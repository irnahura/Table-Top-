package main

import (
	"net/http"
	"sort"
	"strings"
	"time"
)

type Invitation struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	MaxUses   int       `json:"max_uses"`
	Uses      int       `json:"uses"`
	Revoked   bool      `json:"revoked"`
}

// Runs under the database lock, making the last available use atomic.
func (s *Server) invitationAPI(w http.ResponseWriter, r *http.Request, user string, v map[string]any) bool {
	p := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	fail := func(status int, message string) bool {
		s.json(w, status, map[string]string{"error": message})
		return true
	}
	if len(p) >= 5 && p[0] == "api" && p[1] == "v1" && p[2] == "campaigns" && p[4] == "invitations" {
		c, ok := s.db.Campaigns[p[3]]
		if !ok || c.OwnerID != user {
			return fail(403, "PERMISSION_DENIED")
		}
		if len(p) == 5 && r.Method == "GET" {
			out := []Invitation{}
			for token, cid := range s.db.Invites {
				if cid == c.ID {
					rule := s.db.InviteRules[token]
					rule.Token = token
					out = append(out, rule)
				}
			}
			sort.Slice(out, func(i, j int) bool { return out[i].Token < out[j].Token })
			s.json(w, 200, map[string]any{"invitations": out})
			return true
		}
		if len(p) == 5 && r.Method == "POST" {
			hours, uses := 168, 10
			for key, target := range map[string]*int{"expires_in_hours": &hours, "max_uses": &uses} {
				if raw, exists := v[key]; exists {
					n, ok := raw.(float64)
					if !ok || n < 1 || n > 720 || n != float64(int(n)) {
						return fail(400, "Invitation limits must be whole numbers from 1 to 720.")
					}
					*target = int(n)
				}
			}
			rule := Invitation{Token: id(), ExpiresAt: time.Now().UTC().Add(time.Duration(hours) * time.Hour), MaxUses: uses}
			s.db.Invites[rule.Token] = c.ID
			s.db.InviteRules[rule.Token] = rule
			s.db.save()
			s.json(w, 201, rule)
			return true
		}
		if len(p) == 6 && r.Method == "DELETE" {
			if s.db.Invites[p[5]] != c.ID {
				return fail(404, "NOT_FOUND")
			}
			rule := s.db.InviteRules[p[5]]
			rule.Token = p[5]
			rule.Revoked = true
			s.db.InviteRules[p[5]] = rule
			s.db.save()
			s.json(w, 200, map[string]bool{"revoked": true})
			return true
		}
		return fail(405, "METHOD_NOT_ALLOWED")
	}
	if r.URL.Path != "/api/v1/invitations/redeem" || r.Method != "POST" {
		return false
	}
	token := text(v, "token")
	cid, ok := s.db.Invites[token]
	if !ok {
		return fail(404, "Invitation not found.")
	}
	c, ok := s.db.Campaigns[cid]
	if !ok {
		return fail(404, "Campaign not found.")
	}
	rule := s.db.InviteRules[token]
	if rule.Revoked || (!rule.ExpiresAt.IsZero() && !time.Now().Before(rule.ExpiresAt)) || (rule.MaxUses > 0 && rule.Uses >= rule.MaxUses) {
		return fail(409, "This invitation has expired, was revoked, or has no uses remaining.")
	}
	// Existing membership (especially owner authority) must never be overwritten.
	if !member(c, user) {
		c.Members[user] = "player"
		s.db.Campaigns[cid] = c
		rule.Token = token
		rule.Uses++
		s.db.InviteRules[token] = rule
		s.db.save()
	}
	s.json(w, 200, map[string]bool{"redeemed": true})
	return true
}
