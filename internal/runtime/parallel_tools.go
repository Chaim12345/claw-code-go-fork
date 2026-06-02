package runtime

import (
	"claw-code-go/internal/api"
	"claw-code-go/internal/permissions"
	"claw-code-go/internal/tools"
	"context"
	"fmt"
	"sync"
)

// toolExecution represents a single tool execution task
type toolExecution struct {
	tb              toolBlock
	inputMap        map[string]any
	summary         string
	index           int
	needsPermission bool
	isAskUser       bool
}

// toolExecutionResult holds the result of a tool execution
type toolExecutionResult struct {
	index  int
	result api.ContentBlock
	err    error
}

// executeToolsParallel executes independent tools in parallel while maintaining order
func (loop *ConversationLoop) executeToolsParallel(
	ctx context.Context,
	toolBlocks []toolBlock,
	assistantContent []api.ContentBlock,
	events chan<- TurnEvent,
) ([]api.ContentBlock, error) {

	// Prepare tool executions
	executions := make([]toolExecution, 0, len(toolBlocks))
	for i, tb := range toolBlocks {
		var inputMap map[string]any
		for _, cb := range assistantContent {
			if cb.Type == "tool_use" && cb.ID == tb.id {
				inputMap = cb.Input
				break
			}
		}
		if inputMap == nil {
			inputMap = map[string]any{}
		}

		summary := summarizeToolInput(inputMap)

		exec := toolExecution{
			tb:              tb,
			inputMap:        inputMap,
			summary:         summary,
			index:           i,
			needsPermission: loop.PermManager != nil,
			isAskUser:       tb.name == "ask_user",
		}

		executions = append(executions, exec)
	}

	// Separate tools that need sequential execution (permissions, ask_user)
	// from those that can run in parallel
	var sequentialExecs []toolExecution
	var parallelExecs []toolExecution

	for _, exec := range executions {
		if exec.isAskUser {
			// ask_user must be sequential (needs UI interaction)
			sequentialExecs = append(sequentialExecs, exec)
		} else if exec.needsPermission && loop.PermManager != nil {
			decision := loop.PermManager.Check(exec.tb.name, exec.summary)
			if decision == permissions.DecisionAsk {
				// Permission ask needs UI interaction - sequential
				sequentialExecs = append(sequentialExecs, exec)
			} else {
				// Auto-allow or auto-deny can be parallel
				parallelExecs = append(parallelExecs, exec)
			}
		} else {
			// No permission check needed - can be parallel
			parallelExecs = append(parallelExecs, exec)
		}
	}

	// Results slice to maintain order
	results := make([]api.ContentBlock, len(executions))

	// Execute sequential tools first
	for _, exec := range sequentialExecs {
		result, err := loop.executeToolWithPermission(ctx, exec, events)
		if err != nil {
			return nil, err
		}
		results[exec.index] = result
	}

	// Execute parallel tools
	if len(parallelExecs) > 0 {
		var wg sync.WaitGroup
		resultChan := make(chan toolExecutionResult, len(parallelExecs))

		for _, exec := range parallelExecs {
			wg.Add(1)
			go func(e toolExecution) {
				defer wg.Done()

				result, err := loop.executeToolWithPermission(ctx, e, events)
				resultChan <- toolExecutionResult{
					index:  e.index,
					result: result,
					err:    err,
				}
			}(exec)
		}

		// Wait for all parallel executions to complete
		go func() {
			wg.Wait()
			close(resultChan)
		}()

		// Collect results
		for res := range resultChan {
			if res.err != nil {
				return nil, res.err
			}
			results[res.index] = res.result
		}
	}

	return results, nil
}

// executeToolWithPermission handles permission checks and tool execution
func (loop *ConversationLoop) executeToolWithPermission(
	ctx context.Context,
	exec toolExecution,
	events chan<- TurnEvent,
) (api.ContentBlock, error) {

	tb := exec.tb
	inputMap := exec.inputMap
	summary := exec.summary

	// --- Permission check (Phase 5) ---
	if loop.PermManager != nil {
		decision := loop.PermManager.Check(tb.name, summary)

		// Plan mode: describe without executing.
		if loop.PermManager.Mode == permissions.ModePlan {
			return api.ContentBlock{
				Type:      "tool_result",
				ToolUseID: tb.id,
				Content:   []api.ContentBlock{{Type: "text", Text: fmt.Sprintf("[Plan: %s %s]", tb.name, summary)}},
			}, nil
		}

		switch decision {
		case permissions.DecisionDeny:
			return api.ContentBlock{
				Type:      "tool_result",
				ToolUseID: tb.id,
				Content:   []api.ContentBlock{{Type: "text", Text: fmt.Sprintf("Permission denied for tool: %s", tb.name)}},
				IsError:   true,
			}, nil

		case permissions.DecisionAsk:
			replyCh := make(chan PermDecision, 1)
			select {
			case events <- TurnEvent{
				Type:      TurnEventPermissionAsk,
				ToolName:  tb.name,
				ToolInput: summary,
				PermReply: replyCh,
			}:
			case <-ctx.Done():
				return api.ContentBlock{}, ctx.Err()
			}

			var userDecision PermDecision
			select {
			case userDecision = <-replyCh:
			case <-ctx.Done():
				return api.ContentBlock{}, ctx.Err()
			}

			switch userDecision {
			case PermDecisionDeny:
				return api.ContentBlock{
					Type:      "tool_result",
					ToolUseID: tb.id,
					Content:   []api.ContentBlock{{Type: "text", Text: fmt.Sprintf("Permission denied for tool: %s", tb.name)}},
					IsError:   true,
				}, nil
			case PermDecisionAllowAlways:
				loop.PermManager.Remember(tb.name, summary, permissions.DecisionAllow, permissions.ScopeAlways)
			}
			// PermDecisionAllowOnce falls through to execution
		}
		// DecisionAllow falls through to execution
	}

	// ask_user: surface the question to the caller and wait for a reply.
	if tb.name == "ask_user" {
		question, _ := tools.AskUserInput(inputMap)
		if question == "" {
			question = "?"
		}
		replyCh := make(chan string, 1)
		select {
		case events <- TurnEvent{
			Type:         TurnEventAskUser,
			ToolName:     tb.name,
			ToolInput:    question,
			AskUserReply: replyCh,
		}:
		case <-ctx.Done():
			return api.ContentBlock{}, ctx.Err()
		}
		var answer string
		select {
		case answer = <-replyCh:
		case <-ctx.Done():
			return api.ContentBlock{}, ctx.Err()
		}
		return api.ContentBlock{
			Type:      "tool_result",
			ToolUseID: tb.id,
			Content:   []api.ContentBlock{{Type: "text", Text: answer}},
		}, nil
	}

	// Send tool start event
	select {
	case events <- TurnEvent{Type: TurnEventToolStart, ToolName: tb.name, ToolInput: summary}:
	case <-ctx.Done():
		return api.ContentBlock{}, ctx.Err()
	}

	// Execute the tool
	result := loop.ExecuteToolQuiet(tb.name, inputMap)
	result.ToolUseID = tb.id

	// Send tool done event
	resultText := ""
	if len(result.Content) > 0 {
		resultText = result.Content[0].Text
	}
	select {
	case events <- TurnEvent{Type: TurnEventToolDone, ToolName: tb.name, ToolResult: resultText}:
	case <-ctx.Done():
		return api.ContentBlock{}, ctx.Err()
	}

	return result, nil
}
