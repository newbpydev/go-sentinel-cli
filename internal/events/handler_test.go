// Package events provides event handler tests for application event handling
package events

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/go-sentinel/pkg/models"
)

// TestNewAppEventHandler_FactoryFunction tests the factory function
func TestNewAppEventHandler_FactoryFunction(t *testing.T) {
	t.Parallel()

	handler := NewAppEventHandler()
	if handler == nil {
		t.Fatal("NewAppEventHandler should not return nil")
	}

	// Verify interface compliance
	_, ok := handler.(AppEventHandler)
	if !ok {
		t.Fatal("NewAppEventHandler should return AppEventHandler interface")
	}

	// Verify logger is set
	logger := handler.GetLogger()
	if logger == nil {
		t.Fatal("NewAppEventHandler should have a logger")
	}
}

// TestNewAppEventHandlerWithLogger_FactoryFunction tests the factory function with logger
func TestNewAppEventHandlerWithLogger_FactoryFunction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		logger *log.Logger
	}{
		{
			name:   "with_custom_logger",
			logger: log.New(&bytes.Buffer{}, "[test] ", log.LstdFlags),
		},
		{
			name:   "with_nil_logger",
			logger: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := NewAppEventHandlerWithLogger(tt.logger)
			if handler == nil {
				t.Fatal("NewAppEventHandlerWithLogger should not return nil")
			}

			// Verify interface compliance
			_, ok := handler.(AppEventHandler)
			if !ok {
				t.Fatal("NewAppEventHandlerWithLogger should return AppEventHandler interface")
			}

			// Verify logger is set (should use default if nil was passed)
			logger := handler.GetLogger()
			if logger == nil {
				t.Fatal("NewAppEventHandlerWithLogger should have a logger")
			}
		})
	}
}

// TestDefaultAppEventHandler_OnStartup tests startup event handling
func TestDefaultAppEventHandler_OnStartup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		ctx         context.Context
		expectError bool
	}{
		{
			name:        "normal_startup",
			ctx:         context.Background(),
			expectError: false,
		},
		{
			name:        "cancelled_context",
			ctx:         func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx }(),
			expectError: true,
		},
		{
			name: "timeout_context",
			ctx: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
				time.Sleep(1 * time.Millisecond)
				defer cancel()
				return ctx
			}(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			logger := log.New(&buf, "[test] ", 0)
			handler := NewAppEventHandlerWithLogger(logger)

			err := handler.OnStartup(tt.ctx)

			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error for cancelled/timeout context")
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got: %v", err)
				}

				// Verify startup messages were logged
				output := buf.String()
				if !strings.Contains(output, "Application starting up") {
					t.Error("Expected startup message in log output")
				}
				if !strings.Contains(output, "startup completed successfully") {
					t.Error("Expected startup completion message in log output")
				}
			}
		})
	}
}

// TestDefaultAppEventHandler_OnShutdown tests shutdown event handling
func TestDefaultAppEventHandler_OnShutdown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ctx  context.Context
	}{
		{
			name: "normal_shutdown",
			ctx:  context.Background(),
		},
		{
			name: "cancelled_context",
			ctx:  func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx }(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			logger := log.New(&buf, "[test] ", 0)
			handler := NewAppEventHandlerWithLogger(logger)

			err := handler.OnShutdown(tt.ctx)

			// OnShutdown should never return an error
			if err != nil {
				t.Fatalf("OnShutdown should not return error, got: %v", err)
			}

			// Verify shutdown messages were logged
			output := buf.String()
			if !strings.Contains(output, "Application shutting down") {
				t.Error("Expected shutdown message in log output")
			}

			if tt.name == "cancelled_context" {
				if !strings.Contains(output, "context cancellation") {
					t.Error("Expected context cancellation message in log output")
				}
			} else {
				if !strings.Contains(output, "shutdown completed") {
					t.Error("Expected shutdown completion message in log output")
				}
			}
		})
	}
}

