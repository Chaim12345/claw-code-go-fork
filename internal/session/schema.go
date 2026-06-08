// Package session provides SQLite-backed session persistence for
// claw-code-go chat sessions. It stores full conversation history
// (messages, tool calls) so sessions survive server restarts and
// users can search/browse old conversations.
//
// Schema overview:
//
//	sessions  — one row per chat session (metadata)
//	messages  — one row per message (user / assistant)
//	tool_calls — one row per tool invocation (linked to a message)
//
// Foreign keys with ON DELETE CASCADE ensure that deleting a session
// also removes its messages, and deleting a message also removes its
// tool calls.
//
// The store uses modernc.org/sqlite (pure Go, no CGO) and keeps the
// database at ~/.claw-code/sessions.db by default.
package session

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ── Schema versioning ────────────────────────────────────────────

// currentVersion is the active schema version. Increment this when
// adding new migrations. The migrations slice must contain exactly
// currentVersion entries, each idempotent (uses IF NOT EXISTS).
const currentVersion = 2

// ── SQL constants ─────────────────────────────────────────────────

// SchemaV1 creates the initial tables and indexes.
const SchemaV1 = `
-- Schema version 1: initial tables for sessions, messages, tool_calls.

CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
    id               TEXT    PRIMARY KEY,
    created_at       TEXT    NOT NULL,
    last_active_at   TEXT    NOT NULL,
    provider         TEXT    NOT NULL DEFAULT '',
    model            TEXT    NOT NULL DEFAULT '',
    message_count    INTEGER NOT NULL DEFAULT 0,
    first_user_msg   TEXT    DEFAULT '',
    last_assistant_msg TEXT  DEFAULT ''
);

CREATE TABLE IF NOT EXISTS messages (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id   TEXT    NOT NULL,
    role         TEXT    NOT NULL,            -- 'user' | 'assistant'
    content      TEXT    NOT NULL,            -- full message content
    created_at   TEXT    NOT NULL,
    tokens_used  INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_messages_session
    ON messages(session_id, created_at);

CREATE TABLE IF NOT EXISTS tool_calls (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    message_id  INTEGER NOT NULL,
    tool_name   TEXT    NOT NULL,
    input       TEXT    NOT NULL DEFAULT '',  -- JSON-serialized tool input
    output      TEXT    NOT NULL DEFAULT '',  -- truncated tool output
    duration_ms INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_tool_calls_message
    ON tool_calls(message_id);
`

// SchemaV2 adds a full-text search index on message content using FTS5.
// The content=messages, content_rowid=id mapping keeps the FTS index
// in sync with the messages table via triggers.
const SchemaV2 = `
-- Schema version 2: FTS5 full-text search on messages.content.

CREATE VIRTUAL TABLE IF NOT EXISTS messages_fts
USING fts5(content, content=messages, content_rowid=id);

CREATE TRIGGER IF NOT EXISTS messages_fts_ai AFTER INSERT ON messages BEGIN
	INSERT INTO messages_fts(rowid, content) VALUES (new.id, new.content);
END;

CREATE TRIGGER IF NOT EXISTS messages_fts_ad AFTER DELETE ON messages BEGIN
	INSERT INTO messages_fts(messages_fts, rowid, content) VALUES('delete', old.id, old.content);
END;

CREATE TRIGGER IF NOT EXISTS messages_fts_au AFTER UPDATE ON messages BEGIN
	INSERT INTO messages_fts(messages_fts, rowid, content) VALUES('delete', old.id, old.content);
	INSERT INTO messages_fts(rowid, content) VALUES (new.id, new.content);
END;
`

var migrations = []string{SchemaV1, SchemaV2}

// ── Store ─────────────────────────────────────────────────────────

// Store wraps a *sql.DB handle and provides CRUD operations for
// sessions, messages, and tool calls.
type Store struct {
	DB *sql.DB
}

// Open opens (or creates) the SQLite database at dbPath, runs any
// pending migrations, and returns a *Store ready for use.
//
// dbPath is typically "~/.claw-code/sessions.db". The caller should
// expand ~ before calling.
func Open(dbPath string) (*Store, error) {
	// Ensure the parent directory exists.
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("session store: create dir %s: %w", dir, err)
	}

	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("session store: open db: %w", err)
	}

	// Enable foreign keys (not supported via connection string for modernc.org/sqlite).
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("session store: enable foreign keys: %w", err)
	}

	store := &Store{DB: db}
	if err := store.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return store, nil
}

