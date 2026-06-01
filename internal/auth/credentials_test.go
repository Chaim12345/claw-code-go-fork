package auth

import (
	"os"
	"testing"
)

func TestResolveCredentialsEnvPrecedence(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "ant-key")
	t.Setenv("OPENAI_API_KEY", "oai-key")
	t.Setenv("DEEPSEEK_TOKEN", "ds-key")

	// Anthropic should win (env precedence order)
	provider, token, method, err := ResolveCredentials()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if provider != "anthropic" {
		t.Errorf("provider = %q, want anthropic", provider)
	}
	if token != "ant-key" {
		t.Errorf("token = %q, want ant-key", token)
	}
	if method != "api_key" {
		t.Errorf("method = %q, want api_key", method)
	}
}

func TestResolveCredentialsDeepSeekEnv(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("DEEPSEEK_TOKEN", "ds-token")

	provider, token, method, err := ResolveCredentials()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if provider != "deepseek" {
		t.Errorf("provider = %q, want deepseek", provider)
	}
	if token != "ds-token" {
		t.Errorf("token = %q, want ds-token", token)
	}
	if method != "api_key" {
		t.Errorf("method = %q, want api_key", method)
	}
}

func TestResolveCredentialsNoCreds(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("DEEPSEEK_TOKEN", "")

	// Use a temp HOME to avoid picking up real stored creds.
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	_, _, _, err := ResolveCredentials()
	if err == nil {
		t.Error("expected error when no credentials are configured")
	}
}

func TestSetAndLoadProviderAPIKey(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	if err := SetProviderAPIKey("deepseek", "test-ds-key"); err != nil {
		t.Fatalf("SetProviderAPIKey: %v", err)
	}

	store, err := LoadCredentialStore()
	if err != nil {
		t.Fatalf("LoadCredentialStore: %v", err)
	}

	if got := store.ActiveProvider; got != "deepseek" {
		t.Errorf("ActiveProvider = %q, want deepseek", got)
	}
	cred, ok := store.Providers["deepseek"]
	if !ok {
		t.Fatal("deepseek provider not in store")
	}
	if cred.APIKey != "test-ds-key" {
		t.Errorf("APIKey = %q, want test-ds-key", cred.APIKey)
	}
}

func TestGetActiveProviderFallback(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	got := GetActiveProvider()
	if got != "anthropic" {
		t.Errorf("GetActiveProvider() = %q, want anthropic (fallback)", got)
	}
}

func TestCredentialsFileMode(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	if err := SetProviderAPIKey("deepseek", "secret"); err != nil {
		t.Fatalf("SetProviderAPIKey: %v", err)
	}

	// File should exist with restrictive permissions.
	path, _ := credentialsFilePath()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	mode := info.Mode().Perm()
	if mode != 0o600 {
		t.Errorf("credentials file mode = %o, want 0600", mode)
	}
}
