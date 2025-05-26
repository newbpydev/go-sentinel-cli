// Package config provides application argument parsing components
package config

import (
	"io"
)

// AppArgParser provides application-specific argument parsing functionality.
// This interface belongs in the config package as it defines config behavior contracts.
type AppArgParser interface {
	// Parse parses command-line arguments into a structured format
	Parse(args []string) (*AppArguments, error)

	// Help returns help text for the application
	Help() string

	// Version returns version information
	Version() string
}

// AppArgParserFactory creates instances of AppArgParser with proper dependencies injected.
// This follows the Factory pattern for clean dependency management.
type AppArgParserFactory interface {
	// Create creates a new AppArgParser with the specified dependencies
	Create(dependencies AppArgParserDependencies) AppArgParser

	// CreateDefault creates a new AppArgParser with default dependencies
	CreateDefault() AppArgParser
}

// AppArgParserDependencies encapsulates all dependencies needed by the AppArgParser.
// This follows dependency injection principles for better testability.
type AppArgParserDependencies struct {
	// CliParser for parsing underlying CLI arguments
	CliParser ArgParser

	// Writer for output (defaults to os.Stdout)
	Writer io.Writer

	// HelpMode for different help text styles
	HelpMode HelpMode
}

// HelpMode represents different help text styles
type HelpMode int

const (
	// HelpModeDetailed shows full help with examples
	HelpModeDetailed HelpMode = iota

	// HelpModeBrief shows concise help
	HelpModeBrief

	// HelpModeUsageOnly shows only usage line
	HelpModeUsageOnly
)
