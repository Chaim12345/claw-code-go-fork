package deepseek

import (
	"strings"
	"testing"

	"claw-code-go/internal/api"
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
		name        string
		input       string
		wantType    string
		wantThink   bool
		wantSearch  bool
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

func TestExtractSingleJsonToolCall(t *testing.T) {
	text := `{"tool":"read","path":"/tmp/foo"}`
	calls := extractSingleJsonToolCalls(text)
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Name != "read" {
		t.Errorf("name = %q, want read", calls[0].Name)
	}
	if calls[0].Arguments["path"] != "/tmp/foo" {
		t.Errorf("path = %v, want '/tmp/foo'", calls[0].Arguments["path"])
	}
}

func TestExtractXmlToolCalls(t *testing.T) {
	text := `Let me call a tool.
<tool_calls>
[{"name":"write","arguments":{"path":"/tmp/x","content":"y"}}]
</tool_calls>`
	calls := extractXmlToolCalls(text)
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Name != "write" {
		t.Errorf("name = %q, want write", calls[0].Name)
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
