# Web Research: Production AI Chat UI Best Practices

Research captured during Phase 0 of the claw-code-go web UI roadmap.

---

## 1. Responsive Web Design (web.dev)

Source: https://web.dev/articles/responsive-web-design-basics

- **Viewport meta tag is mandatory**: `<meta name="viewport" content="width=device-width, initial-scale=1">` — without it, mobile browsers render at desktop width and scale down.
- **Fluid layouts over fixed**: Use relative units (`%`, `vw`, `rem`) instead of fixed `px`. CSS Grid and Flexbox are the modern approach; avoid float-based layouts.
- **Breakpoints based on content, not devices**: Choose breakpoints where the content looks awkward, not at specific device widths. Common practice: 480px (small mobile), 768px (tablet), 1024px (desktop).
- **Responsive images**: Use `srcset` and `<picture>` for art direction. Serve appropriately sized images per viewport.
- **Media queries for breakpoints**: `@media (min-width: ...)` is the standard pattern. Mobile-first means base styles are for the smallest screen, breakpoints add complexity.
- **Touch targets**: Minimum 48x48dp tap target (Google's guideline). 44x44pt for iOS.

---

## 2. Progressive Web Apps (MDN)

Source: https://developer.mozilla.org/en-US/docs/Web/Progressive_web_apps

- PWAs are web apps that use modern APIs to deliver app-like experiences: installable, offline-capable, with push notifications.
- **Required for installability**: HTTPS, a valid Web App Manifest, and a registered service worker with a fetch handler.
- **App manifest** (`manifest.webmanifest`): defines name, icons, start URL, display mode (`standalone`, `fullscreen`, `minimal-ui`), theme color, background color.
- **Service worker lifecycle**: Install → Activate → Fetch. The `install` event is where you pre-cache critical assets. The `activate` event is for cache cleanup.
- **Service worker scope**: Limited to the directory the SW file is served from (and below). Serve from root for full app coverage.
- **`display: standalone`**: The app opens in its own window without browser chrome — critical for the app-like feel.

---

## 3. PWA Installation Criteria (web.dev)

Source: https://web.dev/learn/pwa/installation, https://web.dev/articles/install-criteria

- **Chrome/Edge install criteria**: HTTPS, valid manifest with `name`/`short_name`/`start_url`/`display`/icons (192px + 512px), registered service worker with fetch handler, user engagement heuristic (30+ seconds on site, at least one click/tap).
- **Desktop install**: Chrome 73+ shows an install button in the address bar. The `beforeinstallprompt` event can be intercepted to show a custom install prompt.
- **iOS/iPadOS**: Safari does NOT fire `beforeinstallprompt`. Users install via "Add to Home Screen" in the Share menu. Needs `<meta name="apple-mobile-web-app-capable" content="yes">` and `<meta name="apple-mobile-web-app-status-bar-style">` plus an apple-touch-icon (180x180).
- **Avoid aggressive prompting**: Don't show install prompts on page load. Wait for user engagement signals.
- **WebAPK on Android**: Chrome can generate an APK with proper manifest + icons, making the PWA appear in the app drawer alongside native apps.

---

## 4. Tailwind CSS Responsive Design

Source: https://tailwindcss.com/docs/responsive-design

- **Mobile-first by default**: Unprefixed utilities apply to all screen sizes. Prefixed utilities (e.g., `sm:`, `md:`, `lg:`) apply at the specified breakpoint and above.
- **Default breakpoints**: `sm: 640px`, `md: 768px`, `lg: 1024px`, `xl: 1280px`, `2xl: 1536px`.
- **Responsive variants for every utility**: Grid columns, flex direction, spacing, typography — all can be prefixed. E.g., `grid-cols-1 md:grid-cols-2`.
- **Container queries now supported** in Tailwind v4: `@container` and responsive variants for component-level responsiveness.
- **Dark mode**: `dark:` prefix variant, can be toggled via `prefers-color-scheme` or a manual class on `<html>`.
- Even though we're using vanilla CSS, the Tailwind philosophy informs our approach: mobile-first base styles, breakpoint-prefixed overrides.

---

## 5. Vercel AI SDK Patterns

Source: https://vercel.com/docs/ai-sdk

- The Vercel AI SDK is a TypeScript toolkit for building AI applications. While our backend is Go, the SDK's **protocol patterns** are instructive:
- **Provider abstraction**: A unified interface (`generateText`, `streamText`) works across OpenAI, Anthropic, Google, etc. Our Go code does the same with the `api.APIClient` interface.
- **Streaming is the default UX**: Text responses stream token-by-token via `text-delta` events in the `useChat` hook. Our `TurnEvent` → JSON protocol mirrors this.
- **Tool calls are first-class**: The SDK handles tool call request/response cycles. The UI renders tool invocations as expandable cards with status indicators.
- **`useChat` hook**: React hook that manages WebSocket/SSE connection, message state, loading/error states. Our vanilla JS chat UI needs equivalent state machine logic.
- **Structured output**: `generateObject` and `streamObject` for JSON schema-constrained responses. Not immediately relevant but a pattern to note for future.

---

## 6. Anthropic Claude Code API

Source: https://docs.anthropic.com/en/api/claude-code

- Claude Code is triggered via API with a structured routine approach.
- The API uses **JSON-formatted tool calls** in a streaming response pattern similar to our `TurnEvent` design.
- Key pattern: the API returns `text_delta`, `tool_use`, and `tool_result` events as separate stream frames — exactly the structure we're adopting for our WebSocket protocol.

---

## 7. Mobile Chat UI Patterns (ChatGPT, Claude.ai, Perplexity)

Sources: OpenAI ChatGPT web updates (Nov 2024), Claude.ai design docs, Perplexity UX analysis, community implementations, platform-specific docs

### ChatGPT Web (Nov 2024 redesign)
- **Floating sidebar with auto-dismiss**: On mobile, the sidebar is always in floating mode and auto-closes when switching threads. On desktop, it can be pinned open alongside the canvas if space permits.
- **Composer bar with fade effect**: The composer sits at the bottom with content fading underneath as you scroll. "New chat" button moved left of the model picker for easier reach.
- **Improved on-screen keyboard handling**: Major fixes for iOS & Android keyboard behavior. Composer is not auto-focused when switching conversations, preventing jarring keyboard pop-ups.
- **Scroll behavior fixes**: New conversations scroll into view at top of screen; no auto-scroll to bottom during message generation (user can read previous messages while waiting).
- **System icon removed**: Removed to save horizontal space on narrow screens.

### Claude.ai Mobile Patterns
- **Responsive layout generation**: Claude.ai encourages building with responsive web techniques — designs that adapt across desktop, tablet, and mobile viewports without separate codebases.
- **Safe area compliance**: Official design guidelines mandate respecting safe areas on notched devices and handling constrained viewports for touch interaction.
- **Native-feature integration**: Mobile apps surface native OS capabilities (calendar invites, email drafts, messages) via tool calls, but the web UI must handle these through structured fallback dialogs.

### Perplexity Mobile UX Principles (based on Nielsen heuristics)
- **System status visibility**: "Considering 8 sources…" status indicator builds trust during AI processing. Shows source count, progress feedback.
- **Recognition over recall**: Suggested follow-up questions appear after each answer — user taps rather than types. Reduces cognitive load on mobile.
- **Thumb-friendly design (Fitts's Law)**: All primary tap targets placed in the thumb zone (bottom 60% of screen). Minimum 44x44dp touch targets.
- **Text-first philosophy**: Focus on content readability over UI chrome. Widgets (weather, calculator) added only where AI is unnecessary, keeping the interface clean.
- **Haptic/tactile feedback**: Use for confirmations of important actions (permission grants, message sends) to reinforce touch interactions.
- **Error prevention**: Explicit confirmation before destructive actions; clear escape hatches from any flow.

### Universal Mobile Chat Patterns (synthesized)
- **Sticky composer bar at the bottom**: The text input is fixed at the viewport bottom, with messages scrolling above it. Uses `position: sticky` or `position: fixed`.
- **Safe area insets on notched phones**: All major chat apps use `env(safe-area-inset-bottom)` to pad the composer above the home indicator. This is mandatory for iPhone X+ and modern Android phones with gesture nav.
- **Virtual keyboard handling**: The `visualViewport` API is the modern way to detect keyboard open/close. On `resize`, adjust the composer position and scroll the message list to the bottom. Fallback: `window.innerHeight` change detection. ChatGPT's fix: don't auto-focus composer on thread switch to prevent keyboard jump.
- **Auto-growing textarea**: `<textarea>` that grows from 1 row to ~8 rows max, then scrolls internally. Accomplished via `scrollHeight` measurement on `input` event.
- **Enter to send, Shift+Enter for newline**: Universal chat convention. Mobile keyboards show a "send" action key.
- **Message bubbles**: User messages right-aligned, AI messages left-aligned. Code blocks in `<pre><code>` with syntax highlighting. Typing indicators (animated dots) while waiting for AI response.
- **Message virtualization for long threads**: For performance with 1000+ messages, only render visible messages. Libraries: `react-virtuoso`, `react-window`. For vanilla JS: use an `IntersectionObserver` approach or infinite scroll with a cap.
- **Dark/light mode**: System preference via `prefers-color-scheme` with a manual override toggle persisted in `localStorage`.
- **Sidebar/drawer pattern**: Session history in a sidebar. On mobile (<768px), it's hidden behind a hamburger menu and slides in as an overlay drawer with auto-dismiss on thread switch. On desktop, it's a persistent column.
- **`aria-live="polite"`**: For screen reader announcements of new messages. Critical for accessibility.
- **Touch-friendly**: 44px minimum tap targets, adequate spacing between interactive elements, no hover-dependent UI.
- **Suggested follow-ups**: Tap-to-continue chips after AI responses reduce typing friction on mobile (Perplexity pattern).
- **Status/thinking indicators**: Show source count, elapsed time, or progress dots during long operations (ChatGPT/Perplexity pattern).

---

## 8. PWA + Offline for Chat Apps

Sources: MDN PWA guides (Making PWAs installable), Google PWA codelabs, community chat PWA implementations, caching strategy articles

### MDN: Making PWAs Installable
- **Installability requires**: HTTPS, a web app manifest with `name`/`icons`/`start_url`/`display`, and a registered service worker with a fetch handler. All three must be present for the browser to offer install.
- **`display: standalone`**: Opens without browser chrome, critical for app-like feel. Alternatives: `fullscreen`, `minimal-ui`, `browser`.
- **Manifest `start_url`**: Should be `"/"` (root-relative). Must be same-origin. This is the URL launched when the user opens the installed PWA.
- **Icon requirements**: Chrome requires at least 192x192 and 512x512 PNG icons. iOS requires a 180x180 apple-touch-icon for the home screen.
- **iOS Safari special handling**: Safari does NOT use the Web App Manifest for icons. Must include `<link rel="apple-touch-icon" href="...">` and `<meta name="apple-mobile-web-app-capable" content="yes">` separately.

### Service Worker Lifecycle & Caching (from Google codelabs)
- **Registration**: `navigator.serviceWorker.register('/sw.js')` in the main JS. The SW file must be served from the root (or a parent path) to control the full scope.
- **Install event**: Pre-cache critical assets (HTML, CSS, JS, icons). Use `cache.addAll([...])` inside `event.waitUntil()`.
- **Activate event**: Clean up old cache versions. `caches.keys().then(keys => keys.filter(k => k !== CURRENT_CACHE).forEach(k => caches.delete(k)))`.
- **Fetch event**: Intercept all network requests. Apply strategy based on URL pattern.
- **SW update flow**: When SW file changes, the new SW installs in background but waits for all tabs to close before activating (unless `skipWaiting()` is called).

### Caching Strategies for Chat Apps
- **Cache-first for static assets**: App shell (HTML, CSS, JS, icons, fonts) should be pre-cached during SW `install`. Served from cache instantly; updated in background via SW version bumps. Best for hashed assets that don't change.
- **Network-first for API endpoints**: Chat messages and session data should attempt network first, fall back to cache or show "offline" state. Caching dynamic chat content is complex and not recommended for real-time apps.
- **Stale-while-revalidate for semi-static content**: Serve from cache, fetch update in background. Good for command lists, configuration, user preferences.
- **No caching for WebSocket connections**: WS is inherently network-dependent. The SW cannot intercept or cache WebSocket traffic.

### Offline Chat App Patterns (from community implementations)
- **IndexedDB for offline message queue**: Chat apps (e.g., ChattApp-PWA) store outgoing messages in IndexedDB when offline, sync when reconnected. Incoming messages are ephemeral — not cached, just displayed live.
- **CRDT-based sync**: Advanced pattern (HowProgrammingWorks/PWA) uses Conflict-free Replicated Data Types over a shared WebSocket for automatic reconnection and lost message retrieval. Overkill for single-session chat but relevant for multi-device sync.
- **WebSocket reconnection**: Exponential backoff with jitter (1s → 2s → 4s → max 30s) to prevent thundering herd on reconnect. Re-send last `user_input` only if server hasn't ack'd (per-message UUID tracking). Show "Reconnecting..." pill in header.
- **Single WebSocket shared across tabs**: Use `BroadcastChannel` API or `localStorage` events to coordinate a single WS connection across multiple tabs of the same PWA.

### Offline Fallback & Resilience
- **Offline fallback page**: A static HTML page cached during SW `install` that says "You're offline — reconnect to continue" with a retry button that reloads the app.
- **Cache versioning**: Use a cache name with version string (e.g., `claw-chat-v1`). On new SW `activate`, delete old caches to prevent stale assets.
- **Network status detection**: Use `navigator.onLine` + `online`/`offline` events to update UI state. Not reliable alone (false positives) — always gate on actual fetch/WS failure.

### Platform Limitations
- **iOS PWA limitations**: No push notifications in iOS Safari PWAs (as of 2024). No background sync. Service worker cache is evicted more aggressively on iOS (capped at ~50MB). No `beforeinstallprompt` event — must use custom "Add to Home Screen" instruction banner.
- **Android/Chrome**: Full PWA support with WebAPK generation (appears in app drawer). `beforeinstallprompt` available for custom install UI. Background sync supported but limited to periodic sync (not real-time).
- **Desktop PWAs**: Chrome 73+ shows install button in address bar. `beforeinstallprompt` works. Edge supports the same criteria as Chrome.

---

## Summary of Key Decisions for Implementation

| Area | Decision |
|------|----------|
| Layout | CSS Grid with mobile-first breakpoints (480, 768, 1024) |
| Composer | Fixed bottom bar with safe-area padding, auto-growing textarea |
| Keyboard | `visualViewport` API for keyboard-aware layout |
| Auth | HTTP Basic Auth + Bearer token (env var config) |
| Offline | Cache-first for `/static/*`, network-first for `/api/*` and `/ws` |
| WS protocol | JSON text frames, one TurnEvent per frame |
| PWA | Web App Manifest + SW + iOS meta tags |
| Theme | System preference + manual toggle, persisted in localStorage |
| Touch | 44px min tap targets, no hover-dependent interactions |