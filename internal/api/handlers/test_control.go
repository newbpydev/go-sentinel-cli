package handlers

import (
	"encoding/json"
	"net/http"
)

type TestRunner interface {
	Trigger(tests []string) error
	Cancel(ids []string) error
}

type TestControlHandler struct {
	runner TestRunner
}

func NewTestControlHandler(runner TestRunner) *TestControlHandler {
	return &TestControlHandler{runner: runner}
}

func (h *TestControlHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.URL.Path == "/api/test-control/trigger" {
		h.handleTrigger(w, r)
		return
	}
	if r.URL.Path == "/api/test-control/cancel" {
		h.handleCancel(w, r)
		return
	}
	http.NotFound(w, r)
}

func (h *TestControlHandler) handleTrigger(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Tests []string `json:"tests"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if err := h.runner.Trigger(req.Tests); err != nil {
		http.Error(w, "trigger failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *TestControlHandler) handleCancel(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Cancel []string `json:"cancel"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if err := h.runner.Cancel(req.Cancel); err != nil {
		http.Error(w, "cancel failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
