package web

import (
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	metricsOnce sync.Once
	m           *Metrics
)

// Metrics holds all Prometheus metrics for the web server.
// It is safe for concurrent use (prometheus client_golang handles
// synchronization internally).
type Metrics struct {
	// SessionsTotal counts chat sessions created, labeled by provider and model.
	SessionsTotal *prometheus.CounterVec

	// ActiveSessions tracks the current number of active chat sessions.
	ActiveSessions prometheus.Gauge

	// MessagesTotal counts messages sent and received, labeled by role
	// (user, assistant) and kind (text, tool_call, tool_result, error).
	MessagesTotal *prometheus.CounterVec

	// RequestDurationSeconds is a histogram of HTTP request durations
	// (excluding WebSocket upgrades, which are long-lived).
	RequestDurationSeconds *prometheus.HistogramVec

	// ErrorsTotal counts errors by kind (parse_error, protocol_violation,
	// rate_limit, auth_failure, server_error).
	ErrorsTotal *prometheus.CounterVec
}

// NewMetrics creates and registers all Prometheus metrics with the
// default registry. Safe to call multiple times (uses sync.Once).
func NewMetrics() *Metrics {
	metricsOnce.Do(func() {
		m = &Metrics{
			SessionsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "claw_web",
				Name:      "sessions_total",
				Help:      "Total number of chat sessions created.",
			}, []string{"provider", "model"}),

			ActiveSessions: prometheus.NewGauge(prometheus.GaugeOpts{
				Namespace: "claw_web",
				Name:      "active_sessions",
				Help:      "Current number of active chat sessions.",
			}),

			MessagesTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "claw_web",
				Name:      "messages_total",
				Help:      "Total number of messages processed.",
			}, []string{"role", "kind"}),

			RequestDurationSeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: "claw_web",
				Name:      "request_duration_seconds",
				Help:      "HTTP request duration in seconds (excluding WebSocket connections).",
				Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
			}, []string{"method", "path"}),

			ErrorsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "claw_web",
				Name:      "errors_total",
				Help:      "Total number of errors by kind.",
			}, []string{"kind"}),
		}

		prometheus.MustRegister(
			m.SessionsTotal,
			m.ActiveSessions,
			m.MessagesTotal,
			m.RequestDurationSeconds,
			m.ErrorsTotal,
		)
	})
	return m
}

// Handler returns an http.Handler that exposes the registered metrics
// in Prometheus text format at GET /metrics.
func (m *Metrics) Handler() http.Handler {
	return promhttp.Handler()
}

// SessionStarted records a new session with the given provider and model.
func (m *Metrics) SessionStarted(provider, model string) {
	m.SessionsTotal.WithLabelValues(provider, model).Inc()
	m.ActiveSessions.Inc()
}

// SessionEnded decrements the active session gauge.
func (m *Metrics) SessionEnded() {
	m.ActiveSessions.Dec()
}

// MessageCounted records a message for the given role and kind.
// role is "user" or "assistant"; kind is "text", "tool_call",
// "tool_result", or "error".
func (m *Metrics) MessageCounted(role, kind string) {
	m.MessagesTotal.WithLabelValues(role, kind).Inc()
}

// RequestDuration records the duration of an HTTP request.
func (m *Metrics) ObserveRequestDuration(method, path string, d time.Duration) {
	m.RequestDurationSeconds.WithLabelValues(method, path).Observe(d.Seconds())
}

// ErrorCounted records an error by kind.
func (m *Metrics) ErrorCounted(kind string) {
	m.ErrorsTotal.WithLabelValues(kind).Inc()
}

// metricsMiddleware wraps an http.Handler to record request duration
// for non-WebSocket requests.
func (m *Metrics) middleware(next http.Handler) http.Handler {
	if m == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't time WebSocket upgrades — they are long-lived connections.
		if r.Header.Get("Upgrade") != "" {
			next.ServeHTTP(w, r)
			return
		}
		start := time.Now()
		// Use a response writer wrapper to capture the status code.
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		m.ObserveRequestDuration(r.Method, routePattern(r.URL.Path), time.Since(start))
		if sw.status >= 400 {
			kind := statusErrorKind(sw.status)
			m.ErrorCounted(kind)
		}
	})
}

// statusWriter wraps http.ResponseWriter to capture the status code.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}

// routePattern maps URL paths to a coarser pattern for metric labels,
// avoiding unbounded cardinality from session IDs or query strings.
func routePattern(path string) string {
	switch {
	case path == "/":
		return "/"
	case path == "/healthz":
		return "/healthz"
	case path == "/readyz":
		return "/readyz"
	case path == "/ws":
		return "/ws"
	case path == "/api/chat/ws":
		return "/api/chat/ws"
	case path == "/metrics":
		return "/metrics"
	case path == "/api/sessions":
		return "/api/sessions"
	case path == "/safe-area-test":
		return "/safe-area-test"
	case path == "/error":
		return "/error"
	case len(path) > 8 && path[:8] == "/static/":
		return "/static/*"
	default:
		return "/other"
	}
}

// statusErrorKind maps HTTP status codes to error kind labels.
func statusErrorKind(code int) string {
	switch {
	case code == 401 || code == 403:
		return "auth_failure"
	case code == 429:
		return "rate_limit"
	case code >= 500:
		return "server_error"
	default:
		return "client_error"
	}
}