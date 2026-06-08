package context

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	maxProjectMapTokens = 8000
	projectMapMaxFiles  = 200
)

var (
	goExportPattern = regexp.MustCompile(`^(?:func\s+(?:\([^)]+\)\s+)?(?:\*?\w+))|(?:type\s+\w+)|(?:func\s+\w+)`)
	ignoreDirs      = map[string]bool{
		".git": true, "node_modules": true, "vendor": true,
		".aider": true, ".claude": true, "dist": true, "build": true,
		".cache": true, "__pycache__": true, ".pytest_cache": true,
	}
	ignoreExts = map[string]bool{
		".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".ico": true,
		".woff": true, ".woff2": true, ".ttf": true, ".eot": true,
		".min.js": true, ".min.css": true, ".sum": true, ".lock": true,
		".db": true, ".sqlite": true, ".sqlite3": true,
	}
)

type ProjectMap struct {
	Root       string
	Structure  string
	KeySymbols string
}

type fileSymbols struct {
	relPath  string
	symbols  []string
	lines    int
}

func BuildProjectMap(root string) *ProjectMap {
	pm := &ProjectMap{Root: root}
	pm.Structure = pm.buildDirectoryTree()
	pm.KeySymbols = pm.extractKeySymbols()
	return pm
}

func (pm *ProjectMap) buildDirectoryTree() string {
	var lines []string
	lines = append(lines, "# Project Structure")
	lines = append(lines, "")

	filepath.Walk(pm.Root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if ignoreDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if info.Size() > 500_000 {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		base := filepath.Base(path)
		if ignoreExts[ext] || ignoreExts[base] {
			return nil
		}

		relPath, _ := filepath.Rel(pm.Root, path)
		parts := strings.Split(relPath, string(filepath.Separator))
		indent := strings.Repeat("  ", len(parts)-1)
		lines = append(lines, fmt.Sprintf("%s%s", indent, parts[len(parts)-1]))

		if len(lines) > 500 {
			return filepath.SkipDir
		}
		return nil
	})

	if len(lines) == 1 {
		return ""
	}
	return strings.Join(lines, "\n")
}

func (pm *ProjectMap) extractKeySymbols() string {
	var files []fileSymbols

	filepath.Walk(pm.Root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}
		if ignoreDirs[filepath.Base(filepath.Dir(path))] {
			return nil
		}

		relPath, _ := filepath.Rel(pm.Root, path)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		content := string(data)
		lines := strings.Split(content, "\n")

		var symbols []string
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if goExportPattern.MatchString(trimmed) && len(trimmed) > 5 {
				if len(trimmed) > 80 {
					trimmed = trimmed[:80] + "..."
				}
				symbols = append(symbols, trimmed)
			}
		}

		if len(symbols) > 0 {
			files = append(files, fileSymbols{
				relPath: relPath,
				symbols: symbols,
				lines:   len(lines),
			})
		}
		return nil
	})

	sort.Slice(files, func(i, j int) bool {
		return len(files[i].symbols) > len(files[j].symbols)
	})

	if len(files) > 50 {
		files = files[:50]
	}

	var sb strings.Builder
	sb.WriteString("# Key Types & Functions\n\n")

	for _, f := range files {
		sb.WriteString(fmt.Sprintf("## %s (%d lines)\n", f.relPath, f.lines))
		for _, sym := range f.symbols {
			sb.WriteString(fmt.Sprintf("  %s\n", sym))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (pm *ProjectMap) TokenEstimate() int {
	totalChars := len(pm.Structure) + len(pm.KeySymbols)
	return totalChars / 4
}

func (pm *ProjectMap) FitToBudget(maxTokens int) string {
	maxChars := maxTokens * 4

	result := pm.Structure + "\n\n" + pm.KeySymbols

	if len(result) <= maxChars {
		return result
	}

	truncated := result[:maxChars-200]
	if nl := strings.LastIndex(truncated, "\n\n"); nl > maxChars-500 {
		truncated = truncated[:nl]
	}
	return truncated + "\n\n[... project map truncated to fit context budget ...]"
}
