package main

import (
	"encoding/json"
	"net"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestMediaSignalingIsBoundToAuthenticatedSessionAndPeer(t *testing.T) {
	s := liveFixture(t)
	s.db.Campaigns["foreign"] = Campaign{ID: "foreign", OwnerID: "outsider", Members: map[string]string{"outsider": "owner"}}
	s.db.Sessions["foreign"] = Session{ID: "foreign", CampaignID: "foreign", Version: 1}
	server := httptest.NewServer(s)
	defer server.Close()
	connect := func(user, sid string) (*websocket.Conn, string) {
		t.Helper()
		c, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http")+"/api/v1/sessions/"+sid+"/live?after=0", nil)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { c.Close() })
		_ = c.WriteJSON(map[string]string{"token": user})
		sync := readLiveType(t, c, "session.sync")
		var id string
		_ = json.Unmarshal(sync["connection_id"], &id)
		_ = c.WriteJSON(map[string]string{"type": "rtc.join"})
		readLiveType(t, c, "presence")
		return c, id
	}
	gm, gmID := connect("gm", "session")
	player, playerID := connect("player", "session")
	other, otherID := connect("outsider", "foreign")
	_ = gm.WriteJSON(map[string]any{"type": "rtc.signal", "to": playerID, "from": "spoofed", "payload": map[string]string{"kind": "offer", "test": "PRIVATE-SIGNAL"}})
	message := readLiveType(t, player, "rtc.signal")
	var sender string
	_ = json.Unmarshal(message["from"], &sender)
	if sender != gmID {
		t.Fatal("signal sender was not authenticated")
	}
	_ = gm.WriteJSON(map[string]any{"type": "rtc.signal", "to": otherID, "payload": map[string]string{"kind": "offer", "test": "CROSS-SESSION"}})
	_ = other.SetReadDeadline(time.Now().Add(250 * time.Millisecond))
	for {
		_, raw, err := other.ReadMessage()
		if err != nil {
			if timeout, ok := err.(net.Error); !ok || !timeout.Timeout() {
				t.Fatal(err)
			}
			break
		}
		if strings.Contains(string(raw), "CROSS-SESSION") {
			t.Fatal("media signal crossed a campaign boundary")
		}
	}
	if len(s.db.Sessions["session"].Events) != 0 {
		t.Fatal("private media negotiation was persisted as a campaign event")
	}
}
