// Package icons provides reference icon implementation with exact Unicode characters
package icons

// ReferenceIcons implements exact Unicode character mapping from visual guidelines
// Unicode points: ✓ U+2713, ✗ U+2717, ⃠ U+20E0, → U+2192, ↳ U+21B3, ^ U+005E, | U+007C, ⏱️ U+23F1
type ReferenceIcons struct {
	name            string
	supportsUnicode bool
	detector        TerminalCapabilityDetector
	unicodeProvider *UnicodeProvider
	asciiProvider   *ASCIIProvider
}

// NewReferenceIcons creates a new reference icons implementation with terminal detection
func NewReferenceIcons(detector TerminalCapabilityDetector) *ReferenceIcons {
	return &ReferenceIcons{
		name:            "reference",
		supportsUnicode: detector.SupportsUnicode(),
		detector:        detector,
		unicodeProvider: NewUnicodeProvider(),
		asciiProvider:   NewASCIIProvider(),
	}
}

// UnicodeIcon represents a Unicode icon with its exact character and metadata
type UnicodeIcon struct {
	Unicode     string // Exact Unicode character (e.g., "✓")
	CodePoint   string // Unicode code point (e.g., "U+2713")
	Name        string // Icon name for reference
	Description string // Description of usage
	Width       int    // Character width (1 for most, 2 for emoji)
}

// Reference Unicode icons from visual guidelines
var (
	// Test status icons - exact Unicode points from guidelines
	ReferenceTestPassed = UnicodeIcon{
		Unicode:     "✓",
		CodePoint:   "U+2713",
		Name:        "test_passed",
		Description: "Check mark for passed tests",
		Width:       1,
	}

	ReferenceTestFailed = UnicodeIcon{
		Unicode:     "✗",
		CodePoint:   "U+2717",
		Name:        "test_failed",
		Description: "X mark for failed tests",
		Width:       1,
	}

	ReferenceTestSkipped = UnicodeIcon{
		Unicode:     "⃠",
		CodePoint:   "U+20E0",
		Name:        "test_skipped",
		Description: "Prohibition sign for skipped tests",
		Width:       1,
	}

	// Navigation and pointer icons
	ReferenceArrowRight = UnicodeIcon{
		Unicode:     "→",
		CodePoint:   "U+2192",
		Name:        "arrow_right",
		Description: "Right arrow for error details",
		Width:       1,
	}

	ReferenceArrowDownRight = UnicodeIcon{
		Unicode:     "↳",
		CodePoint:   "U+21B3",
		Name:        "arrow_down_right",
		Description: "Down-right arrow for stack traces",
		Width:       1,
	}

	// Code context icons
	ReferenceCaret = UnicodeIcon{
		Unicode:     "^",
		CodePoint:   "U+005E",
		Name:        "caret",
		Description: "Caret for pointing to error location in code",
		Width:       1,
	}

	ReferencePipe = UnicodeIcon{
		Unicode:     "|",
		CodePoint:   "U+007C",
		Name:        "pipe",
		Description: "Pipe for separating test stats",
		Width:       1,
	}

	// Timer icon for execution time
	ReferenceTimer = UnicodeIcon{
		Unicode:     "⏱️",
		CodePoint:   "U+23F1",
		Name:        "timer",
		Description: "Timer for execution duration",
		Width:       2, // Emoji width
	}

	// Additional utility icons
	ReferenceInfo = UnicodeIcon{
		Unicode:     "ℹ",
		CodePoint:   "U+2139",
		Name:        "info",
		Description: "Information symbol",
		Width:       1,
	}

	ReferenceWarning = UnicodeIcon{
		Unicode:     "⚠",
		CodePoint:   "U+26A0",
		Name:        "warning",
		Description: "Warning symbol",
		Width:       1,
	}
)

