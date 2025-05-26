package config

import (
	"reflect"
	"testing"
)

// TestConvertPackagesToWatchPaths tests package to watch path conversion
func TestConvertPackagesToWatchPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		packages []string
		expected []string
	}{
		{
			name:     "empty_packages",
			packages: []string{},
			expected: []string{},
		},
		{
			name:     "current_directory_recursive",
			packages: []string{"./..."},
			expected: []string{"."},
		},
		{
			name:     "current_directory_only",
			packages: []string{"."},
			expected: []string{"."},
		},
		{
			name:     "specific_package",
			packages: []string{"internal/config"},
			expected: []string{"internal/config"},
		},
		{
			name:     "package_with_recursive_subdirs",
			packages: []string{"internal/..."},
			expected: []string{"internal"},
		},
		{
			name:     "empty_base_path_recursive",
			packages: []string{"/..."},
			expected: []string{"."},
		},
		{
			name:     "multiple_packages",
			packages: []string{"internal/config", "pkg/models", "cmd/..."},
			expected: []string{"internal/config", "pkg/models", "cmd"},
		},
		{
			name:     "duplicate_paths",
			packages: []string{".", "./...", "internal", "internal/..."},
			expected: []string{".", "internal"},
		},
		{
			name:     "mixed_patterns",
			packages: []string{"./...", "internal/config", "pkg/...", "cmd/cli"},
			expected: []string{".", "internal/config", "pkg", "cmd/cli"},
		},
		{
			name:     "complex_duplicates",
			packages: []string{"internal/config", "internal/config", "pkg/models", "pkg/models"},
			expected: []string{"internal/config", "pkg/models"},
		},
		{
			name:     "root_recursive_pattern",
			packages: []string{"..."},
			expected: []string{"."},
		},
		{
			name:     "nested_recursive_patterns",
			packages: []string{"a/b/c/...", "x/y/z"},
			expected: []string{"a/b/c", "x/y/z"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := ConvertPackagesToWatchPaths(tt.packages)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestConvertPackagesToWatchPaths_EdgeCases tests edge cases for package conversion
func TestConvertPackagesToWatchPaths_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		packages []string
		expected []string
	}{
		{
			name:     "nil_packages",
			packages: nil,
			expected: []string{},
		},
		{
			name:     "empty_string_package",
			packages: []string{""},
			expected: []string{""},
		},
		{
			name:     "only_ellipsis",
			packages: []string{"..."},
			expected: []string{"."},
		},
		{
			name:     "slash_only",
			packages: []string{"/"},
			expected: []string{"/"},
		},
		{
			name:     "multiple_slashes",
			packages: []string{"//..."},
			expected: []string{"/"},
		},
		{
			name:     "whitespace_packages",
			packages: []string{" ", "  "},
			expected: []string{" ", "  "},
		},
		{
			name:     "special_characters",
			packages: []string{"@pkg", "#internal/...", "$cmd"},
			expected: []string{"@pkg", "#internal", "$cmd"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := ConvertPackagesToWatchPaths(tt.packages)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestConvertPackagesToWatchPaths_DuplicateRemoval tests duplicate removal logic
func TestConvertPackagesToWatchPaths_DuplicateRemoval(t *testing.T) {
	t.Parallel()

	// Test that duplicates are properly removed and order is preserved
	packages := []string{"internal", "pkg", "internal", "cmd", "pkg", "internal"}
	expected := []string{"internal", "pkg", "cmd"}

	result := ConvertPackagesToWatchPaths(packages)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}

	// Verify order preservation
	if len(result) != 3 {
		t.Errorf("Expected 3 unique paths, got %d", len(result))
	}
	if result[0] != "internal" {
		t.Errorf("Expected first path to be 'internal', got %q", result[0])
	}
	if result[1] != "pkg" {
		t.Errorf("Expected second path to be 'pkg', got %q", result[1])
	}
	if result[2] != "cmd" {
		t.Errorf("Expected third path to be 'cmd', got %q", result[2])
	}
}

// TestConvertPackagesToWatchPaths_Performance tests performance with large inputs
func TestConvertPackagesToWatchPaths_Performance(t *testing.T) {
	t.Parallel()

	// Create a large slice with many duplicates
	packages := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		packages[i] = "internal/config" // All the same to test deduplication
	}

	result := ConvertPackagesToWatchPaths(packages)

	// Should deduplicate to just one entry
	if len(result) != 1 {
		t.Errorf("Expected 1 unique path, got %d", len(result))
	}
	if result[0] != "internal/config" {
		t.Errorf("Expected 'internal/config', got %q", result[0])
	}
}
