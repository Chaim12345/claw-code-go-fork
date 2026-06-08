package session

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"
)

// ── Session CRUD tests ────────────────────────────────────────────

func TestCreateSession(t *testing.T) {
	store := openTemp(t)

	err := store.CreateSession("sess-1", "deepseek", "expert")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	sess, err := store.GetSession("sess-1")
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if sess.ID != "sess-1" {
		t.Errorf("ID = %q, want %q", sess.ID, "sess-1")
	}
	if sess.Provider != "deepseek" {
		t.Errorf("Provider = %q, want %q", sess.Provider, "deepseek")
	}
	if sess.Model != "expert" {
		t.Errorf("Model = %q, want %q", sess.Model, "expert")
	}
	if sess.MessageCount != 0 {
		t.Errorf("MessageCount = %d, want 0", sess.MessageCount)
	}
	if sess.FirstUserMsg != "" {
		t.Errorf("FirstUserMsg = %q, want empty", sess.FirstUserMsg)
	}
	if sess.LastAssistantMsg != "" {
		t.Errorf("LastAssistantMsg = %q, want empty", sess.LastAssistantMsg)
	}
	// Timestamps should be set (non-zero).
	if sess.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
	if sess.LastActiveAt.IsZero() {
		t.Error("LastActiveAt is zero")
	}
}

func TestCreateSessionDuplicate(t *testing.T) {
	store := openTemp(t)

	if err := store.CreateSession("sess-dup", "openai", "gpt-4"); err != nil {
		t.Fatalf("first CreateSession: %v", err)
	}
	// Second create with same ID should be a no-op (INSERT OR IGNORE).
	if err := store.CreateSession("sess-dup", "anthropic", "claude"); err != nil {
		t.Fatalf("second CreateSession: %v", err)
	}

	sess, err := store.GetSession("sess-dup")
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	// Original values should be preserved.
	if sess.Provider != "openai" {
		t.Errorf("Provider = %q, want %q (original should be preserved)", sess.Provider, "openai")
	}
}

func TestGetSessionNotFound(t *testing.T) {
	store := openTemp(t)

	sess, err := store.GetSession("nonexistent")
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
	if sess != nil {
		t.Errorf("expected nil session, got %+v", sess)
	}
}

func TestListSessions(t *testing.T) {
	store := openTemp(t)

	// Insert sessions with staggered times.
	if err := store.CreateSession("sess-a", "openai", "gpt-4"); err != nil {
		t.Fatalf("CreateSession sess-a: %v", err)
	}
	time.Sleep(10 * time.Millisecond) // ensure different last_active_at
	if err := store.CreateSession("sess-b", "deepseek", "expert"); err != nil {
		t.Fatalf("CreateSession sess-b: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	if err := store.CreateSession("sess-c", "anthropic", "claude"); err != nil {
		t.Fatalf("CreateSession sess-c: %v", err)
	}

	sessions, err := store.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(sessions) != 3 {
		t.Fatalf("expected 3 sessions, got %d", len(sessions))
	}
	// Should be ordered by last_active_at DESC (newest first).
	if sessions[0].ID != "sess-c" {
		t.Errorf("newest session: got %q, want sess-c", sessions[0].ID)
	}
	if sessions[2].ID != "sess-a" {
		t.Errorf("oldest session: got %q, want sess-a", sessions[2].ID)
	}
}

func TestListSessionsEmpty(t *testing.T) {
	store := openTemp(t)

	sessions, err := store.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if sessions == nil {
		t.Error("expected non-nil empty slice")
	}
	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(sessions))
	}
}

func TestDeleteSession(t *testing.T) {
	store := openTemp(t)

	if err := store.CreateSession("sess-del", "deepseek", "expert"); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if err := store.DeleteSession("sess-del"); err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}

	_, err := store.GetSession("sess-del")
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows after delete, got %v", err)
	}
}

