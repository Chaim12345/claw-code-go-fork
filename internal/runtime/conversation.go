package runtime

import (
	"claw-code-go/internal/api"
	clawctx "claw-code-go/internal/context"
	"claw-code-go/internal/mcp"
	"claw-code-go/internal/permissions"
	"claw-code-go/internal/tools"
	"claw-code-go/internal/usage"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const systemPromptBase = `You are Claude Code, an AI assistant for software engineering tasks. You have access to tools for running bash commands, reading and writing files, searching with glob patterns, and grepping for patterns in code. Use these tools to help users with coding tasks.

IMPORTANT: When using tools, prefer XML format over JSON for better streaming compatibility:
<tool_name>
<parameter_name>value</parameter_name>
</tool_name>

JSON format is supported as a fallback but XML is preferred.`

// ConversationLoop manages the agentic conversation loop with tool use.
type ConversationLoop struct {
	Client          api.APIClient // provider-agnostic client interface
	Session         *Session
	Tools           []api.Tool
	Permissions     *Permissions
	PermManager     *permissions.Manager  // Phase 5 permission manager (may be nil)
	Config          *Config
	MCPRegistry     *mcp.Registry         // MCP server registry (may be nil)
	Compaction      CompactionState        // Phase 6 token tracking and compaction state
	CtxAssembler    *clawctx.Assembler    // Phase 12 context assembler (may be nil)
	Usage           *usage.Tracker        // Phase 13 per-session token usage tracker

	// lastStopReason is the stop_reason reported by the most recent
	// streaming turn ("end_turn", "tool_use", "max_tokens", …). Set
	// by runOneTurnStreaming and read by RunTask's end_turn fast
	// path. Zero-value when no turn has run yet.
	lastStopReason string
}

// NewConversationLoop creates a new conversation loop with the given client.
// Use NewProviderClient to create an appropriate client for the configured provider.
func NewConversationLoop(cfg *Config, client api.APIClient) *ConversationLoop {
	workDir, _ := os.Getwd()
	return &ConversationLoop{
		Client:  client,
		Session: NewSession(),
		Tools: []api.Tool{
			tools.BashTool(),
			tools.ReadFileTool(),
			tools.WriteFileTool(),
			tools.GlobTool(),
			tools.GrepTool(),
			tools.FileEditTool(),
			tools.WebFetchTool(),
			tools.WebSearchTool(),
			tools.AskUserQuestionTool(),
			tools.TodoWriteTool(),
		},
		Permissions:  DefaultPermissions(),
		Config:       cfg,
		CtxAssembler: clawctx.NewAssembler(workDir),
		Usage:        usage.NewTracker(cfg.Model),
	}
}

// SystemPrompt returns the rendered system prompt for diagnostic use.
func (loop *ConversationLoop) SystemPrompt() string {
	return loop.systemPrompt()
}

// systemPrompt returns the system prompt, optionally injecting project context,
// compaction summary, and MCP tool context.
func (loop *ConversationLoop) systemPrompt() string {
	var parts []string
	parts = append(parts, systemPromptBase)

	// Inject project context (Phase 12): environment, git status, CLAUDE.md.
	if loop.CtxAssembler != nil {
		if ctx := loop.CtxAssembler.Assemble(); ctx != "" {
			parts = append(parts, ctx)
		}
	}

	// Inject compaction summary when the session has one (Phase 6).
	if loop.Session != nil && loop.Session.CompactionSummary != "" {
		parts = append(parts, FormatCompactSummary(loop.Session.CompactionSummary))
	}

	// Append MCP tool list if any servers are connected.
	if loop.MCPRegistry != nil {
		mcpTools := loop.MCPRegistry.AllTools()
		if len(mcpTools) > 0 {
			names := make([]string, len(mcpTools))
			for i, t := range mcpTools {
				names[i] = t.Name
			}
			parts = append(parts, "Additional tools available via MCP: "+strings.Join(names, ", ")+".")
		}
	}

	return strings.Join(parts, "\n\n")
}

// allTools returns built-in tools merged with any MCP tools.
func (loop *ConversationLoop) allTools() []api.Tool {
	if loop.MCPRegistry == nil {
		return loop.Tools
	}
	mcpAPITools := loop.MCPRegistry.AllAPITools()
	if len(mcpAPITools) == 0 {
		return loop.Tools
	}
	combined := make([]api.Tool, 0, len(loop.Tools)+len(mcpAPITools))
	combined = append(combined, loop.Tools...)
	combined = append(combined, mcpAPITools...)
	return combined
}

// SendMessage sends a user message and runs the full agentic loop.
func (loop *ConversationLoop) SendMessage(ctx context.Context, userText string) error {
	// Create and validate user message
	userMsg := api.Message{
		Role: "user",
		Content: []api.ContentBlock{
			{Type: "text", Text: userText},
		},
	}
	
	// Truncate if message exceeds size limits
	userMsg, wasTruncated := ValidateAndTruncateMessage(userMsg)
	if wasTruncated {
		fmt.Fprintf(os.Stderr, "[message-limit] user message truncated to fit provider limits\n")
	}
	
	// Append validated message
	loop.Session.Messages = append(loop.Session.Messages, userMsg)

	// Compact history if approaching the token budget (Phase 6).
	if ShouldCompact(loop.Compaction.LastInputTokens, loop.Session.Messages, loop.Config) {
		summary, err := CompactSession(ctx, loop.Client, loop.Config, loop.Session)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[compact] warning: %v\n", err)
		} else {
			loop.Compaction.CompactionCount++
			// Prepend a continuation marker to the retained recent messages.
			contMsg := GetContinuationMessage(summary)
			loop.Session.Messages = append([]api.Message{contMsg}, loop.Session.Messages...)
		}
	}

	var totalInput, totalOutput int

	// Agentic loop: keep going until stop_reason is "end_turn".
	// If the provider rejects a turn because the prompt is too large,
	// we compact once and retry before surfacing the error. The retry
	// budget prevents an infinite loop when compaction cannot shrink
	// the history below the model cap.
	const maxPromptRecoveryRetries = 1
	promptRecoveryRetries := 0
retryTurn:
	for {
		stopReason, inTok, outTok, err := loop.runOneTurn(ctx)
		if err != nil {
			if promptRecoveryRetries < maxPromptRecoveryRetries && isPromptTooLargeError(err) {
				promptRecoveryRetries++
				fmt.Fprintf(os.Stderr, "[auto-compact] prompt exceeded model cap; compacting and retrying (%d/%d)\n",
					promptRecoveryRetries, maxPromptRecoveryRetries)
				if _, cerr := loop.CompactNow(ctx); cerr != nil {
					return fmt.Errorf("prompt too large and auto-compact failed: %w (original: %v)", cerr, err)
				}
				continue retryTurn
			}
			return err
		}
		totalInput += inTok
		totalOutput += outTok

		if stopReason != "tool_use" {
			break
		}
	}

	// Update compaction state with the latest token counts (Phase 6).
	// Mirrors the bookkeeping in SendMessageStreaming (line ~344) so
	// the CLI and live-test paths get the same accurate usage data
	// the TUI does. Without this, LastInputTokens stays 0 forever
	// and ShouldCompact falls back to the chars/4 estimator on
	// every turn.
	loop.Compaction.LastInputTokens = totalInput
	loop.Compaction.TotalInputTokens += totalInput
	loop.Compaction.TotalOutputTokens += totalOutput

	// Update the usage tracker (Phase 13).
	if loop.Usage != nil {
		loop.Usage.Add(totalInput, totalOutput, 0, 0)
	}

	return nil
}

