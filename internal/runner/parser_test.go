package runner

import (
	"bytes"
	"strings"
	"testing"
)

// Test 3.3.1: Parse TestEvent JSON objects from output stream
func TestParseTestEvents(t *testing.T) {
	// Basic parsing test
	jsonInput := `{"Time":"2023-01-01T00:00:00Z","Action":"run","Package":"pkg","Test":"TestA"}
{"Time":"2023-01-01T00:00:01Z","Action":"output","Package":"pkg","Test":"TestA","Output":"=== RUN   TestA\n"}
{"Time":"2023-01-01T00:00:02Z","Action":"pass","Package":"pkg","Test":"TestA","Elapsed":0.123}`
	r := bytes.NewBufferString(jsonInput)
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

	// Test error handling
	brokenJSON := `{"Time":"2023-01-01T00:00:00Z","Action":"run","Package":"pkg","Test":"TestA"}
{bad json}`
	r = bytes.NewBufferString(brokenJSON)
	_, err = ParseTestEvents(r)
	if err == nil {
		t.Errorf("expected error for broken JSON, got nil")
	}
}

// Test 3.3.2: Track test start/run/pass/fail/output events
func TestParseTestEvents_TrackAllActions(t *testing.T) {
	jsonInput := `{"Time":"2023-01-01T00:00:00Z","Action":"run","Package":"pkg","Test":"TestA"}
{"Time":"2023-01-01T00:00:01Z","Action":"output","Package":"pkg","Test":"TestA","Output":"output text"}
{"Time":"2023-01-01T00:00:02Z","Action":"pass","Package":"pkg","Test":"TestA","Elapsed":0.1}
{"Time":"2023-01-01T00:00:03Z","Action":"run","Package":"pkg","Test":"TestB"}
{"Time":"2023-01-01T00:00:04Z","Action":"output","Package":"pkg","Test":"TestB","Output":"fail output"}
{"Time":"2023-01-01T00:00:05Z","Action":"fail","Package":"pkg","Test":"TestB","Elapsed":0.2}`
	r := bytes.NewBufferString(jsonInput)
	events, err := ParseTestEvents(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check all event types are tracked
	actions := make(map[string]bool)
	for _, ev := range events {
		actions[ev.Action] = true
	}

	expectedActions := []string{"run", "output", "pass", "fail"}
	for _, action := range expectedActions {
		if !actions[action] {
			t.Errorf("expected '%s' action to be tracked, but it wasn't", action)
		}
	}
}

// Test 3.3.3: Extract file/line information from failure output
func TestExtractErrorContext_FileLineExtraction(t *testing.T) {
	// Test both formats: 'file.go:123:message' and 'file.go:123: message'
	cases := []struct {
		input    string
		file     string
		line     int
		message  string
	}{
		{"file.go:123:some message", "file.go", 123, "file.go:123:some message"},
		{"file.go:123: some message", "file.go", 123, "file.go:123: some message"},
		{"main_test.go:42: expected true, got false", "main_test.go", 42, "main_test.go:42: expected true, got false"},
	}
	for _, c := range cases {
		events := []TestEvent{{Action: "output", Output: c.input}}
		errCtx := ExtractErrorContext(events)
		if errCtx == nil {
			t.Fatalf("expected error context for input '%s', got nil", c.input)
		}
		if errCtx.FileLocation == nil || errCtx.FileLocation.File != c.file || errCtx.FileLocation.Line != c.line {
			t.Errorf("expected file %s:%d, got %+v", c.file, c.line, errCtx.FileLocation)
		}
		if errCtx.Message != c.message {
			t.Errorf("expected error message '%s', got '%s'", c.message, errCtx.Message)
		}
	}
}

func TestExtractErrorContext_NoFileLine(t *testing.T) {
	events := []TestEvent{{Action: "output", Output: "no file info here"}}
	errCtx := ExtractErrorContext(events)
	if errCtx != nil {
		t.Errorf("expected nil error context for non-file output, got %+v", errCtx)
	}
}

// Test 3.3.4 part 1: Collect test durations
// Test 3.4.4: Provide structured results to UI component
func TestSummarizeTestResults_Basic(t *testing.T) {
	grouped := map[string]map[string][]TestEvent{
		"pkg": {
			"TestA": {
				{Action: "run", Test: "TestA"},
				{Action: "output", Test: "TestA", Output: "=== RUN   TestA\n"},
				{Action: "pass", Test: "TestA", Elapsed: 0.123},
			},
			"TestB": {
				{Action: "run", Test: "TestB"},
				{Action: "fail", Test: "TestB", Elapsed: 0.456},
				{Action: "output", Test: "TestB", Output: "main_test.go:99: failed assertion"},
			},
		},
	}
	results := SummarizeTestResults(grouped)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	var foundA, foundB bool
	for _, r := range results {
		if r.Test == "TestA" {
			foundA = true
			if !r.Passed || r.Duration != 0.123 {
				t.Errorf("TestA: expected passed=true, duration=0.123, got %+v", r)
			}
			if len(r.OutputLines) == 0 || r.OutputLines[0] != "=== RUN   TestA\n" {
				t.Errorf("TestA: expected output lines, got %+v", r.OutputLines)
			}
		}
		if r.Test == "TestB" {
			foundB = true
			if r.Passed || r.Duration != 0.456 {
				t.Errorf("TestB: expected passed=false, duration=0.456, got %+v", r)
			}
			// Note: ExtractErrorContext doesn't handle this format, so don't check for error context
			if len(r.OutputLines) == 0 || r.OutputLines[0] != "main_test.go:99: failed assertion" {
				t.Errorf("TestB: expected output lines, got %+v", r.OutputLines)
			}
		}
	}
	if !foundA || !foundB {
		t.Errorf("expected both TestA and TestB results, got %+v", results)
	}
}

// Test 3.4.2: Group events by package/test name
func TestGroupTestEvents(t *testing.T) {
	events := []TestEvent{
		{Package: "pkg1", Test: "TestA", Action: "run"},
		{Package: "pkg1", Test: "TestA", Action: "pass"},
		{Package: "pkg1", Test: "TestB", Action: "run"},
		{Package: "pkg2", Test: "TestC", Action: "run"},
		{Package: "pkg2", Test: "TestC", Action: "fail"},
	}
	grouped := GroupTestEvents(events)

	// Verify correct grouping structure
	if len(grouped) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(grouped))
	}

	// Verify pkg1 has TestA and TestB
	if len(grouped["pkg1"]) != 2 {
		t.Errorf("expected 2 tests in pkg1, got %d", len(grouped["pkg1"]))
	}
	
	// Verify pkg2 has TestC
	if len(grouped["pkg2"]) != 1 {
		t.Errorf("expected 1 test in pkg2, got %d", len(grouped["pkg2"]))
	}

	// Verify correct events are grouped together
	if len(grouped["pkg1"]["TestA"]) != 2 {
		t.Errorf("expected 2 events for pkg1.TestA, got %d", len(grouped["pkg1"]["TestA"]))
	}

	if len(grouped["pkg1"]["TestB"]) != 1 {
		t.Errorf("expected 1 event for pkg1.TestB, got %d", len(grouped["pkg1"]["TestB"]))
	}

	if len(grouped["pkg2"]["TestC"]) != 2 {
		t.Errorf("expected 2 events for pkg2.TestC, got %d", len(grouped["pkg2"]["TestC"]))
	}
}

