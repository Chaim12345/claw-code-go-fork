package deepseek

import (
	"encoding/json"
	"strings"

	"claw-code-go/internal/api"

	"github.com/m-mizutani/jsonex"
)

// knownTools is the set of tool names the conversation loop can execute.
// This list MUST match the case labels in ConversationLoop.ExecuteTool
// and ExecuteToolQuiet (internal/runtime/conversation.go). The runtime
// falls through to "unknown tool" for any name that isn't in both
// places, so a mismatch here is a silent dead-letter path — see
// knownTools_must_match_executor in provider_test.go for the pinned
// consistency test.
var knownTools = map[string]bool{
	"bash":       true,
	"read_file":  true,
	"write_file": true,
	"file_edit":  true,
	"grep":       true,
	"glob":       true,
	"web_fetch":  true,
	"web_search": true,
	"ask_user":   true,
	"todo_write": true,
}

// ToolCall is an alias for api.ToolCall. Kept here so the existing
// test surface (and any other in-package callers) doesn't need to be
// rewritten — every parser in this file produces this canonical
// shape, which is what the provider layer in provider.go marshals
// into api.ContentBlock.ToolUse for the conversation loop.
type ToolCall = api.ToolCall

// ExtractToolCalls scans a chunk of model output and returns any tool calls.
// The model uses a few different shapes — JSON, XML, code-fenced, etc. —
// so we try each in order. Returns nil if nothing was found.
//
// JSON and XML are tried as peers (the model picks one or the other
// depending on its system prompt; the DeepSeek web model defaults to
// XML). JSON wins ties because it can carry richer argument shapes
// (nested objects, arrays) that the XML parameter format would have
// to flatten to strings.
//
// If the model emits a flood of <tool_calls> tags (common with
// reasoning/thinking modes), we collapse the noise down to one copy of
// the first proper tool-call block first. Without this, the XML
// extractor can get stuck scanning tens of empty <tool_calls></tool_calls>
// echoes.
func ExtractToolCalls(text string) []api.ToolCall {
	if len(text) == 0 {
		return nil
	}
	
	// Fast path: check for tool call indicators before expensive parsing
	hasToolIndicators := strings.Contains(text, "tool_calls") ||
		strings.Contains(text, "function_call") ||
		strings.Contains(text, "Action:") ||
		strings.Contains(text, "<tool_") ||
		strings.Contains(text, "{\"tool")
	
	if !hasToolIndicators {
		return nil
	}
	
	// Collapse excessive tool_calls tags before parsing
	if strings.Count(text, "<tool_calls>") > 50 {
		text = collapseExcessiveToolCalls(text)
	}
	
	// Try XML first (preferred for DeepSeek, better streaming compatibility)
	if calls := extractXmlToolCalls(text); len(calls) > 0 {
		return calls
	}
	// Fallback to JSON
	if calls := extractJsonToolCalls(text); len(calls) > 0 {
		return calls
	}
	if calls := extractCodeBlockToolCalls(text); len(calls) > 0 {
		return calls
	}
	if calls := extractReactToolCalls(text); len(calls) > 0 {
		return calls
	}
	if calls := extractFunctionCallToolCalls(text); len(calls) > 0 {
		return calls
	}
	return nil
}

