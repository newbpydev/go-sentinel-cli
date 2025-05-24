package cli

import (
	"io"

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
