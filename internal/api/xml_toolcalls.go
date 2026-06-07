package api

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"html"
	"regexp"
	"strings"
	"unicode/utf8"
)

// firstRequiredParam maps tool names to their first required input
// parameter key. This is used by Pass 2 of ExtractXmlToolCalls when
// the model emits a tool-name wrapper with a non-JSON body
// (e.g. <bash>ls -la</bash>). Without this mapping, the raw body
// would be stored under the key "input", which every tool executor
// rejects with "'<name>' input is required".
//
// The mapping is derived from each tool's InputSchema.Required[0].
// grep and file_edit are not included because they have 2+ required
// params (a raw body can't satisfy them), so they'll still fall
// through to the generic "input" key which produces a clear error.
var firstRequiredParam = map[string]string{
	"bash":       "command",
	"read_file":  "path",
	"write_file": "path",
	"glob":       "pattern",
	"todo_write": "action",
	"web_fetch":  "url",
	"web_search": "query",
	"ask_user":   "question",
}

// XmlToolCallResult is the detailed return type for ExtractXmlToolCallsDetailed.
// It distinguishes between "no tool call syntax detected" and "tool call
// syntax detected but no valid calls extracted" via the SawToolCallSyntax flag.
type XmlToolCallResult struct {
	Calls             []ToolCall
	SawToolCallSyntax bool
}

// ExtractXmlToolCalls parses XML-formatted tool calls from model output.
// The model emits several variants in the wild, and this single
// scanner handles all of them:
//
//  1. DeepSeek DSML (the format chat.deepseek.com uses):
//     <｜DSML｜invoke name="bash">
//     <｜DSML｜parameter name="command" string="true">ls -la</｜DSML｜parameter>
//     </｜DSML｜invoke>
//     wrapped in <｜DSML｜tool_calls>...</｜DSML｜tool_calls>.
//
//  2. Plain XML (when the system prompt doesn't use DSML tokens):
//     <invoke name="bash"><parameter name="command">ls</parameter></invoke>
//     wrapped in <tool_calls>...</tool_calls>.
//
//  3. Direct-JSON body of <invoke> (sglang's "Format 2", observed in
//     v3.2 outputs):
//     <invoke name="bash">{"command":"ls"}</invoke>
//
//  4. Tool-name wrappers (case-insensitive, the rare-but-real case
//     where the model emits a "thought" with the tool name as a
//     PascalCase or lowercase tag, e.g. <Bash>ls</Bash> or
//     <bash>ls</bash>):
//     any tag whose lowercased name is a plausible tool name and
//     whose body is the raw argument string.
//
//  5. Unicode-drifted DSML: fullwidth ASCII drift (`＜`, `＞`, `＝`),
//     CJK angle brackets (`〈`, `〉`), fullwidth bang separators
//     (`！`), Unicode confusables (zero-width chars, Cyrillic/Greek
//     lookalikes), collapsed DSML tags (`<DSMLtool_calls>`), and
//     arbitrary vendor prefixes (`<abc|tool_calls>`,
//     `<vendor_tool_calls>`).
//
//  6. Markdown-fenced tool call examples: tool call syntax inside
//     ``` fenced code blocks is ignored (not executed).
//
// We deliberately do NOT use encoding/xml. The model output has
// CDATA-ish noise (unescaped &, bare < and >, comments, DSML
// multibyte prefixes), self-closing tag variants, and other patterns
// that would make a strict parser fail on the majority of real
// responses. Instead, this is a single-pass byte-level scanner.
//
// Tag matching is case-insensitive, DSML-prefix-tolerant, and
// Unicode-drift-tolerant. The close-tag matcher uses the same
// normalization so a tag opened as <｜DSML｜Invoke> closes on
// </｜DSML｜invoke>. The caller is responsible for filtering Name
// against the per-provider knownTools list — this parser is
// intentionally permissive so a future provider (or a new tool
// name) doesn't require a code change here.
//
// The hot path uses bytes.Index / bytes.IndexByte rather than
// strings.Index so the body never has to be re-encoded — on a 4 KB
// tool block this is roughly 2-3x faster than the equivalent
// strings.Index loop.
//
// Returns nil if no tool calls are found.
func ExtractXmlToolCalls(text string) []ToolCall {
	return ExtractXmlToolCallsDetailed(text).Calls
}

