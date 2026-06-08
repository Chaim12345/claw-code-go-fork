package api

import (
	"strings"
	"testing"
)

func TestExtractXmlToolCalls_DeepSeekDSML(t *testing.T) {
	input := `<｜DSML｜tool_calls>
<｜DSML｜invoke name="bash">
<｜DSML｜parameter name="command" string="true">ls -la</｜DSML｜parameter>
</｜DSML｜invoke>
</｜DSML｜tool_calls>`

	calls := ExtractXmlToolCalls(input)
	if len(calls) != 1 {
		t.Fatalf("len(calls) = %d, want 1", len(calls))
	}
	if calls[0].Name != "bash" {
		t.Errorf("Name = %q, want bash", calls[0].Name)
	}
	if got, ok := calls[0].Arguments["command"].(string); !ok || got != "ls -la" {
		t.Errorf("command = %v, want %q", calls[0].Arguments["command"], "ls -la")
	}
}

func TestExtractXmlToolCalls_PlainXML(t *testing.T) {
	input := `<tool_calls>
<invoke name="bash"><parameter name="command">ls</parameter></invoke>
</tool_calls>`

	calls := ExtractXmlToolCalls(input)
	if len(calls) != 1 || calls[0].Name != "bash" {
		t.Fatalf("unexpected: %+v", calls)
	}
	if calls[0].Arguments["command"] != "ls" {
		t.Errorf("command = %v", calls[0].Arguments["command"])
	}
}

func TestExtractXmlToolCalls_DirectJSONBody(t *testing.T) {
	input := `<invoke name="bash">{"command":"ls -la"}</invoke>`

	calls := ExtractXmlToolCalls(input)
	if len(calls) != 1 || calls[0].Name != "bash" {
		t.Fatalf("unexpected: %+v", calls)
	}
	if calls[0].Arguments["command"] != "ls -la" {
		t.Errorf("command = %v", calls[0].Arguments["command"])
	}
}

func TestExtractXmlToolCalls_ToolNameWrapper(t *testing.T) {
	input := `<bash>ls</bash>`

	calls := ExtractXmlToolCalls(input)
	if len(calls) != 1 || calls[0].Name != "bash" {
		t.Fatalf("unexpected: %+v", calls)
	}
}

func TestExtractXmlToolCalls_MultipleCalls(t *testing.T) {
	input := `<tool_calls>
<invoke name="bash"><parameter name="command">ls</parameter></invoke>
<invoke name="read_file"><parameter name="path">/tmp/x</parameter></invoke>
</tool_calls>`

	calls := ExtractXmlToolCalls(input)
	if len(calls) != 2 {
		t.Fatalf("len(calls) = %d, want 2", len(calls))
	}
	if calls[0].Name != "bash" || calls[1].Name != "read_file" {
		t.Errorf("names = [%q, %q]", calls[0].Name, calls[1].Name)
	}
}

func TestExtractXmlToolCalls_TextAroundCalls(t *testing.T) {
	input := `I'll list the files for you.

<tool_calls>
<invoke name="bash"><parameter name="command">ls</parameter></invoke>
</tool_calls>

Let me know what else you need.`

	calls := ExtractXmlToolCalls(input)
	if len(calls) != 1 || calls[0].Name != "bash" {
		t.Fatalf("unexpected: %+v", calls)
	}
}

func TestExtractXmlToolCallsDetailed_SawSyntaxTrue(t *testing.T) {
	input := `<tool_calls>
<invoke name="bash"><parameter name="command">ls</parameter></invoke>
</tool_calls>`

	res := ExtractXmlToolCallsDetailed(input)
	if !res.SawToolCallSyntax {
		t.Error("SawToolCallSyntax = false, want true")
	}
	if len(res.Calls) != 1 {
		t.Errorf("len(Calls) = %d, want 1", len(res.Calls))
	}
}

func TestExtractXmlToolCallsDetailed_SawSyntaxFalse(t *testing.T) {
	res := ExtractXmlToolCallsDetailed("no tool calls here, just prose")
	if res.SawToolCallSyntax {
		t.Error("SawToolCallSyntax = true, want false")
	}
	if res.Calls != nil {
		t.Errorf("Calls = %v, want nil", res.Calls)
	}
}

func TestExtractXmlToolCallsDetailed_Empty(t *testing.T) {
	res := ExtractXmlToolCallsDetailed("")
	if res.SawToolCallSyntax || res.Calls != nil {
		t.Errorf("empty input = %+v, want zero value", res)
	}
}

