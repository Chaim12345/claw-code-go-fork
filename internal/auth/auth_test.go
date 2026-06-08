package auth

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"
)

// --- oauth.go tests ---

func TestGeneratePKCEPair(t *testing.T) {
	pkce, err := GeneratePKCEPair()
	if err != nil {
		t.Fatalf("GeneratePKCEPair: %v", err)
	}
	if pkce.Verifier == "" {
		t.Error("Verifier is empty")
	}
	if pkce.Challenge == "" {
		t.Error("Challenge is empty")
	}
	if pkce.Verifier == pkce.Challenge {
		t.Error("Verifier and Challenge should differ")
	}
}

func TestGeneratePKCEPair_RandError(t *testing.T) {
	orig := randRead
	randRead = func(b []byte) (int, error) { return 0, errors.New("rand failed") }
	defer func() { randRead = orig }()
	_, err := GeneratePKCEPair()
	if err == nil {
		t.Fatal("expected error from GeneratePKCEPair")
	}
}

func TestGenerateState(t *testing.T) {
	s, err := generateState()
	if err != nil {
		t.Fatalf("generateState: %v", err)
	}
	if s == "" {
		t.Error("state is empty")
	}
}

func TestGenerateState_RandError(t *testing.T) {
	orig := randRead
	randRead = func(b []byte) (int, error) { return 0, errors.New("rand failed") }
	defer func() { randRead = orig }()
	_, err := generateState()
	if err == nil {
		t.Fatal("expected error from generateState")
	}
}

func TestBase64URLEncode(t *testing.T) {
	got := base64URLEncode([]byte("hello"))
	if strings.Contains(got, "+") || strings.Contains(got, "/") || strings.Contains(got, "=") {
		t.Errorf("base64URLEncode produced non-URL-safe result: %q", got)
	}
}

func TestBuildAuthURL(t *testing.T) {
	u := buildAuthURL("http://127.0.0.1:8080/callback", "challenge123", "state123")
	if !strings.HasPrefix(u, AuthorizeURL+"?") {
		t.Errorf("URL does not start with AuthorizeURL: %q", u)
	}
	parsed, err := url.Parse(u)
	if err != nil {
		t.Fatalf("parse auth URL: %v", err)
	}
	q := parsed.Query()
	if q.Get("client_id") != ClientID {
		t.Errorf("client_id = %q, want %q", q.Get("client_id"), ClientID)
	}
	if q.Get("code_challenge") != "challenge123" {
		t.Errorf("code_challenge = %q, want challenge123", q.Get("code_challenge"))
	}
	if q.Get("state") != "state123" {
		t.Errorf("state = %q, want state123", q.Get("state"))
	}
	if q.Get("response_type") != "code" {
		t.Errorf("response_type = %q, want code", q.Get("response_type"))
	}
	if q.Get("code_challenge_method") != "S256" {
		t.Errorf("code_challenge_method = %q, want S256", q.Get("code_challenge_method"))
	}
}

func TestPostTokenRequest_Success(t *testing.T) {
	origHTTP := httpPostForm
	httpPostForm = func(rawURL string, data url.Values) (*http.Response, error) {
		body := `{"access_token":"at_123","refresh_token":"rt_456","expires_in":3600,"token_type":"bearer","scope":"user:inference"}`
		return &http.Response{StatusCode: 200, Body: nopCloser(body)}, nil
	}
	defer func() { httpPostForm = origHTTP }()

	origTime := timeNow
	timeNow = func() time.Time { return time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC) }
	defer func() { timeNow = origTime }()

	td, err := postTokenRequest(url.Values{"grant_type": {"authorization_code"}})
	if err != nil {
		t.Fatalf("postTokenRequest: %v", err)
	}
	if td.AccessToken != "at_123" {
		t.Errorf("AccessToken = %q, want at_123", td.AccessToken)
	}
	if td.RefreshToken != "rt_456" {
		t.Errorf("RefreshToken = %q, want rt_456", td.RefreshToken)
	}
	if td.TokenType != "bearer" {
		t.Errorf("TokenType = %q, want bearer", td.TokenType)
	}
}

func TestPostTokenRequest_PostError(t *testing.T) {
	orig := httpPostForm
	httpPostForm = func(string, url.Values) (*http.Response, error) {
		return nil, errors.New("connection refused")
	}
	defer func() { httpPostForm = orig }()

	_, err := postTokenRequest(url.Values{})
	if err == nil {
		t.Fatal("expected error from postTokenRequest")
	}
}

func TestPostTokenRequest_ReadError(t *testing.T) {
	origHTTP := httpPostForm
	httpPostForm = func(string, url.Values) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: nopCloser("")}, nil
	}
	defer func() { httpPostForm = origHTTP }()

	origIO := ioReadAll
	ioReadAll = func(r io.Reader) ([]byte, error) {
		return nil, errors.New("read failed")
	}
	defer func() { ioReadAll = origIO }()

	_, err := postTokenRequest(url.Values{})
	if err == nil {
		t.Fatal("expected error from postTokenRequest")
	}
}

func TestPostTokenRequest_NonOK(t *testing.T) {
	orig := httpPostForm
	httpPostForm = func(string, url.Values) (*http.Response, error) {
		return &http.Response{StatusCode: 401, Body: nopCloser("unauthorized")}, nil
	}
	defer func() { httpPostForm = orig }()

	_, err := postTokenRequest(url.Values{})
	if err == nil {
		t.Fatal("expected error from postTokenRequest")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("error should contain status code 401: %v", err)
	}
}

