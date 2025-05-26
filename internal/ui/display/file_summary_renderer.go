// Package display provides file summary rendering for test file results
package display

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/newbpydev/go-sentinel/internal/ui/colors"
	"github.com/newbpydev/go-sentinel/internal/ui/icons"
	"github.com/newbpydev/go-sentinel/pkg/models"
)

// FileSummaryRenderer handles rendering file summary lines
type FileSummaryRenderer struct {
	formatter       *colors.ColorFormatter
	icons           icons.IconProvider
	spacingManager  *SpacingManager
	timingFormatter *TimingFormatter
	config          *Config

	// Display options
	showMemoryUsage bool
	showTiming      bool
	maxPathLength   int
	indentLevel     int
}

// FileSummaryRenderOptions configures file summary rendering
type FileSummaryRenderOptions struct {
	ShowMemoryUsage bool
	ShowTiming      bool
	MaxPathLength   int
	IndentLevel     int
}

// NewFileSummaryRenderer creates a new file summary renderer
func NewFileSummaryRenderer(config *Config, options *FileSummaryRenderOptions) *FileSummaryRenderer {
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
		BaseIndent:    0, // No base indent for file summaries
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
		options = &FileSummaryRenderOptions{
			ShowMemoryUsage: true,
			ShowTiming:      true,
			MaxPathLength:   60,
			IndentLevel:     0,
		}
	} else {
		// Apply defaults to zero-value fields
		if options.MaxPathLength == 0 {
			options.MaxPathLength = 60
		}
	}

	return &FileSummaryRenderer{
		formatter:       formatter,
		icons:           iconProvider,
		spacingManager:  spacingManager,
		timingFormatter: timingFormatter,
		config:          config,
		showMemoryUsage: options.ShowMemoryUsage,
		showTiming:      options.ShowTiming,
		maxPathLength:   options.MaxPathLength,
		indentLevel:     options.IndentLevel,
	}
}

// RenderFileSummary renders a file summary line
// Format: "filename (X tests[ | Y failed]) Zms 0 MB heap used"
func (r *FileSummaryRenderer) RenderFileSummary(suite *models.TestSuite) string {
	if suite == nil {
		return ""
	}

	var result strings.Builder

	// Get indentation for this file summary
	indent := r.spacingManager.GetIndentString(r.indentLevel * 2)
	result.WriteString(indent)

	// Render filename (formatted and truncated if needed)
	filename := r.formatFilename(suite.FilePath)
	result.WriteString(filename)

	// Render test counts in parentheses
	testCounts := r.formatTestCounts(suite)
	result.WriteString(" ")
	result.WriteString(testCounts)

	// Render timing if enabled
	if r.showTiming {
		timing := r.timingFormatter.FormatDuration(suite.Duration)
		result.WriteString(" ")
		result.WriteString(timing)
	}

	// Render memory usage if enabled
	if r.showMemoryUsage {
		memoryUsage := r.formatMemoryUsage(suite.MemoryUsage)
		result.WriteString(" ")
		result.WriteString(memoryUsage)
	}

	return result.String()
}

// RenderFileSummaries renders multiple file summary lines
func (r *FileSummaryRenderer) RenderFileSummaries(suites []*models.TestSuite) string {
	if len(suites) == 0 {
		return ""
	}

	var result strings.Builder

	for i, suite := range suites {
		if suite == nil {
			continue
		}

		result.WriteString(r.RenderFileSummary(suite))

		// Add newline between summaries (except for the last one)
		if i < len(suites)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

// formatFilename formats a filename according to display requirements
func (r *FileSummaryRenderer) formatFilename(filename string) string {
	if filename == "" {
		return "(unknown file)"
	}

	// Extract just the filename from the path
	baseName := filepath.Base(filename)

	// Truncate if too long
	if len(baseName) > r.maxPathLength {
		if r.maxPathLength <= 3 {
			return strings.Repeat(".", r.maxPathLength)
		}
		return baseName[:r.maxPathLength-3] + "..."
	}

	return baseName
}

// formatTestCounts formats test counts in the required format
// Format: "(X tests[ | Y failed])"
func (r *FileSummaryRenderer) formatTestCounts(suite *models.TestSuite) string {
	if suite == nil {
		return "(0 tests)"
	}

	totalTests := suite.PassedCount + suite.FailedCount + suite.SkippedCount

	var result strings.Builder
	result.WriteString("(")

	// Always show total test count
	if totalTests == 1 {
		result.WriteString("1 test")
	} else {
		result.WriteString(fmt.Sprintf("%d tests", totalTests))
	}

	// Add failed count if there are failures
	if suite.FailedCount > 0 {
		result.WriteString(" | ")
		if suite.FailedCount == 1 {
			result.WriteString(r.formatter.Red("1 failed"))
		} else {
			result.WriteString(r.formatter.Red(fmt.Sprintf("%d failed", suite.FailedCount)))
		}
	}

	result.WriteString(")")
	return result.String()
}

// formatMemoryUsage formats memory usage in MB
func (r *FileSummaryRenderer) formatMemoryUsage(memoryBytes uint64) string {
	if memoryBytes == 0 {
		return "0 MB heap used"
	}

	// Convert bytes to MB
	memoryMB := float64(memoryBytes) / (1024 * 1024)

	if memoryMB < 0.1 {
		return "0 MB heap used"
	} else if memoryMB < 1.0 {
		return fmt.Sprintf("%.1f MB heap used", memoryMB)
	} else {
		return fmt.Sprintf("%.0f MB heap used", memoryMB)
	}
}

// SetIndentLevel sets the base indentation level
func (r *FileSummaryRenderer) SetIndentLevel(level int) {
	r.indentLevel = level
}

// SetShowTiming enables/disables timing display
func (r *FileSummaryRenderer) SetShowTiming(show bool) {
	r.showTiming = show
}

// SetShowMemoryUsage enables/disables memory usage display
func (r *FileSummaryRenderer) SetShowMemoryUsage(show bool) {
	r.showMemoryUsage = show
}

// SetMaxPathLength sets the maximum path length before truncation
func (r *FileSummaryRenderer) SetMaxPathLength(length int) {
	r.maxPathLength = length
}

// GetIndentLevel returns the current indentation level
func (r *FileSummaryRenderer) GetIndentLevel() int {
	return r.indentLevel
}

// IsShowTiming returns whether timing display is enabled
func (r *FileSummaryRenderer) IsShowTiming() bool {
	return r.showTiming
}

// IsShowMemoryUsage returns whether memory usage display is enabled
func (r *FileSummaryRenderer) IsShowMemoryUsage() bool {
	return r.showMemoryUsage
}

// GetMaxPathLength returns the maximum path length
func (r *FileSummaryRenderer) GetMaxPathLength() int {
	return r.maxPathLength
}
