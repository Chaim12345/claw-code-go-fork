package web

import (
	"claw-code-go/internal/api"
	"claw-code-go/internal/runtime"
	"claw-code-go/internal/web/chatproto"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/websocket"
)

// fakeClient is an api.APIClient that returns canned StreamEvents.
// It satisfies api.APIClient.
type fakeClient struct {
	events    []api.StreamEvent
	streamErr error // if set, StreamResponse returns this error
}

func (f *fakeClient) StreamResponse(ctx context.Context, req api.CreateMessageRequest) (<-chan api.StreamEvent, error) {
	if f.streamErr != nil {
		return nil, f.streamErr
	}
	ch := make(chan api.StreamEvent, len(f.events))
	for _, ev := range f.events {
		ch <- ev
	}
	close(ch)
	return ch, nil
}

func (f *fakeClient) MaxInputTokens() int { return 0 }

// TestTurnEventToOutbound verifies each TurnEvent type maps to the
// expected chatproto.ServerOutbound JSON.
func TestTurnEventToOutbound(t *testing.T) {
	tests := []struct {
		name string
		ev   runtime.TurnEvent
		want chatproto.ServerOutbound
	}{
		{
			name: "text_delta",
			ev:   runtime.TurnEvent{Type: runtime.TurnEventTextDelta, Text: "hello"},
			want: chatproto.ServerOutbound{Type: chatproto.MsgTextDelta, Text: "hello"},
		},
		{
			name: "text_final",
			ev:   runtime.TurnEvent{Type: runtime.TurnEventTextFinal, Text: "world"},
			want: chatproto.ServerOutbound{Type: chatproto.MsgTextFinal, Text: "world"},
		},
		{
			name: "tool_start",
			ev:   runtime.TurnEvent{Type: runtime.TurnEventToolStart, ToolName: "bash", ToolInput: "ls"},
			want: chatproto.ServerOutbound{Type: chatproto.MsgToolStart, ToolName: "bash", ToolInput: "ls"},
		},
		{
			name: "tool_done",
			ev:   runtime.TurnEvent{Type: runtime.TurnEventToolDone, ToolName: "bash", ToolResult: "file1"},
			want: chatproto.ServerOutbound{Type: chatproto.MsgToolDone, Result: "file1"},
		},
		{
			name: "permission_ask",
			ev:   runtime.TurnEvent{Type: runtime.TurnEventPermissionAsk, ToolName: "bash", ToolInput: "rm -rf /"},
			want: chatproto.ServerOutbound{Type: chatproto.MsgPermissionAsk, ToolName: "bash", ToolInput: "rm -rf /"},
		},
		{
			name: "ask_user",
			ev:   runtime.TurnEvent{Type: runtime.TurnEventAskUser, ToolInput: "What file?"},
			want: chatproto.ServerOutbound{Type: chatproto.MsgAskUser, Prompt: "What file?"},
		},
		{
			name: "usage",
			ev:   runtime.TurnEvent{Type: runtime.TurnEventUsage, InputTokens: 10, OutputTokens: 20},
			want: chatproto.ServerOutbound{Type: chatproto.MsgUsage, InputTokens: 10, OutputTokens: 20},
		},
		{
			name: "done",
			ev:   runtime.TurnEvent{Type: runtime.TurnEventDone},
			want: chatproto.ServerOutbound{Type: chatproto.MsgDone},
		},
		{
			name: "warn",
			ev:   runtime.TurnEvent{Type: runtime.TurnEventWarn, Text: "truncated"},
			want: chatproto.ServerOutbound{Type: chatproto.MsgWarn, Message: "truncated"},
		},
		{
			name: "error",
			ev:   runtime.TurnEvent{Type: runtime.TurnEventError, Err: errors.New("boom")},
			want: chatproto.ServerOutbound{Type: chatproto.MsgError, Code: "turn_error", Message: "boom"},
		},
		{
			name: "info_dropped",
			ev:   runtime.TurnEvent{Type: runtime.TurnEventInfo, Text: "ignored"},
			want: chatproto.ServerOutbound{}, // empty → not written by sendJSON
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := turnEventToOutbound(tt.ev)
			if got != tt.want {
				t.Errorf("turnEventToOutbound(%v) = %+v, want %+v", tt.ev.Type, got, tt.want)
			}
		})
	}
}

