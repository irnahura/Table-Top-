package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func liveFixture(t *testing.T) *Server {
	t.Helper()
	db := newDB(filepath.Join(t.TempDir(), "game.json"))
	db.Users["gm"] = User{ID: "gm", DisplayName: "Game master"}
	db.Users["player"] = User{ID: "player", DisplayName: "Player"}
	db.Users["outsider"] = User{ID: "outsider", DisplayName: "Outsider"}
	for _, user := range []string{"gm", "player", "outsider"} {
		db.Tokens[user] = user
		db.TokenExpiry[user] = time.Now().Add(time.Hour)
	}
	db.Campaigns["campaign"] = Campaign{ID: "campaign", OwnerID: "gm", Members: map[string]string{"gm": "owner", "player": "player"}}
	db.Sessions["session"] = Session{ID: "session", CampaignID: "campaign", Version: 1, State: World{Width: 20, Height: 14}}
	return &Server{db: db}
}

func postLiveCommand(t *testing.T, s *Server, user string, v map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(v)
	r := httptest.NewRequest("POST", "/api/v1/sessions/session/commands", strings.NewReader(string(b)))
	r.Header.Set("Authorization", "Bearer "+user)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	return w
}

func TestCommandDeduplicationAndVersions(t *testing.T) {
	s := liveFixture(t)
	payload := map[string]any{"type": "dice.roll", "count": 2, "sides": 6, "bonus": 0}
	v := map[string]any{"command_id": "roll-once", "expected_state_version": 1, "payload": payload}
	first := postLiveCommand(t, s, "gm", v)
	if first.Code != 200 {
		t.Fatal(first.Body.String())
	}
	var accepted map[string]any
	_ = json.Unmarshal(first.Body.Bytes(), &accepted)
	// Retrying after a restart returns the original dice, even with an old version.
	s = &Server{db: newDB(s.db.file)}
	second := postLiveCommand(t, s, "gm", v)
	var duplicate map[string]any
	_ = json.Unmarshal(second.Body.Bytes(), &duplicate)
	a, _ := json.Marshal(accepted["payload"])
	b, _ := json.Marshal(duplicate["payload"])
	if second.Code != 200 || duplicate["duplicate"] != true || string(a) != string(b) || s.db.Sessions["session"].Sequence != 1 {
		t.Fatal("retry changed a roll", second.Body.String())
	}
	v["payload"] = map[string]any{"type": "chat.send", "message": "different"}
	if r := postLiveCommand(t, s, "gm", v); r.Code != 409 {
		t.Fatal("reused ID accepted", r.Body.String())
	}
	v["command_id"] = "stale"
	if r := postLiveCommand(t, s, "gm", v); r.Code != 409 || !strings.Contains(r.Body.String(), "STATE_VERSION_CONFLICT") {
		t.Fatal("stale action accepted", r.Body.String())
	}
	v["expected_state_version"] = 2
	if r := postLiveCommand(t, s, "outsider", v); r.Code != 403 {
		t.Fatal("membership bypass", r.Body.String())
	}
	v["expected_state_version"] = 2.5
	if r := postLiveCommand(t, s, "gm", v); r.Code != 400 {
		t.Fatal("fractional version accepted")
	}
	delete(v, "expected_state_version")
	if r := postLiveCommand(t, s, "gm", v); r.Code != 400 {
		t.Fatal("missing version accepted")
	}
	if s.db.Sessions["session"].Sequence != 1 {
		t.Fatal("rejections mutated sequence")
	}
}

func TestConcurrentCommandsRejectStaleState(t *testing.T) {
	s := liveFixture(t)
	var wg sync.WaitGroup
	results := make(chan int, 2)
	for _, cid := range []string{"first", "second"} {
		wg.Add(1)
		go func(cid string) {
			defer wg.Done()
			results <- postLiveCommand(t, s, "gm", map[string]any{"command_id": cid, "expected_state_version": 1, "payload": map[string]any{"type": "chat.send", "message": cid}}).Code
		}(cid)
	}
	wg.Wait()
	close(results)
	counts := map[int]int{}
	for code := range results {
		counts[code]++
	}
	if counts[200] != 1 || counts[409] != 1 || s.db.Sessions["session"].Sequence != 1 {
		t.Fatal("concurrent commands both applied", counts)
	}
}

