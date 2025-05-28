# 📦 Benchmarks Package

[![Test Coverage](https://img.shields.io/badge/coverage-100.0%25-brightgreen.svg)](https://github.com/newbpydev/go-sentinel/tree/main/internal/test/benchmarks)
[![Go Report Card](https://goreportcard.com/badge/github.com/newbpydev/go-sentinel/internal/test/benchmarks)](https://goreportcard.com/report/github.com/newbpydev/go-sentinel/internal/test/benchmarks)
[![Go Reference](https://pkg.go.dev/badge/github.com/newbpydev/go-sentinel/internal/test/benchmarks.svg)](https://pkg.go.dev/github.com/newbpydev/go-sentinel/internal/test/benchmarks)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## 📖 Overview

The benchmarks package provides comprehensive performance monitoring and regression detection capabilities for Go benchmark results. It enables automated tracking of performance changes, baseline comparison, and intelligent alerting for performance regressions and improvements.

### 🎯 Key Features

- **Benchmark Parsing**: Robust parsing of Go benchmark output with support for various formats
- **Baseline Management**: Automatic baseline creation, storage, and comparison capabilities
- **Regression Detection**: Intelligent detection of performance regressions with configurable thresholds
- **Severity Classification**: Automatic classification of regressions as CRITICAL, MAJOR, or MINOR
- **Improvement Tracking**: Detection and reporting of performance improvements
- **Report Generation**: Human-readable text and machine-readable JSON report generation
- **Recommendation Engine**: Actionable recommendations for addressing performance issues

## 🏗️ Architecture

This package follows clean architecture principles with clear separation of concerns:

- **Single Responsibility**: Focuses exclusively on benchmark performance monitoring
- **Dependency Inversion**: Provides interfaces for extensible monitoring capabilities
- **Factory Pattern**: Clean instantiation through `NewPerformanceMonitor`
- **Strategy Pattern**: Configurable thresholds and severity calculation strategies

### 📦 Package Structure

```
internal/test/benchmarks/
├── monitor.go              # Core performance monitoring implementation
├── monitor_test.go         # Comprehensive test suite (100% coverage)
├── performance_test.go     # Performance and memory leak tests
├── rendering_bench_test.go # Rendering performance benchmarks
├── integration_bench_test.go # Integration benchmarks
├── filesystem_bench_test.go # File system operation benchmarks
├── execution_bench_test.go # Test execution benchmarks
└── README.md              # This documentation
```

## 🚀 Quick Start

### Basic Performance Monitoring

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "github.com/newbpydev/go-sentinel/internal/test/benchmarks"
)

func main() {
    // Create performance monitor
    monitor := benchmarks.NewPerformanceMonitor("baseline.json")
    
    // Parse benchmark output
    benchmarkOutput := `BenchmarkExample-8    1000000    1234 ns/op    456 B/op    7 allocs/op`
    results, err := monitor.ParseBenchmarkOutput(benchmarkOutput)
    if err != nil {
        log.Fatal("Failed to parse benchmark output:", err)
    }
    
    // Compare with baseline and generate report
    report, err := monitor.CompareWithBaseline(results)
    if err != nil {
        log.Fatal("Failed to compare with baseline:", err)
    }
    
    // Generate human-readable report
    err = monitor.GenerateTextReport(report, os.Stdout)
    if err != nil {
        log.Fatal("Failed to generate report:", err)
    }
}
```

### Custom Threshold Configuration

```go
// Configure custom regression thresholds
monitor := benchmarks.NewPerformanceMonitor("baseline.json")

customThresholds := benchmarks.RegressionThresholds{
    MaxSlowdownPercent: 15.0, // 15% slowdown threshold
    MaxMemoryIncrease:  20.0, // 20% memory increase threshold
    MinSampleSize:      5,    // Minimum 5 samples for comparison
}

monitor.SetThresholds(customThresholds)
```

## 🔧 Core Interfaces

### PerformanceMonitor

The main performance monitoring interface providing all benchmark analysis functionality:

```go
type PerformanceMonitor struct {
    baselineFile string
    thresholds   RegressionThresholds
}

// Core monitoring methods
func NewPerformanceMonitor(baselineFile string) *PerformanceMonitor
func (pm *PerformanceMonitor) SetThresholds(thresholds RegressionThresholds)
func (pm *PerformanceMonitor) ParseBenchmarkOutput(output string) ([]BenchmarkResult, error)
func (pm *PerformanceMonitor) CompareWithBaseline(currentResults []BenchmarkResult) (*PerformanceReport, error)
func (pm *PerformanceMonitor) SaveBaseline(results []BenchmarkResult) error
func (pm *PerformanceMonitor) GenerateTextReport(report *PerformanceReport, output io.Writer) error
func (pm *PerformanceMonitor) GenerateJSONReport(report *PerformanceReport, output io.Writer) error
```

### Data Structures

#### BenchmarkResult
```go
type BenchmarkResult struct {
    Name        string    `json:"name"`
    Iterations  int       `json:"iterations"`
    NsPerOp     float64   `json:"ns_per_op"`
    BytesPerOp  int64     `json:"bytes_per_op"`
    AllocsPerOp int64     `json:"allocs_per_op"`
    MBPerSec    float64   `json:"mb_per_sec,omitempty"`
    Timestamp   time.Time `json:"timestamp"`
    GitCommit   string    `json:"git_commit,omitempty"`
    GoVersion   string    `json:"go_version,omitempty"`
    OS          string    `json:"os,omitempty"`
    Arch        string    `json:"arch,omitempty"`
}
```

#### RegressionAlert
```go
type RegressionAlert struct {
    BenchmarkName   string  `json:"benchmark_name"`
    Severity        string  `json:"severity"` // "CRITICAL", "MAJOR", "MINOR"
    SlowdownPercent float64 `json:"slowdown_percent"`
    MemoryIncrease  float64 `json:"memory_increase_percent"`
    PreviousNsPerOp float64 `json:"previous_ns_per_op"`
    CurrentNsPerOp  float64 `json:"current_ns_per_op"`
    PreviousMemory  int64   `json:"previous_memory"`
    CurrentMemory   int64   `json:"current_memory"`
    Recommendation  string  `json:"recommendation"`
}
```

## 🔄 Advanced Usage

### Automated CI/CD Integration

```go
// CI/CD pipeline integration example
func monitorPerformanceInCI() error {
    monitor := benchmarks.NewPerformanceMonitor("ci-baseline.json")
    
    // Set strict thresholds for CI
    monitor.SetThresholds(benchmarks.RegressionThresholds{
        MaxSlowdownPercent: 10.0, // Strict 10% threshold
        MaxMemoryIncrease:  15.0, // Strict 15% memory threshold
        MinSampleSize:      3,
    })
    
    // Parse CI benchmark results
    results, err := monitor.ParseBenchmarkOutput(os.Getenv("BENCHMARK_OUTPUT"))
    if err != nil {
        return fmt.Errorf("failed to parse benchmark output: %w", err)
    }
    
    // Compare and generate report
    report, err := monitor.CompareWithBaseline(results)
    if err != nil {
        return fmt.Errorf("failed to compare with baseline: %w", err)
    }
    
    // Fail CI if critical regressions detected
    if report.Summary.CriticalRegressions > 0 {
        return fmt.Errorf("critical performance regressions detected: %d", 
            report.Summary.CriticalRegressions)
    }
    
    return nil
}
```

### Trend Analysis and Historical Tracking

```go
// Historical performance tracking
func trackPerformanceTrends() {
    monitor := benchmarks.NewPerformanceMonitor("historical-baseline.json")
    
    // Parse multiple benchmark runs
    for _, benchmarkFile := range []string{"run1.txt", "run2.txt", "run3.txt"} {
        output, _ := os.ReadFile(benchmarkFile)
        results, _ := monitor.ParseBenchmarkOutput(string(output))
        
        report, _ := monitor.CompareWithBaseline(results)
        
        // Log trend information
        fmt.Printf("Run %s: %s trend, %d regressions, %d improvements\n",
            benchmarkFile,
            report.Summary.OverallTrend,
            report.Summary.TotalRegressions,
            report.Summary.TotalImprovements)
        
        // Update baseline for next comparison
        monitor.SaveBaseline(results)
    }
}
```

## 🧪 Testing

The package achieves **100% test coverage** through comprehensive TDD implementation:

### Running Tests

```bash
# Run all tests
go test ./internal/test/benchmarks/...

# Run with coverage
go test ./internal/test/benchmarks/... -coverprofile=coverage.out

# View coverage report
go tool cover -html=coverage.out

# Run specific test categories
go test ./internal/test/benchmarks/... -run TestParseBenchmark
go test ./internal/test/benchmarks/... -run TestCompareWithBaseline
go test ./internal/test/benchmarks/... -run TestGenerateReport
```

### Test Categories

- **Unit Tests**: Individual function testing with 100% coverage
- **Integration Tests**: Cross-component interaction validation
- **Error Handling Tests**: Comprehensive error scenario coverage
- **Edge Case Tests**: Boundary condition and malformed input handling
- **Performance Tests**: Memory leak prevention and performance validation
- **Concurrency Tests**: Thread-safety validation

### Test Coverage Breakdown

```
NewPerformanceMonitor           100.0%
SetThresholds                   100.0%
ParseBenchmarkOutput            100.0%
parseBenchmarkLine              100.0%
CompareWithBaseline             100.0%
calculateSeverity               100.0%
generateRecommendation          100.0%
countCriticalRegressions        100.0%
determineOverallTrend           100.0%
createInitialReport             100.0%
SaveBaseline                    100.0%
loadBaseline                    100.0%
GenerateTextReport              100.0%
GenerateJSONReport              100.0%
total:                          100.0%
```

## 📊 Performance

The package is optimized for high-performance benchmark processing:

- **Fast Parsing**: Efficient regex-free benchmark output parsing
- **Memory Efficient**: Minimal memory allocation during processing
- **Concurrent Safe**: Thread-safe operations for parallel processing
- **Scalable**: Handles large benchmark datasets efficiently

### Benchmarks

```bash
# Run performance benchmarks
go test ./internal/test/benchmarks/... -bench=.

# Example results:
BenchmarkParseOutput-8           1000000    1.2μs/op    64B/op
BenchmarkCompareBaseline-8        500000    2.1μs/op    96B/op
BenchmarkGenerateReport-8         200000    5.5μs/op   256B/op
```

### Memory Efficiency

- **Parsing**: ~64 bytes per benchmark result
- **Comparison**: ~96 bytes per baseline comparison
- **Report Generation**: ~256 bytes per report
- **Memory Leak Prevention**: Comprehensive leak detection tests

## 🔍 Error Handling

The package provides robust error handling with specific error types:

### Error Categories

- **Parsing Errors**: Malformed benchmark output handling
- **File System Errors**: Baseline file access and permission issues
- **Validation Errors**: Invalid configuration and threshold validation
- **Processing Errors**: Comparison and calculation error handling

### Error Handling Patterns

```go
// Graceful error handling example
results, err := monitor.ParseBenchmarkOutput(output)
if err != nil {
    // Log error but continue processing
    log.Printf("Warning: Failed to parse some benchmark lines: %v", err)
    // Partial results may still be available
}

// Baseline file handling
report, err := monitor.CompareWithBaseline(results)
if err != nil {
    // First run - create initial baseline
    if os.IsNotExist(err) {
        return monitor.createInitialReport(results), nil
    }
    return nil, fmt.Errorf("baseline comparison failed: %w", err)
}
```

## 🎯 Best Practices

### Configuration Recommendations

1. **Threshold Setting**: Start with default thresholds (20% slowdown, 25% memory) and adjust based on project needs
2. **Baseline Management**: Update baselines regularly but preserve historical data
3. **CI Integration**: Use stricter thresholds in CI/CD pipelines
4. **Report Format**: Use JSON for automated processing, text for human review

### Usage Patterns

1. **Development Workflow**: Regular baseline updates during feature development
2. **Release Validation**: Strict regression checking before releases
3. **Performance Monitoring**: Continuous tracking of performance trends
4. **Optimization Guidance**: Use recommendations for targeted improvements

### Performance Optimization

1. **Batch Processing**: Process multiple benchmark files together
2. **Threshold Tuning**: Adjust thresholds based on application characteristics
3. **Baseline Strategy**: Use rolling baselines for evolving codebases
4. **Report Filtering**: Focus on critical and major regressions first

## 🤝 Contributing

### Development Setup

```bash
# Clone repository
git clone https://github.com/newbpydev/go-sentinel.git
cd go-sentinel/internal/test/benchmarks

# Run tests
go test -v -cover

# Run linting
golangci-lint run

# Run benchmarks
go test -bench=. -benchmem
```

### Quality Standards

- **Test Coverage**: Maintain 100% test coverage
- **Code Quality**: Follow Go best practices and project linting rules
- **Documentation**: Update README for any API changes
- **Performance**: Ensure no performance regressions in benchmarks

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](../../../LICENSE) file for details.

## 🔗 Related Packages

- [`internal/test/processor`](../processor/README.md) - Test result processing and formatting
- [`internal/test/runner`](../runner/README.md) - Test execution and management
- [`internal/test/cache`](../cache/README.md) - Test result caching and optimization
- [`internal/test/metrics`](../metrics/README.md) - Test metrics and complexity analysis
- [`pkg/models`](../../../pkg/models/README.md) - Shared data models and interfaces

---

**Note**: This package achieves 100% test coverage through systematic TDD implementation, ensuring reliability and maintainability for production use in performance monitoring workflows. 