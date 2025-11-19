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
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
)

// LocalBackend implements CacheBackend using the local filesystem.
// This is primarily useful for testing and development.
type LocalBackend struct {
	path   string
	hits   atomic.Int64
	misses atomic.Int64
}

// NewLocalBackend creates a new local filesystem cache backend.
func NewLocalBackend(backendConfig BackendConfig) (*LocalBackend, error) {
	if err := os.MkdirAll(backendConfig.Path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &LocalBackend{
		path: backendConfig.Path,
	}, nil
}

// Get retrieves data from the local filesystem for the given key.
func (b *LocalBackend) Get(ctx context.Context, key CacheKey) ([]byte, error) {
	filePath := b.keyPath(key)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			b.misses.Add(1)
			return nil, ErrCacheMiss
		}
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	b.hits.Add(1)
	return data, nil
}

// Put stores data on the local filesystem with the given key.
func (b *LocalBackend) Put(ctx context.Context, key CacheKey, data []byte) error {
	filePath := b.keyPath(key)

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Write to temp file first, then rename for atomicity
	tempFile := filePath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	if err := os.Rename(tempFile, filePath); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("failed to rename cache file: %w", err)
	}

	return nil
}

// Delete removes data from the local filesystem for the given key.
func (b *LocalBackend) Delete(ctx context.Context, key CacheKey) error {
	filePath := b.keyPath(key)

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to delete cache file: %w", err)
	}

	return nil
}

// Stats returns cache statistics.
func (b *LocalBackend) Stats(ctx context.Context) (*CacheStats, error) {
	stats := &CacheStats{
		Hits:   b.hits.Load(),
		Misses: b.misses.Load(),
	}

	err := filepath.Walk(b.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".bin" {
			stats.Size += info.Size()
			stats.EntryCount++
		}
		return nil
	})

	if err != nil {
		return stats, fmt.Errorf("failed to walk cache directory: %w", err)
	}

	return stats, nil
}

// Clear removes all entries from the local cache.
func (b *LocalBackend) Clear(ctx context.Context) error {
	return filepath.Walk(b.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".bin" {
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("failed to remove cache file %s: %w", path, err)
			}
		}
		return nil
	})
}

// keyPath returns the full file path for the given cache key.
func (b *LocalBackend) keyPath(key CacheKey) string {
	return filepath.Join(b.path, key.Version, key.ContentHash+".bin")
}
