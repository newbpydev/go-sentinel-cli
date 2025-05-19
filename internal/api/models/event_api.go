package models

import (
	runner "github.com/newbpydev/go-sentinel/internal/runner"
)

// APITestEvent is the API-facing version of runner.TestEvent
// with JSON tags suitable for public API responses.
type APITestEvent struct {
	Time    string  `json:"time"`
	Action  string  `json:"action"`
	Package string  `json:"package"`
	Test    string  `json:"test"`
	Output  string  `json:"output"`
	Elapsed float64 `json:"elapsed"`
}

// ConvertRunnerTestEventToAPI converts a runner.TestEvent to APITestEvent
func ConvertRunnerTestEventToAPI(ev runner.TestEvent) APITestEvent {
	return APITestEvent{
		Time:    ev.Time,
		Action:  ev.Action,
		Package: ev.Package,
		Test:    ev.Test,
		Output:  ev.Output,
		Elapsed: ev.Elapsed,
	}
}

// ConvertRunnerTestEventsToAPI converts a slice of runner.TestEvent to []APITestEvent
func ConvertRunnerTestEventsToAPI(events []runner.TestEvent) []APITestEvent {
	apiEvents := make([]APITestEvent, len(events))
	for i, ev := range events {
		apiEvents[i] = ConvertRunnerTestEventToAPI(ev)
	}
	return apiEvents
}
