package execution

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/newbpydev/go-sentinel/internal/cli/core"
)

// InMemoryCacheManager provides an in-memory implementation of cache management
type InMemoryCacheManager struct {
	results      map[string]*core.CachedResult
	dependencies map[string][]string
	fileTimes    map[string]time.Time
	maxSize      int
	mu           sync.RWMutex
}

// NewInMemoryCacheManager creates a new in-memory cache manager
func NewInMemoryCacheManager(maxSize int) *InMemoryCacheManager {
	if maxSize <= 0 {
		maxSize = 1000 // Default size
	}

	return &InMemoryCacheManager{
		results:      make(map[string]*core.CachedResult),
		dependencies: make(map[string][]string),
		fileTimes:    make(map[string]time.Time),
		maxSize:      maxSize,
		mu:           sync.RWMutex{},
	}
}

// GetCachedResult retrieves a cached test result if available and valid
func (c *InMemoryCacheManager) GetCachedResult(target core.TestTarget) (*core.CachedResult, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := c.generateKey(target)
	cached, exists := c.results[key]
	if !exists {
		return nil, false
	}

	// Check if the cached result is still valid
	if !c.isResultValid(cached) {
		return nil, false
	}

	return cached, true
}

// StoreResult caches a test result
func (c *InMemoryCacheManager) StoreResult(target core.TestTarget, result *core.TestResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Enforce size limit by removing oldest entries
	c.enforceSize()

	key := c.generateKey(target)
	dependencies := c.findDependencies(target)

	// Calculate hash of dependencies
	depHash := c.calculateDependencyHash(dependencies)

	cached := &core.CachedResult{
		Result:       result,
		CacheTime:    time.Now(),
		Dependencies: dependencies,
		Hash:         depHash,
		IsValid:      true,
	}

	c.results[key] = cached
	c.dependencies[key] = dependencies

	// Update file times for dependencies
	for _, dep := range dependencies {
		if info, err := os.Stat(dep); err == nil {
			c.fileTimes[dep] = info.ModTime()
		}
	}
}

// InvalidateCache invalidates cache entries based on file changes
func (c *InMemoryCacheManager) InvalidateCache(changes []core.FileChange) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create a map of changed files for faster lookup
	changedFiles := make(map[string]bool)
	for _, change := range changes {
		changedFiles[change.Path] = true
	}

	// Check each cached result for dependency conflicts
	for _, cached := range c.results {
		shouldInvalidate := false

		// Check if any dependencies have changed
		for _, dep := range cached.Dependencies {
			if changedFiles[dep] {
				shouldInvalidate = true
				break
			}

			// Also check if file modification time changed
			if info, err := os.Stat(dep); err == nil {
				if lastTime, exists := c.fileTimes[dep]; exists {
					if info.ModTime().After(lastTime) {
						shouldInvalidate = true
						break
					}
				}
			}
		}

		if shouldInvalidate {
			cached.IsValid = false
		}
	}
}

// Clear removes all cached results
func (c *InMemoryCacheManager) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.results = make(map[string]*core.CachedResult)
	c.dependencies = make(map[string][]string)
	c.fileTimes = make(map[string]time.Time)
}

// GetStats returns cache statistics
func (c *InMemoryCacheManager) GetStats() core.CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	validEntries := 0
	invalidEntries := 0
	var oldestEntry, newestEntry time.Time

	first := true
	for _, cached := range c.results {
		if cached.IsValid {
			validEntries++
		} else {
			invalidEntries++
		}

		if first {
			oldestEntry = cached.CacheTime
			newestEntry = cached.CacheTime
			first = false
		} else {
			if cached.CacheTime.Before(oldestEntry) {
				oldestEntry = cached.CacheTime
			}
			if cached.CacheTime.After(newestEntry) {
				newestEntry = cached.CacheTime
			}
		}
	}

	hitRate := 0.0
	totalEntries := validEntries + invalidEntries
	if totalEntries > 0 {
		hitRate = float64(validEntries) / float64(totalEntries) * 100
	}

	return core.CacheStats{
		TotalEntries:   totalEntries,
		ValidEntries:   validEntries,
		InvalidEntries: invalidEntries,
		HitRate:        hitRate,
		MemoryUsage:    int64(len(c.results) * 1024), // Rough estimate
		OldestEntry:    oldestEntry,
		NewestEntry:    newestEntry,
	}
}

