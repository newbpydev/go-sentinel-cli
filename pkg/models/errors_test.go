package models

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// TestSentinelError_Error tests the Error() method formatting
func TestSentinelError_Error(t *testing.T) {
	testCases := []struct {
		name     string
		error    *SentinelError
		expected string
	}{
		{
			name: "Basic error with type and severity",
			error: &SentinelError{
				Type:     ErrorTypeConfig,
				Severity: SeverityError,
				Message:  "configuration is invalid",
				Context:  ErrorContext{},
			},
			expected: "[CONFIG:ERROR] configuration is invalid",
		},
		{
			name: "Error with component",
			error: &SentinelError{
				Type:     ErrorTypeFileSystem,
				Severity: SeverityError,
				Message:  "file not found",
				Context: ErrorContext{
					Component: "watcher",
				},
			},
			expected: "[FILESYSTEM:ERROR] (watcher) file not found",
		},
		{
			name: "Error with resource and operation",
			error: &SentinelError{
				Type:     ErrorTypeTestExecution,
				Severity: SeverityError,
				Message:  "test failed",
				Context: ErrorContext{
					Component: "test_runner",
					Resource:  "pkg/example",
					Operation: "test_execution",
				},
			},
			expected: "[TEST_EXECUTION:ERROR] (test_runner) test failed resource=pkg/example operation=test_execution",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.error.Error()
			if result != tc.expected {
				t.Errorf("Expected error string '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

// TestSentinelError_Unwrap tests error unwrapping
func TestSentinelError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := WrapError(originalErr, ErrorTypeInternal, SeverityError, "wrapped message")

	unwrapped := wrappedErr.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Expected unwrapped error to be original error, got %v", unwrapped)
	}
}

// TestSentinelError_Is tests error comparison
func TestSentinelError_Is(t *testing.T) {
	err1 := NewError(ErrorTypeConfig, SeverityError, "config error")
	err2 := NewError(ErrorTypeConfig, SeverityError, "config error")
	err3 := NewError(ErrorTypeConfig, SeverityError, "different message")
	err4 := NewError(ErrorTypeFileSystem, SeverityError, "config error")

	// Same type and message should be equal
	if !err1.Is(err2) {
		t.Error("Expected err1.Is(err2) to be true")
	}

	// Different message should not be equal
	if err1.Is(err3) {
		t.Error("Expected err1.Is(err3) to be false")
	}

	// Different type should not be equal
	if err1.Is(err4) {
		t.Error("Expected err1.Is(err4) to be false")
	}

	// Non-SentinelError should not be equal
	regularErr := errors.New("regular error")
	if err1.Is(regularErr) {
		t.Error("Expected err1.Is(regularErr) to be false")
	}
}

// TestSentinelError_UserMessage tests user-safe message generation
func TestSentinelError_UserMessage(t *testing.T) {
	testCases := []struct {
		name     string
		error    *SentinelError
		expected string
	}{
		{
			name: "User-safe error returns original message",
			error: &SentinelError{
				Type:     ErrorTypeValidation,
				Severity: SeverityWarning,
				Message:  "Invalid input provided",
				UserSafe: true,
			},
			expected: "Invalid input provided",
		},
		{
			name: "Internal error returns generic message",
			error: &SentinelError{
				Type:     ErrorTypeInternal,
				Severity: SeverityCritical,
				Message:  "Database connection failed with credentials xyz",
				UserSafe: false,
			},
			expected: "An internal error occurred",
		},
		{
			name: "Config error returns generic config message",
			error: &SentinelError{
				Type:     ErrorTypeConfig,
				Severity: SeverityError,
				Message:  "Secret key not found in environment",
				UserSafe: false,
			},
			expected: "Configuration error occurred",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.error.UserMessage()
			if result != tc.expected {
				t.Errorf("Expected user message '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

// TestSentinelError_WithContext tests context addition
func TestSentinelError_WithContext(t *testing.T) {
	err := NewError(ErrorTypeConfig, SeverityError, "test error")

	// Add context
	err = err.WithContext("key1", "value1")
	err = err.WithContext("key2", "value2")

	// Verify context was added
	if err.Context.Metadata["key1"] != "value1" {
		t.Errorf("Expected metadata key1 to be 'value1', got '%s'", err.Context.Metadata["key1"])
	}
	if err.Context.Metadata["key2"] != "value2" {
		t.Errorf("Expected metadata key2 to be 'value2', got '%s'", err.Context.Metadata["key2"])
	}
}

// TestSentinelError_WithRequestID tests request ID addition
func TestSentinelError_WithRequestID(t *testing.T) {
	err := NewError(ErrorTypeConfig, SeverityError, "test error")
	requestID := "req-123-456"

	err = err.WithRequestID(requestID)

	if err.Context.RequestID != requestID {
		t.Errorf("Expected request ID '%s', got '%s'", requestID, err.Context.RequestID)
	}
}

// TestSentinelError_WithUserID tests user ID addition
func TestSentinelError_WithUserID(t *testing.T) {
	err := NewError(ErrorTypeConfig, SeverityError, "test error")
	userID := "user-789"

	err = err.WithUserID(userID)

	if err.Context.UserID != userID {
		t.Errorf("Expected user ID '%s', got '%s'", userID, err.Context.UserID)
	}
}

// TestNewError tests basic error creation
func TestNewError(t *testing.T) {
	err := NewError(ErrorTypeConfig, SeverityError, "test message")

	if err.Type != ErrorTypeConfig {
		t.Errorf("Expected type %s, got %s", ErrorTypeConfig, err.Type)
	}
	if err.Severity != SeverityError {
		t.Errorf("Expected severity %s, got %s", SeverityError, err.Severity)
	}
	if err.Message != "test message" {
		t.Errorf("Expected message 'test message', got '%s'", err.Message)
	}
	if err.Cause != nil {
		t.Errorf("Expected no cause, got %v", err.Cause)
	}
	if len(err.Stack) == 0 {
		t.Error("Expected stack trace to be captured")
	}
}

// TestWrapError tests error wrapping
func TestWrapError(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := WrapError(originalErr, ErrorTypeFileSystem, SeverityError, "wrapped message")

	if wrappedErr.Type != ErrorTypeFileSystem {
		t.Errorf("Expected type %s, got %s", ErrorTypeFileSystem, wrappedErr.Type)
	}
	if wrappedErr.Severity != SeverityError {
		t.Errorf("Expected severity %s, got %s", SeverityError, wrappedErr.Severity)
	}
	if wrappedErr.Message != "wrapped message" {
		t.Errorf("Expected message 'wrapped message', got '%s'", wrappedErr.Message)
	}
	if wrappedErr.Cause != originalErr {
		t.Errorf("Expected cause to be original error, got %v", wrappedErr.Cause)
	}
	if len(wrappedErr.Stack) == 0 {
		t.Error("Expected stack trace to be captured")
	}
}

// TestWrapError_NilError tests wrapping nil error
func TestWrapError_NilError(t *testing.T) {
	wrappedErr := WrapError(nil, ErrorTypeFileSystem, SeverityError, "wrapped message")
	if wrappedErr != nil {
		t.Errorf("Expected nil when wrapping nil error, got %v", wrappedErr)
	}
}

// TestNewConfigError tests configuration error creation
func TestNewConfigError(t *testing.T) {
	// User-safe config error
	userSafeErr := NewConfigError("Invalid timeout value", true)
	if userSafeErr.Type != ErrorTypeConfig {
		t.Errorf("Expected type %s, got %s", ErrorTypeConfig, userSafeErr.Type)
	}
	if userSafeErr.Severity != SeverityWarning {
		t.Errorf("Expected severity %s, got %s", SeverityWarning, userSafeErr.Severity)
	}
	if !userSafeErr.UserSafe {
		t.Error("Expected user-safe error to be marked as user-safe")
	}
	if userSafeErr.Context.Component != "config" {
		t.Errorf("Expected component 'config', got '%s'", userSafeErr.Context.Component)
	}

	// Internal config error
	internalErr := NewConfigError("Database password missing", false)
	if internalErr.Severity != SeverityError {
		t.Errorf("Expected severity %s, got %s", SeverityError, internalErr.Severity)
	}
	if internalErr.UserSafe {
		t.Error("Expected internal error to not be user-safe")
	}
}

// TestNewValidationError tests validation error creation
func TestNewValidationError(t *testing.T) {
	err := NewValidationError("email", "Email format is invalid")

	if err.Type != ErrorTypeValidation {
		t.Errorf("Expected type %s, got %s", ErrorTypeValidation, err.Type)
	}
	if err.Severity != SeverityWarning {
		t.Errorf("Expected severity %s, got %s", SeverityWarning, err.Severity)
	}
	if !err.UserSafe {
		t.Error("Expected validation error to be user-safe")
	}
	if err.Context.Component != "validation" {
		t.Errorf("Expected component 'validation', got '%s'", err.Context.Component)
	}
	if err.Context.Resource != "email" {
		t.Errorf("Expected resource 'email', got '%s'", err.Context.Resource)
	}
}

// TestNewFileSystemError tests file system error creation
func TestNewFileSystemError(t *testing.T) {
	originalErr := errors.New("permission denied")
	err := NewFileSystemError("read", "/path/to/file", originalErr)

	if err.Type != ErrorTypeFileSystem {
		t.Errorf("Expected type %s, got %s", ErrorTypeFileSystem, err.Type)
	}
	if err.Severity != SeverityError {
		t.Errorf("Expected severity %s, got %s", SeverityError, err.Severity)
	}
	if err.Cause != originalErr {
		t.Errorf("Expected cause to be original error, got %v", err.Cause)
	}
	if err.Context.Operation != "read" {
		t.Errorf("Expected operation 'read', got '%s'", err.Context.Operation)
	}
	if err.Context.Resource != "/path/to/file" {
		t.Errorf("Expected resource '/path/to/file', got '%s'", err.Context.Resource)
	}
	if err.Context.Component != "filesystem" {
		t.Errorf("Expected component 'filesystem', got '%s'", err.Context.Component)
	}
}

// TestNewTestExecutionError tests test execution error creation
func TestNewTestExecutionError(t *testing.T) {
	originalErr := errors.New("test failed")
	err := NewTestExecutionError("pkg/example", originalErr)

	if err.Type != ErrorTypeTestExecution {
		t.Errorf("Expected type %s, got %s", ErrorTypeTestExecution, err.Type)
	}
	if err.Context.Resource != "pkg/example" {
		t.Errorf("Expected resource 'pkg/example', got '%s'", err.Context.Resource)
	}
	if err.Context.Component != "test_runner" {
		t.Errorf("Expected component 'test_runner', got '%s'", err.Context.Component)
	}
}

// TestNewWatchError tests watch error creation
func TestNewWatchError(t *testing.T) {
	originalErr := errors.New("file watcher failed")
	err := NewWatchError("add_path", "/watch/path", originalErr)

	if err.Type != ErrorTypeWatch {
		t.Errorf("Expected type %s, got %s", ErrorTypeWatch, err.Type)
	}
	if err.Context.Operation != "add_path" {
		t.Errorf("Expected operation 'add_path', got '%s'", err.Context.Operation)
	}
	if err.Context.Resource != "/watch/path" {
		t.Errorf("Expected resource '/watch/path', got '%s'", err.Context.Resource)
	}
	if err.Context.Component != "watcher" {
		t.Errorf("Expected component 'watcher', got '%s'", err.Context.Component)
	}
}

// TestNewDependencyError tests dependency error creation
func TestNewDependencyError(t *testing.T) {
	originalErr := errors.New("component not found")
	err := NewDependencyError("testExecutor", originalErr)

	if err.Type != ErrorTypeDependency {
		t.Errorf("Expected type %s, got %s", ErrorTypeDependency, err.Type)
	}
	if err.Severity != SeverityCritical {
		t.Errorf("Expected severity %s, got %s", SeverityCritical, err.Severity)
	}
	if err.Context.Resource != "testExecutor" {
		t.Errorf("Expected resource 'testExecutor', got '%s'", err.Context.Resource)
	}
	if err.Context.Component != "container" {
		t.Errorf("Expected component 'container', got '%s'", err.Context.Component)
	}
}

// TestNewLifecycleError tests lifecycle error creation
func TestNewLifecycleError(t *testing.T) {
	originalErr := errors.New("shutdown failed")
	err := NewLifecycleError("shutdown", originalErr)

	if err.Type != ErrorTypeLifecycle {
		t.Errorf("Expected type %s, got %s", ErrorTypeLifecycle, err.Type)
	}
	if err.Context.Operation != "shutdown" {
		t.Errorf("Expected operation 'shutdown', got '%s'", err.Context.Operation)
	}
	if err.Context.Component != "lifecycle" {
		t.Errorf("Expected component 'lifecycle', got '%s'", err.Context.Component)
	}
}

// TestNewInternalError tests internal error creation
func TestNewInternalError(t *testing.T) {
	originalErr := errors.New("database connection failed")
	err := NewInternalError("database", "connect", originalErr)

	if err.Type != ErrorTypeInternal {
		t.Errorf("Expected type %s, got %s", ErrorTypeInternal, err.Type)
	}
	if err.Severity != SeverityCritical {
		t.Errorf("Expected severity %s, got %s", SeverityCritical, err.Severity)
	}
	if err.UserSafe {
		t.Error("Expected internal error to not be user-safe")
	}
	if err.Context.Component != "database" {
		t.Errorf("Expected component 'database', got '%s'", err.Context.Component)
	}
	if err.Context.Operation != "connect" {
		t.Errorf("Expected operation 'connect', got '%s'", err.Context.Operation)
	}
}

// TestIsErrorType tests error type checking
func TestIsErrorType(t *testing.T) {
	configErr := NewError(ErrorTypeConfig, SeverityError, "config error")
	fileErr := NewError(ErrorTypeFileSystem, SeverityError, "file error")
	regularErr := errors.New("regular error")

	// Positive cases
	if !IsErrorType(configErr, ErrorTypeConfig) {
		t.Error("Expected IsErrorType to return true for matching type")
	}

	// Negative cases
	if IsErrorType(configErr, ErrorTypeFileSystem) {
		t.Error("Expected IsErrorType to return false for non-matching type")
	}
	if IsErrorType(fileErr, ErrorTypeConfig) {
		t.Error("Expected IsErrorType to return false for non-matching type")
	}
	if IsErrorType(regularErr, ErrorTypeConfig) {
		t.Error("Expected IsErrorType to return false for non-SentinelError")
	}
	if IsErrorType(nil, ErrorTypeConfig) {
		t.Error("Expected IsErrorType to return false for nil error")
	}
}

// TestIsErrorSeverity tests error severity checking
func TestIsErrorSeverity(t *testing.T) {
	criticalErr := NewError(ErrorTypeInternal, SeverityCritical, "critical error")
	warningErr := NewError(ErrorTypeValidation, SeverityWarning, "warning error")
	regularErr := errors.New("regular error")

	// Positive cases
	if !IsErrorSeverity(criticalErr, SeverityCritical) {
		t.Error("Expected IsErrorSeverity to return true for matching severity")
	}

	// Negative cases
	if IsErrorSeverity(criticalErr, SeverityWarning) {
		t.Error("Expected IsErrorSeverity to return false for non-matching severity")
	}
	if IsErrorSeverity(warningErr, SeverityCritical) {
		t.Error("Expected IsErrorSeverity to return false for non-matching severity")
	}
	if IsErrorSeverity(regularErr, SeverityCritical) {
		t.Error("Expected IsErrorSeverity to return false for non-SentinelError")
	}
	if IsErrorSeverity(nil, SeverityCritical) {
		t.Error("Expected IsErrorSeverity to return false for nil error")
	}
}

// TestGetErrorContext tests context extraction
func TestGetErrorContext(t *testing.T) {
	err := NewError(ErrorTypeConfig, SeverityError, "test error")
	err.Context.Component = "test_component"
	err.Context.Operation = "test_operation"

	context := GetErrorContext(err)
	if context == nil {
		t.Fatal("Expected context to be returned, got nil")
	}
	if context.Component != "test_component" {
		t.Errorf("Expected component 'test_component', got '%s'", context.Component)
	}
	if context.Operation != "test_operation" {
		t.Errorf("Expected operation 'test_operation', got '%s'", context.Operation)
	}

	// Test with non-SentinelError
	regularErr := errors.New("regular error")
	context = GetErrorContext(regularErr)
	if context != nil {
		t.Errorf("Expected nil context for non-SentinelError, got %v", context)
	}

	// Test with nil error
	context = GetErrorContext(nil)
	if context != nil {
		t.Errorf("Expected nil context for nil error, got %v", context)
	}
}

// TestSanitizeError tests error sanitization
func TestSanitizeError(t *testing.T) {
	// Test with nil error
	sanitized := SanitizeError(nil)
	if sanitized != nil {
		t.Errorf("Expected nil when sanitizing nil error, got %v", sanitized)
	}

	// Test with user-safe SentinelError
	userSafeErr := NewValidationError("field", "Invalid value")
	sanitized = SanitizeError(userSafeErr)
	if sanitized != userSafeErr {
		t.Error("Expected user-safe error to be returned as-is")
	}

	// Test with internal SentinelError
	internalErr := NewInternalError("database", "connect", errors.New("connection failed"))
	sanitized = SanitizeError(internalErr)
	if sentinelErr, ok := sanitized.(*SentinelError); ok {
		if sentinelErr.Message != "An internal error occurred" {
			t.Errorf("Expected sanitized message, got '%s'", sentinelErr.Message)
		}
		if sentinelErr.Type != ErrorTypeInternal {
			t.Errorf("Expected type to be preserved, got %s", sentinelErr.Type)
		}
	} else {
		t.Error("Expected sanitized error to be SentinelError")
	}

	// Test with regular error
	regularErr := errors.New("regular error")
	sanitized = SanitizeError(regularErr)
	if sentinelErr, ok := sanitized.(*SentinelError); ok {
		if sentinelErr.Type != ErrorTypeInternal {
			t.Errorf("Expected type %s, got %s", ErrorTypeInternal, sentinelErr.Type)
		}
		if sentinelErr.Message != "An error occurred" {
			t.Errorf("Expected generic message, got '%s'", sentinelErr.Message)
		}
	} else {
		t.Error("Expected sanitized error to be SentinelError")
	}
}

// TestStackCapture tests that stack traces are captured
func TestStackCapture(t *testing.T) {
	err := NewError(ErrorTypeConfig, SeverityError, "test error")

	if len(err.Stack) == 0 {
		t.Error("Expected stack trace to be captured")
	}

	// Check that the stack contains this test function
	found := false
	for _, frame := range err.Stack {
		if strings.Contains(frame.Function, "TestStackCapture") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected stack trace to contain test function")
	}
}

// TestErrorChaining tests error chaining with fmt.Errorf and errors.Is
func TestErrorChaining(t *testing.T) {
	originalErr := errors.New("original error")
	sentinelErr := WrapError(originalErr, ErrorTypeFileSystem, SeverityError, "wrapped error")
	chainedErr := fmt.Errorf("chained: %w", sentinelErr)

	// Test that errors.Is works through the chain
	if !errors.Is(chainedErr, sentinelErr) {
		t.Error("Expected errors.Is to find SentinelError in chain")
	}
	if !errors.Is(chainedErr, originalErr) {
		t.Error("Expected errors.Is to find original error in chain")
	}

	// Test that errors.As works
	var targetErr *SentinelError
	if !errors.As(chainedErr, &targetErr) {
		t.Error("Expected errors.As to extract SentinelError from chain")
	}
	if targetErr.Type != ErrorTypeFileSystem {
		t.Errorf("Expected extracted error type %s, got %s", ErrorTypeFileSystem, targetErr.Type)
	}
}
