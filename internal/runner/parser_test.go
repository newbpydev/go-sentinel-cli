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