func TestPostTokenRequest_BadJSON(t *testing.T) {
	origHTTP := httpPostForm
	httpPostForm = func(string, url.Values) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: nopCloser("{bad json}")}, nil
	}
	defer func() { httpPostForm = origHTTP }()

	origUnmarshal := jsonUnmarshal
	jsonUnmarshal = func(data []byte, v interface{}) error {
		return errors.New("invalid JSON")
	}
	defer func() { jsonUnmarshal = origUnmarshal }()

	_, err := postTokenRequest(url.Values{})
	if err == nil {
		t.Fatal("expected error from postTokenRequest")
	}
}

func TestExchangeCode(t *testing.T) {
	orig := httpPostForm
	httpPostForm = func(rawURL string, data url.Values) (*http.Response, error) {
		if data.Get("grant_type") != "authorization_code" {
			t.Errorf("grant_type = %q, want authorization_code", data.Get("grant_type"))
		}
		if data.Get("code") != "test_code" {
			t.Errorf("code = %q, want test_code", data.Get("code"))
		}
		if data.Get("code_verifier") != "test_verifier" {
			t.Errorf("code_verifier = %q, want test_verifier", data.Get("code_verifier"))
		}
		body := `{"access_token":"at","refresh_token":"rt","expires_in":7200,"token_type":"bearer","scope":"scope"}`
		return &http.Response{StatusCode: 200, Body: nopCloser(body)}, nil
	}
	defer func() { httpPostForm = orig }()

	origTime := timeNow
	timeNow = func() time.Time { return time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC) }
	defer func() { timeNow = origTime }()

	td, err := ExchangeCode("test_code", "test_verifier", "http://127.0.0.1:0/callback")
	if err != nil {
		t.Fatalf("ExchangeCode: %v", err)
	}
	if td.AccessToken != "at" {
		t.Errorf("AccessToken = %q, want at", td.AccessToken)
	}
}

func TestRefreshToken(t *testing.T) {
	orig := httpPostForm
	httpPostForm = func(rawURL string, data url.Values) (*http.Response, error) {
		if data.Get("grant_type") != "refresh_token" {
			t.Errorf("grant_type = %q, want refresh_token", data.Get("grant_type"))
		}
		if data.Get("refresh_token") != "old_refresh" {
			t.Errorf("refresh_token = %q, want old_refresh", data.Get("refresh_token"))
		}
		body := `{"access_token":"new_at","refresh_token":"new_rt","expires_in":7200,"token_type":"bearer","scope":"user:inference"}`
		return &http.Response{StatusCode: 200, Body: nopCloser(body)}, nil
	}
	defer func() { httpPostForm = orig }()

	origTime := timeNow
	timeNow = func() time.Time { return time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC) }
	defer func() { timeNow = origTime }()

	td, err := RefreshToken("old_refresh")
	if err != nil {
		t.Fatalf("RefreshToken: %v", err)
	}
	if td.AccessToken != "new_at" {
		t.Errorf("AccessToken = %q, want new_at", td.AccessToken)
	}
}

func TestOpenBrowser(t *testing.T) {
	var captured []string
	orig := openBrowserCmd
	openBrowserCmd = func(name string, args ...string) *exec.Cmd {
		captured = append(captured, name)
		captured = append(captured, args...)
		return exec.Command("true")
	}
	defer func() { openBrowserCmd = orig }()

	openBrowser("http://example.com")
	if len(captured) < 2 || captured[0] != "xdg-open" || captured[1] != "http://example.com" {
		t.Errorf("openBrowser called with wrong args: %v", captured)
	}
}

func TestStartOAuthFlow_ListenerError(t *testing.T) {
	orig := netListen
	netListen = func(network, addr string) (net.Listener, error) {
		return nil, errors.New("bind failed")
	}
	defer func() { netListen = orig }()

	_, err := StartOAuthFlow()
	if err == nil {
		t.Fatal("expected error from StartOAuthFlow")
	}
}

