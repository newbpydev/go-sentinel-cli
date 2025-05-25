package watch

import (
	"testing"

	"github.com/newbpydev/go-sentinel/internal/watch/coordinator"
	"github.com/newbpydev/go-sentinel/internal/watch/core"
)

func TestWatchCoordinator_Creation(t *testing.T) {
	options := core.WatchOptions{
		Paths: []string{"."},
		Mode:  core.WatchAll,
	}

	coordinator, err := coordinator.NewTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("Failed to create TestWatchCoordinator: %v", err)
	}

	if coordinator == nil {
		t.Error("Expected coordinator to be created")
	}
}

func TestWatchCoordinator_WithDifferentModes(t *testing.T) {
	testCases := []struct {
		name string
		mode core.WatchMode
	}{
		{
			name: "all mode",
			mode: core.WatchAll,
		},
		{
			name: "changed mode",
			mode: core.WatchChanged,
		},
		{
			name: "related mode",
			mode: core.WatchRelated,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			options := core.WatchOptions{
				Paths: []string{"."},
				Mode:  tc.mode,
			}

			coordinator, err := coordinator.NewTestWatchCoordinator(options)
			if err != nil {
				t.Fatalf("Failed to create TestWatchCoordinator: %v", err)
			}

			if coordinator == nil {
				t.Error("Expected coordinator to be created")
			}

			// Test configuration with the new mode
			newOptions := core.WatchOptions{
				Paths: []string{"."},
				Mode:  tc.mode,
			}

			err = coordinator.Configure(newOptions)
			if err != nil {
				t.Errorf("Failed to configure coordinator with mode %s: %v", tc.mode, err)
			}
		})
	}
}

func TestWatchCoordinator_Configuration(t *testing.T) {
	// Test basic coordinator creation and configuration
	options := core.WatchOptions{
		Paths:          []string{".", "./internal"},
		Mode:           core.WatchAll,
		IgnorePatterns: []string{"*.tmp", "*.log"},
		TestPatterns:   []string{"*_test.go"},
	}

	coordinator, err := coordinator.NewTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("Failed to create TestWatchCoordinator: %v", err)
	}

	if coordinator == nil {
		t.Fatal("Expected coordinator to be created")
	}

	// Test reconfiguration
	newOptions := core.WatchOptions{
		Paths: []string{"./pkg"},
		Mode:  core.WatchChanged,
	}

	err = coordinator.Configure(newOptions)
	if err != nil {
		t.Errorf("Failed to reconfigure coordinator: %v", err)
	}
}

func TestWatchCoordinator_ModeValidation(t *testing.T) {
	validModes := []core.WatchMode{
		core.WatchAll,
		core.WatchChanged,
		core.WatchRelated,
	}

	for _, mode := range validModes {
		t.Run(string(mode), func(t *testing.T) {
			options := core.WatchOptions{
				Paths: []string{"."},
				Mode:  mode,
			}

			coordinator, err := coordinator.NewTestWatchCoordinator(options)
			if err != nil {
				t.Errorf("Failed to create coordinator with valid mode %s: %v", mode, err)
			}

			if coordinator == nil {
				t.Errorf("Expected coordinator to be created for mode %s", mode)
			}
		})
	}
}

func TestWatchCoordinator_PathConfiguration(t *testing.T) {
	testCases := []struct {
		name  string
		paths []string
		valid bool
	}{
		{
			name:  "single path",
			paths: []string{"."},
			valid: true,
		},
		{
			name:  "multiple paths",
			paths: []string{".", "./internal", "./pkg"},
			valid: true,
		},
		{
			name:  "empty paths",
			paths: []string{},
			valid: true, // Should use default
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			options := core.WatchOptions{
				Paths: tc.paths,
				Mode:  core.WatchAll,
			}

			coordinator, err := coordinator.NewTestWatchCoordinator(options)

			if tc.valid {
				if err != nil {
					t.Errorf("Expected valid configuration to succeed, got error: %v", err)
				}
				if coordinator == nil {
					t.Error("Expected coordinator to be created")
				}
			} else {
				if err == nil {
					t.Error("Expected invalid configuration to fail")
				}
			}
		})
	}
}

func TestWatchCoordinator_StatusMessages(t *testing.T) {
	options := core.WatchOptions{
		Paths: []string{"."},
		Mode:  core.WatchAll,
	}

	coordinator, err := coordinator.NewTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("Failed to create TestWatchCoordinator: %v", err)
	}

	if coordinator == nil {
		t.Fatal("Expected coordinator to be created")
	}

	// Test that we can create and configure the coordinator
	// The actual UI output testing would need access to the internal methods
	// or a way to capture the output, which isn't available in the current API

	// For now, we just verify the coordinator is functional
	t.Log("TestWatchCoordinator created successfully")
}
