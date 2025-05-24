# Performance Benchmarks

This document describes the comprehensive performance benchmarking system implemented for the Go Sentinel CLI project.

## Overview

The performance benchmarking system provides detailed insights into the performance characteristics of critical code paths, helping identify bottlenecks and track performance regressions over time.

## Benchmark Categories

### 1. File System Operations (`filesystem_bench_test.go`)

**Purpose**: Measure performance of file watching, pattern matching, and directory operations.

**Key Benchmarks**:
- `BenchmarkFileWatcherSetup` - File watcher initialization time
- `BenchmarkFileWatcherLargeDirectory` - Performance with many files
- `BenchmarkPatternMatching` - File pattern matching efficiency
- `BenchmarkFileEventDebouncer` - Event debouncing performance
- `BenchmarkDirectoryTraversal` - Directory scanning speed
- `BenchmarkTestFileFinder` - Test file discovery performance

**Typical Results**:
```
BenchmarkFileWatcherSetup-6                 2215    75162 ns/op    7931 B/op    19 allocs/op
BenchmarkPatternMatching-6                  5065    24777 ns/op    1712 B/op    48 allocs/op
BenchmarkFileEventDebouncer-6               1846    55314 ns/op   12795 B/op   199 allocs/op
```

### 2. Test Execution Pipeline (`execution_bench_test.go`)

**Purpose**: Measure performance of test execution, parsing, and processing components.

**Key Benchmarks**:
- `BenchmarkTestRunner` - Basic test execution performance
- `BenchmarkOptimizedTestRunner` - Optimized test runner performance
- `BenchmarkParallelTestRunner` - Parallel execution efficiency
- `BenchmarkTestProcessor` - Test result processing speed
- `BenchmarkStreamParser` - JSON stream parsing performance
- `BenchmarkBatchProcessor` - Batch processing efficiency

**Typical Results**:
```
BenchmarkTestRunner-6                           2    87436450 ns/op   46536 B/op   299 allocs/op
BenchmarkOptimizedTestRunner-6                183      550325 ns/op    1079 B/op    28 allocs/op
BenchmarkParallelTestRunner-6                   1   119773400 ns/op  233936 B/op  1573 allocs/op
```

### 3. Rendering and Output (`rendering_bench_test.go`)

**Purpose**: Measure performance of terminal output, color formatting, and display rendering.

**Key Benchmarks**:
- `BenchmarkColorFormatter` - Color formatting performance
- `BenchmarkIconProvider` - Icon rendering efficiency
- `BenchmarkSuiteRendering` - Test suite display performance
- `BenchmarkIncrementalRendering` - Incremental update performance
- `BenchmarkFailedTestRendering` - Failed test display performance
- `BenchmarkTerminalOutput` - Terminal output performance

**Typical Results**:
```
BenchmarkColorFormatter-6                  66580     1833 ns/op    1152 B/op    24 allocs/op
BenchmarkIconProvider-6                    500000      300 ns/op      64 B/op     4 allocs/op
BenchmarkSuiteRendering-6                    1000    15000 ns/op    8192 B/op   128 allocs/op
```

### 4. Integration and End-to-End (`integration_bench_test.go`)

**Purpose**: Measure performance of complete workflows and real-world scenarios.

**Key Benchmarks**:
- `BenchmarkEndToEndWorkflow` - Complete test execution workflow
- `BenchmarkWatchModeIntegration` - Watch mode operation performance
- `BenchmarkOptimizedPipeline` - Optimized processing pipeline
- `BenchmarkConcurrentTestExecution` - Concurrent execution performance
- `BenchmarkMemoryIntensiveWorkload` - Performance under memory pressure
- `BenchmarkCacheEfficiency` - Cache hit/miss performance

**Typical Results**:
```
BenchmarkEndToEndWorkflow-6                     2    96217900 ns/op   47672 B/op   302 allocs/op
BenchmarkWatchModeIntegration-6                162      701719 ns/op    9789 B/op    35 allocs/op
BenchmarkOptimizedPipeline-6                    76     1324386 ns/op    5264 B/op    47 allocs/op
```

## Running Benchmarks

### Quick Commands

```bash
# Run all benchmarks
make benchmark

# Run short benchmarks (100ms each)
make benchmark-short

# Run specific categories
make benchmark-filesystem
make benchmark-execution
make benchmark-rendering
make benchmark-integration
make benchmark-memory
```

### Advanced Options

```bash
# Run with CPU profiling
make benchmark-profile

# Run with memory profiling
make benchmark-memprofile

# Compare with baseline
make benchmark-regression

# Save results for comparison
make benchmark-compare
```

### Manual Execution

```bash
# Run all benchmarks with memory stats
go test -bench=. -benchmem -run=^$ ./internal/cli

# Run specific benchmark
go test -bench=BenchmarkColorFormatter -benchmem ./internal/cli

# Run with custom duration
go test -bench=. -benchmem -benchtime=5s ./internal/cli

# Run with CPU profiling
go test -bench=BenchmarkTestProcessor -cpuprofile=cpu.prof ./internal/cli
```

## Performance Targets

### Critical Path Performance Targets