// Close closes the underlying database handle.
func (s *Store) Close() error {
	return s.DB.Close()
}

// ── Session CRUD ──────────────────────────────────────────────────

// CreateSession inserts a new session row. If a session with the
// given id already exists, it is a no-op (INSERT OR IGNORE).
func (s *Store) CreateSession(id, provider, model string) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := s.DB.Exec(
		`INSERT OR IGNORE INTO sessions (id, created_at, last_active_at, provider, model)
		 VALUES (?, ?, ?, ?, ?)`,
		id, now, now, provider, model,
	)
	return err
}

// GetSession retrieves a single session by id. Returns nil and
// sql.ErrNoRows if not found.
func (s *Store) GetSession(id string) (*Session, error) {
	var sess Session
	var createdAt, lastActiveAt string
	err := s.DB.QueryRow(
		`SELECT id, created_at, last_active_at, provider, model, message_count, first_user_msg, last_assistant_msg
		 FROM sessions WHERE id = ?`, id,
	).Scan(&sess.ID, &createdAt, &lastActiveAt, &sess.Provider, &sess.Model,
		&sess.MessageCount, &sess.FirstUserMsg, &sess.LastAssistantMsg)
	if err != nil {
		return nil, err
	}
	sess.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	sess.LastActiveAt, _ = time.Parse(time.RFC3339Nano, lastActiveAt)
	return &sess, nil
}

