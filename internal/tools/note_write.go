package tools

import "claw-code-go/internal/api"

func NoteWriteTool() api.Tool {
	return api.Tool{
		Name:        "note_write",
		Description: "Save or retrieve session notes for multi-turn memory. Use action=read to get notes for the current session, action=write to save a note. Notes persist across compaction cycles and session resumes.",
		InputSchema: api.InputSchema{
			Type: "object",
			Properties: map[string]api.Property{
				"action": {
					Type:        "string",
					Description: `"read" to retrieve current session notes, "write" to save a note`,
				},
				"key": {
					Type:        "string",
					Description: "Short label for the note (required for action=write). Used for deduplication — writing the same key overwrites the previous note.",
				},
				"content": {
					Type:        "string",
					Description: "Note content (required for action=write)",
				},
			},
			Required: []string{"action"},
		},
	}
}
