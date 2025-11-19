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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPluginMetadataCache(t *testing.T) {
	t.Parallel()

	t.Run("creates database", func(t *testing.T) {
		t.Parallel()
		cache, err := NewPluginMetadataCache(t.TempDir(), nil)
		require.NoError(t, err)
		require.NotNil(t, cache)
		defer cache.Close()
	})

	t.Run("returns error for empty directory", func(t *testing.T) {
		t.Parallel()
		cache, err := NewPluginMetadataCache("", nil)
		require.Error(t, err)
		assert.Nil(t, cache)
	})
}

func TestPluginMetadataCache_PutGet(t *testing.T) {
	t.Parallel()

	cache, err := NewPluginMetadataCache(t.TempDir(), nil)
	require.NoError(t, err)
	defer cache.Close()

	plugin := &CachedPlugin{
		Ref:        "buf.build/protocolbuffers/go:v1.28.1",
		Digest:     "sha256:abc123",
		Type:       PluginTypeWasm,
		LocalPath:  "/cache/plugins/abc123",
		ResolvedAt: time.Now(),
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	}

	// Put
	err = cache.Put(plugin)
	require.NoError(t, err)

	// Get
	result, err := cache.Get(plugin.Ref)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, plugin.Ref, result.Ref)
	assert.Equal(t, plugin.Digest, result.Digest)
	assert.Equal(t, plugin.Type, result.Type)
	assert.Equal(t, plugin.LocalPath, result.LocalPath)
}

