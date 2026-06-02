package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// QuickAction represents a single quick action
type QuickAction struct {
	ID          string
	Icon        string
	Label       string
	Description string
	Shortcut    string
	Handler     func(m Model) (Model, error)
	Condition   func(m Model) bool // Show only if condition met
}

// QuickActionsMenu manages the quick actions overlay
type QuickActionsMenu struct {
	actions  []QuickAction
	filtered []QuickAction
	cursor   int
	visible  bool
	input    textinput.Model
	query    string
}

// NewQuickActionsMenu creates a new quick actions menu
func NewQuickActionsMenu() *QuickActionsMenu {
	ti := textinput.New()
	ti.Placeholder = "Type to filter actions..."
	ti.CharLimit = 256
	ti.Width = 60

	menu := &QuickActionsMenu{
		actions:  []QuickAction{},
		filtered: []QuickAction{},
		cursor:   0,
		visible:  false,
		input:    ti,
		query:    "",
	}

	// Register default actions
	menu.registerDefaultActions()

	return menu
}

// registerDefaultActions registers the built-in quick actions
func (q *QuickActionsMenu) registerDefaultActions() {
	q.actions = []QuickAction{
		{
			ID:          "retry_last",
			Icon:        "🔄",
			Label:       "Retry last command",
			Description: "Re-run the previous command",
			Shortcut:    "Ctrl+R",
			Handler: func(m Model) (Model, error) {
				history := m.history.GetAll()
				if len(history) > 0 {
					lastCmd := history[len(history)-1]
					m.textarea.SetValue(lastCmd)
				}
				return m, nil
			},
			Condition: func(m Model) bool {
				return len(m.history.GetAll()) > 0
			},
		},
		{
			ID:          "copy_last_response",
			Icon:        "📋",
			Label:       "Copy last response",
			Description: "Copy the assistant's last message to clipboard",
			Shortcut:    "",
			Handler: func(m Model) (Model, error) {
				// Get the last assistant message
				messages := m.loop.Session.Messages
				var lastAssistantMsg string
				for i := len(messages) - 1; i >= 0; i-- {
					if messages[i].Role == "assistant" {
						for _, block := range messages[i].Content {
							if block.Type == "text" {
								lastAssistantMsg += block.Text
							}
						}
						break
					}
				}
				if lastAssistantMsg == "" {
					m.viewBuf += warnStyle.Render("No assistant response to copy\n\n")
				} else if CopyToClipboard(lastAssistantMsg) {
					m.viewBuf += statusStyle.Render("✓ Last response copied to clipboard\n\n")
				} else {
					m.viewBuf += errorStyle.Render("Failed to copy to clipboard (no clipboard tool found)\n\n")
				}
				m = m.refreshViewport()
				return m, nil
			},
			Condition: func(m Model) bool {
				return m.loop != nil && m.loop.MessageCount() > 0
			},
		},
		{
			ID:          "search_conversation",
			Icon:        "🔍",
			Label:       "Search conversation",
			Description: "Search through conversation history",
			Shortcut:    "Ctrl+F",
			Handler: func(m Model) (Model, error) {
				// Open conversation search
				m.conversationSearch.Open(m.loop.Session.Messages)
				m.transitionState(stateConvSearch, "user opened conversation search")
				return m, nil
			},
			Condition: func(m Model) bool {
				return m.loop != nil && m.loop.MessageCount() > 0
			},
		},
		{
			ID:          "load_session",
			Icon:        "📂",
			Label:       "Load session",
			Description: "Open session browser",
			Shortcut:    "",
			Handler: func(m Model) (Model, error) {
				metas, err := m.loop.ListSessionsWithMeta()
				if err != nil {
					return m, err
				}
				converted := make([]runtimeSessionMeta, 0, len(metas))
				for _, meta := range metas {
					converted = append(converted, runtimeSessionMeta{
						id:             meta.ID,
						updated:        meta.UpdatedAt.Format("2006-01-02 15:04:05"),
						messageCount:   meta.MessageCount,
						totalInTokens:  meta.TotalInputTokens,
						totalOutTokens: meta.TotalOutputTokens,
					})
				}
				m.sessionPicker.open(converted)
				m.transitionState(stateSessionPicker, "user opened session picker")
				return m, nil
			},
			Condition: func(m Model) bool {
				return true
			},
		},
		{
			ID:          "clear_conversation",
			Icon:        "🗑️",
			Label:       "Clear conversation",
			Description: "Clear all messages",
			Shortcut:    "",
			Handler: func(m Model) (Model, error) {
				m.loop.ClearSession()
				m.viewBuf = statusStyle.Render("Session cleared.\n\n")
				m.streamBuf = ""
				m.inputTokens = 0
				m.outputTokens = 0
				m = m.refreshViewport()
				return m, nil
			},
			Condition: func(m Model) bool {
				return m.loop != nil && m.loop.MessageCount() > 0
			},
		},
		{
			ID:          "change_model",
			Icon:        "🤖",
			Label:       "Change model",
			Description: "Switch to a different AI model",
			Shortcut:    "",
			Handler: func(m Model) (Model, error) {
				m.state = statePicker
				m.pickerCursor = 0
				return m, nil
			},
			Condition: func(m Model) bool {
				return true
			},
		},
		{
			ID:          "change_theme",
			Icon:        "🎨",
			Label:       "Change theme",
			Description: "Toggle between dark and light theme",
			Shortcut:    "",
			Handler: func(m Model) (Model, error) {
				if currentTheme.Name == "dark" {
					SetTheme(LightTheme)
					m.viewBuf += statusStyle.Render("Theme: light\n\n")
				} else {
					SetTheme(DarkTheme)
					m.viewBuf += statusStyle.Render("Theme: dark\n\n")
				}
				m = m.refreshViewport()
				return m, nil
			},
			Condition: func(m Model) bool {
				return true
			},
		},
		{
			ID:          "show_status",
			Icon:        "ℹ️",
			Label:       "Show status",
			Description: "Display current session info",
			Shortcut:    "",
			Handler: func(m Model) (Model, error) {
				m.viewBuf += statusStyle.Render(fmt.Sprintf(
					"Model: %s\nProvider: %s\nMessages: %d\nTokens: %s in / %s out\nSession: %s\n\n",
					m.cfg.Model,
					m.cfg.ProviderName,
					m.loop.MessageCount(),
					formatNum(m.inputTokens),
					formatNum(m.outputTokens),
					m.loop.Session.ID,
				))
				m = m.refreshViewport()
				return m, nil
			},
			Condition: func(m Model) bool {
				return true
			},
		},
		{
			ID:          "show_help",
			Icon:        "❓",
			Label:       "Show help",
			Description: "Display help and commands",
			Shortcut:    "",
			Handler: func(m Model) (Model, error) {
				m.state = stateHelp
				return m, nil
			},
			Condition: func(m Model) bool {
				return true
			},
		},
		{
			ID:          "background_tasks",
			Icon:        "⚙️",
			Label:       "Background tasks",
			Description: "View running and completed tasks",
			Shortcut:    "Ctrl+B",
			Handler: func(m Model) (Model, error) {
				m.backgroundTaskPanel.Show()
				m.state = stateBackgroundTasks
				return m, nil
			},
			Condition: func(m Model) bool {
				return true
			},
		},
		{
			ID:          "command_palette",
			Icon:        "⌘",
			Label:       "Command palette",
			Description: "Open command palette",
			Shortcut:    "Ctrl+P",
			Handler: func(m Model) (Model, error) {
				m.palette.open(m.cfg.ProviderName)
				m.state = statePalette
				return m, nil
			},
			Condition: func(m Model) bool {
				return true
			},
		},
	}
}

