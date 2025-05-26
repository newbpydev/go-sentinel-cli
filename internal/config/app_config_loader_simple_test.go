package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/newbpydev/go-sentinel/pkg/models"
)

// TestNewAppConfigLoader_Factory tests the factory function
func TestNewAppConfigLoader_Factory(t *testing.T) {
	t.Parallel()

	loader := NewAppConfigLoader()
	if loader == nil {
		t.Fatal("NewAppConfigLoader should not return nil")
	}

	// Verify interface compliance
	_, ok := loader.(AppConfigLoader)
	if !ok {
		t.Fatal("NewAppConfigLoader should return AppConfigLoader interface")
	}

	// Verify it's the correct implementation
	_, ok = loader.(*DefaultAppConfigLoader)
	if !ok {
		t.Fatal("NewAppConfigLoader should return *DefaultAppConfigLoader")
	}
}

// TestNewAppConfigLoaderWithDependencies_DependencyInjection tests dependency injection
func TestNewAppConfigLoaderWithDependencies_DependencyInjection(t *testing.T) {
	t.Parallel()

	deps := AppConfigLoaderDependencies{
		CliLoader:      &DefaultConfigLoader{},
		ValidationMode: ValidationModeStrict,
	}

	loader := NewAppConfigLoaderWithDependencies(deps)
	if loader == nil {
		t.Fatal("NewAppConfigLoaderWithDependencies should not return nil")
	}

	defaultLoader := loader.(*DefaultAppConfigLoader)
	if defaultLoader.cliLoader == nil {
		t.Error("CLI loader should be set")
	}
	if defaultLoader.validationMode != ValidationModeStrict {
		t.Error("Validation mode should be set")
	}
}

