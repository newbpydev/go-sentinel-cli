package models

import (
	"testing"
	"time"
)

// Test 1.1.1: Define LegacyTestResult structure with required fields for Vitest-like display
func TestLegacyTestResultStructure(t *testing.T) {
	// Create a test result with all required fields
	result := LegacyTestResult{
		Name:     "TestExample",
		Status:   TestStatusPassed,
		Duration: 50 * time.Millisecond,
		Package:  "github.com/user/project/pkg",
		Test:     "TestExample",
		Output:   "Test output",
	}

	// Validate fields
	if result.Name != "TestExample" {
		t.Errorf("Expected Name to be 'TestExample', got '%s'", result.Name)
	}

	if result.Status != TestStatusPassed {
		t.Errorf("Expected Status to be '%s', got '%s'", TestStatusPassed, result.Status)
	}

	if result.Duration != 50*time.Millisecond {
		t.Errorf("Expected Duration to be 50ms, got '%v'", result.Duration)
	}

	if result.Package != "github.com/user/project/pkg" {
		t.Errorf("Expected Package to be 'github.com/user/project/pkg', got '%s'", result.Package)
	}

	// Check that subtests can be added
	subtest := &LegacyTestResult{
		Name:     "TestExample/SubTest",
		Status:   TestStatusPassed,
		Duration: 20 * time.Millisecond,
		Package:  "github.com/user/project/pkg",
		Test:     "TestExample/SubTest",
		Parent:   "TestExample",
		Output:   "Subtest output",
	}
	result.Subtests = append(result.Subtests, subtest)

	if len(result.Subtests) != 1 {
		t.Errorf("Expected 1 subtest, got %d", len(result.Subtests))
	}

	if result.Subtests[0].Name != "TestExample/SubTest" {
		t.Errorf("Expected subtest Name to be 'TestExample/SubTest', got '%s'", result.Subtests[0].Name)
	}

	if result.Subtests[0].Parent != "TestExample" {
		t.Errorf("Expected subtest Parent to be 'TestExample', got '%s'", result.Subtests[0].Parent)
	}
}

// Test 1.1.2: Define TestSuite structure for organizing tests by file
func TestTestSuiteStructure(t *testing.T) {
	// Create a test suite with all required fields
	suite := TestSuite{
		FilePath:     "pkg/example_test.go",
		Duration:     150 * time.Millisecond,
		MemoryUsage:  1024 * 1024, // 1MB
		TestCount:    5,
		PassedCount:  3,
		FailedCount:  1,
		SkippedCount: 1,
	}

	// Validate fields
	if suite.FilePath != "pkg/example_test.go" {
		t.Errorf("Expected FilePath to be 'pkg/example_test.go', got '%s'", suite.FilePath)
	}

	if suite.Duration != 150*time.Millisecond {
		t.Errorf("Expected Duration to be 150ms, got '%v'", suite.Duration)
	}

	if suite.MemoryUsage != 1024*1024 {
		t.Errorf("Expected MemoryUsage to be 1MB, got '%d'", suite.MemoryUsage)
	}

	if suite.TestCount != 5 {
		t.Errorf("Expected TestCount to be 5, got '%d'", suite.TestCount)
	}

	if suite.PassedCount != 3 {
		t.Errorf("Expected PassedCount to be 3, got '%d'", suite.PassedCount)
	}

	if suite.FailedCount != 1 {
		t.Errorf("Expected FailedCount to be 1, got '%d'", suite.FailedCount)
	}

	if suite.SkippedCount != 1 {
		t.Errorf("Expected SkippedCount to be 1, got '%d'", suite.SkippedCount)
	}

	// Add test results to the suite
	test1 := &LegacyTestResult{Name: "Test1", Status: TestStatusPassed, Duration: 50 * time.Millisecond}
	test2 := &LegacyTestResult{Name: "Test2", Status: TestStatusFailed, Duration: 100 * time.Millisecond}
	suite.Tests = append(suite.Tests, test1, test2)

	if len(suite.Tests) != 2 {
		t.Errorf("Expected 2 tests, got %d", len(suite.Tests))
	}

	if suite.Tests[0].Name != "Test1" || suite.Tests[0].Status != TestStatusPassed {
		t.Errorf("First test not as expected")
	}

	if suite.Tests[1].Name != "Test2" || suite.Tests[1].Status != TestStatusFailed {
		t.Errorf("Second test not as expected")
	}
}

