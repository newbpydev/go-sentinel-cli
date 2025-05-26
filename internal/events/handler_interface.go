// Package events provides application event handling implementation.
// This package follows the Single Responsibility Principle by focusing only on event handling logic.
package events

import (
	"context"
	"log"
)

// AppEventHandler interface for application event handling in the events package.
// This interface is defined in the events package and is implemented by event handlers.
type AppEventHandler interface {
	// Core event handling methods
	OnStartup(ctx context.Context) error
	OnShutdown(ctx context.Context) error
	OnError(err error)
	OnConfigChanged(config *AppConfig)

	// Extended event handling methods
	OnTestStart(testName string)
	OnTestComplete(testName string, success bool)
	OnWatchEvent(filePath string, eventType string)

	// Logger management
	SetLogger(logger *log.Logger)
	SetVerbosity(level int)
	GetLogger() *log.Logger

	// Logging utilities
	LogDebug(format string, args ...interface{})
	LogInfo(format string, args ...interface{})
	LogWarning(format string, args ...interface{})
	LogError(format string, args ...interface{})
}

// AppConfig represents application configuration for event handling.
// This is a lightweight config structure specific to the events package.
type AppConfig struct {
	Colors    bool
	Verbosity int
	Watch     AppWatchConfig
	Paths     AppPathsConfig
}

// AppWatchConfig represents watch configuration for event logging.
type AppWatchConfig struct {
	Enabled      bool
	Debounce     string
	RunOnStart   bool
	ClearOnRerun bool
}

// AppPathsConfig represents paths configuration for event logging.
type AppPathsConfig struct {
	IncludePatterns []string
	ExcludePatterns []string
	IgnorePatterns  []string
}

// AppEventHandlerFactory interface for creating event handlers.
type AppEventHandlerFactory interface {
	CreateEventHandler() AppEventHandler
	CreateEventHandlerWithLogger(logger *log.Logger) AppEventHandler
}

// AppEventHandlerDependencies represents dependencies for event handler creation.
type AppEventHandlerDependencies struct {
	Logger    *log.Logger
	Verbosity int
}