// generateKey creates a unique key for a test target
func (c *InMemoryCacheManager) generateKey(target core.TestTarget) string {
	// Create a key based on path and type
	key := fmt.Sprintf("%s:%s", target.Type, target.Path)

	// Include functions if specified
	if len(target.Functions) > 0 {
		for _, fn := range target.Functions {
			key += ":" + fn
		}
	}

	return key
}

// isResultValid checks if a cached result is still valid
func (c *InMemoryCacheManager) isResultValid(cached *core.CachedResult) bool {
	if !cached.IsValid {
		return false
	}

	// Check if dependencies have changed since cache time
	for _, dep := range cached.Dependencies {
		if info, err := os.Stat(dep); err == nil {
			if info.ModTime().After(cached.CacheTime) {
				return false
			}
		} else {
			// File doesn't exist anymore
			return false
		}
	}

	return true
}

// findDependencies discovers files that a test target depends on
func (c *InMemoryCacheManager) findDependencies(target core.TestTarget) []string {
	var dependencies []string

	// Always include go.mod and go.sum as dependencies
	if _, err := os.Stat("go.mod"); err == nil {
		dependencies = append(dependencies, "go.mod")
	}
	if _, err := os.Stat("go.sum"); err == nil {
		dependencies = append(dependencies, "go.sum")
	}

	switch target.Type {
	case "package":
		// For package targets, include all .go files in the package
		if target.Path != "./..." {
			pattern := filepath.Join(target.Path, "*.go")
			if matches, err := filepath.Glob(pattern); err == nil {
				dependencies = append(dependencies, matches...)
			}
		}

	case "file":
		// For file targets, include the specific file and related files
		dependencies = append(dependencies, target.Path)

		// Include other files in the same package
		dir := filepath.Dir(target.Path)
		pattern := filepath.Join(dir, "*.go")
		if matches, err := filepath.Glob(pattern); err == nil {
			for _, match := range matches {
				if match != target.Path {
					dependencies = append(dependencies, match)
				}
			}
		}

	case "recursive":
		// For recursive targets, this is harder to track efficiently
		// We'll just include the basic dependencies for now
		// In a real implementation, we might scan the entire tree
		break
	}

	return dependencies
}

// calculateDependencyHash creates a hash of dependency contents
func (c *InMemoryCacheManager) calculateDependencyHash(dependencies []string) string {
	hash := md5.New()

	for _, dep := range dependencies {
		// Include file path and modification time in hash
		if info, err := os.Stat(dep); err == nil {
			hash.Write([]byte(dep))
			hash.Write([]byte(info.ModTime().String()))
		}
	}

	return fmt.Sprintf("%x", hash.Sum(nil))
}

// enforceSize removes oldest entries if cache is too large
func (c *InMemoryCacheManager) enforceSize() {
	if len(c.results) < c.maxSize {
		return
	}

	// Find oldest entries to remove
	type entry struct {
		key  string
		time time.Time
	}

	entries := make([]entry, 0, len(c.results))
	for key, cached := range c.results {
		entries = append(entries, entry{key: key, time: cached.CacheTime})
	}

	// Sort by time (oldest first)
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].time.After(entries[j].time) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// Remove oldest 25% of entries
	removeCount := c.maxSize / 4
	for i := 0; i < removeCount && i < len(entries); i++ {
		key := entries[i].key
		delete(c.results, key)
		delete(c.dependencies, key)
	}
}

// FileBasedCacheManager provides a file-based cache implementation
type FileBasedCacheManager struct {
	*InMemoryCacheManager
	cacheDir string
}

// NewFileBasedCacheManager creates a new file-based cache manager
func NewFileBasedCacheManager(cacheDir string, maxSize int) (*FileBasedCacheManager, error) {
	// Ensure cache directory exists
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, core.NewCacheError("create", cacheDir, "failed to create cache directory", err)
	}

	return &FileBasedCacheManager{
		InMemoryCacheManager: NewInMemoryCacheManager(maxSize),
		cacheDir:             cacheDir,
	}, nil
}

// LoadFromDisk loads cache from disk (placeholder for future implementation)
func (c *FileBasedCacheManager) LoadFromDisk() error {
	// TODO: Implement persistent cache loading
	return nil
}

// SaveToDisk saves cache to disk (placeholder for future implementation)
func (c *FileBasedCacheManager) SaveToDisk() error {
	// TODO: Implement persistent cache saving
	return nil
}