// runOneTurn sends the current session messages to the API and processes the response.
// Returns the stop_reason, the input-token count reported on the
// stream's EventMessageStart, and the output-token count reported on
// the final EventMessageDelta. The caller (SendMessage) is responsible
// for aggregating these across the agentic loop and updating
// loop.Compaction / loop.Usage — see SendMessageStreaming for the
// equivalent wiring on the streaming path.
func (loop *ConversationLoop) runOneTurn(ctx context.Context) (string, int, int, error) {
	req := api.CreateMessageRequest{
		Model:     loop.Config.Model,
		MaxTokens: loop.Config.MaxTokens,
		System:    loop.systemPrompt(),
		Messages:  loop.Session.Messages,
		Tools:     loop.allTools(),
		Stream:    true,
	}

	ch, err := loop.Client.StreamResponse(ctx, req)
	if err != nil {
		return "", 0, 0, fmt.Errorf("stream response: %w", err)
	}

	// Accumulators for the current response
	type toolBlock struct {
		id          string
		name        string
		inputBuffer string
	}

	var (
		textBlocks          []api.ContentBlock
		toolBlocks          []toolBlock
		currentText         string
		currentTool         *toolBlock
		stopReason          string
		blockTypeMap        = make(map[int]string) // index -> "text" or "tool_use"
		inputTokens         int
		outputTokens        int
		streamedOutputChars int // total bytes the provider emitted across text deltas + tool argument deltas; used for the outputTokens fallback
	)


	for event := range ch {
		switch event.Type {
		case api.EventError:
			return "", 0, 0, fmt.Errorf("stream error: %s", event.ErrorMessage)

		case api.EventMessageStart:
			// Provider-estimated input token count (DeepSeek web does
			// not publish a per-request input_tokens field, so this is
			// the client's chars/4 estimate). Captured here so the
			// caller can pass it to ShouldCompact as ground truth.
			inputTokens = event.InputTokens

		case api.EventContentBlockStart:
			blockTypeMap[event.Index] = event.ContentBlock.Type
			if event.ContentBlock.Type == "tool_use" {
				tb := toolBlock{
					id:   event.ContentBlock.ID,
					name: event.ContentBlock.Name,
				}
				toolBlocks = append(toolBlocks, tb)
				currentTool = &toolBlocks[len(toolBlocks)-1]
			}

		case api.EventContentBlockDelta:
			switch event.Delta.Type {
			case "text_delta":
				currentText += event.Delta.Text
				streamedOutputChars += len(event.Delta.Text)
				fmt.Fprint(os.Stdout, event.Delta.Text)

			case "input_json_delta":
				if currentTool != nil {
					currentTool.inputBuffer += event.Delta.PartialJSON
				}
				streamedOutputChars += len(event.Delta.PartialJSON)
			}

		case api.EventContentBlockStop:
			bType, ok := blockTypeMap[event.Index]
			if ok && bType == "text" && currentText != "" {
				textBlocks = append(textBlocks, api.ContentBlock{
					Type: "text",
					Text: currentText,
				})
				currentText = ""
			}
			// Reset currentTool pointer (but keep toolBlocks slice)
			currentTool = nil

		case api.EventMessageDelta:
			stopReason = event.StopReason
			// Server-reported output token count (DeepSeek's
			// accumulated_token_usage from the BATCH event, captured
			// by the provider into EventMessageDelta.Usage). Falls
			// back to 0 if the provider doesn't populate it.
			outputTokens = event.Usage.OutputTokens

		case api.EventMessageStop:
			// Stream complete
		}
	}

	// Ensure trailing newline after streaming text
	if len(textBlocks) > 0 || len(toolBlocks) > 0 {
		fmt.Fprintln(os.Stdout)
	}

	// Build the assistant message content
	var assistantContent []api.ContentBlock

	// Add text blocks first, with tool-call syntax stripped so the
	// raw `{"tool_calls":[...]}` JSON (or `<tool_calls>...</tool_calls>`
	// XML the model sometimes emits inline) doesn't end up in the saved
	// session. The streaming path (runOneTurnStreaming) additionally
	// emits TurnEventTextFinal so the TUI can rewrite its in-flight
	// buffer; this path doesn't have a buffer to rewrite, so cleanup
	// happens only at the persistence boundary.
	for i := range textBlocks {
		textBlocks[i].Text = api.TrimEndOfTurnIndicator(api.StripToolCalls(textBlocks[i].Text))
	}
	assistantContent = append(assistantContent, textBlocks...)

	// Add tool_use blocks
	for _, tb := range toolBlocks {
		var inputMap map[string]any
		if tb.inputBuffer != "" {
			if err := json.Unmarshal([]byte(tb.inputBuffer), &inputMap); err != nil {
				inputMap = map[string]any{"raw": tb.inputBuffer}
			}
		} else {
			inputMap = map[string]any{}
		}

		assistantContent = append(assistantContent, api.ContentBlock{
			Type:  "tool_use",
			ID:    tb.id,
			Name:  tb.name,
			Input: inputMap,
		})
	}

	// Append assistant message to session
	if len(assistantContent) > 0 {
		loop.Session.Messages = append(loop.Session.Messages, api.Message{
			Role:    "assistant",
			Content: assistantContent,
		})
	}

	// If stop_reason is tool_use, execute tools and append results
	if stopReason == "tool_use" {
		var toolResults []api.ContentBlock

		for _, tb := range toolBlocks {
			var inputMap map[string]any
			for _, cb := range assistantContent {
				if cb.Type == "tool_use" && cb.ID == tb.id {
					inputMap = cb.Input
					break
				}
			}
			if inputMap == nil {
				inputMap = map[string]any{}
			}

			fmt.Fprintf(os.Stdout, "\n[Tool: %s]\n", tb.name)
			result := loop.ExecuteTool(tb.name, inputMap)
			result.ToolUseID = tb.id
			toolResults = append(toolResults, result)
		}

		// Append tool results as a user message
		loop.Session.Messages = append(loop.Session.Messages, api.Message{
			Role:    "user",
			Content: toolResults,
		})
	}

	// If the provider didn't surface a server-reported output token
	// count (DeepSeek's BATCH accumulated_token_usage is not always
	// emitted — see live exploration), estimate from the total bytes
	// the provider emitted across both text deltas and tool-argument
	// deltas. The previous fallback only counted text blocks, which
	// produced 0 output tokens for tool-only responses (the model
	// emits raw JSON via input_json_delta with no preceding text) and
	// shortchanged mixed responses (the text was counted, but the
	// tool arguments that consumed the model's reasoning budget were
	// not). streamedOutputChars is the total of every delta this
	// loop observed, which is the closest local proxy we have to
	// the server's true output-token count.
	if outputTokens == 0 && streamedOutputChars > 0 {
		outputTokens = streamedOutputChars / charsPerToken
		if outputTokens == 0 {
			outputTokens = 1
		}
	}

	return stopReason, inputTokens, outputTokens, nil
}

