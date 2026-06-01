package runtime

import "testing"

func TestSelectProviderDeepSeek(t *testing.T) {
	p := SelectProvider("deepseek")
	if p == nil {
		t.Fatal("expected non-nil provider for 'deepseek'")
	}
	if p.Name() != "deepseek" {
		t.Errorf("provider name = %q, want 'deepseek'", p.Name())
	}
}

func TestSelectProviderDefault(t *testing.T) {
	p := SelectProvider("nonexistent")
	if p == nil {
		t.Fatal("expected non-nil provider for unknown name (should fall back to anthropic)")
	}
	if p.Name() != "anthropic" {
		t.Errorf("default provider name = %q, want 'anthropic'", p.Name())
	}
}

func TestSelectProviderAllKnown(t *testing.T) {
	for _, name := range []string{"anthropic", "openai", "bedrock", "vertex", "foundry", "deepseek"} {
		p := SelectProvider(name)
		if p == nil {
			t.Errorf("provider %q returned nil", name)
			continue
		}
		if p.Name() != name {
			t.Errorf("provider %q returned name %q", name, p.Name())
		}
	}
}

func TestDetectProviderDeepSeek(t *testing.T) {
	t.Setenv("CLAUDE_CODE_USE_BEDROCK", "")
	t.Setenv("CLAUDE_CODE_USE_VERTEX", "")
	t.Setenv("CLAUDE_CODE_USE_FOUNDRY", "")
	t.Setenv("DEEPSEEK_TOKEN", "test-token")
	if got := detectProvider(); got != "deepseek" {
		t.Errorf("detectProvider() = %q, want deepseek", got)
	}
}

func TestDetectProviderDeepSeekFlag(t *testing.T) {
	t.Setenv("CLAUDE_CODE_USE_BEDROCK", "")
	t.Setenv("CLAUDE_CODE_USE_VERTEX", "")
	t.Setenv("CLAUDE_CODE_USE_FOUNDRY", "")
	t.Setenv("DEEPSEEK_TOKEN", "")
	t.Setenv("CLAUDE_CODE_USE_DEEPSEEK", "1")
	if got := detectProvider(); got != "deepseek" {
		t.Errorf("detectProvider() = %q, want deepseek", got)
	}
}