// Open opens the quick actions menu
func (q *QuickActionsMenu) Open() {
	q.visible = true
	q.query = ""
	q.cursor = 0
	q.input.SetValue("")
	q.input.Focus()
	q.filter("")
}

// Close closes the quick actions menu
func (q *QuickActionsMenu) Close() {
	q.visible = false
	q.input.Blur()
}

// IsVisible returns whether the menu is visible
func (q *QuickActionsMenu) IsVisible() bool {
	return q.visible
}

// UpdateQuery updates the filter query
func (q *QuickActionsMenu) UpdateQuery(query string) {
	q.query = query
	q.cursor = 0
	q.filter(query)
}

// filter filters actions based on query
func (q *QuickActionsMenu) filter(query string) {
	if query == "" {
		q.filtered = q.actions
		return
	}

	lower := strings.ToLower(query)
	q.filtered = []QuickAction{}

	for _, action := range q.actions {
		labelMatch := strings.Contains(strings.ToLower(action.Label), lower)
		descMatch := strings.Contains(strings.ToLower(action.Description), lower)

		if labelMatch || descMatch {
			q.filtered = append(q.filtered, action)
		}
	}
}

// MoveCursor moves the selection cursor
func (q *QuickActionsMenu) MoveCursor(delta int) {
	q.cursor += delta
	if q.cursor < 0 {
		q.cursor = 0
	}
	if q.cursor >= len(q.filtered) {
		q.cursor = len(q.filtered) - 1
	}
	if q.cursor < 0 {
		q.cursor = 0
	}
}

