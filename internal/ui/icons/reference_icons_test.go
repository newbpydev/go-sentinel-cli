// Package icons provides tests for the reference icons implementation
package icons

import (
	"strings"
	"testing"
)

// MockTerminalCapabilityDetector implements TerminalCapabilityDetector for testing
type MockTerminalCapabilityDetector struct {
	supportsUnicode   bool
	supportsEmoji     bool
	supportsNerdFonts bool
	characterWidth    int
	capabilities      *TerminalCapabilities
}

func (m *MockTerminalCapabilityDetector) SupportsUnicode() bool             { return m.supportsUnicode }
func (m *MockTerminalCapabilityDetector) SupportsEmoji() bool               { return m.supportsEmoji }
func (m *MockTerminalCapabilityDetector) SupportsNerdFonts() bool           { return m.supportsNerdFonts }
func (m *MockTerminalCapabilityDetector) GetCharacterWidth(char string) int { return m.characterWidth }
func (m *MockTerminalCapabilityDetector) GetCapabilities() *TerminalCapabilities {
	return m.capabilities
}

func TestNewReferenceIcons(t *testing.T) {
	detector := &MockTerminalCapabilityDetector{
		supportsUnicode:   true,
		supportsEmoji:     false,
		supportsNerdFonts: false,
	}

	icons := NewReferenceIcons(detector)

	if icons.name != "reference" {
		t.Errorf("Expected name 'reference', got '%s'", icons.name)
	}

	if !icons.supportsUnicode {
		t.Error("Expected Unicode support")
	}

	if icons.unicodeProvider == nil {
		t.Error("Expected Unicode provider to be initialized")
	}

	if icons.asciiProvider == nil {
		t.Error("Expected ASCII provider to be initialized")
	}
}

func TestReferenceIcons_GetIcon_Unicode(t *testing.T) {
	detector := &MockTerminalCapabilityDetector{
		supportsUnicode: true,
	}

	icons := NewReferenceIcons(detector)

	tests := []struct {
		name         string
		iconName     string
		expectedIcon string
		shouldExist  bool
	}{
		{
			name:         "test passed icon",
			iconName:     "test_passed",
			expectedIcon: "✓",
			shouldExist:  true,
		},
		{
			name:         "test failed icon",
			iconName:     "test_failed",
			expectedIcon: "✗",
			shouldExist:  true,
		},
		{
			name:         "test skipped icon",
			iconName:     "test_skipped",
			expectedIcon: "⃠",
			shouldExist:  true,
		},
		{
			name:         "arrow right icon",
			iconName:     "arrow_right",
			expectedIcon: "→",
			shouldExist:  true,
		},
		{
			name:         "arrow down right icon",
			iconName:     "arrow_down_right",
			expectedIcon: "↳",
			shouldExist:  true,
		},
		{
			name:         "timer icon",
			iconName:     "timer",
			expectedIcon: "⏱️",
			shouldExist:  true,
		},
		{
			name:         "nonexistent icon",
			iconName:     "nonexistent",
			expectedIcon: "",
			shouldExist:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			icon, exists := icons.GetIcon(tt.iconName)

			if exists != tt.shouldExist {
				t.Errorf("Expected exists=%v, got %v", tt.shouldExist, exists)
			}

			if tt.shouldExist && icon != tt.expectedIcon {
				t.Errorf("Expected icon '%s', got '%s'", tt.expectedIcon, icon)
			}
		})
	}
}

func TestReferenceIcons_GetIcon_ASCII(t *testing.T) {
	detector := &MockTerminalCapabilityDetector{
		supportsUnicode: false, // Force ASCII mode
	}

	icons := NewReferenceIcons(detector)

	tests := []struct {
		name         string
		iconName     string
		expectedIcon string
		shouldExist  bool
	}{
		{
			name:         "test passed icon ASCII",
			iconName:     "test_passed",
			expectedIcon: "[P]",
			shouldExist:  true,
		},
		{
			name:         "test failed icon ASCII",
			iconName:     "test_failed",
			expectedIcon: "[F]",
			shouldExist:  true,
		},
		{
			name:         "test skipped icon ASCII",
			iconName:     "test_skipped",
			expectedIcon: "[S]",
			shouldExist:  true,
		},
		{
			name:         "arrow right icon ASCII",
			iconName:     "arrow_right",
			expectedIcon: "->",
			shouldExist:  true,
		},
		{
			name:         "timer icon ASCII",
			iconName:     "timer",
			expectedIcon: "T",
			shouldExist:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			icon, exists := icons.GetIcon(tt.iconName)

			if exists != tt.shouldExist {
				t.Errorf("Expected exists=%v, got %v", tt.shouldExist, exists)
			}

			if tt.shouldExist && icon != tt.expectedIcon {
				t.Errorf("Expected ASCII icon '%s', got '%s'", tt.expectedIcon, icon)
			}
		})
	}
}

