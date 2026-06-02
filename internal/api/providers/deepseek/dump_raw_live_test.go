//go:build live

package deepseek

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// TestLiveDumpRawSSE captures the verbatim bytes the DeepSeek server
// returns for a chat completion. Use it to map the actual event
// surface — we want to know:
//   - whether response_message_id is in the data lines
//   - whether usage stats (input_tokens / output_tokens) are emitted
//   - what other unknown fields exist
func TestLiveDumpRawSSE(t *testing.T) {
	tok := os.Getenv("DEEPSEEK_TOKEN")
	if tok == "" {
		t.Skip("DEEPSEEK_TOKEN not set")
	}

	wc := NewWebClient(tok)
	sid, err := wc.CreateChatSession()
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	t.Logf("session: %s", sid)

	// Solve PoW.
	powHeader := wc.fetchPowChallenge()
	t.Logf("pow header length: %d", len(powHeader))

	now := time.Now()
	b := make([]byte, 8)
	rand.Read(b)
	streamID := fmt.Sprintf("%04d%02d%02d-%x", now.Year(), now.Month(), now.Day(), b)

	payload, _ := json.Marshal(map[string]interface{}{
		"chat_session_id":   sid,
		"prompt":            "What is 6 * 7? One short sentence, no markdown.",
		"model_type":        "default",
		"stream":            true,
		"ref_file_ids":      []string{},
		"thinking_enabled":  false,
		"search_enabled":    false,
		"preempt":           false,
		"client_stream_id":  streamID,
	})

	req, err := http.NewRequestWithContext(context.Background(), "POST",
		"https://chat.deepseek.com/api/v0/chat/completion", bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	h := wc.buildBaseHeaders()
	for k, vv := range h {
		req.Header.Set(k, vv[0])
	}
	req.Header.Set("x-client-stream-id", streamID)
	if powHeader != "" {
		req.Header.Set("x-ds-pow-response", powHeader)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// Dump response headers first.
	t.Logf("response status: %s", resp.Status)
	for k, v := range resp.Header {
		t.Logf("response header %s: %s", k, strings.Join(v, ", "))
	}

	// Now read the SSE stream line by line and capture every data: line.
	br := bufio.NewReader(resp.Body)
	var allData []string
	var totalBytes int
	for {
		line, err := br.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		totalBytes += len(line)
		line = strings.TrimRight(line, "\r\n")
		if strings.HasPrefix(line, "data:") {
			allData = append(allData, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}
	t.Logf("stream: %d data lines, %d total bytes", len(allData), totalBytes)

	// Print first 5 data lines, then a sample of middle, then the last 3.
	const head, tail = 5, 3
	for i, d := range allData {
		if i < head || i >= len(allData)-tail {
			t.Logf("  data[%d]: %s", i, truncate(d, 400))
		} else if i == head {
			t.Logf("  ... (%d more data lines) ...", len(allData)-head-tail)
		}
	}

	// Hunt for usage info across the whole stream.
	for i, d := range allData {
		low := strings.ToLower(d)
		if strings.Contains(low, "usage") ||
			strings.Contains(low, "input_tokens") ||
			strings.Contains(low, "output_tokens") ||
			strings.Contains(low, "token_count") {
			t.Logf("USAGE-LIKE data[%d]: %s", i, truncate(d, 400))
		}
	}

	// Hunt for the response_message_id that parentMsgID is supposed to
	// come from.
	for i, d := range allData {
		if strings.Contains(d, "response_message_id") ||
			strings.Contains(d, "message_id") ||
			strings.Contains(d, "parent_message_id") {
			t.Logf("ID-LIKE data[%d]: %s", i, truncate(d, 400))
		}
	}

	// Check whether the *last* data line has any closing fields.
	if n := len(allData); n > 0 {
		t.Logf("last data: %s", truncate(allData[n-1], 400))
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…[+]"
}
