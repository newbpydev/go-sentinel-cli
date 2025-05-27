# ⚡ Lifecycle Package

[![Test Coverage](https://img.shields.io/badge/coverage-100.0%25-brightgreen.svg)](https://github.com/newbpydev/go-sentinel/tree/main/internal/lifecycle)
[![Go Report Card](https://goreportcard.com/badge/github.com/newbpydev/go-sentinel/internal/lifecycle)](https://goreportcard.com/report/github.com/newbpydev/go-sentinel/internal/lifecycle)
[![Go Reference](https://pkg.go.dev/badge/github.com/newbpydev/go-sentinel/internal/lifecycle.svg)](https://pkg.go.dev/github.com/newbpydev/go-sentinel/internal/lifecycle)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## 📖 Overview

The `lifecycle` package provides comprehensive application lifecycle management for the Go Sentinel CLI. It handles application startup, graceful shutdown, signal handling, and shutdown hook management with proper resource cleanup and timeout handling.

### 🎯 Key Features

- **Application Startup**: Clean application initialization with context management
- **Graceful Shutdown**: Proper resource cleanup with configurable timeouts
- **Signal Handling**: Automatic handling of OS signals (SIGINT, SIGTERM)
- **Shutdown Hooks**: LIFO (Last In, First Out) execution of cleanup functions
- **Context Management**: Application-wide context with cancellation support
- **Thread Safety**: Concurrent-safe lifecycle operations
- **Factory Pattern**: Clean lifecycle manager creation with dependency injection
- **Timeout Configuration**: Configurable shutdown timeouts for different environments

## 🏗️ Architecture

This package follows clean architecture principles:

- **Single Responsibility**: Focuses only on application lifecycle management
- **Dependency Inversion**: Provides interfaces for lifecycle management contracts
- **Interface Segregation**: Small, focused interfaces for specific concerns
- **Observer Pattern**: Shutdown hooks for decoupled cleanup operations

### 📦 Package Structure

```
internal/lifecycle/
├── manager_interface.go    # Lifecycle manager interfaces and contracts
├── manager.go             # Main lifecycle manager implementation
├── factory.go             # Lifecycle manager factory for creation
└── *_test.go             # Comprehensive test suite (100% coverage)
```

## 🚀 Quick Start

### Basic Lifecycle Management

```go
package main

import (
    "context"
    "log"
    "time"
    "github.com/newbpydev/go-sentinel/internal/lifecycle"
)

func main() {
    // Create lifecycle manager factory
    factory := lifecycle.NewAppLifecycleManagerFactory()
    manager := factory.CreateLifecycleManager()
    
    // Start the application
    ctx := context.Background()
    if err := manager.Startup(ctx); err != nil {
        log.Fatal("Failed to start application:", err)
    }
    
    // Register cleanup functions
    manager.RegisterShutdownHook(func() error {
        log.Println("Cleaning up database connections...")
        return nil
    })
    
    manager.RegisterShutdownHook(func() error {
        log.Println("Saving application state...")
        return nil
    })
    
    // Application logic here...
    log.Println("Application running...")
    
    // Wait for shutdown signal
    <-manager.ShutdownChannel()
    
    // Graceful shutdown
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := manager.Shutdown(shutdownCtx); err != nil {
        log.Printf("Shutdown error: %v", err)
    }
    
    log.Println("Application stopped")
}
```

### Custom Context and Timeout

```go
package main

import (
    "context"
    "time"
    "github.com/newbpydev/go-sentinel/internal/lifecycle"
)

func main() {
    // Create custom context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()
    
    // Create factory
    factory := lifecycle.NewAppLifecycleManagerFactory()
    
    // Create manager with custom context
    manager := factory.CreateLifecycleManagerWithContext(ctx)
    
    // Start with custom context
    if err := manager.Startup(ctx); err != nil {
        log.Fatal("Startup failed:", err)
    }
    
    // Use the manager's context for application operations
    appCtx := manager.Context()
    
    // Run application with context cancellation support
    go func() {
        select {
        case <-appCtx.Done():
            log.Println("Application context cancelled")
            return
        case <-time.After(1 * time.Hour):
            log.Println("Application completed normally")
        }
    }()
    
    // Wait for shutdown
    <-manager.ShutdownChannel()
    manager.Shutdown(context.Background())
}
```

### Advanced Shutdown Hook Management

```go
package main

import (
    "context"
    "database/sql"
    "net/http"
    "github.com/newbpydev/go-sentinel/internal/lifecycle"
)

func main() {
    factory := lifecycle.NewAppLifecycleManagerFactory()
    manager := factory.CreateLifecycleManager()
    
    // Initialize resources
    db, _ := sql.Open("postgres", "connection-string")
    server := &http.Server{Addr: ":8080"}
    
    // Register shutdown hooks in order (executed in reverse order)
    
    // 1. Stop accepting new connections (executed last)
    manager.RegisterShutdownHook(func() error {
        log.Println("Stopping HTTP server...")
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        return server.Shutdown(ctx)
    })
    
    // 2. Close database connections (executed second)
    manager.RegisterShutdownHook(func() error {
        log.Println("Closing database connections...")
        return db.Close()
    })
    
    // 3. Save application state (executed first)
    manager.RegisterShutdownHook(func() error {
        log.Println("Saving application state...")
        return saveApplicationState()
    })
    
    // Start application
    manager.Startup(context.Background())
    
    // Start HTTP server
    go func() {
        if err := server.ListenAndServe(); err != http.ErrServerClosed {
            log.Printf("HTTP server error: %v", err)
        }
    }()
    
    // Wait for shutdown signal
    <-manager.ShutdownChannel()
    
    // Graceful shutdown (hooks executed in LIFO order)
    manager.Shutdown(context.Background())
}
```

## 🔧 Lifecycle Manager Interface

### AppLifecycleManager

The main lifecycle manager interface providing all lifecycle management functionality:

```go
type AppLifecycleManager interface {
    // Core lifecycle methods
    Startup(ctx context.Context) error
    Shutdown(ctx context.Context) error
    IsRunning() bool

    // Shutdown hook management
    RegisterShutdownHook(hook func() error)

    // Context and channel access
    Context() context.Context
    ShutdownChannel() <-chan struct{}
}
```

### Factory Interface

Factory interface for creating lifecycle managers:

```go
type AppLifecycleManagerFactory interface {
    CreateLifecycleManager() AppLifecycleManager
    CreateLifecycleManagerWithContext(ctx context.Context) AppLifecycleManager
}
```

### Dependencies Structure

Configuration for lifecycle manager creation:

```go
type AppLifecycleManagerDependencies struct {
    Context         context.Context
    ShutdownTimeout string // Duration string like "30s"
}
```

## 🔄 Advanced Usage

### Signal Handling Integration

```go
func main() {
    manager := factory.CreateLifecycleManager()
    
    // Start application (automatically sets up signal handling)
    manager.Startup(context.Background())
    
    // The manager automatically handles:
    // - SIGINT (Ctrl+C)
    // - SIGTERM (termination signal)
    // - Context cancellation
    
    // Your application logic
    runApplication(manager.Context())
    
    // Wait for any shutdown trigger
    <-manager.ShutdownChannel()
    
    // Graceful shutdown
    manager.Shutdown(context.Background())
}
```

### Error Handling in Shutdown Hooks

```go
func setupShutdownHooks(manager lifecycle.AppLifecycleManager) {
    // Hook that might fail
    manager.RegisterShutdownHook(func() error {
        if err := cleanupResource1(); err != nil {
            // Log error but don't stop shutdown process
            log.Printf("Failed to cleanup resource1: %v", err)
            return err // This will stop further hook execution
        }
        return nil
    })
    
    // Critical cleanup that should always run
    manager.RegisterShutdownHook(func() error {
        // This runs first (LIFO order)
        return criticalCleanup()
    })
    
    // Graceful cleanup with timeout
    manager.RegisterShutdownHook(func() error {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        
        return gracefulCleanupWithTimeout(ctx)
    })
}
```

### Context Propagation

```go
func runApplicationWithContext(manager lifecycle.AppLifecycleManager) {
    // Get application context from manager
    ctx := manager.Context()
    
    // Use context in goroutines
    go func() {
        ticker := time.NewTicker(1 * time.Second)
        defer ticker.Stop()
        
        for {
            select {
            case <-ctx.Done():
                log.Println("Background task cancelled")
                return
            case <-ticker.C:
                log.Println("Background task tick")
            }
        }
    }()
    
    // Use context in HTTP clients
    client := &http.Client{}
    req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.example.com", nil)
    
    resp, err := client.Do(req)
    if err != nil {
        if ctx.Err() != nil {
            log.Println("Request cancelled due to shutdown")
        } else {
            log.Printf("Request failed: %v", err)
        }
        return
    }
    defer resp.Body.Close()
}
```

## 🧪 Testing

The package achieves **100% test coverage** with comprehensive test suites:

### Running Tests

```bash
# Run all tests
go test ./internal/lifecycle/...

# Run with coverage
go test ./internal/lifecycle/... -coverprofile=coverage.out

# View coverage report
go tool cover -html=coverage.out
```

### Test Categories

- **Unit Tests**: Individual lifecycle manager testing
- **Integration Tests**: Signal handling and context management
- **Concurrency Tests**: Thread-safety validation with 100+ goroutines
- **Shutdown Hook Tests**: LIFO execution order and error handling
- **Timeout Tests**: Shutdown timeout behavior
- **Memory Tests**: Resource cleanup and memory efficiency

### Example Test Structure

```go
func TestAppLifecycleManager_Startup_Success(t *testing.T) {
    t.Parallel()
    
    factory := NewAppLifecycleManagerFactory()
    manager := factory.CreateLifecycleManager()
    
    // Test successful startup
    ctx := context.Background()
    err := manager.Startup(ctx)
    
    assert.NoError(t, err)
    assert.True(t, manager.IsRunning())
    
    // Cleanup
    manager.Shutdown(ctx)
}

func TestAppLifecycleManager_ShutdownHooks_LIFOOrder(t *testing.T) {
    t.Parallel()
    
    manager := factory.CreateLifecycleManager()
    
    var executionOrder []string
    
    // Register hooks in order
    manager.RegisterShutdownHook(func() error {
        executionOrder = append(executionOrder, "first")
        return nil
    })
    
    manager.RegisterShutdownHook(func() error {
        executionOrder = append(executionOrder, "second")
        return nil
    })
    
    manager.RegisterShutdownHook(func() error {
        executionOrder = append(executionOrder, "third")
        return nil
    })
    
    // Start and shutdown
    manager.Startup(context.Background())
    manager.Shutdown(context.Background())
    
    // Verify LIFO execution order
    expected := []string{"third", "second", "first"}
    assert.Equal(t, expected, executionOrder)
}
```

## 📊 Performance

The package is optimized for performance:

- **Fast Startup**: Minimal overhead for application initialization
- **Efficient Signal Handling**: Low-latency signal processing
- **Memory Efficient**: Minimal memory allocation for lifecycle management
- **Concurrent Safe**: Thread-safe operations with minimal locking

### Benchmarks

```bash
# Run performance benchmarks
go test ./internal/lifecycle/... -bench=.

# Example results:
BenchmarkLifecycleManager_Startup-8         1000000    1.2μs/op    64B/op
BenchmarkLifecycleManager_Shutdown-8         500000    2.1μs/op    96B/op
BenchmarkLifecycleManager_IsRunning-8      10000000    0.1μs/op     0B/op
```

## 🔍 Error Handling

The package provides comprehensive error handling:

### Error Types

```go
// Lifecycle operation errors
type LifecycleError struct {
    Operation string
    Cause     error
    Message   string
}

// Shutdown hook errors
type ShutdownHookError struct {
    HookIndex int
    Cause     error
    Message   string
}

// Timeout errors
type TimeoutError struct {
    Operation string
    Timeout   time.Duration
    Message   string
}
```

### Error Handling Examples

```go
// Handle startup errors
err := manager.Startup(ctx)
if err != nil {
    switch e := err.(type) {
    case *lifecycle.LifecycleError:
        log.Printf("Lifecycle operation %s failed: %v", e.Operation, e.Cause)
    default:
        log.Printf("Unexpected startup error: %v", err)
    }
}

// Handle shutdown errors
err = manager.Shutdown(ctx)
if err != nil {
    switch e := err.(type) {
    case *lifecycle.ShutdownHookError:
        log.Printf("Shutdown hook %d failed: %v", e.HookIndex, e.Cause)
    case *lifecycle.TimeoutError:
        log.Printf("Shutdown timed out after %v", e.Timeout)
    default:
        log.Printf("Unexpected shutdown error: %v", err)
    }
}
```

## 🎯 Best Practices

### Startup Sequence

```go
func startApplication() error {
    manager := factory.CreateLifecycleManager()
    
    // 1. Register shutdown hooks first
    registerShutdownHooks(manager)
    
    // 2. Start lifecycle manager
    if err := manager.Startup(context.Background()); err != nil {
        return fmt.Errorf("failed to start lifecycle manager: %w", err)
    }
    
    // 3. Initialize application components
    if err := initializeComponents(manager.Context()); err != nil {
        manager.Shutdown(context.Background())
        return fmt.Errorf("failed to initialize components: %w", err)
    }
    
    return nil
}
```

### Shutdown Hook Registration

```go
func registerShutdownHooks(manager lifecycle.AppLifecycleManager) {
    // Register in dependency order (reverse of startup)
    
    // Last: Stop external interfaces
    manager.RegisterShutdownHook(func() error {
        return stopHTTPServer()
    })
    
    // Middle: Close connections
    manager.RegisterShutdownHook(func() error {
        return closeDatabaseConnections()
    })
    
    // First: Save state
    manager.RegisterShutdownHook(func() error {
        return saveApplicationState()
    })
}
```

### Context Usage

```go
func useLifecycleContext(manager lifecycle.AppLifecycleManager) {
    ctx := manager.Context()
    
    // Use context in all operations
    go backgroundWorker(ctx)
    go httpServer(ctx)
    go databaseWorker(ctx)
    
    // Wait for shutdown
    <-manager.ShutdownChannel()
}

func backgroundWorker(ctx context.Context) {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            log.Println("Background worker shutting down")
            return
        case <-ticker.C:
            // Do work
        }
    }
}
```

## 🔧 Integration Patterns

### Factory Pattern Usage

```go
// Create factory
factory := lifecycle.NewAppLifecycleManagerFactory()

// Create with default context
manager := factory.CreateLifecycleManager()

// Create with custom context
ctx := context.WithValue(context.Background(), "key", "value")
manager := factory.CreateLifecycleManagerWithContext(ctx)
```

### Adapter Pattern Integration

The package integrates with the app package through adapter patterns:

```go
// App package uses lifecycle through adapters
type LifecycleAdapter struct {
    factory *lifecycle.AppLifecycleManagerFactory
    manager lifecycle.AppLifecycleManager
}

func (a *LifecycleAdapter) StartApplication(ctx context.Context) error {
    return a.manager.Startup(ctx)
}

func (a *LifecycleAdapter) StopApplication(ctx context.Context) error {
    return a.manager.Shutdown(ctx)
}

func (a *LifecycleAdapter) RegisterCleanup(cleanup func() error) {
    a.manager.RegisterShutdownHook(cleanup)
}
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](../../CONTRIBUTING.md) for details.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/newbpydev/go-sentinel.git

# Navigate to the lifecycle package
cd go-sentinel/internal/lifecycle

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
- [`internal/events`](../events/README.md) - Event handling system
- [`internal/config`](../config/README.md) - Configuration management
- [`internal/container`](../container/README.md) - Dependency injection

---

**Package Version**: v1.0.0  
**Go Version**: 1.21+  
**Last Updated**: January 2025  
**Maintainer**: Go Sentinel CLI Team 