func TestExtractXmlToolCallsDetailed_SyntaxButNoValidCall(t *testing.T) {
	input := `<tool_calls></tool_calls>`

	res := ExtractXmlToolCallsDetailed(input)
	if !res.SawToolCallSyntax {
		t.Error("SawToolCallSyntax = false, want true (saw <tool_calls>)")
	}
	if len(res.Calls) != 0 {
		t.Errorf("Calls = %v, want empty", res.Calls)
	}
}

func TestExtractXmlToolCalls_FullwidthASCII(t *testing.T) {
	input := "<｜DSML｜tool_calls>\n<｜DSML｜invoke name＝\"bash\">\n<｜DSML｜parameter name＝\"command\" string＝\"true\">ls</｜DSML｜parameter>\n</｜DSML｜invoke>\n</｜DSML｜tool_calls>"

	calls := ExtractXmlToolCalls(input)
	if len(calls) != 1 {
		t.Fatalf("len(calls) = %d, want 1 (fullwidth ASCII not normalized?)", len(calls))
	}
	if calls[0].Name != "bash" {
		t.Errorf("Name = %q, want bash", calls[0].Name)
	}
}

func TestExtractXmlToolCalls_CJKAngleBrackets(t *testing.T) {
	input := "〈｜DSML｜tool_calls〉\n〈｜DSML｜invoke name=\"bash\"〉\n〈｜DSML｜parameter name=\"command\" string=\"true\">ls〈/｜DSML｜parameter〉\n〈/｜DSML｜invoke〉\n〈/｜DSML｜tool_calls〉"

	calls := ExtractXmlToolCalls(input)
	if len(calls) != 1 {
		t.Fatalf("len(calls) = %d, want 1 (CJK drift not normalized?)", len(calls))
	}
	if calls[0].Name != "bash" {
		t.Errorf("Name = %q, want bash", calls[0].Name)
	}
}

func TestExtractXmlToolCalls_CollapsedDSML(t *testing.T) {
	input := `<dsmltool_calls>
<dsmlinvoke name="bash">
<dsmlparameter name="command" string="true">ls</dsmlparameter>
</dsmlinvoke>
</dsmltool_calls>`

	calls := ExtractXmlToolCalls(input)
	if len(calls) != 1 {
		t.Fatalf("len(calls) = %d, want 1 (collapsed DSML not handled?)", len(calls))
	}
}

func TestExtractXmlToolCalls_VendorPrefix(t *testing.T) {
	input := `<vendor|tool_calls>
<vendor|invoke name="bash">
<vendor|parameter name="command">ls</vendor|parameter>
</vendor|invoke>
</vendor|tool_calls>`

	calls := ExtractXmlToolCalls(input)
	if len(calls) != 1 {
		t.Fatalf("len(calls) = %d, want 1 (vendor prefix not stripped?)", len(calls))
	}
}

func TestExtractXmlToolCalls_CaseInsensitiveTag(t *testing.T) {
	input := `<Tool_Calls>
<Invoke name="bash"><Parameter name="command">ls</Parameter></Invoke>
</Tool_Calls>`

	calls := ExtractXmlToolCalls(input)
	if len(calls) != 1 {
		t.Fatalf("len(calls) = %d, want 1 (mixed case not handled?)", len(calls))
	}
}

func TestStripFencedCodeBlocks_TripleBacktick(t *testing.T) {
	input := "Look at this example:\n\n```\n<tool_calls>\n<invoke name=\"bash\"><parameter name=\"command\">rm -rf /</parameter></invoke>\n</tool_calls>\n```\n\nDon't actually run that."

	calls := ExtractXmlToolCalls(input)
	if len(calls) != 0 {
		t.Errorf("fenced example was parsed as real call: %+v", calls)
	}
}

func TestStripFencedCodeBlocks_TripleTilde(t *testing.T) {
	input := "Example:\n\n~~~\n<invoke name=\"bash\"><parameter name=\"command\">ls</parameter></invoke>\n~~~\n\nOK."

	calls := ExtractXmlToolCalls(input)
	if len(calls) != 0 {
		t.Errorf("~~~ fenced example parsed as real call: %+v", calls)
	}
}

func TestStripFencedCodeBlocks_LanguageHint(t *testing.T) {
	input := "```xml\n<invoke name=\"bash\"><parameter name=\"command\">ls</parameter></invoke>\n```"

	calls := ExtractXmlToolCalls(input)
	if len(calls) != 0 {
		t.Errorf("language-tagged fence parsed as real call: %+v", calls)
	}
}

func TestStripFencedCodeBlocks_RealCallOutsideFence(t *testing.T) {
	input := "```\nfake\n```\n<tool_calls><invoke name=\"bash\"><parameter name=\"command\">ls</parameter></invoke></tool_calls>"

	calls := ExtractXmlToolCalls(input)
	if len(calls) != 1 {
		t.Errorf("real call outside fence missed: %+v", calls)
	}
}

