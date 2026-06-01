package runtime

import (
	"claw-code-go/internal/config"
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	DefaultModel     = "claude-sonnet-4-20250514"
	DefaultMaxTokens = 8096
)

// MCPServerConfig describes a single MCP server connection.
type MCPServerConfig struct {
	Name      string            `json:"name"`
	Transport string            `json:"transport"` // "stdio" or "sse"
	Command   string            `json:"command,omitempty"`
	Args      []string          `json:"args,omitempty"`
	URL       string            `json:"url,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
}

// Config holds runtime configuration for the CLI.
type Config struct {
	Model        string
	MaxTokens    int
	SystemPrompt string
	SessionDir   string
	APIKey       string
	BaseURL      string

	// Provider and auth fields (Phase 3).
	// ProviderName is one of: "anthropic", "bedrock", "vertex", "foundry".
	ProviderName string
	// AuthMethod is one of: "api_key", "oauth", "iam", "adc", "azure_identity".
	AuthMethod string
	// OAuthToken is the resolved OAuth access token (set at startup when using OAuth).
	OAuthToken string

	// MCPServers lists MCP server connections (Phase 4).
	MCPServers []MCPServerConfig

	// Compaction settings (Phase 6).
	// CompactionEnabled enables automatic session compaction (default: true).
	CompactionEnabled bool
	// CompactionThreshold is the fraction of CompactionMaxInputTokens (or
	// MaxTokens as a fallback) at which compaction triggers (e.g., 0.75
	// triggers at 75% of the context-window basis).
	CompactionThreshold float64
	// CompactionKeepRecent is the number of most-recent messages retained
	// verbatim after compaction.
	CompactionKeepRecent int
	// CompactionMaxInputTokens overrides the basis used for the compaction
	// threshold. Providers with much larger context windows than the
	// per-turn output budget (e.g. DeepSeek's web API has a 655k-token
	// context for Instant) should populate this so we don't compact based
	// on the per-turn output budget. Zero means "fall back to MaxTokens".
	CompactionMaxInputTokens int

	// Permission settings (Phase 11).
	// PermissionMode is the active permission enforcement mode string.
	PermissionMode string
	// AllowedTools are tool names that are always allowed without prompting.
	AllowedTools []string
	// BlockedTools are tool names that are always denied without prompting.
	BlockedTools []string

	// Theme is the active TUI color theme ("dark" or "light").
	Theme string

	// Autonomous mode (RunTask). When true, RunTask auto-approves
	// every tool call, suppresses the ask_user tool so the model
	// cannot block waiting for human input, and loops SendMessage
	// until an LLM-judge returns DONE or MaxTurns is reached.
	// PermissionMode is force-set to "bypass" for the duration of
	// RunTask regardless of its own value.
	Autonomous bool
	// MaxTurns caps the number of SendMessage iterations inside
	// RunTask. Zero means use DefaultAutonomousMaxTurns.
	MaxTurns int
}

// DefaultAutonomousMaxTurns is the default cap for RunTask iterations
// when Config.MaxTurns is zero. Hard cap is enforced at
// MaxAutonomousTurns so a misconfigured Config can't run forever.
const (
	DefaultAutonomousMaxTurns = 10
	MaxAutonomousTurns        = 50
)

// LoadConfig reads configuration from layered settings files and environment
// variables and applies defaults. Load order (later overrides earlier):
//  1. Defaults
//  2. Layered settings files (user global → project → local)
//  3. Environment variables
//  4. CLI flags (applied by the caller after this function returns)
func LoadConfig() *Config {
	cfg := &Config{
		Model:                DefaultModel,
		MaxTokens:            DefaultMaxTokens,
		PermissionMode:       "default",
		CompactionEnabled:    true,
		CompactionThreshold:  DefaultCompactionThreshold,
		CompactionKeepRecent: DefaultCompactionKeepRecent,
	}

	// Apply layered settings files (user global → project → local).
	s := config.Load()
	if s.Model != "" {
		cfg.Model = s.Model
	}
	if s.MaxTokens != 0 {
		cfg.MaxTokens = s.MaxTokens
	}
	if s.PermissionMode != "" {
		cfg.PermissionMode = s.PermissionMode
	}
	if len(s.AllowedTools) > 0 {
		cfg.AllowedTools = s.AllowedTools
	}
	if len(s.BlockedTools) > 0 {
		cfg.BlockedTools = s.BlockedTools
	}
	if s.Theme != "" {
		cfg.Theme = s.Theme
	}

	// Environment variables override settings files.
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		cfg.APIKey = key
	}
	if model := os.Getenv("ANTHROPIC_MODEL"); model != "" {
		cfg.Model = model
	}
	if baseURL := os.Getenv("ANTHROPIC_BASE_URL"); baseURL != "" {
		cfg.BaseURL = baseURL
	}

	// Default session dir: ~/.claw-code/sessions
	homeDir, err := os.UserHomeDir()
	if err == nil {
		cfg.SessionDir = filepath.Join(homeDir, ".claw-code", "sessions")
	} else {
		cfg.SessionDir = ".claw-code-sessions"
	}

	// Detect the active provider from environment variables.
	cfg.ProviderName = detectProvider()

	// When the deepseek provider is in use and no model was set by
	// settings files / env / CLI flags, default to "expert" — the
	// chain-of-thought variant that the web UI highlights for
	// complex problems. The /settings endpoint reports
	// input_character_limit=163,840 for it (live-probed in
	// TestLiveProbeActualInputLimit/expert — server enforces the
	// same value). The provider's DefaultModel constant is the
	// authoritative source of truth.
	if cfg.ProviderName == "deepseek" && cfg.Model == DefaultModel {
		if v := os.Getenv("DEEPSEEK_MODEL"); v != "" {
			cfg.Model = v
		} else {
			cfg.Model = "expert"
		}
	}

	// Set the per-provider compaction basis so we don't compact at
	// 75% of an 8k output budget when the actual context window is
	// 655k tokens (DeepSeek Instant) or 200k (Anthropic Sonnet 4).
	if cfg.CompactionMaxInputTokens == 0 {
		cfg.CompactionMaxInputTokens = PerProviderCompactionMax(cfg.ProviderName, cfg.Model)
	}

	// Load MCP server configs.
	cfg.MCPServers = loadMCPServers(homeDir)

	return cfg
}

// loadMCPServers reads MCP server configurations from the settings file and
// the CLAUDE_MCP_SERVERS environment variable (JSON override, takes precedence).
func loadMCPServers(homeDir string) []MCPServerConfig {
	// Try env var override first.
	if raw := os.Getenv("CLAUDE_MCP_SERVERS"); raw != "" {
		var servers []MCPServerConfig
		if err := json.Unmarshal([]byte(raw), &servers); err == nil {
			return servers
		}
	}

	// Otherwise read from ~/.claude/settings.json.
	if homeDir == "" {
		return nil
	}
	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return nil
	}

	var settings struct {
		MCPServers []MCPServerConfig `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil
	}

	return settings.MCPServers
}

