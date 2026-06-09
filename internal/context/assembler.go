package context

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"claw-code-go/internal/session"
	"claw-code-go/internal/api/providers/deepseek"
)

// Assembler collects and caches project context for injection into the system prompt.
// Each section (system info, git, memory, project map, recent sessions) is assigned
// a portion of the overall token budget so no single piece dominates the context window.
type Assembler struct {
	WorkDir         string
	SessionStore    session.SessionStore
	CurrentSessionID string

	mu        sync.Mutex
	memCache  string
	memMtimes map[string]int64
	mapCache  *ProjectMap
	mapMtime  int64
}

func NewAssembler(workDir string) *Assembler {
	return &Assembler{WorkDir: workDir}
}

func (a *Assembler) Assemble() string {
	var sections []string
	tokenBudget := 12000

	if info := SystemInfo(a.WorkDir); info != "" {
		sections = append(sections, "# Environment\n\n"+info)
		tokenBudget -= estimateTokens(info)
	}

	if git := GitStatus(a.WorkDir); git != "" {
		sections = append(sections, "# Git Status\n\n"+git)
		tokenBudget -= estimateTokens(git)
	}

	if activity := RecentGitActivity(a.WorkDir); activity != "" && tokenBudget > 1000 {
		sections = append(sections, activity)
		tokenBudget -= estimateTokens(activity)
	}

	if mem := a.loadMemory(); mem != "" && tokenBudget > 500 {
		sections = append(sections, "# Project Instructions (CLAUDE.md)\n\n"+mem)
		tokenBudget -= estimateTokens(mem)
	}

	if pm := a.loadProjectMap(); pm != nil && tokenBudget > 1000 {
		mapText := pm.FitToBudget(tokenBudget / 3)
		if mapText != "" {
			sections = append(sections, mapText)
		}
	}

	// Phase 11: inject recent session history for cross-session context.
	if a.SessionStore != nil && tokenBudget > 1000 {
		if sh := a.buildSessionHistorySection(2000); sh != "" {
			sections = append(sections, sh)
			tokenBudget -= estimateTokens(sh)
		}
	}

	if a.SessionStore != nil && a.CurrentSessionID != "" && tokenBudget > 200 {
		if sn := a.buildSessionNotesSection(500); sn != "" {
			sections = append(sections, sn)
			tokenBudget -= estimateTokens(sn)
		}
	}

	if len(sections) == 0 {
		return ""
	}
	return strings.Join(sections, "\n\n")
}

func estimateTokens(s string) int {
	return len(s) / 4
}

func (a *Assembler) loadMemory() string {
	a.mu.Lock()
	defer a.mu.Unlock()

	current := MemoryFileMtimes(a.WorkDir)
	if !mtimesEqual(current, a.memMtimes) {
		a.memCache = LoadMemoryFiles(a.WorkDir)
		a.memMtimes = current
	}
	return a.memCache
}

func (a *Assembler) loadProjectMap() *ProjectMap {
	a.mu.Lock()
	defer a.mu.Unlock()

	currentMtime := getDirMtime(a.WorkDir)
	if a.mapCache != nil && currentMtime == a.mapMtime {
		return a.mapCache
	}

	pm := BuildProjectMap(a.WorkDir)
	a.mapCache = pm
	a.mapMtime = currentMtime
	return pm
}

// SetSessionStore injects a session store so the Recent Sessions
// section can be populated. Safe to call after construction.
func (a *Assembler) SetSessionStore(ss session.SessionStore) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.SessionStore = ss
}

// buildSessionHistorySection returns a markdown-formatted list of recent
// session summaries, capped to maxTokens using the deepseek token estimator.
// Returns "" when the session store is nil or has no sessions.
func (a *Assembler) buildSessionHistorySection(maxTokens int) string {
	if a.SessionStore == nil {
		return ""
	}

	sessions, err := a.SessionStore.GetRecentSessions(3)
	if err != nil || len(sessions) == 0 {
		return ""
	}

	var lines []string
	for _, s := range sessions {
		summary, err := a.SessionStore.GetSessionSummary(s.ID)
		if err != nil || summary == "" {
			continue
		}
		date := s.LastActiveAt.Format("2006-01-02")
		providerModel := s.Provider
		if s.Model != "" {
			providerModel += "/" + s.Model
		}
		if providerModel == "" {
			providerModel = "unknown"
		}
		lines = append(lines, fmt.Sprintf("- **Session %s** (%s, %s): %s",
			s.ID, date, providerModel, summary))
	}

	if len(lines) == 0 {
		return ""
	}

	text := "# Recent Sessions\n\n" + strings.Join(lines, "\n")
	return deepseek.FitToBudget(text, maxTokens)
}

func (a *Assembler) buildSessionNotesSection(maxTokens int) string {
	if a.SessionStore == nil || a.CurrentSessionID == "" {
		return ""
	}
	notes, err := a.SessionStore.GetSessionNotes(a.CurrentSessionID)
	if err != nil || len(notes) == 0 {
		return ""
	}
	var lines []string
	for _, n := range notes {
		if n.Key != "" {
			lines = append(lines, fmt.Sprintf("- **%s**: %s", n.Key, n.Content))
		} else {
			lines = append(lines, fmt.Sprintf("- %s", n.Content))
		}
	}
	text := "# Session Notes\n\n" + strings.Join(lines, "\n")
	return deepseek.FitToBudget(text, maxTokens)
}

func getDirMtime(dir string) int64 {
	info, err := os.Stat(dir)
	if err != nil {
		return 0
	}
	return info.ModTime().UnixNano()
}

func mtimesEqual(a, b map[string]int64) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}
