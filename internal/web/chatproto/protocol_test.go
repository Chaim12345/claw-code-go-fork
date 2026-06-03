package chatproto

import (
	"encoding/json"
	"testing"
)

func TestClientInbound_JSONRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		msg  ClientInbound
		json string
	}{
		{
			name: "user_input",
			msg:  ClientInbound{Type: "user_input", Text: "Hello, world!"},
			json: `{"type":"user_input","text":"Hello, world!"}`,
		},
		{
			name: "user_input with id",
			msg:  ClientInbound{Type: "user_input", Text: "Hi", MessageID: "msg_001"},
			json: `{"type":"user_input","text":"Hi","message_id":"msg_001"}`,
		},
		{
			name: "permission_reply allow",
			msg:  ClientInbound{Type: "permission_reply", ToolUseID: "tu_01", Decision: "allow"},
			json: `{"type":"permission_reply","tool_use_id":"tu_01","decision":"allow"}`,
		},
		{
			name: "permission_reply deny",
			msg:  ClientInbound{Type: "permission_reply", ToolUseID: "tu_02", Decision: "deny"},
			json: `{"type":"permission_reply","tool_use_id":"tu_02","decision":"deny"}`,
		},
		{
			name: "permission_reply allow_always",
			msg:  ClientInbound{Type: "permission_reply", ToolUseID: "tu_03", Decision: "allow_always"},
			json: `{"type":"permission_reply","tool_use_id":"tu_03","decision":"allow_always"}`,
		},
		{
			name: "resize",
			msg:  ClientInbound{Type: "resize", Cols: 120, Rows: 40},
			json: `{"type":"resize","cols":120,"rows":40}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+"/marshal", func(t *testing.T) {
			data, err := json.Marshal(tt.msg)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if string(data) != tt.json {
				t.Errorf("marshal: got %s, want %s", string(data), tt.json)
			}
		})
		t.Run(tt.name+"/unmarshal", func(t *testing.T) {
			var msg ClientInbound
			if err := json.Unmarshal([]byte(tt.json), &msg); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if msg != tt.msg {
				t.Errorf("unmarshal: got %+v, want %+v", msg, tt.msg)
			}
		})
	}
}

func TestServerOutbound_JSONRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		msg  ServerOutbound
		json string
	}{
		{
			name: "chat_session_init",
			msg:  ServerOutbound{Type: "chat_session_init", SessionID: "sess_abc123"},
			json: `{"type":"chat_session_init","session_id":"sess_abc123"}`,
		},
		{
			name: "text_delta",
			msg:  ServerOutbound{Type: "text_delta", Text: "Hello "},
			json: `{"type":"text_delta","text":"Hello "}`,
		},
		{
			name: "text_final",
			msg:  ServerOutbound{Type: "text_final", Text: "Hello, world!"},
			json: `{"type":"text_final","text":"Hello, world!"}`,
		},
		{
			name: "tool_start",
			msg:  ServerOutbound{Type: "tool_start", ToolUseID: "tu_01", ToolName: "bash", ToolInput: "ls -la"},
			json: `{"type":"tool_start","tool_use_id":"tu_01","tool_name":"bash","tool_input":"ls -la"}`,
		},
		{
			name: "tool_done",
			msg:  ServerOutbound{Type: "tool_done", ToolUseID: "tu_01", Result: "file1\nfile2"},
			json: `{"type":"tool_done","tool_use_id":"tu_01","result":"file1\nfile2"}`,
		},
		{
			name: "permission_ask",
			msg:  ServerOutbound{Type: "permission_ask", ToolUseID: "tu_02", ToolName: "bash", ToolInput: "rm -rf /"},
			json: `{"type":"permission_ask","tool_use_id":"tu_02","tool_name":"bash","tool_input":"rm -rf /"}`,
		},
		{
			name: "ask_user",
			msg:  ServerOutbound{Type: "ask_user", Prompt: "Which file should I edit?"},
			json: `{"type":"ask_user","prompt":"Which file should I edit?"}`,
		},
		{
			name: "usage",
			msg:  ServerOutbound{Type: "usage", InputTokens: 150, OutputTokens: 80},
			json: `{"type":"usage","input_tokens":150,"output_tokens":80}`,
		},
		{
			name: "done",
			msg:  ServerOutbound{Type: "done", Reason: "end_turn"},
			json: `{"type":"done","reason":"end_turn"}`,
		},
		{
			name: "error",
			msg:  ServerOutbound{Type: "error", Code: "rate_limit", Message: "Rate limit exceeded"},
			json: `{"type":"error","code":"rate_limit","message":"Rate limit exceeded"}`,
		},
		{
			name: "ack",
			msg:  ServerOutbound{Type: "ack", MessageID: "msg_xyz"},
			json: `{"type":"ack","message_id":"msg_xyz"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+"/marshal", func(t *testing.T) {
			data, err := json.Marshal(tt.msg)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if string(data) != tt.json {
				t.Errorf("marshal: got %s, want %s", string(data), tt.json)
			}
		})
		t.Run(tt.name+"/unmarshal", func(t *testing.T) {
			var msg ServerOutbound
			if err := json.Unmarshal([]byte(tt.json), &msg); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if msg != tt.msg {
				t.Errorf("unmarshal: got %+v, want %+v", msg, tt.msg)
			}
		})
	}
}

