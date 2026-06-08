package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"claw-code-go/internal/api"
)

// ---- looksLikeToolCallStart tests ----

func TestLooksLikeToolCallStart(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"<tool_calls>", true},
		{"<|DSML|tool_calls>", true},
		{`<invoke name="test">`, true},
		{"<invoke>", true},
		{`{"tool_calls":[]}`, true},
		{"This is normal text", false},
		{"Some <text> that doesn't match", false},
		// HTML-encoded variants
		{"&lt;tool_calls&gt;", true},
		{`&lt;invoke name=&quot;bash&quot;&gt;`, true},
		// Case insensitivity
		{"<TOOL_CALLS>", true},
		{"<INVOKE>", true},
		// DSML variant with pipe
		{"<|dsml|tool_calls>", true},
		// Partial match (should not match)
		{"<tool", false},
		{"<invo", false},
		{"plain text", false},
		// Empty string
		{"", false},
	}

	for _, tt := range tests {
		if got := looksLikeToolCallStart(tt.input); got != tt.want {
			t.Errorf("looksLikeToolCallStart(%q) = %v; want %v", tt.input, got, tt.want)
		}
	}
}

// ---- trimFinishedSentinel tests ----

func TestTrimFinishedSentinel(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		// Space before FINISHED: punctuation is not adjacent to sentinel
		{"Hello FINISHED", "Hello"},
		{"Hello. FINISHED", "Hello."},
		{"Hello! FINISHED", "Hello!"},
		{"Hello? FINISHED", "Hello?"},
		{"Hello FINISHED ", "Hello"},
		{"Hello FINISHED\n", "Hello"},
		// No sentinel
		{"Hello", "Hello"},
		// Just sentinel
		{"FINISHED", ""},
		// Case sensitive: lowercase "finished" is not stripped
		{"Hello finished", "Hello finished"},
		// Sentinel after other text
		{"Not finished FINISHED", "Not finished"},
		// Punctuation directly before sentinel (no space)
		{"Hello.FINISHED", "Hello"},
		{"Hello!FINISHED", "Hello"},
		{"Hello?FINISHED", "Hello"},
		// Empty and whitespace-only
		{"", ""},
		// Single space without FINISHED — not trimmed because TrimRight result "" doesn't end with sentinel
		{" FINISHED ", ""},
	}

	for _, tt := range tests {
		got := trimFinishedSentinel(tt.input)
		if got != tt.want {
			t.Errorf("trimFinishedSentinel(%q) = %q; want %q", tt.input, got, tt.want)
		}
	}
}

// ---- extractTextContent tests ----

