# ROADMAP: Production-Ready, Mobile-Friendly Web UI for claw-code-go

## Context

claw-code-go currently ships a `web` subcommand that wraps the Bubble
Tea TUI in a PTY and serves it to the browser via wterm (Zig/WASM
terminal emulator). The architecture is a "ttyd-style" bridge — it
works, but the user experience is a terminal in a browser, not a
proper web product.

This roadmap transforms the web subcommand into a **production-ready,
mobile-friendly web UI** that runs alongside (or replaces) the TUI
mode.

## Working principles

1. **Each iteration is ONE item.** Read the spec, pick the next
   unchecked item, finish it completely (code + test + commit), then
   mark it done. Don't try to do multiple items in one pass.
2. **Commit after every item.** Use `git add -A && git commit -m
   "web: <short description>"`. If a commit would be empty, skip it.
3. **Don't break what works.** The PTY/wterm mode is the proven path
   — keep `/ws` and the PTY bridge working. Add new endpoints and
   pages alongside it. Default page should become the new chat UI;
   the old terminal view should be available at `/terminal` for
   power users.
4. **Mobile-first.** Every UI decision: assume a 375px-wide viewport
   and a touch user. Desktop is the easy fallback.
5. **Production-grade means deployable.** Auth, HTTPS guidance, rate
   limits, structured logs, config via env, error pages, health
   checks, CSP headers, tests.
6. **If a step reveals more work than the spec accounts for, update
   the spec** by adding new checkboxes at the end. Don't silently
   expand scope mid-item.

---

## Phase 0 — Research (do this FIRST, before any code)

- [x] **Research current best practices for production AI chat UIs**:
  use `web_fetch` to read at least:
    - https://web.dev/articles/responsive-web-design-basics
    - https://web.dev/articles/pwa-install
    - https://developer.mozilla.org/en-US/docs/Web/Progressive_web_apps
    - https://tailwindcss.com/docs/responsive-design
    - https://vercel.com/docs/ai-sdk
    - https://docs.anthropic.com/en/api/claude-code (any web UI patterns they show)
  Capture findings in a new `WEB_RESEARCH.md` at the repo root with
  short bullet-point notes per source. Don't quote at length — paraphrase.
- [x] **Research mobile chat UI patterns**: look at how ChatGPT, Claude.ai,
  and Perplexity render on mobile. Use `web_fetch` on their marketing
  pages and any public design blogs. Note: safe-area insets, virtual
  keyboard handling, sticky input bar, message virtualization for long
  threads. Add bullets to `WEB_RESEARCH.md`.
- [x] **Research PWA + offline for chat apps**: `web_fetch` the MDN PWA
  guide and any "installable chat app" tutorials. Add bullets to
  `WEB_RESEARCH.md`.
- [x] **Rewrite this ROADMAP.md based on research findings.** Add new
  checkboxes at the end of each phase (or new phases) for anything
  the research surfaced that the original list missed. Re-prioritize
  items if research shows a different order is wiser. Do not remove
  items unless the research explicitly contradicts them.
  **Added items**: See "## Items Added from Research" section at end of this document.

---

## Phase 1 — Foundation: structured event API

The current `/ws` endpoint streams raw PTY bytes. Add a parallel
`/api/chat/ws` endpoint that streams structured `TurnEvent` JSON,
suitable for a custom HTML/JS chat UI. The two endpoints share the
same `ConversationLoop` — only the wire format differs.

- [x] **Define a JSON wire format** for chat events in a new file
  `internal/web/chatproto/protocol.go`. Schema:
    - client → server: `{type:"user_input", text:"..."}`,
      `{type:"permission_reply", tool_use_id:"...", decision:"allow|deny|allow_always"}`,
      `{type:"resize", cols:N, rows:N}` (for the TUI view).
    - server → client: turn-event JSON mirroring `runtime.TurnEvent`
      (text_delta, text_final, tool_start, tool_done, permission_ask,
      ask_user, usage, done, error). Plus a `chat_session_init`
      hello message with the session id.
  Document each message with a doc comment and add a small table of
  examples at the top of the file.
- [ ] **Implement `/api/chat/ws`** in `internal/web/server.go`. It
  should:
    - Build a `ConversationLoop` using the same auth/provider/model
      resolution as the existing PTY path.
    - Run the conversation in a goroutine, translating `TurnEvent`s
      into the JSON protocol above.
    - Accept the client messages above and route them to the loop
      (user_input → SendMessage, permission_reply → PermReply, etc.).
    - Be testable in isolation (split into a `runChatSession` function
      that takes a `*ConversationLoop` and a `json.Encoder`/`Decoder`).
