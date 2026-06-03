package web

import (
	"encoding/json"
	"net/http"
	"sort"
	"sync"
	"time"
)

// chatSessionMeta is the metadata stored for each chat session.
type chatSessionMeta struct {
	ID             string    `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	LastActiveAt   time.Time `json:"last_active_at"`
	MessageCount   int       `json:"message_count"`
	FirstUserMsg   string    `json:"first_user_msg,omitempty"`   // truncated preview of first user message
	LastAssistantMsg string  `json:"last_assistant_msg,omitempty"` // truncated preview of last assistant reply
}

// chatSessionStore tracks metadata for all chat WebSocket sessions.
// It is safe for concurrent use.
type chatSessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*chatSessionMeta // sessionID → metadata
	order    []string                    // insertion order (newest first in output)
}

// newChatSessionStore returns an initialized store.
func newChatSessionStore() *chatSessionStore {
	return &chatSessionStore{
		sessions: make(map[string]*chatSessionMeta),
	}
}

// create registers a new session and returns its metadata.
func (s *chatSessionStore) create(id string) *chatSessionMeta {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	m := &chatSessionMeta{
		ID:           id,
		CreatedAt:    now,
		LastActiveAt: now,
	}
	s.sessions[id] = m
	s.order = append(s.order, id)
	return m
}

// recordUserMessage records that a user message was sent, updating
// last_active and optionally storing the message text as the preview.
func (s *chatSessionStore) recordUserMessage(id, text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.sessions[id]
	if !ok {
		return
	}
	m.LastActiveAt = time.Now().UTC()
	m.MessageCount++
	// Only set the first user message preview once.
	if m.FirstUserMsg == "" && text != "" {
		m.FirstUserMsg = truncateText(text, 80)
	}
}

// recordAssistantReply records that an assistant reply was completed,
// storing the last reply text as preview.
func (s *chatSessionStore) recordAssistantReply(id, text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.sessions[id]
	if !ok {
		return
	}
	m.LastActiveAt = time.Now().UTC()
	if text != "" {
		m.LastAssistantMsg = truncateText(text, 120)
	}
}

// list returns all sessions sorted newest-first.
func (s *chatSessionStore) list() []*chatSessionMeta {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*chatSessionMeta, 0, len(s.sessions))
	// Build from order in reverse so newest appears first.
	for i := len(s.order) - 1; i >= 0; i-- {
		id := s.order[i]
		if m, ok := s.sessions[id]; ok {
			result = append(result, m)
		}
	}
	// Also sort by LastActiveAt descending for safety.
	sort.Slice(result, func(i, j int) bool {
		return result[i].LastActiveAt.After(result[j].LastActiveAt)
	})
	return result
}

// handleSessions is the HTTP handler for GET /api/sessions.
func (s *chatSessionStore) handleSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	list := s.list()
	if list == nil {
		list = []*chatSessionMeta{} // always return a JSON array, never null
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	if err := json.NewEncoder(w).Encode(list); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// truncateText returns at most maxLen runes of s, appending "…" if
// truncated. Keeps byte representation under control for preview text.
func truncateText(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "\u2026"
}