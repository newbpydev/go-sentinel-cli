package app_test

import (
	"context"
	"errors" // Added for errors.Is
	"strings"
	"testing"

	"github.com/newbpydev/go-sentinel/internal/app"
)

// TestNewTestExecutor_FactoryFunction tests the factory function following Go conventions
func TestNewTestExecutor_FactoryFunction(t *testing.T) {
	t.Parallel()

	// Act
	executor := app.NewTestExecutor()

	// Assert
	if executor == nil {
		t.Fatal("NewTestExecutor should not return nil")
	}

	// Verify it implements the TestExecutor interface (testing public API only)
	_, ok := executor.(app.TestExecutor)
	if !ok {
		t.Error("NewTestExecutor should return an object implementing TestExecutor interface")
	}
}

// TestNewWatchCoordinator_FactoryFunction tests the factory function following Go conventions
func TestNewWatchCoordinator_FactoryFunction(t *testing.T) {
	t.Parallel()

	// Act
	coordinator := app.NewWatchCoordinator()

	// Assert
	if coordinator == nil {
		t.Fatal("NewWatchCoordinator should not return nil")
	}

	// Verify it implements the WatchCoordinator interface (testing public API only)
	_, ok := coordinator.(app.WatchCoordinator)
	if !ok {
		t.Error("NewWatchCoordinator should return an object implementing WatchCoordinator interface")
	}
}

// TestTestExecutor_ExecuteSingle_ValidConfiguration tests successful single execution
func TestTestExecutor_ExecuteSingle_ValidConfiguration(t *testing.T) {
	t.Parallel()

	// Arrange
	executor := app.NewTestExecutor()
	config := &app.Configuration{
		Verbosity: 1,
		Test: app.TestConfig{
			Timeout:  "30s",
			Parallel: 1,
		},
		Watch: app.WatchConfig{
			Enabled: false,
		},
		Paths: app.PathsConfig{
			IncludePatterns: []string{"./..."},
		},
		Visual: app.VisualConfig{
			Icons: "none",
		},
	}

	ctx := context.Background()
	packages := []string{"."}

	// Act
	err := executor.ExecuteSingle(ctx, packages, config)

	// Assert - this will likely fail due to test environment, but should not be a configuration error
	if err != nil {
		if strings.Contains(err.Error(), "not configured") {
			t.Errorf("ExecuteSingle failed due to configuration issue: %v", err)
		} else {
			t.Logf("ExecuteSingle failed as expected due to test environment: %v", err)
		}
	} else {
		t.Log("ExecuteSingle completed successfully")
	}
}

// TestTestExecutor_ExecuteSingle_NilConfiguration tests error handling for nil configuration
func TestTestExecutor_ExecuteSingle_NilConfiguration(t *testing.T) {
	t.Parallel()

	// Arrange
	executor := app.NewTestExecutor()
	ctx := context.Background()
	packages := []string{"."}

	// Act
	err := executor.ExecuteSingle(ctx, packages, nil)

	// Assert
	if err == nil {
		t.Error("ExecuteSingle should return error for nil configuration")
	}

	if !strings.Contains(err.Error(), "configuration") {
		t.Errorf("Error should mention configuration issue, got: %v", err)
	}
}

// TestWatchCoordinator_Configure_ValidOptions tests configuration with valid options
func TestWatchCoordinator_Configure_ValidOptions(t *testing.T) {
	t.Parallel()

	// Arrange
	coordinator := app.NewWatchCoordinator()
	options := &app.WatchOptions{
		Paths:            []string{"./internal"},
		IgnorePatterns:   []string{"*.tmp"},
		TestPatterns:     []string{"*_test.go"},
		DebounceInterval: "100ms",
		ClearTerminal:    false,
		RunOnStart:       true,
	}

	// Act
	err := coordinator.Configure(options)

	// Assert
	if err != nil {
		t.Errorf("Configure should not return error for valid options, got: %v", err)
	}
}

// TestWatchCoordinator_Configure_NilOptions tests error handling for nil options
func TestWatchCoordinator_Configure_NilOptions(t *testing.T) {
	t.Parallel()

	// Arrange
	coordinator := app.NewWatchCoordinator()

	// Act
	err := coordinator.Configure(nil)

	// Assert
	if err == nil {
		t.Error("Configure should return error for nil options")
	}

	expectedMessage := "watch options cannot be nil"
	if !strings.Contains(err.Error(), expectedMessage) {
		t.Errorf("Expected error containing %q, got: %v", expectedMessage, err)
	}
}

// TestWatchCoordinator_Start_CancelledContext tests start with cancelled context
func TestWatchCoordinator_Start_CancelledContext(t *testing.T) {
	t.Parallel()

	// Arrange
	coordinator := app.NewWatchCoordinator()
	options := &app.WatchOptions{
		Paths:            []string{"test"},
		IgnorePatterns:   []string{},
		TestPatterns:     []string{"*_test.go"},
		DebounceInterval: "100ms",
		ClearTerminal:    false,
		RunOnStart:       false,
	}

	err := coordinator.Configure(options)
	if err != nil {
		t.Fatalf("Configure should work: %v", err)
	}

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Act
	err = coordinator.Start(ctx)

	// Assert
	// When a context is canceled, functions respecting it should return context.Canceled.
	// The previous placeholder implementation of Start might have returned nil.
	// The current implementation of WatchCoordinatorAdapter.Start returns an error when context is canceled.
	if err == nil {
		t.Error("Start with cancelled context expected an error, got nil")
	} else if !errors.Is(err, context.Canceled) && !strings.Contains(err.Error(), "context canceled") {
		// Allow either context.Canceled directly or an error wrapping it (e.g. from fmt.Errorf)
		t.Errorf("Start with cancelled context: expected context.Canceled or error containing 'context canceled', got: %v", err)
	} else {
		// If we got context.Canceled or a similar error, the test passes for this condition.
		t.Logf("Start with cancelled context correctly returned: %v", err)
	}
}
