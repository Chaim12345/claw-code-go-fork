package tui

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func refreshGitContext(workDir string) (branch string, dirty bool) {
	branch = runGitCmd(workDir, "branch", "--show-current")
	if branch == "" {
		return "", false
	}
	status := runGitCmd(workDir, "status", "--porcelain")
	dirty = strings.TrimSpace(status) != ""
	return branch, dirty
}

func refreshWorkingDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	home, err := os.UserHomeDir()
	if err == nil && strings.HasPrefix(wd, home) {
		return "~" + wd[len(home):]
	}
	return wd
}

func shortDir(path string) string {
	if path == "" {
		return ""
	}
	parts := strings.Split(path, string(filepath.Separator))
	if len(parts) <= 3 {
		return path
	}
	return filepath.Join(parts[0], "...", parts[len(parts)-2], parts[len(parts)-1])
}

func runGitCmd(workDir string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = workDir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
