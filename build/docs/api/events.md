package events // import "github.com/newbpydev/go-sentinel/pkg/events"

Package events provides event system interfaces for inter-component
communication.

This package defines the core event system architecture used throughout the
Go Sentinel CLI for decoupled communication between components. It provides
interfaces for event publishing, subscription, filtering, and persistence.

Key components:
  - EventBus: Central event publishing and subscription management
  - Event: Core event interface with metadata and data
  - EventHandler: Interface for processing events
  - EventStore: Interface for event persistence and retrieval
  - EventProcessor: Interface for batch and async event processing

Example usage:

    // Create and publish events
    event := events.NewTestStartedEvent("TestExample", "github.com/example/pkg")
    bus.Publish(ctx, event)

    // Subscribe to events
    subscription, err := bus.Subscribe("test.started", handler)
    if err != nil {
    	log.Fatal(err)
    }
    defer subscription.Cancel()

    // Query stored events
    query := &events.EventQuery{
    	EventTypes: []string{"test.completed"},
    	StartTime:  &yesterday,
    	Limit:      100,
    }
    events, err := store.Retrieve(ctx, query)

Package events provides event system interfaces for inter-component
communication

CONSTANTS

const (
	// Test execution events
	EventTypeTestStarted   = "test.started"
	EventTypeTestCompleted = "test.completed"
	EventTypeTestFailed    = "test.failed"
	EventTypeTestSkipped   = "test.skipped"

	// Package execution events
	EventTypePackageStarted   = "package.started"
	EventTypePackageCompleted = "package.completed"
	EventTypePackageFailed    = "package.failed"

	// Watch mode events
	EventTypeFileChanged     = "file.changed"
	EventTypeWatchStarted    = "watch.started"
	EventTypeWatchStopped    = "watch.stopped"
	EventTypeWatchModeChange = "watch.mode_changed"

	// Application events
	EventTypeAppStarted  = "app.started"
	EventTypeAppStopped  = "app.stopped"
	EventTypeAppError    = "app.error"
	EventTypeAppShutdown = "app.shutdown"

	// Configuration events
	EventTypeConfigLoaded  = "config.loaded"
	EventTypeConfigChanged = "config.changed"
	EventTypeConfigError   = "config.error"

	// Cache events
	EventTypeCacheHit   = "cache.hit"
	EventTypeCacheMiss  = "cache.miss"
	EventTypeCacheStore = "cache.store"
	EventTypeCacheClear = "cache.clear"

	// UI events
	EventTypeDisplayUpdate  = "display.update"
	EventTypeProgressUpdate = "progress.update"
	EventTypeThemeChanged   = "theme.changed"
	EventTypeIconSetChanged = "iconset.changed"
)
    Common event type constants

const (
	// Component sources
	SourceTestRunner      = "test.runner"
	SourceTestProcessor   = "test.processor"
	SourceFileWatcher     = "file.watcher"
	SourceAppController   = "app.controller"
	SourceCacheManager    = "cache.manager"
	SourceDisplayRenderer = "display.renderer"

	// System sources
	SourceFileSystem      = "filesystem"
	SourceOperatingSystem = "os"
	SourceUserInput       = "user"
	SourceConfig          = "config"
)
    Event source constants

const (
	PriorityLow      = 1
	PriorityNormal   = 5
	PriorityHigh     = 10
	PriorityCritical = 15
)
    Priority levels for event handlers


FUNCTIONS

func Example_baseEvent()
    Example_baseEvent demonstrates working with the base event implementation.

    This example shows how to create custom events using the BaseEvent struct
    and how to access event properties.

func Example_eventBusUsage()
    Example_eventBusUsage demonstrates basic event bus operations.

    This example shows how to publish events, subscribe to them, and manage
    subscriptions in a typical application scenario.

func Example_eventConstants()
    Example_eventConstants demonstrates the predefined event constants.

    This example shows the available event types, sources, and priorities that
    are commonly used throughout the application.

