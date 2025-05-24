// Package display provides test result display formatting
package display

import (
	"fmt"
	"io"
	"strings"

	"github.com/newbpydev/go-sentinel/internal/ui/colors"
	"github.com/newbpydev/go-sentinel/pkg/models"
)

// TestDisplayInterface defines the interface for test result display
type TestDisplayInterface interface {
	RenderTestResult(result *models.LegacyTestResult, indentLevel int) error
	SetIndentLevel(level int)
	GetCurrentIndent() string
}

// TestRenderer renders individual test results and implements TestDisplayInterface
type TestRenderer struct {
	writer      io.Writer
	formatter   colors.FormatterInterface
	icons       colors.IconProviderInterface
	indentLevel int
}

// NewTestRenderer creates a new TestRenderer
func NewTestRenderer(writer io.Writer, formatter colors.FormatterInterface, icons colors.IconProviderInterface) *TestRenderer {
	return &TestRenderer{
		writer:      writer,
		formatter:   formatter,
		icons:       icons,
		indentLevel: 0,
	}
}

// NewTestRendererWithDefaults creates a TestRenderer with auto-detected color and icon support
func NewTestRendererWithDefaults(writer io.Writer) *TestRenderer {
	return &TestRenderer{
		writer:      writer,
		formatter:   colors.NewAutoColorFormatter(),
		icons:       colors.NewAutoIconProvider(),
		indentLevel: 0,
	}
}

// SetIndentLevel sets the base indentation level for all rendered tests
func (r *TestRenderer) SetIndentLevel(level int) {
	r.indentLevel = level
}

// GetCurrentIndent returns the current indentation string
func (r *TestRenderer) GetCurrentIndent() string {
	return strings.Repeat("  ", r.indentLevel)
}

// RenderTestResult renders a single test result with proper formatting
func (r *TestRenderer) RenderTestResult(result *models.LegacyTestResult, indentLevel int) error {
	if result == nil {
		return fmt.Errorf("test result cannot be nil")
	}

	// Create indentation (base + provided level)
	totalIndent := r.indentLevel + indentLevel
	indent := strings.Repeat("  ", totalIndent)

	// Format test name and status
	line := r.formatTestLine(result, indent)

	// Write the test line
	if _, err := fmt.Fprintln(r.writer, line); err != nil {
		return fmt.Errorf("failed to write test line: %w", err)
	}

	// Format error if test failed
	if result.Status == models.StatusFailed && result.Error != nil {
		errorIndent := indent + "  "
		errorLines := r.formatErrorLines(result.Error, errorIndent)

		for _, errLine := range errorLines {
			if _, err := fmt.Fprintln(r.writer, errLine); err != nil {
				return fmt.Errorf("failed to write error line: %w", err)
			}
		}
	}

	// Render subtests with increased indentation
	for _, subtest := range result.Subtests {
		if err := r.RenderTestResult(subtest, indentLevel+1); err != nil {
			return fmt.Errorf("failed to render subtest: %w", err)
		}
	}

	return nil
}

// formatTestLine formats a single test result line
func (r *TestRenderer) formatTestLine(result *models.LegacyTestResult, indent string) string {
	// Format icon based on status
	var icon string
	switch result.Status {
	case models.StatusPassed:
		icon = r.formatter.Green(r.icons.CheckMark())
	case models.StatusFailed:
		icon = r.formatter.Red(r.icons.Cross())
	case models.StatusSkipped:
		icon = r.formatter.Yellow(r.icons.Skipped())
	case models.StatusRunning:
		icon = r.formatter.Blue(r.icons.Running())
	default:
		icon = r.formatter.Gray("?")
	}

	// Format test name
	testName := r.formatTestName(result)

	// Format duration - matching Vitest's spacing
	duration := r.formatter.Dim(fmt.Sprintf("%dms", result.Duration.Milliseconds()))

	// Combine with Vitest-like spacing - note that in Vitest, the duration appears at the end with a space
	return fmt.Sprintf("%s%s %s %s", indent, icon, testName, duration)
}

// formatTestName formats the test name, handling subtests appropriately
func (r *TestRenderer) formatTestName(result *models.LegacyTestResult) string {
	if result.Parent != "" {
		// For subtests, only show the part after the parent name
		parts := strings.Split(result.Name, "/")
		if len(parts) > 1 {
			return parts[len(parts)-1]
		}
		return result.Name
	}
	return result.Name
}

// formatErrorLines formats error messages for failed tests
func (r *TestRenderer) formatErrorLines(err *models.LegacyTestError, indent string) []string {
	if err == nil {
		return []string{}
	}

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
			if strings.TrimSpace(stackLines[i]) != "" {
				lines = append(lines, fmt.Sprintf("%s  %s", indent, r.formatter.Dim(stackLines[i])))
			}
		}

		// If there are more lines, add an ellipsis
		if len(stackLines) > maxStackLines {
			lines = append(lines, fmt.Sprintf("%s  %s", indent, r.formatter.Dim("...")))
		}
	}

	return lines
}

// RenderTestResults renders multiple test results
func (r *TestRenderer) RenderTestResults(results []*models.LegacyTestResult) error {
	for _, result := range results {
		if err := r.RenderTestResult(result, 0); err != nil {
			return err
		}
	}
	return nil
}

// RenderTestSummaryLine renders a summary line for a test suite
func (r *TestRenderer) RenderTestSummaryLine(suite *models.TestSuite) error {
	if suite == nil {
		return fmt.Errorf("test suite cannot be nil")
	}

	indent := r.GetCurrentIndent()

	// Format suite icon and name
	var icon string
	if suite.FailedCount > 0 {
		icon = r.formatter.Red(r.icons.Cross())
	} else {
		icon = r.formatter.Green(r.icons.CheckMark())
	}

	// Format suite name (file path)
	suiteName := r.formatter.Bold(suite.FilePath)

	// Format counts and duration
	counts := fmt.Sprintf("(%d/%d passed)", suite.PassedCount, suite.TestCount)
	duration := r.formatter.Dim(fmt.Sprintf("%dms", suite.Duration.Milliseconds()))

	line := fmt.Sprintf("%s%s %s %s %s", indent, icon, suiteName, counts, duration)

	if _, err := fmt.Fprintln(r.writer, line); err != nil {
		return fmt.Errorf("failed to write suite summary: %w", err)
	}

	return nil
}
