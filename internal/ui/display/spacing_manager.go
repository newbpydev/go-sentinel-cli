// Package display provides spacing management for precise indentation control
package display

import (
	"strings"
)

// SpacingManager handles precise spacing and indentation for test display
type SpacingManager struct {
	config *SpacingConfig
}

// SpacingConfig configures spacing behavior
type SpacingConfig struct {
	BaseIndent    int // Base indentation (2 spaces for "  ✓")
	TestIndent    int // Additional indent per test level
	SubtestIndent int // Additional indent for subtests
	ErrorIndent   int // Indentation for error details
}

// NewSpacingManager creates a new spacing manager with configuration
func NewSpacingManager(config *SpacingConfig) *SpacingManager {
	if config == nil {
		config = &SpacingConfig{
			BaseIndent:    2,
			TestIndent:    2,
			SubtestIndent: 4,
			ErrorIndent:   4,
		}
	}

	return &SpacingManager{
		config: config,
	}
}

// GetTestIndent returns the precise indentation for a test at the given level
func (s *SpacingManager) GetTestIndent(level int) string {
	totalIndent := s.config.BaseIndent + (level * s.config.TestIndent)
	return strings.Repeat(" ", totalIndent)
}

// GetSubtestIndent returns the indentation for subtests
func (s *SpacingManager) GetSubtestIndent(level int) string {
	totalIndent := s.config.BaseIndent + (level * s.config.TestIndent) + s.config.SubtestIndent
	return strings.Repeat(" ", totalIndent)
}

// GetErrorIndent returns the indentation for error details
func (s *SpacingManager) GetErrorIndent(level int) string {
	totalIndent := s.config.BaseIndent + (level * s.config.TestIndent) + s.config.ErrorIndent
	return strings.Repeat(" ", totalIndent)
}

// GetIndentString returns a string of spaces for the given count
func (s *SpacingManager) GetIndentString(spaces int) string {
	if spaces <= 0 {
		return ""
	}
	return strings.Repeat(" ", spaces)
}

// CalculateAlignment calculates alignment for columns (e.g., timing alignment)
func (s *SpacingManager) CalculateAlignment(text string, targetWidth int) string {
	textLen := len(text)
	if textLen >= targetWidth {
		return text
	}

	padding := targetWidth - textLen
	return text + strings.Repeat(" ", padding)
}

// AlignRight right-aligns text within the given width
func (s *SpacingManager) AlignRight(text string, width int) string {
	textLen := len(text)
	if textLen >= width {
		return text
	}

	padding := width - textLen
	return strings.Repeat(" ", padding) + text
}

// AlignCenter center-aligns text within the given width
func (s *SpacingManager) AlignCenter(text string, width int) string {
	textLen := len(text)
	if textLen >= width {
		return text
	}

	padding := width - textLen
	leftPadding := padding / 2
	rightPadding := padding - leftPadding

	return strings.Repeat(" ", leftPadding) + text + strings.Repeat(" ", rightPadding)
}

// TruncateWithEllipsis truncates text to fit width, adding ellipsis if needed
func (s *SpacingManager) TruncateWithEllipsis(text string, maxWidth int) string {
	if len(text) <= maxWidth {
		return text
	}

	if maxWidth <= 3 {
		return strings.Repeat(".", maxWidth)
	}

	return text[:maxWidth-3] + "..."
}

// FormatTwoColumn formats text into two columns with proper spacing
func (s *SpacingManager) FormatTwoColumn(left, right string, totalWidth int) string {
	leftLen := len(left)
	rightLen := len(right)

	// If combined length exceeds total width, truncate left column
	if leftLen+rightLen+1 > totalWidth {
		availableLeft := totalWidth - rightLen - 1
		if availableLeft > 0 {
			left = s.TruncateWithEllipsis(left, availableLeft)
		}
	}

	spacingNeeded := totalWidth - len(left) - len(right)
	if spacingNeeded < 1 {
		spacingNeeded = 1
	}

	return left + strings.Repeat(" ", spacingNeeded) + right
}

// GetBaseIndent returns the base indentation amount
func (s *SpacingManager) GetBaseIndent() int {
	return s.config.BaseIndent
}

// GetTestIndentAmount returns the test indentation amount
func (s *SpacingManager) GetTestIndentAmount() int {
	return s.config.TestIndent
}

// GetSubtestIndentAmount returns the subtest indentation amount
func (s *SpacingManager) GetSubtestIndentAmount() int {
	return s.config.SubtestIndent
}

// GetErrorIndentAmount returns the error indentation amount
func (s *SpacingManager) GetErrorIndentAmount() int {
	return s.config.ErrorIndent
}

// SetBaseIndent updates the base indentation amount
func (s *SpacingManager) SetBaseIndent(indent int) {
	s.config.BaseIndent = indent
}

// SetTestIndent updates the test indentation amount
func (s *SpacingManager) SetTestIndent(indent int) {
	s.config.TestIndent = indent
}

// SetSubtestIndent updates the subtest indentation amount
func (s *SpacingManager) SetSubtestIndent(indent int) {
	s.config.SubtestIndent = indent
}

// SetErrorIndent updates the error indentation amount
func (s *SpacingManager) SetErrorIndent(indent int) {
	s.config.ErrorIndent = indent
}
