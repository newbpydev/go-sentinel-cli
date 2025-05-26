// Package app provides factory functions for creating config components with proper dependency injection
package app

import (
	"github.com/newbpydev/go-sentinel/internal/config"
)

// ConfigLoaderFactory creates config loaders for the application.
// This factory ensures proper dependency injection and follows the Factory pattern.
type ConfigLoaderFactory struct{}

// NewConfigLoaderFactory creates a new factory for config loaders.
func NewConfigLoaderFactory() *ConfigLoaderFactory {
	return &ConfigLoaderFactory{}
}

// CreateConfigLoader creates a new config loader with proper dependencies injected.
// This method converts app types to config types to maintain package boundaries.
func (f *ConfigLoaderFactory) CreateConfigLoader() (ConfigurationLoader, error) {
	// Create the config package loader
	configLoader := config.NewAppConfigLoader()

	// Create and return the adapter
	adapter := &configLoaderAdapter{
		factory:   f,
		appLoader: configLoader,
	}

	return adapter, nil
}

// CreateConfigLoaderWithDefaults creates a config loader with default settings.
func (f *ConfigLoaderFactory) CreateConfigLoaderWithDefaults() ConfigurationLoader {
	// Create with defaults
	configLoader := config.NewAppConfigLoader()

	// Create and return the adapter
	return &configLoaderAdapter{
		factory:   f,
		appLoader: configLoader,
	}
}

// convertToConfigArguments converts app Arguments to config AppArguments.
// This conversion maintains clean package boundaries and follows dependency inversion.
func (f *ConfigLoaderFactory) convertToConfigArguments(args *Arguments) *config.AppArguments {
	if args == nil {
		return nil
	}

	return &config.AppArguments{
		Packages:         args.Packages,
		Watch:            args.Watch,
		Verbose:          args.Verbose,
		Colors:           args.Colors,
		Optimized:        args.Optimized,
		OptimizationMode: args.OptimizationMode,
	}
}

// convertFromConfigConfiguration converts config AppConfig to app Configuration.
// This conversion maintains clean package boundaries and follows dependency inversion.
func (f *ConfigLoaderFactory) convertFromConfigConfiguration(appConfig *config.AppConfig) *Configuration {
	if appConfig == nil {
		return nil
	}

	return &Configuration{
		Watch: WatchConfig{
			Enabled:        appConfig.Watch.Enabled,
			IgnorePatterns: appConfig.Watch.IgnorePatterns,
			Debounce:       appConfig.Watch.Debounce,
			RunOnStart:     appConfig.Watch.RunOnStart,
			ClearOnRerun:   appConfig.Watch.ClearOnRerun,
		},
		Paths: PathsConfig{
			IncludePatterns: appConfig.Paths.IncludePatterns,
			ExcludePatterns: appConfig.Paths.ExcludePatterns,
		},
		Visual: VisualConfig{
			Icons:         appConfig.Visual.Icons,
			Theme:         appConfig.Visual.Theme,
			TerminalWidth: appConfig.Visual.TerminalWidth,
		},
		Test: TestConfig{
			Timeout:  appConfig.Test.Timeout,
			Parallel: appConfig.Test.Parallel,
			Coverage: appConfig.Test.Coverage,
		},
		Colors:    appConfig.Colors,
		Verbosity: appConfig.Verbosity,
	}
}

// convertToConfigConfiguration converts app Configuration to config AppConfig.
// This conversion is used for updating config with app-level changes.
func (f *ConfigLoaderFactory) convertToConfigConfiguration(appConfig *Configuration) *config.AppConfig {
	if appConfig == nil {
		return nil
	}

	return &config.AppConfig{
		Watch: config.AppWatchConfig{
			Enabled:        appConfig.Watch.Enabled,
			IgnorePatterns: appConfig.Watch.IgnorePatterns,
			Debounce:       appConfig.Watch.Debounce,
			RunOnStart:     appConfig.Watch.RunOnStart,
			ClearOnRerun:   appConfig.Watch.ClearOnRerun,
		},
		Paths: config.AppPathsConfig{
			IncludePatterns: appConfig.Paths.IncludePatterns,
			ExcludePatterns: appConfig.Paths.ExcludePatterns,
		},
		Visual: config.AppVisualConfig{
			Icons:         appConfig.Visual.Icons,
			Theme:         appConfig.Visual.Theme,
			TerminalWidth: appConfig.Visual.TerminalWidth,
		},
		Test: config.AppTestConfig{
			Timeout:  appConfig.Test.Timeout,
			Parallel: appConfig.Test.Parallel,
			Coverage: appConfig.Test.Coverage,
		},
		Colors:    appConfig.Colors,
		Verbosity: appConfig.Verbosity,
	}
}

// Ensure we're following proper dependency injection patterns
var _ = (*ConfigLoaderFactory)(nil)
