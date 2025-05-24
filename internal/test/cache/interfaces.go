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
	GetStats() *Stats

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

// Storage provides pluggable storage backends for cache data
type Storage interface {
	// Store saves data to the storage backend
	Store(key string, data []byte) error

	// Load retrieves data from the storage backend
	Load(key string) ([]byte, error)

	// Delete removes data from the storage backend
	Delete(key string) error

	// Exists checks if a key exists in the storage backend
	Exists(key string) bool

	// Clear removes all data from the storage backend
	Clear() error

	// Size returns the total size of stored data
	Size() (int64, error)

	// Keys returns all keys in the storage backend
	Keys() ([]string, error)
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

// Stats provides metrics about cache usage and performance
type Stats struct {
	// HitCount is the number of cache hits
	HitCount int64

	// MissCount is the number of cache misses
	MissCount int64

	// StoreCount is the number of cache stores
	StoreCount int64

	// EvictionCount is the number of cache evictions
	EvictionCount int64

	// TotalSize is the total size of cached data
	TotalSize int64

	// LastAccess is the time of the last cache access
	LastAccess time.Time

	// StartTime is when cache metrics started being collected
	StartTime time.Time
}

// Config configures cache behavior and limits
type Config struct {
	// MaxSize is the maximum cache size in bytes
	MaxSize int64

	// MaxEntries is the maximum number of cache entries
	MaxEntries int

	// TTL is the default time-to-live for cache entries
	TTL time.Duration

	// CleanupInterval is how often to run cache cleanup
	CleanupInterval time.Duration

	// PersistToDisk indicates if cache should be persisted
	PersistToDisk bool

	// StoragePath is the path for persistent storage
	StoragePath string

	// CompressionEnabled indicates if data should be compressed
	CompressionEnabled bool
}

// Key represents a cache key with metadata
type Key struct {
	// Value is the actual key value
	Value string

	// Category is the key category for organization
	Category string

	// Tags are metadata tags for the key
	Tags []string

	// CreatedAt is when the key was created
	CreatedAt time.Time

	// AccessedAt is when the key was last accessed
	AccessedAt time.Time
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
