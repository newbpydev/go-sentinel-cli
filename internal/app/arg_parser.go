// Package app provides argument parsing implementation
package app

import (
	"os"

	"github.com/newbpydev/go-sentinel/internal/config"
	"github.com/newbpydev/go-sentinel/pkg/models"
)

// DefaultArgumentParser implements the ArgumentParser interface
type DefaultArgumentParser struct {
	cliParser config.ArgParser
}

// NewArgumentParser creates a new argument parser
func NewArgumentParser() ArgumentParser {
	return &DefaultArgumentParser{
		cliParser: &config.DefaultArgParser{},
	}
}

// Parse parses command-line arguments into a structured format
func (p *DefaultArgumentParser) Parse(args []string) (*Arguments, error) {
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
	appArgs := &Arguments{
		Packages:         cliArgs.Packages,
		Watch:            cliArgs.Watch,
		Verbose:          cliArgs.Verbosity > 0,
		Colors:           cliArgs.Colors,
		Optimized:        cliArgs.Optimized,
		OptimizationMode: cliArgs.OptimizationMode,
		Writer:           os.Stdout,
	}

	return appArgs, nil
}

// Help returns help text for the application
func (p *DefaultArgumentParser) Help() string {
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

// Version returns version information
func (p *DefaultArgumentParser) Version() string {
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

// Ensure DefaultArgumentParser implements ArgumentParser interface
var _ ArgumentParser = (*DefaultArgumentParser)(nil)
