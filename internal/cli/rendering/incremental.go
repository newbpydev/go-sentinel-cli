package rendering

import (
	"fmt"
	"io"
	"time"

	"github.com/newbpydev/go-sentinel/internal/cli/core"
)

// IncrementalRenderer manages incremental test result rendering for watch mode
// This preserves the exact visual style from the original implementation
type IncrementalRenderer struct {
	writer      io.Writer
	formatter   *ColorFormatter
	icons       *IconProvider
	width       int
	lastResults map[string]*core.TestSuite
	lastStats   *core.TestRunStats
	verbose     bool
}

// NewIncrementalRenderer creates a new incremental renderer
func NewIncrementalRenderer(writer io.Writer, useColors bool, verbose bool) *IncrementalRenderer {
	terminal := NewTerminalDetector()
	width := terminal.Width()

	// Auto-detect colors if not explicitly set
	if useColors && !terminal.SupportsColor() {
		useColors = false
	}

	return &IncrementalRenderer{
		writer:      writer,
		formatter:   NewColorFormatter(useColors),
		icons:       NewIconProvider(true), // Unicode support by default
		width:       width,
		lastResults: make(map[string]*core.TestSuite),
		verbose:     verbose,
	}
}

// RenderIncrementalResults renders only changed test results
func (r *IncrementalRenderer) RenderIncrementalResults(currentSuites map[string]*core.TestSuite, currentStats *core.TestRunStats, changes []core.FileChange) error {
	// Render file changes summary
	if err := r.renderChangesSummary(changes); err != nil {
		return err
	}

	// If no test suites were provided, this means no tests were run
	if len(currentSuites) == 0 {
		fmt.Fprintf(r.writer, "%s No test changes detected - tests not needed\n\n", r.icons.GetIcon("info"))
		return nil
	}

	// Identify changed suites
	changedSuites := r.identifyChangedSuites(currentSuites)

	if len(changedSuites) == 0 {
		fmt.Fprintf(r.writer, "%s No test result changes detected\n\n", r.icons.GetIcon("info"))
		return nil
	}

	// Render only changed suites
	for _, suitePath := range changedSuites {
		suite := currentSuites[suitePath]
		if err := r.renderSuiteChange(suitePath, suite); err != nil {
			return err
		}
	}

	// Render incremental summary
	if err := r.renderIncrementalSummary(currentStats, changedSuites); err != nil {
		return err
	}

	// Update cache
	r.updateLastResults(currentSuites, currentStats)

	return nil
}

// renderChangesSummary renders a summary of file changes
func (r *IncrementalRenderer) renderChangesSummary(changes []core.FileChange) error {
	if len(changes) == 0 {
		return nil
	}

	fmt.Fprintf(r.writer, "%s File changes detected:\n", r.icons.GetIcon("watch"))

	for _, change := range changes {
		icon := r.getChangeIcon(change.Type)
		changeType := r.getChangeTypeString(change.Type)

		fmt.Fprintf(r.writer, "   %s %s (%s)\n", icon, change.Path, changeType)
	}

	fmt.Fprintln(r.writer)
	return nil
}

// identifyChangedSuites identifies which test suites have changed results
func (r *IncrementalRenderer) identifyChangedSuites(currentSuites map[string]*core.TestSuite) []string {
	var changed []string

	for suitePath, currentSuite := range currentSuites {
		lastSuite, existed := r.lastResults[suitePath]

		if !existed || r.suiteHasChanged(lastSuite, currentSuite) {
			changed = append(changed, suitePath)
		}
	}

	return changed
}

// suiteHasChanged checks if a test suite has changed since last run
func (r *IncrementalRenderer) suiteHasChanged(lastSuite, currentSuite *core.TestSuite) bool {
	if lastSuite == nil || currentSuite == nil {
		return true
	}

	// Check basic counts
	if lastSuite.TestCount != currentSuite.TestCount ||
		lastSuite.PassedCount != currentSuite.PassedCount ||
		lastSuite.FailedCount != currentSuite.FailedCount ||
		lastSuite.SkippedCount != currentSuite.SkippedCount {
		return true
	}

	// Since core.TestSuite doesn't have individual test details,
	// we'll assume any change in counts means the suite changed
	return false
}

