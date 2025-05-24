package processor

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/newbpydev/go-sentinel/pkg/models"
)

// StreamParser parses Go test JSON output as it arrives
type StreamParser struct {
	testResults map[string]*models.LegacyTestResult
}

// NewStreamParser creates a new StreamParser
func NewStreamParser() *StreamParser {
	return &StreamParser{
		testResults: make(map[string]*models.LegacyTestResult),
	}
}

// Parse reads from the input reader and parses test events, sending TestResult objects to the results channel
func (p *StreamParser) Parse(r io.Reader, results chan<- *models.LegacyTestResult) error {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var event models.TestEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return fmt.Errorf("error parsing JSON: %w", err)
		}

		// Process the event
		p.processEvent(&event, results)
	}

	return scanner.Err()
}

// processEvent processes a single test event
func (p *StreamParser) processEvent(event *models.TestEvent, results chan<- *models.LegacyTestResult) {
	// Skip events without a test name
	if event.Test == "" {
		return
	}

	// Get or create the test result
	key := event.Package + "/" + event.Test
	result, ok := p.testResults[key]
	if !ok {
		result = &models.LegacyTestResult{
			Name:    event.Test,
			Package: event.Package,
			Status:  models.StatusRunning,
		}
		p.testResults[key] = result
	}

	// Update the result based on the event
	switch event.Action {
	case "run":
		result.Status = models.StatusRunning

	case "pass":
		result.Status = models.StatusPassed
		result.Duration = time.Duration(event.Elapsed * float64(time.Second))
		// Send the completed result
		results <- result

	case "fail":
		result.Status = models.StatusFailed
		result.Duration = time.Duration(event.Elapsed * float64(time.Second))
		// Send the completed result
		results <- result

	case "skip":
		result.Status = models.StatusSkipped
		result.Duration = time.Duration(event.Elapsed * float64(time.Second))
		// Send the completed result
		results <- result

	case "output":
		// Process output to extract additional information
		p.processOutput(result, event.Output)
	}
}

// processOutput extracts information from test output
func (p *StreamParser) processOutput(result *models.LegacyTestResult, output string) {
	// Accumulate all output
	result.Output += output

	// Look for test failure information
	if strings.Contains(output, "--- FAIL:") {
		result.Status = models.StatusFailed

		// Try to extract error details
		if result.Error == nil {
			result.Error = &models.LegacyTestError{
				Type:    "TestFailure",
				Message: strings.TrimSpace(strings.TrimPrefix(output, "--- FAIL:")),
			}
		}
	} else if strings.Contains(output, "--- PASS:") {
		result.Status = models.StatusPassed
	} else if strings.Contains(output, "--- SKIP:") {
		result.Status = models.StatusSkipped
	}

	// Extract error location information
	if strings.Contains(output, ".go:") {
		// This might be a location reference
		parts := strings.Split(output, ":")
		if len(parts) >= 3 {
			file := strings.TrimSpace(parts[0])

			// Initialize error if needed
			if result.Error == nil {
				result.Error = &models.LegacyTestError{
					Type:    "TestFailure",
					Message: strings.TrimSpace(output),
				}
			}

			// Set location
			if result.Error.Location == nil {
				result.Error.Location = &models.SourceLocation{
					File: file,
				}
			}
		}
	}
}
