package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/go-sentinel/pkg/models"
)

// TestDefaultConfigLoader_LoadFromDefault tests the LoadFromDefault method
func TestDefaultConfigLoader_LoadFromDefault(t *testing.T) {
	t.Parallel()

	loader := &DefaultConfigLoader{}
	config, err := loader.LoadFromDefault()

	// The function should either return a config or an error, but not panic
	if err != nil {
		// Expected if sentinel.config.json doesn't exist
		if !strings.Contains(err.Error(), "no such file") && !strings.Contains(err.Error(), "cannot find") {
			t.Errorf("Unexpected error type: %v", err)
		}
	} else {
		// If no error, config should not be nil
		if config == nil {
			t.Fatal("LoadFromDefault should not return nil config when no error")
		}
	}
}

// TestConfig_MergeWithCLIArgs_Comprehensive tests the MergeWithCLIArgs method comprehensively
func TestConfig_MergeWithCLIArgs_Comprehensive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		config   *Config
		args     *Args
		validate func(*testing.T, *Config)
	}{
		{
			name: "merge_colors_override",
			config: &Config{
				Colors: true,
				Visual: VisualConfig{Colors: true},
			},
			args: &Args{
				Colors: false,
			},
			validate: func(t *testing.T, merged *Config) {
				if merged.Colors != false {
					t.Errorf("Expected colors false, got %v", merged.Colors)
				}
				if merged.Visual.Colors != false {
					t.Errorf("Expected visual colors false, got %v", merged.Visual.Colors)
				}
			},
		},
		{
			name: "merge_verbosity_override",
			config: &Config{
				Verbosity: 0,
			},
			args: &Args{
				Verbosity: 3,
			},
			validate: func(t *testing.T, merged *Config) {
				if merged.Verbosity != 3 {
					t.Errorf("Expected verbosity 3, got %d", merged.Verbosity)
				}
			},
		},
		{
			name: "merge_watch_enable",
			config: &Config{
				Watch: WatchConfig{Enabled: false},
			},
			args: &Args{
				Watch: true,
			},
			validate: func(t *testing.T, merged *Config) {
				if !merged.Watch.Enabled {
					t.Error("Expected watch enabled to be true")
				}
			},
		},
		{
			name: "merge_test_pattern",
			config: &Config{
				TestPattern: "",
			},
			args: &Args{
				TestPattern: "TestExample",
			},
			validate: func(t *testing.T, merged *Config) {
				if merged.TestPattern != "TestExample" {
					t.Errorf("Expected test pattern 'TestExample', got %q", merged.TestPattern)
				}
			},
		},
		{
			name: "merge_timeout",
			config: &Config{
				Timeout: 5 * time.Minute,
			},
			args: &Args{
				Timeout: "10m",
			},
			validate: func(t *testing.T, merged *Config) {
				if merged.Timeout != 10*time.Minute {
					t.Errorf("Expected timeout 10m, got %v", merged.Timeout)
				}
			},
		},
		{
			name: "merge_invalid_timeout_ignored",
			config: &Config{
				Timeout: 5 * time.Minute,
			},
			args: &Args{
				Timeout: "invalid",
			},
			validate: func(t *testing.T, merged *Config) {
				if merged.Timeout != 5*time.Minute {
					t.Errorf("Expected timeout unchanged at 5m, got %v", merged.Timeout)
				}
			},
		},
		{
			name: "merge_parallel_override",
			config: &Config{
				Parallel: 2,
			},
			args: &Args{
				Parallel: 8,
			},
			validate: func(t *testing.T, merged *Config) {
				if merged.Parallel != 8 {
					t.Errorf("Expected parallel 8, got %d", merged.Parallel)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			merged := tt.config.MergeWithCLIArgs(tt.args)

			if merged == nil {
				t.Fatal("MergeWithCLIArgs should not return nil")
			}

			// Verify it's a copy, not the same instance
			if merged == tt.config {
				t.Error("MergeWithCLIArgs should return a copy, not the same instance")
			}

			tt.validate(t, merged)
		})
	}
}

