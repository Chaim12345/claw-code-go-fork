# TUI Gap Analysis — claw-code-go vs Modern Agentic TUIs

> Comprehensive gap map comparing the current state of the claw-code-go TUI to
> modern agentic coding TUIs (Claude Code TS, OpenCode, Gemini CLI, aider,
> Codex CLI, Continue, Cursor CLI, etc.) as of 2025-2026.
>
> Source of truth: deep code exploration of `internal/tui/` (30+ files, 10K+
> LOC) cross-referenced with public TUI feature conventions.

---

## 1. Inventory snapshot — what is **PRESENT**

| Area | Implementation |
|---|---|
| Bubble Tea runtime + AltScreen | `cmd/claw-code-go/main.go:203-205` |
| Mouse (wheel + click, cell-motion) | `main.go:205`, `model.go:345,2340-2390` |
| Themes (dark / light / highcontrast / reducedmotion) | `theme.go:29-112` |
| Focus-border + unfocus-border tokens | `theme.go:24-25,46-47,68-69,90-91,110-111` |
| Status badges / toast manager | `status_badges.go` |
| Markdown rendering (glamour) | `markdown.go`, `streaming_renderer.go:35-42` |
| Streaming renderer with throttled re-render | `streaming_renderer.go:155-180` |
| Multi-line input (Ctrl+J newline) | `model.go:771-776` |
| Input history (Up/Down) | `model.go:778-800`, `history.go` |
| Slash command menu (fuzzy) | `slash_menu.go` |
| Command palette (Ctrl+P) | `palette.go`, `model.go:658-662` |
| Quick actions (Ctrl+K) | `quick_actions.go`, `model.go:693-697` |
| Conversation search (Ctrl+F) | `conversation_search.go`, `model.go:699-703` |
| Command-history search (Ctrl+R) | `history_search.go`, `model.go:686-690` |
| Session picker (Ctrl+S via /sessions) | `session_picker.go` |
| Session tags + filter | `session_picker.go:30,57,113,129,159,174-180` |
| Tag display in picker table | `session_picker.go:174-180` |
| Tool card expand/collapse (last, all-Ctrl+O) | `model.go:731-742, 1711-1754` |
| Diff approval view | `diff_approval.go` |
| Diff rendering (line-by-line) | `diff.go`, `diff_approval.go:211-235` |
| Permission prompt (y/n/always) | `model.go:548, 2300+` |
| Ask-user question UI | `model.go:555-563, 1439-1525` |
| Todo panel toggle | `model.go:100, 1647`, `todo_panel.go` |
| Background task panel (Ctrl+B) | `background_tasks.go`, `model.go:676-682` |
| Undo/Redo file changes (Ctrl+Z/Y) | `model.go:705-724`, `undo_redo.go` |
| Clipboard copy (system + OSC52) | `clipboard.go`, `quick_actions.go:79-100` |
| Copy code block from viewport | `model.go:1756+` |
| Spinner (dots) | `model.go:263-264, 2850+` |
| "Waiting for browser login" screen | `model.go:2850+` |
| Status bar (mode + tokens + session + Chars + render) | `model.go:2594-2632` |
| Theme switching via `/theme` | `model.go:917-940` |
| Login flow (provider → method → API key / OAuth) | `model.go:91-95, 1237-1455` |
| Resume session | runtime/session.go, TUI session picker |
| Auto-save session | runtime/session.go |
| Debug panel (Ctrl+D) | `debug_handlers.go`, `model.go:664-672` |
| Help overlay (F1) | `model.go:654-656`, `model.go:2730+` |
| Permission modes (default/accept-edits/bypass/plan) | `permissions.go`, `model.go:310-314` |
| Mode indicator in status bar | `model.go:2604-2608` |
| Scrollbar/position indicator | `model.go:2398+` |
| Mouse click in input → focus textarea | `model.go:2355-2361` |
| Read-only / too-small terminal guard | `model.go:2480+` |
| All test files in `tui_test.go` for the new features | PASS |
| Reduced motion / high contrast themes | `theme.go` |

**Implementation footprint:** 30+ files, 10,247 LOC, 6 view-layer files, 7
overlay states, 1 command registry, 1 comprehensive model.

---

## 2. Gaps — features **MISSING** or **PARTIAL**

