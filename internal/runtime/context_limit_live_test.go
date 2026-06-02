//go:build live

package runtime

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"claw-code-go/internal/api"
	"claw-code-go/internal/api/providers/deepseek"
)

// TestLiveContextLimits performs live testing to determine the actual
// context window limits for DeepSeek models by progressively sending
// larger messages until we hit an error.
func TestLiveContextLimits(t *testing.T) {
	tok := os.Getenv("DEEPSEEK_TOKEN")
	if tok == "" {
		t.Skip("DEEPSEEK_TOKEN not set; skipping live context limit test")
	}

	tests := []struct {
		model           string
		startTokens     int
		incrementTokens int
		maxAttempts     int
	}{
		{
			model:           "expert",
			startTokens:     30000,
			incrementTokens: 2000,
			maxAttempts:     20,
		},
		{
			model:           "instant",
			startTokens:     30000,
			incrementTokens: 2000,
			maxAttempts:     20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			provider := deepseek.New()
			client, err := provider.NewClient(api.ProviderConfig{
				APIKey: tok,
				Model:  tt.model,
			})
			if err != nil {
				t.Fatalf("NewClient: %v", err)
			}

			ctx := context.Background()
			currentTokens := tt.startTokens
			lastSuccess := 0

			for attempt := 0; attempt < tt.maxAttempts; attempt++ {
				// Generate text with approximately currentTokens tokens
				// Using 4 chars per token heuristic
				text := generateText(currentTokens * 4)

				t.Logf("Attempt %d: Testing with ~%d tokens (%d chars)",
					attempt+1, currentTokens, len(text))

				req := api.CreateMessageRequest{
					Model:     tt.model,
					MaxTokens: 100, // Small output, we're testing input
					Messages: []api.Message{
						{
							Role: "user",
							Content: []api.ContentBlock{
								{Type: "text", Text: text},
							},
						},
					},
					Stream: true,
				}

				ch, err := client.StreamResponse(ctx, req)
				if err != nil {
					t.Logf("❌ Failed at ~%d tokens: %v", currentTokens, err)
					t.Logf("✅ Last successful: ~%d tokens", lastSuccess)
					return
				}

				// Drain the stream
				success := true
				for event := range ch {
					if event.Type == api.EventError {
						t.Logf("❌ Stream error at ~%d tokens: %s",
							currentTokens, event.ErrorMessage)
						t.Logf("✅ Last successful: ~%d tokens", lastSuccess)
						success = false
						break
					}
				}

				if !success {
					return
				}

				t.Logf("✅ Success at ~%d tokens", currentTokens)
				lastSuccess = currentTokens
				currentTokens += tt.incrementTokens

				// Rate limiting
				time.Sleep(2 * time.Second)
			}

			t.Logf("✅ Final successful: ~%d tokens (reached max attempts)", lastSuccess)
		})
	}
}

