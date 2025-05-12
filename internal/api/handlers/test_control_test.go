package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockRunner struct {
	triggered []string
	cancelled []string
}
func (m *mockRunner) Trigger(tests []string) error {
	m.triggered = append(m.triggered, tests...)
	return nil
}
func (m *mockRunner) Cancel(ids []string) error {
	m.cancelled = append(m.cancelled, ids...)
	return nil
}

func TestTriggerNewTestRun(t *testing.T) {
	runner := &mockRunner{}
	h := NewTestControlHandler(runner)
	body := bytes.NewBufferString(`{"tests":["TestA","TestB"]}`)
	r := httptest.NewRequest("POST", "/api/test-control/trigger", body)
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if len(runner.triggered) != 2 || runner.triggered[0] != "TestA" {
		t.Errorf("unexpected triggered: %+v", runner.triggered)
	}
}

func TestTriggerTestRun_Filter(t *testing.T) {
	runner := &mockRunner{}
	h := NewTestControlHandler(runner)
	body := bytes.NewBufferString(`{"tests":["TestX"]}`)
	r := httptest.NewRequest("POST", "/api/test-control/trigger", body)
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if len(runner.triggered) != 1 || runner.triggered[0] != "TestX" {
		t.Errorf("unexpected triggered: %+v", runner.triggered)
	}
}

func TestCancelRunningTests(t *testing.T) {
	runner := &mockRunner{}
	h := NewTestControlHandler(runner)
	body := bytes.NewBufferString(`{"cancel":["run-123"]}`)
	r := httptest.NewRequest("POST", "/api/test-control/cancel", body)
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if len(runner.cancelled) != 1 || runner.cancelled[0] != "run-123" {
		t.Errorf("unexpected cancelled: %+v", runner.cancelled)
	}
}

func TestTriggerTestRun_BadRequest(t *testing.T) {
	runner := &mockRunner{}
	h := NewTestControlHandler(runner)
	body := bytes.NewBufferString(`notjson`)
	r := httptest.NewRequest("POST", "/api/test-control/trigger", body)
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for bad json, got %d", w.Code)
	}
}
