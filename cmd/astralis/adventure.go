package main

import (
	"encoding/json"
	"errors"
	"fmt"
	rand "math/rand/v2"
	"strconv"
	"strings"
)

// System-neutral records. No licensed rules or automatic spell effects.
type CharacterSheet struct {
	Entries      []SheetEntry `json:"entries,omitempty"`
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	ControllerID string       `json:"controller_id"`
	Shared       bool         `json:"shared"`
	Details      string       `json:"details"`
	Stats        string       `json:"stats"`
	Inventory    string       `json:"inventory"`
	Spells       string       `json:"spells"`
	Resources    string       `json:"resources"`
}
type SheetEntry struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Capacity int    `json:"capacity"`
	Prepared bool   `json:"prepared"`
}
type Quest struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Objectives  string `json:"objectives"`
	Status      string `json:"status"`
	Shared      bool   `json:"shared"`
}
type RollEntry struct {
	Text   string `json:"text"`
	Weight int    `json:"weight"`
	Drawn  bool   `json:"drawn"`
}
type RollTable struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Shared      bool        `json:"shared"`
	Replacement bool        `json:"replacement"`
	Entries     []RollEntry `json:"entries"`
}
type AdventureFeatures struct {
	Sheets     []CharacterSheet `json:"sheets,omitempty"`
	Quests     []Quest          `json:"quests,omitempty"`
	RollTables []RollTable      `json:"roll_tables,omitempty"`
}

func adventureView(a AdventureFeatures, user string, gm bool) AdventureFeatures {
	if gm {
		return a
	}
	out := AdventureFeatures{}
	for _, x := range a.Sheets {
		if x.Shared || x.ControllerID == user {
			out.Sheets = append(out.Sheets, x)
		}
	}
	for _, x := range a.Quests {
		if x.Shared {
			out.Quests = append(out.Quests, x)
		}
	}
	for _, x := range a.RollTables {
		if x.Shared {
			out.RollTables = append(out.RollTables, x)
		}
	}
	return out
}
func adventureCommand(kind string) bool {
	if contains([]string{"sheet.entry.save", "sheet.entry.use", "sheet.entry.delete"}, kind) {
		return true
	}
	switch kind {
	case "sheet.save", "sheet.delete", "sheet.roll", "quest.save", "quest.delete", "rolltable.save", "rolltable.delete", "rolltable.draw", "rolltable.reset":
		return true
	}
	return false
}

