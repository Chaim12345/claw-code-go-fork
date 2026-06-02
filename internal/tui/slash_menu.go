package tui

import (
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// slashMenuItem describes a single slash command shown in the
// autocomplete popup when the user types "/". Command is the full
// slash command string (e.g. "/help"), description is the human-
// readable hint, and args shows the expected arguments syntax.
type slashMenuItem struct {
	command     string // e.g. "/help"
	description string // e.g. "Show available commands"
	args        string // e.g. "[key] [value]" — hint for arguments
}

// slashMenu powers the interactive `/` command autocomplete. When the
// user types `/` in the textarea a floating popup appears above the
// input area listing all available slash commands. As the user
// continues typing, the list filters in real-time using the same
// fuzzy matching as the palette. The user can navigate with ↑/↓,
// autocomplete with Tab, and execute with Enter. Esc dismisses the
// popup without clearing the input.
type slashMenu struct {
	active     bool
	triggerCol int    // column where the "/" was typed
	query      string // text after the "/"
	cursor     int
	items      []slashMenuItem
	filtered   []int // indices into items
}

func newSlashMenu() *slashMenu {
	return &slashMenu{
		items: allSlashCommands(),
	}
}

// allSlashCommands returns the canonical list of all slash commands
// available in the TUI. Each entry carries a human-readable
// description and an optional argument-hint string.
func allSlashCommands() []slashMenuItem {
	return []slashMenuItem{
		{"/help", "Show available commands", ""},
		{"/model", "Change the active model (picker)", ""},
		{"/login", "Multi-provider login flow", ""},
		{"/clear", "Clear conversation history", ""},
		{"/theme", "Switch TUI color theme", "dark|light"},
		{"/status", "Show model/provider/session info", ""},
		{"/cost", "Show token usage this session", ""},
		{"/config", "Show or set config values", "[key] [value]"},
		{"/session", "Manage sessions", "list|save|load [name]"},
		{"/sessions", "Browse saved sessions (picker)", ""},
		{"/todo", "Toggle todo list sidebar", ""},
		{"/init", "Create .claude/settings.json", ""},
		{"/exit", "Exit (session auto-saved)", ""},
		{"/quit", "Exit (session auto-saved)", ""},
	}
}

// update scans the textarea value for an in-progress slash command.
// It activates when a "/" is typed as the first character on a new
// line (or at position 0). Returns true if the menu should be shown.
func (s *slashMenu) update(text string) bool {
	// Only activate if the text starts with "/" (slash command at
	// the beginning of input). Multi-line inputs with "/" later in
	// the text should not trigger the menu.
	slash := strings.Index(text, "/")
	if slash == -1 {
		s.active = false
		return false
	}

	// Activate only if "/" is the first non-whitespace character
	// on the current line. Find the start of the current line.
	lineStart := 0
	lastNewline := strings.LastIndex(text[:slash+1], "\n")
	if lastNewline != -1 {
		lineStart = lastNewline + 1
	}
	beforeSlash := text[lineStart:slash]
	if strings.TrimSpace(beforeSlash) != "" {
		// "/" is not the first thing on the line.
		s.active = false
		return false
	}

	// Capture the query (text from "/" to end of last word).
	rest := text[slash+1:]
	end := len(rest)
	for i := 0; i < len(rest); i++ {
		if isWhitespace(rest[i]) {
			end = i
			break
		}
	}
	q := rest[:end]
	if q != s.query {
		s.query = q
		s.filtered = s.filter(q)
		s.cursor = 0
	}
	s.active = true
	s.triggerCol = slash
	return true
}

// insert replaces the /trigger and partial query in text with the
// selected slash command. Returns the new textarea value and the
// new cursor offset (just after the inserted command).
func (s *slashMenu) insert(text string) (string, int) {
	if !s.active || len(s.filtered) == 0 {
		return text, len(text)
	}
	chosen := s.items[s.filtered[s.cursor]]
	// Build the replacement: the full command string including the "/".
	before := text[:s.triggerCol]
	after := text[s.triggerCol+1+len(s.query):]
	newVal := before + chosen.command + " " + after
	cursor := s.triggerCol + len(chosen.command) + 1
	s.active = false
	return newVal, cursor
}

func (s *slashMenu) moveCursor(delta int) {
	if len(s.filtered) == 0 {
		return
	}
	s.cursor += delta
	if s.cursor < 0 {
		s.cursor = 0
	}
	if s.cursor >= len(s.filtered) {
		s.cursor = len(s.filtered) - 1
	}
}

// filter returns a scored list of indices into s.items matching
// query. Uses the same token-based scoring as palette.refilter().
func (s *slashMenu) filter(query string) []int {
	if query == "" {
		// Show all commands in definition order.
		out := make([]int, len(s.items))
		for i := range s.items {
			out[i] = i
		}
		return out
	}
	q := strings.ToLower(strings.TrimSpace(query))
	tokens := strings.Fields(q)
	type scored struct {
		idx   int
		score int
	}
	var hits []scored
	for i, it := range s.items {
		hay := strings.ToLower(it.command + " " + it.description)
		score, ok := scoreMatch(hay, tokens)
		if ok {
			hits = append(hits, scored{i, score})
		}
	}
	sort.SliceStable(hits, func(a, b int) bool { return hits[a].score > hits[b].score })
	out := make([]int, len(hits))
	for i, h := range hits {
		out[i] = h.idx
	}
	return out
}

// view renders the slash command autocomplete popup.
func (s *slashMenu) view(width, height int) string {
	if !s.active {
		return ""
	}
	var b strings.Builder
	b.WriteString(paletteHeaderStyle.Render("  / Slash Commands"))
	b.WriteString("\n")

	if len(s.filtered) == 0 {
		b.WriteString(paletteHintStyle.Render("  no matches"))
	} else {
		maxVisible := 10
		if maxVisible > len(s.filtered) {
			maxVisible = len(s.filtered)
		}
		for i := 0; i < maxVisible; i++ {
			it := s.items[s.filtered[i]]
			marker := "  "
			style := paletteItemStyle
			if i == s.cursor {
				marker = "▶ "
				style = paletteItemSelectedStyle
			}
			b.WriteString(marker)
			b.WriteString(style.Render(it.command))
			if it.description != "" {
				b.WriteString("  ")
				b.WriteString(paletteHintStyle.Render(it.description))
			}
			b.WriteString("\n")
		}
		// Show argument hints for the selected command.
		if s.cursor < len(s.filtered) {
			sel := s.items[s.filtered[s.cursor]]
			if sel.args != "" {
				b.WriteString("\n")
				b.WriteString(paletteHintStyle.Render("  " + sel.command + " " + sel.args))
				b.WriteString("\n")
			}
		}
	}

	b.WriteString("\n")
	b.WriteString(paletteHintStyle.Render("  ↑↓ select  Tab complete  Enter run  Esc cancel"))

	box := paletteBoxStyle.Width(min(72, width-4)).Render(b.String())
	return lipgloss.Place(width, height, lipgloss.Left, lipgloss.Bottom, box)
}
