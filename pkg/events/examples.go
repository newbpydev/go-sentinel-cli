// Package events provides event system interfaces for inter-component communication.
//
// This package defines the core event system architecture used throughout the Go Sentinel CLI
// for decoupled communication between components. It provides interfaces for event publishing,
// subscription, filtering, and persistence.
//
// Key components:
//   - EventBus: Central event publishing and subscription management
//   - Event: Core event interface with metadata and data
//   - EventHandler: Interface for processing events
//   - EventStore: Interface for event persistence and retrieval
//   - EventProcessor: Interface for batch and async event processing
//
// Example usage:
//
//	// Create and publish events
//	event := events.NewTestStartedEvent("TestExample", "github.com/example/pkg")
//	bus.Publish(ctx, event)
//
//	// Subscribe to events
//	subscription, err := bus.Subscribe("test.started", handler)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer subscription.Cancel()
//
//	// Query stored events
//	query := &events.EventQuery{
//		EventTypes: []string{"test.completed"},
//		StartTime:  &yesterday,
//		Limit:      100,
//	}
//	events, err := store.Retrieve(ctx, query)
package events

import (
	"fmt"
	"time"
)

// Example_eventBusUsage demonstrates basic event bus operations.
//
// This example shows how to publish events, subscribe to them,
// and manage subscriptions in a typical application scenario.
func Example_eventBusUsage() {
	// This is a conceptual example - actual implementation would require
	// a concrete EventBus implementation

	// Create test events
	testStarted := NewTestStartedEvent("TestUserAuth", "github.com/example/auth")
	testCompleted := NewTestCompletedEvent("TestUserAuth", "github.com/example/auth",
		150*time.Millisecond, true)

	fmt.Printf("Test Started Event: %s\n", testStarted.String())
	fmt.Printf("- Type: %s\n", testStarted.Type())
	fmt.Printf("- Source: %s\n", testStarted.Source())
	fmt.Printf("- Timestamp: %s\n", testStarted.Timestamp().Format("15:04:05"))

	fmt.Printf("\nTest Completed Event: %s\n", testCompleted.String())
	fmt.Printf("- Duration: %v\n", testCompleted.Duration)
	fmt.Printf("- Success: %t\n", testCompleted.Success)

	// Output:
	// Test Started Event: test.started:20240101150405.000000
	// - Type: test.started
	// - Source: test.runner
	// - Timestamp: 15:04:05
	//
	// Test Completed Event: test.completed:20240101150405.000000
	// - Duration: 150ms
	// - Success: true
}

// Example_eventQuery demonstrates querying stored events.
//
// This example shows how to construct queries for retrieving
// events from an event store with various filters.
func Example_eventQuery() {
	// Create a query for test events from the last hour
	lastHour := time.Now().Add(-1 * time.Hour)

	query := &EventQuery{
		EventTypes: []string{"test.started", "test.completed"},
		Sources:    []string{"test.runner"},
		StartTime:  &lastHour,
		Limit:      50,
		OrderBy:    "timestamp",
		OrderDesc:  true,
		Metadata: map[string]interface{}{
			"package": "github.com/example/auth",
		},
	}

	fmt.Printf("Event Query Configuration:\n")
	fmt.Printf("- Event Types: %v\n", query.EventTypes)
	fmt.Printf("- Sources: %v\n", query.Sources)
	fmt.Printf("- Start Time: %s\n", query.StartTime.Format("15:04:05"))
	fmt.Printf("- Limit: %d\n", query.Limit)
	fmt.Printf("- Order: %s (%s)\n", query.OrderBy,
		map[bool]string{true: "DESC", false: "ASC"}[query.OrderDesc])

	// Output:
	// Event Query Configuration:
	// - Event Types: [test.started test.completed]
	// - Sources: [test.runner]
	// - Start Time: 14:04:05
	// - Limit: 50
	// - Order: timestamp (DESC)
}