// GetIcon retrieves an icon by name, returning Unicode or ASCII fallback
func (r *ReferenceIcons) GetIcon(name string) (string, bool) {
	// If Unicode is supported, use Unicode provider
	if r.supportsUnicode {
		return r.unicodeProvider.GetIcon(name)
	}

	// Fall back to ASCII provider
	return r.asciiProvider.GetIcon(name)
}

// SetIcon sets an icon for a given name (for custom icons)
func (r *ReferenceIcons) SetIcon(name string, icon string) {
	// Set in both providers to maintain consistency
	r.unicodeProvider.SetIcon(name, icon)
	r.asciiProvider.SetIcon(name, icon)
}

// GetIconSet returns the current icon set
func (r *ReferenceIcons) GetIconSet() *IconSet {
	if r.supportsUnicode {
		return r.unicodeProvider.GetIconSet()
	}
	return r.asciiProvider.GetIconSet()
}

// SetIconSet changes the active icon set
func (r *ReferenceIcons) SetIconSet(iconSet *IconSet) error {
	// Update both providers
	if err := r.unicodeProvider.SetIconSet(iconSet); err != nil {
		return err
	}
	return r.asciiProvider.SetIconSet(iconSet)
}

// SupportsUnicode returns whether Unicode icons are supported
func (r *ReferenceIcons) SupportsUnicode() bool {
	return r.supportsUnicode
}

// GetFallback returns a fallback icon for unsupported characters
func (r *ReferenceIcons) GetFallback(name string) string {
	// Always use ASCII provider for fallbacks
	icon, exists := r.asciiProvider.GetIcon(name)
	if !exists {
		return "?"
	}
	return icon
}

// GetTestStatusIcon returns the appropriate icon for a test status
func (r *ReferenceIcons) GetTestStatusIcon(status string) string {
	switch status {
	case "passed", "pass":
		icon, _ := r.GetIcon("test_passed")
		return icon
	case "failed", "fail":
		icon, _ := r.GetIcon("test_failed")
		return icon
	case "skipped", "skip":
		icon, _ := r.GetIcon("test_skipped")
		return icon
	default:
		return "?"
	}
}

// GetArrowIcon returns arrow icons for different contexts
func (r *ReferenceIcons) GetArrowIcon(direction string) string {
	switch direction {
	case "right":
		icon, _ := r.GetIcon("arrow_right")
		return icon
	case "down_right":
		icon, _ := r.GetIcon("arrow_down_right")
		return icon
	default:
		return "→"
	}
}

// GetCodeContextIcon returns icons for code context display
func (r *ReferenceIcons) GetCodeContextIcon(context string) string {
	switch context {
	case "caret", "pointer":
		icon, _ := r.GetIcon("caret")
		return icon
	case "pipe", "separator":
		icon, _ := r.GetIcon("pipe")
		return icon
	default:
		return "^"
	}
}

// GetTimerIcon returns the timer icon for execution time
func (r *ReferenceIcons) GetTimerIcon() string {
	icon, _ := r.GetIcon("timer")
	return icon
}

// ValidateUnicodeSupport tests if the terminal properly supports required Unicode characters
func (r *ReferenceIcons) ValidateUnicodeSupport() bool {
	// Test the essential characters from our icon set
	testIcons := []UnicodeIcon{
		ReferenceTestPassed,
		ReferenceTestFailed,
		ReferenceTestSkipped,
		ReferenceArrowRight,
		ReferenceArrowDownRight,
	}

	// In a real implementation, this would test character rendering
	// For now, we rely on the detector's Unicode support detection
	_ = testIcons // Mark as used for validation purposes
	return r.detector.SupportsUnicode()
}

// GetName returns the provider name
func (r *ReferenceIcons) GetName() string {
	return r.name
}

// GetDescription returns a description of this icon provider
func (r *ReferenceIcons) GetDescription() string {
	return "Reference icon provider with exact Unicode characters matching Vitest visual guidelines"
}

// ListAvailableIcons returns all available icon names
func (r *ReferenceIcons) ListAvailableIcons() []string {
	return []string{
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
}
