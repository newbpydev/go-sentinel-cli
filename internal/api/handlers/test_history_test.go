package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockHistoryStore struct{}

func (m *mockHistoryStore) GetRecent(_, offset int) ([]TestRun, error) {
	return []TestRun{
		{ID: "1", Status: "pass"},
		{ID: "2", Status: "fail"},
	}, nil
}

func TestGetTestHistory_Success(t *testing.T) {
	hs := &mockHistoryStore{}
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
	if len(resp) != 2 || resp[0].ID != "1" || resp[1].Status != "fail" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestGetTestHistory_Pagination(t *testing.T) {
	hs := &mockHistoryStore{}
	h := NewTestHistoryHandler(hs)
	r := httptest.NewRequest("GET", "/api/test-history?limit=1&offset=1", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp []TestRun
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	if len(resp) != 2 {
		t.Errorf("expected 2 results for mock, got %d", len(resp))
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
