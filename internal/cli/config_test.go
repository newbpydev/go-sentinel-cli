package cli

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
		"watchIgnore":   []string{"*.log", ".git/*"},
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

	expectedDebounce := 500 * time.Millisecond
	if config.Watch.Debounce != expectedDebounce {
		t.Errorf("expected debounce=%v, got %v", expectedDebounce, config.Watch.Debounce)
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

	// Test merging with CLI args
	cliArgs := &Args{
		Colors:      false, // Override config (was true)
		Verbosity:   3,     // Override config (was 0)
		Parallel:    4,     // Override config (was 0)
		Watch:       true,
		TestPattern: "TestUnit",
		Packages:    []string{"./internal"},
	}

	mergedConfig := config.MergeWithCLIArgs(cliArgs)

	// Verify CLI args override config values
	if mergedConfig.Colors != false {
		t.Errorf("expected colors=false (from CLI), got %t", mergedConfig.Colors)
	}
	if mergedConfig.Verbosity != 3 {
		t.Errorf("expected verbosity=3 (from CLI), got %d", mergedConfig.Verbosity)
	}
	if mergedConfig.Parallel != 4 {
		t.Errorf("expected parallel=4 (from CLI), got %d", mergedConfig.Parallel)
	}
	if !mergedConfig.Watch.Enabled {
		t.Error("expected watch=true (from CLI)")
	}
	if mergedConfig.TestPattern != "TestUnit" {
		t.Errorf("expected test pattern='TestUnit' (from CLI), got '%s'", mergedConfig.TestPattern)
	}

	// Verify config values that weren't overridden remain
	if mergedConfig.Visual.Icons != config.Visual.Icons {
		t.Errorf("expected icons to remain from config: %s", config.Visual.Icons)
	}
}

func TestConfig_ValidationErrors(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sentinel-config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name        string
		config      map[string]interface{}
		expectError bool
	}{
		{
			name: "invalid timeout format",
			config: map[string]interface{}{
				"timeout": "invalid-duration",
			},
			expectError: true,
		},
		{
			name: "negative parallel count",
			config: map[string]interface{}{
				"parallel": -1,
			},
			expectError: true,
		},
		{
			name: "invalid verbosity level",
			config: map[string]interface{}{
				"verbosity": 10,
			},
			expectError: true,
		},
		{
			name: "invalid icons type",
			config: map[string]interface{}{
				"icons": "invalid-type",
			},
			expectError: true,
		},
		{
			name: "valid configuration",
			config: map[string]interface{}{
				"timeout":   "30s",
				"parallel":  4,
				"verbosity": 2,
				"icons":     "unicode",
			},
			expectError: false,
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
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				// Perform basic validation on successful configs
				err = ValidateConfig(config)
				if err != nil {
					t.Errorf("config validation failed: %v", err)
				}
			}
		})
	}
}
