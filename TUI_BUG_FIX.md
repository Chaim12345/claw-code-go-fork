# TUI Bug Fix: Repeated "tool_calls pending ⏳" Output

## Problem

The TUI is displaying repeated lines of:
```
tool_calls pending ⏳ tool_calls pending ⏳ tool_calls pending ⏳
```

## Root Cause Analysis

### Issue 1: Overly Greedy XML Parser
**Location:** `internal/tui/tool_parser.go:66-105` (`extractToolCalls`)

The parser treats **ANY** XML tag as a potential tool call:
```go
// Find opening tag
start := strings.Index(content, "<")
if start == -1 {
    break
}
```

This means it picks up:
- Markdown XML (like `<thinking>`, `<result>`)
- HTML entities
- Any other XML-like content

### Issue 2: No Tool Name Validation
**Location:** `internal/tui/tool_parser.go:120-180` (`parseToolCallXML`)

The parser accepts ANY tag name as a tool name without validation:
```go
if toolName == "" {
    return ParsedToolCall{}, fmt.Errorf("no tool name found")
}
// No validation that toolName is a valid tool!
```

### Issue 3: Aggressive XML Removal
**Location:** `internal/tui/tool_parser.go:368-385` (`removeToolCallXML`)

The function removes **ALL** XML tags from the stream:
```go
for _, ch := range text {
    if ch == '<' {
        inTag = true
        continue
    }
    if ch == '>' {
        inTag = false
        continue
    }
    if !inTag {
        result.WriteRune(ch)
    }
}
```

This breaks:
- Markdown with XML tags
- Code examples with XML
- Any legitimate XML content in responses

## Solution

### Fix 1: Whitelist Known Tool Names

Add a list of valid tool names and only parse those:

```go
var validToolNames = map[string]bool{
    "read_file":        true,
    "write_to_file":    true,
    "list_files":       true,
    "search_file_content": true,
    "glob":             true,
    "search_and_replace": true,
    "execute_command":  true,
    "web_fetch":        true,
    "save_memory":      true,
    "attempt_completion": true,
    "switch_mode":      true,
    "insert_content":   true,
    "update_todo_list": true,
    "ask_followup_question": true,
    "apply_diff":       true,
    "replace_regex":    true,
    "restore":          true,
}

func isValidToolName(name string) bool {
    return validToolNames[name]
}
```

### Fix 2: Validate Tool Names in Parser

```go
// In parseToolCallXML, after extracting toolName:
if toolName == "" {
    return ParsedToolCall{}, fmt.Errorf("no tool name found")
}

// ADD THIS:
if !isValidToolName(toolName) {
    return ParsedToolCall{}, fmt.Errorf("invalid tool name: %s", toolName)
}
```

### Fix 3: Only Remove Valid Tool Call XML

```go
func (tcm *ToolCallManager) removeToolCallXML(text string) string {
    result := text
    
    // Only remove XML for known tool names
    for toolName := range validToolNames {
        openTag := "<" + toolName + ">"
        closeTag := "</" + toolName + ">"
        
        // Remove complete tool call blocks
        for {
            start := strings.Index(result, openTag)
            if start == -1 {
                break
            }
            
            end := strings.Index(result[start:], closeTag)
            if end == -1 {
                break
            }
            
            // Remove the entire block
            result = result[:start] + result[start+end+len(closeTag):]
        }
    }
    
    return result
}
```

### Fix 4: Better Tool Call Detection

Instead of looking for ANY `<`, look specifically for tool call patterns:

```go
func (tcp *ToolCallParser) extractToolCalls() []ParsedToolCall {
    content := tcp.buffer.String()
    newCalls := make([]ParsedToolCall, 0)

    // Try each known tool name
    for toolName := range validToolNames {
        openTag := "<" + toolName + ">"
        closeTag := "</" + toolName + ">"
        
        start := strings.Index(content, openTag)
        if start == -1 {
            continue
        }
        
        closeIdx := strings.Index(content[start:], closeTag)
        if closeIdx == -1 {
            // Incomplete, save for next feed
            continue
        }
        
        // Extract and parse
        xmlBlock := content[start : start+closeIdx+len(closeTag)]
        if call, err := tcp.parseToolCallXML(xmlBlock); err == nil {
            newCalls = append(newCalls, call)
            tcp.completedCalls = append(tcp.completedCalls, call)
        }
    }

    return newCalls
}
```

## Implementation Status

### Fix 1: Whitelist Known Tool Names — ✅ COMPLETED
**Location:** `internal/tui/tool_parser.go:12-29`
- `validToolNames` map exists with all 10 actual tool names
- Matches tools registered in `internal/tools/`: `bash`, `read_file`, `write_file`, `glob`, `grep`, `file_edit`, `web_fetch`, `web_search`, `ask_user`, `todo_write`
- Also includes wrapper tags: `tool_calls`, `invoke`

### Fix 2: Validate Tool Names in Parser — ✅ COMPLETED
**Location:** `internal/tui/tool_parser.go:117` and `:212`
- `extractToolCalls()` skips non-valid tags at line 117
- `parseToolCallXML()` returns error for invalid tool names at line 212
- Dual validation ensures both streaming detection and final parsing reject unknown tags

### Fix 3: Only Remove Valid Tool Call XML — ✅ COMPLETED
**Location:** `internal/tui/tool_parser.go:310-380` (`removeToolCallXML`)
- Uses whitelist of `toolTags` + `paramTags`
- Only strips complete blocks matching known tags
- Incomplete blocks are left intact (won't cause display glitches)
- Preserves arbitrary XML-like content in markdown/code

### Fix 4: Better Tool Call Detection — ⚠️ PARTIALLY IMPLEMENTED
**Location:** `internal/tui/tool_parser.go:88-149` (`extractToolCalls`) and `streaming_renderer.go:85-117` (`detectToolCalls`)
- `extractToolCalls` still uses generic `<` scan (line 91) but skips non-valid tags at line 117 — functional but slightly wasteful
- `detectToolCalls` in streaming renderer is smarter: extracts tag name first, checks `validToolNames` before entering tool call mode (line 115)
- Redundancy: `streaming_renderer.go:115` checks `tagName == "tool_calls" || tagName == "invoke"` but these are already in `validToolNames`

### Minor Issue: Redundant check in streaming_renderer.go:115
`validToolNames` already includes `tool_calls` and `invoke`. The additional `|| tagName == "tool_calls" || tagName == "invoke"` check is redundant and can be simplified to just `validToolNames[tagName]`.

## Testing

After fixes, test with:
1. Normal tool calls (should work)
2. Markdown with `<thinking>` tags (should not be parsed as tool)
3. Code examples with XML (should not be removed)
4. Multiple tool calls in one response (should all be parsed correctly)

## Expected Behavior After Fix

- Only actual tool calls are parsed and displayed
- Markdown XML tags are preserved
- No repeated "tool_calls pending" messages
- Clean, readable output