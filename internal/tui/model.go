package tui

import (
	"claw-code-go/internal/auth"
	"claw-code-go/internal/config"
	"claw-code-go/internal/permissions"
	"claw-code-go/internal/runtime"
	"claw-code-go/internal/tui/debug"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	appVersion   = "0.1.0"
	textareaRows = 3 // visible rows in the multi-line input area
)

// modelEntry describes a selectable AI model in the picker overlay.
type modelEntry struct {
	id   string
	desc string
}

var anthropicModels = []modelEntry{
	{"claude-opus-4-6", "Most capable — complex reasoning and analysis"},
	{"claude-sonnet-4-6", "Balanced — great performance at speed"},
	{"claude-haiku-4-5-20251001", "Fast and lightweight — quick tasks"},
}

var openAIModels = []modelEntry{
	{"gpt-4o", "Most capable — multimodal tasks and analysis"},
	{"gpt-4o-mini", "Fast and affordable — everyday tasks"},
	{"o1-mini", "Reasoning model — math and logic"},
}

// deepseekModels lists DeepSeek models available when using the deepseek provider.
var deepseekModels = []modelEntry{
	{"expert", "Deep reasoning model (R1) — great for complex problems"},
	{"expert-thinking", "Expert + thinking — chain-of-thought reasoning"},
	{"vision", "Multimodal model — understands images"},
	{"vision-thinking", "Vision + thinking — reasoning with visual input"},
	{"instant", "Fast model — good for simple tasks"},
	{"instant-thinking", "Instant + thinking — fast with reasoning"},
	{"instant-search", "Instant + search — web search enabled"},
	{"instant-thinking-search", "Instant + thinking + search — fast, reasoning, and web"},
}

// loginProvider describes a selectable AI provider in the /login flow.
type loginProviderEntry struct {
	id   string
	name string
	desc string
}

var loginProviders = []loginProviderEntry{
	{"anthropic", "Anthropic", "Claude Sonnet, Opus, Haiku models"},
	{"openai", "OpenAI", "GPT-4o and GPT-4o-mini models"},
}

// loginMethodEntry describes an auth method choice shown for a given provider.
type loginMethodEntry struct {
	id   string
	name string
	desc string
}

var anthropicAuthMethods = []loginMethodEntry{
	{"oauth", "OAuth (browser)", "Log in with your Claude.ai account"},
	{"api_key", "API Key", "Enter your Anthropic API key manually"},
}

// appState tracks what the TUI is currently doing.
type appState int

const (
	stateInput           appState = iota // waiting for user input
	stateBusy                            // streaming response from API
	statePicker                          // model selection overlay
	stateHelp                            // help panel overlay
	statePermission                      // waiting for permission decision
	stateLoginProvider                   // /login: provider picker
	stateLoginMethod                     // /login: auth-method picker (Anthropic)
	stateLoginAPIKey                     // /login: API key text input
	stateLoginOAuth                      // /login: waiting for OAuth browser flow
	stateAskUser                         // agent has asked the user a question
	statePalette                         // command palette overlay (Ctrl+P)
	stateSessionPicker                   // session browser overlay
	stateMention                         // @-file autocomplete popup
	stateTodoPanel                       // todo list sidebar visible
	stateDebugPanel                      // debug panel overlay (Ctrl+D)
	stateSlashMenu                       // slash command autocomplete popup
	stateBackgroundTasks                 // background tasks panel (Ctrl+B)
	stateHistorySearch                   // command history search (Ctrl+R)
	stateQuickActions                    // quick actions menu (Ctrl+K)
	stateConvSearch                      // conversation search (Ctrl+F)
)

// Bubble Tea messages for async streaming events.
type (
	streamDeltaMsg     struct{ text string }
	streamTextFinalMsg struct{ text string }
	streamToolMsg      struct{ name, input string }
	streamToolDoneMsg  struct{ name, result string }
	streamUsageMsg     struct{ inputTokens, outputTokens int }
	streamDoneMsg      struct{}
	streamErrMsg       struct{ err error }
	streamWarnMsg      struct{ text string }
	streamPermAskMsg   struct {
		name, input string
		reply       chan runtime.PermDecision
	}
	streamAskUserMsg struct {
		question string
		reply    chan string
	}
)

// loginCompleteMsg is sent when a /login flow finishes (success or failure).
type loginCompleteMsg struct {
	provider string // "anthropic" or "openai"
	token    string // API key or OAuth access token
	method   string // "api_key" or "oauth"
	err      error
}

// compactResultMsg is sent when manual session compaction completes.
type compactResultMsg struct {
	summary string
	err     error
}

// Model is the Bubble Tea application model.
type Model struct {
	state  appState
	width  int
	height int
	ready  bool

	viewport viewport.Model
	textarea textarea.Model
	spinner  spinner.Model

	// history for ↑/↓ input navigation
	history *inputHistory

	// model picker state
	pickerCursor int

	// content buffers
	viewBuf   string // finalized history (all complete turns)
	streamBuf string // in-progress streaming content

	// token counts for status bar
	inputTokens  int
	outputTokens int

	// whether any streaming content has arrived (suppresses spinner)
	hasStreamContent bool

	// whether streamTextFinalMsg has already moved the text to viewBuf
	streamTextCommitted bool

	// channel from active streaming goroutine
	streamChan chan runtime.TurnEvent

	// permission ask state
	permToolName  string
	permToolInput string
	permReplyCh   chan runtime.PermDecision

	// ask_user state
	askUserQuestion string
	askUserReplyCh  chan string
	askUserInput    textinput.Model

	// /login flow state
	loginCursor   int             // cursor for provider / method pickers
	loginProvider string          // provider selected during login
	loginKeyInput textinput.Model // API key entry input (single-line, masked)

	// new opencode-style feature state
	palette       *palette
	sessionPicker *sessionPicker
	mention       *mentionAutocomplete
	slashMenu     *slashMenu
	todoPanel     *todoPanel
	toolCards     []toolCard
	permModeOrder []permissions.PermissionMode

	// streaming and rendering
	streamingRenderer *StreamingRenderer
	toolCallManager   *ToolCallManager
	codeBlockRenderer *CodeBlockRenderer

	// debug system
	debugEnabled bool
	stateMachine *debug.StateMachine

	// status badges
	badgeManager *StatusBadgeManager

	// background tasks
	backgroundTaskPanel *BackgroundTaskPanel

	// progressive disclosure
	progressiveDisclosure *ProgressiveDisclosure

	// history search
	historySearch *HistorySearch

	// quick actions menu
	quickActions *QuickActionsMenu

	// conversation search
	conversationSearch *ConversationSearch

	// natural language command parser
	naturalParser *NaturalCommandParser

	// file change history for undo/redo
	fileHistory *FileChangeHistory

	// app deps
	loop *runtime.ConversationLoop
	cfg  *runtime.Config
}

// NewModel creates a new TUI model.
func NewModel(cfg *runtime.Config, loop *runtime.ConversationLoop) Model {
	ta := textarea.New()
	ta.Placeholder = "Type a message or /help..."
	ta.CharLimit = 8192
	ta.ShowLineNumbers = false
	ta.SetHeight(textareaRows)
	// Focus the textarea so the cursor is visible from the first render.
	// The blink Cmd is returned from Init().
	ta.Focus() //nolint:errcheck

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(currentTheme.Primary)

	// Check if debug mode is enabled
	debugEnabled := os.Getenv("CLAW_DEBUG") == "1"
	if debugEnabled {
		debug.Enable()
		debug.Log(debug.EventInfo, "Debug mode enabled", nil)
	}

	return Model{
		state:                 stateInput,
		textarea:              ta,
		spinner:               s,
		history:               newInputHistory(),
		loop:                  loop,
		cfg:                   cfg,
		viewBuf:               RenderLogo(appVersion),
		palette:               newPalette(),
		sessionPicker:         newSessionPicker(),
		mention:               newMentionAutocomplete(),
		slashMenu:             newSlashMenu(),
		todoPanel:             newTodoPanel(),
		streamingRenderer:     NewStreamingRenderer(),
		toolCallManager:       NewToolCallManager(),
		codeBlockRenderer:     NewCodeBlockRenderer(),
		badgeManager:          NewStatusBadgeManager(),
		backgroundTaskPanel:   NewBackgroundTaskPanel(),
		progressiveDisclosure: NewProgressiveDisclosure(".claude"),
		historySearch:         NewHistorySearch(),
		quickActions:          NewQuickActionsMenu(),
		conversationSearch:    NewConversationSearch(),
		naturalParser:         NewNaturalCommandParser(),
		fileHistory:           NewFileChangeHistory(),
		debugEnabled:          debugEnabled,
		stateMachine:          debug.NewStateMachine(debug.StateInput, 100),
		permModeOrder: []permissions.PermissionMode{
			permissions.ModeDefault,
			permissions.ModeAcceptEdits,
			permissions.ModeBypassPermissions,
			permissions.ModePlan,
		},
	}
}

// Init is the Bubble Tea Init function.
func (m Model) Init() tea.Cmd {
	// Start cursor blink for the textarea.
	return m.textarea.Focus()
}

// --- Update -----------------------------------------------------------------

