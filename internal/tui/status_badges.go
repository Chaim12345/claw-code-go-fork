package tui

import (
	"fmt"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// StatusBadge represents a temporary status indicator
type StatusBadge struct {
	Icon      string
	Message   string
	Type      BadgeType
	ExpiresAt time.Time
	Sticky    bool // If true, doesn't auto-expire
}

// BadgeType defines the visual style of a badge
type BadgeType int

const (
	BadgeInfo BadgeType = iota
	BadgeSuccess
	BadgeWarning
	BadgeError
	BadgeProgress
)

// StatusBadgeManager manages temporary status badges
type StatusBadgeManager struct {
	mu         sync.RWMutex
	badges     []StatusBadge
	maxBadges  int
	defaultTTL time.Duration
}

// NewStatusBadgeManager creates a new badge manager
func NewStatusBadgeManager() *StatusBadgeManager {
	return &StatusBadgeManager{
		badges:     []StatusBadge{},
		maxBadges:  5,
		defaultTTL: 5 * time.Second,
	}
}

// Add adds a new status badge
func (m *StatusBadgeManager) Add(icon, message string, badgeType BadgeType) {
	m.AddWithTTL(icon, message, badgeType, m.defaultTTL)
}

// AddWithTTL adds a badge with custom time-to-live
func (m *StatusBadgeManager) AddWithTTL(icon, message string, badgeType BadgeType, ttl time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	badge := StatusBadge{
		Icon:      icon,
		Message:   message,
		Type:      badgeType,
		ExpiresAt: time.Now().Add(ttl),
		Sticky:    false,
	}

	m.badges = append(m.badges, badge)

	// Keep only the most recent badges
	if len(m.badges) > m.maxBadges {
		m.badges = m.badges[len(m.badges)-m.maxBadges:]
	}
}

// AddSticky adds a badge that doesn't auto-expire
func (m *StatusBadgeManager) AddSticky(icon, message string, badgeType BadgeType) {
	m.mu.Lock()
	defer m.mu.Unlock()

	badge := StatusBadge{
		Icon:      icon,
		Message:   message,
		Type:      badgeType,
		ExpiresAt: time.Now().Add(24 * time.Hour), // Far future
		Sticky:    true,
	}

	m.badges = append(m.badges, badge)
}

// RemoveSticky removes all sticky badges
func (m *StatusBadgeManager) RemoveSticky() {
	m.mu.Lock()
	defer m.mu.Unlock()

	filtered := []StatusBadge{}
	for _, badge := range m.badges {
		if !badge.Sticky {
			filtered = append(filtered, badge)
		}
	}
	m.badges = filtered
}

// Clear removes all badges
func (m *StatusBadgeManager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.badges = []StatusBadge{}
}

// Update removes expired badges
func (m *StatusBadgeManager) Update() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	filtered := []StatusBadge{}

	for _, badge := range m.badges {
		if badge.Sticky || now.Before(badge.ExpiresAt) {
			filtered = append(filtered, badge)
		}
	}

	m.badges = filtered
}

// GetActive returns all active badges
func (m *StatusBadgeManager) GetActive() []StatusBadge {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create a copy to avoid race conditions
	badges := make([]StatusBadge, len(m.badges))
	copy(badges, m.badges)
	return badges
}

// HasActive returns true if there are active badges
func (m *StatusBadgeManager) HasActive() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.badges) > 0
}

// Render renders all active badges
func (m *StatusBadgeManager) Render() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.badges) == 0 {
		return ""
	}

	var rendered []string
	for _, badge := range m.badges {
		rendered = append(rendered, m.renderBadge(badge))
	}

	// Join badges horizontally with spacing
	return lipgloss.JoinHorizontal(lipgloss.Left, rendered...)
}

// renderBadge renders a single badge
func (m *StatusBadgeManager) renderBadge(badge StatusBadge) string {
	style := m.getStyleForType(badge.Type)
	content := fmt.Sprintf(" %s %s ", badge.Icon, badge.Message)
	return style.Render(content) + " "
}