// SendMessageStreaming sends a user message and runs the full agentic loop, emitting
// TurnEvents to the provided channel. The channel is NOT closed by this function;
// callers should close it after this returns.
func (loop *ConversationLoop) SendMessageStreaming(ctx context.Context, userText string, events chan<- TurnEvent) error {
	// Create and validate user message
	userMsg := api.Message{
		Role: "user",
		Content: []api.ContentBlock{
			{Type: "text", Text: userText},
		},
	}
	
	// Truncate if message exceeds size limits
	userMsg, wasTruncated := ValidateAndTruncateMessage(userMsg)
	if wasTruncated {
		events <- TurnEvent{
			Type: TurnEventWarn,
			Text: "Your message was truncated to fit provider size limits. The full content is preserved in the conversation history.",
		}
	}
	
	// Append validated message
	loop.Session.Messages = append(loop.Session.Messages, userMsg)

	// Compact history if approaching the token budget (Phase 6).
	if ShouldCompact(loop.Compaction.LastInputTokens, loop.Session.Messages, loop.Config) {
		summary, err := CompactSession(ctx, loop.Client, loop.Config, loop.Session)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[compact] warning: %v\n", err)
		} else {
			loop.Compaction.CompactionCount++
			// Prepend a continuation marker to the retained recent messages.
			contMsg := GetContinuationMessage(summary)
			loop.Session.Messages = append([]api.Message{contMsg}, loop.Session.Messages...)
		}
	}

	var totalInput, totalOutput int

	// Agentic loop: keep going until stop_reason is "end_turn".
	// If the provider rejects a turn because the prompt is too large,
	// compact once and retry before surfacing the error. Mirrors the
	// same recovery path in SendMessage so both the TUI and
	// autonomous (RunTask) paths benefit from it.
	const maxStreamingPromptRecoveryRetries = 1
	streamingPromptRecoveryRetries := 0