func TestExtractTextContent(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"plain string", `"hello world"`, "hello world"},
		{"empty", `""`, ""},
		{"null raw", `null`, ""},
		{"empty raw", ``, ""},
		{"array of content parts", `[{"type":"text","text":"hello"},{"type":"text","text":"world"}]`, "hello\nworld"},
		{"mixed types", `[{"type":"text","text":"first"},{"type":"image","url":"..."}]`, "first"},
		{"empty array", `[]`, ""},
		{"escape sequences", `"hello\nworld"`, "hello\nworld"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractTextContent(json.RawMessage(tt.raw))
			if got != tt.want {
				t.Errorf("extractTextContent(%s) = %q; want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestExtractTextContent_NilRaw(t *testing.T) {
	got := extractTextContent(nil)
	if got != "" {
		t.Errorf("extractTextContent(nil) = %q; want empty", got)
	}
}

func TestExtractTextContent_InvalidJSON(t *testing.T) {
	got := extractTextContent(json.RawMessage(`{broken json`))
	if got == "" {
		t.Error("expected non-empty fallback for invalid JSON")
	}
}

// ---- convertMessages tests ----

func TestConvertMessages_SystemAndUser(t *testing.T) {
	msgs := []chatMessage{
		{Role: "system", Content: json.RawMessage(`"You are helpful"`)},
		{Role: "user", Content: json.RawMessage(`"Hello"`)},
	}

	system, messages := convertMessages(msgs)
	if system != "You are helpful" {
		t.Errorf("system = %q; want %q", system, "You are helpful")
	}
	if len(messages) != 1 {
		t.Fatalf("len(messages) = %d; want 1", len(messages))
	}
	if messages[0].Role != "user" {
		t.Errorf("messages[0].Role = %q; want user", messages[0].Role)
	}
}

func TestConvertMessages_DeveloperRole(t *testing.T) {
	msgs := []chatMessage{
		{Role: "developer", Content: json.RawMessage(`"You are a developer"`)},
	}

	system, _ := convertMessages(msgs)
	if system != "You are a developer" {
		t.Errorf("system = %q; want %q", system, "You are a developer")
	}
}

func TestConvertMessages_MultipleSystemMessages(t *testing.T) {
	msgs := []chatMessage{
		{Role: "system", Content: json.RawMessage(`"First instruction"`)},
		{Role: "system", Content: json.RawMessage(`"Second instruction"`)},
	}

	system, _ := convertMessages(msgs)
	if !strings.Contains(system, "First instruction") || !strings.Contains(system, "Second instruction") {
		t.Errorf("system = %q; want both instructions concatenated", system)
	}
}

func TestConvertMessages_UserWithName(t *testing.T) {
	msgs := []chatMessage{
		{Role: "user", Name: "Alice", Content: json.RawMessage(`"Hi there"`)},
	}

	_, messages := convertMessages(msgs)
	if len(messages) != 1 {
		t.Fatalf("len(messages) = %d; want 1", len(messages))
	}
	text := messages[0].Content[0].Text
	if text != "[Alice] Hi there" {
		t.Errorf("text = %q; want %q", text, "[Alice] Hi there")
	}
}

func TestConvertMessages_ToolCallID(t *testing.T) {
	msgs := []chatMessage{
		{Role: "user", ToolCallID: "call_123", Content: json.RawMessage(`"result data"`)},
	}

	_, messages := convertMessages(msgs)
	if len(messages) != 1 {
		t.Fatalf("len(messages) = %d; want 1", len(messages))
	}
	if messages[0].Role != "user" {
		t.Errorf("Role = %q; want user", messages[0].Role)
	}
	if len(messages[0].Content) != 1 {
		t.Fatalf("len(Content) = %d; want 1", len(messages[0].Content))
	}
	if messages[0].Content[0].Type != "tool_result" {
		t.Errorf("Content[0].Type = %q; want tool_result", messages[0].Content[0].Type)
	}
	if messages[0].Content[0].ToolUseID != "call_123" {
		t.Errorf("ToolUseID = %q; want call_123", messages[0].Content[0].ToolUseID)
	}
}

func TestConvertMessages_AssistantWithToolCalls(t *testing.T) {
	msgs := []chatMessage{
		{
			Role: "assistant",
			ToolCalls: []openAIToolCall{
				{
					ID:   "call_abc",
					Type: "function",
					Function: openAIFunction{
						Name:      "bash",
						Arguments: `{"command":"ls"}`,
					},
				},
			},
		},
	}

	_, messages := convertMessages(msgs)
	if len(messages) != 1 {
		t.Fatalf("len(messages) = %d; want 1", len(messages))
	}
	if messages[0].Role != "assistant" {
		t.Errorf("Role = %q; want assistant", messages[0].Role)
	}
	if len(messages[0].Content) != 1 {
		t.Fatalf("len(Content) = %d; want 1", len(messages[0].Content))
	}
	content := messages[0].Content[0]
	if content.Type != "tool_use" {
		t.Errorf("Type = %q; want tool_use", content.Type)
	}
	if content.Name != "bash" {
		t.Errorf("Name = %q; want bash", content.Name)
	}
	if content.ID != "call_abc" {
		t.Errorf("ID = %q; want call_abc", content.ID)
	}
	cmd, ok := content.Input["command"].(string)
	if !ok || cmd != "ls" {
		t.Errorf("Input[command] = %v; want ls", content.Input["command"])
	}
}

func TestConvertMessages_ToolRole(t *testing.T) {
	msgs := []chatMessage{
		{Role: "tool", ToolCallID: "call_456", Content: json.RawMessage(`"output from tool"`)},
	}

	_, messages := convertMessages(msgs)
	if len(messages) != 1 {
		t.Fatalf("len(messages) = %d; want 1", len(messages))
	}
	if messages[0].Role != "user" {
		t.Errorf("Role = %q; want user", messages[0].Role)
	}
	content := messages[0].Content[0]
	if content.Type != "tool_result" {
		t.Errorf("Type = %q; want tool_result", content.Type)
	}
	if content.ToolUseID != "call_456" {
		t.Errorf("ToolUseID = %q; want call_456", content.ToolUseID)
	}
}

func TestConvertMessages_FunctionRole(t *testing.T) {
	msgs := []chatMessage{
		{Role: "function", Name: "get_weather", Content: json.RawMessage(`"sunny"`)},
	}

	_, messages := convertMessages(msgs)
	if len(messages) != 1 {
		t.Fatalf("len(messages) = %d; want 1", len(messages))
	}
	if messages[0].Role != "user" {
		t.Errorf("Role = %q; want user", messages[0].Role)
	}
	content := messages[0].Content[0]
	if content.Type != "tool_result" {
		t.Errorf("Type = %q; want tool_result", content.Type)
	}
	if content.ToolUseID != "get_weather" {
		t.Errorf("ToolUseID = %q; want get_weather", content.ToolUseID)
	}
}

func TestConvertMessages_EmptyInput(t *testing.T) {
	system, messages := convertMessages(nil)
	if system != "" {
		t.Errorf("system = %q; want empty", system)
	}
	// make([]api.Message, 0, 0) returns empty non-nil slice
	if len(messages) != 0 {
		t.Errorf("len(messages) = %d; want 0", len(messages))
	}
}

func TestConvertMessages_AssistantWithContent(t *testing.T) {
	msgs := []chatMessage{
		{Role: "assistant", Content: json.RawMessage(`"I can help with that"`)},
	}

	_, messages := convertMessages(msgs)
	if len(messages) != 1 {
		t.Fatalf("len(messages) = %d; want 1", len(messages))
	}
	if messages[0].Content[0].Text != "I can help with that" {
		t.Errorf("Text = %q; want %q", messages[0].Content[0].Text, "I can help with that")
	}
}

func TestConvertMessages_RoundtripToolSequence(t *testing.T) {
	msgs := []chatMessage{
		{Role: "user", Content: json.RawMessage(`"What files are in /tmp?"`)},
		{
			Role: "assistant",
			ToolCalls: []openAIToolCall{
				{
					ID:   "call_1",
					Type: "function",
					Function: openAIFunction{
						Name:      "bash",
						Arguments: `{"command":"ls /tmp"}`,
					},
				},
			},
		},
		{Role: "tool", ToolCallID: "call_1", Content: json.RawMessage(`"file1.txt\nfile2.txt"`)},
	}

	system, messages := convertMessages(msgs)
	if system != "" {
		t.Errorf("system = %q; want empty", system)
	}
	if len(messages) != 3 {
		t.Fatalf("len(messages) = %d; want 3", len(messages))
	}
	if messages[0].Role != "user" || messages[0].Content[0].Text != "What files are in /tmp?" {
		t.Errorf("messages[0] incorrect: role=%q text=%q", messages[0].Role, messages[0].Content[0].Text)
	}
	if messages[1].Role != "assistant" || messages[1].Content[0].Type != "tool_use" {
		t.Errorf("messages[1] incorrect: role=%q type=%q", messages[1].Role, messages[1].Content[0].Type)
	}
	if messages[2].Role != "user" || messages[2].Content[0].Type != "tool_result" {
		t.Errorf("messages[2] incorrect: role=%q type=%q", messages[2].Role, messages[2].Content[0].Type)
	}
}

func TestConvertMessages_NilContent(t *testing.T) {
	msgs := []chatMessage{
		{Role: "assistant", Content: nil},
	}
	_, messages := convertMessages(msgs)
	if len(messages) != 0 {
		t.Errorf("len(messages) = %d; want 0 for assistant with nil content and no tool calls", len(messages))
	}
}

func TestConvertMessages_ArrayContentParts(t *testing.T) {
	msgs := []chatMessage{
		{Role: "user", Content: json.RawMessage(`[{"type":"text","text":"part one"},{"type":"text","text":"part two"}]`)},
	}
	_, messages := convertMessages(msgs)
	if len(messages) != 1 {
		t.Fatalf("len(messages) = %d; want 1", len(messages))
	}
	want := "part one\npart two"
	if messages[0].Content[0].Text != want {
		t.Errorf("Text = %q; want %q", messages[0].Content[0].Text, want)
	}
}

func TestConvertMessages_AssistantRefusal(t *testing.T) {
	msgs := []chatMessage{
		{Role: "assistant", Refusal: "I cannot do that", Content: json.RawMessage(`""`)},
	}
	_, messages := convertMessages(msgs)
	if len(messages) != 1 {
		t.Fatalf("len(messages) = %d; want 1", len(messages))
	}
	if messages[0].Content[0].Text != "I cannot do that" {
		t.Errorf("Text = %q; want %q", messages[0].Content[0].Text, "I cannot do that")
	}
}

// ---- extractToolCallsFromText tests ----

func TestExtractToolCallsFromText_PlainXML(t *testing.T) {
	text := `<tool_calls>
<invoke name="bash">
<parameter name="command">ls -la</parameter>
</invoke>
</tool_calls>`

	calls := extractToolCallsFromText(text)
	if len(calls) == 0 {
		t.Fatal("expected tool calls, got none")
	}
	if calls[0].Function.Name != "bash" {
		t.Errorf("Name = %q; want bash", calls[0].Function.Name)
	}
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(calls[0].Function.Arguments), &args); err != nil {
		t.Fatalf("Arguments not valid JSON: %v", err)
	}
	if cmd, ok := args["command"].(string); !ok || cmd != "ls -la" {
		t.Errorf("args[command] = %v; want ls -la", args["command"])
	}
	if calls[0].Type != "function" {
		t.Errorf("Type = %q; want function", calls[0].Type)
	}
	if !strings.HasPrefix(calls[0].ID, "call_") {
		t.Errorf("ID = %q; want call_ prefix", calls[0].ID)
	}
}

func TestExtractToolCallsFromText_XMLMultipleInvokes(t *testing.T) {
	text := `<tool_calls>
<invoke name="bash">
<parameter name="command">ls</parameter>
</invoke>
<invoke name="file_edit">
<parameter name="path">/tmp/test.txt</parameter>
<parameter name="content">hello</parameter>
</invoke>
</tool_calls>`

	calls := extractToolCallsFromText(text)
	if len(calls) < 2 {
		t.Fatalf("expected at least 2 tool calls, got %d", len(calls))
	}
	if calls[0].Function.Name != "bash" {
		t.Errorf("calls[0].Name = %q; want bash", calls[0].Function.Name)
	}
	if calls[1].Function.Name != "file_edit" {
		t.Errorf("calls[1].Name = %q; want file_edit", calls[1].Function.Name)
	}
}

func TestExtractToolCallsFromText_NoToolCalls(t *testing.T) {
	calls := extractToolCallsFromText("Just regular text with no tool calls.")
	if calls != nil {
		t.Errorf("expected nil for plain text, got %v", calls)
	}
}

func TestExtractToolCallsFromText_JSONFormat(t *testing.T) {
	text := `{"tool_calls":[{"name":"bash","arguments":{"command":"ls -la"}}]}`

	calls := extractToolCallsFromText(text)
	if len(calls) == 0 {
		t.Fatal("expected tool calls, got none")
	}
	if calls[0].Function.Name != "bash" {
		t.Errorf("Name = %q; want bash", calls[0].Function.Name)
	}
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(calls[0].Function.Arguments), &args); err != nil {
		t.Fatalf("Arguments not valid JSON: %v", err)
	}
	if cmd, ok := args["command"].(string); !ok || cmd != "ls -la" {
		t.Errorf("args[command] = %v; want ls -la", args["command"])
	}
}

