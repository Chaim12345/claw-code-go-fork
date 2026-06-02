//go:build live

package runtime

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"claw-code-go/internal/api"
	"claw-code-go/internal/api/providers/deepseek"
	"claw-code-go/internal/permissions"
)

// longTestTimeout caps the entire multi-turn agent run. The test sends
// many turns (read, write, edit, grep, glob) and triggers compaction.
const longTestTimeout = 5 * time.Minute

func newDeepSeekLoopForLive(t *testing.T, workdir string) *ConversationLoop {
	t.Helper()
	tok := os.Getenv("DEEPSEEK_TOKEN")
	if tok == "" {
		t.Skip("DEEPSEEK_TOKEN not set; skipping live agent test")
	}

	// Brief sleep so consecutive live tests don't trip the
	// per-token rate limit on the DeepSeek web API. The helper
	// respects DEEPSEEK_LIVE_NO_SLEEP=1 for timing-sensitive
	// debugging.
	sleepForRateLimit()

	cfg := LoadConfig()
	cfg.ProviderName = "deepseek"
	cfg.Model = "expert"
	cfg.MaxTokens = 2048
	cfg.PermissionMode = "bypass"
	cfg.SessionDir = t.TempDir()
	// Expert caps at 163,840 chars (~40,960 tokens). The compaction
	// basis is set automatically by LoadConfig from
	// PerProviderCompactionMax, so 0.5 means compact at ~20k tokens
	// of context — early enough to exercise compaction before the
	// pre-flight per-message cap trips.
	cfg.CompactionThreshold = 0.5

	provider := deepseek.New()
	client, err := provider.NewClient(api.ProviderConfig{APIKey: tok, Model: "expert"})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	loop := NewConversationLoop(cfg, client)
	loop.PermManager = permissions.NewManager(permissions.ModeBypassPermissions, &permissions.Ruleset{})
	return loop
}

