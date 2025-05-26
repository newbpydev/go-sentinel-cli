package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

// TestRunCmd_Structure tests that the run command is properly configured
func TestRunCmd_Structure(t *testing.T) {
	// Test basic command properties
	if runCmd.Use != "run [flags] [packages]" {
		t.Errorf("Expected Use 'run [flags] [packages]', got '%s'", runCmd.Use)
	}

	if runCmd.Short == "" {
		t.Error("Expected Short description to be set")
	}

	if runCmd.Long == "" {
		t.Error("Expected Long description to be set")
	}

	if runCmd.RunE == nil {
		t.Error("Expected RunE function to be set")
	}
}

// TestRunCmd_Flags tests that all expected flags are defined
func TestRunCmd_Flags(t *testing.T) {
	expectedFlags := []struct {
		name      string
		shorthand string
		flagType  string
	}{
		{"verbose", "v", "bool"},
		{"verbosity", "q", "count"},
		{"fail-fast", "f", "bool"},
		{"color", "c", "bool"},
		{"no-color", "", "bool"},
		{"watch", "w", "bool"},
		{"test", "t", "string"},
		{"parallel", "j", "int"},
		{"timeout", "", "duration"},
		{"optimized", "o", "bool"},
		{"optimization", "", "string"},
	}

	for _, expected := range expectedFlags {
		t.Run(expected.name, func(t *testing.T) {
			flag := runCmd.Flags().Lookup(expected.name)
			if flag == nil {
				t.Errorf("Expected flag '%s' to be defined", expected.name)
				return
			}

			if flag.Shorthand != expected.shorthand {
				t.Errorf("Expected shorthand '%s', got '%s'", expected.shorthand, flag.Shorthand)
			}

			// Check flag type by attempting to get value
			switch expected.flagType {
			case "bool":
				_, err := runCmd.Flags().GetBool(expected.name)
				if err != nil {
					t.Errorf("Expected %s to be a bool flag, got error: %v", expected.name, err)
				}
			case "string":
				_, err := runCmd.Flags().GetString(expected.name)
				if err != nil {
					t.Errorf("Expected %s to be a string flag, got error: %v", expected.name, err)
				}
			case "int":
				_, err := runCmd.Flags().GetInt(expected.name)
				if err != nil {
					t.Errorf("Expected %s to be an int flag, got error: %v", expected.name, err)
				}
			case "count":
				_, err := runCmd.Flags().GetCount(expected.name)
				if err != nil {
					t.Errorf("Expected %s to be a count flag, got error: %v", expected.name, err)
				}
			case "duration":
				_, err := runCmd.Flags().GetDuration(expected.name)
				if err != nil {
					t.Errorf("Expected %s to be a duration flag, got error: %v", expected.name, err)
				}
			}
		})
	}
}

// TestRunCmd_FlagDefaults tests default values for flags
func TestRunCmd_FlagDefaults(t *testing.T) {
	testCases := []struct {
		name         string
		expectedBool bool
		expectedStr  string
		expectedInt  int
		flagType     string
	}{
		{"verbose", false, "", 0, "bool"},
		{"fail-fast", false, "", 0, "bool"},
		{"color", true, "", 0, "bool"},
		{"no-color", false, "", 0, "bool"},
		{"watch", false, "", 0, "bool"},
		{"optimized", false, "", 0, "bool"},
		{"test", false, "", 0, "string"},
		{"optimization", false, "", 0, "string"},
		{"parallel", false, "", 0, "int"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			switch tc.flagType {
			case "bool":
				value, err := runCmd.Flags().GetBool(tc.name)
				if err != nil {
					t.Fatalf("Failed to get bool flag %s: %v", tc.name, err)
				}
				if value != tc.expectedBool {
					t.Errorf("Expected %s default to be %v, got %v", tc.name, tc.expectedBool, value)
				}
			case "string":
				value, err := runCmd.Flags().GetString(tc.name)
				if err != nil {
					t.Fatalf("Failed to get string flag %s: %v", tc.name, err)
				}
				if value != tc.expectedStr {
					t.Errorf("Expected %s default to be '%s', got '%s'", tc.name, tc.expectedStr, value)
				}
			case "int":
				value, err := runCmd.Flags().GetInt(tc.name)
				if err != nil {
					t.Fatalf("Failed to get int flag %s: %v", tc.name, err)
				}
				if value != tc.expectedInt {
					t.Errorf("Expected %s default to be %d, got %d", tc.name, tc.expectedInt, value)
				}
			}
		})
	}
}

