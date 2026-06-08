package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestMerge(t *testing.T) {
	dst := &Settings{
		Model:          "old-model",
		PermissionMode: "default",
		MaxTokens:      4096,
	}
	src := &Settings{
		Model:     "new-model",
		MaxTokens: 8192,
	}
	merge(dst, src)

	if dst.Model != "new-model" {
		t.Errorf("Model = %q, want %q", dst.Model, "new-model")
	}
	if dst.PermissionMode != "default" {
		t.Error("PermissionMode should be unchanged")
	}
	if dst.MaxTokens != 8192 {
		t.Errorf("MaxTokens = %d, want 8192", dst.MaxTokens)
	}
}

func TestMerge_EmptySrcNoop(t *testing.T) {
	dst := &Settings{
		Model:     "keep-me",
		MaxTokens: 100,
	}
	src := &Settings{}
	merge(dst, src)

	if dst.Model != "keep-me" {
		t.Errorf("Model = %q, want %q", dst.Model, "keep-me")
	}
	if dst.MaxTokens != 100 {
		t.Errorf("MaxTokens = %d, want 100", dst.MaxTokens)
	}
}

func TestMerge_Slices(t *testing.T) {
	dst := &Settings{
		AllowedTools: []string{"a"},
		BlockedTools: []string{"b"},
	}
	src := &Settings{
		AllowedTools: []string{"x", "y"},
		BlockedTools: []string{"z"},
	}
	merge(dst, src)

	if len(dst.AllowedTools) != 2 || dst.AllowedTools[0] != "x" {
		t.Errorf("AllowedTools = %v, want [x y]", dst.AllowedTools)
	}
	if len(dst.BlockedTools) != 1 || dst.BlockedTools[0] != "z" {
		t.Errorf("BlockedTools = %v, want [z]", dst.BlockedTools)
	}
}

func TestMerge_Theme(t *testing.T) {
	dst := &Settings{Theme: "dark"}
	src := &Settings{Theme: "light"}
	merge(dst, src)
	if dst.Theme != "light" {
		t.Errorf("Theme = %q, want %q", dst.Theme, "light")
	}
}

func TestMerge_CompactionMaxInputTokens(t *testing.T) {
	dst := &Settings{CompactionMaxInputTokens: 100}
	src := &Settings{CompactionMaxInputTokens: 200}
	merge(dst, src)
	if dst.CompactionMaxInputTokens != 200 {
		t.Errorf("CompactionMaxInputTokens = %d, want 200", dst.CompactionMaxInputTokens)
	}
}

func TestWriteProject_PreservesUnmanagedFields(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	existing := map[string]any{
		"mcpServers": map[string]any{"myserver": "value"},
		"model":      "old",
	}
	data, _ := json.MarshalIndent(existing, "", "  ")
	os.MkdirAll(".claude", 0o755)
	os.WriteFile(".claude/settings.json", data, 0o600)

	s := &Settings{Model: "new", Theme: "dark"}
	err := WriteProject(s)
	if err != nil {
		t.Fatalf("WriteProject: %v", err)
	}

	raw, _ := os.ReadFile(".claude/settings.json")
	var got map[string]any
	json.Unmarshal(raw, &got)

	if got["model"] != "new" {
		t.Errorf("model = %v, want 'new'", got["model"])
	}
	if got["theme"] != "dark" {
		t.Errorf("theme = %v, want 'dark'", got["theme"])
	}
	if _, ok := got["mcpServers"]; !ok {
		t.Error("mcpServers was not preserved")
	}
}

func TestInitProject(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	err := InitProject("my-model")
	if err != nil {
		t.Fatalf("InitProject: %v", err)
	}

	data, _ := os.ReadFile(".claude/settings.json")
	var s Settings
	json.Unmarshal(data, &s)

	if s.Model != "my-model" {
		t.Errorf("Model = %q, want %q", s.Model, "my-model")
	}
	if s.PermissionMode != "default" {
		t.Errorf("PermissionMode = %q, want %q", s.PermissionMode, "default")
	}
}

func TestInitProject_DefaultModel(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	InitProject("")
	data, _ := os.ReadFile(".claude/settings.json")
	var s Settings
	json.Unmarshal(data, &s)

	if s.Model != "expert" {
		t.Errorf("Model = %q, want %q", s.Model, "expert")
	}
}

func TestInitProject_AlreadyExists(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	os.MkdirAll(".claude", 0o755)
	os.WriteFile(filepath.Join(".claude", "settings.json"), []byte(`{}`), 0o600)

	err := InitProject("model")
	if err != os.ErrExist {
		t.Errorf("InitProject with existing file: err = %v, want os.ErrExist", err)
	}
}

func TestLoad_NoFiles(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	s := Load()
	if s == nil {
		t.Fatal("Load() returned nil")
	}
	if s.Model != "" {
		t.Errorf("Model = %q, want empty", s.Model)
	}
}

