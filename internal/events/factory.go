// Package events provides factory for creating event handlers
package events

import (
	"log"
)

// DefaultAppEventHandlerFactory implements the AppEventHandlerFactory interface.
// This factory follows the Factory pattern and dependency injection principles.
type DefaultAppEventHandlerFactory struct {
	// Dependencies for creating event handlers
	defaultLogger *log.Logger
}

// NewAppEventHandlerFactory creates a new event handler factory.
func NewAppEventHandlerFactory() AppEventHandlerFactory {
	return &DefaultAppEventHandlerFactory{
		defaultLogger: nil, // Will use default logger in handlers
	}
}

// NewAppEventHandlerFactoryWithDependencies creates a factory with injected dependencies.
func NewAppEventHandlerFactoryWithDependencies(deps AppEventHandlerDependencies) AppEventHandlerFactory {
	return &DefaultAppEventHandlerFactory{
		defaultLogger: deps.Logger,
	}
}

// CreateEventHandler creates a new event handler with default configuration.
func (f *DefaultAppEventHandlerFactory) CreateEventHandler() AppEventHandler {
	if f.defaultLogger != nil {
		return NewAppEventHandlerWithLogger(f.defaultLogger)
	}
	return NewAppEventHandler()
}

// CreateEventHandlerWithLogger creates a new event handler with a custom logger.
func (f *DefaultAppEventHandlerFactory) CreateEventHandlerWithLogger(logger *log.Logger) AppEventHandler {
	return NewAppEventHandlerWithLogger(logger)
}

// CreateEventHandlerWithDefaults creates an event handler using factory defaults.
// This method demonstrates the Factory pattern providing sensible defaults.
func (f *DefaultAppEventHandlerFactory) CreateEventHandlerWithDefaults() AppEventHandler {
	handler := f.CreateEventHandler()

	// Set default verbosity if none set
	handler.SetVerbosity(0)

	return handler
}

// Ensure DefaultAppEventHandlerFactory implements AppEventHandlerFactory interface
var _ AppEventHandlerFactory = (*DefaultAppEventHandlerFactory)(nil)
