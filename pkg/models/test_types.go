package models

import (
	"io"
	"time"
)

// TestEvent represents a JSON event from Go test output
type TestEvent struct {
	Time    string  `json:"Time"`    // RFC3339Nano formatted timestamp
	Action  string  `json:"Action"`  // run, pass, fail, skip, output
	Package string  `json:"Package"` // Package being tested
	Test    string  `json:"Test"`    // Name of the test
	Output  string  `json:"Output"`  // Test output (stdout/stderr)
	Elapsed float64 `json:"Elapsed"` // Test duration in seconds
}

// Legacy constants for backward compatibility during migration
const (
	// StatusPassed indicates a test has passed (legacy)
	StatusPassed = TestStatusPassed
	// StatusFailed indicates a test has failed (legacy)
	StatusFailed = TestStatusFailed
	// StatusSkipped indicates a test was skipped (legacy)
	StatusSkipped = TestStatusSkipped
	// StatusRunning indicates a test is currently running (legacy)
	StatusRunning = TestStatusRunning
)

// TestProgress represents real-time progress information
type TestProgress struct {
	CompletedTests int
	TotalTests     int
	CurrentFile    string
	Status         TestStatus
}

// TestRunStats contains statistics about a test run
type TestRunStats struct {
	// Test file statistics
	TotalFiles  int
	PassedFiles int
	FailedFiles int

	// Test statistics
	TotalTests   int
	PassedTests  int
	FailedTests  int
	SkippedTests int

	// Timing
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration

	// Real phase durations (only populated with actual measurements)
	Phases map[string]time.Duration
}

// TestSuite represents a collection of tests from a single file
type TestSuite struct {
	// FilePath is the path to the test file
	FilePath string
	// Tests is the collection of test results
	Tests []*TestResult
	// Duration is the total time taken to run all tests
	Duration time.Duration
	// MemoryUsage is the memory used during the test run
	MemoryUsage uint64
	// TestCount is the total number of tests
	TestCount int
	// PassedCount is the number of passed tests
	PassedCount int
	// FailedCount is the number of failed tests
	FailedCount int
	// SkippedCount is the number of skipped tests
	SkippedCount int
}

// TestProcessorInterface defines the interface for test processors
type TestProcessorInterface interface {
	ProcessJSONOutput(output string) error
	ProcessStream(r io.Reader, progress chan<- TestProgress) error
	Reset()
	GetStats() *TestRunStats
	RenderResults(showSummary bool) error
	AddTestSuite(suite *TestSuite)
}
