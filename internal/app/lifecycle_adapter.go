// Package app provides adapter for lifecycle management to maintain clean package boundaries
package app

import (
	"context"

	"github.com/newbpydev/go-sentinel/internal/lifecycle"
)

// lifecycleManagerAdapter adapts the lifecycle package manager to the app package interface.
// This adapter pattern allows us to maintain compatibility while moving to proper architecture.
type lifecycleManagerAdapter struct {
	factory *LifecycleManagerFactory
	manager lifecycle.AppLifecycleManager
}

// Startup initializes all application components
func (a *lifecycleManagerAdapter) Startup(ctx context.Context) error {
	return a.manager.Startup(ctx)
}

// Shutdown gracefully stops all application components
func (a *lifecycleManagerAdapter) Shutdown(ctx context.Context) error {
	return a.manager.Shutdown(ctx)
}

// RegisterShutdownHook adds a function to be called during shutdown
func (a *lifecycleManagerAdapter) RegisterShutdownHook(hook func() error) {
	a.manager.RegisterShutdownHook(hook)
}

// LifecycleManagerFactory creates and manages lifecycle manager adapters.
// This factory follows dependency injection principles and maintains package boundaries.
type LifecycleManagerFactory struct {
	lifecycleFactory lifecycle.AppLifecycleManagerFactory
}

// NewLifecycleManagerFactory creates a new lifecycle manager factory.
func NewLifecycleManagerFactory() *LifecycleManagerFactory {
	return &LifecycleManagerFactory{
		lifecycleFactory: lifecycle.NewAppLifecycleManagerFactory(),
	}
}

// CreateLifecycleManagerWithDefaults creates a lifecycle manager adapter with default settings.
func (f *LifecycleManagerFactory) CreateLifecycleManagerWithDefaults() LifecycleManager {
	lifecycleManager := f.lifecycleFactory.CreateLifecycleManager()

	return &lifecycleManagerAdapter{
		factory: f,
		manager: lifecycleManager,
	}
}

// CreateLifecycleManagerWithContext creates a lifecycle manager adapter with custom context.
func (f *LifecycleManagerFactory) CreateLifecycleManagerWithContext(ctx context.Context) LifecycleManager {
	lifecycleManager := f.lifecycleFactory.CreateLifecycleManagerWithContext(ctx)

	return &lifecycleManagerAdapter{
		factory: f,
		manager: lifecycleManager,
	}
}

// NewLifecycleManager creates a new lifecycle manager using the adapter pattern.
// This follows dependency injection principles and maintains package boundaries.
func NewLifecycleManager() LifecycleManager {
	factory := NewLifecycleManagerFactory()
	return factory.CreateLifecycleManagerWithDefaults()
}

// Ensure lifecycleManagerAdapter implements LifecycleManager interface
var _ LifecycleManager = (*lifecycleManagerAdapter)(nil)
