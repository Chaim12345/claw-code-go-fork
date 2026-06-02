# TUI Features Documentation

## Overview

The claw-code-go TUI provides a rich, interactive terminal interface with advanced features inspired by modern AI coding assistants. This document covers all available features, keyboard shortcuts, and usage patterns.


---

## 📏 Message Size Management

### Automatic Message Truncation

The system automatically handles oversized messages to prevent API rejections:

- **Per-Message Limit**: 150,000 characters (~37,500 tokens)
- **Automatic Truncation**: Messages exceeding the limit are automatically truncated
- **User Notification**: TUI displays a warning when truncation occurs
- **Server-Side Preservation**: Full conversation history is retained on the server

### How It Works

1. **Pre-Send Validation**: Each message is validated before sending to the API
2. **Smart Truncation**: Content is truncated at line boundaries when possible for readability
3. **Clear Markers**: Truncated messages include a marker explaining the truncation
4. **Warning Events**: TUI receives `TurnEventWarn` events to notify users

### Technical Details

- **Implementation**: `internal/runtime/message_limits.go`
- **Validation Function**: `ValidateAndTruncateMessage()`
- **Integration Points**: Applied in both `SendMessage()` and `SendMessageStreaming()`
- **Event Type**: New `TurnEventWarn` added to runtime events
- **Conservative Limits**: Set below provider caps (DeepSeek ~160K, Anthropic ~200K)

### User Experience

When a message is truncated:
- **CLI Mode**: Warning printed to stderr
- **TUI Mode**: Yellow warning message displayed in conversation
- **Marker Text**: Clear explanation that content was truncated
- **No Data Loss**: Full message preserved in server-side history


---

## 🎨 Pretty Tool Call Rendering

Tool calls are now displayed in beautiful bordered boxes with icons, badges, and syntax highlighting.

### Collapsed View (Default)
```
  ⚡ [SHELL] bash ls -la ✓
  📝 [FILE] write_file Created config.json ✓
  🌐 [WEB] web_search Found 5 results ✓
```

### Expanded View (Ctrl+T)
```
╭──────────────────────────────────────────────────────────────────────────────╮
│ ⚡ bash                                                                       │
├──────────────────────────────────────────────────────────────────────────────┤
│ Input:                                                                        │
│   command: "ls -la"                                                          │
│   timeout: 30                                                                │
├──────────────────────────────────────────────────────────────────────────────┤
│ Result:                                                                       │
│   total 48                                                                   │
│   drwxr-xr-x  12 user  staff   384 Jun  2 08:20 .                          │
│   drwxr-xr-x   5 user  staff   160 Jun  1 10:15 ..                         │
╰──────────────────────────────────────────────────────────────────────────────╯
```

### Tool Icons & Badges

| Tool | Icon | Badge | Color |
|------|------|-------|-------|
| bash | ⚡ | [SHELL] | Orange |
| read_file | 📖 | [FILE] | Blue |
| write_file | ✏️ | [FILE] | Blue |
| file_edit | 📝 | [FILE] | Blue |
| web_fetch | 🌐 | [WEB] | Green |
| web_search | 🔍 | [WEB] | Green |
| ask_user | ❓ | [USER] | Purple |
| todo_write | ✅ | [TOOL] | Gray |

---

## ⌨️ Keyboard Shortcuts

### Global Shortcuts
| Key | Action | Description |
|-----|--------|-------------|
| **Ctrl+P** | Command Palette | Fuzzy search all commands |
| **Ctrl+T** | Toggle Tool Card | Expand/collapse most recent tool |
| **Shift+Tab** | Cycle Permission Mode | default → accept-edits → bypass → plan |
| **Ctrl+C** | Quit | Exit the application |
| **Enter** | Send Message | Submit current input |
| **Ctrl+J** | Insert Newline | Multi-line input mode |
| **↑ / ↓** | History Navigation | Browse previous messages (single-line) |
| **PgUp / PgDn** | Scroll Viewport | Navigate conversation history |

### Input Shortcuts
| Key | Action | Description |
|-----|--------|-------------|
| **@** | File Mention | Trigger file autocomplete |
| **Tab** | Accept Mention | Insert selected file path |
| **Esc** | Cancel | Close overlay/cancel action |

---

## 🎯 Command Palette (Ctrl+P)

Fuzzy-searchable overlay for all commands and actions.

