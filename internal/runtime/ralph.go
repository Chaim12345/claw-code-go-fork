package runtime

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/template"
	"time"

	"claw-code-go/internal/api"
)

// RalphConfig configures a single Ralph loop run.
type RalphConfig struct {
	// SpecPath is the path to the spec/roadmap file the agent
	// reads on each iteration. Required.
	SpecPath string

	// MaxIterations caps the number of fresh-context agent calls.
	// 0 means use DefaultRalphMaxIterations. Hard-capped at
	// MaxRalphIterations to prevent runaway runs.
	MaxIterations int

	// DoneSentinel is the literal token the model must emit on its
	// own line to signal "all items in the spec are complete."
	// Default: "RALPH_DONE"
	DoneSentinel string

	// PromptTemplate is the Go text/template body sent to the
	// model on every iteration. Receives {{.Spec}} and
	// {{.Iteration}} and {{.MaxIterations}} and {{.Sentinel}}.
	// Default: a built-in template (defined separately).
	PromptTemplate string

	// Model optionally overrides the loop's model. Empty means
	// use the loop's configured model.
	Model string

	// WorkingDir is the cwd the model sees. Empty means inherit.
	WorkingDir string

	// SelfDebug enables a one-shot "self-debug" pass when an
	// iteration has exhausted all retries. The loop captures the
	// last error, mutates the prompt template to inject a
	// "## Self-debug" section that names the error, and runs one
	// more iteration. If that debug iteration also fails (or
	// reports blocked), the loop gives up. If it succeeds (verdict
	// Continue or Done), the loop proceeds as if the original
	// iteration had worked.
	//
	// This is the answer to "the agent got an exit-127 from bash
	// and now needs to fix its own PATH/environment." Instead of
	// dying, ralph sees the error and is tasked with debugging.
	//
	// Default: true (opt-out for tests).
	SelfDebug bool

	// LastError is set by the loop when SelfDebug triggers. It is
	// not part of the public constructor; the loop mutates the
	// config in-flight to inject the error into the next prompt.
	// Users should not set this directly.
	LastError string

	// DeltaMode enables incremental context: each iteration sends
	// only the new spec prompt instead of the full system+tools+
	// spec context. Requires a provider with server-side session
	// support (e.g., deepseek web chat API). When enabled, the
	// provider's ResetSession() is NOT called between iterations
	// so the server maintains the full conversation.
	DeltaMode bool

	// IterationTimeout caps how long a single iteration may run.
	// 0 means use DefaultRalphIterationTimeout. Prevents the model
	// from hanging indefinitely on a single turn.
	IterationTimeout time.Duration

	// ProgressNotes is injected into the next iteration's prompt
	// so the model knows what happened in prior iterations. The
	// loop populates this automatically after each iteration when
	// ProgressTracking is enabled.
	ProgressNotes string

	// ProgressTracking enables per-iteration git diff summary and
	// tool-call tracking. When true, after each iteration the loop
	// captures git diff stats and checks whether the model only
	// read files without writing. Results are injected into
	// ProgressNotes for the next iteration.
	ProgressTracking bool
}

// Defaults for the Ralph loop.
const (
	DefaultRalphSpecPath          = "ROADMAP.md"
	DefaultRalphDoneSentinel      = "RALPH_DONE"
	DefaultRalphMaxIterations     = 50
	MaxRalphIterations            = 500
	DefaultRalphIterationTimeout  = 5 * time.Minute
	defaultRalphPromptTemplate = `Ralph autonomous agent — iteration {{.Iteration}}/{{.MaxIterations}}.

Read spec at {{.SpecPath}}. Pick the NEXT unchecked item. Implement it NOW.

WORKFLOW (strict order):
1. Read spec file, identify next [ ] item
2. Read ONLY the files you must change (batch all reads)
3. Write ALL edits (batch all file_edit/write_file calls)
4. Run build+test (ONE bash call)
5. Mark item done in spec ([x])
6. Commit

HARD RULES:
- NEVER read a file without editing it in the SAME iteration
- NEVER read files one-at-a-time — batch ALL reads first
- NEVER edit one field at a time — batch ALL edits
- NEVER leave an iteration without at least one file change
- If you cannot implement an item, mark it [BLOCKED] and move on
- Output RALPH_DONE on its own line when ALL items are checked

Spec ({{.SpecPath}}):

{{.Spec}}

{{if .ProgressNotes}}
LAST ITERATION:
{{.ProgressNotes}}
{{end}}
{{if .LastError}}
SELF-DEBUG — previous iteration crashed:
{{.LastError}}
Fix the root cause, then continue on the spec.
{{end}}`
)

