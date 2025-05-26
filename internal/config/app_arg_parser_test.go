package config

import (
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
