package web

import (
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// rateLimiter is a per-IP token bucket rate limiter. It limits the
// rate of new WebSocket sessions to /api/chat/ws and the rate of
// messages per session.
//
// Configuration is read from CLAW_WEB_RATE_RPM (requests per minute
// per IP). If unset or ≤ 0, rate limiting is disabled.
type rateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor

	// sessionRate is the per-IP rate for new chat sessions.
	sessionRate rate.Limit
	// messageRate is the per-IP rate for individual messages within a
	// session. Each message consumes a token from the message bucket.
	messageRate rate.Limit
}

// visitor holds the token buckets for a single IP address.
type visitor struct {
	sessions *rate.Limiter // 5 new sessions per minute by default
	messages *rate.Limiter // 60 messages per minute by default
	lastSeen time.Time
}

// newRateLimiter creates a rate limiter from the CLAW_WEB_RATE_RPM
// environment variable. The value is the requests-per-minute ceiling.
// Defaults: 5 sessions/min, 60 messages/min.
//
// Set CLAW_WEB_RATE_RPM=0 or leave unset to disable rate limiting.
func newRateLimiter() *rateLimiter {
	rpm := rateRPMFromEnv()
	if rpm <= 0 {
		return nil // disabled
	}

	// sessionRate: 5 new sessions per minute (per IP)
	sessionPerMin := 5.0
	// messageRate: 60 messages per minute (per IP)
	messagePerMin := 60.0

	// If CLAW_WEB_RATE_RPM is set, use it as the message rate and
	// keep session rate at roughly 1/12th of that.
	if rpm > 0 {
		messagePerMin = float64(rpm)
		sessionPerMin = float64(rpm) / 12.0
		if sessionPerMin < 1 {
			sessionPerMin = 1
		}
	}

	rl := &rateLimiter{
		visitors:    make(map[string]*visitor),
		sessionRate: rate.Limit(sessionPerMin / 60.0),
		messageRate: rate.Limit(messagePerMin / 60.0),
	}

	// Background cleanup of stale visitors every 10 minutes.
	go rl.cleanupLoop()

	return rl
}

// rateRPMFromEnv reads CLAW_WEB_RATE_RPM. Returns 0 if unset,
// invalid, or ≤ 0.
func rateRPMFromEnv() int {
	v := os.Getenv("CLAW_WEB_RATE_RPM")
	if v == "" {
		return 0
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return 0
	}
	return n
}

// getVisitor returns the visitor for ip, creating one if needed.
func (rl *rateLimiter) getVisitor(ip string) *visitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, ok := rl.visitors[ip]
	if !ok {
		burst := 3
		v = &visitor{
			sessions: rate.NewLimiter(rl.sessionRate, burst),
			messages: rate.NewLimiter(rl.messageRate, burst),
		}
		rl.visitors[ip] = v
	}
	v.lastSeen = time.Now()
	return v
}

// AllowSession reports whether a new session is allowed for ip.
func (rl *rateLimiter) AllowSession(ip string) bool {
	if rl == nil {
		return true
	}
	return rl.getVisitor(ip).sessions.Allow()
}

// AllowMessage reports whether a new message is allowed for ip.
func (rl *rateLimiter) AllowMessage(ip string) bool {
	if rl == nil {
		return true
	}
	return rl.getVisitor(ip).messages.Allow()
}

// cleanupLoop periodically removes visitors that haven't been seen
// for 15 minutes.
func (rl *rateLimiter) cleanupLoop() {
	t := time.NewTicker(10 * time.Minute)
	defer t.Stop()
	for range t.C {
		rl.cleanup()
	}
}

func (rl *rateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	for ip, v := range rl.visitors {
		if time.Since(v.lastSeen) > 15*time.Minute {
			delete(rl.visitors, ip)
		}
	}
}

// extractIP returns the client IP address from the request. It
// respects X-Forwarded-For when present (trusts the first entry).
func extractIP(r *http.Request) string {
	// Check X-Forwarded-For header first.
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		// Take the first IP in the chain.
		ip, _, _ := net.SplitHostPort(fwd)
		if ip == "" {
			ip = fwd
		}
		return ip
	}
	// Fall back to RemoteAddr.
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// rateLimitMiddleware returns an HTTP middleware that applies rate
// limiting to the /api/chat/ws endpoint. Session rate is checked
// before the WebSocket upgrade; message rate is checked inside
// runChatSession (see chat_session.go).
func (rl *rateLimiter) middleware(next http.Handler) http.Handler {
	if rl == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only rate-limit the chat WS endpoint.
		if r.URL.Path != "/api/chat/ws" {
			next.ServeHTTP(w, r)
			return
		}
		ip := extractIP(r)
		if !rl.AllowSession(ip) {
			http.Error(w, "rate limit exceeded; try again later", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}