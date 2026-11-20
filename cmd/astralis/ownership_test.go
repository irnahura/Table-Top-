package main

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func ownershipFixture(t *testing.T) *Server {
	s := liveFixture(t)
	s.db.Users["other"] = User{ID: "other", DisplayName: "Other player"}
	s.db.Tokens["other"] = "other"
	s.db.TokenExpiry["other"] = time.Now().Add(time.Hour)
	c := s.db.Campaigns["campaign"]
	c.Members["other"] = "player"
	s.db.Campaigns[c.ID] = c
	ss := s.db.Sessions["session"]
	ss.State = World{Width: 20, Height: 14, Round: 1, ActiveActorID: "mayank", Actors: []Actor{
		{ID: "mayank", Name: "Mayank", X: 8, Y: 5, HP: 30, MaxHP: 30, Movement: 6, Damage: 4},
		{ID: "lira", Name: "Lira", X: 4, Y: 3, HP: 30, MaxHP: 30, Movement: 6},
		{ID: "bandit", Name: "Bandit", X: 10, Y: 5, HP: 30, MaxHP: 30, Hostile: true},
	}}
	s.db.Sessions[ss.ID] = ss
	return s
}

func assignment(token, controller string) map[string]any {
	return map[string]any{"type": "token.assign", "token_id": token, "controller_id": controller}
}

func ownershipCommand(t *testing.T, s *Server, user string, payload map[string]any, status int) map[string]any {
	t.Helper()
	before, _ := json.Marshal(s.db.Sessions["session"])
	r := postLiveCommand(t, s, user, map[string]any{"payload": payload})
	if r.Code != status {
		t.Fatalf("%s %v: %d want %d: %s", user, payload, r.Code, status, r.Body.String())
	}
	if status != 200 {
		after, _ := json.Marshal(s.db.Sessions["session"])
		if string(before) != string(after) {
			t.Fatal("rejected command changed session")
		}
	}
	var result map[string]any
	_ = json.Unmarshal(r.Body.Bytes(), &result)
	return result
}

func TestGameplayRequiresAssignedController(t *testing.T) {
	commands := []map[string]any{
		{"type": "character.move", "character_id": "mayank", "target": map[string]any{"x": 9, "y": 5}},
		{"type": "combat.attack", "attacker_id": "mayank", "target_id": "bandit"},
		{"type": "object.interact", "character_id": "mayank", "object_id": "old-bridge"},
		{"type": "turn.end", "character_id": "mayank"},
	}
	for _, command := range commands {
		t.Run(command["type"].(string), func(t *testing.T) {
			s := ownershipFixture(t)
			// Old saves have no controller: only the GM can operate those actors.
			ownershipCommand(t, s, "player", command, 403)
			ownershipCommand(t, s, "gm", assignment("mayank", "player"), 200)
			ownershipCommand(t, s, "other", command, 403)
			ownershipCommand(t, s, "player", command, 200)
		})
	}
}

func TestAssignmentValidationAndGMOverride(t *testing.T) {
	s := ownershipFixture(t)
	ownershipCommand(t, s, "player", assignment("mayank", "player"), 403)
	ownershipCommand(t, s, "outsider", assignment("mayank", "outsider"), 403)
	for _, payload := range []map[string]any{assignment("missing", "player"), assignment("bandit", "player"), assignment("mayank", "outsider"), assignment("mayank", "missing"), {"type": "token.assign", "token_id": "mayank"}, {"type": "token.assign", "token_id": "mayank", "controller_id": 42}} {
		ownershipCommand(t, s, "gm", payload, 409)
	}
	ownershipCommand(t, s, "gm", assignment("mayank", "player"), 200)
	move := map[string]any{"type": "character.move", "character_id": "mayank", "target": map[string]any{"x": 9, "y": 5}}
	ownershipCommand(t, s, "gm", move, 200) // GM override remains available after assignment.
	ownershipCommand(t, s, "gm", assignment("mayank", "other"), 200)
	ownershipCommand(t, s, "player", move, 403)
	ownershipCommand(t, s, "gm", assignment("mayank", ""), 200)
	ownershipCommand(t, s, "other", move, 403)
	ss := s.db.Sessions["session"]
	if ss.State.Actors[0].ControllerID != "" || ss.State.Actors[0].ControllerName != "" {
		t.Fatal("clear assignment failed")
	}
	// A role field alone never grants GM authority.
	c := s.db.Campaigns["campaign"]
	c.Members["player"] = "owner"
	s.db.Campaigns[c.ID] = c
	ownershipCommand(t, s, "player", assignment("mayank", "player"), 403)
}