streamingRetryTurn:
	for {
		stopReason, inTok, outTok, err := loop.runOneTurnStreaming(ctx, events)
		if err != nil {
			if streamingPromptRecoveryRetries < maxStreamingPromptRecoveryRetries && isPromptTooLargeError(err) {
				streamingPromptRecoveryRetries++
				fmt.Fprintf(os.Stderr, "[auto-compact] prompt exceeded model cap; compacting and retrying (%d/%d)\n",
					streamingPromptRecoveryRetries, maxStreamingPromptRecoveryRetries)
				if _, cerr := loop.CompactNow(ctx); cerr != nil {
					events <- TurnEvent{Type: TurnEventError, Err: fmt.Errorf("prompt too large and auto-compact failed: %w (original: %v)", cerr, err)}
					return fmt.Errorf("prompt too large and auto-compact failed: %w (original: %v)", cerr, err)
				}
				continue streamingRetryTurn
			}
			events <- TurnEvent{Type: TurnEventError, Err: err}
			return err
		}
		totalInput += inTok
		totalOutput += outTok

		if stopReason != "tool_use" {
			break
		}
	}

	// Update compaction state with the latest token counts (Phase 6).
	loop.Compaction.LastInputTokens = totalInput
	loop.Compaction.TotalInputTokens += totalInput
	loop.Compaction.TotalOutputTokens += totalOutput

	// Update the usage tracker (Phase 13).
	if loop.Usage != nil {
		loop.Usage.Add(totalInput, totalOutput, 0, 0)
	}

	events <- TurnEvent{
		Type:         TurnEventUsage,
		InputTokens:  totalInput,
		OutputTokens: totalOutput,
	}
	events <- TurnEvent{Type: TurnEventDone}
	return nil
}

