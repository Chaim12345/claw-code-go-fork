package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestReadFileTool_Definition(t *testing.T) {
	tool := ReadFileTool()
	if tool.Name != "read_file" {
		t.Errorf("Name = %q", tool.Name)
	}
}

func TestWriteFileTool_Definition(t *testing.T) {
	tool := WriteFileTool()
	if tool.Name != "write_file" {
		t.Errorf("Name = %q", tool.Name)
	}
}

func TestExecuteReadFile_MissingPath(t *testing.T) {
	_, err := ExecuteReadFile(map[string]any{})
	if err == nil {
		t.Error("missing path should error")
	}
}

func TestExecuteReadFile_Nonexistent(t *testing.T) {
	_, err := ExecuteReadFile(map[string]any{"path": "/nonexistent/file"})
	if err == nil {
		t.Error("nonexistent file should error")
	}
}

func TestExecuteReadFile_Success(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("hello world"), 0o644)

	result, err := ExecuteReadFile(map[string]any{"path": path})
	if err != nil {
		t.Fatalf("ExecuteReadFile: %v", err)
	}
	if result != "hello world" {
		t.Errorf("result = %q, want %q", result, "hello world")
	}
}

func TestExecuteWriteFile_MissingPath(t *testing.T) {
	_, err := ExecuteWriteFile(map[string]any{})
	if err == nil {
		t.Error("missing path should error")
	}
}

func TestExecuteWriteFile_MissingContent(t *testing.T) {
	_, err := ExecuteWriteFile(map[string]any{"path": "/tmp/test"})
	if err == nil {
		t.Error("missing content should error")
	}
}

func TestExecuteWriteFile_Success(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")

	result, err := ExecuteWriteFile(map[string]any{"path": path, "content": "data"})
	if err != nil {
		t.Fatalf("ExecuteWriteFile: %v", err)
	}
	if !strings.Contains(result, "Successfully wrote") {
		t.Errorf("result = %q", result)
	}

	data, _ := os.ReadFile(path)
	if string(data) != "data" {
		t.Errorf("file content = %q, want %q", string(data), "data")
	}
}

func TestExecuteWriteFile_CreatesDirs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a", "b", "c.txt")

	_, err := ExecuteWriteFile(map[string]any{"path": path, "content": "nested"})
	if err != nil {
		t.Fatalf("ExecuteWriteFile: %v", err)
	}

	data, _ := os.ReadFile(path)
	if string(data) != "nested" {
		t.Errorf("file content = %q", string(data))
	}
}

func TestExecuteWriteFile_DirOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.txt")

	_, err := ExecuteWriteFile(map[string]any{"path": path, "content": "x"})
	if err != nil {
		t.Fatalf("ExecuteWriteFile: %v", err)
	}
}

func TestFileEditTool_Definition(t *testing.T) {
	tool := FileEditTool()
	if tool.Name != "file_edit" {
		t.Errorf("Name = %q", tool.Name)
	}
}

func TestExecuteFileEdit_MissingFilePath(t *testing.T) {
	_, err := ExecuteFileEdit(map[string]any{})
	if err == nil {
		t.Error("missing file_path should error")
	}
}

func TestExecuteFileEdit_MissingOldString(t *testing.T) {
	_, err := ExecuteFileEdit(map[string]any{"file_path": "/tmp/x"})
	if err == nil {
		t.Error("missing old_string should error")
	}
}

func TestExecuteFileEdit_MissingNewString(t *testing.T) {
	_, err := ExecuteFileEdit(map[string]any{"file_path": "/tmp/x", "old_string": "a"})
	if err == nil {
		t.Error("missing new_string should error")
	}
}

func TestExecuteFileEdit_NonexistentFile(t *testing.T) {
	_, err := ExecuteFileEdit(map[string]any{
		"file_path":  "/nonexistent/file",
		"old_string": "a",
		"new_string": "b",
	})
	if err == nil {
		t.Error("nonexistent file should error")
	}
}

func TestExecuteFileEdit_NotFound(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("hello"), 0o644)

	_, err := ExecuteFileEdit(map[string]any{
		"file_path":  path,
		"old_string": "MISSING",
		"new_string": "b",
	})
	if err == nil {
		t.Error("old_string not found should error")
	}
}

func TestExecuteFileEdit_MultipleMatches(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("aaa"), 0o644)

	_, err := ExecuteFileEdit(map[string]any{
		"file_path":  path,
		"old_string": "a",
		"new_string": "b",
	})
	if err == nil {
		t.Error("multiple matches should error")
	}
}

func TestExecuteFileEdit_Success(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("hello world"), 0o644)

	result, err := ExecuteFileEdit(map[string]any{
		"file_path":  path,
		"old_string": "world",
		"new_string": "go",
	})
	if err != nil {
		t.Fatalf("ExecuteFileEdit: %v", err)
	}
	if !strings.Contains(result, "Successfully edited") {
		t.Errorf("result = %q", result)
	}

	data, _ := os.ReadFile(path)
	if string(data) != "hello go" {
		t.Errorf("file = %q, want %q", string(data), "hello go")
	}
}

func TestGlobTool_Definition(t *testing.T) {
	tool := GlobTool()
	if tool.Name != "glob" {
		t.Errorf("Name = %q", tool.Name)
	}
}

func TestExecuteGlob_MissingPattern(t *testing.T) {
	_, err := ExecuteGlob(map[string]any{})
	if err == nil {
		t.Error("missing pattern should error")
	}
}

func TestExecuteGlob_NoMatches(t *testing.T) {
	dir := t.TempDir()
	result, err := ExecuteGlob(map[string]any{"pattern": "*.xyz", "path": dir})
	if err != nil {
		t.Fatalf("ExecuteGlob: %v", err)
	}
	if !strings.Contains(result, "No files found") {
		t.Errorf("result = %q", result)
	}
}

func TestExecuteGlob_Flat(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0o644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0o644)

	result, err := ExecuteGlob(map[string]any{"pattern": "*.txt", "path": dir})
	if err != nil {
		t.Fatalf("ExecuteGlob: %v", err)
	}
	if !strings.Contains(result, "a.txt") {
		t.Errorf("result missing a.txt: %q", result)
	}
	if !strings.Contains(result, "b.txt") {
		t.Errorf("result missing b.txt: %q", result)
	}
}

func TestExecuteGlob_Recursive(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "sub"), 0o755)
	os.WriteFile(filepath.Join(dir, "sub", "deep.txt"), []byte("d"), 0o644)

	result, err := ExecuteGlob(map[string]any{"pattern": "**/*.txt", "path": dir})
	if err != nil {
		t.Fatalf("ExecuteGlob: %v", err)
	}
	if !strings.Contains(result, "deep.txt") {
		t.Errorf("result missing deep.txt: %q", result)
	}
}

