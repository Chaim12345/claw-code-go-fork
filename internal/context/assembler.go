package context

import (
	"os"
	"strings"
	"sync"
)

// Assembler collects and caches project context for injection into the system prompt.
type Assembler struct {
	WorkDir string

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
