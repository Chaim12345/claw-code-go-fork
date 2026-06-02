package tui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// ToolCallParser extracts and parses tool calls from streaming XML
type ToolCallParser struct {
	mu              sync.RWMutex
	buffer          strings.Builder
	inToolCall      bool
	currentToolName string
	currentParams   map[string]string
	completedCalls  []ParsedToolCall
	partialXML      string
}

// ParsedToolCall represents a fully parsed tool call
type ParsedToolCall struct {
	ID         string
	Name       string
	Parameters map[string]string
	StartTime  time.Time
	EndTime    time.Time
	Result     string
	Status     ToolCallStatus
	Error      string
}

// ToolCallStatus represents the execution status
type ToolCallStatus string

const (
	ToolCallPending   ToolCallStatus = "pending"
	ToolCallRunning   ToolCallStatus = "running"
	ToolCallSuccess   ToolCallStatus = "success"
	ToolCallFailed    ToolCallStatus = "failed"
	ToolCallCancelled ToolCallStatus = "cancelled"
)

// NewToolCallParser creates a new tool call parser
func NewToolCallParser() *ToolCallParser {
	return &ToolCallParser{
		currentParams:  make(map[string]string),
		completedCalls: make([]ParsedToolCall, 0),
	}
}

// Feed adds new streaming text to the parser
func (tcp *ToolCallParser) Feed(text string) []ParsedToolCall {
	tcp.mu.Lock()
	defer tcp.mu.Unlock()

	tcp.buffer.WriteString(text)
	
	// Try to extract complete tool calls
	return tcp.extractToolCalls()
}

// extractToolCalls finds and parses complete tool call XML blocks
func (tcp *ToolCallParser) extractToolCalls() []ParsedToolCall {
	content := tcp.buffer.String()
	newCalls := make([]ParsedToolCall, 0)

	// Look for tool call patterns: <tool_name>...</tool_name>
	for {
		// Find opening tag
		start := strings.Index(content, "<")
		if start == -1 {
			break
		}

		// Find tag name
		tagEnd := strings.Index(content[start:], ">")
		if tagEnd == -1 {
			// Incomplete tag, save for next feed
			tcp.partialXML = content[start:]
			tcp.buffer.Reset()
			tcp.buffer.WriteString(tcp.partialXML)
			break
		}

		tagName := content[start+1 : start+tagEnd]
		tagName = strings.TrimSpace(tagName)
		
		// Skip if it's a closing tag or parameter tag
		if strings.HasPrefix(tagName, "/") || strings.Contains(tagName, " ") {
			content = content[start+tagEnd+1:]
			continue
		}

		// Look for closing tag
		closingTag := "</" + tagName + ">"
		closeIdx := strings.Index(content[start:], closingTag)
		if closeIdx == -1 {
			// Incomplete tool call, save for next feed
			tcp.partialXML = content[start:]
			tcp.buffer.Reset()
			tcp.buffer.WriteString(tcp.partialXML)
			break
		}

		// Extract complete tool call XML
		xmlBlock := content[start : start+closeIdx+len(closingTag)]
		
		// Parse the tool call
		if call, err := tcp.parseToolCallXML(xmlBlock); err == nil {
			newCalls = append(newCalls, call)
			tcp.completedCalls = append(tcp.completedCalls, call)
		}

		// Move past this tool call
		content = content[start+closeIdx+len(closingTag):]
	}

	return newCalls
}

