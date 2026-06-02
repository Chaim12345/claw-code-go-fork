package tui

import "github.com/charmbracelet/lipgloss"

// All TUI styles are derived from currentTheme.
// Call rebuildStyles (via SetTheme) to refresh after a theme switch.
var (
	headerStyle              lipgloss.Style
	modelTagStyle            lipgloss.Style
	userLabelStyle           lipgloss.Style
	assistantLabelStyle      lipgloss.Style
	toolRunningStyle         lipgloss.Style
	toolDoneStyle            lipgloss.Style
	toolFailedStyle          lipgloss.Style
	toolExpandedHeaderStyle  lipgloss.Style
	toolInputStyle           lipgloss.Style
	statusStyle              lipgloss.Style
	warnStyle                lipgloss.Style
	errorStyle               lipgloss.Style
	helpBoxStyle             lipgloss.Style
	dividerStyle             lipgloss.Style
	inputPromptStyle         lipgloss.Style
	pickerHeaderStyle        lipgloss.Style
	selectedModelStyle       lipgloss.Style
	unselectedModelStyle     lipgloss.Style
	paletteHeaderStyle       lipgloss.Style
	paletteQueryStyle        lipgloss.Style
	paletteItemStyle         lipgloss.Style
	paletteItemSelectedStyle lipgloss.Style
	paletteHintStyle         lipgloss.Style
	paletteBoxStyle          lipgloss.Style
	diffHeaderStyle          lipgloss.Style
	diffAddStyle             lipgloss.Style
	diffDelStyle             lipgloss.Style
	diffCtxStyle             lipgloss.Style
	todoHeaderStyle          lipgloss.Style
	todoSectionStyle         lipgloss.Style
	todoHighStyle            lipgloss.Style
	modeBadgeStyle           lipgloss.Style
	toolCardBorderStyle      lipgloss.Style
	toolSectionStyle         lipgloss.Style
	toolSuccessStyle         lipgloss.Style
	toolCompactStyle         lipgloss.Style
	toolBadgeShellStyle      lipgloss.Style
	toolBadgeFileStyle       lipgloss.Style
	toolBadgeWebStyle        lipgloss.Style
	toolBadgeInteractStyle   lipgloss.Style
	toolBadgeDefaultStyle    lipgloss.Style
	toolJSONKeyStyle         lipgloss.Style
	toolJSONValueStyle       lipgloss.Style
)

// init seeds styles from the default theme before the first render.
func init() { rebuildStyles(currentTheme) }

// rebuildStyles recreates all styles from the given theme tokens.
func rebuildStyles(t Theme) {
	headerStyle = lipgloss.NewStyle().
		Foreground(t.Primary).
		Bold(true)

	modelTagStyle = lipgloss.NewStyle().
		Foreground(t.Muted)

	userLabelStyle = lipgloss.NewStyle().
		Foreground(t.UserLabel).
		Bold(true)

	assistantLabelStyle = lipgloss.NewStyle().
		Foreground(t.AssistantLabel).
		Bold(true)

	toolRunningStyle = lipgloss.NewStyle().
		Foreground(t.ToolRunning).
		Italic(true)

	toolDoneStyle = lipgloss.NewStyle().
		Foreground(t.ToolDone)

	toolFailedStyle = lipgloss.NewStyle().
		Foreground(t.ToolFailed).
		Bold(true)

	statusStyle = lipgloss.NewStyle().
		Foreground(t.Muted)

	warnStyle = lipgloss.NewStyle().
		Foreground(t.Warning)

	errorStyle = lipgloss.NewStyle().
		Foreground(t.Error)

	helpBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Secondary).
		Padding(1, 2)

	dividerStyle = lipgloss.NewStyle().
		Foreground(t.Subtle)

	inputPromptStyle = lipgloss.NewStyle().
		Foreground(t.InputPrompt).
		Bold(true)

	pickerHeaderStyle = lipgloss.NewStyle().
		Foreground(t.Primary).
		Bold(true).
		Padding(0, 1)

	selectedModelStyle = lipgloss.NewStyle().
		Foreground(t.SelectedItem).
		Bold(true)

	unselectedModelStyle = lipgloss.NewStyle().
		Foreground(t.UnselectedItem)

	toolExpandedHeaderStyle = lipgloss.NewStyle().
		Foreground(t.Primary).
		Bold(true)

	toolInputStyle = lipgloss.NewStyle().
		Foreground(t.Muted)

	paletteHeaderStyle = lipgloss.NewStyle().
		Foreground(t.Primary).
		Bold(true)

	paletteQueryStyle = lipgloss.NewStyle().
		Foreground(t.AssistantLabel)

	paletteItemStyle = lipgloss.NewStyle().
		Foreground(t.UnselectedItem)

	paletteItemSelectedStyle = lipgloss.NewStyle().
		Foreground(t.SelectedItem).
		Bold(true)

	paletteHintStyle = lipgloss.NewStyle().
		Foreground(t.Muted)

	paletteBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Secondary).
		Padding(1, 2)

	diffHeaderStyle = lipgloss.NewStyle().
		Foreground(t.Primary).
		Bold(true)

	diffAddStyle = lipgloss.NewStyle().
		Foreground(t.Success)

	diffDelStyle = lipgloss.NewStyle().
		Foreground(t.Error)

	diffCtxStyle = lipgloss.NewStyle().
		Foreground(t.Muted)

	todoHeaderStyle = lipgloss.NewStyle().
		Foreground(t.Primary).
		Bold(true)

	todoSectionStyle = lipgloss.NewStyle().
		Foreground(t.Muted)

	todoHighStyle = lipgloss.NewStyle().
		Foreground(t.Warning).
		Bold(true)

	modeBadgeStyle = lipgloss.NewStyle().
		Foreground(t.Primary).
		Bold(true)

	toolCardBorderStyle = lipgloss.NewStyle().
		Foreground(t.Secondary)

	toolSectionStyle = lipgloss.NewStyle().
		Foreground(t.Primary).
		Bold(true)

	toolSuccessStyle = lipgloss.NewStyle().
		Foreground(t.Success)

	toolCompactStyle = lipgloss.NewStyle().
		Foreground(t.Muted)

	toolBadgeShellStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Bold(true)

	toolBadgeFileStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("33")).
		Bold(true)

	toolBadgeWebStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("82")).
		Bold(true)

	toolBadgeInteractStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true)

	toolBadgeDefaultStyle = lipgloss.NewStyle().
		Foreground(t.Muted).
		Bold(true)

	toolJSONKeyStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("33"))

	toolJSONValueStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("82"))
}
