package main

import (
	"errors"
	"fmt"
	"math"
	rand "math/rand/v2"
)

func systemCommand(k string) bool {
	return contains([]string{"system.select", "system.check", "dice.quick", "fx.burst", "measure.share"}, k)
}

// Profiles configure reference tools, never silently replace the encounter engine.
func (s *Server) systemCommand(original World, v map[string]any, c Campaign, user string) (World, string, string, any, []string, bool, error) {
	w := original
	kind := text(v, "type")
	gm := c.OwnerID == user
	data := map[string]any{"by": s.db.Users[user].DisplayName}
	var audience []string
	fail := func(msg string) (World, string, string, any, []string, bool, error) {
		return original, "", "", nil, nil, false, errors.New(msg)
	}
	visibility := func() bool {
		mode := text(v, "visibility")
		switch mode {
		case "", "all":
		case "gm":
			audience = []string{user, c.OwnerID}
		case "secret":
			audience = []string{user}
		default:
			return false
		}
		data["visibility"] = mode
		data["private"] = len(audience) > 0
		return true
	}
	switch kind {
	case "system.select":
		if !gm {
			return fail("PERMISSION_DENIED: only the game master chooses the system")
		}
		system := text(v, "system")
		if !contains([]string{"astralis", "dnd5e", "dnd5e2024", "warhammer40k", "pf2e", "fate", "percentile"}, system) {
			return fail("INVALID_SYSTEM")
		}
		w.GameSystem = system
		data["system"] = system
		data["message"] = "Game profile selected: " + system + ". Scene scale and encounter combat rules are unchanged."
	case "dice.quick", "system.check":
		if !visibility() {
			return fail("INVALID_VISIBILITY")
		}
		bonus, ok := integer(v, "bonus", -100, 100)
		if !ok {
			return fail("INVALID_BONUS")
		}
		rolls := []int{}
		total := bonus
		formula := ""
		outcome := ""
		if kind == "dice.quick" {
			count, a := integer(v, "count", 1, 20)
			sides, b := integer(v, "sides", 2, 100)
			if !a || !b {
				return fail("INVALID_DICE")
			}
			for i := 0; i < count; i++ {
				r := rand.IntN(sides) + 1
				rolls = append(rolls, r)
				total += r
			}
			formula = fmt.Sprintf("%dd%d%+d", count, sides, bonus)
		} else {
			system := w.GameSystem
			if system == "" {
				system = "astralis"
			}
			data["system"] = system
			target, a := integer(v, "target", 1, 100)
			if !a {
				return fail("INVALID_TARGET: choose 1 to 100")
			}
			mode := text(v, "mode")
			if !contains([]string{"", "normal", "advantage", "disadvantage"}, mode) {
				return fail("INVALID_CHECK_MODE")
			}
			if mode != "" && mode != "normal" && system != "dnd5e" && system != "dnd5e2024" {
				return fail("INVALID_CHECK_MODE: advantage applies to D&D checks")
			}
			switch system {
			case "warhammer40k":
				count, b := integer(v, "count", 1, 100)
				if !b || target > 6 || target < 2 || bonus != 0 {
					return fail("INVALID_POOL: 1 to 100 d6, threshold 2 to 6, bonus 0")
				}
				total = 0
				for i := 0; i < count; i++ {
					r := rand.IntN(6) + 1
					rolls = append(rolls, r)
					if r >= target {
						total++
					}
				}
				formula = fmt.Sprintf("%dd6, %d+", count, target)
				outcome = fmt.Sprintf("%d successes; resolve rerolls, modifiers and subsequent steps at the table", total)
			case "fate":
				for i := 0; i < 4; i++ {
					r := rand.IntN(3) - 1
					rolls = append(rolls, r)
					total += r
				}
				formula = fmt.Sprintf("4dF%+d", bonus)
				outcome = "Compare this total with opposition; apply aspects and costs at the table"
			case "percentile":
				if bonus != 0 {
					return fail("INVALID_BONUS: percentile checks use target only")
				}
				r := rand.IntN(100) + 1
				rolls = append(rolls, r)
				total = r
				formula = fmt.Sprintf("1d100 ≤ %d", target)
				outcome = "Failure"
				if total <= target {
					outcome = "Success"
				}
			default:
				r := rand.IntN(20) + 1
				rolls = append(rolls, r)
				if mode == "advantage" || mode == "disadvantage" {
					r2 := rand.IntN(20) + 1
					rolls = append(rolls, r2)
					if mode == "advantage" {
						r = max(r, r2)
					} else {
						r = min(r, r2)
					}
				}
				total += r
				formula = fmt.Sprintf("1d20%+d (%s)", bonus, mode)
				outcome = "Failure"
				if total >= target {
					outcome = "Success"
				}
				if system == "pf2e" {
					outcome = pathfinderDegree(r, total, target)
				}
			}
			data["target"] = target
			data["outcome"] = outcome
		}
		data["rolls"] = rolls
		data["total"] = total
		data["formula"] = formula
		kind = "dice.roll"
	case "fx.burst", "measure.share":
		if kind == "fx.burst" && !gm {
			return fail("PERMISSION_DENIED: effects are GM managed")
		}
		number := func(key string, upper int) (float64, bool) {
			f, ok := v[key].(float64)
			return f, ok && !math.IsNaN(f) && !math.IsInf(f, 0) && f >= 0 && f <= float64(upper-1)
		}
		x, a := number("x", w.Width)
		y, b := number("y", w.Height)
		if !a || !b {
			return fail("INVALID_CELL")
		}
		tx, ty := x, y
		if _, ok := v["to_x"]; ok {
			var a, b bool
			tx, a = number("to_x", w.Width)
			ty, b = number("to_y", w.Height)
			if !a || !b {
				return fail("INVALID_CELL")
			}
		} else if kind == "measure.share" {
			return fail("INVALID_CELL")
		}
		shape := text(v, "shape")
		if shape == "" {
			shape = "line"
		}
		if !contains([]string{"line", "square", "circle", "cone", "beam"}, shape) {
			return fail("INVALID_SHAPE")
		}
		data["point"] = map[string]float64{"x": x, "y": y}
		data["to"] = map[string]float64{"x": tx, "y": ty}
		data["shape"] = shape
		data["scene_id"] = w.SceneID
		if kind == "fx.burst" {
			effect := text(v, "effect")
			if effect == "" {
				effect = "burst"
			}
			element := text(v, "element")
			if element == "" {
				element = "magic"
			}
			if !contains([]string{"burst", "breath", "beam", "nova"}, effect) || !contains([]string{"fire", "charm", "acid", "death", "holy", "blood", "frost", "slime", "smoke", "water", "magic"}, element) {
				return fail("INVALID_EFFECT")
			}
			data["effect"] = effect
			data["element"] = element
			data["message"] = "Map effect: " + element + " " + effect
		} else {
			dx, dy := math.Abs(tx-x), math.Abs(ty-y)
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
			data["distance"] = distance * scale
			data["unit"] = w.Unit
			data["message"] = fmt.Sprintf("Shared %s measurement: %.2f %s", shape, distance*scale, w.Unit)
		}
		// Geometry can cross hidden cells. Keep all spatial effects GM-only while
		// sight is enabled, avoiding leaks through every event transport/replay.
		if w.SightEnabled {
			audience = []string{c.OwnerID}
			if !gm {
				visible := sightCells(w, user)
				if !visible[Point{int(x), int(y)}] || !visible[Point{int(tx), int(ty)}] {
					return fail("OUTSIDE_SIGHT")
				}
				audience = append(audience, user)
			}
		}
	}
	return w, kind, user, data, audience, false, nil
}

func pathfinderDegree(face, total, target int) string {
	degree := 1
	if total >= target+10 {
		degree = 3
	} else if total >= target {
		degree = 2
	} else if total <= target-10 {
		degree = 0
	}
	if face == 20 {
		degree = min(3, degree+1)
	}
	if face == 1 {
		degree = max(0, degree-1)
	}
	return []string{"Critical failure", "Failure", "Success", "Critical success"}[degree]
}