// Update is the Bubble Tea Update function.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.logResize(msg.Width, msg.Height)
		if !m.ready {
			m.ready = true
			m = m.initViewport()
		} else {
			m = m.resizeViewport()
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.MouseMsg:
		return m.handleMouse(msg)

	case streamDeltaMsg:
		if !m.hasStreamContent {
			m.hasStreamContent = true
		}

		// Log streaming event (only in debug mode to reduce overhead)
		if m.debugEnabled {
			debug.Log(debug.EventStreamChunk, "Received stream chunk", map[string]interface{}{
				"length": len(msg.text),
			})
		}

		// Strip tool-call XML from visible text so users don't see raw
		// <tool_calls>...</tool_calls> markup. We do NOT feed text to
		// the toolCallManager parser here — tool calls are tracked via
		// the runtime's TurnEventToolStart/TurnEventToolDone event
		// system (streamToolMsg/streamToolDoneMsg). The XML parser
		// creates ParsedToolCall entries with IDs that don't match the
		// toolCard IDs, leaving them permanently "pending" (⏳) and
		// causing duplicate rendering in streamDoneMsg.
		cleanText := m.toolCallManager.RemoveToolCallXML(msg.text)

		// Add to streaming renderer (buffered, will render on interval)
		m.streamingRenderer.Append(cleanText)

		// Get rendered content (throttled internally by renderer)
		renderedContent := m.streamingRenderer.Render()

		// Only update if render produced new content
		if renderedContent != "" {
			m.streamBuf = renderedContent
			m = m.refreshViewport()
		}

		return m, waitForStream(m.streamChan)

	case streamTextFinalMsg:
		// Append final cleaned text to viewBuf (preserve all content)
		if m.debugEnabled {
			debug.Log(debug.EventStreamChunk, "Received final cleaned text", map[string]interface{}{
				"length": len(msg.text),
			})
		}

		// Move streamed content to permanent viewBuf
		m.viewBuf += m.streamBuf

		// Clear streaming state (but keep the same renderer instance so
		// in-flight deltas still land in the correct buffer).
		m.streamBuf = ""
		m.streamingRenderer.Reset()
		m.streamTextCommitted = true

		m = m.refreshViewport()
		return m, waitForStream(m.streamChan)

	case streamToolMsg:
		if !m.hasStreamContent {
			m.hasStreamContent = true
		}

		// Log tool call event
		if m.debugEnabled {
			startTime := time.Now()
			debug.Log(debug.EventToolCall, fmt.Sprintf("Tool call: %s", msg.name), map[string]interface{}{
				"tool_name": msg.name,
				"input":     msg.input,
				"timestamp": startTime,
			})
		}

		// Track the tool call as a card so the user can expand it
		// later with Ctrl+T. The first event for a tool is "running";
		// we still create a card so the start line is visible.
		card := toolCard{
			id:    fmt.Sprintf("tc_%d", len(m.toolCards)),
			name:  msg.name,
			input: msg.input,
		}
		if msg.name == "edit" || msg.name == "write" || msg.name == "file_edit" || msg.name == "write_file" {
			card.hasDiff = true
		}
		m.toolCards = append(m.toolCards, card)

		// Limit tool card history to prevent memory leak (keep last 50)
		if len(m.toolCards) > 50 {
			m.toolCards = m.toolCards[len(m.toolCards)-50:]
		}

		// Update tool call manager status to running
		m.toolCallManager.UpdateStatus(card.id, ToolCallRunning, "", "")

		// Show badge for tool use
		m.badgeManager.ShowToolUse(msg.name)

		m.streamBuf += formatToolCard(card) + "\n"
		// If the agent updated the todo list, refresh the panel.
		if msg.name == "todo_write" {
			m.refreshTodosFromDisk()
		}
		m = m.refreshViewport()
		return m, waitForStream(m.streamChan)

	case streamToolDoneMsg:
		// Log tool completion
		if m.debugEnabled {
			debug.Log(debug.EventToolComplete, fmt.Sprintf("Tool completed: %s", msg.name), map[string]interface{}{
				"tool_name": msg.name,
				"result":    msg.result,
			})
		}

		// Update the most recent matching card with the result. If the
		// tool wrote/edited a file, attempt to compute an inline diff
		// for the card so the user can see the change visually.
		for i := len(m.toolCards) - 1; i >= 0; i-- {
			if m.toolCards[i].name == msg.name && m.toolCards[i].result == "" {
				m.toolCards[i].result = msg.result
				if m.toolCards[i].hasDiff {
					m.toolCards[i].diffInline = m.computeToolDiff(m.toolCards[i])
				}

				// Record file change for undo/redo
				m.recordFileChange(m.toolCards[i])

				// Update tool call manager status to success
				m.toolCallManager.UpdateStatus(m.toolCards[i].id, ToolCallSuccess, msg.result, "")

				// Show success badge
				m.badgeManager.ShowCompleted(msg.name)
				break
			}
		}
		m.streamBuf = m.renderToolCards() + "\n"
		m = m.refreshViewport()
		return m, waitForStream(m.streamChan)

	case streamUsageMsg:
		m.inputTokens = msg.inputTokens
		m.outputTokens = msg.outputTokens
		return m, waitForStream(m.streamChan)

	case streamDoneMsg:
		// Log stream completion
		if m.debugEnabled {
			debug.Log(debug.EventStreamComplete, "Stream completed", map[string]interface{}{
				"input_tokens":  m.inputTokens,
				"output_tokens": m.outputTokens,
			})
		}

		// If streamTextFinalMsg already committed the text, only
		// append the token line. Tool cards are already in streamBuf
		// via streamToolMsg/streamToolDoneMsg calls.
		if m.streamTextCommitted {
			tokLine := statusStyle.Render(fmt.Sprintf(
				"\nTokens: %s in / %s out\n\n",
				formatNum(m.inputTokens),
				formatNum(m.outputTokens),
			))
			if m.streamBuf != "" {
				m.viewBuf += m.streamBuf + tokLine
				m.streamBuf = ""
			} else {
				m.viewBuf += tokLine
			}
		} else {
			// No streamTextFinalMsg arrived — commit whatever we have.
			if m.streamBuf != "" || m.hasStreamContent {
				tokLine := statusStyle.Render(fmt.Sprintf(
					"\n\nTokens: %s in / %s out\n\n",
					formatNum(m.inputTokens),
					formatNum(m.outputTokens),
				))
				m.viewBuf += m.streamBuf + tokLine
				m.streamBuf = ""
			}
		}
		m.hasStreamContent = false
		m.streamTextCommitted = false
		m.toolCards = nil // Reset tool cards for next turn

		// Reset streaming components
		m.streamingRenderer.Reset()
		m.toolCallManager.Reset()

		// Clear progress badges
		m.badgeManager.HideProgress()

		m.transitionState(stateInput, "stream completed")
		m = m.refreshViewport()
		m.viewport.GotoBottom()
		return m, nil

	case streamWarnMsg:
		m.viewBuf += warnStyle.Render(fmt.Sprintf("Warning: %s\n\n", msg.text))
		m = m.refreshViewport()
		return m, waitForStream(m.streamChan)

	case streamPermAskMsg:
		m.transitionState(statePermission, "permission requested by agent")
		m.permToolName = msg.name
		m.permToolInput = msg.input
		m.permReplyCh = msg.reply
		m = m.refreshViewport()
		return m, nil

	case streamAskUserMsg:
		ti := textinput.New()
		ti.Placeholder = "Type your answer and press Enter..."
		ti.CharLimit = 2048
		ti.Focus()
		m.askUserInput = ti
		m.askUserQuestion = msg.question
		m.askUserReplyCh = msg.reply
		m.transitionState(stateAskUser, "agent asked user a question")
		m = m.refreshViewport()
		return m, nil

	case streamErrMsg:
		m.logError("stream", msg.err)
		m.viewBuf += errorStyle.Render(fmt.Sprintf("Error: %v\n\n", msg.err))
		m.streamBuf = ""
		m.hasStreamContent = false
		m.transitionState(stateInput, "stream error")
		m = m.refreshViewport()
		return m, nil

	case loginCompleteMsg:
		return m.handleLoginComplete(msg)

	case compactResultMsg:
		if msg.err != nil {
			m.viewBuf += errorStyle.Render(fmt.Sprintf("Compaction failed: %v\n\n", msg.err))
		} else {
			m.viewBuf += statusStyle.Render(fmt.Sprintf("Session compacted successfully.\n\nSummary:\n%s\n\n", msg.summary))
			// Update token counts after compaction
			m.inputTokens = runtime.EstimateTokens(m.loop.Session.Messages)
		}
		m = m.refreshViewport()
		return m, nil

	case spinner.TickMsg:
		if (m.state == stateBusy && !m.hasStreamContent) || m.state == stateLoginOAuth {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	return m, nil
}

// handleKey dispatches key events based on current state.
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case statePalette:
		return m.handlePaletteKey(msg)
	case stateSessionPicker:
		return m.handleSessionPickerKey(msg)
	case stateMention:
		return m.handleMentionKey(msg)
	case stateSlashMenu:
		return m.handleSlashMenuKey(msg)
	case statePicker:
		return m.handlePickerKey(msg)
	case stateHelp:
		return m.handleHelpKey(msg)
	case statePermission:
		return m.handlePermissionKey(msg)
	case stateAskUser:
		return m.handleAskUserKey(msg)
	case stateLoginProvider:
		return m.handleLoginProviderKey(msg)
	case stateLoginMethod:
		return m.handleLoginMethodKey(msg)
	case stateLoginAPIKey:
		return m.handleLoginAPIKeyKey(msg)
	case stateLoginOAuth:
		return m.handleLoginOAuthKey(msg)
	case stateDebugPanel:
		return m.handleDebugPanelKey(msg)
	case stateBackgroundTasks:
		return m.handleBackgroundTasksKey(msg)
	case stateHistorySearch:
		return m.handleHistorySearchKey(msg)
	case stateQuickActions:
		return m.handleQuickActionsKey(msg)
	case stateConvSearch:
		return m.handleConvSearchKey(msg)
	case stateBusy:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		return m, nil
	}

	// stateInput
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit

	case tea.KeyCtrlP:
		// Open the command palette (opencode's signature key binding).
		m.palette.open(m.cfg.ProviderName)
		m.transitionState(statePalette, "user opened palette")
		return m, nil

	case tea.KeyCtrlD:
		// Toggle debug panel
		if m.debugEnabled {
			debug.TogglePanel()
			if debug.IsPanelVisible() {
				m.transitionState(stateDebugPanel, "user opened debug panel")
			} else {
				m.transitionState(stateInput, "user closed debug panel")
			}
		}
		return m, nil

	case tea.KeyCtrlB:
		// Toggle background tasks panel
		m.backgroundTaskPanel.Toggle()
		if m.backgroundTaskPanel.IsVisible() {
			m.transitionState(stateBackgroundTasks, "user opened background tasks")
		} else {
			m.transitionState(stateInput, "user closed background tasks")
		}
		return m, nil

	case tea.KeyCtrlR:
		// Open command history search
		historyItems := m.history.GetAll()
		m.historySearch.Open(historyItems)
		m.transitionState(stateHistorySearch, "user opened history search")
		return m, nil

	case tea.KeyCtrlK:
		// Open quick actions menu
		m.quickActions.Open()
		m.transitionState(stateQuickActions, "user opened quick actions")
		return m, nil

	case tea.KeyCtrlF:
		// Open conversation search
		m.conversationSearch.Open(m.loop.Session.Messages)
		m.transitionState(stateConvSearch, "user opened conversation search")
		return m, nil

	case tea.KeyCtrlZ:
		// Undo last file change
		if change, err := m.fileHistory.Undo(); err == nil {
			m.viewBuf += statusStyle.Render(fmt.Sprintf("⏪ Undid %s: %s\n\n", change.Operation, change.FilePath))
			m.badgeManager.ShowSuccess(fmt.Sprintf("Undid %s", change.Operation))
		} else {
			m.viewBuf += statusStyle.Render(fmt.Sprintf("Nothing to undo\n\n"))
		}
		m = m.refreshViewport()
		return m, nil

	case tea.KeyCtrlY:
		// Redo last undone change
		if change, err := m.fileHistory.Redo(); err == nil {
			m.viewBuf += statusStyle.Render(fmt.Sprintf("⏩ Redid %s: %s\n\n", change.Operation, change.FilePath))
			m.badgeManager.ShowSuccess(fmt.Sprintf("Redid %s", change.Operation))
		} else {
			m.viewBuf += statusStyle.Render(fmt.Sprintf("Nothing to redo\n\n"))
		}
		m = m.refreshViewport()
		return m, nil

	case tea.KeyShiftTab:
		// Cycle the permission mode.
		return m.cyclePermissionMode()

	case tea.KeyCtrlT:
		// Toggle the most recent tool card (opencode behaviour).
		return m.toggleLastToolCard()

	case tea.KeyEnter:
		// Submit the message.
		return m.handleSubmit()

	case tea.KeyTab:
		// If the @-mention popup is active, Tab inserts the highlighted file.
		if m.mention.active && len(m.mention.matches) > 0 {
			return m.applyMentionSelection()
		}
		// Otherwise, fall through to textarea.
		m.history.Reset()
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd

	case tea.KeyCtrlJ:
		// Ctrl+J inserts a real newline into the multi-line input.
		m.history.Reset()
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(tea.KeyMsg{Type: tea.KeyEnter})
		return m, cmd

	case tea.KeyUp:
		// Navigate to previous history entry when input is single-line.
		if !strings.Contains(m.textarea.Value(), "\n") {
			prev := m.history.Prev(m.textarea.Value())
			m.textarea.SetValue(prev)
			return m, nil
		}
		// Multi-line: let textarea handle cursor movement.
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd

	case tea.KeyDown:
		// Navigate to next history entry when input is single-line.
		if !strings.Contains(m.textarea.Value(), "\n") {
			next := m.history.Next()
			m.textarea.SetValue(next)
			return m, nil
		}
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd

	case tea.KeyPgUp, tea.KeyPgDown:
		// Scroll the conversation viewport.
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd

	default:
		// All other keys go to the textarea. Reset history navigation on
		// any edit so the draft is not accidentally discarded.
		m.history.Reset()
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		// After every keystroke, refresh the @-mention popup.
		if m.mention.update(m.textarea.Value()) {
			m.transitionState(stateMention, "mention triggered")
			return m, cmd
		}
		// Also check for slash command popup.
		if m.slashMenu.update(m.textarea.Value()) {
			m.transitionState(stateSlashMenu, "slash menu triggered")
			return m, cmd
		}
		if m.state == stateMention || m.state == stateSlashMenu {
			m.transitionState(stateInput, "popup deactivated")
		}
		return m, cmd
	}
}

