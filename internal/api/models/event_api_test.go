package models

import (
	"encoding/json"
	"testing"
	"time"

	runner "github.com/newbpydev/go-sentinel/internal/runner"
)

func TestConvertRunnerTestEventToAPIModel(t *testing.T) {
	in := runner.TestEvent{
		Time:    "2025-05-12T10:00:00Z",
		Action:  "pass",
		Package: "github.com/newbpydev/go-sentinel/internal/ui",
		Test:    "TestFoo",
		Output:  "",
		Elapsed: 0.12,
	}
	api := APITestEvent{
		Time:    in.Time,
		Action:  in.Action,
		Package: in.Package,
		Test:    in.Test,
		Output:  in.Output,
		Elapsed: in.Elapsed,
	}
	b, err := json.Marshal(api)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out APITestEvent
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.Action != "pass" || out.Test != "TestFoo" || out.Elapsed != 0.12 {
		t.Errorf("unexpected output: %+v", out)
	}
}

func TestAPITestEventHandlesAllTypes(t *testing.T) {
	cases := []runner.TestEvent{
		{Action: "run", Test: "TestBar"},
		{Action: "fail", Test: "TestBaz", Output: "failure msg"},
		{Action: "output", Output: "some output"},
	}
	for _, c := range cases {
		api := APITestEvent{
			Time:    c.Time,
			Action:  c.Action,
			Package: c.Package,
			Test:    c.Test,
			Output:  c.Output,
			Elapsed: c.Elapsed,
		}
		b, err := json.Marshal(api)
		if err != nil {
			t.Errorf("marshal: %v", err)
		}
		var out APITestEvent
		if err := json.Unmarshal(b, &out); err != nil {
			t.Errorf("unmarshal: %v", err)
		}
		if out.Action != c.Action || out.Test != c.Test || out.Output != c.Output {
			t.Errorf("mismatch: got %+v want %+v", out, c)
		}
	}
}

func TestAPITestEventNestedSerialization(t *testing.T) {
	parent := APITestEvent{
		Time:    "2025-05-12T10:00:00Z",
		Action:  "run",
		Package: "pkg",
		Test:    "TestParent",
	}
	child := APITestEvent{
		Time:    "2025-05-12T10:00:01Z",
		Action:  "pass",
		Package: "pkg",
		Test:    "TestChild",
		Elapsed: 0.05,
	}
	wrapped := struct {
		Parent   APITestEvent   `json:"parent"`
		Children []APITestEvent `json:"children"`
	}{
		Parent:   parent,
		Children: []APITestEvent{child},
	}
	b, err := json.Marshal(wrapped)
	if err != nil {
		t.Fatalf("marshal nested: %v", err)
	}
	var out struct {
		Parent   APITestEvent   `json:"parent"`
		Children []APITestEvent `json:"children"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal nested: %v", err)
	}
	if len(out.Children) != 1 || out.Children[0].Test != "TestChild" {
		t.Errorf("nested serialization failed: %+v", out)
	}
}

func TestConvertRunnerTestEventsToAPI(t *testing.T) {
	events := []runner.TestEvent{
		{
			Time:    "2025-05-12T10:00:00Z",
			Action:  "run",
			Package: "pkg/foo",
			Test:    "TestA",
		},
		{
			Time:    "2025-05-12T10:00:01Z",
			Action:  "pass",
			Package: "pkg/foo",
			Test:    "TestA",
			Elapsed: 0.1,
		},
		{
			Time:    "2025-05-12T10:00:02Z",
			Action:  "run",
			Package: "pkg/bar",
			Test:    "TestB",
		},
		{
			Time:    "2025-05-12T10:00:03Z",
			Action:  "fail",
			Package: "pkg/bar",
			Test:    "TestB",
			Output:  "test failed",
			Elapsed: 0.2,
		},
	}

	apiEvents := ConvertRunnerTestEventsToAPI(events)

	if len(apiEvents) != len(events) {
		t.Errorf("expected %d events, got %d", len(events), len(apiEvents))
	}

	// Check conversion of each event
	for i, ev := range events {
		api := apiEvents[i]
		if api.Time != ev.Time {
			t.Errorf("event %d: expected time %s, got %s", i, ev.Time, api.Time)
		}
		if api.Action != ev.Action {
			t.Errorf("event %d: expected action %s, got %s", i, ev.Action, api.Action)
		}
		if api.Package != ev.Package {
			t.Errorf("event %d: expected package %s, got %s", i, ev.Package, api.Package)
		}
		if api.Test != ev.Test {
			t.Errorf("event %d: expected test %s, got %s", i, ev.Test, api.Test)
		}
		if api.Output != ev.Output {
			t.Errorf("event %d: expected output %s, got %s", i, ev.Output, api.Output)
		}
		if api.Elapsed != ev.Elapsed {
			t.Errorf("event %d: expected elapsed %f, got %f", i, ev.Elapsed, api.Elapsed)
		}
	}
}

func TestAPITestEvent_JSONEncoding(t *testing.T) {
	tests := []struct {
		name     string
		event    APITestEvent
		wantJSON string
	}{
		{
			name: "full event",
			event: APITestEvent{
				Time:    "2025-05-12T10:00:00Z",
				Action:  "pass",
				Package: "pkg/foo",
				Test:    "TestA",
				Output:  "test output",
				Elapsed: 0.123,
			},
			wantJSON: `{"time":"2025-05-12T10:00:00Z","action":"pass","package":"pkg/foo","test":"TestA","output":"test output","elapsed":0.123}`,
		},
		{
			name: "minimal event",
			event: APITestEvent{
				Action:  "run",
				Package: "pkg/foo",
			},
			wantJSON: `{"time":"","action":"run","package":"pkg/foo","test":"","output":"","elapsed":0}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.event)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if string(b) != tc.wantJSON {
				t.Errorf("got JSON %s, want %s", string(b), tc.wantJSON)
			}

			// Test round-trip
			var decoded APITestEvent
			if err := json.Unmarshal(b, &decoded); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if decoded != tc.event {
				t.Errorf("round-trip: got %+v, want %+v", decoded, tc.event)
			}
		})
	}
}

func TestAPITestEvent_TimeValidation(t *testing.T) {
	tests := []struct {
		name    string
		time    string
		wantErr bool
	}{
		{
			name:    "valid RFC3339",
			time:    "2025-05-12T10:00:00Z",
			wantErr: false,
		},
		{
			name:    "valid with timezone",
			time:    "2025-05-12T10:00:00+02:00",
			wantErr: false,
		},
		{
			name:    "invalid format",
			time:    "2025-05-12 10:00:00",
			wantErr: true,
		},
		{
			name:    "empty string",
			time:    "",
			wantErr: false, // Empty is allowed for events that don't need time
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			event := APITestEvent{Time: tc.time}
			if tc.time != "" {
				_, err := time.Parse(time.RFC3339, event.Time)
				if tc.wantErr {
					if err == nil {
						t.Error("expected error parsing time")
					}
				} else {
					if err != nil {
						t.Errorf("unexpected error parsing time: %v", err)
					}
				}
			}
		})
	}
}
