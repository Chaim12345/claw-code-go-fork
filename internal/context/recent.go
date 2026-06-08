package context

import (
	"fmt"
	"strings"
)

const recentActivityMaxChars = 4000

func RecentGitActivity(workDir string) string {
	var sections []string

	log := runGitCmd(workDir, "log", "--oneline", "--name-status", "-n", "10", "--format=commit %h %s")
	if log != "" {
		sections = append(sections, "# Recent Git Activity\n\n"+log)
	}

	stash := runGitCmd(workDir, "stash", "list")
	if stash != "" {
		stashLines := strings.Split(stash, "\n")
		if len(stashLines) > 3 {
			stash = strings.Join(stashLines[:3], "\n") + fmt.Sprintf("\n... and %d more", len(stashLines)-3)
		}
		sections = append(sections, "# Git Stash\n\n"+stash)
	}

	if len(sections) == 0 {
		return ""
	}

	result := strings.Join(sections, "\n\n")
	if len(result) > recentActivityMaxChars {
		result = result[:recentActivityMaxChars] + "\n... (truncated)"
	}
	return result
}
