// Package lifecycle provides factory for creating lifecycle managers
package lifecycle

import (
	"context"
	"time"
)

// DefaultAppLifecycleManagerFactory implements the AppLifecycleManagerFactory interface.
// This factory follows the Factory pattern and dependency injection principles.
type DefaultAppLifecycleManagerFactory struct {
	// Dependencies for creating lifecycle managers
	defaultContext context.Context
	defaultTimeout time.Duration
}

// NewAppLifecycleManagerFactory creates a new lifecycle manager factory.
func NewAppLifecycleManagerFactory() AppLifecycleManagerFactory {
	return &DefaultAppLifecycleManagerFactory{
		defaultContext: context.Background(),
		defaultTimeout: 30 * time.Second,
	}
}

// NewAppLifecycleManagerFactoryWithDependencies creates a factory with injected dependencies.
func NewAppLifecycleManagerFactoryWithDependencies(deps AppLifecycleManagerDependencies) AppLifecycleManagerFactory {
	timeout := 30 * time.Second
	if deps.ShutdownTimeout != "" {
		if parsed, err := time.ParseDuration(deps.ShutdownTimeout); err == nil {
			timeout = parsed
		}
	}

	ctx := deps.Context
	if ctx == nil {
		ctx = context.Background()
	}

	return &DefaultAppLifecycleManagerFactory{
		defaultContext: ctx,
		defaultTimeout: timeout,
	}
}

// CreateLifecycleManager creates a new lifecycle manager with default configuration.
func (f *DefaultAppLifecycleManagerFactory) CreateLifecycleManager() AppLifecycleManager {
	manager := NewAppLifecycleManager()

	// Apply factory defaults if we have a concrete implementation
	if concrete, ok := manager.(*DefaultAppLifecycleManager); ok {
		concrete.SetShutdownTimeout(f.defaultTimeout)
	}

	return manager
}

// CreateLifecycleManagerWithContext creates a new lifecycle manager with a custom context.
func (f *DefaultAppLifecycleManagerFactory) CreateLifecycleManagerWithContext(ctx context.Context) AppLifecycleManager {
	manager := NewAppLifecycleManagerWithContext(ctx)

	// Apply factory defaults if we have a concrete implementation
	if concrete, ok := manager.(*DefaultAppLifecycleManager); ok {
		concrete.SetShutdownTimeout(f.defaultTimeout)
	}

	return manager
}

// CreateLifecycleManagerWithDefaults creates a lifecycle manager using factory defaults.
// This method demonstrates the Factory pattern providing sensible defaults.
func (f *DefaultAppLifecycleManagerFactory) CreateLifecycleManagerWithDefaults() AppLifecycleManager {
	return f.CreateLifecycleManagerWithContext(f.defaultContext)
}

// Ensure DefaultAppLifecycleManagerFactory implements AppLifecycleManagerFactory interface
var _ AppLifecycleManagerFactory = (*DefaultAppLifecycleManagerFactory)(nil)
