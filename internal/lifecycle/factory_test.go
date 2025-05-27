// Package lifecycle provides comprehensive tests for lifecycle manager factory
package lifecycle

import (
	"context"
	"testing"
	"time"
)

// TestNewAppLifecycleManagerFactory_FactoryFunction tests the factory creation
func TestNewAppLifecycleManagerFactory_FactoryFunction(t *testing.T) {
	t.Parallel()

	factory := NewAppLifecycleManagerFactory()
	if factory == nil {
		t.Fatal("NewAppLifecycleManagerFactory should not return nil")
	}

	// Verify interface compliance
	_, ok := factory.(AppLifecycleManagerFactory)
	if !ok {
		t.Fatal("NewAppLifecycleManagerFactory should return AppLifecycleManagerFactory interface")
	}
}

// TestNewAppLifecycleManagerFactoryWithDependencies_FactoryFunction tests factory with dependencies
func TestNewAppLifecycleManagerFactoryWithDependencies_FactoryFunction(t *testing.T) {
	t.Parallel()

	deps := AppLifecycleManagerDependencies{
		Context:         context.Background(),
		ShutdownTimeout: "10s",
	}

	factory := NewAppLifecycleManagerFactoryWithDependencies(deps)
	if factory == nil {
		t.Fatal("NewAppLifecycleManagerFactoryWithDependencies should not return nil")
	}

	// Verify interface compliance
	_, ok := factory.(AppLifecycleManagerFactory)
	if !ok {
		t.Fatal("NewAppLifecycleManagerFactoryWithDependencies should return AppLifecycleManagerFactory interface")
	}
}

// TestNewAppLifecycleManagerFactoryWithDependencies_NilContext tests factory with nil context
func TestNewAppLifecycleManagerFactoryWithDependencies_NilContext(t *testing.T) {
	t.Parallel()

	deps := AppLifecycleManagerDependencies{
		Context:         nil, // Nil context should use default
		ShutdownTimeout: "5s",
	}

	factory := NewAppLifecycleManagerFactoryWithDependencies(deps)
	if factory == nil {
		t.Fatal("Factory should handle nil context gracefully")
	}

	// Create manager and verify it works
	manager := factory.CreateLifecycleManager()
	if manager == nil {
		t.Fatal("Factory should create valid manager even with nil context")
	}

	if manager.Context() == nil {
		t.Error("Manager should have valid context even when factory created with nil")
	}
}

// TestNewAppLifecycleManagerFactoryWithDependencies_InvalidTimeout tests factory with invalid timeout
func TestNewAppLifecycleManagerFactoryWithDependencies_InvalidTimeout(t *testing.T) {
	t.Parallel()

	deps := AppLifecycleManagerDependencies{
		Context:         context.Background(),
		ShutdownTimeout: "invalid-timeout", // Invalid timeout should use default
	}

	factory := NewAppLifecycleManagerFactoryWithDependencies(deps)
	if factory == nil {
		t.Fatal("Factory should handle invalid timeout gracefully")
	}

	// Create manager and verify it works
	manager := factory.CreateLifecycleManager()
	if manager == nil {
		t.Fatal("Factory should create valid manager even with invalid timeout")
	}
}

// TestNewAppLifecycleManagerFactoryWithDependencies_EmptyTimeout tests factory with empty timeout
func TestNewAppLifecycleManagerFactoryWithDependencies_EmptyTimeout(t *testing.T) {
	t.Parallel()

	deps := AppLifecycleManagerDependencies{
		Context:         context.Background(),
		ShutdownTimeout: "", // Empty timeout should use default
	}

	factory := NewAppLifecycleManagerFactoryWithDependencies(deps)
	if factory == nil {
		t.Fatal("Factory should handle empty timeout gracefully")
	}

	// Create manager and verify it works
	manager := factory.CreateLifecycleManager()
	if manager == nil {
		t.Fatal("Factory should create valid manager even with empty timeout")
	}
}

// TestDefaultAppLifecycleManagerFactory_CreateLifecycleManager tests manager creation
func TestDefaultAppLifecycleManagerFactory_CreateLifecycleManager(t *testing.T) {
	t.Parallel()

	factory := NewAppLifecycleManagerFactory()
	manager := factory.CreateLifecycleManager()

	if manager == nil {
		t.Fatal("CreateLifecycleManager should not return nil")
	}

	// Verify interface compliance
	_, ok := manager.(AppLifecycleManager)
	if !ok {
		t.Fatal("CreateLifecycleManager should return AppLifecycleManager interface")
	}

	// Verify initial state
	if manager.IsRunning() {
		t.Error("New manager should not be running initially")
	}

	// Verify context is available
	if manager.Context() == nil {
		t.Error("New manager should have a context")
	}

	// Verify shutdown channel is available
	if manager.ShutdownChannel() == nil {
		t.Error("New manager should have a shutdown channel")
	}
}