// extractJsonToolCalls walks valid JSON objects in `text` using jsonex
// (RFC 8259 compliant, ignores surrounding prose / markdown noise) and
// turns any tool-call shapes it finds into []ToolCall.
//
// We run two passes and dedupe the results. The two passes are
// complementary because jsonex's two APIs have different failure modes:
//
//  1. jsonex.New(...).Decode — streaming decoder. Returns EVERY valid
//     object in the text, so it correctly handles the "model emits two
//     tool-call blocks separated by reasoning" case. BUT it bails on
//     the first `{` it sees, so a leading fragment like `{not valid}`
//     (common when the model is mid-thought) prevents it from ever
//     reaching the real tool call that follows.
//
//  2. jsonex.Unmarshal — extracts the LONGEST valid JSON object. It
//     handles the `{not valid}` noise case (it can skip past invalid
//     fragments), but only returns ONE object. So if the model emits
//     multiple tool-call blocks of equal length, we only see one.
//
// Running both and deduping gets us the union: multi-block coverage
// from the streaming pass, plus robustness to leading noise from the
// Unmarshal pass. This is the same "strict first, repair on failure"
// strategy used by json_repair (Python) and parsePartialJson in
// ax-llm/ax, just generalised across jsonex's two APIs.
func extractJsonToolCalls(text string) []api.ToolCall {
	if len(text) == 0 {
		return nil
	}
	seen := make(map[string]bool)
	var calls []ToolCall
		add := func(callsToAdd []api.ToolCall) {
		for _, c := range callsToAdd {
			key := toolCallKey(c)
			if seen[key] {
				continue
			}
			seen[key] = true
			calls = append(calls, c)
		}
	}

	// Pass 1: streaming decoder for multi-block coverage.
	dec := jsonex.New(strings.NewReader(text))
	for {
		var obj map[string]interface{}
		err := dec.Decode(&obj)
		if err != nil {
			break
		}
		add(callsFromObject(obj))
	}

	// Pass 2: Unmarshal, which recovers from a leading invalid
	// `{...}` fragment that the streaming decoder choked on. It also
	// catches the case where the streaming decoder found nothing
	// (e.g. entirely noise-prefixed text).
	var obj map[string]interface{}
	if err := jsonex.Unmarshal([]byte(text), &obj); err == nil {
		add(callsFromObject(obj))
	}
	return calls
}

// toolCallKey produces a stable, comparable identity for a tool call
// so we can dedupe across the two passes. JSON-marshaling the args
// (sorted) gives us a canonical key that ignores map iteration order.
func toolCallKey(c api.ToolCall) string {
	if c.Arguments == nil {
		return c.Name + "|"
	}
	b, _ := json.Marshal(c.Arguments)
	return c.Name + "|" + string(b)
}

// callsFromObject extracts all tool calls found in a single decoded JSON
// object. Handles three shapes:
//
//  1. Wrapped tool_calls array:     {"tool_calls":[{"name":..,"arguments":..}]}
//  2. OpenAI-style with "function": {"tool_calls":[{"function":{"name":..,"arguments":..}}]}
//  3. Single-tool object:          {"tool":"bash","command":"ls"}
//
// The flat shape (#3 with extra fields alongside `tool`) is covered by the
// single-tool branch — all non-"tool" keys become the args map.
func callsFromObject(obj map[string]interface{}) []api.ToolCall {
	var calls []api.ToolCall
	for _, key := range []string{"tool_calls", "_calls"} {
		if arr, ok := obj[key].([]interface{}); ok {
			calls = append(calls, buildCallsFromArray(arr)...)
		}
	}
	if len(calls) > 0 {
		return calls
	}
	if name, ok := obj["tool"].(string); ok && knownTools[name] {
		args := map[string]interface{}{}
		for k, v := range obj {
			if k == "tool" {
				continue
			}
			args[k] = v
		}
			calls = append(calls, api.ToolCall{Name: name, Arguments: args})
	}
	return calls
}

func buildCallsFromArray(arr []interface{}) []api.ToolCall {
	var calls []api.ToolCall
	for _, c := range arr {
		m, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		name := ""
		if n, ok := m["name"].(string); ok {
			name = n
		}
		if name == "" {
			if fn, ok := m["function"].(map[string]interface{}); ok {
				if n, ok := fn["name"].(string); ok {
					name = n
				}
			}
		}
		if !knownTools[name] {
			continue
		}
		calls = append(calls, ToolCall{Name: name, Arguments: normalizeToolArgs(m)})
	}
	return calls
}

