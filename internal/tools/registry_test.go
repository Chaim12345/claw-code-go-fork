package tools

import (
	"testing"

	"claw-code-go/internal/api"
)

func TestSchemaRegistry_AllBuiltinsRegistered(t *testing.T) {
	// Every tool that has an Execute* function should be in the
	// schema registry. If a new tool is added without registering,
	// dispatch will skip schema validation silently — this test
	// catches that drift.
	expected := []string{
		"bash", "read_file", "write_file", "file_edit",
		"glob", "grep", "web_fetch", "web_search",
		"ask_user", "todo_write", "pty_run",
	}
	for _, name := range expected {
		schema, ok := Schema(name)
		if !ok {
			t.Errorf("Schema(%q) not registered", name)
			continue
		}
		if schema.Type == "" {
			t.Errorf("Schema(%q) missing type", name)
		}
	}
}

func TestSchemaRegistry_UnknownReturnsFalse(t *testing.T) {
	if _, ok := Schema("definitely_not_a_real_tool"); ok {
		t.Error("expected unknown tool to return false")
	}
}

func TestSchemaRegistry_BashRequired(t *testing.T) {
	schema, ok := Schema("bash")
	if !ok {
		t.Fatal("bash not registered")
	}
	hasCommand := false
	for _, r := range schema.Required {
		if r == "command" {
			hasCommand = true
		}
	}
	if !hasCommand {
		t.Errorf("bash schema should require 'command', got: %v", schema.Required)
	}
}

func TestSchemaRegistry_InputSchemaValid(t *testing.T) {
	// Each registered schema must be usable with the validator.
	// This catches cases where someone updates a tool's schema
	// in a way that breaks the validator contract.
	for name, schema := range allSchemas() {
		if schema.Type != "object" {
			t.Errorf("%s: schema.Type = %q, want \"object\"", name, schema.Type)
		}
		_ = api.ValidateToolCallArguments(name, map[string]any{}, schema)
	}
}

func allSchemas() map[string]api.InputSchema {
	out := map[string]api.InputSchema{}
	for _, name := range []string{
		"bash", "read_file", "write_file", "file_edit",
		"glob", "grep", "web_fetch", "web_search",
		"ask_user", "todo_write", "pty_run",
	} {
		if s, ok := Schema(name); ok {
			out[name] = s
		}
	}
	return out
}
