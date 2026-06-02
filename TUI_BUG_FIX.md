# TUI Rendering Bug Fix

## Problem

The TUI displays repeated content when cycling through permission modes (Shift+Tab):

```
claw-code-go v0.1.0
Permission mode: plan
Permission mode: default
Permission mode: accept-edits
Permission mode: bypass
Permission mode: plan
Permission mode: default
...
```

## Root Cause

**Location:** `internal/tui/model.go:1220`

```go
func (m Model) cyclePermissionMode() (tea.Model, tea.Cmd) {
    // ... mode cycling logic ...
    m.viewBuf += statusStyle.Render(fmt.Sprintf("Permission mode: %s\n\n", next))
    m = m.refreshViewport()
    return m, nil
}
```

**Issue:** Each time the user presses Shift+Tab to cycle permission modes, the function **appends** the new mode message to `viewBuf`. Since `viewBuf` is the persistent conversation history buffer, these messages accumulate and never get cleared.

## Why This Happens

1. User presses Shift+Tab → `cyclePermissionMode()` is called
2. Mode changes from "plan" → "default"
3. Message "Permission mode: default" is **appended** to `viewBuf`
4. User presses Shift+Tab again → mode changes to "accept-edits"
5. Message "Permission mode: accept-edits" is **appended** to `viewBuf`
6. This continues indefinitely, creating a growing list of mode messages

## Solution

The permission mode is already displayed in the status bar (line 1613-1632):

```go
func (m Model) renderStatusBar() string {
    mode := "default"
    if m.loop != nil && m.loop.PermManager != nil {
        mode = m.loop.PermManager.Mode.String()
    }
    modeBadge := modeBadgeStyle.Render("[" + mode + "]")
    // ...
}
```

**Fix:** Remove the redundant `viewBuf` append since the mode is already visible in the status bar.

### Option 1: Remove the append (Recommended)

```go
func (m Model) cyclePermissionMode() (tea.Model, tea.Cmd) {
    if m.loop == nil || m.loop.PermManager == nil {
        m.viewBuf += warnStyle.Render("No active permission manager.\n\n")
        m = m.refreshViewport()
        return m, nil
    }
    cur := m.loop.PermManager.Mode
    idx := 0
    for i, p := range m.permModeOrder {
        if p == cur {
            idx = i
            break
        }
    }
    idx = (idx + 1) % len(m.permModeOrder)
    next := m.permModeOrder[idx]
    m.loop.PermManager.Mode = next
    m.cfg.PermissionMode = next.String()
    // REMOVED: m.viewBuf += statusStyle.Render(fmt.Sprintf("Permission mode: %s\n\n", next))
    m = m.refreshViewport()
    return m, nil
}
```

### Option 2: Use a temporary overlay message

If we want to show a brief confirmation, use a temporary message system instead of appending to the persistent buffer:

```go
func (m Model) cyclePermissionMode() (tea.Model, tea.Cmd) {
    // ... existing logic ...
    next := m.permModeOrder[idx]
    m.loop.PermManager.Mode = next
    m.cfg.PermissionMode = next.String()
    
    // Show temporary message (would need to implement a flash message system)
    // m.flashMessage = fmt.Sprintf("Permission mode: %s", next)
    
    m = m.refreshViewport()
    return m, nil
}
```

## Implementation

**File:** `internal/tui/model.go`  
**Line:** 1220  
**Action:** Delete or comment out the line that appends to `viewBuf`

```diff
  idx = (idx + 1) % len(m.permModeOrder)
  next := m.permModeOrder[idx]
  m.loop.PermManager.Mode = next
  m.cfg.PermissionMode = next.String()
- m.viewBuf += statusStyle.Render(fmt.Sprintf("Permission mode: %s\n\n", next))
  m = m.refreshViewport()
  return m, nil
```

## Testing

After applying the fix:

1. Start the TUI
2. Press Shift+Tab multiple times to cycle through permission modes
3. Verify that:
   - The status bar shows the current mode
   - No duplicate "Permission mode: X" messages appear in the conversation history
   - The logo appears only once at the top

## Additional Improvements

### 1. Add Debug Logging

Add a debug flag to log TUI state changes:

```go
// Add to model.go
var debugTUI = os.Getenv("CLAW_DEBUG_TUI") == "1"

func debugLog(format string, args ...interface{}) {
    if debugTUI {
        fmt.Fprintf(os.Stderr, "[TUI-DEBUG] "+format+"\n", args...)
    }
}

// Use in cyclePermissionMode
func (m Model) cyclePermissionMode() (tea.Model, tea.Cmd) {
    // ...
    debugLog("Permission mode changed: %s -> %s", cur, next)
    // ...
}
```

### 2. Build with Debug Mode

Add build tags for debug builds:

```bash
# Build with debug logging enabled
CLAW_DEBUG_TUI=1 go build -tags debug -o claw-code-go-debug ./cmd/claw-code-go

# Or use ldflags to inject version info
go build -ldflags "-X main.version=0.1.0-debug" ./cmd/claw-code-go
```

### 3. System Prompt XML Preference

Update the system prompt to prefer XML tool calls over JSON (for DeepSeek compatibility):

**File:** `internal/runtime/conversation.go`  
**Location:** `systemPromptBase` constant

```go
const systemPromptBase = `You are Claude Code, an AI assistant for software engineering tasks. You have access to tools for running bash commands, reading and writing files, searching with glob patterns, and grepping for patterns in code. Use these tools to help users with coding tasks.

IMPORTANT: When using tools, prefer XML format over JSON:
<tool_name>
<parameter_name>value</parameter_name>
</tool_name>

JSON format is supported as a fallback but XML is preferred for better streaming compatibility.`
```

### 4. Optimize XML Streaming Parser

**File:** `internal/api/providers/deepseek/stream.go`

Current issue: The parser tries multiple strategies on every chunk, even when no tool calls are present.

Optimization:

```go
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
    
    // Try XML first (preferred for DeepSeek)
    if calls := extractXmlToolCalls(text); len(calls) > 0 {
        return calls
    }
    
    // Fallback to JSON
    if calls := extractJsonToolCalls(text); len(calls) > 0 {
        return calls
    }
    
    // Other fallbacks...
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
```

## Summary

The bug is caused by accumulating permission mode messages in the persistent conversation buffer. The fix is simple: remove the line that appends to `viewBuf` since the mode is already displayed in the status bar.
