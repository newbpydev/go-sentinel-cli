package watcher

import (
	"testing"

	"github.com/newbpydev/go-sentinel/internal/watch/core"
)

// Test NewPatternMatcher factory function
func TestNewPatternMatcher_Creation(t *testing.T) {
	t.Parallel()

	matcher := NewPatternMatcher()

	if matcher == nil {
		t.Fatal("NewPatternMatcher should not return nil")
	}

	// Verify interface compliance
	var _ core.PatternMatcher = matcher

	// Verify it's properly initialized (patterns slice should be empty)
	patternMatcher := matcher.(*PatternMatcher)
	if patternMatcher.patterns == nil {
		t.Error("patterns slice should be initialized")
	}
	if len(patternMatcher.patterns) != 0 {
		t.Error("patterns slice should be empty initially")
	}
}

// Test MatchesAny functionality
func TestPatternMatcher_MatchesAny(t *testing.T) {
	matcher := NewPatternMatcher()

	tests := []struct {
		name     string
		path     string
		patterns []string
		expected bool
	}{
		{
			name:     "Single exact match",
			path:     "main.go",
			patterns: []string{"main.go"},
			expected: true,
		},
		{
			name:     "Multiple patterns - first matches",
			path:     "test.go",
			patterns: []string{"test.go", "*.js", "*.py"},
			expected: true,
		},
		{
			name:     "Multiple patterns - last matches",
			path:     "script.py",
			patterns: []string{"*.js", "*.go", "*.py"},
			expected: true,
		},
		{
			name:     "No match",
			path:     "readme.txt",
			patterns: []string{"*.go", "*.js", "*.py"},
			expected: false,
		},
		{
			name:     "Empty patterns",
			path:     "any.go",
			patterns: []string{},
			expected: false,
		},
		{
			name:     "Wildcard pattern",
			path:     "my_test.go",
			patterns: []string{"*_test.go"},
			expected: true,
		},
		{
			name:     "Directory pattern",
			path:     "src/main.go",
			patterns: []string{"src"},
			expected: true,
		},
		{
			name:     "Cross-platform path",
			path:     "src\\main.go", // Windows-style path
			patterns: []string{"src"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := matcher.MatchesAny(tt.path, tt.patterns)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for path %s with patterns %v", tt.expected, result, tt.path, tt.patterns)
			}
		})
	}
}

// Test MatchesPattern functionality with comprehensive patterns
func TestPatternMatcher_MatchesPattern(t *testing.T) {
	matcher := NewPatternMatcher()

	tests := []struct {
		name     string
		path     string
		pattern  string
		expected bool
	}{
		// Exact matches
		{
			name:     "Exact filename match",
			path:     "main.go",
			pattern:  "main.go",
			expected: true,
		},
		{
			name:     "Exact path match",
			path:     "src/main.go",
			pattern:  "src/main.go",
			expected: true,
		},

		// Wildcard patterns
		{
			name:     "Simple wildcard match",
			path:     "test.go",
			pattern:  "*.go",
			expected: true,
		},
		{
			name:     "Prefix wildcard match",
			path:     "my_test.go",
			pattern:  "*_test.go",
			expected: true,
		},
		{
			name:     "Suffix wildcard match",
			path:     "test_helper.py",
			pattern:  "test_*",
			expected: true,
		},

		// Directory patterns
		{
			name:     "Directory in path",
			path:     "src/utils/helper.go",
			pattern:  "utils",
			expected: true,
		},
		{
			name:     "Directory prefix",
			path:     "src/main.go",
			pattern:  "src/",
			expected: true,
		},
		{
			name:     "Directory exact match",
			path:     "/node_modules/package.json",
			pattern:  "node_modules",
			expected: true,
		},

		// Recursive patterns
		{
			name:     "Recursive pattern - double asterisk",
			path:     "deeply/nested/src/main.go",
			pattern:  "src/**",
			expected: true,
		},
		{
			name:     "Recursive pattern - prefix match",
			path:     "vendor/github.com/pkg/errors/errors.go",
			pattern:  "vendor/**",
			expected: true,
		},

		// Cross-platform paths
		{
			name:     "Windows path normalized",
			path:     "src\\main.go",
			pattern:  "src",
			expected: true,
		},
		{
			name:     "Mixed slashes",
			path:     "src/utils\\helper.go",
			pattern:  "utils",
			expected: true,
		},

		// No matches
		{
			name:     "No match - different extension",
			path:     "main.js",
			pattern:  "*.go",
			expected: false,
		},
		{
			name:     "No match - different directory",
			path:     "tests/main.go",
			pattern:  "src",
			expected: false,
		},
		{
			name:     "No match - exact mismatch",
			path:     "main.go",
			pattern:  "test.go",
			expected: false,
		},

		// Edge cases
		{
			name:     "Empty path",
			path:     "",
			pattern:  "*.go",
			expected: false,
		},
		{
			name:     "Empty pattern",
			path:     "main.go",
			pattern:  "",
			expected: false,
		},
		{
			name:     "Both empty",
			path:     "",
			pattern:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := matcher.MatchesPattern(tt.path, tt.pattern)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for path %s with pattern %s", tt.expected, result, tt.path, tt.pattern)
			}
		})
	}
}