func normalizeToolArgs(m map[string]interface{}) map[string]interface{} {
	// Shape 1: Anthropic-style wrapped arguments.
	//   {"name":"bash","arguments":{"command":"..."}}
	if args, ok := m["arguments"]; ok {
		switch a := args.(type) {
		case map[string]interface{}:
			return a
		case string:
			var p map[string]interface{}
			if json.Unmarshal([]byte(a), &p) == nil {
				return p
			}
		}
	}
	// Shape 2: OpenAI-style wrapped arguments.
	//   {"name":"bash","function":{"name":"bash","arguments":"{...}"}}
	if fn, ok := m["function"].(map[string]interface{}); ok {
		if args, ok := fn["arguments"]; ok {
			switch a := args.(type) {
			case map[string]interface{}:
				return a
			case string:
				var p map[string]interface{}
				if json.Unmarshal([]byte(a), &p) == nil {
					return p
				}
			}
		}
	}
	// Shape 3: DeepSeek web model — flat shape with tool metadata and
	// arguments at the same level. The model emits this in TUI
	// sessions in the wild:
	//   {"tool_calls":[{"id":"x1","name":"bash","command":"find ..."}]}
	// Strip the metadata keys (name, id, function, tool_calls) and
	// return whatever's left as the args map.
	reserved := map[string]bool{
		"name": true, "id": true, "function": true, "tool_calls": true, "_calls": true,
	}
	flat := make(map[string]interface{}, len(m))
	for k, v := range m {
		if reserved[k] {
			continue
		}
		flat[k] = v
	}
	if len(flat) == 0 {
		return map[string]interface{}{}
	}
	return flat
}

// extractSingleJsonToolCalls is preserved as a public-ish entry point
// for the single-tool {"tool":"bash",...} shape. It now delegates to
// the jsonex-based extractJsonToolCalls (which handles every JSON shape
// — the single-tool object is one of them, via callsFromObject). Kept
// as a separate function so callers and tests that target this shape
// specifically continue to work.
func extractSingleJsonToolCalls(text string) []api.ToolCall {
	return extractJsonToolCalls(text)
}

// extractXmlToolCalls is a thin wrapper around api.ExtractXmlToolCalls
// that filters the results against knownTools. The byte-level scanner
// is in internal/api/ so other providers can use it; this layer is
// responsible for the per-provider "which tool names does THIS
// conversation loop's executor know how to run" gate.
func extractXmlToolCalls(text string) []api.ToolCall {
	all := api.ExtractXmlToolCalls(text)
	if len(all) == 0 {
		return nil
	}
	filtered := make([]api.ToolCall, 0, len(all))
	for _, c := range all {
		if !knownTools[c.Name] {
			continue
		}
		filtered = append(filtered, c)
	}
	if len(filtered) == 0 {
		return nil
	}
	return filtered
}

func extractCodeBlockToolCalls(text string) []api.ToolCall {
	idx := 0
	for {
		start := strings.Index(text[idx:], "```")
		if start == -1 {
			break
		}
		start += idx + 3
		// Skip optional language tag (e.g. ```json).
		if nl := strings.IndexByte(text[start:], '\n'); nl != -1 && nl < 20 {
			tag := strings.TrimSpace(text[start : start+nl])
			if len(tag) > 0 && !strings.Contains(tag, "{") {
				start += nl + 1
			}
		}
		end := strings.Index(text[start:], "```")
		if end == -1 {
			break
		}
		inner := strings.TrimSpace(text[start : start+end])
		if !strings.HasPrefix(inner, "{") {
			idx = start + end + 3
			continue
		}
		// extractJsonToolCalls is jsonex-based and now handles every
		// JSON shape we care about (tool_calls array, wrapped, flat,
		// and the single-tool {"tool":"bash",...} object) — so one call
		// replaces the previous two-tier (array-then-single) fallback.
		if calls := extractJsonToolCalls(inner); len(calls) > 0 {
			return calls
		}
		idx = start + end + 3
	}
	return nil
}