func TestStartOAuthFlow_CallbackError(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port

	netListen = func(network, addr string) (net.Listener, error) { return ln, nil }
	openBrowserCmd = func(name string, args ...string) *exec.Cmd { return exec.Command("true") }

	resultCh := make(chan error, 1)
	go func() {
		_, err := StartOAuthFlow()
		resultCh <- err
	}()

	time.Sleep(100 * time.Millisecond)
	http.Get(fmt.Sprintf("http://127.0.0.1:%d/callback?error=access_denied&error_description=User+denied", port))

	select {
	case err := <-resultCh:
		if err == nil {
			t.Fatal("expected error from callback")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for StartOAuthFlow")
	}
}

func TestStartOAuthFlow_StateMismatch(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port

	netListen = func(network, addr string) (net.Listener, error) { return ln, nil }
	openBrowserCmd = func(name string, args ...string) *exec.Cmd { return exec.Command("true") }

	resultCh := make(chan error, 1)
	go func() {
		_, err := StartOAuthFlow()
		resultCh <- err
	}()

	time.Sleep(100 * time.Millisecond)
	http.Get(fmt.Sprintf("http://127.0.0.1:%d/callback?code=abc&state=wrong_state", port))

	select {
	case err := <-resultCh:
		if err == nil {
			t.Fatal("expected error from state mismatch")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for StartOAuthFlow")
	}
}

func TestStartOAuthFlow_Success(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port

	var mu sync.Mutex
	var capturedURL string

	netListen = func(network, addr string) (net.Listener, error) { return ln, nil }
	openBrowserCmd = func(name string, args ...string) *exec.Cmd {
		mu.Lock()
		capturedURL = args[0]
		mu.Unlock()
		return exec.Command("true")
	}
	httpPostForm = func(rawURL string, data url.Values) (*http.Response, error) {
		body := `{"access_token":"new_at","refresh_token":"new_rt","expires_in":3600,"token_type":"bearer","scope":"user:inference"}`
		return &http.Response{StatusCode: 200, Body: nopCloser(body)}, nil
	}
	origTime := timeNow
	timeNow = func() time.Time { return time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC) }
	defer func() {
		netListen = net.Listen
		openBrowserCmd = exec.Command
		httpPostForm = http.PostForm
		timeNow = origTime
	}()

	type result struct {
		td  *TokenData
		err error
	}
	resultCh := make(chan result, 1)
	go func() {
		td, err := StartOAuthFlow()
		resultCh <- result{td, err}
	}()

	for i := 0; i < 50; i++ {
		mu.Lock()
		u := capturedURL
		mu.Unlock()
		if u != "" {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	mu.Lock()
	u := capturedURL
	mu.Unlock()
	if u == "" {
		t.Fatal("openBrowser was never called")
	}

	parsed, _ := url.Parse(u)
	state := parsed.Query().Get("state")

	http.Get(fmt.Sprintf("http://127.0.0.1:%d/callback?code=test_code&state=%s", port, state))

	select {
	case r := <-resultCh:
		if r.err != nil {
			t.Fatalf("StartOAuthFlow: %v", r.err)
		}
		if r.td == nil {
			t.Fatal("expected non-nil TokenData")
		}
		if r.td.AccessToken != "new_at" {
			t.Errorf("AccessToken = %q, want new_at", r.td.AccessToken)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for StartOAuthFlow")
	}
}

func TestPrepareOAuthFlow(t *testing.T) {
	session, err := PrepareOAuthFlow()
	if err != nil {
		t.Fatalf("PrepareOAuthFlow: %v", err)
	}
	if session.AuthURL == "" {
		t.Error("AuthURL is empty")
	}
	if !strings.HasPrefix(session.AuthURL, AuthorizeURL) {
		t.Errorf("AuthURL = %q, want prefix %q", session.AuthURL, AuthorizeURL)
	}
	session.listener.Close()
}

func TestPrepareOAuthFlow_ListenerError(t *testing.T) {
	orig := netListen
	netListen = func(network, addr string) (net.Listener, error) {
		return nil, errors.New("bind failed")
	}
	defer func() { netListen = orig }()

	_, err := PrepareOAuthFlow()
	if err == nil {
		t.Fatal("expected error from PrepareOAuthFlow")
	}
}

func TestPrepareOAuthFlow_GenerateStateError(t *testing.T) {
	origRand := randRead
	randRead = func(b []byte) (int, error) { return 0, errors.New("rand failed") }
	defer func() { randRead = origRand }()

	_, err := PrepareOAuthFlow()
	if err == nil {
		t.Fatal("expected error from PrepareOAuthFlow")
	}
}

func TestComplete_Success(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port

	// Create a session manually
	pkce, _ := GeneratePKCEPair()
	state := "test_state_val"
	redirectURI := fmt.Sprintf("http://127.0.0.1:%d/callback", port)
	session := &OAuthSession{
		AuthURL:     buildAuthURL(redirectURI, pkce.Challenge, state),
		listener:    ln,
		redirectURI: redirectURI,
		pkce:        pkce,
		state:       state,
	}

	openBrowserCmd = func(name string, args ...string) *exec.Cmd { return exec.Command("true") }
	httpPostForm = func(rawURL string, data url.Values) (*http.Response, error) {
		body := `{"access_token":"comp_at","refresh_token":"comp_rt","expires_in":3600,"token_type":"bearer","scope":"user:inference"}`
		return &http.Response{StatusCode: 200, Body: nopCloser(body)}, nil
	}
	origTime := timeNow
	timeNow = func() time.Time { return time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC) }
	defer func() {
		openBrowserCmd = exec.Command
		httpPostForm = http.PostForm
		timeNow = origTime
	}()

	type result struct {
		td  *TokenData
		err error
	}
	resultCh := make(chan result, 1)
	go func() {
		td, err := session.Complete()
		resultCh <- result{td, err}
	}()

	time.Sleep(100 * time.Millisecond)
	http.Get(fmt.Sprintf("http://127.0.0.1:%d/callback?code=comp_code&state=%s", port, state))

	select {
	case r := <-resultCh:
		if r.err != nil {
			t.Fatalf("Complete: %v", r.err)
		}
		if r.td == nil {
			t.Fatal("expected non-nil TokenData")
		}
		if r.td.AccessToken != "comp_at" {
			t.Errorf("AccessToken = %q, want comp_at", r.td.AccessToken)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for Complete")
	}
}

func TestComplete_ErrorCallback(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port

	pkce, _ := GeneratePKCEPair()
	session := &OAuthSession{
		AuthURL:     "http://example.com",
		listener:    ln,
		redirectURI: fmt.Sprintf("http://127.0.0.1:%d/callback", port),
		pkce:        pkce,
		state:       "s",
	}

	openBrowserCmd = func(name string, args ...string) *exec.Cmd { return exec.Command("true") }
	defer func() { openBrowserCmd = exec.Command }()

	type result struct {
		td  *TokenData
		err error
	}
	resultCh := make(chan result, 1)
	go func() {
		td, err := session.Complete()
		resultCh <- result{td, err}
	}()

	time.Sleep(100 * time.Millisecond)
	http.Get(fmt.Sprintf("http://127.0.0.1:%d/callback?error=denied&error_description=no+way", port))

	select {
	case r := <-resultCh:
		if r.err == nil {
			t.Fatal("expected error from callback")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for Complete")
	}
}

func TestComplete_StateMismatch(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port

	pkce, _ := GeneratePKCEPair()
	session := &OAuthSession{
		AuthURL:     "http://example.com",
		listener:    ln,
		redirectURI: fmt.Sprintf("http://127.0.0.1:%d/callback", port),
		pkce:        pkce,
		state:       "correct_state",
	}

	openBrowserCmd = func(name string, args ...string) *exec.Cmd { return exec.Command("true") }
	defer func() { openBrowserCmd = exec.Command }()

	type result struct {
		td  *TokenData
		err error
	}
	resultCh := make(chan result, 1)
	go func() {
		td, err := session.Complete()
		resultCh <- result{td, err}
	}()

	time.Sleep(100 * time.Millisecond)
	http.Get(fmt.Sprintf("http://127.0.0.1:%d/callback?code=abc&state=wrong", port))

	select {
	case r := <-resultCh:
		if r.err == nil {
			t.Fatal("expected error from state mismatch")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for Complete")
	}
}

// --- storage.go tests ---

func TestAuthFilePath(t *testing.T) {
	orig := osUserHomeDir
	osUserHomeDir = func() (string, error) { return "/tmp/testhome", nil }
	defer func() { osUserHomeDir = orig }()

	path, err := authFilePath()
	if err != nil {
		t.Fatalf("authFilePath: %v", err)
	}
	if path != "/tmp/testhome/.claw-code/auth.json" {
		t.Errorf("path = %q", path)
	}
}

func TestAuthFilePath_Error(t *testing.T) {
	orig := osUserHomeDir
	osUserHomeDir = func() (string, error) { return "", errors.New("no home") }
	defer func() { osUserHomeDir = orig }()

	_, err := authFilePath()
	if err == nil {
		t.Fatal("expected error from authFilePath")
	}
}

func TestLoadTokens_Success(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// SaveTokens encrypts, LoadTokens decrypts — full round-trip.
	want := &TokenData{AccessToken: "at", RefreshToken: "rt", ExpiresAt: time.Now(), TokenType: "bearer", Scope: "user:inference"}
	if err := SaveTokens(want); err != nil {
		t.Fatalf("SaveTokens: %v", err)
	}

	td, err := LoadTokens()
	if err != nil {
		t.Fatalf("LoadTokens: %v", err)
	}
	if td.AccessToken != "at" {
		t.Errorf("AccessToken = %q, want at", td.AccessToken)
	}
}

func TestLoadTokens_FileNotExist(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	_, err := LoadTokens()
	if err == nil {
		t.Fatal("expected error from LoadTokens")
	}
}

func TestLoadTokens_BadJSON(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	os.MkdirAll(tmpHome+"/.claw-code", 0700)
	os.WriteFile(tmpHome+"/.claw-code/auth.json", []byte("{bad json}"), 0600)

	_, err := LoadTokens()
	if err == nil {
		t.Fatal("expected error from LoadTokens")
	}
}

func TestLoadTokens_ReadFileError(t *testing.T) {
	orig := osReadFile
	osReadFile = func(name string) ([]byte, error) { return nil, errors.New("permission denied") }
	defer func() { osReadFile = orig }()

	_, err := LoadTokens()
	if err == nil {
		t.Fatal("expected error from LoadTokens")
	}
}

func TestLoadTokens_UnmarshalError(t *testing.T) {
	origUnmarshal := jsonUnmarshal
	jsonUnmarshal = func(data []byte, v interface{}) error { return errors.New("unmarshal failed") }
	defer func() { jsonUnmarshal = origUnmarshal }()

	orig := osReadFile
	osReadFile = func(name string) ([]byte, error) {
		return []byte("v1:valid-looking-ciphertext"), nil
	}
	defer func() { osReadFile = orig }()

	_, err := LoadTokens()
	if err == nil {
		t.Fatal("expected error from LoadTokens")
	}
}

func TestSaveTokens_Success(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	td := &TokenData{AccessToken: "at", RefreshToken: "rt", ExpiresAt: time.Now(), TokenType: "bearer"}
	if err := SaveTokens(td); err != nil {
		t.Fatalf("SaveTokens: %v", err)
	}

	loaded, err := LoadTokens()
	if err != nil {
		t.Fatalf("LoadTokens: %v", err)
	}
	if loaded.AccessToken != "at" {
		t.Errorf("AccessToken = %q, want at", loaded.AccessToken)
	}
}

func TestSaveTokens_MkdirError(t *testing.T) {
	orig := osMkdirAll
	osMkdirAll = func(path string, perm os.FileMode) error { return errors.New("mkdir failed") }
	defer func() { osMkdirAll = orig }()

	err := SaveTokens(&TokenData{})
	if err == nil {
		t.Fatal("expected error from SaveTokens")
	}
}

func TestSaveTokens_MarshalError(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	orig := jsonMarshalIndent
	jsonMarshalIndent = func(v interface{}, prefix, indent string) ([]byte, error) {
		return nil, errors.New("marshal failed")
	}
	defer func() { jsonMarshalIndent = orig }()

	err := SaveTokens(&TokenData{})
	if err == nil {
		t.Fatal("expected error from SaveTokens")
	}
}

func TestSaveTokens_WriteFileError(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	orig := osWriteFile
	osWriteFile = func(name string, data []byte, perm os.FileMode) error { return errors.New("write failed") }
	defer func() { osWriteFile = orig }()

	err := SaveTokens(&TokenData{})
	if err == nil {
		t.Fatal("expected error from SaveTokens")
	}
}

func TestSaveTokens_AuthFilePathError(t *testing.T) {
	orig := osUserHomeDir
	osUserHomeDir = func() (string, error) { return "", errors.New("no home") }
	defer func() { osUserHomeDir = orig }()

	err := SaveTokens(&TokenData{})
	if err == nil {
		t.Fatal("expected error from SaveTokens")
	}
}

func TestClearTokens_Success(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	SaveTokens(&TokenData{AccessToken: "at"})
	if err := ClearTokens(); err != nil {
		t.Fatalf("ClearTokens: %v", err)
	}

	_, err := LoadTokens()
	if err == nil {
		t.Fatal("expected error after clearing tokens")
	}
}

func TestClearTokens_NotExist(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	if err := ClearTokens(); err != nil {
		t.Fatalf("ClearTokens: %v", err)
	}
}

func TestClearTokens_RemoveError(t *testing.T) {
	orig := osRemove
	osRemove = func(name string) error { return errors.New("remove failed") }
	defer func() { osRemove = orig }()

	err := ClearTokens()
	if err == nil {
		t.Fatal("expected error from ClearTokens")
	}
}

func TestClearTokens_AuthFilePathError(t *testing.T) {
	orig := osUserHomeDir
	osUserHomeDir = func() (string, error) { return "", errors.New("no home") }
	defer func() { osUserHomeDir = orig }()

	err := ClearTokens()
	if err == nil {
		t.Fatal("expected error from ClearTokens")
	}
}

func TestIsExpired_ZeroTime(t *testing.T) {
	td := &TokenData{ExpiresAt: time.Time{}}
	if IsExpired(td) {
		t.Error("zero time should not be expired")
	}
}

func TestIsExpired_Expired(t *testing.T) {
	orig := timeNow
	timeNow = func() time.Time { return time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC) }
	defer func() { timeNow = orig }()

	td := &TokenData{ExpiresAt: time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC)}
	if !IsExpired(td) {
		t.Error("expected token to be expired")
	}
}

func TestIsExpired_NotExpired(t *testing.T) {
	orig := timeNow
	timeNow = func() time.Time { return time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC) }
	defer func() { timeNow = orig }()

	td := &TokenData{ExpiresAt: time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)}
	if IsExpired(td) {
		t.Error("expected token to not be expired")
	}
}

// --- manager.go tests ---

func TestGetAccessToken_APIKey(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "sk-test-key")
	key, err := GetAccessToken()
	if err != nil {
		t.Fatalf("GetAccessToken: %v", err)
	}
	if key != "sk-test-key" {
		t.Errorf("key = %q, want sk-test-key", key)
	}
}

func TestGetAccessToken_NoCreds(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	origLoad := loadTokensFunc
	loadTokensFunc = func() (*TokenData, error) { return nil, errors.New("no tokens") }
	defer func() { loadTokensFunc = origLoad }()

	_, err := GetAccessToken()
	if err == nil {
		t.Fatal("expected error from GetAccessToken")
	}
}

func TestGetAccessToken_ValidToken(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	origLoad := loadTokensFunc
	loadTokensFunc = func() (*TokenData, error) {
		return &TokenData{AccessToken: "valid_at", ExpiresAt: time.Now().Add(time.Hour)}, nil
	}
	defer func() { loadTokensFunc = origLoad }()

	origExpired := isExpiredFunc
	isExpiredFunc = func(td *TokenData) bool { return false }
	defer func() { isExpiredFunc = origExpired }()

	key, err := GetAccessToken()
	if err != nil {
		t.Fatalf("GetAccessToken: %v", err)
	}
	if key != "valid_at" {
		t.Errorf("key = %q, want valid_at", key)
	}
}

func TestGetAccessToken_ExpiredNoRefresh(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	origLoad := loadTokensFunc
	loadTokensFunc = func() (*TokenData, error) {
		return &TokenData{AccessToken: "expired_at", ExpiresAt: time.Now().Add(-time.Hour)}, nil
	}
	defer func() { loadTokensFunc = origLoad }()

	origExpired := isExpiredFunc
	isExpiredFunc = func(td *TokenData) bool { return true }
	defer func() { isExpiredFunc = origExpired }()

	_, err := GetAccessToken()
	if err == nil {
		t.Fatal("expected error from GetAccessToken")
	}
}

func TestGetAccessToken_ExpiredRefreshSuccess(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	origLoad := loadTokensFunc
	loadTokensFunc = func() (*TokenData, error) {
		return &TokenData{
			AccessToken:  "old_at",
			RefreshToken: "refresh_tok",
			ExpiresAt:    time.Now().Add(-time.Hour),
		}, nil
	}
	defer func() { loadTokensFunc = origLoad }()

	origExpired := isExpiredFunc
	isExpiredFunc = func(td *TokenData) bool { return true }
	defer func() { isExpiredFunc = origExpired }()

	origRefresh := refreshTokenFunc
	refreshTokenFunc = func(rt string) (*TokenData, error) {
		return &TokenData{AccessToken: "new_at", ExpiresAt: time.Now().Add(time.Hour)}, nil
	}
	defer func() { refreshTokenFunc = origRefresh }()

	key, err := GetAccessToken()
	if err != nil {
		t.Fatalf("GetAccessToken: %v", err)
	}
	if key != "new_at" {
		t.Errorf("key = %q, want new_at", key)
	}
}

func TestGetAccessToken_ExpiredRefreshFails(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	origLoad := loadTokensFunc
	loadTokensFunc = func() (*TokenData, error) {
		return &TokenData{
			AccessToken:  "old_at",
			RefreshToken: "refresh_tok",
			ExpiresAt:    time.Now().Add(-time.Hour),
		}, nil
	}
	defer func() { loadTokensFunc = origLoad }()

	origExpired := isExpiredFunc
	isExpiredFunc = func(td *TokenData) bool { return true }
	defer func() { isExpiredFunc = origExpired }()

	origRefresh := refreshTokenFunc
	refreshTokenFunc = func(rt string) (*TokenData, error) {
		return nil, errors.New("refresh failed")
	}
	defer func() { refreshTokenFunc = origRefresh }()

	_, err := GetAccessToken()
	if err == nil {
		t.Fatal("expected error from GetAccessToken")
	}
}

func TestGetAccessToken_ExpiredRefreshSaveFails(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	origLoad := loadTokensFunc
	loadTokensFunc = func() (*TokenData, error) {
		return &TokenData{
			AccessToken:  "old_at",
			RefreshToken: "refresh_tok",
			ExpiresAt:    time.Now().Add(-time.Hour),
		}, nil
	}
	defer func() { loadTokensFunc = origLoad }()

	origExpired := isExpiredFunc
	isExpiredFunc = func(td *TokenData) bool { return true }
	defer func() { isExpiredFunc = origExpired }()

	origRefresh := refreshTokenFunc
	refreshTokenFunc = func(rt string) (*TokenData, error) {
		return &TokenData{AccessToken: "new_at", ExpiresAt: time.Now().Add(time.Hour)}, nil
	}
	defer func() { refreshTokenFunc = origRefresh }()

	origSave := saveTokensFunc
	saveTokensFunc = func(td *TokenData) error { return errors.New("save failed") }
	defer func() { saveTokensFunc = origSave }()

	key, err := GetAccessToken()
	if err != nil {
		t.Fatalf("GetAccessToken: %v", err)
	}
	if key != "new_at" {
		t.Errorf("key = %q, want new_at", key)
	}
}

func TestIsOAuthAuth_APIKeySet(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "key")
	if IsOAuthAuth() {
		t.Error("expected false when API key is set")
	}
}

