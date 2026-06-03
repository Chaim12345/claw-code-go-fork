package web

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestAuthFromEnv_NotSet(t *testing.T) {
	os.Unsetenv("CLAW_WEB_AUTH")
	os.Unsetenv("CLAW_WEB_AUTH_FILE")
	user, pass := authFromEnv()
	if user != "" || pass != "" {
		t.Errorf("expected empty credentials, got user=%q pass=%q", user, pass)
	}
}

func TestAuthFromEnv_Direct(t *testing.T) {
	t.Setenv("CLAW_WEB_AUTH", "admin:secret")
	user, pass := authFromEnv()
	if user != "admin" || pass != "secret" {
		t.Errorf("expected admin:secret, got %s:%s", user, pass)
	}
}

func TestAuthFromEnv_NoColon(t *testing.T) {
	t.Setenv("CLAW_WEB_AUTH", "invalidformat")
	user, pass := authFromEnv()
	if user != "" || pass != "" {
		t.Errorf("expected empty when no colon, got %s:%s", user, pass)
	}
}

func TestAuthFromEnv_PasswordContainsColon(t *testing.T) {
	// strings.Cut splits on the first colon only; password may contain additional colons.
	t.Setenv("CLAW_WEB_AUTH", "user:pass:with:colons")
	user, pass := authFromEnv()
	if user != "user" || pass != "pass:with:colons" {
		t.Errorf("expected user:pass:with:colons, got %s:%s", user, pass)
	}
}