func TestExtractToolCallsFromText_XMLOverJSON(t *testing.T) {
	text := `<tool_calls>
<invoke name="grep">
<parameter name="pattern">TODO</parameter>
</invoke>
</tool_calls>`

	calls := extractToolCallsFromText(text)
	if len(calls) == 0 {
		t.Fatal("expected tool calls, got none")
	}
	if calls[0].Function.Name != "grep" {
		t.Errorf("Name = %q; want grep", calls[0].Function.Name)
	}
}

// ---- extractJSONToolCalls tests ----

func TestExtractJSONToolCalls_Single(t *testing.T) {
	text := `{"tool_calls":[{"name":"bash","arguments":{"command":"ls"}}]}`

	calls := extractJSONToolCalls(text)
	if len(calls) != 1 {
		t.Fatalf("len(calls) = %d; want 1", len(calls))
	}
	if calls[0].Name != "bash" {
		t.Errorf("Name = %q; want bash", calls[0].Name)
	}
	if cmd, ok := calls[0].Arguments["command"].(string); !ok || cmd != "ls" {
		t.Errorf("Arguments[command] = %v; want ls", calls[0].Arguments["command"])
	}
}

func TestExtractJSONToolCalls_Multiple(t *testing.T) {
	text := `{"tool_calls":[{"name":"bash","arguments":{"command":"ls"}},{"name":"file_edit","arguments":{"path":"/tmp/x","content":"hi"}}]}`

	calls := extractJSONToolCalls(text)
	if len(calls) != 2 {
		t.Fatalf("len(calls) = %d; want 2", len(calls))
	}
	if calls[0].Name != "bash" {
		t.Errorf("calls[0].Name = %q; want bash", calls[0].Name)
	}
	if calls[1].Name != "file_edit" {
		t.Errorf("calls[1].Name = %q; want file_edit", calls[1].Name)
	}
}

