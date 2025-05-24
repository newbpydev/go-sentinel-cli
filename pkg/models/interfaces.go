// Package models provides shared data models and value objects
package models

import (
	"time"
)

// TestResult represents the result of a test execution
type TestResult struct {
	// ID is the unique test result identifier
	ID string

	// Name is the test name
	Name string

	// Package is the package containing the test
	Package string

	// Status is the test execution status
	Status TestStatus

	// Duration is the test execution time
	Duration time.Duration

	// StartTime is when the test started
	StartTime time.Time

	// EndTime is when the test finished
	EndTime time.Time

	// Output contains the test output lines
	Output []string

	// Error contains error details if the test failed
	Error *TestError

	// Coverage contains coverage information
	Coverage *TestCoverage

	// Subtests contains any subtest results
	Subtests []*TestResult

	// Parent is the parent test name (for subtests)
	Parent string

	// Metadata contains additional test metadata
	Metadata map[string]interface{}
}

// PackageResult represents the result of testing a package
type PackageResult struct {
	// Package is the package name/path
	Package string

	// Success indicates if all tests in the package passed
	Success bool

	// Duration is the total execution time for the package
	Duration time.Duration

	// StartTime is when package testing started
	StartTime time.Time

	// EndTime is when package testing finished
	EndTime time.Time

	// Tests contains individual test results
	Tests []*TestResult

	// Coverage contains package coverage information
	Coverage *PackageCoverage

	// TestCount is the total number of tests
	TestCount int

	// PassedCount is the number of passed tests
	PassedCount int

	// FailedCount is the number of failed tests
	FailedCount int

	// SkippedCount is the number of skipped tests
	SkippedCount int

	// Output contains the raw package output
	Output string

	// Error contains any package-level error
	Error error

	// Metadata contains additional package metadata
	Metadata map[string]interface{}
}

// TestSummary contains aggregated test statistics
type TestSummary struct {
	// TotalTests is the total number of tests executed
	TotalTests int

	// PassedTests is the number of tests that passed
	PassedTests int

	// FailedTests is the number of tests that failed
	FailedTests int

	// SkippedTests is the number of tests that were skipped
	SkippedTests int

	// TotalDuration is the total execution time
	TotalDuration time.Duration

	// AverageDuration is the average test execution time
	AverageDuration time.Duration

	// PackageCount is the number of packages tested
	PackageCount int

	// CoveragePercentage is the overall coverage percentage
	CoveragePercentage float64

	// Success indicates if all tests passed
	Success bool

	// StartTime is when testing started
	StartTime time.Time

	// EndTime is when testing finished
	EndTime time.Time

	// FailedPackages contains names of packages with failed tests
	FailedPackages []string

	// Metadata contains additional summary metadata
	Metadata map[string]interface{}
}

// TestError contains detailed error information for a failed test
type TestError struct {
	// Message is the primary error message
	Message string

	// Type is the error type or category
	Type string

	// StackTrace contains the stack trace lines
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

	// Expected contains the expected value (for assertion errors)
	Expected string

	// Actual contains the actual value (for assertion errors)
	Actual string

	// Metadata contains additional error metadata
	Metadata map[string]interface{}
}

// TestCoverage contains coverage information for a test
type TestCoverage struct {
	// Percentage is the coverage percentage
	Percentage float64

	// CoveredLines is the number of covered lines
	CoveredLines int

	// TotalLines is the total number of lines
	TotalLines int

	// CoveredStatements is the number of covered statements
	CoveredStatements int

	// TotalStatements is the total number of statements
	TotalStatements int

	// Files contains per-file coverage information
	Files map[string]*FileCoverage

	// Metadata contains additional coverage metadata
	Metadata map[string]interface{}
}

// PackageCoverage contains coverage information for a package
type PackageCoverage struct {
	// Package is the package name
	Package string

	// Percentage is the overall coverage percentage
	Percentage float64

	// CoveredLines is the total number of covered lines
	CoveredLines int

	// TotalLines is the total number of lines in the package
	TotalLines int

	// CoveredStatements is the total number of covered statements
	CoveredStatements int

	// TotalStatements is the total number of statements in the package
	TotalStatements int

	// Files contains coverage for each file in the package
	Files map[string]*FileCoverage

	// Functions contains coverage for each function
	Functions map[string]*FunctionCoverage

	// Metadata contains additional coverage metadata
	Metadata map[string]interface{}
}

