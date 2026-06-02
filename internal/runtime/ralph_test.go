package runtime

import "testing"

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
