package eval

import (
	"errors"
	"testing"

	"claw-code-go/internal/api"
	"claw-code-go/internal/tools"
)

// P4 — Schema Validation Coverage: exercises
// ValidateToolCallArguments against every built-in tool
// registered in the schema registry. Validates that
// required-field enforcement and type checking work
// for each tool's actual schema, not just a synthetic one.

func TestSchemaValidation_Bash(t *testing.T) {
	schema, ok := tools.Schema("bash")
	if !ok {
		t.Fatal("bash not in schema registry")
	}
	err := api.ValidateToolCallArguments("bash", map[string]any{"command": "ls"}, schema)
	if err != nil {
		t.Errorf("valid bash args rejected: %v", err)
	}
	err = api.ValidateToolCallArguments("bash", map[string]any{}, schema)
	if err == nil {
		t.Error("missing required 'command' should fail validation")
	}
}

func TestSchemaValidation_ReadFile(t *testing.T) {
	schema, ok := tools.Schema("read_file")
	if !ok {
		t.Fatal("read_file not in schema registry")
	}
	err := api.ValidateToolCallArguments("read_file", map[string]any{"path": "/tmp/x"}, schema)
	if err != nil {
		t.Errorf("valid read_file args rejected: %v", err)
	}
	err = api.ValidateToolCallArguments("read_file", map[string]any{}, schema)
	if err == nil {
		t.Error("missing required 'path' should fail validation")
	}
}

func TestSchemaValidation_WriteFile(t *testing.T) {
	schema, ok := tools.Schema("write_file")
	if !ok {
		t.Fatal("write_file not in schema registry")
	}
	err := api.ValidateToolCallArguments("write_file", map[string]any{
		"path":    "/tmp/out.txt",
		"content": "hello",
	}, schema)
	if err != nil {
		t.Errorf("valid write_file args rejected: %v", err)
	}
	err = api.ValidateToolCallArguments("write_file", map[string]any{"path": "/tmp/out.txt"}, schema)
	if err == nil {
		t.Error("missing required 'content' should fail validation")
	}
}

func TestSchemaValidation_FileEdit(t *testing.T) {
	schema, ok := tools.Schema("file_edit")
	if !ok {
		t.Fatal("file_edit not in schema registry")
	}
	err := api.ValidateToolCallArguments("file_edit", map[string]any{
		"file_path":   "/tmp/x.go",
		"old_string":  "foo",
		"new_string":  "bar",
	}, schema)
	if err != nil {
		t.Errorf("valid file_edit args rejected: %v", err)
	}
	err = api.ValidateToolCallArguments("file_edit", map[string]any{"file_path": "/tmp/x.go"}, schema)
	if err == nil {
		t.Error("missing required 'old_string' and 'new_string' should fail validation")
	}
}

func TestSchemaValidation_Glob(t *testing.T) {
	schema, ok := tools.Schema("glob")
	if !ok {
		t.Fatal("glob not in schema registry")
	}
	err := api.ValidateToolCallArguments("glob", map[string]any{"pattern": "**/*.go"}, schema)
	if err != nil {
		t.Errorf("valid glob args rejected: %v", err)
	}
	err = api.ValidateToolCallArguments("glob", map[string]any{}, schema)
	if err == nil {
		t.Error("missing required 'pattern' should fail validation")
	}
}

func TestSchemaValidation_Grep(t *testing.T) {
	schema, ok := tools.Schema("grep")
	if !ok {
		t.Fatal("grep not in schema registry")
	}
	err := api.ValidateToolCallArguments("grep", map[string]any{"pattern": "TODO", "path": "/tmp"}, schema)
	if err != nil {
		t.Errorf("valid grep args rejected: %v", err)
	}
	err = api.ValidateToolCallArguments("grep", map[string]any{"pattern": "TODO"}, schema)
	if err == nil {
		t.Error("missing required 'path' should fail validation")
	}
}

