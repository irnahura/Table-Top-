package main

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
)

func TestDoorsLightAndDarkvision(t *testing.T) {
	s := ownershipFixture(t)
	ownershipCommand(t, s, "gm", assignment("mayank", "player"), 200)
	ownershipCommand(t, s, "gm", map[string]any{"type": "sight.settings", "enabled": true, "radius": 8}, 200)
	door := map[string]any{"type": "door.save", "name": "Gate", "x": 9, "y": 5, "open": false, "locked": false, "secret": false, "window": false}
	ownershipCommand(t, s, "player", door, 403)
	ownershipCommand(t, s, "gm", door, 200)
	w := s.db.Sessions["session"].State
	did := w.Doors[0].ID
	if !blocked(w, Point{9, 5}) || clearRay(w, Point{8, 5}, Point{10, 5}) {
		t.Fatal("closed door does not block movement/sight")
	}
	ownershipCommand(t, s, "player", map[string]any{"type": "door.toggle", "door_id": did}, 200)
	w = s.db.Sessions["session"].State
	if blocked(w, Point{9, 5}) || !clearRay(w, Point{8, 5}, Point{10, 5}) {
		t.Fatal("opened door still blocks")
	}
	door["door_id"] = did
	door["window"] = true
	ownershipCommand(t, s, "gm", door, 200)
	w = s.db.Sessions["session"].State
	if !blocked(w, Point{9, 5}) || !clearRay(w, Point{8, 5}, Point{10, 5}) {
		t.Fatal("window semantics incorrect")
	}
	door["secret"] = true
	door["locked"] = true
	ownershipCommand(t, s, "gm", door, 200)
	if len(s.sessionView(s.db.Sessions["session"], "player").State.Doors) != 0 {
		t.Fatal("secret door leaked")
	}
	ownershipCommand(t, s, "player", map[string]any{"type": "door.toggle", "door_id": did}, 409)
	ownershipCommand(t, s, "gm", map[string]any{"type": "lighting.settings", "require_light": true}, 200)
	w = s.db.Sessions["session"].State
	if sightCells(w, "player")[Point{10, 5}] {
		t.Fatal("dark room visible without light")
	}
	light := map[string]any{"type": "light.save", "name": "Torch", "x": 8, "y": 5, "radius": 4, "color": "#ffbb55", "enabled": true}
	ownershipCommand(t, s, "gm", light, 200)
	w = s.db.Sessions["session"].State
	if !sightCells(w, "player")[Point{10, 5}] {
		t.Fatal("light did not illuminate through window")
	}
	ownershipCommand(t, s, "gm", map[string]any{"type": "light.delete", "light_id": w.Lights[0].ID}, 200)
	ownershipCommand(t, s, "gm", map[string]any{"type": "token.vision", "token_id": "mayank", "darkvision": 4}, 200)
	w = s.db.Sessions["session"].State
	if !sightCells(w, "player")[Point{10, 5}] {
		t.Fatal("darkvision did not see darkness")
	}
	ownershipCommand(t, s, "gm", map[string]any{"type": "table.settings", "width": 9, "height": 14, "scale": 1, "unit": "spaces", "metric": "manhattan"}, 409)
	s = &Server{db: newDB(s.db.file)}
	w = s.db.Sessions["session"].State
	if !w.RequireLight || len(w.Doors) != 1 || w.Actors[0].Darkvision != 4 {
		t.Fatal("environment not persisted")
	}
}

func TestMapCropPixelsAndRejection(t *testing.T) {
	s := ownershipFixture(t)
	img := image.NewRGBA(image.Rect(0, 0, 4, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{R: uint8(40 * x), G: uint8(80 * y), A: 255})
		}
	}
	var b bytes.Buffer
	_ = png.Encode(&b, img)
	ownershipCommand(t, s, "gm", map[string]any{"type": "map.upload", "image_data": "data:image/png;base64," + base64.StdEncoding.EncodeToString(b.Bytes())}, 200)
	crop := map[string]any{"type": "map.calibrate", "left": 50, "top": 0, "right": 100, "bottom": 100, "rotation": 90}
	ownershipCommand(t, s, "player", crop, 403)
	ownershipCommand(t, s, "gm", crop, 200)
	raw, _ := base64.StdEncoding.DecodeString(strings.SplitN(s.db.Sessions["session"].State.ImageData, ",", 2)[1])
	result, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil || result.Bounds().Dx() != 2 || result.Bounds().Dy() != 2 {
		t.Fatal("wrong crop size")
	}
	r, g, _, _ := result.At(0, 0).RGBA()
	if r != 80*257 || g != 80*257 {
		t.Fatalf("wrong rotated crop pixel %d,%d", r, g)
	}
	crop["left"] = 100
	ownershipCommand(t, s, "gm", crop, 409)
}
