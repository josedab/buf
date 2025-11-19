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

// Package bufremotepluginwasm provides WASM plugin caching functionality.
package bufremotepluginwasm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// CompiledModuleCache provides caching for pre-compiled WASM modules.
// It stores compiled modules on disk indexed by their content hash,
// allowing fast instantiation of previously-compiled modules.
type CompiledModuleCache struct {
	cacheDir string
	logger   *slog.Logger
	mu       sync.RWMutex
	stats    CacheStats
}

// CacheStats tracks cache hit/miss statistics.
type CacheStats struct {
	Hits       int64
	Misses     int64
	TotalBytes int64
	ModuleCount int
}

// CachedModule represents a cached compiled WASM module.
type CachedModule struct {
	Hash       string
	Path       string
	Size       int64
	CompiledAt time.Time
}

// NewCompiledModuleCache creates a new CompiledModuleCache.
func NewCompiledModuleCache(cacheDir string, logger *slog.Logger) (*CompiledModuleCache, error) {
	if cacheDir == "" {
		return nil, errors.New("cache directory is required")
	}
	if logger == nil {
		logger = slog.Default()
	}

	// Ensure cache directory exists
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	cache := &CompiledModuleCache{
		cacheDir: cacheDir,
		logger:   logger,
	}

	// Initialize stats by scanning existing cache
	if err := cache.updateStats(); err != nil {
		logger.Warn("failed to initialize cache stats", slog.Any("error", err))
	}

	return cache, nil
}

// GetOrCompile returns a cached compiled module or compiles and caches a new one.
// The compileFunc is called if the module is not in cache.
func (c *CompiledModuleCache) GetOrCompile(
	ctx context.Context,
	wasmBytes []byte,
	compileFunc func(ctx context.Context, wasmBytes []byte) ([]byte, error),
) ([]byte, error) {
	hash := c.hashBytes(wasmBytes)
	cachePath := c.getCachePath(hash)

	// Try loading from cache
	c.mu.RLock()
	if compiled, err := os.ReadFile(cachePath); err == nil {
		c.mu.RUnlock()
		c.recordHit()
		c.logger.Debug("cache hit for WASM module",
			slog.String("hash", hash),
			slog.Int("size", len(compiled)),
		)
		return compiled, nil
	}
	c.mu.RUnlock()

	// Cache miss - compile and store
	c.recordMiss()
	c.logger.Debug("cache miss for WASM module, compiling",
		slog.String("hash", hash),
		slog.Int("wasmSize", len(wasmBytes)),
	)

	compiled, err := compileFunc(ctx, wasmBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to compile WASM module: %w", err)
	}

	// Store in cache
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create parent directory if needed
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		c.logger.Warn("failed to create cache subdirectory",
			slog.Any("error", err),
			slog.String("path", cachePath),
		)
		// Return compiled module even if caching fails
		return compiled, nil
	}

	if err := os.WriteFile(cachePath, compiled, 0644); err != nil {
		c.logger.Warn("failed to cache compiled module",
			slog.Any("error", err),
			slog.String("hash", hash),
		)
		// Return compiled module even if caching fails
		return compiled, nil
	}

	c.logger.Debug("cached compiled WASM module",
		slog.String("hash", hash),
		slog.Int("compiledSize", len(compiled)),
	)

	return compiled, nil
}

// Get retrieves a cached module by its content hash.
func (c *CompiledModuleCache) Get(hash string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cachePath := c.getCachePath(hash)
	compiled, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("module not found in cache: %s", hash)
		}
		return nil, fmt.Errorf("failed to read cached module: %w", err)
	}

	c.recordHit()
	return compiled, nil
}

// Put stores a compiled module in the cache.
func (c *CompiledModuleCache) Put(hash string, compiled []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cachePath := c.getCachePath(hash)

	// Create parent directory if needed
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		return fmt.Errorf("failed to create cache subdirectory: %w", err)
	}

	if err := os.WriteFile(cachePath, compiled, 0644); err != nil {
		return fmt.Errorf("failed to write cached module: %w", err)
	}

	return nil
}

// Has checks if a module with the given hash exists in the cache.
func (c *CompiledModuleCache) Has(hash string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cachePath := c.getCachePath(hash)
	_, err := os.Stat(cachePath)
	return err == nil
}

// Delete removes a cached module by its hash.
func (c *CompiledModuleCache) Delete(hash string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cachePath := c.getCachePath(hash)
	if err := os.Remove(cachePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete cached module: %w", err)
	}

	return nil
}

// Clear removes all cached modules.
func (c *CompiledModuleCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	entries, err := os.ReadDir(c.cacheDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	var errs []error
	for _, entry := range entries {
		path := filepath.Join(c.cacheDir, entry.Name())
		if err := os.RemoveAll(path); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to clear some cache entries: %v", errs)
	}

	c.stats = CacheStats{}
	return nil
}

// Stats returns the current cache statistics.
func (c *CompiledModuleCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stats
}

// List returns all cached modules.
func (c *CompiledModuleCache) List() ([]CachedModule, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var modules []CachedModule

	err := filepath.Walk(c.cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(c.cacheDir, path)
		if err != nil {
			return err
		}

		modules = append(modules, CachedModule{
			Hash:       relPath,
			Path:       path,
			Size:       info.Size(),
			CompiledAt: info.ModTime(),
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list cached modules: %w", err)
	}

	return modules, nil
}

// Prune removes cached modules older than the specified duration.
func (c *CompiledModuleCache) Prune(maxAge time.Duration) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	var removed int

	err := filepath.Walk(c.cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		if info.ModTime().Before(cutoff) {
			if err := os.Remove(path); err != nil {
				c.logger.Warn("failed to prune cached module",
					slog.Any("error", err),
					slog.String("path", path),
				)
			} else {
				removed++
			}
		}

		return nil
	})

	if err != nil {
		return removed, fmt.Errorf("failed to prune cache: %w", err)
	}

	// Update stats after pruning
	if err := c.updateStats(); err != nil {
		c.logger.Warn("failed to update cache stats after pruning", slog.Any("error", err))
	}

	return removed, nil
}

// HashBytes computes the SHA256 hash of the given bytes and returns it as a hex string.
func (c *CompiledModuleCache) HashBytes(data []byte) string {
	return c.hashBytes(data)
}

func (c *CompiledModuleCache) hashBytes(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (c *CompiledModuleCache) getCachePath(hash string) string {
	// Use first 2 characters as subdirectory for better filesystem performance
	if len(hash) >= 2 {
		return filepath.Join(c.cacheDir, hash[:2], hash)
	}
	return filepath.Join(c.cacheDir, hash)
}

func (c *CompiledModuleCache) recordHit() {
	c.mu.Lock()
	c.stats.Hits++
	c.mu.Unlock()
}

func (c *CompiledModuleCache) recordMiss() {
	c.mu.Lock()
	c.stats.Misses++
	c.mu.Unlock()
}

func (c *CompiledModuleCache) updateStats() error {
	var totalBytes int64
	var moduleCount int

	err := filepath.Walk(c.cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalBytes += info.Size()
			moduleCount++
		}
		return nil
	})

	if err != nil && !os.IsNotExist(err) {
		return err
	}

	c.stats.TotalBytes = totalBytes
	c.stats.ModuleCount = moduleCount
	return nil
}
