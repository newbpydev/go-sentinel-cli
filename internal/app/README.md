# App Package

The `app` package is the central orchestration layer for the Go Sentinel CLI application. It coordinates all components, manages the application lifecycle, and provides dependency injection for clean component interaction.

## 🎯 Purpose

This package implements the **Application Layer** of our clean architecture, responsible for:
- **Orchestrating** all application components
- **Managing** application lifecycle (startup, shutdown, cleanup)
- **Coordinating** cross-cutting concerns (logging, monitoring, error handling)
- **Providing** dependency injection for component integration
- **Handling** high-level application flow and business workflows

## 🏗️ Architecture

The app package follows the **Facade** and **Coordinator** patterns to provide a single entry point for application functionality while keeping components decoupled.

```
app/
├── controller.go           # Main application controller
├── lifecycle.go           # Application lifecycle management
├── container.go           # Dependency injection container
├── interfaces.go          # Core interfaces and contracts
├── event_handler.go       # Application-level event handling
├── monitoring.go          # Application monitoring and metrics
├── monitoring_dashboard.go # Advanced monitoring dashboard
└── integration_test.go    # End-to-end integration tests
```

## 🔌 Key Interfaces

### ApplicationController
```go
type ApplicationController interface {
    Run(args []string) error           // Execute main application flow
    Initialize() error                 // Set up application with dependencies
    Shutdown(ctx context.Context) error // Gracefully shut down application
}
```

**Implementation**: Coordinates the entire CLI execution flow:
1. Parse CLI arguments using `config/` package
2. Initialize all required components
3. Execute the requested command (run tests, watch mode, etc.)
4. Handle cleanup and shutdown

### LifecycleManager
```go
type LifecycleManager interface {
    Startup(ctx context.Context) error     // Initialize all components
    Shutdown(ctx context.Context) error    // Stop all components gracefully
    IsRunning() bool                       // Check if application is running
    RegisterShutdownHook(hook func() error) // Register cleanup functions
}
```

**Implementation**: Manages component lifecycle:
- **Startup**: Initialize components in correct dependency order
- **Signal Handling**: Listen for SIGINT/SIGTERM for graceful shutdown
- **Cleanup**: Ensure all resources are properly released
- **Hook System**: Allow components to register cleanup functions

### DependencyContainer
```go
type DependencyContainer interface {
    Register(name string, component interface{}) error     // Register component
    Resolve(name string) (interface{}, error)              // Get component
    ResolveAs(name string, target interface{}) error       // Get component with type casting
    Initialize() error                                      // Initialize all components
    Cleanup() error                                         // Clean up all components
}
```

**Implementation**: Provides dependency injection:
- **Registration**: Components register themselves with the container
- **Resolution**: Components can request dependencies by name/type
- **Lifecycle**: Manages initialization and cleanup of all registered components
- **Type Safety**: Provides type-safe resolution methods

## 📋 Core Components

### 1. Application Controller (`controller.go`)
The main orchestrator that implements the primary application workflow:

```go
func (c *Controller) Run(args []string) error {
    // 1. Parse arguments
    parsedArgs, err := c.argParser.Parse(args)
    if err != nil {
        return fmt.Errorf("failed to parse arguments: %w", err)
    }

    // 2. Load configuration
    config, err := c.configLoader.LoadFromDefaults()
    if err != nil {
        return fmt.Errorf("failed to load configuration: %w", err)
    }

    // 3. Merge CLI args with config
    finalConfig := config.MergeWithCLIArgs(parsedArgs)

    // 4. Initialize components
    if err := c.initializeComponents(finalConfig); err != nil {
        return fmt.Errorf("failed to initialize components: %w", err)
    }

    // 5. Execute the requested command
    return c.executeCommand(parsedArgs, finalConfig)
}
```

### 2. Lifecycle Manager (`lifecycle.go`)
Handles application startup, shutdown, and signal management:

```go
func (l *LifecycleManager) Startup(ctx context.Context) error {
    // Set up signal handling
    l.setupSignalHandling()

    // Initialize components in dependency order
    for _, component := range l.components {
        if err := component.Initialize(ctx); err != nil {
            return fmt.Errorf("failed to initialize %T: %w", component, err)
        }
    }

    l.running = true
    return nil
}
```

### 3. Dependency Container (`container.go`)
Provides dependency injection using reflection and interface matching:

```go
func (c *Container) Register(name string, component interface{}) error {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    if _, exists := c.components[name]; exists {
        return fmt.Errorf("component %s already registered", name)
    }
    
    c.components[name] = component
    return nil
}
```

### 4. Event Handler (`event_handler.go`)
Handles application-level events and cross-cutting concerns:

```go
func (h *EventHandler) OnStartup(ctx context.Context) error {
    // Log application startup
    log.Info("Go Sentinel CLI starting up", 
        "version", h.version,
        "config", h.config.Path)

    // Initialize monitoring
    return h.monitoring.Start(ctx)
}
```

## 🔄 Component Integration

The app package integrates with all other internal packages:

### Configuration Integration
```go
// Load and validate configuration
config, err := c.configLoader.LoadFromDefaults()
if err != nil {
    return fmt.Errorf("configuration error: %w", err)
}

// Apply CLI argument overrides
mergedConfig := config.MergeWithCLIArgs(args)
```

### Test Execution Integration
```go
// Initialize test runner with configuration
testRunner := c.container.MustResolve("testRunner").(test.Runner)

// Execute tests based on command
if args.Watch {
    return c.executeWatchMode(testRunner, config)
} else {
    return c.executeTestRun(testRunner, config)
}
```

### Watch Mode Integration
```go
// Set up watch coordinator
watchCoordinator := c.container.MustResolve("watchCoordinator").(watch.Coordinator)

// Start file monitoring
if err := watchCoordinator.Start(ctx, config.Watch); err != nil {
    return fmt.Errorf("failed to start watch mode: %w", err)
}
```

