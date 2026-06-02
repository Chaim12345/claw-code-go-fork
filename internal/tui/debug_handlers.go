package tui

import (
	"claw-code-go/internal/tui/debug"

	tea "github.com/charmbracelet/bubbletea"
)

// handleDebugPanelKey handles key input when the debug panel is open
func (m Model) handleDebugPanelKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit

	case tea.KeyEsc, tea.KeyCtrlD:
		// Close debug panel
		debug.TogglePanel()
		m.transitionState(stateInput, "user closed debug panel")
		return m, nil

	case tea.KeyTab:
		// Next tab
		debug.NextPanelTab()
		return m, nil

	case tea.KeyShiftTab:
		// Previous tab
		debug.PrevPanelTab()
		return m, nil

	case tea.KeyUp:
		// Scroll up
		debug.ScrollPanelUp()
		return m, nil

	case tea.KeyDown:
		// Scroll down
		debug.ScrollPanelDown()
		return m, nil

	default:
		// Handle character keys
		switch msg.String() {
		case "c":
			// Clear events
			debug.Clear()
			return m, nil
		case "f":
			// Toggle filter (cycle through event types)
			// This is a simple implementation - could be enhanced
			return m, nil
		case "q":
			// Quit debug panel
			debug.TogglePanel()
			m.transitionState(stateInput, "user closed debug panel")
			return m, nil
		}
	}

	return m, nil
}
