// Package models provides shared data models and value objects for the Go Sentinel CLI.
//
// This package contains core data structures used throughout the application for representing
// test results, error handling, file changes, and configuration. It follows the principle of
// providing clean value objects without business logic.
//
// Key components:
//   - Error handling: SentinelError with comprehensive error context and stack traces
//   - Test results: TestResult, PackageResult, and TestSummary for test execution data
//   - File changes: FileChange for representing file system modifications
//   - Configuration: TestConfiguration and WatchConfiguration for application settings
//
// Example usage:
//
//	// Creating and using test results
//	result := models.NewTestResult("TestExample", "github.com/example/pkg")
//	result.Status = models.TestStatusPassed
//	result.Duration = 100 * time.Millisecond
//
//	// Creating and handling errors
//	err := models.NewValidationError("config.timeout", "timeout must be positive")
//	if models.IsErrorType(err, models.ErrorTypeValidation) {
//		fmt.Println("Validation error:", err.UserMessage())
//	}
//
//	// Creating file change events
//	change := models.NewFileChange("main.go", models.ChangeTypeModified)
//	fmt.Printf("File %s was %s at %v\n", change.FilePath, change.ChangeType, change.Timestamp)
package models

import (
	"fmt"
	"time"
)

// Example_errorHandling demonstrates the comprehensive error handling system.
//
// This example shows how to create, wrap, and handle different types of errors
// with proper context and user-safe messaging.
func Example_errorHandling() {
	// Create a configuration error with user-safe message
	configErr := NewConfigError("invalid timeout value: must be greater than 0", true)
	fmt.Printf("Config error: %s\n", configErr.UserMessage())

	// Create a validation error for a specific field
	validationErr := NewValidationError("email", "email format is invalid")
	fmt.Printf("Validation error: %s\n", validationErr.UserMessage())

	// Create a file system error by wrapping an underlying error
	originalErr := fmt.Errorf("permission denied")
	fsErr := NewFileSystemError("read", "/etc/config.yaml", originalErr)
	fmt.Printf("File system error: %s\n", fsErr.Error())

	// Check error types
	if IsErrorType(configErr, ErrorTypeConfig) {
		fmt.Println("This is a configuration error")
	}

	// Get error context
	context := GetErrorContext(fsErr)
	if context != nil {
		fmt.Printf("Operation: %s, Resource: %s\n", context.Operation, context.Resource)
	}

	// Sanitize errors for user display
	sanitized := SanitizeError(fsErr)
	fmt.Printf("User-safe message: %s\n", sanitized.Error())

	// Output:
	// Config error: invalid timeout value: must be greater than 0
	// Validation error: email format is invalid
	// File system error: [FILESYSTEM:ERROR] (filesystem) file system operation failed: read resource=/etc/config.yaml operation=read
	// This is a configuration error
	// Operation: read, Resource: /etc/config.yaml
	// User-safe message: [FILESYSTEM:ERROR] File system error occurred
}

// Example_testResults demonstrates working with test execution results.
//
// This example shows how to create test results, manage package results,
// and generate comprehensive test summaries.
func Example_testResults() {
	// Create sample tests
	test := createExamplePassingTest()
	failingTest := createExampleFailingTest()

	// Create package and summary
	pkg := createExamplePackageResult(test, failingTest)
	summary := createExampleTestSummary(pkg)

	// Display results
	displayTestResults(pkg, summary, test, failingTest)

	// Output:
	// Package: github.com/example/auth
	// Success rate: 50.0%
	// Tests: 1 passed, 1 failed
	// Overall success rate: 50.0%
	// Test TestUserLogin passed
	// Test TestInvalidPassword failed: Expected authentication to fail, but got success
}

// createExamplePassingTest creates a sample passing test result.
func createExamplePassingTest() *TestResult {
	test := NewTestResult("TestUserLogin", "github.com/example/auth")
	test.Status = TestStatusPassed
	test.Duration = 150 * time.Millisecond
	test.Output = []string{"=== RUN   TestUserLogin", "--- PASS: TestUserLogin (0.15s)"}
	return test
}

