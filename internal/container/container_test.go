// Package container provides dependency injection container implementation tests
package container

import (
	"errors"
	"fmt"
	"sync"
	"testing"
)

// Mock components for testing
type MockComponent struct {
	Name  string
	Value int
}

type MockInitializer struct {
	InitializeCalled bool
	InitializeError  error
}

func (m *MockInitializer) Initialize() error {
	m.InitializeCalled = true
	return m.InitializeError
}

type MockCleaner struct {
	CleanupCalled bool
	CleanupError  error
}

func (m *MockCleaner) Cleanup() error {
	m.CleanupCalled = true
	return m.CleanupError
}

type MockInitializerCleaner struct {
	*MockInitializer
	*MockCleaner
}

// TestNewAppDependencyContainer_FactoryFunction tests the factory function
func TestNewAppDependencyContainer_FactoryFunction(t *testing.T) {
	t.Parallel()

	container := NewAppDependencyContainer()
	if container == nil {
		t.Fatal("NewAppDependencyContainer should not return nil")
	}

	// Verify interface compliance
	_, ok := container.(AppDependencyContainer)
	if !ok {
		t.Fatal("NewAppDependencyContainer should return AppDependencyContainer interface")
	}
}

// TestNewAppDependencyContainerWithCapacity_FactoryFunction tests the capacity factory function
func TestNewAppDependencyContainerWithCapacity_FactoryFunction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		capacity int
	}{
		{"zero_capacity", 0},
		{"small_capacity", 5},
		{"large_capacity", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			container := NewAppDependencyContainerWithCapacity(tt.capacity)
			if container == nil {
				t.Fatal("NewAppDependencyContainerWithCapacity should not return nil")
			}

			// Verify interface compliance
			_, ok := container.(AppDependencyContainer)
			if !ok {
				t.Fatal("NewAppDependencyContainerWithCapacity should return AppDependencyContainer interface")
			}
		})
	}
}

// TestDefaultAppDependencyContainer_Register tests component registration
func TestDefaultAppDependencyContainer_Register(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		componentName string
		component     interface{}
		expectError   bool
		errorContains string
	}{
		{
			name:          "valid_component",
			componentName: "test-component",
			component:     &MockComponent{Name: "test", Value: 42},
			expectError:   false,
		},
		{
			name:          "empty_name",
			componentName: "",
			component:     &MockComponent{},
			expectError:   true,
			errorContains: "component name cannot be empty",
		},
		{
			name:          "nil_component",
			componentName: "test-component",
			component:     nil,
			expectError:   true,
			errorContains: "component cannot be nil",
		},
		{
			name:          "factory_function",
			componentName: "factory-component",
			component: func() (interface{}, error) {
				return &MockComponent{Name: "factory", Value: 100}, nil
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			container := NewAppDependencyContainer()
			err := container.Register(tt.componentName, tt.component)

			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error containing '%s', got nil", tt.errorContains)
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Fatalf("Expected error containing '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got: %v", err)
				}

				// Verify component was registered
				if !container.HasComponent(tt.componentName) {
					t.Fatalf("Component '%s' should be registered", tt.componentName)
				}
			}
		})
	}
}

// TestDefaultAppDependencyContainer_RegisterSingleton tests singleton registration
func TestDefaultAppDependencyContainer_RegisterSingleton(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		componentName string
		factory       AppComponentFactory
		expectError   bool
		errorContains string
	}{
		{
			name:          "valid_singleton",
			componentName: "singleton-component",
			factory: func() (interface{}, error) {
				return &MockComponent{Name: "singleton", Value: 42}, nil
			},
			expectError: false,
		},
		{
			name:          "empty_name",
			componentName: "",
			factory: func() (interface{}, error) {
				return &MockComponent{}, nil
			},
			expectError:   true,
			errorContains: "component name cannot be empty",
		},
		{
			name:          "nil_factory",
			componentName: "test-component",
			factory:       nil,
			expectError:   true,
			errorContains: "factory cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			container := NewAppDependencyContainer()
			err := container.RegisterSingleton(tt.componentName, tt.factory)

			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error containing '%s', got nil", tt.errorContains)
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Fatalf("Expected error containing '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got: %v", err)
				}

				// Verify singleton was registered
				if !container.HasComponent(tt.componentName) {
					t.Fatalf("Singleton '%s' should be registered", tt.componentName)
				}
			}
		})
	}
}

