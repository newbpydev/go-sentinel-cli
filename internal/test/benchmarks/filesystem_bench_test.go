package benchmarks

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

// BenchmarkTestFileFinder benchmarks finding test files in a directory structure
func BenchmarkTestFileFinder(b *testing.B) {
	tempDir := b.TempDir()

	// Create test files in various subdirectories
	testDirs := []string{"pkg1", "pkg2", "internal/app", "internal/test"}
	for _, dir := range testDirs {
		fullDir := filepath.Join(tempDir, dir)
		if err := os.MkdirAll(fullDir, 0755); err != nil {
			b.Fatalf("Failed to create directory: %v", err)
		}

		for i := 0; i < 20; i++ {
			filename := filepath.Join(fullDir, fmt.Sprintf("test_%d_test.go", i))
			content := fmt.Sprintf("package %s\nfunc Test%d() {}\n", filepath.Base(dir), i)
			if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
				b.Fatalf("Failed to create test file: %v", err)
			}
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		testFiles, err := findTestFiles(tempDir)
		if err != nil {
			b.Fatalf("Failed to find test files: %v", err)
		}
		_ = testFiles
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

// BenchmarkFileChangeAnalysis benchmarks file change analysis performance
func BenchmarkFileChangeAnalysis(b *testing.B) {
	tempDir := b.TempDir()

	// Create test files
	testFiles := make([]string, 100)
	for i := 0; i < 100; i++ {
		filename := filepath.Join(tempDir, fmt.Sprintf("file_%d.go", i))
		content := fmt.Sprintf("package main\nfunc Function%d() {}\n", i)
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			b.Fatalf("Failed to create test file: %v", err)
		}
		testFiles[i] = filename
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, file := range testFiles {
			// Simulate file change analysis
			info, err := os.Stat(file)
			if err != nil {
				continue
			}
			_ = info.ModTime()
			_ = info.Size()
		}
	}
}

// BenchmarkFileReading benchmarks reading multiple files
func BenchmarkFileReading(b *testing.B) {
	tempDir := b.TempDir()

	// Create test files with content
	testFiles := make([]string, 50)
	for i := 0; i < 50; i++ {
		filename := filepath.Join(tempDir, fmt.Sprintf("file_%d.go", i))
		content := fmt.Sprintf("package main\n\nfunc Function%d() {\n\t// Some code here\n\treturn\n}\n", i)
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			b.Fatalf("Failed to create test file: %v", err)
		}
		testFiles[i] = filename
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, file := range testFiles {
			content, err := os.ReadFile(file)
			if err != nil {
				continue
			}
			_ = len(content)
		}
	}
}

// BenchmarkFileWriting benchmarks writing to multiple files
func BenchmarkFileWriting(b *testing.B) {
	tempDir := b.TempDir()

	content := []byte("package main\n\nfunc BenchmarkFunction() {\n\t// Benchmark content\n}\n")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		filename := filepath.Join(tempDir, fmt.Sprintf("bench_file_%d.go", i))
		if err := os.WriteFile(filename, content, 0644); err != nil {
			b.Fatalf("Failed to write file: %v", err)
		}
	}
}

// Helper functions

// findTestFiles finds all test files in a directory
func findTestFiles(rootDir string) ([]string, error) {
	var testFiles []string

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, "_test.go") {
			testFiles = append(testFiles, path)
		}

		return nil
	})

	return testFiles, err
}

// matchesAnyPattern checks if a path matches any of the given patterns
func matchesAnyPattern(path string, patterns []string) bool {
	for _, pattern := range patterns {
		// Simple pattern matching - in real implementation would use more sophisticated matching
		if strings.Contains(pattern, "*") {
			// Handle simple wildcard patterns
			if strings.HasPrefix(pattern, "*.") {
				ext := strings.TrimPrefix(pattern, "*")
				if strings.HasSuffix(path, ext) {
					return true
				}
			}
		} else {
			// Exact match or contains
			if strings.Contains(path, pattern) {
				return true
			}
		}
	}
	return false
}
