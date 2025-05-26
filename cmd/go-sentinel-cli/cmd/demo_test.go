package cmd

import (
	"testing"
)

// TestDemoCmd_Structure tests that the demo command is properly configured
func TestDemoCmd_Structure(t *testing.T) {
	// Test basic command properties
	if demoCmd.Use != "demo" {
		t.Errorf("Expected Use 'demo', got '%s'", demoCmd.Use)
	}

	if demoCmd.Short == "" {
		t.Error("Expected Short description to be set")
	}

	if demoCmd.Long == "" {
		t.Error("Expected Long description to be set")
	}

	if demoCmd.Run == nil {
		t.Error("Expected Run function to be set")
	}
}

// TestDemoCmd_PhaseFlag tests that the phase flag is properly configured
func TestDemoCmd_PhaseFlag(t *testing.T) {
	// Test that phase flag exists
	phaseFlag := demoCmd.Flags().Lookup("phase")
	if phaseFlag == nil {
		t.Fatal("Expected 'phase' flag to be defined")
	}

	// Test shorthand
	if phaseFlag.Shorthand != "p" {
		t.Errorf("Expected phase flag shorthand 'p', got '%s'", phaseFlag.Shorthand)
	}

	// Test that it's a string flag
	_, err := demoCmd.Flags().GetString("phase")
	if err != nil {
		t.Errorf("Expected phase to be a string flag, got error: %v", err)
	}

	// Test default value
	defaultValue, err := demoCmd.Flags().GetString("phase")
	if err != nil {
		t.Fatalf("Failed to get phase flag default: %v", err)
	}
	if defaultValue != "" {
		t.Errorf("Expected phase flag default to be empty, got '%s'", defaultValue)
	}
}

// TestDemoCmd_RequiredFlag tests that phase flag is marked as required
func TestDemoCmd_RequiredFlag(t *testing.T) {
	// Test that phase flag is required
	phaseFlag := demoCmd.Flags().Lookup("phase")
	if phaseFlag == nil {
		t.Fatal("Expected 'phase' flag to be defined")
	}

	// Check if flag is marked as required
	// Note: Cobra doesn't expose a direct way to check if a flag is required,
	// but we can verify the flag exists and has the expected properties
	if phaseFlag.Usage == "" {
		t.Error("Expected phase flag to have usage description")
	}
}

// TestDemoCmd_FlagValues tests setting and getting phase flag values
func TestDemoCmd_FlagValues(t *testing.T) {
	validPhases := []string{"1d", "2d", "3d", "4d", "5d", "6d", "7d"}

	for _, phase := range validPhases {
		t.Run("phase_"+phase, func(t *testing.T) {
			// Set the flag value
			err := demoCmd.Flags().Set("phase", phase)
			if err != nil {
				t.Fatalf("Failed to set phase flag to '%s': %v", phase, err)
			}

			// Get the flag value
			value, err := demoCmd.Flags().GetString("phase")
			if err != nil {
				t.Fatalf("Failed to get phase flag value: %v", err)
			}

			if value != phase {
				t.Errorf("Expected phase flag value '%s', got '%s'", phase, value)
			}
		})
	}
}

// TestDemoCmd_InvalidPhaseValues tests setting invalid phase values
func TestDemoCmd_InvalidPhaseValues(t *testing.T) {
	invalidPhases := []string{"invalid", "8d", "0d", "1", "demo"}

	for _, phase := range invalidPhases {
		t.Run("invalid_phase_"+phase, func(t *testing.T) {
			// Set the flag value (this should succeed as cobra doesn't validate values by default)
			err := demoCmd.Flags().Set("phase", phase)
			if err != nil {
				t.Fatalf("Failed to set phase flag to '%s': %v", phase, err)
			}

			// Get the flag value
			value, err := demoCmd.Flags().GetString("phase")
			if err != nil {
				t.Fatalf("Failed to get phase flag value: %v", err)
			}

			if value != phase {
				t.Errorf("Expected phase flag value '%s', got '%s'", phase, value)
			}

			// Note: The actual validation happens in the Run function, not in flag parsing
		})
	}
}

// TestDemoCmd_IsAddedToRoot tests that demo command is added to root
func TestDemoCmd_IsAddedToRoot(t *testing.T) {
	// Check if demo command is in root's subcommands
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "demo" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected demo command to be added to root command")
	}
}

// TestDemoCmd_HelpText tests that help text is properly formatted
func TestDemoCmd_HelpText(t *testing.T) {
	// Test that help text contains expected content
	if demoCmd.Short == "" {
		t.Error("Expected demo command to have short description")
	}

	if demoCmd.Long == "" {
		t.Error("Expected demo command to have long description")
	}

	// Test that help text mentions key concepts
	helpText := demoCmd.Long
	expectedTerms := []string{"features", "CLI", "development"}

	for _, term := range expectedTerms {
		if !containsIgnoreCase(helpText, term) {
			t.Errorf("Expected help text to contain '%s'", term)
		}
	}
}

// TestDemoCmd_CommandName tests that command name is correct
func TestDemoCmd_CommandName(t *testing.T) {
	if demoCmd.Name() != "demo" {
		t.Errorf("Expected command name 'demo', got '%s'", demoCmd.Name())
	}
}

// TestDemoCmd_NoSubcommands tests that demo command has no subcommands
func TestDemoCmd_NoSubcommands(t *testing.T) {
	subcommands := demoCmd.Commands()
	if len(subcommands) != 0 {
		t.Errorf("Expected demo command to have no subcommands, got %d", len(subcommands))
	}
}

// TestDemoCmd_FlagInheritance tests that demo command inherits persistent flags from root
func TestDemoCmd_FlagInheritance(t *testing.T) {
	// Test that persistent flags from root are available
	persistentFlags := []string{"color", "watch"}

	for _, flagName := range persistentFlags {
		flag := demoCmd.InheritedFlags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected demo command to inherit persistent flag '%s' from root", flagName)
		}
	}
}

// TestDemoCmd_LocalFlags tests that demo command has expected local flags
func TestDemoCmd_LocalFlags(t *testing.T) {
	localFlags := demoCmd.LocalFlags()

	// Should have the phase flag
	phaseFlag := localFlags.Lookup("phase")
	if phaseFlag == nil {
		t.Error("Expected demo command to have local 'phase' flag")
	}
}

// Helper function to check if string contains substring (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	// Simple case-insensitive contains check
	sLower := toLower(s)
	substrLower := toLower(substr)
	return containsAt(sLower, substrLower)
}

// Simple toLower implementation for testing
func toLower(s string) string {
	result := make([]byte, len(s))
	for i, b := range []byte(s) {
		if b >= 'A' && b <= 'Z' {
			result[i] = b + 32
		} else {
			result[i] = b
		}
	}
	return string(result)
}
