package cost

import (
	"math"
	"testing"
)

func TestCalculateKnownModels(t *testing.T) {
	c := New()

	tests := []struct {
		model  string
		input  int
		output int
		want   float64
	}{
		// GPT-4o: $2.50/1M input, $10/1M output
		{"gpt-4o", 1000, 500, 0.0025 + 0.005},
		// GPT-4o-mini: $0.15/1M input, $0.60/1M output
		{"gpt-4o-mini", 10000, 5000, 0.0015 + 0.003},
		// Claude Opus: $15/1M input, $75/1M output
		{"claude-opus-4-6", 1000, 1000, 0.015 + 0.075},
		// Claude Sonnet: $3/1M input, $15/1M output
		{"claude-sonnet-4-6", 2000, 1000, 0.006 + 0.015},
		// Ollama (free)
		{"llama3", 100000, 50000, 0},
		// Zero tokens
		{"gpt-4o", 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			got := c.Calculate(tt.model, tt.input, tt.output)
			if math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("Calculate(%q, %d, %d) = %f, want %f", tt.model, tt.input, tt.output, got, tt.want)
			}
		})
	}
}

func TestCalculatePrefixMatching(t *testing.T) {
	c := New()

	// Versioned model name should match base model pricing
	cost1 := c.Calculate("gpt-4o-2025-03-15", 1000, 500)
	cost2 := c.Calculate("gpt-4o", 1000, 500)
	if math.Abs(cost1-cost2) > 1e-9 {
		t.Errorf("prefix match failed: gpt-4o-2025-03-15 = %f, gpt-4o = %f", cost1, cost2)
	}
}

func TestCalculateUnknownModel(t *testing.T) {
	c := New()

	// Unknown model returns 0 cost
	got := c.Calculate("some-unknown-model-xyz", 1000, 1000)
	if got != 0 {
		t.Errorf("unknown model should return 0, got %f", got)
	}
}

func TestKnownModel(t *testing.T) {
	c := New()

	if !c.KnownModel("gpt-4o") {
		t.Error("gpt-4o should be known")
	}
	if !c.KnownModel("gpt-4o-2025-01-01") {
		t.Error("gpt-4o-2025-01-01 should match via prefix")
	}
	if c.KnownModel("totally-fake-model") {
		t.Error("totally-fake-model should not be known")
	}
}

func TestSetPricing(t *testing.T) {
	c := New()

	c.SetPricing("my-custom-model", 5.0, 20.0)
	got := c.Calculate("my-custom-model", 1_000_000, 1_000_000)
	want := 5.0 + 20.0
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("custom pricing: got %f, want %f", got, want)
	}
}

func TestCalculateLargeTokenCounts(t *testing.T) {
	c := New()

	// 1M input + 1M output on gpt-4o
	got := c.Calculate("gpt-4o", 1_000_000, 1_000_000)
	want := 2.50 + 10.00
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("large tokens: got %f, want %f", got, want)
	}
}
