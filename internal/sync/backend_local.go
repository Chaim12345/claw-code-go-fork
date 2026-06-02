package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// LocalBackend implements SyncBackend using local filesystem
type LocalBackend struct {
	mu      sync.RWMutex
	baseDir string
}

// NewLocalBackend creates a new local filesystem backend
func NewLocalBackend(baseDir string) (*LocalBackend, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backend directory: %w", err)
	}

	return &LocalBackend{
		baseDir: baseDir,
	}, nil
}

// Upload uploads a session to local storage
func (b *LocalBackend) Upload(ctx context.Context, sessionID string, data []byte, meta SessionMeta) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Write session data
	dataPath := filepath.Join(b.baseDir, sessionID+".json")
	if err := os.WriteFile(dataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write session data: %w", err)
	}

	// Write metadata
	metaPath := filepath.Join(b.baseDir, sessionID+".meta.json")
	metaData, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metaPath, metaData, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

// Download downloads a session from local storage
func (b *LocalBackend) Download(ctx context.Context, sessionID string) ([]byte, SessionMeta, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Read session data
	dataPath := filepath.Join(b.baseDir, sessionID+".json")
	data, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, SessionMeta{}, fmt.Errorf("failed to read session data: %w", err)
	}

	// Read metadata
	metaPath := filepath.Join(b.baseDir, sessionID+".meta.json")
	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, SessionMeta{}, fmt.Errorf("failed to read metadata: %w", err)
	}

	var meta SessionMeta
	if err := json.Unmarshal(metaData, &meta); err != nil {
		return nil, SessionMeta{}, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return data, meta, nil
}

// List lists all sessions in local storage
func (b *LocalBackend) List(ctx context.Context) ([]SessionMeta, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	entries, err := os.ReadDir(b.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var metas []SessionMeta
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process .meta.json files
		name := entry.Name()
		if len(name) < 10 || name[len(name)-10:] != ".meta.json" {
			continue
		}

		metaPath := filepath.Join(b.baseDir, name)
		metaData, err := os.ReadFile(metaPath)
		if err != nil {
			continue
		}

		var meta SessionMeta
		if err := json.Unmarshal(metaData, &meta); err != nil {
			continue
		}

		metas = append(metas, meta)
	}

	return metas, nil
}

// Delete deletes a session from local storage
func (b *LocalBackend) Delete(ctx context.Context, sessionID string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Remove data file
	dataPath := filepath.Join(b.baseDir, sessionID+".json")
	if err := os.Remove(dataPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove data file: %w", err)
	}

	// Remove metadata file
	metaPath := filepath.Join(b.baseDir, sessionID+".meta.json")
	if err := os.Remove(metaPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove metadata file: %w", err)
	}

	return nil
}

// GetMeta retrieves metadata for a session
func (b *LocalBackend) GetMeta(ctx context.Context, sessionID string) (SessionMeta, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	metaPath := filepath.Join(b.baseDir, sessionID+".meta.json")
	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		return SessionMeta{}, fmt.Errorf("failed to read metadata: %w", err)
	}

	var meta SessionMeta
	if err := json.Unmarshal(metaData, &meta); err != nil {
		return SessionMeta{}, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return meta, nil
}

// Exists checks if a session exists in local storage
func (b *LocalBackend) Exists(ctx context.Context, sessionID string) (bool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	dataPath := filepath.Join(b.baseDir, sessionID+".json")
	_, err := os.Stat(dataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// Name returns the backend name
func (b *LocalBackend) Name() string {
	return "local"
}
