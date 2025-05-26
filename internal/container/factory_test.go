// Package container provides factory tests for dependency injection container
package container

import (
	"testing"
)

// TestNewAppDependencyContainerFactory_FactoryFunction tests the factory creation
func TestNewAppDependencyContainerFactory_FactoryFunction(t *testing.T) {
	t.Parallel()

	factory := NewAppDependencyContainerFactory()
	if factory == nil {
		t.Fatal("NewAppDependencyContainerFactory should not return nil")
	}

	// Verify interface compliance
	_, ok := factory.(AppDependencyContainerFactory)
	if !ok {
		t.Fatal("NewAppDependencyContainerFactory should return AppDependencyContainerFactory interface")
	}
}

// TestNewAppDependencyContainerFactoryWithDependencies_FactoryFunction tests factory with dependencies
func TestNewAppDependencyContainerFactoryWithDependencies_FactoryFunction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		dependencies AppDependencyContainerDependencies
	}{
		{
			name:         "zero_capacity",
			dependencies: AppDependencyContainerDependencies{InitialCapacity: 0},
		},
		{
			name:         "positive_capacity",
			dependencies: AppDependencyContainerDependencies{InitialCapacity: 50},
		},
		{
			name:         "large_capacity",
			dependencies: AppDependencyContainerDependencies{InitialCapacity: 1000},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			factory := NewAppDependencyContainerFactoryWithDependencies(tt.dependencies)
			if factory == nil {
				t.Fatal("NewAppDependencyContainerFactoryWithDependencies should not return nil")
			}

			// Verify interface compliance
			_, ok := factory.(AppDependencyContainerFactory)
			if !ok {
				t.Fatal("NewAppDependencyContainerFactoryWithDependencies should return AppDependencyContainerFactory interface")
			}
		})
	}
}

// TestDefaultAppDependencyContainerFactory_CreateContainer tests basic container creation
func TestDefaultAppDependencyContainerFactory_CreateContainer(t *testing.T) {
	t.Parallel()

	factory := NewAppDependencyContainerFactory()
	container := factory.CreateContainer()

	if container == nil {
		t.Fatal("CreateContainer should not return nil")
	}

	// Verify interface compliance
	_, ok := container.(AppDependencyContainer)
	if !ok {
		t.Fatal("CreateContainer should return AppDependencyContainer interface")
	}

	// Verify container is functional
	testComponent := &MockComponent{Name: "test", Value: 42}
	err := container.Register("test-component", testComponent)
	if err != nil {
		t.Fatalf("Container should be functional, got error: %v", err)
	}

	if !container.HasComponent("test-component") {
		t.Fatal("Container should have registered component")
	}
}

// TestDefaultAppDependencyContainerFactory_CreateContainerWithDefaults tests container with defaults
func TestDefaultAppDependencyContainerFactory_CreateContainerWithDefaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		dependencies AppDependencyContainerDependencies
	}{
		{
			name:         "default_factory",
			dependencies: AppDependencyContainerDependencies{}, // Use default factory
		},
		{
			name:         "custom_capacity",
			dependencies: AppDependencyContainerDependencies{InitialCapacity: 25},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var factory AppDependencyContainerFactory
			if tt.name == "default_factory" {
				factory = NewAppDependencyContainerFactory()
			} else {
				factory = NewAppDependencyContainerFactoryWithDependencies(tt.dependencies)
			}

			container := factory.CreateContainerWithDefaults()

			if container == nil {
				t.Fatal("CreateContainerWithDefaults should not return nil")
			}

			// Verify interface compliance
			_, ok := container.(AppDependencyContainer)
			if !ok {
				t.Fatal("CreateContainerWithDefaults should return AppDependencyContainer interface")
			}

			// Verify container is functional
			testComponent := &MockComponent{Name: "test", Value: 42}
			err := container.Register("test-component", testComponent)
			if err != nil {
				t.Fatalf("Container should be functional, got error: %v", err)
			}

			if !container.HasComponent("test-component") {
				t.Fatal("Container should have registered component")
			}
		})
	}
}

// TestDefaultAppDependencyContainerFactory_MultipleContainers tests creating multiple containers
func TestDefaultAppDependencyContainerFactory_MultipleContainers(t *testing.T) {
	t.Parallel()

	factory := NewAppDependencyContainerFactory()

	// Create multiple containers
	container1 := factory.CreateContainer()
	container2 := factory.CreateContainer()
	container3 := factory.CreateContainerWithDefaults()

	// Verify they are different instances
	if container1 == container2 {
		t.Fatal("Factory should create different container instances")
	}
	if container1 == container3 {
		t.Fatal("Factory should create different container instances")
	}
	if container2 == container3 {
		t.Fatal("Factory should create different container instances")
	}

	// Verify they are independent
	testComponent1 := &MockComponent{Name: "test1", Value: 1}
	testComponent2 := &MockComponent{Name: "test2", Value: 2}

	err := container1.Register("component", testComponent1)
	if err != nil {
		t.Fatalf("Failed to register in container1: %v", err)
	}

	err = container2.Register("component", testComponent2)
	if err != nil {
		t.Fatalf("Failed to register in container2: %v", err)
	}

	// Verify isolation
	resolved1, err := container1.Resolve("component")
	if err != nil {
		t.Fatalf("Failed to resolve from container1: %v", err)
	}

	resolved2, err := container2.Resolve("component")
	if err != nil {
		t.Fatalf("Failed to resolve from container2: %v", err)
	}

	if resolved1 == resolved2 {
		t.Fatal("Containers should be isolated from each other")
	}

	// Verify container3 doesn't have the component
	if container3.HasComponent("component") {
		t.Fatal("Container3 should not have components from other containers")
	}
}

