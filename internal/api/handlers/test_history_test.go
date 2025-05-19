package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockHistoryStore struct {
	runs []TestRun
}

func (m *mockHistoryStore) Add(run TestRun) error {
	m.runs = append(m.runs, run)
	return nil
}

func (m *mockHistoryStore) GetRecent(limit, offset int) ([]TestRun, error) {
	if offset > len(m.runs) {
		return []TestRun{}, nil
	}
	end := offset + limit
	if end > len(m.runs) {
		end = len(m.runs)
	}
	return m.runs[offset:end], nil
}

func TestGetTestHistory_Success(t *testing.T) {
	hs := &mockHistoryStore{
		runs: []TestRun{
			{ID: "1"},
			{ID: "2"},
		},
	}
	h := NewTestHistoryHandler(hs)
	r := httptest.NewRequest("GET", "/api/test-history?limit=2", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp []TestRun
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	if len(resp) != 2 || resp[0].ID != "1" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestGetTestHistory_Pagination(t *testing.T) {
	hs := &mockHistoryStore{
		runs: []TestRun{
			{ID: "1"},
			{ID: "2"},
			{ID: "3"},
		},
	}
	h := NewTestHistoryHandler(hs)
	r := httptest.NewRequest("GET", "/api/test-history?limit=2&offset=1", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp []TestRun
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	if len(resp) != 2 || resp[0].ID != "2" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestGetTestHistory_Error(t *testing.T) {
	hs := &mockHistoryStore{}
	h := NewTestHistoryHandler(hs)
	r := httptest.NewRequest("GET", "/api/test-history?limit=bad", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for bad limit, got %d", w.Code)
	}
}
