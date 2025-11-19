// Copyright 2020-2025 Buf Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bufimage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"sync"
	"time"

	"github.com/bufbuild/buf/private/bufpkg/bufmodule"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

// CompilationCache provides caching for compiled file descriptors.
type CompilationCache struct {
	mu      sync.RWMutex
	entries map[string]*cacheEntry
	maxSize int
	ttl     time.Duration
}

type cacheEntry struct {
	result    *descriptorpb.FileDescriptorProto
	hash      string
	timestamp time.Time
}

// NewCompilationCache creates a new CompilationCache.
func NewCompilationCache(options ...CompilationCacheOption) *CompilationCache {
	opts := &compilationCacheOptions{
		maxSize: 10000, // Default max entries
		ttl:     time.Hour,
	}
	for _, opt := range options {
		opt(opts)
	}
	return &CompilationCache{
		entries: make(map[string]*cacheEntry),
		maxSize: opts.maxSize,
		ttl:     opts.ttl,
	}
}

// CompilationCacheOption is an option for CompilationCache.
type CompilationCacheOption func(*compilationCacheOptions)

// WithCompilationCacheMaxSize sets the maximum number of entries in the cache.
func WithCompilationCacheMaxSize(maxSize int) CompilationCacheOption {
	return func(opts *compilationCacheOptions) {
		opts.maxSize = maxSize
	}
}

// WithCompilationCacheTTL sets the time-to-live for cache entries.
func WithCompilationCacheTTL(ttl time.Duration) CompilationCacheOption {
	return func(opts *compilationCacheOptions) {
		opts.ttl = ttl
	}
}

type compilationCacheOptions struct {
	maxSize int
	ttl     time.Duration
}

// Get retrieves a cached result for a file with the given content hash.
func (c *CompilationCache) Get(file string, contentHash string) (*descriptorpb.FileDescriptorProto, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[file]
	if !ok {
		return nil, false
	}

	// Check if entry has expired
	if time.Since(entry.timestamp) > c.ttl {
		return nil, false
	}

	// Check if content hash matches
	if entry.hash != contentHash {
		return nil, false
	}

	return entry.result, true
}

// Put stores a compiled result in the cache.
func (c *CompilationCache) Put(file string, contentHash string, result *descriptorpb.FileDescriptorProto) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict old entries if at capacity
	if len(c.entries) >= c.maxSize {
		c.evictOldest()
	}

	c.entries[file] = &cacheEntry{
		result:    result,
		hash:      contentHash,
		timestamp: time.Now(),
	}
}

// Invalidate removes a file from the cache.
func (c *CompilationCache) Invalidate(file string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, file)
}

// InvalidateAll clears the entire cache.
func (c *CompilationCache) InvalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*cacheEntry)
}

// Size returns the current number of entries in the cache.
func (c *CompilationCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// HitRate returns the current cache hit rate (0.0 to 1.0).
// Note: This is a simplified implementation; production would track hits/misses.
func (c *CompilationCache) HitRate() float64 {
	// Would need to track hits and misses for accurate rate
	return 0.0
}

func (c *CompilationCache) evictOldest() {
	var oldestFile string
	var oldestTime time.Time

	for file, entry := range c.entries {
		if oldestFile == "" || entry.timestamp.Before(oldestTime) {
			oldestFile = file
			oldestTime = entry.timestamp
		}
	}

	if oldestFile != "" {
		delete(c.entries, oldestFile)
	}
}

// CachedCompiler wraps a StreamingCompiler with caching capabilities.
type CachedCompiler struct {
	cache    *CompilationCache
	compiler StreamingCompiler
	logger   *slog.Logger
}

// NewCachedCompiler creates a new CachedCompiler.
func NewCachedCompiler(
	cache *CompilationCache,
	compiler StreamingCompiler,
	logger *slog.Logger,
) *CachedCompiler {
	return &CachedCompiler{
		cache:    cache,
		compiler: compiler,
		logger:   logger,
	}
}

// CompileStream returns results as they become available, using cache when possible.
func (c *CachedCompiler) CompileStream(
	ctx context.Context,
	moduleReadBucket bufmodule.ModuleReadBucket,
) (<-chan CompileResult, error) {
	results := make(chan CompileResult, 100)

	go func() {
		defer close(results)

		// First, check which files are cached
		targetFileInfos, err := bufmodule.GetTargetFileInfos(ctx, moduleReadBucket)
		if err != nil {
			results <- CompileResult{Error: err}
			return
		}

		var toCompile []string
		cachedResults := make([]CompileResult, 0)

		for _, fileInfo := range targetFileInfos {
			path := fileInfo.Path()

			// Compute content hash
			hash, err := c.computeFileHash(ctx, moduleReadBucket, path)
			if err != nil {
				// If we can't compute hash, compile it
				toCompile = append(toCompile, path)
				continue
			}

			// Check cache
			if cached, ok := c.cache.Get(path, hash); ok {
				c.logger.Debug("cache hit", "file", path)
				cachedResults = append(cachedResults, CompileResult{
					File:                path,
					FileDescriptorProto: cached,
				})
			} else {
				c.logger.Debug("cache miss", "file", path)
				toCompile = append(toCompile, path)
			}
		}

		// Send cached results immediately
		for _, result := range cachedResults {
			select {
			case results <- result:
			case <-ctx.Done():
				return
			}
		}

		// Compile uncached files
		if len(toCompile) > 0 {
			stream, err := c.compiler.CompileStream(ctx, moduleReadBucket)
			if err != nil {
				results <- CompileResult{Error: err}
				return
			}

			for result := range stream {
				// Cache successful results
				if result.Error == nil && result.FileDescriptorProto != nil {
					hash, _ := c.computeFileHash(ctx, moduleReadBucket, result.File)
					if hash != "" {
						c.cache.Put(result.File, hash, result.FileDescriptorProto)
					}
				}

				select {
				case results <- result:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return results, nil
}

func (c *CachedCompiler) computeFileHash(
	ctx context.Context,
	moduleReadBucket bufmodule.ModuleReadBucket,
	path string,
) (string, error) {
	file, err := moduleReadBucket.GetFile(ctx, path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Read file content
	content := make([]byte, 0, 4096)
	buf := make([]byte, 4096)
	for {
		n, err := file.Read(buf)
		if n > 0 {
			content = append(content, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	// Compute SHA256 hash
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:]), nil
}

// GetCache returns the underlying CompilationCache.
func (c *CachedCompiler) GetCache() *CompilationCache {
	return c.cache
}

// CacheStats provides statistics about cache usage.
type CacheStats struct {
	// Size is the current number of entries in the cache.
	Size int
	// MaxSize is the maximum number of entries allowed.
	MaxSize int
	// HitRate is the cache hit rate (0.0 to 1.0).
	HitRate float64
}

// Stats returns current cache statistics.
func (c *CachedCompiler) Stats() CacheStats {
	return CacheStats{
		Size:    c.cache.Size(),
		MaxSize: c.cache.maxSize,
		HitRate: c.cache.HitRate(),
	}
}

// HashFileDescriptor computes a hash for a FileDescriptorProto for cache key generation.
func HashFileDescriptor(fdp *descriptorpb.FileDescriptorProto) string {
	data, err := proto.Marshal(fdp)
	if err != nil {
		return ""
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
