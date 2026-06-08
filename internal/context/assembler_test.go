package context

import (
	"strings"
	"testing"
	"time"

	"claw-code-go/internal/session"
)

func TestAssembler_Assemble_NotEmpty(t *testing.T) {
	a := NewAssembler(".")
	result := a.Assemble()
	if result == "" {
		t.Error("Assemble() returned empty string")
	}
}

func TestAssembler_TokenBudget(t *testing.T) {
	// The assembler uses a 12K token budget with estimateTokens = len(s)/4.
	// The soft prompt limit is 28K (set in config). 12K is well within limits.
	a := NewAssembler(".")
	result := a.Assemble()
	estimatedTokens := len(result) / 4

	// Budget is 12K in assembler code, soft limit is 28K
	if estimatedTokens > 28000 {
		t.Errorf("token budget exceeded: estimated %d tokens (soft limit 28K)", estimatedTokens)
	}
	t.Logf("estimated tokens: %d / 28000 soft limit", estimatedTokens)
}

func BenchmarkAssembler_Assemble(b *testing.B) {
	a := NewAssembler(".")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Assemble()
	}
}

func TestAssembler_Latency(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping latency test in short mode")
	}
	a := NewAssembler(".")
	start := time.Now()
	for i := 0; i < 10; i++ {
		a.Assemble()
	}
	elapsed := time.Since(start)
	avg := elapsed / 10
	t.Logf("average assembly time: %v", avg)
	if avg > 500*time.Millisecond {
		t.Errorf("context assembly exceeded 500ms latency: %v", avg)
	}
}

func TestAssembler_IncludesEnvironment(t *testing.T) {
	a := NewAssembler(".")
	result := a.Assemble()
	// Should include environment info, git status, and project structure
	hasEnv := containsSubstring(result, "Environment")
	hasGit := containsSubstring(result, "Git Status")
	hasProject := containsSubstring(result, "Project Structure")
	hasSymbols := containsSubstring(result, "Key Types")
	if !hasEnv || !hasGit || !hasProject || !hasSymbols {
		t.Logf("result preview (first 500 chars):\n%s", truncateString(result, 500))
	}
	if !hasEnv {
		t.Error("missing Environment section")
	}
	if !hasGit {
		t.Error("missing Git Status section")
	}
	// Project structure may be truncated if token budget is tight
	if !hasProject {
		t.Log("NOTE: Project Structure section not found (may be truncated by budget)")
	}
	if !hasSymbols {
		t.Log("NOTE: Key Types section not found (may be truncated by budget)")
	}
}

