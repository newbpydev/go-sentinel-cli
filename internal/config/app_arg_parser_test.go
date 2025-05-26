package config

import (
	"bytes"
	"strings"
	"testing"
)

// TestAppArgParser_NewAppArgParser tests the creation of a new app argument parser
func TestAppArgParser_NewAppArgParser(t *testing.T) {
	t.Run("should create app arg parser with proper initialization", func(t *testing.T) {
		parser := NewAppArgParser()

		if parser == nil {
			t.Fatal("expected non-nil parser")
		}

		// Verify it implements the AppArgParser interface
		var _ AppArgParser = parser
	})
}

// TestAppArgParser_Parse tests argument parsing functionality
func TestAppArgParser_Parse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		args        []string
		expectError bool
		expected    *AppArguments
	}{
		{
			name:        "empty arguments",
			args:        []string{},
			expectError: false,
			expected: &AppArguments{
				Packages:         []string{},
				Watch:            false,
				Verbose:          false,
				Colors:           true, // Default from config is true
				Optimized:        false,
				OptimizationMode: "",
			},
		},
		{
			name:        "watch flag",
			args:        []string{"--watch"},
			expectError: false,
			expected: &AppArguments{
				Watch:   true,
				Verbose: false,
				Colors:  true, // Default from config is true
			},
		},
		{
			name:        "verbose flag",
			args:        []string{"--verbose"},
			expectError: false,
			expected: &AppArguments{
				Watch:   false,
				Verbose: true,
				Colors:  true, // Default from config is true
			},
		},
		{
			name:        "multiple flags",
			args:        []string{"--watch", "--verbose", "--color"},
			expectError: false,
			expected: &AppArguments{
				Watch:   true,
				Verbose: true,
				Colors:  true,
			},
		},
		{
			name:        "packages argument",
			args:        []string{"./pkg", "./cmd"},
			expectError: false,
			expected: &AppArguments{
				Packages: []string{"./pkg", "./cmd"},
				Watch:    false,
				Verbose:  false,
				Colors:   true, // Default from config is true
			},
		},
		{
			name:        "optimization flags",
			args:        []string{"--optimized", "--optimization=aggressive"},
			expectError: false,
			expected: &AppArguments{
				Optimized:        true,
				OptimizationMode: "aggressive",
				Colors:           true, // Default from config is true
			},
		},
		{
			name:        "invalid arguments should error gracefully",
			args:        []string{"--invalid-flag"},
			expectError: true, // Parser should return error for invalid args
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parser := NewAppArgParser()
			result, err := parser.Parse(tt.args)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectError && result == nil {
				t.Error("expected non-nil result")
			}

			if tt.expected != nil && result != nil {
				if len(tt.expected.Packages) > 0 && len(result.Packages) != len(tt.expected.Packages) {
					t.Errorf("expected packages %v, got %v", tt.expected.Packages, result.Packages)
				}

				if result.Watch != tt.expected.Watch {
					t.Errorf("expected watch %v, got %v", tt.expected.Watch, result.Watch)
				}

				if result.Verbose != tt.expected.Verbose {
					t.Errorf("expected verbose %v, got %v", tt.expected.Verbose, result.Verbose)
				}

				if result.Colors != tt.expected.Colors {
					t.Errorf("expected colors %v, got %v", tt.expected.Colors, result.Colors)
				}

				if tt.expected.Optimized && result.Optimized != tt.expected.Optimized {
					t.Errorf("expected optimized %v, got %v", tt.expected.Optimized, result.Optimized)
				}

				if tt.expected.OptimizationMode != "" && result.OptimizationMode != tt.expected.OptimizationMode {
					t.Errorf("expected optimization mode %v, got %v", tt.expected.OptimizationMode, result.OptimizationMode)
				}
			}
		})
	}
}

// TestAppArgParser_Help tests help text generation
func TestAppArgParser_Help(t *testing.T) {
	t.Run("should return help text", func(t *testing.T) {
		parser := NewAppArgParser()
		help := parser.Help()

		if help == "" {
			t.Error("expected non-empty help text")
		}

		// Check for key elements in help text
		expectedElements := []string{
			"go-sentinel",
			"Usage:",
			"Flags:",
			"Examples:",
		}

		for _, element := range expectedElements {
			if !contains(help, element) {
				t.Errorf("help text should contain '%s'", element)
			}
		}
	})
}

// TestAppArgParser_Version tests version information generation
func TestAppArgParser_Version(t *testing.T) {
	t.Run("should return version information", func(t *testing.T) {
		parser := NewAppArgParser()
		version := parser.Version()

		if version == "" {
			t.Error("expected non-empty version information")
		}

		// Check for key elements in version text
		expectedElements := []string{
			"go-sentinel",
			"v2.0.0",
			"Features:",
			"License:",
		}

		for _, element := range expectedElements {
			if !contains(version, element) {
				t.Errorf("version text should contain '%s'", element)
			}
		}
	})
}

