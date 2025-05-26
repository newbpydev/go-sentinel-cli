// Package display provides final summary rendering for test execution completion
package display

import (
	"strings"
	"time"

	"github.com/newbpydev/go-sentinel/internal/ui/colors"
	"github.com/newbpydev/go-sentinel/internal/ui/icons"
	"github.com/newbpydev/go-sentinel/pkg/models"
)

// FinalSummaryRenderer handles rendering the final test execution summary
type FinalSummaryRenderer struct {
	formatter       *colors.ColorFormatter
	icons           icons.IconProvider
	spacingManager  *SpacingManager
	timingFormatter *TimingFormatter
	config          *Config

	// Display options
	terminalWidth int
	showTiming    bool
	showMemory    bool
	showCoverage  bool
	indentLevel   int
}

// FinalSummaryRenderOptions configures final summary rendering
type FinalSummaryRenderOptions struct {
	TerminalWidth int
	ShowTiming    bool
	ShowMemory    bool
	ShowCoverage  bool
	IndentLevel   int
}

// NewFinalSummaryRenderer creates a new final summary renderer
func NewFinalSummaryRenderer(config *Config, options *FinalSummaryRenderOptions) *FinalSummaryRenderer {
	formatter := colors.NewAutoColorFormatter()

	// Detect terminal capabilities for icon selection
	detector := colors.NewTerminalDetector()
	var iconProvider icons.IconProvider
	if detector.SupportsUnicode() {
		iconProvider = icons.NewUnicodeProvider()
	} else {
		iconProvider = icons.NewASCIIProvider()
	}

	spacingManager := NewSpacingManager(&SpacingConfig{
		BaseIndent:    0, // No base indent for summary
		TestIndent:    2,
		SubtestIndent: 4,
		ErrorIndent:   4,
	})

	timingFormatter := NewTimingFormatter(&TimingConfig{
		ShowMilliseconds: true,
		ShowMicroseconds: false,
		IntegerFormat:    true,
		MinWidth:         4,
	})

	// Set defaults if options not provided or apply defaults to unset fields
	if options == nil {
		options = &FinalSummaryRenderOptions{
			TerminalWidth: 110,
			ShowTiming:    true,
			ShowMemory:    true,
			ShowCoverage:  false,
			IndentLevel:   0,
		}
	} else {
		// Apply defaults to zero-value fields
		if options.TerminalWidth == 0 {
			options.TerminalWidth = 110
		}
	}

	return &FinalSummaryRenderer{
		formatter:       formatter,
		icons:           iconProvider,
		spacingManager:  spacingManager,
		timingFormatter: timingFormatter,
		config:          config,
		terminalWidth:   options.TerminalWidth,
		showTiming:      options.ShowTiming,
		showMemory:      options.ShowMemory,
		showCoverage:    options.ShowCoverage,
		indentLevel:     options.IndentLevel,
	}
}

// RenderFinalSummary renders the complete final summary section
func (r *FinalSummaryRenderer) RenderFinalSummary(summary *models.TestSummary) string {
	if summary == nil {
		return ""
	}

	var result strings.Builder

	// Render section separator with "Test Summary" centered
	separator := r.renderSummarySeparator("Test Summary")
	result.WriteString(separator)
	result.WriteString("\n")

	// Render test file statistics
	testFilesLine := r.renderTestFilesStats(summary)
	result.WriteString(testFilesLine)
	result.WriteString("\n")

	// Render test statistics
	testsLine := r.renderTestsStats(summary)
	result.WriteString(testsLine)
	result.WriteString("\n")

	// Render timing information
	timingLine := r.renderTimingStats(summary)
	result.WriteString(timingLine)
	result.WriteString("\n")

	// Render final separator
	finalSeparator := r.renderFinalSeparator()
	result.WriteString(finalSeparator)
	result.WriteString("\n")

	// Render completion message with timing icon
	completionLine := r.renderCompletionMessage(summary)
	result.WriteString(completionLine)

	return result.String()
}

