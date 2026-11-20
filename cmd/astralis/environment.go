package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"strings"
)

type Door struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Open   bool   `json:"open"`
	Locked bool   `json:"locked"`
	Secret bool   `json:"secret"`
	Window bool   `json:"window"`
}
type Light struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
	Radius  int    `json:"radius"`
	Color   string `json:"color"`
	Enabled bool   `json:"enabled"`
}
type Environment struct {
	Doors        []Door  `json:"doors,omitempty"`
	Lights       []Light `json:"lights,omitempty"`
	RequireLight bool    `json:"require_light,omitempty"`
}

func opaque(w World, p Point) bool {
	for _, b := range w.Blocked {
		if b == p {
			return true
		}
	}
	for _, d := range w.Doors {
		if d.X == p.X && d.Y == p.Y && !d.Open && !d.Window {
			return true
		}
	}
	return false
}
func illuminated(w World, p Point) bool {
	for _, l := range w.Lights {
		if l.Enabled && math.Hypot(float64(l.X-p.X), float64(l.Y-p.Y)) <= float64(l.Radius) && clearRay(w, Point{l.X, l.Y}, p) {
			return true
		}
	}
	return false
}
func environmentView(w World, user string, gm bool) Environment {
	if gm {
		return w.Environment
	}
	out := Environment{RequireLight: w.RequireLight}
	visible := sightCells(w, user)
	for _, d := range w.Doors {
		if !d.Secret && (!w.SightEnabled || visible[Point{d.X, d.Y}]) {
			out.Doors = append(out.Doors, d)
		}
	}
	for _, l := range w.Lights {
		if !w.SightEnabled || visible[Point{l.X, l.Y}] {
			out.Lights = append(out.Lights, l)
		}
	}
	return out
}
func environmentCommand(kind string) bool {
	switch kind {
	case "door.save", "door.delete", "door.toggle", "light.save", "light.delete", "lighting.settings", "token.vision", "map.calibrate":
		return true
	}
	return false
}
func (s *Server) environmentCommand(original World, v map[string]any, c Campaign, user string) (World, string, string, any, []string, bool, error) {
	fail := func(msg string) (World, string, string, any, []string, bool, error) {
		return original, "", "", nil, nil, false, errors.New(msg)
	}
	w := World{}
	b, _ := json.Marshal(original)
	_ = json.Unmarshal(b, &w)
	kind := text(v, "type")
	gm := c.OwnerID == user
	if !gm && kind != "door.toggle" {
		return fail("PERMISSION_DENIED: only the GM configures lighting and map setup")
	}
	cell := func() (int, int, bool) {
		x, a := integer(v, "x", 0, w.Width-1)
		y, b := integer(v, "y", 0, w.Height-1)
		return x, y, a && b
	}
	if strings.HasPrefix(kind, "door.") {
		index := -1
		for i, d := range w.Doors {
			if d.ID == text(v, "door_id") {
				index = i
			}
		}
		if kind != "door.save" && index < 0 {
			return fail("DOOR_NOT_FOUND")
		}
		switch kind {
		case "door.save":
			if text(v, "door_id") != "" && index < 0 {
				return fail("DOOR_NOT_FOUND")
			}
			x, y, ok := cell()
			opened, a := v["open"].(bool)
			locked, b := v["locked"].(bool)
			secret, d := v["secret"].(bool)
			window, e := v["window"].(bool)
			if !ok || !a || !b || !d || !e || text(v, "name") == "" || len(text(v, "name")) > 80 || opened && locked {
				return fail("INVALID_DOOR: choose a cell, name and compatible open/locked states")
			}
			for i, door := range w.Doors {
				if i != index && door.X == x && door.Y == y {
					return fail("A door already occupies that cell")
				}
			}
			if !opened {
				for _, actor := range w.Actors {
					if x >= actor.X && x < actor.X+max(1, actor.Size) && y >= actor.Y && y < actor.Y+max(1, actor.Size) {
						return fail("Move tokens out of the doorway before closing it")
					}
				}
			}
			door := Door{ID: id(), Name: text(v, "name"), X: x, Y: y, Open: opened, Locked: locked, Secret: secret, Window: window}
			if index >= 0 {
				door.ID = w.Doors[index].ID
				w.Doors[index] = door
			} else {
				if len(w.Doors) >= 100 {
					return fail("Maximum 100 doors")
				}
				w.Doors = append(w.Doors, door)
			}
		case "door.delete":
			w.Doors = append(w.Doors[:index], w.Doors[index+1:]...)
		case "door.toggle":
			door := &w.Doors[index]
			if door.Locked {
				return fail("This door is locked. The GM must unlock it.")
			}
			if !gm {
				if door.Secret || (w.SightEnabled && !sightCells(w, user)[Point{door.X, door.Y}]) {
					return fail("PERMISSION_DENIED: door is not visible")
				}
				near := false
				for _, a := range w.Actors {
					if a.ControllerID == user && !a.Hostile && a.HP > 0 && a.Layer != "gm" && a.Layer != "map" && dist(Point{a.X, a.Y}, Point{door.X, door.Y}) <= 1 {
						near = true
					}
				}
				if !near {
					return fail("Move an assigned living character next to the door first")
				}
			}
			if door.Open {
				for _, a := range w.Actors {
					if door.X >= a.X && door.X < a.X+max(1, a.Size) && door.Y >= a.Y && door.Y < a.Y+max(1, a.Size) {
						return fail("Move tokens out of the doorway before closing it")
					}
				}
			}
			door.Open = !door.Open
		}
	} else if strings.HasPrefix(kind, "light.") {
		index := -1
		for i, l := range w.Lights {
			if l.ID == text(v, "light_id") {
				index = i
			}
		}
		if kind == "light.delete" {
			if index < 0 {
				return fail("LIGHT_NOT_FOUND")
			}
			w.Lights = append(w.Lights[:index], w.Lights[index+1:]...)
		} else {
			if text(v, "light_id") != "" && index < 0 {
				return fail("LIGHT_NOT_FOUND")
			}
			x, y, ok := cell()
			radius, a := integer(v, "radius", 1, 30)
			enabled, b := v["enabled"].(bool)
			name := text(v, "name")
			color := text(v, "color")
			if !ok || !a || !b || name == "" || len(name) > 80 || !colorPattern.MatchString(color) {
				return fail("INVALID_LIGHT: choose a name, map cell, color and radius 1–30")
			}
			light := Light{ID: id(), Name: name, X: x, Y: y, Radius: radius, Color: color, Enabled: enabled}
			if index >= 0 {
				light.ID = w.Lights[index].ID
				w.Lights[index] = light
			} else {
				if len(w.Lights) >= 50 {
					return fail("Maximum 50 lights")
				}
				w.Lights = append(w.Lights, light)
			}
		}
	} else if kind == "lighting.settings" {
		enabled, ok := v["require_light"].(bool)
		if !ok {
			return fail("INVALID_LIGHTING")
		}
		w.RequireLight = enabled
	} else if kind == "token.vision" {
		radius, ok := integer(v, "darkvision", 0, 30)
		_, a := actor(w, text(v, "token_id"))
		if !ok || a == nil {
			return fail("INVALID_VISION: choose a token and darkvision from 0–30")
		}
		a.Darkvision = radius
	} else if kind == "map.calibrate" {
		if w.ImageData == "" {
			return fail("Upload a custom map before cropping it")
		}
		left, a := integer(v, "left", 0, 99)
		top, b := integer(v, "top", 0, 99)
		right, d := integer(v, "right", 1, 100)
		bottom, e := integer(v, "bottom", 1, 100)
		rotation, f := integer(v, "rotation", 0, 270)
		if !a || !b || !d || !e || !f || left >= right || top >= bottom || rotation%90 != 0 {
			return fail("INVALID_CROP: use percentage bounds and rotation 0/90/180/270")
		}
		parts := strings.SplitN(w.ImageData, ",", 2)
		if len(parts) != 2 {
			return fail("INVALID_IMAGE")
		}
		raw, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			return fail("INVALID_IMAGE")
		}
		source, _, err := image.Decode(bytes.NewReader(raw))
		if err != nil {
			return fail("INVALID_IMAGE")
		}
		bounds := source.Bounds()
		crop := image.Rect(left*bounds.Dx()/100, top*bounds.Dy()/100, right*bounds.Dx()/100, bottom*bounds.Dy()/100).Add(bounds.Min)
		if crop.Empty() {
			return fail("Crop is smaller than a pixel")
		}
		width, height := crop.Dx(), crop.Dy()
		if rotation == 90 || rotation == 270 {
			width, height = height, width
		}
		ratio := math.Min(1, 2048/float64(max(width, height)))
		outW, outH := max(1, int(float64(width)*ratio)), max(1, int(float64(height)*ratio))
		out := image.NewRGBA(image.Rect(0, 0, outW, outH))
		draw.Draw(out, out.Bounds(), image.NewUniform(color.Black), image.Point{}, draw.Src)
		for y := 0; y < outH; y++ {
			for x := 0; x < outW; x++ {
				rx, ry := x*width/outW, y*height/outH
				var sx, sy int
				switch rotation {
				case 90:
					sx, sy = ry, crop.Dy()-1-rx
				case 180:
					sx, sy = crop.Dx()-1-rx, crop.Dy()-1-ry
				case 270:
					sx, sy = crop.Dx()-1-ry, rx
				default:
					sx, sy = rx, ry
				}
				out.Set(x, y, source.At(crop.Min.X+sx, crop.Min.Y+sy))
			}
		}
		var encoded bytes.Buffer
		if png.Encode(&encoded, out) != nil {
			return fail("Could not encode cropped map")
		}
		if encoded.Len() > 5<<20 {
			return fail("Cropped map exceeds 5 MB; choose a smaller crop")
		}
		w.ImageData = "data:image/png;base64," + base64.StdEncoding.EncodeToString(encoded.Bytes())
		w.MapImage = "custom"
	}
	// Configuration and secret door details are never included in public events.
	return w, kind, user, map[string]any{"by": s.db.Users[user].DisplayName}, nil, true, nil
}