// detectProvider reads env vars to determine which provider to use.
func detectProvider() string {
	switch {
	case os.Getenv("DEEPSEEK_TOKEN") != "" || os.Getenv("CLAUDE_CODE_USE_DEEPSEEK") == "1":
		return "deepseek"
	case os.Getenv("CLAUDE_CODE_USE_BEDROCK") == "1":
		return "bedrock"
	case os.Getenv("CLAUDE_CODE_USE_VERTEX") == "1":
		return "vertex"
	case os.Getenv("CLAUDE_CODE_USE_FOUNDRY") == "1":
		return "foundry"
	default:
		return "anthropic"
	}
}

// PerProviderCompactionMax returns the provider's per-variant context-window
// size in tokens, used as the basis for the compaction threshold so we don't
// compact prematurely on providers with very large windows. Zero means
// "use the default MaxTokens basis".
//
// Values come from live probing (TestLiveProbeActualInputLimit) — the
// server actually enforces these, not just the /settings JSON.
func PerProviderCompactionMax(providerName, model string) int {
	switch providerName {
	case "deepseek":
		// Use the model's own cap so the threshold tracks what the
		// server will actually accept. /settings reports:
		//   instant:  2,621,440 chars = 655,360 tokens
		//   expert:     163,840 chars =  40,960 tokens
		//   vision:   2,621,440 chars = 655,360 tokens
		// Live-probe (probe_limit_live_test.go) confirmed both the
		// client-side pre-flight cap and the server's enforcement
		// match these values exactly.
		const charsPerToken = 4
		switch model {
		case "instant", "default":
			return 2_621_440 / charsPerToken
		case "expert":
			return 163_840 / charsPerToken
		case "vision":
			return 2_621_440 / charsPerToken
		}
		// Unknown deepseek model — fall back to the conservative
		// Expert cap (smallest of the known variants).
		return 163_840 / charsPerToken
	}
	_ = model
	return 0
}
