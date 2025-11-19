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
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"sync/atomic"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Backend implements CacheBackend using AWS S3.
type S3Backend struct {
	client *s3.Client
	bucket string
	prefix string
	hits   atomic.Int64
	misses atomic.Int64
}

// NewS3Backend creates a new S3 cache backend.
func NewS3Backend(ctx context.Context, backendConfig BackendConfig) (*S3Backend, error) {
	var opts []func(*config.LoadOptions) error
	if backendConfig.Region != "" {
		opts = append(opts, config.WithRegion(backendConfig.Region))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	return &S3Backend{
		client: client,
		bucket: backendConfig.Bucket,
		prefix: backendConfig.Prefix,
	}, nil
}

// Get retrieves data from S3 for the given key.
func (b *S3Backend) Get(ctx context.Context, key CacheKey) ([]byte, error) {
	result, err := b.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(b.keyPath(key)),
	})
	if err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			b.misses.Add(1)
			return nil, ErrCacheMiss
		}
		// Also check for NotFound status
		var notFound *types.NotFound
		if errors.As(err, &notFound) {
			b.misses.Add(1)
			return nil, ErrCacheMiss
		}
		return nil, fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object body: %w", err)
	}

	b.hits.Add(1)
	return data, nil
}

// Put stores data in S3 with the given key.
func (b *S3Backend) Put(ctx context.Context, key CacheKey, data []byte) error {
	_, err := b.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(b.keyPath(key)),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		return fmt.Errorf("failed to put object to S3: %w", err)
	}
	return nil
}

// Delete removes data from S3 for the given key.
func (b *S3Backend) Delete(ctx context.Context, key CacheKey) error {
	_, err := b.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(b.keyPath(key)),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object from S3: %w", err)
	}
	return nil
}

// Stats returns cache statistics.
func (b *S3Backend) Stats(ctx context.Context) (*CacheStats, error) {
	stats := &CacheStats{
		Hits:   b.hits.Load(),
		Misses: b.misses.Load(),
	}

	// List objects to calculate size and count
	paginator := s3.NewListObjectsV2Paginator(b.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(b.bucket),
		Prefix: aws.String(b.prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return stats, fmt.Errorf("failed to list objects: %w", err)
		}
		for _, obj := range page.Contents {
			stats.Size += *obj.Size
			stats.EntryCount++
		}
	}

	return stats, nil
}

// keyPath returns the full S3 key path for the given cache key.
func (b *S3Backend) keyPath(key CacheKey) string {
	if b.prefix == "" {
		return fmt.Sprintf("%s/%s.bin", key.Version, key.ContentHash)
	}
	return fmt.Sprintf("%s/%s/%s.bin", b.prefix, key.Version, key.ContentHash)
}
