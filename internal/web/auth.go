package web

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"
)

// authFromEnv reads CLAW_WEB_AUTH (format "user:pass") or
// CLAW_WEB_AUTH_FILE (path to a file containing "user:pass").
// Returns the username and password, or empty strings if no auth
// is configured.
func authFromEnv() (user, pass string) {
	raw := os.Getenv("CLAW_WEB_AUTH")
	if raw == "" {
		if path := os.Getenv("CLAW_WEB_AUTH_FILE"); path != "" {
			data, err := os.ReadFile(path)
			if err != nil {
				return "", ""
			}
			raw = strings.TrimSpace(string(data))
		}
	}
	if raw == "" {
		return "", ""
	}
	user, pass, ok := strings.Cut(raw, ":")
	if !ok {
		return "", ""
	}
	return user, pass
}

// basicAuthMiddleware returns an HTTP middleware that requires HTTP
// Basic Auth when CLAW_WEB_AUTH (or CLAW_WEB_AUTH_FILE) is set.
// Health endpoints (/healthz, /readyz, /metrics) are excluded.
// Uses constant-time comparison to prevent timing attacks.
func basicAuthMiddleware(next http.Handler) http.Handler {
	user, pass := authFromEnv()
	if user == "" {
		// No auth configured — pass through.
		return next
	}

	expectedUser := []byte(user)
	expectedPass := []byte(pass)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow health/metrics endpoints without auth.
		if isHealthPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		reqUser, reqPass, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="claw-code-go", charset="UTF-8"`)
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		// Constant-time comparison.
		if subtle.ConstantTimeCompare([]byte(reqUser), expectedUser) != 1 ||
			subtle.ConstantTimeCompare([]byte(reqPass), expectedPass) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="claw-code-go", charset="UTF-8"`)
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// isHealthPath returns true for paths that should be accessible
// without authentication (health checks, metrics).
func isHealthPath(path string) bool {
	switch path {
	case "/healthz", "/readyz", "/metrics":
		return true
	}
	return false
}