// TestLiveHybridContextManagement tests the hybrid context management
// approach with real API calls to verify it works correctly.
func TestLiveHybridContextManagement(t *testing.T) {
	tok := os.Getenv("DEEPSEEK_TOKEN")
	if tok == "" {
		t.Skip("DEEPSEEK_TOKEN not set; skipping live hybrid test")
	}

	provider := deepseek.New()
	client, err := provider.NewClient(api.ProviderConfig{
		APIKey: tok,
		Model:  "expert",
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	// Create a conversation with many messages
	messages := generateLargeConversation(100)

	t.Logf("Generated conversation with %d messages", len(messages))
	t.Logf("Estimated tokens: %d", EstimateTokens(messages))

	// Test Path 3: Importance-based pruning
	t.Run("ImportancePruning", func(t *testing.T) {
		targetTokens := 20000
		pruned := PruneByImportance(messages, targetTokens)

		t.Logf("Pruned from %d to %d messages", len(messages), len(pruned))
		t.Logf("Token reduction: %d -> %d",
			EstimateTokens(messages), EstimateTokens(pruned))

		// Verify we can send the pruned conversation
		ctx := context.Background()
		req := api.CreateMessageRequest{
			Model:     "expert",
			MaxTokens: 100,
			Messages: append(pruned, api.Message{
				Role: "user",
				Content: []api.ContentBlock{
					{Type: "text", Text: "Summarize our conversation."},
				},
			}),
			Stream: true,
		}

		ch, err := client.StreamResponse(ctx, req)
		if err != nil {
			t.Fatalf("Failed to send pruned conversation: %v", err)
		}

		// Drain stream
		for event := range ch {
			if event.Type == api.EventError {
				t.Fatalf("Stream error: %s", event.ErrorMessage)
			}
		}

		t.Logf("✅ Successfully sent pruned conversation")
	})

	// Test Path 1: Compaction with summarization
	t.Run("Compaction", func(t *testing.T) {
		cfg := &Config{
			Model:                    "expert",
			MaxTokens:                2048,
			CompactionEnabled:        true,
			CompactionThreshold:      0.75,
			CompactionKeepRecent:     10,
			CompactionMaxInputTokens: 38000,
		}

		session := &Session{
			ID:       "test-session",
			Messages: messages,
		}

		ctx := context.Background()
		summary, err := CompactSession(ctx, client, cfg, session)
		if err != nil {
			t.Fatalf("CompactSession failed: %v", err)
		}

		t.Logf("✅ Compaction successful")
		t.Logf("Summary length: %d chars", len(summary))
		t.Logf("Messages after compaction: %d", len(session.Messages))
		t.Logf("Summary preview: %s", truncateText(summary, 200))
	})
}

// generateText creates a string with approximately the specified number of characters
func generateText(chars int) string {
	const paragraph = "This is a test paragraph that will be repeated many times to create a large text block for testing context window limits. It contains various words and punctuation to simulate real conversation content. "

	repetitions := chars / len(paragraph)
	if repetitions < 1 {
		repetitions = 1
	}

	var sb strings.Builder
	sb.WriteString("Please analyze the following large text and provide a brief summary:\n\n")

	for i := 0; i < repetitions; i++ {
		sb.WriteString(paragraph)
	}

	return sb.String()
}

// generateLargeConversation creates a realistic conversation with many messages
func generateLargeConversation(numMessages int) []api.Message {
	messages := make([]api.Message, 0, numMessages)

	for i := 0; i < numMessages; i++ {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}

		var content []api.ContentBlock

		if role == "user" {
			// User messages
			texts := []string{
				"Can you help me with this task?",
				"Please write a function to calculate fibonacci numbers.",
				"There's an error in the code, can you fix it?",
				"Run the tests and check if they pass.",
				"What's the status of the project?",
			}
			content = []api.ContentBlock{
				{Type: "text", Text: texts[i%len(texts)]},
			}
		} else {
			// Assistant messages with tool calls
			if i%4 == 1 {
				// Tool call message
				content = []api.ContentBlock{
					{Type: "text", Text: "I'll help you with that."},
					{
						Type: "tool_use",
						ID:   fmt.Sprintf("tool_%d", i),
						Name: "write_to_file",
						Input: map[string]interface{}{
							"file_path": "/tmp/test.go",
							"content":   "package main\n\nfunc main() {}\n",
						},
					},
				}
			} else if i%4 == 3 {
				// Tool result message
				content = []api.ContentBlock{
					{
						Type:      "tool_result",
						ToolUseID: fmt.Sprintf("tool_%d", i-2),
						Content: []api.ContentBlock{
							{Type: "text", Text: "Successfully wrote file."},
						},
					},
				}
			} else {
				// Regular text message
				content = []api.ContentBlock{
					{Type: "text", Text: "Here's what I found: The code looks good and should work correctly."},
				}
			}
		}

		messages = append(messages, api.Message{
			Role:    role,
			Content: content,
		})
	}

	return messages
}

func truncateText(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
