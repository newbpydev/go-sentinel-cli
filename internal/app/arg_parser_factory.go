// Package app provides factory functions for creating argument parser components with proper dependency injection
package app

import (
	"io"

	"github.com/newbpydev/go-sentinel/internal/config"
)

// ArgParserFactory creates argument parsers for the application.
// This factory ensures proper dependency injection and follows the Factory pattern.
type ArgParserFactory struct{}

// NewArgParserFactory creates a new factory for argument parsers.
func NewArgParserFactory() *ArgParserFactory {
	return &ArgParserFactory{}
}

// CreateArgParser creates a new argument parser with proper dependencies injected.
// This method converts app types to config types to maintain package boundaries.
func (f *ArgParserFactory) CreateArgParser() (ArgumentParser, error) {
	// Create the config package parser
	configParser := config.NewAppArgParser()

	// Create and return the adapter
	adapter := &argParserAdapter{
		factory:   f,
		appParser: configParser,
	}

	return adapter, nil
}

// CreateArgParserWithDefaults creates an argument parser with default settings.
func (f *ArgParserFactory) CreateArgParserWithDefaults() ArgumentParser {
	// Create with defaults
	configParser := config.NewAppArgParser()

	// Create and return the adapter
	return &argParserAdapter{
		factory:   f,
		appParser: configParser,
	}
}

// CreateArgParserWithDependencies creates an argument parser with specific dependencies.
func (f *ArgParserFactory) CreateArgParserWithDependencies(writer io.Writer, helpMode config.HelpMode) ArgumentParser {
	// Create with dependencies
	deps := config.AppArgParserDependencies{
		CliParser: &config.DefaultArgParser{},
		Writer:    writer,
		HelpMode:  helpMode,
	}
	configParser := config.NewAppArgParserWithDependencies(deps)

	// Create and return the adapter
	return &argParserAdapter{
		factory:   f,
		appParser: configParser,
	}
}

// convertFromConfigArguments converts config AppArguments to app Arguments.
// This conversion maintains clean package boundaries and follows dependency inversion.
func (f *ArgParserFactory) convertFromConfigArguments(configArgs *config.AppArguments) *Arguments {
	if configArgs == nil {
		return nil
	}

	return &Arguments{
		Packages:         configArgs.Packages,
		Watch:            configArgs.Watch,
		Verbose:          configArgs.Verbose,
		Colors:           configArgs.Colors,
		Optimized:        configArgs.Optimized,
		OptimizationMode: configArgs.OptimizationMode,
		Writer:           nil, // Set by caller if needed
	}
}

// convertToConfigArguments converts app Arguments to config AppArguments.
// This conversion is used for updating config with app-level changes.
func (f *ArgParserFactory) convertToConfigArguments(appArgs *Arguments) *config.AppArguments {
	if appArgs == nil {
		return nil
	}

	return &config.AppArguments{
		Packages:         appArgs.Packages,
		Watch:            appArgs.Watch,
		Verbose:          appArgs.Verbose,
		Colors:           appArgs.Colors,
		Optimized:        appArgs.Optimized,
		OptimizationMode: appArgs.OptimizationMode,
	}
}

// Ensure we're following proper dependency injection patterns
var _ = (*ArgParserFactory)(nil)
