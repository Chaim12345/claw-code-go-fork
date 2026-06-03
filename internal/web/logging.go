package web

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"os"
	"time"
)

// ctxKey is an unexported type used for context keys to avoid
// collisions with keys defined in other packages.
type ctxKey string

const loggerCtxKey ctxKey = "slog-logger"

// newRequestID generates a short random hex string for use as a
// request_id in structured logs. Falls back to a timestamp if
// crypto/rand fails (should never happen on Linux).
func newRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format("20060102T150405.000")))
	}
	return hex.EncodeToString(b)
}

// requestLogger returns a *slog.Logger with context attributes from
// the HTTP request. The returned logger has request_id, remote_addr,
// and user_agent pre-populated. It is also stored in the request
// context so downstream handlers can retrieve it via LoggerFromCtx.
//
// Uses the default slog logger (configured via slog.SetDefault or
// CLAW_WEB_LOG_FORMAT/CLAW_WEB_LOG_LEVEL env vars in SetupLogging).
func requestLogger(r *http.Request) (*slog.Logger, *http.Request) {
	requestID := newRequestID()
	remoteAddr := extractIP(r)
	userAgent := r.UserAgent()

	logger := slog.Default().With(
		slog.String("request_id", requestID),
		slog.String("remote_addr", remoteAddr),
		slog.String("user_agent", userAgent),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
	)

	ctx := context.WithValue(r.Context(), loggerCtxKey, logger)
	return logger, r.WithContext(ctx)
}

// LoggerFromCtx extracts the *slog.Logger that was stored in the
// context by the logging middleware. Returns slog.Default() if no
// logger was found (should not happen if the middleware is applied).
func LoggerFromCtx(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerCtxKey).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

// loggingMiddleware injects a structured logger into every request.
// It generates a request_id and attaches remote_addr and user_agent
// attributes. The logger is available to all downstream handlers via
// LoggerFromCtx(r.Context()).
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger, r2 := requestLogger(r)
		logger.LogAttrs(r.Context(), slog.LevelInfo, "request_start")
		next.ServeHTTP(w, r2)
	})
}

// SetupLogging configures the default slog logger. It reads
// CLAW_WEB_LOG_FORMAT (json or text) and CLAW_WEB_LOG_LEVEL
// (debug, info, warn, error) to set the output format and level.
// Call this once during server startup (not during init).
func SetupLogging() {
	format := os.Getenv("CLAW_WEB_LOG_FORMAT")
	level := slog.LevelInfo
	if lv := os.Getenv("CLAW_WEB_LOG_LEVEL"); lv != "" {
		switch lv {
		case "debug":
			level = slog.LevelDebug
		case "warn":
			level = slog.LevelWarn
		case "error":
			level = slog.LevelError
		}
	}

	var handler slog.Handler
	if format == "json" {
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	} else {
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	}
	slog.SetDefault(slog.New(handler))
}