// Test MatchesPattern edge cases for 100% coverage
func TestPatternMatcher_MatchesPattern_EdgeCases(t *testing.T) {
	matcher := NewPatternMatcher()

	tests := []struct {
		name     string
		path     string
		pattern  string
		expected bool
	}{
		// Empty string edge cases
		{
			name:     "Empty path",
			path:     "",
			pattern:  "*.go",
			expected: false,
		},
		{
			name:     "Empty pattern",
			path:     "main.go",
			pattern:  "",
			expected: false,
		},
		{
			name:     "Both empty",
			path:     "",
			pattern:  "",
			expected: false,
		},
		// Directory prefix patterns
		{
			name:     "Directory prefix with slash",
			path:     "src/main.go",
			pattern:  "src/",
			expected: true,
		},
		{
			name:     "Directory prefix without slash",
			path:     "src/main.go",
			pattern:  "src",
			expected: true,
		},
		{
			name:     "Nested directory prefix",
			path:     "src/internal/main.go",
			pattern:  "internal",
			expected: true,
		},
		{
			name:     "Directory contains pattern",
			path:     "project/src/main.go",
			pattern:  "src",
			expected: true,
		},
		// Recursive pattern matching (**)
		{
			name:     "Recursive pattern with prefix",
			path:     "src/deep/nested/main.go",
			pattern:  "src/**",
			expected: true,
		},
		{
			name:     "Recursive pattern without prefix",
			path:     "any/path/main.go",
			pattern:  "**/main.go",
			expected: true,
		},
		{
			name:     "Recursive pattern middle",
			path:     "src/any/deep/main.go",
			pattern:  "src/**/main.go",
			expected: true,
		},
		// Wildcard error cases (invalid patterns)
		{
			name:     "Invalid wildcard pattern",
			path:     "main.go",
			pattern:  "[invalid",
			expected: false,
		},
		// Cross-platform path normalization
		{
			name:     "Windows path normalization",
			path:     "src\\main.go",
			pattern:  "src/main.go",
			expected: true,
		},
		{
			name:     "Mixed path separators",
			path:     "src/internal\\main.go",
			pattern:  "internal",
			expected: true,
		},
		// Complex directory matching
		{
			name:     "Directory at start of path",
			path:     "src/main.go",
			pattern:  "src",
			expected: true,
		},
		{
			name:     "Directory in middle of path",
			path:     "project/src/main.go",
			pattern:  "src",
			expected: true,
		},
		{
			name:     "Directory not matching",
			path:     "project/source/main.go",
			pattern:  "src",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := matcher.MatchesPattern(tt.path, tt.pattern)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for path %q with pattern %q", tt.expected, result, tt.path, tt.pattern)
			}
		})
	}
}

// Test AddPattern functionality
func TestPatternMatcher_AddPattern(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		expectError bool
	}{
		{
			name:        "Simple pattern",
			pattern:     "*.go",
			expectError: false,
		},
		{
			name:        "Directory pattern",
			pattern:     ".git",
			expectError: false,
		},
		{
			name:        "Recursive pattern",
			pattern:     "vendor/**",
			expectError: false,
		},
		{
			name:        "Complex pattern",
			pattern:     "*_test.go",
			expectError: false,
		},
		{
			name:        "Empty pattern",
			pattern:     "",
			expectError: false, // Should be allowed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			matcher := NewPatternMatcher()
			pm := matcher.(*PatternMatcher)

			initialCount := len(pm.patterns)

			err := matcher.AddPattern(tt.pattern)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Verify pattern was added
				if len(pm.patterns) != initialCount+1 {
					t.Errorf("Expected %d patterns, got %d", initialCount+1, len(pm.patterns))
				}

				// Verify the pattern was added correctly
				addedPattern := pm.patterns[len(pm.patterns)-1]
				if addedPattern.Pattern != tt.pattern {
					t.Errorf("Expected pattern %s, got %s", tt.pattern, addedPattern.Pattern)
				}

				// Verify pattern properties are set correctly
				expectedRecursive := contains(tt.pattern, "**")
				if addedPattern.Recursive != expectedRecursive {
					t.Errorf("Expected recursive %v, got %v", expectedRecursive, addedPattern.Recursive)
				}

				if addedPattern.Type != core.PatternTypeGlob {
					t.Errorf("Expected pattern type %v, got %v", core.PatternTypeGlob, addedPattern.Type)
				}

				if !addedPattern.CaseSensitive {
					t.Error("Expected case sensitive to be true")
				}
			}
		})
	}
}

