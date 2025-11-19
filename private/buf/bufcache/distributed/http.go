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
	"fmt"
	"io"
	"net/http"
	"os"
	"sync/atomic"
	"time"
)

// HTTPBackend implements CacheBackend using an HTTP server.
type HTTPBackend struct {
	baseURL string
	client  *http.Client
	token   string
	hits    atomic.Int64
	misses  atomic.Int64
}

// NewHTTPBackend creates a new HTTP cache backend.
func NewHTTPBackend(backendConfig BackendConfig) (*HTTPBackend, error) {
	timeout := backendConfig.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	var token string
	if backendConfig.TokenEnv != "" {
		token = os.Getenv(backendConfig.TokenEnv)
	}

	return &HTTPBackend{
		baseURL: backendConfig.URL,
		client: &http.Client{
			Timeout: timeout,
		},
		token: token,
	}, nil
}

// Get retrieves data from the HTTP server for the given key.
func (b *HTTPBackend) Get(ctx context.Context, key CacheKey) ([]byte, error) {
	url := fmt.Sprintf("%s/cache/%s/%s", b.baseURL, key.Version, key.ContentHash)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if b.token != "" {
		req.Header.Set("Authorization", "Bearer "+b.token)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		b.misses.Add(1)
		return nil, ErrCacheMiss
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	b.hits.Add(1)
	return data, nil
}

// Put stores data on the HTTP server with the given key.
func (b *HTTPBackend) Put(ctx context.Context, key CacheKey, data []byte) error {
	url := fmt.Sprintf("%s/cache/%s/%s", b.baseURL, key.Version, key.ContentHash)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if b.token != "" {
		req.Header.Set("Authorization", "Bearer "+b.token)
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := b.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// Delete removes data from the HTTP server for the given key.
func (b *HTTPBackend) Delete(ctx context.Context, key CacheKey) error {
	url := fmt.Sprintf("%s/cache/%s/%s", b.baseURL, key.Version, key.ContentHash)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if b.token != "" {
		req.Header.Set("Authorization", "Bearer "+b.token)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// Stats returns cache statistics.
// Note: HTTP backend stats are limited to local hit/miss counters.
// Full stats require the server to implement a /stats endpoint.
func (b *HTTPBackend) Stats(ctx context.Context) (*CacheStats, error) {
	stats := &CacheStats{
		Hits:   b.hits.Load(),
		Misses: b.misses.Load(),
	}

	// Try to get full stats from server
	url := fmt.Sprintf("%s/stats", b.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return stats, nil
	}

	if b.token != "" {
		req.Header.Set("Authorization", "Bearer "+b.token)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return stats, nil
	}
	defer resp.Body.Close()

	// If stats endpoint not available, return local stats
	if resp.StatusCode != http.StatusOK {
		return stats, nil
	}

	// Could parse server stats here if needed
	return stats, nil
}
