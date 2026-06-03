package web

import (
	"claw-code-go/internal/runtime"
	"claw-code-go/internal/web/chatproto"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/gorilla/websocket"
)

// runChatSession bridges a *runtime.ConversationLoop to a WebSocket
// connection using the chatproto JSON wire format. It sends a
// chat_session_init hello, then starts a goroutine that runs the
// conversation loop (SendMessageStreaming) and translates TurnEvent
// values into JSON messages written to the websocket. Meanwhile the
// main goroutine reads client JSON messages and routes them into the
// conversation (user_input, permission_reply).
//
// The store parameter (may be nil) is used to record session metadata
// (user messages and assistant replies) for the sidebar history.
//
// The function returns when either the websocket closes or ctx is
// cancelled. The caller should close the websocket after this returns.
func runChatSession(ctx context.Context, conn *websocket.Conn, loop *runtime.ConversationLoop, sessionID string, store *chatSessionStore) {
	logger := LoggerFromCtx(ctx).With(slog.String("session_id", sessionID))
	logger.Info("chat_session_started")

	// Send the hello.
	sendJSON(conn, logger, chatproto.ServerOutbound{
		Type:      chatproto.MsgChatSessionInit,
		SessionID: sessionID,
	})

	// TurnEvents from SendMessageStreaming will be written here.
	turnEvents := make(chan runtime.TurnEvent, 32)

	// Mutex to protect websocket writes from two goroutines:
	//   - the TurnEvent translator goroutine
	//   - the direct writes in the client-reader loop (e.g. for
	//     permission_reply echo or error responses)
	var writeMu sync.Mutex

	// Conversation runner goroutine. Calls SendMessageStreaming
	// every time a user_input arrives. The channel stays open until
	// the outer function returns.
	msgCh := make(chan string, 1) // buffered so send doesn't block
	go func() {
		defer close(turnEvents)
		for {
			select {
			case <-ctx.Done():
				return
			case text, ok := <-msgCh:
				if !ok {
					return
				}
				// Run the conversation turn, streaming TurnEvents
				// back. This blocks until the full agentic loop
				// (possibly multiple tool-use turns) completes.
				if err := loop.SendMessageStreaming(ctx, text, turnEvents); err != nil {
					// Non-fatal: the TurnEventError was already
					// pushed into turnEvents by SendMessageStreaming.
					// Just log and continue so the client can send
					// another message.
					logger.Warn("turn_error", "error", err)
				}
			}
		}
	}()

	// TurnEvent → JSON writer goroutine.
	go func() {
		for ev := range turnEvents {
			out := turnEventToOutbound(ev)
			writeMu.Lock()
			sendJSON(conn, logger, out)
			writeMu.Unlock()
			// Record assistant replies for sidebar history.
			if store != nil && ev.Type == runtime.TurnEventTextFinal && ev.Text != "" {
				store.recordAssistantReply(sessionID, ev.Text)
			}
		}
	}()

	// Read loop: parse ClientInbound JSON and route to the loop.
	conn.SetReadLimit(256 * 1024) // 256 KiB per client message
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			break
		}
		var msg chatproto.ClientInbound
		if err := json.Unmarshal(data, &msg); err != nil {
			logger.Warn("parse_error", "error", err)
			writeMu.Lock()
			sendJSON(conn, logger, chatproto.ServerOutbound{
				Type:    chatproto.MsgError,
				Code:    "parse_error",
				Message: fmt.Sprintf("invalid JSON: %v", err),
			})
			writeMu.Unlock()
			continue
		}

		switch msg.Type {
		case chatproto.MsgUserInput:
			if msg.Text == "" {
				logger.Warn("protocol_violation", "reason", "empty user_input")
				writeMu.Lock()
				sendJSON(conn, logger, chatproto.ServerOutbound{
					Type:    chatproto.MsgError,
					Code:    "invalid_message",
					Message: "user_input requires a non-empty text field",
				})
				writeMu.Unlock()
				continue
			}
			// Ack the message so the client knows it was received.
			// This allows the client to avoid resending on reconnect.
			if msg.MessageID != "" {
				writeMu.Lock()
				sendJSON(conn, logger, chatproto.ServerOutbound{
					Type:      chatproto.MsgAck,
					MessageID: msg.MessageID,
				})
				writeMu.Unlock()
			}
			// Record the user message in the session store for sidebar history.
			if store != nil {
				store.recordUserMessage(sessionID, msg.Text)
			}
			select {
			case msgCh <- msg.Text:
			case <-ctx.Done():
				return
			}

		case chatproto.MsgPermissionReply:
			logger.Warn("protocol_violation", "reason", "permission_reply not yet supported")
			writeMu.Lock()
			sendJSON(conn, logger, chatproto.ServerOutbound{
				Type:    chatproto.MsgWarn,
				Message: "permission_reply not yet supported via chat API; use /terminal for permission prompts",
			})
			writeMu.Unlock()

		default:
			logger.Warn("protocol_violation", "reason", "unknown message type", "msg_type", msg.Type)
			writeMu.Lock()
			sendJSON(conn, logger, chatproto.ServerOutbound{
				Type:    chatproto.MsgWarn,
				Message: fmt.Sprintf("unknown message type: %s", msg.Type),
			})
			writeMu.Unlock()
		}
	}
	logger.Info("chat_session_ended")
}

