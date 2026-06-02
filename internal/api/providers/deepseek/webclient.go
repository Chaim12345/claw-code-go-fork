// Package deepseek implements the api.Provider and api.APIClient interfaces
// for DeepSeek's web chat API (chat.deepseek.com) — the same interface used
// in a browser session. This avoids needing a DeepSeek-issued API key and
// instead reuses a session token captured from a logged-in browser.
package deepseek

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"claw-code-go/internal/api"
)

const (
	defaultBaseURL = "https://chat.deepseek.com"
	appVersion     = "20241129.1"
	clientVersion  = "2.0.0"
	headerOrigin   = "https://chat.deepseek.com"
)

// Hardcoded tokens for automatic rotation on rate limits.
var hardcodedTokens = []string{
	"3evJVVhmV+zcNoOO57Xjo627uluj61CAPNsZtwclRXDAssfXPHlYbYDOnEvSq/Nt",
	"tG4oKk2pZD8y8ShOVjyhz1CttmhzF6OYaHe7BQnIOOOB3n4X7iVQSJfr1sloGsL2",
}

// TokenManager handles automatic token rotation on rate limits.
type TokenManager struct {
	mu           sync.RWMutex
	tokens       []string
	currentIndex int
}

// NewTokenManager creates a token manager with the given tokens.
func NewTokenManager(tokens []string) *TokenManager {
	return &TokenManager{
		tokens:       tokens,
		currentIndex: 0,
	}
}

// CurrentToken returns the active token.
func (tm *TokenManager) CurrentToken() string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	if len(tm.tokens) == 0 {
		return ""
	}
	return tm.tokens[tm.currentIndex]
}

// RotateToNext switches to the next token and returns it.
func (tm *TokenManager) RotateToNext() string {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if len(tm.tokens) == 0 {
		return ""
	}
	tm.currentIndex = (tm.currentIndex + 1) % len(tm.tokens)
	fmt.Fprintf(os.Stderr, "[deepseek] rotated to token %d/%d\n", tm.currentIndex+1, len(tm.tokens))
	return tm.tokens[tm.currentIndex]
}

// IsRateLimitError checks if an error is a rate limit error from DeepSeek.
func IsRateLimitError(errMsg string) bool {
	low := strings.ToLower(errMsg)
	return strings.Contains(low, "too frequent") ||
		strings.Contains(low, "rate_limit_reached") ||
		strings.Contains(low, "rate limit")
}

