package handlers

import (
	"encoding/json"
	"net/http"
)

type ConfigStore interface {
	Get() map[string]interface{}
	Set(cfg map[string]interface{}) error
	Validate(cfg map[string]interface{}) error
}

type ConfigHandler struct {
	store ConfigStore
}

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

func (h *ConfigHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	cfg := h.store.Get()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cfg)
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
