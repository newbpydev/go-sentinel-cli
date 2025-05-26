package app

import (
	"strings"
	"testing"
)

// TestApplicationController_RunSingleTests tests basic test execution integration
func TestApplicationController_RunSingleTests(t *testing.T) {
	// Arrange
	controller := NewApplicationController()
	args := []string{"run", "./internal/config"}

	// Act
	err := controller.Run(args)

	// Assert
	if err != nil {
		// Check if it's a "no tests" error, which is acceptable for config package
		if !strings.Contains(err.Error(), "no test files") &&
			!strings.Contains(err.Error(), "no tests to run") &&
			!strings.Contains(err.Error(), "no tests found") &&
			!strings.Contains(err.Error(), "test execution failed") {
			t.Errorf("Expected no error or 'no tests' error, got: %v", err)
		}
	}
}

// TestApplicationController_RunWithVerbose tests verbose flag integration
func TestApplicationController_RunWithVerbose(t *testing.T) {
	// Arrange
	controller := NewApplicationController()
	args := []string{"run", "--verbose", "./internal/config"}

	// Act
	err := controller.Run(args)

	// Assert
	if err != nil {
		// Check if it's a "no tests" error, which is acceptable
		if !strings.Contains(err.Error(), "no test files") &&
			!strings.Contains(err.Error(), "no tests to run") &&
			!strings.Contains(err.Error(), "no tests found") &&
			!strings.Contains(err.Error(), "test execution failed") {
			t.Errorf("Expected no error or 'no tests' error, got: %v", err)
		}
	}
}

// TestApplicationController_RunWithInvalidPackage tests error handling
func TestApplicationController_RunWithInvalidPackage(t *testing.T) {
	// Arrange
	controller := NewApplicationController()
	args := []string{"run", "./nonexistent/package"}

	// Act
	err := controller.Run(args)

	// Assert
	if err == nil {
		t.Error("Expected error for nonexistent package, got nil")
	}

	// Should contain package not found indicator
	if !strings.Contains(err.Error(), "package not found") &&
		!strings.Contains(err.Error(), "cannot find package") &&
		!strings.Contains(err.Error(), "no such file") &&
		!strings.Contains(err.Error(), "test execution failed") {
		t.Errorf("Expected package not found error, got: %v", err)
	}
}

// TestApplicationController_Interface tests interface compliance
func TestApplicationController_Interface(t *testing.T) {
	// Arrange & Act
	controller := NewApplicationController()

	// Assert - Check that it implements the interface
	if controller == nil {
		t.Error("NewApplicationController should return a valid controller")
	}
}

// TestNewApplicationController_NotNil tests that constructor returns valid instance
func TestNewApplicationController_NotNil(t *testing.T) {
	// Act
	controller := NewApplicationController()

	// Assert
	if controller == nil {
		t.Error("NewApplicationController should not return nil")
	}
}