func TestSanitizeLooseCDATA_Unterminated(t *testing.T) {
	input := `<tool_calls><![CDATA[<invoke name="bash"><parameter name="command">ls</parameter></invoke></tool_calls>`

	calls := ExtractXmlToolCalls(input)
	if len(calls) != 1 {
		t.Fatalf("len(calls) = %d, want 1 (unterminated CDATA crashed scanner?)", len(calls))
	}
	if calls[0].Name != "bash" {
		t.Errorf("Name = %q, want bash", calls[0].Name)
	}
}

func TestSanitizeLooseCDATA_TerminatedIsFine(t *testing.T) {
	input := `<invoke name="bash"><parameter name="command"><![CDATA[ls -la]]></parameter></invoke>`

	calls := ExtractXmlToolCalls(input)
	if len(calls) != 1 {
		t.Fatalf("len(calls) = %d, want 1", len(calls))
	}
	cmd, _ := calls[0].Arguments["command"].(string)
	if !strings.Contains(cmd, "ls -la") {
		t.Errorf("command = %q, want contains 'ls -la'", cmd)
	}
}

func TestStripIgnorables(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"noop", "hello", "hello"},
		{"zero-width space", "he\u200Bllo", "hello"},
		{"BOM", "\uFEFFhello", "hello"},
		{"ZWJ", "he\u200Dllo", "hello"},
		{"multiple", "\u200B\u200C\uFEFFx\u200D\u200E", "x"},
		{"plain preserved", "abc 123", "abc 123"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := stripIgnorables(tc.in); got != tc.want {
				t.Errorf("stripIgnorables(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestNormalizeFullwidthASCII(t *testing.T) {
	if got := normalizeFullwidthASCII("a｜b"); got != "a|b" {
		t.Errorf("fullwidth | not converted: got %q", got)
	}
	if got := normalizeFullwidthASCII("a＝b"); got != "a=b" {
		t.Errorf("fullwidth = not converted: got %q", got)
	}
	if got := normalizeFullwidthASCII("a|b=c"); got != "a|b=c" {
		t.Errorf("ASCII changed: got %q", got)
	}
}

func TestStripVendorPrefix(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"abc|tool_calls", "tool_calls"},
		{"abc|invoke", "invoke"},
		{"abc|parameter", "parameter"},
		{"no-separator", ""},
		{"tool_calls", ""},
		{"random", ""},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			if got := stripVendorPrefix(tc.in); got != tc.want {
				t.Errorf("stripVendorPrefix(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestHasDSMLOrDrift(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"plain", false},
		{"<｜DSML｜tool_calls>", true},
		{"<｜DSML｜invoke name=...>", true},
		{"<tool_calls>", false},
		{"DSML", true},
		{"dsmltool_calls", true},
		{"<！DSML！tool_calls>", true},
		{"〈DSMLtool_calls〉", true},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			if got := hasDSMLOrDrift(tc.in); got != tc.want {
				t.Errorf("hasDSMLOrDrift(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestRepairLooseJSON_AlreadyValid(t *testing.T) {
	in := `{"command":"ls","path":"/tmp"}`
	if got := RepairLooseJSON(in); got != in {
		t.Errorf("valid JSON changed: %q → %q", in, got)
	}
}

func TestRepairLooseJSON_TrailingComma(t *testing.T) {
	in := `{"command":"ls",}`
	got := RepairLooseJSON(in)
	if !strings.Contains(got, `"command"`) {
		t.Errorf("trailing comma repair lost field: %q", got)
	}
}

func TestRepairLooseJSON_SingleQuotes(t *testing.T) {
	in := `{'command':'ls'}`
	got := RepairLooseJSON(in)
	if !strings.Contains(got, `"command"`) {
		t.Errorf("single-quote repair failed: %q", got)
	}
}

func TestRepairLooseJSON_UnquotedKeys(t *testing.T) {
	in := `{command:"ls"}`
	got := RepairLooseJSON(in)
	if !strings.Contains(got, `"command"`) {
		t.Errorf("unquoted key repair failed: %q", got)
	}
}

func TestRepairLooseJSON_PythonStyleBooleans(t *testing.T) {
	in := `{"flag":True,"other":False}`
	got := RepairLooseJSON(in)
	if !strings.Contains(got, `"flag":true`) || !strings.Contains(got, `"other":false`) {
		t.Errorf("Python booleans not repaired: %q", got)
	}
}

func TestRepairLooseJSON_NoneAndInfinity(t *testing.T) {
	in := `{"x":None,"y":Infinity}`
	got := RepairLooseJSON(in)
	if !strings.Contains(got, `"x":null`) {
		t.Errorf("None not converted to null: %q", got)
	}
	if !strings.Contains(got, `"y":null`) {
		t.Errorf("Infinity not converted to null: %q", got)
	}
}

func TestParseCDATAFragment_SimpleElement(t *testing.T) {
	cdata := `<item><name>x</name></item>`
	got, ok := parseCDATAFragment(cdata)
	if !ok {
		t.Fatal("parseCDATAFragment returned false")
	}
	m, ok := got.(map[string]any)
	if !ok {
		t.Fatalf("type = %T, want map[string]any", got)
	}
	inner, ok := m["name"].(map[string]any)
	if !ok {
		t.Fatalf("name type = %T", m["name"])
	}
	if inner["#text"] != "x" {
		t.Errorf("#text = %v, want x", inner["#text"])
	}
}

func TestParseCDATAFragment_RepeatedElements(t *testing.T) {
	cdata := `<root><x>1</x><x>2</x><x>3</x></root>`
	got, ok := parseCDATAFragment(cdata)
	if !ok {
		t.Fatal("parseCDATAFragment returned false")
	}
	m := got.(map[string]any)
	xs, ok := m["x"].([]any)
	if !ok {
		t.Fatalf("x is not a slice: %T (%+v)", m["x"], m)
	}
	if len(xs) != 3 {
		t.Errorf("len(xs) = %d, want 3", len(xs))
	}
}

func TestParseCDATAFragment_Attributes(t *testing.T) {
	cdata := `<item id="42">x</item>`
	got, ok := parseCDATAFragment(cdata)
	if !ok {
		t.Fatal("parseCDATAFragment returned false")
	}
	m := got.(map[string]any)
	if m["@id"] != "42" {
		t.Errorf("@id = %v, want 42", m["@id"])
	}
	if m["#text"] != "x" {
		t.Errorf("#text = %v, want x", m["#text"])
	}
}

func TestParseCDATAFragment_NotXML(t *testing.T) {
	if _, ok := parseCDATAFragment("just text"); ok {
		t.Error("plain text parsed as XML")
	}
	if _, ok := parseCDATAFragment(""); ok {
		t.Error("empty string parsed as XML")
	}
}

func TestExtractXmlToolCalls_NoCrashOnHostileInput(t *testing.T) {
	hostile := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"control bytes", "\x00\x01\x02"},
		{"bare angle brackets", "<>"},
		{"nested bare brackets", "<<>>"},
		{"many open brackets", "<<<<<"},
		{"unterminated open", "<tool_calls"},
		{"unterminated open invoke", "<invoke"},
		{"many closes", "</invoke></invoke></invoke>"},
		{"unterminated", "<invoke name=\"bash\"><parameter name=\"command\">"},
		{"truncated parameter", "<invoke name=\"bash\"><parameter"},
		{"space in tag name", "<inv oke name=\"bash\"/>"},
		{"unquoted attrs", "<invoke name=bash><parameter name=command>ls</parameter></invoke>"},
		{"stress", strings.Repeat("<invoke name=\"bash\"/>", 100)},
	}
	for _, tc := range hostile {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panicked on %q: %v", tc.in, r)
				}
			}()
			_ = ExtractXmlToolCalls(tc.in)
		})
	}
}

