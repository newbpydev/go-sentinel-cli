package watcher

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

type Event struct {
	Name string
}

type Watcher struct {
	Events chan Event
	fsw    *fsnotify.Watcher
	done   chan struct{}
}

func NewWatcher(root string) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	w := &Watcher{
		Events: make(chan Event, 32),
		fsw:    fsw,
		done:   make(chan struct{}),
	}

	// Recursively add directories
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// Permission denied, skip
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if name == "vendor" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			if isSymlink(path) {
				return filepath.SkipDir
			}
			return fsw.Add(path)
		}
		return nil
	})
	if err != nil {
		fsw.Close()
		return nil, err
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
	for {
		select {
		case ev, ok := <-w.fsw.Events:
			if !ok {
				return
			}
			if ev.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove) != 0 {
				if !excluded(ev.Name) {
					w.Events <- Event{Name: ev.Name}
				}
			}
		case <-w.done:
			return
		}
	}
}

func (w *Watcher) Close() error {
	close(w.done)
	w.fsw.Close()
	if w.Events != nil {
		close(w.Events)
		w.Events = nil
		return nil
	}
	return errors.New("watcher already closed")
}

func isSymlink(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}

