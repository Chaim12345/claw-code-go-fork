package session

import "time"

// SessionStore defines the contract for session, message, and tool-call
// persistence. The concrete Store (SQLite-backed) implements this
// interface, making it possible to swap in mock stores for testing or
// alternative backends in the future.
type SessionStore interface {
	// ── Sessions ───────────────────────────────────────────────

	// CreateSession inserts a new session. No-op if the id already exists.
	CreateSession(id, provider, model string) error

	// GetSession retrieves a session by id. Returns sql.ErrNoRows if missing.
	GetSession(id string) (*Session, error)

	// ListSessions returns all sessions, newest first.
	ListSessions() ([]*Session, error)

	// GetRecentSessions returns the most recent sessions up to limit,
	// ordered by last_active_at descending.
	GetRecentSessions(limit int) ([]*Session, error)

	// GetSessionSummary returns a condensed summary of a session:
	// the first user message (what was being worked on), the last
	// assistant reply, and any tool calls made with their outcomes.
	// Suitable for injecting as "resume context" on reconnect.
	GetSessionSummary(id string) (string, error)

	// SearchSessions performs a full-text search across all message content
	// and returns matching sessions ordered by last_active_at DESC.
	// It uses the FTS5 index created in SchemaV2.
	SearchSessions(query string) ([]*Session, error)

	// FilterSessions returns sessions matching the given provider, model,
	// and/or time range filters. Empty/zero-value parameters are ignored.
	// Results are ordered by last_active_at DESC.
	FilterSessions(provider, model string, after, before time.Time) ([]*Session, error)

	// DeleteSession removes a session and cascades to its messages/tool calls.
	DeleteSession(id string) error

	// UpdateSessionPreview updates the session's preview fields and message count.
	// firstUserMsg is set only the first time (never overwritten).
	UpdateSessionPreview(id, firstUserMsg, lastAssistantMsg string, msgCount int) error

	// ── Messages ───────────────────────────────────────────────

	// RecordMessage inserts a message and returns its auto-generated id.
	RecordMessage(sessionID, role, content string, tokensUsed int) (int64, error)

	// GetMessages returns all messages for a session, chronological order.
	GetMessages(sessionID string) ([]*Message, error)

	// ── Tool Calls ─────────────────────────────────────────────

	// RecordToolCall inserts a tool-call row linked to a message.
	RecordToolCall(messageID int64, toolName, input, output string, durationMs int) error

	// GetToolCalls returns all tool calls for a message.
	GetToolCalls(messageID int64) ([]*ToolCall, error)

	// ── Export / Import ──────────────────────────────────────────

	// ExportSessionMarkdown returns the session and all its messages/tool calls
	// formatted as markdown (## Session heading, ### role subheadings,
	// #### Tool Calls blocks under assistant messages).
	ExportSessionMarkdown(id string) (string, error)

	// ExportSessionJSON returns the full session tree (session metadata +
	// messages with nested tool calls) as pretty-printed JSON.
	ExportSessionJSON(id string) ([]byte, error)

	// ImportSessionJSON parses a JSON ExportedSession, validates required
	// fields (provider, model), creates the session and all messages/tool
	// calls in a transaction, and returns the new session id.
	ImportSessionJSON(data []byte) (string, error)

	// ImportSessionMarkdown parses the markdown export format (### role
	// headers, #### Tool Calls blocks), creates a session and messages,
	// and returns the new session id.
	ImportSessionMarkdown(data []byte) (string, error)

	// ── Lifecycle ──────────────────────────────────────────────

	// Close releases the underlying database handle.
	Close() error
}

// Compile-time assertion: *Store implements SessionStore.
var _ SessionStore = (*Store)(nil)