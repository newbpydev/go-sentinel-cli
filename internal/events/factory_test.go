// Package events provides factory tests for event handler creation
package events

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/newbpydev/go-sentinel/pkg/models"
)

// TestNewAppEventHandlerFactory_FactoryFunction tests the factory creation
func TestNewAppEventHandlerFactory_FactoryFunction(t *testing.T) {
	t.Parallel()

	factory := NewAppEventHandlerFactory()
	if factory == nil {
		t.Fatal("NewAppEventHandlerFactory should not return nil")
	}

	// Verify interface compliance
	_, ok := factory.(AppEventHandlerFactory)
	if !ok {
		t.Fatal("NewAppEventHandlerFactory should return AppEventHandlerFactory interface")
	}
}

// TestNewAppEventHandlerFactoryWithDependencies_FactoryFunction tests factory with dependencies
func TestNewAppEventHandlerFactoryWithDependencies_FactoryFunction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		dependencies AppEventHandlerDependencies
	}{
		{
			name:         "nil_logger",
			dependencies: AppEventHandlerDependencies{Logger: nil, Verbosity: 0},
		},
		{
			name:         "custom_logger",
			dependencies: AppEventHandlerDependencies{Logger: log.New(&bytes.Buffer{}, "[test] ", 0), Verbosity: 1},
		},
		{
			name:         "high_verbosity",
			dependencies: AppEventHandlerDependencies{Logger: nil, Verbosity: 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			factory := NewAppEventHandlerFactoryWithDependencies(tt.dependencies)
			if factory == nil {
				t.Fatal("NewAppEventHandlerFactoryWithDependencies should not return nil")
			}

			// Verify interface compliance
			_, ok := factory.(AppEventHandlerFactory)
			if !ok {
				t.Fatal("NewAppEventHandlerFactoryWithDependencies should return AppEventHandlerFactory interface")
			}
		})
	}
}

// TestDefaultAppEventHandlerFactory_CreateEventHandler tests basic handler creation
func TestDefaultAppEventHandlerFactory_CreateEventHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		dependencies *AppEventHandlerDependencies
	}{
		{
			name:         "default_factory",
			dependencies: nil,
		},
		{
			name:         "factory_with_logger",
			dependencies: &AppEventHandlerDependencies{Logger: log.New(&bytes.Buffer{}, "[factory] ", 0)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var factory AppEventHandlerFactory
			if tt.dependencies == nil {
				factory = NewAppEventHandlerFactory()
			} else {
				factory = NewAppEventHandlerFactoryWithDependencies(*tt.dependencies)
			}

			handler := factory.CreateEventHandler()

			if handler == nil {
				t.Fatal("CreateEventHandler should not return nil")
			}

			// Verify interface compliance
			_, ok := handler.(AppEventHandler)
			if !ok {
				t.Fatal("CreateEventHandler should return AppEventHandler interface")
			}

			// Verify handler is functional
			logger := handler.GetLogger()
			if logger == nil {
				t.Fatal("Handler should have a logger")
			}
		})
	}
}

// TestDefaultAppEventHandlerFactory_CreateEventHandlerWithLogger tests handler creation with custom logger
func TestDefaultAppEventHandlerFactory_CreateEventHandlerWithLogger(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		logger *log.Logger
	}{
		{
			name:   "custom_logger",
			logger: log.New(&bytes.Buffer{}, "[custom] ", log.LstdFlags),
		},
		{
			name:   "nil_logger",
			logger: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			factory := NewAppEventHandlerFactory()
			handler := factory.CreateEventHandlerWithLogger(tt.logger)

			if handler == nil {
				t.Fatal("CreateEventHandlerWithLogger should not return nil")
			}

			// Verify interface compliance
			_, ok := handler.(AppEventHandler)
			if !ok {
				t.Fatal("CreateEventHandlerWithLogger should return AppEventHandler interface")
			}

			// Verify handler has a logger (should use default if nil was passed)
			logger := handler.GetLogger()
			if logger == nil {
				t.Fatal("Handler should have a logger")
			}
		})
	}
}