- [ ] **Add Go tests** for the chat WS endpoint in
  `internal/web/chat_test.go`:
    - round-trip a fake client (gorilla `websocket` test client) and
      verify the server emits `chat_session_init` then echoes a
      `text_final` after a `user_input`.
    - verify a `permission_reply` decodes and reaches `PermReply`.
  No live API calls required — use a fake `api.APIClient` that
  emits canned events from a channel.
- [x] **Document the protocol** in `docs/web-protocol.md` (new file)
  with a Mermaid sequence diagram showing one full turn (user input →
  text deltas → tool call → tool result → end_turn).

---

## Phase 2 — Auth & security baseline

- [x] **Basic auth via env var**: add `CLAW_WEB_AUTH=user:pass` (or
  `CLAW_WEB_AUTH_FILE=path`). When set, all `/`, `/api/*`, `/ws`,
  `/static/*` require HTTP Basic Auth. Implement as a middleware in
  `internal/web/auth.go` with constant-time compare. Health
  endpoints (`/healthz`) stay open.
- [x] **Optional bearer token auth**: `CLAW_WEB_TOKEN=<secret>` —
  accept as `Authorization: Bearer ...` OR `?token=...` query
  parameter (so a browser can bookmark a session). Document that
  query-param token leaks via referer; recommend HTTPS.
- [ ] **CSP and security headers** middleware in
  `internal/web/security.go`. Add: `Content-Security-Policy`
  (default-src 'self'; script-src 'self' 'wasm-unsafe-eval'; …),
  `X-Content-Type-Options: nosniff`, `Referrer-Policy: no-referrer`,
  `X-Frame-Options: DENY`, `Permissions-Policy: camera=(),
  microphone=(), geolocation=()`. The wasm-unsafe-eval is required
  for wterm.
- [ ] **Rate limiting** in `internal/web/ratelimit.go`: per-IP token
  bucket for the chat WS endpoint (e.g. 5 new sessions per minute,
  60 messages per minute). Use `golang.org/x/time/rate` (add the
  dep). Configurable via `CLAW_WEB_RATE_RPM`.
- [ ] **Static asset caching**: serve `/static/*` with
  `Cache-Control: public, max-age=3600` for hashed assets, `no-cache`
  for `index.html`. Add content-hashed subdirs for JS/CSS if not
  already present.
- [ ] **TLS termination guide**: new `docs/web-deployment.md` with
  copy-paste configs for Caddy and nginx (TLS, reverse proxy, large
  response buffering for SSE, websocket upgrade headers).

---

## Phase 3 — Mobile-friendly chat UI (vanilla HTML + CSS + JS, no framework)

Build a proper chat UI as a single-page app served from
`internal/web/static/chat.html`. It must work on a 375px-wide
viewport with touch input, and gracefully scale up to desktop.

- [ ] **HTML scaffold** (`chat.html`): semantic structure with
  `<header>`, `<main id="messages">`, `<form id="composer">`,
  `<footer>`. Viewport meta: `width=device-width, initial-scale=1,
  viewport-fit=cover`. Theme color meta for the OS chrome.
- [ ] **CSS reset + design tokens** (`chat.css`): CSS custom
  properties for colors, spacing, type scale. Light + dark mode
  via `prefers-color-scheme` AND a manual toggle persisted in
  `localStorage`. Use `clamp()` for fluid typography.
- [ ] **Responsive layout** with `display: grid`:
  - mobile (< 768px): full-width, single column, sticky composer at
    bottom, hamburger menu for sidebar.
  - tablet (768-1024px): sidebar visible, content centered.
  - desktop (> 1024px): two-column (sidebar + main), max-width
    720px on the message column.
- [ ] **Safe area insets**: `padding:
  max(12px, env(safe-area-inset-top)) … max(12px,
  env(safe-area-inset-bottom))` on the composer. Add a CSS env
  test page (`/safe-area-test`) that draws the safe-area outlines
  for visual verification.
- [ ] **Virtual keyboard handling**: use `visualViewport` API to
  detect the keyboard and resize the composer / scroll-to-bottom on
  `resize` event. Add a test that mounts a mock visualViewport and
  asserts the composer height adjusts.
- [ ] **Message rendering**: render server `text_delta` events into
  message bubbles with `aria-live="polite"` for screen readers.
  Group consecutive `text_delta`s from the same turn into one
  bubble. Use a `<pre><code>` for code blocks; the JS calls
  `Prism.highlight` (or a similar small syntax highlighter — add
  the dep) for the languages the agent is likely to emit (go,
  typescript, python, bash, json, html, css).
