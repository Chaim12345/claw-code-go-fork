package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// DiffApprovalState represents the state of diff approval UI
type DiffApprovalState int

const (
	DiffApprovalView DiffApprovalState = iota
	DiffApprovalEdit
	DiffApprovalConfirm
)

// DiffApproval manages interactive diff approval
type DiffApproval struct {
	FilePath   string
	OldContent string
	NewContent string
	DiffLines  []DiffLine
	State      DiffApprovalState
	Cursor     int
	Approved   bool
	Rejected   bool
	Visible    bool
	ToolCallID string
}

// DiffLine represents a single line in a diff
type DiffLine struct {
	Type    DiffLineType
	Content string
	OldNum  int
	NewNum  int
}

// DiffLineType represents the type of diff line
type DiffLineType int

const (
	DiffLineContext DiffLineType = iota
	DiffLineAdd
	DiffLineDelete
	DiffLineHeader
)

// NewDiffApproval creates a new diff approval instance
func NewDiffApproval(filePath, oldContent, newContent, toolCallID string) *DiffApproval {
	da := &DiffApproval{
		FilePath:   filePath,
		OldContent: oldContent,
		NewContent: newContent,
		State:      DiffApprovalView,
		Cursor:     0,
		Visible:    true,
		ToolCallID: toolCallID,
	}
	da.DiffLines = da.computeDiff()
	return da
}

// computeDiff generates diff lines from old and new content
func (da *DiffApproval) computeDiff() []DiffLine {
	oldLines := strings.Split(da.OldContent, "\n")
	newLines := strings.Split(da.NewContent, "\n")

	var diffLines []DiffLine

	// Simple line-by-line diff (could be enhanced with proper diff algorithm)
	maxLen := max(len(oldLines), len(newLines))

	for i := 0; i < maxLen; i++ {
		if i < len(oldLines) && i < len(newLines) {
			if oldLines[i] == newLines[i] {
				// Context line
				diffLines = append(diffLines, DiffLine{
					Type:    DiffLineContext,
					Content: oldLines[i],
					OldNum:  i + 1,
					NewNum:  i + 1,
				})
			} else {
				// Changed line - show as delete + add
				diffLines = append(diffLines, DiffLine{
					Type:    DiffLineDelete,
					Content: oldLines[i],
					OldNum:  i + 1,
					NewNum:  0,
				})
				diffLines = append(diffLines, DiffLine{
					Type:    DiffLineAdd,
					Content: newLines[i],
					OldNum:  0,
					NewNum:  i + 1,
				})
			}
		} else if i < len(oldLines) {
			// Deleted line
			diffLines = append(diffLines, DiffLine{
				Type:    DiffLineDelete,
				Content: oldLines[i],
				OldNum:  i + 1,
				NewNum:  0,
			})
		} else {
			// Added line
			diffLines = append(diffLines, DiffLine{
				Type:    DiffLineAdd,
				Content: newLines[i],
				OldNum:  0,
				NewNum:  i + 1,
			})
		}
	}

	return diffLines
}

// Approve approves the diff and applies the changes
func (da *DiffApproval) Approve() error {
	if err := os.WriteFile(da.FilePath, []byte(da.NewContent), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	da.Approved = true
	da.Visible = false
	return nil
}

// Reject rejects the diff
func (da *DiffApproval) Reject() {
	da.Rejected = true
	da.Visible = false
}

// OpenInEditor opens the file in the user's default editor
func (da *DiffApproval) OpenInEditor() error {
	// Write to temp file for editing
	tmpFile := da.FilePath + ".tmp"
	if err := os.WriteFile(tmpFile, []byte(da.NewContent), 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Get editor from environment variables
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		// Fallback to common editors
		if _, err := os.Stat("/usr/bin/nano"); err == nil {
			editor = "nano"
		} else if _, err := os.Stat("/usr/bin/vim"); err == nil {
			editor = "vim"
		} else if _, err := os.Stat("/usr/bin/vi"); err == nil {
			editor = "vi"
		} else {
			return fmt.Errorf("no editor found: set $EDITOR environment variable")
		}
	}

	// Open the editor
	cmd := exec.Command(editor, tmpFile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor exited with error: %w", err)
	}

	// Read the edited content
	editedContent, err := os.ReadFile(tmpFile)
	if err != nil {
		return fmt.Errorf("failed to read edited file: %w", err)
	}

	// Update new content
	da.NewContent = string(editedContent)

	// Clean up temp file
	os.Remove(tmpFile)

	// Recompute diff
	da.DiffLines = da.computeDiff()

	return nil
}

// View renders the diff approval UI
func (da *DiffApproval) View(width, height int) string {
	if !da.Visible {
		return ""
	}

	var b strings.Builder

	// Header
	header := diffApprovalHeaderStyle.Render(fmt.Sprintf("📝 File Edit: %s", da.FilePath))
	b.WriteString(header + "\n")
	b.WriteString(strings.Repeat("─", min(width-4, 80)) + "\n\n")

	// Diff content (show up to 20 lines)
	visibleLines := min(len(da.DiffLines), 20)
	for i := 0; i < visibleLines; i++ {
		line := da.DiffLines[i]
		b.WriteString(da.renderDiffLine(line) + "\n")
	}

	if len(da.DiffLines) > visibleLines {
		b.WriteString(statusStyle.Render(fmt.Sprintf("\n... %d more lines ...\n", len(da.DiffLines)-visibleLines)))
	}

	b.WriteString("\n")

	// Actions
	actions := diffApprovalActionsStyle.Render("[a]pprove  [r]eject  [e]dit  [d]iff tool  [Esc] cancel")
	b.WriteString(actions + "\n")

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

// renderDiffLine renders a single diff line with appropriate styling
func (da *DiffApproval) renderDiffLine(line DiffLine) string {
	var prefix string
	var style lipgloss.Style

	switch line.Type {
	case DiffLineAdd:
		prefix = "+ "
		style = diffAddStyle
	case DiffLineDelete:
		prefix = "- "
		style = diffDelStyle
	case DiffLineContext:
		prefix = "  "
		style = diffCtxStyle
	case DiffLineHeader:
		prefix = "@@"
		style = diffHeaderStyle
	}

	// Line numbers
	lineNum := ""
	if line.OldNum > 0 && line.NewNum > 0 {
		lineNum = fmt.Sprintf("%4d ", line.NewNum)
	} else if line.OldNum > 0 {
		lineNum = fmt.Sprintf("%4d ", line.OldNum)
	} else if line.NewNum > 0 {
		lineNum = fmt.Sprintf("%4d ", line.NewNum)
	} else {
		lineNum = "     "
	}

	return statusStyle.Render(lineNum) + style.Render(prefix+line.Content)
}

// GetStats returns statistics about the diff
func (da *DiffApproval) GetStats() (additions, deletions int) {
	for _, line := range da.DiffLines {
		switch line.Type {
		case DiffLineAdd:
			additions++
		case DiffLineDelete:
			deletions++
		}
	}
	return
}

// Diff approval styles
var (
	diffApprovalHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("33")).
				Bold(true)

	diffApprovalActionsStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("250"))
)