// ExtractXmlToolCallsDetailed is the same as ExtractXmlToolCalls
// but also reports whether tool call syntax was detected (even if no
// valid calls were extracted). This lets callers distinguish between
// "model said nothing" and "model tried to call a tool but it was
// rejected/empty".
func ExtractXmlToolCallsDetailed(text string) XmlToolCallResult {
	if len(text) == 0 {
		return XmlToolCallResult{}
	}

	// Pre-processing pipeline (order matters):
	// 1. Strip markdown fenced code blocks (``` and ~~~) so tool call
	//    syntax in examples is not executed.
	// 2. Repair malformed CDATA (unterminated <![CDATA[) so the
	//    remaining content can still be parsed.
	// 3. Canonicalize Unicode-drifted DSML markup to canonical ASCII
	//    so the main scanner doesn't have to handle every drift
	//    variant.
	processed := text
	processed = stripFencedCodeBlocks(processed)
	processed = sanitizeLooseCDATA(processed)
	processed = normalizeDSMLMarkup(processed)

	b := []byte(processed)

	var calls []ToolCall
	sawSyntax := bytes.Contains(b, []byte("<tool_calls")) ||
		bytes.Contains(b, []byte("</tool_calls>")) ||
		bytes.Contains(b, []byte("<function_calls")) ||
		hasDSMLOrDrift(processed)

	// Pass 1: <invoke ...> blocks (the documented DeepSeek/Qwen
	// format, both with and without the DSML prefix and with both
	// XML-parameter and direct-JSON body variants).
	pos := 0
	for pos < len(b) {
		start, _, name, ok := findOpeningTag(b, pos, "invoke")
		if !ok {
			break
		}
		sawSyntax = true
		bodyStart := findTagEnd(b, start)
		if bodyStart < 0 {
			break
		}
		bodyEnd, bodyOK := findCloseTag(b, bodyStart, "invoke")
		if !bodyOK {
			break
		}
		if name == "" {
			// Bare <invoke> with no name attribute — skip
			// rather than producing a tool call named "".
			pos = bodyEnd
			continue
		}
		args := scanInvokeBody(b[bodyStart:bodyEnd], name)
		calls = append(calls, ToolCall{Name: name, Arguments: args})
		pos = bodyEnd
	}

	// Pass 1b: <function name="..."> blocks — the same DSML-style
	// parameter format as <invoke>, but with a different wrapper
	// tag. The DeepSeek expert model emits this when the system
	// prompt doesn't explicitly prescribe a wrapper. Only runs if
	// pass 1 found nothing (same dedup logic as pass 2).
	if len(calls) == 0 {
		pos := 0
		for pos < len(b) {
			start, _, name, ok := findOpeningTag(b, pos, "function")
			if !ok {
				break
			}
			sawSyntax = true
			bodyStart := findTagEnd(b, start)
			if bodyStart < 0 {
				break
			}
			bodyEnd, bodyOK := findCloseTag(b, bodyStart, "function")
			if !bodyOK {
				break
			}
			if name == "" {
				pos = bodyEnd
				continue
			}
			args := scanInvokeBody(b[bodyStart:bodyEnd], name)
			calls = append(calls, ToolCall{Name: name, Arguments: args})
			pos = bodyEnd
		}
	}

	// Pass 2: tool-name wrappers (e.g. <Bash>ls</Bash>,
	// <read_file>path</read_file>). We only consider this as a
	// fallback if passes 1 and 1b found nothing, otherwise a
	// model that emits both an <invoke> AND a verbal
	// "<bash>...</bash>" thought would produce a duplicate call.
	// The body is taken as the single raw argument value.
	if len(calls) == 0 {
		pos := 0
		for pos < len(b) {
			start, name, _, ok := findOpeningTag(b, pos, "")
			if !ok {
				break
			}
			if !isToolNameLike(name) || name == "invoke" || name == "parameter" ||
				name == "tool_calls" || name == "function_calls" || name == "_calls" {
				pos = start + 1
				continue
			}
			sawSyntax = true
			bodyStart := findTagEnd(b, start)
			if bodyStart < 0 {
				break
			}
			bodyEnd, bodyOK := findCloseTag(b, bodyStart, name)
			if !bodyOK {
				break
			}
			raw := strings.TrimSpace(unescapeXml(string(b[bodyStart:bodyEnd])))
			if raw != "" {
				var args map[string]interface{}
				if err := json.Unmarshal([]byte(raw), &args); err == nil {
					calls = append(calls, ToolCall{Name: name, Arguments: args})
				} else {
					key := "input"
					if k, ok := firstRequiredParam[name]; ok {
						key = k
					}
					calls = append(calls, ToolCall{Name: name, Arguments: map[string]interface{}{key: raw}})
				}
			}
			pos = bodyEnd
		}
	}

	return XmlToolCallResult{Calls: calls, SawToolCallSyntax: sawSyntax || len(calls) > 0}
}

