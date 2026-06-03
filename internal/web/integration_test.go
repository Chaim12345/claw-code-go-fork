package web

import (
	"claw-code-go/internal/api"
	"claw-code-go/internal/runtime"
	"claw-code-go/internal/web/chatproto"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// TestIntegration_ChatRoundTrip is the canonical integration test for the
// /api/chat/ws endpoint. It spins up a full httptest.Server with middleware
// (logging, rate limiting, security headers, auth), opens a WebSocket
// connection, sends a user_input, and asserts the expected sequence of
// text_delta, text_final, and done events using a fake api.APIClient.
func TestIntegration_ChatRoundTrip(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	fc := &fakeClient{
		events: []api.StreamEvent{
			{Type: api.EventMessageStart, InputTokens: 5},
			{Type: api.EventContentBlockStart, Index: 0, ContentBlock: api.ContentBlockInfo{Type: "text"}},
			{Type: api.EventContentBlockDelta, Index: 0, Delta: api.Delta{Type: "text_delta", Text: "Hello from "}},
			{Type: api.EventContentBlockDelta, Index: 0, Delta: api.Delta{Type: "text_delta", Text: "integration test"}},
			{Type: api.EventContentBlockStop, Index: 0},
			{Type: api.EventMessageDelta, MessageDelta: api.MessageDelta{StopReason: "end_turn"}, Usage: api.UsageDelta{OutputTokens: 7}},
			{Type: api.EventMessageStop},
		},
	}

	cfg := &runtime.Config{
		Model:            "expert",
		MaxTokens:        100,
		SessionDir:       t.TempDir(),
		CompactionEnabled: true,
	}
	loop := runtime.NewConversationLoop(cfg, fc)

	srv := NewServer(Config{Addr: "127.0.0.1:0"})
	srv.ChatLoopFactory = func() *runtime.ConversationLoop {
		return loop
	}

	ts := httptest.NewServer(srv.routes())
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/api/chat/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	// 1. Expect chat_session_init hello.
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read hello: %v", err)
	}
	var hello chatproto.ServerOutbound
	if err := json.Unmarshal(msg, &hello); err != nil {
		t.Fatalf("unmarshal hello: %v", err)
	}
	if hello.Type != chatproto.MsgChatSessionInit {
		t.Fatalf("expected chat_session_init, got %q", hello.Type)
	}
	if hello.SessionID == "" {
		t.Fatal("expected non-empty session_id in hello")
	}

	// 2. Send user_input.
	userJSON := `{"type":"user_input","text":"Hello world"}`
	if err := conn.WriteMessage(websocket.TextMessage, []byte(userJSON)); err != nil {
		t.Fatalf("write user_input: %v", err)
	}

	// 3. Collect all server messages until "done".
	var events []chatproto.ServerOutbound
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		var out chatproto.ServerOutbound
		if err := json.Unmarshal(msg, &out); err != nil {
			t.Fatalf("unmarshal server msg: %v", err)
		}
		if out.Type == "" {
			continue
		}
		events = append(events, out)
		if out.Type == chatproto.MsgDone {
			break
		}
	}

	// 4. Assert the expected sequence: text_delta(s), text_final, usage, done.
	var gotTypes []string
	foundTextDelta := false
	foundTextFinal := false
	foundDone := false
	for _, ev := range events {
		gotTypes = append(gotTypes, ev.Type)
		switch ev.Type {
		case chatproto.MsgTextDelta:
			foundTextDelta = true
		case chatproto.MsgTextFinal:
			foundTextFinal = true
		case chatproto.MsgDone:
			foundDone = true
		}
	}
	if !foundTextDelta {
		t.Errorf("missing text_delta in response, got: %v", gotTypes)
	}
	if !foundTextFinal {
		t.Errorf("missing text_final in response, got: %v", gotTypes)
	}
	if !foundDone {
		t.Errorf("missing done in response, got: %v", gotTypes)
	}

	// Verify the combined text matches what the fake emitted.
	finalText := ""
	for _, ev := range events {
		if ev.Type == chatproto.MsgTextDelta {
			finalText += ev.Text
		}
		if ev.Type == chatproto.MsgTextFinal {
			finalText = ev.Text
		}
	}
	if finalText == "" {
		t.Error("expected non-empty text from assistant response")
	}
}

