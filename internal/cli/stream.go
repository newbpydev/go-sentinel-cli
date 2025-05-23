package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// StreamParser parses Go test JSON output as it arrives
type StreamParser struct {
	testResults map[string]*TestResult
}

// NewStreamParser creates a new StreamParser
func NewStreamParser() *StreamParser {
	return &StreamParser{
		testResults: make(map[string]*TestResult),
	}
}

// Parse reads from the input reader and parses test events, sending TestResult objects to the results channel
func (p *StreamParser) Parse(r io.Reader, results chan<- *TestResult) error {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var event TestEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return fmt.Errorf("error parsing JSON: %w", err)
		}

		// Process the event
		p.processEvent(&event, results)
	}

	return scanner.Err()
}

// processEvent processes a single test event
func (p *StreamParser) processEvent(event *TestEvent, results chan<- *TestResult) {
	// Skip events without a test name
	if event.Test == "" {
		return
	}

	// Get or create the test result
	key := event.Package + "/" + event.Test
	result, ok := p.testResults[key]
	if !ok {
		result = &TestResult{
			Name:    event.Test,
			Package: event.Package,
			Status:  StatusRunning,
		}
		p.testResults[key] = result
	}

	// Update the result based on the event
	switch event.Action {
	case "run":
		result.Status = StatusRunning

	case "pass":
		result.Status = StatusPassed
		result.Duration = time.Duration(event.Elapsed * float64(time.Second))
		// Send the completed result
		results <- result

	case "fail":
		result.Status = StatusFailed
		result.Duration = time.Duration(event.Elapsed * float64(time.Second))
		// Send the completed result
		results <- result

	case "skip":
		result.Status = StatusSkipped
		result.Duration = time.Duration(event.Elapsed * float64(time.Second))
		// Send the completed result
		results <- result

	case "output":
		// Process output to extract additional information
		p.processOutput(result, event.Output)
	}
}

// processOutput extracts information from test output
func (p *StreamParser) processOutput(result *TestResult, output string) {
	// Accumulate all output
	result.Output += output

	// Look for test failure information
	if strings.Contains(output, "--- FAIL:") {
		result.Status = StatusFailed

		// Try to extract error details
		if result.Error == nil {
			result.Error = &TestError{
				Type:    "TestFailure",
				Message: strings.TrimSpace(strings.TrimPrefix(output, "--- FAIL:")),
			}
		}
	} else if strings.Contains(output, "--- PASS:") {
		result.Status = StatusPassed
	} else if strings.Contains(output, "--- SKIP:") {
		result.Status = StatusSkipped
	}

	// Extract error location information
	if strings.Contains(output, ".go:") {
		// This might be a location reference
		parts := strings.Split(output, ":")
		if len(parts) >= 3 {
			file := strings.TrimSpace(parts[0])

			// Initialize error if needed
			if result.Error == nil {
				result.Error = &TestError{
					Type:    "TestFailure",
					Message: strings.TrimSpace(output),
				}
			}

			// Set location
			if result.Error.Location == nil {
				result.Error.Location = &SourceLocation{
					File: file,
				}
			}
		}
	}
}

