package main

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestInvitationLifecycle(t *testing.T) {
	s := liveFixture(t)
	request := func(user, method, path, body string, status int) map[string]any {
		t.Helper()
		r := httptest.NewRequest(method, "/api/v1"+path, strings.NewReader(body))
		r.Header.Set("Authorization", "Bearer "+user)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, r)
		if w.Code != status {
			t.Fatalf("%s: %d %s", path, w.Code, w.Body.String())
		}
		var out map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &out)
		return out
	}
	path := "/campaigns/campaign/invitations"
	request("player", "POST", path, `{}`, 403)
	request("gm", "POST", path, `{"max_uses":1.5}`, 400)
	invite := request("gm", "POST", path, `{"max_uses":1,"expires_in_hours":1}`, 201)["token"].(string)
	redeem := `{"token":"` + invite + `"}`
	request("gm", "POST", "/invitations/redeem", redeem, 200)
	if s.db.Campaigns["campaign"].Members["gm"] != "owner" || s.db.InviteRules[invite].Uses != 0 {
		t.Fatal("owner role or uses changed")
	}
	request("outsider", "POST", "/invitations/redeem", redeem, 200)
	if s.db.InviteRules[invite].Uses != 1 {
		t.Fatal("missing use")
	}
	request("player", "POST", "/invitations/redeem", redeem, 409)
	s = &Server{db: newDB(s.db.file)}
	request("player", "GET", path, ``, 403)
	request("gm", "GET", path, ``, 200)
	request("gm", "DELETE", path+"/"+invite, ``, 200)
	if !s.db.InviteRules[invite].Revoked {
		t.Fatal("revoke not saved")
	}
	request("outsider", "POST", "/invitations/redeem", redeem, 409)
	expired := request("gm", "POST", path, `{}`, 201)["token"].(string)
	rule := s.db.InviteRules[expired]
	rule.ExpiresAt = time.Now().Add(-time.Second)
	s.db.InviteRules[expired] = rule
	request("outsider", "POST", "/invitations/redeem", `{"token":"`+expired+`"}`, 409)
	s.db.Invites["legacy"] = "campaign"
	request("outsider", "POST", "/invitations/redeem", `{"token":"legacy"}`, 200)
}

func TestAdventurePrivacyPermissionsAndPersistence(t *testing.T) {
	s := ownershipFixture(t)
	sheet := map[string]any{"type": "sheet.save", "name": "Hero", "controller_id": "player", "shared": false, "details": "PRIVATE-BACKSTORY", "stats": "Agility +2", "inventory": "Rope x1", "spells": "Spark: 2 uses", "resources": "Gold: 10"}
	result := ownershipCommand(t, s, "gm", sheet, 200)
	sid := result["payload"].(map[string]any)["sheet_id"].(string)
	if len(s.sessionView(s.db.Sessions["session"], "other").State.Sheets) != 0 {
		t.Fatal("private sheet leaked")
	}
	sheet["sheet_id"] = sid
	sheet["inventory"] = "Rope x2"
	ownershipCommand(t, s, "other", sheet, 403)
	ownershipCommand(t, s, "player", sheet, 200)
	entry := map[string]any{"type": "sheet.entry.save", "sheet_id": sid, "kind": "spell", "name": "Spark", "quantity": 1, "capacity": 2, "prepared": false}
	ownershipCommand(t, s, "player", entry, 200)
	entryID := s.db.Sessions["session"].State.Sheets[0].Entries[0].ID
	use := map[string]any{"type": "sheet.entry.use", "sheet_id": sid, "entry_id": entryID}
	ownershipCommand(t, s, "other", use, 403)
	ownershipCommand(t, s, "player", use, 409)
	entry["entry_id"] = entryID
	entry["prepared"] = true
	ownershipCommand(t, s, "player", entry, 200)
	ownershipCommand(t, s, "player", use, 200)
	ownershipCommand(t, s, "player", use, 409)
	// Editing sheet notes must retain tracked inventory and spell quantities.
	ownershipCommand(t, s, "player", sheet, 200)
	if len(s.db.Sessions["session"].State.Sheets[0].Entries) != 1 || s.db.Sessions["session"].State.Sheets[0].Entries[0].Quantity != 0 {
		t.Fatal("sheet edit lost tracked entries")
	}
	ownershipCommand(t, s, "player", map[string]any{"type": "sheet.roll", "sheet_id": sid, "formula": "2d6+3"}, 200)
	quest := map[string]any{"type": "quest.save", "title": "Bridge", "description": "PRIVATE-QUEST", "objectives": "[ ] Cross", "status": "active", "shared": false}
	ownershipCommand(t, s, "gm", quest, 200)
	ownershipCommand(t, s, "player", quest, 403)
	table := map[string]any{"type": "rolltable.save", "name": "Private loot", "entries": "1 | SECRET-LOOT", "shared": false, "replacement": false}
	result = ownershipCommand(t, s, "gm", table, 200)
	tid := result["payload"].(map[string]any)["table_id"].(string)
	draw := map[string]any{"type": "rolltable.draw", "table_id": tid}
	ownershipCommand(t, s, "gm", draw, 200)
	ownershipCommand(t, s, "gm", draw, 409)
	view, _ := json.Marshal(s.sessionView(s.db.Sessions["session"], "other"))
	if strings.Contains(string(view), "PRIVATE-") || strings.Contains(string(view), "SECRET-LOOT") {
		t.Fatal("private content leaked")
	}
	ownershipCommand(t, s, "gm", map[string]any{"type": "rolltable.reset", "table_id": tid}, 200)
	ownershipCommand(t, s, "gm", draw, 200)
	ownershipCommand(t, s, "gm", map[string]any{"type": "scene.create", "name": "Elsewhere", "duplicate": false}, 200)
	ss := s.db.Sessions["session"]
	page := ss.State.Pages[0].ID
	ownershipCommand(t, s, "gm", map[string]any{"type": "scene.switch", "scene_id": page}, 200)
	s = &Server{db: newDB(s.db.file)}
	ss = s.db.Sessions["session"]
	if len(ss.State.Sheets) != 1 || ss.State.Sheets[0].Inventory != "Rope x2" || len(ss.State.Quests) != 1 || len(ss.State.RollTables) != 1 {
		t.Fatal("adventure data lost on scene switch/restart")
	}
}
