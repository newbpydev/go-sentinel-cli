// Package cli provides a Vitest-like CLI for Go testing
package cli

import (
	"time"
)

// TestStatus represents the status of a test
type TestStatus string

const (
	// StatusPassed indicates a test has passed
	StatusPassed TestStatus = "passed"
	// StatusFailed indicates a test has failed
	StatusFailed TestStatus = "failed"
	// StatusSkipped indicates a test was skipped
	StatusSkipped TestStatus = "skipped"
	// StatusRunning indicates a test is currently running
	StatusRunning TestStatus = "running"
)

// TestResult represents the result of a single test
type TestResult struct {
	// Name is the name of the test
	Name string
	// Status is the test status (passed, failed, skipped)
	Status TestStatus
	// Duration is how long the test took to run
	Duration time.Duration
	// Error contains error information if the test failed
	Error *TestError
	// Package is the Go package the test belongs to
	Package string
	// Test is the Go test name
	Test string
	// Output contains any test output
	Output string
	// Parent indicates the parent test for subtests
	Parent string
	// Subtests contains any subtests
	Subtests []*TestResult
}

// TestError contains details about a test failure
type TestError struct {
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
}

// SourceLocation represents a location in source code
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

// TestPackage represents all tests from a single package
type TestPackage struct {
	// Package is the package name/path
	Package string
	// Tests is all tests in this package
	Tests []*TestResult
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

// FailedTestDetail represents detailed information about a failed test
type FailedTestDetail struct {
	// Result is the test result
	Result *TestResult
	// Suite is the test suite the test belongs to
	Suite *TestSuite
	// SourceCode contains the relevant source code
	SourceCode []string
	// ErrorLine is the line number where the error occurred
	ErrorLine int
	// FormattedError is the formatted error message
	FormattedError string
}
