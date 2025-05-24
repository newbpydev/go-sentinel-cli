// Package app provides application lifecycle management
package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// DefaultLifecycleManager implements the LifecycleManager interface
type DefaultLifecycleManager struct {
	mu            sync.RWMutex
	isRunning     bool
	shutdownHooks []func() error
	shutdownCh    chan struct{}
	signalCh      chan os.Signal
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewLifecycleManager creates a new lifecycle manager
func NewLifecycleManager() LifecycleManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &DefaultLifecycleManager{
		shutdownHooks: make([]func() error, 0),
		shutdownCh:    make(chan struct{}),
		signalCh:      make(chan os.Signal, 1),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Startup implements the LifecycleManager interface
func (lm *DefaultLifecycleManager) Startup(ctx context.Context) error {
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

// Shutdown implements the LifecycleManager interface
func (lm *DefaultLifecycleManager) Shutdown(ctx context.Context) error {
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
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 30*time.Second)
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

// IsRunning implements the LifecycleManager interface
func (lm *DefaultLifecycleManager) IsRunning() bool {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return lm.isRunning
}

// RegisterShutdownHook implements the LifecycleManager interface
func (lm *DefaultLifecycleManager) RegisterShutdownHook(hook func() error) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.shutdownHooks = append(lm.shutdownHooks, hook)
}

// handleSignals handles OS signals for graceful shutdown
func (lm *DefaultLifecycleManager) handleSignals() {
	select {
	case sig := <-lm.signalCh:
		fmt.Printf("\n🛑 Received signal %s, shutting down gracefully...\n", sig)

		// Create shutdown context with timeout
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
func (lm *DefaultLifecycleManager) executeShutdownHooks(ctx context.Context) error {
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

// Context returns the lifecycle context
func (lm *DefaultLifecycleManager) Context() context.Context {
	return lm.ctx
}

// ShutdownChannel returns a channel that closes when shutdown is initiated
func (lm *DefaultLifecycleManager) ShutdownChannel() <-chan struct{} {
	return lm.shutdownCh
}
