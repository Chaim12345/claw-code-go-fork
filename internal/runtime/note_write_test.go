package runtime

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"claw-code-go/internal/session"
)

type noteStore struct {
	notes map[string][]*session.Note
	nextID int64
}

func newNoteStore() *noteStore {
	return &noteStore{notes: make(map[string][]*session.Note), nextID: 1}
}

func (s *noteStore) SaveSessionNote(sessionID, key, content string) error {
	for i, n := range s.notes[sessionID] {
		if n.Key == key {
			s.notes[sessionID][i].Content = content
			s.notes[sessionID][i].UpdatedAt = time.Now()
			return nil
		}
	}
	n := &session.Note{ID: s.nextID, SessionID: sessionID, Key: key, Content: content, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	s.nextID++
	s.notes[sessionID] = append(s.notes[sessionID], n)
	return nil
}

func (s *noteStore) GetSessionNotes(sessionID string) ([]*session.Note, error) {
	return s.notes[sessionID], nil
}

func (s *noteStore) DeleteSessionNote(id int64) error {
	for sid, notes := range s.notes {
		for i, n := range notes {
			if n.ID == id {
				s.notes[sid] = append(notes[:i], notes[i+1:]...)
				return nil
			}
		}
	}
	return nil
}

func (s *noteStore) CreateSession(_, _, _ string) error                                             { return nil }
func (s *noteStore) GetSession(_ string) (*session.Session, error)                                  { return nil, nil }
func (s *noteStore) ListSessions() ([]*session.Session, error)                                      { return nil, nil }
func (s *noteStore) GetRecentSessions(_ int) ([]*session.Session, error)                            { return nil, nil }
func (s *noteStore) GetSessionSummary(_ string) (string, error)                                     { return "", nil }
func (s *noteStore) SearchSessions(_ string) ([]*session.Session, error)                            { return nil, nil }
func (s *noteStore) FilterSessions(_, _ string, _, _ time.Time) ([]*session.Session, error)         { return nil, nil }
func (s *noteStore) DeleteSession(_ string) error                                                   { return nil }
func (s *noteStore) UpdateSessionPreview(_, _, _ string, _ int) error                               { return nil }
func (s *noteStore) RecordMessage(_, _, _ string, _ int) (int64, error)                             { return 0, nil }
func (s *noteStore) GetMessages(_ string) ([]*session.Message, error)                               { return nil, nil }
func (s *noteStore) RecordToolCall(_ int64, _, _, _ string, _ int) error                            { return nil }
func (s *noteStore) GetToolCalls(_ int64) ([]*session.ToolCall, error)                              { return nil, nil }
func (s *noteStore) ExportSessionMarkdown(_ string) (string, error)                                 { return "", nil }
func (s *noteStore) ExportSessionJSON(_ string) ([]byte, error)                                     { return nil, nil }
func (s *noteStore) ImportSessionJSON(_ []byte) (string, error)                                     { return "", nil }
func (s *noteStore) ImportSessionMarkdown(_ []byte) (string, error)                                 { return "", nil }
func (s *noteStore) Close() error                                                                   { return nil }

func TestExecuteNoteWrite_ReadEmpty(t *testing.T) {
	loop := &ConversationLoop{SessionStore: newNoteStore(), CurrentSessionID: "s1"}
	result, err := loop.executeNoteWrite(map[string]any{"action": "read"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "[]" {
		t.Errorf("expected '[]', got %q", result)
	}
}

func TestExecuteNoteWrite_WriteThenRead(t *testing.T) {
	loop := &ConversationLoop{SessionStore: newNoteStore(), CurrentSessionID: "s1"}
	result, err := loop.executeNoteWrite(map[string]any{"action": "write", "key": "pref", "content": "tabs not spaces"})
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	if result != "note saved" {
		t.Errorf("write: expected 'note saved', got %q", result)
	}

	result, err = loop.executeNoteWrite(map[string]any{"action": "read"})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(result, "pref") || !strings.Contains(result, "tabs not spaces") {
		t.Errorf("read: expected note content, got %q", result)
	}
}

func TestExecuteNoteWrite_WriteUpserts(t *testing.T) {
	loop := &ConversationLoop{SessionStore: newNoteStore(), CurrentSessionID: "s1"}
	loop.executeNoteWrite(map[string]any{"action": "write", "key": "pref", "content": "old"})
	loop.executeNoteWrite(map[string]any{"action": "write", "key": "pref", "content": "new"})
	result, _ := loop.executeNoteWrite(map[string]any{"action": "read"})
	if strings.Contains(result, `"old"`) {
		t.Errorf("upsert: old value still present, got %q", result)
	}
	if !strings.Contains(result, `"new"`) {
		t.Errorf("upsert: new value missing, got %q", result)
	}
}

func TestExecuteNoteWrite_WriteNoContent(t *testing.T) {
	loop := &ConversationLoop{SessionStore: newNoteStore(), CurrentSessionID: "s1"}
	_, err := loop.executeNoteWrite(map[string]any{"action": "write", "key": "pref"})
	if err == nil {
		t.Error("expected error for missing content")
	}
}

func TestExecuteNoteWrite_NoSession(t *testing.T) {
	loop := &ConversationLoop{SessionStore: newNoteStore()}
	_, err := loop.executeNoteWrite(map[string]any{"action": "write", "key": "k", "content": "v"})
	if err == nil {
		t.Error("expected error for no session")
	}
}

func TestExecuteNoteWrite_UnknownAction(t *testing.T) {
	loop := &ConversationLoop{SessionStore: newNoteStore(), CurrentSessionID: "s1"}
	_, err := loop.executeNoteWrite(map[string]any{"action": "delete"})
	if err == nil {
		t.Error("expected error for unknown action")
	}
}

func TestExecuteNoteWrite_ReadNilStore(t *testing.T) {
	loop := &ConversationLoop{}
	result, err := loop.executeNoteWrite(map[string]any{"action": "read"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "[]" {
		t.Errorf("expected '[]' with nil store, got %q", result)
	}
}

func TestExecuteNoteWrite_ReadReturnsValidJSON(t *testing.T) {
	store := newNoteStore()
	store.SaveSessionNote("s1", "a", "alpha")
	store.SaveSessionNote("s1", "b", "beta")
	loop := &ConversationLoop{SessionStore: store, CurrentSessionID: "s1"}
	result, err := loop.executeNoteWrite(map[string]any{"action": "read"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var parsed []map[string]any
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, result)
	}
	if len(parsed) != 2 {
		t.Errorf("expected 2 notes, got %d", len(parsed))
	}
}