func TestIsOAuthAuth_NoTokens(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	origLoad := loadTokensFunc
	loadTokensFunc = func() (*TokenData, error) { return nil, errors.New("no tokens") }
	defer func() { loadTokensFunc = origLoad }()

	if IsOAuthAuth() {
		t.Error("expected false when no tokens")
	}
}

func TestIsOAuthAuth_HasTokens(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	origLoad := loadTokensFunc
	loadTokensFunc = func() (*TokenData, error) {
		return &TokenData{AccessToken: "at"}, nil
	}
	defer func() { loadTokensFunc = origLoad }()

	if !IsOAuthAuth() {
		t.Error("expected true when tokens exist")
	}
}

func TestGetStatus_APIKey(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "key")
	s := GetStatus()
	if !s.Authenticated {
		t.Error("expected Authenticated=true")
	}
	if s.Method != "api_key" {
		t.Errorf("Method = %q, want api_key", s.Method)
	}
}

func TestGetStatus_NoTokens(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	origLoad := loadTokensFunc
	loadTokensFunc = func() (*TokenData, error) { return nil, errors.New("no tokens") }
	defer func() { loadTokensFunc = origLoad }()

	s := GetStatus()
	if s.Authenticated {
		t.Error("expected Authenticated=false")
	}
	if s.Method != "none" {
		t.Errorf("Method = %q, want none", s.Method)
	}
}

