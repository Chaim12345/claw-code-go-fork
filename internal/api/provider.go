package api

import "context"

// AuthMethod describes how a provider authenticates.
type AuthMethod string

const (
	AuthMethodOAuth         AuthMethod = "oauth"
	AuthMethodAPIKey        AuthMethod = "api_key"
	AuthMethodIAM           AuthMethod = "iam"            // AWS IAM (Bedrock)
	AuthMethodADC           AuthMethod = "adc"            // GCP Application Default Credentials (Vertex)
	AuthMethodAzureIdentity AuthMethod = "azure_identity" // Azure Managed Identity (Foundry)
)

// ProviderConfig holds the credentials and settings needed to create a provider client.
type ProviderConfig struct {
	APIKey     string // API key (Anthropic direct, Azure Foundry)
	OAuthToken string // OAuth 2.0 access token
	BaseURL    string // Override base URL (empty = provider default)
	Model      string // Model ID in the provider's native format
	MaxTokens  int
	DeltaMode  bool // Send only new messages instead of full history (DeepSeek only)
}

// APIClient is the interface all provider clients must implement.
type APIClient interface {
	StreamResponse(ctx context.Context, req CreateMessageRequest) (<-chan StreamEvent, error)
	// MaxInputTokens returns the provider/model's maximum input token limit,
	// or 0 if unknown. Used for compaction threshold calculations.
	MaxInputTokens() int
}

// SessionResetter is an optional interface that providers can implement
// when they maintain server-side conversation state that must be cleared
// after a client-side compaction. For example, DeepSeek's delta mode uses
// parent_message_id chaining — the server tracks the full conversation,
// so when the client compacts its message list, the server-side state
// must be reset to avoid silently overflowing the model's context window
// with stale history.
type SessionResetter interface {
	ResetSession()
}

// SessionStateProvider is an optional interface for providers that
// maintain server-side conversation state (e.g. DeepSeek's
// chat_session_id / parent_message_id) that should be persisted
// across program restarts. When a session is loaded from disk, the
// conversation loop calls RestoreSessionState to rehydrate the
// provider's in-memory state. Before saving, it calls
// MarshalSessionState to capture the current state as an opaque
// JSON blob for storage in the Session struct.
type SessionStateProvider interface {
	// MarshalSessionState returns the current provider-side session
	// state as a JSON string suitable for persistence.
	MarshalSessionState() string
	// RestoreSessionState rehydrates the provider's in-memory state
	// from a previously marshalled JSON blob. Returns an error if
	// the blob is malformed or belongs to a different provider.
	RestoreSessionState(state string) error
}

// Provider is the interface all AI providers must implement.
type Provider interface {
	// Name returns the provider identifier (e.g., "anthropic", "bedrock").
	Name() string
	// NewClient creates an API client configured for this provider.
	NewClient(cfg ProviderConfig) (APIClient, error)
	// AuthMethod returns the primary authentication method used by this provider.
	AuthMethod() AuthMethod
}
