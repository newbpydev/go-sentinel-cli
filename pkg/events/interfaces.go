// Package events provides event system interfaces for inter-component communication
package events

import (
	"context"
	"time"
)

// EventBus manages event publishing and subscription
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

// EventHandler processes events
type EventHandler interface {
	// Handle processes an event
	Handle(ctx context.Context, event Event) error

	// CanHandle returns whether this handler can process the event
	CanHandle(event Event) bool

	// Priority returns the handler priority (higher values = higher priority)
	Priority() int
}

// EventFilter filters events based on criteria
type EventFilter interface {
	// Match returns whether the event matches the filter
	Match(event Event) bool

	// String returns a string representation of the filter
	String() string
}

// Subscription represents an active event subscription
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

// EventStore provides event persistence and retrieval
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

// EventProcessor processes events with various strategies
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

// Event represents a system event
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

// BaseEvent provides a basic event implementation
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

// ID implements the Event interface
func (e *BaseEvent) ID() string {
	return e.EventID
}

// Type implements the Event interface
func (e *BaseEvent) Type() string {
	return e.EventType
}

// Timestamp implements the Event interface
func (e *BaseEvent) Timestamp() time.Time {
	return e.EventTimestamp
}

// Source implements the Event interface
func (e *BaseEvent) Source() string {
	return e.EventSource
}

// Data implements the Event interface
func (e *BaseEvent) Data() interface{} {
	return e.EventData
}

// Metadata implements the Event interface
func (e *BaseEvent) Metadata() map[string]interface{} {
	return e.EventMetadata
}

// String implements the Event interface
func (e *BaseEvent) String() string {
	return e.EventType + ":" + e.EventID
}

// EventQuery represents criteria for querying events
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

// EventBusMetrics provides metrics about event bus usage
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

// SubscriptionStats provides statistics about a subscription
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

// ProcessingStats provides statistics about event processing
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

// Common event type constants
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

// Event source constants
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

// Priority levels for event handlers
const (
	PriorityLow      = 1
	PriorityNormal   = 5
	PriorityHigh     = 10
	PriorityCritical = 15
)

// NewBaseEvent creates a new BaseEvent
func NewBaseEvent(eventType, source string, data interface{}) *BaseEvent {
	return &BaseEvent{
		EventID:        generateEventID(),
		EventType:      eventType,
		EventTimestamp: time.Now(),
		EventSource:    source,
		EventData:      data,
		EventMetadata:  make(map[string]interface{}),
	}
}

// generateEventID generates a unique event ID
func generateEventID() string {
	// Simple timestamp-based ID for now
	return time.Now().Format("20060102150405.000000")
}

// TestStartedEvent represents a test starting
type TestStartedEvent struct {
	*BaseEvent
	TestName    string
	PackageName string
}

// TestCompletedEvent represents a test completion
type TestCompletedEvent struct {
	*BaseEvent
	TestName    string
	PackageName string
	Duration    time.Duration
	Success     bool
}

// FileChangedEvent represents a file change
type FileChangedEvent struct {
	*BaseEvent
	FilePath   string
	ChangeType string
}

// NewTestStartedEvent creates a new test started event
func NewTestStartedEvent(testName, packageName string) *TestStartedEvent {
	return &TestStartedEvent{
		BaseEvent:   NewBaseEvent(EventTypeTestStarted, SourceTestRunner, nil),
		TestName:    testName,
		PackageName: packageName,
	}
}

// NewTestCompletedEvent creates a new test completed event
func NewTestCompletedEvent(testName, packageName string, duration time.Duration, success bool) *TestCompletedEvent {
	return &TestCompletedEvent{
		BaseEvent:   NewBaseEvent(EventTypeTestCompleted, SourceTestRunner, nil),
		TestName:    testName,
		PackageName: packageName,
		Duration:    duration,
		Success:     success,
	}
}

// NewFileChangedEvent creates a new file changed event
func NewFileChangedEvent(filePath, changeType string) *FileChangedEvent {
	return &FileChangedEvent{
		BaseEvent:  NewBaseEvent(EventTypeFileChanged, SourceFileWatcher, nil),
		FilePath:   filePath,
		ChangeType: changeType,
	}
}
