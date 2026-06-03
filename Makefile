.PHONY: web/lint web/lint-html web/lint-js web/lint-fix

# ── Web asset linting ─────────────────────────────────────────────
# Requires: Node.js and npm dependencies (html-validate, eslint).
# Run `npm ci` to install deps before using these targets.
# Uses direct node_modules/.bin paths to avoid npx network overhead.

ESLINT   := node_modules/.bin/eslint
HTMLV    := node_modules/.bin/html-validate

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