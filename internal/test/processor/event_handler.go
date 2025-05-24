package processor

import (
	"strings"
	"time"

	"github.com/newbpydev/go-sentinel/pkg/models"
)

// onTestRun handles a test run event
func (p *TestProcessor) onTestRun(event models.TestEvent) {
	// Track the first test start time (end of setup phase)
	if p.firstTestTime.IsZero() {
		p.firstTestTime = time.Now()
	}

	// Create a new test result
	result := &models.LegacyTestResult{
		Name:     event.Test,
		Package:  event.Package,
		Status:   models.StatusRunning,
		Duration: 0,
		Output:   "",
	}

	// Find or create the test suite first
	suitePath := event.Package
	suite, ok := p.suites[suitePath]
	if !ok {
		suite = &models.TestSuite{
			FilePath: suitePath,
		}
		p.suites[suitePath] = suite
	}

	// Check if this is a subtest
	if strings.Contains(event.Test, "/") {
		// Extract parent test name
		parts := strings.SplitN(event.Test, "/", 2)
		result.Parent = parts[0]

		// Add to parent test if it exists
		for _, test := range suite.Tests {
			if test.Name == result.Parent {
				test.Subtests = append(test.Subtests, result)
				// FIXED: Count subtests in suite.TestCount too
				suite.TestCount++
				return
			}
		}
	}

	// Add the test to the suite
	suite.Tests = append(suite.Tests, result)
	suite.TestCount++
}

// onTestPass handles a test pass event
func (p *TestProcessor) onTestPass(event models.TestEvent) {
	// Update last test completion time
	p.lastTestTime = time.Now()

	// Find the test
	for _, suite := range p.suites {
		for _, test := range suite.Tests {
			if test.Name == event.Test {
				test.Status = models.StatusPassed
				test.Duration = time.Duration(event.Elapsed * float64(time.Second))
				suite.PassedCount++
				p.statistics.PassedTests++
				p.statistics.TotalTests++
				return
			}

			// Check subtests
			for _, subtest := range test.Subtests {
				if subtest.Name == event.Test {
					subtest.Status = models.StatusPassed
					subtest.Duration = time.Duration(event.Elapsed * float64(time.Second))
					suite.PassedCount++
					p.statistics.PassedTests++
					p.statistics.TotalTests++
					return
				}
			}
		}
	}
}

// onTestFail handles a test fail event
func (p *TestProcessor) onTestFail(event models.TestEvent) {
	// Update last test completion time
	p.lastTestTime = time.Now()

	// Find the test
	for _, suite := range p.suites {
		for _, test := range suite.Tests {
			if test.Name == event.Test {
				test.Status = models.StatusFailed
				test.Duration = time.Duration(event.Elapsed * float64(time.Second))

				// Create error details from accumulated output
				test.Error = p.createTestError(test, event)

				suite.FailedCount++
				p.statistics.FailedTests++
				p.statistics.TotalTests++
				return
			}

			// Check subtests
			for _, subtest := range test.Subtests {
				if subtest.Name == event.Test {
					subtest.Status = models.StatusFailed
					subtest.Duration = time.Duration(event.Elapsed * float64(time.Second))

					// Create error details from accumulated output
					subtest.Error = p.createTestError(subtest, event)

					suite.FailedCount++
					p.statistics.FailedTests++
					p.statistics.TotalTests++
					return
				}
			}
		}
	}
}

// onTestSkip handles a test skip event
func (p *TestProcessor) onTestSkip(event models.TestEvent) {
	// Update last test completion time
	p.lastTestTime = time.Now()

	// Find the test
	for _, suite := range p.suites {
		for _, test := range suite.Tests {
			if test.Name == event.Test {
				test.Status = models.StatusSkipped
				test.Duration = time.Duration(event.Elapsed * float64(time.Second))
				suite.SkippedCount++
				p.statistics.SkippedTests++
				p.statistics.TotalTests++
				return
			}

			// Check subtests
			for _, subtest := range test.Subtests {
				if subtest.Name == event.Test {
					subtest.Status = models.StatusSkipped
					subtest.Duration = time.Duration(event.Elapsed * float64(time.Second))
					suite.SkippedCount++
					p.statistics.SkippedTests++
					p.statistics.TotalTests++
					return
				}
			}
		}
	}
}

// onTestOutput handles test output
func (p *TestProcessor) onTestOutput(event models.TestEvent) {
	// Find the test
	for _, suite := range p.suites {
		for _, test := range suite.Tests {
			if test.Name == event.Test {
				test.Output += event.Output
				return
			}

			// Check subtests
			for _, subtest := range test.Subtests {
				if subtest.Name == event.Test {
					subtest.Output += event.Output
					return
				}
			}
		}
	}
}

// createTestError creates a TestError from test result and event (placeholder)
// This will be moved to error_processor.go
func (p *TestProcessor) createTestError(test *models.LegacyTestResult, event models.TestEvent) *models.LegacyTestError {
	return &models.LegacyTestError{
		Message: "Test failed",
		Type:    "failure",
	}
}
