package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/newbpydev/go-sentinel/pkg/models"
)

// TestNewTestResultCache_Creation verifies cache initialization
func TestNewTestResultCache_Creation(t *testing.T) {
	// Act
	cache := NewTestResultCache()

	// Assert
	if cache == nil {
		t.Fatal("Expected cache to be created, got nil")
	}
	if cache.results == nil {
		t.Error("Expected results map to be initialized")
	}
	if cache.fileTimes == nil {
		t.Error("Expected fileTimes map to be initialized")
	}
	if cache.testTimes == nil {
		t.Error("Expected testTimes map to be initialized")
	}
	if len(cache.results) != 0 {
		t.Error("Expected results to be empty initially")
	}
	if len(cache.fileTimes) != 0 {
		t.Error("Expected fileTimes to be empty initially")
	}
	if len(cache.testTimes) != 0 {
		t.Error("Expected testTimes to be empty initially")
	}
}

// TestAnalyzeChange_TestFile tests analyzing changes to test files
func TestAnalyzeChange_TestFile(t *testing.T) {
	// Arrange
	cache := NewTestResultCache()

	// Create a temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "example_test.go")
	err := os.WriteFile(testFile, []byte("package main\nfunc TestExample(t *testing.T) {}"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Act
	change, err := cache.AnalyzeChange(testFile)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if change == nil {
		t.Fatal("Expected change to be returned, got nil")
	}
	if change.Path != testFile {
		t.Errorf("Expected path '%s', got '%s'", testFile, change.Path)
	}
	if change.Type != ChangeTypeTest {
		t.Errorf("Expected ChangeTypeTest, got %v", change.Type)
	}
	if !change.IsNew {
		t.Error("Expected IsNew to be true for new file")
	}
	if change.Hash == "" {
		t.Error("Expected hash to be calculated")
	}
}

// TestAnalyzeChange_SourceFile tests analyzing changes to source files
func TestAnalyzeChange_SourceFile(t *testing.T) {
	// Arrange
	cache := NewTestResultCache()

	// Create a temporary source file
	tempDir := t.TempDir()
	sourceFile := filepath.Join(tempDir, "example.go")
	err := os.WriteFile(sourceFile, []byte("package main\nfunc Hello() string { return \"world\" }"), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Act
	change, err := cache.AnalyzeChange(sourceFile)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if change.Type != ChangeTypeSource {
		t.Errorf("Expected ChangeTypeSource, got %v", change.Type)
	}
}

// TestAnalyzeChange_ConfigFile tests analyzing changes to config files
func TestAnalyzeChange_ConfigFile(t *testing.T) {
	// Arrange
	cache := NewTestResultCache()

	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "go.mod")
	err := os.WriteFile(configFile, []byte("module example.com/test\ngo 1.19"), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Act
	change, err := cache.AnalyzeChange(configFile)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if change.Type != ChangeTypeDependency {
		t.Errorf("Expected ChangeTypeDependency, got %v", change.Type)
	}
}

// TestAnalyzeChange_NonexistentFile tests analyzing changes to nonexistent files
func TestAnalyzeChange_NonexistentFile(t *testing.T) {
	// Arrange
	cache := NewTestResultCache()
	nonexistentFile := "/path/to/nonexistent/file.go"

	// Act
	change, err := cache.AnalyzeChange(nonexistentFile)

	// Assert
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
	if change != nil {
		t.Error("Expected change to be nil for nonexistent file")
	}
}

// TestMarkFileAsProcessed_UpdatesFileTime tests file processing tracking
func TestMarkFileAsProcessed_UpdatesFileTime(t *testing.T) {
	// Arrange
	cache := NewTestResultCache()
	filePath := "test.go"
	processTime := time.Now()

	// Act
	cache.MarkFileAsProcessed(filePath, processTime)

	// Assert
	cache.mutex.RLock()
	storedTime, exists := cache.fileTimes[filePath]
	cache.mutex.RUnlock()

	if !exists {
		t.Error("Expected file time to be stored")
	}
	if !storedTime.Equal(processTime) {
		t.Errorf("Expected stored time %v, got %v", processTime, storedTime)
	}
}

// TestShouldRunTests_NoChanges tests when no changes are provided
func TestShouldRunTests_NoChanges(t *testing.T) {
	// Arrange
	cache := NewTestResultCache()
	var changes []*FileChange

	// Act
	shouldRun, testPaths := cache.ShouldRunTests(changes)

	// Assert
	if shouldRun {
		t.Error("Expected shouldRun to be false with no changes")
	}
	if testPaths != nil {
		t.Error("Expected testPaths to be nil with no changes")
	}
}

// TestShouldRunTests_WithNewChanges tests when new changes are provided
func TestShouldRunTests_WithNewChanges(t *testing.T) {
	// Arrange
	cache := NewTestResultCache()
	changes := []*FileChange{
		{
			Path:  "test.go",
			Type:  ChangeTypeTest,
			IsNew: true,
		},
	}

	// Act
	shouldRun, testPaths := cache.ShouldRunTests(changes)

	// Assert
	if !shouldRun {
		t.Error("Expected shouldRun to be true with new changes")
	}
	if len(testPaths) == 0 {
		t.Error("Expected testPaths to contain paths")
	}
}

// TestGetStaleTests_TestFileChange tests stale test detection for test file changes
func TestGetStaleTests_TestFileChange(t *testing.T) {
	// Arrange
	cache := NewTestResultCache()
	changes := []*FileChange{
		{
			Path: "pkg/example/example_test.go",
			Type: ChangeTypeTest,
		},
	}

	// Act
	staleTests := cache.GetStaleTests(changes)

	// Assert
	if len(staleTests) != 1 {
		t.Errorf("Expected 1 stale test, got %d", len(staleTests))
	}
	expectedPath := filepath.Dir("pkg/example/example_test.go")
	if staleTests[0] != expectedPath {
		t.Errorf("Expected '%s', got '%s'", expectedPath, staleTests[0])
	}
}

// TestGetStaleTests_ConfigChange tests stale test detection for config changes
func TestGetStaleTests_ConfigChange(t *testing.T) {
	// Arrange
	cache := NewTestResultCache()

	// Add some cached results first
	suite := &models.TestSuite{
		FilePath: "test1",
		Duration: time.Millisecond,
	}
	cache.CacheResult("test1", suite)
	cache.CacheResult("test2", suite)

	changes := []*FileChange{
		{
			Path: "go.mod",
			Type: ChangeTypeConfig,
		},
	}

	// Act
	staleTests := cache.GetStaleTests(changes)

	// Assert
	if len(staleTests) != 2 {
		t.Errorf("Expected 2 stale tests, got %d", len(staleTests))
	}
}

// TestCacheResult_StoresResult tests caching test results
func TestCacheResult_StoresResult(t *testing.T) {
	// Arrange
	cache := NewTestResultCache()
	testPath := "pkg/example"
	suite := &models.TestSuite{
		FilePath:     "pkg/example/example_test.go",
		TestCount:    5,
		PassedCount:  4,
		FailedCount:  1,
		SkippedCount: 0,
		Duration:     100 * time.Millisecond,
	}

	// Act
	cache.CacheResult(testPath, suite)

	// Assert
	cache.mutex.RLock()
	cached, exists := cache.results[testPath]
	cache.mutex.RUnlock()

	if !exists {
		t.Error("Expected result to be cached")
	}
	if cached.Suite != suite {
		t.Error("Expected cached suite to match original")
	}
	if cached.Status != models.StatusFailed {
		t.Errorf("Expected status to be Failed, got %v", cached.Status)
	}
	if cached.DependsOn == nil {
		t.Error("Expected DependsOn field to be initialized")
	}
}

// TestGetCachedResult_ValidResult tests retrieving valid cached results
func TestGetCachedResult_ValidResult(t *testing.T) {
	// Arrange
	cache := NewTestResultCache()
	testPath := "pkg/example"
	suite := &models.TestSuite{
		FilePath: "pkg/example/example_test.go",
		Duration: time.Millisecond,
	}
	cache.CacheResult(testPath, suite)

	// Act
	result, exists := cache.GetCachedResult(testPath)

	// Assert
	if !exists {
		t.Error("Expected cached result to exist")
	}
	if result == nil {
		t.Error("Expected result to be returned")
	}
	if result.Suite != suite {
		t.Error("Expected cached suite to match original")
	}
}

// TestGetCachedResult_NonexistentResult tests retrieving nonexistent results
func TestGetCachedResult_NonexistentResult(t *testing.T) {
	// Arrange
	cache := NewTestResultCache()
	testPath := "nonexistent/path"

	// Act
	result, exists := cache.GetCachedResult(testPath)

	// Assert
	if exists {
		t.Error("Expected cached result to not exist")
	}
	if result != nil {
		t.Error("Expected result to be nil")
	}
}

// TestGetCachedResult_InvalidatedByDependency tests cache invalidation
func TestGetCachedResult_InvalidatedByDependency(t *testing.T) {
	// Arrange
	cache := NewTestResultCache()
	testPath := "pkg/example"
	suite := &models.TestSuite{
		FilePath: "pkg/example/example_test.go",
		Duration: time.Millisecond,
	}

	// Cache a result
	cache.CacheResult(testPath, suite)

	// Simulate a dependency change after caching
	time.Sleep(10 * time.Millisecond)
	cache.MarkFileAsProcessed("go.mod", time.Now())

	// Act
	result, exists := cache.GetCachedResult(testPath)

	// Assert - result should still exist since go.mod wasn't a dependency when cached
	if !exists {
		t.Error("Expected cached result to exist")
	}
	if result == nil {
		t.Error("Expected result to be returned")
	}
}

// TestClear_RemovesAllData tests clearing all cached data
func TestClear_RemovesAllData(t *testing.T) {
	// Arrange
	cache := NewTestResultCache()
	suite := &models.TestSuite{
		FilePath: "test.go",
		Duration: time.Millisecond,
	}
	cache.CacheResult("test", suite)
	cache.MarkFileAsProcessed("file.go", time.Now())

	// Act
	cache.Clear()

	// Assert
	if len(cache.results) != 0 {
		t.Error("Expected results to be cleared")
	}
	if len(cache.fileTimes) != 0 {
		t.Error("Expected fileTimes to be cleared")
	}
	if len(cache.testTimes) != 0 {
		t.Error("Expected testTimes to be cleared")
	}
}

// TestGetStats_ReturnsCorrectStats tests statistics retrieval
func TestGetStats_ReturnsCorrectStats(t *testing.T) {
	// Arrange
	cache := NewTestResultCache()
	suite := &models.TestSuite{
		FilePath: "test.go",
		Duration: time.Millisecond,
	}
	cache.CacheResult("test1", suite)
	cache.CacheResult("test2", suite)
	cache.MarkFileAsProcessed("file1.go", time.Now())
	cache.MarkFileAsProcessed("file2.go", time.Now())
	cache.MarkFileAsProcessed("file3.go", time.Now())

	// Act
	stats := cache.GetStats()

	// Assert
	if stats["cached_results"] != 2 {
		t.Errorf("Expected 2 cached results, got %v", stats["cached_results"])
	}
	if stats["tracked_files"] != 3 {
		t.Errorf("Expected 3 tracked files, got %v", stats["tracked_files"])
	}
	if stats["tracked_tests"] != 2 {
		t.Errorf("Expected 2 tracked tests, got %v", stats["tracked_tests"])
	}
}

// TestConcurrentAccess_SafeAccess tests thread-safe access
func TestConcurrentAccess_SafeAccess(t *testing.T) {
	// Arrange
	cache := NewTestResultCache()
	suite := &models.TestSuite{
		FilePath: "test.go",
		Duration: time.Millisecond,
	}

	// Act - concurrent operations
	done := make(chan bool, 3)

	go func() {
		for i := 0; i < 100; i++ {
			cache.CacheResult("test", suite)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			cache.GetCachedResult("test")
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			cache.GetStats()
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	// Assert - no panic should occur, and cache should be in valid state
	stats := cache.GetStats()
	if stats == nil {
		t.Error("Expected stats to be returned")
	}
}

// TestChangeType_Constants tests change type constants
func TestChangeType_Constants(t *testing.T) {
	// Assert
	if ChangeTypeTest != 0 {
		t.Errorf("Expected ChangeTypeTest to be 0, got %d", ChangeTypeTest)
	}
	if ChangeTypeSource != 1 {
		t.Errorf("Expected ChangeTypeSource to be 1, got %d", ChangeTypeSource)
	}
	if ChangeTypeConfig != 2 {
		t.Errorf("Expected ChangeTypeConfig to be 2, got %d", ChangeTypeConfig)
	}
	if ChangeTypeDependency != 3 {
		t.Errorf("Expected ChangeTypeDependency to be 3, got %d", ChangeTypeDependency)
	}
}

// TestFileChange_StructFields tests FileChange struct fields
func TestFileChange_StructFields(t *testing.T) {
	// Arrange
	change := &FileChange{
		Path:          "test.go",
		Type:          ChangeTypeTest,
		IsNew:         true,
		Hash:          "abc123",
		Timestamp:     time.Now(),
		AffectedTests: []string{"test1", "test2"},
	}

	// Assert
	if change.Path != "test.go" {
		t.Errorf("Expected Path to be 'test.go', got '%s'", change.Path)
	}
	if change.Type != ChangeTypeTest {
		t.Errorf("Expected Type to be ChangeTypeTest, got %v", change.Type)
	}
	if !change.IsNew {
		t.Error("Expected IsNew to be true")
	}
	if change.Hash != "abc123" {
		t.Errorf("Expected Hash to be 'abc123', got '%s'", change.Hash)
	}
	if len(change.AffectedTests) != 2 {
		t.Errorf("Expected 2 affected tests, got %d", len(change.AffectedTests))
	}
}

// TestCachedTestResult_StructFields tests CachedTestResult struct fields
func TestCachedTestResult_StructFields(t *testing.T) {
	// Arrange
	suite := &models.TestSuite{
		FilePath: "test.go",
		Duration: time.Millisecond,
	}
	result := &CachedTestResult{
		Suite:     suite,
		FileHash:  "abc123",
		LastRun:   time.Now(),
		Duration:  100 * time.Millisecond,
		Status:    models.StatusPassed,
		DependsOn: []string{"dep1.go", "dep2.go"},
	}

	// Assert
	if result.Suite != suite {
		t.Error("Expected Suite to match")
	}
	if result.FileHash != "abc123" {
		t.Errorf("Expected FileHash to be 'abc123', got '%s'", result.FileHash)
	}
	if result.Duration != 100*time.Millisecond {
		t.Errorf("Expected Duration to be 100ms, got %v", result.Duration)
	}
	if result.Status != models.StatusPassed {
		t.Errorf("Expected Status to be StatusPassed, got %v", result.Status)
	}
	if len(result.DependsOn) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(result.DependsOn))
	}
}

// TestShouldRunTests_TableDriven covers all edge cases for ShouldRunTests
func TestShouldRunTests_TableDriven(t *testing.T) {
	tempDir := t.TempDir()
	cache := NewTestResultCache()

	suite := &models.TestSuite{FilePath: filepath.Join(tempDir, "example_test.go"), Duration: time.Millisecond}
	cache.CacheResult("pkg/example", suite)

	tests := []struct {
		name    string
		changes []*FileChange
		wantRun bool
		wantLen int
	}{
		{
			name:    "no changes",
			changes: nil,
			wantRun: false,
			wantLen: 0,
		},
		{
			name:    "all changes not new, no stale tests",
			changes: []*FileChange{{Path: "sentinel.config.json", Type: ChangeTypeConfig, IsNew: false}},
			wantRun: false,
			wantLen: 0,
		},
		{
			name:    "all changes not new, with stale tests",
			changes: []*FileChange{{Path: "pkg/example/example_test.go", Type: ChangeTypeTest, IsNew: false}},
			wantRun: true,
			wantLen: 1,
		},
		{
			name:    "all changes new",
			changes: []*FileChange{{Path: "foo.go", Type: ChangeTypeSource, IsNew: true}},
			wantRun: true,
			wantLen: 1,
		},
		{
			name:    "mixed changes, one new",
			changes: []*FileChange{{Path: "foo.go", Type: ChangeTypeSource, IsNew: false}, {Path: "bar.go", Type: ChangeTypeSource, IsNew: true}},
			wantRun: true,
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "all changes not new, no stale tests" {
				cache.Clear()
			}
			shouldRun, testPaths := cache.ShouldRunTests(tt.changes)
			if shouldRun != tt.wantRun {
				t.Errorf("shouldRun: got %v, want %v", shouldRun, tt.wantRun)
			}
			if tt.wantLen == 0 && testPaths != nil && len(testPaths) != 0 {
				t.Errorf("expected no testPaths, got %v", testPaths)
			}
			if tt.wantLen > 0 && len(testPaths) != tt.wantLen {
				t.Errorf("expected %d testPaths, got %d", tt.wantLen, len(testPaths))
			}
			// For the "all changes not new, no stale tests" case, expect testPaths to be [.] if shouldRun is true
			if tt.name == "all changes not new, no stale tests" && shouldRun && len(testPaths) == 1 && testPaths[0] == "." {
				// Acceptable on some platforms
			}
		})
	}
}

// TestGetStaleTests_TableDriven covers all edge cases for GetStaleTests
func TestGetStaleTests_TableDriven(t *testing.T) {
	cache := NewTestResultCache()
	suite := &models.TestSuite{FilePath: "pkg/example/example_test.go", Duration: time.Millisecond}
	cache.CacheResult("pkg/example", suite)
	cache.CacheResult("pkg/other", suite)

	tests := []struct {
		name     string
		changes  []*FileChange
		wantLen  int
		wantPath string
	}{
		{
			name:    "empty changes",
			changes: nil,
			wantLen: 0,
		},
		{
			name:     "test file change",
			changes:  []*FileChange{{Path: "pkg/example/example_test.go", Type: ChangeTypeTest}},
			wantLen:  1,
			wantPath: "pkg/example",
		},
		{
			name:     "source file change",
			changes:  []*FileChange{{Path: "pkg/example/example.go", Type: ChangeTypeSource}},
			wantLen:  1,
			wantPath: "pkg/example",
		},
		{
			name:    "config change with cached results",
			changes: []*FileChange{{Path: "sentinel.config.json", Type: ChangeTypeConfig}},
			wantLen: 2,
		},
		{
			name:    "config change with no cached results",
			changes: []*FileChange{{Path: "sentinel.config.json", Type: ChangeTypeConfig}},
			wantLen: 0,
		},
		{
			name:     "dependency change with affected tests",
			changes:  []*FileChange{{Path: "go.mod", Type: ChangeTypeDependency, AffectedTests: []string{"pkg/example"}}},
			wantLen:  1,
			wantPath: "pkg/example",
		},
		{
			name:    "dependency change with no affected tests",
			changes: []*FileChange{{Path: "go.mod", Type: ChangeTypeDependency, AffectedTests: nil}},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "config change with no cached results" {
				cache.Clear()
			}
			staleTests := cache.GetStaleTests(tt.changes)
			if len(staleTests) != tt.wantLen {
				t.Errorf("expected %d stale tests, got %d", tt.wantLen, len(staleTests))
			}
			if tt.wantPath != "" && tt.wantLen > 0 {
				actual := staleTests[0]
				if filepath.ToSlash(actual) != tt.wantPath {
					t.Errorf("expected path %q, got %q", tt.wantPath, actual)
				}
			}
		})
	}
}

// TestGetCachedResult_EdgeCases covers edge cases for GetCachedResult
func TestGetCachedResult_EdgeCases(t *testing.T) {
	cache := NewTestResultCache()
	testPath := "pkg/example"
	suite := &models.TestSuite{FilePath: "pkg/example/example_test.go", Duration: time.Millisecond}
	cache.CacheResult(testPath, suite)

	// No dependencies changed
	result, exists := cache.GetCachedResult(testPath)
	if !exists || result == nil {
		t.Error("Expected cached result to exist and be valid")
	}

	// Add a dependency and mark it as changed
	cache.mutex.Lock()
	cache.results[testPath].DependsOn = []string{"dep.go"}
	cache.mutex.Unlock()
	cache.MarkFileAsProcessed("dep.go", time.Now().Add(time.Hour))
	result, exists = cache.GetCachedResult(testPath)
	if exists || result != nil {
		t.Error("Expected cached result to be invalidated by changed dependency")
	}

	// No dependencies
	cache.CacheResult("noDeps", suite)
	cache.mutex.Lock()
	cache.results["noDeps"].DependsOn = nil
	cache.mutex.Unlock()
	result, exists = cache.GetCachedResult("noDeps")
	if !exists || result == nil {
		t.Error("Expected cached result to exist with no dependencies")
	}

	// Multiple dependencies, one changed
	cache.CacheResult("multiDeps", suite)
	cache.mutex.Lock()
	cache.results["multiDeps"].DependsOn = []string{"a.go", "b.go"}
	cache.mutex.Unlock()
	cache.MarkFileAsProcessed("a.go", time.Now().Add(time.Hour))
	result, exists = cache.GetCachedResult("multiDeps")
	if exists || result != nil {
		t.Error("Expected cached result to be invalidated if any dependency changed")
	}
}

// TestCalculateFileHash_EdgeCases covers edge cases for calculateFileHash
func TestCalculateFileHash_EdgeCases(t *testing.T) {
	cache := NewTestResultCache()
	tempDir := t.TempDir()
	file := filepath.Join(tempDir, "file.go")
	os.WriteFile(file, []byte("hello world"), 0644)
	hash, err := cache.calculateFileHash(file)
	if err != nil || hash == "" {
		t.Error("Expected hash for existing file")
	}
	// Nonexistent file
	hash, err = cache.calculateFileHash(filepath.Join(tempDir, "nope.go"))
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
	// Empty file
	emptyFile := filepath.Join(tempDir, "empty.go")
	os.WriteFile(emptyFile, []byte(""), 0644)
	hash, err = cache.calculateFileHash(emptyFile)
	if err != nil || hash == "" {
		t.Error("Expected hash for empty file")
	}
}

// TestDetermineChangeType_AllCases covers all file type cases
func TestDetermineChangeType_AllCases(t *testing.T) {
	cache := NewTestResultCache()
	tests := []struct {
		file         string
		typeExpected ChangeType
	}{
		{"go.mod", ChangeTypeDependency},
		{"go.sum", ChangeTypeDependency},
		{"sentinel.config.json", ChangeTypeConfig},
		{".golangci.yml", ChangeTypeConfig},
		{"foo_test.go", ChangeTypeTest},
		{"foo.go", ChangeTypeSource},
		{"main.go", ChangeTypeConfig},
		{"unknown.txt", ChangeTypeConfig},
	}
	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			typeGot := cache.determineChangeType(tt.file)
			if typeGot != tt.typeExpected {
				t.Errorf("expected %v, got %v", tt.typeExpected, typeGot)
			}
		})
	}
}

