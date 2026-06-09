package search

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIndex_BuildAndSearch(t *testing.T) {
	dir := t.TempDir()
	writeGoFile(t, dir, "handler.go", `
package main

func HandleRequest(w http.ResponseWriter, r *http.Request) {
	data := fetchFromDB(r.URL.Query().Get("id"))
	w.Write(data)
}

func fetchFromDB(query string) []byte {
	return []byte("result")
}
`)
	writeGoFile(t, dir, "auth.go", `
package main

func Authenticate(user, pass string) bool {
	return user == "admin" && pass == "secret"
}
`)

	idx := NewIndex()
	if err := idx.Build(dir); err != nil {
		t.Fatal(err)
	}

	results := idx.Search("handle request", 5)
	if len(results) == 0 {
		t.Fatal("expected results for 'handle request', got none")
	}
	if results[0].FilePath != "handler.go" {
		t.Errorf("expected handler.go as top result, got %s", results[0].FilePath)
	}
}

func TestIndex_EmptyQuery(t *testing.T) {
	dir := t.TempDir()
	writeGoFile(t, dir, "main.go", `package main; func main() {}`)
	idx := NewIndex()
	if err := idx.Build(dir); err != nil {
		t.Fatal(err)
	}
	results := idx.Search("", 10)
	if results != nil {
		t.Errorf("expected nil for empty query, got %v", results)
	}
}

func TestIndex_NoMatch(t *testing.T) {
	dir := t.TempDir()
	writeGoFile(t, dir, "main.go", `package main; func main() {}`)
	idx := NewIndex()
	if err := idx.Build(dir); err != nil {
		t.Fatal(err)
	}
	results := idx.Search("xyzzy_nonexistent_token", 10)
	if len(results) != 0 {
		t.Errorf("expected no results for nonsense query, got %d", len(results))
	}
}

func TestIndex_SkipsTestFiles(t *testing.T) {
	dir := t.TempDir()
	writeGoFile(t, dir, "handler.go", `package main; func HandleThings() {}`)
	writeGoFile(t, dir, "handler_test.go", `package main; func TestHandleThings() {}`)
	idx := NewIndex()
	if err := idx.Build(dir); err != nil {
		t.Fatal(err)
	}
	results := idx.Search("HandleThings", 10)
	for _, r := range results {
		if filepath.Ext(r.FilePath) == ".go" && filepath.Base(r.FilePath) != "handler.go" {
			t.Errorf("test file should be skipped: %s", r.FilePath)
		}
	}
}

func TestIndex_SkipsGitDir(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	os.MkdirAll(gitDir, 0755)
	writeGoFile(t, dir, ".git/config.go", `package git; func Config() {}`)
	writeGoFile(t, dir, "main.go", `package main; func Main() {}`)
	idx := NewIndex()
	if err := idx.Build(dir); err != nil {
		t.Fatal(err)
	}
	for _, doc := range idx.docs {
		if filepath.Base(filepath.Dir(doc)) == ".git" || filepath.Base(doc) == "config.go" {
			t.Errorf(".git dir should be skipped, found: %s", doc)
		}
	}
}

func TestSearchCodebase(t *testing.T) {
	dir := t.TempDir()
	writeGoFile(t, dir, "search.go", `package search; func SearchCodebase() {}`)
	results, err := SearchCodebase(dir, "search codebase", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) == 0 {
		t.Fatal("expected results, got none")
	}
}

func TestSearchCodebase_EmptyQuery(t *testing.T) {
	dir := t.TempDir()
	writeGoFile(t, dir, "main.go", `package main; func main() {}`)
	results, err := SearchCodebase(dir, "", 5)
	if err != nil {
		t.Fatal(err)
	}
	if results != nil {
		t.Errorf("expected nil for empty query, got %v", results)
	}
}

func TestIndex_Snippet(t *testing.T) {
	dir := t.TempDir()
	writeGoFile(t, dir, "api.go", `package main

import "net/http"

func CreateHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
}
`)
	idx := NewIndex()
	if err := idx.Build(dir); err != nil {
		t.Fatal(err)
	}
	results := idx.Search("create handler", 1)
	if len(results) == 0 {
		t.Fatal("expected results")
	}
	if results[0].Snippet == "" {
		t.Error("expected non-empty snippet")
	}
}

func writeGoFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