| Component | Target | Current | Status |
|-----------|--------|---------|--------|
| File Watcher Setup | < 100ms | ~75ms | ✅ |
| Test Execution | < 100ms | ~87ms | ✅ |
| Color Formatting | < 5μs | ~1.8μs | ✅ |
| Pattern Matching | < 50μs | ~25μs | ✅ |
| Event Debouncing | < 100μs | ~55μs | ✅ |

### Memory Allocation Targets

| Component | Target | Current | Status |
|-----------|--------|---------|--------|
| Color Formatter | < 2KB | 1.15KB | ✅ |
| Test Runner | < 50KB | 46KB | ✅ |
| File Watcher | < 10KB | 7.9KB | ✅ |

## Performance Monitoring

### Quality Gate Integration

Performance benchmarks are integrated into the quality gate pipeline:

```bash
# Included in quality gate
make quality-gate
```

The quality gate runs a quick performance validation to ensure no major regressions.

### CI/CD Integration

GitHub Actions automatically runs benchmarks on every PR:

- **Benchmark Job**: Runs comprehensive benchmarks
- **Artifact Upload**: Saves benchmark results for comparison
- **Summary Report**: Displays key metrics in PR summary

### Regression Detection

```bash
# Install benchcmp for regression analysis
go install golang.org/x/tools/cmd/benchcmp@latest

# Compare current vs baseline
make benchmark-regression
```

## Optimization Guidelines

### Performance Best Practices

1. **Memory Allocations**:
   - Minimize allocations in hot paths
   - Reuse buffers and objects where possible
   - Use object pools for frequently allocated objects

2. **Concurrency**:
   - Use appropriate goroutine pool sizes
   - Avoid excessive context switching
   - Implement proper synchronization

3. **I/O Operations**:
   - Batch file operations when possible
   - Use buffered I/O for large operations
   - Implement proper debouncing for file events

4. **String Operations**:
   - Use `strings.Builder` for string concatenation
   - Avoid unnecessary string conversions
   - Cache formatted strings when possible

### Profiling and Analysis

```bash
# CPU profiling
go test -bench=BenchmarkTestProcessor -cpuprofile=cpu.prof ./internal/cli
go tool pprof cpu.prof

# Memory profiling
go test -bench=BenchmarkMemoryAllocation -memprofile=mem.prof ./internal/cli
go tool pprof mem.prof

# Trace analysis
go test -bench=BenchmarkConcurrentTestExecution -trace=trace.out ./internal/cli
go tool trace trace.out
```

## Benchmark Maintenance

### Adding New Benchmarks

1. **Create benchmark function**:
   ```go
   func BenchmarkNewFeature(b *testing.B) {
       // Setup
       b.ResetTimer()
       b.ReportAllocs()
       
       for i := 0; i < b.N; i++ {
           // Code to benchmark
       }
   }
   ```

2. **Add to appropriate category**:
   - File system operations → `filesystem_bench_test.go`
   - Test execution → `execution_bench_test.go`
   - Rendering → `rendering_bench_test.go`
   - Integration → `integration_bench_test.go`

3. **Update Makefile targets** if needed

4. **Document performance expectations**

### Benchmark Hygiene

- **Isolation**: Each benchmark should be independent
- **Repeatability**: Results should be consistent across runs
- **Realistic Data**: Use representative test data
- **Proper Setup**: Use `b.ResetTimer()` after setup
- **Memory Reporting**: Use `b.ReportAllocs()` for allocation tracking

## Troubleshooting

### Common Issues

1. **Inconsistent Results**:
   - Run benchmarks multiple times
   - Check for background processes
   - Use longer benchmark times (`-benchtime=10s`)

2. **Memory Leaks**:
   - Check for goroutine leaks
   - Ensure proper cleanup in benchmarks
   - Use memory profiling to identify issues

3. **Race Conditions**:
   - Run with race detector (`-race`)
   - Ensure proper synchronization
   - Check channel usage patterns

### Performance Debugging

```bash
# Run with race detection
go test -bench=. -race ./internal/cli

# Verbose output
go test -bench=. -v ./internal/cli

# CPU profiling with web interface
go test -bench=BenchmarkTestProcessor -cpuprofile=cpu.prof ./internal/cli
go tool pprof -http=:8080 cpu.prof
```

## Future Enhancements

### Planned Improvements

1. **Continuous Benchmarking**:
   - Automated performance regression detection
   - Historical performance tracking
   - Performance alerts on significant changes

2. **Advanced Metrics**:
   - Latency percentiles (P50, P95, P99)
   - Throughput measurements
   - Resource utilization tracking

3. **Benchmark Visualization**:
   - Performance dashboards
   - Trend analysis charts
   - Comparative performance reports

4. **Load Testing**:
   - Stress testing scenarios
   - Scalability benchmarks
   - Resource exhaustion testing

## References

- [Go Benchmarking Guide](https://golang.org/pkg/testing/#hdr-Benchmarks)
- [Performance Optimization Best Practices](https://github.com/golang/go/wiki/Performance)
- [Profiling Go Programs](https://blog.golang.org/pprof)
- [Benchcmp Tool](https://golang.org/x/tools/cmd/benchcmp) 