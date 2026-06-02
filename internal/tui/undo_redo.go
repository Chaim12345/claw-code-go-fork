package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FileChange represents a single file modification
type FileChange struct {
	FilePath  string
	Before    string
	After     string
	Timestamp time.Time
	Operation string // "write", "edit", "delete"
}

// FileChangeHistory manages undo/redo for file changes
type FileChangeHistory struct {
	changes []FileChange
	cursor  int
	maxSize int
}

// NewFileChangeHistory creates a new file change history
func NewFileChangeHistory() *FileChangeHistory {
	return &FileChangeHistory{
		changes: []FileChange{},
		cursor:  -1,
		maxSize: 50, // Keep last 50 changes
	}
}

// Record adds a new file change to history
func (h *FileChangeHistory) Record(change FileChange) {
	// Truncate future if we're not at the end
	if h.cursor < len(h.changes)-1 {
		h.changes = h.changes[:h.cursor+1]
	}

	h.changes = append(h.changes, change)
	h.cursor = len(h.changes) - 1

	// Limit history size
	if len(h.changes) > h.maxSize {
		h.changes = h.changes[1:]
		h.cursor--
	}
}

// RecordFileWrite records a file write operation
func (h *FileChangeHistory) RecordFileWrite(filePath, before, after string) {
	change := FileChange{
		FilePath:  filePath,
		Before:    before,
		After:     after,
		Timestamp: time.Now(),
		Operation: "write",
	}
	h.Record(change)
}

// RecordFileEdit records a file edit operation
func (h *FileChangeHistory) RecordFileEdit(filePath, before, after string) {
	change := FileChange{
		FilePath:  filePath,
		Before:    before,
		After:     after,
		Timestamp: time.Now(),
		Operation: "edit",
	}
	h.Record(change)
}

// RecordFileDelete records a file deletion
func (h *FileChangeHistory) RecordFileDelete(filePath, content string) {
	change := FileChange{
		FilePath:  filePath,
		Before:    content,
		After:     "",
		Timestamp: time.Now(),
		Operation: "delete",
	}
	h.Record(change)
}

// CanUndo returns true if there are changes to undo
func (h *FileChangeHistory) CanUndo() bool {
	return h.cursor >= 0
}

// CanRedo returns true if there are changes to redo
func (h *FileChangeHistory) CanRedo() bool {
	return h.cursor < len(h.changes)-1
}

// Undo reverts the last change
func (h *FileChangeHistory) Undo() (*FileChange, error) {
	if !h.CanUndo() {
		return nil, fmt.Errorf("nothing to undo")
	}

	change := h.changes[h.cursor]
	h.cursor--

	// Restore previous content — applyChange handles file-not-found
	// errors naturally (we don't pre-check existence to avoid TOCTOU).
	if err := h.applyChange(change.FilePath, change.Before, change.Operation); err != nil {
		h.cursor++ // Rollback cursor on error
		return nil, fmt.Errorf("undo failed: %w", err)
	}

	return &change, nil
}

// Redo reapplies a previously undone change
func (h *FileChangeHistory) Redo() (*FileChange, error) {
	if !h.CanRedo() {
		return nil, fmt.Errorf("nothing to redo")
	}

	h.cursor++
	change := h.changes[h.cursor]

	// Reapply change
	if err := h.applyChange(change.FilePath, change.After, change.Operation); err != nil {
		h.cursor-- // Rollback cursor on error
		return nil, fmt.Errorf("redo failed: %w", err)
	}

	return &change, nil
}

// applyChange applies a change to the filesystem
func (h *FileChangeHistory) applyChange(filePath, content, operation string) error {
	switch operation {
	case "delete":
		// For undo of delete, restore the file
		if content != "" {
			return h.writeFile(filePath, content)
		}
		// For redo of delete, remove the file
		return os.Remove(filePath)

	case "write", "edit":
		if content == "" {
			// Empty content means file was deleted
			return os.Remove(filePath)
		}
		return h.writeFile(filePath, content)

	default:
		return fmt.Errorf("unknown operation: %s", operation)
	}
}

// writeFile writes content to a file, creating directories if needed
func (h *FileChangeHistory) writeFile(filePath, content string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// GetHistory returns all changes in chronological order
func (h *FileChangeHistory) GetHistory() []FileChange {
	return h.changes
}

// GetCurrentPosition returns the current cursor position
func (h *FileChangeHistory) GetCurrentPosition() int {
	return h.cursor
}

// Clear removes all history
func (h *FileChangeHistory) Clear() {
	h.changes = []FileChange{}
	h.cursor = -1
}

// GetRecentChanges returns the N most recent changes
func (h *FileChangeHistory) GetRecentChanges(n int) []FileChange {
	if n <= 0 || len(h.changes) == 0 {
		return []FileChange{}
	}

	start := len(h.changes) - n
	if start < 0 {
		start = 0
	}

	return h.changes[start:]
}

// GetChangesByFile returns all changes for a specific file
func (h *FileChangeHistory) GetChangesByFile(filePath string) []FileChange {
	var fileChanges []FileChange
	for _, change := range h.changes {
		if change.FilePath == filePath {
			fileChanges = append(fileChanges, change)
		}
	}
	return fileChanges
}

// Summary returns a human-readable summary of the history
func (h *FileChangeHistory) Summary() string {
	if len(h.changes) == 0 {
		return "No changes recorded"
	}

	undoCount := h.cursor + 1
	redoCount := len(h.changes) - undoCount

	return fmt.Sprintf("%d changes (%d can undo, %d can redo)",
		len(h.changes), undoCount, redoCount)
}
