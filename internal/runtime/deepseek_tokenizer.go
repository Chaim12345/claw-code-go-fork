package runtime

/*
#cgo LDFLAGS: ${SRCDIR}/libtokenizers.a -lstdc++ -lpthread -ldl -lm
*/
import "C"

import (
	_ "embed"
	"strings"
	"sync"

	"claw-code-go/internal/api"
	"github.com/cohere-ai/tokenizers"
)

//go:embed tokenizer.json
var tokenizerJSON []byte

var (
	ctk     *tokenizers.Tokenizer
	ctkOnce sync.Once
	ctkErr  error
)

// initTokenizer loads the embedded tokenizer.json through cohere-ai/tokenizers
// (Rust-based HuggingFace Tokenizers bindings). Safe for concurrent use;
// the first caller triggers a one-time synchronous load.
func initTokenizer() {
	ctkOnce.Do(func() {
		if len(tokenizerJSON) == 0 {
			ctkErr = &tokenizerLoadError{"tokenizer.json data is empty"}
			return
		}
		ctk, ctkErr = tokenizers.FromBytes(tokenizerJSON)
	})
}

type tokenizerLoadError struct{ msg string }

func (e *tokenizerLoadError) Error() string { return e.msg }

// CountTokens returns the exact number of BPE tokens in text using the
// DeepSeek tokenizer. Falls back to a chars/4 estimate if the tokenizer
// is not yet loaded. Safe for concurrent use.
func CountTokens(text string) int {
	initTokenizer()
	if ctkErr != nil || ctk == nil {
		return fallbackTokens(text)
	}
	ids, _ := ctk.Encode(text, false) // addSpecialTokens=false
	if len(ids) == 0 && text != "" {
		return fallbackTokens(text)
	}
	return len(ids)
}

// CountMessagesTokens returns the token count for a slice of messages
// by encoding their textual content through the tokenizer. Falls back
// to the chars/4 heuristic if the tokenizer is not yet available.
func CountMessagesTokens(messages []api.Message) int {
	initTokenizer()
	if ctkErr != nil || ctk == nil {
		return fallbackMessagesTokens(messages)
	}
	total := 0
	for _, msg := range messages {
		total += countOneMessage(ctk, msg)
	}
	return total
}

func fallbackTokens(text string) int {
	if len(text) == 0 {
		return 0
	}
	return (len(text) + 3) / 4
}

func countOneMessage(tk *tokenizers.Tokenizer, msg api.Message) int {
	total := 0
	for _, cb := range msg.Content {
		switch cb.Type {
		case "text":
			ids, _ := tk.Encode(cb.Text, false)
			if len(ids) > 0 {
				total += len(ids)
			} else {
				total += fallbackTokens(cb.Text)
			}
		case "tool_use":
			if cb.Input != nil {
				serialized := serializeInput(cb.Input)
				ids, _ := tk.Encode(serialized, false)
				if len(ids) > 0 {
					total += len(ids)
				} else {
					total += fallbackTokens(serialized)
				}
			}
		}
		for _, inner := range cb.Content {
			ids, _ := tk.Encode(inner.Text, false)
			if len(ids) > 0 {
				total += len(ids)
			} else {
				total += fallbackTokens(inner.Text)
			}
		}
	}
	return total
}

func serializeInput(input map[string]any) string {
	var sb strings.Builder
	for k, v := range input {
		sb.WriteString(k)
		if s, ok := v.(string); ok {
			sb.WriteString(s)
		}
	}
	return sb.String()
}

func fallbackMessagesTokens(messages []api.Message) int {
	var total int
	for _, msg := range messages {
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
	}
	return total
}

// TokenizerAvailable reports whether the DeepSeek tokenizer has been
// successfully loaded and is ready for accurate token counting.
func TokenizerAvailable() bool {
	initTokenizer()
	return ctkErr == nil && ctk != nil
}
