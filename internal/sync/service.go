package sync

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Service manages session synchronization
type Service struct {
	mu       sync.RWMutex
	config   *SyncConfig
	cache    *Cache
	resolver ConflictResolver

	// Track sync status for each session
	status map[string]SyncStatus

	// Background sync control
	stopCh   chan struct{}
	syncDone chan struct{}
}

// NewService creates a new sync service
func NewService(config *SyncConfig) (*Service, error) {
	if config == nil {
		config = DefaultSyncConfig()
	}

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(config.CacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	cache, err := NewCache(config.CacheDir, config.MaxCacheSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache: %w", err)
	}

	resolver := NewDefaultResolver(config.ConflictStrategy)

	s := &Service{
		config:   config,
		cache:    cache,
		resolver: resolver,
		status:   make(map[string]SyncStatus),
		stopCh:   make(chan struct{}),
		syncDone: make(chan struct{}),
	}

	// Start background sync if enabled
	if config.AutoSync {
		go s.backgroundSync()
	}

	return s, nil
}

// Upload uploads a session to the remote backend
func (s *Service) Upload(ctx context.Context, sessionID string, data []byte) error {
	if s.config.Backend == nil {
		return fmt.Errorf("no backend configured")
	}

	s.mu.Lock()
	s.status[sessionID] = SyncStatusPending
	s.mu.Unlock()

	// Calculate checksum
	checksum := calculateChecksum(data)

	// Create metadata
	meta := SessionMeta{
		ID:           sessionID,
		Version:      time.Now().Unix(),
		LastModified: time.Now(),
		DeviceID:     s.config.DeviceID,
		Checksum:     checksum,
		Size:         int64(len(data)),
	}

	// Try to extract message count from session data
	if count, err := extractMessageCount(data); err == nil {
		meta.MessageCount = count
	}

	// Upload with retries
	var lastErr error
	for attempt := 0; attempt <= s.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(s.config.RetryDelay)
		}

		err := s.config.Backend.Upload(ctx, sessionID, data, meta)
		if err == nil {
			// Success - update cache and status
			if err := s.cache.Put(sessionID, data, meta); err != nil {
				// Log but don't fail - cache is optional
				fmt.Fprintf(os.Stderr, "Warning: failed to cache session: %v\n", err)
			}

			s.mu.Lock()
			s.status[sessionID] = SyncStatusSynced
			s.mu.Unlock()

			return nil
		}

		lastErr = err
	}

	s.mu.Lock()
	s.status[sessionID] = SyncStatusLocal
	s.mu.Unlock()

	return &SyncError{
		Op:      "upload",
		Session: sessionID,
		Err:     lastErr,
	}
}

// Download downloads a session from the remote backend
func (s *Service) Download(ctx context.Context, sessionID string) ([]byte, SessionMeta, error) {
	if s.config.Backend == nil {
		return nil, SessionMeta{}, fmt.Errorf("no backend configured")
	}

	// Check cache first
	if data, meta, err := s.cache.Get(sessionID); err == nil {
		// Verify cache is still valid by checking remote metadata
		if remoteMeta, err := s.config.Backend.GetMeta(ctx, sessionID); err == nil {
			if meta.Version == remoteMeta.Version && meta.Checksum == remoteMeta.Checksum {
				return data, meta, nil
			}
		}
	}

	// Download with retries
	var lastErr error
	for attempt := 0; attempt <= s.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(s.config.RetryDelay)
		}

		data, meta, err := s.config.Backend.Download(ctx, sessionID)
		if err == nil {
			// Verify checksum
			if checksum := calculateChecksum(data); checksum != meta.Checksum {
				return nil, SessionMeta{}, fmt.Errorf("checksum mismatch: expected %s, got %s", meta.Checksum, checksum)
			}

			// Update cache
			if err := s.cache.Put(sessionID, data, meta); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to cache session: %v\n", err)
			}

			s.mu.Lock()
			s.status[sessionID] = SyncStatusSynced
			s.mu.Unlock()

			return data, meta, nil
		}

		lastErr = err
	}

	return nil, SessionMeta{}, &SyncError{
		Op:      "download",
		Session: sessionID,
		Err:     lastErr,
	}
}