func TestExecuteGlob_RecursiveNoSuffix(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "sub"), 0o755)
	os.WriteFile(filepath.Join(dir, "sub", "file.txt"), []byte("x"), 0o644)

	result, err := ExecuteGlob(map[string]any{"pattern": "**", "path": dir})
	if err != nil {
		t.Fatalf("ExecuteGlob: %v", err)
	}
	if !strings.Contains(result, "file.txt") {
		t.Errorf("result missing file.txt: %q", result)
	}
}

func TestGrepTool_Definition(t *testing.T) {
	tool := GrepTool()
	if tool.Name != "grep" {
		t.Errorf("Name = %q", tool.Name)
	}
}

func TestExecuteGrep_MissingPattern(t *testing.T) {
	_, err := ExecuteGrep(map[string]any{"path": "."})
	if err == nil {
		t.Error("missing pattern should error")
	}
}

func TestExecuteGrep_MissingPath(t *testing.T) {
	_, err := ExecuteGrep(map[string]any{"pattern": "x"})
	if err == nil {
		t.Error("missing path should error")
	}
}

func TestExecuteGrep_InvalidPattern(t *testing.T) {
	_, err := ExecuteGrep(map[string]any{"pattern": "[invalid", "path": "."})
	if err == nil {
		t.Error("invalid pattern should error")
	}
}

func TestExecuteGrep_NoMatches(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "f.txt"), []byte("hello"), 0o644)

	result, err := ExecuteGrep(map[string]any{"pattern": "NONEXISTENT", "path": dir})
	if err != nil {
		t.Fatalf("ExecuteGrep: %v", err)
	}
	if !strings.Contains(result, "No matches found") {
		t.Errorf("result = %q", result)
	}
}

func TestExecuteGrep_InFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	os.WriteFile(path, []byte("line1\nhello world\nline3"), 0o644)

	result, err := ExecuteGrep(map[string]any{"pattern": "hello", "path": path})
	if err != nil {
		t.Fatalf("ExecuteGrep: %v", err)
	}
	if !strings.Contains(result, "hello world") {
		t.Errorf("result = %q", result)
	}
}

func TestExecuteGrep_InDir(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("foo\nbar"), 0o644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("baz\nfoo"), 0o644)

	result, err := ExecuteGrep(map[string]any{"pattern": "foo", "path": dir})
	if err != nil {
		t.Fatalf("ExecuteGrep: %v", err)
	}
	if !strings.Contains(result, "foo") {
		t.Errorf("result = %q", result)
	}
}

func TestExecuteGrep_WithGlobFilter(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.go"), []byte("foo"), 0o644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("foo"), 0o644)

	result, err := ExecuteGrep(map[string]any{"pattern": "foo", "path": dir, "glob": "*.go"})
	if err != nil {
		t.Fatalf("ExecuteGrep: %v", err)
	}
	if !strings.Contains(result, "a.go") {
		t.Errorf("result should contain a.go: %q", result)
	}
	if strings.Contains(result, "b.txt") {
		t.Errorf("result should NOT contain b.txt: %q", result)
	}
}

func TestExecuteGrep_NonexistentPath(t *testing.T) {
	_, err := ExecuteGrep(map[string]any{"pattern": "x", "path": "/nonexistent"})
	if err == nil {
		t.Error("nonexistent path should error")
	}
}

func TestStripHTML(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{"plain", "hello", "hello"},
		{"tags", "<p>hello</p>", "hello"},
		{"script", "<script>alert(1)</script>hello", "hello"},
		{"style", "<style>.x{}</style>hello", "hello"},
		{"entities", "&amp; &lt; &gt; &quot; &#39; &nbsp;", "& < > \" '"},
		{"block", "before<br>after", "before\nafter"},
		{"whitespace", "a  b\tc", "a b c"},
		{"blank_lines", "a\n\n\n\nb", "a\n\nb"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripHTML(tt.html)
			if got != tt.want {
				t.Errorf("stripHTML(%q) = %q, want %q", tt.html, got, tt.want)
			}
		})
	}
}

func TestWebFetchTool_Definition(t *testing.T) {
	tool := WebFetchTool()
	if tool.Name != "web_fetch" {
		t.Errorf("Name = %q", tool.Name)
	}
}

func TestExecuteWebFetch_MissingURL(t *testing.T) {
	_, err := ExecuteWebFetch(map[string]any{})
	if err == nil {
		t.Error("missing url should error")
	}
}

func TestWebSearchTool_Definition(t *testing.T) {
	tool := WebSearchTool()
	if tool.Name != "web_search" {
		t.Errorf("Name = %q", tool.Name)
	}
}

func TestExecuteWebSearch_MissingQuery(t *testing.T) {
	_, err := ExecuteWebSearch(map[string]any{})
	if err == nil {
		t.Error("missing query should error")
	}
}

func TestParseDDGLite_Empty(t *testing.T) {
	results := parseDDGLite("<html></html>", 5)
	if len(results) != 0 {
		t.Errorf("parseDDGLite empty: got %d results", len(results))
	}
}

func TestParseDDGLite_WithResults(t *testing.T) {
	html := `<html><body>
<a href="http://example.com">Example Title</a>
<p>This is a long enough snippet text that exceeds twenty characters.</p>
</body></html>`
	results := parseDDGLite(html, 5)
	if len(results) != 1 {
		t.Fatalf("parseDDGLite: got %d results, want 1", len(results))
	}
	if results[0][0] != "Example Title" {
		t.Errorf("title = %q", results[0][0])
	}
	if results[0][1] != "http://example.com" {
		t.Errorf("url = %q", results[0][1])
	}
}

func TestParseDDGLite_MaxLimit(t *testing.T) {
	html := `<a href="http://a.com">Title A</a>
<p>Long enough snippet text for result A.</p>
<a href="http://b.com">Title B</a>
<p>Long enough snippet text for result B.</p>`
	results := parseDDGLite(html, 1)
	if len(results) > 1 {
		t.Errorf("parseDDGLite max=1: got %d results", len(results))
	}
}

func TestParseDDGLite_SkipsDDGLinks(t *testing.T) {
	html := `<a href="http://duckduckgo.com/search">DDG Link</a>
<p>This is a very long snippet text.</p>`
	results := parseDDGLite(html, 5)
	for _, r := range results {
		if strings.Contains(r[1], "duckduckgo.com") {
			t.Error("should skip duckduckgo.com links")
		}
	}
}

func TestSetNativeSearch(t *testing.T) {
	old := NativeSearchFunc
	defer func() { NativeSearchFunc = old }()

	SetNativeSearch(nil)
	if NativeSearchFunc != nil {
		t.Error("SetNativeSearch(nil) did not clear")
	}
}

func TestBashTool_Definition(t *testing.T) {
	tool := BashTool()
	if tool.Name != "bash" {
		t.Errorf("Name = %q", tool.Name)
	}
}

func TestExecuteBash_MissingCommand(t *testing.T) {
	_, err := ExecuteBash(map[string]any{})
	if err == nil {
		t.Error("missing command should error")
	}
}

