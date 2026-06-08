package runtime

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"claw-code-go/internal/api"
)

func TestRalphConfig_Defaults(t *testing.T) {
	cfg := DefaultRalphConfig()
	if cfg.SpecPath == "" {
		t.Error("SpecPath should have a default")
	}
	if cfg.MaxIterations <= 0 {
		t.Error("MaxIterations should have a positive default")
	}
	if cfg.DoneSentinel == "" {
		t.Error("DoneSentinel should have a default")
	}
	if cfg.PromptTemplate == "" {
		t.Error("PromptTemplate should have a default")
	}
}

func TestRenderRalphPrompt(t *testing.T) {
	cfg := DefaultRalphConfig()
	cfg.SpecPath = "ROADMAP.md"
	rendered, err := RenderRalphPrompt(cfg, "do thing one", 1, 10)
	if err != nil {
		t.Fatalf("RenderRalphPrompt: %v", err)
	}
	if !strings.Contains(rendered, "iteration 1/10") {
		t.Error("expected iteration/max in prompt")
	}
	if !strings.Contains(rendered, "ROADMAP.md") {
		t.Error("expected spec path in prompt")
	}
	if !strings.Contains(rendered, "do thing one") {
		t.Error("expected spec body in prompt")
	}
	if !strings.Contains(rendered, cfg.DoneSentinel) {
		t.Error("expected sentinel in prompt")
	}
}

func TestRenderRalphPrompt_RejectsBadTemplate(t *testing.T) {
	cfg := DefaultRalphConfig()
	cfg.PromptTemplate = "{{ .Undeclared }"
	_, err := RenderRalphPrompt(cfg, "x", 1, 1)
	if err == nil {
		t.Error("expected error on bad template")
	}
}

func TestRalphDetectsSentinel(t *testing.T) {
	cfg := DefaultRalphConfig()
	cases := []struct {
		text     string
		expected bool
	}{
		{"I marked item done.\nRALPH_DONE\n", true},
		{"RALPH_DONE", true},
		{"  RALPH_DONE  \n", true},
		{"RALPH_DONE_NEAR_BUT_NOT_IT", false},
		{"I should now RALPH_DONE with the spec but didn't on its own line", false},
		{"", false},
		{"ralph_done", false}, // case-sensitive on purpose
	}
	for _, c := range cases {
		got := RalphDetectedSentinel(c.text, cfg.DoneSentinel)
		if got != c.expected {
			t.Errorf("for %q: got %v, want %v", c.text, got, c.expected)
		}
	}
}

func TestRalphSpecIsComplete(t *testing.T) {
	cases := []struct {
		name     string
		spec     string
		expected bool
	}{
		{"all-checked", "- [x] one\n- [x] two\n", true},
		{"some-unchecked", "- [x] one\n- [ ] two\n", false},
		{"no-items", "## Notes\nJust prose here.\n", true},
		{"mixed-todos", "- [x] done\n- [TODO] pending\n- [ ] pending2\n", false},
		{"empty", "", true},
		{"only-completed-heading", "## Completed\n- [x] one\n", true},
	}
	for _, c := range cases {
		got := RalphSpecIsComplete(c.spec)
		if got != c.expected {
			t.Errorf("%s: got %v, want %v", c.name, got, c.expected)
		}
	}
}

func TestRalphIteration_DetectsSentinel(t *testing.T) {
	cfg := DefaultRalphConfig()
	verdict := RalphVerdict("All done.\nRALPH_DONE\n", "- [x] one\n- [x] two\n", cfg)
	if verdict != RalphVerdictDone {
		t.Errorf("got %v, want %v", verdict, RalphVerdictDone)
	}
	verdict = RalphVerdict("I am still working on item three.", "- [x] one\n- [ ] three\n", cfg)
	if verdict != RalphVerdictContinue {
		t.Errorf("got %v, want %v", verdict, RalphVerdictContinue)
	}
	verdict = RalphVerdict("Spec looks empty to me.", "", cfg)
	if verdict != RalphVerdictDone {
		t.Errorf("got %v, want %v", verdict, RalphVerdictDone)
	}
}

func TestRalphLoop_StopsOnVerdict(t *testing.T) {
	called := 0
	iterFn := func(ctx context.Context, iteration, maxIter int) (RalphVerdictKind, string, error) {
		called++
		if iteration < 3 {
			return RalphVerdictContinue, "still going", nil
		}
		return RalphVerdictDone, "RALPH_DONE", nil
	}
	cfg := DefaultRalphConfig()
	cfg.MaxIterations = 10
	err := RunRalphLoopWithIter(context.Background(), &cfg, iterFn)
	if err != nil {
		t.Fatalf("RunRalphLoopWithIter: %v", err)
	}
	if called != 3 {
		t.Errorf("expected 3 iterations, got %d", called)
	}
}

func TestRalphLoop_StopsAtMax(t *testing.T) {
	called := 0
	iterFn := func(ctx context.Context, iteration, maxIter int) (RalphVerdictKind, string, error) {
		called++
		// Return unique text each iteration to avoid stupid-loop detection.
		return RalphVerdictContinue, fmt.Sprintf("iter-%d", iteration), nil
	}
	cfg := DefaultRalphConfig()
	cfg.MaxIterations = 5
	err := RunRalphLoopWithIter(context.Background(), &cfg, iterFn)
	if err == nil {
		t.Error("expected error on max iterations")
	}
	if called != 5 {
		t.Errorf("expected 5 iterations, got %d", called)
	}
}