// TestDefaultAppDependencyContainer_Resolve tests component resolution
func TestDefaultAppDependencyContainer_Resolve(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupFunc     func(container AppDependencyContainer)
		resolveName   string
		expectError   bool
		errorContains string
		validateFunc  func(t *testing.T, component interface{})
	}{
		{
			name: "resolve_direct_component",
			setupFunc: func(container AppDependencyContainer) {
				component := &MockComponent{Name: "direct", Value: 42}
				container.Register("direct-component", component)
			},
			resolveName: "direct-component",
			expectError: false,
			validateFunc: func(t *testing.T, component interface{}) {
				mock, ok := component.(*MockComponent)
				if !ok {
					t.Fatalf("Expected *MockComponent, got %T", component)
				}
				if mock.Name != "direct" || mock.Value != 42 {
					t.Fatalf("Expected Name='direct', Value=42, got Name='%s', Value=%d", mock.Name, mock.Value)
				}
			},
		},
		{
			name: "resolve_factory_component",
			setupFunc: func(container AppDependencyContainer) {
				factory := func() (interface{}, error) {
					return &MockComponent{Name: "factory", Value: 100}, nil
				}
				container.RegisterSingleton("factory-component", factory)
			},
			resolveName: "factory-component",
			expectError: false,
			validateFunc: func(t *testing.T, component interface{}) {
				mock, ok := component.(*MockComponent)
				if !ok {
					t.Fatalf("Expected *MockComponent, got %T", component)
				}
				if mock.Name != "factory" || mock.Value != 100 {
					t.Fatalf("Expected Name='factory', Value=100, got Name='%s', Value=%d", mock.Name, mock.Value)
				}
			},
		},
		{
			name:          "resolve_nonexistent_component",
			setupFunc:     func(container AppDependencyContainer) {},
			resolveName:   "nonexistent",
			expectError:   true,
			errorContains: "component 'nonexistent' not found",
		},
		{
			name: "resolve_factory_error",
			setupFunc: func(container AppDependencyContainer) {
				factory := func() (interface{}, error) {
					return nil, errors.New("factory error")
				}
				container.RegisterSingleton("error-component", factory)
			},
			resolveName:   "error-component",
			expectError:   true,
			errorContains: "failed to create component 'error-component'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			container := NewAppDependencyContainer()
			tt.setupFunc(container)

			component, err := container.Resolve(tt.resolveName)

			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error containing '%s', got nil", tt.errorContains)
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Fatalf("Expected error containing '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got: %v", err)
				}
				if component == nil {
					t.Fatal("Expected component, got nil")
				}
				if tt.validateFunc != nil {
					tt.validateFunc(t, component)
				}
			}
		})
	}
}

// TestDefaultAppDependencyContainer_Resolve_Singleton tests singleton behavior
func TestDefaultAppDependencyContainer_Resolve_Singleton(t *testing.T) {
	t.Parallel()

	container := NewAppDependencyContainer()

	// Register a singleton factory
	factory := func() (interface{}, error) {
		return &MockComponent{Name: "singleton", Value: 42}, nil
	}
	err := container.RegisterSingleton("singleton-component", factory)
	if err != nil {
		t.Fatalf("Failed to register singleton: %v", err)
	}

	// Resolve twice and verify it's the same instance
	component1, err := container.Resolve("singleton-component")
	if err != nil {
		t.Fatalf("Failed to resolve singleton first time: %v", err)
	}

	component2, err := container.Resolve("singleton-component")
	if err != nil {
		t.Fatalf("Failed to resolve singleton second time: %v", err)
	}

	// Verify it's the same instance (pointer comparison)
	if component1 != component2 {
		t.Fatal("Singleton should return the same instance on multiple resolves")
	}
}

// TestDefaultAppDependencyContainer_ResolveAs tests type-safe resolution
func TestDefaultAppDependencyContainer_ResolveAs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupFunc     func(container AppDependencyContainer)
		resolveName   string
		target        interface{}
		expectError   bool
		errorContains string
		validateFunc  func(t *testing.T, target interface{})
	}{
		{
			name: "resolve_as_valid_type",
			setupFunc: func(container AppDependencyContainer) {
				component := &MockComponent{Name: "test", Value: 42}
				container.Register("test-component", component)
			},
			resolveName: "test-component",
			target:      new(*MockComponent),
			expectError: false,
			validateFunc: func(t *testing.T, target interface{}) {
				ptr := target.(**MockComponent)
				if *ptr == nil {
					t.Fatal("Target should not be nil")
				}
				if (*ptr).Name != "test" || (*ptr).Value != 42 {
					t.Fatalf("Expected Name='test', Value=42, got Name='%s', Value=%d", (*ptr).Name, (*ptr).Value)
				}
			},
		},
		{
			name:          "resolve_as_nil_target",
			setupFunc:     func(container AppDependencyContainer) {},
			resolveName:   "test-component",
			target:        nil,
			expectError:   true,
			errorContains: "target cannot be nil",
		},
		{
			name:          "resolve_as_non_pointer",
			setupFunc:     func(container AppDependencyContainer) {},
			resolveName:   "test-component",
			target:        MockComponent{},
			expectError:   true,
			errorContains: "target must be a pointer",
		},
		{
			name: "resolve_as_incompatible_type",
			setupFunc: func(container AppDependencyContainer) {
				component := &MockComponent{Name: "test", Value: 42}
				container.Register("test-component", component)
			},
			resolveName:   "test-component",
			target:        new(string),
			expectError:   true,
			errorContains: "is not assignable to target type",
		},
		{
			name:          "resolve_as_nonexistent_component",
			setupFunc:     func(container AppDependencyContainer) {},
			resolveName:   "nonexistent",
			target:        new(*MockComponent),
			expectError:   true,
			errorContains: "component 'nonexistent' not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			container := NewAppDependencyContainer()
			tt.setupFunc(container)

			err := container.ResolveAs(tt.resolveName, tt.target)

			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error containing '%s', got nil", tt.errorContains)
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Fatalf("Expected error containing '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got: %v", err)
				}
				if tt.validateFunc != nil {
					tt.validateFunc(t, tt.target)
				}
			}
		})
	}
}

