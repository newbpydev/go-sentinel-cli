// Package colors provides terminal color formatting and detection capabilities
package colors

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
	ColorWhite   = "\033[37m"

	// Background colors
	ColorBgRed     = "\033[41m"
	ColorBgGreen   = "\033[42m"
	ColorBgYellow  = "\033[43m"
	ColorBgBlue    = "\033[44m"
	ColorBgMagenta = "\033[45m"
	ColorBgCyan    = "\033[46m"
)

// FormatterInterface defines the interface for color formatting
type FormatterInterface interface {
	// Color methods
	Red(text string) string
	Green(text string) string
	Yellow(text string) string
	Blue(text string) string
	Magenta(text string) string
	Cyan(text string) string
	Gray(text string) string
	White(text string) string

	// Style methods
	Bold(text string) string
	Dim(text string) string

	// Background methods
	BgRed(text string) string

	// Generic colorize method
	Colorize(text, colorName string) string

	// Check if colors are enabled
	IsEnabled() bool
}

// ColorFormatter formats text with ANSI colors and implements FormatterInterface
type ColorFormatter struct {
	useColors bool
}

// NewColorFormatter creates a new ColorFormatter
func NewColorFormatter(useColors bool) *ColorFormatter {
	return &ColorFormatter{
		useColors: useColors,
	}
}

// NewAutoColorFormatter creates a ColorFormatter with automatic color detection
func NewAutoColorFormatter() *ColorFormatter {
	detector := NewTerminalDetector()
	return &ColorFormatter{
		useColors: detector.SupportsColor(),
	}
}

// IsEnabled returns whether colors are enabled
func (f *ColorFormatter) IsEnabled() bool {
	return f.useColors
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

// White formats text with white color
func (f *ColorFormatter) White(text string) string {
	return f.colorize(text, ColorWhite)
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
	return fmt.Sprintf("%s%s%s", ColorBgRed, text, ColorReset)
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

// DetectorInterface defines the interface for terminal detection
type DetectorInterface interface {
	SupportsColor() bool
	SupportsUnicode() bool
	SupportsTrueColor() bool
	Supports256Color() bool
	Width() int
}

// TerminalDetector detects terminal capabilities and implements DetectorInterface
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
		"tmux", "tmux-256color", "rxvt", "color", "ansi",
	}

	for _, colorTerm := range colorTerms {
		if strings.Contains(termType, colorTerm) {
			return true
		}
	}

	// Check for Windows Terminal and ConEmu
	if os.Getenv("WT_SESSION") != "" || os.Getenv("ConEmuANSI") == "ON" {
		return true
	}

	// Default to false for unknown terminals
	return false
}

// SupportsUnicode detects if the terminal supports Unicode characters
func (t *TerminalDetector) SupportsUnicode() bool {
	// Check environment variables that indicate Unicode support
	lang := os.Getenv("LANG")
	lcAll := os.Getenv("LC_ALL")

	// Look for UTF-8 encoding
	if strings.Contains(strings.ToLower(lang), "utf-8") ||
		strings.Contains(strings.ToLower(lcAll), "utf-8") {
		return true
	}

	// Windows Terminal and modern terminals generally support Unicode
	if os.Getenv("WT_SESSION") != "" {
		return true
	}

	// Check terminal type for Unicode support
	termType := os.Getenv("TERM")
	unicodeTerms := []string{"xterm", "screen", "tmux"}
	for _, unicodeTerm := range unicodeTerms {
		if strings.Contains(termType, unicodeTerm) {
			return true
		}
	}

	// Default to false for safety
	return false
}

// Width returns the terminal width in columns
func (t *TerminalDetector) Width() int {
	width, _, err := term.GetSize(t.fd)
	if err != nil {
		return 80 // Default width
	}
	return width
}

// SupportsTrueColor detects if the terminal supports 24-bit true color
func (t *TerminalDetector) SupportsTrueColor() bool {
	// Check if colors are supported at all
	if !t.SupportsColor() {
		return false
	}

	// Check for true color support via environment variables
	colorTerm := os.Getenv("COLORTERM")
	if colorTerm == "truecolor" || colorTerm == "24bit" {
		return true
	}

	// Check terminal capabilities
	termType := os.Getenv("TERM")
	trueColorTerms := []string{
		"xterm-direct", "tmux-direct", "screen-direct",
	}

	for _, term := range trueColorTerms {
		if strings.Contains(termType, term) {
			return true
		}
	}

	// Modern terminals that support true color
	if os.Getenv("WT_SESSION") != "" || // Windows Terminal
		os.Getenv("ITERM_SESSION_ID") != "" || // iTerm2
		os.Getenv("KITTY_WINDOW_ID") != "" { // Kitty
		return true
	}

	// Check for modern terminal versions
	if strings.Contains(termType, "256color") {
		// Many modern 256-color terminals also support true color
		return true
	}

	return false
}

// Supports256Color detects if the terminal supports 256-color palette
func (t *TerminalDetector) Supports256Color() bool {
	// Check if colors are supported at all
	if !t.SupportsColor() {
		return false
	}

	// Check terminal type
	termType := os.Getenv("TERM")
	color256Terms := []string{
		"256color", "xterm-256", "screen-256", "tmux-256",
	}

	for _, term := range color256Terms {
		if strings.Contains(termType, term) {
			return true
		}
	}

	// Modern terminals typically support 256 colors
	if os.Getenv("WT_SESSION") != "" || // Windows Terminal
		os.Getenv("ITERM_SESSION_ID") != "" || // iTerm2
		os.Getenv("KITTY_WINDOW_ID") != "" { // Kitty
		return true
	}

	// Fall back to true color support check
	return t.SupportsTrueColor()
}

// Ensure ColorFormatter implements FormatterInterface
var _ FormatterInterface = (*ColorFormatter)(nil)

// Ensure TerminalDetector implements DetectorInterface
var _ DetectorInterface = (*TerminalDetector)(nil)
