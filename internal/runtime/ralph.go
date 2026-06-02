package runtime

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"

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
}

// Defaults for the Ralph loop.
const (
	DefaultRalphSpecPath      = "ROADMAP.md"
	DefaultRalphDoneSentinel  = "RALPH_DONE"
	DefaultRalphMaxIterations = 50
	MaxRalphIterations        = 500
	defaultRalphPromptTemplate = `You are an autonomous coding agent running in a Ralph loop.

Iteration {{.Iteration}} of {{.MaxIterations}}.

A spec file lives at {{.SpecPath}}. Read it. It contains a list of items (tasks, fixes, features) the codebase should implement. Your job, in this iteration, is to make progress on exactly one item:

  1. Read the spec file at {{.SpecPath}}.
  2. Find the next item that is NOT marked done (e.g. checkbox is [ ] or status is TODO/pending).
  3. Implement it: read, edit, build, test. Use the full tool set.
  4. Verify the change actually works (run the test, build the binary, or whatever proves the item is done).
  5. Mark the item done in the spec file (e.g. change [ ] to [x], or add a "Done" annotation, or move it under a "Completed" heading).
  6. Commit the change with a clear conventional-commit message.
  7. Output a one-paragraph summary of what you did.

If all items in the spec are already done, output the literal sentinel {{.Sentinel}} on its own line and stop.

If you hit a real blocker (a missing dependency, a question only the human can answer, a contradiction in the spec), document it under a "## Blockers" heading at the bottom of the spec file with enough detail for the next iteration to pick up — then exit normally. The next iteration will start fresh with your note.

The spec file is your persistent memory. Each iteration you will get a fresh context. Do not assume anything from prior turns survives; write everything important to the spec.

The spec content is shown below for convenience — it is the live file, so re-read it from disk if you suspect it has changed.

--- BEGIN SPEC ({{.SpecPath}}) ---

{{.Spec}}

--- END SPEC ---`
)

// DefaultRalphConfig returns a RalphConfig with all defaults
// populated. Callers may override individual fields after.
func DefaultRalphConfig() RalphConfig {
	return RalphConfig{
		SpecPath:       DefaultRalphSpecPath,
		MaxIterations:  DefaultRalphMaxIterations,
		DoneSentinel:   DefaultRalphDoneSentinel,
		PromptTemplate: defaultRalphPromptTemplate,
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
	specBody, err := readRalphSpec(cfg.SpecPath)
	if err != nil {
		return RalphVerdictContinue, "", fmt.Errorf("read spec: %w", err)
	}
	prompt, err := RenderRalphPrompt(cfg, specBody, iteration, maxIter)
	if err != nil {
		return RalphVerdictContinue, "", err
	}
	if err := loop.SendMessage(ctx, prompt); err != nil {
		return RalphVerdictContinue, "", fmt.Errorf("send: %w", err)
	}
	// SendMessage populates loop.Session.Messages with the new
	// turn. The final assistant text is the last assistant
	// message's text content.
	finalText := lastAssistantText(loop.Session.Messages)
	// Re-read the spec in case the model edited it.
	updated, err := readRalphSpec(cfg.SpecPath)
	if err != nil {
		// Spec disappeared? Treat as done (we'll exit on next check).
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

// IterFn is the per-iteration hook used by RunRalphLoopWithIter.
// Production code calls RunRalphLoop; tests inject a fake.
type IterFn func(ctx context.Context, iteration, maxIter int) (RalphVerdictKind, string, error)

// RunRalphLoopWithIter drives the Ralph loop with an injected
// iteration function. Most callers want RunRalphLoop; this
// version is exposed for testability.
func RunRalphLoopWithIter(ctx context.Context, cfg RalphConfig, iter IterFn) error {
	maxIter := cfg.MaxIterations
	if maxIter <= 0 {
		maxIter = DefaultRalphMaxIterations
	}
	if maxIter > MaxRalphIterations {
		maxIter = MaxRalphIterations
	}
	for i := 1; i <= maxIter; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		verdict, text, err := iter(ctx, i, maxIter)
		if err != nil {
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
	iter := func(ctx context.Context, i, max int) (RalphVerdictKind, string, error) {
		// Reset the provider session so each iteration gets a
		// fresh context. The spec file is the only persistent
		// memory.
		if resetter, ok := loop.Client.(interface{ ResetSession() error }); ok {
			_ = resetter.ResetSession()
		}
		return RalphOneIteration(ctx, loop, cfg, i, max)
	}
	return RunRalphLoopWithIter(ctx, cfg, iter)
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
		Spec          string
		SpecPath      string
		Iteration     int
		MaxIterations int
		Sentinel      string
	}{
		Spec:          specBody,
		SpecPath:      cfg.SpecPath,
		Iteration:     iteration,
		MaxIterations: maxIter,
		Sentinel:      cfg.DoneSentinel,
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return buf.String(), nil
}
