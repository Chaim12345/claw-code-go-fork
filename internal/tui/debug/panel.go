package debug

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Panel represents the debug panel UI
type Panel struct {
	visible      bool
	width        int
	height       int
	selectedTab  int
	scrollOffset int
	filterType   EventType
	showMetrics  bool
}

// Tab represents a debug panel tab
type Tab struct {
	Name string
	Icon string
}

var debugTabs = []Tab{
	{"Events", "📋"},
	{"States", "🔄"},
	{"Metrics", "📊"},
	{"Tools", "🔧"},
}

// NewPanel creates a new debug panel
func NewPanel() *Panel {
	return &Panel{
		visible:     false,
		selectedTab: 0,
	}
}

// Toggle toggles the panel visibility
func (p *Panel) Toggle() {
	p.visible = !p.visible
}

// IsVisible returns whether the panel is visible
func (p *Panel) IsVisible() bool {
	return p.visible
}

// SetSize sets the panel dimensions
func (p *Panel) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// NextTab moves to the next tab
func (p *Panel) NextTab() {
	p.selectedTab = (p.selectedTab + 1) % len(debugTabs)
	p.scrollOffset = 0
}

// PrevTab moves to the previous tab
func (p *Panel) PrevTab() {
	p.selectedTab--
	if p.selectedTab < 0 {
		p.selectedTab = len(debugTabs) - 1
	}
	p.scrollOffset = 0
}

// ScrollUp scrolls the content up
func (p *Panel) ScrollUp() {
	if p.scrollOffset > 0 {
		p.scrollOffset--
	}
}

// ScrollDown scrolls the content down
func (p *Panel) ScrollDown() {
	p.scrollOffset++
}

// SetFilter sets the event type filter
func (p *Panel) SetFilter(eventType EventType) {
	p.filterType = eventType
	p.scrollOffset = 0
}

// ClearFilter clears the event type filter
func (p *Panel) ClearFilter() {
	p.filterType = ""
	p.scrollOffset = 0
}

// Render renders the debug panel
func (p *Panel) Render() string {
	if !p.visible {
		return ""
	}

	// Styles
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("cyan")).
		Bold(true).
		Padding(0, 1)

	tabStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(lipgloss.Color("240"))

	activeTabStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(lipgloss.Color("cyan")).
		Bold(true).
		Underline(true)

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("cyan")).
		Padding(1, 2)

	// Header
	header := headerStyle.Render("🐛 Debug Panel")

	// Tabs
	var tabs []string
	for i, tab := range debugTabs {
		style := tabStyle
		if i == p.selectedTab {
			style = activeTabStyle
		}
		tabs = append(tabs, style.Render(tab.Icon+" "+tab.Name))
	}
	tabBar := lipgloss.JoinHorizontal(lipgloss.Left, tabs...)

	// Content based on selected tab
	var content string
	switch p.selectedTab {
	case 0:
		content = p.renderEventsTab()
	case 1:
		content = p.renderStatesTab()
	case 2:
		content = p.renderMetricsTab()
	case 3:
		content = p.renderToolsTab()
	}

	// Footer with controls
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("Tab/Shift+Tab: switch tabs  •  ↑↓: scroll  •  f: filter  •  c: clear  •  Esc: close")

	// Combine all parts
	panel := lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		tabBar,
		"",
		content,
		"",
		footer,
	)

	return borderStyle.Width(p.width - 4).Height(p.height - 4).Render(panel)
}

// renderEventsTab renders the events tab content
func (p *Panel) renderEventsTab() string {
	events := GetRecentEvents(100)

	if p.filterType != "" {
		filtered := make([]Event, 0)
		for _, e := range events {
			if e.Type == p.filterType {
				filtered = append(filtered, e)
			}
		}
		events = filtered
	}

	if len(events) == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("No events to display")
	}

	// Apply scroll offset
	if p.scrollOffset >= len(events) {
		p.scrollOffset = len(events) - 1
	}
	if p.scrollOffset < 0 {
		p.scrollOffset = 0
	}

	var lines []string
	maxLines := p.height - 12 // Account for header, tabs, footer

	start := p.scrollOffset
	end := start + maxLines
	if end > len(events) {
		end = len(events)
	}

	for i := start; i < end; i++ {
		e := events[i]
		lines = append(lines, p.formatEvent(e))
	}

	if len(lines) == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("No events in view")
	}

	return strings.Join(lines, "\n")
}

// formatEvent formats a single event for display
func (p *Panel) formatEvent(e Event) string {
	// Time
	timeStr := e.Timestamp.Format("15:04:05.000")
	timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	// Type with color
	var typeColor lipgloss.Color
	switch e.Type {
	case EventError:
		typeColor = lipgloss.Color("red")
	case EventWarning:
		typeColor = lipgloss.Color("yellow")
	case EventStateChange:
		typeColor = lipgloss.Color("cyan")
	case EventToolCall:
		typeColor = lipgloss.Color("green")
	case EventStreamChunk:
		typeColor = lipgloss.Color("blue")
	default:
		typeColor = lipgloss.Color("white")
	}
	typeStyle := lipgloss.NewStyle().Foreground(typeColor).Width(15)

	// Message
	msg := e.Message
	if len(msg) > 60 {
		msg = msg[:57] + "..."
	}

	// Duration if present
	durationStr := ""
	if e.Duration > 0 {
		durationStr = fmt.Sprintf(" (%s)", e.Duration)
	}

	return fmt.Sprintf("%s %s %s%s",
		timeStyle.Render(timeStr),
		typeStyle.Render(string(e.Type)),
		msg,
		durationStr,
	)
}

