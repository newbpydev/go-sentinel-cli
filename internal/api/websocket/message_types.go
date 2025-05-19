package websocket

import (
	"encoding/json"
	"errors"
)

// MessageType represents the type of WebSocket message being sent or received.
// It's used to categorize messages for proper routing and handling.
type MessageType string

// WebSocket message type constants
const (
	// MessageTypeTestResult is used for messages containing test execution results
	MessageTypeTestResult MessageType = "test_result"
	// MessageTypeCommand is used for messages containing commands to be executed
	MessageTypeCommand    MessageType = "command"
)

// TestMessage represents the standard message format for WebSocket communication.
// It contains a type field to identify the message category and a payload with the actual data.
type TestMessage struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// TestResultPayload contains information about a test execution result.
// It includes the test identifier and its execution status (pass/fail).
type TestResultPayload struct {
	TestID string `json:"test_id"`
	Status string `json:"status"`
}

// CommandPayload contains information about a command to be executed.
// It includes the command name and any arguments needed for execution.
type CommandPayload struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

// EncodeMessage serializes a message with the given type and payload into JSON format
// for transmission over a WebSocket connection.
func EncodeMessage(msgType MessageType, payload interface{}) ([]byte, error) {
	p, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	msg := TestMessage{Type: msgType, Payload: p}
	return json.Marshal(msg)
}

// DecodeMessage deserializes a JSON message received over a WebSocket connection
// and returns its type and payload. The payload type depends on the message type.
func DecodeMessage(data []byte) (MessageType, interface{}, error) {
	var msg TestMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return "", nil, err
	}
	switch msg.Type {
	case MessageTypeTestResult:
		var p TestResultPayload
		if err := json.Unmarshal(msg.Payload, &p); err != nil {
			return msg.Type, nil, err
		}
		return msg.Type, p, nil
	case MessageTypeCommand:
		var p CommandPayload
		if err := json.Unmarshal(msg.Payload, &p); err != nil {
			return msg.Type, nil, err
		}
		return msg.Type, p, nil
	default:
		return msg.Type, nil, errors.New("unknown message type")
	}
}
