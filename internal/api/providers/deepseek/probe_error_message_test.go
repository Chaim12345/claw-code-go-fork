//go:build live

package deepseek

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"claw-code-go/internal/api"
)

// TestProbeExactErrorMessage sends oversized messages across all model variants
// and captures the EXACT error text the server returns.
func TestProbeExactErrorMessage(t *testing.T) {
	tok := os.Getenv("DEEPSEEK_TOKEN")
	if tok == "" {
		t.Skip("DEEPSEEK_TOKEN not set; skipping live test")
	}

	models := []string{"instant", "expert"}
	for _, model := range models {
		t.Run(model, func(t *testing.T) {
			provider := New()
			client, err := provider.NewClient(api.ProviderConfig{
				APIKey: tok,
				Model:  model,
			})
			if err != nil {
				t.Fatalf("NewClient: %v", err)
			}
			dsClient, ok := client.(*Client)
			if !ok {
				t.Skip("client is not *deepseek.Client")
			}

			sid, err := dsClient.web.CreateChatSession()
			if err != nil {
				t.Fatalf("CreateChatSession: %v", err)
			}
			dsClient.chatSessionID = sid
			dsClient.parentMessageID = ""

			sizes := []int{100_000, 500_000, 1_000_000, 2_000_000}
			for _, size := range sizes {
				t.Run(fmt.Sprintf("size_%dk", size/1000), func(t *testing.T) {
					captureViaStreamResponse(t, dsClient, model, size)
				})
				time.Sleep(2 * time.Second)
			}

			t.Run("delta_mode", func(t *testing.T) {
				dsClient.mu.Lock()
				dsClient.deltaMode = true
				dsClient.mu.Unlock()

				for _, size := range []int{100_000, 500_000, 1_000_000} {
					t.Run(fmt.Sprintf("size_%dk", size/1000), func(t *testing.T) {
						captureViaStreamResponse(t, dsClient, model, size)
					})
					time.Sleep(2 * time.Second)
				}

				dsClient.mu.Lock()
				dsClient.deltaMode = false
				dsClient.mu.Unlock()
			})
		})
	}
}

func captureViaStreamResponse(t *testing.T, c *Client, model string, size int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	prompt := strings.Repeat("A", size)
	t.Logf("Sending %d chars via StreamResponse (delta=%v)", size, c.deltaMode)

	req := api.CreateMessageRequest{
		Model:     model,
		MaxTokens: 32,
		Messages: []api.Message{
			{Role: "user", Content: []api.ContentBlock{{Type: "text", Text: prompt}}},
		},
		Stream: true,
	}

	ch, err := c.StreamResponse(ctx, req)
	if err != nil {
		t.Logf("=== PRE-FLIGHT ERROR ===")
		t.Logf("%v", err)
		t.Logf("=== END PRE-FLIGHT ERROR ===")
		return
	}

	for ev := range ch {
		if ev.Type == api.EventError {
			t.Logf("=== EXACT STREAM ERROR ===")
			t.Logf("%s", ev.ErrorMessage)
			t.Logf("=== END STREAM ERROR ===")
			return
		}
	}
	t.Log("No error — message accepted")
}