// handleSubmit processes the current textarea value.
func (m Model) handleSubmit() (tea.Model, tea.Cmd) {
	text := strings.TrimSpace(m.textarea.Value())
	if text == "" {
		return m, nil
	}
	m.textarea.Reset()
	m.history.Push(text)
	m.history.Reset()

	// Try parsing as natural language command first
	cmd, args, isNatural := m.naturalParser.Parse(text)
	if cmd != "" {
		// It's a command (natural or slash)
		if isNatural {
			// Show feedback that natural language was recognized
			m.viewBuf += statusStyle.Render(fmt.Sprintf("💬 Understood: %s → %s\n", text, cmd))
			m = m.refreshViewport()
		}

		// Execute the command with any parsed arguments
		if len(args) > 0 {
			// Reconstruct command with args
			fullCmd := cmd + " " + strings.Join(args, " ")
			return m.handleSlashCommand(fullCmd)
		}
		return m.handleSlashCommand(cmd)
	}

	// Not a command - treat as regular message
	return m.startMessage(text)
}

// handleSlashCommand processes built-in slash commands.
func (m Model) handleSlashCommand(cmd string) (tea.Model, tea.Cmd) {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return m, nil
	}

	switch parts[0] {
	case "/model":
		m.transitionState(statePicker, "user opened model picker")
		m.pickerCursor = 0
		for i, km := range m.activeModels() {
			if km.id == m.cfg.Model {
				m.pickerCursor = i
				break
			}
		}
		return m, nil

	case "/help":
		m.transitionState(stateHelp, "user opened help")
		return m, nil

	case "/login":
		m.transitionState(stateLoginProvider, "user started /login")
		m.loginCursor = 0
		return m, nil

	case "/clear":
		m.loop.ClearSession()
		m.viewBuf = statusStyle.Render("Session cleared.\n\n")
		m.streamBuf = ""
		m.inputTokens = 0
		m.outputTokens = 0
		m = m.refreshViewport()
		return m, nil

	case "/compact":
		return m.handleCompactCommand()

	case "/session-list":
		metas, err := m.loop.ListSessionsWithMeta()
		if err != nil {
			m.viewBuf += errorStyle.Render(fmt.Sprintf("Error listing sessions: %v\n\n", err))
		} else if len(metas) == 0 {
			m.viewBuf += statusStyle.Render("No saved sessions.\n\n")
		} else {
			m.viewBuf += statusStyle.Render(formatSessionList(metas) + "\n\n")
		}
		m = m.refreshViewport()
		return m, nil

	case "/theme":
		theme := "dark"
		if len(parts) > 1 {
			theme = parts[1]
		}
		switch theme {
		case "light":
			SetTheme(LightTheme)
			m.viewBuf += statusStyle.Render("Theme: light.\n\n")
		default:
			SetTheme(DarkTheme)
			m.viewBuf += statusStyle.Render("Theme: dark.\n\n")
		}
		m = m.refreshViewport()
		return m, nil

	case "/auth":
		sub := "status"
		if len(parts) > 1 {
			sub = parts[1]
		}
		msg := m.handleAuthSubcommand(sub)
		m.viewBuf += statusStyle.Render(msg + "\n\n")
		m = m.refreshViewport()
		return m, nil

	case "/session":
		return m.handleSessionCommand(parts)

	case "/todo":
		m.todoPanel.toggle()
		if m.todoPanel.visible {
			m.refreshTodosFromDisk()
			m.viewBuf += statusStyle.Render("Todo panel: on\n\n")
		} else {
			m.viewBuf += statusStyle.Render("Todo panel: off\n\n")
		}
		m = m.refreshViewport()
		return m, nil

	case "/sessions":
		return m.openSessionPicker()

	case "/status":
		return m.handleStatus()

	case "/init":
		return m.handleInit()

	case "/cost":
		return m.handleCost()

	case "/config":
		return m.handleConfig(parts)

	case "/exit", "/quit":
		return m, tea.Quit

	default:
		m.viewBuf += errorStyle.Render(fmt.Sprintf("Unknown command: %s  (type /help for commands)\n\n", parts[0]))
		m = m.refreshViewport()
		return m, nil
	}
}

// handleSessionCommand handles /session list|save|load <name>.
func (m Model) handleSessionCommand(parts []string) (tea.Model, tea.Cmd) {
	sub := "list"
	if len(parts) > 1 {
		sub = parts[1]
	}
	switch sub {
	case "list":
		metas, err := m.loop.ListSessionsWithMeta()
		if err != nil {
			m.viewBuf += errorStyle.Render(fmt.Sprintf("Error listing sessions: %v\n\n", err))
		} else if len(metas) == 0 {
			m.viewBuf += statusStyle.Render("No saved sessions.\n\n")
		} else {
			m.viewBuf += statusStyle.Render(formatSessionList(metas) + "\n\n")
		}
	case "save":
		name := ""
		if len(parts) > 2 {
			name = parts[2]
		}
		if name != "" {
			m.loop.Session.ID = name
		}
		if err := m.loop.SaveCurrentSession(); err != nil {
			m.viewBuf += errorStyle.Render(fmt.Sprintf("Error saving session: %v\n\n", err))
		} else {
			m.viewBuf += statusStyle.Render(fmt.Sprintf("Session saved: %s\n\n", m.loop.Session.ID))
		}
	case "load":
		if len(parts) < 3 {
			return m.openSessionPicker()
		}
		id := parts[2]
		if err := m.loop.LoadNamedSession(id); err != nil {
			m.viewBuf += errorStyle.Render(fmt.Sprintf("Error loading session %q: %v\n\n", id, err))
		} else {
			m.viewBuf += statusStyle.Render(fmt.Sprintf("Session loaded: %s (%d messages)\n\n", id, m.loop.MessageCount()))
		}
	default:
		m.viewBuf += errorStyle.Render(fmt.Sprintf("Unknown /session subcommand %q. Usage: /session list|save|load <name>\n\n", sub))
	}
	m = m.refreshViewport()
	return m, nil
}

// handleStatus shows current model, provider, permission mode, and session info.
func (m Model) handleStatus() (tea.Model, tea.Cmd) {
	permMode := "default"
	if m.cfg.PermissionMode != "" {
		permMode = m.cfg.PermissionMode
	}
	if m.loop.PermManager != nil {
		permMode = m.loop.PermManager.Mode.String()
	}
	lines := []string{
		fmt.Sprintf("Provider       : %s", m.cfg.ProviderName),
		fmt.Sprintf("Model          : %s", m.cfg.Model),
		fmt.Sprintf("Permission mode: %s", permMode),
		fmt.Sprintf("Session ID     : %s", m.loop.Session.ID),
		fmt.Sprintf("Messages       : %d", m.loop.MessageCount()),
		fmt.Sprintf("Tokens in/out  : %s / %s", formatNum(m.inputTokens), formatNum(m.outputTokens)),
	}
	m.viewBuf += statusStyle.Render(strings.Join(lines, "\n") + "\n\n")
	m = m.refreshViewport()
	return m, nil
}

// handleInit creates .claude/settings.json with defaults.
func (m Model) handleInit() (tea.Model, tea.Cmd) {
	err := config.InitProject(m.cfg.Model)
	switch {
	case err == nil:
		m.viewBuf += statusStyle.Render("Created .claude/settings.json with defaults.\n\n")
	case os.IsExist(err):
		m.viewBuf += statusStyle.Render(".claude/settings.json already exists — no changes made.\n\n")
	default:
		m.viewBuf += errorStyle.Render(fmt.Sprintf("init: %v\n\n", err))
	}
	m = m.refreshViewport()
	return m, nil
}

// handleCost shows token usage and best-effort cost estimate for the session.
func (m Model) handleCost() (tea.Model, tea.Cmd) {
	var report string
	if m.loop.Usage != nil && m.loop.Usage.Turns > 0 {
		report = m.loop.Usage.FormatSummary()
		if m.loop.Compaction.CompactionCount > 0 {
			report += fmt.Sprintf("Compactions    : %d\n", m.loop.Compaction.CompactionCount)
		}
	} else {
		// No turns recorded yet; fall back to compaction-state totals.
		c := m.loop.Compaction
		lines := []string{
			fmt.Sprintf("Input tokens   : %s", formatNum(c.TotalInputTokens)),
			fmt.Sprintf("Output tokens  : %s", formatNum(c.TotalOutputTokens)),
			fmt.Sprintf("Compactions    : %d", c.CompactionCount),
			"Cost           : unavailable (no completed turns yet)",
		}
		report = strings.Join(lines, "\n")
	}
	m.viewBuf += statusStyle.Render(report + "\n\n")
	m = m.refreshViewport()
	return m, nil
}

// handleConfig handles /config [key [value]].
func (m Model) handleConfig(parts []string) (tea.Model, tea.Cmd) {
	if len(parts) == 1 {
		// Show all config.
		permMode := m.cfg.PermissionMode
		if permMode == "" {
			permMode = "default"
		}
		lines := []string{
			fmt.Sprintf("model          = %s", m.cfg.Model),
			fmt.Sprintf("permissionMode = %s", permMode),
			fmt.Sprintf("maxTokens      = %d", m.cfg.MaxTokens),
			fmt.Sprintf("theme          = %s", m.cfg.Theme),
		}
		m.viewBuf += statusStyle.Render(strings.Join(lines, "\n") + "\n\n")
		m = m.refreshViewport()
		return m, nil
	}

	key := parts[1]
	if len(parts) == 2 {
		// Show single key.
		val := m.configGet(key)
		if val == "" {
			m.viewBuf += errorStyle.Render(fmt.Sprintf("Unknown config key: %s\n\n", key))
		} else {
			m.viewBuf += statusStyle.Render(fmt.Sprintf("%s = %s\n\n", key, val))
		}
		m = m.refreshViewport()
		return m, nil
	}

	// Set value.
	value := strings.Join(parts[2:], " ")
	if err := m.configSet(key, value); err != nil {
		m.viewBuf += errorStyle.Render(fmt.Sprintf("config set: %v\n\n", err))
	} else {
		m.viewBuf += statusStyle.Render(fmt.Sprintf("Set %s = %s\n\n", key, value))
	}
	m = m.refreshViewport()
	return m, nil
}

