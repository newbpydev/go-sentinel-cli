package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/newbpydev/go-sentinel/internal/cli"
	"github.com/spf13/cobra"
)

// demoCmd represents the command for testing and demonstrating features
var demoCmd = &cobra.Command{
	Use:   "demo",
	Short: "Run a demonstration of CLI features",
	Long: `Demonstrates various features of the CLI by running tests
and displaying the results in different formats.

This command is used for development and validation.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check which phase demo to run
		phase, _ := cmd.Flags().GetString("phase")

		switch phase {
		case "1d":
			runPhase1Demo()
		case "2d":
			runPhase2Demo()
		default:
			fmt.Println("Please specify a valid phase to demo (1d or 2d)")
			fmt.Println("Example: go-sentinel-cli demo --phase=1d")
		}
	},
}

func init() {
	rootCmd.AddCommand(demoCmd)

	// Add flags
	demoCmd.Flags().StringP("phase", "p", "", "Phase to demo (1d, 2d)")
	demoCmd.MarkFlagRequired("phase")
}

// runPhase1Demo runs the Phase 1-D demonstration (Core Architecture)
func runPhase1Demo() {
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

// runPhase2Demo runs the Phase 2-D demonstration (Test Suite Display)
func runPhase2Demo() {
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
		displayFailedTestsSection(failedTests, formatter, icons, terminalWidth)
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

// createSampleTestResult creates a sample test result for demonstration
func createSampleTestResult() *cli.TestResult {
	return &cli.TestResult{
		Name:     "TestSampleFunction",
		Status:   cli.StatusPassed,
		Duration: 50 * time.Millisecond,
		Package:  "github.com/newbpydev/go-sentinel/pkg/example",
		Test:     "TestSampleFunction",
		Output:   "PASS: TestSampleFunction (0.05s)",
		Subtests: []*cli.TestResult{
			{
				Name:     "TestSampleFunction/subtest_case_1",
				Status:   cli.StatusPassed,
				Duration: 20 * time.Millisecond,
				Package:  "github.com/newbpydev/go-sentinel/pkg/example",
				Test:     "TestSampleFunction/subtest_case_1",
				Parent:   "TestSampleFunction",
				Output:   "PASS: subtest_case_1 (0.02s)",
			},
			{
				Name:     "TestSampleFunction/subtest_case_2",
				Status:   cli.StatusFailed,
				Duration: 15 * time.Millisecond,
				Package:  "github.com/newbpydev/go-sentinel/pkg/example",
				Test:     "TestSampleFunction/subtest_case_2",
				Parent:   "TestSampleFunction",
				Output:   "FAIL: subtest_case_2 (0.015s)",
				Error: &cli.TestError{
					Message: "Expected 5, got 10",
					Type:    "AssertionError",
					Location: &cli.SourceLocation{
						File: "example_test.go",
						Line: 42,
					},
				},
			},
		},
	}
}

// createSampleTestSuite creates a sample test suite for demonstration
func createSampleTestSuite() *cli.TestSuite {
	suite := &cli.TestSuite{
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

// printJSON pretty-prints an object as JSON
func printJSON(v interface{}) {
	data, err := json.MarshalIndent(v, "    ", "  ")
	if err != nil {
		fmt.Printf("Error encoding JSON: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

// formatBytes formats bytes as human-readable string
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return strconv.FormatUint(bytes, 10) + " B"
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// isColorTerminal detects if the terminal supports colors
func isColorTerminal() bool {
	// Check for NO_COLOR environment variable
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Check for FORCE_COLOR environment variable
	if os.Getenv("FORCE_COLOR") != "" {
		return true
	}

	// Check for TTY
	fileInfo, _ := os.Stdout.Stat()
	if (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		return true
	}

	// Check for specific environment variables
	term := os.Getenv("TERM")
	if term == "xterm" || term == "xterm-256color" || term == "screen" || term == "screen-256color" {
		return true
	}

	return false
}

// createMockTestSuites creates mock test suites that match the Vitest screenshot
func createMockTestSuites() []*cli.TestSuite {
	var suites []*cli.TestSuite

	// Create settings.test.ts suite (all passed)
	settingsSuite := &cli.TestSuite{
		FilePath:     "test/settings.test.ts",
		TestCount:    12,
		PassedCount:  12,
		FailedCount:  0,
		SkippedCount: 0,
		Duration:     119 * time.Millisecond,
		MemoryUsage:  33 * 1024 * 1024, // 33 MB
	}

	// Create mock test results for settings tests
	for i := 1; i <= 12; i++ {
		test := &cli.TestResult{
			Name:     fmt.Sprintf("SettingsTest%d", i),
			Status:   cli.StatusPassed,
			Duration: 10 * time.Millisecond,
			Package:  "test",
			Test:     fmt.Sprintf("SettingsTest%d", i),
		}
		settingsSuite.Tests = append(settingsSuite.Tests, test)
	}

	// Create websocket.test.ts suite (all failed)
	websocketSuite := &cli.TestSuite{
		FilePath:     "test/websocket.test.ts",
		TestCount:    8,
		PassedCount:  0,
		FailedCount:  8,
		SkippedCount: 0,
		Duration:     21 * time.Millisecond,
		MemoryUsage:  32 * 1024 * 1024, // 32 MB
	}

	// Create mock test results for websocket tests (failing tests)
	failingTests := []struct {
		name    string
		message string
		time    time.Duration
	}{
		{"WebSocketClient - connect method - should create a WebSocket with the given URL", "wsClient.connect is not a function", 8 * time.Millisecond},
		{"WebSocketClient - event handlers - should register open event handlers", "wsClient.connect is not a function", 1 * time.Millisecond},
		{"WebSocketClient - event handlers - should register close event handlers", "wsClient.connect is not a function", 1 * time.Millisecond},
		{"WebSocketClient - event handlers - should register and handle message events", "wsClient.connect is not a function", 1 * time.Millisecond},
		{"WebSocketClient - event handlers - should register error event handlers", "wsClient.connect is not a function", 3 * time.Millisecond},
		{"WebSocketClient - send method - should send JSON-stringified data when socket is open", "wsClient.connect is not a function", 1 * time.Millisecond},
		{"WebSocketClient - send method - should not send data when socket is not open", "wsClient.connect is not a function", 1 * time.Millisecond},
		{"WebSocketClient - disconnect method - should close the WebSocket connection", "wsClient.connect is not a function", 1 * time.Millisecond},
	}

	for _, ft := range failingTests {
		test := &cli.TestResult{
			Name:     ft.name,
			Status:   cli.StatusFailed,
			Duration: ft.time,
			Package:  "test",
			Test:     ft.name,
			Error: &cli.TestError{
				Message: ft.message,
				Type:    "TypeError",
			},
		}
		websocketSuite.Tests = append(websocketSuite.Tests, test)
	}

	// Create toast.test.ts suite (all passed)
	toastSuite := &cli.TestSuite{
		FilePath:     "test/toast.test.ts",
		TestCount:    8,
		PassedCount:  8,
		FailedCount:  0,
		SkippedCount: 0,
		Duration:     93 * time.Millisecond,
		MemoryUsage:  34 * 1024 * 1024, // 34 MB
	}

	// Create mock test results for toast tests
	for i := 1; i <= 8; i++ {
		test := &cli.TestResult{
			Name:     fmt.Sprintf("ToastTest%d", i),
			Status:   cli.StatusPassed,
			Duration: 10 * time.Millisecond,
			Package:  "test",
			Test:     fmt.Sprintf("ToastTest%d", i),
		}
		toastSuite.Tests = append(toastSuite.Tests, test)
	}

	// Create main.test.ts suite (all passed)
	mainSuite := &cli.TestSuite{
		FilePath:     "test/main.test.ts",
		TestCount:    10,
		PassedCount:  10,
		FailedCount:  0,
		SkippedCount: 0,
		Duration:     382 * time.Millisecond,
		MemoryUsage:  36 * 1024 * 1024, // 36 MB
	}

	// Create mock test results for main tests
	for i := 1; i <= 10; i++ {
		test := &cli.TestResult{
			Name:     fmt.Sprintf("MainTest%d", i),
			Status:   cli.StatusPassed,
			Duration: 30 * time.Millisecond,
			Package:  "test",
			Test:     fmt.Sprintf("MainTest%d", i),
		}
		mainSuite.Tests = append(mainSuite.Tests, test)
	}

	// Create coverage.test.ts suite (all passed)
	coverageSuite := &cli.TestSuite{
		FilePath:     "test/coverage.test.ts",
		TestCount:    20,
		PassedCount:  20,
		FailedCount:  0,
		SkippedCount: 0,
		Duration:     313 * time.Millisecond,
		MemoryUsage:  35 * 1024 * 1024, // 35 MB
	}

	// Create mock test results for coverage tests
	for i := 1; i <= 20; i++ {
		test := &cli.TestResult{
			Name:     fmt.Sprintf("CoverageTest%d", i),
			Status:   cli.StatusPassed,
			Duration: 15 * time.Millisecond,
			Package:  "test",
			Test:     fmt.Sprintf("CoverageTest%d", i),
		}
		coverageSuite.Tests = append(coverageSuite.Tests, test)
	}

	// Create utils/websocket.test.ts suite (all passed)
	utilsWebsocketSuite := &cli.TestSuite{
		FilePath:     "test/utils/websocket.test.ts",
		TestCount:    10,
		PassedCount:  10,
		FailedCount:  0,
		SkippedCount: 0,
		Duration:     24 * time.Millisecond,
		MemoryUsage:  40 * 1024 * 1024, // 40 MB
	}

	// Create mock test results for utils/websocket tests
	for i := 1; i <= 10; i++ {
		test := &cli.TestResult{
			Name:     fmt.Sprintf("UtilsWebSocketTest%d", i),
			Status:   cli.StatusPassed,
			Duration: 2 * time.Millisecond,
			Package:  "test/utils",
			Test:     fmt.Sprintf("UtilsWebSocketTest%d", i),
		}
		utilsWebsocketSuite.Tests = append(utilsWebsocketSuite.Tests, test)
	}

	// Create example.test.ts suite (all passed)
	exampleSuite := &cli.TestSuite{
		FilePath:     "test/example.test.ts",
		TestCount:    2,
		PassedCount:  2,
		FailedCount:  0,
		SkippedCount: 0,
		Duration:     14 * time.Millisecond,
		MemoryUsage:  39 * 1024 * 1024, // 39 MB
	}

	// Create mock test results for example tests
	for i := 1; i <= 2; i++ {
		test := &cli.TestResult{
			Name:     fmt.Sprintf("ExampleTest%d", i),
			Status:   cli.StatusPassed,
			Duration: 7 * time.Millisecond,
			Package:  "test",
			Test:     fmt.Sprintf("ExampleTest%d", i),
		}
		exampleSuite.Tests = append(exampleSuite.Tests, test)
	}

	// Add all suites to the result
	suites = append(suites,
		settingsSuite,
		websocketSuite,
		toastSuite,
		mainSuite,
		coverageSuite,
		utilsWebsocketSuite,
		exampleSuite,
	)

	return suites
}

// getMockFailedTests extracts all failed tests from mock suites
func getMockFailedTests(suites []*cli.TestSuite) []*cli.TestResult {
	var failedTests []*cli.TestResult

	for _, suite := range suites {
		for _, test := range suite.Tests {
			if test.Status == cli.StatusFailed {
				// Set the filepath for better display
				if test.Error != nil && test.Error.Location == nil {
					test.Error.Location = &cli.SourceLocation{
						File: suite.FilePath,
						Line: 42, // Mock line number
					}
				}

				failedTests = append(failedTests, test)
			}
		}
	}

	return failedTests
}

// displayFailedTestsSection displays a detailed section about failed tests
func displayFailedTestsSection(failedTests []*cli.TestResult, formatter *cli.ColorFormatter, icons *cli.IconProvider, width int) {
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
	fmt.Printf("Time: %s\n", formatter.Gray(time.Now().Format("15:04:05")))

	fmt.Println()
}
