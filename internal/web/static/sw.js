/**
 * claw-code-go Service Worker
 *
 * Cache strategy:
 *   - /static/*  → cache-first (immutable, content-hashed assets)
 *   - /api/*, /ws → network-first (live data)
 *   - /           → network-first (HTML shell)
 *
 * Offline: shows a fallback page with a retry button.
 */

const CACHE_NAME = 'claw-code-v1';

// Assets to pre-cache on install.
const PRECACHE_URLS = [
  '/',
  '/static/css/chat.css',
  '/static/js/chat.js',
  '/static/manifest.webmanifest'
];

// ── Install: pre-cache critical assets ──────────────────────────
self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => {
      return Promise.allSettled(
        PRECACHE_URLS.map((url) =>
          cache.add(url).catch(() => {
            // Ignore individual failures; not all assets may exist yet.
          })
        )
      );
    }).then(() => self.skipWaiting())
  );
});

// ── Activate: clean old caches ───────────────────────────────────
self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((keys) =>
      Promise.all(
        keys
          .filter((key) => key !== CACHE_NAME)
          .map((key) => caches.delete(key))
      )
    ).then(() => self.clients.claim())
  );
});

// ── Fetch: routing by URL pattern ────────────────────────────────
self.addEventListener('fetch', (event) => {
  const url = new URL(event.request.url);

  // Only handle same-origin GET requests.
  if (url.origin !== self.location.origin) return;
  if (event.request.method !== 'GET') return;

  // ── /static/* : cache-first ─────────────────────────────────
  if (url.pathname.startsWith('/static/')) {
    event.respondWith(cacheFirst(event.request));
    return;
  }

  // ── /api/* and /ws : network-first ──────────────────────────
  if (url.pathname.startsWith('/api/') || url.pathname === '/ws') {
    event.respondWith(networkFirst(event.request));
    return;
  }

  // ── / and other HTML pages : network-first with offline fallback
  event.respondWith(
    networkFirst(event.request).catch(() => {
      return caches.match('/') || offlineFallback();
    })
  );
});

// ── Cache-first strategy ─────────────────────────────────────────
async function cacheFirst(request) {
  const cached = await caches.match(request);
  if (cached) return cached;

  try {
    const response = await fetch(request);
    if (response && response.status === 200) {
      const cache = await caches.open(CACHE_NAME);
      cache.put(request, response.clone());
    }
    return response;
  } catch (err) {
    // If not cached and network fails, return the offline fallback
    // for navigation requests; for assets just let the error through.
    return new Response('', { status: 504, statusText: 'Gateway Timeout' });
  }
}

// ── Network-first strategy ───────────────────────────────────────
async function networkFirst(request) {
  try {
    const response = await fetch(request);
    // Cache successful GET responses.
    if (response && response.status === 200) {
      const cache = await caches.open(CACHE_NAME);
      cache.put(request, response.clone());
    }
    return response;
  } catch (err) {
    const cached = await caches.match(request);
    if (cached) return cached;
    throw err;
  }
}

// ── Offline fallback page ────────────────────────────────────────
function offlineFallback() {
  const body = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="theme-color" content="#1a1a2e">
<title>Offline — claw-code-go</title>
<style>
  * { margin:0; padding:0; box-sizing:border-box; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    background: #1a1a2e; color: #e0e0e0;
    display: flex; align-items: center; justify-content: center;
    min-height: 100vh; min-height: 100dvh;
    padding: 24px;
  }
  .card {
    text-align: center; max-width: 360px;
    background: rgba(255,255,255,0.05); border-radius: 16px; padding: 40px 32px;
  }
  .icon { font-size: 64px; margin-bottom: 16px; }
  h1 { font-size: 1.5rem; margin-bottom: 8px; color: #fff; }
  p { font-size: 0.95rem; color: #a0a0b0; margin-bottom: 24px; line-height: 1.5; }
  button {
    background: #4fc3f7; color: #1a1a2e; border: none;
    padding: 12px 32px; border-radius: 8px; font-size: 1rem;
    font-weight: 600; cursor: pointer; transition: background 0.2s;
    touch-action: manipulation; -webkit-tap-highlight-color: transparent;
  }
  button:hover, button:active { background: #29b6f6; }
</style>
</head>
<body>
  <div class="card">
    <div class="icon">📡</div>
    <h1>You're offline</h1>
    <p>claw-code-go requires a connection to the server. Check your network and try again.</p>
    <button onclick="location.reload()">Retry</button>
  </div>
</body>
</html>`;

  return new Response(body, {
    status: 200,
    statusText: 'OK',
    headers: { 'Content-Type': 'text/html; charset=utf-8' }
  });
}