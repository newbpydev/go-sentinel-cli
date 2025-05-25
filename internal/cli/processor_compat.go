package cli

import (
	"io"
	"strings"
	"time"

	"github.com/newbpydev/go-sentinel/internal/test/cache"
	"github.com/newbpydev/go-sentinel/internal/test/processor"
	"github.com/newbpydev/go-sentinel/internal/test/runner"
	"github.com/newbpydev/go-sentinel/internal/ui/colors"
	"github.com/newbpydev/go-sentinel/internal/ui/display"
	"github.com/newbpydev/go-sentinel/internal/watch/coordinator"
	"github.com/newbpydev/go-sentinel/internal/watch/core"
	"github.com/newbpydev/go-sentinel/internal/watch/debouncer"
	"github.com/newbpydev/go-sentinel/internal/watch/watcher"
	"github.com/newbpydev/go-sentinel/pkg/models"
)

// Re-export types from internal/test/processor for backward compatibility during migration
// These will be removed once all files are migrated to use internal/test/processor directly

// TestProcessor re-exports processor.TestProcessor
type TestProcessor = processor.TestProcessor

// SourceExtractor re-exports processor.SourceExtractor
type SourceExtractor = processor.SourceExtractor

// Parser re-exports processor.Parser
type Parser = processor.Parser

// StreamParser re-exports processor.StreamParser
type StreamParser = processor.StreamParser

// Re-export types from internal/test/runner for backward compatibility

// TestRunner re-exports runner.TestRunner (BasicTestRunner)
type TestRunner = runner.TestRunner

// TestRunnerInterface re-exports runner.TestRunnerInterface
type TestRunnerInterface = runner.TestRunnerInterface

// ParallelTestRunner re-exports runner.ParallelTestRunner
type ParallelTestRunner = runner.ParallelTestRunner

// ParallelTestResult re-exports runner.ParallelTestResult
type ParallelTestResult = runner.ParallelTestResult

// OptimizedTestRunner re-exports runner.OptimizedTestRunner
type OptimizedTestRunner = runner.OptimizedTestRunner

// SmartTestCache re-exports runner.SmartTestCache
type SmartTestCache = runner.SmartTestCache

// OptimizedTestResult re-exports runner.OptimizedTestResult
type OptimizedTestResult = runner.OptimizedTestResult

// TestExecutionResult re-exports runner.TestExecutionResult
type TestExecutionResult = runner.TestExecutionResult

// FileChangeInterface re-exports runner.FileChangeInterface
type FileChangeInterface = runner.FileChangeInterface

// FileChangeAdapter re-exports runner.FileChangeAdapter
type FileChangeAdapter = runner.FileChangeAdapter

// OptimizedTestProcessor re-exports runner.OptimizedTestProcessor
type OptimizedTestProcessor = runner.OptimizedTestProcessor

// MemoryStats re-exports runner.MemoryStats
type MemoryStats = runner.MemoryStats

// OptimizedStreamParser re-exports runner.OptimizedStreamParser
type OptimizedStreamParser = runner.OptimizedStreamParser

// BatchProcessor re-exports runner.BatchProcessor
type BatchProcessor = runner.BatchProcessor

// LazyRenderer re-exports runner.LazyRenderer
type LazyRenderer = runner.LazyRenderer

// Re-export types from internal/test/cache for backward compatibility during TIER 5 migration

// TestResultCache re-exports cache.TestResultCache
type TestResultCache = cache.TestResultCache

// CachedTestResult re-exports cache.CachedTestResult
type CachedTestResult = cache.CachedTestResult

// ChangeType re-exports cache.ChangeType
type ChangeType = cache.ChangeType

// FileChange re-exports cache.FileChange
type FileChange = cache.FileChange

// CacheInterface re-exports cache.CacheInterface
type CacheInterface = cache.CacheInterface

// Re-export constants from cache package
const (
	ChangeTypeTest       = cache.ChangeTypeTest
	ChangeTypeSource     = cache.ChangeTypeSource
	ChangeTypeConfig     = cache.ChangeTypeConfig
	ChangeTypeDependency = cache.ChangeTypeDependency
)

// Re-export config types for backward compatibility

// Re-export constructor functions

// NewTestProcessor re-exports processor.NewTestProcessor
func NewTestProcessor(writer io.Writer, formatter *ColorFormatter, icons *IconProvider, width int) *TestProcessor {
	return processor.NewTestProcessor(writer, formatter, icons, width)
}

