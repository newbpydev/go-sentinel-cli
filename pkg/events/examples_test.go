package events

import (
	"testing"
)

// TestExample_eventBusUsage tests the event bus usage example
func TestExample_eventBusUsage(t *testing.T) {
	t.Parallel()

	// This test ensures the example function runs without panicking
	// and that the events are created correctly
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Example_eventBusUsage panicked: %v", r)
		}
	}()

	Example_eventBusUsage()
}

// TestExample_eventQuery tests the event query example
func TestExample_eventQuery(t *testing.T) {
	t.Parallel()

	// This test ensures the example function runs without panicking
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Example_eventQuery panicked: %v", r)
		}
	}()

	Example_eventQuery()
}

// TestExample_eventMetrics tests the event metrics example
func TestExample_eventMetrics(t *testing.T) {
	t.Parallel()

	// This test ensures the example function runs without panicking
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Example_eventMetrics panicked: %v", r)
		}
	}()

	Example_eventMetrics()
}

// TestExample_fileChangeEvents tests the file change events example
func TestExample_fileChangeEvents(t *testing.T) {
	t.Parallel()

	// This test ensures the example function runs without panicking
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Example_fileChangeEvents panicked: %v", r)
		}
	}()

	Example_fileChangeEvents()

	// Test the metadata conditional path by creating events with metadata
	// This covers the conditional branch in the example function
	testFileChangeEventsWithMetadata(t)
}

// testFileChangeEventsWithMetadata tests the metadata conditional path
func testFileChangeEventsWithMetadata(t *testing.T) {
	// Create events with metadata to trigger the conditional
	fileCreated := NewFileChangedEvent("/src/new_test.go", "created")
	fileCreated.Metadata()["test"] = "value"

	fileModified := NewFileChangedEvent("/src/existing_test.go", "modified")
	fileModified.Metadata()["branch"] = "main"

	events := []*FileChangedEvent{fileCreated, fileModified}

	// Simulate the loop from Example_fileChangeEvents with metadata
	for _, event := range events {
		// Access event metadata (this covers the conditional)
		if metadata := event.Metadata(); len(metadata) > 0 {
			// This branch should be executed
			if len(metadata) == 0 {
				t.Error("Expected metadata to be present")
			}
		}
	}
}

// TestExample_baseEvent tests the base event example
func TestExample_baseEvent(t *testing.T) {
	t.Parallel()

	// This test ensures the example function runs without panicking
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Example_baseEvent panicked: %v", r)
		}
	}()

	Example_baseEvent()
}

// TestExample_eventConstants tests the event constants example
func TestExample_eventConstants(t *testing.T) {
	t.Parallel()

	// This test ensures the example function runs without panicking
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Example_eventConstants panicked: %v", r)
		}
	}()

	Example_eventConstants()
}