// DefaultRalphConfig returns a RalphConfig with all defaults
// populated. Callers may override individual fields after.
func DefaultRalphConfig() RalphConfig {
	return RalphConfig{
		SpecPath:         DefaultRalphSpecPath,
		MaxIterations:    DefaultRalphMaxIterations,
		DoneSentinel:     DefaultRalphDoneSentinel,
		PromptTemplate:   defaultRalphPromptTemplate,
		SelfDebug:        true,
		IterationTimeout: DefaultRalphIterationTimeout,
		ProgressTracking: true,
	}
}

// readRalphSpec reads the spec file and returns its contents.
// Returns os.ErrNotExist if the file does not exist.
func readRalphSpec(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// RalphDetectedSentinel returns true if the sentinel token appears
// on its own line (allowing leading/trailing whitespace). False
// otherwise. Case-sensitive.
//
// This is the completion signal — the model emits it only when it
// believes the spec is fully implemented.
func RalphDetectedSentinel(text, sentinel string) bool {
	for _, line := range strings.Split(text, "\n") {
		if strings.TrimSpace(line) == sentinel {
			return true
		}
	}
	return false
}

// A spec is "complete" when every line matching a checkbox or
// TODO marker shows it done. We are conservative: any line that
// looks like an open task (- [ ], - [TODO], * [ ], 1. [ ], etc.)
// makes the spec incomplete.
//
// We do NOT use this as the primary completion signal (the
// sentinel is). It's a safety net — if the model never emits
// the sentinel, we can still detect a spec where every line is
// marked done.
var openTaskPattern = regexp.MustCompile(`(?m)^\s*(?:-|\*|\d+\.)\s+\[(?:\s|todo|pending|wip)\]`)

// RalphSpecIsComplete returns true if no open task markers are
// present in the spec body. Empty or all-checked specs return true.
func RalphSpecIsComplete(spec string) bool {
	return !openTaskPattern.MatchString(spec)
}

// ralphToolSummary counts read vs write tool calls in the session
// messages from the most recent iteration. Returns (reads, writes).
func ralphToolSummary(messages []api.Message) (reads, writes int) {
	for _, msg := range messages {
		if msg.Role != "assistant" {
			continue
		}
		for _, cb := range msg.Content {
			if cb.Type != "tool_use" {
				continue
			}
			switch cb.Name {
			case "read_file", "glob", "grep":
				reads++
			case "write_file", "file_edit", "bash":
				writes++
			}
		}
	}
	return
}

// ralphGitDiff returns the git diff --stat output for the working
// directory, or an empty string if git is unavailable.
func ralphGitDiff() string {
	out, err := exec.Command("git", "diff", "--stat").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// ralphBuildProgressNotes generates a progress summary for injection
// into the next iteration's prompt. It includes git diff stats and
// a warning if the iteration only read files without writing.
func ralphBuildProgressNotes(reads, writes int, gitDiff string) string {
	var b strings.Builder
	if gitDiff != "" {
		b.WriteString("Git changes:\n")
		b.WriteString(gitDiff)
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, "Tool calls: %d reads, %d writes\n", reads, writes)
	if writes == 0 && reads > 0 {
		b.WriteString("WARNING: last iteration only READ files — no edits made. You MUST write code this iteration.\n")
	}
	return b.String()
}

// RalphVerdictKind is the outcome of a single iteration.
//
// (Originally named `RalphVerdict` in the plan; renamed to avoid
// the type/function name collision with the RalphVerdict()
// decision function below.)
type RalphVerdictKind int

const (
	// RalphVerdictContinue means the spec still has open work
	// and the model did not emit the sentinel.
	RalphVerdictContinue RalphVerdictKind = iota
	// RalphVerdictDone means either the model emitted the
	// sentinel, or every line in the spec looks done.
	RalphVerdictDone
	// RalphVerdictBlocked means the model said it was blocked
	// (e.g. output matched a "Blockers" header in its response
	// text). The loop should record this and stop gracefully.
	RalphVerdictBlocked
)

// RalphVerdict decides what to do next based on the final
// assistant text and the current spec body.
//
// Priority:
//  1. Sentinel on its own line → Done
//  2. Spec is fully checked off (or empty) → Done
//  3. Otherwise → Continue
func RalphVerdict(finalText, specBody string, cfg RalphConfig) RalphVerdictKind {
	if RalphDetectedSentinel(finalText, cfg.DoneSentinel) {
		return RalphVerdictDone
	}
	if RalphSpecIsComplete(specBody) {
		return RalphVerdictDone
	}
	return RalphVerdictContinue
}

// RalphOneIteration runs ONE fresh-context agent turn. It:
//  1. Reads the spec from disk.
//  2. Renders the prompt with the current spec body.
//  3. Calls loop.SendMessage(ctx, rendered).
//  4. Re-reads the spec (the model may have edited it).
//  5. Returns (verdict, finalText, error).
//
// The loop's conversation history grows across iterations
// (because we use the same loop), which is intentional — it
// keeps the model oriented. To get true fresh-context behavior,
// call ResetSession() on the client before each iteration (see
// RunRalphLoop for the wiring).
func RalphOneIteration(ctx context.Context, loop *ConversationLoop, cfg RalphConfig, iteration, maxIter int) (RalphVerdictKind, string, error) {
	timeout := cfg.IterationTimeout
	if timeout <= 0 {
		timeout = DefaultRalphIterationTimeout
	}
	iterCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	specBody, err := readRalphSpec(cfg.SpecPath)
	if err != nil {
		return RalphVerdictContinue, "", fmt.Errorf("read spec: %w", err)
	}
	prompt, err := RenderRalphPrompt(cfg, specBody, iteration, maxIter)
	if err != nil {
		return RalphVerdictContinue, "", err
	}
	if err := loop.SendMessage(iterCtx, prompt); err != nil {
		return RalphVerdictContinue, "", fmt.Errorf("send: %w", err)
	}
	finalText := lastAssistantText(loop.Session.Messages)
	updated, err := readRalphSpec(cfg.SpecPath)
	if err != nil {
		updated = specBody
	}
	return RalphVerdict(finalText, updated, cfg), finalText, nil
}

// lastAssistantText extracts the final text from the most recent
// assistant message in the session, or "" if none.
func lastAssistantText(messages []api.Message) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role != "assistant" {
			continue
		}
		var sb strings.Builder
		for _, cb := range messages[i].Content {
			if cb.Type == "text" {
				sb.WriteString(cb.Text)
			}
		}
		return sb.String()
	}
	return ""
}

