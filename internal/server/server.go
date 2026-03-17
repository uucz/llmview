package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/llmview/llmview/internal/cost"
	"github.com/llmview/llmview/internal/proxy"
	"github.com/llmview/llmview/internal/storage"
	"github.com/llmview/llmview/internal/ws"
)

// Server is the main HTTP server that mounts proxy, API, WebSocket, and UI.
type Server struct {
	port      int
	store     *storage.Store
	hub       *ws.Hub
	calc      *cost.Calculator
	proxy     *proxy.Proxy
	sessionID string
}

// Config holds server configuration.
type Config struct {
	Port      int
	DBPath    string
	SessionID string
}

// New creates and configures the server.
func New(cfg Config) (*Server, error) {
	store, err := storage.New(cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("init storage: %w", err)
	}

	hub := ws.NewHub()
	calc := cost.New()

	// Create session
	sess := &storage.Session{
		ID: cfg.SessionID,
	}
	if err := store.CreateSession(sess); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	p := proxy.New(store, hub, calc, cfg.SessionID)

	return &Server{
		port:      cfg.Port,
		store:     store,
		hub:       hub,
		calc:      calc,
		proxy:     p,
		sessionID: cfg.SessionID,
	}, nil
}

// Start begins listening and serving.
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Proxy routes — all LLM traffic goes through /proxy/*
	mux.Handle("/proxy/", s.proxy.Handler())

	// WebSocket for real-time UI updates
	mux.HandleFunc("/ws", s.hub.HandleWS)

	// REST API for historical data
	mux.HandleFunc("/api/session", s.handleSession)
	mux.HandleFunc("/api/calls", s.handleCalls)
	mux.HandleFunc("/api/calls/", s.handleCallDetail)

	// Health check
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// UI — embedded static files (TODO: go:embed in production, dev proxy in development)
	mux.HandleFunc("/", s.handleUI)

	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("llmview listening on http://localhost%s", addr)
	log.Printf("")
	log.Printf("  Configure your LLM clients:")
	log.Printf("    export OPENAI_BASE_URL=http://localhost%s/proxy/openai/v1", addr)
	log.Printf("    export ANTHROPIC_BASE_URL=http://localhost%s/proxy/anthropic", addr)
	log.Printf("")
	log.Printf("  Open http://localhost%s to view the dashboard", addr)
	log.Printf("")

	return http.ListenAndServe(addr, mux)
}

func (s *Server) handleSession(w http.ResponseWriter, r *http.Request) {
	sess, err := s.store.GetSession(s.sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sess)
}

func (s *Server) handleCalls(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	calls, err := s.store.ListCalls(s.sessionID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(calls)
}

func (s *Server) handleCallDetail(w http.ResponseWriter, r *http.Request) {
	// Extract call ID from path: /api/calls/{id}
	id := r.URL.Path[len("/api/calls/"):]
	if id == "" {
		http.Error(w, "missing call id", http.StatusBadRequest)
		return
	}

	detail, err := s.store.GetCallDetail(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detail)
}

func (s *Server) handleUI(w http.ResponseWriter, r *http.Request) {
	// TODO: In production, serve embedded Svelte build via go:embed
	// For now, serve a minimal placeholder that connects to WebSocket
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, minimalUI)
}

// Close cleans up server resources.
func (s *Server) Close() error {
	return s.store.Close()
}

const minimalUI = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>llmview</title>
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body { background: #0a0a0a; color: #e0e0e0; font-family: 'SF Mono', 'Fira Code', monospace; font-size: 14px; }
  .header { display: flex; justify-content: space-between; align-items: center; padding: 16px 24px; border-bottom: 1px solid #1a1a2e; }
  .logo { font-size: 18px; font-weight: 700; color: #00d4ff; }
  .stats { display: flex; gap: 24px; }
  .stat { text-align: center; }
  .stat-value { font-size: 24px; font-weight: 700; color: #fff; }
  .stat-label { font-size: 11px; color: #666; text-transform: uppercase; letter-spacing: 1px; }
  .cost { color: #00ff88; }
  .timeline { padding: 16px 24px; }
  .call { display: flex; align-items: center; gap: 16px; padding: 12px 16px; border-radius: 8px; margin-bottom: 8px; background: #111; border: 1px solid #1a1a2e; cursor: pointer; transition: border-color 0.2s; }
  .call:hover { border-color: #00d4ff33; }
  .call.streaming { border-left: 3px solid #00d4ff; }
  .provider { font-size: 11px; padding: 2px 8px; border-radius: 4px; font-weight: 600; text-transform: uppercase; }
  .provider.openai { background: #10a37f22; color: #10a37f; }
  .provider.anthropic { background: #d4a27422; color: #d4a274; }
  .provider.ollama { background: #666; color: #fff; }
  .model { color: #fff; font-weight: 500; flex: 1; }
  .tokens { color: #888; font-size: 12px; }
  .duration { color: #888; font-size: 12px; min-width: 60px; text-align: right; }
  .call-cost { color: #00ff88; font-size: 12px; min-width: 60px; text-align: right; }
  .empty { text-align: center; padding: 80px 24px; color: #444; }
  .empty h2 { font-size: 20px; margin-bottom: 16px; color: #666; }
  .empty code { background: #1a1a2e; padding: 8px 16px; border-radius: 6px; display: inline-block; margin: 4px 0; color: #00d4ff; }
  .stream-text { margin-top: 8px; padding: 8px 12px; background: #0a0a0a; border-radius: 4px; font-size: 13px; color: #aaa; max-height: 200px; overflow-y: auto; white-space: pre-wrap; word-break: break-all; display: none; }
  .call.expanded .stream-text { display: block; }
</style>
</head>
<body>
<div class="header">
  <div class="logo">llmview</div>
  <div class="stats">
    <div class="stat">
      <div class="stat-value" id="total-cost">$0.00</div>
      <div class="stat-label">Session Cost</div>
    </div>
    <div class="stat">
      <div class="stat-value" id="total-tokens">0</div>
      <div class="stat-label">Tokens</div>
    </div>
    <div class="stat">
      <div class="stat-value" id="total-requests">0</div>
      <div class="stat-label">Requests</div>
    </div>
  </div>
</div>
<div class="timeline" id="timeline">
  <div class="empty">
    <h2>Waiting for LLM traffic...</h2>
    <p>Set your base URL to proxy through llmview:</p>
    <br>
    <code>export OPENAI_BASE_URL=http://localhost:4700/proxy/openai/v1</code>
    <br>
    <code>export ANTHROPIC_BASE_URL=http://localhost:4700/proxy/anthropic</code>
  </div>
</div>

<script>
const timeline = document.getElementById('timeline');
const calls = {};
let hasContent = false;

function clearEmpty() {
  if (!hasContent) {
    timeline.innerHTML = '';
    hasContent = true;
  }
}

function formatDuration(ms) {
  if (ms < 1000) return ms + 'ms';
  return (ms / 1000).toFixed(1) + 's';
}

function formatCost(cost) {
  if (cost === 0) return 'free';
  if (cost < 0.01) return '$' + cost.toFixed(4);
  return '$' + cost.toFixed(2);
}

function formatTokens(n) {
  if (n >= 1000000) return (n / 1000000).toFixed(1) + 'M';
  if (n >= 1000) return (n / 1000).toFixed(1) + 'k';
  return n.toString();
}

const ws = new WebSocket('ws://' + location.host + '/ws');

ws.onmessage = function(e) {
  const event = JSON.parse(e.data);

  switch (event.type) {
    case 'call_start': {
      clearEmpty();
      const d = event.data;
      const el = document.createElement('div');
      el.className = 'call' + (d.streaming ? ' streaming' : '');
      el.id = 'call-' + d.id;
      el.innerHTML =
        '<span class="provider ' + d.provider + '">' + d.provider + '</span>' +
        '<span class="model">' + (d.model || d.endpoint) + '</span>' +
        '<span class="tokens" id="tokens-' + d.id + '">...</span>' +
        '<span class="duration" id="dur-' + d.id + '">...</span>' +
        '<span class="call-cost" id="cost-' + d.id + '">...</span>' +
        '<div class="stream-text" id="stream-' + d.id + '"></div>';
      el.onclick = function() { el.classList.toggle('expanded'); };
      timeline.prepend(el);
      calls[d.id] = { streamText: '' };
      break;
    }

    case 'call_chunk': {
      const d = event.data;
      if (calls[d.id]) {
        calls[d.id].streamText += d.delta;
        const streamEl = document.getElementById('stream-' + d.id);
        if (streamEl) {
          streamEl.textContent = calls[d.id].streamText;
          streamEl.scrollTop = streamEl.scrollHeight;
        }
        // Auto-expand while streaming
        const callEl = document.getElementById('call-' + d.id);
        if (callEl && !callEl.classList.contains('expanded')) {
          callEl.classList.add('expanded');
        }
      }
      break;
    }

    case 'call_end': {
      const d = event.data;
      const tokEl = document.getElementById('tokens-' + d.id);
      const durEl = document.getElementById('dur-' + d.id);
      const costEl = document.getElementById('cost-' + d.id);
      if (tokEl) tokEl.textContent = formatTokens(d.input_tokens) + ' → ' + formatTokens(d.output_tokens);
      if (durEl) durEl.textContent = formatDuration(d.duration_ms);
      if (costEl) costEl.textContent = formatCost(d.cost);
      if (d.error) {
        const callEl = document.getElementById('call-' + d.id);
        if (callEl) callEl.style.borderColor = '#ff4444';
      }
      break;
    }

    case 'session_update': {
      const d = event.data;
      document.getElementById('total-cost').textContent = formatCost(d.total_cost);
      document.getElementById('total-tokens').textContent = formatTokens(d.total_tokens);
      document.getElementById('total-requests').textContent = d.request_count;
      break;
    }
  }
};

ws.onclose = function() {
  console.log('WebSocket closed, reconnecting...');
  setTimeout(function() { location.reload(); }, 2000);
};
</script>
</body>
</html>`