// Example_eventMetrics demonstrates working with event system metrics.
//
// This example shows how to monitor event bus performance and
// subscription statistics for system observability.
func Example_eventMetrics() {
	// Create sample metrics (in real usage, these would come from the event bus)
	busMetrics := &EventBusMetrics{
		TotalEvents:           1250,
		TotalSubscriptions:    8,
		EventsPerSecond:       15.7,
		AverageProcessingTime: 25 * time.Millisecond,
		ErrorCount:            3,
		LastEventTime:         time.Now(),
	}

	subscriptionStats := &SubscriptionStats{
		EventsReceived:        450,
		EventsProcessed:       447,
		ProcessingErrors:      3,
		AverageProcessingTime: 12 * time.Millisecond,
		LastEventTime:         time.Now(),
		CreatedAt:             time.Now().Add(-2 * time.Hour),
	}

	processingStats := &ProcessingStats{
		TotalProcessed:        1247,
		TotalErrors:           3,
		AverageProcessingTime: 18 * time.Millisecond,
		ProcessingRate:        69.3,
		QueueSize:             12,
		MaxQueueSize:          45,
	}

	// Display metrics
	fmt.Printf("Event Bus Metrics:\n")
	fmt.Printf("- Total Events: %d\n", busMetrics.TotalEvents)
	fmt.Printf("- Active Subscriptions: %d\n", busMetrics.TotalSubscriptions)
	fmt.Printf("- Events/Second: %.1f\n", busMetrics.EventsPerSecond)
	fmt.Printf("- Avg Processing Time: %v\n", busMetrics.AverageProcessingTime)
	fmt.Printf("- Error Count: %d\n", busMetrics.ErrorCount)

	fmt.Printf("\nSubscription Statistics:\n")
	fmt.Printf("- Events Received: %d\n", subscriptionStats.EventsReceived)
	fmt.Printf("- Events Processed: %d\n", subscriptionStats.EventsProcessed)
	fmt.Printf("- Processing Errors: %d\n", subscriptionStats.ProcessingErrors)
	fmt.Printf("- Success Rate: %.1f%%\n",
		float64(subscriptionStats.EventsProcessed)/float64(subscriptionStats.EventsReceived)*100)

	fmt.Printf("\nProcessing Statistics:\n")
	fmt.Printf("- Total Processed: %d\n", processingStats.TotalProcessed)
	fmt.Printf("- Total Errors: %d\n", processingStats.TotalErrors)
	fmt.Printf("- Processing Rate: %.1f events/sec\n", processingStats.ProcessingRate)
	fmt.Printf("- Current Queue Size: %d\n", processingStats.QueueSize)
	fmt.Printf("- Max Queue Size: %d\n", processingStats.MaxQueueSize)

	// Output:
	// Event Bus Metrics:
	// - Total Events: 1250
	// - Active Subscriptions: 8
	// - Events/Second: 15.7
	// - Avg Processing Time: 25ms
	// - Error Count: 3
	//
	// Subscription Statistics:
	// - Events Received: 450
	// - Events Processed: 447
	// - Processing Errors: 3
	// - Success Rate: 99.3%
	//
	// Processing Statistics:
	// - Total Processed: 1247
	// - Total Errors: 3
	// - Processing Rate: 69.3 events/sec
	// - Current Queue Size: 12
	// - Max Queue Size: 45
}

// Example_fileChangeEvents demonstrates working with file change events.
//
// This example shows how to create and handle file system change events
// that are commonly used in watch mode functionality.
func Example_fileChangeEvents() {
	// Create different types of file change events
	fileCreated := NewFileChangedEvent("/src/new_test.go", "created")
	fileModified := NewFileChangedEvent("/src/existing_test.go", "modified")
	fileDeleted := NewFileChangedEvent("/src/old_test.go", "deleted")

	events := []*FileChangedEvent{fileCreated, fileModified, fileDeleted}

	fmt.Printf("File Change Events:\n")
	for i, event := range events {
		fmt.Printf("%d. %s\n", i+1, event.String())
		fmt.Printf("   - File: %s\n", event.FilePath)
		fmt.Printf("   - Change: %s\n", event.ChangeType)
		fmt.Printf("   - Time: %s\n", event.Timestamp().Format("15:04:05"))

		// Access event metadata
		if metadata := event.Metadata(); len(metadata) > 0 {
			fmt.Printf("   - Metadata: %v\n", metadata)
		}
	}

	// Output:
	// File Change Events:
	// 1. file.changed:20240101150405.000000
	//    - File: /src/new_test.go
	//    - Change: created
	//    - Time: 15:04:05
	// 2. file.changed:20240101150405.000000
	//    - File: /src/existing_test.go
	//    - Change: modified
	//    - Time: 15:04:05
	// 3. file.changed:20240101150405.000000
	//    - File: /src/old_test.go
	//    - Change: deleted
	//    - Time: 15:04:05
}

