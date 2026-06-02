package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// HistorySearch manages fuzzy search through command history
type HistorySearch struct {
	query      string
	results    []string
	cursor     int
	visible    bool
	input      textinput.Model
	history    []string
	maxResults int
}

// NewHistorySearch creates a new history search instance
func NewHistorySearch() *HistorySearch {
	ti := textinput.New()
	ti.Placeholder = "Search history..."
	ti.CharLimit = 256
	ti.Width = 60

	return &HistorySearch{
		query:      "",
		results:    []string{},
		cursor:     0,
		visible:    false,
		input:      ti,
		history:    []string{},
		maxResults: 10,
	}
}

// Open opens the history search with the given history
func (h *HistorySearch) Open(history []string) {
	h.visible = true
	h.history = history
	h.query = ""
	h.cursor = 0
	h.input.SetValue("")
	h.input.Focus()
	h.search("")
}

// Close closes the history search
func (h *HistorySearch) Close() {
	h.visible = false
	h.input.Blur()
}

// IsVisible returns whether the search is visible
func (h *HistorySearch) IsVisible() bool {
	return h.visible
}

// UpdateQuery updates the search query and refreshes results
func (h *HistorySearch) UpdateQuery(query string) {
	h.query = query
	h.cursor = 0
	h.search(query)
}

// search performs fuzzy search on history
func (h *HistorySearch) search(query string) {
	if query == "" {
		// Show recent history when no query
		h.results = h.getRecentHistory(h.maxResults)
		return
	}

	lower := strings.ToLower(query)
	var matches []searchMatch

	for _, cmd := range h.history {
		score := h.fuzzyScore(cmd, lower)
		if score > 0 {
			matches = append(matches, searchMatch{
				command: cmd,
				score:   score,
			})
		}
	}

	// Sort by score (highest first)
	for i := 0; i < len(matches); i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[j].score > matches[i].score {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	// Extract commands
	h.results = []string{}
	for i := 0; i < len(matches) && i < h.maxResults; i++ {
		h.results = append(h.results, matches[i].command)
	}
}

// searchMatch represents a search result with score
type searchMatch struct {
	command string
	score   int
}

// fuzzyScore calculates a fuzzy match score
func (h *HistorySearch) fuzzyScore(text, query string) int {
	text = strings.ToLower(text)

	// Exact match gets highest score
	if text == query {
		return 1000
	}

	// Contains match gets high score
	if strings.Contains(text, query) {
		// Bonus for match at start
		if strings.HasPrefix(text, query) {
			return 800
		}
		return 500
	}

	// Fuzzy match - all query chars must appear in order
	score := 0
	textIdx := 0
	queryIdx := 0

	for textIdx < len(text) && queryIdx < len(query) {
		if text[textIdx] == query[queryIdx] {
			score += 10
			// Bonus for consecutive matches
			if queryIdx > 0 && textIdx > 0 && text[textIdx-1] == query[queryIdx-1] {
				score += 5
			}
			queryIdx++
		}
		textIdx++
	}

	// All query chars must match
	if queryIdx != len(query) {
		return 0
	}

	return score
}

// getRecentHistory returns the most recent N commands
func (h *HistorySearch) getRecentHistory(n int) []string {
	if len(h.history) == 0 {
		return []string{}
	}

	start := 0
	if len(h.history) > n {
		start = len(h.history) - n
	}

	// Return in reverse order (most recent first)
	recent := make([]string, 0, n)
	for i := len(h.history) - 1; i >= start; i-- {
		recent = append(recent, h.history[i])
	}

	return recent
}

// MoveCursor moves the selection cursor
func (h *HistorySearch) MoveCursor(delta int) {
	h.cursor += delta
	if h.cursor < 0 {
		h.cursor = 0
	}
	if h.cursor >= len(h.results) {
		h.cursor = len(h.results) - 1
	}
	if h.cursor < 0 {
		h.cursor = 0
	}
}

// GetSelected returns the currently selected command
func (h *HistorySearch) GetSelected() string {
	if h.cursor >= 0 && h.cursor < len(h.results) {
		return h.results[h.cursor]
	}
	return ""
}

// HasResults returns true if there are search results
func (h *HistorySearch) HasResults() bool {
	return len(h.results) > 0
}

// View renders the history search UI
func (h *HistorySearch) View(width, height int) string {
	if !h.visible {
		return ""
	}

	var b strings.Builder

	// Header
	header := historySearchHeaderStyle.Render("🔍 Command History")
	b.WriteString(header + "\n\n")

	// Search input
	b.WriteString(h.input.View() + "\n\n")

	// Results
	if len(h.results) == 0 {
		if h.query == "" {
			b.WriteString(statusStyle.Render("  No history yet\n"))
		} else {
			b.WriteString(statusStyle.Render("  No matches found\n"))
		}
	} else {
		for i, cmd := range h.results {
			cursor := "  "
			style := historySearchItemStyle
			if i == h.cursor {
				cursor = "▶ "
				style = historySearchSelectedStyle
			}

			// Truncate long commands
			display := cmd
			if len(display) > 70 {
				display = display[:67] + "..."
			}

			b.WriteString(cursor + style.Render(display) + "\n")
		}
	}

	b.WriteString("\n")

	// Footer
	footer := historySearchHintStyle.Render("[↑↓] Navigate  [Enter] Select  [Esc] Cancel")
	b.WriteString(footer)

	// Wrap in box
	content := b.String()
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("33")).
		Padding(1, 2).
		Width(min(width-4, 80))

	box := boxStyle.Render(content)

	// Center on screen
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}

// History search styles
var (
	historySearchHeaderStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("33")).
					Bold(true)

	historySearchItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("250"))

	historySearchSelectedStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("33")).
					Bold(true)

	historySearchHintStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240"))
)
