# 📦 Debouncer Package

[![Test Coverage](https://img.shields.io/badge/coverage-97.0%25-brightgreen.svg)](https://github.com/newbpydev/go-sentinel/tree/main/internal/watch/debouncer)
[![Go Report Card](https://goreportcard.com/badge/github.com/newbpydev/go-sentinel/internal/watch/debouncer)](https://goreportcard.com/report/github.com/newbpydev/go-sentinel/internal/watch/debouncer)
[![Go Reference](https://pkg.go.dev/badge/github.com/newbpydev/go-sentinel/internal/watch/debouncer.svg)](https://pkg.go.dev/github.com/newbpydev/go-sentinel/internal/watch/debouncer)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## 📖 Overview

The debouncer package provides event temporal processing capabilities for the Go Sentinel CLI file watching system. It implements two complementary debouncing strategies to prevent excessive test execution during rapid file changes.

### 🎯 Key Features

- **Dual Implementation**: Two debouncing strategies for different use cases
- **Event Deduplication**: Automatically deduplicates events by file path
- **Configurable Intervals**: Adjustable debounce timing for different scenarios
- **Thread-Safe Operations**: Concurrent access patterns with proper synchronization
- **Graceful Shutdown**: Clean resource cleanup with pending event flushing
- **Channel-Based Architecture**: Non-blocking event processing with buffered channels

## 🏗️ Architecture

The package follows clean architecture principles with two main implementations:

### 📦 Package Structure

```
internal/watch/debouncer/
├── debouncer.go              # Core debouncer implementation (149 lines)
├── file_debouncer.go         # File-specific debouncer implementation (180 lines)
├── debouncer_test.go         # Comprehensive test suite (969+ lines)
├── file_debouncer_test.go    # File debouncer tests (400 lines)
└── README.md                 # This documentation
```

### 🔧 Core Interfaces

Both implementations satisfy the `core.EventDebouncer` interface:

```go
type EventDebouncer interface {
    AddEvent(event FileEvent)
    Events() <-chan []FileEvent
    SetInterval(interval time.Duration)
    Stop() error
}
```

## 🚀 Quick Start

### Basic Debouncer Usage

```go
package main

import (
    "time"
    "github.com/newbpydev/go-sentinel/internal/watch/debouncer"
    "github.com/newbpydev/go-sentinel/internal/watch/core"
)

func main() {
    // Create a debouncer with 250ms interval
    deb := debouncer.NewDebouncer(250 * time.Millisecond)
    defer deb.Stop()
    
    // Add file events
    deb.AddEvent(core.FileEvent{
        Path: "main.go",
        Type: "write",
        Timestamp: time.Now(),
    })
    
    // Process debounced events
    select {
    case events := <-deb.Events():
        fmt.Printf("Processing %d debounced events\n", len(events))
        for _, event := range events {
            fmt.Printf("File: %s, Type: %s\n", event.Path, event.Type)
        }
    case <-time.After(1 * time.Second):
        fmt.Println("No events received")
    }
}
```

### File Event Debouncer Usage

```go
package main

import (
    "time"
    "github.com/newbpydev/go-sentinel/internal/watch/debouncer"
    "github.com/newbpydev/go-sentinel/internal/watch/core"
)

func main() {
    // Create a file event debouncer with 500ms interval
    deb := debouncer.NewFileEventDebouncer(500 * time.Millisecond)
    defer deb.Stop()
    
    // Add multiple events for the same file (will be deduplicated)
    deb.AddEvent(core.FileEvent{Path: "test.go", Type: "write"})
    deb.AddEvent(core.FileEvent{Path: "test.go", Type: "modify"})
    deb.AddEvent(core.FileEvent{Path: "test.go", Type: "create"}) // Final event wins
    
    // Process deduplicated events
    events := <-deb.Events()
    fmt.Printf("Received %d deduplicated events\n", len(events))
    // Output: Received 1 deduplicated events (type: "create")
}
```

## 🔄 Advanced Usage

### Dynamic Interval Adjustment

```go
deb := debouncer.NewDebouncer(100 * time.Millisecond)
defer deb.Stop()

// Adjust interval based on system load
if highLoad {
    deb.SetInterval(500 * time.Millisecond) // Longer debounce for high load
} else {
    deb.SetInterval(50 * time.Millisecond)  // Shorter debounce for low load
}
```

### Concurrent Event Processing

```go
deb := debouncer.NewFileEventDebouncer(250 * time.Millisecond)
defer deb.Stop()

// Multiple goroutines can safely add events
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        deb.AddEvent(core.FileEvent{
            Path: fmt.Sprintf("file_%d.go", id),
            Type: "write",
            Timestamp: time.Now(),
        })
    }(i)
}

wg.Wait()

// Process all events
events := <-deb.Events()
fmt.Printf("Processed %d concurrent events\n", len(events))
```

## 🧪 Testing

The package achieves **97.0% test coverage** with comprehensive test suites covering:

### Test Categories

- **Unit Tests**: Individual debouncer functionality testing
- **Integration Tests**: Cross-component interaction validation  
- **Concurrency Tests**: Thread-safety validation with 100+ goroutines
- **Edge Case Tests**: Boundary conditions and error handling
- **Performance Tests**: Memory efficiency and timing validation
- **Race Condition Tests**: Concurrent access pattern verification

### Running Tests

```bash
# Run all tests
go test ./internal/watch/debouncer/...

# Run with coverage
go test ./internal/watch/debouncer/... -coverprofile=coverage.out

# View coverage report
go tool cover -html=coverage.out
```

### Test Coverage Breakdown

```bash
$ go test -cover
PASS
coverage: 97.0% of statements

# Function-level coverage:
NewDebouncer                100.0%
AddEvent                    100.0%
Events                      100.0%
SetInterval                 100.0%
Stop                        100.0%
flushPendingEvents          75.0%  # Race conditions in concurrent code
NewFileEventDebouncer       100.0%
FileEventDebouncer.AddEvent 100.0%
FileEventDebouncer.Events   100.0%
FileEventDebouncer.SetInterval 100.0%
FileEventDebouncer.Stop     100.0%
FileEventDebouncer.run      100.0%
FileEventDebouncer.flushPendingEvents 100.0%
```

## 📊 Performance

The package is optimized for performance with:

### Benchmarks

```bash
# Run performance benchmarks
go test ./internal/watch/debouncer/... -bench=.

# Example results:
BenchmarkDebouncer_AddEvent-8           1000000    1.2μs/op    64B/op
BenchmarkDebouncer_Events-8              500000    2.1μs/op    96B/op
BenchmarkFileEventDebouncer_AddEvent-8   800000    1.5μs/op    72B/op
```

### Memory Efficiency

- **Minimal Allocations**: Efficient data structure management
- **Bounded Channels**: 10-event buffer prevents memory bloat
- **Event Deduplication**: Map-based deduplication reduces memory usage
- **Graceful Cleanup**: Proper resource management prevents leaks

### Concurrency Performance

- **Lock-Free Channels**: Primary communication via Go channels
- **Minimal Locking**: RWMutex only for critical sections
- **Non-Blocking Operations**: Default cases prevent goroutine blocking
- **Efficient Timer Management**: Proper timer cleanup and reset

## 🔍 Error Handling

The package implements comprehensive error handling:

### Error Types and Handling

```go
// Graceful degradation when stopped
deb.AddEvent(event) // Safe to call after Stop() - events ignored

// Non-blocking channel operations
select {
case events := <-deb.Events():
    // Process events
default:
    // Channel might be full - handle gracefully
}

// Safe concurrent access
err := deb.Stop() // Always returns nil, safe to call multiple times
```

### Edge Cases Handled

- **Post-Stop Operations**: Events added after Stop() are safely ignored
- **Channel Blocking**: Default cases prevent goroutine deadlocks
- **Timer Race Conditions**: Proper synchronization prevents timer leaks
- **Concurrent Stop**: Multiple Stop() calls are safe and idempotent
- **Empty Event Batches**: Graceful handling of empty pending events

## 🎯 Best Practices

### Usage Recommendations

1. **Choose the Right Implementation**:
   - Use `Debouncer` for simple event batching
   - Use `FileEventDebouncer` for file-specific deduplication

2. **Interval Selection**:
   - **Fast Response**: 50-100ms for interactive applications
   - **Balanced**: 250-500ms for general file watching
   - **Conservative**: 1-2s for high-load scenarios

3. **Resource Management**:
   ```go
   deb := debouncer.NewDebouncer(interval)
   defer deb.Stop() // Always ensure cleanup
   ```

4. **Event Processing**:
   ```go
   // Non-blocking event consumption
   select {
   case events := <-deb.Events():
       processEvents(events)
   case <-time.After(timeout):
       // Handle timeout gracefully
   }
   ```

5. **Concurrent Usage**:
   ```go
   // Safe concurrent event addition
   go func() {
       deb.AddEvent(event) // Thread-safe
   }()
   ```

## 🤝 Contributing

### Development Setup

1. **Clone Repository**: `git clone https://github.com/newbpydev/go-sentinel`
2. **Navigate to Package**: `cd internal/watch/debouncer`
3. **Run Tests**: `go test -v`
4. **Check Coverage**: `go test -cover`

### Quality Standards

- **Test Coverage**: Maintain ≥95% coverage for new code
- **Concurrency Safety**: All public methods must be thread-safe
- **Performance**: Benchmark critical paths for regressions
- **Documentation**: Update README for API changes

### Testing Guidelines

- **TDD Approach**: Write tests before implementation
- **Edge Cases**: Test boundary conditions and error scenarios
- **Race Conditions**: Use `go test -race` for concurrency validation
- **Performance**: Include benchmarks for performance-critical code

## 📄 License

This package is licensed under the MIT License. See the [LICENSE](../../../LICENSE) file for details.

## 🔗 Related Packages

- [`internal/watch/core`](../core/README.md) - Core watch interfaces and types
- [`internal/watch/watcher`](../watcher/README.md) - File system monitoring
- [`internal/watch/coordinator`](../coordinator/README.md) - Watch coordination
- [`pkg/models`](../../../pkg/models/README.md) - Shared data models

---

**Note**: The remaining 3% of uncovered code consists of race condition edge cases in concurrent timer management that are extremely difficult to trigger deterministically in tests. The package is production-ready with excellent coverage of all critical paths and error conditions. 