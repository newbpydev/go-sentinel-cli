package config

import (
	"strings"
	"testing"
	"time"
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
					Icons: "unicode",
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

// TestNewAppConfigLoaderWithDependencies tests loader creation with custom dependencies
func TestNewAppConfigLoaderWithDependencies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		dependencies AppConfigLoaderDependencies
		validateFunc func(*testing.T, AppConfigLoader)
	}{
		{
			name: "with_all_dependencies",
			dependencies: AppConfigLoaderDependencies{
				CliLoader:      &DefaultConfigLoader{},
				ValidationMode: ValidationModeStrict,
			},
			validateFunc: func(t *testing.T, loader AppConfigLoader) {
				if loader == nil {
					t.Fatal("Loader should not be nil")
				}
				// Verify interface compliance
				_, ok := loader.(AppConfigLoader)
				if !ok {
					t.Fatal("Should implement AppConfigLoader interface")
				}
			},
		},
		{
			name: "with_nil_cli_loader",
			dependencies: AppConfigLoaderDependencies{
				CliLoader:      nil,
				ValidationMode: ValidationModeLenient,
			},
			validateFunc: func(t *testing.T, loader AppConfigLoader) {
				if loader == nil {
					t.Fatal("Loader should not be nil")
				}
				// Should use default CLI loader when nil is provided
				defaultLoader := loader.(*DefaultAppConfigLoader)
				if defaultLoader.cliLoader == nil {
					t.Error("Should have default CLI loader when nil is provided")
				}
			},
		},
		{
			name: "with_validation_mode_off",
			dependencies: AppConfigLoaderDependencies{
				CliLoader:      &DefaultConfigLoader{},
				ValidationMode: ValidationModeOff,
			},
			validateFunc: func(t *testing.T, loader AppConfigLoader) {
				if loader == nil {
					t.Fatal("Loader should not be nil")
				}
				defaultLoader := loader.(*DefaultAppConfigLoader)
				if defaultLoader.validationMode != ValidationModeOff {
					t.Error("Should preserve validation mode")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			loader := NewAppConfigLoaderWithDependencies(tt.dependencies)
			tt.validateFunc(t, loader)
		})
	}
}

// TestDefaultAppConfigLoader_LoadFromDefaults tests loading default configuration
func TestDefaultAppConfigLoader_LoadFromDefaults(t *testing.T) {
	t.Parallel()

	loader := NewAppConfigLoader()
	config := loader.LoadFromDefaults()

	if config == nil {
		t.Fatal("LoadFromDefaults should not return nil")
	}

	// Verify default values are set correctly
	if config.Watch.Enabled {
		t.Error("Watch should be disabled by default")
	}

	if !config.Colors {
		t.Error("Colors should be enabled by default")
	}

	if config.Verbosity != 0 {
		t.Errorf("Expected verbosity 0, got %d", config.Verbosity)
	}

	if config.Visual.Icons == "" {
		t.Error("Icons should have a default value")
	}

	if config.Visual.Theme == "" {
		t.Error("Theme should have a default value")
	}

	if config.Visual.TerminalWidth != 80 {
		t.Errorf("Expected terminal width 80, got %d", config.Visual.TerminalWidth)
	}
}

