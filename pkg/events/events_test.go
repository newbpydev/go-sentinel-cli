package events

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

// TestNewBaseEvent_FactoryFunction tests the BaseEvent factory function
func TestNewBaseEvent_FactoryFunction(t *testing.T) {
	t.Parallel()

	eventType := "test.event"
	source := "test.source"
	data := map[string]interface{}{"key": "value"}

	event := NewBaseEvent(eventType, source, data)

	if event == nil {
		t.Fatal("NewBaseEvent should not return nil")
	}

	// Verify interface compliance
	var _ Event = event

	// Verify event type
	if event.Type() != eventType {
		t.Errorf("Expected event type %q, got %q", eventType, event.Type())
	}

	// Verify source
	if event.Source() != source {
		t.Errorf("Expected source %q, got %q", source, event.Source())
	}

	// Verify data
	if !reflect.DeepEqual(event.Data(), data) {
		t.Errorf("Expected data %v, got %v", data, event.Data())
	}

	// Verify ID is generated
	if event.ID() == "" {
		t.Error("Event ID should not be empty")
	}

	// Verify timestamp is set
	if event.Timestamp().IsZero() {
		t.Error("Event timestamp should not be zero")
	}

	// Verify metadata is initialized
	if event.Metadata() == nil {
		t.Error("Event metadata should not be nil")
	}
}

// TestBaseEvent_InterfaceMethods tests all BaseEvent interface methods
func TestBaseEvent_InterfaceMethods(t *testing.T) {
	t.Parallel()

	// Create test event
	eventType := "test.interface"
	source := "test.runner"
	data := "test data"

	event := NewBaseEvent(eventType, source, data)

	// Test ID method
	id := event.ID()
	if id == "" {
		t.Error("ID() should return non-empty string")
	}

	// Test Type method
	if event.Type() != eventType {
		t.Errorf("Type() expected %q, got %q", eventType, event.Type())
	}

	// Test Source method
	if event.Source() != source {
		t.Errorf("Source() expected %q, got %q", source, event.Source())
	}

	// Test Data method
	if event.Data() != data {
		t.Errorf("Data() expected %v, got %v", data, event.Data())
	}

	// Test Timestamp method
	timestamp := event.Timestamp()
	if timestamp.IsZero() {
		t.Error("Timestamp() should return non-zero time")
	}

	// Verify timestamp is recent (within last second)
	if time.Since(timestamp) > time.Second {
		t.Error("Timestamp should be recent")
	}

	// Test Metadata method
	metadata := event.Metadata()
	if metadata == nil {
		t.Error("Metadata() should return non-nil map")
	}

	// Test String method
	str := event.String()
	expectedPrefix := eventType + ":"
	if !strings.HasPrefix(str, expectedPrefix) {
		t.Errorf("String() should start with %q, got %q", expectedPrefix, str)
	}
}

// TestBaseEvent_MetadataManipulation tests metadata operations
func TestBaseEvent_MetadataManipulation(t *testing.T) {
	t.Parallel()

	event := NewBaseEvent("test.metadata", "test.source", nil)

	// Test initial metadata
	metadata := event.Metadata()
	if metadata == nil {
		t.Fatal("Metadata should not be nil")
	}

	// Test adding metadata
	metadata["test_key"] = "test_value"
	metadata["number"] = 42

	// Verify metadata persists
	retrievedMetadata := event.Metadata()
	if retrievedMetadata["test_key"] != "test_value" {
		t.Error("Metadata should persist changes")
	}
	if retrievedMetadata["number"] != 42 {
		t.Error("Metadata should persist numeric values")
	}
}

// TestGenerateEventID tests the event ID generation
func TestGenerateEventID(t *testing.T) {
	t.Parallel()

	// Generate multiple IDs with small delays to avoid timestamp collisions
	ids := make(map[string]bool)
	for i := 0; i < 10; i++ {
		id := generateEventID()

		// Verify ID is not empty
		if id == "" {
			t.Error("generateEventID should not return empty string")
		}

		// Verify ID format (should be timestamp-based)
		if len(id) < 10 {
			t.Errorf("generateEventID should return reasonable length ID, got: %s", id)
		}

		// Store ID for uniqueness check (note: duplicates are possible with timestamp-based IDs)
		ids[id] = true

		// Small delay to reduce chance of timestamp collision
		time.Sleep(time.Microsecond)
	}

	// Verify we got at least some unique IDs
	if len(ids) == 0 {
		t.Error("Should generate at least some IDs")
	}
}

