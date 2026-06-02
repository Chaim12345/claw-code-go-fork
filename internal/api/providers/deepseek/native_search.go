package deepseek

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

// NativeSearch performs a web search using the DeepSeek client's WebClient.
// It creates a standalone chat session with the Instant model (search enabled)
// and sends a single-turn search prompt. The response is streamed, collected,
// and returned as the result.
//
// This is designed to be plugged into the tools package's NativeSearchFunc
// so that the web_search tool delegates to DeepSeek's server-side search
// when the DeepSeek provider is active. The function creates its own
// short-lived chat session and does not interfere with the main
// conversation's session.
func (c *Client) NativeSearch(ctx context.Context, query string, numResults int) (string, error) {
	debugLog("NativeSearch: creating session for query: %s", query)

	sessID, err := c.web.CreateChatSession()
	if err != nil {
		return "", fmt.Errorf("deepseek native search: create session: %w", err)
	}
	debugLog("NativeSearch: session created: %s", sessID)

	// Build a prompt that asks for search results in a structured format.
	prompt := fmt.Sprintf(
		`Search the web for: %s

Return the top %d search results in this exact format for each result:
- Title of the result
- URL of the result
- A brief snippet or description

List each result with a number. Be concise and factual. Only return real search results from the web.`,
		query, numResults,
	)

	spec := ModelSpec{
		ModelType:       "default", // Instant model
		ThinkingEnabled: false,     // No thinking needed for simple search
		SearchEnabled:   true,      // Enable web search
	}

	var fullText strings.Builder
	sawFinished := false

	_, err = c.web.ChatCompletionStream(
		CompletionOpts{
			SessionID: sessID,
			Prompt:    prompt,
			Spec:      spec,
		},
		func(event StreamEvent) bool {
			if event.Event == "content" {
				fullText.WriteString(event.Data)
				return true
			}
			if event.Event == "status" && event.Data == "FINISHED" {
				sawFinished = true
				return false // stop parsing
			}
			return true
		},
	)

	if err != nil {
		// Log the partial output if we have any, then return the error.
		partial := fullText.String()
		if partial != "" {
			fmt.Fprintf(os.Stderr, "[web_search] partial results before error: %s\n", truncateForLog(partial, 200))
		}
		return "", fmt.Errorf("deepseek native search: %w", err)
	}

	text := fullText.String()
	if sawFinished {
		text = trimFinishedSentinel(text)
	}
	text = strings.TrimSpace(text)

	if text == "" {
		return "No search results found.", nil
	}

	// Add a header so the model knows this came from web search.
	result := fmt.Sprintf("Web search results for: %s\n\n%s\n\n[Search performed at %s; results may include real-time information.]",
		query, text, time.Now().Format(time.RFC3339))

	return result, nil
}

// truncateForLog shortens a string for log output.
func truncateForLog(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
