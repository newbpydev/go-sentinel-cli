package cli

import (
	"bytes"
	"strconv"
	"testing"
)

func TestFailedTestSectionHeader(t *testing.T) {
	tests := []struct {
		name       string
		failCount  int
		wantHeader bool
		wantCount  bool
	}{
		{
			name:       "shows header with count when tests failed",
			failCount:  8,
			wantHeader: true,
			wantCount:  true,
		},
		{
			name:       "shows header without tests when no failures",
			failCount:  0,
			wantHeader: false,
			wantCount:  false,
		},
		{
			name:       "shows header with single failure",
			failCount:  1,
			wantHeader: true,
			wantCount:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			var buf bytes.Buffer
			formatter := NewColorFormatter(false) // No colors for testing
			icons := NewIconProvider(false)       // No unicode icons for testing
			renderer := NewFailedTestRenderer(&buf, formatter, icons, 80)

			// Execute
			err := renderer.RenderFailedTestsHeader(tt.failCount)

			// Assert
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := buf.String()

			// Check if header exists or not as expected
			if tt.wantHeader && len(output) == 0 {
				t.Errorf("expected header, got empty output")
			}

			if !tt.wantHeader && len(output) > 0 {
				t.Errorf("expected no header, got: %s", output)
			}

			// Check if count is displayed
			if tt.wantCount {
				expectedCountStr := "Failed Tests " + itoa(tt.failCount)
				if !contains(output, expectedCountStr) {
					t.Errorf("expected count %d in header, got: %s", tt.failCount, output)
				}
			}
		})
	}
}

func TestRenderFailedTest(t *testing.T) {
	tests := []struct {
		name           string
		testResult     *TestResult
		wantFailBadge  bool
		wantErrorType  bool
		wantErrorMsg   bool
		wantSourceCode bool
	}{
		{
			name: "shows detailed error for failed test",
			testResult: &TestResult{
				Name:   "WebSocketClient - connect method - should create a WebSocket with the given URL",
				Status: StatusFailed,
				Error: &TestError{
					Type:    "TypeError",
					Message: "wsClient.connect is not a function",
					Location: &SourceLocation{
						File: "test/websocket.test.ts",
						Line: 61,
					},
					SourceContext: []string{
						"// When",
						"// When",
						"wsClient.connect(testUrl);",
						"",
						"// Then",
					},
					HighlightedLine: 2, // 0-based index into SourceContext
				},
			},
			wantFailBadge:  true,
			wantErrorType:  true,
			wantErrorMsg:   true,
			wantSourceCode: true,
		},
		{
			name: "handles missing source context",
			testResult: &TestResult{
				Name:   "WebSocketClient - connect method - should create a WebSocket with the given URL",
				Status: StatusFailed,
				Error: &TestError{
					Type:    "TypeError",
					Message: "wsClient.connect is not a function",
					Location: &SourceLocation{
						File: "test/websocket.test.ts",
						Line: 61,
					},
				},
			},
			wantFailBadge:  true,
			wantErrorType:  true,
			wantErrorMsg:   true,
			wantSourceCode: false,
		},
		{
			name: "skips non-failed tests",
			testResult: &TestResult{
				Name:   "WebSocketClient - some test",
				Status: StatusPassed,
			},
			wantFailBadge:  false,
			wantErrorType:  false,
			wantErrorMsg:   false,
			wantSourceCode: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			var buf bytes.Buffer
			formatter := NewColorFormatter(false) // No colors for testing
			icons := NewIconProvider(false)       // No unicode icons for testing
			renderer := NewFailedTestRenderer(&buf, formatter, icons, 80)

			// Execute
			err := renderer.RenderFailedTest(tt.testResult)

			// Assert
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := buf.String()

			// For non-failed tests, we expect no output
			if tt.testResult.Status != StatusFailed && len(output) > 0 {
				t.Errorf("expected no output for non-failed test, got: %s", output)
				return
			}

			// For failed tests, check expected components
			if tt.testResult.Status == StatusFailed {
				if tt.wantFailBadge && !contains(output, "FAIL") {
					t.Errorf("expected FAIL badge, not found in: %s", output)
				}

				if tt.wantErrorType && !contains(output, tt.testResult.Error.Type) {
					t.Errorf("expected error type %s, not found in: %s", tt.testResult.Error.Type, output)
				}

				if tt.wantErrorMsg && !contains(output, tt.testResult.Error.Message) {
					t.Errorf("expected error message %s, not found in: %s", tt.testResult.Error.Message, output)
				}

				if tt.wantSourceCode {
					if tt.testResult.Error.SourceContext != nil {
						// Check if highlighted line is present
						highlightedLine := tt.testResult.Error.SourceContext[tt.testResult.Error.HighlightedLine]
						if !contains(output, highlightedLine) {
							t.Errorf("expected highlighted line %s, not found in: %s", highlightedLine, output)
						}
					}
				}
			}
		})
	}
}

func TestRenderFailedTests(t *testing.T) {
	tests := []struct {
		name        string
		failedTests []*TestResult
		wantHeader  bool
		wantCount   int
	}{
		{
			name: "renders multiple failed tests",
			failedTests: []*TestResult{
				{
					Name:   "WebSocketClient - connect method - should create a WebSocket with the given URL",
					Status: StatusFailed,
					Error: &TestError{
						Type:    "TypeError",
						Message: "wsClient.connect is not a function",
						Location: &SourceLocation{
							File: "test/websocket.test.ts",
							Line: 61,
						},
					},
				},
				{
					Name:   "WebSocketClient - event handlers - should register open event handlers",
					Status: StatusFailed,
					Error: &TestError{
						Type:    "TypeError",
						Message: "wsClient.connect is not a function",
						Location: &SourceLocation{
							File: "test/websocket.test.ts",
							Line: 72,
						},
					},
				},
			},
			wantHeader: true,
			wantCount:  2,
		},
		{
			name:        "handles empty failed tests list",
			failedTests: []*TestResult{},
			wantHeader:  false,
			wantCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			var buf bytes.Buffer
			formatter := NewColorFormatter(false) // No colors for testing
			icons := NewIconProvider(false)       // No unicode icons for testing
			renderer := NewFailedTestRenderer(&buf, formatter, icons, 80)

			// Execute
			err := renderer.RenderFailedTests(tt.failedTests)

			// Assert
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := buf.String()

			// Check if header exists as expected
			if tt.wantHeader && !contains(output, "Failed Tests") {
				t.Errorf("expected Failed Tests header, not found in: %s", output)
			}

			if !tt.wantHeader && contains(output, "Failed Tests") {
				t.Errorf("expected no Failed Tests header, but found in: %s", output)
			}

			// Check if all tests are rendered
			for _, test := range tt.failedTests {
				if !contains(output, test.Name) {
					t.Errorf("expected test name %s, not found in: %s", test.Name, output)
				}

				if test.Error != nil && !contains(output, test.Error.Message) {
					t.Errorf("expected error message %s, not found in: %s", test.Error.Message, output)
				}
			}
		})
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}

// itoa converts int to string
func itoa(i int) string {
	return strconv.Itoa(i)
}
