// Package app provides adapter for dependency container to maintain clean package boundaries
package app

import (
	"github.com/newbpydev/go-sentinel/internal/container"
)

// dependencyContainerAdapter adapts the container package to the app package interface.
// This adapter pattern allows us to maintain compatibility while moving to proper architecture.
type dependencyContainerAdapter struct {
	factory   *DependencyContainerFactory
	container container.AppDependencyContainer
}

// Register registers a component with the container
func (a *dependencyContainerAdapter) Register(name string, component interface{}) error {
	return a.container.Register(name, component)
}

// Resolve retrieves a component from the container
func (a *dependencyContainerAdapter) Resolve(name string) (interface{}, error) {
	return a.container.Resolve(name)
}

// ResolveAs retrieves a component and casts it to the specified type
func (a *dependencyContainerAdapter) ResolveAs(name string, target interface{}) error {
	return a.container.ResolveAs(name, target)
}

// Initialize initializes all registered components
func (a *dependencyContainerAdapter) Initialize() error {
	return a.container.Initialize()
}

// Cleanup cleans up all registered components
func (a *dependencyContainerAdapter) Cleanup() error {
	return a.container.Cleanup()
}

// DependencyContainerFactory creates and manages dependency container adapters.
// This factory follows dependency injection principles and maintains package boundaries.
type DependencyContainerFactory struct {
	containerFactory container.AppDependencyContainerFactory
}

// NewDependencyContainerFactory creates a new dependency container factory.
func NewDependencyContainerFactory() *DependencyContainerFactory {
	return &DependencyContainerFactory{
		containerFactory: container.NewAppDependencyContainerFactory(),
	}
}

// CreateContainerWithDefaults creates a dependency container adapter with default settings.
func (f *DependencyContainerFactory) CreateContainerWithDefaults() DependencyContainer {
	containerImpl := f.containerFactory.CreateContainerWithDefaults()

	return &dependencyContainerAdapter{
		factory:   f,
		container: containerImpl,
	}
}

// NewContainer creates a new dependency container using the adapter pattern.
// This follows dependency injection principles and maintains package boundaries.
func NewContainer() DependencyContainer {
	factory := NewDependencyContainerFactory()
	return factory.CreateContainerWithDefaults()
}

// ComponentFactory is a function that creates a component instance (for compatibility)
type ComponentFactory func() (interface{}, error)

// Initializer interface for components that need initialization (for compatibility)
type Initializer interface {
	Initialize() error
}

// Cleaner interface for components that need cleanup (for compatibility)
type Cleaner interface {
	Cleanup() error
}

// Ensure dependencyContainerAdapter implements DependencyContainer interface
var _ DependencyContainer = (*dependencyContainerAdapter)(nil)
