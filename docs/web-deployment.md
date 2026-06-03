# Web UI Deployment Guide

This document covers production deployment of the claw-code-go web UI
behind a TLS-terminating reverse proxy. The web server listens on a plain
HTTP port (`--addr`) and should always be placed behind a reverse proxy
(Caddy or nginx) that handles TLS, request buffering, and WebSocket upgrade.

---

## Quick start (local testing)

```bash
# Run the web server on localhost
./claw-code-go web --addr 127.0.0.1:7777

# With basic auth
CLAW_WEB_AUTH=user:pass ./claw-code-go web --addr 0.0.0.0:7777

# With bearer token (accepts ?token=... query param for browser bookmarking)
CLAW_WEB_TOKEN=my-secret-token ./claw-code-go web --addr 0.0.0.0:7777
```

---

## Caddy (recommended)

Caddy auto-provisions Let's Encrypt certificates and handles WebSocket
upgrade transparently. No special configuration is needed for SSE or
large response buffering — Caddy's defaults work well.

### Caddyfile

```caddyfile
chat.example.com {
    reverse_proxy localhost:7777
}
```

### With basic auth at the proxy level (optional)

If you prefer to handle auth at the reverse proxy instead of via
`CLAW_WEB_AUTH`, you can use Caddy's built-in `basicauth`:

```caddyfile
chat.example.com {
    basicauth {
        alice $2a$14$...
    }
    reverse_proxy localhost:7777
}
```

---

## nginx

nginx requires explicit configuration for WebSocket upgrade and
buffering. The configuration below is production-tested for ttyd-style
(PTY bridge) and chat WebSocket endpoints.

### nginx site config

```nginx
server {
    listen 443 ssl;
    server_name chat.example.com;

    ssl_certificate     /etc/letsencrypt/live/chat.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/chat.example.com/privkey.pem;

    # --- WebSocket endpoints ---
    location /ws {
        proxy_pass http://localhost:7777;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Disable buffering for PTY/SSE streams
        proxy_buffering off;
        proxy_read_timeout 86400s;
    }

    location /api/chat/ws {
        proxy_pass http://localhost:7777;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Disable buffering for SSE streams sent over WS
        proxy_buffering off;
        proxy_read_timeout 86400s;
    }

    # --- Everything else (static assets, health checks, API) ---
    location / {
        proxy_pass http://localhost:7777;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Large response buffering for SSE (future /api/chat/sse endpoint)
        proxy_buffering off;
    }
}
```

---

## TLS certificate provisioning

### Let's Encrypt with Caddy (automatic)

Caddy handles certificate provisioning and renewal automatically.
No additional steps required beyond the Caddyfile above.

### Let's Encrypt with nginx + certbot

```bash
# Install certbot
apt install certbot python3-certbot-nginx

# Obtain and install certificate
certbot --nginx -d chat.example.com

# Auto-renewal (certbot adds a systemd timer automatically)
certbot renew --dry-run
```

---

## Operating system service

### systemd unit file (`/etc/systemd/system/claw-code-web.service`)

```ini
[Unit]
Description=claw-code-go web UI
After=network.target

[Service]
Type=simple
User=clawcode
Group=clawcode
WorkingDirectory=/home/clawcode
ExecStart=/usr/local/bin/claw-code-go web --addr 127.0.0.1:7777
Restart=always
RestartSec=5

# Environment
Environment=CLAW_WEB_AUTH_FILE=/etc/clawcode/web-auth
Environment=CLAW_WEB_RATE_RPM=30

# Security hardening
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ProtectHome=read-only
ReadWritePaths=/home/clawcode/.claw-code-go

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
systemctl daemon-reload
systemctl enable --now claw-code-web
```

---

## Firewall considerations