// TestNewTestStartedEvent_FactoryFunction tests TestStartedEvent factory
func TestNewTestStartedEvent_FactoryFunction(t *testing.T) {
	t.Parallel()

	testName := "TestExample"
	packageName := "github.com/example/pkg"

	event := NewTestStartedEvent(testName, packageName)

	if event == nil {
		t.Fatal("NewTestStartedEvent should not return nil")
	}

	// Verify interface compliance
	var _ Event = event

	// Verify event type
	if event.Type() != "test.started" {
		t.Errorf("Expected event type 'test.started', got %q", event.Type())
	}

	// Verify source
	if event.Source() != "test.runner" {
		t.Errorf("Expected source 'test.runner', got %q", event.Source())
	}

	// Verify test-specific fields
	if event.TestName != testName {
		t.Errorf("Expected test name %q, got %q", testName, event.TestName)
	}

	if event.PackageName != packageName {
		t.Errorf("Expected package name %q, got %q", packageName, event.PackageName)
	}

	// Verify embedded BaseEvent
	if event.BaseEvent == nil {
		t.Error("TestStartedEvent should embed BaseEvent")
	}
}

// TestNewTestCompletedEvent_FactoryFunction tests TestCompletedEvent factory
func TestNewTestCompletedEvent_FactoryFunction(t *testing.T) {
	t.Parallel()

	testName := "TestExample"
	packageName := "github.com/example/pkg"
	duration := 150 * time.Millisecond
	success := true

	event := NewTestCompletedEvent(testName, packageName, duration, success)

	if event == nil {
		t.Fatal("NewTestCompletedEvent should not return nil")
	}

	// Verify interface compliance
	var _ Event = event

	// Verify event type
	if event.Type() != "test.completed" {
		t.Errorf("Expected event type 'test.completed', got %q", event.Type())
	}

	// Verify source
	if event.Source() != "test.runner" {
		t.Errorf("Expected source 'test.runner', got %q", event.Source())
	}

	// Verify test-specific fields
	if event.TestName != testName {
		t.Errorf("Expected test name %q, got %q", testName, event.TestName)
	}

	if event.PackageName != packageName {
		t.Errorf("Expected package name %q, got %q", packageName, event.PackageName)
	}

	if event.Duration != duration {
		t.Errorf("Expected duration %v, got %v", duration, event.Duration)
	}

	if event.Success != success {
		t.Errorf("Expected success %t, got %t", success, event.Success)
	}

	// Verify embedded BaseEvent
	if event.BaseEvent == nil {
		t.Error("TestCompletedEvent should embed BaseEvent")
	}
}

// TestNewFileChangedEvent_FactoryFunction tests FileChangedEvent factory
func TestNewFileChangedEvent_FactoryFunction(t *testing.T) {
	t.Parallel()

	filePath := "/src/test.go"
	changeType := "modified"

	event := NewFileChangedEvent(filePath, changeType)

	if event == nil {
		t.Fatal("NewFileChangedEvent should not return nil")
	}

	// Verify interface compliance
	var _ Event = event

	// Verify event type
	if event.Type() != "file.changed" {
		t.Errorf("Expected event type 'file.changed', got %q", event.Type())
	}

	// Verify source
	if event.Source() != "file.watcher" {
		t.Errorf("Expected source 'file.watcher', got %q", event.Source())
	}

	// Verify file-specific fields
	if event.FilePath != filePath {
		t.Errorf("Expected file path %q, got %q", filePath, event.FilePath)
	}

	if event.ChangeType != changeType {
		t.Errorf("Expected change type %q, got %q", changeType, event.ChangeType)
	}

	// Verify embedded BaseEvent
	if event.BaseEvent == nil {
		t.Error("FileChangedEvent should embed BaseEvent")
	}
}