// TestGetDefaultConfig tests the GetDefaultConfig function
func TestGetDefaultConfig(t *testing.T) {
	t.Parallel()

	config := GetDefaultConfig()

	if config == nil {
		t.Fatal("GetDefaultConfig should not return nil")
	}

	// Verify default values
	if !config.Colors {
		t.Error("Default colors should be true")
	}

	if config.Verbosity != 0 {
		t.Errorf("Default verbosity should be 0, got %d", config.Verbosity)
	}

	if config.Parallel != 0 {
		t.Errorf("Default parallel should be 0, got %d", config.Parallel)
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("Default timeout should be 30s, got %v", config.Timeout)
	}

	if config.Visual.Icons != "unicode" {
		t.Errorf("Default icons should be 'unicode', got %q", config.Visual.Icons)
	}

	if config.Visual.Theme != "dark" {
		t.Errorf("Default theme should be 'dark', got %q", config.Visual.Theme)
	}

	if config.Watch.Enabled {
		t.Error("Default watch should be disabled")
	}

	if config.Watch.Debounce != 250*time.Millisecond {
		t.Errorf("Default watch debounce should be 250ms, got %v", config.Watch.Debounce)
	}

	if !config.Watch.RunOnStart {
		t.Error("Default watch RunOnStart should be true")
	}

	if !config.Watch.ClearOnRerun {
		t.Error("Default watch ClearOnRerun should be true")
	}
}

// TestNewConfigLoader tests the NewConfigLoader factory function
func TestNewConfigLoader(t *testing.T) {
	t.Parallel()

	loader := NewConfigLoader()

	if loader == nil {
		t.Fatal("NewConfigLoader should not return nil")
	}

	// Verify interface compliance
	_, ok := loader.(ConfigLoader)
	if !ok {
		t.Fatal("NewConfigLoader should return ConfigLoader interface")
	}

	// Verify it's the correct implementation
	_, ok = loader.(*DefaultConfigLoader)
	if !ok {
		t.Fatal("NewConfigLoader should return *DefaultConfigLoader")
	}
}

// TestDefaultConfigLoader_LoadFromFile_CompleteConfig tests loading a complete config file
func TestDefaultConfigLoader_LoadFromFile_CompleteConfig(t *testing.T) {
	t.Parallel()

	// Create a comprehensive config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "complete.json")
	configData := `{
		"colors": false,
		"verbosity": 2,
		"timeout": "30s",
		"parallel": 4,
		"icons": "ascii",
		"theme": "dark",
		"watchMode": true,
		"watchIgnore": ["*.tmp", "*.log"],
		"watchDebounce": "200ms",
		"runOnStart": true,
		"clearOnRerun": false,
		"includePatterns": ["./pkg", "./cmd"],
		"excludePatterns": ["./vendor"],
		"testCommand": "go test -v"
	}`

	err := os.WriteFile(configFile, []byte(configData), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	loader := &DefaultConfigLoader{}
	config, err := loader.LoadFromFile(configFile)

	if err != nil {
		t.Errorf("LoadFromFile should not error: %v", err)
	}

	if config == nil {
		t.Fatal("LoadFromFile should not return nil config")
	}

	// Verify all settings were parsed correctly
	if config.Colors {
		t.Error("Colors should be false")
	}
	if config.Verbosity != 2 {
		t.Errorf("Expected verbosity 2, got %d", config.Verbosity)
	}
	if config.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", config.Timeout)
	}
	if config.Parallel != 4 {
		t.Errorf("Expected parallel 4, got %d", config.Parallel)
	}
	if config.Visual.Icons != "ascii" {
		t.Errorf("Expected icons 'ascii', got %q", config.Visual.Icons)
	}
	if config.Visual.Theme != "dark" {
		t.Errorf("Expected theme 'dark', got %q", config.Visual.Theme)
	}
	if !config.Watch.Enabled {
		t.Error("Watch should be enabled")
	}
	if len(config.Watch.IgnorePatterns) != 2 {
		t.Errorf("Expected 2 ignore patterns, got %d", len(config.Watch.IgnorePatterns))
	}
	if config.Watch.Debounce != 200*time.Millisecond {
		t.Errorf("Expected debounce 200ms, got %v", config.Watch.Debounce)
	}
	if !config.Watch.RunOnStart {
		t.Error("RunOnStart should be true")
	}
	if config.Watch.ClearOnRerun {
		t.Error("ClearOnRerun should be false")
	}
	if len(config.Paths.IncludePatterns) != 2 {
		t.Errorf("Expected 2 include patterns, got %d", len(config.Paths.IncludePatterns))
	}
	if len(config.Paths.ExcludePatterns) != 1 {
		t.Errorf("Expected 1 exclude pattern, got %d", len(config.Paths.ExcludePatterns))
	}
	if config.TestCommand != "go test -v" {
		t.Errorf("Expected test command 'go test -v', got %q", config.TestCommand)
	}
}

