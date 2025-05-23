package rendering

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// ANSI color codes - EXACT COPY from original
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

// Vitest-inspired icons - EXACT COPY from original
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

// ColorFormatter formats text with ANSI colors - EXACT COPY from original
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

// TerminalDetector detects terminal capabilities - EXACT COPY from original
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
	colorTerms := []string{
		"xterm", "xterm-256color", "screen", "screen-256color",
		"tmux", "tmux-256color", "rxvt", "ansi", "cygwin",
	}

	for _, colorTerm := range colorTerms {
		if strings.Contains(termType, colorTerm) {
			return true
		}
	}

	return false
}

// Width returns the terminal width
func (t *TerminalDetector) Width() int {
	width, _, err := term.GetSize(t.fd)
	if err != nil {
		return 80 // Default width
	}
	return width
}

// IconProvider provides icons for different contexts - EXACT COPY from original
type IconProvider struct {
	unicodeSupport bool
}

// NewIconProvider creates a new IconProvider
func NewIconProvider(unicodeSupport bool) *IconProvider {
	return &IconProvider{
		unicodeSupport: unicodeSupport,
	}
}

// CheckMark returns a checkmark icon
func (i *IconProvider) CheckMark() string {
	if i.unicodeSupport {
		return IconCheckMark
	}
	return IconCheckMarkASCII
}

// Cross returns a cross/X icon
func (i *IconProvider) Cross() string {
	if i.unicodeSupport {
		return IconCross
	}
	return IconCrossASCII
}

// Skipped returns a skipped icon
func (i *IconProvider) Skipped() string {
	if i.unicodeSupport {
		return IconSkipped
	}
	return IconSkippedASCII
}

// Running returns a running/spinning icon
func (i *IconProvider) Running() string {
	if i.unicodeSupport {
		return IconRunning
	}
	return IconRunningASCII
}

// GetIcon returns an icon by type name
func (i *IconProvider) GetIcon(iconType string) string {
	if i.unicodeSupport {
		return i.getUnicodeIcon(iconType)
	}
	return i.getAsciiIcon(iconType)
}

// getUnicodeIcon returns Unicode icons for various types
func (i *IconProvider) getUnicodeIcon(iconType string) string {
	switch iconType {
	case "checkmark", "pass", "passed", "success":
		return "✓"
	case "cross", "fail", "failed", "error":
		return "✗"
	case "skip", "skipped", "pending":
		return "⃠"
	case "running", "spin":
		return "⟳"
	case "package", "pkg":
		return "📦"
	case "file", "files":
		return "📁"
	case "watch", "eye":
		return "👀"
	case "rocket", "launch":
		return "🚀"
	case "lightning", "fast", "optimized":
		return "⚡"
	case "clock", "time", "timer":
		return "⏱️"
	case "info", "information":
		return "ℹ️"
	case "warning", "warn":
		return "⚠️"
	case "new", "plus":
		return "+"
	case "modified", "change":
		return "~"
	case "deleted", "remove":
		return "-"
	default:
		return "•"
	}
}

// getAsciiIcon returns ASCII fallback icons
func (i *IconProvider) getAsciiIcon(iconType string) string {
	switch iconType {
	case "checkmark", "pass", "passed", "success":
		return "√"
	case "cross", "fail", "failed", "error":
		return "×"
	case "skip", "skipped", "pending":
		return "○"
	case "running", "spin":
		return "~"
	case "package", "pkg":
		return "[PKG]"
	case "file", "files":
		return "[FILE]"
	case "watch", "eye":
		return "[WATCH]"
	case "rocket", "launch":
		return "[RUN]"
	case "lightning", "fast", "optimized":
		return "[OPT]"
	case "clock", "time", "timer":
		return "[TIME]"
	case "info", "information":
		return "[INFO]"
	case "warning", "warn":
		return "[WARN]"
	case "new", "plus":
		return "+"
	case "modified", "change":
		return "~"
	case "deleted", "remove":
		return "-"
	default:
		return "*"
	}
}

// Legacy compatibility types from original types.go
type TestStatus string

const (
	StatusPassed  TestStatus = "passed"
	StatusFailed  TestStatus = "failed"
	StatusSkipped TestStatus = "skipped"
	StatusRunning TestStatus = "running"
)

// FormatTestStatus formats a test status with appropriate icon and color
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
		return formatter.Gray("?")
	}
}
