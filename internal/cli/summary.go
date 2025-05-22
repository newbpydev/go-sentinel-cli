package cli

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// TestRunStats contains statistics about a test run
type TestRunStats struct {
	// Test file statistics
	TotalFiles  int
	PassedFiles int
	FailedFiles int

	// Test statistics
	TotalTests   int
	PassedTests  int
	FailedTests  int
	SkippedTests int

	// Timing
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	Phases    map[string]time.Duration
}

// SummaryRenderer renders a summary of test results
type SummaryRenderer struct {
	writer    io.Writer
	formatter *ColorFormatter
	icons     *IconProvider
	width     int
}

// NewSummaryRenderer creates a new SummaryRenderer
func NewSummaryRenderer(writer io.Writer, formatter *ColorFormatter, icons *IconProvider, width int) *SummaryRenderer {
	return &SummaryRenderer{
		writer:    writer,
		formatter: formatter,
		icons:     icons,
		width:     width,
	}
}

// RenderSummary renders a summary of test results
func (r *SummaryRenderer) RenderSummary(stats *TestRunStats) error {
	// Add separator line before summary
	fmt.Fprintln(r.writer, r.formatter.Dim(strings.Repeat("─", r.width)))

	// Display summary header
	fmt.Fprintln(r.writer, r.formatter.Bold("Test Summary:"))

	// Test files
	fileStats := fmt.Sprintf("Test Files: %s",
		r.formatter.Green(fmt.Sprintf("%d passed", stats.PassedFiles)))

	if stats.FailedFiles > 0 {
		fileStats += fmt.Sprintf(", %s",
			r.formatter.Red(fmt.Sprintf("%d failed", stats.FailedFiles)))
	}

	fileStats += fmt.Sprintf(" (total: %d)", stats.TotalFiles)
	fmt.Fprintln(r.writer, fileStats)

	// Tests
	testStats := fmt.Sprintf("Tests: %s",
		r.formatter.Green(fmt.Sprintf("%d passed", stats.PassedTests)))

	if stats.FailedTests > 0 {
		testStats += fmt.Sprintf(", %s",
			r.formatter.Red(fmt.Sprintf("%d failed", stats.FailedTests)))
	}

	if stats.SkippedTests > 0 {
		testStats += fmt.Sprintf(", %s",
			r.formatter.Yellow(fmt.Sprintf("%d skipped", stats.SkippedTests)))
	}

	testStats += fmt.Sprintf(" (total: %d)", stats.TotalTests)
	fmt.Fprintln(r.writer, testStats)

	// Start time
	fmt.Fprintf(r.writer, "Start at: %s\n",
		r.formatter.Gray(stats.StartTime.Format("15:04:05")))

	// Duration
	durationText := fmt.Sprintf("Duration: %s",
		r.formatter.Gray(formatDuration(stats.Duration)))

	// Add phase timing if available
	if len(stats.Phases) > 0 {
		durationText += " ("
		first := true

		for name, duration := range stats.Phases {
			if !first {
				durationText += ", "
			}
			durationText += fmt.Sprintf("%s: %s",
				name, r.formatter.Gray(formatDuration(duration)))
			first = false
		}

		durationText += ")"
	}

	fmt.Fprintln(r.writer, durationText)

	// Add another separator for clarity
	fmt.Fprintln(r.writer, r.formatter.Dim(strings.Repeat("─", r.width)))

	return nil
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	// Round to milliseconds for display
	ms := d.Milliseconds()

	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}

	sec := float64(ms) / 1000.0
	return fmt.Sprintf("%.2fs", sec)
}