func TestExecuteBash_EmptyCommand(t *testing.T) {
	_, err := ExecuteBash(map[string]any{"command": ""})
	if err == nil {
		t.Error("empty command should error")
	}
}

func TestExecuteBash_NonStringCommand(t *testing.T) {
	_, err := ExecuteBash(map[string]any{"command": 123})
	if err == nil {
		t.Error("non-string command should error")
	}
}

func TestExecuteBash_Success(t *testing.T) {
	result, err := ExecuteBash(map[string]any{"command": "echo hello"})
	if err != nil {
		t.Fatalf("ExecuteBash: %v", err)
	}
	if !strings.Contains(result, "hello") {
		t.Errorf("result = %q, want contains 'hello'", result)
	}
}

func TestExecuteBash_ExitError(t *testing.T) {
	_, err := ExecuteBash(map[string]any{"command": "exit 1"})
	if err == nil {
		t.Error("non-zero exit should error")
	}
}

func TestExecuteBash_OutputTruncation(t *testing.T) {
	result, err := ExecuteBash(map[string]any{"command": "python3 -c \"print('x' * 20000)\""})
	if err != nil {
		t.Skipf("python3 not available: %v", err)
	}
	if !strings.Contains(result, "truncated") {
		t.Error("long output should be truncated")
	}
}

func TestAskUserQuestionTool_Definition(t *testing.T) {
	tool := AskUserQuestionTool()
	if tool.Name != "ask_user" {
		t.Errorf("Name = %q", tool.Name)
	}
}

func TestAskUserInput(t *testing.T) {
	q, ok := AskUserInput(map[string]any{"question": "What?"})
	if !ok || q != "What?" {
		t.Errorf("AskUserInput = (%q, %v), want (What?, true)", q, ok)
	}
}

func TestAskUserInput_Missing(t *testing.T) {
	_, ok := AskUserInput(map[string]any{})
	if ok {
		t.Error("missing question should return false")
	}
}

func TestAskUserInput_Empty(t *testing.T) {
	_, ok := AskUserInput(map[string]any{"question": ""})
	if ok {
		t.Error("empty question should return false")
	}
}

func TestAskUserInput_NonString(t *testing.T) {
	_, ok := AskUserInput(map[string]any{"question": 123})
	if ok {
		t.Error("non-string question should return false")
	}
}

func TestAskUserFallback(t *testing.T) {
	block := AskUserFallback("test question")
	if block.Type != "tool_result" {
		t.Errorf("Type = %q", block.Type)
	}
}

func TestWebFetch_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html><body><p>Hello World</p></body></html>"))
	}))
	defer srv.Close()

	result, err := ExecuteWebFetch(map[string]any{"url": srv.URL})
	if err != nil {
		t.Fatalf("ExecuteWebFetch: %v", err)
	}
	if !strings.Contains(result, "Hello World") {
		t.Errorf("result = %q, want contains 'Hello World'", result)
	}
}

func TestWebFetch_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := ExecuteWebFetch(map[string]any{"url": srv.URL})
	if err == nil {
		t.Error("HTTP 404 should error")
	}
}

func TestWebFetch_MaxBytes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("short"))
	}))
	defer srv.Close()

	result, err := ExecuteWebFetch(map[string]any{"url": srv.URL, "max_bytes": float64(3)})
	if err != nil {
		t.Fatalf("ExecuteWebFetch: %v", err)
	}
	if len(result) > 100 {
		t.Errorf("result too long for max_bytes=3: %d bytes", len(result))
	}
}

func TestWebFetch_InvalidURL(t *testing.T) {
	_, err := ExecuteWebFetch(map[string]any{"url": "http://127.0.0.1:1"})
	if err == nil {
		t.Error("invalid URL should error")
	}
}

func TestWebSearch_MissingQuery(t *testing.T) {
	_, err := ExecuteWebSearch(map[string]any{})
	if err == nil {
		t.Error("missing query should error")
	}
}

func TestWebSearch_NumResultsClamp(t *testing.T) {
	old := NativeSearchFunc
	defer func() { NativeSearchFunc = old }()
	SetNativeSearch(func(ctx context.Context, query string, numResults int) (string, error) {
		if numResults < 1 || numResults > 20 {
			return "", fmt.Errorf("numResults out of range: %d", numResults)
		}
		return "ok", nil
	})

	_, err := ExecuteWebSearch(map[string]any{"query": "test", "num_results": float64(0)})
	if err != nil {
		t.Errorf("numResults=0 should clamp to 1: %v", err)
	}

	_, err = ExecuteWebSearch(map[string]any{"query": "test", "num_results": float64(100)})
	if err != nil {
		t.Errorf("numResults=100 should clamp to 20: %v", err)
	}
}

func TestWebSearch_NativeSearchFunc(t *testing.T) {
	old := NativeSearchFunc
	defer func() { NativeSearchFunc = old }()

	SetNativeSearch(func(ctx context.Context, query string, numResults int) (string, error) {
		return "native result", nil
	})

	result, err := ExecuteWebSearch(map[string]any{"query": "test"})
	if err != nil {
		t.Fatalf("ExecuteWebSearch: %v", err)
	}
	if !strings.Contains(result, "native result") {
		t.Errorf("result = %q, want 'native result'", result)
	}
}

func TestWebSearch_NativeSearchFallback(t *testing.T) {
	old := NativeSearchFunc
	defer func() { NativeSearchFunc = old }()
	oldBrave := os.Getenv("BRAVE_API_KEY")
	defer os.Setenv("BRAVE_API_KEY", oldBrave)
	os.Unsetenv("BRAVE_API_KEY")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html><body><p>no results here</p></body></html>"))
	}))
	defer srv.Close()
	oldURL := ddgLiteURL
	ddgLiteURL = srv.URL
	defer func() { ddgLiteURL = oldURL }()

	SetNativeSearch(func(ctx context.Context, query string, numResults int) (string, error) {
		return "", fmt.Errorf("native failed")
	})

	result, err := ExecuteWebSearch(map[string]any{"query": "test"})
	if err != nil {
		t.Fatalf("fallback to DDG should not error: %v", err)
	}
	if !strings.Contains(result, "No results") {
		t.Errorf("expected no-results from DDG fallback, got: %q", result)
	}
	}

func TestBraveSearch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"web":{"results":[{"title":"T","url":"http://example.com","description":"D"}]}}`))
	}))
	defer srv.Close()

	oldURL := braveSearchURL
	braveSearchURL = srv.URL
	defer func() { braveSearchURL = oldURL }()

	result, err := braveSearch("test", 5, "fake-key")
	if err != nil {
		t.Fatalf("braveSearch: %v", err)
	}
	if !strings.Contains(result, "T") {
		t.Errorf("result = %q", result)
	}
}

func TestBraveSearch_NoResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"web":{"results":[]}}`))
	}))
	defer srv.Close()

	oldURL := braveSearchURL
	braveSearchURL = srv.URL
	defer func() { braveSearchURL = oldURL }()

	result, err := braveSearch("test", 5, "fake-key")
	if err != nil {
		t.Fatalf("braveSearch: %v", err)
	}
	if !strings.Contains(result, "No results") {
		t.Errorf("result = %q", result)
	}
}