// ErrRalphMaxIterations is returned by RunRalphLoop when the loop
// hits MaxIterations without the spec being completed.
var ErrRalphMaxIterations = errors.New("ralph: max iterations reached without completion")

// ErrRalphGiveUp is returned when every retry+rotation attempt has
// failed for the same iteration. The loop exits with this error so
// callers can distinguish "spec not done yet" from "the upstream
// LLM is genuinely unreachable."
var ErrRalphGiveUp = errors.New("ralph: gave up after exhausting retries+rotations")

// IterFn is the per-iteration hook used by RunRalphLoopWithIter.
// Production code calls RunRalphLoop; tests inject a fake.
type IterFn func(ctx context.Context, iteration, maxIter int) (RalphVerdictKind, string, error)

// ralphRetryConfig controls how RunRalphLoopWithIter handles
// transient per-iteration errors and "stupid loop" detection. All
// fields have sensible defaults; tests can override them via the
// ralphRetryEnv hook.
type ralphRetryConfig struct {
	// maxRetriesPerIter is how many times to retry a single
	// iteration on a transient error before giving up. The deepseek
	// provider already retries 3 times internally per stream; this
	// layer handles errors that survive that (e.g. session
	// poisoning, sustained rate limit) by re-running the whole
	// iteration with a fresh context.
	maxRetriesPerIter int
	// baseBackoff is the first retry delay; subsequent retries
	// double it (capped at maxBackoff).
	baseBackoff time.Duration
	maxBackoff  time.Duration
	// stupidLoopThreshold: if the same finalText repeats this many
	// times in a row, the loop assumes the provider is stuck (e.g.
	// rate limit mid-response, model echoing itself) and forces a
	// backoff before the next attempt.
	stupidLoopThreshold int
}

var defaultRalphRetry = ralphRetryConfig{
	maxRetriesPerIter:   5,
	baseBackoff:         30 * time.Second,
	maxBackoff:          5 * time.Minute,
	stupidLoopThreshold: 2,
}

// looksLikeTransientRalphError matches errors that warrant a retry
// of the whole iteration. We deliberately err on the side of
// retrying — a wasted 30s sleep is much cheaper than a human
// restarting the loop.
func looksLikeTransientRalphError(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	// Context cancellation is never transient.
	if strings.Contains(s, "context canceled") || strings.Contains(s, "context deadline exceeded") {
		return false
	}
	hints := []string{
		"stream error",
		"server is busy",
		"too frequent",
		"rate limit",
		"rate_limit_reached",
		"429",
		"500",
		"502",
		"503",
		"504",
		"connection reset",
		"connection refused",
		"timeout",
		"eof",
		"temporarily unavailable",
		"no session id",
		"pow",
		"length limit",
		"prompt too large",
		"context length exceeded",
		"maximum context length",
		"context_length_exceeded",
		"input is too long",
	}
	for _, h := range hints {
		if strings.Contains(s, h) {
			return true
		}
	}
	return false
}

