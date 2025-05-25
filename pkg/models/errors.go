// Package models provides shared error types and error handling utilities
package models

import (
	"fmt"
	"runtime"
	"strings"
)

// ErrorType represents the category of error
type ErrorType string

const (
	// Configuration errors
	ErrorTypeConfig ErrorType = "CONFIG"

	// File system and I/O errors
	ErrorTypeFileSystem ErrorType = "FILESYSTEM"

	// Test execution errors
	ErrorTypeTestExecution ErrorType = "TEST_EXECUTION"

	// Watch system errors
	ErrorTypeWatch ErrorType = "WATCH"

	// Dependency injection errors
	ErrorTypeDependency ErrorType = "DEPENDENCY"

	// Application lifecycle errors
	ErrorTypeLifecycle ErrorType = "LIFECYCLE"

	// Validation errors
	ErrorTypeValidation ErrorType = "VALIDATION"

	// Network/external service errors
	ErrorTypeExternal ErrorType = "EXTERNAL"

	// Internal system errors
	ErrorTypeInternal ErrorType = "INTERNAL"
)

// ErrorSeverity represents how critical an error is
type ErrorSeverity string

const (
	SeverityInfo     ErrorSeverity = "INFO"
	SeverityWarning  ErrorSeverity = "WARNING"
	SeverityError    ErrorSeverity = "ERROR"
	SeverityCritical ErrorSeverity = "CRITICAL"
)

// SentinelError is the base error type for all application errors
type SentinelError struct {
	Type     ErrorType     `json:"type"`
	Severity ErrorSeverity `json:"severity"`
	Message  string        `json:"message"`
	Cause    error         `json:"cause,omitempty"`
	Context  ErrorContext  `json:"context"`
	Stack    []StackFrame  `json:"stack,omitempty"`
	UserSafe bool          `json:"userSafe"` // Whether safe to show to end users
}

// ErrorContext provides additional context about where/when the error occurred
type ErrorContext struct {
	Operation string            `json:"operation"` // What operation was being performed
	Component string            `json:"component"` // Which component generated the error
	Resource  string            `json:"resource"`  // What resource was involved (file, package, etc.)
	Metadata  map[string]string `json:"metadata"`  // Additional context-specific data
	RequestID string            `json:"requestId"` // For tracing across operations
	UserID    string            `json:"userId"`    // For user-specific operations
}

// StackFrame represents a single frame in the call stack
type StackFrame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

// Error implements the error interface
func (e *SentinelError) Error() string {
	var parts []string

	// Add type and severity prefix
	parts = append(parts, fmt.Sprintf("[%s:%s]", e.Type, e.Severity))

	// Add component if available
	if e.Context.Component != "" {
		parts = append(parts, fmt.Sprintf("(%s)", e.Context.Component))
	}

	// Add main message
	parts = append(parts, e.Message)

	// Add resource context if available
	if e.Context.Resource != "" {
		parts = append(parts, fmt.Sprintf("resource=%s", e.Context.Resource))
	}

	// Add operation context if available
	if e.Context.Operation != "" {
		parts = append(parts, fmt.Sprintf("operation=%s", e.Context.Operation))
	}

	return strings.Join(parts, " ")
}

// Unwrap returns the underlying cause for error unwrapping
func (e *SentinelError) Unwrap() error {
	return e.Cause
}

// Is implements error comparison for errors.Is
func (e *SentinelError) Is(target error) bool {
	if t, ok := target.(*SentinelError); ok {
		return e.Type == t.Type && e.Message == t.Message
	}
	return false
}

// UserMessage returns a sanitized message safe for end users
func (e *SentinelError) UserMessage() string {
	if !e.UserSafe {
		return getGenericErrorMessage(e.Type)
	}
	return e.Message
}

// getGenericErrorMessage returns a generic user-safe message for the given error type
func getGenericErrorMessage(errorType ErrorType) string {
	switch errorType {
	case ErrorTypeConfig:
		return "Configuration error occurred"
	case ErrorTypeFileSystem:
		return "File system error occurred"
	case ErrorTypeTestExecution:
		return "Test execution failed"
	case ErrorTypeWatch:
		return "File watching error occurred"
	case ErrorTypeDependency:
		return "Dependency resolution failed"
	case ErrorTypeLifecycle:
		return "Application lifecycle error"
	case ErrorTypeValidation:
		return "Validation failed"
	case ErrorTypeExternal:
		return "External service error"
	default:
		return "An internal error occurred"
	}
}

// WithContext adds additional context to the error
func (e *SentinelError) WithContext(key, value string) *SentinelError {
	if e.Context.Metadata == nil {
		e.Context.Metadata = make(map[string]string)
	}
	e.Context.Metadata[key] = value
	return e
}

// WithRequestID adds a request ID for tracing
func (e *SentinelError) WithRequestID(requestID string) *SentinelError {
	e.Context.RequestID = requestID
	return e
}

// WithUserID adds a user ID for user-specific operations
func (e *SentinelError) WithUserID(userID string) *SentinelError {
	e.Context.UserID = userID
	return e
}

