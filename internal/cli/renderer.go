package cli

import (
	"fmt"
	"io"
	"log"
	"strings"
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

	// Render the summary
	r.renderSummary(run)
}

// renderSummary is the single source of truth for rendering test summaries
func (r *Renderer) renderSummary(run *TestRun) {
	// Calculate file statistics
	passedFiles := 0
	failedFiles := 0
	for _, suite := range run.Suites {
		if suite.NumFailed > 0 {
			failedFiles++
		} else if suite.NumTotal > 0 {
			passedFiles++
		}
	}

	// Add a divider line before summary
	r.writeln("")

	// Format summaries with consistent spacing
	r.writeln(r.style.FormatTestSummary("Test Files", failedFiles, passedFiles, 0, len(run.Suites)))
	r.writeln(r.style.FormatTestSummary("Tests", run.NumFailed, run.NumPassed, run.NumSkipped, run.NumTotal))

	r.writeln("")
	r.writeln(r.style.FormatTimestamp("Start at", run.StartTime))
	if !run.EndTime.IsZero() {
		r.writeln(r.style.FormatTimestamp("End at", run.EndTime))
	}

	// Calculate total duration from all components
	totalDuration := run.Duration
	mainDurationStr := FormatDurationAdaptive(totalDuration)
	formattedMainDuration := r.style.FormatDuration("Duration", mainDurationStr)

	// Add breakdown details
	breakdownParts := []string{}

	// Setup duration
	if run.SetupDuration > 0 {
		breakdownParts = append(breakdownParts, fmt.Sprintf("setup %s", FormatDurationAdaptive(run.SetupDuration)))
	}

	// Collect duration
	if run.CollectDuration > 0 {
		breakdownParts = append(breakdownParts, fmt.Sprintf("collect %s", FormatDurationAdaptive(run.CollectDuration)))
	}

	// Tests duration
	if run.TestsDuration > 0 {
		breakdownParts = append(breakdownParts, fmt.Sprintf("tests %s", FormatDurationAdaptive(run.TestsDuration)))
	}

	// Parse duration
	if run.ParseDuration > 0 {
		breakdownParts = append(breakdownParts, fmt.Sprintf("parse %s", FormatDurationAdaptive(run.ParseDuration)))
	}

	// Add breakdown in parentheses with proper styling
	if len(breakdownParts) > 0 {
		formattedMainDuration += " " + r.style.FormatBreakdownText(fmt.Sprintf("(%s)", strings.Join(breakdownParts, ", ")))
	}

	r.writeln(formattedMainDuration)
	r.writeln("")

	// Show failed tests if any
	if run.NumFailed > 0 {
		r.writeln(r.style.FormatErrorHeader(" FAILED Tests "))
		r.writeln("")

		// Show failed test files/suites
		for _, suite := range run.Suites {
			if suite.NumFailed > 0 {
				r.writeln(r.style.FormatFailedSuite(suite.FilePath))

				// Only show the failing tests, not all tests in the suite
				for _, test := range suite.Tests {
					if test.Status == TestStatusFailed {
						// Simple test name only
						testName := test.Name
						// If it's a subtest, show just the subtest part
						if strings.Contains(testName, "/") {
							parts := strings.Split(testName, "/")
							testName = parts[len(parts)-1]
						}

						r.writeln("    %s", r.style.FormatFailedTest(testName))

						// Show error details compactly
						if test.Error != nil {
							if test.Error.Message != "" {
								msg := strings.TrimSpace(test.Error.Message)
								// Extract just the first line of the error message
								if idx := strings.Index(msg, "\n"); idx > 0 {
									msg = msg[:idx]
								}
								r.writeln("    %s", r.style.FormatErrorMessage(msg))
							}
							if test.Error.Location != nil {
								r.writeln("    %s", r.style.FormatErrorLocation(test.Error.Location))
							}
						}

						r.writeln("")
					}
				}
			}
		}
	}
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
		duration := FormatDurationPrecise(result.Duration)
		// Pad name and duration for better alignment, matching the example in the image
		// Use 30 chars for the name like in the example
		name = fmt.Sprintf("%-30s %s", name, duration)
	}

	// Add indentation for subtests
	indent := strings.Repeat("  ", result.Depth)
	r.writeln("%s%s", indent, name)

	// Show error details for failed tests
	if result.Status == TestStatusFailed && result.Error != nil {
		// If we have a source location, show it on a new line
		if result.Error.Location != nil {
			r.writeln("%s  at %s", indent, r.style.FormatErrorLocation(result.Error.Location))
		}
		// Only show error message details if there's a message
		if result.Error.Message != "" && strings.TrimSpace(result.Error.Message) != "" {
			r.renderError(result.Error, result.Depth+1)
		}
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

	// Error message
	if err.Message != "" {
		r.writeln("%s%s", indent, r.style.FormatErrorMessage(strings.TrimSpace(err.Message)))
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
			r.writeln("%sExpected: %s", indent, r.style.FormatErrorValue(err.Expected))
		}
		if err.Actual != "" {
			r.writeln("%s  Actual: %s", indent, r.style.FormatErrorValue(err.Actual))
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

// RenderFinalSummary renders the final test summary
func (r *Renderer) RenderFinalSummary(run *TestRun) {
	// Use the consolidated summary rendering
	r.renderSummary(run)
}

// RenderTestSummary is deprecated and should not be used
func (r *Renderer) RenderTestSummary(run *TestRun) {
	// This function is deprecated and should not be used
	// Use RenderFinalSummary instead
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
	r.writeln("  Time: %s", FormatDurationPrecise(suite.Duration))
	r.writeln("")
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