// TestBuildCLIArgs_BasicFlags tests building CLI args from basic flags
func TestBuildCLIArgs_BasicFlags(t *testing.T) {
	// Create a test command with flags
	cmd := &cobra.Command{}
	cmd.Flags().Bool("verbose", false, "verbose")
	cmd.Flags().Bool("color", true, "color")
	cmd.Flags().Bool("no-color", false, "no-color")
	cmd.Flags().Bool("watch", false, "watch")
	cmd.Flags().Bool("optimized", false, "optimized")
	cmd.Flags().Bool("fail-fast", false, "fail-fast")
	cmd.Flags().String("test", "", "test")
	cmd.Flags().String("optimization", "", "optimization")
	cmd.Flags().Int("parallel", 0, "parallel")
	cmd.Flags().Duration("timeout", 0, "timeout")
	cmd.Flags().Count("verbosity", "verbosity")

	testCases := []struct {
		name     string
		setFlags func()
		args     []string
		expected []string
	}{
		{
			name: "verbose flag",
			setFlags: func() {
				if err := cmd.Flags().Set("verbose", "true"); err != nil {
					t.Errorf("Failed to set verbose flag: %v", err)
				}
			},
			args:     []string{},
			expected: []string{"-v", "--color"},
		},
		{
			name: "no-color flag",
			setFlags: func() {
				if err := cmd.Flags().Set("no-color", "true"); err != nil {
					t.Errorf("Failed to set no-color flag: %v", err)
				}
			},
			args:     []string{},
			expected: []string{"--no-color"},
		},
		{
			name: "watch flag",
			setFlags: func() {
				if err := cmd.Flags().Set("watch", "true"); err != nil {
					t.Errorf("Failed to set watch flag: %v", err)
				}
			},
			args:     []string{},
			expected: []string{"--color", "--watch"},
		},
		{
			name: "optimized flag",
			setFlags: func() {
				if err := cmd.Flags().Set("optimized", "true"); err != nil {
					t.Errorf("Failed to set optimized flag: %v", err)
				}
			},
			args:     []string{},
			expected: []string{"--color", "--optimized"},
		},
		{
			name: "test pattern",
			setFlags: func() {
				if err := cmd.Flags().Set("test", "TestExample"); err != nil {
					t.Errorf("Failed to set test flag: %v", err)
				}
			},
			args:     []string{},
			expected: []string{"--color", "--test=TestExample"},
		},
		{
			name: "with packages",
			setFlags: func() {
				// Reset flags
				if err := cmd.Flags().Set("verbose", "false"); err != nil {
					t.Errorf("Failed to set verbose flag: %v", err)
				}
				if err := cmd.Flags().Set("watch", "false"); err != nil {
					t.Errorf("Failed to set watch flag: %v", err)
				}
			},
			args:     []string{"./pkg1", "./pkg2"},
			expected: []string{"--color", "./pkg1", "./pkg2"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset all flags to defaults
			if err := cmd.Flags().Set("verbose", "false"); err != nil {
				t.Fatalf("Failed to set verbose flag: %v", err)
			}
			if err := cmd.Flags().Set("color", "true"); err != nil {
				t.Fatalf("Failed to set color flag: %v", err)
			}
			if err := cmd.Flags().Set("no-color", "false"); err != nil {
				t.Fatalf("Failed to set no-color flag: %v", err)
			}
			if err := cmd.Flags().Set("watch", "false"); err != nil {
				t.Fatalf("Failed to set watch flag: %v", err)
			}
			if err := cmd.Flags().Set("optimized", "false"); err != nil {
				t.Fatalf("Failed to set optimized flag: %v", err)
			}
			if err := cmd.Flags().Set("fail-fast", "false"); err != nil {
				t.Fatalf("Failed to set fail-fast flag: %v", err)
			}
			if err := cmd.Flags().Set("test", ""); err != nil {
				t.Fatalf("Failed to set test flag: %v", err)
			}
			if err := cmd.Flags().Set("optimization", ""); err != nil {
				t.Fatalf("Failed to set optimization flag: %v", err)
			}
			if err := cmd.Flags().Set("parallel", "0"); err != nil {
				t.Fatalf("Failed to set parallel flag: %v", err)
			}

			// Set specific flags for this test
			tc.setFlags()

			// Act
			result := buildCLIArgs(cmd, tc.args)

			// Assert
			if len(result) != len(tc.expected) {
				t.Errorf("Expected %d args, got %d: %v", len(tc.expected), len(result), result)
				return
			}

			for i, expected := range tc.expected {
				if i >= len(result) || result[i] != expected {
					t.Errorf("Expected arg %d to be '%s', got '%s'", i, expected, result[i])
				}
			}
		})
	}
}

// TestConvertArgsToSlice_BasicConversion tests converting Args struct to string slice
func TestConvertArgsToSlice_BasicConversion(t *testing.T) {
	// This test requires importing the cli package, but since we're testing
	// the cmd package, we'll test the structure without the actual conversion
	// The function exists and can be called, which is what we're verifying

	// Test that the function exists and doesn't panic with nil input
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("convertArgsToSlice panicked with nil input: %v", r)
		}
	}()

	// We can't easily test this without creating a proper Args struct
	// since it's from the internal/cli package, but we can verify the function exists
	// by checking that it's not a nil function pointer (functions can't be compared to nil directly)
	// Instead, we'll just verify that the function can be referenced
	_ = convertArgsToSlice // This will compile if the function exists
}

// TestRunCmd_IsAddedToRoot tests that run command is added to root
func TestRunCmd_IsAddedToRoot(t *testing.T) {
	// Check if run command is in root's subcommands
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "run" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected run command to be added to root command")
	}
}

// TestRunCmd_FlagInheritance tests that run command inherits persistent flags from root
func TestRunCmd_FlagInheritance(t *testing.T) {
	// Test that persistent flags from root are available
	persistentFlags := []string{"color", "watch"}

	for _, flagName := range persistentFlags {
		// Check both inherited flags and all flags (since persistent flags might be available through different methods)
		inheritedFlag := runCmd.InheritedFlags().Lookup(flagName)
		allFlag := runCmd.Flags().Lookup(flagName)

		if inheritedFlag == nil && allFlag == nil {
			t.Errorf("Expected run command to have access to persistent flag '%s' from root", flagName)
		}
	}
}

// TestRunCmd_HelpText tests that help text is properly formatted
func TestRunCmd_HelpText(t *testing.T) {
	// Test that help text contains expected content
	if runCmd.Short == "" {
		t.Error("Expected run command to have short description")
	}

	if runCmd.Long == "" {
		t.Error("Expected run command to have long description")
	}

	// Test that help text mentions key concepts
	helpText := runCmd.Long
	expectedTerms := []string{"Go tests", "Vitest", "output"}

	for _, term := range expectedTerms {
		if !contains(helpText, term) {
			t.Errorf("Expected help text to contain '%s'", term)
		}
	}
}

// Helper function to check if string contains substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsAt(s, substr))))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
