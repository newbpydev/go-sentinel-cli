// Package watcher provides pattern matching capabilities for file paths
package watcher

import (
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
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

// normalizePath ensures forward slashes and cleans the path.
func (pm *PatternMatcher) normalizePath(path string) string {
	if path == "" {
		return "" // Return empty for empty.
	}
	// Explicitly replace backslashes first for cross-platform consistency before Clean.
	s := strings.ReplaceAll(path, "\\", "/")
	// Clean will handle . and .. etc., and ToSlash ensures forward slashes again.
	return filepath.ToSlash(filepath.Clean(s))
}

// MatchesAny implements the PatternMatcher interface
func (pm *PatternMatcher) MatchesAny(path string, patterns []string) bool {
	// Path normalization is handled within MatchesPattern
	for _, pattern := range patterns {
		if pm.MatchesPattern(path, pattern) {
			return true
		}
	}
	return false
}

// MatchesPattern implements the PatternMatcher interface
func (pm *PatternMatcher) MatchesPattern(path string, pattern string) bool {
	if pattern == "" { // Empty pattern never matches anything.
		return false
	}
	// An empty path can only be matched by specific globs like "**" or by an empty pattern.
	if path == "" {
        // doublestar.Match("**", "") is true
        // doublestar.Match("", "") is true only if pattern is "" (already handled)
		return pattern == "**" 
	}

	normalizedPath := pm.normalizePath(path)
	normalizedPattern := pm.normalizePath(pattern) 

	// After normalization, if a pattern became empty (e.g., original was "." or similar), 
	// and path is not empty, it's not a match.
	if normalizedPattern == "" && normalizedPath != "" {
		return false
	}
    // If path normalized to empty (e.g. original was ".") and pattern is not, it's not a match unless pattern is "**"
    if normalizedPath == "" && normalizedPattern != "" {
        return normalizedPattern == "**"
    }
    // If both normalized to empty (e.g. path="." and pattern=".")
    if normalizedPath == "" && normalizedPattern == "" { // This implies original path and pattern were like "." or ""
        return false // "" vs "" test expects false. "." vs "." should be true via exact match if not caught here.
    }


	// Explicit check for pattern "dir/" vs path "dir" -> should be false
	if strings.HasSuffix(normalizedPattern, "/") && normalizedPath == strings.TrimSuffix(normalizedPattern, "/") {
		return false
	}

	// 1. Exact match after full normalization
	if normalizedPath == normalizedPattern { return true }

	// 2. Comprehensive glob match using doublestar.
	if matched, _ := doublestar.Match(normalizedPattern, normalizedPath); matched {
		return true
	}

	// 3. Fallback for simple "directory name" patterns (no globs, no slashes).
	if !strings.ContainsAny(normalizedPattern, "*?[]{}") && !strings.Contains(normalizedPattern, "/") {
		pathComponents := strings.Split(normalizedPath, "/")
		for _, component := range pathComponents {
			if component == normalizedPattern {
				return true
			}
		}
	}

	// 4. Fallback for "filename glob" or "directory component glob" patterns 
	if !strings.Contains(normalizedPattern, "/") && strings.ContainsAny(normalizedPattern, "*?[]{}") {
		baseName := filepath.Base(normalizedPath)
		if baseMatch, _ := doublestar.Match(normalizedPattern, baseName); baseMatch {
			return true
		}
		pathComponents := strings.Split(normalizedPath, "/")
		for _, component := range pathComponents {
			if componentMatch, _ := doublestar.Match(normalizedPattern, component); componentMatch {
				return true
			}
		}
	}
	
    // 5. Handle directory prefix patterns that end with a slash (e.g., "src/").
    //    Path "src" should NOT match "src/". Path "src/foo" SHOULD match "src/".
    if strings.HasSuffix(normalizedPattern, "/") {
        // Ensure path is genuinely inside the directory or is the directory itself (if path also ends with /)
        if strings.HasPrefix(normalizedPath, normalizedPattern) {
            return true
        }
    }
    
    // 6. Handle directory prefix patterns that do NOT end with a slash (e.g. "src/main")
    //    This should match paths that start with pattern + "/"
    //    Example: pattern "src/main" should match "src/main/foo.go"
    if !strings.ContainsAny(normalizedPattern, "*?[]{}") && strings.Contains(normalizedPattern, "/") && !strings.HasSuffix(normalizedPattern, "/"){
        if strings.HasPrefix(normalizedPath, normalizedPattern+"/") {
            return true
        }
    }

	return false
}

// AddPattern implements the PatternMatcher interface
func (pm *PatternMatcher) AddPattern(pattern string) error {
	slashNormalizedPattern := strings.ReplaceAll(pattern, "\\", "/")
	filePattern := core.FilePattern{
		Pattern:       pattern, 
		Type:          core.PatternTypeGlob,
		Recursive:     strings.Contains(slashNormalizedPattern, "**"),
		CaseSensitive: true,
	}
	pm.patterns = append(pm.patterns, filePattern)
	return nil
}

// RemovePattern implements the PatternMatcher interface
func (pm *PatternMatcher) RemovePattern(pattern string) error {
	newPatterns := make([]core.FilePattern, 0, len(pm.patterns))
	for _, p := range pm.patterns {
		if p.Pattern != pattern {
			newPatterns = append(newPatterns, p)
		}
	}
	pm.patterns = newPatterns
	return nil
}
