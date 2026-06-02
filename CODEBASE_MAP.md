# Claw-Code-Go Codebase Map

## Project Overview

**claw-code-go** is a Go implementation of an AI coding assistant (inspired by Claude Code). It provides a terminal-based interface (TUI) for conversational coding with multi-provider AI backends, autonomous task execution, and a comprehensive tool system.

- **Language:** Go 1.24.2
- **TUI Framework:** Bubble Tea (charmbracelet) + lipgloss + glamour
- **Module path:** `claw-code-go`
- **Binary output:** `main` (26MB compiled)
- **Entry point:** `cmd/claw-code-go/main.go` (211 lines)
- **Total Go source:** ~6,400 lines (TUI) + ~4,500 lines (runtime/providers/tools) = ~11,000 lines

---

## Top-Level Directory Structure

```
/home/chaim/claw-code-go/
├── assets/                    # Static assets (logo, screenshot)
├── cmd/claw-code-go/          # Main binary entry point
│   └── main.go                # CLI flag parsing, provider init, TUI launch
├── config/                    # Configuration files
│   └── mcporter.json          # MCP port bridge config
├── .bob/                      # Bob build tool workspace
├── .claude/                   # Claude project config (todos, settings)
├── .codegraph/                # Code graph analysis data
├── .cursor/                   # Cursor IDE rules
├── internal/                  # All core application code (14 sub-packages)
│   ├── api/                   # Provider interface + HTTP clients
│   │   ├── providers/         # Multi-provider implementations
│   │   │   ├── anthropic/     # Anthropic Claude API
│   │   │   ├── bedrock/       # AWS Bedrock
│   │   │   ├── deepseek/      # DeepSeek web chat API (newest, most complex)
│   │   │   ├── foundry/       # Azure AI Foundry
│   │   │   ├── openai/        # OpenAI API
│   │   │   └── vertex/        # Google Vertex AI
│   │   ├── client.go          # Anthropic HTTP streaming client
│   │   ├── provider.go        # Provider/APIClient interfaces
│   │   ├── types.go           # ContentBlock, Tool, StreamEvent types
│   │   ├── toolcall.go        # Tool call parsing utilities
│   │   ├── xml_toolcalls.go   # XML-format tool call parsing
│   │   └── strip.go           # Response text stripping utilities
│   ├── auth/                  # Multi-provider credential management
│   │   ├── credentials.go     # Env-var + file credential resolution
│   │   ├── manager.go         # Credential manager (login/logout/status)
│   │   ├── oauth.go           # OAuth 2.0 device flow
│   │   └── storage.go         # JSON file credential persistence
│   ├── commands/              # Slash-command handlers (/login, /mcp, etc.)
│   │   ├── registry.go        # Command registry and dispatcher
│   │   ├── auth.go            # /login, /logout, /status commands
│   │   └── mcp.go             # /mcp command for MCP server management
│   ├── compat/                # Compatibility/utility subcommands
│   │   ├── harness.go         # Subcommand runners (dump-manifests, etc.)
│   │   └── manifest.go        # Tool/slash-command manifest definitions
│   ├── config/                # Configuration loading
│   │   └── loader.go          # JSON config file loader
│   ├── context/               # System prompt context assembly
│   │   ├── assembler.go       # Environment, git, memory context injection
│   │   ├── git.go             # Git status/commit context
│   │   ├── memory.go          # CLAUDE.md / project memory files
│   │   └── sysinfo.go         # System information context
│   ├── mcp/                   # Model Context Protocol support
│   │   ├── client.go          # MCP client interface
│   │   ├── registry.go        # MCP server registry
│   │   ├── sse.go             # SSE transport
│   │   ├── stdio.go           # Stdio transport
│   │   └── types.go           # MCP type definitions
│   ├── permissions/           # Permission management system
│   │   ├── manager.go         # Permission manager
│   │   ├── mode.go            # Permission modes (default, accept-edits, bypass, plan)
│   │   └── rules.go           # Permission rules engine
│   ├── runtime/               # Core runtime: conversation loop, streaming, sessions
│   │   ├── conversation.go    # ConversationLoop - THE CORE (~1156 lines)
│   │   ├── autonomous.go      # Autonomous task execution (RunTask)
│   │   ├── config.go          # Runtime configuration
│   │   ├── session.go         # Session management
│   │   ├── provider_factory.go # Multi-provider client factory
│   │   ├── compact.go         # Token compaction/context compression
│   │   ├── events.go          # Event types
│   │   ├── message_limits.go  # Message/token limit enforcement
│   │   └── permissions.go     # Permission integration
│   ├── tools/                 # All 10 tool implementations
│   │   ├── bash.go            # Shell command execution
│   │   ├── files.go           # read_file, write_file tools
│   │   ├── file_edit.go       # String replacement file editing
│   │   ├── glob.go            # File glob pattern matching
│   │   ├── grep.go            # Regex search in files
│   │   ├── web_fetch.go       # URL content fetching
│   │   ├── web_search.go      # Web search (Brave API / DuckDuckGo)
│   │   ├── ask_user.go        # User question tool
│   │   └── todo_write.go      # Task list management
│   ├── tui/                   # Terminal UI - Bubble Tea (~6,400 lines)
│   │   ├── model.go           # Main TUI model (2,163 lines) - CORE UI
│   │   ├── tool_parser.go     # Smart tool call XML parsing (518 lines)
│   │   ├── streaming_renderer.go # Real-time markdown streaming (345 lines)
│   │   ├── debug_handlers.go  # Debug key bindings (61 lines)
│   │   ├── debug/             # Event-driven debug subsystem
│   │   │   ├── events.go      # Structured event logging (277 lines)
│   │   │   ├── panel.go       # Debug panel UI with 4 tabs (460 lines)
│   │   │   └── state_machine.go # State transition tracking (308 lines)
│   │   ├── diff.go            # Diff display rendering (145 lines)
│   │   ├── history.go         # Command history (63 lines)
│   │   ├── keymap.go          # Key binding definitions (17 lines)
│   │   ├── logo.go            # Logo display (56 lines)
│   │   ├── markdown.go        # Markdown rendering (103 lines)
│   │   ├── mention.go         # @mention autocomplete (211 lines)
│   │   ├── palette.go         # Ctrl+P command palette (232 lines)
│   │   ├── session_picker.go  # Session browser/loader (128 lines)
│   │   ├── slash_menu.go      # / slash command autocomplete (221 lines) [NEW]
│   │   ├── state_helpers.go   # State management utilities (97 lines)
│   │   ├── styles.go          # Lipgloss styles (212 lines)
│   │   ├── theme.go           # Theme definitions (73 lines)
│   │   ├── todo_panel.go      # Todo list sidebar (108 lines)
│   │   ├── tool_card.go       # Tool call card rendering (237 lines)
│   │   └── tui_test.go        # TUI tests (349 lines)
│   └── usage/                 # Token usage tracking
│       └── tracker.go         # Per-session usage tracker
├── mimo-cli/                  # External Node.js CLI tool (separate project)
│   └── node_modules/          # npm dependencies (chrome-devtools-mcp, puppeteer, etc.)
├── go.mod                     # Go module definition
├── go.sum                     # Dependency checksums
├── README.md                  # Project documentation
├── CODEBASE_MAP.md            # This file
├── CODE_REVIEW_FINDINGS.md    # Code review findings document
├── CONNECTION_DROP_FIX.md     # Connection drop fix analysis
├── REVIEW_AND_RECOMMENDATIONS.md # Review and recommendations
├── SLASH_COMMAND_MENU_DESIGN.md  # Slash command menu design spec
├── TUI_BUG_FIX.md             # TUI bug fix analysis (repeated tool_calls)
├── TUI_ENHANCEMENTS.md        # TUI enhancements documentation
├── TUI_FEATURES.md            # TUI features overview
└── test-session-recording.log # Test session recording
```

