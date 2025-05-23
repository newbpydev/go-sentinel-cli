package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// NewTestProcessor creates a new TestProcessor
func NewTestProcessor(writer io.Writer, formatter *ColorFormatter, icons *IconProvider, width int) *TestProcessor {
	return &TestProcessor{
		writer:     writer,
		formatter:  formatter,
		icons:      icons,
		width:      width,
		suites:     make(map[string]*TestSuite),
		statistics: &TestRunStats{},
		startTime:  time.Now(),
	}
}

// ProcessJSONOutput processes the JSON output from Go test and updates the processor state
func (p *TestProcessor) ProcessJSONOutput(output string) error {
	// Reset the processor state
	p.Reset()

	// Split the output into lines
	lines := strings.Split(output, "\n")

	// Process each line as a JSON object
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var event TestEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}

		// Process the event based on its action
		switch event.Action {
		case "run":
			// A test is about to run
			p.onTestRun(event)
		case "pass":
			// A test has passed
			p.onTestPass(event)
		case "fail":
			// A test has failed
			p.onTestFail(event)
		case "skip":
			// A test was skipped
			p.onTestSkip(event)
		case "output":
			// Test output - add to current test
			p.onTestOutput(event)
		}
	}

	// Finalize the state
	p.finalize()

	return nil
}

// Reset resets the processor state for a new test run
func (p *TestProcessor) Reset() {
	p.statistics = &TestRunStats{
		StartTime: time.Now(),
	}
	p.suites = make(map[string]*TestSuite)
}

// onTestRun handles a test run event
func (p *TestProcessor) onTestRun(event TestEvent) {
	// Create a new test result
	result := &TestResult{
		Name:     event.Test,
		Package:  event.Package,
		Status:   StatusRunning,
		Duration: 0,
	}

	// Find or create the test suite first
	suitePath := event.Package
	suite, ok := p.suites[suitePath]
	if !ok {
		suite = &TestSuite{
			FilePath: suitePath,
		}
		p.suites[suitePath] = suite
	}

	// Check if this is a subtest
	if strings.Contains(event.Test, "/") {
		// Extract parent test name
		parts := strings.SplitN(event.Test, "/", 2)
		result.Parent = parts[0]

		// Add to parent test if it exists
		for _, test := range suite.Tests {
			if test.Name == result.Parent {
				test.Subtests = append(test.Subtests, result)
				// FIXED: Count subtests in suite.TestCount too
				suite.TestCount++
				return
			}
		}
	}

	// Add the test to the suite
	suite.Tests = append(suite.Tests, result)
	suite.TestCount++
}

// onTestPass handles a test pass event
func (p *TestProcessor) onTestPass(event TestEvent) {
	// Find the test
	for _, suite := range p.suites {
		for _, test := range suite.Tests {
			if test.Name == event.Test {
				test.Status = StatusPassed
				test.Duration = time.Duration(event.Elapsed * float64(time.Second))
				suite.PassedCount++
				p.statistics.PassedTests++
				p.statistics.TotalTests++
				return
			}

			// Check subtests
			for _, subtest := range test.Subtests {
				if subtest.Name == event.Test {
					subtest.Status = StatusPassed
					subtest.Duration = time.Duration(event.Elapsed * float64(time.Second))
					suite.PassedCount++
					p.statistics.PassedTests++
					p.statistics.TotalTests++
					return
				}
			}
		}
	}
}

// onTestFail handles a test fail event
func (p *TestProcessor) onTestFail(event TestEvent) {
	// Find the test
	for _, suite := range p.suites {
		for _, test := range suite.Tests {
			if test.Name == event.Test {
				test.Status = StatusFailed
				test.Duration = time.Duration(event.Elapsed * float64(time.Second))

				// Create error details from accumulated output
				test.Error = p.createTestError(test, event)

				suite.FailedCount++
				p.statistics.FailedTests++
				p.statistics.TotalTests++
				return
			}

			// Check subtests
			for _, subtest := range test.Subtests {
				if subtest.Name == event.Test {
					subtest.Status = StatusFailed
					subtest.Duration = time.Duration(event.Elapsed * float64(time.Second))

					// Create error details from accumulated output
					subtest.Error = p.createTestError(subtest, event)

					suite.FailedCount++
					p.statistics.FailedTests++
					p.statistics.TotalTests++
					return
				}
			}
		}
	}
}

// onTestSkip handles a test skip event
func (p *TestProcessor) onTestSkip(event TestEvent) {
	// Find the test
	for _, suite := range p.suites {
		for _, test := range suite.Tests {
			if test.Name == event.Test {
				test.Status = StatusSkipped
				test.Duration = time.Duration(event.Elapsed * float64(time.Second))
				suite.SkippedCount++
				p.statistics.SkippedTests++
				p.statistics.TotalTests++
				return
			}

			// Check subtests
			for _, subtest := range test.Subtests {
				if subtest.Name == event.Test {
					subtest.Status = StatusSkipped
					subtest.Duration = time.Duration(event.Elapsed * float64(time.Second))
					suite.SkippedCount++
					p.statistics.SkippedTests++
					p.statistics.TotalTests++
					return
				}
			}
		}
	}
}