// Example_baseEvent demonstrates working with the base event implementation.
//
// This example shows how to create custom events using the BaseEvent
// struct and how to access event properties.
func Example_baseEvent() {
	// Create a custom event using BaseEvent
	customData := map[string]interface{}{
		"severity":  "high",
		"component": "test-runner",
		"details":   "Memory usage exceeded threshold",
	}

	event := NewBaseEvent("system.alert", "monitoring", customData)

	// Add metadata
	event.EventMetadata = map[string]interface{}{
		"host": "test-server-01",
		"pid":  12345,
		"user": "ci-runner",
	}

	fmt.Printf("Custom Event Details:\n")
	fmt.Printf("- ID: %s\n", event.ID())
	fmt.Printf("- Type: %s\n", event.Type())
	fmt.Printf("- Source: %s\n", event.Source())
	fmt.Printf("- Timestamp: %s\n", event.Timestamp().Format("2006-01-02 15:04:05"))
	fmt.Printf("- String: %s\n", event.String())

	// Access event data
	if data := event.Data(); data != nil {
		if dataMap, ok := data.(map[string]interface{}); ok {
			fmt.Printf("- Data:\n")
			for key, value := range dataMap {
				fmt.Printf("  - %s: %v\n", key, value)
			}
		}
	}

	// Access metadata
	if metadata := event.Metadata(); len(metadata) > 0 {
		fmt.Printf("- Metadata:\n")
		for key, value := range metadata {
			fmt.Printf("  - %s: %v\n", key, value)
		}
	}

	// Output:
	// Custom Event Details:
	// - ID: 20240101150405.000000
	// - Type: system.alert
	// - Source: monitoring
	// - Timestamp: 2024-01-01 15:04:05
	// - String: system.alert:20240101150405.000000
	// - Data:
	//   - severity: high
	//   - component: test-runner
	//   - details: Memory usage exceeded threshold
	// - Metadata:
	//   - host: test-server-01
	//   - pid: 12345
	//   - user: ci-runner
}

// Example_eventConstants demonstrates the predefined event constants.
//
// This example shows the available event types, sources, and priorities
// that are commonly used throughout the application.
func Example_eventConstants() {
	// Event types
	eventTypes := []string{
		EventTypeTestStarted,
		EventTypeTestCompleted,
		EventTypeFileChanged,
		EventTypeConfigChanged,
		EventTypeAppStarted,
	}

	// Event sources
	eventSources := []string{
		SourceTestRunner,
		SourceFileWatcher,
		SourceAppController,
		SourceConfig,
	}

	// Event priorities
	priorities := []struct {
		name  string
		value int
	}{
		{"Low", PriorityLow},
		{"Normal", PriorityNormal},
		{"High", PriorityHigh},
		{"Critical", PriorityCritical},
	}

	fmt.Printf("Available Event Types:\n")
	for i, eventType := range eventTypes {
		fmt.Printf("  %d. %s\n", i+1, eventType)
	}

	fmt.Printf("\nAvailable Event Sources:\n")
	for i, source := range eventSources {
		fmt.Printf("  %d. %s\n", i+1, source)
	}

	fmt.Printf("\nEvent Priorities:\n")
	for _, priority := range priorities {
		fmt.Printf("  - %s: %d\n", priority.name, priority.value)
	}

	// Output:
	// Available Event Types:
	//   1. test.started
	//   2. test.completed
	//   3. file.changed
	//   4. config.changed
	//   5. app.started
	//
	// Available Event Sources:
	//   1. test.runner
	//   2. file.watcher
	//   3. app.controller
	//   4. config
	//
	// Event Priorities:
	//   - Low: 1
	//   - Normal: 5
	//   - High: 10
	//   - Critical: 15
}
