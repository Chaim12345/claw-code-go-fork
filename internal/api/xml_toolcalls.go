package api

import (
	"bytes"
	"encoding/json"
	"strings"
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
// We deliberately do NOT use encoding/xml. The model output has
// CDATA-ish noise (unescaped &, bare < and >, comments, DSML
// multibyte prefixes), self-closing tag variants, and other patterns
// that would make a strict parser fail on the majority of real
// responses. Instead, this is a single-pass byte-level scanner.
//
// Tag matching is case-insensitive and DSML-prefix-tolerant. The
// close-tag matcher uses the same normalization so a tag opened as
// <｜DSML｜Invoke> closes on </｜DSML｜invoke>. The caller is
// responsible for filtering Name against the per-provider knownTools
// list — this parser is intentionally permissive so a future provider
// (or a new tool name) doesn't require a code change here.
//
// The hot path uses bytes.Index / bytes.IndexByte rather than
// strings.Index so the body never has to be re-encoded — on a 4 KB
// tool block this is roughly 2-3x faster than the equivalent
// strings.Index loop.
//
// Returns nil if no tool calls are found.
func ExtractXmlToolCalls(text string) []ToolCall {
	if len(text) == 0 {
		return nil
	}
	b := []byte(text)

	var calls []ToolCall

	// Pass 1: <invoke ...> blocks (the documented DeepSeek/Qwen
	// format, both with and without the DSML prefix and with both
	// XML-parameter and direct-JSON body variants).
	pos := 0
	for pos < len(b) {
		start, _, name, ok := findOpeningTag(b, pos, "invoke")
		if !ok {
			break
		}
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

	if len(calls) == 0 {
		return nil
	}
	return calls
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
		if err := json.Unmarshal(trimmed, &args); err == nil {
			return args
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
				// Malformed JSON in a string="false"
				// param — fall back to the raw string so
				// the call still has something for the
				// executor to look at.
				args[pname] = raw
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
// matches `want` (case-insensitive, DSML-prefix-stripped). If `want`
// is empty, matches ANY tag and returns its normalized name — used
// by pass 2 to discover tool-name wrappers like <Bash>.
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

// normalizeTagName lowercases the name and strips any DSML prefix.
// The DeepSeek DSML spec wraps the tag name with `｜DSML｜` (the
// fullwidth vertical bar, 3 bytes in UTF-8), `|DSML|` (halfwidth),
// or `DSML|` (no opening bar, model skipped the first bar). All
// three are normalised to the bare name so a single match works
// against any of them. The prefix is always at the START of the
// tag name, not the end.
func normalizeTagName(name string) string {
	n := strings.ToLower(name)
	for _, prefix := range []string{"｜dsml｜", "|dsml|", "dsml|", "dsml"} {
		if strings.HasPrefix(n, prefix) && len(n) > len(prefix) {
			return n[len(prefix):]
		}
	}
	return n
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
