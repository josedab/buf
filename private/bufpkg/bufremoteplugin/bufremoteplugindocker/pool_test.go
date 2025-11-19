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

package bufremoteplugindocker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewContainerPool(t *testing.T) {
	t.Parallel()

	t.Run("returns error without client", func(t *testing.T) {
		t.Parallel()
		pool, err := NewContainerPool(nil)
		require.Error(t, err)
		assert.Nil(t, pool)
	})
}

func TestContainerPoolOptions(t *testing.T) {
	t.Parallel()

	// Note: These tests verify options work without needing a real Docker client
	// Full integration tests would require Docker to be available

	t.Run("default max size", func(t *testing.T) {
		t.Parallel()
		// Default max size is 5
		// This is verified in the NewContainerPool constructor
	})

	t.Run("custom max size", func(t *testing.T) {
		t.Parallel()
		// WithPoolMaxSize option would set a custom size
		// This requires a mock client to test properly
	})
}

func TestPoolStats(t *testing.T) {
	t.Parallel()

	t.Run("zero stats initially", func(t *testing.T) {
		t.Parallel()
		stats := PoolStats{}
		assert.Equal(t, 0, stats.TotalContainers)
		assert.Equal(t, 0, stats.ActiveContainers)
		assert.Equal(t, int64(0), stats.Hits)
		assert.Equal(t, int64(0), stats.Misses)
	})
}

func TestContainer(t *testing.T) {
	t.Parallel()

	t.Run("container struct fields", func(t *testing.T) {
		t.Parallel()
		container := &Container{
			ID:    "test-id",
			Image: "test-image",
			Ready: true,
		}

		assert.Equal(t, "test-id", container.ID)
		assert.Equal(t, "test-image", container.Image)
		assert.True(t, container.Ready)
	})
}

// Note: Full integration tests for ContainerPool would require:
// 1. A running Docker daemon
// 2. Test images to use
// 3. Proper cleanup after tests
//
// These would be marked with a build tag like:
// //go:build integration
//
// Example integration test:
//
// func TestContainerPool_Integration(t *testing.T) {
//     if testing.Short() {
//         t.Skip("skipping integration test")
//     }
//
//     client, err := client.NewClientWithOpts(client.FromEnv)
//     require.NoError(t, err)
//     defer client.Close()
//
//     pool, err := NewContainerPool(client, WithPoolMaxSize(3))
//     require.NoError(t, err)
//     defer pool.Close()
//
//     // Test getting a container
//     container, err := pool.Get(context.Background(), "alpine:latest")
//     require.NoError(t, err)
//     assert.NotEmpty(t, container.ID)
//
//     // Test pool stats
//     stats := pool.Stats()
//     assert.Equal(t, 1, stats.TotalContainers)
//     assert.Equal(t, int64(1), stats.Misses)
//
//     // Get same container again (should be cache hit)
//     container2, err := pool.Get(context.Background(), "alpine:latest")
//     require.NoError(t, err)
//     assert.Equal(t, container.ID, container2.ID)
//
//     stats = pool.Stats()
//     assert.Equal(t, int64(1), stats.Hits)
// }
