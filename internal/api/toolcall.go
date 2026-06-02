package api

// ToolCall is a parsed tool invocation from the model output.
// The parser pipelines in ExtractToolCalls all return this canonical
// shape, regardless of whether the model emitted JSON, XML, ReAct, or
// a code-fenced payload. The conversation loop dispatches on Name and
// marshals Arguments straight into the executor's input map, so any
// new format added here just needs to produce the same struct.
type ToolCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}
