package storage

import "time"

// Session groups API calls made during a single llmview run.
type Session struct {
	ID           string    `json:"id"`
	StartedAt    time.Time `json:"started_at"`
	TotalCost    float64   `json:"total_cost"`
	TotalTokens  int       `json:"total_tokens"`
	RequestCount int       `json:"request_count"`
}

// APICall represents a single intercepted LLM API request/response pair.
type APICall struct {
	ID           string        `json:"id"`
	SessionID    string        `json:"session_id"`
	Provider     Provider      `json:"provider"`
	Model        string        `json:"model"`
	Endpoint     string        `json:"endpoint"`
	Method       string        `json:"method"`
	RequestBody  []byte        `json:"-"`
	ResponseBody []byte        `json:"-"`
	StatusCode   int           `json:"status_code"`
	StartedAt    time.Time     `json:"started_at"`
	Duration     time.Duration `json:"duration_ms"`
	InputTokens  int           `json:"input_tokens"`
	OutputTokens int           `json:"output_tokens"`
	Cost         float64       `json:"cost"`
	Streaming    bool          `json:"streaming"`
	Error        string        `json:"error,omitempty"`
}

// APICallDetail includes full request/response bodies (for detail view).
type APICallDetail struct {
	APICall
	RequestBody  string `json:"request_body"`
	ResponseBody string `json:"response_body"`
}

// Provider identifies the LLM API provider.
type Provider string

const (
	ProviderOpenAI    Provider = "openai"
	ProviderAnthropic Provider = "anthropic"
	ProviderOllama    Provider = "ollama"
	ProviderUnknown   Provider = "unknown"
)

// WSEvent is pushed to the frontend via WebSocket.
type WSEvent struct {
	Type string      `json:"type"` // "call_start", "call_chunk", "call_end", "session_update"
	Data interface{} `json:"data"`
}

// CallStartEvent is sent when a new API call begins.
type CallStartEvent struct {
	ID        string   `json:"id"`
	Provider  Provider `json:"provider"`
	Model     string   `json:"model"`
	Endpoint  string   `json:"endpoint"`
	Streaming bool     `json:"streaming"`
	StartedAt int64    `json:"started_at"` // unix ms
}

// CallChunkEvent is sent for each streaming token/chunk.
type CallChunkEvent struct {
	ID    string `json:"id"`
	Delta string `json:"delta"` // text content delta
	Index int    `json:"index"` // chunk sequence number
}

// CallEndEvent is sent when an API call completes.
type CallEndEvent struct {
	ID           string  `json:"id"`
	StatusCode   int     `json:"status_code"`
	DurationMs   int64   `json:"duration_ms"`
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	Cost         float64 `json:"cost"`
	Error        string  `json:"error,omitempty"`
}

// SessionUpdateEvent is sent when session totals change.
type SessionUpdateEvent struct {
	TotalCost    float64 `json:"total_cost"`
	TotalTokens  int     `json:"total_tokens"`
	RequestCount int     `json:"request_count"`
}
