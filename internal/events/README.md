# 🎯 Events Package

[![Test Coverage](https://img.shields.io/badge/coverage-100.0%25-brightgreen.svg)](https://github.com/newbpydev/go-sentinel/tree/main/internal/events)
[![Go Report Card](https://goreportcard.com/badge/github.com/newbpydev/go-sentinel/internal/events)](https://goreportcard.com/report/github.com/newbpydev/go-sentinel/internal/events)
[![Go Reference](https://pkg.go.dev/badge/github.com/newbpydev/go-sentinel/internal/events.svg)](https://pkg.go.dev/github.com/newbpydev/go-sentinel/internal/events)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## 📖 Overview

The `events` package provides comprehensive application event handling implementation for the Go Sentinel CLI. It manages application lifecycle events, test execution events, file system events, and provides structured logging with configurable verbosity levels.

### 🎯 Key Features

- **Lifecycle Event Handling**: Startup, shutdown, and error event management
- **Test Event Processing**: Test start, completion, and result tracking
- **File System Events**: Watch mode file change event handling
- **Structured Logging**: Configurable verbosity with debug, info, warning, and error levels
- **Configuration Integration**: Dynamic configuration updates and event-driven responses
- **Factory Pattern**: Clean event handler creation with dependency injection
- **Thread Safety**: Concurrent event handling with proper synchronization

## 🏗️ Architecture

This package follows clean architecture principles:

- **Single Responsibility**: Focuses only on event handling and logging
- **Dependency Inversion**: Provides interfaces for event handling contracts
- **Interface Segregation**: Small, focused interfaces for specific concerns
- **Observer Pattern**: Event-driven architecture for loose coupling

### 📦 Package Structure

```
internal/events/
├── handler_interface.go    # Event handler interfaces and contracts
├── handler.go             # Main event handler implementation
├── factory.go             # Event handler factory for creation
└── *_test.go             # Comprehensive test suite (100% coverage)
```

## 🚀 Quick Start

### Basic Event Handler Usage

```go
package main

import (
    "context"
    "log"
    "os"
    "github.com/newbpydev/go-sentinel/internal/events"
)

func main() {
    // Create event handler factory
    factory := events.NewAppEventHandlerFactory()
    
    // Create event handler with custom logger
    logger := log.New(os.Stdout, "[EVENTS] ", log.LstdFlags)
    handler := factory.CreateEventHandlerWithLogger(logger)
    
    // Set verbosity level
    handler.SetVerbosity(2) // Info level
    
    // Handle application startup
    ctx := context.Background()
    if err := handler.OnStartup(ctx); err != nil {
        log.Fatal("Failed to handle startup:", err)
    }
    
    // Handle test events
    handler.OnTestStart("TestExample")
    handler.OnTestComplete("TestExample", true)
    
    // Handle file system events
    handler.OnWatchEvent("/path/to/file.go", "modified")
    
    // Handle application shutdown
    if err := handler.OnShutdown(ctx); err != nil {
        log.Fatal("Failed to handle shutdown:", err)
    }
}
```

### Configuration-Driven Event Handling

```go
package main

import (
    "github.com/newbpydev/go-sentinel/internal/events"
)

func main() {
    factory := events.NewAppEventHandlerFactory()
    handler := factory.CreateEventHandler()
    
    // Configure event handler with application config
    config := &events.AppConfig{
        Colors:    true,
        Verbosity: 3, // Debug level
        Watch: events.AppWatchConfig{
            Enabled:      true,
            Debounce:     "500ms",
            RunOnStart:   true,
            ClearOnRerun: true,
        },
        Paths: events.AppPathsConfig{
            IncludePatterns: []string{"**/*.go"},
            ExcludePatterns: []string{"vendor/**"},
            IgnorePatterns:  []string{"**/.git/**"},
        },
    }
    
    // Apply configuration
    handler.OnConfigChanged(config)
    
    // Event handler now uses the new configuration
    handler.LogInfo("Configuration updated successfully")
}
```

### Structured Logging

```go
package main

import (
    "github.com/newbpydev/go-sentinel/internal/events"
)

func main() {
    factory := events.NewAppEventHandlerFactory()
    handler := factory.CreateEventHandler()
    
    // Set high verbosity for detailed logging
    handler.SetVerbosity(4)
    
    // Use structured logging methods
    handler.LogDebug("Debug message: %s", "detailed information")
    handler.LogInfo("Application started successfully")
    handler.LogWarning("Configuration file not found, using defaults")
    
    // Handle errors with structured logging
    err := someOperation()
    if err != nil {
        handler.OnError(err)
        handler.LogError("Operation failed: %v", err)
    }
}
```

## 🔧 Event Handler Interface

### AppEventHandler

The main event handler interface providing all event handling functionality:

```go
type AppEventHandler interface {
    // Core event handling methods
    OnStartup(ctx context.Context) error
    OnShutdown(ctx context.Context) error
    OnError(err error)
    OnConfigChanged(config *AppConfig)

    // Extended event handling methods
    OnTestStart(testName string)
    OnTestComplete(testName string, success bool)
    OnWatchEvent(filePath string, eventType string)

    // Logger management
    SetLogger(logger *log.Logger)
    SetVerbosity(level int)
    GetLogger() *log.Logger

    // Logging utilities
    LogDebug(format string, args ...interface{})
    LogInfo(format string, args ...interface{})
    LogWarning(format string, args ...interface{})
    LogError(format string, args ...interface{})
}
```

### Configuration Types

Event-specific configuration structures:

```go
// AppConfig represents application configuration for event handling
type AppConfig struct {
    Colors    bool
    Verbosity int
    Watch     AppWatchConfig
    Paths     AppPathsConfig
}

// AppWatchConfig represents watch configuration for event logging
type AppWatchConfig struct {
    Enabled      bool
    Debounce     string
    RunOnStart   bool
    ClearOnRerun bool
}

// AppPathsConfig represents paths configuration for event logging
type AppPathsConfig struct {
    IncludePatterns []string
    ExcludePatterns []string
    IgnorePatterns  []string
}
```

## 🔄 Advanced Usage

### Custom Event Handler Implementation

```go
type CustomEventHandler struct {
    *events.DefaultAppEventHandler
    metrics *MetricsCollector
}

func (c *CustomEventHandler) OnTestStart(testName string) {
    // Call parent implementation
    c.DefaultAppEventHandler.OnTestStart(testName)
    
    // Add custom metrics collection
    c.metrics.IncrementTestsStarted()
    c.metrics.RecordTestStart(testName, time.Now())
}

func (c *CustomEventHandler) OnTestComplete(testName string, success bool) {
    // Call parent implementation
    c.DefaultAppEventHandler.OnTestComplete(testName, success)
    
    // Add custom metrics collection
    if success {
        c.metrics.IncrementTestsPassed()
    } else {
        c.metrics.IncrementTestsFailed()
    }
    c.metrics.RecordTestComplete(testName, time.Now())
}
```

### Event-Driven Configuration Updates

```go
func main() {
    handler := factory.CreateEventHandler()
    
    // Watch for configuration changes
    configWatcher := &ConfigWatcher{
        handler: handler,
    }
    
    // When configuration changes, update event handler
    configWatcher.OnConfigFileChanged = func(newConfig *events.AppConfig) {
        handler.OnConfigChanged(newConfig)
        handler.LogInfo("Configuration reloaded from file")
    }
    
    // Start watching for changes
    configWatcher.Start()
}
```

### Error Handling with Context

```go
func handleApplicationError(handler events.AppEventHandler, err error) {
    // Log the error with context
    handler.OnError(err)
    
    // Determine error severity and log appropriately
    switch {
    case isRecoverableError(err):
        handler.LogWarning("Recoverable error occurred: %v", err)
    case isCriticalError(err):
        handler.LogError("Critical error occurred: %v", err)
        // Trigger shutdown sequence
        ctx := context.Background()
        handler.OnShutdown(ctx)
    default:
        handler.LogInfo("Minor error occurred: %v", err)
    }
}
```

### File System Event Processing

```go
func setupFileWatcher(handler events.AppEventHandler) {
    watcher := &FileSystemWatcher{
        handler: handler,
    }
    
    // Configure event processing
    watcher.OnFileChanged = func(filePath, eventType string) {
        handler.OnWatchEvent(filePath, eventType)
        
        // Process different event types
        switch eventType {
        case "created":
            handler.LogInfo("New file created: %s", filePath)
        case "modified":
            handler.LogDebug("File modified: %s", filePath)
        case "deleted":
            handler.LogWarning("File deleted: %s", filePath)
        case "renamed":
            handler.LogInfo("File renamed: %s", filePath)
        }
    }
    
    watcher.Start()
}
```

## 🧪 Testing

The package achieves **100% test coverage** with comprehensive test suites:

### Running Tests

```bash
# Run all tests
go test ./internal/events/...

# Run with coverage
go test ./internal/events/... -coverprofile=coverage.out

# View coverage report
go tool cover -html=coverage.out
```

### Test Categories

- **Unit Tests**: Individual event handler testing
- **Integration Tests**: Multi-component event workflows
- **Logging Tests**: Structured logging verification
- **Configuration Tests**: Dynamic configuration updates
- **Concurrency Tests**: Thread-safety validation

### Example Test Structure

```go
func TestAppEventHandler_OnStartup_Success(t *testing.T) {
    t.Parallel()
    
    factory := NewAppEventHandlerFactory()
    handler := factory.CreateEventHandler()
    
    // Test successful startup handling
    ctx := context.Background()
    err := handler.OnStartup(ctx)
    
    assert.NoError(t, err)
}

func TestAppEventHandler_LoggingLevels(t *testing.T) {
    t.Parallel()
    
    var buf bytes.Buffer
    logger := log.New(&buf, "", 0)
    
    factory := NewAppEventHandlerFactory()
    handler := factory.CreateEventHandlerWithLogger(logger)
    handler.SetVerbosity(3) // Debug level
    
    // Test different logging levels
    handler.LogDebug("Debug message")
    handler.LogInfo("Info message")
    handler.LogWarning("Warning message")
    handler.LogError("Error message")
    
    output := buf.String()
    assert.Contains(t, output, "Debug message")
    assert.Contains(t, output, "Info message")
    assert.Contains(t, output, "Warning message")
    assert.Contains(t, output, "Error message")
}
```

## 📊 Performance

The package is optimized for performance:

- **Efficient Logging**: Minimal overhead for disabled log levels
- **Non-blocking Events**: Asynchronous event processing where appropriate
- **Memory Efficient**: Minimal memory allocation for event handling
- **Fast Configuration Updates**: O(1) configuration changes

### Benchmarks

```bash
# Run performance benchmarks
go test ./internal/events/... -bench=.

# Example results:
BenchmarkEventHandler_OnTestStart-8      2000000    0.8μs/op    32B/op
BenchmarkEventHandler_LogInfo-8          1000000    1.2μs/op    64B/op
BenchmarkEventHandler_OnConfigChanged-8  5000000    0.3μs/op    16B/op
```

## 🔍 Error Handling

The package provides comprehensive error handling:

### Error Types

```go
// Event handling errors
type EventHandlingError struct {
    EventType string
    Cause     error
    Message   string
}

// Configuration errors
type ConfigurationError struct {
    Field   string
    Value   interface{}
    Message string
}

// Logging errors
type LoggingError struct {
    Level   string
    Message string
    Cause   error
}
```

### Error Handling Examples

```go
// Handle event processing errors
err := handler.OnStartup(ctx)
if err != nil {
    switch e := err.(type) {
    case *events.EventHandlingError:
        log.Printf("Failed to handle %s event: %v", e.EventType, e.Cause)
    default:
        log.Printf("Unexpected event handling error: %v", err)
    }
}

// Handle configuration errors
handler.OnConfigChanged(config)
// Configuration errors are logged internally
```

## 🎯 Verbosity Levels

The package supports multiple verbosity levels:

### Level 0: Silent
- No output except critical errors
- Minimal logging overhead

### Level 1: Error Only
- Error messages only
- Critical application failures

### Level 2: Warning + Error
- Warning and error messages
- Important application events

### Level 3: Info + Warning + Error
- Informational, warning, and error messages
- General application flow

### Level 4: Debug + Info + Warning + Error
- All logging levels enabled
- Detailed debugging information

### Level 5: Verbose Debug
- Maximum verbosity
- Extremely detailed logging

```go
// Set verbosity level
handler.SetVerbosity(3) // Info level

// Check current verbosity
if handler.GetVerbosity() >= 4 {
    handler.LogDebug("Detailed debug information")
}
```

## 🔧 Integration Patterns

### Factory Pattern Usage

```go
// Create factory
factory := events.NewAppEventHandlerFactory()

// Create with default logger
handler := factory.CreateEventHandler()

// Create with custom logger
logger := log.New(os.Stdout, "[APP] ", log.LstdFlags)
handler := factory.CreateEventHandlerWithLogger(logger)
```

### Adapter Pattern Integration

The package integrates with the app package through adapter patterns:

```go
// App package uses events through adapters
type EventHandlerAdapter struct {
    factory *events.AppEventHandlerFactory
    handler events.AppEventHandler
}

func (a *EventHandlerAdapter) HandleStartup(ctx context.Context) error {
    return a.handler.OnStartup(ctx)
}

func (a *EventHandlerAdapter) HandleTestEvent(testName string, success bool) {
    a.handler.OnTestStart(testName)
    a.handler.OnTestComplete(testName, success)
}
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](../../CONTRIBUTING.md) for details.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/newbpydev/go-sentinel.git

# Navigate to the events package
cd go-sentinel/internal/events

# Run tests
go test ./...

# Run tests with coverage
go test ./... -coverprofile=coverage.out

# View coverage
go tool cover -html=coverage.out
```

### Code Quality Standards

- **Test Coverage**: Maintain 100% test coverage
- **Documentation**: All exported symbols must have documentation
- **Linting**: Code must pass `golangci-lint` checks
- **Formatting**: Use `go fmt` for consistent formatting

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](../../LICENSE) file for details.

## 🔗 Related Packages

- [`internal/app`](../app/README.md) - Application orchestration layer
- [`internal/lifecycle`](../lifecycle/README.md) - Application lifecycle management
- [`internal/config`](../config/README.md) - Configuration management
- [`pkg/events`](../../pkg/events/README.md) - Event system interfaces

---

**Package Version**: v1.0.0  
**Go Version**: 1.21+  
**Last Updated**: January 2025  
**Maintainer**: Go Sentinel CLI Team 