### 2.1 High-priority (table stakes in 2025-2026)

| # | Feature | Status | Notes |
|---|---|---|---|
| G1 | **@-mention file completion** | PARTIAL | `mention.go` exists; wired in `model.go:821-822`, but I should verify it opens with `@` reliably. |
| G2 | **@-mention accepts paths outside cwd** | MISSING | Only current working directory. |
| G3 | **@-mention accepts directories** | MISSING | Files only. |
| G4 | **@-mention shows file size, modified time** | MISSING | Just path display. |
| G5 | **@-mention supports fuzzy + path prefix** | PARTIAL | Only substring containment. |
| G6 | **Auto-suggest / ghost-text input** | MISSING | No fish-shell-style inline completion from history. |
| G7 | **Context-window % indicator in status bar** | MISSING | Only "Tokens: in/out" — no `45% / 200k` visual. |
| G8 | **Plan mode display (read-only agent)** | PARTIAL | Mode is in `permissions.go`/status, no dedicated UI/badge for "PLAN active". |
| G9 | **Plan display pane (plan document)** | MISSING | No `statePlanViewer`. |
| G10 | **Sub-agent / task delegation display** | MISSING | No `stateSubAgent` or sub-agent panel. |
| G11 | **Custom user-defined slash commands** | MISSING | Only hard-coded `commandRegistry`. |
| G12 | **Image / picture support in chat** | MISSING | Vision model reference but no `ImageBlock` rendering. |
| G13 | **Image paste from clipboard** | MISSING | No image paste handler. |
| G14 | **Image preview (chafa/kitty/sixel/iTerm)** | MISSING | No image protocol emitter. |
| G15 | **Markdown export to file** | PARTIAL | Implemented in tests (`TestExportSessionAsMarkdown*`) but no UI command binding. |
| G16 | **Vim mode in input** | MISSING | No keymap mode flag. |
| G17 | **Emacs mode toggle** | MISSING | Default only. |
| G18 | **MCP server indicator / status** | MISSING | No display of connected MCP servers. |
| G19 | **MCP tool discovery panel** | MISSING | No list of MCP tools. |
| G20 | **Git branch + dirty indicator in status bar** | MISSING | Status bar has no git context. |
| G21 | **Working directory display in status bar** | MISSING | Status bar shows session/tokens but not cwd. |
| G22 | **Config hot-reload** | MISSING | No watcher. |
| G23 | **Input placeholder / hint text** | MISSING | No "Type a message…" placeholder. |
| G24 | **Input keybinding hints in input area** | MISSING | No `(shift+enter for newline)` hint. |
| G25 | **Display char/token-count threshold warning** | MISSING | No "approaching limit" warning. |
| G26 | **Auto-suggest based on conversation** | MISSING | No `/agent suggest "fix"`-style completion. |
| G27 | **Plan mode quick toggle (`/plan`)** | PARTIAL | Exists in `permissions.go` mode list but not exposed as a single command. |
| G28 | **Queuing messages during stream** | MISSING | Cannot type & submit while busy; `stateBusy` only handles Ctrl+C. |
| G29 | **Streaming-interrupt button (Esc) during stream** | PARTIAL | `stateBusy` ignores everything except Ctrl+C — Esc does nothing. |
| G30 | **Stream-replace final text without re-render flicker** | PARTIAL | `streaming_renderer.go` throttles, but no "live diff" or "typewriter" toggle. |

### 2.2 UX polish (medium priority)