// renderStatesTab renders the state transitions tab
func (p *Panel) renderStatesTab() string {
	history := GetStateHistory()

	if len(history) == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("No state transitions recorded")
	}

	var lines []string
	maxLines := p.height - 12

	// Show most recent transitions
	start := len(history) - maxLines - p.scrollOffset
	if start < 0 {
		start = 0
	}
	end := len(history) - p.scrollOffset
	if end > len(history) {
		end = len(history)
	}

	for i := start; i < end; i++ {
		t := history[i]
		lines = append(lines, p.formatTransition(t))
	}

	// Add current state info
	current := GetCurrentState()
	timeInState := GetTimeInCurrentState()

	currentInfo := lipgloss.NewStyle().
		Foreground(lipgloss.Color("cyan")).
		Bold(true).
		Render(fmt.Sprintf("\nCurrent: %s (for %s)", current, timeInState.Round(time.Millisecond)))

	return strings.Join(lines, "\n") + "\n" + currentInfo
}

// formatTransition formats a state transition for display
func (p *Panel) formatTransition(t StateTransition) string {
	timeStr := t.Timestamp.Format("15:04:05.000")
	timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	fromStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("yellow"))
	toStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("green"))
	arrowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	durationStr := ""
	if t.Duration > 0 {
		durationStr = fmt.Sprintf(" (%s)", t.Duration.Round(time.Millisecond))
	}

	return fmt.Sprintf("%s %s %s %s%s",
		timeStyle.Render(timeStr),
		fromStyle.Render(string(t.From)),
		arrowStyle.Render("→"),
		toStyle.Render(string(t.To)),
		durationStr,
	)
}

// renderMetricsTab renders the metrics tab
func (p *Panel) renderMetricsTab() string {
	metrics := GetStateMetrics()
	stats := GetStats()

	if len(metrics) == 0 && len(stats) == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("No metrics available")
	}

	var lines []string

	// State metrics
	lines = append(lines, lipgloss.NewStyle().
		Foreground(lipgloss.Color("cyan")).
		Bold(true).
		Render("State Metrics:"))
	lines = append(lines, "")

	for state, m := range metrics {
		lines = append(lines, fmt.Sprintf("  %s:", state))
		lines = append(lines, fmt.Sprintf("    Enters: %d", m.EnterCount))
		lines = append(lines, fmt.Sprintf("    Total time: %s", m.TotalDuration.Round(time.Millisecond)))
		lines = append(lines, fmt.Sprintf("    Avg time: %s", m.AvgDuration.Round(time.Millisecond)))
		lines = append(lines, fmt.Sprintf("    Max time: %s", m.MaxDuration.Round(time.Millisecond)))
		lines = append(lines, "")
	}

	// Event stats
	lines = append(lines, lipgloss.NewStyle().
		Foreground(lipgloss.Color("cyan")).
		Bold(true).
		Render("Event Statistics:"))
	lines = append(lines, "")

	for eventType, count := range stats {
		lines = append(lines, fmt.Sprintf("  %s: %d", eventType, count))
	}

	return strings.Join(lines, "\n")
}

// renderToolsTab renders the tools tab
func (p *Panel) renderToolsTab() string {
	toolEvents := GetEventsByType(EventToolCall)

	if len(toolEvents) == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("No tool calls recorded")
	}

	var lines []string
	maxLines := p.height - 12

	start := len(toolEvents) - maxLines - p.scrollOffset
	if start < 0 {
		start = 0
	}
	end := len(toolEvents) - p.scrollOffset
	if end > len(toolEvents) {
		end = len(toolEvents)
	}

	for i := start; i < end; i++ {
		e := toolEvents[i]
		lines = append(lines, p.formatToolEvent(e))
	}

	return strings.Join(lines, "\n")
}

// formatToolEvent formats a tool call event for display
func (p *Panel) formatToolEvent(e Event) string {
	timeStr := e.Timestamp.Format("15:04:05.000")
	timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	toolName := "unknown"
	if name, ok := e.Data["tool_name"].(string); ok {
		toolName = name
	}

	toolStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("green"))

	durationStr := ""
	if e.Duration > 0 {
		durationStr = fmt.Sprintf(" (%s)", e.Duration.Round(time.Millisecond))
	}

	return fmt.Sprintf("%s %s%s",
		timeStyle.Render(timeStr),
		toolStyle.Render(toolName),
		durationStr,
	)
}

// Global panel instance
var globalPanel = NewPanel()

// Global convenience functions
func TogglePanel() {
	globalPanel.Toggle()
}

func IsPanelVisible() bool {
	return globalPanel.IsVisible()
}

func SetPanelSize(width, height int) {
	globalPanel.SetSize(width, height)
}

func NextPanelTab() {
	globalPanel.NextTab()
}

func PrevPanelTab() {
	globalPanel.PrevTab()
}

func ScrollPanelUp() {
	globalPanel.ScrollUp()
}

func ScrollPanelDown() {
	globalPanel.ScrollDown()
}

func RenderPanel() string {
	return globalPanel.Render()
}