// GetSelected returns the currently selected action
func (q *QuickActionsMenu) GetSelected() *QuickAction {
	if q.cursor >= 0 && q.cursor < len(q.filtered) {
		return &q.filtered[q.cursor]
	}
	return nil
}

// HasActions returns true if there are filtered actions
func (q *QuickActionsMenu) HasActions() bool {
	return len(q.filtered) > 0
}

// GetVisibleActions returns actions that should be shown based on conditions
func (q *QuickActionsMenu) GetVisibleActions(m Model) []QuickAction {
	visible := []QuickAction{}
	for _, action := range q.filtered {
		if action.Condition == nil || action.Condition(m) {
			visible = append(visible, action)
		}
	}
	return visible
}

// View renders the quick actions menu
func (q *QuickActionsMenu) View(m Model, width, height int) string {
	if !q.visible {
		return ""
	}

	var b strings.Builder

	// Header
	header := quickActionsHeaderStyle.Render("⚡ Quick Actions")
	b.WriteString(header + "\n\n")

	// Search input
	b.WriteString(q.input.View() + "\n\n")

	// Get visible actions based on conditions
	visibleActions := q.GetVisibleActions(m)

	// Actions
	if len(visibleActions) == 0 {
		if q.query == "" {
			b.WriteString(statusStyle.Render("  No actions available\n"))
		} else {
			b.WriteString(statusStyle.Render("  No matching actions\n"))
		}
	} else {
		for i, action := range visibleActions {
			cursor := "  "
			style := quickActionsItemStyle
			if i == q.cursor {
				cursor = "▶ "
				style = quickActionsSelectedStyle
			}

			line := fmt.Sprintf("%s %s", action.Icon, action.Label)
			if action.Shortcut != "" {
				line += " " + quickActionsShortcutStyle.Render(action.Shortcut)
			}

			b.WriteString(cursor + style.Render(line) + "\n")

			// Show description for selected item
			if i == q.cursor {
				b.WriteString("    " + quickActionsDescStyle.Render(action.Description) + "\n")
			}
		}
	}

	b.WriteString("\n")

	// Footer
	footer := quickActionsHintStyle.Render("[↑↓] Navigate  [Enter] Execute  [Esc] Cancel")
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

// Quick actions styles
var (
	quickActionsHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("33")).
				Bold(true)

	quickActionsItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("250"))

	quickActionsSelectedStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("33")).
					Bold(true)

	quickActionsShortcutStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("240"))

	quickActionsDescStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Italic(true)

	quickActionsHintStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240"))
)
