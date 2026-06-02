package tui

import (
	"claw-code-go/internal/api"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// ConversationSearch manages search through conversation messages
type ConversationSearch struct {
	query      string
	results    []searchResult
	cursor     int
	visible    bool
	input      textinput.Model
	messages   []api.Message
	maxResults int
}

// searchResult represents a matched message in search results
type searchResult struct {
	index   int
	role    string
	content string
	preview string
}

// NewConversationSearch creates a new conversation search instance
func NewConversationSearch() *ConversationSearch {
	ti := textinput.New()
	ti.Placeholder = "Search conversation..."
	ti.CharLimit = 256
	ti.Width = 60

	return &ConversationSearch{
		query:      "",
		results:    []searchResult{},
		cursor:     0,
		visible:    false,
		input:      ti,
		messages:   []api.Message{},
		maxResults: 20,
	}
}

// Open opens the conversation search with the given messages
func (cs *ConversationSearch) Open(messages []api.Message) {
	cs.visible = true
	cs.messages = messages
	cs.query = ""
	cs.cursor = 0
	cs.input.SetValue("")
	cs.input.Focus()
	cs.search("")
}

// Close closes the conversation search
func (cs *ConversationSearch) Close() {
	cs.visible = false
	cs.input.Blur()
}

// IsVisible returns whether the search is visible
func (cs *ConversationSearch) IsVisible() bool {
	return cs.visible
}

// UpdateQuery updates the search query and refreshes results
func (cs *ConversationSearch) UpdateQuery(query string) {
	cs.query = query
	cs.cursor = 0
	cs.search(query)
}

// search performs search through conversation messages
func (cs *ConversationSearch) search(query string) {
	if query == "" {
		cs.results = []searchResult{}
		return
	}

	lowerQuery := strings.ToLower(query)
	cs.results = []searchResult{}

	for msgIdx, msg := range cs.messages {
		// Extract text content from message
		var content strings.Builder
		for _, block := range msg.Content {
			if block.Type == "text" {
				content.WriteString(block.Text)
				content.WriteString("\n")
			} else if block.Type == "tool_use" {
				// Include tool use info
				content.WriteString("[Tool: " + block.Name + "]\n")
			} else if block.Type == "tool_result" {
				for _, subBlock := range block.Content {
					if subBlock.Type == "text" {
						content.WriteString(subBlock.Text)
						content.WriteString("\n")
					}
				}
			}
		}
		fullText := content.String()

		if strings.Contains(strings.ToLower(fullText), lowerQuery) {
			// Create preview (first 100 chars around match)
			preview := cs.createPreview(fullText, query)
			role := msg.Role
			if role == "user" {
				role = "You"
			} else if role == "assistant" {
				role = "Assistant"
			}

			cs.results = append(cs.results, searchResult{
				index:   msgIdx,
				role:    role,
				content: fullText,
				preview: preview,
			})
		}

		if len(cs.results) >= cs.maxResults {
			break
		}
	}
}

// createPreview creates a preview snippet around the first match
func (cs *ConversationSearch) createPreview(text, query string) string {
	lowerText := strings.ToLower(text)
	lowerQuery := strings.ToLower(query)

	idx := strings.Index(lowerText, lowerQuery)
	if idx == -1 {
		// Fallback: first 80 chars
		if len(text) > 80 {
			return text[:80] + "..."
		}
		return text
	}

	start := idx - 40
	if start < 0 {
		start = 0
	}
	end := idx + len(query) + 40
	if end > len(text) {
		end = len(text)
	}

	preview := text[start:end]
	if start > 0 {
		preview = "..." + preview
	}
	if end < len(text) {
		preview = preview + "..."
	}
	return preview
}

// MoveCursor moves the selection cursor
func (cs *ConversationSearch) MoveCursor(delta int) {
	cs.cursor += delta
	if cs.cursor < 0 {
		cs.cursor = 0
	}
	if cs.cursor >= len(cs.results) {
		cs.cursor = len(cs.results) - 1
	}
	if cs.cursor < 0 {
		cs.cursor = 0
	}
}

// GetSelected returns the selected search result
func (cs *ConversationSearch) GetSelected() *searchResult {
	if cs.cursor >= 0 && cs.cursor < len(cs.results) {
		return &cs.results[cs.cursor]
	}
	return nil
}

// HasResults returns true if there are search results
func (cs *ConversationSearch) HasResults() bool {
	return len(cs.results) > 0
}

// GetSelectedMessage returns the full content of the selected message
func (cs *ConversationSearch) GetSelectedMessage() (int, string) {
	if cs.cursor >= 0 && cs.cursor < len(cs.results) {
		return cs.results[cs.cursor].index, cs.results[cs.cursor].content
	}
	return -1, ""
}

// View renders the conversation search UI
func (cs *ConversationSearch) View(width, height int) string {
	if !cs.visible {
		return ""
	}

	var b strings.Builder

	// Header
	header := convSearchHeaderStyle.Render("💬 Search Conversation")
	b.WriteString(header + "\n\n")

	// Search input
	b.WriteString(cs.input.View() + "\n\n")

	// Results
	if len(cs.results) == 0 {
		if cs.query == "" {
			b.WriteString(statusStyle.Render("  Enter a search term\n"))
		} else {
			b.WriteString(statusStyle.Render("  No matches found\n"))
		}
	} else {
		for i, res := range cs.results {
			cursor := "  "
			style := convSearchItemStyle
			if i == cs.cursor {
				cursor = "▶ "
				style = convSearchSelectedStyle
			}

			// Header line: role and preview
			headerLine := style.Render(res.role + ": " + res.preview)
			b.WriteString(cursor + headerLine + "\n")
		}
	}

	b.WriteString("\n")

	// Footer
	footer := convSearchHintStyle.Render("[↑↓] Navigate  [Enter] Jump to message  [Esc] Cancel")
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

// Conversation search styles
var (
	convSearchHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("33")).
				Bold(true)

	convSearchItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("250"))

	convSearchSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("33")).
				Bold(true)

	convSearchHintStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240"))
)
