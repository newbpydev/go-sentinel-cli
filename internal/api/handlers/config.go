// Package handlers provides HTTP request handlers for the API
package handlers

import (
	"encoding/json"
	"net/http"
)

// ConfigStore defines the interface for configuration storage
type ConfigStore interface {
	Get() map[string]interface{}
	Set(cfg map[string]interface{}) error
	Validate(cfg map[string]interface{}) error
}

// ConfigHandler handles configuration-related HTTP requests
type ConfigHandler struct {
	store ConfigStore
}

// NewConfigHandler creates a new ConfigHandler with the given store
func NewConfigHandler(store ConfigStore) *ConfigHandler {
	return &ConfigHandler{store: store}
}

func (h *ConfigHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, r)
	case http.MethodPost:
		h.handleSet(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGet handles GET requests for configuration
func (h *ConfigHandler) handleGet(w http.ResponseWriter, _ *http.Request) {
	cfg := h.store.Get()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(cfg); err != nil {
		http.Error(w, "failed to encode config", http.StatusInternalServerError)
		return
	}
}

func (h *ConfigHandler) handleSet(w http.ResponseWriter, r *http.Request) {
	var cfg map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if err := h.store.Validate(cfg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.store.Set(cfg); err != nil {
		http.Error(w, "set failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