func TestReferenceIcons_GetTestStatusIcon(t *testing.T) {
	detector := &MockTerminalCapabilityDetector{
		supportsUnicode: true,
	}

	icons := NewReferenceIcons(detector)

	tests := []struct {
		name         string
		status       string
		expectedIcon string
	}{
		{
			name:         "passed status",
			status:       "passed",
			expectedIcon: "✓",
		},
		{
			name:         "pass status",
			status:       "pass",
			expectedIcon: "✓",
		},
		{
			name:         "failed status",
			status:       "failed",
			expectedIcon: "✗",
		},
		{
			name:         "fail status",
			status:       "fail",
			expectedIcon: "✗",
		},
		{
			name:         "skipped status",
			status:       "skipped",
			expectedIcon: "⃠",
		},
		{
			name:         "skip status",
			status:       "skip",
			expectedIcon: "⃠",
		},
		{
			name:         "unknown status",
			status:       "unknown",
			expectedIcon: "?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			icon := icons.GetTestStatusIcon(tt.status)

			if icon != tt.expectedIcon {
				t.Errorf("Expected icon '%s' for status '%s', got '%s'", tt.expectedIcon, tt.status, icon)
			}
		})
	}
}

func TestReferenceIcons_GetArrowIcon(t *testing.T) {
	detector := &MockTerminalCapabilityDetector{
		supportsUnicode: true,
	}

	icons := NewReferenceIcons(detector)

	tests := []struct {
		name         string
		direction    string
		expectedIcon string
	}{
		{
			name:         "right arrow",
			direction:    "right",
			expectedIcon: "→",
		},
		{
			name:         "down right arrow",
			direction:    "down_right",
			expectedIcon: "↳",
		},
		{
			name:         "unknown direction",
			direction:    "unknown",
			expectedIcon: "→", // Default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			icon := icons.GetArrowIcon(tt.direction)

			if icon != tt.expectedIcon {
				t.Errorf("Expected icon '%s' for direction '%s', got '%s'", tt.expectedIcon, tt.direction, icon)
			}
		})
	}
}

func TestReferenceIcons_GetCodeContextIcon(t *testing.T) {
	detector := &MockTerminalCapabilityDetector{
		supportsUnicode: true,
	}

	icons := NewReferenceIcons(detector)

	tests := []struct {
		name         string
		context      string
		expectedIcon string
	}{
		{
			name:         "caret context",
			context:      "caret",
			expectedIcon: "^",
		},
		{
			name:         "pointer context",
			context:      "pointer",
			expectedIcon: "^",
		},
		{
			name:         "pipe context",
			context:      "pipe",
			expectedIcon: "|",
		},
		{
			name:         "separator context",
			context:      "separator",
			expectedIcon: "|",
		},
		{
			name:         "unknown context",
			context:      "unknown",
			expectedIcon: "^", // Default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			icon := icons.GetCodeContextIcon(tt.context)

			if icon != tt.expectedIcon {
				t.Errorf("Expected icon '%s' for context '%s', got '%s'", tt.expectedIcon, tt.context, icon)
			}
		})
	}
}

func TestReferenceIcons_GetTimerIcon(t *testing.T) {
	detector := &MockTerminalCapabilityDetector{
		supportsUnicode: true,
	}

	icons := NewReferenceIcons(detector)
	icon := icons.GetTimerIcon()

	if icon != "⏱️" {
		t.Errorf("Expected timer icon '⏱️', got '%s'", icon)
	}
}

func TestReferenceIcons_GetFallback(t *testing.T) {
	detector := &MockTerminalCapabilityDetector{
		supportsUnicode: true,
	}

	icons := NewReferenceIcons(detector)

	tests := []struct {
		name         string
		iconName     string
		expectedIcon string
	}{
		{
			name:         "test passed fallback",
			iconName:     "test_passed",
			expectedIcon: "[P]",
		},
		{
			name:         "test failed fallback",
			iconName:     "test_failed",
			expectedIcon: "[F]",
		},
		{
			name:         "nonexistent fallback",
			iconName:     "nonexistent",
			expectedIcon: "?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			icon := icons.GetFallback(tt.iconName)

			if icon != tt.expectedIcon {
				t.Errorf("Expected fallback icon '%s', got '%s'", tt.expectedIcon, icon)
			}
		})
	}
}

