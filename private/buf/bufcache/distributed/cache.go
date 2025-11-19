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
	"log/slog"

	"github.com/bufbuild/buf/private/bufpkg/bufimage"
	imagev1 "github.com/bufbuild/buf/private/gen/proto/go/buf/alpha/image/v1"
	"google.golang.org/protobuf/proto"
)

// Cache provides distributed caching functionality for buf images.
type Cache struct {
	backend CacheBackend
	version string
	logger  *slog.Logger
}

// NewCache creates a new distributed cache.
func NewCache(backend CacheBackend, version string, logger *slog.Logger) *Cache {
	if logger == nil {
		logger = slog.Default()
	}
	return &Cache{
		backend: backend,
		version: version,
		logger:  logger,
	}
}

// GetImage retrieves a cached image for the given files and options.
// Returns ErrCacheMiss if the image is not in the cache.
func (c *Cache) GetImage(ctx context.Context, files []File, options CompileOptions) (bufimage.Image, error) {
	key := GenerateCacheKey(files, c.version, options)

	data, err := c.backend.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	image, err := deserializeImage(data)
	if err != nil {
		c.logger.Warn("cache corruption detected", slog.String("key", key.ContentHash), slog.Any("error", err))
		return nil, ErrCacheCorruption
	}

	c.logger.Debug("cache hit", slog.String("key", key.ContentHash))
	return image, nil
}

// PutImage stores an image in the cache for the given files and options.
func (c *Cache) PutImage(ctx context.Context, files []File, options CompileOptions, image bufimage.Image) error {
	key := GenerateCacheKey(files, c.version, options)

	data, err := serializeImage(image)
	if err != nil {
		return err
	}

	if err := c.backend.Put(ctx, key, data); err != nil {
		c.logger.Warn("failed to cache image", slog.String("key", key.ContentHash), slog.Any("error", err))
		return err
	}

	c.logger.Debug("image cached", slog.String("key", key.ContentHash), slog.Int("size", len(data)))
	return nil
}

// PutImageAsync stores an image in the cache asynchronously.
// Errors are logged but not returned.
func (c *Cache) PutImageAsync(ctx context.Context, files []File, options CompileOptions, image bufimage.Image) {
	go func() {
		// Use background context since the parent context may be canceled
		bgCtx := context.Background()
		if err := c.PutImage(bgCtx, files, options, image); err != nil {
			c.logger.Warn("async cache put failed", slog.Any("error", err))
		}
	}()
}

// Stats returns cache statistics.
func (c *Cache) Stats(ctx context.Context) (*CacheStats, error) {
	return c.backend.Stats(ctx)
}

// Backend returns the underlying cache backend.
func (c *Cache) Backend() CacheBackend {
	return c.backend
}

// serializeImage converts an Image to bytes for storage.
func serializeImage(image bufimage.Image) ([]byte, error) {
	protoImage, err := bufimage.ImageToProtoImage(image)
	if err != nil {
		return nil, err
	}
	return proto.Marshal(protoImage)
}

// deserializeImage converts bytes back to an Image.
func deserializeImage(data []byte) (bufimage.Image, error) {
	protoImage := &imagev1.Image{}
	if err := proto.Unmarshal(data, protoImage); err != nil {
		return nil, err
	}
	return bufimage.NewImageForProto(protoImage)
}
