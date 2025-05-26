// Package config provides application configuration loading implementation
package config

import (
	"fmt"
	"time"

	"github.com/newbpydev/go-sentinel/pkg/models"
)

// DefaultAppConfigLoader implements the AppConfigLoader interface using modular config components.
// This implementation follows dependency injection principles and the Single Responsibility Principle.
type DefaultAppConfigLoader struct {
	// Config component dependencies (injected)
	cliLoader      ConfigLoader
	validationMode ValidationMode
}

// NewAppConfigLoader creates a new DefaultAppConfigLoader with default dependencies.
// This follows the Factory pattern and dependency injection principles.
func NewAppConfigLoader() AppConfigLoader {
	return &DefaultAppConfigLoader{
		cliLoader:      &DefaultConfigLoader{},
		validationMode: ValidationModeStrict,
	}
}

// NewAppConfigLoaderWithDependencies creates a new DefaultAppConfigLoader with injected dependencies.
// This constructor promotes testability and follows dependency inversion principles.
func NewAppConfigLoaderWithDependencies(deps AppConfigLoaderDependencies) AppConfigLoader {
	cliLoader := deps.CliLoader
	if cliLoader == nil {
		cliLoader = &DefaultConfigLoader{}
	}

	return &DefaultAppConfigLoader{
		cliLoader:      cliLoader,
		validationMode: deps.ValidationMode,
	}
}

// LoadFromFile loads configuration from a file
func (l *DefaultAppConfigLoader) LoadFromFile(path string) (*AppConfig, error) {
	if path == "" {
		return nil, models.WrapError(
			fmt.Errorf("config path cannot be empty"),
			models.ErrorTypeValidation,
			models.SeverityError,
			"failed to load configuration from file",
		).WithContext("config_path", path)
	}

	// Use the existing CLI config loader
	cliConfig, err := l.cliLoader.LoadFromFile(path)
	if err != nil {
		return nil, models.WrapError(
			err,
			models.ErrorTypeConfig,
			models.SeverityError,
			fmt.Sprintf("failed to load configuration from file: %s", path),
		).WithContext("config_file", path)
	}

	// Convert CLI config to app Configuration
	appConfig := l.convertCliConfigToAppConfig(cliConfig)
	return appConfig, nil
}

// LoadFromDefaults returns default configuration
func (l *DefaultAppConfigLoader) LoadFromDefaults() *AppConfig {
	// Get default CLI config
	cliConfig := GetDefaultConfig()

	// Convert to app Configuration
	return l.convertCliConfigToAppConfig(cliConfig)
}

// Merge merges CLI arguments with configuration
func (l *DefaultAppConfigLoader) Merge(config *AppConfig, args *AppArguments) *AppConfig {
	if config == nil || args == nil {
		return config
	}

	// Create a copy to avoid modifying the original
	merged := *config

	// Override config values with CLI arguments
	if args.Watch {
		merged.Watch.Enabled = true
	}

	if args.Colors {
		merged.Colors = true
	}

	if args.Verbose {
		merged.Verbosity = 1
	}

	// Set packages from CLI args
	if len(args.Packages) > 0 {
		merged.Paths.IncludePatterns = args.Packages
	}

	return &merged
}

// Validate validates the final configuration
func (l *DefaultAppConfigLoader) Validate(config *AppConfig) error {
	if config == nil {
		return models.NewValidationError("config", "configuration cannot be nil")
	}

	// Skip validation if mode is off (for testing)
	if l.validationMode == ValidationModeOff {
		return nil
	}

	// Validate watch configuration
	if config.Watch.Enabled {
		if config.Watch.Debounce != "" {
			if _, err := time.ParseDuration(config.Watch.Debounce); err != nil {
				if l.validationMode == ValidationModeStrict {
					return models.NewValidationError(
						"watch.debounce",
						fmt.Sprintf("invalid debounce duration: %s", config.Watch.Debounce),
					)
				}
				// In lenient mode, set default
				config.Watch.Debounce = "100ms"
			}
		}
	}

	// Validate test configuration
	if config.Test.Timeout != "" {
		if _, err := time.ParseDuration(config.Test.Timeout); err != nil {
			if l.validationMode == ValidationModeStrict {
				return models.NewValidationError(
					"test.timeout",
					fmt.Sprintf("invalid timeout duration: %s", config.Test.Timeout),
				)
			}
			// In lenient mode, set default
			config.Test.Timeout = "30s"
		}
	}

	// Validate visual configuration (consistent with main config loader)
	validIcons := map[string]bool{
		"unicode": true,
		"ascii":   true,
		"minimal": true,
		"none":    true,
	}
	if !validIcons[config.Visual.Icons] {
		if l.validationMode == ValidationModeStrict {
			return models.NewValidationError(
				"visual.icons",
				fmt.Sprintf("invalid icons setting: %s (valid: unicode, ascii, minimal, none)", config.Visual.Icons),
			)
		}
		// In lenient mode, set default
		config.Visual.Icons = "unicode"
	}

	return nil
}

// convertCliConfigToAppConfig converts a CLI config to an app Configuration
func (l *DefaultAppConfigLoader) convertCliConfigToAppConfig(cliConfig *Config) *AppConfig {
	return &AppConfig{
		Watch: AppWatchConfig{
			Enabled:        false, // Will be set by CLI args
			IgnorePatterns: cliConfig.Watch.IgnorePatterns,
			Debounce:       cliConfig.Watch.Debounce.String(),
			RunOnStart:     cliConfig.Watch.RunOnStart,
			ClearOnRerun:   cliConfig.Watch.ClearOnRerun,
		},
		Paths: AppPathsConfig{
			IncludePatterns: cliConfig.Paths.IncludePatterns,
			ExcludePatterns: cliConfig.Paths.ExcludePatterns,
		},
		Visual: AppVisualConfig{
			Icons:         cliConfig.Visual.Icons,
			Theme:         cliConfig.Visual.Theme,
			TerminalWidth: 80, // Default
		},
		Test: AppTestConfig{
			Timeout:  cliConfig.Timeout.String(),
			Parallel: cliConfig.Parallel,
			Coverage: false, // Default
		},
		Colors:    cliConfig.Colors,
		Verbosity: cliConfig.Verbosity,
	}
}

// Ensure DefaultAppConfigLoader implements AppConfigLoader interface
var _ AppConfigLoader = (*DefaultAppConfigLoader)(nil)
