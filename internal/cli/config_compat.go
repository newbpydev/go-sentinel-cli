package cli

import (
	"strings"

	"github.com/newbpydev/go-sentinel/internal/config"
)

// Re-export types from internal/config for backward compatibility during migration
// These will be removed once all files are migrated to use internal/config directly

// Config re-exports config.Config
type Config = config.Config

// VisualConfig re-exports config.VisualConfig
type VisualConfig = config.VisualConfig

// PathsConfig re-exports config.PathsConfig
type PathsConfig = config.PathsConfig

// WatchConfig re-exports config.WatchConfig
type WatchConfig = config.WatchConfig

// Args re-exports config.Args
type Args = config.Args

// ConfigLoader re-exports config.ConfigLoader
type ConfigLoader = config.ConfigLoader

// DefaultConfigLoader re-exports config.DefaultConfigLoader
type DefaultConfigLoader = config.DefaultConfigLoader

// ArgParser re-exports config.ArgParser
type ArgParser = config.ArgParser

// DefaultArgParser re-exports config.DefaultArgParser
type DefaultArgParser = config.DefaultArgParser

// Re-export functions for backward compatibility

// GetDefaultConfig re-exports config.GetDefaultConfig
var GetDefaultConfig = config.GetDefaultConfig

// ValidateConfig re-exports config.ValidateConfig
var ValidateConfig = config.ValidateConfig

// NewConfigLoader re-exports config.NewConfigLoader
var NewConfigLoader = config.NewConfigLoader

// GetDefaultArgs re-exports config.GetDefaultArgs
var GetDefaultArgs = config.GetDefaultArgs

// ValidateArgs re-exports config.ValidateArgs
var ValidateArgs = config.ValidateArgs

// NewArgParser re-exports config.NewArgParser
var NewArgParser = config.NewArgParser

// convertPackagesToWatchPaths converts Go package patterns to file system paths for watching
// This is a temporary helper function to maintain compatibility during migration
func convertPackagesToWatchPaths(packages []string) []string {
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