---

## Architecture

### Core Loop: ConversationLoop

The heart of the application is `ConversationLoop` (~1156 lines in `internal/runtime/conversation.go`), which manages:
- **Agentic conversation** with tool-use loops
- **Streaming responses** from AI providers
- **Tool execution** and result injection back into the conversation
- **Context assembly** (environment, git, CLAUDE.md)
- **Token tracking** and compaction
- **MCP server integration** for extended tool capabilities
- **Permission checks** before tool execution

### Key Interfaces

```go
// All provider clients implement this single interface
APIClient interface {
    StreamResponse(ctx, CreateMessageRequest) (<-chan StreamEvent, error)
}

// All providers implement this factory interface
Provider interface {
    Name() string
    NewClient(ProviderConfig) (APIClient, error)
    AuthMethod() AuthMethod
}
```

### Tool System

10 tools exposed to the AI:
1. **bash** - Shell command execution (30s timeout, 10KB output limit)
2. **read_file** / **write_file** - File I/O
3. **file_edit** - Exact string replacement in files
4. **glob** - File pattern matching
5. **grep** - Regex search
6. **web_fetch** - URL content retrieval
7. **web_search** - Web search (Brave API / DuckDuckGo)
8. **ask_user** - Pause for user input
9. **todo_write** - Task list management

Each tool defines its JSON schema via `api.Tool` and implements `Execute(input map[string]any) (string, error)`.

---

## TUI Architecture (Recent Enhancements)

The TUI has been significantly expanded in recent commits with a comprehensive enhancement system:

### TUI State Machine

