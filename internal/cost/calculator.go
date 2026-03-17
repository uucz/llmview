package cost

// ModelPricing stores per-million-token pricing for a model.
type ModelPricing struct {
	InputPerMillion  float64
	OutputPerMillion float64
}

// Calculator computes API call costs based on model and token usage.
type Calculator struct {
	prices map[string]ModelPricing
}

// New creates a Calculator with default model pricing (as of early 2026).
func New() *Calculator {
	return &Calculator{
		prices: map[string]ModelPricing{
			// OpenAI
			"gpt-4o":                {InputPerMillion: 2.50, OutputPerMillion: 10.00},
			"gpt-4o-2024-11-20":    {InputPerMillion: 2.50, OutputPerMillion: 10.00},
			"gpt-4o-mini":          {InputPerMillion: 0.15, OutputPerMillion: 0.60},
			"gpt-4-turbo":          {InputPerMillion: 10.00, OutputPerMillion: 30.00},
			"gpt-4":                {InputPerMillion: 30.00, OutputPerMillion: 60.00},
			"gpt-3.5-turbo":        {InputPerMillion: 0.50, OutputPerMillion: 1.50},
			"o1":                   {InputPerMillion: 15.00, OutputPerMillion: 60.00},
			"o1-mini":              {InputPerMillion: 3.00, OutputPerMillion: 12.00},
			"o3":                   {InputPerMillion: 10.00, OutputPerMillion: 40.00},
			"o3-mini":              {InputPerMillion: 1.10, OutputPerMillion: 4.40},
			"o4-mini":              {InputPerMillion: 1.10, OutputPerMillion: 4.40},

			// Anthropic
			"claude-opus-4-6":        {InputPerMillion: 15.00, OutputPerMillion: 75.00},
			"claude-sonnet-4-6":      {InputPerMillion: 3.00, OutputPerMillion: 15.00},
			"claude-haiku-4-5":       {InputPerMillion: 0.80, OutputPerMillion: 4.00},
			"claude-3-5-sonnet-latest": {InputPerMillion: 3.00, OutputPerMillion: 15.00},
			"claude-3-5-haiku-latest":  {InputPerMillion: 0.80, OutputPerMillion: 4.00},

			// Ollama / local models — free
			"llama3":   {InputPerMillion: 0, OutputPerMillion: 0},
			"llama3.1": {InputPerMillion: 0, OutputPerMillion: 0},
			"mistral":  {InputPerMillion: 0, OutputPerMillion: 0},
			"qwen2.5":  {InputPerMillion: 0, OutputPerMillion: 0},
			"phi-3":    {InputPerMillion: 0, OutputPerMillion: 0},
		},
	}
}

// Calculate returns the cost in USD for the given model and token counts.
func (c *Calculator) Calculate(model string, inputTokens, outputTokens int) float64 {
	pricing, ok := c.prices[model]
	if !ok {
		// Try prefix matching for versioned model names
		pricing = c.findByPrefix(model)
	}

	inputCost := float64(inputTokens) / 1_000_000 * pricing.InputPerMillion
	outputCost := float64(outputTokens) / 1_000_000 * pricing.OutputPerMillion
	return inputCost + outputCost
}

func (c *Calculator) findByPrefix(model string) ModelPricing {
	// Match longest prefix: "gpt-4o-2025-01-01" → "gpt-4o"
	best := ModelPricing{}
	bestLen := 0
	for name, pricing := range c.prices {
		if len(name) > bestLen && len(model) >= len(name) && model[:len(name)] == name {
			best = pricing
			bestLen = len(name)
		}
	}
	return best
}

// SetPricing allows overriding or adding model pricing at runtime.
func (c *Calculator) SetPricing(model string, input, output float64) {
	c.prices[model] = ModelPricing{
		InputPerMillion:  input,
		OutputPerMillion: output,
	}
}

// KnownModel returns true if pricing data exists for this model.
func (c *Calculator) KnownModel(model string) bool {
	if _, ok := c.prices[model]; ok {
		return true
	}
	return c.findByPrefix(model) != (ModelPricing{})
}
