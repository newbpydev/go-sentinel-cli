package watcher

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/newbpydev/go-sentinel/internal/watch/core"
)

// Test TestFileFinder Factory Function
func TestNewTestFileFinder_Creation(t *testing.T) {
	tests := []struct {
		name    string
		rootDir string
	}{
		{
			name:    "Valid root directory",
			rootDir: ".",
		},
		{
			name:    "Empty root directory",
			rootDir: "",
		},
		{
			name:    "Relative path",
			rootDir: "../",
		},
		{
			name:    "Absolute path",
			rootDir: "/tmp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			finder := NewTestFileFinder(tt.rootDir)

			if finder == nil {
				t.Fatal("NewTestFileFinder should not return nil")
			}

			// Verify interface compliance
			var _ core.TestFileFinder = finder

			// Verify rootDir is set correctly
			if finder.rootDir != tt.rootDir {
				t.Errorf("Expected rootDir %s, got %s", tt.rootDir, finder.rootDir)
			}
		})
	}
}

// Test FindTestFile functionality
func TestTestFileFinder_FindTestFile(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_finder_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	implFile := filepath.Join(tempDir, "example.go")
	testFile := filepath.Join(tempDir, "example_test.go")

	err = os.WriteFile(implFile, []byte("package main"), 0644)
	if err != nil {
		t.Fatalf("Failed to create implementation file: %v", err)
	}

	err = os.WriteFile(testFile, []byte("package main"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	finder := NewTestFileFinder(tempDir)

	tests := []struct {
		name        string
		filePath    string
		expectError bool
		expected    string
		errorMsg    string
	}{
		{
			name:        "Implementation file with existing test",
			filePath:    implFile,
			expectError: false,
			expected:    testFile,
		},
		{
			name:        "Test file returns itself",
			filePath:    testFile,
			expectError: false,
			expected:    testFile,
		},
		{
			name:        "Non-existent implementation file",
			filePath:    filepath.Join(tempDir, "non_existent.go"),
			expectError: true,
			errorMsg:    "test file not found",
		},
		{
			name:        "Implementation file without test",
			filePath:    filepath.Join(tempDir, "orphan.go"),
			expectError: true,
			errorMsg:    "test file not found",
		},
	}

	// Create the orphan.go file for the last test case
	orphanFile := filepath.Join(tempDir, "orphan.go")
	err = os.WriteFile(orphanFile, []byte("package main"), 0644)
	if err != nil {
		t.Fatalf("Failed to create orphan file: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := finder.FindTestFile(tt.filePath)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorMsg != "" && err != nil && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, result)
				}
			}
		})
	}
}

// Test FindImplementationFile functionality
func TestTestFileFinder_FindImplementationFile(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_finder_impl_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	implFile := filepath.Join(tempDir, "service.go")
	testFile := filepath.Join(tempDir, "service_test.go")

	err = os.WriteFile(implFile, []byte("package service"), 0644)
	if err != nil {
		t.Fatalf("Failed to create implementation file: %v", err)
	}

	err = os.WriteFile(testFile, []byte("package service"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	finder := NewTestFileFinder(tempDir)

	tests := []struct {
		name        string
		testPath    string
		expectError bool
		expected    string
		errorMsg    string
	}{
		{
			name:        "Test file with existing implementation",
			testPath:    testFile,
			expectError: false,
			expected:    implFile,
		},
		{
			name:        "Non-test file",
			testPath:    implFile,
			expectError: true,
			errorMsg:    "not a test file",
		},
		{
			name:        "Test file without implementation",
			testPath:    filepath.Join(tempDir, "orphan_test.go"),
			expectError: true,
			errorMsg:    "implementation file not found",
		},
		{
			name:        "Non-existent test file",
			testPath:    filepath.Join(tempDir, "non_existent_test.go"),
			expectError: true,
			errorMsg:    "implementation file not found",
		},
	}

	// Create the orphan test file for testing
	orphanTestFile := filepath.Join(tempDir, "orphan_test.go")
	err = os.WriteFile(orphanTestFile, []byte("package orphan"), 0644)
	if err != nil {
		t.Fatalf("Failed to create orphan test file: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := finder.FindImplementationFile(tt.testPath)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorMsg != "" && err != nil && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, result)
				}
			}
		})
	}
}

