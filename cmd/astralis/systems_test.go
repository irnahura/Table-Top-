package main

import (
	"encoding/json"
	"math"
	"reflect"
	"strings"
	"testing"
)

func TestPathfinderDegreeBoundaries(t *testing.T) {
	for _, tc := range []struct {
		face, total, target int
		want                string
	}{
		{10, 30, 20, "Critical success"}, {10, 29, 20, "Success"}, {10, 20, 20, "Success"}, {10, 19, 20, "Failure"}, {10, 10, 20, "Critical failure"}, {10, 11, 20, "Failure"},
		{20, 20, 40, "Failure"}, {20, 20, 25, "Success"}, {20, 30, 30, "Critical success"}, {1, 51, 10, "Success"}, {1, 20, 20, "Failure"}, {1, 10, 20, "Critical failure"},
	} {
		if got := pathfinderDegree(tc.face, tc.total, tc.target); got != tc.want {
			t.Errorf("face %d total %d DC %d: got %s, want %s", tc.face, tc.total, tc.target, got, tc.want)
		}
	}
}

func TestSystemSelectionPersistenceAndAuthorization(t *testing.T) {
	s := ownershipFixture(t)
	ownershipCommand(t, s, "player", map[string]any{"type": "system.select", "system": "dnd5e"}, 403)
	ownershipCommand(t, s, "gm", map[string]any{"type": "system.select", "system": "unknown"}, 409)
	before := s.db.Sessions["session"].State
	ownershipCommand(t, s, "gm", map[string]any{"type": "system.select", "system": "dnd5e2024"}, 200)
	w := s.db.Sessions["session"].State
	if w.GameSystem != "dnd5e2024" || w.Scale != before.Scale || !reflect.DeepEqual(w.Actors, before.Actors) {
		t.Fatal("profile changed encounter stats or scale")
	}
	ownershipCommand(t, s, "gm", map[string]any{"type": "scene.create", "name": "Other", "duplicate": false}, 200)
	ownershipCommand(t, s, "gm", map[string]any{"type": "system.select", "system": "fate"}, 200)
	ownershipCommand(t, s, "gm", map[string]any{"type": "scene.switch", "scene_id": s.db.Sessions["session"].State.Pages[0].ID}, 200)
	if s.db.Sessions["session"].State.GameSystem != "fate" {
		t.Fatal("scene switch lost global profile")
	}
	restored := newDB(s.db.file)
	if restored.Sessions["session"].State.GameSystem != "fate" {
		t.Fatal("profile not persisted")
	}
}

func TestPrivateQuickDiceAcrossReadPaths(t *testing.T) {
	s := ownershipFixture(t)
	for _, mode := range []string{"gm", "secret"} {
		r := ownershipCommand(t, s, "player", map[string]any{"type": "dice.quick", "count": 3, "sides": 6, "bonus": 2, "visibility": mode}, 200)
		p := r["payload"].(map[string]any)
		if len(p["rolls"].([]any)) != 3 {
			t.Fatal(p)
		}
		sequence := s.db.Sessions["session"].Sequence
		for _, user := range []string{"gm", "player", "other"} {
			events := s.sessionView(s.db.Sessions["session"], user).Events
			found := false
			for _, ev := range events {
				if ev.Sequence == sequence {
					found = true
				}
			}
			want := user == "player" || mode == "gm" && user == "gm"
			if found != want {
				t.Fatalf("%s audience %s: visible %v", mode, user, found)
			}
		}
	}
	for _, path := range []string{"/sessions/session/snapshot", "/sessions/session/events?after=0", "/sessions"} {
		if strings.Contains(tableRequest(t, s, "other", path), "dice.roll") {
			t.Fatal("private dice leaked", path)
		}
	}
	ownershipCommand(t, s, "player", map[string]any{"type": "dice.quick", "count": 1, "sides": 20, "bonus": 0, "visibility": "everyone"}, 409)
}

