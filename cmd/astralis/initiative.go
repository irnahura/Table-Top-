package main

import (
	"encoding/json"
	"errors"
	"sort"
)

// Reordering does not grant extra actions or reset the active token's budget.
func (s *Server) manageInitiative(original World, v map[string]any, c Campaign, user string) (World, string, string, any, []string, bool, error) {
	fail := func(msg string) (World, string, string, any, []string, bool, error) {
		return original, "", "", nil, nil, false, errors.New(msg)
	}
	if c.OwnerID != user {
		return fail("PERMISSION_DENIED: only the GM edits turn order")
	}
	w := World{}
	b, _ := json.Marshal(original)
	_ = json.Unmarshal(b, &w)
	action := text(v, "action")
	token := text(v, "token_id")
	index := -1
	for i, x := range w.Initiative {
		if x.TokenID == token {
			index = i
		}
	}
	switch action {
	case "visibility":
		visible, ok := v["visible"].(bool)
		if !ok {
			return fail("INVALID_TRACKER")
		}
		w.TrackerVisible = visible
	case "clear":
		w.Initiative = nil
	case "score":
		_, a := actor(w, token)
		score, ok := integer(v, "score", -100, 100)
		if !ok || a == nil || a.Hostile || a.HP <= 0 || hiddenLayer(a.Layer) || a.Layer == "map" {
			return fail("Choose a living party character and score from -100 to 100")
		}
		if index < 0 {
			w.Initiative = append(w.Initiative, Initiative{TokenID: token, Score: score})
		} else {
			w.Initiative[index].Score = score
		}
	case "remove":
		if index < 0 {
			return fail("TURN_NOT_FOUND")
		}
		w.Initiative = append(w.Initiative[:index], w.Initiative[index+1:]...)
	case "up", "down":
		if index < 0 {
			return fail("TURN_NOT_FOUND")
		}
		next := index - 1
		if action == "down" {
			next = index + 1
		}
		if next < 0 || next >= len(w.Initiative) {
			return fail("Turn is already at the end of the list")
		}
		w.Initiative[index], w.Initiative[next] = w.Initiative[next], w.Initiative[index]
	case "before":
		if index < 0 {
			return fail("TURN_NOT_FOUND")
		}
		target := text(v, "before_id")
		found := -1
		for i, x := range w.Initiative {
			if x.TokenID == target {
				found = i
			}
		}
		if found < 0 || target == token {
			return fail("INVALID_TURN_ORDER")
		}
		entry := w.Initiative[index]
		w.Initiative = append(w.Initiative[:index], w.Initiative[index+1:]...)
		found = 0
		for i, x := range w.Initiative {
			if x.TokenID == target {
				found = i
			}
		}
		w.Initiative = append(w.Initiative, Initiative{})
		copy(w.Initiative[found+1:], w.Initiative[found:])
		w.Initiative[found] = entry
	case "sort":
		sort.SliceStable(w.Initiative, func(i, j int) bool { return w.Initiative[i].Score > w.Initiative[j].Score })
	default:
		return fail("INVALID_TRACKER_ACTION")
	}
	return w, "initiative.manage", user, map[string]any{"by": s.db.Users[user].DisplayName, "message": "Turn tracker updated."}, nil, false, nil
}