// createExampleFailingTest creates a sample failing test result.
func createExampleFailingTest() *TestResult {
	failingTest := NewTestResult("TestInvalidPassword", "github.com/example/auth")
	failingTest.Status = TestStatusFailed
	failingTest.Duration = 50 * time.Millisecond
	failingTest.Error = &TestError{
		Message:    "Expected authentication to fail, but got success",
		Type:       "assertion",
		SourceFile: "auth_test.go",
		SourceLine: 42,
		Expected:   "false",
		Actual:     "true",
	}
	return failingTest
}

// createExamplePackageResult creates a sample package result with tests.
func createExamplePackageResult(test, failingTest *TestResult) *PackageResult {
	pkg := NewPackageResult("github.com/example/auth")
	pkg.AddTest(test)
	pkg.AddTest(failingTest)
	pkg.Duration = 200 * time.Millisecond
	return pkg
}

// createExampleTestSummary creates a sample test summary.
func createExampleTestSummary(pkg *PackageResult) *TestSummary {
	summary := NewTestSummary()
	summary.AddPackageResult(pkg)
	return summary
}

// displayTestResults displays test results and status information.
func displayTestResults(pkg *PackageResult, summary *TestSummary, test, failingTest *TestResult) {
	// Display package results
	fmt.Printf("Package: %s\n", pkg.Package)
	fmt.Printf("Success rate: %.1f%%\n", pkg.GetSuccessRate()*100)
	fmt.Printf("Tests: %d passed, %d failed\n", pkg.PassedCount, pkg.FailedCount)
	fmt.Printf("Overall success rate: %.1f%%\n", summary.GetSuccessRate()*100)

	// Check individual test status
	if test.IsSuccess() {
		fmt.Printf("Test %s passed\n", test.Name)
	}
	if failingTest.IsFailure() {
		fmt.Printf("Test %s failed: %s\n", failingTest.Name, failingTest.Error.Message)
	}
}

// Example_fileChanges demonstrates tracking file system changes.
//
// This example shows how to create and work with file change events
// for watch mode functionality.
func Example_fileChanges() {
	// Create different types of file changes
	created := NewFileChange("new_test.go", ChangeTypeCreated)
	created.Size = 1024
	created.Checksum = "abc123def456"

	modified := NewFileChange("existing_test.go", ChangeTypeModified)
	modified.Size = 2048
	modified.OldPath = "old_test.go" // For renamed files

	deleted := NewFileChange("obsolete_test.go", ChangeTypeDeleted)

	// Track changes over time
	changes := []*FileChange{created, modified, deleted}

	fmt.Printf("File change summary:\n")
	for _, change := range changes {
		fmt.Printf("- %s: %s (at %s)\n",
			change.ChangeType,
			change.FilePath,
			change.Timestamp.Format("15:04:05"))

		if change.Size > 0 {
			fmt.Printf("  Size: %d bytes\n", change.Size)
		}
		if change.OldPath != "" {
			fmt.Printf("  Previous path: %s\n", change.OldPath)
		}
	}

	// Output:
	// File change summary:
	// - created: new_test.go (at 15:04:05)
	//   Size: 1024 bytes
	// - modified: existing_test.go (at 15:04:05)
	//   Size: 2048 bytes
	//   Previous path: old_test.go
	// - deleted: obsolete_test.go (at 15:04:05)
}

// Example_configuration demonstrates creating and using configuration objects.
//
// This example shows how to set up test and watch configurations
// for different scenarios.
func Example_configuration() {
	// Create configurations
	testConfig := createExampleTestConfiguration()
	watchConfig := createExampleWatchConfiguration()

	// Display configurations
	displayConfigurations(testConfig, watchConfig)

	// Output:
	// Test Configuration:
	// - Packages: [./internal/... ./pkg/...]
	// - Coverage: true
	// - Parallel: 4
	// - Timeout: 5m0s
	//
	// Watch Configuration:
	// - Enabled: true
	// - Paths: [./internal ./pkg ./cmd]
	// - Debounce: 500ms
	// - Clear on rerun: true
}

