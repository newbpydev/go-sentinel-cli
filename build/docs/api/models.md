package models // import "github.com/newbpydev/go-sentinel/pkg/models"

Package models provides core data structures for test execution and result
management.

This file contains fundamental data models used throughout the Go Sentinel CLI
for representing test results, error contexts, and package-level information.
These models form the foundation of the test execution and reporting system.

Key components:
  - SourceLocation: Precise location information for code references
  - TestPackage: Package-level test aggregation and statistics
  - FailedTestDetail: Comprehensive information about test failures
  - LegacyTestResult: Backward compatibility during migration
  - LegacyTestError: Error representation for legacy compatibility

Design principles:
  - Immutable data structures where possible
  - Rich metadata for debugging and reporting
  - Backward compatibility during migration phase
  - Clear separation between data and behavior

Example usage:

    // Creating a test package result
    pkg := &TestPackage{
    	Package:     "github.com/example/auth",
    	Duration:    150 * time.Millisecond,
    	TestCount:   5,
    	PassedCount: 4,
    	FailedCount: 1,
    	Passed:      false,
    }

    // Creating source location information
    location := &SourceLocation{
    	File:        "auth.go",
    	Line:        42,
    	Column:      15,
    	Function:    "ValidatePassword",
    	Context:     []string{"func ValidatePassword(pass string) bool {", "    return len(pass) > 8", "}"},
    	ContextLine: 1,
    }

# Package models provides shared error types and error handling utilities

Package models provides shared data models and value objects for the Go Sentinel
CLI.

This package contains core data structures used throughout the application for
representing test results, error handling, file changes, and configuration. It
follows the principle of providing clean value objects without business logic.

Key components:
  - Error handling: SentinelError with comprehensive error context and stack
    traces
  - Test results: TestResult, PackageResult, and TestSummary for test execution
    data
  - File changes: FileChange for representing file system modifications
  - Configuration: TestConfiguration and WatchConfiguration for application
    settings

Example usage:

    // Creating and using test results
    result := models.NewTestResult("TestExample", "github.com/example/pkg")
    result.Status = models.TestStatusPassed
    result.Duration = 100 * time.Millisecond

    // Creating and handling errors
    err := models.NewValidationError("config.timeout", "timeout must be positive")
    if models.IsErrorType(err, models.ErrorTypeValidation) {
    	fmt.Println("Validation error:", err.UserMessage())
    }

    // Creating file change events
    change := models.NewFileChange("main.go", models.ChangeTypeModified)
    fmt.Printf("File %s was %s at %v\n", change.FilePath, change.ChangeType, change.Timestamp)

# Package models provides shared data models and value objects

Package models provides type definitions and interfaces for test execution and
processing.

This file contains legacy types and interfaces used during the migration
from the monolithic CLI package to the modular architecture. These types
provide backward compatibility while the system transitions to the new model
architecture.

Key components:
  - TestEvent: JSON event structure from Go test output
  - TestProgress: Real-time test execution progress tracking
  - TestRunStats: Comprehensive statistics about test runs
  - TestSuite: Collection of tests from a single file
  - TestProcessorInterface: Interface for test output processors

Legacy compatibility:
  - StatusPassed, StatusFailed, etc.: Deprecated constants for backward
    compatibility
  - Will be removed in favor of TestStatus* constants

Example usage:

    // Processing test events
    event := &TestEvent{
    	Time:    time.Now().Format(time.RFC3339Nano),
    	Action:  "pass",
    	Package: "github.com/example/pkg",
    	Test:    "TestExample",
    	Elapsed: 0.15,
    }

    // Tracking test progress
    progress := &TestProgress{
    	CompletedTests: 5,
    	TotalTests:     10,
    	CurrentFile:    "main_test.go",
    	Status:         TestStatusRunning,
    }

CONSTANTS

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
    Legacy constants for backward compatibility during migration


FUNCTIONS

func Example_configuration()
    Example_configuration demonstrates creating and using configuration objects.

    This example shows how to set up test and watch configurations for different
    scenarios.

func Example_coverage()
    Example_coverage demonstrates working with test coverage data.

    This example shows how to create and analyze coverage information at
    different levels (test, file, function, package).

func Example_errorHandling()
    Example_errorHandling demonstrates the comprehensive error handling system.

    This example shows how to create, wrap, and handle different types of errors
    with proper context and user-safe messaging.

func Example_fileChanges()
    Example_fileChanges demonstrates tracking file system changes.

    This example shows how to create and work with file change events for watch
    mode functionality.

func Example_testResults()
    Example_testResults demonstrates working with test execution results.

    This example shows how to create test results, manage package results,
    and generate comprehensive test summaries.

func Example_testStatus()
    Example_testStatus demonstrates working with test status values.

    This example shows the different test statuses and how to check them.

func IsErrorSeverity(err error, severity ErrorSeverity) bool
    IsErrorSeverity checks if an error has a specific severity

func IsErrorType(err error, errorType ErrorType) bool
    IsErrorType checks if an error is of a specific type