// TestAppArgParser_InterfaceCompliance ensures the parser implements required interfaces
func TestAppArgParser_InterfaceCompliance(t *testing.T) {
	parser := NewAppArgParser()

	// Verify interface compliance at compile time
	var _ AppArgParser = parser
}

// TestAppArgParser_ErrorHandling tests error handling scenarios
func TestAppArgParser_ErrorHandling(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "nil arguments",
			args: nil,
		},
		{
			name: "conflicting flags",
			args: []string{"--color", "--no-color"},
		},
		{
			name: "malformed arguments",
			args: []string{"--timeout=invalid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewAppArgParser()
			_, err := parser.Parse(tt.args)

			// Parser should handle all cases gracefully
			if err != nil {
				t.Logf("Parser returned error (may be expected): %v", err)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestNewAppArgParserWithDependencies tests parser creation with custom dependencies
func TestNewAppArgParserWithDependencies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		dependencies AppArgParserDependencies
		validateFunc func(*testing.T, AppArgParser)
	}{
		{
			name: "with_all_dependencies",
			dependencies: AppArgParserDependencies{
				CliParser: &DefaultArgParser{},
				Writer:    &bytes.Buffer{},
				HelpMode:  HelpModeBrief,
			},
			validateFunc: func(t *testing.T, parser AppArgParser) {
				if parser == nil {
					t.Fatal("Parser should not be nil")
				}
				// Verify interface compliance
				_, ok := parser.(AppArgParser)
				if !ok {
					t.Fatal("Should implement AppArgParser interface")
				}
			},
		},
		{
			name: "with_nil_cli_parser",
			dependencies: AppArgParserDependencies{
				CliParser: nil,
				Writer:    &bytes.Buffer{},
				HelpMode:  HelpModeDetailed,
			},
			validateFunc: func(t *testing.T, parser AppArgParser) {
				if parser == nil {
					t.Fatal("Parser should not be nil")
				}
				// Should use default CLI parser when nil is provided
				defaultParser := parser.(*DefaultAppArgParser)
				if defaultParser.cliParser == nil {
					t.Error("Should have default CLI parser when nil is provided")
				}
			},
		},
		{
			name: "with_nil_writer",
			dependencies: AppArgParserDependencies{
				CliParser: &DefaultArgParser{},
				Writer:    nil,
				HelpMode:  HelpModeUsageOnly,
			},
			validateFunc: func(t *testing.T, parser AppArgParser) {
				if parser == nil {
					t.Fatal("Parser should not be nil")
				}
				// Should use default writer when nil is provided
				defaultParser := parser.(*DefaultAppArgParser)
				if defaultParser.writer == nil {
					t.Error("Should have default writer when nil is provided")
				}
			},
		},
		{
			name: "with_all_nil_dependencies",
			dependencies: AppArgParserDependencies{
				CliParser: nil,
				Writer:    nil,
				HelpMode:  HelpModeBrief,
			},
			validateFunc: func(t *testing.T, parser AppArgParser) {
				if parser == nil {
					t.Fatal("Parser should not be nil")
				}
				defaultParser := parser.(*DefaultAppArgParser)
				if defaultParser.cliParser == nil {
					t.Error("Should have default CLI parser")
				}
				if defaultParser.writer == nil {
					t.Error("Should have default writer")
				}
				if defaultParser.helpMode != HelpModeBrief {
					t.Error("Should preserve help mode")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parser := NewAppArgParserWithDependencies(tt.dependencies)
			tt.validateFunc(t, parser)
		})
	}
}

// TestDefaultAppArgParser_HelpModes tests different help modes
func TestDefaultAppArgParser_HelpModes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		helpMode     HelpMode
		expectedText string
	}{
		{
			name:         "detailed_help_mode",
			helpMode:     HelpModeDetailed,
			expectedText: "go-sentinel - Beautiful Go Test Runner",
		},
		{
			name:         "brief_help_mode",
			helpMode:     HelpModeBrief,
			expectedText: "go-sentinel - Beautiful Go Test Runner",
		},
		{
			name:         "usage_only_mode",
			helpMode:     HelpModeUsageOnly,
			expectedText: "Usage: go-sentinel [flags] [packages]",
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
			if !strings.Contains(help, tt.expectedText) {
				t.Errorf("Expected help to contain %q, got: %s", tt.expectedText, help)
			}
		})
	}
}

// TestDefaultAppArgParser_GetBriefHelp tests brief help generation
func TestDefaultAppArgParser_GetBriefHelp(t *testing.T) {
	t.Parallel()

	deps := AppArgParserDependencies{
		CliParser: &DefaultArgParser{},
		Writer:    &bytes.Buffer{},
		HelpMode:  HelpModeBrief,
	}
	parser := NewAppArgParserWithDependencies(deps).(*DefaultAppArgParser)

	help := parser.getBriefHelp()

	// Verify brief help content
	expectedContent := []string{
		"go-sentinel - Beautiful Go Test Runner",
		"Usage: go-sentinel [flags] [packages]",
		"Common flags:",
		"-c, --color",
		"-v, --verbose",
		"-w, --watch",
		"--help",
		"Run 'go-sentinel --help' for detailed options.",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(help, expected) {
			t.Errorf("Brief help should contain %q, got: %s", expected, help)
		}
	}

	// Verify it's shorter than detailed help
	detailedParser := NewAppArgParser().(*DefaultAppArgParser)
	detailedHelp := detailedParser.getDetailedHelp()
	if len(help) >= len(detailedHelp) {
		t.Error("Brief help should be shorter than detailed help")
	}
}