// configGet returns the string value of a config key from the active Config.
func (m Model) configGet(key string) string {
	switch key {
	case "model":
		return m.cfg.Model
	case "permissionMode":
		if m.cfg.PermissionMode == "" {
			return "default"
		}
		return m.cfg.PermissionMode
	case "maxTokens":
		return fmt.Sprintf("%d", m.cfg.MaxTokens)
	case "theme":
		return m.cfg.Theme
	default:
		return ""
	}
}

// configSet updates a config key in-memory and persists to project settings.json.
func (m *Model) configSet(key, value string) error {
	s := &config.Settings{}
	switch key {
	case "model":
		m.cfg.Model = value
		m.loop.Config.Model = value
		s.Model = value
	case "permissionMode":
		m.cfg.PermissionMode = value
		s.PermissionMode = value
	case "theme":
		m.cfg.Theme = value
		s.Theme = value
		switch value {
		case "light":
			SetTheme(LightTheme)
		default:
			SetTheme(DarkTheme)
		}
	default:
		return fmt.Errorf("unknown config key %q (valid: model, permissionMode, theme)", key)
	}
	return config.WriteProject(s)
}

// handleAuthSubcommand executes a legacy /auth subcommand and returns output text.
func (m Model) handleAuthSubcommand(sub string) string {
	switch sub {
	case "login":
		td, err := auth.StartOAuthFlow()
		if err != nil {
			return fmt.Sprintf("Auth login error: %v", err)
		}
		if err := auth.SaveTokens(td); err != nil {
			return fmt.Sprintf("Login succeeded but could not save token: %v", err)
		}
		return "Login successful. Token saved to ~/.claw-code/auth.json"

	case "logout":
		if err := auth.ClearTokens(); err != nil {
			return fmt.Sprintf("Logout error: %v", err)
		}
		return "Logged out. Stored tokens cleared."

	case "status":
		s := auth.GetStatus()
		lines := []string{
			fmt.Sprintf("Provider       : %s", m.cfg.ProviderName),
			fmt.Sprintf("Authenticated  : %v", s.Authenticated),
			fmt.Sprintf("Method         : %s", s.Method),
		}
		if s.Method == "oauth" && !s.ExpiresAt.IsZero() {
			lines = append(lines,
				fmt.Sprintf("Token expires  : %s", s.ExpiresAt.Format("2006-01-02 15:04:05 MST")),
				fmt.Sprintf("Has refresh    : %v", s.HasRefresh),
			)
		}
		return strings.Join(lines, "\n")

	default:
		return fmt.Sprintf("Unknown auth subcommand %q. Usage: /auth login | logout | status", sub)
	}
}

// --- /login flow ------------------------------------------------------------

// handleLoginProviderKey handles key input on the provider picker screen.
func (m Model) handleLoginProviderKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEsc:
		m.state = stateInput
		return m, nil
	case tea.KeyUp:
		if m.loginCursor > 0 {
			m.loginCursor--
		}
		return m, nil
	case tea.KeyDown:
		if m.loginCursor < len(loginProviders)-1 {
			m.loginCursor++
		}
		return m, nil
	case tea.KeyEnter:
		chosen := loginProviders[m.loginCursor]
		m.loginProvider = chosen.id
		m.loginCursor = 0

		switch chosen.id {
		case "anthropic":
			m.transitionState(stateLoginMethod, "user chose anthropic oauth/key")
		default:
			m = m.startAPIKeyInput()
		}
		return m, nil
	}
	return m, nil
}

// handleLoginMethodKey handles key input on the auth-method picker (Anthropic).
func (m Model) handleLoginMethodKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEsc:
		m.state = stateLoginProvider
		m.loginCursor = 0
		return m, nil
	case tea.KeyUp:
		if m.loginCursor > 0 {
			m.loginCursor--
		}
		return m, nil
	case tea.KeyDown:
		if m.loginCursor < len(anthropicAuthMethods)-1 {
			m.loginCursor++
		}
		return m, nil
	case tea.KeyEnter:
		chosen := anthropicAuthMethods[m.loginCursor]
		m.loginCursor = 0
		switch chosen.id {
		case "oauth":
			return m.startOAuthLogin()
		default:
			m = m.startAPIKeyInput()
			return m, nil
		}
	}
	return m, nil
}

// startAPIKeyInput transitions to the API key entry state.
func (m Model) startAPIKeyInput() Model {
	ti := textinput.New()
	ti.Placeholder = "Paste API key and press Enter..."
	ti.EchoMode = textinput.EchoPassword
	ti.CharLimit = 512
	ti.Focus()
	m.loginKeyInput = ti
	m.transitionState(stateLoginAPIKey, "entering api key")
	return m
}

// handleLoginAPIKeyKey handles key input when the user is typing an API key.
func (m Model) handleLoginAPIKeyKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEsc:
		m.state = stateInput
		return m, nil
	case tea.KeyEnter:
		apiKey := strings.TrimSpace(m.loginKeyInput.Value())
		if apiKey == "" {
			m.viewBuf += errorStyle.Render("API key cannot be empty.\n\n")
			m.transitionState(stateInput, "api key empty - rejected")
			m = m.refreshViewport()
			return m, nil
		}
		saveErr := auth.SetProviderAPIKey(m.loginProvider, apiKey)
		if saveErr != nil {
			return m.handleLoginComplete(loginCompleteMsg{err: saveErr})
		}
		return m.handleLoginComplete(loginCompleteMsg{
			provider: m.loginProvider,
			token:    apiKey,
			method:   "api_key",
		})
	}

	var cmd tea.Cmd
	m.loginKeyInput, cmd = m.loginKeyInput.Update(msg)
	return m, cmd
}

// startOAuthLogin prepares the OAuth session, shows the URL, and waits in background.
func (m Model) startOAuthLogin() (Model, tea.Cmd) {
	session, err := auth.PrepareOAuthFlow()
	if err != nil {
		m.viewBuf += errorStyle.Render(fmt.Sprintf("OAuth setup failed: %v\n\n", err))
		m.transitionState(stateInput, "oauth setup failed")
		m = m.refreshViewport()
		return m, nil
	}

	m.transitionState(stateLoginOAuth, "oauth flow started")
	m.viewBuf += statusStyle.Render(fmt.Sprintf(
		"Opening browser for Anthropic OAuth login...\n"+
			"If your browser doesn't open, visit:\n  %s\n\n"+
			"Waiting for callback… (5-minute timeout)\n\n",
		session.AuthURL,
	))
	m = m.refreshViewport()

	return m, tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			td, err := session.Complete()
			if err != nil {
				return loginCompleteMsg{err: err}
			}
			if err := auth.SetProviderOAuth("anthropic", td); err != nil {
				return loginCompleteMsg{err: fmt.Errorf("save token: %w", err)}
			}
			return loginCompleteMsg{
				provider: "anthropic",
				token:    td.AccessToken,
				method:   "oauth",
			}
		},
	)
}

// handleLoginOAuthKey lets the user Ctrl+C to abort the OAuth wait.
func (m Model) handleLoginOAuthKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyCtrlC {
		return m, tea.Quit
	}
	return m, nil
}

// handleLoginComplete is called when a /login flow finishes (success or error).
func (m Model) handleLoginComplete(result loginCompleteMsg) (tea.Model, tea.Cmd) {
	m.state = stateInput

	if result.err != nil {
		m.viewBuf += errorStyle.Render(fmt.Sprintf("Login failed: %v\n\n", result.err))
		m = m.refreshViewport()
		return m, nil
	}

	m.cfg.ProviderName = result.provider
	m.cfg.AuthMethod = result.method
	if result.method == "oauth" {
		m.cfg.OAuthToken = result.token
		m.cfg.APIKey = ""
	} else {
		m.cfg.APIKey = result.token
		m.cfg.OAuthToken = ""
	}

	switch result.provider {
	case "openai":
		m.cfg.Model = "gpt-4o"
	default:
		m.cfg.Model = runtime.DefaultModel
	}
	m.loop.Config.Model = m.cfg.Model

	client, err := runtime.NewProviderClient(m.cfg)
	if err != nil {
		m.viewBuf += errorStyle.Render(fmt.Sprintf(
			"Login succeeded but could not create provider client: %v\n\n", err))
		m = m.refreshViewport()
		return m, nil
	}
	m.loop.Client = client

	m.viewBuf += statusStyle.Render(fmt.Sprintf(
		"Logged in to %s via %s. Model set to %s. Ready!\n\n",
		result.provider, result.method, m.cfg.Model,
	))
	m = m.refreshViewport()
	return m, nil
}

// handleAskUserKey handles key input when the agent is waiting for a user answer.
func (m Model) handleAskUserKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEsc:
		// Cancel — send empty string so the agent loop can proceed.
		ch := m.askUserReplyCh
		m.askUserReplyCh = nil
		m.askUserQuestion = ""
		m.transitionState(stateBusy, "ask_user answered")
		m = m.refreshViewport()
		return m, tea.Batch(
			func() tea.Msg { ch <- ""; return nil },
			waitForStream(m.streamChan),
		)
	case tea.KeyEnter:
		answer := strings.TrimSpace(m.askUserInput.Value())
		ch := m.askUserReplyCh
		m.askUserReplyCh = nil
		m.askUserQuestion = ""
		m.transitionState(stateBusy, "ask_user answered")
		m = m.refreshViewport()
		return m, tea.Batch(
			func() tea.Msg { ch <- answer; return nil },
			waitForStream(m.streamChan),
		)
	}
	var cmd tea.Cmd
	m.askUserInput, cmd = m.askUserInput.Update(msg)
	return m, cmd
}

// viewAskUser renders the ask-user question overlay.
func (m Model) viewAskUser() string {
	q := m.askUserQuestion
	if len(q) > 200 {
		q = q[:200] + "..."
	}
	content := lipgloss.JoinVertical(lipgloss.Left,
		headerStyle.Render("Agent Question"),
		"",
		"  "+q,
		"",
		"  "+m.askUserInput.View(),
		"",
		statusStyle.Render("  Enter to answer  •  Esc to skip  •  Ctrl+C to quit"),
	)
	box := helpBoxStyle.Width(min(72, m.width-4)).Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

// startMessage begins a streaming conversation turn.
func (m Model) startMessage(text string) (tea.Model, tea.Cmd) {
	m.viewBuf += userLabelStyle.Render("You") + ": " + text + "\n\n"
	m.viewBuf += assistantLabelStyle.Render("Claude") + ": "
	m.transitionState(stateBusy, "message submitted")
	m.hasStreamContent = false

	ch := make(chan runtime.TurnEvent, 64)
	m.streamChan = ch

	loop := m.loop
	go func() {
		err := loop.SendMessageStreaming(context.Background(), text, ch)
		if err != nil {
			// Always surface errors to the TUI. The conversation loop
			// already emits TurnEventError before returning on most paths,
			// but if it returns a non-nil error without emitting (legacy
			// path or cancelled-context case), send one explicitly so
			// the user sees the failure instead of an empty response.
			select {
			case ch <- runtime.TurnEvent{Type: runtime.TurnEventError, Err: err}:
			default:
				// Channel buffer is full; error is already being processed
			}
		}
		close(ch)
	}()

	m = m.refreshViewport()
	return m, tea.Batch(
		m.spinner.Tick,
		waitForStream(ch),
	)
}

// handlePaletteKey handles keys when the command palette is open.
func (m Model) handlePaletteKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	chosen, shouldClose := m.palette.updateKey(msg)
	if !shouldClose {
		return m, nil
	}
	if chosen == nil {
		m.palette.close()
		m.transitionState(stateInput, "palette closed (esc)")
		return m, nil
	}
	cmd := *chosen
	m.palette.close()
	m.state = stateInput

	// Route the chosen command. System actions are handled directly;
	// others are funnelled through the slash-command pipeline so the
	// existing TUI code keeps a single source of truth.
	if cmd.system != "" {
		return m.runSystemAction(cmd.system)
	}
	if cmd.action != "" {
		return m.handleSlashCommand("/" + cmd.action)
	}
	return m, nil
}

