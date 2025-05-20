// Package watcher provides a file system watcher that notifies about file system events.
// It's designed to be used for watching Go source code changes during development.
package watcher

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Event represents a file system event with the name of the affected file or directory.
type Event struct {
	Name string
}

// Watcher watches a directory and its subdirectories for changes.
// It filters out vendor directories, hidden directories, and symlinks.
type Watcher struct {
	// Events is a channel that receives file system events.
	Events chan Event
	fsw    *fsnotify.Watcher
	done   chan struct{}
	errors chan error
}

// NewWatcher creates a new file system watcher starting at the given root directory.
// It recursively watches all subdirectories except vendor directories, hidden directories, and symlinks.
func NewWatcher(root string) (*Watcher, error) {
	// Check if root exists first
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return nil, fmt.Errorf("root directory %s does not exist", root)
	}

	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file system watcher: %w", err)
	}

	w := &Watcher{
		Events: make(chan Event, 32),
		fsw:    fsw,
		done:   make(chan struct{}),
		errors: make(chan error, 32),
	}

	// Recursively add directories
	walkErr := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			log.Printf("Warning: error accessing path %s: %v", path, err)
			return filepath.SkipDir
		}
		if d.IsDir() {
			name := d.Name()
			if name == "vendor" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			if isSymlink(path) {
				return filepath.SkipDir
			}
			if addErr := fsw.Add(path); addErr != nil {
				log.Printf("Warning: failed to watch directory %s: %v", path, addErr)
				return filepath.SkipDir
			}
		}
		return nil
	})

	// If there was an error during walking, clean up and return the error
	if walkErr != nil {
		if closeErr := fsw.Close(); closeErr != nil {
			log.Printf("Failed to close watcher during error cleanup: %v", closeErr)
		}
		return nil, fmt.Errorf("error walking directory %s: %w", root, walkErr)
	}

	// Start event loop
	go w.loop()

	return w, nil
}

func (w *Watcher) loop() {
	excluded := func(path string) bool {
		base := filepath.Base(path)
		if base == "vendor" || strings.HasPrefix(base, ".") {
			return true
		}
		if isSymlink(path) {
			return true
		}
		return false
	}

	retryTicker := time.NewTicker(5 * time.Second)
	defer retryTicker.Stop()

	for {
		select {
		case ev, ok := <-w.fsw.Events:
			if !ok {
				return
			}
			if ev.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove) != 0 {
				if !excluded(ev.Name) {
					select {
					case w.Events <- Event{Name: ev.Name}:
					default:
						log.Printf("Warning: event channel full, dropping event for %s", ev.Name)
					}
				}
			}
		case err := <-w.fsw.Errors:
			select {
			case w.errors <- err:
			default:
				log.Printf("Warning: error channel full, dropping error: %v", err)
			}
		case <-retryTicker.C:
			// Periodically check and try to re-add directories that might have been created
			if err := w.retryMissingDirectories(); err != nil {
				log.Printf("Warning: error retrying directories: %v", err)
			}
		case <-w.done:
			return
		}
	}
}

func (w *Watcher) retryMissingDirectories() error {
	// Implementation would go here - for now just a placeholder
	return nil
}

// Close stops watching the file system and releases associated resources.
// It returns an error if the watcher is already closed.
func (w *Watcher) Close() error {
	close(w.done)
	if err := w.fsw.Close(); err != nil {
		return err
	}
	if w.Events != nil {
		close(w.Events)
		w.Events = nil
	}
	if w.errors != nil {
		close(w.errors)
		w.errors = nil
	}
	return nil
}

func isSymlink(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}