// scanInvokeBody handles the inside of an <invoke> block. The body
// can be either:
//   - a sequence of <parameter name="x" string="true|false">VALUE</parameter>
//     children, where string="false" means VALUE is JSON, or
//   - a single direct-JSON object (sglang "Format 2").
func scanInvokeBody(body []byte, invokeName string) map[string]interface{} {
	// Try direct-JSON body first (sglang Format 2). If the body
	// starts with '{' and is a valid object, use it as the args.
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) > 0 && trimmed[0] == '{' {
		var args map[string]interface{}
		raw := string(trimmed)
		if err := json.Unmarshal(trimmed, &args); err == nil {
			return args
		}
		// Try repairing invalid backslashes (Windows paths).
		repaired := repairInvalidJSONBackslashes(raw)
		if repaired != raw {
			if err := json.Unmarshal([]byte(repaired), &args); err == nil {
				return args
			}
		}
		// Try repairing loose JSON (unquoted keys, missing brackets).
		repaired = RepairLooseJSON(raw)
		if repaired != raw {
			if err := json.Unmarshal([]byte(repaired), &args); err == nil {
				return args
			}
		}
	}
	// Fall through: parse <parameter> children.
	return scanParameters(body, invokeName)
}

// scanParameters finds every <parameter name="x">...</parameter>
// child inside an <invoke> body and returns their values as a map.
// The optional `string="true|false"` attribute on each <parameter>
// controls how the value is interpreted: string="false" (the default
// for non-string types in the DSML spec) parses the value as JSON,
// string="true" stores it as a raw string. The caller-supplied
// invokeName is only used in error comments.
func scanParameters(body []byte, invokeName string) map[string]interface{} {
	if len(body) == 0 {
		return nil
	}
	args := make(map[string]interface{})

	pos := 0
	for pos < len(body) {
		start, _, pname, ok := findOpeningTag(body, pos, "parameter")
		if !ok {
			break
		}
		if pname == "" {
			// <parameter> with no name attribute — skip.
			// Advance past the close tag to avoid an
			// infinite loop on the same tag.
			bodyEnd, bodyOK := findCloseTag(body, findTagEnd(body, start), "parameter")
			if !bodyOK {
				break
			}
			pos = bodyEnd
			continue
		}
		// Pull optional string="true|false" attribute. When false
		// the value body is JSON, so we json.Unmarshal it before
		// storing. When true (or absent, which the DSML spec
		// treats as string) we store the trimmed raw string.
		tagEnd := findTagEnd(body, start)
		if tagEnd < 0 {
			break
		}
		isStringAttr, _ := scanAttr(body, start, "string")
		// bodyStart is past the opening tag, bodyEnd is just
		// before the matching </parameter>.
		bodyStart := tagEnd
		bodyEnd, bodyOK := findCloseTag(body, tagEnd, "parameter")
		if !bodyOK {
			break
		}
		raw := string(body[bodyStart:bodyEnd])
		raw = strings.TrimSpace(unescapeXml(raw))
		_ = invokeName
		switch {
		case raw == "":
			// An empty <parameter name="x"></parameter> still
			// becomes args["x"] = "" so the executor sees the
			// key exists. The model may use empty values to
			// express "set this arg to the empty string"
			// deliberately.
			args[pname] = ""
		case isStringAttr == "false":
			// DSML marks non-string params (numbers, bools,
			// arrays, objects) with string="false" and the
			// body as JSON. Parse it so downstream tools
			// see the right Go type instead of a stringified
			// "5".
			var v interface{}
			if err := json.Unmarshal([]byte(raw), &v); err == nil {
				args[pname] = v
			} else {
				// Try repair before giving up.
				repaired := repairInvalidJSONBackslashes(raw)
				if repaired != raw {
					if err := json.Unmarshal([]byte(repaired), &v); err == nil {
						args[pname] = v
					} else {
						args[pname] = raw
					}
				} else {
					// Malformed JSON in a string="false"
					// param — fall back to the raw string so
					// the call still has something for the
					// executor to look at.
					args[pname] = raw
				}
			}
		default:
			// string="true" (or absent): store as raw
			// string. The executor's input map expects
			// strings for string-shaped arguments, so
			// this round-trips correctly.
			args[pname] = raw
		}
		pos = bodyEnd
	}
	if len(args) == 0 {
		return nil
	}
	return args
}

// findOpeningTag locates the next opening tag whose normalized name
// matches `want` (case-insensitive, DSML-prefix-stripped,
// Unicode-drift-normalized). If `want` is empty, matches ANY tag and
// returns its normalized name — used by pass 2 to discover tool-name
// wrappers like <Bash>.
//
// The returned name is always the lowercased, prefix-stripped form
// (e.g. "invoke" for both <invoke> and <｜DSML｜INVOKE>). If the
// tag has a `name="..."` attribute, the unescaped attribute value
// is returned as attrName — used by callers to learn the tool name
// from <invoke name="bash"> and the arg name from
// <parameter name="command">.
func findOpeningTag(body []byte, start int, want string) (int, string, string, bool) {
	pos := start
	for pos < len(body) {
		lt := bytes.IndexByte(body[pos:], '<')
		if lt < 0 {
			return 0, "", "", false
		}
		lt += pos
		// Skip past the '<' and any leading non-name characters
		// (handles '<br/>' and '<!-- ... -->' corner cases —
		// though we don't try to handle the latter).
		p := lt + 1
		if p < len(body) && (body[p] == '/' || body[p] == '!' || body[p] == '?') {
			pos = lt + 1
			continue
		}
		// Read the tag name up to the first delimiter.
		nameStart := p
		for p < len(body) {
			c := body[p]
			if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '>' || c == '/' {
				break
			}
			p++
		}
		if p == nameStart {
			pos = lt + 1
			continue
		}
		rawName := string(body[nameStart:p])
		name := normalizeTagName(rawName)
		if want != "" && name != want {
			pos = lt + 1
			continue
		}
		// Optional name="..." attribute. Done as a second pass so
		// the matching loop stays tight; tags without a name
		// attribute (e.g. <invoke>...</invoke> with the name on a
		// sibling) return an empty attrName.
		attrName, _ := scanAttr(body, lt, "name")
		return lt, name, attrName, true
	}
	return 0, "", "", false
}

