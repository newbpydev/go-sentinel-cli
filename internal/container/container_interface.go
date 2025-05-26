// Package container provides dependency injection container implementation.
// This package follows the Single Responsibility Principle by focusing only on dependency management.
package container

// AppDependencyContainer interface for dependency injection in the container package.
// This interface is defined in the container package and is implemented by dependency containers.
type AppDependencyContainer interface {
	// Core dependency management methods
	Register(name string, component interface{}) error
	Resolve(name string) (interface{}, error)
	ResolveAs(name string, target interface{}) error

	// Lifecycle management
	Initialize() error
	Cleanup() error

	// Advanced registration methods
	RegisterSingleton(name string, factory AppComponentFactory) error

	// Inspection methods
	ListComponents() []string
	HasComponent(name string) bool
}

// AppComponentFactory is a function that creates a component instance.
type AppComponentFactory func() (interface{}, error)

// AppDependencyContainerFactory interface for creating dependency containers.
type AppDependencyContainerFactory interface {
	CreateContainer() AppDependencyContainer
	CreateContainerWithDefaults() AppDependencyContainer
}

// AppDependencyContainerDependencies represents dependencies for container creation.
type AppDependencyContainerDependencies struct {
	// Future extensibility for dependencies
	InitialCapacity int
}

// AppInitializer interface for components that need initialization.
type AppInitializer interface {
	Initialize() error
}

// AppCleaner interface for components that need cleanup.
type AppCleaner interface {
	Cleanup() error
}
