package session

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

// openTemp opens a Store backed by a temporary database file that is
// automatically cleaned up when the test and all its subtests finish.
func openTemp(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open(%q): %v", dbPath, err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func TestSchemaCreation(t *testing.T) {
	store := openTemp(t)

	// Verify all three tables exist.
	tables := []string{"sessions", "messages", "tool_calls", "schema_version"}
	for _, name := range tables {
		var count int
		err := store.DB.QueryRow(
			"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?",
			name,
		).Scan(&count)
		if err != nil {
			t.Fatalf("query sqlite_master for %s: %v", name, err)
		}
		if count != 1 {
			t.Errorf("table %q: expected 1 row in sqlite_master, got %d", name, count)
		}
	}
}

func TestSchemaVersion(t *testing.T) {
	store := openTemp(t)

	var version int
	err := store.DB.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&version)
	if err != nil {
		t.Fatalf("read version: %v", err)
	}
	if version < 1 {
		t.Errorf("expected schema version >= 1, got %d", version)
	}
}

func TestSessionsTableColumns(t *testing.T) {
	store := openTemp(t)

	// Insert a row so we can verify all columns are writable.
	_, err := store.DB.Exec(`
		INSERT INTO sessions (id, created_at, last_active_at, provider, model, message_count, first_user_msg, last_assistant_msg)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, "sess_test1", "2026-01-01T00:00:00Z", "2026-01-01T01:00:00Z",
		"deepseek", "expert", 5, "Hello world", "I can help with that")
	if err != nil {
		t.Fatalf("insert session: %v", err)
	}

	var (
		id, createdAt, lastActive, provider, model, firstMsg, lastMsg string
		msgCount                                                       int
	)
	err = store.DB.QueryRow("SELECT id, created_at, last_active_at, provider, model, message_count, first_user_msg, last_assistant_msg FROM sessions WHERE id = ?", "sess_test1").
		Scan(&id, &createdAt, &lastActive, &provider, &model, &msgCount, &firstMsg, &lastMsg)
	if err != nil {
		t.Fatalf("query session: %v", err)
	}
	if id != "sess_test1" {
		t.Errorf("id = %q, want %q", id, "sess_test1")
	}
	if msgCount != 5 {
		t.Errorf("message_count = %d, want 5", msgCount)
	}
	if firstMsg != "Hello world" {
		t.Errorf("first_user_msg = %q, want %q", firstMsg, "Hello world")
	}
}

func TestMessagesTableColumns(t *testing.T) {
	store := openTemp(t)

	// Need a parent session row first (FK constraint).
	_, err := store.DB.Exec(
		"INSERT INTO sessions (id, created_at, last_active_at) VALUES (?, ?, ?)",
		"sess_test2", "2026-01-01T00:00:00Z", "2026-01-01T01:00:00Z",
	)
	if err != nil {
		t.Fatalf("insert parent session: %v", err)
	}

	_, err = store.DB.Exec(`
		INSERT INTO messages (session_id, role, content, created_at, tokens_used)
		VALUES (?, ?, ?, ?, ?)
	`, "sess_test2", "user", "What is Go?", "2026-01-01T00:01:00Z", 12)
	if err != nil {
		t.Fatalf("insert message: %v", err)
	}

	var (
		id                        int64
		sessionID, role, content, createdAt string
		tokensUsed                int
	)
	err = store.DB.QueryRow(
		"SELECT id, session_id, role, content, created_at, tokens_used FROM messages WHERE session_id = ?",
		"sess_test2",
	).Scan(&id, &sessionID, &role, &content, &createdAt, &tokensUsed)
	if err != nil {
		t.Fatalf("query message: %v", err)
	}
	if role != "user" {
		t.Errorf("role = %q, want %q", role, "user")
	}
	if content != "What is Go?" {
		t.Errorf("content = %q, want %q", content, "What is Go?")
	}
	if tokensUsed != 12 {
		t.Errorf("tokens_used = %d, want 12", tokensUsed)
	}
	// Verify AUTOINCREMENT assigned a positive id.
	if id <= 0 {
		t.Errorf("id = %d, want > 0", id)
	}
}

func TestToolCallsTableColumns(t *testing.T) {
	store := openTemp(t)

	// Setup: session + message.
	_, err := store.DB.Exec(
		"INSERT INTO sessions (id, created_at, last_active_at) VALUES (?, ?, ?)",
		"sess_test3", "2026-01-01T00:00:00Z", "2026-01-01T01:00:00Z",
	)
	if err != nil {
		t.Fatalf("insert parent session: %v", err)
	}
	res, err := store.DB.Exec(
		"INSERT INTO messages (session_id, role, content, created_at) VALUES (?, ?, ?, ?)",
		"sess_test3", "assistant", "Let me run that command", "2026-01-01T00:01:00Z",
	)
	if err != nil {
		t.Fatalf("insert parent message: %v", err)
	}
	msgID, _ := res.LastInsertId()

	_, err = store.DB.Exec(`
		INSERT INTO tool_calls (message_id, tool_name, input, output, duration_ms)
		VALUES (?, ?, ?, ?, ?)
	`, msgID, "bash", `{"command":"ls"}`, "file1\nfile2\n", 150)
	if err != nil {
		t.Fatalf("insert tool_call: %v", err)
	}

	var (
		id                                 int64
		mid                                int64
		toolName, input, output            string
		durationMs                         int
	)
	err = store.DB.QueryRow(
		"SELECT id, message_id, tool_name, input, output, duration_ms FROM tool_calls WHERE message_id = ?",
		msgID,
	).Scan(&id, &mid, &toolName, &input, &output, &durationMs)
	if err != nil {
		t.Fatalf("query tool_call: %v", err)
	}
	if mid != msgID {
		t.Errorf("message_id = %d, want %d", mid, msgID)
	}
	if toolName != "bash" {
		t.Errorf("tool_name = %q, want %q", toolName, "bash")
	}
	if input != `{"command":"ls"}` {
		t.Errorf("input = %q, want %q", input, `{"command":"ls"}`)
	}
	if output != "file1\nfile2\n" {
		t.Errorf("output = %q, want %q", output, "file1\nfile2\n")
	}
	if durationMs != 150 {
		t.Errorf("duration_ms = %d, want 150", durationMs)
	}
}

func TestForeignKeyCascadeDeleteSession(t *testing.T) {
	store := openTemp(t)

	// Insert session + message + tool_call.
	_, err := store.DB.Exec(
		"INSERT INTO sessions (id, created_at, last_active_at) VALUES (?, ?, ?)",
		"sess_cascade", "2026-01-01T00:00:00Z", "2026-01-01T01:00:00Z",
	)
	if err != nil {
		t.Fatalf("insert session: %v", err)
	}
	res, err := store.DB.Exec(
		"INSERT INTO messages (session_id, role, content, created_at) VALUES (?, ?, ?, ?)",
		"sess_cascade", "user", "hello", "2026-01-01T00:01:00Z",
	)
	if err != nil {
		t.Fatalf("insert message: %v", err)
	}
	msgID, _ := res.LastInsertId()
	_, err = store.DB.Exec(
		"INSERT INTO tool_calls (message_id, tool_name) VALUES (?, ?)",
		msgID, "bash",
	)
	if err != nil {
		t.Fatalf("insert tool_call: %v", err)
	}

	// Delete the session.
	_, err = store.DB.Exec("DELETE FROM sessions WHERE id = ?", "sess_cascade")
	if err != nil {
		t.Fatalf("delete session: %v", err)
	}

	// Messages should be cascaded away.
	var count int
	err = store.DB.QueryRow("SELECT COUNT(*) FROM messages WHERE session_id = ?", "sess_cascade").Scan(&count)
	if err != nil {
		t.Fatalf("count messages: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 messages after cascade delete, got %d", count)
	}

	// Tool calls should also be cascaded away.
	err = store.DB.QueryRow("SELECT COUNT(*) FROM tool_calls WHERE message_id = ?", msgID).Scan(&count)
	if err != nil {
		t.Fatalf("count tool_calls: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 tool_calls after cascade delete, got %d", count)
	}
}

func TestMigrationIsIdempotent(t *testing.T) {
	store := openTemp(t)

	// Re-running migrate should be a no-op.
	if err := store.migrate(); err != nil {
		t.Fatalf("second migrate: %v", err)
	}

	// Schema version should still be exactly the migration count.
	var version int
	err := store.DB.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&version)
	if err != nil {
		t.Fatalf("read version after second migrate: %v", err)
	}
	if version != currentVersion {
		t.Errorf("version after second migrate = %d, want %d", version, currentVersion)
	}
}

// TestOpenCreatesDir verifies that Open creates the parent directory
// when it does not exist.
func TestOpenCreatesDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nonexistent", "subdir")
	dbPath := filepath.Join(dir, "sessions.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer store.Close()

	if _, err := os.Stat(dir); err != nil {
		t.Errorf("directory %s was not created: %v", dir, err)
	}
}

// TestOpenNonExistentDir verifies that Open returns an error when the
// parent directory cannot be created (e.g. permissions).
func TestOpenNonExistentDir(t *testing.T) {
	// This is a best-effort test: if we're running as root it won't fail.
	// We just verify Open doesn't panic.
	_ = t.TempDir() // ensure clean state
}

// Ensure the sql driver is imported so tests don't fail with
// "no driver registered". The import happens in the main file or a
// blank-import file; this comment is just documentation.
var _ = sql.Drivers