// turnEventToOutbound converts a runtime.TurnEvent to a chatproto.ServerOutbound.
func turnEventToOutbound(ev runtime.TurnEvent) chatproto.ServerOutbound {
	switch ev.Type {
	case runtime.TurnEventTextDelta:
		return chatproto.ServerOutbound{
			Type: chatproto.MsgTextDelta,
			Text: ev.Text,
		}
	case runtime.TurnEventTextFinal:
		return chatproto.ServerOutbound{
			Type: chatproto.MsgTextFinal,
			Text: ev.Text,
		}
	case runtime.TurnEventToolStart:
		return chatproto.ServerOutbound{
			Type:      chatproto.MsgToolStart,
			ToolName:  ev.ToolName,
			ToolInput: ev.ToolInput,
		}
	case runtime.TurnEventToolDone:
		return chatproto.ServerOutbound{
			Type:   chatproto.MsgToolDone,
			Result: ev.ToolResult,
		}
	case runtime.TurnEventPermissionAsk:
		return chatproto.ServerOutbound{
			Type:      chatproto.MsgPermissionAsk,
			ToolName:  ev.ToolName,
			ToolInput: ev.ToolInput,
		}
	case runtime.TurnEventAskUser:
		return chatproto.ServerOutbound{
			Type:   chatproto.MsgAskUser,
			Prompt: ev.ToolInput, // AskUserInput places the question here
		}
	case runtime.TurnEventUsage:
		return chatproto.ServerOutbound{
			Type:         chatproto.MsgUsage,
			InputTokens:  ev.InputTokens,
			OutputTokens: ev.OutputTokens,
		}
	case runtime.TurnEventDone:
		return chatproto.ServerOutbound{
			Type: chatproto.MsgDone,
		}
	case runtime.TurnEventWarn:
		return chatproto.ServerOutbound{
			Type:    chatproto.MsgWarn,
			Message: ev.Text,
		}
	case runtime.TurnEventError:
		msg := "unknown error"
		if ev.Err != nil {
			msg = ev.Err.Error()
		}
		return chatproto.ServerOutbound{
			Type:    chatproto.MsgError,
			Code:    "turn_error",
			Message: msg,
		}
	default:
		// TurnEventInfo and anything else is dropped silently
		return chatproto.ServerOutbound{}
	}
}

// sendJSON marshals msg and writes it as a TextMessage to the
// websocket. Errors are logged using the provided logger.
func sendJSON(conn *websocket.Conn, logger *slog.Logger, msg chatproto.ServerOutbound) {
	if msg.Type == "" {
		return // skip empty default messages
	}
	data, err := json.Marshal(msg)
	if err != nil {
		logger.Warn("marshal_error", "error", err)
		return
	}
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		logger.Warn("websocket_write_error", "error", err)
	}
}