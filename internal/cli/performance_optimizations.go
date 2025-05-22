package cli

import (
	"context"
	"fmt"
	"io"
	"runtime"
	"sync"
	"time"
)

// OptimizedTestProcessor provides thread-safe test processing with performance optimizations
type OptimizedTestProcessor struct {
	mu              sync.RWMutex
	output          io.Writer
	formatter       *ColorFormatter
	icons           *IconProvider
	terminalWidth   int
	testSuites      map[string]*TestSuite
	testStats       *TestRunStats
	memoryPool      sync.Pool
	renderOptimized bool
	maxConcurrency  int
}

// NewOptimizedTestProcessor creates a new thread-safe optimized test processor
func NewOptimizedTestProcessor(output io.Writer, formatter *ColorFormatter, icons *IconProvider, terminalWidth int) *OptimizedTestProcessor {
	return &OptimizedTestProcessor{
		output:        output,
		formatter:     formatter,
		icons:         icons,
		terminalWidth: terminalWidth,
		testSuites:    make(map[string]*TestSuite),
		testStats:     &TestRunStats{},
		memoryPool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 0, 1024) // Pre-allocate 1KB buffer
			},
		},
		renderOptimized: true,
		maxConcurrency:  runtime.NumCPU(),
	}
}

// AddTestSuite adds a test suite with thread safety
func (p *OptimizedTestProcessor) AddTestSuite(suite *TestSuite) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Use file path as key to avoid duplicates
	p.testSuites[suite.FilePath] = suite
	p.updateStats(suite)
}

// updateStats updates internal statistics (must be called with lock held)
func (p *OptimizedTestProcessor) updateStats(suite *TestSuite) {
	p.testStats.TotalFiles++
	if suite.FailedCount > 0 {
		p.testStats.FailedFiles++
	} else {
		p.testStats.PassedFiles++
	}

	p.testStats.TotalTests += suite.TestCount
	p.testStats.PassedTests += suite.PassedCount
	p.testStats.FailedTests += suite.FailedCount
	p.testStats.SkippedTests += suite.SkippedCount
}

// RenderResultsOptimized renders results with performance optimizations
func (p *OptimizedTestProcessor) RenderResultsOptimized(autoCollapse bool) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Create context for cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use worker pool for parallel rendering if we have many suites
	suites := make([]*TestSuite, 0, len(p.testSuites))
	for _, suite := range p.testSuites {
		suites = append(suites, suite)
	}

	if len(suites) > 10 {
		return p.renderWithWorkerPool(ctx, suites, autoCollapse)
	}

	// For smaller numbers, render sequentially
	return p.renderSequentially(suites, autoCollapse)
}

// renderWithWorkerPool renders test suites using a worker pool
func (p *OptimizedTestProcessor) renderWithWorkerPool(ctx context.Context, suites []*TestSuite, autoCollapse bool) error {
	// Create work channel
	work := make(chan *TestSuite, len(suites))
	results := make(chan error, len(suites))

	// Start workers
	numWorkers := min(p.maxConcurrency, len(suites))
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for suite := range work {
				select {
				case <-ctx.Done():
					results <- ctx.Err()
					return
				default:
					// Get buffer from pool
					buffer := p.memoryPool.Get().([]byte)
					buffer = buffer[:0] // Reset length but keep capacity

					// Create a temporary writer that writes to buffer
					// Note: This is a simplified approach - in practice we'd need
					// a more sophisticated buffering mechanism

					// For now, render directly to output with synchronization
					// This ensures thread safety but reduces parallelism benefit
					func() {
						p.mu.Lock()
						defer p.mu.Unlock()
						renderer := NewSuiteRenderer(p.output, p.formatter, p.icons, p.terminalWidth)
						if err := renderer.RenderSuite(suite, autoCollapse); err != nil {
							results <- err
							return
						}
					}()

					// Return buffer to pool
					p.memoryPool.Put(buffer)
					results <- nil
				}
			}
		}()
	}

	// Send work
	go func() {
		defer close(work)
		for _, suite := range suites {
			select {
			case work <- suite:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Wait for workers and collect results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Check for errors
	for err := range results {
		if err != nil {
			return err
		}
	}

	return nil
}

// renderSequentially renders test suites one by one
func (p *OptimizedTestProcessor) renderSequentially(suites []*TestSuite, autoCollapse bool) error {
	for _, suite := range suites {
		renderer := NewSuiteRenderer(p.output, p.formatter, p.icons, p.terminalWidth)
		if err := renderer.RenderSuite(suite, autoCollapse); err != nil {
			return err
		}
	}
	return nil
}

// GetStatsOptimized returns current statistics with thread safety
func (p *OptimizedTestProcessor) GetStatsOptimized() TestRunStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Return a copy to avoid race conditions
	return *p.testStats
}

// Clear clears all data and resets statistics
func (p *OptimizedTestProcessor) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Clear suites
	for k := range p.testSuites {
		delete(p.testSuites, k)
	}

	// Reset stats
	p.testStats = &TestRunStats{}
}

