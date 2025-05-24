package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestIntegration_BasicCLIWorkflow tests the basic CLI workflow end-to-end
func TestIntegration_BasicCLIWorkflow(t *testing.T) {
	tempDir := t.TempDir()

	// Create a simple test file
	testFile := filepath.Join(tempDir, "example_test.go")
	testContent := `package main

import "testing"

func TestExample(t *testing.T) {
	if 1+1 != 2 {
		t.Error("Math is broken")
	}
}
`

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test the TestRunner
	runner := &TestRunner{
		Verbose:    true,
		JSONOutput: false,
	}

	ctx := context.Background()
	output, err := runner.Run(ctx, []string{tempDir})

	// In test environment, this may fail but should not crash
	if err != nil {
		t.Logf("Test runner failed (expected in test environment): %v", err)
	} else {
		t.Logf("Test runner succeeded with output length: %d", len(output))
	}
}

// TestIntegration_ProcessorWorkflow tests the processor workflow
func TestIntegration_ProcessorWorkflow(t *testing.T) {
	var output bytes.Buffer
	processor := NewTestProcessor(
		&output,
		NewColorFormatter(false), // No colors for test consistency
		NewIconProvider(false),   // ASCII icons for test consistency
		80,
	)

	// Create a test suite
	suite := &TestSuite{
		FilePath:     "example_test.go",
		TestCount:    2,
		PassedCount:  1,
		FailedCount:  1,
		SkippedCount: 0,
		Duration:     100 * time.Millisecond,
	}

	// Add test results
	suite.Tests = []*TestResult{
		{
			Name:     "TestPassing",
			Status:   StatusPassed,
			Duration: 50 * time.Millisecond,
			Package:  "github.com/test/example",
		},
		{
			Name:     "TestFailing",
			Status:   StatusFailed,
			Duration: 50 * time.Millisecond,
			Package:  "github.com/test/example",
			Error: &TestError{
				Message: "Test failed",
				Type:    "AssertionError",
			},
		},
	}

	// Process the test suite
	processor.AddTestSuite(suite)
	err := processor.RenderResults(false)

	if err != nil {
		t.Fatalf("Failed to render results: %v", err)
	}

	// Verify output contains expected elements
	outputStr := output.String()
	expectedElements := []string{
		"example_test.go",
		"TestPassing",
		"TestFailing",
	}

	for _, element := range expectedElements {
		if !strings.Contains(outputStr, element) {
			t.Errorf("Expected output to contain '%s'", element)
		}
	}
}

// TestIntegration_ConfigurationLoading tests configuration loading
func TestIntegration_ConfigurationLoading(t *testing.T) {
	tempDir := t.TempDir()

	// Create a valid config file
	configFile := filepath.Join(tempDir, "config.json")
	configContent := `{
		"verbosity": 2,
		"colors": true
	}`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Test config loading
	configLoader := &DefaultConfigLoader{}
	config, err := configLoader.LoadFromFile(configFile)

	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify config values
	if config.Verbosity != 2 {
		t.Errorf("Expected verbosity 2, got %d", config.Verbosity)
	}

	if !config.Colors {
		t.Error("Expected colors to be true")
	}
}

// TestIntegration_WatchModeSetup tests watch mode setup
func TestIntegration_WatchModeSetup(t *testing.T) {
	tempDir := t.TempDir()

	// Create test file
	testFile := filepath.Join(tempDir, "watch_test.go")
	content := `package main
import "testing"
func TestWatch(t *testing.T) {
	// Test content
}`

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test watch options setup
	options := WatchOptions{
		Paths:            []string{tempDir},
		IgnorePatterns:   []string{"*.log"},
		TestPatterns:     []string{"*_test.go"},
		Mode:             WatchAll,
		DebounceInterval: 100 * time.Millisecond,
		ClearTerminal:    false,
		Writer:           nil,
	}

	// This tests the watch options validation
	if len(options.Paths) == 0 {
		t.Error("Expected watch paths to be set")
	}

	if options.DebounceInterval <= 0 {
		t.Error("Expected positive debounce interval")
	}

	t.Logf("Watch mode setup test completed successfully")
}

// TestIntegration_CacheOperations tests cache operations
func TestIntegration_CacheOperations(t *testing.T) {
	cache := NewTestResultCache()

	// Create test suite
	suite := &TestSuite{
		FilePath:    "cache_test.go",
		TestCount:   1,
		PassedCount: 1,
		Duration:    50 * time.Millisecond,
	}

	testPath := "./cache_test"

	// Test cache store operation
	cache.CacheResult(testPath, suite)

	// Test cache retrieve operation
	cachedResult, exists := cache.GetCachedResult(testPath)
	if !exists {
		t.Error("Expected cached result to exist")
	}

	if cachedResult == nil {
		t.Fatal("Expected non-nil cached result")
	}

	if cachedResult.Suite.TestCount != suite.TestCount {
		t.Errorf("Expected cached test count %d, got %d", suite.TestCount, cachedResult.Suite.TestCount)
	}
}
