package coverage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func ensureTestdataDir(t *testing.T) string {
	dir := "testdata"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("Failed to create testdata dir: %v", err)
	}
	return dir
}

func TestNewCollector(t *testing.T) {
	dir := ensureTestdataDir(t)
	// Create a temporary coverage profile in ./testdata
	tempFile, err := os.CreateTemp(dir, "coverage-*.out")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func(name string) {
		if removeErr := os.Remove(name); removeErr != nil {
			t.Fatalf("Failed to remove temp file: %v", removeErr)
		}
	}(tempFile.Name())
	// Write a sample cover profile content
	sampleContent := `mode: set
github.com/newbpydev/go-sentinel/internal/sample/sample.go:10.39,12.2 1 1
github.com/newbpydev/go-sentinel/internal/sample/sample.go:14.52,16.13 2 1
github.com/newbpydev/go-sentinel/internal/sample/sample.go:19.2,20.12 2 1
github.com/newbpydev/go-sentinel/internal/sample/sample.go:16.13,18.3 1 0
`
	if _, writeErr := tempFile.Write([]byte(sampleContent)); writeErr != nil {
		t.Fatalf("Failed to write to temp file: %v", writeErr)
	}
	if closeErr := tempFile.Close(); closeErr != nil {
		t.Fatalf("Failed to close temp file: %v", closeErr)
	}

	// Use relative path for collector
	relPath, err := filepath.Rel(".", tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to get relative path: %v", err)
	}
	// Replace backslashes with forward slashes for validation compatibility
	relPath = strings.ReplaceAll(relPath, "\\", "/")
	collector, err := NewCollector(relPath)
	if err != nil {
		t.Fatalf("Failed to create collector: %v", err)
	}

	if collector == nil {
		t.Fatal("Expected collector to be created, got nil")
	}

	if len(collector.Profiles) == 0 {
		t.Error("Expected profiles to be parsed, got none")
	}
}

func TestCalculateMetrics(t *testing.T) {
	dir := ensureTestdataDir(t)
	tempFile, err := os.CreateTemp(dir, "coverage-*.out")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func(name string) {
		if removeErr := os.Remove(name); removeErr != nil {
			t.Fatalf("Failed to remove temp file: %v", removeErr)
		}
	}(tempFile.Name())
	sampleContent := `mode: set
github.com/newbpydev/go-sentinel/internal/sample/sample.go:10.39,12.2 1 1
github.com/newbpydev/go-sentinel/internal/sample/sample.go:14.52,16.13 2 1
github.com/newbpydev/go-sentinel/internal/sample/sample.go:19.2,20.12 2 1
github.com/newbpydev/go-sentinel/internal/sample/sample.go:16.13,18.3 1 0
`
	if _, writeErr := tempFile.Write([]byte(sampleContent)); writeErr != nil {
		t.Fatalf("Failed to write to temp file: %v", writeErr)
	}
	if closeErr := tempFile.Close(); closeErr != nil {
		t.Fatalf("Failed to close temp file: %v", closeErr)
	}

	relPath, err := filepath.Rel(".", tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to get relative path: %v", err)
	}
	// Replace backslashes with forward slashes for validation compatibility
	relPath = strings.ReplaceAll(relPath, "\\", "/")
	collector, err := NewCollector(relPath)
	if err != nil {
		t.Fatalf("Failed to create collector: %v", err)
	}

	metrics, err := collector.CalculateMetrics()
	if err != nil {
		t.Fatalf("Failed to calculate metrics: %v", err)
	}

	if metrics == nil {
		t.Fatal("Expected metrics to be calculated, got nil")
	}

	if metrics.LineCoverage < 74.0 || metrics.LineCoverage > 76.0 {
		t.Errorf("Expected LineCoverage to be around 75%%, got %.2f%%", metrics.LineCoverage)
	}

	if len(metrics.FileMetrics) == 0 {
		t.Error("Expected file metrics to be calculated, got none")
	}
}

func TestFileMetricsCalculation(t *testing.T) {
	dir := ensureTestdataDir(t)
	tempFile, err := os.CreateTemp(dir, "coverage-*.out")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func(name string) {
		if removeErr := os.Remove(name); removeErr != nil {
			t.Fatalf("Failed to remove temp file: %v", removeErr)
		}
	}(tempFile.Name())

	sampleContent := `mode: set
github.com/newbpydev/go-sentinel/internal/sample/sample.go:10.39,12.2 1 1
github.com/newbpydev/go-sentinel/internal/sample/sample.go:14.52,16.13 2 1
github.com/newbpydev/go-sentinel/internal/sample/sample.go:19.2,20.12 2 1
github.com/newbpydev/go-sentinel/internal/sample/sample.go:16.13,18.3 1 0
`
	if _, writeErr := tempFile.Write([]byte(sampleContent)); writeErr != nil {
		t.Fatalf("Failed to write to temp file: %v", writeErr)
	}
	if closeErr := tempFile.Close(); closeErr != nil {
		t.Fatalf("Failed to close temp file: %v", closeErr)
	}

	relPath, err := filepath.Rel(".", tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to get relative path: %v", err)
	}
	// Replace backslashes with forward slashes for validation compatibility
	relPath = strings.ReplaceAll(relPath, "\\", "/")
	collector, err := NewCollector(relPath)
	if err != nil {
		t.Fatalf("Failed to create collector: %v", err)
	}

	metrics, err := collector.CalculateMetrics()
	if err != nil {
		t.Fatalf("Failed to calculate metrics: %v", err)
	}

	sampleFilePath := "github.com/newbpydev/go-sentinel/internal/sample/sample.go"
	fileMetrics, ok := metrics.FileMetrics[sampleFilePath]
	if !ok {
		t.Fatalf("Expected metrics for file %s, not found", sampleFilePath)
	}

	if fileMetrics.LineCoverage < 74.0 || fileMetrics.LineCoverage > 76.0 {
		t.Errorf("Expected file LineCoverage to be around 75%%, got %.2f%%", fileMetrics.LineCoverage)
	}

	if len(fileMetrics.LineExecutionCounts) == 0 {
		t.Error("Expected line execution counts to be calculated, got none")
	}

	if fileMetrics.BranchCoverage < 0 || fileMetrics.BranchCoverage > 100 {
		t.Errorf("Expected valid branch coverage percentage, got %.2f%%", fileMetrics.BranchCoverage)
	}
}