func SanitizeError(err error) error
    SanitizeError returns a user-safe version of any error


TYPES

type ChangeType string
    ChangeType represents the type of file change

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
type ErrorContext struct {
	Operation string            `json:"operation"` // What operation was being performed
	Component string            `json:"component"` // Which component generated the error
	Resource  string            `json:"resource"`  // What resource was involved (file, package, etc.)
	Metadata  map[string]string `json:"metadata"`  // Additional context-specific data
	RequestID string            `json:"requestId"` // For tracing across operations
	UserID    string            `json:"userId"`    // For user-specific operations
}
    ErrorContext provides additional context about where/when the error occurred

func GetErrorContext(err error) *ErrorContext
    GetErrorContext extracts context from a SentinelError

type ErrorSeverity string
    ErrorSeverity represents how critical an error is

const (
	SeverityInfo     ErrorSeverity = "INFO"
	SeverityWarning  ErrorSeverity = "WARNING"
	SeverityError    ErrorSeverity = "ERROR"
	SeverityCritical ErrorSeverity = "CRITICAL"
)
type ErrorType string
    ErrorType represents the category of error

const (
	// Configuration errors
	ErrorTypeConfig ErrorType = "CONFIG"

	// File system and I/O errors
	ErrorTypeFileSystem ErrorType = "FILESYSTEM"

	// Test execution errors
	ErrorTypeTestExecution ErrorType = "TEST_EXECUTION"

	// Watch system errors
	ErrorTypeWatch ErrorType = "WATCH"

	// Dependency injection errors
	ErrorTypeDependency ErrorType = "DEPENDENCY"

	// Application lifecycle errors
	ErrorTypeLifecycle ErrorType = "LIFECYCLE"

	// Validation errors
	ErrorTypeValidation ErrorType = "VALIDATION"

	// Network/external service errors
	ErrorTypeExternal ErrorType = "EXTERNAL"

	// Internal system errors
	ErrorTypeInternal ErrorType = "INTERNAL"
)
type FailedTestDetail struct {
	// Result is the test result
	Result *LegacyTestResult
	// Suite is the test suite the test belongs to
	Suite *TestSuite
	// SourceCode contains the relevant source code
	SourceCode []string
	// ErrorLine is the line number where the error occurred
	ErrorLine int
	// FormattedError is the formatted error message
	FormattedError string
}
    FailedTestDetail represents detailed information about a failed test

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
    FileChange represents a change to a file

func NewFileChange(filePath string, changeType ChangeType) *FileChange
    NewFileChange creates a new FileChange

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
    FileCoverage contains coverage information for a single file

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
    FunctionCoverage contains coverage information for a single function

type LegacyTestError struct {
	// Message is the error message
	Message string
	// Type is the error type (e.g., "TypeError")
	Type string
	// Stack is the error stack trace
	Stack string
	// Expected value in assertions
	Expected string
	// Actual value in assertions
	Actual string
	// Location information
	Location *SourceLocation
	// SourceContext contains lines of source code around the error
	SourceContext []string
	// HighlightedLine is the index in SourceContext that contains the error
	HighlightedLine int
}
    LegacyTestError for backward compatibility during migration

type LegacyTestResult struct {
	// Name is the name of the test
	Name string
	// Status is the test status (passed, failed, skipped)
	Status TestStatus
	// Duration is how long the test took to run
	Duration time.Duration
	// Error contains error information if the test failed
	Error *LegacyTestError
	// Package is the Go package the test belongs to
	Package string
	// Test is the Go test name
	Test string
	// Output contains any test output
	Output string
	// Parent indicates the parent test for subtests
	Parent string
	// Subtests contains any subtests
	Subtests []*LegacyTestResult
}
    LegacyTestResult for backward compatibility during migration This matches
    the structure from internal/cli/models.go

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
    PackageCoverage contains coverage information for a package

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
    PackageResult represents the result of testing a package

func NewPackageResult(pkg string) *PackageResult
    NewPackageResult creates a new PackageResult with default values

func (pr *PackageResult) AddTest(test *TestResult)
    AddTest adds a test result to the package

func (pr *PackageResult) GetSuccessRate() float64
    GetSuccessRate returns the success rate for the package

type SentinelError struct {
	Type     ErrorType     `json:"type"`
	Severity ErrorSeverity `json:"severity"`
	Message  string        `json:"message"`
	Cause    error         `json:"cause,omitempty"`
	Context  ErrorContext  `json:"context"`
	Stack    []StackFrame  `json:"stack,omitempty"`
	UserSafe bool          `json:"userSafe"` // Whether safe to show to end users
}
    SentinelError is the base error type for all application errors

func NewConfigError(message string, userSafe bool) *SentinelError
    NewConfigError creates a configuration-related error

func NewDependencyError(component string, cause error) *SentinelError
    NewDependencyError creates a dependency injection error

func NewError(errorType ErrorType, severity ErrorSeverity, message string) *SentinelError
    NewError creates a new SentinelError with the specified type and message