// parseToolCallXML parses a complete tool call XML block
func (tcp *ToolCallParser) parseToolCallXML(xmlBlock string) (ParsedToolCall, error) {
	// Simple XML parsing - extract tool name and parameters
	lines := strings.Split(xmlBlock, "\n")
	
	var toolName string
	params := make(map[string]string)
	var currentParam string
	var paramValue strings.Builder

	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Extract tool name from opening tag
		if strings.HasPrefix(line, "<") && !strings.HasPrefix(line, "</") && toolName == "" {
			tagEnd := strings.Index(line, ">")
			if tagEnd > 0 {
				toolName = line[1:tagEnd]
			}
			continue
		}

		// Extract parameter name
		if strings.HasPrefix(line, "<") && !strings.HasPrefix(line, "</") {
			tagEnd := strings.Index(line, ">")
			if tagEnd > 0 {
				currentParam = line[1:tagEnd]
				// Check if value is on same line
				remaining := line[tagEnd+1:]
				closeTag := "</" + currentParam + ">"
				if strings.Contains(remaining, closeTag) {
					value := strings.Split(remaining, closeTag)[0]
					params[currentParam] = value
					currentParam = ""
				} else {
					paramValue.Reset()
				}
			}
			continue
		}

		// Extract parameter closing tag
		if strings.HasPrefix(line, "</") && currentParam != "" {
			params[currentParam] = paramValue.String()
			currentParam = ""
			paramValue.Reset()
			continue
		}

		// Accumulate parameter value
		if currentParam != "" {
			if paramValue.Len() > 0 {
				paramValue.WriteString("\n")
			}
			paramValue.WriteString(line)
		}
	}

	if toolName == "" {
		return ParsedToolCall{}, fmt.Errorf("no tool name found")
	}

	return ParsedToolCall{
		ID:         fmt.Sprintf("tool_%d", time.Now().UnixNano()),
		Name:       toolName,
		Parameters: params,
		StartTime:  time.Now(),
		Status:     ToolCallPending,
	}, nil
}

// GetCompletedCalls returns all completed tool calls
func (tcp *ToolCallParser) GetCompletedCalls() []ParsedToolCall {
	tcp.mu.RLock()
	defer tcp.mu.RUnlock()

	calls := make([]ParsedToolCall, len(tcp.completedCalls))
	copy(calls, tcp.completedCalls)
	return calls
}

// UpdateCallStatus updates the status of a tool call
func (tcp *ToolCallParser) UpdateCallStatus(id string, status ToolCallStatus, result string, err string) {
	tcp.mu.Lock()
	defer tcp.mu.Unlock()

	for i := range tcp.completedCalls {
		if tcp.completedCalls[i].ID == id {
			tcp.completedCalls[i].Status = status
			tcp.completedCalls[i].Result = result
			tcp.completedCalls[i].Error = err
			tcp.completedCalls[i].EndTime = time.Now()
			break
		}
	}
}

// Reset clears the parser state
func (tcp *ToolCallParser) Reset() {
	tcp.mu.Lock()
	defer tcp.mu.Unlock()

	tcp.buffer.Reset()
	tcp.inToolCall = false
	tcp.currentToolName = ""
	tcp.currentParams = make(map[string]string)
	tcp.completedCalls = make([]ParsedToolCall, 0)
	tcp.partialXML = ""
}

// RenderToolCall renders a tool call as a pretty card
func RenderToolCall(call ParsedToolCall, expanded bool) string {
	// Status icon and color
	var statusIcon string
	var statusColor lipgloss.Color
	
	switch call.Status {
	case ToolCallPending:
		statusIcon = "⏳"
		statusColor = lipgloss.Color("yellow")
	case ToolCallRunning:
		statusIcon = "⚡"
		statusColor = lipgloss.Color("blue")
	case ToolCallSuccess:
		statusIcon = "✓"
		statusColor = lipgloss.Color("green")
	case ToolCallFailed:
		statusIcon = "✗"
		statusColor = lipgloss.Color("red")
	case ToolCallCancelled:
		statusIcon = "⊘"
		statusColor = lipgloss.Color("gray")
	}

	// Tool name style
	toolNameStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("cyan")).
		Bold(true)

	// Status style
	statusStyle := lipgloss.NewStyle().
		Foreground(statusColor)

	// Header line
	header := fmt.Sprintf("%s %s %s",
		statusStyle.Render(statusIcon),
		toolNameStyle.Render(call.Name),
		statusStyle.Render(string(call.Status)),
	)

	if !expanded {
		return header
	}

	// Expanded view with parameters
	var content strings.Builder
	content.WriteString(header)
	content.WriteString("\n")

	// Parameters
	if len(call.Parameters) > 0 {
		content.WriteString("\n")
		paramStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		
		for key, value := range call.Parameters {
			// Truncate long values
			displayValue := value
			if len(displayValue) > 100 {
				displayValue = displayValue[:100] + "..."
			}
			
			content.WriteString(paramStyle.Render(fmt.Sprintf("  %s: %s\n", key, displayValue)))
		}
	}

	// Result or error
	if call.Status == ToolCallSuccess && call.Result != "" {
		content.WriteString("\n")
		resultStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("green"))
		content.WriteString(resultStyle.Render("  Result: " + call.Result))
		content.WriteString("\n")
	}

	if call.Status == ToolCallFailed && call.Error != "" {
		content.WriteString("\n")
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("red"))
		content.WriteString(errorStyle.Render("  Error: " + call.Error))
		content.WriteString("\n")
	}

	// Duration
	if !call.EndTime.IsZero() {
		duration := call.EndTime.Sub(call.StartTime)
		durationStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		content.WriteString(durationStyle.Render(fmt.Sprintf("  Duration: %s\n", duration)))
	}

	// Box style
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)

	return boxStyle.Render(content.String())
}