// runOneTurnStreaming streams one API turn and sends TurnEvents.
// Returns stop_reason, inputTokens, outputTokens, error.
func (loop *ConversationLoop) runOneTurnStreaming(ctx context.Context, events chan<- TurnEvent) (string, int, int, error) {
	req := api.CreateMessageRequest{
		Model:     loop.Config.Model,
		MaxTokens: loop.Config.MaxTokens,
		System:    loop.systemPrompt(),
		Messages:  loop.Session.Messages,
		Tools:     loop.allTools(),
		Stream:    true,
	}

	ch, err := loop.Client.StreamResponse(ctx, req)
	if err != nil {
		return "", 0, 0, fmt.Errorf("stream response: %w", err)
	}

	type toolBlock struct {
		id          string
		name        string
		inputBuffer string
	}

	var (
		textBlocks          []api.ContentBlock
		toolBlocks          []toolBlock
		currentText         string
		currentTool         *toolBlock
		stopReason          string
		blockTypeMap        = make(map[int]string)
		inputTokens         int
		outputTokens        int
		streamedOutputChars int // see runOneTurn for rationale; this path is consumed by the TUI which renders the same usage line
	)

	for event := range ch {
		switch event.Type {
		case api.EventError:
			return "", 0, 0, fmt.Errorf("stream error: %s", event.ErrorMessage)

		case api.EventMessageStart:
			inputTokens = event.InputTokens

		case api.EventContentBlockStart:
			blockTypeMap[event.Index] = event.ContentBlock.Type
			if event.ContentBlock.Type == "tool_use" {
				tb := toolBlock{
					id:   event.ContentBlock.ID,
					name: event.ContentBlock.Name,
				}
				toolBlocks = append(toolBlocks, tb)
				currentTool = &toolBlocks[len(toolBlocks)-1]
			}

		case api.EventContentBlockDelta:
			switch event.Delta.Type {
			case "text_delta":
				currentText += event.Delta.Text
				streamedOutputChars += len(event.Delta.Text)
				select {
				case events <- TurnEvent{Type: TurnEventTextDelta, Text: event.Delta.Text}:
				case <-ctx.Done():
					return "", 0, 0, ctx.Err()
				}
			case "input_json_delta":
				if currentTool != nil {
					currentTool.inputBuffer += event.Delta.PartialJSON
				}
				streamedOutputChars += len(event.Delta.PartialJSON)
			}

		case api.EventContentBlockStop:
			if bType, ok := blockTypeMap[event.Index]; ok && bType == "text" && currentText != "" {
				textBlocks = append(textBlocks, api.ContentBlock{Type: "text", Text: currentText})
				currentText = ""
			}
			currentTool = nil

		case api.EventMessageDelta:
			stopReason = event.StopReason
			outputTokens = event.Usage.OutputTokens
			loop.lastStopReason = stopReason

		case api.EventMessageStop:
			// stream complete
		}
	}

	// Build assistant message content. Strip tool-call syntax from the
	// accumulated text so the raw JSON / XML the model sometimes emits
	// inline (e.g. `{"tool_calls":[{"name":"bash",...}]}` or
	// `<tool_calls>...</tool_calls>`) never lands in the saved session.
	// Also trim the end-of-turn FINISHED sentinel the model sometimes
	// appends as a learned stop marker.
	for i := range textBlocks {
		textBlocks[i].Text = api.TrimEndOfTurnIndicator(api.StripToolCalls(textBlocks[i].Text))
	}
	var assistantContent []api.ContentBlock
	assistantContent = append(assistantContent, textBlocks...)

	// Tell the UI to replace whatever in-flight text it has buffered
	// with the cleaned version. Without this, the TUI commits the
	// raw streamBuf (built from text deltas) into viewBuf on
	// TurnEventDone, and the user sees tool-call syntax forever in
	// their history. The TUI re-renders the assistant line in place.
	if len(textBlocks) > 0 {
		var joined string
		for i, b := range textBlocks {
			if i > 0 {
				joined += "\n"
			}
			joined += b.Text
		}
		select {
		case events <- TurnEvent{Type: TurnEventTextFinal, Text: joined}:
		case <-ctx.Done():
			return "", 0, 0, ctx.Err()
		}
	}

	for _, tb := range toolBlocks {
		var inputMap map[string]any
		if tb.inputBuffer != "" {
			if err := json.Unmarshal([]byte(tb.inputBuffer), &inputMap); err != nil {
				inputMap = map[string]any{"raw": tb.inputBuffer}
			}
		} else {
			inputMap = map[string]any{}
		}
		assistantContent = append(assistantContent, api.ContentBlock{
			Type:  "tool_use",
			ID:    tb.id,
			Name:  tb.name,
			Input: inputMap,
		})
	}

	if len(assistantContent) > 0 {
		loop.Session.Messages = append(loop.Session.Messages, api.Message{
			Role:    "assistant",
			Content: assistantContent,
		})
	}

	// Execute tools if needed
	if stopReason == "tool_use" {
		var toolResults []api.ContentBlock

		for _, tb := range toolBlocks {
			var inputMap map[string]any
			for _, cb := range assistantContent {
				if cb.Type == "tool_use" && cb.ID == tb.id {
					inputMap = cb.Input
					break
				}
			}
			if inputMap == nil {
				inputMap = map[string]any{}
			}

			summary := summarizeToolInput(inputMap)

			// --- Permission check (Phase 5) ---
			if loop.PermManager != nil {
				decision := loop.PermManager.Check(tb.name, summary)

				// Plan mode: describe without executing.
				if loop.PermManager.Mode == permissions.ModePlan {
					planResult := api.ContentBlock{
						Type:    "tool_result",
						ToolUseID: tb.id,
						Content: []api.ContentBlock{{Type: "text", Text: fmt.Sprintf("[Plan: %s %s]", tb.name, summary)}},
					}
					toolResults = append(toolResults, planResult)
					continue
				}

				switch decision {
				case permissions.DecisionDeny:
					denied := api.ContentBlock{
						Type:    "tool_result",
						ToolUseID: tb.id,
						Content: []api.ContentBlock{{Type: "text", Text: fmt.Sprintf("Permission denied for tool: %s", tb.name)}},
						IsError: true,
					}
					toolResults = append(toolResults, denied)
					continue

				case permissions.DecisionAsk:
					replyCh := make(chan PermDecision, 1)
					select {
					case events <- TurnEvent{
						Type:      TurnEventPermissionAsk,
						ToolName:  tb.name,
						ToolInput: summary,
						PermReply: replyCh,
					}:
					case <-ctx.Done():
						return "", 0, 0, ctx.Err()
					}

					var userDecision PermDecision
					select {
					case userDecision = <-replyCh:
					case <-ctx.Done():
						return "", 0, 0, ctx.Err()
					}

					switch userDecision {
					case PermDecisionDeny:
						denied := api.ContentBlock{
							Type:    "tool_result",
							ToolUseID: tb.id,
							Content: []api.ContentBlock{{Type: "text", Text: fmt.Sprintf("Permission denied for tool: %s", tb.name)}},
							IsError: true,
						}
						toolResults = append(toolResults, denied)
						continue
					case PermDecisionAllowAlways:
						loop.PermManager.Remember(tb.name, summary, permissions.DecisionAllow, permissions.ScopeAlways)
					}
					// PermDecisionAllowOnce falls through to execution
				}
				// DecisionAllow falls through to execution
			}

			// ask_user: surface the question to the caller and wait for a reply.
			if tb.name == "ask_user" {
				question, _ := tools.AskUserInput(inputMap)
				if question == "" {
					question = "?"
				}
				replyCh := make(chan string, 1)
				select {
				case events <- TurnEvent{
					Type:         TurnEventAskUser,
					ToolName:     tb.name,
					ToolInput:    question,
					AskUserReply: replyCh,
				}:
				case <-ctx.Done():
					return "", 0, 0, ctx.Err()
				}
				var answer string
				select {
				case answer = <-replyCh:
				case <-ctx.Done():
					return "", 0, 0, ctx.Err()
				}
				toolResults = append(toolResults, api.ContentBlock{
					Type:      "tool_result",
					ToolUseID: tb.id,
					Content:   []api.ContentBlock{{Type: "text", Text: answer}},
				})
				continue
			}

			select {
			case events <- TurnEvent{Type: TurnEventToolStart, ToolName: tb.name, ToolInput: summary}:
			case <-ctx.Done():
				return "", 0, 0, ctx.Err()
			}

			result := loop.ExecuteToolQuiet(tb.name, inputMap)
			result.ToolUseID = tb.id
			toolResults = append(toolResults, result)

			resultText := ""
			if len(result.Content) > 0 {
				resultText = result.Content[0].Text
			}
			select {
			case events <- TurnEvent{Type: TurnEventToolDone, ToolName: tb.name, ToolResult: resultText}:
			case <-ctx.Done():
				return "", 0, 0, ctx.Err()
			}
		}

		loop.Session.Messages = append(loop.Session.Messages, api.Message{
			Role:    "user",
			Content: toolResults,
		})
	}

	// Fallback estimate for output tokens when the provider didn't
	// surface one (see runOneTurn for full rationale). The TUI
	// surfaces outputTokens in the usage line at the bottom of the
	// prompt, so a tool-only response should still report >0 tokens
	// — otherwise the user gets a confusing "0 tokens used" right
	// after a multi-tool turn.
	if outputTokens == 0 && streamedOutputChars > 0 {
		outputTokens = streamedOutputChars / charsPerToken
		if outputTokens == 0 {
			outputTokens = 1
		}
	}

	return stopReason, inputTokens, outputTokens, nil
}

