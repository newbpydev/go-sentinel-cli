// Package processor provides test output processing interfaces and implementations
package processor

import (
	"io"
	"time"
)

// OutputProcessor handles processing of test output in various formats
type OutputProcessor interface {
	// ProcessJSON processes JSON output from go test
	ProcessJSON(output string) (*ProcessingResult, error)

	// ProcessText processes plain text output from go test
	ProcessText(output string) (*ProcessingResult, error)

	// ProcessStream processes streaming output from go test
	ProcessStream(reader io.Reader) (*ProcessingResult, error)

	// Reset clears the processor state for a new test run
	Reset() error

	// GetResult returns the current processing result
	GetResult() *ProcessingResult
}

// EventProcessor handles individual test events
type EventProcessor interface {
	// ProcessEvent processes a single test event
	ProcessEvent(event *TestEvent) error

	// ProcessBatch processes multiple test events
	ProcessBatch(events []*TestEvent) error

	// GetProcessedEvents returns all processed events
	GetProcessedEvents() []*TestEvent
}

// TestEventParser parses test events from raw output
type TestEventParser interface {
	// ParseJSON parses JSON test events
	ParseJSON(jsonLine string) (*TestEvent, error)

	// ParseTextLine parses a single line of text output
	ParseTextLine(line string) (*TestEvent, error)

	// SupportedFormats returns the supported output formats
	SupportedFormats() []OutputFormat
}

// ResultAggregator aggregates test results across packages and suites
type ResultAggregator interface {
	// AddTestSuite adds a test suite to the aggregation
	AddTestSuite(suite *TestSuite) error

	// AddTestResult adds a single test result
	AddTestResult(result *TestResult) error

	// GetSummary returns aggregated summary statistics
	GetSummary() *TestSummary

	// GetFailedTests returns all failed tests
	GetFailedTests() []*TestResult

	// Clear clears all aggregated results
	Clear() error
}

// ProcessingResult represents the result of test output processing
type ProcessingResult struct {
	// Suites contains all processed test suites
	Suites map[string]*TestSuite

	// Summary contains aggregated statistics
	Summary *TestSummary

	// Events contains all processed events
	Events []*TestEvent

	// Errors contains any processing errors
	Errors []error

	// StartTime indicates when processing started
	StartTime time.Time

	// EndTime indicates when processing finished
	EndTime time.Time

	// ProcessingDuration is the time spent processing
	ProcessingDuration time.Duration
}

// TestEvent represents a single test event from go test output
type TestEvent struct {
	// Time indicates when the event occurred
	Time time.Time `json:"Time"`

	// Action is the event action (run, pass, fail, skip, output, etc.)
	Action string `json:"Action"`

	// Package is the package being tested
	Package string `json:"Package"`

	// Test is the test name (if applicable)
	Test string `json:"Test"`

	// Output is the test output (for output events)
	Output string `json:"Output"`

	// Elapsed is the elapsed time in seconds
	Elapsed float64 `json:"Elapsed"`
}

// TestSuite represents a collection of tests in a package
type TestSuite struct {
	// FilePath is the package path
	FilePath string

	// Tests contains all tests in the suite
	Tests []*TestResult

	// TestCount is the total number of tests
	TestCount int

	// PassedCount is the number of passed tests
	PassedCount int

	// FailedCount is the number of failed tests
	FailedCount int

	// SkippedCount is the number of skipped tests
	SkippedCount int

	// Duration is the total execution time
	Duration time.Duration

	// Coverage is the coverage percentage
	Coverage float64
}

// TestResult represents the result of a single test
type TestResult struct {
	// Name is the test name
	Name string

	// Package is the package containing the test
	Package string

	// Status is the test status
	Status TestStatus

	// Duration is the test execution time
	Duration time.Duration

	// Error contains error details if the test failed
	Error *TestError

	// Output contains the test output
	Output []string

	// Parent is the parent test name (for subtests)
	Parent string

	// Subtests contains any subtests
	Subtests []*TestResult

	// StartTime indicates when the test started
	StartTime time.Time

	// EndTime indicates when the test finished
	EndTime time.Time
}

// TestError represents detailed error information for a failed test
type TestError struct {
	// Message is the error message
	Message string

	// StackTrace contains the stack trace
	StackTrace []string

	// SourceFile is the source file where the error occurred
	SourceFile string

	// SourceLine is the line number where the error occurred
	SourceLine int

	// SourceColumn is the column number where the error occurred
	SourceColumn int

	// SourceContext contains surrounding source code lines
	SourceContext []string

	// ContextStartLine is the starting line number for the context
	ContextStartLine int
}

// TestSummary contains aggregated test statistics
type TestSummary struct {
	// TotalTests is the total number of tests
	TotalTests int

	// PassedTests is the number of passed tests
	PassedTests int

	// FailedTests is the number of failed tests
	FailedTests int

	// SkippedTests is the number of skipped tests
	SkippedTests int

	// TotalDuration is the total execution time
	TotalDuration time.Duration

	// AverageDuration is the average test execution time
	AverageDuration time.Duration

	// CoveragePercentage is the overall coverage percentage
	CoveragePercentage float64

	// PackageCount is the number of packages tested
	PackageCount int

	// Success indicates if all tests passed
	Success bool
}

// TestStatus represents the status of a test
type TestStatus string

const (
	// StatusRunning indicates the test is currently running
	StatusRunning TestStatus = "running"

	// StatusPassed indicates the test passed
	StatusPassed TestStatus = "passed"

	// StatusFailed indicates the test failed
	StatusFailed TestStatus = "failed"

	// StatusSkipped indicates the test was skipped
	StatusSkipped TestStatus = "skipped"
)

// OutputFormat represents different test output formats
type OutputFormat string

const (
	// OutputFormatJSON represents JSON output format
	OutputFormatJSON OutputFormat = "json"

	// OutputFormatText represents plain text output format
	OutputFormatText OutputFormat = "text"

	// OutputFormatTAP represents TAP (Test Anything Protocol) format
	OutputFormatTAP OutputFormat = "tap"
)