// createExampleTestConfiguration creates a sample test configuration.
func createExampleTestConfiguration() *TestConfiguration {
	return &TestConfiguration{
		Packages:        []string{"./internal/...", "./pkg/..."},
		Verbose:         true,
		Coverage:        true,
		JSONOutput:      true,
		Parallel:        4,
		Timeout:         5 * time.Minute,
		Tags:            []string{"unit", "integration"},
		Environment:     map[string]string{"TEST_ENV": "development"},
		CoverageProfile: "coverage.out",
	}
}

// createExampleWatchConfiguration creates a sample watch configuration.
func createExampleWatchConfiguration() *WatchConfiguration {
	return &WatchConfiguration{
		Enabled:          true,
		Paths:            []string{"./internal", "./pkg", "./cmd"},
		IgnorePatterns:   []string{"*.tmp", "vendor/", ".git/"},
		TestPatterns:     []string{"*_test.go"},
		DebounceInterval: 500 * time.Millisecond,
		RunOnStart:       true,
		ClearOnRerun:     true,
		NotifyOnFailure:  true,
	}
}

// displayConfigurations displays test and watch configuration information.
func displayConfigurations(testConfig *TestConfiguration, watchConfig *WatchConfiguration) {
	// Display test configuration
	fmt.Printf("Test Configuration:\n")
	fmt.Printf("- Packages: %v\n", testConfig.Packages)
	fmt.Printf("- Coverage: %t\n", testConfig.Coverage)
	fmt.Printf("- Parallel: %d\n", testConfig.Parallel)
	fmt.Printf("- Timeout: %v\n", testConfig.Timeout)

	// Display watch configuration
	fmt.Printf("\nWatch Configuration:\n")
	fmt.Printf("- Enabled: %t\n", watchConfig.Enabled)
	fmt.Printf("- Paths: %v\n", watchConfig.Paths)
	fmt.Printf("- Debounce: %v\n", watchConfig.DebounceInterval)
	fmt.Printf("- Clear on rerun: %t\n", watchConfig.ClearOnRerun)
}

// Example_testStatus demonstrates working with test status values.
//
// This example shows the different test statuses and how to check them.
func Example_testStatus() {
	// Create tests with different statuses
	statuses := []TestStatus{
		TestStatusPending,
		TestStatusRunning,
		TestStatusPassed,
		TestStatusFailed,
		TestStatusSkipped,
		TestStatusTimeout,
		TestStatusError,
	}

	fmt.Printf("Test status examples:\n")
	for _, status := range statuses {
		// Create a test result with this status
		test := NewTestResult("ExampleTest", "example/pkg")
		test.Status = status

		// Check status conditions
		var condition string
		switch {
		case test.IsSuccess():
			condition = "success"
		case test.IsFailure():
			condition = "failure"
		case test.IsComplete():
			condition = "complete"
		default:
			condition = "in progress"
		}

		fmt.Printf("- %s: %s\n", status, condition)
	}

	// Output:
	// Test status examples:
	// - pending: in progress
	// - running: in progress
	// - passed: success
	// - failed: failure
	// - skipped: complete
	// - timeout: failure
	// - error: failure
}

// Example_coverage demonstrates working with test coverage data.
//
// This example shows how to create and analyze coverage information
// at different levels (test, file, function, package).
func Example_coverage() {
	// Create sample coverage data
	funcCoverage := createExampleFunctionCoverage()
	fileCoverage := createExampleFileCoverage()
	pkgCoverage := createExamplePackageCoverage(fileCoverage, funcCoverage)
	testCoverage := createExampleTestCoverage(fileCoverage)

	// Display coverage information
	displayCoverageInformation(pkgCoverage, fileCoverage, funcCoverage, testCoverage)

	// Output:
	// Package Coverage: github.com/example/handlers
	// - Overall: 82.3% (234/284 lines)
	// File Coverage: handler.go
	// - Coverage: 78.5% (89/113 statements)
	// Function Coverage: ProcessRequest
	// - Coverage: 85.7% (called 15 times)
	// - Lines: 10-25 in handler.go
	// ✓ Coverage meets threshold of 80.0%
}

