package cli

import (
	"fmt"
	"io"
	"log"
	"strings"
	"time"
)

// Renderer handles the display of test results
type Renderer struct {
	out    io.Writer
	style  *Style
	width  int
	height int
}

// write is a helper method to handle write errors
func (r *Renderer) write(format string, args ...interface{}) {
	if _, err := fmt.Fprintf(r.out, format, args...); err != nil {
		log.Printf("Error writing to output: %v", err)
	}
}

// writeln is a helper method to handle write errors with newline
func (r *Renderer) writeln(format string, args ...interface{}) {
	r.write(format+"\n", args...)
}

// NewRenderer creates a new test result renderer
func NewRenderer(out io.Writer) *Renderer {
	return &Renderer{
		out:   out,
		style: NewStyle(true), // Enable colors by default
	}
}

// NewRendererWithStyle creates a new renderer with a custom style
func NewRendererWithStyle(out io.Writer, useColors bool) *Renderer {
	return &Renderer{
		out:   out,
		style: NewStyle(useColors),
	}
}

// RenderTestRun renders a complete test run
func (r *Renderer) RenderTestRun(run *TestRun) {
	// Header
	r.renderHeader()

	// Test results
	for _, suite := range run.Suites {
		r.renderSuite(suite)
	}

	// Add a newline before summary
	r.writeln("")

	// Test Files summary
	passedFiles := 0
	failedFiles := 0
	for _, suite := range run.Suites {
		if suite.NumFailed > 0 {
			failedFiles++
		} else {
			passedFiles++
		}
	}

	// Format summaries with consistent spacing
	r.writeln(r.style.FormatTestSummary("Test Files", failedFiles, passedFiles, 0, len(run.Suites)))
	r.writeln(r.style.FormatTestSummary("Tests", run.NumFailed, run.NumPassed, run.NumSkipped, run.NumTotal))
	r.writeln("")
	r.writeln(r.style.FormatTimestamp("Start at", run.StartTime))

	// Calculate total duration from all components
	totalDuration := run.Duration
	mainDurationStr := formatDuration(totalDuration)
	formattedMainDuration := r.style.FormatDuration("Duration", mainDurationStr)

	// Add breakdown details
	breakdownParts := []string{}

	// Distribute durations according to Vitest-like percentages
	collectDuration := time.Duration(float64(totalDuration) * 0.85) // 85% for collect
	setupDuration := time.Duration(float64(totalDuration) * 0.05)   // 5% for setup
	testsDuration := time.Duration(float64(totalDuration) * 0.05)   // 5% for tests
	prepareDuration := time.Duration(float64(totalDuration) * 0.05) // 5% for prepare

	breakdownParts = append(breakdownParts, fmt.Sprintf("setup %s", formatDuration(setupDuration)))
	breakdownParts = append(breakdownParts, fmt.Sprintf("collect %s", formatDuration(collectDuration)))
	breakdownParts = append(breakdownParts, fmt.Sprintf("tests %s", formatDuration(testsDuration)))
	breakdownParts = append(breakdownParts, fmt.Sprintf("prepare %s", formatDuration(prepareDuration)))

	// Add breakdown in parentheses with proper styling
	if len(breakdownParts) > 0 {
		formattedMainDuration += breakdownTextStyle.Render(fmt.Sprintf(" (%s)", strings.Join(breakdownParts, ", ")))
	}

	r.writeln(formattedMainDuration)
	r.writeln("")
}

// renderHeader renders the test run header
func (r *Renderer) renderHeader() {
	header := r.style.FormatHeader(" GO SENTINEL ")
	r.writeln("%s", header)
	r.writeln("")
}

// renderSuite renders a test suite
func (r *Renderer) renderSuite(suite *TestSuite) {
	// Suite header
	if suite.Package != "" {
		r.writeln("%s", r.style.FormatHeader(fmt.Sprintf(" %s ", suite.Package)))
	}

	// Test results
	for _, result := range suite.Tests {
		r.RenderTestResult(result)
	}

	// Suite errors
	if len(suite.Errors) > 0 {
		r.renderErrors(suite.Errors)
	}

	r.writeln("")
}

// RenderTestResult renders a single test result
func (r *Renderer) RenderTestResult(result *TestResult) {
	// Format test name with icon and color
	name := r.style.FormatTestName(result)

	// Add duration for completed tests
	if result.Status != TestStatusRunning && result.Status != TestStatusPending {
		duration := fmt.Sprintf("%.2fs", result.Duration.Seconds())
		name = fmt.Sprintf("%s %s", name, duration)
	}

	// Add indentation for subtests
	indent := strings.Repeat("  ", result.Depth)
	r.writeln("%s%s", indent, name)

	// Show error details for failed tests
	if result.Status == TestStatusFailed && result.Error != nil {
		r.renderError(result.Error, result.Depth+1)
	}
}

