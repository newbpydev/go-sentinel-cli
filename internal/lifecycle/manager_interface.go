// Package lifecycle provides application lifecycle management implementation.
// This package follows the Single Responsibility Principle by focusing only on lifecycle management.
package lifecycle

import (
	"context"
)

// AppLifecycleManager interface for application lifecycle management in the lifecycle package.
// This interface is defined in the lifecycle package and is implemented by lifecycle managers.
type AppLifecycleManager interface {
	// Core lifecycle methods
	Startup(ctx context.Context) error
	Shutdown(ctx context.Context) error
	IsRunning() bool

	// Shutdown hook management
	RegisterShutdownHook(hook func() error)

	// Context and channel access
	Context() context.Context
	ShutdownChannel() <-chan struct{}
}

// AppLifecycleManagerFactory interface for creating lifecycle managers.
type AppLifecycleManagerFactory interface {
	CreateLifecycleManager() AppLifecycleManager
	CreateLifecycleManagerWithContext(ctx context.Context) AppLifecycleManager
}

// AppLifecycleManagerDependencies represents dependencies for lifecycle manager creation.
type AppLifecycleManagerDependencies struct {
	Context         context.Context
	ShutdownTimeout string // Duration string like "30s"
}
