// Package colors provides tests for the reference theme implementation
package colors

import (
	"strings"
	"testing"
)

// MockDetector implements DetectorInterface for testing
type MockDetector struct {
	supportsColor     bool
	supportsUnicode   bool
	supportsTrueColor bool
	supports256Color  bool
	width             int
}

func (m *MockDetector) SupportsColor() bool     { return m.supportsColor }
func (m *MockDetector) SupportsUnicode() bool   { return m.supportsUnicode }
func (m *MockDetector) SupportsTrueColor() bool { return m.supportsTrueColor }
func (m *MockDetector) Supports256Color() bool  { return m.supports256Color }
func (m *MockDetector) Width() int              { return m.width }

func TestNewReferenceTheme(t *testing.T) {
	detector := &MockDetector{
		supportsColor:     true,
		supportsUnicode:   true,
		supportsTrueColor: true,
		supports256Color:  true,
		width:             80,
	}

	theme := NewReferenceTheme(detector)

	if theme.name != "reference" {
		t.Errorf("Expected theme name 'reference', got '%s'", theme.name)
	}

	if !theme.supportsTrue {
		t.Error("Expected theme to support true color")
	}

	if !theme.supports256 {
		t.Error("Expected theme to support 256 color")
	}
}

func TestReferenceTheme_GetColorCode(t *testing.T) {
	tests := []struct {
		name              string
		supportsTrueColor bool
		supports256Color  bool
		color             HexColor
		expectedPattern   string
	}{
		{
			name:              "true color support",
			supportsTrueColor: true,
			supports256Color:  true,
			color:             ColorSuccess,
			expectedPattern:   "\033[38;2;16;185;129m",
		},
		{
			name:              "256 color support",
			supportsTrueColor: false,
			supports256Color:  true,
			color:             ColorSuccess,
			expectedPattern:   "\033[38;5;48m",
		},
		{
			name:              "16 color fallback",
			supportsTrueColor: false,
			supports256Color:  false,
			color:             ColorSuccess,
			expectedPattern:   "\033[32m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := &MockDetector{
				supportsColor:     true,
				supportsTrueColor: tt.supportsTrueColor,
				supports256Color:  tt.supports256Color,
			}

			theme := NewReferenceTheme(detector)
			result := theme.GetColorCode(tt.color)

			if result != tt.expectedPattern {
				t.Errorf("Expected color code '%s', got '%s'", tt.expectedPattern, result)
			}
		})
	}
}

func TestReferenceTheme_FormatSuccess(t *testing.T) {
	tests := []struct {
		name          string
		supportsColor bool
		text          string
		expectColor   bool
	}{
		{
			name:          "with color support",
			supportsColor: true,
			text:          "Test passed",
			expectColor:   true,
		},
		{
			name:          "without color support",
			supportsColor: false,
			text:          "Test passed",
			expectColor:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := &MockDetector{
				supportsColor:     tt.supportsColor,
				supportsTrueColor: true,
				supports256Color:  true,
			}

			theme := NewReferenceTheme(detector)
			result := theme.FormatSuccess(tt.text)

			if tt.expectColor {
				if !strings.Contains(result, "\033[38;2;16;185;129m") {
					t.Error("Expected colored output")
				}
				if !strings.Contains(result, "\033[0m") {
					t.Error("Expected reset sequence")
				}
				if !strings.Contains(result, tt.text) {
					t.Error("Expected text to be preserved")
				}
			} else {
				if result != tt.text {
					t.Errorf("Expected plain text '%s', got '%s'", tt.text, result)
				}
			}
		})
	}
}

func TestReferenceTheme_FormatError(t *testing.T) {
	detector := &MockDetector{
		supportsColor:     true,
		supportsTrueColor: true,
		supports256Color:  true,
	}

	theme := NewReferenceTheme(detector)
	result := theme.FormatError("Test failed")

	// Should contain exact error color
	if !strings.Contains(result, "\033[38;2;239;68;68m") {
		t.Error("Expected exact error color (#ef4444)")
	}

	if !strings.Contains(result, "Test failed") {
		t.Error("Expected text to be preserved")
	}

	if !strings.Contains(result, "\033[0m") {
		t.Error("Expected reset sequence")
	}
}