func Example_eventMetrics()
    Example_eventMetrics demonstrates working with event system metrics.

    This example shows how to monitor event bus performance and subscription
    statistics for system observability.

func Example_eventQuery()
    Example_eventQuery demonstrates querying stored events.

    This example shows how to construct queries for retrieving events from an
    event store with various filters.

func Example_fileChangeEvents()
    Example_fileChangeEvents demonstrates working with file change events.

    This example shows how to create and handle file system change events that
    are commonly used in watch mode functionality.


TYPES

type BaseEvent struct {
	// EventID is the unique event identifier
	EventID string

	// EventType is the type of event
	EventType string

	// EventTimestamp is when the event occurred
	EventTimestamp time.Time

	// EventSource is the source of the event
	EventSource string

	// EventData is the event payload
	EventData interface{}

	// EventMetadata contains additional event information
	EventMetadata map[string]interface{}
}
    BaseEvent provides a basic event implementation

func NewBaseEvent(eventType, source string, data interface{}) *BaseEvent
    NewBaseEvent creates a new BaseEvent

func (e *BaseEvent) Data() interface{}
    Data implements the Event interface

func (e *BaseEvent) ID() string
    ID implements the Event interface

func (e *BaseEvent) Metadata() map[string]interface{}
    Metadata implements the Event interface

func (e *BaseEvent) Source() string
    Source implements the Event interface

func (e *BaseEvent) String() string
    String implements the Event interface

func (e *BaseEvent) Timestamp() time.Time
    Timestamp implements the Event interface

func (e *BaseEvent) Type() string
    Type implements the Event interface

type Event interface {
	// ID returns the event ID
	ID() string

	// Type returns the event type
	Type() string

	// Timestamp returns when the event occurred
	Timestamp() time.Time

	// Source returns the event source
	Source() string

	// Data returns the event data
	Data() interface{}

	// Metadata returns event metadata
	Metadata() map[string]interface{}

	// String returns a string representation of the event
	String() string
}
    Event represents a system event

type EventBus interface {
	// Publish publishes an event to all subscribers
	Publish(ctx context.Context, event Event) error

	// PublishAsync publishes an event asynchronously
	PublishAsync(ctx context.Context, event Event) error

	// Subscribe subscribes to events of a specific type
	Subscribe(eventType string, handler EventHandler) (Subscription, error)

	// SubscribeWithFilter subscribes with a custom filter
	SubscribeWithFilter(filter EventFilter, handler EventHandler) (Subscription, error)

	// Unsubscribe removes a subscription
	Unsubscribe(subscription Subscription) error

	// Close closes the event bus and cleans up resources
	Close() error

	// GetMetrics returns event bus metrics
	GetMetrics() *EventBusMetrics
}
    EventBus manages event publishing and subscription

type EventBusMetrics struct {
	// TotalEvents is the total number of events published
	TotalEvents int64

	// TotalSubscriptions is the total number of active subscriptions
	TotalSubscriptions int

	// EventsPerSecond is the current events per second rate
	EventsPerSecond float64

	// AverageProcessingTime is the average event processing time
	AverageProcessingTime time.Duration

	// ErrorCount is the number of processing errors
	ErrorCount int64

	// LastEventTime is when the last event was published
	LastEventTime time.Time
}
    EventBusMetrics provides metrics about event bus usage

type EventFilter interface {
	// Match returns whether the event matches the filter
	Match(event Event) bool

	// String returns a string representation of the filter
	String() string
}
    EventFilter filters events based on criteria

type EventHandler interface {
	// Handle processes an event
	Handle(ctx context.Context, event Event) error

	// CanHandle returns whether this handler can process the event
	CanHandle(event Event) bool

	// Priority returns the handler priority (higher values = higher priority)
	Priority() int
}
    EventHandler processes events