### UI Integration
```go
// Initialize display renderer
renderer := c.container.MustResolve("displayRenderer").(ui.Renderer)

// Configure based on terminal capabilities
renderer.Configure(ui.Config{
    Colors: config.Colors,
    Icons:  config.Visual.Icons,
    Theme:  config.Visual.Theme,
})
```

## 🧪 Testing

### Unit Tests
Test individual components in isolation:

```go
func TestController_Run_Success(t *testing.T) {
    // Arrange
    mockArgParser := &MockArgumentParser{}
    mockConfigLoader := &MockConfigLoader{}
    controller := NewController(mockArgParser, mockConfigLoader)

    // Act
    err := controller.Run([]string{"run", "./internal/config"})

    // Assert
    assert.NoError(t, err)
    assert.True(t, mockArgParser.ParseCalled)
    assert.True(t, mockConfigLoader.LoadCalled)
}
```

### Integration Tests
Test complete application workflows:

```go
func TestApplicationFlow_EndToEnd(t *testing.T) {
    // Test complete CLI execution from args to output
    app := setupTestApplication(t)
    
    output, err := app.RunWithOutput([]string{"run", "--verbose", "./testdata"})
    
    assert.NoError(t, err)
    assert.Contains(t, output, "Tests completed")
    assert.Contains(t, output, "✓") // Success icons
}
```

### Running Tests
```bash
# Run app package tests
go test ./internal/app/

# Run with coverage
go test -cover ./internal/app/

# Run integration tests specifically
go test -run TestApplicationFlow ./internal/app/

# Benchmark application startup
go test -bench=BenchmarkStartup ./internal/app/
```

## 📊 Monitoring & Observability

The app package includes comprehensive monitoring:

### Metrics Collection
```go
// Application-level metrics
AppStartupDuration   = prometheus.NewHistogram(...)
ComponentInitTime    = prometheus.NewHistogramVec(...)
CommandExecutionTime = prometheus.NewHistogramVec(...)
ErrorCount          = prometheus.NewCounterVec(...)
```

### Health Checks
```go
func (c *Controller) HealthCheck() error {
    // Check all critical components
    for name, component := range c.container.components {
        if healthChecker, ok := component.(HealthChecker); ok {
            if err := healthChecker.HealthCheck(); err != nil {
                return fmt.Errorf("component %s unhealthy: %w", name, err)
            }
        }
    }
    return nil
}
```

### Advanced Monitoring Dashboard
The package includes a comprehensive monitoring dashboard (`monitoring_dashboard.go`) with:
- **Real-time Metrics**: Live application performance data
- **Health Monitoring**: Component health and status tracking
- **Error Tracking**: Comprehensive error monitoring and alerting
- **Performance Analytics**: Historical performance trends and analysis

## 🔧 Configuration

The app package respects all configuration options:

```json
{
  "app": {
    "logLevel": "info",
    "shutdownTimeout": "30s",
    "maxConcurrency": 4,
    "enableMetrics": true,
    "enableHealthChecks": true
  }
}
```

### Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `logLevel` | `"info"` | Application logging level |
| `shutdownTimeout` | `"30s"` | Maximum time for graceful shutdown |
| `maxConcurrency` | `4` | Maximum concurrent operations |
| `enableMetrics` | `true` | Enable metrics collection |
| `enableHealthChecks` | `true` | Enable health check endpoints |

## 🚀 Performance

### Startup Performance
- **Target**: < 500ms application startup time
- **Optimization**: Lazy initialization of expensive components
- **Monitoring**: Startup time metrics and alerts

### Memory Management
- **Resource Cleanup**: Proper cleanup of all components
- **Memory Leaks**: Monitoring for goroutine and memory leaks
- **Resource Limits**: Configurable resource limits

### Concurrency
- **Thread Safety**: All components are thread-safe
- **Context Propagation**: Proper context usage throughout
- **Graceful Shutdown**: Clean shutdown of all goroutines

## 🔗 Dependencies

### Internal Dependencies
- `internal/config` - Configuration management
- `internal/test` - Test execution coordination
- `internal/watch` - Watch mode coordination
- `internal/ui` - Display coordination
- `pkg/models` - Shared data structures
- `pkg/events` - Event system integration

### External Dependencies
- `context` - Context management
- `fmt` - Error formatting
- `log/slog` - Structured logging
- `sync` - Concurrency primitives
- `os/signal` - Signal handling

## 📚 Examples

### Basic Application Setup
```go
func main() {
    // Create application components
    controller := app.NewController()
    
    // Initialize the application
    if err := controller.Initialize(); err != nil {
        log.Fatal("Failed to initialize application:", err)
    }
    
    // Run the application
    if err := controller.Run(os.Args[1:]); err != nil {
        log.Fatal("Application error:", err)
    }
}
```

### Custom Component Registration
```go
func setupApplication() *app.Controller {
    container := app.NewContainer()
    
    // Register custom test runner
    customRunner := &MyTestRunner{}
    container.Register("testRunner", customRunner)
    
    // Register custom display renderer
    customRenderer := &MyRenderer{}
    container.Register("displayRenderer", customRenderer)
    
    return app.NewControllerWithContainer(container)
}
```

### Event Handling
```go
func (h *MyEventHandler) OnStartup(ctx context.Context) error {
    log.Info("Custom startup logic")
    return h.initializeCustomComponents()
}

func (h *MyEventHandler) OnShutdown(ctx context.Context) error {
    log.Info("Custom shutdown logic")
    return h.cleanupCustomComponents()
}
```

---

The app package is the foundation that enables all other components to work together seamlessly, providing a clean, testable, and maintainable application architecture. 