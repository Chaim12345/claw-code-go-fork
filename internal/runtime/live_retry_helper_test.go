//go:build live

package runtime

import (
	"os"
	"strings"
	"time"
)

func getenv(k string) string { return os.Getenv(k) }

// sleepForRateLimit waits a short, fixed time at the start of a
// test to give any previous test's rate-limit budget time to relax.
// The DeepSeek web API throttles per-token after a burst of
// tool-call-heavy requests; the first tool call in a fresh session
// sometimes lands on a rate-limited / empty response. The 3s sleep
// stabilises the live suite significantly when tests run back-to-
// back. Set DEEPSEEK_LIVE_NO_SLEEP=1 to disable (e.g. for
// debugging, when you want raw timing).
func sleepForRateLimit() {
	if getenv("DEEPSEEK_LIVE_NO_SLEEP") == "1" {
		return
	}
	time.Sleep(3 * time.Second)
}

// isTransientStreamErr reports whether s looks like a deepseek
// rate-limit / empty-response / network error. The test bodies
// can use this to decide whether a failure is environmental
// (skip with a note) versus a real regression (fail).
func isTransientStreamErr(s string) bool {
	s = strings.ToLower(s)
	hints := []string{
		"rate", "limit", "throttle", "temporarily", "try again",
		"empty response", "connection reset", "connection refused",
		"http 429", "http 502", "http 503", "http 504",
	}
	for _, h := range hints {
		if strings.Contains(s, h) {
			return true
		}
	}
	return false
}
