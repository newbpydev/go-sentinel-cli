// Package colors provides reference theme implementation with exact hex color matching
package colors

import (
	"fmt"
	"strconv"
	"strings"
)

// ReferenceTheme implements exact color matching with specified hex codes
// from the visual guidelines: #10b981 (green), #ef4444 (red), #f59e0b (amber)
type ReferenceTheme struct {
	name         string
	supportsTrue bool // True color (24-bit) support
	supports256  bool // 256-color support
	detector     DetectorInterface
}

// NewReferenceTheme creates a new reference theme with terminal detection
func NewReferenceTheme(detector DetectorInterface) *ReferenceTheme {
	return &ReferenceTheme{
		name:         "reference",
		supportsTrue: detector.SupportsTrueColor(),
		supports256:  detector.Supports256Color(),
		detector:     detector,
	}
}

// HexColor represents a color with its hex value and fallbacks
type HexColor struct {
	Hex     string // Hex color code (e.g., "#10b981")
	RGB     RGB    // RGB values extracted from hex
	ANSI256 int    // 256-color fallback
	ANSI16  string // 16-color fallback
	Name    string // Color name for debugging
}

// RGB represents RGB color values
type RGB struct {
	R, G, B int
}

// Vitest reference colors with exact hex codes from guidelines
var (
	// Success color: #10b981 (green)
	ColorSuccess = HexColor{
		Hex:     "#10b981",
		RGB:     RGB{R: 16, G: 185, B: 129},
		ANSI256: 48,   // Closest 256-color match
		ANSI16:  "32", // Green
		Name:    "success",
	}

	// Error color: #ef4444 (red)
	ColorError = HexColor{
		Hex:     "#ef4444",
		RGB:     RGB{R: 239, G: 68, B: 68},
		ANSI256: 203,  // Closest 256-color match
		ANSI16:  "31", // Red
		Name:    "error",
	}

	// Warning/Skip color: #f59e0b (amber)
	ColorWarning = HexColor{
		Hex:     "#f59e0b",
		RGB:     RGB{R: 245, G: 158, B: 11},
		ANSI256: 214,  // Closest 256-color match
		ANSI16:  "33", // Yellow
		Name:    "warning",
	}

	// Info color: #6b7280 (gray)
	ColorInfo = HexColor{
		Hex:     "#6b7280",
		RGB:     RGB{R: 107, G: 114, B: 128},
		ANSI256: 243,  // Closest 256-color match
		ANSI16:  "90", // Bright black (gray)
		Name:    "info",
	}

	// Muted color: #9ca3af (light gray)
	ColorMuted = HexColor{
		Hex:     "#9ca3af",
		RGB:     RGB{R: 156, G: 163, B: 175},
		ANSI256: 250,  // Closest 256-color match
		ANSI16:  "37", // White
		Name:    "muted",
	}
)

// GetColorCode returns the appropriate color code based on terminal capabilities
func (t *ReferenceTheme) GetColorCode(color HexColor) string {
	// Use true color if supported (24-bit)
	if t.supportsTrue {
		return fmt.Sprintf("\033[38;2;%d;%d;%dm", color.RGB.R, color.RGB.G, color.RGB.B)
	}

	// Use 256-color if supported
	if t.supports256 {
		return fmt.Sprintf("\033[38;5;%dm", color.ANSI256)
	}

	// Fall back to 16-color
	return fmt.Sprintf("\033[%sm", color.ANSI16)
}

// FormatSuccess formats text with exact success color (#10b981)
func (t *ReferenceTheme) FormatSuccess(text string) string {
	if !t.detector.SupportsColor() {
		return text
	}
	return t.GetColorCode(ColorSuccess) + text + "\033[0m"
}

// FormatError formats text with exact error color (#ef4444)
func (t *ReferenceTheme) FormatError(text string) string {
	if !t.detector.SupportsColor() {
		return text
	}
	return t.GetColorCode(ColorError) + text + "\033[0m"
}