// TestDefaultAppEventHandler_OnError tests error event handling
func TestDefaultAppEventHandler_OnError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		err            error
		expectedOutput []string
	}{
		{
			name:           "nil_error",
			err:            nil,
			expectedOutput: []string{}, // No output expected for nil error
		},
		{
			name:           "regular_error",
			err:            errors.New("test error"),
			expectedOutput: []string{"ERROR: test error"},
		},
		{
			name: "sentinel_error_critical",
			err: &models.SentinelError{
				Type:     models.ErrorTypeValidation,
				Message:  "critical validation error",
				Severity: models.SeverityCritical,
				Context:  models.ErrorContext{Metadata: map[string]string{"key": "value"}},
				Cause:    errors.New("underlying cause"),
			},
			expectedOutput: []string{
				"CRITICAL ERROR [VALIDATION]: critical validation error",
				"Error context:",
				"Caused by: underlying cause",
			},
		},
		{
			name: "sentinel_error_warning",
			err: &models.SentinelError{
				Type:     models.ErrorTypeConfig,
				Message:  "config warning",
				Severity: models.SeverityWarning,
			},
			expectedOutput: []string{"WARNING [CONFIG]: config warning"},
		},
		{
			name: "sentinel_error_info",
			err: &models.SentinelError{
				Type:     models.ErrorTypeTestExecution,
				Message:  "info message",
				Severity: models.SeverityInfo,
			},
			expectedOutput: []string{"INFO [TEST_EXECUTION]: info message"},
		},
		{
			name: "sentinel_error_default",
			err: &models.SentinelError{
				Type:     models.ErrorTypeInternal,
				Message:  "unknown error",
				Severity: "invalid_severity",
			},
			expectedOutput: []string{"ERROR [INTERNAL]: unknown error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			logger := log.New(&buf, "", 0) // No prefix for easier testing
			handler := NewAppEventHandlerWithLogger(logger)

			handler.OnError(tt.err)

			output := buf.String()

			if len(tt.expectedOutput) == 0 {
				if output != "" {
					t.Errorf("Expected no output for nil error, got: %s", output)
				}
			} else {
				for _, expected := range tt.expectedOutput {
					if !strings.Contains(output, expected) {
						t.Errorf("Expected output to contain '%s', got: %s", expected, output)
					}
				}
			}
		})
	}
}

// TestDefaultAppEventHandler_OnConfigChanged tests configuration change handling
func TestDefaultAppEventHandler_OnConfigChanged(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		config         *AppConfig
		expectedOutput []string
	}{
		{
			name:           "nil_config",
			config:         nil,
			expectedOutput: []string{"Configuration updated: <nil>"},
		},
		{
			name: "basic_config",
			config: &AppConfig{
				Colors:    true,
				Verbosity: 1,
				Watch: AppWatchConfig{
					Enabled: false,
				},
			},
			expectedOutput: []string{
				"Configuration updated:",
				"Colors: true",
				"Verbosity: 1",
				"Watch enabled: false",
			},
		},
		{
			name: "full_config_with_watch",
			config: &AppConfig{
				Colors:    false,
				Verbosity: 2,
				Watch: AppWatchConfig{
					Enabled:      true,
					Debounce:     "500ms",
					RunOnStart:   true,
					ClearOnRerun: false,
				},
				Paths: AppPathsConfig{
					IncludePatterns: []string{"*.go", "*.test"},
					ExcludePatterns: []string{"vendor/*"},
					IgnorePatterns:  []string{"*.log", "*.tmp"},
				},
			},
			expectedOutput: []string{
				"Configuration updated:",
				"Colors: false",
				"Verbosity: 2",
				"Watch enabled: true",
				"Watch debounce: 500ms",
				"Watch run on start: true",
				"Watch clear on rerun: false",
				"Include patterns: [*.go *.test]",
				"Exclude patterns: [vendor/*]",
				"Watch ignore patterns: [*.log *.tmp]",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			logger := log.New(&buf, "", 0)
			handler := NewAppEventHandlerWithLogger(logger)

			handler.OnConfigChanged(tt.config)

			output := buf.String()

			for _, expected := range tt.expectedOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain '%s', got: %s", expected, output)
				}
			}
		})
	}
}

// TestDefaultAppEventHandler_TestEvents tests test-related event handling
func TestDefaultAppEventHandler_TestEvents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		verbosity      int
		testName       string
		success        bool
		expectedOutput []string
	}{
		{
			name:           "test_start_low_verbosity",
			verbosity:      0,
			testName:       "TestExample",
			expectedOutput: []string{}, // No output expected for low verbosity
		},
		{
			name:           "test_start_high_verbosity",
			verbosity:      2,
			testName:       "TestExample",
			expectedOutput: []string{"Test started: TestExample"},
		},
		{
			name:           "test_complete_success_high_verbosity",
			verbosity:      2,
			testName:       "TestExample",
			success:        true,
			expectedOutput: []string{"Test completed: TestExample [PASS]"},
		},
		{
			name:           "test_complete_failure_high_verbosity",
			verbosity:      2,
			testName:       "TestExample",
			success:        false,
			expectedOutput: []string{"Test completed: TestExample [FAIL]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			logger := log.New(&buf, "", 0)
			handler := NewAppEventHandlerWithLogger(logger)

			// Set verbosity
			handler.SetVerbosity(tt.verbosity)

			if strings.Contains(tt.name, "start") {
				handler.OnTestStart(tt.testName)
			} else {
				handler.OnTestComplete(tt.testName, tt.success)
			}

			output := buf.String()

			if len(tt.expectedOutput) == 0 {
				if output != "" {
					t.Errorf("Expected no output for low verbosity, got: %s", output)
				}
			} else {
				for _, expected := range tt.expectedOutput {
					if !strings.Contains(output, expected) {
						t.Errorf("Expected output to contain '%s', got: %s", expected, output)
					}
				}
			}
		})
	}
}