// TestDefaultAppEventHandlerFactory_VerbosityConfiguration tests verbosity configuration through dependencies
func TestDefaultAppEventHandlerFactory_VerbosityConfiguration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		dependencies AppEventHandlerDependencies
		expectDebug  bool
	}{
		{
			name:         "low_verbosity",
			dependencies: AppEventHandlerDependencies{Verbosity: 0},
			expectDebug:  false,
		},
		{
			name:         "high_verbosity",
			dependencies: AppEventHandlerDependencies{Verbosity: 2},
			expectDebug:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			factory := NewAppEventHandlerFactoryWithDependencies(tt.dependencies)
			handler := factory.CreateEventHandler()

			if handler == nil {
				t.Fatal("CreateEventHandler should not return nil")
			}

			// Verify interface compliance
			_, ok := handler.(AppEventHandler)
			if !ok {
				t.Fatal("CreateEventHandler should return AppEventHandler interface")
			}

			// Test verbosity behavior
			var buf bytes.Buffer
			testLogger := log.New(&buf, "", 0)
			handler.SetLogger(testLogger)
			handler.SetVerbosity(tt.dependencies.Verbosity)

			handler.LogDebug("debug message")
			output := buf.String()

			if tt.expectDebug {
				if output == "" {
					t.Error("Debug messages should appear with high verbosity")
				}
			} else {
				if output != "" {
					t.Error("Debug messages should not appear with low verbosity")
				}
			}
		})
	}
}

// TestDefaultAppEventHandlerFactory_MultipleHandlers tests creating multiple handlers
func TestDefaultAppEventHandlerFactory_MultipleHandlers(t *testing.T) {
	t.Parallel()

	factory := NewAppEventHandlerFactory()

	// Create multiple handlers
	handler1 := factory.CreateEventHandler()
	handler2 := factory.CreateEventHandler()
	handler3 := factory.CreateEventHandlerWithLogger(nil)

	// Verify they are different instances
	if handler1 == handler2 {
		t.Fatal("Factory should create different handler instances")
	}
	if handler1 == handler3 {
		t.Fatal("Factory should create different handler instances")
	}
	if handler2 == handler3 {
		t.Fatal("Factory should create different handler instances")
	}

	// Verify they are independent
	handler1.SetVerbosity(1)
	handler2.SetVerbosity(2)

	// Test independence by checking debug logging behavior
	var buf1, buf2 bytes.Buffer
	logger1 := log.New(&buf1, "", 0)
	logger2 := log.New(&buf2, "", 0)

	handler1.SetLogger(logger1)
	handler2.SetLogger(logger2)

	handler1.LogDebug("debug1")
	handler2.LogDebug("debug2")

	output1 := buf1.String()
	output2 := buf2.String()

	// handler1 (verbosity 1) should not show debug
	if output1 != "" {
		t.Error("Handler1 should not show debug at verbosity 1")
	}

	// handler2 (verbosity 2) should show debug
	if output2 == "" {
		t.Error("Handler2 should show debug at verbosity 2")
	}
}

// TestDefaultAppEventHandlerFactory_DependencyInjection tests dependency injection pattern
func TestDefaultAppEventHandlerFactory_DependencyInjection(t *testing.T) {
	t.Parallel()

	// Test with different dependency configurations
	tests := []struct {
		name         string
		dependencies AppEventHandlerDependencies
		description  string
	}{
		{
			name:         "minimal_dependencies",
			dependencies: AppEventHandlerDependencies{Logger: nil, Verbosity: 0},
			description:  "Factory with minimal dependencies",
		},
		{
			name:         "custom_logger_dependency",
			dependencies: AppEventHandlerDependencies{Logger: log.New(&bytes.Buffer{}, "[injected] ", 0), Verbosity: 1},
			description:  "Factory with custom logger dependency",
		},
		{
			name:         "high_verbosity_dependency",
			dependencies: AppEventHandlerDependencies{Logger: nil, Verbosity: 2},
			description:  "Factory with high verbosity dependency",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			factory := NewAppEventHandlerFactoryWithDependencies(tt.dependencies)

			// Create handlers using different methods
			handler1 := factory.CreateEventHandler()
			handler2 := factory.CreateEventHandlerWithLogger(tt.dependencies.Logger)

			// Verify both handlers work
			if handler1 == nil {
				t.Fatal("CreateEventHandler should not return nil")
			}
			if handler2 == nil {
				t.Fatal("CreateEventHandlerWithDefaults should not return nil")
			}

			// Test functionality
			var buf bytes.Buffer
			testLogger := log.New(&buf, "", 0)

			handler1.SetLogger(testLogger)
			handler1.LogInfo("test message")

			output := buf.String()
			if !bytes.Contains([]byte(output), []byte("test message")) {
				t.Error("Handler should log info messages")
			}
		})
	}
}