| # | Feature | Status | Notes |
|---|---|---|---|
| G31 | **Line numbers in code blocks** | MISSING | `ta.ShowLineNumbers = false` (model.go:255); code blocks also lack line numbers. |
| G32 | **Copy code-block shortcut visible in tool card** | PARTIAL | Code-block copy exists (model.go:1756) but no keybind hint. |
| G33 | **Tab-cycle between tool cards** | MISSING | No "next tool card" key. |
| G34 | **Search forward in conversation (`/`)** | PARTIAL | Only "next match" not "open search bar". |
| G35 | **Inline image preview in streaming** | MISSING | — |
| G36 | **Per-tool card copy-result shortcut** | MISSING | Quick actions has it (model.go:79-100) but not a single keystroke. |
| G37 | **Permission "always for session" prompt** | PARTIAL | Permission modes exist; per-tool "always" not exposed in UI. |
| G38 | **Diff approval: arrow keys to navigate hunks** | PARTIAL | Diff approval exists; need to check hunk navigation. |
| G39 | **Edit-and-resend last message** | MISSING | No edit-last-user-message flow. |
| G40 | **Branching sessions / fork from a message** | MISSING | — |
| G41 | **Session rename (label)** | PARTIAL | `Label` field exists on `Session` struct but no UI to set it. |
| G42 | **Session notes / annotations** | MISSING | — |
| G43 | **Status bar: elapsed time** | MISSING | No per-turn or session timer. |
| G44 | **Status bar: background task count** | PARTIAL | Panel exists (Ctrl+B) but not a passive indicator. |
| G45 | **Settings panel (full config UI)** | MISSING | `/config` exists but only text-based. |
| G46 | **MCP install UI** | MISSING | No `/mcp add <name> <command>` wizard. |
| G47 | **Working directory picker (cd from TUI)** | MISSING | — |
| G48 | **File-tree side panel** | MISSING | — |
| G49 | **Read-only mode / live preview toggle** | MISSING | Preview toggle exists (Ctrl+L) but not full RO mode. |
| G50 | **Window/scroll position in conversation** | PARTIAL | Scrollbar exists; no "Line 45 of 230" indicator. |
| G51 | **Theme per-session override** | MISSING | Global only. |
| G52 | **Color picker / theme editor** | MISSING | — |
| G53 | **Notification sound (bell) on completion** | MISSING | No `\a` on stream end. |
| G54 | **Status dot for streaming vs idle** | PARTIAL | Spinner exists; no idle dot. |
| G55 | **Pause/resume a long task** | MISSING | Only cancel. |
| G56 | **Session export to JSON** | MISSING | Markdown only. |
| G57 | **Session share-link** | MISSING | — |
| G58 | **Clipboard OSC52 fallback (SSH)** | PARTIAL | `clipboard.go` uses external tools; OSC52 path not exercised. |
| G59 | **Hyperlinks (OSC 8) for paths/URLs** | MISSING | No terminal hyperlink emitter. |
| G60 | **Tabbed panels (multiple chats)** | MISSING | Single chat only. |

### 2.3 Input/Editor (medium priority)

| # | Feature | Status | Notes |
|---|---|---|---|
| G61 | **Auto-pair brackets/quotes** | MISSING | textarea default. |
| G62 | **Tab-key to insert literal tab character** | PARTIAL | `KeyTab` falls through to textarea; for actual tab insertion in non-shell, no explicit. |
| G63 | **Smart Home/End (jump over leading whitespace)** | MISSING | — |
| G64 | **Selection / copy in textarea** | PARTIAL | Mouse works, keyboard selection not bound. |
| G65 | **Undo in textarea (Ctrl+Z)** | MISSING | `KeyCtrlZ` is bound to "undo file change" — input has no undo. |
| G66 | **Redo in textarea (Ctrl+Y)** | MISSING | Same — file-change redo. |
| G67 | **Word-jump (Ctrl+Left/Right)** | MISSING | — |
| G68 | **Select-all (Ctrl+A)** | MISSING | — |
| G69 | **Cut-line (Ctrl+K)** | MISSING | Conflicts with quick-actions — needs disambiguation by state. |
| G70 | **Paste stripping (paste handler)** | PARTIAL | Bracketed paste is the bubbletea default; no custom handling. |
| G71 | **Drag-and-drop file path in input** | MISSING | — |
| G72 | **Syntax-aware indent / dedent** | MISSING | — |

### 2.4 Streaming / display (medium)

