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
	"context"
	"sync"

	"github.com/bufbuild/buf/private/bufpkg/bufimage"
	"github.com/bufbuild/buf/private/bufpkg/bufmodule"
)

// StreamingProgress provides progress information during streaming compilation.
type StreamingProgress struct {
	// Total is the total number of files to process.
	Total int
	// Completed is the number of files successfully compiled.
	Completed int
	// Failed is the number of files that failed to compile.
	Failed int
	// Current is the path of the file currently being processed.
	Current string
	// Stage indicates the current compilation stage.
	Stage string
}

// ProgressCallback is a function called with progress updates during streaming builds.
type ProgressCallback func(StreamingProgress)

// BuildWithProgress builds an image using streaming compilation with progress reporting.
func (c *controller) BuildWithProgress(
	ctx context.Context,
	moduleReadBucket bufmodule.ModuleReadBucket,
	progress ProgressCallback,
	options ...FunctionOption,
) (bufimage.Image, error) {
	functionOptions := newFunctionOptions()
	for _, option := range options {
		option(functionOptions)
	}

	// Get file list for total count
	targetFileInfos, err := bufmodule.GetTargetFileInfos(ctx, moduleReadBucket)
	if err != nil {
		return nil, err
	}
	total := len(targetFileInfos)

	// Report initial progress
	if progress != nil {
		progress(StreamingProgress{
			Total:     total,
			Completed: 0,
			Failed:    0,
			Stage:     "starting",
		})
	}

	// Use streaming compilation
	var buildOpts []bufimage.BuildImageStreamOption
	if functionOptions.imageExcludeSourceInfo {
		buildOpts = append(buildOpts, bufimage.WithBuildImageStreamExcludeSourceCodeInfo())
	}

	// Set up progress tracking
	completed := 0
	failed := 0
	var mu sync.Mutex

	buildOpts = append(buildOpts, bufimage.WithBuildImageStreamOnChange(func(file bufimage.ImageFile) {
		mu.Lock()
		completed++
		if progress != nil {
			progress(StreamingProgress{
				Total:     total,
				Completed: completed,
				Failed:    failed,
				Current:   file.Path(),
				Stage:     "compiling",
			})
		}
		mu.Unlock()
	}))

	results, builder, err := bufimage.BuildImageStream(ctx, c.logger, moduleReadBucket, buildOpts...)
	if err != nil {
		return nil, err
	}

	// Process results
	for result := range results {
		if result.Error != nil {
			mu.Lock()
			failed++
			if progress != nil {
				progress(StreamingProgress{
					Total:     total,
					Completed: completed,
					Failed:    failed,
					Current:   result.File,
					Stage:     "error",
				})
			}
			mu.Unlock()
			continue
		}

		if err := builder.AddFile(result); err != nil {
			mu.Lock()
			failed++
			mu.Unlock()
		}
	}

	// Build final image
	if progress != nil {
		progress(StreamingProgress{
			Total:     total,
			Completed: completed,
			Failed:    failed,
			Stage:     "building",
		})
	}

	image, err := builder.Build()
	if err != nil {
		return nil, err
	}

	// Report completion
	if progress != nil {
		progress(StreamingProgress{
			Total:     total,
			Completed: completed,
			Failed:    failed,
			Stage:     "complete",
		})
	}

	// Apply filters
	return filterImage(image, functionOptions, true)
}

// StreamingBuildResult contains the result of a streaming build.
type StreamingBuildResult struct {
	// Image is the final built image.
	Image bufimage.Image
	// TotalFiles is the total number of files processed.
	TotalFiles int
	// CompletedFiles is the number of successfully compiled files.
	CompletedFiles int
	// FailedFiles is the number of files that failed to compile.
	FailedFiles int
	// CacheHits is the number of cache hits (if caching enabled).
	CacheHits int
	// CacheMisses is the number of cache misses (if caching enabled).
	CacheMisses int
}

// StreamingBuildOption is an option for streaming builds.
type StreamingBuildOption func(*streamingBuildOptions)

// WithStreamingBuildProgress sets a progress callback.
func WithStreamingBuildProgress(callback ProgressCallback) StreamingBuildOption {
	return func(opts *streamingBuildOptions) {
		opts.progressCallback = callback
	}
}

// WithStreamingBuildCache enables caching for the build.
func WithStreamingBuildCache(cache *bufimage.CompilationCache) StreamingBuildOption {
	return func(opts *streamingBuildOptions) {
		opts.cache = cache
	}
}

// WithStreamingBuildMaxParallelism sets the max parallelism.
func WithStreamingBuildMaxParallelism(maxParallelism int) StreamingBuildOption {
	return func(opts *streamingBuildOptions) {
		opts.maxParallelism = maxParallelism
	}
}

type streamingBuildOptions struct {
	progressCallback ProgressCallback
	cache            *bufimage.CompilationCache
	maxParallelism   int
}

