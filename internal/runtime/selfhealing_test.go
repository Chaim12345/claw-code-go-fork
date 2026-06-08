package runtime

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// P3 — self-healing: replace {"raw": ...} fallback with informative
// error tool_result. The model emits malformed JSON or missing args;
// instead of silently passing {"raw": "..."} to the executor (which
// then fails with "'command' input is required"), we emit a
// structured error tool_result that tells the model exactly what
// to fix. On the next turn the model can re-emit correctly.

func TestSelfHealing_MalformedJSON_ProducesErrorToolResult(t *testing.T) {
	// Simulate a toolBlock with malformed JSON in inputBuffer.
	// The parser failed to produce valid JSON, so inputBuffer
	// contains something like '{"command":' (truncated).
	raw := `{"command":`
	inputMap := map[string]any{}
	err := json.Unmarshal([]byte(raw), &inputMap)

	// The old fallback: inputMap = {"raw": raw}
	// The new behavior: produce an error tool_result with a
	// model-friendly message.
	healed := HealToolInput("bash", raw, err)

	if healed == nil {
		t.Fatal("expected healed tool_result, got nil")
	}
	if healed.Type != "tool_result" {
		t.Errorf("healed.Type = %q, want tool_result", healed.Type)
	}
	if !healed.IsError {
		t.Error("healed.IsError = false, want true")
	}
	if len(healed.Content) != 1 || healed.Content[0].Type != "text" {
		t.Errorf("healed.Content = %+v, want single text block", healed.Content)
	}
	text := healed.Content[0].Text
	if text == "" {
		t.Error("healed text empty")
	}
	// The message must mention the tool name and the JSON problem
	// so the model knows what to re-emit.
	if !strings.Contains(text, "bash") {
		t.Errorf("healed text missing tool name: %q", text)
	}
	if !strings.Contains(text, "JSON") {
		t.Errorf("healed text missing 'JSON': %q", text)
	}
}

func TestSelfHealing_EmptyArgsWithRequired_ProducesErrorToolResult(t *testing.T) {
	// The model emitted <bash></bash> or {"command": ""} — the
	// args are empty but the schema requires "command".
	raw := `{}`
	inputMap := map[string]any{}
	_ = json.Unmarshal([]byte(raw), &inputMap)

	healed := HealToolInput("bash", raw, nil)

	if healed == nil {
		t.Fatal("expected healed tool_result for missing args")
	}
	if !healed.IsError {
		t.Error("want IsError=true for missing required args")
	}
	text := healed.Content[0].Text
	if !strings.Contains(text, "command") {
		t.Errorf("healed text missing required field name: %q", text)
	}
}

func TestSelfHealing_ValidJSON_NoHealing(t *testing.T) {
	// Valid JSON should pass through without wrapping.
	raw := `{"command": "ls"}`
	inputMap := map[string]any{}
	err := json.Unmarshal([]byte(raw), &inputMap)

	healed := HealToolInput("bash", raw, err)

	if healed != nil {
		t.Errorf("expected nil for valid JSON, got: %+v", healed)
	}
}

func BenchmarkHealToolInput_ParseError(b *testing.B) {
	raw := `{"command":`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HealToolInput("bash", raw, fmt.Errorf("json parse error"))
	}
}

func BenchmarkHealToolInput_ValidJSON(b *testing.B) {
	raw := `{"command":"ls -la"}`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HealToolInput("bash", raw, nil)
	}
}
