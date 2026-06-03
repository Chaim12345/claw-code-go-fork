# Changelog

All notable changes to claw-code-go will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added — Production-Ready Web UI

A major new feature: a **mobile-first, installable web UI** for the
`web` subcommand. Previously, `claw-code-go web` only streamed a PTY
terminal to the browser. Now it serves a full chat application
alongside the terminal mode.

#### Architecture

- **Structured WebSocket API** (`/api/chat/ws`): JSON protocol mirroring
  the internal `TurnEvent` stream — text deltas, tool calls, permission
  prompts, usage stats, and completion events. See
  [docs/web-protocol.md](docs/web-protocol.md) for the full wire format
  and Mermaid sequence diagram.
- **Chat UI** (`/`): Vanilla HTML/CSS/JS single-page app (no framework).
  Default landing page replaces the old terminal-only page.
- **Terminal mode** (`/terminal`): The existing PTY/wterm path remains
  available for power users who prefer the TUI experience.
- Both modes share the same `ConversationLoop`, auth, provider, and model
  resolution.

#### Chat UI Features

- **Mobile-first responsive layout**: CSS Grid with breakpoints at 768px
  and 1024px. Safe area insets for notched phones. `visualViewport` API
  for keyboard-aware layout. Auto-growing textarea composer with
  Enter-to-send and Shift+Enter for newlines. 44px minimum touch targets.
- **Streaming message rendering**: Server `text_delta` events group into
  message bubbles with `aria-live="polite"`. Code blocks in `<pre><code>`
  with Prism.js syntax highlighting (go, python, bash, json, html, css,
  typescript).
- **Tool-call cards**: Collapsible cards for `tool_start`/`tool_done`
  events with tool name, input summary, result details, duration timer,
  and copy button.
- **Permission prompts**: Modal dialogs with Allow once / Always allow /
  Deny buttons. On mobile, renders as a bottom sheet filling 60% of the
  viewport.
- **Slash command menu**: Typing `/` opens a filtered command list from
  `/api/commands`. Arrow keys / tap to select, Enter to insert.
- **Session history sidebar**: Lists past sessions from `/api/sessions`.
  Slide-in drawer on mobile, persistent column on desktop.
- **Reconnection logic**: Exponential backoff (1s→2s→4s→max 30s) on
  WebSocket close. Per-message UUID tracking with server ack to avoid
  duplicate sends on reconnect.
- **Suggested follow-up chips**: Tappable chip buttons with follow-up
  questions after each AI response (Perplexity pattern).
- **Theme switcher**: System / light / dark modes persisted in
  `localStorage`. Font size control (S/M/L, 12-18px range).
- **Keyboard shortcuts overlay**: `?` toggles a shortcuts panel; "?"
  button in the header for touch users.
- **Code block actions**: Copy, download-as-file, and "run in terminal"
  buttons on every code block.
- **Empty states**: Friendly placeholder for no messages, "thinking…"
  indicator with elapsed time, and rate-limit countdown banner.

#### PWA (Progressive Web App)

- **Web App Manifest**: `name`, `short_name`, `start_url`, `display:
  standalone`, `theme_color`, `background_color`, and 192px/512px/180px
  (apple-touch-icon) icons.
- **Service Worker**: Cache-first for `/static/*` assets, network-first
  for `/api/*` and `/ws`, offline fallback page with retry button.
- **iOS support**: `apple-mobile-web-app-capable` meta tag, status bar
  style, one-time "Add to Home Screen" instruction banner for Safari.
- **Install verification**: Documented flow for desktop Chrome, Android
  Chrome (WebAPK), and iOS Safari in [docs/web-deployment.md](docs/web-deployment.md).

#### Security

- **HTTP Basic Auth**: `CLAW_WEB_AUTH=user:pass` or
  `CLAW_WEB_AUTH_FILE=path`. Constant-time comparison. Protects all
  routes except `/healthz`, `/readyz`, `/metrics`, and `/error`.
- **Bearer token auth**: `CLAW_WEB_TOKEN=<secret>`. Accepted via
  `Authorization: Bearer ...` header or `?token=...` query param.
- **CSP and security headers**: `Content-Security-Policy` (including
  `'wasm-unsafe-eval'` for wterm), `X-Content-Type-Options: nosniff`,
  `Referrer-Policy: no-referrer`, `X-Frame-Options: DENY`,
  `Permissions-Policy` (no camera, microphone, geolocation).
- **Rate limiting**: Per-IP token bucket for the chat WebSocket
  endpoint (5 new sessions/min, 60 messages/min). Configurable via
  `CLAW_WEB_RATE_RPM`.
- **Static asset caching**: `Cache-Control` headers with hashed subdirs.

#### Observability

- **Structured logging**: `log/slog` with `request_id`, `remote_addr`,
  `user_agent` on every WebSocket connection. Lifecycle events at INFO,
  protocol violations and rate-limit hits at WARN.
- **Prometheus metrics** at `/metrics`:
  `claw_web_sessions_total{provider,model}`,
  `claw_web_active_sessions`, `claw_web_messages_total{role,kind}`,
  `claw_web_request_duration_seconds`, `claw_web_errors_total{kind}`.
- **Health checks**: `/healthz` returns JSON with status, version, uptime,
  active sessions, last provider error timestamp. `/readyz` pings the
  configured AI provider for a deep health check.
- **Error page**: `/error?code=...&msg=...` with friendly UI and
  copy-paste correlation ID.

#### Documentation

- **[docs/web-protocol.md](docs/web-protocol.md)**: Wire format spec with
  Mermaid sequence diagram.
- **[docs/web-deployment.md](docs/web-deployment.md)**: Caddy and nginx
  reverse-proxy configs, TLS termination, systemd unit, firewall rules,
  PWA install verification, and Prometheus metrics reference.
- **[WEB_RESEARCH.md](WEB_RESEARCH.md)**: Research notes from responsive
  design, PWA, mobile chat UI patterns, and AI SDK best practices.
- **README.md**: New "Web UI" section with quick start, auth config,
  architecture diagram, and mobile install instructions.

#### Tests

- **Unit tests**: `auth_test.go`, `ratelimit_test.go`,
  `security_test.go`, `chatproto_test.go` — ≥80% coverage on all new
  packages.