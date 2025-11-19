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

package bufcheck

import (
	"context"
	"errors"
	"sync"

	"github.com/bufbuild/buf/private/bufpkg/bufanalysis"
	"github.com/bufbuild/buf/private/bufpkg/bufconfig"
	"github.com/bufbuild/buf/private/bufpkg/bufimage"
	"github.com/bufbuild/buf/private/pkg/thread"
)

// ModuleLintConfig pairs a module's image with its lint configuration.
type ModuleLintConfig struct {
	// Image is the compiled image for the module.
	Image bufimage.Image
	// Config is the lint configuration for the module.
	Config bufconfig.LintConfig
	// ModuleID is an identifier for the module (e.g., OpaqueID).
	ModuleID string
}

// ModuleBreakingConfig pairs a module's images with its breaking configuration.
type ModuleBreakingConfig struct {
	// Image is the compiled image for the module.
	Image bufimage.Image
	// AgainstImage is the image to compare against for breaking changes.
	AgainstImage bufimage.Image
	// Config is the breaking configuration for the module.
	Config bufconfig.BreakingConfig
	// ModuleID is an identifier for the module (e.g., OpaqueID).
	ModuleID string
}

// ParallelLintResult contains the result of linting a single module.
type ParallelLintResult struct {
	// ModuleID is the identifier of the module.
	ModuleID string
	// Err is the error returned by the lint operation, if any.
	// This may be a bufanalysis.FileAnnotationSet for lint failures.
	Err error
}

// ParallelBreakingResult contains the result of breaking detection for a single module.
type ParallelBreakingResult struct {
	// ModuleID is the identifier of the module.
	ModuleID string
	// Err is the error returned by the breaking operation, if any.
	// This may be a bufanalysis.FileAnnotationSet for breaking failures.
	Err error
}

// LintParallel lints multiple modules in parallel using the given client.
// Returns results for all modules, including any errors.
// If any module has lint failures, the error will be of type bufanalysis.FileAnnotationSet.
func LintParallel(
	ctx context.Context,
	client Client,
	modules []ModuleLintConfig,
	options ...LintOption,
) ([]ParallelLintResult, error) {
	if len(modules) == 0 {
		return nil, nil
	}

	results := make([]ParallelLintResult, len(modules))
	var mu sync.Mutex

	jobs := make([]func(context.Context) error, len(modules))
	for i, module := range modules {
		i, module := i, module // Capture for goroutine
		jobs[i] = func(ctx context.Context) error {
			err := client.Lint(ctx, module.Config, module.Image, options...)

			mu.Lock()
			results[i] = ParallelLintResult{
				ModuleID: module.ModuleID,
				Err:      err,
			}
			mu.Unlock()

			// Don't return error here - we want to continue linting other modules
			// and collect all results
			return nil
		}
	}

	// Execute all lint operations in parallel
	if err := thread.Parallelize(ctx, jobs); err != nil {
		return nil, err
	}

	return results, nil
}

// LintParallelWithMergedAnnotations lints multiple modules in parallel and merges
// all file annotations into a single FileAnnotationSet.
// This is useful when you want a single consolidated error with all lint issues.
func LintParallelWithMergedAnnotations(
	ctx context.Context,
	client Client,
	modules []ModuleLintConfig,
	options ...LintOption,
) error {
	results, err := LintParallel(ctx, client, modules, options...)
	if err != nil {
		return err
	}

	// Collect all file annotations from results
	var allAnnotations []bufanalysis.FileAnnotation
	for _, result := range results {
		if result.Err != nil {
			var annotationSet bufanalysis.FileAnnotationSet
			if errors.As(result.Err, &annotationSet) {
				allAnnotations = append(allAnnotations, annotationSet.FileAnnotations()...)
			} else {
				// Non-annotation error
				return result.Err
			}
		}
	}

	if len(allAnnotations) > 0 {
		return bufanalysis.NewFileAnnotationSet(allAnnotations...)
	}

	return nil
}

// BreakingParallel checks multiple modules for breaking changes in parallel.
// Returns results for all modules, including any errors.
func BreakingParallel(
	ctx context.Context,
	client Client,
	modules []ModuleBreakingConfig,
	options ...BreakingOption,
) ([]ParallelBreakingResult, error) {
	if len(modules) == 0 {
		return nil, nil
	}

	results := make([]ParallelBreakingResult, len(modules))
	var mu sync.Mutex

	jobs := make([]func(context.Context) error, len(modules))
	for i, module := range modules {
		i, module := i, module // Capture for goroutine
		jobs[i] = func(ctx context.Context) error {
			err := client.Breaking(ctx, module.Config, module.Image, module.AgainstImage, options...)

			mu.Lock()
			results[i] = ParallelBreakingResult{
				ModuleID: module.ModuleID,
				Err:      err,
			}
			mu.Unlock()

			return nil
		}
	}

	if err := thread.Parallelize(ctx, jobs); err != nil {
		return nil, err
	}

	return results, nil
}

// BreakingParallelWithMergedAnnotations checks multiple modules for breaking changes
// and merges all file annotations into a single FileAnnotationSet.
func BreakingParallelWithMergedAnnotations(
	ctx context.Context,
	client Client,
	modules []ModuleBreakingConfig,
	options ...BreakingOption,
) error {
	results, err := BreakingParallel(ctx, client, modules, options...)
	if err != nil {
		return err
	}

	var allAnnotations []bufanalysis.FileAnnotation
	for _, result := range results {
		if result.Err != nil {
			var annotationSet bufanalysis.FileAnnotationSet
			if errors.As(result.Err, &annotationSet) {
				allAnnotations = append(allAnnotations, annotationSet.FileAnnotations()...)
			} else {
				return result.Err
			}
		}
	}

	if len(allAnnotations) > 0 {
		return bufanalysis.NewFileAnnotationSet(allAnnotations...)
	}

	return nil
}