// TestDefaultAppConfigLoader_LoadFromFile tests the LoadFromFile method
func TestDefaultAppConfigLoader_LoadFromFile(t *testing.T) {
	t.Parallel()

	// Create temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.json")
	configData := `{
		"colors": false,
		"verbosity": 2
	}`

	err := os.WriteFile(configFile, []byte(configData), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	loader := NewAppConfigLoader()
	config, err := loader.LoadFromFile(configFile)

	if err != nil {
		t.Errorf("LoadFromFile should not error: %v", err)
	}

	if config == nil {
		t.Fatal("LoadFromFile should not return nil config")
	}

	if config.Colors {
		t.Error("Colors should be false")
	}
	if config.Verbosity != 2 {
		t.Errorf("Expected verbosity 2, got %d", config.Verbosity)
	}
}

// TestDefaultAppConfigLoader_LoadFromFile_EmptyPath_Simple tests error handling for empty path
func TestDefaultAppConfigLoader_LoadFromFile_EmptyPath_Simple(t *testing.T) {
	t.Parallel()

	loader := NewAppConfigLoader()
	_, err := loader.LoadFromFile("")

	if err == nil {
		t.Fatal("Expected error when loading with empty path")
	}

	// Verify error is wrapped properly
	sentinelErr, ok := err.(*models.SentinelError)
	if !ok {
		t.Fatalf("Expected SentinelError, got %T", err)
	}

	if sentinelErr.Type != models.ErrorTypeValidation {
		t.Errorf("Expected error type %v, got %v", models.ErrorTypeValidation, sentinelErr.Type)
	}

	if !strings.Contains(sentinelErr.Message, "failed to load configuration from file") {
		t.Errorf("Expected error message to contain 'failed to load configuration from file', got: %s", sentinelErr.Message)
	}
}

// TestDefaultAppConfigLoader_LoadFromDefaults_Simple tests the LoadFromDefaults method
func TestDefaultAppConfigLoader_LoadFromDefaults_Simple(t *testing.T) {
	t.Parallel()

	loader := NewAppConfigLoader()
	config := loader.LoadFromDefaults()

	if config == nil {
		t.Fatal("LoadFromDefaults should not return nil config")
	}

	// Verify default values are set correctly
	if !config.Colors {
		t.Error("Default colors should be true")
	}

	if config.Verbosity != 0 {
		t.Errorf("Default verbosity should be 0, got %d", config.Verbosity)
	}

	if config.Watch.Enabled {
		t.Error("Default watch should be disabled")
	}
}

// TestDefaultAppConfigLoader_Merge_Simple tests the Merge method
func TestDefaultAppConfigLoader_Merge_Simple(t *testing.T) {
	t.Parallel()

	config := &AppConfig{
		Colors:    false,
		Verbosity: 0,
		Watch:     AppWatchConfig{Enabled: false},
	}

	args := &AppArguments{
		Watch:   true,
		Colors:  true,
		Verbose: true,
	}

	loader := NewAppConfigLoader()
	merged := loader.Merge(config, args)

	if merged == nil {
		t.Fatal("Merge should not return nil")
	}

	// Verify it's a copy, not the same instance
	if merged == config {
		t.Error("Merge should return a copy, not the same instance")
	}

	if !merged.Watch.Enabled {
		t.Error("Watch should be enabled after merge")
	}

	if !merged.Colors {
		t.Error("Colors should be enabled after merge")
	}

	if merged.Verbosity != 1 {
		t.Errorf("Expected verbosity 1 after merge, got %d", merged.Verbosity)
	}
}

// TestDefaultAppConfigLoader_Validate_Simple tests the Validate method
func TestDefaultAppConfigLoader_Validate_Simple(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      *AppConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_config",
			config: &AppConfig{
				Colors:    true,
				Verbosity: 2,
				Watch:     AppWatchConfig{Enabled: false},
				Visual:    AppVisualConfig{Icons: "unicode"},
			},
			expectError: false,
		},
		{
			name: "invalid_icons",
			config: &AppConfig{
				Visual: AppVisualConfig{Icons: "invalid"},
			},
			expectError: true,
			errorMsg:    "invalid icons setting",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			loader := NewAppConfigLoader()
			err := loader.Validate(tt.config)

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

// TestDefaultAppConfigLoader_ConvertCliConfigToAppConfig_Simple tests the conversion function
func TestDefaultAppConfigLoader_ConvertCliConfigToAppConfig_Simple(t *testing.T) {
	t.Parallel()

	cliConfig := &Config{
		Colors:    false,
		Verbosity: 3,
		Visual: VisualConfig{
			Icons: "ascii",
			Theme: "dark",
		},
	}

	loader := NewAppConfigLoader().(*DefaultAppConfigLoader)
	appConfig := loader.convertCliConfigToAppConfig(cliConfig)

	if appConfig == nil {
		t.Fatal("convertCliConfigToAppConfig should not return nil")
	}

	if appConfig.Colors {
		t.Error("Colors should be false")
	}
	if appConfig.Verbosity != 3 {
		t.Errorf("Expected verbosity 3, got %d", appConfig.Verbosity)
	}
	if appConfig.Visual.Icons != "ascii" {
		t.Errorf("Expected icons 'ascii', got %q", appConfig.Visual.Icons)
	}
	if appConfig.Visual.Theme != "dark" {
		t.Errorf("Expected theme 'dark', got %q", appConfig.Visual.Theme)
	}
}

// TestDefaultAppConfigLoader_Integration tests integration scenarios
func TestDefaultAppConfigLoader_Integration(t *testing.T) {
	t.Parallel()

	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "integration.json")
	configData := `{
		"colors": false,
		"verbosity": 1
	}`

	err := os.WriteFile(configFile, []byte(configData), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	loader := NewAppConfigLoader()

	// Test loading from file
	fileConfig, err := loader.LoadFromFile(configFile)
	if err != nil {
		t.Fatalf("LoadFromFile should not error: %v", err)
	}

	// Test loading from defaults
	defaultConfig := loader.LoadFromDefaults()
	if defaultConfig == nil {
		t.Fatal("LoadFromDefaults should not return nil")
	}

	// Test merging
	args := &AppArguments{
		Watch:  true,
		Colors: true,
	}
	merged := loader.Merge(fileConfig, args)
	if merged == nil {
		t.Fatal("Merge should not return nil")
	}

	// Test validation
	err = loader.Validate(merged)
	if err != nil {
		t.Fatalf("Validate should not error: %v", err)
	}

	// Verify final config has expected values
	if !merged.Colors {
		t.Error("Colors should be true from args override")
	}

	if merged.Verbosity != 1 {
		t.Errorf("Expected verbosity 1 from file config, got %d", merged.Verbosity)
	}

	if !merged.Watch.Enabled {
		t.Error("Watch should be enabled from args")
	}
}
