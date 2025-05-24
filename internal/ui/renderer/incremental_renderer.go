// Package renderer provides incremental rendering capabilities for watch mode
package renderer

import (
	"fmt"
	"io"
	"time"

	"github.com/newbpydev/go-sentinel/internal/test/cache"
	"github.com/newbpydev/go-sentinel/internal/ui/colors"
	"github.com/newbpydev/go-sentinel/pkg/models"
)

// IncrementalRendererInterface defines the interface for incremental rendering
type IncrementalRendererInterface interface {
	// RenderIncrementalResults renders only changed test results
	RenderIncrementalResults(currentSuites map[string]*models.TestSuite, currentStats *models.TestRunStats, changes []*cache.FileChange) error

	// UpdateLastResults updates the stored last results
	UpdateLastResults(suites map[string]*models.TestSuite, stats *models.TestRunStats)

	// GetWriter returns the output writer
	GetWriter() io.Writer
}

// IncrementalRenderer manages incremental test result rendering for watch mode
type IncrementalRenderer struct {
	writer      io.Writer
	formatter   colors.FormatterInterface
	icons       colors.IconProviderInterface
	width       int
	lastResults map[string]*models.TestSuite
	lastStats   *models.TestRunStats
	cache       cache.CacheInterface
}

// NewIncrementalRenderer creates a new incremental renderer
func NewIncrementalRenderer(writer io.Writer, formatter colors.FormatterInterface, icons colors.IconProviderInterface, width int, cacheImpl cache.CacheInterface) *IncrementalRenderer {
	return &IncrementalRenderer{
		writer:      writer,
		formatter:   formatter,
		icons:       icons,
		width:       width,
		lastResults: make(map[string]*models.TestSuite),
		cache:       cacheImpl,
	}
}

// RenderIncrementalResults renders only changed test results
func (r *IncrementalRenderer) RenderIncrementalResults(currentSuites map[string]*models.TestSuite, currentStats *models.TestRunStats, changes []*cache.FileChange) error {
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
	r.UpdateLastResults(currentSuites, currentStats)

	return nil
}

// GetWriter returns the output writer
func (r *IncrementalRenderer) GetWriter() io.Writer {
	return r.writer
}

// UpdateLastResults updates the stored last results
func (r *IncrementalRenderer) UpdateLastResults(currentSuites map[string]*models.TestSuite, currentStats *models.TestRunStats) {
	// Deep copy current suites
	for path, suite := range currentSuites {
		suiteCopy := *suite
		testsCopy := make([]*models.LegacyTestResult, len(suite.Tests))
		for i, test := range suite.Tests {
			testCopy := *test
			testsCopy[i] = &testCopy
		}
		suiteCopy.Tests = testsCopy
		r.lastResults[path] = &suiteCopy
	}

	// Copy stats
	if currentStats != nil {
		statsCopy := *currentStats
		r.lastStats = &statsCopy
	}
}

