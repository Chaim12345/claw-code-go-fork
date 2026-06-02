package tui

import (
	"os"
	"os/exec"
	"runtime"
)

// CopyToClipboard copies the given text to the system clipboard.
// It tries multiple clipboard tools based on the OS and available commands.
// Returns true if successful, false otherwise.
func CopyToClipboard(text string) bool {
	if text == "" {
		return false
	}

	// Try OS-specific clipboard commands
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		// macOS: pbcopy
		cmd = exec.Command("pbcopy")
	case "linux":
		// Linux: try xclip or wl-copy (Wayland)
		if hasCommand("xclip") {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if hasCommand("wl-copy") {
			cmd = exec.Command("wl-copy")
		} else if hasCommand("xsel") {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		} else {
			return false
		}
	case "windows":
		// Windows: clip
		cmd = exec.Command("clip")
	default:
		return false
	}

	// Feed the text to the command's stdin
	cmd.Stdin = os.Stdin
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return false
	}

	if err := cmd.Start(); err != nil {
		return false
	}

	if _, err := stdin.Write([]byte(text)); err != nil {
		return false
	}
	stdin.Close()

	if err := cmd.Wait(); err != nil {
		return false
	}

	return true
}

// hasCommand checks if a command exists in PATH.
func hasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