// Test 1.1.3: Define FailedTestDetail structure for detailed error reporting
func TestFailedTestDetailStructure(t *testing.T) {
	// Create error location
	location := &SourceLocation{
		File:        "pkg/example.go",
		Line:        42,
		Column:      15,
		Function:    "ExampleFunc",
		Context:     []string{"line 40", "line 41", "error on line 42", "line 43", "line 44"},
		ContextLine: 2, // Index of the error line in Context array
	}

	// Create test error
	testError := &LegacyTestError{
		Message:  "Expected 5, got 10",
		Type:     "AssertionError",
		Stack:    "stack trace...",
		Expected: "5",
		Actual:   "10",
		Location: location,
	}

	// Create failed test result
	result := &LegacyTestResult{
		Name:     "FailingTest",
		Status:   TestStatusFailed,
		Duration: 75 * time.Millisecond,
		Error:    testError,
	}

	// Create test suite
	suite := &TestSuite{
		FilePath:    "pkg/example_test.go",
		Tests:       []*LegacyTestResult{result},
		Duration:    75 * time.Millisecond,
		TestCount:   1,
		FailedCount: 1,
	}

	// Create failed test detail
	detail := FailedTestDetail{
		Result:         result,
		Suite:          suite,
		SourceCode:     []string{"line 40", "line 41", "error on line 42", "line 43", "line 44"},
		ErrorLine:      42,
		FormattedError: "AssertionError: Expected 5, got 10",
	}

	// Validate fields
	if detail.Result.Name != "FailingTest" {
		t.Errorf("Expected Result.Name to be 'FailingTest', got '%s'", detail.Result.Name)
	}

	if detail.Result.Status != TestStatusFailed {
		t.Errorf("Expected Result.Status to be 'failed', got '%s'", detail.Result.Status)
	}

	if detail.Suite.FilePath != "pkg/example_test.go" {
		t.Errorf("Expected Suite.FilePath to be 'pkg/example_test.go', got '%s'", detail.Suite.FilePath)
	}

	if detail.ErrorLine != 42 {
		t.Errorf("Expected ErrorLine to be 42, got '%d'", detail.ErrorLine)
	}

	if detail.FormattedError != "AssertionError: Expected 5, got 10" {
		t.Errorf("Expected FormattedError to be 'AssertionError: Expected 5, got 10', got '%s'", detail.FormattedError)
	}

	if len(detail.SourceCode) != 5 {
		t.Errorf("Expected 5 lines of source code, got %d", len(detail.SourceCode))
	}

	if detail.SourceCode[2] != "error on line 42" {
		t.Errorf("Expected error line to be 'error on line 42', got '%s'", detail.SourceCode[2])
	}
}

// TestNewTestResult tests the new TestResult constructor
func TestNewTestResult(t *testing.T) {
	result := NewTestResult("TestExample", "github.com/user/project")

	if result.Name != "TestExample" {
		t.Errorf("Expected Name to be 'TestExample', got '%s'", result.Name)
	}

	if result.Package != "github.com/user/project" {
		t.Errorf("Expected Package to be 'github.com/user/project', got '%s'", result.Package)
	}

	if result.Status != TestStatusPending {
		t.Errorf("Expected Status to be 'pending', got '%s'", result.Status)
	}

	if result.ID == "" {
		t.Error("Expected ID to be generated, got empty string")
	}

	if result.Output == nil {
		t.Error("Expected Output to be initialized, got nil")
	}

	if result.Subtests == nil {
		t.Error("Expected Subtests to be initialized, got nil")
	}

	if result.Metadata == nil {
		t.Error("Expected Metadata to be initialized, got nil")
	}
}

// TestTestResult_Methods tests TestResult methods
func TestTestResult_Methods(t *testing.T) {
	result := NewTestResult("TestExample", "github.com/user/project")

	// Test IsSuccess
	result.Status = TestStatusPassed
	if !result.IsSuccess() {
		t.Error("Expected IsSuccess to return true for passed test")
	}

	// Test IsFailure
	result.Status = TestStatusFailed
	if !result.IsFailure() {
		t.Error("Expected IsFailure to return true for failed test")
	}

	// Test IsComplete
	if !result.IsComplete() {
		t.Error("Expected IsComplete to return true for failed test")
	}

	result.Status = TestStatusRunning
	if result.IsComplete() {
		t.Error("Expected IsComplete to return false for running test")
	}

	// Test AddSubtest
	subtest := NewTestResult("SubTest", "github.com/user/project")
	result.AddSubtest(subtest)

	if len(result.Subtests) != 1 {
		t.Errorf("Expected 1 subtest, got %d", len(result.Subtests))
	}

	if result.Subtests[0].Parent != "TestExample" {
		t.Errorf("Expected subtest parent to be 'TestExample', got '%s'", result.Subtests[0].Parent)
	}
}
