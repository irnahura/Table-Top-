package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
)

func TestSightMasksPixelsTokensAndHistoricalEvents(t *testing.T) {
	s := ownershipFixture(t)
	art := image.NewRGBA(image.Rect(0, 0, 20, 14))
	for y := 0; y < 14; y++ {
		for x := 0; x < 20; x++ {
			art.Set(x, y, color.White)
		}
	}
	var out bytes.Buffer
	_ = png.Encode(&out, art)
	ss := s.db.Sessions["session"]
	ss.State.ImageData = "data:image/png;base64," + base64.StdEncoding.EncodeToString(out.Bytes())
	ss.State.MapImage = "custom"
	ss.State.Blocked = []Point{{9, 5}}
	s.db.Sessions[ss.ID] = ss
	ownershipCommand(t, s, "gm", assignment("mayank", "player"), 200)
	ownershipCommand(t, s, "gm", map[string]any{"type": "sight.settings", "enabled": true, "radius": 3}, 200)
	ownershipCommand(t, s, "gm", map[string]any{"type": "token.update", "token_id": "bandit", "name": "HIDDEN-HISTORY", "hp": 30}, 200)
	var view Session
	_ = json.Unmarshal([]byte(tableRequest(t, s, "player", "/sessions/session/snapshot")), &view)
	if len(view.State.Actors) != 1 || view.State.Actors[0].ID != "mayank" {
		t.Fatal("tokens outside sight were delivered", view.State.Actors)
	}
	if len(view.State.VisibleCells) == 0 {
		t.Fatal("missing sight mask")
	}
	for _, ev := range view.Events {
		raw, _ := json.Marshal(ev)
		if strings.Contains(string(raw), "HIDDEN-HISTORY") {
			t.Fatal("hidden event leaked")
		}
	}
	encoded := strings.SplitN(view.State.ImageData, ",", 2)
	raw, err := base64.StdEncoding.DecodeString(encoded[1])
	if err != nil {
		t.Fatal(err)
	}
	masked, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	r, _, _, _ := masked.At(10, 5).RGBA()
	if r > 10000 {
		t.Fatal("hidden custom map pixel was not masked")
	}
	r, _, _, _ = masked.At(8, 5).RGBA()
	if r < 60000 {
		t.Fatal("visible custom map pixel was obscured")
	}
	ownershipCommand(t, s, "player", map[string]any{"type": "combat.attack", "attacker_id": "mayank", "target_id": "bandit"}, 403)
	ownershipCommand(t, s, "gm", map[string]any{"type": "map.block", "x": 9, "y": 5}, 200)
	_ = json.Unmarshal([]byte(tableRequest(t, s, "player", "/sessions/session/snapshot")), &view)
	if len(view.State.Actors) != 2 {
		t.Fatal("clearing wall did not restore line of sight")
	}
	for _, ev := range view.Events {
		raw, _ := json.Marshal(ev)
		if strings.Contains(string(raw), "HIDDEN-HISTORY") {
			t.Fatal("revealing a token exposed its hidden history")
		}
	}
	view = Session{}
	_ = json.Unmarshal([]byte(tableRequest(t, s, "other", "/sessions/session/snapshot")), &view)
	if len(view.State.Actors) != 0 || len(view.State.VisibleCells) != 0 {
		t.Fatal("unassigned observer saw tokens")
	}
}

func TestSightDoesNotPassThroughDiagonalWallCorners(t *testing.T) {
	w := World{Width: 8, Height: 8, Blocked: []Point{{2, 1}, {1, 2}}}
	if clearRay(w, Point{1, 1}, Point{3, 3}) {
		t.Fatal("sight leaked through a closed diagonal corner")
	}
	if !clearRay(w, Point{1, 1}, Point{2, 1}) {
		t.Fatal("facing wall cell should be visible")
	}
}