// renderSuiteChange renders changes for a specific test suite
func (r *IncrementalRenderer) renderSuiteChange(suitePath string, suite *core.TestSuite) error {
	lastSuite := r.lastResults[suitePath]

	if lastSuite == nil {
		// New suite - render normally
		return r.renderNewSuite(suitePath, suite)
	}

	// Compare and render changes
	return r.renderSuiteComparison(suitePath, lastSuite, suite)
}

// renderNewSuite renders a completely new test suite
func (r *IncrementalRenderer) renderNewSuite(suitePath string, suite *core.TestSuite) error {
	fmt.Fprintf(r.writer, "%s %s\n", r.icons.GetIcon("package"), suitePath)

	// Display suite-level information
	icon := r.getTestStatusIcon(suite.Status)
	color := r.getTestStatusColor(suite.Status)
	duration := r.formatDuration(suite.Duration)

	fmt.Fprintf(r.writer, "  %s %s tests %s\n",
		icon,
		r.formatter.Colorize(fmt.Sprintf("%d", suite.TestCount), color),
		r.formatter.Colorize(duration, "dim"))

	fmt.Fprintln(r.writer)
	return nil
}

// renderSuiteComparison renders changes between old and new test suite
func (r *IncrementalRenderer) renderSuiteComparison(suitePath string, lastSuite, currentSuite *core.TestSuite) error {
	fmt.Fprintf(r.writer, "%s %s\n", r.icons.GetIcon("package"), suitePath)

	hasChanges := false

	// Check for status changes
	if lastSuite.Status != currentSuite.Status {
		oldIcon := r.getTestStatusIcon(lastSuite.Status)
		newIcon := r.getTestStatusIcon(currentSuite.Status)
		color := r.getTestStatusColor(currentSuite.Status)
		duration := r.formatDuration(currentSuite.Duration)

		fmt.Fprintf(r.writer, "  %s %s→%s %s %s\n",
			r.icons.GetIcon("changed"),
			oldIcon,
			newIcon,
			r.formatter.Colorize(fmt.Sprintf("%d tests", currentSuite.TestCount), color),
			r.formatter.Colorize(duration, "dim"))
		hasChanges = true
	}

	// Check for count changes
	if lastSuite.TestCount != currentSuite.TestCount ||
		lastSuite.PassedCount != currentSuite.PassedCount ||
		lastSuite.FailedCount != currentSuite.FailedCount ||
		lastSuite.SkippedCount != currentSuite.SkippedCount {

		fmt.Fprintf(r.writer, "  %s Test counts changed: %d→%d passed, %d→%d failed, %d→%d skipped\n",
			r.icons.GetIcon("changed"),
			lastSuite.PassedCount, currentSuite.PassedCount,
			lastSuite.FailedCount, currentSuite.FailedCount,
			lastSuite.SkippedCount, currentSuite.SkippedCount)
		hasChanges = true
	}

	if !hasChanges {
		fmt.Fprintf(r.writer, "  %s No changes detected\n", r.icons.GetIcon("info"))
	}

	fmt.Fprintln(r.writer)
	return nil
}

