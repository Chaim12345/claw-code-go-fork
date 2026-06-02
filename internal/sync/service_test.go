package sync

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSyncService(t *testing.T) {
	// Create temporary directories
	tmpDir := t.TempDir()
	backendDir := filepath.Join(tmpDir, "backend")
	cacheDir := filepath.Join(tmpDir, "cache")

	// Create backend
	backend, err := NewLocalBackend(backendDir)
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}

	// Create config
	config := &SyncConfig{
		Backend:          backend,
		ConflictStrategy: StrategyNewest,
		AutoSync:         false,
		DeviceID:         "test-device",
		CacheDir:         cacheDir,
		MaxCacheSize:     10 * 1024 * 1024,
		RetryAttempts:    2,
		RetryDelay:       100 * time.Millisecond,
	}

	// Create service
	service, err := NewService(config)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	// Don't defer Stop() - it can cause test hangs

	ctx := context.Background()

	// Test upload
	t.Run("Upload", func(t *testing.T) {
		sessionID := "test-session-1"
		data := []byte(`{"messages": [{"role": "user", "content": "Hello"}]}`)

		err := service.Upload(ctx, sessionID, data)
		if err != nil {
			t.Fatalf("Upload failed: %v", err)
		}

		// Verify status
		status := service.GetStatus(sessionID)
		if status != SyncStatusSynced {
			t.Errorf("Expected status %v, got %v", SyncStatusSynced, status)
		}
	})

	// Test download
	t.Run("Download", func(t *testing.T) {
		sessionID := "test-session-1"

		data, meta, err := service.Download(ctx, sessionID)
		if err != nil {
			t.Fatalf("Download failed: %v", err)
		}

		if len(data) == 0 {
			t.Error("Downloaded data is empty")
		}

		if meta.ID != sessionID {
			t.Errorf("Expected session ID %s, got %s", sessionID, meta.ID)
		}
	})

	// Test sync with no conflict
	t.Run("SyncNoConflict", func(t *testing.T) {
		sessionID := "test-session-2"
		data := []byte(`{"messages": [{"role": "user", "content": "Test"}]}`)

		// Upload first
		err := service.Upload(ctx, sessionID, data)
		if err != nil {
			t.Fatalf("Upload failed: %v", err)
		}

		// Sync with same data
		result, err := service.Sync(ctx, sessionID, data)
		if err != nil {
			t.Fatalf("Sync failed: %v", err)
		}

		if string(result) != string(data) {
			t.Error("Synced data doesn't match")
		}
	})

	// Test list
	t.Run("List", func(t *testing.T) {
		sessions, err := service.List(ctx)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if len(sessions) < 2 {
			t.Errorf("Expected at least 2 sessions, got %d", len(sessions))
		}
	})

	// Test delete
	t.Run("Delete", func(t *testing.T) {
		sessionID := "test-session-1"

		err := service.Delete(ctx, sessionID)
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		// Verify it's gone
		exists, err := backend.Exists(ctx, sessionID)
		if err != nil {
			t.Fatalf("Exists check failed: %v", err)
		}

		if exists {
			t.Error("Session still exists after delete")
		}
	})
}

func TestCache(t *testing.T) {
	tmpDir := t.TempDir()

	cache, err := NewCache(tmpDir, 1024*1024)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Test put and get
	t.Run("PutGet", func(t *testing.T) {
		sessionID := "test-cache-1"
		data := []byte("test data")
		meta := SessionMeta{
			ID:           sessionID,
			Version:      1,
			LastModified: time.Now(),
			DeviceID:     "test",
			Checksum:     "abc123",
			Size:         int64(len(data)),
		}

		err := cache.Put(sessionID, data, meta)
		if err != nil {
			t.Fatalf("Put failed: %v", err)
		}

		retrieved, retrievedMeta, err := cache.Get(sessionID)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		if string(retrieved) != string(data) {
			t.Error("Retrieved data doesn't match")
		}

		if retrievedMeta.ID != sessionID {
			t.Error("Retrieved meta doesn't match")
		}
	})

	// Test list
	t.Run("List", func(t *testing.T) {
		sessions := cache.ListSessions()
		if len(sessions) != 1 {
			t.Errorf("Expected 1 session, got %d", len(sessions))
		}
	})

	// Test delete
	t.Run("Delete", func(t *testing.T) {
		sessionID := "test-cache-1"

		err := cache.Delete(sessionID)
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		_, _, err = cache.Get(sessionID)
		if err == nil {
			t.Error("Expected error getting deleted session")
		}
	})
}

func TestLocalBackend(t *testing.T) {
	tmpDir := t.TempDir()

	backend, err := NewLocalBackend(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}

	ctx := context.Background()

	// Test upload
	t.Run("Upload", func(t *testing.T) {
		sessionID := "test-backend-1"
		data := []byte("test data")
		meta := SessionMeta{
			ID:           sessionID,
			Version:      1,
			LastModified: time.Now(),
			DeviceID:     "test",
			Checksum:     "abc123",
			Size:         int64(len(data)),
		}

		err := backend.Upload(ctx, sessionID, data, meta)
		if err != nil {
			t.Fatalf("Upload failed: %v", err)
		}

		// Verify files exist
		dataPath := filepath.Join(tmpDir, sessionID+".json")
		if _, err := os.Stat(dataPath); err != nil {
			t.Error("Data file not created")
		}

		metaPath := filepath.Join(tmpDir, sessionID+".meta.json")
		if _, err := os.Stat(metaPath); err != nil {
			t.Error("Meta file not created")
		}
	})

	// Test download
	t.Run("Download", func(t *testing.T) {
		sessionID := "test-backend-1"

		data, meta, err := backend.Download(ctx, sessionID)
		if err != nil {
			t.Fatalf("Download failed: %v", err)
		}

		if string(data) != "test data" {
			t.Error("Downloaded data doesn't match")
		}

		if meta.ID != sessionID {
			t.Error("Downloaded meta doesn't match")
		}
	})

	// Test exists
	t.Run("Exists", func(t *testing.T) {
		exists, err := backend.Exists(ctx, "test-backend-1")
		if err != nil {
			t.Fatalf("Exists check failed: %v", err)
		}

		if !exists {
			t.Error("Session should exist")
		}

		exists, err = backend.Exists(ctx, "nonexistent")
		if err != nil {
			t.Fatalf("Exists check failed: %v", err)
		}

		if exists {
			t.Error("Nonexistent session should not exist")
		}
	})

	// Test list
	t.Run("List", func(t *testing.T) {
		sessions, err := backend.List(ctx)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if len(sessions) != 1 {
			t.Errorf("Expected 1 session, got %d", len(sessions))
		}
	})

	// Test delete
	t.Run("Delete", func(t *testing.T) {
		sessionID := "test-backend-1"

		err := backend.Delete(ctx, sessionID)
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		exists, err := backend.Exists(ctx, sessionID)
		if err != nil {
			t.Fatalf("Exists check failed: %v", err)
		}

		if exists {
			t.Error("Session should not exist after delete")
		}
	})
}