// Test RemovePattern functionality
func TestPatternMatcher_RemovePattern(t *testing.T) {
	tests := []struct {
		name            string
		initialPatterns []string
		removePattern   string
		expectError     bool
		expectedCount   int
	}{
		{
			name:            "Remove existing pattern",
			initialPatterns: []string{"*.go", "*.js", "*.py"},
			removePattern:   "*.js",
			expectError:     false,
			expectedCount:   2,
		},
		{
			name:            "Remove first pattern",
			initialPatterns: []string{"*.go", "*.js"},
			removePattern:   "*.go",
			expectError:     false,
			expectedCount:   1,
		},
		{
			name:            "Remove last pattern",
			initialPatterns: []string{"*.go", "*.js"},
			removePattern:   "*.js",
			expectError:     false,
			expectedCount:   1,
		},
		{
			name:            "Remove non-existent pattern",
			initialPatterns: []string{"*.go", "*.js"},
			removePattern:   "*.py",
			expectError:     false, // Not an error according to implementation
			expectedCount:   2,     // Count should remain same
		},
		{
			name:            "Remove from empty list",
			initialPatterns: []string{},
			removePattern:   "*.go",
			expectError:     false,
			expectedCount:   0,
		},
		{
			name:            "Remove empty pattern",
			initialPatterns: []string{"*.go", "", "*.js"},
			removePattern:   "",
			expectError:     false,
			expectedCount:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			matcher := NewPatternMatcher()

			// Add initial patterns
			for _, pattern := range tt.initialPatterns {
				err := matcher.AddPattern(pattern)
				if err != nil {
					t.Fatalf("Failed to add initial pattern %s: %v", pattern, err)
				}
			}

			// Remove pattern
			err := matcher.RemovePattern(tt.removePattern)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Verify pattern count
				pm := matcher.(*PatternMatcher)
				if len(pm.patterns) != tt.expectedCount {
					t.Errorf("Expected %d patterns, got %d", tt.expectedCount, len(pm.patterns))
				}

				// Verify the specific pattern was removed
				for _, pattern := range pm.patterns {
					if pattern.Pattern == tt.removePattern {
						t.Errorf("Pattern %s should have been removed but still exists", tt.removePattern)
					}
				}
			}
		})
	}
}

// Test pattern types and properties
func TestPatternMatcher_PatternProperties(t *testing.T) {
	matcher := NewPatternMatcher()

	tests := []struct {
		name              string
		pattern           string
		expectedRecursive bool
		expectedType      core.PatternType
	}{
		{
			name:              "Simple glob pattern",
			pattern:           "*.go",
			expectedRecursive: false,
			expectedType:      core.PatternTypeGlob,
		},
		{
			name:              "Recursive pattern",
			pattern:           "vendor/**",
			expectedRecursive: true,
			expectedType:      core.PatternTypeGlob,
		},
		{
			name:              "Directory pattern",
			pattern:           ".git",
			expectedRecursive: false,
			expectedType:      core.PatternTypeGlob,
		},
		{
			name:              "Complex recursive pattern",
			pattern:           "src/**/test/**",
			expectedRecursive: true,
			expectedType:      core.PatternTypeGlob,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := matcher.AddPattern(tt.pattern)
			if err != nil {
				t.Fatalf("Failed to add pattern: %v", err)
			}

			pm := matcher.(*PatternMatcher)
			addedPattern := pm.patterns[len(pm.patterns)-1]

			if addedPattern.Recursive != tt.expectedRecursive {
				t.Errorf("Expected recursive %v, got %v", tt.expectedRecursive, addedPattern.Recursive)
			}

			if addedPattern.Type != tt.expectedType {
				t.Errorf("Expected type %v, got %v", tt.expectedType, addedPattern.Type)
			}

			if !addedPattern.CaseSensitive {
				t.Error("Expected case sensitive to be true by default")
			}
		})
	}
}

// Test edge cases and error conditions
func TestPatternMatcher_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		operation func(*PatternMatcher) interface{}
		verify    func(t *testing.T, result interface{})
	}{
		{
			name: "Multiple add and remove operations",
			operation: func(pm *PatternMatcher) interface{} {
				patterns := []string{"*.go", "*.js", "*.py", "*.go"} // Duplicate
				for _, pattern := range patterns {
					pm.AddPattern(pattern)
				}
				pm.RemovePattern("*.js")
				return len(pm.patterns)
			},
			verify: func(t *testing.T, result interface{}) {
				count := result.(int)
				if count != 3 { // *.go (twice) and *.py
					t.Errorf("Expected 3 patterns after operations, got %d", count)
				}
			},
		},
		{
			name: "MatchesAny with nil patterns",
			operation: func(pm *PatternMatcher) interface{} {
				return pm.MatchesAny("test.go", nil)
			},
			verify: func(t *testing.T, result interface{}) {
				matched := result.(bool)
				if matched {
					t.Error("Expected false for nil patterns")
				}
			},
		},
		{
			name: "MatchesPattern with special characters",
			operation: func(pm *PatternMatcher) interface{} {
				return pm.MatchesPattern("file[1].go", "file*.go")
			},
			verify: func(t *testing.T, result interface{}) {
				matched := result.(bool)
				if !matched {
					t.Error("Expected pattern to match file with brackets")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			matcher := NewPatternMatcher().(*PatternMatcher)
			result := tt.operation(matcher)
			tt.verify(t, result)
		})
	}
}
