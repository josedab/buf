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

package bufremoteplugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// PluginType represents the type of plugin.
type PluginType string

const (
	// PluginTypeWasm represents a WASM plugin.
	PluginTypeWasm PluginType = "wasm"
	// PluginTypeDocker represents a Docker plugin.
	PluginTypeDocker PluginType = "docker"
	// PluginTypeBinary represents a binary plugin.
	PluginTypeBinary PluginType = "binary"
)

// InvalidationStrategy defines how cache entries are invalidated.
type InvalidationStrategy int

const (
	// TTLBased invalidates entries after a time-to-live period.
	TTLBased InvalidationStrategy = iota
	// VersionBased invalidates entries when the version changes.
	VersionBased
	// Manual invalidates entries only through explicit user action.
	Manual
)

// Default TTLs for different cache types.
var defaultTTLs = map[string]time.Duration{
	"remote-plugin": 24 * time.Hour,
	"compiled-wasm": 7 * 24 * time.Hour,
	"container-pool": 1 * time.Hour,
}

// CachedPlugin represents a cached plugin entry.
type CachedPlugin struct {
	Ref        string     `json:"ref"`
	Digest     string     `json:"digest"`
	Type       PluginType `json:"type"`
	LocalPath  string     `json:"local_path"`
	ResolvedAt time.Time  `json:"resolved_at"`
	ExpiresAt  time.Time  `json:"expires_at"`
}

// PluginMetadataCache provides file-based caching for plugin metadata.
type PluginMetadataCache struct {
	cacheDir string
	logger   *slog.Logger
	mu       sync.RWMutex
	stats    MetadataCacheStats
}

// NewPluginMetadataCache creates a new metadata cache backed by JSON files.
func NewPluginMetadataCache(cacheDir string, logger *slog.Logger) (*PluginMetadataCache, error) {
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

	cache := &PluginMetadataCache{
		cacheDir: cacheDir,
		logger:   logger,
	}

	return cache, nil
}

// Get retrieves a cached plugin by reference.
func (c *PluginMetadataCache) Get(ref string) (*CachedPlugin, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	filePath := c.getPluginFilePath(ref)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			c.recordMiss()
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read plugin cache: %w", err)
	}

	var plugin CachedPlugin
	if err := json.Unmarshal(data, &plugin); err != nil {
		return nil, fmt.Errorf("failed to unmarshal plugin cache: %w", err)
	}

	// Check if expired
	if time.Now().After(plugin.ExpiresAt) {
		c.recordMiss()
		return nil, nil
	}

	c.recordHit()
	return &plugin, nil
}

// Put stores a plugin in the cache.
func (c *PluginMetadataCache) Put(plugin *CachedPlugin) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	filePath := c.getPluginFilePath(plugin.Ref)

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	data, err := json.MarshalIndent(plugin, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal plugin: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write plugin cache: %w", err)
	}

	return nil
}

// Delete removes a plugin from the cache.
func (c *PluginMetadataCache) Delete(ref string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	filePath := c.getPluginFilePath(ref)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete plugin cache: %w", err)
	}

	return nil
}

// Clear removes all cached plugins.
func (c *PluginMetadataCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	metadataDir := filepath.Join(c.cacheDir, "metadata")
	if err := os.RemoveAll(metadataDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clear plugin cache: %w", err)
	}

	// Reset stats
	c.stats = MetadataCacheStats{}
	return nil
}

// Prune removes expired entries from the cache.
func (c *PluginMetadataCache) Prune() (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	metadataDir := filepath.Join(c.cacheDir, "metadata")
	var removed int

	err := filepath.Walk(metadataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		var plugin CachedPlugin
		if err := json.Unmarshal(data, &plugin); err != nil {
			return nil
		}

		if time.Now().After(plugin.ExpiresAt) {
			if err := os.Remove(path); err == nil {
				removed++
			}
		}

		return nil
	})

	if err != nil {
		return removed, fmt.Errorf("failed to prune cache: %w", err)
	}

	return removed, nil
}

// List returns all cached plugins.
func (c *PluginMetadataCache) List() ([]*CachedPlugin, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	metadataDir := filepath.Join(c.cacheDir, "metadata")
	var plugins []*CachedPlugin

	err := filepath.Walk(metadataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		var plugin CachedPlugin
		if err := json.Unmarshal(data, &plugin); err != nil {
			return nil
		}

		plugins = append(plugins, &plugin)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list plugins: %w", err)
	}

	return plugins, nil
}

// Count returns the number of cached plugins.
func (c *PluginMetadataCache) Count() (int, error) {
	plugins, err := c.List()
	if err != nil {
		return 0, err
	}
	return len(plugins), nil
}

// MetadataCacheStats contains cache statistics.
type MetadataCacheStats struct {
	Hits   int64
	Misses int64
	Count  int
}

// Stats returns the cache statistics.
func (c *PluginMetadataCache) Stats() (MetadataCacheStats, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := c.stats
	count, err := c.Count()
	if err != nil {
		return stats, err
	}
	stats.Count = count

	return stats, nil
}

// Close is a no-op for file-based cache.
func (c *PluginMetadataCache) Close() error {
	return nil
}

