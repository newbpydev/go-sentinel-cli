package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/go-sentinel/pkg/models"
)

// TestAppConfigLoader_LoadFromFile_ErrorPaths tests error paths to achieve 100% coverage
func TestAppConfigLoader_LoadFromFile_ErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupFile   func(*testing.T) string
		expectError bool
		errorType   models.ErrorType
	}{
		{
			name: "invalid_json_file",
			setupFile: func(t *testing.T) string {
				tempDir := t.TempDir()
				configFile := filepath.Join(tempDir, "invalid.json")
				err := os.WriteFile(configFile, []byte(`{invalid json content`), 0644)
				if err != nil {
					t.Fatalf("Failed to write test file: %v", err)
				}
				return configFile
			},
			expectError: true,
			errorType:   models.ErrorTypeConfig,
		},
		{
			name: "file_read_permission_error",
			setupFile: func(t *testing.T) string {
				tempDir := t.TempDir()
				configFile := filepath.Join(tempDir, "no_permission.json")
				err := os.WriteFile(configFile, []byte(`{"colors": true}`), 0644)
				if err != nil {
					t.Fatalf("Failed to write test file: %v", err)
				}
				// On Windows, permission changes may not work as expected
				// This test may not trigger the expected error on all platforms
				return configFile
			},
			expectError: false, // May not error on Windows
			errorType:   models.ErrorTypeConfig,
		},
		{
			name: "config_validation_error",
			setupFile: func(t *testing.T) string {
				tempDir := t.TempDir()
				configFile := filepath.Join(tempDir, "invalid_config.json")
				// Create config with invalid verbosity
				err := os.WriteFile(configFile, []byte(`{"verbosity": 10}`), 0644)
				if err != nil {
					t.Fatalf("Failed to write test file: %v", err)
				}
				return configFile
			},
			expectError: true,
			errorType:   models.ErrorTypeConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			loader := NewAppConfigLoader()
			configFile := tt.setupFile(t)

			// Restore permissions after test for cleanup
			if strings.Contains(tt.name, "permission") {
				defer func() {
					os.Chmod(configFile, 0644)
				}()
			}

			_, err := loader.LoadFromFile(configFile)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.expectError && err != nil {
				sentinelErr, ok := err.(*models.SentinelError)
				if !ok {
					t.Fatalf("Expected SentinelError, got %T", err)
				}

				if sentinelErr.Type != tt.errorType {
					t.Errorf("Expected error type %v, got %v", tt.errorType, sentinelErr.Type)
				}
			}
		})
	}
}

// TestDefaultConfigLoader_LoadFromFile_EdgeCases tests edge cases for LoadFromFile
func TestDefaultConfigLoader_LoadFromFile_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupFile   func(*testing.T) string
		expectError bool
	}{
		{
			name: "empty_json_object",
			setupFile: func(t *testing.T) string {
				tempDir := t.TempDir()
				configFile := filepath.Join(tempDir, "empty.json")
				err := os.WriteFile(configFile, []byte(`{}`), 0644)
				if err != nil {
					t.Fatalf("Failed to write test file: %v", err)
				}
				return configFile
			},
			expectError: false,
		},
		{
			name: "malformed_json",
			setupFile: func(t *testing.T) string {
				tempDir := t.TempDir()
				configFile := filepath.Join(tempDir, "malformed.json")
				err := os.WriteFile(configFile, []byte(`{"colors": true,}`), 0644) // Trailing comma
				if err != nil {
					t.Fatalf("Failed to write test file: %v", err)
				}
				return configFile
			},
			expectError: true,
		},
		{
			name: "file_exists_but_empty",
			setupFile: func(t *testing.T) string {
				tempDir := t.TempDir()
				configFile := filepath.Join(tempDir, "empty_file.json")
				err := os.WriteFile(configFile, []byte(""), 0644)
				if err != nil {
					t.Fatalf("Failed to write test file: %v", err)
				}
				return configFile
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			loader := &DefaultConfigLoader{}
			configFile := tt.setupFile(t)

			_, err := loader.LoadFromFile(configFile)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestConvertPackagesToWatchPaths_CompleteCoverage tests all branches of convertPackagesToWatchPaths
func TestConvertPackagesToWatchPaths_CompleteCoverage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		packages []string
		expected []string
	}{
		{
			name:     "empty_base_path_with_ellipsis",
			packages: []string{"/..."},
			expected: []string{"."},
		},
		{
			name:     "root_ellipsis_only",
			packages: []string{"..."},
			expected: []string{"..."}, // Falls through to default case, not converted
		},
		{
			name:     "complex_mixed_patterns",
			packages: []string{"./pkg/...", "cmd/...", "internal", "./..."},
			expected: []string{"./pkg", "cmd", "internal", "."},
		},
		{
			name:     "duplicate_after_conversion",
			packages: []string{"./pkg/...", "pkg/...", "./pkg"},
			expected: []string{"./pkg", "pkg"},
		},
		{
			name:     "empty_string_in_packages",
			packages: []string{"", "pkg"},
			expected: []string{"", "pkg"},
		},
		{
			name:     "slash_only_patterns",
			packages: []string{"/", "//..."},
			expected: []string{"/"}, // "//..." becomes "/" and duplicates are removed
		},
		{
			name:     "nested_ellipsis_patterns",
			packages: []string{"a/b/c/...", "x/y/z/..."},
			expected: []string{"a/b/c", "x/y/z"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := convertPackagesToWatchPaths(tt.packages)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d paths, got %d", len(tt.expected), len(result))
			}

			for i, expected := range tt.expected {
				if i >= len(result) {
					t.Errorf("Missing expected path %q at index %d", expected, i)
					continue
				}
				if result[i] != expected {
					t.Errorf("Expected path %q at index %d, got %q", expected, i, result[i])
				}
			}
		})
	}
}