// TestDefaultAppConfigLoader_Merge tests configuration merging
func TestDefaultAppConfigLoader_Merge(t *testing.T) {
	t.Parallel()

	loader := NewAppConfigLoader()

	tests := []struct {
		name     string
		config   *AppConfig
		args     *AppArguments
		expected *AppConfig
	}{
		{
			name:     "nil_config",
			config:   nil,
			args:     &AppArguments{Watch: true},
			expected: nil,
		},
		{
			name:     "nil_args",
			config:   &AppConfig{Watch: AppWatchConfig{Enabled: false}},
			args:     nil,
			expected: &AppConfig{Watch: AppWatchConfig{Enabled: false}},
		},
		{
			name: "merge_watch_enabled",
			config: &AppConfig{
				Watch: AppWatchConfig{Enabled: false},
			},
			args: &AppArguments{Watch: true},
			expected: &AppConfig{
				Watch: AppWatchConfig{Enabled: true},
			},
		},
		{
			name: "merge_colors_enabled",
			config: &AppConfig{
				Colors: false,
			},
			args: &AppArguments{Colors: true},
			expected: &AppConfig{
				Colors: true,
			},
		},
		{
			name: "merge_verbose_enabled",
			config: &AppConfig{
				Verbosity: 0,
			},
			args: &AppArguments{Verbose: true},
			expected: &AppConfig{
				Verbosity: 1,
			},
		},
		{
			name: "merge_packages",
			config: &AppConfig{
				Paths: AppPathsConfig{
					IncludePatterns: []string{"old/path"},
				},
			},
			args: &AppArguments{
				Packages: []string{"new/path", "another/path"},
			},
			expected: &AppConfig{
				Paths: AppPathsConfig{
					IncludePatterns: []string{"new/path", "another/path"},
				},
			},
		},
		{
			name: "merge_all_options",
			config: &AppConfig{
				Watch:     AppWatchConfig{Enabled: false},
				Colors:    false,
				Verbosity: 0,
				Paths:     AppPathsConfig{IncludePatterns: []string{"old"}},
			},
			args: &AppArguments{
				Watch:    true,
				Colors:   true,
				Verbose:  true,
				Packages: []string{"new"},
			},
			expected: &AppConfig{
				Watch:     AppWatchConfig{Enabled: true},
				Colors:    true,
				Verbosity: 1,
				Paths:     AppPathsConfig{IncludePatterns: []string{"new"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := loader.Merge(tt.config, tt.args)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil result, got %+v", result)
				}
				return
			}

			if result == nil {
				t.Fatal("Expected non-nil result")
			}

			// Verify specific fields
			if result.Watch.Enabled != tt.expected.Watch.Enabled {
				t.Errorf("Expected watch enabled %v, got %v", tt.expected.Watch.Enabled, result.Watch.Enabled)
			}

			if result.Colors != tt.expected.Colors {
				t.Errorf("Expected colors %v, got %v", tt.expected.Colors, result.Colors)
			}

			if result.Verbosity != tt.expected.Verbosity {
				t.Errorf("Expected verbosity %d, got %d", tt.expected.Verbosity, result.Verbosity)
			}

			if len(result.Paths.IncludePatterns) != len(tt.expected.Paths.IncludePatterns) {
				t.Errorf("Expected %d include patterns, got %d", len(tt.expected.Paths.IncludePatterns), len(result.Paths.IncludePatterns))
			}
		})
	}
}

