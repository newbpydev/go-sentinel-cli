// Package icons provides icon providers and visual element interfaces
package icons

// IconProvider manages icons and visual elements for different terminal capabilities
type IconProvider interface {
	// GetIcon retrieves an icon by name
	GetIcon(name string) (string, bool)

	// SetIcon sets an icon for a given name
	SetIcon(name string, icon string)

	// GetIconSet returns the current icon set
	GetIconSet() *IconSet

	// SetIconSet changes the active icon set
	SetIconSet(iconSet *IconSet) error

	// SupportsUnicode returns whether Unicode icons are supported
	SupportsUnicode() bool

	// GetFallback returns a fallback icon for unsupported characters
	GetFallback(name string) string
}

// IconSetManager manages different icon sets and their capabilities
type IconSetManager interface {
	// LoadIconSet loads an icon set by name
	LoadIconSet(name string) (*IconSet, error)

	// ListIconSets returns available icon set names
	ListIconSets() []string

	// GetDefaultIconSet returns the default icon set
	GetDefaultIconSet() *IconSet

	// CreateCustomIconSet creates a new custom icon set
	CreateCustomIconSet(name string, definition *IconSetDefinition) (*IconSet, error)

	// DetectBestIconSet automatically detects the best icon set for the terminal
	DetectBestIconSet() (*IconSet, error)
}

// TerminalCapabilityDetector detects terminal capabilities for icon rendering
type TerminalCapabilityDetector interface {
	// SupportsUnicode detects if the terminal supports Unicode
	SupportsUnicode() bool

	// SupportsEmoji detects if the terminal supports emoji
	SupportsEmoji() bool

	// SupportsNerdFonts detects if Nerd Fonts are available
	SupportsNerdFonts() bool

	// GetCharacterWidth returns the width of a character in the terminal
	GetCharacterWidth(char string) int

	// GetCapabilities returns detailed terminal capabilities
	GetCapabilities() *TerminalCapabilities
}

// SpinnerProvider provides animated spinner characters
type SpinnerProvider interface {
	// GetSpinner returns a spinner by name
	GetSpinner(name string) (*Spinner, bool)

	// ListSpinners returns available spinner names
	ListSpinners() []string

	// GetDefaultSpinner returns the default spinner
	GetDefaultSpinner() *Spinner

	// CreateCustomSpinner creates a custom spinner
	CreateCustomSpinner(name string, frames []string, interval int) *Spinner
}

// IconSet represents a complete set of icons for terminal display
type IconSet struct {
	// Name is the icon set name
	Name string

	// Description is the icon set description
	Description string

	// RequiresUnicode indicates if Unicode support is required
	RequiresUnicode bool

	// RequiresEmoji indicates if emoji support is required
	RequiresEmoji bool

	// RequiresNerdFonts indicates if Nerd Fonts are required
	RequiresNerdFonts bool

	// Icons contains the icon mappings
	Icons map[string]string

	// Fallbacks contains fallback mappings for unsupported icons
	Fallbacks map[string]string

	// Metadata contains additional icon set metadata
	Metadata map[string]interface{}
}

// IconSetDefinition represents the definition for creating an icon set
type IconSetDefinition struct {
	// Name is the icon set name
	Name string

	// Description is the icon set description
	Description string

	// BaseIconSet is the base icon set to inherit from
	BaseIconSet string

	// IconOverrides contains icon overrides
	IconOverrides map[string]string

	// FallbackOverrides contains fallback overrides
	FallbackOverrides map[string]string

	// Requirements specifies the requirements for this icon set
	Requirements *IconRequirements
}

// IconRequirements specifies the requirements for an icon set
type IconRequirements struct {
	// Unicode indicates if Unicode support is required
	Unicode bool

	// Emoji indicates if emoji support is required
	Emoji bool

	// NerdFonts indicates if Nerd Fonts are required
	NerdFonts bool

	// MinTerminalWidth is the minimum required terminal width
	MinTerminalWidth int
}

// TerminalCapabilities represents terminal capabilities for icon rendering
type TerminalCapabilities struct {
	// Unicode indicates Unicode support
	Unicode bool

	// Emoji indicates emoji support
	Emoji bool

	// NerdFonts indicates Nerd Fonts support
	NerdFonts bool

	// TerminalWidth is the current terminal width
	TerminalWidth int

	// TerminalHeight is the current terminal height
	TerminalHeight int

	// FontName is the detected font name
	FontName string
}

// Spinner represents an animated spinner
type Spinner struct {
	// Name is the spinner name
	Name string

	// Frames contains the animation frames
	Frames []string

	// Interval is the animation interval in milliseconds
	Interval int

	// RequiresUnicode indicates if Unicode support is required
	RequiresUnicode bool
}

// Icon name constants for common test states and UI elements
const (
	// Test status icons
	IconTestPassed  = "test_passed"
	IconTestFailed  = "test_failed"
	IconTestSkipped = "test_skipped"
	IconTestRunning = "test_running"

	// General status icons
	IconSuccess = "success"
	IconError   = "error"
	IconWarning = "warning"
	IconInfo    = "info"

	// Progress and activity icons
	IconSpinner  = "spinner"
	IconProgress = "progress"
	IconLoading  = "loading"

	// File and package icons
	IconPackage = "package"
	IconFile    = "file"
	IconFolder  = "folder"

	// Coverage icons
	IconCoverage     = "coverage"
	IconCoverageHigh = "coverage_high"
	IconCoverageLow  = "coverage_low"

	// Watch mode icons
	IconWatch      = "watch"
	IconWatchStart = "watch_start"
	IconWatchStop  = "watch_stop"

	// UI elements
	IconArrowRight = "arrow_right"
	IconArrowLeft  = "arrow_left"
	IconArrowUp    = "arrow_up"
	IconArrowDown  = "arrow_down"
	IconBullet     = "bullet"
	IconDash       = "dash"
	IconPlus       = "plus"
	IconMinus      = "minus"
	IconCross      = "cross"
	IconCheck      = "check"
)

