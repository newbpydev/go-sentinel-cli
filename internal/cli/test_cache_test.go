package cli

import (
	"os"
	"path/filepath"
	"testing"
	"time"
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
		t.Error("Expected testPaths to be provided with new changes")
	}
}

// TestGetStaleTests_TestFileChange tests stale test detection for test file changes
func TestGetStaleTests_TestFileChange(t *testing.T) {
	// Arrange
	cache := NewTestResultCache()
	changes := []*FileChange{
		{
			Path: "pkg/example_test.go",
			Type: ChangeTypeTest,
		},
	}

	// Act
	staleTests := cache.GetStaleTests(changes)

	// Assert
	if len(staleTests) != 1 {
		t.Errorf("Expected 1 stale test, got %d", len(staleTests))
	}
	if staleTests[0] != "pkg" {
		t.Errorf("Expected stale test 'pkg', got '%s'", staleTests[0])
	}
}

// TestGetStaleTests_ConfigChange tests stale test detection for config changes
func TestGetStaleTests_ConfigChange(t *testing.T) {
	// Arrange
	cache := NewTestResultCache()

	// Add some cached results first
	cache.results["pkg1"] = &CachedTestResult{}
	cache.results["pkg2"] = &CachedTestResult{}

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
		t.Errorf("Expected 2 stale tests for config change, got %d", len(staleTests))
	}
}

