package deepseek

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"claw-code-go/internal/api"
)

// Provider implements api.Provider for DeepSeek's web chat API.
type Provider struct{}

// New returns a new DeepSeek Provider.
func New() *Provider { return &Provider{} }

func (p *Provider) Name() string               { return "deepseek" }
func (p *Provider) AuthMethod() api.AuthMethod { return api.AuthMethodAPIKey }

// NewClient creates a DeepSeek API client from the given config.
func (p *Provider) NewClient(cfg api.ProviderConfig) (api.APIClient, error) {
	token := cfg.APIKey
	if token == "" {
		// Fall back to ~/.deepseek / env var, the same sources the dpp Go
		// harness uses. This lets users run `claw-code-go --provider deepseek`
		// without any extra setup beyond dropping a token in ~/.deepseek/.
		token = LoadAuth()
	}
	if token == "" {
		return nil, fmt.Errorf("deepseek: no auth token — set DEEPSEEK_TOKEN or create ~/.deepseek/deepseek_token.txt")
	}
	model := cfg.Model
	if model == "" {
		model = DefaultModel
	}
	_ = cfg.BaseURL
	wc := NewWebClient(token)
	if cfg.BaseURL != "" {
		wc.BaseURL = cfg.BaseURL
	}
	return &Client{
		model: model,
		web:   wc,
	}, nil
}

const DefaultModel = "expert"

// Client is the api.APIClient implementation for DeepSeek's web chat API.
// It translates claw-code-go's Anthropic-shaped request (system + messages +
// tools) into the web chat session format, and re-emits the streamed chunks
// as api.StreamEvent values compatible with the conversation loop.
type Client struct {
	model string
	web   *WebClient

	mu              sync.Mutex
	chatSessionID   string
	parentMessageID string

	// Per-model settings discovered from /api/v0/client/settings?scope=model.
	// Populated lazily on the first request and reused for the lifetime of
	// the client. Falls back to the hard-coded maxPromptTokens if the fetch
	// fails (network error, expired session, etc.).
	settingsOnce sync.Once
	settingsErr  error
	settings     map[string]ModelConfig
}

// fetchSettings lazily loads and caches the per-model settings from the
// server. Called from the pre-flight check so the limit can be picked from
// the actual server-advertised cap rather than a hard-coded guess.
func (c *Client) fetchSettings() (map[string]ModelConfig, error) {
	c.settingsOnce.Do(func() {
		c.settings, c.settingsErr = c.web.FetchModelSettings()
	})
	return c.settings, c.settingsErr
}

// limitForModel returns the per-model input-token cap, or 0 if unknown. The
// model_type is the only relevant dimension (thinking/search don't change
// the input cap on the web API).
func (c *Client) limitForModel(spec ModelSpec) int {
	if settings, err := c.fetchSettings(); err == nil {
		if cfg, ok := settings[spec.ModelType]; ok && cfg.InputCharacterLimit > 0 {
			return cfg.MaxInputTokens()
		}
	}
	// Fallback: hard-coded default for the model_type.
	switch spec.ModelType {
	case "expert":
		// Expert's input_character_limit is 163,840 → 40,960 tokens at 4
		// chars/token. Round down slightly for safety margin.
		return 38_000
	case "vision":
		return 600_000
	default:
		return maxPromptTokens
	}
}

// MaxInputTokens returns the provider's approximate input token limit
// for the currently configured model.
func (c *Client) MaxInputTokens() int {
	spec := ParseModelName(c.model)
	if c.model == "" {
		spec = ParseModelName(DefaultModel)
	}
	return c.limitForModel(spec)
}

// ensureSession creates a chat session on the DeepSeek side if we don't
// have one. Reused across all turns of a single conversation.
func (c *Client) ensureSession(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.chatSessionID != "" {
		debugLog("Reusing existing session: %s", c.chatSessionID)
		return nil
	}
	debugLog("Creating new chat session")
	sessID, err := c.web.CreateChatSession()
	if err != nil {
		return fmt.Errorf("deepseek: create session: %w", err)
	}
	c.chatSessionID = sessID
	debugLog("Created session: %s", sessID)
	return nil
}

// maxRetries is how many times to retry on transient errors (network blips,
// 5xx responses, PoW failures, etc.) before giving up.
const maxRetries = 3

// transientErrors contains substrings that indicate a retryable failure.
var transientErrors = []string{
	"connection reset",
	"connection refused",
	"EOF",
	"timeout",
	"PoW",
	"500",
	"502",
	"503",
	"504",
	"temporarily unavailable",
	"stream error",
	"no session ID",
}

