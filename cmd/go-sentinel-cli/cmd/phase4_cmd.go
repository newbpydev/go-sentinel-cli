package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/newbpydev/go-sentinel/internal/cli"
	"github.com/spf13/cobra"
)

// phase4Cmd represents the phase4 command
var phase4Cmd = &cobra.Command{
	Use:   "phase4",
	Short: "Demonstrates Phase 4: Real-time Processing & Summary",
	Long: `Phase 4 focuses on real-time processing of test output and 
rendering a summary at the end. This matches the Vitest-like CLI 
format shown in the example screenshot.`,
	Run: func(cmd *cobra.Command, args []string) {
		demoPhase4()
	},
}

func init() {
	rootCmd.AddCommand(phase4Cmd)
}

func demoPhase4() {
	// Create renderers
	formatter := cli.NewColorFormatter(true)
	icons := cli.NewIconProvider(true)

	// Create test processor
	processor := cli.NewTestProcessor(os.Stdout, formatter, icons, 80)

	// Simulate test run with mock data
	stats := mockTestRun(processor)

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
}

func mockTestRun(processor *cli.TestProcessor) *cli.TestRunStats {
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
					File: "example_test.go",
					Line: 42,
				},
				SourceContext: []string{
					"40  func TestFailingExample(t *testing.T) {",
					"41      result := Calculate(5)",
					"42      if result != 5 {",
					"43          t.Errorf(\"Expected 5, got %d\", result)",
					"44      }",
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
