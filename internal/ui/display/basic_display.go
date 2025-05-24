// Package display provides basic display formatting and rendering capabilities
package display

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/newbpydev/go-sentinel/internal/ui/colors"
	"github.com/newbpydev/go-sentinel/pkg/models"
)

// RendererInterface defines the interface for display renderers
type RendererInterface interface {
	// RenderSuiteHeader renders a test suite header
	RenderSuiteHeader(suite *models.TestSuite) error

	// GetWriter returns the output writer
	GetWriter() io.Writer

	// SetWidth sets the display width
	SetWidth(width int)
}

// HeaderRenderer renders test suite headers and implements RendererInterface
type HeaderRenderer struct {
	writer    io.Writer
	formatter colors.FormatterInterface
	icons     colors.IconProviderInterface
	width     int
}

// NewHeaderRenderer creates a new HeaderRenderer
func NewHeaderRenderer(writer io.Writer, formatter colors.FormatterInterface, icons colors.IconProviderInterface, width int) *HeaderRenderer {
	return &HeaderRenderer{
		writer:    writer,
		formatter: formatter,
		icons:     icons,
		width:     width,
	}
}

// RenderSuiteHeader renders a test suite header
func (r *HeaderRenderer) RenderSuiteHeader(suite *models.TestSuite) error {
	// Format file path
	filePath := FormatFilePath(r.formatter, suite.FilePath)

	// Format test counts
	counts := formatTestCounts(r.formatter, suite.TestCount, suite.PassedCount, suite.FailedCount, suite.SkippedCount)

	// Format duration
	duration := FormatDuration(r.formatter, suite.Duration)

	// Format memory usage
	memory := FormatMemoryUsage(r.formatter, suite.MemoryUsage)

	// Combine all parts
	header := fmt.Sprintf("%s %s %s %s\n",
		filePath,
		counts,
		duration,
		memory)

	// Write to output
	_, err := fmt.Fprint(r.writer, header)
	return err
}

// GetWriter returns the output writer
func (r *HeaderRenderer) GetWriter() io.Writer {
	return r.writer
}

// SetWidth sets the display width
func (r *HeaderRenderer) SetWidth(width int) {
	r.width = width
}

// PathFormatterInterface defines the interface for path formatting
type PathFormatterInterface interface {
	FormatFilePath(path string) string
}

// PathFormatter handles file path formatting
type PathFormatter struct {
	formatter colors.FormatterInterface
}

// NewPathFormatter creates a new PathFormatter
func NewPathFormatter(formatter colors.FormatterInterface) *PathFormatter {
	return &PathFormatter{
		formatter: formatter,
	}
}

// FormatFilePath formats a file path with colorized file name
func (p *PathFormatter) FormatFilePath(path string) string {
	return FormatFilePath(p.formatter, path)
}

// FormatFilePath formats a file path with colorized file name (standalone function)
func FormatFilePath(formatter colors.FormatterInterface, path string) string {
	// Get directory and file name
	dir, file := filepath.Split(path)

	// Normalize path separators to forward slashes for consistent display
	dir = strings.ReplaceAll(dir, "\\", "/")

	// Clean up directory
	dir = strings.TrimSuffix(dir, string(filepath.Separator))
	dir = strings.TrimSuffix(dir, "/")

	// Format with colors
	formattedDir := formatter.Dim(dir)
	formattedFile := formatter.Bold(formatter.Cyan(file))

	// Combine
	if dir == "" {
		return formattedFile
	}

	return fmt.Sprintf("%s/%s", formattedDir, formattedFile)
}

// formatTestCounts formats test counts with color-coded results
func formatTestCounts(formatter colors.FormatterInterface, total, passed, failed, skipped int) string {
	var parts []string

	// Format as "(X tests | Y passed | Z failed | ...)"
	totalPart := fmt.Sprintf("%d %s", total, pluralize("test", total))
	parts = append(parts, totalPart)

	// Add passed, failed, skipped counts with appropriate colors
	if passed > 0 {
		passedPart := fmt.Sprintf("%s %s", formatter.Green(fmt.Sprintf("%d", passed)), "passed")
		parts = append(parts, passedPart)
	}

	if failed > 0 {
		failedPart := fmt.Sprintf("%s %s", formatter.Red(fmt.Sprintf("%d", failed)), "failed")
		parts = append(parts, failedPart)
	}

	if skipped > 0 {
		skippedPart := fmt.Sprintf("%s %s", formatter.Yellow(fmt.Sprintf("%d", skipped)), "skipped")
		parts = append(parts, skippedPart)
	}

	return fmt.Sprintf("(%s)", strings.Join(parts, " | "))
}

// DurationFormatterInterface defines the interface for duration formatting
type DurationFormatterInterface interface {
	FormatDuration(d time.Duration) string
}

// DurationFormatter handles duration formatting
type DurationFormatter struct {
	formatter colors.FormatterInterface
}

// NewDurationFormatter creates a new DurationFormatter
func NewDurationFormatter(formatter colors.FormatterInterface) *DurationFormatter {
	return &DurationFormatter{
		formatter: formatter,
	}
}

// FormatDuration formats a duration with appropriate units and precision
func (d *DurationFormatter) FormatDuration(duration time.Duration) string {
	return FormatDuration(d.formatter, duration)
}

// FormatDuration formats a duration with appropriate units and precision (standalone function)
func FormatDuration(formatter colors.FormatterInterface, d time.Duration) string {
	var result string

	// Format based on duration magnitude
	switch {
	case d.Milliseconds() < 1:
		// Very small durations
		result = "0ms"
	case d < time.Second:
		// Milliseconds
		result = fmt.Sprintf("%dms", d.Milliseconds())
	case d < time.Minute:
		// Seconds with decimal
		seconds := float64(d) / float64(time.Second)
		result = fmt.Sprintf("%.1fs", seconds)
	default:
		// Minutes and seconds
		minutes := d / time.Minute
		seconds := (d % time.Minute) / time.Second
		result = fmt.Sprintf("%dm %ds", minutes, seconds)
	}

	return formatter.Gray(result)
}

// MemoryFormatterInterface defines the interface for memory usage formatting
type MemoryFormatterInterface interface {
	FormatMemoryUsage(bytes uint64) string
}

// MemoryFormatter handles memory usage formatting
type MemoryFormatter struct {
	formatter colors.FormatterInterface
}

// NewMemoryFormatter creates a new MemoryFormatter
func NewMemoryFormatter(formatter colors.FormatterInterface) *MemoryFormatter {
	return &MemoryFormatter{
		formatter: formatter,
	}
}

// FormatMemoryUsage formats memory usage in appropriate units
func (m *MemoryFormatter) FormatMemoryUsage(bytes uint64) string {
	return FormatMemoryUsage(m.formatter, bytes)
}

// FormatMemoryUsage formats memory usage in appropriate units (standalone function)
func FormatMemoryUsage(formatter colors.FormatterInterface, bytes uint64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	var result string

	// Format based on magnitude
	switch {
	case bytes < KB:
		result = fmt.Sprintf("%d B", bytes)
	case bytes < MB:
		result = fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	case bytes < GB:
		result = fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	default:
		result = fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	}

	// Remove .0 suffix for whole numbers
	result = strings.Replace(result, ".0 ", " ", 1)

	return formatter.Gray(fmt.Sprintf("%s heap used", result))
}

// pluralize returns singular or plural form based on count
func pluralize(word string, count int) string {
	if count == 1 {
		return word
	}
	return word + "s"
}

// Ensure HeaderRenderer implements RendererInterface
var _ RendererInterface = (*HeaderRenderer)(nil)