| # | Feature | Status | Notes |
|---|---|---|---|
| G73 | **Inline diff in streaming output** | PARTIAL | Tool cards have diffs but not in main text. |
| G74 | **Citations / sources footer on message** | MISSING | — |
| G75 | **Footnotes / footnote style refs** | MISSING | — |
| G76 | **Tables in markdown** | PARTIAL | glamour supports; verify rendering. |
| G77 | **Task list in markdown (GFM checkboxes)** | PARTIAL | glamour supports; verify. |
| G78 | **Mermaid / diagram rendering** | MISSING | — |
| G79 | **Math / LaTeX rendering** | MISSING | — |
| G80 | **Collapsible section headers (`<details>`)** | MISSING | — |
| G81 | **Quote-block collapse when long** | MISSING | — |
| G82 | **Sticky input while streaming** | PARTIAL | Input is always visible — verify behavior at terminal bottom edge. |
| G83 | **Auto-scroll to bottom on new content** | PARTIAL | Default viewport behavior; verify pin-to-bottom toggle. |
| G84 | **Pin-to-bottom toggle** | MISSING | No key. |
| G85 | **Streaming rate (tokens/sec) display** | MISSING | — |
| G86 | **Time-to-first-token indicator** | MISSING | — |

### 2.5 Accessibility & a11y (medium)

| # | Feature | Status | Notes |
|---|---|---|---|
| G87 | **Screen reader / NVDA support** | N/A | TUI limitation, but bracket-style metadata helps. |
| G88 | **High-contrast / color-blind palettes** | PARTIAL | HC theme exists; not customizable. |
| G89 | **Reduced motion** | PRESENT | `reducedmotion` theme. |
| G90 | **No-color / ASCII-only mode** | MISSING | — |
| G91 | **Adjustable font size** | N/A | Terminal only. |
| G92 | **Width-adaptive layout (<60 cols)** | PARTIAL | Too-small guard exists; no compact layout. |
| G93 | **Width-adaptive layout (>200 cols)** | MISSING | Doesn't use wide layout. |
| G94 | **Touch / mobile-friendly** | MISSING | TUI. |
| G95 | **Help overlay describes all keys** | PARTIAL | `model.go:2730+` is a static list; not context-aware. |
| G96 | **Read aloud / TTS** | N/A | TUI. |

### 2.6 Sub-agent / advanced agent features (high-priority for parity with claude-code-ts)

| # | Feature | Status | Notes |
|---|---|---|---|
| G97 | **Sub-agent panel** | MISSING | No display for spawned sub-agents. |
| G98 | **Active agents in conversation** | MISSING | — |
| G99 | **Per-agent progress** | MISSING | — |
| G100 | **Cancel a single sub-agent** | MISSING | — |
| G101 | **Task list (todo_write tool) — in input bar** | PARTIAL | Todo panel exists (Ctrl+T) but not as inline pills. |
| G102 | **Worktree isolation display** | MISSING | — |
| G103 | **Sandbox indicator (running inside sandbox?)** | MISSING | — |
| G104 | **Yolo-mode indicator** | PARTIAL | Permission mode in status; "yolo" not visible. |
| G105 | **Plan-mode input bar lock** | MISSING | — |
| G106 | **Plan-mode auto-display of plan** | MISSING | — |
| G107 | **Compaction pending toast** | PARTIAL | `/compact` exists; no "compacting now" status. |
| G108 | **Hook display (PreToolUse, PostToolUse)** | MISSING | — |
| G109 | **MCP server list panel** | MISSING | — |
| G110 | **Plugin / extension list** | MISSING | — |
| G111 | **Skills list (slash commands) panel** | MISSING | — |
| G112 | **Memory directory browser** | MISSING | — |
| G113 | **Context window breakdown (system/tools/messages)** | MISSING | — |

### 2.7 Performance / reliability (low)

| # | Feature | Status | Notes |
|---|---|---|---|
| G114 | **Throttled re-render** | PRESENT | `streaming_renderer.go:155-180`. |
| G115 | **Render timing in status bar** | PARTIAL | Only in debug mode. |
| G116 | **Lazy glamour renderer init** | PARTIAL | `markdown.go:30-40`. |
| G117 | **Memory pressure indicator** | MISSING | — |
| G118 | **Backpressure during very long output** | PARTIAL | `toolCards` capped at 50 (model.go:433-435). |
| G119 | **Resize debounce** | MISSING | — |
| G120 | **Crash recovery / auto-resume on restart** | PARTIAL | Sessions saved; no prompt to resume on startup. |

---

## 3. Quick wins — what to fill first (ordered by ROI)

Ranked by **user value × implementation ease**.