func isTransient(errMsg string) bool {
	low := strings.ToLower(errMsg)
	for _, hint := range transientErrors {
		if strings.Contains(low, strings.ToLower(hint)) {
			return true
		}
	}
	return false
}

// debugDeepSeek enables debug logging when DEEPSEEK_DEBUG=1
var debugDeepSeek = os.Getenv("DEEPSEEK_DEBUG") == "1"

func debugLog(format string, args ...interface{}) {
	if debugDeepSeek {
		fmt.Fprintf(os.Stderr, "[deepseek-debug] "+format+"\n", args...)
	}
}

// resetSession clears the session state, forcing a new session on next request
func (c *Client) resetSession() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.chatSessionID = ""
	c.parentMessageID = ""
	debugLog("Session reset - will create new session on next request")
}

// isSessionError checks if an error indicates session expiry/invalidity
func isSessionError(errMsg string) bool {
	low := strings.ToLower(errMsg)
	return strings.Contains(low, "session") &&
		(strings.Contains(low, "expired") ||
			strings.Contains(low, "invalid") ||
			strings.Contains(low, "not found"))
}

// StreamResponse sends a streaming request to DeepSeek and returns a channel
// of api.StreamEvent values matching the conversation loop's expectations.
func (c *Client) StreamResponse(ctx context.Context, req api.CreateMessageRequest) (<-chan api.StreamEvent, error) {
	if err := c.ensureSession(ctx); err != nil {
		return nil, err
	}

	prompt := buildPrompt(req.System, req.Messages, req.Tools)

	// Resolve the model spec for this request. cfg.Model is the friendly
	// name (e.g. "expert-thinking", "instant-search"). When the caller
	// doesn't set one we default to the Expert model — the deeper reasoning
	// variant that the web UI highlights for "complex problems". The
	// spec tells us which model_type and feature flags to send to the
	// server, and which per-model input cap to use.
	spec := ParseModelName(c.model)
	if c.model == "" {
		spec = ParseModelName(DefaultModel)
	}
	modelCap := c.limitForModel(spec)

	// Pre-flight: estimate input tokens and reject if we'd blow past the
	// model-specific cap. The conversation loop's ShouldCompact uses
	// EstimateTokens as a fallback, but the actual prompt we send is bigger
	// (we flatten the system prompt + tool descriptions + messages into a
	// single string), so check against the true prompt.
	estimatedInput := EstimateTokens(prompt)
	if modelCap > 0 && estimatedInput > modelCap {
		return nil, fmt.Errorf(
			"deepseek: prompt too large for %s (~%d tokens, limit %d). Compact the session or reduce history",
			spec.CanonicalName(), estimatedInput, modelCap,
		)
	}
	if estimatedInput > softPromptLimit {
		fmt.Fprintf(os.Stderr, "[deepseek] warning: prompt at ~%d tokens (soft limit %d) for %s\n",
			estimatedInput, softPromptLimit, spec.CanonicalName())
	}

	// Compute the output token budget. The request sets MaxTokens; we use it
	// to truncate the streamed response so the conversation loop and tool
	// detection see a coherent (cut-off) answer rather than something that
	// might have been truncated mid-thought by the upstream server.
	outputBudget := req.MaxTokens
	if outputBudget <= 0 {
		outputBudget = 4096
	}
	maxOutputChars := outputBudget * charsPerToken

	c.mu.Lock()
	parentID := c.parentMessageID
	sessID := c.chatSessionID
	c.mu.Unlock()
	debugLog("StreamResponse: session=%s parent=%s", sessID, parentID)
	var parentPtr *string
	if parentID != "" {
		parentPtr = &parentID
	}

	ch := make(chan api.StreamEvent, 64)

	go func() {
		defer close(ch)

		send := func(ev api.StreamEvent) bool {
			select {
			case ch <- ev:
				return true
			case <-ctx.Done():
				return true
			}
		}

		// Kick off the turn with a message_start, including the estimated
		// input token count so the conversation loop's compaction logic
		// (which checks CompactionState.LastInputTokens) fires on time.
		if !send(api.StreamEvent{
			Type:        api.EventMessageStart,
			InputTokens: estimatedInput,
		}) {
			return
		}

		// Track all streamed content so we can run tool-call detection on it.
		var fullText strings.Builder
		// Track output tokens for the EventMessageDelta Usage field.
		var outputChars int
		// Track the server-reported *output* token count from
		// accumulated_token_usage BATCH events (live-mapped in
		// TestLiveDumpRawSSE). This is the real number the DeepSeek
		// server saw for the response — we surface it on
		// EventMessageDelta so the conversation loop's usage tracker
		// and cost estimator get accurate data instead of the local
		// chars/4 guess.
		var serverOutputTokens int
		// Hit the budget? Stop accepting new content but keep the goroutine
		// running so we can still emit tool-call blocks and the final events.
		var budgetExceeded bool
		// sawFinishedStatus records whether the server's authoritative
		// FINISHED event arrived. When set, a trailing "FINISHED" token
		// in the model's own text is a sentinel artefact (the model
		// learned to emit it as an end-of-turn marker) and we can trim
		// it without risking a false positive on the English word
		// "FINISHED" in legitimate prose. Without the SSE signal we
		// can't distinguish the two cases and the trim stays off.
		var sawFinishedStatus bool

		// Open a text content block for the streamed response.
		if !send(api.StreamEvent{
			Type:         api.EventContentBlockStart,
			Index:        0,
			ContentBlock: api.ContentBlockInfo{Type: "text", Index: 0},
		}) {
			return
		}

		// Retry loop for transient errors. We retry the whole stream
		// because the DeepSeek web API has no concept of resumable
		// streaming — if a request fails mid-stream, we have to start over.
		var (
			newID   string
			err     error
			lastErr error
		)
		for attempt := 0; attempt < maxRetries; attempt++ {
			if attempt > 0 {
				// Exponential backoff. Rate limits get a much longer wait
				// (DeepSeek's "Server is busy" can last for minutes per key)
				// while generic transient errors get the original short backoff.
				var backoff time.Duration
				if lastErr != nil && IsRateLimitError(lastErr.Error()) {
					backoff = time.Duration(1<<attempt) * 30 * time.Second
					// Rotate the token so the next attempt uses a fresh
					// quota. The token manager has up to 3 keys (the user
					// token + the 2 hardcoded fallbacks).
					if newTok := c.web.RotateToken(); newTok != "" {
						fmt.Fprintf(os.Stderr, "[deepseek] rate-limited: rotated to new token and resetting session\n")
						c.resetSession()
					}
				} else {
					backoff = time.Duration(1<<attempt) * 200 * time.Millisecond
				}
				select {
				case <-time.After(backoff):
				case <-ctx.Done():
					return
				}
				fmt.Fprintf(os.Stderr, "[deepseek] retry %d/%d after %v\n", attempt+1, maxRetries, backoff)
				// Clear the buffer so a partial response doesn't leak into
				// the next attempt's tool detection.
				fullText.Reset()
				outputChars = 0
				serverOutputTokens = 0
				budgetExceeded = false
			}

			newID, err = c.web.ChatCompletionStream(
				CompletionOpts{
					SessionID:   sessID,
					ParentMsgID: parentPtr,
					Prompt:      prompt,
					Spec:        spec,
				},
				func(event StreamEvent) bool {
					if event.Event != "content" {
						if event.Event == "message_id" {
							c.mu.Lock()
							c.parentMessageID = event.Data
							c.mu.Unlock()
							return true
						}
						if event.Event == "usage" {
							// Server-reported output token count for this
							// message. Parsed from the BATCH update frame
							// {"p":"response","o":"BATCH","v":[{"p":
							//   "accumulated_token_usage","v":N}, ...]}.
							var n int
							fmt.Sscanf(event.Data, "%d", &n)
							if n > 0 {
								serverOutputTokens = n
							}
							return true
						}
						if event.Event == "status" && event.Data == "FINISHED" {
							// The server has signalled that the turn is
							// complete. parseSSE will not invoke this
							// handler again. We keep what we already
							// accumulated in fullText; the model
							// frequently echoes a literal "FINISHED"
							// token as a training-time end-of-turn
							// marker inside its content stream. We
							// remember that the SSE event fired so the
							// post-stream cleanup below can trim that
							// sentinel safely (without this guard we
							// would risk clipping legitimate English
							// prose that happens to end with the word
							// "FINISHED").
							sawFinishedStatus = true
							return true
						}
						// "debug" lines and any unknown status values
						// are ignored — keep reading.
						return true
					}
					// Always append to fullText — even after the budget trips —
					// so tool-call detection can still see what the model tried
					// to emit (it's better to extract a partial tool call than
					// none at all).
					fullText.WriteString(event.Data)
					outputChars += len(event.Data)
					if budgetExceeded {
						return true
					}
					if outputChars > maxOutputChars {
						// We've hit the output budget. Note it for later and
						// stop pushing more deltas — but keep the partial
						// response in fullText so tool detection still works.
						budgetExceeded = true
						fmt.Fprintf(os.Stderr, "[deepseek] output budget exhausted at %d chars; truncating stream\n", outputChars)
						return true
					}
					if !send(api.StreamEvent{
						Type:  api.EventContentBlockDelta,
						Index: 0,
						Delta: api.Delta{Type: "text_delta", Text: event.Data},
					}) {
						return true
					}
					return true
				},
			)

			if err == nil {
				break
			}
			lastErr = err
			if !isTransient(err.Error()) {
				break
			}
			fmt.Fprintf(os.Stderr, "[deepseek] transient error: %v\n", err)
		}
		if err != nil {
			// Check if this is a session error and reset for recovery
			if isSessionError(err.Error()) {
				debugLog("Session error detected, resetting session: %v", err)
				c.resetSession()
			}
			send(api.StreamEvent{
				Type:         api.EventError,
				ErrorMessage: fmt.Sprintf("deepseek: %s", err.Error()),
			})
			return
		}
		if newID != "" {
			c.mu.Lock()
			c.parentMessageID = newID
			c.mu.Unlock()
			debugLog("Parent message ID updated: %s", newID)
		} else {
			debugLog("Warning: no parent message ID received from stream")
		}

		// If we hit the output budget mid-tool-call, append a marker so the
		// model (next turn) understands the previous turn was cut short.
		text := fullText.String()
		if sawFinishedStatus {
			text = trimFinishedSentinel(text)
		}
		if budgetExceeded && !strings.HasSuffix(strings.TrimSpace(text), "…") {
			text += "\n\n[…response truncated by client to fit max_tokens…]"
		}

		calls := ExtractToolCalls(text)

		// Close the text content block.
		if !send(api.StreamEvent{Type: api.EventContentBlockStop, Index: 0}) {
			return
		}

		// Emit one tool_use block per detected call. The conversation loop
		// recognises EventContentBlockStart with type=tool_use and the
		// following input_json_delta as a tool invocation.
		for i, tc := range calls {
			idx := i + 1
			argsJSON, _ := json.Marshal(tc.Arguments)
			if !send(api.StreamEvent{
				Type:  api.EventContentBlockStart,
				Index: idx,
				ContentBlock: api.ContentBlockInfo{
					Type:  "tool_use",
					Index: idx,
					ID:    fmt.Sprintf("toolu_%s_%d", newID, i),
					Name:  tc.Name,
				},
			}) {
				return
			}
			if !send(api.StreamEvent{
				Type:  api.EventContentBlockDelta,
				Index: idx,
				Delta: api.Delta{Type: "input_json_delta", PartialJSON: string(argsJSON)},
			}) {
				return
			}
			if !send(api.StreamEvent{Type: api.EventContentBlockStop, Index: idx}) {
				return
			}
		}

		stopReason := "end_turn"
		if len(calls) > 0 {
			stopReason = "tool_use"
		}
		// Surface the real server-reported output token count (or our
		// local chars/4 estimate as a fallback) on the message_delta
		// event. The conversation loop reads
		// event.Usage.OutputTokens to update the usage tracker and
		// the per-turn cost estimate.
		outTokens := serverOutputTokens
		if outTokens == 0 {
			outTokens = EstimateTokens(fullText.String())
		}
		send(api.StreamEvent{
			Type:       api.EventMessageDelta,
			StopReason: stopReason,
			Usage:      api.UsageDelta{OutputTokens: outTokens},
		})
		send(api.StreamEvent{Type: api.EventMessageStop})
	}()

	return ch, nil
}

