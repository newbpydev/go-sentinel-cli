package runner

import (
	"context"
	"fmt"
	"io"
	"runtime"
	"sync"
	"time"

	"github.com/newbpydev/go-sentinel/internal/test/processor"
	"github.com/newbpydev/go-sentinel/pkg/models"
)

// OptimizedTestProcessor provides thread-safe test processing with performance optimizations
type OptimizedTestProcessor struct {
	mu              sync.RWMutex
	outputMu        sync.Mutex // Separate mutex for output synchronization
	output          io.Writer
	processor       *processor.TestProcessor
	memoryPool      sync.Pool
	renderOptimized bool
	maxConcurrency  int
}

// MemoryStats represents memory usage statistics
type MemoryStats struct {
	AllocBytes      uint64
	TotalAllocBytes uint64
	SysBytes        uint64
	NumGC           uint32
	LastGCTime      time.Time
}

// OptimizedStreamParser provides buffered stream parsing
type OptimizedStreamParser struct {
	bufferSize     int
	maxLineLength  int
	reusableBuffer []byte
	mu             sync.Mutex
}

// BatchProcessor processes test results in batches for efficiency
type BatchProcessor struct {
	batchSize int
	timeout   time.Duration
	buffer    []*models.LegacyTestResult
	mu        sync.Mutex
}

// LazyRenderer provides lazy rendering for large test suites
type LazyRenderer struct {
	threshold       int // Number of tests before switching to lazy mode
	summaryOnly     bool
	detailsOnDemand bool
}

// ProcessorInterface defines the interface for test processors
type ProcessorInterface interface {
	AddTestSuite(suite *models.TestSuite)
	GetStats() *models.TestRunStats
	GetSuites() map[string]*models.TestSuite
}

// NewOptimizedTestProcessor creates a new thread-safe optimized test processor
func NewOptimizedTestProcessor(output io.Writer, proc *processor.TestProcessor) *OptimizedTestProcessor {
	return &OptimizedTestProcessor{
		output:    output,
		processor: proc,
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
func (p *OptimizedTestProcessor) AddTestSuite(suite *models.TestSuite) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.processor != nil {
		p.processor.AddTestSuite(suite)
	}
}

// GetStats returns current statistics with thread safety
func (p *OptimizedTestProcessor) GetStats() *models.TestRunStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.processor != nil {
		return p.processor.GetStats()
	}
	return &models.TestRunStats{}
}

// GetSuites returns the test suites with thread safety
func (p *OptimizedTestProcessor) GetSuites() map[string]*models.TestSuite {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.processor != nil {
		return p.processor.GetSuites()
	}
	return make(map[string]*models.TestSuite)
}

// GetStatsOptimized returns current statistics with thread safety (alias for GetStats)
func (p *OptimizedTestProcessor) GetStatsOptimized() *models.TestRunStats {
	return p.GetStats()
}

// RenderResultsOptimized renders results with performance optimizations
func (p *OptimizedTestProcessor) RenderResultsOptimized(autoCollapse bool) error {
	// Create context for cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get a snapshot of suites with minimal lock time
	var suites []*models.TestSuite
	func() {
		p.mu.RLock()
		defer p.mu.RUnlock()
		testSuites := p.GetSuites()
		suites = make([]*models.TestSuite, 0, len(testSuites))
		for _, suite := range testSuites {
			suites = append(suites, suite)
		}
	}()

	if len(suites) > 10 {
		return p.renderWithWorkerPool(ctx, suites, autoCollapse)
	}

	// For smaller numbers, render sequentially
	return p.renderSequentially(suites, autoCollapse)
}

