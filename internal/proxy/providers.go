package proxy

import (
	"encoding/json"

	"github.com/uucz/llmview/internal/storage"
)

// ProviderConfig defines the upstream target for each provider.
type ProviderConfig struct {
	Provider  storage.Provider
	Upstream  string // e.g., "https://api.openai.com"
	PathStrip string // prefix to strip, e.g., "/proxy/openai"
}

// DefaultProviders returns the built-in provider routing table.
func DefaultProviders() []ProviderConfig {
	return []ProviderConfig{
		{
			Provider:  storage.ProviderOpenAI,
			Upstream:  "https://api.openai.com",
			PathStrip: "/proxy/openai",
		},
		{
			Provider:  storage.ProviderAnthropic,
			Upstream:  "https://api.anthropic.com",
			PathStrip: "/proxy/anthropic",
		},
		{
			Provider:  storage.ProviderOllama,
			Upstream:  "http://localhost:11434",
			PathStrip: "/proxy/ollama",
		},
	}
}

// ExtractModelOpenAI extracts the model name from an OpenAI-format request body.
func ExtractModelOpenAI(body []byte) string {
	var req struct {
		Model string `json:"model"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return ""
	}
	return req.Model
}

// ExtractModelAnthropic extracts the model name from an Anthropic-format request body.
func ExtractModelAnthropic(body []byte) string {
	var req struct {
		Model string `json:"model"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return ""
	}
	return req.Model
}

// ExtractStreamFlag checks if the request has "stream": true.
func ExtractStreamFlag(body []byte) bool {
	var req struct {
		Stream interface{} `json:"stream"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return false
	}
	switch v := req.Stream.(type) {
	case bool:
		return v
	case map[string]interface{}:
		// OpenAI stream_options format: {"stream": true} or presence of stream field
		return true
	}
	return false
}

// TokenUsage holds extracted token counts from a response.
type TokenUsage struct {
	InputTokens  int
	OutputTokens int
}

// ExtractUsageOpenAI extracts token usage from an OpenAI response body.
func ExtractUsageOpenAI(body []byte) TokenUsage {
	var resp struct {
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return TokenUsage{}
	}
	return TokenUsage{
		InputTokens:  resp.Usage.PromptTokens,
		OutputTokens: resp.Usage.CompletionTokens,
	}
}

// ExtractUsageAnthropic extracts token usage from an Anthropic response body.
func ExtractUsageAnthropic(body []byte) TokenUsage {
	var resp struct {
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return TokenUsage{}
	}
	return TokenUsage{
		InputTokens:  resp.Usage.InputTokens,
		OutputTokens: resp.Usage.OutputTokens,
	}
}

// ExtractStreamDelta extracts the text content delta from a streaming SSE chunk.
// Works for both OpenAI and Anthropic streaming formats.
func ExtractStreamDelta(provider storage.Provider, data []byte) string {
	switch provider {
	case storage.ProviderOpenAI:
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal(data, &chunk); err != nil || len(chunk.Choices) == 0 {
			return ""
		}
		return chunk.Choices[0].Delta.Content

	case storage.ProviderAnthropic:
		var chunk struct {
			Type  string `json:"type"`
			Delta struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"delta"`
		}
		if err := json.Unmarshal(data, &chunk); err != nil {
			return ""
		}
		if chunk.Type == "content_block_delta" {
			return chunk.Delta.Text
		}
		return ""

	default:
		return ""
	}
}

// ExtractStreamUsage extracts final token usage from the last SSE event.
func ExtractStreamUsage(provider storage.Provider, data []byte) TokenUsage {
	switch provider {
	case storage.ProviderOpenAI:
		return ExtractUsageOpenAI(data)
	case storage.ProviderAnthropic:
		var msg struct {
			Type  string `json:"type"`
			Usage struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			} `json:"usage"`
		}
		if err := json.Unmarshal(data, &msg); err != nil {
			return TokenUsage{}
		}
		if msg.Type == "message_delta" || msg.Type == "message_stop" {
			return TokenUsage{
				InputTokens:  msg.Usage.InputTokens,
				OutputTokens: msg.Usage.OutputTokens,
			}
		}
		return TokenUsage{}
	default:
		return TokenUsage{}
	}
}