### Usage
1. Press **Ctrl+P** to open
2. Type to filter commands (fuzzy matching)
3. Use **↑/↓** to navigate
4. Press **Enter** to execute
5. Press **Esc** to cancel

### Available Commands
- **Switch model** - Open model picker
- **Change theme** - Toggle dark/light theme
- **Clear session** - Wipe conversation history
- **Save session** - Save to disk
- **Load session** - Browse saved sessions
- **Show status** - Display current config
- **Show cost** - Token usage summary
- **Toggle permission mode** - Cycle modes
- **Toggle todo panel** - Show/hide todos
- **Show help** - Open help panel
- **Exit** - Quit application

### Multi-Token Search
Search with multiple words: `"show config"` matches "Show config"

---

## 📁 @ File Mentions

Autocomplete file paths by typing `@` in the input.

### Usage
```
You: Can you review @src/main.go and @internal/tui/model.go?
```

### Features
- **Fuzzy search** - Type partial names
- **Respects .gitignore** - Skips ignored files
- **Depth scoring** - Prefers shallower files
- **Tab/Enter** - Insert selected path
- **Esc** - Cancel autocomplete

### Example
```
Type: @mod
Suggestions:
  ▶ internal/tui/model.go
    go.mod
    internal/runtime/model.go
```

---

## 📊 Inline Diff Viewer

Visual file change preview with color-coded +/- lines.

### Features
- **LCS-based diff** - Line-level comparison
- **Color coding** - Green (+), Red (-), Gray (context)
- **Automatic** - Shows for write_file/file_edit tools
- **Compact stats** - "+12 -3" summary in collapsed view

### Example
```
╭──────────────────────────────────────────────────────────────────────────────╮
│ 📝 file_edit                                                                 │
├──────────────────────────────────────────────────────────────────────────────┤
│     src/config.go                                                            │
│     func LoadConfig() {                                                      │
│   -     port := 8080                                                         │
│   +     port := 3000                                                         │
│       log.Printf("Starting on port %d", port)                               │
│     }                                                                        │
╰──────────────────────────────────────────────────────────────────────────────╯
```

---

## 📋 Todo Panel

Task list sidebar that auto-shows when `todo_write` runs.

### Features
- **Auto-display** - Shows when todos are created
- **Priority sorting** - in_progress → pending → done
- **Toggle** - `/todo` or Ctrl+P → "Toggle todo panel"
- **Sidebar layout** - On wide terminals (>120 cols)

### Status Icons
- `[-]` In Progress (yellow)
- `[ ]` Pending (white)
- `[x]` Done (green, dimmed)

### Example
```
┌─ Todos ─────────────┐
│ [-] Implement auth  │
│ [ ] Write tests     │
│ [x] Setup database  │
└─────────────────────┘
```

---

## 🔐 Permission Modes

Control how the AI interacts with your system.

### Modes
1. **default** - Ask for each file write/bash command
2. **accept-edits** - Auto-accept file edits, ask for bash
3. **bypass** - Auto-accept everything (autonomous mode)
4. **plan** - Planning mode, no execution

### Toggle
- **Shift+Tab** - Cycle through modes
- **Status bar** - Shows current mode badge
- **Command palette** - "Toggle permission mode"

### Status Bar Display
```
[default] | [accept-edits] | [bypass] | [plan]
```

---

## 💬 Session Management

Save, load, and browse conversation sessions.

### Commands
- `/session save [name]` - Save current session
- `/session load <name>` - Load saved session
- `/session list` - List all sessions
- **Ctrl+P** → "Load session" - Visual picker

### Session Picker
```
┌─ Sessions ──────────────────────────────────────────────┐
│ ▶ project-refactor    2026-06-02 08:15    45 msgs      │
│   bug-investigation   2026-06-01 14:30    23 msgs      │
│   feature-planning    2026-05-31 09:00    67 msgs      │
└─────────────────────────────────────────────────────────┘
```

---

## 🎨 Themes

Switch between dark and light color schemes.

### Commands
- `/theme dark` - Dark theme (default)
- `/theme light` - Light theme
- **Ctrl+P** → "Change theme" - Toggle

### Theme Features
- **Markdown rendering** - Matches theme
- **Syntax highlighting** - Theme-aware colors
- **Tool cards** - Consistent styling

---

## 🔍 Search & Navigation

