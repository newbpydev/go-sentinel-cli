package core

import (
	"fmt"
	"time"
)

// Error types for different failure scenarios

// ConfigError represents configuration-related errors
type ConfigError struct {
	Field   string
	Value   interface{}
	Message string
	Cause   error
}

func (e *ConfigError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("config error for field '%s': %s (caused by: %v)", e.Field, e.Message, e.Cause)
	}
	return fmt.Sprintf("config error for field '%s': %s", e.Field, e.Message)
}

func (e *ConfigError) Unwrap() error {
	return e.Cause
}

// NewConfigError creates a new configuration error
func NewConfigError(field string, value interface{}, message string, cause error) *ConfigError {
	return &ConfigError{
		Field:   field,
		Value:   value,
		Message: message,
		Cause:   cause,
	}
}

// TestExecutionError represents test execution failures
type TestExecutionError struct {
	Target   TestTarget
	Command  string
	ExitCode int
	Output   string
	Duration time.Duration
	Cause    error
}

func (e *TestExecutionError) Error() string {
	msg := fmt.Sprintf("test execution failed for target '%s'", e.Target.Path)
	if e.ExitCode != 0 {
		msg += fmt.Sprintf(" with exit code %d", e.ExitCode)
	}
	if e.Duration > 0 {
		msg += fmt.Sprintf(" after %v", e.Duration)
	}
	if e.Cause != nil {
		msg += fmt.Sprintf(" (caused by: %v)", e.Cause)
	}
	return msg
}

func (e *TestExecutionError) Unwrap() error {
	return e.Cause
}

// NewTestExecutionError creates a new test execution error
func NewTestExecutionError(target TestTarget, command string, exitCode int, output string, duration time.Duration, cause error) *TestExecutionError {
	return &TestExecutionError{
		Target:   target,
		Command:  command,
		ExitCode: exitCode,
		Output:   output,
		Duration: duration,
		Cause:    cause,
	}
}

// FileWatchError represents file watching errors
type FileWatchError struct {
	Path      string
	Operation string
	Message   string
	Cause     error
}

func (e *FileWatchError) Error() string {
	msg := fmt.Sprintf("file watch error for path '%s'", e.Path)
	if e.Operation != "" {
		msg += fmt.Sprintf(" during operation '%s'", e.Operation)
	}
	if e.Message != "" {
		msg += fmt.Sprintf(": %s", e.Message)
	}
	if e.Cause != nil {
		msg += fmt.Sprintf(" (caused by: %v)", e.Cause)
	}
	return msg
}

func (e *FileWatchError) Unwrap() error {
	return e.Cause
}

// NewFileWatchError creates a new file watch error
func NewFileWatchError(path, operation, message string, cause error) *FileWatchError {
	return &FileWatchError{
		Path:      path,
		Operation: operation,
		Message:   message,
		Cause:     cause,
	}
}

// CacheError represents caching-related errors
type CacheError struct {
	Operation string
	Key       string
	Message   string
	Cause     error
}

func (e *CacheError) Error() string {
	msg := fmt.Sprintf("cache error during operation '%s'", e.Operation)
	if e.Key != "" {
		msg += fmt.Sprintf(" for key '%s'", e.Key)
	}
	if e.Message != "" {
		msg += fmt.Sprintf(": %s", e.Message)
	}
	if e.Cause != nil {
		msg += fmt.Sprintf(" (caused by: %v)", e.Cause)
	}
	return msg
}

func (e *CacheError) Unwrap() error {
	return e.Cause
}

// NewCacheError creates a new cache error
func NewCacheError(operation, key, message string, cause error) *CacheError {
	return &CacheError{
		Operation: operation,
		Key:       key,
		Message:   message,
		Cause:     cause,
	}
}

// RenderError represents rendering/output errors
type RenderError struct {
	Component string
	Message   string
	Cause     error
}

func (e *RenderError) Error() string {
	msg := fmt.Sprintf("render error in component '%s'", e.Component)
	if e.Message != "" {
		msg += fmt.Sprintf(": %s", e.Message)
	}
	if e.Cause != nil {
		msg += fmt.Sprintf(" (caused by: %v)", e.Cause)
	}
	return msg
}

