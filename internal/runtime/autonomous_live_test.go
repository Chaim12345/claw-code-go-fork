//go:build live

package runtime

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"claw-code-go/internal/api"
	"claw-code-go/internal/api/providers/deepseek"
	"claw-code-go/internal/permissions"
)

// TestLiveAutonomousRunTask is the first end-to-end test of the
// autonomous mode (RunTask + LLM-judge). It seeds a Go project in a
// temp dir, gives the agent one high-level prompt that requires
// multiple tool turns (write code, run tests, fix failures), and
// verifies the loop self-terminates with all assertions true.
//
// If the loop is genuinely working, we expect:
//   - the agent writes a non-trivial .go file
//   - the agent runs `go test ./...` (or similar)
//   - on failure, it reads the error, edits the file, retries
//   - the LLM-judge returns DONE when the tests pass
//   - RunTask returns in under MaxTurns * ~30s with no human input
//
// Skipped when DEEPSEEK_TOKEN is unset.
func TestLiveAutonomousRunTask(t *testing.T) {
	tok := os.Getenv("DEEPSEEK_TOKEN")
	if tok == "" {
		t.Skip("DEEPSEEK_TOKEN not set; skipping live autonomous test")
	}
	sleepForRateLimit()

	workdir := t.TempDir()
	oldwd, _ := os.Getwd()
	if err := os.Chdir(workdir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldwd) })

	// Bootstrap a buildable Go module so `go test` is a real verifier.
	if out, err := exec.Command("go", "mod", "init", "autodemo").CombinedOutput(); err != nil {
		t.Fatalf("go mod init: %v\n%s", err, out)
	}

	// Build the loop using the same setup as the existing live
	// tests: deepseek/expert, bypass permissions, low compaction
	// threshold so the run fits.
	cfg := LoadConfig()
	cfg.ProviderName = "deepseek"
	cfg.Model = "expert"
	cfg.MaxTokens = 2048
	cfg.PermissionMode = "bypass"
	cfg.SessionDir = t.TempDir()
	cfg.CompactionThreshold = 0.5
	cfg.Autonomous = true
	cfg.MaxTurns = 15

	// Reuse the same client-builder pattern as the other live tests.
	provider := deepseek.New()
	client, err := provider.NewClient(api.ProviderConfig{APIKey: tok, Model: "expert"})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	loop := NewConversationLoop(cfg, client)
	loop.PermManager = permissions.NewManager(permissions.ModeBypassPermissions, &permissions.Ruleset{})

	// One high-level prompt that requires: write file, run tests,
	// read failure, fix, run tests again. No human input allowed.
	prompt := "Write the file mathutil/divide.go with a single function Divide(a, b float64) (float64, error) that returns a/b and an error when b==0. " +
		"Then write mathutil/divide_test.go with two tests: TestDivideHappy and TestDivideByZero. " +
		"Then run `go test ./mathutil/...` and confirm the tests pass. " +
		"If the tests fail, read the error and fix the code. Do not stop until `go test ./mathutil/...` reports ok."

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	start := time.Now()
	final, err := loop.RunTask(ctx, prompt)
	elapsed := time.Since(start)

	t.Logf("RunTask returned in %s (err=%v)", elapsed, err)
	t.Logf("final assistant text (last 800 chars): %s", tail(final, 800))

	// --- Assertions ---

	// 1. RunTask must have returned (no infinite loop / context cancel).
	if err != nil {
		if errors.Is(err, ErrMaxTurns) {
			t.Fatalf("RunTask hit MaxTurns without DONE — autonomous loop did not converge: %v", err)
		}
		if errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("RunTask exceeded the 6-minute wall budget: %v", err)
		}
		t.Fatalf("RunTask returned error: %v", err)
	}

	// 2. The Go files the agent was asked to write must exist.
	divide := filepath.Join(workdir, "mathutil", "divide.go")
	tests := filepath.Join(workdir, "mathutil", "divide_test.go")
	if _, err := os.Stat(divide); err != nil {
		t.Fatalf("expected %s to exist: %v", divide, err)
	}
	if _, err := os.Stat(tests); err != nil {
		t.Fatalf("expected %s to exist: %v", tests, err)
	}

	// 3. The package must build and tests must pass — the agent's
	//    own verifier is `go test ./mathutil/...`; we re-run it here
	//    to confirm the on-disk state is correct.
	cmd := exec.Command("go", "test", "./mathutil/...")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go test ./mathutil/... failed (autonomous agent should have fixed this before DONE):\n%s\n%v", out, err)
	}
	t.Logf("go test ./mathutil/... passed: %s", strings.TrimSpace(string(out)))

	// 4. The session must have grown — if the agent only sent one
	//    turn and exited, the loop is broken.
	if loop.MessageCount() < 4 {
		t.Fatalf("expected multi-turn conversation, got %d messages", loop.MessageCount())
	}
	t.Logf("session has %d messages, %d total turns, %d in / %d out tokens",
		loop.MessageCount(), loop.Usage.Turns,
		loop.Usage.TotalInput, loop.Usage.TotalOutput)

	// 5. The final text should not contain the FINISHED sentinel
	//    (provider-side trim is a separate concern but worth
	//    pinning here too).
	if strings.Contains(final, "FINISHED") {
		t.Errorf("final text still contains FINISHED sentinel: %q", tail(final, 200))
	}
}

// tail returns the last n bytes of s, prefixed with "..." if s was
// truncated. Used to keep test logs readable when the model emits a
// long final text.
func tail(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return "..." + s[len(s)-n:]
}

// containsString is a tiny local helper; we don't import slices
// just to keep the test self-contained.
func containsString(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