// ExecuteToolQuiet dispatches to the appropriate tool without printing to stdout/stderr.
func (loop *ConversationLoop) ExecuteToolQuiet(name string, input map[string]any) api.ContentBlock {
	if !CheckPermission(loop.Permissions, name) {
		return api.ContentBlock{
			Type:    "tool_result",
			Content: []api.ContentBlock{{Type: "text", Text: fmt.Sprintf("Permission denied for tool: %s", name)}},
			IsError: true,
		}
	}

	var result string
	var err error

	switch name {
	case "bash":
		result, err = tools.ExecuteBash(input)
	case "read_file":
		result, err = tools.ExecuteReadFile(input)
	case "write_file":
		result, err = tools.ExecuteWriteFile(input)
	case "glob":
		result, err = tools.ExecuteGlob(input)
	case "grep":
		result, err = tools.ExecuteGrep(input)
	case "file_edit":
		result, err = tools.ExecuteFileEdit(input)
	case "web_fetch":
		result, err = tools.ExecuteWebFetch(input)
	case "web_search":
		result, err = tools.ExecuteWebSearch(input)
	case "ask_user":
		q, ok := tools.AskUserInput(input)
		if !ok {
			err = fmt.Errorf("ask_user: 'question' is required")
		} else {
			return tools.AskUserFallback(q)
		}
	case "todo_write":
		result, err = tools.ExecuteTodoWrite(input)
	default:
		// Fall back to MCP registry.
		if loop.MCPRegistry != nil {
			if client, _, ok := loop.MCPRegistry.FindTool(name); ok {
				mcpResult, mcpErr := client.CallTool(context.Background(), name, input)
				if mcpErr != nil {
					return api.ContentBlock{
						Type:    "tool_result",
						Content: []api.ContentBlock{{Type: "text", Text: fmt.Sprintf("Error: %v", mcpErr)}},
						IsError: true,
					}
				}
				text := mcpResultText(mcpResult)
				return api.ContentBlock{
					Type:    "tool_result",
					Content: []api.ContentBlock{{Type: "text", Text: text}},
					IsError: mcpResult.IsError,
				}
			}
		}
		err = fmt.Errorf("unknown tool: %s", name)
	}

	isError := err != nil
	text := result
	if err != nil {
		text = fmt.Sprintf("Error: %v", err)
	}

	return api.ContentBlock{
		Type:    "tool_result",
		Content: []api.ContentBlock{{Type: "text", Text: text}},
		IsError: isError,
	}
}

// summarizeToolInput returns a short human-readable summary of tool inputs.
func summarizeToolInput(input map[string]any) string {
	for _, key := range []string{"command", "path", "file_path", "pattern", "url", "query", "question"} {
		if v, ok := input[key].(string); ok {
			if len(v) > 60 {
				return v[:60] + "..."
			}
			return v
		}
	}
	return ""
}