// onTestOutput handles test output
func (p *TestProcessor) onTestOutput(event TestEvent) {
	// Find the test
	for _, suite := range p.suites {
		for _, test := range suite.Tests {
			if test.Name == event.Test {
				test.Output += event.Output
				return
			}

			// Check subtests
			for _, subtest := range test.Subtests {
				if subtest.Name == event.Test {
					subtest.Output += event.Output
					return
				}
			}
		}
	}
}

// finalize updates final statistics
func (p *TestProcessor) finalize() {
	// Calculate suite statistics
	for _, suite := range p.suites {
		if suite.FailedCount > 0 {
			p.statistics.FailedFiles++
		} else if suite.TestCount > 0 {
			p.statistics.PassedFiles++
		}
		p.statistics.TotalFiles++

		// Calculate total duration (simple sum for now)
		for _, test := range suite.Tests {
			suite.Duration += test.Duration
		}
	}

	// Set end time and calculate duration
	p.statistics.EndTime = time.Now()
	p.statistics.Duration = p.statistics.EndTime.Sub(p.statistics.StartTime)
}

// GetStats returns the current test run statistics
func (p *TestProcessor) GetStats() *TestRunStats {
	return p.statistics
}

// RenderResults renders the current test results
func (p *TestProcessor) RenderResults(showSummary bool) error {
	// Render each test suite
	suiteRenderer := NewSuiteRenderer(p.writer, p.formatter, p.icons, p.width)

	for _, suite := range p.suites {
		// FIXED: Don't auto-collapse by default to show individual tests
		if err := suiteRenderer.RenderSuite(suite, false); err != nil {
			return err
		}
		_, _ = fmt.Fprintln(p.writer)
	}

	// Collect failed tests
	var failedTests []*TestResult
	for _, suite := range p.suites {
		for _, test := range suite.Tests {
			if test.Status == StatusFailed {
				failedTests = append(failedTests, test)
			}
		}
	}

	// Render failed tests section if any
	if len(failedTests) > 0 {
		failedRenderer := NewFailedTestRenderer(p.writer, p.formatter, p.icons, p.width)
		if err := failedRenderer.RenderFailedTests(failedTests); err != nil {
			return err
		}
	}

	// Render summary if requested
	if showSummary {
		summaryRenderer := NewSummaryRenderer(p.writer, p.formatter, p.icons, p.width)
		if err := summaryRenderer.RenderSummary(p.statistics); err != nil {
			return err
		}
	}

	return nil
}

// AddTestSuite adds a test suite to the processor's state
func (p *TestProcessor) AddTestSuite(suite *TestSuite) {
	if suite == nil {
		return
	}

	p.suites[suite.FilePath] = suite

	// Update statistics
	p.statistics.TotalTests += suite.TestCount
	p.statistics.PassedTests += suite.PassedCount
	p.statistics.FailedTests += suite.FailedCount
	p.statistics.SkippedTests += suite.SkippedCount

	// Update file statistics
	p.statistics.TotalFiles++
	if suite.FailedCount > 0 {
		p.statistics.FailedFiles++
	} else {
		p.statistics.PassedFiles++
	}
}

// createTestError creates a TestError from test output and event information
func (p *TestProcessor) createTestError(test *TestResult, event TestEvent) *TestError {
	// Extract error message from test output
	output := strings.TrimSpace(test.Output)
	lines := strings.Split(output, "\n")

	var errorMessage string
	var errorType string = "TestFailure"
	var sourceFile string
	var sourceLine int

	// Parse output to extract error information
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Look for error messages (lines that start with test file and line number)
		if strings.Contains(trimmed, ".go:") && strings.Contains(trimmed, ": ") {
			// This looks like an error line: filename.go:123: error message
			parts := strings.SplitN(trimmed, ": ", 2)
			if len(parts) == 2 {
				// Extract file and line info
				fileLinePart := parts[0]
				errorMessage = parts[1]

				// Try to extract file name and line number
				if lastSpace := strings.LastIndex(fileLinePart, " "); lastSpace != -1 {
					fileLinePart = fileLinePart[lastSpace+1:]
				}

				if colonIndex := strings.LastIndex(fileLinePart, ":"); colonIndex != -1 {
					sourceFile = fileLinePart[:colonIndex]
					if lineStr := fileLinePart[colonIndex+1:]; lineStr != "" {
						if line, err := strconv.Atoi(lineStr); err == nil {
							sourceLine = line
						}
					}
				}
				break
			}
		}

		// If no structured error found, use the first non-empty line as error message
		if errorMessage == "" && trimmed != "" && !strings.HasPrefix(trimmed, "=== ") {
			errorMessage = trimmed
		}
	}

	// Default error message if none found
	if errorMessage == "" {
		errorMessage = "Test failed"
	}

	// Detect error type from message content
	if strings.Contains(errorMessage, "panic") {
		errorType = "Panic"
	} else if strings.Contains(errorMessage, "timeout") {
		errorType = "Timeout"
	} else if strings.Contains(errorMessage, "Expected") || strings.Contains(errorMessage, "expected") {
		errorType = "AssertionError"
	}

	// Create location if we have file info
	var location *SourceLocation
	if sourceFile != "" && sourceLine > 0 {
		location = &SourceLocation{
			File:     sourceFile,
			Line:     sourceLine,
			Column:   1,
			Function: test.Name,
		}
	}

	return &TestError{
		Message:  errorMessage,
		Type:     errorType,
		Stack:    output, // Store full output as stack trace
		Location: location,
	}
}
