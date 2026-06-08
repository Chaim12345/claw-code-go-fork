# Ralph Spec: Remaining Phases

Working directory: /root/claw-code-go-fork
Branch: dsml
DB path: ~/.claw-code/sessions.db (SQLite, schema v2 with FTS5)
Key packages: internal/session/, internal/context/, internal/web/, internal/runtime/
All prior phases (0-10) are COMPLETE. Only these items remain.

## Phase 11 — Session History Context (remaining items)

11.1 and 11.2 are DONE (GetRecentSessions and GetSessionSummary already exist in internal/session/).

- [ ] 11.3 Update assembler to include "Recent Sessions" section
  - Edit internal/context/assembler.go
  - After loading project map, add a new section that calls session.GetRecentSessions(3) and session.GetSessionSummary() for each
  - Format as markdown: "# Recent Sessions\n\n- **Session ABC** (2026-06-08, deepseek/expert): Summary text...\n"
  - Deduct from tokenBudget like other sections
  - The assembler needs a SessionStore reference — add it to the Assembler struct (optional, nil means skip)
  - In internal/web/server.go, when creating the assembler, inject the session store

- [ ] 11.4 Cap session history context at 2K tokens
  - Use FitToBudget from internal/api/providers/deepseek/limits.go (or copy the pattern)
  - Ensure the "Recent Sessions" section never exceeds 2K tokens (8000 chars)
  - If truncated, append "[... truncated ...]" marker

- [ ] 11.5 Add tests for session history context
  - Add test in internal/context/assembler_test.go
  - Test that Assemble() includes "Recent Sessions" when SessionStore is provided
  - Test that section is capped at 2K tokens
  - Test that nil SessionStore skips the section (no crash)

### Success Criteria
- [ ] After returning, AI knows what was discussed last session
- [ ] Token budget stays within limits
- [ ] No performance degradation on session creation

---

## Phase 12 — Codebase Search Integration

### Goal
Add semantic search over the codebase so the AI can find relevant code without reading entire files.

### Tasks

- [ ] 12.1 Add BM25/TF-IDF index for Go files (stdlib only)
  - Create internal/search/index.go
  - Walk .go files in the project root (excluding .git, vendor, node_modules)
  - Tokenize: split on whitespace/punctuation, lowercase, filter Go keywords and common stopwords
  - Build inverted index: map[token]map[filePath][]positions
  - Compute BM25 scores: IDF = log((N - df + 0.5) / (df + 0.5)), TF component uses term frequency / doc length
  - Store index in memory, rebuild on demand (no persistence needed for MVP)
  - Provide Index type with Build(dir string) error and Search(query string, limit int) []SearchResult

- [ ] 12.2 Add `SearchCodebase(query string, limit int) ([]SearchResult, error)` function
  - Create internal/search/search.go
  - SearchResult struct: FilePath string, Score float64, Snippet string (3-line context around best match)
  - Tokenize query the same way as index
  - Score each document using BM25, sort descending, return top N
  - Empty query returns nil (no error)

- [ ] 12.3 Add `/api/search` HTTP endpoint
  - Edit internal/web/server.go
  - GET /api/search?q=...&limit=10
  - Build index lazily (once per server start, or on first request)
  - Return JSON array of SearchResult objects
  - Add auth check (same as other /api/* endpoints)

- [ ] 12.4 Update context assembler to include top search results for user query
  - Edit internal/context/assembler.go
  - Add SearchIndex field to Assembler struct (optional, nil means skip)
  - In Assemble(), if SearchIndex is set, search for relevant code and include as "# Relevant Code\n\n" section
  - Cap at 1K tokens
  - NOTE: assembler.Assemble() currently doesn't receive the user query — that's OK for now, skip this sub-item if assembler doesn't have access to the query. It can be wired later.

- [ ] 12.5 Add tests for codebase search
  - internal/search/index_test.go: test Build + Search on a temp dir with .go files
  - internal/search/search_test.go: test SearchCodebase returns correct results
  - internal/web/ test for /api/search endpoint (if time permits)

### Success Criteria
- [ ] AI can find relevant code snippets by semantic query
- [ ] Search completes in <100ms for typical codebase
- [ ] No false positives on unrelated files

---

## Phase 13 — Multi-Turn Session Memory

### Goal
Improve multi-turn conversation quality by maintaining better context across turns.

### Tasks

- [ ] 13.1 Add automatic context compaction for long sessions
  - Edit internal/runtime/conversation.go
  - When session exceeds 20 turns (40 messages), compact older messages
  - Keep first 2 messages (system + first user) and last 10 messages intact
  - Replace middle messages with a summary: "Messages 3-30 summarized: [key topics discussed]"
  - Trigger compaction in the conversation loop before sending to provider
  - Use a simple heuristic: count messages, if > 40, compact the middle
  - Do NOT delete from the session store — only compact the in-memory messages sent to the provider

- [ ] 13.2 Add "session notes" feature — AI can write persistent notes per session
  - Add session_notes table to SQLite schema (migration v3)
  - Schema: id, session_id, key TEXT, value TEXT, created_at, updated_at
  - Add SaveSessionNote(sessionID, key, value string) error to SessionStore
  - Add GetSessionNotes(sessionID string) (map[string]string, error) to SessionStore
  - Add /api/sessions/{id}/notes GET and POST endpoints
  - Notes are injected into the assembler as "# Session Notes\n\n" section (capped at 500 tokens)

- [ ] 13.3 Add session context to system prompt (what was discussed, key decisions)
  - When a session is resumed (via --session flag or WebUI session select), load the session's notes
  - Include in assembler output so the AI has context from the previous conversation
  - This builds on 13.2: if session has notes, inject them

- [ ] 13.4 Add tests for multi-turn memory
  - Test compaction: create 50 messages, call compact, verify first 2 and last 10 are preserved
  - Test session notes CRUD: SaveSessionNote + GetSessionNotes roundtrip
  - Test session notes in assembler: when SessionStore has notes, Assemble includes them

### Success Criteria
- [ ] Long sessions (>20 turns) maintain context quality
- [ ] AI can recall decisions from early in conversation
- [ ] No context window overflow

---

## Technical Constraints

- Use stdlib only where possible (no external deps for search/index)
- SQLite via modernc.org/sqlite (pure Go, no CGO)
- All new code must have tests
- `go build ./...` must pass
- `go test ./internal/session/... ./internal/context/... ./internal/search/...` must pass
- Do NOT modify any existing test files that are .bak or in eval/
- Pre-existing failures in delta_test.go, auth_test.go, pipeline_integration_test.go.bak — do not touch
