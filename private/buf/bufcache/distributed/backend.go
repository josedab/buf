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
	"errors"
	"time"
)

var (
	// ErrCacheMiss is returned when a cache key is not found.
	ErrCacheMiss = errors.New("cache miss")
	// ErrCacheCorruption is returned when cached data is corrupted.
	ErrCacheCorruption = errors.New("cache corruption detected")
)

// CacheBackend is the interface that all cache backends must implement.
type CacheBackend interface {
	// Get retrieves data from the cache for the given key.
	// Returns ErrCacheMiss if the key is not found.
	Get(ctx context.Context, key CacheKey) ([]byte, error)

	// Put stores data in the cache with the given key.
	Put(ctx context.Context, key CacheKey, data []byte) error

	// Delete removes data from the cache for the given key.
	// Returns nil if the key does not exist.
	Delete(ctx context.Context, key CacheKey) error

	// Stats returns cache statistics.
	Stats(ctx context.Context) (*CacheStats, error)
}

// CacheStats contains cache statistics.
type CacheStats struct {
	// Hits is the number of cache hits.
	Hits int64 `json:"hits"`
	// Misses is the number of cache misses.
	Misses int64 `json:"misses"`
	// Size is the total size of cached data in bytes.
	Size int64 `json:"size"`
	// EntryCount is the number of entries in the cache.
	EntryCount int64 `json:"entry_count"`
}

// HitRate returns the cache hit rate as a percentage.
func (s *CacheStats) HitRate() float64 {
	total := s.Hits + s.Misses
	if total == 0 {
		return 0
	}
	return float64(s.Hits) / float64(total) * 100
}

// BackendConfig contains common configuration for cache backends.
type BackendConfig struct {
	// Type is the backend type (s3, gcs, http, local).
	Type string `json:"type" yaml:"type"`
	// Bucket is the bucket name for S3/GCS backends.
	Bucket string `json:"bucket,omitempty" yaml:"bucket,omitempty"`
	// Prefix is the key prefix for S3/GCS backends.
	Prefix string `json:"prefix,omitempty" yaml:"prefix,omitempty"`
	// Region is the AWS region for S3 backend.
	Region string `json:"region,omitempty" yaml:"region,omitempty"`
	// URL is the base URL for HTTP backend.
	URL string `json:"url,omitempty" yaml:"url,omitempty"`
	// TokenEnv is the environment variable containing the auth token for HTTP backend.
	TokenEnv string `json:"token_env,omitempty" yaml:"token_env,omitempty"`
	// Path is the local directory path for local backend.
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
	// Timeout is the timeout for backend operations.
	Timeout time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

// Validate validates the backend configuration.
func (c *BackendConfig) Validate() error {
	switch c.Type {
	case "s3":
		if c.Bucket == "" {
			return errors.New("bucket is required for S3 backend")
		}
	case "gcs":
		if c.Bucket == "" {
			return errors.New("bucket is required for GCS backend")
		}
	case "http":
		if c.URL == "" {
			return errors.New("url is required for HTTP backend")
		}
	case "local":
		if c.Path == "" {
			return errors.New("path is required for local backend")
		}
	default:
		return errors.New("unknown backend type: " + c.Type)
	}
	return nil
}

// NewBackend creates a new cache backend based on the configuration.
func NewBackend(ctx context.Context, config BackendConfig) (CacheBackend, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	switch config.Type {
	case "s3":
		return NewS3Backend(ctx, config)
	case "gcs":
		return NewGCSBackend(ctx, config)
	case "http":
		return NewHTTPBackend(config)
	case "local":
		return NewLocalBackend(config)
	default:
		return nil, errors.New("unknown backend type: " + config.Type)
	}
}
