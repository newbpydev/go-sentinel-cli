package cli

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// ANSI color codes
const (
	// Basic formatting
	ColorReset  = "\033[0m"
	ColorBold   = "\033[1m"
	ColorDim    = "\033[2m"
	ColorItalic = "\033[3m"

	// Foreground colors
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorCyan    = "\033[36m"
	ColorGray    = "\033[90m"

	// Background colors
	ColorBgRed     = "\033[41m"
	ColorBgGreen   = "\033[42m"
	ColorBgYellow  = "\033[43m"
	ColorBgBlue    = "\033[44m"
	ColorBgMagenta = "\033[45m"
	ColorBgCyan    = "\033[46m"
)

// Vitest-inspired icons
const (
	// Unicode symbols
	IconCheckMark = "✓"
	IconCross     = "✗"
	IconSkipped   = "⃠"
	IconRunning   = "⟳"

	// ASCII fallbacks
	IconCheckMarkASCII = "√"
	IconCrossASCII     = "×"
	IconSkippedASCII   = "○"
	IconRunningASCII   = "~"
)

// ColorFormatter formats text with ANSI colors
type ColorFormatter struct {
	useColors bool
}

// NewColorFormatter creates a new ColorFormatter
func NewColorFormatter(useColors bool) *ColorFormatter {
	return &ColorFormatter{
		useColors: useColors,
	}
}

// colorize applies a color to text if colors are enabled
func (f *ColorFormatter) colorize(text, color string) string {
	if !f.useColors {
		return text
	}
	return color + text + ColorReset
}

// Red formats text with red color
func (f *ColorFormatter) Red(text string) string {
	return f.colorize(text, ColorRed)
}

// Green formats text with green color
func (f *ColorFormatter) Green(text string) string {
	return f.colorize(text, ColorGreen)
}

// Yellow formats text with yellow color
func (f *ColorFormatter) Yellow(text string) string {
	return f.colorize(text, ColorYellow)
}

// Blue formats text with blue color
func (f *ColorFormatter) Blue(text string) string {
	return f.colorize(text, ColorBlue)
}

// Magenta formats text with magenta color
func (f *ColorFormatter) Magenta(text string) string {
	return f.colorize(text, ColorMagenta)
}

// Cyan formats text with cyan color
func (f *ColorFormatter) Cyan(text string) string {
	return f.colorize(text, ColorCyan)
}

// Gray formats text with gray color
func (f *ColorFormatter) Gray(text string) string {
	return f.colorize(text, ColorGray)
}

// Bold formats text as bold
func (f *ColorFormatter) Bold(text string) string {
	return f.colorize(text, ColorBold)
}

// Dim formats text as dim
func (f *ColorFormatter) Dim(text string) string {
	return f.colorize(text, ColorDim)
}

// BgRed returns text with a red background
func (f *ColorFormatter) BgRed(text string) string {
	if !f.useColors {
		return text
	}
	return fmt.Sprintf("\033[41m%s\033[0m", text)
}

// White returns white colored text
func (f *ColorFormatter) White(text string) string {
	if !f.useColors {
		return text
	}
	return fmt.Sprintf("\033[37m%s\033[0m", text)
}

// Colorize applies a color name to text
func (f *ColorFormatter) Colorize(text, colorName string) string {
	if !f.useColors {
		return text
	}

	switch colorName {
	case "red":
		return f.Red(text)
	case "green":
		return f.Green(text)
	case "yellow":
		return f.Yellow(text)
	case "blue":
		return f.Blue(text)
	case "magenta":
		return f.Magenta(text)
	case "cyan":
		return f.Cyan(text)
	case "gray":
		return f.Gray(text)
	case "white":
		return f.White(text)
	case "bold":
		return f.Bold(text)
	case "dim":
		return f.Dim(text)
	default:
		return text
	}
}

// TerminalDetector detects terminal capabilities
type TerminalDetector struct {
	fd int
}

// NewTerminalDetector creates a new TerminalDetector
func NewTerminalDetector() *TerminalDetector {
	return &TerminalDetector{
		fd: int(os.Stdout.Fd()),
	}
}

