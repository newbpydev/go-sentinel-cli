// Package models provides core data structures for test execution and result management.
//
// This file contains fundamental data models used throughout the Go Sentinel CLI
// for representing test results, error contexts, and package-level information.
// These models form the foundation of the test execution and reporting system.
//
// Key components:
//   - SourceLocation: Precise location information for code references
//   - TestPackage: Package-level test aggregation and statistics
//   - FailedTestDetail: Comprehensive information about test failures
//   - LegacyTestResult: Backward compatibility during migration
//   - LegacyTestError: Error representation for legacy compatibility
//
// Design principles:
//   - Immutable data structures where possible
//   - Rich metadata for debugging and reporting
//   - Backward compatibility during migration phase
//   - Clear separation between data and behavior
//
// Example usage:
//
//	// Creating a test package result
//	pkg := &TestPackage{
//		Package:     "github.com/example/auth",
//		Duration:    150 * time.Millisecond,
//		TestCount:   5,
//		PassedCount: 4,
//		FailedCount: 1,
//		Passed:      false,
//	}
//
//	// Creating source location information
//	location := &SourceLocation{
//		File:        "auth.go",
//		Line:        42,
//		Column:      15,
//		Function:    "ValidatePassword",
//		Context:     []string{"func ValidatePassword(pass string) bool {", "    return len(pass) > 8", "}"},
//		ContextLine: 1,
//	}
package models

import (
	"time"
)

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

// FailedTestDetail represents detailed information about a failed test
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

// LegacyTestResult for backward compatibility during migration
// This matches the structure from internal/cli/models.go
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

// LegacyTestError for backward compatibility during migration
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
