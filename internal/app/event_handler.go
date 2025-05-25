// Package app provides application event handling implementation
package app

import (
	"context"
	"log"
	"os"

	"github.com/newbpydev/go-sentinel/pkg/models"
)

// DefaultApplicationEventHandler implements the ApplicationEventHandler interface
type DefaultApplicationEventHandler struct {
	logger *log.Logger
	config *Configuration
}

// NewApplicationEventHandler creates a new application event handler
func NewApplicationEventHandler() ApplicationEventHandler {
	return &DefaultApplicationEventHandler{
		logger: log.New(os.Stderr, "[go-sentinel] ", log.LstdFlags|log.Lshortfile),
	}
}

// OnStartup is called when the application starts
func (h *DefaultApplicationEventHandler) OnStartup(ctx context.Context) error {
	h.logger.Printf("Application starting up...")

	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Log startup success
	h.logger.Printf("Application startup completed successfully")
	return nil
}

// OnShutdown is called when the application shuts down
func (h *DefaultApplicationEventHandler) OnShutdown(ctx context.Context) error {
	h.logger.Printf("Application shutting down...")

	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		// Still try to log shutdown even if context is cancelled
		h.logger.Printf("Application shutdown initiated due to context cancellation")
	default:
	}

	// Log shutdown completion
	h.logger.Printf("Application shutdown completed")
	return nil
}

// OnError is called when an error occurs
func (h *DefaultApplicationEventHandler) OnError(err error) {
	if err == nil {
		return
	}

	// Check if it's a SentinelError with additional context
	if sentinelErr, ok := err.(*models.SentinelError); ok {
		// Log with severity level
		switch sentinelErr.Severity {
		case models.SeverityCritical:
			h.logger.Printf("CRITICAL ERROR [%s]: %s", sentinelErr.Type, sentinelErr.Message)
		case models.SeverityError:
			h.logger.Printf("ERROR [%s]: %s", sentinelErr.Type, sentinelErr.Message)
		case models.SeverityWarning:
			h.logger.Printf("WARNING [%s]: %s", sentinelErr.Type, sentinelErr.Message)
		case models.SeverityInfo:
			h.logger.Printf("INFO [%s]: %s", sentinelErr.Type, sentinelErr.Message)
		default:
			h.logger.Printf("ERROR [%s]: %s", sentinelErr.Type, sentinelErr.Message)
		}

		// Log context if available
		if len(sentinelErr.Context.Metadata) > 0 {
			h.logger.Printf("Error context: %+v", sentinelErr.Context.Metadata)
		}

		// Log cause if available
		if sentinelErr.Cause != nil {
			h.logger.Printf("Caused by: %v", sentinelErr.Cause)
		}
	} else {
		// Log regular error
		h.logger.Printf("ERROR: %v", err)
	}
}

// OnConfigChanged is called when configuration changes
func (h *DefaultApplicationEventHandler) OnConfigChanged(config *Configuration) {
	if config == nil {
		h.logger.Printf("Configuration updated: <nil>")
		return
	}

	h.config = config

	// Log configuration changes
	h.logger.Printf("Configuration updated:")
	h.logger.Printf("  Colors: %t", config.Colors)
	h.logger.Printf("  Verbosity: %d", config.Verbosity)
	h.logger.Printf("  Watch enabled: %t", config.Watch.Enabled)

	if config.Watch.Enabled {
		h.logger.Printf("  Watch debounce: %s", config.Watch.Debounce)
		h.logger.Printf("  Watch run on start: %t", config.Watch.RunOnStart)
		h.logger.Printf("  Watch clear on rerun: %t", config.Watch.ClearOnRerun)
	}

	if len(config.Paths.IncludePatterns) > 0 {
		h.logger.Printf("  Include patterns: %v", config.Paths.IncludePatterns)
	}

	if len(config.Paths.ExcludePatterns) > 0 {
		h.logger.Printf("  Exclude patterns: %v", config.Paths.ExcludePatterns)
	}

	if len(config.Watch.IgnorePatterns) > 0 {
		h.logger.Printf("  Watch ignore patterns: %v", config.Watch.IgnorePatterns)
	}
}

// OnTestStart is called when a test starts (optional extension)
func (h *DefaultApplicationEventHandler) OnTestStart(testName string) {
	if h.config != nil && h.config.Verbosity > 1 {
		h.logger.Printf("Test started: %s", testName)
	}
}

// OnTestComplete is called when a test completes (optional extension)
func (h *DefaultApplicationEventHandler) OnTestComplete(testName string, success bool) {
	if h.config != nil && h.config.Verbosity > 1 {
		status := "PASS"
		if !success {
			status = "FAIL"
		}
		h.logger.Printf("Test completed: %s [%s]", testName, status)
	}
}

// OnWatchEvent is called when a file watch event occurs (optional extension)
func (h *DefaultApplicationEventHandler) OnWatchEvent(filePath string, eventType string) {
	if h.config != nil && h.config.Verbosity > 0 {
		h.logger.Printf("File watch event: %s [%s]", filePath, eventType)
	}
}

// SetLogger sets a custom logger
func (h *DefaultApplicationEventHandler) SetLogger(logger *log.Logger) {
	if logger != nil {
		h.logger = logger
	}
}

// SetVerbosity sets the verbosity level for logging
func (h *DefaultApplicationEventHandler) SetVerbosity(level int) {
	if h.config == nil {
		h.config = &Configuration{}
	}
	h.config.Verbosity = level
}

// GetLogger returns the current logger
func (h *DefaultApplicationEventHandler) GetLogger() *log.Logger {
	return h.logger
}

// LogDebug logs a debug message if verbosity is high enough
func (h *DefaultApplicationEventHandler) LogDebug(format string, args ...interface{}) {
	if h.config != nil && h.config.Verbosity > 1 {
		h.logger.Printf("[DEBUG] "+format, args...)
	}
}

// LogInfo logs an info message
func (h *DefaultApplicationEventHandler) LogInfo(format string, args ...interface{}) {
	h.logger.Printf("[INFO] "+format, args...)
}

// LogWarning logs a warning message
func (h *DefaultApplicationEventHandler) LogWarning(format string, args ...interface{}) {
	h.logger.Printf("[WARNING] "+format, args...)
}

// LogError logs an error message
func (h *DefaultApplicationEventHandler) LogError(format string, args ...interface{}) {
	h.logger.Printf("[ERROR] "+format, args...)
}

// Ensure DefaultApplicationEventHandler implements ApplicationEventHandler interface
var _ ApplicationEventHandler = (*DefaultApplicationEventHandler)(nil)
