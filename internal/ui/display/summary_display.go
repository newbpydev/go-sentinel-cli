package display

import (
	"fmt"
	"io"
	"time"

	"github.com/newbpydev/go-sentinel/internal/ui/colors"
)

// SummaryDisplayInterface defines the contract for summary display functionality
type SummaryDisplayInterface interface {
	// RenderSummary renders the complete test run summary
	RenderSummary(stats *TestRunStats) error

	// RenderTestFilesSummary renders the test files summary line
	RenderTestFilesSummary(totalFiles, passedFiles, failedFiles int) error

	// RenderTestsSummary renders the tests summary line
	RenderTestsSummary(totalTests, passedTests, failedTests int) error

	// RenderTimingSummary renders the timing information
	RenderTimingSummary(startTime time.Time, duration time.Duration) error
}

// TestRunStats contains statistics for a test run
type TestRunStats struct {
	TotalFiles  int
	PassedFiles int
	FailedFiles int
	TotalTests  int
	PassedTests int
	FailedTests int
	StartTime   time.Time
	Duration    time.Duration
}

// SummaryRenderer renders test run summary information
type SummaryRenderer struct {
	writer    io.Writer
	formatter colors.FormatterInterface
	icons     colors.IconProviderInterface
	width     int
}

// NewSummaryRenderer creates a new SummaryRenderer
func NewSummaryRenderer(writer io.Writer, formatter colors.FormatterInterface, icons colors.IconProviderInterface, width int) *SummaryRenderer {
	if writer == nil {
		panic("writer cannot be nil")
	}
	if formatter == nil {
		panic("formatter cannot be nil")
	}
	if icons == nil {
		panic("icons cannot be nil")
	}

	return &SummaryRenderer{
		writer:    writer,
		formatter: formatter,
		icons:     icons,
		width:     width,
	}
}

// NewSummaryRendererWithDefaults creates a SummaryRenderer with auto-detected defaults
func NewSummaryRendererWithDefaults(writer io.Writer) *SummaryRenderer {
	formatter := colors.NewAutoColorFormatter()
	icons := colors.NewAutoIconProvider()
	return NewSummaryRenderer(writer, formatter, icons, 80)
}

// RenderSummary renders the complete test run summary
func (r *SummaryRenderer) RenderSummary(stats *TestRunStats) error {
	if stats == nil {
		return fmt.Errorf("stats cannot be nil")
	}

	// Render test files summary
	if err := r.RenderTestFilesSummary(stats.TotalFiles, stats.PassedFiles, stats.FailedFiles); err != nil {
		return fmt.Errorf("failed to render test files summary: %w", err)
	}

	// Render tests summary
	if err := r.RenderTestsSummary(stats.TotalTests, stats.PassedTests, stats.FailedTests); err != nil {
		return fmt.Errorf("failed to render tests summary: %w", err)
	}

	// Render timing summary
	if err := r.RenderTimingSummary(stats.StartTime, stats.Duration); err != nil {
		return fmt.Errorf("failed to render timing summary: %w", err)
	}

	return nil
}

// RenderTestFilesSummary renders the test files summary line
func (r *SummaryRenderer) RenderTestFilesSummary(totalFiles, passedFiles, failedFiles int) error {
	// Format: "Test Files  1 failed | 7 passed (8)"
	var parts []string

	if failedFiles > 0 {
		failedText := r.formatter.Red(fmt.Sprintf("%d failed", failedFiles))
		parts = append(parts, failedText)
	}

	if passedFiles > 0 {
		passedText := r.formatter.Green(fmt.Sprintf("%d passed", passedFiles))
		parts = append(parts, passedText)
	}

	// Join parts with " | "
	var statusText string
	if len(parts) > 0 {
		statusText = fmt.Sprintf(" %s", joinParts(parts, " | "))
	}

	// Add total in parentheses if there are any files
	if totalFiles > 0 {
		statusText += fmt.Sprintf(" (%d)", totalFiles)
	}

	line := fmt.Sprintf("Test Files%s", statusText)
	_, err := fmt.Fprintln(r.writer, line)
	if err != nil {
		return fmt.Errorf("failed to write test files summary: %w", err)
	}

	return nil
}

// RenderTestsSummary renders the tests summary line
func (r *SummaryRenderer) RenderTestsSummary(totalTests, passedTests, failedTests int) error {
	// Format: "Tests       8 failed | 70 passed (78)"
	var parts []string

	if failedTests > 0 {
		failedText := r.formatter.Red(fmt.Sprintf("%d failed", failedTests))
		parts = append(parts, failedText)
	}

	if passedTests > 0 {
		passedText := r.formatter.Green(fmt.Sprintf("%d passed", passedTests))
		parts = append(parts, passedText)
	}

	// Join parts with " | "
	var statusText string
	if len(parts) > 0 {
		statusText = fmt.Sprintf(" %s", joinParts(parts, " | "))
	}

	// Add total in parentheses if there are any tests
	if totalTests > 0 {
		statusText += fmt.Sprintf(" (%d)", totalTests)
	}

	line := fmt.Sprintf("Tests%s", statusText)
	_, err := fmt.Fprintln(r.writer, line)
	if err != nil {
		return fmt.Errorf("failed to write tests summary: %w", err)
	}

	return nil
}

// RenderTimingSummary renders the timing information
func (r *SummaryRenderer) RenderTimingSummary(startTime time.Time, duration time.Duration) error {
	// Format start time as "Start at  11:39:32"
	timeStr := startTime.Format("15:04:05")
	startLine := fmt.Sprintf("Start at  %s", timeStr)

	_, err := fmt.Fprintln(r.writer, startLine)
	if err != nil {
		return fmt.Errorf("failed to write start time: %w", err)
	}

	// Format duration
	var durationStr string
	if duration < time.Second {
		durationStr = fmt.Sprintf("%dms", duration.Milliseconds())
	} else {
		durationStr = fmt.Sprintf("%.2fs", duration.Seconds())
	}

	durationLine := fmt.Sprintf("Duration  %s", durationStr)
	_, err = fmt.Fprintln(r.writer, durationLine)
	if err != nil {
		return fmt.Errorf("failed to write duration: %w", err)
	}

	return nil
}

// joinParts joins string parts with a separator
func joinParts(parts []string, separator string) string {
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}

	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += separator + parts[i]
	}
	return result
}
