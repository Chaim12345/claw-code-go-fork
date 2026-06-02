package sync

import (
	"context"
	"time"
)

// SessionMeta contains metadata about a synced session
type SessionMeta struct {
	ID                string    `json:"id"`
	Version           int64     `json:"version"`
	LastModified      time.Time `json:"last_modified"`
	DeviceID          string    `json:"device_id"`
	MessageCount      int       `json:"message_count"`
	TotalInputTokens  int       `json:"total_input_tokens"`
	TotalOutputTokens int       `json:"total_output_tokens"`
	Checksum          string    `json:"checksum"` // SHA256 of session content
	Size              int64     `json:"size"`     // Size in bytes
}

// SyncStatus represents the synchronization state
type SyncStatus int

const (
	SyncStatusUnknown  SyncStatus = iota
	SyncStatusLocal               // Only exists locally
	SyncStatusRemote              // Only exists remotely
	SyncStatusSynced              // Local and remote are in sync
	SyncStatusConflict            // Local and remote have diverged
	SyncStatusPending             // Sync operation in progress
)

func (s SyncStatus) String() string {
	switch s {
	case SyncStatusLocal:
		return "local"
	case SyncStatusRemote:
		return "remote"
	case SyncStatusSynced:
		return "synced"
	case SyncStatusConflict:
		return "conflict"
	case SyncStatusPending:
		return "pending"
	default:
		return "unknown"
	}
}

// SyncBackend defines the interface for session storage backends
type SyncBackend interface {
	// Upload uploads a session to the backend
	Upload(ctx context.Context, sessionID string, data []byte, meta SessionMeta) error

	// Download downloads a session from the backend
	Download(ctx context.Context, sessionID string) ([]byte, SessionMeta, error)

	// List lists all sessions in the backend
	List(ctx context.Context) ([]SessionMeta, error)

	// Delete deletes a session from the backend
	Delete(ctx context.Context, sessionID string) error

	// GetMeta retrieves metadata for a session without downloading content
	GetMeta(ctx context.Context, sessionID string) (SessionMeta, error)

	// Exists checks if a session exists in the backend
	Exists(ctx context.Context, sessionID string) (bool, error)

	// Name returns the backend name (e.g., "s3", "gcs", "local")
	Name() string
}

// ConflictResolutionStrategy defines how to handle sync conflicts
type ConflictResolutionStrategy int

const (
	// StrategyNewest keeps the most recently modified version
	StrategyNewest ConflictResolutionStrategy = iota

	// StrategyLocal always keeps the local version
	StrategyLocal

	// StrategyRemote always keeps the remote version
	StrategyRemote

	// StrategyMerge attempts to merge both versions (not implemented yet)
	StrategyMerge

	// StrategyAsk prompts the user to choose
	StrategyAsk
)

func (s ConflictResolutionStrategy) String() string {
	switch s {
	case StrategyNewest:
		return "newest"
	case StrategyLocal:
		return "local"
	case StrategyRemote:
		return "remote"
	case StrategyMerge:
		return "merge"
	case StrategyAsk:
		return "ask"
	default:
		return "unknown"
	}
}

// ConflictResolver handles session conflicts
type ConflictResolver interface {
	// Resolve resolves a conflict between local and remote sessions
	Resolve(ctx context.Context, local, remote SessionData) (SessionData, error)
}

// SessionData represents a complete session with metadata
type SessionData struct {
	Meta    SessionMeta
	Content []byte
}

// SyncConfig contains configuration for the sync service
type SyncConfig struct {
	// Backend to use for syncing
	Backend SyncBackend

	// ConflictStrategy defines how to handle conflicts
	ConflictStrategy ConflictResolutionStrategy

	// AutoSync enables automatic background syncing
	AutoSync bool

	// SyncInterval is the interval for automatic syncing (if AutoSync is true)
	SyncInterval time.Duration

	// DeviceID uniquely identifies this device
	DeviceID string

	// CacheDir is the local directory for caching synced sessions
	CacheDir string

	// MaxCacheSize is the maximum size of the local cache in bytes
	MaxCacheSize int64

	// RetryAttempts is the number of times to retry failed sync operations
	RetryAttempts int

	// RetryDelay is the delay between retry attempts
	RetryDelay time.Duration
}

// DefaultSyncConfig returns a default sync configuration
func DefaultSyncConfig() *SyncConfig {
	return &SyncConfig{
		ConflictStrategy: StrategyNewest,
		AutoSync:         false,
		SyncInterval:     5 * time.Minute,
		DeviceID:         generateDeviceID(),
		CacheDir:         ".claude/sync-cache",
		MaxCacheSize:     100 * 1024 * 1024, // 100 MB
		RetryAttempts:    3,
		RetryDelay:       2 * time.Second,
	}
}

// generateDeviceID generates a unique device identifier
func generateDeviceID() string {
	// Simple implementation - in production, use a more robust method
	// that persists across restarts
	return time.Now().Format("20060102150405")
}

// SyncError represents a sync operation error
type SyncError struct {
	Op      string // Operation that failed (e.g., "upload", "download")
	Session string // Session ID
	Err     error  // Underlying error
}

func (e *SyncError) Error() string {
	return "sync " + e.Op + " failed for session " + e.Session + ": " + e.Err.Error()
}

func (e *SyncError) Unwrap() error {
	return e.Err
}

// ConflictError represents a sync conflict
type ConflictError struct {
	SessionID string
	Local     SessionMeta
	Remote    SessionMeta
}

func (e *ConflictError) Error() string {
	return "sync conflict for session " + e.SessionID +
		": local version " + e.Local.LastModified.Format(time.RFC3339) +
		" vs remote version " + e.Remote.LastModified.Format(time.RFC3339)
}
