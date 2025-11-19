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

package bufremotepluginwasm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCompiledModuleCache(t *testing.T) {
	t.Parallel()

	t.Run("creates cache directory", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		cacheDir := filepath.Join(tempDir, "cache")

		cache, err := NewCompiledModuleCache(cacheDir, nil)
		require.NoError(t, err)
		require.NotNil(t, cache)

		// Verify directory was created
		info, err := os.Stat(cacheDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("returns error for empty directory", func(t *testing.T) {
		t.Parallel()
		cache, err := NewCompiledModuleCache("", nil)
		require.Error(t, err)
		assert.Nil(t, cache)
	})
}

func TestCompiledModuleCache_GetOrCompile(t *testing.T) {
	t.Parallel()

	t.Run("caches compiled module", func(t *testing.T) {
		t.Parallel()
		cache, err := NewCompiledModuleCache(t.TempDir(), nil)
		require.NoError(t, err)

		wasmBytes := []byte("test wasm module")
		compiledBytes := []byte("compiled wasm module")
		compileCount := 0

		compileFunc := func(ctx context.Context, wasm []byte) ([]byte, error) {
			compileCount++
			return compiledBytes, nil
		}

		// First call should compile
		result, err := cache.GetOrCompile(context.Background(), wasmBytes, compileFunc)
		require.NoError(t, err)
		assert.Equal(t, compiledBytes, result)
		assert.Equal(t, 1, compileCount)

		// Second call should use cache
		result, err = cache.GetOrCompile(context.Background(), wasmBytes, compileFunc)
		require.NoError(t, err)
		assert.Equal(t, compiledBytes, result)
		assert.Equal(t, 1, compileCount) // Should not have called compile again
	})

	t.Run("tracks stats", func(t *testing.T) {
		t.Parallel()
		cache, err := NewCompiledModuleCache(t.TempDir(), nil)
		require.NoError(t, err)

		wasmBytes := []byte("test wasm")
		compileFunc := func(ctx context.Context, wasm []byte) ([]byte, error) {
			return []byte("compiled"), nil
		}

		// First call - miss
		_, err = cache.GetOrCompile(context.Background(), wasmBytes, compileFunc)
		require.NoError(t, err)

		stats := cache.Stats()
		assert.Equal(t, int64(1), stats.Misses)
		assert.Equal(t, int64(0), stats.Hits)

		// Second call - hit
		_, err = cache.GetOrCompile(context.Background(), wasmBytes, compileFunc)
		require.NoError(t, err)

		stats = cache.Stats()
		assert.Equal(t, int64(1), stats.Misses)
		assert.Equal(t, int64(1), stats.Hits)
	})
}

func TestCompiledModuleCache_PutGet(t *testing.T) {
	t.Parallel()

	cache, err := NewCompiledModuleCache(t.TempDir(), nil)
	require.NoError(t, err)

	hash := "testhash123456789012345678901234567890123456"
	data := []byte("compiled module data")

	// Put data
	err = cache.Put(hash, data)
	require.NoError(t, err)

	// Get data
	result, err := cache.Get(hash)
	require.NoError(t, err)
	assert.Equal(t, data, result)

	// Check Has
	assert.True(t, cache.Has(hash))
	assert.False(t, cache.Has("nonexistent"))
}

func TestCompiledModuleCache_Delete(t *testing.T) {
	t.Parallel()

	cache, err := NewCompiledModuleCache(t.TempDir(), nil)
	require.NoError(t, err)

	hash := "testhash123456789012345678901234567890123456"
	data := []byte("test data")

	// Put and verify
	err = cache.Put(hash, data)
	require.NoError(t, err)
	assert.True(t, cache.Has(hash))

	// Delete
	err = cache.Delete(hash)
	require.NoError(t, err)
	assert.False(t, cache.Has(hash))
}

func TestCompiledModuleCache_Clear(t *testing.T) {
	t.Parallel()

	cache, err := NewCompiledModuleCache(t.TempDir(), nil)
	require.NoError(t, err)

	// Add some data
	for i := 0; i < 5; i++ {
		hash := sha256.Sum256([]byte{byte(i)})
		err = cache.Put(hex.EncodeToString(hash[:]), []byte("data"))
		require.NoError(t, err)
	}

	// Clear
	err = cache.Clear()
	require.NoError(t, err)

	// Verify empty
	modules, err := cache.List()
	require.NoError(t, err)
	assert.Empty(t, modules)
}

func TestCompiledModuleCache_List(t *testing.T) {
	t.Parallel()

	cache, err := NewCompiledModuleCache(t.TempDir(), nil)
	require.NoError(t, err)

	// Add some data
	hashes := make([]string, 3)
	for i := 0; i < 3; i++ {
		hash := sha256.Sum256([]byte{byte(i)})
		hashes[i] = hex.EncodeToString(hash[:])
		err = cache.Put(hashes[i], []byte("data"))
		require.NoError(t, err)
	}

	// List
	modules, err := cache.List()
	require.NoError(t, err)
	assert.Len(t, modules, 3)
}

func TestCompiledModuleCache_Prune(t *testing.T) {
	t.Parallel()

	cacheDir := t.TempDir()
	cache, err := NewCompiledModuleCache(cacheDir, nil)
	require.NoError(t, err)

	// Add a file and modify its time to be old
	hash := sha256.Sum256([]byte("old"))
	hashStr := hex.EncodeToString(hash[:])
	err = cache.Put(hashStr, []byte("old data"))
	require.NoError(t, err)

	// Get the path and modify its modification time
	cachePath := filepath.Join(cacheDir, hashStr[:2], hashStr)
	oldTime := time.Now().Add(-48 * time.Hour)
	err = os.Chtimes(cachePath, oldTime, oldTime)
	require.NoError(t, err)

	// Add a recent file
	hash2 := sha256.Sum256([]byte("new"))
	hashStr2 := hex.EncodeToString(hash2[:])
	err = cache.Put(hashStr2, []byte("new data"))
	require.NoError(t, err)

	// Prune with 24h max age
	removed, err := cache.Prune(24 * time.Hour)
	require.NoError(t, err)
	assert.Equal(t, 1, removed)

	// Verify only new file remains
	assert.False(t, cache.Has(hashStr))
	assert.True(t, cache.Has(hashStr2))
}

func TestCompiledModuleCache_HashBytes(t *testing.T) {
	t.Parallel()

	cache, err := NewCompiledModuleCache(t.TempDir(), nil)
	require.NoError(t, err)

	data := []byte("test data")
	hash := cache.HashBytes(data)

	// Verify it's a valid SHA256 hex string
	assert.Len(t, hash, 64)

	// Verify consistency
	hash2 := cache.HashBytes(data)
	assert.Equal(t, hash, hash2)

	// Verify different data produces different hash
	hash3 := cache.HashBytes([]byte("different data"))
	assert.NotEqual(t, hash, hash3)
}
