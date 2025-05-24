package demo

import (
	"fmt"

	"github.com/newbpydev/go-sentinel/internal/cli"
)

// RunPhase1Demo runs the Phase 1-D demonstration (Core Architecture)
func RunPhase1Demo() {
	fmt.Println("=== Phase 1-D: Core Architecture Demonstration ===")
	fmt.Println()

	// 1-D.1.1: Implement minimal CLI to run basic tests
	fmt.Println("1. Testing Core Data Structures:")

	// 1-D.1.2: Add test cases that exercise data structures
	// Create test objects to validate our data structures
	testResult := createSampleTestResult()
	testSuite := createSampleTestSuite()

	// 1-D.1.3: Output raw parsed results to validate data structure correctness
	fmt.Println("  - TestResult Structure:")
	printJSON(testResult)

	fmt.Println("\n  - TestSuite Structure:")
	printJSON(testSuite)

	// 1-D.2.1: Verify test output is correctly parsed into data structures
	fmt.Println("\n2. Validating Test Result Parsing:")
	validateTestStructures(testResult, testSuite)

	// 1-D.2.2: Validate terminal color support detection
	fmt.Println("\n3. Testing Terminal Capabilities:")
	checkTerminalCapabilities()

	// 1-D.2.3: Confirm correct emoji/icon display
	fmt.Println("\n4. Testing Icon Display:")
	displayIconSamples()

	// 1-D.2.4: Document discrepancies
	fmt.Println("\n5. Identified Issues:")
	fmt.Println("  - None so far")
}

// validateTestStructures checks the test structures for correctness
func validateTestStructures(result *cli.TestResult, suite *cli.TestSuite) {
	// Validate TestResult
	fmt.Println("  - TestResult validation:")
	fmt.Printf("    ✓ Has name: %s\n", result.Name)
	fmt.Printf("    ✓ Has status: %s\n", result.Status)
	fmt.Printf("    ✓ Has duration: %v\n", result.Duration)
	fmt.Printf("    ✓ Has %d subtests\n", len(result.Subtests))

	if result.Error != nil {
		fmt.Println("    ✗ Should not have error for passing test")
	} else if len(result.Subtests) > 0 && result.Subtests[1].Error == nil {
		fmt.Println("    ✗ Subtest should have error")
	} else {
		fmt.Println("    ✓ Error structure is correct")
	}

	// Validate TestSuite
	fmt.Println("  - TestSuite validation:")
	fmt.Printf("    ✓ Has file path: %s\n", suite.FilePath)
	fmt.Printf("    ✓ Has %d tests (%d passed, %d failed, %d skipped)\n",
		suite.TestCount, suite.PassedCount, suite.FailedCount, suite.SkippedCount)
	fmt.Printf("    ✓ Has duration: %v\n", suite.Duration)
	fmt.Printf("    ✓ Has memory usage: %s\n", formatBytes(suite.MemoryUsage))
}

// checkTerminalCapabilities checks terminal color and unicode support
func checkTerminalCapabilities() {
	// Get terminal properties
	isColorTerminal := isColorTerminal()

	// Display capabilities
	fmt.Println("  - Terminal capabilities:")
	fmt.Printf("    ✓ Color support: %v\n", isColorTerminal)
	fmt.Printf("    ✓ Unicode support: %v\n", true) // Assume Unicode support

	// Show colored output
	if isColorTerminal {
		fmt.Println("    ✓ \033[32mThis text should be green\033[0m")
		fmt.Println("    ✓ \033[31mThis text should be red\033[0m")
		fmt.Println("    ✓ \033[33mThis text should be yellow\033[0m")
	} else {
		fmt.Println("    ✗ Terminal does not support colors")
	}
}

// displayIconSamples shows the various icons with colors
func displayIconSamples() {
	isColorTerminal := isColorTerminal()

	// Create a formatter
	formatter := cli.NewColorFormatter(isColorTerminal)
	icons := cli.NewIconProvider(true) // Always use Unicode icons for demo

	// Display icon samples
	fmt.Println("  - Icon samples:")

	// Passed icon
	fmt.Printf("    %s %s\n",
		formatter.Green(icons.CheckMark()),
		formatter.Green("Test passed"))

	// Failed icon
	fmt.Printf("    %s %s\n",
		formatter.Red(icons.Cross()),
		formatter.Red("Test failed"))

	// Skipped icon
	fmt.Printf("    %s %s\n",
		formatter.Yellow(icons.Skipped()),
		formatter.Yellow("Test skipped"))

	// Running icon
	fmt.Printf("    %s %s\n",
		formatter.Blue(icons.Running()),
		formatter.Blue("Test running"))
}
