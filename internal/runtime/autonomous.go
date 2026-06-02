package runtime

import (
	"claw-code-go/internal/api"
	"claw-code-go/internal/permissions"
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

// RunTask is the autonomous-mode entry point. Given a high-level
// prompt, it runs the agentic loop repeatedly until the task is
// verifiably done, returning the final assistant text.
//
// It composes with existing infrastructure rather than duplicating
// it: PermissionMode is forced to "bypass" (so PermManager
// auto-allows), ask_user is stripped from the tool list (so the
// model cannot block), and an LLM-judge call decides DONE vs
// CONTINUE between SendMessage iterations. MaxTurns caps the run.
//
// The returned string is the assistant's final text from the last
// iteration, with tool-call syntax already stripped (RunTask
// reuses SendMessageStreaming's TextFinal path).
//
// Stop conditions, in order:
//  1. LLM-judge returns DONE.
//  2. MaxTurns reached (returns ErrMaxTurns with the partial result).
//  3. Context cancelled (returns ctx.Err()).
//  4. SendMessageStreaming returns an error.
func (loop *ConversationLoop) RunTask(ctx context.Context, prompt string) (string, error) {
	if !loop.Config.Autonomous {
		return "", fmt.Errorf("RunTask called with Config.Autonomous=false; set it in cfg or call SendMessage instead")
	}

	maxTurns := loop.Config.MaxTurns
	if maxTurns <= 0 {
		maxTurns = DefaultAutonomousMaxTurns
	}
	if maxTurns > MaxAutonomousTurns {
		maxTurns = MaxAutonomousTurns
	}

	// Save and restore the tool list + permission mode around the
	// run so the same loop can be used for both interactive and
	// autonomous sessions. The original loop.Config is shared with
	// the TUI, so we mutate a copy and point loop.Config at it for
	// the duration of RunTask.
	// Save the original Config pointer so the deferred restore
	// puts back the exact *Config the caller owns (not a copy on
	// the stack). The runConfig copy is only used for the duration
	// of RunTask.
	originalConfig := loop.Config
	originalTools := loop.Tools
	originalPermMgr := loop.PermManager
	runConfig := *originalConfig
	runConfig.PermissionMode = "bypass"
	runConfig.Autonomous = true
	loop.Config = &runConfig
	loop.PermManager = permissions.NewManager(permissions.ModeBypassPermissions, &permissions.Ruleset{})

	// Strip ask_user from the tool list so the model cannot block
	// on a human question. Save and restore the slice header — the
	// underlying Tool structs are shared and must not be mutated.
	loop.Tools = filterOutTool(originalTools, "ask_user")

	defer func() {
		loop.Config = originalConfig
		loop.Tools = originalTools
		loop.PermManager = originalPermMgr
	}()

	var lastText string
	for turn := 1; turn <= maxTurns; turn++ {
		if err := ctx.Err(); err != nil {
			return lastText, err
		}
		fmt.Fprintf(os.Stderr, "[autonomous] turn %d/%d\n", turn, maxTurns)

		// Send the prompt only on the first turn; on subsequent
		// turns the LLM-judge response is the user message.
		userText := prompt
		if turn > 1 {
			userText = "continue"
		}

		turnText, err := runAutonomousTurn(ctx, loop, userText)
		if err != nil {
			return lastText, fmt.Errorf("autonomous turn %d: %w", turn, err)
		}
		lastText = turnText

		// Stop condition: model produced a final text response
		// with no tool calls (end_turn). Many models naturally
		// end on a "task complete" summary when they think the
		// job is done; honor that as a fast path and skip the
		// judge call to save tokens.
		if loop.lastStopReason == "end_turn" {
			fmt.Fprintf(os.Stderr, "[autonomous] model produced end_turn; skipping judge\n")
			return lastText, nil
		}

		verdict, err := judgeTaskDone(ctx, loop, prompt, lastText)
		if err != nil {
			return lastText, fmt.Errorf("autonomous judge (turn %d): %w", turn, err)
		}
		fmt.Fprintf(os.Stderr, "[autonomous] judge: %s\n", verdict)
		if verdict == "DONE" {
			return lastText, nil
		}
	}

	return lastText, fmt.Errorf("autonomous: max turns (%d) reached without verdict DONE: %w", maxTurns, ErrMaxTurns)
}

// lastStopReason is the most recent stop_reason observed in the
// streaming turn, set by runOneTurnStreaming. Used by the fast path
// in RunTask to skip the judge when the model already ended its
// turn.
var ErrMaxTurns = fmt.Errorf("autonomous run hit max turns without DONE")

// runAutonomousTurn drives one SendMessageStreaming iteration and
// collects the final assistant text, auto-replying to permission
// asks and ask_user (the latter shouldn't fire because ask_user is
// stripped from the tool list — but handle it defensively).
func runAutonomousTurn(ctx context.Context, loop *ConversationLoop, userText string) (string, error) {
	events := make(chan TurnEvent, 64)
	errCh := make(chan error, 1)
	go func() {
		errCh <- loop.SendMessageStreaming(ctx, userText, events)
		close(events)
	}()

	var lastText string
	for ev := range events {
		switch ev.Type {
		case TurnEventTextFinal:
			lastText = ev.Text
		case TurnEventPermissionAsk:
			// Auto-allow in autonomous mode (PermManager is
			// already bypass-mode, but defend against a future
			// call site that swaps the manager).
			if ev.PermReply != nil {
				ev.PermReply <- PermDecisionAllowOnce
			}
		case TurnEventAskUser:
			// Should be unreachable (ask_user is stripped), but
			// be defensive: feed the model an answer that says
			// "keep going on your own judgement".
			if ev.AskUserReply != nil {
				ev.AskUserReply <- "No human available in autonomous mode; continue with your best judgement."
			}
		case TurnEventError:
			return lastText, ev.Err
		}
	}
	if err := <-errCh; err != nil {
		return lastText, err
	}
	return lastText, nil
}

// judgeTaskDone calls the LLM once with a fixed template and
// returns the verdict. The judge uses the same client/model as the
// main agent (no second model dependency) but with tools stripped
// and a 50-token cap so the call is cheap.
func judgeTaskDone(ctx context.Context, loop *ConversationLoop, prompt, lastText string) (string, error) {
	judgeCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	judgePrompt := fmt.Sprintf(
		`You are a verifier. You will be given an original task and the most recent assistant output. Reply with exactly one word: DONE if the task is verifiably accomplished, CONTINUE if the assistant should keep working.

Rules:
- If the assistant's most recent text says it is done and the actions taken look sufficient, reply DONE.
- If the assistant is mid-thought, encountered an error it can recover from, or its work is incomplete, reply CONTINUE.
- Reply with ONLY the single word, nothing else.

Original task:
%s

Most recent assistant output:
%s`,
		prompt, truncate(lastText, 4000),
	)

	req := api.CreateMessageRequest{
		Model:     loop.Config.Model,
		MaxTokens: 8, // one token is plenty; cap at 8 to be safe
		System:    "You are a binary verdict classifier. Reply with exactly one word: DONE or CONTINUE.",
		Messages: []api.Message{
			{Role: "user", Content: []api.ContentBlock{{Type: "text", Text: judgePrompt}}},
		},
		Stream: true,
		// No Tools field — judge must not call tools.
	}

	ch, err := loop.Client.StreamResponse(judgeCtx, req)
	if err != nil {
		return "", fmt.Errorf("judge stream: %w", err)
	}
	var text string
	for ev := range ch {
		if ev.Type == api.EventContentBlockDelta && ev.Delta.Type == "text_delta" {
			text += ev.Delta.Text
		}
	}
	text = strings.ToUpper(strings.TrimSpace(text))
	if strings.HasPrefix(text, "DONE") {
		return "DONE", nil
	}
	return "CONTINUE", nil
}

// filterOutTool returns a copy of tools with the named tool removed.
// Used to strip ask_user from the model-visible list in autonomous
// mode without mutating the original slice.
func filterOutTool(tools []api.Tool, name string) []api.Tool {
	out := make([]api.Tool, 0, len(tools))
	for _, t := range tools {
		if t.Name == name {
			continue
		}
		out = append(out, t)
	}
	return out
}

// truncate returns the first n runes of s, appending an ellipsis
// marker if truncation occurred. Used to keep the judge prompt
// bounded against huge tool-result pastes.
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "\n[...truncated...]"
}
