package tui

import (
	"claw-code-go/internal/tui/debug"
	"fmt"
)

// setState is a helper to transition state with optional debug logging
func (m *Model) setState(to appState, reason string) {
	if m.debugEnabled {
		from := m.state
		m.stateMachine.Transition(debug.State(stateToString(to)), reason)
		debug.Log(debug.EventStateChange, fmt.Sprintf("State: %s -> %s", stateToString(from), stateToString(to)), map[string]interface{}{
			"from":   stateToString(from),
			"to":     stateToString(to),
			"reason": reason,
		})
	}
	m.state = to
}

// logKeyPress logs keyboard input when debug is enabled
func (m *Model) logKeyPress(key string) {
	if m.debugEnabled {
		debug.Log(debug.EventKeyPress, fmt.Sprintf("Key pressed: %s", key), map[string]interface{}{
			"key":   key,
			"state": stateToString(m.state),
		})
	}
}

// logResize logs window resize events
func (m *Model) logResize(width, height int) {
	if m.debugEnabled {
		debug.Log(debug.EventResize, fmt.Sprintf("Window resized: %dx%d", width, height), map[string]interface{}{
			"width":  width,
			"height": height,
		})
	}
}

// logRender logs render events with timing
func (m *Model) logRender(component string, duration int64) {
	if m.debugEnabled {
		debug.Log(debug.EventRender, fmt.Sprintf("Rendered: %s", component), map[string]interface{}{
			"component": component,
			"duration_ns": duration,
		})
	}
}

// logError logs error events
func (m *Model) logError(context string, err error) {
	if m.debugEnabled {
		debug.Log(debug.EventError, fmt.Sprintf("Error in %s: %v", context, err), map[string]interface{}{
			"context": context,
			"error":   err.Error(),
		})
	}
}

// logWarning logs warning events
func (m *Model) logWarning(message string, data map[string]interface{}) {
	if m.debugEnabled {
		debug.Log(debug.EventWarning, message, data)
	}
}

// logSession logs session-related events
func (m *Model) logSession(action string, sessionID string) {
	if m.debugEnabled {
		debug.Log(debug.EventSession, fmt.Sprintf("Session %s: %s", action, sessionID), map[string]interface{}{
			"action":     action,
			"session_id": sessionID,
		})
	}
}

// logCompaction logs compaction events
func (m *Model) logCompaction(messageCount int, tokensSaved int) {
	if m.debugEnabled {
		debug.Log(debug.EventCompaction, fmt.Sprintf("Compacted %d messages, saved %d tokens", messageCount, tokensSaved), map[string]interface{}{
			"message_count": messageCount,
			"tokens_saved":  tokensSaved,
		})
	}
}

// logPermission logs permission decision events
func (m *Model) logPermission(tool string, decision string) {
	if m.debugEnabled {
		debug.Log(debug.EventPermission, fmt.Sprintf("Permission for %s: %s", tool, decision), map[string]interface{}{
			"tool":     tool,
			"decision": decision,
		})
	}
}
