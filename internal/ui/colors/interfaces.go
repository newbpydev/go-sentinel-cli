// Package colors provides color formatting and theme management interfaces
package colors

import (
	"io"
)

// ColorFormatter handles color formatting and terminal capabilities
type ColorFormatter interface {
	// Format applies color formatting to text
	Format(text string, style *Style) (string, error)

	// FormatWithColor applies a single color to text
	FormatWithColor(text string, color Color) (string, error)

	// StripColors removes color formatting from text
	StripColors(text string) string

	// SupportsColor returns whether color output is supported
	SupportsColor() bool

	// Enable enables color formatting
	Enable()

	// Disable disables color formatting
	Disable()

	// IsEnabled returns whether color formatting is enabled
	IsEnabled() bool
}

// ThemeProvider manages color themes and style definitions
type ThemeProvider interface {
	// LoadTheme loads a theme by name
	LoadTheme(name string) (*Theme, error)

	// GetTheme returns the currently active theme
	GetTheme() *Theme

	// SetTheme sets the active theme
	SetTheme(theme *Theme) error

	// ListThemes returns available theme names
	ListThemes() []string

	// CreateCustomTheme creates a new custom theme
	CreateCustomTheme(name string, definition *ThemeDefinition) (*Theme, error)

	// GetDefaultTheme returns the default theme
	GetDefaultTheme() *Theme
}

// TerminalDetector detects terminal capabilities and characteristics
type TerminalDetector interface {
	// SupportsColor detects if the terminal supports color
	SupportsColor() bool

	// Supports256Color detects if the terminal supports 256 colors
	Supports256Color() bool

	// SupportsTrueColor detects if the terminal supports true color (24-bit)
	SupportsTrueColor() bool

	// GetTerminalType returns the terminal type
	GetTerminalType() string

	// IsInteractive returns whether output is going to an interactive terminal
	IsInteractive(output io.Writer) bool

	// GetColorCapabilities returns detailed color capabilities
	GetColorCapabilities() *ColorCapabilities
}

// ColorPalette provides predefined color collections
type ColorPalette interface {
	// GetSuccessColors returns colors for success states
	GetSuccessColors() *ColorSet

	// GetErrorColors returns colors for error states
	GetErrorColors() *ColorSet

	// GetWarningColors returns colors for warning states
	GetWarningColors() *ColorSet

	// GetInfoColors returns colors for informational states
	GetInfoColors() *ColorSet

	// GetNeutralColors returns neutral colors
	GetNeutralColors() *ColorSet

	// GetColor returns a specific color by name
	GetColor(name string) (Color, bool)
}

// StyleBuilder provides fluent API for building styles
type StyleBuilder interface {
	// Foreground sets the foreground color
	Foreground(color Color) StyleBuilder

	// Background sets the background color
	Background(color Color) StyleBuilder

	// Bold applies bold formatting
	Bold() StyleBuilder

	// Italic applies italic formatting
	Italic() StyleBuilder

	// Underline applies underline formatting
	Underline() StyleBuilder

	// Strikethrough applies strikethrough formatting
	Strikethrough() StyleBuilder

	// Dim applies dim formatting
	Dim() StyleBuilder

	// Reverse applies reverse video formatting
	Reverse() StyleBuilder

	// Build creates the final style
	Build() *Style

	// Reset resets the builder to default state
	Reset() StyleBuilder
}

// Style represents a complete text style including colors and formatting
type Style struct {
	// Foreground is the text color
	Foreground Color

	// Background is the background color
	Background Color

	// Bold indicates bold formatting
	Bold bool

	// Italic indicates italic formatting
	Italic bool

	// Underline indicates underline formatting
	Underline bool

	// Strikethrough indicates strikethrough formatting
	Strikethrough bool

	// Dim indicates dim formatting
	Dim bool

	// Reverse indicates reverse video formatting
	Reverse bool
}

// Color represents a color value
type Color struct {
	// Type is the color type (basic, extended, rgb)
	Type ColorType

	// Value is the color value (depends on type)
	Value interface{}

	// Name is the optional color name
	Name string
}

// Theme represents a complete color theme
type Theme struct {
	// Name is the theme name
	Name string

	// Description is the theme description
	Description string

	// Colors contains named color definitions
	Colors map[string]Color

	// Styles contains named style definitions
	Styles map[string]*Style

	// Metadata contains additional theme metadata
	Metadata map[string]interface{}
}

// ThemeDefinition represents the definition for creating a theme
type ThemeDefinition struct {
	// Name is the theme name
	Name string

	// Description is the theme description
	Description string

	// BaseTheme is the base theme to inherit from
	BaseTheme string

	// ColorOverrides contains color overrides
	ColorOverrides map[string]Color

	// StyleOverrides contains style overrides
	StyleOverrides map[string]*Style
}

// ColorSet represents a collection of related colors
type ColorSet struct {
	// Primary is the primary color
	Primary Color

	// Secondary is the secondary color
	Secondary Color

	// Light is the light variant
	Light Color

	// Dark is the dark variant
	Dark Color

	// Accent is the accent color
	Accent Color
}

