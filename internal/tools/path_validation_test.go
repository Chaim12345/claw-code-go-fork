package tools

import (
	"claw-code-go/internal/errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultAllowedDirs(t *testing.T) {
	dirs := defaultAllowedDirs()
	if len(dirs) == 0 {
		t.Fatal("defaultAllowedDirs returned empty slice")
	}

	wantEntries := []string{"/tmp", "/var/tmp"}
	for _, want := range wantEntries {
		found := false
		for _, d := range dirs {
			if d == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("defaultAllowedDirs = %v, want %q included", dirs, want)
		}
	}

	cwd, _ := os.Getwd()
	foundCWD := false
	for _, d := range dirs {
		if d == cwd {
			foundCWD = true
			break
		}
	}
	if !foundCWD {
		t.Errorf("defaultAllowedDirs = %v, want CWD %q included", dirs, cwd)
	}

	parentCount := 0
	for p := filepath.Dir(cwd); p != filepath.Dir(p); p = filepath.Dir(p) {
		parentCount++
	}
	parentDirsFound := 0
	for _, d := range dirs {
		if strings.HasPrefix(cwd, d) && d != cwd {
			parentDirsFound++
		}
	}
	if parentDirsFound < parentCount/2 {
		t.Errorf("defaultAllowedDirs seems to miss parent dirs: cwd=%s dirs=%v", cwd, dirs)
	}
}

func TestValidatePath_NilAllowedDirs(t *testing.T) {
	// path_validation.go:20-21 — nil allowedDirs triggers defaultAllowedDirs
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("hello"), 0o644)

	// With nil, uses defaults (/tmp, /var/tmp). TempDir may not be under those.
	// So we test that nil doesn't panic and falls through to defaults.
	err := ValidatePath(path, nil)
	// Result depends on whether TempDir is under /tmp — on many systems it is.
	if err != nil {
		// TempDir not under default allowed dirs — expected on some systems
		t.Logf("ValidatePath with nil allowedDirs: %v (tempdir may not be under /tmp)", err)
	}
}

func TestValidatePath_AllowedPath(t *testing.T) {
	// path_validation.go:52-60 — path within allowed directory
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("hello"), 0o644)

	err := ValidatePath(path, []string{dir})
	if err != nil {
		t.Fatalf("ValidatePath(%q, [%q]) = %v, want nil", path, dir, err)
	}
}

func TestValidatePath_AllowedDirExact(t *testing.T) {
	// path_validation.go:58 — realPath == absDir
	dir := t.TempDir()

	err := ValidatePath(dir, []string{dir})
	if err != nil {
		t.Fatalf("ValidatePath(%q, [%q]) = %v, want nil", dir, dir, err)
	}
}

func TestValidatePath_BlockedPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("hello"), 0o644)

	err := ValidatePath(path, []string{"/some/other/dir"})
	if err == nil {
		t.Fatal("expected path traversal error for path outside allowed dirs")
	}
	if !errors.IsRestricted(err) {
		t.Errorf("expected RestrictedError for out-of-scope path, got: %v", err)
	}
}

func TestValidatePath_NonexistentFile_ParentAllowed(t *testing.T) {
	// path_validation.go:29-48 — file doesn't exist, validate parent
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.txt")

	err := ValidatePath(path, []string{dir})
	if err != nil {
		t.Fatalf("ValidatePath for nonexistent file in allowed dir: %v", err)
	}
}

func TestValidatePath_NonexistentFile_ParentBlocked(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "nonexistent.txt")
	os.MkdirAll(filepath.Join(dir, "subdir"), 0o755)

	err := ValidatePath(path, []string{"/some/other/dir"})
	if err == nil {
		t.Fatal("expected path traversal error for nonexistent file outside allowed dirs")
	}
	if !errors.IsRestricted(err) {
		t.Errorf("expected RestrictedError for out-of-scope path, got: %v", err)
	}
}

func TestValidatePath_NonexistentFile_ParentUnresolvable(t *testing.T) {
	// path_validation.go:35-36 — parent dir also can't be resolved
	// Use a deeply nested nonexistent path where even parent doesn't exist
	path := "/nonexistent_root_dir_xyz/sub/file.txt"

	err := ValidatePath(path, []string{"/nonexistent_root_dir_xyz"})
	if err == nil {
		t.Fatal("expected error for unresolvable parent directory")
	}
}