func TestExtractJSONToolCalls_NoMatch(t *testing.T) {
	calls := extractJSONToolCalls("no tool calls here")
	if calls != nil {
		t.Errorf("expected nil, got %v", calls)
	}
}

func TestExtractJSONToolCalls_EmptyToolCalls(t *testing.T) {
	text := `{"tool_calls":[]}`

	calls := extractJSONToolCalls(text)
	if len(calls) != 0 {
		t.Errorf("expected 0 tool calls, got %d", len(calls))
	}
}

func TestExtractJSONToolCalls_WithPrefix(t *testing.T) {
	text := `Here is the result: {"tool_calls":[{"name":"bash","arguments":{"command":"pwd"}}]}`

	calls := extractJSONToolCalls(text)
	if len(calls) != 1 {
		t.Fatalf("len(calls) = %d; want 1", len(calls))
	}
	if calls[0].Name != "bash" {
		t.Errorf("Name = %q; want bash", calls[0].Name)
	}
}

func TestExtractJSONToolCalls_InvalidJSON(t *testing.T) {
	text := `{"tool_calls":[{"name":"bash"`
	calls := extractJSONToolCalls(text)
	if calls != nil {
		t.Errorf("expected nil for invalid JSON, got %v", calls)
	}
}

func TestExtractJSONToolCalls_NestedBraces(t *testing.T) {
	text := `{"tool_calls":[{"name":"bash","arguments":{"command":"echo '{\"key\": \"val\"}'"}}]}`

	calls := extractJSONToolCalls(text)
	if len(calls) != 1 {
		t.Fatalf("len(calls) = %d; want 1", len(calls))
	}
	if calls[0].Name != "bash" {
		t.Errorf("Name = %q; want bash", calls[0].Name)
	}
}

