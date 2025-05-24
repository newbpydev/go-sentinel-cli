package cli

import (
	"io"
	"time"

	"github.com/newbpydev/go-sentinel/internal/test/cache"
	"github.com/newbpydev/go-sentinel/internal/test/processor"
	"github.com/newbpydev/go-sentinel/internal/test/runner"
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
