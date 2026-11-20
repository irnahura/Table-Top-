package main

import (
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestTableToolsPersistenceAndPermissions(t *testing.T) {
	file := filepath.Join(t.TempDir(), "table.json")
	s := &Server{db: newDB(file)}
	token := ""
	request := func(method, path string, payload any, status int) map[string]any {
		t.Helper()
		b, _ := json.Marshal(payload)
		r := httptest.NewRequest(method, "/api/v1"+path, strings.NewReader(string(b)))
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		s.ServeHTTP(w, r)
		if w.Code != status {
			t.Fatalf("%s %s: %d want %d: %s", method, path, w.Code, status, w.Body.String())
		}
		var out map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &out)
		return out
	}
	demo := request("POST", "/auth/demo", map[string]any{}, 201)
	token = demo["token"].(string)
	ownerToken := token
	ss := demo["snapshot"].(map[string]any)
	sid := ss["id"].(string)
	cid := demo["campaign"].(map[string]any)["id"].(string)
	command := func(v map[string]any, status int) map[string]any {
		return request("POST", "/sessions/"+sid+"/commands", map[string]any{"payload": v}, status)
	}
	command(map[string]any{"type": "map.block", "x": 6, "y": 8}, 200)
	command(map[string]any{"type": "map.fog", "x": 2, "y": 9}, 200)
	command(map[string]any{"type": "map.draw", "x": 1, "y": 8, "to_x": 4, "to_y": 10, "shape": "rectangle"}, 200)
	command(map[string]any{"type": "map.draw", "x": -1, "y": 8, "to_x": 4, "to_y": 10, "shape": "line"}, 409)
	command(map[string]any{"type": "notes.save", "notes": "Inspect the old bridge."}, 200)
	command(map[string]any{"type": "token.add", "name": "Innkeeper", "x": 5, "y": 5, "hp": 20, "hostile": false}, 200)
	command(map[string]any{"type": "token.update", "token_id": "lira", "name": "Lira the scout", "hp": 42}, 200)
	command(map[string]any{"type": "token.update", "token_id": "lira", "name": "Lira", "hp": -1}, 409)
	command(map[string]any{"type": "token.place", "token_id": "lira", "x": 6, "y": 8}, 409) // blocked
	command(map[string]any{"type": "token.place", "token_id": "lira", "x": 5, "y": 8}, 200)
	roll := command(map[string]any{"type": "dice.roll", "count": 3, "sides": 6, "bonus": 2}, 200)["payload"].(map[string]any)
	sum := 2.
	for _, x := range roll["rolls"].([]any) {
		n := x.(float64)
		if n < 1 || n > 6 {
			t.Fatal("invalid die")
		}
		sum += n
	}
	if sum != roll["total"] {
		t.Fatal("roll total mismatch")
	}
	command(map[string]any{"type": "dice.roll", "count": 10000, "sides": 6, "bonus": 0}, 409)
	command(map[string]any{"type": "map.upload", "image_data": "data:image/png;base64,not-an-image"}, 409)
	invite := request("POST", "/campaigns/"+cid+"/invitations", map[string]any{}, 201)["token"].(string)
	// A second sandbox cannot inspect or mutate the first one.
	second := request("POST", "/auth/demo", map[string]any{}, 201)
	token = second["token"].(string)
	request("GET", "/sessions/"+sid+"/snapshot", nil, 403)
	request("GET", "/sessions/"+sid+"/events", nil, 403)
	command(map[string]any{"type": "notes.save", "notes": "unauthorized"}, 403)
	request("POST", "/maps", map[string]any{"campaign_id": cid, "name": "foreign"}, 403)
	request("POST", "/invitations/redeem", map[string]any{"token": invite}, 200)
	request("GET", "/sessions/"+sid+"/snapshot", nil, 200)
	command(map[string]any{"type": "notes.save", "notes": "player edit"}, 403)
	command(map[string]any{"type": "map.block", "x": 3, "y": 9}, 403)
	command(map[string]any{"type": "chat.send", "message": "Ready for the next encounter."}, 200)
	command(map[string]any{"type": "map.ping", "x": 3, "y": 9}, 200)
	command(map[string]any{"type": "dice.roll", "count": 1, "sides": 20, "bonus": 0}, 200)
	// Restart the DB and verify authentication and every saved map field.
	s = &Server{db: newDB(file)}
	token = ownerToken
	snapshot := request("GET", "/sessions/"+sid+"/snapshot", nil, 200)
	world := snapshot["state"].(map[string]any)
	if world["notes"] != "Inspect the old bridge." || len(world["fog"].([]any)) != 1 || len(world["drawings"].([]any)) != 1 || len(world["actors"].([]any)) != 5 {
		t.Fatal("saved tools missing after restart")
	}
	request("GET", "/sessions", nil, 200)
	request("POST", "/auth/logout", map[string]any{}, 200)
	request("GET", "/sessions/"+sid+"/snapshot", nil, 401)
}

func TestPasswordsSurviveRestart(t *testing.T) {
	file := filepath.Join(t.TempDir(), "auth.json")
	s := &Server{db: newDB(file)}
	credentials := `{"email":"persist@example.test","password":"working-password","display_name":"Persist"}`
	request := func(path string, status int) {
		t.Helper()
		r := httptest.NewRequest("POST", "/api/v1/auth/"+path, strings.NewReader(credentials))
		w := httptest.NewRecorder()
		s.ServeHTTP(w, r)
		if w.Code != status {
			t.Fatalf("%s: %d %s", path, w.Code, w.Body.String())
		}
		if strings.Contains(w.Body.String(), "password_hash") || strings.Contains(w.Body.String(), "passwords") {
			t.Fatal("password exposed")
		}
	}
	request("register", 201)
	request("register", 409)
	s = &Server{db: newDB(file)}
	request("login", 200)
}

func TestTurnBudgetAndRejectedMutation(t *testing.T) {
	w := World{Width: 20, Height: 14, ActiveActorID: "a", Round: 1, Actors: []Actor{{ID: "a", X: 0, Y: 0, HP: 10, Movement: 2}, {ID: "b", X: 4, Y: 4, HP: 10, Movement: 2}}}
	moved, _, _, _, err := rule(w, map[string]any{"type": "character.move", "character_id": "a", "target": map[string]any{"x": float64(2), "y": float64(0)}})
	if err != nil {
		t.Fatal(err)
	}
	if w.Actors[0].X != 0 {
		t.Fatal("rule mutated original")
	}
	_, _, _, _, err = rule(moved, map[string]any{"type": "character.move", "character_id": "a", "target": map[string]any{"x": float64(3), "y": float64(0)}})
	if err == nil {
		t.Fatal("movement budget not enforced")
	}
	next, _, _, _, err := rule(moved, map[string]any{"type": "turn.end", "character_id": "a"})
	if err != nil || next.ActiveActorID != "b" || next.Round != 1 || next.MovementUsed != 0 {
		t.Fatal("party turn did not advance")
	}
	next, _, _, _, err = rule(next, map[string]any{"type": "turn.end", "character_id": "b"})
	if err != nil || next.ActiveActorID != "a" || next.Round != 2 {
		t.Fatal("round did not wrap")
	}
}
