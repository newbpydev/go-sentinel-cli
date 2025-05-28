# 📦 cache

[![Test Coverage](https://img.shields.io/badge/coverage-100.0%25-brightgreen.svg)](https://github.com/newbpydev/go-sentinel/tree/main/internal/test/cache)
[![Go Report Card](https://goreportcard.com/badge/github.com/newbpydev/go-sentinel/internal/test/cache)](https://goreportcard.com/report/github.com/newbpydev/go-sentinel/internal/test/cache)
[![Go Reference](https://pkg.go.dev/badge/github.com/newbpydev/go-sentinel/internal/test/cache.svg)](https://pkg.go.dev/github.com/newbpydev/go-sentinel/internal/test/cache)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## 📖 Overview
The `cache` package provides robust, thread-safe caching for test results, file hashes, and dependency relationships to optimize incremental test execution in the Go Sentinel CLI. It supports fine-grained invalidation, dependency tracking, and pluggable storage backends.

### 🎯 Key Features
- **Test Result Caching**: Store and retrieve test results for fast incremental runs.
- **File Hash Tracking**: Detect file changes efficiently using hash caching.
- **Dependency Management**: Track and invalidate tests based on file and config changes.
- **Thread-Safe**: All cache operations are safe for concurrent use.
- **Extensible Interfaces**: Pluggable storage and cache strategies.

## 🏗️ Architecture
- **Single Responsibility**: Each cache type (test results, file hashes, dependencies) is encapsulated in its own interface.
- **Interface Segregation**: Interfaces are small and focused for easy mocking and extension.
- **Dependency Inversion**: Consumers depend on interfaces, not concrete implementations.
- **Thread Safety**: Uses `sync.RWMutex` for safe concurrent access.

### 📦 Package Structure
```
internal/test/cache/
├── result_cache.go        # Main implementation of test result cache
├── result_cache_test.go  # Comprehensive test suite (100% coverage)
├── interfaces.go         # Core cache interfaces and types
└── README.md             # This documentation
```

## 🚀 Quick Start
```go
package main

import (
    "fmt"
    "github.com/newbpydev/go-sentinel/internal/test/cache"
    "github.com/newbpydev/go-sentinel/pkg/models"
)

func main() {
    c := cache.NewTestResultCache()
    suite := &models.TestSuite{FilePath: "example_test.go", TestCount: 1, PassedCount: 1}
    c.CacheResult("example", suite)
    result, exists := c.GetCachedResult("example")
    if exists {
        fmt.Println("Cached result:", result.Suite.FilePath)
    }
}
```

## 🔧 Core Interfaces
```go
type CacheInterface interface {
    AnalyzeChange(filePath string) (*FileChange, error)
    MarkFileAsProcessed(filePath string, processTime time.Time)
    ShouldRunTests(changes []*FileChange) (bool, []string)
    GetStaleTests(changes []*FileChange) []string
    CacheResult(testPath string, suite *models.TestSuite)
    GetCachedResult(testPath string) (*CachedTestResult, bool)
    Clear()
    GetStats() map[string]interface{}
}
```

## 🔄 Advanced Usage
- **Dependency Invalidation**: Cached results are automatically invalidated if any dependency file changes.
- **Custom Storage**: Implement the `Storage` interface to use a custom backend (e.g., Redis, disk).
- **Fine-Grained Invalidation**: Use `AnalyzeChange` and `GetStaleTests` to determine exactly which tests to rerun after a change.

## 🧪 Testing
- **100% coverage**: All logic, edge cases, and concurrency scenarios are tested.
- **How to run:**
  ```bash
  go test ./internal/test/cache -cover
  go tool cover -html=coverage.out
  ```
- **Test categories:**
  - Unit tests for all exported and unexported functions
  - Table-driven edge case tests
  - Concurrency and race condition tests

## 📊 Performance
- **Efficient lookups**: Uses maps and minimal locking for fast access.
- **Low memory overhead**: Only stores necessary metadata and results.
- **Benchmarks:**
  ```go
  func BenchmarkCacheResult(b *testing.B) {
      c := cache.NewTestResultCache()
      suite := &models.TestSuite{FilePath: "bench.go"}
      for i := 0; i < b.N; i++ {
          c.CacheResult("bench", suite)
      }
  }
  ```

## 🔍 Error Handling
- All errors include context (e.g., file path, operation).
- Graceful handling of missing files, permission errors, and dependency changes.
- No panics on invalid input; always returns error or false.

## 🎯 Best Practices
- Always use the provided interfaces for cache operations.
- Use `Clear()` before large test runs to avoid stale data.
- Use `GetStats()` to monitor cache health and usage.

## 🤝 Contributing
- Follow Go formatting and linting standards.
- Add table-driven tests for all new logic.
- Document all exported symbols and interfaces.

## 📄 License
MIT License. See [LICENSE](../../LICENSE).

## 🔗 Related Packages
- [`internal/test/processor`](../processor): Test result processing and reporting
- [`internal/test/runner`](../runner): Test execution engine
- [`pkg/models`](../../../pkg/models): Shared data models 