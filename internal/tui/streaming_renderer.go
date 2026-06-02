package tui

import (
	"bytes"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// StreamingRenderer handles real-time markdown rendering during streaming
type StreamingRenderer struct {
	mu              sync.RWMutex
	buffer          strings.Builder
	renderedCache   string
	lastRenderTime  time.Time
	renderInterval  time.Duration
	codeBlockOpen   bool
	codeBlockLang   string
	codeBlockBuffer strings.Builder
	inToolCall      bool
	toolCallBuffer  strings.Builder
	glamourRenderer *glamour.TermRenderer
}

// NewStreamingRenderer creates a new streaming markdown renderer
func NewStreamingRenderer() *StreamingRenderer {
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100),
	)

	return &StreamingRenderer{
		renderInterval:  50 * time.Millisecond, // Render every 50ms max
		glamourRenderer: renderer,
	}
}

// Append adds new text to the streaming buffer
func (sr *StreamingRenderer) Append(text string) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	sr.buffer.WriteString(text)
	
	// Check for code block markers
	sr.detectCodeBlocks(text)
	
	// Check for tool call markers
	sr.detectToolCalls(text)
}

// detectCodeBlocks tracks code block state for syntax highlighting
func (sr *StreamingRenderer) detectCodeBlocks(text string) {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Opening code block
		if strings.HasPrefix(trimmed, "```") && !sr.codeBlockOpen {
			sr.codeBlockOpen = true
			sr.codeBlockLang = strings.TrimPrefix(trimmed, "```")
			sr.codeBlockBuffer.Reset()
			continue
		}
		
		// Closing code block
		if strings.HasPrefix(trimmed, "```") && sr.codeBlockOpen {
			sr.codeBlockOpen = false
			sr.codeBlockLang = ""
			continue
		}
		
		// Inside code block
		if sr.codeBlockOpen {
			sr.codeBlockBuffer.WriteString(line)
			sr.codeBlockBuffer.WriteString("\n")
		}
	}
}

// detectToolCalls tracks tool call XML markers
func (sr *StreamingRenderer) detectToolCalls(text string) {
	// Simple detection - look for <tool_name> patterns
	if strings.Contains(text, "<") && !strings.Contains(text, "</") {
		// Potential opening tag
		if !sr.inToolCall {
			sr.inToolCall = true
			sr.toolCallBuffer.Reset()
		}
	}
	
	if sr.inToolCall {
		sr.toolCallBuffer.WriteString(text)
		
		// Check for closing tag
		if strings.Contains(text, "</") {
			sr.inToolCall = false
		}
	}
}

// Render returns the current rendered content
func (sr *StreamingRenderer) Render() string {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	// Throttle rendering
	now := time.Now()
	if now.Sub(sr.lastRenderTime) < sr.renderInterval && sr.renderedCache != "" {
		return sr.renderedCache
	}

	content := sr.buffer.String()
	
	// If we're in a tool call, hide the raw XML
	if sr.inToolCall {
		// Return content up to the tool call start
		toolStart := strings.LastIndex(content, "<")
		if toolStart > 0 {
			content = content[:toolStart]
		}
	}

	// Render markdown
	rendered := sr.renderMarkdown(content)
	
	sr.renderedCache = rendered
	sr.lastRenderTime = now
	
	return rendered
}

// renderMarkdown converts markdown to styled terminal output
func (sr *StreamingRenderer) renderMarkdown(content string) string {
	if sr.glamourRenderer == nil {
		return content
	}

	// Try to render with glamour
	rendered, err := sr.glamourRenderer.Render(content)
	if err != nil {
		// Fallback to plain text if rendering fails
		return content
	}

	return rendered
}

// GetCurrentCodeBlock returns the current code block being streamed
func (sr *StreamingRenderer) GetCurrentCodeBlock() (lang string, code string, isOpen bool) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	return sr.codeBlockLang, sr.codeBlockBuffer.String(), sr.codeBlockOpen
}

// GetCurrentToolCall returns the current tool call being streamed
func (sr *StreamingRenderer) GetCurrentToolCall() (xml string, isOpen bool) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	return sr.toolCallBuffer.String(), sr.inToolCall
}

// Reset clears the streaming buffer
func (sr *StreamingRenderer) Reset() {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	sr.buffer.Reset()
	sr.renderedCache = ""
	sr.codeBlockOpen = false
	sr.codeBlockLang = ""
	sr.codeBlockBuffer.Reset()
	sr.inToolCall = false
	sr.toolCallBuffer.Reset()
}

