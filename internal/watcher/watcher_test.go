package watcher

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDetectsFileChanges(t *testing.T) {
	dir, err := ioutil.TempDir("", "gosentinel-test-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	// Create a Go file
	filePath := filepath.Join(dir, "main.go")
	if err := ioutil.WriteFile(filePath, []byte("package main"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// Start watcher (to be implemented)
	w, err := NewWatcher(dir)
	if err != nil {
		t.Fatalf("failed to start watcher: %v", err)
	}
	defer w.Close()

	// Modify file
	if err := ioutil.WriteFile(filePath, []byte("package main // changed"), 0644); err != nil {
		t.Fatalf("failed to modify file: %v", err)
	}

	// Expect event
	select {
	case ev := <-w.Events:
		if ev.Name != filePath {
			t.Errorf("expected event for %s, got %s", filePath, ev.Name)
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for file change event")
	}
}

func TestIgnoresVendorAndHiddenDirs(t *testing.T) {
	dir, err := ioutil.TempDir("", "gosentinel-test-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	vendorDir := filepath.Join(dir, "vendor")
	os.Mkdir(vendorDir, 0755)
	vendorFile := filepath.Join(vendorDir, "ignore.go")
	ioutil.WriteFile(vendorFile, []byte("package vendor"), 0644)

	hiddenDir := filepath.Join(dir, ".hidden")
	os.Mkdir(hiddenDir, 0755)
	hiddenFile := filepath.Join(hiddenDir, "ignore.go")
	ioutil.WriteFile(hiddenFile, []byte("package hidden"), 0644)

	w, err := NewWatcher(dir)
	if err != nil {
		t.Fatalf("failed to start watcher: %v", err)
	}
	defer w.Close()

	// Modify vendor file
	ioutil.WriteFile(vendorFile, []byte("package vendor // changed"), 0644)
	// Modify hidden file
	ioutil.WriteFile(hiddenFile, []byte("package hidden // changed"), 0644)

	// Should NOT receive events for these
	select {
	case ev := <-w.Events:
		t.Errorf("unexpected event for excluded dir: %v", ev)
	case <-time.After(300 * time.Millisecond):
		// Success: no event
	}
}

func TestHandlesFileEvents(t *testing.T) {
	dir, err := ioutil.TempDir("", "gosentinel-test-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	w, err := NewWatcher(dir)
	if err != nil {
		t.Fatalf("failed to start watcher: %v", err)
	}
	defer w.Close()

	filePath := filepath.Join(dir, "test.go")
	// Create
	if err := ioutil.WriteFile(filePath, []byte("package test"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	// Write
	if err := ioutil.WriteFile(filePath, []byte("package test // changed"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	// Remove
	if err := os.Remove(filePath); err != nil {
		t.Fatalf("failed to remove file: %v", err)
	}

	received := 0
	for i := 0; i < 3; i++ {
		select {
		case <-w.Events:
			received++
		case <-time.After(time.Second):
			t.Errorf("timeout waiting for file event %d", i)
		}
	}
	if received < 3 {
		t.Errorf("expected 3 file events, got %d", received)
	}
}
