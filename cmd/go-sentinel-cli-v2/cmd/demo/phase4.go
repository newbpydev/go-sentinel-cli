package demo

import (
	"fmt"
	"os"
	"time"

	"github.com/newbpydev/go-sentinel/internal/cli"
)

// RunPhase4Demo runs the Phase 4-D demonstration (Real-time Processing & Summary)
func RunPhase4Demo() {
	fmt.Println("=== Phase 4-D: Real-time Processing & Summary Demonstration ===")
	fmt.Println()

	// Create renderers
	formatter := cli.NewColorFormatter(true)
	icons := cli.NewIconProvider(true)

	// Create test processor
	processor := cli.NewTestProcessor(os.Stdout, formatter, icons, 80)

	// Simulate test run with mock data
	stats := mockPhase4TestRun(processor)

	// Render the results
	err := processor.RenderResults(true)
	if err != nil {
		fmt.Printf("Error rendering results: %v\n", err)
		return
	}

	// For demonstration, print the raw stats
	fmt.Println("\n\n--- Raw Statistics ---")
	fmt.Printf("Test Files: %d total, %d passed, %d failed\n",
		stats.TotalFiles, stats.PassedFiles, stats.FailedFiles)
	fmt.Printf("Tests: %d total, %d passed, %d failed, %d skipped\n",
		stats.TotalTests, stats.PassedTests, stats.FailedTests, stats.SkippedTests)
	fmt.Printf("Duration: %v\n", stats.Duration)

	// Add a summary and comparison
	fmt.Println("\nComparison with Vitest Summary Display:")
	fmt.Println("1. Clear test file statistics with pass/fail counts ✓")
	fmt.Println("2. Detailed test statistics showing passed/failed/skipped tests ✓")
	fmt.Println("3. Test start time displayed in HH:MM:SS format ✓")
	fmt.Println("4. Total duration with phase timing information ✓")
	fmt.Println("5. Colorized output matching Vitest style ✓")
}

// mockPhase4TestRun simulates a test run for Phase 4 demonstration
func mockPhase4TestRun(processor *cli.TestProcessor) *cli.TestRunStats {
	// Create mock test results
	results := []*cli.TestResult{
		{
			Name:     "TestExample",
			Package:  "github.com/test/example",
			Status:   cli.StatusPassed,
			Duration: 100 * time.Millisecond,
		},
		{
			Name:     "TestAnotherExample",
			Package:  "github.com/test/example",
			Status:   cli.StatusPassed,
			Duration: 50 * time.Millisecond,
		},
		{
			Name:     "TestFailingExample",
			Package:  "github.com/test/example",
			Status:   cli.StatusFailed,
			Duration: 75 * time.Millisecond,
			Error: &cli.TestError{
				Type:    "AssertionError",
				Message: "Expected 5, got 10",
				Location: &cli.SourceLocation{
					File:   "example_test.go",
					Line:   42,
					Column: 4,
				},
				SourceContext: []string{
					"func TestFailingExample(t *testing.T) {",
					"    result := Calculate(5)",
					"    if result != 5 {",
					"        t.Errorf(\"Expected 5, got %d\", result)",
					"    }",
				},
				HighlightedLine: 2,
			},
		},
		{
			Name:     "TestSkippedExample",
			Package:  "github.com/test/another",
			Status:   cli.StatusSkipped,
			Duration: 10 * time.Millisecond,
		},
	}

	// Create a test suite
	exampleSuite := &cli.TestSuite{
		FilePath:    "example_test.go",
		TestCount:   3,
		PassedCount: 2,
		FailedCount: 1,
		Tests:       results[:3],
	}

	anotherSuite := &cli.TestSuite{
		FilePath:     "another_test.go",
		TestCount:    1,
		SkippedCount: 1,
		Tests:        results[3:],
	}

	// Set processor state
	processor.AddTestSuite(exampleSuite)
	processor.AddTestSuite(anotherSuite)

	// Return stats
	return processor.GetStats()
}
