package cli

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
)

// Icon constants
const (
	// Unicode icons
	IconPass    = "✓"
	IconFail    = "✕"
	IconSkip    = "○"
	IconRunning = "⠋"

	// ASCII icons
	ASCIIIconPass    = "+"
	ASCIIIconFail    = "-"
	ASCIIIconSkip    = "o"
	ASCIIIconRunning = "*"

	// Windows icons
	WinIconPass    = "√"
	WinIconFail    = "×"
	WinIconSkip    = "o"
	WinIconRunning = "*"
)

// Style handles terminal styling and icons
type Style struct {
	useColors    bool
	useIcons     bool
	useEmoji     bool
	isWindows    bool
	forceColors  bool
	forceNoColor bool
}

var (
	// Colors
	successColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00"))
	errorColor   = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	skipColor    = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffff00"))
	dimColor     = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	boldStyle    = lipgloss.NewStyle().Bold(true)

	// Icons
	passedIcon  = "✓"
	failedIcon  = "✕"
	skippedIcon = "○"
	runningIcon = "⠋"

	// Headers
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#1a1a1a")).
			Padding(0, 1)

	// Test name styles
	testNameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff"))

	// Duration styles
	durationStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	// Error styles
	errorHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ff0000")).
				Bold(true)
	errorBodyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff6666"))
	errorLocationStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#666666"))
)

// NewStyle creates a new Style instance
func NewStyle() *Style {
	s := &Style{
		isWindows: runtime.GOOS == "windows",
	}
	s.detect()
	return s
}

// detect determines the terminal capabilities
func (s *Style) detect() {
	// Check if colors should be forced on or off
	s.forceColors = os.Getenv("FORCE_COLOR") != ""
	s.forceNoColor = os.Getenv("NO_COLOR") != "" || os.Getenv("NOCOLOR") != ""

	// Determine if we should use colors
	s.useColors = s.forceColors || (!s.forceNoColor && isatty.IsTerminal(os.Stdout.Fd()))

	// Determine if we can use Unicode icons/emoji
	s.useIcons = s.useColors && (!s.isWindows || s.forceColors)

	// On Windows, only use Unicode if running in a modern terminal
	if s.isWindows && !s.forceColors {
		// Check for modern Windows terminals that support Unicode
		if os.Getenv("WT_SESSION") != "" || // Windows Terminal
			os.Getenv("TERM_PROGRAM") == "vscode" || // VS Code terminal
			os.Getenv("CMDER_ROOT") != "" { // Cmder
			s.useIcons = true
		}
	}

	// Emoji support follows icon support
	s.useEmoji = s.useIcons
}

// FormatTestName formats a test result with appropriate color and icon
func (s *Style) FormatTestName(result *TestResult) string {
	var icon, name string

	// Choose icon and color based on test status
	switch result.Status {
	case TestStatusPassed:
		if s.useIcons {
			icon = IconPass
		} else {
			icon = ASCIIIconPass
		}
		name = result.Name
	case TestStatusFailed:
		if s.useIcons {
			icon = IconFail
		} else {
			icon = ASCIIIconFail
		}
		name = result.Name
	case TestStatusSkipped:
		if s.useIcons {
			icon = IconSkip
		} else {
			icon = ASCIIIconSkip
		}
		name = result.Name
	case TestStatusRunning:
		if s.useIcons {
			icon = IconRunning
		} else {
			icon = ASCIIIconRunning
		}
		name = result.Name
	default:
		icon = " "
		name = result.Name
	}

	if s.useColors {
		switch result.Status {
		case TestStatusPassed:
			name = successColor.Render(name)
		case TestStatusFailed:
			name = errorColor.Render(name)
		case TestStatusSkipped:
			name = skipColor.Render(name)
		case TestStatusRunning:
			name = testNameStyle.Render(name)
		}
	}

	return fmt.Sprintf("%s %s", icon, name)
}

// FormatDuration formats a duration in a human-readable way
func (s *Style) FormatDuration(seconds float64) string {
	d := time.Duration(seconds * float64(time.Second))
	if d < time.Second {
		return durationStyle.Render(fmt.Sprintf("%dms", d.Milliseconds()))
	}
	return durationStyle.Render(fmt.Sprintf("%.2fs", seconds))
}

// FormatErrorLocation formats a source location for error output
func (s *Style) FormatErrorLocation(loc *SourceLocation) string {
	if loc == nil {
		return ""
	}
	return errorLocationStyle.Render(fmt.Sprintf("%s:%d", loc.File, loc.Line))
}

// FormatErrorSnippet formats a code snippet for error output
func (s *Style) FormatErrorSnippet(snippet string, errorLine int) string {
	// Return the snippet as is, preserving original formatting
	return strings.TrimSpace(snippet)
}

// FormatSummary formats test run summary information
func (s *Style) FormatSummary(run *TestRun) string {
	var parts []string

	// Total tests
	parts = append(parts, fmt.Sprintf("Total: %d", run.NumTotal))

	// Passed tests
	if run.NumPassed > 0 {
		parts = append(parts, fmt.Sprintf("Passed: %d", run.NumPassed))
	}

	// Failed tests
	if run.NumFailed > 0 {
		parts = append(parts, fmt.Sprintf("Failed: %d", run.NumFailed))
	}

	// Skipped tests
	if run.NumSkipped > 0 {
		parts = append(parts, fmt.Sprintf("Skipped: %d", run.NumSkipped))
	}

	// Duration
	parts = append(parts, fmt.Sprintf("Time: %.2fs", run.Duration.Seconds()))

	return strings.Join(parts, " ")
}

// formatIcon formats an icon with color if enabled
func (s *Style) formatIcon(icon string, col color.Attribute) string {
	if !s.useIcons {
		return " "
	}
	if s.useColors {
		return color.New(col).Sprint(icon)
	}
	return icon
}

// FormatHeader formats a header with background color
func (s *Style) FormatHeader(text string) string {
	return headerStyle.Render(text)
}

// FormatErrorHeader formats an error header
func (s *Style) FormatErrorHeader(text string) string {
	return errorHeaderStyle.Render(text)
}

// FormatErrorBody formats error details
func (s *Style) FormatErrorBody(text string) string {
	return errorBodyStyle.Render(text)
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