### Conversation History
- **PgUp/PgDn** - Scroll viewport
- **Home/End** - Jump to top/bottom
- **Mouse wheel** - Scroll (if supported)

### Input History
- **↑** - Previous message (single-line mode)
- **↓** - Next message
- **Ctrl+J** - Multi-line mode (disables history nav)

---

## 📝 Slash Commands

Quick commands for common actions.

### Core Commands
| Command | Description |
|---------|-------------|
| `/help` | Show help panel |
| `/model` | Open model picker |
| `/clear` | Clear conversation |
| `/status` | Show current config |
| `/cost` | Token usage summary |
| `/config [key] [value]` | View/set config |
| `/init` | Create .claude/settings.json |
| `/theme <dark\|light>` | Change theme |
| `/exit` or `/quit` | Exit application |

### Session Commands
| Command | Description |
|---------|-------------|
| `/session list` | List saved sessions |
| `/session save [name]` | Save current session |
| `/session load <name>` | Load saved session |
| `/compact` | Compact session context |

### Auth Commands
| Command | Description |
|---------|-------------|
| `/login` | Multi-provider login flow |
| `/auth login` | OAuth login (legacy) |
| `/auth logout` | Clear stored tokens |
| `/auth status` | Show auth status |

---

## 🎯 Advanced Features

### Token Budget Management
- **Status bar** - Shows in/out tokens
- **/cost** - Detailed usage breakdown
- **Auto-compaction** - Prevents context overflow

### Markdown Rendering
- **Glamour integration** - Beautiful formatting
- **Code blocks** - Syntax highlighting
- **Tables** - Proper alignment
- **Lists** - Nested support

### Error Handling
- **Graceful degradation** - Falls back to plain text
- **Error messages** - Color-coded in red
- **Warnings** - Yellow highlights

---

## 🐛 Troubleshooting

### Terminal Too Small
- **Minimum size**: 10 lines × 40 columns
- **Recommended**: 24 lines × 80 columns
- **Optimal**: 40 lines × 120 columns (for todo sidebar)

### Rendering Issues
- **Colors wrong?** - Check terminal supports 256 colors
- **Borders broken?** - Ensure UTF-8 encoding
- **Slow rendering?** - Reduce viewport size with PgUp/PgDn

### Performance
- **Large sessions** - Use `/compact` to reduce context
- **Many tool calls** - Collapsed view is faster
- **Long outputs** - Results auto-truncate at 4KB

---

## 📚 Tips & Tricks

1. **Quick model switch** - Ctrl+P → type "opus" → Enter
2. **Batch file review** - Use @ mentions for multiple files
3. **Safe experimentation** - Use "plan" mode to preview
4. **Session snapshots** - Save before risky operations
5. **Tool inspection** - Ctrl+T to see full tool I/O
6. **Fuzzy everything** - Command palette and @ mentions support fuzzy search
7. **Multi-line prompts** - Ctrl+J for complex instructions
8. **History recall** - ↑ to reuse previous prompts

---

## 🚀 What's New

### Latest Features (v1.0.4+)
- ✨ **Pretty tool cards** with icons, badges, and borders
- 🎨 **Syntax highlighting** for JSON in tool inputs
- 📊 **Enhanced diff viewer** with better formatting
- ⚡ **Performance improvements** in markdown rendering
- 🔒 **Thread-safe** markdown renderer (race condition fixed)
- 🛡️ **Panic protection** for tiny terminals
- 🧹 **State cleanup** fixes for tool card boundaries

### From go-claw Integration
- 🎯 Command palette (Ctrl+P)
- 📁 @ file mentions with autocomplete
- 📋 Todo panel sidebar
- 🔄 Permission mode toggle (Shift+Tab)
- 📊 Session picker overlay
- 🎨 Enhanced markdown rendering
- 🔧 Tool card expansion (Ctrl+T)

---

## 📖 Further Reading

- **Configuration**: See `.claude/settings.json` for project config
- **Providers**: Supports Anthropic, OpenAI, Bedrock, Vertex, Foundry, DeepSeek
- **Tools**: 10+ built-in tools (bash, files, web, etc.)
- **Testing**: Run `go test ./internal/tui/...` for TUI tests

---

**Version**: 1.0.4+  
**Last Updated**: 2026-06-02  
**Maintainer**: claw-code-go team