// TestDefaultAppEventHandler_OnWatchEvent tests watch event handling
func TestDefaultAppEventHandler_OnWatchEvent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		verbosity      int
		filePath       string
		eventType      string
		expectedOutput []string
	}{
		{
			name:           "watch_event_no_verbosity",
			verbosity:      0,
			filePath:       "/path/to/file.go",
			eventType:      "WRITE",
			expectedOutput: []string{}, // No output expected for verbosity 0
		},
		{
			name:           "watch_event_with_verbosity",
			verbosity:      1,
			filePath:       "/path/to/file.go",
			eventType:      "WRITE",
			expectedOutput: []string{"File watch event: /path/to/file.go [WRITE]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			logger := log.New(&buf, "", 0)
			handler := NewAppEventHandlerWithLogger(logger)

			// Set verbosity
			handler.SetVerbosity(tt.verbosity)

			handler.OnWatchEvent(tt.filePath, tt.eventType)

			output := buf.String()

			if len(tt.expectedOutput) == 0 {
				if output != "" {
					t.Errorf("Expected no output for low verbosity, got: %s", output)
				}
			} else {
				for _, expected := range tt.expectedOutput {
					if !strings.Contains(output, expected) {
						t.Errorf("Expected output to contain '%s', got: %s", expected, output)
					}
				}
			}
		})
	}
}

// TestDefaultAppEventHandler_LoggerManagement tests logger management
func TestDefaultAppEventHandler_LoggerManagement(t *testing.T) {
	t.Parallel()

	handler := NewAppEventHandler()

	// Test initial logger
	initialLogger := handler.GetLogger()
	if initialLogger == nil {
		t.Fatal("Initial logger should not be nil")
	}

	// Test setting custom logger
	var buf bytes.Buffer
	customLogger := log.New(&buf, "[custom] ", 0)
	handler.SetLogger(customLogger)

	newLogger := handler.GetLogger()
	if newLogger != customLogger {
		t.Error("SetLogger should update the logger")
	}

	// Test setting nil logger (should be ignored)
	handler.SetLogger(nil)
	afterNilLogger := handler.GetLogger()
	if afterNilLogger != customLogger {
		t.Error("Setting nil logger should not change the current logger")
	}
}

// TestDefaultAppEventHandler_VerbosityManagement tests verbosity management
func TestDefaultAppEventHandler_VerbosityManagement(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	handler := NewAppEventHandlerWithLogger(logger)

	// Test setting verbosity levels
	verbosityLevels := []int{0, 1, 2, 3}

	for _, level := range verbosityLevels {
		t.Run(fmt.Sprintf("verbosity_%d", level), func(t *testing.T) {
			buf.Reset()
			handler.SetVerbosity(level)

			// Test debug logging (should only appear at verbosity > 1)
			handler.LogDebug("debug message")
			output := buf.String()

			if level > 1 {
				if !strings.Contains(output, "[DEBUG] debug message") {
					t.Errorf("Expected debug message at verbosity %d", level)
				}
			} else {
				if strings.Contains(output, "debug message") {
					t.Errorf("Did not expect debug message at verbosity %d", level)
				}
			}
		})
	}
}

// TestDefaultAppEventHandler_LoggingMethods tests logging utility methods
func TestDefaultAppEventHandler_LoggingMethods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		logFunc        func(AppEventHandler)
		expectedPrefix string
	}{
		{
			name: "log_info",
			logFunc: func(h AppEventHandler) {
				h.LogInfo("info message with %s", "args")
			},
			expectedPrefix: "[INFO]",
		},
		{
			name: "log_warning",
			logFunc: func(h AppEventHandler) {
				h.LogWarning("warning message with %d", 42)
			},
			expectedPrefix: "[WARNING]",
		},
		{
			name: "log_error",
			logFunc: func(h AppEventHandler) {
				h.LogError("error message with %v", errors.New("test"))
			},
			expectedPrefix: "[ERROR]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			logger := log.New(&buf, "", 0)
			handler := NewAppEventHandlerWithLogger(logger)

			tt.logFunc(handler)

			output := buf.String()
			if !strings.Contains(output, tt.expectedPrefix) {
				t.Errorf("Expected output to contain '%s', got: %s", tt.expectedPrefix, output)
			}
		})
	}
}

// TestDefaultAppEventHandler_InterfaceCompliance tests interface compliance
func TestDefaultAppEventHandler_InterfaceCompliance(t *testing.T) {
	t.Parallel()

	// Test that DefaultAppEventHandler implements the interface
	var _ AppEventHandler = (*DefaultAppEventHandler)(nil)

	// Test factory functions return the interface
	handler1 := NewAppEventHandler()
	handler2 := NewAppEventHandlerWithLogger(nil)

	_, ok1 := handler1.(AppEventHandler)
	_, ok2 := handler2.(AppEventHandler)

	if !ok1 {
		t.Fatal("NewAppEventHandler should return AppEventHandler")
	}
	if !ok2 {
		t.Fatal("NewAppEventHandlerWithLogger should return AppEventHandler")
	}
}

// Note: fmt.Sprintf is already imported via the "fmt" package
