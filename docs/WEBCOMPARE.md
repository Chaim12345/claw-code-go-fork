# WebUI Comparison: OpenCode vs Claw-Code-Go

## Executive Summary

| Dimension | OpenCode | Claw-Code-Go (Our) | Gap |
|-----------|----------|-------------------|-----|
| **Architecture** | SolidJS + Vite + TypeScript, SSE transport | Go HTTP + vanilla JS + WebSocket | Different paradigms, both valid |
| **Frontend** | React/SolidJS SPA, component-based | Single HTML file + embedded JS | We're simpler but less maintainable |
| **Backend** | Go HTTP server, SSE streaming | Go HTTP server, WebSocket streaming | Comparable |
| **Session Mgmt** | SQLite-backed, persistent | In-memory map, ephemeral | **CRITICAL GAP** |
| **Agent System** | 6 specialized agents (build, plan, explore, etc.) | Single conversation loop | **MAJOR GAP** |
| **Tool Integration** | 20+ tools, MCP, LSP, batch execution | 10 core tools, MCP | We're catching up |
| **Desktop App** | Tauri 2 (macOS/Linux/Windows) | PWA only | **MAJOR GAP** |
| **Terminal** | Built-in xterm.js (WebGL) | PTY bridge via wterm WASM | Comparable |
| **Mobile** | Full mobile optimization, safe areas | Basic mobile support | We're behind |
| **Themes** | 3 built-in themes + custom CSS | System/light/dark toggle | We're behind |
| **File Management** | File browser, diff viewer, multi-file diff | None | **MAJOR GAP** |
| **Slash Commands** | Full set (/help, /new, /models, /export, /compact, etc.) | Basic (/help, /clear, /session-list, /model, /compact, /cost, /exit) | We're behind |
| **@ Mentions** | @file autocomplete | None | **MAJOR GAP** |
| **Search** | Full-text search across sessions | None | **MAJOR GAP** |
| **Export** | Session export (markdown, JSON) | None | **MAJOR GAP** |
| **Docker** | Full Docker deployment with gateway | Systemd service only | We're behind |
| **PWA** | Full PWA with offline support | Basic PWA (manifest + service worker) | Comparable |
| **Keyboard Shortcuts** | Customizable key bindings | Basic shortcuts (?) overlay | We're behind |
| **Browser Notifications** | AI reply notifications | None | **MAJOR GAP** |
| **Multi-language** | English, Chinese, Japanese | English only | We're behind |

## Architecture Comparison

### OpenCode Architecture
```
packages/
├── opencode/     ← Core: CLI, TUI, Server, all business logic
├── app/          ← Web UI (SolidJS + Vite + Tailwind)
├── desktop/      ← Desktop app (Tauri v2, wraps app)
├── web/          ← Marketing/docs site (Astro + Starlight)
├── ui/           ← Shared SolidJS component library
├── sdk/js/       ← TypeScript SDK (auto-generated from OpenAPI spec)
├── plugin/       ← Plugin API
```

**Key Architectural Patterns:**
- **Event Bus**: Type-safe event publishing/subscribing
- **Agent System**: 6 specialized agents with different permission rulesets
- **Tool Abstraction**: Lazy initialization, metadata callbacks for real-time UI updates
- **Edit Tool**: 9 fallback strategies for precise code editing
- **Session Prompt Loop**: Continuously looping cycle until task completion
- **Permission System**: Layered rules with BashArity for fine-grained control

### Our Architecture
```
internal/web/
├── chatproto/          ← WebSocket protocol definitions
├── chat_session.go     ← WebSocket-to-ConversationLoop bridge
├── server.go           ← HTTP server, PTY bridge, WebSocket handlers
├── session_store.go    ← In-memory session metadata
├── static/             ← Embedded HTML/JS/CSS
│   ├── chat.html       ← Main chat UI
│   ├── index.html      ← Terminal view
│   ├── css/            ← Stylesheets
│   ├── js/             ← JavaScript
│   └── manifest.webmanifest
```