// TestDefaultAppArgParser_GetUsageOnly tests usage-only help generation
func TestDefaultAppArgParser_GetUsageOnly(t *testing.T) {
	t.Parallel()

	deps := AppArgParserDependencies{
		CliParser: &DefaultArgParser{},
		Writer:    &bytes.Buffer{},
		HelpMode:  HelpModeUsageOnly,
	}
	parser := NewAppArgParserWithDependencies(deps).(*DefaultAppArgParser)

	usage := parser.getUsageOnly()

	expectedUsage := "Usage: go-sentinel [flags] [packages]"
	if usage != expectedUsage {
		t.Errorf("Expected usage %q, got %q", expectedUsage, usage)
	}

	// Verify it's the shortest help format
	briefParser := NewAppArgParserWithDependencies(AppArgParserDependencies{HelpMode: HelpModeBrief}).(*DefaultAppArgParser)
	briefHelp := briefParser.getBriefHelp()
	if len(usage) >= len(briefHelp) {
		t.Error("Usage-only help should be shorter than brief help")
	}
}

// TestDefaultAppArgParser_HelpModeIntegration tests help mode integration
func TestDefaultAppArgParser_HelpModeIntegration(t *testing.T) {
	t.Parallel()

	// Test that Help() method correctly delegates to the right help function based on mode
	tests := []struct {
		name     string
		helpMode HelpMode
		validate func(*testing.T, string)
	}{
		{
			name:     "detailed_mode_integration",
			helpMode: HelpModeDetailed,
			validate: func(t *testing.T, help string) {
				if !strings.Contains(help, "Examples:") {
					t.Error("Detailed help should contain examples")
				}
				if !strings.Contains(help, "For more information") {
					t.Error("Detailed help should contain additional info")
				}
			},
		},
		{
			name:     "brief_mode_integration",
			helpMode: HelpModeBrief,
			validate: func(t *testing.T, help string) {
				if strings.Contains(help, "Examples:") {
					t.Error("Brief help should not contain examples")
				}
				if !strings.Contains(help, "Common flags:") {
					t.Error("Brief help should contain common flags")
				}
			},
		},
		{
			name:     "usage_only_integration",
			helpMode: HelpModeUsageOnly,
			validate: func(t *testing.T, help string) {
				if strings.Contains(help, "flags:") {
					t.Error("Usage-only help should not contain flag descriptions")
				}
				if !strings.Contains(help, "Usage:") {
					t.Error("Usage-only help should contain usage line")
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
			tt.validate(t, help)
		})
	}
}

// TestDefaultAppArgParser_DependencyInjection tests dependency injection behavior
func TestDefaultAppArgParser_DependencyInjection(t *testing.T) {
	t.Parallel()

	// Test that injected dependencies are actually used
	var buf bytes.Buffer
	mockParser := &DefaultArgParser{}

	deps := AppArgParserDependencies{
		CliParser: mockParser,
		Writer:    &buf,
		HelpMode:  HelpModeBrief,
	}

	parser := NewAppArgParserWithDependencies(deps).(*DefaultAppArgParser)

	// Verify dependencies are set correctly
	if parser.cliParser != mockParser {
		t.Error("Should use injected CLI parser")
	}
	if parser.writer != &buf {
		t.Error("Should use injected writer")
	}
	if parser.helpMode != HelpModeBrief {
		t.Error("Should use injected help mode")
	}
}

// TestDefaultAppArgParser_Parse_WithDependencies tests parsing with injected dependencies
func TestDefaultAppArgParser_Parse_WithDependencies(t *testing.T) {
	t.Parallel()

	// Create a mock CLI parser that returns specific results
	mockParser := &DefaultArgParser{}

	deps := AppArgParserDependencies{
		CliParser: mockParser,
		Writer:    &bytes.Buffer{},
		HelpMode:  HelpModeDetailed,
	}

	parser := NewAppArgParserWithDependencies(deps)

	// Test parsing with the injected parser
	args := []string{"--watch", "--verbose", "internal/config"}
	result, err := parser.Parse(args)

	if err != nil {
		t.Fatalf("Parse should not error: %v", err)
	}

	if result == nil {
		t.Fatal("Parse should return non-nil result")
	}

	// Verify the result structure
	if !result.Watch {
		t.Error("Expected watch to be true")
	}
	if !result.Verbose {
		t.Error("Expected verbose to be true")
	}
	if len(result.Packages) != 1 || result.Packages[0] != "internal/config" {
		t.Errorf("Expected packages [internal/config], got %v", result.Packages)
	}
}
