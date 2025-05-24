// Package colors provides icon formatting capabilities
package colors

import (
	"fmt"

	"github.com/newbpydev/go-sentinel/pkg/models"
)

// Vitest-inspired icons
const (
	// Unicode symbols
	IconCheckMark = "✓"
	IconCross     = "✗"
	IconSkipped   = "⃠"
	IconRunning   = "⟳"

	// ASCII fallbacks
	IconCheckMarkASCII = "√"
	IconCrossASCII     = "×"
	IconSkippedASCII   = "○"
	IconRunningASCII   = "~"
)

// IconProviderInterface defines the interface for icon providers
type IconProviderInterface interface {
	// Basic test status icons
	CheckMark() string
	Cross() string
	Skipped() string
	Running() string

	// Generic icon retrieval
	GetIcon(iconType string) string

	// Check if Unicode is supported
	SupportsUnicode() bool
}

// IconProvider provides icons for test status and implements IconProviderInterface
type IconProvider struct {
	unicodeSupport bool
}

// NewIconProvider creates a new IconProvider
func NewIconProvider(unicodeSupport bool) *IconProvider {
	return &IconProvider{
		unicodeSupport: unicodeSupport,
	}
}

// NewAutoIconProvider creates an IconProvider with automatic Unicode detection
func NewAutoIconProvider() *IconProvider {
	detector := NewTerminalDetector()
	return &IconProvider{
		unicodeSupport: detector.SupportsUnicode(),
	}
}

// SupportsUnicode returns whether Unicode icons are supported
func (i *IconProvider) SupportsUnicode() bool {
	return i.unicodeSupport
}

// CheckMark returns the checkmark icon
func (i *IconProvider) CheckMark() string {
	if i.unicodeSupport {
		return IconCheckMark
	}
	return IconCheckMarkASCII
}

// Cross returns the cross icon
func (i *IconProvider) Cross() string {
	if i.unicodeSupport {
		return IconCross
	}
	return IconCrossASCII
}

// Skipped returns the skipped icon
func (i *IconProvider) Skipped() string {
	if i.unicodeSupport {
		return IconSkipped
	}
	return IconSkippedASCII
}

// Running returns the running icon
func (i *IconProvider) Running() string {
	if i.unicodeSupport {
		return IconRunning
	}
	return IconRunningASCII
}

// GetIcon returns an icon for various UI elements
func (i *IconProvider) GetIcon(iconType string) string {
	if !i.unicodeSupport {
		return i.getAsciiIcon(iconType)
	}
	return i.getUnicodeIcon(iconType)
}

func (i *IconProvider) getUnicodeIcon(iconType string) string {
	switch iconType {
	// Basic test status icons
	case "pass":
		return "✓"
	case "fail":
		return "✗"
	case "skip":
		return "⊝"
	case "running":
		return "⟳"

	// File change icons
	case "watch":
		return "👀"
	case "test":
		return "🧪"
	case "code":
		return "📝"
	case "config":
		return "⚙️"
	case "dependency":
		return "📦"
	case "file":
		return "📄"

	// Change type icons
	case "new":
		return "✨"
	case "change":
		return "🔄"
	case "unchanged":
		return "➖"

	// UI icons
	case "package":
		return "📁"
	case "summary":
		return "📊"
	case "info":
		return "ℹ️"
	case "unknown":
		return "❓"

	default:
		return "•"
	}
}

func (i *IconProvider) getAsciiIcon(iconType string) string {
	switch iconType {
	// Basic test status icons
	case "pass":
		return "[PASS]"
	case "fail":
		return "[FAIL]"
	case "skip":
		return "[SKIP]"
	case "running":
		return "[RUN ]"

	// File change icons
	case "watch":
		return "[WATCH]"
	case "test":
		return "[TEST]"
	case "code":
		return "[CODE]"
	case "config":
		return "[CONF]"
	case "dependency":
		return "[DEP ]"
	case "file":
		return "[FILE]"

	// Change type icons
	case "new":
		return "[NEW ]"
	case "change":
		return "[CHG ]"
	case "unchanged":
		return "[----]"

	// UI icons
	case "package":
		return "[PKG ]"
	case "summary":
		return "[SUM ]"
	case "info":
		return "[INFO]"
	case "unknown":
		return "[??? ]"

	default:
		return "[ • ]"
	}
}

// FormatTestStatus formats a test status with appropriate coloring and icon
func FormatTestStatus(status models.TestStatus, formatter FormatterInterface, icons IconProviderInterface) string {
	switch status {
	case models.StatusPassed:
		return formatter.Green(icons.CheckMark())
	case models.StatusFailed:
		return formatter.Red(icons.Cross())
	case models.StatusSkipped:
		return formatter.Yellow(icons.Skipped())
	case models.StatusRunning:
		return formatter.Blue(icons.Running())
	default:
		return fmt.Sprintf("[%s]", status)
	}
}

// Ensure IconProvider implements IconProviderInterface
var _ IconProviderInterface = (*IconProvider)(nil)
