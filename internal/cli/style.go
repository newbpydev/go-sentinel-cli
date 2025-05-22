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

// Colors using the more vibrant palette seen in the example
const (
	// Bright, modern colors like in the provided image
	ColorSuccess   = "#4ADE80" // Bright green for passing tests
	ColorError     = "#F87171" // Bright red for failing tests
	ColorWarning   = "#FBBF24" // Amber yellow for skipped tests
	ColorRunning   = "#60A5FA" // Bright blue for running tests
	ColorDim       = "#94A3B8" // Slate gray for dimmed text
	ColorHeaderBg  = "#1E293B" // Dark slate blue for header backgrounds
	ColorText      = "#E2E8F0" // Light gray for normal text
	ColorLabelText = "#94A3B8" // Slate gray for labels
	ColorTimeText  = "#CBD5E1" // Light slate for timestamps
)

// Styles for test output
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(ColorText)).
			Background(lipgloss.Color(ColorHeaderBg)).
			Padding(0, 1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorLabelText))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorSuccess))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorError))

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorWarning))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorDim))

	// Test status styles
	passedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorSuccess)).
			SetString("✓")

	failedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorError)).
			SetString("✕")

	skippedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorWarning)).
			SetString("○")

	runningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorRunning)).
			SetString("⠋")

	// Summary styles
	summaryLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorLabelText))

	summaryFailedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(ColorError))

	summaryPassedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(ColorSuccess))

	summarySkippedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(ColorWarning))

	summaryValueStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(ColorTimeText))

	breakdownTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorDim))

	// Error formatting styles
	errorMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorError))

	errorLocationStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorDim))

	errorSnippetStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorText))

	errorValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorDim))
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
	// Right-align label with 12 characters padding
	formattedLabel := fmt.Sprintf("%-12s", label)
	if s.useColors {
		formattedLabel = summaryLabelStyle.Render(formattedLabel)
	}

	var parts []string
	if failed > 0 {
		failedStr := fmt.Sprintf("%d failed", failed)
		if s.useColors {
			failedStr = summaryFailedStyle.Bold(true).Render(failedStr)
		}
		parts = append(parts, failedStr)
	}
	if passed > 0 {
		passedStr := fmt.Sprintf("%d passed", passed)
		if s.useColors {
			passedStr = summaryPassedStyle.Render(passedStr)
		}
		parts = append(parts, passedStr)
	}
	if skipped > 0 {
		skippedStr := fmt.Sprintf("%d skipped", skipped)
		if s.useColors {
			skippedStr = summarySkippedStyle.Render(skippedStr)
		}
		parts = append(parts, skippedStr)
	}

	summary := strings.Join(parts, " | ")
	totalStr := fmt.Sprintf("(%d)", total)
	if s.useColors {
		totalStr = summaryValueStyle.Render(totalStr)
	}

	if summary != "" {
		summary = fmt.Sprintf("%s %s", summary, totalStr)
	} else if total > 0 {
		passedStr := fmt.Sprintf("%d passed", total)
		if s.useColors {
			passedStr = summaryPassedStyle.Render(passedStr)
		}
		summary = fmt.Sprintf("%s %s", passedStr, totalStr)
	}

	return fmt.Sprintf("%s %s", formattedLabel, summary)
}

// FormatTimestamp formats a timestamp line with consistent padding
func (s *Style) FormatTimestamp(label string, t time.Time) string {
	labelPart := fmt.Sprintf("%12s  ", summaryLabelStyle.Render(label))
	timeStr := summaryValueStyle.Render(t.Format("15:04:05"))
	return fmt.Sprintf("%s%s", labelPart, timeStr)
}

// FormatDuration formats the main duration value and breakdown
func (s *Style) FormatDuration(label string, mainDuration string) string {
	labelPart := fmt.Sprintf("%12s  ", summaryLabelStyle.Render(label))
	durationStr := summaryValueStyle.Render(mainDuration)
	return fmt.Sprintf("%s%s", labelPart, durationStr)
}

// FormatHeader formats a header line
func (s *Style) FormatHeader(text string) string {
	if s.useColors {
		// Create a modern header with background color and padding
		return titleStyle.Copy().
			Padding(0, 1).
			Background(lipgloss.Color(ColorHeaderBg)).
			Foreground(lipgloss.Color(ColorText)).
			Bold(true).
			Render(text)
	}
	return text
}

// FormatErrorHeader formats an error header
func (s *Style) FormatErrorHeader(text string) string {
	if s.useColors {
		// Create a bold red header with background for the error section
		return errorStyle.Copy().
			Bold(true).
			Padding(0, 1).
			Background(lipgloss.Color("#2D1414")). // Dark red background
			Render(text)
	}
	return "FAILED Tests"
}

// FormatFailedSuite formats a failed test suite path
func (s *Style) FormatFailedSuite(path string) string {
	return fmt.Sprintf("  %s", dimStyle.Render(path))
}

// FormatFailedTest formats a failed test name
func (s *Style) FormatFailedTest(name string) string {
	return fmt.Sprintf("  %s", errorStyle.Bold(true).Render(name))
}

// FormatErrorMessage formats an error message
func (s *Style) FormatErrorMessage(msg string) string {
	if s.useColors {
		return errorMessageStyle.Render(msg)
	}
	return msg
}

// FormatErrorLocation formats a source location
func (s *Style) FormatErrorLocation(loc *SourceLocation) string {
	if s.useColors {
		return errorLocationStyle.Render(fmt.Sprintf("at %s:%d", loc.File, loc.Line))
	}
	return fmt.Sprintf("at %s:%d", loc.File, loc.Line)
}

// FormatErrorSnippet formats a code snippet
func (s *Style) FormatErrorSnippet(snippet string, line int) string {
	lines := strings.Split(strings.TrimSpace(snippet), "\n")
	var formattedLines []string

	for i, l := range lines {
		lineNum := line + i
		lineStr := fmt.Sprintf("  %d | %s", lineNum, strings.TrimSpace(l))
		if s.useColors {
			formattedLines = append(formattedLines, errorSnippetStyle.Render(lineStr))
		} else {
			formattedLines = append(formattedLines, lineStr)
		}
	}

	return strings.Join(formattedLines, "\n")
}

// FormatErrorValue formats an expected or actual value
func (s *Style) FormatErrorValue(value string) string {
	if s.useColors {
		return errorValueStyle.Render(value)
	}
	return value
}

// StatusIcon returns an icon for the test status
func (s *Style) StatusIcon(status TestStatus) string {
	if s.useIcons {
		if s.isWindows && !s.useEmoji {
			// Windows without emoji support
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
		} else {
			// Modern terminals or emoji-compatible Windows
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
	} else {
		// ASCII mode
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

// FormatBreakdownText formats the breakdown text in the duration line
func (s *Style) FormatBreakdownText(text string) string {
	if s.useColors {
		return breakdownTextStyle.Render(text)
	}
	return text
}
