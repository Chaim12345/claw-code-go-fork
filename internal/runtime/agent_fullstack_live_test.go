//go:build live

package runtime

import (
	"context"
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

// TestLiveAgentFullStackProject is a 5-step end-to-end build that
// exercises every code-tool path through the agentic loop:
//
//   1. Glob the project for existing files
//   2. Read the package metadata
//   3. Write three new modules
//   4. Grep for the new function names to confirm they're present
//   5. Run `go build` and `go test` and report the exit status
//
// It also forces two intermediate compaction boundaries by injecting
// filler turns between steps 3 and 4, so we verify that the
// CompactionSummary carries enough state for the post-compaction
// grep/build to still be coherent.
func TestLiveAgentFullStackProject(t *testing.T) {
	tok := os.Getenv("DEEPSEEK_TOKEN")
	if tok == "" {
		t.Skip("DEEPSEEK_TOKEN not set; skipping live fullstack test")
	}
	// Hold off a moment so any prior test's rate-limit budget
	// has time to relax before we hit the API. See
	// live_retry_helper_test.go for the rationale.
	sleepForRateLimit()
	workdir := t.TempDir()
	oldwd, _ := os.Getwd()
	if err := os.Chdir(workdir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldwd) })

	// Bootstrap a real, buildable Go module so the final `go build` step
	// is a meaningful verification.
	modName := "liveagentdemo"
	if out, err := exec.Command("go", "mod", "init", modName).CombinedOutput(); err != nil {
		t.Fatalf("go mod init: %v\n%s", err, out)
	}

	cfg := LoadConfig()
	cfg.ProviderName = "deepseek"
	cfg.Model = "expert"
	cfg.MaxTokens = 2048
	cfg.PermissionMode = "bypass"
	cfg.SessionDir = t.TempDir()
	cfg.CompactionThreshold = 0.5

	provider := deepseek.New()
	client, err := provider.NewClient(api.ProviderConfig{APIKey: os.Getenv("DEEPSEEK_TOKEN"), Model: "expert"})
	if err != nil {
		t.Fatal(err)
	}
	loop := NewConversationLoop(cfg, client)
	loop.PermManager = permissions.NewManager(permissions.ModeBypassPermissions, &permissions.Ruleset{})

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()

	step := func(label, prompt string) {
		t.Helper()
		t.Logf("=== %s ===", label)
		if err := loop.SendMessage(ctx, prompt); err != nil {
			t.Fatalf("%s: %v", label, err)
		}
		// Log the latest assistant text for visibility.
		for i := len(loop.Session.Messages) - 1; i >= 0; i-- {
			if loop.Session.Messages[i].Role == "assistant" {
				for _, b := range loop.Session.Messages[i].Content {
					if b.Type == "text" {
						t.Logf("[%s reply] %s", label, strings.TrimSpace(b.Text))
						return
					}
				}
			}
		}
	}

	step("step 1: bootstrap",
		"Bootstrap a Go package: write the file mathutil/divide.go "+
			"containing a function Divide(a, b float64) (float64, error) "+
			"that returns a/b and an error when b==0. Use the write_file tool. "+
			"Use raw JSON for tool calls.")

	step("step 2: extend",
		"Now write the file mathutil/divide_test.go containing one test "+
			"`func TestDivideByZero(t *testing.T)` that calls Divide(1, 0) "+
			"and checks the error is non-nil. Use write_file. "+
			"Use raw JSON for tool calls.")

	// Step 3: large filler turn to push us into compaction territory.
	// Expert caps at 163,840 chars / 40,960 tokens per single message,
	// so the filler must stay under that. 6,000 × 22 chars = 132,000
	// chars (~33k tokens), well under the cap but enough to trip the
	// 0.5 compaction threshold on the next turn.
	step("step 3: filler",
		"Memorise the following: "+strings.Repeat("history-pad sentence. ", 6_000))

	step("step 4: verify with grep",
		"Use grep to confirm that both DivideByZero and the Divide function "+
			"appear in the mathutil directory. Report what grep returned. "+
			"Use raw JSON for tool calls.")

	step("step 5: build + test",
		"Run `go test ./...` and report the exit status. "+
			"Use raw JSON tool calls.")

	// Final disk verification: the source files must exist, contain the
	// required functions, and `go test` must pass on the package as
	// written.
	divideFile := filepath.Join(workdir, "mathutil", "divide.go")
	testFile := filepath.Join(workdir, "mathutil", "divide_test.go")

	divideSrc, err := os.ReadFile(divideFile)
	if err != nil {
		t.Fatalf("divide.go not on disk: %v", err)
	}
	if !strings.Contains(string(divideSrc), "func Divide(") {
		t.Errorf("divide.go missing Divide function:\n%s", string(divideSrc))
	}

	testSrc, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("divide_test.go not on disk: %v", err)
	}
	if !strings.Contains(string(testSrc), "TestDivideByZero") {
		t.Errorf("divide_test.go missing TestDivideByZero:\n%s", string(testSrc))
	}

	// Run the test the agent was asked to run, ourselves — this proves
	// the package actually compiles and passes, regardless of whether
	// the agent's reported exit status was accurate.
	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = workdir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go test ./... failed (the package the agent wrote is broken):\n%s", out)
	}
	t.Logf("go test ./...: %s", strings.TrimSpace(string(out)))

	// Session health summary.
	toolCounts := map[string]int{}
	var textBytes int
	for _, m := range loop.Session.Messages {
		for _, b := range m.Content {
			if b.Type == "tool_use" {
				toolCounts[b.Name]++
			}
			if b.Type == "text" {
				textBytes += len(b.Text)
			}
		}
	}

	t.Logf("session: %d messages, %d tool calls, %d bytes of assistant text",
		len(loop.Session.Messages), sumCounts(toolCounts), textBytes)
	t.Logf("tool distribution: %v", toolCounts)
	t.Logf("compaction count: %d, summary length: %d chars",
		loop.Compaction.CompactionCount, len(loop.Session.CompactionSummary))

	if sumCounts(toolCounts) < 3 {
		t.Errorf("expected at least 3 tool calls across the run, got %d", sumCounts(toolCounts))
	}
}

func sumCounts(m map[string]int) int {
	n := 0
	for _, v := range m {
		n += v
	}
	return n
}