// TestDefaultAppDependencyContainerFactory_InterfaceCompliance tests interface compliance
func TestDefaultAppDependencyContainerFactory_InterfaceCompliance(t *testing.T) {
	t.Parallel()

	// Test that DefaultAppDependencyContainerFactory implements the interface
	var _ AppDependencyContainerFactory = (*DefaultAppDependencyContainerFactory)(nil)

	// Test factory creation methods
	factory1 := NewAppDependencyContainerFactory()
	factory2 := NewAppDependencyContainerFactoryWithDependencies(AppDependencyContainerDependencies{
		InitialCapacity: 20,
	})

	// Verify both implement the interface
	_, ok1 := factory1.(AppDependencyContainerFactory)
	_, ok2 := factory2.(AppDependencyContainerFactory)

	if !ok1 {
		t.Fatal("NewAppDependencyContainerFactory should return AppDependencyContainerFactory")
	}
	if !ok2 {
		t.Fatal("NewAppDependencyContainerFactoryWithDependencies should return AppDependencyContainerFactory")
	}
}

// TestDefaultAppDependencyContainerFactory_DependencyInjection tests dependency injection pattern
func TestDefaultAppDependencyContainerFactory_DependencyInjection(t *testing.T) {
	t.Parallel()

	// Test with different dependency configurations
	tests := []struct {
		name         string
		dependencies AppDependencyContainerDependencies
		description  string
	}{
		{
			name:         "minimal_dependencies",
			dependencies: AppDependencyContainerDependencies{InitialCapacity: 1},
			description:  "Factory with minimal capacity",
		},
		{
			name:         "standard_dependencies",
			dependencies: AppDependencyContainerDependencies{InitialCapacity: 10},
			description:  "Factory with standard capacity",
		},
		{
			name:         "high_capacity_dependencies",
			dependencies: AppDependencyContainerDependencies{InitialCapacity: 100},
			description:  "Factory with high capacity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			factory := NewAppDependencyContainerFactoryWithDependencies(tt.dependencies)

			// Create containers using both methods
			container1 := factory.CreateContainer()
			container2 := factory.CreateContainerWithDefaults()

			// Verify both containers work
			if container1 == nil {
				t.Fatal("CreateContainer should not return nil")
			}
			if container2 == nil {
				t.Fatal("CreateContainerWithDefaults should not return nil")
			}

			// Test functionality
			testComponent := &MockComponent{Name: "test", Value: 42}

			err := container1.Register("test", testComponent)
			if err != nil {
				t.Fatalf("Container1 should work: %v", err)
			}

			err = container2.Register("test", testComponent)
			if err != nil {
				t.Fatalf("Container2 should work: %v", err)
			}
		})
	}
}

// TestDefaultAppDependencyContainerFactory_FactoryPattern tests factory pattern implementation
func TestDefaultAppDependencyContainerFactory_FactoryPattern(t *testing.T) {
	t.Parallel()

	// Test that factory encapsulates creation logic
	factory := NewAppDependencyContainerFactory()

	// Create multiple containers and verify they follow the same pattern
	containers := make([]AppDependencyContainer, 5)
	for i := 0; i < 5; i++ {
		containers[i] = factory.CreateContainer()
		if containers[i] == nil {
			t.Fatalf("Container %d should not be nil", i)
		}
	}

	// Verify all containers have the same interface but are different instances
	for i := 0; i < len(containers); i++ {
		for j := i + 1; j < len(containers); j++ {
			if containers[i] == containers[j] {
				t.Fatalf("Containers %d and %d should be different instances", i, j)
			}
		}
	}

	// Verify all containers work independently
	for i, container := range containers {
		testComponent := &MockComponent{Name: "test", Value: i}
		err := container.Register("component", testComponent)
		if err != nil {
			t.Fatalf("Container %d should work: %v", i, err)
		}

		resolved, err := container.Resolve("component")
		if err != nil {
			t.Fatalf("Container %d should resolve: %v", i, err)
		}

		mock, ok := resolved.(*MockComponent)
		if !ok {
			t.Fatalf("Container %d should return MockComponent", i)
		}

		if mock.Value != i {
			t.Fatalf("Container %d should have value %d, got %d", i, i, mock.Value)
		}
	}
}
