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

package bufgen

import (
	"context"
	"sort"
	"sync"

	"github.com/bufbuild/buf/private/bufpkg/bufconfig"
	"github.com/bufbuild/buf/private/bufpkg/bufimage"
	"github.com/bufbuild/buf/private/pkg/thread"
)

// PluginGroup represents a group of plugins that can be executed in parallel.
// Plugins within a group have different output directories and don't conflict.
type PluginGroup struct {
	Plugins []bufconfig.GeneratePluginConfig
}

// GroupPluginsByOutput groups plugins by their output directory.
// Plugins with the same output directory must be executed sequentially
// to maintain insertion point ordering. Plugins with different output
// directories can be executed in parallel.
func GroupPluginsByOutput(plugins []bufconfig.GeneratePluginConfig) []PluginGroup {
	if len(plugins) == 0 {
		return nil
	}

	// Group plugins by output directory
	outToPlugins := make(map[string][]bufconfig.GeneratePluginConfig)
	var outOrder []string

	for _, plugin := range plugins {
		out := plugin.Out()
		if _, exists := outToPlugins[out]; !exists {
			outOrder = append(outOrder, out)
		}
		outToPlugins[out] = append(outToPlugins[out], plugin)
	}

	// Sort output directories for deterministic ordering
	sort.Strings(outOrder)

	// Create plugin groups
	groups := make([]PluginGroup, len(outOrder))
	for i, out := range outOrder {
		groups[i] = PluginGroup{
			Plugins: outToPlugins[out],
		}
	}

	return groups
}

// ParallelGenerateResult contains the result of generating for a single image.
type ParallelGenerateResult struct {
	// ImageIndex is the index of the image in the input slice.
	ImageIndex int
	// Err is the error returned by the generate operation, if any.
	Err error
}

// ParallelGenerator provides parallel generation capabilities.
type ParallelGenerator struct {
	generator Generator
}

// NewParallelGenerator creates a new ParallelGenerator.
func NewParallelGenerator(generator Generator) *ParallelGenerator {
	return &ParallelGenerator{
		generator: generator,
	}
}

// GenerateImages generates code for multiple images in parallel.
// Each image is processed independently, and results are collected.
func (g *ParallelGenerator) GenerateImages(
	ctx context.Context,
	images []bufimage.Image,
	generateFunc func(ctx context.Context, image bufimage.Image) error,
) ([]ParallelGenerateResult, error) {
	if len(images) == 0 {
		return nil, nil
	}

	results := make([]ParallelGenerateResult, len(images))
	var mu sync.Mutex

	jobs := make([]func(context.Context) error, len(images))
	for i, image := range images {
		i, image := i, image
		jobs[i] = func(ctx context.Context) error {
			err := generateFunc(ctx, image)

			mu.Lock()
			results[i] = ParallelGenerateResult{
				ImageIndex: i,
				Err:        err,
			}
			mu.Unlock()

			// Return error to stop on failure (optional)
			return err
		}
	}

	if err := thread.Parallelize(ctx, jobs, thread.ParallelizeWithCancelOnFailure()); err != nil {
		// Return results collected so far along with the error
		return results, err
	}

	return results, nil
}

// PluginExecutionPlan represents a plan for executing plugins in parallel.
type PluginExecutionPlan struct {
	// ParallelGroups contains groups of plugins that can be executed in parallel.
	// Within each group, plugins have no conflicts and can run concurrently.
	ParallelGroups [][]bufconfig.GeneratePluginConfig
}

// CreatePluginExecutionPlan creates an execution plan that maximizes parallelism
// while respecting plugin dependencies and output directory conflicts.
func CreatePluginExecutionPlan(plugins []bufconfig.GeneratePluginConfig) PluginExecutionPlan {
	groups := GroupPluginsByOutput(plugins)

	if len(groups) == 0 {
		return PluginExecutionPlan{}
	}

	// Find the maximum number of plugins in any group
	maxLen := 0
	for _, group := range groups {
		if len(group.Plugins) > maxLen {
			maxLen = len(group.Plugins)
		}
	}

	// Create parallel execution batches
	// Each batch contains one plugin from each group (at the same index)
	// This allows maximum parallelism while maintaining order within groups
	parallelGroups := make([][]bufconfig.GeneratePluginConfig, maxLen)

	for i := 0; i < maxLen; i++ {
		var batch []bufconfig.GeneratePluginConfig
		for _, group := range groups {
			if i < len(group.Plugins) {
				batch = append(batch, group.Plugins[i])
			}
		}
		parallelGroups[i] = batch
	}

	return PluginExecutionPlan{
		ParallelGroups: parallelGroups,
	}
}

// ExecutePlan executes a plugin execution plan.
// Each parallel group is executed concurrently, but groups are processed sequentially.
func ExecutePlan(
	ctx context.Context,
	plan PluginExecutionPlan,
	executeFunc func(ctx context.Context, plugin bufconfig.GeneratePluginConfig) error,
) error {
	for _, batch := range plan.ParallelGroups {
		jobs := make([]func(context.Context) error, len(batch))
		for i, plugin := range batch {
			plugin := plugin
			jobs[i] = func(ctx context.Context) error {
				return executeFunc(ctx, plugin)
			}
		}

		if err := thread.Parallelize(ctx, jobs, thread.ParallelizeWithCancelOnFailure()); err != nil {
			return err
		}
	}

	return nil
}

// ParallelPluginStats contains statistics about parallel plugin execution.
type ParallelPluginStats struct {
	// TotalPlugins is the total number of plugins.
	TotalPlugins int
	// ParallelGroups is the number of parallel execution groups.
	ParallelGroups int
	// MaxParallelism is the maximum number of plugins that can run in parallel.
	MaxParallelism int
	// UniqueOutputDirs is the number of unique output directories.
	UniqueOutputDirs int
}

// GetParallelStats returns statistics about the parallel execution plan.
func GetParallelStats(plugins []bufconfig.GeneratePluginConfig) ParallelPluginStats {
	plan := CreatePluginExecutionPlan(plugins)
	groups := GroupPluginsByOutput(plugins)

	maxParallel := 0
	for _, batch := range plan.ParallelGroups {
		if len(batch) > maxParallel {
			maxParallel = len(batch)
		}
	}

	return ParallelPluginStats{
		TotalPlugins:     len(plugins),
		ParallelGroups:   len(plan.ParallelGroups),
		MaxParallelism:   maxParallel,
		UniqueOutputDirs: len(groups),
	}
}
