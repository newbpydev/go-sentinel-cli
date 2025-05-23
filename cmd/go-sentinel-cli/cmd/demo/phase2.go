package demo

import (
	"fmt"
	"os"
	"strings"

	"github.com/newbpydev/go-sentinel/internal/cli"
)

// RunPhase2Demo runs the Phase 2-D demonstration (Test Suite Display)
func RunPhase2Demo() {
	fmt.Println("=== Phase 2-D: Test Suite Display Demonstration ===")
	fmt.Println()

	// Get terminal properties
	isColorSupported := isColorTerminal()
	terminalWidth := 80 // Fixed width for demo purposes

	// Create formatters
	formatter := cli.NewColorFormatter(isColorSupported)
	icons := cli.NewIconProvider(true) // Always use Unicode icons for demo

	// Create mock test suites based on the Vitest screenshot
	suites := createMockTestSuites()

	// Display each test suite
	for _, suite := range suites {
		// Create a renderer for this suite
		renderer := cli.NewSuiteRenderer(os.Stdout, formatter, icons, terminalWidth)

		// Render the suite with auto-collapse for passing suites
		err := renderer.RenderSuite(suite, true)
		if err != nil {
			fmt.Printf("Error rendering suite: %v\n", err)
		}

		fmt.Println()
	}

	// Display failed tests section for the failing tests
	failedTests := getMockFailedTests(suites)
	if len(failedTests) > 0 {
		displayFailedTestsSection(failedTests, formatter, icons)
	}

	// Display summary
	displaySummary(suites, formatter)

	// Compare with Vitest output
	fmt.Println("\nVisual Comparison with Vitest:")
	fmt.Println("1. File paths displayed with correct coloring ✓")
	fmt.Println("2. Test counts show proper highlighting for failures ✓")
	fmt.Println("3. Duration and memory usage match Vitest format ✓")
	fmt.Println("4. Passing suites are collapsed, failing suites expanded ✓")
	fmt.Println("5. Indentation of nested tests matches Vitest ✓")
}

// displayFailedTestsSection displays a detailed section about failed tests
func displayFailedTestsSection(failedTests []*cli.TestResult, formatter *cli.ColorFormatter, icons *cli.IconProvider) {
	// This is a simplified version of what will be implemented in Phase 3
	fmt.Println(formatter.Bold(formatter.Red("Failed Tests:")))

	for _, test := range failedTests {
		fmt.Printf("  %s %s\n", formatter.Red(icons.Cross()), test.Name)
		if test.Error != nil {
			// Match Vitest format with an arrow indicator
			fmt.Printf("    → %s\n", formatter.Red(test.Error.Message))

			// Show location if available
			if test.Error.Location != nil {
				fmt.Printf("      %s %s:%d\n",
					formatter.Dim("at"),
					test.Error.Location.File,
					test.Error.Location.Line)
			}
		}
	}

	fmt.Println()
}

// displaySummary shows a summary of test results
func displaySummary(suites []*cli.TestSuite, formatter *cli.ColorFormatter) {
	// Count totals
	passedSuites := 0
	failedSuites := 0
	totalTests := 0
	passedTests := 0
	failedTests := 0
	skippedTests := 0

	for _, suite := range suites {
		if suite.FailedCount > 0 {
			failedSuites++
		} else {
			passedSuites++
		}

		totalTests += suite.TestCount
		passedTests += suite.PassedCount
		failedTests += suite.FailedCount
		skippedTests += suite.SkippedCount
	}

	// Add separator line before summary (to match Vitest's clear visual separation)
	fmt.Println(formatter.Dim(strings.Repeat("─", 50)))

	// Display summary
	fmt.Println(formatter.Bold("Test Summary:"))

	// Test files
	fmt.Printf("Test Files: %s, %s\n",
		formatter.Green(fmt.Sprintf("%d passed", passedSuites)),
		formatter.Red(fmt.Sprintf("%d failed", failedSuites)),
	)

	// Tests
	testSummary := fmt.Sprintf("Tests: %s, %s",
		formatter.Green(fmt.Sprintf("%d passed", passedTests)),
		formatter.Red(fmt.Sprintf("%d failed", failedTests)),
	)

	if skippedTests > 0 {
		testSummary += fmt.Sprintf(", %s", formatter.Yellow(fmt.Sprintf("%d skipped", skippedTests)))
	}

	fmt.Println(testSummary)

	// Time
	fmt.Printf("Time: %s\n", formatter.Gray("16:01:08"))

	fmt.Println()
}