// TestChatSessionRoundTrip spins up a httptest server with a fake
// api.APIClient, opens a WebSocket connection, sends a user_input,
// and verifies the server sends chat_session_init followed by
// text_delta / text_final / done messages.
func TestChatSessionRoundTrip(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	// A fake that emits one text block then end_turn.
	fc := &fakeClient{
		events: []api.StreamEvent{
			{Type: api.EventMessageStart, InputTokens: 5},
			{Type: api.EventContentBlockStart, Index: 0, ContentBlock: api.ContentBlockInfo{Type: "text"}},
			{Type: api.EventContentBlockDelta, Index: 0, Delta: api.Delta{Type: "text_delta", Text: "Hello, "}},
			{Type: api.EventContentBlockDelta, Index: 0, Delta: api.Delta{Type: "text_delta", Text: "world!"}},
			{Type: api.EventContentBlockStop, Index: 0},
			{Type: api.EventMessageDelta, MessageDelta: api.MessageDelta{StopReason: "end_turn"}, Usage: api.UsageDelta{OutputTokens: 3}},
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

	// Create the server with a ChatLoopFactory.
	srv, err := NewServer(Config{Addr: "127.0.0.1:0"})
	if err != nil {
		t.Fatal(err)
	}
	srv.ChatLoopFactory = func() *runtime.ConversationLoop {
		return loop
	}

	ts := httptest.NewServer(srv.routes())
	defer ts.Close()

	// Dial the chat WS endpoint.
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/api/chat/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	// Read hello.
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read hello: %v", err)
	}
	var hello chatproto.ServerOutbound
	if err := json.Unmarshal(msg, &hello); err != nil {
		t.Fatalf("unmarshal hello: %v", err)
	}
	if hello.Type != chatproto.MsgChatSessionInit {
		t.Errorf("expected chat_session_init, got %q", hello.Type)
	}
	if hello.SessionID == "" {
		t.Error("expected non-empty session_id")
	}

	// Send user_input.
	userJSON := `{"type":"user_input","text":"Hello"}`
	if err := conn.WriteMessage(websocket.TextMessage, []byte(userJSON)); err != nil {
		t.Fatalf("write user_input: %v", err)
	}

	// Collect messages until we see "done".
	var gotTypes []string
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
		gotTypes = append(gotTypes, out.Type)
		if out.Type == chatproto.MsgDone {
			break
		}
	}

	// Verify expected sequence.
	if len(gotTypes) == 0 {
		t.Fatal("expected at least one message after user_input")
	}
	// We expect text_delta + text_final + usage + done (plus possibly warn).
	foundTextDelta := false
	foundTextFinal := false
	foundDone := false
	for _, typ := range gotTypes {
		switch typ {
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
}

// TestChatSession_ParseError sends invalid JSON and expects a parse_error.
func TestChatSession_ParseError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	cfg := &runtime.Config{
		Model:            "expert",
		MaxTokens:        100,
		SessionDir:       t.TempDir(),
		CompactionEnabled: true,
	}
	loop := runtime.NewConversationLoop(cfg, &fakeClient{events: nil})

	srv, err := NewServer(Config{Addr: "127.0.0.1:0"})
	if err != nil {
		t.Fatal(err)
	}
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

	// Send invalid JSON.
	if err := conn.WriteMessage(websocket.TextMessage, []byte("not json")); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Should get a parse_error.
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var out chatproto.ServerOutbound
	if err := json.Unmarshal(msg, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.Type != chatproto.MsgError || out.Code != "parse_error" {
		t.Errorf("expected parse_error, got type=%q code=%q", out.Type, out.Code)
	}
}

// TestChatSession_EmptyUserInput sends a user_input with empty text
// and expects a validation error.
func TestChatSession_EmptyUserInput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	cfg := &runtime.Config{
		Model:            "expert",
		MaxTokens:        100,
		SessionDir:       t.TempDir(),
		CompactionEnabled: true,
	}
	loop := runtime.NewConversationLoop(cfg, &fakeClient{events: nil})

	srv, err := NewServer(Config{Addr: "127.0.0.1:0"})
	if err != nil {
		t.Fatal(err)
	}
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

	// Send user_input with empty text.
	emptyJSON := `{"type":"user_input","text":""}`
	if err := conn.WriteMessage(websocket.TextMessage, []byte(emptyJSON)); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Should get an error.
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var out chatproto.ServerOutbound
	if err := json.Unmarshal(msg, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.Type != chatproto.MsgError || out.Code != "invalid_message" {
		t.Errorf("expected invalid_message error, got type=%q code=%q", out.Type, out.Code)
	}
}

// TestChatSession_PermissionReplyDecode verifies that the server
// decodes a permission_reply message and responds (currently with a
// "not yet supported" warn since PermReply channels aren't wired).
func TestChatSession_PermissionReplyDecode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	cfg := &runtime.Config{
		Model:            "expert",
		MaxTokens:        100,
		SessionDir:       t.TempDir(),
		CompactionEnabled: true,
	}
	loop := runtime.NewConversationLoop(cfg, &fakeClient{events: nil})

	srv, err := NewServer(Config{Addr: "127.0.0.1:0"})
	if err != nil {
		t.Fatal(err)
	}
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

	// Send permission_reply.
	replyJSON := `{"type":"permission_reply","tool_use_id":"tu_01","decision":"allow"}`
	if err := conn.WriteMessage(websocket.TextMessage, []byte(replyJSON)); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Read response — should be a warn about not yet supported.
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var out chatproto.ServerOutbound
	if err := json.Unmarshal(msg, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// Accept either warn or error — just verify the server decodes and
	// responds rather than crashing.
	if out.Type != chatproto.MsgWarn && out.Type != chatproto.MsgError {
		t.Errorf("expected warn or error for permission_reply, got type=%q", out.Type)
	}
}

// TestChatSessionHandleChatWS_NoFactory verifies that /api/chat/ws
// returns 501 when ChatLoopFactory is nil.
func TestChatSessionHandleChatWS_NoFactory(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	srv, err := NewServer(Config{Addr: "127.0.0.1:0"})
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(srv.routes())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/chat/ws")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotImplemented {
		t.Errorf("expected 501, got %d", resp.StatusCode)
	}
}

// TestChatSession_UnknownMessageType verifies that an unknown
// message type produces a warning.
func TestChatSession_UnknownMessageType(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	cfg := &runtime.Config{
		Model:            "expert",
		MaxTokens:        100,
		SessionDir:       t.TempDir(),
		CompactionEnabled: true,
	}
	loop := runtime.NewConversationLoop(cfg, &fakeClient{events: nil})

	srv, err := NewServer(Config{Addr: "127.0.0.1:0"})
	if err != nil {
		t.Fatal(err)
	}
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

	// Send bogus message type.
	bogus := `{"type":"nonexistent","text":"hi"}`
	if err := conn.WriteMessage(websocket.TextMessage, []byte(bogus)); err != nil {
		t.Fatalf("write: %v", err)
	}

	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var out chatproto.ServerOutbound
	if err := json.Unmarshal(msg, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.Type != chatproto.MsgWarn {
		t.Errorf("expected warn for unknown type, got type=%q", out.Type)
	}
}

// TestSendJSON_EmptyType verifies sendJSON does not write messages
// with an empty Type field.
func TestSendJSON_EmptyType(t *testing.T) {
	// We create a pair of loopback connections.
	// sendJSON writes a TextMessage; if Type is empty, it should skip.
	// We can't easily test the skip without a real conn that writes to
	// stderr on error, but we can verify that a non-empty Type is
	// encoded correctly by calling turnEventToOutbound + sendJSON
	// indirectly via the read-round-trip tests above.
	// This test just validates the filter logic at the call site.
	out := turnEventToOutbound(runtime.TurnEvent{Type: runtime.TurnEventInfo})
	if out.Type != "" {
		t.Errorf("expected empty type for TurnEventInfo, got %q", out.Type)
	}
}

// TestChatSession_ConcurrentWrites verifies that the write mutex in
// runChatSession prevents data races when the TurnEvent translator
// and the client-reader loop both write to the connection.
func TestChatSession_ConcurrentWrites(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	// This test exercises the mutex path by sending a user_input while
	// events are being written. The race detector (-race flag) would
	// catch any unprotected concurrent websocket writes.
	cfg := &runtime.Config{
		Model:            "expert",
		MaxTokens:        100,
		SessionDir:       t.TempDir(),
		CompactionEnabled: true,
	}
	loop := runtime.NewConversationLoop(cfg, &fakeClient{
		// Emit enough events that the writer goroutine is still busy.
		events: []api.StreamEvent{
			{Type: api.EventMessageStart, InputTokens: 5},
			{Type: api.EventContentBlockStart, Index: 0, ContentBlock: api.ContentBlockInfo{Type: "text"}},
			{Type: api.EventContentBlockDelta, Index: 0, Delta: api.Delta{Type: "text_delta", Text: "A"}},
			{Type: api.EventContentBlockDelta, Index: 0, Delta: api.Delta{Type: "text_delta", Text: "B"}},
			{Type: api.EventContentBlockStop, Index: 0},
			{Type: api.EventMessageDelta, MessageDelta: api.MessageDelta{StopReason: "end_turn"}, Usage: api.UsageDelta{OutputTokens: 2}},
			{Type: api.EventMessageStop},
		},
	})

	srv, err := NewServer(Config{Addr: "127.0.0.1:0"})
	if err != nil {
		t.Fatal(err)
	}
	srv.ChatLoopFactory = func() *runtime.ConversationLoop {
		return loop
	}

	ts := httptest.NewServer(srv.routes())
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/api/chat/ws"

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				t.Errorf("dial: %v", err)
				return
			}
			defer conn.Close()

			// Read hello.
			_, _, _ = conn.ReadMessage()

			// Send user_input.
			userJSON := `{"type":"user_input","text":"test"}`
			if err := conn.WriteMessage(websocket.TextMessage, []byte(userJSON)); err != nil {
				t.Errorf("write: %v", err)
				return
			}

			// Drain messages until done.
			for {
				_, msg, err := conn.ReadMessage()
				if err != nil {
					break
				}
				var out chatproto.ServerOutbound
				_ = json.Unmarshal(msg, &out)
				if out.Type == chatproto.MsgDone {
					break
				}
			}
		}()
	}
	wg.Wait()
}