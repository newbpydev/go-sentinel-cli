package config

import (
	"bytes"
	"strings"
	"testing"

	"github.com/newbpydev/go-sentinel/pkg/models"
)

// TestNewAppArgParser_Factory tests the factory function
func TestNewAppArgParser_Factory(t *testing.T) {
	t.Parallel()

	parser := NewAppArgParser()
	if parser == nil {
		t.Fatal("NewAppArgParser should not return nil")
	}

	// Verify interface compliance
	_, ok := parser.(AppArgParser)
	if !ok {
		t.Fatal("NewAppArgParser should return AppArgParser interface")
	}

	// Verify it's the correct implementation
	defaultParser, ok := parser.(*DefaultAppArgParser)
	if !ok {
		t.Fatal("NewAppArgParser should return *DefaultAppArgParser")
	}

	// Verify default dependencies are set
	if defaultParser.cliParser == nil {
		t.Error("Default CLI parser should be set")
	}

	if defaultParser.writer == nil {
		t.Error("Default writer should be set")
	}

	if defaultParser.helpMode != HelpModeDetailed {
		t.Errorf("Expected help mode %v, got %v", HelpModeDetailed, defaultParser.helpMode)
	}
}

// TestNewAppArgParserWithDependencies_DependencyInjection tests dependency injection
func TestNewAppArgParserWithDependencies_DependencyInjection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		deps     AppArgParserDependencies
		validate func(*testing.T, AppArgParser)
	}{
		{
			name: "all_dependencies_provided",
			deps: AppArgParserDependencies{
				CliParser: &DefaultArgParser{},
				Writer:    &bytes.Buffer{},
				HelpMode:  HelpModeBrief,
			},
			validate: func(t *testing.T, parser AppArgParser) {
				defaultParser := parser.(*DefaultAppArgParser)
				if defaultParser.helpMode != HelpModeBrief {
					t.Errorf("Expected help mode %v, got %v", HelpModeBrief, defaultParser.helpMode)
				}
			},
		},
		{
			name: "nil_cli_parser_uses_default",
			deps: AppArgParserDependencies{
				CliParser: nil,
				Writer:    &bytes.Buffer{},
				HelpMode:  HelpModeUsageOnly,
			},
			validate: func(t *testing.T, parser AppArgParser) {
				defaultParser := parser.(*DefaultAppArgParser)
				if defaultParser.cliParser == nil {
					t.Error("CLI parser should not be nil when default is used")
				}
				_, ok := defaultParser.cliParser.(*DefaultArgParser)
				if !ok {
					t.Error("Should use DefaultArgParser when nil provided")
				}
			},
		},
		{
			name: "nil_writer_uses_default",
			deps: AppArgParserDependencies{
				CliParser: &DefaultArgParser{},
				Writer:    nil,
				HelpMode:  HelpModeDetailed,
			},
			validate: func(t *testing.T, parser AppArgParser) {
				defaultParser := parser.(*DefaultAppArgParser)
				if defaultParser.writer == nil {
					t.Error("Writer should not be nil when default is used")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parser := NewAppArgParserWithDependencies(tt.deps)
			if parser == nil {
				t.Fatal("NewAppArgParserWithDependencies should not return nil")
			}

			tt.validate(t, parser)
		})
	}
}

// TestDefaultAppArgParser_Parse tests the Parse method
func TestDefaultAppArgParser_Parse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		args        []string
		expectError bool
		validate    func(*testing.T, *AppArguments)
	}{
		{
			name: "empty_args",
			args: []string{},
			validate: func(t *testing.T, appArgs *AppArguments) {
				if appArgs == nil {
					t.Fatal("AppArguments should not be nil")
				}
				if appArgs.Watch {
					t.Error("Watch should be false by default")
				}
				if appArgs.Verbose {
					t.Error("Verbose should be false by default")
				}
				if !appArgs.Colors {
					t.Error("Colors should be true by default")
				}
			},
		},
		{
			name: "watch_flag",
			args: []string{"-w"},
			validate: func(t *testing.T, appArgs *AppArguments) {
				if !appArgs.Watch {
					t.Error("Watch should be true when -w flag is provided")
				}
			},
		},
		{
			name: "verbose_flag",
			args: []string{"-v"},
			validate: func(t *testing.T, appArgs *AppArguments) {
				if !appArgs.Verbose {
					t.Error("Verbose should be true when -v flag is provided")
				}
			},
		},
		{
			name: "packages_specified",
			args: []string{"./pkg", "./cmd"},
			validate: func(t *testing.T, appArgs *AppArguments) {
				if len(appArgs.Packages) != 2 {
					t.Errorf("Expected 2 packages, got %d", len(appArgs.Packages))
				}
				if appArgs.Packages[0] != "./pkg" {
					t.Errorf("Expected first package './pkg', got %q", appArgs.Packages[0])
				}
				if appArgs.Packages[1] != "./cmd" {
					t.Errorf("Expected second package './cmd', got %q", appArgs.Packages[1])
				}
			},
		},
		{
			name: "optimized_flag",
			args: []string{"--optimized"},
			validate: func(t *testing.T, appArgs *AppArguments) {
				if !appArgs.Optimized {
					t.Error("Optimized should be true when --optimized flag is provided")
				}
			},
		},
		{
			name: "optimization_mode",
			args: []string{"--optimization=aggressive"},
			validate: func(t *testing.T, appArgs *AppArguments) {
				if appArgs.OptimizationMode != "aggressive" {
					t.Errorf("Expected optimization mode 'aggressive', got %q", appArgs.OptimizationMode)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parser := NewAppArgParser()
			appArgs, err := parser.Parse(tt.args)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectError && appArgs != nil {
				tt.validate(t, appArgs)
			}
		})
	}
}

