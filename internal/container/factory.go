// Package container provides factory for creating dependency containers
package container

// DefaultAppDependencyContainerFactory implements the AppDependencyContainerFactory interface.
// This factory follows the Factory pattern and dependency injection principles.
type DefaultAppDependencyContainerFactory struct {
	// Dependencies for creating containers
	defaultCapacity int
}

// NewAppDependencyContainerFactory creates a new dependency container factory.
func NewAppDependencyContainerFactory() AppDependencyContainerFactory {
	return &DefaultAppDependencyContainerFactory{
		defaultCapacity: 10, // Default initial capacity
	}
}

// NewAppDependencyContainerFactoryWithDependencies creates a factory with injected dependencies.
func NewAppDependencyContainerFactoryWithDependencies(deps AppDependencyContainerDependencies) AppDependencyContainerFactory {
	capacity := 10 // Default
	if deps.InitialCapacity > 0 {
		capacity = deps.InitialCapacity
	}

	return &DefaultAppDependencyContainerFactory{
		defaultCapacity: capacity,
	}
}

// CreateContainer creates a new dependency container with default configuration.
func (f *DefaultAppDependencyContainerFactory) CreateContainer() AppDependencyContainer {
	return NewAppDependencyContainerWithCapacity(f.defaultCapacity)
}

// CreateContainerWithDefaults creates a dependency container using factory defaults.
// This method demonstrates the Factory pattern providing sensible defaults.
func (f *DefaultAppDependencyContainerFactory) CreateContainerWithDefaults() AppDependencyContainer {
	return NewAppDependencyContainerWithCapacity(f.defaultCapacity)
}

// Ensure DefaultAppDependencyContainerFactory implements AppDependencyContainerFactory interface
var _ AppDependencyContainerFactory = (*DefaultAppDependencyContainerFactory)(nil)
