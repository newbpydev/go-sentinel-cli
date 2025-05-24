// Package watcher provides file system monitoring capabilities
package watcher

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/newbpydev/go-sentinel/internal/watch/core"
)

// FSWatcher implements the FileSystemWatcher interface using fsnotify
type FSWatcher struct {
	watcher        *fsnotify.Watcher
	paths          []string
	ignorePatterns []string
	testPatterns   []string
	patternMatcher core.PatternMatcher
}

// NewFSWatcher creates a new file system watcher
func NewFSWatcher(paths []string, ignorePatterns []string) (*FSWatcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	return &FSWatcher{
		watcher:        fsWatcher,
		paths:          paths,
		ignorePatterns: ignorePatterns,
		testPatterns:   []string{"*_test.go"},
		patternMatcher: NewPatternMatcher(),
	}, nil
}

// Watch implements the FileSystemWatcher interface
func (w *FSWatcher) Watch(ctx context.Context, events chan<- core.FileEvent) error {
	// Add all paths to the watcher
	for _, path := range w.paths {
		if err := w.AddPath(path); err != nil {
			return fmt.Errorf("failed to add path %s: %w", path, err)
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

			// Process the file system event
			fileEvent := w.processEvent(event)
			if fileEvent != nil {
				events <- *fileEvent
			}

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return errors.New("watcher error channel closed")
			}
			return fmt.Errorf("watcher error: %w", err)
		}
	}
}

// AddPath implements the FileSystemWatcher interface
func (w *FSWatcher) AddPath(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for %s: %w", path, err)
	}

	// Check if path exists
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("failed to stat path %s: %w", absPath, err)
	}

	if info.IsDir() {
		// Walk through all subdirectories
		return filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				// Skip directories that match ignore patterns
				if w.patternMatcher.MatchesAny(path, w.ignorePatterns) {
					return filepath.SkipDir
				}

				// Add directory to watcher
				if err := w.watcher.Add(path); err != nil {
					return fmt.Errorf("failed to add directory %s to watcher: %w", path, err)
				}
			}
			return nil
		})
	} else {
		// Add the directory containing the file
		dir := filepath.Dir(absPath)
		if err := w.watcher.Add(dir); err != nil {
			return fmt.Errorf("failed to add directory %s to watcher: %w", dir, err)
		}
	}

	// Add to paths list if not already present
	for _, existingPath := range w.paths {
		if existingPath == path {
			return nil // Already present
		}
	}
	w.paths = append(w.paths, path)

	return nil
}

// RemovePath implements the FileSystemWatcher interface
func (w *FSWatcher) RemovePath(path string) error {
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

// Close implements the FileSystemWatcher interface
func (w *FSWatcher) Close() error {
	return w.watcher.Close()
}

// processEvent converts fsnotify.Event to core.FileEvent
func (w *FSWatcher) processEvent(event fsnotify.Event) *core.FileEvent {
	// Skip events for ignored files
	if w.patternMatcher.MatchesAny(event.Name, w.ignorePatterns) {
		return nil
	}

	// Only watch for write and create events
	if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
		return nil
	}

	// Skip directories for file events
	info, err := os.Stat(event.Name)
	if err == nil && info.IsDir() {
		// Add new directories to the watcher
		if event.Op&fsnotify.Create != 0 {
			_ = w.watcher.Add(event.Name) // Ignore error for now
		}
		return nil
	}

	// Determine if this is a test file
	isTest := w.patternMatcher.MatchesAny(event.Name, w.testPatterns)

	return &core.FileEvent{
		Path:      event.Name,
		Type:      convertEventType(event.Op),
		Timestamp: time.Now(),
		IsTest:    isTest,
	}
}

// convertEventType converts fsnotify.Op to string
func convertEventType(op fsnotify.Op) string {
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
