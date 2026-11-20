package main

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func tableRequest(t *testing.T, s *Server, user, path string) string {
	t.Helper()
	r := httptest.NewRequest("GET", "/api/v1"+path, nil)
	r.Header.Set("Authorization", "Bearer "+user)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	return w.Body.String()
}
func configureToken(token, layer string) map[string]any {
	return map[string]any{"type": "token.configure", "token_id": token, "layer": layer, "rotation": 90, "size": 1, "markers": []string{"blessed"}, "locked": false, "initiative_bonus": 2}
}

func TestTablePrivacyAcrossEveryReadPath(t *testing.T) {
	s := ownershipFixture(t)
	ownershipCommand(t, s, "gm", configureToken("bandit", "gm"), 200)
	ownershipCommand(t, s, "gm", map[string]any{"type": "token.update", "token_id": "bandit", "name": "SECRET-CREATURE", "hp": 30}, 200)
	ownershipCommand(t, s, "gm", map[string]any{"type": "handout.save", "title": "SECRET-HANDOUT", "text": "Hidden treasure coordinates", "shared": false}, 200)
	ownershipCommand(t, s, "gm", map[string]any{"type": "draw.add", "shape": "text", "x": 1, "y": 1, "to_x": 1, "to_y": 1, "layer": "gm", "text": "SECRET-LABEL", "color": "#aabbcc", "stroke": 2}, 200)
	ownershipCommand(t, s, "gm", map[string]any{"type": "chat.post", "message": "/w gm SECRET-WHISPER"}, 200)
	ownershipCommand(t, s, "gm", map[string]any{"type": "macro.save", "name": "SECRET-MACRO", "formula": "1d20+2"}, 200)
	paths := []string{"/sessions/session/snapshot", "/sessions/session/events?after=0", "/sessions"}
	for _, path := range paths {
		body := tableRequest(t, s, "player", path)
		if strings.Contains(body, "SECRET-") {
			t.Fatalf("private content leaked through %s: %s", path, body)
		}
	}
	conflict := postLiveCommand(t, s, "player", map[string]any{"command_id": "stale-chat", "expected_state_version": 1, "payload": map[string]any{"type": "chat.post", "message": "Hi"}})
	if conflict.Code != 409 || strings.Contains(conflict.Body.String(), "SECRET-") {
		t.Fatal("conflict leaked private content", conflict.Body.String())
	}
	v := map[string]any{"command_id": "public-chat", "expected_state_version": s.db.Sessions["session"].Version, "payload": map[string]any{"type": "chat.post", "message": "Public chat"}}
	for attempt := 0; attempt < 2; attempt++ {
		r := postLiveCommand(t, s, "player", v)
		if r.Code != 200 || strings.Contains(r.Body.String(), "SECRET-") {
			t.Fatal("command/duplicate leaked private content", r.Body.String())
		}
	}
	preview := tableRequest(t, s, "gm", "/sessions/session/snapshot?view=player")
	if strings.Contains(preview, "SECRET-") {
		t.Fatal("player preview leaked GM content")
	}
	server := httptest.NewServer(s)
	defer server.Close()
	c, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http")+"/api/v1/sessions/session/live?after=0", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	_ = c.WriteJSON(map[string]string{"token": "player"})
	sync := readLiveType(t, c, "session.sync")
	raw, _ := json.Marshal(sync)
	if strings.Contains(string(raw), "SECRET-") {
		t.Fatal("live sync/replay leaked GM content")
	}
	ownershipCommand(t, s, "gm", map[string]any{"type": "token.update", "token_id": "bandit", "name": "SECRET-CHANGED", "hp": 20}, 200)
	updated := readLiveType(t, c, "session.updated")
	raw, _ = json.Marshal(updated)
	if strings.Contains(string(raw), "SECRET-") {
		t.Fatal("live update leaked GM content")
	}
	if body := tableRequest(t, s, "gm", "/sessions/session/snapshot"); !strings.Contains(body, "SECRET-CHANGED") || !strings.Contains(body, "SECRET-HANDOUT") {
		t.Fatal("GM lost access to private content")
	}
}

