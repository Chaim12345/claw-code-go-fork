// Package web serves the claw-code-go TUI over HTTP and WebSocket using
// a PTY bridge. The browser loads the wterm terminal emulator (Zig WASM)
// and pipes raw PTY bytes both directions.
//
// Architecture:
//
//	browser ──wterm WASM──> WebSocket ──bytes──> PTY ──> claw-code-go --repl
//	browser <──wterm WASM──< WebSocket <──bytes──< PTY <── claw-code-go --repl
//
// This is a thin "ttyd-style" wrapper. The existing Bubble Tea TUI runs
// unmodified inside the PTY; the browser is just a remote terminal.
package web

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"claw-code-go/internal/runtime"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
)

//go:embed static
var staticFS embed.FS

func init() {
	// Register the web manifest MIME type so browsers parse it correctly.
	mime.AddExtensionType(".webmanifest", "application/manifest+json")
}

// Config configures the web server.
type Config struct {
	// Addr is the listen address, e.g. "127.0.0.1:7777" or ":7777".
	Addr string

	// BinaryPath is the path to the claw-code-go binary to spawn inside
	// the PTY. Defaults to os.Args[0] (the currently running binary).
	BinaryPath string

	// Workdir is the cwd the spawned TUI sees. Empty inherits.
	Workdir string

	// Args are extra flags passed to the spawned TUI.
	// `--repl` is always added.
	Args []string

	// Env are extra environment variables for the spawned TUI.
	// Inherits the parent process env by default.
	Env []string
}

// Server is the HTTP+WS server that hosts the web UI.
type Server struct {
	cfg     Config
	upgrader websocket.Upgrader

	// ChatLoopFactory, if set, builds a *runtime.ConversationLoop for
	// each /api/chat/ws connection. If nil, the endpoint returns 501
	// Not Implemented. Set by the caller after constructing the
	// runtime loop (e.g. in cmd/claw-code-go/main.go).
	ChatLoopFactory func() *runtime.ConversationLoop

	rateLimiter *rateLimiter

	// chatSessions tracks metadata for chat WS sessions (the /api/chat/ws
	// path) so the sidebar can show session history.
	chatSessions *chatSessionStore

	mu        sync.Mutex
	sessions  map[*ptySession]struct{}
}

// NewServer returns a Server ready to listen on cfg.Addr.
func NewServer(cfg Config) *Server {
	if cfg.Addr == "" {
		cfg.Addr = "127.0.0.1:7777"
	}
	if cfg.BinaryPath == "" {
		cfg.BinaryPath = os.Args[0]
	}
	return &Server{
		cfg:          cfg,
		upgrader:     websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		rateLimiter:  newRateLimiter(),
		chatSessions: newChatSessionStore(),
		sessions:     make(map[*ptySession]struct{}),
	}
}

// ListenAndServe starts the HTTP server. Blocks until ctx is done or
// the server returns a non-nil error.
func (s *Server) ListenAndServe(ctx context.Context) error {
	srv := &http.Server{
		Addr:              s.cfg.Addr,
		Handler:           s.routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Shutdown on context cancel.
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
		s.killAll()
	}()

	fmt.Fprintf(os.Stderr, "[web] listening on http://%s\n", s.cfg.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// routes returns the HTTP mux wired up with all endpoints. Exposed so
// tests can exercise the routing without binding a real socket.
func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	// Serve the static frontend from the embedded FS.
	staticSub, err := fs.Sub(staticFS, "static")
	if err != nil {
		// Static is required for the page to load; panic at startup
		// rather than serving a broken UI silently.
		panic(fmt.Errorf("embed sub: %w", err))
	}
	mux.Handle("/static/", staticCacheHandler(staticSub))
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/ws", s.handleWS)
	mux.HandleFunc("/api/chat/ws", s.handleChatWS)
	mux.HandleFunc("/api/sessions", s.chatSessions.handleSessions)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	// Safe area insets visual test page for mobile development.
	mux.HandleFunc("/safe-area-test", s.handleSafeAreaTest)
	return s.rateLimiter.middleware(securityHeadersMiddleware(basicAuthMiddleware(mux)))
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	data, err := staticFS.ReadFile("static/chat.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(data)
}

// handleSafeAreaTest serves the visual safe-area insets test page
// for mobile development and debugging on notched/rounded devices.
func (s *Server) handleSafeAreaTest(w http.ResponseWriter, r *http.Request) {
	data, err := staticFS.ReadFile("static/safe-area-test.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(data)
}

// handleWS upgrades the HTTP request to a WebSocket, spawns a PTY
// running the TUI binary, and bridges bytes both directions.
func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[web] ws upgrade: %v\n", err)
		return
	}
	conn.SetReadLimit(64 * 1024) // 64 KiB; PTY input bursts are small

	sess, err := s.spawnPTY(r)
	if err != nil {
		_ = conn.WriteMessage(websocket.TextMessage, []byte("\r\n\x1b[31m[web] failed to start TUI: "+err.Error()+"\x1b[0m\r\n"))
		_ = conn.Close()
		return
	}
	s.register(sess)
	defer s.unregister(sess)

	// PTY -> WS (binary frames so the browser can do term.write(Uint8Array))
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := sess.ptmx.Read(buf)
			if n > 0 {
				if werr := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); werr != nil {
					return
				}
			}
			if err != nil {
				if !errors.Is(err, io.EOF) {
					fmt.Fprintf(os.Stderr, "[web] pty read: %v\n", err)
				}
				return
			}
		}
	}()

	// WS -> PTY. Two message types:
	//   - text JSON  {type:"resize", cols:N, rows:N} → TIOCSWINSZ on the PTY
	//   - text/binary raw bytes                     → written to PTY stdin
	for {
		mt, data, err := conn.ReadMessage()
		if err != nil {
			break
		}
		if mt == websocket.TextMessage && len(data) > 0 && data[0] == '{' {
			if handled, rerr := sess.handleControlMessage(data); handled {
				if rerr != nil {
					fmt.Fprintf(os.Stderr, "[web] control: %v\n", rerr)
				}
				continue
			}
		}
		if _, err := sess.ptmx.Write(data); err != nil {
			fmt.Fprintf(os.Stderr, "[web] pty write: %v\n", err)
			break
		}
	}

	// Reader hung up; tear down the PTY.
	sess.close()
}