**Key Architectural Patterns:**
- **WebSocket Transport**: Bidirectional JSON messages
- **PTY Bridge**: Raw terminal bytes to browser
- **Embedded Static Files**: Go embed for single-binary deployment
- **Rate Limiting**: Per-IP token buckets
- **Security Headers**: CSP, auth middleware

## Feature Gap Analysis

### CRITICAL GAPS (Must Fix)

#### 1. Persistent Session Storage
**OpenCode**: SQLite-backed, survives restarts, full history
**Ours**: In-memory map, lost on restart

**Impact**: Users lose all conversation history on server restart
**Fix**: Implement SQLite or file-based session persistence

#### 2. Agent System
**OpenCode**: 6 specialized agents (build, plan, explore, compaction, title, general)
**Ours**: Single conversation loop

**Impact**: No task delegation, no read-only planning mode, no specialized exploration
**Fix**: Implement agent delegation with permission rulesets

### MAJOR GAPS (High Priority)

#### 3. File Management
**OpenCode**: File browser, diff viewer, multi-file diff
**Ours**: No file management UI

**Impact**: Users can't browse files or see changes visually
**Fix**: Add file tree component, diff viewer, file preview

#### 4. Desktop Application
**OpenCode**: Tauri 2 native apps (macOS/Linux/Windows)
**Ours**: PWA only

**Impact**: No native desktop experience, no system integration
**Fix**: Consider Tauri or Electron wrapper

#### 5. Search & Export
**OpenCode**: Full-text search, session export (markdown, JSON)
**Ours**: No search, no export

**Impact**: Users can't find old conversations or export work
**Fix**: Add search index, export functionality

#### 6. @ Mentions & Slash Commands
**OpenCode**: @file autocomplete, full slash command set
**Ours**: Basic slash commands, no @ mentions

**Impact**: Less efficient file referencing, limited command set
**Fix**: Implement @ mention with file autocomplete, expand slash commands

### MODERATE GAPS (Medium Priority)

#### 7. Theme System
**OpenCode**: 3 built-in themes + custom CSS
**Ours**: System/light/dark toggle

**Impact**: Less customization, no brand consistency
**Fix**: Add theme presets, custom CSS support

#### 8. Mobile Optimization
**OpenCode**: Full mobile optimization, safe areas, touch gestures
**Ours**: Basic mobile support

**Impact**: Poor mobile experience
**Fix**: Improve touch handling, safe area insets, responsive design

#### 9. Keyboard Shortcuts
**OpenCode**: Customizable key bindings
**Ours**: Basic shortcuts (?) overlay

**Impact**: Power users can't customize workflow
**Fix**: Add shortcut customization, more shortcuts

#### 10. Browser Notifications
**OpenCode**: AI reply notifications
**Ours**: None

**Impact**: Users miss responses when tab is inactive
**Fix**: Add Notification API integration

### MINOR GAPS (Low Priority)

#### 11. Multi-language Support
**OpenCode**: English, Chinese, Japanese
**Ours**: English only

**Impact**: Limited international audience
**Fix**: Add i18n framework, translations

#### 12. Docker Deployment
**OpenCode**: Full Docker with gateway, router, frontend, backend
**Ours**: Systemd service only

**Impact**: Harder to deploy in containers
**Fix**: Add Dockerfile, docker-compose.yml

## Roadmap

### Phase 1: Session Persistence (Week 1-2)
**Goal**: Survive server restarts, full conversation history

**Tasks:**
- [ ] Design SQLite schema for sessions, messages, tool calls
- [ ] Implement session store with CRUD operations
- [ ] Migrate in-memory store to SQLite
- [ ] Add session search and filtering
- [ ] Implement session export (markdown, JSON)
- [ ] Add session import functionality

**Success Criteria:**
- Sessions persist across server restarts
- Users can search old conversations
- Export works for individual sessions

### Phase 2: File Management (Week 3-4)
**Goal**: Browse files, see changes, preview content

