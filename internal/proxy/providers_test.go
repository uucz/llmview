package proxy

import (
	"testing"

	"github.com/uucz/llmview/internal/storage"
)

func TestExtractModelOpenAI(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{"standard", `{"model":"gpt-4o","messages":[]}`, "gpt-4o"},
		{"mini", `{"model":"gpt-4o-mini","messages":[{"role":"user","content":"hi"}]}`, "gpt-4o-mini"},
		{"empty", `{}`, ""},
		{"invalid json", `not json`, ""},
		{"no model field", `{"messages":[]}`, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractModelOpenAI([]byte(tt.body))
			if got != tt.want {
				t.Errorf("ExtractModelOpenAI(%q) = %q, want %q", tt.body, got, tt.want)
			}
		})
	}
}

func TestExtractModelAnthropic(t *testing.T) {
	body := `{"model":"claude-sonnet-4-6","max_tokens":1024,"messages":[{"role":"user","content":"hello"}]}`
	got := ExtractModelAnthropic([]byte(body))
	if got != "claude-sonnet-4-6" {
		t.Errorf("got %q, want claude-sonnet-4-6", got)
	}
}

func TestExtractStreamFlag(t *testing.T) {
	tests := []struct {
		name string
		body string
		want bool
	}{
		{"true", `{"model":"gpt-4o","stream":true}`, true},
		{"false", `{"model":"gpt-4o","stream":false}`, false},
		{"absent", `{"model":"gpt-4o"}`, false},
		{"invalid", `not json`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractStreamFlag([]byte(tt.body))
			if got != tt.want {
				t.Errorf("ExtractStreamFlag(%q) = %v, want %v", tt.body, got, tt.want)
			}
		})
	}
}

func TestExtractUsageOpenAI(t *testing.T) {
	body := `{"id":"chatcmpl-123","choices":[{"message":{"content":"hi"}}],"usage":{"prompt_tokens":10,"completion_tokens":20,"total_tokens":30}}`
	usage := ExtractUsageOpenAI([]byte(body))
	if usage.InputTokens != 10 {
		t.Errorf("input tokens: got %d, want 10", usage.InputTokens)
	}
	if usage.OutputTokens != 20 {
		t.Errorf("output tokens: got %d, want 20", usage.OutputTokens)
	}
}

func TestExtractUsageAnthropic(t *testing.T) {
	body := `{"id":"msg_123","type":"message","role":"assistant","content":[{"type":"text","text":"hi"}],"usage":{"input_tokens":15,"output_tokens":25}}`
	usage := ExtractUsageAnthropic([]byte(body))
	if usage.InputTokens != 15 {
		t.Errorf("input tokens: got %d, want 15", usage.InputTokens)
	}
	if usage.OutputTokens != 25 {
		t.Errorf("output tokens: got %d, want 25", usage.OutputTokens)
	}
}

func TestExtractUsageInvalidJSON(t *testing.T) {
	usage := ExtractUsageOpenAI([]byte("not json"))
	if usage.InputTokens != 0 || usage.OutputTokens != 0 {
		t.Errorf("invalid JSON should return zero usage, got %+v", usage)
	}
}

func TestExtractStreamDeltaOpenAI(t *testing.T) {
	chunk := `{"choices":[{"delta":{"content":"Hello"}}]}`
	got := ExtractStreamDelta(storage.ProviderOpenAI, []byte(chunk))
	if got != "Hello" {
		t.Errorf("got %q, want %q", got, "Hello")
	}
}

func TestExtractStreamDeltaAnthropic(t *testing.T) {
	chunk := `{"type":"content_block_delta","delta":{"type":"text_delta","text":"World"}}`
	got := ExtractStreamDelta(storage.ProviderAnthropic, []byte(chunk))
	if got != "World" {
		t.Errorf("got %q, want %q", got, "World")
	}
}

func TestExtractStreamDeltaAnthropicNonDelta(t *testing.T) {
	// message_start events should not return text
	chunk := `{"type":"message_start","message":{"id":"msg_123"}}`
	got := ExtractStreamDelta(storage.ProviderAnthropic, []byte(chunk))
	if got != "" {
		t.Errorf("non-delta event should return empty, got %q", got)
	}
}

func TestExtractStreamDeltaEmptyChoices(t *testing.T) {
	chunk := `{"choices":[]}`
	got := ExtractStreamDelta(storage.ProviderOpenAI, []byte(chunk))
	if got != "" {
		t.Errorf("empty choices should return empty, got %q", got)
	}
}

func TestExtractStreamUsageAnthropic(t *testing.T) {
	chunk := `{"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"input_tokens":100,"output_tokens":50}}`
	usage := ExtractStreamUsage(storage.ProviderAnthropic, []byte(chunk))
	if usage.InputTokens != 100 {
		t.Errorf("input: got %d, want 100", usage.InputTokens)
	}
	if usage.OutputTokens != 50 {
		t.Errorf("output: got %d, want 50", usage.OutputTokens)
	}
}

func TestDefaultProviders(t *testing.T) {
	providers := DefaultProviders()
	if len(providers) != 3 {
		t.Fatalf("expected 3 providers, got %d", len(providers))
	}

	names := map[storage.Provider]bool{}
	for _, p := range providers {
		names[p.Provider] = true
		if p.Upstream == "" {
			t.Errorf("provider %s has empty upstream", p.Provider)
		}
		if p.PathStrip == "" {
			t.Errorf("provider %s has empty path strip", p.Provider)
		}
	}

	for _, expected := range []storage.Provider{storage.ProviderOpenAI, storage.ProviderAnthropic, storage.ProviderOllama} {
		if !names[expected] {
			t.Errorf("missing provider: %s", expected)
		}
	}
}