// TestDefaultAppArgParser_Parse_ErrorHandling tests error handling in Parse
func TestDefaultAppArgParser_Parse_ErrorHandling(t *testing.T) {
	t.Parallel()

	// Create a mock CLI parser that returns an error
	mockCliParser := &MockArgParser{
		ParseFunc: func(args []string) (*Args, error) {
			return nil, models.NewError(
				models.ErrorTypeValidation,
				models.SeverityError,
				"mock parse error",
			)
		},
	}

	deps := AppArgParserDependencies{
		CliParser: mockCliParser,
		Writer:    &bytes.Buffer{},
		HelpMode:  HelpModeDetailed,
	}

	parser := NewAppArgParserWithDependencies(deps)
	_, err := parser.Parse([]string{"invalid"})

	if err == nil {
		t.Fatal("Expected error when CLI parser fails")
	}

	// Verify error is wrapped properly
	sentinelErr, ok := err.(*models.SentinelError)
	if !ok {
		t.Fatalf("Expected SentinelError, got %T", err)
	}

	if sentinelErr.Type != models.ErrorTypeValidation {
		t.Errorf("Expected error type %v, got %v", models.ErrorTypeValidation, sentinelErr.Type)
	}

	if !strings.Contains(sentinelErr.Message, "failed to parse command line arguments") {
		t.Errorf("Expected error message to contain 'failed to parse command line arguments', got: %s", sentinelErr.Message)
	}
}

// TestDefaultAppArgParser_Help tests the Help method with different modes
func TestDefaultAppArgParser_Help(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		helpMode HelpMode
		validate func(*testing.T, string)
	}{
		{
			name:     "detailed_help",
			helpMode: HelpModeDetailed,
			validate: func(t *testing.T, help string) {
				if !strings.Contains(help, "go-sentinel - Beautiful Go Test Runner") {
					t.Error("Detailed help should contain title")
				}
				if !strings.Contains(help, "Usage:") {
					t.Error("Detailed help should contain usage")
				}
				if !strings.Contains(help, "Flags:") {
					t.Error("Detailed help should contain flags section")
				}
				if !strings.Contains(help, "Examples:") {
					t.Error("Detailed help should contain examples")
				}
				if !strings.Contains(help, "--watch") {
					t.Error("Detailed help should contain watch flag")
				}
			},
		},
		{
			name:     "brief_help",
			helpMode: HelpModeBrief,
			validate: func(t *testing.T, help string) {
				if !strings.Contains(help, "go-sentinel - Beautiful Go Test Runner") {
					t.Error("Brief help should contain title")
				}
				if !strings.Contains(help, "Usage:") {
					t.Error("Brief help should contain usage")
				}
				if !strings.Contains(help, "Common flags:") {
					t.Error("Brief help should contain common flags section")
				}
				if strings.Contains(help, "Examples:") {
					t.Error("Brief help should not contain examples section")
				}
			},
		},
		{
			name:     "usage_only",
			helpMode: HelpModeUsageOnly,
			validate: func(t *testing.T, help string) {
				if !strings.Contains(help, "Usage:") {
					t.Error("Usage only help should contain usage")
				}
				if strings.Contains(help, "Flags:") {
					t.Error("Usage only help should not contain flags section")
				}
				if strings.Contains(help, "Examples:") {
					t.Error("Usage only help should not contain examples section")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			deps := AppArgParserDependencies{
				CliParser: &DefaultArgParser{},
				Writer:    &bytes.Buffer{},
				HelpMode:  tt.helpMode,
			}

			parser := NewAppArgParserWithDependencies(deps)
			help := parser.Help()

			if help == "" {
				t.Fatal("Help should not be empty")
			}

			tt.validate(t, help)
		})
	}
}

// TestDefaultAppArgParser_Version tests the Version method
func TestDefaultAppArgParser_Version(t *testing.T) {
	t.Parallel()

	parser := NewAppArgParser()
	version := parser.Version()

	if version == "" {
		t.Fatal("Version should not be empty")
	}

	// Verify version contains expected information
	if !strings.Contains(version, "go-sentinel") {
		t.Error("Version should contain 'go-sentinel'")
	}

	if !strings.Contains(version, "v2.0.0") {
		t.Error("Version should contain version number")
	}

	if !strings.Contains(version, "Features:") {
		t.Error("Version should contain features section")
	}

	if !strings.Contains(version, "Copyright") {
		t.Error("Version should contain copyright")
	}

	if !strings.Contains(version, "License: MIT") {
		t.Error("Version should contain license information")
	}
}

// MockArgParser for testing error scenarios
type MockArgParser struct {
	ParseFunc          func(args []string) (*Args, error)
	ParseFromCobraFunc func(watchFlag, colorFlag, verboseFlag, failFastFlag, optimizedFlag bool, packages []string, testPattern, optimizationMode string) *Args
}

func (m *MockArgParser) Parse(args []string) (*Args, error) {
	if m.ParseFunc != nil {
		return m.ParseFunc(args)
	}
	return &Args{Colors: true}, nil
}

func (m *MockArgParser) ParseFromCobra(watchFlag, colorFlag, verboseFlag, failFastFlag, optimizedFlag bool, packages []string, testPattern, optimizationMode string) *Args {
	if m.ParseFromCobraFunc != nil {
		return m.ParseFromCobraFunc(watchFlag, colorFlag, verboseFlag, failFastFlag, optimizedFlag, packages, testPattern, optimizationMode)
	}
	return &Args{Colors: true}
}

// Ensure MockArgParser implements ArgParser interface
var _ ArgParser = (*MockArgParser)(nil)
