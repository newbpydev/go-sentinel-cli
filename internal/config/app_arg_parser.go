// Package config provides application argument parsing implementation
package config

import (
	"io"
	"os"

	"github.com/newbpydev/go-sentinel/pkg/models"
)

// DefaultAppArgParser implements the AppArgParser interface using modular config components.
// This implementation follows dependency injection principles and the Single Responsibility Principle.
type DefaultAppArgParser struct {
	// Config component dependencies (injected)
	cliParser ArgParser
	writer    io.Writer
	helpMode  HelpMode
}

// NewAppArgParser creates a new DefaultAppArgParser with default dependencies.
// This follows the Factory pattern and dependency injection principles.
func NewAppArgParser() AppArgParser {
	return &DefaultAppArgParser{
		cliParser: &DefaultArgParser{},
		writer:    os.Stdout,
		helpMode:  HelpModeDetailed,
	}
}

// NewAppArgParserWithDependencies creates a new DefaultAppArgParser with injected dependencies.
// This constructor promotes testability and follows dependency inversion principles.
func NewAppArgParserWithDependencies(deps AppArgParserDependencies) AppArgParser {
	cliParser := deps.CliParser
	if cliParser == nil {
		cliParser = &DefaultArgParser{}
	}

	writer := deps.Writer
	if writer == nil {
		writer = os.Stdout
	}

	return &DefaultAppArgParser{
		cliParser: cliParser,
		writer:    writer,
		helpMode:  deps.HelpMode,
	}
}

// Parse parses command-line arguments into a structured format
func (p *DefaultAppArgParser) Parse(args []string) (*AppArguments, error) {
	// Use the existing CLI argument parser
	cliArgs, err := p.cliParser.Parse(args)
	if err != nil {
		return nil, models.WrapError(
			err,
			models.ErrorTypeValidation,
			models.SeverityWarning,
			"failed to parse command line arguments",
		).WithContext("operation", "argument_parsing")
	}

	// Convert CLI args to app Arguments
	appArgs := &AppArguments{
		Packages:         cliArgs.Packages,
		Watch:            cliArgs.Watch,
		Verbose:          cliArgs.Verbosity > 0,
		Colors:           cliArgs.Colors,
		Optimized:        cliArgs.Optimized,
		OptimizationMode: cliArgs.OptimizationMode,
	}

	return appArgs, nil
}

// Help returns help text for the application
func (p *DefaultAppArgParser) Help() string {
	switch p.helpMode {
	case HelpModeBrief:
		return p.getBriefHelp()
	case HelpModeUsageOnly:
		return p.getUsageOnly()
	default:
		return p.getDetailedHelp()
	}
}

// getDetailedHelp returns full help text with examples
func (p *DefaultAppArgParser) getDetailedHelp() string {
	return `go-sentinel - Beautiful Go Test Runner

Usage:
  go-sentinel [flags] [packages]

Flags:
  -c, --color          Use colored output (default: auto-detected)
  -v, --verbose        Enable verbose output
  -w, --watch          Enable watch mode
  -j, --parallel=N     Run tests in parallel (default: number of CPUs)
  -t, --test=PATTERN   Run only tests matching pattern
  -f, --fail-fast      Stop on first failure
  --optimized          Enable optimized test execution with Go's built-in caching
  --optimization=MODE  Set optimization mode (conservative, balanced, aggressive)
  --config=FILE        Path to config file (default: sentinel.config.json)
  --timeout=DURATION   Timeout for test execution
  --coverage=MODE      Coverage mode
  --help               Show this help message
  --version            Show version information

Examples:
  go-sentinel                    # Run tests in current directory
  go-sentinel ./...              # Run all tests recursively
  go-sentinel --watch ./pkg      # Watch ./pkg directory for changes
  go-sentinel -v --color ./cmd   # Run tests with verbose, colored output
  go-sentinel --optimized        # Run with optimization enabled

For more information, visit: https://github.com/newbpydev/go-sentinel
`
}

// getBriefHelp returns concise help text
func (p *DefaultAppArgParser) getBriefHelp() string {
	return `go-sentinel - Beautiful Go Test Runner

Usage: go-sentinel [flags] [packages]

Common flags:
  -c, --color     Use colored output
  -v, --verbose   Enable verbose output
  -w, --watch     Enable watch mode
  --help          Show full help

Run 'go-sentinel --help' for detailed options.
`
}

// getUsageOnly returns only usage line
func (p *DefaultAppArgParser) getUsageOnly() string {
	return "Usage: go-sentinel [flags] [packages]"
}

// Version returns version information
func (p *DefaultAppArgParser) Version() string {
	return `go-sentinel v2.0.0
A beautiful, Vitest-style test runner for Go projects
Built with ❤️  for the Go community

Features:
  ✨ Beautiful, intuitive output
  🚀 Optimized test execution
  👀 Smart watch mode
  🎨 Customizable themes
  ⚡ Fast incremental updates

Copyright (c) 2025 NewBpyDev
License: MIT
`
}

// Ensure DefaultAppArgParser implements AppArgParser interface
var _ AppArgParser = (*DefaultAppArgParser)(nil)