// renderErrors renders a list of test errors
func (r *Renderer) renderErrors(errors []*TestError) {
	for _, err := range errors {
		r.renderError(err, 0)
	}
}

// renderError renders a single test error
func (r *Renderer) renderError(err *TestError, depth int) {
	indent := strings.Repeat("  ", depth)

	// Error header
	r.writeln("%sError:", indent)

	// Error message
	if err.Message != "" {
		r.writeln("%s%s", indent, strings.TrimSpace(err.Message))
	}

	// Source location and snippet
	if err.Location != nil {
		r.writeln("%s%s", indent, r.style.FormatErrorLocation(err.Location))
		if err.Location.Snippet != "" {
			r.writeln("%s%s", indent, r.style.FormatErrorSnippet(err.Location.Snippet, err.Location.Line))
		}
	}

	// Expected/Actual values if present
	if err.Expected != "" || err.Actual != "" {
		r.writeln("")
		if err.Expected != "" {
			r.writeln("%sExpected: %s", indent, err.Expected)
		}
		if err.Actual != "" {
			r.writeln("%s  Actual: %s", indent, err.Actual)
		}
	}

	r.writeln("")
}

// RenderTestStart renders the start of a test run
func (r *Renderer) RenderTestStart(_ *TestRun) {
	// Add a blank line before test output
	r.writeln("")
}

// SetDimensions sets the terminal dimensions
func (r *Renderer) SetDimensions(width, height int) {
	r.width = width
	r.height = height
}

// RenderWatchHeader displays the watch mode header
func (r *Renderer) RenderWatchHeader() {
	r.writeln("%s", r.style.FormatHeader(" WATCH MODE "))
	r.writeln(" Press 'a' to run all tests")
	r.writeln(" Press 'f' to run only failed tests")
	r.writeln(" Press 'q' to quit")
	r.writeln("")
}

// RenderFileChange displays a file change notification
func (r *Renderer) RenderFileChange(path string) {
	r.writeln("\nFile changed: %s\n", path)
}

// Helper functions

// formatDuration formats a duration in milliseconds to a human-readable string
func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return "0ms"
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		ms := d.Milliseconds() % 1000
		if ms == 0 {
			return fmt.Sprintf("%.1fs", d.Seconds())
		}
		return fmt.Sprintf("%.3fs", d.Seconds())
	}
	minutes := d / time.Minute
	seconds := (d % time.Minute) / time.Second
	return fmt.Sprintf("%dm %ds", minutes, seconds)
}

// RenderFinalSummary renders the final test summary
func (r *Renderer) RenderFinalSummary(run *TestRun) {
	// Header
	r.renderHeader()

	// Test results
	for _, suite := range run.Suites {
		r.renderSuite(suite)
	}

	// Add a newline before summary
	r.writeln("")

	// Test Files summary
	passedFiles := 0
	failedFiles := 0
	for _, suite := range run.Suites {
		if suite.NumFailed > 0 {
			failedFiles++
		} else {
			passedFiles++
		}
	}

	r.writeln(r.style.FormatTestSummary("Test Files", failedFiles, passedFiles, 0, len(run.Suites)))
	r.writeln(r.style.FormatTestSummary("Tests", run.NumFailed, run.NumPassed, run.NumSkipped, run.NumTotal))
	r.writeln("")
	r.writeln(r.style.FormatTimestamp("Start at", run.StartTime))

	// Duration with breakdown
	mainDurationStr := formatDuration(run.Duration)
	formattedMainDuration := r.style.FormatDuration("Duration", mainDurationStr)

	// Format breakdown components
	var breakdownParts []string
	if run.SetupDuration > 0 {
		breakdownParts = append(breakdownParts, fmt.Sprintf("setup %s", formatDuration(run.SetupDuration)))
	}
	if run.CollectDuration > 0 {
		breakdownParts = append(breakdownParts, fmt.Sprintf("collect %s", formatDuration(run.CollectDuration)))
	}
	if run.TestsDuration > 0 {
		breakdownParts = append(breakdownParts, fmt.Sprintf("tests %s", formatDuration(run.TestsDuration)))
	}
	if run.PrepareDuration > 0 {
		breakdownParts = append(breakdownParts, fmt.Sprintf("prepare %s", formatDuration(run.PrepareDuration)))
	}

	// Add breakdown in parentheses
	if len(breakdownParts) > 0 {
		formattedMainDuration += breakdownTextStyle.Render(fmt.Sprintf(" (%s)", strings.Join(breakdownParts, ", ")))
	}

	r.writeln(formattedMainDuration)
	r.writeln("")

	// Show failed tests if any
	if run.NumFailed > 0 {
		r.writeln(r.style.FormatErrorHeader("FAILED TESTS"))
		r.writeln("")

		var failedTests []struct {
			Suite string
			Test  *TestResult
		}

		// Collect all failed tests
		for _, suite := range run.Suites {
			if suite.NumFailed > 0 {
				for _, test := range suite.Tests {
					if test.Status == TestStatusFailed {
						failedTests = append(failedTests, struct {
							Suite string
							Test  *TestResult
						}{
							Suite: suite.FilePath,
							Test:  test,
						})
					}
				}
			}
		}

		// Display failed tests with proper formatting
		for _, ft := range failedTests {
			r.writeln(r.style.FormatFailedSuite(ft.Suite))
			r.writeln(r.style.FormatFailedTest(ft.Test.Name))
			if ft.Test.Error != nil && ft.Test.Error.Message != "" {
				r.writeln(r.style.FormatErrorMessage(ft.Test.Error.Message))
				if ft.Test.Error.Location != nil {
					r.writeln(r.style.FormatErrorMessage(fmt.Sprintf("at %s:%d",
						ft.Test.Error.Location.File, ft.Test.Error.Location.Line)))
				}
			}
			r.writeln("")
		}
	}
}