func TestPluginMetadataCache_GetExpired(t *testing.T) {
	t.Parallel()

	cache, err := NewPluginMetadataCache(t.TempDir(), nil)
	require.NoError(t, err)
	defer cache.Close()

	plugin := &CachedPlugin{
		Ref:        "buf.build/expired/plugin:v1.0.0",
		Digest:     "sha256:expired",
		Type:       PluginTypeDocker,
		LocalPath:  "/cache/plugins/expired",
		ResolvedAt: time.Now().Add(-48 * time.Hour),
		ExpiresAt:  time.Now().Add(-24 * time.Hour), // Already expired
	}

	// Put expired plugin
	err = cache.Put(plugin)
	require.NoError(t, err)

	// Get should return nil for expired
	result, err := cache.Get(plugin.Ref)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestPluginMetadataCache_Delete(t *testing.T) {
	t.Parallel()

	cache, err := NewPluginMetadataCache(t.TempDir(), nil)
	require.NoError(t, err)
	defer cache.Close()

	plugin := &CachedPlugin{
		Ref:        "buf.build/test/plugin:v1.0.0",
		Digest:     "sha256:test",
		Type:       PluginTypeWasm,
		LocalPath:  "/cache/plugins/test",
		ResolvedAt: time.Now(),
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	}

	// Put and verify
	err = cache.Put(plugin)
	require.NoError(t, err)
	result, err := cache.Get(plugin.Ref)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Delete
	err = cache.Delete(plugin.Ref)
	require.NoError(t, err)

	// Verify deleted
	result, err = cache.Get(plugin.Ref)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestPluginMetadataCache_Clear(t *testing.T) {
	t.Parallel()

	cache, err := NewPluginMetadataCache(t.TempDir(), nil)
	require.NoError(t, err)
	defer cache.Close()

	// Add multiple plugins
	for i := 0; i < 5; i++ {
		plugin := &CachedPlugin{
			Ref:        "buf.build/test/plugin" + string(rune('a'+i)) + ":v1.0.0",
			Digest:     "sha256:test",
			Type:       PluginTypeWasm,
			LocalPath:  "/cache/plugins/test",
			ResolvedAt: time.Now(),
			ExpiresAt:  time.Now().Add(24 * time.Hour),
		}
		err = cache.Put(plugin)
		require.NoError(t, err)
	}

	// Verify count
	count, err := cache.Count()
	require.NoError(t, err)
	assert.Equal(t, 5, count)

	// Clear
	err = cache.Clear()
	require.NoError(t, err)

	// Verify empty
	count, err = cache.Count()
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestPluginMetadataCache_List(t *testing.T) {
	t.Parallel()

	cache, err := NewPluginMetadataCache(t.TempDir(), nil)
	require.NoError(t, err)
	defer cache.Close()

	// Add plugins
	plugins := []*CachedPlugin{
		{
			Ref:        "buf.build/protocolbuffers/go:v1.28.1",
			Digest:     "sha256:abc",
			Type:       PluginTypeWasm,
			LocalPath:  "/cache/1",
			ResolvedAt: time.Now(),
			ExpiresAt:  time.Now().Add(24 * time.Hour),
		},
		{
			Ref:        "buf.build/grpc/go:v1.2.0",
			Digest:     "sha256:def",
			Type:       PluginTypeDocker,
			LocalPath:  "/cache/2",
			ResolvedAt: time.Now(),
			ExpiresAt:  time.Now().Add(24 * time.Hour),
		},
	}

	for _, p := range plugins {
		err = cache.Put(p)
		require.NoError(t, err)
	}

	// List
	result, err := cache.List()
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestPluginMetadataCache_Prune(t *testing.T) {
	t.Parallel()

	cache, err := NewPluginMetadataCache(t.TempDir(), nil)
	require.NoError(t, err)
	defer cache.Close()

	// Add expired plugin
	expired := &CachedPlugin{
		Ref:        "buf.build/expired/plugin:v1.0.0",
		Digest:     "sha256:expired",
		Type:       PluginTypeWasm,
		LocalPath:  "/cache/expired",
		ResolvedAt: time.Now().Add(-48 * time.Hour),
		ExpiresAt:  time.Now().Add(-24 * time.Hour),
	}
	err = cache.Put(expired)
	require.NoError(t, err)

	// Add valid plugin
	valid := &CachedPlugin{
		Ref:        "buf.build/valid/plugin:v1.0.0",
		Digest:     "sha256:valid",
		Type:       PluginTypeWasm,
		LocalPath:  "/cache/valid",
		ResolvedAt: time.Now(),
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	}
	err = cache.Put(valid)
	require.NoError(t, err)

	// Prune
	removed, err := cache.Prune()
	require.NoError(t, err)
	assert.Equal(t, 1, removed)

	// Verify only valid remains
	count, err := cache.Count()
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestPluginMetadataCache_Stats(t *testing.T) {
	t.Parallel()

	cache, err := NewPluginMetadataCache(t.TempDir(), nil)
	require.NoError(t, err)
	defer cache.Close()

	// Initial stats
	stats, err := cache.Stats()
	require.NoError(t, err)
	assert.Equal(t, 0, stats.Count)
	assert.Equal(t, int64(0), stats.Hits)
	assert.Equal(t, int64(0), stats.Misses)

	// Add a plugin
	plugin := &CachedPlugin{
		Ref:        "buf.build/test/plugin:v1.0.0",
		Digest:     "sha256:test",
		Type:       PluginTypeWasm,
		LocalPath:  "/cache/test",
		ResolvedAt: time.Now(),
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	}
	err = cache.Put(plugin)
	require.NoError(t, err)

	// Miss (non-existent)
	_, err = cache.Get("nonexistent")
	require.NoError(t, err)

	// Hit
	_, err = cache.Get(plugin.Ref)
	require.NoError(t, err)

	stats, err = cache.Stats()
	require.NoError(t, err)
	assert.Equal(t, 1, stats.Count)
	assert.Equal(t, int64(1), stats.Hits)
	assert.Equal(t, int64(1), stats.Misses)
}

func TestCacheConfig_Default(t *testing.T) {
	t.Parallel()

	config := DefaultCacheConfig()
	assert.True(t, config.Enabled)
	assert.Equal(t, 24*time.Hour, config.Plugins.TTL)
	assert.Equal(t, int64(500*1024*1024), config.Plugins.MaxSize)
	assert.True(t, config.WASM.Compile)
	assert.Equal(t, 5, config.Docker.PoolSize)
}

func TestCacheManager(t *testing.T) {
	t.Parallel()

	t.Run("creates with default config", func(t *testing.T) {
		t.Parallel()
		config := DefaultCacheConfig()
		config.Directory = t.TempDir()

		manager, err := NewCacheManager(config, nil)
		require.NoError(t, err)
		require.NotNil(t, manager)
		defer manager.Close()
	})

	t.Run("respects disabled config", func(t *testing.T) {
		t.Parallel()
		config := DefaultCacheConfig()
		config.Directory = t.TempDir()
		config.Enabled = false

		manager, err := NewCacheManager(config, nil)
		require.NoError(t, err)
		defer manager.Close()

		// Cache operations should be no-ops when disabled
		result, err := manager.GetPlugin(context.Background(), "test-ref")
		require.NoError(t, err)
		assert.Nil(t, result)

		err = manager.CachePlugin(context.Background(), &CachedPlugin{})
		require.NoError(t, err)
	})

	t.Run("caches and retrieves plugins", func(t *testing.T) {
		t.Parallel()
		config := DefaultCacheConfig()
		config.Directory = t.TempDir()

		manager, err := NewCacheManager(config, nil)
		require.NoError(t, err)
		defer manager.Close()

		plugin := &CachedPlugin{
			Ref:        "buf.build/test/plugin:v1.0.0",
			Digest:     "sha256:test",
			Type:       PluginTypeWasm,
			LocalPath:  "/cache/test",
			ResolvedAt: time.Now(),
		}

		err = manager.CachePlugin(context.Background(), plugin)
		require.NoError(t, err)

		result, err := manager.GetPlugin(context.Background(), plugin.Ref)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, plugin.Ref, result.Ref)
	})

	t.Run("status returns cache info", func(t *testing.T) {
		t.Parallel()
		config := DefaultCacheConfig()
		config.Directory = t.TempDir()

		manager, err := NewCacheManager(config, nil)
		require.NoError(t, err)
		defer manager.Close()

		status, err := manager.Status(context.Background())
		require.NoError(t, err)
		assert.NotNil(t, status)
	})

	t.Run("clean removes cached data", func(t *testing.T) {
		t.Parallel()
		config := DefaultCacheConfig()
		config.Directory = t.TempDir()

		manager, err := NewCacheManager(config, nil)
		require.NoError(t, err)
		defer manager.Close()

		// Add some data
		plugin := &CachedPlugin{
			Ref:        "buf.build/test/plugin:v1.0.0",
			Digest:     "sha256:test",
			Type:       PluginTypeWasm,
			LocalPath:  "/cache/test",
			ResolvedAt: time.Now(),
		}
		err = manager.CachePlugin(context.Background(), plugin)
		require.NoError(t, err)

		// Clean
		err = manager.Clean(context.Background(), true, false, false, false)
		require.NoError(t, err)

		// Verify cleaned
		result, err := manager.GetPlugin(context.Background(), plugin.Ref)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestPluginType(t *testing.T) {
	t.Parallel()

	assert.Equal(t, PluginType("wasm"), PluginTypeWasm)
	assert.Equal(t, PluginType("docker"), PluginTypeDocker)
	assert.Equal(t, PluginType("binary"), PluginTypeBinary)
}

func TestInvalidationStrategy(t *testing.T) {
	t.Parallel()

	assert.Equal(t, InvalidationStrategy(0), TTLBased)
	assert.Equal(t, InvalidationStrategy(1), VersionBased)
	assert.Equal(t, InvalidationStrategy(2), Manual)
}

func TestDefaultTTLs(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 24*time.Hour, defaultTTLs["remote-plugin"])
	assert.Equal(t, 7*24*time.Hour, defaultTTLs["compiled-wasm"])
	assert.Equal(t, 1*time.Hour, defaultTTLs["container-pool"])
}
