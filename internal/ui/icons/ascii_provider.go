// Package icons provides ASCII icon provider implementation
package icons

import "fmt"

// ASCIIProvider implements IconProvider for ASCII characters
type ASCIIProvider struct {
	iconSet *IconSet
}

// NewASCIIProvider creates a new ASCII icon provider with fallback characters
func NewASCIIProvider() *ASCIIProvider {
	iconSet := &IconSet{
		Name:              "ascii",
		Description:       "ASCII icon set with fallback characters for limited terminals",
		RequiresUnicode:   false,
		RequiresEmoji:     false,
		RequiresNerdFonts: false,
		Icons: map[string]string{
			// Test status icons - ASCII fallbacks from guidelines
			"test_passed":  "[P]", // Pass
			"test_failed":  "[F]", // Fail
			"test_skipped": "[S]", // Skip

			// Navigation icons
			"arrow_right":      "->",   // Simple arrow
			"arrow_down_right": "\\->", // Down-right arrow

			// Code context icons
			"caret": "^", // Caret (same in ASCII)
			"pipe":  "|", // Pipe (same in ASCII)

			// Timer and utility icons
			"timer":   "T",   // Timer
			"info":    "[i]", // Info
			"warning": "[!]", // Warning

			// Legacy support
			"success": "[P]", // Same as test_passed
			"error":   "[F]", // Same as test_failed
		},
		Fallbacks: map[string]string{
			// ASCII provider uses same icons as fallbacks
			"test_passed":      "[P]",
			"test_failed":      "[F]",
			"test_skipped":     "[S]",
			"arrow_right":      "->",
			"arrow_down_right": "\\->",
			"caret":            "^",
			"pipe":             "|",
			"timer":            "T",
			"info":             "[i]",
			"warning":          "[!]",
			"success":          "[P]",
			"error":            "[F]",
		},
		Metadata: map[string]interface{}{
			"version":    "1.0",
			"source":     "vitest-visual-guidelines-ascii",
			"ascii_only": true,
		},
	}

	return &ASCIIProvider{
		iconSet: iconSet,
	}
}

// GetIcon retrieves an icon by name
func (a *ASCIIProvider) GetIcon(name string) (string, bool) {
	icon, exists := a.iconSet.Icons[name]
	return icon, exists
}

// SetIcon sets an icon for a given name
func (a *ASCIIProvider) SetIcon(name string, icon string) {
	if a.iconSet.Icons == nil {
		a.iconSet.Icons = make(map[string]string)
	}
	a.iconSet.Icons[name] = icon
}

// GetIconSet returns the current icon set
func (a *ASCIIProvider) GetIconSet() *IconSet {
	return a.iconSet
}

// SetIconSet changes the active icon set
func (a *ASCIIProvider) SetIconSet(iconSet *IconSet) error {
	if iconSet == nil {
		return fmt.Errorf("icon set cannot be nil")
	}
	a.iconSet = iconSet
	return nil
}

// SupportsUnicode returns false for ASCII provider
func (a *ASCIIProvider) SupportsUnicode() bool {
	return false
}

// GetFallback returns a fallback icon for unsupported characters
func (a *ASCIIProvider) GetFallback(name string) string {
	if fallback, exists := a.iconSet.Fallbacks[name]; exists {
		return fallback
	}
	return "?"
}
