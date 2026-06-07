//go:build live

package deepseek

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"claw-code-go/internal/api"
)

// TestLiveProbeActualInputLimit is a binary-search probe that measures the
// real per-message character limit the DeepSeek web API enforces, instead
// of trusting the input_character_limit value in /api/v0/client/settings.
//
// Why: the /settings endpoint publishes 2,621,440 for "default" and
// 163,840 for "expert", but that's a configured knob — the server may
// actually cap lower (e.g., per-account limits, current server load,
// per-message token ceilings that the configured char count doesn't
// translate to linearly). This probe establishes ground truth by
// actually sending prompts and observing success/failure.
//
// Direction: probes from LARGE to SMALL. Each model case starts with a
// `ceiling` (best-guess upper bound from /settings or simply "very large")
// and tests that first. If the ceiling is accepted, the model's true limit
// is at least that high. If the ceiling is rejected, we binary-search DOWN
// between `floor` (smallest known good) and `ceiling` to find the largest
// size the server actually accepts.
//
// Why large-to-small: gives a strong first signal (ceiling fails) that
// validates the harness is exercising the right code path, and converges
// to a usable "largest OK" value in O(log n) probes.
//
// Method:
//   - Pick a target model.
//   - For each candidate prompt size, send a single user message via the
//     real StreamResponse path (the same call the agent uses) and wait
//     for either a normal message_start event or a server error.
//   - Report the largest size that succeeds and the smallest that fails.
func TestLiveProbeActualInputLimit(t *testing.T) {
	if os.Getenv("DEEPSEEK_TOKEN") == "" {
		t.Skip("DEEPSEEK_TOKEN not set; skipping live test")
	}

	cases := []struct {
		name  string
		model string
		floor int // smallest known good (probe won't go below this)
		ceil  int // probe this first; if it fails, binary-search down to floor
	}{
		{name: "default", model: "default", floor: 100, ceil: 3_000_000},
		{name: "expert", model: "expert", floor: 100, ceil: 200_000},
		{name: "vision", model: "vision", floor: 100, ceil: 3_000_000},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			provider := New()
			client, err := provider.NewClient(api.ProviderConfig{
				APIKey: os.Getenv("DEEPSEEK_TOKEN"),
				Model:  tc.model,
			})
			if err != nil {
				t.Fatalf("NewClient: %v", err)
			}
			dsClient, ok := client.(*Client)
			if !ok {
				t.Skip("client is not *deepseek.Client")
			}

			// Best-effort: log /settings if available, but don't fail the
			// test if it's broken (the endpoint currently returns
			// biz_code:2 INVALID_PARAM).
			if settings, err := dsClient.fetchSettings(); err == nil {
				if cfg, ok := settings[tc.model]; ok {
					t.Logf("[%s] /settings input_character_limit = %d", tc.name, cfg.InputCharacterLimit)
				}
			} else {
				t.Logf("[%s] /settings unavailable: %v (proceeding with probe)", tc.name, err)
			}

			// Start with a fresh chat session per model.
			sid, err := dsClient.web.CreateChatSession()
			if err != nil {
				t.Fatalf("CreateChatSession: %v", err)
			}
			dsClient.chatSessionID = sid
			dsClient.parentMessageID = ""
			// Bypass the client-side pre-flight cap (provider.go:270) so we
			// measure the SERVER's enforcement, not the client's estimate.
			// In delta mode the pre-flight is skipped; the prompt we send
			// is still a single user message (buildPrompt path) since
			// parentMessageID is empty.
			dsClient.deltaMode = true

			// Phase 1: probe the ceiling first.
			t.Logf("[%s] ceiling probe: size=%d", tc.name, tc.ceil)
			ok, msg := probeSize(t, dsClient, tc.model, tc.ceil)
			t.Logf("[%s] size=%d -> ok=%v (%s)", tc.name, tc.ceil, ok, msg)
			if ok {
				t.Logf("[%s] RESULT: limit >= %d chars (ceiling accepted, no upper bound within tested range)",
					tc.name, tc.ceil)
				return
			}

			// Phase 2: ceiling failed. Binary-search DOWN between floor and
			// ceiling to find the largest size that succeeds.
			lo, hi := tc.floor, tc.ceil
			lastOK := lo
			attempts := 1
			for hi-lo > maxInt(hi/20, 1000) {
				mid := (lo + hi) / 2
				attempts++
				ok, msg := probeSize(t, dsClient, tc.model, mid)
				t.Logf("[%s] attempt %d: size=%d -> ok=%v (%s)", tc.name, attempts, mid, ok, msg)
				if ok {
					lastOK = mid
					lo = mid
				} else {
					hi = mid
				}
				if attempts > 12 {
					t.Logf("[%s] (stopping after %d attempts)", tc.name, attempts)
					break
				}
			}
			t.Logf("[%s] RESULT: largest OK = %d chars, smallest fail = %d chars (gap = %d)",
				tc.name, lastOK, hi, hi-lastOK)
		})
	}
}

// probeSize sends a single message of `size` bytes (ASCII filler) and
// returns whether the server accepted the request (started streaming a
// response) or rejected it (returned an error).
func probeSize(t *testing.T, c *Client, model string, size int) (bool, string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	prompt := strings.Repeat("a", size)
	req := api.CreateMessageRequest{
		Model:     model,
		MaxTokens: 32,
		System:    "You are a test harness. Reply with the single token: ok",
		Messages:  []api.Message{{Role: "user", Content: []api.ContentBlock{{Type: "text", Text: prompt}}}},
	}

	events, err := c.StreamResponse(ctx, req)
	if err != nil {
		// Pre-flight rejection (caller-side cap) is also a useful signal.
		return false, fmt.Sprintf("preflight err: %v", err)
	}

	gotStart := false
	gotErr := ""
	gotText := false
	for ev := range events {
		switch ev.Type {
		case api.EventMessageStart:
			gotStart = true
		case api.EventContentBlockDelta:
			if ev.Delta.Type == "text_delta" {
				gotText = true
			}
		case api.EventError:
			gotErr = ev.ErrorMessage
		}
	}
	if gotErr != "" {
		return false, "server error: " + gotErr
	}
	if !gotStart {
		return false, "no message_start (no accept, no error)"
	}
	// We only care whether the server accepted the request (gotStart).
	// Whether text streamed back is a property of model speed/queue, not
	// input-size enforcement, so report success on start.
	_ = gotText
	return gotStart, "streamed"
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
