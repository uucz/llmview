package proxy

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/uucz/llmview/internal/cost"
	"github.com/uucz/llmview/internal/storage"
	"github.com/uucz/llmview/internal/ws"
)

// Proxy intercepts LLM API calls, logs them, and forwards to upstream.
type Proxy struct {
	providers  []ProviderConfig
	store      *storage.Store
	hub        *ws.Hub
	calc       *cost.Calculator
	sessionID  string
	httpClient *http.Client
}

// New creates a new Proxy instance.
func New(store *storage.Store, hub *ws.Hub, calc *cost.Calculator, sessionID string) *Proxy {
	return &Proxy{
		providers: DefaultProviders(),
		store:     store,
		hub:       hub,
		calc:      calc,
		sessionID: sessionID,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // LLM calls can be slow
		},
	}
}

// Handler returns an http.Handler that routes to the correct provider proxy.
func (p *Proxy) Handler() http.Handler {
	mux := http.NewServeMux()
	for _, pc := range p.providers {
		pc := pc // capture
		mux.HandleFunc(pc.PathStrip+"/", func(w http.ResponseWriter, r *http.Request) {
			p.handleProxy(w, r, pc)
		})
	}
	return mux
}

func (p *Proxy) handleProxy(w http.ResponseWriter, r *http.Request, pc ProviderConfig) {
	callID := generateID()
	startedAt := time.Now()

	// Read request body
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	r.Body.Close()

	// Extract metadata from request
	model := extractModel(pc.Provider, reqBody)
	streaming := ExtractStreamFlag(reqBody)

	// Broadcast call start
	p.hub.Broadcast(storage.WSEvent{
		Type: "call_start",
		Data: storage.CallStartEvent{
			ID:        callID,
			Provider:  pc.Provider,
			Model:     model,
			Endpoint:  r.URL.Path,
			Streaming: streaming,
			StartedAt: startedAt.UnixMilli(),
		},
	})

	// Build upstream request
	upstreamPath := strings.TrimPrefix(r.URL.Path, pc.PathStrip)
	upstreamURL := pc.Upstream + upstreamPath
	if r.URL.RawQuery != "" {
		upstreamURL += "?" + r.URL.RawQuery
	}

	upReq, err := http.NewRequestWithContext(r.Context(), r.Method, upstreamURL, bytes.NewReader(reqBody))
	if err != nil {
		http.Error(w, "failed to create upstream request", http.StatusInternalServerError)
		return
	}

	// Forward all headers
	for key, vals := range r.Header {
		for _, val := range vals {
			upReq.Header.Add(key, val)
		}
	}

	// Execute upstream request
	resp, err := p.httpClient.Do(upReq)
	if err != nil {
		log.Printf("[proxy] upstream error: %v", err)
		http.Error(w, "upstream request failed", http.StatusBadGateway)
		p.recordError(callID, r.URL.Path, pc, model, reqBody, startedAt, err)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, vals := range resp.Header {
		for _, val := range vals {
			w.Header().Add(key, val)
		}
	}
	w.WriteHeader(resp.StatusCode)

	// Handle streaming vs non-streaming
	var respBody []byte
	var usage TokenUsage

	if streaming && resp.StatusCode == http.StatusOK {
		respBody, usage = p.handleStreaming(w, resp, pc.Provider, callID)
	} else {
		respBody, err = io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("[proxy] read response error: %v", err)
		}
		w.Write(respBody)
		usage = extractUsage(pc.Provider, respBody)
	}

	duration := time.Since(startedAt)
	callCost := p.calc.Calculate(model, usage.InputTokens, usage.OutputTokens)

	// Store the call
	call := &storage.APICall{
		ID:           callID,
		SessionID:    p.sessionID,
		Provider:     pc.Provider,
		Model:        model,
		Endpoint:     r.URL.Path,
		Method:       r.Method,
		RequestBody:  reqBody,
		ResponseBody: respBody,
		StatusCode:   resp.StatusCode,
		StartedAt:    startedAt,
		DurationMs:   duration.Milliseconds(),
		InputTokens:  usage.InputTokens,
		OutputTokens: usage.OutputTokens,
		Cost:         callCost,
		Streaming:    streaming,
	}

	if err := p.store.InsertCall(call); err != nil {
		log.Printf("[proxy] store error: %v", err)
	}

	// Broadcast call end
	p.hub.Broadcast(storage.WSEvent{
		Type: "call_end",
		Data: storage.CallEndEvent{
			ID:           callID,
			StatusCode:   resp.StatusCode,
			DurationMs:   duration.Milliseconds(),
			InputTokens:  usage.InputTokens,
			OutputTokens: usage.OutputTokens,
			Cost:         callCost,
		},
	})

	// Broadcast session update
	if sess, err := p.store.GetSession(p.sessionID); err == nil {
		p.hub.Broadcast(storage.WSEvent{
			Type: "session_update",
			Data: storage.SessionUpdateEvent{
				TotalCost:    sess.TotalCost,
				TotalTokens:  sess.TotalTokens,
				RequestCount: sess.RequestCount,
			},
		})
	}

	log.Printf("[proxy] %s %s → %d | %s | %d+%d tokens | $%.4f | %s",
		r.Method, r.URL.Path, resp.StatusCode, model,
		usage.InputTokens, usage.OutputTokens, callCost, duration.Round(time.Millisecond))
}

