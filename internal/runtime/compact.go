package runtime

import (
	"context"
	"fmt"
	"strings"

	"claw-code-go/internal/api"
)

const (
	// DefaultCompactionThreshold triggers compaction when input tokens reach
	// this fraction of MaxTokens (e.g., 0.75 = 75%).
	DefaultCompactionThreshold = 0.75

	// DefaultCompactionKeepRecent is the number of recent messages retained
	// verbatim after compaction.
	DefaultCompactionKeepRecent = 10

	// charsPerToken is the approximate character-to-token ratio used for estimation.
	charsPerToken = 4
)

// CompactionState tracks token usage and compaction history across turns.
type CompactionState struct {
	LastInputTokens   int // input tokens from the most recently completed turn
	TotalInputTokens  int // cumulative input tokens across all turns
	TotalOutputTokens int // cumulative output tokens across all turns
	CompactionCount   int // number of times the session has been compacted
}

// EstimateTokens returns the token count for a slice of messages using the
// DeepSeek tokenizer when available, falling back to a chars/4 heuristic.
// Prefer CountMessagesTokens for new code; this wrapper exists for callers
// that predate the tokenizer integration.
func EstimateTokens(messages []api.Message) int {
	return CountMessagesTokens(messages)
}

// estimateTokensFallback is the original chars-per-token heuristic, kept for
// reference. CountMessagesTokens uses this as its fallback when the
// tokenizer is not loaded.
func estimateTokensFallback(messages []api.Message) int {
	var total int
	for _, msg := range messages {
		for _, cb := range msg.Content {
			switch cb.Type {
			case "text":
				total += len(cb.Text) / charsPerToken
			case "tool_use":
				// Tool use blocks carry their input as a map; estimate
				// from the serialised form. The provider sends this as
				// input_json_delta so counting chars/4 is consistent with
				// how we estimate streamed output.
				if cb.Input != nil {
					for k, v := range cb.Input {
						total += len(k) / charsPerToken
						if s, ok := v.(string); ok {
							total += len(s) / charsPerToken
						}
					}
				}
			}
			// Nested content blocks (tool_result inner blocks)
			for _, inner := range cb.Content {
				total += len(inner.Text) / charsPerToken
			}
		}
	}
	return total
}

// ShouldCompact returns true when the session should be compacted.
// It uses the actual API-reported input token count when available (> 0),
// falling back to EstimateTokens.
//
// In delta mode (DeepSeek), the API-reported input token count reflects
// only the delta prompt size (system + tools + last message), which is
// typically 2-3K tokens — far below the real threshold. In that case we
// always fall back to EstimateTokens on the full message list so we
// don't silently ignore server-side context growth.
//
// The threshold is computed against the client's MaxInputTokens() when the
// client supports it (non-zero), then cfg.CompactionMaxInputTokens, then
// cfg.MaxTokens as a last resort. Providers with very large context windows
// (e.g. DeepSeek's web API) advertise their real limit via the client
// interface so we don't compact based on the per-turn output budget.
func ShouldCompact(inputTokens int, messages []api.Message, cfg *Config, client api.APIClient) bool {
	if !cfg.CompactionEnabled {
		return false
	}
	if inputTokens <= 0 {
		inputTokens = EstimateTokens(messages)
	}
	// In delta mode the API-reported input token count reflects only
	// In delta mode (and the DeepSeek web API, which uses server-side
	// conversation tracking regardless of the DeltaMode flag), the
	// API-reported input token count reflects only the delta prompt
	// (system + tools + last message), typically 2-3K tokens per turn.
	// The server tracks the full conversation, so we must estimate from
	// the complete message list.
	//
	// Detection strategy:
	//   1. If DeltaMode is explicitly enabled, always use the estimate.
	//   2. Otherwise, if the full message-list estimate exceeds the
	//      reported input, the provider is almost certainly using
	//      server-side conversation tracking (delta-like). Use the
	//      estimate. A true full-history API would report input >=
	//      estimate since the estimate is a lower bound.
	//      The old 5x heuristic failed because accumulated per-turn
	//      delta input (sum of ~2-3K/turn across N turns) eventually
	//      exceeds estimatedFromMessages / 5, making the ratio < 5x
	//      and preventing compaction from ever triggering.
	if cfg.DeltaMode {
		inputTokens = EstimateTokens(messages)
	} else {
		estimatedFromMessages := EstimateTokens(messages)
		if estimatedFromMessages > inputTokens {
			inputTokens = estimatedFromMessages
		}
	}
	// Prefer the client's own limit (single source of truth for the model).
	basis := 0
	if client != nil {
		basis = client.MaxInputTokens()
	}
	if basis <= 0 {
		basis = cfg.CompactionMaxInputTokens
	}
	if basis <= 0 {
		basis = cfg.MaxTokens
	}
	threshold := int(float64(basis) * cfg.CompactionThreshold)
	return inputTokens >= threshold
}

