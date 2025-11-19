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

package bufbenchmark

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/bufbuild/buf/private/buf/buftarget"
	"github.com/bufbuild/buf/private/buf/bufworkspace"
	"github.com/bufbuild/buf/private/bufpkg/bufcheck"
	"github.com/bufbuild/buf/private/bufpkg/bufconfig"
	"github.com/bufbuild/buf/private/bufpkg/bufimage"
	"github.com/bufbuild/buf/private/bufpkg/bufmodule"
	"github.com/bufbuild/buf/private/bufpkg/bufplugin"
	"github.com/bufbuild/buf/private/pkg/slogtestext"
	"github.com/bufbuild/buf/private/pkg/storage/storageos"
	"github.com/bufbuild/buf/private/pkg/wasm"
	"github.com/stretchr/testify/require"
)

// Compilation benchmarks

func BenchmarkCompileSmall(b *testing.B) {
	benchmarkCompile(b, "small")
}

func BenchmarkCompileMedium(b *testing.B) {
	benchmarkCompile(b, "medium")
}

func BenchmarkCompileLarge(b *testing.B) {
	benchmarkCompile(b, "large")
}

// Linting benchmarks

func BenchmarkLintSmall(b *testing.B) {
	benchmarkLint(b, "small")
}

func BenchmarkLintMedium(b *testing.B) {
	benchmarkLint(b, "medium")
}

func BenchmarkLintLarge(b *testing.B) {
	benchmarkLint(b, "large")
}

// Breaking change benchmarks

func BenchmarkBreakingSmall(b *testing.B) {
	benchmarkBreaking(b, "small")
}

func BenchmarkBreakingMedium(b *testing.B) {
	benchmarkBreaking(b, "medium")
}

func BenchmarkBreakingLarge(b *testing.B) {
	benchmarkBreaking(b, "large")
}

// Image operation benchmarks

func BenchmarkImageSerialize(b *testing.B) {
	benchmarkImageSerialize(b, "medium")
}

func BenchmarkImageDeserialize(b *testing.B) {
	benchmarkImageDeserialize(b, "medium")
}

func BenchmarkImageFilter(b *testing.B) {
	benchmarkImageFilter(b, "medium")
}

// benchmarkCompile benchmarks the compilation of protobuf files.
func benchmarkCompile(b *testing.B, size string) {
	b.Helper()

	ctx := context.Background()
	dirPath := filepath.Join("testdata", size)
	logger := slogtestext.NewLogger(b)

	storageosProvider := storageos.NewProvider(storageos.ProviderWithSymlinks())

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		readWriteBucket, err := storageosProvider.NewReadWriteBucket(
			dirPath,
			storageos.ReadWriteBucketWithSymlinksIfSupported(),
		)
		require.NoError(b, err)

		bucketTargeting, err := buftarget.NewBucketTargeting(
			ctx,
			logger,
			readWriteBucket,
			".",
			nil,
			nil,
			buftarget.TerminateAtControllingWorkspace,
		)
		require.NoError(b, err)

		workspace, err := bufworkspace.NewWorkspaceProvider(
			logger,
			bufmodule.NopGraphProvider,
			bufmodule.NopModuleDataProvider,
			bufmodule.NopCommitProvider,
			bufplugin.NopPluginKeyProvider,
		).GetWorkspaceForBucket(
			ctx,
			readWriteBucket,
			bucketTargeting,
		)
		require.NoError(b, err)

		opaqueID := getRootOpaqueIDForBenchmark(b, workspace)
		moduleSet, err := workspace.WithTargetOpaqueIDs(opaqueID)
		require.NoError(b, err)

		moduleReadBucket := bufmodule.ModuleSetToModuleReadBucketWithOnlyProtoFiles(moduleSet)
		_, err = bufimage.BuildImage(
			ctx,
			logger,
			moduleReadBucket,
		)
		require.NoError(b, err)
	}
}

// benchmarkLint benchmarks the linting of protobuf files.
func benchmarkLint(b *testing.B, size string) {
	b.Helper()

	ctx := context.Background()
	data := LoadTestData(b, size)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Lint returns an error if there are lint issues
		// We ignore the error here since we're just measuring performance
		_ = data.CheckClient.Lint(
			ctx,
			data.LintConfig,
			data.Image,
		)
	}
}

