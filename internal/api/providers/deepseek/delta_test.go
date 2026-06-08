package deepseek

import (
	"fmt"
	"testing"

	"claw-code-go/internal/api"
)

func TestBuildDeltaPrompt_OnlyLatestUserMessage(t *testing.T) {
	messages := []api.Message{
		{Role: "user", Content: []api.ContentBlock{{Type: "text", Text: "first"}}},
		{Role: "assistant", Content: []api.ContentBlock{{Type: "text", Text: "response"}}},
		{Role: "user", Content: []api.ContentBlock{{Type: "text", Text: "second"}}},
		{Role: "assistant", Content: []api.ContentBlock{{Type: "text", Text: "response"}}},
		{Role: "user", Content: []api.ContentBlock{{Type: "text", Text: "third"}}},
	}

	prompt := buildDeltaPrompt(messages)

	if prompt == "" {
		t.Fatal("buildDeltaPrompt returned empty")
	}

	allText := ""
	for _, msg := range messages {
		for _, cb := range msg.Content {
			if cb.Type == "text" {
				allText += cb.Text
			}
		}
	}

	t.Logf("All messages combined text: %d chars", len(allText))
	t.Logf("Delta prompt length: %d chars", len(prompt))

	if !containsSubstr(prompt, "third") {
		t.Error("prompt should contain last user message")
	}
	if containsSubstr(prompt, "first") {
		t.Error("prompt should NOT contain first user message")
	}
	if containsSubstr(prompt, "second") {
		t.Error("prompt should NOT contain second user message")
	}
}

func TestBuildDeltaPrompt_EmptyMessages(t *testing.T) {
	prompt := buildDeltaPrompt(nil)
	if prompt != "" {
		t.Errorf("expected empty prompt for nil messages, got %q", prompt)
	}

	prompt = buildDeltaPrompt([]api.Message{})
	if prompt != "" {
		t.Errorf("expected empty prompt for empty messages, got %q", prompt)
	}
}

func TestBuildDeltaPrompt_OnlyLastUserMessage(t *testing.T) {
	messages := []api.Message{
		{Role: "user", Content: []api.ContentBlock{{Type: "text", Text: "first user msg"}}},
		{Role: "assistant", Content: []api.ContentBlock{{Type: "text", Text: "assistant response"}}},
		{Role: "user", Content: []api.ContentBlock{{Type: "text", Text: "last user msg"}}},
	}

	prompt := buildDeltaPrompt(messages)

	if prompt == "" {
		t.Fatal("buildDeltaPrompt returned empty")
	}

	if !containsSubstr(prompt, "last user msg") {
		t.Errorf("prompt should contain last user message, got: %s", prompt)
	}
	if containsSubstr(prompt, "first user msg") {
		t.Errorf("prompt should NOT contain first user message, got: %s", prompt)
	}
	if containsSubstr(prompt, "assistant response") {
		t.Errorf("prompt should NOT contain assistant response, got: %s", prompt)
	}
}

func TestBuildDeltaPrompt_WithToolResults(t *testing.T) {
	messages := []api.Message{
		{Role: "user", Content: []api.ContentBlock{
			{Type: "text", Text: "user question"},
			{Type: "tool_result", ToolUseID: "tool1", Content: []api.ContentBlock{
				{Type: "text", Text: "tool output"},
			}},
		}},
	}

	prompt := buildDeltaPrompt(messages)

	if prompt == "" {
		t.Fatal("buildDeltaPrompt returned empty")
	}
	if !containsSubstr(prompt, "user question") {
		t.Errorf("prompt should contain user text, got: %s", prompt)
	}
	if !containsSubstr(prompt, "tool output") {
		t.Errorf("prompt should contain tool result, got: %s", prompt)
	}
}

func TestIsLengthLimitError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"content is too long", fmt.Errorf("stream error: Content is too long. Please shorten it and try again."), true},
		{"content is too long lowercase", fmt.Errorf("content is too long"), true},
		{"length limit reached", fmt.Errorf("length limit reached"), true},
		{"length limit", fmt.Errorf("some length limit error"), true},
		{"unrelated error", fmt.Errorf("rate limit reached"), false},
		{"empty error", fmt.Errorf(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isLengthLimitError(tt.err)
			if got != tt.want {
				t.Errorf("isLengthLimitError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestIsRateLimitError_DoesNotMatchLengthError(t *testing.T) {
	errMsg := "stream error: Content is too long. Please shorten it and try again."
	if IsRateLimitError(errMsg) {
		t.Errorf("IsRateLimitError should NOT match length limit error: %s", errMsg)
	}
}

func totalTextChars(messages []api.Message) int {
	total := 0
	for _, msg := range messages {
		for _, cb := range msg.Content {
			if cb.Type == "text" {
				total += len(cb.Text)
			}
		}
	}
	return total
}

func containsSubstr(s, substr string) bool {
	return len(s) >= len(substr) && (substr == "" || findSubstr(s, substr))
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
