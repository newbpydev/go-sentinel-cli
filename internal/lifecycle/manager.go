// Package lifecycle provides application lifecycle management implementation
package lifecycle

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// DefaultAppLifecycleManager implements the AppLifecycleManager interface.
// This implementation follows the Single Responsibility Principle by focusing only on lifecycle management.
type DefaultAppLifecycleManager struct {
	mu              sync.RWMutex
	isRunning       bool
	shutdownHooks   []func() error
	shutdownCh      chan struct{}
	signalCh        chan os.Signal
	ctx             context.Context
	cancel          context.CancelFunc
	shutdownTimeout time.Duration
}

// NewAppLifecycleManager creates a new application lifecycle manager with default configuration.
func NewAppLifecycleManager() AppLifecycleManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &DefaultAppLifecycleManager{
		shutdownHooks:   make([]func() error, 0),
		shutdownCh:      make(chan struct{}),
		signalCh:        make(chan os.Signal, 1),
		ctx:             ctx,
		cancel:          cancel,
		shutdownTimeout: 30 * time.Second, // Default timeout
	}
}

// NewAppLifecycleManagerWithContext creates a new lifecycle manager with custom context.
func NewAppLifecycleManagerWithContext(ctx context.Context) AppLifecycleManager {
	managerCtx, cancel := context.WithCancel(ctx)

	return &DefaultAppLifecycleManager{
		shutdownHooks:   make([]func() error, 0),
		shutdownCh:      make(chan struct{}),
		signalCh:        make(chan os.Signal, 1),
		ctx:             managerCtx,
		cancel:          cancel,
		shutdownTimeout: 30 * time.Second, // Default timeout
	}
}

// Startup initializes the lifecycle manager and sets up signal handling
func (lm *DefaultAppLifecycleManager) Startup(ctx context.Context) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if lm.isRunning {
		return fmt.Errorf("lifecycle manager is already running")
	}

	// Setup signal handling for graceful shutdown
	signal.Notify(lm.signalCh, os.Interrupt, syscall.SIGTERM)

	// Start signal handler goroutine
	go lm.handleSignals()

	lm.isRunning = true
	return nil
}

// Shutdown gracefully stops the lifecycle manager and executes shutdown hooks
func (lm *DefaultAppLifecycleManager) Shutdown(ctx context.Context) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if !lm.isRunning {
		return nil // Already shut down
	}

	// Stop signal notifications
	signal.Stop(lm.signalCh)

	// Cancel context
	lm.cancel()

	// Create timeout context for shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, lm.shutdownTimeout)
	defer shutdownCancel()

	// Execute shutdown hooks
	if err := lm.executeShutdownHooks(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown hooks failed: %w", err)
	}

	// Close shutdown channel
	close(lm.shutdownCh)

	lm.isRunning = false
	return nil
}

// IsRunning returns whether the lifecycle manager is currently running
func (lm *DefaultAppLifecycleManager) IsRunning() bool {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return lm.isRunning
}

// RegisterShutdownHook adds a function to be called during shutdown
func (lm *DefaultAppLifecycleManager) RegisterShutdownHook(hook func() error) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.shutdownHooks = append(lm.shutdownHooks, hook)
}

// Context returns the lifecycle context
func (lm *DefaultAppLifecycleManager) Context() context.Context {
	return lm.ctx
}

// ShutdownChannel returns a channel that closes when shutdown is initiated
func (lm *DefaultAppLifecycleManager) ShutdownChannel() <-chan struct{} {
	return lm.shutdownCh
}

// handleSignals handles OS signals for graceful shutdown
func (lm *DefaultAppLifecycleManager) handleSignals() {
	select {
	case sig := <-lm.signalCh:
		fmt.Printf("\n🛑 Received signal %s, shutting down gracefully...\n", sig)

		// Create shutdown context with timeout
		shutdownCtx, cancel := context.WithTimeout(context.Background(), lm.shutdownTimeout)
		defer cancel()

		// Shutdown the application
		if err := lm.Shutdown(shutdownCtx); err != nil {
			fmt.Printf("Error during shutdown: %v\n", err)
			os.Exit(1)
		}

		os.Exit(0)

	case <-lm.ctx.Done():
		// Context was cancelled, normal shutdown
		return
	}
}

// executeShutdownHooks executes all registered shutdown hooks
func (lm *DefaultAppLifecycleManager) executeShutdownHooks(ctx context.Context) error {
	// Execute hooks in reverse order (LIFO)
	for i := len(lm.shutdownHooks) - 1; i >= 0; i-- {
		hook := lm.shutdownHooks[i]

		// Execute hook with timeout
		done := make(chan error, 1)
		go func() {
			done <- hook()
		}()

		select {
		case err := <-done:
			if err != nil {
				return fmt.Errorf("shutdown hook failed: %w", err)
			}
		case <-ctx.Done():
			return fmt.Errorf("shutdown hook timed out: %w", ctx.Err())
		}
	}

	return nil
}

// SetShutdownTimeout sets the timeout for shutdown operations
func (lm *DefaultAppLifecycleManager) SetShutdownTimeout(timeout time.Duration) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.shutdownTimeout = timeout
}

// Ensure DefaultAppLifecycleManager implements AppLifecycleManager interface
var _ AppLifecycleManager = (*DefaultAppLifecycleManager)(nil)