// CompactNow forces a compaction regardless of the ShouldCompact threshold.
// Used by the /compact command and by the auto-recovery path when the
// provider rejects a request with a prompt-too-large error.
func (loop *ConversationLoop) CompactNow(ctx context.Context) (string, error) {
	if !loop.Config.CompactionEnabled {
		return "", fmt.Errorf("compaction is disabled in config")
	}
	summary, err := CompactSession(ctx, loop.Client, loop.Config, loop.Session)
	if err != nil {
		return "", err
	}
	loop.Compaction.CompactionCount++
	contMsg := GetContinuationMessage(summary)
	loop.Session.Messages = append([]api.Message{contMsg}, loop.Session.Messages...)
	return summary, nil
}

// ClearSession resets the conversation history in the current session.
func (loop *ConversationLoop) ClearSession() {
	loop.Session.Messages = []api.Message{}
}

// ListSessions returns all session IDs saved in the configured session directory.
func (loop *ConversationLoop) ListSessions() ([]string, error) {
	return ListSessions(loop.Config.SessionDir)
}

// SaveCurrentSession persists the active session to disk, including usage data.
func (loop *ConversationLoop) SaveCurrentSession() error {
	if loop.Usage != nil {
		loop.Session.TotalInputTokens = loop.Usage.TotalInput
		loop.Session.TotalOutputTokens = loop.Usage.TotalOutput
		loop.Session.TotalTurns = loop.Usage.Turns
	}
	return SaveSession(loop.Config.SessionDir, loop.Session)
}

// LoadNamedSession replaces the active session with one loaded from disk by ID.
// Usage tracker state is restored from persisted session data.
func (loop *ConversationLoop) LoadNamedSession(id string) error {
	sess, err := LoadSession(loop.Config.SessionDir, id)
	if err != nil {
		return err
	}
	loop.Session = sess
	if loop.Usage != nil && sess.TotalTurns > 0 {
		loop.Usage.TotalInput = sess.TotalInputTokens
		loop.Usage.TotalOutput = sess.TotalOutputTokens
		loop.Usage.Turns = sess.TotalTurns
	}
	return nil
}

// ListSessionsWithMeta returns metadata for all saved sessions, sorted newest first.
func (loop *ConversationLoop) ListSessionsWithMeta() ([]SessionMeta, error) {
	return ListSessionsWithMeta(loop.Config.SessionDir)
}

// MessageCount returns the number of messages in the active session.
func (loop *ConversationLoop) MessageCount() int {
	if loop.Session == nil {
		return 0
	}
	return len(loop.Session.Messages)
}

// ExecuteTool dispatches to the appropriate tool implementation.
func (loop *ConversationLoop) ExecuteTool(name string, input map[string]any) api.ContentBlock {
	if !CheckPermission(loop.Permissions, name) {
		return api.ContentBlock{
			Type:    "tool_result",
			Content: []api.ContentBlock{{Type: "text", Text: fmt.Sprintf("Permission denied for tool: %s", name)}},
			IsError: true,
		}
	}

	var result string
	var err error

	switch name {
	case "bash":
		result, err = tools.ExecuteBash(input)
	case "read_file":
		result, err = tools.ExecuteReadFile(input)
	case "write_file":
		result, err = tools.ExecuteWriteFile(input)
	case "glob":
		result, err = tools.ExecuteGlob(input)
	case "grep":
		result, err = tools.ExecuteGrep(input)
	case "file_edit":
		result, err = tools.ExecuteFileEdit(input)
	case "web_fetch":
		result, err = tools.ExecuteWebFetch(input)
	case "web_search":
		result, err = tools.ExecuteWebSearch(input)
	case "ask_user":
		q, ok := tools.AskUserInput(input)
		if !ok {
			err = fmt.Errorf("ask_user: 'question' is required")
		} else {
			cb := tools.AskUserFallback(q)
			fmt.Fprintf(os.Stdout, "%s\n", cb.Content[0].Text)
			return cb
		}
	case "todo_write":
		result, err = tools.ExecuteTodoWrite(input)
	default:
		// Fall back to MCP registry.
		if loop.MCPRegistry != nil {
			if client, _, ok := loop.MCPRegistry.FindTool(name); ok {
				mcpResult, mcpErr := client.CallTool(context.Background(), name, input)
				if mcpErr != nil {
					fmt.Fprintf(os.Stderr, "[MCP tool %s error]: %v\n", name, mcpErr)
					return api.ContentBlock{
						Type:    "tool_result",
						Content: []api.ContentBlock{{Type: "text", Text: fmt.Sprintf("Error: %v", mcpErr)}},
						IsError: true,
					}
				}
				text := mcpResultText(mcpResult)
				fmt.Fprintf(os.Stdout, "%s\n", text)
				return api.ContentBlock{
					Type:    "tool_result",
					Content: []api.ContentBlock{{Type: "text", Text: text}},
					IsError: mcpResult.IsError,
				}
			}
		}
		err = fmt.Errorf("unknown tool: %s", name)
	}

	isError := err != nil
	text := result
	if err != nil {
		text = fmt.Sprintf("Error: %v", err)
		fmt.Fprintf(os.Stderr, "[Tool %s error]: %v\n", name, err)
	} else {
		fmt.Fprintf(os.Stdout, "%s\n", result)
	}

	return api.ContentBlock{
		Type: "tool_result",
		Content: []api.ContentBlock{
			{Type: "text", Text: text},
		},
		IsError: isError,
	}
}

