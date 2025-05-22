package cli

import (
	"fmt"
	"io"
)

// SuiteRenderer renders test suites with collapsing/expanding
type SuiteRenderer struct {
	writer    io.Writer
	formatter *ColorFormatter
	icons     *IconProvider
	width     int
	header    *HeaderRenderer
	test      *TestRenderer
}

// NewSuiteRenderer creates a new SuiteRenderer
func NewSuiteRenderer(writer io.Writer, formatter *ColorFormatter, icons *IconProvider, width int) *SuiteRenderer {
	return &SuiteRenderer{
		writer:    writer,
		formatter: formatter,
		icons:     icons,
		width:     width,
		header:    NewHeaderRenderer(writer, formatter, icons, width),
		test:      NewTestRenderer(writer, formatter, icons),
	}
}

// RenderSuite renders a test suite to the output writer
func (r *SuiteRenderer) RenderSuite(suite *TestSuite, autoCollapse bool) error {
	// Determine if we should collapse this suite
	shouldCollapse := autoCollapse && suite.FailedCount == 0

	// Format and write the suite header
	header := r.formatSuiteHeader(suite)
	_, err := fmt.Fprintln(r.writer, header)
	if err != nil {
		return err
	}

	// If collapsed, just show a summary line
	if shouldCollapse {
		statusIcon := r.formatter.Green(r.icons.CheckMark())
		_, err = fmt.Fprintf(r.writer, "  %s %s\n", statusIcon, r.formatter.Green(fmt.Sprintf("Suite passed (%d tests)", suite.TestCount)))
		return err
	}

	// If not collapsed, render each test
	testRenderer := NewTestRenderer(r.writer, r.formatter, r.icons)
	for _, test := range suite.Tests {
		err := testRenderer.RenderTestResult(test, 1)
		if err != nil {
			return err
		}
	}

	return nil
}

// formatSuiteHeader creates a header line for a test suite
func (r *SuiteRenderer) formatSuiteHeader(suite *TestSuite) string {
	// Get file path
	formattedPath := FormatFilePath(r.formatter, suite.FilePath)

	// Format test counts
	var testCountStr string
	if suite.FailedCount > 0 {
		// Red if any tests failed
		testCountStr = fmt.Sprintf("(%d tests | %s)",
			suite.TestCount,
			r.formatter.Red(fmt.Sprintf("%d failed", suite.FailedCount)),
		)
	} else if suite.SkippedCount > 0 && suite.SkippedCount == suite.TestCount {
		// Yellow if all tests skipped
		testCountStr = fmt.Sprintf("(%d tests | %s)",
			suite.TestCount,
			r.formatter.Yellow(fmt.Sprintf("%d skipped", suite.SkippedCount)),
		)
	} else {
		// Just green count if all passed
		skippedPart := ""
		if suite.SkippedCount > 0 {
			skippedPart = " | " + r.formatter.Yellow(fmt.Sprintf("%d skipped", suite.SkippedCount))
		}
		testCountStr = fmt.Sprintf("(%d tests%s)",
			suite.TestCount,
			skippedPart,
		)
	}

	// Format duration and memory
	durationStr := r.formatter.Dim(fmt.Sprintf("%dms", suite.Duration.Milliseconds()))
	memoryStr := r.formatter.Dim(fmt.Sprintf("%d MB heap used", suite.MemoryUsage/(1024*1024)))

	// Combine all parts
	return fmt.Sprintf("%s %s %s %s",
		formattedPath,
		testCountStr,
		durationStr,
		memoryStr,
	)
}
