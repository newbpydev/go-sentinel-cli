// Package app provides adapter for config loading to maintain clean package boundaries
package app

import (
	"github.com/newbpydev/go-sentinel/internal/config"
)

// configLoaderAdapter adapts the config package loader to the app package interface.
// This adapter pattern allows us to maintain compatibility while moving to proper architecture.
type configLoaderAdapter struct {
	factory   *ConfigLoaderFactory
	appLoader config.AppConfigLoader
}

// LoadFromFile loads configuration from a file using the config package implementation.
func (a *configLoaderAdapter) LoadFromFile(path string) (*Configuration, error) {
	// Delegate to config package
	appConfig, err := a.appLoader.LoadFromFile(path)
	if err != nil {
		return nil, err
	}

	// Convert from config package types to app package types
	return a.factory.convertFromConfigConfiguration(appConfig), nil
}

// LoadFromDefaults returns default configuration using the config package implementation.
func (a *configLoaderAdapter) LoadFromDefaults() *Configuration {
	// Delegate to config package
	appConfig := a.appLoader.LoadFromDefaults()

	// Convert from config package types to app package types
	return a.factory.convertFromConfigConfiguration(appConfig)
}

// Merge merges CLI arguments with configuration using the config package implementation.
func (a *configLoaderAdapter) Merge(config *Configuration, args *Arguments) *Configuration {
	// Convert app package types to config package types
	configAppConfig := a.factory.convertToConfigConfiguration(config)
	configArgs := a.factory.convertToConfigArguments(args)

	// Delegate to config package
	mergedConfig := a.appLoader.Merge(configAppConfig, configArgs)

	// Convert back to app package types
	return a.factory.convertFromConfigConfiguration(mergedConfig)
}

// Validate validates the final configuration using the config package implementation.
func (a *configLoaderAdapter) Validate(config *Configuration) error {
	// Convert app package types to config package types
	configAppConfig := a.factory.convertToConfigConfiguration(config)

	// Delegate to config package
	return a.appLoader.Validate(configAppConfig)
}

// NewConfigurationLoader creates a new configuration loader using the adapter pattern.
// This follows dependency injection principles and maintains package boundaries.
func NewConfigurationLoader() ConfigurationLoader {
	factory := NewConfigLoaderFactory()
	return factory.CreateConfigLoaderWithDefaults()
}

// Ensure configLoaderAdapter implements ConfigurationLoader interface
var _ ConfigurationLoader = (*configLoaderAdapter)(nil)
