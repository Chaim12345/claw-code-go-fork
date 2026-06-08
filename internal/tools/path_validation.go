package tools

import (
	"claw-code-go/internal/errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var filepathAbs = filepath.Abs

var filepathEvalSymlinks = filepath.EvalSymlinks

func defaultAllowedDirs() []string {
	var dirs []string

	if cwd, err := os.Getwd(); err == nil {
		dirs = append(dirs, cwd)
		for p := filepath.Dir(cwd); p != filepath.Dir(p); p = filepath.Dir(p) {
			dirs = append(dirs, p)
		}
	}

	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, home)
	}

	dirs = append(dirs, "/tmp", "/var/tmp")
	return dirs
}

func ValidatePath(path string, allowedDirs []string) error {
	if allowedDirs == nil {
		allowedDirs = defaultAllowedDirs()
	}

	absPath, err := filepathAbs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}
	realPath, err := filepathEvalSymlinks(absPath)
	if err == nil {
		return pathInsideAllowed(realPath, allowedDirs)
	}
	current := absPath
	for {
		parent := filepath.Dir(current)
		if parent == current {
			if restrictedBypass {
				return nil
			}
			return errors.Restricted(fmt.Errorf("path traversal blocked: cannot resolve %s: %w", absPath, err))
		}
		if parentReal, perr := filepathEvalSymlinks(parent); perr == nil {
			return pathInsideAllowed(parentReal, allowedDirs)
		}
		current = parent
	}
}

func pathInsideAllowed(realPath string, allowedDirs []string) error {
	for _, dir := range allowedDirs {
		absDir, dirErr := filepathAbs(dir)
		if dirErr != nil {
			continue
		}
		if strings.HasPrefix(realPath, absDir+string(os.PathSeparator)) ||
			realPath == absDir {
			return nil
		}
	}
	if restrictedBypass {
		return nil
	}
	return errors.Restricted(fmt.Errorf("path traversal blocked: %s is outside allowed directories", realPath))
}
