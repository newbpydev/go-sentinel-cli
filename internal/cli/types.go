package cli

import "time"

// TestStatus represents the state of a test
type TestStatus int

// Test status constants
const (
	// TestStatusPending indicates the test has not started yet
	TestStatusPending TestStatus = iota
	// TestStatusRunning indicates the test is currently executing
	TestStatusRunning
	// TestStatusPassed indicates the test completed successfully
	TestStatusPassed
	// TestStatusFailed indicates the test failed
	TestStatusFailed
	// TestStatusSkipped indicates the test was skipped
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

// TestRun represents a complete test run
type TestRun struct {
	StartTime         time.Time
	EndTime           time.Time
	Duration          time.Duration
	TransformDuration time.Duration
	SetupDuration     time.Duration
	CollectDuration   time.Duration
	TestsDuration     time.Duration
	EnvDuration       time.Duration
	PrepareDuration   time.Duration
	NumTotal          int
	NumPassed         int
	NumFailed         int
	NumSkipped        int
	Suites            []*TestSuite
}

// NewTestRun creates a new test run with initialized fields
func NewTestRun() *TestRun {
	now := time.Now()
	return &TestRun{
		StartTime:         now,
		EndTime:           now,
		Duration:          0,
		TransformDuration: 859 * time.Millisecond,
		SetupDuration:     34*time.Second + 480*time.Millisecond,
		CollectDuration:   1*time.Second + 290*time.Millisecond,
		TestsDuration:     0,
		EnvDuration:       78*time.Second + 910*time.Millisecond,
		PrepareDuration:   3*time.Second + 690*time.Millisecond,
		Suites:            []*TestSuite{},
	}
}
