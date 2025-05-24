package cli

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// IncrementalRenderer manages incremental test result rendering for watch mode
type IncrementalRenderer struct {
	writer      io.Writer
	formatter   *ColorFormatter
	icons       *IconProvider
	width       int
	lastResults map[string]*TestSuite
	lastStats   *TestRunStats
	cache       *TestResultCache
}

// NewIncrementalRenderer creates a new incremental renderer
func NewIncrementalRenderer(writer io.Writer, formatter *ColorFormatter, icons *IconProvider, width int, cache *TestResultCache) *IncrementalRenderer {
	return &IncrementalRenderer{
		writer:      writer,
		formatter:   formatter,
		icons:       icons,
		width:       width,
		lastResults: make(map[string]*TestSuite),
		cache:       cache,
	}
}

// RenderIncrementalResults renders only changed test results
func (r *IncrementalRenderer) RenderIncrementalResults(currentSuites map[string]*TestSuite, currentStats *TestRunStats, changes []*FileChange) error {
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
func (r *IncrementalRenderer) renderChangesSummary(changes []*FileChange) error {
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
func (r *IncrementalRenderer) identifyChangedSuites(currentSuites map[string]*TestSuite) []string {
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
func (r *IncrementalRenderer) suiteHasChanged(lastSuite, currentSuite *TestSuite) bool {
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

	// Check individual test status changes
	lastTestMap := make(map[string]*TestResult)
	for _, test := range lastSuite.Tests {
		lastTestMap[test.Name] = test
	}

	for _, currentTest := range currentSuite.Tests {
		lastTest, existed := lastTestMap[currentTest.Name]
		if !existed || lastTest.Status != currentTest.Status {
			return true
		}
	}

	return false
}

// renderSuiteChange renders changes for a specific test suite
func (r *IncrementalRenderer) renderSuiteChange(suitePath string, suite *TestSuite) error {
	lastSuite := r.lastResults[suitePath]

	if lastSuite == nil {
		// New suite - render normally
		return r.renderNewSuite(suitePath, suite)
	}

	// Compare and render changes
	return r.renderSuiteComparison(suitePath, lastSuite, suite)
}

// renderNewSuite renders a completely new test suite
func (r *IncrementalRenderer) renderNewSuite(suitePath string, suite *TestSuite) error {
	fmt.Fprintf(r.writer, "%s %s\n", r.icons.GetIcon("package"), suitePath)

	for _, test := range suite.Tests {
		icon := r.getTestStatusIcon(test.Status)
		color := r.getTestStatusColor(test.Status)
		duration := formatDuration(test.Duration)

		fmt.Fprintf(r.writer, "  %s %s %s\n",
			icon,
			r.formatter.Colorize(test.Name, color),
			r.formatter.Colorize(duration, "dim"))
	}

	fmt.Fprintln(r.writer)
	return nil
}

// renderSuiteComparison renders changes between old and new test suite
func (r *IncrementalRenderer) renderSuiteComparison(suitePath string, lastSuite, currentSuite *TestSuite) error {
	fmt.Fprintf(r.writer, "%s %s\n", r.icons.GetIcon("package"), suitePath)

	// Create maps for easy lookup
	lastTestMap := make(map[string]*TestResult)
	for _, test := range lastSuite.Tests {
		lastTestMap[test.Name] = test
	}

	hasChanges := false

	for _, currentTest := range currentSuite.Tests {
		lastTest, existed := lastTestMap[currentTest.Name]

		if !existed {
			// New test
			icon := r.getTestStatusIcon(currentTest.Status)
			color := r.getTestStatusColor(currentTest.Status)
			duration := formatDuration(currentTest.Duration)

			fmt.Fprintf(r.writer, "  %s %s %s %s\n",
				r.icons.GetIcon("new"),
				icon,
				r.formatter.Colorize(currentTest.Name, color),
				r.formatter.Colorize(duration, "dim"))
			hasChanges = true
		} else if lastTest.Status != currentTest.Status {
			// Status changed
			oldIcon := r.getTestStatusIcon(lastTest.Status)
			newIcon := r.getTestStatusIcon(currentTest.Status)
			color := r.getTestStatusColor(currentTest.Status)
			duration := formatDuration(currentTest.Duration)

			fmt.Fprintf(r.writer, "  %s %s → %s %s %s\n",
				r.icons.GetIcon("change"),
				oldIcon,
				newIcon,
				r.formatter.Colorize(currentTest.Name, color),
				r.formatter.Colorize(duration, "dim"))
			hasChanges = true
		}
	}

	if !hasChanges {
		fmt.Fprintf(r.writer, "  %s No changes\n", r.icons.GetIcon("unchanged"))
	}

	fmt.Fprintln(r.writer)
	return nil
}

// renderIncrementalSummary renders a summary of changes
func (r *IncrementalRenderer) renderIncrementalSummary(currentStats *TestRunStats, changedSuites []string) error {
	if r.lastStats == nil {
		// First run - render full summary
		return r.renderFullSummary(currentStats)
	}

	// Calculate deltas
	deltaTests := currentStats.TotalTests - r.lastStats.TotalTests
	deltaPassed := currentStats.PassedTests - r.lastStats.PassedTests
	deltaFailed := currentStats.FailedTests - r.lastStats.FailedTests
	deltaSkipped := currentStats.SkippedTests - r.lastStats.SkippedTests

	// Render summary header
	fmt.Fprintf(r.writer, "%s", r.formatter.Colorize(strings.Repeat("─", r.width), "dim"))
	fmt.Fprintln(r.writer)

	fmt.Fprintf(r.writer, "%s Incremental Test Results\n", r.icons.GetIcon("summary"))

	// Show changes
	if deltaTests != 0 {
		fmt.Fprintf(r.writer, "Tests: %s%+d (%d total)\n",
			r.getDeltaColor(deltaTests), deltaTests, currentStats.TotalTests)
	}

	if deltaPassed != 0 {
		fmt.Fprintf(r.writer, "Passed: %s%+d (%d total)\n",
			r.getDeltaColor(deltaPassed), deltaPassed, currentStats.PassedTests)
	}

	if deltaFailed != 0 {
		fmt.Fprintf(r.writer, "Failed: %s%+d (%d total)\n",
			r.getDeltaColor(deltaFailed), deltaFailed, currentStats.FailedTests)
	}

	if deltaSkipped != 0 {
		fmt.Fprintf(r.writer, "Skipped: %s%+d (%d total)\n",
			r.getDeltaColor(deltaSkipped), deltaSkipped, currentStats.SkippedTests)
	}

	// Duration and timing
	fmt.Fprintf(r.writer, "Duration: %v\n", currentStats.Duration)
	fmt.Fprintf(r.writer, "Changed suites: %d\n", len(changedSuites))

	fmt.Fprintf(r.writer, "%s", r.formatter.Colorize(strings.Repeat("─", r.width), "dim"))
	fmt.Fprintln(r.writer)

	return nil
}

// renderFullSummary renders a complete summary (for first run)
func (r *IncrementalRenderer) renderFullSummary(stats *TestRunStats) error {
	summaryRenderer := NewSummaryRenderer(r.writer, r.formatter, r.icons, r.width)
	return summaryRenderer.RenderSummary(stats)
}

// updateLastResults updates the cached results
func (r *IncrementalRenderer) updateLastResults(currentSuites map[string]*TestSuite, currentStats *TestRunStats) {
	// Deep copy current results for comparison next time
	r.lastResults = make(map[string]*TestSuite)
	for path, suite := range currentSuites {
		r.lastResults[path] = suite // Note: this is a shallow copy, which should be fine for our use case
	}
	r.lastStats = currentStats
}

// Helper methods

func (r *IncrementalRenderer) getChangeIcon(changeType ChangeType) string {
	switch changeType {
	case ChangeTypeTest:
		return r.icons.GetIcon("test")
	case ChangeTypeSource:
		return r.icons.GetIcon("code")
	case ChangeTypeConfig:
		return r.icons.GetIcon("config")
	case ChangeTypeDependency:
		return r.icons.GetIcon("dependency")
	default:
		return r.icons.GetIcon("file")
	}
}

func (r *IncrementalRenderer) getChangeTypeString(changeType ChangeType) string {
	switch changeType {
	case ChangeTypeTest:
		return "test file"
	case ChangeTypeSource:
		return "source file"
	case ChangeTypeConfig:
		return "config file"
	case ChangeTypeDependency:
		return "dependency"
	default:
		return "file"
	}
}

func (r *IncrementalRenderer) getTestStatusIcon(status TestStatus) string {
	switch status {
	case StatusPassed:
		return r.icons.GetIcon("pass")
	case StatusFailed:
		return r.icons.GetIcon("fail")
	case StatusSkipped:
		return r.icons.GetIcon("skip")
	default:
		return r.icons.GetIcon("unknown")
	}
}

func (r *IncrementalRenderer) getTestStatusColor(status TestStatus) string {
	switch status {
	case StatusPassed:
		return "green"
	case StatusFailed:
		return "red"
	case StatusSkipped:
		return "yellow"
	default:
		return "white"
	}
}

func (r *IncrementalRenderer) getDeltaColor(delta int) string {
	if delta > 0 {
		return r.formatter.Colorize("", "green")
	} else if delta < 0 {
		return r.formatter.Colorize("", "red")
	}
	return ""
}

// formatDuration is a simple wrapper around FormatDuration for backward compatibility
func formatDuration(d time.Duration) string {
	// Create a dummy formatter for the existing FormatDuration function
	formatter := &ColorFormatter{useColors: false}
	return FormatDuration(formatter, d)
}

// SummaryRenderer renders test summary information
type SummaryRenderer struct {
	writer    io.Writer
	formatter *ColorFormatter
	icons     *IconProvider
	width     int
}

// NewSummaryRenderer creates a new SummaryRenderer
func NewSummaryRenderer(writer io.Writer, formatter *ColorFormatter, icons *IconProvider, width int) *SummaryRenderer {
	return &SummaryRenderer{
		writer:    writer,
		formatter: formatter,
		icons:     icons,
		width:     width,
	}
}

// RenderSummary renders a test run summary
func (r *SummaryRenderer) RenderSummary(stats *TestRunStats) error {
	// Render test files summary
	fmt.Fprintf(r.writer, "Test Files  %s\n", r.formatFileStats(stats))

	// Render tests summary
	fmt.Fprintf(r.writer, "Tests       %s\n", r.formatTestStats(stats))

	// Render timing information
	fmt.Fprintf(r.writer, "Start at    %s\n", stats.StartTime.Format("15:04:05"))
	fmt.Fprintf(r.writer, "Duration    %s\n", FormatDuration(r.formatter, stats.Duration))

	return nil
}

// formatFileStats formats file statistics
func (r *SummaryRenderer) formatFileStats(stats *TestRunStats) string {
	var parts []string

	total := stats.TotalFiles
	parts = append(parts, fmt.Sprintf("%d total", total))

	if stats.PassedFiles > 0 {
		parts = append(parts, r.formatter.Green(fmt.Sprintf("%d passed", stats.PassedFiles)))
	}

	if stats.FailedFiles > 0 {
		parts = append(parts, r.formatter.Red(fmt.Sprintf("%d failed", stats.FailedFiles)))
	}

	return strings.Join(parts, " | ")
}

// formatTestStats formats test statistics
func (r *SummaryRenderer) formatTestStats(stats *TestRunStats) string {
	var parts []string

	total := stats.TotalTests
	parts = append(parts, fmt.Sprintf("%d total", total))

	if stats.PassedTests > 0 {
		parts = append(parts, r.formatter.Green(fmt.Sprintf("%d passed", stats.PassedTests)))
	}

	if stats.FailedTests > 0 {
		parts = append(parts, r.formatter.Red(fmt.Sprintf("%d failed", stats.FailedTests)))
	}

	if stats.SkippedTests > 0 {
		parts = append(parts, r.formatter.Yellow(fmt.Sprintf("%d skipped", stats.SkippedTests)))
	}

	return strings.Join(parts, " | ")
}
