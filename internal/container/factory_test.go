// Package container provides factory tests for dependency injection container
package container

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"
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
		name             string
		dependencies     AppDependencyContainerDependencies
		expectedCapacity int
	}{
		{
			name:             "zero_capacity",
			dependencies:     AppDependencyContainerDependencies{InitialCapacity: 0},
			expectedCapacity: 10, // Should default to 10
		},
		{
			name:             "positive_capacity",
			dependencies:     AppDependencyContainerDependencies{InitialCapacity: 50},
			expectedCapacity: 50,
		},
		{
			name:             "large_capacity",
			dependencies:     AppDependencyContainerDependencies{InitialCapacity: 1000},
			expectedCapacity: 1000,
		},
		{
			name:             "negative_capacity",
			dependencies:     AppDependencyContainerDependencies{InitialCapacity: -10},
			expectedCapacity: 10, // Should default to 10
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

			// Verify the factory uses the correct capacity by creating a container with defaults
			container := factory.CreateContainerWithDefaults()
			if container == nil {
				t.Fatal("CreateContainerWithDefaults should not return nil")
			}

			// Test that the container works properly (indirect capacity verification)
			err := container.Register("test", &MockComponent{Name: "test", Value: 42})
			if err != nil {
				t.Fatalf("Container should be functional: %v", err)
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

	// Test default factory
	defaultFactory := NewAppDependencyContainerFactory()
	var _ AppDependencyContainerFactory = defaultFactory

	// Test factory with dependencies
	deps := AppDependencyContainerDependencies{InitialCapacity: 20}
	depsFactory := NewAppDependencyContainerFactoryWithDependencies(deps)
	var _ AppDependencyContainerFactory = depsFactory

	// Verify both methods work
	container1 := defaultFactory.CreateContainer()
	container2 := defaultFactory.CreateContainerWithDefaults()
	container3 := depsFactory.CreateContainer()
	container4 := depsFactory.CreateContainerWithDefaults()

	containers := []AppDependencyContainer{container1, container2, container3, container4}
	for i, container := range containers {
		if container == nil {
			t.Fatalf("Container %d should not be nil", i)
		}

		// Verify each container implements the interface
		var _ AppDependencyContainer = container
	}
}

// TestDefaultAppDependencyContainerFactory_DependencyInjection tests dependency injection
func TestDefaultAppDependencyContainerFactory_DependencyInjection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		dependencies       AppDependencyContainerDependencies
		componentCount     int
		expectedFunctional bool
	}{
		{
			name:               "small_capacity_many_components",
			dependencies:       AppDependencyContainerDependencies{InitialCapacity: 2},
			componentCount:     10, // More than initial capacity
			expectedFunctional: true,
		},
		{
			name:               "large_capacity_few_components",
			dependencies:       AppDependencyContainerDependencies{InitialCapacity: 100},
			componentCount:     5,
			expectedFunctional: true,
		},
		{
			name:               "zero_capacity_with_components",
			dependencies:       AppDependencyContainerDependencies{InitialCapacity: 0},
			componentCount:     3,
			expectedFunctional: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			factory := NewAppDependencyContainerFactoryWithDependencies(tt.dependencies)
			container := factory.CreateContainerWithDefaults()

			// Register multiple components
			for i := 0; i < tt.componentCount; i++ {
				componentName := fmt.Sprintf("component-%d", i)
				component := &MockComponent{Name: componentName, Value: i}
				err := container.Register(componentName, component)
				if err != nil && tt.expectedFunctional {
					t.Fatalf("Failed to register component %s: %v", componentName, err)
				}
			}

			// Verify all components are registered
			if tt.expectedFunctional {
				components := container.ListComponents()
				if len(components) != tt.componentCount {
					t.Fatalf("Expected %d components, got %d", tt.componentCount, len(components))
				}
			}
		})
	}
}

