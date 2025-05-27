# 📦 File System Watcher Package

[![Test Coverage](https://img.shields.io/badge/coverage-91.0%25-brightgreen.svg)](https://github.com/newbpydev/go-sentinel/tree/main/internal/watch/watcher)
[![Go Report Card](https://goreportcard.com/badge/github.com/newbpydev/go-sentinel/internal/watch/watcher)](https://goreportcard.com/report/github.com/newbpydev/go-sentinel/internal/watch/watcher)
[![Go Reference](https://pkg.go.dev/badge/github.com/newbpydev/go-sentinel/internal/watch/watcher.svg)](https://pkg.go.dev/github.com/newbpydev/go-sentinel/internal/watch/watcher)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## 📖 Overview

The `internal/watch/watcher` package provides comprehensive file system monitoring, pattern matching, and test file discovery capabilities for the Go Sentinel CLI. It implements real-time file watching with intelligent filtering, cross-platform path handling, and robust test file detection.

### 🎯 Key Features

- **Real-time File Watching**: Monitor file system changes with fsnotify-based implementation supporting create, write, remove, rename, and chmod events
- **Advanced Pattern Matching**: Flexible pattern matching with wildcard support (`*.go`), directory patterns (`src/`), and recursive patterns (`src/**`)
- **Test File Discovery**: Intelligent test file detection and related implementation file discovery for Go projects
- **Cross-platform Compatibility**: Seamless handling of Windows and Unix path separators with automatic normalization
- **Concurrent Safety**: Thread-safe operations with proper resource cleanup and graceful shutdown handling
- **Performance Optimized**: Efficient event debouncing and ignore pattern filtering to reduce noise

## 🏗️ Architecture

This package follows clean architecture principles and implements several design patterns:

- **Factory Pattern**: Clean object creation with `NewFileSystemWatcher`, `NewPatternMatcher`, and `NewTestFileFinder`
- **Interface Segregation**: Small, focused interfaces (`FileSystemWatcher`, `PatternMatcher`) defined in consuming packages
- **Strategy Pattern**: Pluggable pattern matching strategies with different matching algorithms
- **Observer Pattern**: Event-driven file monitoring with channel-based communication
- **Single Responsibility**: Each component has one clear purpose (watching, pattern matching, test discovery)

### 📦 Package Structure

```
internal/watch/watcher/
├── fs_watcher.go              # Main file system watcher implementation (341 lines)
├── patterns.go                # Pattern matching engine (131 lines)
├── fs_watcher_test.go         # Comprehensive watcher tests (677+ lines)
├── pattern_matcher_test.go    # Pattern matching tests (677+ lines)
├── test_file_finder_test.go   # Test discovery tests (200+ lines)
└── README.md                  # This documentation
```

**Core Components:**
- **FileSystemWatcher**: Real-time file system monitoring with event filtering
- **PatternMatcher**: Advanced pattern matching with wildcard and recursive support
- **TestFileFinder**: Go-specific test file discovery and relationship mapping

## 🚀 Quick Start

### Basic File Watching

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/newbpydev/go-sentinel/internal/watch/watcher"
    "github.com/newbpydev/go-sentinel/internal/watch/core"
)

func main() {
    // Create file system watcher with paths and ignore patterns
    watcher, err := watcher.NewFileSystemWatcher(
        []string{"./src", "./tests"},           // Watch these paths
        []string{".git", "node_modules", "*.tmp"}, // Ignore these patterns
    )
    if err != nil {
        log.Fatal("Failed to create watcher:", err)
    }
    defer watcher.Close()

    // Set up context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Create event channel
    events := make(chan core.FileEvent, 100)

    // Start watching in background
    go func() {
        if err := watcher.Watch(ctx, events); err != nil {
            log.Printf("Watch error: %v", err)
        }
    }()

    // Process events
    for {
        select {
        case event := <-events:
            fmt.Printf("File %s: %s (test: %v)\n", 
                event.Type, event.Path, event.IsTest)
        case <-ctx.Done():
            fmt.Println("Watching stopped")
            return
        }
    }
}
```

### Pattern Matching

```go
package main

import (
    "fmt"
    "github.com/newbpydev/go-sentinel/internal/watch/watcher"
)

func main() {
    matcher := watcher.NewPatternMatcher()

    // Test various patterns
    patterns := []string{
        "*.go",        // All Go files
        "*_test.go",   // Test files
        "vendor/**",   // Vendor directory (recursive)
        ".git",        // Git directory
    }

    testPaths := []string{
        "main.go",
        "main_test.go", 
        "vendor/pkg/lib.go",
        ".git/config",
        "src/utils.py",
    }

    for _, path := range testPaths {
        if matcher.MatchesAny(path, patterns) {
            fmt.Printf("✓ %s matches ignore patterns\n", path)
        } else {
            fmt.Printf("○ %s should be watched\n", path)
        }
    }
}
```

### Test File Discovery

```go
package main

import (
    "fmt"
    "log"
    "github.com/newbpydev/go-sentinel/internal/watch/watcher"
)

func main() {
    finder := watcher.NewTestFileFinder("./myproject")

    // Find test file for implementation
    testFile, err := finder.FindTestFile("src/calculator.go")
    if err != nil {
        log.Printf("No test file found: %v", err)
    } else {
        fmt.Printf("Test file: %s\n", testFile)
    }

    // Find implementation for test
    implFile, err := finder.FindImplementationFile("src/calculator_test.go")
    if err != nil {
        log.Printf("No implementation found: %v", err)
    } else {
        fmt.Printf("Implementation: %s\n", implFile)
    }

    // Find all tests in package
    tests, err := finder.FindPackageTests("src/utils.go")
    if err != nil {
        log.Printf("Error finding package tests: %v", err)
    } else {
        fmt.Printf("Package tests: %v\n", tests)
    }
}
```

## 🔧 Core Interfaces

### FileSystemWatcher

The main file watching interface with real-time monitoring capabilities:

```go
type FileSystemWatcher interface {
    // Watch starts monitoring for file changes
    Watch(ctx context.Context, events chan<- core.FileEvent) error
    
    // AddPath adds a new path to monitor
    AddPath(path string) error
    
    // RemovePath removes a path from monitoring
    RemovePath(path string) error
    
    // Close releases all resources
    Close() error
}
```

**Usage Example:**
```go
watcher, err := watcher.NewFileSystemWatcher([]string{"."}, []string{".git"})
if err != nil {
    return err
}
defer watcher.Close()

events := make(chan core.FileEvent, 10)
ctx := context.Background()

go watcher.Watch(ctx, events)
```

### PatternMatcher

Advanced pattern matching with wildcard and recursive support:

```go
type PatternMatcher interface {
    // MatchesAny checks if path matches any pattern
    MatchesAny(path string, patterns []string) bool
    
    // MatchesPattern checks if path matches specific pattern
    MatchesPattern(path string, pattern string) bool
    
    // AddPattern adds a new pattern to the matcher
    AddPattern(pattern string) error
    
    // RemovePattern removes a pattern from the matcher
    RemovePattern(pattern string) error
}
```

**Supported Pattern Types:**
- **Exact Match**: `main.go`, `src/utils.go`
- **Wildcard**: `*.go`, `*_test.go`, `test_*`
- **Directory**: `src/`, `.git`, `node_modules`
- **Recursive**: `vendor/**`, `src/**/test/**`
- **Cross-platform**: Handles both `/` and `\` separators

### TestFileFinder

Go-specific test file discovery and relationship mapping:

```go
type TestFileFinder interface {
    // FindTestFile finds the test file for an implementation file
    FindTestFile(filePath string) (string, error)
    
    // FindImplementationFile finds implementation for a test file
    FindImplementationFile(testPath string) (string, error)
    
    // FindPackageTests finds all test files in the same package
    FindPackageTests(filePath string) ([]string, error)
    
    // IsTestFile checks if a file is a test file
    IsTestFile(filePath string) bool
}
```

## 🔄 Advanced Usage

### Dynamic Path Management

```go
watcher, err := watcher.NewFileSystemWatcher([]string{}, []string{".git"})
if err != nil {
    return err
}

// Add paths dynamically
err = watcher.AddPath("./src")
if err != nil {
    log.Printf("Failed to add path: %v", err)
}

// Remove paths when no longer needed
err = watcher.RemovePath("./old-src")
if err != nil {
    log.Printf("Failed to remove path: %v", err)
}
```

### Custom Pattern Matching

```go
matcher := watcher.NewPatternMatcher()

// Add custom patterns
matcher.AddPattern("*.generated.go")  // Generated files
matcher.AddPattern("vendor/**")       // Vendor directory
matcher.AddPattern("**/node_modules") // Node modules anywhere

// Test against custom patterns
isIgnored := matcher.MatchesPattern("src/proto.generated.go", "*.generated.go")
```

### Error Handling and Timeouts

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

events := make(chan core.FileEvent, 100)

go func() {
    if err := watcher.Watch(ctx, events); err != nil {
        if err == context.Canceled {
            log.Println("Watching cancelled by user")
        } else {
            log.Printf("Watch error: %v", err)
        }
    }
}()

// Handle events with timeout
for {
    select {
    case event := <-events:
        // Process file event
        processFileEvent(event)
    case <-ctx.Done():
        log.Println("Watching stopped due to timeout")
        return
    }
}
```

### Concurrent Usage

```go
var wg sync.WaitGroup
numWatchers := 3

for i := 0; i < numWatchers; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        
        w, err := watcher.NewFileSystemWatcher(
            []string{fmt.Sprintf("./dir%d", id)}, 
            []string{".git"},
        )
        if err != nil {
            return
        }
        defer w.Close()
        
        // Each watcher monitors different directory
        events := make(chan core.FileEvent, 10)
        ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
        defer cancel()
        
        w.Watch(ctx, events)
    }(i)
}

wg.Wait()
```

## 🧪 Testing

The package has **91.0% test coverage** with comprehensive test suites covering all functionality.

### Running Tests

```bash
# Run all tests
go test ./internal/watch/watcher/...

# Run with coverage
go test ./internal/watch/watcher/... -coverprofile=coverage.out

# View coverage report
go tool cover -html=coverage.out

# Run specific test suites
go test -run TestFileSystemWatcher ./internal/watch/watcher
go test -run TestPatternMatcher ./internal/watch/watcher
go test -run TestFileFinder ./internal/watch/watcher
```

### Test Categories

- **Unit Tests**: Individual component testing with factory functions, interface compliance, and edge cases
- **Integration Tests**: Cross-component interaction testing with real file system operations
- **Concurrency Tests**: Thread-safety validation with 100+ concurrent goroutines and race condition detection
- **Platform Tests**: Cross-platform path handling and Windows/Unix compatibility
- **Performance Tests**: Memory efficiency testing and goroutine leak detection
- **Error Handling Tests**: Comprehensive error scenario coverage including context cancellation and resource cleanup

### Coverage Breakdown

```bash
github.com/newbpydev/go-sentinel/internal/watch/watcher/fs_watcher.go:
NewFileSystemWatcher    75.0%   # Factory function
Watch                   73.9%   # Core watching logic
AddPath                 81.5%   # Path management
RemovePath              91.7%   # Path removal
Close                   100.0%  # Resource cleanup
matchesAnyPattern       100.0%  # Pattern filtering
eventTypeString         100.0%  # Event type conversion

github.com/newbpydev/go-sentinel/internal/watch/watcher/patterns.go:
NewPatternMatcher       100.0%  # Factory function
MatchesAny              100.0%  # Multi-pattern matching
MatchesPattern          93.5%   # Core pattern logic
AddPattern              100.0%  # Pattern management
RemovePattern           100.0%  # Pattern removal

Total Coverage:         91.0%   # Excellent coverage
```

## 📊 Performance

The package is optimized for performance with efficient algorithms and minimal memory allocation:

- **Fast Startup**: Minimal overhead for watcher initialization (<1ms)
- **Efficient Pattern Matching**: Optimized wildcard and recursive pattern algorithms
- **Memory Efficient**: Minimal memory allocation with reusable data structures
- **Concurrent Safe**: Thread-safe operations with minimal locking overhead
- **Resource Management**: Proper cleanup prevents memory leaks and goroutine leaks

### Benchmarks

```bash
# Run performance benchmarks
go test ./internal/watch/watcher/... -bench=. -benchmem

# Example results (platform-dependent):
BenchmarkFileSystemWatcher_AddPath-8           1000000    1.2μs/op    64B/op
BenchmarkPatternMatcher_MatchesPattern-8       5000000    0.3μs/op    16B/op
BenchmarkTestFileFinder_FindTestFile-8          500000    2.1μs/op    128B/op
BenchmarkPatternMatcher_MatchesAny-8           2000000    0.8μs/op    32B/op
```

### Memory Efficiency

- **Low Allocation**: Minimal heap allocations during normal operation
- **Reusable Buffers**: Internal buffer reuse for path processing
- **Efficient Data Structures**: Optimized slice and map usage for pattern storage
- **Garbage Collection Friendly**: Minimal GC pressure with proper object lifecycle management

## 🔍 Error Handling

The package provides comprehensive error handling with context-rich error messages:

### Error Types

- **Path Errors**: Invalid paths, permission issues, non-existent directories
- **Pattern Errors**: Malformed wildcard patterns, invalid recursive patterns
- **Context Errors**: Cancellation, timeouts, deadline exceeded
- **Resource Errors**: File descriptor limits, watcher creation failures

### Error Handling Patterns

```go
// Path validation errors
watcher, err := watcher.NewFileSystemWatcher([]string{"./invalid"}, nil)
if err != nil {
    if strings.Contains(err.Error(), "failed to stat path") {
        log.Printf("Invalid path provided: %v", err)
        return
    }
}

// Context cancellation handling
ctx, cancel := context.WithCancel(context.Background())
go func() {
    err := watcher.Watch(ctx, events)
    if err == context.Canceled {
        log.Println("Watching cancelled gracefully")
    } else if err != nil {
        log.Printf("Unexpected watch error: %v", err)
    }
}()

// Pattern matching error handling
matcher := watcher.NewPatternMatcher()
if !matcher.MatchesPattern("file.go", "[invalid") {
    // Invalid patterns are handled gracefully (return false)
    log.Println("Pattern matching failed safely")
}
```

### Best Error Practices

- **Validate Early**: All inputs are validated at function entry points
- **Context Awareness**: Proper context handling for cancellation and timeouts
- **Resource Cleanup**: Automatic cleanup in error paths prevents resource leaks
- **Meaningful Messages**: Error messages include operation context and specific failure details

## 🎯 Best Practices

### Watcher Usage

```go
// ✅ DO: Use context for timeout control
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

// ✅ DO: Buffer event channels appropriately
events := make(chan core.FileEvent, 100) // Sufficient buffer

// ✅ DO: Always close watchers
defer watcher.Close()

// ❌ DON'T: Create watchers without cleanup
watcher, _ := watcher.NewFileSystemWatcher(paths, patterns)
// Missing: defer watcher.Close()
```

### Pattern Management

```go
// ✅ DO: Use specific patterns
patterns := []string{"*.go", "*_test.go", "vendor/**"}

// ✅ DO: Combine related patterns
ignorePatterns := []string{".git", "node_modules", "*.tmp", "*.log"}

// ❌ DON'T: Use overly broad patterns
badPatterns := []string{"*"} // Too broad, matches everything
```

### Performance Optimization

```go
// ✅ DO: Reuse pattern matchers
matcher := watcher.NewPatternMatcher()
for _, pattern := range patterns {
    matcher.AddPattern(pattern)
}

// ✅ DO: Use appropriate buffer sizes
events := make(chan core.FileEvent, 100) // Good buffer size

// ❌ DON'T: Create new matchers repeatedly
for _, file := range files {
    // Bad: creates new matcher each time
    newMatcher := watcher.NewPatternMatcher()
    newMatcher.MatchesPattern(file, pattern)
}
```

## 🤝 Contributing

### Development Setup

1. **Clone the repository**:
   ```bash
   git clone https://github.com/newbpydev/go-sentinel.git
   cd go-sentinel/internal/watch/watcher
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Run tests**:
   ```bash
   go test ./... -v
   ```

4. **Check coverage**:
   ```bash
   go test ./... -coverprofile=coverage.out
   go tool cover -html=coverage.out
   ```

### Quality Standards

- **Test Coverage**: Maintain ≥ 90% test coverage for all new code
- **Code Style**: Follow standard Go formatting (`go fmt`) and linting rules
- **Documentation**: Add comprehensive tests and documentation for new features
- **Error Handling**: Include proper error handling with context-rich messages
- **Performance**: Consider performance implications and add benchmarks for critical paths

### Contribution Workflow

1. **Fork and branch**: Create feature branch from main
2. **Implement with TDD**: Write tests first, then implementation
3. **Verify quality**: Ensure tests pass and coverage is maintained
4. **Document changes**: Update README and add examples if needed
5. **Submit PR**: Include clear description and test results

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](../../../LICENSE) file for details.

## 🔗 Related Packages

- [`internal/watch/core`](../core/) - Core interfaces and types for the watch system
- [`internal/watch/coordinator`](../coordinator/) - Watch coordination and orchestration
- [`internal/watch/debouncer`](../debouncer/) - Event debouncing and rate limiting
- [`internal/test/runner`](../../test/runner/) - Test execution engine that uses this watcher
- [`internal/ui/display`](../../ui/display/) - UI components that display watch results
- [`pkg/models`](../../../pkg/models/) - Shared data models used across packages

---

**🎯 Package Status**: Production-ready with 91.0% test coverage, comprehensive error handling, and cross-platform compatibility. Actively maintained and following clean architecture principles. 