func TestSchemaValidation_WebFetch(t *testing.T) {
	schema, ok := tools.Schema("web_fetch")
	if !ok {
		t.Fatal("web_fetch not in schema registry")
	}
	err := api.ValidateToolCallArguments("web_fetch", map[string]any{"url": "https://example.com"}, schema)
	if err != nil {
		t.Errorf("valid web_fetch args rejected: %v", err)
	}
	err = api.ValidateToolCallArguments("web_fetch", map[string]any{}, schema)
	if err == nil {
		t.Error("missing required 'url' should fail validation")
	}
}

func TestSchemaValidation_WebSearch(t *testing.T) {
	schema, ok := tools.Schema("web_search")
	if !ok {
		t.Fatal("web_search not in schema registry")
	}
	err := api.ValidateToolCallArguments("web_search", map[string]any{"query": "golang testing"}, schema)
	if err != nil {
		t.Errorf("valid web_search args rejected: %v", err)
	}
	err = api.ValidateToolCallArguments("web_search", map[string]any{}, schema)
	if err == nil {
		t.Error("missing required 'query' should fail validation")
	}
}

func TestSchemaValidation_AskUser(t *testing.T) {
	schema, ok := tools.Schema("ask_user")
	if !ok {
		t.Fatal("ask_user not in schema registry")
	}
	err := api.ValidateToolCallArguments("ask_user", map[string]any{"question": "continue?"}, schema)
	if err != nil {
		t.Errorf("valid ask_user args rejected: %v", err)
	}
	err = api.ValidateToolCallArguments("ask_user", map[string]any{}, schema)
	if err == nil {
		t.Error("missing required 'question' should fail validation")
	}
}

func TestSchemaValidation_TodoWrite(t *testing.T) {
	schema, ok := tools.Schema("todo_write")
	if !ok {
		t.Fatal("todo_write not in schema registry")
	}
	err := api.ValidateToolCallArguments("todo_write", map[string]any{"action": "add"}, schema)
	if err != nil {
		t.Errorf("valid todo_write args rejected: %v", err)
	}
	err = api.ValidateToolCallArguments("todo_write", map[string]any{}, schema)
	if err == nil {
		t.Error("missing required 'action' should fail validation")
	}
}

func TestSchemaValidation_PTYRun(t *testing.T) {
	schema, ok := tools.Schema("pty_run")
	if !ok {
		t.Fatal("pty_run not in schema registry")
	}
	err := api.ValidateToolCallArguments("pty_run", map[string]any{"command": "vim"}, schema)
	if err != nil {
		t.Errorf("valid pty_run args rejected: %v", err)
	}
	err = api.ValidateToolCallArguments("pty_run", map[string]any{}, schema)
	if err == nil {
		t.Error("missing required 'command' should fail validation")
	}
}

func TestSchemaValidation_TypeMismatch(t *testing.T) {
	schema, ok := tools.Schema("bash")
	if !ok {
		t.Fatal("bash not in schema registry")
	}
	err := api.ValidateToolCallArguments("bash", map[string]any{"command": 42}, schema)
	if err == nil {
		t.Error("integer value for string-typed 'command' should fail")
	}
	if err != nil && !isSchemaError(err) {
		t.Errorf("expected ErrSchema-wrapped error, got: %v", err)
	}
}

func TestSchemaValidation_UnknownToolNoPanic(t *testing.T) {
	err := api.ValidateToolCallArguments("nonexistent_tool", map[string]any{"x": 1}, api.InputSchema{})
	if err != nil {
		t.Errorf("unknown tool with empty schema should not error, got: %v", err)
	}
}

func TestSchemaValidation_AllToolsRegistered(t *testing.T) {
	builtinTools := []string{
		"bash", "read_file", "write_file", "file_edit",
		"glob", "grep", "web_fetch", "web_search",
		"ask_user", "todo_write", "pty_run",
	}
	for _, name := range builtinTools {
		_, ok := tools.Schema(name)
		if !ok {
			t.Errorf("tools.Schema(%q) returned false — tool not registered", name)
		}
	}
}

func isSchemaError(err error) bool {
	return errors.Is(err, api.ErrSchema)
}