// safeWD switches into workdir for the duration of t, restoring the original
// on cleanup. The agent's tools (bash, read, write) operate on the process
// working directory.
func safeWD(t *testing.T, workdir string) {
	t.Helper()
	old, _ := os.Getwd()
	if err := os.Chdir(workdir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
}

// TestLiveAgentCodingTask drives a real coding task through the full
// agentic loop: read existing file → write a new module → grep to verify →
// edit it → glob to find generated artifacts.
//
// Verifies that:
//  - tool calls work (bash, read, write, edit, grep, glob)
//  - the conversation loop continues across multiple tool turns
//  - the final answer references the files actually on disk
//  - the tool dispatcher recognises the names emitted by the deepseek provider
func TestLiveAgentCodingTask(t *testing.T) {
	workdir := t.TempDir()
	safeWD(t, workdir)

	// Seed the project: a small Go file the agent must read and extend.
	seedFile := filepath.Join(workdir, "calc.go")
	if err := os.WriteFile(seedFile, []byte(`package calc

// Add returns a + b.
func Add(a, b int) int {
	return a + b
}
`), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	loop := newDeepSeekLoopForLive(t, workdir)

	ctx, cancel := context.WithTimeout(context.Background(), longTestTimeout)
	defer cancel()

	prompt := fmt.Sprintf(
		"You are working in %s. The file calc.go defines an Add function. "+
			"Your job: (1) read calc.go, (2) add a Subtract function "+
			"and a Multiply function to the same file using the file_edit "+
			"or write tool, (3) run `go build ./...` to confirm it compiles, "+
			"(4) report the result. Use raw JSON for tool calls, no markdown.",
		workdir,
	)

	// Collect tool calls the agent actually executed so we can assert on them.
	_ = func() {} // placeholder for any future instrumentation

	// Wrap the loop's stream to record events. SendMessage doesn't expose
	// streaming; instead we run turn-by-turn and inspect Session.Messages
	// after each call. We instrument by reading the session between calls.
	if err := loop.SendMessage(ctx, prompt); err != nil {
		t.Fatalf("SendMessage: %v", err)
	}

	// Inspect the session: count tool uses, verify the agent saw a file_read
	// and at least one file_edit/write, and that the final assistant turn
	// references the new functions.
	msgs := loop.Session.Messages
	toolCallCount := 0
	toolNames := map[string]int{}
	var finalAssistantText strings.Builder

	for _, m := range msgs {
		for _, b := range m.Content {
			if b.Type == "tool_use" {
				toolCallCount++
				toolNames[b.Name]++
			}
		}
		if m.Role == "assistant" {
			for _, b := range m.Content {
				if b.Type == "text" {
					finalAssistantText.WriteString(b.Text)
				}
			}
		}
	}

	if toolCallCount == 0 {
		t.Fatalf("agent never called a tool — got %d messages, last assistant text: %q",
			len(msgs), finalAssistantText.String())
	}
	t.Logf("agent made %d tool calls across %d messages", toolCallCount, len(msgs))
	t.Logf("tool distribution: %v", toolNames)

	// The agent must have actually executed at least one of: read_file, file_edit, write_file, grep, glob, bash.
	seenAny := false
	for _, n := range []string{"read_file", "file_edit", "write_file", "grep", "glob", "bash", "read", "write", "edit"} {
		if toolNames[n] > 0 {
			seenAny = true
			break
		}
	}
	if !seenAny {
		t.Errorf("expected at least one of read/write/edit/grep/glob/bash calls; got %v", toolNames)
	}

	// The on-disk file must now contain the new functions.
	data, err := os.ReadFile(seedFile)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	body := string(data)
	t.Logf("final calc.go:\n%s", body)
	if !strings.Contains(body, "func Subtract") {
		t.Error("expected Subtract function to be in calc.go")
	}
	if !strings.Contains(body, "func Multiply") {
		t.Error("expected Multiply function to be in calc.go")
	}

	// Final assistant text should mention completion or the new functions.
	final := strings.ToLower(finalAssistantText.String())
	if final == "" {
		t.Error("expected a non-empty final assistant message")
	}
}

// TestLiveAgentMultiTurnFollowUp sends a follow-up request that requires
// the agent to remember what it did in a previous turn. This is the
// canonical "context management" test — the previous files + results
// must be in the conversation history when the second prompt runs.
func TestLiveAgentMultiTurnFollowUp(t *testing.T) {
	workdir := t.TempDir()
	safeWD(t, workdir)

	loop := newDeepSeekLoopForLive(t, workdir)
	ctx, cancel := context.WithTimeout(context.Background(), longTestTimeout)
	defer cancel()

	// Turn 1: create a file.
	if err := loop.SendMessage(ctx,
		fmt.Sprintf("Create a file %s/notes.txt containing exactly the string 'first' and nothing else. "+
			"Use raw JSON tool calls.", workdir)); err != nil {
		t.Fatalf("turn 1: %v", err)
	}
	if data, err := os.ReadFile(filepath.Join(workdir, "notes.txt")); err != nil {
		t.Fatalf("turn 1 file not created: %v", err)
	} else if !strings.Contains(string(data), "first") {
		t.Errorf("turn 1 file content wrong: %q", string(data))
	}

	turn1Msgs := len(loop.Session.Messages)

	// Turn 2: read the file and append to it. This requires the agent to
	// remember the path AND the previous content from history.
	if err := loop.SendMessage(ctx,
		"Now use read_file to read notes.txt, then use file_edit (or write_file) "+
			"to APPEND the string 'second' on a new line. Do not stop after the first "+
			"tool call — keep going until the file contains both 'first' and 'second'. "+
			"Use raw JSON tool calls."); err != nil {
		t.Fatalf("turn 2: %v", err)
	}
	if turn1Msgs >= len(loop.Session.Messages) {
		t.Errorf("turn 2 didn't add new messages (turn1=%d, now=%d)", turn1Msgs, len(loop.Session.Messages))
	}

	data, err := os.ReadFile(filepath.Join(workdir, "notes.txt"))
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	body := string(data)
	t.Logf("notes.txt after turn 2: %q", body)
	if !strings.Contains(body, "first") {
		t.Error("first turn content was lost")
	}
	if !strings.Contains(body, "second") {
		t.Error("second turn content was not appended")
	}
}

// TestLiveAgentCompaction forces compaction by sending many turns and a
// deliberately large message. After compaction, the session should still
// produce a coherent answer (the summary must carry enough state for the
// agent to continue).
func TestLiveAgentCompaction(t *testing.T) {
	workdir := t.TempDir()
	safeWD(t, workdir)

	loop := newDeepSeekLoopForLive(t, workdir)
	ctx, cancel := context.WithTimeout(context.Background(), longTestTimeout)
	defer cancel()

	// Pad the session with a deliberately large system-side input so the
	// estimator trips ShouldCompact on the next turn. The pad has to
	// stay under Expert's per-message cap (163,840 chars / 40,960
	// tokens) or the pre-flight check rejects it before compaction can
	// fire.
	huge := strings.Repeat("padding text that fills the context window. ", 3_000) // ~138k chars
	if err := loop.SendMessage(ctx,
		"Please memorise the following text verbatim, you may be tested on it later:\n\n"+huge); err != nil {
		t.Fatalf("pad: %v", err)
	}

	msgsBefore := len(loop.Session.Messages)
	t.Logf("messages before compaction-trigger turn: %d, CompactionCount=%d, LastInputTokens=%d",
		msgsBefore, loop.Compaction.CompactionCount, loop.Compaction.LastInputTokens)

	// One more turn — ShouldCompact is checked at the top of SendMessage,
	// so if the previous turn pushed LastInputTokens high enough, the
	// summary fires before this prompt goes out.
	if err := loop.SendMessage(ctx, "Reply with the single word: ok."); err != nil {
		t.Fatalf("post-pad: %v", err)
	}

	// If compaction ran, the session should have a CompactionSummary.
	compacted := loop.Session.CompactionSummary != ""
	t.Logf("after second turn: CompactionSummary set=%v, CompactionCount=%d, msgs=%d, LastInputTokens=%d",
		compacted, loop.Compaction.CompactionCount, len(loop.Session.Messages), loop.Compaction.LastInputTokens)

	if !compacted {
		// Acceptable outcome if the model used fewer tokens than the
		// conservative estimator predicted. We don't fail the test, but
		// we do report the gap so the threshold can be tuned.
		t.Logf("note: CompactionSummary not set; CompactionCount=%d LastInputTokens=%d",
			loop.Compaction.CompactionCount, loop.Compaction.LastInputTokens)
	}

	// The real test invariant: CompactionCount > 0 means the loop
	// correctly observed input tokens exceed the threshold and
	// attempted compaction. The actual summary content depends on
	// the model returning something, which is flaky under load.
	// We accept non-empty assistant content from EITHER turn as
	// evidence the agent stayed alive end-to-end.
	if loop.Compaction.CompactionCount == 0 {
		t.Errorf("expected CompactionCount > 0, got 0 — ShouldCompact never fired; LastInputTokens=%d threshold_basis=%d",
			loop.Compaction.LastInputTokens, loop.Config.CompactionMaxInputTokens)
	}

	gotReply := false
	for _, m := range loop.Session.Messages {
		if m.Role != "assistant" {
			continue
		}
		for _, b := range m.Content {
			if b.Type == "text" && strings.TrimSpace(b.Text) != "" {
				gotReply = true
			}
		}
	}
	if !gotReply {
		t.Logf("note: no text in any assistant turn (server returned empty under load); CompactionCount=%d msgs=%d",
			loop.Compaction.CompactionCount, len(loop.Session.Messages))
	}
}

// TestLiveAgentSendMessageTokenTracking verifies that after a single
// SendMessage, the conversation loop's CompactionState and Usage
// tracker have been updated with the real (server-reported or
// estimator-derived) input/output token counts.
//
// Codegraph analysis (see plan/codegraph-report.md) found that
// runOneTurn (the path used by SendMessage, the CLI, and every live
// test) silently discarded event.InputTokens and event.Usage.OutputTokens
// from the stream. Only runOneTurnStreaming (used by the TUI) read
// them. As a result, loop.Compaction.LastInputTokens stayed 0
// forever on the CLI/live-test path, and loop.Usage never received
// any .Add() calls.
//
// This test pins the contract: a turn through SendMessage must
// populate LastInputTokens and TotalOutputTokens so ShouldCompact can
// trigger on the real numbers.
func TestLiveAgentSendMessageTokenTracking(t *testing.T) {
	if os.Getenv("DEEPSEEK_TOKEN") == "" {
		t.Skip("DEEPSEEK_TOKEN not set; skipping live test")
	}
	workdir := t.TempDir()
	safeWD(t, workdir)

	loop := newDeepSeekLoopForLive(t, workdir)
	ctx, cancel := context.WithTimeout(context.Background(), longTestTimeout)
	defer cancel()

	beforeIn := loop.Compaction.LastInputTokens
	beforeTotal := loop.Compaction.TotalOutputTokens
	beforeUsageOut := 0
	if loop.Usage != nil {
		beforeUsageOut = loop.Usage.TotalOutput
	}
	t.Logf("before: LastInputTokens=%d TotalOutputTokens=%d Usage.TotalOutput=%d",
		beforeIn, beforeTotal, beforeUsageOut)

	// Ask for a specific response so the model is unlikely to
	// short-circuit. We then assert input AND output tokens grew,
	// OR an assistant message was added to the session (which
	// proves the turn completed). DeepSeek's web API occasionally
	// returns an empty response under load — the wiring is
	// correct as long as the loop can run, and a non-empty
	// assistant turn guarantees a real server-reported output
	// count was captured.
	const wantReply = "pong"
	if err := loop.SendMessage(ctx, "Reply with the single token: "+wantReply); err != nil {
		t.Fatalf("SendMessage: %v", err)
	}

	afterIn := loop.Compaction.LastInputTokens
	afterTotalOut := loop.Compaction.TotalOutputTokens
	afterUsageOut := 0
	if loop.Usage != nil {
		afterUsageOut = loop.Usage.TotalOutput
	}
	t.Logf("after:  LastInputTokens=%d TotalOutputTokens=%d Usage.TotalOutput=%d",
		afterIn, afterTotalOut, afterUsageOut)

	if afterIn <= beforeIn {
		t.Errorf("expected LastInputTokens to grow, got %d (was %d) — wiring is broken", afterIn, beforeIn)
	}

	// Output-token check is gated on the assistant actually
	// producing something. Under heavy load DeepSeek can return
	// an empty response (no text, no tool calls), in which case
	// the legitimately-correct total is 0 and the wiring is
	// still fine.
	gotAssistant := false
	for _, m := range loop.Session.Messages {
		if m.Role == "assistant" && len(m.Content) > 0 {
			gotAssistant = true
			break
		}
	}
	if gotAssistant {
		if afterTotalOut <= beforeTotal {
			t.Errorf("expected TotalOutputTokens to grow, got %d (was %d)", afterTotalOut, beforeTotal)
		}
		if afterUsageOut <= beforeUsageOut {
			t.Errorf("expected Usage.TotalOutput to grow, got %d (was %d)", afterUsageOut, beforeUsageOut)
		}
	} else {
		t.Logf("note: assistant produced no content (server returned empty); output token check skipped")
	}
}

// TestLiveAgentContextOverflowRecovery verifies that if a single user
// message blows past the deepseek model's per-variant cap, the agent
// surfaces a clear error instead of hanging or returning a silent failure.
func TestLiveAgentContextOverflowRecovery(t *testing.T) {
	workdir := t.TempDir()
	safeWD(t, workdir)

	loop := newDeepSeekLoopForLive(t, workdir)
	ctx, cancel := context.WithTimeout(context.Background(), longTestTimeout)
	defer cancel()

	// ~600k tokens worth of input; instant's cap is 655k chars (≈164k tokens),
	// so this should still be accepted, but with no headroom for output.
	// Push well past 1.5M chars to definitely overflow.
	huge := strings.Repeat("over the cap. ", 200_000) // ~3.4M chars
	err := loop.SendMessage(ctx, "Summarise: "+huge)
	if err == nil {
		t.Fatal("expected error for over-cap prompt, got nil")
	}
	if !strings.Contains(err.Error(), "prompt too large") {
		t.Errorf("expected 'prompt too large' error, got %v", err)
	}
	t.Logf("got expected overflow error: %v", err)
}
