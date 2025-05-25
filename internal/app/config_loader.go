// Package app provides configuration loading implementation
package app

import (
	"fmt"
	"time"

	"github.com/newbpydev/go-sentinel/internal/config"
	"github.com/newbpydev/go-sentinel/pkg/models"
)

// DefaultConfigurationLoader implements the ConfigurationLoader interface
type DefaultConfigurationLoader struct {
	cliLoader config.ConfigLoader
}

// NewConfigurationLoader creates a new configuration loader
func NewConfigurationLoader() ConfigurationLoader {
	return &DefaultConfigurationLoader{
		cliLoader: &config.DefaultConfigLoader{},
	}
}

// LoadFromFile loads configuration from a file
func (l *DefaultConfigurationLoader) LoadFromFile(path string) (*Configuration, error) {
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
func (l *DefaultConfigurationLoader) LoadFromDefaults() *Configuration {
	// Get default CLI config
	cliConfig := config.GetDefaultConfig()

	// Convert to app Configuration
	return l.convertCliConfigToAppConfig(cliConfig)
}

// Merge merges CLI arguments with configuration
func (l *DefaultConfigurationLoader) Merge(config *Configuration, args *Arguments) *Configuration {
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
func (l *DefaultConfigurationLoader) Validate(config *Configuration) error {
	if config == nil {
		return models.NewValidationError("config", "configuration cannot be nil")
	}

	// Validate watch configuration
	if config.Watch.Enabled {
		if config.Watch.Debounce != "" {
			if _, err := time.ParseDuration(config.Watch.Debounce); err != nil {
				return models.NewValidationError(
					"watch.debounce",
					fmt.Sprintf("invalid debounce duration: %s", config.Watch.Debounce),
				)
			}
		}
	}

	// Validate test configuration
	if config.Test.Timeout != "" {
		if _, err := time.ParseDuration(config.Test.Timeout); err != nil {
			return models.NewValidationError(
				"test.timeout",
				fmt.Sprintf("invalid timeout duration: %s", config.Test.Timeout),
			)
		}
	}

	// Validate visual configuration
	validIcons := map[string]bool{
		"none":   true,
		"simple": true,
		"rich":   true,
	}
	if !validIcons[config.Visual.Icons] {
		return models.NewValidationError(
			"visual.icons",
			fmt.Sprintf("invalid icons setting: %s (valid: none, simple, rich)", config.Visual.Icons),
		)
	}

	return nil
}

// convertCliConfigToAppConfig converts a CLI config to an app Configuration
func (l *DefaultConfigurationLoader) convertCliConfigToAppConfig(cliConfig *config.Config) *Configuration {
	return &Configuration{
		Watch: WatchConfig{
			Enabled:        false, // Will be set by CLI args
			IgnorePatterns: cliConfig.Watch.IgnorePatterns,
			Debounce:       cliConfig.Watch.Debounce.String(),
			RunOnStart:     cliConfig.Watch.RunOnStart,
			ClearOnRerun:   cliConfig.Watch.ClearOnRerun,
		},
		Paths: PathsConfig{
			IncludePatterns: cliConfig.Paths.IncludePatterns,
			ExcludePatterns: cliConfig.Paths.ExcludePatterns,
		},
		Visual: VisualConfig{
			Icons:         cliConfig.Visual.Icons,
			Theme:         cliConfig.Visual.Theme,
			TerminalWidth: 80, // Default
		},
		Test: TestConfig{
			Timeout:  cliConfig.Timeout.String(),
			Parallel: cliConfig.Parallel,
			Coverage: false, // Default
		},
		Colors:    cliConfig.Colors,
		Verbosity: cliConfig.Verbosity,
	}
}

// Ensure DefaultConfigurationLoader implements ConfigurationLoader interface
var _ ConfigurationLoader = (*DefaultConfigurationLoader)(nil)
