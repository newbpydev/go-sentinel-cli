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

	// Timing information
	StartTime     time.Time
	Duration      time.Duration
	PhaseDuration map[string]time.Duration
}

// SummaryRenderer renders the test run summary
type SummaryRenderer struct {
	writer    io.Writer
	formatter *ColorFormatter
}

// NewSummaryRenderer creates a new SummaryRenderer
func NewSummaryRenderer(writer io.Writer, formatter *ColorFormatter) *SummaryRenderer {
	return &SummaryRenderer{
		writer:    writer,
		formatter: formatter,
	}
}

// RenderSummary renders the summary of a test run
func (r *SummaryRenderer) RenderSummary(stats *TestRunStats) error {
	// Skip if we don't have any stats
	if stats == nil {
		return nil
	}

	// Format and render the test files line
	filesLine := r.formatTestFilesLine(stats)
	if _, err := fmt.Fprintln(r.writer, filesLine); err != nil {
		return err
	}

	// Format and render the tests line
	testsLine := r.formatTestsLine(stats)
	if _, err := fmt.Fprintln(r.writer, testsLine); err != nil {
		return err
	}

	// Format and render the start time line
	startLine := r.formatStartTimeLine(stats)
	if _, err := fmt.Fprintln(r.writer, startLine); err != nil {
		return err
	}

	// Format and render the duration line
	durationLine := r.formatDurationLine(stats)
	if _, err := fmt.Fprintln(r.writer, durationLine); err != nil {
		return err
	}

	return nil
}

// formatTestFilesLine formats the "Test Files" line
func (r *SummaryRenderer) formatTestFilesLine(stats *TestRunStats) string {
	var parts []string

	// Add the failed files count if any
	if stats.FailedFiles > 0 {
		failed := fmt.Sprintf("%d failed", stats.FailedFiles)
		parts = append(parts, r.formatter.Red(failed))
	}

	// Add the passed files count if any
	if stats.PassedFiles > 0 {
		passed := fmt.Sprintf("%d passed", stats.PassedFiles)
		parts = append(parts, r.formatter.Green(passed))
	}

	// Format with totals
	total := ""
	if stats.TotalFiles > 0 {
		total = fmt.Sprintf("(%d)", stats.TotalFiles)
	}

	// Format the final line
	return fmt.Sprintf("Test Files %s %s",
		strings.Join(parts, " | "),
		r.formatter.Dim(total),
	)
}

// formatTestsLine formats the "Tests" line
func (r *SummaryRenderer) formatTestsLine(stats *TestRunStats) string {
	var parts []string

	// Add the failed tests count if any
	if stats.FailedTests > 0 {
		failed := fmt.Sprintf("%d failed", stats.FailedTests)
		parts = append(parts, r.formatter.Red(failed))
	}

	// Add the passed tests count if any
	if stats.PassedTests > 0 {
		passed := fmt.Sprintf("%d passed", stats.PassedTests)
		parts = append(parts, r.formatter.Green(passed))
	}

	// Add the skipped tests count if any
	if stats.SkippedTests > 0 {
		skipped := fmt.Sprintf("%d skipped", stats.SkippedTests)
		parts = append(parts, r.formatter.Yellow(skipped))
	}

	// Format with totals
	total := ""
	if stats.TotalTests > 0 {
		total = fmt.Sprintf("(%d)", stats.TotalTests)
	}

	// Format the final line
	return fmt.Sprintf("     Tests %s %s",
		strings.Join(parts, " | "),
		r.formatter.Dim(total),
	)
}

// formatStartTimeLine formats the "Start at" line
func (r *SummaryRenderer) formatStartTimeLine(stats *TestRunStats) string {
	// Format the start time
	timeStr := stats.StartTime.Format("15:04:05")
	return fmt.Sprintf("  Start at %s", timeStr)
}

// formatDurationLine formats the "Duration" line with phase timing
func (r *SummaryRenderer) formatDurationLine(stats *TestRunStats) string {
	// Format the total duration
	durationStr := formatDuration(stats.Duration)

	// Format the phase timings if any
	phaseStr := ""
	if len(stats.PhaseDuration) > 0 {
		var phases []string
		for phase, duration := range stats.PhaseDuration {
			phases = append(phases, fmt.Sprintf("%s %s", phase, formatDuration(duration)))
		}
		phaseStr = fmt.Sprintf(" (%s)", strings.Join(phases, ", "))
	}

	return fmt.Sprintf("  Duration %s%s", durationStr, r.formatter.Dim(phaseStr))
}

// formatDuration formats a duration in a human-readable format
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}

	// Format as seconds with 2 decimal places
	seconds := float64(d) / float64(time.Second)
	return fmt.Sprintf("%.2fs", seconds)
}
