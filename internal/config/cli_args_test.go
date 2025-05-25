package config

import (
	"testing"
)

func TestCLIArgs_ParseWatchFlag(t *testing.T) {
	parser := &DefaultArgParser{}

	// Test watch flag short form
	cliArgs, err := parser.Parse([]string{"-w"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cliArgs.Watch {
		t.Errorf("expected watch=true, got watch=false")
	}

	// Test watch flag long form
	cliArgs, err = parser.Parse([]string{"--watch"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cliArgs.Watch {
		t.Errorf("expected watch=true, got watch=false")
	}

	// Test no watch flag
	cliArgs, err = parser.Parse([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cliArgs.Watch {
		t.Errorf("expected watch=false, got watch=true")
	}
}

func TestCLIArgs_ParsePackagePatterns(t *testing.T) {
	parser := &DefaultArgParser{}

	// Test single package
	cliArgs, err := parser.Parse([]string{"./internal"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cliArgs.Packages) != 1 || cliArgs.Packages[0] != "./internal" {
		t.Errorf("expected packages=[./internal], got packages=%v", cliArgs.Packages)
	}

	// Test multiple packages
	cliArgs, err = parser.Parse([]string{"./internal", "./cmd"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cliArgs.Packages) != 2 {
		t.Errorf("expected 2 packages, got %d", len(cliArgs.Packages))
	}
}

func TestCLIArgs_ParseTestNamePattern(t *testing.T) {
	parser := &DefaultArgParser{}

	// Test pattern short form
	cliArgs, err := parser.Parse([]string{"-t", "TestFunction"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cliArgs.TestPattern != "TestFunction" {
		t.Errorf("expected test pattern='TestFunction', got '%s'", cliArgs.TestPattern)
	}

	// Test pattern long form
	cliArgs, err = parser.Parse([]string{"--test", "TestSuite"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cliArgs.TestPattern != "TestSuite" {
		t.Errorf("expected test pattern='TestSuite', got '%s'", cliArgs.TestPattern)
	}
}

func TestCLIArgs_ParseVerbosityLevel(t *testing.T) {
	parser := &DefaultArgParser{}

	// Test single verbose flag
	cliArgs, err := parser.Parse([]string{"-v"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cliArgs.Verbosity != 1 {
		t.Errorf("expected verbosity=1, got %d", cliArgs.Verbosity)
	}

	// Test multiple verbose flags
	cliArgs, err = parser.Parse([]string{"-vvv"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cliArgs.Verbosity != 3 {
		t.Errorf("expected verbosity=3, got %d", cliArgs.Verbosity)
	}

	// Test verbosity level with number
	cliArgs, err = parser.Parse([]string{"--verbosity=2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cliArgs.Verbosity != 2 {
		t.Errorf("expected verbosity=2, got %d", cliArgs.Verbosity)
	}
}

func TestCLIArgs_ParseFromCobra(t *testing.T) {
	parser := &DefaultArgParser{}

	cliArgs := parser.ParseFromCobra(true, true, false, false, true, []string{"./internal"}, "TestUnit", "aggressive")

	if !cliArgs.Watch {
		t.Errorf("expected watch=true, got watch=false")
	}
	if !cliArgs.Colors {
		t.Errorf("expected colors=true, got colors=false")
	}
	if cliArgs.Verbosity != 0 {
		t.Errorf("expected verbosity=0, got verbosity=%d", cliArgs.Verbosity)
	}
	if cliArgs.TestPattern != "TestUnit" {
		t.Errorf("expected test pattern='TestUnit', got '%s'", cliArgs.TestPattern)
	}
	if !cliArgs.Optimized {
		t.Errorf("expected optimized=true, got optimized=false")
	}
	if cliArgs.OptimizationMode != "aggressive" {
		t.Errorf("expected optimization mode='aggressive', got '%s'", cliArgs.OptimizationMode)
	}
}

func TestCLIArgs_ErrorHandling(t *testing.T) {
	parser := &DefaultArgParser{}

	// Test invalid verbosity level
	_, err := parser.Parse([]string{"--verbosity=invalid"})
	if err == nil {
		t.Errorf("expected error but got none")
	}

	// Test negative verbosity
	_, err = parser.Parse([]string{"--verbosity=-1"})
	if err == nil {
		t.Errorf("expected error but got none")
	}

	// Test verbosity too high
	_, err = parser.Parse([]string{"--verbosity=10"})
	if err == nil {
		t.Errorf("expected error but got none")
	}
}