// TestFindAffectedTests_AllCases covers all change types for findAffectedTests
func TestFindAffectedTests_AllCases(t *testing.T) {
	cache := NewTestResultCache()
	cache.CacheResult("pkg/example", &models.TestSuite{FilePath: "pkg/example/example_test.go"})
	cache.mutex.Lock()
	cache.results["pkg/example"].DependsOn = []string{"dep.go"}
	cache.mutex.Unlock()

	tests := []struct {
		name       string
		file       string
		changeType ChangeType
		expected   []string
	}{
		{"test file", "pkg/example/example_test.go", ChangeTypeTest, []string{"pkg/example"}},
		{"source file", "pkg/example/example.go", ChangeTypeSource, []string{"pkg/example"}},
		{"dependency match", "dep.go", ChangeTypeDependency, []string{"pkg/example"}},
		{"dependency no match", "other.go", ChangeTypeDependency, []string{}},
		{"config", "sentinel.config.json", ChangeTypeConfig, []string{"pkg/example"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cache.findAffectedTests(tt.file, tt.changeType)
			if len(got) != len(tt.expected) {
				t.Errorf("expected %d, got %d", len(tt.expected), len(got))
			}
			for i := range got {
				if filepath.ToSlash(got[i]) != tt.expected[i] {
					t.Errorf("expected %q, got %q", tt.expected[i], got[i])
				}
			}
		})
	}
}

