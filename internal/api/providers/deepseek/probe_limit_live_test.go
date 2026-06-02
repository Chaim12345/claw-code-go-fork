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
// Method:
//   - Pick a target model.
//   - For each candidate prompt size, send a single user message and
//     wait for either a normal stream-start (request_message_id frame)
//     or an error.
//   - Binary search for the largest size that succeeds.
func TestLiveProbeActualInputLimit(t *testing.T) {
	if os.Getenv("DEEPSEEK_TOKEN") == "" {
		t.Skip("DEEPSEEK_TOKEN not set; skipping live test")
	}

	cases := []struct {
		name      string
		model     string
		lowOK     int    // size known to succeed (from a prior run, ~100 chars)
		highFail  int    // size well above /settings limit, must fail
		settingCh int    // configured input_character_limit from /settings
	}{
		{name: "instant", model: "instant", lowOK: 100, highFail: 3_000_000, settingCh: 2_621_440},
		{name: "expert", model: "expert", lowOK: 100, highFail: 200_000, settingCh: 163_840},
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

			// Verify the configured /settings value matches what we expect
			// so we know our test harness is correct.
			settings, err := dsClient.fetchSettings()
			if err != nil {
				t.Fatalf("fetchSettings: %v", err)
			}
			cfg, ok := settings[tc.model]
			if !ok {
				// Map "instant" -> "default" and "expert" -> "expert"
				if alt, ok2 := settings["default"]; ok2 && tc.model == "instant" {
					cfg = alt
				} else {
					t.Fatalf("settings missing model %q (have: %v)", tc.model, settings)
				}
			}
			t.Logf("[%s] /settings input_character_limit = %d", tc.name, cfg.InputCharacterLimit)

			// Start with a fresh chat session per model.
			sid, err := dsClient.web.CreateChatSession()
			if err != nil {
				t.Fatalf("CreateChatSession: %v", err)
			}
			dsClient.chatSessionID = sid
			dsClient.parentMessageID = ""

			// Binary search between lowOK and highFail for the largest
			// size that produces a successful stream.
			lo, hi := tc.lowOK, tc.highFail
			lastOK := lo
			for hi-lo > maxInt(hi/20, 1000) {
				mid := (lo + hi) / 2
				ok, msg := probeSize(t, dsClient, tc.model, mid)
				t.Logf("[%s] probe size=%d -> ok=%v (%s)", tc.name, mid, ok, msg)
				if ok {
					lastOK = mid
					lo = mid
				} else {
					hi = mid
				}
			}

			// Take the larger of: configured limit, observed lastOK.
			// The probe's lastOK is the largest size that succeeded.
			// We log both so we can see if the server's enforced limit
			// matches the /settings value.
			t.Logf("[%s] /settings says: %d chars, server actually accepted: %d chars",
				tc.name, cfg.InputCharacterLimit, lastOK)
			if lastOK < cfg.InputCharacterLimit/2 {
				t.Errorf("[%s] server-enforced limit (%d) is less than half of /settings value (%d) — possible new server-side cap",
					tc.name, lastOK, cfg.InputCharacterLimit)
			}
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
	return gotText, "streamed"
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
