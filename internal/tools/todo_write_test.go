package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestTodoWriteTool_Definition(t *testing.T) {
	tool := TodoWriteTool()
	if tool.Name != "todo_write" {
		t.Errorf("Name = %q, want %q", tool.Name, "todo_write")
	}
	if len(tool.InputSchema.Required) == 0 {
		t.Error("InputSchema should have required fields")
	}
}

func TestExecuteTodoWrite_MissingAction(t *testing.T) {
	_, err := ExecuteTodoWrite(map[string]any{})
	if err == nil {
		t.Error("missing action should return error")
	}
}

func TestExecuteTodoWrite_UnknownAction(t *testing.T) {
	_, err := ExecuteTodoWrite(map[string]any{"action": "invalid"})
	if err == nil {
		t.Error("unknown action should return error")
	}
}

func TestExecuteTodoWrite_ReadNoFile(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	result, err := ExecuteTodoWrite(map[string]any{"action": "read"})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if result != "[]" {
		t.Errorf("read empty: result = %q, want %q", result, "[]")
	}
}

func TestExecuteTodoWrite_WriteAndRead(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	todos := []TodoItem{
		{ID: "1", Content: "Task 1", Status: "pending", Priority: "high"},
		{ID: "2", Content: "Task 2", Status: "done", Priority: "low"},
	}

	writeResult, err := ExecuteTodoWrite(map[string]any{
		"action": "write",
		"todos":  todos,
	})
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	if writeResult == "" {
		t.Error("write result should not be empty")
	}

	data, err := os.ReadFile(todosPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var got []TodoItem
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("todos in file = %d, want 2", len(got))
	}

	readResult, err := ExecuteTodoWrite(map[string]any{"action": "read"})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var readBack []TodoItem
	json.Unmarshal([]byte(readResult), &readBack)
	if len(readBack) != 2 {
		t.Errorf("read back = %d items, want 2", len(readBack))
	}
}

func TestExecuteTodoWrite_MissingTodosField(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	_, err := ExecuteTodoWrite(map[string]any{"action": "write"})
	if err == nil {
		t.Error("write without todos should return error")
	}
}

func TestExecuteTodoWrite_InvalidStatus(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	todos := []TodoItem{
		{ID: "1", Content: "Task", Status: "invalid", Priority: "high"},
	}
	_, err := ExecuteTodoWrite(map[string]any{"action": "write", "todos": todos})
	if err == nil {
		t.Error("invalid status should return error")
	}
}

func TestExecuteTodoWrite_InvalidPriority(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	todos := []TodoItem{
		{ID: "1", Content: "Task", Status: "pending", Priority: "invalid"},
	}
	_, err := ExecuteTodoWrite(map[string]any{"action": "write", "todos": todos})
	if err == nil {
		t.Error("invalid priority should return error")
	}
}

func TestExecuteTodoWrite_MissingID(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	todos := []TodoItem{
		{Content: "Task", Status: "pending", Priority: "high"},
	}
	_, err := ExecuteTodoWrite(map[string]any{"action": "write", "todos": todos})
	if err == nil {
		t.Error("missing id should return error")
	}
}

func TestExecuteTodoWrite_MissingContent(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	todos := []TodoItem{
		{ID: "1", Status: "pending", Priority: "high"},
	}
	_, err := ExecuteTodoWrite(map[string]any{"action": "write", "todos": todos})
	if err == nil {
		t.Error("missing content should return error")
	}
}

func TestExecuteTodoWrite_WriteCreatesDir(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	if _, err := os.Stat(".claude"); !os.IsNotExist(err) {
		t.Skip(".claude already exists")
	}

	todos := []TodoItem{
		{ID: "1", Content: "Task", Status: "pending", Priority: "high"},
	}
	_, err := ExecuteTodoWrite(map[string]any{"action": "write", "todos": todos})
	if err != nil {
		t.Fatalf("write: %v", err)
	}

	if _, err := os.Stat(filepath.Join(".claude", "todos.json")); os.IsNotExist(err) {
		t.Error("todos.json was not created")
	}
}