func TestGetStatus_OAuthValid(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	origLoad := loadTokensFunc
	loadTokensFunc = func() (*TokenData, error) {
		return &TokenData{
			AccessToken:  "at",
			RefreshToken: "rt",
			ExpiresAt:    time.Now().Add(time.Hour),
		}, nil
	}
	defer func() { loadTokensFunc = origLoad }()

	origExpired := isExpiredFunc
	isExpiredFunc = func(td *TokenData) bool { return false }
	defer func() { isExpiredFunc = origExpired }()

	s := GetStatus()
	if !s.Authenticated {
		t.Error("expected Authenticated=true")
	}
	if s.Method != "oauth" {
		t.Errorf("Method = %q, want oauth", s.Method)
	}
	if !s.HasRefresh {
		t.Error("expected HasRefresh=true")
	}
}

func TestGetStatus_OAuthExpired(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	origLoad := loadTokensFunc
	loadTokensFunc = func() (*TokenData, error) {
		return &TokenData{
			AccessToken:  "at",
			RefreshToken: "rt",
			ExpiresAt:    time.Now().Add(-time.Hour),
		}, nil
	}
	defer func() { loadTokensFunc = origLoad }()

	origExpired := isExpiredFunc
	isExpiredFunc = func(td *TokenData) bool { return true }
	defer func() { isExpiredFunc = origExpired }()

	s := GetStatus()
	if s.Authenticated {
		t.Error("expected Authenticated=false")
	}
	if s.Method != "oauth" {
		t.Errorf("Method = %q, want oauth", s.Method)
	}
}