// ColorCapabilities represents terminal color capabilities
type ColorCapabilities struct {
	// BasicColors indicates support for basic 8/16 colors
	BasicColors bool

	// ExtendedColors indicates support for 256 colors
	ExtendedColors bool

	// TrueColor indicates support for 24-bit true color
	TrueColor bool

	// ColorCount is the maximum number of supported colors
	ColorCount int

	// TerminalType is the detected terminal type
	TerminalType string
}

// ColorType represents the type of color representation
type ColorType string

const (
	// ColorTypeBasic represents basic ANSI colors (0-15)
	ColorTypeBasic ColorType = "basic"

	// ColorTypeExtended represents extended colors (0-255)
	ColorTypeExtended ColorType = "extended"

	// ColorTypeRGB represents RGB true color
	ColorTypeRGB ColorType = "rgb"

	// ColorTypeHex represents hexadecimal color notation
	ColorTypeHex ColorType = "hex"

	// ColorTypeNamed represents named colors
	ColorTypeNamed ColorType = "named"
)

// Basic ANSI color constants - using variables since Go doesn't support struct constants
var (
	// Basic colors
	Black   = Color{Type: ColorTypeBasic, Value: 0, Name: "black"}
	Red     = Color{Type: ColorTypeBasic, Value: 1, Name: "red"}
	Green   = Color{Type: ColorTypeBasic, Value: 2, Name: "green"}
	Yellow  = Color{Type: ColorTypeBasic, Value: 3, Name: "yellow"}
	Blue    = Color{Type: ColorTypeBasic, Value: 4, Name: "blue"}
	Magenta = Color{Type: ColorTypeBasic, Value: 5, Name: "magenta"}
	Cyan    = Color{Type: ColorTypeBasic, Value: 6, Name: "cyan"}
	White   = Color{Type: ColorTypeBasic, Value: 7, Name: "white"}

	// Bright colors
	BrightBlack   = Color{Type: ColorTypeBasic, Value: 8, Name: "bright_black"}
	BrightRed     = Color{Type: ColorTypeBasic, Value: 9, Name: "bright_red"}
	BrightGreen   = Color{Type: ColorTypeBasic, Value: 10, Name: "bright_green"}
	BrightYellow  = Color{Type: ColorTypeBasic, Value: 11, Name: "bright_yellow"}
	BrightBlue    = Color{Type: ColorTypeBasic, Value: 12, Name: "bright_blue"}
	BrightMagenta = Color{Type: ColorTypeBasic, Value: 13, Name: "bright_magenta"}
	BrightCyan    = Color{Type: ColorTypeBasic, Value: 14, Name: "bright_cyan"}
	BrightWhite   = Color{Type: ColorTypeBasic, Value: 15, Name: "bright_white"}
)

// Semantic color mappings for test states
var (
	// SuccessColor is the default success color
	SuccessColor = Green

	// ErrorColor is the default error color
	ErrorColor = Red

	// WarningColor is the default warning color
	WarningColor = Yellow

	// InfoColor is the default info color
	InfoColor = Blue

	// SkipColor is the default skip color
	SkipColor = Yellow

	// PassedColor is the default passed test color
	PassedColor = Green

	// FailedColor is the default failed test color
	FailedColor = Red

	// RunningColor is the default running test color
	RunningColor = Blue
)

// PresetStyles provides common predefined styles
var (
	// DefaultStyle is the default text style
	DefaultStyle = &Style{}

	// BoldStyle applies bold formatting
	BoldStyle = &Style{Bold: true}

	// ItalicStyle applies italic formatting
	ItalicStyle = &Style{Italic: true}

	// UnderlineStyle applies underline formatting
	UnderlineStyle = &Style{Underline: true}

	// DimStyle applies dim formatting
	DimStyle = &Style{Dim: true}

	// SuccessStyle for success messages
	SuccessStyle = &Style{Foreground: SuccessColor, Bold: true}

	// ErrorStyle for error messages
	ErrorStyle = &Style{Foreground: ErrorColor, Bold: true}

	// WarningStyle for warning messages
	WarningStyle = &Style{Foreground: WarningColor, Bold: true}

	// InfoStyle for info messages
	InfoStyle = &Style{Foreground: InfoColor}
)

// RGB represents an RGB color value
type RGB struct {
	// R is the red component (0-255)
	R uint8

	// G is the green component (0-255)
	G uint8

	// B is the blue component (0-255)
	B uint8
}

// NewRGBColor creates a new RGB color
func NewRGBColor(r, g, b uint8) Color {
	return Color{
		Type:  ColorTypeRGB,
		Value: RGB{R: r, G: g, B: b},
		Name:  "",
	}
}

// NewHexColor creates a new color from hex string
func NewHexColor(hex string) Color {
	return Color{
		Type:  ColorTypeHex,
		Value: hex,
		Name:  "",
	}
}

// NewExtendedColor creates a new extended color (0-255)
func NewExtendedColor(value uint8) Color {
	return Color{
		Type:  ColorTypeExtended,
		Value: value,
		Name:  "",
	}
}

// NewNamedColor creates a new named color
func NewNamedColor(name string) Color {
	return Color{
		Type:  ColorTypeNamed,
		Value: nil,
		Name:  name,
	}
}
