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

package bufctl

import (
	"testing"

	"github.com/bufbuild/buf/private/bufpkg/bufimage"
	"github.com/stretchr/testify/assert"
)

func TestStreamingProgress(t *testing.T) {
	t.Parallel()

	progress := StreamingProgress{
		Total:     100,
		Completed: 50,
		Failed:    2,
		Current:   "test.proto",
		Stage:     "compiling",
	}

	assert.Equal(t, 100, progress.Total)
	assert.Equal(t, 50, progress.Completed)
	assert.Equal(t, 2, progress.Failed)
	assert.Equal(t, "test.proto", progress.Current)
	assert.Equal(t, "compiling", progress.Stage)
}

func TestStreamingBuildResult(t *testing.T) {
	t.Parallel()

	result := StreamingBuildResult{
		Image:          nil,
		TotalFiles:     100,
		CompletedFiles: 98,
		FailedFiles:    2,
		CacheHits:      80,
		CacheMisses:    18,
	}

	assert.Nil(t, result.Image)
	assert.Equal(t, 100, result.TotalFiles)
	assert.Equal(t, 98, result.CompletedFiles)
	assert.Equal(t, 2, result.FailedFiles)
	assert.Equal(t, 80, result.CacheHits)
	assert.Equal(t, 18, result.CacheMisses)
}

func TestStreamingBuildOptions(t *testing.T) {
	t.Parallel()

	opts := &streamingBuildOptions{}

	// Test progress callback option
	var callbackCalled bool
	WithStreamingBuildProgress(func(p StreamingProgress) {
		callbackCalled = true
	})(opts)
	assert.NotNil(t, opts.progressCallback)
	opts.progressCallback(StreamingProgress{})
	assert.True(t, callbackCalled)

	// Test cache option
	cache := bufimage.NewCompilationCache()
	WithStreamingBuildCache(cache)(opts)
	assert.Equal(t, cache, opts.cache)

	// Test max parallelism option
	WithStreamingBuildMaxParallelism(8)(opts)
	assert.Equal(t, 8, opts.maxParallelism)
}

func TestGetImageStreamOptions(t *testing.T) {
	t.Parallel()

	opts := &getImageStreamOptions{}

	// Test progress callback option
	var callbackCalled bool
	callback := func(p StreamingProgress) {
		callbackCalled = true
	}
	progressOpt := WithProgressCallback(callback)
	progressOpt.applyToGetImageStream(opts)
	assert.NotNil(t, opts.progressCallback)
	opts.progressCallback(StreamingProgress{})
	assert.True(t, callbackCalled)

	// Test cache option
	cache := bufimage.NewCompilationCache()
	cacheOpt := WithCache(cache)
	cacheOpt.applyToGetImageStream(opts)
	assert.Equal(t, cache, opts.cache)
}

func TestProgressStages(t *testing.T) {
	t.Parallel()

	stages := []string{"starting", "compiling", "error", "building", "complete"}

	for _, stage := range stages {
		progress := StreamingProgress{
			Total:     100,
			Completed: 50,
			Stage:     stage,
		}
		assert.Equal(t, stage, progress.Stage)
	}
}
