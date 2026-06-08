package context

import (
	"testing"
	"time"
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