// Predefined icon sets
var (
	// NoneIconSet provides text-only icons
	NoneIconSet = &IconSet{
		Name:              "none",
		Description:       "Text-only icons for basic terminals",
		RequiresUnicode:   false,
		RequiresEmoji:     false,
		RequiresNerdFonts: false,
		Icons: map[string]string{
			IconTestPassed:  "PASS",
			IconTestFailed:  "FAIL",
			IconTestSkipped: "SKIP",
			IconTestRunning: "RUN",
			IconSuccess:     "OK",
			IconError:       "ERR",
			IconWarning:     "WARN",
			IconInfo:        "INFO",
			IconSpinner:     "|",
			IconProgress:    "=",
			IconPackage:     "PKG",
			IconFile:        "FILE",
			IconFolder:      "DIR",
			IconCoverage:    "COV",
			IconWatch:       "WATCH",
			IconArrowRight:  ">",
			IconArrowLeft:   "<",
			IconArrowUp:     "^",
			IconArrowDown:   "v",
			IconBullet:      "*",
			IconDash:        "-",
			IconPlus:        "+",
			IconMinus:       "-",
			IconCross:       "x",
			IconCheck:       "✓",
		},
		Fallbacks: map[string]string{},
	}

	// SimpleIconSet provides basic Unicode icons
	SimpleIconSet = &IconSet{
		Name:              "simple",
		Description:       "Simple Unicode icons for most terminals",
		RequiresUnicode:   true,
		RequiresEmoji:     false,
		RequiresNerdFonts: false,
		Icons: map[string]string{
			IconTestPassed:  "✓",
			IconTestFailed:  "✗",
			IconTestSkipped: "⊝",
			IconTestRunning: "●",
			IconSuccess:     "✓",
			IconError:       "✗",
			IconWarning:     "⚠",
			IconInfo:        "ℹ",
			IconSpinner:     "◐",
			IconProgress:    "█",
			IconPackage:     "📦",
			IconFile:        "📄",
			IconFolder:      "📁",
			IconCoverage:    "📊",
			IconWatch:       "👁",
			IconArrowRight:  "→",
			IconArrowLeft:   "←",
			IconArrowUp:     "↑",
			IconArrowDown:   "↓",
			IconBullet:      "•",
			IconDash:        "─",
			IconPlus:        "➕",
			IconMinus:       "➖",
			IconCross:       "❌",
			IconCheck:       "✅",
		},
		Fallbacks: map[string]string{
			IconTestPassed:  "PASS",
			IconTestFailed:  "FAIL",
			IconTestSkipped: "SKIP",
		},
	}

	// RichIconSet provides rich icons with emoji and special characters
	RichIconSet = &IconSet{
		Name:              "rich",
		Description:       "Rich icons with emoji and Unicode symbols",
		RequiresUnicode:   true,
		RequiresEmoji:     true,
		RequiresNerdFonts: false,
		Icons: map[string]string{
			IconTestPassed:  "✅",
			IconTestFailed:  "❌",
			IconTestSkipped: "⏭️",
			IconTestRunning: "🏃",
			IconSuccess:     "🎉",
			IconError:       "💥",
			IconWarning:     "⚠️",
			IconInfo:        "ℹ️",
			IconSpinner:     "🔄",
			IconProgress:    "▓",
			IconPackage:     "📦",
			IconFile:        "📄",
			IconFolder:      "📁",
			IconCoverage:    "📊",
			IconWatch:       "👀",
			IconArrowRight:  "▶️",
			IconArrowLeft:   "◀️",
			IconArrowUp:     "🔼",
			IconArrowDown:   "🔽",
			IconBullet:      "🔸",
			IconDash:        "➖",
			IconPlus:        "➕",
			IconMinus:       "➖",
			IconCross:       "❌",
			IconCheck:       "✅",
		},
		Fallbacks: map[string]string{
			IconTestPassed:  "✓",
			IconTestFailed:  "✗",
			IconTestSkipped: "⊝",
		},
	}
)

// Predefined spinners
var (
	// DotsSpinner is a simple dots spinner
	DotsSpinner = &Spinner{
		Name:            "dots",
		Frames:          []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		Interval:        80,
		RequiresUnicode: true,
	}

	// LineSpinner is a simple line spinner
	LineSpinner = &Spinner{
		Name:            "line",
		Frames:          []string{"|", "/", "-", "\\"},
		Interval:        100,
		RequiresUnicode: false,
	}

	// ArrowSpinner is an arrow-based spinner
	ArrowSpinner = &Spinner{
		Name:            "arrow",
		Frames:          []string{"←", "↖", "↑", "↗", "→", "↘", "↓", "↙"},
		Interval:        120,
		RequiresUnicode: true,
	}

	// EmojiSpinner is an emoji-based spinner
	EmojiSpinner = &Spinner{
		Name:            "emoji",
		Frames:          []string{"🕐", "🕑", "🕒", "🕓", "🕔", "🕕", "🕖", "🕗", "🕘", "🕙", "🕚", "🕛"},
		Interval:        100,
		RequiresUnicode: true,
	}
)
