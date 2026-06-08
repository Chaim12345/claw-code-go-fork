package runtime

import (
	"encoding/json"

	"claw-code-go/internal/api"
	"claw-code-go/internal/tools"
)

// HealToolInput attempts to parse the raw tool input JSON. If
// parsing fails or required fields are missing, it returns an
// error tool_result with a model-friendly message. Returns nil
// if the input is valid JSON with all required fields present.
//
// The returned tool_result is meant to be appended to the
// conversation as a tool_result so the model sees the error
// on its next turn and can self-correct by re-emitting the
// tool call with valid arguments.
func HealToolInput(toolName, raw string, parseErr error) *api.ContentBlock {
	if parseErr != nil {
		return &api.ContentBlock{
			Type: "tool_result",
			Content: []api.ContentBlock{{
				Type: "text",
				Text: "Error: tool " + toolName + " received malformed JSON arguments: " +
					parseErr.Error() + ". Please re-emit the tool call with valid JSON.",
			}},
			IsError: true,
		}
	}

	var inputMap map[string]any
	if err := json.Unmarshal([]byte(raw), &inputMap); err != nil {
		return &api.ContentBlock{
			Type: "tool_result",
			Content: []api.ContentBlock{{
				Type: "text",
				Text: "Error: tool " + toolName + " received malformed JSON. Please re-emit with valid JSON.",
			}},
			IsError: true,
		}
	}

	schema, ok := tools.Schema(toolName)
	if !ok {
		return nil
	}

	for _, req := range schema.Required {
		if _, ok := inputMap[req]; !ok {
			return &api.ContentBlock{
				Type: "tool_result",
				Content: []api.ContentBlock{{
					Type: "text",
					Text: "Error: tool " + toolName + " missing required field " +
						"'" + req + "'. Please re-emit with all required arguments.",
				}},
				IsError: true,
			}
		}
	}

	return nil
}