// --- credentials.go tests ---

func TestCredentialsFilePath(t *testing.T) {
	orig := osUserHomeDir
	osUserHomeDir = func() (string, error) { return "/tmp/testhome", nil }
	defer func() { osUserHomeDir = orig }()

	path, err := credentialsFilePath()
	if err != nil {
		t.Fatalf("credentialsFilePath: %v", err)
	}
	if path != "/tmp/testhome/.claw-code/credentials.json" {
		t.Errorf("path = %q", path)
	}
}

func TestCredentialsFilePath_Error(t *testing.T) {
	orig := osUserHomeDir
	osUserHomeDir = func() (string, error) { return "", errors.New("no home") }
	defer func() { osUserHomeDir = orig }()

	_, err := credentialsFilePath()
	if err == nil {
		t.Fatal("expected error from credentialsFilePath")
	}
}

func TestLoadCredentialStore_FileNotExist(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	store, err := LoadCredentialStore()
	if err != nil {
		t.Fatalf("LoadCredentialStore: %v", err)
	}
	if store == nil {
		t.Fatal("expected non-nil store")
	}
	if len(store.Providers) != 0 {
		t.Errorf("expected empty providers, got %d", len(store.Providers))
	}
}

func TestLoadCredentialStore_ReadError(t *testing.T) {
	orig := osReadFile
	osReadFile = func(name string) ([]byte, error) { return nil, errors.New("permission denied") }
	defer func() { osReadFile = orig }()

	_, err := LoadCredentialStore()
	if err == nil {
		t.Fatal("expected error from LoadCredentialStore")
	}
}

