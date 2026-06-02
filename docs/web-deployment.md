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

## Related documents

- [Web protocol reference](./web-protocol.md) — structured chat API
- [ROADMAP](../ROADMAP.md) — planned features and status