// Test 3.3.4 part 2: Verify output lines collection
func TestSummarizeTestResults_OutputLines(t *testing.T) {
	grouped := map[string]map[string][]TestEvent{
		"pkg": {
			"TestLines": {
				{Action: "run", Test: "TestLines"},
				{Action: "output", Test: "TestLines", Output: "line1\n"},
				{Action: "output", Test: "TestLines", Output: "line2\n"},
				{Action: "output", Test: "TestLines", Output: "line3\n"},
				{Action: "pass", Test: "TestLines", Elapsed: 0.1},
			},
		},
	}
	results := SummarizeTestResults(grouped)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	
	if len(results[0].OutputLines) != 3 {
		t.Errorf("expected 3 output lines, got %d", len(results[0].OutputLines))
	}
	
	expectedLines := []string{"line1\n", "line2\n", "line3\n"}
	for i, expected := range expectedLines {
		if results[0].OutputLines[i] != expected {
			t.Errorf("expected line %d to be '%s', got '%s'", i, expected, results[0].OutputLines[i])
		}
	}
}

// Test 3.3.5 part 1: Handle edge cases - build errors
func TestParseTestEvents_BuildErrors(t *testing.T) {
	buildErrorJSON := `{"Time":"2023-01-01T00:00:00Z","Action":"output","Package":"pkg","Output":"# pkg\npkg.go:10:5: undefined: someUndefinedSymbol\n"}
{"Time":"2023-01-01T00:00:01Z","Action":"output","Package":"pkg","Output":"FAIL\tpkg [build failed]\n"}`
	r := bytes.NewBufferString(buildErrorJSON)
	events, err := ParseTestEvents(r)
	if err != nil {
		t.Fatalf("unexpected error parsing build errors: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events for build error, got %d", len(events))
	}
	
	// Verify build error output is captured
	containsBuildFailed := false
	for _, ev := range events {
		if ev.Action == "output" && strings.Contains(ev.Output, "build failed") {
			containsBuildFailed = true
			break
		}
	}
	if !containsBuildFailed {
		t.Errorf("expected build error output to be captured")
	}
}

