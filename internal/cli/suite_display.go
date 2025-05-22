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

// RenderSuite renders a test suite with appropriate collapsing/expanding
func (r *SuiteRenderer) RenderSuite(suite *TestSuite, autoCollapse bool) error {
	// Always render the header
	if err := r.header.RenderSuiteHeader(suite); err != nil {
		return err
	}

	// Determine if suite should be collapsed
	collapsed := autoCollapse && suite.FailedCount == 0 && suite.TestCount > 0

	// If collapsed, show summary line
	if collapsed {
		return r.renderCollapsedSummary(suite)
	}

	// Otherwise show expanded tests
	return r.renderExpandedTests(suite)
}

// renderCollapsedSummary renders a summary line for collapsed suites
func (r *SuiteRenderer) renderCollapsedSummary(suite *TestSuite) error {
	var summary string

	if suite.SkippedCount == suite.TestCount {
		// All tests skipped
		summary = fmt.Sprintf("  %s All tests skipped (%d %s)",
			r.formatter.Yellow(r.icons.Skipped()),
			suite.SkippedCount,
			pluralize("test", suite.SkippedCount),
		)
	} else {
		// Passing tests
		summary = fmt.Sprintf("  %s Suite %s (%d %s)",
			r.formatter.Green(r.icons.CheckMark()),
			r.formatter.Green("passed"),
			suite.TestCount,
			pluralize("test", suite.TestCount),
		)
	}

	_, err := fmt.Fprintln(r.writer, summary)
	return err
}

// renderExpandedTests renders all tests in a suite
func (r *SuiteRenderer) renderExpandedTests(suite *TestSuite) error {
	// For empty suites, show special message
	if suite.TestCount == 0 {
		emptyMsg := fmt.Sprintf("  %s No tests found",
			r.formatter.Yellow("•"),
		)
		_, err := fmt.Fprintln(r.writer, emptyMsg)
		return err
	}

	// Render each test (only top-level tests, subtests are handled by TestRenderer)
	for _, test := range suite.Tests {
		if err := r.test.RenderTestResult(test, 1); err != nil {
			return err
		}
	}

	return nil
}