func TestRalphToolSummary(t *testing.T) {
	msgs := []api.Message{
		{Role: "assistant", Content: []api.ContentBlock{
			{Type: "tool_use", Name: "read_file"},
			{Type: "tool_use", Name: "glob"},
		}},
		{Role: "user", Content: []api.ContentBlock{
			{Type: "tool_result"},
		}},
		{Role: "assistant", Content: []api.ContentBlock{
			{Type: "tool_use", Name: "file_edit"},
			{Type: "tool_use", Name: "bash"},
		}},
	}
	reads, writes := ralphToolSummary(msgs)
	if reads != 2 {
		t.Errorf("reads: got %d, want 2", reads)
	}
	if writes != 2 {
		t.Errorf("writes: got %d, want 2", writes)
	}
}

func TestRalphToolSummary_ReadOnly(t *testing.T) {
	msgs := []api.Message{
		{Role: "assistant", Content: []api.ContentBlock{
			{Type: "tool_use", Name: "read_file"},
			{Type: "tool_use", Name: "grep"},
			{Type: "tool_use", Name: "glob"},
		}},
	}
	reads, writes := ralphToolSummary(msgs)
	if reads != 3 {
		t.Errorf("reads: got %d, want 3", reads)
	}
	if writes != 0 {
		t.Errorf("writes: got %d, want 0", writes)
	}
}

func TestRalphBuildProgressNotes_ReadOnlyWarning(t *testing.T) {
	notes := ralphBuildProgressNotes(5, 0, "")
	if !strings.Contains(notes, "WARNING") {
		t.Error("expected WARNING for read-only iteration")
	}
	if !strings.Contains(notes, "5 reads, 0 writes") {
		t.Errorf("expected tool call count, got: %s", notes)
	}
}

func TestRalphBuildProgressNotes_WithWrites(t *testing.T) {
	notes := ralphBuildProgressNotes(2, 3, "")
	if strings.Contains(notes, "WARNING") {
		t.Error("did not expect WARNING when writes exist")
	}
	if !strings.Contains(notes, "2 reads, 3 writes") {
		t.Errorf("expected tool call count, got: %s", notes)
	}
}

func TestRalphBuildProgressNotes_WithGitDiff(t *testing.T) {
	notes := ralphBuildProgressNotes(1, 1, " file.go | 5 +-")
	if !strings.Contains(notes, "Git changes") {
		t.Error("expected git changes header")
	}
	if !strings.Contains(notes, "file.go") {
		t.Error("expected file.go in diff")
	}
}

func TestRenderRalphPrompt_ProgressNotes(t *testing.T) {
	cfg := DefaultRalphConfig()
	cfg.ProgressNotes = "WARNING: last iteration only READ files"
	rendered, err := RenderRalphPrompt(cfg, "spec body", 2, 10)
	if err != nil {
		t.Fatalf("RenderRalphPrompt: %v", err)
	}
	if !strings.Contains(rendered, "LAST ITERATION") {
		t.Error("expected LAST ITERATION header when ProgressNotes set")
	}
	if !strings.Contains(rendered, "WARNING: last iteration only READ files") {
		t.Error("expected progress notes content in prompt")
	}
}

func TestRenderRalphPrompt_NoProgressNotes(t *testing.T) {
	cfg := DefaultRalphConfig()
	rendered, err := RenderRalphPrompt(cfg, "spec body", 1, 10)
	if err != nil {
		t.Fatalf("RenderRalphPrompt: %v", err)
	}
	if strings.Contains(rendered, "LAST ITERATION") {
		t.Error("did not expect LAST ITERATION when ProgressNotes empty")
	}
}

func TestRalphConfig_NewDefaults(t *testing.T) {
	cfg := DefaultRalphConfig()
	if cfg.IterationTimeout != DefaultRalphIterationTimeout {
		t.Errorf("IterationTimeout: got %v, want %v", cfg.IterationTimeout, DefaultRalphIterationTimeout)
	}
	if !cfg.ProgressTracking {
		t.Error("ProgressTracking should default to true")
	}
}

func TestRalphIterationTimeout(t *testing.T) {
	timeout := DefaultRalphIterationTimeout
	if timeout <= 0 {
		t.Error("DefaultRalphIterationTimeout should be positive")
	}
	cfg := DefaultRalphConfig()
	if cfg.IterationTimeout != timeout {
		t.Errorf("IterationTimeout: got %v, want %v", cfg.IterationTimeout, timeout)
	}
	// Verify that a very short timeout causes context cancellation
	shortTimeout := 1 * time.Millisecond
	iterCtx, cancel := context.WithTimeout(context.Background(), shortTimeout)
	defer cancel()
	// Sleep past the deadline
	time.Sleep(10 * time.Millisecond)
	if iterCtx.Err() == nil {
		t.Error("expected context to be cancelled after short timeout")
	}
}
