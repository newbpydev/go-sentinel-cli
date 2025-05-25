package demo

import (
	"fmt"
	"time"

	"github.com/newbpydev/go-sentinel/pkg/models"
)

// createSampleTestResult creates a sample test result for demonstration
func createSampleTestResult() *models.LegacyTestResult {
	return &models.LegacyTestResult{
		Name:     "TestSampleFunction",
		Status:   models.StatusPassed,
		Duration: 50 * time.Millisecond,
		Package:  "github.com/newbpydev/go-sentinel/pkg/example",
		Test:     "TestSampleFunction",
		Output:   "PASS: TestSampleFunction (0.05s)",
	}
}

// createSampleTestSuite creates a sample test suite for demonstration
func createSampleTestSuite() *models.TestSuite {
	suite := &models.TestSuite{
		FilePath:     "github.com/newbpydev/go-sentinel/pkg/example/example_test.go",
		Duration:     100 * time.Millisecond,
		MemoryUsage:  10 * 1024 * 1024, // 10 MB
		TestCount:    3,
		PassedCount:  2,
		FailedCount:  1,
		SkippedCount: 0,
	}

	// Add the tests to the suite
	suite.Tests = append(suite.Tests, createSampleTestResult())

	return suite
}

// createMockTestSuites creates mock test suites for demonstration
func createMockTestSuites() []*models.TestSuite {
	var suites []*models.TestSuite

	// Create a simple test suite
	suite := &models.TestSuite{
		FilePath:     "test/example.test.go",
		TestCount:    5,
		PassedCount:  4,
		FailedCount:  1,
		SkippedCount: 0,
		Duration:     150 * time.Millisecond,
		MemoryUsage:  25 * 1024 * 1024, // 25 MB
	}

	// Create mock test results
	for i := 1; i <= 5; i++ {
		status := models.StatusPassed
		if i == 3 {
			status = models.StatusFailed
		}

		test := &models.LegacyTestResult{
			Name:     fmt.Sprintf("TestExample%d", i),
			Status:   status,
			Duration: 30 * time.Millisecond,
			Package:  "test",
			Test:     fmt.Sprintf("TestExample%d", i),
		}

		if status == models.StatusFailed {
			test.Error = &models.LegacyTestError{
				Message: "Expected true, got false",
				Type:    "AssertionError",
			}
		}

		suite.Tests = append(suite.Tests, test)
	}

	suites = append(suites, suite)
	return suites
}

// getMockFailedTests returns mock failed tests for demonstration
func getMockFailedTests(suites []*models.TestSuite) []*models.LegacyTestResult {
	var failedTests []*models.LegacyTestResult

	for _, suite := range suites {
		for _, test := range suite.Tests {
			if test.Status == models.StatusFailed {
				failedTests = append(failedTests, test)
			}
		}
	}

	return failedTests
}

// createMockFailedTestsWithSourceContext creates mock failed tests with source context
func createMockFailedTestsWithSourceContext() []*models.LegacyTestResult {
	return []*models.LegacyTestResult{
		{
			Name:     "TestFailedExample",
			Status:   models.StatusFailed,
			Duration: 25 * time.Millisecond,
			Package:  "test",
			Test:     "TestFailedExample",
			Error: &models.LegacyTestError{
				Message: "Expected 42, got 24",
				Type:    "AssertionError",
				Location: &models.SourceLocation{
					File:     "example_test.go",
					Line:     15,
					Function: "TestFailedExample",
				},
			},
		},
	}
}
