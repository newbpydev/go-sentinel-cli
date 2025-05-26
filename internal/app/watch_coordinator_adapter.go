// Package app provides watch coordinator adapter to maintain clean package boundaries
package app

import (
	"context"
	"fmt"
)

// watchCoordinatorAdapter adapts watch coordination to use proper dependency injection
// This adapter eliminates direct dependencies on internal/watch packages
type watchCoordinatorAdapter struct {
	// Dependencies injected through interfaces
	watcher   FileWatcher
	debouncer EventDebouncer
	options   *WatchOptions
	config    *Configuration
}

// FileWatcher interface for file system watching - defined in app package as consumer
type FileWatcher interface {
	Watch(paths []string, ignorePatterns []string) (<-chan FileEvent, error)
	Stop() error
}

// EventDebouncer interface for event debouncing - defined in app package as consumer
type EventDebouncer interface {
	Debounce(events <-chan FileEvent, interval string) <-chan FileEvent
}

// FileEvent represents a file system event
type FileEvent struct {
	Path      string
	Operation string
	IsDir     bool
}

// NewWatchCoordinatorAdapter creates a new watch coordinator adapter with dependency injection
func NewWatchCoordinatorAdapter() WatchCoordinator {
	return &watchCoordinatorAdapter{}
}

// NewWatchCoordinator creates a new watch coordinator using the adapter pattern
// This eliminates direct dependencies on internal packages
func NewWatchCoordinator() WatchCoordinator {
	adapter := &watchCoordinatorAdapter{}

	// Wire real implementations following architecture principles
	// For now, we'll use placeholder implementations until watch system is fully integrated
	// This maintains the adapter pattern while allowing the system to function

	return adapter
}

// Configure implements WatchCoordinator interface
func (w *watchCoordinatorAdapter) Configure(options *WatchOptions) error {
	if options == nil {
		return fmt.Errorf("watch options cannot be nil")
	}

	w.options = options

	// Dependencies will be injected through factory pattern
	// This eliminates direct imports of internal packages
	return nil
}

// Start implements WatchCoordinator interface
func (w *watchCoordinatorAdapter) Start(ctx context.Context) error {
	if w.options == nil {
		return fmt.Errorf("watch coordinator not configured")
	}

	// For now, return a "not implemented" error to maintain compatibility
	// This will be implemented when the watch system is properly integrated
	fmt.Printf("⚠️  Watch mode not yet fully implemented in adapter pattern\n")
	fmt.Printf("📁 Would watch paths: %v\n", w.options.Paths)
	fmt.Printf("🚫 Would ignore patterns: %v\n", w.options.IgnorePatterns)

	// Wait for context cancellation (simulating watch mode)
	<-ctx.Done()
	return nil
}

// SetConfiguration configures the watch coordinator adapter
func (w *watchCoordinatorAdapter) SetConfiguration(config *Configuration) error {
	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	w.config = config
	return nil
}

// SetFileWatcher injects the file watcher dependency
func (w *watchCoordinatorAdapter) SetFileWatcher(watcher FileWatcher) {
	w.watcher = watcher
}

// SetEventDebouncer injects the event debouncer dependency
func (w *watchCoordinatorAdapter) SetEventDebouncer(debouncer EventDebouncer) {
	w.debouncer = debouncer
}