// TestDefaultAppEventHandlerFactory_FactoryPattern tests factory pattern implementation
func TestDefaultAppEventHandlerFactory_FactoryPattern(t *testing.T) {
	t.Parallel()

	// Test that factory encapsulates creation logic
	factory := NewAppEventHandlerFactory()

	// Create multiple handlers and verify they follow the same pattern
	handlers := make([]AppEventHandler, 5)
	for i := 0; i < 5; i++ {
		handlers[i] = factory.CreateEventHandler()
		if handlers[i] == nil {
			t.Fatalf("Handler %d should not be nil", i)
		}
	}

	// Verify all handlers have the same interface but are different instances
	for i := 0; i < len(handlers); i++ {
		for j := i + 1; j < len(handlers); j++ {
			if handlers[i] == handlers[j] {
				t.Fatalf("Handlers %d and %d should be different instances", i, j)
			}
		}
	}

	// Verify all handlers work independently
	for i, handler := range handlers {
		var buf bytes.Buffer
		testLogger := log.New(&buf, "", 0)
		handler.SetLogger(testLogger)

		testMessage := fmt.Sprintf("test message %d", i)
		handler.LogInfo(testMessage)

		output := buf.String()
		if !bytes.Contains([]byte(output), []byte(testMessage)) {
			t.Errorf("Handler %d should log its specific message", i)
		}
	}
}

// TestDefaultAppEventHandlerFactory_InterfaceCompliance tests interface compliance
func TestDefaultAppEventHandlerFactory_InterfaceCompliance(t *testing.T) {
	t.Parallel()

	// Test that DefaultAppEventHandlerFactory implements the interface
	var _ AppEventHandlerFactory = (*DefaultAppEventHandlerFactory)(nil)

	// Test factory creation methods
	factory1 := NewAppEventHandlerFactory()
	factory2 := NewAppEventHandlerFactoryWithDependencies(AppEventHandlerDependencies{
		Logger:    log.New(&bytes.Buffer{}, "[test] ", 0),
		Verbosity: 1,
	})

	// Verify both implement the interface
	_, ok1 := factory1.(AppEventHandlerFactory)
	_, ok2 := factory2.(AppEventHandlerFactory)

	if !ok1 {
		t.Fatal("NewAppEventHandlerFactory should return AppEventHandlerFactory")
	}
	if !ok2 {
		t.Fatal("NewAppEventHandlerFactoryWithDependencies should return AppEventHandlerFactory")
	}
}

// TestDefaultAppEventHandlerFactory_LoggerInjection tests logger injection behavior
func TestDefaultAppEventHandlerFactory_LoggerInjection(t *testing.T) {
	t.Parallel()

	// Test factory without injected logger
	factory1 := NewAppEventHandlerFactory()
	handler1 := factory1.CreateEventHandler()

	logger1 := handler1.GetLogger()
	if logger1 == nil {
		t.Fatal("Handler from factory without injected logger should still have a logger")
	}

	// Test factory with injected logger
	var buf bytes.Buffer
	injectedLogger := log.New(&buf, "[injected] ", 0)
	factory2 := NewAppEventHandlerFactoryWithDependencies(AppEventHandlerDependencies{
		Logger: injectedLogger,
	})

	handler2 := factory2.CreateEventHandler()
	logger2 := handler2.GetLogger()

	if logger2 == nil {
		t.Fatal("Handler from factory with injected logger should have a logger")
	}

	// Test that the injected logger is used
	handler2.LogInfo("test message")
	output := buf.String()

	if !bytes.Contains([]byte(output), []byte("[injected]")) {
		t.Error("Handler should use the injected logger")
	}
	if !bytes.Contains([]byte(output), []byte("test message")) {
		t.Error("Handler should log the test message")
	}
}