// captureStack captures the current call stack
func captureStack(skip int) []StackFrame {
	var frames []StackFrame

	// Capture up to 10 stack frames
	for i := skip; i < skip+10; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		frames = append(frames, StackFrame{
			Function: fn.Name(),
			File:     file,
			Line:     line,
		})
	}

	return frames
}

// NewError creates a new SentinelError with the specified type and message
func NewError(errorType ErrorType, severity ErrorSeverity, message string) *SentinelError {
	return &SentinelError{
		Type:     errorType,
		Severity: severity,
		Message:  message,
		Context:  ErrorContext{},
		Stack:    captureStack(2), // Skip NewError and caller
		UserSafe: severity == SeverityInfo || severity == SeverityWarning,
	}
}

// WrapError wraps an existing error with additional context
func WrapError(err error, errorType ErrorType, severity ErrorSeverity, message string) *SentinelError {
	if err == nil {
		return nil
	}

	return &SentinelError{
		Type:     errorType,
		Severity: severity,
		Message:  message,
		Cause:    err,
		Context:  ErrorContext{},
		Stack:    captureStack(2), // Skip WrapError and caller
		UserSafe: severity == SeverityInfo || severity == SeverityWarning,
	}
}

// NewConfigError creates a configuration-related error
func NewConfigError(message string, userSafe bool) *SentinelError {
	severity := SeverityError
	if userSafe {
		severity = SeverityWarning
	}

	err := NewError(ErrorTypeConfig, severity, message)
	err.UserSafe = userSafe
	err.Context.Component = "config"
	return err
}

// NewValidationError creates a validation error (always user-safe)
func NewValidationError(field, message string) *SentinelError {
	err := NewError(ErrorTypeValidation, SeverityWarning, message)
	err.UserSafe = true
	err.Context.Component = "validation"
	err.Context.Resource = field
	return err
}

// NewFileSystemError creates a file system error
func NewFileSystemError(operation, path string, cause error) *SentinelError {
	message := fmt.Sprintf("file system operation failed: %s", operation)
	err := WrapError(cause, ErrorTypeFileSystem, SeverityError, message)
	err.Context.Operation = operation
	err.Context.Resource = path
	err.Context.Component = "filesystem"
	return err
}

// NewTestExecutionError creates a test execution error
func NewTestExecutionError(testPath string, cause error) *SentinelError {
	message := "test execution failed"
	err := WrapError(cause, ErrorTypeTestExecution, SeverityError, message)
	err.Context.Operation = "test_execution"
	err.Context.Resource = testPath
	err.Context.Component = "test_runner"
	return err
}

// NewWatchError creates a watch system error
func NewWatchError(operation, path string, cause error) *SentinelError {
	message := fmt.Sprintf("watch operation failed: %s", operation)
	err := WrapError(cause, ErrorTypeWatch, SeverityError, message)
	err.Context.Operation = operation
	err.Context.Resource = path
	err.Context.Component = "watcher"
	return err
}

// NewDependencyError creates a dependency injection error
func NewDependencyError(component string, cause error) *SentinelError {
	message := fmt.Sprintf("dependency resolution failed for: %s", component)
	err := WrapError(cause, ErrorTypeDependency, SeverityCritical, message)
	err.Context.Operation = "dependency_resolution"
	err.Context.Resource = component
	err.Context.Component = "container"
	return err
}

// NewLifecycleError creates an application lifecycle error
func NewLifecycleError(operation string, cause error) *SentinelError {
	message := fmt.Sprintf("lifecycle operation failed: %s", operation)
	err := WrapError(cause, ErrorTypeLifecycle, SeverityError, message)
	err.Context.Operation = operation
	err.Context.Component = "lifecycle"
	return err
}

// NewInternalError creates an internal system error (never user-safe)
func NewInternalError(component, operation string, cause error) *SentinelError {
	message := "internal system error occurred"
	err := WrapError(cause, ErrorTypeInternal, SeverityCritical, message)
	err.Context.Operation = operation
	err.Context.Component = component
	err.UserSafe = false
	return err
}

// IsErrorType checks if an error is of a specific type
func IsErrorType(err error, errorType ErrorType) bool {
	if sentinelErr, ok := err.(*SentinelError); ok {
		return sentinelErr.Type == errorType
	}
	return false
}

// IsErrorSeverity checks if an error has a specific severity
func IsErrorSeverity(err error, severity ErrorSeverity) bool {
	if sentinelErr, ok := err.(*SentinelError); ok {
		return sentinelErr.Severity == severity
	}
	return false
}

// GetErrorContext extracts context from a SentinelError
func GetErrorContext(err error) *ErrorContext {
	if sentinelErr, ok := err.(*SentinelError); ok {
		return &sentinelErr.Context
	}
	return nil
}

// SanitizeError returns a user-safe version of any error
func SanitizeError(err error) error {
	if err == nil {
		return nil
	}

	if sentinelErr, ok := err.(*SentinelError); ok {
		if sentinelErr.UserSafe {
			return sentinelErr
		}
		// Return sanitized version
		return NewError(sentinelErr.Type, sentinelErr.Severity, sentinelErr.UserMessage())
	}

	// For non-SentinelError, return generic message
	return NewError(ErrorTypeInternal, SeverityError, "An error occurred")
}