| Priority | Gap | Why |
|---|---|---|
| **P0** | G7: Context window % in status bar | Users constantly ask "how much room do I have?" — 30 LOC fix. |
| **P0** | G20: Git branch in status bar | Universal expectation from modern TUIs. |
| **P0** | G21: Working directory in status bar | Pairs with G20. |
| **P0** | G6: Auto-suggest / ghost-text input | Big UX win, ~150 LOC with bubbles `textinput`. |
| **P0** | G65-G69: Textarea keybindings (Ctrl+A, Ctrl+Z, Ctrl+←/→) | Currently hijacked for file changes — needs mode-aware dispatch. |
| **P0** | G11: Custom user slash commands | Open `~/.claw-code/commands/*.md` and register. |
| **P0** | G41: Session label UI | Field exists, no UI to set. |
| **P1** | G28-G29: Queue messages + Esc to interrupt | Critical during long streams. |
| **P1** | G84: Pin-to-bottom toggle | Easy: `m.pinToBottom bool` + `Ctrl+End`. |
| **P1** | G95: Context-aware help overlay | Big discoverability win. |
| **P1** | G23-G24: Placeholder + keybinding hints in input | Easy polish. |
| **P1** | G43: Elapsed-time indicator | Trivial. |
| **P1** | G44: Background-task badge in status bar | Pair with `Ctrl+B` panel. |
| **P1** | G97-G100: Sub-agent panel | Important if/when sub-agents are added. |
| **P2** | G12-G14: Image paste + preview | Requires kitty/chafa. |
| **P2** | G16-G17: Vim/Emacs mode toggle | Larger refactor. |
| **P2** | G19, G109, G111, G112: Panels for MCP / plugins / skills / memory | Discoverability of installed capabilities. |
| **P2** | G25: Token-count warning indicator | Easy. |
| **P2** | G34: Search-forward `/` key in viewport | Easy. |
| **P2** | G50: Line-of-N total indicator | Easy. |
| **P2** | G56: JSON export | Easy. |
| **P2** | G58: OSC52 clipboard fallback | Easy. |
| **P2** | G59: OSC 8 hyperlinks | Easy. |
| **P3** | G31: Line numbers in code blocks | `chroma.WithLineNumbers()`. |
| **P3** | G60: Tabbed panels / multi-chat | Big refactor. |
| **P3** | G85-G86: Streaming rate / TTFT | Easy but low value. |
| **P3** | G78-G80: Mermaid / LaTeX / collapsible | Out of scope (terminal). |

---

## 4. Architectural findings

- **Single 3,066-line `model.go` file** — split candidate: `state_handlers.go`, `rendering.go`, `keymap.go` (currently 17 LOC), `input.go`.
- **State machine** uses iota with 20 states — clear, but no FSM table/state transition map.
- **Command registry** is solid (single source of truth for slash / palette / quick actions).
- **Mouse** fully wired (wheel + click) — solid.
- **Streaming renderer** throttles well — solid.
- **Tests** are concentrated in `tui_test.go` (1,135 LOC, 50+ tests).
- **No virtualization for very long tool-card lists** (capped at 50).

---

## 5. Dependencies — quick scan

`go list -m -u all` shows many available upgrades:

- `chroma v2.20.0 → v2.26.1` (syntax highlighting)
- `bubbletea` (current: v1.3.10) — many ecosystem upgrades
- `lipgloss v1.1.1-...` (current dev build)
- `clipperhouse/displaywidth v0.9.0 → v0.11.0`
- `charmbracelet/x/exp/golden` (latest)
- `charmbracelet/x/exp/slice` (latest)
- `aymerick/douceur v0.2.0` (CSS parser)
- `bits-and-blooms/bitset v1.24.4` (latest)

Recommend staged upgrade: dev/minor first, runtime majors last (dependency-upgrade skill workflow).

---

## 6. Next step (recommendation)

Pursue **P0 items** first as a single milestone. Specifically:

1. Status bar: cwd + git branch + context % + elapsed time
2. Input editor keybindings (Ctrl+A, Ctrl+Z in input, word jump)
3. Auto-suggest / ghost-text
4. Custom slash commands from `~/.claw-code/commands/*.md`
5. Session label UI (write to `Label` field)
6. Pin-to-bottom + Esc to interrupt

After P0: P1 panels (MCP / plugins / skills), sub-agent display, image paste.
