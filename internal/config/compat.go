package config

import "strings"

// convertPackagesToWatchPaths converts Go package patterns to file system paths for watching
// This is a helper function to support Go package syntax in watch mode
func ConvertPackagesToWatchPaths(packages []string) []string {
	var paths []string
	for _, pkg := range packages {
		switch pkg {
		case "./...":
			// Watch current directory and all subdirectories
			paths = append(paths, ".")
		case ".":
			// Watch current directory only
			paths = append(paths, ".")
		default:
			if strings.HasSuffix(pkg, "/...") {
				// Package with recursive subdirectories
				basePath := strings.TrimSuffix(pkg, "/...")
				if basePath == "" {
					basePath = "."
				}
				paths = append(paths, basePath)
			} else {
				// Specific package path
				paths = append(paths, pkg)
			}
		}
	}

	// Remove duplicates
	seen := make(map[string]bool)
	uniquePaths := []string{}
	for _, path := range paths {
		if !seen[path] {
			seen[path] = true
			uniquePaths = append(uniquePaths, path)
		}
	}

	return uniquePaths
}