// handleChatWS upgrades the HTTP request to a WebSocket and runs the
// structured chat protocol via a *runtime.ConversationLoop.
// Requires ChatLoopFactory to be set on the Server.
func (s *Server) handleChatWS(w http.ResponseWriter, r *http.Request) {
	if s.ChatLoopFactory == nil {
		http.Error(w, "chat mode not available (set ChatLoopFactory)", http.StatusNotImplemented)
		return
	}
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[web] chat ws upgrade: %v\n", err)
		return
	}
	loop := s.ChatLoopFactory()
	sessionID := newSessionID()
	s.chatSessions.create(sessionID)
	ctx := r.Context()
	runChatSession(ctx, conn, loop, sessionID, s.chatSessions)
	conn.Close()
}

// ptySession owns a single PTY + child process.
type ptySession struct {
	ptmx   *os.File
	cmd    *exec.Cmd
	cancel context.CancelFunc
}

func (s *Server) spawnPTY(r *http.Request) (*ptySession, error) {
	ctx, cancel := context.WithCancel(context.Background())

	args := append([]string{"--repl"}, s.cfg.Args...)
	cmd := exec.CommandContext(ctx, s.cfg.BinaryPath, args...)
	cmd.Env = append(os.Environ(), s.cfg.Env...)
	if s.cfg.Workdir != "" {
		cmd.Dir = s.cfg.Workdir
	} else if wd, err := os.Getwd(); err == nil {
		cmd.Dir = wd
	}
	// The child should think it's connected to a real terminal so
	// Bubble Tea doesn't refuse to render the alt-screen.
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	// Default 80x24 — client resizes once it has measured the viewport.
	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{Rows: 24, Cols: 80})
	if err != nil {
		cancel()
		return nil, fmt.Errorf("pty start: %w", err)
	}

	return &ptySession{ptmx: ptmx, cmd: cmd, cancel: cancel}, nil
}

// handleControlMessage tries to interpret a text frame as a JSON
// control message. Returns handled=true even on error so the caller
// doesn't write the JSON to the PTY as raw keystrokes.
func (s *ptySession) handleControlMessage(data []byte) (bool, error) {
	var msg struct {
		Type string `json:"type"`
		Cols int    `json:"cols"`
		Rows int    `json:"rows"`
	}
	if err := jsonUnmarshal(data, &msg); err != nil {
		return true, err
	}
	if msg.Type != "resize" {
		return true, nil // unknown, drop silently
	}
	if msg.Cols < 2 || msg.Rows < 2 || msg.Cols > 500 || msg.Rows > 500 {
		return true, fmt.Errorf("resize out of range: %dx%d", msg.Cols, msg.Rows)
	}
	return true, pty.Setsize(s.ptmx, &pty.Winsize{Rows: uint16(msg.Rows), Cols: uint16(msg.Cols)})
}

func (s *ptySession) close() {
	s.cancel()
	_ = s.ptmx.Close()
	if s.cmd.Process != nil {
		_ = s.cmd.Process.Signal(syscall.SIGTERM)
		// Give it a moment, then SIGKILL.
		done := make(chan struct{})
		go func() { _ = s.cmd.Wait(); close(done) }()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			_ = s.cmd.Process.Kill()
			<-done
		}
	}
}

func (s *Server) register(sess *ptySession) {
	s.mu.Lock()
	s.sessions[sess] = struct{}{}
	s.mu.Unlock()
}

func (s *Server) unregister(sess *ptySession) {
	s.mu.Lock()
	delete(s.sessions, sess)
	s.mu.Unlock()
}

func (s *Server) killAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for sess := range s.sessions {
		sess.close()
	}
	s.sessions = map[*ptySession]struct{}{}
}

// newSessionID returns a short random hex string suitable for
// identifying a chat session in logs and to the client.
func newSessionID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand.Read is documented to always return nil on
		// Linux; fall back to a timestamp as a last resort.
		return fmt.Sprintf("sess_%d", time.Now().UnixNano())
	}
	return "sess_" + hex.EncodeToString(b)
}

// staticCacheHandler serves files from the embedded filesystem fsys
// under the /static/ prefix. It sets Cache-Control: public, max-age=3600
// for all static assets (JS, CSS, WASM, etc.) since they are served
// from content-hashed subdirectories like @wterm/.
func staticCacheHandler(fsys fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(fsys))
	return http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set long-lived cache for all static assets.
		w.Header().Set("Cache-Control", "public, max-age=3600")
		// Ensure .webmanifest gets application/manifest+json content type.
		if strings.HasSuffix(r.URL.Path, ".webmanifest") {
			w.Header().Set("Content-Type", "application/manifest+json")
		}
		fileServer.ServeHTTP(w, r)
	}))
}