func TestExtractXmlToolCalls_NumericArguments(t *testing.T) {
	input := `<invoke name="bash"><parameter name="count" string="true">42</parameter></invoke>`
	calls := ExtractXmlToolCalls(input)
	if len(calls) != 1 {
		t.Fatalf("len = %d", len(calls))
	}
	if calls[0].Arguments["count"] != "42" {
		t.Errorf("count = %v", calls[0].Arguments["count"])
	}
}

func TestExtractXmlToolCalls_EmptyArguments(t *testing.T) {
	input := `<invoke name="bash"></invoke>`
	calls := ExtractXmlToolCalls(input)
	if len(calls) != 1 {
		t.Fatalf("len = %d", len(calls))
	}
	if calls[0].Name != "bash" {
		t.Errorf("name = %q", calls[0].Name)
	}
}

func TestExtractXmlToolCalls_StringAttrFlag(t *testing.T) {
	input := `<invoke name="write_file">
<parameter name="path" string="true">/tmp/file.json</parameter>
<parameter name="content" string="true">{"key": "value with \"escapes\""}</parameter>
</invoke>`

	calls := ExtractXmlToolCalls(input)
	if len(calls) != 1 {
		t.Fatalf("len = %d", len(calls))
	}
	content := calls[0].Arguments["content"].(string)
	if !strings.Contains(content, `"value with`) {
		t.Errorf("content lost JSON braces: %q", content)
	}
}