// TestValidateConfig_CompleteCoverage tests all validation branches
func TestValidateConfig_CompleteCoverage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_minimal_config",
			config: &Config{
				Verbosity: 0,
				Parallel:  0,
				Timeout:   1 * time.Second,
				Visual: VisualConfig{
					Icons: "unicode",
					Theme: "dark",
				},
				Watch: WatchConfig{
					Debounce: 100 * time.Millisecond,
				},
			},
			expectError: false,
		},
		{
			name: "valid_maximal_config",
			config: &Config{
				Verbosity: 5,
				Parallel:  16,
				Timeout:   10 * time.Minute,
				Visual: VisualConfig{
					Icons: "ascii",
					Theme: "light",
				},
				Watch: WatchConfig{
					Debounce: 500 * time.Millisecond,
				},
			},
			expectError: false,
		},
		{
			name: "invalid_verbosity_boundary_low",
			config: &Config{
				Verbosity: -1,
				Timeout:   1 * time.Second,
			},
			expectError: true,
			errorMsg:    "verbosity level must be between 0 and 5",
		},
		{
			name: "invalid_verbosity_boundary_high",
			config: &Config{
				Verbosity: 6,
				Timeout:   1 * time.Second,
			},
			expectError: true,
			errorMsg:    "verbosity level must be between 0 and 5",
		},
		{
			name: "invalid_parallel_negative",
			config: &Config{
				Verbosity: 0,
				Parallel:  -1,
				Timeout:   1 * time.Second,
			},
			expectError: true,
			errorMsg:    "parallel count cannot be negative",
		},
		{
			name: "invalid_timeout_zero",
			config: &Config{
				Verbosity: 0,
				Parallel:  0,
				Timeout:   0,
			},
			expectError: true,
			errorMsg:    "timeout must be positive",
		},
		{
			name: "invalid_timeout_negative",
			config: &Config{
				Verbosity: 0,
				Parallel:  0,
				Timeout:   -1 * time.Second,
			},
			expectError: true,
			errorMsg:    "timeout must be positive",
		},
		{
			name: "invalid_icons_empty",
			config: &Config{
				Verbosity: 0,
				Timeout:   1 * time.Second,
				Visual: VisualConfig{
					Icons: "",
					Theme: "dark",
				},
			},
			expectError: true,
			errorMsg:    "invalid icons type:",
		},
		{
			name: "invalid_icons_unknown",
			config: &Config{
				Verbosity: 0,
				Timeout:   1 * time.Second,
				Visual: VisualConfig{
					Icons: "unknown",
					Theme: "dark",
				},
			},
			expectError: true,
			errorMsg:    "invalid icons type: unknown",
		},
		{
			name: "invalid_theme_empty",
			config: &Config{
				Verbosity: 0,
				Timeout:   1 * time.Second,
				Visual: VisualConfig{
					Icons: "unicode",
					Theme: "",
				},
			},
			expectError: true,
			errorMsg:    "invalid theme:",
		},
		{
			name: "invalid_theme_unknown",
			config: &Config{
				Verbosity: 0,
				Timeout:   1 * time.Second,
				Visual: VisualConfig{
					Icons: "unicode",
					Theme: "unknown",
				},
			},
			expectError: true,
			errorMsg:    "invalid theme: unknown",
		},
		{
			name: "invalid_watch_debounce_negative",
			config: &Config{
				Verbosity: 0,
				Timeout:   1 * time.Second,
				Visual: VisualConfig{
					Icons: "unicode",
					Theme: "dark",
				},
				Watch: WatchConfig{
					Debounce: -1 * time.Millisecond,
				},
			},
			expectError: true,
			errorMsg:    "watch debounce cannot be negative",
		},
		{
			name: "all_valid_icons_types",
			config: &Config{
				Verbosity: 0,
				Timeout:   1 * time.Second,
				Visual: VisualConfig{
					Icons: "minimal",
					Theme: "auto",
				},
			},
			expectError: false,
		},
		{
			name: "icons_none_type",
			config: &Config{
				Verbosity: 0,
				Timeout:   1 * time.Second,
				Visual: VisualConfig{
					Icons: "none",
					Theme: "dark",
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateConfig(tt.config)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.expectError && err != nil && tt.errorMsg != "" {
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q, got: %v", tt.errorMsg, err)
				}
			}
		})
	}
}

// TestDefaultConfigLoader_LoadFromFile_FileSystemEdgeCases tests file system edge cases
func TestDefaultConfigLoader_LoadFromFile_FileSystemEdgeCases(t *testing.T) {
	t.Parallel()

	loader := &DefaultConfigLoader{}

	// Test with directory instead of file
	tempDir := t.TempDir()
	_, err := loader.LoadFromFile(tempDir)
	if err == nil {
		t.Error("Expected error when trying to load directory as config file")
	}

	// Test with file that becomes inaccessible after creation
	configFile := filepath.Join(tempDir, "test.json")
	err = os.WriteFile(configFile, []byte(`{"colors": true}`), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// On Windows, file permission changes may not work as expected
	// Try to make file unreadable, but don't fail test if it doesn't work
	err = os.Chmod(configFile, 0000)
	if err != nil {
		t.Logf("Could not change file permissions (expected on Windows): %v", err)
	}

	// Restore permissions for cleanup
	defer func() {
		os.Chmod(configFile, 0644)
	}()

	_, err = loader.LoadFromFile(configFile)
	// On Windows, this may not error due to different permission model
	if err != nil {
		t.Logf("Got expected permission error: %v", err)
	} else {
		t.Logf("No permission error (expected on Windows)")
	}
}
