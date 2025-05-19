package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type TestRun struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type HistoryStore interface {
	GetRecent(limit, offset int) ([]TestRun, error)
}

type TestHistoryHandler struct {
	store HistoryStore
}

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