// renderSummarySeparator renders the test summary section separator with centered title
func (r *FinalSummaryRenderer) renderSummarySeparator(title string) string {
	// Calculate dash count for each side
	totalWidth := r.terminalWidth
	if totalWidth < 110 {
		totalWidth = 110
	}

	titleLen := len(title)
	dashCount := (totalWidth - titleLen - 2) / 2 // 2 spaces around title

	leftDashes := strings.Repeat("─", dashCount)
	rightDashes := strings.Repeat("─", totalWidth-dashCount-titleLen-2)

	return leftDashes + " " + title + " " + rightDashes
}

// renderFinalSeparator renders the final separator line
func (r *FinalSummaryRenderer) renderFinalSeparator() string {
	width := r.terminalWidth
	if width < 110 {
		width = 110
	}
	return strings.Repeat("─", width)
}

// renderTestFilesStats renders test file statistics
// Format: "Test Files: 1 passed | 1 failed (2)"
func (r *FinalSummaryRenderer) renderTestFilesStats(summary *models.TestSummary) string {
	var result strings.Builder

	result.WriteString("Test Files: ")

	// Calculate file stats (assuming one file per package for now)
	passedFiles := 0
	failedFiles := len(summary.FailedPackages)
	totalFiles := summary.PackageCount

	if summary.PackageCount > failedFiles {
		passedFiles = summary.PackageCount - failedFiles
	}

	// Render passed count
	if passedFiles > 0 {
		result.WriteString(r.formatter.Green(r.intToString(passedFiles)))
		result.WriteString(" passed")
	}

	// Add separator if both passed and failed
	if passedFiles > 0 && failedFiles > 0 {
		result.WriteString(" | ")
	}

	// Render failed count
	if failedFiles > 0 {
		result.WriteString(r.formatter.Red(r.intToString(failedFiles)))
		result.WriteString(" failed")
	}

	// Add total in parentheses
	result.WriteString(" (")
	result.WriteString(r.intToString(totalFiles))
	result.WriteString(")")

	return result.String()
}

// renderTestsStats renders test statistics
// Format: "Tests: 142 passed | 26 failed | 7 skipped (175)"
func (r *FinalSummaryRenderer) renderTestsStats(summary *models.TestSummary) string {
	var result strings.Builder

	result.WriteString("Tests: ")

	parts := []string{}

	// Render passed count
	if summary.PassedTests > 0 {
		passedPart := r.formatter.Green(r.intToString(summary.PassedTests)) + " passed"
		parts = append(parts, passedPart)
	}

	// Render failed count
	if summary.FailedTests > 0 {
		failedPart := r.formatter.Red(r.intToString(summary.FailedTests)) + " failed"
		parts = append(parts, failedPart)
	}

	// Render skipped count
	if summary.SkippedTests > 0 {
		skippedPart := r.formatter.Yellow(r.intToString(summary.SkippedTests)) + " skipped"
		parts = append(parts, skippedPart)
	}

	// Join parts with " | "
	result.WriteString(strings.Join(parts, " | "))

	// Add total in parentheses
	result.WriteString(" (")
	result.WriteString(r.intToString(summary.TotalTests))
	result.WriteString(")")

	return result.String()
}

// renderTimingStats renders timing information
// Format: "Start at: 12:17:10" / "End at: 12:17:23" / "Duration: 12.96s (setup 7.61s, tests 4.36s, teardown 979ms)"
func (r *FinalSummaryRenderer) renderTimingStats(summary *models.TestSummary) string {
	if !r.showTiming {
		return ""
	}

	var result strings.Builder

	// Start time
	startTime := r.formatTime(summary.StartTime)
	result.WriteString("Start at: ")
	result.WriteString(startTime)
	result.WriteString("\n")

	// End time
	endTime := r.formatTime(summary.EndTime)
	result.WriteString("End at: ")
	result.WriteString(endTime)
	result.WriteString("\n")

	// Duration breakdown - use the simpler TimingFormatter formatting
	duration := r.timingFormatter.FormatDuration(summary.TotalDuration)
	result.WriteString("Duration: ")
	result.WriteString(duration)

	// Add breakdown if available (placeholder for now)
	// TODO: Add actual breakdown when available in models
	if summary.TotalDuration > time.Second {
		result.WriteString(" (setup 0ms, tests ")
		result.WriteString(duration)
		result.WriteString(", teardown 0ms)")
	}

	return result.String()
}

