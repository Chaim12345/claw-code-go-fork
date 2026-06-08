package usage

import (
	"strings"
	"testing"
)

func TestNewTracker(t *testing.T) {
	tr := NewTracker("gpt-4o")
	if tr.model != "gpt-4o" {
		t.Errorf("model = %q, want %q", tr.model, "gpt-4o")
	}
	if tr.Turns != 0 {
		t.Errorf("Turns = %d, want 0", tr.Turns)
	}
}

func TestAdd(t *testing.T) {
	tr := NewTracker("gpt-4o")
	tr.Add(100, 50, 10, 5)
	tr.Add(200, 80, 0, 0)

	if tr.TotalInput != 300 {
		t.Errorf("TotalInput = %d, want 300", tr.TotalInput)
	}
	if tr.TotalOutput != 130 {
		t.Errorf("TotalOutput = %d, want 130", tr.TotalOutput)
	}
	if tr.CacheWrite != 10 {
		t.Errorf("CacheWrite = %d, want 10", tr.CacheWrite)
	}
	if tr.CacheRead != 5 {
		t.Errorf("CacheRead = %d, want 5", tr.CacheRead)
	}
	if tr.Turns != 2 {
		t.Errorf("Turns = %d, want 2", tr.Turns)
	}
}

func TestTotalTokens(t *testing.T) {
	tr := NewTracker("gpt-4o")
	tr.Add(100, 200, 0, 0)
	if got := tr.TotalTokens(); got != 300 {
		t.Errorf("TotalTokens() = %d, want 300", got)
	}
}

func TestSetModel(t *testing.T) {
	tr := NewTracker("gpt-4o")
	tr.SetModel("claude-sonnet-4-6")
	if tr.model != "claude-sonnet-4-6" {
		t.Errorf("model after SetModel = %q, want %q", tr.model, "claude-sonnet-4-6")
	}
}

func TestCostEstimate_KnownModel(t *testing.T) {
	tr := NewTracker("gpt-4o")
	tr.Add(1_000_000, 1_000_000, 0, 0) // 1M input + 1M output
	cost := tr.CostEstimate()
	expected := 2.50 + 10.00
	if cost != expected {
		t.Errorf("CostEstimate() = %f, want %f", cost, expected)
	}
}

func TestCostEstimate_UnknownModel(t *testing.T) {
	tr := NewTracker("unknown-model")
	tr.Add(1000, 500, 0, 0)
	if cost := tr.CostEstimate(); cost != -1 {
		t.Errorf("CostEstimate() = %f, want -1", cost)
	}
}

func TestPricingKnown(t *testing.T) {
	tr := NewTracker("gpt-4o")
	if !tr.PricingKnown() {
		t.Error("PricingKnown() = false for gpt-4o, want true")
	}
	tr.SetModel("nonexistent")
	if tr.PricingKnown() {
		t.Error("PricingKnown() = true for nonexistent model, want false")
	}
}

func TestFormatSummary_WithCost(t *testing.T) {
	tr := NewTracker("gpt-4o")
	tr.Add(1000, 500, 100, 50)
	summary := tr.FormatSummary()

	if !strings.Contains(summary, "gpt-4o") {
		t.Error("FormatSummary missing model name")
	}
	if !strings.Contains(summary, "Turns") {
		t.Error("FormatSummary missing turns")
	}
	if !strings.Contains(summary, "Input tokens") {
		t.Error("FormatSummary missing input tokens")
	}
	if !strings.Contains(summary, "Output tokens") {
		t.Error("FormatSummary missing output tokens")
	}
	if !strings.Contains(summary, "Cache write") {
		t.Error("FormatSummary missing cache write")
	}
	if !strings.Contains(summary, "Cache read") {
		t.Error("FormatSummary missing cache read")
	}
	if !strings.Contains(summary, "Est. cost (USD)") {
		t.Error("FormatSummary missing cost estimate")
	}
}

func TestFormatSummary_UnknownModel(t *testing.T) {
	tr := NewTracker("unknown-model")
	tr.Add(100, 50, 0, 0)
	summary := tr.FormatSummary()
	if !strings.Contains(summary, "pricing not known") {
		t.Error("FormatSummary missing unknown-model message")
	}
}

func TestFormatSummary_NoCache(t *testing.T) {
	tr := NewTracker("gpt-4o")
	tr.Add(100, 50, 0, 0)
	summary := tr.FormatSummary()
	if strings.Contains(summary, "Cache write") {
		t.Error("FormatSummary should not contain cache info when cache is 0")
	}
}

func TestFormatNum(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{999, "999"},
		{1000, "1,000"},
		{1234567, "1,234,567"},
		{100, "100"},
		{10000, "10,000"},
	}
	for _, tt := range tests {
		got := formatNum(tt.input)
		if got != tt.want {
			t.Errorf("formatNum(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