func (c *PluginMetadataCache) getPluginFilePath(ref string) string {
	// Create a safe filename from the ref
	// Replace special characters
	safeRef := filepath.Base(ref)
	return filepath.Join(c.cacheDir, "metadata", safeRef+".json")
}

func (c *PluginMetadataCache) recordHit() {
	c.stats.Hits++
}

func (c *PluginMetadataCache) recordMiss() {
	c.stats.Misses++
}

// CacheConfig holds configuration for the plugin cache system.
type CacheConfig struct {
	Enabled   bool          `yaml:"enabled"`
	Directory string        `yaml:"directory"`
	Plugins   PluginsTTL    `yaml:"plugins"`
	WASM      WASMConfig    `yaml:"wasm"`
	Docker    DockerConfig  `yaml:"docker"`
}

// PluginsTTL configures plugin cache TTL settings.
type PluginsTTL struct {
	TTL     time.Duration `yaml:"ttl"`
	MaxSize int64         `yaml:"max_size"` // in bytes
}

// WASMConfig configures WASM caching.
type WASMConfig struct {
	Compile bool `yaml:"compile"`
}

// DockerConfig configures Docker container pooling.
type DockerConfig struct {
	PoolSize int `yaml:"pool_size"`
}

// DefaultCacheConfig returns the default cache configuration.
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		Enabled:   true,
		Directory: "", // Will be set to ~/.cache/buf at runtime
		Plugins: PluginsTTL{
			TTL:     24 * time.Hour,
			MaxSize: 500 * 1024 * 1024, // 500MB
		},
		WASM: WASMConfig{
			Compile: true,
		},
		Docker: DockerConfig{
			PoolSize: 5,
		},
	}
}

// CacheManager coordinates all caching components.
type CacheManager struct {
	config       CacheConfig
	metadataCache *PluginMetadataCache
	logger       *slog.Logger
	mu           sync.RWMutex
}

// NewCacheManager creates a new cache manager.
func NewCacheManager(config CacheConfig, logger *slog.Logger) (*CacheManager, error) {
	if logger == nil {
		logger = slog.Default()
	}

	if config.Directory == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		config.Directory = filepath.Join(homeDir, ".cache", "buf", "v2")
	}

	// Create metadata cache
	metadataCache, err := NewPluginMetadataCache(config.Directory, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create metadata cache: %w", err)
	}

	return &CacheManager{
		config:        config,
		metadataCache: metadataCache,
		logger:        logger,
	}, nil
}

// GetPlugin retrieves a cached plugin by reference.
func (m *CacheManager) GetPlugin(ctx context.Context, ref string) (*CachedPlugin, error) {
	if !m.config.Enabled {
		return nil, nil
	}
	return m.metadataCache.Get(ref)
}

// CachePlugin stores a plugin in the cache.
func (m *CacheManager) CachePlugin(ctx context.Context, plugin *CachedPlugin) error {
	if !m.config.Enabled {
		return nil
	}

	// Set expiration based on TTL
	if plugin.ExpiresAt.IsZero() {
		plugin.ExpiresAt = time.Now().Add(m.config.Plugins.TTL)
	}

	return m.metadataCache.Put(plugin)
}

// CacheStatus represents the overall cache status.
type CacheStatus struct {
	PluginCount      int
	PluginCacheSize  int64 // in bytes
	WASMModules      int
	WASMCacheSize    int64 // in bytes
	WarmContainers   int
	MetadataStats    MetadataCacheStats
}

// Status returns the overall cache status.
func (m *CacheManager) Status(ctx context.Context) (*CacheStatus, error) {
	status := &CacheStatus{}

	// Get metadata stats
	stats, err := m.metadataCache.Stats()
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata stats: %w", err)
	}
	status.MetadataStats = stats
	status.PluginCount = stats.Count

	// Calculate plugin cache size
	pluginDir := filepath.Join(m.config.Directory, "plugins")
	status.PluginCacheSize, _ = dirSize(pluginDir)

	// Calculate WASM cache size
	wasmDir := filepath.Join(m.config.Directory, "wasm")
	status.WASMCacheSize, _ = dirSize(wasmDir)

	return status, nil
}

// Clean removes cached data based on the provided options.
func (m *CacheManager) Clean(ctx context.Context, plugins, wasm, containers, all bool) error {
	if all {
		plugins, wasm, containers = true, true, true
	}

	if plugins {
		if err := m.metadataCache.Clear(); err != nil {
			return fmt.Errorf("failed to clear plugin cache: %w", err)
		}
		// Also remove plugin files
		pluginDir := filepath.Join(m.config.Directory, "plugins")
		if err := os.RemoveAll(pluginDir); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove plugin directory: %w", err)
		}
		m.logger.Info("cleared plugin cache")
	}

	if wasm {
		wasmDir := filepath.Join(m.config.Directory, "wasm")
		if err := os.RemoveAll(wasmDir); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove WASM cache: %w", err)
		}
		m.logger.Info("cleared WASM cache")
	}

	if containers {
		// Note: Container pool cleanup is handled separately
		m.logger.Info("container pool cleanup requested")
	}

	return nil
}

// Close closes the cache manager and releases resources.
func (m *CacheManager) Close() error {
	return m.metadataCache.Close()
}

// dirSize calculates the total size of a directory.
func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}