func TestAuthFromEnv_File(t *testing.T) {
	os.Unsetenv("CLAW_WEB_AUTH")
	dir := t.TempDir()
	authFile := filepath.Join(dir, "auth.txt")
	if err := os.WriteFile(authFile, []byte("fileuser:filepass\n"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLAW_WEB_AUTH_FILE", authFile)
	user, pass := authFromEnv()
	if user != "fileuser" || pass != "filepass" {
		t.Errorf("expected fileuser:filepass from file, got %s:%s", user, pass)
	}
}

func TestAuthFromEnv_FileWithWhitespace(t *testing.T) {
	os.Unsetenv("CLAW_WEB_AUTH")
	dir := t.TempDir()
	authFile := filepath.Join(dir, "auth.txt")
	if err := os.WriteFile(authFile, []byte("  spaced:value  \n"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLAW_WEB_AUTH_FILE", authFile)
	user, pass := authFromEnv()
	if user != "spaced" || pass != "value" {
		t.Errorf("expected trimmed credentials, got %s:%s", user, pass)
	}
}

func TestAuthFromEnv_FileNotFound(t *testing.T) {
	os.Unsetenv("CLAW_WEB_AUTH")
	t.Setenv("CLAW_WEB_AUTH_FILE", "/nonexistent/path/auth.txt")
	user, pass := authFromEnv()
	if user != "" || pass != "" {
		t.Errorf("expected empty when file missing, got %s:%s", user, pass)
	}
}

func TestAuthFromEnv_DirectTakesPrecedence(t *testing.T) {
	dir := t.TempDir()
	authFile := filepath.Join(dir, "auth.txt")
	if err := os.WriteFile(authFile, []byte("fileuser:filepass\n"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLAW_WEB_AUTH", "direct:secret")
	t.Setenv("CLAW_WEB_AUTH_FILE", authFile)
	user, pass := authFromEnv()
	// Direct env var takes precedence over file.
	if user != "direct" || pass != "secret" {
		t.Errorf("expected direct:secret, got %s:%s", user, pass)
	}
}

func TestBearerFromRequest_Header(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer tok123")
	tok := bearerFromRequest(r)
	if tok != "tok123" {
		t.Errorf("expected tok123, got %q", tok)
	}
}

func TestBearerFromRequest_QueryParam(t *testing.T) {
	r := httptest.NewRequest("GET", "/?token=qp456", nil)
	tok := bearerFromRequest(r)
	if tok != "qp456" {
		t.Errorf("expected qp456, got %q", tok)
	}
}

func TestBearerFromRequest_HeaderPreferred(t *testing.T) {
	r := httptest.NewRequest("GET", "/?token=query_token", nil)
	r.Header.Set("Authorization", "Bearer header_token")
	tok := bearerFromRequest(r)
	if tok != "header_token" {
		t.Errorf("expected header_token (header preferred), got %q", tok)
	}
}

func TestBearerFromRequest_None(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	tok := bearerFromRequest(r)
	if tok != "" {
		t.Errorf("expected empty, got %q", tok)
	}
}

func TestBearerFromRequest_EmptyQueryParam(t *testing.T) {
	r := httptest.NewRequest("GET", "/?token=", nil)
	tok := bearerFromRequest(r)
	if tok != "" {
		t.Errorf("expected empty for empty token query param, got %q", tok)
	}
}

func TestIsHealthPath(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"/healthz", true},
		{"/readyz", true},
		{"/metrics", true},
		{"/", false},
		{"/api/chat/ws", false},
		{"/static/app.js", false},
		{"/health", false},
		{"/healthz/extra", false},
	}
	for _, tt := range tests {
		got := isHealthPath(tt.path)
		if got != tt.want {
			t.Errorf("isHealthPath(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

// okHandler is a simple http.Handler that writes 200 OK.
type okHandler struct{}

func (h okHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func TestBasicAuthMiddleware_NoAuth(t *testing.T) {
	os.Unsetenv("CLAW_WEB_AUTH")
	os.Unsetenv("CLAW_WEB_AUTH_FILE")
	os.Unsetenv("CLAW_WEB_TOKEN")

	handler := basicAuthMiddleware(okHandler{})
	// Without auth configured, it should return the original handler.
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("expected 200 without auth, got %d", resp.StatusCode)
	}
}

func TestBasicAuthMiddleware_HealthBypass(t *testing.T) {
	t.Setenv("CLAW_WEB_AUTH", "admin:secret")

	handler := basicAuthMiddleware(okHandler{})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	for _, path := range []string{"/healthz", "/readyz", "/metrics"} {
		resp, err := http.Get(ts.URL + path)
		if err != nil {
			t.Errorf("%s: %v", path, err)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Errorf("%s: expected 200, got %d", path, resp.StatusCode)
		}
	}
}

func TestBasicAuthMiddleware_RequiresAuth(t *testing.T) {
	t.Setenv("CLAW_WEB_AUTH", "admin:secret")

	handler := basicAuthMiddleware(okHandler{})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
	// Should have WWW-Authenticate header.
	if wwwAuth := resp.Header.Get("WWW-Authenticate"); wwwAuth == "" {
		t.Error("expected WWW-Authenticate header")
	}
}

func TestBasicAuthMiddleware_ValidBasicAuth(t *testing.T) {
	t.Setenv("CLAW_WEB_AUTH", "admin:secret")

	handler := basicAuthMiddleware(okHandler{})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL+"/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth("admin", "secret")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestBasicAuthMiddleware_InvalidBasicAuth(t *testing.T) {
	t.Setenv("CLAW_WEB_AUTH", "admin:secret")

	handler := basicAuthMiddleware(okHandler{})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL+"/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth("admin", "wrongpass")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestBasicAuthMiddleware_InvalidUser(t *testing.T) {
	t.Setenv("CLAW_WEB_AUTH", "admin:secret")

	handler := basicAuthMiddleware(okHandler{})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL+"/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth("hacker", "secret")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestBasicAuthMiddleware_BearerToken(t *testing.T) {
	// Only bearer token configured, no basic auth.
	os.Unsetenv("CLAW_WEB_AUTH")
	os.Unsetenv("CLAW_WEB_AUTH_FILE")
	t.Setenv("CLAW_WEB_TOKEN", "mysecrettoken")

	handler := basicAuthMiddleware(okHandler{})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL+"/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer mysecrettoken")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("expected 200 with valid bearer token, got %d", resp.StatusCode)
	}
}

func TestBasicAuthMiddleware_BearerTokenQueryParam(t *testing.T) {
	os.Unsetenv("CLAW_WEB_AUTH")
	os.Unsetenv("CLAW_WEB_AUTH_FILE")
	t.Setenv("CLAW_WEB_TOKEN", "querytoken")

	handler := basicAuthMiddleware(okHandler{})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/?token=querytoken")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("expected 200 with valid query token, got %d", resp.StatusCode)
	}
}

func TestBasicAuthMiddleware_InvalidBearerToken(t *testing.T) {
	os.Unsetenv("CLAW_WEB_AUTH")
	os.Unsetenv("CLAW_WEB_AUTH_FILE")
	t.Setenv("CLAW_WEB_TOKEN", "mysecrettoken")

	handler := basicAuthMiddleware(okHandler{})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL+"/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer wrongtoken")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 with invalid bearer token, got %d", resp.StatusCode)
	}
}

func TestBasicAuthMiddleware_BothAuthMethods(t *testing.T) {
	// Both basic and bearer configured. Either should work.
	t.Setenv("CLAW_WEB_AUTH", "admin:secret")
	t.Setenv("CLAW_WEB_TOKEN", "mysecrettoken")

	handler := basicAuthMiddleware(okHandler{})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	// Bearer should work.
	req1, _ := http.NewRequest("GET", ts.URL+"/", nil)
	req1.Header.Set("Authorization", "Bearer mysecrettoken")
	resp1, err := http.DefaultClient.Do(req1)
	if err != nil {
		t.Fatal(err)
	}
	resp1.Body.Close()
	if resp1.StatusCode != 200 {
		t.Errorf("bearer auth: expected 200, got %d", resp1.StatusCode)
	}

	// Basic should also work.
	req2, _ := http.NewRequest("GET", ts.URL+"/", nil)
	req2.SetBasicAuth("admin", "secret")
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatal(err)
	}
	resp2.Body.Close()
	if resp2.StatusCode != 200 {
		t.Errorf("basic auth: expected 200, got %d", resp2.StatusCode)
	}
}

func TestBasicAuthMiddleware_AuthFile(t *testing.T) {
	os.Unsetenv("CLAW_WEB_AUTH")
	dir := t.TempDir()
	authFile := filepath.Join(dir, "auth.txt")
	if err := os.WriteFile(authFile, []byte("fileuser:filepass\n"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLAW_WEB_AUTH_FILE", authFile)
	os.Unsetenv("CLAW_WEB_TOKEN")

	handler := basicAuthMiddleware(okHandler{})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL+"/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth("fileuser", "filepass")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("expected 200 with file-based auth, got %d", resp.StatusCode)
	}
}