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
	"testing"

	"buf.build/go/bufplugin/check"
	"github.com/bufbuild/buf/private/bufpkg/bufanalysis"
	"github.com/bufbuild/buf/private/bufpkg/bufconfig"
	"github.com/bufbuild/buf/private/bufpkg/bufimage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLintParallelEmpty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := &mockClient{}

	results, err := LintParallel(ctx, client, nil)
	require.NoError(t, err)
	assert.Nil(t, results)
}

func TestLintParallelSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := &mockClient{
		lintFunc: func(_ context.Context, _ bufconfig.LintConfig, _ bufimage.Image, _ ...LintOption) error {
			return nil
		},
	}

	modules := []ModuleLintConfig{
		{ModuleID: "module1"},
		{ModuleID: "module2"},
		{ModuleID: "module3"},
	}

	results, err := LintParallel(ctx, client, modules)
	require.NoError(t, err)
	require.Len(t, results, 3)

	for i, result := range results {
		assert.Equal(t, modules[i].ModuleID, result.ModuleID)
		assert.NoError(t, result.Err)
	}
}

func TestLintParallelWithFailures(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create mock annotations
	annotation := &mockFileAnnotation{
		path:    "test.proto",
		message: "test error",
	}

	client := &mockClient{
		lintFunc: func(_ context.Context, _ bufconfig.LintConfig, _ bufimage.Image, _ ...LintOption) error {
			return bufanalysis.NewFileAnnotationSet(annotation)
		},
	}

	modules := []ModuleLintConfig{
		{ModuleID: "module1"},
		{ModuleID: "module2"},
	}

	results, err := LintParallel(ctx, client, modules)
	require.NoError(t, err)
	require.Len(t, results, 2)

	// All results should have annotation errors
	for _, result := range results {
		assert.Error(t, result.Err)
		var annotationSet bufanalysis.FileAnnotationSet
		assert.True(t, errors.As(result.Err, &annotationSet))
	}
}

func TestLintParallelWithMergedAnnotations(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	callCount := 0
	client := &mockClient{
		lintFunc: func(_ context.Context, _ bufconfig.LintConfig, _ bufimage.Image, _ ...LintOption) error {
			callCount++
			annotation := &mockFileAnnotation{
				path:    "test.proto",
				message: "error " + string(rune('0'+callCount)),
			}
			return bufanalysis.NewFileAnnotationSet(annotation)
		},
	}

	modules := []ModuleLintConfig{
		{ModuleID: "module1"},
		{ModuleID: "module2"},
	}

	err := LintParallelWithMergedAnnotations(ctx, client, modules)
	require.Error(t, err)

	var annotationSet bufanalysis.FileAnnotationSet
	require.True(t, errors.As(err, &annotationSet))
	assert.Len(t, annotationSet.FileAnnotations(), 2)
}

func TestLintParallelWithMergedAnnotationsSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := &mockClient{
		lintFunc: func(_ context.Context, _ bufconfig.LintConfig, _ bufimage.Image, _ ...LintOption) error {
			return nil
		},
	}

	modules := []ModuleLintConfig{
		{ModuleID: "module1"},
	}

	err := LintParallelWithMergedAnnotations(ctx, client, modules)
	assert.NoError(t, err)
}

func TestBreakingParallelEmpty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := &mockClient{}

	results, err := BreakingParallel(ctx, client, nil)
	require.NoError(t, err)
	assert.Nil(t, results)
}

func TestBreakingParallelSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := &mockClient{
		breakingFunc: func(_ context.Context, _ bufconfig.BreakingConfig, _, _ bufimage.Image, _ ...BreakingOption) error {
			return nil
		},
	}

	modules := []ModuleBreakingConfig{
		{ModuleID: "module1"},
		{ModuleID: "module2"},
	}

	results, err := BreakingParallel(ctx, client, modules)
	require.NoError(t, err)
	require.Len(t, results, 2)

	for i, result := range results {
		assert.Equal(t, modules[i].ModuleID, result.ModuleID)
		assert.NoError(t, result.Err)
	}
}

func TestBreakingParallelWithMergedAnnotations(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	client := &mockClient{
		breakingFunc: func(_ context.Context, _ bufconfig.BreakingConfig, _, _ bufimage.Image, _ ...BreakingOption) error {
			annotation := &mockFileAnnotation{
				path:    "test.proto",
				message: "breaking change",
			}
			return bufanalysis.NewFileAnnotationSet(annotation)
		},
	}

	modules := []ModuleBreakingConfig{
		{ModuleID: "module1"},
	}

	err := BreakingParallelWithMergedAnnotations(ctx, client, modules)
	require.Error(t, err)

	var annotationSet bufanalysis.FileAnnotationSet
	require.True(t, errors.As(err, &annotationSet))
}

// mockClient is a mock implementation of Client for testing
type mockClient struct {
	lintFunc     func(context.Context, bufconfig.LintConfig, bufimage.Image, ...LintOption) error
	breakingFunc func(context.Context, bufconfig.BreakingConfig, bufimage.Image, bufimage.Image, ...BreakingOption) error
}

func (c *mockClient) Lint(ctx context.Context, config bufconfig.LintConfig, image bufimage.Image, options ...LintOption) error {
	if c.lintFunc != nil {
		return c.lintFunc(ctx, config, image, options...)
	}
	return nil
}

func (c *mockClient) Breaking(ctx context.Context, config bufconfig.BreakingConfig, image bufimage.Image, againstImage bufimage.Image, options ...BreakingOption) error {
	if c.breakingFunc != nil {
		return c.breakingFunc(ctx, config, image, againstImage, options...)
	}
	return nil
}

func (c *mockClient) ConfiguredRules(_ context.Context, _ check.RuleType, _ bufconfig.CheckConfig, _ ...ConfiguredRulesOption) ([]Rule, error) {
	return nil, nil
}

func (c *mockClient) AllRules(_ context.Context, _ check.RuleType, _ bufconfig.FileVersion, _ ...AllRulesOption) ([]Rule, error) {
	return nil, nil
}

func (c *mockClient) AllCategories(_ context.Context, _ bufconfig.FileVersion, _ ...AllCategoriesOption) ([]Category, error) {
	return nil, nil
}

// mockFileAnnotation is a mock implementation of bufanalysis.FileAnnotation
type mockFileAnnotation struct {
	path    string
	message string
}

func (a *mockFileAnnotation) Path() string                     { return a.path }
func (a *mockFileAnnotation) ExternalPath() string             { return a.path }
func (a *mockFileAnnotation) StartLine() int                   { return 1 }
func (a *mockFileAnnotation) StartColumn() int                 { return 1 }
func (a *mockFileAnnotation) EndLine() int                     { return 1 }
func (a *mockFileAnnotation) EndColumn() int                   { return 1 }
func (a *mockFileAnnotation) Type() string                     { return "ERROR" }
func (a *mockFileAnnotation) Message() string                  { return a.message }
func (a *mockFileAnnotation) String() string                   { return a.message }
func (a *mockFileAnnotation) FileInfo() bufanalysis.FileInfo   { return nil }
func (a *mockFileAnnotation) WithExternalPath(string) bufanalysis.FileAnnotation { return a }
