package processor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/newbpydev/go-sentinel/pkg/models"
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
			resultCh := make(chan *models.LegacyTestResult, 10)
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
			var results []*models.LegacyTestResult
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

func TestTestProcessor_ProcessStream(t *testing.T) {
	tests := []struct {
		name           string
		events         []models.TestEvent
		expectProgress bool
		expectEvents   int
	}{
		{
			name: "processes complete test output",
			events: []models.TestEvent{
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
			expectProgress: true,
			expectEvents:   2, // Two completed tests (pass and fail)
		},
		{
			name:           "handles empty events",
			events:         []models.TestEvent{},
			expectProgress: false,
			expectEvents:   0,
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
			formatter := &MockColorFormatter{}
			icons := &MockIconProvider{}

			processor := NewTestProcessor(&output, formatter, icons, 80)

			// Create a channel to receive progress updates
			progressCh := make(chan models.TestProgress, 10)

			// Start processing
			err := processor.ProcessStream(reader, progressCh)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Count progress updates received
			progressCount := 0
			for {
				select {
				case <-progressCh:
					progressCount++
				default:
					goto done
				}
			}
		done:

			// Check if we received the expected number of progress updates
			if tt.expectProgress && progressCount == 0 {
				t.Errorf("expected progress updates, got none")
			}

			if !tt.expectProgress && progressCount > 0 {
				t.Errorf("expected no progress updates, got %d", progressCount)
			}

			if tt.expectEvents > 0 && progressCount != tt.expectEvents {
				t.Errorf("expected %d progress events, got %d", tt.expectEvents, progressCount)
			}

			// Note: ProcessStream doesn't update the processor's internal statistics
			// It only sends progress updates. The processor's statistics remain empty.
		})
	}
}

// TestNewStreamParser_Creation tests NewStreamParser factory function
func TestNewStreamParser_Creation(t *testing.T) {
	t.Parallel()

	parser := NewStreamParser()
	if parser == nil {
		t.Fatal("Expected stream parser to be created, got nil")
	}

	// Verify parser can be used with a complete test sequence
	testInput := `{"Time":"2023-10-01T12:00:00Z","Action":"run","Package":"test","Test":"TestExample"}
{"Time":"2023-10-01T12:00:01Z","Action":"pass","Package":"test","Test":"TestExample","Elapsed":0.1}`
	reader := strings.NewReader(testInput)
	results := make(chan *models.LegacyTestResult, 1)

	go func() {
		defer close(results)
		_ = parser.Parse(reader, results)
	}()

	// Should be able to process at least one result
	resultCount := 0
	for range results {
		resultCount++
	}

	if resultCount == 0 {
		t.Error("Stream parser should be able to process test results")
	}
}

// TestStreamParser_EdgeCases tests stream parser with edge cases
func TestStreamParser_EdgeCases(t *testing.T) {
	t.Parallel()

	parser := NewStreamParser()

	testCases := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "Empty input",
			input:       "",
			expectError: false,
		},
		{
			name:        "Malformed JSON",
			input:       `{"Time":"invalid"`,
			expectError: true,
		},
		{
			name:        "Valid single event",
			input:       `{"Time":"2023-10-01T12:00:00Z","Action":"run","Package":"test"}`,
			expectError: false,
		},
		{
			name:        "Multiple events",
			input:       `{"Time":"2023-10-01T12:00:00Z","Action":"run","Package":"test","Test":"TestExample"}` + "\n" + `{"Time":"2023-10-01T12:00:01Z","Action":"pass","Package":"test","Test":"TestExample","Elapsed":0.1}`,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reader := strings.NewReader(tc.input)
			results := make(chan *models.LegacyTestResult, 10)

			err := parser.Parse(reader, results)
			close(results)

			if tc.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Count results
			resultCount := 0
			for range results {
				resultCount++
			}

			if tc.name == "Multiple events" && resultCount == 0 {
				t.Error("Expected at least one result for multiple events")
			}
		})
	}
}

// TestStreamParser_ConcurrentAccess tests stream parser with concurrent access
func TestStreamParser_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	parser := NewStreamParser()

	// Test multiple concurrent parsing operations
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(routineID int) {
			defer wg.Done()

			testInput := fmt.Sprintf(`{"Time":"2023-10-01T12:00:00Z","Action":"run","Package":"test%d","Test":"TestExample"}
{"Time":"2023-10-01T12:00:01Z","Action":"pass","Package":"test%d","Test":"TestExample","Elapsed":0.1}`, routineID, routineID)
			reader := strings.NewReader(testInput)
			results := make(chan *models.LegacyTestResult, 1)

			err := parser.Parse(reader, results)
			close(results)

			if err != nil {
				t.Errorf("Routine %d: unexpected error: %v", routineID, err)
			}
		}(i)
	}

	wg.Wait()
}
