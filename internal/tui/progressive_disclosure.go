package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// UserProgress tracks user's interaction history for progressive disclosure
type UserProgress struct {
	MessageCount     int             `json:"message_count"`
	FeaturesUnlocked map[string]bool `json:"features_unlocked"`
	FirstUseDate     time.Time       `json:"first_use_date"`
	LastUseDate      time.Time       `json:"last_use_date"`
	FeatureUsage     map[string]int  `json:"feature_usage"`
	DismissedHints   map[string]bool `json:"dismissed_hints"`
}

// ProgressiveDisclosure manages feature revelation based on user experience
type ProgressiveDisclosure struct {
	progress     *UserProgress
	configPath   string
	hintsShown   map[string]bool
	currentLevel int
}

// NewProgressiveDisclosure creates a new progressive disclosure manager
func NewProgressiveDisclosure(configDir string) *ProgressiveDisclosure {
	pd := &ProgressiveDisclosure{
		configPath: filepath.Join(configDir, "user_progress.json"),
		hintsShown: make(map[string]bool),
	}
	pd.load()
	return pd
}

// load loads user progress from disk
func (pd *ProgressiveDisclosure) load() {
	data, err := os.ReadFile(pd.configPath)
	if err != nil {
		// First time user - initialize
		pd.progress = &UserProgress{
			MessageCount:     0,
			FeaturesUnlocked: make(map[string]bool),
			FirstUseDate:     time.Now(),
			LastUseDate:      time.Now(),
			FeatureUsage:     make(map[string]int),
			DismissedHints:   make(map[string]bool),
		}
		return
	}

	var progress UserProgress
	if err := json.Unmarshal(data, &progress); err != nil {
		// Corrupted file - reinitialize
		pd.progress = &UserProgress{
			MessageCount:     0,
			FeaturesUnlocked: make(map[string]bool),
			FirstUseDate:     time.Now(),
			LastUseDate:      time.Now(),
			FeatureUsage:     make(map[string]int),
			DismissedHints:   make(map[string]bool),
		}
		return
	}

	pd.progress = &progress
	pd.progress.LastUseDate = time.Now()
	pd.updateLevel()
}

