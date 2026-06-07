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

// EstimateTokens estimates the number of tokens in a slice of messages.
// When the DeepSeek V3 tokenizer is available (tokenizer.json bundled),
// it uses accurate BPE token counting. Falls back to chars/4 heuristic
// when the tokenizer is unavailable.
func EstimateTokens(messages []api.Message) int {
	text := buildMessagesText(messages)
	return dsTokenizer.CountTokens(text)
}

// buildMessageText converts a single api.Message to the text
// representation that the tokenizer will count.
func buildMessageText(msg api.Message) string {
	var sb strings.Builder
	sb.WriteString(msg.Role)
	sb.WriteByte('\n')
	for _, cb := range msg.Content {
		switch cb.Type {
		case "text":
			sb.WriteString(cb.Text)
		case "tool_use":
			sb.WriteString(cb.Name)
			if cb.Input != nil {
				for k, v := range cb.Input {
					sb.WriteString(k)
					if s, ok := v.(string); ok {
						sb.WriteString(s)
					}
				}
			}
		case "tool_result":
			// tool_result blocks carry their text in inner Content blocks.
			// Adding an explicit case makes the intent clear and guards
			// against future data-model changes where Text might be set directly.
		}
		for _, inner := range cb.Content {
			sb.WriteString(inner.Text)
		}
	}
	return sb.String()
}

// buildMessagesText flattens a message list into a single string for
// accurate token counting. It concatenates role+content for each
// message, matching how the DeepSeek API serializes messages.
func buildMessagesText(messages []api.Message) string {
	var sb strings.Builder
	for _, msg := range messages {
		sb.WriteString(msg.Role)
		sb.WriteByte('\n')
		for _, cb := range msg.Content {
			switch cb.Type {
			case "text":
				sb.WriteString(cb.Text)
			case "tool_use":
				sb.WriteString(cb.Name)
				if cb.Input != nil {
					for k, v := range cb.Input {
						sb.WriteString(k)
						if s, ok := v.(string); ok {
							sb.WriteString(s)
						}
					}
				}
			case "tool_result":
				// tool_result blocks carry their text in inner Content blocks.
			}
			for _, inner := range cb.Content {
				sb.WriteString(inner.Text)
			}
		}
	}
	return sb.String()
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
// client supports it (non-zero), then cfg.CompactionMaxInputTokens. When
// neither is available a conservative default of 50000 is used — this is a
// sensible floor for most modern LLMs (even small models have ≥128k context).
// cfg.MaxTokens is NOT used because it represents the output token budget,
// not the input context window (e.g. 8192 output vs 65536 input).
//
// The inputTokens parameter is treated as a hint. When it's ≤0 or stale
// (i.e. from a prior turn's LastInputTokens rather than the current
// message list), the function falls back to EstimateTokens(messages) which
// computes a fresh chars/4 estimate of the full message list. Callers
// should pass 0 or the current estimate to get an accurate comparison.
func ShouldCompact(inputTokens int, messages []api.Message, cfg *Config, client api.APIClient) bool {
	if !cfg.CompactionEnabled {
		return false
	}

	// Always compute a fresh estimate from the message list. Use the
	// passed-in inputTokens only when it's more recent than the
	// message-list estimate — this guards against stale LastInputTokens
	// from a prior SendMessage call.
	estimatedFromMessages := EstimateTokens(messages)
	if inputTokens <= 0 || estimatedFromMessages > inputTokens {
		inputTokens = estimatedFromMessages
	}

	// Delta-mode detection: when the API-reported input token count is
	// far below the full message-list estimate, we're in delta mode and
	// the server tracks state we don't see. Always use the full estimate
	// in that case so we don't silently ignore server-side context growth.
	if cfg.DeltaMode {
		inputTokens = estimatedFromMessages
	}

	// Basis: prefer client.MaxInputTokens() (single source of truth),
	// then cfg.CompactionMaxInputTokens, then a conservative default.
	basis := 0
	if client != nil {
		basis = client.MaxInputTokens()
	}
	if basis <= 0 {
		basis = cfg.CompactionMaxInputTokens
	}
	if basis <= 0 {
		basis = 50000 // conservative default: modern LLMs have ≥128k context
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
	if dsTokenizer != nil {
		text := buildMessageText(msg)
		if tokens := dsTokenizer.CountTokens(text); tokens > 0 {
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