func (e *RenderError) Unwrap() error {
	return e.Cause
}

// NewRenderError creates a new render error
func NewRenderError(component, message string, cause error) *RenderError {
	return &RenderError{
		Component: component,
		Message:   message,
		Cause:     cause,
	}
}

// ValidationError represents validation failures
type ValidationError struct {
	Field   string
	Value   interface{}
	Rule    string
	Message string
}

func (e *ValidationError) Error() string {
	msg := fmt.Sprintf("validation failed for field '%s'", e.Field)
	if e.Rule != "" {
		msg += fmt.Sprintf(" (rule: %s)", e.Rule)
	}
	if e.Message != "" {
		msg += fmt.Sprintf(": %s", e.Message)
	}
	return msg
}

// NewValidationError creates a new validation error
func NewValidationError(field string, value interface{}, rule, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Rule:    rule,
		Message: message,
	}
}

// TimeoutError represents timeout failures
type TimeoutError struct {
	Operation string
	Timeout   time.Duration
	Elapsed   time.Duration
	Cause     error
}

func (e *TimeoutError) Error() string {
	msg := fmt.Sprintf("timeout during operation '%s' after %v", e.Operation, e.Elapsed)
	if e.Timeout > 0 {
		msg += fmt.Sprintf(" (timeout: %v)", e.Timeout)
	}
	if e.Cause != nil {
		msg += fmt.Sprintf(" (caused by: %v)", e.Cause)
	}
	return msg
}

func (e *TimeoutError) Unwrap() error {
	return e.Cause
}

// NewTimeoutError creates a new timeout error
func NewTimeoutError(operation string, timeout, elapsed time.Duration, cause error) *TimeoutError {
	return &TimeoutError{
		Operation: operation,
		Timeout:   timeout,
		Elapsed:   elapsed,
		Cause:     cause,
	}
}

// DependencyError represents dependency-related errors
type DependencyError struct {
	Component   string
	Dependency  string
	Message     string
	Suggestions []string
	Cause       error
}

func (e *DependencyError) Error() string {
	msg := fmt.Sprintf("dependency error in component '%s'", e.Component)
	if e.Dependency != "" {
		msg += fmt.Sprintf(" for dependency '%s'", e.Dependency)
	}
	if e.Message != "" {
		msg += fmt.Sprintf(": %s", e.Message)
	}
	if len(e.Suggestions) > 0 {
		msg += fmt.Sprintf(" (suggestions: %v)", e.Suggestions)
	}
	if e.Cause != nil {
		msg += fmt.Sprintf(" (caused by: %v)", e.Cause)
	}
	return msg
}

func (e *DependencyError) Unwrap() error {
	return e.Cause
}

// NewDependencyError creates a new dependency error
func NewDependencyError(component, dependency, message string, suggestions []string, cause error) *DependencyError {
	return &DependencyError{
		Component:   component,
		Dependency:  dependency,
		Message:     message,
		Suggestions: suggestions,
		Cause:       cause,
	}
}

// Error severity levels
type ErrorSeverity int

const (
	SeverityInfo ErrorSeverity = iota
	SeverityWarning
	SeverityError
	SeverityFatal
)

func (s ErrorSeverity) String() string {
	switch s {
	case SeverityInfo:
		return "info"
	case SeverityWarning:
		return "warning"
	case SeverityError:
		return "error"
	case SeverityFatal:
		return "fatal"
	default:
		return "unknown"
	}
}

// ContextualError adds context and severity to errors
type ContextualError struct {
	Severity ErrorSeverity
	Context  map[string]interface{}
	Message  string
	Cause    error
}

func (e *ContextualError) Error() string {
	msg := fmt.Sprintf("[%s] %s", e.Severity, e.Message)
	if e.Cause != nil {
		msg += fmt.Sprintf(" (caused by: %v)", e.Cause)
	}
	return msg
}

func (e *ContextualError) Unwrap() error {
	return e.Cause
}

// NewContextualError creates a new contextual error
func NewContextualError(severity ErrorSeverity, message string, context map[string]interface{}, cause error) *ContextualError {
	return &ContextualError{
		Severity: severity,
		Context:  context,
		Message:  message,
		Cause:    cause,
	}
}
