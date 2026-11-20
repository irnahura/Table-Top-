package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	rand "math/rand/v2"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Handout struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Text   string `json:"text"`
	Shared bool   `json:"shared"`
}
type Macro struct {
	ID      string `json:"id"`
	OwnerID string `json:"owner_id"`
	Name    string `json:"name"`
	Formula string `json:"formula"`
}
type AudioState struct {
	URL       string    `json:"url"`
	Title     string    `json:"title"`
	Playing   bool      `json:"playing"`
	StartedAt time.Time `json:"started_at"`
}
type Page struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	State *World `json:"state"`
}
type Initiative struct {
	TokenID string `json:"token_id"`
	Score   int    `json:"score"`
}
type TableFeatures struct {
	GameSystem     string `json:"game_system,omitempty"`
	TrackerVisible bool   `json:"tracker_visible,omitempty"`
	AdventureFeatures
	SightEnabled   bool         `json:"sight_enabled,omitempty"`
	SightRadius    int          `json:"sight_radius,omitempty"`
	VisibleCells   []Point      `json:"visible_cells,omitempty"`
	ArtAssets      []ArtAsset   `json:"art_assets,omitempty"`
	SceneID        string       `json:"scene_id,omitempty"`
	SceneName      string       `json:"scene_name,omitempty"`
	Pages          []Page       `json:"pages,omitempty"`
	Scale          float64      `json:"scale,omitempty"`
	Unit           string       `json:"unit,omitempty"`
	Metric         string       `json:"metric,omitempty"`
	MovementPaused bool         `json:"movement_paused,omitempty"`
	Handouts       []Handout    `json:"handouts,omitempty"`
	Macros         []Macro      `json:"macros,omitempty"`
	Audio          AudioState   `json:"audio"`
	Initiative     []Initiative `json:"initiative,omitempty"`
}
type ArtAsset struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ImageData string `json:"image_data"`
}

func contains(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}
func hiddenLayer(layer string) bool { return layer == "gm" }

// Construct a fresh view without mutating shared state. Every read transport,
// conflict, replay and command response uses this same projection.
func (s *Server) sessionView(ss Session, user string) Session {
	gm := s.db.Campaigns[ss.CampaignID].OwnerID == user
	w := ss.State
	w.Environment = environmentView(ss.State, user, gm)
	w.AdventureFeatures = adventureView(w.AdventureFeatures, user, gm)
	w.Macros = []Macro{}
	for _, m := range ss.State.Macros {
		if m.OwnerID == user {
			w.Macros = append(w.Macros, m)
		}
	}
	if !gm {
		visible := sightCells(ss.State, user)
		w.Actors = []Actor{}
		for _, a := range ss.State.Actors {
			if !hiddenLayer(a.Layer) && (!w.SightEnabled || visible[Point{a.X, a.Y}]) {
				w.Actors = append(w.Actors, a)
			}
		}
		w.Drawings = []Drawing{}
		for _, d := range ss.State.Drawings {
			inSight := true
			if w.SightEnabled {
				for _, p := range append([]Point{d.From, d.To}, d.Points...) {
					if !visible[p] {
						inSight = false
					}
				}
			}
			if !hiddenLayer(d.Layer) && inSight {
				for _, p := range d.Path {
					if w.SightEnabled && !visible[Point{int(p.X), int(p.Y)}] {
						inSight = false
					}
				}
			}
			if !hiddenLayer(d.Layer) && inSight {
				w.Drawings = append(w.Drawings, d)
			}
		}
		w.Handouts = []Handout{}
		for _, h := range ss.State.Handouts {
			if h.Shared {
				w.Handouts = append(w.Handouts, h)
			}
		}
		w.Pages = nil // Archived scenes are GM content until the GM switches scenes.
		w.ArtAssets = nil
		if w.SightEnabled {
			w.Blocked = nil
			w.Fog = nil
			for _, p := range ss.State.Blocked {
				if visible[p] {
					w.Blocked = append(w.Blocked, p)
				}
			}
			for _, p := range ss.State.Fog {
				if visible[p] {
					w.Fog = append(w.Fog, p)
				}
			}
		}
		if w.SightEnabled {
			w.VisibleCells = nil
			for y := 0; y < w.Height; y++ {
				for x := 0; x < w.Width; x++ {
					if visible[Point{x, y}] {
						w.VisibleCells = append(w.VisibleCells, Point{x, y})
					}
				}
			}
			w.ImageData = s.maskedMap(ss.State, visible)
			if w.ImageData == "" {
				w.MapImage = "blank"
			}
		}
		w.Initiative = nil
		for _, entry := range ss.State.Initiative {
			if _, a := actor(w, entry.TokenID); a != nil {
				w.Initiative = append(w.Initiative, entry)
			}
		}
		if _, a := actor(w, w.ActiveActorID); a == nil {
			w.ActiveActorID = ""
		}
	}
	ss.State = w
	events := []Event{}
	for _, ev := range ss.Events {
		if ev.GMOnly && !gm {
			continue
		}
		if len(ev.Audience) > 0 && !contains(ev.Audience, user) {
			continue
		}
		events = append(events, ev)
	}
	ss.Events = events
	return ss
}

