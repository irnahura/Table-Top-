package main

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestSheetTokenLinkPlacementPrivacyAndPersistence(t *testing.T) {
	s := ownershipFixture(t)
	saved := ownershipCommand(t, s, "gm", map[string]any{"type": "sheet.save", "name": "River scout", "controller_id": "player", "shared": false, "details": "PRIVATE-RIVER-NOTES"}, 200)
	sheetID := saved["payload"].(map[string]any)["sheet_id"].(string)
	before := s.db.Sessions["session"].State
	_, original := actor(before, "mayank")
	originalActor := *original
	link := map[string]any{"type": "sheet.link", "sheet_id": sheetID, "token_id": "mayank", "unlink": false}
	ownershipCommand(t, s, "player", link, 403)
	ownershipCommand(t, s, "gm", link, 200)
	_, linked := actor(s.db.Sessions["session"].State, "mayank")
	if linked.SheetID != sheetID || linked.HP != originalActor.HP || linked.ControllerID != originalActor.ControllerID || linked.AttackBonus != originalActor.AttackBonus {
		t.Fatal("link changed stats/control or failed to persist")
	}
	_, playerToken := actor(s.sessionView(s.db.Sessions["session"], "player").State, "mayank")
	_, otherToken := actor(s.sessionView(s.db.Sessions["session"], "other").State, "mayank")
	if playerToken.SheetID != sheetID || otherToken.SheetID != "" {
		t.Fatal("private sheet reference projection failed")
	}
	place := map[string]any{"type": "sheet.place", "sheet_id": sheetID, "name": "River scout", "hp": 12, "x": 0, "y": 0, "hostile": false, "layer": "tokens"}
	ownershipCommand(t, s, "player", place, 403)
	result := ownershipCommand(t, s, "gm", place, 200)
	tokenID := result["payload"].(map[string]any)["token_id"].(string)
	_, placed := actor(s.db.Sessions["session"].State, tokenID)
	if placed.SheetID != sheetID || placed.ControllerID != "player" || placed.HP != 12 || placed.Movement != 6 || placed.Hostile {
		t.Fatal("new party token did not inherit sheet control and explicit/default stats")
	}
	// A rejected placement must neither change the actor slice nor consume state.
	unchanged := s.db.Sessions["session"].State
	ownershipCommand(t, s, "gm", place, 409)
	if !reflect.DeepEqual(unchanged, s.db.Sessions["session"].State) {
		t.Fatal("occupied placement mutated state")
	}
	place["x"], place["y"], place["layer"], place["hostile"] = 1, 0, "gm", true
	hidden := ownershipCommand(t, s, "gm", place, 200)["payload"].(map[string]any)["token_id"].(string)
	_, npc := actor(s.db.Sessions["session"].State, hidden)
	if npc.ControllerID != "" {
		t.Fatal("hostile NPC inherited player control")
	}
	view := s.sessionView(s.db.Sessions["session"], "player")
	for _, ev := range view.Events {
		b, _ := json.Marshal(ev)
		if strings.Contains(string(b), hidden) {
			t.Fatal("hidden sheet token receipt leaked into player event history")
		}
	}
	if _, a := actor(view.State, hidden); a != nil {
		t.Fatal("GM-layer sheet token leaked")
	}
	// Reassignment changes sheet editing only, not already placed token authority.
	ownershipCommand(t, s, "gm", map[string]any{"type": "sheet.save", "sheet_id": sheetID, "name": "River scout", "controller_id": "other", "shared": false}, 200)
	_, placed = actor(s.db.Sessions["session"].State, tokenID)
	if placed.ControllerID != "player" {
		t.Fatal("sheet reassignment silently changed existing token control")
	}
	_, playerToken = actor(s.sessionView(s.db.Sessions["session"], "player").State, tokenID)
	if playerToken.SheetID != "" {
		t.Fatal("revoked private sheet reference still visible")
	}
	ownershipCommand(t, s, "gm", map[string]any{"type": "scene.create", "name": "Linked archive", "duplicate": true}, 200)
	restarted := &Server{db: newDB(s.db.file)}
	_, restored := actor(restarted.db.Sessions["session"].State, tokenID)
	if restored == nil || restored.SheetID != sheetID {
		t.Fatal("link did not survive scene creation and restart")
	}
	ownershipCommand(t, restarted, "gm", map[string]any{"type": "sheet.delete", "sheet_id": sheetID}, 200)
	var verify func(World)
	verify = func(w World) {
		for _, a := range w.Actors {
			if a.SheetID == sheetID {
				t.Fatal("deleted sheet retained a token link")
			}
		}
		for _, page := range w.Pages {
			if page.State != nil {
				verify(*page.State)
			}
		}
	}
	verify(restarted.db.Sessions["session"].State)
}

func TestSheetTokenValidationAndUnlink(t *testing.T) {
	s := ownershipFixture(t)
	result := ownershipCommand(t, s, "gm", map[string]any{"type": "sheet.save", "name": "Navigator", "shared": true}, 200)
	sid := result["payload"].(map[string]any)["sheet_id"].(string)
	for _, p := range []map[string]any{
		{"type": "sheet.link", "sheet_id": "missing", "token_id": "mayank", "unlink": false},
		{"type": "sheet.link", "sheet_id": sid, "token_id": "missing", "unlink": false},
		{"type": "sheet.link", "sheet_id": sid, "token_id": "mayank", "unlink": "false"},
		{"type": "sheet.link", "sheet_id": sid, "token_id": "mayank", "unlink": true},
		{"type": "sheet.place", "sheet_id": sid, "name": "Navigator", "hp": 20, "x": -1, "y": 0, "hostile": false, "layer": "tokens"},
		{"type": "sheet.place", "sheet_id": sid, "name": "Navigator", "hp": 20, "x": 1.5, "y": 0, "hostile": false, "layer": "tokens"},
		{"type": "sheet.place", "sheet_id": sid, "name": "Navigator", "hp": 0, "x": 0, "y": 0, "hostile": false, "layer": "tokens"},
		{"type": "sheet.place", "sheet_id": sid, "name": "Navigator", "hp": 20, "x": 0, "y": 0, "hostile": false, "layer": "unknown"},
	} {
		before, _ := json.Marshal(s.db.Sessions["session"])
		ownershipCommand(t, s, "gm", p, 409)
		after, _ := json.Marshal(s.db.Sessions["session"])
		if string(before) != string(after) {
			t.Fatal("rejected link or placement mutated saved session")
		}
	}
	ownershipCommand(t, s, "gm", map[string]any{"type": "sheet.link", "sheet_id": sid, "token_id": "mayank", "unlink": false}, 200)
	ownershipCommand(t, s, "gm", map[string]any{"type": "sheet.link", "sheet_id": sid, "token_id": "mayank", "unlink": true}, 200)
	_, a := actor(s.db.Sessions["session"].State, "mayank")
	if a.SheetID != "" {
		t.Fatal("unlink failed")
	}
}