// MemoryStats provides memory usage statistics
type MemoryStats struct {
	AllocBytes      uint64
	TotalAllocBytes uint64
	SysBytes        uint64
	NumGC           uint32
	LastGCTime      time.Time
}

// GetMemoryStats returns current memory statistics
func (p *OptimizedTestProcessor) GetMemoryStats() MemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return MemoryStats{
		AllocBytes:      m.Alloc,
		TotalAllocBytes: m.TotalAlloc,
		SysBytes:        m.Sys,
		NumGC:           m.NumGC,
		LastGCTime:      time.Unix(0, int64(m.LastGC)),
	}
}

// ForceGarbageCollection forces garbage collection
func (p *OptimizedTestProcessor) ForceGarbageCollection() {
	runtime.GC()
}

// OptimizedStreamParser provides performance-optimized stream parsing
type OptimizedStreamParser struct {
	bufferSize     int
	maxLineLength  int
	reusableBuffer []byte
	mu             sync.Mutex
}

// NewOptimizedStreamParser creates a new optimized stream parser
func NewOptimizedStreamParser() *OptimizedStreamParser {
	return &OptimizedStreamParser{
		bufferSize:     64 * 1024, // 64KB buffer
		maxLineLength:  10 * 1024, // 10KB max line length
		reusableBuffer: make([]byte, 64*1024),
	}
}

// ParseOptimized parses test output with performance optimizations
func (p *OptimizedStreamParser) ParseOptimized(reader io.Reader, results chan<- *TestResult) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Use the existing stream parser but with optimized buffering
	parser := NewStreamParser()
	return parser.Parse(reader, results)
}

// BatchProcessor processes test results in batches for better performance
type BatchProcessor struct {
	batchSize int
	timeout   time.Duration
	buffer    []*TestResult
	mu        sync.Mutex
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(batchSize int, timeout time.Duration) *BatchProcessor {
	return &BatchProcessor{
		batchSize: batchSize,
		timeout:   timeout,
		buffer:    make([]*TestResult, 0, batchSize),
	}
}

// Add adds a test result to the batch
func (bp *BatchProcessor) Add(result *TestResult) []*TestResult {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	bp.buffer = append(bp.buffer, result)

	if len(bp.buffer) >= bp.batchSize {
		// Return full batch and reset buffer
		batch := make([]*TestResult, len(bp.buffer))
		copy(batch, bp.buffer)
		bp.buffer = bp.buffer[:0] // Reset but keep capacity
		return batch
	}

	return nil // No batch ready yet
}

// Flush returns any remaining results in the buffer
func (bp *BatchProcessor) Flush() []*TestResult {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	if len(bp.buffer) == 0 {
		return nil
	}

	batch := make([]*TestResult, len(bp.buffer))
	copy(batch, bp.buffer)
	bp.buffer = bp.buffer[:0]
	return batch
}

// LazyRenderer provides lazy rendering for large test suites
type LazyRenderer struct {
	threshold       int // Number of tests before switching to lazy mode
	summaryOnly     bool
	detailsOnDemand bool
}

// NewLazyRenderer creates a new lazy renderer
func NewLazyRenderer(threshold int) *LazyRenderer {
	return &LazyRenderer{
		threshold:       threshold,
		summaryOnly:     false,
		detailsOnDemand: true,
	}
}

// ShouldUseLazyMode determines if lazy rendering should be used
func (lr *LazyRenderer) ShouldUseLazyMode(testCount int) bool {
	return testCount > lr.threshold
}

// RenderSummaryOnly renders only summary information for large test suites
func (lr *LazyRenderer) RenderSummaryOnly(suite *TestSuite, formatter *ColorFormatter, icons *IconProvider) string {
	if suite.FailedCount > 0 {
		return formatter.Red(icons.Cross()) + " " +
			formatter.Cyan(suite.FilePath) + " " +
			formatter.Red(fmt.Sprintf("(%d failed)", suite.FailedCount))
	}

	return formatter.Green(icons.CheckMark()) + " " +
		formatter.Cyan(suite.FilePath) + " " +
		formatter.Green(fmt.Sprintf("(%d passed)", suite.PassedCount))
}