// Save saves user progress to disk
func (pd *ProgressiveDisclosure) Save() error {
	// Ensure directory exists
	dir := filepath.Dir(pd.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(pd.progress, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(pd.configPath, data, 0644)
}

// IncrementMessageCount increments the message count and updates level
func (pd *ProgressiveDisclosure) IncrementMessageCount() {
	pd.progress.MessageCount++
	pd.updateLevel()
	pd.Save()
}

// RecordFeatureUsage records that a feature was used
func (pd *ProgressiveDisclosure) RecordFeatureUsage(feature string) {
	if pd.progress.FeatureUsage == nil {
		pd.progress.FeatureUsage = make(map[string]int)
	}
	pd.progress.FeatureUsage[feature]++
	pd.Save()
}

// DismissHint marks a hint as dismissed
func (pd *ProgressiveDisclosure) DismissHint(hint string) {
	if pd.progress.DismissedHints == nil {
		pd.progress.DismissedHints = make(map[string]bool)
	}
	pd.progress.DismissedHints[hint] = true
	pd.Save()
}

// updateLevel calculates the current experience level
func (pd *ProgressiveDisclosure) updateLevel() {
	switch {
	case pd.progress.MessageCount < 5:
		pd.currentLevel = 1 // Beginner
	case pd.progress.MessageCount < 15:
		pd.currentLevel = 2 // Intermediate
	case pd.progress.MessageCount < 30:
		pd.currentLevel = 3 // Advanced
	default:
		pd.currentLevel = 4 // Expert
	}
}

// ShouldShowFeature determines if a feature should be visible
func (pd *ProgressiveDisclosure) ShouldShowFeature(feature string) bool {
	// Check if already unlocked
	if pd.progress.FeaturesUnlocked[feature] {
		return true
	}

	// Check if should be unlocked based on level
	shouldUnlock := false
	switch feature {
	// Level 1 (Always visible)
	case "basic_input", "help", "clear":
		shouldUnlock = true

	// Level 2 (After 5 messages)
	case "command_palette", "mention_files", "slash_commands":
		shouldUnlock = pd.currentLevel >= 2

	// Level 3 (After 15 messages)
	case "session_management", "todo_panel", "background_tasks":
		shouldUnlock = pd.currentLevel >= 3

	// Level 4 (After 30 messages)
	case "debug_panel", "advanced_shortcuts", "custom_modes":
		shouldUnlock = pd.currentLevel >= 4

	default:
		// Unknown features are shown by default
		shouldUnlock = true
	}

	if shouldUnlock && !pd.progress.FeaturesUnlocked[feature] {
		pd.progress.FeaturesUnlocked[feature] = true
		pd.Save()
	}

	return shouldUnlock
}

// GetCurrentHint returns the most relevant hint for the current level
func (pd *ProgressiveDisclosure) GetCurrentHint() string {
	// Don't show hints that have been dismissed
	checkDismissed := func(hint string) bool {
		return pd.progress.DismissedHints != nil && pd.progress.DismissedHints[hint]
	}

	// Level-based hints
	switch pd.currentLevel {
	case 1:
		if !checkDismissed("natural_language") {
			return "💡 Tip: You can use natural language! Try 'help', 'clear', or 'switch model'"
		}
		if !checkDismissed("slash_commands") {
			return "💡 Tip: Type / to see available commands"
		}

	case 2:
		if !checkDismissed("command_palette") && pd.ShouldShowFeature("command_palette") {
			return "💡 New feature unlocked! Press Ctrl+P to open the command palette"
		}
		if !checkDismissed("mention_files") && pd.ShouldShowFeature("mention_files") {
			return "💡 Tip: Use @ to mention files and include them in your message"
		}

	case 3:
		if !checkDismissed("background_tasks") && pd.ShouldShowFeature("background_tasks") {
			return "💡 New feature! Press Ctrl+B to view background tasks"
		}
		if !checkDismissed("session_management") && pd.ShouldShowFeature("session_management") {
			return "💡 Tip: Use /session save to save your conversation for later"
		}

	case 4:
		if !checkDismissed("debug_panel") && pd.ShouldShowFeature("debug_panel") {
			return "💡 Advanced: Press Ctrl+D to open the debug panel"
		}
	}

	return ""
}

// GetHintsForState returns contextual hints based on current state
func (pd *ProgressiveDisclosure) GetHintsForState(state string, inputValue string) string {
	// Context-aware hints
	switch state {
	case "input":
		if inputValue == "" && pd.progress.MessageCount == 0 {
			return "💡 Welcome! Type a message or 'help' to get started"
		}
		if len(inputValue) > 0 && inputValue[0] == '@' {
			return "💡 @-mention: Type a filename to include it in context"
		}
		if len(inputValue) > 0 && inputValue[0] == '/' {
			return "💡 Slash command: Press Tab to autocomplete"
		}

	case "busy":
		if pd.ShouldShowFeature("background_tasks") {
			return "💡 Long operation? Press Ctrl+B to view background tasks"
		}
	}

	return ""
}

// GetShortcutHints returns keyboard shortcuts based on experience level
func (pd *ProgressiveDisclosure) GetShortcutHints() []string {
	hints := []string{"Enter=send"}

	if pd.currentLevel >= 1 {
		hints = append(hints, "Ctrl+J=newline")
	}

	if pd.currentLevel >= 2 && pd.ShouldShowFeature("command_palette") {
		hints = append(hints, "Ctrl+P=palette")
	}

	if pd.currentLevel >= 2 {
		hints = append(hints, "↑↓=history")
	}

	if pd.currentLevel >= 3 && pd.ShouldShowFeature("background_tasks") {
		hints = append(hints, "Ctrl+B=tasks")
	}

	if pd.currentLevel >= 3 {
		hints = append(hints, "Shift+Tab=mode")
	}

	if pd.currentLevel >= 4 && pd.ShouldShowFeature("debug_panel") {
		hints = append(hints, "Ctrl+D=debug")
	}

	return hints
}

// GetLevel returns the current experience level
func (pd *ProgressiveDisclosure) GetLevel() int {
	return pd.currentLevel
}

// GetMessageCount returns the total message count
func (pd *ProgressiveDisclosure) GetMessageCount() int {
	return pd.progress.MessageCount
}

// IsFirstTime returns true if this is the user's first session
func (pd *ProgressiveDisclosure) IsFirstTime() bool {
	return pd.progress.MessageCount == 0
}

// GetDaysSinceFirstUse returns days since first use
func (pd *ProgressiveDisclosure) GetDaysSinceFirstUse() int {
	return int(time.Since(pd.progress.FirstUseDate).Hours() / 24)
}