// FileCoverage contains coverage information for a single file
type FileCoverage struct {
	// FilePath is the path to the file
	FilePath string

	// Percentage is the coverage percentage for this file
	Percentage float64

	// CoveredLines is the number of covered lines
	CoveredLines int

	// TotalLines is the total number of lines
	TotalLines int

	// CoveredStatements is the number of covered statements
	CoveredStatements int

	// TotalStatements is the total number of statements
	TotalStatements int

	// LinesCovered contains the specific lines that are covered
	LinesCovered []int

	// LinesUncovered contains the specific lines that are not covered
	LinesUncovered []int

	// Metadata contains additional file coverage metadata
	Metadata map[string]interface{}
}

// FunctionCoverage contains coverage information for a single function
type FunctionCoverage struct {
	// Name is the function name
	Name string

	// FilePath is the file containing the function
	FilePath string

	// StartLine is the starting line of the function
	StartLine int

	// EndLine is the ending line of the function
	EndLine int

	// Percentage is the coverage percentage for this function
	Percentage float64

	// IsCovered indicates if the function is covered by tests
	IsCovered bool

	// CallCount is the number of times the function was called during tests
	CallCount int

	// Metadata contains additional function coverage metadata
	Metadata map[string]interface{}
}

// FileChange represents a change to a file
type FileChange struct {
	// FilePath is the path to the changed file
	FilePath string

	// ChangeType is the type of change (created, modified, deleted, renamed)
	ChangeType ChangeType

	// Timestamp is when the change occurred
	Timestamp time.Time

	// OldPath is the old path (for rename operations)
	OldPath string

	// Size is the file size after the change
	Size int64

	// Checksum is the file checksum after the change
	Checksum string

	// Metadata contains additional change metadata
	Metadata map[string]interface{}
}

// TestConfiguration represents test execution configuration
type TestConfiguration struct {
	// Packages contains the packages to test
	Packages []string

	// TestFiles contains specific test files to run
	TestFiles []string

	// TestPatterns contains test name patterns to match
	TestPatterns []string

	// Verbose enables verbose output
	Verbose bool

	// Coverage enables coverage reporting
	Coverage bool

	// CoverageProfile specifies the coverage profile file
	CoverageProfile string

	// JSONOutput enables JSON output format
	JSONOutput bool

	// Parallel specifies the number of parallel test processes
	Parallel int

	// Timeout specifies the test execution timeout
	Timeout time.Duration

	// Tags contains build tags to use
	Tags []string

	// Args contains additional arguments to pass to go test
	Args []string

	// Environment contains environment variables for test execution
	Environment map[string]string

	// WorkingDirectory specifies the working directory for execution
	WorkingDirectory string

	// Metadata contains additional configuration metadata
	Metadata map[string]interface{}
}

// WatchConfiguration represents watch mode configuration
type WatchConfiguration struct {
	// Enabled indicates if watch mode is enabled
	Enabled bool

	// Paths contains the paths to watch
	Paths []string

	// IgnorePatterns contains patterns to ignore
	IgnorePatterns []string

	// TestPatterns contains test file patterns
	TestPatterns []string

	// DebounceInterval is the debounce interval for file changes
	DebounceInterval time.Duration

	// RunOnStart indicates if tests should run on startup
	RunOnStart bool

	// ClearOnRerun indicates if the terminal should be cleared between runs
	ClearOnRerun bool

	// NotifyOnSuccess indicates if notifications should be sent on success
	NotifyOnSuccess bool

	// NotifyOnFailure indicates if notifications should be sent on failure
	NotifyOnFailure bool

	// Metadata contains additional watch configuration metadata
	Metadata map[string]interface{}
}

// TestStatus represents the status of a test
type TestStatus string

const (
	// TestStatusPending indicates the test is pending execution
	TestStatusPending TestStatus = "pending"

	// TestStatusRunning indicates the test is currently running
	TestStatusRunning TestStatus = "running"

	// TestStatusPassed indicates the test passed
	TestStatusPassed TestStatus = "passed"

	// TestStatusFailed indicates the test failed
	TestStatusFailed TestStatus = "failed"

	// TestStatusSkipped indicates the test was skipped
	TestStatusSkipped TestStatus = "skipped"

	// TestStatusTimeout indicates the test timed out
	TestStatusTimeout TestStatus = "timeout"

	// TestStatusError indicates an error occurred during test execution
	TestStatusError TestStatus = "error"
)