// RunRalphLoopWithIter drives the Ralph loop with an injected
// iteration function. Most callers want RunRalphLoop; this
// version is exposed for testability.
//
// cfg is passed by pointer so the self-debug pass can mutate
// cfg.LastError to inject the last error into the next prompt.
//
// Resilience features (all opt-out via test override):
//   - Per-iteration transient errors are retried with exponential
//     backoff (30s → 60s → 120s → 240s → 300s).
//   - When the same finalText repeats stupidLoopThreshold times in
//     a row, the loop assumes the provider is stuck and backs off
//     before the next attempt. The deepseek provider's own token
//     rotation is what unsticks it.
//   - After maxRetriesPerIter consecutive transient failures the
//     loop runs ONE self-debug pass (if cfg.SelfDebug is true):
//     the next iter gets a "## Self-debug" prompt section naming
//     the error and is tasked with fixing the root cause. Only
//     if the debug pass also fails does the loop return
//     ErrRalphGiveUp.
func RunRalphLoopWithIter(ctx context.Context, cfg *RalphConfig, iter IterFn) error {
	maxIter := cfg.MaxIterations
	if maxIter <= 0 {
		maxIter = DefaultRalphMaxIterations
	}
	if maxIter > MaxRalphIterations {
		maxIter = MaxRalphIterations
	}
	rc := defaultRalphRetry

	// Track the last few finalText outputs to detect stupid loops.
	recent := make([]string, 0, rc.stupidLoopThreshold)
	consecutiveStuck := 0

	for i := 1; i <= maxIter; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		// Stupid-loop guard: if the model has produced the same
		// finalText N times in a row, back off before trying again.
		// This is the "toggling 2 keys" case: the provider is rate
		// limited and the model is echoing itself.
		if consecutiveStuck >= rc.stupidLoopThreshold {
			backoff := rc.baseBackoff
			fmt.Fprintf(os.Stderr, "[ralph] iter %d: stuck on same output %d times; backing off %v before next attempt\n", i, consecutiveStuck, backoff)
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return ctx.Err()
			}
			consecutiveStuck = 0
		}

		var (
			verdict RalphVerdictKind
			text    string
			err     error
			lastErr error
		)
		for attempt := 0; attempt <= rc.maxRetriesPerIter; attempt++ {
			if err := ctx.Err(); err != nil {
				return err
			}
			verdict, text, err = iter(ctx, i, maxIter)
			if err == nil {
				break
			}
			lastErr = err
			if !looksLikeTransientRalphError(err) {
				// Non-transient: skip retries, fall through to the
				// self-debug pass below.
				break
			}
			if attempt == rc.maxRetriesPerIter {
				// Exhausted transient retries; fall through to the
				// self-debug pass.
				break
			}
			backoff := rc.baseBackoff << attempt
			if backoff > rc.maxBackoff {
				backoff = rc.maxBackoff
			}
			fmt.Fprintf(os.Stderr, "[ralph] iter %d: transient error (attempt %d/%d), retrying in %v: %v\n", i, attempt+1, rc.maxRetriesPerIter+1, backoff, err)
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// Self-debug pass: if the iter failed (either non-transient
		// or transient-after-exhaustion) and SelfDebug is on, give
		// the model ONE more iteration with a meta-prompt that names
		// the error and tells it to fix the root cause. The
		// captured cfg pointer lets us inject the error string.
		if err != nil && cfg.SelfDebug {
			fmt.Fprintf(os.Stderr, "[ralph] iter %d: entering self-debug pass (last error: %v)\n", i, err)
			cfg.LastError = err.Error()
			// Brief backoff so the model isn't immediately thrashed.
			select {
			case <-time.After(10 * time.Second):
			case <-ctx.Done():
				return ctx.Err()
			}
			debugVerdict, debugText, debugErr := iter(ctx, i, maxIter)
			fmt.Fprintf(os.Stderr, "[ralph] iter %d self-debug: verdict=%v text=%q err=%v\n", i, debugVerdict, debugText, debugErr)
			// Clear LastError so subsequent iters get a clean prompt.
			cfg.LastError = ""
			if debugErr == nil {
				verdict, text, err = debugVerdict, debugText, nil
			} else {
				// Self-debug also failed. Fall through to give up.
				fmt.Fprintf(os.Stderr, "[ralph] iter %d: self-debug also failed; giving up\n", i)
				if looksLikeTransientRalphError(err) {
					return fmt.Errorf("%w (iter=%d, last_err=%v, debug_err=%v)", ErrRalphGiveUp, i, lastErr, debugErr)
				}
				return fmt.Errorf("%w (iter=%d, last_err=%v)", ErrRalphGiveUp, i, err)
			}
		} else if err != nil {
			// SelfDebug off or last error was already terminal.
			if looksLikeTransientRalphError(err) {
				return fmt.Errorf("%w (iter=%d, last_err=%v)", ErrRalphGiveUp, i, err)
			}
			return fmt.Errorf("iteration %d: %w", i, err)
		}

		fmt.Fprintf(os.Stderr, "[ralph] iter %d/%d: verdict=%v text=%q\n", i, maxIter, verdict, text)
		if verdict == RalphVerdictDone {
			return nil
		}
		if verdict == RalphVerdictBlocked {
			// Documented exit: the model said it's blocked.
			// Treat as done so the loop doesn't run forever.
			fmt.Fprintf(os.Stderr, "[ralph] iter %d reported blocked; exiting cleanly\n", i)
			return nil
		}

		// Stupid-loop bookkeeping. We only count "echo" of the
		// same text; an empty or whitespace-only text never counts
		// (that's an iteration that produced no output, not a stuck
		// model).
		trimmed := strings.TrimSpace(text)
		if trimmed == "" {
			consecutiveStuck = 0
		} else if len(recent) > 0 && recent[len(recent)-1] == trimmed {
			consecutiveStuck++
		} else {
			consecutiveStuck = 1
		}
		recent = append(recent, trimmed)
		if len(recent) > rc.stupidLoopThreshold {
			recent = recent[len(recent)-rc.stupidLoopThreshold:]
		}
	}
	return fmt.Errorf("%w (max=%d)", ErrRalphMaxIterations, cfg.MaxIterations)
}