// LoadAuth resolves a DeepSeek auth token from one of:
//   - DEEPSEEK_TOKEN env var
//   - ~/.deepseek/deepseek_token.txt (plain text)
//   - ~/.deepseek/auth.json (Chrome localStorage export)
func LoadAuth() string {
	if tok := strings.TrimSpace(os.Getenv("DEEPSEEK_TOKEN")); tok != "" {
		return tok
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	tokenFile := filepath.Join(home, ".deepseek", "deepseek_token.txt")
	if data, err := os.ReadFile(tokenFile); err == nil {
		if t := strings.TrimSpace(string(data)); t != "" {
			return t
		}
	}

	authFile := filepath.Join(home, ".deepseek", "auth.json")
	if data, err := os.ReadFile(authFile); err == nil {
		var authObj struct {
			Origins []struct {
				LocalStorage []struct {
					Key   string `json:"key"`
					Value string `json:"value"`
				} `json:"localStorage"`
			} `json:"origins"`
		}
		if json.Unmarshal(data, &authObj) == nil {
			for _, origin := range authObj.Origins {
				for _, ls := range origin.LocalStorage {
					if ls.Key == "token" || ls.Key == "ds_token" || ls.Key == "deepseek_token" {
						return ls.Value
					}
				}
			}
		}
	}

	return ""
}

// WebClient talks to chat.deepseek.com's HTTP API directly, the same way the
// website does in a browser. It manages a chat session, the PoW challenge
// (solved in WASM), and an optional WAF cookie.
type WebClient struct {
	BaseURL      string
	Token        string
	CookieHeader string
	client       *http.Client
	solver       *WasmSolver
	solverMu     sync.Mutex
	chatSession  string
	parentMsgID  *string
	tokenMgr     *TokenManager
}

// NewWebClient constructs a client. If a solver is required (most sessions)
// and WASM init fails, the client still works but PoW challenges are skipped.
func NewWebClient(auth string) *WebClient {
	// Build token list: provided auth first, then hardcoded tokens (deduped)
	tokens := []string{}
	seen := map[string]bool{}
	if auth != "" {
		tokens = append(tokens, auth)
		seen[auth] = true
	}
	for _, t := range hardcodedTokens {
		if !seen[t] {
			tokens = append(tokens, t)
			seen[t] = true
		}
	}
	return &WebClient{
		BaseURL:  defaultBaseURL,
		Token:    auth,
		client:   &http.Client{Timeout: 120 * time.Second},
		tokenMgr: NewTokenManager(tokens),
	}
}

func (wc *WebClient) getSolver() *WasmSolver {
	wc.solverMu.Lock()
	defer wc.solverMu.Unlock()
	if wc.solver == nil {
		s, err := NewWasmSolver()
		if err != nil {
			fmt.Fprintf(os.Stderr, "[POW] WASM solver init error: %v\n", err)
			return nil
		}
		wc.solver = s
	}
	return wc.solver
}

// RotateToken switches to the next available token. Returns the new token.
func (wc *WebClient) RotateToken() string {
	newToken := wc.tokenMgr.RotateToNext()
	wc.Token = newToken
	return newToken
}

func (wc *WebClient) buildBaseHeaders() http.Header {
	h := http.Header{}
	h.Set("Accept", "application/json, text/plain, */*")
	h.Set("Content-Type", "application/json")
	h.Set("Origin", headerOrigin)
	h.Set("Referer", headerOrigin+"/")
	h.Set("x-app-version", appVersion)
	h.Set("x-client-locale", "en_US")
	h.Set("x-client-platform", "web")
	h.Set("x-client-timezone-offset", timezoneOffset())
	h.Set("x-client-version", clientVersion)
	h.Set("sec-fetch-dest", "empty")
	h.Set("sec-fetch-mode", "cors")
	h.Set("sec-fetch-site", "same-origin")
	if wc.Token != "" {
		h.Set("Authorization", "Bearer "+wc.Token)
	}
	if wc.CookieHeader != "" {
		h.Set("Cookie", wc.CookieHeader)
	}
	return h
}

// CreateChatSession opens a new chat session and returns the session ID.
func (wc *WebClient) CreateChatSession() (string, error) {
	url := wc.BaseURL + "/api/v0/chat_session/create"
	body := map[string]interface{}{"from": "sidebar"}
	payload, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("request create: %w", err)
	}
	req.Header = wc.buildBaseHeaders()

	resp, err := wc.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("create session status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data struct {
			BizData struct {
				ID          string `json:"id"`
				ChatSession struct {
					ID string `json:"id"`
				} `json:"chat_session"`
			} `json:"biz_data"`
			ChatSessionID string `json:"chat_session_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse create response: %w", err)
	}
	sessionID := result.Data.BizData.ChatSession.ID
	if sessionID == "" {
		sessionID = result.Data.BizData.ID
	}
	if sessionID == "" {
		sessionID = result.Data.ChatSessionID
	}
	if sessionID == "" {
		return "", fmt.Errorf("no session ID in response: %s", string(respBody))
	}
	return sessionID, nil
}

// fetchPowChallenge asks the server for a PoW challenge and solves it.
// Returns the base64-encoded JSON response header value, or "" if skipped.
func (wc *WebClient) fetchPowChallenge() string {
	url := wc.BaseURL + "/api/v0/chat/create_pow_challenge"
	payload := map[string]string{"target_path": "/api/v0/chat/completion"}
	data, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return ""
	}
	req.Header = wc.buildBaseHeaders()

	resp, err := wc.client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[POW] request error: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return ""
	}

	var result struct {
		Data struct {
			BizData struct {
				Challenge struct {
					Algorithm  string `json:"algorithm"`
					Challenge  string `json:"challenge"`
					Salt       string `json:"salt"`
					Difficulty int    `json:"difficulty"`
					ExpireAt   int64  `json:"expire_at"`
					Signature  string `json:"signature"`
					TargetPath string `json:"target_path"`
				} `json:"challenge"`
			} `json:"biz_data"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return ""
	}

	ch := result.Data.BizData.Challenge
	if ch.Challenge == "" || ch.Algorithm != "DeepSeekHashV1" {
		return ""
	}

	solver := wc.getSolver()
	if solver == nil {
		return ""
	}

	powResult, solveErr := solver.Solve(ch.Challenge, ch.Salt, ch.ExpireAt, ch.Difficulty, ch.Signature, ch.TargetPath)
	if solveErr != nil {
		fmt.Fprintf(os.Stderr, "[POW] solve error: %v\n", solveErr)
		return ""
	}
	return powResult
}