// TestDefaultAppLifecycleManagerFactory_CreateLifecycleManagerWithContext tests manager creation with context
func TestDefaultAppLifecycleManagerFactory_CreateLifecycleManagerWithContext(t *testing.T) {
	t.Parallel()

	factory := NewAppLifecycleManagerFactory()
	ctx := context.Background()
	manager := factory.CreateLifecycleManagerWithContext(ctx)

	if manager == nil {
		t.Fatal("CreateLifecycleManagerWithContext should not return nil")
	}

	// Verify interface compliance
	_, ok := manager.(AppLifecycleManager)
	if !ok {
		t.Fatal("CreateLifecycleManagerWithContext should return AppLifecycleManager interface")
	}

	// Verify initial state
	if manager.IsRunning() {
		t.Error("New manager should not be running initially")
	}
}

// TestDefaultAppLifecycleManagerFactory_CreateLifecycleManagerWithDefaults tests manager creation with defaults
func TestDefaultAppLifecycleManagerFactory_CreateLifecycleManagerWithDefaults(t *testing.T) {
	t.Parallel()

	factory := NewAppLifecycleManagerFactory()

	// Test that CreateLifecycleManagerWithDefaults method exists
	if concrete, ok := factory.(*DefaultAppLifecycleManagerFactory); ok {
		manager := concrete.CreateLifecycleManagerWithDefaults()
		if manager == nil {
			t.Fatal("CreateLifecycleManagerWithDefaults should not return nil")
		}

		// Verify interface compliance
		_, ok := manager.(AppLifecycleManager)
		if !ok {
			t.Fatal("CreateLifecycleManagerWithDefaults should return AppLifecycleManager interface")
		}
	} else {
		t.Error("Factory should be concrete DefaultAppLifecycleManagerFactory type")
	}
}

// TestDefaultAppLifecycleManagerFactory_TimeoutConfiguration tests timeout configuration
func TestDefaultAppLifecycleManagerFactory_TimeoutConfiguration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		timeoutString   string
		expectValid     bool
		expectedDefault bool
	}{
		{
			name:            "Valid timeout",
			timeoutString:   "5s",
			expectValid:     true,
			expectedDefault: false,
		},
		{
			name:            "Valid timeout minutes",
			timeoutString:   "2m",
			expectValid:     true,
			expectedDefault: false,
		},
		{
			name:            "Invalid timeout",
			timeoutString:   "invalid",
			expectValid:     false,
			expectedDefault: true,
		},
		{
			name:            "Empty timeout",
			timeoutString:   "",
			expectValid:     false,
			expectedDefault: true,
		},
		{
			name:            "Zero timeout",
			timeoutString:   "0s",
			expectValid:     true,
			expectedDefault: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			deps := AppLifecycleManagerDependencies{
				Context:         context.Background(),
				ShutdownTimeout: tt.timeoutString,
			}

			factory := NewAppLifecycleManagerFactoryWithDependencies(deps)
			manager := factory.CreateLifecycleManager()

			if manager == nil {
				t.Fatal("Factory should create manager regardless of timeout validity")
			}

			// Verify manager is functional
			if manager.Context() == nil {
				t.Error("Manager should have valid context")
			}
		})
	}
}

// TestDefaultAppLifecycleManagerFactory_ContextPropagation tests context propagation
func TestDefaultAppLifecycleManagerFactory_ContextPropagation(t *testing.T) {
	t.Parallel()

	// Create factory with custom context
	parentCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deps := AppLifecycleManagerDependencies{
		Context:         parentCtx,
		ShutdownTimeout: "30s",
	}

	factory := NewAppLifecycleManagerFactoryWithDependencies(deps)

	// Create manager using factory defaults (should use factory's context)
	if concrete, ok := factory.(*DefaultAppLifecycleManagerFactory); ok {
		manager := concrete.CreateLifecycleManagerWithDefaults()

		// Cancel parent context
		cancel()

		// Manager context should be cancelled
		select {
		case <-manager.Context().Done():
			// Expected - context should be cancelled
		case <-time.After(100 * time.Millisecond):
			t.Error("Manager context should be cancelled when factory context is cancelled")
		}
	} else {
		t.Error("Factory should be concrete DefaultAppLifecycleManagerFactory type")
	}
}