// RunRalphLoop is the production entry point. It drives a
// ConversationLoop through the Ralph pattern, reading the spec
// file from disk on each iteration and giving the model a fresh
// context (we reset the provider session between iterations so
// the model genuinely starts each turn with the spec prompt as
// its only context).
//
// Stop conditions:
//   - Sentinel detected in the model's final text → return nil
//   - Spec is fully checked off → return nil
//   - MaxIterations reached → return ErrRalphMaxIterations
//   - Context cancelled → return ctx.Err()
//   - Per-iteration error → return wrapped error
func RunRalphLoop(ctx context.Context, loop *ConversationLoop, cfg RalphConfig) error {
	if cfg.SpecPath == "" {
		cfg.SpecPath = DefaultRalphSpecPath
	}
	if _, err := readRalphSpec(cfg.SpecPath); err != nil {
		return fmt.Errorf("read spec %q: %w", cfg.SpecPath, err)
	}
	// Capture cfg by pointer so the self-debug pass (inside
	// RunRalphLoopWithIter) can mutate cfg.LastError to inject the
	// last error into the next prompt.
	cfgPtr := &cfg
	iter := func(ctx context.Context, i, max int) (RalphVerdictKind, string, error) {
		if !cfgPtr.DeltaMode {
			if resetter, ok := loop.Client.(api.SessionResetter); ok {
				resetter.ResetSession()
			}
		}
		loop.ClearSession()
		v, text, err := RalphOneIteration(ctx, loop, *cfgPtr, i, max)
		if err == nil && cfgPtr.ProgressTracking {
			reads, writes := ralphToolSummary(loop.Session.Messages)
			diff := ralphGitDiff()
			cfgPtr.ProgressNotes = ralphBuildProgressNotes(reads, writes, diff)
		}
		return v, text, err
	}
	return RunRalphLoopWithIter(ctx, cfgPtr, iter)
}

// RenderRalphPrompt executes the prompt template with the spec
// body, current iteration, max iterations, and sentinel. Returns
// an error if the template is invalid.
func RenderRalphPrompt(cfg RalphConfig, specBody string, iteration, maxIter int) (string, error) {
	tpl, err := template.New("ralph").Parse(cfg.PromptTemplate)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}
	data := struct {
		Spec           string
		SpecPath       string
		Iteration      int
		MaxIterations  int
		Sentinel       string
		LastError      string
		ProgressNotes  string
	}{
		Spec:           specBody,
		SpecPath:       cfg.SpecPath,
		Iteration:      iteration,
		MaxIterations:  maxIter,
		Sentinel:       cfg.DoneSentinel,
		LastError:      cfg.LastError,
		ProgressNotes:  cfg.ProgressNotes,
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return buf.String(), nil
}
