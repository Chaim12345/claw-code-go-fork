# TUI Gap Analysis: claw-code-go vs Modern TUI Standards

**Date:** 2026-06-03  
**Scope:** Full TUI implementation review  
**Framework:** charmbracelet stack (bubbletea, lipgloss, bubbles, glamour)

---

## Executive Summary

The claw-code-go TUI has **solid foundations** (streaming, themes, tool cards, palettes) but **critical gaps** in polish, accessibility, and competitive feature parity. This document maps every gap and provides implementation guidance.

---

## 1. FRAMEWORK & ARCHITECTURE

### Current State
- **Framework:** charmbracelet/bubbletea (Elm-architecture, proven)
- **Styling:** charmbracelet/lipgloss (14 color tokens, 2 themes)
- **Components:** charmbracelet/bubbles (textarea, textinput, viewport, spinner)
- **Markdown:** charmbracelet/glamour (200ms throttle streaming)
- **Architecture:** Central `model.go` (~2845 lines), flat component structure

### Gaps Identified

| Gap | Severity | Priority | Notes |
|-----|----------|----------|-------|
| `model.go` is monolithic (2845 lines) | Medium | Medium | Should decompose into focused components |
| No error boundary / graceful degradation | High | High | Glamour failures silently drop content |
| No terminal size adaptation strategy | High | Medium | Hardcoded widths (80 cols) |
| No accessibility layer (a11y) | High | Medium | Screen reader, high contrast, reduced motion |
| Missing TUI state persistence (scroll position) | Low | Low | Viewport resets on re-render |

---

## 2. INPUT & COMPOSITION

### Current State
- Multi-line textarea (3 rows visible)
- `/` slash command autocomplete
- `@` file mention autocomplete
- Ctrl+P command palette
- Input history (ring buffer, 500 entries)
- History search
- Undo/redo (50 entries)

### Gaps Identified

| Gap | Severity | Priority | Current | Expected |
|-----|----------|----------|---------|----------|
| **No syntax highlighting in input** | Medium | Low | Plain text | Tree-sitter or regex-based highlighting for code blocks in input |
| **No autocomplete for tool names** | Low | Low | None | When typing tool-like patterns, suggest tool names |
| **No multi-cursor / selection** | Low | Low | None | Standard text selection with Shift+arrows |
| **No clipboard paste detection** | Low | Low | Basic clipboard.go | Detect paste vs typed (bracketed paste mode) |
| **No smart indentation** | Low | Low | None | Auto-indent on Enter, auto-dedent on `}` |
| **No input preview** | Medium | Low | None | Live preview of markdown/code as you type |
| **No character count / token estimate** | Low | Low | None | Show approximate token count in input area |
| **No input validation feedback** | Low | Low | None | Visual feedback for invalid commands before execution |

---

## 3. OUTPUT & DISPLAY

### Current State
- Streaming markdown rendering (glamour, 200ms throttle)
- Tool cards (collapsible, with diff)
- Inline diff renderer (LCS-based)
- Progress spinner
- Status bar (provider, model, permission mode)

### Gaps Identified

| Gap | Severity | Priority | Current | Expected |
|-----|----------|----------|---------|----------|
| **No lazy rendering for long outputs** | High | Medium | Full render on each update | Paginated or virtual-scroll rendering for 1000+ line outputs |
| **No search within conversation** | High | Medium | Conversation search exists but limited | Full regex search with highlighting, jump-to-result |
| **No copy code block button** | High | High | None | One-key copy for code blocks (Ctrl+C when code block focused) |
| **No link handling** | Medium | Low | Links render as text | Detect URLs, offer to open in browser |
| **No image display** | Low | Low | None | Sixel/iTerm2 protocol for inline images (if terminal supports) |
| **No collapsible sections** | Medium | Low | Tool cards only | Any output section should be collapsible |
| **No table rendering** | Medium | Low | Plain text | Render markdown tables with proper alignment |
| **No syntax highlighting in output** | High | High | Glamour handles some | Ensure all code blocks get syntax highlighting |
| **No word wrap control** | Low | Low | Hardcoded 100 cols | User-configurable word wrap width |
| **No scroll position memory** | Low | Low | Viewport resets | Remember scroll position per message |

---

## 4. NAVIGATION & KEYBINDINGS

### Current State
- ↑/↓ history navigation
- Tab completion
- Esc to cancel
- Ctrl+P command palette
- Shift+Tab detection

### Gaps Identified

