package config

import (
	"testing"
)

// TestAppConfigLoader_NewAppConfigLoader tests the creation of a new app config loader
func TestAppConfigLoader_NewAppConfigLoader(t *testing.T) {
	t.Run("should create app config loader with proper initialization", func(t *testing.T) {
		loader := NewAppConfigLoader()

		if loader == nil {
			t.Fatal("expected non-nil loader")
		}

		// Verify it implements the AppConfigLoader interface
		var _ AppConfigLoader = loader
	})
}

// TestAppConfigLoader_LoadFromFile tests configuration loading from file
func TestAppConfigLoader_LoadFromFile(t *testing.T) {
	tests := []struct {
		name        string
		configPath  string
		expectError bool
	}{
		{
			name:        "valid config file",
			configPath:  "sentinel.config.json",
			expectError: false,
		},
		{
			name:        "non-existent file returns default config",
			configPath:  "non-existent.json",
			expectError: false, // Config loader returns default config for missing files
		},
		{
			name:        "empty path",
			configPath:  "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewAppConfigLoader()
			config, err := loader.LoadFromFile(tt.configPath)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectError && config == nil {
				t.Error("expected non-nil config")
			}
		})
	}
}

// TestAppConfigLoader_LoadFromDefaults tests default configuration loading
func TestAppConfigLoader_LoadFromDefaults(t *testing.T) {
	t.Run("should load default configuration", func(t *testing.T) {
		loader := NewAppConfigLoader()
		config := loader.LoadFromDefaults()

		if config == nil {
			t.Fatal("expected non-nil config")
		}

		// Verify default values
		if config.Visual.Icons == "" {
			t.Error("expected default icons setting")
		}

		if config.Visual.TerminalWidth <= 0 {
			t.Error("expected positive terminal width")
		}
	})
}

// TestAppConfigLoader_Merge tests configuration merging with CLI arguments
func TestAppConfigLoader_Merge(t *testing.T) {
	tests := []struct {
		name           string
		config         *AppConfig
		args           *AppArguments
		expectedColors bool
		expectedWatch  bool
	}{
		{
			name: "merge with watch enabled",
			config: &AppConfig{
				Colors: false,
				Watch:  AppWatchConfig{Enabled: false},
			},
			args: &AppArguments{
				Watch:  true,
				Colors: true,
			},
			expectedColors: true,
			expectedWatch:  true,
		},
		{
			name: "merge with nil args",
			config: &AppConfig{
				Colors: true,
				Watch:  AppWatchConfig{Enabled: false},
			},
			args:           nil,
			expectedColors: true,
			expectedWatch:  false,
		},
		{
			name:           "merge with nil config",
			config:         nil,
			args:           &AppArguments{Watch: true},
			expectedColors: false,
			expectedWatch:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewAppConfigLoader()
			merged := loader.Merge(tt.config, tt.args)

			if tt.config == nil {
				if merged != nil {
					t.Error("expected nil when config is nil")
				}
				return
			}

			if merged == nil {
				t.Fatal("expected non-nil merged config")
			}

			if merged.Colors != tt.expectedColors {
				t.Errorf("expected colors %v, got %v", tt.expectedColors, merged.Colors)
			}

			if merged.Watch.Enabled != tt.expectedWatch {
				t.Errorf("expected watch %v, got %v", tt.expectedWatch, merged.Watch.Enabled)
			}
		})
	}
}

// TestAppConfigLoader_Validate tests configuration validation
func TestAppConfigLoader_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *AppConfig
		expectError bool
	}{
		{
			name: "valid configuration",
			config: &AppConfig{
				Watch: AppWatchConfig{
					Enabled:  true,
					Debounce: "100ms",
				},
				Test: AppTestConfig{
					Timeout: "30s",
				},
				Visual: AppVisualConfig{
					Icons: "simple",
				},
			},
			expectError: false,
		},
		{
			name:        "nil configuration",
			config:      nil,
			expectError: true,
		},
		{
			name: "invalid debounce duration",
			config: &AppConfig{
				Watch: AppWatchConfig{
					Enabled:  true,
					Debounce: "invalid",
				},
			},
			expectError: true,
		},
		{
			name: "invalid timeout duration",
			config: &AppConfig{
				Test: AppTestConfig{
					Timeout: "invalid",
				},
			},
			expectError: true,
		},
		{
			name: "invalid icons setting",
			config: &AppConfig{
				Visual: AppVisualConfig{
					Icons: "invalid",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewAppConfigLoader()
			err := loader.Validate(tt.config)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestAppConfigLoader_InterfaceCompliance ensures the loader implements required interfaces
func TestAppConfigLoader_InterfaceCompliance(t *testing.T) {
	loader := NewAppConfigLoader()

	// Verify interface compliance at compile time
	var _ AppConfigLoader = loader
}
