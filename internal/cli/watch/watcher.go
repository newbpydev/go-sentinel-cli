// Package watch provides file system monitoring for test execution
package watch

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/newbpydev/go-sentinel/internal/cli/core"
)

// FileSystemWatcher implements the core.ChangeAnalyzer interface
// while preserving all original watcher functionality
type FileSystemWatcher struct {
	watcher        *fsnotify.Watcher
	paths          []string
	ignorePatterns []string
	testPatterns   []string
}

// NewFileSystemWatcher creates a new file system watcher
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

// AnalyzeChanges implements core.ChangeAnalyzer interface
func (w *FileSystemWatcher) AnalyzeChanges(paths []string) ([]core.FileChange, error) {
	// For initial implementation, convert paths to changes
	changes := make([]core.FileChange, 0, len(paths))

	for _, path := range paths {
		changeType := w.determineChangeType(path)
		changes = append(changes, core.FileChange{
			Path:      path,
			Type:      changeType,
			IsNew:     false,
			Timestamp: time.Now(),
		})
	}

	return changes, nil
}

// WatchForChanges starts watching for file changes and sends events to the provided channel
func (w *FileSystemWatcher) WatchForChanges(events chan<- core.FileChange) error {
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
	}

	// Start watching for events
	for {
		select {
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

			// Determine change type
			changeType := w.determineChangeType(event.Name)

			// Send the event as a FileChange
			change := core.FileChange{
				Path:      event.Name,
				Type:      changeType,
				IsNew:     event.Op&fsnotify.Create != 0,
				Timestamp: time.Now(),
			}

			events <- change

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return errors.New("watcher error channel closed")
			}
			return fmt.Errorf("watcher error: %w", err)
		}
	}
}

// Close closes the watcher
func (w *FileSystemWatcher) Close() error {
	return w.watcher.Close()
}

// Helper methods (preserved from original)

// determineChangeType determines the type of change based on file path
func (w *FileSystemWatcher) determineChangeType(filePath string) core.ChangeType {
	if w.matchesAnyPattern(filePath, w.testPatterns) {
		return core.ChangeTypeTest
	}

	if strings.HasSuffix(filePath, ".go") {
		return core.ChangeTypeSource
	}

	if strings.Contains(filePath, "go.mod") || strings.Contains(filePath, "go.sum") {
		return core.ChangeTypeDependency
	}

	return core.ChangeTypeConfig
}

// matchesAnyPattern checks if a path matches any of the given patterns
func (w *FileSystemWatcher) matchesAnyPattern(path string, patterns []string) bool {
	for _, pattern := range patterns {
		if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
			return true
		}

		// Also check if the path contains the pattern as a substring
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}

// TestFileFinder helps find test files related to implementation files
// Preserved from original implementation
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

	return testFiles, nil
}

// FindPackageTests finds all test files in the same package as the given file
func (f *TestFileFinder) FindPackageTests(filePath string) ([]string, error) {
	// Find all test files in the package
	return f.FindAllTestFiles(filePath)
}

// Legacy compatibility types for smooth migration
type FileEvent struct {
	Path      string    // Full path to the changed file
	Type      string    // Type of event (create, write, remove, rename, chmod)
	Timestamp time.Time // When the event occurred
	IsTest    bool      // Whether this is a test file
}

// FileWatcher provides backward compatibility with the original FileWatcher
type FileWatcher struct {
	fsWatcher *FileSystemWatcher
}

// NewFileWatcher creates a new FileWatcher with backward compatibility
func NewFileWatcher(paths []string, ignorePatterns []string) (*FileWatcher, error) {
	fsWatcher, err := NewFileSystemWatcher(paths, ignorePatterns)
	if err != nil {
		return nil, err
	}

	return &FileWatcher{
		fsWatcher: fsWatcher,
	}, nil
}

// Watch starts watching for file changes (legacy interface)
func (w *FileWatcher) Watch(events chan<- FileEvent) error {
	// Create a channel for the new FileChange events
	newEvents := make(chan core.FileChange, 10)

	// Start the new watcher
	go func() {
		err := w.fsWatcher.WatchForChanges(newEvents)
		if err != nil {
			// Handle error appropriately
			close(newEvents)
		}
	}()

	// Convert new events to legacy events
	for change := range newEvents {
		isTest := change.Type == core.ChangeTypeTest
		eventType := w.convertEventType(change)

		event := FileEvent{
			Path:      change.Path,
			Type:      eventType,
			Timestamp: change.Timestamp,
			IsTest:    isTest,
		}

		events <- event
	}

	return nil
}

// Close closes the watcher
func (w *FileWatcher) Close() error {
	return w.fsWatcher.Close()
}

// convertEventType converts core.ChangeType to legacy event type string
func (w *FileWatcher) convertEventType(change core.FileChange) string {
	if change.IsNew {
		return "create"
	}
	return "write"
}
