package main

import (
	"bytes"
	"crypto/pbkdf2"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	rand "math/rand/v2"
	"net/http"
	"sort"
	"strings"
	"time"
)

type Drawing struct {
	Path   []DrawPoint `json:"path,omitempty"`
	Layer  string      `json:"layer,omitempty"`
	Text   string      `json:"text,omitempty"`
	Points []Point     `json:"points,omitempty"`
	Color  string      `json:"color,omitempty"`
	Stroke int         `json:"stroke,omitempty"`
	ID     string      `json:"id"`
	From   Point       `json:"from"`
	To     Point       `json:"to"`
	Shape  string      `json:"shape"`
}
type DrawPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func text(v map[string]any, k string) string { s, _ := v[k].(string); return strings.TrimSpace(s) }
func integer(v map[string]any, k string, min, max int) (int, bool) {
	n, ok := v[k].(float64)
	return int(n), ok && !math.IsNaN(n) && n == math.Trunc(n) && n >= float64(min) && n <= float64(max)
}
func passwordHash(password string) string {
	salt := id()
	key, _ := pbkdf2.Key(sha256.New, password, []byte(salt), 600000, 32)
	return salt + ":" + hex.EncodeToString(key)
}
func verifyPassword(stored, password string) bool {
	parts := strings.Split(stored, ":")
	if len(parts) != 2 || len(password) > 128 {
		return false
	}
	key, _ := pbkdf2.Key(sha256.New, password, []byte(parts[0]), 600000, 32)
	return subtle.ConstantTimeCompare([]byte(parts[1]), []byte(hex.EncodeToString(key))) == 1
}

// Each demo entry gets its own saved sandbox, without sharing another visitor's table.
func (s *Server) demoAuth(w http.ResponseWriter, r *http.Request) bool {
	if r.URL.Path != "/api/v1/auth/demo" {
		return false
	}
	if r.Method != "POST" {
		s.json(w, 405, map[string]string{"error": "METHOD_NOT_ALLOWED"})
		return true
	}
	s.db.mu.Lock()
	defer s.db.mu.Unlock()
	u := User{ID: id(), DisplayName: "Demo Game Master", Email: "demo-" + id()[:8] + "@astralis.local"}
	s.db.Users[u.ID] = u
	c := Campaign{ID: id(), OwnerID: u.ID, Name: "Riverside • Demo table", Description: "Your own saved sandbox. Try map tools, dice, tokens and notes.", Members: map[string]string{u.ID: "owner"}, CreatedAt: time.Now().UTC()}
	s.db.Campaigns[c.ID] = c
	m := GameMap{ID: id(), OwnerID: u.ID, CampaignID: c.ID, Name: "Riverside village", Width: 20, Height: 14}
	s.db.Maps[m.ID] = m
	world := World{Width: 20, Height: 14, Blocked: []Point{{7, 4}, {7, 5}, {7, 6}, {8, 6}, {9, 6}}, Actors: []Actor{{ID: "mayank", Name: "Mayank", X: 2, Y: 2, HP: 82, MaxHP: 100, Movement: 6, AttackBonus: 5, Damage: 8}, {ID: "lira", Name: "Lira", X: 4, Y: 3, HP: 100, MaxHP: 100, Movement: 6, AttackBonus: 4, Damage: 6}, {ID: "bandit-scout", Name: "Bandit Scout", X: 10, Y: 5, HP: 30, MaxHP: 30, Movement: 6, AttackBonus: 3, Damage: 4, Hostile: true}, {ID: "bandit-archer", Name: "Bandit Archer", X: 13, Y: 4, HP: 24, MaxHP: 24, Movement: 6, AttackBonus: 3, Damage: 4, Hostile: true}}, ActiveActorID: "mayank", Round: 1}
	ss := Session{ID: id(), CampaignID: c.ID, MapID: m.ID, Status: "active", Version: 1, State: world, CreatedAt: time.Now().UTC()}
	s.db.Sessions[ss.ID] = ss
	t := id()
	s.db.Tokens[t] = u.ID
	s.db.TokenExpiry[t] = time.Now().Add(7 * 24 * time.Hour)
	s.db.save()
	s.json(w, 201, map[string]any{"user": u, "token": t, "campaign": c, "snapshot": ss})
	return true
}