func generateClientStreamID() string {
	now := time.Now()
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%04d%02d%02d-%x", now.Year(), now.Month(), now.Day(), b)
}

// timezoneOffset returns the local timezone offset in seconds, formatted as
// a string (e.g. "10800" for UTC+3). DeepSeek's web API includes this
// header for session fingerprinting.
func timezoneOffset() string {
	_, offset := time.Now().Zone()
	return fmt.Sprintf("%d", offset)
}

// ChatCompletionStream sends a streaming completion request and invokes the
// handler for each parsed event. The responseMessageID is captured for the
// caller to chain follow-up requests.
func (wc *WebClient) ChatCompletionStream(opts CompletionOpts, handler StreamHandler) (string, error) {
	url := wc.BaseURL + "/api/v0/chat/completion"

	// Resolve model_type and feature flags. Callers can pass either the
	// legacy Model string (e.g. "expert" or "default") or a fully-decomposed
	// Spec. We honour Spec when its model_type is non-empty; otherwise we
	// fall back to parsing Model as a friendly name.
	spec := opts.Spec
	if spec.ModelType == "" {
		if opts.Model != "" {
			spec = ParseModelName(opts.Model)
		} else {
			spec = ModelSpec{ModelType: "default"}
		}
	}

	clientStreamID := generateClientStreamID()
	powHeader := wc.fetchPowChallenge()

	payload := map[string]interface{}{
		"chat_session_id":  opts.SessionID,
		"prompt":           opts.Prompt,
		"model_type":       spec.ModelType,
		"stream":           true,
		"ref_file_ids":     []string{},
		"thinking_enabled": spec.ThinkingEnabled,
		"search_enabled":   spec.SearchEnabled,
		"preempt":          false,
		"client_stream_id": clientStreamID,
	}
	if opts.ParentMsgID != nil && *opts.ParentMsgID != "" {
		var id int64
		fmt.Sscanf(*opts.ParentMsgID, "%d", &id)
		if id > 0 {
			payload["parent_message_id"] = id
		}
	}
	data, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("request completion: %w", err)
	}
	req.Header = wc.buildBaseHeaders()
	req.Header.Set("x-client-stream-id", clientStreamID)
	if powHeader != "" {
		req.Header.Set("x-ds-pow-response", powHeader)
	}

	resp, err := wc.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("chat completion: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("completion status %d: %s", resp.StatusCode, string(body))
	}

	var lastID string
	var streamErr error
	parseSSE(resp.Body, func(event StreamEvent) bool {
		if event.Event == "error" {
			// Store error and stop parsing
			streamErr = fmt.Errorf("stream error: %s", event.Data)
			return false
		}
		if event.Event == "content" {
			return handler(event)
		} else if event.Event == "debug" {
			fmt.Fprintf(os.Stderr, "[DEBUG] %s\n", event.Data)
		} else if event.Event == "message_id" {
			lastID = event.Data
			// Forward message_id to handler so provider can track parent_message_id
			return handler(event)
		} else if event.Event == "status" {
			// Forward status events (FINISHED, etc.) to handler
			return handler(event)
		} else if event.Event == "usage" {
			// Forward usage events to handler for token tracking
			return handler(event)
		}
		return true
	})

	if streamErr != nil {
		return "", streamErr
	}
	return lastID, nil
}

