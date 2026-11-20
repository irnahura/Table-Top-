package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSmoothDrawingPrecisionAndPrivacy(t *testing.T) {
	s := ownershipFixture(t)
	draw := map[string]any{"type": "draw.add", "shape": "freehand", "layer": "tokens", "color": "#e7c98c", "stroke": 2, "x": 1, "y": 1, "to_x": 3, "to_y": 2, "points": []any{map[string]any{"x": 1, "y": 1}, map[string]any{"x": 3, "y": 2}}, "path": []any{map[string]any{"x": 1.125, "y": 1.5}, map[string]any{"x": 2.375, "y": 1.75}, map[string]any{"x": 3.25, "y": 2.125}}}
	ownershipCommand(t, s, "player", draw, 403)
	ownershipCommand(t, s, "gm", draw, 200)
	d := s.db.Sessions["session"].State.Drawings[0]
	if len(d.Path) != 3 || d.Path[0].X != 1.125 {
		t.Fatal("smooth coordinates lost")
	}
	draw["path"] = []any{map[string]any{"x": -1, "y": 1}, map[string]any{"x": 2, "y": 2}}
	ownershipCommand(t, s, "gm", draw, 409)
	draw["path"] = []any{map[string]any{"x": "NaN", "y": 1}, map[string]any{"x": 2, "y": 2}}
	ownershipCommand(t, s, "gm", draw, 409)
	draw["path"] = []any{map[string]any{"x": 1.125, "y": 1.5}, map[string]any{"x": 2.375, "y": 1.75}}
	draw["layer"] = "gm"
	draw["text"] = "PRIVATE-STROKE"
	ownershipCommand(t, s, "gm", draw, 200)
	data, _ := json.Marshal(s.sessionView(s.db.Sessions["session"], "player"))
	if strings.Contains(string(data), "PRIVATE-STROKE") {
		t.Fatal("private path leaked")
	}
	s = &Server{db: newDB(s.db.file)}
	if s.db.Sessions["session"].State.Drawings[0].Path[0].X != 1.125 {
		t.Fatal("path not persisted")
	}
}
func TestTurnTrackerEditingPermissionsAndBudgets(t *testing.T) {
	s := ownershipFixture(t)
	ss := s.db.Sessions["session"]
	ss.State.MovementUsed = 3
	ss.State.AttackUsed = true
	s.db.Sessions["session"] = ss
	manage := func(action, token string, score int) map[string]any {
		return map[string]any{"type": "initiative.manage", "action": action, "token_id": token, "score": score}
	}
	ownershipCommand(t, s, "player", manage("score", "mayank", 10), 403)
	ownershipCommand(t, s, "gm", manage("score", "mayank", 10), 200)
	ownershipCommand(t, s, "gm", manage("score", "lira", 18), 200)
	ownershipCommand(t, s, "gm", manage("score", "bandit", 20), 409)
	ownershipCommand(t, s, "gm", manage("sort", "", 0), 200)
	if s.db.Sessions["session"].State.Initiative[0].TokenID != "lira" {
		t.Fatal("sort failed")
	}
	ownershipCommand(t, s, "gm", manage("up", "mayank", 0), 200)
	state := s.db.Sessions["session"].State
	if state.Initiative[0].TokenID != "mayank" || state.MovementUsed != 3 || !state.AttackUsed {
		t.Fatal("reorder changed action budgets")
	}
	before := map[string]any{"type": "initiative.manage", "action": "before", "token_id": "lira", "before_id": "mayank"}
	ownershipCommand(t, s, "player", before, 403)
	ownershipCommand(t, s, "gm", before, 200)
	state = s.db.Sessions["session"].State
	if state.Initiative[0].TokenID != "lira" || state.MovementUsed != 3 || !state.AttackUsed {
		t.Fatal("drag reorder changed budgets")
	}
	before["before_id"] = "missing"
	ownershipCommand(t, s, "gm", before, 409)
	ownershipCommand(t, s, "gm", map[string]any{"type": "initiative.manage", "action": "visibility", "visible": true}, 200)
	s = &Server{db: newDB(s.db.file)}
	if !s.db.Sessions["session"].State.TrackerVisible {
		t.Fatal("tracker visibility not persisted")
	}
	ownershipCommand(t, s, "gm", manage("remove", "lira", 0), 200)
	ownershipCommand(t, s, "gm", manage("clear", "", 0), 200)
}
