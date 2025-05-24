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
	// Normalize pattern
	pattern = filepath.ToSlash(pattern)

	// Try exact filename match first (most common case)
	matched, err := filepath.Match(pattern, filepath.Base(path))
	if err == nil && matched {
		return true
	}

	// Check for directory patterns with ** (recursive)
	if strings.Contains(pattern, "**") {
		parts := strings.Split(pattern, "**")
		if len(parts) == 2 && strings.HasPrefix(path, parts[0]) {
			return true
		}
	}

	// Check for exact directory matches
	if strings.Contains(path, "/"+pattern+"/") || strings.HasPrefix(path, pattern+"/") {
		return true
	}

	// Check for wildcard directory patterns (e.g., "*.log", ".git/*")
	if strings.Contains(pattern, "*") {
		if matched, err := filepath.Match(pattern, path); err == nil && matched {
			return true
		}

		// Check if pattern matches any directory component
		pathParts := strings.Split(path, "/")
		for _, part := range pathParts {
			if matched, err := filepath.Match(pattern, part); err == nil && matched {
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