// renderChangesSummary renders a summary of file changes
func (r *IncrementalRenderer) renderChangesSummary(changes []*cache.FileChange) error {
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
func (r *IncrementalRenderer) identifyChangedSuites(currentSuites map[string]*models.TestSuite) []string {
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
func (r *IncrementalRenderer) suiteHasChanged(lastSuite, currentSuite *models.TestSuite) bool {
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
	lastTestMap := make(map[string]*models.LegacyTestResult)
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
func (r *IncrementalRenderer) renderSuiteChange(suitePath string, suite *models.TestSuite) error {
	lastSuite := r.lastResults[suitePath]

	if lastSuite == nil {
		// New suite - render normally
		return r.renderNewSuite(suitePath, suite)
	}

	// Compare and render changes
	return r.renderSuiteComparison(suitePath, lastSuite, suite)
}

// renderNewSuite renders a completely new test suite
func (r *IncrementalRenderer) renderNewSuite(suitePath string, suite *models.TestSuite) error {
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
func (r *IncrementalRenderer) renderSuiteComparison(suitePath string, lastSuite, currentSuite *models.TestSuite) error {
	fmt.Fprintf(r.writer, "%s %s\n", r.icons.GetIcon("package"), suitePath)

	// Create maps for easy lookup
	lastTestMap := make(map[string]*models.LegacyTestResult)
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

// renderIncrementalSummary renders a summary for the incremental update
func (r *IncrementalRenderer) renderIncrementalSummary(currentStats *models.TestRunStats, changedSuites []string) error {
	if currentStats == nil {
		return nil
	}

	// Summary header
	fmt.Fprintf(r.writer, "%s Updated %d test %s\n",
		r.icons.GetIcon("summary"),
		len(changedSuites),
		pluralize("suite", len(changedSuites)))

	// Test counts with deltas if we have previous stats
	if r.lastStats != nil {
		passedDelta := currentStats.PassedTests - r.lastStats.PassedTests
		failedDelta := currentStats.FailedTests - r.lastStats.FailedTests
		skippedDelta := currentStats.SkippedTests - r.lastStats.SkippedTests

		fmt.Fprintf(r.writer, "  Tests:   %s", r.formatTestCountWithDelta("passed", currentStats.PassedTests, passedDelta))
		fmt.Fprintf(r.writer, " | %s", r.formatTestCountWithDelta("failed", currentStats.FailedTests, failedDelta))
		fmt.Fprintf(r.writer, " | %s", r.formatTestCountWithDelta("skipped", currentStats.SkippedTests, skippedDelta))
		fmt.Fprintln(r.writer)
	} else {
		// No deltas available
		fmt.Fprintf(r.writer, "  Tests:   %s passed | %s failed | %s skipped\n",
			r.formatter.Green(fmt.Sprintf("%d", currentStats.PassedTests)),
			r.formatter.Red(fmt.Sprintf("%d", currentStats.FailedTests)),
			r.formatter.Yellow(fmt.Sprintf("%d", currentStats.SkippedTests)))
	}

	fmt.Fprintln(r.writer)
	return nil
}

// formatTestCountWithDelta formats a test count with delta indicator
func (r *IncrementalRenderer) formatTestCountWithDelta(label string, count, delta int) string {
	var color string
	switch label {
	case "passed":
		color = "green"
	case "failed":
		color = "red"
	case "skipped":
		color = "yellow"
	default:
		color = "white"
	}

	if delta == 0 {
		return r.formatter.Colorize(fmt.Sprintf("%d %s", count, label), color)
	}

	deltaStr := ""
	if delta > 0 {
		deltaStr = fmt.Sprintf(" (+%d)", delta)
	} else {
		deltaStr = fmt.Sprintf(" (%d)", delta)
	}

	return r.formatter.Colorize(fmt.Sprintf("%d %s", count, label), color) +
		r.formatter.Colorize(deltaStr, r.getDeltaColor(delta))
}

// getChangeIcon returns an icon for a change type
func (r *IncrementalRenderer) getChangeIcon(changeType cache.ChangeType) string {
	switch changeType {
	case cache.ChangeTypeTest:
		return r.icons.GetIcon("test")
	case cache.ChangeTypeSource:
		return r.icons.GetIcon("code")
	case cache.ChangeTypeConfig:
		return r.icons.GetIcon("config")
	case cache.ChangeTypeDependency:
		return r.icons.GetIcon("dependency")
	default:
		return r.icons.GetIcon("file")
	}
}

// getChangeTypeString returns a string representation of a change type
func (r *IncrementalRenderer) getChangeTypeString(changeType cache.ChangeType) string {
	switch changeType {
	case cache.ChangeTypeTest:
		return "test file"
	case cache.ChangeTypeSource:
		return "source file"
	case cache.ChangeTypeConfig:
		return "config file"
	case cache.ChangeTypeDependency:
		return "dependency"
	default:
		return "file"
	}
}

// getTestStatusIcon returns an icon for a test status
func (r *IncrementalRenderer) getTestStatusIcon(status models.TestStatus) string {
	switch status {
	case models.StatusPassed:
		return r.icons.CheckMark()
	case models.StatusFailed:
		return r.icons.Cross()
	case models.StatusSkipped:
		return r.icons.Skipped()
	case models.StatusRunning:
		return r.icons.Running()
	default:
		return r.icons.GetIcon("unknown")
	}
}

// getTestStatusColor returns a color for a test status
func (r *IncrementalRenderer) getTestStatusColor(status models.TestStatus) string {
	switch status {
	case models.StatusPassed:
		return "green"
	case models.StatusFailed:
		return "red"
	case models.StatusSkipped:
		return "yellow"
	case models.StatusRunning:
		return "blue"
	default:
		return "white"
	}
}

// getDeltaColor returns a color for delta values
func (r *IncrementalRenderer) getDeltaColor(delta int) string {
	if delta > 0 {
		return "green"
	} else if delta < 0 {
		return "red"
	}
	return "dim"
}

// formatDuration formats a duration with appropriate precision
func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return "0ms"
	} else if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	} else {
		return fmt.Sprintf("%.1fs", float64(d)/float64(time.Second))
	}
}

// pluralize returns singular or plural form based on count
func pluralize(word string, count int) string {
	if count == 1 {
		return word
	}
	return word + "s"
}

// Ensure IncrementalRenderer implements IncrementalRendererInterface
var _ IncrementalRendererInterface = (*IncrementalRenderer)(nil)
