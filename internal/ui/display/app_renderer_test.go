package display

import (
	"bytes"
	"context"
	"testing"

	"github.com/newbpydev/go-sentinel/pkg/models"
)

// TestAppRenderer_NewAppRenderer tests the creation of a new app renderer
func TestAppRenderer_NewAppRenderer(t *testing.T) {
	t.Run("should create app renderer with default settings", func(t *testing.T) {
		renderer := NewAppRenderer()

		if renderer == nil {
			t.Fatal("expected non-nil renderer")
		}

		// Verify it implements the AppRenderer interface
		var _ AppRenderer = renderer
	})
}

// TestAppRenderer_SetConfiguration tests configuration setting
func TestAppRenderer_SetConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		config      *AppConfig
		expectError bool
	}{
		{
			name: "valid configuration",
			config: &AppConfig{
				Colors: true,
				Visual: struct {
					Icons         string
					TerminalWidth int
				}{
					Icons:         "rich",
					TerminalWidth: 120,
				},
			},
			expectError: false,
		},
		{
			name:        "nil configuration",
			config:      nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewAppRenderer()
			err := renderer.SetConfiguration(tt.config)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestAppRenderer_RenderResults tests basic result rendering
func TestAppRenderer_RenderResults(t *testing.T) {
	tests := []struct {
		name            string
		configured      bool
		contextCanceled bool
		expectError     bool
	}{
		{
			name:            "successful render with configured renderer",
			configured:      true,
			contextCanceled: false,
			expectError:     false,
		},
		{
			name:            "error when not configured",
			configured:      false,
			contextCanceled: false,
			expectError:     true,
		},
		{
			name:            "handle cancelled context",
			configured:      true,
			contextCanceled: true,
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewAppRenderer()

			// Configure if needed
			if tt.configured {
				config := &AppConfig{
					Colors: true,
					Visual: struct {
						Icons         string
						TerminalWidth int
					}{
						Icons:         "simple",
						TerminalWidth: 80,
					},
				}
				if err := renderer.SetConfiguration(config); err != nil {
					t.Fatalf("failed to configure renderer: %v", err)
				}
			}

			// Create context
			ctx := context.Background()
			if tt.contextCanceled {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel() // Cancel immediately
			}

			err := renderer.RenderResults(ctx)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestAppRenderer_RenderTestResults tests individual test result rendering
func TestAppRenderer_RenderTestResults(t *testing.T) {
	renderer := NewAppRenderer()

	// Configure the renderer
	config := &AppConfig{
		Colors: true,
		Visual: struct {
			Icons         string
			TerminalWidth int
		}{
			Icons:         "simple",
			TerminalWidth: 80,
		},
	}
	if err := renderer.SetConfiguration(config); err != nil {
		t.Fatalf("failed to configure renderer: %v", err)
	}

	// Create test results
	results := []*models.TestResult{
		{
			Name:    "TestExample",
			Package: "github.com/example/pkg",
			Status:  models.TestStatusPassed,
			Output:  []string{"=== RUN   TestExample", "--- PASS: TestExample (0.00s)"},
		},
		{
			Name:    "TestFailed",
			Package: "github.com/example/pkg",
			Status:  models.TestStatusFailed,
			Output:  []string{"=== RUN   TestFailed", "--- FAIL: TestFailed (0.01s)"},
		},
	}

	ctx := context.Background()
	err := renderer.RenderTestResults(ctx, results)

	if err != nil {
		t.Errorf("unexpected error rendering test results: %v", err)
	}
}

// TestAppRenderer_SetWriter tests writer functionality
func TestAppRenderer_SetWriter(t *testing.T) {
	renderer := NewAppRenderer()

	// Test setting a custom writer
	var buf bytes.Buffer
	renderer.SetWriter(&buf)

	writer := renderer.GetWriter()
	if writer != &buf {
		t.Error("writer was not set correctly")
	}
}

// TestAppRenderer_InterfaceCompliance ensures the renderer implements required interfaces
func TestAppRenderer_InterfaceCompliance(t *testing.T) {
	renderer := NewAppRenderer()

	// Verify interface compliance at compile time
	var _ AppRenderer = renderer

	// Verify the renderer has expected methods
	if renderer.GetWriter() == nil {
		t.Error("GetWriter() should not return nil")
	}
}
