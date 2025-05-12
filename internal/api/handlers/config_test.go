package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type mockConfigStore struct {
	cfg map[string]interface{}
}
func (m *mockConfigStore) Get() map[string]interface{} {
	return m.cfg
}
func (m *mockConfigStore) Set(cfg map[string]interface{}) error {
	m.cfg = cfg
	return nil
}
func (m *mockConfigStore) Validate(cfg map[string]interface{}) error {
	if cfg["invalid"] == true {
		return &ConfigError{"invalid config"}
	}
	return nil
}

type ConfigError struct{ msg string }
func (e *ConfigError) Error() string { return e.msg }

func TestGetConfiguration(t *testing.T) {
	store := &mockConfigStore{cfg: map[string]interface{}{"foo": "bar"}}
	h := NewConfigHandler(store)
	r := httptest.NewRequest("GET", "/api/config", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	if !reflect.DeepEqual(resp, store.cfg) {
		t.Errorf("expected %v, got %v", store.cfg, resp)
	}
}

func TestUpdateConfiguration(t *testing.T) {
	store := &mockConfigStore{cfg: map[string]interface{}{}}
	h := NewConfigHandler(store)
	body := bytes.NewBufferString(`{"foo":"baz"}`)
	r := httptest.NewRequest("POST", "/api/config", body)
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if store.cfg["foo"] != "baz" {
		t.Errorf("expected foo to be 'baz', got %v", store.cfg["foo"])
	}
}

func TestValidateConfiguration(t *testing.T) {
	store := &mockConfigStore{cfg: map[string]interface{}{}}
	h := NewConfigHandler(store)
	body := bytes.NewBufferString(`{"invalid":true}`)
	r := httptest.NewRequest("POST", "/api/config", body)
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid config, got %d", w.Code)
	}
}