```
input ←→ busy ←→ permission
  ↓       ↓
picker  ask_user
  ↓
palette (Ctrl+P)
  ↓
session_picker
  ↓
slash_menu (typing / in textarea)
  ↓
mention (typing @ in textarea)
  ↓
debug_panel (Ctrl+D when CLAW_DEBUG=1)
```

### Key TUI Components

| File | Lines | Purpose |
|------|-------|---------|
| `model.go` | 2,163 | Main TUI model with all state handling, key dispatch, view rendering |
| `tool_parser.go` | 518 | Real-time XML tool call extraction, card rendering, status tracking |
| `streaming_renderer.go` | 345 | Throttled real-time markdown rendering during API streaming |
| `tool_card.go` | 237 | Pretty tool call cards with expand/collapse |
| `palette.go` | 232 | Ctrl+P command palette with fuzzy matching |
| `slash_menu.go` | 221 | **NEW** Interactive `/` autocomplete (like @mention but for slash commands) |
| `mention.go` | 211 | @file autocomplete with fuzzy file matching |
| `styles.go` | 212 | Lipgloss styles for all TUI elements |
| `diff.go` | 145 | Diff rendering for file edits |
| `markdown.go` | 103 | Glamour-based markdown with syntax highlighting |
| `todo_panel.go` | 108 | Sidebar todo list display |
| `debug/panel.go` | 460 | Debug overlay with Events/States/Metrics/Tools tabs |
| `debug/events.go` | 277 | Structured event logging with circular buffer |
| `debug/state_machine.go` | 308 | State transition validation and metrics |

### Streaming Pipeline

```
API Stream Chunk
    ↓
ToolCallParser.ProcessStream()  — extracts XML tool calls, removes from visible text
    ↓
StreamingRenderer.Append()      — appends clean text, throttles markdown rendering
    ↓
StreamingRenderer.Render()      — glamour markdown → styled terminal output
    ↓
Viewport display                — merged with tool call cards
```

### Slash Command Menu (NEW - `slash_menu.go`)

Interactive `/` command autocomplete that appears when typing `/` in the textarea:
- Floating popup above input area with all 14 slash commands
- Real-time fuzzy filtering as user types
- ↑/↓ navigation, Tab to autocomplete, Enter to select
- Esc to dismiss without clearing input
- Uses same token-based scoring as command palette

---

## Provider Architecture

Six AI providers supported via a uniform `api.APIClient` interface:

| Provider | Package | Auth Method | Status |
|----------|---------|-------------|--------|
| Anthropic | `anthropic` | API Key / OAuth | Stable |
| OpenAI | `openai` | API Key | Stable |
| AWS Bedrock | `bedrock` | IAM | Stable |
| Vertex AI | `vertex` | ADC | Stable |
| Azure Foundry | `foundry` | Azure Identity | Stable |
| DeepSeek | `deepseek` | Token (web chat) | Newest, most complex |

Provider selection flows through `runtime.SelectProvider()` → `provider.NewClient()`.

### DeepSeek Provider (Newest Addition)

The DeepSeek provider is the most complex, using:
- **WASM solver** (`sha3_wasm_bg.wasm`, `wasm_solver.go`) for challenge-response auth
- **Web client** (`webclient.go`) for HTTP polling
- **Stream parsing** (`stream.go`) for SSE + polling hybrid
- **Models** (`models.go`) with rate limiting and TPD tracking
- **Rate limits** (`limits.go`) for usage quota management
- **Live tests** for full integration testing

---

## Permission System

Four permission modes:
- **default** - Ask user for each file write/bash execution
- **accept-edits** - Auto-accept file edits, ask for bash
- **bypass** - Auto-accept everything (used by autonomous mode)
- **plan** - Plan mode (no execution)

Ruleset loaded from `.claude/settings.json` with allowedTools/blockedTools lists.

---

## Autonomous Mode

`RunTask()` in `autonomous.go` enables fully autonomous execution:
- Forces `bypass` permission mode
- Strips `ask_user` from tool list
- Uses LLM-judge for DONE/CONTINUE decisions
- Respects `MaxTurns` limit

---

## Context Assembly

The `context/assembler.go` injects environment context into the system prompt:
- Working directory and OS info
- Git branch, status, and recent commits
- CLAUDE.md / project memory files
- Date/time

---

## Data Flow

```
User Input (TUI textarea)
    │
    ▼
ConversationLoop.SendMessage()
    │
    ▼
Build system prompt + message history + tools
    │
    ▼
APIClient.StreamResponse()  ← provider-specific HTTP/SSE
    │
    ▼
Parse streaming events (text / tool_use / stop)
    │
    ├─► Text delta → emit to TUI as streamDeltaMsg
    │       │
    │       ▼
    │   ToolCallParser.ProcessStream()  (hides XML tool calls)
    │       │
    │       ▼
    │   StreamingRenderer.Append()  (real-time markdown)
    │       │
    │       ▼
    │   Viewport display
    │
    └─► Tool use → Execute tool → inject tool_result → loop
```

