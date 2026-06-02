package sync

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Cache manages local caching of synced sessions
type Cache struct {
	mu          sync.RWMutex
	dir         string
	maxSize     int64
	currentSize int64
	index       map[string]CacheEntry
}

// CacheEntry represents a cached session
type CacheEntry struct {
	SessionID    string      `json:"session_id"`
	Meta         SessionMeta `json:"meta"`
	FilePath     string      `json:"file_path"`
	Size         int64       `json:"size"`
	CachedAt     time.Time   `json:"cached_at"`
	LastAccessed time.Time   `json:"last_accessed"`
}

// NewCache creates a new cache instance
func NewCache(dir string, maxSize int64) (*Cache, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	c := &Cache{
		dir:     dir,
		maxSize: maxSize,
		index:   make(map[string]CacheEntry),
	}

	// Load existing cache index
	if err := c.loadIndex(); err != nil {
		// If index doesn't exist or is corrupted, start fresh
		fmt.Fprintf(os.Stderr, "Warning: failed to load cache index: %v\n", err)
	}

	return c, nil
}

// Put stores a session in the cache
func (c *Cache) Put(sessionID string, data []byte, meta SessionMeta) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to evict entries to make space
	dataSize := int64(len(data))
	if c.currentSize+dataSize > c.maxSize {
		if err := c.evictLRU(dataSize); err != nil {
			return fmt.Errorf("failed to evict cache entries: %w", err)
		}
	}

	// Write session data to file
	filePath := filepath.Join(c.dir, sessionID+".json")
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	// Update index
	now := time.Now()
	entry := CacheEntry{
		SessionID:    sessionID,
		Meta:         meta,
		FilePath:     filePath,
		Size:         dataSize,
		CachedAt:     now,
		LastAccessed: now,
	}

	// Remove old entry size if updating
	if oldEntry, exists := c.index[sessionID]; exists {
		c.currentSize -= oldEntry.Size
	}

	c.index[sessionID] = entry
	c.currentSize += dataSize

	// Save index
	if err := c.saveIndex(); err != nil {
		return fmt.Errorf("failed to save cache index: %w", err)
	}

	return nil
}

// Get retrieves a session from the cache
func (c *Cache) Get(sessionID string) ([]byte, SessionMeta, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.index[sessionID]
	if !exists {
		return nil, SessionMeta{}, fmt.Errorf("session not in cache")
	}

	// Read data from file
	data, err := os.ReadFile(entry.FilePath)
	if err != nil {
		// File missing - remove from index
		delete(c.index, sessionID)
		c.currentSize -= entry.Size
		c.saveIndex()
		return nil, SessionMeta{}, fmt.Errorf("cache file missing: %w", err)
	}

	// Update last accessed time
	entry.LastAccessed = time.Now()
	c.index[sessionID] = entry
	c.saveIndex()

	return data, entry.Meta, nil
}

// Delete removes a session from the cache
func (c *Cache) Delete(sessionID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.index[sessionID]
	if !exists {
		return nil // Already deleted
	}

	// Remove file
	if err := os.Remove(entry.FilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove cache file: %w", err)
	}

	// Update index
	delete(c.index, sessionID)
	c.currentSize -= entry.Size

	if err := c.saveIndex(); err != nil {
		return fmt.Errorf("failed to save cache index: %w", err)
	}

	return nil
}

// ListSessions returns all cached session IDs
func (c *Cache) ListSessions() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	sessions := make([]string, 0, len(c.index))
	for id := range c.index {
		sessions = append(sessions, id)
	}
	return sessions
}

// GetEntry returns cache entry metadata
func (c *Cache) GetEntry(sessionID string) (CacheEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.index[sessionID]
	return entry, exists
}

// Size returns the current cache size in bytes
func (c *Cache) Size() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.currentSize
}

// Clear removes all entries from the cache
func (c *Cache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove all files
	for _, entry := range c.index {
		if err := os.Remove(entry.FilePath); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Warning: failed to remove cache file %s: %v\n", entry.FilePath, err)
		}
	}

	// Clear index
	c.index = make(map[string]CacheEntry)
	c.currentSize = 0

	return c.saveIndex()
}

// evictLRU evicts least recently used entries to make space
func (c *Cache) evictLRU(neededSpace int64) error {
	// Sort entries by last accessed time
	type entryWithID struct {
		id    string
		entry CacheEntry
	}

	entries := make([]entryWithID, 0, len(c.index))
	for id, entry := range c.index {
		entries = append(entries, entryWithID{id, entry})
	}

	// Simple bubble sort by last accessed time (oldest first)
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].entry.LastAccessed.After(entries[j].entry.LastAccessed) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// Evict entries until we have enough space
	freedSpace := int64(0)
	for _, e := range entries {
		if c.currentSize-freedSpace+neededSpace <= c.maxSize {
			break
		}

		// Remove file
		if err := os.Remove(e.entry.FilePath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove cache file: %w", err)
		}

		// Update tracking
		delete(c.index, e.id)
		freedSpace += e.entry.Size
	}

	c.currentSize -= freedSpace

	return nil
}

// loadIndex loads the cache index from disk
func (c *Cache) loadIndex() error {
	indexPath := filepath.Join(c.dir, "index.json")

	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No index yet
		}
		return err
	}

	var index map[string]CacheEntry
	if err := json.Unmarshal(data, &index); err != nil {
		return err
	}

	// Verify files exist and calculate total size
	validIndex := make(map[string]CacheEntry)
	totalSize := int64(0)

	for id, entry := range index {
		if _, err := os.Stat(entry.FilePath); err == nil {
			validIndex[id] = entry
			totalSize += entry.Size
		}
	}

	c.index = validIndex
	c.currentSize = totalSize

	return nil
}

// saveIndex saves the cache index to disk
func (c *Cache) saveIndex() error {
	indexPath := filepath.Join(c.dir, "index.json")

	data, err := json.MarshalIndent(c.index, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(indexPath, data, 0644)
}
