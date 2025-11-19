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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestNewCompilationCache(t *testing.T) {
	t.Parallel()

	cache := NewCompilationCache()
	assert.NotNil(t, cache)
	assert.Equal(t, 0, cache.Size())
}

func TestCompilationCacheWithOptions(t *testing.T) {
	t.Parallel()

	cache := NewCompilationCache(
		WithCompilationCacheMaxSize(100),
		WithCompilationCacheTTL(time.Minute),
	)
	assert.NotNil(t, cache)
}

func TestCompilationCachePutAndGet(t *testing.T) {
	t.Parallel()

	cache := NewCompilationCache()

	fdp := &descriptorpb.FileDescriptorProto{
		Name:   stringPtr("test.proto"),
		Syntax: stringPtr("proto3"),
	}

	// Put entry
	cache.Put("test.proto", "hash123", fdp)
	assert.Equal(t, 1, cache.Size())

	// Get entry with correct hash
	result, ok := cache.Get("test.proto", "hash123")
	require.True(t, ok)
	assert.Equal(t, "test.proto", result.GetName())

	// Get entry with wrong hash
	result, ok = cache.Get("test.proto", "wronghash")
	assert.False(t, ok)
	assert.Nil(t, result)

	// Get non-existent entry
	result, ok = cache.Get("nonexistent.proto", "hash123")
	assert.False(t, ok)
	assert.Nil(t, result)
}

func TestCompilationCacheInvalidate(t *testing.T) {
	t.Parallel()

	cache := NewCompilationCache()

	fdp := &descriptorpb.FileDescriptorProto{
		Name:   stringPtr("test.proto"),
		Syntax: stringPtr("proto3"),
	}

	cache.Put("test.proto", "hash123", fdp)
	assert.Equal(t, 1, cache.Size())

	// Invalidate entry
	cache.Invalidate("test.proto")
	assert.Equal(t, 0, cache.Size())

	// Should not be retrievable
	result, ok := cache.Get("test.proto", "hash123")
	assert.False(t, ok)
	assert.Nil(t, result)
}

func TestCompilationCacheInvalidateAll(t *testing.T) {
	t.Parallel()

	cache := NewCompilationCache()

	// Add multiple entries
	for i := 0; i < 10; i++ {
		name := "test" + string(rune('0'+i)) + ".proto"
		fdp := &descriptorpb.FileDescriptorProto{
			Name:   &name,
			Syntax: stringPtr("proto3"),
		}
		cache.Put(name, "hash"+string(rune('0'+i)), fdp)
	}
	assert.Equal(t, 10, cache.Size())

	// Invalidate all
	cache.InvalidateAll()
	assert.Equal(t, 0, cache.Size())
}

func TestCompilationCacheTTL(t *testing.T) {
	t.Parallel()

	// Create cache with very short TTL
	cache := NewCompilationCache(
		WithCompilationCacheTTL(time.Millisecond * 10),
	)

	fdp := &descriptorpb.FileDescriptorProto{
		Name:   stringPtr("test.proto"),
		Syntax: stringPtr("proto3"),
	}

	cache.Put("test.proto", "hash123", fdp)

	// Should be retrievable immediately
	result, ok := cache.Get("test.proto", "hash123")
	require.True(t, ok)
	assert.NotNil(t, result)

	// Wait for TTL to expire
	time.Sleep(time.Millisecond * 20)

	// Should no longer be retrievable
	result, ok = cache.Get("test.proto", "hash123")
	assert.False(t, ok)
	assert.Nil(t, result)
}

func TestCompilationCacheEviction(t *testing.T) {
	t.Parallel()

	// Create cache with small max size
	cache := NewCompilationCache(
		WithCompilationCacheMaxSize(3),
	)

	// Add entries up to capacity
	for i := 0; i < 3; i++ {
		name := "test" + string(rune('0'+i)) + ".proto"
		fdp := &descriptorpb.FileDescriptorProto{
			Name:   &name,
			Syntax: stringPtr("proto3"),
		}
		cache.Put(name, "hash"+string(rune('0'+i)), fdp)
	}
	assert.Equal(t, 3, cache.Size())

	// Add one more entry, should trigger eviction
	fdp := &descriptorpb.FileDescriptorProto{
		Name:   stringPtr("test3.proto"),
		Syntax: stringPtr("proto3"),
	}
	cache.Put("test3.proto", "hash3", fdp)

	// Size should still be 3 (one evicted)
	assert.Equal(t, 3, cache.Size())
}

func TestCompilationCacheUpdate(t *testing.T) {
	t.Parallel()

	cache := NewCompilationCache()

	fdp1 := &descriptorpb.FileDescriptorProto{
		Name:    stringPtr("test.proto"),
		Syntax:  stringPtr("proto3"),
		Package: stringPtr("old"),
	}

	fdp2 := &descriptorpb.FileDescriptorProto{
		Name:    stringPtr("test.proto"),
		Syntax:  stringPtr("proto3"),
		Package: stringPtr("new"),
	}

	// Put first version
	cache.Put("test.proto", "hash1", fdp1)

	// Update with new version and new hash
	cache.Put("test.proto", "hash2", fdp2)
	assert.Equal(t, 1, cache.Size())

	// Old hash should not work
	result, ok := cache.Get("test.proto", "hash1")
	assert.False(t, ok)
	assert.Nil(t, result)

	// New hash should work
	result, ok = cache.Get("test.proto", "hash2")
	require.True(t, ok)
	assert.Equal(t, "new", result.GetPackage())
}

func TestNewCachedCompiler(t *testing.T) {
	t.Parallel()

	cache := NewCompilationCache()
	compiler := NewStreamingCompiler(nil)
	cachedCompiler := NewCachedCompiler(cache, compiler, nil)

	assert.NotNil(t, cachedCompiler)
	assert.Equal(t, cache, cachedCompiler.GetCache())
}

func TestCachedCompilerStats(t *testing.T) {
	t.Parallel()

	cache := NewCompilationCache(WithCompilationCacheMaxSize(100))
	compiler := NewStreamingCompiler(nil)
	cachedCompiler := NewCachedCompiler(cache, compiler, nil)

	stats := cachedCompiler.Stats()
	assert.Equal(t, 0, stats.Size)
	assert.Equal(t, 100, stats.MaxSize)
}

func TestHashFileDescriptor(t *testing.T) {
	t.Parallel()

	fdp1 := &descriptorpb.FileDescriptorProto{
		Name:   stringPtr("test.proto"),
		Syntax: stringPtr("proto3"),
	}

	fdp2 := &descriptorpb.FileDescriptorProto{
		Name:   stringPtr("test.proto"),
		Syntax: stringPtr("proto3"),
	}

	fdp3 := &descriptorpb.FileDescriptorProto{
		Name:    stringPtr("test.proto"),
		Syntax:  stringPtr("proto3"),
		Package: stringPtr("different"),
	}

	// Same content should produce same hash
	hash1 := HashFileDescriptor(fdp1)
	hash2 := HashFileDescriptor(fdp2)
	assert.Equal(t, hash1, hash2)

	// Different content should produce different hash
	hash3 := HashFileDescriptor(fdp3)
	assert.NotEqual(t, hash1, hash3)
}

func TestHashFileDescriptorNil(t *testing.T) {
	t.Parallel()

	// Nil should return empty string (or handle gracefully)
	hash := HashFileDescriptor(nil)
	// proto.Marshal on nil returns empty bytes, so hash is of empty input
	assert.NotEmpty(t, hash)
}
