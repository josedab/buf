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
)

// GCSBackend implements CacheBackend using Google Cloud Storage.
// NOTE: This is a stub implementation. To use GCS backend, add the following dependencies:
//   - cloud.google.com/go/storage
//   - google.golang.org/api/iterator
type GCSBackend struct {
	bucket string
	prefix string
}

// ErrGCSNotAvailable is returned when GCS support is not available.
var ErrGCSNotAvailable = errors.New("GCS backend is not available: cloud.google.com/go/storage dependency not installed")

// NewGCSBackend creates a new GCS cache backend.
// NOTE: This is a stub that returns an error. To use GCS backend,
// rebuild with GCS dependencies installed.
func NewGCSBackend(ctx context.Context, backendConfig BackendConfig) (*GCSBackend, error) {
	return nil, ErrGCSNotAvailable
}

// Get retrieves data from GCS for the given key.
func (b *GCSBackend) Get(ctx context.Context, key CacheKey) ([]byte, error) {
	return nil, ErrGCSNotAvailable
}

// Put stores data in GCS with the given key.
func (b *GCSBackend) Put(ctx context.Context, key CacheKey, data []byte) error {
	return ErrGCSNotAvailable
}

// Delete removes data from GCS for the given key.
func (b *GCSBackend) Delete(ctx context.Context, key CacheKey) error {
	return ErrGCSNotAvailable
}

// Stats returns cache statistics.
func (b *GCSBackend) Stats(ctx context.Context) (*CacheStats, error) {
	return nil, ErrGCSNotAvailable
}

// Close closes the GCS client.
func (b *GCSBackend) Close() error {
	return nil
}
