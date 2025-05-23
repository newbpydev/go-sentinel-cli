// Package cli provides command-line interface components for the test runner
package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FileEvent represents a file system event
type FileEvent struct {
	Path      string    // Full path to the changed file
	Type      string    // Type of event (create, write, remove, rename, chmod)
	Timestamp time.Time // When the event occurred
	IsTest    bool      // Whether this is a test file
}

// FileWatcher watches for file changes in specified directories
type FileWatcher struct {
	watcher        *fsnotify.Watcher
	paths          []string
	ignorePatterns []string
	testPatterns   []string
}

// NewFileWatcher creates a new FileWatcher
func NewFileWatcher(paths []string, ignorePatterns []string) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	return &FileWatcher{
		watcher:        watcher,
		paths:          paths,
		ignorePatterns: ignorePatterns,
		testPatterns:   []string{"*_test.go"},
	}, nil
}

// Watch starts watching for file changes and sends events to the provided channel
func (w *FileWatcher) Watch(events chan<- FileEvent) error {
	// Add all paths to the watcher
	for _, path := range w.paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for %s: %w", path, err)
		}

		// Add the directory itself
		info, err := os.Stat(absPath)
		if err != nil {
			return fmt.Errorf("failed to stat path %s: %w", absPath, err)
		}

		if info.IsDir() {
			// Walk through all subdirectories
			err = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if info.IsDir() {
					// Skip directories that match ignore patterns
					if matchesAnyPattern(path, w.ignorePatterns) {
						return filepath.SkipDir
					}

					// Add directory to watcher
					if err := w.watcher.Add(path); err != nil {
						return fmt.Errorf("failed to add directory %s to watcher: %w", path, err)
					}
				}
				return nil
			})

			if err != nil {
				return fmt.Errorf("failed to walk directory %s: %w", absPath, err)
			}
		} else {
			// Add the single file to the watcher
			if err := w.watcher.Add(filepath.Dir(absPath)); err != nil {
				return fmt.Errorf("failed to add file %s to watcher: %w", absPath, err)
			}
		}
	}

	// Start watching for events
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return errors.New("watcher channel closed")
			}

			// Skip events for ignored files
			if matchesAnyPattern(event.Name, w.ignorePatterns) {
				continue
			}

			// Only watch for write and create events
			if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
				continue
			}

			// Skip directories
			info, err := os.Stat(event.Name)
			if err == nil && info.IsDir() {
				// Add the new directory to the watcher
				if err := w.watcher.Add(event.Name); err != nil {
					return fmt.Errorf("failed to add new directory %s to watcher: %w", event.Name, err)
				}
				continue
			}

			// Determine if this is a test file
			isTest := matchesAnyPattern(event.Name, w.testPatterns)

			// Send the event
			events <- FileEvent{
				Path:      event.Name,
				Type:      eventTypeString(event.Op),
				Timestamp: time.Now(),
				IsTest:    isTest,
			}

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return errors.New("watcher error channel closed")
			}
			return fmt.Errorf("watcher error: %w", err)
		}
	}
}

// Close closes the watcher
func (w *FileWatcher) Close() error {
	return w.watcher.Close()
}

// TestFileFinder helps find test files related to implementation files
type TestFileFinder struct {
	rootDir string
}

// NewTestFileFinder creates a new TestFileFinder
func NewTestFileFinder(rootDir string) *TestFileFinder {
	return &TestFileFinder{
		rootDir: rootDir,
	}
}