func TestClientInbound_OmitEmptyFields(t *testing.T) {
	// Only the type field and non-zero fields should be present.
	// Verify by checking the marshalled JSON string does not contain
	// the empty-value field keys.
	msg := ClientInbound{Type: "user_input", Text: "hi"}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	str := string(data)
	// Zero-value fields should not appear in the JSON.
	for _, field := range []string{"message_id", "tool_use_id", "decision", "cols", "rows"} {
		if containsStr(str, `"`+field+`"`) {
			t.Errorf("zero-value field %q should be omitted from JSON: %s", field, str)
		}
	}
	if !containsStr(str, `"type":"user_input"`) {
		t.Errorf("expected type field, got: %s", str)
	}
	if !containsStr(str, `"text":"hi"`) {
		t.Errorf("expected text field, got: %s", str)
	}
}

func TestServerOutbound_OmitEmptyFields(t *testing.T) {
	msg := ServerOutbound{Type: "text_final", Text: "done"}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	str := string(data)
	omittedFields := []string{"session_id", "message_id", "tool_use_id", "tool_name",
		"tool_input", "result", "prompt", "input_tokens", "output_tokens", "reason", "code", "message"}
	for _, field := range omittedFields {
		if containsStr(str, `"`+field+`"`) {
			t.Errorf("zero-value field %q should be omitted from JSON: %s", field, str)
		}
	}
	if !containsStr(str, `"type":"text_final"`) {
		t.Errorf("expected type field, got: %s", str)
	}
	if !containsStr(str, `"text":"done"`) {
		t.Errorf("expected text field, got: %s", str)
	}
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestWellKnownConstants(t *testing.T) {
	// Verify the well-known constants are all unique and non-empty.
	clientTypes := map[string]bool{}
	for _, c := range []string{MsgUserInput, MsgPermissionReply, MsgResize} {
		if c == "" {
			t.Error("client message type constant is empty")
		}
		if clientTypes[c] {
			t.Errorf("duplicate client message type: %s", c)
		}
		clientTypes[c] = true
	}
	if len(clientTypes) != 3 {
		t.Errorf("expected 3 client message types, got %d", len(clientTypes))
	}

	serverTypes := map[string]bool{}
	for _, c := range []string{MsgChatSessionInit, MsgTextDelta, MsgTextFinal, MsgToolStart, MsgToolDone, MsgPermissionAsk, MsgAskUser, MsgUsage, MsgDone, MsgError, MsgWarn, MsgAck} {
		if c == "" {
			t.Error("server message type constant is empty")
		}
		if serverTypes[c] {
			t.Errorf("duplicate server message type: %s", c)
		}
		serverTypes[c] = true
	}
	if len(serverTypes) != 12 {
		t.Errorf("expected 12 server message types, got %d", len(serverTypes))
	}
}