// TestFindDependencies_EdgeCases covers edge cases for findDependencies
func TestFindDependencies_EdgeCases(t *testing.T) {
	cache := NewTestResultCache()
	tempDir := t.TempDir()
	os.WriteFile(filepath.Join(tempDir, "foo.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tempDir, "bar.go"), []byte("package main"), 0644)
	os.WriteFile("go.mod", []byte("module test"), 0644)
	os.WriteFile("go.sum", []byte("sum"), 0644)
	deps := cache.findDependencies(tempDir)
	if len(deps) < 2 {
		t.Error("Expected at least go.mod and go.sum as dependencies")
	}
	// Remove go.mod and go.sum
	os.Remove("go.mod")
	os.Remove("go.sum")
	deps = cache.findDependencies(tempDir)
	for _, dep := range deps {
		if dep == "go.mod" || dep == "go.sum" {
			t.Error("Did not expect go.mod or go.sum as dependencies after removal")
		}
	}
	// testPath is a file
	file := filepath.Join(tempDir, "foo.go")
	deps = cache.findDependencies(file)
	if len(deps) == 0 {
		t.Error("Expected at least one dependency for file path")
	}
}

// TestCalculateSuiteStatus_AllCases covers all status combinations
func TestCalculateSuiteStatus_AllCases(t *testing.T) {
	cache := NewTestResultCache()
	tests := []struct {
		name     string
		suite    *models.TestSuite
		expected models.TestStatus
	}{
		{"failed", &models.TestSuite{FailedCount: 1}, models.StatusFailed},
		{"skipped", &models.TestSuite{SkippedCount: 1, PassedCount: 0}, models.StatusSkipped},
		{"passed", &models.TestSuite{PassedCount: 1}, models.StatusPassed},
		{"all zero", &models.TestSuite{}, models.StatusPassed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cache.calculateSuiteStatus(tt.suite)
			if got != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}