func TestLoad_ProjectSettings(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	os.MkdirAll(".claude", 0o755)
	data := `{"model": "test-model", "maxTokens": 4096}`
	os.WriteFile(filepath.Join(".claude", "settings.json"), []byte(data), 0o600)

	s := Load()
	if s.Model != "test-model" {
		t.Errorf("Model = %q, want %q", s.Model, "test-model")
	}
	if s.MaxTokens != 4096 {
		t.Errorf("MaxTokens = %d, want 4096", s.MaxTokens)
	}
}

func TestLoad_LocalOverridesProject(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	os.MkdirAll(".claude", 0o755)
	os.WriteFile(filepath.Join(".claude", "settings.json"), []byte(`{"model": "project-model"}`), 0o600)
	os.WriteFile(filepath.Join(".claude", "settings.local.json"), []byte(`{"model": "local-model"}`), 0o600)

	s := Load()
	if s.Model != "local-model" {
		t.Errorf("Model = %q, want %q (local should override project)", s.Model, "local-model")
	}
}

func TestLoad_InvalidJSONSkipped(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	os.MkdirAll(".claude", 0o755)
	os.WriteFile(filepath.Join(".claude", "settings.json"), []byte(`not json`), 0o600)

	s := Load()
	if s == nil {
		t.Fatal("Load() returned nil with invalid JSON")
	}
}

func TestWriteProject_MkdirAllError(t *testing.T) {
	orig := osMkdirAll
	osMkdirAll = func(string, os.FileMode) error {
		return os.ErrPermission
	}
	defer func() { osMkdirAll = orig }()

	s := &Settings{Model: "test"}
	err := WriteProject(s)
	if err == nil {
		t.Fatal("expected error when osMkdirAll fails")
	}
	if !errors.Is(err, os.ErrPermission) {
		t.Errorf("err = %v, want os.ErrPermission", err)
	}
}

func TestWriteProject_MarshalIndentError(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	origMarshal := jsonMarshalIndent
	jsonMarshalIndent = func(v any, prefix string, indent string) ([]byte, error) {
		return nil, os.ErrInvalid
	}
	defer func() { jsonMarshalIndent = origMarshal }()

	s := &Settings{Model: "test"}
	err := WriteProject(s)
	if err == nil {
		t.Fatal("expected error when jsonMarshalIndent fails")
	}
	if !errors.Is(err, os.ErrInvalid) {
		t.Errorf("err = %v, want os.ErrInvalid", err)
	}
}

func TestWriteProject_AllFields(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	s := &Settings{
		Model:          "m",
		PermissionMode: "bypass",
		AllowedTools:   []string{"a"},
		BlockedTools:   []string{"b"},
		MaxTokens:      100,
		Theme:          "dark",
	}
	if err := WriteProject(s); err != nil {
		t.Fatalf("WriteProject: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(".claude", "settings.json"))
	var got map[string]any
	json.Unmarshal(data, &got)

	if got["model"] != "m" {
		t.Errorf("model = %v", got["model"])
	}
	if got["permissionMode"] != "bypass" {
		t.Errorf("permissionMode = %v", got["permissionMode"])
	}
	if got["maxTokens"] != float64(100) {
		t.Errorf("maxTokens = %v", got["maxTokens"])
	}
	if got["theme"] != "dark" {
		t.Errorf("theme = %v", got["theme"])
	}
}

func TestLoad_AllFields(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	os.MkdirAll(".claude", 0o755)
	data := `{
		"model": "m",
		"permissionMode": "bypass",
		"allowedTools": ["a"],
		"blockedTools": ["b"],
		"maxTokens": 100,
		"compactionMaxInputTokens": 200,
		"theme": "dark"
	}`
	os.WriteFile(filepath.Join(".claude", "settings.json"), []byte(data), 0o600)

	s := Load()
	if s.Model != "m" {
		t.Errorf("Model = %q", s.Model)
	}
	if s.PermissionMode != "bypass" {
		t.Errorf("PermissionMode = %q", s.PermissionMode)
	}
	if len(s.AllowedTools) != 1 || s.AllowedTools[0] != "a" {
		t.Errorf("AllowedTools = %v", s.AllowedTools)
	}
	if len(s.BlockedTools) != 1 || s.BlockedTools[0] != "b" {
		t.Errorf("BlockedTools = %v", s.BlockedTools)
	}
	if s.MaxTokens != 100 {
		t.Errorf("MaxTokens = %d", s.MaxTokens)
	}
	if s.CompactionMaxInputTokens != 200 {
		t.Errorf("CompactionMaxInputTokens = %d", s.CompactionMaxInputTokens)
	}
	if s.Theme != "dark" {
		t.Errorf("Theme = %q", s.Theme)
	}
}
