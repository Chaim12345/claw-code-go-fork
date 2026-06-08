// Package errors defines sentinel error values and categories used across
// claw-code-go. Centralising them lets callers distinguish failure modes
// with errors.Is / errors.As instead of fragile string matching, and gives
// the CLI a single place to render uniform error output.
package errors

import (
	"errors"
	"fmt"
)

// Sentinel error categories. Use Is/As to test:
//   if errors.Is(err, errors.ErrPathTraversal) { ... }
var (
	// ErrPathTraversal is returned when a tool operation is blocked because
	// the target path escapes the allowed directory set.
	ErrPathTraversal = errors.New("path traversal blocked")

	// ErrFileTooLarge is returned when read_file / grep is asked to process
	// a file that exceeds the configured size limit.
	ErrFileTooLarge = errors.New("file too large")

	// ErrPermission is returned when a tool operation is denied by the
	// permission system (auto/ask/deny modes).
	ErrPermission = errors.New("permission denied")

	// ErrBashBlocked is returned when a bash invocation is rejected by
	// the pattern-blocklist before any command runs.
	ErrBashBlocked = errors.New("bash command blocked")

	// ErrProvider is returned when an AI provider stream fails, authenticates
	// incorrectly, or returns a malformed response.
	ErrProvider = errors.New("provider error")

	// ErrSession is returned when a session cannot be loaded, parsed, or
	// saved.
	ErrSession = errors.New("session error")

	// ErrAuth is returned when credentials are missing, expired, or
	// cannot be refreshed.
	ErrAuth = errors.New("authentication error")

	// ErrContext is returned when the agentic loop aborts because the
	// caller's context was cancelled.
	ErrContext = errors.New("context cancelled")

	// ErrToolNotFound is returned when a tool name doesn't match any
	// registered tool.
	ErrToolNotFound = errors.New("tool not found")

	// ErrInvalidInput is returned when tool input fails validation
	// (missing required fields, wrong type, etc.).
	ErrInvalidInput = errors.New("invalid input")
)

// RestrictedError wraps any error to signal that the operation should be
// escalated to the user as a permission prompt rather than hard-denied.
// The permission layer checks for this type: if a tool returns a
// RestrictedError, the loop emits a DecisionAsk instead of failing outright.
type RestrictedError struct {
	Err error
}

func (e *RestrictedError) Error() string { return e.Err.Error() }
func (e *RestrictedError) Unwrap() error { return e.Err }

// Restricted wraps err in a RestrictedError. Returns nil if err is nil.
func Restricted(err error) error {
	if err == nil {
		return nil
	}
	return &RestrictedError{Err: err}
}

// IsRestricted reports whether err is (or wraps) a RestrictedError.
func IsRestricted(err error) bool {
	var re *RestrictedError
	return errors.As(err, &re)
}

// Wrapped is a small helper to produce an error that is both formatted and
// has Is/As match against a sentinel.
func Wrapped(sentinel error, format string, args ...any) error {
	return fmt.Errorf("%w: %s", sentinel, fmt.Sprintf(format, args...))
}
