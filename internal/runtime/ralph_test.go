package runtime

import (
	"strings"
	"testing"
)

func TestRalphConfig_Defaults(t *testing.T) {
	cfg := DefaultRalphConfig()
	if cfg.SpecPath == "" {
		t.Error("SpecPath should have a default")
	}
	if cfg.MaxIterations <= 0 {
		t.Error("MaxIterations should have a positive default")
	}
	if cfg.DoneSentinel == "" {
		t.Error("DoneSentinel should have a default")
	}
	if cfg.PromptTemplate == "" {
		t.Error("PromptTemplate should have a default")
	}
}

func TestRenderRalphPrompt(t *testing.T) {
	cfg := DefaultRalphConfig()
	cfg.SpecPath = "ROADMAP.md"
	rendered, err := RenderRalphPrompt(cfg, "do thing one", 1, 10)
	if err != nil {
		t.Fatalf("RenderRalphPrompt: %v", err)
	}
	if !strings.Contains(rendered, "Iteration 1 of 10") {
		t.Error("expected iteration/max in prompt")
	}
	if !strings.Contains(rendered, "ROADMAP.md") {
		t.Error("expected spec path in prompt")
	}
	if !strings.Contains(rendered, "do thing one") {
		t.Error("expected spec body in prompt")
	}
	if !strings.Contains(rendered, cfg.DoneSentinel) {
		t.Error("expected sentinel in prompt")
	}
}

func TestRenderRalphPrompt_RejectsBadTemplate(t *testing.T) {
	cfg := DefaultRalphConfig()
	cfg.PromptTemplate = "{{ .Undeclared }"
	_, err := RenderRalphPrompt(cfg, "x", 1, 1)
	if err == nil {
		t.Error("expected error on bad template")
	}
}