// TestDefaultAppDependencyContainer_Initialize tests component initialization
func TestDefaultAppDependencyContainer_Initialize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupFunc     func(container AppDependencyContainer) []*MockInitializer
		expectError   bool
		errorContains string
		validateFunc  func(t *testing.T, initializers []*MockInitializer)
	}{
		{
			name: "initialize_success",
			setupFunc: func(container AppDependencyContainer) []*MockInitializer {
				init1 := &MockInitializer{}
				init2 := &MockInitializer{}
				container.Register("init1", init1)
				container.Register("init2", init2)
				return []*MockInitializer{init1, init2}
			},
			expectError: false,
			validateFunc: func(t *testing.T, initializers []*MockInitializer) {
				for i, init := range initializers {
					if !init.InitializeCalled {
						t.Fatalf("Initializer %d should have been called", i)
					}
				}
			},
		},
		{
			name: "initialize_error",
			setupFunc: func(container AppDependencyContainer) []*MockInitializer {
				init1 := &MockInitializer{InitializeError: errors.New("init error")}
				container.Register("init1", init1)
				return []*MockInitializer{init1}
			},
			expectError:   true,
			errorContains: "failed to initialize component 'init1'",
		},
		{
			name: "initialize_mixed_components",
			setupFunc: func(container AppDependencyContainer) []*MockInitializer {
				init := &MockInitializer{}
				regular := &MockComponent{Name: "regular", Value: 42}
				container.Register("init", init)
				container.Register("regular", regular)
				return []*MockInitializer{init}
			},
			expectError: false,
			validateFunc: func(t *testing.T, initializers []*MockInitializer) {
				if !initializers[0].InitializeCalled {
					t.Fatal("Initializer should have been called")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			container := NewAppDependencyContainer()
			initializers := tt.setupFunc(container)

			err := container.Initialize()

			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error containing '%s', got nil", tt.errorContains)
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Fatalf("Expected error containing '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got: %v", err)
				}
				if tt.validateFunc != nil {
					tt.validateFunc(t, initializers)
				}
			}
		})
	}
}

// TestDefaultAppDependencyContainer_Cleanup tests component cleanup
func TestDefaultAppDependencyContainer_Cleanup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupFunc     func(container AppDependencyContainer) ([]*MockCleaner, []*MockCleaner)
		expectError   bool
		errorContains string
		validateFunc  func(t *testing.T, regularCleaners, singletonCleaners []*MockCleaner)
	}{
		{
			name: "cleanup_success",
			setupFunc: func(container AppDependencyContainer) ([]*MockCleaner, []*MockCleaner) {
				cleaner1 := &MockCleaner{}
				cleaner2 := &MockCleaner{}
				container.Register("cleaner1", cleaner1)

				// Create a singleton cleaner
				factory := func() (interface{}, error) {
					return cleaner2, nil
				}
				container.RegisterSingleton("singleton-cleaner", factory)
				// Resolve to create the singleton
				container.Resolve("singleton-cleaner")

				return []*MockCleaner{cleaner1}, []*MockCleaner{cleaner2}
			},
			expectError: false,
			validateFunc: func(t *testing.T, regularCleaners, singletonCleaners []*MockCleaner) {
				for i, cleaner := range regularCleaners {
					if !cleaner.CleanupCalled {
						t.Fatalf("Regular cleaner %d should have been called", i)
					}
				}
				for i, cleaner := range singletonCleaners {
					if !cleaner.CleanupCalled {
						t.Fatalf("Singleton cleaner %d should have been called", i)
					}
				}
			},
		},
		{
			name: "cleanup_error",
			setupFunc: func(container AppDependencyContainer) ([]*MockCleaner, []*MockCleaner) {
				cleaner := &MockCleaner{CleanupError: errors.New("cleanup error")}
				container.Register("cleaner", cleaner)
				return []*MockCleaner{cleaner}, []*MockCleaner{}
			},
			expectError:   true,
			errorContains: "cleanup errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			container := NewAppDependencyContainer()
			regularCleaners, singletonCleaners := tt.setupFunc(container)

			err := container.Cleanup()

			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error containing '%s', got nil", tt.errorContains)
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Fatalf("Expected error containing '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got: %v", err)
				}
				if tt.validateFunc != nil {
					tt.validateFunc(t, regularCleaners, singletonCleaners)
				}
			}

			// Verify container is empty after cleanup
			components := container.ListComponents()
			if len(components) != 0 {
				t.Fatalf("Container should be empty after cleanup, got %d components", len(components))
			}
		})
	}
}

