package runner

import (
	"strings"
	"testing"
)

func TestFormatMillis_Conversion(t *testing.T) {
	tests := []struct {
		name     string
		seconds  float64
		expected string
	}{
		{"zero", 0.0, "0ms"},
		{"one_second", 1.0, "1000ms"},
		{"fraction", 0.0123, "12ms"},
		{"small_fraction", 0.001, "1ms"},
		{"large_number", 123.456, "123456ms"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatMillis(tc.seconds)
			if result != tc.expected {
				t.Errorf("FormatMillis(%v) = %v, want %v", tc.seconds, result, tc.expected)
			}
		})
	}
}

func TestFormatCoverage_Percentage(t *testing.T) {
	tests := []struct {
		name     string
		coverage float64
		expected string
	}{
		{"zero", 0.0, "0.00%"},
		{"hundred_percent", 1.0, "100.00%"},
		{"fifty_percent", 0.5, "50.00%"},
		{"decimal_percent", 0.7523, "75.23%"},
		{"already_percentage", 75.23, "75.23%"},
		{"over_hundred", 150.0, "150.00%"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatCoverage(tc.coverage)
			if result != tc.expected {
				t.Errorf("FormatCoverage(%v) = %v, want %v", tc.coverage, result, tc.expected)
			}
		})
	}
}

func TestFormatTestOutput_PassingTest(t *testing.T) {
	events := []TestEvent{
		{
			Time:    "2025-05-12T10:00:00Z",
			Action:  "run",
			Package: "pkg/foo",
			Test:    "TestPass",
		},
		{
			Time:    "2025-05-12T10:00:01Z",
			Action:  "pass",
			Package: "pkg/foo",
			Test:    "TestPass",
			Elapsed: 0.123,
		},
	}

	output := FormatTestOutput(events)

	// Debug output
	t.Logf("Actual output:\n%s", output)

	// Check for expected output format
	if !strings.Contains(output, "=== RUN   TestPass") {
		t.Error("missing run marker")
	}
	if !strings.Contains(output, "--- PASS: TestPass (0.123s)") {
		t.Error("missing pass marker with elapsed time")
	}
	if !strings.Contains(output, "ok   \tpkg/foo") && !strings.Contains(output, "ok   pkg/foo") {
		t.Error("missing package status")
	}
}

func TestFormatTestOutput_FailingTest(t *testing.T) {
	events := []TestEvent{
		{
			Time:    "2025-05-12T10:00:00Z",
			Action:  "run",
			Package: "pkg/foo",
			Test:    "TestFail",
		},
		{
			Time:    "2025-05-12T10:00:01Z",
			Action:  "fail",
			Package: "pkg/foo",
			Test:    "TestFail",
			Output:  "test failed: expected true, got false",
			Elapsed: 0.056,
		},
	}

	output := FormatTestOutput(events)

	// Check for expected output format
	if !strings.Contains(output, "=== RUN   TestFail") {
		t.Error("missing run marker")
	}
	if !strings.Contains(output, "--- FAIL: TestFail (0.056s)") {
		t.Error("missing fail marker with elapsed time")
	}
	if !strings.Contains(output, "test failed: expected true, got false") {
		t.Error("missing error message")
	}
	if !strings.Contains(output, "FAIL	pkg/foo") {
		t.Error("missing package status")
	}
}

func TestFormatTestOutput_SkippedTest(t *testing.T) {
	events := []TestEvent{
		{
			Time:    "2025-05-12T10:00:00Z",
			Action:  "run",
			Package: "pkg/foo",
			Test:    "TestSkip",
		},
		{
			Time:    "2025-05-12T10:00:01Z",
			Action:  "skip",
			Package: "pkg/foo",
			Test:    "TestSkip",
			Output:  "skipping test in short mode",
			Elapsed: 0.001,
		},
	}

	output := FormatTestOutput(events)

	// Check for expected output format
	if !strings.Contains(output, "=== RUN   TestSkip") {
		t.Error("missing run marker")
	}
	if !strings.Contains(output, "--- SKIP: TestSkip (0.001s)") {
		t.Error("missing skip marker with elapsed time")
	}
	if !strings.Contains(output, "skipping test in short mode") {
		t.Error("missing skip message")
	}
}