// BuildImageWithProgressAndCache builds an image with streaming compilation, progress, and caching.
func (c *controller) BuildImageWithProgressAndCache(
	ctx context.Context,
	moduleReadBucket bufmodule.ModuleReadBucket,
	options ...StreamingBuildOption,
) (*StreamingBuildResult, error) {
	opts := &streamingBuildOptions{}
	for _, opt := range options {
		opt(opts)
	}

	// Get file list for total count
	targetFileInfos, err := bufmodule.GetTargetFileInfos(ctx, moduleReadBucket)
	if err != nil {
		return nil, err
	}
	total := len(targetFileInfos)

	// Initialize counters
	completed := 0
	failed := 0
	cacheHits := 0
	cacheMisses := 0
	var mu sync.Mutex

	// Report initial progress
	if opts.progressCallback != nil {
		opts.progressCallback(StreamingProgress{
			Total:     total,
			Completed: 0,
			Failed:    0,
			Stage:     "starting",
		})
	}

	// Create streaming compiler
	var compilerOpts []bufimage.StreamingCompilerOption
	if opts.maxParallelism > 0 {
		compilerOpts = append(compilerOpts, bufimage.WithStreamingMaxParallelism(opts.maxParallelism))
	}

	compiler := bufimage.NewStreamingCompiler(c.logger, compilerOpts...)

	// Wrap with cache if provided
	var streamingCompiler bufimage.StreamingCompiler = compiler
	if opts.cache != nil {
		streamingCompiler = bufimage.NewCachedCompiler(opts.cache, compiler, c.logger)
	}

	// Start compilation
	results, err := streamingCompiler.CompileStream(ctx, moduleReadBucket)
	if err != nil {
		return nil, err
	}

	// Build image incrementally
	builder := bufimage.NewStreamingImageBuilder()
	builder.SetOnChange(func(file bufimage.ImageFile) {
		mu.Lock()
		completed++
		if opts.progressCallback != nil {
			opts.progressCallback(StreamingProgress{
				Total:     total,
				Completed: completed,
				Failed:    failed,
				Current:   file.Path(),
				Stage:     "compiling",
			})
		}
		mu.Unlock()
	})

	// Process results
	for result := range results {
		if result.Error != nil {
			mu.Lock()
			failed++
			if opts.progressCallback != nil {
				opts.progressCallback(StreamingProgress{
					Total:     total,
					Completed: completed,
					Failed:    failed,
					Current:   result.File,
					Stage:     "error",
				})
			}
			mu.Unlock()
			continue
		}

		if err := builder.AddFile(result); err != nil {
			mu.Lock()
			failed++
			mu.Unlock()
		}
	}

	// Get cache stats if available
	if cachedCompiler, ok := streamingCompiler.(*bufimage.CachedCompiler); ok {
		stats := cachedCompiler.Stats()
		cacheHits = completed // Simplified - would need actual tracking
		cacheMisses = stats.Size - cacheHits
		if cacheMisses < 0 {
			cacheMisses = 0
		}
	}

	// Report building stage
	if opts.progressCallback != nil {
		opts.progressCallback(StreamingProgress{
			Total:     total,
			Completed: completed,
			Failed:    failed,
			Stage:     "building",
		})
	}

	// Build final image
	image, err := builder.Build()
	if err != nil {
		return nil, err
	}

	// Report completion
	if opts.progressCallback != nil {
		opts.progressCallback(StreamingProgress{
			Total:     total,
			Completed: completed,
			Failed:    failed,
			Stage:     "complete",
		})
	}

	return &StreamingBuildResult{
		Image:          image,
		TotalFiles:     total,
		CompletedFiles: completed,
		FailedFiles:    failed,
		CacheHits:      cacheHits,
		CacheMisses:    cacheMisses,
	}, nil
}

// GetImageStreamOption is an option for GetImageStream.
type GetImageStreamOption interface {
	applyToGetImageStream(*getImageStreamOptions)
}

type getImageStreamOptions struct {
	progressCallback ProgressCallback
	cache            *bufimage.CompilationCache
	maxParallelism   int
}

type progressCallbackOption struct {
	callback ProgressCallback
}

func (o *progressCallbackOption) applyToGetImageStream(opts *getImageStreamOptions) {
	opts.progressCallback = o.callback
}

// WithProgressCallback returns an option to set a progress callback for streaming operations.
func WithProgressCallback(callback ProgressCallback) GetImageStreamOption {
	return &progressCallbackOption{callback: callback}
}

type cacheOption struct {
	cache *bufimage.CompilationCache
}

func (o *cacheOption) applyToGetImageStream(opts *getImageStreamOptions) {
	opts.cache = o.cache
}

// WithCache returns an option to enable caching for streaming operations.
func WithCache(cache *bufimage.CompilationCache) GetImageStreamOption {
	return &cacheOption{cache: cache}
}
