package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTestResultsFromJSON_ValidFile(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_results.json")
	testData := `[
		{"Time":"2023-01-01T00:00:00Z","Action":"run","Package":"pkg/foo","Test":"TestA"},
		{"Time":"2023-01-01T00:00:01Z","Action":"pass","Package":"pkg/foo","Test":"TestA","Elapsed":0.1}
	]`
	if err := os.WriteFile(tmpFile, []byte(testData), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test loading
	results, err := LoadTestResultsFromJSON(tmpFile)
	if err != nil {
		t.Fatalf("LoadTestResultsFromJSON failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Check first result
	if results[0].Package != "pkg/foo" || results[0].Test != "TestA" {
		t.Errorf("Unexpected first result: %+v", results[0])
	}
}

func TestLoadTestResultsFromJSON_InvalidFile(t *testing.T) {
	// Test with non-existent file
	_, err := LoadTestResultsFromJSON("nonexistent.json")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}

	// Test with invalid JSON
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.json")
	if err := os.WriteFile(tmpFile, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err = LoadTestResultsFromJSON(tmpFile)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestConvertTestResultsToTree_BasicStructure(t *testing.T) {
	results := []TestResult{
		{
			Time:    "2023-01-01T00:00:00Z",
			Action:  "run",
			Package: "pkg/foo",
			Test:    "TestA",
		},
		{
			Time:    "2023-01-01T00:00:01Z",
			Action:  "pass",
			Package: "pkg/foo",
			Test:    "TestA",
			Elapsed: 0.1,
		},
		{
			Time:    "2023-01-01T00:00:02Z",
			Action:  "run",
			Package: "pkg/bar",
			Test:    "TestB",
		},
		{
			Time:    "2023-01-01T00:00:03Z",
			Action:  "fail",
			Package: "pkg/bar",
			Test:    "TestB",
			Elapsed: 0.2,
		},
	}

	root := ConvertTestResultsToTree(results)

	// Check root node
	if root.Title != "" || !root.Expanded {
		t.Errorf("Unexpected root node: %+v", root)
	}

	// Should have one child for "pkg"
	if len(root.Children) != 1 {
		t.Fatalf("Expected 1 top-level package, got %d", len(root.Children))
	}

	pkg := root.Children[0]
	if pkg.Title != "pkg" {
		t.Errorf("Expected package 'pkg', got %q", pkg.Title)
	}

	// Should have two children: foo and bar
	if len(pkg.Children) != 2 {
		t.Fatalf("Expected 2 subpackages, got %d", len(pkg.Children))
	}

	// Check foo package
	foo := pkg.Children[0]
	if foo.Title != "foo" {
		t.Errorf("Expected package 'foo', got %q", foo.Title)
	}
	if len(foo.Children) != 1 {
		t.Fatalf("Expected 1 test in foo, got %d", len(foo.Children))
	}
	testA := foo.Children[0]
	if testA.Title != "TestA" || !*testA.Passed || testA.Duration != 0.1 {
		t.Errorf("Unexpected TestA node: %+v", testA)
	}

	// Check bar package
	bar := pkg.Children[1]
	if bar.Title != "bar" {
		t.Errorf("Expected package 'bar', got %q", bar.Title)
	}
	if len(bar.Children) != 1 {
		t.Fatalf("Expected 1 test in bar, got %d", len(bar.Children))
	}
	testB := bar.Children[0]
	if testB.Title != "TestB" || *testB.Passed || testB.Duration != 0.2 {
		t.Errorf("Unexpected TestB node: %+v", testB)
	}
}

func TestConvertTestResultsToTree_SkippedPackages(t *testing.T) {
	results := []TestResult{
		{
			Time:    "2023-01-01T00:00:00Z",
			Action:  "skip",
			Package: "pkg/empty",
		},
		{
			Time:    "2023-01-01T00:00:01Z",
			Action:  "output",
			Package: "pkg/empty",
			Output:  "?   \tpkg/empty\t[no test files]\n",
		},
	}

	root := ConvertTestResultsToTree(results)

	// Navigate to the empty package node
	pkg := root.Children[0]
	empty := pkg.Children[0]

	if empty.Title != "empty [skip]" {
		t.Errorf("Expected 'empty [skip]', got %q", empty.Title)
	}
	if empty.Error != "skip" {
		t.Errorf("Expected error 'skip', got %q", empty.Error)
	}
}

func TestLooksLikeTestName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"test function", "TestFoo", true},
		{"benchmark", "BenchmarkBar", true},
		{"example", "ExampleBaz", true},
		{"regular function", "foo", false},
		{"empty string", "", false},
		{"test prefix only", "Test", true},
		{"benchmark prefix only", "Benchmark", true},
		{"example prefix only", "Example", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := looksLikeTestName(tc.input); got != tc.expected {
				t.Errorf("looksLikeTestName(%q) = %v, want %v", tc.input, got, tc.expected)
			}
		})
	}
}

func TestSplitPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name:     "simple path",
			path:     "a/b/c",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "path with empty segments",
			path:     "a//b///c",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "single segment",
			path:     "a",
			expected: []string{"a"},
		},
		{
			name:     "empty path",
			path:     "",
			expected: []string{},
		},
		{
			name:     "path with trailing slash",
			path:     "a/b/c/",
			expected: []string{"a", "b", "c"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := splitPath(tc.path)
			if len(got) != len(tc.expected) {
				t.Errorf("splitPath(%q) returned %d segments, want %d", tc.path, len(got), len(tc.expected))
				return
			}
			for i := range got {
				if got[i] != tc.expected[i] {
					t.Errorf("splitPath(%q)[%d] = %q, want %q", tc.path, i, got[i], tc.expected[i])
				}
			}
		})
	}
}

func TestExtractTestName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple test", "ok   pkg/foo/TestAlpha", "TestAlpha"},
		{"no slashes", "TestBeta", "TestBeta"},
		{"empty string", "", ""},
		{"multiple slashes", "a/b/c/TestGamma", "TestGamma"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := extractTestName(tc.input)
			if result != tc.expected {
				t.Errorf("extractTestName(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestGetModulePrefix_ValidFile(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Logf("Warning: failed to restore working directory: %v", err)
		}
	}()

	// Create a temporary go.mod file
	modContent := []byte("module example.com/mymodule\n\ngo 1.21\n")
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), modContent, 0644); err != nil {
		t.Fatalf("Failed to create test go.mod: %v", err)
	}

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Test module prefix extraction
	prefix := getModulePrefix()
	if prefix != "example.com/mymodule" {
		t.Errorf("Expected module prefix 'example.com/mymodule', got %q", prefix)
	}
}

func TestGetModulePrefix_NoFile(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Logf("Warning: failed to restore working directory: %v", err)
		}
	}()

	// Change to temp directory (which has no go.mod)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Test module prefix extraction
	prefix := getModulePrefix()
	if prefix != "" {
		t.Errorf("Expected empty module prefix, got %q", prefix)
	}
}

func TestGetModulePrefix_InvalidFile(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Logf("Warning: failed to restore working directory: %v", err)
		}
	}()

	// Create an invalid go.mod file
	modContent := []byte("invalid content\n")
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), modContent, 0644); err != nil {
		t.Fatalf("Failed to create test go.mod: %v", err)
	}

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Test module prefix extraction
	prefix := getModulePrefix()
	if prefix != "" {
		t.Errorf("Expected empty module prefix, got %q", prefix)
	}
}

func TestGetModulePrefix_EmptyFile(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Logf("Warning: failed to restore working directory: %v", err)
		}
	}()

	// Create an empty go.mod file
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create test go.mod: %v", err)
	}

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Test module prefix extraction
	prefix := getModulePrefix()
	if prefix != "" {
		t.Errorf("Expected empty module prefix, got %q", prefix)
	}
}