// findCloseTag locates the matching close tag for an opening tag
// that started at `bodyStart` (just past its '>'). The close tag's
// name is matched case-insensitively and DSML-stripped the same way
// as opening tags. Returns the position of the '<' of the close tag
// (not past the '>') so the caller can do body[bodyStart:bodyEnd]
// to get the inner content. Callers that want to advance past the
// close tag should use findTagEnd on the returned position.
func findCloseTag(body []byte, bodyStart int, name string) (int, bool) {
	// Search for "</NAME" (any closing-tag opener) and verify the
	// normalized name matches.
	needle := []byte("</")
	pos := bodyStart
	for pos < len(body) {
		idx := bytes.Index(body[pos:], needle)
		if idx < 0 {
			return 0, false
		}
		idx += pos
		p := idx + 2
		// Read the close tag's name.
		nameStart := p
		for p < len(body) {
			c := body[p]
			if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '>' {
				break
			}
			p++
		}
		rawName := string(body[nameStart:p])
		closeName := normalizeTagName(rawName)
		if closeName == name {
			return idx, true
		}
		pos = idx + 1
	}
	return 0, false
}

// normalizeTagName lowercases the name, strips any DSML prefix, and
// handles Unicode-drifted variants.
//
// The DeepSeek DSML spec wraps the tag name with `｜DSML｜` (the
// fullwidth vertical bar, 3 bytes in UTF-8), `|DSML|` (halfwidth),
// or `DSML|` (no opening bar, model skipped the first bar). All
// three are normalised to the bare name so a single match works
// against any of them. The prefix is always at the START of the
// tag name, not the end.
//
// Additional drift variants handled:
//   - Collapsed: `DSMLtool_calls`, `DSMLinvoke`, `DSMLparameter`
//   - Vendor prefix: `abc|tool_calls`, `vendor_tool_calls`, `agent - tool_calls`
//   - CJK/Fullwidth: `〈invoke〉`, `＜invoke＞`, `！DSML！invoke`
//   - Ideographic comma: `、DSML、invoke`
func normalizeTagName(name string) string {
	n := strings.ToLower(name)
	// Strip zero-width and BOM characters that some models inject.
	n = stripIgnorables(n)
	// Fullwidth ASCII range ＀-～ maps to ASCII by subtracting 0xFEE0.
	n = normalizeFullwidthASCII(n)
	// Map CJK angle brackets 〈〉 to < > (already handled by fullwidth
	// normalization, but keep explicit for clarity).
	n = strings.ReplaceAll(n, "〈", "<")
	n = strings.ReplaceAll(n, "〉", ">")

	// Try separator-based DSML prefixes first.
	for _, prefix := range []string{"｜dsml｜", "|dsml|", "、dsml、", "！dsml！", "dsml|", "dsml-", "dsml_", "dsml ", "dsml"} {
		if strings.HasPrefix(n, prefix) && len(n) > len(prefix) {
			return n[len(prefix):]
		}
	}

	// Try collapsed DSML prefix (no separator): dsmltool_calls,
	// dsmlinvoke, dsmlparameter. Reject lookalikes like
	// dsmltool_calls_extra (the extra suffix would not match a
	// canonical name).
	if strings.HasPrefix(n, "dsml") && len(n) > 4 {
		rest := n[4:]
		if rest == "tool_calls" || rest == "invoke" || rest == "parameter" {
			return rest
		}
	}

	// Try arbitrary vendor prefix: <anything><sep>tool_calls
	// where <sep> is |, _, -, or space. The canonical name must be
	// one of the known tag names.
	if canonical := stripVendorPrefix(n); canonical != "" {
		return canonical
	}

	return n
}

