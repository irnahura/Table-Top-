package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type CommandReceipt struct {
	Fingerprint string `json:"fingerprint"`
	Sequence    int    `json:"sequence"`
	Version     int    `json:"version"`
}

// Legacy clients may omit both fields. Guarded commands must supply both.
// Membership is checked before this method; retries never bypass authorization.
func (s *Server) checkCommand(w http.ResponseWriter, v, payload map[string]any, user string, ss Session) (string, string, bool) {
	_, hasID := v["command_id"]
	_, hasVersion := v["expected_state_version"]
	if !hasID && !hasVersion {
		return "", "", false
	}
	cid := text(v, "command_id")
	version, valid := integer(v, "expected_state_version", 1, 1<<53-1)
	if !hasID || !hasVersion || cid == "" || len(cid) > 128 || !valid {
		s.json(w, 400, map[string]string{"error": "Provide command_id (1–128 characters) and an integer expected_state_version."})
		return "", "", true
	}
	b, _ := json.Marshal(payload)
	fingerprint := hash(string(b))
	key := ss.ID + ":" + user + ":" + cid
	if receipt, exists := s.db.Receipts[key]; exists {
		if receipt.Fingerprint != fingerprint {
			s.json(w, 409, map[string]string{"error": "COMMAND_ID_REUSED: use a new ID for a different action"})
		} else {
			ev := ss.Events[receipt.Sequence-1]
			s.json(w, 200, map[string]any{"status": "accepted", "duplicate": true, "sequence_number": receipt.Sequence, "state_version": receipt.Version, "event_type": ev.Type, "payload": ev.Payload, "snapshot": s.sessionView(ss, user)})
		}
		return "", "", true
	}
	if version != ss.Version {
		s.json(w, 409, map[string]any{"error": "STATE_VERSION_CONFLICT: the table changed; review the refreshed encounter and try again", "snapshot": s.sessionView(ss, user)})
		return "", "", true
	}
	return key, fingerprint, false
}

type livePeer struct {
	connectionID         string
	media                bool
	session, user, token string
	send                 chan []byte
	stop                 chan struct{}
}

// All peer operations run under db.mu. Writes to sockets never hold that lock.
func (s *Server) queue(peer *livePeer, message any) {
	b, _ := json.Marshal(message)
	select {
	case peer.send <- b:
	default:
		// Disconnect slow readers; reconnect replays from the last applied sequence.
		select {
		case <-peer.stop:
		default:
			close(peer.stop)
		}
	}
}

func (s *Server) allowed(peer *livePeer) bool {
	ss, exists := s.db.Sessions[peer.session]
	expiry, valid := s.db.TokenExpiry[peer.token]
	return exists && valid && time.Now().Before(expiry) && s.db.Tokens[peer.token] == peer.user && member(s.db.Campaigns[ss.CampaignID], peer.user)
}

func (s *Server) broadcastSession(ss Session) {
	for peer := range s.peers {
		if peer.session != ss.ID {
			continue
		}
		if !s.allowed(peer) {
			select {
			case <-peer.stop:
			default:
				close(peer.stop)
			}
			continue
		}
		s.queue(peer, map[string]any{"type": "session.updated", "snapshot": s.sessionView(ss, peer.user)})
	}
}

func (s *Server) presence(sid string) {
	users := map[string]string{}
	clients := map[string]any{}
	for peer := range s.peers {
		if peer.session == sid && s.allowed(peer) {
			users[peer.user] = s.db.Users[peer.user].DisplayName
			clients[peer.connectionID] = map[string]any{"user_id": peer.user, "name": s.db.Users[peer.user].DisplayName, "media": peer.media}
		}
	}
	for peer := range s.peers {
		if peer.session == sid && s.allowed(peer) {
			s.queue(peer, map[string]any{"type": "presence", "users": users, "clients": clients})
		}
	}
}