// getStyleForType returns the lipgloss style for a badge type
func (m *StatusBadgeManager) getStyleForType(badgeType BadgeType) lipgloss.Style {
	base := lipgloss.NewStyle().
		Padding(0, 1).
		MarginRight(1)

	switch badgeType {
	case BadgeInfo:
		return base.
			Background(lipgloss.Color("33")).
			Foreground(lipgloss.Color("0"))
	case BadgeSuccess:
		return base.
			Background(lipgloss.Color("42")).
			Foreground(lipgloss.Color("0"))
	case BadgeWarning:
		return base.
			Background(lipgloss.Color("214")).
			Foreground(lipgloss.Color("0"))
	case BadgeError:
		return base.
			Background(lipgloss.Color("196")).
			Foreground(lipgloss.Color("15"))
	case BadgeProgress:
		return base.
			Background(lipgloss.Color("99")).
			Foreground(lipgloss.Color("0"))
	default:
		return base.
			Background(lipgloss.Color("240")).
			Foreground(lipgloss.Color("15"))
	}
}

// Convenience methods for common badge types

// ShowInfo shows an info badge
func (m *StatusBadgeManager) ShowInfo(message string) {
	m.Add("ℹ️", message, BadgeInfo)
}

// ShowSuccess shows a success badge
func (m *StatusBadgeManager) ShowSuccess(message string) {
	m.Add("✓", message, BadgeSuccess)
}

// ShowWarning shows a warning badge
func (m *StatusBadgeManager) ShowWarning(message string) {
	m.Add("⚠️", message, BadgeWarning)
}

// ShowError shows an error badge
func (m *StatusBadgeManager) ShowError(message string) {
	m.Add("✗", message, BadgeError)
}

// ShowProgress shows a progress badge
func (m *StatusBadgeManager) ShowProgress(message string) {
	m.Add("⏳", message, BadgeProgress)
}

// ShowSaving shows a saving indicator
func (m *StatusBadgeManager) ShowSaving(filename string) {
	m.Add("💾", fmt.Sprintf("Saving %s", filename), BadgeProgress)
}

// ShowSaved shows a saved confirmation
func (m *StatusBadgeManager) ShowSaved(filename string) {
	m.Add("✓", fmt.Sprintf("Saved %s", filename), BadgeSuccess)
}

// ShowRunning shows a command running indicator
func (m *StatusBadgeManager) ShowRunning(command string) {
	m.AddSticky("▶️", fmt.Sprintf("Running: %s", command), BadgeProgress)
}

// ShowCompleted shows a command completed indicator
func (m *StatusBadgeManager) ShowCompleted(command string) {
	m.RemoveSticky() // Remove running indicator
	m.Add("✓", fmt.Sprintf("Completed: %s", command), BadgeSuccess)
}

// ShowFailed shows a command failed indicator
func (m *StatusBadgeManager) ShowFailed(command string) {
	m.RemoveSticky() // Remove running indicator
	m.Add("✗", fmt.Sprintf("Failed: %s", command), BadgeError)
}

// ShowThinking shows AI thinking indicator
func (m *StatusBadgeManager) ShowThinking() {
	m.AddSticky("🤔", "Thinking...", BadgeProgress)
}

// ShowStreaming shows AI streaming indicator
func (m *StatusBadgeManager) ShowStreaming() {
	m.AddSticky("💬", "Streaming response...", BadgeProgress)
}

// ShowToolUse shows tool use indicator
func (m *StatusBadgeManager) ShowToolUse(toolName string) {
	m.AddSticky("🔧", fmt.Sprintf("Using: %s", toolName), BadgeProgress)
}

// HideProgress removes all progress badges
func (m *StatusBadgeManager) HideProgress() {
	m.mu.Lock()
	defer m.mu.Unlock()

	filtered := []StatusBadge{}
	for _, badge := range m.badges {
		if badge.Type != BadgeProgress {
			filtered = append(filtered, badge)
		}
	}
	m.badges = filtered
}