// TestDefaultAppDependencyContainer_ListComponents tests component listing
func TestDefaultAppDependencyContainer_ListComponents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		setupFunc    func(container AppDependencyContainer)
		expectedList []string
	}{
		{
			name:         "empty_container",
			setupFunc:    func(container AppDependencyContainer) {},
			expectedList: []string{},
		},
		{
			name: "single_component",
			setupFunc: func(container AppDependencyContainer) {
				container.Register("component1", &MockComponent{})
			},
			expectedList: []string{"component1"},
		},
		{
			name: "multiple_components",
			setupFunc: func(container AppDependencyContainer) {
				container.Register("component1", &MockComponent{})
				container.Register("component2", &MockComponent{})
				factory := func() (interface{}, error) {
					return &MockComponent{}, nil
				}
				container.RegisterSingleton("singleton1", factory)
			},
			expectedList: []string{"component1", "component2", "singleton1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			container := NewAppDependencyContainer()
			tt.setupFunc(container)

			components := container.ListComponents()

			if len(components) != len(tt.expectedList) {
				t.Fatalf("Expected %d components, got %d", len(tt.expectedList), len(components))
			}

			// Convert to map for easier comparison (order doesn't matter)
			componentMap := make(map[string]bool)
			for _, comp := range components {
				componentMap[comp] = true
			}

			for _, expected := range tt.expectedList {
				if !componentMap[expected] {
					t.Fatalf("Expected component '%s' not found in list", expected)
				}
			}
		})
	}
}

// TestDefaultAppDependencyContainer_HasComponent tests component existence check
func TestDefaultAppDependencyContainer_HasComponent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupFunc     func(container AppDependencyContainer)
		componentName string
		expected      bool
	}{
		{
			name:          "empty_container",
			setupFunc:     func(container AppDependencyContainer) {},
			componentName: "nonexistent",
			expected:      false,
		},
		{
			name: "existing_component",
			setupFunc: func(container AppDependencyContainer) {
				container.Register("test-component", &MockComponent{})
			},
			componentName: "test-component",
			expected:      true,
		},
		{
			name: "existing_singleton",
			setupFunc: func(container AppDependencyContainer) {
				factory := func() (interface{}, error) {
					return &MockComponent{}, nil
				}
				container.RegisterSingleton("singleton-component", factory)
			},
			componentName: "singleton-component",
			expected:      true,
		},
		{
			name: "nonexistent_component",
			setupFunc: func(container AppDependencyContainer) {
				container.Register("test-component", &MockComponent{})
			},
			componentName: "nonexistent",
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			container := NewAppDependencyContainer()
			tt.setupFunc(container)

			result := container.HasComponent(tt.componentName)

			if result != tt.expected {
				t.Fatalf("Expected HasComponent('%s') to return %v, got %v", tt.componentName, tt.expected, result)
			}
		})
	}
}

// TestDefaultAppDependencyContainer_Concurrency tests thread safety
func TestDefaultAppDependencyContainer_Concurrency(t *testing.T) {
	t.Parallel()

	container := NewAppDependencyContainer()

	// Number of goroutines to run concurrently
	numGoroutines := 10
	numOperations := 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Run concurrent operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				componentName := fmt.Sprintf("component-%d-%d", id, j)
				component := &MockComponent{Name: componentName, Value: j}

				// Register component
				err := container.Register(componentName, component)
				if err != nil {
					t.Errorf("Failed to register component %s: %v", componentName, err)
					return
				}

				// Check if component exists
				if !container.HasComponent(componentName) {
					t.Errorf("Component %s should exist", componentName)
					return
				}

				// Resolve component
				resolved, err := container.Resolve(componentName)
				if err != nil {
					t.Errorf("Failed to resolve component %s: %v", componentName, err)
					return
				}

				// Verify resolved component
				if resolved != component {
					t.Errorf("Resolved component should be the same instance")
					return
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify final state
	components := container.ListComponents()
	expectedCount := numGoroutines * numOperations
	if len(components) != expectedCount {
		t.Fatalf("Expected %d components, got %d", expectedCount, len(components))
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 0; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}