func extractReactToolCalls(text string) []api.ToolCall {
	lines := strings.Split(text, "\n")
	var calls []api.ToolCall
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "Action:") || i+1 >= len(lines) {
			continue
		}
		nextLine := strings.TrimSpace(lines[i+1])
		if !strings.HasPrefix(nextLine, "Action Input:") {
			continue
		}
		name := strings.TrimSpace(strings.TrimPrefix(trimmed, "Action:"))
		input := strings.TrimSpace(strings.TrimPrefix(nextLine, "Action Input:"))
		if !knownTools[name] {
			continue
		}
		var args map[string]interface{}
		if json.Unmarshal([]byte(input), &args) == nil {
		calls = append(calls, api.ToolCall{Name: name, Arguments: args})
			continue
		}
		// Action Input wasn't a JSON object. Two reasonable choices:
		//   (a) skip the call entirely, or
		//   (b) wrap the raw text as {"input": "..."} so the tool
		//       sees *something* and can decide whether to retry.
		// We used to do (b) and it was inconsistent — most paths return
		// a real map. Skipping is the safer default; the model's next
		// turn (or the XML/JSON fallbacks above) will pick up the call.
		_ = input
	}
	return calls
}

// extractFunctionCallToolCalls handles the <tool_call>...</tool_call>
// XML format the model sometimes emits. The previous version extracted
// only the name attribute and returned an empty `{}` arg map, which
// always caused the downstream tool to fail with "input is required".
//
// The body of the tag is usually a JSON object describing the arguments
// (the format the OpenAI/Anthropic function-call spec uses). We try
// jsonex on the body first (handles nested braces, escaped quotes, etc.)
// and fall back to the standard library. If neither parses, we skip the
// call rather than return a broken empty-args stub.
func extractFunctionCallToolCalls(text string) []api.ToolCall {
	idx := 0
	for {
		start := strings.Index(text[idx:], "<function_call")
		if start == -1 {
			break
		}
		start += idx
		endTag := "</function_call>"
		end := strings.Index(text[start:], endTag)
		if end == -1 {
			break
		}
		inner := text[start : start+end]
		nameStart := strings.Index(inner, "name=\"")
		if nameStart == -1 {
			idx = start + end + len(endTag)
			continue
		}
		nameStart += 6
		nameEnd := strings.Index(inner[nameStart:], "\"")
		if nameEnd == -1 {
			idx = start + end + len(endTag)
			continue
		}
		name := inner[nameStart : nameStart+nameEnd]
		if !knownTools[name] {
			idx = start + end + len(endTag)
			continue
		}
		// Extract everything after the opening tag's `name="..."` attr as
		// the candidate args body, then try to parse it as JSON.
		body := inner[nameStart+nameEnd+1:]
		var args map[string]interface{}
		if err := jsonex.Unmarshal([]byte(body), &args); err != nil {
			if err := json.Unmarshal([]byte(body), &args); err != nil {
				// Body wasn't JSON. Skip — the empty-args stub used to be
				// returned here, and it always failed downstream.
				idx = start + end + len(endTag)
				continue
			}
		}
		idx = start + end + len(endTag)
		return []api.ToolCall{{Name: name, Arguments: args}}
	}
	return nil
}

// collapseExcessiveToolCalls takes text that contains an unreasonable number
// of <tool_calls> tags (common when the model is in reasoning/thinking mode
// and echoes the tag pattern) and collapses the noise down to one copy of
// the first proper tool-call block, preserving its content and the prose
// around it.
//
// The previous version lost everything before the first non-empty block
// when it had to skip past several empty <tool_calls></tool_calls> echoes
// to find the real one. We now splice the chosen block back into the
// original text rather than rebuilding from the suffix, so surrounding
// reasoning is preserved.
func collapseExcessiveToolCalls(text string) string {
	open := "<tool_calls>"
	close := "</tool_calls>"
	for {
		start := strings.Index(text, open)
		if start == -1 {
			break
		}
		end := strings.Index(text[start+len(open):], close)
		if end == -1 {
			break
		}
		end += start + len(open)
		inner := strings.TrimSpace(text[start+len(open) : end])
		if inner != "" {
			prefix := text[:start]
			prefix = strings.TrimRight(prefix, " \t")
			suffix := text[end+len(close):]
			return prefix + open + inner + close + suffix
		}
		text = text[:start] + text[end+len(close):]
	}
	return text
}

// StripToolCalls re-exports api.StripToolCalls so the existing deepseek
// test surface keeps working without duplicating the implementation.
// The canonical home is internal/api/strip.go — any new caller should
// import "claw-code-go/internal/api" directly.
func StripToolCalls(text string) string {
	return api.StripToolCalls(text)
}
