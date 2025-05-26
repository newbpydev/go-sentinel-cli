// Package app provides adapter for argument parsing to maintain clean package boundaries
package app

import (
	"github.com/newbpydev/go-sentinel/internal/config"
)

// argParserAdapter adapts the config package parser to the app package interface.
// This adapter pattern allows us to maintain compatibility while moving to proper architecture.
type argParserAdapter struct {
	factory   *ArgParserFactory
	appParser config.AppArgParser
}

// Parse parses command-line arguments using the config package implementation.
func (a *argParserAdapter) Parse(args []string) (*Arguments, error) {
	// Delegate to config package
	configArgs, err := a.appParser.Parse(args)
	if err != nil {
		return nil, err
	}

	// Convert from config package types to app package types
	return a.factory.convertFromConfigArguments(configArgs), nil
}

// Help returns help text using the config package implementation.
func (a *argParserAdapter) Help() string {
	// Delegate to config package
	return a.appParser.Help()
}

// Version returns version information using the config package implementation.
func (a *argParserAdapter) Version() string {
	// Delegate to config package
	return a.appParser.Version()
}

// NewArgumentParser creates a new argument parser using the adapter pattern.
// This follows dependency injection principles and maintains package boundaries.
func NewArgumentParser() ArgumentParser {
	factory := NewArgParserFactory()
	return factory.CreateArgParserWithDefaults()
}

// Ensure argParserAdapter implements ArgumentParser interface
var _ ArgumentParser = (*argParserAdapter)(nil)
