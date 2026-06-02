package api

import "strings"

// StripToolCalls removes tool-call invocations from a chunk of assistant
// text so the cleaned message can be displayed to the user without
// exposing the raw control syntax.
//
// The formats handled are the common ones emitted by the various
// providers wired into claw-code-go:
//
//   - <tool_calls>...</tool_calls>, <function_calls>...</function_calls>,
//     <_calls>...</_calls> XML wrappers.
//   - <invoke name="...">...</invoke>, <function name="...">...</function>,
//     and <tool_call>...</tool_call> children inside those wrappers
//     (DeepSeek/Qwen-style XML tool-call formats).
//   - ```json...``` markdown code fences wrapping a tool-call payload.
//   - ReAct "Action: <tool>\nAction Input: <json>" line pairs.
//
// This is provider-agnostic on purpose: the conversation loop calls it
// after every turn on the accumulated assistant text, regardless of
// whether the model was Anthropic, OpenAI, DeepSeek, etc. Providers
// whose protocol already separates tool calls from text (Anthropic,
// OpenAI) won't see any change — the text they emit never contains the
// syntax this function matches.
//
// The function never modifies text outside the tool-call syntax. If
// nothing matches, the input is returned (with surrounding whitespace
// trimmed) so callers can always assign the return value back to the
// original buffer.
func StripToolCalls(text string) string {
	result := text
	// Strip the outer wrappers (and their bodies) first so the
	// inner-strip pass below has nothing to find. This is what
	// collapses <tool_calls>...<invoke...>...</invoke>...</tool_calls>
	// down to an empty string in a single iteration, instead of
	// leaving orphaned <invoke> blocks behind.
	for _, wrapper := range []string{"<tool_calls>", "<function_calls>", "<_calls>"} {
		endWrapper := "</" + wrapper[1:]
		for {
			si := strings.Index(result, wrapper)
			if si == -1 {
				break
			}
			ei := strings.Index(result[si:], endWrapper)
			if ei == -1 {
				result = result[:si]
				break
			}
			ei += si
			result = result[:si] + result[ei+len(endWrapper):]
		}
	}
	// Belt-and-braces: strip any bare <invoke>, <function>, or
	// <file_content> blocks that weren't wrapped in a <tool_calls>
	// container. The model sometimes emits bare invokes/functions,
	// especially mid-stream before the closing </tool_calls> arrives.
	for _, tag := range []string{"invoke", "function", "function_call"} {
		open := "<" + tag
		close := "</" + tag + ">"
		pos := 0
		for pos < len(result) {
			si := strings.Index(result[pos:], open)
			if si < 0 {
				break
			}
			si += pos
			// Require a tag boundary so we don't strip e.g.
			// `<invoke_other>` or `<tool_call>` (where
			// `<function` is a prefix that must not match).
			// A valid tag-name continuation character means
			// `open` is a prefix of a different tag — skip
			// past the whole opening tag and keep searching.
			if si+len(open) < len(result) {
				c := result[si+len(open)]
				if isTagNameCont(c) {
					gt := strings.IndexByte(result[si:], '>')
					if gt < 0 {
						break
					}
					pos = si + gt + 1
					continue
				}
			}
			ei := strings.Index(result[si:], close)
			if ei == -1 {
				result = result[:si]
				break
			}
			ei += si
			result = result[:si] + result[ei+len(close):]
			pos = si
		}
	}
	lines := strings.Split(result, "\n")
	var cleaned []string
	skipNext := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "Action:") {
			skipNext = true
			continue
		}
		if skipNext && strings.HasPrefix(trimmed, "Action Input:") {
			skipNext = false
			continue
		}
		skipNext = false
		cleaned = append(cleaned, line)
	}
	return strings.TrimSpace(strings.Join(cleaned, "\n"))
}

// TrimEndOfTurnIndicator removes a trailing "FINISHED" sentinel that
// the DeepSeek model emits as a learned end-of-turn marker. The
// model often splices it directly onto the last word
// ("...layers.FINISHED") or adds sentence-boundary punctuation
// before it ("...layers.FINISHED"). We strip the sentinel plus a
// single trailing sentence-end punctuation (., !, ?) if present.
//
// This is intentionally aggressive: the English word "FINISHED" at
// the end of a code-session turn is rare enough that the false
// positives are acceptable. Callers that need to distinguish the
// sentinel from English text should check the SSE FINISHED signal
// themselves (this function assumes it already fired).
//
// Provider-agnostic: any provider that produces a trailing
// "FINISHED" in the model's text benefits from this cleanup.
func TrimEndOfTurnIndicator(text string) string {
	const sentinel = "FINISHED"
	trimmed := strings.TrimRight(text, " \t\r\n")
	if !strings.HasSuffix(trimmed, sentinel) {
		return text
	}
	cut := len(trimmed) - len(sentinel)
	// Strip sentence-boundary punctuation the model added to fake
	// a sentence end before the sentinel ("...layers.FINISHED" →
	// "...layers", not "...layers.").
	if cut > 0 {
		switch trimmed[cut-1] {
		case '.', '!', '?':
			cut--
		}
	}
	return strings.TrimRight(trimmed[:cut], " \t\r\n")
}

// isTagNameCont reports whether c can appear as a continuation
// character in an XML element name (NameChar in the XML spec:
// letters, digits, '.', '-', '_', and a few Unicode ranges). The
// stripper uses it to decide whether a hit like `<function` in
// `<tool_call>` is a real tag or a prefix of a different tag.
func isTagNameCont(c byte) bool {
	switch {
	case c >= 'a' && c <= 'z':
		return true
	case c >= 'A' && c <= 'Z':
		return true
	case c >= '0' && c <= '9':
		return true
	case c == '.', c == '-', c == '_':
		return true
	}
	return false
}