// runSystemAction executes a palette-only command that has no slash
// equivalent. Kept tiny so the palette file stays UI-only.
func (m Model) runSystemAction(system string) (tea.Model, tea.Cmd) {
	switch system {
	case "theme-toggle":
		if currentTheme.Name == "dark" {
			SetTheme(LightTheme)
			m.viewBuf += statusStyle.Render("Theme: light.\n\n")
		} else {
			SetTheme(DarkTheme)
			m.viewBuf += statusStyle.Render("Theme: dark.\n\n")
		}
		// Rebuild the markdown renderer to match the new theme.
		_ = rebuildMarkdownRenderer()
		m = m.refreshViewport()
		return m, nil
	case "clear":
		m.loop.ClearSession()
		m.viewBuf = statusStyle.Render("Session cleared.\n\n")
		m.streamBuf = ""
		m.inputTokens = 0
		m.outputTokens = 0
		m = m.refreshViewport()
		return m, nil
	case "compact":
		return m.handleCompactCommand()
	case "toggle-mode":
		return m.cyclePermissionMode()
	case "open-session-picker":
		return m.openSessionPicker()
	case "toggle-todo":
		m.todoPanel.toggle()
		if m.todoPanel.visible {
			m.refreshTodosFromDisk()
			m.viewBuf += statusStyle.Render("Todo panel: on\n\n")
		} else {
			m.viewBuf += statusStyle.Render("Todo panel: off\n\n")
		}
		m = m.refreshViewport()
		return m, nil
	}
	// "set-model <id>" — system actions can carry arguments after a space.
	if strings.HasPrefix(system, "set-model ") {
		id := strings.TrimPrefix(system, "set-model ")
		m.cfg.Model = id
		m.loop.Config.Model = id
		m.viewBuf += statusStyle.Render(fmt.Sprintf("Model changed to %s\n\n", id))
		m = m.refreshViewport()
		return m, nil
	}
	return m, nil
}

// transitionState changes state with debug logging
func (m *Model) transitionState(to appState, reason string) {
	if m.debugEnabled {
		from := m.state
		m.stateMachine.Transition(debug.State(stateToString(to)), reason)
		debug.Log(debug.EventStateChange, fmt.Sprintf("State: %s -> %s", stateToString(from), stateToString(to)), map[string]interface{}{
			"from":   stateToString(from),
			"to":     stateToString(to),
			"reason": reason,
		})
	}
	m.state = to
}

// stateToString converts appState to string for debugging
func stateToString(s appState) string {
	switch s {
	case stateInput:
		return "input"
	case stateBusy:
		return "busy"
	case statePicker:
		return "picker"
	case stateHelp:
		return "help"
	case statePermission:
		return "permission"
	case stateLoginProvider:
		return "login_provider"
	case stateLoginMethod:
		return "login_method"
	case stateLoginAPIKey:
		return "login_api_key"
	case stateLoginOAuth:
		return "login_oauth"
	case stateAskUser:
		return "ask_user"
	case statePalette:
		return "palette"
	case stateSessionPicker:
		return "session_picker"
	case stateMention:
		return "mention"
	case stateTodoPanel:
		return "todo_panel"
	case stateDebugPanel:
		return "debug_panel"
	default:
		return "unknown"
	}
}

// cyclePermissionMode advances the mode indicator to the next entry in
// permModeOrder. The badge in the status bar reflects the change
// immediately.
func (m Model) cyclePermissionMode() (tea.Model, tea.Cmd) {
	if m.loop == nil || m.loop.PermManager == nil {
		m.viewBuf += warnStyle.Render("No active permission manager.\n\n")
		m = m.refreshViewport()
		return m, nil
	}
	cur := m.loop.PermManager.Mode
	idx := 0
	for i, p := range m.permModeOrder {
		if p == cur {
			idx = i
			break
		}
	}
	idx = (idx + 1) % len(m.permModeOrder)
	next := m.permModeOrder[idx]
	m.loop.PermManager.Mode = next
	m.cfg.PermissionMode = next.String()

	// Persist the mode to project settings.json so it survives restarts
	s := &config.Settings{PermissionMode: next.String()}
	if err := config.WriteProject(s); err != nil {
		m.viewBuf += warnStyle.Render(fmt.Sprintf("Warning: could not save permission mode: %v\n\n", err))
	}

	// Mode is already displayed in status bar - no need to append to viewBuf
	m = m.refreshViewport()
	return m, nil
}

// toggleLastToolCard flips the expanded state of the most recent tool
// card. Mirrors opencode's Ctrl+T keybinding.

// handleCompactCommand manually triggers session compaction, summarizing
// the conversation history to save context space.
func (m Model) handleCompactCommand() (tea.Model, tea.Cmd) {
	if m.loop == nil {
		m.viewBuf += warnStyle.Render("No active session to compact.\n\n")
		m = m.refreshViewport()
		return m, nil
	}

	// Check if there are messages to compact
	if len(m.loop.Session.Messages) == 0 {
		m.viewBuf += warnStyle.Render("No messages to compact.\n\n")
		m = m.refreshViewport()
		return m, nil
	}

	m.viewBuf += statusStyle.Render("Compacting session...\n\n")
	m = m.refreshViewport()

	// Trigger compaction asynchronously
	return m, func() tea.Msg {
		ctx := context.Background()
		summary, err := runtime.CompactSession(ctx, m.loop.Client, m.loop.Config, m.loop.Session)
		if err != nil {
			return compactResultMsg{err: err}
		}
		return compactResultMsg{summary: summary}
	}
}


func (m Model) toggleLastToolCard() (tea.Model, tea.Cmd) {
	if len(m.toolCards) == 0 {
		return m, nil
	}
	idx := len(m.toolCards) - 1
	m.toolCards[idx].expanded = !m.toolCards[idx].expanded
	m.streamBuf = m.renderToolCards()
	m = m.refreshViewport()
	return m, nil
}

// handleSessionPickerKey handles keys when the session picker is open.
func (m Model) handleSessionPickerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	id, shouldClose := m.sessionPicker.updateKey(msg)
	if !shouldClose {
		return m, nil
	}
	m.sessionPicker.close()
	m.transitionState(stateInput, "session picker closed")
	if id == "" {
		return m, nil
	}
	if err := m.loop.LoadNamedSession(id); err != nil {
		m.viewBuf += errorStyle.Render(fmt.Sprintf("Error loading session %q: %v\n\n", id, err))
	} else {
		m.viewBuf += statusStyle.Render(fmt.Sprintf("Session loaded: %s (%d messages)\n\n", id, m.loop.MessageCount()))
	}
	m = m.refreshViewport()
	return m, nil
}

// openSessionPicker fetches session metadata and opens the picker.
func (m Model) openSessionPicker() (tea.Model, tea.Cmd) {
	metas, err := m.loop.ListSessionsWithMeta()
	if err != nil {
		m.viewBuf += errorStyle.Render(fmt.Sprintf("Error listing sessions: %v\n\n", err))
		m = m.refreshViewport()
		return m, nil
	}
	converted := make([]runtimeSessionMeta, 0, len(metas))
	for _, m := range metas {
		converted = append(converted, runtimeSessionMeta{
			id:             m.ID,
			updated:        m.UpdatedAt.Format("2006-01-02 15:04:05"),
			messageCount:   m.MessageCount,
			totalInTokens:  m.TotalInputTokens,
			totalOutTokens: m.TotalOutputTokens,
		})
	}
	m.sessionPicker.open(converted)
	m.transitionState(stateSessionPicker, "session picker opened")
	return m, nil
}

// handleBackgroundTasksKey handles keys when the background tasks panel is open.
func (m Model) handleBackgroundTasksKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+b", "esc", "q":
		// Close the panel
		m.backgroundTaskPanel.Hide()
		m.transitionState(stateInput, "user closed background tasks")
		return m, nil

	case "up", "k":
		// Move cursor up
		m.backgroundTaskPanel.MoveCursor(-1)
		return m, nil

	case "down", "j":
		// Move cursor down
		m.backgroundTaskPanel.MoveCursor(1)
		return m, nil

	case "c":
		// Clear completed tasks
		m.backgroundTaskPanel.ClearCompleted()
		return m, nil

	case "enter":
		// Show details of selected task
		task := m.backgroundTaskPanel.GetSelectedTask()
		if task != nil {
			m.viewBuf += statusStyle.Render(fmt.Sprintf("\n=== Task Details: %s ===\n", task.Name))
			m.viewBuf += statusStyle.Render(fmt.Sprintf("Status: %s\n", task.Status.String()))
			m.viewBuf += statusStyle.Render(fmt.Sprintf("Duration: %s\n", formatDuration(task.Duration())))
			if task.Description != "" {
				m.viewBuf += statusStyle.Render(fmt.Sprintf("Description: %s\n", task.Description))
			}
			if task.Output != "" {
				m.viewBuf += statusStyle.Render(fmt.Sprintf("Output:\n%s\n", task.Output))
			}
			if task.Error != nil {
				m.viewBuf += errorStyle.Render(fmt.Sprintf("Error: %v\n", task.Error))
			}
			m.viewBuf += "\n"
			m = m.refreshViewport()
		}
		return m, nil
	}

	return m, nil
}

// handleHistorySearchKey handles keys when the history search is open.
func (m Model) handleHistorySearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+r", "esc":
		// Close the search
		m.historySearch.Close()
		m.transitionState(stateInput, "user closed history search")
		return m, nil

	case "up", "ctrl+p":
		// Move cursor up
		m.historySearch.MoveCursor(-1)
		return m, nil

	case "down", "ctrl+n":
		// Move cursor down
		m.historySearch.MoveCursor(1)
		return m, nil

	case "enter":
		// Select the highlighted command
		if m.historySearch.HasResults() {
			selected := m.historySearch.GetSelected()
			if selected != "" {
				m.textarea.SetValue(selected)
				m.historySearch.Close()
				m.transitionState(stateInput, "user selected from history")
			}
		}
		return m, nil

	default:
		// Update the search input
		var cmd tea.Cmd
		m.historySearch.input, cmd = m.historySearch.input.Update(msg)
		m.historySearch.UpdateQuery(m.historySearch.input.Value())
		return m, cmd
	}
}