func turnActors(w World) []Actor {
	out := []Actor{}
	seen := map[string]bool{}
	for _, entry := range w.Initiative {
		if _, a := actor(w, entry.TokenID); a != nil && !a.Hostile && !hiddenLayer(a.Layer) && a.Layer != "map" {
			out = append(out, *a)
			seen[a.ID] = true
		}
	}
	for _, a := range w.Actors {
		if !seen[a.ID] && !a.Hostile && !hiddenLayer(a.Layer) && a.Layer != "map" {
			out = append(out, a)
		}
	}
	return out
}

// Size is measured in square grid cells. Artwork and GM-only tokens do not
// secretly block a player's route; token and foreground actors do.
func canOccupy(w World, p Point, tokenID string, size int) bool {
	if size < 1 {
		size = 1
		if _, a := actor(w, tokenID); a != nil {
			size = max(1, a.Size)
		}
	}
	if p.X < 0 || p.Y < 0 || p.X+size > w.Width || p.Y+size > w.Height {
		return false
	}
	for _, b := range w.Blocked {
		if b.X >= p.X && b.X < p.X+size && b.Y >= p.Y && b.Y < p.Y+size {
			return false
		}
	}
	for _, a := range w.Actors {
		other := max(1, a.Size)
		if a.ID != tokenID && a.HP > 0 && !hiddenLayer(a.Layer) && a.Layer != "map" && p.X < a.X+other && p.X+size > a.X && p.Y < a.Y+other && p.Y+size > a.Y {
			return false
		}
	}
	return true
}

func extendedCommand(kind string) bool {
	if systemCommand(kind) {
		return true
	}
	if kind == "initiative.manage" {
		return true
	}
	if environmentCommand(kind) {
		return true
	}
	if adventureCommand(kind) {
		return true
	}
	switch kind {
	case "sight.settings", "asset.save", "asset.delete", "asset.map", "asset.token", "table.settings", "table.pause", "scene.create", "scene.switch", "scene.rename", "scene.delete", "draw.add", "draw.delete", "draw.undo", "token.configure", "token.delete", "token.duplicate", "initiative.roll", "initiative.start", "handout.save", "handout.delete", "macro.save", "macro.delete", "macro.run", "audio.set", "chat.post", "chat.send", "fx.burst", "measure.share":
		return true
	}
	return false
}

var colorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
var dicePattern = regexp.MustCompile(`(?i)^([0-9]{0,2})d([0-9]{1,3})([+-][0-9]{1,3})?$`)

func diceFormula(formula string) (map[string]any, error) {
	m := dicePattern.FindStringSubmatch(strings.ReplaceAll(formula, " ", ""))
	if m == nil {
		return nil, errors.New("Use NdS+bonus, for example 2d6+3")
	}
	count := 1
	if m[1] != "" {
		count, _ = strconv.Atoi(m[1])
	}
	sides, _ := strconv.Atoi(m[2])
	bonus := 0
	if m[3] != "" {
		bonus, _ = strconv.Atoi(m[3])
	}
	if count < 1 || count > 20 || sides < 2 || sides > 100 || bonus < -100 || bonus > 100 {
		return nil, errors.New("Use 1–20 dice, 2–100 sides and bonus -100 to 100")
	}
	rolls := []int{}
	total := bonus
	for i := 0; i < count; i++ {
		r := rand.IntN(sides) + 1
		rolls = append(rolls, r)
		total += r
	}
	return map[string]any{"rolls": rolls, "total": total, "formula": fmt.Sprintf("%dd%d%+d", count, sides, bonus)}, nil
}

