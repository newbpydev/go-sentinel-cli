package watcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDetectsFileChanges(t *testing.T) {
	dir, dirErr := os.MkdirTemp("", "gosentinel-test-")
	if dirErr != nil {
		t.Fatalf("failed to create temp dir: %v", dirErr)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Logf("warning: failed to clean up temp dir: %v", err)
		}
	}()

	// Create a Go file
	filePath := filepath.Join(dir, "main.go")
	if writeErr := os.WriteFile(filePath, []byte("package main"), 0600); writeErr != nil {
		t.Fatalf("failed to write file: %v", writeErr)
	}

	// Start watcher
	w, watcherErr := NewWatcher(dir)
	if watcherErr != nil {
		t.Fatalf("failed to start watcher: %v", watcherErr)
	}
	defer func() {
		if closeErr := w.Close(); closeErr != nil {
			t.Logf("warning: error closing watcher: %v", closeErr)
		}
	}()

	// Modify file
	if modifyErr := os.WriteFile(filePath, []byte("package main // changed"), 0600); modifyErr != nil {
		t.Fatalf("failed to modify file: %v", modifyErr)
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
	tempDir, err := os.MkdirTemp("", "gosentinel-test-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("warning: failed to clean up temp dir: %v", err)
		}
	}()

	// Setup vendor directory
	vendorDir := filepath.Join(tempDir, "vendor")
	if err := os.Mkdir(vendorDir, 0700); err != nil {
		t.Fatalf("failed to create vendor dir: %v", err)
	}
	vendorFile := filepath.Join(vendorDir, "ignore.go")
	if err := os.WriteFile(vendorFile, []byte("package vendor"), 0600); err != nil {
		t.Fatalf("failed to create vendor file: %v", err)
	}

	// Setup hidden directory
	hiddenDir := filepath.Join(tempDir, ".hidden")
	if err := os.Mkdir(hiddenDir, 0700); err != nil {
		t.Fatalf("failed to create hidden dir: %v", err)
	}
	hiddenFile := filepath.Join(hiddenDir, "ignore.go")
	if err := os.WriteFile(hiddenFile, []byte("package hidden"), 0600); err != nil {
		t.Fatalf("failed to create hidden file: %v", err)
	}

	// Initialize watcher
	w, err := NewWatcher(tempDir)
	if err != nil {
		t.Fatalf("failed to start watcher: %v", err)
	}
	defer func() {
		if err := w.Close(); err != nil {
			t.Logf("warning: error closing watcher: %v", err)
		}
	}()

	// Modify files in excluded directories
	if err := os.WriteFile(vendorFile, []byte("package vendor // changed"), 0600); err != nil {
		t.Fatalf("failed to modify vendor file: %v", err)
	}
	if err := os.WriteFile(hiddenFile, []byte("package hidden // changed"), 0600); err != nil {
		t.Fatalf("failed to modify hidden file: %v", err)
	}

	// Should NOT receive events for these
	select {
	case ev := <-w.Events:
		t.Errorf("unexpected event for excluded dir: %v", ev)
	case <-time.After(300 * time.Millisecond):
		// Success: no event
	}
}

func TestHandlesFileEvents(t *testing.T) {
	tempDir, dirErr := os.MkdirTemp("", "gosentinel-test-")
	if dirErr != nil {
		t.Fatalf("failed to create temp dir: %v", dirErr)
	}
	defer func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			t.Logf("warning: failed to clean up temp dir: %v", removeErr)
		}
	}()

	w, watcherErr := NewWatcher(tempDir)
	if watcherErr != nil {
		t.Fatalf("failed to start watcher: %v", watcherErr)
	}
	defer func() {
		if closeErr := w.Close(); closeErr != nil {
			t.Logf("warning: error closing watcher: %v", closeErr)
		}
	}()

	filePath := filepath.Join(tempDir, "test.go")
	// Create file
	if writeErr := os.WriteFile(filePath, []byte("package test"), 0600); writeErr != nil {
		t.Fatalf("failed to write file: %v", writeErr)
	}
	// Modify file
	if modifyErr := os.WriteFile(filePath, []byte("package test // changed"), 0600); modifyErr != nil {
		t.Fatalf("failed to modify file: %v", modifyErr)
	}
	// Remove file
	if removeErr := os.Remove(filePath); removeErr != nil {
		t.Fatalf("failed to remove file: %v", removeErr)
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