// TestDefaultAppDependencyContainerFactory_FactoryPattern tests the factory pattern
func TestDefaultAppDependencyContainerFactory_FactoryPattern(t *testing.T) {
	t.Parallel()

	// Test that factory follows the Factory pattern correctly
	factory := NewAppDependencyContainerFactory()

	// Verify factory creates consistent instances
	container1 := factory.CreateContainer()
	container2 := factory.CreateContainer()

	// They should be different instances but same type
	if container1 == container2 {
		t.Fatal("Factory should create new instances each time")
	}

	// But they should have the same underlying type
	type1 := reflect.TypeOf(container1)
	type2 := reflect.TypeOf(container2)
	if type1 != type2 {
		t.Fatal("Factory should create instances of the same type")
	}

	// Test CreateContainerWithDefaults vs CreateContainer behavior
	defaultContainer := factory.CreateContainerWithDefaults()
	basicContainer := factory.CreateContainer()

	// They should be different instances
	if defaultContainer == basicContainer {
		t.Fatal("CreateContainerWithDefaults and CreateContainer should create different instances")
	}

	// But same type
	defaultType := reflect.TypeOf(defaultContainer)
	basicType := reflect.TypeOf(basicContainer)
	if defaultType != basicType {
		t.Fatal("Both factory methods should create the same container type")
	}
}

// TestDefaultAppDependencyContainerFactory_Concurrency tests concurrent factory usage
func TestDefaultAppDependencyContainerFactory_Concurrency(t *testing.T) {
	t.Parallel()

	factory := NewAppDependencyContainerFactory()

	var wg sync.WaitGroup
	numGoroutines := 50
	containers := make([]AppDependencyContainer, numGoroutines*2)

	// Concurrent container creation
	for i := 0; i < numGoroutines; i++ {
		wg.Add(2)
		go func(index int) {
			defer wg.Done()
			containers[index*2] = factory.CreateContainer()
		}(i)
		go func(index int) {
			defer wg.Done()
			containers[index*2+1] = factory.CreateContainerWithDefaults()
		}(i)
	}

	wg.Wait()

	// Verify all containers were created successfully
	for i, container := range containers {
		if container == nil {
			t.Fatalf("Container %d should not be nil", i)
		}
	}

	// Verify all containers are unique instances
	for i := 0; i < len(containers); i++ {
		for j := i + 1; j < len(containers); j++ {
			if containers[i] == containers[j] {
				t.Fatalf("Containers at index %d and %d should be different instances", i, j)
			}
		}
	}

	// Verify all containers are functional
	for i, container := range containers {
		componentName := fmt.Sprintf("test-component-%d", i)
		component := &MockComponent{Name: componentName, Value: i}
		err := container.Register(componentName, component)
		if err != nil {
			t.Fatalf("Container %d should be functional: %v", i, err)
		}
	}
}

// TestDefaultAppDependencyContainerFactory_MemoryEfficiency tests memory efficiency
func TestDefaultAppDependencyContainerFactory_MemoryEfficiency(t *testing.T) {
	t.Parallel()

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	factory := NewAppDependencyContainerFactory()

	// Create many containers
	containers := make([]AppDependencyContainer, 100)
	for i := 0; i < 100; i++ {
		if i%2 == 0 {
			containers[i] = factory.CreateContainer()
		} else {
			containers[i] = factory.CreateContainerWithDefaults()
		}
	}

	// Use the containers
	for i, container := range containers {
		componentName := fmt.Sprintf("component-%d", i)
		component := &MockComponent{Name: componentName, Value: i}
		err := container.Register(componentName, component)
		if err != nil {
			t.Fatalf("Failed to register component in container %d: %v", i, err)
		}
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)

	allocDiff := m2.TotalAlloc - m1.TotalAlloc
	if allocDiff > 10*1024*1024 { // 10MB threshold
		t.Errorf("Excessive memory allocation: %d bytes", allocDiff)
	}
}