// trimFinishedSentinel removes a trailing "FINISHED" token that the
// model emits as a learned end-of-turn marker. It only runs when
// the caller has already confirmed the SSE FINISHED event fired
// (otherwise we can't tell the sentinel from the English word).
//
// In practice, on chat.deepseek.com the model splices the sentinel
// directly onto the last word of its response ("...layers.FINISHED")
// or emits it on its own line ("...\n\nFINISHED"). The English
// word "FINISHED" at the end of a code-session turn is rare enough
// that we accept the false positive in exchange for the UX win of
// a clean response. Callers that need to distinguish the two
// should check the SSE FINISHED signal themselves (this function
// assumes it already fired).
//
// When the sentinel follows a sentence-boundary punctuation
// (`.`, `!`, `?`) we also strip that punctuation, because the model
// emits it only to fake a "sentence end" before the sentinel —
// a real English sentence ending in "FINISHED" would not have a
// period after the word ("the build is FINISHED" without period is
// already a valid stylization).
func trimFinishedSentinel(text string) string {
	const sentinel = "FINISHED"
	trimmed := strings.TrimRight(text, " \t\r\n")
	if !strings.HasSuffix(trimmed, sentinel) {
		return text
	}
	cut := len(trimmed) - len(sentinel)
	// If the char immediately before "FINISHED" is a
	// sentence-boundary punctuation the model added to fake a
	// sentence end, also strip it ("...layers.FINISHED" →
	// "...layers", not "...layers.").
	if cut > 0 {
		switch trimmed[cut-1] {
		case '.', '!', '?':
			cut--
		}
	}
	return strings.TrimRight(trimmed[:cut], " \t\r\n")
}

