// Package coordinator provides watch system orchestration capabilities
package coordinator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/newbpydev/go-sentinel/internal/watch/core"
	"github.com/newbpydev/go-sentinel/pkg/models"
)

// Coordinator implements the WatchCoordinator interface
type Coordinator struct {
	mu           sync.RWMutex
	fsWatcher    core.FileSystemWatcher
	debouncer    core.EventDebouncer
	testTrigger  core.TestTrigger
	options      core.WatchOptions
	status       core.WatchStatus
	eventChannel chan core.FileEvent
	ctx          context.Context
	cancel       context.CancelFunc
	stopCh       chan struct{}
	stopped      bool
}

// NewCoordinator creates a new watch coordinator
func NewCoordinator(
	fsWatcher core.FileSystemWatcher,
	debouncer core.EventDebouncer,
	testTrigger core.TestTrigger,
) core.WatchCoordinator {
	return &Coordinator{
		fsWatcher:    fsWatcher,
		debouncer:    debouncer,
		testTrigger:  testTrigger,
		eventChannel: make(chan core.FileEvent, 100),
		stopCh:       make(chan struct{}),
		stopped:      false,
		status: core.WatchStatus{
			IsRunning:    false,
			WatchedPaths: []string{},
			Mode:         core.WatchAll,
			EventCount:   0,
			ErrorCount:   0,
		},
	}
}

// Start implements the WatchCoordinator interface
func (c *Coordinator) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.status.IsRunning {
		return models.NewWatchError("start", "", nil).
			WithContext("reason", "already_running").
			WithContext("component", "coordinator")
	}

	// Create context for this watch session
	c.ctx, c.cancel = context.WithCancel(ctx)

	// Update status
	c.status.IsRunning = true
	c.status.StartTime = time.Now()
	c.status.WatchedPaths = c.options.Paths

	// Start the file system watcher
	if err := c.fsWatcher.Watch(c.ctx, c.eventChannel); err != nil && err != context.Canceled {
		c.incrementErrorCount()
		return models.NewWatchError("start_file_watcher", "", err).
			WithContext("component", "file_watcher")
	}

	// Start the event processing loop
	go c.processEvents()

	return nil
}

// Stop implements the WatchCoordinator interface
func (c *Coordinator) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.stopped {
		return nil
	}

	c.stopped = true

	// Cancel the context to stop the file system watcher
	if c.cancel != nil {
		c.cancel()
	}

	// Stop the debouncer
	if err := c.debouncer.Stop(); err != nil {
		c.status.ErrorCount++
		return models.NewWatchError("stop_debouncer", "", err).
			WithContext("component", "debouncer")
	}

	// Close the file system watcher
	if err := c.fsWatcher.Close(); err != nil {
		c.status.ErrorCount++
		return models.NewWatchError("stop_file_watcher", "", err).
			WithContext("component", "file_watcher")
	}

	// Close channels
	close(c.stopCh)
	close(c.eventChannel)

	// Update status
	c.status.IsRunning = false

	return nil
}

// HandleFileChanges implements the WatchCoordinator interface
func (c *Coordinator) HandleFileChanges(changes []core.FileEvent) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.status.IsRunning {
		return models.NewWatchError("handle_changes", "", nil).
			WithContext("reason", "not_running").
			WithContext("component", "coordinator")
	}

	// Update last event time
	c.status.LastEventTime = time.Now()

	// Trigger tests based on watch mode
	switch c.options.Mode {
	case core.WatchAll:
		for _, change := range changes {
			if err := c.testTrigger.TriggerTestsForFile(c.ctx, change.Path); err != nil {
				c.incrementErrorCount()
				return models.NewWatchError("trigger_tests", change.Path, err).
					WithContext("mode", "watch_all").
					WithContext("change_type", change.Type)
			}
		}

	case core.WatchChanged:
		for _, change := range changes {
			if err := c.testTrigger.TriggerTestsForFile(c.ctx, change.Path); err != nil {
				c.incrementErrorCount()
				return models.NewWatchError("trigger_tests", change.Path, err).
					WithContext("mode", "watch_changed").
					WithContext("change_type", change.Type)
			}
		}

	case core.WatchRelated:
		for _, change := range changes {
			if err := c.testTrigger.TriggerRelatedTests(c.ctx, change.Path); err != nil {
				c.incrementErrorCount()
				return models.NewWatchError("trigger_related_tests", change.Path, err).
					WithContext("mode", "watch_related").
					WithContext("change_type", change.Type)
			}
		}

	default:
		return models.NewWatchError("handle_changes", "", nil).
			WithContext("reason", "unknown_mode").
			WithContext("mode", string(c.options.Mode))
	}

	return nil
}

// Configure implements the WatchCoordinator interface
func (c *Coordinator) Configure(options core.WatchOptions) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Store the options
	c.options = options

	// Configure the debouncer interval
	c.debouncer.SetInterval(options.DebounceInterval)

	// Update watch mode in status
	c.status.Mode = options.Mode

	return nil
}

// GetStatus implements the WatchCoordinator interface
func (c *Coordinator) GetStatus() core.WatchStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy of the status
	return c.status
}

// processEvents handles the event processing loop
func (c *Coordinator) processEvents() {
	for {
		select {
		case event := <-c.eventChannel:
			c.incrementEventCount()
			c.debouncer.AddEvent(event)

		case debouncedEvents := <-c.debouncer.Events():
			if len(debouncedEvents) > 0 {
				if err := c.HandleFileChanges(debouncedEvents); err != nil {
					c.incrementErrorCount()
					fmt.Printf("Error handling file changes: %v\n", err)
				}
			}

		case <-c.stopCh:
			return

		case <-c.ctx.Done():
			return
		}
	}
}

// incrementEventCount safely increments the event count
func (c *Coordinator) incrementEventCount() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.status.EventCount++
}

// incrementErrorCount safely increments the error count
func (c *Coordinator) incrementErrorCount() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.status.ErrorCount++
}
