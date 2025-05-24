package colors

import (
	"os"
	"testing"

	"github.com/newbpydev/go-sentinel/pkg/models"
)

func TestColorFormatter(t *testing.T) {
	tests := []struct {
		name      string
		useColors bool
		text      string
		expected  string
	}{
		{
			name:      "Colors enabled",
			useColors: true,
			text:      "test",
			expected:  ColorRed + "test" + ColorReset,
		},
		{
			name:      "Colors disabled",
			useColors: false,
			text:      "test",
			expected:  "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewColorFormatter(tt.useColors)
			result := formatter.Red(tt.text)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestColorFormatterIsEnabled(t *testing.T) {
	tests := []struct {
		name      string
		useColors bool
		expected  bool
	}{
		{
			name:      "Colors enabled",
			useColors: true,
			expected:  true,
		},
		{
			name:      "Colors disabled",
			useColors: false,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewColorFormatter(tt.useColors)
			if formatter.IsEnabled() != tt.expected {
				t.Errorf("expected %t, got %t", tt.expected, formatter.IsEnabled())
			}
		})
	}
}

func TestIconProvider(t *testing.T) {
	tests := []struct {
		name           string
		unicodeSupport bool
		iconMethod     func(*IconProvider) string
		expectedUni    string
		expectedASCII  string
	}{
		{
			name:           "CheckMark",
			unicodeSupport: true,
			iconMethod:     (*IconProvider).CheckMark,
			expectedUni:    IconCheckMark,
			expectedASCII:  IconCheckMarkASCII,
		},
		{
			name:           "Cross",
			unicodeSupport: true,
			iconMethod:     (*IconProvider).Cross,
			expectedUni:    IconCross,
			expectedASCII:  IconCrossASCII,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Unicode support
			provider := NewIconProvider(true)
			result := tt.iconMethod(provider)
			if result != tt.expectedUni {
				t.Errorf("Unicode: expected %q, got %q", tt.expectedUni, result)
			}

			// Test ASCII fallback
			provider = NewIconProvider(false)
			result = tt.iconMethod(provider)
			if result != tt.expectedASCII {
				t.Errorf("ASCII: expected %q, got %q", tt.expectedASCII, result)
			}
		})
	}
}

func TestFormatTestStatus(t *testing.T) {
	formatter := NewColorFormatter(true)
	icons := NewIconProvider(true)

	tests := []struct {
		status   models.TestStatus
		expected string
	}{
		{models.StatusPassed, ColorGreen + IconCheckMark + ColorReset},
		{models.StatusFailed, ColorRed + IconCross + ColorReset},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			result := FormatTestStatus(tt.status, formatter, icons)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestTerminalDetector(t *testing.T) {
	detector := NewTerminalDetector()

	// Test that detector returns reasonable values
	if detector.Width() < 0 {
		t.Error("Width should not be negative")
	}

	// Test that methods don't panic
	_ = detector.SupportsColor()
	_ = detector.SupportsUnicode()
}

func TestTerminalDetectorNoColor(t *testing.T) {
	detector := NewTerminalDetector()

	// Test NO_COLOR environment variable
	oldNoColor := os.Getenv("NO_COLOR")
	os.Setenv("NO_COLOR", "1")
	defer func() {
		if oldNoColor == "" {
			os.Unsetenv("NO_COLOR")
		} else {
			os.Setenv("NO_COLOR", oldNoColor)
		}
	}()

	if detector.SupportsColor() {
		t.Error("Should not support colors when NO_COLOR is set")
	}
}