| Gap | Severity | Priority | Keybinding | Notes |
|-----|----------|----------|------------|-------|
| **No Ctrl+R reverse history search** | High | High | Ctrl+R | Standard shell feature, currently only through palette |
| **No Ctrl+A / Ctrl+E line navigation** | Medium | Medium | Home/End | Basic text editing shortcuts |
| **No Ctrl+W word delete** | Medium | Medium | Ctrl+W | Standard shell feature |
| **No Ctrl+U / Ctrl+K line kill** | Medium | Medium | Ctrl+U/K | Standard shell features |
| **No Ctrl+T tool card toggle** | High | High | Ctrl+T | Mentioned in tool_card.go but not implemented |
| **No Ctrl+O expand all tool cards** | Medium | Medium | Ctrl+O | Referenced in tool_card.go comments |
| **No page up/down in viewport** | High | Medium | PgUp/PgDn | Essential for long outputs |
| **No Ctrl+L screen clear** | Low | Low | Ctrl+L | Standard terminal clear |
| **No F1 help** | Low | Low | F1 | Quick help overlay |
| **No mouse support** | Medium | Low | Scroll wheel | For scrolling, clicking tool cards |
| **No Ctrl+F / Ctrl+G find** | Medium | Medium | Ctrl+F/G | Find in conversation |

---

## 5. VISUAL DESIGN & THEMING

### Current State
- Dark/Light themes (14 color tokens each)
- Theme toggle via Ctrl+P palette
- Derived styles in `styles.go` (~50+ styles)
- Glamour auto-style for markdown

### Gaps Identified

| Gap | Severity | Priority | Notes |
|-----|----------|----------|-------|
| **No custom theme support** | Medium | Medium | Users can't define their own themes |
| **No true color detection** | Medium | Low | Falls back to 256-color palette |
| **No italic/bold detection** | Low | Low | Some terminals don't support |
| **No font size awareness** | Low | Low | TUI assumes fixed cell width |
| **No high contrast mode** | Medium | Medium | Accessibility requirement |
| **No colorblind-friendly palette** | Medium | Low | Alternative color schemes |
| **No transparent/opacity support** | Low | Low | Some terminals support true transparency |
| **No animation control** | Low | Low | Spinner always animated (battery drain) |
| **No focus indicators** | Medium | Medium | Which pane is active not always clear |

---

## 6. SESSION & STATE MANAGEMENT

### Current State
- Session persistence (JSON files)
- Session picker overlay
- Input history persistence
- Token usage tracking
- Cost estimation

### Gaps Identified

| Gap | Severity | Priority | Notes |
|-----|----------|----------|-------|
| **No session search** | High | Medium | Can't search across sessions |
| **No session tagging/labeling** | Medium | Low | Sessions are just IDs |
| **No session diff** | Medium | Low | Compare two sessions |
| **No session export** | Low | Low | Export as markdown/JSON |
| **No session branching** | Medium | Low | Fork a session from a point |
| **No auto-save on crash** | High | Medium | Should persist on SIGTERM/SIGINT |
| **No session metrics dashboard** | Low | Low | Aggregate usage stats |

---

## 7. TOOL INTEGRATION

### Current State
- Tool cards (collapsible)
- Diff viewer (inline)
- Tool icons and badges
- Tool result truncation (4096 chars)

### Gaps Identified

| Gap | Severity | Priority | Notes |
|-----|----------|----------|-------|
| **No tool result search** | Medium | Low | Search within tool outputs |
| **No tool result copy** | High | High | Copy tool result to clipboard |
| **No tool result re-run** | Medium | Low | Re-execute a specific tool call |
| **No tool result comparison** | Low | Low | Diff two tool results |
| **No tool result history** | Low | Low | Browse past tool calls |
| **No tool timeout control** | Medium | Medium | User should be able to set timeout per tool |
| **No tool approval UI improvement** | Medium | Medium | Current permission flow is basic |

---

## 8. ACCESSIBILITY

### Current State
- Keyboard navigation (primary)
- Color themes

### Gaps Identified

| Gap | Severity | Priority | Notes |
|-----|----------|----------|-------|
| **No screen reader support** | High | High | No ARIA-like labels, no semantic markup |
| **No high contrast mode** | High | Medium | WCAG AAA compliance |
| **No reduced motion mode** | Medium | Medium | Disable spinner, animations |
| **No font size scaling** | Medium | Low | Respect terminal font size |
| **No focus management** | Medium | Medium | Clear focus indicators |
| **No keyboard shortcut reference** | High | High | Built-in shortcut cheat sheet |
| **No color-only information** | Medium | Medium | Don't rely solely on color for status |

---

## 9. PERFORMANCE

### Current State
- 200ms render throttle
- Streaming with buffer
- Lazy code block detection

### Gaps Identified

| Gap | Severity | Priority | Notes |
|-----|----------|----------|-------|
| **No render profiling** | Low | Low | No way to measure render time |
| **No memory monitoring** | Low | Low | Long sessions may leak |
| **No viewport virtualization** | High | Medium | Large outputs cause lag |
| **No debounced input** | Low | Low | Input events may flood |
| **No incremental rendering** | Medium | Low | Full re-render each frame |

---

## 10. COMPETITIVE PARITY (vs Claude Code, Aider, Continue.dev)

