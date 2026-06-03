package web

import (
	"os/exec"
	"testing"
)

// TestWebAssetsLint shells out to the lint tools if they are available.
// This test is wired into `go test ./...` so that CI catches lint regressions.
// The test is skipped when Node.js or the lint tools are not installed.
func TestWebAssetsLint(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping web asset lint in short mode")
	}

	// Require Node.js to be available for linting.
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node not found in PATH; skipping web asset lint")
	}

	// Run html-validate on HTML assets (uses direct node_modules/.bin path).
	t.Run("html-validate", func(t *testing.T) {
		cmd := exec.Command("node_modules/.bin/html-validate",
			"--config", ".htmlvalidate.json",
			"internal/web/static/chat.html",
			"internal/web/static/index.html",
			"internal/web/static/error.html",
			"internal/web/static/safe-area-test.html",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("html-validate failed:\n%s", string(out))
		}
	})

	// Run eslint on JS assets (uses direct node_modules/.bin path).
	t.Run("eslint", func(t *testing.T) {
		cmd := exec.Command("node_modules/.bin/eslint",
			"--config", "eslint.config.js",
			"internal/web/static/js/chat.js",
			"internal/web/static/sw.js",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("eslint failed:\n%s", string(out))
		}
	})
}