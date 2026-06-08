package eval

import (
	"testing"

	"claw-code-go/internal/api"
)

// P4 — Format Resilience: table-driven tests that stress
// ExtractXmlToolCalls with noisy, malformed, or edge-case
// XML payloads that DeepSeek emits in the wild. Each case
// specifies the input text and the expected tool calls (name
// + key argument checks). The parser must extract the right
// calls without panicking or returning nil when tool calls
// are present.

func TestFormatResilience_DSMLInvoke(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantN   int           // expected number of tool calls
		checks  []toolCallCheck // per-call assertions
	}{
		{
			name: "standard DSML invoke",
			input: `<｜DSML｜tool_calls>
<｜DSML｜invoke name="bash">
<｜DSML｜parameter name="command" string="true">ls -la</｜DSML｜parameter>
</｜DSML｜invoke>
</｜DSML｜tool_calls>`,
			wantN:  1,
			checks: []toolCallCheck{{name: "bash", key: "command", val: "ls -la"}},
		},
		{
			name: "fullwidth bracket variant ［DSML］",
			input: `<［DSML］tool_calls>
<［DSML］invoke name="read_file">
<［DSML］parameter name="path" string="true">/tmp/x</［DSML］parameter>
</［DSML］invoke>
</［DSML］tool_calls>`,
			wantN:  1,
			checks: []toolCallCheck{{name: "read_file", key: "path", val: "/tmp/x"}},
		},
		{
			name: "mixed bracket ｜DSML］",
			input: `<｜DSML］tool_calls>
<｜DSML］invoke name="glob">
<｜DSML］parameter name="pattern" string="true">**/*.go</｜DSML］parameter>
</｜DSML］invoke>
</｜DSML］tool_calls>`,
			wantN:  1,
			checks: []toolCallCheck{{name: "glob", key: "pattern", val: "**/*.go"}},
		},
		{
			name: "bare DSML prefix (no opening bar)",
			input: `<DSML|tool_calls>
<DSML|invoke name="bash">
<DSML|parameter name="command" string="true">echo hi</DSML|parameter>
</DSML|invoke>
</DSML|tool_calls>`,
			wantN:  1,
			checks: []toolCallCheck{{name: "bash", key: "command", val: "echo hi"}},
		},
		{
			name: "plain invoke (no DSML prefix)",
			input: `<tool_calls>
<invoke name="write_file">
<parameter name="path" string="true">/tmp/out.txt</parameter>
<parameter name="content" string="true">hello world</parameter>
</invoke>
</tool_calls>`,
			wantN:  1,
			checks: []toolCallCheck{{name: "write_file", key: "path", val: "/tmp/out.txt"}},
		},
		{
			name: "invoke with direct-JSON body",
			input: `<invoke name="bash">{"command":"ls"}</invoke>`,
			wantN:  1,
			checks: []toolCallCheck{{name: "bash", key: "command", val: "ls"}},
		},
		{
			name: "function wrapper (pass 1b)",
			input: `<function name="bash">
<parameter name="command" string="true">pwd</parameter>
</function>`,
			wantN:  1,
			checks: []toolCallCheck{{name: "bash", key: "command", val: "pwd"}},
		},
		{
			name: "tool-name wrapper (pass 2)",
			input: `<bash>ls -la /tmp</bash>`,
			wantN:  1,
			checks: []toolCallCheck{{name: "bash", key: "command", val: "ls -la /tmp"}},
		},
		{
			name: "tool-name wrapper PascalCase",
			input: `<Bash>echo hello</Bash>`,
			wantN:  1,
			checks: []toolCallCheck{{name: "bash", key: "command", val: "echo hello"}},
		},
		{
			name: "multiple invoke blocks",
			input: `<invoke name="bash"><parameter name="command" string="true">ls</parameter></invoke>
<invoke name="read_file"><parameter name="path" string="true">/etc/hosts</parameter></invoke>`,
			wantN:  2,
			checks: []toolCallCheck{
				{name: "bash", key: "command", val: "ls"},
				{name: "read_file", key: "path", val: "/etc/hosts"},
			},
		},
		{
			name:    "empty input",
			input:   "",
			wantN:   0,
			checks:  nil,
		},
		{
			name:    "plain text with no tool calls",
			input:   "I'll help you with that task.",
			wantN:   0,
			checks:  nil,
		},
		{
			name:    "bare invoke no name attribute",
			input:   `<invoke>something</invoke>`,
			wantN:   0,
			checks:  nil,
		},
		{
			name: "parameter with string=false parses JSON number",
			input: `<invoke name="bash">
<parameter name="command" string="true">sleep</parameter>
<parameter name="timeout" string="false">30</parameter>
</invoke>`,
			wantN:  1,
			checks: []toolCallCheck{{name: "bash", key: "command", val: "sleep"}},
		},
		{
			name: "self-closing tag not treated as tool call",
			input: `Some text<br/>more text`,
			wantN:  0,
			checks: nil,
		},
		{
			name: "XML entities in parameter value",
			input: `<invoke name="bash">
<parameter name="command" string="true">echo &quot;hello&amp;world&quot;</parameter>
</invoke>`,
			wantN:  1,
			checks: []toolCallCheck{{name: "bash", key: "command", val: `echo "hello&world"`}},
		},
		{
			name: "pipe-delimited DSML prefix |DSML|",
			input: `<|DSML|tool_calls>
<|DSML|invoke name="grep">
<|DSML|parameter name="pattern" string="true">TODO</|DSML|parameter>
</|DSML|invoke>
</|DSML|tool_calls>`,
			wantN:  1,
			checks: []toolCallCheck{{name: "grep", key: "pattern", val: "TODO"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := api.ExtractXmlToolCalls(tt.input)
			if len(calls) != tt.wantN {
				t.Fatalf("ExtractXmlToolCalls returned %d calls, want %d", len(calls), tt.wantN)
			}
			for i, check := range tt.checks {
				if i >= len(calls) {
					break
				}
				if calls[i].Name != check.name {
					t.Errorf("call[%d].Name = %q, want %q", i, calls[i].Name, check.name)
				}
				val, ok := calls[i].Arguments[check.key]
				if !ok {
					t.Errorf("call[%d] missing key %q in arguments: %+v", i, check.key, calls[i].Arguments)
					continue
				}
				if s, _ := val.(string); s != check.val {
					t.Errorf("call[%d].Arguments[%q] = %q, want %q", i, check.key, s, check.val)
				}
			}
		})
	}
}

type toolCallCheck struct {
	name string
	key  string
	val  string
}
