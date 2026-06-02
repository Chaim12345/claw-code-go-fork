# TUI Enhancements - Event-Driven Debug System & Real-Time Rendering

## Overview

This document describes the comprehensive TUI enhancements added to claw-code-go, including:
- Event-driven debugging system with state machine tracking
- Real-time markdown rendering during streaming
- Smart tool call parsing and display (hiding raw XML)
- Debug panel overlay for live system inspection

## Features

### 1. Event-Driven Debugging System

**Location:** `internal/tui/debug/`

#### Components

**`events.go`** - Event logging system
- Structured event logging with types: StateChange, ToolCall, StreamChunk, etc.
- Circular buffer (default 1000 events)
- Event filtering by type
- Event handlers for custom processing
- Thread-safe with mutex protection

**`state_machine.go`** - State transition tracking
- Validates state transitions
- Tracks state history (default 100 transitions)
- Collects metrics per state (enter count, duration, avg/max/min)
- Detects invalid transitions

**`panel.go`** - Debug panel UI
- 4 tabs: Events, States, Metrics, Tools
- Real-time event display with filtering
- State transition history with durations
- Performance metrics per state
- Tool call timeline

#### Usage

Enable debug mode:
```bash
CLAW_DEBUG=1 ./claw-code-go --provider deepseek
```

Open debug panel: **Ctrl+D**

**Debug Panel Controls:**
- `Tab` / `Shift+Tab` - Switch tabs
- `↑` / `↓` - Scroll content
- `c` - Clear events
- `f` - Filter events (future enhancement)
- `q` or `Esc` - Close panel

#### Event Types

```go
EventStateChange    // State transitions (input -> busy -> input)
EventToolCall       // Tool execution start
EventToolComplete   // Tool execution complete
EventStreamChunk    // Streaming text received
EventStreamComplete // Stream finished
EventStreamError    // Stream error
EventRender         // UI render event
EventKeyPress       // Key input
EventResize         // Window resize
EventPermission     // Permission request
EventSession        // Session operations
EventCompaction     // Context compaction
EventError          // Error events
EventWarning        // Warning events
EventInfo           // Info events
```

### 2. Real-Time Markdown Rendering

**Location:** `internal/tui/streaming_renderer.go`

#### StreamingRenderer

Renders markdown incrementally as it arrives from the API:

**Features:**
- Throttled rendering (50ms interval) to reduce CPU usage
- Code block detection and tracking
- Tool call XML detection (for hiding)
- Glamour-based markdown rendering
- Fallback to plain text on render errors

**Usage:**
```go
renderer := NewStreamingRenderer()
renderer.Append("# Hello\n")
renderer.Append("This is **bold**\n")
rendered := renderer.Render() // Returns styled terminal output
```

#### ProgressiveMarkdownRenderer

Alternative renderer that processes chunks independently:

**Features:**
- Chunk-based rendering
- Only renders new content since last call
- Maintains rendered chunk cache
- Useful for very long documents

#### CodeBlockRenderer

Specialized renderer for code blocks with syntax highlighting:

**Features:**
- Language-specific syntax highlighting
- Render caching for performance
- Fallback styling when glamour fails
- Border and padding for visual separation

### 3. Smart Tool Call Parsing

**Location:** `internal/tui/tool_parser.go`

#### ToolCallParser

Extracts and parses tool calls from streaming XML:

**Features:**
- Real-time XML parsing from stream
- Handles incomplete XML (partial buffering)
- Extracts tool name and parameters
- Tracks tool call status (pending, running, success, failed)
- Thread-safe operation

**Tool Call Lifecycle:**
1. **Pending** - Tool call detected, not yet executed
2. **Running** - Tool is executing
3. **Success** - Tool completed successfully
4. **Failed** - Tool execution failed
5. **Cancelled** - Tool execution cancelled

#### ToolCallManager

Manages tool call display and state:

**Features:**
- Processes streaming text to extract tool calls
- Removes raw XML from visible output
- Renders tool calls as pretty cards
- Expandable/collapsible tool details
- Status tracking with icons

**Tool Call Card Display:**
```
⚡ read_file running
  
  file_path: /path/to/file.txt
  
  Duration: 150ms
```

**Expanded View:**
```
✓ read_file success

  file_path: /path/to/file.txt
  
  Result: File content retrieved (1024 bytes)
  
  Duration: 150ms
```

### 4. Integration with TUI

**Location:** `internal/tui/model.go`

#### New Model Fields

```go
// Streaming and rendering
streamingRenderer *StreamingRenderer
toolCallManager   *ToolCallManager
codeBlockRenderer *CodeBlockRenderer

// Debug system
debugEnabled bool
stateMachine *debug.StateMachine
```

#### State Transitions with Logging

All state transitions now go through `transitionState()`:

```go
func (m *Model) transitionState(to appState, reason string) {
    if m.debugEnabled {
        // Log to debug system
        m.stateMachine.Transition(debug.State(stateToString(to)), reason)
        debug.Log(debug.EventStateChange, ...)
    }
    m.state = to
}
```

#### Stream Processing

Stream chunks are now processed through multiple layers:

1. **Tool Call Manager** - Extracts and hides XML tool calls
2. **Streaming Renderer** - Renders markdown in real-time
3. **Viewport** - Displays final styled output

