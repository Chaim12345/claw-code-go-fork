//go:build live

package deepseek

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"claw-code-go/internal/api"
)

// liveTestTimeout caps each live API call so a stuck PoW or network hang
// doesn't run away with the test process.
const liveTestTimeout = 60 * time.Second

// skipIfNoToken short-circuits live tests when DEEPSEEK_TOKEN isn't set.
// CI without secrets should `go test -tags live -run ^$ ./...` to skip.
func skipIfNoToken(t *testing.T) string {
	t.Helper()
	tok := os.Getenv("DEEPSEEK_TOKEN")
	if tok == "" {
		t.Skip("DEEPSEEK_TOKEN not set; skipping live test")
	}
	return tok
}

func TestLiveFetchModelSettings(t *testing.T) {
	skipIfNoToken(t)
	wc := NewWebClient(os.Getenv("DEEPSEEK_TOKEN"))

	ctx, cancel := context.WithTimeout(context.Background(), liveTestTimeout)
	defer cancel()

	done := make(chan struct {
		settings map[string]ModelConfig
		err      error
	}, 1)
	go func() {
		s, e := wc.FetchModelSettings()
		done <- struct {
			settings map[string]ModelConfig
			err      error
		}{s, e}
	}()

	select {
	case r := <-done:
		if r.err != nil {
			t.Fatalf("FetchModelSettings: %v", r.err)
		}
		if len(r.settings) == 0 {
			t.Fatal("expected at least one model config")
		}
		// The web API publishes Instant / Expert / Vision configs.
		for _, want := range []string{"default", "expert", "vision"} {
			if _, ok := r.settings[want]; !ok {
				t.Errorf("expected config for model_type %q", want)
			}
		}
		// Each config should have a positive input limit.
		for k, cfg := range r.settings {
			if cfg.InputCharacterLimit <= 0 {
				t.Errorf("config %q has non-positive input_character_limit", k)
			}
		}
		t.Logf("fetched %d model configs", len(r.settings))
	case <-ctx.Done():
		t.Fatal("FetchModelSettings timed out")
	}
}

func TestLiveCreateChatSession(t *testing.T) {
	skipIfNoToken(t)
	wc := NewWebClient(os.Getenv("DEEPSEEK_TOKEN"))

	ctx, cancel := context.WithTimeout(context.Background(), liveTestTimeout)
	defer cancel()

	type result struct {
		sid string
		err error
	}
	done := make(chan result, 1)
	go func() {
		sid, err := wc.CreateChatSession()
		done <- result{sid, err}
	}()

	select {
	case r := <-done:
		if r.err != nil {
			t.Fatalf("CreateChatSession: %v", r.err)
		}
		if r.sid == "" {
			t.Fatal("expected non-empty session ID")
		}
		t.Logf("created session: %s", r.sid)
	case <-ctx.Done():
		t.Fatal("CreateChatSession timed out")
	}
}

func TestLiveStreamResponseSimple(t *testing.T) {
	skipIfNoToken(t)
	provider := New()
	client, err := provider.NewClient(api.ProviderConfig{
		APIKey: os.Getenv("DEEPSEEK_TOKEN"),
		Model:  "instant", // default fast variant — most likely to succeed
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), liveTestTimeout)
	defer cancel()

	req := api.CreateMessageRequest{
		Model:     "instant",
		MaxTokens: 256,
		System:    "You are a concise assistant. Reply in one sentence.",
		Messages: []api.Message{
			{Role: "user", Content: []api.ContentBlock{{Type: "text", Text: "What is 2+2?"}}},
		},
	}

	events, err := client.StreamResponse(ctx, req)
	if err != nil {
		t.Fatalf("StreamResponse: %v", err)
	}

	var gotText strings.Builder
	var gotStart, gotStop, gotBlockStart, gotBlockStop bool
	var inputTokens int

	for ev := range events {
		switch ev.Type {
		case api.EventMessageStart:
			gotStart = true
			inputTokens = ev.InputTokens
		case api.EventMessageStop:
			gotStop = true
		case api.EventContentBlockStart:
			gotBlockStart = true
		case api.EventContentBlockStop:
			gotBlockStop = true
		case api.EventContentBlockDelta:
			if ev.Delta.Type == "text_delta" {
				gotText.WriteString(ev.Delta.Text)
			}
		case api.EventError:
			t.Fatalf("got error event: %s", ev.ErrorMessage)
		}
	}

	if !gotStart {
		t.Error("missing message_start event")
	}
	if !gotStop {
		t.Error("missing message_stop event")
	}
	if !gotBlockStart || !gotBlockStop {
		t.Errorf("missing content block events (start=%v stop=%v)", gotBlockStart, gotBlockStop)
	}
	if gotText.Len() == 0 {
		t.Error("expected at least one text delta")
	}
	if inputTokens <= 0 {
		t.Errorf("expected positive input token estimate, got %d", inputTokens)
	}
	t.Logf("response: %q", gotText.String())
	t.Logf("estimated input tokens: %d", inputTokens)
}