// handleQuickActionsKey handles keys when the quick actions menu is open.
func (m Model) handleQuickActionsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+k", "esc":
		// Close the menu
		m.quickActions.Close()
		m.transitionState(stateInput, "user closed quick actions")
		return m, nil

	case "up", "ctrl+p":
		// Move cursor up
		m.quickActions.MoveCursor(-1)
		return m, nil

	case "down", "ctrl+n":
		// Move cursor down
		m.quickActions.MoveCursor(1)
		return m, nil

	case "enter":
		// Execute the selected action
		if m.quickActions.HasActions() {
			action := m.quickActions.GetSelected()
			if action != nil {
				m.quickActions.Close()
				m.transitionState(stateInput, "executing quick action")

				// Execute the action handler
				var err error
				m, err = action.Handler(m)
				if err != nil {
					m.viewBuf += errorStyle.Render(fmt.Sprintf("Error: %v\n\n", err))
					m = m.refreshViewport()
				}
			}
		}
		return m, nil

	default:
		// Update the search input
		var cmd tea.Cmd
		m.quickActions.input, cmd = m.quickActions.input.Update(msg)
		m.quickActions.UpdateQuery(m.quickActions.input.Value())
		return m, cmd
	}
}

// handleConvSearchKey handles keys when the conversation search is open.
func (m Model) handleConvSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+f", "esc":
		// Close the search
		m.conversationSearch.Close()
		m.transitionState(stateInput, "user closed conversation search")
		return m, nil

	case "up", "ctrl+p":
		// Move cursor up
		m.conversationSearch.MoveCursor(-1)
		return m, nil

	case "down", "ctrl+n":
		// Move cursor down
		m.conversationSearch.MoveCursor(1)
		return m, nil

	case "enter":
		// Jump to the selected message
		if m.conversationSearch.HasResults() {
			selected := m.conversationSearch.GetSelected()
			if selected != nil {
				// Jump to the message in the viewport
				// We need to reconstruct the conversation view to show the message
				m.viewBuf += statusStyle.Render(fmt.Sprintf("Jumping to message #%d\n\n", selected.index+1))
				// In a full implementation, we would scroll the viewport to the message
				m.conversationSearch.Close()
				m.transitionState(stateInput, "jumped to message")
			}
		}
		return m, nil

	default:
		// Update the search input
		var cmd tea.Cmd
		m.conversationSearch.input, cmd = m.conversationSearch.input.Update(msg)
		m.conversationSearch.UpdateQuery(m.conversationSearch.input.Value())
		return m, cmd
	}
}

// handleMentionKey handles keys when the @-mention popup is active.
func (m Model) handleMentionKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		// Cancel: drop the @-trigger and return to plain input.
		m.mention.active = false
		m.transitionState(stateInput, "mention cancelled (esc)")
		return m, nil
	case tea.KeyEnter:
		return m.applyMentionSelection()
	case tea.KeyTab:
		return m.applyMentionSelection()
	case tea.KeyUp:
		m.mention.moveCursor(-1)
		return m, nil
	case tea.KeyDown:
		m.mention.moveCursor(1)
		return m, nil
	case tea.KeyBackspace:
		// Let the textarea handle the backspace; we'll re-evaluate the
		// popup state via the default key path on the next render.
		m.history.Reset()
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		if !m.mention.update(m.textarea.Value()) {
			m.transitionState(stateInput, "mention deactivated")
		}
		return m, cmd
	}
	// Default: pass through to textarea; re-evaluate mention state.
	m.history.Reset()
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	if !m.mention.update(m.textarea.Value()) {
		m.transitionState(stateInput, "mention deactivated")
	}
	return m, cmd
}

// handleSlashMenuKey handles keys when the slash command menu is open.
func (m Model) handleSlashMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		// Cancel: drop the slash menu and return to plain input.
		m.slashMenu.active = false
		m.transitionState(stateInput, "slash menu cancelled (esc)")
		return m, nil
	case tea.KeyEnter:
		return m.applySlashMenuSelection()
	case tea.KeyTab:
		return m.applySlashMenuSelection()
	case tea.KeyUp:
		m.slashMenu.moveCursor(-1)
		return m, nil
	case tea.KeyDown:
		m.slashMenu.moveCursor(1)
		return m, nil
	case tea.KeyBackspace:
		// Let the textarea handle the backspace; we'll re-evaluate the
		// popup state via the default key path on the next render.
		m.history.Reset()
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		if !m.slashMenu.update(m.textarea.Value()) {
			m.transitionState(stateInput, "slash menu deactivated")
		}
		return m, cmd
	}
	// Default: pass through to textarea; re-evaluate slash menu state.
	m.history.Reset()
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	if !m.slashMenu.update(m.textarea.Value()) {
		m.transitionState(stateInput, "slash menu deactivated")
	}
	return m, cmd
}

// applySlashMenuSelection replaces the /trigger and partial query in the
// textarea with the currently highlighted slash command.
func (m Model) applySlashMenuSelection() (tea.Model, tea.Cmd) {
	newVal, cursor := m.slashMenu.insert(m.textarea.Value())
	m.textarea.SetValue(newVal)
	m.textarea.SetCursor(cursor)
	m.slashMenu.active = false
	m.transitionState(stateInput, "slash menu cancelled (esc)")
	return m, nil
}

// applyMentionSelection replaces the @trigger and partial query in the
// textarea with the currently highlighted file path.
func (m Model) applyMentionSelection() (tea.Model, tea.Cmd) {
	newVal, cursor := m.mention.insert(m.textarea.Value())
	m.textarea.SetValue(newVal)
	m.textarea.SetCursor(cursor)
	m.mention.active = false
	m.transitionState(stateInput, "mention cancelled (esc)")
	return m, nil
}

// computeToolDiff produces a coloured diff for a file edit/write card.
// For a write, we use the file's current contents (after the write) as
// the "new" side and an empty string as "old" — that lets the user
// review the new file contents. For an edit, we attempt to read the
// original content from .claude/edit_previews/<id>.txt (a side-channel
// the runtime writes before applying the change); if that file is
// missing we just show the model's `new_string` input as additions.
func (m Model) computeToolDiff(c toolCard) string {
	// Pull the path and the new content from the card input. The input
	// is a JSON-shaped blob in the runtime; we don't try to fully
	// parse it here — we look for a "path" key and a "content" or
	// "new_string" key. If we can't find one we render a short stub.
	idx := strings.Index(c.input, `"path"`)
	path := ""
	if idx != -1 {
		path = extractJSONString(c.input[idx:])
	}
	newContent := ""
	if nci := strings.Index(c.input, `"new_string"`); nci != -1 {
		newContent = extractJSONString(c.input[nci:])
	} else if ci := strings.Index(c.input, `"content"`); ci != -1 {
		newContent = extractJSONString(c.input[ci:])
	}
	if path == "" && newContent == "" {
		return ""
	}
	// Read the previous version (if available). For write tools this
	// is empty; for edit tools we look for a side-channel.
	var oldContent string
	if c.name == "edit" || c.name == "file_edit" {
		// Best effort: try a couple of well-known paths.
		for _, p := range []string{".claude/edit_previews/" + c.id + ".txt", path + ".bak"} {
			if data, err := os.ReadFile(p); err == nil {
				oldContent = string(data)
				break
			}
		}
	}
	return inlineDiff(path, oldContent, newContent)
}

// extractJSONString pulls the value of a string field out of a JSON
// fragment. The runtime's tool input is a free-form blob so we do a
// minimal scan: find the colon after the key, the opening quote, and
// the matching closing quote (handling \" and \\ escapes).
func extractJSONString(s string) string {
	colon := strings.Index(s, ":")
	if colon == -1 {
		return ""
	}
	rest := s[colon+1:]
	for i := 0; i < len(rest); i++ {
		if rest[i] == '"' {
			var b strings.Builder
			for j := i + 1; j < len(rest); j++ {
				c := rest[j]
				if c == '\\' && j+1 < len(rest) {
					next := rest[j+1]
					switch next {
					case 'n':
						b.WriteByte('\n')
					case 't':
						b.WriteByte('\t')
					case '"':
						b.WriteByte('"')
					case '\\':
						b.WriteByte('\\')
					default:
						b.WriteByte(next)
					}
					j++
					continue
				}
				if c == '"' {
					return b.String()
				}
				b.WriteByte(c)
			}
			return b.String()
		}
		if rest[i] != ' ' && rest[i] != '\t' && rest[i] != '\n' {
			return ""
		}
	}
	return ""
}

// renderToolCards renders the in-progress tool cards into a string.
// Used when toggling expansion so the streaming view stays in sync.
func (m Model) renderToolCards() string {
	var b strings.Builder
	for _, c := range m.toolCards {
		b.WriteString(formatToolCard(c))
		b.WriteString("\n")
	}
	return b.String()
}

// refreshTodosFromDisk re-reads `.claude/todos.json` and updates the
// todo panel. Called whenever the agent runs todo_write so the panel
// reflects progress without us adding a dedicated event.
func (m Model) refreshTodosFromDisk() {
	data, err := os.ReadFile(".claude/todos.json")
	if err != nil {
		return
	}
	var items []struct {
		ID       string `json:"id"`
		Content  string `json:"content"`
		Status   string `json:"status"`
		Priority string `json:"priority"`
	}
	if err := json.Unmarshal(data, &items); err != nil {
		return
	}
	conv := make([]todoItem, len(items))
	for i, it := range items {
		conv[i] = todoItem{id: it.ID, content: it.Content, status: it.Status, priority: it.Priority}
	}
	m.todoPanel.setItems(conv)
}

// handlePickerKey handles keys when the model picker overlay is shown.
func (m Model) handlePickerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	models := m.activeModels()
	switch msg.Type {
	case tea.KeyEsc:
		m.state = stateInput
		return m, nil
	case tea.KeyEnter:
		chosen := models[m.pickerCursor]
		m.cfg.Model = chosen.id
		m.loop.Config.Model = chosen.id
		m.viewBuf += statusStyle.Render(fmt.Sprintf("Model changed to %s\n\n", chosen.id))
		m.transitionState(stateInput, "model selected from picker")
		m = m.refreshViewport()
		return m, nil
	case tea.KeyUp:
		if m.pickerCursor > 0 {
			m.pickerCursor--
		}
		return m, nil
	case tea.KeyDown:
		if m.pickerCursor < len(models)-1 {
			m.pickerCursor++
		}
		return m, nil
	case tea.KeyCtrlC:
		return m, tea.Quit
	}
	return m, nil
}

// handleHelpKey handles keys when the help overlay is shown.
func (m Model) handleHelpKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.Type == tea.KeyEsc, msg.Type == tea.KeyEnter, msg.String() == "q":
		m.state = stateInput
		return m, nil
	case msg.Type == tea.KeyCtrlC:
		return m, tea.Quit
	}
	return m, nil
}

// handlePermissionKey handles keys when a permission decision is pending.
func (m Model) handlePermissionKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var decision runtime.PermDecision
	handled := true

	switch msg.String() {
	case "y", "Y":
		decision = runtime.PermDecisionAllowOnce
	case "a", "A":
		decision = runtime.PermDecisionAllowAlways
	case "n", "N":
		decision = runtime.PermDecisionDeny
	default:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		handled = false
	}

	if !handled {
		return m, nil
	}

	ch := m.permReplyCh
	m.permReplyCh = nil
	m.permToolName = ""
	m.permToolInput = ""
	m.transitionState(stateBusy, "permission decision made")
	m = m.refreshViewport()

	return m, tea.Batch(
		func() tea.Msg {
			ch <- decision
			return nil
		},
		waitForStream(m.streamChan),
	)
}