func TestOwnershipCannotBeForgedOrBypassedOnRetry(t *testing.T) {
	s := ownershipFixture(t)
	for _, command := range []map[string]any{
		{"type": "character.move", "character_id": "mayank", "controller_id": "player", "user_id": "gm", "role": "owner", "target": map[string]any{"x": 9, "y": 5}},
		{"type": "combat.attack", "attacker_id": "mayank", "character_id": "lira", "controller_id": "player", "target_id": "bandit"},
		{"type": "turn.end", "character_id": "mayank", "attacker_id": "lira"},
	} {
		ownershipCommand(t, s, "player", command, 403)
	}
	ownershipCommand(t, s, "gm", assignment("mayank", "player"), 200)
	// Combat must authorize and execute the same actor even with conflicting IDs.
	attack := ownershipCommand(t, s, "player", map[string]any{"type": "combat.attack", "attacker_id": "mayank", "character_id": "lira", "target_id": "bandit"}, 200)
	if attack["payload"].(map[string]any)["attacker_id"] != "mayank" {
		t.Fatal("combat actor confused")
	}
	v := map[string]any{"command_id": "one-move", "expected_state_version": s.db.Sessions["session"].Version, "payload": map[string]any{"type": "character.move", "character_id": "mayank", "target": map[string]any{"x": 9, "y": 5}}}
	if r := postLiveCommand(t, s, "player", v); r.Code != 200 {
		t.Fatal(r.Body.String())
	}
	ownershipCommand(t, s, "gm", assignment("mayank", "other"), 200)
	seq := s.db.Sessions["session"].Sequence
	if r := postLiveCommand(t, s, "player", v); r.Code != 403 {
		t.Fatal("revoked assignee bypassed authorization with a receipt", r.Body.String())
	}
	if s.db.Sessions["session"].Sequence != seq {
		t.Fatal("denied retry mutated state")
	}
	// A hostile token cannot be controlled even if old/corrupt data assigns it.
	ss := s.db.Sessions["session"]
	ss.State.Actors[2].ControllerID = "player"
	s.db.Sessions[ss.ID] = ss
	ownershipCommand(t, s, "player", map[string]any{"type": "turn.end", "character_id": "bandit"}, 403)
}

func TestAssignmentsSurviveRestartAndMembershipLoss(t *testing.T) {
	s := ownershipFixture(t)
	ownershipCommand(t, s, "gm", assignment("mayank", "player"), 200)
	s = &Server{db: newDB(s.db.file)}
	a := s.db.Sessions["session"].State.Actors[0]
	if a.ControllerID != "player" || a.ControllerName != "Player" {
		t.Fatal("assignment lost on restart", a)
	}
	move := map[string]any{"type": "character.move", "character_id": "mayank", "target": map[string]any{"x": 9, "y": 5}}
	ownershipCommand(t, s, "player", move, 200)
	c := s.db.Campaigns["campaign"]
	delete(c.Members, "player")
	s.db.Campaigns[c.ID] = c
	ownershipCommand(t, s, "player", move, 403)
}

func TestCampaignMembersAreScopedAndPrivate(t *testing.T) {
	s := ownershipFixture(t)
	for _, test := range []struct {
		user   string
		status int
	}{{"gm", 200}, {"player", 200}, {"outsider", 403}, {"", 401}} {
		r := httptest.NewRequest("GET", "/api/v1/campaigns/campaign/members", nil)
		r.Header.Set("Authorization", "Bearer "+test.user)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, r)
		if w.Code != test.status {
			t.Fatal(test.user, w.Code, w.Body.String())
		}
		if w.Code == 200 {
			var result struct {
				Members []map[string]any `json:"members"`
			}
			_ = json.Unmarshal(w.Body.Bytes(), &result)
			if len(result.Members) != 3 {
				t.Fatal("wrong campaign members")
			}
			for _, m := range result.Members {
				if len(m) != 3 || m["id"] == "outsider" || m["display_name"] == nil || m["role"] == nil {
					t.Fatal("unexpected member data", m)
				}
			}
			if strings.Contains(w.Body.String(), "password") || strings.Contains(w.Body.String(), "token") || strings.Contains(w.Body.String(), "email") {
				t.Fatal("private account data exposed")
			}
		}
	}
}