// ProcessStream processes a stream of test output
func (p *TestProcessor) ProcessStream(r io.Reader, progress chan<- TestProgress) error {
	p.startTime = time.Now()

	// Create a stream parser
	parser := NewStreamParser()

	// Create a channel for test results
	resultCh := make(chan *TestResult, 10)

	// Start a goroutine to parse the stream
	go func() {
		if err := parser.Parse(r, resultCh); err != nil && err != io.EOF {
			_, _ = fmt.Fprintf(p.writer, "Error parsing test output: %v\n", err)
		}
		close(resultCh)
	}()

	// Process test results as they arrive with real-time updates
	totalTests := 0
	completedTests := 0
	suiteHeaders := make(map[string]bool) // Track which suite headers we've shown

	// Clear initial area for live updates
	fmt.Fprint(p.writer, "\033[s") // Save cursor position

	for result := range resultCh {
		totalTests++

		// Determine proper suite name (use package path instead of test file name)
		suiteName := p.formatSuiteName(result.Package)
		suitePath := filepath.Base(result.Package) + "_test.go"

		suite, ok := p.suites[suitePath]
		if !ok {
			suite = &TestSuite{
				FilePath: suitePath,
			}
			p.suites[suitePath] = suite
		}

		// Show suite header if this is the first test from this suite
		if !suiteHeaders[suiteName] {
			// Add spacing before new suite (except for the very first one)
			if len(suiteHeaders) > 0 {
				fmt.Fprintln(p.writer)
			}
			p.renderSuiteHeader(suiteName)
			suiteHeaders[suiteName] = true
		}

		// Enhance failed test errors with our advanced processing
		if result.Status == StatusFailed && result.Error != nil {
			// Create a mock event for our enhanced error processing
			event := TestEvent{
				Test:    result.Name,
				Package: result.Package,
				Action:  "fail",
			}
			// Replace simple error with enhanced error
			result.Error = p.createTestError(result, event)
		}

		// Add the test to the suite
		suite.Tests = append(suite.Tests, result)

		// Update suite statistics
		suite.TestCount++
		switch result.Status {
		case StatusPassed:
			suite.PassedCount++
			p.statistics.PassedTests++
		case StatusFailed:
			suite.FailedCount++
			p.statistics.FailedTests++
		case StatusSkipped:
			suite.SkippedCount++
			p.statistics.SkippedTests++
		}

		// If the test is completed, increment counter and show live update
		if result.Status != StatusRunning {
			completedTests++

			// Show live progress update
			p.renderLiveUpdate(suite, result)
		}

		// Send progress update
		if progress != nil {
			progress <- TestProgress{
				CompletedTests: completedTests,
				TotalTests:     totalTests,
				CurrentFile:    suitePath,
				Status:         result.Status,
			}
		}
	}

	// Update summary statistics
	p.statistics.TotalTests = totalTests
	p.statistics.TotalFiles = len(p.suites)
	p.statistics.EndTime = time.Now()
	p.statistics.Duration = p.statistics.EndTime.Sub(p.statistics.StartTime)

	// Track different phases (similar to Vitest)
	totalDuration := p.statistics.Duration

	// Estimate phase durations based on typical test execution patterns
	setupTime := totalDuration / 20       // ~5% for setup
	collectTime := totalDuration / 10     // ~10% for test discovery/collection
	testsTime := totalDuration * 7 / 10   // ~70% for actual test execution
	teardownTime := totalDuration / 20    // ~5% for teardown
	environmentTime := totalDuration / 10 // ~10% for environment setup

	// Store phases
	if p.statistics.Phases == nil {
		p.statistics.Phases = make(map[string]time.Duration)
	}
	p.statistics.Phases["setup"] = setupTime
	p.statistics.Phases["collect"] = collectTime
	p.statistics.Phases["tests"] = testsTime
	p.statistics.Phases["teardown"] = teardownTime
	p.statistics.Phases["environment"] = environmentTime

	// Count passed and failed files
	for _, suite := range p.suites {
		if suite.FailedCount > 0 {
			p.statistics.FailedFiles++
		} else {
			p.statistics.PassedFiles++
		}
	}

	// Add visual separation after live output
	if completedTests > 0 {
		fmt.Fprintln(p.writer)
	}

	return nil
}

// formatSuiteName converts package path to a user-friendly suite name
func (p *TestProcessor) formatSuiteName(packagePath string) string {
	// If it's command-line-arguments, try to infer from current working directory
	if packagePath == "command-line-arguments" {
		// Check if we're in a specific test directory
		if wd, err := os.Getwd(); err == nil {
			if strings.Contains(wd, "stress_tests") {
				return "stress_tests"
			}
			// Try to get the last directory component
			if parts := strings.Split(filepath.Clean(wd), string(filepath.Separator)); len(parts) > 0 {
				lastPart := parts[len(parts)-1]
				if lastPart != "." && lastPart != "" {
					return lastPart
				}
			}
		}
		return "."
	}

	// For named packages, use the full package path
	return packagePath
}

// renderSuiteHeader shows the package/suite name when streaming starts
func (p *TestProcessor) renderSuiteHeader(suiteName string) {
	// Display suite name without extra indentation
	fmt.Fprintln(p.writer, suiteName)
}

// renderLiveUpdate shows real-time test results as they complete
func (p *TestProcessor) renderLiveUpdate(suite *TestSuite, latestResult *TestResult) {
	// Format test result with appropriate icon
	var icon string
	var color func(string) string

	switch latestResult.Status {
	case StatusPassed:
		icon = "✓"
		color = func(s string) string { return fmt.Sprintf("\033[32m%s\033[0m", s) } // Green
	case StatusFailed:
		icon = "✗"
		color = func(s string) string { return fmt.Sprintf("\033[31m%s\033[0m", s) } // Red
	case StatusSkipped:
		icon = "⃠"
		color = func(s string) string { return fmt.Sprintf("\033[33m%s\033[0m", s) } // Yellow
	default:
		icon = "○"
		color = func(s string) string { return s }
	}

	// Show live test result with consistent indentation under suite
	duration := fmt.Sprintf("%dms", latestResult.Duration.Milliseconds())
	testLine := fmt.Sprintf("  %s %s %s", color(icon), latestResult.Name, color(duration))

	fmt.Fprintln(p.writer, testLine)
}