// GetRawContent returns the raw unrendered content
func (sr *StreamingRenderer) GetRawContent() string {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	return sr.buffer.String()
}

// ProgressiveMarkdownRenderer renders markdown incrementally as it arrives
type ProgressiveMarkdownRenderer struct {
	mu              sync.RWMutex
	chunks          []string
	renderedChunks  []string
	lastChunkIndex  int
	glamourRenderer *glamour.TermRenderer
}

// NewProgressiveMarkdownRenderer creates a new progressive renderer
func NewProgressiveMarkdownRenderer() *ProgressiveMarkdownRenderer {
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100),
	)

	return &ProgressiveMarkdownRenderer{
		chunks:          make([]string, 0),
		renderedChunks:  make([]string, 0),
		glamourRenderer: renderer,
	}
}

// AddChunk adds a new chunk of markdown
func (pmr *ProgressiveMarkdownRenderer) AddChunk(chunk string) {
	pmr.mu.Lock()
	defer pmr.mu.Unlock()

	pmr.chunks = append(pmr.chunks, chunk)
}

// RenderNew renders only new chunks since last call
func (pmr *ProgressiveMarkdownRenderer) RenderNew() string {
	pmr.mu.Lock()
	defer pmr.mu.Unlock()

	if pmr.lastChunkIndex >= len(pmr.chunks) {
		return ""
	}

	var newContent strings.Builder
	for i := pmr.lastChunkIndex; i < len(pmr.chunks); i++ {
		newContent.WriteString(pmr.chunks[i])
	}

	rendered := pmr.renderChunk(newContent.String())
	pmr.renderedChunks = append(pmr.renderedChunks, rendered)
	pmr.lastChunkIndex = len(pmr.chunks)

	return rendered
}

// RenderAll renders all chunks
func (pmr *ProgressiveMarkdownRenderer) RenderAll() string {
	pmr.mu.RLock()
	defer pmr.mu.RUnlock()

	var result strings.Builder
	for _, rendered := range pmr.renderedChunks {
		result.WriteString(rendered)
	}

	return result.String()
}

// renderChunk renders a single chunk of markdown
func (pmr *ProgressiveMarkdownRenderer) renderChunk(chunk string) string {
	if pmr.glamourRenderer == nil {
		return chunk
	}

	rendered, err := pmr.glamourRenderer.Render(chunk)
	if err != nil {
		return chunk
	}

	return rendered
}

// Reset clears all chunks
func (pmr *ProgressiveMarkdownRenderer) Reset() {
	pmr.mu.Lock()
	defer pmr.mu.Unlock()

	pmr.chunks = make([]string, 0)
	pmr.renderedChunks = make([]string, 0)
	pmr.lastChunkIndex = 0
}

// CodeBlockRenderer handles syntax highlighting for code blocks
type CodeBlockRenderer struct {
	mu    sync.RWMutex
	cache map[string]string // lang+code -> rendered
}

// NewCodeBlockRenderer creates a new code block renderer
func NewCodeBlockRenderer() *CodeBlockRenderer {
	return &CodeBlockRenderer{
		cache: make(map[string]string),
	}
}

// Render renders a code block with syntax highlighting
func (cbr *CodeBlockRenderer) Render(lang, code string) string {
	cbr.mu.Lock()
	defer cbr.mu.Unlock()

	cacheKey := lang + ":" + code
	if rendered, ok := cbr.cache[cacheKey]; ok {
		return rendered
	}

	// Use glamour for syntax highlighting
	markdown := "```" + lang + "\n" + code + "\n```"
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100),
	)

	rendered, err := renderer.Render(markdown)
	if err != nil {
		// Fallback to simple styling
		rendered = cbr.fallbackRender(lang, code)
	}

	cbr.cache[cacheKey] = rendered
	return rendered
}

// fallbackRender provides basic styling when glamour fails
func (cbr *CodeBlockRenderer) fallbackRender(lang, code string) string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2)

	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color("99")).
		Bold(true).
		Render(lang)

	var buf bytes.Buffer
	buf.WriteString(header)
	buf.WriteString("\n\n")
	buf.WriteString(code)

	return style.Render(buf.String())
}

// ClearCache clears the rendering cache
func (cbr *CodeBlockRenderer) ClearCache() {
	cbr.mu.Lock()
	defer cbr.mu.Unlock()

	cbr.cache = make(map[string]string)
}
