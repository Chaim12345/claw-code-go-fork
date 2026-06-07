//go:build live

package runtime

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"claw-code-go/internal/api"
	"claw-code-go/internal/api/providers/deepseek"
)

// TestLiveTruncationShootout compares all truncation approaches by:
// 1. Finding the rejection point (how many tokens the server actually rejects)
// 2. Testing each approach's ability to recover from rejection
//
// Approaches tested:
//   A. Simple FIFO (drop oldest messages, keep only last N)
//   B. Per-message truncation (truncate individual large blocks)
//   C. Smart compaction (summarize then keep recent)
//
// Run with: go test -tags=live -run TestLiveTruncationShootout -v -timeout 10m
func TestLiveTruncationShootout(t *testing.T) {
	tok := os.Getenv("DEEPSEEK_TOKEN")
	if tok == "" {
		t.Skip("DEEPSEEK_TOKEN not set; skipping live truncation shootout")
	}

	for _, model := range []string{"expert", "instant"} {
		t.Run(model, func(t *testing.T) {
			runShootoutForModel(t, tok, model)
		})
	}
}

func runShootoutForModel(t *testing.T, tok, model string) {
	t.Helper()

	provider := deepseek.New()
	client, err := provider.NewClient(api.ProviderConfig{
		APIKey: tok,
		Model:  model,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ctx := context.Background()
	limit := client.MaxInputTokens()
	t.Logf("Model %q MaxInputTokens(): %d tokens", model, limit)

	// ── Phase 1: Find the actual rejection point ──────────────────────
	t.Log("── Phase 1: Finding rejection point ──")
	rejectPoint := findRejectionPoint(t, ctx, client, model)
	t.Logf("Rejection point: ~%d messages, ~%d tokens (%d chars)",
		rejectPoint.msgCount, rejectPoint.tokens, rejectPoint.chars)
	if rejectPoint.msgCount == 0 {
		t.Skip("Could not find rejection point (model may accept very large input)")
		return
	}

	// ── Phase 2: Test each truncation approach ────────────────────────
	t.Log("── Phase 2: Testing truncation approaches ──")

	// Generate messages at the rejection point (slightly above to force rejection)
	overLimit := rejectPoint.msgCount + 10
	overLimitMsgs := generateUniformConversation(overLimit)
	initialTokens := EstimateTokens(overLimitMsgs)
	initialChars := totalChars(overLimitMsgs)
	t.Logf("Generated %d messages (~%d tokens, %d chars) - should be rejected",
		len(overLimitMsgs), initialTokens, initialChars)

	// Verify baseline rejection
	baselineOK, baselineErr := sendAndCheck(t, ctx, client, model, overLimitMsgs)
	t.Logf("Baseline (no truncation): ok=%v err=%v", baselineOK, baselineErr)
	if baselineOK {
		t.Logf("WARNING: Baseline not rejected at %d msgs — model may accept more than expected", len(overLimitMsgs))
	}

	// Approach A: Simple FIFO - keep only recent messages
	t.Run("A_FIFO_Truncation", func(t *testing.T) {
		testFIFOTruncation(t, ctx, client, model, overLimitMsgs, limit, rejectPoint.msgCount)
	})

	// Approach B: Per-message truncation - truncate large messages
	t.Run("B_PerMessage_Truncation", func(t *testing.T) {
		testPerMessageTruncation(t, ctx, client, model, overLimitMsgs, limit)
	})

	// Approach C: Smart compaction (existing CompactSession)
	t.Run("C_Smart_Compaction", func(t *testing.T) {
		testSmartCompaction(t, ctx, client, model, overLimitMsgs, tok)
	})
}

// ── Rejection point discovery ────────────────────────────────────────

type rejectionPoint struct {
	msgCount int
	tokens   int
	chars    int
}

func findRejectionPoint(t *testing.T, ctx context.Context, client api.APIClient, model string) rejectionPoint {
	t.Helper()
	// Strategy: start from what we know is rejected, then binary search down.
	// Each message is ~34.5 tokens (from generateUniformConversation).
	// Start at a point comfortably above the model's max input tokens.

	modelLimit := client.MaxInputTokens()
	tokensPerMsg := 35 // approximate, ~34.5 from generateUniformConversation

	// Start at ~1.5x the model's token limit (guaranteed rejection)
	startMsgs := (modelLimit * 3 / 2) / tokensPerMsg
	if startMsgs < 32 {
		startMsgs = 32
	}
	if startMsgs > 10000 {
		startMsgs = 10000
	}
	t.Logf("Starting rejection search from %d msgs (~%d tokens, model limit=%d)",
		startMsgs, startMsgs*tokensPerMsg, modelLimit)

	// Find a point that IS rejected (walk up if needed)
	hi := startMsgs
	for {
		msgs := generateUniformConversation(hi)
		ok, err := sendAndCheck(t, ctx, client, model, msgs)
		if !ok {
			t.Logf("  confirmed rejected at %d msgs: %v", hi, truncateFirstLine(err.Error(), 80))
			break
		}
		t.Logf("  %d msgs accepted (unexpected), doubling...", hi)
		if hi > 20000 {
			t.Logf("No rejection at %d messages — model may have very large context", hi)
			return rejectionPoint{}
		}
		hi *= 2
		time.Sleep(1 * time.Second)
	}

	// Binary search down from hi to find the exact boundary.
	// lo = last known accepted, hi = first known rejected.
	lo := 1
	lastAccepted := rejectionPoint{}
	for lo <= hi {
		mid := (lo + hi) / 2
		msgs := generateUniformConversation(mid)
		ok, err := sendAndCheck(t, ctx, client, model, msgs)
		if ok {
			lastAccepted.msgCount = mid
			lastAccepted.tokens = EstimateTokens(msgs)
			lastAccepted.chars = totalChars(msgs)
			t.Logf("  accepted at %d msgs (~%d tokens)", mid, lastAccepted.tokens)
			lo = mid + 1
		} else {
			t.Logf("  rejected at %d msgs: %v", mid, truncateFirstLine(err.Error(), 80))
			hi = mid - 1
		}
		time.Sleep(1 * time.Second)
	}
	return lastAccepted
}

// ── Approach A: Simple FIFO ─────────────────────────────────────────

func testFIFOTruncation(t *testing.T, ctx context.Context, client api.APIClient, model string, msgs []api.Message, modelLimit, knownGood int) {
	t.Helper()
	t.Logf("Testing FIFO: %d messages -> keep last N", len(msgs))

	// Test various keep counts to find the sweet spot
	testCases := []struct {
		keep int
		desc string
	}{
		{keep: 5, desc: "very aggressive"},
		{keep: 10, desc: "aggressive (default)"},
		{keep: 20, desc: "moderate"},
		{keep: 50, desc: "generous"},
		{keep: knownGood / 2, desc: "half of known-good"},
		{keep: knownGood, desc: "known-good count"},
	}

	for _, tc := range testCases {
		if tc.keep <= 0 || tc.keep >= len(msgs) {
			continue
		}
		truncated := append([]api.Message{}, msgs[len(msgs)-tc.keep:]...)
		est := EstimateTokens(truncated)
		toksAsPct := float64(est) / float64(modelLimit) * 100

		ok, err := sendAndCheck(t, ctx, client, model, truncated)
		status := "✅"
		if !ok {
			status = "❌"
		}
		t.Logf("  FIFO keep=%d (%s): %d msgs, ~%d tokens (%.0f%% of limit) -> %s",
			tc.keep, tc.desc, len(truncated), est, toksAsPct, status)
		if err != nil {
			t.Logf("    error: %v", truncateFirstLine(err.Error(), 100))
		}
		time.Sleep(1 * time.Second)
	}
}

// ── Approach B: Per-message truncation ──────────────────────────────

func testPerMessageTruncation(t *testing.T, ctx context.Context, client api.APIClient, model string, msgs []api.Message, modelLimit int) {
	t.Helper()
	t.Logf("Testing per-message truncation: %d messages", len(msgs))

	// Strategy: truncate each message to a max size, keeping all messages
	// but shorter. Different max sizes per message.
	testCases := []struct {
		maxPerMsg  int
		desc       string
	}{
		{maxPerMsg: 50, desc: "very short"},
		{maxPerMsg: 100, desc: "short"},
		{maxPerMsg: 200, desc: "moderate"},
		{maxPerMsg: 500, desc: "generous"},
	}

	for _, tc := range testCases {
		truncated := make([]api.Message, len(msgs))
		for i, m := range msgs {
			truncated[i] = truncateAllBlocks(m, tc.maxPerMsg)
		}
		est := EstimateTokens(truncated)
		toksAsPct := float64(est) / float64(modelLimit) * 100

		ok, err := sendAndCheck(t, ctx, client, model, truncated)
		status := "✅"
		if !ok {
			status = "❌"
		}
		t.Logf("  Per-msg max=%d (%s): %d msgs, ~%d tokens (%.0f%% of limit) -> %s",
			tc.maxPerMsg, tc.desc, len(truncated), est, toksAsPct, status)
		if err != nil {
			t.Logf("    error: %v", truncateFirstLine(err.Error(), 100))
		}
		time.Sleep(1 * time.Second)
	}
}

func truncateAllBlocks(msg api.Message, maxChars int) api.Message {
	result := api.Message{Role: msg.Role, Content: make([]api.ContentBlock, len(msg.Content))}
	for i, cb := range msg.Content {
		result.Content[i] = cb
		switch cb.Type {
		case "text":
			if len(cb.Text) > maxChars {
				result.Content[i].Text = cb.Text[:maxChars] + "..."
			}
		case "tool_result":
			for j, inner := range cb.Content {
				if inner.Type == "text" && len(inner.Text) > maxChars {
					result.Content[i].Content[j].Text = inner.Text[:maxChars] + "..."
				}
			}
		}
	}
	return result
}

// ── Approach C: Smart Compaction ────────────────────────────────────

func testSmartCompaction(t *testing.T, ctx context.Context, client api.APIClient, model string, msgs []api.Message, tok string) {
	t.Helper()
	t.Logf("Testing smart compaction: %d messages", len(msgs))

	cfg := &Config{
		Model:                    model,
		MaxTokens:                2048,
		CompactionEnabled:        true,
		CompactionThreshold:      0.75,
		CompactionKeepRecent:     10,
		CompactionMaxInputTokens: client.MaxInputTokens(),
	}

	session := &Session{
		ID:       "shootout-test",
		Messages: append([]api.Message{}, msgs...),
	}

	summary, err := CompactSession(ctx, client, cfg, session)
	if err != nil {
		t.Logf("  ❌ Compaction failed: %v", err)
		return
	}

	t.Logf("  Summary: %d chars", len(summary))
	t.Logf("  Messages after: %d (was %d)", len(session.Messages), len(msgs))
	est := EstimateTokens(session.Messages)
	t.Logf("  Estimated tokens after: ~%d", est)

	// Send the compacted session (just the recent messages + the compaction
	// was already applied to the session). We add a final user message.
	finalMsgs := append([]api.Message{GetContinuationMessage(summary)}, session.Messages...)
	finalMsgs = append(finalMsgs, api.Message{
		Role: "user",
		Content: []api.ContentBlock{
			{Type: "text", Text: "What did we just discuss?"},
		},
	})

	ok, err := sendAndCheck(t, ctx, client, model, finalMsgs)
	status := "✅"
	if !ok {
		status = "❌"
	}
	t.Logf("  Compaction result: %s", status)
	if err != nil {
		t.Logf("    error: %v", truncateFirstLine(err.Error(), 100))
	}
}

// ── Phase 3: Head-to-head with model fallback ──────────────────────

// TestLiveTruncationRecovery tests that the auto-recovery path
// (isPromptTooLargeError → CompactNow) works end-to-end.
func TestLiveTruncationRecovery(t *testing.T) {
	tok := os.Getenv("DEEPSEEK_TOKEN")
	if tok == "" {
		t.Skip("DEEPSEEK_TOKEN not set")
	}

	provider := deepseek.New()
	client, err := provider.NewClient(api.ProviderConfig{
		APIKey: tok,
		Model:  "expert",
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ctx := context.Background()

	// Find rejection point for expert
	rp := findRejectionPoint(t, ctx, client, "expert")
	if rp.msgCount == 0 {
		t.Skip("Could not find rejection point")
	}
	t.Logf("Expert rejection at ~%d messages", rp.msgCount)

	// Create messages that will definitely be rejected
	overLimitMsgs := generateUniformConversation(rp.msgCount + 20)

	// Test that the auto-compact recovery works
	cfg := &Config{
		Model:                    "expert",
		MaxTokens:                4096,
		CompactionEnabled:        true,
		CompactionThreshold:      0.75,
		CompactionKeepRecent:     10,
		CompactionMaxInputTokens: 50000,
	}

	loop := &ConversationLoop{
		Client:  client,
		Session: &Session{ID: "recovery-test", Messages: overLimitMsgs},
		Config:  cfg,
	}

	// Try to send — should auto-compact and succeed
	userMsg := api.Message{
		Role: "user",
		Content: []api.ContentBlock{
			{Type: "text", Text: "Summarize our conversation."},
		},
	}
	loop.Session.Messages = append(loop.Session.Messages, userMsg)

	_, _, _, sendErr := loop.runOneTurn(ctx)
	if sendErr != nil {
		t.Logf("Initial send error (expected): %v", sendErr)

		if isPromptTooLargeError(sendErr) {
			t.Log("✅ Detected as prompt-too-large")
			summary, compactErr := loop.CompactNow(ctx)
			if compactErr != nil {
				t.Fatalf("CompactNow failed: %v", compactErr)
			}
			t.Logf("Compaction succeeded: %d chars summary, %d messages remain",
				len(summary), len(loop.Session.Messages))

			// Retry — should succeed
			_, _, _, retryErr := loop.runOneTurn(ctx)
			if retryErr != nil {
				t.Logf("❌ Retry after compaction failed: %v", retryErr)
			} else {
				t.Log("✅ Retry after compaction succeeded!")
			}
		}
	} else {
		t.Log("✅ Send succeeded without compaction (unexpected but OK)")
	}

	// Now try the recovery with model fallback to instant
	t.Log("── Testing model fallback (expert → instant) ──")
	instClient, err := provider.NewClient(api.ProviderConfig{
		APIKey: tok,
		Model:  "instant",
	})
	if err != nil {
		t.Fatalf("NewClient(instant): %v", err)
	}

	// Make a copy for instant test
	overLimitMsgs2 := generateUniformConversation(rp.msgCount + 20)
	loop2 := &ConversationLoop{
		Client:  instClient,
		Session: &Session{ID: "recovery-test-instant", Messages: overLimitMsgs2},
		Config:  cfg,
	}
	loop2.Session.Messages = append(loop2.Session.Messages, api.Message{
		Role: "user",
		Content: []api.ContentBlock{
			{Type: "text", Text: "Summarize our conversation."},
		},
	})

	// Compact first (since instant has bigger context, but still test compaction)
	summary, err := loop2.CompactNow(ctx)
	if err != nil {
		t.Logf("CompactNow on instant failed: %v", err)
	} else {
		t.Logf("CompactNow on instant: %d chars summary", len(summary))
		_, _, _, retryErr := loop2.runOneTurn(ctx)
		if retryErr != nil {
			t.Logf("❌ Retry after instant compaction failed: %v", retryErr)
		} else {
			t.Log("✅ Retry after instant compaction succeeded!")
		}
	}
}

// ── Helpers ─────────────────────────────────────────────────────────

func generateUniformConversation(n int) []api.Message {
	msgs := make([]api.Message, n)
	for i := 0; i < n; i++ {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		text := fmt.Sprintf("Message %d in the conversation. This is a %s message containing some content for testing context window limits with the DeepSeek API.", i+1, role)
		msgs[i] = api.Message{
			Role:    role,
			Content: []api.ContentBlock{{Type: "text", Text: text}},
		}
	}
	return msgs
}

func sendAndCheck(t *testing.T, ctx context.Context, client api.APIClient, model string, msgs []api.Message) (bool, error) {
	t.Helper()
	req := api.CreateMessageRequest{
		Model:     model,
		MaxTokens: 50,
		Messages:  msgs,
		Stream:    true,
	}

	ch, err := client.StreamResponse(ctx, req)
	if err != nil {
		return false, err
	}

	for event := range ch {
		if event.Type == api.EventError {
			return false, fmt.Errorf("stream error: %s", event.ErrorMessage)
		}
	}
	return true, nil
}

func totalChars(msgs []api.Message) int {
	total := 0
	for _, m := range msgs {
		for _, cb := range m.Content {
			switch cb.Type {
			case "text":
				total += len(cb.Text)
			}
			for _, inner := range cb.Content {
				total += len(inner.Text)
			}
		}
	}
	return total
}

func truncateFirstLine(s string, maxLen int) string {
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		s = s[:idx]
	}
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}