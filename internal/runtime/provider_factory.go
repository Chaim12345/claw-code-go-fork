package runtime

import (
	"claw-code-go/internal/api"
	anthropicprovider "claw-code-go/internal/api/providers/anthropic"
	bedrockprovider "claw-code-go/internal/api/providers/bedrock"
	deepseekprovider "claw-code-go/internal/api/providers/deepseek"
	foundryprovider "claw-code-go/internal/api/providers/foundry"
	openaiprovider "claw-code-go/internal/api/providers/openai"
	vertexprovider "claw-code-go/internal/api/providers/vertex"
	"context"
	"fmt"
)

// SelectProvider returns the Provider for the given name.
// Supported names: "anthropic" (default), "openai", "bedrock", "vertex", "foundry", "deepseek".
func SelectProvider(name string) api.Provider {
	switch name {
	case "openai":
		return openaiprovider.New()
	case "bedrock":
		return bedrockprovider.New()
	case "vertex":
		return vertexprovider.New()
	case "foundry":
		return foundryprovider.New()
	case "deepseek":
		return deepseekprovider.New()
	default:
		return anthropicprovider.New()
	}
}

// NewProviderClient creates an API client for the provider named in cfg.ProviderName.
// Returns an error for stub providers that are not yet implemented.
func NewProviderClient(cfg *Config) (api.APIClient, error) {
	provider := SelectProvider(cfg.ProviderName)
	return provider.NewClient(api.ProviderConfig{
		APIKey:     cfg.APIKey,
		OAuthToken: cfg.OAuthToken,
		BaseURL:    cfg.BaseURL,
		Model:      cfg.Model,
		MaxTokens:  cfg.MaxTokens,
		DeltaMode:  cfg.DeltaMode,
	})
}

// NewCompactClient returns a client suitable for compaction (summarization) calls.
// When the primary provider is DeepSeek with a low-cap model (e.g. Expert at 40k
// tokens), compaction itself may hit the same pre-flight limit. This helper
// returns a client with the "instant" model (655k context) so compaction can
// always proceed regardless of how large the session is. For other providers it
// returns the original client unchanged.
func NewCompactClient(cfg *Config, original api.APIClient) api.APIClient {
	if cfg.ProviderName != "deepseek" {
		return original
	}
	provider := SelectProvider("deepseek")
	// Use a generous output budget for compaction summaries (the
	// request body itself uses MaxTokens: 2048, but the client-level
	// budget is a ceiling that the provider enforces). Use 4096 so
	// the compact client isn't artificially capped at a low value.
	compactClient, err := provider.NewClient(api.ProviderConfig{
		APIKey:     cfg.APIKey,
		OAuthToken: cfg.OAuthToken,
		BaseURL:    cfg.BaseURL,
		Model:      "instant",
		MaxTokens:  4096,
	})
	if err != nil {
		return original
	}
	return compactClient
}

// ----- NoAuthClient ----------------------------------------------------------

// NoAuthClient is a placeholder APIClient used when no credentials are configured.
// Every call returns a friendly error directing the user to /login.
type NoAuthClient struct{}

// NewNoAuthClient returns a NoAuthClient as an api.APIClient interface value.
func NewNoAuthClient() api.APIClient {
	return &NoAuthClient{}
}

func (c *NoAuthClient) StreamResponse(_ context.Context, _ api.CreateMessageRequest) (<-chan api.StreamEvent, error) {
	return nil, fmt.Errorf("not authenticated — type /login to connect to an AI provider")
}

// MaxInputTokens returns 0 (unknown) since no auth means no model.
func (c *NoAuthClient) MaxInputTokens() int { return 0 }