// Test FindPackageTests functionality
func TestTestFileFinder_FindPackageTests(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_finder_package_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create multiple files
	files := []struct {
		name     string
		isTest   bool
		expected bool
	}{
		{"service.go", false, false},
		{"service_test.go", true, true},
		{"handler.go", false, false},
		{"handler_test.go", true, true},
		{"utils_test.go", true, true},
		{"config.json", false, false}, // Non-go file
	}

	var expectedTestFiles []string
	for _, file := range files {
		filePath := filepath.Join(tempDir, file.name)
		err = os.WriteFile(filePath, []byte("package test"), 0644)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", file.name, err)
		}

		if file.expected {
			expectedTestFiles = append(expectedTestFiles, filePath)
		}
	}

	finder := NewTestFileFinder(tempDir)

	tests := []struct {
		name          string
		filePath      string
		expectError   bool
		expectedCount int
		errorMsg      string
	}{
		{
			name:          "Directory with test files",
			filePath:      filepath.Join(tempDir, "service.go"),
			expectError:   false,
			expectedCount: 3, // service_test.go, handler_test.go, utils_test.go
		},
		{
			name:        "Non-existent directory",
			filePath:    filepath.Join(tempDir, "non_existent", "file.go"),
			expectError: true,
			errorMsg:    "failed to read directory",
		},
	}

	// Create empty directory for testing
	emptyDir := filepath.Join(tempDir, "empty")
	err = os.Mkdir(emptyDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create empty directory: %v", err)
	}

	tests = append(tests, struct {
		name          string
		filePath      string
		expectError   bool
		expectedCount int
		errorMsg      string
	}{
		name:        "Empty directory",
		filePath:    filepath.Join(emptyDir, "file.go"),
		expectError: true,
		errorMsg:    "no test files found",
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := finder.FindPackageTests(tt.filePath)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorMsg != "" && err != nil && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(result) != tt.expectedCount {
					t.Errorf("Expected %d test files, got %d", tt.expectedCount, len(result))
				}

				// Verify all returned files are test files
				for _, testFile := range result {
					if !finder.IsTestFile(testFile) {
						t.Errorf("Expected %s to be a test file", testFile)
					}
				}
			}
		})
	}
}

// Test IsTestFile functionality
func TestTestFileFinder_IsTestFile(t *testing.T) {
	finder := NewTestFileFinder(".")

	tests := []struct {
		name     string
		filePath string
		expected bool
	}{
		{
			name:     "Valid test file",
			filePath: "example_test.go",
			expected: true,
		},
		{
			name:     "Valid test file with path",
			filePath: "/path/to/service_test.go",
			expected: true,
		},
		{
			name:     "Implementation file",
			filePath: "service.go",
			expected: false,
		},
		{
			name:     "Non-go file",
			filePath: "config.json",
			expected: false,
		},
		{
			name:     "File ending with test but not _test.go",
			filePath: "mytest.go",
			expected: false,
		},
		{
			name:     "Empty string",
			filePath: "",
			expected: false,
		},
		{
			name:     "File with test in middle",
			filePath: "test_helper.go",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := finder.IsTestFile(tt.filePath)
			if result != tt.expected {
				t.Errorf("Expected %v for file %s, got %v", tt.expected, tt.filePath, result)
			}
		})
	}
}

// Test edge cases and error conditions for TestFileFinder
func TestTestFileFinder_EdgeCases(t *testing.T) {
	finder := NewTestFileFinder(".")

	tests := []struct {
		name        string
		operation   func() (interface{}, error)
		expectError bool
		errorMsg    string
	}{
		{
			name: "FindTestFile with empty string",
			operation: func() (interface{}, error) {
				return finder.FindTestFile("")
			},
			expectError: true,
		},
		{
			name: "FindImplementationFile with empty string",
			operation: func() (interface{}, error) {
				return finder.FindImplementationFile("")
			},
			expectError: true,
			errorMsg:    "test path cannot be empty",
		},
		{
			name: "FindPackageTests with empty string",
			operation: func() (interface{}, error) {
				return finder.FindPackageTests("")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := tt.operation()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorMsg != "" && err != nil && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