type EventProcessor interface {
	// ProcessSync processes events synchronously
	ProcessSync(ctx context.Context, events []Event) error

	// ProcessAsync processes events asynchronously
	ProcessAsync(ctx context.Context, events []Event) error

	// ProcessBatch processes events in batches
	ProcessBatch(ctx context.Context, events []Event, batchSize int) error

	// GetProcessingStats returns processing statistics
	GetProcessingStats() *ProcessingStats
}
    EventProcessor processes events with various strategies

type EventQuery struct {
	// EventTypes filters by event types
	EventTypes []string

	// Sources filters by event sources
	Sources []string

	// StartTime filters events after this time
	StartTime *time.Time

	// EndTime filters events before this time
	EndTime *time.Time

	// Limit limits the number of results
	Limit int

	// Offset skips this number of results
	Offset int

	// OrderBy specifies the ordering field
	OrderBy string

	// OrderDesc specifies descending order
	OrderDesc bool

	// Metadata filters by metadata values
	Metadata map[string]interface{}
}
    EventQuery represents criteria for querying events

type EventStore interface {
	// Store persists an event
	Store(ctx context.Context, event Event) error

	// Retrieve retrieves events by criteria
	Retrieve(ctx context.Context, query *EventQuery) ([]Event, error)

	// Count returns the number of events matching criteria
	Count(ctx context.Context, query *EventQuery) (int, error)

	// Delete removes events by criteria
	Delete(ctx context.Context, query *EventQuery) error

	// Clear removes all events
	Clear(ctx context.Context) error
}
    EventStore provides event persistence and retrieval

type FileChangedEvent struct {
	*BaseEvent
	FilePath   string
	ChangeType string
}
    FileChangedEvent represents a file change

func NewFileChangedEvent(filePath, changeType string) *FileChangedEvent
    NewFileChangedEvent creates a new file changed event

type ProcessingStats struct {
	// TotalProcessed is the total number of events processed
	TotalProcessed int64

	// TotalErrors is the total number of processing errors
	TotalErrors int64

	// AverageProcessingTime is the average processing time per event
	AverageProcessingTime time.Duration

	// ProcessingRate is the current processing rate (events/second)
	ProcessingRate float64

	// QueueSize is the current queue size
	QueueSize int

	// MaxQueueSize is the maximum queue size reached
	MaxQueueSize int
}
    ProcessingStats provides statistics about event processing

type Subscription interface {
	// ID returns the subscription ID
	ID() string

	// EventType returns the subscribed event type
	EventType() string

	// IsActive returns whether the subscription is active
	IsActive() bool

	// Cancel cancels the subscription
	Cancel() error

	// GetStats returns subscription statistics
	GetStats() *SubscriptionStats
}
    Subscription represents an active event subscription

type SubscriptionStats struct {
	// EventsReceived is the number of events received
	EventsReceived int64

	// EventsProcessed is the number of events successfully processed
	EventsProcessed int64

	// ProcessingErrors is the number of processing errors
	ProcessingErrors int64

	// AverageProcessingTime is the average processing time
	AverageProcessingTime time.Duration

	// LastEventTime is when the last event was received
	LastEventTime time.Time

	// CreatedAt is when the subscription was created
	CreatedAt time.Time
}
    SubscriptionStats provides statistics about a subscription

type TestCompletedEvent struct {
	*BaseEvent
	TestName    string
	PackageName string
	Duration    time.Duration
	Success     bool
}
    TestCompletedEvent represents a test completion

func NewTestCompletedEvent(testName, packageName string, duration time.Duration, success bool) *TestCompletedEvent
    NewTestCompletedEvent creates a new test completed event

type TestStartedEvent struct {
	*BaseEvent
	TestName    string
	PackageName string
}
    TestStartedEvent represents a test starting

func NewTestStartedEvent(testName, packageName string) *TestStartedEvent
    NewTestStartedEvent creates a new test started event