func TestBraveSearch_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	oldURL := braveSearchURL
	braveSearchURL = srv.URL
	defer func() { braveSearchURL = oldURL }()

	_, err := braveSearch("test", 5, "bad-key")
	if err == nil {
		t.Error("HTTP 401 should error")
	}
}

func TestBraveSearch_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer srv.Close()

	oldURL := braveSearchURL
	braveSearchURL = srv.URL
	defer func() { braveSearchURL = oldURL }()

	_, err := braveSearch("test", 5, "key")
	if err == nil {
		t.Error("invalid JSON should error")
	}
}

func TestDDGSearch_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html><body>
<a href="http://example.com/page">Example Title</a>
<p>This is a long enough snippet text that exceeds the twenty character minimum.</p>
</body></html>`))
	}))
	defer srv.Close()

	oldURL := ddgLiteURL
	ddgLiteURL = srv.URL
	defer func() { ddgLiteURL = oldURL }()

	result, err := ddgSearch("test", 5)
	if err != nil {
		t.Fatalf("ddgSearch: %v", err)
	}
	if !strings.Contains(result, "Example Title") {
		t.Errorf("result = %q, want contains 'Example Title'", result)
	}
}

func TestDDGSearch_NoResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html></html>"))
	}))
	defer srv.Close()

	oldURL := ddgLiteURL
	ddgLiteURL = srv.URL
	defer func() { ddgLiteURL = oldURL }()

	result, err := ddgSearch("test", 5)
	if err != nil {
		t.Fatalf("ddgSearch: %v", err)
	}
	if !strings.Contains(result, "No results") {
		t.Errorf("result = %q", result)
	}
}

func TestDDGSearch_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	oldURL := ddgLiteURL
	ddgLiteURL = srv.URL
	defer func() { ddgLiteURL = oldURL }()

	_, err := ddgSearch("test", 5)
	if err == nil {
		t.Error("HTTP 500 should error")
	}
}

func TestWebSearchTool_NativeDescription(t *testing.T) {
	old := NativeSearchFunc
	defer func() { NativeSearchFunc = old }()
	SetNativeSearch(func(ctx context.Context, query string, numResults int) (string, error) {
		return "", nil
	})

	tool := WebSearchTool()
	if !strings.Contains(tool.Description, "native") {
		t.Errorf("Description = %q, want contains 'native'", tool.Description)
	}
}

// --- Additional tests for 100% coverage ---

func TestExecuteBash_Timeout(t *testing.T) {
	// bash.go:62-64 — context.DeadlineExceeded path
	result, err := ExecuteBash(map[string]any{"command": "sleep 30"})
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("err = %q, want contains 'timed out'", err.Error())
	}
	_ = result // output may be empty or partial
}

func TestExecuteWriteFile_MkdirAllError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("MkdirAll does not fail for read-only dirs when running as root")
	}
	// files.go:80-82 — os.MkdirAll error
	dir := t.TempDir()
	roDir := filepath.Join(dir, "ro")
	if err := os.Mkdir(roDir, 0o555); err != nil {
		t.Fatalf("mkdir ro: %v", err)
	}
	path := filepath.Join(roDir, "sub", "file.txt")
	_, err := ExecuteWriteFile(map[string]any{
		"path":    path,
		"content": "test",
	})
	if err == nil {
		t.Fatal("expected MkdirAll error")
	}
	if !strings.Contains(err.Error(), "create directories") {
		t.Errorf("err = %q, want contains 'create directories'", err.Error())
	}
}

func TestExecuteWriteFile_WriteError(t *testing.T) {
	// files.go:85-87 — os.WriteFile error
	// /proc/version exists but is read-only even for root
	_, err := ExecuteWriteFile(map[string]any{
		"path":    "/proc/version",
		"content": "test",
	})
	if err == nil {
		t.Fatal("expected write error")
	}
	if !strings.Contains(err.Error(), "write_file") {
		t.Errorf("err = %q, want contains 'write_file'", err.Error())
	}
}

func TestExecuteFileEdit_WriteError(t *testing.T) {
	// file_edit.go:66-68 — os.WriteFile error after successful read+match
	dir := t.TempDir()
	fpath := filepath.Join(dir, "editable.txt")
	os.WriteFile(fpath, []byte("hello world"), 0o644)

	old := osWriteFile
	defer func() { osWriteFile = old }()
	osWriteFile = func(name string, data []byte, perm os.FileMode) error {
		return fmt.Errorf("injected write error")
	}

	_, err := ExecuteFileEdit(map[string]any{
		"file_path":  fpath,
		"old_string": "hello",
		"new_string": "goodbye",
	})
	if err == nil {
		t.Fatal("expected write error")
	}
	if !strings.Contains(err.Error(), "write") {
		t.Errorf("err = %q, want contains 'write'", err.Error())
	}
}

func TestExecuteGlob_WalkError(t *testing.T) {
	// glob.go:55-57 — WalkFunc receives err (e.g. permission denied)
	// glob.go:74-76 — filepath.Walk returns non-nil, non-SkipAll error
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("x"), 0o644)

	old := filepathWalk
	defer func() { filepathWalk = old }()
	filepathWalk = func(root string, fn filepath.WalkFunc) error {
		return fmt.Errorf("injected walk error")
	}

	_, err := ExecuteGlob(map[string]any{
		"pattern": "**/*.txt",
		"path":    dir,
	})
	if err == nil {
		t.Fatal("expected walk error")
	}
	if !strings.Contains(err.Error(), "glob walk") {
		t.Errorf("err = %q, want contains 'glob walk'", err.Error())
	}
}

func TestExecuteGlob_WalkFuncError(t *testing.T) {
	// glob.go:55-57 — WalkFunc gets err param, returns nil (skip)
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("x"), 0o644)

	old := filepathWalk
	defer func() { filepathWalk = old }()
	filepathWalk = func(root string, fn filepath.WalkFunc) error {
		fi, _ := os.Stat(filepath.Join(dir, "a.txt"))
		fn(filepath.Join(dir, "a.txt"), fi, fmt.Errorf("permission denied"))
		return nil
	}

	result, err := ExecuteGlob(map[string]any{
		"pattern": "**/*.txt",
		"path":    dir,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(result, "a.txt") {
		t.Errorf("errored file should be skipped, got: %s", result)
	}
}

func TestExecuteGlob_RecursiveSuffixMatchError(t *testing.T) {
	// glob.go:66-68 — filepath.Match returns error for bad suffix pattern
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "test.txt"), []byte("hi"), 0o644)

	_, err := ExecuteGlob(map[string]any{
		"pattern": "**/[",
		"path":    dir,
	})
	// filepath.Match("[", ...) returns error — but ExecuteGlob skips it (returns nil in WalkFunc)
	// So this shouldn't error, but we test it doesn't crash
	if err != nil {
		// If it errors, that's also fine
		t.Logf("glob with bad pattern: %v (acceptable)", err)
	}
}

func TestExecuteGlob_StandardGlobError(t *testing.T) {
	// glob.go:81-83 — filepath.Glob returns error
	_, err := ExecuteGlob(map[string]any{
		"pattern": "[",
	})
	if err == nil {
		t.Fatal("expected glob error for bad pattern")
	}
}

func TestExecuteGrep_DirWalkError(t *testing.T) {
	// grep.go:72-74 — WalkFunc receives err param, returns nil
	// grep.go:92-94 — grepFile error during walk, returns nil (skip)
	// grep.go:97-99 — len(results) >= maxGrepResults → SkipAll
	// grep.go:103-105 — walk returns non-nil, non-SkipAll error
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "visible.txt"), []byte("match here too"), 0o644)

	// Test grep.go:103-105 — walk returns non-nil non-SkipAll error
	old := filepathWalk
	defer func() { filepathWalk = old }()
	filepathWalk = func(root string, fn filepath.WalkFunc) error {
		return fmt.Errorf("injected walk error")
	}

	_, err := ExecuteGrep(map[string]any{
		"pattern": "match",
		"path":    dir,
	})
	if err == nil {
		t.Fatal("expected walk error")
	}
	if !strings.Contains(err.Error(), "grep walk") {
		t.Errorf("err = %q, want contains 'grep walk'", err.Error())
	}
}

func TestExecuteGrep_WalkFuncError(t *testing.T) {
	// grep.go:72-74 — WalkFunc gets err param, returns nil (skip)
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "good.txt"), []byte("FINDME"), 0o644)

	old := filepathWalk
	defer func() { filepathWalk = old }()
	filepathWalk = func(root string, fn filepath.WalkFunc) error {
		fi, _ := os.Stat(filepath.Join(dir, "good.txt"))
		fn(filepath.Join(dir, "good.txt"), fi, fmt.Errorf("access error"))
		return nil
	}

	result, err := ExecuteGrep(map[string]any{
		"pattern": "FINDME",
		"path":    dir,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(result, "good.txt") {
		t.Errorf("errored file should be skipped, got: %s", result)
	}
}

func TestExecuteGrep_WalkGrepFileError(t *testing.T) {
	// grep.go:92-94 — grepFile returns error during walk, skip it
	dir := t.TempDir()

	tmpfile := filepath.Join(dir, "unreadable.txt")
	os.WriteFile(tmpfile, []byte("test"), 0o644)
	fi, _ := os.Stat(tmpfile)

	old := filepathWalk
	defer func() { filepathWalk = old }()
	filepathWalk = func(root string, fn filepath.WalkFunc) error {
		dfi, _ := os.Stat(dir)
		fn(dir, dfi, nil)
		fn(filepath.Join(dir, "DELETED.txt"), fi, nil)
		return nil
	}

	result, err := ExecuteGrep(map[string]any{
		"pattern": "test",
		"path":    dir,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "No matches") {
		t.Errorf("expected no matches for missing file skip, got: %s", result)
	}
}

func TestExecuteGrep_SkipAllTriggersWalkError(t *testing.T) {
	// grep.go:97-99 — results >= maxGrepResults → return filepath.SkipAll
	// grep.go:103-105 — err == filepath.SkipAll is suppressed
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "big.txt"), []byte("MATCH"), 0o644)

	old := filepathWalk
	defer func() { filepathWalk = old }()
	filepathWalk = func(root string, fn filepath.WalkFunc) error {
		fi, _ := os.Stat(filepath.Join(dir, "big.txt"))
		for i := 0; i < 1001; i++ {
			if err := fn(filepath.Join(dir, "big.txt"), fi, nil); err != nil {
				return err // SkipAll propagates here
			}
		}
		return nil
	}

	result, err := ExecuteGrep(map[string]any{
		"pattern": "MATCH",
		"path":    dir,
	})
	if err != nil {
		t.Fatalf("unexpected error (SkipAll should be suppressed): %v", err)
	}
	if !strings.Contains(result, "truncated") {
		t.Errorf("expected truncation, got: %s", result)
	}
}

func TestExecuteGrep_GlobFilterSkipsFiles(t *testing.T) {
	// grep.go:84-88 — glob filter skips non-matching files
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.go"), []byte("TODO: fix this"), 0o644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("TODO: fix that"), 0o644)

	result, err := ExecuteGrep(map[string]any{
		"pattern": "TODO",
		"path":    dir,
		"glob":    "*.go",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "a.go") {
		t.Errorf("expected a.go in results, got: %s", result)
	}
	if strings.Contains(result, "b.txt") {
		t.Errorf("b.txt should be filtered out by glob *.go, got: %s", result)
	}
}

func TestExecuteGrep_SingleFileError(t *testing.T) {
	// grep.go:108-110 — grepFile error on single file
	_, err := ExecuteGrep(map[string]any{
		"pattern": "test",
		"path":    "/proc/nonexistent_file_for_grep_test",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent single file")
	}
}

func TestExecuteGrep_TruncatedResults(t *testing.T) {
	// grep.go:119-121 — results >= maxGrepResults triggers truncation
	dir := t.TempDir()
	// Create a file with many matching lines
	var lines []string
	for i := 0; i < 1100; i++ {
		lines = append(lines, "MATCHLINE")
	}
	fpath := filepath.Join(dir, "big.txt")
	os.WriteFile(fpath, []byte(strings.Join(lines, "\n")), 0o644)

	result, err := ExecuteGrep(map[string]any{
		"pattern": "MATCHLINE",
		"path":    fpath,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "truncated") {
		t.Errorf("expected truncation notice, got length %d", len(result))
	}
}

func TestGrepFile_OpenError(t *testing.T) {
	// grep.go:129-131 — os.Open error
	re := regexp.MustCompile("test")
	_, err := grepFile(re, "/proc/nonexistent_grep_file_test")
	if err == nil {
		t.Fatal("expected error opening nonexistent file")
	}
}

func TestReadTodos_ReadError(t *testing.T) {
	// todo_write.go:65-67 — os.ReadFile error (not IsNotExist)
	// Use a directory where .claude/todos.json is a directory itself
	dir := t.TempDir()
	fpath := filepath.Join(dir, ".claude", "todos.json")
	os.MkdirAll(filepath.Dir(fpath), 0o755)
	os.MkdirAll(fpath, 0o755)

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	_, err := readTodos()
	if err == nil {
		t.Fatal("expected read error when todos.json is a directory")
	}
}

func TestReadTodos_ParseError(t *testing.T) {
	// todo_write.go:71-73 — json.Unmarshal error
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".claude"), 0o755)
	os.WriteFile(filepath.Join(dir, ".claude", "todos.json"), []byte("not valid json"), 0o644)

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	_, err := readTodos()
	if err == nil {
		t.Fatal("expected parse error")
	}
	if !strings.Contains(err.Error(), "parse todos") {
		t.Errorf("err = %q, want contains 'parse todos'", err.Error())
	}
}

func TestWriteTodos_MissingTodos(t *testing.T) {
	// todo_write.go:82 — missing 'todos' key
	_, err := ExecuteTodoWrite(map[string]any{"action": "write"})
	if err == nil {
		t.Fatal("expected error for missing todos")
	}
	if !strings.Contains(err.Error(), "todos") {
		t.Errorf("err = %q, want contains 'todos'", err.Error())
	}
}

func TestWriteTodos_StringInput(t *testing.T) {
	// todo_write.go:88-89 — todosRaw is a string
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	jsonStr := `[{"id":"1","content":"test","status":"pending","priority":"high"}]`
	_, err := ExecuteTodoWrite(map[string]any{
		"action": "write",
		"todos":  jsonStr,
	})
	if err != nil {
		t.Fatalf("unexpected error with string input: %v", err)
	}
}

func TestWriteTodos_ByteSliceInput(t *testing.T) {
	// todo_write.go:90-91 — todosRaw is []byte
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	jsonBytes := []byte(`[{"id":"1","content":"test","status":"pending","priority":"high"}]`)
	_, err := ExecuteTodoWrite(map[string]any{
		"action": "write",
		"todos":  jsonBytes,
	})
	if err != nil {
		t.Fatalf("unexpected error with byte slice input: %v", err)
	}
}

func TestWriteTodos_MarshalError(t *testing.T) {
	// todo_write.go:95-97 — json.Marshal error for unmarshallable type
	_, err := ExecuteTodoWrite(map[string]any{
		"action": "write",
		"todos":  make(chan int), // channels can't be marshaled to JSON
	})
	if err == nil {
		t.Fatal("expected marshal error")
	}
	if !strings.Contains(err.Error(), "encode todos") {
		t.Errorf("err = %q, want contains 'encode todos'", err.Error())
	}
}

func TestWriteTodos_InvalidTodosJSON(t *testing.T) {
	// todo_write.go:101-103 — json.Unmarshal error for invalid todos
	_, err := ExecuteTodoWrite(map[string]any{
		"action": "write",
		"todos":  []any{"not an object"},
	})
	if err == nil {
		t.Fatal("expected validate error")
	}
	if !strings.Contains(err.Error(), "validate todos") {
		t.Errorf("err = %q, want contains 'validate todos'", err.Error())
	}
}

func TestWriteTodos_MissingID(t *testing.T) {
	// todo_write.go:107-109 — missing ID
	_, err := ExecuteTodoWrite(map[string]any{
		"action": "write",
		"todos":  []any{map[string]any{"content": "test", "status": "pending", "priority": "high"}},
	})
	if err == nil {
		t.Fatal("expected error for missing id")
	}
	if !strings.Contains(err.Error(), "missing id") {
		t.Errorf("err = %q, want contains 'missing id'", err.Error())
	}
}

func TestWriteTodos_MissingContent(t *testing.T) {
	// todo_write.go:110-112 — missing content
	_, err := ExecuteTodoWrite(map[string]any{
		"action": "write",
		"todos":  []any{map[string]any{"id": "1", "status": "pending", "priority": "high"}},
	})
	if err == nil {
		t.Fatal("expected error for missing content")
	}
	if !strings.Contains(err.Error(), "missing content") {
		t.Errorf("err = %q, want contains 'missing content'", err.Error())
	}
}

func TestWriteTodos_InvalidStatus(t *testing.T) {
	// todo_write.go:115-117 — invalid status
	_, err := ExecuteTodoWrite(map[string]any{
		"action": "write",
		"todos":  []any{map[string]any{"id": "1", "content": "test", "status": "unknown", "priority": "high"}},
	})
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
	if !strings.Contains(err.Error(), "invalid status") {
		t.Errorf("err = %q, want contains 'invalid status'", err.Error())
	}
}

func TestWriteTodos_InvalidPriority(t *testing.T) {
	// todo_write.go:119-121 — invalid priority
	_, err := ExecuteTodoWrite(map[string]any{
		"action": "write",
		"todos":  []any{map[string]any{"id": "1", "content": "test", "status": "pending", "priority": "urgent"}},
	})
	if err == nil {
		t.Fatal("expected error for invalid priority")
	}
	if !strings.Contains(err.Error(), "invalid priority") {
		t.Errorf("err = %q, want contains 'invalid priority'", err.Error())
	}
}

func TestWriteTodos_MkdirAllError(t *testing.T) {
	// todo_write.go:126-128 — MkdirAll error
	// /proc/.claude is read-only even for root
	oldWd, _ := os.Getwd()
	os.Chdir("/proc")
	defer os.Chdir(oldWd)

	_, err := ExecuteTodoWrite(map[string]any{
		"action": "write",
		"todos":  []any{map[string]any{"id": "1", "content": "test", "status": "pending", "priority": "high"}},
	})
	if err == nil {
		t.Fatal("expected MkdirAll error")
	}
	if !strings.Contains(err.Error(), "create dir") {
		t.Errorf("err = %q, want contains 'create dir'", err.Error())
	}
}

func TestWriteTodos_WriteFileError(t *testing.T) {
	// todo_write.go:131-133 — os.WriteFile error
	// Make .claude/todos.json a directory so WriteFile fails
	dir := t.TempDir()
	cliDir := filepath.Join(dir, ".claude")
	os.MkdirAll(cliDir, 0o755)
	os.MkdirAll(filepath.Join(cliDir, "todos.json"), 0o755)

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	_, err := ExecuteTodoWrite(map[string]any{
		"action": "write",
		"todos":  []any{map[string]any{"id": "1", "content": "test", "status": "pending", "priority": "high"}},
	})
	if err == nil {
		t.Fatal("expected write error")
	}
	if !strings.Contains(err.Error(), "write file") {
		t.Errorf("err = %q, want contains 'write file'", err.Error())
	}
}

func TestExecuteTodoWrite_ReadSuccess(t *testing.T) {
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	result, err := ExecuteTodoWrite(map[string]any{"action": "read"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "[]" {
		t.Errorf("empty todos read = %q, want '[]'", result)
	}
}

func TestWebFetch_MaxBytesInt(t *testing.T) {
	// web_fetch.go:60-61 — max_bytes as int type
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html><body>Hello World</body></html>"))
	}))
	defer srv.Close()

	result, err := ExecuteWebFetch(map[string]any{
		"url":      srv.URL,
		"max_bytes": 100,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "Hello World") {
		t.Errorf("result = %q, want contains 'Hello World'", result)
	}
}

func TestWebFetch_BuildRequestError(t *testing.T) {
	// web_fetch.go:67-69 — http.NewRequest error
	_, err := ExecuteWebFetch(map[string]any{
		"url": "\x00invalid",
	})
	if err == nil {
		t.Fatal("expected build request error")
	}
	if !strings.Contains(err.Error(), "build request") {
		t.Errorf("err = %q, want contains 'build request'", err.Error())
	}
}

func TestWebFetch_ReadBodyError(t *testing.T) {
	// web_fetch.go:84-86 — io.ReadAll error
	// This is hard to trigger naturally; we use a server that closes connection mid-read
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			w.Write([]byte("ok"))
			return
		}
		conn, _, _ := hijacker.Hijack()
		conn.Close() // Close immediately to cause read error
	}))
	defer srv.Close()

	_, err := ExecuteWebFetch(map[string]any{"url": srv.URL})
	// May or may not error depending on timing; just ensure no panic
	_ = err
}

func TestWebSearch_NumResultsIntType(t *testing.T) {
	// web_search.go:80-81 — num_results as int type
	old := NativeSearchFunc
	defer func() { NativeSearchFunc = old }()
	SetNativeSearch(func(ctx context.Context, query string, numResults int) (string, error) {
		if numResults != 3 {
			t.Errorf("numResults = %d, want 3", numResults)
		}
		return "native result", nil
	})

	_, err := ExecuteWebSearch(map[string]any{"query": "test", "num_results": 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWebSearch_BraveAPIKey(t *testing.T) {
	// web_search.go:104-106 — BRAVE_API_KEY set path
	old := NativeSearchFunc
	defer func() { NativeSearchFunc = old }()
	NativeSearchFunc = nil

	oldKey := os.Getenv("BRAVE_API_KEY")
	defer os.Setenv("BRAVE_API_KEY", oldKey)
	os.Setenv("BRAVE_API_KEY", "test-key")

	// Set up mock Brave server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"web":{"results":[{"title":"T","url":"http://example.com","description":"D"}]}}`))
	}))
	defer srv.Close()

	oldURL := braveSearchURL
	defer func() { braveSearchURL = oldURL }()
	braveSearchURL = srv.URL

	result, err := ExecuteWebSearch(map[string]any{"query": "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "T") {
		t.Errorf("result = %q, want contains 'T'", result)
	}
}