// renderWithWorkerPool renders test suites using a worker pool
func (p *OptimizedTestProcessor) renderWithWorkerPool(ctx context.Context, suites []*models.TestSuite, autoCollapse bool) error {
	// Create work channel
	work := make(chan *models.TestSuite, len(suites))
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
					bufferInterface := p.memoryPool.Get()
					var buffer []byte
					if bufferInterface != nil {
						if buf, ok := bufferInterface.([]byte); ok {
							buffer = buf
						} else if bufPtr, ok := bufferInterface.(*[]byte); ok {
							buffer = *bufPtr
						} else {
							// Fallback: create new buffer if type assertion fails
							buffer = make([]byte, 0, 1024)
						}
					} else {
						buffer = make([]byte, 0, 1024)
					}
					buffer = buffer[:0] // Reset length but keep capacity

					// Render with output synchronization only
					// This ensures thread safety for output while allowing parallel processing
					func() {
						p.outputMu.Lock()
						defer p.outputMu.Unlock()
						// Simplified rendering - just output suite info
						if suite != nil {
							fmt.Fprintf(p.output, "Suite: %s (%d tests)\n", suite.FilePath, suite.TestCount)
						}
					}()

					// Return buffer to pool
					p.memoryPool.Put(&buffer)
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
func (p *OptimizedTestProcessor) renderSequentially(suites []*models.TestSuite, autoCollapse bool) error {
	for _, suite := range suites {
		if suite != nil {
			fmt.Fprintf(p.output, "Suite: %s (%d tests)\n", suite.FilePath, suite.TestCount)
		}
	}
	return nil
}

// Clear clears all test data
func (p *OptimizedTestProcessor) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.processor != nil {
		p.processor.Reset()
	}
}

// GetMemoryStats returns current memory usage statistics
func (p *OptimizedTestProcessor) GetMemoryStats() MemoryStats {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return MemoryStats{
		AllocBytes:      memStats.Alloc,
		TotalAllocBytes: memStats.TotalAlloc,
		SysBytes:        memStats.Sys,
		NumGC:           memStats.NumGC,
		LastGCTime:      time.Unix(0, int64(memStats.LastGC)),
	}
}

// ForceGarbageCollection forces a garbage collection cycle
func (p *OptimizedTestProcessor) ForceGarbageCollection() {
	runtime.GC()
}

// NewOptimizedStreamParser creates a new optimized stream parser
func NewOptimizedStreamParser() *OptimizedStreamParser {
	return &OptimizedStreamParser{
		bufferSize:     8192,
		maxLineLength:  65536,
		reusableBuffer: make([]byte, 8192),
	}
}

// ParseOptimized parses a reader stream with optimizations
func (p *OptimizedStreamParser) ParseOptimized(reader io.Reader, results chan<- *models.LegacyTestResult) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Simple implementation - just read and process
	// In a full implementation, this would use buffered reading and parsing
	buffer := make([]byte, p.bufferSize)
	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			// Process the buffer content
			// This is a simplified version - would need actual JSON parsing
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}

	return nil
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(batchSize int, timeout time.Duration) *BatchProcessor {
	return &BatchProcessor{
		batchSize: batchSize,
		timeout:   timeout,
		buffer:    make([]*models.LegacyTestResult, 0, batchSize),
	}
}

// Add adds a test result to the batch
func (bp *BatchProcessor) Add(result *models.LegacyTestResult) []*models.LegacyTestResult {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	bp.buffer = append(bp.buffer, result)

	if len(bp.buffer) >= bp.batchSize {
		// Return the batch and reset
		batch := make([]*models.LegacyTestResult, len(bp.buffer))
		copy(batch, bp.buffer)
		bp.buffer = bp.buffer[:0]
		return batch
	}

	return nil
}

// Flush returns all remaining results in the batch
func (bp *BatchProcessor) Flush() []*models.LegacyTestResult {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	if len(bp.buffer) == 0 {
		return nil
	}

	batch := make([]*models.LegacyTestResult, len(bp.buffer))
	copy(batch, bp.buffer)
	bp.buffer = bp.buffer[:0]
	return batch
}

// NewLazyRenderer creates a new lazy renderer
func NewLazyRenderer(threshold int) *LazyRenderer {
	return &LazyRenderer{
		threshold:       threshold,
		summaryOnly:     false,
		detailsOnDemand: true,
	}
}

// ShouldUseLazyMode determines if lazy mode should be used
func (lr *LazyRenderer) ShouldUseLazyMode(testCount int) bool {
	return testCount > lr.threshold
}

// RenderSummaryOnly renders only a summary of the test suite
func (lr *LazyRenderer) RenderSummaryOnly(suite *models.TestSuite) string {
	if suite == nil {
		return "No test suite data available"
	}

	return fmt.Sprintf("Suite: %s - %d tests (%d passed, %d failed, %d skipped)",
		suite.FilePath,
		suite.TestCount,
		suite.PassedCount,
		suite.FailedCount,
		suite.SkippedCount,
	)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
