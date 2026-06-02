package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// toolCard represents a single tool invocation that has been streamed
// into the conversation viewport. It tracks a stable ID so the TUI can
// toggle expansion per-card with the same key the user pressed in
// opencode (`Ctrl+T` to toggle the most recent, or `Ctrl+O` to expand
// all). For now the model keeps a slice of these and lets the user
// toggle the last one with a single key.
type toolCard struct {
	id         string
	name       string
	input      string
	result     string
	expanded   bool
	hasDiff    bool
	diffInline string
}

// formatToolCard renders a tool card as a pretty bordered block. Collapsed
// view shows a compact summary with icon. Expanded view shows full details
// in a bordered box with syntax highlighting for JSON inputs.
func formatToolCard(c toolCard) string {
	var b strings.Builder
	if c.expanded {
		// Expanded view: bordered box with full details
		boxWidth := 80
		topBorder := toolCardBorderStyle.Render("╭" + strings.Repeat("─", boxWidth-2) + "╮")
		bottomBorder := toolCardBorderStyle.Render("╰" + strings.Repeat("─", boxWidth-2) + "╯")

		b.WriteString(topBorder + "\n")

		// Header with icon
		icon := getToolIcon(c.name)
		header := fmt.Sprintf("│ %s %s", icon, c.name)
		b.WriteString(toolExpandedHeaderStyle.Render(header))
		b.WriteString(toolCardBorderStyle.Render(strings.Repeat(" ", boxWidth-len(header)-1) + "│\n"))

		// Separator
		b.WriteString(toolCardBorderStyle.Render("├" + strings.Repeat("─", boxWidth-2) + "┤\n"))

		// Content
		if c.hasDiff && c.diffInline != "" {
			// Show diff in box
			diffLines := strings.Split(c.diffInline, "\n")
			for _, line := range diffLines {
				if line != "" {
					b.WriteString(toolCardBorderStyle.Render("│ "))
					b.WriteString(line)
					padding := boxWidth - lipgloss.Width(line) - 3
					if padding > 0 {
						b.WriteString(strings.Repeat(" ", padding))
					}
					b.WriteString(toolCardBorderStyle.Render("│\n"))
				}
			}
		} else {
			// Show input
			b.WriteString(toolCardBorderStyle.Render("│ "))
			b.WriteString(toolSectionStyle.Render("Input:"))
			b.WriteString(toolCardBorderStyle.Render(strings.Repeat(" ", boxWidth-10) + "│\n"))

			inputLines := formatToolInput(c.input, boxWidth-4)
			for _, line := range inputLines {
				b.WriteString(toolCardBorderStyle.Render("│ "))
				b.WriteString(line)
				padding := boxWidth - lipgloss.Width(line) - 3
				if padding > 0 {
					b.WriteString(strings.Repeat(" ", padding))
				}
				b.WriteString(toolCardBorderStyle.Render("│\n"))
			}
		}

		// Result section
		if c.result != "" {
			b.WriteString(toolCardBorderStyle.Render("├" + strings.Repeat("─", boxWidth-2) + "┤\n"))
			b.WriteString(toolCardBorderStyle.Render("│ "))
			b.WriteString(toolSectionStyle.Render("Result:"))
			b.WriteString(toolCardBorderStyle.Render(strings.Repeat(" ", boxWidth-11) + "│\n"))

			preview := c.result
			if len(preview) > 4096 {
				preview = preview[:4096] + "\n…(truncated)"
			}

			resultLines := formatToolResult(preview, boxWidth-4)
			for _, line := range resultLines {
				b.WriteString(toolCardBorderStyle.Render("│ "))
				b.WriteString(line)
				padding := boxWidth - lipgloss.Width(line) - 3
				if padding > 0 {
					b.WriteString(strings.Repeat(" ", padding))
				}
				b.WriteString(toolCardBorderStyle.Render("│\n"))
			}
		}

		b.WriteString(bottomBorder)
	} else {
		// Collapsed view: compact single line with icon and badge
		icon := getToolIcon(c.name)
		badge := getToolBadge(c.name)

		summary := truncate(c.input, 50)
		status := ""
		if c.result != "" {
			status = toolSuccessStyle.Render(" ✓")
		}

		line := fmt.Sprintf("  %s %s %s %s%s", icon, badge, c.name, summary, status)
		b.WriteString(toolCompactStyle.Render(line))
	}
	return b.String()
}

// getToolIcon returns an emoji icon for the tool type
func getToolIcon(toolName string) string {
	icons := map[string]string{
		"bash":           "⚡",
		"read_file":      "📖",
		"write_file":     "✏️",
		"file_edit":      "📝",
		"glob":           "🔍",
		"grep":           "🔎",
		"web_fetch":      "🌐",
		"web_search":     "🔍",
		"ask_user":       "❓",
		"todo_write":     "✅",
		"list_files":     "📁",
		"search_replace": "🔄",
	}
	if icon, ok := icons[toolName]; ok {
		return icon
	}
	return "🔧"
}

// getToolBadge returns a colored badge for the tool category
func getToolBadge(toolName string) string {
	switch {
	case toolName == "bash":
		return toolBadgeShellStyle.Render("[SHELL]")
	case strings.Contains(toolName, "file") || strings.Contains(toolName, "write") || strings.Contains(toolName, "read"):
		return toolBadgeFileStyle.Render("[FILE]")
	case strings.Contains(toolName, "web") || strings.Contains(toolName, "search"):
		return toolBadgeWebStyle.Render("[WEB]")
	case toolName == "ask_user":
		return toolBadgeInteractStyle.Render("[USER]")
	default:
		return toolBadgeDefaultStyle.Render("[TOOL]")
	}
}

// formatToolInput formats tool input with syntax highlighting for JSON
func formatToolInput(input string, maxWidth int) []string {
	// Try to detect and pretty-print JSON
	if strings.HasPrefix(strings.TrimSpace(input), "{") {
		return formatJSON(input, maxWidth)
	}
	return wrapText(input, maxWidth)
}

// formatToolResult formats tool result with truncation and wrapping
func formatToolResult(result string, maxWidth int) []string {
	return wrapText(result, maxWidth)
}

// formatJSON attempts to format JSON with basic indentation
func formatJSON(s string, maxWidth int) []string {
	lines := strings.Split(s, "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// Add syntax coloring for JSON keys and values
		if strings.Contains(trimmed, ":") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				key := toolJSONKeyStyle.Render(parts[0])
				val := toolJSONValueStyle.Render(parts[1])
				line = key + ":" + val
			}
		}
		wrapped := wrapText(line, maxWidth)
		result = append(result, wrapped...)
	}
	return result
}

// wrapText wraps text to fit within maxWidth
func wrapText(text string, maxWidth int) []string {
	if text == "" {
		return []string{"  (empty)"}
	}

	lines := strings.Split(text, "\n")
	var result []string

	for _, line := range lines {
		if len(line) <= maxWidth {
			result = append(result, "  "+line)
			continue
		}

		// Wrap long lines
		for len(line) > maxWidth {
			result = append(result, "  "+line[:maxWidth])
			line = line[maxWidth:]
		}
		if len(line) > 0 {
			result = append(result, "  "+line)
		}
	}

	return result
}

// indentBlock prefixes every line of s with prefix.
func indentBlock(s, prefix string) string {
	if s == "" {
		return prefix + "(empty)\n"
	}
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	for i, line := range lines {
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n") + "\n"
}