**Tasks:**
- [ ] Design file tree component
- [ ] Implement file browser with lazy loading
- [ ] Add file preview (syntax highlighted)
- [ ] Implement diff viewer (unified, side-by-side)
- [ ] Add multi-file diff view
- [ ] Integrate with tool calls (show file changes)

**Success Criteria:**
- Users can browse project files
- File changes are visible in chat
- Diff viewer works for code changes

### Phase 3: Agent System (Week 5-6)
**Goal**: Task delegation, specialized agents, permission control

**Tasks:**
- [ ] Design agent interface and permission system
- [ ] Implement build agent (full access)
- [ ] Implement plan agent (read-only)
- [ ] Implement explore agent (search only)
- [ ] Add agent delegation via task tool
- [ ] Implement permission prompts in chat UI

**Success Criteria:**
- Agents can delegate to sub-agents
- Permission system controls tool access
- Plan mode prevents file modifications

### Phase 4: Enhanced UX (Week 7-8)
**Goal**: Better mobile, themes, shortcuts, notifications

**Tasks:**
- [ ] Add theme presets (Eucalyptus, Claude, Breeze)
- [ ] Implement custom CSS support
- [ ] Improve mobile touch handling
- [ ] Add safe area insets
- [ ] Implement customizable keyboard shortcuts
- [ ] Add browser notifications for AI replies
- [ ] Add @ mention with file autocomplete
- [ ] Expand slash command set

**Success Criteria:**
- Mobile experience is smooth
- Themes work correctly
- Shortcuts are customizable
- Notifications work

### Phase 5: Desktop & Deployment (Week 9-10)
**Goal**: Native apps, Docker, production ready

**Tasks:**
- [ ] Evaluate Tauri vs Electron
- [ ] Implement desktop app wrapper
- [ ] Add Dockerfile and docker-compose.yml
- [ ] Implement gateway for routing
- [ ] Add health checks and monitoring
- [ ] Implement rate limiting per user
- [ ] Add audit logging

**Success Criteria:**
- Desktop app works on macOS/Linux/Windows
- Docker deployment is documented
- Production deployment is stable

## Implementation Priority

### P0 (Must Do)
1. Session persistence (SQLite)
2. File browser + diff viewer
3. Agent system with permissions

### P1 (Should Do)
4. Theme system
5. Mobile optimization
6. Search & export
7. @ mentions & slash commands

### P2 (Nice to Have)
8. Desktop app (Tauri)
9. Docker deployment
10. Multi-language support
11. Browser notifications

## Technical Debt

### Current Issues
1. **Single HTML file**: chat.html is 168 lines, hard to maintain
2. **No component system**: vanilla JS, no reusability
3. **In-memory sessions**: data lost on restart
4. **No tests for frontend**: only Go tests
5. **No CI/CD**: manual deployment

### Recommended Fixes
1. **Componentize frontend**: Use React/SolidJS or Web Components
2. **Add frontend tests**: Jest/Vitest for JS, Playwright for E2E
3. **Implement CI/CD**: GitHub Actions for build/test/deploy
4. **Add error tracking**: Sentry or similar
5. **Implement logging**: Structured logging for frontend

## Success Metrics

### User Experience
- **Session persistence**: 100% of sessions survive restarts
- **File management**: Users can browse 100% of project files
- **Mobile**: 90% of features work on mobile
- **Performance**: < 100ms response time for UI interactions

### Technical
- **Test coverage**: > 80% for Go, > 60% for JS
- **Build time**: < 2 minutes for full build
- **Deployment**: < 5 minutes for new version
- **Uptime**: 99.9% availability

## Conclusion

Our WebUI is functional but lacks the polish and features of OpenCode. The biggest gaps are session persistence, file management, and the agent system. By following this roadmap, we can close these gaps and create a competitive product.

**Immediate Next Steps:**
1. Start Phase 1: Session persistence
2. Design SQLite schema
3. Implement session store
4. Test persistence across restarts

**Long-term Vision:**
- Full-featured WebUI comparable to OpenCode
- Desktop application for power users
- Production-ready deployment with Docker
- Extensible plugin system for custom tools