// stripIgnorables removes zero-width characters, BOMs, and other
// format characters that models sometimes inject.
func stripIgnorables(s string) string {
	if !strings.ContainsAny(s, ignorableRunes) {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case '\u200B', '\u200C', '\u200D', '\uFEFF',
			'\u200E', '\u200F',
			'\u202A', '\u202B', '\u202C', '\u202D', '\u202E',
			'\u2060':
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

const ignorableRunes = "\u200B\u200C\u200D\uFEFF\u200E\u200F\u202A\u202B\u202C\u202D\u202E\u2060"

// normalizeFullwidthASCII maps fullwidth ASCII characters to their
// halfwidth equivalents. The fullwidth range is ＀ (U+FF00) to
//  ～ (U+FF5E), offset 0xFEE0 from ASCII.
func normalizeFullwidthASCII(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	changed := false
	for _, r := range s {
		if r >= '！' && r <= '～' {
			b.WriteRune(r - 0xFEE0)
			changed = true
		} else {
			b.WriteRune(r)
		}
	}
	if !changed {
		return s
	}
	return b.String()
}

// stripVendorPrefix removes an arbitrary vendor prefix from a tag
// name. The prefix must end with a separator (|, _, -, space) and
// the remainder must be a canonical tag name. Returns "" if no
// vendor prefix is found.
func stripVendorPrefix(n string) string {
	// Try each separator.
	for _, sep := range []string{"|", "_", "-", " "} {
		idx := strings.LastIndex(n, sep)
		if idx < 1 || idx >= len(n)-1 {
			continue
		}
		candidate := n[idx+1:]
		if candidate == "tool_calls" || candidate == "invoke" || candidate == "parameter" {
			// Reject lookalikes: if the prefix contains "dsml"
			// it should have been handled by the DSML prefix
			// logic above. If it doesn't contain dsml, it's a
			// vendor prefix.
			prefix := n[:idx]
			if strings.Contains(prefix, "dsml") || strings.Contains(prefix, "tool") {
				continue
			}
			return candidate
		}
	}
	return ""
}

// isToolNameLike returns true for strings that look like plausible
// tool names: lowercase, alphanumeric or underscores, not one of
// the XML/DSML control tags. Used by pass 2 to filter the universe
// of opening tags down to candidates that *could* be tool wrappers.
func isToolNameLike(name string) bool {
	if name == "" {
		return false
	}
	for _, c := range name {
		switch {
		case c >= 'a' && c <= 'z':
		case c >= '0' && c <= '9':
		case c == '_' || c == '-':
		default:
			return false
		}
	}
	return true
}

// scanAttr finds `attrName="value"` inside the tag starting at
// tagStart (which points at '<'). Returns the value and ok. We only
// handle double-quoted attribute values — that's all the model emits
// — and decode the same five XML entities as the value decoder so
// an attribute like name="read&quot;_file" resolves to `read"_file`.
func scanAttr(text []byte, tagStart int, attrName string) (string, bool) {
	needle := []byte(attrName + `="`)
	rel := bytes.Index(text[tagStart:], needle)
	if rel < 0 {
		return "", false
	}
	valStart := tagStart + rel + len(needle)
	valEnd := bytes.IndexByte(text[valStart:], '"')
	if valEnd < 0 {
		return "", false
	}
	val := unescapeXml(string(text[valStart : valStart+valEnd]))
	return val, true
}

// findTagEnd returns the byte position immediately after the closing
// '>' of a tag starting at `tagStart` (which points at '<'). Honours
// quoted attribute values so a '>' inside `name="3>2"` does not
// terminate the tag prematurely. Returns -1 if the closing '>' is
// missing (truncated input).
func findTagEnd(text []byte, tagStart int) int {
	if tagStart >= len(text) || text[tagStart] != '<' {
		return -1
	}
	inQuote := byte(0)
	for i := tagStart; i < len(text); i++ {
		c := text[i]
		if inQuote != 0 {
			if c == inQuote {
				inQuote = 0
			}
			continue
		}
		switch c {
		case '"', '\'':
			inQuote = c
		case '>':
			return i + 1
		}
	}
	return -1
}

// unescapeXml decodes the five XML predefined entities. The output
// is only escaped if the input contains '&', which keeps the common
// (no-entity) case a single-byte check with no allocation.
func unescapeXml(s string) string {
	if !strings.ContainsAny(s, "&") {
		return s
	}
	// Order matters: do &amp; LAST so we don't double-decode
	// &amp;lt; into '<' when the model actually meant the literal
	// string "&lt;".
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", `"`)
	s = strings.ReplaceAll(s, "&apos;", "'")
	s = strings.ReplaceAll(s, "&amp;", "&")
	return s
}

// ─────────────────────────────────────────────────────────────────
// Pre-processing pipeline helpers (added from ds2api parser features)
// ─────────────────────────────────────────────────────────────────

// stripFencedCodeBlocks removes content inside ``` and ~~~ markdown
// fenced code blocks. Tool call syntax inside code examples is
// documentation, not executable, and should not be parsed.
//
// The function preserves content before an unclosed fence (so a
// truncated model response doesn't lose everything after the fence
// opener). It also handles CDATA blocks inside fences correctly by
// not treating their content as fence markers.
func stripFencedCodeBlocks(text string) string {
	if text == "" {
		return ""
	}
	if !strings.Contains(text, "```") && !strings.Contains(text, "~~~") {
		return text
	}
	var b strings.Builder
	b.Grow(len(text))

	inFence := false
	fenceMarker := ""
	beforeFenceLen := 0

	lines := strings.SplitAfter(text, "\n")
	for _, line := range lines {
		if !inFence {
			trimmed := strings.TrimLeft(line, " \t")
			if marker, ok := parseFenceOpen(trimmed); ok {
				inFence = true
				fenceMarker = marker
				beforeFenceLen = b.Len()
				continue
			}
			b.WriteString(line)
			continue
		}
		// Inside a fence: check for closing fence.
		trimmed := strings.TrimLeft(line, " \t")
		if isFenceClose(trimmed, fenceMarker) {
			inFence = false
			fenceMarker = ""
		}
	}

	if inFence {
		// Unclosed fence: preserve content before the fence started.
		result := b.String()
		if beforeFenceLen > 0 && beforeFenceLen <= len(result) {
			return result[:beforeFenceLen]
		}
		return ""
	}
	return b.String()
}

// parseFenceOpen checks if a line opens a markdown code fence.
// Returns the fence marker (e.g. "```" or "~~~") and ok.
func parseFenceOpen(line string) (string, bool) {
	if len(line) < 3 {
		return "", false
	}
	ch := line[0]
	if ch != '`' && ch != '~' {
		return "", false
	}
	count := 0
	for count < len(line) && line[count] == ch {
		count++
	}
	if count < 3 {
		return "", false
	}
	return strings.Repeat(string(ch), count), true
}

// isFenceClose checks if a line closes a markdown code fence.
func isFenceClose(line, marker string) bool {
	if marker == "" {
		return false
	}
	ch := marker[0]
	if line == "" || line[0] != ch {
		return false
	}
	count := 0
	for count < len(line) && line[count] == ch {
		count++
	}
	if count < len(marker) {
		return false
	}
	rest := strings.TrimSpace(line[count:])
	return rest == ""
}

// sanitizeLooseCDATA repairs malformed CDATA blocks. If a
// <![CDATA[ opening is found without a matching ]]> close, the
// opening is stripped so the remaining content can still be parsed.
// Properly closed CDATA blocks are left untouched.
func sanitizeLooseCDATA(text string) string {
	if text == "" {
		return ""
	}
	if !strings.Contains(text, "<![CDATA[") && !strings.Contains(text, "<![cdata[") {
		return text
	}
	var b strings.Builder
	b.Grow(len(text))
	changed := false
	pos := 0
	for pos < len(text) {
		// Find next CDATA open (case-insensitive).
		openStart := findLooseCDATAOpen(text, pos)
		if openStart < 0 {
			b.WriteString(text[pos:])
			break
		}
		// Write everything before the CDATA open.
		b.WriteString(text[pos:openStart])
		// Check if this CDATA is properly closed.
		closeStart := strings.Index(text[openStart:], "]]>")
		if closeStart >= 0 {
			// Properly closed: keep as-is.
			end := openStart + closeStart + 3
			b.WriteString(text[openStart:end])
			pos = end
			continue
		}
		// Unterminated: strip the opening tag and keep the content.
		// The content after <![CDATA[ is preserved.
		openLen := looseCDATAOpenLen(text, openStart)
		b.WriteString(text[openStart+openLen:])
		pos = len(text)
		changed = true
	}
	if !changed {
		return text
	}
	return b.String()
}

// findLooseCDATAOpen finds the next <![CDATA[ opening (case-insensitive)
// at or after `start`. Returns -1 if not found.
func findLooseCDATAOpen(text string, start int) int {
	lower := strings.ToLower(text)
	idx := strings.Index(lower[start:], "<![cdata[")
	if idx < 0 {
		return -1
	}
	return start + idx
}

// looseCDATAOpenLen returns the byte length of the CDATA opening tag
// at `idx` in `text` (always 9 for "<![CDATA[").
func looseCDATAOpenLen(text string, idx int) int {
	_ = text
	return len("<![CDATA[")
}

// ─────────────────────────────────────────────────────────────────
// JSON repair helpers (added from ds2api parser features)
// ─────────────────────────────────────────────────────────────────

// repairInvalidJSONBackslashes fixes invalid backslash sequences in
// JSON strings. The most common case is Windows paths like
// `C:\Users\name` which need to become `C:\\Users\name` for valid
// JSON. Also fixes invalid unicode escapes like `\u123` (3 digits
// instead of 4) by escaping the backslash.
func repairInvalidJSONBackslashes(s string) string {
	if !strings.Contains(s, "\\") {
		return s
	}
	var out strings.Builder
	out.Grow(len(s) + 10)
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		if runes[i] == '\\' {
			if i+1 < len(runes) {
				next := runes[i+1]
				switch next {
				case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
					out.WriteRune('\\')
					out.WriteRune(next)
					i++
					continue
				case 'u':
					// Check for valid 4-digit hex escape.
					if i+5 < len(runes) {
						isHex := true
						for j := 1; j <= 4; j++ {
							r := runes[i+1+j]
							if (r < '0' || r > '9') && (r < 'a' || r > 'f') && (r < 'A' || r > 'F') {
								isHex = false
								break
							}
						}
						if isHex {
							out.WriteRune('\\')
							out.WriteRune('u')
							for j := 1; j <= 4; j++ {
								out.WriteRune(runes[i+1+j])
							}
							i += 5
							continue
						}
					}
				}
			}
			// Not a valid escape: double the backslash.
			out.WriteString("\\\\")
		} else {
			out.WriteRune(runes[i])
		}
	}
	return out.String()
}

var unquotedKeyPattern = regexp.MustCompile(`([{,]\s*)([a-zA-Z_][a-zA-Z0-9_]*)\s*:`)

// missingArrayBracketsPattern identifies a sequence of two or more
// JSON objects separated by commas that immediately follow a colon,
// which indicates a missing array bracket `[` `]`. E.g.,
// `"key": {"a": 1}, {"b": 2}` → `"key": [{"a": 1}, {"b": 2}]`.
var missingArrayBracketsPattern = regexp.MustCompile(`(:\s*)(\{(?:[^{}]|\{[^{}]*\})*\}(?:\s*,\s*\{(?:[^{}]|\{[^{}]*\})*\})+)`)

// RepairLooseJSON repairs common JSON formatting issues:
//  1. Single quotes → double quotes: `{'k':'v'}` → `{"k":"v"}`
//  2. Unquoted keys: `{key: val}` → `{"key": val}`
//  3. Trailing commas: `{a:1,}` → `{a:1}`
//  4. Python booleans: `True`/`False`/`None` → `true`/`false`/`null`
//  5. Numeric drift: `NaN`/`Infinity`/`-Infinity` → `null` (JSON has no
//     literal for these; null is the safest interpretation)
//  6. Missing array brackets: `[{a: 1}, {b: 2}]` → `[{a: 1}, {b: 2}]`
//     (the objects after a colon separated by commas should be an array)
func RepairLooseJSON(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	s = singleQuoteToDouble(s)
	s = unquotedKeyPattern.ReplaceAllString(s, `$1"$2":`)
	s = trailingCommaPattern.ReplaceAllString(s, "")
	s = pythonBoolPattern.ReplaceAllStringFunc(s, pyBoolReplacer)
	s = missingArrayBracketsPattern.ReplaceAllString(s, `$1[$2]`)
	return s
}

var trailingCommaPattern = regexp.MustCompile(`,\s*([}\]])`)

var pythonBoolPattern = regexp.MustCompile(`\b(True|False|None|NaN|Infinity|-Infinity)\b`)

func pyBoolReplacer(m string) string {
	switch m {
	case "True":
		return "true"
	case "False":
		return "false"
	case "None":
		return "null"
	case "NaN", "Infinity", "-Infinity":
		return "null"
	}
	return m
}

// singleQuoteToDouble converts single-quoted strings to double-quoted
// ones. It deliberately walks the string respecting backslash escapes so
// an apostrophe inside a double-quoted string isn't treated as a quote
// boundary.
func singleQuoteToDouble(s string) string {
	if !strings.ContainsRune(s, '\'') {
		return s
	}
	var b strings.Builder
	b.Grow(len(s) + 8)
	inSingle := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '\\' && i+1 < len(s) {
			b.WriteByte(c)
			b.WriteByte(s[i+1])
			i++
			continue
		}
		if c == '\'' {
			b.WriteByte('"')
			inSingle = !inSingle
			continue
		}
		if c == '"' && inSingle {
			b.WriteByte('\\')
		}
		b.WriteByte(c)
	}
	return b.String()
}

// ─────────────────────────────────────────────────────────────────
// Unicode normalization helpers (added from ds2api parser features)
// ─────────────────────────────────────────────────────────────────

// normalizeRune normalizes a single rune to its ASCII equivalent
// for tag scanning purposes. Handles fullwidth ASCII, CJK angle
// brackets, curly quotes, and other common drift variants.
func normalizeRune(r rune) rune {
	switch r {
	case '〈', '＜', '﹤':
		return '<'
	case '〉', '＞', '﹥':
		return '>'
	case '＝', '﹦', '꞊':
		return '='
	case '／', '∕', '⁄', '⧸':
		return '/'
	case '“', '”', '＂':
		return '"'
	case '‘', '’', '＇':
		return '\''
	}
	if r >= '！' && r <= '～' {
		return r - 0xFEE0
	}
	return r
}

// normalizedASCIIAt returns the ASCII-normalized byte at idx and
// the byte size of the rune. Returns (0, 0) if out of bounds.
func normalizedASCIIAt(text string, idx int) (byte, int) {
	if idx < 0 || idx >= len(text) {
		return 0, 0
	}
	r, size := utf8.DecodeRuneInString(text[idx:])
	if r == utf8.RuneError && size == 0 {
		return 0, 0
	}
	normalized := normalizeRune(r)
	if normalized > 0x7f {
		return 0, 0
	}
	return byte(normalized), size
}

// asciiLower lowercases an ASCII byte.
func asciiLower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + ('a' - 'A')
	}
	return b
}

