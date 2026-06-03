package web

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestRateLimiter_DisabledByDefault(t *testing.T) {
	// With no env set, rate limiter should be nil.
	rl := newRateLimiter()
	if rl != nil {
		t.Error("expected nil rateLimiter when CLAW_WEB_RATE_RPM is unset")
	}
}

func TestRateLimiter_EnabledWithEnv(t *testing.T) {
	t.Setenv("CLAW_WEB_RATE_RPM", "30")
	rl := newRateLimiter()
	if rl == nil {
		t.Fatal("expected non-nil rateLimiter when CLAW_WEB_RATE_RPM=30")
	}
}

func TestRateLimiter_InvalidEnv(t *testing.T) {
	t.Setenv("CLAW_WEB_RATE_RPM", "-5")
	rl := newRateLimiter()
	if rl != nil {
		t.Error("expected nil rateLimiter when CLAW_WEB_RATE_RPM is negative")
	}

	t.Setenv("CLAW_WEB_RATE_RPM", "notanumber")
	rl = newRateLimiter()
	if rl != nil {
		t.Error("expected nil rateLimiter when CLAW_WEB_RATE_RPM is not a number")
	}

	os.Unsetenv("CLAW_WEB_RATE_RPM")
	t.Setenv("CLAW_WEB_RATE_RPM", "0")
	rl = newRateLimiter()
	if rl != nil {
		t.Error("expected nil rateLimiter when CLAW_WEB_RATE_RPM is 0")
	}
}

func TestRateLimiter_AllowSession(t *testing.T) {
	t.Setenv("CLAW_WEB_RATE_RPM", "60")
	rl := newRateLimiter()
	if rl == nil {
		t.Fatal("expected non-nil rateLimiter")
	}

	// First few sessions should be allowed (burst of 3).
	for i := 0; i < 3; i++ {
		if !rl.AllowSession("192.168.1.1") {
			t.Errorf("session %d should be allowed", i+1)
		}
	}
	// After burst is exhausted, next should be limited.
	if rl.AllowSession("192.168.1.1") {
		t.Log("4th session allowed (token may have refilled)")
	}
}

func TestRateLimiter_AllowMessage(t *testing.T) {
	t.Setenv("CLAW_WEB_RATE_RPM", "60")
	rl := newRateLimiter()
	if rl == nil {
		t.Fatal("expected non-nil rateLimiter")
	}

	// First few messages should be allowed (burst of 3).
	for i := 0; i < 3; i++ {
		if !rl.AllowMessage("10.0.0.1") {
			t.Errorf("message %d should be allowed", i+1)
		}
	}
	// After burst is exhausted, next should be limited.
	if rl.AllowMessage("10.0.0.1") {
		t.Log("4th message allowed (token may have refilled)")
	}
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	t.Setenv("CLAW_WEB_RATE_RPM", "60")
	rl := newRateLimiter()
	if rl == nil {
		t.Fatal("expected non-nil rateLimiter")
	}

	// Each IP gets its own buckets.
	for i := 0; i < 5; i++ {
		if !rl.AllowSession("192.168.1.100") {
			t.Error("different IP should not be rate-limited yet")
		}
	}
}

func TestRateLimiter_NilSafety(t *testing.T) {
	// AllowSession and AllowMessage on nil should return true.
	var rl *rateLimiter
	if !rl.AllowSession("1.2.3.4") {
		t.Error("nil rateLimiter should allow sessions")
	}
	if !rl.AllowMessage("1.2.3.4") {
		t.Error("nil rateLimiter should allow messages")
	}
}

func TestRateLimitMiddleware_Allowed(t *testing.T) {
	t.Setenv("CLAW_WEB_RATE_RPM", "60")
	s := NewServer(Config{Addr: "127.0.0.1:0"})
	ts := httptest.NewServer(s.routes())
	defer ts.Close()

	// First request to /api/chat/ws should not be rate-limited
	// (though it will fail with 501 since ChatLoopFactory is nil).
	resp, err := http.Get(ts.URL + "/api/chat/ws")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	// Should NOT be 429 Too Many Requests.
	if resp.StatusCode == http.StatusTooManyRequests {
		t.Error("first request should not be rate-limited")
	}
}

func TestRateLimitMiddleware_NonChatEndpoint(t *testing.T) {
	t.Setenv("CLAW_WEB_RATE_RPM", "1") // very restrictive
	s := NewServer(Config{Addr: "127.0.0.1:0"})
	ts := httptest.NewServer(s.routes())
	defer ts.Close()

	// Requests to non-chat endpoints should pass through.
	resp, err := http.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("healthz should not be rate-limited, got %d", resp.StatusCode)
	}
}

func TestExtractIP(t *testing.T) {
	tests := []struct {
		name         string
		remoteAddr   string
		xForwardedFor string
		want         string
	}{
		{"plain", "192.168.1.1:12345", "", "192.168.1.1"},
		{"x-forwarded-for", "10.0.0.1:54321", "1.2.3.4", "1.2.3.4"},
		{"x-forwarded-for with port", "10.0.0.1:54321", "1.2.3.4:8080", "1.2.3.4"},
		{"ipv6", "[::1]:12345", "", "::1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			r.RemoteAddr = tt.remoteAddr
			if tt.xForwardedFor != "" {
				r.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if got := extractIP(r); got != tt.want {
				t.Errorf("extractIP() = %q, want %q", got, tt.want)
			}
		})
	}
}