// buildPrompt flattens an Anthropic-style request into the single-prompt
// shape that chat.deepseek.com's /completion endpoint expects.
func buildPrompt(system string, messages []api.Message, tools []api.Tool) string {
	var parts []string

	toolDesc := describeTools(tools)

	if system != "" {
		parts = append(parts, system)
	}
	if toolDesc != "" {
		parts = append(parts, toolDesc)
	}

	for _, msg := range messages {
		switch msg.Role {
		case "user":
			var textParts []string
			var toolResults []string
			for _, block := range msg.Content {
				switch block.Type {
				case "text":
					if block.Text != "" {
						textParts = append(textParts, block.Text)
					}
				case "tool_result":
					content := extractToolResultText(block)
					toolResults = append(toolResults, fmt.Sprintf("[Tool:%s result]\n%s", block.ToolUseID, content))
				}
			}
			if len(textParts) > 0 {
				parts = append(parts, "[User]\n"+strings.Join(textParts, "\n"))
			}
			if len(toolResults) > 0 {
				parts = append(parts, strings.Join(toolResults, "\n\n"))
			}
		case "assistant":
			var textParts []string
			for _, block := range msg.Content {
				switch block.Type {
				case "text":
					if block.Text != "" {
						textParts = append(textParts, block.Text)
					}
				case "tool_use":
					argsJSON, _ := json.Marshal(block.Input)
					textParts = append(textParts, fmt.Sprintf("[Assistant tool call: %s(%s)]", block.Name, string(argsJSON)))
				}
			}
			if len(textParts) > 0 {
				parts = append(parts, "[Assistant]\n"+strings.Join(textParts, "\n"))
			}
		}
	}

	parts = append(parts, "\nRespond with the next assistant turn. To call a tool, output a JSON object like {\"tool_calls\":[{\"name\":\"...\",\"arguments\":{...}}]} on its own line, or use XML: <tool_calls><invoke name=\"tool_name\"><parameter name=\"arg\">value</parameter></invoke></tool_calls>.")

	return strings.Join(parts, "\n\n")
}