// TestDefaultAppEventHandlerFactory_CreateEventHandlerWithDefaults tests the CreateEventHandlerWithDefaults method
func TestDefaultAppEventHandlerFactory_CreateEventHandlerWithDefaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		dependencies *AppEventHandlerDependencies
	}{
		{
			name:         "default_factory",
			dependencies: nil,
		},
		{
			name:         "factory_with_logger",
			dependencies: &AppEventHandlerDependencies{Logger: log.New(&bytes.Buffer{}, "[test] ", 0), Verbosity: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var factory *DefaultAppEventHandlerFactory
			if tt.dependencies == nil {
				factory = NewAppEventHandlerFactory().(*DefaultAppEventHandlerFactory)
			} else {
				factory = NewAppEventHandlerFactoryWithDependencies(*tt.dependencies).(*DefaultAppEventHandlerFactory)
			}

			// Call the method directly on the concrete type since it's not in the interface
			handler := factory.CreateEventHandlerWithDefaults()

			if handler == nil {
				t.Fatal("CreateEventHandlerWithDefaults should not return nil")
			}

			// Verify interface compliance
			_, ok := handler.(AppEventHandler)
			if !ok {
				t.Fatal("CreateEventHandlerWithDefaults should return AppEventHandler interface")
			}

			// Verify handler is functional
			logger := handler.GetLogger()
			if logger == nil {
				t.Fatal("Handler should have a logger")
			}

			// Test that defaults are applied (verbosity should be set to 0)
			var buf bytes.Buffer
			testLogger := log.New(&buf, "", 0)
			handler.SetLogger(testLogger)

			// Debug messages should not appear with default verbosity (0)
			handler.LogDebug("debug message")
			output := buf.String()
			if output != "" {
				t.Error("Debug messages should not appear with default verbosity")
			}
		})
	}
}

// TestDefaultAppEventHandler_OnError_EmptyMetadata tests OnError with SentinelError having empty metadata
func TestDefaultAppEventHandler_OnError_EmptyMetadata(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	handler := NewAppEventHandlerWithLogger(logger)

	// Test SentinelError with empty metadata (to cover the missing line)
	err := &models.SentinelError{
		Type:     models.ErrorTypeValidation,
		Message:  "test error",
		Severity: models.SeverityCritical,
		Context:  models.ErrorContext{Metadata: map[string]string{}}, // Empty metadata
		Cause:    nil,                                                // No cause
	}

	handler.OnError(err)

	output := buf.String()
	if !strings.Contains(output, "CRITICAL ERROR [VALIDATION]: test error") {
		t.Errorf("Expected critical error message, got: %s", output)
	}

	// Should not contain context or cause messages since they're empty
	if strings.Contains(output, "Error context:") {
		t.Error("Should not log empty context")
	}
	if strings.Contains(output, "Caused by:") {
		t.Error("Should not log nil cause")
	}
}

// TestDefaultAppEventHandler_OnError_WithMetadataAndCause tests OnError with SentinelError having metadata and cause
func TestDefaultAppEventHandler_OnError_WithMetadataAndCause(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	handler := NewAppEventHandlerWithLogger(logger)

	// Test SentinelError with metadata and cause (to cover the missing lines)
	err := &models.SentinelError{
		Type:     models.ErrorTypeValidation,
		Message:  "test error with context",
		Severity: models.SeverityError,
		Context: models.ErrorContext{
			Metadata: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		Cause: fmt.Errorf("underlying cause error"),
	}

	handler.OnError(err)

	output := buf.String()
	if !strings.Contains(output, "ERROR [VALIDATION]: test error with context") {
		t.Errorf("Expected error message, got: %s", output)
	}

	// Should contain context and cause messages
	if !strings.Contains(output, "Error context:") {
		t.Error("Should log non-empty context")
	}
	if !strings.Contains(output, "Caused by: underlying cause error") {
		t.Error("Should log cause when present")
	}
}
