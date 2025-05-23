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

	// Real phase durations (only populated with actual measurements)
	Phases map[string]time.Duration
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
	// Render centered "Test Summary" header with separator line
	r.renderSummaryHeader()

	// Test files
	fileStats := fmt.Sprintf("Test Files: ")

	if stats.PassedFiles > 0 {
		fileStats += r.formatter.Green(fmt.Sprintf("%d passed", stats.PassedFiles))
	}

	if stats.FailedFiles > 0 {
		if stats.PassedFiles > 0 {
			fileStats += " | "
		}
		fileStats += r.formatter.Red(fmt.Sprintf("%d failed", stats.FailedFiles))
	}

	fileStats += fmt.Sprintf(" (%d)", stats.TotalFiles)
	fmt.Fprintln(r.writer, fileStats)

	// Tests
	testStats := fmt.Sprintf("Tests: ")

	if stats.PassedTests > 0 {
		testStats += r.formatter.Green(fmt.Sprintf("%d passed", stats.PassedTests))
	}

	if stats.FailedTests > 0 {
		if stats.PassedTests > 0 {
			testStats += " | "
		}
		testStats += r.formatter.Red(fmt.Sprintf("%d failed", stats.FailedTests))
	}

	if stats.SkippedTests > 0 {
		if stats.PassedTests > 0 || stats.FailedTests > 0 {
			testStats += " | "
		}
		testStats += r.formatter.Yellow(fmt.Sprintf("%d skipped", stats.SkippedTests))
	}

	testStats += fmt.Sprintf(" (%d)", stats.TotalTests)
	fmt.Fprintln(r.writer, testStats)

	// Start and End times
	startTime := stats.StartTime.Format("15:04:05")
	endTime := stats.EndTime.Format("15:04:05")

	fmt.Fprintf(r.writer, "Start at: %s\n", r.formatter.Gray(startTime))
	fmt.Fprintf(r.writer, "End at: %s\n", r.formatter.Gray(endTime))

	// Duration with phase breakdown (only when real measurements are available)
	durationText := fmt.Sprintf("Duration: %s", r.formatter.Gray(formatDuration(stats.Duration)))

	// Add phase timing if we have real measurements (not estimates)
	if len(stats.Phases) > 0 {
		durationText += " ("
		phaseTexts := []string{}

		// Define order for consistent display with what each phase represents:
		// - setup: Time from start until first test begins
		// - collect: Time spent discovering and preparing tests
		// - tests: Time actually executing test code
		// - teardown: Time after tests complete until end
		phaseOrder := []string{"setup", "collect", "tests", "teardown"}

		for _, phaseName := range phaseOrder {
			if duration, exists := stats.Phases[phaseName]; exists && duration > 0 {
				phaseTexts = append(phaseTexts, fmt.Sprintf("%s %s",
					phaseName, r.formatter.Gray(formatDuration(duration))))
			}
		}

		// Add any other phases not in the predefined order
		for phaseName, duration := range stats.Phases {
			if duration <= 0 {
				continue // Skip zero durations
			}
			found := false
			for _, orderedPhase := range phaseOrder {
				if orderedPhase == phaseName {
					found = true
					break
				}
			}
			if !found {
				phaseTexts = append(phaseTexts, fmt.Sprintf("%s %s",
					phaseName, r.formatter.Gray(formatDuration(duration))))
			}
		}

		if len(phaseTexts) > 0 {
			durationText += strings.Join(phaseTexts, ", ") + ")"
		}
	}

	fmt.Fprintln(r.writer, durationText)

	// Add exactly one blank line after duration, then separator
	fmt.Fprintln(r.writer) // One blank line after duration
	fmt.Fprintln(r.writer, r.formatter.Dim(strings.Repeat("─", r.width)))

	return nil
}

// renderSummaryHeader renders a centered "Test Summary" header on a separator line
func (r *SummaryRenderer) renderSummaryHeader() {
	headerText := " Test Summary " // Add spaces around the text

	// Calculate padding to center the header
	headerLength := len(headerText)
	totalPadding := r.width - headerLength
	leftPadding := totalPadding / 2
	rightPadding := totalPadding - leftPadding

	// Create the centered header line
	var headerLine strings.Builder
	headerLine.WriteString(strings.Repeat("─", leftPadding))
	headerLine.WriteString(headerText)
	headerLine.WriteString(strings.Repeat("─", rightPadding))

	// Add padding above and below the header (one line each)
	fmt.Fprintln(r.writer) // Padding line above
	fmt.Fprintln(r.writer, r.formatter.Dim(headerLine.String()))
	fmt.Fprintln(r.writer) // Padding line below
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
