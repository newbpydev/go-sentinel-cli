package runner

import (
	"strings"
	"testing"
)

func TestParseTestEvents_SimpleStream(t *testing.T) {
	input := `{"Time":"2025-05-08T13:03:22.67","Action":"run","Package":"github.com/yourusername/go-sentinel/internal/runner/testdata/passonly","Test":"TestAlwaysPass"}
{"Time":"2025-05-08T13:03:22.68","Action":"output","Package":"github.com/yourusername/go-sentinel/internal/runner/testdata/passonly","Test":"TestAlwaysPass","Output":"=== RUN   TestAlwaysPass\n"}
{"Time":"2025-05-08T13:03:22.69","Action":"pass","Package":"github.com/yourusername/go-sentinel/internal/runner/testdata/passonly","Test":"TestAlwaysPass","Elapsed":0}`
	r := strings.NewReader(input)
	events, err := ParseTestEvents(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
	if events[0].Action != "run" || events[1].Action != "output" || events[2].Action != "pass" {
		t.Errorf("unexpected event actions: %+v", events)
	}
	if events[2].Test != "TestAlwaysPass" {
		t.Errorf("expected TestAlwaysPass, got %s", events[2].Test)
	}
}

func TestParseTestEvents_TrackAllActions(t *testing.T) {
	input := `{"Time":"2025-05-08T13:03:22.67","Action":"start","Package":"pkg"}
{"Time":"2025-05-08T13:03:22.68","Action":"run","Package":"pkg","Test":"TestA"}
{"Time":"2025-05-08T13:03:22.69","Action":"output","Package":"pkg","Test":"TestA","Output":"=== RUN   TestA\n"}
{"Time":"2025-05-08T13:03:22.70","Action":"pass","Package":"pkg","Test":"TestA","Elapsed":0.002}
{"Time":"2025-05-08T13:03:22.71","Action":"run","Package":"pkg","Test":"TestB"}
{"Time":"2025-05-08T13:03:22.72","Action":"fail","Package":"pkg","Test":"TestB","Elapsed":0.003}
{"Time":"2025-05-08T13:03:22.73","Action":"output","Package":"pkg","Test":"TestB","Output":"--- FAIL: TestB (0.00s)\n"}`
	r := strings.NewReader(input)
	events, err := ParseTestEvents(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 7 {
		t.Fatalf("expected 7 events, got %d", len(events))
	}
	wantActions := []string{"start", "run", "output", "pass", "run", "fail", "output"}
	for i, act := range wantActions {
		if events[i].Action != act {
			t.Errorf("event %d: want action %s, got %s", i, act, events[i].Action)
		}
	}
	if events[3].Test != "TestA" || events[5].Test != "TestB" {
		t.Errorf("expected TestA and TestB in correct places, got %+v", events)
	}
}

func TestParseTestEvents_ExtractFileLineFromFailureOutput(t *testing.T) {
	input := `{"Time":"2025-05-08T13:03:22.72","Action":"fail","Package":"pkg","Test":"TestB","Elapsed":0.003}
{"Time":"2025-05-08T13:03:22.73","Action":"output","Package":"pkg","Test":"TestB","Output":"main_test.go:42: expected true, got false\n"}`
	r := strings.NewReader(input)
	events, err := ParseTestEvents(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	failEvent := events[0]
	outputEvent := events[1]
	if failEvent.Action != "fail" || outputEvent.Action != "output" {
		t.Fatalf("unexpected actions: %s, %s", failEvent.Action, outputEvent.Action)
	}
	// if outputEvent.Output == "" || outputEvent.Output[:11] != "main_test.go" {
	// 	t.Errorf("expected file info in output, got %q", outputEvent.Output)
	// }
	if outputEvent.Output == "" || !strings.Contains(outputEvent.Output, "main_test.go:42:") {
		t.Errorf("expected file info in output, got %q", outputEvent.Output)
	}
}