func TestValidatePath_InvalidAllowedDir(t *testing.T) {
	// path_validation.go:40-41 / 53-55 — filepath.Abs error on allowed dir entry
	// This is hard to trigger directly, but we can test that a valid path
	// still works when one allowed dir is bad (continue behavior)
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("hello"), 0o644)

	// Empty string won't cause filepath.Abs to error on most systems,
	// but we verify the loop continues to the next entry.
	err := ValidatePath(path, []string{"", dir})
	if err != nil {
		t.Fatalf("ValidatePath with empty allowedDir entry: %v", err)
	}
}

func TestValidatePath_SymlinkWithinAllowed(t *testing.T) {
	// path_validation.go:29 — symlink resolution
	dir := t.TempDir()
	realPath := filepath.Join(dir, "real.txt")
	os.WriteFile(realPath, []byte("hello"), 0o644)

	linkPath := filepath.Join(dir, "link.txt")
	os.Symlink(realPath, linkPath)

	err := ValidatePath(linkPath, []string{dir})
	if err != nil {
		t.Fatalf("ValidatePath for symlink in allowed dir: %v", err)
	}
}

func TestValidatePath_FilepathAbsError(t *testing.T) {
	// path_validation.go:30-32 — filepathAbs error
	orig := filepathAbs
	filepathAbs = func(path string) (string, error) {
		return "", fmt.Errorf("abs error")
	}
	defer func() { filepathAbs = orig }()

	err := ValidatePath("/tmp/test", []string{"/tmp"})
	if err == nil {
		t.Fatal("expected error when filepathAbs fails")
	}
	if !strings.Contains(err.Error(), "invalid path") {
		t.Errorf("err = %q, want contains 'invalid path'", err.Error())
	}
}

func TestValidatePath_AbsErrorOnAllowedDir_ParentBranch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.txt")
	orig := filepathAbs
	absCallCount := 0
	filepathAbs = func(p string) (string, error) {
		absCallCount++
		if absCallCount == 1 {
			return filepath.Abs(p)
		}
		return "", fmt.Errorf("abs error on dir")
	}
	defer func() { filepathAbs = orig }()

	err := ValidatePath(path, []string{"/bad_allowed_dir"})
	if err == nil {
		t.Fatal("expected error when filepathAbs fails on allowed dir")
	}
	if !strings.Contains(err.Error(), "path traversal blocked") {
		t.Errorf("err = %q, want contains 'path traversal blocked'", err.Error())
	}
}

func TestValidatePath_BlockedPath_WithBypass(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("hello"), 0o644)

	restrictedBypass = true
	defer func() { restrictedBypass = false }()

	err := ValidatePath(path, []string{"/some/other/dir"})
	if err != nil {
		t.Fatalf("expected bypass to allow out-of-scope path, got: %v", err)
	}
}

func TestValidatePath_CWDAutoAllowed(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Skip("cannot get cwd")
	}
	path := filepath.Join(cwd, "some_test_file.txt")

	err = ValidatePath(path, nil)
	if err != nil {
		t.Fatalf("path under CWD should be auto-allowed with nil allowedDirs, got: %v", err)
	}
}

func TestValidatePath_ParentAutoAllowed(t *testing.T) {
	cwd, _ := os.Getwd()
	parent := filepath.Dir(cwd)
	if parent == cwd {
		t.Skip("cwd is root")
	}
	path := filepath.Join(parent, "some_parent_file.txt")

	err := ValidatePath(path, nil)
	if err != nil {
		t.Fatalf("path under parent of CWD should be auto-allowed, got: %v", err)
	}
}

func TestRestrictedBypass(t *testing.T) {
	if restrictedBypass != false {
		t.Fatal("restrictedBypass should start as false")
	}
	SetRestrictedBypass(true)
	if restrictedBypass != true {
		t.Fatal("SetRestrictedBypass(true) should set restrictedBypass to true")
	}
	SetRestrictedBypass(false)
	if restrictedBypass != false {
		t.Fatal("SetRestrictedBypass(false) should set restrictedBypass to false")
	}
}

func TestValidatePath_AbsErrorOnAllowedDir_RealPathBranch(t *testing.T) {
	dir := t.TempDir()
	realFile := filepath.Join(dir, "test.txt")
	os.WriteFile(realFile, []byte("hello"), 0o644)

	orig := filepathAbs
	absCallCount := 0
	filepathAbs = func(p string) (string, error) {
		absCallCount++
		if absCallCount == 1 {
			return filepath.Abs(p)
		}
		return "", fmt.Errorf("abs error on dir")
	}
	defer func() { filepathAbs = orig }()

	err := ValidatePath(realFile, []string{"/bad_dir"})
	if err == nil {
		t.Fatal("expected error when filepathAbs fails on allowed dir")
	}
	if !strings.Contains(err.Error(), "path traversal blocked") {
		t.Errorf("err = %q, want contains 'path traversal blocked'", err.Error())
	}
}