- [ ] **Tool-call cards**: render `tool_start`/`tool_done` as
  collapsible cards showing the tool name, a one-line summary of
  the input, and a `<details>` block with the result. Card has
  a small icon, a duration timer, and a copy button.
- [ ] **Permission prompts** as proper modal-style dialogs (not
  text prompts): centered card with the tool name, a formatted view
  of the input, and three buttons: Allow once / Always allow / Deny.
  Send the corresponding `permission_reply` over the WS. On mobile
  the dialog fills the bottom 60% as a sheet.
- [ ] **Composer**: `<textarea>` with `auto-grow` (height adjusts
  to content up to a max of 8 rows). Submit on Enter, newline on
  Shift+Enter. Send button is a 44x44 touch target. Show character
  count when approaching the per-message limit.
- [ ] **Slash command menu**: typing `/` opens a filtered list of
  available commands fetched once at session start (server exposes
  `/api/commands`). Arrow keys / tap to select, Enter to insert.
- [ ] **Reconnection logic**: on WS close, retry with exponential
  backoff (1s, 2s, 4s, max 30s). Show a small "reconnecting…" pill
  in the header. On reconnect, resend the last `user_input` only if
  the server hasn't ack'd it (use a per-message UUID + server
  ack).
- [ ] **Session history sidebar**: list of past sessions from
  `/api/sessions`. Tap to load. Active session is highlighted. On
  mobile, sidebar is a slide-in drawer; on desktop, a fixed column.

---

## Phase 4 — PWA (installable, offline-ready)

- [ ] **Web app manifest** at `/static/manifest.webmanifest`:
  name, short_name, start_url=`/`, display=`standalone`, theme_color,
  background_color, icons in 192px and 512px (and 180px for iOS apple-touch-icon).
- [ ] **Service worker** at `/static/sw.js` (registered from
  `chat.html`):
    - cache-first for `/static/*` assets
    - network-first for `/api/*` and `/ws`
    - offline fallback page that shows "you're offline; reconnect
      to continue" with a retry button
