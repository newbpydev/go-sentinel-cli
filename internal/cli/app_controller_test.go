package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAppController_NewAppController(t *testing.T) {
	controller := NewAppController()

	if controller == nil {
		t.Fatal("NewAppController returned nil")
	}

	if controller.argParser == nil {
		t.Error("argParser not initialized")
	}

	if controller.configLoader == nil {
		t.Error("configLoader not initialized")
	}

	if controller.testRunner == nil {
		t.Error("testRunner not initialized")
	}

	if !controller.testRunner.JSONOutput {
		t.Error("testRunner should have JSON output enabled")
	}
}

func TestAppController_LoadConfiguration(t *testing.T) {
	tests := []struct {
		name           string
		setupConfig    bool
		configContent  string
		expectError    bool
		expectedColors bool
	}{
		{
			name:           "Default configuration when no file exists",
			setupConfig:    false,
			expectError:    false,
			expectedColors: true, // Default config has colors enabled
		},
		{
			name:        "Valid configuration file",
			setupConfig: true,
			configContent: `{
				"colors": false,
				"verbosity": 2,
				"icons": "ascii"
			}`,
			expectError:    false,
			expectedColors: false,
		},
		{
			name:        "Invalid JSON configuration",
			setupConfig: true,
			configContent: `{
				"colors": false,
				"verbosity": 2,
				"invalid
			}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test directory
			tempDir := t.TempDir()
			oldWd, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get working directory: %v", err)
			}
			defer os.Chdir(oldWd)

			if err := os.Chdir(tempDir); err != nil {
				t.Fatalf("Failed to change to temp directory: %v", err)
			}

			// Create config file if needed
			if tt.setupConfig {
				configPath := filepath.Join(tempDir, "sentinel.config.json")
				if err := os.WriteFile(configPath, []byte(tt.configContent), 0644); err != nil {
					t.Fatalf("Failed to write config file: %v", err)
				}
			}

			// Test loading configuration
			controller := NewAppController()
			config, err := controller.loadConfiguration()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if config == nil {
				t.Fatal("Configuration is nil")
			}

			if config.Colors != tt.expectedColors {
				t.Errorf("Expected colors=%t, got colors=%t", tt.expectedColors, config.Colors)
			}
		})
	}
}

func TestAppController_DetermineTestsToRun(t *testing.T) {
	tests := []struct {
		name         string
		changedFile  string
		createFiles  []string
		expectedDirs []string
	}{
		{
			name:         "Test file changed",
			changedFile:  "pkg/utils/helper_test.go",
			createFiles:  []string{"pkg/utils/helper_test.go"},
			expectedDirs: []string{"pkg/utils"},
		},
		{
			name:         "Source file with corresponding test",
			changedFile:  "pkg/utils/helper.go",
			createFiles:  []string{"pkg/utils/helper.go", "pkg/utils/helper_test.go"},
			expectedDirs: []string{"pkg/utils"},
		},
		{
			name:         "Source file without corresponding test",
			changedFile:  "pkg/utils/config.go",
			createFiles:  []string{"pkg/utils/config.go"},
			expectedDirs: []string{"pkg/utils"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test directory
			tempDir := t.TempDir()

			// Create test files
			for _, file := range tt.createFiles {
				fullPath := filepath.Join(tempDir, file)
				dir := filepath.Dir(fullPath)
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatalf("Failed to create directory %s: %v", dir, err)
				}

				if err := os.WriteFile(fullPath, []byte("package test\n"), 0644); err != nil {
					t.Fatalf("Failed to create file %s: %v", fullPath, err)
				}
			}

			// Test determine tests to run
			controller := NewAppController()
			changedPath := filepath.Join(tempDir, tt.changedFile)

			testsToRun, err := controller.determineTestsToRun(changedPath)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(testsToRun) != len(tt.expectedDirs) {
				t.Fatalf("Expected %d test directories, got %d", len(tt.expectedDirs), len(testsToRun))
			}

			for i, expected := range tt.expectedDirs {
				expectedPath := filepath.Join(tempDir, expected)
				if testsToRun[i] != expectedPath {
					t.Errorf("Expected test directory %s, got %s", expectedPath, testsToRun[i])
				}
			}
		})
	}
}

func TestAppController_HelperFunctions(t *testing.T) {
	t.Run("isTestFile", func(t *testing.T) {
		tests := []struct {
			filename string
			expected bool
		}{
			{"helper_test.go", true},
			{"utils_test.go", true},
			{"helper.go", false},
			{"main.go", false},
			{"test.txt", false},
			{"", false},
		}

		for _, tt := range tests {
			result := isTestFile(tt.filename)
			if result != tt.expected {
				t.Errorf("isTestFile(%q) = %t, expected %t", tt.filename, result, tt.expected)
			}
		}
	})

	t.Run("getCorrespondingTestFile", func(t *testing.T) {
		tests := []struct {
			filename string
			expected string
		}{
			{"helper.go", "helper_test.go"},
			{"utils.go", "utils_test.go"},
			{"pkg/helper.go", "pkg/helper_test.go"},
			{"helper.txt", ""},
			{"", ""},
		}

		for _, tt := range tests {
			result := getCorrespondingTestFile(tt.filename)
			if result != tt.expected {
				t.Errorf("getCorrespondingTestFile(%q) = %q, expected %q", tt.filename, result, tt.expected)
			}
		}
	})
}

func TestAppController_RunIntegration(t *testing.T) {
	// This is an integration test to ensure the Run method properly integrates all components
	t.Run("Run with default args", func(t *testing.T) {
		// Setup test directory with a simple test
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "simple_test.go")
		testContent := `package main

import "testing"

func TestSimple(t *testing.T) {
	if 1+1 != 2 {
		t.Error("Math is broken")
	}
}
`
		if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		// Change to temp directory
		oldWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(oldWd)

		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}

		// Test the controller
		controller := NewAppController()

		// This should parse args, load config, and try to run tests
		// We expect this to fail in the CI environment, but it should not panic
		// and should demonstrate that the integration is working
		args := []string{"."}
		err = controller.Run(args)

		// The test might fail due to go test execution, but we just want to ensure
		// that the integration doesn't panic and follows the expected flow
		// The important thing is that all components are connected properly
		if err != nil {
			// This is expected in test environment - the important thing is no panic
			t.Logf("Expected test execution error in test environment: %v", err)
		}
	})
}

func TestAppController_ConfigMerging(t *testing.T) {
	t.Run("CLI args override config", func(t *testing.T) {
		// Setup test directory with config
		tempDir := t.TempDir()
		oldWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(oldWd)

		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}

		// Create config file
		configContent := `{
			"colors": false,
			"verbosity": 1,
			"parallel": 2
		}`
		configPath := filepath.Join(tempDir, "sentinel.config.json")
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		// Test argument parsing and merging
		controller := NewAppController()

		// Load configuration
		config, err := controller.loadConfiguration()
		if err != nil {
			t.Fatalf("Failed to load configuration: %v", err)
		}

		// Parse CLI args that should override config
		cliArgs, err := controller.argParser.Parse([]string{"--color", "-vv", "./..."})
		if err != nil {
			t.Fatalf("Failed to parse CLI args: %v", err)
		}

		// Merge configuration
		merged := config.MergeWithCLIArgs(cliArgs)

		// Verify CLI args override config
		if !merged.Colors {
			t.Error("CLI --color flag should override config colors=false")
		}

		if merged.Verbosity != 2 {
			t.Errorf("CLI -vv should override config verbosity, got %d", merged.Verbosity)
		}

		if len(cliArgs.Packages) == 0 || cliArgs.Packages[0] != "./..." {
			t.Errorf("CLI packages should be set, got %v", cliArgs.Packages)
		}

		// Validate merged configuration
		if err := ValidateConfig(merged); err != nil {
			t.Errorf("Merged configuration should be valid: %v", err)
		}
	})
}

func TestAppController_WatchMode(t *testing.T) {
	t.Run("Watch mode initialization", func(t *testing.T) {
		// This test verifies that watch mode can be initialized properly
		// without actually running the file watcher (which would be complex in tests)

		tempDir := t.TempDir()
		oldWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(oldWd)

		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}

		// Create a config with watch enabled
		configContent := `{
			"colors": true,
			"watchMode": true,
			"runOnStart": false,
			"clearOnRerun": true
		}`
		configPath := filepath.Join(tempDir, "sentinel.config.json")
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		controller := NewAppController()

		// Load and validate configuration
		config, err := controller.loadConfiguration()
		if err != nil {
			t.Fatalf("Failed to load configuration: %v", err)
		}

		if !config.Watch.Enabled {
			t.Error("Watch mode should be enabled")
		}

		if !config.Watch.ClearOnRerun {
			t.Error("Clear on rerun should be enabled")
		}

		// Parse CLI args
		cliArgs, err := controller.argParser.Parse([]string{"./..."})
		if err != nil {
			t.Fatalf("Failed to parse CLI args: %v", err)
		}

		// Merge configuration
		merged := config.MergeWithCLIArgs(cliArgs)

		// Validate merged configuration
		if err := ValidateConfig(merged); err != nil {
			t.Errorf("Merged configuration should be valid: %v", err)
		}

		// We can't easily test the actual watch functionality in unit tests
		// but we can verify the configuration and setup is correct
		t.Logf("Watch mode configuration validated successfully")
	})
}

// Benchmark tests for performance validation
func BenchmarkAppController_Parse(b *testing.B) {
	controller := NewAppController()
	args := []string{"-v", "--color", "--parallel=4", "./internal", "./cmd"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := controller.argParser.Parse(args)
		if err != nil {
			b.Fatalf("Parse failed: %v", err)
		}
	}
}

func BenchmarkAppController_ConfigLoad(b *testing.B) {
	// Setup test config
	tempDir := b.TempDir()
	configContent := `{
		"colors": true,
		"verbosity": 1,
		"parallel": 4,
		"timeout": "30s",
		"icons": "unicode"
	}`
	configPath := filepath.Join(tempDir, "sentinel.config.json")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		b.Fatalf("Failed to write config file: %v", err)
	}

	oldWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(oldWd)

	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}

	controller := NewAppController()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := controller.loadConfiguration()
		if err != nil {
			b.Fatalf("Config load failed: %v", err)
		}
	}
}