// createExampleFunctionCoverage creates a sample function coverage object.
func createExampleFunctionCoverage() *FunctionCoverage {
	return &FunctionCoverage{
		Name:       "ProcessRequest",
		FilePath:   "handler.go",
		StartLine:  10,
		EndLine:    25,
		Percentage: 85.7,
		IsCovered:  true,
		CallCount:  15,
	}
}

// createExampleFileCoverage creates a sample file coverage object.
func createExampleFileCoverage() *FileCoverage {
	return &FileCoverage{
		FilePath:          "handler.go",
		Percentage:        78.5,
		CoveredLines:      157,
		TotalLines:        200,
		CoveredStatements: 89,
		TotalStatements:   113,
		LinesCovered:      []int{1, 2, 3, 5, 7, 8, 10, 11, 12},
		LinesUncovered:    []int{4, 6, 9, 13, 14},
	}
}

// createExamplePackageCoverage creates a sample package coverage object.
func createExamplePackageCoverage(fileCoverage *FileCoverage, funcCoverage *FunctionCoverage) *PackageCoverage {
	return &PackageCoverage{
		Package:           "github.com/example/handlers",
		Percentage:        82.3,
		CoveredLines:      234,
		TotalLines:        284,
		CoveredStatements: 156,
		TotalStatements:   189,
		Files: map[string]*FileCoverage{
			"handler.go": fileCoverage,
		},
		Functions: map[string]*FunctionCoverage{
			"ProcessRequest": funcCoverage,
		},
	}
}

// createExampleTestCoverage creates a sample test coverage object.
func createExampleTestCoverage(fileCoverage *FileCoverage) *TestCoverage {
	return &TestCoverage{
		Percentage:        82.3,
		CoveredLines:      234,
		TotalLines:        284,
		CoveredStatements: 156,
		TotalStatements:   189,
		Files: map[string]*FileCoverage{
			"handler.go": fileCoverage,
		},
	}
}

// displayCoverageInformation displays coverage information for different levels.
func displayCoverageInformation(pkgCoverage *PackageCoverage, fileCoverage *FileCoverage, funcCoverage *FunctionCoverage, testCoverage *TestCoverage) {
	// Display package coverage
	fmt.Printf("Package Coverage: %s\n", pkgCoverage.Package)
	fmt.Printf("- Overall: %.1f%% (%d/%d lines)\n",
		pkgCoverage.Percentage, pkgCoverage.CoveredLines, pkgCoverage.TotalLines)

	// Display file coverage
	fmt.Printf("File Coverage: %s\n", fileCoverage.FilePath)
	fmt.Printf("- Coverage: %.1f%% (%d/%d statements)\n",
		fileCoverage.Percentage, fileCoverage.CoveredStatements, fileCoverage.TotalStatements)

	// Display function coverage
	fmt.Printf("Function Coverage: %s\n", funcCoverage.Name)
	fmt.Printf("- Coverage: %.1f%% (called %d times)\n",
		funcCoverage.Percentage, funcCoverage.CallCount)
	fmt.Printf("- Lines: %d-%d in %s\n",
		funcCoverage.StartLine, funcCoverage.EndLine, funcCoverage.FilePath)

	// Check coverage thresholds
	threshold := 80.0
	if testCoverage.Percentage >= threshold {
		fmt.Printf("✓ Coverage meets threshold of %.1f%%\n", threshold)
	} else {
		fmt.Printf("✗ Coverage below threshold of %.1f%%\n", threshold)
	}
}