// FindTestFile finds the test file corresponding to the given file
func (f *TestFileFinder) FindTestFile(filePath string) (string, error) {
	// If it's already a test file, just return it
	if strings.HasSuffix(filePath, "_test.go") {
		return filePath, nil
	}

	// Construct the expected test file path
	dir := filepath.Dir(filePath)
	base := filepath.Base(filePath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	testFile := filepath.Join(dir, name+"_test"+ext)

	// Check if the test file exists
	_, err := os.Stat(testFile)
	if err != nil {
		return "", fmt.Errorf("test file not found for %s: %w", filePath, err)
	}

	return testFile, nil
}

// FindAllTestFiles finds all test files in the package containing the given file
func (f *TestFileFinder) FindAllTestFiles(filePath string) ([]string, error) {
	dir := filepath.Dir(filePath)

	// Read all files in the directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	// Filter for test files
	var testFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasSuffix(name, "_test.go") {
			testFiles = append(testFiles, filepath.Join(dir, name))
		}
	}

	if len(testFiles) == 0 {
		return nil, fmt.Errorf("no test files found in %s", dir)
	}

	return testFiles, nil
}

// FindPackageTests finds all tests in the same package as the given file
func (f *TestFileFinder) FindPackageTests(filePath string) ([]string, error) {
	dir := filepath.Dir(filePath)

	// Get the package name from the file
	packageName, err := getGoFilePackage(filePath)
	if err != nil {
		return nil, err
	}

	// Find all Go files in the directory with the same package
	var packageFiles []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".go") {
			continue
		}

		fullPath := filepath.Join(dir, name)
		filePackage, err := getGoFilePackage(fullPath)
		if err != nil {
			continue
		}

		if filePackage == packageName || filePackage == packageName+"_test" {
			packageFiles = append(packageFiles, fullPath)
		}
	}

	// Filter for test files
	var testFiles []string
	for _, file := range packageFiles {
		if strings.HasSuffix(file, "_test.go") {
			testFiles = append(testFiles, file)
		}
	}

	if len(testFiles) == 0 {
		return nil, fmt.Errorf("no test files found for package %s", packageName)
	}

	return testFiles, nil
}

// Utility functions

// matchesAnyPattern checks if the path matches any of the given patterns
func matchesAnyPattern(path string, patterns []string) bool {
	// Normalize path for cross-platform compatibility
	cleanPath := filepath.ToSlash(path)

	for _, pattern := range patterns {
		// Normalize pattern
		pattern = filepath.ToSlash(pattern)

		// Try exact filename match first (most common case)
		matched, err := filepath.Match(pattern, filepath.Base(cleanPath))
		if err == nil && matched {
			return true
		}

		// Check for directory patterns with ** (recursive)
		if strings.Contains(pattern, "**") {
			parts := strings.Split(pattern, "**")
			if len(parts) == 2 && strings.HasPrefix(cleanPath, parts[0]) {
				return true
			}
		}

		// Check for exact directory matches
		if strings.Contains(cleanPath, "/"+pattern+"/") || strings.HasPrefix(cleanPath, pattern+"/") {
			return true
		}

		// Check for wildcard directory patterns (e.g., "*.log", ".git/*")
		if strings.Contains(pattern, "*") {
			if matched, err := filepath.Match(pattern, cleanPath); err == nil && matched {
				return true
			}

			// Check if pattern matches any directory component
			pathParts := strings.Split(cleanPath, "/")
			for _, part := range pathParts {
				if matched, err := filepath.Match(pattern, part); err == nil && matched {
					return true
				}
			}
		}
	}
	return false
}

// eventTypeString converts fsnotify.Op to a string description
func eventTypeString(op fsnotify.Op) string {
	switch {
	case op&fsnotify.Create != 0:
		return "create"
	case op&fsnotify.Write != 0:
		return "write"
	case op&fsnotify.Remove != 0:
		return "remove"
	case op&fsnotify.Rename != 0:
		return "rename"
	case op&fsnotify.Chmod != 0:
		return "chmod"
	default:
		return "unknown"
	}
}

// getGoFilePackage extracts the package name from a Go file
func getGoFilePackage(filePath string) (string, error) {
	// #nosec G304 - Safe file read from Go source files
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "package ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "package")), nil
		}
	}

	return "", fmt.Errorf("no package declaration found in %s", filePath)
}
