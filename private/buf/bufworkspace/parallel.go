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

package bufworkspace

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/bufbuild/buf/private/bufpkg/bufimage"
	"github.com/bufbuild/buf/private/bufpkg/bufmodule"
	"github.com/bufbuild/buf/private/pkg/thread"
)

// ParallelCompiler compiles modules in parallel while respecting dependencies.
type ParallelCompiler struct {
	logger   *slog.Logger
	options  ParallelOptions
	reporter ProgressReporter
}

// NewParallelCompiler creates a new ParallelCompiler.
func NewParallelCompiler(
	logger *slog.Logger,
	reporter ProgressReporter,
	opts ...ParallelOption,
) *ParallelCompiler {
	options := ApplyOptions(opts...)
	if reporter == nil {
		reporter = NoopProgressReporter{}
	}
	return &ParallelCompiler{
		logger:   logger,
		options:  options,
		reporter: reporter,
	}
}

// ModuleImage pairs a module with its compiled image.
type ModuleImage struct {
	Module bufmodule.Module
	Image  bufimage.Image
}

// CompileModules compiles the given modules in parallel, respecting dependencies.
// Returns a slice of ModuleImage in the same order as the input modules.
func (c *ParallelCompiler) CompileModules(
	ctx context.Context,
	modules []bufmodule.Module,
) ([]ModuleImage, error) {
	if len(modules) == 0 {
		return nil, nil
	}

	// If parallel is disabled, process sequentially
	if !c.options.Enabled {
		return c.compileSequential(ctx, modules)
	}

	// Build dependency graph
	graph, err := NewDependencyGraph(modules)
	if err != nil {
		return nil, err
	}

	// Get topological levels
	levels, err := graph.TopologicalLevels()
	if err != nil {
		return nil, err
	}

	// Create result map for thread-safe storage
	resultMap := make(map[string]ModuleImage)
	var resultMu sync.Mutex

	// Process each level in parallel
	for levelIdx, level := range levels {
		c.reporter.LevelStarted(levelIdx, level)
		levelStart := time.Now()

		if err := c.compileLevel(ctx, level, resultMap, &resultMu); err != nil {
			return nil, err
		}

		c.reporter.LevelCompleted(levelIdx, level, time.Since(levelStart))
	}

	// Reconstruct results in input order
	results := make([]ModuleImage, len(modules))
	for i, module := range modules {
		results[i] = resultMap[module.OpaqueID()]
	}

	return results, nil
}

// compileLevel compiles all modules in a level in parallel.
func (c *ParallelCompiler) compileLevel(
	ctx context.Context,
	modules []bufmodule.Module,
	resultMap map[string]ModuleImage,
	resultMu *sync.Mutex,
) error {
	jobs := make([]func(context.Context) error, len(modules))

	for i, module := range modules {
		module := module // Capture for goroutine
		jobs[i] = func(ctx context.Context) error {
			c.reporter.ModuleStarted(module)
			start := time.Now()

			image, err := c.compileModule(ctx, module)
			if err != nil {
				c.reporter.ModuleFailed(module, err)
				return err
			}

			resultMu.Lock()
			resultMap[module.OpaqueID()] = ModuleImage{
				Module: module,
				Image:  image,
			}
			resultMu.Unlock()

			c.reporter.ModuleCompleted(module, time.Since(start))
			return nil
		}
	}

	var parallelOpts []thread.ParallelizeOption
	if c.options.CancelOnFailure {
		parallelOpts = append(parallelOpts, thread.ParallelizeWithCancelOnFailure())
	}

	return thread.Parallelize(ctx, jobs, parallelOpts...)
}

// compileModule compiles a single module to an image.
func (c *ParallelCompiler) compileModule(
	ctx context.Context,
	module bufmodule.Module,
) (bufimage.Image, error) {
	// Convert module to read bucket with only proto files
	moduleReadBucket := bufmodule.ModuleReadBucketWithOnlyProtoFiles(module)

	// Build image from the module
	return bufimage.BuildImage(
		ctx,
		c.logger,
		moduleReadBucket,
	)
}

// compileSequential compiles modules sequentially (fallback mode).
func (c *ParallelCompiler) compileSequential(
	ctx context.Context,
	modules []bufmodule.Module,
) ([]ModuleImage, error) {
	results := make([]ModuleImage, len(modules))

	for i, module := range modules {
		c.reporter.ModuleStarted(module)
		start := time.Now()

		image, err := c.compileModule(ctx, module)
		if err != nil {
			c.reporter.ModuleFailed(module, err)
			return nil, err
		}

		results[i] = ModuleImage{
			Module: module,
			Image:  image,
		}
		c.reporter.ModuleCompleted(module, time.Since(start))
	}

	return results, nil
}

// CompileWorkspace compiles all target modules in a workspace in parallel.
func CompileWorkspace(
	ctx context.Context,
	logger *slog.Logger,
	workspace Workspace,
	reporter ProgressReporter,
	opts ...ParallelOption,
) ([]ModuleImage, error) {
	modules := bufmodule.ModuleSetTargetModules(workspace)
	compiler := NewParallelCompiler(logger, reporter, opts...)
	return compiler.CompileModules(ctx, modules)
}

// CompileModuleSet compiles all modules in a module set in parallel.
func CompileModuleSet(
	ctx context.Context,
	logger *slog.Logger,
	moduleSet bufmodule.ModuleSet,
	reporter ProgressReporter,
	opts ...ParallelOption,
) ([]ModuleImage, error) {
	modules := moduleSet.Modules()
	compiler := NewParallelCompiler(logger, reporter, opts...)
	return compiler.CompileModules(ctx, modules)
}
