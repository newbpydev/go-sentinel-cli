package tui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/go-sentinel/internal/ui"
)

// TestSafeGoroutine tests that safeGoroutine recovers from panics
func TestSafeGoroutine(t *testing.T) {
	// Create test app
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a proper mock program to avoid nil pointer issues
	mockProg := &mockProgramImplementation{
		messageReceived: make(chan tea.Msg, 10),
	}

	app := &App{
		errorCh:    make(chan error, 10),
		ctx:        ctx,
		testEvents: make(chan []byte, 10),
		program:    mockProg,
	}

	// Run a goroutine to check the error channel
	errorReceived := make(chan bool, 1)
	go func() {
		select {
		case err := <-app.errorCh:
			// Verify the error contains the panic message
			// The format is "panic in testFunc: test panic"
			if err != nil && err.Error() == "panic in testFunc: test panic" {
				errorReceived <- true
			} else {
				t.Errorf("Got unexpected error: %v", err)
				errorReceived <- false
			}
		case <-time.After(1 * time.Second):
			t.Error("Timeout waiting for panic recovery")
			errorReceived <- false
		}
	}()

	// Run a function that will panic in a safeGoroutine
	app.safeGoroutine("testFunc", func() {
		panic("test panic")
	})

	// Wait for the test result
	result := <-errorReceived
	if !result {
		t.Fail()
	}
}

// mockMsg is a simple test message
type mockMsg struct {
	Content string
}

// mockProgramImplementation is a mock program for testing
type mockProgramImplementation struct {
	messageReceived chan tea.Msg
}

func (mp *mockProgramImplementation) Send(msg tea.Msg) {
	mp.messageReceived <- msg
}

func (mp *mockProgramImplementation) Start() error {
	return nil
}

func (mp *mockProgramImplementation) Quit() {
	// Do nothing
}

// TestErrorHandlerLoop verifies that errors are properly processed
func TestErrorHandlerLoop(t *testing.T) {
	// Create a context we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a mock program to receive messages
	mockProg := &mockProgramImplementation{
		messageReceived: make(chan tea.Msg, 10),
	}

	// Create the app with our mock
	app := &App{
		errorCh:    make(chan error, 10),
		ctx:        ctx,
		program:    mockProg,
		testEvents: make(chan []byte, 10),
	}

	// Start the error handler loop
	go app.errorHandlerLoop()

	// Send a test error
	testErr := errors.New("test error message")
	app.errorCh <- testErr

	// Check if we receive the error message
	// Create a timeout for the entire test
	testTimeout := time.After(2 * time.Second)
	
	// Keep reading messages until we find one with our error
	found := false
	for !found {
		select {
		case msg := <-mockProg.messageReceived:
			if logEntry, ok := msg.(ui.LogEntryMsg); ok {
				// Check if this log entry contains our error message
				if strings.Contains(logEntry.Content, testErr.Error()) {
					// Found it! Test passed
					found = true
				}
			}
			// Ignore other message types (ShowLogViewMsg, etc.)
		case <-testTimeout:
			t.Fatalf("Timeout waiting for log message containing '%s'", testErr.Error())
			return
		}
	}
	
	// If we get here with found=true, the test passes
}

// TestPipelineRecovery verifies that components in the pipeline can recover from panics
func TestPipelineRecovery(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a mock program to receive messages
	mockProg := &mockProgramImplementation{
		messageReceived: make(chan tea.Msg, 10),
	}

	// Create the app with our mock
	app := &App{
		errorCh:    make(chan error, 10),
		ctx:        ctx,
		program:    mockProg,
		testEvents: make(chan []byte, 10),
	}

	// Start the error handler loop
	go app.errorHandlerLoop()

	// Simulate a component failing with panic
	app.safeGoroutine("pipelineComponent", func() {
		// Do some "normal" work first
		time.Sleep(50 * time.Millisecond)
		
		// Then panic
		panic("pipeline component failure")
	})

	// Define what we're looking for in the error message
	expectedFragment := "pipeline component failure"
	
	// Create a timeout for the entire test
	testTimeout := time.After(2 * time.Second)
	
	// Keep reading messages until we find one with our panic message
	found := false
	for !found {
		select {
		case msg := <-mockProg.messageReceived:
			if logEntry, ok := msg.(ui.LogEntryMsg); ok {
				// Check if this log entry contains our panic message
				if strings.Contains(logEntry.Content, expectedFragment) {
					// Found it! Test passed
					found = true
				}
			}
			// Ignore other message types (ShowLogViewMsg, etc.)
		case <-testTimeout:
			t.Fatalf("Timeout waiting for log message containing '%s'", expectedFragment)
			return
		}
	}
	
	// If we get here with found=true, the test passes
}
