package eval

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"claw-code-go/internal/api"
	"claw-code-go/internal/runtime"
	"claw-code-go/internal/tools"
)

// P4 — Recovery Rate: end-to-end tests that simulate
// malformed input flowing through the self-healing path
// (healToolInput) and verify an error tool_result is
// produced with actionable information. Also tests the
// full pipeline: XML parsing → schema validation →
// self-healing on failure.

func TestRecovery_MalformedJSON_HealsWithErrorMessage(t *testing.T) {
	raw := `{command: ls}` // unquoted key — invalid JSON
	var m map[string]any
	parseErr := json.Unmarshal([]byte(raw), &m)

	healed := runtime.HealToolInput("bash", raw, parseErr)
	if healed == nil {
		t.Fatal("healToolInput should return error tool_result for malformed JSON")
	}
	assertErrorToolResult(t, healed, "bash")
	if !strings.Contains(healed.Content[0].Text, "JSON") {
		t.Errorf("healed text should mention JSON: %q", healed.Content[0].Text)
	}
}

func TestRecovery_TruncatedJSON_HealsWithErrorMessage(t *testing.T) {
	raw := `{"command":` // truncated — missing value and closing brace
	var m map[string]any
	parseErr := json.Unmarshal([]byte(raw), &m)

	healed := runtime.HealToolInput("bash", raw, parseErr)
	if healed == nil {
		t.Fatal("healToolInput should return error tool_result for truncated JSON")
	}
	assertErrorToolResult(t, healed, "bash")
}

func TestRecovery_MissingRequiredField_HealsWithFieldName(t *testing.T) {
	raw := `{"extra_field": "value"}` // valid JSON but missing "command"
	healed := runtime.HealToolInput("bash", raw, nil)
	if healed == nil {
		t.Fatal("healToolInput should return error tool_result for missing required field")
	}
	assertErrorToolResult(t, healed, "bash")
	if !strings.Contains(healed.Content[0].Text, "command") {
		t.Errorf("healed text should mention missing field 'command': %q", healed.Content[0].Text)
	}
}

func TestRecovery_ValidArgs_NoHealingNeeded(t *testing.T) {
	raw := `{"command": "ls -la"}`
	healed := runtime.HealToolInput("bash", raw, nil)
	if healed != nil {
		t.Errorf("valid args should not trigger healing, got: %+v", healed)
	}
}

func TestRecovery_UnknownTool_NoHealing(t *testing.T) {
	raw := `{"x": 1}`
	healed := runtime.HealToolInput("nonexistent_tool", raw, nil)
	if healed != nil {
		t.Errorf("unknown tool with valid JSON should not trigger healing, got: %+v", healed)
	}
}

func TestRecovery_EmptyObject_WithRequiredFields(t *testing.T) {
	raw := `{}`
	healed := runtime.HealToolInput("bash", raw, nil)
	if healed == nil {
		t.Fatal("empty object should trigger healing for tools with required fields")
	}
	assertErrorToolResult(t, healed, "bash")
}

func TestRecovery_ParseErrorTakesPriority(t *testing.T) {
	raw := `{bad json`
	var m map[string]any
	parseErr := json.Unmarshal([]byte(raw), &m)

	healed := runtime.HealToolInput("bash", raw, parseErr)
	if healed == nil {
		t.Fatal("parse error should trigger healing")
	}
	if !strings.Contains(healed.Content[0].Text, "malformed JSON") {
		t.Errorf("parse error message should mention 'malformed JSON': %q", healed.Content[0].Text)
	}
}

func TestRecovery_FullPipeline_XmlParseToSchemaValidation(t *testing.T) {
	input := `<invoke name="bash"><parameter name="command" string="true">ls</parameter></invoke>`
	calls := api.ExtractXmlToolCalls(input)
	if len(calls) != 1 {
		t.Fatalf("ExtractXmlToolCalls returned %d calls, want 1", len(calls))
	}
	call := calls[0]
	if call.Name != "bash" {
		t.Errorf("call.Name = %q, want bash", call.Name)
	}

	schema, ok := tools.Schema(call.Name)
	if !ok {
		t.Fatalf("tools.Schema(%q) not found", call.Name)
	}

	argsBytes, _ := json.Marshal(call.Arguments)
	var args map[string]any
	_ = json.Unmarshal(argsBytes, &args)

	err := api.ValidateToolCallArguments(call.Name, args, schema)
	if err != nil {
		t.Errorf("valid pipeline args should pass schema validation: %v", err)
	}
}

func TestRecovery_FullPipeline_InvalidType_CaughtBySchema(t *testing.T) {
	input := `<invoke name="bash">{"command": 42}</invoke>`
	calls := api.ExtractXmlToolCalls(input)
	if len(calls) != 1 {
		t.Fatalf("ExtractXmlToolCalls returned %d calls, want 1", len(calls))
	}

	schema, _ := tools.Schema("bash")
	argsBytes, _ := json.Marshal(calls[0].Arguments)
	var args map[string]any
	_ = json.Unmarshal(argsBytes, &args)

	err := api.ValidateToolCallArguments("bash", args, schema)
	if err == nil {
		t.Error("integer value for 'command' (string type) should fail schema validation")
	}
	if !errors.Is(err, api.ErrSchema) {
		t.Errorf("validation error should wrap ErrSchema, got: %v", err)
	}
}

func TestRecovery_FullPipeline_MissingField_Healed(t *testing.T) {
	input := `<invoke name="bash"><parameter name="wrong_key" string="true">ls</parameter></invoke>`
	calls := api.ExtractXmlToolCalls(input)
	if len(calls) != 1 {
		t.Fatalf("ExtractXmlToolCalls returned %d calls, want 1", len(calls))
	}

	argsBytes, _ := json.Marshal(calls[0].Arguments)
	raw := string(argsBytes)

	healed := runtime.HealToolInput("bash", raw, nil)
	if healed == nil {
		t.Fatal("missing required 'command' should trigger healing")
	}
	assertErrorToolResult(t, healed, "bash")
}

func assertErrorToolResult(t *testing.T, cb *api.ContentBlock, toolName string) {
	t.Helper()
	if cb.Type != "tool_result" {
		t.Errorf("Type = %q, want tool_result", cb.Type)
	}
	if !cb.IsError {
		t.Error("IsError = false, want true")
	}
	if len(cb.Content) != 1 || cb.Content[0].Type != "text" {
		t.Errorf("Content = %+v, want single text block", cb.Content)
		return
	}
	if !strings.Contains(cb.Content[0].Text, toolName) {
		t.Errorf("error text should mention tool name %q: %q", toolName, cb.Content[0].Text)
	}
}