// SupportsColor detects if the terminal supports colors
func (t *TerminalDetector) SupportsColor() bool {
	// Check if NO_COLOR environment variable is set
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Check if FORCE_COLOR environment variable is set
	if os.Getenv("FORCE_COLOR") != "" {
		return true
	}

	// Check if stdout is a terminal
	if !term.IsTerminal(t.fd) {
		return false
	}

	// Check terminal type
	termType := os.Getenv("TERM")
	if termType == "dumb" {
		return false
	}

	// Check for known color-supporting terminals
	if strings.Contains(termType, "color") ||
		strings.Contains(termType, "xterm") ||
		strings.Contains(termType, "256") ||
		strings.Contains(termType, "ansi") {
		return true
	}

	// Default to true if we're in a terminal
	return true
}

// Width returns the terminal width
func (t *TerminalDetector) Width() int {
	width, _, err := term.GetSize(t.fd)
	if err != nil {
		return 80 // Default width
	}
	return width
}

// IconProvider provides icons for test status
type IconProvider struct {
	unicodeSupport bool
}

// NewIconProvider creates a new IconProvider
func NewIconProvider(unicodeSupport bool) *IconProvider {
	return &IconProvider{
		unicodeSupport: unicodeSupport,
	}
}

// CheckMark returns the checkmark icon
func (i *IconProvider) CheckMark() string {
	if i.unicodeSupport {
		return IconCheckMark
	}
	return IconCheckMarkASCII
}

// Cross returns the cross icon
func (i *IconProvider) Cross() string {
	if i.unicodeSupport {
		return IconCross
	}
	return IconCrossASCII
}

// Skipped returns the skipped icon
func (i *IconProvider) Skipped() string {
	if i.unicodeSupport {
		return IconSkipped
	}
	return IconSkippedASCII
}

// Running returns the running icon
func (i *IconProvider) Running() string {
	if i.unicodeSupport {
		return IconRunning
	}
	return IconRunningASCII
}

// GetIcon returns an icon for various UI elements
func (i *IconProvider) GetIcon(iconType string) string {
	if !i.unicodeSupport {
		return i.getAsciiIcon(iconType)
	}
	return i.getUnicodeIcon(iconType)
}

func (i *IconProvider) getUnicodeIcon(iconType string) string {
	switch iconType {
	// Basic test status icons
	case "pass":
		return "✓"
	case "fail":
		return "✗"
	case "skip":
		return "⊝"
	case "running":
		return "⟳"

	// File change icons
	case "watch":
		return "👀"
	case "test":
		return "🧪"
	case "code":
		return "📝"
	case "config":
		return "⚙️"
	case "dependency":
		return "📦"
	case "file":
		return "📄"

	// Change type icons
	case "new":
		return "✨"
	case "change":
		return "🔄"
	case "unchanged":
		return "➖"

	// UI icons
	case "package":
		return "📁"
	case "summary":
		return "📊"
	case "info":
		return "ℹ️"
	case "unknown":
		return "❓"

	default:
		return "•"
	}
}

func (i *IconProvider) getAsciiIcon(iconType string) string {
	switch iconType {
	// Basic test status icons
	case "pass":
		return "[PASS]"
	case "fail":
		return "[FAIL]"
	case "skip":
		return "[SKIP]"
	case "running":
		return "[RUN ]"

	// File change icons
	case "watch":
		return "[WATCH]"
	case "test":
		return "[TEST]"
	case "code":
		return "[CODE]"
	case "config":
		return "[CONF]"
	case "dependency":
		return "[DEP ]"
	case "file":
		return "[FILE]"

	// Change type icons
	case "new":
		return "[NEW ]"
	case "change":
		return "[CHG ]"
	case "unchanged":
		return "[----]"

	// UI icons
	case "package":
		return "[PKG ]"
	case "summary":
		return "[SUM ]"
	case "info":
		return "[INFO]"
	case "unknown":
		return "[??? ]"

	default:
		return "[ • ]"
	}
}

// FormatTestStatus formats a test status with appropriate coloring and icon
func FormatTestStatus(status TestStatus, formatter *ColorFormatter, icons *IconProvider) string {
	switch status {
	case StatusPassed:
		return formatter.Green(icons.CheckMark())
	case StatusFailed:
		return formatter.Red(icons.Cross())
	case StatusSkipped:
		return formatter.Yellow(icons.Skipped())
	case StatusRunning:
		return formatter.Blue(icons.Running())
	default:
		return fmt.Sprintf("[%s]", status)
	}
}
