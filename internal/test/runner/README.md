# 📦 Test Runner Package

[![Test Coverage](https://img.shields.io/badge/coverage-80.0%25-green.svg)](https://github.com/newbpydev/go-sentinel/tree/main/internal/test/runner)
[![Go Report Card](https://goreportcard.com/badge/github.com/newbpydev/go-sentinel/internal/test/runner)](https://goreportcard.com/report/github.com/newbpydev/go-sentinel/internal/test/runner)
[![Go Reference](https://pkg.go.dev/badge/github.com/newbpydev/go-sentinel/internal/test/runner.svg)](https://pkg.go.dev/github.com/newbpydev/go-sentinel/internal/test/runner)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## 📖 Overview

The Test Runner package provides a comprehensive suite of test execution engines for the Go Sentinel CLI. It implements multiple execution strategies including basic sequential execution, optimized smart execution with caching, and parallel execution with concurrency control.

### 🎯 Key Features

- **Basic Test Runner**: Simple, reliable test execution with streaming output support
- **Optimized Test Runner**: Smart execution with dependency analysis, caching, and selective test running
- **Parallel Test Runner**: Concurrent test execution with configurable worker pools and progress monitoring
- **Performance Optimizer**: Advanced rendering and processing optimizations for large test suites
- **Safety Mechanisms**: Infinite loop prevention, context cancellation, and resource cleanup
- **Streaming Support**: Real-time test output processing with JSON parsing capabilities

## 🏗️ Architecture

This package follows clean architecture principles with clear separation of concerns:

- **Single Responsibility**: Each runner type focuses on a specific execution strategy
- **Dependency Inversion**: All runners implement common interfaces for interchangeability
- **Interface Segregation**: Small, focused interfaces for specific concerns (TestRunnerInterface, TestExecutor)
- **Factory Pattern**: Centralized creation of runner instances with proper configuration
- **Observer Pattern**: Progress monitoring and event handling for test execution
- **Strategy Pattern**: Multiple execution strategies (basic, optimized, parallel) with runtime selection

### 📦 Package Structure

```
internal/test/runner/
├── basic_runner.go           # Basic sequential test execution
├── basic_runner_test.go      # Comprehensive basic runner tests
├── executor.go               # Core test execution engine with process management
├── executor_test.go          # Extensive executor tests with safety measures
├── optimized_runner.go       # Smart execution with caching and optimization
├── optimized_runner_test.go  # Optimized runner functionality tests
├── parallel_runner.go        # Concurrent execution with worker pools
├── parallel_runner_test.go   # Parallel execution and safety tests
├── performance_optimizer.go  # Advanced rendering and processing optimizations
├── interfaces.go             # Core interfaces and contracts
├── additional_coverage_test.go # Additional tests for edge cases and coverage
└── README.md                # This comprehensive documentation
```

## 🚀 Quick Start

### Basic Test Execution

```go
package main

import (
    "context"
    "log"
    "github.com/newbpydev/go-sentinel/internal/test/runner"
)

func main() {
    // Create a basic test runner
    testRunner := runner.NewBasicTestRunner(true, true) // verbose=true, jsonOutput=true
    
    // Execute tests
    ctx := context.Background()
    output, err := testRunner.Run(ctx, []string{"./pkg/..."})
    if err != nil {
        log.Fatal("Test execution failed:", err)
    }
    
    log.Println("Test output:", output)
}
```

### Streaming Test Execution

```go
package main

import (
    "context"
    "io"
    "log"
    "github.com/newbpydev/go-sentinel/internal/test/runner"
)

func main() {
    testRunner := runner.NewBasicTestRunner(false, true)
    
    // Get streaming output
    stream, err := testRunner.RunStream(context.Background(), []string{"./..."})
    if err != nil {
        log.Fatal("Failed to start test stream:", err)
    }
    defer stream.Close()
    
    // Process stream in real-time
    buffer := make([]byte, 1024)
    for {
        n, err := stream.Read(buffer)
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Printf("Stream error: %v", err)
            break
        }
        
        // Process test output
        log.Printf("Test output: %s", string(buffer[:n]))
    }
}
```

### Parallel Test Execution

```go
package main

import (
    "context"
    "log"
    "github.com/newbpydev/go-sentinel/internal/config"
    "github.com/newbpydev/go-sentinel/internal/test/runner"
    "github.com/newbpydev/go-sentinel/internal/test/cache"
)

func main() {
    // Create components
    basicRunner := runner.NewBasicTestRunner(true, true)
    testCache := cache.NewTestResultCache()
    
    // Create parallel runner with 4 workers
    parallelRunner := runner.NewParallelTestRunner(4, basicRunner, testCache)
    
    // Configure execution
    cfg := &config.Config{
        Verbosity: 1,
        Timeout:   30 * time.Second,
    }
    
    // Execute tests in parallel
    testPaths := []string{"./pkg/models", "./pkg/events", "./internal/app"}
    results, err := parallelRunner.RunParallel(context.Background(), testPaths, cfg)
    if err != nil {
        log.Fatal("Parallel execution failed:", err)
    }
    
    // Process results
    for _, result := range results {
        if result.Error != nil {
            log.Printf("Test path %s failed: %v", result.TestPath, result.Error)
        } else {
            log.Printf("Test path %s completed in %v (from cache: %v)", 
                result.TestPath, result.Duration, result.FromCache)
        }
    }
}
```

## 🔧 Core Interfaces

### TestRunnerInterface

The main interface for test execution:

```go
type TestRunnerInterface interface {
    Run(ctx context.Context, testPaths []string) (string, error)
    RunStream(ctx context.Context, testPaths []string) (io.ReadCloser, error)
}
```

### TestExecutor

Advanced test execution with process control:

```go
type TestExecutor interface {
    Execute(ctx context.Context, packages []string, options *ExecutionOptions) (*ExecutionResult, error)
    Cancel() error
    IsRunning() bool
}
```

### FileChangeInterface

For optimized test execution based on file changes:

```go
type FileChangeInterface interface {
    GetPath() string
    GetType() ChangeType
    IsNewChange() bool
}
```

## 🔄 Advanced Usage

### Optimized Test Execution

```go
package main

import (
    "context"
    "log"
    "time"
    "github.com/newbpydev/go-sentinel/internal/test/runner"
    "github.com/newbpydev/go-sentinel/pkg/models"
)

func main() {
    // Create optimized runner
    optimizedRunner := runner.NewOptimizedTestRunner()
    
    // Configure optimization settings
    optimizedRunner.SetCacheEnabled(true)
    optimizedRunner.SetOnlyRunChangedTests(true)
    optimizedRunner.SetOptimizationMode("aggressive")
    
    // Create file changes
    changes := []runner.FileChangeInterface{
        &runner.FileChangeAdapter{
            FileChange: &models.FileChange{
                FilePath:   "./pkg/models/user.go",
                ChangeType: models.ChangeTypeModified,
                Timestamp:  time.Now(),
            },
        },
    }
    
    // Run optimized tests
    result, err := optimizedRunner.RunOptimized(context.Background(), changes)
    if err != nil {
        log.Fatal("Optimized execution failed:", err)
    }
    
    // Get efficiency statistics
    stats := result.GetEfficiencyStats()
    log.Printf("Tests run: %v, Cache hits: %v", stats["tests_run"], stats["cache_hits"])
}
```

### Performance Optimization

```go
package main

import (
    "log"
    "github.com/newbpydev/go-sentinel/internal/test/runner"
    "github.com/newbpydev/go-sentinel/internal/test/processor"
    "github.com/newbpydev/go-sentinel/pkg/models"
)

func main() {
    // Create performance optimizer
    testProcessor := processor.NewTestProcessor(os.Stdout, colorFormatter, iconProvider, 120)
    optimizer := runner.NewOptimizedTestProcessor(os.Stdout, testProcessor)
    
    // Add test suites
    for i := 0; i < 1000; i++ {
        suite := &models.TestSuite{
            FilePath:     fmt.Sprintf("test_%d.go", i),
            TestCount:    10,
            PassedCount:  8,
            FailedCount:  2,
        }
        optimizer.AddTestSuite(suite)
    }
    
    // Render with optimizations
    err := optimizer.RenderResultsOptimized(false)
    if err != nil {
        log.Fatal("Optimized rendering failed:", err)
    }
    
    // Get memory statistics
    memStats := optimizer.GetMemoryStats()
    log.Printf("Memory usage: %+v", memStats)
}
```

## 🧪 Testing

The package achieves **80.0% test coverage** with comprehensive test suites covering:

### Test Categories

- **Unit Tests**: Individual component testing with mocks and stubs
- **Integration Tests**: Cross-component interaction testing
- **Concurrency Tests**: Thread-safety validation with 100+ goroutines
- **Safety Tests**: Infinite loop prevention and resource cleanup
- **Performance Tests**: Memory efficiency and execution speed validation
- **Error Handling Tests**: Comprehensive error scenario coverage

### Running Tests

```bash
# Run all tests
go test ./internal/test/runner/...

# Run with coverage
go test ./internal/test/runner/... -coverprofile=coverage.out

# View coverage report
go tool cover -html=coverage.out

# Run specific test categories
go test ./internal/test/runner -run TestBasicTestRunner
go test ./internal/test/runner -run TestParallelRunner
go test ./internal/test/runner -run TestOptimizedTestRunner
```

### Test Safety Features

The test suite includes critical safety mechanisms:

- **Infinite Loop Prevention**: Timeout protections and iteration limits
- **Context Cancellation**: Proper cleanup on cancellation
- **Resource Management**: Goroutine leak detection and cleanup
- **Process Control**: Safe process termination and signal handling

## 📊 Performance

The package is optimized for high-performance test execution:

- **Fast Startup**: Minimal overhead for test initialization
- **Efficient Parallel Execution**: Optimal worker pool utilization
- **Memory Efficient**: Streaming processing with bounded memory usage
- **Smart Caching**: Dependency-aware test result caching
- **Concurrent Safe**: Thread-safe operations with minimal locking

### Benchmarks

```bash
# Run performance benchmarks
go test ./internal/test/runner/... -bench=.

# Example results:
BenchmarkBasicTestRunner_Run-8              100    12.5ms/op    1.2MB/op
BenchmarkParallelTestRunner_RunParallel-8    50    25.1ms/op    2.4MB/op
BenchmarkOptimizedTestRunner_RunOptimized-8  75    18.3ms/op    1.8MB/op
```

### Memory Optimization

- **Streaming Processing**: Processes test output without loading entire results into memory
- **Bounded Channels**: Prevents memory buildup in concurrent operations
- **Garbage Collection**: Proactive cleanup of temporary resources
- **Pool Reuse**: Object pooling for frequently allocated structures

## 🔍 Error Handling

The package implements comprehensive error handling:

### Error Types

- **ExecutionError**: Test execution failures with context
- **TimeoutError**: Context deadline exceeded during execution
- **CancellationError**: User-initiated cancellation
- **ValidationError**: Invalid input parameters or configuration
- **ResourceError**: System resource limitations or failures

### Error Handling Patterns

```go
// Graceful error handling with context
result, err := testRunner.Run(ctx, testPaths)
if err != nil {
    switch {
    case errors.Is(err, context.DeadlineExceeded):
        log.Printf("Test execution timed out: %v", err)
    case errors.Is(err, context.Canceled):
        log.Printf("Test execution was cancelled: %v", err)
    default:
        log.Printf("Test execution failed: %v", err)
    }
    return
}
```

### Recovery Mechanisms

- **Automatic Retry**: Configurable retry logic for transient failures
- **Graceful Degradation**: Fallback to basic execution on optimization failures
- **Resource Recovery**: Automatic cleanup of leaked resources
- **State Recovery**: Restoration of consistent state after failures

## 🎯 Best Practices

### Usage Recommendations

1. **Choose the Right Runner**:
   - Use `BasicTestRunner` for simple, reliable execution
   - Use `OptimizedTestRunner` for large codebases with frequent changes
   - Use `ParallelTestRunner` for independent test packages

2. **Configure Appropriately**:
   - Set reasonable timeouts based on test complexity
   - Configure concurrency based on system resources
   - Enable caching for repeated test runs

3. **Handle Errors Gracefully**:
   - Always check error returns
   - Implement proper cancellation handling
   - Log errors with sufficient context

4. **Monitor Performance**:
   - Use benchmarks to validate performance improvements
   - Monitor memory usage in long-running processes
   - Profile critical execution paths

### Integration Patterns

```go
// Factory pattern for runner selection
func CreateTestRunner(strategy string, config *Config) TestRunnerInterface {
    switch strategy {
    case "basic":
        return runner.NewBasicTestRunner(config.Verbose, config.JSONOutput)
    case "optimized":
        optimized := runner.NewOptimizedTestRunner()
        optimized.SetOptimizationMode(config.OptimizationMode)
        return optimized
    case "parallel":
        basic := runner.NewBasicTestRunner(config.Verbose, config.JSONOutput)
        cache := cache.NewTestResultCache()
        return runner.NewParallelTestRunner(config.Concurrency, basic, cache)
    default:
        return runner.NewBasicTestRunner(false, true)
    }
}
```

## 🤝 Contributing

### Development Setup

1. Clone the repository
2. Install Go 1.21 or later
3. Run `go mod download` to install dependencies
4. Run `go test ./internal/test/runner/...` to verify setup

### Quality Standards

- Maintain **≥80%** test coverage for all new code
- Follow Go formatting standards (`go fmt`)
- Pass all linter checks (`golangci-lint run`)
- Include comprehensive error handling
- Add benchmarks for performance-critical code

### Testing Guidelines

- Write tests before implementation (TDD)
- Include both positive and negative test cases
- Test concurrent access patterns
- Verify resource cleanup and leak prevention
- Document test scenarios and expected behaviors

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](../../../LICENSE) file for details.

## 🔗 Related Packages

- [`internal/test/processor`](../processor/README.md) - Test result processing and formatting
- [`internal/test/cache`](../cache/README.md) - Test result caching and optimization
- [`internal/config`](../../config/README.md) - Configuration management
- [`pkg/models`](../../../pkg/models/README.md) - Core data models and interfaces
- [`pkg/events`](../../../pkg/events/README.md) - Event system for test notifications

---

**Note**: This package is part of the Go Sentinel CLI project and follows the established architecture principles for maintainability, testability, and performance. 