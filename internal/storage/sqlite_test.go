package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func tempDB(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")
	s, err := New(path)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestCreateAndGetSession(t *testing.T) {
	s := tempDB(t)

	sess := &Session{
		ID:        "test-session-1",
		StartedAt: time.Now(),
	}
	if err := s.CreateSession(sess); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	got, err := s.GetSession("test-session-1")
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if got.ID != "test-session-1" {
		t.Errorf("session ID: got %q, want %q", got.ID, "test-session-1")
	}
	if got.TotalCost != 0 {
		t.Errorf("initial cost should be 0, got %f", got.TotalCost)
	}
	if got.RequestCount != 0 {
		t.Errorf("initial request count should be 0, got %d", got.RequestCount)
	}
}

func TestInsertCallAndList(t *testing.T) {
	s := tempDB(t)

	sess := &Session{ID: "s1", StartedAt: time.Now()}
	s.CreateSession(sess)

	call := &APICall{
		ID:           "call-1",
		SessionID:    "s1",
		Provider:     ProviderOpenAI,
		Model:        "gpt-4o",
		Endpoint:     "/v1/chat/completions",
		Method:       "POST",
		RequestBody:  []byte(`{"model":"gpt-4o"}`),
		ResponseBody: []byte(`{"choices":[]}`),
		StatusCode:   200,
		StartedAt:    time.Now(),
		DurationMs:   1500,
		InputTokens:  100,
		OutputTokens: 50,
		Cost:         0.00075,
		Streaming:    false,
	}

	if err := s.InsertCall(call); err != nil {
		t.Fatalf("InsertCall: %v", err)
	}

	// Verify session was updated
	got, _ := s.GetSession("s1")
	if got.RequestCount != 1 {
		t.Errorf("request count: got %d, want 1", got.RequestCount)
	}
	if got.TotalTokens != 150 {
		t.Errorf("total tokens: got %d, want 150", got.TotalTokens)
	}
	if got.TotalCost != 0.00075 {
		t.Errorf("total cost: got %f, want 0.00075", got.TotalCost)
	}

	// List calls
	calls, err := s.ListCalls("s1", 10, 0)
	if err != nil {
		t.Fatalf("ListCalls: %v", err)
	}
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Model != "gpt-4o" {
		t.Errorf("model: got %q, want %q", calls[0].Model, "gpt-4o")
	}
	if calls[0].DurationMs != 1500 {
		t.Errorf("duration: got %d, want 1500", calls[0].DurationMs)
	}
}

func TestGetCallDetail(t *testing.T) {
	s := tempDB(t)

	sess := &Session{ID: "s1", StartedAt: time.Now()}
	s.CreateSession(sess)

	reqBody := `{"model":"gpt-4o","messages":[{"role":"user","content":"hello"}]}`
	respBody := `{"choices":[{"message":{"content":"hi there"}}],"usage":{"prompt_tokens":5,"completion_tokens":3}}`

	call := &APICall{
		ID:           "call-detail-1",
		SessionID:    "s1",
		Provider:     ProviderOpenAI,
		Model:        "gpt-4o",
		Endpoint:     "/v1/chat/completions",
		Method:       "POST",
		RequestBody:  []byte(reqBody),
		ResponseBody: []byte(respBody),
		StatusCode:   200,
		StartedAt:    time.Now(),
		DurationMs:   800,
		InputTokens:  5,
		OutputTokens: 3,
		Cost:         0.0001,
	}
	s.InsertCall(call)

	detail, err := s.GetCallDetail("call-detail-1")
	if err != nil {
		t.Fatalf("GetCallDetail: %v", err)
	}
	if detail.RequestBody != reqBody {
		t.Errorf("request body mismatch")
	}
	if detail.ResponseBody != respBody {
		t.Errorf("response body mismatch")
	}
}

func TestListCallsOrdering(t *testing.T) {
	s := tempDB(t)

	sess := &Session{ID: "s1", StartedAt: time.Now()}
	s.CreateSession(sess)

	now := time.Now()
	for i := 0; i < 5; i++ {
		call := &APICall{
			ID:        "call-" + string(rune('a'+i)),
			SessionID: "s1",
			Provider:  ProviderOpenAI,
			Model:     "gpt-4o",
			Endpoint:  "/v1/chat/completions",
			Method:    "POST",
			StartedAt: now.Add(time.Duration(i) * time.Second),
		}
		s.InsertCall(call)
	}

	calls, _ := s.ListCalls("s1", 10, 0)
	if len(calls) != 5 {
		t.Fatalf("expected 5 calls, got %d", len(calls))
	}

	// Should be newest first
	for i := 1; i < len(calls); i++ {
		if calls[i].StartedAt.After(calls[i-1].StartedAt) {
			t.Errorf("calls not sorted newest first at index %d", i)
		}
	}
}

func TestListCallsLimitOffset(t *testing.T) {
	s := tempDB(t)

	sess := &Session{ID: "s1", StartedAt: time.Now()}
	s.CreateSession(sess)

	now := time.Now()
	for i := 0; i < 10; i++ {
		call := &APICall{
			ID:        "call-" + string(rune('a'+i)),
			SessionID: "s1",
			Provider:  ProviderOpenAI,
			Model:     "gpt-4o",
			Endpoint:  "/v1/chat/completions",
			Method:    "POST",
			StartedAt: now.Add(time.Duration(i) * time.Second),
		}
		s.InsertCall(call)
	}

	calls, _ := s.ListCalls("s1", 3, 0)
	if len(calls) != 3 {
		t.Errorf("limit 3: got %d calls", len(calls))
	}

	calls, _ = s.ListCalls("s1", 3, 5)
	if len(calls) != 3 {
		t.Errorf("limit 3 offset 5: got %d calls", len(calls))
	}
}

func TestMultipleCallsAccumulateSession(t *testing.T) {
	s := tempDB(t)

	sess := &Session{ID: "s1", StartedAt: time.Now()}
	s.CreateSession(sess)

	for i := 0; i < 3; i++ {
		call := &APICall{
			ID:           "call-" + string(rune('a'+i)),
			SessionID:    "s1",
			Provider:     ProviderOpenAI,
			Model:        "gpt-4o",
			Endpoint:     "/v1/chat/completions",
			Method:       "POST",
			StartedAt:    time.Now(),
			InputTokens:  100,
			OutputTokens: 50,
			Cost:         0.01,
		}
		s.InsertCall(call)
	}

	got, _ := s.GetSession("s1")
	if got.RequestCount != 3 {
		t.Errorf("request count: got %d, want 3", got.RequestCount)
	}
	if got.TotalTokens != 450 {
		t.Errorf("total tokens: got %d, want 450", got.TotalTokens)
	}
	if got.TotalCost < 0.029 || got.TotalCost > 0.031 {
		t.Errorf("total cost: got %f, want ~0.03", got.TotalCost)
	}
}

func TestDefaultDBPath(t *testing.T) {
	// Just verify it doesn't crash — uses $HOME/.llmview/
	s, err := New("")
	if err != nil {
		t.Fatalf("New with empty path: %v", err)
	}
	s.Close()

	// Clean up
	home, _ := os.UserHomeDir()
	os.Remove(filepath.Join(home, ".llmview", "llmview.db"))
}