// TestDefaultAppLifecycleManagerFactory_MultipleManagers tests creating multiple managers
func TestDefaultAppLifecycleManagerFactory_MultipleManagers(t *testing.T) {
	t.Parallel()

	factory := NewAppLifecycleManagerFactory()

	// Create multiple managers
	manager1 := factory.CreateLifecycleManager()
	manager2 := factory.CreateLifecycleManager()
	manager3 := factory.CreateLifecycleManagerWithContext(context.Background())

	// All should be valid but independent
	managers := []AppLifecycleManager{manager1, manager2, manager3}
	for i, manager := range managers {
		if manager == nil {
			t.Fatalf("Manager %d should not be nil", i+1)
		}

		if manager.IsRunning() {
			t.Errorf("Manager %d should not be running initially", i+1)
		}

		if manager.Context() == nil {
			t.Errorf("Manager %d should have a context", i+1)
		}
	}

	// Managers should be independent - starting one shouldn't affect others
	ctx := context.Background()
	err := manager1.Startup(ctx)
	if err != nil {
		t.Fatalf("Manager1 startup should not error: %v", err)
	}

	if manager2.IsRunning() || manager3.IsRunning() {
		t.Error("Other managers should not be affected by manager1 startup")
	}

	// Cleanup
	_ = manager1.Shutdown(ctx)
}

// TestDefaultAppLifecycleManagerFactory_InterfaceCompliance tests interface compliance
func TestDefaultAppLifecycleManagerFactory_InterfaceCompliance(t *testing.T) {
	t.Parallel()

	// Test that factory implements the interface
	var factory AppLifecycleManagerFactory = NewAppLifecycleManagerFactory()

	// Test all interface methods
	manager1 := factory.CreateLifecycleManager()
	if manager1 == nil {
		t.Error("CreateLifecycleManager should not return nil")
	}

	manager2 := factory.CreateLifecycleManagerWithContext(context.Background())
	if manager2 == nil {
		t.Error("CreateLifecycleManagerWithContext should not return nil")
	}

	// Both managers should implement AppLifecycleManager
	_, ok1 := manager1.(AppLifecycleManager)
	_, ok2 := manager2.(AppLifecycleManager)

	if !ok1 || !ok2 {
		t.Error("All created managers should implement AppLifecycleManager interface")
	}
}

// TestDefaultAppLifecycleManagerFactory_EdgeCases tests various edge cases
func TestDefaultAppLifecycleManagerFactory_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func() AppLifecycleManagerFactory
		operation func(AppLifecycleManagerFactory) AppLifecycleManager
		expectNil bool
	}{
		{
			name: "Factory with nil dependencies context",
			setup: func() AppLifecycleManagerFactory {
				deps := AppLifecycleManagerDependencies{
					Context:         nil,
					ShutdownTimeout: "30s",
				}
				return NewAppLifecycleManagerFactoryWithDependencies(deps)
			},
			operation: func(f AppLifecycleManagerFactory) AppLifecycleManager {
				return f.CreateLifecycleManager()
			},
			expectNil: false,
		},
		{
			name: "Factory with zero dependencies",
			setup: func() AppLifecycleManagerFactory {
				deps := AppLifecycleManagerDependencies{}
				return NewAppLifecycleManagerFactoryWithDependencies(deps)
			},
			operation: func(f AppLifecycleManagerFactory) AppLifecycleManager {
				return f.CreateLifecycleManager()
			},
			expectNil: false,
		},
		{
			name: "Factory with cancelled context",
			setup: func() AppLifecycleManagerFactory {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately
				deps := AppLifecycleManagerDependencies{
					Context:         ctx,
					ShutdownTimeout: "30s",
				}
				return NewAppLifecycleManagerFactoryWithDependencies(deps)
			},
			operation: func(f AppLifecycleManagerFactory) AppLifecycleManager {
				return f.CreateLifecycleManager()
			},
			expectNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			factory := tt.setup()
			manager := tt.operation(factory)

			if tt.expectNil {
				if manager != nil {
					t.Error("Expected nil manager")
				}
			} else {
				if manager == nil {
					t.Error("Expected non-nil manager")
				} else {
					// Verify manager is functional
					if manager.Context() == nil {
						t.Error("Manager should have valid context")
					}
				}
			}
		})
	}
}
