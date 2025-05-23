package cli

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TestResultCache manages cached test results for incremental testing
type TestResultCache struct {
	results   map[string]*CachedTestResult
	fileTimes map[string]time.Time
	testTimes map[string]time.Time
	mutex     sync.RWMutex
}

// CachedTestResult represents a cached test result
type CachedTestResult struct {
	Suite     *TestSuite
	FileHash  string
	LastRun   time.Time
	Duration  time.Duration
	Status    TestStatus
	DependsOn []string // Files this test depends on
}

// ChangeType represents the type of file change
type ChangeType int

const (
	ChangeTypeTest ChangeType = iota
	ChangeTypeSource
	ChangeTypeConfig
	ChangeTypeDependency
)

// FileChange represents a file change with analysis
type FileChange struct {
	Path          string
	Type          ChangeType
	IsNew         bool
	Hash          string
	Timestamp     time.Time
	AffectedTests []string
}

// NewTestResultCache creates a new test result cache
func NewTestResultCache() *TestResultCache {
	return &TestResultCache{
		results:   make(map[string]*CachedTestResult),
		fileTimes: make(map[string]time.Time),
		testTimes: make(map[string]time.Time),
		mutex:     sync.RWMutex{},
	}
}

// AnalyzeChange analyzes a file change and determines its impact
func (c *TestResultCache) AnalyzeChange(filePath string) (*FileChange, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Get file info
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file %s: %w", filePath, err)
	}

	// Calculate file hash
	hash, err := c.calculateFileHash(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate hash for %s: %w", filePath, err)
	}

	// Determine change type
	changeType := c.determineChangeType(filePath)

	// Check if file is new
	lastTime, exists := c.fileTimes[filePath]
	isNew := !exists || info.ModTime().After(lastTime)

	// Update file time
	c.fileTimes[filePath] = info.ModTime()

	change := &FileChange{
		Path:      filePath,
		Type:      changeType,
		IsNew:     isNew,
		Hash:      hash,
		Timestamp: info.ModTime(),
	}

	// Determine affected tests
	change.AffectedTests = c.findAffectedTests(filePath, changeType)

	return change, nil
}

// GetStaleTests returns tests that need to be re-run based on changes
func (c *TestResultCache) GetStaleTests(changes []*FileChange) []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	staleTests := make(map[string]bool)

	for _, change := range changes {
		switch change.Type {
		case ChangeTypeTest:
			// Test file changed - only run this specific test
			testPackage := filepath.Dir(change.Path)
			staleTests[testPackage] = true

		case ChangeTypeSource:
			// Source file changed - run tests in same package
			packageDir := filepath.Dir(change.Path)
			staleTests[packageDir] = true

		case ChangeTypeConfig:
			// Config changed - mark all tests as stale
			for testPath := range c.results {
				staleTests[testPath] = true
			}

		case ChangeTypeDependency:
			// Dependency changed - run affected tests
			for _, testPath := range change.AffectedTests {
				staleTests[testPath] = true
			}
		}
	}

	// Convert to slice
	result := make([]string, 0, len(staleTests))
	for testPath := range staleTests {
		result = append(result, testPath)
	}

	return result
}

// CacheResult stores a test result in the cache
func (c *TestResultCache) CacheResult(testPath string, suite *TestSuite) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Calculate dependencies
	dependencies := c.findDependencies(testPath)

	cached := &CachedTestResult{
		Suite:     suite,
		LastRun:   time.Now(),
		Duration:  suite.Duration,
		Status:    c.calculateSuiteStatus(suite),
		DependsOn: dependencies,
	}

	c.results[testPath] = cached
	c.testTimes[testPath] = time.Now()
}

// GetCachedResult retrieves a cached test result if still valid
func (c *TestResultCache) GetCachedResult(testPath string) (*CachedTestResult, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	result, exists := c.results[testPath]
	if !exists {
		return nil, false
	}

	// Check if dependencies have changed
	for _, dep := range result.DependsOn {
		if fileTime, exists := c.fileTimes[dep]; exists {
			if fileTime.After(result.LastRun) {
				return nil, false // Dependencies changed, cache invalid
			}
		}
	}

	return result, true
}

// calculateFileHash calculates MD5 hash of a file
func (c *TestResultCache) calculateFileHash(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", md5.Sum(content)), nil
}

// determineChangeType determines the type of change based on file path
func (c *TestResultCache) determineChangeType(filePath string) ChangeType {
	base := filepath.Base(filePath)
	ext := filepath.Ext(filePath)

	switch {
	case base == "go.mod" || base == "go.sum":
		return ChangeTypeDependency
	case base == "sentinel.config.json" || base == ".golangci.yml":
		return ChangeTypeConfig
	case ext == ".go" && filepath.Base(filePath) != "main.go":
		if isTestFile(filePath) {
			return ChangeTypeTest
		}
		return ChangeTypeSource
	default:
		return ChangeTypeConfig // Default to config for unknown files
	}
}

// findAffectedTests finds tests that might be affected by a file change
func (c *TestResultCache) findAffectedTests(filePath string, changeType ChangeType) []string {
	var affected []string

	switch changeType {
	case ChangeTypeTest:
		// Only affects the test itself
		affected = append(affected, filepath.Dir(filePath))

	case ChangeTypeSource:
		// Affects tests in the same package
		packageDir := filepath.Dir(filePath)
		affected = append(affected, packageDir)

	case ChangeTypeDependency:
		// Affects all tests that might import this dependency
		for testPath, cached := range c.results {
			for _, dep := range cached.DependsOn {
				if dep == filePath {
					affected = append(affected, testPath)
					break
				}
			}
		}

	case ChangeTypeConfig:
		// Affects all tests
		for testPath := range c.results {
			affected = append(affected, testPath)
		}
	}

	return affected
}

// findDependencies finds files that a test depends on
func (c *TestResultCache) findDependencies(testPath string) []string {
	var dependencies []string

	// Add go.mod and go.sum as dependencies
	if _, err := os.Stat("go.mod"); err == nil {
		dependencies = append(dependencies, "go.mod")
	}
	if _, err := os.Stat("go.sum"); err == nil {
		dependencies = append(dependencies, "go.sum")
	}

	// Add source files in the same package
	err := filepath.Walk(testPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if filepath.Ext(path) == ".go" && !isTestFile(path) {
			dependencies = append(dependencies, path)
		}

		return nil
	})

	if err != nil {
		// If walk fails, just return basic dependencies
	}

	return dependencies
}

// calculateSuiteStatus calculates overall status of a test suite
func (c *TestResultCache) calculateSuiteStatus(suite *TestSuite) TestStatus {
	if suite.FailedCount > 0 {
		return StatusFailed
	}
	if suite.SkippedCount > 0 && suite.PassedCount == 0 {
		return StatusSkipped
	}
	return StatusPassed
}

// Clear clears all cached results
func (c *TestResultCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.results = make(map[string]*CachedTestResult)
	c.fileTimes = make(map[string]time.Time)
	c.testTimes = make(map[string]time.Time)
}

// GetStats returns cache statistics
func (c *TestResultCache) GetStats() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return map[string]interface{}{
		"cached_results": len(c.results),
		"tracked_files":  len(c.fileTimes),
		"tracked_tests":  len(c.testTimes),
	}
}