---

## Key Files by Line Count

| File | Lines | Role |
|------|-------|------|
| `internal/tui/model.go` | 2,163 | Main TUI model |
| `internal/runtime/conversation.go` | 1,156 | Core agent loop |
| `internal/api/providers/deepseek/provider_test.go` | 732 | DeepSeek tests |
| `internal/api/providers/deepseek/provider.go` | 608 | DeepSeek provider |
| `internal/tui/tool_parser.go` | 518 | Tool call parsing & display |
| `internal/api/xml_toolcalls.go` | 466 | XML tool call parsing |
| `internal/api/providers/deepseek/webclient.go` | 475 | DeepSeek HTTP client |
| `internal/api/providers/deepseek/stream.go` | 465 | DeepSeek stream handling |
| `internal/tui/debug/panel.go` | 460 | Debug panel UI |
| `internal/tui/tui_test.go` | 349 | TUI tests |
| `internal/tui/streaming_renderer.go` | 345 | Real-time markdown streaming |
| `internal/tui/debug/state_machine.go` | 308 | State machine tracking |
| `internal/tui/debug/events.go` | 277 | Event logging system |
| `internal/runtime/autonomous.go` | 248 | Autonomous task runner |
| `internal/tui/tool_card.go` | 237 | Tool card rendering |
| `internal/tui/palette.go` | 232 | Command palette |
| `internal/api/client.go` | 231 | Anthropic HTTP client |
| `internal/tools/web_search.go` | 226 | Web search implementation |
| `internal/tui/slash_menu.go` | 221 | Slash command autocomplete [NEW] |
| `internal/api/providers/deepseek/models.go` | 218 | DeepSeek model definitions |
| `internal/tui/styles.go` | 212 | Lipgloss styles |
| `cmd/claw-code-go/main.go` | 211 | CLI entry point |
| `internal/tui/mention.go` | 211 | @mention autocomplete |

---

## Current Git Status

- **Branch:** master
- **Modified (uncommitted):**
  - `TUI_BUG_FIX.md` - Analysis of repeated "tool_calls pending" bug
  - `internal/tui/model.go` - Main TUI model modifications
  - `internal/tui/tool_parser.go` - Tool parser fixes
  - `main` - Compiled binary
- **Untracked (new):**
  - `internal/tui/slash_menu.go` - Slash command autocomplete menu (221 lines)
- **Last commits:**
  - `a69825d` - Fix critical integration issues in TUI enhancement system
  - `6ec425e` - Add comprehensive TUI enhancement system
  - `086b957` - chore: initial import of claw-code-go fork
  - `535bb60` - feat: integrate deepseek provider into claw-code-go

---

## Build & Dependencies

- **Go version:** 1.24.2
- **Core dependency:** Bubble Tea (charmbracelet) v1.3.10 for TUI
- **Styling:** lipgloss v1.1.1, glamour v1.0.0 (markdown), chroma v2.20.0 (syntax highlighting)
- **WASM runtime:** wazero v1.11.0 (tetratelabs) for DeepSeek challenge solver
- **No CGO or system dependencies** — pure Go binary
- **Binary size:** ~26MB (static, no external libs)

---

## Slash Commands (14 total)

| Command | Description | Args |
|---------|-------------|------|
| `/help` | Show available commands | |
| `/model` | Change the active model (picker) | |
| `/login` | Multi-provider login flow | |
| `/clear` | Clear conversation history | |
| `/theme` | Switch TUI color theme | dark\|light |
| `/status` | Show model/provider/session info | |
| `/cost` | Show token usage this session | |
| `/config` | Show or set config values | [key] [value] |
| `/session` | Manage sessions | list\|save\|load [name] |
| `/sessions` | Browse saved sessions (picker) | |
| `/todo` | Toggle todo list sidebar | |
| `/init` | Create .claude/settings.json | |
| `/exit` | Exit (session auto-saved) | |
| `/quit` | Exit (session auto-saved) | |

---

## Key Design Patterns

1. **Bubble Tea Elm Architecture:** Model/Update/View pattern with message-driven state
2. **Provider Interface:** Uniform `APIClient` interface for all 6 AI providers
3. **Stream Processing Pipeline:** Chained processors (tool parser → markdown renderer → viewport)
4. **Event-Driven Debugging:** Structured event logging with state machine tracking
5. **Fuzzy Matching:** Token-based scoring used in both palette and slash menu
6. **Graceful Degradation:** Falls back to NoAuthClient when no credentials, allows /login in TUI
7. **Session Persistence:** Auto-save on exit, resume via `--session` flag