// Called under the database lock.
func (s *Server) extraAPI(w http.ResponseWriter, r *http.Request, u string, v map[string]any) bool {
	p := r.URL.Path
	if s.campaignMembers(w, r, u) {
		return true
	}
	if p == "/api/v1/auth/me" && r.Method == "GET" {
		s.json(w, 200, map[string]any{"user": s.db.Users[u]})
		return true
	}
	if p == "/api/v1/auth/logout" && r.Method == "POST" {
		delete(s.db.Tokens, strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		delete(s.db.TokenExpiry, strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		s.db.save()
		s.json(w, 200, map[string]bool{"ok": true})
		return true
	}
	if p == "/api/v1/sessions" && r.Method == "GET" {
		out := []Session{}
		for _, ss := range s.db.Sessions {
			if member(s.db.Campaigns[ss.CampaignID], u) {
				out = append(out, s.sessionView(ss, u))
			}
		}
		sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
		s.json(w, 200, map[string]any{"sessions": out})
		return true
	}
	if p == "/api/v1/maps" && r.Method == "GET" {
		out := []GameMap{}
		for _, m := range s.db.Maps {
			if member(s.db.Campaigns[m.CampaignID], u) {
				out = append(out, m)
			}
		}
		s.json(w, 200, map[string]any{"maps": out})
		return true
	}
	return false
}
func pathDistance(w World, start, end Point, actorID string) int {
	for budget := 0; budget <= w.Width*w.Height; budget++ {
		if reachable(w, start, end, budget, actorID) {
			return budget
		}
	}
	return w.Width*w.Height + 1
}
func reachable(w World, start, end Point, budget int, actorID string) bool {
	queue := []Point{start}
	steps := map[Point]int{start: 0}
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		if p == end {
			return true
		}
		if steps[p] >= budget {
			continue
		}
		for _, d := range []Point{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			n := Point{p.X + d.X, p.Y + d.Y}
			if _, seen := steps[n]; seen || !canOccupy(w, n, actorID, 0) {
				continue
			}
			{
				steps[n] = steps[p] + 1
				queue = append(queue, n)
			}
		}
	}
	return false
}
func isToolCommand(kind string) bool {
	return strings.HasPrefix(kind, "map.") || strings.HasPrefix(kind, "token.") || kind == "dice.roll" || kind == "notes.save" || kind == "chat.send"
}
func toolRule(w World, v map[string]any, gm bool, name string) (World, string, string, any, error) {
	kind := text(v, "type")
	fail := func(msg string) (World, string, string, any, error) { return w, "", "", nil, errors.New(msg) }
	if kind != "dice.roll" && kind != "chat.send" && kind != "map.ping" && !gm {
		return fail("PERMISSION_DENIED: only the game master can edit the scene")
	}
	w.Actors = append([]Actor(nil), w.Actors...)
	w.Blocked = append([]Point(nil), w.Blocked...)
	w.Fog = append([]Point(nil), w.Fog...)
	w.Drawings = append([]Drawing(nil), w.Drawings...)
	data := map[string]any{"by": name}
	event := kind
	cell := func(prefix string) (Point, bool) {
		x, a := integer(v, prefix+"x", 0, w.Width-1)
		y, b := integer(v, prefix+"y", 0, w.Height-1)
		return Point{x, y}, a && b
	}
	switch kind {
	case "map.block", "map.fog", "map.ping":
		p, ok := cell("")
		if !ok {
			return fail("INVALID_CELL: choose a cell inside the map")
		}
		data["point"] = p
		if kind == "map.ping" {
			break
		}
		list := &w.Blocked
		if kind == "map.fog" {
			list = &w.Fog
		}
		found := -1
		for i, x := range *list {
			if x == p {
				found = i
				break
			}
		}
		if found >= 0 {
			*list = append((*list)[:found], (*list)[found+1:]...)
			data["enabled"] = false
		} else {
			if kind == "map.block" {
				for _, a := range w.Actors {
					if a.HP > 0 && p.X >= a.X && p.X < a.X+max(1, a.Size) && p.Y >= a.Y && p.Y < a.Y+max(1, a.Size) {
						return fail("OCCUPIED: move the token before blocking this cell")
					}
				}
			}
			*list = append(*list, p)
			data["enabled"] = true
		}
	case "map.draw":
		a, ok := cell("")
		b, ok2 := cell("to_")
		shape := text(v, "shape")
		if !ok || !ok2 || (shape != "line" && shape != "rectangle") {
			return fail("INVALID_DRAWING")
		}
		if len(w.Drawings) >= 200 {
			return fail("DRAWING_LIMIT: remove an annotation first")
		}
		d := Drawing{ID: id(), From: a, To: b, Shape: shape}
		w.Drawings = append(w.Drawings, d)
		data["drawing"] = d
	case "map.undo":
		if len(w.Drawings) == 0 {
			return fail("NO_ANNOTATIONS: nothing to undo")
		}
		w.Drawings = w.Drawings[:len(w.Drawings)-1]
	case "map.scene":
		scene := text(v, "scene")
		if scene != "village" && scene != "blank" {
			return fail("INVALID_SCENE")
		}
		w.MapImage = scene
		w.ImageData = ""
		data["scene"] = scene
	case "map.upload":
		dataURL := text(v, "image_data")
		parts := strings.SplitN(dataURL, ",", 2)
		if len(parts) != 2 || (parts[0] != "data:image/png;base64" && parts[0] != "data:image/jpeg;base64") {
			return fail("INVALID_IMAGE: choose a PNG or JPEG map")
		}
		decoded, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil || len(decoded) > 5<<20 {
			return fail("INVALID_IMAGE: maximum file size is 5 MB")
		}
		config, format, err := image.DecodeConfig(bytes.NewReader(decoded))
		if err != nil || config.Width < 1 || config.Height < 1 || config.Width*config.Height > 20000000 || (format != "png" && format != "jpeg") {
			return fail("INVALID_IMAGE: use a PNG or JPEG under 20 megapixels")
		}
		w.ImageData = dataURL
		w.MapImage = "custom"
		data["scene"] = "custom map"
	case "token.add":
		p, ok := cell("")
		hp, hpok := integer(v, "hp", 1, 999)
		n := text(v, "name")
		if !ok || !hpok || len(n) == 0 || len(n) > 60 {
			return fail("INVALID_TOKEN: name and HP from 1 to 999 required")
		}
		if len(w.Actors) >= 50 {
			return fail("TOKEN_LIMIT: 50 tokens per scene")
		}
		if blocked(w, p) {
			return fail("BLOCKED_PATH")
		}
		for _, a := range w.Actors {
			if a.HP > 0 && p.X >= a.X && p.X < a.X+max(1, a.Size) && p.Y >= a.Y && p.Y < a.Y+max(1, a.Size) {
				return fail("OCCUPIED")
			}
		}
		hostile, _ := v["hostile"].(bool)
		a := Actor{ID: id(), Name: n, X: p.X, Y: p.Y, HP: hp, MaxHP: hp, Movement: 6, AttackBonus: 3, Damage: 4, Hostile: hostile}
		w.Actors = append(w.Actors, a)
		if w.ActiveActorID == "" && !hostile {
			w.ActiveActorID = a.ID
		}
		data["name"] = n
	case "token.update", "token.place":
		_, a := actor(w, text(v, "token_id"))
		if a == nil {
			return fail("TOKEN_NOT_FOUND")
		}
		if kind == "token.place" {
			p, ok := cell("")
			if !ok || !canOccupy(w, p, a.ID, 0) {
				return fail("INVALID_CELL")
			}
			for _, other := range w.Actors {
				if other.ID != a.ID && other.HP > 0 && other.X == p.X && other.Y == p.Y {
					return fail("OCCUPIED")
				}
			}
			a.X = p.X
			a.Y = p.Y
			data["point"] = p
		} else {
			hp, ok := integer(v, "hp", 0, a.MaxHP)
			n := text(v, "name")
			if !ok || n == "" || len(n) > 60 {
				return fail("INVALID_TOKEN: check name and hit points")
			}
			a.Name = n
			a.HP = hp
		}
		data["name"] = a.Name
	case "dice.roll":
		count, a := integer(v, "count", 1, 20)
		sides, b := integer(v, "sides", 2, 100)
		bonus, c := integer(v, "bonus", -100, 100)
		if !a || !b || !c {
			return fail("INVALID_DICE: use 1 to 20 dice, 2 to 100 sides and bonus -100 to 100")
		}
		rolls := []int{}
		total := bonus
		for i := 0; i < count; i++ {
			n := rand.IntN(sides) + 1
			rolls = append(rolls, n)
			total += n
		}
		data["rolls"] = rolls
		data["total"] = total
		data["formula"] = fmt.Sprintf("%dd%d%+d", count, sides, bonus)
	case "notes.save":
		notes, _ := v["notes"].(string)
		if len(notes) > 12000 {
			return fail("NOTES_TOO_LONG: 12000 character limit")
		}
		w.Notes = notes
	case "chat.send":
		msg := text(v, "message")
		if len(msg) == 0 || len(msg) > 1000 {
			return fail("INVALID_MESSAGE: enter up to 1000 characters")
		}
		data["message"] = msg
	default:
		return fail("INVALID_COMMAND")
	}
	return w, event, "", data, nil
}
