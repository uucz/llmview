package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Store handles all SQLite persistence.
type Store struct {
	db *sql.DB
	mu sync.RWMutex
}

// New creates a Store, initializing the database at the given path.
// If path is empty, uses ~/.llmview/llmview.db
func New(path string) (*Store, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("get home dir: %w", err)
		}
		dir := filepath.Join(home, ".llmview")
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create data dir: %w", err)
		}
		path = filepath.Join(dir, "llmview.db")
	}

	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return s, nil
}

func (s *Store) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS sessions (
		id          TEXT PRIMARY KEY,
		started_at  DATETIME NOT NULL,
		total_cost  REAL DEFAULT 0,
		total_tokens INTEGER DEFAULT 0,
		request_count INTEGER DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS api_calls (
		id            TEXT PRIMARY KEY,
		session_id    TEXT NOT NULL REFERENCES sessions(id),
		provider      TEXT NOT NULL,
		model         TEXT NOT NULL DEFAULT '',
		endpoint      TEXT NOT NULL,
		method        TEXT NOT NULL DEFAULT 'POST',
		request_body  BLOB,
		response_body BLOB,
		status_code   INTEGER DEFAULT 0,
		started_at    DATETIME NOT NULL,
		duration_ms   INTEGER DEFAULT 0,
		input_tokens  INTEGER DEFAULT 0,
		output_tokens INTEGER DEFAULT 0,
		cost          REAL DEFAULT 0,
		streaming     BOOLEAN DEFAULT 0,
		error         TEXT DEFAULT ''
	);

	CREATE INDEX IF NOT EXISTS idx_api_calls_session ON api_calls(session_id);
	CREATE INDEX IF NOT EXISTS idx_api_calls_started ON api_calls(started_at DESC);
	`
	_, err := s.db.Exec(schema)
	return err
}

// CreateSession inserts a new session record.
func (s *Store) CreateSession(sess *Session) error {
	_, err := s.db.Exec(
		`INSERT INTO sessions (id, started_at) VALUES (?, ?)`,
		sess.ID, sess.StartedAt,
	)
	return err
}

// InsertCall inserts a completed API call record.
func (s *Store) InsertCall(call *APICall) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO api_calls
			(id, session_id, provider, model, endpoint, method,
			 request_body, response_body, status_code, started_at,
			 duration_ms, input_tokens, output_tokens, cost, streaming, error)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		call.ID, call.SessionID, call.Provider, call.Model, call.Endpoint, call.Method,
		call.RequestBody, call.ResponseBody, call.StatusCode, call.StartedAt,
		call.Duration.Milliseconds(), call.InputTokens, call.OutputTokens,
		call.Cost, call.Streaming, call.Error,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		UPDATE sessions SET
			total_cost = total_cost + ?,
			total_tokens = total_tokens + ? + ?,
			request_count = request_count + 1
		WHERE id = ?`,
		call.Cost, call.InputTokens, call.OutputTokens, call.SessionID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetSession returns the current session stats.
func (s *Store) GetSession(id string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	row := s.db.QueryRow(`SELECT id, started_at, total_cost, total_tokens, request_count FROM sessions WHERE id = ?`, id)
	sess := &Session{}
	err := row.Scan(&sess.ID, &sess.StartedAt, &sess.TotalCost, &sess.TotalTokens, &sess.RequestCount)
	if err != nil {
		return nil, err
	}
	return sess, nil
}

// ListCalls returns recent API calls for the current session, newest first.
func (s *Store) ListCalls(sessionID string, limit, offset int) ([]APICall, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 100
	}

	rows, err := s.db.Query(`
		SELECT id, session_id, provider, model, endpoint, method,
		       status_code, started_at, duration_ms, input_tokens, output_tokens,
		       cost, streaming, error
		FROM api_calls
		WHERE session_id = ?
		ORDER BY started_at DESC
		LIMIT ? OFFSET ?`, sessionID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var calls []APICall
	for rows.Next() {
		var c APICall
		var durMs int64
		err := rows.Scan(
			&c.ID, &c.SessionID, &c.Provider, &c.Model, &c.Endpoint, &c.Method,
			&c.StatusCode, &c.StartedAt, &durMs, &c.InputTokens, &c.OutputTokens,
			&c.Cost, &c.Streaming, &c.Error,
		)
		if err != nil {
			return nil, err
		}
		c.Duration = time.Duration(durMs) * time.Millisecond
		calls = append(calls, c)
	}
	return calls, rows.Err()
}

// GetCallDetail returns full request/response bodies for a single call.
func (s *Store) GetCallDetail(callID string) (*APICallDetail, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	row := s.db.QueryRow(`
		SELECT id, session_id, provider, model, endpoint, method,
		       request_body, response_body, status_code, started_at,
		       duration_ms, input_tokens, output_tokens, cost, streaming, error
		FROM api_calls WHERE id = ?`, callID)

	var d APICallDetail
	var durMs int64
	var reqBody, respBody []byte
	err := row.Scan(
		&d.ID, &d.SessionID, &d.Provider, &d.Model, &d.Endpoint, &d.Method,
		&reqBody, &respBody, &d.StatusCode, &d.StartedAt,
		&durMs, &d.InputTokens, &d.OutputTokens, &d.Cost, &d.Streaming, &d.Error,
	)
	if err != nil {
		return nil, err
	}
	d.Duration = time.Duration(durMs) * time.Millisecond
	d.RequestBody = string(reqBody)
	d.ResponseBody = string(respBody)
	return &d, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}
