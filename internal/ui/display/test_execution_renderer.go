// Package display provides test execution rendering for individual test results
package display

import (
	"strings"

	"github.com/newbpydev/go-sentinel/internal/ui/colors"
	"github.com/newbpydev/go-sentinel/internal/ui/icons"
	"github.com/newbpydev/go-sentinel/pkg/models"
)

// TestExecutionRenderer handles rendering individual test execution lines
type TestExecutionRenderer struct {
	formatter       *colors.ColorFormatter
	icons           icons.IconProvider
	spacingManager  *SpacingManager
	timingFormatter *TimingFormatter
	config          *Config

	// Display options
	showTiming    bool
	showSubtests  bool
	maxNameLength int
	indentLevel   int
}

// TestExecutionRenderOptions configures test execution rendering
type TestExecutionRenderOptions struct {
	ShowTiming    bool
	ShowSubtests  bool
	MaxNameLength int
	IndentLevel   int
}

// NewTestExecutionRenderer creates a new test execution renderer
func NewTestExecutionRenderer(config *Config, options *TestExecutionRenderOptions) *TestExecutionRenderer {
	formatter := colors.NewAutoColorFormatter()

	// Detect terminal capabilities for icon selection
	detector := colors.NewTerminalDetector()
	var iconProvider icons.IconProvider
	if detector.SupportsUnicode() {
		iconProvider = icons.NewUnicodeProvider()
	} else {
		iconProvider = icons.NewASCIIProvider()
	}

	spacingManager := NewSpacingManager(&SpacingConfig{
		BaseIndent:    2,
		TestIndent:    2,
		SubtestIndent: 4,
		ErrorIndent:   4,
	})

	timingFormatter := NewTimingFormatter(&TimingConfig{
		ShowMilliseconds: true,
		ShowMicroseconds: false,
		IntegerFormat:    true,
		MinWidth:         4,
	})

	// Set defaults if options not provided or apply defaults to unset fields
	if options == nil {
		options = &TestExecutionRenderOptions{
			ShowTiming:    true,
			ShowSubtests:  true,
			MaxNameLength: 80,
			IndentLevel:   0,
		}
	} else {
		// Apply defaults to zero-value fields
		if options.MaxNameLength == 0 {
			options.MaxNameLength = 80
		}
	}

	return &TestExecutionRenderer{
		formatter:       formatter,
		icons:           iconProvider,
		spacingManager:  spacingManager,
		timingFormatter: timingFormatter,
		config:          config,
		showTiming:      options.ShowTiming,
		showSubtests:    options.ShowSubtests,
		maxNameLength:   options.MaxNameLength,
		indentLevel:     options.IndentLevel,
	}
}

// RenderTestExecution renders an individual test execution line
func (r *TestExecutionRenderer) RenderTestExecution(test *models.TestResult) string {
	if test == nil {
		return ""
	}

	var result strings.Builder

	// Get base indentation for this test
	indent := r.spacingManager.GetTestIndent(r.indentLevel)
	result.WriteString(indent)

	// Render status icon
	icon := r.getStatusIcon(test.Status)
	coloredIcon := r.colorizeIcon(icon, test.Status)
	result.WriteString(coloredIcon)
	result.WriteString(" ")

	// Render test name (formatted and truncated if needed)
	testName := r.formatTestName(test.Name)
	coloredName := r.colorizeTestName(testName, test.Status)
	result.WriteString(coloredName)

	// Render timing if enabled
	if r.showTiming {
		timing := r.timingFormatter.FormatDuration(test.Duration)
		result.WriteString(" ")
		result.WriteString(r.formatter.Dim(timing))
	}

	// Render subtests if enabled and present
	if r.showSubtests && len(test.Subtests) > 0 {
		subtestResults := r.renderSubtests(test.Subtests)
		if subtestResults != "" {
			result.WriteString("\n")
			result.WriteString(subtestResults)
		}
	}

	return result.String()
}

