// Package chatproto defines the JSON wire format for the chat WebSocket API
// (`/api/chat/ws`). It provides structured message types for both
// client-to-server and server-to-client communication, mirroring the
// internal TurnEvent stream but in a form suitable for a custom HTML/JS
// chat UI.
//
// # Example messages
//
// | Direction | Type                    | Example payload
// |-----------|-------------------------|-----------------
// | Client→S  | user_input             | {"type":"user_input","text":"Hello"}
// | Client→S  | permission_reply       | {"type":"permission_reply","tool_use_id":"tu_01","decision":"allow"}
// | Client→S  | resize                 | {"type":"resize","cols":120,"rows":40}
// | S→Client  | chat_session_init      | {"type":"chat_session_init","session_id":"sess_abc123"}
// | S→Client  | text_delta             | {"type":"text_delta","text":"Hello "}
// | S→Client  | text_final             | {"type":"text_final","text":"Hello, world!"}
// | S→Client  | tool_start             | {"type":"tool_start","tool_use_id":"tu_01","tool_name":"bash","tool_input":"ls -la"}
// | S→Client  | tool_done              | {"type":"tool_done","tool_use_id":"tu_01","result":"file1 file2"}
// | S→Client  | permission_ask         | {"type":"permission_ask","tool_use_id":"tu_01","tool_name":"bash","tool_input":"rm -rf /"}
// | S→Client  | ask_user               | {"type":"ask_user","prompt":"Which file?"}
// | S→Client  | usage                  | {"type":"usage","input_tokens":100,"output_tokens":200}
// | S→Client  | done                   | {"type":"done","reason":"end_turn"}
// | S→Client  | error                  | {"type":"error","code":"rate_limit","message":"Rate limited"}
package chatproto

// ClientInbound is any message the browser sends to the server over the WS.
// The JSON "type" field discriminates.
type ClientInbound struct {
	Type string `json:"type"`

	// user_input
	Text string `json:"text,omitempty"`

	// permission_reply
	ToolUseID string `json:"tool_use_id,omitempty"`
	Decision  string `json:"decision,omitempty"` // "allow", "deny", "allow_always"

	// resize
	Cols int `json:"cols,omitempty"`
	Rows int `json:"rows,omitempty"`
}

// ServerOutbound is any message the server sends to the browser over the WS.
type ServerOutbound struct {
	Type string `json:"type"`

	// chat_session_init
	SessionID string `json:"session_id,omitempty"`

	// text_delta, text_final
	Text string `json:"text,omitempty"`

	// tool_start
	ToolName  string `json:"tool_name,omitempty"`
	ToolInput string `json:"tool_input,omitempty"`

	// tool_done
	Result string `json:"result,omitempty"`

	// ask_user
	Prompt string `json:"prompt,omitempty"`

	// usage
	InputTokens  int `json:"input_tokens,omitempty"`
	OutputTokens int `json:"output_tokens,omitempty"`

	// done
	Reason string `json:"reason,omitempty"` // "end_turn", "stop", "max_turns"

	// error
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// Well-known message types for client→server messages.
const (
	MsgUserInput       = "user_input"
	MsgPermissionReply = "permission_reply"
	MsgResize          = "resize"
)

// Well-known message types for server→client messages.
const (
	MsgChatSessionInit = "chat_session_init"
	MsgTextDelta       = "text_delta"
	MsgTextFinal       = "text_final"
	MsgToolStart       = "tool_start"
	MsgToolDone        = "tool_done"
	MsgPermissionAsk   = "permission_ask"
	MsgAskUser         = "ask_user"
	MsgUsage           = "usage"
	MsgDone            = "done"
	MsgError           = "error"
)