// TestDefaultAppConfigLoader_Validate tests configuration validation
func TestDefaultAppConfigLoader_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		config         *AppConfig
		validationMode ValidationMode
		expectError    bool
		errorMsg       string
	}{
		{
			name:        "nil_config",
			config:      nil,
			expectError: true,
			errorMsg:    "configuration cannot be nil",
		},
		{
			name: "validation_mode_off",
			config: &AppConfig{
				Watch: AppWatchConfig{Debounce: "invalid"},
			},
			validationMode: ValidationModeOff,
			expectError:    false,
		},
		{
			name: "valid_config",
			config: &AppConfig{
				Watch: AppWatchConfig{
					Enabled:  true,
					Debounce: "100ms",
				},
				Test: AppTestConfig{
					Timeout: "30s",
				},
				Visual: AppVisualConfig{
					Icons: "unicode",
				},
			},
			validationMode: ValidationModeStrict,
			expectError:    false,
		},
		{
			name: "invalid_debounce_strict",
			config: &AppConfig{
				Watch: AppWatchConfig{
					Enabled:  true,
					Debounce: "invalid",
				},
			},
			validationMode: ValidationModeStrict,
			expectError:    true,
			errorMsg:       "invalid debounce duration",
		},
		{
			name: "invalid_debounce_lenient",
			config: &AppConfig{
				Watch: AppWatchConfig{
					Enabled:  true,
					Debounce: "invalid",
				},
			},
			validationMode: ValidationModeLenient,
			expectError:    false,
		},
		{
			name: "invalid_timeout_strict",
			config: &AppConfig{
				Test: AppTestConfig{
					Timeout: "invalid",
				},
			},
			validationMode: ValidationModeStrict,
			expectError:    true,
			errorMsg:       "invalid timeout duration",
		},
		{
			name: "invalid_timeout_lenient",
			config: &AppConfig{
				Test: AppTestConfig{
					Timeout: "invalid",
				},
			},
			validationMode: ValidationModeLenient,
			expectError:    false,
		},
		{
			name: "invalid_icons_strict",
			config: &AppConfig{
				Visual: AppVisualConfig{
					Icons: "invalid",
				},
			},
			validationMode: ValidationModeStrict,
			expectError:    true,
			errorMsg:       "invalid icons setting",
		},
		{
			name: "invalid_icons_lenient",
			config: &AppConfig{
				Visual: AppVisualConfig{
					Icons: "invalid",
				},
			},
			validationMode: ValidationModeLenient,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			deps := AppConfigLoaderDependencies{
				CliLoader:      &DefaultConfigLoader{},
				ValidationMode: tt.validationMode,
			}
			loader := NewAppConfigLoaderWithDependencies(deps)

			err := loader.Validate(tt.config)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestDefaultAppConfigLoader_ConvertCliConfigToAppConfig tests CLI config conversion
func TestDefaultAppConfigLoader_ConvertCliConfigToAppConfig(t *testing.T) {
	t.Parallel()

	loader := NewAppConfigLoader().(*DefaultAppConfigLoader)

	// Create a sample CLI config
	cliConfig := &Config{
		Watch: WatchConfig{
			IgnorePatterns: []string{"*.tmp", "*.log"},
			Debounce:       100 * time.Millisecond,
			RunOnStart:     true,
			ClearOnRerun:   false,
		},
		Paths: PathsConfig{
			IncludePatterns: []string{"./pkg", "./internal"},
			ExcludePatterns: []string{"./vendor"},
		},
		Visual: VisualConfig{
			Icons: "unicode",
			Theme: "dark",
		},
		Timeout:   30 * time.Second,
		Parallel:  4,
		Colors:    true,
		Verbosity: 2,
	}

	appConfig := loader.convertCliConfigToAppConfig(cliConfig)

	if appConfig == nil {
		t.Fatal("Converted config should not be nil")
	}

	// Verify watch config conversion
	if appConfig.Watch.Enabled {
		t.Error("Watch should be disabled by default in app config")
	}
	if len(appConfig.Watch.IgnorePatterns) != 2 {
		t.Errorf("Expected 2 ignore patterns, got %d", len(appConfig.Watch.IgnorePatterns))
	}
	if appConfig.Watch.Debounce != "100ms" {
		t.Errorf("Expected debounce '100ms', got %q", appConfig.Watch.Debounce)
	}
	if !appConfig.Watch.RunOnStart {
		t.Error("RunOnStart should be preserved")
	}
	if appConfig.Watch.ClearOnRerun {
		t.Error("ClearOnRerun should be preserved")
	}

	// Verify paths config conversion
	if len(appConfig.Paths.IncludePatterns) != 2 {
		t.Errorf("Expected 2 include patterns, got %d", len(appConfig.Paths.IncludePatterns))
	}
	if len(appConfig.Paths.ExcludePatterns) != 1 {
		t.Errorf("Expected 1 exclude pattern, got %d", len(appConfig.Paths.ExcludePatterns))
	}

	// Verify visual config conversion
	if appConfig.Visual.Icons != "unicode" {
		t.Errorf("Expected icons 'unicode', got %q", appConfig.Visual.Icons)
	}
	if appConfig.Visual.Theme != "dark" {
		t.Errorf("Expected theme 'dark', got %q", appConfig.Visual.Theme)
	}
	if appConfig.Visual.TerminalWidth != 80 {
		t.Errorf("Expected terminal width 80, got %d", appConfig.Visual.TerminalWidth)
	}

	// Verify test config conversion
	if appConfig.Test.Timeout != "30s" {
		t.Errorf("Expected timeout '30s', got %q", appConfig.Test.Timeout)
	}
	if appConfig.Test.Parallel != 4 {
		t.Errorf("Expected parallel 4, got %d", appConfig.Test.Parallel)
	}
	if appConfig.Test.Coverage {
		t.Error("Coverage should be false by default")
	}

	// Verify other fields
	if !appConfig.Colors {
		t.Error("Colors should be preserved")
	}
	if appConfig.Verbosity != 2 {
		t.Errorf("Expected verbosity 2, got %d", appConfig.Verbosity)
	}
}

// TestDefaultAppConfigLoader_LoadFromFile_EmptyPath tests loading with empty path
func TestDefaultAppConfigLoader_LoadFromFile_EmptyPath(t *testing.T) {
	t.Parallel()

	loader := NewAppConfigLoader()

	_, err := loader.LoadFromFile("")

	if err == nil {
		t.Error("Expected error for empty path")
	}

	if !strings.Contains(err.Error(), "failed to load configuration from file") {
		t.Errorf("Expected error about loading configuration, got: %v", err)
	}
}

// TestDefaultAppConfigLoader_ValidationModeIntegration tests validation mode integration
func TestDefaultAppConfigLoader_ValidationModeIntegration(t *testing.T) {
	t.Parallel()

	// Test that validation mode affects behavior correctly
	invalidConfig := &AppConfig{
		Watch: AppWatchConfig{
			Enabled:  true,
			Debounce: "invalid",
		},
		Test: AppTestConfig{
			Timeout: "invalid",
		},
		Visual: AppVisualConfig{
			Icons: "invalid",
		},
	}

	// Test strict mode
	strictLoader := NewAppConfigLoaderWithDependencies(AppConfigLoaderDependencies{
		ValidationMode: ValidationModeStrict,
	})
	err := strictLoader.Validate(invalidConfig)
	if err == nil {
		t.Error("Strict mode should return error for invalid config")
	}

	// Test lenient mode (should fix invalid values)
	lenientLoader := NewAppConfigLoaderWithDependencies(AppConfigLoaderDependencies{
		ValidationMode: ValidationModeLenient,
	})
	lenientConfig := *invalidConfig // Copy to avoid modifying original
	err = lenientLoader.Validate(&lenientConfig)
	if err != nil {
		t.Errorf("Lenient mode should not return error: %v", err)
	}

	// Verify lenient mode fixed the values
	if lenientConfig.Watch.Debounce != "100ms" {
		t.Errorf("Expected lenient mode to fix debounce to '100ms', got %q", lenientConfig.Watch.Debounce)
	}
	if lenientConfig.Test.Timeout != "30s" {
		t.Errorf("Expected lenient mode to fix timeout to '30s', got %q", lenientConfig.Test.Timeout)
	}
	if lenientConfig.Visual.Icons != "unicode" {
		t.Errorf("Expected lenient mode to fix icons to 'unicode', got %q", lenientConfig.Visual.Icons)
	}

	// Test off mode
	offLoader := NewAppConfigLoaderWithDependencies(AppConfigLoaderDependencies{
		ValidationMode: ValidationModeOff,
	})
	err = offLoader.Validate(invalidConfig)
	if err != nil {
		t.Errorf("Off mode should not return error: %v", err)
	}
}
