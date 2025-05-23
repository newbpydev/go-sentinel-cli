package rendering

import (
	"fmt"
	"io"
	"time"

	"github.com/newbpydev/go-sentinel/internal/cli/core"
)

// StructuredRenderer provides the original go-sentinel visual style
// while working with the new modular architecture
type StructuredRenderer struct {
	writer    io.Writer
	formatter *ColorFormatter
	icons     *IconProvider
	terminal  *TerminalDetector
	verbose   bool
}

// NewStructuredRenderer creates a renderer with original go-sentinel styling
func NewStructuredRenderer(writer io.Writer, useColors bool, verbose bool) *StructuredRenderer {
	terminal := NewTerminalDetector()

	// Auto-detect colors if not explicitly set
	if useColors && !terminal.SupportsColor() {
		useColors = false
	}

	return &StructuredRenderer{
		writer:    writer,
		formatter: NewColorFormatter(useColors),
		icons:     NewIconProvider(true), // Unicode support by default
		terminal:  terminal,
		verbose:   verbose,
	}
}

// RenderStartup renders the original startup message
func (r *StructuredRenderer) RenderStartup(optimized bool, optimizationMode string) {
	fmt.Fprintf(r.writer, "%s Running tests with go-sentinel...\n", r.icons.GetIcon("rocket"))

	if optimized {
		fmt.Fprintf(r.writer, "%s Optimized mode enabled (%s) - leveraging Go's built-in caching!\n",
			r.icons.GetIcon("lightning"), optimizationMode)
	}
	fmt.Fprintln(r.writer)
}

// RenderWatchStart renders the watch mode startup message
func (r *StructuredRenderer) RenderWatchStart() {
	fmt.Fprintf(r.writer, "%s Starting watch mode...\n", r.icons.GetIcon("watch"))
}

// RenderTestResult renders a single test result in the original format
func (r *StructuredRenderer) RenderTestResult(result *core.TestResult) {
	if result == nil {
		return
	}

	// Format duration
	duration := r.formatDuration(result.Duration)

	// Get status icon and color
	var icon, statusText string
	switch result.Status {
	case core.StatusPassed:
		icon = r.formatter.Green(r.icons.CheckMark())
		statusText = "passed"
	case core.StatusFailed:
		icon = r.formatter.Red(r.icons.Cross())
		statusText = "failed"
	case core.StatusSkipped:
		icon = r.formatter.Yellow(r.icons.Skipped())
		statusText = "skipped"
	default:
		icon = r.formatter.Gray("?")
		statusText = "unknown"
	}

	// Render main result line
	fmt.Fprintf(r.writer, "%s Tests %s", icon, statusText)

	// Add cache indicator
	if result.CacheHit {
		fmt.Fprintf(r.writer, " %s", r.formatter.Dim("(cached)"))
	}

	// Add duration
	if result.Duration > 0 {
		fmt.Fprintf(r.writer, " %s", r.formatter.Dim(fmt.Sprintf("in %s", duration)))
	}

	fmt.Fprintln(r.writer)

	// Show detailed output if verbose or failed
	if (r.verbose || result.Status == core.StatusFailed) && result.Output != "" {
		fmt.Fprintln(r.writer)
		fmt.Fprintf(r.writer, "%s Test Output %s\n",
			r.formatter.Dim("---"),
			r.formatter.Dim("---"))
		fmt.Fprintln(r.writer, result.Output)
	}
}

// RenderCacheStats renders cache statistics in verbose mode
func (r *StructuredRenderer) RenderCacheStats(stats core.CacheStats) {
	if !r.verbose {
		return
	}

	fmt.Fprintf(r.writer, "\n%s Cache Statistics:\n", r.icons.GetIcon("info"))
	fmt.Fprintf(r.writer, "   Total entries: %d\n", stats.TotalEntries)
	fmt.Fprintf(r.writer, "   Valid entries: %d\n", stats.ValidEntries)
	fmt.Fprintf(r.writer, "   Hit rate: %.1f%%\n", stats.HitRate)
}

// RenderFileChanges renders file changes detected in watch mode
func (r *StructuredRenderer) RenderFileChanges(changes []core.FileChange) {
	if len(changes) == 0 {
		return
	}

	fmt.Fprintf(r.writer, "%s File changes detected:\n", r.icons.GetIcon("files"))

	for _, change := range changes {
		icon := r.getChangeIcon(change.Type)
		changeType := r.getChangeTypeString(change.Type)

		fmt.Fprintf(r.writer, "   %s %s (%s)\n", icon, change.Path, changeType)
	}

	fmt.Fprintln(r.writer)
}

// RenderNoTestsNeeded renders the message when no tests are needed
func (r *StructuredRenderer) RenderNoTestsNeeded() {
	fmt.Fprintf(r.writer, "%s No test changes detected - tests not needed\n\n",
		r.icons.GetIcon("info"))
}

// RenderCompletion renders the completion message with timing
func (r *StructuredRenderer) RenderCompletion(duration time.Duration) {
	fmt.Fprintf(r.writer, "\n%s Tests completed in %s\n",
		r.icons.GetIcon("timer"),
		r.formatDuration(duration))
}

// RenderWatchModeInfo renders watch mode information
func (r *StructuredRenderer) RenderWatchModeInfo() {
	fmt.Fprintf(r.writer, "%s Watching for file changes... (Press Ctrl+C to stop)\n",
		r.icons.GetIcon("watch"))
}

// RenderError renders an error message with appropriate styling
func (r *StructuredRenderer) RenderError(err error) {
	fmt.Fprintf(r.writer, "%s Error: %v\n",
		r.formatter.Red(r.icons.Cross()),
		err)
}

// Helper methods

// formatDuration formats a duration in a human-readable format
func (r *StructuredRenderer) formatDuration(d time.Duration) string {
	if d == 0 {
		return "0ms"
	}

	switch {
	case d < time.Millisecond:
		return fmt.Sprintf("%.0fμs", float64(d.Nanoseconds())/1000)
	case d < time.Second:
		return fmt.Sprintf("%.0fms", float64(d.Nanoseconds())/1000000)
	case d < time.Minute:
		return fmt.Sprintf("%.1fs", d.Seconds())
	default:
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
}

// getChangeIcon returns an icon for a file change type
func (r *StructuredRenderer) getChangeIcon(changeType core.ChangeType) string {
	switch changeType {
	case core.ChangeTypeTest:
		return r.icons.GetIcon("modified")
	case core.ChangeTypeSource:
		return r.icons.GetIcon("modified")
	case core.ChangeTypeConfig:
		return r.icons.GetIcon("modified")
	case core.ChangeTypeDependency:
		return r.icons.GetIcon("modified")
	default:
		return r.icons.GetIcon("modified")
	}
}

// getChangeTypeString returns a string description of a change type
func (r *StructuredRenderer) getChangeTypeString(changeType core.ChangeType) string {
	switch changeType {
	case core.ChangeTypeTest:
		return "test"
	case core.ChangeTypeSource:
		return "source"
	case core.ChangeTypeConfig:
		return "config"
	case core.ChangeTypeDependency:
		return "dependency"
	default:
		return "unknown"
	}
}
