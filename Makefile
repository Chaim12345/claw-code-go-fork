.PHONY: web/lint web/lint-html web/lint-js web/lint-fix web/e2e web/lighthouse

# ── Web asset linting ─────────────────────────────────────────────
# Requires: Node.js and npm dependencies (html-validate, eslint).
# Run `npm ci` to install deps before using these targets.
# Uses direct node_modules/.bin paths to avoid npx network overhead.

ESLINT   := node_modules/.bin/eslint
HTMLV    := node_modules/.bin/html-validate
PW       := node_modules/.bin/playwright

web/lint: web/lint-html web/lint-js
	@echo "✅ All web lints passed."

web/lint-html:
	@echo "🔍 Linting HTML files…"
	$(HTMLV) --config .htmlvalidate.json \
		internal/web/static/chat.html \
		internal/web/static/index.html \
		internal/web/static/error.html \
		internal/web/static/safe-area-test.html

web/lint-js:
	@echo "🔍 Linting JavaScript files…"
	$(ESLINT) --config eslint.config.js \
		internal/web/static/js/chat.js \
		internal/web/static/sw.js

web/lint-fix:
	$(ESLINT) --config eslint.config.js --fix \
		internal/web/static/js/chat.js \
		internal/web/static/sw.js

# ── Mobile visual regression (Playwright) ───────────────────────
# Requires: Node.js, @playwright/test, and a running server.
# Set E2E_BASE_URL to override the target (default: http://127.0.0.1:7777).
#
#   make web/e2e              – run the visual regression tests
#   make web/e2e-update       – regenerate baselines from scratch
#
web/e2e:
	@echo "📱 Running mobile visual regression tests…"
	$(PW) test --config web/e2e/playwright.config.ts web/e2e/mobile.spec.ts

web/e2e-update:
	@echo "📱 Regenerating baselines…"
	rm -f web/e2e/baselines/*.png
	$(PW) test --config web/e2e/playwright.config.ts web/e2e/mobile.spec.ts || true
	@echo "✅ Baselines regenerated. Review web/e2e/baselines/*.png and commit."

# ── PWA installability audit (Lighthouse CI) ────────────────────
# Requires: Node.js, @lhci/cli, Chromium, and a running server.
# Set LIGHTHOUSE_BASE_URL to override the target (default: http://127.0.0.1:7777).
#
#   make web/lighthouse         – collect + assert PWA score ≥ 90
#   make web/lighthouse-server  – run an autorun server (useful for CI)
#
LHCI        := node_modules/.bin/lhci
LIGHTHOUSE_BASE_URL ?= http://127.0.0.1:7777

web/lighthouse:
	@echo "🔍 Running Lighthouse PWA audit against $(LIGHTHOUSE_BASE_URL)…"
	LIGHTHOUSE_BASE_URL="$(LIGHTHOUSE_BASE_URL)" $(LHCI) autorun

# Convenience target: runs LHCI in server mode (collects from one URL then asserts).
# Useful for CI pipelines where you control the browser environment.
web/lighthouse-server:
	@echo "🔍 Running Lighthouse CI server mode…"
	$(LHCI) server &
	sleep 2
	$(LHCI) collect --url="$(LIGHTHOUSE_BASE_URL)" --settings.chrome-flags="--headless --no-sandbox"
	$(LHCI) assert --preset=lighthouse:recommended --assertions.pwa=error --assertions.categories.pwa.minScore=0.9
	kill %1 || true