// collectStreamText drains a StreamEvent channel and returns the concatenated
// text response, or an error if the stream encountered one.
func collectStreamText(ch <-chan api.StreamEvent) (string, error) {
	var sb strings.Builder
	for event := range ch {
		switch event.Type {
		case api.EventError:
			return "", fmt.Errorf("stream error: %s", event.ErrorMessage)
		case api.EventContentBlockDelta:
			if event.Delta.Type == "text_delta" {
				sb.WriteString(event.Delta.Text)
			}
		}
	}
	return sb.String(), nil
}

// buildTranscript constructs a plain-text transcript of the messages suitable
// for submission to the summarization model.
func buildTranscript(messages []api.Message) string {
	var sb strings.Builder
	sb.WriteString("Please provide a concise but thorough summary of the following conversation. ")
	sb.WriteString("Preserve all important technical details: file paths modified, commands run, ")
	sb.WriteString("decisions made, errors encountered, and the current state of any ongoing work. ")
	sb.WriteString("The summary will replace the conversation history and must be self-contained.\n\n")
	sb.WriteString("---CONVERSATION---\n")

	for _, msg := range messages {
		fmt.Fprintf(&sb, "\n[%s]:\n", strings.ToUpper(msg.Role))
		for _, cb := range msg.Content {
			switch cb.Type {
			case "text":
				if len(cb.Text) > 1000 {
					sb.WriteString(cb.Text[:1000])
					sb.WriteString("... [truncated]\n")
				} else {
					sb.WriteString(cb.Text)
					sb.WriteString("\n")
				}
			case "tool_use":
				fmt.Fprintf(&sb, "[Tool call: %s]\n", cb.Name)
			case "tool_result":
				for _, inner := range cb.Content {
					if len(inner.Text) > 300 {
						fmt.Fprintf(&sb, "[Tool result: %s... (truncated)]\n", inner.Text[:300])
					} else if inner.Text != "" {
						fmt.Fprintf(&sb, "[Tool result: %s]\n", inner.Text)
					}
				}
			}
		}
	}
	sb.WriteString("\n---END CONVERSATION---\n")
	return sb.String()
}

// CompactSession summarizes the session's message history by calling the model,
// stores the summary in the session, and trims the message list to the most
// recent cfg.CompactionKeepRecent messages. Returns the summary text.
//
// Uses a dedicated instant model client for reliability:
// - Creates a separate client configured for instant model
// - Ensures compaction works even if expert model is overloaded or failing
// - The summary is then used in the main session (which may use expert model)
func CompactSession(ctx context.Context, client api.APIClient, cfg *Config, session *Session) (string, error) {
	if len(session.Messages) == 0 {
		return "", nil
	}

	transcript := buildTranscript(session.Messages)

	// Create a dedicated compaction client using instant model
	// This ensures compaction always succeeds even if expert model fails
	compactClient := NewCompactClient(cfg, client)

	req := api.CreateMessageRequest{
		Model:     "instant", // Always use instant for compaction
		MaxTokens: 2048,
		Messages: []api.Message{
			{
				Role: "user",
				Content: []api.ContentBlock{
					{Type: "text", Text: transcript},
				},
			},
		},
		Stream: true,
	}

	ch, err := compactClient.StreamResponse(ctx, req)
	if err != nil {
		return "", fmt.Errorf("compact: stream response: %w", err)
	}

	summary, err := collectStreamText(ch)
	if err != nil {
		return "", fmt.Errorf("compact: collect stream: %w", err)
	}

	// Retain the most recent N messages verbatim.
	keepCount := cfg.CompactionKeepRecent
	if keepCount > len(session.Messages) {
		keepCount = len(session.Messages)
	}
	recent := make([]api.Message, keepCount)
	copy(recent, session.Messages[len(session.Messages)-keepCount:])

	session.CompactionSummary = summary
	session.CompactionCount++
	session.Messages = recent

	return summary, nil
}

// FormatCompactSummary wraps a compaction summary for injection into the system
// prompt so the model is aware of earlier conversation context.
func FormatCompactSummary(summary string) string {
	return fmt.Sprintf(
		"<compacted_context>\nThe following is a summary of earlier conversation history that has been compacted to save context space:\n\n%s\n</compacted_context>",
		summary,
	)
}