// RenderTestExecutions renders multiple test execution lines
func (r *TestExecutionRenderer) RenderTestExecutions(tests []*models.TestResult) string {
	if len(tests) == 0 {
		return ""
	}

	var result strings.Builder

	for i, test := range tests {
		if test == nil {
			continue
		}

		result.WriteString(r.RenderTestExecution(test))

		// Add newline between tests (except for the last one)
		if i < len(tests)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

// getStatusIcon returns the appropriate icon for a test status
func (r *TestExecutionRenderer) getStatusIcon(status models.TestStatus) string {
	switch status {
	case models.TestStatusPassed:
		icon, _ := r.icons.GetIcon("test_passed")
		return icon
	case models.TestStatusFailed:
		icon, _ := r.icons.GetIcon("test_failed")
		return icon
	case models.TestStatusSkipped:
		icon, _ := r.icons.GetIcon("test_skipped")
		return icon
	case models.TestStatusRunning:
		// Use a neutral icon for running tests
		return "⋯"
	default:
		return "?"
	}
}

// colorizeIcon applies appropriate color to an icon based on test status
func (r *TestExecutionRenderer) colorizeIcon(icon string, status models.TestStatus) string {
	switch status {
	case models.TestStatusPassed:
		return r.formatter.Green(icon)
	case models.TestStatusFailed:
		return r.formatter.Red(icon)
	case models.TestStatusSkipped:
		return r.formatter.Yellow(icon)
	case models.TestStatusRunning:
		return r.formatter.Cyan(icon)
	default:
		return r.formatter.Dim(icon)
	}
}

// formatTestName formats a test name according to display requirements
func (r *TestExecutionRenderer) formatTestName(name string) string {
	if name == "" {
		return "(unnamed test)"
	}

	// Truncate if too long
	if len(name) > r.maxNameLength {
		if r.maxNameLength <= 3 {
			return strings.Repeat(".", r.maxNameLength)
		}
		return name[:r.maxNameLength-3] + "..."
	}

	return name
}

// colorizeTestName applies appropriate color to test name based on status
func (r *TestExecutionRenderer) colorizeTestName(name string, status models.TestStatus) string {
	switch status {
	case models.TestStatusPassed:
		return name // No coloring for passed test names
	case models.TestStatusFailed:
		return name // No coloring for failed test names
	case models.TestStatusSkipped:
		return r.formatter.Dim(name)
	case models.TestStatusRunning:
		return name
	default:
		return r.formatter.Dim(name)
	}
}

// renderSubtests renders subtests with increased indentation
func (r *TestExecutionRenderer) renderSubtests(subtests []*models.TestResult) string {
	if len(subtests) == 0 {
		return ""
	}

	var result strings.Builder

	// Create a new renderer with increased indent level for subtests
	subtestRenderer := *r
	subtestRenderer.indentLevel = r.indentLevel + 1

	for i, subtest := range subtests {
		if subtest == nil {
			continue
		}

		result.WriteString(subtestRenderer.RenderTestExecution(subtest))

		// Add newline between subtests (except for the last one)
		if i < len(subtests)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

// SetIndentLevel sets the base indentation level
func (r *TestExecutionRenderer) SetIndentLevel(level int) {
	r.indentLevel = level
}

// SetShowTiming enables/disables timing display
func (r *TestExecutionRenderer) SetShowTiming(show bool) {
	r.showTiming = show
}

// SetShowSubtests enables/disables subtest display
func (r *TestExecutionRenderer) SetShowSubtests(show bool) {
	r.showSubtests = show
}

// SetMaxNameLength sets the maximum test name length before truncation
func (r *TestExecutionRenderer) SetMaxNameLength(length int) {
	r.maxNameLength = length
}

// GetIndentLevel returns the current indentation level
func (r *TestExecutionRenderer) GetIndentLevel() int {
	return r.indentLevel
}

// IsShowTiming returns whether timing display is enabled
func (r *TestExecutionRenderer) IsShowTiming() bool {
	return r.showTiming
}

// IsShowSubtests returns whether subtest display is enabled
func (r *TestExecutionRenderer) IsShowSubtests() bool {
	return r.showSubtests
}

// GetMaxNameLength returns the maximum test name length
func (r *TestExecutionRenderer) GetMaxNameLength() int {
	return r.maxNameLength
}
