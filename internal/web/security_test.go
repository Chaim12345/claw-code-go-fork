package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeadersMiddleware(t *testing.T) {
	handler := securityHeadersMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// Verify Content-Security-Policy is set and contains key directives.
	csp := resp.Header.Get("Content-Security-Policy")
	if csp == "" {
		t.Error("expected Content-Security-Policy header")
	}
	cspContains := func(directive string) bool {
		return len(csp) > 0 && containsStr(csp, directive)
	}
	if !cspContains("default-src 'self'") {
		t.Errorf("CSP missing default-src 'self': %s", csp)
	}
	if !cspContains("script-src 'self' 'wasm-unsafe-eval'") {
		t.Errorf("CSP missing wasm-unsafe-eval: %s", csp)
	}
	if !cspContains("style-src 'self' 'unsafe-inline'") {
		t.Errorf("CSP missing style-src: %s", csp)
	}
	if !cspContains("connect-src 'self' ws: wss:") {
		t.Errorf("CSP missing connect-src ws: wss:: %s", csp)
	}
	if !cspContains("img-src 'self' data:") {
		t.Errorf("CSP missing img-src data:: %s", csp)
	}
	if !cspContains("font-src 'self'") {
		t.Errorf("CSP missing font-src: %s", csp)
	}
	if !cspContains("manifest-src 'self'") {
		t.Errorf("CSP missing manifest-src: %s", csp)
	}
	if !cspContains("frame-ancestors 'none'") {
		t.Errorf("CSP missing frame-ancestors 'none': %s", csp)
	}
	if !cspContains("base-uri 'self'") {
		t.Errorf("CSP missing base-uri: %s", csp)
	}
	if !cspContains("form-action 'self'") {
		t.Errorf("CSP missing form-action: %s", csp)
	}

	// X-Content-Type-Options
	xcto := resp.Header.Get("X-Content-Type-Options")
	if xcto != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q, want nosniff", xcto)
	}

	// Referrer-Policy
	rp := resp.Header.Get("Referrer-Policy")
	if rp != "no-referrer" {
		t.Errorf("Referrer-Policy = %q, want no-referrer", rp)
	}

	// X-Frame-Options
	xfo := resp.Header.Get("X-Frame-Options")
	if xfo != "DENY" {
		t.Errorf("X-Frame-Options = %q, want DENY", xfo)
	}

	// Permissions-Policy
	pp := resp.Header.Get("Permissions-Policy")
	if pp == "" {
		t.Error("expected Permissions-Policy header")
	}
	if !containsStr(pp, "camera=()") {
		t.Errorf("Permissions-Policy missing camera=(): %s", pp)
	}
	if !containsStr(pp, "microphone=()") {
		t.Errorf("Permissions-Policy missing microphone=(): %s", pp)
	}
	if !containsStr(pp, "geolocation=()") {
		t.Errorf("Permissions-Policy missing geolocation=(): %s", pp)
	}
}

func TestSecurityHeadersMiddleware_HeadersPerRequest(t *testing.T) {
	// Each request gets fresh headers (not cached from previous).
	handler := securityHeadersMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	ts := httptest.NewServer(handler)
	defer ts.Close()

	for i := 0; i < 3; i++ {
		resp, err := http.Get(ts.URL + "/")
		if err != nil {
			t.Fatalf("request %d: %v", i, err)
		}
		if csp := resp.Header.Get("Content-Security-Policy"); csp == "" {
			t.Errorf("request %d: missing CSP", i)
		}
		resp.Body.Close()
	}
}

// containsStr reports whether s contains substr.
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && hasSubstr(s, substr)
}

func hasSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestContainsStr(t *testing.T) {
	if !containsStr("hello world", "world") {
		t.Error("containsStr broken")
	}
	if containsStr("hello", "xxx") {
		t.Error("containsStr broken: false positive")
	}
}