// benchmarkBreaking benchmarks breaking change detection.
func benchmarkBreaking(b *testing.B, size string) {
	b.Helper()

	ctx := context.Background()
	data := LoadTestData(b, size)
	logger := slogtestext.NewLogger(b)

	// Get breaking config from workspace
	opaqueID := getRootOpaqueIDForBenchmark(b, data.Workspace)
	breakingConfig := data.Workspace.GetBreakingConfigForOpaqueID(opaqueID)
	require.NotNil(b, breakingConfig)

	// For breaking changes, we compare the image against itself
	// In a real scenario, this would be against a previous version
	againstImage := cloneImage(b, data.Image, logger)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = data.CheckClient.Breaking(
			ctx,
			breakingConfig,
			data.Image,
			againstImage,
		)
	}
}

// benchmarkImageSerialize benchmarks image serialization to proto format.
func benchmarkImageSerialize(b *testing.B, size string) {
	b.Helper()

	data := LoadTestData(b, size)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := bufimage.ImageToProtoImage(data.Image)
		require.NoError(b, err)
	}
}

// benchmarkImageDeserialize benchmarks image deserialization from proto format.
func benchmarkImageDeserialize(b *testing.B, size string) {
	b.Helper()

	data := LoadTestData(b, size)

	// First serialize to get proto format
	protoImage, err := bufimage.ImageToProtoImage(data.Image)
	require.NoError(b, err)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := bufimage.NewImageForProto(protoImage)
		require.NoError(b, err)
	}
}

// benchmarkImageFilter benchmarks image filtering operations.
func benchmarkImageFilter(b *testing.B, size string) {
	b.Helper()

	data := LoadTestData(b, size)

	// Get some paths to filter by
	files := data.Image.Files()
	if len(files) == 0 {
		b.Skip("No files in image")
		return
	}

	// Filter to first file only
	filterPaths := []string{files[0].Path()}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := bufimage.ImageWithOnlyPathsAllowNotExist(data.Image, filterPaths, nil)
		require.NoError(b, err)
	}
}

// getRootOpaqueIDForBenchmark returns the opaque ID of the root module in the workspace.
func getRootOpaqueIDForBenchmark(b *testing.B, workspace bufworkspace.Workspace) string {
	b.Helper()

	for _, module := range workspace.Modules() {
		if module.IsTarget() {
			return module.OpaqueID()
		}
	}

	return "."
}

// cloneImage creates a deep copy of an image for comparison purposes.
func cloneImage(b *testing.B, image bufimage.Image, _ interface{}) bufimage.Image {
	b.Helper()

	// Serialize and deserialize to create a deep copy
	protoImage, err := bufimage.ImageToProtoImage(image)
	require.NoError(b, err)

	clonedImage, err := bufimage.NewImageForProto(protoImage)
	require.NoError(b, err)

	return clonedImage
}

// loadBreakingConfig loads breaking change configuration.
func loadBreakingConfig(b *testing.B, size string) bufconfig.BreakingConfig {
	b.Helper()

	ctx := context.Background()
	dirPath := filepath.Join("testdata", size)
	logger := slogtestext.NewLogger(b)

	storageosProvider := storageos.NewProvider(storageos.ProviderWithSymlinks())
	readWriteBucket, err := storageosProvider.NewReadWriteBucket(
		dirPath,
		storageos.ReadWriteBucketWithSymlinksIfSupported(),
	)
	require.NoError(b, err)

	bucketTargeting, err := buftarget.NewBucketTargeting(
		ctx,
		logger,
		readWriteBucket,
		".",
		nil,
		nil,
		buftarget.TerminateAtControllingWorkspace,
	)
	require.NoError(b, err)

	workspace, err := bufworkspace.NewWorkspaceProvider(
		logger,
		bufmodule.NopGraphProvider,
		bufmodule.NopModuleDataProvider,
		bufmodule.NopCommitProvider,
		bufplugin.NopPluginKeyProvider,
	).GetWorkspaceForBucket(
		ctx,
		readWriteBucket,
		bucketTargeting,
	)
	require.NoError(b, err)

	opaqueID := getRootOpaqueIDForBenchmark(b, workspace)
	breakingConfig := workspace.GetBreakingConfigForOpaqueID(opaqueID)
	require.NotNil(b, breakingConfig)

	return breakingConfig
}

// Helper to create a check client for benchmarks
func newCheckClient(b *testing.B) bufcheck.Client {
	b.Helper()

	ctx := context.Background()
	logger := slogtestext.NewLogger(b)

	wasmRuntime, err := wasm.NewRuntime(ctx)
	require.NoError(b, err)

	b.Cleanup(func() {
		require.NoError(b, wasmRuntime.Close(ctx))
	})

	client, err := bufcheck.NewClient(
		logger,
		bufcheck.ClientWithRunnerProvider(bufcheck.NewLocalRunnerProvider(wasmRuntime)),
	)
	require.NoError(b, err)

	return client
}
