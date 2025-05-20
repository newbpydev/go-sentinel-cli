// Package cli provides command-line interface functionality for go-sentinel.
package cli

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
)

// Test status icons
const (
	IconPass         = "✓"
	IconFail         = "✕"
	IconSkip         = "○"
	IconRunning      = "⠋"
	ASCIIIconPass    = "+"
	ASCIIIconFail    = "x"
	ASCIIIconSkip    = "o"
	ASCIIIconRunning = "*"
	WinIconPass      = "+"
	WinIconFail      = "x"
	WinIconSkip      = "o"
	WinIconRunning   = "*"
)

// Styles for test output
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff"))

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#16a34a")) // Vitest green

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#dc2626")) // Vitest red

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ca8a04")) // Vitest yellow

	dimStyle = lipgloss.NewStyle().
			Faint(true)

	// Test status styles
	passedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#16a34a")).
			SetString("✓")

	failedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#dc2626")).
			SetString("✕")

	skippedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ca8a04")).
			SetString("○")

	runningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3b82f6")).
			SetString("⠋")
)

// Style handles terminal styling and formatting
type Style struct {
	useColors bool
	useIcons  bool
	isWindows bool
	useEmoji  bool
}

// NewStyle creates a new style instance
func NewStyle(useColors bool) *Style {
	s := &Style{
		useColors: useColors,
		useIcons:  true,
		isWindows: runtime.GOOS == "windows",
		useEmoji:  true,
	}
	s.Detect()
	return s
}

// FormatTestName formats a test name with status icon and color
func (s *Style) FormatTestName(result *TestResult) string {
	icon := s.StatusIcon(result.Status)
	name := result.Name

	if s.useColors {
		switch result.Status {
		case TestStatusPassed:
			return fmt.Sprintf("%s %s", passedStyle.Render(icon), successStyle.Render(name))
		case TestStatusFailed:
			return fmt.Sprintf("%s %s", failedStyle.Render(icon), errorStyle.Render(name))
		case TestStatusSkipped:
			return fmt.Sprintf("%s %s", skippedStyle.Render(icon), warningStyle.Render(name))
		case TestStatusRunning:
			return fmt.Sprintf("%s %s", runningStyle.Render(icon), name)
		default:
			return fmt.Sprintf("%s %s", icon, name)
		}
	}

	return fmt.Sprintf("%s %s", icon, name)
}

// FormatTestSummary formats a test summary line with colors
func (s *Style) FormatTestSummary(label string, failed, passed, skipped, total int) string {
	// Add label with consistent padding (Test Files is the longest at 10 chars)
	parts := []string{fmt.Sprintf("%-12s", label)}

	// Use more vibrant colors for the numbers
	boldFailStyle := errorStyle.Bold(true)
	boldPassStyle := successStyle.Bold(true)
	boldSkipStyle := warningStyle.Bold(true)

	// Vitest style: If no failures, everything is green; otherwise show failures in red
	if failed == 0 && skipped == 0 {
		// Everything passed - all green
		parts = append(parts, boldPassStyle.Render(fmt.Sprintf("%d passed", passed)))
	} else {
		// Show detailed breakdown with vibrant colors
		var stats []string
		if failed > 0 {
			stats = append(stats, boldFailStyle.Render(fmt.Sprintf("%d failed", failed)))
		}
		if passed > 0 {
			stats = append(stats, boldPassStyle.Render(fmt.Sprintf("%d passed", passed)))
		}
		if skipped > 0 {
			stats = append(stats, boldSkipStyle.Render(fmt.Sprintf("%d skipped", skipped)))
		}

		// Join stats with separator
		if len(stats) > 0 {
			parts = append(parts, strings.Join(stats, " | "))
		}
	}

	// Add total in parentheses with dim style
	parts = append(parts, dimStyle.Render(fmt.Sprintf("(%d)", total)))

	return strings.Join(parts, " ")
}

// FormatTimestamp formats a timestamp line with consistent padding
func (s *Style) FormatTimestamp(label string, t time.Time) string {
	// Add label with consistent padding (Test Files is the longest at 10 chars)
	parts := []string{fmt.Sprintf("%-12s", label)}

	// Format time as HH:MM:SS
	timeStr := t.Format("15:04:05")
	parts = append(parts, timeStr)

	return strings.Join(parts, " ")
}