// ---- toPropertyMap tests ----

func TestToPropertyMap(t *testing.T) {
	raw := map[string]interface{}{
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "The command to run",
			},
		},
	}

	props := toPropertyMap(raw)
	if len(props) != 1 {
		t.Fatalf("len(props) = %d; want 1", len(props))
	}
	p, ok := props["command"]
	if !ok {
		t.Fatal("expected 'command' property")
	}
	if p.Type != "string" {
		t.Errorf("Type = %q; want string", p.Type)
	}
	if p.Description != "The command to run" {
		t.Errorf("Description = %q; want 'The command to run'", p.Description)
	}
}

func TestToPropertyMap_NoProperties(t *testing.T) {
	raw := map[string]interface{}{}
	props := toPropertyMap(raw)
	if props != nil {
		t.Errorf("expected nil, got %v", props)
	}
}

func TestToPropertyMap_InvalidPropertiesType(t *testing.T) {
	raw := map[string]interface{}{
		"properties": "not an object",
	}
	props := toPropertyMap(raw)
	if props != nil {
		t.Errorf("expected nil for invalid properties type, got %v", props)
	}
}

func TestToPropertyMap_PropertyWithInvalidSubType(t *testing.T) {
	raw := map[string]interface{}{
		"properties": map[string]interface{}{
			"bad": "not an object",
		},
	}
	props := toPropertyMap(raw)
	if len(props) != 0 {
		t.Errorf("expected 0 props for invalid sub-type, got %d", len(props))
	}
}

