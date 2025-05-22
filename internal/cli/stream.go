package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"
)

// GoTestEvent represents a test event from the Go test JSON output
type GoTestEvent struct {
	Action  string  `json:"Action"`
	Test    string  `json:"Test"`
	Package string  `json:"Package"`
	Output  string  `json:"Output"`
	Elapsed float64 `json:"Elapsed"`
}

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

		var event GoTestEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return fmt.Errorf("error parsing JSON: %w", err)
		}

		// Process the event
		p.processEvent(&event, results)
	}

	return scanner.Err()
}

// processEvent processes a single test event
func (p *StreamParser) processEvent(event *GoTestEvent, results chan<- *TestResult) {
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

// TestProgress represents real-time progress information
type TestProgress struct {
	CompletedTests int
	TotalTests     int
	CurrentFile    string
	Status         TestStatus
}

// TestProcessor processes test output and tracks statistics
type TestProcessor struct {
	writer     io.Writer
	formatter  *ColorFormatter
	icons      *IconProvider
	width      int
	suites     map[string]*TestSuite
	statistics *TestRunStats
	startTime  time.Time
}

// NewTestProcessor creates a new TestProcessor
func NewTestProcessor(writer io.Writer, formatter *ColorFormatter, icons *IconProvider, width int) *TestProcessor {
	return &TestProcessor{
		writer:    writer,
		formatter: formatter,
		icons:     icons,
		width:     width,
		suites:    make(map[string]*TestSuite),
		statistics: &TestRunStats{
			StartTime:     time.Now(),
			PhaseDuration: make(map[string]time.Duration),
		},
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

	// Process test results as they arrive
	totalTests := 0
	completedTests := 0

	for result := range resultCh {
		totalTests++

		// Get or create the suite for this result
		suitePath := filepath.Base(result.Package) + "_test.go"
		suite, ok := p.suites[suitePath]
		if !ok {
			suite = &TestSuite{
				FilePath: suitePath,
			}
			p.suites[suitePath] = suite
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

		// If the test is completed, increment counter
		if result.Status != StatusRunning {
			completedTests++
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
	p.statistics.Duration = time.Since(p.startTime)

	// Count passed and failed files
	for _, suite := range p.suites {
		if suite.FailedCount > 0 {
			p.statistics.FailedFiles++
		} else {
			p.statistics.PassedFiles++
		}
	}

	return nil
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
		if err := suiteRenderer.RenderSuite(suite, true); err != nil {
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
		summaryRenderer := NewSummaryRenderer(p.writer, p.formatter)
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
