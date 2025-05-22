package cli

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestRenderer_RenderTestResult(t *testing.T) {
	var buf bytes.Buffer
	r := NewRenderer(&buf)
	r.style.useColors = false // Disable colors for predictable output
	r.style.useIcons = false  // Use ASCII icons for predictable output

	// Test passing test
	result := &TestResult{
		Name:     "TestPass",
		Status:   TestStatusPassed,
		Duration: 100 * time.Millisecond,
	}
	r.RenderTestResult(result)
	output := buf.String()
	buf.Reset()

	if !strings.Contains(output, "TestPass") {
		t.Errorf("Output should contain test name: %s", output)
	}
	if !strings.Contains(output, "100ms") {
		t.Errorf("Output should contain duration: %s", output)
	}

	// Test failing test with error
	result = &TestResult{
		Name:     "TestFail",
		Status:   TestStatusFailed,
		Duration: 200 * time.Millisecond,
		Error: &TestError{
			Message: "assertion failed",
			Location: &SourceLocation{
				File:      "test_file.go",
				Line:      42,
				Snippet:   "assert.Equal(t, want, got)",
				StartLine: 40,
			},
			Expected: "5",
			Actual:   "3",
		},
	}
	r.RenderTestResult(result)
	output = buf.String()
	buf.Reset()

	expectedParts := []string{
		"TestFail",
		"200ms",
		"assertion failed",
		"test_file.go:42",
		"assert.Equal",
		"Expected: 5",
		"Actual: 3",
	}
	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("Output should contain %q: %s", part, output)
		}
	}
}

func TestRenderer_RenderFinalSummary(t *testing.T) {
	var buf bytes.Buffer
	r := NewRenderer(&buf)

	run := &TestRun{
		StartTime:  time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		Duration:   3 * time.Second,
		NumTotal:   10,
		NumPassed:  7,
		NumFailed:  2,
		NumSkipped: 1,
		Suites: []*TestSuite{
			{
				FilePath: "pkg/foo/foo_test.go",
				Tests: []*TestResult{
					{
						Name:   "TestFailed1",
						Status: TestStatusFailed,
						Error:  &TestError{Message: "expected true, got false"},
					},
					{
						Name:   "TestFailed2",
						Status: TestStatusFailed,
						Error:  &TestError{Message: "expected 42, got 41"},
					},
				},
				NumFailed: 2,
			},
		},
	}

	r.RenderFinalSummary(run)
	output := buf.String()

	expectedStrings := []string{
		"Test Files",
		"1 failed",
		"Tests     ",
		"2 failed",
		"7 passed",
		"1 skipped",
		"(10)",
		"Start at  ",
		"Duration  ",
		"FAILED Tests",
		"TestFailed1",
		"TestFailed2",
		"pkg/foo/foo_test.go",
	}

	for _, s := range expectedStrings {
		if !strings.Contains(output, s) {
			t.Errorf("Output should contain %q:\n%s", s, output)
		}
	}
}

func TestRenderer_RenderProgress(t *testing.T) {
	var buf bytes.Buffer
	r := NewRenderer(&buf)

	run := &TestRun{
		NumTotal:   4,
		NumPassed:  1,
		NumFailed:  1,
		NumSkipped: 0,
	}

	r.RenderProgress(run)
	output := buf.String()
	buf.Reset()

	expectedParts := []string{
		"Running tests",
		"50%",
		"2/4",
	}

	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("Output should contain %q: %s", part, output)
		}
	}
}

func TestRenderer_RenderWatchHeader(t *testing.T) {
	var buf bytes.Buffer
	r := NewRenderer(&buf)

	r.RenderWatchHeader()
	output := buf.String()

	expectedParts := []string{
		"WATCH MODE",
		"Press 'a' to run all tests",
		"Press 'f' to run only failed tests",
		"Press 'q' to quit",
	}

	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("Output should contain %q: %s", part, output)
		}
	}
}

func TestRenderer_RenderFileChange(t *testing.T) {
	var buf bytes.Buffer
	r := NewRenderer(&buf)

	path := "pkg/foo/foo_test.go"
	r.RenderFileChange(path)
	output := buf.String()

	if !strings.Contains(output, "File changed: "+path) {
		t.Errorf("Output should contain file change notification: %s", output)
	}
}

func TestRenderer_RenderSuiteSummary(t *testing.T) {
	var buf bytes.Buffer
	r := NewRenderer(&buf)
	r.style.useColors = false // Disable colors for predictable output

	// Test suite with failures
	suite := &TestSuite{
		FilePath:   "pkg/foo/foo_test.go",
		NumTotal:   3,
		NumPassed:  1,
		NumFailed:  1,
		NumSkipped: 1,
		Duration:   2 * time.Second,
	}

	r.RenderSuiteSummary(suite)
	output := buf.String()
	buf.Reset()

	expectedParts := []string{
		"Suite",
		"pkg/foo/foo_test.go",
		"Total: 3",
		"Passed: 1",
		"Failed: 1",
		"Skipped: 1",
		"Time: 2.00s",
	}

	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("Output should contain %q: %s", part, output)
		}
	}

	// Test suite without failures (should not output anything)
	suite = &TestSuite{
		FilePath:  "pkg/bar/bar_test.go",
		NumTotal:  2,
		NumPassed: 2,
	}

	r.RenderSuiteSummary(suite)
	output = buf.String()

	if output != "" {
		t.Errorf("Output should be empty for suite with no failures: %q", output)
	}
}
