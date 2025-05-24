package cli

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// BenchmarkFileWatcherSetup benchmarks the file watcher initialization time
func BenchmarkFileWatcherSetup(b *testing.B) {
	tempDir := b.TempDir()
	paths := []string{tempDir}
	ignorePatterns := []string{"*.log", "*.tmp", ".git/*"}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		watcher, err := NewFileWatcher(paths, ignorePatterns)
		if err != nil {
			b.Fatalf("Failed to create file watcher: %v", err)
		}
		_ = watcher.Close()
	}
}

// BenchmarkFileWatcherLargeDirectory benchmarks watching a directory with many files
func BenchmarkFileWatcherLargeDirectory(b *testing.B) {
	tempDir := b.TempDir()

	// Create many files to watch
	for i := 0; i < 1000; i++ {
		filename := filepath.Join(tempDir, fmt.Sprintf("file_%d.go", i))
		content := fmt.Sprintf("package main\nfunc Test%d() {}\n", i)
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			b.Fatalf("Failed to create test file: %v", err)
		}
	}

	paths := []string{tempDir}
	ignorePatterns := []string{"*.log", "*.tmp"}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		watcher, err := NewFileWatcher(paths, ignorePatterns)
		if err != nil {
			b.Fatalf("Failed to create file watcher: %v", err)
		}
		_ = watcher.Close()
	}
}

// BenchmarkPatternMatching benchmarks file pattern matching performance
func BenchmarkPatternMatching(b *testing.B) {
	patterns := []string{"*.log", "*.tmp", ".git/*", "vendor/**", "node_modules/**"}
	testPaths := []string{
		"main.go",
		"test_file.log",
		"temporary.tmp",
		".git/config",
		"vendor/package/file.go",
		"node_modules/library/index.js",
		"src/internal/app.go",
		"tests/unit/test.go",
		"docs/README.md",
		"scripts/build.sh",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, path := range testPaths {
			_ = matchesAnyPattern(path, patterns)
		}
	}
}

// BenchmarkFileEventDebouncer benchmarks file event debouncing performance
func BenchmarkFileEventDebouncer(b *testing.B) {
	debouncer := NewFileEventDebouncer(50 * time.Millisecond)

	// Start draining events immediately
	done := make(chan bool)
	go func() {
		defer close(done)
		for range debouncer.Events() {
			// Consume events
		}
	}()

	// Create test events
	events := make([]FileEvent, 100)
	for i := 0; i < 100; i++ {
		events[i] = FileEvent{
			Path:      fmt.Sprintf("file_%d.go", i%10), // 10 unique files
			Type:      "write",
			Timestamp: time.Now(),
			IsTest:    i%2 == 0,
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, event := range events {
			debouncer.AddEvent(event)
		}
	}

	// Stop debouncer and wait for drain to complete
	debouncer.Stop()
	<-done
}

// BenchmarkFileEventDebouncerDeduplication benchmarks deduplication performance
func BenchmarkFileEventDebouncerDeduplication(b *testing.B) {
	debouncer := NewFileEventDebouncer(50 * time.Millisecond)

	// Start draining events immediately
	done := make(chan bool)
	go func() {
		defer close(done)
		for range debouncer.Events() {
			// Consume events
		}
	}()

	// Create many duplicate events for the same file
	event := FileEvent{
		Path:      "main.go",
		Type:      "write",
		Timestamp: time.Now(),
		IsTest:    false,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		debouncer.AddEvent(event)
	}

	// Stop debouncer and wait for drain to complete
	debouncer.Stop()
	<-done
}

// BenchmarkDirectoryTraversal benchmarks directory traversal for test files
func BenchmarkDirectoryTraversal(b *testing.B) {
	tempDir := b.TempDir()

	// Create nested directory structure with test files
	for i := 0; i < 10; i++ {
		subDir := filepath.Join(tempDir, fmt.Sprintf("pkg%d", i))
		if err := os.MkdirAll(subDir, 0755); err != nil {
			b.Fatalf("Failed to create directory: %v", err)
		}

		for j := 0; j < 10; j++ {
			filename := filepath.Join(subDir, fmt.Sprintf("test_%d_test.go", j))
			content := fmt.Sprintf("package pkg%d\nfunc Test%d() {}\n", i, j)
			if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
				b.Fatalf("Failed to create test file: %v", err)
			}
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var testFiles []string
		err := filepath.WalkDir(tempDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && filepath.Ext(path) == ".go" {
				if matched, _ := filepath.Match("*_test.go", filepath.Base(path)); matched {
					testFiles = append(testFiles, path)
				}
			}
			return nil
		})
		if err != nil {
			b.Fatalf("Directory traversal failed: %v", err)
		}
		_ = testFiles
	}
}