// --- Mouse & Scroll Support -------------------------------------------------

// handleMouse processes mouse events (wheel, click, motion) for the TUI.
func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		// Scroll viewport up (3 lines per wheel click for smooth feel)
		m.viewport.LineUp(3)
		return m, nil

	case tea.MouseButtonWheelDown:
		// Scroll viewport down (3 lines per wheel click)
		m.viewport.LineDown(3)
		return m, nil

	case tea.MouseButtonLeft:
		// Click in input area: focus textarea
		inputStartY := m.viewportHeight() + 1 // +1 for header
		if msg.Y >= inputStartY && msg.Y <= inputStartY+textareaRows {
			// Click was in the text input area, focus it
			m.textarea.Focus() //nolint:errcheck
			return m, nil
		}

		// Click in viewport area: delegate to viewport for selection if needed
		if msg.Y >= 1 && msg.Y < inputStartY {
			// In the viewport area - scroll to approximate position
			// This provides click-to-position-on-scrollbar behavior
			vpHeight := m.viewportHeight()
			if vpHeight > 0 && msg.Y-1 >= 0 && msg.Y-1 < vpHeight {
				// Approximate: set cursor to y-offset relative to total content
				totalLines := m.viewport.TotalLineCount()
				if totalLines > vpHeight {
					ratio := float64(msg.Y-1) / float64(vpHeight)
					targetLine := int(ratio * float64(totalLines))
					m.viewport.SetYOffset(targetLine)
				}
			}
			return m, nil
		}

		// Click in hint/status bar area: no-op for now
		return m, nil

	case tea.MouseButtonRight:
		// Right click: nothing special for now
		return m, nil
	}

	// For all other mouse events (motion, release, etc.), delegate to viewport
	// which may handle text selection internally
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// renderViewportWithScrollbar wraps the viewport content with a scrollbar indicator.
// This shows the user their position within the conversation history.
func (m Model) renderViewportWithScrollbar() string {
	vpContent := m.viewport.View()

	totalLines := m.viewport.TotalLineCount()
	vpHeight := m.viewportHeight()

	// If content fits within viewport, no scrollbar needed
	if totalLines <= vpHeight {
		return vpContent
	}

	// Calculate scrollbar position and height
	offset := m.viewport.YOffset
	scrollPercent := float64(offset) / float64(totalLines-vpHeight)
	if scrollPercent > 1.0 {
		scrollPercent = 1.0
	}
	if scrollPercent < 0.0 {
		scrollPercent = 0.0
	}

	// Build the scrollbar on the right edge
	scrollHeight := max(1, int(float64(vpHeight)*float64(vpHeight)/float64(totalLines)))
	scrollPos := int(scrollPercent * float64(vpHeight-scrollHeight))

	// Create scrollbar column
	scrollStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	scrollTrackStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("235"))

	// Build the scrollbar
	lines := strings.Split(vpContent, "\n")
	for len(lines) < vpHeight {
		lines = append(lines, "")
	}

	// Truncate/pad to exact viewport height
	if len(lines) > vpHeight {
		lines = lines[:vpHeight]
	}

	sb := strings.Builder{}
	for i := 0; i < vpHeight; i++ {
		sb.WriteString(lines[i])
		sb.WriteString(" ") // spacer

		// Draw scrollbar indicator
		if i >= scrollPos && i < scrollPos+scrollHeight {
			sb.WriteString(scrollStyle.Render("█"))
		} else {
			sb.WriteString(scrollTrackStyle.Render("│"))
		}
		sb.WriteString("\n")
	}

	// Show percentage at bottom of scrollbar area
	scrollPct := int(scrollPercent * 100)
	if offset == 0 {
		scrollPct = 0
	} else if offset >= totalLines-vpHeight {
		scrollPct = 100
	}

	return sb.String()[:sb.Len()-1] + " " + statusStyle.Render(fmt.Sprintf("%d%%", scrollPct))
}

// --- View -------------------------------------------------------------------

// View is the Bubble Tea View function.
func (m Model) View() string {
	if !m.ready {
		return "Initializing…\n"
	}

	switch m.state {
	case statePicker:
		return m.viewPicker()
	case stateHelp:
		return m.viewHelp()
	case statePermission:
		return m.viewPermission()
	case stateAskUser:
		return m.viewAskUser()
	case stateLoginProvider:
		return m.viewLoginProvider()
	case stateLoginMethod:
		return m.viewLoginMethod()
	case stateLoginAPIKey:
		return m.viewLoginAPIKey()
	case stateLoginOAuth:
		return m.viewLoginOAuth()
	case statePalette:
		return m.palette.view(m.width, m.height)
	case stateSessionPicker:
		return m.sessionPicker.view(m.width, m.height)
	case stateDebugPanel:
		if m.debugEnabled {
			debug.SetPanelSize(m.width, m.height)
			return debug.RenderPanel()
		}
		return m.View() // Fallback if debug disabled
	case stateBackgroundTasks:
		return m.backgroundTaskPanel.View(m.width, m.height)
	case stateHistorySearch:
		return m.historySearch.View(m.width, m.height)
	case stateQuickActions:
		return m.quickActions.View(m, m.width, m.height)
	case stateConvSearch:
		return m.conversationSearch.View(m.width, m.height)
	}

	header := m.renderHeader()
	divider := dividerStyle.Render(strings.Repeat("─", m.width))

	// Use progressive disclosure for hints
	shortcuts := m.progressiveDisclosure.GetShortcutHints()
	hint := statusStyle.Render(strings.Join(shortcuts, "  "))

	// Add contextual hint if available
	contextHint := m.progressiveDisclosure.GetHintsForState("input", m.textarea.Value())
	if contextHint == "" {
		contextHint = m.progressiveDisclosure.GetCurrentHint()
	}

	statusLine := m.renderStatusBar()
	inputArea := m.renderInputArea()

	// Use scrollbar-enhanced viewport when content overflows
	vpView := m.renderViewportWithScrollbar()

	// Compose the main column (header, viewport, divider, input, hint, status).
	// When the todo panel is active on a wide terminal, lay it out as a
	// two-column row so the conversation can scroll beside the tasks.
	parts := []string{
		header,
		vpView,
		divider,
		inputArea,
		hint,
	}

	// Add contextual hint if available
	if contextHint != "" {
		parts = append(parts, statusStyle.Render(contextHint))
	}

	parts = append(parts, statusLine)

	mainCol := lipgloss.JoinVertical(lipgloss.Left, parts...)

	if m.todoPanel.isVisible() && m.width >= 100 {
		sidebar := lipgloss.NewStyle().Width(m.todoPanel.width).Render(
			m.todoPanel.view(m.viewportHeight()),
		)
		body := lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(m.width-m.todoPanel.width-2).Render(mainCol),
			"  ",
			sidebar,
		)
		out := body
		// Mention / slash-menu popups sit over the input area.
		if m.state == stateMention {
			out = lipgloss.JoinVertical(lipgloss.Left, out, m.mention.view(m.width, m.height))
		}
		if m.state == stateSlashMenu {
			out = lipgloss.JoinVertical(lipgloss.Left, out, m.slashMenu.view(m.width, m.height))
		}
		return out
	}

	out := mainCol
	if m.state == stateMention {
		out = lipgloss.JoinVertical(lipgloss.Left, out, m.mention.view(m.width, m.height))
	}
	if m.state == stateSlashMenu {
		out = lipgloss.JoinVertical(lipgloss.Left, out, m.slashMenu.view(m.width, m.height))
	}
	return out
}

func (m Model) renderHeader() string {
	title := headerStyle.Render("Claw Code v" + appVersion)
	tag := modelTagStyle.Render(fmt.Sprintf("  [%s] %s", m.cfg.ProviderName, m.cfg.Model))
	return title + tag
}

func (m Model) renderStatusBar() string {
	var parts []string

	// Render active status badges
	badges := m.badgeManager.Render()
	if badges != "" {
		parts = append(parts, badges)
	}

	// Add permission mode indicator
	mode := "default"
	if m.loop != nil && m.loop.PermManager != nil {
		mode = m.loop.PermManager.Mode.String()
	}
	modeIndicator := statusStyle.Render(fmt.Sprintf("[%s]", mode))
	parts = append(parts, modeIndicator)

	// Add token usage if available
	if m.inputTokens > 0 || m.outputTokens > 0 {
		tokens := statusStyle.Render(fmt.Sprintf("Tokens: %s in / %s out", formatNum(m.inputTokens), formatNum(m.outputTokens)))
		parts = append(parts, tokens)
	}

	// Add session info
	if m.loop != nil && m.loop.Session != nil {
		sessionInfo := statusStyle.Render(fmt.Sprintf("Session: %s (%d msgs)", m.loop.Session.ID, m.loop.MessageCount()))
		parts = append(parts, sessionInfo)
	}

	return strings.Join(parts, " │ ")
}

// renderInputArea renders the multi-line input with a "> " prefix on the first line.
func (m Model) renderInputArea() string {
	prompt := inputPromptStyle.Render("> ")
	tv := m.textarea.View()
	lines := strings.Split(tv, "\n")
	result := make([]string, len(lines))
	for i, line := range lines {
		if i == 0 {
			result[i] = prompt + line
		} else {
			result[i] = "  " + line
		}
	}
	return strings.Join(result, "\n")
}

// activeModels returns the model list appropriate for the current provider.
func (m Model) activeModels() []modelEntry {
	switch m.cfg.ProviderName {
	case "openai":
		return openAIModels
	case "deepseek":
		return deepseekModels
	default:
		return anthropicModels
	}
}

func (m Model) viewPicker() string {
	var b strings.Builder
	b.WriteString(pickerHeaderStyle.Render("Select Model") + "\n")
	b.WriteString(statusStyle.Render("  ↑/↓ navigate  Enter select  Esc cancel") + "\n\n")

	for i, km := range m.activeModels() {
		cursor := "  "
		style := unselectedModelStyle
		if i == m.pickerCursor {
			cursor = "▶ "
			style = selectedModelStyle
		}
		b.WriteString(cursor + style.Render(km.id) + "\n")
		b.WriteString("    " + statusStyle.Render(km.desc) + "\n")
	}

	return b.String()
}

