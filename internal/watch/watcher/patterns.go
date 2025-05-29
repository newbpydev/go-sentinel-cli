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
	// Ensure "fmt" is removed from imports if no longer needed (after removing debug statements).

	if pattern == "" {
		return false
	}
	if path == "" {
		// doublestar.Match can handle some empty path cases, e.g. pattern "**"
		// An empty path string specifically should only match "**" or an empty pattern (already handled).
		return pattern == "**" 
	}

	normalizedPath := pm.normalizePath(path) // e.g., path "src" -> "src"

	// Check for the specific case: original pattern "dir/" vs path "dir"
	// This must happen BEFORE 'pattern' is normalized in a way that removes its trailing slash.
	originalPatternEndsWithSlash := strings.HasSuffix(pattern, "/") || strings.HasSuffix(pattern, "\\")
	
	if originalPatternEndsWithSlash {
		// Normalize the pattern *without* its trailing slash for comparison with normalizedPath
		trimmedPatternBase := ""
		if len(pattern) > 1 {
			trimmedPatternBase = pm.normalizePath(pattern[:len(pattern)-1])
		} else { // Original pattern was just "/" or ""
			trimmedPatternBase = pm.normalizePath(pattern) // normalizePath("/") is "/"
		}

		if normalizedPath == trimmedPatternBase && normalizedPath != "" {
			// This means path is "src" and original pattern was "src/" (or "src\").
			// normalizedPath ("src") == trimmedPatternBase ("src") -> true.
			// This should be false.
			return false
		}
	}
    
	normalizedPattern := pm.normalizePath(pattern) // e.g., pattern "src/" becomes "src"; pattern "src" stays "src"

	// Post-normalization checks for empty/'.' paths
	if normalizedPattern == "" && normalizedPath != "" { // Pattern like "." normalized to "", path is not empty
		return false
	}
    if normalizedPath == "" && normalizedPattern != "" { // Path like "." normalized to "", pattern is not empty
        return normalizedPattern == "**" // Only globstar matches effectively empty path
    }
    if normalizedPath == "" && normalizedPattern == "" { // Path and pattern were like "." or empty
        return false // Considered not a match for "empty vs empty"
    }

	// 1. Exact match after full normalization (e.g., path "src", pattern "src")
	if normalizedPath == normalizedPattern {
		return true
	}

	// 2. Comprehensive glob match using doublestar.
	if matched, _ := doublestar.Match(normalizedPattern, normalizedPath); matched {
		return true
	}
    
	// 3. Fallback for simple "directory name" patterns (original pattern had no globs, no slashes).
	//    normalizedPattern will be the directory name.
	if !strings.ContainsAny(normalizedPattern, "*?[]{}") && !strings.Contains(normalizedPattern, "/") {
		pathComponents := strings.Split(normalizedPath, "/")
		for _, component := range pathComponents {
			if component == normalizedPattern {
				return true
			}
		}
	}

	// 4. Fallback for "filename glob" or "directory component glob" patterns 
	//    (original pattern had globs, but no slashes). normalizedPattern is the glob.
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
	
    // 5. Handle directory prefix patterns where the original pattern ended with a slash (e.g., "src/").
    //    The special check at the top already handled `path="src"` vs `pattern="src/"` (returned false).
    //    This section is for `path="src/foo.go"` vs `pattern="src/"`.
    if originalPatternEndsWithSlash {
        // `normalizedPattern` for "src/" is "src".
        // Path "src/foo.go" should match pattern "src/"
        // Check if normalizedPath starts with normalizedPattern + "/"
        // (or just normalizedPattern if normalizedPattern is already "/")
        prefixToMatch := normalizedPattern
        if prefixToMatch == "/" { // If original pattern was just "/"
             // normalizedPath must start with "/" (it will if it's under root)
            if strings.HasPrefix(normalizedPath, prefixToMatch) {
                return true
            }
        } else if prefixToMatch != "" { // For patterns like "src/"
            if strings.HasPrefix(normalizedPath, prefixToMatch + "/") {
                 return true
            }
        }
        // Also consider if normalizedPath is identical to normalizedPattern and original pattern was like "src/"
        // e.g. path "src/", pattern "src/" -> normalizedPath="src/", normalizedPattern="src"
        // This case should be true. The `normalizedPath == trimmedPatternBase` check at top returns false for this.
        // `if normalizedPath == normalizedPattern` (src/ == src) is false.
        // `doublestar.Match("src", "src/")` is false.
        // So, if normalizedPath itself ends with a slash and matches the pattern base:
        if strings.HasSuffix(normalizedPath, "/") && normalizedPath == normalizedPattern + "/" {
            return true
        }
    }
    
    // 6. Handle directory prefix patterns that did NOT originally end with a slash but contain slashes
    //    (e.g. pattern "src/main" should match path "src/main/foo.go")
    //    Here, normalizedPattern is "src/main".
    if !originalPatternEndsWithSlash && strings.Contains(normalizedPattern, "/") {
        if strings.HasPrefix(normalizedPath, normalizedPattern + "/") {
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
