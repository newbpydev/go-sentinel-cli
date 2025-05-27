// Package watcher provides pattern matching capabilities for file paths
package watcher

import (
	"path/filepath"
	"strings"

	"github.com/newbpydev/go-sentinel/internal/watch/core"
)

// PatternMatcher implements the core.PatternMatcher interface
type PatternMatcher struct {
	patterns []core.FilePattern
}

// NewPatternMatcher creates a new pattern matcher
func NewPatternMatcher() core.PatternMatcher {
	return &PatternMatcher{
		patterns: make([]core.FilePattern, 0),
	}
}

// MatchesAny implements the PatternMatcher interface
func (pm *PatternMatcher) MatchesAny(path string, patterns []string) bool {
	// Normalize path for cross-platform compatibility
	cleanPath := filepath.ToSlash(path)

	for _, pattern := range patterns {
		if pm.MatchesPattern(cleanPath, pattern) {
			return true
		}
	}
	return false
}

// MatchesPattern implements the PatternMatcher interface
func (pm *PatternMatcher) MatchesPattern(path string, pattern string) bool {
	if path == "" || pattern == "" {
		return false
	}

	// Normalize paths for cross-platform compatibility
	normalizedPath := filepath.ToSlash(path)
	normalizedPattern := filepath.ToSlash(pattern)

	// Exact match first
	if normalizedPath == normalizedPattern {
		return true
	}

	// Check exact filename match
	if filepath.Base(normalizedPath) == normalizedPattern {
		return true
	}

	// Wildcard matching using filepath.Match for filename
	if matched, err := filepath.Match(normalizedPattern, filepath.Base(normalizedPath)); err == nil && matched {
		return true
	}

	// Full path wildcard matching
	if matched, err := filepath.Match(normalizedPattern, normalizedPath); err == nil && matched {
		return true
	}

	// Directory pattern matching
	pathComponents := strings.Split(normalizedPath, "/")

	// Check if pattern matches any directory component
	for _, component := range pathComponents {
		if component == normalizedPattern {
			return true
		}
		// Wildcard match against directory component
		if matched, err := filepath.Match(normalizedPattern, component); err == nil && matched {
			return true
		}
	}

	// Directory prefix matching (src/ should match src/main.go)
	if strings.HasSuffix(normalizedPattern, "/") {
		prefix := strings.TrimSuffix(normalizedPattern, "/")
		if strings.HasPrefix(normalizedPath, prefix+"/") {
			return true
		}
	}

	// Directory contains matching (src should match src/main.go)
	if strings.Contains(normalizedPath, "/"+normalizedPattern+"/") ||
		strings.HasPrefix(normalizedPath, normalizedPattern+"/") {
		return true
	}

	// Recursive pattern matching (**)
	if strings.Contains(normalizedPattern, "**") {
		parts := strings.Split(normalizedPattern, "**")
		if len(parts) >= 2 {
			prefix := strings.TrimSuffix(parts[0], "/")
			if prefix == "" || strings.Contains(normalizedPath, prefix) {
				return true
			}
		}
	}

	return false
}

// AddPattern implements the PatternMatcher interface
func (pm *PatternMatcher) AddPattern(pattern string) error {
	filePattern := core.FilePattern{
		Pattern:       pattern,
		Type:          core.PatternTypeGlob, // Default to glob
		Recursive:     strings.Contains(pattern, "**"),
		CaseSensitive: true, // Default to case sensitive
	}

	pm.patterns = append(pm.patterns, filePattern)
	return nil
}

// RemovePattern implements the PatternMatcher interface
func (pm *PatternMatcher) RemovePattern(pattern string) error {
	for i, p := range pm.patterns {
		if p.Pattern == pattern {
			pm.patterns = append(pm.patterns[:i], pm.patterns[i+1:]...)
			return nil
		}
	}
	return nil // Pattern not found, but that's not an error
}