func readLiveType(t *testing.T, c *websocket.Conn, kind string) map[string]json.RawMessage {
	t.Helper()
	_ = c.SetReadDeadline(time.Now().Add(3 * time.Second))
	for {
		var message map[string]json.RawMessage
		if err := c.ReadJSON(&message); err != nil {
			t.Fatal(err)
		}
		var got string
		_ = json.Unmarshal(message["type"], &got)
		if got == kind {
			return message
		}
	}
}

func TestLiveReplayPresenceAndAuthorization(t *testing.T) {
	s := liveFixture(t)
	server := httptest.NewServer(s)
	defer server.Close()
	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/v1/sessions/session/live?after=0"
	connect := func(token string) *websocket.Conn {
		t.Helper()
		c, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { c.Close() })
		if err = c.WriteJSON(map[string]string{"token": token}); err != nil {
			t.Fatal(err)
		}
		return c
	}
	gm := connect("gm")
	readLiveType(t, gm, "session.sync")
	player := connect("player")
	readLiveType(t, player, "session.sync")
	presence := readLiveType(t, player, "presence")
	var users map[string]string
	_ = json.Unmarshal(presence["users"], &users)
	if len(users) != 2 {
		t.Fatal("missing presence", users)
	}
	v := map[string]any{"command_id": "chat", "expected_state_version": 1, "payload": map[string]any{"type": "chat.send", "message": "Hello live table"}}
	if r := postLiveCommand(t, s, "player", v); r.Code != 200 {
		t.Fatal(r.Body.String())
	}
	for _, c := range []*websocket.Conn{gm, player} {
		m := readLiveType(t, c, "session.updated")
		var ss Session
		_ = json.Unmarshal(m["snapshot"], &ss)
		if ss.Sequence != 1 {
			t.Fatal("missing live event")
		}
	}
	player.Close()
	// A disconnected member catches up using the last received sequence.
	postLiveCommand(t, s, "gm", map[string]any{"payload": map[string]any{"type": "chat.send", "message": "While away"}})
	url = strings.TrimSuffix(url, "0") + "1"
	reconnected := connect("player")
	m := readLiveType(t, reconnected, "session.sync")
	var events []Event
	_ = json.Unmarshal(m["events"], &events)
	if len(events) != 1 || events[0].Sequence != 2 {
		t.Fatal("reconnect replay is wrong", string(m["events"]))
	}
	for _, token := range []string{"outsider", "invalid"} {
		c := connect(token)
		_ = c.SetReadDeadline(time.Now().Add(3 * time.Second))
		if _, _, err := c.ReadMessage(); !websocket.IsCloseError(err, 1008) {
			t.Fatal("unauthorized connection accepted", err)
		}
	}
	header := http.Header{"Origin": []string{"https://unrelated.example"}}
	if c, response, err := websocket.DefaultDialer.Dial(url, header); err == nil {
		c.Close()
		t.Fatal("cross-origin socket accepted")
	} else if response == nil || response.StatusCode != 403 {
		t.Fatal("unexpected origin rejection", err)
	}
	// Revocation is checked again before any new state is delivered.
	s.db.mu.Lock()
	delete(s.db.Tokens, "player")
	s.db.mu.Unlock()
	postLiveCommand(t, s, "gm", map[string]any{"payload": map[string]any{"type": "chat.send", "message": "After logout"}})
	_ = reconnected.SetReadDeadline(time.Now().Add(3 * time.Second))
	for {
		_, data, err := reconnected.ReadMessage()
		if err != nil {
			break
		}
		if strings.Contains(string(data), "After logout") {
			t.Fatal("revoked token received state")
		}
	}
}