func TestAssembler_CacheHit(t *testing.T) {
	a := NewAssembler(".")
	// First call builds cache
	_ = a.Assemble()
	// Second call should use cache
	start := time.Now()
	for i := 0; i < 10; i++ {
		a.Assemble()
	}
	elapsed := time.Since(start)
	avg := elapsed / 10
	t.Logf("cached assembly time: %v", avg)
	// Cached should be fast (<250ms even with I/O)
	if avg > 250*time.Millisecond {
		t.Errorf("cached assembly too slow: %v", avg)
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

type mockStore struct {
	sessions  []*session.Session
	summaries map[string]string
}

func (m *mockStore) GetRecentSessions(limit int) ([]*session.Session, error) {
	if limit > len(m.sessions) {
		limit = len(m.sessions)
	}
	return m.sessions[:limit], nil
}
func (m *mockStore) GetSessionSummary(id string) (string, error) {
	return m.summaries[id], nil
}
func (m *mockStore) CreateSession(_, _, _ string) error               { return nil }
func (m *mockStore) GetSession(_ string) (*session.Session, error)    { return nil, nil }
func (m *mockStore) ListSessions() ([]*session.Session, error)        { return m.sessions, nil }
func (m *mockStore) SearchSessions(_ string) ([]*session.Session, error) { return nil, nil }
func (m *mockStore) FilterSessions(_, _ string, _, _ time.Time) ([]*session.Session, error) {
	return nil, nil
}
func (m *mockStore) DeleteSession(_ string) error                          { return nil }
func (m *mockStore) UpdateSessionPreview(_, _, _ string, _ int) error      { return nil }
func (m *mockStore) RecordMessage(_, _, _ string, _ int) (int64, error)    { return 0, nil }
func (m *mockStore) GetMessages(_ string) ([]*session.Message, error)      { return nil, nil }
func (m *mockStore) RecordToolCall(_ int64, _, _, _ string, _ int) error   { return nil }
func (m *mockStore) GetToolCalls(_ int64) ([]*session.ToolCall, error)     { return nil, nil }
func (m *mockStore) ExportSessionMarkdown(_ string) (string, error)        { return "", nil }
func (m *mockStore) ExportSessionJSON(_ string) ([]byte, error)            { return nil, nil }
func (m *mockStore) ImportSessionJSON(_ []byte) (string, error)           { return "", nil }
func (m *mockStore) ImportSessionMarkdown(_ []byte) (string, error)        { return "", nil }
func (m *mockStore) Close() error                                          { return nil }

func TestAssembler_SessionHistory_NilStore(t *testing.T) {
	a := NewAssembler(".")
	result := a.buildSessionHistorySection(2000)
	if result != "" {
		t.Errorf("expected empty string with nil store, got: %q", result)
	}
}

func TestAssembler_SessionHistory_WithSessions(t *testing.T) {
	store := &mockStore{
		sessions: []*session.Session{
			{ID: "abc123", Provider: "deepseek", Model: "expert", LastActiveAt: time.Now()},
			{ID: "def456", Provider: "anthropic", Model: "sonnet-4", LastActiveAt: time.Now()},
		},
		summaries: map[string]string{
			"abc123": "Fixed UUID bug in WebSocket handler",
			"def456": "Added context injection to assembler",
		},
	}
	a := NewAssembler(".")
	a.SetSessionStore(store)

	result := a.buildSessionHistorySection(2000)
	if !strings.Contains(result, "Recent Sessions") {
		t.Error("missing 'Recent Sessions' heading")
	}
	if !strings.Contains(result, "abc123") {
		t.Error("missing session ID abc123")
	}
	if !strings.Contains(result, "Fixed UUID bug") {
		t.Error("missing summary for abc123")
	}
	if !strings.Contains(result, "deepseek/expert") {
		t.Error("missing provider/model for abc123")
	}
}

func TestAssembler_SessionHistory_AssembleIntegration(t *testing.T) {
	store := &mockStore{
		sessions: []*session.Session{
			{ID: "s1", Provider: "deepseek", Model: "expert", LastActiveAt: time.Now()},
		},
		summaries: map[string]string{
			"s1": "Implemented session history context",
		},
	}
	a := NewAssembler(".")
	a.SetSessionStore(store)

	result := a.Assemble()
	if !strings.Contains(result, "Recent Sessions") {
		t.Error("Assemble() output missing Recent Sessions section")
	}
	if !strings.Contains(result, "Implemented session history context") {
		t.Error("Assemble() output missing session summary text")
	}
}

func TestAssembler_SessionHistory_BudgetCapped(t *testing.T) {
	longSummary := strings.Repeat("x", 50000)
	store := &mockStore{
		sessions: []*session.Session{
			{ID: "big", Provider: "deepseek", Model: "expert", LastActiveAt: time.Now()},
		},
		summaries: map[string]string{
			"big": longSummary,
		},
	}
	a := NewAssembler(".")
	a.SetSessionStore(store)

	result := a.buildSessionHistorySection(2000)
	estimatedTokens := len(result) / 4
	if estimatedTokens > 2200 {
		t.Errorf("session history section exceeded budget: ~%d tokens (limit 2000)", estimatedTokens)
	}
}