// ---- requiredFields tests ----

func TestRequiredFields(t *testing.T) {
	raw := map[string]interface{}{
		"required": []interface{}{"command", "path"},
	}
	fields := requiredFields(raw)
	if len(fields) != 2 {
		t.Fatalf("len(fields) = %d; want 2", len(fields))
	}
	if fields[0] != "command" || fields[1] != "path" {
		t.Errorf("fields = %v; want [command path]", fields)
	}
}

func TestRequiredFields_NoRequired(t *testing.T) {
	raw := map[string]interface{}{}
	fields := requiredFields(raw)
	if fields != nil {
		t.Errorf("expected nil, got %v", fields)
	}
}

func TestRequiredFields_InvalidType(t *testing.T) {
	raw := map[string]interface{}{
		"required": "not an array",
	}
	fields := requiredFields(raw)
	if fields != nil {
		t.Errorf("expected nil for invalid type, got %v", fields)
	}
}

func TestRequiredFields_NonStringElements(t *testing.T) {
	raw := map[string]interface{}{
		"required": []interface{}{"command", 42, true},
	}
	fields := requiredFields(raw)
	if len(fields) != 1 {
		t.Fatalf("len(fields) = %d; want 1 (only string elements)", len(fields))
	}
	if fields[0] != "command" {
		t.Errorf("fields[0] = %q; want command", fields[0])
	}
}

// ---- streamingChunk tests ----

func TestStreamingChunk(t *testing.T) {
	content := map[string]interface{}{
		"content": "hello",
	}
	toolCall := map[string]interface{}{
		"id":       "call_1",
		"function": map[string]interface{}{"name": "bash"},
	}

	chunk := streamingChunk("chatcmpl-1", "deepseek-chat", 12345, content, toolCall)

	if chunk["id"] != "chatcmpl-1" {
		t.Errorf("id = %v; want chatcmpl-1", chunk["id"])
	}
	if chunk["object"] != "chat.completion.chunk" {
		t.Errorf("object = %v; want chat.completion.chunk", chunk["object"])
	}
	if chunk["model"] != "deepseek-chat" {
		t.Errorf("model = %v; want deepseek-chat", chunk["model"])
	}

	choices, ok := chunk["choices"].([]interface{})
	if !ok || len(choices) != 1 {
		t.Fatalf("choices = %v; want single-element array", chunk["choices"])
	}
	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		t.Fatal("choice is not a map")
	}
	delta, ok := choice["delta"].(map[string]interface{})
	if !ok {
		t.Fatal("delta is not a map")
	}
	// Content keys are merged into delta
	if delta["content"] != "hello" {
		t.Errorf("delta[content] = %v; want hello", delta["content"])
	}
	// Tool calls are nested under delta.tool_calls
	tcArr, ok := delta["tool_calls"].([]interface{})
	if !ok || len(tcArr) != 1 {
		t.Fatalf("delta[tool_calls] = %v; want single-element array", delta["tool_calls"])
	}
	tcMap, ok := tcArr[0].(map[string]interface{})
	if !ok {
		t.Fatalf("delta[tool_calls][0] is not a map, got %T", tcArr[0])
	}
	if tcMap["id"] != "call_1" {
		t.Errorf("delta[tool_calls][0].id = %v; want call_1", tcMap["id"])
	}
	// choice should not have a separate tool_calls key
	if _, exists := choice["tool_calls"]; exists {
		t.Error("choice should not have a tool_calls key; it should be nested in delta")
	}
}

func TestStreamingChunk_ContentOnly(t *testing.T) {
	content := map[string]interface{}{
		"content": "just text",
	}

	chunk := streamingChunk("chatcmpl-3", "deepseek-chat", 99999, content, nil)

	choices := chunk["choices"].([]interface{})
	choice := choices[0].(map[string]interface{})
	delta := choice["delta"].(map[string]interface{})

	if delta["content"] != "just text" {
		t.Errorf("delta[content] = %v; want just text", delta["content"])
	}
	if _, exists := delta["tool_calls"]; exists {
		t.Error("delta should not have tool_calls when toolCall is nil")
	}
}

