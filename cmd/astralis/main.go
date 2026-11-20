package main

import (
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	mrand "math/rand/v2"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}
type Actor struct {
	Darkvision      int      `json:"darkvision,omitempty"`
	Layer           string   `json:"layer,omitempty"`
	Rotation        int      `json:"rotation,omitempty"`
	Size            int      `json:"size,omitempty"`
	Markers         []string `json:"markers,omitempty"`
	Locked          bool     `json:"locked,omitempty"`
	ImageData       string   `json:"token_image,omitempty"`
	InitiativeBonus int      `json:"initiative_bonus,omitempty"`
	ControllerID    string   `json:"controller_id,omitempty"`
	ControllerName  string   `json:"controller_name,omitempty"`
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	X               int      `json:"x"`
	Y               int      `json:"y"`
	HP              int      `json:"hp"`
	MaxHP           int      `json:"max_hp"`
	Movement        int      `json:"movement"`
	AttackBonus     int      `json:"attack_bonus"`
	Damage          int      `json:"damage"`
	Hostile         bool     `json:"hostile"`
}
type World struct {
	Environment
	TableFeatures
	ImageData     string    `json:"image_data"`
	MovementUsed  int       `json:"movement_used"`
	AttackUsed    bool      `json:"attack_used"`
	Fog           []Point   `json:"fog"`
	Drawings      []Drawing `json:"drawings"`
	MapImage      string    `json:"map_image"`
	Notes         string    `json:"notes"`
	Width         int       `json:"width"`
	Height        int       `json:"height"`
	Blocked       []Point   `json:"blocked"`
	Actors        []Actor   `json:"actors"`
	ActiveActorID string    `json:"active_actor_id"`
	Round         int       `json:"round"`
}
type User struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	DisplayName  string `json:"display_name"`
}
type Campaign struct {
	ID          string            `json:"id"`
	OwnerID     string            `json:"owner_id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Members     map[string]string `json:"members"`
	CreatedAt   time.Time         `json:"created_at"`
}
type Character struct {
	ID         string `json:"id"`
	OwnerID    string `json:"owner_id"`
	CampaignID string `json:"campaign_id"`
	Name       string `json:"name"`
	ClassName  string `json:"class_name"`
	HP         int    `json:"hp"`
	MaxHP      int    `json:"max_hp"`
	X          int    `json:"x"`
	Y          int    `json:"y"`
}
type GameMap struct {
	ID         string  `json:"id"`
	OwnerID    string  `json:"owner_id"`
	CampaignID string  `json:"campaign_id"`
	Name       string  `json:"name"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	Blocked    []Point `json:"blocked"`
}
type Event struct {
	Audience  []string  `json:"audience,omitempty"`
	GMOnly    bool      `json:"gm_only,omitempty"`
	Sequence  int       `json:"sequence"`
	Type      string    `json:"event_type"`
	ActorID   string    `json:"actor_id"`
	Payload   any       `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}
type Session struct {
	ID         string    `json:"id"`
	CampaignID string    `json:"campaign_id"`
	MapID      string    `json:"map_id"`
	Status     string    `json:"status"`
	Version    int       `json:"state_version"`
	Sequence   int       `json:"sequence"`
	State      World     `json:"state"`
	Events     []Event   `json:"events"`
	CreatedAt  time.Time `json:"created_at"`
}
type DB struct {
	Receipts    map[string]CommandReceipt `json:"command_receipts,omitempty"`
	Passwords   map[string]string         `json:"passwords"`
	Users       map[string]User           `json:"users"`
	Campaigns   map[string]Campaign       `json:"campaigns"`
	Characters  map[string]Character      `json:"characters"`
	Maps        map[string]GameMap        `json:"maps"`
	Sessions    map[string]Session        `json:"sessions"`
	Invites     map[string]string         `json:"invites"`
	InviteRules map[string]Invitation     `json:"invitation_rules,omitempty"`
	Tokens      map[string]string         `json:"tokens"`
	TokenExpiry map[string]time.Time      `json:"token_expiry"`
	mu          sync.RWMutex
	file        string
}

func newDB(file string) *DB {
	d := &DB{Users: map[string]User{}, Campaigns: map[string]Campaign{}, Characters: map[string]Character{}, Maps: map[string]GameMap{}, Sessions: map[string]Session{}, Invites: map[string]string{}, Tokens: map[string]string{}, file: file}
	b, _ := os.ReadFile(file)
	if len(b) > 0 {
		_ = json.Unmarshal(b, d)
	}
	if d.Passwords == nil {
		d.Passwords = map[string]string{}
	}
	if d.TokenExpiry == nil {
		d.TokenExpiry = map[string]time.Time{}
	}
	if d.Receipts == nil {
		d.Receipts = map[string]CommandReceipt{}
	}
	if d.InviteRules == nil {
		d.InviteRules = map[string]Invitation{}
	}
	return d
}
func (d *DB) save() {
	b, _ := json.MarshalIndent(d, "", "  ")
	_ = os.MkdirAll(filepath.Dir(d.file), 0755)
	if err := os.WriteFile(d.file+".tmp", b, 0600); err != nil {
		log.Printf("save failed: %v", err)
		return
	}
	if err := os.Rename(d.file+".tmp", d.file); err != nil {
		log.Printf("save failed: %v", err)
	}
}
func id() string                       { b := make([]byte, 16); _, _ = crand.Read(b); return hex.EncodeToString(b) }
func hash(s string) string             { h := sha256.Sum256([]byte(s)); return hex.EncodeToString(h[:]) }
func member(c Campaign, u string) bool { _, ok := c.Members[u]; return ok }
func actor(w World, id string) (int, *Actor) {
	for i := range w.Actors {
		if w.Actors[i].ID == id {
			return i, &w.Actors[i]
		}
	}
	return -1, nil
}
func dist(a, b Point) int {
	x := a.X - b.X
	if x < 0 {
		x = -x
	}
	y := a.Y - b.Y
	if y < 0 {
		y = -y
	}
	return x + y
}
func blocked(w World, p Point) bool {
	for _, d := range w.Doors {
		if d.X == p.X && d.Y == p.Y && !d.Open {
			return true
		}
	}
	for _, b := range w.Blocked {
		if b == p {
			return true
		}
	}
	return false
}
func rule(w World, cmd map[string]any) (World, string, string, any, error) {
	// Commands work on a copy: a rejected command cannot mutate the live snapshot.
	w.Actors = append([]Actor(nil), w.Actors...)
	kind, _ := cmd["type"].(string)
	actorID, _ := cmd["character_id"].(string)
	if kind == "combat.attack" {
		actorID, _ = cmd["attacker_id"].(string)
	}
	i, a := actor(w, actorID)
	if i < 0 || a == nil {
		return w, "", "", nil, errors.New("NOT_FOUND: actor not found")
	}
	if kind == "character.move" {
		t, _ := cmd["target"].(map[string]any)
		x, xok := integer(t, "x", 0, w.Width-1)
		y, yok := integer(t, "y", 0, w.Height-1)
		if !xok || !yok || a.HP <= 0 {
			return w, "", "", nil, errors.New("INVALID_MOVE: choose a cell inside the map with a living actor")
		}
		p := Point{x, y}
		if w.ActiveActorID != actorID {
			return w, "", "", nil, errors.New("NOT_YOUR_TURN: only the active actor can move")
		}
		if p.X < 0 || p.Y < 0 || p.X >= w.Width || p.Y >= w.Height {
			return w, "", "", nil, errors.New("OUT_OF_RANGE: target is outside map")
		}
		if blocked(w, p) {
			return w, "", "", nil, errors.New("BLOCKED_PATH: target is blocked")
		}
		if !reachable(w, Point{a.X, a.Y}, p, a.Movement-w.MovementUsed, actorID) {
			return w, "", "", nil, errors.New("BLOCKED_PATH: no traversable route within movement range")
		}
		if dist(Point{a.X, a.Y}, p) > a.Movement {
			return w, "", "", nil, errors.New("OUT_OF_RANGE: movement budget exceeded")
		}
		old := Point{a.X, a.Y}
		w.MovementUsed += pathDistance(w, old, p, actorID)
		a.X = p.X
		a.Y = p.Y
		return w, "character.moved", actorID, map[string]any{"character_id": actorID, "from": old, "to": p}, nil
	}
	if kind == "object.interact" {
		if cmd["object_id"] != "old-bridge" || dist(Point{a.X, a.Y}, Point{8, 5}) > 2 {
			return w, "", "", nil, errors.New("OUT_OF_RANGE: move closer to the old bridge")
		}
		return w, "object.interacted", actorID, map[string]any{"character_id": actorID, "object_id": "old-bridge", "result": "The bridge bears fresh wagon marks."}, nil
	}
	if kind == "combat.attack" {
		if a.HP <= 0 || w.AttackUsed {
			return w, "", "", nil, errors.New("ACTION_USED: one attack per turn for a living actor")
		}
		targetID, _ := cmd["target_id"].(string)
		j, t := actor(w, targetID)
		if j < 0 || t == nil || !t.Hostile || t.HP <= 0 {
			return w, "", "", nil, errors.New("NOT_FOUND: hostile target not found")
		}
		if hiddenLayer(t.Layer) || t.Layer == "map" {
			return w, "", "", nil, errors.New("NOT_FOUND: target is not on the token layer")
		}
		if w.ActiveActorID != actorID {
			return w, "", "", nil, errors.New("NOT_YOUR_TURN: not your turn")
		}
		if dist(Point{a.X, a.Y}, Point{t.X, t.Y}) > 5 {
			return w, "", "", nil, errors.New("OUT_OF_RANGE: target is too far away")
		}
		roll := mrand.IntN(20) + 1
		w.AttackUsed = true
		total := roll + a.AttackBonus
		hit := total >= 12
		damage := 0
		if hit {
			damage = a.Damage
			t.HP -= damage
			if t.HP < 0 {
				t.HP = 0
			}
		}
		return w, "attack.resolved", actorID, map[string]any{"attacker_id": actorID, "target_id": targetID, "roll": roll, "total": total, "hit": hit, "damage": damage, "remaining_hp": t.HP}, nil
	}
	if kind == "turn.end" {
		if w.ActiveActorID != actorID {
			return w, "", "", nil, errors.New("NOT_YOUR_TURN: only the active actor can end the turn")
		}
		next := ""
		order := turnActors(w)
		for index, x := range order {
			if x.ID == actorID {
				i = index
				break
			}
		}
		// The basic ruleset cycles party characters; hostile NPCs are GM controlled.
		for step := 1; step <= len(order); step++ {
			index := (i + step) % len(order)
			x := order[index]
			if x.HP > 0 && !x.Hostile {
				next = x.ID
				if index <= i {
					w.Round++
				}
				break
			}
		}
		if next == "" {
			return w, "", "", nil, errors.New("SESSION_NOT_ACTIVE: no active actors")
		}
		w.ActiveActorID = next
		w.MovementUsed = 0
		w.AttackUsed = false
		return w, "turn.advanced", actorID, map[string]any{"active_actor_id": next, "round": w.Round}, nil
	}
	return w, "", "", nil, errors.New("INVALID_COMMAND: unsupported command")
}

type Server struct {
	db          *DB
	web, design string
	peers       map[*livePeer]bool // protected by db.mu
	visionCache map[string]string  // protected by db.mu
}

func (s *Server) json(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func body(r *http.Request) (map[string]any, error) {
	var v map[string]any
	decoder := json.NewDecoder(io.LimitReader(r.Body, 8<<20))
	err := decoder.Decode(&v)
	if err == nil {
		var extra any
		if decoder.Decode(&extra) != io.EOF {
			err = errors.New("multiple JSON values")
		}
	}
	return v, err
}
func (s *Server) auth(r *http.Request) string {
	token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	s.db.mu.RLock()
	defer s.db.mu.RUnlock()
	if expiry, ok := s.db.TokenExpiry[token]; !ok || time.Now().After(expiry) {
		return ""
	}
	return s.db.Tokens[token]
}
func (s *Server) static(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/")
	if name == "" {
		name = "index.html"
	}
	p := filepath.Join(s.web, name)
	if _, err := os.Stat(p); err != nil {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, p)
}
func (s *Server) designFile(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/design/")
	if name == "astralis-tabletop.html" {
		name = "astralis-tabletop-v2.html"
	}
	http.ServeFile(w, r, filepath.Join(s.design, name))
}
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.WriteHeader(204)
		return
	}
	p := r.URL.Path
	if p == "/health" {
		s.json(w, 200, map[string]any{"status": "ok", "service": "astralis-go"})
		return
	}
	if strings.HasPrefix(p, "/design/") {
		s.designFile(w, r)
		return
	}
	if !strings.HasPrefix(p, "/api/") {
		s.static(w, r)
		return
	}
	if strings.HasPrefix(p, "/api/v1/sessions/") && strings.HasSuffix(p, "/live") && r.Method == "GET" {
		s.live(w, r)
		return
	}
	if s.demoAuth(w, r) {
		return
	}
	if p == "/api/v1/auth/register" && r.Method == "POST" {
		v, e := body(r)
		if e != nil {
			s.json(w, 400, map[string]string{"error": "INVALID_JSON"})
			return
		}
		email, password, name := text(v, "email"), text(v, "password"), text(v, "display_name")
		if !strings.Contains(email, "@") || len(email) > 254 || len(password) < 8 || len(password) > 128 || name == "" || len(name) > 80 {
			s.json(w, 400, map[string]string{"error": "Enter a name, valid email and password of 8 to 128 characters."})
			return
		}
		u := User{ID: id(), Email: strings.ToLower(email), DisplayName: name}
		s.db.mu.Lock()
		for _, existing := range s.db.Users {
			if existing.Email == u.Email {
				s.db.mu.Unlock()
				s.json(w, 409, map[string]string{"error": "This email already has an account."})
				return
			}
		}
		s.db.Passwords[u.ID] = passwordHash(password)
		s.db.Users[u.ID] = u
		s.db.save()
		s.db.mu.Unlock()
		u.PasswordHash = ""
		s.json(w, 201, map[string]any{"user": u})
		return
	}
	if p == "/api/v1/auth/login" && r.Method == "POST" {
		v, _ := body(r)
		email := strings.ToLower(fmt.Sprint(v["email"]))
		password := text(v, "password")
		s.db.mu.Lock()
		defer s.db.mu.Unlock()
		for _, u := range s.db.Users {
			if u.Email == email && (verifyPassword(s.db.Passwords[u.ID], password) || (u.PasswordHash != "" && u.PasswordHash == hash(password))) {
				t := id()
				s.db.Tokens[t] = u.ID
				s.db.TokenExpiry[t] = time.Now().Add(7 * 24 * time.Hour)
				s.db.save()
				u.PasswordHash = ""
				s.json(w, 200, map[string]any{"token": t, "user": u})
				return
			}
		}
		s.json(w, 401, map[string]string{"error": "AUTH_REQUIRED"})
		return
	}
	u := s.auth(r)
	if u == "" {
		s.json(w, 401, map[string]string{"error": "AUTH_REQUIRED"})
		return
	}
	v, e := body(r)
	if e != nil && r.Method == "POST" {
		s.json(w, 400, map[string]string{"error": "INVALID_JSON"})
		return
	}
	s.db.mu.Lock()
	defer s.db.mu.Unlock()
	if s.extraAPI(w, r, u, v) {
		return
	}
	if p == "/api/v1/campaigns" && r.Method == "POST" {
		if text(v, "name") == "" || len(text(v, "name")) > 120 || len(text(v, "description")) > 4000 {
			s.json(w, 400, map[string]string{"error": "Enter a campaign name up to 120 characters and description up to 4000 characters."})
			return
		}
		c := Campaign{ID: id(), OwnerID: u, Name: fmt.Sprint(v["name"]), Description: fmt.Sprint(v["description"]), Members: map[string]string{u: "owner"}, CreatedAt: time.Now().UTC()}
		s.db.Campaigns[c.ID] = c
		s.db.save()
		s.json(w, 201, map[string]any{"campaign": c})
		return
	}
	if p == "/api/v1/campaigns" && r.Method == "GET" {
		out := []Campaign{}
		for _, c := range s.db.Campaigns {
			if member(c, u) {
				out = append(out, c)
			}
		}
		s.json(w, 200, map[string]any{"campaigns": out})
		return
	}
	parts := strings.Split(strings.Trim(p, "/"), "/")
	if s.invitationAPI(w, r, u, v) {
		return
	}
	if p == "/api/v1/characters" && r.Method == "POST" {
		c, ok := s.db.Campaigns[text(v, "campaign_id")]
		if !ok || c.OwnerID != u {
			s.json(w, 403, map[string]string{"error": "PERMISSION_DENIED"})
			return
		}
		character := Character{ID: id(), OwnerID: u, CampaignID: fmt.Sprint(v["campaign_id"]), Name: fmt.Sprint(v["name"]), ClassName: fmt.Sprint(v["class_name"]), HP: 100, MaxHP: 100, X: 2, Y: 2}
		s.db.Characters[character.ID] = character
		s.db.save()
		s.json(w, 201, map[string]any{"character": character})
		return
	}
	if p == "/api/v1/maps" && r.Method == "POST" {
		c, ok := s.db.Campaigns[text(v, "campaign_id")]
		if !ok || c.OwnerID != u {
			s.json(w, 403, map[string]string{"error": "PERMISSION_DENIED"})
			return
		}
		m := GameMap{ID: id(), OwnerID: u, CampaignID: fmt.Sprint(v["campaign_id"]), Name: fmt.Sprint(v["name"]), Width: 20, Height: 14, Blocked: []Point{{7, 4}, {7, 5}, {7, 6}, {8, 6}, {9, 6}}}
		s.db.Maps[m.ID] = m
		s.db.save()
		s.json(w, 201, map[string]any{"map": m})
		return
	}
	if len(parts) >= 5 && parts[0] == "api" && parts[2] == "campaigns" && parts[4] == "sessions" && r.Method == "POST" {
		cid := parts[3]
		c, ok := s.db.Campaigns[cid]
		if !ok || c.OwnerID != u {
			s.json(w, 403, map[string]string{"error": "PERMISSION_DENIED"})
			return
		}
		mid := fmt.Sprint(v["map_id"])
		m, exists := s.db.Maps[mid]
		if !exists || m.CampaignID != cid {
			s.json(w, 400, map[string]string{"error": "MAP_NOT_IN_CAMPAIGN"})
			return
		}
		wld := World{Width: 20, Height: 14, Blocked: []Point{{7, 4}, {7, 5}, {7, 6}, {8, 6}, {9, 6}}, Actors: []Actor{{ID: "mayank", Name: "Mayank", X: 2, Y: 2, HP: 82, MaxHP: 100, Movement: 6, AttackBonus: 5, Damage: 8}, {ID: "lira", Name: "Lira", X: 4, Y: 3, HP: 100, MaxHP: 100, Movement: 6, AttackBonus: 4, Damage: 6}, {ID: "bandit-scout", Name: "Bandit Scout", X: 10, Y: 5, HP: 30, MaxHP: 30, Hostile: true}, {ID: "bandit-archer", Name: "Bandit Archer", X: 13, Y: 4, HP: 24, MaxHP: 24, Hostile: true}}, ActiveActorID: "mayank", Round: 1}
		ss := Session{ID: id(), CampaignID: cid, MapID: mid, Status: "active", Version: 1, State: wld, CreatedAt: time.Now().UTC()}
		s.db.Sessions[ss.ID] = ss
		s.db.save()
		s.json(w, 201, map[string]any{"session_id": ss.ID, "snapshot": ss})
		return
	}
	if len(parts) >= 5 && parts[0] == "api" && parts[2] == "sessions" {
		sid := parts[3]
		ss, ok := s.db.Sessions[sid]
		if !ok {
			s.json(w, 404, map[string]string{"error": "NOT_FOUND"})
			return
		}
		campaign := s.db.Campaigns[ss.CampaignID]
		if !member(campaign, u) {
			s.json(w, 403, map[string]string{"error": "PERMISSION_DENIED"})
			return
		}
		if parts[4] == "snapshot" && r.Method == "GET" {
			viewer := u
			if r.URL.Query().Get("view") == "player" && campaign.OwnerID == u {
				viewer = "__player_preview__"
			}
			s.json(w, 200, s.sessionView(ss, viewer))
			return
		}
		if parts[4] == "events" && r.Method == "GET" {
			after := 0
			fmt.Sscanf(r.URL.Query().Get("after"), "%d", &after)
			out := []Event{}
			for _, ev := range s.sessionView(ss, u).Events {
				if ev.Sequence > after {
					out = append(out, ev)
				}
			}
			s.json(w, 200, map[string]any{"events": out})
			return
		}
		if parts[4] == "commands" && r.Method == "POST" {
			commandPayload, ok := v["payload"].(map[string]any)
			if !ok {
				s.json(w, 400, map[string]string{"error": "INVALID_COMMAND"})
				return
			}
			if err := authorizeActor(ss.State, commandPayload, campaign, u); err != nil {
				s.json(w, 403, map[string]string{"error": err.Error()})
				return
			}
			key, fingerprint, handled := s.checkCommand(w, v, commandPayload, u, ss)
			if handled {
				return
			}
			var payload World
			var kind, actorID string
			var data any
			var err error
			var audience []string
			var gmOnly bool
			if extendedCommand(text(commandPayload, "type")) {
				payload, kind, actorID, data, audience, gmOnly, err = s.tableCommand(ss.State, commandPayload, campaign, u)
			} else if text(commandPayload, "type") == "token.assign" {
				payload, kind, actorID, data, err = s.assignActor(ss.State, commandPayload, campaign, u)
			} else if isToolCommand(text(commandPayload, "type")) {
				payload, kind, actorID, data, err = toolRule(ss.State, commandPayload, campaign.OwnerID == u, s.db.Users[u].DisplayName)
			} else {
				payload, kind, actorID, data, err = rule(ss.State, commandPayload)
			}
			if err != nil {
				status := 409
				if strings.HasPrefix(err.Error(), "PERMISSION_DENIED") {
					status = 403
				}
				s.json(w, status, map[string]string{"error": err.Error()})
				return
			}
			gmOnly = gmOnly || privateToolEvent(ss.State, payload, commandPayload)
			audience = sightAudience(ss.State, payload, commandPayload, campaign, audience)
			ss.State = payload
			ss.Version++
			ss.Sequence++
			ss.Events = append(ss.Events, Event{Sequence: ss.Sequence, Type: kind, ActorID: actorID, Payload: data, Audience: audience, GMOnly: gmOnly, CreatedAt: time.Now().UTC()})
			if key != "" {
				s.db.Receipts[key] = CommandReceipt{Fingerprint: fingerprint, Sequence: ss.Sequence, Version: ss.Version}
			}
			s.db.Sessions[sid] = ss
			s.db.save()
			s.broadcastSession(ss)
			s.json(w, 200, map[string]any{"status": "accepted", "sequence_number": ss.Sequence, "state_version": ss.Version, "event_type": kind, "payload": data, "snapshot": s.sessionView(ss, u)})
			return
		}
	}
	s.json(w, 404, map[string]string{"error": "NOT_FOUND"})
}

func main() {
	root, _ := os.Getwd()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	dataFile := os.Getenv("ASTRALIS_DATA_FILE")
	if dataFile == "" {
		dataFile = filepath.Join(root, "data", "astralis.json")
	}
	s := &Server{db: newDB(dataFile), web: filepath.Join(root, "frontend"), design: filepath.Join(root, "Astralis-Game-Table")}
	log.Printf("Astralis Go server listening on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, s))
}
