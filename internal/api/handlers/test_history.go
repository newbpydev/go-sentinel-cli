package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

// TestRun represents a single execution of tests.
type TestRun struct {
	ID           string       `json:"id"`
	Timestamp    time.Time    `json:"timestamp"`
	TotalTests   int          `json:"totalTests"`
	PassedTests  int          `json:"passedTests"`
	FailedTests  int          `json:"failedTests"`
	TotalTime    string       `json:"totalTime"` // Duration as string for display
	TestResults  []TestResult `json:"testResults"`
	Branch       string       `json:"branch,omitempty"`
	Commit       string       `json:"commit,omitempty"`
	TriggeredBy  string       `json:"triggeredBy,omitempty"`
	BuildVersion string       `json:"buildVersion,omitempty"`
}

// TestResult represents a single test result (minimal stub for test history).
type TestResult struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// HistoryStore defines the interface for test run history storage.
type HistoryStore interface {
	Add(run TestRun) error
	GetRecent(limit, offset int) ([]TestRun, error)
}

// TestHistoryHandler handles test run history operations.
type TestHistoryHandler struct {
	store HistoryStore
}

// NewTestHistoryHandler creates a new TestHistoryHandler.
func NewTestHistoryHandler(store HistoryStore) *TestHistoryHandler {
	return &TestHistoryHandler{store: store}
}

func (h *TestHistoryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	limit := 10
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		v, err := strconv.Atoi(l)
		if err != nil || v < 1 {
			http.Error(w, "bad limit", http.StatusBadRequest)
			return
		}
		limit = v
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		v, err := strconv.Atoi(o)
		if err != nil || v < 0 {
			http.Error(w, "bad offset", http.StatusBadRequest)
			return
		}
		offset = v
	}
	runs, err := h.store.GetRecent(limit, offset)
	if err != nil {
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(runs); err != nil {
		http.Error(w, "failed to encode test runs", http.StatusInternalServerError)
		return
	}
}
