package main

import (
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestPointJSONCoordinates(t *testing.T) {
	encoded, err := json.Marshal(Point{X: 7, Y: 4})
	if err != nil {
		t.Fatal(err)
	}
	var coordinates map[string]int
	if err := json.Unmarshal(encoded, &coordinates); err != nil {
		t.Fatal(err)
	}
	if len(coordinates) != 2 || coordinates["x"] != 7 || coordinates["y"] != 4 {
		t.Fatalf("map coordinates were lost: %s", encoded)
	}
}

func TestFrontendGameJourney(t *testing.T) {
	s := &Server{db: newDB(filepath.Join(t.TempDir(), "state.json"))}
	token := ""
	request := func(method, path, payload string, status int) map[string]any {
		t.Helper()
		r := httptest.NewRequest(method, "/api/v1"+path, strings.NewReader(payload))
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		s.ServeHTTP(w, r)
		if w.Code != status {
			t.Fatalf("%s %s: got %d, want %d: %s", method, path, w.Code, status, w.Body.String())
		}
		var data map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &data); err != nil {
			t.Fatal(err)
		}
		return data
	}
	credentials := `{"display_name":"Design QA","email":"design@example.test","password":"test-password"}`
	request("POST", "/auth/register", credentials, 201)
	login := request("POST", "/auth/login", credentials, 200)
	token = login["token"].(string)
	c := request("POST", "/campaigns", `{"name":"QA world","description":"Isolated test"}`, 201)["campaign"].(map[string]any)
	cid := c["id"].(string)
	request("GET", "/campaigns", "", 200)
	m := request("POST", "/maps", `{"campaign_id":"`+cid+`","name":"Riverside Village"}`, 201)["map"].(map[string]any)
	snapshot := request("POST", "/campaigns/"+cid+"/sessions", `{"map_id":"`+m["id"].(string)+`"}`, 201)["snapshot"].(map[string]any)
	sid := snapshot["id"].(string)
	request("GET", "/sessions/"+sid+"/snapshot", "", 200)
	moved := request("POST", "/sessions/"+sid+"/commands", `{"payload":{"type":"character.move","character_id":"mayank","target":{"x":3,"y":2}}}`, 200)
	if moved["snapshot"].(map[string]any)["sequence"] != float64(1) {
		t.Fatal("accepted move did not advance snapshot sequence")
	}
	request("POST", "/sessions/"+sid+"/commands", `{"payload":{"type":"character.move","character_id":"mayank","target":{"x":7,"y":4}}}`, 409)
	request("POST", "/sessions/"+sid+"/commands", `{"payload":{"type":"turn.end","character_id":"mayank"}}`, 200)
}