func (s *Server) adventureCommand(original World, v map[string]any, c Campaign, user string) (World, string, string, any, []string, bool, error) {
	fail := func(message string) (World, string, string, any, []string, bool, error) {
		return original, "", "", nil, nil, false, errors.New(message)
	}
	w := World{}
	b, _ := json.Marshal(original)
	_ = json.Unmarshal(b, &w)
	kind := text(v, "type")
	gm := c.OwnerID == user
	data := map[string]any{"by": s.db.Users[user].DisplayName}
	var audience []string
	private := false
	if strings.HasPrefix(kind, "sheet.") {
		index := -1
		for i, x := range w.Sheets {
			if x.ID == text(v, "sheet_id") {
				index = i
			}
		}
		if !gm && (index < 0 || w.Sheets[index].ControllerID != user) {
			return fail("PERMISSION_DENIED: only the assigned player or GM can edit or roll this sheet")
		}
		if kind != "sheet.save" && index < 0 {
			return fail("SHEET_NOT_FOUND")
		}
		switch kind {
		case "sheet.entry.save", "sheet.entry.use", "sheet.entry.delete":
			sheet := &w.Sheets[index]
			entryIndex := -1
			for i, entry := range sheet.Entries {
				if entry.ID == text(v, "entry_id") {
					entryIndex = i
				}
			}
			if kind == "sheet.entry.save" {
				if text(v, "entry_id") != "" && entryIndex < 0 {
					return fail("ENTRY_NOT_FOUND")
				}
				quantity, a := integer(v, "quantity", 0, 9999)
				capacity, b := integer(v, "capacity", 1, 9999)
				prepared, d := v["prepared"].(bool)
				entry := SheetEntry{ID: id(), Kind: text(v, "kind"), Name: text(v, "name"), Quantity: quantity, Capacity: capacity, Prepared: prepared}
				if !a || !b || !d || quantity > capacity || entry.Name == "" || len(entry.Name) > 100 || !contains([]string{"inventory", "spell", "resource"}, entry.Kind) {
					return fail("INVALID_ENTRY: enter a name, kind, and quantity within capacity 1–9999")
				}
				if entryIndex >= 0 {
					entry.ID = sheet.Entries[entryIndex].ID
					sheet.Entries[entryIndex] = entry
				} else {
					if len(sheet.Entries) >= 100 {
						return fail("Maximum 100 tracked entries per sheet")
					}
					sheet.Entries = append(sheet.Entries, entry)
				}
			} else {
				if entryIndex < 0 {
					return fail("ENTRY_NOT_FOUND")
				}
				if kind == "sheet.entry.delete" {
					sheet.Entries = append(sheet.Entries[:entryIndex], sheet.Entries[entryIndex+1:]...)
				} else {
					entry := &sheet.Entries[entryIndex]
					if entry.Quantity <= 0 {
						return fail("No uses remaining")
					}
					if entry.Kind == "spell" && !entry.Prepared {
						return fail("Prepare this spell before using it")
					}
					entry.Quantity--
				}
			}
		case "sheet.delete":
			if !gm {
				return fail("PERMISSION_DENIED: only the GM can delete sheets")
			}
			w.Sheets = append(w.Sheets[:index], w.Sheets[index+1:]...)
		case "sheet.roll":
			roll, err := diceFormula(text(v, "formula"))
			if err != nil {
				return fail(err.Error())
			}
			data = roll
			data["by"] = w.Sheets[index].Name
			kind = "dice.roll"
		case "sheet.save":
			if text(v, "sheet_id") != "" && index < 0 {
				return fail("SHEET_NOT_FOUND")
			}
			x := CharacterSheet{ID: id(), Name: text(v, "name"), ControllerID: text(v, "controller_id"), Details: text(v, "details"), Stats: text(v, "stats"), Inventory: text(v, "inventory"), Spells: text(v, "spells"), Resources: text(v, "resources")}
			shared, ok := v["shared"].(bool)
			if !ok {
				return fail("INVALID_SHEET: shared must be a boolean")
			}
			x.Shared = shared
			if x.Name == "" || len(x.Name) > 100 {
				return fail("INVALID_SHEET: enter a name up to 100 characters")
			}
			for _, val := range []string{x.Details, x.Stats, x.Inventory, x.Spells, x.Resources} {
				if len(val) > 12000 {
					return fail("INVALID_SHEET: each section is limited to 12000 bytes")
				}
			}
			if x.ControllerID != "" && !member(c, x.ControllerID) {
				return fail("INVALID_SHEET: controller must be a campaign member")
			}
			if index >= 0 {
				x.ID = w.Sheets[index].ID
				x.Entries = w.Sheets[index].Entries
				if !gm {
					x.ControllerID = w.Sheets[index].ControllerID
					x.Shared = w.Sheets[index].Shared
				}
				w.Sheets[index] = x
			} else {
				if len(w.Sheets) >= 50 {
					return fail("Maximum 50 sheets")
				}
				w.Sheets = append(w.Sheets, x)
			}
			data["sheet_id"] = x.ID
		}
		// Sheet content remains in the filtered snapshot, never in public history.
		if kind != "dice.roll" {
			private = true
			if index >= 0 && original.Sheets[index].ControllerID != "" {
				audience = []string{c.OwnerID, original.Sheets[index].ControllerID}
				private = false
			}
		}
	} else {
		if !gm {
			return fail("PERMISSION_DENIED: only the GM manages quests and random tables")
		}
		switch kind {
		case "quest.save", "quest.delete":
			index := -1
			for i, x := range w.Quests {
				if x.ID == text(v, "quest_id") {
					index = i
				}
			}
			if kind == "quest.delete" {
				if index < 0 {
					return fail("QUEST_NOT_FOUND")
				}
				w.Quests = append(w.Quests[:index], w.Quests[index+1:]...)
			} else {
				if text(v, "quest_id") != "" && index < 0 {
					return fail("QUEST_NOT_FOUND")
				}
				x := Quest{ID: id(), Title: text(v, "title"), Description: text(v, "description"), Objectives: text(v, "objectives"), Status: text(v, "status")}
				shared, ok := v["shared"].(bool)
				x.Shared = shared
				if !ok || x.Title == "" || len(x.Title) > 100 || len(x.Description) > 12000 || len(x.Objectives) > 12000 || !contains([]string{"planned", "active", "completed", "failed"}, x.Status) {
					return fail("INVALID_QUEST: enter a title and valid status; text sections allow 12000 bytes")
				}
				if index >= 0 {
					x.ID = w.Quests[index].ID
					w.Quests[index] = x
				} else {
					if len(w.Quests) >= 50 {
						return fail("Maximum 50 quests")
					}
					w.Quests = append(w.Quests, x)
				}
				data["quest_id"] = x.ID
			}
			private = true
		case "rolltable.save", "rolltable.delete", "rolltable.draw", "rolltable.reset":
			index := -1
			for i, x := range w.RollTables {
				if x.ID == text(v, "table_id") {
					index = i
				}
			}
			if kind != "rolltable.save" && index < 0 {
				return fail("TABLE_NOT_FOUND")
			}
			if kind == "rolltable.save" {
				if text(v, "table_id") != "" && index < 0 {
					return fail("TABLE_NOT_FOUND")
				}
				x := RollTable{ID: id(), Name: text(v, "name")}
				shared, a := v["shared"].(bool)
				replacement, b := v["replacement"].(bool)
				x.Shared = shared
				x.Replacement = replacement
				if !a || !b || x.Name == "" || len(x.Name) > 100 || len(text(v, "entries")) > 20000 {
					return fail("INVALID_TABLE: enter a name and weighted entries")
				}
				for _, line := range strings.Split(strings.TrimSpace(text(v, "entries")), "\n") {
					pair := strings.SplitN(strings.TrimSpace(line), "|", 2)
					if len(pair) != 2 {
						return fail("Use weight | result, one entry per line")
					}
					weight, err := strconv.Atoi(strings.TrimSpace(pair[0]))
					result := strings.TrimSpace(pair[1])
					if err != nil || weight < 1 || weight > 1000 || result == "" || len(result) > 1000 {
						return fail("Weights must be 1–1000 and results 1–1000 bytes")
					}
					x.Entries = append(x.Entries, RollEntry{Text: result, Weight: weight})
				}
				if len(x.Entries) > 100 {
					return fail("Maximum 100 entries")
				}
				if index >= 0 {
					x.ID = w.RollTables[index].ID
					w.RollTables[index] = x
				} else {
					if len(w.RollTables) >= 30 {
						return fail("Maximum 30 random tables")
					}
					w.RollTables = append(w.RollTables, x)
				}
				data["table_id"] = x.ID
				private = true
			} else if kind == "rolltable.delete" {
				w.RollTables = append(w.RollTables[:index], w.RollTables[index+1:]...)
				private = true
			} else if kind == "rolltable.reset" {
				for i := range w.RollTables[index].Entries {
					w.RollTables[index].Entries[i].Drawn = false
				}
				private = true
			} else {
				x := &w.RollTables[index]
				total := 0
				for _, entry := range x.Entries {
					if !entry.Drawn || x.Replacement {
						total += entry.Weight
					}
				}
				if total == 0 {
					return fail("This table is exhausted. Reset it to draw again.")
				}
				roll := rand.IntN(total)
				for i, entry := range x.Entries {
					if entry.Drawn && !x.Replacement {
						continue
					}
					roll -= entry.Weight
					if roll < 0 {
						if !x.Replacement {
							x.Entries[i].Drawn = true
						}
						data["message"] = fmt.Sprintf("%s: %s", x.Name, entry.Text)
						data["table_id"] = x.ID
						break
					}
				}
				private = !x.Shared
			}
		}
	}
	return w, kind, user, data, audience, private, nil
}