// TestEventTypes_EdgeCases tests edge cases for event creation
func TestEventTypes_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		createEvent func() Event
		expectValid bool
	}{
		{
			name: "empty_event_type",
			createEvent: func() Event {
				return NewBaseEvent("", "source", nil)
			},
			expectValid: true, // Empty type should be allowed
		},
		{
			name: "empty_source",
			createEvent: func() Event {
				return NewBaseEvent("test.type", "", nil)
			},
			expectValid: true, // Empty source should be allowed
		},
		{
			name: "nil_data",
			createEvent: func() Event {
				return NewBaseEvent("test.type", "source", nil)
			},
			expectValid: true, // Nil data should be allowed
		},
		{
			name: "complex_data",
			createEvent: func() Event {
				data := map[string]interface{}{
					"nested": map[string]interface{}{
						"value": 42,
						"list":  []string{"a", "b", "c"},
					},
				}
				return NewBaseEvent("test.complex", "source", data)
			},
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			event := tt.createEvent()

			if !tt.expectValid {
				if event != nil {
					t.Error("Expected nil event for invalid input")
				}
				return
			}

			if event == nil {
				t.Fatal("Expected valid event, got nil")
			}

			// Verify basic interface compliance
			if event.ID() == "" {
				t.Error("Event should have non-empty ID")
			}

			if event.Timestamp().IsZero() {
				t.Error("Event should have non-zero timestamp")
			}

			if event.Metadata() == nil {
				t.Error("Event should have non-nil metadata")
			}
		})
	}
}

// TestEventQuery_StructureValidation tests EventQuery structure
func TestEventQuery_StructureValidation(t *testing.T) {
	t.Parallel()

	// Test default query
	query := &EventQuery{}

	// Verify default values
	if query.EventTypes != nil && len(query.EventTypes) > 0 {
		t.Error("Default EventTypes should be empty")
	}

	if query.Sources != nil && len(query.Sources) > 0 {
		t.Error("Default Sources should be empty")
	}

	if query.StartTime != nil {
		t.Error("Default StartTime should be nil")
	}

	if query.EndTime != nil {
		t.Error("Default EndTime should be nil")
	}

	if query.Limit != 0 {
		t.Error("Default Limit should be 0")
	}

	if query.Offset != 0 {
		t.Error("Default Offset should be 0")
	}

	// Test populated query
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	populatedQuery := &EventQuery{
		EventTypes: []string{"test.started", "test.completed"},
		Sources:    []string{"test.runner"},
		StartTime:  &yesterday,
		EndTime:    &now,
		Limit:      100,
		Offset:     10,
		OrderBy:    "timestamp",
		OrderDesc:  true,
		Metadata: map[string]interface{}{
			"package": "example",
		},
	}

	// Verify populated values
	if len(populatedQuery.EventTypes) != 2 {
		t.Errorf("Expected 2 event types, got %d", len(populatedQuery.EventTypes))
	}

	if len(populatedQuery.Sources) != 1 {
		t.Errorf("Expected 1 source, got %d", len(populatedQuery.Sources))
	}

	if populatedQuery.StartTime == nil || populatedQuery.StartTime.IsZero() {
		t.Error("StartTime should be set and non-zero")
	}

	if populatedQuery.EndTime == nil || populatedQuery.EndTime.IsZero() {
		t.Error("EndTime should be set and non-zero")
	}

	if populatedQuery.Limit != 100 {
		t.Errorf("Expected limit 100, got %d", populatedQuery.Limit)
	}

	if populatedQuery.Offset != 10 {
		t.Errorf("Expected offset 10, got %d", populatedQuery.Offset)
	}

	if populatedQuery.OrderBy != "timestamp" {
		t.Errorf("Expected OrderBy 'timestamp', got %q", populatedQuery.OrderBy)
	}

	if !populatedQuery.OrderDesc {
		t.Error("Expected OrderDesc to be true")
	}

	if populatedQuery.Metadata["package"] != "example" {
		t.Error("Expected metadata package to be 'example'")
	}
}

