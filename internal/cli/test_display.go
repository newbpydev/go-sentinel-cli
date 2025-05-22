package cli

import (
	"fmt"
	"io"
	"strings"
)

// TestRenderer renders individual test results
type TestRenderer struct {
	writer    io.Writer
	formatter *ColorFormatter
	icons     *IconProvider
}

// NewTestRenderer creates a new TestRenderer
func NewTestRenderer(writer io.Writer, formatter *ColorFormatter, icons *IconProvider) *TestRenderer {
	return &TestRenderer{
		writer:    writer,
		formatter: formatter,
		icons:     icons,
	}
}

// RenderTestResult renders a single test result with proper formatting
func (r *TestRenderer) RenderTestResult(result *TestResult, indentLevel int) error {
	// Create indentation
	indent := strings.Repeat("  ", indentLevel)

	// Format test name and status
	line := r.formatTestLine(result, indent)

	// Write the test line
	if _, err := fmt.Fprintln(r.writer, line); err != nil {
		return err
	}

	// Format error if test failed
	if result.Status == StatusFailed && result.Error != nil {
		errorIndent := indent + "  "
		errorLines := r.formatErrorLines(result.Error, errorIndent)

		for _, errLine := range errorLines {
			if _, err := fmt.Fprintln(r.writer, errLine); err != nil {
				return err
			}
		}
	}

	// Render subtests with increased indentation
	for _, subtest := range result.Subtests {
		if err := r.RenderTestResult(subtest, indentLevel+1); err != nil {
			return err
		}
	}

	return nil
}

// formatTestLine formats a single test result line
func (r *TestRenderer) formatTestLine(result *TestResult, indent string) string {
	// Format icon based on status
	var icon string
	switch result.Status {
	case StatusPassed:
		icon = r.formatter.Green(r.icons.CheckMark())
	case StatusFailed:
		icon = r.formatter.Red(r.icons.Cross())
	case StatusSkipped:
		icon = r.formatter.Yellow(r.icons.Skipped())
	case StatusRunning:
		icon = r.formatter.Blue(r.icons.Running())
	default:
		icon = "?"
	}

	// Format test name
	var testName string
	if result.Parent != "" {
		// For subtests, only show the part after the parent name
		parts := strings.Split(result.Name, "/")
		if len(parts) > 1 {
			testName = parts[len(parts)-1]
		} else {
			testName = result.Name
		}
	} else {
		testName = result.Name
	}

	// Format duration - matching Vitest's spacing
	duration := r.formatter.Dim(fmt.Sprintf("%dms", result.Duration.Milliseconds()))

	// Combine with Vitest-like spacing - note that in Vitest, the duration appears at the end with a space
	return fmt.Sprintf("%s%s %s %s", indent, icon, testName, duration)
}

// formatErrorLines formats error messages for failed tests
func (r *TestRenderer) formatErrorLines(err *TestError, indent string) []string {
	lines := []string{}

	// Format error type and message - using a consistent spacing pattern to match Vitest
	// Vitest format: "→ wsClient.connect is not a function"
	errorMessage := fmt.Sprintf("%s→ %s",
		indent,
		r.formatter.Red(err.Message),
	)
	lines = append(lines, errorMessage)

	// Add expected/actual for assertion errors
	if err.Type == "AssertionError" && err.Expected != "" && err.Actual != "" {
		expected := fmt.Sprintf("%s  %s %s",
			indent,
			r.formatter.Dim("Expected:"),
			r.formatter.Green(err.Expected),
		)
		lines = append(lines, expected)

		actual := fmt.Sprintf("%s  %s %s",
			indent,
			r.formatter.Dim("Received:"),
			r.formatter.Red(err.Actual),
		)
		lines = append(lines, actual)
	}

	// Add location if available
	if err.Location != nil {
		location := fmt.Sprintf("%s%s %s:%d",
			indent,
			r.formatter.Dim("at"),
			err.Location.File,
			err.Location.Line,
		)
		lines = append(lines, location)
	}

	// Add abbreviated stack trace for panics
	if err.Type == "Panic" && err.Stack != "" {
		lines = append(lines, fmt.Sprintf("%s%s", indent, r.formatter.Dim("Stack trace:")))

		// Add first few lines of stack trace
		stackLines := strings.Split(err.Stack, "\n")
		maxStackLines := 3 // Limit to 3 lines for brevity

		for i := 0; i < len(stackLines) && i < maxStackLines; i++ {
			if stackLines[i] != "" {
				lines = append(lines, fmt.Sprintf("%s%s", indent, r.formatter.Dim(stackLines[i])))
			}
		}

		// If there are more lines, add an ellipsis
		if len(stackLines) > maxStackLines {
			lines = append(lines, fmt.Sprintf("%s%s", indent, r.formatter.Dim("...")))
		}
	}

	return lines
}