// mcpResultText extracts the concatenated text from an MCP tool result.
func mcpResultText(r mcp.MCPToolResult) string {
	var parts []string
	for _, c := range r.Content {
		if c.Text != "" {
			parts = append(parts, c.Text)
		}
	}
	return strings.Join(parts, "\n")
}

// InitMCPFromConfig connects to all MCP servers defined in the config.
// Errors are printed but do not abort startup.
func (loop *ConversationLoop) InitMCPFromConfig(ctx context.Context) {
	if len(loop.Config.MCPServers) == 0 {
		return
	}
	if loop.MCPRegistry == nil {
		loop.MCPRegistry = mcp.NewRegistry()
	}
	for _, srv := range loop.Config.MCPServers {
		var transport mcp.Transport
		var err error

		switch strings.ToLower(srv.Transport) {
		case "stdio":
			var envPairs []string
			for k, v := range srv.Env {
				envPairs = append(envPairs, k+"="+v)
			}
			transport, err = mcp.NewStdioTransport(srv.Command, srv.Args, envPairs)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[MCP] failed to start stdio server %q: %v\n", srv.Name, err)
				continue
			}
		case "sse", "http":
			auth := ""
			if tok, ok := srv.Env["AUTHORIZATION"]; ok {
				auth = tok
			}
			transport = mcp.NewSSETransport(srv.URL, auth)
		default:
			fmt.Fprintf(os.Stderr, "[MCP] unknown transport %q for server %q\n", srv.Transport, srv.Name)
			continue
		}

		if err := loop.MCPRegistry.AddServer(ctx, srv.Name, transport); err != nil {
			fmt.Fprintf(os.Stderr, "[MCP] failed to connect to server %q: %v\n", srv.Name, err)
		} else {
			toolCount := len(loop.MCPRegistry.ServerTools(srv.Name))
			fmt.Fprintf(os.Stdout, "[MCP] connected to %q (%d tools)\n", srv.Name, toolCount)
		}
	}
}

// MCPConnect connects to a named MCP server defined in config.
func (loop *ConversationLoop) MCPConnect(ctx context.Context, name string) error {
	if loop.MCPRegistry == nil {
		loop.MCPRegistry = mcp.NewRegistry()
	}
	for _, srv := range loop.Config.MCPServers {
		if srv.Name != name {
			continue
		}
		var transport mcp.Transport
		var err error
		switch strings.ToLower(srv.Transport) {
		case "stdio":
			var envPairs []string
			for k, v := range srv.Env {
				envPairs = append(envPairs, k+"="+v)
			}
			transport, err = mcp.NewStdioTransport(srv.Command, srv.Args, envPairs)
			if err != nil {
				return err
			}
		case "sse", "http":
			auth := ""
			if tok, ok := srv.Env["AUTHORIZATION"]; ok {
				auth = tok
			}
			transport = mcp.NewSSETransport(srv.URL, auth)
		default:
			return fmt.Errorf("unknown transport %q", srv.Transport)
		}
		return loop.MCPRegistry.AddServer(ctx, name, transport)
	}
	return fmt.Errorf("MCP server %q not found in config", name)
}

// MCPDisconnect disconnects from a named MCP server.
func (loop *ConversationLoop) MCPDisconnect(name string) error {
	if loop.MCPRegistry == nil {
		return fmt.Errorf("no MCP servers connected")
	}
	return loop.MCPRegistry.Disconnect(name)
}

// MCPList returns a human-readable summary of connected MCP servers and their tools.
func (loop *ConversationLoop) MCPList() string {
	if loop.MCPRegistry == nil {
		return "No MCP servers connected.\n"
	}
	names := loop.MCPRegistry.ServerNames()
	if len(names) == 0 {
		return "No MCP servers connected.\n"
	}
	var sb strings.Builder
	for _, name := range names {
		tools := loop.MCPRegistry.ServerTools(name)
		fmt.Fprintf(&sb, "Server: %s (%d tools)\n", name, len(tools))
		for _, t := range tools {
			desc := t.Description
			if len(desc) > 60 {
				desc = desc[:60] + "..."
			}
			fmt.Fprintf(&sb, "  - %s: %s\n", t.Name, desc)
		}
	}
	return sb.String()
}

// promptTooLargeMarkers are substrings providers commonly use when the
// input exceeds the model's context cap. The auto-recovery path in
// SendMessage watches for these and triggers a compaction + retry. Keep
// the list tight — false positives would silently compact when the real
// issue is something else.
var promptTooLargeMarkers = []string{
	"prompt too large",
	"context length exceeded",
	"maximum context length",
	"context_length_exceeded",
	"input is too long",
}

func isPromptTooLargeError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	for _, marker := range promptTooLargeMarkers {
		if strings.Contains(msg, marker) {
			return true
		}
	}
	return false
}