// ToolCallManager manages tool call display and state
type ToolCallManager struct {
	mu            sync.RWMutex
	parser        *ToolCallParser
	activeCalls   map[string]*ParsedToolCall
	expandedCalls map[string]bool
}

// NewToolCallManager creates a new tool call manager
func NewToolCallManager() *ToolCallManager {
	return &ToolCallManager{
		parser:        NewToolCallParser(),
		activeCalls:   make(map[string]*ParsedToolCall),
		expandedCalls: make(map[string]bool),
	}
}

// ProcessStream processes streaming text and extracts tool calls
func (tcm *ToolCallManager) ProcessStream(text string) string {
	tcm.mu.Lock()
	defer tcm.mu.Unlock()

	// Feed to parser
	newCalls := tcm.parser.Feed(text)

	// Track new calls
	for _, call := range newCalls {
		callCopy := call
		tcm.activeCalls[call.ID] = &callCopy
	}

	// Remove tool call XML from visible text
	cleanText := tcm.removeToolCallXML(text)
	
	return cleanText
}

// removeToolCallXML strips tool call XML from text
func (tcm *ToolCallManager) removeToolCallXML(text string) string {
	// Simple approach: remove anything between < and >
	var result strings.Builder
	inTag := false
	
	for _, ch := range text {
		if ch == '<' {
			inTag = true
			continue
		}
		if ch == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(ch)
		}
	}
	
	return result.String()
}

// RenderActiveCalls renders all active tool calls
func (tcm *ToolCallManager) RenderActiveCalls() string {
	tcm.mu.RLock()
	defer tcm.mu.RUnlock()

	if len(tcm.activeCalls) == 0 {
		return ""
	}

	var result strings.Builder
	for id, call := range tcm.activeCalls {
		expanded := tcm.expandedCalls[id]
		result.WriteString(RenderToolCall(*call, expanded))
		result.WriteString("\n")
	}

	return result.String()
}

// ToggleExpanded toggles the expanded state of a tool call
func (tcm *ToolCallManager) ToggleExpanded(id string) {
	tcm.mu.Lock()
	defer tcm.mu.Unlock()

	tcm.expandedCalls[id] = !tcm.expandedCalls[id]
}

// UpdateStatus updates a tool call's status
func (tcm *ToolCallManager) UpdateStatus(id string, status ToolCallStatus, result string, err string) {
	tcm.mu.Lock()
	defer tcm.mu.Unlock()

	if call, ok := tcm.activeCalls[id]; ok {
		call.Status = status
		call.Result = result
		call.Error = err
		call.EndTime = time.Now()
	}

	tcm.parser.UpdateCallStatus(id, status, result, err)
}

// Reset clears all tool calls
func (tcm *ToolCallManager) Reset() {
	tcm.mu.Lock()
	defer tcm.mu.Unlock()

	tcm.parser.Reset()
	tcm.activeCalls = make(map[string]*ParsedToolCall)
	tcm.expandedCalls = make(map[string]bool)
}