func NewFileSystemError(operation, path string, cause error) *SentinelError
    NewFileSystemError creates a file system error

func NewInternalError(component, operation string, cause error) *SentinelError
    NewInternalError creates an internal system error (never user-safe)

func NewLifecycleError(operation string, cause error) *SentinelError
    NewLifecycleError creates an application lifecycle error

func NewTestExecutionError(testPath string, cause error) *SentinelError
    NewTestExecutionError creates a test execution error

func NewValidationError(field, message string) *SentinelError
    NewValidationError creates a validation error (always user-safe)

func NewWatchError(operation, path string, cause error) *SentinelError
    NewWatchError creates a watch system error

func WrapError(err error, errorType ErrorType, severity ErrorSeverity, message string) *SentinelError
    WrapError wraps an existing error with additional context

func (e *SentinelError) Error() string
    Error implements the error interface

func (e *SentinelError) Is(target error) bool
    Is implements error comparison for errors.Is

func (e *SentinelError) Unwrap() error
    Unwrap returns the underlying cause for error unwrapping

func (e *SentinelError) UserMessage() string
    UserMessage returns a sanitized message safe for end users

func (e *SentinelError) WithContext(key, value string) *SentinelError
    WithContext adds additional context to the error

func (e *SentinelError) WithRequestID(requestID string) *SentinelError
    WithRequestID adds a request ID for tracing

func (e *SentinelError) WithUserID(userID string) *SentinelError
    WithUserID adds a user ID for user-specific operations

type SourceLocation struct {
	// File is the file path
	File string
	// Line is the line number
	Line int
	// Column is the column number
	Column int
	// Function is the function name
	Function string
	// Context contains the lines of code around the error
	Context []string
	// ContextLine is the index of the error line in Context
	ContextLine int
}
    SourceLocation represents a location in source code

type StackFrame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}
    StackFrame represents a single frame in the call stack

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
    TestConfiguration represents test execution configuration

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
    TestCoverage contains coverage information for a test

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
    TestError contains detailed error information for a failed test

type TestEvent struct {
	Time    string  `json:"Time"`    // RFC3339Nano formatted timestamp
	Action  string  `json:"Action"`  // run, pass, fail, skip, output
	Package string  `json:"Package"` // Package being tested
	Test    string  `json:"Test"`    // Name of the test
	Output  string  `json:"Output"`  // Test output (stdout/stderr)
	Elapsed float64 `json:"Elapsed"` // Test duration in seconds
}
    TestEvent represents a JSON event from Go test output

type TestPackage struct {
	// Package is the package name/path
	Package string
	// Tests is all tests in this package
	Tests []*LegacyTestResult
	// Duration is the total duration for all tests in the package
	Duration time.Duration
	// BuildFailed indicates if there was a build error
	BuildFailed bool
	// BuildError contains any build error message
	BuildError string
	// Passed indicates if all tests in the package passed
	Passed bool
	// MemoryUsage is the memory used by the package tests
	MemoryUsage uint64
	// TestCount is the total number of tests in the package
	TestCount int
	// PassedCount is the number of passed tests
	PassedCount int
	// FailedCount is the number of failed tests
	FailedCount int
	// SkippedCount is the number of skipped tests
	SkippedCount int
}
    TestPackage represents all tests from a single package

type TestProcessorInterface interface {
	ProcessJSONOutput(output string) error
	ProcessStream(r io.Reader, progress chan<- TestProgress) error
	Reset()
	GetStats() *TestRunStats
	RenderResults(showSummary bool) error
	AddTestSuite(suite *TestSuite)
}
    TestProcessorInterface defines the interface for test processors

type TestProgress struct {
	CompletedTests int
	TotalTests     int
	CurrentFile    string
	Status         TestStatus
}
    TestProgress represents real-time progress information

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
    TestResult represents the result of a test execution

func NewTestResult(name, pkg string) *TestResult
    NewTestResult creates a new TestResult with default values

func (tr *TestResult) AddSubtest(subtest *TestResult)
    AddSubtest adds a subtest result

func (tr *TestResult) IsComplete() bool
    IsComplete returns whether the test has completed execution

func (tr *TestResult) IsFailure() bool
    IsFailure returns whether the test result represents a failed test

func (tr *TestResult) IsSuccess() bool
    IsSuccess returns whether the test result represents a successful test

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
    TestRunStats contains statistics about a test run

type TestStatus string
    TestStatus represents the status of a test

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
type TestSuite struct {
	// FilePath is the path to the test file
	FilePath string
	// Tests is the collection of test results
	Tests []*LegacyTestResult
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
    TestSuite represents a collection of tests from a single file

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
    TestSummary contains aggregated test statistics

func NewTestSummary() *TestSummary
    NewTestSummary creates a new TestSummary with default values

func (ts *TestSummary) AddPackageResult(pkg *PackageResult)
    AddPackageResult adds a package result to the summary

func (ts *TestSummary) GetSuccessRate() float64
    GetSuccessRate returns the overall success rate

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
    WatchConfiguration represents watch mode configuration

