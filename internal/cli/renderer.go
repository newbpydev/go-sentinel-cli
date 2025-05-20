package cli

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff"))

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00ff00"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff0000"))

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffff00"))

	dimStyle = lipgloss.NewStyle().
			Faint(true)

	// Test status styles
	passedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00ff00")).
			SetString("✓")

	failedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff0000")).
			SetString("✕")

	skippedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffff00")).
			SetString("○")

	runningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0000ff")).
			SetString("⠋")
)

// Renderer handles the display of test results
type Renderer struct {
	out    io.Writer
	style  *Style
	width  int
	height int
}

// NewRenderer creates a new renderer instance
func NewRenderer(out io.Writer) *Renderer {
	return &Renderer{
		out:   out,
		style: NewStyle(),
		width: 80, // Default width
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

	// Summary
	r.renderSummary(run)
}

// renderHeader renders the test run header
func (r *Renderer) renderHeader() {
	header := r.style.FormatHeader(" GO SENTINEL ")
	fmt.Fprintln(r.out, header)
	fmt.Fprintln(r.out)
}

// renderSuite renders a test suite
func (r *Renderer) renderSuite(suite *TestSuite) {
	// Suite header
	if suite.Package != "" {
		fmt.Fprintf(r.out, "%s\n", r.style.FormatHeader(fmt.Sprintf(" %s ", suite.Package)))
	}

	// Test results
	for _, result := range suite.Tests {
		r.RenderTestResult(result)
	}

	// Suite errors
	if len(suite.Errors) > 0 {
		r.renderErrors(suite.Errors)
	}

	fmt.Fprintln(r.out)
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
	fmt.Fprintf(r.out, "%s%s\n", indent, name)

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
	fmt.Fprintf(r.out, "%sError:\n", indent)

	// Error message
	if err.Message != "" {
		fmt.Fprintf(r.out, "%s%s\n", indent, strings.TrimSpace(err.Message))
	}

	// Source location and snippet
	if err.Location != nil {
		fmt.Fprintf(r.out, "%s%s\n", indent, r.style.FormatErrorLocation(err.Location))
		if err.Location.Snippet != "" {
			fmt.Fprintf(r.out, "%s%s\n", indent, r.style.FormatErrorSnippet(err.Location.Snippet, err.Location.Line))
		}
	}

	// Expected/Actual values if present
	if err.Expected != "" || err.Actual != "" {
		fmt.Fprintln(r.out)
		if err.Expected != "" {
			fmt.Fprintf(r.out, "%sExpected: %s\n", indent, err.Expected)
		}
		if err.Actual != "" {
			fmt.Fprintf(r.out, "%s  Actual: %s\n", indent, err.Actual)
		}
	}

	fmt.Fprintln(r.out)
}

// renderSummary renders the final test run summary
func (r *Renderer) renderSummary(run *TestRun) {
	// Separator
	fmt.Fprintln(r.out, lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Render(strings.Repeat("─", r.width)))

	// Summary line
	fmt.Fprintln(r.out, r.style.FormatSummary(run))
	fmt.Fprintln(r.out)
}

// RenderTestStart renders the initial test run message
func (r *Renderer) RenderTestStart(run *TestRun) {
	fmt.Fprintln(r.out, r.style.FormatHeader(" RUNNING TESTS "))
	fmt.Fprintln(r.out)
}

// SetDimensions sets the terminal dimensions
func (r *Renderer) SetDimensions(width, height int) {
	r.width = width
	r.height = height
}

// RenderWatchHeader displays the watch mode header
func (r *Renderer) RenderWatchHeader() {
	fmt.Fprintln(r.out, titleStyle.Render("Watch Mode"))
	fmt.Fprintln(r.out, " Press 'a' to run all tests")
	fmt.Fprintln(r.out, " Press 'f' to run only failed tests")
	fmt.Fprintln(r.out, " Press 'q' to quit")
	fmt.Fprintln(r.out)
}

// RenderFileChange displays a file change notification
func (r *Renderer) RenderFileChange(path string) {
	fmt.Fprintf(r.out, "\nFile changed: %s\n\n", path)
}

// Helper functions

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

// RenderFinalSummary renders the final test run summary
func (r *Renderer) RenderFinalSummary(run *TestRun) {
	fmt.Fprintln(r.out, titleStyle.Render("Test Run"))
	fmt.Fprintln(r.out)

	// Summary line
	fmt.Fprintf(r.out, "Total: %d Passed: %d Failed: %d Skipped: %d Time: %.2fs\n",
		run.NumTotal, run.NumPassed, run.NumFailed, run.NumSkipped, run.Duration.Seconds())

	// Failed tests section
	if run.NumFailed > 0 {
		fmt.Fprintln(r.out)
		fmt.Fprintln(r.out, errorStyle.Render("Failed Tests:"))
		for _, suite := range run.Suites {
			for _, test := range suite.Tests {
				if test.Status == TestStatusFailed {
					fmt.Fprintf(r.out, "  %s (%s)\n", test.Name, suite.FilePath)
				}
			}
		}
	}

	// Duration
	fmt.Fprintf(r.out, "\nTotal Duration: %.2fs\n", run.Duration.Seconds())
}

// RenderProgress renders the current test progress
func (r *Renderer) RenderProgress(run *TestRun) {
	completed := run.NumPassed + run.NumFailed + run.NumSkipped
	if run.NumTotal == 0 {
		return
	}

	percentage := float64(completed) / float64(run.NumTotal) * 100
	fmt.Fprintf(r.out, "Running tests... %.0f%% (%d/%d)\n", percentage, completed, run.NumTotal)
}

// RenderSuiteSummary renders a test suite summary
func (r *Renderer) RenderSuiteSummary(suite *TestSuite) {
	// Only show summary for suites with failures
	if suite.NumFailed == 0 {
		return
	}

	fmt.Fprintln(r.out, titleStyle.Render("Suite"))
	fmt.Fprintf(r.out, "  %s\n", suite.FilePath)
	fmt.Fprintf(r.out, "  Total: %d\n", suite.NumTotal)
	fmt.Fprintf(r.out, "  Passed: %d\n", suite.NumPassed)
	fmt.Fprintf(r.out, "  Failed: %d\n", suite.NumFailed)
	fmt.Fprintf(r.out, "  Skipped: %d\n", suite.NumSkipped)
	fmt.Fprintf(r.out, "  Time: %.2fs\n", suite.Duration.Seconds())
	fmt.Fprintln(r.out)
}