// NewTestRunner creates a new test runner for backward compatibility
func NewTestRunner(verbose, jsonOutput bool) *TestRunner {
	return runner.NewTestRunner(verbose, jsonOutput)
}

// NewParallelTestRunner creates a new parallel test runner
func NewParallelTestRunner(maxConcurrency int, testRunner *TestRunner, cache *TestResultCache) *ParallelTestRunner {
	// We need to adapt the cache interface here
	var cacheAdapter runner.CacheInterface
	if cache != nil {
		cacheAdapter = &cacheAdapterImpl{cache: cache}
	}
	return runner.NewParallelTestRunner(maxConcurrency, testRunner, cacheAdapter)
}

// cacheAdapterImpl adapts TestResultCache to runner.CacheInterface
type cacheAdapterImpl struct {
	cache *TestResultCache
}

func (c *cacheAdapterImpl) GetCachedResult(testPath string) (*runner.CachedResult, bool) {
	if result, exists := c.cache.GetCachedResult(testPath); exists {
		return &runner.CachedResult{Suite: result.Suite}, true
	}
	return nil, false
}

func (c *cacheAdapterImpl) CacheResult(testPath string, suite *models.TestSuite) {
	c.cache.CacheResult(testPath, suite)
}

// MergeResults merges parallel test results
func MergeResults(processor *TestProcessor, results []*ParallelTestResult) {
	runner.MergeResults(processor, results)
}

// NewSourceExtractor re-exports processor.NewSourceExtractor
func NewSourceExtractor() *SourceExtractor {
	return processor.NewSourceExtractor()
}

// NewParser re-exports processor.NewParser
func NewParser() *Parser {
	return processor.NewParser()
}

// NewStreamParser re-exports processor.NewStreamParser
func NewStreamParser() *StreamParser {
	return processor.NewStreamParser()
}

// NewOptimizedTestRunner creates a new optimized test runner
func NewOptimizedTestRunner() *OptimizedTestRunner {
	return runner.NewOptimizedTestRunner()
}

// NewSmartTestCache creates a new smart test cache
func NewSmartTestCache() *SmartTestCache {
	return runner.NewSmartTestCache()
}

// NewOptimizedTestProcessor creates a new optimized test processor (new signature)
func NewOptimizedTestProcessor(output io.Writer, proc *TestProcessor) *OptimizedTestProcessor {
	return runner.NewOptimizedTestProcessor(output, proc)
}

// NewOptimizedTestProcessorWithUI creates a new optimized test processor with UI components (legacy signature)
func NewOptimizedTestProcessorWithUI(output io.Writer, formatter *ColorFormatter, icons *IconProvider, terminalWidth int) *OptimizedTestProcessor {
	// Create a basic test processor first
	proc := processor.NewTestProcessor(output, formatter, icons, terminalWidth)
	return runner.NewOptimizedTestProcessor(output, proc)
}

