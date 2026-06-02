package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"claw-code-go/internal/runtime"
)

func toModel(t *testing.T, raw tea.Model) Model {
	t.Helper()
	m, ok := raw.(Model)
	if !ok {
		t.Fatalf("expected tui.Model, got %T", raw)
	}
	return m
}

func feed(t *testing.T, m Model, msgs ...tea.Msg) Model {
	t.Helper()
	for _, msg := range msgs {
		nm, _ := m.Update(msg)
		m = toModel(t, nm)
	}
	return m
}

// TestStreamTextFinalPreservesToolIndicators reproduces the user
// report: the conversation loop streams text deltas that contain
// raw <invoke> syntax, a tool runs (◆ + ✓ indicators get appended
// to streamBuf), then the conversation loop sends the cleaned
// text as TurnEventTextFinal. Before the fix, the handler
// overwrote streamBuf entirely, erasing the tool indicators AND
// dropping any text outside the cleaned window. After the fix, the
// text portion is replaced and the tool indicators are preserved.
func TestStreamTextFinalPreservesToolIndicators(t *testing.T) {
	cfg := &runtime.Config{Model: "expert"}
	loop := runtime.NewConversationLoop(cfg, runtime.NewNoAuthClient())
	m := NewModel(cfg, loop)

	// Step 1: stream raw text deltas with embedded <invoke>.
	deltas := []string{
		"I'll map the codebase. ",
		"<invoke name=\"bash\">\n<parameter name=\"command\">ls -la</parameter>\n</invoke>\n",
		"\nClaude: Let me read a few more key files to complete the mapping.\n",
		"\nFINISHED", // model literally emitted this in its text
	}
	for _, d := range deltas {
		nm, _ := m.Update(streamDeltaMsg{text: d})
		m = toModel(t, nm)
	}

	// Step 2: a tool runs. These two messages append indicators.
	m = feed(t, m, streamToolMsg{name: "bash", input: `{"command":"ls -la"}`})
	m = feed(t, m, streamToolDoneMsg{name: "bash", result: "file1.go\nfile2.go"})

	// Step 3: cleaned text arrives. <invoke> block is gone, but
	// the post-text "Claude: Let me read..." should remain.
	cleaned := "I'll map the codebase. \n\nClaude: Let me read a few more key files to complete the mapping."
	m = feed(t, m, streamTextFinalMsg{text: cleaned})

	// Step 4: turn ends, commit to viewBuf.
	m = feed(t, m, streamDoneMsg{})

	view := m.viewBuf
	checks := []struct {
		desc, want, dontWant string
	}{
		{"pre-text preserved", "I'll map the codebase", ""},
		{"post-text preserved", "read a few more key files", ""},
		{"raw <invoke> stripped", "", "<invoke"},
		{"raw </invoke> stripped", "", "</invoke>"},
		{"tool running indicator preserved", "bash", ""},
		{"FINISHED stripped from model output", "", "FINISHED"},
	}
	for _, c := range checks {
		if c.want != "" && !strings.Contains(view, c.want) {
			t.Errorf("%s: viewBuf missing %q\nviewBuf=%q", c.desc, c.want, view)
		}
		if c.dontWant != "" && strings.Contains(view, c.dontWant) {
			t.Errorf("%s: viewBuf still contains %q\nviewBuf=%q", c.desc, c.dontWant, view)
		}
	}
}

// TestStreamTextFinalNoToolsOverwritesEntirely: when a turn has
// no tool calls, streamTextLen is 0, so the prefix replacement
// degenerates to a full overwrite (the original behaviour for
// tool-free turns).
func TestStreamTextFinalNoToolsOverwritesEntirely(t *testing.T) {
	cfg := &runtime.Config{Model: "expert"}
	loop := runtime.NewConversationLoop(cfg, runtime.NewNoAuthClient())
	m := NewModel(cfg, loop)

	// Pure text, no tool calls.
	for _, d := range []string{"Hello, ", "world!"} {
		nm, _ := m.Update(streamDeltaMsg{text: d})
		m = toModel(t, nm)
	}
	// Cleaned text = same thing (no tool calls to strip).
	m = feed(t, m, streamTextFinalMsg{text: "Hello, world!"})
	m = feed(t, m, streamDoneMsg{})

	if !strings.Contains(m.viewBuf, "Hello, world!") {
		t.Errorf("expected 'Hello, world!' in viewBuf, got %q", m.viewBuf)
	}
}

// TestStreamTextFinalBoundaryResetsAcrossTurns: streamTextLen
// must not leak across turns. If turn N had a tool (so
// streamTextLen was set) and turn N+1 does not, the boundary from
// turn N must not cause TextFinal in turn N+1 to do a prefix
// replacement against a stale offset.
func TestStreamTextFinalBoundaryResetsAcrossTurns(t *testing.T) {
	cfg := &runtime.Config{Model: "expert"}
	loop := runtime.NewConversationLoop(cfg, runtime.NewNoAuthClient())
	m := NewModel(cfg, loop)

	// Turn 1: text + tool + final.
	for _, d := range []string{"first "} {
		nm, _ := m.Update(streamDeltaMsg{text: d})
		m = toModel(t, nm)
	}
	m = feed(t, m,
		streamToolMsg{name: "bash", input: "x"},
		streamToolDoneMsg{name: "bash", result: "y"},
		streamTextFinalMsg{text: "first"},
		streamDoneMsg{},
	)

	// Turn 2: pure text, no tools. The boundary MUST be reset.
	for _, d := range []string{"second"} {
		nm, _ := m.Update(streamDeltaMsg{text: d})
		m = toModel(t, nm)
	}
	m = feed(t, m,
		streamTextFinalMsg{text: "second"},
		streamDoneMsg{},
	)

	view := m.viewBuf
	if !strings.Contains(view, "first") {
		t.Errorf("turn 1 text lost: %q", view)
	}
	if !strings.Contains(view, "second") {
		t.Errorf("turn 2 text lost (stale boundary bug): %q", view)
	}
}
