package main

import (
	"errors"
	"net/http"
	"sort"
	"strings"
)

// Authority comes from the signed-in account and persisted assignment, never
// from user/role/controller fields submitted in a gameplay payload.
func authorizeActor(world World, payload map[string]any, campaign Campaign, user string) error {
	kind := text(payload, "type")
	if (kind == "sheet.place" || kind == "sheet.link") && campaign.OwnerID != user {
		return errors.New("PERMISSION_DENIED: only the game master can place or link sheet tokens")
	}
	if strings.HasPrefix(kind, "sheet.") && campaign.OwnerID != user {
		for _, sheet := range world.Sheets {
			if sheet.ID == text(payload, "sheet_id") && sheet.ControllerID == user && kind != "sheet.delete" {
				return nil
			}
		}
		return errors.New("PERMISSION_DENIED: this sheet is not assigned to you")
	}
	if kind == "token.assign" {
		if campaign.OwnerID != user {
			return errors.New("PERMISSION_DENIED: only the game master can assign characters")
		}
		return nil
	}
	field := "character_id"
	switch kind {
	case "character.move", "object.interact", "turn.end":
	case "combat.attack":
		field = "attacker_id"
	default:
		return nil // Tool commands retain their separate GM/member checks.
	}
	if campaign.OwnerID == user {
		return nil
	}
	_, a := actor(world, text(payload, field))
	if a == nil || a.Layer == "gm" || a.Layer == "map" || a.Hostile || a.ControllerID == "" || a.ControllerID != user {
		return errors.New("PERMISSION_DENIED: this character is controlled by another player or the game master")
	}
	if kind == "combat.attack" && world.SightEnabled {
		_, target := actor(world, text(payload, "target_id"))
		if target == nil || !sightCells(world, user)[Point{target.X, target.Y}] {
			return errors.New("PERMISSION_DENIED: target is outside your characters' line of sight")
		}
	}
	if kind == "character.move" && (a.Locked || world.MovementPaused) {
		return errors.New("PERMISSION_DENIED: the game master has locked movement for this token or table")
	}
	return nil
}

func (s *Server) assignActor(world World, payload map[string]any, campaign Campaign, user string) (World, string, string, any, error) {
	fail := func(message string) (World, string, string, any, error) {
		return world, "", "", nil, errors.New(message)
	}
	if campaign.OwnerID != user {
		return fail("PERMISSION_DENIED: only the game master can assign characters")
	}
	controller, valid := payload["controller_id"].(string)
	if !valid {
		return fail("INVALID_ASSIGNMENT: choose a campaign player or game master control")
	}
	_, a := actor(world, text(payload, "token_id"))
	if a == nil || a.Hostile {
		return fail("INVALID_ASSIGNMENT: choose a party character")
	}
	name := "Game master"
	if controller != "" {
		profile, exists := s.db.Users[controller]
		if !exists || !member(campaign, controller) {
			return fail("INVALID_ASSIGNMENT: the player must belong to this campaign")
		}
		name = profile.DisplayName
	}
	// Validate before copying/mutating the live actor slice.
	world.Actors = append([]Actor(nil), world.Actors...)
	_, a = actor(world, a.ID)
	a.ControllerID = controller
	a.ControllerName = ""
	if controller != "" {
		a.ControllerName = name
	}
	return world, "token.assigned", user, map[string]any{"token_id": a.ID, "name": a.Name, "controller_id": controller, "controller_name": name}, nil
}

// Called under db.mu. Return only profiles belonging to this campaign; no
// credentials, email addresses or unrelated account data are exposed.
func (s *Server) campaignMembers(w http.ResponseWriter, r *http.Request, user string) bool {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 5 || parts[0] != "api" || parts[1] != "v1" || parts[2] != "campaigns" || parts[4] != "members" || r.Method != "GET" {
		return false
	}
	campaign, exists := s.db.Campaigns[parts[3]]
	if !exists || !member(campaign, user) {
		s.json(w, 403, map[string]string{"error": "PERMISSION_DENIED"})
		return true
	}
	type profile struct {
		ID          string `json:"id"`
		DisplayName string `json:"display_name"`
		Role        string `json:"role"`
	}
	out := []profile{}
	for id, role := range campaign.Members {
		if u, ok := s.db.Users[id]; ok {
			out = append(out, profile{ID: id, DisplayName: u.DisplayName, Role: role})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].DisplayName == out[j].DisplayName {
			return out[i].ID < out[j].ID
		}
		return out[i].DisplayName < out[j].DisplayName
	})
	s.json(w, 200, map[string]any{"members": out})
	return true
}