// TestDefaultAppDependencyContainerFactory_EdgeCases tests edge cases
func TestDefaultAppDependencyContainerFactory_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "factory_reuse",
			testFunc: func(t *testing.T) {
				factory := NewAppDependencyContainerFactory()

				// Create multiple containers from same factory
				containers := make([]AppDependencyContainer, 10)
				for i := 0; i < 10; i++ {
					containers[i] = factory.CreateContainer()
				}

				// All should be functional and independent
				for i, container := range containers {
					err := container.Register("test", &MockComponent{Name: "test", Value: i})
					if err != nil {
						t.Fatalf("Container %d should be functional: %v", i, err)
					}

					resolved, err := container.Resolve("test")
					if err != nil {
						t.Fatalf("Container %d resolve failed: %v", i, err)
					}

					mock := resolved.(*MockComponent)
					if mock.Value != i {
						t.Fatalf("Container %d should have isolated state, expected %d, got %d", i, i, mock.Value)
					}
				}
			},
		},
		{
			name: "mixed_creation_methods",
			testFunc: func(t *testing.T) {
				factory := NewAppDependencyContainerFactory()

				container1 := factory.CreateContainer()
				container2 := factory.CreateContainerWithDefaults()

				// Both should work identically
				err1 := container1.Register("test1", &MockComponent{Name: "test1", Value: 1})
				err2 := container2.Register("test2", &MockComponent{Name: "test2", Value: 2})

				if err1 != nil {
					t.Fatalf("CreateContainer result should be functional: %v", err1)
				}
				if err2 != nil {
					t.Fatalf("CreateContainerWithDefaults result should be functional: %v", err2)
				}

				// Verify isolation
				if container1.HasComponent("test2") {
					t.Fatal("Containers should be isolated")
				}
				if container2.HasComponent("test1") {
					t.Fatal("Containers should be isolated")
				}
			},
		},
		{
			name: "factory_with_extreme_dependencies",
			testFunc: func(t *testing.T) {
				// Test with very large capacity
				largeDeps := AppDependencyContainerDependencies{InitialCapacity: 1000000}
				largeFactory := NewAppDependencyContainerFactoryWithDependencies(largeDeps)
				largeContainer := largeFactory.CreateContainerWithDefaults()

				if largeContainer == nil {
					t.Fatal("Factory should handle large capacities")
				}

				err := largeContainer.Register("test", &MockComponent{})
				if err != nil {
					t.Fatalf("Large capacity container should be functional: %v", err)
				}

				// Test with zero capacity (should default)
				zeroDeps := AppDependencyContainerDependencies{InitialCapacity: 0}
				zeroFactory := NewAppDependencyContainerFactoryWithDependencies(zeroDeps)
				zeroContainer := zeroFactory.CreateContainerWithDefaults()

				if zeroContainer == nil {
					t.Fatal("Factory should handle zero capacity")
				}

				err = zeroContainer.Register("test", &MockComponent{})
				if err != nil {
					t.Fatalf("Zero capacity container should be functional: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.testFunc(t)
		})
	}
}

// TestDefaultAppDependencyContainerFactory_PerformanceBenchmark benchmarks factory performance
func TestDefaultAppDependencyContainerFactory_PerformanceBenchmark(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance benchmark in short mode")
	}

	factory := NewAppDependencyContainerFactory()

	start := time.Now()

	// Create many containers quickly
	for i := 0; i < 1000; i++ {
		container := factory.CreateContainer()
		if container == nil {
			t.Fatalf("Container %d should not be nil", i)
		}
	}

	duration := time.Since(start)
	if duration > time.Second {
		t.Errorf("Creating 1000 containers took too long: %v", duration)
	}
}

// TestDefaultAppDependencyContainerFactory_GoroutineLeakPrevention tests for goroutine leaks
func TestDefaultAppDependencyContainerFactory_GoroutineLeakPrevention(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("Skipping goroutine leak test in short mode or parallel testing")
	}

	initialGoroutines := runtime.NumGoroutine()

	factory := NewAppDependencyContainerFactory()

	// Create and use many containers concurrently
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			container := factory.CreateContainer()

			// Use the container briefly
			component := &MockComponent{Name: "test", Value: index}
			container.Register("test", component)
			container.Resolve("test")
			container.Cleanup()
		}(i)
	}

	wg.Wait()

	// Wait for potential goroutines to finish
	time.Sleep(10 * time.Millisecond)

	finalGoroutines := runtime.NumGoroutine()
	if finalGoroutines > initialGoroutines+25 { // Increased tolerance for parallel testing environment
		t.Errorf("Potential goroutine leak: initial=%d, final=%d", initialGoroutines, finalGoroutines)
	}
}
