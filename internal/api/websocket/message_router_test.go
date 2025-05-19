package websocket

import (
	"errors"
	"testing"
)

func TestRouter_MessageRouting(t *testing.T) {
	r := NewRouter()
	hit := false
	r.Register(MessageTypeTestResult, func(payload interface{}) error {
		p, ok := payload.(TestResultPayload)
		if !ok || p.Status != "pass" {
			t.Errorf("unexpected payload: %+v", payload)
		}
		hit = true
		return nil
	})
	err := r.Route(MessageTypeTestResult, TestResultPayload{TestID: "t1", Status: "pass"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !hit {
		t.Error("handler not called")
	}
}

func TestRouter_UnknownType(t *testing.T) {
	r := NewRouter()
	err := r.Route("unknown", nil)
	if err == nil {
		t.Error("expected error for unknown type")
	}
}

func TestRouter_HandlerErrorPropagates(t *testing.T) {
	r := NewRouter()
	r.Register(MessageTypeCommand, func(_ interface{}) error {
		return errors.New("fail")
	})
	err := r.Route(MessageTypeCommand, CommandPayload{Command: "run"})
	if err == nil || err.Error() != "fail" {
		t.Errorf("expected fail error, got %v", err)
	}
}