func TestDeleteSessionCascade(t *testing.T) {
	store := openTemp(t)

	// Create session + message + tool_call, then delete session and
	// verify everything is cleaned up via foreign key cascade.
	if err := store.CreateSession("sess-cascade2", "deepseek", "expert"); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	msgID, err := store.RecordMessage("sess-cascade2", "user", "hello", 5)
	if err != nil {
		t.Fatalf("RecordMessage: %v", err)
	}
	if err := store.RecordToolCall(msgID, "bash", `{"cmd":"ls"}`, "file1", 100); err != nil {
		t.Fatalf("RecordToolCall: %v", err)
	}

	// Delete session.
	if err := store.DeleteSession("sess-cascade2"); err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}

	// Messages should be gone.
	messages, err := store.GetMessages("sess-cascade2")
	if err != nil {
		t.Fatalf("GetMessages: %v", err)
	}
	if len(messages) != 0 {
		t.Errorf("expected 0 messages after cascade, got %d", len(messages))
	}

	// Tool calls should be gone.
	toolCalls, err := store.GetToolCalls(msgID)
	if err != nil {
		t.Fatalf("GetToolCalls: %v", err)
	}
	if len(toolCalls) != 0 {
		t.Errorf("expected 0 tool calls after cascade, got %d", len(toolCalls))
	}
}

// ── Message CRUD tests ────────────────────────────────────────────