func TestStreamingChunkWithUsage(t *testing.T) {
	chunk := streamingChunkWithUsage("chatcmpl-2", "deepseek-chat", 12345, "stop", 100, 50)

	if chunk["id"] != "chatcmpl-2" {
		t.Errorf("id = %v; want chatcmpl-2", chunk["id"])
	}

	choices, ok := chunk["choices"].([]interface{})
	if !ok || len(choices) != 1 {
		t.Fatalf("choices = %v; want single-element array", chunk["choices"])
	}
	choice := choices[0].(map[string]interface{})
	if choice["finish_reason"] != "stop" {
		t.Errorf("finish_reason = %v; want stop", choice["finish_reason"])
	}

	usage, ok := chunk["usage"].(map[string]interface{})
	if !ok {
		t.Fatal("usage is not a map")
	}
	if usage["prompt_tokens"] != 100 {
		t.Errorf("prompt_tokens = %v; want 100", usage["prompt_tokens"])
	}
	if usage["completion_tokens"] != 50 {
		t.Errorf("completion_tokens = %v; want 50", usage["completion_tokens"])
	}
	if usage["total_tokens"] != 150 {
		t.Errorf("total_tokens = %v; want 150", usage["total_tokens"])
	}
}

// ---- writeSSE / writeOpenAIError tests ----

func TestWriteSSE_Nil(t *testing.T) {
	rec := httptest.NewRecorder()
	writeSSE(rec, nil)
	body := rec.Body.String()
	if body != "data: [DONE]\n\n" {
		t.Errorf("body = %q; want %q", body, "data: [DONE]\n\n")
	}
}

func TestWriteSSE_WithPayload(t *testing.T) {
	rec := httptest.NewRecorder()
	writeSSE(rec, map[string]string{"role": "assistant"})
	body := rec.Body.String()
	if !strings.HasPrefix(body, "data: ") {
		t.Errorf("body should start with 'data: ', got %q", body)
	}
	if !strings.Contains(body, `"role":"assistant"`) && !strings.Contains(body, `"role": "assistant"`) {
		t.Errorf("body should contain role:assistant, got %q", body)
	}
	if !strings.HasSuffix(body, "\n\n") {
		t.Errorf("body should end with double newline, got %q", body)
	}
}

func TestWriteOpenAIError(t *testing.T) {
	rec := httptest.NewRecorder()
	writeOpenAIError(rec, "/v1/chat/completions", "invalid_request_error", "model not found", "", http.StatusNotFound)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusNotFound)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q; want application/json", ct)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	errObj, ok := resp["error"].(map[string]interface{})
	if !ok {
		t.Fatal("response has no 'error' object")
	}
	if errObj["message"] != "model not found" {
		t.Errorf("error.message = %v; want model not found", errObj["message"])
	}
	if errObj["type"] != "invalid_request_error" {
		t.Errorf("error.type = %v; want invalid_request_error", errObj["type"])
	}
}

// ---- knownModels tests ----

func TestKnownModels(t *testing.T) {
	if len(knownModels) == 0 {
		t.Error("knownModels is empty after init")
	}
	found := false
	for _, m := range knownModels {
		if strings.Contains(m.ID, "deepseek") {
			found = true
			break
		}
	}
	if !found {
		t.Error("no deepseek model found in knownModels")
	}
}

// ---- api.ExtractXmlToolCalls round-trip ----

func TestExtractXmlToolCalls_Integration(t *testing.T) {
	text := `<tool_calls>
<invoke name="bash">
<parameter name="command">echo hello</parameter>
</invoke>
</tool_calls>`

	calls := api.ExtractXmlToolCalls(text)
	if len(calls) != 1 {
		t.Fatalf("len(calls) = %d; want 1", len(calls))
	}
	if calls[0].Name != "bash" {
		t.Errorf("Name = %q; want bash", calls[0].Name)
	}
	if cmd, ok := calls[0].Arguments["command"].(string); !ok || cmd != "echo hello" {
		t.Errorf("Arguments[command] = %v; want echo hello", calls[0].Arguments["command"])
	}
}

func TestExtractXmlToolCalls_NoMatch(t *testing.T) {
	calls := api.ExtractXmlToolCalls("plain text without any tool calls")
	if calls != nil {
		t.Errorf("expected nil, got %v", calls)
	}
}