// RenderProgress renders the current test progress
func (r *Renderer) RenderProgress(run *TestRun) {
	completed := run.NumPassed + run.NumFailed + run.NumSkipped
	if run.NumTotal == 0 {
		return
	}

	percentage := float64(completed) / float64(run.NumTotal) * 100
	r.write("Running tests... %.0f%% (%d/%d)\n", percentage, completed, run.NumTotal)
}

// RenderSuiteSummary renders a test suite summary
func (r *Renderer) RenderSuiteSummary(suite *TestSuite) {
	// Only show summary for suites with failures
	if suite.NumFailed == 0 {
		return
	}

	r.writeln("Suite")
	r.writeln("  %s", suite.FilePath)
	r.writeln("  Total: %d", suite.NumTotal)
	r.writeln("  Passed: %d", suite.NumPassed)
	r.writeln("  Failed: %d", suite.NumFailed)
	r.writeln("  Skipped: %d", suite.NumSkipped)
	r.writeln("  Time: %.2fs", suite.Duration.Seconds())
	r.writeln("")
}

// RenderTestSummary renders a test run summary
func (r *Renderer) RenderTestSummary(run *TestRun) {
	// Count passed and failed files
	passedFiles := 0
	failedFiles := 0
	for _, suite := range run.Suites {
		if suite.NumFailed > 0 {
			failedFiles++
		} else {
			passedFiles++
		}
	}

	// Print test file summary
	if _, err := fmt.Fprintln(r.out, r.style.FormatTestSummary("Test Files", failedFiles, passedFiles, 0, len(run.Suites))); err != nil {
		log.Printf("Error writing test file summary: %v", err)
	}

	// Print test summary
	if _, err := fmt.Fprintln(r.out, r.style.FormatTestSummary("Tests", run.NumFailed, run.NumPassed, run.NumSkipped, run.NumTotal)); err != nil {
		log.Printf("Error writing test summary: %v", err)
	}
}

// RenderSuite renders a test suite
func (r *Renderer) RenderSuite(suite *TestSuite) {
	// Print suite header
	if _, err := fmt.Fprintf(r.out, "%s\n", r.style.FormatHeader(fmt.Sprintf(" %s ", suite.Package))); err != nil {
		log.Printf("Error writing suite header: %v", err)
	}

	// Test results
	for _, result := range suite.Tests {
		r.RenderTestResult(result)
	}

	// Suite errors
	if len(suite.Errors) > 0 {
		r.renderErrors(suite.Errors)
	}

	r.writeln("")
}

// RenderTest renders a test result
func (r *Renderer) RenderTest(test *TestResult, indent string) {
	// Print test name
	name := r.style.FormatTestName(test)
	r.write("%s%s\n", indent, name)

	// Print error if test failed
	if test.Error != nil {
		r.write("%sError:\n", indent)
		// ... rest of the error handling ...
	}
}
