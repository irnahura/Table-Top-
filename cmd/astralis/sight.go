package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
)

// Square-grid sight with blocked cells as opaque walls. A wall's facing cell
// is visible, but a ray cannot pass through it to reveal anything behind it.
func clearRay(w World, a, b Point) bool {
	x, y := a.X, a.Y
	dx := int(math.Abs(float64(b.X - x)))
	dy := -int(math.Abs(float64(b.Y - y)))
	sx, sy := -1, -1
	if x < b.X {
		sx = 1
	}
	if y < b.Y {
		sy = 1
	}
	err := dx + dy
	for {
		if x == b.X && y == b.Y {
			return true
		}
		if (x != a.X || y != a.Y) && opaque(w, Point{x, y}) {
			return false
		}
		e2 := 2 * err
		oldX, oldY := x, y
		if e2 >= dy {
			err += dy
			x += sx
		}
		if e2 <= dx {
			err += dx
			y += sy
		}
		if x != oldX && y != oldY && opaque(w, Point{x, oldY}) && opaque(w, Point{oldX, y}) {
			return false
		}
	}
}
func sightCells(w World, user string) map[Point]bool {
	visible := map[Point]bool{}
	if !w.SightEnabled {
		return visible
	}
	radius := w.SightRadius
	if radius < 1 {
		radius = 8
	}
	for _, a := range w.Actors {
		if a.ControllerID != user || a.Hostile || hiddenLayer(a.Layer) || a.Layer == "map" || a.HP <= 0 {
			continue
		}
		origin := Point{a.X, a.Y}
		reach := max(radius, a.Darkvision)
		for y := max(0, a.Y-reach); y < min(w.Height, a.Y+reach+1); y++ {
			for x := max(0, a.X-reach); x < min(w.Width, a.X+reach+1); x++ {
				p := Point{x, y}
				distance := math.Hypot(float64(x-a.X), float64(y-a.Y))
				if clearRay(w, origin, p) && (distance <= float64(a.Darkvision) || (distance <= float64(radius) && (!w.RequireLight || illuminated(w, p))) || p == origin) {
					visible[p] = true
				}
			}
		}
	}
	return visible
}

// For token-specific events, capture who could see them when they happened.
// A token becoming visible later must not expose its previously hidden history.
func sightAudience(before, after World, v map[string]any, c Campaign, audience []string) []string {
	if !before.SightEnabled && !after.SightEnabled {
		return audience
	}
	refs := []string{}
	for _, key := range []string{"token_id", "character_id", "attacker_id", "target_id"} {
		if target := text(v, key); target != "" {
			refs = append(refs, target)
		}
	}
	for _, a := range after.Actors {
		if _, old := actor(before, a.ID); old == nil {
			refs = append(refs, a.ID)
		}
	}
	if len(refs) == 0 {
		return audience
	}
	out := []string{}
	for user := range c.Members {
		if len(audience) > 0 && !contains(audience, user) {
			continue
		}
		if user == c.OwnerID {
			out = append(out, user)
			continue
		}
		visibleBefore, visibleAfter := sightCells(before, user), sightCells(after, user)
		allowed := true
		for _, ref := range refs {
			if _, a := actor(before, ref); a != nil && before.SightEnabled && !visibleBefore[Point{a.X, a.Y}] {
				allowed = false
			}
			if _, a := actor(after, ref); a != nil && after.SightEnabled && !visibleAfter[Point{a.X, a.Y}] {
				allowed = false
			}
		}
		if allowed {
			out = append(out, user)
		}
	}
	if len(out) == 0 {
		return []string{c.OwnerID}
	}
	return out
}

// Custom uploaded maps are masked before delivery, rather than merely covered
// by CSS. The included sample artwork is also public on the landing page.
func (s *Server) maskedMap(w World, visible map[Point]bool) string {
	if w.MapImage == "blank" {
		return ""
	}
	var key strings.Builder
	fmt.Fprintf(&key, "%s|%d|%d|", hash(w.ImageData), w.Width, w.Height)
	for y := 0; y < w.Height; y++ {
		for x := 0; x < w.Width; x++ {
			if visible[Point{x, y}] {
				fmt.Fprintf(&key, "%d,%d;", x, y)
			}
		}
	}
	cacheKey := hash(key.String())
	if value := s.visionCache[cacheKey]; value != "" {
		return value
	}
	var raw []byte
	if w.ImageData != "" {
		parts := strings.SplitN(w.ImageData, ",", 2)
		if len(parts) != 2 {
			return ""
		}
		raw, _ = base64.StdEncoding.DecodeString(parts[1])
	} else {
		raw, _ = os.ReadFile(filepath.Join(s.web, "assets", "riverside-battlemap.png"))
	}
	source, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return ""
	}
	bounds := source.Bounds()
	masked := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(masked, masked.Bounds(), image.NewUniform(color.RGBA{R: 18, G: 18, B: 18, A: 255}), image.Point{}, draw.Src)
	for p := range visible {
		rect := image.Rect(p.X*bounds.Dx()/w.Width, p.Y*bounds.Dy()/w.Height, (p.X+1)*bounds.Dx()/w.Width, (p.Y+1)*bounds.Dy()/w.Height)
		draw.Draw(masked, rect, source, rect.Min.Add(bounds.Min), draw.Src)
	}
	var encoded bytes.Buffer
	if png.Encode(&encoded, masked) != nil {
		return ""
	}
	value := "data:image/png;base64," + base64.StdEncoding.EncodeToString(encoded.Bytes())
	if s.visionCache == nil || len(s.visionCache) >= 16 {
		s.visionCache = map[string]string{}
	}
	s.visionCache[cacheKey] = value
	return value
}