// Sync synchronizes a session (upload if local is newer, download if remote is newer)
func (s *Service) Sync(ctx context.Context, sessionID string, localData []byte) ([]byte, error) {
	if s.config.Backend == nil {
		return localData, fmt.Errorf("no backend configured")
	}

	// Check if session exists remotely
	exists, err := s.config.Backend.Exists(ctx, sessionID)
	if err != nil {
		return localData, fmt.Errorf("failed to check remote existence: %w", err)
	}

	if !exists {
		// Upload local session
		if err := s.Upload(ctx, sessionID, localData); err != nil {
			return localData, err
		}
		return localData, nil
	}

	// Get remote metadata
	remoteMeta, err := s.config.Backend.GetMeta(ctx, sessionID)
	if err != nil {
		return localData, fmt.Errorf("failed to get remote metadata: %w", err)
	}

	// Calculate local metadata
	localChecksum := calculateChecksum(localData)
	localMeta := SessionMeta{
		ID:           sessionID,
		Version:      time.Now().Unix(),
		LastModified: time.Now(),
		DeviceID:     s.config.DeviceID,
		Checksum:     localChecksum,
		Size:         int64(len(localData)),
	}

	// Check if already in sync
	if localChecksum == remoteMeta.Checksum {
		s.mu.Lock()
		s.status[sessionID] = SyncStatusSynced
		s.mu.Unlock()
		return localData, nil
	}

	// Conflict detected - resolve it
	remoteData, _, err := s.config.Backend.Download(ctx, sessionID)
	if err != nil {
		return localData, fmt.Errorf("failed to download remote session: %w", err)
	}

	s.mu.Lock()
	s.status[sessionID] = SyncStatusConflict
	s.mu.Unlock()

	// Resolve conflict
	resolved, err := s.resolver.Resolve(ctx,
		SessionData{Meta: localMeta, Content: localData},
		SessionData{Meta: remoteMeta, Content: remoteData},
	)
	if err != nil {
		return localData, fmt.Errorf("failed to resolve conflict: %w", err)
	}

	// Upload resolved version
	if err := s.Upload(ctx, sessionID, resolved.Content); err != nil {
		return localData, fmt.Errorf("failed to upload resolved session: %w", err)
	}

	return resolved.Content, nil
}

// List lists all sessions (local and remote)
func (s *Service) List(ctx context.Context) ([]SessionMeta, error) {
	if s.config.Backend == nil {
		return nil, fmt.Errorf("no backend configured")
	}

	return s.config.Backend.List(ctx)
}

// Delete deletes a session from both local cache and remote backend
func (s *Service) Delete(ctx context.Context, sessionID string) error {
	var errs []error

	// Delete from cache
	if err := s.cache.Delete(sessionID); err != nil {
		errs = append(errs, fmt.Errorf("cache delete: %w", err))
	}

	// Delete from backend
	if s.config.Backend != nil {
		if err := s.config.Backend.Delete(ctx, sessionID); err != nil {
			errs = append(errs, fmt.Errorf("backend delete: %w", err))
		}
	}

	s.mu.Lock()
	delete(s.status, sessionID)
	s.mu.Unlock()

	if len(errs) > 0 {
		return fmt.Errorf("delete errors: %v", errs)
	}

	return nil
}

// GetStatus returns the sync status for a session
func (s *Service) GetStatus(sessionID string) SyncStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status, ok := s.status[sessionID]
	if !ok {
		return SyncStatusUnknown
	}
	return status
}

// backgroundSync runs periodic synchronization
func (s *Service) backgroundSync() {
	ticker := time.NewTicker(s.config.SyncInterval)
	defer ticker.Stop()
	defer close(s.syncDone)

	for {
		select {
		case <-ticker.C:
			// Get all cached sessions
			sessions := s.cache.ListSessions()

			for _, sessionID := range sessions {
				// Skip if already syncing
				if s.GetStatus(sessionID) == SyncStatusPending {
					continue
				}

				// Get cached data
				data, _, err := s.cache.Get(sessionID)
				if err != nil {
					continue
				}

				// Sync in background (don't block)
				go func(id string, d []byte) {
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()

					if _, err := s.Sync(ctx, id, d); err != nil {
						fmt.Fprintf(os.Stderr, "Background sync failed for %s: %v\n", id, err)
					}
				}(sessionID, data)
			}

		case <-s.stopCh:
			return
		}
	}
}

// Stop stops the background sync goroutine
func (s *Service) Stop() {
	if s.config.AutoSync {
		close(s.stopCh)
		<-s.syncDone
	}
}

// calculateChecksum calculates SHA256 checksum of data
func calculateChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// extractMessageCount attempts to extract message count from session JSON
func extractMessageCount(data []byte) (int, error) {
	var session struct {
		Messages []interface{} `json:"messages"`
	}

	if err := json.Unmarshal(data, &session); err != nil {
		return 0, err
	}

	return len(session.Messages), nil
}
