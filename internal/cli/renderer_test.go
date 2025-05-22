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
	r.style.useIcons = false  // Use Windows icons for predictable output
	r.style.isWindows = true  // Force Windows mode

	tests := []struct {
		name           string
		result         *TestResult
		expectedParts  []string
		notExpectParts []string
	}{
		{
			name: "simple passing test",
			result: &TestResult{
				Name:     "TestWebSocketClient_Connect",
				Status:   TestStatusPassed,
				Duration: 100 * time.Millisecond,
			},
			expectedParts: []string{
				"+ Web socket client connect  100ms",
			},
		},
		{
			name: "nested test with parent",
			result: &TestResult{
				Name:     "TestWebSocketClient/Connect_WithURL",
				Status:   TestStatusPassed,
				Duration: 150 * time.Millisecond,
			},
			expectedParts: []string{
				"Web socket client › Connect with URL  150ms",
			},
		},
		{
			name: "deeply nested test",
			result: &TestResult{
				Name:     "TestWebSocketClient/Connect/WithURL/AndHeaders",
				Status:   TestStatusPassed,
				Duration: 200 * time.Millisecond,
			},
			expectedParts: []string{
				"Web socket client › Connect › With URL › And headers  200ms",
			},
		},
		{
			name: "failing test with error",
			result: &TestResult{
				Name:     "TestWebSocketClient_SendMethod",
				Status:   TestStatusFailed,
				Duration: 200 * time.Millisecond,
				Error: &TestError{
					Message: "wsClient.connect is not a function",
					Location: &SourceLocation{
						File:      "test_file.go",
						Line:      42,
						Snippet:   "assert.Equal(t, want, got)",
						StartLine: 40,
					},
					Expected: "5",
					Actual:   "3",
				},
			},
			expectedParts: []string{
				"x Web socket client send method  200ms",
				"→ wsClient.connect is not a function",
				"at test_file.go:42",
				"40 │ assert.Equal(t, want, got)",
				"Expected",
				"5",
				"Actual",
				"3",
			},
		},
		{
			name: "test with numbers",
			result: &TestResult{
				Name:     "TestHTTP2_Request",
				Status:   TestStatusPassed,
				Duration: 100 * time.Millisecond,
			},
			expectedParts: []string{
				"+ HTTP2 request  100ms",
			},
		},
		{
			name: "test with common abbreviations",
			result: &TestResult{
				Name:     "TestParseJSON_WithURLAndSSL",
				Status:   TestStatusPassed,
				Duration: 100 * time.Millisecond,
			},
			expectedParts: []string{
				"+ Parse JSON with URL and SSL  100ms",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			r.RenderTestResult(tt.result)
			output := buf.String()

			for _, expected := range tt.expectedParts {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, but got:\n%s", expected, output)
				}
			}

			for _, notExpected := range tt.notExpectParts {
				if strings.Contains(output, notExpected) {
					t.Errorf("Expected output to NOT contain %q, but got:\n%s", notExpected, output)
				}
			}
		})
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
		"Time: 2.0s",
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