// handleStreaming tees SSE chunks to the client and captures for storage.
func (p *Proxy) handleStreaming(w http.ResponseWriter, resp *http.Response, provider storage.Provider, callID string) ([]byte, TokenUsage) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		// Fallback: read all and write
		body, _ := io.ReadAll(resp.Body)
		w.Write(body)
		return body, TokenUsage{}
	}

	var fullBody bytes.Buffer
	var finalUsage TokenUsage
	chunkIndex := 0
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // 1MB max line

	for scanner.Scan() {
		line := scanner.Text()

		// Write to client immediately
		fmt.Fprintf(w, "%s\n", line)
		flusher.Flush()

		// Accumulate full response
		fullBody.WriteString(line)
		fullBody.WriteByte('\n')

		// Parse SSE data lines
		if strings.HasPrefix(line, "data: ") {
			data := []byte(strings.TrimPrefix(line, "data: "))

			if string(data) == "[DONE]" {
				continue
			}

			// Extract text delta for real-time streaming to UI
			delta := ExtractStreamDelta(provider, data)
			if delta != "" {
				p.hub.Broadcast(storage.WSEvent{
					Type: "call_chunk",
					Data: storage.CallChunkEvent{
						ID:    callID,
						Delta: delta,
						Index: chunkIndex,
					},
				})
				chunkIndex++
			}

			// Check for usage in final chunks
			if usage := ExtractStreamUsage(provider, data); usage.InputTokens > 0 || usage.OutputTokens > 0 {
				finalUsage = usage
			}
		}
	}

	return fullBody.Bytes(), finalUsage
}

func (p *Proxy) recordError(callID, endpoint string, pc ProviderConfig, model string, reqBody []byte, startedAt time.Time, err error) {
	call := &storage.APICall{
		ID:          callID,
		SessionID:   p.sessionID,
		Provider:    pc.Provider,
		Model:       model,
		Endpoint:    endpoint,
		Method:      "POST",
		RequestBody: reqBody,
		StatusCode:  502,
		StartedAt:   startedAt,
		DurationMs:  time.Since(startedAt).Milliseconds(),
		Error:       err.Error(),
	}
	if storeErr := p.store.InsertCall(call); storeErr != nil {
		log.Printf("[proxy] store error on error record: %v", storeErr)
	}

	p.hub.Broadcast(storage.WSEvent{
		Type: "call_end",
		Data: storage.CallEndEvent{
			ID:         callID,
			StatusCode: 502,
			DurationMs: call.DurationMs,
			Error:      err.Error(),
		},
	})
}

func extractModel(provider storage.Provider, body []byte) string {
	switch provider {
	case storage.ProviderOpenAI, storage.ProviderOllama:
		return ExtractModelOpenAI(body)
	case storage.ProviderAnthropic:
		return ExtractModelAnthropic(body)
	default:
		return ""
	}
}

func extractUsage(provider storage.Provider, body []byte) TokenUsage {
	switch provider {
	case storage.ProviderOpenAI, storage.ProviderOllama:
		return ExtractUsageOpenAI(body)
	case storage.ProviderAnthropic:
		return ExtractUsageAnthropic(body)
	default:
		return TokenUsage{}
	}
}

func generateID() string {
	b := make([]byte, 12)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