// ChangeType represents the type of file change
type ChangeType string

const (
	// ChangeTypeCreated indicates a file was created
	ChangeTypeCreated ChangeType = "created"

	// ChangeTypeModified indicates a file was modified
	ChangeTypeModified ChangeType = "modified"

	// ChangeTypeDeleted indicates a file was deleted
	ChangeTypeDeleted ChangeType = "deleted"

	// ChangeTypeRenamed indicates a file was renamed
	ChangeTypeRenamed ChangeType = "renamed"

	// ChangeTypeMoved indicates a file was moved
	ChangeTypeMoved ChangeType = "moved"
)

// NewTestResult creates a new TestResult with default values
func NewTestResult(name, pkg string) *TestResult {
	return &TestResult{
		ID:        generateID(),
		Name:      name,
		Package:   pkg,
		Status:    TestStatusPending,
		StartTime: time.Now(),
		Output:    make([]string, 0),
		Subtests:  make([]*TestResult, 0),
		Metadata:  make(map[string]interface{}),
	}
}

// NewPackageResult creates a new PackageResult with default values
func NewPackageResult(pkg string) *PackageResult {
	return &PackageResult{
		Package:   pkg,
		StartTime: time.Now(),
		Tests:     make([]*TestResult, 0),
		Metadata:  make(map[string]interface{}),
	}
}

// NewTestSummary creates a new TestSummary with default values
func NewTestSummary() *TestSummary {
	return &TestSummary{
		StartTime:      time.Now(),
		FailedPackages: make([]string, 0),
		Metadata:       make(map[string]interface{}),
	}
}

// NewFileChange creates a new FileChange
func NewFileChange(filePath string, changeType ChangeType) *FileChange {
	return &FileChange{
		FilePath:   filePath,
		ChangeType: changeType,
		Timestamp:  time.Now(),
		Metadata:   make(map[string]interface{}),
	}
}

// IsSuccess returns whether the test result represents a successful test
func (tr *TestResult) IsSuccess() bool {
	return tr.Status == TestStatusPassed
}

// IsFailure returns whether the test result represents a failed test
func (tr *TestResult) IsFailure() bool {
	return tr.Status == TestStatusFailed || tr.Status == TestStatusTimeout || tr.Status == TestStatusError
}

// IsComplete returns whether the test has completed execution
func (tr *TestResult) IsComplete() bool {
	return tr.Status != TestStatusPending && tr.Status != TestStatusRunning
}

// AddSubtest adds a subtest result
func (tr *TestResult) AddSubtest(subtest *TestResult) {
	subtest.Parent = tr.Name
	tr.Subtests = append(tr.Subtests, subtest)
}

// GetSuccessRate returns the success rate for the package
func (pr *PackageResult) GetSuccessRate() float64 {
	if pr.TestCount == 0 {
		return 0.0
	}
	return float64(pr.PassedCount) / float64(pr.TestCount)
}

// AddTest adds a test result to the package
func (pr *PackageResult) AddTest(test *TestResult) {
	pr.Tests = append(pr.Tests, test)
	pr.TestCount++

	switch test.Status {
	case TestStatusPassed:
		pr.PassedCount++
	case TestStatusFailed, TestStatusTimeout, TestStatusError:
		pr.FailedCount++
		pr.Success = false
	case TestStatusSkipped:
		pr.SkippedCount++
	}
}

// GetSuccessRate returns the overall success rate
func (ts *TestSummary) GetSuccessRate() float64 {
	if ts.TotalTests == 0 {
		return 0.0
	}
	return float64(ts.PassedTests) / float64(ts.TotalTests)
}

// AddPackageResult adds a package result to the summary
func (ts *TestSummary) AddPackageResult(pkg *PackageResult) {
	ts.PackageCount++
	ts.TotalTests += pkg.TestCount
	ts.PassedTests += pkg.PassedCount
	ts.FailedTests += pkg.FailedCount
	ts.SkippedTests += pkg.SkippedCount
	ts.TotalDuration += pkg.Duration

	if !pkg.Success {
		ts.Success = false
		ts.FailedPackages = append(ts.FailedPackages, pkg.Package)
	}

	// Update average duration
	if ts.TotalTests > 0 {
		ts.AverageDuration = ts.TotalDuration / time.Duration(ts.TotalTests)
	}
}

// generateID generates a unique identifier
func generateID() string {
	return time.Now().Format("20060102150405.000000")
}