// NewOptimizedStreamParser creates a new optimized stream parser
func NewOptimizedStreamParser() *OptimizedStreamParser {
	return runner.NewOptimizedStreamParser()
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(batchSize int, timeout time.Duration) *BatchProcessor {
	return runner.NewBatchProcessor(batchSize, timeout)
}

// NewLazyRenderer creates a new lazy renderer
func NewLazyRenderer(threshold int) *LazyRenderer {
	return runner.NewLazyRenderer(threshold)
}

// NewTestResultCache creates a new test result cache
func NewTestResultCache() *TestResultCache {
	return cache.NewTestResultCache()
}

// AdaptFileChanges converts CLI FileChange slice to FileChangeInterface slice
func AdaptFileChanges(changes []*FileChange) []FileChangeInterface {
	if changes == nil {
		return nil
	}

	result := make([]FileChangeInterface, len(changes))
	for i, change := range changes {
		// Convert CLI FileChange to models.FileChange first, then wrap in adapter
		modelChange := &models.FileChange{
			FilePath:   change.Path,
			ChangeType: adaptChangeType(change.Type),
			Timestamp:  change.Timestamp,
			Checksum:   change.Hash,
			Metadata: map[string]interface{}{
				"is_new":         change.IsNew,
				"affected_tests": change.AffectedTests,
			},
		}
		result[i] = &FileChangeAdapter{FileChange: modelChange}
	}
	return result
}

// adaptChangeType converts CLI ChangeType to models ChangeType
func adaptChangeType(cliType ChangeType) models.ChangeType {
	switch cliType {
	case ChangeTypeTest, ChangeTypeSource, ChangeTypeConfig, ChangeTypeDependency:
		// These map to Created/Modified based on whether file is new
		return models.ChangeTypeModified
	default:
		return models.ChangeTypeModified
	}
}

// AdaptCacheFileChanges converts cache FileChange slice to runner FileChangeInterface slice
func AdaptCacheFileChanges(changes []*cache.FileChange) []FileChangeInterface {
	if changes == nil {
		return nil
	}

	result := make([]FileChangeInterface, len(changes))
	for i, change := range changes {
		// Convert cache FileChange to models.FileChange first, then wrap in adapter
		modelChange := &models.FileChange{
			FilePath:   change.Path,
			ChangeType: adaptCacheChangeType(change.Type),
			Timestamp:  change.Timestamp,
			Checksum:   change.Hash,
			Metadata: map[string]interface{}{
				"is_new":         change.IsNew,
				"affected_tests": change.AffectedTests,
			},
		}
		result[i] = &FileChangeAdapter{FileChange: modelChange}
	}
	return result
}

// adaptCacheChangeType converts cache ChangeType to models ChangeType
func adaptCacheChangeType(cacheType cache.ChangeType) models.ChangeType {
	switch cacheType {
	case cache.ChangeTypeTest, cache.ChangeTypeSource, cache.ChangeTypeConfig, cache.ChangeTypeDependency:
		// These map to Created/Modified based on whether file is new
		return models.ChangeTypeModified
	default:
		return models.ChangeTypeModified
	}
}

// Helper functions re-exported from runner package

// IsGoTestFile returns true if the file is a Go test file
func IsGoTestFile(path string) bool {
	return runner.IsGoTestFile(path)
}

// IsGoFile returns true if the file is a Go source file
func IsGoFile(path string) bool {
	return runner.IsGoFile(path)
}

// discardWriter is a writer that discards all writes (for testing)
type discardWriter struct{}

func (d *discardWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// Legacy type aliases for backward compatibility
// These map to the legacy models that match the old CLI structure

// TestEvent is an alias for models.TestEvent
type TestEvent = models.TestEvent

// TestProgress is an alias for models.TestProgress
type TestProgress = models.TestProgress

// TestRunStats is an alias for models.TestRunStats
type TestRunStats = models.TestRunStats

// TestSuite is an alias for models.TestSuite
type TestSuite = models.TestSuite

// TestResult is an alias for models.LegacyTestResult (maintains old structure)
type TestResult = models.LegacyTestResult

// TestError is an alias for models.LegacyTestError (maintains old structure)
type TestError = models.LegacyTestError

// TestStatus is an alias for models.TestStatus
type TestStatus = models.TestStatus

// SourceLocation is an alias for models.SourceLocation
type SourceLocation = models.SourceLocation

// TestPackage is an alias for models.TestPackage
type TestPackage = models.TestPackage

// FailedTestDetail is an alias for models.FailedTestDetail
type FailedTestDetail = models.FailedTestDetail

// TestProcessorInterface is an alias for models.TestProcessorInterface
type TestProcessorInterface = models.TestProcessorInterface

// Legacy status constants for backward compatibility
const (
	StatusPassed  = models.StatusPassed
	StatusFailed  = models.StatusFailed
	StatusSkipped = models.StatusSkipped
	StatusRunning = models.StatusRunning
)

// Re-export types from internal/watch/debouncer for backward compatibility
type FileEventDebouncer = debouncer.FileEventDebouncer

// Re-export types from internal/watch/core for backward compatibility
type FileEvent = core.FileEvent

// AdaptFileEventToCoreEvent converts CLI FileEvent to core.FileEvent
func AdaptFileEventToCoreEvent(event FileEvent) core.FileEvent {
	return core.FileEvent{
		Path:      event.Path,
		Type:      event.Type,
		Timestamp: event.Timestamp,
		IsTest:    event.IsTest,
	}
}

// AdaptCoreEventToFileEvent converts core.FileEvent to CLI FileEvent
func AdaptCoreEventToFileEvent(event core.FileEvent) FileEvent {
	return FileEvent{
		Path:      event.Path,
		Type:      event.Type,
		Timestamp: event.Timestamp,
		IsTest:    event.IsTest,
	}
}

// AdaptCoreEventsToFileEvents converts slice of core.FileEvent to CLI FileEvent
func AdaptCoreEventsToFileEvents(events []core.FileEvent) []FileEvent {
	result := make([]FileEvent, len(events))
	for i, event := range events {
		result[i] = AdaptCoreEventToFileEvent(event)
	}
	return result
}

// Re-export constructor functions for watch components
func NewFileEventDebouncer(interval time.Duration) *debouncer.FileEventDebouncer {
	return debouncer.NewFileEventDebouncer(interval)
}

// Re-export types from internal/watch/watcher for backward compatibility
type FileWatcher = watcher.FileSystemWatcher
type TestFileFinder = watcher.TestFileFinder

func NewFileWatcher(paths []string, ignorePatterns []string) (*watcher.FileSystemWatcher, error) {
	return watcher.NewFileSystemWatcher(paths, ignorePatterns)
}

func NewTestFileFinder(rootDir string) *watcher.TestFileFinder {
	return watcher.NewTestFileFinder(rootDir)
}

// Re-export types from internal/watch/coordinator for backward compatibility
type TestWatchCoordinator = coordinator.TestWatchCoordinator

func NewTestWatchCoordinator(options core.WatchOptions) (*coordinator.TestWatchCoordinator, error) {
	return coordinator.NewTestWatchCoordinator(options)
}

// Re-export types from internal/ui/colors for backward compatibility

// Note: We cannot directly alias the old CLI types to new UI types because they have different interfaces
// Instead, we keep the original CLI types and add compatibility adapters where needed

// The original CLI ColorFormatter and IconProvider are defined in colors.go
// This compatibility layer provides bridges to the new modular UI components

// Re-export types from internal/ui/display for backward compatibility

// HeaderRenderer re-exports display.HeaderRenderer
// type HeaderRenderer = display.HeaderRenderer

// PathFormatter re-exports display.PathFormatter
// type PathFormatter = display.PathFormatter

// DurationFormatter re-exports display.DurationFormatter
// type DurationFormatter = display.DurationFormatter

// MemoryFormatter re-exports display.MemoryFormatter
// type MemoryFormatter = display.MemoryFormatter

// Re-export types from internal/ui/renderer for backward compatibility

// IncrementalRenderer re-exports renderer.IncrementalRenderer
// type IncrementalRenderer = renderer.IncrementalRenderer

// Re-export UI constructor functions

// NewColorFormatter creates a new ColorFormatter
// func NewColorFormatter(useColors bool) *ColorFormatter {
// 	return colors.NewColorFormatter(useColors)
// }

// NewIconProvider creates a new IconProvider
// func NewIconProvider(unicodeSupport bool) *IconProvider {
// 	return colors.NewIconProvider(unicodeSupport)
// }

// NewTerminalDetector creates a new TerminalDetector
// func NewTerminalDetector() *TerminalDetector {
// 	return colors.NewTerminalDetector()
// }

// NewHeaderRenderer creates a new HeaderRenderer
// func NewHeaderRenderer(writer io.Writer, formatter *ColorFormatter, icons *IconProvider, width int) *HeaderRenderer {
// 	return display.NewHeaderRenderer(writer, formatter, icons, width)
// }

// NewIncrementalRenderer creates a new IncrementalRenderer
// func NewIncrementalRenderer(writer io.Writer, formatter *ColorFormatter, icons *IconProvider, width int, cache *TestResultCache) *IncrementalRenderer {
// 	return renderer.NewIncrementalRenderer(writer, formatter, icons, width, cache)
// }

// FormatFilePath formats a file path with colorized components
// func FormatFilePath(formatter *ColorFormatter, path string) string {
// 	return display.FormatFilePath(formatter, path)
// }

// FormatDuration formats a duration with appropriate units
// func FormatDuration(formatter *ColorFormatter, d time.Duration) string {
// 	return display.FormatDuration(formatter, d)
// }

// FormatMemoryUsage formats memory usage in appropriate units
// func FormatMemoryUsage(formatter *ColorFormatter, bytes uint64) string {
// 	return display.FormatMemoryUsage(formatter, bytes)
// }

// FormatTestStatus formats a test status with appropriate coloring and icon
// func FormatTestStatus(status TestStatus, formatter *ColorFormatter, icons *IconProvider) string {
// 	return colors.FormatTestStatus(status, formatter, icons)
// }

// Re-export UI types from internal/ui for backward compatibility during TIER 7 migration
// Note: FailedTestRenderer is already declared in failed_tests.go, so we use a compatibility wrapper

// FailedTestRendererCompat provides backward compatibility for the old FailedTestRenderer
type FailedTestRendererCompat struct {
	writer    io.Writer
	formatter *ColorFormatter
	icons     *IconProvider
	width     int
	// New modular components
	failureRenderer display.FailureDisplayInterface
	errorFormatter  display.ErrorFormatterInterface
}

// NewFailedTestRendererCompat creates a new FailedTestRenderer compatibility wrapper
func NewFailedTestRendererCompat(writer io.Writer, formatter *ColorFormatter, icons *IconProvider, width int) *FailedTestRendererCompat {
	// Create the new modular color formatter and icon provider
	// Access the useColors field directly since old formatter doesn't have IsEnabled() method
	colorFormatter := colors.NewColorFormatter(formatter.useColors)
	// Access the unicodeSupport field directly since old icon provider doesn't have SupportsUnicode() method
	iconProvider := colors.NewIconProvider(icons.unicodeSupport)

	// Create the new modular components
	errorFormatter := display.NewErrorFormatterWithDefaults(writer, colorFormatter)
	failureRenderer := display.NewFailureRenderer(writer, colorFormatter, iconProvider, errorFormatter, width)

	return &FailedTestRendererCompat{
		writer:          writer,
		formatter:       formatter,
		icons:           icons,
		width:           width,
		failureRenderer: failureRenderer,
		errorFormatter:  errorFormatter,
	}
}

// RenderFailedTestsHeaderCompat renders the header for the failed tests section (compatibility method)
func (r *FailedTestRendererCompat) RenderFailedTestsHeaderCompat(failCount int) error {
	return r.failureRenderer.RenderFailedTestsHeader(failCount)
}

// RenderFailedTestCompat renders a detailed view of a failed test with source context (compatibility method)
func (r *FailedTestRendererCompat) RenderFailedTestCompat(test *TestResult) error {
	// Convert old TestResult to new models.TestResult
	newTest := r.convertTestResult(test)
	return r.failureRenderer.RenderFailedTest(newTest)
}

// RenderFailedTestsCompat renders a section with all failed tests (compatibility method)
func (r *FailedTestRendererCompat) RenderFailedTestsCompat(tests []*TestResult) error {
	// Convert old TestResult slice to new models.TestResult slice
	newTests := make([]*models.TestResult, len(tests))
	for i, test := range tests {
		newTests[i] = r.convertTestResult(test)
	}
	return r.failureRenderer.RenderFailedTests(newTests)
}

// getTerminalWidthCompat returns the current terminal width or default (compatibility method)
func (r *FailedTestRendererCompat) getTerminalWidthCompat() int {
	return r.failureRenderer.GetTerminalWidth()
}

// convertTestResult converts old CLI TestResult to new models.TestResult
func (r *FailedTestRendererCompat) convertTestResult(oldTest *TestResult) *models.TestResult {
	if oldTest == nil {
		return nil
	}

	newTest := &models.TestResult{
		Name:     oldTest.Name,
		Status:   models.TestStatus(oldTest.Status),
		Duration: oldTest.Duration,
		Package:  oldTest.Package,
		// Convert single string Output to []string slice
		Output: []string{oldTest.Output},
	}

	// Convert error information if present
	if oldTest.Error != nil {
		newTest.Error = &models.TestError{
			Message:    oldTest.Error.Message,
			Type:       oldTest.Error.Type,
			StackTrace: strings.Split(oldTest.Error.Stack, "\n"),
			Expected:   oldTest.Error.Expected,
			Actual:     oldTest.Error.Actual,
		}

		// Convert location information
		if oldTest.Error.Location != nil {
			newTest.Error.SourceFile = oldTest.Error.Location.File
			newTest.Error.SourceLine = oldTest.Error.Location.Line
			newTest.Error.SourceColumn = oldTest.Error.Location.Column
		}

		// Convert source context information
		if len(oldTest.Error.SourceContext) > 0 {
			newTest.Error.SourceContext = oldTest.Error.SourceContext
			// Calculate context start line based on highlighted line
			if oldTest.Error.HighlightedLine >= 0 && oldTest.Error.HighlightedLine < len(oldTest.Error.SourceContext) {
				if oldTest.Error.Location != nil {
					newTest.Error.ContextStartLine = oldTest.Error.Location.Line - oldTest.Error.HighlightedLine
				}
			}
		}
	}

	return newTest
}