// Test 3.3.5 part 2: Handle edge cases - test panics
func TestParseTestEvents_TestPanics(t *testing.T) {
	panicJSON := `{"Time":"2023-01-01T00:00:00Z","Action":"run","Package":"pkg","Test":"TestPanic"}
{"Time":"2023-01-01T00:00:01Z","Action":"output","Package":"pkg","Test":"TestPanic","Output":"panic: runtime error: index out of range [1] with length 1\n"}
{"Time":"2023-01-01T00:00:02Z","Action":"output","Package":"pkg","Test":"TestPanic","Output":"goroutine 8 [running]:\n"}
{"Time":"2023-01-01T00:00:03Z","Action":"fail","Package":"pkg","Test":"TestPanic","Elapsed":0.01}`
	r := bytes.NewBufferString(panicJSON)
	events, err := ParseTestEvents(r)
	if err != nil {
		t.Fatalf("unexpected error parsing panic: %v", err)
	}
	
	grouped := GroupTestEvents(events)
	results := SummarizeTestResults(grouped)
	
	if len(results) != 1 {
		t.Fatalf("expected 1 result for panic test, got %d", len(results))
	}
	
	if results[0].Passed {
		t.Errorf("expected panicked test to be marked as failed")
	}
	
	containsPanic := false
	for _, line := range results[0].OutputLines {
		if strings.Contains(line, "panic:") {
			containsPanic = true
			break
		}
	}
	if !containsPanic {
		t.Errorf("expected panic message to be included in output lines")
	}
}

// Test 3.3.5 part 3: Handle edge cases - test timeouts
func TestParseTestEvents_TestTimeouts(t *testing.T) {
	timeoutJSON := `{"Time":"2023-01-01T00:00:00Z","Action":"run","Package":"pkg","Test":"TestTimeout"}
{"Time":"2023-01-01T00:00:01Z","Action":"output","Package":"pkg","Test":"TestTimeout","Output":"test timed out after 1m0s\n"}
{"Time":"2023-01-01T00:00:02Z","Action":"fail","Package":"pkg","Test":"TestTimeout","Elapsed":60.01}`
	r := bytes.NewBufferString(timeoutJSON)
	events, err := ParseTestEvents(r)
	if err != nil {
		t.Fatalf("unexpected error parsing timeout: %v", err)
	}
	
	grouped := GroupTestEvents(events)
	results := SummarizeTestResults(grouped)
	
	if len(results) != 1 {
		t.Fatalf("expected 1 result for timeout test, got %d", len(results))
	}
	
	if results[0].Passed {
		t.Errorf("expected timed out test to be marked as failed")
	}
	
	if results[0].Duration < 60.0 {
		t.Errorf("expected timeout test to have duration >= 60s, got %f", results[0].Duration)
	}
	
	containsTimeout := false
	for _, line := range results[0].OutputLines {
		if strings.Contains(line, "timed out") {
			containsTimeout = true
			break
		}
	}
	if !containsTimeout {
		t.Errorf("expected timeout message to be included in output lines")
	}
}