```go
case streamDeltaMsg:
    // Process through tool call manager to hide XML
    cleanText := m.toolCallManager.ProcessStream(msg.text)
    
    // Add to streaming renderer for real-time markdown rendering
    m.streamingRenderer.Append(cleanText)
    
    // Get rendered content
    m.streamBuf = m.streamingRenderer.Render()
    
    // Add tool call cards
    toolCallsRendered := m.toolCallManager.RenderActiveCalls()
    if toolCallsRendered != "" {
        m.streamBuf += "\n" + toolCallsRendered
    }
```

## Architecture

### Event Flow

```
User Input
    ↓
Key Handler → Debug Log (EventKeyPress)
    ↓
State Transition → Debug Log (EventStateChange)
    ↓
API Call
    ↓
Stream Chunks → Debug Log (EventStreamChunk)
    ↓
Tool Call Parser → Debug Log (EventToolCall)
    ↓
Streaming Renderer
    ↓
Viewport Update → Debug Log (EventRender)
```

### State Machine

```
input ←→ busy ←→ permission
  ↓       ↓
picker  ask_user
  ↓
palette
  ↓
session_picker
  ↓
debug_panel
```

## Performance Considerations

### Rendering Throttling

- Markdown rendering throttled to 50ms intervals
- Prevents excessive CPU usage during fast streaming
- Cached results reused within throttle window

### Event Buffer Management

- Circular buffer prevents unbounded memory growth
- Default 1000 events (configurable)
- Old events automatically discarded

### Tool Call Caching

- Code block rendering cached by language+content
- Reduces redundant syntax highlighting
- Cache cleared on session reset

## Future Enhancements

### Planned Features

1. **Event Filtering UI**
   - Interactive filter selection in debug panel
   - Multiple filter types simultaneously
   - Save/load filter presets

2. **Performance Profiling**
   - FPS counter in debug panel
   - Render time tracking
   - Memory usage monitoring

3. **Event Export**
   - Export events to JSON
   - Export state transitions to CSV
   - Generate performance reports

4. **Advanced Tool Display**
   - Diff rendering for file edits
   - Syntax highlighting in tool parameters
   - Tool execution timeline visualization

5. **Replay System**
   - Record and replay sessions
   - Step through state transitions
   - Debug specific scenarios

## Troubleshooting

### Debug Mode Not Working

Check environment variable:
```bash
echo $CLAW_DEBUG
# Should output: 1
```

### Panel Not Opening

- Ensure debug mode is enabled (`CLAW_DEBUG=1`)
- Press `Ctrl+D` to toggle
- Check terminal size (minimum 80x24)

### Rendering Issues

- Update terminal to support 256 colors
- Check glamour compatibility
- Fallback rendering should work in all terminals

### Performance Issues

- Reduce event buffer size in code
- Increase render throttle interval
- Disable debug mode in production

## Examples

### Example 1: Monitoring State Transitions

```bash
CLAW_DEBUG=1 ./claw-code-go --provider deepseek
# Press Ctrl+D to open debug panel
# Navigate to "States" tab
# Watch state transitions in real-time
```

### Example 2: Tracking Tool Calls

```bash
CLAW_DEBUG=1 ./claw-code-go --provider deepseek
# Send message: "Read the README.md file"
# Press Ctrl+D to open debug panel
# Navigate to "Tools" tab
# See tool call timeline with durations
```

### Example 3: Performance Analysis

```bash
CLAW_DEBUG=1 ./claw-code-go --provider deepseek
# Have a long conversation
# Press Ctrl+D to open debug panel
# Navigate to "Metrics" tab
# Review state durations and enter counts
```

## API Reference

### Debug Logger

```go
// Log an event
debug.Log(debug.EventToolCall, "Tool executed", map[string]interface{}{
    "tool_name": "read_file",
    "duration": "150ms",
})

// Log with duration
debug.LogWithDuration(debug.EventRender, "Viewport rendered", nil, duration)

// Get events
events := debug.GetEvents()
filtered := debug.GetEventsByType(debug.EventToolCall)
recent := debug.GetRecentEvents(10)

// Control logging
debug.Enable()
debug.Disable()
debug.Clear()
```

### State Machine

```go
// Transition state
err := debug.TransitionState(debug.StateBusy, "user sent message")

// Get current state
current := debug.GetCurrentState()
previous := debug.GetPreviousState()

// Get metrics
metrics := debug.GetStateMetrics()
timeInState := debug.GetTimeInCurrentState()
```

### Streaming Renderer

```go
// Create renderer
renderer := NewStreamingRenderer()

// Add content
renderer.Append("# Title\n")
renderer.Append("Some **bold** text\n")

// Get rendered output
output := renderer.Render()

// Check for code blocks
lang, code, isOpen := renderer.GetCurrentCodeBlock()

// Reset for new stream
renderer.Reset()
```

### Tool Call Manager

```go
// Create manager
manager := NewToolCallManager()

// Process stream
cleanText := manager.ProcessStream(streamChunk)

// Render active calls
rendered := manager.RenderActiveCalls()

// Update status
manager.UpdateStatus(callID, ToolCallSuccess, "result", "")

// Toggle expansion
manager.ToggleExpanded(callID)
```

## Contributing

When adding new features:

1. Add appropriate debug events
2. Update state machine transitions if needed
3. Add metrics tracking for performance-critical code
4. Document new event types
5. Update this README

## License

Same as claw-code-go main project.