// renderCompletionMessage renders the final completion message with timing icon
// Format: "⏱️  Tests completed in 13.1472234s"
func (r *FinalSummaryRenderer) renderCompletionMessage(summary *models.TestSummary) string {
	var result strings.Builder

	// Add timing icon
	timingIcon, _ := r.icons.GetIcon("timing")
	result.WriteString(timingIcon)
	result.WriteString("  Tests completed in ")

	// Format duration with higher precision for completion message
	duration := r.formatPreciseDuration(summary.TotalDuration)
	result.WriteString(duration)

	return result.String()
}

// formatTime formats a time for display (HH:MM:SS format)
func (r *FinalSummaryRenderer) formatTime(t time.Time) string {
	if t.IsZero() {
		return "00:00:00"
	}

	hour := t.Hour()
	minute := t.Minute()
	second := t.Second()

	return r.padZero(hour) + ":" + r.padZero(minute) + ":" + r.padZero(second)
}

// formatPreciseDuration formats duration with higher precision for final message
func (r *FinalSummaryRenderer) formatPreciseDuration(duration time.Duration) string {
	seconds := duration.Seconds()
	return r.floatToString(seconds) + "s"
}

// padZero pads a number with leading zero if needed
func (r *FinalSummaryRenderer) padZero(n int) string {
	if n < 10 {
		return "0" + r.intToString(n)
	}
	return r.intToString(n)
}

// floatToString converts float64 to string with appropriate precision
func (r *FinalSummaryRenderer) floatToString(f float64) string {
	// Convert to integer part and fractional part
	intPart := int(f)
	fracPart := f - float64(intPart)

	result := r.intToString(intPart)

	// Add fractional part with 7 decimal places (matching the example)
	result += "."

	// Extract exactly 7 decimal places to match expected precision
	fracPart *= 10000000 // multiply by 10^7
	fracDigits := int(fracPart) % 10000000

	// Format with leading zeros if needed
	fracStr := r.intToString(fracDigits)
	for len(fracStr) < 7 {
		fracStr = "0" + fracStr
	}
	result += fracStr

	return result
}

// intToString converts an integer to string without importing fmt
func (r *FinalSummaryRenderer) intToString(n int) string {
	if n == 0 {
		return "0"
	}

	var result []byte
	negative := n < 0
	if negative {
		n = -n
	}

	for n > 0 {
		result = append([]byte{byte('0' + n%10)}, result...)
		n /= 10
	}

	if negative {
		result = append([]byte{'-'}, result...)
	}

	return string(result)
}

// SetTerminalWidth sets the terminal width for formatting
func (r *FinalSummaryRenderer) SetTerminalWidth(width int) {
	if width < 110 {
		width = 110
	}
	r.terminalWidth = width
}

// SetShowTiming enables/disables timing display
func (r *FinalSummaryRenderer) SetShowTiming(show bool) {
	r.showTiming = show
}

// SetShowMemory enables/disables memory usage display
func (r *FinalSummaryRenderer) SetShowMemory(show bool) {
	r.showMemory = show
}

// SetShowCoverage enables/disables coverage display
func (r *FinalSummaryRenderer) SetShowCoverage(show bool) {
	r.showCoverage = show
}

// GetTerminalWidth returns the current terminal width
func (r *FinalSummaryRenderer) GetTerminalWidth() int {
	return r.terminalWidth
}

// IsShowTiming returns whether timing display is enabled
func (r *FinalSummaryRenderer) IsShowTiming() bool {
	return r.showTiming
}

// IsShowMemory returns whether memory display is enabled
func (r *FinalSummaryRenderer) IsShowMemory() bool {
	return r.showMemory
}

// IsShowCoverage returns whether coverage display is enabled
func (r *FinalSummaryRenderer) IsShowCoverage() bool {
	return r.showCoverage
}