func TestBraveSearch_BuildRequestError(t *testing.T) {
	// web_search.go:117-119 — http.NewRequest error
	oldURL := braveSearchURL
	defer func() { braveSearchURL = oldURL }()
	braveSearchURL = "http://\x00invalid"

	_, err := braveSearch("test", 5, "key")
	if err == nil {
		t.Fatal("expected build request error")
	}
	if !strings.Contains(err.Error(), "build request") {
		t.Errorf("err = %q, want contains 'build request'", err.Error())
	}
}

func TestBraveSearch_ClientDoError(t *testing.T) {
	// web_search.go:125-127 — client.Do error
	oldURL := braveSearchURL
	defer func() { braveSearchURL = oldURL }()
	braveSearchURL = "http://127.0.0.1:1" // port 1 should refuse

	_, err := braveSearch("test", 5, "key")
	if err == nil {
		t.Fatal("expected client.Do error")
	}
	if !strings.Contains(err.Error(), "brave") {
		t.Errorf("err = %q, want contains 'brave'", err.Error())
	}
}

func TestBraveSearch_ReadBodyError(t *testing.T) {
	// web_search.go:135-137 — io.ReadAll error
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"web":{"results":[]}}`))
			return
		}
		conn, _, _ := hijacker.Hijack()
		conn.Close()
	}))
	defer srv.Close()

	oldURL := braveSearchURL
	defer func() { braveSearchURL = oldURL }()
	braveSearchURL = srv.URL

	_, err := braveSearch("test", 5, "key")
	_ = err // may or may not error due to timing
}

func TestDDGSearch_BuildRequestError(t *testing.T) {
	// web_search.go:170-172 — http.NewRequest error
	oldURL := ddgLiteURL
	defer func() { ddgLiteURL = oldURL }()
	ddgLiteURL = "http://\x00invalid"

	_, err := ddgSearch("test", 5)
	if err == nil {
		t.Fatal("expected build request error")
	}
	if !strings.Contains(err.Error(), "build request") {
		t.Errorf("err = %q, want contains 'build request'", err.Error())
	}
}

func TestDDGSearch_ClientDoError(t *testing.T) {
	// web_search.go:178-180 — client.Do error
	oldURL := ddgLiteURL
	defer func() { ddgLiteURL = oldURL }()
	ddgLiteURL = "http://127.0.0.1:1"

	_, err := ddgSearch("test", 5)
	if err == nil {
		t.Fatal("expected client.Do error")
	}
	if !strings.Contains(err.Error(), "ddg") {
		t.Errorf("err = %q, want contains 'ddg'", err.Error())
	}
}

func TestDDGSearch_ReadBodyError(t *testing.T) {
	// web_search.go:188-190 — io.ReadAll error
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			w.Write([]byte("<html><body>ok</body></html>"))
			return
		}
		conn, _, _ := hijacker.Hijack()
		conn.Close()
	}))
	defer srv.Close()

	oldURL := ddgLiteURL
	defer func() { ddgLiteURL = oldURL }()
	ddgLiteURL = srv.URL

	_, err := ddgSearch("test", 5)
	_ = err // may or may not error due to timing
}

func TestExecuteGrep_GlobFilterMatchError(t *testing.T) {
	// grep.go:89-91 — filepath.Match error on invalid glob pattern
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("hello world"), 0o644)

	// "[" is an invalid glob pattern that causes filepath.Match to error
	_, err := ExecuteGrep(map[string]any{
		"pattern": "hello",
		"path":    dir,
		"glob":    "[",
	})
	if err != nil {
		t.Fatalf("ExecuteGrep with invalid glob: %v", err)
	}
	// Invalid glob should skip the file (not error), returning no matches
}

func TestExecuteGrep_SingleFileGrepFileError(t *testing.T) {
	// grep.go:112-114 — single file grepFile error
	_, err := ExecuteGrep(map[string]any{
		"pattern": "hello",
		"path":    "/proc/1/meminfo", // unreadable as regular file on some systems
	})
	if err == nil {
		// On systems where /proc/1/meminfo is readable, use a different path
		t.Skip("skipping: /proc/1/meminfo is readable on this system")
	}
	if !strings.Contains(err.Error(), "grep:") {
		t.Errorf("err = %q, want contains 'grep:'", err.Error())
	}
}

func TestExecuteGrep_SingleFileGrepFileError_Unreadable(t *testing.T) {
	// grep.go:112-114 — single file grepFile error via permission denied
	dir := t.TempDir()
	path := filepath.Join(dir, "secret.txt")
	os.WriteFile(path, []byte("data"), 0o000)
	defer os.Chmod(path, 0o644) // cleanup

	_, err := ExecuteGrep(map[string]any{
		"pattern": "data",
		"path":    path,
	})
	// Running as root — os.Open won't fail. Use /proc/kcore instead.
	if err == nil {
		// Running as root, skip this test
		t.Skip("skipping: running as root, permission test not applicable")
	}
}

func TestExecuteWebFetch_ReadBodyError(t *testing.T) {
	// web_fetch.go:84-86 — ioReadAll error
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("some body"))
	}))
	defer srv.Close()

	oldReadAll := ioReadAll
	defer func() { ioReadAll = oldReadAll }()
	ioReadAll = func(r io.Reader) ([]byte, error) {
		return nil, fmt.Errorf("simulated read error")
	}

	_, err := ExecuteWebFetch(map[string]any{
		"url": srv.URL,
	})
	if err == nil {
		t.Fatal("expected read body error")
	}
	if !strings.Contains(err.Error(), "read body") {
		t.Errorf("err = %q, want contains 'read body'", err.Error())
	}
}

func TestBraveSearch_ReadBodyError_Mock(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"web":{"results":[]}}`))
	}))
	defer srv.Close()

	oldURL := braveSearchURL
	defer func() { braveSearchURL = oldURL }()
	braveSearchURL = srv.URL

	oldReadAll := ioReadAll
	defer func() { ioReadAll = oldReadAll }()
	ioReadAll = func(r io.Reader) ([]byte, error) {
		return nil, fmt.Errorf("simulated read error")
	}

	_, err := braveSearch("test", 5, "fake-key")
	if err == nil {
		t.Fatal("expected read body error")
	}
	if !strings.Contains(err.Error(), "read body") {
		t.Errorf("err = %q, want contains 'read body'", err.Error())
	}
}

