# Connection Drop After First Message - Analysis & Fix

## Problem

After sending the first message successfully, subsequent messages fail and the connection drops. The agent doesn't respond to the second message.

## Root Cause Analysis

### DeepSeek Session Management

The DeepSeek provider uses a stateful session model:

1. **Chat Session ID**: Created once per client instance, reused for all messages
2. **Parent Message ID**: Updated after each turn to chain messages together

**File:** `internal/api/providers/deepseek/provider.go`

```go
type Client struct {
    model string
    web   *WebClient

    mu              sync.Mutex
    chatSessionID   string      // Created once, reused
    parentMessageID string      // Updated after each turn
    // ...
}
```

### Message Flow

**First Message (Works):**
1. `ensureSession()` creates new chat session → `chatSessionID` set
2. `parentMessageID` is empty (no parent)
3. Message sent successfully
4. Response streams back
5. `parentMessageID` updated from stream events

**Second Message (Fails):**
1. `ensureSession()` returns early (session exists)
2. Uses existing `chatSessionID` and `parentMessageID`
3. **PROBLEM**: If `parentMessageID` wasn't set correctly, or session expired, request fails
4. Connection drops, no retry

### Identified Issues

#### 1. **Race Condition in Parent ID Update**

**Location:** Lines 304 and 386

```go
// During streaming (line 304)
if event.Event == "message_id" {
    c.mu.Lock()
    c.parentMessageID = event.Data
    c.mu.Unlock()
    return true
}

// After streaming completes (line 386)
if newID != "" {
    c.mu.Lock()
    c.parentMessageID = newID
    c.mu.Unlock()
}
```

**Problem:** Two different places update `parentMessageID`. If the stream is interrupted or the "message_id" event doesn't arrive, the parent ID might not be set.

#### 2. **No Session Expiry Handling**

DeepSeek sessions can expire. There's no logic to detect expired sessions and recreate them.

#### 3. **No Validation of Parent ID**

The code doesn't validate that `parentMessageID` was actually set after a successful turn.

#### 4. **Client Instance Reuse**

**Location:** `cmd/claw-code-go/main.go:108`

```go
loop := runtime.NewConversationLoop(cfg, realClient)
```

The same `Client` instance is reused for the entire TUI session. If the DeepSeek session expires or becomes invalid, there's no way to recover.

## Solution

### Fix 1: Add Session Recovery Logic

**File:** `internal/api/providers/deepseek/provider.go`

Add method to reset session state:

```go
// resetSession clears the session state, forcing a new session on next request
func (c *Client) resetSession() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.chatSessionID = ""
    c.parentMessageID = ""
}

// isSessionError checks if an error indicates session expiry/invalidity
func isSessionError(errMsg string) bool {
    low := strings.ToLower(errMsg)
    return strings.Contains(low, "session") &&
           (strings.Contains(low, "expired") ||
            strings.Contains(low, "invalid") ||
            strings.Contains(low, "not found"))
}
```

### Fix 2: Add Session Recovery in StreamResponse

**Location:** After line 386 (after setting parentMessageID)

```go
if err != nil {
    // Check if this is a session error
    if isSessionError(err.Error()) {
        fmt.Fprintf(os.Stderr, "[deepseek] session error detected, will reset on next request: %v\n", err)
        c.resetSession()
    }
    
    send(api.StreamEvent{
        Type:         api.EventError,
        ErrorMessage: fmt.Sprintf("deepseek: %s", err.Error()),
    })
    return
}
```

### Fix 3: Validate Parent ID After Streaming

**Location:** After line 386

```go
if newID != "" {
    c.mu.Lock()
    c.parentMessageID = newID
    c.mu.Unlock()
} else {
    // Warning: no parent ID received, next turn might fail
    fmt.Fprintf(os.Stderr, "[deepseek] warning: no parent message ID received from stream\n")
}
```

### Fix 4: Add Debug Logging

Add environment variable for debug logging:

```go
var debugDeepSeek = os.Getenv("DEEPSEEK_DEBUG") == "1"

func debugLog(format string, args ...interface{}) {
    if debugDeepSeek {
        fmt.Fprintf(os.Stderr, "[deepseek-debug] "+format+"\n", args...)
    }
}

// Use in key places:
func (c *Client) StreamResponse(ctx context.Context, req api.CreateMessageRequest) (<-chan api.StreamEvent, error) {
    c.mu.Lock()
    sessID := c.chatSessionID
    parentID := c.parentMessageID
    c.mu.Unlock()
    
    debugLog("StreamResponse: session=%s parent=%s", sessID, parentID)
    
    // ... rest of implementation
}
```

### Fix 5: Add Session Validation

**Location:** In `ensureSession()` after line 119

```go
func (c *Client) ensureSession(ctx context.Context) error {
    c.mu.Lock()
    defer c.mu.Unlock()
    if c.chatSessionID != "" {
        debugLog("Reusing existing session: %s", c.chatSessionID)
        return nil
    }
    
    debugLog("Creating new chat session")
    sessID, err := c.web.CreateChatSession()
    if err != nil {
        return fmt.Errorf("deepseek: create session: %w", err)
    }
    c.chatSessionID = sessID
    debugLog("Created session: %s", sessID)
    return nil
}
```

## Testing

### Enable Debug Logging

```bash
export DEEPSEEK_DEBUG=1
./claw-code-go --provider deepseek
```

### Test Scenario

1. Send first message → should work
2. Check debug output for session ID and parent message ID
3. Send second message → should work now
4. Verify parent message ID is being set correctly

### Expected Debug Output

```
[deepseek-debug] Creating new chat session
[deepseek-debug] Created session: abc123...
[deepseek-debug] StreamResponse: session=abc123... parent=
[deepseek-debug] Parent message ID set: msg456...
[deepseek-debug] StreamResponse: session=abc123... parent=msg456...
```

## Additional Improvements

### 1. Add Session Health Check

```go
func (c *Client) validateSession(ctx context.Context) error {
    c.mu.Lock()
    sessID := c.chatSessionID
    c.mu.Unlock()
    
    if sessID == "" {
        return nil // Will be created on next request
    }
    
    // TODO: Add API call to validate session is still active
    // For now, just log the current state
    debugLog("Session validation: session=%s", sessID)
    return nil
}
```

### 2. Add Retry with Session Reset

```go
// In StreamResponse, wrap the retry loop:
const maxSessionRetries = 2
for sessionRetry := 0; sessionRetry < maxSessionRetries; sessionRetry++ {
    if sessionRetry > 0 {
        fmt.Fprintf(os.Stderr, "[deepseek] session retry %d/%d\n", sessionRetry+1, maxSessionRetries)
        c.resetSession()
        if err := c.ensureSession(ctx); err != nil {
            return nil, err
        }
    }
    
    // Existing retry loop for transient errors
    for attempt := 0; attempt < maxRetries; attempt++ {
        // ... existing code
    }
    
    // If we got a session error, retry with new session
    if err != nil && isSessionError(err.Error()) {
        continue
    }
    break
}
```

### 3. Add Metrics

```go
type SessionMetrics struct {
    TotalRequests      int
    SessionResets      int
    ParentIDMismatches int
}

var metrics SessionMetrics
var metricsMu sync.Mutex

func recordSessionReset() {
    metricsMu.Lock()
    metrics.SessionResets++
    metricsMu.Unlock()
}

// Export metrics for monitoring
func GetSessionMetrics() SessionMetrics {
    metricsMu.Lock()
    defer metricsMu.Unlock()
    return metrics
}
```

## Implementation Priority

1. ✅ **HIGH**: Add `resetSession()` and `isSessionError()` methods
2. ✅ **HIGH**: Add session error detection and recovery in StreamResponse
3. ✅ **HIGH**: Add debug logging with DEEPSEEK_DEBUG env var
4. ✅ **MEDIUM**: Validate parent ID after streaming
5. ✅ **MEDIUM**: Add session validation in ensureSession
6. 🔵 **LOW**: Add session health check API call
7. 🔵 **LOW**: Add comprehensive metrics

## Summary

The connection drop is likely caused by:
1. Session expiry not being detected
2. Parent message ID not being set correctly
3. No recovery mechanism when session becomes invalid

The fix adds:
- Session reset capability
- Error detection for session issues
- Debug logging to diagnose problems
- Validation of parent message ID
- Automatic recovery on session errors