// GetContinuationMessage creates a synthetic user message that announces the
// compaction event, suitable for prepending to the retained recent messages.
func GetContinuationMessage(summary string) api.Message {
	text := fmt.Sprintf(
		"[System: Conversation history was automatically compacted to stay within context limits.\n\nSummary of prior context:\n%s\n\nContinuing from here.]",
		summary,
	)
	return api.Message{
		Role: "user",
		Content: []api.ContentBlock{
			{Type: "text", Text: text},
		},
	}
}

// resetProviderSessionAfterCompaction calls ResetSession on the provider
// client if it implements the SessionResetter interface. This is critical
// for delta mode: after compaction trims the client-side message list, the
// DeepSeek server still holds the full pre-compaction conversation via
// parent_message_id chaining. Without this reset, the server-side context
// silently overflows the model's context window even though the client
// thinks it has been compacted.
func resetProviderSessionAfterCompaction(client api.APIClient) {
	if r, ok := client.(api.SessionResetter); ok {
		r.ResetSession()
	}
}

// TruncateFIFO removes the oldest messages, keeping only the most recent N
// messages that fit within fractionBasis of the model's token budget. This is
// a last-resort recovery strategy when smart compaction and model fallback
// have both failed. Based on live shootout results (2026-06-07): FIFO works
// reliably even at 141%+ of the reported token limit — the server's actual
// rejection threshold is well above the advertised cap.
//
// fractionBasis controls how aggressively we trim. 0.50 means keep messages
// whose estimated tokens total ≤ 50% of the model's MaxInputTokens.
func TruncateFIFO(session *Session, client api.APIClient, fractionBasis float64) int {
	if len(session.Messages) == 0 || client == nil {
		return 0
	}

	limit := client.MaxInputTokens()
	if limit <= 0 {
		limit = 50000 // fallback for clients that don't report MaxInputTokens
	}

	target := int(float64(limit) * fractionBasis)
	if target < 500 {
		target = 500 // don't trim below a reasonable minimum
	}

	// Walk from newest to oldest, accumulating token estimates.
	kept := 0
	cumulative := 0
	for i := len(session.Messages) - 1; i >= 0; i-- {
		msgTokens := estimateMessageTokens(session.Messages[i])
		if cumulative+msgTokens > target {
			break
		}
		cumulative += msgTokens
		kept++
	}

	if kept >= len(session.Messages) {
		return 0 // nothing to trim
	}

	trimmed := len(session.Messages) - kept
	newMsgs := make([]api.Message, kept)
	copy(newMsgs, session.Messages[len(session.Messages)-kept:])

	// Prepend a FIFO truncation notice so the model knows context was lost.
	notice := api.Message{
		Role: "user",
		Content: []api.ContentBlock{
			{Type: "text", Text: fmt.Sprintf(
				"[System: Conversation history was truncated (FIFO) to stay within context limits. "+
					"%d oldest messages were dropped. Some earlier context may be missing.]",
				trimmed,
			)},
		},
	}
	session.Messages = append([]api.Message{notice}, newMsgs...)
	return trimmed
}

// estimateMessageTokens estimates the token count for a single message.
// Uses the accurate DeepSeek V3 tokenizer when available, falling back
// to chars/4 heuristic.
func estimateMessageTokens(msg api.Message) int {
	// Try accurate tokenizer first.
	if TokenizerAvailable() {
		text := buildMessageText(msg)
		if tokens := CountTokens(text); tokens > 0 {
			return tokens
		}
	}

	// Fallback: chars/4 heuristic.
	total := 0
	for _, cb := range msg.Content {
		switch cb.Type {
		case "text":
			total += len(cb.Text) / charsPerToken
		case "tool_use":
			if cb.Input != nil {
				for k, v := range cb.Input {
					total += len(k) / charsPerToken
					if s, ok := v.(string); ok {
						total += len(s) / charsPerToken
					}
				}
			}
		}
		for _, inner := range cb.Content {
			total += len(inner.Text) / charsPerToken
		}
	}
	if total == 0 {
		total = 1
	}
	return total
}

// buildMessageText extracts the textual content from a message for token counting.
func buildMessageText(msg api.Message) string {
	var parts []string
	for _, cb := range msg.Content {
		switch cb.Type {
		case "text":
			if cb.Text != "" {
				parts = append(parts, cb.Text)
			}
		case "tool_use":
			if cb.Input != nil {
				for k, v := range cb.Input {
					parts = append(parts, k)
					if s, ok := v.(string); ok {
						parts = append(parts, s)
					}
				}
			}
		}
		for _, inner := range cb.Content {
			if inner.Text != "" {
				parts = append(parts, inner.Text)
			}
		}
	}
	result := ""
	for _, p := range parts {
		result += p
	}
	return result
}