// FormatDuration formats a duration line with consistent padding
func (s *Style) FormatDuration(label string, duration string) string {
	// Add label with consistent padding (Test Files is the longest at 10 chars)
	parts := []string{fmt.Sprintf("%-12s", label)}
	parts = append(parts, duration)
	return strings.Join(parts, " ")
}

// FormatHeader formats a header line
func (s *Style) FormatHeader(text string) string {
	if s.useColors {
		return titleStyle.Render(text)
	}
	return text
}

// FormatErrorHeader formats an error header
func (s *Style) FormatErrorHeader(text string) string {
	if s.useColors {
		return errorStyle.Render(text)
	}
	return text
}

// FormatFailedSuite formats a failed suite path
func (s *Style) FormatFailedSuite(path string) string {
	if s.useColors {
		return errorStyle.Render(fmt.Sprintf("  %s", path))
	}
	return fmt.Sprintf("  %s", path)
}

// FormatFailedTest formats a failed test name
func (s *Style) FormatFailedTest(name string) string {
	if s.useColors {
		return errorStyle.Render(fmt.Sprintf("    %s", name))
	}
	return fmt.Sprintf("    %s", name)
}

// FormatErrorMessage formats an error message
func (s *Style) FormatErrorMessage(message string) string {
	if s.useColors {
		return errorStyle.Render(fmt.Sprintf("      %s", message))
	}
	return fmt.Sprintf("      %s", message)
}

// FormatErrorLocation formats a test error location
func (s *Style) FormatErrorLocation(loc *SourceLocation) string {
	if s.useColors {
		return subtitleStyle.Render(fmt.Sprintf("at %s:%d", loc.File, loc.Line))
	}
	return fmt.Sprintf("at %s:%d", loc.File, loc.Line)
}

// FormatErrorSnippet formats a test error code snippet
func (s *Style) FormatErrorSnippet(snippet string, line int) string {
	lines := strings.Split(snippet, "\n")
	var formattedLines []string

	for i, l := range lines {
		if s.useColors {
			formattedLines = append(formattedLines, subtitleStyle.Render(fmt.Sprintf("  %d | %s", line+i, l)))
		} else {
			formattedLines = append(formattedLines, fmt.Sprintf("  %d | %s", line+i, l))
		}
	}

	return strings.Join(formattedLines, "\n")
}

// StatusIcon returns the appropriate icon for a test status
func (s *Style) StatusIcon(status TestStatus) string {
	if !s.useIcons {
		// Use ASCII icons when Unicode is not supported
		switch status {
		case TestStatusPassed:
			return ASCIIIconPass
		case TestStatusFailed:
			return ASCIIIconFail
		case TestStatusSkipped:
			return ASCIIIconSkip
		case TestStatusRunning:
			return ASCIIIconRunning
		default:
			return " "
		}
	}

	if s.isWindows && !s.useEmoji {
		// Use Windows-compatible icons
		switch status {
		case TestStatusPassed:
			return WinIconPass
		case TestStatusFailed:
			return WinIconFail
		case TestStatusSkipped:
			return WinIconSkip
		case TestStatusRunning:
			return WinIconRunning
		default:
			return " "
		}
	}

	// Use Unicode icons
	switch status {
	case TestStatusPassed:
		return IconPass
	case TestStatusFailed:
		return IconFail
	case TestStatusSkipped:
		return IconSkip
	case TestStatusRunning:
		return IconRunning
	default:
		return " "
	}
}

// Detect checks terminal capabilities and adjusts settings accordingly
func (s *Style) Detect() {
	// Check if colors are forced
	if os.Getenv("FORCE_COLOR") != "" {
		s.useColors = true
		s.useIcons = true
		return
	}

	// Check if colors are disabled
	if os.Getenv("NO_COLOR") != "" {
		s.useColors = false
		s.useIcons = false
		return
	}

	// Check if terminal supports colors
	if !isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		s.useColors = false
		s.useIcons = false
		return
	}

	// Check if terminal supports Unicode
	if s.isWindows {
		s.useEmoji = false
	}
}