func TestLoadCredentialStore_BadJSON(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	os.MkdirAll(tmpHome+"/.claw-code", 0700)
	os.WriteFile(tmpHome+"/.claw-code/credentials.json", []byte("{bad}"), 0600)

	store, err := LoadCredentialStore()
	if err != nil {
		t.Fatalf("LoadCredentialStore: %v", err)
	}
	if store == nil {
		t.Fatal("expected non-nil store")
	}
}

func TestLoadCredentialStore_NilProviders(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	os.MkdirAll(tmpHome+"/.claw-code", 0700)
	os.WriteFile(tmpHome+"/.claw-code/credentials.json", []byte(`{"active_provider":"anthropic"}`), 0600)

	store, err := LoadCredentialStore()
	if err != nil {
		t.Fatalf("LoadCredentialStore: %v", err)
	}
	if len(store.Providers) != 0 {
		t.Errorf("expected empty providers, got %d", len(store.Providers))
	}
}

func TestLoadCredentialStore_HomeError(t *testing.T) {
	orig := osUserHomeDir
	osUserHomeDir = func() (string, error) { return "", errors.New("no home") }
	defer func() { osUserHomeDir = orig }()

	store, err := LoadCredentialStore()
	if err != nil {
		t.Fatalf("LoadCredentialStore: %v", err)
	}
	if store == nil {
		t.Fatal("expected non-nil store")
	}
}

func TestSetProviderAPIKey_LoadError(t *testing.T) {
	orig := osUserHomeDir
	osUserHomeDir = func() (string, error) { return "", errors.New("no home") }
	defer func() { osUserHomeDir = orig }()

	err := SetProviderAPIKey("anthropic", "sk-test")
	if err == nil {
		t.Fatal("expected error from SetProviderAPIKey")
	}
}

func TestSetProviderOAuth_LoadError(t *testing.T) {
	orig := osUserHomeDir
	osUserHomeDir = func() (string, error) { return "", errors.New("no home") }
	defer func() { osUserHomeDir = orig }()

	err := SetProviderOAuth("anthropic", &TokenData{AccessToken: "tok"})
	if err == nil {
		t.Fatal("expected error from SetProviderOAuth")
	}
}

func TestSaveCredentialStore_Success(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	store := &CredentialStore{
		ActiveProvider: "anthropic",
		Providers: map[string]*ProviderCredentials{
			"anthropic": {AuthMethod: "api_key", APIKey: "sk-test"},
		},
	}
	if err := SaveCredentialStore(store); err != nil {
		t.Fatalf("SaveCredentialStore: %v", err)
	}

	loaded, err := LoadCredentialStore()
	if err != nil {
		t.Fatalf("LoadCredentialStore: %v", err)
	}
	if loaded.ActiveProvider != "anthropic" {
		t.Errorf("ActiveProvider = %q, want anthropic", loaded.ActiveProvider)
	}
}

func TestSaveCredentialStore_MkdirError(t *testing.T) {
	orig := osMkdirAll
	osMkdirAll = func(path string, perm os.FileMode) error { return errors.New("mkdir failed") }
	defer func() { osMkdirAll = orig }()

	err := SaveCredentialStore(&CredentialStore{})
	if err == nil {
		t.Fatal("expected error from SaveCredentialStore")
	}
}

func TestSaveCredentialStore_MarshalError(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	orig := jsonMarshalIndent
	jsonMarshalIndent = func(v interface{}, prefix, indent string) ([]byte, error) {
		return nil, errors.New("marshal failed")
	}
	defer func() { jsonMarshalIndent = orig }()

	err := SaveCredentialStore(&CredentialStore{})
	if err == nil {
		t.Fatal("expected error from SaveCredentialStore")
	}
}

func TestSaveCredentialStore_WriteFileError(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	orig := osWriteFile
	osWriteFile = func(name string, data []byte, perm os.FileMode) error { return errors.New("write failed") }
	defer func() { osWriteFile = orig }()

	err := SaveCredentialStore(&CredentialStore{})
	if err == nil {
		t.Fatal("expected error from SaveCredentialStore")
	}
}

func TestSaveCredentialStore_CredentialsFilePathError(t *testing.T) {
	orig := osUserHomeDir
	osUserHomeDir = func() (string, error) { return "", errors.New("no home") }
	defer func() { osUserHomeDir = orig }()

	err := SaveCredentialStore(&CredentialStore{})
	if err == nil {
		t.Fatal("expected error from SaveCredentialStore")
	}
}

func TestSetProviderOAuth(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	td := &TokenData{AccessToken: "oauth_at", RefreshToken: "oauth_rt", ExpiresAt: time.Now().Add(time.Hour)}
	if err := SetProviderOAuth("anthropic", td); err != nil {
		t.Fatalf("SetProviderOAuth: %v", err)
	}

	store, _ := LoadCredentialStore()
	cred, ok := store.Providers["anthropic"]
	if !ok {
		t.Fatal("anthropic provider not in store")
	}
	if cred.AuthMethod != "oauth" {
		t.Errorf("AuthMethod = %q, want oauth", cred.AuthMethod)
	}
	if cred.OAuth == nil || cred.OAuth.AccessToken != "oauth_at" {
		t.Error("OAuth token data not stored correctly")
	}
}

func TestGetActiveProvider_SetValue(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	SetProviderAPIKey("openai", "oai-key")
	got := GetActiveProvider()
	if got != "openai" {
		t.Errorf("GetActiveProvider() = %q, want openai", got)
	}
}

func TestResolveCredentials_StoredAPIKey(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("DEEPSEEK_TOKEN", "")

	SetProviderAPIKey("anthropic", "stored-key")
	provider, token, method, err := ResolveCredentials()
	if err != nil {
		t.Fatalf("ResolveCredentials: %v", err)
	}
	if provider != "anthropic" || token != "stored-key" || method != "api_key" {
		t.Errorf("got %q/%q/%q", provider, token, method)
	}
}

