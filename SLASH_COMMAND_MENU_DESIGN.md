# Interactive Slash Command Menu Design

## Current State Analysis

### How slash commands work today

1. User types `/something` in the textarea
2. Presses Enter
3. `handleSubmit()` checks `strings.HasPrefix(text, "/")`
4. Routes to `handleSlashCommand()` which does a big `switch` on `parts[0]`
5. Dispatches to specific handler or prints "Unknown command"

**Problems:**
- No discoverability — user must know exact command names
- No filtering/fuzzy matching — typo = "Unknown command"
- No inline feedback while typing — must submit to see if valid
- Existing palette (`Ctrl+P`) is a separate overlay, not inline

### Existing similar implementations in codebase

#### 1. `@mention` autocomplete (`mention.go`)
- **Triggers on `@` character in textarea**
- Intercepts keystrokes in `handleKey` default case: calls `m.mention.update(m.textarea.Value())`
- Switches to `stateMention` when active
- Handles Tab/Enter for selection, Esc to cancel
- Renders a floating popup at bottom of screen
- **This is the closest analog — the `/` menu should work EXACTLY like this**

#### 2. Command palette (`palette.go`)
- Opens on `Ctrl+P`, full-screen overlay
- Has its own search box (not inline in textarea)
- Fuzzy filtered with scoring
- Separate state `statePalette`

### TUI architecture (bubbletea)

- Single `Model` struct with `state appState` field
- `Update(msg tea.Msg)` dispatches to `handleKey(msg tea.KeyMsg)`
- `handleKey` has a big switch on `m.state` for modal states, then on `msg.Type` for `stateInput`
- Textarea gets keystrokes via `m.textarea.Update(msg)` in the default branch
- `@mention` detection happens AFTER textarea.Update, by checking `m.mention.update(m.textarea.Value())`

## Design: Interactive `/` Command Menu

### Behavior (like OpenCode)

1. User types `/` in the textarea
2. A popup appears **above the input area** showing all slash commands
3. As user continues typing, the list filters in real-time
4. User can navigate with ↑/↓ arrows
5. **Enter** selects the highlighted command and either:
   - Executes it immediately (if it takes no args like `/help`, `/clear`)
   - Fills the textarea with the command prefix and lets user continue typing args
6. **Tab** autocompletes the command name
7. **Esc** closes the popup (but keeps the `/` text in textarea)
8. **Backspace** to delete the `/` closes the popup and returns to normal input

### Implementation plan

#### Step 1: Add `slashMenu` struct and state

Create new file: `internal/tui/slash_menu.go`

```go
type slashMenuItem struct {
    command     string // e.g., "/help"
    description string // e.g., "Show available commands"
    args        string // e.g., "[key] [value]" — hint for arguments
}

type slashMenu struct {
    active   bool
    cursor   int
    query    string   // text after the "/"
    items    []slashMenuItem
    filtered []int
}
```

#### Step 2: Add `stateSlashMenu` to appState

In `model.go`:
```go
const (
    // ...existing states...
    stateSlashMenu  // slash command autocomplete popup
)
```

Add field to `Model`:
```go
slashMenu *slashMenu
```

#### Step 3: Detect `/` in handleKey default case

In `handleKey`, in the `default` branch, after `m.textarea.Update(msg)`:

```go
// Check for @-mention AND /-command
if m.mention.update(m.textarea.Value()) {
    m.state = stateMention
} else if m.slashMenu.update(m.textarea.Value()) {
    m.state = stateSlashMenu
} else if m.state == stateMention || m.state == stateSlashMenu {
    m.state = stateInput
}
```

#### Step 4: Handle keys in slashMenu state

```go
case stateSlashMenu:
    return m.handleSlashMenuKey(msg)
```

`handleSlashMenuKey` needs:
- **Esc** → close menu, stay in input (keep text)
- **Enter** → select command, execute or fill textarea
- **Tab** → autocomplete command name in textarea
- **↑/↓** → move cursor in filtered list
- **Backspace** → pass to textarea; if no more `/`, close menu
- **All other keys** → pass to textarea for typing filter query

#### Step 5: Render slash menu in View()

In `View()`, add case `stateSlashMenu` rendering. The popup should appear **above the input area**, positioned at the bottom of the viewport. Use the existing `paletteBoxStyle` and similar layout to `mention.view()`.

#### Step 6: Command list definition

```go
func allSlashCommands() []slashMenuItem {
    return []slashMenuItem{
        {"/help", "Show available commands", ""},
        {"/model", "Change the active model (picker)", ""},
        {"/login", "Multi-provider login flow", ""},
        {"/clear", "Clear conversation history", ""},
        {"/theme", "Switch TUI color theme", "dark|light"},
        {"/status", "Show model/provider/session info", ""},
        {"/cost", "Show token usage this session", ""},
        {"/config", "Show or set config values", "[key] [value]"},
        {"/session", "Manage sessions", "list|save|load [name]"},
        {"/sessions", "Browse saved sessions (picker)", ""},
        {"/todo", "Toggle todo list sidebar", ""},
        {"/init", "Create .claude/settings.json", ""},
        {"/exit", "Exit (session auto-saved)", ""},
        {"/quit", "Exit (session auto-saved)", ""},
    }
}
```

### Filtering logic

Use the same fuzzy-match scoring as `palette.refilter()`:
- Case-insensitive substring matching
- Multi-token queries (type "se li" matches "/session list")
- Prefix bonus for scoring
- Show description next to each item

### View layout

```
┌──────────────────────────────────┐
│  / Slash Commands                │
│                                  │
│  ▶ /session   Manage sessions    │
│    /sessions  Browse sessions    │
│    /status    Show status info   │
│    /help      Show help          │
│                                  │
│  ↑↓ select  Tab complete  Esc    │
└──────────────────────────────────┘
> /ses█
```

### Key decisions

| Question | Answer |
|----------|--------|
| Should `/` still work without the menu? | Yes — if user types `/help` and hits Enter without using the menu, existing behavior is preserved |
| Execute on Enter or fill textarea? | **Fill textarea** with command, let user add args. For no-arg commands, execute immediately |
| Close on selection? | Yes — fill textarea, close menu, keep cursor position after inserted text |
| Match `@mention` UX? | Yes — same placement, same styles, same keyboard shortcuts |
| Should typing a space close the menu? | No — some commands have args (e.g., `/config model claude-sonnet`) — let user keep typing |

### Files to modify/create

1. **NEW** `internal/tui/slash_menu.go` — the slashMenu struct and logic
2. **MODIFY** `internal/tui/model.go`:
   - Add `stateSlashMenu` to appState
   - Add `slashMenu` field to Model
   - Initialize in NewModel
   - Add detection in handleKey default case
   - Add `handleSlashMenuKey` method
   - Add dispatch in handleKey state switch
   - Add rendering in View()
3. **MODIFY** `internal/tui/debug/state_machine.go` — add StateSlashMenu

### Bonus: Command argument hints

When a command like `/config` is selected, show a subtle hint below the list:
```
/config [key] [value]
```
This helps discoverability of subcommands and arguments.