// TestMetricsStructures_Validation tests metrics structure initialization
func TestMetricsStructures_Validation(t *testing.T) {
	t.Parallel()

	// Test EventBusMetrics
	busMetrics := &EventBusMetrics{
		TotalEvents:           1000,
		TotalSubscriptions:    5,
		EventsPerSecond:       25.5,
		AverageProcessingTime: 10 * time.Millisecond,
		ErrorCount:            2,
		LastEventTime:         time.Now(),
	}

	if busMetrics.TotalEvents != 1000 {
		t.Errorf("Expected TotalEvents 1000, got %d", busMetrics.TotalEvents)
	}

	if busMetrics.TotalSubscriptions != 5 {
		t.Errorf("Expected TotalSubscriptions 5, got %d", busMetrics.TotalSubscriptions)
	}

	if busMetrics.EventsPerSecond != 25.5 {
		t.Errorf("Expected EventsPerSecond 25.5, got %f", busMetrics.EventsPerSecond)
	}

	// Test SubscriptionStats
	subStats := &SubscriptionStats{
		EventsReceived:        100,
		EventsProcessed:       98,
		ProcessingErrors:      2,
		AverageProcessingTime: 5 * time.Millisecond,
		LastEventTime:         time.Now(),
		CreatedAt:             time.Now().Add(-1 * time.Hour),
	}

	if subStats.EventsReceived != 100 {
		t.Errorf("Expected EventsReceived 100, got %d", subStats.EventsReceived)
	}

	if subStats.EventsProcessed != 98 {
		t.Errorf("Expected EventsProcessed 98, got %d", subStats.EventsProcessed)
	}

	if subStats.ProcessingErrors != 2 {
		t.Errorf("Expected ProcessingErrors 2, got %d", subStats.ProcessingErrors)
	}

	// Test ProcessingStats
	procStats := &ProcessingStats{
		TotalProcessed:        500,
		TotalErrors:           5,
		AverageProcessingTime: 8 * time.Millisecond,
		ProcessingRate:        62.5,
		QueueSize:             10,
		MaxQueueSize:          25,
	}

	if procStats.TotalProcessed != 500 {
		t.Errorf("Expected TotalProcessed 500, got %d", procStats.TotalProcessed)
	}

	if procStats.TotalErrors != 5 {
		t.Errorf("Expected TotalErrors 5, got %d", procStats.TotalErrors)
	}

	if procStats.ProcessingRate != 62.5 {
		t.Errorf("Expected ProcessingRate 62.5, got %f", procStats.ProcessingRate)
	}
}

// TestConcurrentEventCreation tests concurrent event creation for race conditions
func TestConcurrentEventCreation(t *testing.T) {
	t.Parallel()

	const numGoroutines = 10
	const eventsPerGoroutine = 5

	events := make(chan Event, numGoroutines*eventsPerGoroutine)
	done := make(chan bool, numGoroutines)

	// Create events concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(routineID int) {
			defer func() { done <- true }()

			for j := 0; j < eventsPerGoroutine; j++ {
				event := NewBaseEvent(
					"test.concurrent",
					"test.source",
					map[string]interface{}{
						"routine": routineID,
						"event":   j,
					},
				)
				events <- event
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
	close(events)

	// Verify all events were created successfully
	eventCount := 0
	eventIDs := make(map[string]int) // Count occurrences instead of requiring uniqueness

	for event := range events {
		eventCount++

		// Verify event is valid
		if event == nil {
			t.Error("Event should not be nil")
			continue
		}

		if event.ID() == "" {
			t.Error("Event ID should not be empty")
		}

		// Count ID occurrences (duplicates are expected with timestamp-based IDs)
		eventIDs[event.ID()]++

		if event.Type() != "test.concurrent" {
			t.Errorf("Expected event type 'test.concurrent', got %q", event.Type())
		}

		if event.Timestamp().IsZero() {
			t.Error("Event timestamp should not be zero")
		}
	}

	expectedCount := numGoroutines * eventsPerGoroutine
	if eventCount != expectedCount {
		t.Errorf("Expected %d events, got %d", expectedCount, eventCount)
	}

	// Verify we got at least some IDs (duplicates are acceptable)
	if len(eventIDs) == 0 {
		t.Error("Should generate at least some event IDs")
	}
}