func (s *Server) live(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	after, err := strconv.Atoi(r.URL.Query().Get("after"))
	if len(parts) != 5 || err != nil || after < 0 {
		s.json(w, 400, map[string]string{"error": "INVALID_REPLAY_SEQUENCE"})
		return
	}
	// The default upgrader rejects cross-origin browser connections.
	upgrader := websocket.Upgrader{HandshakeTimeout: 5 * time.Second}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	conn.SetReadLimit(4096)
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	var credentials struct {
		Token string `json:"token"`
	}
	if conn.ReadJSON(&credentials) != nil {
		return
	}
	// Authenticate the first frame so bearer tokens never appear in URLs/logs.
	authRequest := r.Clone(r.Context())
	authRequest.Header = r.Header.Clone()
	authRequest.Header.Set("Authorization", "Bearer "+credentials.Token)
	user := s.auth(authRequest)
	peer := &livePeer{connectionID: id(), session: parts[3], user: user, token: credentials.Token, send: make(chan []byte, 16), stop: make(chan struct{})}
	s.db.mu.Lock()
	if user == "" || !s.allowed(peer) {
		s.db.mu.Unlock()
		_ = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(1008, "Sign in as a campaign member"), time.Now().Add(time.Second))
		return
	}
	if s.peers == nil {
		s.peers = map[*livePeer]bool{}
	}
	s.peers[peer] = true
	ss := s.db.Sessions[peer.session]
	events := []Event{}
	view := s.sessionView(ss, user)
	for _, ev := range view.Events {
		if ev.Sequence > after {
			events = append(events, ev)
		}
	}
	// Subscribe and capture replay in one critical section, with no event gap.
	s.queue(peer, map[string]any{"type": "session.sync", "events": events, "snapshot": view, "connection_id": peer.connectionID})
	s.presence(peer.session)
	s.db.mu.Unlock()
	defer func() { s.db.mu.Lock(); delete(s.peers, peer); s.presence(peer.session); s.db.mu.Unlock() }()
	_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetReadLimit(64 << 10)
	conn.SetPongHandler(func(string) error { return conn.SetReadDeadline(time.Now().Add(60 * time.Second)) })
	go func() {
		defer func() {
			select {
			case <-peer.stop:
			default:
				// Only the writer/queue closes stop under db.mu.
				s.db.mu.Lock()
				select {
				case <-peer.stop:
				default:
					close(peer.stop)
				}
				s.db.mu.Unlock()
			}
		}()
		for {
			var signal struct {
				Type    string          `json:"type"`
				To      string          `json:"to"`
				Payload json.RawMessage `json:"payload"`
			}
			if err := conn.ReadJSON(&signal); err != nil {
				return
			}
			s.db.mu.Lock()
			if !s.allowed(peer) {
				s.db.mu.Unlock()
				return
			}
			if signal.Type == "rtc.join" || signal.Type == "rtc.leave" {
				peer.media = signal.Type == "rtc.join"
				s.presence(peer.session)
			}
			if signal.Type == "rtc.signal" && peer.media && len(signal.Payload) < 60<<10 {
				for target := range s.peers {
					if target.connectionID == signal.To && target.session == peer.session && target.media && s.allowed(target) {
						s.queue(target, map[string]any{"type": "rtc.signal", "from": peer.connectionID, "payload": signal.Payload})
					}
				}
			}
			s.db.mu.Unlock()
		}
	}()
	tick := time.NewTicker(20 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-peer.stop:
			return
		case message := <-peer.send:
			s.db.mu.RLock()
			valid := s.allowed(peer)
			s.db.mu.RUnlock()
			if !valid {
				return
			}
			_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if conn.WriteMessage(websocket.TextMessage, message) != nil {
				return
			}
		case <-tick.C:
			s.db.mu.RLock()
			valid := s.allowed(peer)
			s.db.mu.RUnlock()
			if !valid {
				return
			}
			if conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second)) != nil {
				return
			}
		}
	}
}