func TestWhispersAndMacrosHaveAccountScopedPrivacy(t *testing.T) {
	s := ownershipFixture(t)
	ownershipCommand(t, s, "player", map[string]any{"type": "chat.send", "message": "/w \"Other player\" SECRET-BETWEEN-PLAYERS"}, 200)
	ownershipCommand(t, s, "player", map[string]any{"type": "chat.post", "message": "/gmroll 2d6+3"}, 200)
	ownershipCommand(t, s, "player", map[string]any{"type": "macro.save", "name": "SECRET-PLAYER-MACRO", "formula": "2d6+3"}, 200)
	gm := tableRequest(t, s, "gm", "/sessions/session/snapshot")
	if strings.Contains(gm, "SECRET-BETWEEN") || strings.Contains(gm, "SECRET-PLAYER-MACRO") {
		t.Fatal("GM received other players' private data", gm)
	}
	other := tableRequest(t, s, "other", "/sessions/session/snapshot")
	if !strings.Contains(other, "SECRET-BETWEEN") || strings.Contains(other, "SECRET-PLAYER-MACRO") || strings.Contains(other, "2d6+3") {
		t.Fatal("whisper/private roll visibility is wrong", other)
	}
	macro := s.db.Sessions["session"].State.Macros[0]
	ownershipCommand(t, s, "other", map[string]any{"type": "macro.run", "macro_id": macro.ID}, 409)
	roll := ownershipCommand(t, s, "player", map[string]any{"type": "macro.run", "macro_id": macro.ID}, 200)
	if roll["event_type"] != "dice.roll" {
		t.Fatal("macro did not roll")
	}
	ownershipCommand(t, s, "player", map[string]any{"type": "chat.post", "message": "not forged", "as_character": "mayank"}, 403)
	ownershipCommand(t, s, "player", map[string]any{"type": "chat.post", "message": "/roll 100d1000"}, 409)
}

func TestScenesGridDrawingsAndInitiativePersist(t *testing.T) {
	s := ownershipFixture(t)
	ownershipCommand(t, s, "gm", map[string]any{"type": "table.settings", "width": 30, "height": 20, "scale": 5, "unit": "ft", "metric": "euclidean"}, 200)
	draw := map[string]any{"type": "draw.add", "shape": "ellipse", "x": 1, "y": 1, "to_x": 3, "to_y": 4, "layer": "foreground", "color": "#aa3344", "stroke": 3}
	ownershipCommand(t, s, "gm", draw, 200)
	ownershipCommand(t, s, "gm", map[string]any{"type": "initiative.roll"}, 200)
	ownershipCommand(t, s, "gm", map[string]any{"type": "initiative.start"}, 200)
	before := s.db.Sessions["session"].State
	first := before.Initiative[0].TokenID
	ownershipCommand(t, s, "gm", map[string]any{"type": "turn.end", "character_id": first}, 200)
	if s.db.Sessions["session"].State.ActiveActorID == first {
		t.Fatal("turn did not follow initiative")
	}
	ownershipCommand(t, s, "gm", map[string]any{"type": "scene.create", "name": "Dungeon", "duplicate": false}, 200)
	w := s.db.Sessions["session"].State
	if len(w.Actors) != 0 || len(w.Drawings) != 0 || w.MapImage != "blank" || len(w.Pages) != 1 {
		t.Fatal("new scene did not isolate state")
	}
	oldID := w.Pages[0].ID
	ownershipCommand(t, s, "gm", map[string]any{"type": "scene.switch", "scene_id": oldID}, 200)
	w = s.db.Sessions["session"].State
	if w.Width != 30 || len(w.Drawings) != 1 || len(w.Actors) != 3 || len(w.Initiative) != 2 {
		t.Fatal("scene state lost")
	}
	s = &Server{db: newDB(s.db.file)}
	w = s.db.Sessions["session"].State
	if w.Width != 30 || w.Scale != 5 || len(w.Pages) != 1 {
		t.Fatal("scene/grid not persisted")
	}
	ownershipCommand(t, s, "player", map[string]any{"type": "scene.create", "name": "hijack"}, 403)
	ownershipCommand(t, s, "gm", map[string]any{"type": "table.settings", "width": 2, "height": 2, "scale": 5, "unit": "ft", "metric": "euclidean"}, 409)
	draw["color"] = "url(javascript:bad)"
	ownershipCommand(t, s, "gm", draw, 409)
}

func TestTokenFootprintsAndMovementLocks(t *testing.T) {
	s := ownershipFixture(t)
	v := configureToken("lira", "tokens")
	v["size"] = 2
	v["locked"] = true
	ownershipCommand(t, s, "gm", v, 200)
	ownershipCommand(t, s, "gm", assignment("lira", "player"), 200)
	ownershipCommand(t, s, "player", map[string]any{"type": "character.move", "character_id": "lira", "target": map[string]any{"x": 5, "y": 3}}, 403)
	if canOccupy(s.db.Sessions["session"].State, Point{5, 4}, "mayank", 1) {
		t.Fatal("large token did not occupy its full footprint")
	}
	ownershipCommand(t, s, "gm", map[string]any{"type": "table.pause", "paused": true}, 200)
	ownershipCommand(t, s, "gm", assignment("mayank", "player"), 200)
	ownershipCommand(t, s, "player", map[string]any{"type": "character.move", "character_id": "mayank", "target": map[string]any{"x": 9, "y": 5}}, 403)
	ownershipCommand(t, s, "player", map[string]any{"type": "turn.end", "character_id": "mayank"}, 200) // A movement lock does not block ending a turn.
	ownershipCommand(t, s, "gm", map[string]any{"type": "audio.set", "url": "javascript:alert(1)", "playing": true}, 409)
}