// renderIncrementalSummary renders a summary of incremental changes
func (r *IncrementalRenderer) renderIncrementalSummary(currentStats *core.TestRunStats, changedSuites []string) error {
	if currentStats == nil {
		return nil
	}

	// Calculate differences
	var passedDelta, failedDelta, skippedDelta int
	if r.lastStats != nil {
		passedDelta = currentStats.PassedTests - r.lastStats.PassedTests
		failedDelta = currentStats.FailedTests - r.lastStats.FailedTests
		skippedDelta = currentStats.SkippedTests - r.lastStats.SkippedTests
	}

	// Render summary header
	fmt.Fprintf(r.writer, "%s Test Summary:\n", r.icons.GetIcon("summary"))

	// Render test counts with deltas
	if passedDelta != 0 {
		deltaColor := r.getDeltaColor(passedDelta)
		deltaText := r.formatDelta(passedDelta)
		fmt.Fprintf(r.writer, "  %s Passed: %d %s\n",
			r.formatter.Green(r.icons.CheckMark()),
			currentStats.PassedTests,
			r.formatter.Colorize(deltaText, deltaColor))
	} else {
		fmt.Fprintf(r.writer, "  %s Passed: %d\n",
			r.formatter.Green(r.icons.CheckMark()),
			currentStats.PassedTests)
	}

	if currentStats.FailedTests > 0 || failedDelta != 0 {
		if failedDelta != 0 {
			deltaColor := r.getDeltaColor(failedDelta)
			deltaText := r.formatDelta(failedDelta)
			fmt.Fprintf(r.writer, "  %s Failed: %d %s\n",
				r.formatter.Red(r.icons.Cross()),
				currentStats.FailedTests,
				r.formatter.Colorize(deltaText, deltaColor))
		} else {
			fmt.Fprintf(r.writer, "  %s Failed: %d\n",
				r.formatter.Red(r.icons.Cross()),
				currentStats.FailedTests)
		}
	}

	if currentStats.SkippedTests > 0 || skippedDelta != 0 {
		if skippedDelta != 0 {
			deltaColor := r.getDeltaColor(skippedDelta)
			deltaText := r.formatDelta(skippedDelta)
			fmt.Fprintf(r.writer, "  %s Skipped: %d %s\n",
				r.formatter.Yellow(r.icons.Skipped()),
				currentStats.SkippedTests,
				r.formatter.Colorize(deltaText, deltaColor))
		} else {
			fmt.Fprintf(r.writer, "  %s Skipped: %d\n",
				r.formatter.Yellow(r.icons.Skipped()),
				currentStats.SkippedTests)
		}
	}

	// Render duration
	duration := r.formatDuration(currentStats.Duration)
	fmt.Fprintf(r.writer, "  %s Duration: %s\n",
		r.icons.GetIcon("timer"),
		r.formatter.Colorize(duration, "dim"))

	fmt.Fprintln(r.writer)
	return nil
}

// updateLastResults updates the cache of last results
func (r *IncrementalRenderer) updateLastResults(currentSuites map[string]*core.TestSuite, currentStats *core.TestRunStats) {
	r.lastResults = make(map[string]*core.TestSuite)
	for k, v := range currentSuites {
		r.lastResults[k] = v
	}
	r.lastStats = currentStats
}

// Helper methods

// getChangeIcon returns an icon for a file change type
func (r *IncrementalRenderer) getChangeIcon(changeType core.ChangeType) string {
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
func (r *IncrementalRenderer) getChangeTypeString(changeType core.ChangeType) string {
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

// getTestStatusIcon returns an icon for a test status
func (r *IncrementalRenderer) getTestStatusIcon(status core.TestStatus) string {
	switch status {
	case core.StatusPassed:
		return r.formatter.Green(r.icons.CheckMark())
	case core.StatusFailed:
		return r.formatter.Red(r.icons.Cross())
	case core.StatusSkipped:
		return r.formatter.Yellow(r.icons.Skipped())
	case core.StatusRunning:
		return r.formatter.Blue(r.icons.GetIcon("refresh"))
	default:
		return r.formatter.Gray("?")
	}
}

// getTestStatusColor returns a color for a test status
func (r *IncrementalRenderer) getTestStatusColor(status core.TestStatus) string {
	switch status {
	case core.StatusPassed:
		return "green"
	case core.StatusFailed:
		return "red"
	case core.StatusSkipped:
		return "yellow"
	case core.StatusRunning:
		return "blue"
	default:
		return "gray"
	}
}

// getDeltaColor returns a color for a delta value
func (r *IncrementalRenderer) getDeltaColor(delta int) string {
	if delta > 0 {
		return "green"
	} else if delta < 0 {
		return "red"
	}
	return "gray"
}

// formatDelta formats a delta value with appropriate sign
func (r *IncrementalRenderer) formatDelta(delta int) string {
	if delta > 0 {
		return fmt.Sprintf("(+%d)", delta)
	} else if delta < 0 {
		return fmt.Sprintf("(%d)", delta)
	}
	return ""
}

// formatDuration formats a duration in a human-readable format
func (r *IncrementalRenderer) formatDuration(d time.Duration) string {
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
