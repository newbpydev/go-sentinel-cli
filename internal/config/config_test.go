package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfig_LoadFromFile(t *testing.T) {
	// Create temporary directory for test config files
	tmpDir, err := os.MkdirTemp("", "sentinel-config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test loading valid configuration file
	configPath := filepath.Join(tmpDir, "sentinel.config.json")
	configData := map[string]interface{}{
		"colors":          true,
		"icons":           "unicode",
		"watchMode":       false,
		"verbosity":       2,
		"timeout":         "30s",
		"includePatterns": []string{"./internal", "./pkg"},
		"excludePatterns": []string{"./vendor", "./node_modules"},
		"watchIgnore":     []string{"*.tmp", "*.log"},
		"testCommand":     "go test",
		"parallel":        4,
	}

	configJSON, err := json.Marshal(configData)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(configPath, configJSON, 0644)
	if err != nil {
		t.Fatal(err)
	}

	loader := &DefaultConfigLoader{}
	config, err := loader.LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !config.Colors {
		t.Errorf("expected colors=true, got %t", config.Colors)
	}
	if config.Icons != "unicode" {
		t.Errorf("expected icons='unicode', got '%s'", config.Icons)
	}
	if config.Verbosity != 2 {
		t.Errorf("expected verbosity=2, got %d", config.Verbosity)
	}
	if config.Parallel != 4 {
		t.Errorf("expected parallel=4, got %d", config.Parallel)
	}
}

func TestConfig_LoadFromFileNotFound(t *testing.T) {
	loader := &DefaultConfigLoader{}
	config, err := loader.LoadFromFile("nonexistent-config.json")

	// Should return default config when file not found
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defaultConfig := GetDefaultConfig()
	if config.Colors != defaultConfig.Colors {
		t.Errorf("expected default colors, got %t", config.Colors)
	}
}

func TestConfig_VisualStyleConfiguration(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sentinel-config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name     string
		config   map[string]interface{}
		expected VisualConfig
	}{
		{
			name: "unicode icons with colors",
			config: map[string]interface{}{
				"colors": true,
				"icons":  "unicode",
				"theme":  "dark",
			},
			expected: VisualConfig{
				Colors: true,
				Icons:  "unicode",
				Theme:  "dark",
			},
		},
		{
			name: "ascii icons without colors",
			config: map[string]interface{}{
				"colors": false,
				"icons":  "ascii",
				"theme":  "light",
			},
			expected: VisualConfig{
				Colors: false,
				Icons:  "ascii",
				Theme:  "light",
			},
		},
		{
			name: "minimal icons",
			config: map[string]interface{}{
				"colors": true,
				"icons":  "minimal",
			},
			expected: VisualConfig{
				Colors: true,
				Icons:  "minimal",
				Theme:  "dark", // default
			},
		},
	}

	loader := &DefaultConfigLoader{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := filepath.Join(tmpDir, tt.name+".json")
			configJSON, err := json.Marshal(tt.config)
			if err != nil {
				t.Fatal(err)
			}

			err = os.WriteFile(configPath, configJSON, 0644)
			if err != nil {
				t.Fatal(err)
			}

			config, err := loader.LoadFromFile(configPath)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if config.Visual.Colors != tt.expected.Colors {
				t.Errorf("expected colors=%t, got %t", tt.expected.Colors, config.Visual.Colors)
			}
			if config.Visual.Icons != tt.expected.Icons {
				t.Errorf("expected icons='%s', got '%s'", tt.expected.Icons, config.Visual.Icons)
			}
			if config.Visual.Theme != tt.expected.Theme {
				t.Errorf("expected theme='%s', got '%s'", tt.expected.Theme, config.Visual.Theme)
			}
		})
	}
}

func TestConfig_PathPatternConfiguration(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sentinel-config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "patterns.json")
	configData := map[string]interface{}{
		"includePatterns": []string{"./internal/...", "./pkg/...", "./cmd/..."},
		"excludePatterns": []string{"./vendor", ".*_test.go", "./tmp"},
		"watchIgnore":     []string{"*.log", "*.tmp", ".git/*"},
	}

	configJSON, err := json.Marshal(configData)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(configPath, configJSON, 0644)
	if err != nil {
		t.Fatal(err)
	}

	loader := &DefaultConfigLoader{}
	config, err := loader.LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedIncludes := []string{"./internal/...", "./pkg/...", "./cmd/..."}
	if len(config.Paths.IncludePatterns) != len(expectedIncludes) {
		t.Errorf("expected %d include patterns, got %d", len(expectedIncludes), len(config.Paths.IncludePatterns))
	}

	expectedExcludes := []string{"./vendor", ".*_test.go", "./tmp"}
	if len(config.Paths.ExcludePatterns) != len(expectedExcludes) {
		t.Errorf("expected %d exclude patterns, got %d", len(expectedExcludes), len(config.Paths.ExcludePatterns))
	}

	expectedWatchIgnore := []string{"*.log", "*.tmp", ".git/*"}
	if len(config.Watch.IgnorePatterns) != len(expectedWatchIgnore) {
		t.Errorf("expected %d watch ignore patterns, got %d", len(expectedWatchIgnore), len(config.Watch.IgnorePatterns))
	}
}

func TestConfig_WatchBehaviorConfiguration(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sentinel-config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "watch.json")
	configData := map[string]interface{}{
		"watchMode":     true,
		"watchDebounce": "500ms",
		"clearOnRerun":  true,
		"runOnStart":    false,
	}

	configJSON, err := json.Marshal(configData)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(configPath, configJSON, 0644)
	if err != nil {
		t.Fatal(err)
	}

	loader := &DefaultConfigLoader{}
	config, err := loader.LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !config.Watch.Enabled {
		t.Errorf("expected watch enabled=true, got %t", config.Watch.Enabled)
	}
	if config.Watch.Debounce != 500*time.Millisecond {
		t.Errorf("expected debounce=500ms, got %v", config.Watch.Debounce)
	}
	if !config.Watch.ClearOnRerun {
		t.Errorf("expected clearOnRerun=true, got %t", config.Watch.ClearOnRerun)
	}
	if config.Watch.RunOnStart {
		t.Errorf("expected runOnStart=false, got %t", config.Watch.RunOnStart)
	}
}

func TestConfig_MergeWithCLIArgs(t *testing.T) {
	config := GetDefaultConfig()
	args := &Args{
		Colors:    true,
		Verbosity: 3,
		Parallel:  8,
		Packages:  []string{"./internal", "./pkg"},
	}

	merged := config.MergeWithCLIArgs(args)

	if !merged.Colors {
		t.Errorf("expected colors=true, got %t", merged.Colors)
	}
	if merged.Verbosity != 3 {
		t.Errorf("expected verbosity=3, got %d", merged.Verbosity)
	}
	if merged.Parallel != 8 {
		t.Errorf("expected parallel=8, got %d", merged.Parallel)
	}
}

func TestConfig_ValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name:        "valid config",
			config:      GetDefaultConfig(),
			expectError: false,
		},
		{
			name: "negative verbosity",
			config: &Config{
				Verbosity: -1,
			},
			expectError: true,
		},
		{
			name: "negative parallel",
			config: &Config{
				Parallel: -1,
			},
			expectError: true,
		},
		{
			name: "negative timeout",
			config: &Config{
				Timeout: -1 * time.Second,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if tt.expectError && err == nil {
				t.Error("expected validation error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}
