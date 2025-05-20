package cli

import "time"

// TestStatus represents the state of a test
type TestStatus int

const (
	TestStatusPending TestStatus = iota
	TestStatusRunning
	TestStatusPassed
	TestStatusFailed
	TestStatusSkipped
)

// SourceLocation represents a location in source code
type SourceLocation struct {
	File      string
	Line      int
	Column    int
	Snippet   string
	StartLine int // Starting line for context
}

// TestError represents a test failure
type TestError struct {
	Message  string
	Location *SourceLocation
	Snippet  string
	Expected string // Expected value for assertions
	Actual   string // Actual value for assertions
}

// TestResult represents the result of a single test
type TestResult struct {
	Name      string
	Status    TestStatus
	Duration  time.Duration
	Error     *TestError
	Depth     int // For subtests
	StartTime time.Time
	EndTime   time.Time
}

// TestSuite represents a collection of tests from a package
type TestSuite struct {
	Package     string
	PackageName string // Full package import path
	FilePath    string // Path to test file
	Tests       []*TestResult
	Errors      []*TestError
	NumTotal    int
	NumPassed   int
	NumFailed   int
	NumSkipped  int
	Duration    time.Duration
	StartTime   time.Time
	EndTime     time.Time
}

// TestRun represents a complete test execution
type TestRun struct {
	Suites     []*TestSuite
	NumTotal   int
	NumPassed  int
	NumFailed  int
	NumSkipped int
	Duration   time.Duration
	StartTime  time.Time
	EndTime    time.Time
}