func TestResolveCredentials_StoredOAuth(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("DEEPSEEK_TOKEN", "")

	td := &TokenData{AccessToken: "oauth_tok", ExpiresAt: time.Now().Add(time.Hour)}
	SetProviderOAuth("anthropic", td)

	provider, token, method, err := ResolveCredentials()
	if err != nil {
		t.Fatalf("ResolveCredentials: %v", err)
	}
	if provider != "anthropic" || token != "oauth_tok" || method != "oauth" {
		t.Errorf("got %q/%q/%q", provider, token, method)
	}
}

func TestResolveCredentials_StoredOAuthExpired_HasRefresh(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("DEEPSEEK_TOKEN", "")

	td := &TokenData{
		AccessToken:  "old_tok",
		RefreshToken: "refresh_tok",
		ExpiresAt:    time.Now().Add(-time.Hour),
	}
	SetProviderOAuth("anthropic", td)

	origPost := httpPostForm
	httpPostForm = func(url string, data url.Values) (*http.Response, error) {
		respJSON := `{"access_token":"refreshed_tok","token_type":"bearer","expires_in":3600}`
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(respJSON)),
		}, nil
	}
	defer func() { httpPostForm = origPost }()

	provider, token, method, err := ResolveCredentials()
	if err != nil {
		t.Fatalf("ResolveCredentials: %v", err)
	}
	if provider != "anthropic" || token != "refreshed_tok" || method != "oauth" {
		t.Errorf("got %q/%q/%q", provider, token, method)
	}
}

func TestResolveCredentials_StoredOAuthExpired_NoRefresh(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("DEEPSEEK_TOKEN", "")

	td := &TokenData{
		AccessToken: "old_tok",
		ExpiresAt:   time.Now().Add(-time.Hour),
	}
	SetProviderOAuth("anthropic", td)

	_, _, _, err := ResolveCredentials()
	if err == nil {
		t.Fatal("expected error from ResolveCredentials")
	}
}

func TestResolveCredentials_StoredOAuthRefreshFails(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("DEEPSEEK_TOKEN", "")

	td := &TokenData{
		AccessToken:  "old_tok",
		RefreshToken: "refresh_tok",
		ExpiresAt:    time.Now().Add(-time.Hour),
	}
	SetProviderOAuth("anthropic", td)

	origRefresh := refreshTokenFunc
	refreshTokenFunc = func(rt string) (*TokenData, error) {
		return nil, errors.New("refresh failed")
	}
	defer func() { refreshTokenFunc = origRefresh }()

	_, _, _, err := ResolveCredentials()
	if err == nil {
		t.Fatal("expected error from ResolveCredentials")
	}
}

func TestResolveCredentials_LegacyFallback(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("DEEPSEEK_TOKEN", "")

	// Write encrypted legacy auth.json via SaveTokens (round-trip).
	want := &TokenData{AccessToken: "legacy_at", ExpiresAt: time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC), TokenType: "bearer"}
	if err := SaveTokens(want); err != nil {
		t.Fatalf("SaveTokens: %v", err)
	}

	provider, token, method, err := ResolveCredentials()
	if err != nil {
		t.Fatalf("ResolveCredentials: %v", err)
	}
	if provider != "anthropic" || token != "legacy_at" || method != "oauth" {
		t.Errorf("got %q/%q/%q", provider, token, method)
	}
}

func TestFingerprintKey(t *testing.T) {
	fp := fingerprintKey()
	if fp == "unknown" && os.Getenv("HOME") == "" {
		t.Skip("no HOME set")
	}
	if len(fp) != 8 {
		t.Errorf("fingerprintKey = %q, want 8 hex chars", fp)
	}
}

func TestMachineKeyPath_HomeError(t *testing.T) {
	orig := osUserHomeDir
	osUserHomeDir = func() (string, error) { return "", errors.New("no home") }
	defer func() { osUserHomeDir = orig }()

	_, err := machineKeyPath()
	if err == nil {
		t.Fatal("expected error from machineKeyPath")
	}
}

func TestDecryptTokenData_BadFormat(t *testing.T) {
	_, err := decryptTokenData([]byte("plaintext-no-magic"))
	if err == nil {
		t.Fatal("expected error from decryptTokenData")
	}
}

func TestDecryptTokenData_Truncated(t *testing.T) {
	_, err := decryptTokenData([]byte("v1:short"))
	if err == nil {
		t.Fatal("expected error from decryptTokenData")
	}
}

func TestEncryptTokenData_HomeError(t *testing.T) {
	orig := osUserHomeDir
	osUserHomeDir = func() (string, error) { return "", errors.New("no home") }
	defer func() { osUserHomeDir = orig }()

	keyOnce = sync.Once{}
	keyData = nil
	keyErr = nil
	defer func() { keyOnce = sync.Once{} }()

	_, err := encryptTokenData([]byte("test"))
	if err == nil {
		t.Fatal("expected error from encryptTokenData")
	}
}

func TestDecryptTokenData_HomeError(t *testing.T) {
	orig := osUserHomeDir
	osUserHomeDir = func() (string, error) { return "", errors.New("no home") }
	defer func() { osUserHomeDir = orig }()

	keyOnce = sync.Once{}
	keyData = nil
	keyErr = nil
	defer func() { keyOnce = sync.Once{} }()

	_, err := decryptTokenData([]byte("v1:" + string(make([]byte, 32))))
	if err == nil {
		t.Fatal("expected error from decryptTokenData")
	}
}

func nopCloser(s string) *nopReadCloser {
	return &nopReadCloser{strings.NewReader(s)}
}

type nopReadCloser struct {
	*strings.Reader
}

func (n *nopReadCloser) Close() error { return nil }