// FormatWarning formats text with exact warning color (#f59e0b)
func (t *ReferenceTheme) FormatWarning(text string) string {
	if !t.detector.SupportsColor() {
		return text
	}
	return t.GetColorCode(ColorWarning) + text + "\033[0m"
}

// FormatInfo formats text with exact info color (#6b7280)
func (t *ReferenceTheme) FormatInfo(text string) string {
	if !t.detector.SupportsColor() {
		return text
	}
	return t.GetColorCode(ColorInfo) + text + "\033[0m"
}

// FormatMuted formats text with exact muted color (#9ca3af)
func (t *ReferenceTheme) FormatMuted(text string) string {
	if !t.detector.SupportsColor() {
		return text
	}
	return t.GetColorCode(ColorMuted) + text + "\033[0m"
}

// FormatWithHex formats text with a custom hex color
func (t *ReferenceTheme) FormatWithHex(text, hexColor string) string {
	if !t.detector.SupportsColor() {
		return text
	}

	color, err := ParseHexColor(hexColor)
	if err != nil {
		// Fall back to no color on parse error
		return text
	}

	return t.GetColorCode(color) + text + "\033[0m"
}

// ParseHexColor parses a hex color string into a HexColor struct
func ParseHexColor(hex string) (HexColor, error) {
	// Remove # prefix if present
	hex = strings.TrimPrefix(hex, "#")

	// Validate hex length
	if len(hex) != 6 {
		return HexColor{}, fmt.Errorf("invalid hex color format: %s", hex)
	}

	// Parse RGB components
	r, err := strconv.ParseInt(hex[0:2], 16, 0)
	if err != nil {
		return HexColor{}, fmt.Errorf("invalid red component: %s", hex[0:2])
	}

	g, err := strconv.ParseInt(hex[2:4], 16, 0)
	if err != nil {
		return HexColor{}, fmt.Errorf("invalid green component: %s", hex[2:4])
	}

	b, err := strconv.ParseInt(hex[4:6], 16, 0)
	if err != nil {
		return HexColor{}, fmt.Errorf("invalid blue component: %s", hex[4:6])
	}

	return HexColor{
		Hex:     "#" + hex,
		RGB:     RGB{R: int(r), G: int(g), B: int(b)},
		ANSI256: rgbTo256Color(int(r), int(g), int(b)),
		ANSI16:  rgbTo16Color(int(r), int(g), int(b)),
		Name:    "custom",
	}, nil
}

// rgbTo256Color converts RGB to closest 256-color code
func rgbTo256Color(r, g, b int) int {
	// Simplified conversion - in practice, you'd use a more sophisticated algorithm
	// or lookup table for better color matching
	if r == g && g == b {
		// Grayscale
		if r < 8 {
			return 16
		}
		if r > 248 {
			return 231
		}
		return int(232 + (r-8)/10)
	}

	// Color cube
	r6 := r * 5 / 255
	g6 := g * 5 / 255
	b6 := b * 5 / 255
	return 16 + 36*r6 + 6*g6 + b6
}

// rgbTo16Color converts RGB to closest 16-color ANSI code
func rgbTo16Color(r, g, b int) string {
	// Simple brightness-based fallback
	brightness := (r + g + b) / 3

	// Determine dominant color
	max := r
	color := "31" // red
	if g > max {
		max = g
		color = "32" // green
	}
	if b > max {
		max = b
		color = "34" // blue
	}

	// Use bright variant if brightness is high
	if brightness > 128 {
		switch color {
		case "31":
			return "91" // bright red
		case "32":
			return "92" // bright green
		case "34":
			return "94" // bright blue
		}
	}

	return color
}

// GetName returns the theme name
func (t *ReferenceTheme) GetName() string {
	return t.name
}

// SupportsColors returns whether colors are supported
func (t *ReferenceTheme) SupportsColors() bool {
	return t.detector.SupportsColor()
}
