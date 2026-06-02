//go:build live

package deepseek

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// TestLiveExploreRawStream calls DeepSeek directly and dumps the raw
// stream so we can see exactly what events, headers, and fields the
// server actually emits. Useful for mapping the API surface and
// deciding what to put in EventMessageStart.InputTokens /
// EventMessageDelta.Usage.
func TestLiveExploreRawStream(t *testing.T) {
	tok := os.Getenv("DEEPSEEK_TOKEN")
	if tok == "" {
		t.Skip("DEEPSEEK_TOKEN not set")
	}

	wc := NewWebClient(tok)

	// Step 1: open a session.
	sid, err := wc.CreateChatSession()
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	t.Logf("session id: %s", sid)

	// Step 2: dump the model-config table.
	settings, err := wc.FetchModelSettings()
	if err != nil {
		t.Fatalf("fetch settings: %v", err)
	}
	for k, v := range settings {
		b, _ := json.MarshalIndent(v, "", "  ")
		t.Logf("config[%s] = %s", k, string(b))
	}

	// Step 3: drive a stream and capture every line the server sends.
	// We do this by temporarily swapping the handler in ChatCompletionStream
	// with one that records raw lines.
	rec := &recorder{}
	_, _ = wc.ChatCompletionStream(
		CompletionOpts{
			SessionID: sid,
			Prompt:    "What is 6 * 7? One sentence.",
			Spec:      ModelSpec{ModelType: "default"},
		},
		rec.handler,
	)

	if rec.text.Len() == 0 {
		t.Fatal("no content received")
	}
	t.Logf("raw assembled text: %q", rec.text.String())
	t.Logf("saw %d message_id events", rec.msgIDCount)

	// Verify our existing parser can read the events.
	gotText := ExtractToolCalls(rec.text.String())
	t.Logf("tool calls extracted: %d", len(gotText))
}

type recorder struct {
	text        strings.Builder
	rawLines    []string
	msgIDCount  int
}

func (r *recorder) handler(ev StreamEvent) bool {
	if ev.Event == "content" {
		r.text.WriteString(ev.Data)
		return true
	}
	if ev.Event == "message_id" {
		r.msgIDCount++
		return true
	}
	if ev.Event == "debug" {
		// Capture full debug line for inspection.
		r.rawLines = append(r.rawLines, ev.Data)
	}
	return true
}
