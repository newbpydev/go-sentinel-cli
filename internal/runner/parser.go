package runner

import (
	"encoding/json"
	"io"
)

// TestEvent represents a single event from 'go test -json' output.
type TestEvent struct {
	Time    string  `json:"Time"`
	Action  string  `json:"Action"`
	Package string  `json:"Package"`
	Test    string  `json:"Test,omitempty"`
	Output  string  `json:"Output,omitempty"`
	Elapsed float64 `json:"Elapsed,omitempty"`
}

// ParseTestEvents reads a stream of JSON lines from r and returns parsed TestEvents.
func ParseTestEvents(r io.Reader) ([]TestEvent, error) {
	var events []TestEvent
	dec := json.NewDecoder(r)
	for {
		var ev TestEvent
		if err := dec.Decode(&ev); err != nil {
			if err == io.EOF {
				break
			}
			return events, err
		}
		events = append(events, ev)
	}
	return events, nil
}