// ListSessions returns all sessions sorted by last_active_at
// descending (newest first).
func (s *Store) ListSessions() ([]*Session, error) {
	rows, err := s.DB.Query(
		`SELECT id, created_at, last_active_at, provider, model, message_count, first_user_msg, last_assistant_msg
		 FROM sessions ORDER BY last_active_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		var sess Session
		var createdAt, lastActiveAt string
		if err := rows.Scan(&sess.ID, &createdAt, &lastActiveAt, &sess.Provider, &sess.Model,
			&sess.MessageCount, &sess.FirstUserMsg, &sess.LastAssistantMsg); err != nil {
			return nil, err
		}
		sess.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		sess.LastActiveAt, _ = time.Parse(time.RFC3339Nano, lastActiveAt)
		sessions = append(sessions, &sess)
	}
	if sessions == nil {
		sessions = []*Session{}
	}
	return sessions, rows.Err()
}

// GetRecentSessions returns the most recent sessions up to limit,
// ordered by last_active_at descending (newest first). If limit <= 0,
// returns all sessions (same as ListSessions).
func (s *Store) GetRecentSessions(limit int) ([]*Session, error) {
	if limit <= 0 {
		return s.ListSessions()
	}
	rows, err := s.DB.Query(
		`SELECT id, created_at, last_active_at, provider, model, message_count, first_user_msg, last_assistant_msg
		 FROM sessions ORDER BY last_active_at DESC LIMIT ?`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		var sess Session
		var createdAt, lastActiveAt string
		if err := rows.Scan(&sess.ID, &createdAt, &lastActiveAt, &sess.Provider, &sess.Model,
			&sess.MessageCount, &sess.FirstUserMsg, &sess.LastAssistantMsg); err != nil {
			return nil, err
		}
		sess.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		sess.LastActiveAt, _ = time.Parse(time.RFC3339Nano, lastActiveAt)
		sessions = append(sessions, &sess)
	}
	if sessions == nil {
		sessions = []*Session{}
	}
	return sessions, rows.Err()
}

// GetSessionSummary returns a condensed summary of a session suitable for
// "resume context" injection when a user reconnects. It includes:
//   - Session metadata (provider, model, when it was active)
//   - The first user message (what they were working on)
//   - The last N messages (up to 10 most recent exchanges)
//   - Tool calls made and their outcomes (truncated)
//
// The output is capped at roughly 2K characters.
func (s *Store) GetSessionSummary(id string) (string, error) {
	sess, err := s.GetSession(id)
	if err != nil {
		return "", fmt.Errorf("get session summary: %w", err)
	}

	messages, err := s.GetMessages(id)
	if err != nil {
		return "", fmt.Errorf("get session summary: %w", err)
	}

	const maxSummaryMessages = 10

	var b strings.Builder

	// Header
	b.WriteString("Previous session: ")
	if sess.Provider != "" || sess.Model != "" {
		b.WriteString(fmt.Sprintf("%s/%s — ", sess.Provider, sess.Model))
	}
	b.WriteString(sess.LastActiveAt.Format("2006-01-02 15:04:05"))
	b.WriteString("\n")

	// What they were working on (first user message, truncated)
	if sess.FirstUserMsg != "" {
		b.WriteString("Topic: ")
		first := sess.FirstUserMsg
		if len(first) > 200 {
			first = first[:200] + "..."
		}
		b.WriteString(first)
		b.WriteString("\n\n")
	}

	// Last N messages (skip if there's only the first user msg shown above)
	start := 0
	if len(messages) > maxSummaryMessages {
		start = len(messages) - maxSummaryMessages
	}

	for i := start; i < len(messages); i++ {
		msg := messages[i]
		content := msg.Content
		// Truncate long messages
		if len(content) > 300 {
			content = content[:300] + "..."
		}

		roleLabel := "User"
		if msg.Role == "assistant" {
			roleLabel = "Assistant"
		}
		b.WriteString(fmt.Sprintf("%s: %s\n", roleLabel, content))

		// Include tool calls for assistant messages
		if msg.Role == "assistant" {
			calls, err := s.GetToolCalls(msg.ID)
			if err == nil && len(calls) > 0 {
				for _, tc := range calls {
					out := tc.Output
					if len(out) > 150 {
						out = out[:150] + "..."
					}
					b.WriteString(fmt.Sprintf("  [tool: %s → %s]\n", tc.ToolName, out))
				}
			}
		}
	}

	result := b.String()
	// Cap at ~2K chars
	if len(result) > 2048 {
		cut := strings.LastIndex(result[:2048], "\n")
		if cut < 0 {
			cut = 2000
		}
		result = result[:cut] + "\n...\n"
	}

	return result, nil
}

// SearchSessions performs a full-text search across message content using
// the FTS5 index. It returns sessions whose messages match the query,
// ordered by last_active_at DESC.
func (s *Store) SearchSessions(query string) ([]*Session, error) {
	rows, err := s.DB.Query(
		`SELECT DISTINCT s.id, s.created_at, s.last_active_at, s.provider, s.model,
		        s.message_count, s.first_user_msg, s.last_assistant_msg
		 FROM sessions s
		 JOIN messages m ON m.session_id = s.id
		 JOIN messages_fts fts ON fts.rowid = m.id
		 WHERE messages_fts MATCH ?
		 ORDER BY s.last_active_at DESC`, query,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		var sess Session
		var createdAt, lastActiveAt string
		if err := rows.Scan(&sess.ID, &createdAt, &lastActiveAt, &sess.Provider, &sess.Model,
			&sess.MessageCount, &sess.FirstUserMsg, &sess.LastAssistantMsg); err != nil {
			return nil, err
		}
		sess.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		sess.LastActiveAt, _ = time.Parse(time.RFC3339Nano, lastActiveAt)
		sessions = append(sessions, &sess)
	}
	if sessions == nil {
		sessions = []*Session{}
	}
	return sessions, rows.Err()
}

// FilterSessions returns sessions matching the given criteria.
// Empty strings and zero-time values are treated as "no filter".
func (s *Store) FilterSessions(provider, model string, after, before time.Time) ([]*Session, error) {
	query := `SELECT id, created_at, last_active_at, provider, model,
	                 message_count, first_user_msg, last_assistant_msg
	          FROM sessions WHERE 1=1`
	var args []any

	if provider != "" {
		query += ` AND provider = ?`
		args = append(args, provider)
	}
	if model != "" {
		query += ` AND model = ?`
		args = append(args, model)
	}
	if !after.IsZero() {
		query += ` AND last_active_at >= ?`
		args = append(args, after.UTC().Format(time.RFC3339Nano))
	}
	if !before.IsZero() {
		query += ` AND last_active_at <= ?`
		args = append(args, before.UTC().Format(time.RFC3339Nano))
	}
	query += ` ORDER BY last_active_at DESC`

	rows, err := s.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		var sess Session
		var createdAt, lastActiveAt string
		if err := rows.Scan(&sess.ID, &createdAt, &lastActiveAt, &sess.Provider, &sess.Model,
			&sess.MessageCount, &sess.FirstUserMsg, &sess.LastAssistantMsg); err != nil {
			return nil, err
		}
		sess.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		sess.LastActiveAt, _ = time.Parse(time.RFC3339Nano, lastActiveAt)
		sessions = append(sessions, &sess)
	}
	if sessions == nil {
		sessions = []*Session{}
	}
	return sessions, rows.Err()
}

// DeleteSession removes a session and all associated messages and
// tool_calls (via ON DELETE CASCADE).
func (s *Store) DeleteSession(id string) error {
	_, err := s.DB.Exec(`DELETE FROM sessions WHERE id = ?`, id)
	return err
}

// UpdateSessionPreview updates the session metadata after recording
// a message. Called after user messages (sets first_user_msg) and
// assistant replies (sets last_assistant_msg).
func (s *Store) UpdateSessionPreview(id, firstUserMsg, lastAssistantMsg string, msgCount int) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	query := `UPDATE sessions SET last_active_at = ?, message_count = ?`
	args := []any{now, msgCount}

	if firstUserMsg != "" {
		query += `, first_user_msg = CASE WHEN first_user_msg = '' THEN ? ELSE first_user_msg END`
		args = append(args, firstUserMsg)
	}
	if lastAssistantMsg != "" {
		query += `, last_assistant_msg = ?`
		args = append(args, lastAssistantMsg)
	}
	query += ` WHERE id = ?`
	args = append(args, id)
	_, err := s.DB.Exec(query, args...)
	return err
}

// ── Message CRUD ──────────────────────────────────────────────────

// RecordMessage inserts a new message for the given session. Returns
// the auto-generated message id. The session's last_active_at and
// message_count are updated.
func (s *Store) RecordMessage(sessionID, role, content string, tokensUsed int) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	res, err := s.DB.Exec(
		`INSERT INTO messages (session_id, role, content, created_at, tokens_used)
		 VALUES (?, ?, ?, ?, ?)`,
		sessionID, role, content, now, tokensUsed,
	)
	if err != nil {
		return 0, err
	}
	msgID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return msgID, nil
}

// GetMessages returns all messages for a session, ordered by
// created_at ascending (chronological order).
func (s *Store) GetMessages(sessionID string) ([]*Message, error) {
	rows, err := s.DB.Query(
		`SELECT id, session_id, role, content, created_at, tokens_used
		 FROM messages WHERE session_id = ? ORDER BY created_at ASC`, sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		var msg Message
		var createdAt string
		if err := rows.Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content,
			&createdAt, &msg.TokensUsed); err != nil {
			return nil, err
		}
		msg.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		messages = append(messages, &msg)
	}
	if messages == nil {
		messages = []*Message{}
	}
	return messages, rows.Err()
}

// ── Tool Call CRUD ────────────────────────────────────────────────

// RecordToolCall inserts a new tool_call row linked to a message.
func (s *Store) RecordToolCall(messageID int64, toolName, input, output string, durationMs int) error {
	_, err := s.DB.Exec(
		`INSERT INTO tool_calls (message_id, tool_name, input, output, duration_ms)
		 VALUES (?, ?, ?, ?, ?)`,
		messageID, toolName, input, output, durationMs,
	)
	return err
}

// GetToolCalls returns all tool calls for a given message.
func (s *Store) GetToolCalls(messageID int64) ([]*ToolCall, error) {
	rows, err := s.DB.Query(
		`SELECT id, message_id, tool_name, input, output, duration_ms
		 FROM tool_calls WHERE message_id = ? ORDER BY id ASC`, messageID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var calls []*ToolCall
	for rows.Next() {
		var tc ToolCall
		if err := rows.Scan(&tc.ID, &tc.MessageID, &tc.ToolName,
			&tc.Input, &tc.Output, &tc.DurationMs); err != nil {
			return nil, err
		}
		calls = append(calls, &tc)
	}
	if calls == nil {
		calls = []*ToolCall{}
	}
	return calls, rows.Err()
}

// ── Export ────────────────────────────────────────────────────────

// ExportSessionMarkdown returns the session with all messages and tool calls
// formatted as markdown. Format:
//
//	## Session: <id> (<provider>/<model>)
//	Created: <created_at>, Last active: <last_active_at>
//	Messages: <count>
//
//	### User
//	<message content>
//
//	### Assistant
//	<message content>
//
//	#### Tool Calls
//	- **<tool_name>** (<duration_ms>ms)
//	  Input: <input>
//	  Output: <output>
func (s *Store) ExportSessionMarkdown(id string) (string, error) {
	sess, err := s.GetSession(id)
	if err != nil {
		return "", fmt.Errorf("export markdown: %w", err)
	}
	messages, err := s.GetMessages(id)
	if err != nil {
		return "", fmt.Errorf("export markdown: %w", err)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "## Session: %s (%s/%s)\n", sess.ID, sess.Provider, sess.Model)
	fmt.Fprintf(&b, "Created: %s\n", sess.CreatedAt.Format(time.RFC3339))
	fmt.Fprintf(&b, "Last active: %s\n", sess.LastActiveAt.Format(time.RFC3339))
	fmt.Fprintf(&b, "Messages: %d\n\n", sess.MessageCount)

	for _, msg := range messages {
		role := "User"
		if msg.Role == "assistant" {
			role = "Assistant"
		}
		fmt.Fprintf(&b, "### %s\n", role)
		fmt.Fprintf(&b, "%s\n\n", msg.Content)

		if msg.Role == "assistant" {
			calls, err := s.GetToolCalls(msg.ID)
			if err != nil {
				return "", fmt.Errorf("export markdown: tool calls for msg %d: %w", msg.ID, err)
			}
			if len(calls) > 0 {
				fmt.Fprintf(&b, "#### Tool Calls\n")
				for _, tc := range calls {
					fmt.Fprintf(&b, "- **%s** (%dms)\n", tc.ToolName, tc.DurationMs)
					fmt.Fprintf(&b, "  Input: %s\n", tc.Input)
					fmt.Fprintf(&b, "  Output: %s\n", tc.Output)
				}
				fmt.Fprintf(&b, "\n")
			}
		}
	}
	return b.String(), nil
}

// ExportSessionJSON returns the full session tree as pretty-printed JSON.
// The tree includes session metadata and all messages with their tool calls
// nested inline.
func (s *Store) ExportSessionJSON(id string) ([]byte, error) {
	sess, err := s.GetSession(id)
	if err != nil {
		return nil, fmt.Errorf("export json: %w", err)
	}
	messages, err := s.GetMessages(id)
	if err != nil {
		return nil, fmt.Errorf("export json: %w", err)
	}

	exported := &ExportedSession{Session: sess}
	for _, msg := range messages {
		em := &ExportedMessage{
			ID:         msg.ID,
			Role:       msg.Role,
			Content:    msg.Content,
			CreatedAt:  msg.CreatedAt,
			TokensUsed: msg.TokensUsed,
		}
		if msg.Role == "assistant" {
			calls, err := s.GetToolCalls(msg.ID)
			if err != nil {
				return nil, fmt.Errorf("export json: tool calls for msg %d: %w", msg.ID, err)
			}
			if len(calls) > 0 {
				em.ToolCalls = calls
			}
		}
		exported.Messages = append(exported.Messages, em)
	}

	data, err := json.MarshalIndent(exported, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("export json: marshal: %w", err)
	}
	return data, nil
}

// ── Import ────────────────────────────────────────────────────────

// ImportSessionJSON parses a JSON ExportedSession, creates a new session
// with all messages and tool calls, and returns the new session id.
// The provider and model fields are required. A new session id is always
// generated; the original id is preserved in the session metadata.
func (s *Store) ImportSessionJSON(data []byte) (string, error) {
	var exported ExportedSession
	if err := json.Unmarshal(data, &exported); err != nil {
		return "", fmt.Errorf("import json: parse: %w", err)
	}
	if exported.Session == nil {
		return "", fmt.Errorf("import json: missing session field")
	}
	if exported.Session.Provider == "" || exported.Session.Model == "" {
		return "", fmt.Errorf("import json: provider and model are required")
	}

	newID := fmt.Sprintf("imported-%d", time.Now().UnixNano())

	tx, err := s.DB.Begin()
	if err != nil {
		return "", fmt.Errorf("import json: begin tx: %w", err)
	}
	defer tx.Rollback()

	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = tx.Exec(
		`INSERT INTO sessions (id, created_at, last_active_at, provider, model, message_count, first_user_msg, last_assistant_msg)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		newID, now, now, exported.Session.Provider, exported.Session.Model,
		len(exported.Messages), exported.Session.FirstUserMsg, exported.Session.LastAssistantMsg,
	)
	if err != nil {
		return "", fmt.Errorf("import json: create session: %w", err)
	}

	for _, em := range exported.Messages {
		res, err := tx.Exec(
			`INSERT INTO messages (session_id, role, content, created_at, tokens_used) VALUES (?, ?, ?, ?, ?)`,
			newID, em.Role, em.Content, now, em.TokensUsed,
		)
		if err != nil {
			return "", fmt.Errorf("import json: insert message: %w", err)
		}
		msgID, _ := res.LastInsertId()
		for _, tc := range em.ToolCalls {
			_, err := tx.Exec(
				`INSERT INTO tool_calls (message_id, tool_name, input, output, duration_ms) VALUES (?, ?, ?, ?, ?)`,
				msgID, tc.ToolName, tc.Input, tc.Output, tc.DurationMs,
			)
			if err != nil {
				return "", fmt.Errorf("import json: insert tool call: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("import json: commit: %w", err)
	}
	return newID, nil
}

// ImportSessionMarkdown parses the markdown export format and creates
// a new session with messages and tool calls. Returns the new session id.
//
// Expected format:
//
//	## Session: <id> (<provider>/<model>)
//	Created: ...
//	Last active: ...
//	Messages: ...
//
//	### User
//	message content
//
//	### Assistant
//	message content
//
//	#### Tool Calls
//	- **tool_name** (Nms)
//	  Input: ...
//	  Output: ...
func (s *Store) ImportSessionMarkdown(data []byte) (string, error) {
	content := string(data)
	lines := strings.Split(content, "\n")

	var sessionID, provider, model string
	var messages []struct {
		Role      string
		Content   string
		ToolCalls []*ToolCall
	}

	// Parse header
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "## Session: ") {
			// Format: ## Session: <id> (<provider>/<model>)
			header := strings.TrimPrefix(line, "## Session: ")
			parenOpen := strings.LastIndex(header, "(")
			if parenOpen > 0 {
				sessionID = strings.TrimSpace(header[:parenOpen])
				pm := strings.TrimRight(header[parenOpen+1:], ")")
				parts := strings.SplitN(pm, "/", 2)
				if len(parts) == 2 {
					provider = parts[0]
					model = parts[1]
				}
			} else {
				sessionID = strings.TrimSpace(header)
			}
			continue
		}
		if strings.HasPrefix(line, "### ") {
			// New message block
			role := strings.TrimPrefix(line, "### ")
			roleLower := "user"
			if strings.EqualFold(role, "assistant") {
				roleLower = "assistant"
			}
			// Collect content until next ### or #### or end
			var contentLines []string
			var toolCalls []*ToolCall
			i++
			for i < len(lines) {
				if strings.HasPrefix(lines[i], "### ") || strings.HasPrefix(lines[i], "## ") {
					i-- // back up so outer loop re-processes
					break
				}
				if strings.HasPrefix(lines[i], "#### Tool Calls") {
					i++
					for i < len(lines) {
						l := lines[i]
						if strings.HasPrefix(l, "### ") || strings.HasPrefix(l, "## ") || strings.HasPrefix(l, "#### ") && !strings.HasPrefix(l, "#### Tool Calls") {
							i-- // back up
							break
						}
						if strings.HasPrefix(l, "- **") {
							tc := parseToolCallLine(l, lines, &i)
							if tc != nil {
								toolCalls = append(toolCalls, tc)
							}
						} else {
							i++
						}
					}
					break
				}
				contentLines = append(contentLines, lines[i])
				i++
			}
			msgContent := strings.Join(contentLines, "\n")
			msgContent = strings.TrimSpace(msgContent)
			messages = append(messages, struct {
				Role      string
				Content   string
				ToolCalls []*ToolCall
			}{Role: roleLower, Content: msgContent, ToolCalls: toolCalls})
			continue
		}
	}

	if provider == "" {
		provider = "imported"
	}
	if model == "" {
		model = "unknown"
	}
	sessionID = fmt.Sprintf("imported-md-%d", time.Now().UnixNano())

	// Create session and messages in a transaction
	tx, err := s.DB.Begin()
	if err != nil {
		return "", fmt.Errorf("import markdown: begin tx: %w", err)
	}
	defer tx.Rollback()

	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = tx.Exec(
		`INSERT INTO sessions (id, created_at, last_active_at, provider, model, message_count)
		VALUES (?, ?, ?, ?, ?, ?)`,
		sessionID, now, now, provider, model, len(messages),
	)
	if err != nil {
		return "", fmt.Errorf("import markdown: create session: %w", err)
	}

	for _, msg := range messages {
		res, err := tx.Exec(
			`INSERT INTO messages (session_id, role, content, created_at, tokens_used) VALUES (?, ?, ?, ?, 0)`,
			sessionID, msg.Role, msg.Content, now,
		)
		if err != nil {
			return "", fmt.Errorf("import markdown: insert message: %w", err)
		}
		msgID, _ := res.LastInsertId()
		for _, tc := range msg.ToolCalls {
			_, err := tx.Exec(
				`INSERT INTO tool_calls (message_id, tool_name, input, output, duration_ms) VALUES (?, ?, ?, ?, ?)`,
				msgID, tc.ToolName, tc.Input, tc.Output, tc.DurationMs,
			)
			if err != nil {
				return "", fmt.Errorf("import markdown: insert tool call: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("import markdown: commit: %w", err)
	}
	return sessionID, nil
}

// parseToolCallLine parses "- **tool_name** (Nms)" and the following
// Input/Output lines, advancing the line index.
func parseToolCallLine(line string, lines []string, idx *int) *ToolCall {
	// Format: - **tool_name** (Nms)
	if !strings.HasPrefix(line, "- **") {
		return nil
	}
	line = strings.TrimPrefix(line, "- **")
	closeIdx := strings.Index(line, "**")
	if closeIdx < 0 {
		return nil
	}
	toolName := line[:closeIdx]
	rest := line[closeIdx+2:]
	rest = strings.TrimSpace(rest)

	var durationMs int
	if strings.HasPrefix(rest, "(") && strings.HasSuffix(rest, "ms)") {
		numStr := strings.TrimSuffix(strings.TrimPrefix(rest, "("), "ms)")
		durationMs, _ = strconv.Atoi(numStr)
	}

	tc := &ToolCall{ToolName: toolName, DurationMs: durationMs}

	// Read next lines for Input: and Output:
	for *idx+1 < len(lines) {
		*idx++
		l := strings.TrimSpace(lines[*idx])
		if strings.HasPrefix(l, "Input: ") {
			tc.Input = strings.TrimPrefix(l, "Input: ")
		} else if strings.HasPrefix(l, "Output: ") {
			tc.Output = strings.TrimPrefix(l, "Output: ")
			break
		} else if strings.HasPrefix(l, "- **") || strings.HasPrefix(l, "### ") || strings.HasPrefix(l, "## ") {
			*idx-- // back up
			break
		}
	}
	return tc
}

// migrate applies any pending schema migrations.
func (s *Store) migrate() error {
	// Ensure the version table exists (SchemaV1 does this, but
	// defensive code in case a migration failed partway).
	if _, err := s.DB.Exec(`CREATE TABLE IF NOT EXISTS schema_version (version INTEGER NOT NULL)`); err != nil {
		return fmt.Errorf("session store: create version table: %w", err)
	}

	var current int
	err := s.DB.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_version`).Scan(&current)
	if err != nil {
		return fmt.Errorf("session store: read schema version: %w", err)
	}

	for i := current; i < len(migrations); i++ {
		tx, err := s.DB.Begin()
		if err != nil {
			return fmt.Errorf("session store: begin tx for migration %d: %w", i+1, err)
		}
		if _, err := tx.Exec(migrations[i]); err != nil {
			tx.Rollback()
			return fmt.Errorf("session store: apply migration %d: %w", i+1, err)
		}
		if _, err := tx.Exec(`INSERT INTO schema_version (version) VALUES (?)`, i+1); err != nil {
			tx.Rollback()
			return fmt.Errorf("session store: record migration %d: %w", i+1, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("session store: commit migration %d: %w", i+1, err)
		}
	}
	return nil
}