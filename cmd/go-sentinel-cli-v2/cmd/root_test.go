package cmd

import (
	"testing"
)

// TestRootCmd_Structure tests that the root command is properly configured
func TestRootCmd_Structure(t *testing.T) {
	// Test basic command properties
	if rootCmd.Use != "go-sentinel" {
		t.Errorf("Expected Use 'go-sentinel', got '%s'", rootCmd.Use)
	}

	if rootCmd.Short == "" {
		t.Error("Expected Short description to be set")
	}

	if rootCmd.Long == "" {
		t.Error("Expected Long description to be set")
	}
}

// TestRootCmd_Flags tests that persistent flags are properly configured
func TestRootCmd_Flags(t *testing.T) {
	// Test that color flag exists
	colorFlag := rootCmd.PersistentFlags().Lookup("color")
	if colorFlag == nil {
		t.Error("Expected 'color' flag to be defined")
	} else {
		if colorFlag.Shorthand != "c" {
			t.Errorf("Expected color flag shorthand 'c', got '%s'", colorFlag.Shorthand)
		}
		if colorFlag.DefValue != "true" {
			t.Errorf("Expected color flag default 'true', got '%s'", colorFlag.DefValue)
		}
	}

	// Test that watch flag exists
	watchFlag := rootCmd.PersistentFlags().Lookup("watch")
	if watchFlag == nil {
		t.Error("Expected 'watch' flag to be defined")
	} else {
		if watchFlag.Shorthand != "w" {
			t.Errorf("Expected watch flag shorthand 'w', got '%s'", watchFlag.Shorthand)
		}
		if watchFlag.DefValue != "false" {
			t.Errorf("Expected watch flag default 'false', got '%s'", watchFlag.DefValue)
		}
	}
}

// TestRootCmd_FlagValues tests setting and getting flag values
func TestRootCmd_FlagValues(t *testing.T) {
	// Test setting color flag
	err := rootCmd.PersistentFlags().Set("color", "false")
	if err != nil {
		t.Errorf("Failed to set color flag: %v", err)
	}

	colorValue, err := rootCmd.PersistentFlags().GetBool("color")
	if err != nil {
		t.Errorf("Failed to get color flag value: %v", err)
	}
	if colorValue != false {
		t.Errorf("Expected color flag value false, got %v", colorValue)
	}

	// Test setting watch flag
	err = rootCmd.PersistentFlags().Set("watch", "true")
	if err != nil {
		t.Errorf("Failed to set watch flag: %v", err)
	}

	watchValue, err := rootCmd.PersistentFlags().GetBool("watch")
	if err != nil {
		t.Errorf("Failed to get watch flag value: %v", err)
	}
	if watchValue != true {
		t.Errorf("Expected watch flag value true, got %v", watchValue)
	}

	// Reset flags to defaults for other tests
	rootCmd.PersistentFlags().Set("color", "true")
	rootCmd.PersistentFlags().Set("watch", "false")
}

// TestExecute_FunctionExists tests that Execute function exists
func TestExecute_FunctionExists(t *testing.T) {
	// We can't easily test Execute() without actually running the command
	// but we can verify the rootCmd is properly initialized which Execute depends on
	if rootCmd == nil {
		t.Error("Expected rootCmd to be initialized")
	}

	// Verify that rootCmd has the expected structure for Execute to work
	if rootCmd.Use == "" {
		t.Error("Expected rootCmd.Use to be set for Execute to work properly")
	}
}
