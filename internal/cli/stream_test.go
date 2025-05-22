package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"
)

func TestStreamParser(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectResults int
		expectError   bool
	}{
		{
			name: "parses complete test output",
			input: `{"Action":"run","Test":"TestExample"}
{"Action":"output","Test":"TestExample","Output":"=== RUN   TestExample\n"}
{"Action":"output","Test":"TestExample","Output":"--- PASS: TestExample (0.00s)\n"}
{"Action":"pass","Test":"TestExample","Elapsed":0.1}
{"Action":"run","Test":"TestExample2"}
{"Action":"output","Test":"TestExample2","Output":"=== RUN   TestExample2\n"}
{"Action":"output","Test":"TestExample2","Output":"--- FAIL: TestExample2 (0.00s)\n"}
{"Action":"output","Test":"TestExample2","Output":"    example_test.go:42: Expected 5, got 10\n"}
{"Action":"fail","Test":"TestExample2","Elapsed":0.1}`,
			expectResults: 2,
			expectError:   false,
		},
		{
			name: "handles incomplete JSON",
			input: `{"Action":"run","Test":"TestExample"}
{"Action":"output","Test":"TestExample","Output":"=== RUN   TestExample\n"}
{"Action":"output"`,
			expectResults: 0,
			expectError:   true,
		},
		{
			name:          "handles empty input",
			input:         "",
			expectResults: 0,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create input reader
			reader := strings.NewReader(tt.input)

			// Create the stream parser
			parser := NewStreamParser()

			// Create a channel to receive results
			resultCh := make(chan *TestResult, 10)
			errCh := make(chan error, 1)

			// Start parsing
			go func() {
				err := parser.Parse(reader, resultCh)
				if err != nil && err != io.EOF {
					errCh <- err
				}
				close(resultCh)
			}()

			// Collect results
			var results []*TestResult
			for result := range resultCh {
				results = append(results, result)
			}

			// Check for errors
			var err error
			select {
			case err = <-errCh:
			default:
			}

			// Verify expectations
			if tt.expectError && err == nil {
				t.Errorf("expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if len(results) != tt.expectResults {
				t.Errorf("expected %d results, got %d", tt.expectResults, len(results))
			}
		})
	}
}

func TestTestProcessor(t *testing.T) {
	tests := []struct {
		name           string
		events         []TestEvent
		expectSuites   int
		expectPassed   int
		expectFailed   int
		expectProgress bool
	}{
		{
			name: "processes complete test output",
			events: []TestEvent{
				{Action: "run", Test: "TestExample", Package: "github.com/test/example"},
				{Action: "output", Test: "TestExample", Package: "github.com/test/example", Output: "=== RUN   TestExample\n"},
				{Action: "output", Test: "TestExample", Package: "github.com/test/example", Output: "--- PASS: TestExample (0.10s)\n"},
				{Action: "pass", Test: "TestExample", Package: "github.com/test/example", Elapsed: 0.1},
				{Action: "run", Test: "TestExample2", Package: "github.com/test/example"},
				{Action: "output", Test: "TestExample2", Package: "github.com/test/example", Output: "=== RUN   TestExample2\n"},
				{Action: "output", Test: "TestExample2", Package: "github.com/test/example", Output: "--- FAIL: TestExample2 (0.05s)\n"},
				{Action: "output", Test: "TestExample2", Package: "github.com/test/example", Output: "    example_test.go:42: Expected 5, got 10\n"},
				{Action: "fail", Test: "TestExample2", Package: "github.com/test/example", Elapsed: 0.05},
			},
			expectSuites:   1,
			expectPassed:   1,
			expectFailed:   1,
			expectProgress: true,
		},
		{
			name:           "handles empty events",
			events:         []TestEvent{},
			expectSuites:   0,
			expectPassed:   0,
			expectFailed:   0,
			expectProgress: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test input
			var buf bytes.Buffer
			for _, event := range tt.events {
				data, _ := json.Marshal(event)
				buf.Write(data)
				buf.WriteString("\n")
			}

			reader := bytes.NewReader(buf.Bytes())

			// Create the processor
			var output bytes.Buffer
			formatter := NewColorFormatter(false)
			icons := NewIconProvider(false)

			processor := NewTestProcessor(&output, formatter, icons, 80)

			// Create a channel to receive progress updates
			progressCh := make(chan TestProgress, 10)

			// Start processing
			err := processor.ProcessStream(reader, progressCh)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Check if we received progress updates
			gotProgress := false
			select {
			case <-progressCh:
				gotProgress = true
			default:
			}

			if tt.expectProgress && !gotProgress {
				t.Errorf("expected progress updates, got none")
			}

			// Check the processor's state
			stats := processor.GetStats()

			if stats.TotalFiles != tt.expectSuites {
				t.Errorf("expected %d suites, got %d", tt.expectSuites, stats.TotalFiles)
			}

			if stats.PassedTests != tt.expectPassed {
				t.Errorf("expected %d passed tests, got %d", tt.expectPassed, stats.PassedTests)
			}

			if stats.FailedTests != tt.expectFailed {
				t.Errorf("expected %d failed tests, got %d", tt.expectFailed, stats.FailedTests)
			}
		})
	}
}