// TestDefaultConfigLoader_LoadFromFile_InvalidValues tests handling of invalid config values
func TestDefaultConfigLoader_LoadFromFile_InvalidValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		configData  string
		expectError bool
		errorMsg    string
	}{
		{
			name: "invalid_verbosity_negative",
			configData: `{
				"verbosity": -1
			}`,
			expectError: true,
			errorMsg:    "verbosity level must be between 0 and 5",
		},
		{
			name: "invalid_verbosity_high",
			configData: `{
				"verbosity": 6
			}`,
			expectError: true,
			errorMsg:    "verbosity level must be between 0 and 5",
		},
		{
			name: "invalid_parallel_negative",
			configData: `{
				"parallel": -1
			}`,
			expectError: true,
			errorMsg:    "parallel count cannot be negative",
		},
		{
			name: "invalid_timeout_format",
			configData: `{
				"timeout": "invalid-duration"
			}`,
			expectError: true,
			errorMsg:    "invalid timeout format",
		},
		{
			name: "invalid_icons",
			configData: `{
				"icons": "invalid"
			}`,
			expectError: true,
			errorMsg:    "invalid icons type",
		},
		{
			name: "invalid_watch_debounce",
			configData: `{
				"watchDebounce": "invalid-duration"
			}`,
			expectError: true,
			errorMsg:    "invalid watch debounce format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			configFile := filepath.Join(tempDir, "invalid.json")
			err := os.WriteFile(configFile, []byte(tt.configData), 0644)
			if err != nil {
				t.Fatalf("Failed to write test config file: %v", err)
			}

			loader := &DefaultConfigLoader{}
			_, err = loader.LoadFromFile(configFile)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.expectError && err != nil && tt.errorMsg != "" {
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q, got: %v", tt.errorMsg, err)
				}
			}
		})
	}
}

// TestConvertPackagesToWatchPaths_Comprehensive tests edge cases for convertPackagesToWatchPaths
func TestConvertPackagesToWatchPaths_Comprehensive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		packages []string
		expected []string
	}{
		{
			name:     "empty_packages",
			packages: []string{},
			expected: []string{},
		},
		{
			name:     "single_dot",
			packages: []string{"."},
			expected: []string{"."},
		},
		{
			name:     "triple_dot_pattern",
			packages: []string{"..."},
			expected: []string{"..."},
		},
		{
			name:     "relative_paths",
			packages: []string{"./pkg", "../other"},
			expected: []string{"./pkg", "../other"},
		},
		{
			name:     "absolute_paths",
			packages: []string{"/usr/local/go", "/home/user/project"},
			expected: []string{"/usr/local/go", "/home/user/project"},
		},
		{
			name:     "mixed_patterns",
			packages: []string{"./pkg", "...", "./cmd"},
			expected: []string{"./pkg", "...", "./cmd"},
		},
		{
			name:     "duplicates_with_conversion",
			packages: []string{"./pkg", "...", "./pkg", "..."},
			expected: []string{"./pkg", "..."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := convertPackagesToWatchPaths(tt.packages)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d paths, got %d", len(tt.expected), len(result))
			}

			for i, expected := range tt.expected {
				if i >= len(result) {
					t.Errorf("Missing expected path %q at index %d", expected, i)
					continue
				}
				if result[i] != expected {
					t.Errorf("Expected path %q at index %d, got %q", expected, i, result[i])
				}
			}
		})
	}
}

