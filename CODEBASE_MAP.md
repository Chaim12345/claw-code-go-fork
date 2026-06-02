# Claw-Code-Go Codebase Map

## Project Overview

**claw-code-go** is a Go implementation of an AI coding assistant (inspired by Claude Code). It provides a terminal-based interface (TUI) for conversational coding with multi-provider AI backends, autonomous task execution, and a comprehensive tool system.

- **Language:** Go 1.24.2
- **TUI Framework:** Bubble Tea (charmbracelet)
- **Module path:** `claw-code-go`
- **Binary output:** `main` (26MB compiled)
- **Entry point:** `cmd/claw-code-go/main.go` (211 lines)
- **Total Go source:** ~5,400+ lines (excluding test/auxiliary files)

---

## Top-Level Directory Structure

```
/home/chaim/claw-code-go/
├── assets/                    # Static assets (logo, etc.)
├── cmd/claw-code-go/          # Main binary entry point
│   └── main.go                # CLI flag parsing, provider init, TUI launch
├── config/                    # Configuration files
├── internal/                  # All core application code (14 sub-packages)
│   ├── api/                   # Provider interface + Anthropic HTTP client
│   │   ├── providers/         # Multi-provider implementations
│   │   │   ├── anthropic/     # Anthropic Claude API
│   │   │   ├── bedrock/       # AWS Bedrock
│   │   │   ├── deepseek/      # DeepSeek web chat API (newest)
│   │   │   ├── foundry/       # Azure AI Foundry
│   │   │   ├── openai/        # OpenAI API
│   │   │   └── vertex/        # Google Vertex AI
│   │   ├── client.go          # Anthropic HTTP streaming client
│   │   ├── provider.go        # Provider/APIClient interfaces
│   │   ├── types.go           # ContentBlock, Tool, StreamEvent types
│   │   ├── toolcall.go        # Tool call parsing utilities
│   │   ├── xml_toolcalls.go   # XML-format tool call parsing (466 lines)
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
│   │   ├── manager.go         # Permission manager (Phase 5)
│   │   ├── mode.go            # Permission modes (default, accept-edits, bypass, plan)
│   │   └── rules.go           # Permission rules engine
│   ├── runtime/               # Core runtime: conversation loop, streaming, sessions
│   │   ├── conversation.go    # ConversationLoop (1156 lines) - THE CORE
│   │   ├── autonomous.go      # Autonomous task execution (RunTask)
│   │   ├── config.go          # Runtime configuration
│   │   ├── session.go         # Session management
│   │   ├── provider_factory.go # Multi-provider client factory
│   │   ├── compact.go         # Token compaction/context compression
│   │   ├── events.go          # Event types
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
│   ├── tui/                   # Terminal UI (Bubble Tea)
│   │   ├── model.go           # Main TUI model
│   │   ├── markdown.go        # Markdown rendering
│   │   ├── history.go         # Command history
│   │   ├── logo.go            # Logo display
│   │   ├── styles.go          # Lipgloss styles
│   │   └── theme.go           # Theme definitions
│   └── usage/                 # Token usage tracking
│       └── tracker.go         # Per-session usage tracker
├── mimo-cli/                  # External Node.js CLI tool (separate project)
│   ├── chat.js                # Main entry point (9,671 lines)
│   ├── src/                   # Source directory
│   └── package.json           # Node dependencies
├── go.mod                     # Go module definition
├── go.sum                     # Dependency checksums
├── README.md                  # Project documentation
└── CODEBASE_MAP.md            # This file
```

---

## Architecture

### Core Loop: ConversationLoop

The heart of the application is `ConversationLoop` (1156 lines in `internal/runtime/conversation.go`), which manages:
- **Agentic conversation** with tool-use loops
- **Streaming responses** from AI providers
- **Tool execution** and result injection back into the conversation
- **Context assembly** (environment, git, CLAUDE.md)
- **Token tracking** and compaction

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

### Provider Architecture

Six AI providers supported via a uniform `api.APIClient` interface:
| Provider | Package | Auth Method |
|----------|---------|-------------|
| Anthropic | `anthropic` | API Key / OAuth |
| OpenAI | `openai` | API Key |
| AWS Bedrock | `bedrock` | IAM |
| Vertex AI | `vertex` | ADC |
| Azure Foundry | `foundry` | Azure Identity |
| DeepSeek | `deepseek` | Token (web chat) |

Provider selection flows through `runtime.SelectProvider()` → `provider.NewClient()`.

### DeepSeek Provider (Newest Addition)

The DeepSeek provider is the most complex, using:
- **WASM solver** (`sha3_wasm_bg.wasm`, `wasm_solver.go`) for challenge-response auth
- **Web client** (`webclient.go`, 475 lines) for HTTP polling
- **Stream parsing** (`stream.go`, 465 lines) for SSE + polling hybrid
- **Models** (`models.go`, 218 lines) with rate limiting and TPD tracking
- **Provider tests** (`provider_test.go`, 732 lines)

### Permission System

Four permission modes:
- **default** - Ask user for each file write/bash execution
- **accept-edits** - Auto-accept file edits, ask for bash
- **bypass** - Auto-accept everything (used by autonomous mode)
- **plan** - Plan mode (no execution)

### Autonomous Mode

`RunTask()` in `autonomous.go` enables fully autonomous execution:
- Forces `bypass` permission mode
- Strips `ask_user` from tool list
- Uses LLM-judge for DONE/CONTINUE decisions
- Respects `MaxTurns` limit

### Context Assembly

The `context/assembler.go` injects environment context into the system prompt:
- Working directory and OS info
- Git branch, status, and recent commits
- CLAUDE.md / project memory files
- Date/time

---

## Data Flow

```
User Input (TUI)
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
    ├─► Text delta → emit to TUI
    │
    └─► Tool use → Execute tool → inject tool_result → loop
```

---

## Key Files by Importance

| File | Lines | Role |
|------|-------|------|
| `internal/runtime/conversation.go` | 1156 | Core agent loop |
| `internal/api/providers/deepseek/provider.go` | 608 | DeepSeek provider (newest) |
| `internal/api/providers/deepseek/stream.go` | 465 | DeepSeek stream handling |
| `internal/api/providers/deepseek/webclient.go` | 475 | DeepSeek HTTP client |
| `internal/api/xml_toolcalls.go` | 466 | XML tool call parsing |
| `internal/api/providers/deepseek/provider_test.go` | 732 | DeepSeek tests |
| `internal/api/client.go` | 231 | Anthropic HTTP client |
| `internal/tools/web_search.go` | 226 | Web search implementation |
| `cmd/claw-code-go/main.go` | 211 | CLI entry point |
| `internal/runtime/autonomous.go` | 248 | Autonomous task runner |

---

## Build & Dependencies

- **Go version:** 1.24.2
- **Core dependency:** Bubble Tea (charmbracelet) for TUI
- **WASM runtime:** wazero (tetratelabs) for DeepSeek challenge solver
- **Markdown:** glamour (charmbracelet) for TUI rendering
- **No external CGO or system dependencies**

---

## Current Git Status

- **Branch:** master
- **Last commits:** 
  - `535bb60` - feat: integrate deepseek provider
  - `23c653b` - forked from daolmedo/claw-code-go
- **Modified files:** 18 files across deepseek provider, commands, permissions, and runtime
- **Untracked:** 22 files (including `main` binary, `.codegraph/`, `CODEBASE_MAP.md`)
