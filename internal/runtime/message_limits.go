package runtime

import (
	"claw-code-go/internal/api"
	"fmt"
	"os"
	"strings"
)

// Message size limits to prevent individual messages from exceeding provider caps.
// These are conservative estimates that work across all providers.
const (
	// maxMessageChars is the maximum character count for a single message content block.
	// Set conservatively below most provider limits (DeepSeek ~160K, Anthropic ~200K).
	maxMessageChars = 150_000

	// maxMessageTokens is the estimated token budget for a single message.
	// Using 4 chars/token heuristic: 150K chars ≈ 37.5K tokens.
	maxMessageTokens = 37_500

	// truncationMarker is appended to truncated content.
	truncationMarker = "\n\n[... content truncated to fit message size limit ...]\n\nNote: This message was automatically truncated because it exceeded the maximum size. The conversation history is preserved on the server side."
)

// ValidateAndTruncateMessage checks if a message exceeds size limits and truncates if needed.
// This prevents individual messages from being rejected by the provider API.
// Returns the potentially modified message and a boolean indicating if truncation occurred.
func ValidateAndTruncateMessage(msg api.Message) (api.Message, bool) {
	truncated := false

	// Check each content block
	for i := range msg.Content {
		block := &msg.Content[i]

		// Only truncate text blocks (images, tool results, etc. are handled differently)
		if block.Type != "text" {
			continue
		}

		textLen := len(block.Text)
		if textLen <= maxMessageChars {
			continue
		}

		// Truncate at a line boundary when possible for readability
		cut := maxMessageChars - len(truncationMarker) - 100
		if cut <= 0 {
			block.Text = truncationMarker
			truncated = true
			continue
		}

		truncatedText := block.Text[:cut]

		// Try to cut at a line boundary
		if nl := strings.LastIndex(truncatedText, "\n"); nl > cut-500 && nl > 0 {
			truncatedText = truncatedText[:nl]
		}

		block.Text = truncatedText + truncationMarker
		truncated = true

		fmt.Fprintf(os.Stderr, "[message-truncate] truncated text block from %d to %d chars\n",
			textLen, len(block.Text))
	}

	return msg, truncated
}

// EstimateMessageTokens returns a rough token estimate for a message.
// Uses the 4 chars/token heuristic that works reasonably well for English + code.
func EstimateMessageTokens(msg api.Message) int {
	totalChars := 0
	for _, block := range msg.Content {
		switch block.Type {
		case "text":
			totalChars += len(block.Text)
		case "tool_result":
			// Tool results can be large; estimate their content
			if block.Content != nil {
				for _, subBlock := range block.Content {
					if subBlock.Type == "text" {
						totalChars += len(subBlock.Text)
					}
				}
			}
		case "tool_use":
			// Tool use blocks have JSON input; estimate size
			if block.Input != nil {
				// Rough estimate: JSON serialization is ~2x the field count
				totalChars += len(fmt.Sprintf("%v", block.Input)) * 2
			}
		}
	}

	const charsPerToken = 4
	if totalChars == 0 {
		return 0
	}
	return (totalChars + charsPerToken - 1) / charsPerToken
}

// ShouldTruncateMessage returns true if a message should be truncated before sending.
func ShouldTruncateMessage(msg api.Message) bool {
	return EstimateMessageTokens(msg) > maxMessageTokens
}

// ValidateAndTruncateResults truncates individual tool-result content blocks
// that exceed maxMessageChars. This prevents massive file reads or command
// outputs from filling the session with a single bloated message.
func ValidateAndTruncateResults(blocks []api.ContentBlock) []api.ContentBlock {
	for i := range blocks {
		block := &blocks[i]
		if block.Type != "tool_result" {
			continue
		}
		for j := range block.Content {
			inner := &block.Content[j]
			if inner.Type != "text" {
				continue
			}
			textLen := len(inner.Text)
			if textLen <= maxMessageChars {
				continue
			}
			cut := maxMessageChars - len(truncationMarker) - 100
			if cut <= 0 {
				inner.Text = truncationMarker
				continue
			}
			truncatedText := inner.Text[:cut]
			if nl := strings.LastIndex(truncatedText, "\n"); nl > cut-500 && nl > 0 {
				truncatedText = truncatedText[:nl]
			}
			inner.Text = truncatedText + truncationMarker
			fmt.Fprintf(os.Stderr, "[tool-result-truncate] truncated tool result from %d to %d chars\n",
				textLen, len(inner.Text))
		}
	}
	return blocks
}
