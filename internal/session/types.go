package session

import "time"

// Session represents a chat session stored in the sessions table.
type Session struct {
	ID               string    `json:"id"`
	CreatedAt        time.Time `json:"created_at"`
	LastActiveAt     time.Time `json:"last_active_at"`
	Provider         string    `json:"provider"`
	Model            string    `json:"model"`
	MessageCount     int       `json:"message_count"`
	FirstUserMsg     string    `json:"first_user_msg,omitempty"`
	LastAssistantMsg string    `json:"last_assistant_msg,omitempty"`
}

// Message represents a single message (user or assistant) in a session.
type Message struct {
	ID         int64     `json:"id"`
	SessionID  string    `json:"session_id"`
	Role       string    `json:"role"` // "user" or "assistant"
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
	TokensUsed int       `json:"tokens_used"`
}

// ToolCall represents a single tool invocation linked to a message.
type ToolCall struct {
	ID         int64  `json:"id"`
	MessageID  int64  `json:"message_id"`
	ToolName   string `json:"tool_name"`
	Input      string `json:"input"`  // JSON-serialized tool input
	Output     string `json:"output"` // truncated tool output
	DurationMs int    `json:"duration_ms"`
}

// ExportedMessage is a message with its associated tool calls, used for
// session export/import. This nested structure allows JSON export to
// include tool calls inline under each message.
type ExportedMessage struct {
	ID        int64      `json:"id"`
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	CreatedAt time.Time  `json:"created_at"`
	TokensUsed int       `json:"tokens_used"`
	ToolCalls []*ToolCall `json:"tool_calls,omitempty"`
}

// ExportedSession is the full session tree used for export/import.
// It contains the session metadata and all messages with their tool calls.
type ExportedSession struct {
	Session  *Session          `json:"session"`
	Messages []*ExportedMessage `json:"messages"`
}