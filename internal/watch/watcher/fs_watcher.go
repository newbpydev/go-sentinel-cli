// Package watcher provides file system monitoring capabilities
package watcher

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/newbpydev/go-sentinel/internal/watch/core"
)

// FileSystemWatcher watches for file changes in specified directories
// Implements the core.FileSystemWatcher interface
type FileSystemWatcher struct {
	watcher        *fsnotify.Watcher
	paths          []string
	ignorePatterns []string
	testPatterns   []string
}

// NewFileSystemWatcher creates a new FileSystemWatcher
func NewFileSystemWatcher(paths []string, ignorePatterns []string) (*FileSystemWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	return &FileSystemWatcher{
		watcher:        watcher,
		paths:          paths,
		ignorePatterns: ignorePatterns,
		testPatterns:   []string{"*_test.go"},
	}, nil
}

// Watch starts monitoring for file changes and sends events to the channel
// Implements core.FileSystemWatcher.Watch
func (w *FileSystemWatcher) Watch(ctx context.Context, events chan<- core.FileEvent) error {
	// Add all paths to the watcher
	for _, path := range w.paths {
		if err := w.AddPath(path); err != nil {
			return err
		}
	}

	// Start watching for events
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case event, ok := <-w.watcher.Events:
			if !ok {
				return errors.New("watcher channel closed")
			}

			// Skip events for ignored files
			if w.matchesAnyPattern(event.Name, w.ignorePatterns) {
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
			isTest := w.matchesAnyPattern(event.Name, w.testPatterns)

			// Send the event
			select {
			case events <- core.FileEvent{
				Path:      event.Name,
				Type:      w.eventTypeString(event.Op),
				Timestamp: time.Now(),
				IsTest:    isTest,
			}:
			case <-ctx.Done():
				return ctx.Err()
			}

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return errors.New("watcher error channel closed")
			}
			return fmt.Errorf("watcher error: %w", err)
		}
	}
}

// AddPath adds a new path to be monitored
// Implements core.FileSystemWatcher.AddPath
func (w *FileSystemWatcher) AddPath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

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
				if w.matchesAnyPattern(path, w.ignorePatterns) {
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

	// Add to paths list if not already present
	for _, existingPath := range w.paths {
		if existingPath == path {
			return nil // Already exists
		}
	}
	w.paths = append(w.paths, path)
	return nil
}

// RemovePath removes a path from monitoring
// Implements core.FileSystemWatcher.RemovePath
func (w *FileSystemWatcher) RemovePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for %s: %w", path, err)
	}

	// Remove from fsnotify watcher
	if err := w.watcher.Remove(absPath); err != nil {
		return fmt.Errorf("failed to remove path %s from watcher: %w", absPath, err)
	}

	// Remove from paths list
	for i, existingPath := range w.paths {
		if existingPath == path {
			w.paths = append(w.paths[:i], w.paths[i+1:]...)
			break
		}
	}

	return nil
}

// Close releases all resources used by the watcher
// Implements core.FileSystemWatcher.Close
func (w *FileSystemWatcher) Close() error {
	return w.watcher.Close()
}

// matchesAnyPattern checks if a path matches any of the provided patterns
func (w *FileSystemWatcher) matchesAnyPattern(path string, patterns []string) bool {
	if len(patterns) == 0 {
		return false
	}

	// Use PatternMatcher for consistent pattern matching
	matcher := NewPatternMatcher()
	return matcher.MatchesAny(path, patterns)
}

// eventTypeString converts fsnotify operation to string
func (w *FileSystemWatcher) eventTypeString(op fsnotify.Op) string {
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

// TestFileFinder helps find test files related to implementation files
// Implements the core.TestFileFinder interface
type TestFileFinder struct {
	rootDir string
}

// NewTestFileFinder creates a new TestFileFinder
func NewTestFileFinder(rootDir string) *TestFileFinder {
	return &TestFileFinder{
		rootDir: rootDir,
	}
}

// FindTestFile finds the test file corresponding to the given implementation file
// Implements core.TestFileFinder.FindTestFile
func (f *TestFileFinder) FindTestFile(filePath string) (string, error) {
	if filePath == "" {
		return "", fmt.Errorf("file path cannot be empty")
	}

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

// FindImplementationFile finds the implementation file for a given test file
// Implements core.TestFileFinder.FindImplementationFile
func (f *TestFileFinder) FindImplementationFile(testPath string) (string, error) {
	if testPath == "" {
		return "", fmt.Errorf("test path cannot be empty")
	}

	if !strings.HasSuffix(testPath, "_test.go") {
		return "", fmt.Errorf("not a test file: %s", testPath)
	}

	// Construct the expected implementation file path
	dir := filepath.Dir(testPath)
	base := filepath.Base(testPath)
	name := strings.TrimSuffix(base, "_test.go")
	implFile := filepath.Join(dir, name+".go")

	// Check if the implementation file exists
	_, err := os.Stat(implFile)
	if err != nil {
		return "", fmt.Errorf("implementation file not found for %s: %w", testPath, err)
	}

	return implFile, nil
}

// FindPackageTests finds all test files in the same package as the given file
// Implements core.TestFileFinder.FindPackageTests
func (f *TestFileFinder) FindPackageTests(filePath string) ([]string, error) {
	if filePath == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}

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

// IsTestFile determines if the given file is a test file
// Implements core.TestFileFinder.IsTestFile
func (f *TestFileFinder) IsTestFile(filePath string) bool {
	return strings.HasSuffix(filePath, "_test.go")
}

// Ensure FileSystemWatcher implements the FileSystemWatcher interface
var _ core.FileSystemWatcher = (*FileSystemWatcher)(nil)

// Ensure TestFileFinder implements the TestFileFinder interface
var _ core.TestFileFinder = (*TestFileFinder)(nil)
