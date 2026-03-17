package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/uucz/llmview/internal/cost"
	"github.com/uucz/llmview/internal/proxy"
	"github.com/uucz/llmview/internal/storage"
	"github.com/uucz/llmview/internal/ws"
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

	// UI — embedded Svelte build
	mux.Handle("/", embeddedUI())

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

// Close cleans up server resources.
func (s *Server) Close() error {
	return s.store.Close()
}
