package runtime

import (
	"fmt"
	"testing"
)

func TestIsTransientError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"stream error content too long", fmt.Errorf("stream error: Content is too long. Please shorten it and try again."), false},
		{"stream error server busy", fmt.Errorf("stream error: server is busy"), true},
		{"stream error timeout", fmt.Errorf("stream error: timeout"), true},
		{"server is busy", fmt.Errorf("server is busy"), true},
		{"rate limit", fmt.Errorf("rate limit reached"), true},
		{"429", fmt.Errorf("429 Too Many Requests"), true},
		{"500", fmt.Errorf("500 Internal Server Error"), true},
		{"502", fmt.Errorf("502 Bad Gateway"), true},
		{"503", fmt.Errorf("503 Service Unavailable"), true},
		{"504", fmt.Errorf("504 Gateway Timeout"), true},
		{"connection reset", fmt.Errorf("connection reset by peer"), true},
		{"timeout", fmt.Errorf("i/o timeout"), true},
		{"broken pipe", fmt.Errorf("broken pipe"), true},
		{"context canceled", fmt.Errorf("context canceled"), false},
		{"context deadline exceeded", fmt.Errorf("context deadline exceeded"), false},
		{"prompt too large", fmt.Errorf("prompt too large"), false},
		{"maximum context length", fmt.Errorf("maximum context length"), false},
		{"content is too long", fmt.Errorf("content is too long"), false},
		{"length limit reached", fmt.Errorf("length limit reached"), false},
		{"length limit", fmt.Errorf("some length limit error"), false},
		{"unrelated error", fmt.Errorf("something went wrong"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isTransientError(tt.err)
			if got != tt.want {
				t.Errorf("isTransientError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestIsPromptTooLargeError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"content is too long", fmt.Errorf("stream error: Content is too long. Please shorten it and try again."), true},
		{"content is too long lowercase", fmt.Errorf("content is too long"), true},
		{"prompt too large", fmt.Errorf("prompt too large"), true},
		{"context length exceeded", fmt.Errorf("context length exceeded"), true},
		{"maximum context length", fmt.Errorf("maximum context length"), true},
		{"context_length_exceeded", fmt.Errorf("context_length_exceeded"), true},
		{"input is too long", fmt.Errorf("input is too long"), true},
		{"length limit reached", fmt.Errorf("length limit reached"), true},
		{"length limit", fmt.Errorf("some length limit error"), true},
		{"unrelated error", fmt.Errorf("rate limit reached"), false},
		{"empty error", fmt.Errorf(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPromptTooLargeError(tt.err)
			if got != tt.want {
				t.Errorf("isPromptTooLargeError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