func TestReferenceTheme_FormatWarning(t *testing.T) {
	detector := &MockDetector{
		supportsColor:     true,
		supportsTrueColor: true,
		supports256Color:  true,
	}

	theme := NewReferenceTheme(detector)
	result := theme.FormatWarning("Test skipped")

	// Should contain exact warning color
	if !strings.Contains(result, "\033[38;2;245;158;11m") {
		t.Error("Expected exact warning color (#f59e0b)")
	}

	if !strings.Contains(result, "Test skipped") {
		t.Error("Expected text to be preserved")
	}
}

func TestParseHexColor(t *testing.T) {
	tests := []struct {
		name        string
		hex         string
		expectedRGB RGB
		expectError bool
	}{
		{
			name:        "valid hex with #",
			hex:         "#10b981",
			expectedRGB: RGB{R: 16, G: 185, B: 129},
			expectError: false,
		},
		{
			name:        "valid hex without #",
			hex:         "ef4444",
			expectedRGB: RGB{R: 239, G: 68, B: 68},
			expectError: false,
		},
		{
			name:        "invalid hex length",
			hex:         "#fff",
			expectError: true,
		},
		{
			name:        "invalid hex characters",
			hex:         "#gggggg",
			expectError: true,
		},
		{
			name:        "empty string",
			hex:         "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseHexColor(tt.hex)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result.RGB != tt.expectedRGB {
				t.Errorf("Expected RGB %+v, got %+v", tt.expectedRGB, result.RGB)
			}

			// Check if hex is properly formatted
			expectedHex := tt.hex
			if !strings.HasPrefix(expectedHex, "#") {
				expectedHex = "#" + expectedHex
			}
			if result.Hex != expectedHex {
				t.Errorf("Expected hex '%s', got '%s'", expectedHex, result.Hex)
			}
		})
	}
}

func TestReferenceColors(t *testing.T) {
	// Test the predefined reference colors
	tests := []struct {
		name  string
		color HexColor
	}{
		{"ColorSuccess", ColorSuccess},
		{"ColorError", ColorError},
		{"ColorWarning", ColorWarning},
		{"ColorInfo", ColorInfo},
		{"ColorMuted", ColorMuted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify hex format
			if !strings.HasPrefix(tt.color.Hex, "#") {
				t.Errorf("Expected hex to start with #, got '%s'", tt.color.Hex)
			}

			if len(tt.color.Hex) != 7 {
				t.Errorf("Expected hex length 7, got %d", len(tt.color.Hex))
			}

			// Verify RGB values are within range
			if tt.color.RGB.R < 0 || tt.color.RGB.R > 255 {
				t.Errorf("RGB.R out of range: %d", tt.color.RGB.R)
			}
			if tt.color.RGB.G < 0 || tt.color.RGB.G > 255 {
				t.Errorf("RGB.G out of range: %d", tt.color.RGB.G)
			}
			if tt.color.RGB.B < 0 || tt.color.RGB.B > 255 {
				t.Errorf("RGB.B out of range: %d", tt.color.RGB.B)
			}

			// Verify ANSI256 is valid
			if tt.color.ANSI256 < 0 || tt.color.ANSI256 > 255 {
				t.Errorf("ANSI256 out of range: %d", tt.color.ANSI256)
			}

			// Verify name is set
			if tt.color.Name == "" {
				t.Error("Color name should not be empty")
			}
		})
	}
}

func TestReferenceTheme_FormatWithHex(t *testing.T) {
	detector := &MockDetector{
		supportsColor:     true,
		supportsTrueColor: true,
		supports256Color:  true,
	}

	theme := NewReferenceTheme(detector)

	// Test with valid hex
	result := theme.FormatWithHex("Custom color", "#ff5500")
	if !strings.Contains(result, "\033[38;2;255;85;0m") {
		t.Error("Expected custom hex color")
	}

	// Test with invalid hex (should return plain text)
	result = theme.FormatWithHex("Custom color", "invalid")
	if result != "Custom color" {
		t.Error("Expected plain text for invalid hex")
	}
}

func TestReferenceTheme_GetName(t *testing.T) {
	detector := &MockDetector{}
	theme := NewReferenceTheme(detector)

	if theme.GetName() != "reference" {
		t.Errorf("Expected name 'reference', got '%s'", theme.GetName())
	}
}

func TestReferenceTheme_SupportsColors(t *testing.T) {
	tests := []struct {
		name          string
		supportsColor bool
	}{
		{"with color support", true},
		{"without color support", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := &MockDetector{
				supportsColor: tt.supportsColor,
			}
			theme := NewReferenceTheme(detector)

			if theme.SupportsColors() != tt.supportsColor {
				t.Errorf("Expected SupportsColors() to return %v", tt.supportsColor)
			}
		})
	}
}
