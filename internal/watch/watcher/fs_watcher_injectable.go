// Package watcher provides file system monitoring capabilities with dependency injection for 100% test coverage
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

// FsnotifyWatcher interface wraps fsnotify.Watcher for dependency injection
type FsnotifyWatcher interface {
	Add(name string) error
	Remove(name string) error
	Close() error
	Events() <-chan fsnotify.Event
	Errors() <-chan error
}

// FileSystem interface wraps filesystem operations for dependency injection
type FileSystem interface {
	Stat(name string) (os.FileInfo, error)
	Walk(root string, walkFn filepath.WalkFunc) error
	Abs(path string) (string, error)
}

// TimeProvider interface for time operations (dependency injection)
type TimeProvider interface {
	Now() time.Time
}

// WatcherFactory interface for creating watchers (dependency injection)
type WatcherFactory interface {
	NewWatcher() (FsnotifyWatcher, error)
}

// Real implementations for production use

// realFsnotifyWatcher wraps the actual fsnotify.Watcher
type realFsnotifyWatcher struct {
	*fsnotify.Watcher
}

func (w *realFsnotifyWatcher) Events() <-chan fsnotify.Event {
	return w.Watcher.Events
}

func (w *realFsnotifyWatcher) Errors() <-chan error {
	return w.Watcher.Errors
}

// realFileSystem implements FileSystem using actual OS operations
type realFileSystem struct{}

func (fs *realFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (fs *realFileSystem) Walk(root string, walkFn filepath.WalkFunc) error {
	return filepath.Walk(root, walkFn)
}

func (fs *realFileSystem) Abs(path string) (string, error) {
	return filepath.Abs(path)
}

// realTimeProvider implements TimeProvider using actual time
type realTimeProvider struct{}

func (tp *realTimeProvider) Now() time.Time {
	return time.Now()
}

// realWatcherFactory implements WatcherFactory using fsnotify
type realWatcherFactory struct{}

func (f *realWatcherFactory) NewWatcher() (FsnotifyWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &realFsnotifyWatcher{Watcher: watcher}, nil
}

// InjectableFileSystemWatcher with dependency injection for 100% test coverage
type InjectableFileSystemWatcher struct {
	watcher        FsnotifyWatcher
	paths          []string
	ignorePatterns []string
	testPatterns   []string
	fs             FileSystem
	timeProvider   TimeProvider
	factory        WatcherFactory
}

// Dependencies struct for constructor injection
type Dependencies struct {
	FileSystem   FileSystem
	TimeProvider TimeProvider
	Factory      WatcherFactory
}

// NewInjectableFileSystemWatcher creates a new watcher with dependency injection
func NewInjectableFileSystemWatcher(paths []string, ignorePatterns []string, deps *Dependencies) (*InjectableFileSystemWatcher, error) {
	// Use default dependencies if not provided (JIT injection pattern)
	if deps == nil {
		deps = &Dependencies{
			FileSystem:   &realFileSystem{},
			TimeProvider: &realTimeProvider{},
			Factory:      &realWatcherFactory{},
		}
	}

	// Fill in missing dependencies (JIT injection pattern)
	if deps.FileSystem == nil {
		deps.FileSystem = &realFileSystem{}
	}
	if deps.TimeProvider == nil {
		deps.TimeProvider = &realTimeProvider{}
	}
	if deps.Factory == nil {
		deps.Factory = &realWatcherFactory{}
	}

	watcher, err := deps.Factory.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	return &InjectableFileSystemWatcher{
		watcher:        watcher,
		paths:          paths,
		ignorePatterns: ignorePatterns,
		testPatterns:   []string{"*_test.go"},
		fs:             deps.FileSystem,
		timeProvider:   deps.TimeProvider,
		factory:        deps.Factory,
	}, nil
}

// Watch starts monitoring for file changes with full dependency injection
func (w *InjectableFileSystemWatcher) Watch(ctx context.Context, events chan<- core.FileEvent) error {
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

		case event, ok := <-w.watcher.Events():
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
			info, err := w.fs.Stat(event.Name)
			if err == nil && info.IsDir() {
				// Add the new directory to the watcher
				if err := w.watcher.Add(event.Name); err != nil {
					return fmt.Errorf("failed to add new directory %s to watcher: %w", event.Name, err)
				}
				continue
			}

			// Determine if this is a test file
			isTest := w.matchesAnyPattern(event.Name, w.testPatterns)

			// Send the event using injected time provider
			select {
			case events <- core.FileEvent{
				Path:      event.Name,
				Type:      w.eventTypeString(event.Op),
				Timestamp: w.timeProvider.Now(),
				IsTest:    isTest,
			}:
			case <-ctx.Done():
				return ctx.Err()
			}

		case err, ok := <-w.watcher.Errors():
			if !ok {
				return errors.New("watcher error channel closed")
			}
			return fmt.Errorf("watcher error: %w", err)
		}
	}
}

// AddPath with dependency injection for filesystem operations
func (w *InjectableFileSystemWatcher) AddPath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	absPath, err := w.fs.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for %s: %w", path, err)
	}

	// Add the directory itself
	info, err := w.fs.Stat(absPath)
	if err != nil {
		return fmt.Errorf("failed to stat path %s: %w", absPath, err)
	}

	if info.IsDir() {
		// Walk through all subdirectories using injected filesystem
		err = w.fs.Walk(absPath, func(path string, info os.FileInfo, err error) error {
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

// RemovePath with dependency injection
func (w *InjectableFileSystemWatcher) RemovePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	absPath, err := w.fs.Abs(path)
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

// Close releases all resources
func (w *InjectableFileSystemWatcher) Close() error {
	return w.watcher.Close()
}

// matchesAnyPattern remains the same
func (w *InjectableFileSystemWatcher) matchesAnyPattern(path string, patterns []string) bool {
	if len(patterns) == 0 {
		return false
	}

	for _, pattern := range patterns {
		if strings.Contains(path, pattern) {
			return true
		}

		// Simple wildcard matching
		if strings.HasPrefix(pattern, "*") {
			suffix := pattern[1:]
			if strings.HasSuffix(path, suffix) {
				return true
			}
		}

		if strings.HasSuffix(pattern, "*") {
			prefix := pattern[:len(pattern)-1]
			if strings.HasPrefix(path, prefix) {
				return true
			}
		}
	}

	return false
}

// eventTypeString remains the same
func (w *InjectableFileSystemWatcher) eventTypeString(op fsnotify.Op) string {
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