// ─────────────────────────────────────────────────────────────────
// Markup normalization (added from ds2api parser features)
// ─────────────────────────────────────────────────────────────────

// normalizeDSMLMarkup rewrites Unicode-drifted DSML markup to its
// canonical ASCII form. This allows the main scanner to work with
// a single set of tag names without handling every drift variant.
//
// Handles:
//   - Fullwidth tags: ＜｜DSML｜tool_calls＞ → <|DSML|tool_calls>
//   - CJK angle: 〈｜DSML｜invoke〉 → <|DSML|invoke>
//   - Fullwidth separators: ！DSML！invoke → <|DSML|invoke>
//   - Mixed DSML/canonical: normalizes all to canonical
func normalizeDSMLMarkup(text string) string {
	if text == "" {
		return text
	}
	// Quick check: if the text doesn't contain any DSML-like markers
	// or Unicode drift characters, return as-is.
	if !hasDSMLOrDrift(text) {
		return text
	}
	var b strings.Builder
	b.Grow(len(text))
	for _, r := range text {
		normalized := normalizeRune(r)
		if normalized != r && normalized > 0 && normalized <= 0x7f {
			b.WriteRune(normalized)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// hasDSMLOrDrift checks if the text contains DSML markers or
// Unicode drift characters that need normalization.
func hasDSMLOrDrift(text string) bool {
	if strings.Contains(text, "DSML") || strings.Contains(text, "dsml") {
		return true
	}
	if strings.ContainsAny(text, "〈〉＜＞＝／") {
		return true
	}
	// Check for fullwidth ASCII range.
	for _, r := range text {
		if r >= '！' && r <= '～' {
			return true
		}
	}
	return false
}

// ─────────────────────────────────────────────────────────────────
// CDATA and XML fragment parsing (added from ds2api parser features)
// ─────────────────────────────────────────────────────────────────
//
// These helpers are used for parsing CDATA content that contains
// XML fragments (e.g., when a model puts a nested structure inside
// a CDATA section). They are only called when CDATA content is
// detected and looks like structured data.
//
// Currently not invoked from the main parser pipeline — they are
// available for callers that need to parse CDATA content separately.
// The main pipeline treats CDATA content as raw text.

// parseCDATAFragment parses CDATA content that contains an XML
// fragment into a map[string]any. Repeated child elements become
// arrays. This is used for structured CDATA like:
//
//	<![CDATA[<item><name>x</name></item>]]>
//
// which parses to `map[string]any{"item": map[string]any{"name": "x"}}`.
func parseCDATAFragment(cdata string) (any, bool) {
	cdata = strings.TrimSpace(cdata)
	if cdata == "" {
		return nil, false
	}
	if !strings.HasPrefix(cdata, "<") {
		return nil, false
	}
	// Unescape HTML entities first.
	cdata = html.UnescapeString(cdata)
	// Use encoding/xml to parse the fragment.
	return parseXMLFragmentValue(cdata)
}

// parseXMLFragmentValue parses a single XML element (or document fragment)
// into a map[string]any. Repeated sibling elements are folded into a slice.
// Attributes are stored under "@name". Text content is stored under "#text".
// Self-closing or empty elements map to an empty map.
func parseXMLFragmentValue(s string) (any, bool) {
	dec := xml.NewDecoder(strings.NewReader(s))
	dec.Strict = false
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, false
		}
		if tok == nil {
			return nil, false
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		return decodeXMLNode(dec, se), true
	}
}

func decodeXMLNode(dec *xml.Decoder, start xml.StartElement) any {
	node := make(map[string]any)
	for _, attr := range start.Attr {
		node["@"+attr.Name.Local] = attr.Value
	}
	var text strings.Builder
	for {
		tok, err := dec.Token()
		if err != nil {
			return node
		}
		if tok == nil {
			return node
		}
		switch t := tok.(type) {
		case xml.StartElement:
			child := decodeXMLNode(dec, t)
			if existing, ok := node[t.Name.Local]; ok {
				switch e := existing.(type) {
				case []any:
					node[t.Name.Local] = append(e, child)
				default:
					node[t.Name.Local] = []any{e, child}
				}
			} else {
				node[t.Name.Local] = child
			}
		case xml.CharData:
			text.Write(t)
		case xml.EndElement:
			if t.Name == start.Name {
				trimmed := strings.TrimSpace(text.String())
				if trimmed != "" {
					node["#text"] = trimmed
				}
				return node
			}
		}
	}
}
