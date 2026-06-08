package deepseek

import (
	"strings"
	"testing"

	"claw-code-go/internal/api"
	"claw-code-go/internal/tools"
)

func TestProviderName(t *testing.T) {
	p := New()
	if p.Name() != "deepseek" {
		t.Errorf("expected name 'deepseek', got %q", p.Name())
	}
}

func TestProviderAuthMethod(t *testing.T) {
	p := New()
	if p.AuthMethod() != api.AuthMethodAPIKey {
		t.Errorf("expected AuthMethodAPIKey, got %v", p.AuthMethod())
	}
}

func TestProviderNewClientMissingToken(t *testing.T) {
	t.Setenv("DEEPSEEK_TOKEN", "")
	t.Setenv("HOME", t.TempDir()) // ensure no ~/.deepseek reads
	p := New()
	_, err := p.NewClient(api.ProviderConfig{})
	if err == nil {
		t.Fatal("expected error when no token is configured")
	}
	if !strings.Contains(err.Error(), "no auth token") {
		t.Errorf("expected 'no auth token' error, got %v", err)
	}
}

func TestProviderNewClientWithToken(t *testing.T) {
	t.Setenv("DEEPSEEK_TOKEN", "test-token")
	p := New()
	client, err := p.NewClient(api.ProviderConfig{Model: "expert"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestProviderNewClientFallsBackToAPIKey(t *testing.T) {
	t.Setenv("DEEPSEEK_TOKEN", "")
	t.Setenv("HOME", t.TempDir())
	p := New()
	client, err := p.NewClient(api.ProviderConfig{APIKey: "inline-token"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client from inline APIKey")
	}
}

func TestParseModelName(t *testing.T) {
	cases := []struct {
		name       string
		input      string
		wantType   string
		wantThink  bool
		wantSearch bool
	}{
		{"empty", "", "default", false, false},
		{"instant", "instant", "default", false, false},
		{"default", "default", "default", false, false},
		{"deepseek-chat", "deepseek-chat", "default", false, false},
		{"instant-thinking", "instant-thinking", "default", true, false},
		{"instant-search", "instant-search", "default", false, true},
		{"instant-thinking-search", "instant-thinking-search", "default", true, true},
		{"expert", "expert", "expert", false, false},
		{"reasoner", "reasoner", "expert", false, false},
		{"r1", "r1", "expert", false, false},
		{"deepseek-reasoner", "deepseek-reasoner", "expert", false, false},
		{"expert-thinking", "expert-thinking", "expert", true, false},
		{"vision", "vision", "vision", false, false},
		{"vision-thinking", "vision-thinking", "vision", true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseModelName(tc.input)
			if got.ModelType != tc.wantType {
				t.Errorf("ModelType = %q, want %q", got.ModelType, tc.wantType)
			}
			if got.ThinkingEnabled != tc.wantThink {
				t.Errorf("ThinkingEnabled = %v, want %v", got.ThinkingEnabled, tc.wantThink)
			}
			if got.SearchEnabled != tc.wantSearch {
				t.Errorf("SearchEnabled = %v, want %v", got.SearchEnabled, tc.wantSearch)
			}
		})
	}
}

func TestModelSpecCanonicalName(t *testing.T) {
	cases := []struct {
		name string
		spec ModelSpec
		want string
	}{
		{"instant", ModelSpec{ModelType: "default"}, "Instant"},
		{"instant-thinking", ModelSpec{ModelType: "default", ThinkingEnabled: true}, "Instant+Thinking"},
		{"instant-search", ModelSpec{ModelType: "default", SearchEnabled: true}, "Instant+Search"},
		{"instant-both", ModelSpec{ModelType: "default", ThinkingEnabled: true, SearchEnabled: true}, "Instant+Thinking+Search"},
		{"expert", ModelSpec{ModelType: "expert"}, "Expert"},
		{"expert-thinking", ModelSpec{ModelType: "expert", ThinkingEnabled: true}, "Expert+Thinking"},
		{"vision", ModelSpec{ModelType: "vision"}, "Vision"},
		{"vision-thinking", ModelSpec{ModelType: "vision", ThinkingEnabled: true}, "Vision+Thinking"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.spec.CanonicalName(); got != tc.want {
				t.Errorf("CanonicalName() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestExtractJsonToolCalls(t *testing.T) {
	text := `Some preamble text.
{"tool_calls":[{"name":"bash","arguments":{"command":"ls -la"}}]}
Trailing text.`
	calls := extractJsonToolCalls(text)
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Name != "bash" {
		t.Errorf("name = %q, want bash", calls[0].Name)
	}
	if cmd, _ := calls[0].Arguments["command"].(string); cmd != "ls -la" {
		t.Errorf("command = %q, want 'ls -la'", cmd)
	}
}

func TestExtractJsonToolCallsUnknownTool(t *testing.T) {
	text := `{"tool_calls":[{"name":"unknown_tool","arguments":{}}]}`
	calls := extractJsonToolCalls(text)
	if len(calls) != 0 {
		t.Errorf("expected 0 calls for unknown tool, got %d", len(calls))
	}
}

// TestExtractJsonToolCallsFlatShape covers the shape the DeepSeek web
// model actually emits in the wild: arguments are flat alongside the
// tool name, NOT wrapped in an "arguments" object. This is the format
// observed in TUI sessions where the bash tool is called with
// {"tool_calls":[{"name":"bash","command":"find ..."}]}.
//
// The existing extractors only handle the wrapped form
// {"tool_calls":[{"name":"bash","arguments":{"command":"..."}}]},
// so the flat shape parses to an empty arguments map and the bash
// tool fails with "'command' input is required".
func TestExtractJsonToolCallsFlatShape(t *testing.T) {
	cases := []struct {
		name     string
		text     string
		wantTool string
		wantArg  string
		wantKey  string
	}{
		{
			name:     "bash flat",
			text:     `{"tool_calls":[{"name":"bash","command":"ls -la"}]}`,
			wantTool: "bash",
			wantKey:  "command",
			wantArg:  "ls -la",
		},
		{
			name:     "read_file flat",
			text:     `{"tool_calls":[{"name":"read_file","path":"/etc/hosts"}]}`,
			wantTool: "read_file",
			wantKey:  "path",
			wantArg:  "/etc/hosts",
		},
		{
			name:     "wrapped still works (regression)",
			text:     `{"tool_calls":[{"name":"bash","arguments":{"command":"echo hi"}}]}`,
			wantTool: "bash",
			wantKey:  "command",
			wantArg:  "echo hi",
		},
		{
			name:     "flat with metadata keys (name, id) — must strip",
			text:     `{"tool_calls":[{"id":"x1","name":"bash","command":"pwd"}]}`,
			wantTool: "bash",
			wantKey:  "command",
			wantArg:  "pwd",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			calls := extractJsonToolCalls(tc.text)
			if len(calls) != 1 {
				t.Fatalf("expected 1 call, got %d (text=%q)", len(calls), tc.text)
			}
			if calls[0].Name != tc.wantTool {
				t.Errorf("name = %q, want %q", calls[0].Name, tc.wantTool)
			}
			if got, _ := calls[0].Arguments[tc.wantKey].(string); got != tc.wantArg {
				t.Errorf("arguments[%q] = %q, want %q (full args: %+v)",
					tc.wantKey, got, tc.wantArg, calls[0].Arguments)
			}
			// Metadata keys must not leak into the args map.
			if _, leaked := calls[0].Arguments["name"]; leaked {
				t.Errorf("'name' leaked into arguments: %+v", calls[0].Arguments)
			}
			if _, leaked := calls[0].Arguments["id"]; leaked {
				t.Errorf("'id' leaked into arguments: %+v", calls[0].Arguments)
			}
		})
	}
}

func TestExtractSingleJsonToolCall(t *testing.T) {
	text := `{"tool":"read_file","path":"/tmp/foo"}`
	calls := extractSingleJsonToolCalls(text)
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Name != "read_file" {
		t.Errorf("name = %q, want read_file", calls[0].Name)
	}
	if calls[0].Arguments["path"] != "/tmp/foo" {
		t.Errorf("path = %v, want '/tmp/foo'", calls[0].Arguments["path"])
	}
}

// TestExtractJsonToolCallsNoisyText covers the case the new
// jsonex-based extractor was introduced to fix: tool calls embedded in
// prose, reasoning, or other contamination that the previous
// hand-rolled scanner would have rejected. jsonex extracts the longest
// valid JSON object, ignoring surrounding noise.
func TestExtractJsonToolCallsNoisyText(t *testing.T) {
	cases := []struct {
		name    string
		text    string
		want    string // expected tool name
		wantKey string
		wantArg string
	}{
		{
			name:    "prose before and after",
			text:    `Sure! I'll run that for you.\n\n{"tool_calls":[{"name":"bash","arguments":{"command":"ls -la"}}]}\n\nLet me know if you need anything else.`,
			want:    "bash",
			wantKey: "command",
			wantArg: "ls -la",
		},
		{
			name:    "reasoning then tool call",
			text:    `I need to check the directory structure. Let me list it.\n{"tool_calls":[{"name":"bash","command":"ls"}]}`,
			want:    "bash",
			wantKey: "command",
			wantArg: "ls",
		},
		{
			name:    "single tool in prose",
			text:    `Here's the call: {"tool":"read_file","path":"/etc/hosts"} -- done.`,
			want:    "read_file",
			wantKey: "path",
			wantArg: "/etc/hosts",
		},
		{
			name:    "escaped quotes in command",
			text:    `{"tool_calls":[{"name":"bash","arguments":{"command":"echo \"hi\""}}]}`,
			want:    "bash",
			wantKey: "command",
			wantArg: `echo "hi"`,
		},
		{
			name:    "nested braces in command",
			text:    `{"tool_calls":[{"name":"bash","arguments":{"command":"echo {1,2,3}"}}]}`,
			want:    "bash",
			wantKey: "command",
			wantArg: "echo {1,2,3}",
		},
		{
			name:    "object with non-JSON noise before it",
			text:    `blah blah {not valid} {"tool_calls":[{"name":"grep","pattern":"TODO"}]}`,
			want:    "grep",
			wantKey: "pattern",
			wantArg: "TODO",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Use unescaped \n — the test data above is literal
			// text, not a Go string with escapes. Rebuild it that
			// way to keep the test cases readable.
			text := tc.text
			if idx := strings.Index(text, `\n`); idx != -1 {
				text = strings.ReplaceAll(text, `\n`, "\n")
			}
			calls := ExtractToolCalls(text)
			if len(calls) == 0 {
				t.Fatalf("expected at least 1 call, got 0 (text=%q)", text)
			}
			if calls[0].Name != tc.want {
				t.Errorf("name = %q, want %q", calls[0].Name, tc.want)
			}
			if got, _ := calls[0].Arguments[tc.wantKey].(string); got != tc.wantArg {
				t.Errorf("arguments[%q] = %q, want %q (full args: %+v)",
					tc.wantKey, got, tc.wantArg, calls[0].Arguments)
			}
		})
	}
}

// TestExtractJsonToolCallsMultipleBlocks covers the case the streaming
// jsonex decoder was introduced to handle: the model emits more than
// one tool-call block in its response (perhaps separated by reasoning).
// The old extractor tried only the first occurrence and would miss
// the second.
func TestExtractJsonToolCallsMultipleBlocks(t *testing.T) {
	text := `
First I'll check the directory:
{"tool_calls":[{"name":"bash","arguments":{"command":"ls"}}]}

Now the file count:
{"tool_calls":[{"name":"bash","arguments":{"command":"wc -l file.txt"}}]}
`
	calls := ExtractToolCalls(text)
	if len(calls) < 2 {
		t.Fatalf("expected at least 2 calls across the two blocks, got %d", len(calls))
	}
	// We don't assert order — jsonex scans sequentially, but
	// normalisation may emit either order depending on shape. Just
	// confirm we got both.
	got := map[string]bool{}
	for _, c := range calls {
		got[c.Arguments["command"].(string)] = true
	}
	if !got["ls"] || !got["wc -l file.txt"] {
		t.Errorf("expected both commands in calls, got: %+v", calls)
	}
}

// TestCollapseExcessiveToolCallsPreservesPrefix locks in the fix for
// the prefix-loss bug: when several empty <tool_calls></tool_calls>
// blocks precede the real one, the result must keep the prose that
// came before the noise run.
func TestCollapseExcessiveToolCallsPreservesPrefix(t *testing.T) {
	noise := strings.Repeat("<tool_calls></tool_calls>", 5)
	prefix := "Let me think about this for a moment."
	real := `<tool_calls>[{"name":"bash","command":"ls"}]</tool_calls>`
	suffix := "Then I'll continue."
	text := prefix + noise + real + suffix

	got := collapseExcessiveToolCalls(text)
	if !strings.Contains(got, prefix) {
		t.Errorf("expected prefix %q to be preserved, got %q", prefix, got)
	}
	if !strings.Contains(got, real) {
		t.Errorf("expected real block %q to be preserved, got %q", real, got)
	}
	if !strings.Contains(got, suffix) {
		t.Errorf("expected suffix %q to be preserved, got %q", suffix, got)
	}
	// And the empty echoes must be gone (only one <tool_calls> remains).
	if strings.Count(got, "<tool_calls>") != 1 {
		t.Errorf("expected exactly 1 <tool_calls>, got %d in %q",
			strings.Count(got, "<tool_calls>"), got)
	}
}

// TestExtractFunctionCallToolCallsJSONBody covers the previous bug
// where <tool_call> XML only had the name extracted and the args
// body was silently dropped (returning `{}` and breaking the
// downstream tool). The fix parses the body as JSON.
func TestExtractFunctionCallToolCallsJSONBody(t *testing.T) {
	text := `<function_call name="bash">
{"command": "ls -la"}
</function_call>`
	calls := extractFunctionCallToolCalls(text)
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Name != "bash" {
		t.Errorf("name = %q, want bash", calls[0].Name)
	}
	if got, _ := calls[0].Arguments["command"].(string); got != "ls -la" {
		t.Errorf("command = %q, want 'ls -la' (full args: %+v)", got, calls[0].Arguments)
	}
}

// TestExtractFunctionCallToolCallsUnparseableBody verifies the fix
// chose to skip rather than return an empty-args stub when the body
// isn't JSON. The old behaviour returned `{}` and the tool would
// fail with "input is required" — skipping lets the XML/JSON
// fallbacks pick up the call on the next extractor.
func TestExtractFunctionCallToolCallsUnparseableBody(t *testing.T) {
	text := `<function_call name="bash">not json at all</function_call>`
	calls := extractFunctionCallToolCalls(text)
	if len(calls) != 0 {
		t.Errorf("expected 0 calls (unparseable body should be skipped, not return broken empty-args), got %d: %+v", len(calls), calls)
	}
}

// TestExtractReactToolCallsNoFakeShape ensures the ReAct extractor no
// longer returns a fabricated {"input": "..."} arg map when the
// Action Input line isn't valid JSON. The old shape was inconsistent
// with the rest of the extractors and broke tools that expected a
// typed schema.
func TestExtractReactToolCallsNoFakeShape(t *testing.T) {
	text := "Action: bash\nAction Input: not json"
	calls := extractReactToolCalls(text)
	if len(calls) != 0 {
		t.Errorf("expected 0 calls (unparseable input should be skipped), got %d: %+v", len(calls), calls)
	}
}

func TestExtractXmlToolCalls(t *testing.T) {
	// Real DeepSeek/Qwen-style XML: <invoke name="tool"> with
	// <parameter name="arg">value</parameter> children, inside a
	// <tool_calls> wrapper. The previous parser only handled
	// JSON-in-XML-wrapper and silently dropped this format — the
	// exact bug that produced "tools are still not executed or
	// parsed" reports. The new api.ExtractXmlToolCalls must
	// extract both invokes below with their parameters intact.
	text := `Let me call a tool.
<tool_calls>
<invoke name="write_file">
<parameter name="path">/tmp/x</parameter>
<parameter name="content">y</parameter>
</invoke>
<invoke name="bash">
<parameter name="command">ls -la</parameter>
</invoke>
</tool_calls>`
	calls := extractXmlToolCalls(text)
	if len(calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(calls))
	}
	if calls[0].Name != "write_file" || calls[1].Name != "bash" {
		t.Errorf("names = [%q, %q], want [write_file, bash]", calls[0].Name, calls[1].Name)
	}
	if got := calls[0].Arguments["path"]; got != "/tmp/x" {
		t.Errorf("write_file path = %q, want /tmp/x", got)
	}
	if got := calls[1].Arguments["command"]; got != "ls -la" {
		t.Errorf("bash command = %q, want 'ls -la'", got)
	}
}

func TestExtractCodeBlockToolCalls(t *testing.T) {
	text := "Here you go:\n```\n{\"tool\":\"bash\",\"command\":\"echo hi\"}\n```\nDone."
	calls := extractCodeBlockToolCalls(text)
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Name != "bash" {
		t.Errorf("name = %q, want bash", calls[0].Name)
	}
}

func TestStripToolCallsRemovesBlock(t *testing.T) {
	text := "Hello.\n<tool_calls>\n[{\"name\":\"bash\",\"arguments\":{\"command\":\"ls\"}}]\n</tool_calls>\nWorld."
	got := StripToolCalls(text)
	if strings.Contains(got, "tool_calls") {
		t.Errorf("expected tool_calls to be removed, got %q", got)
	}
	if !strings.Contains(got, "Hello.") || !strings.Contains(got, "World.") {
		t.Errorf("expected non-tool text to remain, got %q", got)
	}
}

func TestEstimateTokens(t *testing.T) {
	cases := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"a", 1},
		{"abcd", 1},
		{"abcde", 2},
		{strings.Repeat("x", 400), 100},
	}
	for _, tc := range cases {
		if got := EstimateTokens(tc.input); got != tc.want {
			t.Errorf("EstimateTokens(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestFitToBudget(t *testing.T) {
	short := "hello"
	if got := FitToBudget(short, 100); got != short {
		t.Errorf("short text should be unchanged, got %q", got)
	}

	long := strings.Repeat("a", 5000)
	got := FitToBudget(long, 100) // 100 tokens * 4 chars = 400 max
	if len(got) >= len(long) {
		t.Errorf("long text should be truncated, len(got)=%d len(long)=%d", len(got), len(long))
	}
	if !strings.Contains(got, "truncated") {
		t.Errorf("expected truncation marker, got %q", got)
	}
}

func TestBuildPrompt(t *testing.T) {
	system := "You are a helper."
	messages := []api.Message{
		{Role: "user", Content: []api.ContentBlock{{Type: "text", Text: "hello"}}},
		{Role: "assistant", Content: []api.ContentBlock{{Type: "text", Text: "hi there"}}},
	}
	tools := []api.Tool{
		{Name: "bash", Description: "run commands", InputSchema: api.InputSchema{
			Type: "object",
			Properties: map[string]api.Property{
				"command": {Type: "string", Description: "the command"},
			},
			Required: []string{"command"},
		}},
	}

	got := buildPrompt(system, messages, tools)
	if !strings.Contains(got, system) {
		t.Error("system prompt missing from built prompt")
	}
	if !strings.Contains(got, "[User]\nhello") {
		t.Error("user turn missing from built prompt")
	}
	if !strings.Contains(got, "[Assistant]\nhi there") {
		t.Error("assistant turn missing from built prompt")
	}
	if !strings.Contains(got, "Available tools") {
		t.Error("tool description missing from built prompt")
	}
	if !strings.Contains(got, "bash") {
		t.Error("bash tool name missing")
	}
}

func TestWebClientInit(t *testing.T) {
	wc := NewWebClient("test-token")
	if wc == nil {
		t.Fatal("expected non-nil WebClient")
	}
	if wc.BaseURL != "https://chat.deepseek.com" {
		t.Errorf("BaseURL = %q, want https://chat.deepseek.com", wc.BaseURL)
	}
	if wc.Token != "test-token" {
		t.Errorf("Token = %q, want test-token", wc.Token)
	}
}

func TestParseDeepSeekSseData(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "v.response.fragments",
			input: `{"v":{"response":{"fragments":[{"content":"hello"}]}}}`,
			want:  "hello",
		},
		{
			name:  "p response/fragments append",
			input: `{"p":"response/fragments","o":"APPEND","v":[{"content":"world"}]}`,
			want:  "world",
		},
		{
			name:  "p ending in /content",
			input: `{"p":"foo/content","v":"hi"}`,
			want:  "hi",
		},
		{
			name:  "openai delta fallback",
			input: `{"choices":[{"delta":{"content":"ok"}}]}`,
			want:  "ok",
		},
		{
			name:  "empty",
			input: `{}`,
			want:  "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := parseDeepSeekSseData(tc.input); got != tc.want {
				t.Errorf("parseDeepSeekSseData() = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestParseSSEStopsOnStatusFinished pins the new contract that
// parseSSE aborts reading as soon as the handler returns false on a
// quasi_status=FINISHED BATCH frame. The previous implementation
// relied on string-suffixing a literal "FINISHED" off content deltas,
// which silently corrupted user-visible text whenever the model
// happened to use the word "FINISHED" naturally. The new contract
// uses the server's authoritative FINISHED event to clip the stream
// at the right boundary, and keeps the content deltas verbatim.
func TestParseSSEStopsOnStatusFinished(t *testing.T) {
	const body = `data: {"p":"response","o":"BATCH","v":[{"p":"quasi_status","v":"FINISHED"}]}

data: {"p":"response/fragments","o":"APPEND","v":[{"content":"should be ignored"}]}

data: [DONE]

`
	var got []string
	parseSSE(strings.NewReader(body), func(ev StreamEvent) bool {
		got = append(got, ev.Event+":"+ev.Data)
		// Stop on the first status FINISHED frame, mimicking
		// the provider's behavior in StreamResponse.
		if ev.Event == "status" && ev.Data == "FINISHED" {
			return false
		}
		return true
	})

	want := []string{"status:FINISHED"}
	if len(got) != len(want) {
		t.Fatalf("parseSSE invoked handler %d times, want %d (events: %v)", len(got), len(want), got)
	}
	if got[0] != want[0] {
		t.Errorf("parseSSE first event = %q, want %q", got[0], want[0])
	}
}

// TestParseSSEPropagatesStatusFinishedContent verifies that content
// events delivered BEFORE the FINISHED status frame are emitted to
// the handler in order, so the provider's fullText buffer sees the
// real model output. (The previous test only checks the stop
// boundary; this one pins the accumulate-then-stop ordering the
// provider depends on.)
func TestParseSSEPropagatesStatusFinishedContent(t *testing.T) {
	const body = `data: {"p":"response","o":"BATCH","v":[{"p":"accumulated_token_usage","v":42}]}

data: {"p":"response/fragments","o":"APPEND","v":[{"content":"hello "}]}

data: {"p":"response/fragments","o":"APPEND","v":[{"content":"world"}]}

data: {"p":"response","o":"BATCH","v":[{"p":"quasi_status","v":"FINISHED"}]}

`
	var got []StreamEvent
	parseSSE(strings.NewReader(body), func(ev StreamEvent) bool {
		got = append(got, ev)
		return ev.Event != "status" || ev.Data != "FINISHED"
	})

	want := []struct{ event, data string }{
		{"usage", "42"},
		{"content", "hello "},
		{"content", "world"},
		{"status", "FINISHED"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d events, want %d (events: %+v)", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i].Event != w.event || got[i].Data != w.data {
			t.Errorf("event %d = {%q, %q}, want {%q, %q}", i, got[i].Event, got[i].Data, w.event, w.data)
		}
	}
}

// TestKnownToolsMatchExecutor pins the contract between the parser's
// knownTools map and the runtime's tool executor (tools.*Tool()
// constructors and the corresponding case labels in
// ConversationLoop.ExecuteTool / ExecuteToolQuiet in
// internal/runtime/conversation.go). Any drift between the two
// surfaces causes the parser to accept a tool name that the
// executor then rejects as "unknown tool" — a silent dead-letter
// path. If you add a new tool to one place, add it to the other and
// this test will tell you if you forgot.
func TestKnownToolsMatchExecutor(t *testing.T) {
	executable := map[string]bool{
		tools.BashTool().Name:            true,
		tools.ReadFileTool().Name:        true,
		tools.WriteFileTool().Name:       true,
		tools.FileEditTool().Name:        true,
		tools.GlobTool().Name:            true,
		tools.GrepTool().Name:            true,
		tools.WebFetchTool().Name:        true,
		tools.WebSearchTool().Name:       true,
		tools.AskUserQuestionTool().Name: true,
		tools.TodoWriteTool().Name:       true,
	}
	// The parser must accept every tool the executor can run.
	for name := range executable {
		if !knownTools[name] {
			t.Errorf("parser knownTools is missing %q (executor can run it)", name)
		}
	}
	// And it must NOT accept tools the executor can't run — otherwise
	// the conversation loop will dispatch an unknown name to the MCP
	// fallback and produce a confusing "unknown tool" error.
	for name := range knownTools {
		if !executable[name] {
			t.Errorf("parser knownTools advertises %q but the executor has no case for it (will produce 'unknown tool' at runtime)", name)
		}
	}
}

// TestTrimFinishedSentinel pins the heuristic for stripping the
// model's learned end-of-turn marker. The DeepSeek web model
// splices the token onto the last word ("...layers.FINISHED") or
// emits it on its own line; the trim must handle both.
func TestTrimFinishedSentinel(t *testing.T) {
	cases := []struct {
		desc, in, want string
	}{
		{"splice onto last word", "...layers.FINISHED", "...layers"},
		{"splice onto mid-sentence word", "answer hereFINISHED", "answer here"},
		{"sentinel on its own paragraph", "first paragraph\n\nFINISHED", "first paragraph"},
		{"sentinel on its own line", "first line\nFINISHED", "first line"},
		{"sentinel with trailing whitespace", "answer.FINISHED   \n", "answer"},
		{"text without FINISHED is untouched", "no marker here", "no marker here"},
		{"empty stays empty", "", ""},
		{"just FINISHED alone becomes empty", "FINISHED", ""},
		{"FINISHED in middle is untouched", "FINISHED then more", "FINISHED then more"},
		{"English ending is also trimmed (acceptable false positive)", "we are FINISHED", "we are"},
	}
	for _, c := range cases {
		got := trimFinishedSentinel(c.in)
		if got != c.want {
			t.Errorf("%s: in=%q got=%q want=%q", c.desc, c.in, got, c.want)
		}
	}
}

// TestParseSSEDetectsRawMuteJSON verifies that parseSSE detects a raw JSON
// mute response (no "data:" prefix) and emits an error event instead of
// silently skipping it. This is the primary fix for the "empty 200 on mute"
// bug — the DeepSeek server returns a JSON body without SSE framing when
// the user is rate-limited.
func TestParseSSEDetectsRawMuteJSON(t *testing.T) {
	const body = `{"code":0,"msg":"","data":{"biz_code":5,"biz_msg":"user is muted","biz_data":{"is_muted":1,"mute_until":1735689600}}}`
	var got []StreamEvent
	parseSSE(strings.NewReader(body), func(ev StreamEvent) bool {
		got = append(got, ev)
		return true
	})
	if len(got) != 1 {
		t.Fatalf("expected 1 error event, got %d: %+v", len(got), got)
	}
	if got[0].Event != "error" {
		t.Errorf("event type = %q, want %q", got[0].Event, "error")
	}
	if !strings.Contains(got[0].Data, "user is muted") {
		t.Errorf("event data = %q, want it to contain 'user is muted'", got[0].Data)
	}
}

// TestParseSSEDetectsRawMsgError verifies that a raw JSON response with a
// top-level "msg" error field (but no biz_msg) is also caught.
func TestParseSSEDetectsRawMsgError(t *testing.T) {
	const body = `{"code":1,"msg":"session expired","data":null}`
	var got []StreamEvent
	parseSSE(strings.NewReader(body), func(ev StreamEvent) bool {
		got = append(got, ev)
		return true
	})
	if len(got) != 1 {
		t.Fatalf("expected 1 error event, got %d: %+v", len(got), got)
	}
	if got[0].Event != "error" {
		t.Errorf("event type = %q, want %q", got[0].Event, "error")
	}
	if !strings.Contains(got[0].Data, "session expired") {
		t.Errorf("event data = %q, want it to contain 'session expired'", got[0].Data)
	}
}

// TestParseSSESkipsNonErrorNonDataLines verifies that parseSSE still silently
// skips lines that are neither "data:" prefixed nor raw JSON error objects.
func TestParseSSESkipsNonErrorNonDataLines(t *testing.T) {
	const body = `: this is a comment
event: ping

data: {"p":"response/fragments","o":"APPEND","v":[{"content":"hello"}]}
`
	var got []StreamEvent
	parseSSE(strings.NewReader(body), func(ev StreamEvent) bool {
		got = append(got, ev)
		return true
	})
	// Should only get the content event from the data: line
	if len(got) != 1 {
		t.Fatalf("expected 1 content event, got %d: %+v", len(got), got)
	}
	if got[0].Event != "content" {
		t.Errorf("event type = %q, want %q", got[0].Event, "content")
	}
}

// TestIsRateLimitErrorMuted verifies that "user is muted" is detected as a
// rate limit error, so the provider's retry+rotation logic kicks in.
func TestIsRateLimitErrorMuted(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"user is muted", true},
		{"User Is Muted", true},
		{"deepseek: user is muted (retry after 2025-01-01T00:00:00Z)", true},
		{"is_muted", true},
		{"too frequent", true},
		{"server is busy", true},
		{"normal error", false},
		{"connection reset", false},
	}
	for _, c := range cases {
		got := IsRateLimitError(c.input)
		if got != c.want {
			t.Errorf("IsRateLimitError(%q) = %v, want %v", c.input, got, c.want)
		}
	}
}