func (m Model) viewHelp() string {
	content := lipgloss.JoinVertical(lipgloss.Left,
		headerStyle.Render("Claw Code — Commands"),
		"",
		statusStyle.Render("Auth & Provider"),
		"  "+userLabelStyle.Render("/login")+"                          Multi-provider login flow",
		"  "+userLabelStyle.Render("/auth")+" login|logout|status       Legacy OAuth commands",
		"",
		statusStyle.Render("Model & Config"),
		"  "+userLabelStyle.Render("/model")+"                          Change the active model (picker)",
		"  "+userLabelStyle.Render("/config")+"                         Show all config values",
		"  "+userLabelStyle.Render("/config")+" <key>                   Show one config value",
		"  "+userLabelStyle.Render("/config")+" <key> <value>           Set config value (saves to project)",
		"  "+userLabelStyle.Render("/init")+"                           Create .claude/settings.json",
		"  "+userLabelStyle.Render("/theme")+" dark|light               Switch TUI color theme",
		"  "+userLabelStyle.Render("/status")+"                         Show model/provider/session info",
		"  "+userLabelStyle.Render("/cost")+"                           Show token usage this session",
		"",
		statusStyle.Render("Session"),
		"  "+userLabelStyle.Render("/clear")+"                          Clear conversation history",
		"  "+userLabelStyle.Render("/session")+" list                   List saved sessions",
		"  "+userLabelStyle.Render("/session")+" save [name]            Save current session",
		"  "+userLabelStyle.Render("/session")+" load <name>            Load a saved session",
		"  "+userLabelStyle.Render("/session-list")+"                   Alias for /session list",
		"",
		statusStyle.Render("Other"),
		"  "+userLabelStyle.Render("/help")+"                           Show this help",
		"  "+userLabelStyle.Render("/exit")+" / "+userLabelStyle.Render("/quit")+"                     Exit (session auto-saved)",
		"",
		statusStyle.Render("Input:"),
		"  "+userLabelStyle.Render("Enter")+"          Send message",
		"  "+userLabelStyle.Render("Ctrl+J")+"         Insert newline (multi-line input)",
		"  "+userLabelStyle.Render("↑ / ↓")+"          Navigate input history (single-line mode)",
		"  "+userLabelStyle.Render("PgUp / PgDn")+"    Scroll conversation",
		"  "+userLabelStyle.Render("Mouse wheel")+"     Scroll conversation smoothly",
		"  "+userLabelStyle.Render("Ctrl+C")+"         Exit",
		"",
		statusStyle.Render("Esc / Enter / q to close this panel"),
	)
	box := helpBoxStyle.Width(min(80, m.width-4)).Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m Model) viewPermission() string {
	tool := userLabelStyle.Render(m.permToolName)
	inp := m.permToolInput
	if len(inp) > 60 {
		inp = inp[:60] + "..."
	}
	prompt := fmt.Sprintf("Allow %s: %s?", tool, inp)
	hint := statusStyle.Render("[y]es-once  [a]lways  [n]o")
	content := lipgloss.JoinVertical(lipgloss.Left,
		headerStyle.Render("Permission Required"),
		"",
		"  "+prompt,
		"",
		"  "+hint,
	)
	box := helpBoxStyle.Width(min(72, m.width-4)).Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

// viewLoginProvider renders the provider selection screen.
func (m Model) viewLoginProvider() string {
	var b strings.Builder
	b.WriteString(pickerHeaderStyle.Render("Login — Choose Provider") + "\n")
	b.WriteString(statusStyle.Render("  ↑/↓ navigate  Enter select  Esc cancel") + "\n\n")
	for i, p := range loginProviders {
		cursor := "  "
		style := unselectedModelStyle
		if i == m.loginCursor {
			cursor = "▶ "
			style = selectedModelStyle
		}
		b.WriteString(cursor + style.Render(p.name) + "\n")
		b.WriteString("    " + statusStyle.Render(p.desc) + "\n")
	}
	box := helpBoxStyle.Width(min(60, m.width-4)).Render(b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

// viewLoginMethod renders the auth-method selection screen (Anthropic).
func (m Model) viewLoginMethod() string {
	var b strings.Builder
	b.WriteString(pickerHeaderStyle.Render("Login — Anthropic Auth Method") + "\n")
	b.WriteString(statusStyle.Render("  ↑/↓ navigate  Enter select  Esc back") + "\n\n")
	for i, meth := range anthropicAuthMethods {
		cursor := "  "
		style := unselectedModelStyle
		if i == m.loginCursor {
			cursor = "▶ "
			style = selectedModelStyle
		}
		b.WriteString(cursor + style.Render(meth.name) + "\n")
		b.WriteString("    " + statusStyle.Render(meth.desc) + "\n")
	}
	box := helpBoxStyle.Width(min(60, m.width-4)).Render(b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

// viewLoginAPIKey renders the API key entry screen.
func (m Model) viewLoginAPIKey() string {
	providerName := m.loginProvider
	if providerName == "" {
		providerName = "provider"
	}
	content := lipgloss.JoinVertical(lipgloss.Left,
		pickerHeaderStyle.Render(fmt.Sprintf("Login — %s API Key", strings.Title(providerName))), //nolint:staticcheck
		"",
		"  "+statusStyle.Render("Paste your API key below (input is masked):"),
		"  "+m.loginKeyInput.View(),
		"",
		"  "+statusStyle.Render("Enter to confirm  •  Esc to cancel"),
	)
	box := helpBoxStyle.Width(min(70, m.width-4)).Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

// viewLoginOAuth renders the OAuth waiting screen.
func (m Model) viewLoginOAuth() string {
	content := lipgloss.JoinVertical(lipgloss.Left,
		pickerHeaderStyle.Render("Login — Anthropic OAuth"),
		"",
		"  "+m.spinner.View()+" "+statusStyle.Render("Waiting for browser login…"),
		"",
		"  "+statusStyle.Render("Complete the login in your browser, then return here."),
		"  "+statusStyle.Render("Ctrl+C to quit."),
	)
	box := helpBoxStyle.Width(min(70, m.width-4)).Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

// --- Helpers ----------------------------------------------------------------

// waitForStream returns a tea.Cmd that reads the next event from the stream channel.
func waitForStream(ch <-chan runtime.TurnEvent) tea.Cmd {
	return func() tea.Msg {
		for {
			ev, ok := <-ch
			if !ok {
				return streamDoneMsg{}
			}
			switch ev.Type {
			case runtime.TurnEventTextDelta:
				return streamDeltaMsg{text: ev.Text}
			case runtime.TurnEventTextFinal:
				return streamTextFinalMsg{text: ev.Text}
			case runtime.TurnEventToolStart:
				return streamToolMsg{name: ev.ToolName, input: ev.ToolInput}
			case runtime.TurnEventToolDone:
				return streamToolDoneMsg{name: ev.ToolName, result: ev.ToolResult}
			case runtime.TurnEventUsage:
				return streamUsageMsg{inputTokens: ev.InputTokens, outputTokens: ev.OutputTokens}
			case runtime.TurnEventDone:
				return streamDoneMsg{}
			case runtime.TurnEventError:
				return streamErrMsg{err: ev.Err}
			case runtime.TurnEventWarn:
				return streamWarnMsg{text: ev.Text}
			case runtime.TurnEventPermissionAsk:
				return streamPermAskMsg{name: ev.ToolName, input: ev.ToolInput, reply: ev.PermReply}
			case runtime.TurnEventAskUser:
				return streamAskUserMsg{question: ev.ToolInput, reply: ev.AskUserReply}
			}
		}
	}
}

// initViewport creates the viewport and sets textarea width on first resize.
func (m Model) initViewport() Model {
	vpHeight := m.viewportHeight()
	m.viewport = viewport.New(m.width, vpHeight)
	m.viewport.SetContent(m.viewBuf)
	m.viewport.GotoBottom()
	m.textarea.SetWidth(max(m.width-2, 10))
	return m
}

// resizeViewport updates dimensions after a window resize.
func (m Model) resizeViewport() Model {
	m.viewport.Width = m.width
	m.viewport.Height = m.viewportHeight()
	m.textarea.SetWidth(max(m.width-2, 10))
	return m
}

// viewportHeight calculates the viewport height from the terminal height.
// Layout overhead: header(1) + divider(1) + textarea(textareaRows) + hint(1) + status(1).
func (m Model) viewportHeight() int {
	// Minimum terminal height to avoid panic: 10 lines
	// (header + minimal viewport + textarea + hints)
	const minTerminalHeight = 10
	if m.height < minTerminalHeight {
		return 1 // Return minimal viewport, UI will be cramped but won't panic
	}

	overhead := 4 + textareaRows // header + divider + textarea + hint + status
	h := m.height - overhead
	if h < 1 {
		h = 1
	}
	return h
}

// refreshViewport rebuilds viewport content from current buffers.
// Optimized to reduce expensive SetContent calls during streaming.
func (m Model) refreshViewport() Model {
	// Build content once
	var content strings.Builder
	content.WriteString(m.viewBuf)
	content.WriteString(m.streamBuf)

	if m.state == stateBusy && !m.hasStreamContent {
		content.WriteString(m.spinner.View())
		content.WriteString(statusStyle.Render(" Thinking…\n"))
	}

	// Only update if content changed (use length check as fast path)
	newContent := content.String()
	currentContent := m.viewport.View()

	// Fast path: if lengths differ, content definitely changed
	if len(currentContent) != len(newContent) || currentContent != newContent {
		m.viewport.SetContent(newContent)
		m.viewport.GotoBottom()
	}

	return m
}

// truncate shortens s to at most n runes, appending "…" if truncated.
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "…"
}

// formatNum formats an integer with comma separators.
// formatSessionList renders a table-style listing of session metadata.
func formatSessionList(metas []runtime.SessionMeta) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%-40s  %-19s  %6s  %8s  %8s\n",
		"ID", "Updated", "Msgs", "In tok", "Out tok"))
	sb.WriteString(strings.Repeat("-", 90) + "\n")
	for _, m := range metas {
		ts := m.UpdatedAt.Format("2006-01-02 15:04:05")
		id := m.ID
		if len(id) > 38 {
			id = id[:35] + "..."
		}
		sb.WriteString(fmt.Sprintf("%-40s  %-19s  %6d  %8s  %8s\n",
			id, ts, m.MessageCount,
			formatNum(m.TotalInputTokens),
			formatNum(m.TotalOutputTokens)))
	}
	return sb.String()
}

// extractFilePath extracts file path from tool input JSON
func extractFilePath(input string) string {
	// Look for "path" or "file_path" field
	if idx := strings.Index(input, `"path"`); idx != -1 {
		return extractJSONString(input[idx:])
	}
	if idx := strings.Index(input, `"file_path"`); idx != -1 {
		return extractJSONString(input[idx:])
	}
	return ""
}

// recordFileChange extracts file change info from a tool card and records it
func (m *Model) recordFileChange(card toolCard) {
	// Only record for file modification tools
	if card.name != "write_file" && card.name != "file_edit" && card.name != "edit" && card.name != "write" {
		return
	}

	// Try to extract file path from input
	filePath := extractFilePath(card.input)
	if filePath == "" {
		return
	}

	// Read current content (after change)
	after, err := os.ReadFile(filePath)
	if err != nil {
		return // File might not exist yet
	}

	// For write operations, before is empty or previous content
	// For edit operations, try to get before content from backup
	before := ""
	if card.name == "file_edit" || card.name == "edit" {
		// Try to read backup file if it exists
		backupPath := filePath + ".bak"
		if data, err := os.ReadFile(backupPath); err == nil {
			before = string(data)
		}
	}

	// Record the change
	operation := "write"
	if card.name == "file_edit" || card.name == "edit" {
		operation = "edit"
	}

	m.fileHistory.Record(FileChange{
		FilePath:  filePath,
		Before:    before,
		After:     string(after),
		Timestamp: time.Now(),
		Operation: operation,
	})
}

func formatNum(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var result []byte
	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, c)
	}
	return string(result)
}