| Feature | Claude Code | Aider | claw-code | Gap |
|---------|-------------|-------|-----------|-----|
| Streaming output | ✅ | ✅ | ✅ | — |
| Syntax highlighting | ✅ | ✅ | Partial | Needs improvement |
| Inline diff | ✅ | ✅ | ✅ | — |
| Code block copy | ✅ | ✅ | ❌ | **Critical** |
| Ctrl+R history search | ✅ | ✅ | ❌ | **Critical** |
| Session management | ✅ | ✅ | ✅ | — |
| Multi-model support | ✅ | ✅ | ✅ | — |
| Tool approval UI | ✅ | ✅ | Basic | Needs polish |
| Responsive layout | ✅ | ✅ | ❌ | **High** |
| Mouse support | ✅ | ❌ | ❌ | **Medium** |
| Accessibility | Basic | ❌ | ❌ | **Medium** |
| Custom themes | ❌ | ❌ | ❌ | — |
| Session branching | ❌ | ❌ | ❌ | — |
| Image display | ❌ | ❌ | ❌ | — |

---

## PRIORITY MATRIX

### P0 — Critical (Must Fix)
1. **Code block copy** (Ctrl+C when code block focused)
2. **Ctrl+R reverse history search**
3. **Responsive terminal layout** (handle small terminals gracefully)
4. **Graceful error handling** (glamour failures, API errors)

### P1 — High (Should Fix)
1. **Ctrl+T tool card toggle** (referenced but not implemented)
2. **Page Up/Down in viewport**
3. **Screen reader basics** (semantic labels)
4. **High contrast mode**
5. **Tool result copy**
6. **Session search**
7. **Keyboard shortcut reference**

### P2 — Medium (Nice to Have)
1. **Mouse support** (scroll wheel, clicking)
2. **Custom theme support**
3. **Reduced motion mode**
4. **Session branching**
5. **Tool timeout control**
6. **Improved permission approval UI**
7. **Input preview** (markdown live preview)

### P3 — Low (Future)
1. **Image display** (sixel/iTerm2)
2. **Session export**
3. **Session metrics dashboard**
4. **Render profiling**
5. **Font size awareness**

---

## IMPLEMENTATION ROADMAP

### Phase 1: Critical Fixes (1-2 days)
- [ ] Implement code block copy (detect code block under cursor, copy to clipboard)
- [ ] Add Ctrl+R reverse history search
- [ ] Add terminal size detection and graceful degradation
- [ ] Add error boundary for glamour rendering failures

### Phase 2: Core UX (3-5 days)
- [ ] Implement Ctrl+T tool card toggle
- [ ] Add Page Up/Down to viewport
- [ ] Build keyboard shortcut reference panel (F1)
- [ ] Add basic screen reader labels
- [ ] Implement tool result copy

### Phase 3: Polish (1-2 weeks)
- [ ] Mouse support (scroll wheel, clicks)
- [ ] High contrast / reduced motion modes
- [ ] Custom theme loading
- [ ] Session search
- [ ] Input preview

### Phase 4: Advanced (2+ weeks)
- [ ] Image display protocol detection
- [ ] Session branching
- [ ] Session export
- [ ] Advanced accessibility

---

## APPENDIX: FILE-BY-FILE GAPS

### `model.go` (2845 lines)
- **Decompose:** Extract `Update()` into focused handler functions per state
- **Add:** Error boundary wrapper around glamour calls
- **Add:** Terminal size change handling (`tea.WindowSizeMsg`)
- **Add:** Focus management state

### `streaming_renderer.go` (393 lines)
- **Fix:** Glamour error handling (currently silent failures)
- **Add:** Lazy rendering for large outputs
- **Add:** Search within rendered content

### `tool_card.go` (237 lines)
- **Implement:** Ctrl+T toggle (referenced in comments)
- **Implement:** Ctrl+O expand all (referenced in comments)
- **Add:** Tool result copy to clipboard
- **Fix:** Dynamic box width (currently hardcoded 80)

### `slash_menu.go` (186 lines)
- **Add:** Keyboard shortcut hints
- **Add:** Category icons
- **Fix:** Scroll for long command lists

### `palette.go` (232 lines)
- **Add:** Keyboard shortcut display per command
- **Add:** Recently used commands section
- **Fix:** Fuzzy matching improvements

### `theme.go` (73 lines)
- **Add:** Custom theme loading from config file
- **Add:** True color detection and fallback
- **Add:** High contrast theme variant

### `styles.go` (~50+ styles)
- **Add:** High contrast style set
- **Add:** Reduced motion style variants
- **Add:** Focus indicator styles

### `debug/panel.go`
- **Add:** Performance metrics tab
- **Add:** Memory usage tracking
- **Add:** Network request log

---

*Analysis completed. Ready to implement fixes starting with P0 items.*
