package tui

import (
	"os"
	"path/filepath"
	"strings"
)

type userCommand struct {
	Name        string
	Description string
	Content     string
}

func loadUserCommands() []userCommand {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	dir := filepath.Join(home, ".claw-code", "commands")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var cmds []userCommand
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".md")
		path := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		content := string(data)
		desc := name
		if idx := strings.Index(content, "\n"); idx > 0 {
			firstLine := strings.TrimSpace(content[:idx])
			if strings.HasPrefix(firstLine, "#") {
				desc = strings.TrimSpace(strings.TrimPrefix(firstLine, "#"))
			}
		}
		cmds = append(cmds, userCommand{Name: name, Description: desc, Content: content})
	}
	return cmds
}
