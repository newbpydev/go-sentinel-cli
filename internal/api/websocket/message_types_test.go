package websocket

import (
	"encoding/json"
	"testing"
)

type TestMessage struct {
	Type string `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type TestResultPayload struct {
	TestID string `json:"test_id"`
	Status string `json:"status"`
}

type CommandPayload struct {
	Command string `json:"command"`
	Args []string `json:"args"`
}

func TestMessageEncodingDecoding(t *testing.T) {
	payload := TestResultPayload{TestID: "abc", Status: "pass"}
	payloadBytes, err := json.Marshal(payload)
	if err != nil { t.Fatal(err) }
	msg := TestMessage{Type: "test_result", Payload: payloadBytes}
	msgBytes, err := json.Marshal(msg)
	if err != nil { t.Fatal(err) }
	var decoded TestMessage
	if err := json.Unmarshal(msgBytes, &decoded); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if decoded.Type != "test_result" {
		t.Errorf("expected type 'test_result', got %s", decoded.Type)
	}
	var decodedPayload TestResultPayload
	if err := json.Unmarshal(decoded.Payload, &decodedPayload); err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}
	if decodedPayload.TestID != "abc" || decodedPayload.Status != "pass" {
		t.Errorf("payload mismatch: %+v", decodedPayload)
	}
}

func TestHandleDifferentMessageTypes(t *testing.T) {
	msg1 := TestMessage{Type: "test_result", Payload: []byte(`{"test_id":"t1","status":"fail"}`)}
	msg2 := TestMessage{Type: "command", Payload: []byte(`{"command":"run","args":["suite"]}`)}
	// Simulate routing
	switch msg1.Type {
	case "test_result":
		var p TestResultPayload
		if err := json.Unmarshal(msg1.Payload, &p); err != nil { t.Fatal(err) }
		if p.Status != "fail" { t.Errorf("expected fail, got %s", p.Status) }
	case "command":
		t.Error("should not route to command")
	}
	switch msg2.Type {
	case "command":
		var p CommandPayload
		if err := json.Unmarshal(msg2.Payload, &p); err != nil { t.Fatal(err) }
		if p.Command != "run" { t.Errorf("expected run, got %s", p.Command) }
	case "test_result":
		t.Error("should not route to test_result")
	}
}

func TestMalformedMessageHandling(t *testing.T) {
	bad := []byte(`{"type":"test_result","payload":"notjson"}`)
	var msg TestMessage
	err := json.Unmarshal(bad, &msg)
	if err != nil {
		// Should still decode the top-level, but payload is not valid JSON
		return
	}
	var payload TestResultPayload
	err = json.Unmarshal(msg.Payload, &payload)
	if err == nil {
		t.Error("expected error when decoding malformed payload")
	}
}
