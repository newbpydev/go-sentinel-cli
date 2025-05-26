// Package container provides dependency injection container implementation
package container

import (
	"fmt"
	"reflect"
	"sync"
)

// DefaultAppDependencyContainer implements the AppDependencyContainer interface.
// This implementation follows the Single Responsibility Principle by focusing only on dependency management.
type DefaultAppDependencyContainer struct {
	mu         sync.RWMutex
	components map[string]interface{}
	factories  map[string]func() (interface{}, error)
	singletons map[string]interface{}
}

// NewAppDependencyContainer creates a new dependency injection container with default configuration.
func NewAppDependencyContainer() AppDependencyContainer {
	return &DefaultAppDependencyContainer{
		components: make(map[string]interface{}),
		factories:  make(map[string]func() (interface{}, error)),
		singletons: make(map[string]interface{}),
	}
}

// NewAppDependencyContainerWithCapacity creates a new container with specified initial capacity.
func NewAppDependencyContainerWithCapacity(capacity int) AppDependencyContainer {
	return &DefaultAppDependencyContainer{
		components: make(map[string]interface{}, capacity),
		factories:  make(map[string]func() (interface{}, error), capacity),
		singletons: make(map[string]interface{}, capacity),
	}
}

// Register registers a component with the container
func (c *DefaultAppDependencyContainer) Register(name string, component interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if name == "" {
		return fmt.Errorf("component name cannot be empty")
	}

	if component == nil {
		return fmt.Errorf("component cannot be nil")
	}

	// Check if it's a factory function
	if factory, ok := component.(func() (interface{}, error)); ok {
		c.factories[name] = factory
	} else {
		c.components[name] = component
	}

	return nil
}

// RegisterSingleton registers a component as a singleton
func (c *DefaultAppDependencyContainer) RegisterSingleton(name string, factory AppComponentFactory) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if name == "" {
		return fmt.Errorf("component name cannot be empty")
	}

	if factory == nil {
		return fmt.Errorf("factory cannot be nil")
	}

	c.factories[name] = factory
	return nil
}

// Resolve retrieves a component from the container
func (c *DefaultAppDependencyContainer) Resolve(name string) (interface{}, error) {
	c.mu.RLock()

	// Check if it's already a created singleton
	if singleton, exists := c.singletons[name]; exists {
		c.mu.RUnlock()
		return singleton, nil
	}

	// Check if it's a direct component
	if component, exists := c.components[name]; exists {
		c.mu.RUnlock()
		return component, nil
	}

	// Check if it's a factory
	factory, exists := c.factories[name]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("component '%s' not found", name)
	}

	// Create component using factory
	component, err := factory()
	if err != nil {
		return nil, fmt.Errorf("failed to create component '%s': %w", name, err)
	}

	// Store as singleton if it was registered as a factory
	c.mu.Lock()
	c.singletons[name] = component
	c.mu.Unlock()

	return component, nil
}

// ResolveAs retrieves a component and casts it to the specified type
func (c *DefaultAppDependencyContainer) ResolveAs(name string, target interface{}) error {
	if target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer")
	}

	targetElement := targetValue.Elem()
	if !targetElement.CanSet() {
		return fmt.Errorf("target cannot be set")
	}

	// Resolve the component
	component, err := c.Resolve(name)
	if err != nil {
		return err
	}

	componentValue := reflect.ValueOf(component)

	// Check if types are compatible
	if !componentValue.Type().AssignableTo(targetElement.Type()) {
		return fmt.Errorf("component '%s' of type %T is not assignable to target type %s",
			name, component, targetElement.Type())
	}

	// Assign the component to the target
	targetElement.Set(componentValue)
	return nil
}

// Initialize initializes all registered components
func (c *DefaultAppDependencyContainer) Initialize() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Initialize any components that implement AppInitializer interface
	for name, component := range c.components {
		if initializer, ok := component.(AppInitializer); ok {
			if err := initializer.Initialize(); err != nil {
				return fmt.Errorf("failed to initialize component '%s': %w", name, err)
			}
		}
	}

	return nil
}

// Cleanup cleans up all registered components
func (c *DefaultAppDependencyContainer) Cleanup() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var errors []error

	// Cleanup singletons first
	for name, component := range c.singletons {
		if cleaner, ok := component.(AppCleaner); ok {
			if err := cleaner.Cleanup(); err != nil {
				errors = append(errors, fmt.Errorf("failed to cleanup singleton '%s': %w", name, err))
			}
		}
	}

	// Cleanup regular components
	for name, component := range c.components {
		if cleaner, ok := component.(AppCleaner); ok {
			if err := cleaner.Cleanup(); err != nil {
				errors = append(errors, fmt.Errorf("failed to cleanup component '%s': %w", name, err))
			}
		}
	}

	// Clear all maps
	c.components = make(map[string]interface{})
	c.factories = make(map[string]func() (interface{}, error))
	c.singletons = make(map[string]interface{})

	// Return combined errors if any
	if len(errors) > 0 {
		return fmt.Errorf("cleanup errors: %v", errors)
	}

	return nil
}

// ListComponents returns a list of all registered component names
func (c *DefaultAppDependencyContainer) ListComponents() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var names []string
	for name := range c.components {
		names = append(names, name)
	}
	for name := range c.factories {
		// Avoid duplicates
		found := false
		for _, existing := range names {
			if existing == name {
				found = true
				break
			}
		}
		if !found {
			names = append(names, name)
		}
	}
	return names
}

// HasComponent returns true if a component with the given name is registered
func (c *DefaultAppDependencyContainer) HasComponent(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, hasComponent := c.components[name]
	_, hasFactory := c.factories[name]
	return hasComponent || hasFactory
}

// Ensure DefaultAppDependencyContainer implements AppDependencyContainer interface
var _ AppDependencyContainer = (*DefaultAppDependencyContainer)(nil)