// TestValidateConfig tests the ValidateConfig function
func TestValidateConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_config",
			config: &Config{
				Colors:    true,
				Verbosity: 1,
				Parallel:  4,
				Timeout:   5 * time.Minute,
				Visual: VisualConfig{
					Icons: "unicode",
					Theme: "dark",
				},
			},
			expectError: false,
		},
		{
			name: "negative_verbosity",
			config: &Config{
				Verbosity: -1,
			},
			expectError: true,
			errorMsg:    "verbosity level must be between 0 and 5",
		},
		{
			name: "high_verbosity",
			config: &Config{
				Verbosity: 6,
			},
			expectError: true,
			errorMsg:    "verbosity level must be between 0 and 5",
		},
		{
			name: "negative_parallel",
			config: &Config{
				Parallel: -1,
			},
			expectError: true,
			errorMsg:    "parallel count cannot be negative",
		},
		{
			name: "invalid_icons",
			config: &Config{
				Timeout: 1 * time.Second, // Valid timeout to avoid timeout error
				Visual: VisualConfig{
					Icons: "invalid",
				},
			},
			expectError: true,
			errorMsg:    "invalid icons type",
		},
		{
			name: "zero_timeout",
			config: &Config{
				Timeout: 0,
			},
			expectError: true,
			errorMsg:    "timeout must be positive",
		},
		{
			name: "negative_timeout",
			config: &Config{
				Timeout: -1 * time.Second,
			},
			expectError: true,
			errorMsg:    "timeout must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateConfig(tt.config)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.expectError && err != nil && tt.errorMsg != "" {
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q, got: %v", tt.errorMsg, err)
				}
			}
		})
	}
}

// TestDefaultConfigLoader_LoadFromFile_ErrorHandling tests error handling in LoadFromFile
func TestDefaultConfigLoader_LoadFromFile_ErrorHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupFile   func(*testing.T) string
		expectError bool
	}{
		{
			name: "invalid_json",
			setupFile: func(t *testing.T) string {
				tempDir := t.TempDir()
				configFile := filepath.Join(tempDir, "invalid.json")
				err := os.WriteFile(configFile, []byte(`{invalid json}`), 0644)
				if err != nil {
					t.Fatalf("Failed to write test file: %v", err)
				}
				return configFile
			},
			expectError: true,
		},
		{
			name: "empty_file",
			setupFile: func(t *testing.T) string {
				tempDir := t.TempDir()
				configFile := filepath.Join(tempDir, "empty.json")
				err := os.WriteFile(configFile, []byte(""), 0644)
				if err != nil {
					t.Fatalf("Failed to write test file: %v", err)
				}
				return configFile
			},
			expectError: true,
		},
		{
			name: "non_existent_file",
			setupFile: func(t *testing.T) string {
				return "/non/existent/path/config.json"
			},
			expectError: false, // CLI loader returns default config for non-existent files
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			loader := &DefaultConfigLoader{}
			configFile := tt.setupFile(t)

			_, err := loader.LoadFromFile(configFile)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestValidateConfig_EdgeCases tests additional edge cases for ValidateConfig
func TestValidateConfig_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "invalid_theme",
			config: &Config{
				Timeout: 1 * time.Second, // Valid timeout to avoid timeout error
				Visual: VisualConfig{
					Icons: "unicode", // Valid icons
					Theme: "invalid",
				},
			},
			expectError: true,
			errorMsg:    "invalid theme: invalid",
		},
		{
			name: "zero_timeout_edge_case",
			config: &Config{
				Timeout: 0,
				Visual: VisualConfig{
					Icons: "unicode", // Valid icons
					Theme: "dark",    // Valid theme
				},
			},
			expectError: true,
			errorMsg:    "timeout must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateConfig(tt.config)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.expectError && err != nil && tt.errorMsg != "" {
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q, got: %v", tt.errorMsg, err)
				}
			}
		})
	}
}

// TestAppConfigLoader_LoadFromFile_ErrorCoverage tests error paths in app config loader
func TestAppConfigLoader_LoadFromFile_ErrorCoverage(t *testing.T) {
	t.Parallel()

	loader := NewAppConfigLoader()

	// Test with non-existent file - CLI loader returns default config, not error
	config, err := loader.LoadFromFile("/non/existent/path/config.json")
	if err != nil {
		t.Errorf("LoadFromFile should not error for non-existent file: %v", err)
	}

	if config == nil {
		t.Error("LoadFromFile should return default config for non-existent file")
	}

	// Test with empty path - this should error
	_, err = loader.LoadFromFile("")
	if err == nil {
		t.Error("Expected error when loading with empty path")
	}

	// Verify error is wrapped properly
	sentinelErr, ok := err.(*models.SentinelError)
	if !ok {
		t.Fatalf("Expected SentinelError, got %T", err)
	}

	if sentinelErr.Type != models.ErrorTypeValidation {
		t.Errorf("Expected error type %v, got %v", models.ErrorTypeValidation, sentinelErr.Type)
	}
}
