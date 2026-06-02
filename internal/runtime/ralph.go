package runtime

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"
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
