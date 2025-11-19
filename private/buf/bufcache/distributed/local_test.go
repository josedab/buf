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

package distributed

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalBackendGetPut(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	config := BackendConfig{
		Type: "local",
		Path: tempDir,
	}

	backend, err := NewLocalBackend(config)
	require.NoError(t, err)

	ctx := context.Background()

	key := CacheKey{
		ContentHash: "testhash123",
		Version:     "1.0.0",
		Options:     CompileOptions{},
	}
	data := []byte("test data content")

	// Test Put
	err = backend.Put(ctx, key, data)
	require.NoError(t, err)

	// Test Get
	retrieved, err := backend.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, data, retrieved)
}

func TestLocalBackendGetMiss(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	config := BackendConfig{
		Type: "local",
		Path: tempDir,
	}

	backend, err := NewLocalBackend(config)
	require.NoError(t, err)

	ctx := context.Background()

	key := CacheKey{
		ContentHash: "nonexistent",
		Version:     "1.0.0",
		Options:     CompileOptions{},
	}

	_, err = backend.Get(ctx, key)
	assert.ErrorIs(t, err, ErrCacheMiss)
}

func TestLocalBackendDelete(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	config := BackendConfig{
		Type: "local",
		Path: tempDir,
	}

	backend, err := NewLocalBackend(config)
	require.NoError(t, err)

	ctx := context.Background()

	key := CacheKey{
		ContentHash: "todelete",
		Version:     "1.0.0",
		Options:     CompileOptions{},
	}
	data := []byte("test data content")

	// Put first
	err = backend.Put(ctx, key, data)
	require.NoError(t, err)

	// Delete
	err = backend.Delete(ctx, key)
	require.NoError(t, err)

	// Verify deleted
	_, err = backend.Get(ctx, key)
	assert.ErrorIs(t, err, ErrCacheMiss)
}

func TestLocalBackendDeleteNonexistent(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	config := BackendConfig{
		Type: "local",
		Path: tempDir,
	}

	backend, err := NewLocalBackend(config)
	require.NoError(t, err)

	ctx := context.Background()

	key := CacheKey{
		ContentHash: "nonexistent",
		Version:     "1.0.0",
		Options:     CompileOptions{},
	}

	// Should not error when deleting non-existent key
	err = backend.Delete(ctx, key)
	assert.NoError(t, err)
}

func TestLocalBackendStats(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	config := BackendConfig{
		Type: "local",
		Path: tempDir,
	}

	backend, err := NewLocalBackend(config)
	require.NoError(t, err)

	ctx := context.Background()

	// Empty stats
	stats, err := backend.Stats(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(0), stats.EntryCount)
	assert.Equal(t, int64(0), stats.Size)

	// Add entries
	key1 := CacheKey{ContentHash: "hash1", Version: "1.0.0"}
	key2 := CacheKey{ContentHash: "hash2", Version: "1.0.0"}
	data1 := []byte("data1")
	data2 := []byte("longer data 2")

	require.NoError(t, backend.Put(ctx, key1, data1))
	require.NoError(t, backend.Put(ctx, key2, data2))

	// Check stats
	stats, err = backend.Stats(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(2), stats.EntryCount)
	assert.Equal(t, int64(len(data1)+len(data2)), stats.Size)
}

func TestLocalBackendClear(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	config := BackendConfig{
		Type: "local",
		Path: tempDir,
	}

	backend, err := NewLocalBackend(config)
	require.NoError(t, err)

	ctx := context.Background()

	// Add entries
	key1 := CacheKey{ContentHash: "hash1", Version: "1.0.0"}
	key2 := CacheKey{ContentHash: "hash2", Version: "1.0.0"}
	require.NoError(t, backend.Put(ctx, key1, []byte("data1")))
	require.NoError(t, backend.Put(ctx, key2, []byte("data2")))

	// Clear
	err = backend.Clear(ctx)
	require.NoError(t, err)

	// Verify cleared
	stats, err := backend.Stats(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(0), stats.EntryCount)
}

func TestCacheStatsHitRate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		hits     int64
		misses   int64
		expected float64
	}{
		{"no requests", 0, 0, 0},
		{"all hits", 10, 0, 100},
		{"all misses", 0, 10, 0},
		{"50% hit rate", 5, 5, 50},
		{"80% hit rate", 80, 20, 80},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			stats := &CacheStats{
				Hits:   tt.hits,
				Misses: tt.misses,
			}
			assert.Equal(t, tt.expected, stats.HitRate())
		})
	}
}