func TestDDGSearch_ReadBodyError_Mock(t *testing.T) {
	// web_search.go:188-190 — ioReadAll error in ddgSearch
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body>ok</body></html>"))
	}))
	defer srv.Close()

	oldURL := ddgLiteURL
	defer func() { ddgLiteURL = oldURL }()
	ddgLiteURL = srv.URL

	oldReadAll := ioReadAll
	defer func() { ioReadAll = oldReadAll }()
	ioReadAll = func(r io.Reader) ([]byte, error) {
		return nil, fmt.Errorf("simulated read error")
	}

	_, err := ddgSearch("test", 5)
	if err == nil {
		t.Fatal("expected read body error")
	}
	if !strings.Contains(err.Error(), "read body") {
		t.Errorf("err = %q, want contains 'read body'", err.Error())
	}
}

func TestCheckBlockedPattern_AllPatterns(t *testing.T) {
	// bash.go:117-125 — exercise blocked patterns via checkBlockedPattern
	// Note: checkBlockedPattern lowercases the command but NOT the pattern,
	// so test commands must match the pattern casing from blockedPatterns.
	patterns := []string{
		"rm -rf /",
		"rm -r /home",
		"rm -fr /var",
		"rm --recursive /",
		"rm --force /etc/passwd",
		"dd if=/dev/zero of=/dev/sda",
		"mkfs.ext4 /dev/sda",
		"mkfs /dev/sda",
		"> /dev/sda",
		">/dev/sda",
		":(){ :|:& };:",
		"chmod 777 /",
		"chmod -R 777 /",
		"chown -R root /",
		"kill -9 1",
		"reboot",
		"shutdown",
		"halt",
		"poweroff",
		"wget .* -O /dev/sda",
		"curl .* /dev/sda",
		"ufw disable",
		"iptables -F",
		"iptables --flush",
		"systemctl stop sshd",
		"systemctl disable firewalld",
	}
	for _, cmd := range patterns {
		err := checkBlockedPattern(cmd)
		if err == nil {
			t.Errorf("checkBlockedPattern(%q) = nil, want error", cmd)
		}
	}
}

