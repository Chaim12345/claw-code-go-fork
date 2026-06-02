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
//
// If both CLAW_WEB_AUTH and CLAW_WEB_TOKEN are set, either method
// grants access (the first successful check wins).
func basicAuthMiddleware(next http.Handler) http.Handler {
	basicUser, basicPass := authFromEnv()
	bearerToken := os.Getenv("CLAW_WEB_TOKEN")

	// If no auth is configured at all, pass through.
	if basicUser == "" && bearerToken == "" {
		return next
	}

	expectedUser := []byte(basicUser)
	expectedPass := []byte(basicPass)
	expectedBearer := []byte(bearerToken)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow health/metrics endpoints without auth.
		if isHealthPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// 1. Try Bearer token (Authorization header or ?token= query param).
		if len(expectedBearer) > 0 {
			token := bearerFromRequest(r)
			if token != "" && subtle.ConstantTimeCompare([]byte(token), expectedBearer) == 1 {
				next.ServeHTTP(w, r)
				return
			}
		}

		// 2. Try Basic Auth.
		if basicUser != "" {
			reqUser, reqPass, ok := r.BasicAuth()
			if ok &&
				subtle.ConstantTimeCompare([]byte(reqUser), expectedUser) == 1 &&
				subtle.ConstantTimeCompare([]byte(reqPass), expectedPass) == 1 {
				next.ServeHTTP(w, r)
				return
			}
		}

		// If basic auth is configured, respond with a Basic challenge.
		if basicUser != "" {
			w.Header().Set("WWW-Authenticate", `Basic realm="claw-code-go", charset="UTF-8"`)
		}
		http.Error(w, "Authentication required", http.StatusUnauthorized)
	})
}

// bearerFromRequest extracts a bearer token from the Authorization
// header or the ?token= query parameter.
// Query-param tokens may leak via Referer; recommend HTTPS.
func bearerFromRequest(r *http.Request) string {
	// Authorization: Bearer <token>
	if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	// ?token=... query parameter
	if tok := r.URL.Query().Get("token"); tok != "" {
		return tok
	}
	return ""
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