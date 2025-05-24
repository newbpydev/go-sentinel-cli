// Package cache provides test result caching interfaces and implementations
package cache

import (
	"context"
	"time"
)

// ResultCache handles caching of test results for optimization
type ResultCache interface {
	// Get retrieves a cached result by key
	Get(ctx context.Context, key string) (*CachedResult, error)

	// Set stores a result in the cache
	Set(ctx context.Context, key string, result *CachedResult, ttl time.Duration) error

	// Delete removes a result from the cache
	Delete(ctx context.Context, key string) error

	// Clear clears all cached results
	Clear(ctx context.Context) error

	// GetStats returns cache statistics
	GetStats() *CacheStats

	// Close closes the cache and releases resources
	Close() error
}

// FileHashCache tracks file hashes for change detection
type FileHashCache interface {
	// GetFileHash retrieves the cached hash for a file
	GetFileHash(filePath string) (string, bool)

	// SetFileHash stores the hash for a file
	SetFileHash(filePath string, hash string)

	// RemoveFileHash removes the hash for a file
	RemoveFileHash(filePath string)

	// HasChanged checks if a file has changed since last cached
	HasChanged(filePath string) (bool, error)

	// UpdateFileHash updates the hash for a file with current content
	UpdateFileHash(filePath string) error

	// GetAllHashes returns all cached file hashes
	GetAllHashes() map[string]string

	// Clear clears all cached hashes
	Clear()
}

// DependencyCache tracks dependencies between files and tests
type DependencyCache interface {
	// GetDependencies returns the dependencies for a file
	GetDependencies(filePath string) ([]string, bool)

	// SetDependencies stores the dependencies for a file
	SetDependencies(filePath string, dependencies []string)

	// GetReverseDependencies returns files that depend on the given file
	GetReverseDependencies(filePath string) []string

	// AddDependency adds a dependency relationship
	AddDependency(filePath, dependency string)

	// RemoveDependency removes a dependency relationship
	RemoveDependency(filePath, dependency string)

	// Clear clears all dependency relationships
	Clear()
}

// CacheStorage provides the underlying storage mechanism for cache
type CacheStorage interface {
	// Get retrieves data from storage
	Get(key string) ([]byte, error)

	// Set stores data in storage
	Set(key string, data []byte, ttl time.Duration) error

	// Delete removes data from storage
	Delete(key string) error

	// Keys returns all keys in storage
	Keys() ([]string, error)

	// Clear clears all data from storage
	Clear() error

	// Size returns the current storage size
	Size() (int64, error)

	// Close closes the storage
	Close() error
}

// CachedResult represents a cached test result
type CachedResult struct {
	// Key is the cache key
	Key string

	// TestPackage is the package that was tested
	TestPackage string

	// FileHashes contains hashes of files when test was run
	FileHashes map[string]string

	// Success indicates if the test passed
	Success bool

	// Duration is the test execution time
	Duration time.Duration

	// Output is the test output
	Output string

	// Coverage is the coverage percentage
	Coverage float64

	// CachedAt is when the result was cached
	CachedAt time.Time

	// AccessedAt is when the result was last accessed
	AccessedAt time.Time

	// AccessCount is how many times the result has been accessed
	AccessCount int
}

// CacheStats provides statistics about cache usage
type CacheStats struct {
	// TotalKeys is the total number of cached keys
	TotalKeys int

	// HitCount is the number of cache hits
	HitCount int64

	// MissCount is the number of cache misses
	MissCount int64

	// HitRatio is the cache hit ratio (hits / total requests)
	HitRatio float64

	// MemoryUsage is the memory usage in bytes
	MemoryUsage int64

	// DiskUsage is the disk usage in bytes
	DiskUsage int64

	// OldestEntry is the timestamp of the oldest cached entry
	OldestEntry time.Time

	// NewestEntry is the timestamp of the newest cached entry
	NewestEntry time.Time
}

// CacheConfig configures cache behavior
type CacheConfig struct {
	// MaxMemorySize is the maximum memory size in bytes
	MaxMemorySize int64

	// MaxDiskSize is the maximum disk size in bytes
	MaxDiskSize int64

	// DefaultTTL is the default time-to-live for cached entries
	DefaultTTL time.Duration

	// CleanupInterval is how often to run cache cleanup
	CleanupInterval time.Duration

	// PersistentStorage indicates if cache should persist to disk
	PersistentStorage bool

	// StoragePath is the path for persistent storage
	StoragePath string

	// Compression indicates if cache data should be compressed
	Compression bool
}

// CacheKey represents a cache key with metadata
type CacheKey struct {
	// Package is the test package
	Package string

	// Files are the files involved in the test
	Files []string

	// Checksum is a checksum of the key components
	Checksum string

	// CreatedAt is when the key was created
	CreatedAt time.Time
}

// TestInvalidationReason represents why a cached test was invalidated
type TestInvalidationReason string

const (
	// InvalidationFileChanged indicates a file was modified
	InvalidationFileChanged TestInvalidationReason = "file_changed"

	// InvalidationDependencyChanged indicates a dependency was modified
	InvalidationDependencyChanged TestInvalidationReason = "dependency_changed"

	// InvalidationTTLExpired indicates the cache entry expired
	InvalidationTTLExpired TestInvalidationReason = "ttl_expired"

	// InvalidationManual indicates manual cache invalidation
	InvalidationManual TestInvalidationReason = "manual"

	// InvalidationConfigChanged indicates test configuration changed
	InvalidationConfigChanged TestInvalidationReason = "config_changed"
)
