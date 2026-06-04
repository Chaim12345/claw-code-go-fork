package tui

import (
	"sort"
	"strings"
)

// command is a single user-facing command in the unified command system.
// It can be invoked via slash command (/cmd), the command palette (Ctrl+P),
// or directly through keyboard shortcuts.
type command struct {
	ID          string // unique identifier, also the slash-command name (e.g. "help")
	Label       string // human-readable label for palette/quick actions
	Description string // one-line description of what this does
	Category    string // "session", "config", "view", "action", "system"
	Shortcut    string // optional keyboard shortcut hint (e.g. "Ctrl+P")
	Args        string // optional argument syntax hint (e.g. "[name]")
}

// commandRegistry is the single source of truth for all user-facing commands.
// The slash menu, command palette, and quick actions all derive their content
// from this registry, ensuring consistent labels, descriptions, and behavior.
type commandRegistry struct {
	all []command
}

// newCommandRegistry returns the canonical command list.
func newCommandRegistry() *commandRegistry {
	return &commandRegistry{all: []command{
		{ID: "help", Label: "Show help", Description: "Open the help panel with all commands", Category: "view", Shortcut: "F1"},
		{ID: "model", Label: "Change model", Description: "Open the model picker", Category: "config"},
		{ID: "theme", Label: "Toggle theme", Description: "Switch between dark and light theme", Category: "config", Shortcut: "Ctrl+T"},
		{ID: "status", Label: "Show status", Description: "Provider, model, session info", Category: "view"},
		{ID: "cost", Label: "Show cost", Description: "Token usage and cost for this session", Category: "view"},
		{ID: "config", Label: "Show config", Description: "View or set configuration values", Category: "config", Args: "[key] [value]"},

		// Session management
		{ID: "clear", Label: "Clear session", Description: "Wipe conversation history", Category: "session", Shortcut: "Ctrl+L"},
		{ID: "compact", Label: "Compact session", Description: "Summarize and compress conversation history", Category: "session"},
		{ID: "session save", Label: "Save session", Description: "Save current session to disk", Category: "session", Args: "[name]"},
		{ID: "sessions", Label: "Browse sessions", Description: "Open the session picker to load a session", Category: "session"},
		{ID: "session list", Label: "List sessions", Description: "Show all saved sessions", Category: "session"},

		// Actions
		{ID: "login", Label: "Login", Description: "Connect to an AI provider", Category: "action"},
		{ID: "todo", Label: "Toggle todo panel", Description: "Show/hide the todo list sidebar", Category: "view"},
		{ID: "init", Label: "Initialize project", Description: "Create .claude/settings.json", Category: "action"},

		// System
		{ID: "exit", Label: "Exit", Description: "Quit the application (session auto-saved)", Category: "system", Shortcut: "Ctrl+C"},
	}}
}

// filtered returns commands matching the given query string using fuzzy
// token-based scoring. An empty query returns all commands in order.
func (r *commandRegistry) filtered(query string) []command {
	if query == "" {
		return r.all
	}
	q := strings.ToLower(strings.TrimSpace(query))
	tokens := strings.Fields(q)
	type scored struct {
		idx   int
		score int
	}
	var hits []scored
	for i, c := range r.all {
		hay := strings.ToLower(c.Label + " " + c.Description + " " + c.Category + " " + c.ID)
		score, ok := scoreMatchTokens(hay, tokens)
		if ok {
			hits = append(hits, scored{i, score})
		}
	}
	sort.SliceStable(hits, func(a, b int) bool { return hits[a].score > hits[b].score })
	out := make([]command, len(hits))
	for i, h := range hits {
		out[i] = r.all[h.idx]
	}
	return out
}

// get returns the command with the given ID, or nil if not found.
func (r *commandRegistry) get(id string) *command {
	for i := range r.all {
		if r.all[i].ID == id {
			return &r.all[i]
		}
	}
	return nil
}

// byCategory groups commands by their Category field.
func (r *commandRegistry) byCategory() map[string][]command {
	groups := make(map[string][]command)
	for _, c := range r.all {
		groups[c.Category] = append(groups[c.Category], c)
	}
	return groups
}

// scoreMatchTokens returns a score and match-ok for the given tokens
// against a haystack string. Higher scores indicate better matches
// (prefix matches score higher than substring matches).
func scoreMatchTokens(hay string, tokens []string) (int, bool) {
	score := 0
	for _, t := range tokens {
		if !strings.Contains(hay, t) {
			return 0, false
		}
		if strings.HasPrefix(hay, t) {
			score += 5
		} else {
			score += 1
		}
	}
	return score, true
}

// categoryOrder defines the display order for command categories.
var categoryOrder = []string{"action", "config", "session", "view", "system"}

// categoryLabel returns a human-readable label for a category.
func categoryLabel(cat string) string {
	switch cat {
	case "action":
		return "Actions"
	case "config":
		return "Configuration"
	case "session":
		return "Session"
	case "view":
		return "View & Info"
	case "system":
		return "System"
	default:
		return cat
	}
}