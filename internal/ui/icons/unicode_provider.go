// Package icons provides Unicode icon provider implementation
package icons

import "fmt"

// UnicodeProvider implements IconProvider for Unicode characters
type UnicodeProvider struct {
	iconSet *IconSet
}

// NewUnicodeProvider creates a new Unicode icon provider with default icons
func NewUnicodeProvider() *UnicodeProvider {
	iconSet := &IconSet{
		Name:              "unicode",
		Description:       "Unicode icon set with exact character mappings",
		RequiresUnicode:   true,
		RequiresEmoji:     false,
		RequiresNerdFonts: false,
		Icons: map[string]string{
			// Test status icons - exact Unicode from guidelines
			"test_passed":  "✓", // U+2713
			"test_failed":  "✗", // U+2717
			"test_skipped": "⃠", // U+20E0

			// Navigation icons
			"arrow_right":      "→", // U+2192
			"arrow_down_right": "↳", // U+21B3

			// Code context icons
			"caret": "^", // U+005E
			"pipe":  "|", // U+007C

			// Timer and utility icons
			"timer":   "⏱️", // U+23F1
			"info":    "ℹ",  // U+2139
			"warning": "⚠",  // U+26A0

			// Legacy support
			"success": "✓", // Same as test_passed
			"error":   "✗", // Same as test_failed
		},
		Fallbacks: map[string]string{
			"test_passed":      "[P]",
			"test_failed":      "[F]",
			"test_skipped":     "[S]",
			"arrow_right":      "->",
			"arrow_down_right": "\\->",
			"caret":            "^",
			"pipe":             "|",
			"timer":            "⏱",
			"info":             "[i]",
			"warning":          "[!]",
			"success":          "[P]",
			"error":            "[F]",
		},
		Metadata: map[string]interface{}{
			"version":         "1.0",
			"source":          "vitest-visual-guidelines",
			"unicode_version": "13.0",
		},
	}

	return &UnicodeProvider{
		iconSet: iconSet,
	}
}

// GetIcon retrieves an icon by name
func (u *UnicodeProvider) GetIcon(name string) (string, bool) {
	icon, exists := u.iconSet.Icons[name]
	return icon, exists
}

// SetIcon sets an icon for a given name
func (u *UnicodeProvider) SetIcon(name string, icon string) {
	if u.iconSet.Icons == nil {
		u.iconSet.Icons = make(map[string]string)
	}
	u.iconSet.Icons[name] = icon
}

// GetIconSet returns the current icon set
func (u *UnicodeProvider) GetIconSet() *IconSet {
	return u.iconSet
}

// SetIconSet changes the active icon set
func (u *UnicodeProvider) SetIconSet(iconSet *IconSet) error {
	if iconSet == nil {
		return fmt.Errorf("icon set cannot be nil")
	}
	u.iconSet = iconSet
	return nil
}

// SupportsUnicode returns true for Unicode provider
func (u *UnicodeProvider) SupportsUnicode() bool {
	return true
}

// GetFallback returns a fallback icon for unsupported characters
func (u *UnicodeProvider) GetFallback(name string) string {
	if fallback, exists := u.iconSet.Fallbacks[name]; exists {
		return fallback
	}
	return "?"
}