func privateToolEvent(before, after World, v map[string]any) bool {
	if hiddenLayer(text(v, "layer")) {
		return true
	}
	for _, world := range []World{before, after} {
		for _, key := range []string{"token_id", "character_id", "attacker_id", "target_id"} {
			if _, a := actor(world, text(v, key)); a != nil && hiddenLayer(a.Layer) {
				return true
			}
		}
	}
	return false
}

// Runs under db.mu. Deep copy first: failed validation cannot mutate any nested
// page, token marker, drawing point or handout in the authoritative state.
func (s *Server) tableCommand(original World, v map[string]any, c Campaign, user string) (World, string, string, any, []string, bool, error) {
	if systemCommand(text(v, "type")) {
		return s.systemCommand(original, v, c, user)
	}
	if text(v, "type") == "initiative.manage" {
		return s.manageInitiative(original, v, c, user)
	}
	if environmentCommand(text(v, "type")) {
		return s.environmentCommand(original, v, c, user)
	}
	if adventureCommand(text(v, "type")) {
		return s.adventureCommand(original, v, c, user)
	}
	w := World{}
	b, _ := json.Marshal(original)
	_ = json.Unmarshal(b, &w)
	kind := text(v, "type")
	gm := c.OwnerID == user
	name := s.db.Users[user].DisplayName
	data := map[string]any{"by": name}
	var audience []string
	gmOnly := false
	fail := func(message string) (World, string, string, any, []string, bool, error) {
		return original, "", "", nil, nil, false, errors.New(message)
	}
	memberTool := kind == "chat.post" || kind == "chat.send" || kind == "macro.save" || kind == "macro.delete" || kind == "macro.run" || kind == "measure.share"
	if !gm && !memberTool {
		return fail("PERMISSION_DENIED: this tool is managed by the game master")
	}
	cell := func(xKey, yKey string) (Point, bool) {
		x, a := integer(v, xKey, 0, w.Width-1)
		y, b := integer(v, yKey, 0, w.Height-1)
		return Point{x, y}, a && b
	}
	switch kind {
	case "sight.settings":
		enabled, a := v["enabled"].(bool)
		radius, b := integer(v, "radius", 1, 30)
		if !a || !b {
			return fail("INVALID_SIGHT: choose a radius from 1 to 30 grid cells")
		}
		w.SightEnabled = enabled
		w.SightRadius = radius
		data["message"] = "Token line of sight updated."
	case "asset.save", "asset.delete", "asset.map", "asset.token":
		index := -1
		for i, a := range w.ArtAssets {
			if a.ID == text(v, "asset_id") {
				index = i
			}
		}
		if kind == "asset.save" {
			gmOnly = true
			n := text(v, "name")
			imageData := text(v, "image_data")
			if n == "" || len(n) > 80 || len(imageData) > 3<<20 || len(w.ArtAssets) >= 10 {
				return fail("INVALID_ASSET: name required, image up to 2 MB; maximum 10 assets")
			}
			temp, _, _, _, err := toolRule(w, map[string]any{"type": "map.upload", "image_data": imageData}, true, name)
			if err != nil {
				return fail(err.Error())
			}
			parts := strings.SplitN(temp.ImageData, ",", 2)
			decoded, _ := base64.StdEncoding.DecodeString(parts[1])
			if len(decoded) > 2<<20 {
				return fail("INVALID_ASSET: maximum decoded image size is 2 MB")
			}
			w.ArtAssets = append(w.ArtAssets, ArtAsset{ID: id(), Name: n, ImageData: temp.ImageData})
			data["message"] = "Art saved: " + n
		} else {
			if index < 0 {
				return fail("ASSET_NOT_FOUND")
			}
			asset := w.ArtAssets[index]
			if kind == "asset.delete" {
				gmOnly = true
				w.ArtAssets = append(w.ArtAssets[:index], w.ArtAssets[index+1:]...)
				data["message"] = "Library image removed; placed copies remain."
			} else if kind == "asset.map" {
				w.MapImage = "custom"
				w.ImageData = asset.ImageData
				data["message"] = "Map image applied."
			} else {
				_, a := actor(w, text(v, "token_id"))
				if a == nil {
					return fail("TOKEN_NOT_FOUND")
				}
				a.ImageData = asset.ImageData
				gmOnly = hiddenLayer(a.Layer)
				data["message"] = "Token artwork applied."
			}
		}
	case "table.settings":
		width, a := integer(v, "width", 2, 100)
		height, b := integer(v, "height", 2, 100)
		scale, ok := v["scale"].(float64)
		unit := text(v, "unit")
		metric := text(v, "metric")
		if !a || !b || !ok || math.IsNaN(scale) || scale <= 0 || scale > 1000 || len(unit) < 1 || len(unit) > 16 || !contains([]string{"manhattan", "euclidean", "chebyshev"}, metric) {
			return fail("INVALID_GRID: dimensions 2–100, positive scale up to 1000, and a distance metric are required")
		}
		for _, actor := range w.Actors {
			size := max(1, actor.Size)
			if actor.X < 0 || actor.Y < 0 || actor.X+size > width || actor.Y+size > height {
				return fail("INVALID_GRID: move tokens inside the new map before shrinking it")
			}
		}
		for _, d := range w.Doors {
			if d.X >= width || d.Y >= height {
				return fail("INVALID_GRID: remove outlying doors before shrinking")
			}
		}
		for _, l := range w.Lights {
			if l.X >= width || l.Y >= height {
				return fail("INVALID_GRID: remove outlying lights before shrinking")
			}
		}
		for _, d := range w.Drawings {
			points := append([]Point{d.From, d.To}, d.Points...)
			for _, p := range d.Path {
				points = append(points, Point{int(p.X), int(p.Y)})
			}
			for _, p := range points {
				if p.X >= width || p.Y >= height {
					return fail("INVALID_GRID: remove drawings outside the new map before shrinking it")
				}
			}
		}
		for _, p := range append(append([]Point{}, w.Blocked...), w.Fog...) {
			if p.X >= width || p.Y >= height {
				return fail("INVALID_GRID: clear terrain/fog outside the new map before shrinking it")
			}
		}
		w.Width = width
		w.Height = height
		w.Scale = scale
		w.Unit = unit
		w.Metric = metric
		data["message"] = "Grid dimensions and measurement scale updated."
	case "table.pause":
		paused, ok := v["paused"].(bool)
		if !ok {
			return fail("INVALID_PAUSE")
		}
		w.MovementPaused = paused
		data["message"] = "Player movement resumed."
		if paused {
			data["message"] = "Player movement paused by the game master."
		}
	case "scene.create", "scene.switch", "scene.rename", "scene.delete":
		gmOnly = true
		if w.SceneID == "" {
			w.SceneID = id()
			w.SceneName = "Riverside village"
		}
		if kind == "scene.create" {
			n := text(v, "name")
			if n == "" || len(n) > 80 || len(w.Pages) >= 9 {
				return fail("INVALID_SCENE: enter a name; maximum 10 scenes")
			}
			current := w
			current.Pages = nil
			current.Macros = nil
			current.Handouts = nil
			current.AdventureFeatures = AdventureFeatures{}
			current.ArtAssets = nil
			w.Pages = append(w.Pages, Page{ID: w.SceneID, Name: w.SceneName, State: &current})
			duplicate, _ := v["duplicate"].(bool)
			if !duplicate {
				w.Environment = Environment{}
				w.Actors = nil
				w.Drawings = nil
				w.Blocked = nil
				w.Fog = nil
				w.ImageData = ""
				w.MapImage = "blank"
				w.ActiveActorID = ""
				w.MovementUsed = 0
				w.AttackUsed = false
				w.Round = 1
				w.Initiative = nil
			}
			w.SceneID = id()
			w.SceneName = n
			data["message"] = "Scene created: " + n
		} else if kind == "scene.switch" {
			target := text(v, "scene_id")
			found := -1
			for i, p := range w.Pages {
				if p.ID == target {
					found = i
				}
			}
			if found < 0 || w.Pages[found].State == nil {
				return fail("SCENE_NOT_FOUND")
			}
			page := w.Pages[found]
			current := w
			current.Pages = nil
			current.Macros = nil
			current.Handouts = nil
			current.AdventureFeatures = AdventureFeatures{}
			current.ArtAssets = nil
			pages := append([]Page(nil), w.Pages...)
			pages[found] = Page{ID: w.SceneID, Name: w.SceneName, State: &current}
			old := w
			w = *page.State
			w.Pages = pages
			w.SceneID = page.ID
			w.SceneName = page.Name
			w.Handouts = old.Handouts
			w.AdventureFeatures = old.AdventureFeatures
			w.GameSystem = old.GameSystem
			w.Macros = old.Macros
			w.Audio = old.Audio
			w.Notes = old.Notes
			w.ArtAssets = old.ArtAssets
			w.MovementPaused = old.MovementPaused
			data["message"] = "Active scene: " + w.SceneName
		} else if kind == "scene.rename" {
			n := text(v, "name")
			if n == "" || len(n) > 80 {
				return fail("INVALID_SCENE_NAME")
			}
			w.SceneName = n
			data["message"] = "Scene renamed to " + n
		} else {
			target := text(v, "scene_id")
			found := -1
			for i, p := range w.Pages {
				if p.ID == target {
					found = i
				}
			}
			if found < 0 {
				return fail("SCENE_NOT_FOUND: only archived scenes can be deleted")
			}
			w.Pages = append(w.Pages[:found], w.Pages[found+1:]...)
			data["message"] = "Archived scene deleted."
		}
	case "draw.add":
		p, valid := cell("x", "y")
		q, valid2 := cell("to_x", "to_y")
		shape := text(v, "shape")
		layer := text(v, "layer")
		color := text(v, "color")
		stroke, ok := integer(v, "stroke", 1, 12)
		if !valid || !valid2 || !ok || !colorPattern.MatchString(color) || !contains([]string{"line", "rectangle", "ellipse", "text", "freehand", "pin", "cone"}, shape) || !contains([]string{"map", "tokens", "gm", "foreground"}, layer) || len(w.Drawings) >= 200 {
			return fail("INVALID_DRAWING: check shape, layer, color and map coordinates")
		}
		d := Drawing{ID: id(), From: p, To: q, Shape: shape, Layer: layer, Color: color, Stroke: stroke, Text: text(v, "text")}
		if len(d.Text) > 500 || ((shape == "text" || shape == "pin") && d.Text == "") {
			return fail("INVALID_TEXT: enter up to 500 characters")
		}
		if shape == "freehand" {
			if raw, exists := v["path"]; exists {
				path, valid := raw.([]any)
				if !valid || len(path) < 2 || len(path) > 1000 {
					return fail("INVALID_PATH: use 2–1000 smooth path points")
				}
				for _, value := range path {
					p, ok := value.(map[string]any)
					if !ok {
						return fail("INVALID_PATH")
					}
					x, a := p["x"].(float64)
					y, b := p["y"].(float64)
					if !a || !b || math.IsNaN(x) || math.IsNaN(y) || math.IsInf(x, 0) || math.IsInf(y, 0) || x < 0 || y < 0 || x >= float64(w.Width) || y >= float64(w.Height) {
						return fail("INVALID_PATH: smooth points must be inside the map")
					}
					d.Path = append(d.Path, DrawPoint{x, y})
				}
			}
			points, ok := v["points"].([]any)
			if !ok || len(points) < 2 || len(points) > 250 {
				return fail("INVALID_PATH: use 2–250 points")
			}
			for _, p := range points {
				m, ok := p.(map[string]any)
				if !ok {
					return fail("INVALID_PATH")
				}
				x, a := integer(m, "x", 0, w.Width-1)
				y, b := integer(m, "y", 0, w.Height-1)
				if !a || !b {
					return fail("INVALID_PATH")
				}
				d.Points = append(d.Points, Point{x, y})
			}
		}
		w.Drawings = append(w.Drawings, d)
		gmOnly = hiddenLayer(layer)
		data["message"] = "Saved " + shape + " on " + layer + " layer."
		data["drawing_id"] = d.ID
	case "draw.delete", "draw.undo":
		index := -1
		for i, d := range w.Drawings {
			if (kind == "draw.delete" && d.ID == text(v, "drawing_id")) || (kind == "draw.undo" && d.Layer == text(v, "layer")) {
				index = i
			}
		}
		if index < 0 {
			return fail("DRAWING_NOT_FOUND")
		}
		gmOnly = hiddenLayer(w.Drawings[index].Layer)
		w.Drawings = append(w.Drawings[:index], w.Drawings[index+1:]...)
		data["message"] = "Drawing removed."
	case "token.configure", "token.duplicate", "token.delete":
		index, a := actor(w, text(v, "token_id"))
		if a == nil {
			return fail("TOKEN_NOT_FOUND")
		}
		gmOnly = hiddenLayer(a.Layer)
		if kind == "token.configure" {
			layer := text(v, "layer")
			rotation, ok := integer(v, "rotation", 0, 359)
			size, ok2 := integer(v, "size", 1, 4)
			bonus, ok3 := integer(v, "initiative_bonus", -20, 20)
			locked, ok4 := v["locked"].(bool)
			if !ok || !ok2 || !ok3 || !ok4 || !contains([]string{"tokens", "gm", "map", "foreground"}, layer) || a.X+size > w.Width || a.Y+size > w.Height {
				return fail("INVALID_TOKEN_SETTINGS")
			}
			markers := []string{}
			raw, ok := v["markers"].([]any)
			if !ok || len(raw) > 6 {
				return fail("INVALID_MARKERS")
			}
			for _, m := range raw {
				value, ok := m.(string)
				if !ok || !contains([]string{"poisoned", "prone", "stunned", "blessed", "concentrating", "invisible"}, value) {
					return fail("INVALID_MARKERS")
				}
				if !contains(markers, value) {
					markers = append(markers, value)
				}
			}
			if (hiddenLayer(layer) || layer == "map") && w.ActiveActorID == a.ID {
				return fail("ACTIVE_TOKEN: end this actor's turn before moving it off the token layer")
			}
			if layer != "map" && !hiddenLayer(layer) && !canOccupy(w, Point{a.X, a.Y}, a.ID, size) {
				return fail("OCCUPIED: the token footprint overlaps terrain or another token")
			}
			a.Layer = layer
			a.Rotation = rotation
			a.Size = size
			a.Locked = locked
			a.InitiativeBonus = bonus
			a.Markers = markers
			gmOnly = gmOnly || hiddenLayer(layer)
			data["message"] = "Token settings updated: " + a.Name
		} else if kind == "token.duplicate" {
			if len(w.Actors) >= 50 {
				return fail("TOKEN_LIMIT")
			}
			p, ok := cell("x", "y")
			if !ok || !canOccupy(w, p, a.ID, max(1, a.Size)) {
				return fail("INVALID_CELL")
			}
			for _, other := range w.Actors {
				if other.X == p.X && other.Y == p.Y && other.HP > 0 {
					return fail("OCCUPIED")
				}
			}
			copy := *a
			copy.ID = id()
			copy.Name = a.Name + " copy"
			copy.X = p.X
			copy.Y = p.Y
			if p.X+max(1, copy.Size) > w.Width || p.Y+max(1, copy.Size) > w.Height {
				return fail("INVALID_CELL")
			}
			w.Actors = append(w.Actors, copy)
			data["message"] = "Token duplicated."
		} else {
			if a.ID == w.ActiveActorID {
				return fail("ACTIVE_TOKEN: end this actor's turn before deleting it")
			}
			w.Actors = append(w.Actors[:index], w.Actors[index+1:]...)
			data["message"] = "Token deleted."
		}
	case "initiative.roll", "initiative.start":
		if kind == "initiative.roll" {
			w.Initiative = nil
			for _, a := range turnActors(w) {
				if a.HP > 0 {
					w.Initiative = append(w.Initiative, Initiative{TokenID: a.ID, Score: rand.IntN(20) + 1 + a.InitiativeBonus})
				}
			}
			sort.SliceStable(w.Initiative, func(i, j int) bool { return w.Initiative[i].Score > w.Initiative[j].Score })
			data["message"] = "Initiative rolled for living party characters."
		} else {
			if len(w.Initiative) == 0 {
				return fail("ROLL_INITIATIVE_FIRST")
			}
			_, a := actor(w, w.Initiative[0].TokenID)
			if a == nil || a.HP <= 0 || a.Hostile || hiddenLayer(a.Layer) || a.Layer == "map" {
				return fail("ROLL_INITIATIVE_FIRST")
			}
			w.ActiveActorID = a.ID
			w.Round = 1
			w.AttackUsed = false
			w.MovementUsed = 0
			data["message"] = "Encounter started in initiative order."
		}
	case "handout.save", "handout.delete":
		index := -1
		for i, h := range w.Handouts {
			if h.ID == text(v, "handout_id") {
				index = i
			}
		}
		if kind == "handout.delete" {
			if index < 0 {
				return fail("HANDOUT_NOT_FOUND")
			}
			gmOnly = !w.Handouts[index].Shared
			w.Handouts = append(w.Handouts[:index], w.Handouts[index+1:]...)
			data["message"] = "Handout removed."
		} else {
			title := text(v, "title")
			body := text(v, "text")
			shared, ok := v["shared"].(bool)
			if title == "" || len(title) > 100 || len(body) > 20000 || !ok {
				return fail("INVALID_HANDOUT")
			}
			if index < 0 && len(w.Handouts) >= 50 {
				return fail("HANDOUT_LIMIT")
			}
			h := Handout{ID: id(), Title: title, Text: body, Shared: shared}
			if index >= 0 {
				h.ID = w.Handouts[index].ID
				w.Handouts[index] = h
			} else {
				w.Handouts = append(w.Handouts, h)
			}
			gmOnly = !shared
			data["message"] = "Handout saved: " + title
		}
	case "macro.save", "macro.delete", "macro.run":
		index := -1
		for i, m := range w.Macros {
			if m.ID == text(v, "macro_id") && m.OwnerID == user {
				index = i
			}
		}
		switch kind {
		case "macro.delete":
			if index < 0 {
				return fail("MACRO_NOT_FOUND")
			}
			w.Macros = append(w.Macros[:index], w.Macros[index+1:]...)
			audience = []string{user}
			data["message"] = "Macro deleted."
		case "macro.save":
			n := text(v, "name")
			formula := text(v, "formula")
			if n == "" || len(n) > 60 || len(formula) > 30 || !dicePattern.MatchString(strings.ReplaceAll(formula, " ", "")) {
				return fail("INVALID_MACRO")
			}
			if _, err := diceFormula(formula); err != nil {
				return fail(err.Error())
			}
			count := 0
			for _, m := range w.Macros {
				if m.OwnerID == user {
					count++
				}
			}
			if index < 0 && count >= 30 {
				return fail("MACRO_LIMIT")
			}
			m := Macro{ID: id(), OwnerID: user, Name: n, Formula: formula}
			if index >= 0 {
				m.ID = w.Macros[index].ID
				w.Macros[index] = m
			} else {
				w.Macros = append(w.Macros, m)
			}
			audience = []string{user}
			data["message"] = "Dice macro saved."
		case "macro.run":
			if index < 0 {
				return fail("MACRO_NOT_FOUND")
			}
			roll, err := diceFormula(w.Macros[index].Formula)
			if err != nil {
				return fail(err.Error())
			}
			for k, v := range roll {
				data[k] = v
			}
			data["message"] = "Macro: " + w.Macros[index].Name
			kind = "dice.roll"
		}
	case "audio.set":
		address := text(v, "url")
		title := text(v, "title")
		playing, ok := v["playing"].(bool)
		parsed, err := url.Parse(address)
		if !ok || len(address) > 2000 || len(title) > 100 || (address != "" && (err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" || parsed.User != nil)) || (playing && address == "") {
			return fail("INVALID_AUDIO: enter a public HTTP(S) audio URL")
		}
		w.Audio = AudioState{URL: address, Title: title, Playing: playing, StartedAt: time.Now().UTC()}
		data["message"] = "Shared audio updated. Each listener must enable audio in their browser."
	case "chat.post", "chat.send":
		msg := text(v, "message")
		if msg == "" || len(msg) > 1000 {
			return fail("INVALID_MESSAGE: enter up to 1000 characters")
		}
		by := name
		if as := text(v, "as_character"); as != "" {
			_, a := actor(w, as)
			if a == nil || hiddenLayer(a.Layer) || (!gm && (a.ControllerID != user || a.Hostile)) {
				return fail("PERMISSION_DENIED: you may speak only as a character you control")
			}
			by = a.Name
		}
		data["by"] = by
		if strings.HasPrefix(msg, "/roll ") || strings.HasPrefix(msg, "/r ") || strings.HasPrefix(msg, "/gmroll ") || strings.HasPrefix(msg, "/gr ") {
			parts := strings.SplitN(msg, " ", 2)
			roll, err := diceFormula(parts[1])
			if err != nil {
				return fail(err.Error())
			}
			for k, v := range roll {
				data[k] = v
			}
			kind = "dice.roll"
			if parts[0] == "/gmroll" || parts[0] == "/gr" {
				audience = []string{user, c.OwnerID}
				data["private"] = true
			}
		} else {
			if strings.HasPrefix(msg, "/w ") {
				rest := strings.TrimSpace(strings.TrimPrefix(msg, "/w "))
				var target string
				if strings.HasPrefix(rest, "\"") {
					end := strings.Index(rest[1:], "\"")
					if end < 0 {
						return fail("INVALID_WHISPER: quote the recipient's full display name")
					}
					target = rest[1 : end+1]
					msg = strings.TrimSpace(rest[end+2:])
				} else {
					parts := strings.SplitN(rest, " ", 2)
					if len(parts) != 2 {
						return fail("INVALID_WHISPER")
					}
					target = parts[0]
					msg = strings.TrimSpace(parts[1])
				}
				match := ""
				if strings.EqualFold(target, "gm") {
					match = c.OwnerID
				} else {
					for id := range c.Members {
						if id == target || strings.EqualFold(s.db.Users[id].DisplayName, target) {
							if match != "" && match != id {
								return fail("AMBIGUOUS_PLAYER: use the member ID or a unique display name")
							}
							match = id
						}
					}
				}
				if match == "" || msg == "" {
					return fail("WHISPER_RECIPIENT_NOT_FOUND")
				}
				audience = []string{user, match}
				data["private"] = true
			} else if strings.HasPrefix(msg, "/em ") || strings.HasPrefix(msg, "/me ") {
				msg = strings.TrimSpace(strings.SplitN(msg, " ", 2)[1])
				data["emote"] = true
			} else if strings.HasPrefix(msg, "/") {
				return fail("UNKNOWN_CHAT_COMMAND: use /roll, /gmroll, /w, or /em")
			}
			data["message"] = msg
			kind = "chat.send"
		}
	case "fx.burst", "measure.share":
		p, a := cell("x", "y")
		if !a {
			return fail("INVALID_CELL")
		}
		data["point"] = p
		if kind == "measure.share" {
			q, b := cell("to_x", "to_y")
			if !b {
				return fail("INVALID_CELL")
			}
			dx := math.Abs(float64(q.X - p.X))
			dy := math.Abs(float64(q.Y - p.Y))
			distance := dx + dy
			if w.Metric == "euclidean" {
				distance = math.Hypot(dx, dy)
			} else if w.Metric == "chebyshev" {
				distance = math.Max(dx, dy)
			}
			scale := w.Scale
			if scale <= 0 {
				scale = 1
			}
			data["to"] = q
			data["distance"] = distance * scale
			data["unit"] = w.Unit
			data["message"] = fmt.Sprintf("Shared measurement: %.2f %s", distance*scale, w.Unit)
		} else {
			data["message"] = "Map effect"
		}
	}
	return w, kind, user, data, audience, gmOnly, nil
}