- [ ] **iOS install hints**: `<meta name="apple-mobile-web-app-capable"
  content="yes">`, `<meta name="apple-mobile-web-app-status-bar-style"
  content="black-translucent">`, and a one-time banner that shows
  the "Add to Home Screen" instructions for iOS Safari (which
  doesn't fire `beforeinstallprompt`).
- [ ] **Verify install on desktop Chrome and on Android**: document
  the install flow in `docs/web-deployment.md`.

---

## Phase 5 — Observability

- [ ] **Structured logs** via `log/slog` (stdlib in Go 1.21+).
  Every WS connection gets a `slog.Logger` with `request_id`,
  `remote_addr`, `user_agent`. Log lifecycle events at INFO;
  protocol violations and rate-limit hits at WARN.
- [ ] **Prometheus metrics** at `/metrics` (use the stdlib
  `expvar`-style handler or `github.com/prometheus/client_golang`).
  Expose at minimum:
    - `claw_web_sessions_total{provider,model}` (counter)
    - `claw_web_active_sessions` (gauge)
    - `claw_web_messages_total{role,kind}` (counter)
    - `claw_web_request_duration_seconds` (histogram)
    - `claw_web_errors_total{kind}` (counter)
  Document the metrics in `docs/web-deployment.md`.
- [ ] **Error page** at `/error?code=...&msg=...` rendered when the
  server fails to set up a session (e.g. provider down). Friendly
  UI, copy-paste correlation id.
- [ ] **Health check enrichment**: `/healthz` returns JSON with
  status, version, uptime, active sessions, last provider error
  timestamp. Add a deep check `/readyz` that pings the configured
  provider.

---

## Phase 6 — Polish & docs

- [ ] **Keyboard shortcuts overlay** (`?` to toggle, lists all
  shortcuts). Touch users get a "?" button in the header.
- [ ] **Theme switcher** with 3 options (system, light, dark),
  persisted per-device.
- [ ] **Font size control** (S/M/L) with 12-18px range, persisted.
- [ ] **Code block actions**: copy button, "run in terminal" button
  (if a TUI session is also open), download as file (uses the
  filename from a leading comment if present).
- [ ] **Empty states**: friendly placeholder when there are no
  messages; "thinking…" indicator with elapsed time during long
  model calls; "model is rate-limited" message with a countdown.
- [ ] **Update README.md** with a new "Web UI" section:
    - Quick start (`./claw-code-go web --addr 0.0.0.0:7777`)
    - Auth configuration
    - Production deployment (TLS, reverse proxy)
    - Architecture overview (PTY vs chat mode, when to use each)
    - Mobile install instructions
- [ ] **CHANGELOG.md entry** summarizing the new web UI.

---

## Phase 7 — Tests & CI

- [ ] **Unit tests** for every new package
  (`internal/web/auth_test.go`, `ratelimit_test.go`,
  `security_test.go`, `chatproto_test.go`). Aim for ≥80% coverage
  on new code.
- [ ] **Integration test** in `internal/web/integration_test.go`:
  spins up a `httptest.Server`, opens a `websocket` connection,
  sends a `user_input`, asserts the expected sequence of
  `text_delta` and `text_final` events. Use a fake `api.APIClient`.
- [ ] **HTML/CSS/JS lint**: add a `web/lint` Makefile target that
  runs `html-validate` (or similar) and `eslint` on the chat assets.
  Wire it into `go test ./...` via a `TestWebAssets` Go test that
  shells out, or document it as a separate manual step.
- [ ] **Mobile visual regression**: a Playwright script
  (`web/e2e/mobile.spec.ts`) that opens the chat UI in a 390x844
  viewport (iPhone 14 size), sends a message, screenshots the
  result, and diffs against `web/e2e/baselines/`. Document how to
  regenerate baselines.

---

## Items Added from Research

The following items were added after reviewing the research findings
in `WEB_RESEARCH.md`. They address gaps the original spec didn't
cover but the research showed are important.

### Phase 3 additions — Mobile chat UI

- [ ] **Suggested follow-up chips**: after each AI response, render
  tappable chip buttons with suggested follow-up questions (Perplexity
  pattern). Reduces typing friction on mobile. Implement in
  `chat.html` JS.
- [ ] **Thinking/status indicator with elapsed time**: show a small
  indicator (e.g. "Thinking… 12s") during long model calls, visible
  in the message stream. Distinct from the Phase 6 empty-state
  placeholder — this is an in-stream element. Research shows ChatGPT
  and Perplexity both do this.

### Phase 4 additions — PWA & offline

- [ ] **Network status detection**: add `navigator.onLine` +
  `online`/`offline` event listeners to show a persistent offline
  banner when connectivity is lost. Gating logic: always confirm with
  an actual fetch/WS failure, since `navigator.onLine` has false
  positives.
- [ ] **iOS "Add to Home Screen" instruction banner**: a one-time
  dismissible banner that shows iOS Safari users how to install
  (tap Share → "Add to Home Screen"). Only shown on iOS Safari,
  hidden after dismissal (stored in `localStorage`).

### Phase 7 additions — Tests & CI

- [ ] **PWA installability audit**: add a `web/lighthouse` Makefile
  target that runs Lighthouse (or `@lhci/cli`) against a running
  server and asserts PWA installability score ≥ 90. Document in
  `docs/web-deployment.md`.

---

## Definition of done

- `./claw-code-go web --addr 0.0.0.0:7777` starts the new chat UI
  in a browser.
- The same command on a phone (Android Chrome + iOS Safari) renders
  the chat UI correctly, supports touch input, and is installable
  as a PWA.
- The PTY/wterm mode still works at `/terminal` for users who
  prefer the TUI.
- `CLAW_WEB_AUTH=user:pass` blocks unauthenticated access.
- `curl -H "Authorization: Bearer $TOKEN" …/metrics` returns
  Prometheus metrics.
- `go test ./...` passes.
- The README's Web UI section is up to date.

## Blockers

### ~~Phase 1 item 2 (Implement `/api/chat/ws`) and item 3 (Go tests)~~ — RESOLVED 2026-06-02

The previous iteration incorrectly reported that `ConversationLoop`
does not exist in the codebase. **It does** — see
`internal/runtime/conversation.go:39`:

- `type ConversationLoop struct { ... }` (line 39)
- `func NewConversationLoop(cfg *Config, client api.APIClient) *ConversationLoop` (line 60)
- `func (loop *ConversationLoop) SendMessage(ctx, userText) error` (line 138)
- `func (loop *ConversationLoop) SendMessageStreaming(ctx, userText, events) error` (line 417)
- `type TurnEvent` channel is used in `runOneTurnStreaming` (line 506)

All Phase 1 requirements (auth/provider/model resolution,
`SendMessage`, `TurnEvent` channel, importable from `internal/web`)
are already met by the existing type. The next iteration should
proceed directly to wiring `internal/web/chatproto` to
`*ConversationLoop` — **do not** waste cycles "implementing"
`ConversationLoop` again.

If you can't find it: `grep -rn "type ConversationLoop" internal/`
returns exactly one match at `internal/runtime/conversation.go:39`.