func TestReferenceIcons_ValidateUnicodeSupport(t *testing.T) {
	tests := []struct {
		name            string
		supportsUnicode bool
		expected        bool
	}{
		{
			name:            "with Unicode support",
			supportsUnicode: true,
			expected:        true,
		},
		{
			name:            "without Unicode support",
			supportsUnicode: false,
			expected:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := &MockTerminalCapabilityDetector{
				supportsUnicode: tt.supportsUnicode,
			}

			icons := NewReferenceIcons(detector)
			result := icons.ValidateUnicodeSupport()

			if result != tt.expected {
				t.Errorf("Expected Unicode validation %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestReferenceIcons_SupportsUnicode(t *testing.T) {
	tests := []struct {
		name            string
		supportsUnicode bool
	}{
		{
			name:            "with Unicode support",
			supportsUnicode: true,
		},
		{
			name:            "without Unicode support",
			supportsUnicode: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := &MockTerminalCapabilityDetector{
				supportsUnicode: tt.supportsUnicode,
			}

			icons := NewReferenceIcons(detector)

			if icons.SupportsUnicode() != tt.supportsUnicode {
				t.Errorf("Expected SupportsUnicode() to return %v", tt.supportsUnicode)
			}
		})
	}
}

func TestReferenceIcons_GetName(t *testing.T) {
	detector := &MockTerminalCapabilityDetector{}
	icons := NewReferenceIcons(detector)

	if icons.GetName() != "reference" {
		t.Errorf("Expected name 'reference', got '%s'", icons.GetName())
	}
}

func TestReferenceIcons_GetDescription(t *testing.T) {
	detector := &MockTerminalCapabilityDetector{}
	icons := NewReferenceIcons(detector)

	description := icons.GetDescription()
	if description == "" {
		t.Error("Expected non-empty description")
	}

	expectedKeywords := []string{"Reference", "Unicode", "Vitest"}
	for _, keyword := range expectedKeywords {
		if !strings.Contains(description, keyword) {
			t.Errorf("Expected description to contain '%s'", keyword)
		}
	}
}

func TestReferenceIcons_ListAvailableIcons(t *testing.T) {
	detector := &MockTerminalCapabilityDetector{}
	icons := NewReferenceIcons(detector)

	availableIcons := icons.ListAvailableIcons()

	expectedIcons := []string{
		"test_passed",
		"test_failed",
		"test_skipped",
		"arrow_right",
		"arrow_down_right",
		"caret",
		"pipe",
		"timer",
		"info",
		"warning",
	}

	if len(availableIcons) != len(expectedIcons) {
		t.Errorf("Expected %d icons, got %d", len(expectedIcons), len(availableIcons))
	}

	for _, expectedIcon := range expectedIcons {
		if !contains(availableIcons, expectedIcon) {
			t.Errorf("Expected icon '%s' to be available", expectedIcon)
		}
	}
}

func TestReferenceUnicodeIcons(t *testing.T) {
	// Test the predefined Unicode icons
	tests := []struct {
		name string
		icon UnicodeIcon
	}{
		{"ReferenceTestPassed", ReferenceTestPassed},
		{"ReferenceTestFailed", ReferenceTestFailed},
		{"ReferenceTestSkipped", ReferenceTestSkipped},
		{"ReferenceArrowRight", ReferenceArrowRight},
		{"ReferenceArrowDownRight", ReferenceArrowDownRight},
		{"ReferenceCaret", ReferenceCaret},
		{"ReferencePipe", ReferencePipe},
		{"ReferenceTimer", ReferenceTimer},
		{"ReferenceInfo", ReferenceInfo},
		{"ReferenceWarning", ReferenceWarning},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify Unicode character is set
			if tt.icon.Unicode == "" {
				t.Error("Unicode character should not be empty")
			}

			// Verify CodePoint is set
			if tt.icon.CodePoint == "" {
				t.Error("CodePoint should not be empty")
			}

			// Verify CodePoint format
			if len(tt.icon.CodePoint) < 6 || tt.icon.CodePoint[:2] != "U+" {
				t.Errorf("Expected CodePoint to start with 'U+', got '%s'", tt.icon.CodePoint)
			}

			// Verify Name is set
			if tt.icon.Name == "" {
				t.Error("Name should not be empty")
			}

			// Verify Description is set
			if tt.icon.Description == "" {
				t.Error("Description should not be empty")
			}

			// Verify Width is valid
			if tt.icon.Width < 1 || tt.icon.Width > 2 {
				t.Errorf("Width should be 1 or 2, got %d", tt.icon.Width)
			}
		})
	}
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