func extractToolResultText(block api.ContentBlock) string {
	var parts []string
	for _, c := range block.Content {
		if c.Text != "" {
			parts = append(parts, c.Text)
		}
	}
	return strings.Join(parts, "\n")
}

func describeTools(tools []api.Tool) string {
	if len(tools) == 0 {
		return ""
	}
	var lines []string
	lines = append(lines, "Available tools. To call a tool, output XML in this EXACT format (nothing else works):")
	lines = append(lines, "<tool_calls><invoke name=\"TOOL_NAME\"><parameter name=\"ARG_NAME\">ARG_VALUE</parameter></invoke></tool_calls>):")
	for _, t := range tools {
		lines = append(lines, fmt.Sprintf("- %s: %s", t.Name, t.Description))
		required := map[string]bool{}
		for _, r := range t.InputSchema.Required {
			required[r] = true
		}
		for k, p := range t.InputSchema.Properties {
			req := ""
			if required[k] {
				req = " (required)"
			}
			lines = append(lines, fmt.Sprintf("    - %s [%s]%s: %s", k, p.Type, req, p.Description))
		}
	}
	return strings.Join(lines, "\n")
}

// SaveToken persists a token to ~/.deepseek/deepseek_token.txt for
// subsequent runs. Called by /login flow.
func SaveToken(token string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".deepseek")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "deepseek_token.txt"), []byte(strings.TrimSpace(token)+"\n"), 0o600)
}
