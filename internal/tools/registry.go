package tools

import "claw-code-go/internal/api"

// Schema returns the InputSchema for a built-in tool by name.
// Returns false for tools not in the registry. The registry is
// the single source of truth for schema validation in
// ConversationLoop.ExecuteTool — adding a new tool without
// registering it here silently bypasses validation, so the
// registry_test.go drift check enforces the full set.
func Schema(name string) (api.InputSchema, bool) {
	switch name {
	case "bash":
		return BashTool().InputSchema, true
	case "read_file":
		return ReadFileTool().InputSchema, true
	case "write_file":
		return WriteFileTool().InputSchema, true
	case "file_edit":
		return FileEditTool().InputSchema, true
	case "glob":
		return GlobTool().InputSchema, true
	case "grep":
		return GrepTool().InputSchema, true
	case "web_fetch":
		return WebFetchTool().InputSchema, true
	case "web_search":
		return WebSearchTool().InputSchema, true
	case "ask_user":
		return AskUserQuestionTool().InputSchema, true
	case "todo_write":
		return TodoWriteTool().InputSchema, true
	case "pty_run":
		return PTYRunTool().InputSchema, true
	}
	return api.InputSchema{}, false
}
