package websocket

import (
	"encoding/json"
	"errors"
)

type MessageType string

const (
	MessageTypeTestResult MessageType = "test_result"
	MessageTypeCommand    MessageType = "command"
)

type TestMessage struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type TestResultPayload struct {
	TestID string `json:"test_id"`
	Status string `json:"status"`
}

type CommandPayload struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

func EncodeMessage(msgType MessageType, payload interface{}) ([]byte, error) {
	p, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	msg := TestMessage{Type: msgType, Payload: p}
	return json.Marshal(msg)
}

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