// TestCacheResult_StoresResult tests storing test results in cache
func TestCacheResult_StoresResult(t *testing.T) {
	// Arrange
	cache := NewTestResultCache()
	testPath := "pkg/test"
	suite := &TestSuite{
		FilePath:     "pkg/test",
		Duration:     100 * time.Millisecond,
		PassedCount:  5,
		FailedCount:  0,
		SkippedCount: 1,
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
	if cached == nil {
		t.Fatal("Expected cached result to be non-nil")
	}
	if cached.Suite != suite {
		t.Error("Expected suite to be stored correctly")
	}
	if cached.Duration != suite.Duration {
		t.Errorf("Expected duration %v, got %v", suite.Duration, cached.Duration)
	}
	if cached.Status != StatusPassed {
		t.Errorf("Expected status StatusPassed, got %v", cached.Status)
	}
}

// TestGetCachedResult_ValidResult tests retrieving valid cached results
func TestGetCachedResult_ValidResult(t *testing.T) {
	// Arrange
	cache := NewTestResultCache()
	testPath := "pkg/test"

	// Store a result first
	suite := &TestSuite{FilePath: testPath}
	cache.CacheResult(testPath, suite)

	// Act
	cached, valid := cache.GetCachedResult(testPath)

	// Assert
	if !valid {
		t.Error("Expected cached result to be valid")
	}
	if cached == nil {
		t.Error("Expected cached result to be returned")
		return // Exit early to avoid nil pointer dereference
	}
	if cached.Suite != suite {
		t.Error("Expected correct suite to be returned")
	}
}

// TestGetCachedResult_NonexistentResult tests retrieving nonexistent results
func TestGetCachedResult_NonexistentResult(t *testing.T) {
	// Arrange
	cache := NewTestResultCache()
	testPath := "nonexistent"

	// Act
	cached, valid := cache.GetCachedResult(testPath)

	// Assert
	if valid {
		t.Error("Expected result to be invalid for nonexistent path")
	}
	if cached != nil {
		t.Error("Expected cached result to be nil for nonexistent path")
	}
}

// TestGetCachedResult_InvalidatedByDependency tests cache invalidation by dependencies
func TestGetCachedResult_InvalidatedByDependency(t *testing.T) {
	// Arrange
	cache := NewTestResultCache()
	testPath := "pkg/test"
	dependencyPath := "pkg/source.go"

	// Store a result with dependencies
	suite := &TestSuite{FilePath: testPath}
	cache.CacheResult(testPath, suite)

	// Mark dependency as changed after cache
	time.Sleep(1 * time.Millisecond) // Ensure time difference
	cache.MarkFileAsProcessed(dependencyPath, time.Now())

	// Update the cached result to include this dependency
	cache.mutex.Lock()
	if cached, exists := cache.results[testPath]; exists {
		cached.DependsOn = []string{dependencyPath}
	}
	cache.mutex.Unlock()

	// Act
	cached, valid := cache.GetCachedResult(testPath)

	// Assert
	if valid {
		t.Error("Expected result to be invalid due to dependency change")
	}
	if cached != nil {
		t.Error("Expected cached result to be nil due to dependency change")
	}
}

// TestClear_RemovesAllData tests clearing all cache data
func TestClear_RemovesAllData(t *testing.T) {
	// Arrange
	cache := NewTestResultCache()

	// Add some data
	cache.results["test1"] = &CachedTestResult{}
	cache.fileTimes["file1.go"] = time.Now()
	cache.testTimes["test1"] = time.Now()

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

// TestGetStats_ReturnsCorrectStats tests cache statistics
func TestGetStats_ReturnsCorrectStats(t *testing.T) {
	// Arrange
	cache := NewTestResultCache()

	// Add some data
	cache.results["test1"] = &CachedTestResult{}
	cache.results["test2"] = &CachedTestResult{}
	cache.fileTimes["file1.go"] = time.Now()
	cache.testTimes["test1"] = time.Now()

	// Act
	stats := cache.GetStats()

	// Assert
	if stats == nil {
		t.Fatal("Expected stats to be returned")
	}
	if stats["cached_results"].(int) != 2 {
		t.Errorf("Expected 2 cached results, got %v", stats["cached_results"])
	}
	if stats["tracked_files"].(int) != 1 {
		t.Errorf("Expected 1 tracked file, got %v", stats["tracked_files"])
	}
	if stats["tracked_tests"].(int) != 1 {
		t.Errorf("Expected 1 tracked test, got %v", stats["tracked_tests"])
	}
}

// TestConcurrentAccess_SafeAccess tests concurrent access to cache
func TestConcurrentAccess_SafeAccess(t *testing.T) {
	// Arrange
	cache := NewTestResultCache()
	done := make(chan bool, 3)

	// Act - Concurrent operations
	// Goroutine 1: Store results
	go func() {
		for i := 0; i < 10; i++ {
			suite := &TestSuite{FilePath: "test"}
			cache.CacheResult("test", suite)
		}
		done <- true
	}()

	// Goroutine 2: Read results
	go func() {
		for i := 0; i < 10; i++ {
			cache.GetCachedResult("test")
		}
		done <- true
	}()

	// Goroutine 3: Clear cache
	go func() {
		cache.Clear()
		done <- true
	}()

	// Assert - Wait for all operations to complete without panic
	for i := 0; i < 3; i++ {
		select {
		case <-done:
		case <-time.After(1 * time.Second):
			t.Error("Timeout waiting for concurrent operations")
		}
	}
}

// TestChangeType_Constants tests that change type constants are properly defined
func TestChangeType_Constants(t *testing.T) {
	// Test that all constants have expected values
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

// TestFileChange_StructFields tests FileChange struct field access
func TestFileChange_StructFields(t *testing.T) {
	// Arrange
	change := &FileChange{
		Path:          "test.go",
		Type:          ChangeTypeTest,
		IsNew:         true,
		Hash:          "abc123",
		Timestamp:     time.Now(),
		AffectedTests: []string{"pkg1", "pkg2"},
	}

	// Assert
	if change.Path != "test.go" {
		t.Errorf("Expected Path 'test.go', got '%s'", change.Path)
	}
	if change.Type != ChangeTypeTest {
		t.Errorf("Expected Type ChangeTypeTest, got %v", change.Type)
	}
	if !change.IsNew {
		t.Error("Expected IsNew to be true")
	}
	if change.Hash != "abc123" {
		t.Errorf("Expected Hash 'abc123', got '%s'", change.Hash)
	}
	if len(change.AffectedTests) != 2 {
		t.Errorf("Expected 2 affected tests, got %d", len(change.AffectedTests))
	}
}

// TestCachedTestResult_StructFields tests CachedTestResult struct field access
func TestCachedTestResult_StructFields(t *testing.T) {
	// Arrange
	suite := &TestSuite{FilePath: "test.go"}
	lastRun := time.Now()
	duration := 100 * time.Millisecond

	cached := &CachedTestResult{
		Suite:     suite,
		FileHash:  "hash123",
		LastRun:   lastRun,
		Duration:  duration,
		Status:    StatusPassed,
		DependsOn: []string{"dep1.go", "dep2.go"},
	}

	// Assert
	if cached.Suite != suite {
		t.Error("Expected Suite to be set correctly")
	}
	if cached.FileHash != "hash123" {
		t.Errorf("Expected FileHash 'hash123', got '%s'", cached.FileHash)
	}
	if !cached.LastRun.Equal(lastRun) {
		t.Errorf("Expected LastRun %v, got %v", lastRun, cached.LastRun)
	}
	if cached.Duration != duration {
		t.Errorf("Expected Duration %v, got %v", duration, cached.Duration)
	}
	if cached.Status != StatusPassed {
		t.Errorf("Expected Status StatusPassed, got %v", cached.Status)
	}
	if len(cached.DependsOn) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(cached.DependsOn))
	}
}
