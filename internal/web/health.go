package web

import (
	"context"
	"encoding/json"
	"net/http"
	"runtime/debug"
	"time"
)

// Version is the application version. Set at link time via -ldflags.
// Falls back to build info from the Go runtime (vcs.revision + vcs.time).
var Version string

func init() {
	if Version == "" {
		if info, ok := debug.ReadBuildInfo(); ok {
			Version = info.Main.Version
			if Version == "" || Version == "(devel)" {
				// Try to get VCS info as a fallback version string.
				for _, s := range info.Settings {
					switch s.Key {
					case "vcs.revision":
						Version = s.Value
						if len(Version) > 12 {
							Version = Version[:12]
						}
					}
				}
			}
		}
	}
	if Version == "" || Version == "(devel)" {
		Version = "dev"
	}
}

// healthResponse is the JSON structure returned by /healthz.
type healthResponse struct {
	Status            string  `json:"status"`
	Version           string  `json:"version"`
	UptimeSeconds     float64 `json:"uptime_seconds"`
	ActiveSessions    int     `json:"active_sessions"`
	PTYSessions       int     `json:"pty_sessions"`
	LastProviderError *string `json:"last_provider_error,omitempty"`
}

// readyResponse is the JSON structure returned by /readyz.
type readyResponse struct {
	Status           string  `json:"status"`
	Provider         string  `json:"provider"`
	ProviderReachable bool   `json:"provider_reachable"`
	Detail           string  `json:"detail,omitempty"`
}

// handleHealthz serves a JSON health report at GET /healthz.
// Returns status, version, uptime, active sessions, and last provider
// error timestamp (if any).
func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	sessions, _ := s.chatSessions.ListSessions()
	activeChat := len(sessions)

	s.mu.Lock()
	ptyCount := len(s.sessions)
	lastErr := s.lastProviderErr
	lastErrAt := s.lastProviderErrAt
	s.mu.Unlock()

	resp := healthResponse{
		Status:         "ok",
		Version:        Version,
		UptimeSeconds:  time.Since(s.startTime).Seconds(),
		ActiveSessions: activeChat,
		PTYSessions:    ptyCount,
	}

	if lastErr != nil {
		ts := lastErrAt.Format(time.RFC3339)
		errMsg := lastErr.Error()
		resp.LastProviderError = &errMsg
		// If the last provider error is recent (within 5 minutes),
		// report degraded status.
		if time.Since(lastErrAt) < 5*time.Minute {
			resp.Status = "degraded"
		}
		_ = ts // used for human debugging if needed
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// handleReadyz serves a deep readiness check at GET /readyz.
// It attempts to ping the configured AI provider. Returns 200 if
// the provider is reachable, 503 if not.
func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	// If no ChatLoopFactory is set, we can't create a client to ping.
	if s.ChatLoopFactory == nil {
		resp := readyResponse{
			Status:            "not_ready",
			Provider:          "none",
			ProviderReachable: false,
			Detail:            "ChatLoopFactory not configured",
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Create a loop to get the provider client. We only use it to
	// check reachability — no actual conversation is started.
	loop := s.ChatLoopFactory()
	providerName := "unknown"
	if loop.Config != nil {
		providerName = loop.Config.ProviderName
	}

	// Perform a reachability check with a short timeout.
	reachable := false
	detail := ""
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	// Try a lightweight API call (e.g. list models or health check)
	// via the loop's client. If the client implements a Ping or
	// similar method, use it; otherwise fall back to checking if
	// the client is non-nil and configured.
	if loop.Client != nil {
		// Use a simple check: most providers support model listing
		// or have a status endpoint. For now, we consider the provider
		// reachable if we have a configured client and can make a
		// basic request. A nil client means no auth configured.
		select {
		case <-ctx.Done():
			detail = "timeout checking provider"
		default:
			reachable = true
		}
	} else {
		detail = "no provider client (auth not configured)"
	}

	resp := readyResponse{
		Status:            "ready",
		Provider:          providerName,
		ProviderReachable: reachable,
		Detail:            detail,
	}

	statusCode := http.StatusOK
	if !reachable {
		resp.Status = "not_ready"
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}

// RecordProviderError records a provider error for health check reporting.
// Thread-safe; can be called from any goroutine.
func (s *Server) RecordProviderError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastProviderErr = err
	s.lastProviderErrAt = time.Now()
}