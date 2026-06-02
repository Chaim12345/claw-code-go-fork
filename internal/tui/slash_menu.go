package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// slashMenu powers the interactive `/` command autocomplete. When the
// user types `/` in the textarea a floating popup appears above the
// input area listing all available slash commands. As the user
// continues typing, the list filters in real-time using fuzzy matching
// backed by the unified commandRegistry. The user can navigate with
// ↑/↓, autocomplete with Tab, and execute with Enter. Esc dismisses
// the popup without clearing the input.
type slashMenu struct {
	active     bool
	triggerCol int // column where the "/" was typed
	query      string
	cursor     int
	registry   *commandRegistry
	filtered   []command
}

func newSlashMenu() *slashMenu {
	return &slashMenu{
		registry: newCommandRegistry(),
	}
}

// update scans the textarea value for an in-progress slash command.
// It activates when a "/" is typed as the first character on a new
// line (or at position 0). Returns true if the menu should be shown.
func (s *slashMenu) update(text string) bool {
	// Only activate if the text starts with "/" (slash command at
	// the beginning of input).
	slash := strings.Index(text, "/")
	if slash == -1 {
		s.active = false
		return false
	}

	// Activate only if "/" is the first non-whitespace character
	// on the current line.
	lineStart := 0
	lastNewline := strings.LastIndex(text[:slash+1], "\n")
	if lastNewline != -1 {
		lineStart = lastNewline + 1
	}
	beforeSlash := text[lineStart:slash]
	if strings.TrimSpace(beforeSlash) != "" {
		s.active = false
		return false
	}

	// Capture the query (text from "/" to end of word).
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
		s.filtered = s.registry.filtered(q)
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
	chosen := s.filtered[s.cursor]
	cmdStr := "/" + chosen.ID
	before := text[:s.triggerCol]
	after := text[s.triggerCol+1+len(s.query):]
	newVal := before + cmdStr + " " + after
	cursor := s.triggerCol + len(cmdStr) + 1
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

// view renders the slash command autocomplete popup. Commands are
// grouped by category for easier scanning.
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
		// Group filtered commands by category
		groups := groupCommandsByCategory(s.filtered)
		lineNum := 0
		maxVisible := 12
		for _, cat := range categoryOrder {
			cmds, ok := groups[cat]
			if !ok || len(cmds) == 0 {
				continue
			}
			// Section header
			if lineNum > 0 {
				b.WriteString("\n")
			}
			b.WriteString(paletteHintStyle.Render("  " + categoryLabel(cat)))
			b.WriteString("\n")
			for _, cmd := range cmds {
				if lineNum >= maxVisible {
					break
				}
				marker := "  "
				style := paletteItemStyle
				if lineNum == s.cursor {
					marker = "▶ "
					style = paletteItemSelectedStyle
				}
				b.WriteString(marker)
				b.WriteString(style.Render("/" + cmd.ID))
				if cmd.Description != "" {
					b.WriteString("  ")
					b.WriteString(paletteHintStyle.Render(cmd.Description))
				}
				if cmd.Shortcut != "" {
					b.WriteString(" ")
					b.WriteString(paletteHintStyle.Render("[" + cmd.Shortcut + "]"))
				}
				b.WriteString("\n")
				lineNum++
			}
			if lineNum >= maxVisible {
				break
			}
		}
		// Show argument hints for the selected command.
		if s.cursor < len(s.filtered) {
			sel := s.filtered[s.cursor]
			if sel.Args != "" {
				b.WriteString("\n")
				b.WriteString(paletteHintStyle.Render("  /" + sel.ID + " " + sel.Args))
				b.WriteString("\n")
			}
		}
	}

	b.WriteString("\n")
	b.WriteString(paletteHintStyle.Render("  ↑↓ select  Tab complete  Enter run  Esc cancel"))

	box := paletteBoxStyle.Width(min(72, width-4)).Render(b.String())
	return lipgloss.Place(width, height, lipgloss.Left, lipgloss.Bottom, box)
}

// groupCommandsByCategory groups a slice of commands by their Category field.
func groupCommandsByCategory(cmds []command) map[string][]command {
	groups := make(map[string][]command)
	for _, c := range cmds {
		groups[c.Category] = append(groups[c.Category], c)
	}
	return groups
}