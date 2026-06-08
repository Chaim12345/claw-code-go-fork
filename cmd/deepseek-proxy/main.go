// Package main implements a fully OpenAI-compatible HTTP proxy for the
// claw-code-go DeepSeek provider (chat.deepseek.com web API).
//
// Endpoints:
//   POST /v1/chat/completions — OpenAI Chat Completions (streaming + non-streaming)
//   GET  /v1/models           — model listing
//   GET  /healthz             — liveness check
//
// Env:
//   DEEPSEEK_API_KEY     optional browser-token for chat.deepseek.com
//   DEEPSEEK_MODEL       default model slug (default: "expert")
//   LISTEN_ADDR          listen address (default: ":5001")
//   DEEPSEEK_PROXY_TOKEN optional Bearer token clients must present (zero-config when unset)
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	deepseekprovider "claw-code-go/internal/api/providers/deepseek"
	"claw-code-go/internal/api"
)

// ---- OpenAI-compatible request types ----

// chatMessage mirrors the OpenAI ChatCompletionRequestMessage shape.
type chatMessage struct {
	Role       string            `json:"role"`
	Content    json.RawMessage   `json:"content"`             // string or array of content parts
	Name       string            `json:"name,omitempty"`
	ToolCalls  []openAIToolCall  `json:"tool_calls,omitempty"`
	ToolCallID string            `json:"tool_call_id,omitempty"`
	Refusal    string            `json:"refusal,omitempty"`
}

type openAIToolCall struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Function openAIFunction `json:"function"`
}

type openAIFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// chatRequest mirrors the full OpenAI /v1/chat/completions request shape.
// All optional fields use pointers so we can distinguish "not set" from zero.
type chatRequest struct {
	Model            string           `json:"model"`
	Messages         []chatMessage    `json:"messages"`
	MaxTokens        *int             `json:"max_tokens,omitempty"`
	MaxCompletionTokens *int          `json:"max_completion_tokens,omitempty"`
	Temperature      *float64         `json:"temperature,omitempty"`
	TopP             *float64         `json:"top_p,omitempty"`
	N                *int             `json:"n,omitempty"`
	Stream           *bool            `json:"stream,omitempty"`
	StreamOptions    *streamOptions   `json:"stream_options,omitempty"`
	Stop             json.RawMessage  `json:"stop,omitempty"`        // string or []string
	PresencePenalty  *float64         `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64         `json:"frequency_penalty,omitempty"`
	LogitBias        map[string]int   `json:"logit_bias,omitempty"`
	User             string           `json:"user,omitempty"`
	Seed             *int             `json:"seed,omitempty"`
	Tools            []openAITool     `json:"tools,omitempty"`
	ToolChoice       json.RawMessage  `json:"tool_choice,omitempty"` // string or object
	ParallelToolCalls *bool           `json:"parallel_tool_calls,omitempty"`
	ResponseFormat   json.RawMessage  `json:"response_format,omitempty"`
	Logprobs         *bool            `json:"logprobs,omitempty"`
	TopLogprobs      *int             `json:"top_logprobs,omitempty"`
}

type streamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type openAITool struct {
	Type     string            `json:"type"`
	Function openAIFunctionDef `json:"function"`
}

type openAIFunctionDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

// ---- plumbing ----

var (
	listenAddr   string
	apiKey       string
	defaultModel string
	systemFingerprint = "claw-code-go-deepseek-proxy/v0"
	proxyToken   string // Bearer token clients must present; "" = no auth required (zero-token mode)
)

func init() {
	listenAddr = os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = ":5001"
	}
	apiKey = strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY"))
	defaultModel = os.Getenv("DEEPSEEK_MODEL")
	if defaultModel == "" {
		defaultModel = "expert"
	}
	// Zero-token mode: if DEEPSEEK_PROXY_TOKEN is not set, no client auth is required.
	// This makes the proxy a "universal" web API that any client can hit without setup.
	proxyToken = strings.TrimSpace(os.Getenv("DEEPSEEK_PROXY_TOKEN"))
	if proxyToken == "" {
		fmt.Fprintf(os.Stderr, "[proxy] zero-token mode — no client authentication required (set DEEPSEEK_PROXY_TOKEN to enable)\n")
	}
}

// ---- model catalog ----

type modelEntry struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

var knownModels = []modelEntry{
	{ID: "instant", Object: "model", Created: 1700000000, OwnedBy: "deepseek"},
	{ID: "instant-thinking", Object: "model", Created: 1700000000, OwnedBy: "deepseek"},
	{ID: "instant-search", Object: "model", Created: 1700000000, OwnedBy: "deepseek"},
	{ID: "instant-thinking-search", Object: "model", Created: 1700000000, OwnedBy: "deepseek"},
	{ID: "expert", Object: "model", Created: 1700000000, OwnedBy: "deepseek"},
	{ID: "expert-thinking", Object: "model", Created: 1700000000, OwnedBy: "deepseek"},
	{ID: "vision", Object: "model", Created: 1700000000, OwnedBy: "deepseek"},
	{ID: "vision-thinking", Object: "model", Created: 1700000000, OwnedBy: "deepseek"},
	{ID: "deepseek-chat", Object: "model", Created: 1700000000, OwnedBy: "deepseek"},
	{ID: "deepseek-reasoner", Object: "model", Created: 1700000000, OwnedBy: "deepseek"},
}

func init() {
	// Sort models by ID for deterministic output.
	sort.Slice(knownModels, func(i, j int) bool { return knownModels[i].ID < knownModels[j].ID })
}

// ---- main ----

func main() {
	provider := deepseekprovider.New()
	cfg := api.ProviderConfig{
		APIKey: apiKey,
		Model:  defaultModel,
	}
	client, err := provider.NewClient(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: failed to build DeepSeek client: %v\n", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		handleChatCompletions(w, r, client)
	})
	mux.HandleFunc("/v1/models", func(w http.ResponseWriter, r *http.Request) {
		handleModels(w, r)
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	handler := requestLog(corsMiddleware(mux))
	srv := &http.Server{Addr: listenAddr, Handler: handler}
	fmt.Fprintf(os.Stderr, "[proxy] listening on %s  (model=%s)\n", listenAddr, defaultModel)
	if err := srv.ListenAndServe(); err != nil {
		fmt.Fprintf(os.Stderr, "[proxy] shutdown: %v\n", err)
	}
}

// ---- middleware ----

func requestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(os.Stderr, "[proxy] %s %s\n", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, x-request-id")
		w.Header().Set("Access-Control-Expose-Headers", "x-request-id")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// bearerAuth checks the Authorization header against proxyToken.
// Returns true if access is granted (zero-token mode or valid token).
func bearerAuth(r *http.Request) bool {
	if proxyToken == "" {
		return true // zero-token mode
	}
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return false
	}
	return strings.TrimPrefix(auth, "Bearer ") == proxyToken
}

// ---- /v1/models ----

func handleModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeOpenAIError(w, r.URL.Path, "method_not_allowed", "Only GET is supported", "", http.StatusMethodNotAllowed)
		return
	}
	if !bearerAuth(r) {
		writeOpenAIError(w, r.URL.Path, "authentication_error", "Invalid or missing Authorization header", "", http.StatusUnauthorized)
		return
	}

	type listResponse struct {
		Object string        `json:"object"`
		Data   []modelEntry  `json:"data"`
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(listResponse{
		Object: "list",
		Data:   knownModels,
	})
}

// ---- /v1/chat/completions ----

func handleChatCompletions(w http.ResponseWriter, r *http.Request, client api.APIClient) {
	if r.Method != http.MethodPost {
		writeOpenAIError(w, r.URL.Path, "method_not_allowed", "Only POST is supported", "", http.StatusMethodNotAllowed)
		return
	}
	if !bearerAuth(r) {
		writeOpenAIError(w, r.URL.Path, "authentication_error", "Invalid or missing Authorization header", "", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeOpenAIError(w, r.URL.Path, "invalid_request_error", err.Error(), "", http.StatusBadRequest)
		return
	}
	r.Body.Close()

	var req chatRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeOpenAIError(w, r.URL.Path, "invalid_request_error", fmt.Sprintf("bad request: %v", err), "", http.StatusBadRequest)
		return
	}

	// Resolve max_tokens: prefer max_completion_tokens over max_tokens (OpenAI v2 naming).
	maxTokens := 4096
	if req.MaxCompletionTokens != nil && *req.MaxCompletionTokens > 0 {
		maxTokens = *req.MaxCompletionTokens
	} else if req.MaxTokens != nil && *req.MaxTokens > 0 {
		maxTokens = *req.MaxTokens
	}

	model := req.Model
	if model == "" {
		model = defaultModel
	}
	stream := req.Stream != nil && *req.Stream

	// Extract system message and convert messages.
	system, messages := convertMessages(req.Messages)

	// Convert tools.
	var tools []api.Tool
	for _, t := range req.Tools {
		var props map[string]interface{}
		_ = json.Unmarshal(t.Function.Parameters, &props)
		tools = append(tools, api.Tool{
			Name:        t.Function.Name,
			Description: t.Function.Description,
			InputSchema: api.InputSchema{
				Type:       "object",
				Properties: toPropertyMap(props),
				Required:   requiredFields(props),
			},
		})
	}

	internalReq := api.CreateMessageRequest{
		Model:     model,
		MaxTokens: maxTokens,
		System:    system,
		Messages:  messages,
		Tools:     tools,
		Stream:    stream,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	ch, err := client.StreamResponse(ctx, internalReq)
	if err != nil {
		writeOpenAIError(w, r.URL.Path, "upstream_error", fmt.Sprintf("deepseek upstream error: %v", err), "", http.StatusBadGateway)
		return
	}

	if stream {
		handleStreamingResponse(w, ch, model)
	} else {
		handleNonStreamingResponse(w, ch, model)
	}
}

// ---- streaming response ----

func handleStreamingResponse(w http.ResponseWriter, ch <-chan api.StreamEvent, model string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	flusher, ok := w.(http.Flusher)
	if !ok {
		fmt.Fprintf(os.Stderr, "[proxy] streaming unsupported by response writer\n")
		return
	}

	msgID := fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano())
	created := time.Now().Unix()
	var finishReason string
	var outputTokens, inputTokens int
	var pendingDelta string
	var allText strings.Builder
	var inToolCallBlock bool
	var hasNativeToolCalls bool

	flushPending := func(trim bool) {
		if pendingDelta == "" {
			return
		}
		text := pendingDelta
		pendingDelta = ""
		if trim {
			text = trimFinishedSentinel(text)
		}
		if text == "" {
			return
		}
		allText.WriteString(text)
		if !inToolCallBlock && looksLikeToolCallStart(allText.String()) {
			inToolCallBlock = true
		}
		if !inToolCallBlock {
			writeSSE(w, streamingChunk(msgID, model, created, map[string]interface{}{"content": text}, nil))
			flusher.Flush()
		}
	}

	for ev := range ch {
		switch ev.Type {
		case api.EventMessageStart:
			flushPending(false)
			inputTokens = ev.InputTokens
			writeSSE(w, streamingChunk(msgID, model, created, map[string]interface{}{"role": "assistant"}, nil))
			flusher.Flush()

		case api.EventContentBlockStart:
			flushPending(false)
			if ev.ContentBlock.Type == "tool_use" {
				hasNativeToolCalls = true
				tc := map[string]interface{}{
					"index": 0,
					"id":    fmt.Sprintf("call_%s_%d", msgID, ev.ContentBlock.Index),
					"type":  "function",
					"function": map[string]interface{}{
						"name":      ev.ContentBlock.Name,
						"arguments": "",
					},
				}
				writeSSE(w, streamingChunk(msgID, model, created, nil, tc))
				flusher.Flush()
			}

		case api.EventContentBlockDelta:
			if ev.Delta.Type == "input_json_delta" {
				flushPending(false)
				tc := map[string]interface{}{
					"index": 0,
					"function": map[string]interface{}{
						"arguments": ev.Delta.PartialJSON,
					},
				}
				writeSSE(w, streamingChunk(msgID, model, created, nil, tc))
				flusher.Flush()
				continue
			}
			if ev.Delta.Type == "text_delta" && ev.Delta.Text != "" {
				unescaped := html.UnescapeString(ev.Delta.Text)
				if strings.TrimSpace(unescaped) == "FINISHED" {
					flushPending(true)
					continue
				}
				flushPending(false)
				pendingDelta = unescaped
			}

		case api.EventMessageDelta:
			flushPending(true)
			if ev.StopReason != "" {
				finishReason = ev.StopReason
			}
			if ev.Usage.OutputTokens > 0 {
				outputTokens = ev.Usage.OutputTokens
			}

		case api.EventError:
			flushPending(false)
			fmt.Fprintf(os.Stderr, "[proxy] upstream error: %s\n", ev.ErrorMessage)
		}
	}

	// Flush final buffered delta.
	flushPending(true)

	extractedCalls := extractToolCallsFromText(allText.String())

	if finishReason == "" || finishReason == "end_turn" {
		finishReason = "stop"
	} else if finishReason == "tool_use" {
		finishReason = "tool_calls"
	} else if finishReason == "max_tokens" {
		finishReason = "length"
	}

	if !hasNativeToolCalls && len(extractedCalls) > 0 {
		for _, tc := range extractedCalls {
			writeSSE(w, streamingChunk(msgID, model, created, nil, map[string]interface{}{
				"index": 0,
				"id":    tc.ID,
				"type":  "function",
				"function": map[string]interface{}{
					"name":      tc.Function.Name,
					"arguments": tc.Function.Arguments,
				},
			}))
		}
		flusher.Flush()
		finishReason = "tool_calls"
	}

	// Final chunk with finish_reason + usage.
	writeSSE(w, streamingChunkWithUsage(msgID, model, created, finishReason, inputTokens, outputTokens))
	flusher.Flush()

	// [DONE] marker.
	writeSSE(w, nil)
	flusher.Flush()
}

func streamingChunk(id, model string, created int64, content map[string]interface{}, toolCall map[string]interface{}) map[string]interface{} {
	chunk := map[string]interface{}{
		"id":                id,
		"object":            "chat.completion.chunk",
		"created":           created,
		"model":             model,
		"system_fingerprint": systemFingerprint,
	}
	delta := map[string]interface{}{}
	if content != nil {
		for k, v := range content {
			delta[k] = v
		}
	}
	choice := map[string]interface{}{
		"index":         0,
		"delta":         delta,
		"logprobs":      nil,
		"finish_reason": nil,
	}
	if toolCall != nil {
		// For tool_calls, nest under delta.tool_calls as an array.
		delta["tool_calls"] = []interface{}{toolCall}
	}
	chunk["choices"] = []interface{}{choice}
	return chunk
}

func streamingChunkWithUsage(id, model string, created int64, finish string, inTok, outTok int) map[string]interface{} {
	chunk := map[string]interface{}{
		"id":                id,
		"object":            "chat.completion.chunk",
		"created":           created,
		"model":             model,
		"system_fingerprint": systemFingerprint,
		"choices": []interface{}{
			map[string]interface{}{
				"index":         0,
				"delta":         map[string]interface{}{},
				"logprobs":      nil,
				"finish_reason": finish,
			},
		},
		"usage": map[string]interface{}{
			"prompt_tokens":     inTok,
			"completion_tokens": outTok,
			"total_tokens":      inTok + outTok,
		},
	}
	return chunk
}

// ---- non-streaming response ----

func handleNonStreamingResponse(w http.ResponseWriter, ch <-chan api.StreamEvent, model string) {
	msgID := fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano())
	created := time.Now().Unix()
	var fullText strings.Builder
	var toolCalls []api.ToolCall
	var finishReason string
	var outputTokens, inputTokens int
	var lastToolCallIdx int

	for ev := range ch {
		switch ev.Type {
		case api.EventMessageStart:
			inputTokens = ev.InputTokens

		case api.EventContentBlockStart:
			if ev.ContentBlock.Type == "tool_use" {
				lastToolCallIdx = len(toolCalls)
				toolCalls = append(toolCalls, api.ToolCall{
					Name:      ev.ContentBlock.Name,
					Arguments: make(map[string]interface{}),
				})
			}

		case api.EventContentBlockDelta:
			if ev.Delta.Type == "input_json_delta" && len(toolCalls) > 0 {
				_ = json.Unmarshal([]byte(ev.Delta.PartialJSON), &toolCalls[lastToolCallIdx].Arguments)
				continue
			}
			if ev.Delta.Type == "text_delta" {
				unescaped := html.UnescapeString(ev.Delta.Text)
				if strings.TrimSpace(unescaped) == "FINISHED" {
					continue
				}
				fullText.WriteString(unescaped)
			}

		case api.EventMessageDelta:
			if ev.StopReason != "" {
				finishReason = ev.StopReason
			}
			if ev.Usage.OutputTokens > 0 {
				outputTokens = ev.Usage.OutputTokens
			}

		case api.EventError:
			fmt.Fprintf(os.Stderr, "[proxy] upstream error: %s\n", ev.ErrorMessage)
		}
	}

	// Trim FINISHED sentinel from the end of the full text.
	text := fullText.String()
	text = trimFinishedSentinel(text)

	textToolCalls := extractToolCallsFromText(fullText.String())

	if len(textToolCalls) > 0 {
		finishReason = "tool_calls"
	} else if finishReason == "" || finishReason == "end_turn" {
		finishReason = "stop"
	} else if finishReason == "tool_use" {
		finishReason = "tool_calls"
	}

	message := map[string]interface{}{
		"role": "assistant",
	}
	if text != "" && len(textToolCalls) == 0 {
		message["content"] = text
	} else {
		message["content"] = nil
	}

	if len(textToolCalls) > 0 {
		var openAIToolCalls []map[string]interface{}
		for _, tc := range textToolCalls {
			openAIToolCalls = append(openAIToolCalls, map[string]interface{}{
				"id":   tc.ID,
				"type": "function",
				"function": map[string]interface{}{
					"name":      tc.Function.Name,
					"arguments": tc.Function.Arguments,
				},
			})
		}
		message["tool_calls"] = openAIToolCalls
	} else if len(toolCalls) > 0 {
		var openAIToolCalls []map[string]interface{}
		for _, tc := range toolCalls {
			argsJSON, _ := json.Marshal(tc.Arguments)
			openAIToolCalls = append(openAIToolCalls, map[string]interface{}{
				"id":   fmt.Sprintf("call_%s_%d", msgID, len(openAIToolCalls)),
				"type": "function",
				"function": map[string]interface{}{
					"name":      tc.Name,
					"arguments": string(argsJSON),
				},
			})
		}
		message["tool_calls"] = openAIToolCalls
	}

	resp := map[string]interface{}{
		"id":                msgID,
		"object":            "chat.completion",
		"created":           created,
		"model":             model,
		"system_fingerprint": systemFingerprint,
		"choices": []map[string]interface{}{
			{
				"index":         0,
				"message":       message,
				"logprobs":      nil,
				"finish_reason": finishReason,
			},
		},
		"usage": map[string]interface{}{
			"prompt_tokens":     inputTokens,
			"completion_tokens": outputTokens,
			"total_tokens":      inputTokens + outputTokens,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// ---- message conversion ----

// extractTextContent extracts the text from an OpenAI message content field,
// which can be either a plain string or an array of content parts.
func extractTextContent(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	// Try string first.
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	// Try array of content parts.
	var parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(raw, &parts) == nil {
		var texts []string
		for _, p := range parts {
			if p.Type == "text" && p.Text != "" {
				texts = append(texts, p.Text)
			}
		}
		return strings.Join(texts, "\n")
	}
	// Fallback: return raw as string.
	return string(raw)
}

func convertMessages(openai []chatMessage) (string, []api.Message) {
	var systemBuf strings.Builder
	msgs := make([]api.Message, 0, len(openai))

	for _, m := range openai {
		switch m.Role {
		case "system", "developer":
			text := extractTextContent(m.Content)
			if systemBuf.Len() > 0 {
				systemBuf.WriteString("\n\n")
			}
			systemBuf.WriteString(text)

		case "user":
			text := extractTextContent(m.Content)
			if m.ToolCallID != "" {
				// This is a tool result message (role="user" with tool_call_id).
				msgs = append(msgs, api.Message{
					Role: "user",
					Content: []api.ContentBlock{{
						Type:      "tool_result",
						ToolUseID: m.ToolCallID,
						Content:   []api.ContentBlock{{Type: "text", Text: text}},
					}},
				})
			} else if m.Name != "" {
				msgs = append(msgs, api.Message{
					Role:    "user",
					Content: []api.ContentBlock{{Type: "text", Text: fmt.Sprintf("[%s] %s", m.Name, text)}},
				})
			} else {
				msgs = append(msgs, api.Message{
					Role:    "user",
					Content: []api.ContentBlock{{Type: "text", Text: text}},
				})
			}

		case "assistant":
			if len(m.ToolCalls) > 0 {
				for _, tc := range m.ToolCalls {
					args := make(map[string]interface{})
					_ = json.Unmarshal([]byte(tc.Function.Arguments), &args)
					msgs = append(msgs, api.Message{
						Role: "assistant",
						Content: []api.ContentBlock{{
							Type:  "tool_use",
							ID:    tc.ID,
							Name:  tc.Function.Name,
							Input: args,
						}},
					})
				}
			} else if m.Content != nil {
				text := extractTextContent(m.Content)
				if text != "" || m.Refusal != "" {
					display := text
					if m.Refusal != "" {
						display = m.Refusal
					}
					msgs = append(msgs, api.Message{
						Role:    "assistant",
						Content: []api.ContentBlock{{Type: "text", Text: display}},
					})
				}
			}

		case "tool":
			text := extractTextContent(m.Content)
			msgs = append(msgs, api.Message{
				Role: "user",
				Content: []api.ContentBlock{{
					Type:      "tool_result",
					ToolUseID: m.ToolCallID,
					Content:   []api.ContentBlock{{Type: "text", Text: text}},
				}},
			})

		case "function":
			// Deprecated OpenAI function role — treat as tool.
			text := extractTextContent(m.Content)
			msgs = append(msgs, api.Message{
				Role: "user",
				Content: []api.ContentBlock{{
					Type:      "tool_result",
					ToolUseID: m.Name, // function role used name as the call identifier
					Content:   []api.ContentBlock{{Type: "text", Text: text}},
				}},
			})
		}
	}
	return systemBuf.String(), msgs
}

func toPropertyMap(raw map[string]interface{}) map[string]api.Property {
	out := make(map[string]api.Property, len(raw))
	for k, v := range raw {
		m, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		p := api.Property{}
		if d, ok := m["description"].(string); ok {
			p.Description = d
		}
		if t, ok := m["type"].(string); ok {
			p.Type = t
		}
		out[k] = p
	}
	return out
}

func requiredFields(raw map[string]interface{}) []string {
	v, ok := raw["required"]
	if !ok {
		return nil
	}
	arr, ok := v.([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, r := range arr {
		if s, ok := r.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// trimFinishedSentinel removes a trailing "FINISHED" token that
// DeepSeek's web API model emits as a learned end-of-turn marker.
func trimFinishedSentinel(text string) string {
	const sentinel = "FINISHED"
	trimmed := strings.TrimRight(text, " \t\r\n")
	if !strings.HasSuffix(trimmed, sentinel) {
		return text
	}
	cut := len(trimmed) - len(sentinel)
	if cut > 0 {
		switch trimmed[cut-1] {
		case '.', '!', '?':
			cut--
		}
	}
	return strings.TrimRight(trimmed[:cut], " \t\r\n")
}

func looksLikeToolCallStart(text string) bool {
	unescaped := html.UnescapeString(text)
	lower := strings.ToLower(unescaped)
	return strings.Contains(lower, "<tool_calls") ||
		strings.Contains(lower, "<|dsml|tool_calls") ||
		strings.Contains(lower, "<invoke name=") ||
		strings.Contains(lower, `{"tool_calls"`) ||
		strings.Contains(lower, `{"tool_calls`)
}

// extractToolCallsFromText parses XML or JSON tool calls from accumulated text.
func extractToolCallsFromText(text string) []openAIToolCall {
	unescaped := html.UnescapeString(text)
	calls := api.ExtractXmlToolCalls(unescaped)
	if len(calls) == 0 {
		calls = extractJSONToolCalls(unescaped)
	}
	if len(calls) == 0 {
		return nil
	}
	var result []openAIToolCall
	for i, tc := range calls {
		argsJSON, _ := json.Marshal(tc.Arguments)
		result = append(result, openAIToolCall{
			ID:   fmt.Sprintf("call_%d", time.Now().UnixNano()+int64(i)),
			Type: "function",
			Function: openAIFunction{
				Name:      tc.Name,
				Arguments: string(argsJSON),
			},
		})
	}
	return result
}

// extractJSONToolCalls parses {"tool_calls":[{"name":"...","arguments":{...}}]} format.
func extractJSONToolCalls(text string) []api.ToolCall {
	text = strings.TrimSpace(text)
	idx := strings.Index(text, `{"tool_calls"`)
	if idx < 0 {
		return nil
	}
	depth := 0
	start := -1
	for i := idx; i < len(text); i++ {
		switch text[i] {
		case '{':
			if depth == 0 {
				start = i
			}
			depth++
		case '}':
			depth--
			if depth == 0 && start >= 0 {
				var wrapper struct {
					ToolCalls []struct {
						Name      string                 `json:"name"`
						Arguments map[string]interface{} `json:"arguments"`
					} `json:"tool_calls"`
				}
				if err := json.Unmarshal([]byte(text[start:i+1]), &wrapper); err == nil {
					var result []api.ToolCall
					for _, tc := range wrapper.ToolCalls {
						result = append(result, api.ToolCall{
							Name:      tc.Name,
							Arguments: tc.Arguments,
						})
					}
					return result
				}
				return nil
			}
		}
	}
	return nil
}

// ---- SSE + error helpers ----

func writeSSE(w http.ResponseWriter, v interface{}) {
	if v == nil {
		fmt.Fprintf(w, "data: [DONE]\n\n")
		return
	}
	b, _ := json.Marshal(v)
	fmt.Fprintf(w, "data: %s\n\n", b)
}

// writeOpenAIError writes an OpenAI-compatible error response.
func writeOpenAIError(w http.ResponseWriter, path, typ, message, param string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"message": message,
			"type":    typ,
			"param":   param,
			"code":    code,
		},
	})
}