func parseSSE(r io.Reader, handler StreamHandler) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			break
		}

		// Try to capture response_message_id from any data line. The
		// first data frame looks like
		//   {"request_message_id":1,"response_message_id":2,"model_type":"default"}
		// — see TestLiveDumpRawSSE for the full surface.
		var probe map[string]interface{}
		if json.Unmarshal([]byte(data), &probe) == nil {
			if rid, ok := probe["response_message_id"].(float64); ok {
				if !handler(StreamEvent{Event: "message_id", Data: fmt.Sprintf("%.0f", rid)}) {
					return
				}
			}
			// BATCH update frames look like
			//   {"p":"response","o":"BATCH","v":[
			//      {"p":"accumulated_token_usage","v":55},
			//      {"p":"quasi_status","v":"FINISHED"}
			//   ]}
			// accumulated_token_usage is the server-reported *output* token
			// count for this message — we surface it so the conversation
			// loop's usage tracker gets real numbers, not local estimates.
			// quasi_status=FINISHED is the server's authoritative "done"
			// signal (independent of stream EOF). When the provider sees
			// it, it returns false from the handler to stop parseSSE
			// from reading further — this is how we end the stream
			// cleanly without a brittle literal-suffix strip on the
			// content deltas.
			if p, _ := probe["p"].(string); p == "response" && probe["o"] == "BATCH" {
				if arr, ok := probe["v"].([]interface{}); ok {
					for _, item := range arr {
						m, ok := item.(map[string]interface{})
						if !ok {
							continue
						}
						switch m["p"] {
						case "accumulated_token_usage":
							if n, ok := m["v"].(float64); ok {
								if !handler(StreamEvent{Event: "usage", Data: fmt.Sprintf("%.0f", n)}) {
									return
								}
							}
						case "quasi_status":
							if s, ok := m["v"].(string); ok {
								if !handler(StreamEvent{Event: "status", Data: s}) {
									return
								}
							}
						}
					}
				}
			}
		}

		// Check for error events (rate limit, etc.)
		if errType, _ := probe["type"].(string); errType == "error" {
			if content, _ := probe["content"].(string); content != "" {
				// Return error as a special event that the caller can detect
				if !handler(StreamEvent{Event: "error", Data: content}) {
					return
				}
			}
		}

		if content := parseDeepSeekSseData(data); content != "" {
			if !handler(StreamEvent{Event: "content", Data: content}) {
				return
			}
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "[deepseek] SSE scanner error: %v\n", err)
	}
}

func parseDeepSeekSseData(dataStr string) string {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
		return ""
	}

	if v, ok := data["v"].(map[string]interface{}); ok {
		if resp, ok := v["response"].(map[string]interface{}); ok {
			if frags, ok := resp["fragments"].([]interface{}); ok {
				for _, f := range frags {
					if frag, ok := f.(map[string]interface{}); ok {
						if content, _ := frag["content"].(string); content != "" {
							return content
						}
					}
				}
			}
		}
	}

	if p, _ := data["p"].(string); p == "response/fragments" && data["o"] == "APPEND" {
		if arr, ok := data["v"].([]interface{}); ok {
			for _, item := range arr {
				if frag, ok := item.(map[string]interface{}); ok {
					if content, _ := frag["content"].(string); content != "" {
						return content
					}
				}
			}
		}
	}

	if p, _ := data["p"].(string); strings.HasSuffix(p, "/content") {
		if v, ok := data["v"].(string); ok && v != "" {
			return v
		}
	}

	if v, ok := data["v"].(string); ok && v != "" {
		return v
	}

	if choices, ok := data["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if delta, ok := choice["delta"].(map[string]interface{}); ok {
				if content, ok := delta["content"].(string); ok && content != "" {
					return content
				}
			}
		}
	}

	return ""
}

// Compile-time interface check.
var _ api.APIClient = (*Client)(nil)
var _ api.Provider = (*Provider)(nil)
