# 🌐 Events Package (pkg/events)

[![Test Coverage](https://img.shields.io/badge/coverage-100.0%25-brightgreen.svg)](https://github.com/newbpydev/go-sentinel/tree/main/pkg/events)
[![Go Report Card](https://goreportcard.com/badge/github.com/newbpydev/go-sentinel/pkg/events)](https://goreportcard.com/report/github.com/newbpydev/go-sentinel/pkg/events)
[![Go Reference](https://pkg.go.dev/badge/github.com/newbpydev/go-sentinel/pkg/events.svg)](https://pkg.go.dev/github.com/newbpydev/go-sentinel/pkg/events)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## 📖 Overview

The `pkg/events` package provides a comprehensive event system for inter-component communication in the Go Sentinel CLI. It implements the Observer pattern with event buses, handlers, filters, and persistence capabilities for building loosely coupled, event-driven architectures.

### 🎯 Key Features

- **Event Bus**: Publish-subscribe messaging system with async support
- **Event Handlers**: Prioritized event processing with filtering capabilities
- **Event Store**: Persistent event storage with querying and retrieval
- **Event Processing**: Synchronous, asynchronous, and batch processing strategies
- **Event Filtering**: Custom filters for selective event handling
- **Subscription Management**: Active subscription tracking with statistics
- **Metrics & Analytics**: Comprehensive event bus and processing metrics
- **Type Safety**: Strongly typed event interfaces with metadata support

## 🏗️ Architecture

This package follows clean architecture principles:

- **Interface Segregation**: Small, focused interfaces for specific concerns
- **Dependency Inversion**: Abstractions for event system components
- **Observer Pattern**: Decoupled event publishing and subscription
- **Strategy Pattern**: Pluggable event processing strategies

### 📦 Package Structure

```
pkg/events/
├── interfaces.go       # Core event system interfaces
├── examples.go        # Concrete implementations and examples
└── *_test.go         # Comprehensive test suite (100% coverage)
```

## 🚀 Quick Start

### Basic Event Bus Usage

```go
package main

import (
    "context"
    "log"
    "github.com/newbpydev/go-sentinel/pkg/events"
)

func main() {
    // Create event bus (implementation would be provided)
    bus := NewEventBus()
    
    // Create event handler
    handler := &MyEventHandler{}
    
    // Subscribe to test events
    subscription, err := bus.Subscribe("test.started", handler)
    if err != nil {
        log.Fatal("Failed to subscribe:", err)
    }
    defer subscription.Cancel()
    
    // Create and publish event
    event := events.NewTestStartedEvent("TestExample", "example")
    ctx := context.Background()
    
    if err := bus.Publish(ctx, event); err != nil {
        log.Fatal("Failed to publish event:", err)
    }
    
    // Publish asynchronously
    if err := bus.PublishAsync(ctx, event); err != nil {
        log.Fatal("Failed to publish async:", err)
    }
}

type MyEventHandler struct{}

func (h *MyEventHandler) Handle(ctx context.Context, event events.Event) error {
    log.Printf("Received event: %s from %s", event.Type(), event.Source())
    return nil
}

func (h *MyEventHandler) CanHandle(event events.Event) bool {
    return event.Type() == "test.started"
}

func (h *MyEventHandler) Priority() int {
    return 1
}
```

### Event Store Usage

```go
package main

import (
    "context"
    "time"
    "github.com/newbpydev/go-sentinel/pkg/events"
)

func main() {
    // Create event store (implementation would be provided)
    store := NewEventStore()
    ctx := context.Background()
    
    // Store events
    event1 := events.NewTestStartedEvent("TestA", "package1")
    event2 := events.NewTestCompletedEvent("TestA", "package1", 100*time.Millisecond, true)
    
    store.Store(ctx, event1)
    store.Store(ctx, event2)
    
    // Query events
    query := &events.EventQuery{
        EventTypes: []string{"test.started", "test.completed"},
        Sources:    []string{"package1"},
        Limit:      10,
    }
    
    retrievedEvents, err := store.Retrieve(ctx, query)
    if err != nil {
        log.Fatal("Failed to retrieve events:", err)
    }
    
    log.Printf("Retrieved %d events", len(retrievedEvents))
    
    // Count events
    count, err := store.Count(ctx, query)
    if err != nil {
        log.Fatal("Failed to count events:", err)
    }
    
    log.Printf("Total matching events: %d", count)
}
```

### Event Processing

```go
package main

import (
    "context"
    "github.com/newbpydev/go-sentinel/pkg/events"
)

func main() {
    // Create event processor (implementation would be provided)
    processor := NewEventProcessor()
    ctx := context.Background()
    
    // Create events to process
    events := []events.Event{
        events.NewTestStartedEvent("Test1", "pkg1"),
        events.NewTestStartedEvent("Test2", "pkg2"),
        events.NewTestCompletedEvent("Test1", "pkg1", 50*time.Millisecond, true),
    }
    
    // Process synchronously
    if err := processor.ProcessSync(ctx, events); err != nil {
        log.Fatal("Sync processing failed:", err)
    }
    
    // Process asynchronously
    if err := processor.ProcessAsync(ctx, events); err != nil {
        log.Fatal("Async processing failed:", err)
    }
    
    // Process in batches
    batchSize := 2
    if err := processor.ProcessBatch(ctx, events, batchSize); err != nil {
        log.Fatal("Batch processing failed:", err)
    }
    
    // Get processing statistics
    stats := processor.GetProcessingStats()
    log.Printf("Processed: %d, Errors: %d, Rate: %.2f/sec", 
        stats.TotalProcessed, stats.TotalErrors, stats.ProcessingRate)
}
```

## 🔧 Core Interfaces

### Event Interface

The fundamental event interface:

```go
type Event interface {
    ID() string
    Type() string
    Timestamp() time.Time
    Source() string
    Data() interface{}
    Metadata() map[string]interface{}
    String() string
}
```

### EventBus Interface

Central event distribution system:

```go
type EventBus interface {
    Publish(ctx context.Context, event Event) error
    PublishAsync(ctx context.Context, event Event) error
    Subscribe(eventType string, handler EventHandler) (Subscription, error)
    SubscribeWithFilter(filter EventFilter, handler EventHandler) (Subscription, error)
    Unsubscribe(subscription Subscription) error
    Close() error
    GetMetrics() *EventBusMetrics
}
```

### EventHandler Interface

Event processing interface:

```go
type EventHandler interface {
    Handle(ctx context.Context, event Event) error
    CanHandle(event Event) bool
    Priority() int
}
```

### EventStore Interface

Event persistence interface:

```go
type EventStore interface {
    Store(ctx context.Context, event Event) error
    Retrieve(ctx context.Context, query *EventQuery) ([]Event, error)
    Count(ctx context.Context, query *EventQuery) (int, error)
    Delete(ctx context.Context, query *EventQuery) error
    Clear(ctx context.Context) error
}
```

## 🎯 Built-in Event Types

### Test Events

Pre-defined events for test execution:

```go
// Test started event
type TestStartedEvent struct {
    *BaseEvent
    TestName    string
    PackageName string
}

// Test completed event
type TestCompletedEvent struct {
    *BaseEvent
    TestName    string
    PackageName string
    Duration    time.Duration
    Success     bool
}

// File changed event
type FileChangedEvent struct {
    *BaseEvent
    FilePath   string
    ChangeType string
}
```

### Event Creation Functions

```go
// Create test events
testStarted := events.NewTestStartedEvent("TestExample", "mypackage")
testCompleted := events.NewTestCompletedEvent("TestExample", "mypackage", 100*time.Millisecond, true)
fileChanged := events.NewFileChangedEvent("/path/to/file.go", "modified")

// Create custom events
customEvent := events.NewBaseEvent("custom.event", "myservice", map[string]interface{}{
    "key": "value",
    "count": 42,
})
```

## 🔄 Advanced Usage

### Custom Event Handler with Priority

```go
type PriorityEventHandler struct {
    name     string
    priority int
    filter   func(events.Event) bool
}

func (h *PriorityEventHandler) Handle(ctx context.Context, event events.Event) error {
    log.Printf("[%s] Processing event: %s", h.name, event.Type())
    
    // Custom processing logic
    switch event.Type() {
    case "test.started":
        return h.handleTestStarted(ctx, event)
    case "test.completed":
        return h.handleTestCompleted(ctx, event)
    default:
        return nil
    }
}

func (h *PriorityEventHandler) CanHandle(event events.Event) bool {
    if h.filter != nil {
        return h.filter(event)
    }
    return true
}

func (h *PriorityEventHandler) Priority() int {
    return h.priority
}

func (h *PriorityEventHandler) handleTestStarted(ctx context.Context, event events.Event) error {
    // Handle test started event
    return nil
}

func (h *PriorityEventHandler) handleTestCompleted(ctx context.Context, event events.Event) error {
    // Handle test completed event
    return nil
}
```

### Custom Event Filter

```go
type PackageEventFilter struct {
    allowedPackages []string
}

func (f *PackageEventFilter) Match(event events.Event) bool {
    // Check if event source is in allowed packages
    for _, pkg := range f.allowedPackages {
        if event.Source() == pkg {
            return true
        }
    }
    return false
}

func (f *PackageEventFilter) String() string {
    return fmt.Sprintf("PackageFilter(%v)", f.allowedPackages)
}

// Usage
filter := &PackageEventFilter{
    allowedPackages: []string{"internal/test", "pkg/models"},
}

subscription, err := bus.SubscribeWithFilter(filter, handler)
```

### Event Metrics and Monitoring

```go
func monitorEventBus(bus events.EventBus) {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            metrics := bus.GetMetrics()
            
            log.Printf("Event Bus Metrics:")
            log.Printf("  Total Events: %d", metrics.TotalEvents)
            log.Printf("  Active Subscriptions: %d", metrics.TotalSubscriptions)
            log.Printf("  Events/sec: %.2f", metrics.EventsPerSecond)
            log.Printf("  Avg Processing Time: %v", metrics.AverageProcessingTime)
            log.Printf("  Error Count: %d", metrics.ErrorCount)
            log.Printf("  Last Event: %v", metrics.LastEventTime)
        }
    }
}
```

### Subscription Statistics

```go
func trackSubscriptionStats(subscription events.Subscription) {
    go func() {
        ticker := time.NewTicker(5 * time.Second)
        defer ticker.Stop()
        
        for subscription.IsActive() {
            select {
            case <-ticker.C:
                stats := subscription.GetStats()
                
                log.Printf("Subscription %s Stats:", subscription.ID())
                log.Printf("  Events Received: %d", stats.EventsReceived)
                log.Printf("  Events Processed: %d", stats.EventsProcessed)
                log.Printf("  Processing Errors: %d", stats.ProcessingErrors)
                log.Printf("  Avg Processing Time: %v", stats.AverageProcessingTime)
                log.Printf("  Last Event: %v", stats.LastEventTime)
            }
        }
    }()
}
```

## 🧪 Testing

The package achieves **100% test coverage** with comprehensive test suites:

### Running Tests

```bash
# Run all tests
go test ./pkg/events/...

# Run with coverage
go test ./pkg/events/... -coverprofile=coverage.out

# View coverage report
go tool cover -html=coverage.out
```

### Test Categories

- **Unit Tests**: Individual interface and implementation testing
- **Integration Tests**: Multi-component event workflows
- **Performance Tests**: Event throughput and latency validation
- **Concurrency Tests**: Thread-safety with multiple publishers/subscribers
- **Memory Tests**: Event storage and cleanup efficiency

### Example Test Structure

```go
func TestEventBus_PublishSubscribe(t *testing.T) {
    t.Parallel()
    
    bus := NewTestEventBus()
    handler := &TestEventHandler{}
    
    // Subscribe to events
    subscription, err := bus.Subscribe("test.event", handler)
    assert.NoError(t, err)
    defer subscription.Cancel()
    
    // Publish event
    event := events.NewBaseEvent("test.event", "test", "data")
    ctx := context.Background()
    
    err = bus.Publish(ctx, event)
    assert.NoError(t, err)
    
    // Verify handler received event
    assert.Eventually(t, func() bool {
        return handler.ReceivedCount() > 0
    }, time.Second, 10*time.Millisecond)
}

func TestEventStore_QueryEvents(t *testing.T) {
    t.Parallel()
    
    store := NewTestEventStore()
    ctx := context.Background()
    
    // Store test events
    events := []events.Event{
        events.NewTestStartedEvent("Test1", "pkg1"),
        events.NewTestStartedEvent("Test2", "pkg2"),
        events.NewTestCompletedEvent("Test1", "pkg1", 100*time.Millisecond, true),
    }
    
    for _, event := range events {
        err := store.Store(ctx, event)
        assert.NoError(t, err)
    }
    
    // Query events
    query := &events.EventQuery{
        EventTypes: []string{"test.started"},
        Limit:      10,
    }
    
    retrieved, err := store.Retrieve(ctx, query)
    assert.NoError(t, err)
    assert.Len(t, retrieved, 2)
}
```

## 📊 Performance

The package is optimized for high-performance event processing:

- **Async Publishing**: Non-blocking event publication
- **Efficient Filtering**: Fast event filtering with minimal overhead
- **Batch Processing**: Optimized batch event processing
- **Memory Pooling**: Reusable event objects to reduce allocations

### Benchmarks

```bash
# Run performance benchmarks
go test ./pkg/events/... -bench=.

# Example results:
BenchmarkEventBus_Publish-8           1000000    1.2μs/op    128B/op
BenchmarkEventBus_PublishAsync-8      2000000    0.8μs/op     64B/op
BenchmarkEventHandler_Handle-8        5000000    0.3μs/op     32B/op
BenchmarkEventStore_Store-8            500000    3.2μs/op    256B/op
```

## 🔍 Error Handling

The package provides comprehensive error handling:

### Error Types

```go
// Event publishing errors
type PublishError struct {
    EventType string
    EventID   string
    Cause     error
}

// Subscription errors
type SubscriptionError struct {
    EventType string
    Handler   string
    Cause     error
}

// Storage errors
type StorageError struct {
    Operation string
    EventID   string
    Cause     error
}
```

### Error Handling Examples

```go
// Handle publishing errors
err := bus.Publish(ctx, event)
if err != nil {
    switch e := err.(type) {
    case *events.PublishError:
        log.Printf("Failed to publish %s event %s: %v", e.EventType, e.EventID, e.Cause)
    default:
        log.Printf("Unexpected publish error: %v", err)
    }
}

// Handle subscription errors
subscription, err := bus.Subscribe("test.event", handler)
if err != nil {
    switch e := err.(type) {
    case *events.SubscriptionError:
        log.Printf("Failed to subscribe to %s: %v", e.EventType, e.Cause)
    default:
        log.Printf("Unexpected subscription error: %v", err)
    }
}
```

## 🎯 Best Practices

### Event Design

```go
// Use descriptive event types
const (
    EventTestStarted    = "test.started"
    EventTestCompleted  = "test.completed"
    EventTestFailed     = "test.failed"
    EventFileChanged    = "file.changed"
    EventConfigUpdated  = "config.updated"
)

// Include relevant metadata
event := events.NewBaseEvent(EventTestStarted, "test-runner", testData)
event.Metadata()["package"] = packageName
event.Metadata()["duration"] = expectedDuration
event.Metadata()["tags"] = testTags
```

### Handler Implementation

```go
// Implement idempotent handlers
func (h *TestHandler) Handle(ctx context.Context, event events.Event) error {
    // Check if already processed
    if h.isProcessed(event.ID()) {
        return nil
    }
    
    // Process event
    if err := h.processEvent(ctx, event); err != nil {
        return fmt.Errorf("failed to process event %s: %w", event.ID(), err)
    }
    
    // Mark as processed
    h.markProcessed(event.ID())
    return nil
}

// Use context for cancellation
func (h *TestHandler) processEvent(ctx context.Context, event events.Event) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // Process event
        return nil
    }
}
```

### Subscription Management

```go
// Always cancel subscriptions
subscription, err := bus.Subscribe("test.event", handler)
if err != nil {
    return err
}
defer subscription.Cancel()

// Monitor subscription health
go func() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for subscription.IsActive() {
        select {
        case <-ticker.C:
            stats := subscription.GetStats()
            if stats.ProcessingErrors > 100 {
                log.Printf("High error rate in subscription %s", subscription.ID())
            }
        }
    }
}()
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](../../CONTRIBUTING.md) for details.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/newbpydev/go-sentinel.git

# Navigate to the events package
cd go-sentinel/pkg/events

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

- [`internal/events`](../../internal/events/README.md) - Internal event handling implementation
- [`pkg/models`](../models/README.md) - Shared data models and types
- [`internal/app`](../../internal/app/README.md) - Application orchestration layer

---

**Package Version**: v1.0.0  
**Go Version**: 1.21+  
**Last Updated**: January 2025  
**Maintainer**: Go Sentinel CLI Team 