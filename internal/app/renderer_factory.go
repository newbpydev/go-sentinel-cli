// Package app provides factory functions for creating UI components with proper dependency injection
package app

import (
	"github.com/newbpydev/go-sentinel/internal/ui/display"
)

// DisplayRendererFactory creates display renderers for the application.
// This factory ensures proper dependency injection and follows the Factory pattern.
type DisplayRendererFactory struct{}

// NewDisplayRendererFactory creates a new factory for display renderers.
func NewDisplayRendererFactory() *DisplayRendererFactory {
	return &DisplayRendererFactory{}
}

// CreateDisplayRenderer creates a new display renderer with proper dependencies injected.
// This method converts app Configuration to UI AppConfig to maintain package boundaries.
func (f *DisplayRendererFactory) CreateDisplayRenderer(config *Configuration) (display.AppRenderer, error) {
	// Convert app Configuration to UI AppConfig
	// This conversion respects package boundaries and dependency direction
	uiConfig := f.convertToUIConfig(config)

	// Create the renderer using the UI package's factory function
	renderer := display.NewAppRenderer()

	// Configure the renderer with the converted configuration
	if err := renderer.SetConfiguration(uiConfig); err != nil {
		return nil, err
	}

	return renderer, nil
}

// CreateDisplayRendererWithDefaults creates a display renderer with default settings.
func (f *DisplayRendererFactory) CreateDisplayRendererWithDefaults() display.AppRenderer {
	renderer := display.NewAppRenderer()

	// Set up basic default configuration
	defaultConfig := &display.AppConfig{
		Colors: true, // Enable colors by default
		Visual: struct {
			Icons         string
			TerminalWidth int
		}{
			Icons:         "simple",
			TerminalWidth: 80,
		},
	}

	// Configure with defaults (ignore error since config is valid)
	_ = renderer.SetConfiguration(defaultConfig)

	return renderer
}

// convertToUIConfig converts app Configuration to UI AppConfig.
// This conversion maintains clean package boundaries and follows dependency inversion.
func (f *DisplayRendererFactory) convertToUIConfig(config *Configuration) *display.AppConfig {
	if config == nil {
		// Return sensible defaults if config is nil
		return &display.AppConfig{
			Colors: true,
			Visual: struct {
				Icons         string
				TerminalWidth int
			}{
				Icons:         "simple",
				TerminalWidth: 80,
			},
		}
	}

	// Convert app Configuration to UI AppConfig
	return &display.AppConfig{
		Colors: config.Colors,
		Visual: struct {
			Icons         string
			TerminalWidth int
		}{
			Icons:         config.Visual.Icons,
			TerminalWidth: config.Visual.TerminalWidth,
		},
	}
}

// Ensure we're following proper dependency injection patterns
var _ = (*DisplayRendererFactory)(nil)