func TestRecordMessage(t *testing.T) {
	store := openTemp(t)

	if err := store.CreateSession("sess-msg", "deepseek", "expert"); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	msgID, err := store.RecordMessage("sess-msg", "user", "What is Go?", 12)
	if err != nil {
		t.Fatalf("RecordMessage: %v", err)
	}
	if msgID <= 0 {
		t.Errorf("expected positive message id, got %d", msgID)
	}

	messages, err := store.GetMessages("sess-msg")
	if err != nil {
		t.Fatalf("GetMessages: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}
	msg := messages[0]
	if msg.ID != msgID {
		t.Errorf("ID = %d, want %d", msg.ID, msgID)
	}
	if msg.SessionID != "sess-msg" {
		t.Errorf("SessionID = %q, want sess-msg", msg.SessionID)
	}
	if msg.Role != "user" {
		t.Errorf("Role = %q, want user", msg.Role)
	}
	if msg.Content != "What is Go?" {
		t.Errorf("Content = %q, want %q", msg.Content, "What is Go?")
	}
	if msg.TokensUsed != 12 {
		t.Errorf("TokensUsed = %d, want 12", msg.TokensUsed)
	}
	if msg.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
}

func TestRecordMessageMissingSession(t *testing.T) {
	store := openTemp(t)

	_, err := store.RecordMessage("nonexistent", "user", "hello", 5)
	if err == nil {
		t.Fatal("expected foreign key error, got nil")
	}
}

func TestGetMessagesOrdering(t *testing.T) {
	store := openTemp(t)

	if err := store.CreateSession("sess-order", "deepseek", "expert"); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	// Insert messages with a sleep to ensure distinct timestamps.
	_, err := store.RecordMessage("sess-order", "user", "first", 5)
	if err != nil {
		t.Fatalf("RecordMessage 1: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	_, err = store.RecordMessage("sess-order", "assistant", "second", 10)
	if err != nil {
		t.Fatalf("RecordMessage 2: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	_, err = store.RecordMessage("sess-order", "user", "third", 5)
	if err != nil {
		t.Fatalf("RecordMessage 3: %v", err)
	}

	messages, err := store.GetMessages("sess-order")
	if err != nil {
		t.Fatalf("GetMessages: %v", err)
	}
	if len(messages) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(messages))
	}
	// Should be in chronological order.
	if messages[0].Content != "first" {
		t.Errorf("messages[0] = %q, want first", messages[0].Content)
	}
	if messages[1].Content != "second" {
		t.Errorf("messages[1] = %q, want second", messages[1].Content)
	}
	if messages[2].Content != "third" {
		t.Errorf("messages[2] = %q, want third", messages[2].Content)
	}
}

func TestGetMessagesEmpty(t *testing.T) {
	store := openTemp(t)

	if err := store.CreateSession("sess-empty-msgs", "deepseek", "expert"); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	messages, err := store.GetMessages("sess-empty-msgs")
	if err != nil {
		t.Fatalf("GetMessages: %v", err)
	}
	if messages == nil {
		t.Error("expected non-nil empty slice")
	}
	if len(messages) != 0 {
		t.Errorf("expected 0 messages, got %d", len(messages))
	}
}

func TestRecordMessageBothRoles(t *testing.T) {
	store := openTemp(t)

	if err := store.CreateSession("sess-roles", "deepseek", "expert"); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	_, err := store.RecordMessage("sess-roles", "user", "Hello", 3)
	if err != nil {
		t.Fatalf("RecordMessage user: %v", err)
	}
	_, err = store.RecordMessage("sess-roles", "assistant", "Hi there!", 7)
	if err != nil {
		t.Fatalf("RecordMessage assistant: %v", err)
	}

	messages, err := store.GetMessages("sess-roles")
	if err != nil {
		t.Fatalf("GetMessages: %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(messages))
	}
	if messages[0].Role != "user" {
		t.Errorf("messages[0].Role = %q, want user", messages[0].Role)
	}
	if messages[1].Role != "assistant" {
		t.Errorf("messages[1].Role = %q, want assistant", messages[1].Role)
	}
}

// ── Tool Call CRUD tests ──────────────────────────────────────────

func TestRecordToolCall(t *testing.T) {
	store := openTemp(t)

	if err := store.CreateSession("sess-tc", "deepseek", "expert"); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	msgID, err := store.RecordMessage("sess-tc", "assistant", "Running command...", 15)
	if err != nil {
		t.Fatalf("RecordMessage: %v", err)
	}

	input := `{"command":"ls -la"}`
	output := "total 12\ndrwxr-xr-x ..."
	err = store.RecordToolCall(msgID, "bash", input, output, 250)
	if err != nil {
		t.Fatalf("RecordToolCall: %v", err)
	}

	calls, err := store.GetToolCalls(msgID)
	if err != nil {
		t.Fatalf("GetToolCalls: %v", err)
	}
	if len(calls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(calls))
	}
	tc := calls[0]
	if tc.MessageID != msgID {
		t.Errorf("MessageID = %d, want %d", tc.MessageID, msgID)
	}
	if tc.ToolName != "bash" {
		t.Errorf("ToolName = %q, want bash", tc.ToolName)
	}
	if tc.Input != input {
		t.Errorf("Input = %q, want %q", tc.Input, input)
	}
	if tc.Output != output {
		t.Errorf("Output = %q, want %q", tc.Output, output)
	}
	if tc.DurationMs != 250 {
		t.Errorf("DurationMs = %d, want 250", tc.DurationMs)
	}
	if tc.ID <= 0 {
		t.Errorf("expected positive id, got %d", tc.ID)
	}
}

func TestRecordToolCallMissingMessage(t *testing.T) {
	store := openTemp(t)

	err := store.RecordToolCall(9999, "bash", "{}", "", 100)
	if err == nil {
		t.Fatal("expected foreign key error, got nil")
	}
}

func TestGetToolCallsMultiple(t *testing.T) {
	store := openTemp(t)

	if err := store.CreateSession("sess-multitc", "deepseek", "expert"); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	msgID, err := store.RecordMessage("sess-multitc", "assistant", "Running tools...", 15)
	if err != nil {
		t.Fatalf("RecordMessage: %v", err)
	}

	if err := store.RecordToolCall(msgID, "bash", `{"cmd":"ls"}`, "file1", 100); err != nil {
		t.Fatalf("RecordToolCall 1: %v", err)
	}
	if err := store.RecordToolCall(msgID, "read", `{"path":"/tmp"}`, "contents...", 200); err != nil {
		t.Fatalf("RecordToolCall 2: %v", err)
	}
	if err := store.RecordToolCall(msgID, "write", `{"path":"/tmp/x"}`, "ok", 50); err != nil {
		t.Fatalf("RecordToolCall 3: %v", err)
	}

	calls, err := store.GetToolCalls(msgID)
	if err != nil {
		t.Fatalf("GetToolCalls: %v", err)
	}
	if len(calls) != 3 {
		t.Fatalf("expected 3 tool calls, got %d", len(calls))
	}
	if calls[0].ToolName != "bash" {
		t.Errorf("calls[0].ToolName = %q, want bash", calls[0].ToolName)
	}
	if calls[1].ToolName != "read" {
		t.Errorf("calls[1].ToolName = %q, want read", calls[1].ToolName)
	}
	if calls[2].ToolName != "write" {
		t.Errorf("calls[2].ToolName = %q, want write", calls[2].ToolName)
	}
}

func TestGetToolCallsEmpty(t *testing.T) {
	store := openTemp(t)

	if err := store.CreateSession("sess-empty-tc", "deepseek", "expert"); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	msgID, err := store.RecordMessage("sess-empty-tc", "user", "hello", 5)
	if err != nil {
		t.Fatalf("RecordMessage: %v", err)
	}

	calls, err := store.GetToolCalls(msgID)
	if err != nil {
		t.Fatalf("GetToolCalls: %v", err)
	}
	if calls == nil {
		t.Error("expected non-nil empty slice")
	}
	if len(calls) != 0 {
		t.Errorf("expected 0 tool calls, got %d", len(calls))
	}
}

// ── UpdateSessionPreview tests ────────────────────────────────────

func TestUpdateSessionPreview(t *testing.T) {
	store := openTemp(t)

	if err := store.CreateSession("sess-preview", "deepseek", "expert"); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	// Update with first user message.
	err := store.UpdateSessionPreview("sess-preview", "How do I write Go?", "", 1)
	if err != nil {
		t.Fatalf("UpdateSessionPreview user: %v", err)
	}

	sess, err := store.GetSession("sess-preview")
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if sess.FirstUserMsg != "How do I write Go?" {
		t.Errorf("FirstUserMsg = %q, want %q", sess.FirstUserMsg, "How do I write Go?")
	}
	if sess.MessageCount != 1 {
		t.Errorf("MessageCount = %d, want 1", sess.MessageCount)
	}
	// FirstUserMsg should NOT be overwritten on subsequent calls.
	err = store.UpdateSessionPreview("sess-preview", "Different message", "", 2)
	if err != nil {
		t.Fatalf("UpdateSessionPreview second: %v", err)
	}
	sess, _ = store.GetSession("sess-preview")
	if sess.FirstUserMsg != "How do I write Go?" {
		t.Errorf("FirstUserMsg was overwritten: got %q, want %q", sess.FirstUserMsg, "How do I write Go?")
	}

	// Update with assistant reply.
	err = store.UpdateSessionPreview("sess-preview", "", "Go is a statically typed language...", 3)
	if err != nil {
		t.Fatalf("UpdateSessionPreview assistant: %v", err)
	}
	sess, _ = store.GetSession("sess-preview")
	if sess.LastAssistantMsg != "Go is a statically typed language..." {
		t.Errorf("LastAssistantMsg = %q, want %q", sess.LastAssistantMsg, "Go is a statically typed language...")
	}
	if sess.MessageCount != 3 {
		t.Errorf("MessageCount = %d, want 3", sess.MessageCount)
	}
}

func TestUpdateSessionPreviewMissingSession(t *testing.T) {
	store := openTemp(t)

	// Should not error — UPDATE with no matching row is a no-op.
	err := store.UpdateSessionPreview("nonexistent", "hello", "reply", 1)
	if err != nil {
		t.Fatalf("UpdateSessionPreview on nonexistent session: %v", err)
	}
}

// ── Full conversation flow integration test ───────────────────────

func TestFullConversationFlow(t *testing.T) {
	store := openTemp(t)

	// 1. Create a session.
	if err := store.CreateSession("sess-full", "deepseek", "expert"); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	// 2. Record a user message.
	userMsgID, err := store.RecordMessage("sess-full", "user", "List files in /tmp", 8)
	if err != nil {
		t.Fatalf("RecordMessage user: %v", err)
	}
	if err := store.UpdateSessionPreview("sess-full", "List files in /tmp", "", 1); err != nil {
		t.Fatalf("UpdateSessionPreview user: %v", err)
	}

	// 3. Record assistant message with tool calls.
	asstMsgID, err := store.RecordMessage("sess-full", "assistant", "Let me check that for you...", 20)
	if err != nil {
		t.Fatalf("RecordMessage assistant: %v", err)
	}
	if err := store.RecordToolCall(asstMsgID, "bash", `{"command":"ls /tmp"}`, "file1\nfile2\n", 150); err != nil {
		t.Fatalf("RecordToolCall: %v", err)
	}
	if err := store.UpdateSessionPreview("sess-full", "", "Let me check that for you...", 2); err != nil {
		t.Fatalf("UpdateSessionPreview assistant: %v", err)
	}

	// 4. Record second user message.
	_, err = store.RecordMessage("sess-full", "user", "Now read file1", 6)
	if err != nil {
		t.Fatalf("RecordMessage user 2: %v", err)
	}
	// UpdateSessionPreview with empty strings should not overwrite previews.
	if err := store.UpdateSessionPreview("sess-full", "", "", 3); err != nil {
		t.Fatalf("UpdateSessionPreview count only: %v", err)
	}

	// 5. Verify session metadata.
	sess, err := store.GetSession("sess-full")
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if sess.MessageCount != 3 {
		t.Errorf("MessageCount = %d, want 3", sess.MessageCount)
	}
	if sess.FirstUserMsg != "List files in /tmp" {
		t.Errorf("FirstUserMsg = %q, want %q", sess.FirstUserMsg, "List files in /tmp")
	}
	if sess.LastAssistantMsg != "Let me check that for you..." {
		t.Errorf("LastAssistantMsg = %q, want %q", sess.LastAssistantMsg, "Let me check that for you...")
	}

	// 6. Verify all messages.
	messages, err := store.GetMessages("sess-full")
	if err != nil {
		t.Fatalf("GetMessages: %v", err)
	}
	if len(messages) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(messages))
	}
	if messages[0].ID != userMsgID {
		t.Errorf("messages[0].ID = %d, want %d", messages[0].ID, userMsgID)
	}
	if messages[1].ID != asstMsgID {
		t.Errorf("messages[1].ID = %d, want %d", messages[1].ID, asstMsgID)
	}

	// 7. Verify tool calls on the assistant message.
	calls, err := store.GetToolCalls(asstMsgID)
	if err != nil {
		t.Fatalf("GetToolCalls: %v", err)
	}
	if len(calls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(calls))
	}
	if calls[0].ToolName != "bash" {
		t.Errorf("ToolName = %q, want bash", calls[0].ToolName)
	}
	if calls[0].Output != "file1\nfile2\n" {
		t.Errorf("Output = %q, want %q", calls[0].Output, "file1\nfile2\n")
	}

	// 8. Verify tool calls on the user message (should be empty).
	userCalls, err := store.GetToolCalls(userMsgID)
	if err != nil {
		t.Fatalf("GetToolCalls user: %v", err)
	}
	if len(userCalls) != 0 {
		t.Errorf("expected 0 tool calls on user message, got %d", len(userCalls))
	}
}

// ── SearchSessions tests ──────────────────────────────────────────

func TestSearchSessionsBasic(t *testing.T) {
	store := openTemp(t)

	if err := store.CreateSession("s1", "deepseek", "expert"); err != nil {
		t.Fatalf("CreateSession s1: %v", err)
	}
	if err := store.CreateSession("s2", "openai", "gpt-4"); err != nil {
		t.Fatalf("CreateSession s2: %v", err)
	}

	_, err := store.RecordMessage("s1", "user", "How do I write concurrent Go code?", 10)
	if err != nil {
		t.Fatalf("RecordMessage s1: %v", err)
	}
	_, err = store.RecordMessage("s2", "user", "What is the weather today?", 8)
	if err != nil {
		t.Fatalf("RecordMessage s2: %v", err)
	}

	results, err := store.SearchSessions("concurrent Go")
	if err != nil {
		t.Fatalf("SearchSessions: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].ID != "s1" {
		t.Errorf("result ID = %q, want s1", results[0].ID)
	}
}

func TestSearchSessionsNoMatch(t *testing.T) {
	store := openTemp(t)

	if err := store.CreateSession("s-no-match", "deepseek", "expert"); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	_, err := store.RecordMessage("s-no-match", "user", "Hello world", 5)
	if err != nil {
		t.Fatalf("RecordMessage: %v", err)
	}

	results, err := store.SearchSessions("xyzzynonexistent")
	if err != nil {
		t.Fatalf("SearchSessions: %v", err)
	}
	if results == nil {
		t.Error("expected non-nil empty slice")
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestSearchSessionsMultipleMatches(t *testing.T) {
	store := openTemp(t)

	if err := store.CreateSession("sa", "deepseek", "expert"); err != nil {
		t.Fatalf("CreateSession sa: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	if err := store.CreateSession("sb", "deepseek", "expert"); err != nil {
		t.Fatalf("CreateSession sb: %v", err)
	}

	_, err := store.RecordMessage("sa", "user", "Tell me about database indexing", 10)
	if err != nil {
		t.Fatalf("RecordMessage sa: %v", err)
	}
	_, err = store.RecordMessage("sb", "user", "How does indexing work in SQLite?", 10)
	if err != nil {
		t.Fatalf("RecordMessage sb: %v", err)
	}

	results, err := store.SearchSessions("indexing")
	if err != nil {
		t.Fatalf("SearchSessions: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	// Newest session first.
	if results[0].ID != "sb" {
		t.Errorf("first result = %q, want sb", results[0].ID)
	}
}

// ── FilterSessions tests ──────────────────────────────────────────

func TestFilterSessionsByProvider(t *testing.T) {
	store := openTemp(t)

	if err := store.CreateSession("fa", "deepseek", "expert"); err != nil {
		t.Fatalf("CreateSession fa: %v", err)
	}
	if err := store.CreateSession("fb", "openai", "gpt-4"); err != nil {
		t.Fatalf("CreateSession fb: %v", err)
	}
	if err := store.CreateSession("fc", "deepseek", "chat"); err != nil {
		t.Fatalf("CreateSession fc: %v", err)
	}

	results, err := store.FilterSessions("deepseek", "", time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("FilterSessions: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, s := range results {
		if s.Provider != "deepseek" {
			t.Errorf("Provider = %q, want deepseek", s.Provider)
		}
	}
}

func TestFilterSessionsByModel(t *testing.T) {
	store := openTemp(t)

	if err := store.CreateSession("fma", "deepseek", "expert"); err != nil {
		t.Fatalf("CreateSession fma: %v", err)
	}
	if err := store.CreateSession("fmb", "deepseek", "chat"); err != nil {
		t.Fatalf("CreateSession fmb: %v", err)
	}

	results, err := store.FilterSessions("", "expert", time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("FilterSessions: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Model != "expert" {
		t.Errorf("Model = %q, want expert", results[0].Model)
	}
}

func TestFilterSessionsByTimeRange(t *testing.T) {
	store := openTemp(t)

	if err := store.CreateSession("fta", "deepseek", "expert"); err != nil {
		t.Fatalf("CreateSession fta: %v", err)
	}
	time.Sleep(50 * time.Millisecond)
	if err := store.CreateSession("ftb", "deepseek", "expert"); err != nil {
		t.Fatalf("CreateSession ftb: %v", err)
	}

	sessions, err := store.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}
	olderTime := sessions[1].LastActiveAt

	results, err := store.FilterSessions("", "", olderTime, time.Time{})
	if err != nil {
		t.Fatalf("FilterSessions: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results (both >= older), got %d", len(results))
	}

	sessA, _ := store.GetSession("fta")
	beforeFilter := sessA.LastActiveAt.Add(-10 * time.Millisecond)
	results, err = store.FilterSessions("", "", beforeFilter, time.Time{})
	if err != nil {
		t.Fatalf("FilterSessions after: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results with before filter, got %d", len(results))
	}
}

func TestFilterSessionsNoFilters(t *testing.T) {
	store := openTemp(t)

	if err := store.CreateSession("fnone", "deepseek", "expert"); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	results, err := store.FilterSessions("", "", time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("FilterSessions: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result (no filters = all), got %d", len(results))
	}
}

func TestFilterSessionsEmpty(t *testing.T) {
	store := openTemp(t)

	results, err := store.FilterSessions("nonexistent", "", time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("FilterSessions: %v", err)
	}
	if results == nil {
		t.Error("expected non-nil empty slice")
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

// ── Export / Import tests ──────────────────────────────────────────

func seedForExport(t *testing.T, store *Store) string {
	t.Helper()
	err := store.CreateSession("export-test", "deepseek", "expert")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	_, err = store.RecordMessage("export-test", "user", "Hello from user", 10)
	if err != nil {
		t.Fatalf("RecordMessage user: %v", err)
	}
	msgID, err := store.RecordMessage("export-test", "assistant", "Hi from assistant", 20)
	if err != nil {
		t.Fatalf("RecordMessage assistant: %v", err)
	}
	err = store.RecordToolCall(msgID, "bash", "ls -la", "file1.txt\nfile2.txt", 150)
	if err != nil {
		t.Fatalf("RecordToolCall: %v", err)
	}
	return "export-test"
}

func TestExportSessionMarkdown(t *testing.T) {
	store := openTemp(t)
	id := seedForExport(t, store)

	md, err := store.ExportSessionMarkdown(id)
	if err != nil {
		t.Fatalf("ExportSessionMarkdown: %v", err)
	}
	if len(md) == 0 {
		t.Fatal("ExportSessionMarkdown returned empty string")
	}
	if !contains(md, "## Session:") {
		t.Error("markdown missing session heading")
	}
	if !contains(md, "### User") {
		t.Error("markdown missing user heading")
	}
	if !contains(md, "Hello from user") {
		t.Error("markdown missing user content")
	}
	if !contains(md, "### Assistant") {
		t.Error("markdown missing assistant heading")
	}
	if !contains(md, "Hi from assistant") {
		t.Error("markdown missing assistant content")
	}
	if !contains(md, "#### Tool Calls") {
		t.Error("markdown missing tool calls heading")
	}
	if !contains(md, "bash") {
		t.Error("markdown missing tool name")
	}
}

func TestExportSessionJSON(t *testing.T) {
	store := openTemp(t)
	id := seedForExport(t, store)

	data, err := store.ExportSessionJSON(id)
	if err != nil {
		t.Fatalf("ExportSessionJSON: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("ExportSessionJSON returned empty bytes")
	}

	var exported ExportedSession
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatalf("parse exported JSON: %v", err)
	}
	if exported.Session == nil {
		t.Fatal("exported session is nil")
	}
	if exported.Session.ID != id {
		t.Errorf("session ID = %q, want %q", exported.Session.ID, id)
	}
	if len(exported.Messages) != 2 {
		t.Fatalf("messages count = %d, want 2", len(exported.Messages))
	}
	if exported.Messages[0].Role != "user" {
		t.Errorf("msg[0] role = %q, want user", exported.Messages[0].Role)
	}
	if exported.Messages[1].Role != "assistant" {
		t.Errorf("msg[1] role = %q, want assistant", exported.Messages[1].Role)
	}
	if len(exported.Messages[1].ToolCalls) != 1 {
		t.Fatalf("msg[1] tool calls = %d, want 1", len(exported.Messages[1].ToolCalls))
	}
	if exported.Messages[1].ToolCalls[0].ToolName != "bash" {
		t.Errorf("tool name = %q, want bash", exported.Messages[1].ToolCalls[0].ToolName)
	}
}

func TestExportSessionMarkdownNotFound(t *testing.T) {
	store := openTemp(t)
	_, err := store.ExportSessionMarkdown("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}

func TestExportSessionJSONNotFound(t *testing.T) {
	store := openTemp(t)
	_, err := store.ExportSessionJSON("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}

func TestImportSessionJSON(t *testing.T) {
	store := openTemp(t)
	id := seedForExport(t, store)

	data, err := store.ExportSessionJSON(id)
	if err != nil {
		t.Fatalf("ExportSessionJSON: %v", err)
	}

	newID, err := store.ImportSessionJSON(data)
	if err != nil {
		t.Fatalf("ImportSessionJSON: %v", err)
	}
	if newID == "" {
		t.Error("ImportSessionJSON returned empty id")
	}

	sess, err := store.GetSession(newID)
	if err != nil {
		t.Fatalf("GetSession after import: %v", err)
	}
	if sess.Provider != "deepseek" {
		t.Errorf("imported provider = %q, want deepseek", sess.Provider)
	}
	if sess.Model != "expert" {
		t.Errorf("imported model = %q, want expert", sess.Model)
	}

	msgs, err := store.GetMessages(newID)
	if err != nil {
		t.Fatalf("GetMessages after import: %v", err)
	}
	if len(msgs) != 2 {
		t.Errorf("imported messages count = %d, want 2", len(msgs))
	}
}

func TestImportSessionJSONMissingSession(t *testing.T) {
	store := openTemp(t)
	_, err := store.ImportSessionJSON([]byte(`{"messages":[]}`))
	if err == nil {
		t.Error("expected error for missing session field")
	}
}

func TestImportSessionJSONMissingProvider(t *testing.T) {
	store := openTemp(t)
	_, err := store.ImportSessionJSON([]byte(`{"session":{"model":"gpt-4"},"messages":[]}`))
	if err == nil {
		t.Error("expected error for missing provider")
	}
}

func TestImportSessionMarkdown(t *testing.T) {
	store := openTemp(t)
	id := seedForExport(t, store)

	md, err := store.ExportSessionMarkdown(id)
	if err != nil {
		t.Fatalf("ExportSessionMarkdown: %v", err)
	}

	newID, err := store.ImportSessionMarkdown([]byte(md))
	if err != nil {
		t.Fatalf("ImportSessionMarkdown: %v", err)
	}
	if newID == "" {
		t.Error("ImportSessionMarkdown returned empty id")
	}

	sess, err := store.GetSession(newID)
	if err != nil {
		t.Fatalf("GetSession after import: %v", err)
	}
	if sess.Provider != "deepseek" {
		t.Errorf("imported provider = %q, want deepseek", sess.Provider)
	}

	msgs, err := store.GetMessages(newID)
	if err != nil {
		t.Fatalf("GetMessages after import: %v", err)
	}
	if len(msgs) != 2 {
		t.Errorf("imported messages count = %d, want 2", len(msgs))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}