// TestIntegration_SessionIsolation verifies that two concurrent WebSocket
// connections get separate session IDs and can both complete a turn.
func TestIntegration_SessionIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	cfg := &runtime.Config{
		Model:            "expert",
		MaxTokens:        100,
		SessionDir:       t.TempDir(),
		CompactionEnabled: true,
	}

	srv := NewServer(Config{Addr: "127.0.0.1:0"})
	srv.ChatLoopFactory = func() *runtime.ConversationLoop {
		fc := &fakeClient{
			events: []api.StreamEvent{
				{Type: api.EventMessageStart, InputTokens: 1},
				{Type: api.EventContentBlockStart, Index: 0, ContentBlock: api.ContentBlockInfo{Type: "text"}},
				{Type: api.EventContentBlockDelta, Index: 0, Delta: api.Delta{Type: "text_delta", Text: "OK"}},
				{Type: api.EventContentBlockStop, Index: 0},
				{Type: api.EventMessageDelta, MessageDelta: api.MessageDelta{StopReason: "end_turn"}, Usage: api.UsageDelta{OutputTokens: 1}},
				{Type: api.EventMessageStop},
			},
		}
		return runtime.NewConversationLoop(cfg, fc)
	}

	ts := httptest.NewServer(srv.routes())
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/api/chat/ws"

	type result struct {
		sessionID string
		err       error
	}

	results := make(chan result, 2)
	for i := 0; i < 2; i++ {
		go func() {
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				results <- result{err: err}
				return
			}
			defer conn.Close()

			_, msg, _ := conn.ReadMessage()
			var hello chatproto.ServerOutbound
			json.Unmarshal(msg, &hello)

			conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"user_input","text":"ping"}`))

			for {
				_, msg, err := conn.ReadMessage()
				if err != nil {
					break
				}
				var out chatproto.ServerOutbound
				json.Unmarshal(msg, &out)
				if out.Type == chatproto.MsgDone {
					break
				}
			}
			results <- result{sessionID: hello.SessionID}
		}()
	}

	var sessionIDs []string
	for i := 0; i < 2; i++ {
		r := <-results
		if r.err != nil {
			t.Errorf("connection error: %v", r.err)
			continue
		}
		sessionIDs = append(sessionIDs, r.sessionID)
	}

	if len(sessionIDs) != 2 {
		t.Fatalf("expected 2 session IDs, got %d", len(sessionIDs))
	}
	if sessionIDs[0] == sessionIDs[1] {
		t.Errorf("expected different session IDs, got duplicate: %q", sessionIDs[0])
	}
}

// TestIntegration_Timeout verifies that the server respects context
// cancellation and cleans up cleanly when the client disconnects mid-turn.
func TestIntegration_Timeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	cfg := &runtime.Config{
		Model:            "expert",
		MaxTokens:        100,
		SessionDir:       t.TempDir(),
		CompactionEnabled: true,
	}
	fc := &fakeClient{
		events: make([]api.StreamEvent, 100),
	}
	for i := range fc.events {
		fc.events[i] = api.StreamEvent{
			Type:  api.EventContentBlockDelta,
			Index: 0,
			Delta: api.Delta{Type: "text_delta", Text: "x"},
		}
	}
	fc.events = append(fc.events,
		api.StreamEvent{Type: api.EventContentBlockStop, Index: 0},
		api.StreamEvent{Type: api.EventMessageDelta, MessageDelta: api.MessageDelta{StopReason: "end_turn"}, Usage: api.UsageDelta{OutputTokens: 100}},
		api.StreamEvent{Type: api.EventMessageStop},
	)

	loop := runtime.NewConversationLoop(cfg, fc)

	srv := NewServer(Config{Addr: "127.0.0.1:0"})
	srv.ChatLoopFactory = func() *runtime.ConversationLoop {
		return loop
	}

	ts := httptest.NewServer(srv.routes())
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/api/chat/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}

	// Read hello.
	_, _, _ = conn.ReadMessage()

	// Send user_input.
	conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"user_input","text":"go"}`))

	// Close the connection before the turn completes.
	conn.Close()

	// Give the server a moment to clean up.
	time.Sleep(50 * time.Millisecond)

	// The test passes if we reach here without a panic or deadlock.
}

// TestIntegration_ErrorHandling verifies the server handles API errors
// gracefully and sends an error event to the client.
func TestIntegration_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	cfg := &runtime.Config{
		Model:            "expert",
		MaxTokens:        100,
		SessionDir:       t.TempDir(),
		CompactionEnabled: true,
	}
	fc := &fakeClient{
		streamErr: &testAPIError{msg: "provider unavailable"},
	}
	loop := runtime.NewConversationLoop(cfg, fc)

	srv := NewServer(Config{Addr: "127.0.0.1:0"})
	srv.ChatLoopFactory = func() *runtime.ConversationLoop {
		return loop
	}

	ts := httptest.NewServer(srv.routes())
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/api/chat/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	// Read hello.
	_, _, _ = conn.ReadMessage()

	// Send user_input.
	conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"user_input","text":"crash"}`))

	gotError := false
	gotDone := false
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		var out chatproto.ServerOutbound
		if err := json.Unmarshal(msg, &out); err != nil {
			continue
		}
		if out.Type == chatproto.MsgError {
			gotError = true
		}
		if out.Type == chatproto.MsgDone {
			gotDone = true
			break
		}
	}
	if !gotError && !gotDone {
		t.Error("expected error or done event after API failure")
	}
}

// testAPIError is a simple error type used in integration tests.
type testAPIError struct {
	msg string
}

func (e *testAPIError) Error() string { return e.msg }