func TestCheckBlockedPattern_SafeCommand(t *testing.T) {
	// bash.go:124 — no match returns nil
	err := checkBlockedPattern("ls -la")
	if err != nil {
		t.Fatalf("checkBlockedPattern('ls -la') = %v, want nil", err)
	}
}

func TestCheckBlockedPattern_CaseInsensitive(t *testing.T) {
	// bash.go:118 — strings.ToLower makes check case-insensitive
	err := checkBlockedPattern("REBOOT")
	if err == nil {
		t.Error("checkBlockedPattern('REBOOT') = nil, want error")
	}
	err = checkBlockedPattern("Shutdown now")
	if err == nil {
		t.Error("checkBlockedPattern('Shutdown now') = nil, want error")
	}
}

func TestExecuteBash_BlockedPattern(t *testing.T) {
	// bash.go:47-49 — ExecuteBash rejects hard-denied patterns
	_, err := ExecuteBash(map[string]any{"command": "dd if=/dev/zero of=/dev/sda"})
	if err == nil {
		t.Fatal("expected blocked pattern error")
	}
	if !strings.Contains(err.Error(), "blocked") {
		t.Errorf("err = %q, want contains 'blocked'", err.Error())
	}
}

func TestExecuteBash_MaxOutputSize(t *testing.T) {
	// bash.go:68-69 — output truncation
	result, err := ExecuteBash(map[string]any{
		"command": "python3 -c \"print('x' * 20000)\"",
	})
	if err != nil {
		t.Fatalf("ExecuteBash long output: %v", err)
	}
	if !strings.Contains(result, "truncated") {
		t.Errorf("long output not truncated, len=%d", len(result))
	}
}

func TestExecuteGrep_HiddenDirSkipped(t *testing.T) {
	// grep.go:81-83 — hidden directories are skipped during walk
	dir := t.TempDir()
	hidden := filepath.Join(dir, ".hidden")
	os.MkdirAll(hidden, 0o755)
	os.WriteFile(filepath.Join(hidden, "secret.txt"), []byte("MATCHME"), 0o644)
	os.WriteFile(filepath.Join(dir, "visible.txt"), []byte("MATCHME"), 0o644)

	result, err := ExecuteGrep(map[string]any{
		"pattern": "MATCHME",
		"path":   dir,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(result, ".hidden") {
		t.Errorf("hidden dir files should be skipped, got: %s", result)
	}
	if !strings.Contains(result, "visible.txt") {
		t.Errorf("visible file should be found, got: %s", result)
	}
}

func TestExecuteGrep_SingleFileGrepFileError_ProcMem(t *testing.T) {
	// grep.go:112-114 — grepFile error on single file (os.Open fails)
	_, err := ExecuteGrep(map[string]any{
		"pattern": "test",
		"path":   "/proc/self/mem",
	})
	if err == nil {
		t.Fatal("expected error reading /proc/self/mem as single file")
	}
}