func TestFormatTestOutput_SubtestPassing(t *testing.T) {
	events := []TestEvent{
		{
			Time:    "2025-05-12T10:00:00Z",
			Action:  "run",
			Package: "pkg/foo",
			Test:    "TestParent",
		},
		{
			Time:    "2025-05-12T10:00:01Z",
			Action:  "run",
			Package: "pkg/foo",
			Test:    "TestParent/SubtestA",
		},
		{
			Time:    "2025-05-12T10:00:02Z",
			Action:  "pass",
			Package: "pkg/foo",
			Test:    "TestParent/SubtestA",
			Elapsed: 0.050,
		},
		{
			Time:    "2025-05-12T10:00:03Z",
			Action:  "run",
			Package: "pkg/foo",
			Test:    "TestParent/SubtestB",
		},
		{
			Time:    "2025-05-12T10:00:04Z",
			Action:  "pass",
			Package: "pkg/foo",
			Test:    "TestParent/SubtestB",
			Elapsed: 0.050,
		},
		{
			Time:    "2025-05-12T10:00:05Z",
			Action:  "pass",
			Package: "pkg/foo",
			Test:    "TestParent",
			Elapsed: 0.100,
		},
	}

	output := FormatTestOutput(events)

	// Check for expected output format
	expectedLines := []string{
		"=== RUN   TestParent",
		"=== RUN   TestParent/SubtestA",
		"=== RUN   TestParent/SubtestB",
		"--- PASS: TestParent (0.100s)",
		"    --- PASS: TestParent/SubtestA (0.050s)",
		"    --- PASS: TestParent/SubtestB (0.050s)",
	}

	for _, line := range expectedLines {
		if !strings.Contains(output, line) {
			t.Errorf("missing expected line: %q", line)
		}
	}
}

func TestFormatTestOutput_SubtestFailing(t *testing.T) {
	events := []TestEvent{
		{
			Time:    "2025-05-12T10:00:00Z",
			Action:  "run",
			Package: "pkg/foo",
			Test:    "TestParent",
		},
		{
			Time:    "2025-05-12T10:00:01Z",
			Action:  "run",
			Package: "pkg/foo",
			Test:    "TestParent/SubtestA",
		},
		{
			Time:    "2025-05-12T10:00:02Z",
			Action:  "pass",
			Package: "pkg/foo",
			Test:    "TestParent/SubtestA",
			Elapsed: 0.050,
		},
		{
			Time:    "2025-05-12T10:00:03Z",
			Action:  "run",
			Package: "pkg/foo",
			Test:    "TestParent/SubtestB",
		},
		{
			Time:    "2025-05-12T10:00:04Z",
			Action:  "fail",
			Package: "pkg/foo",
			Test:    "TestParent/SubtestB",
			Output:  "subtest B failed",
			Elapsed: 0.050,
		},
		{
			Time:    "2025-05-12T10:00:05Z",
			Action:  "fail",
			Package: "pkg/foo",
			Test:    "TestParent",
			Elapsed: 0.100,
		},
	}

	output := FormatTestOutput(events)

	// Check for expected output format
	expectedLines := []string{
		"=== RUN   TestParent",
		"=== RUN   TestParent/SubtestA",
		"=== RUN   TestParent/SubtestB",
		"--- FAIL: TestParent (0.100s)",
		"    --- PASS: TestParent/SubtestA (0.050s)",
		"    --- FAIL: TestParent/SubtestB (0.050s)",
		"        subtest B failed",
		"FAIL",
	}

	for _, line := range expectedLines {
		if !strings.Contains(output, line) {
			t.Errorf("missing expected line: %q", line)
		}
	}
}

func TestFormatTestOutput_EmptyEvents(t *testing.T) {
	output := FormatTestOutput(nil)
	if output != "" {
		t.Errorf("expected empty output for nil events, got %q", output)
	}

	output = FormatTestOutput([]TestEvent{})
	if output != "" {
		t.Errorf("expected empty output for empty events, got %q", output)
	}
}

func TestFormatTestOutput_Coverage(t *testing.T) {
	events := []TestEvent{
		{
			Time:    "2025-05-12T10:00:00Z",
			Action:  "run",
			Package: "pkg/foo",
			Test:    "TestPass",
		},
		{
			Time:    "2025-05-12T10:00:01Z",
			Action:  "pass",
			Package: "pkg/foo",
			Test:    "TestPass",
			Output:  "coverage: 85.2% of statements",
			Elapsed: 0.123,
		},
	}

	output := FormatTestOutput(events)

	// Check for coverage information
	if !strings.Contains(output, "coverage: 85.2% of statements") {
		t.Error("missing coverage information")
	}
}

func TestFormatTestOutput_MultiplePackages(t *testing.T) {
	events := []TestEvent{
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
			Elapsed: 0.100,
		},
		{
			Time:    "2025-05-12T10:00:02Z",
			Action:  "run",
			Package: "pkg/bar",
			Test:    "TestB",
		},
		{
			Time:    "2025-05-12T10:00:03Z",
			Action:  "pass",
			Package: "pkg/bar",
			Test:    "TestB",
			Elapsed: 0.050,
		},
	}

	output := FormatTestOutput(events)

	// Debug output
	t.Logf("Actual output:\n%s", output)

	// Check for package separation
	if !strings.Contains(output, "ok   \tpkg/foo") && !strings.Contains(output, "ok   pkg/foo") {
		t.Error("missing status for first package")
	}
	if !strings.Contains(output, "ok   \tpkg/bar") && !strings.Contains(output, "ok   pkg/bar") {
		t.Error("missing status for second package")
	}

	// Check order of events
	firstPkg := strings.Index(output, "pkg/foo")
	secondPkg := strings.Index(output, "pkg/bar")
	if firstPkg >= secondPkg {
		t.Error("packages not in chronological order")
	}
}
