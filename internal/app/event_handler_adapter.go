// Package app provides adapter for event handling to maintain clean package boundaries
package app

import (
	"context"

	"github.com/newbpydev/go-sentinel/internal/events"
)

// eventHandlerAdapter adapts the events package handler to the app package interface.
// This adapter pattern allows us to maintain compatibility while moving to proper architecture.
type eventHandlerAdapter struct {
	factory *EventHandlerFactory
	handler events.AppEventHandler
}

// OnStartup is called when the application starts
func (a *eventHandlerAdapter) OnStartup(ctx context.Context) error {
	return a.handler.OnStartup(ctx)
}

// OnShutdown is called when the application shuts down
func (a *eventHandlerAdapter) OnShutdown(ctx context.Context) error {
	return a.handler.OnShutdown(ctx)
}

// OnConfigChanged is called when configuration changes
func (a *eventHandlerAdapter) OnConfigChanged(config *Configuration) {
	// Convert app configuration to events configuration
	eventsConfig := a.factory.convertToEventsConfig(config)
	a.handler.OnConfigChanged(eventsConfig)
}

// OnError is called when an error occurs
func (a *eventHandlerAdapter) OnError(err error) {
	a.handler.OnError(err)
}

// EventHandlerFactory creates and manages event handler adapters.
// This factory follows dependency injection principles and maintains package boundaries.
type EventHandlerFactory struct {
	eventsFactory events.AppEventHandlerFactory
}

// NewEventHandlerFactory creates a new event handler factory.
func NewEventHandlerFactory() *EventHandlerFactory {
	return &EventHandlerFactory{
		eventsFactory: events.NewAppEventHandlerFactory(),
	}
}

// CreateEventHandlerWithDefaults creates an event handler adapter with default settings.
func (f *EventHandlerFactory) CreateEventHandlerWithDefaults() ApplicationEventHandler {
	eventsHandler := f.eventsFactory.CreateEventHandler()

	return &eventHandlerAdapter{
		factory: f,
		handler: eventsHandler,
	}
}

// convertToEventsConfig converts app Configuration to events AppConfig.
// This conversion maintains separation of concerns between packages.
func (f *EventHandlerFactory) convertToEventsConfig(appConfig *Configuration) *events.AppConfig {
	if appConfig == nil {
		return nil
	}

	return &events.AppConfig{
		Colors:    appConfig.Colors,
		Verbosity: appConfig.Verbosity,
		Watch: events.AppWatchConfig{
			Enabled:      appConfig.Watch.Enabled,
			Debounce:     appConfig.Watch.Debounce,
			RunOnStart:   appConfig.Watch.RunOnStart,
			ClearOnRerun: appConfig.Watch.ClearOnRerun,
		},
		Paths: events.AppPathsConfig{
			IncludePatterns: appConfig.Paths.IncludePatterns,
			ExcludePatterns: appConfig.Paths.ExcludePatterns,
			IgnorePatterns:  appConfig.Watch.IgnorePatterns,
		},
	}
}

// NewApplicationEventHandler creates a new application event handler using the adapter pattern.
// This follows dependency injection principles and maintains package boundaries.
func NewApplicationEventHandler() ApplicationEventHandler {
	factory := NewEventHandlerFactory()
	return factory.CreateEventHandlerWithDefaults()
}

// Ensure eventHandlerAdapter implements ApplicationEventHandler interface
var _ ApplicationEventHandler = (*eventHandlerAdapter)(nil)