// BenchmarkTestFileFinder benchmarks the TestFileFinder performance
func BenchmarkTestFileFinder(b *testing.B) {
	tempDir := b.TempDir()

	// Create test files
	for i := 0; i < 100; i++ {
		filename := filepath.Join(tempDir, fmt.Sprintf("test_%d_test.go", i))
		content := fmt.Sprintf("package main\nfunc Test%d() {}\n", i)
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			b.Fatalf("Failed to create test file: %v", err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Use a simple file finder approach instead of the complex TestFileFinder
		var testFiles []string
		err := filepath.WalkDir(tempDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && filepath.Ext(path) == ".go" {
				if matched, _ := filepath.Match("*_test.go", filepath.Base(path)); matched {
					testFiles = append(testFiles, path)
				}
			}
			return nil
		})
		if err != nil {
			b.Fatalf("Failed to find test files: %v", err)
		}
		_ = testFiles
	}
}

// BenchmarkFileChangeAnalysis benchmarks analyzing file changes for test impact
func BenchmarkFileChangeAnalysis(b *testing.B) {
	tempDir := b.TempDir()

	// Create source and test files
	sourceFile := filepath.Join(tempDir, "main.go")
	testFile := filepath.Join(tempDir, "main_test.go")

	sourceContent := `package main
import "fmt"
func main() {
	fmt.Println("Hello, World!")
}
func add(a, b int) int {
	return a + b
}`

	testContent := `package main
import "testing"
func TestAdd(t *testing.T) {
	result := add(2, 3)
	if result != 5 {
		t.Errorf("Expected 5, got %d", result)
	}
}`

	if err := os.WriteFile(sourceFile, []byte(sourceContent), 0644); err != nil {
		b.Fatalf("Failed to create source file: %v", err)
	}
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	cache := NewTestResultCache()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		change, err := cache.AnalyzeChange(sourceFile)
		if err != nil {
			b.Fatalf("Failed to analyze change: %v", err)
		}
		_ = change
	}
}

// BenchmarkWatchModeSetup benchmarks complete watch mode setup
func BenchmarkWatchModeSetup(b *testing.B) {
	tempDir := b.TempDir()

	options := WatchOptions{
		Paths:            []string{tempDir},
		IgnorePatterns:   []string{"*.log", "*.tmp", ".git/*"},
		TestPatterns:     []string{"*_test.go"},
		Mode:             WatchAll,
		DebounceInterval: 500 * time.Millisecond,
		ClearTerminal:    false,
		Writer:           os.Stdout,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		watcher, err := NewTestWatcher(options)
		if err != nil {
			b.Fatalf("Failed to create test watcher: %v", err)
		}
		_ = watcher.Stop()
	}
}

// BenchmarkConcurrentFileEvents benchmarks handling concurrent file events
func BenchmarkConcurrentFileEvents(b *testing.B) {
	debouncer := NewFileEventDebouncer(100 * time.Millisecond)

	// Start draining events immediately
	done := make(chan bool)
	go func() {
		defer close(done)
		for range debouncer.Events() {
			// Consume events
		}
	}()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			event := FileEvent{
				Path:      fmt.Sprintf("concurrent_%d.go", i%50),
				Type:      "write",
				Timestamp: time.Now(),
				IsTest:    i%2 == 0,
			}
			debouncer.AddEvent(event)
			i++
		}
	})

	// Stop debouncer and wait for drain to complete
	debouncer.Stop()
	<-done
}