func TestSystemDiceAndValidation(t *testing.T) {
	s := ownershipFixture(t)
	for _, system := range []string{"dnd5e", "dnd5e2024", "pf2e", "warhammer40k", "fate", "percentile"} {
		ownershipCommand(t, s, "gm", map[string]any{"type": "system.select", "system": system}, 200)
		for attempt := 0; attempt < 8; attempt++ {
			target := 15
			count := 1
			bonus := 3
			mode := "normal"
			if system == "warhammer40k" {
				target = 3
				count = 100
				bonus = 0
			}
			if system == "percentile" {
				target = 100
				bonus = 0
			}
			if strings.HasPrefix(system, "dnd") {
				mode = "advantage"
			}
			r := ownershipCommand(t, s, "player", map[string]any{"type": "system.check", "target": target, "count": count, "bonus": bonus, "mode": mode, "visibility": "all"}, 200)
			p := r["payload"].(map[string]any)
			rolls := p["rolls"].([]any)
			total := p["total"].(float64)
			switch system {
			case "dnd5e", "dnd5e2024":
				if len(rolls) != 2 || total != math.Max(rolls[0].(float64), rolls[1].(float64))+3 {
					t.Fatal("advantage", p)
				}
			case "warhammer40k":
				success := 0
				for _, r := range rolls {
					if r.(float64) >= 3 {
						success++
					}
				}
				if len(rolls) != 100 || int(total) != success {
					t.Fatal("pool", p)
				}
			case "fate":
				sum := 3.0
				for _, r := range rolls {
					f := r.(float64)
					if f < -1 || f > 1 {
						t.Fatal(p)
					}
					sum += f
				}
				if len(rolls) != 4 || sum != total {
					t.Fatal("Fate", p)
				}
			case "percentile":
				if total < 1 || total > 100 || p["outcome"] != "Success" {
					t.Fatal(p)
				}
			case "pf2e":
				if !contains([]string{"Critical failure", "Failure", "Success", "Critical success"}, p["outcome"].(string)) {
					t.Fatal(p)
				}
			}
		}
	}
	ownershipCommand(t, s, "player", map[string]any{"type": "system.check", "target": 20, "bonus": 0, "mode": "advantage"}, 409)
	ownershipCommand(t, s, "gm", map[string]any{"type": "system.select", "system": "warhammer40k"}, 200)
	ownershipCommand(t, s, "player", map[string]any{"type": "system.check", "target": 7, "count": 101, "bonus": 0}, 409)
}

func TestFXAndRulerGeometryPrivacy(t *testing.T) {
	s := ownershipFixture(t)
	cmd := map[string]any{"type": "fx.burst", "x": 1, "y": 1, "to_x": 4, "to_y": 5, "effect": "breath", "element": "fire"}
	ownershipCommand(t, s, "player", cmd, 403)
	ownershipCommand(t, s, "gm", cmd, 200)
	cmd["element"] = "invalid"
	ownershipCommand(t, s, "gm", cmd, 409)
	cmd["element"] = "fire"
	ownershipCommand(t, s, "gm", map[string]any{"type": "table.settings", "width": 20, "height": 14, "scale": 5, "unit": "ft", "metric": "euclidean"}, 200)
	for _, shape := range []string{"line", "circle", "square", "cone", "beam"} {
		r := ownershipCommand(t, s, "player", map[string]any{"type": "measure.share", "x": 1.25, "y": 1.25, "to_x": 4.25, "to_y": 5.25, "shape": shape}, 200)
		p := r["payload"].(map[string]any)
		if p["distance"] != float64(25) || p["shape"] != shape {
			t.Fatal(p)
		}
	}
	ownershipCommand(t, s, "gm", assignment("mayank", "player"), 200)
	ownershipCommand(t, s, "gm", map[string]any{"type": "sight.settings", "enabled": true, "radius": 2}, 200)
	ownershipCommand(t, s, "gm", cmd, 200)
	last := s.db.Sessions["session"].Sequence
	for _, ev := range s.sessionView(s.db.Sessions["session"], "player").Events {
		if ev.Sequence == last {
			t.Fatal("hidden FX exposed to player")
		}
	}
	ownershipCommand(t, s, "player", map[string]any{"type": "measure.share", "x": 1, "y": 1, "to_x": 4, "to_y": 5}, 409)
	ownershipCommand(t, s, "player", map[string]any{"type": "measure.share", "x": 8, "y": 5, "to_x": 9, "to_y": 5, "shape": "circle"}, 200)
	last = s.db.Sessions["session"].Sequence
	for _, ev := range s.sessionView(s.db.Sessions["session"], "other").Events {
		if ev.Sequence == last {
			t.Fatal("other player saw sight-restricted ruler")
		}
	}
	before, _ := json.Marshal(s.db.Sessions["session"])
	ownershipCommand(t, s, "player", map[string]any{"type": "measure.share", "x": 1, "y": 1, "to_x": 200, "to_y": 5}, 409)
	after, _ := json.Marshal(s.db.Sessions["session"])
	if string(before) != string(after) {
		t.Fatal("rejection changed state")
	}
}