The web server should bind to `127.0.0.1` (localhost) and not be
directly exposed to the internet. Only the reverse proxy (Caddy/nginx)
needs port 443 (and optionally 80 for Let's Encrypt HTTP challenges).

```bash
# UFW example
ufw allow 443/tcp
ufw allow 80/tcp   # for Let's Encrypt HTTP-01 challenge
ufw deny 7777      # ensure the app port is not externally reachable
```

---

## Health check monitoring

The `/healthz` endpoint returns `200 OK` and the string `ok`.
Configure your load balancer or monitoring system to poll it:

```bash
curl -f http://localhost:7777/healthz
```

---

## PWA install verification

The chat UI is a Progressive Web App (PWA) that can be installed to
the home screen on desktop Chrome, Android Chrome, and iOS Safari
(using the "Add to Home Screen" flow). This section documents how to
verify that the install flow works on each platform.

### Prerequisites for installability

For the browser to offer the install prompt, all of these must be
true:

1. **HTTPS** — the site must be served over TLS (or `localhost` for
   development). The easiest path is running Caddy in front of the
   app (see the Caddy section above). Let's Encrypt auto-provisions
   a certificate.
2. **Valid manifest** — the server must serve
   `/static/manifest.webmanifest` with `Content-Type:
   application/manifest+json`. The manifest must include `name` (or
   `short_name`), `start_url`, `icons` with at least a 192×192 and a
   512×512 PNG, and `display: standalone`.
3. **Registered service worker** — `/static/sw.js` must register
   successfully with a `fetch` handler. Open DevTools → Application →
   Service Workers to confirm it shows "activated and is running".
4. **User engagement** — on desktop Chrome, the browser also requires
   some user interaction (click, scroll, keystroke) before showing
   the install prompt. On Android Chrome this heuristic is relaxed.

### Verifying the checklist programmatically

Open Chrome DevTools → Application → Manifest. The "Installability"
section shows a green checkmark for each criterion. If any criterion
fails, the section explains why (e.g. "No matching service worker
detected").

Alternatively, run a Lighthouse audit (DevTools → Lighthouse →
Progressive Web App) and check the "Installable" group. All items in
that group must pass.

### Desktop Chrome install flow

1. Open the chat UI at `https://chat.example.com` in Chrome.
2. Open DevTools → Application → Manifest. Verify all installability
   checks pass (green checkmarks).
3. Look for the install icon (a monitor with a down-arrow) in the
   address bar, on the right side. If it doesn't appear immediately,
   interact with the page (click a message, type in the composer).
4. Click the install icon. Chrome shows a dialog: "Install app?"
   with the app name and icon. Click **Install**.
5. The app opens in a standalone window (no browser chrome, no
   address bar). Verify:
   - The window title bar shows "claw-code-go".
   - Resizing the window reflows the chat layout correctly.
   - The `/terminal` page works inside the PWA window (navigate by
     typing the URL, or add a link in the sidebar).
   - The service worker is active (DevTools → Application → Service
     Workers).
6. **Uninstall**: click the three-dot menu in the PWA title bar →
   "Uninstall claw-code-go…". This removes the app from
   `chrome://apps`.

### Android Chrome install flow

1. Open the chat UI at `https://chat.example.com` in Chrome on an
   Android device (tested on Android 12+ with Chrome 110+).
2. Chrome automatically shows a bottom-sheet install prompt after
   the first page load if all PWA criteria are met. The sheet says
   "Add claw-code-go to Home screen".
3. If the automatic prompt doesn't appear (or was dismissed), tap
   the three-dot menu → "Add to Home screen".
4. In the dialog that appears:
   - The app name is shown as "claw-code-go" (from `short_name` in
     the manifest).
   - The icon is the 192×192 maskable PNG from the manifest.
   - Tap **Add**. On some Android launchers you can also drag to
     place the icon manually.
5. After adding, tap the claw-code-go icon on the home screen. The
   app opens in **standalone** mode (full screen, no browser
   chrome). Verify:
   - The status bar color matches the manifest `theme_color`
     (`#1a1a2e`).
   - The splash screen background matches `background_color`.
   - The chat UI fills the screen edge-to-edge.
   - The soft keyboard works correctly with the composer
     (visualViewport handling).
   - Switch to airplane mode → the offline fallback page appears
     with a "Retry" button. Switch back online → tap Retry → chat
     reloads.
   - The Android back button (gesture or hardware) closes the PWA
     and returns to the home screen (does not navigate back in
     browser history).
6. **Uninstall**: long-press the home screen icon → "Uninstall", or
   go to Settings → Apps → claw-code-go → Uninstall.

### iOS Safari "Add to Home Screen" flow

iOS Safari does not support the `beforeinstallprompt` event or
automatic PWA install. Instead, users must manually add the app via
the Share menu. The chat UI includes an iOS-specific instruction
banner that guides users through this flow.

1. Open the chat UI at `https://chat.example.com` in Safari on an
   iPhone or iPad (iOS 15+).
2. The iOS install banner appears at the top of the page (one-time;
   dismissed banner is stored in `localStorage` and won't reappear).
3. Tap the **Share** icon (square with an up-arrow) in the Safari
   toolbar.
4. Scroll down and tap **"Add to Home Screen"**.
5. The dialog shows the app name ("claw-code-go") and the
   apple-touch-icon (180×180). Tap **Add**.
6. The icon appears on the iOS home screen. Tap it to open. Verify:
   - The app opens in standalone mode (no Safari toolbar or tab bar).
   - The status bar is black-translucent (matches the
     `apple-mobile-web-app-status-bar-style` meta tag).
   - Safe area insets work (content doesn't overlap the notch or
     home indicator).
   - The soft keyboard pushes the composer up correctly.
   - Multitasking (swipe up from the bottom or double-click home)
     shows the app as a separate card from Safari.
7. **Uninstall**: long-press the home screen icon → "Remove App" →
   "Delete App".

### Common PWA install issues and fixes

| Symptom | Likely cause | Fix |
|---|---|---|
| No install icon in Chrome address bar | Manifest not served with correct `Content-Type` | Ensure the Go server sets `Content-Type: application/manifest+json` for `.webmanifest` files |
| Manifest "Installability" shows a red X for icons | Icon files missing or wrong size | Verify `/static/icons/icon-192.png` and `icon-512.png` exist and are valid PNGs of the declared sizes |
| Service worker doesn't activate | SW script has a syntax error | Check DevTools → Application → Service Workers for error messages; test `sw.js` in isolation |
| Android install prompt doesn't appear | Site not served over HTTPS | Deploy behind Caddy/nginx with TLS; `localhost` works for dev but not on a real device |
| iOS standalone mode shows Safari chrome | Missing `apple-mobile-web-app-capable` meta | Verify `<meta name="apple-mobile-web-app-capable" content="yes">` is in `<head>` |
| PWA opens as browser tab instead of standalone | `display` in manifest is not `standalone` | Check manifest: must have `"display": "standalone"` |

---

## Mobile visual regression tests (Playwright)

The project includes Playwright-based visual regression tests that
capture screenshots of the chat UI at iPhone 14 viewport size (390×844)
and compare them against stored baselines. This catches unintended CSS
layout regressions.

### Prerequisites

```bash
npm ci                          # install deps including @playwright/test
npx playwright install chromium # install Chromium browser
./claw-code-go web --addr 127.0.0.1:7777  # start the server
```

### Running the tests

```bash
# Run against a running server
make web/e2e

# Or override the server URL
E2E_BASE_URL=http://192.168.1.50:7777 make web/e2e
```

### Regenerating baselines

When the UI changes intentionally (e.g. new features, CSS redesign),
baselines must be regenerated:

```bash
# Automated: remove old baselines, run tests (first run saves new baselines)
make web/e2e-update

# Manual steps:
# 1. Start the server: ./claw-code-go web --addr 127.0.0.1:7777
# 2. Remove old baselines: rm -f web/e2e/baselines/*.png
# 3. Run: npx playwright test --config web/e2e/playwright.config.ts web/e2e/mobile.spec.ts
#    (First run will fail, saving new baselines to web/e2e/baselines/)
# 4. Visually review each .png in web/e2e/baselines/
# 5. Run again: npx playwright test --config web/e2e/playwright.config.ts web/e2e/mobile.spec.ts
#    (Second run should pass against the new baselines)
# 6. Commit the new baselines: git add web/e2e/baselines/ && git commit -m "web: update visual baselines"
```

### Test coverage

The visual regression suite covers:
- Empty state (no messages, placeholder visible)
- Composer with text input
- Sidebar toggle (mobile hamburger menu)
- Dark mode rendering
- Keyboard shortcuts overlay
- Offline banner on network disconnect
- Sticky composer with scrolled messages

---

---

## PWA installability audit (Lighthouse CI)

The project includes an automated Lighthouse CI target that audits the
chat UI for PWA installability and asserts a score ≥ 90. This catches
regressions in the manifest, service worker, or security headers that
would block PWA installation.

### Prerequisites

```bash
npm ci                          # install deps including @lhci/cli
npx playwright install chromium # LHCI reuses the Playwright Chromium
./claw-code-go web --addr 127.0.0.1:7777  # start the server
```

### Running the audit

```bash
# Run Lighthouse against a running server and assert PWA score ≥ 90
make web/lighthouse

# Or override the server URL
LIGHTHOUSE_BASE_URL=http://192.168.1.50:7777 make web/lighthouse

# CI-friendly variant (server mode with explicit collect + assert steps)
make web/lighthouse-server
```

### What it checks

The `.lighthouserc.json` config file asserts:

- **PWA category score ≥ 0.9** (90%). This covers:
  - Valid web app manifest with icons, name, start_url, display
  - Registered service worker with fetch handler
  - HTTPS (or localhost) — LHCI won't flag localhost
  - Proper viewport meta tag
  - Redirects HTTP to HTTPS (if applicable)
  - `apple-mobile-web-app-capable` and iOS meta tags
- All **lighthouse:recommended** assertions for performance,
  accessibility, best-practices, and SEO are also evaluated
  (warnings only; PWA is the gating failure).

### Interpreting failures

If `make web/lighthouse` exits non-zero, the LHCI output will show
which assertion failed. Common causes:

| Failure | Fix |
|---|---|
| `categories:pwa < 0.9` | Check manifest validity, service worker registration, and HTTPS in DevTools → Application → Manifest |
| `installable-manifest` | Verify `/static/manifest.webmanifest` Content-Type and icon sizes |
| `service-worker` | Check `/static/sw.js` has a `fetch` handler and registers without errors |
| `is-on-https` | Run behind Caddy/nginx with TLS; localhost is exempt |

### Configuration

The audit is configured in `.lighthouserc.json` at the repo root.
The `LIGHTHOUSE_BASE_URL` env var (default `http://127.0.0.1:7777`)
is expanded at runtime. To change assertion thresholds, edit the
`assertions` block in that file.

---

## Related documents

- [Web protocol reference](./web-protocol.md) — structured chat API
- [ROADMAP](../ROADMAP.md) — planned features and status

---

## Prometheus metrics

The server exposes a `/metrics` endpoint in Prometheus text format
for scraping by Prometheus, Grafana, or any OpenMetrics-compatible
collector.

### Available metrics

| Metric | Type | Labels | Description |
|---|---|---|---|
| `claw_web_sessions_total` | counter | `provider`, `model` | Total chat sessions created |
| `claw_web_active_sessions` | gauge | — | Currently active chat sessions |
| `claw_web_messages_total` | counter | `role` (user\|assistant), `kind` (text\|tool_call\|tool_result) | Messages processed |
| `claw_web_request_duration_seconds` | histogram | `method`, `path` | HTTP request duration (excludes WS upgrades) |
| `claw_web_errors_total` | counter | `kind` (auth_failure\|rate_limit\|server_error\|client_error) | Errors by category |

### Scraping configuration

**Prometheus** (`prometheus.yml`):

```yaml
scrape_configs:
  - job_name: claw-code-go
    static_configs:
      - targets: ['localhost:7777']
    metrics_path: /metrics
    # If CLAW_WEB_TOKEN is set, include it as a query param:
    # params:
    #   token: ['<your-token>']
```

**Grafana**: import a Prometheus dashboard and create panels from the
metrics above. The `claw_web_request_duration_seconds` histogram is
particularly useful for setting up RED (Rate, Errors, Duration)
dashboards.

### Example curl

```bash
# No auth
curl -s http://localhost:7777/metrics | head -20

# With bearer token
curl -s -H "Authorization: Bearer $CLAW_WEB_TOKEN" http://localhost:7777/metrics | grep claw_web
```

Note: the `/metrics` endpoint is subject to the same auth middleware
as the rest of the app. If `CLAW_WEB_AUTH` or `CLAW_WEB_TOKEN` is set,
the scraper must authenticate (use `Authorization: Bearer <token>` or
`?token=<token>` query parameter).

---

## Related documents

- [Web protocol reference](./web-protocol.md) — structured chat API
- [ROADMAP](../ROADMAP.md) — planned features and status
