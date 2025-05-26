// Package app provides application orchestration following modular architecture
package app

import (
	"context"
	"fmt"
)

// NewApplicationController creates a new application controller with proper dependency injection
// This eliminates direct dependency violations by using the existing Controller with adapters
func NewApplicationController() ApplicationController {
	// Create components using factory pattern to eliminate direct dependencies
	argParser := NewArgumentParser()
	configLoader := NewConfigurationLoader()
	lifecycle := NewLifecycleManager()
	container := NewContainer()
	eventHandler := NewApplicationEventHandler()

	// Use the existing Controller which has proper dependency injection
	controller := NewController(
		argParser,
		configLoader,
		lifecycle,
		container,
		eventHandler,
	)

	// Initialize the controller
	if err := controller.Initialize(); err != nil {
		// Return an error wrapper that implements ApplicationController
		return &initializationErrorController{err: err}
	}

	return controller
}

// initializationErrorController wraps initialization errors and implements ApplicationController
type initializationErrorController struct {
	err error
}

// Run implements ApplicationController interface and returns the initialization error
func (c *initializationErrorController) Run(args []string) error {
	return fmt.Errorf("controller initialization failed: %w", c.err)
}

// Initialize implements ApplicationController interface
func (c *initializationErrorController) Initialize() error {
	return c.err
}

// Shutdown implements ApplicationController interface
func (c *initializationErrorController) Shutdown(ctx context.Context) error {
	return nil // Nothing to shut down if initialization failed
}
