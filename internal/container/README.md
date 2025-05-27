# 📦 Container Package

[![Test Coverage](https://img.shields.io/badge/coverage-100.0%25-brightgreen.svg)](https://github.com/newbpydev/go-sentinel/tree/main/internal/container)
[![Go Report Card](https://goreportcard.com/badge/github.com/newbpydev/go-sentinel/internal/container)](https://goreportcard.com/report/github.com/newbpydev/go-sentinel/internal/container)
[![Go Reference](https://pkg.go.dev/badge/github.com/newbpydev/go-sentinel/internal/container.svg)](https://pkg.go.dev/github.com/newbpydev/go-sentinel/internal/container)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## 📖 Overview

The `container` package provides a comprehensive dependency injection container implementation for the Go Sentinel CLI. It manages component registration, resolution, lifecycle, and provides advanced features like singleton patterns and factory functions for clean dependency management.

### 🎯 Key Features

- **Dependency Registration**: Register components by name with type safety
- **Dependency Resolution**: Resolve components with automatic type casting
- **Singleton Support**: Register singleton components with factory functions
- **Lifecycle Management**: Initialize and cleanup components automatically
- **Type Safety**: Compile-time and runtime type checking
- **Factory Pattern**: Clean component creation with dependency injection
- **Interface Compliance**: Components can implement initialization and cleanup interfaces

## 🏗️ Architecture

This package follows clean architecture principles:

- **Single Responsibility**: Focuses only on dependency injection and container management
- **Dependency Inversion**: Provides interfaces for dependency management contracts
- **Interface Segregation**: Small, focused interfaces for specific concerns
- **Factory Pattern**: Clean object creation with proper dependency management

### 📦 Package Structure

```
internal/container/
├── container_interface.go    # Container interfaces and contracts
├── container.go             # Main container implementation
├── factory.go               # Container factory for creation
└── *_test.go               # Comprehensive test suite (100% coverage)
```

## 🚀 Quick Start

### Basic Container Usage

```go
package main

import (
    "github.com/newbpydev/go-sentinel/internal/container"
)

func main() {
    // Create container factory
    factory := container.NewAppDependencyContainerFactory()
    container := factory.CreateContainer()
    
    // Initialize container
    if err := container.Initialize(); err != nil {
        log.Fatal("Failed to initialize container:", err)
    }
    defer container.Cleanup()
    
    // Register a component
    myService := &MyService{Name: "example"}
    if err := container.Register("myService", myService); err != nil {
        log.Fatal("Failed to register service:", err)
    }
    
    // Resolve the component
    resolved, err := container.Resolve("myService")
    if err != nil {
        log.Fatal("Failed to resolve service:", err)
    }
    
    service := resolved.(*MyService)
    fmt.Printf("Service name: %s\n", service.Name)
}
```

### Singleton Registration

```go
package main

import (
    "github.com/newbpydev/go-sentinel/internal/container"
)

func main() {
    factory := container.NewAppDependencyContainerFactory()
    container := factory.CreateContainer()
    
    // Register singleton with factory function
    err := container.RegisterSingleton("database", func() (interface{}, error) {
        return &Database{
            ConnectionString: "localhost:5432",
            MaxConnections:   10,
        }, nil
    })
    
    if err != nil {
        log.Fatal("Failed to register singleton:", err)
    }
    
    // Resolve singleton (same instance returned each time)
    db1, _ := container.Resolve("database")
    db2, _ := container.Resolve("database")
    
    // db1 and db2 are the same instance
    fmt.Printf("Same instance: %v\n", db1 == db2) // true
}
```

### Type-Safe Resolution

```go
package main

import (
    "github.com/newbpydev/go-sentinel/internal/container"
)

func main() {
    factory := container.NewAppDependencyContainerFactory()
    container := factory.CreateContainer()
    
    // Register component
    container.Register("config", &Config{Debug: true})
    
    // Resolve with type safety
    var config *Config
    if err := container.ResolveAs("config", &config); err != nil {
        log.Fatal("Failed to resolve config:", err)
    }
    
    fmt.Printf("Debug mode: %v\n", config.Debug)
}
```

## 🔧 Container Interface

### AppDependencyContainer

The main container interface providing all dependency management functionality:

```go
type AppDependencyContainer interface {
    // Core dependency management methods
    Register(name string, component interface{}) error
    Resolve(name string) (interface{}, error)
    ResolveAs(name string, target interface{}) error

    // Lifecycle management
    Initialize() error
    Cleanup() error

    // Advanced registration methods
    RegisterSingleton(name string, factory AppComponentFactory) error

    // Inspection methods
    ListComponents() []string
    HasComponent(name string) bool
}
```

### Component Factory

Factory function type for creating components:

```go
type AppComponentFactory func() (interface{}, error)
```

### Lifecycle Interfaces

Components can implement these interfaces for automatic lifecycle management:

```go
// AppInitializer interface for components that need initialization
type AppInitializer interface {
    Initialize() error
}

// AppCleaner interface for components that need cleanup
type AppCleaner interface {
    Cleanup() error
}
```

## 🔄 Advanced Usage

### Component Lifecycle Management

```go
type DatabaseService struct {
    connection *sql.DB
}

// Implement AppInitializer
func (d *DatabaseService) Initialize() error {
    conn, err := sql.Open("postgres", "connection-string")
    if err != nil {
        return err
    }
    d.connection = conn
    return nil
}

// Implement AppCleaner
func (d *DatabaseService) Cleanup() error {
    if d.connection != nil {
        return d.connection.Close()
    }
    return nil
}

func main() {
    container := factory.CreateContainer()
    
    // Register component with lifecycle
    dbService := &DatabaseService{}
    container.Register("database", dbService)
    
    // Initialize will call Initialize() on all registered components
    container.Initialize()
    
    // Cleanup will call Cleanup() on all registered components
    defer container.Cleanup()
}
```

### Factory Pattern with Dependencies

```go
func main() {
    container := factory.CreateContainer()
    
    // Register dependencies first
    container.Register("config", &Config{DatabaseURL: "localhost:5432"})
    
    // Register component with factory that uses dependencies
    container.RegisterSingleton("userService", func() (interface{}, error) {
        // Resolve dependencies
        config, err := container.Resolve("config")
        if err != nil {
            return nil, err
        }
        
        // Create component with dependencies
        return &UserService{
            Config: config.(*Config),
        }, nil
    })
    
    // Resolve the service
    userService, err := container.Resolve("userService")
    if err != nil {
        log.Fatal("Failed to resolve user service:", err)
    }
}
```

### Container Inspection

```go
func main() {
    container := factory.CreateContainer()
    
    // Register some components
    container.Register("service1", &Service1{})
    container.Register("service2", &Service2{})
    
    // List all registered components
    components := container.ListComponents()
    fmt.Printf("Registered components: %v\n", components)
    
    // Check if component exists
    if container.HasComponent("service1") {
        fmt.Println("service1 is registered")
    }
    
    // Check non-existent component
    if !container.HasComponent("service3") {
        fmt.Println("service3 is not registered")
    }
}
```

## 🧪 Testing

The package achieves **100% test coverage** with comprehensive test suites:

### Running Tests

```bash
# Run all tests
go test ./internal/container/...

# Run with coverage
go test ./internal/container/... -coverprofile=coverage.out

# View coverage report
go tool cover -html=coverage.out
```

### Test Categories

- **Unit Tests**: Individual component testing
- **Integration Tests**: Multi-component workflows
- **Lifecycle Tests**: Component initialization and cleanup
- **Error Handling Tests**: Edge cases and error conditions
- **Concurrency Tests**: Thread-safety validation

### Example Test Structure

```go
func TestAppDependencyContainer_Register_Success(t *testing.T) {
    t.Parallel()
    
    factory := NewAppDependencyContainerFactory()
    container := factory.CreateContainer()
    
    // Test successful registration
    component := &TestComponent{Name: "test"}
    err := container.Register("testComponent", component)
    
    assert.NoError(t, err)
    assert.True(t, container.HasComponent("testComponent"))
}

func TestAppDependencyContainer_Resolve_TypeSafety(t *testing.T) {
    t.Parallel()
    
    container := factory.CreateContainer()
    container.Register("config", &Config{Debug: true})
    
    // Test type-safe resolution
    var config *Config
    err := container.ResolveAs("config", &config)
    
    assert.NoError(t, err)
    assert.NotNil(t, config)
    assert.True(t, config.Debug)
}
```

## 📊 Performance

The package is optimized for performance:

- **Fast Registration**: O(1) component registration
- **Fast Resolution**: O(1) component resolution with caching
- **Memory Efficient**: Minimal memory overhead per component
- **Singleton Caching**: Singleton instances cached for reuse

### Benchmarks

```bash
# Run performance benchmarks
go test ./internal/container/... -bench=.

# Example results:
BenchmarkContainer_Register-8        1000000    1.2μs/op    64B/op
BenchmarkContainer_Resolve-8         2000000    0.8μs/op    32B/op
BenchmarkContainer_ResolveAs-8       1500000    1.1μs/op    48B/op
```

## 🔍 Error Handling

The package provides comprehensive error handling:

### Error Types

```go
// Component registration errors
type RegistrationError struct {
    ComponentName string
    Cause         error
    Message       string
}

// Component resolution errors
type ResolutionError struct {
    ComponentName string
    RequestedType string
    ActualType    string
    Message       string
}

// Lifecycle errors
type LifecycleError struct {
    ComponentName string
    Operation     string
    Cause         error
}
```

### Error Handling Examples

```go
// Handle registration errors
err := container.Register("service", myService)
if err != nil {
    switch e := err.(type) {
    case *container.RegistrationError:
        log.Printf("Failed to register %s: %v", e.ComponentName, e.Cause)
    default:
        log.Printf("Unexpected registration error: %v", err)
    }
}

// Handle resolution errors
component, err := container.Resolve("nonexistent")
if err != nil {
    switch e := err.(type) {
    case *container.ResolutionError:
        log.Printf("Component %s not found", e.ComponentName)
    default:
        log.Printf("Unexpected resolution error: %v", err)
    }
}
```

## 🔧 Integration Patterns

### Factory Pattern Usage

```go
// Create factory with default settings
factory := container.NewAppDependencyContainerFactory()
container := factory.CreateContainer()

// Create container with custom settings
container := factory.CreateContainerWithDefaults()
```

### Adapter Pattern Integration

The package integrates with the app package through adapter patterns:

```go
// App package uses container through adapters
type ContainerAdapter struct {
    factory   *container.AppDependencyContainerFactory
    container container.AppDependencyContainer
}

func (a *ContainerAdapter) RegisterComponent(name string, component interface{}) error {
    return a.container.Register(name, component)
}

func (a *ContainerAdapter) GetComponent(name string) (interface{}, error) {
    return a.container.Resolve(name)
}
```

## 🎯 Best Practices

### Component Registration

```go
// Register components in dependency order
container.Register("config", config)
container.Register("database", database)
container.Register("userService", userService) // depends on config and database
```

### Singleton Usage

```go
// Use singletons for expensive resources
container.RegisterSingleton("httpClient", func() (interface{}, error) {
    return &http.Client{
        Timeout: 30 * time.Second,
    }, nil
})
```

### Error Handling

```go
// Always check registration and resolution errors
if err := container.Register("service", service); err != nil {
    return fmt.Errorf("failed to register service: %w", err)
}

component, err := container.Resolve("service")
if err != nil {
    return fmt.Errorf("failed to resolve service: %w", err)
}
```

### Lifecycle Management

```go
// Always initialize and cleanup
if err := container.Initialize(); err != nil {
    return fmt.Errorf("failed to initialize container: %w", err)
}
defer func() {
    if err := container.Cleanup(); err != nil {
        log.Printf("Failed to cleanup container: %v", err)
    }
}()
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](../../CONTRIBUTING.md) for details.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/newbpydev/go-sentinel.git

# Navigate to the container package
cd go-sentinel/internal/container

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
- [`internal/events`](../events/README.md) - Event handling system
- [`internal/config`](../config/README.md) - Configuration management

---

**Package Version**: v1.0.0  
**Go Version**: 1.21+  
**Last Updated**: January 2025  
**Maintainer**: Go Sentinel CLI Team 