func TestLiveStreamResponseExpert(t *testing.T) {
	skipIfNoToken(t)
	provider := New()
	client, err := provider.NewClient(api.ProviderConfig{
		APIKey: os.Getenv("DEEPSEEK_TOKEN"),
		Model:  "expert",
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*liveTestTimeout)
	defer cancel()

	// Force a plain answer: explicitly forbid tool calls and ask for
	// a single number. The expert model otherwise tries to call
	// "calc_art" or similar (treating the request as a tool call).
	req := api.CreateMessageRequest{
		Model:     "expert",
		MaxTokens: 512,
		System:    "You are a math assistant. Answer in plain text. Do not use any tool. Do not output JSON.",
		Messages: []api.Message{
			{Role: "user", Content: []api.ContentBlock{{Type: "text", Text: "Compute 12 * 13. Reply with just the number, nothing else."}}},
		},
	}

	events, err := client.StreamResponse(ctx, req)
	if err != nil {
		t.Fatalf("StreamResponse: %v", err)
	}

	var gotText strings.Builder
	for ev := range events {
		if ev.Type == api.EventContentBlockDelta && ev.Delta.Type == "text_delta" {
			gotText.WriteString(ev.Delta.Text)
		}
		if ev.Type == api.EventError {
			t.Fatalf("error: %s", ev.ErrorMessage)
		}
	}

	if gotText.Len() == 0 {
		t.Fatal("no text received from expert model")
	}
	t.Logf("expert response: %q", gotText.String())
	if !strings.Contains(gotText.String(), "156") {
		t.Errorf("expected response to contain '156', got %q", gotText.String())
	}
}

func TestLiveToolCallExtraction(t *testing.T) {
	skipIfNoToken(t)
	provider := New()
	client, err := provider.NewClient(api.ProviderConfig{
		APIKey: os.Getenv("DEEPSEEK_TOKEN"),
		Model:  "instant",
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), liveTestTimeout)
	defer cancel()

	// Ask the model to call a tool — system prompt instructs it to.
	req := api.CreateMessageRequest{
		Model:     "instant",
		MaxTokens: 512,
		System: `You have access to a tool called "bash". When you need to run a command, output a JSON object like:
{"tool_calls":[{"name":"bash","arguments":{"command":"..."}}]}
Do not use markdown fences. Output raw JSON on its own line.`,
		Messages: []api.Message{
			{Role: "user", Content: []api.ContentBlock{{Type: "text", Text: "Please list the files in the current directory by running `ls`."}}},
		},
	}

	events, err := client.StreamResponse(ctx, req)
	if err != nil {
		t.Fatalf("StreamResponse: %v", err)
	}

	var gotText strings.Builder
	for ev := range events {
		if ev.Type == api.EventContentBlockDelta && ev.Delta.Type == "text_delta" {
			gotText.WriteString(ev.Delta.Text)
		}
		if ev.Type == api.EventError {
			t.Fatalf("error: %s", ev.ErrorMessage)
		}
	}

	text := gotText.String()
	t.Logf("raw response: %s", text)
	calls := ExtractToolCalls(text)
	if len(calls) == 0 {
		t.Skipf("model did not emit a recognisable tool call in this run; raw=%q", text)
	}
	if calls[0].Name != "bash" {
		t.Errorf("expected tool name 'bash', got %q", calls[0].Name)
	}
	t.Logf("extracted tool call: %s(%v)", calls[0].Name, calls[0].Arguments)
}

func TestLiveModelLimitEnforced(t *testing.T) {
	skipIfNoToken(t)
	provider := New()
	client, err := provider.NewClient(api.ProviderConfig{
		APIKey: os.Getenv("DEEPSEEK_TOKEN"),
		Model:  "instant",
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	// Force a deliberately oversized prompt that exceeds the cap for
	// 'instant' (most variants cap around 28-60k tokens). This proves
	// the pre-flight limit check fires *before* a network round-trip.
	big := strings.Repeat("filler word here. ", 200_000) // ~2.4M chars ~ 600k tokens
	dsClient, ok := client.(*Client)
	if !ok {
		t.Skip("client is not *deepseek.Client; cannot access limit internals")
	}

	settings, err := dsClient.fetchSettings()
	if err != nil {
		t.Fatalf("fetchSettings: %v", err)
	}
	cap := settings["default"].MaxInputTokens()
	if cap == 0 {
		t.Skip("server didn't advertise a limit; cannot exercise pre-flight check")
	}

	estimated := EstimateTokens(big)
	if estimated <= cap {
		t.Fatalf("test setup wrong: estimated %d tokens, cap is %d", estimated, cap)
	}

	req := api.CreateMessageRequest{
		Model:     "instant",
		MaxTokens: 256,
		System:    "x",
		Messages:  []api.Message{{Role: "user", Content: []api.ContentBlock{{Type: "text", Text: big}}}},
	}

	_, err = client.StreamResponse(context.Background(), req)
	if err == nil {
		t.Fatal("expected pre-flight limit error, got nil")
	}
	if !strings.Contains(err.Error(), "prompt too large") {
		t.Errorf("expected 'prompt too large' error, got %v", err)
	}
	t.Logf("got expected error: %v", err)
}
