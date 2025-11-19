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

// Package bufbenchmark provides comprehensive benchmarks for buf operations.
package bufbenchmark

import (
	"context"
	"log/slog"
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

// TestData holds pre-loaded test data for benchmarks.
type TestData struct {
	Image       bufimage.Image
	LintConfig  bufconfig.LintConfig
	Workspace   bufworkspace.Workspace
	WasmRuntime wasm.Runtime
	CheckClient bufcheck.Client
}

// LoadTestData loads test data from the specified testdata directory.
func LoadTestData(b *testing.B, size string) *TestData {
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

	// Get the root module's opaque ID
	opaqueID := getRootOpaqueID(b, workspace)

	moduleSet, err := workspace.WithTargetOpaqueIDs(opaqueID)
	require.NoError(b, err)

	moduleReadBucket := bufmodule.ModuleSetToModuleReadBucketWithOnlyProtoFiles(moduleSet)
	image, err := bufimage.BuildImage(
		ctx,
		logger,
		moduleReadBucket,
	)
	require.NoError(b, err)

	lintConfig := workspace.GetLintConfigForOpaqueID(opaqueID)
	require.NotNil(b, lintConfig)

	wasmRuntime, err := wasm.NewRuntime(ctx)
	require.NoError(b, err)

	client, err := bufcheck.NewClient(
		logger,
		bufcheck.ClientWithRunnerProvider(bufcheck.NewLocalRunnerProvider(wasmRuntime)),
	)
	require.NoError(b, err)

	b.Cleanup(func() {
		require.NoError(b, wasmRuntime.Close(ctx))
	})

	return &TestData{
		Image:       image,
		LintConfig:  lintConfig,
		Workspace:   workspace,
		WasmRuntime: wasmRuntime,
		CheckClient: client,
	}
}

// getRootOpaqueID returns the opaque ID of the root module in the workspace.
func getRootOpaqueID(b *testing.B, workspace bufworkspace.Workspace) string {
	b.Helper()

	for _, module := range workspace.Modules() {
		if module.IsTarget() {
			return module.OpaqueID()
		}
	}

	// Fallback to "." for single-module workspaces
	return "."
}

// buildImageForDirectory builds a buf image from the specified directory.
func buildImageForDirectory(b *testing.B, dirPath string, logger *slog.Logger) bufimage.Image {
	b.Helper()

	ctx := context.Background()
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

	opaqueID := getRootOpaqueID(b, workspace)
	moduleSet, err := workspace.WithTargetOpaqueIDs(opaqueID)
	require.NoError(b, err)

	moduleReadBucket := bufmodule.ModuleSetToModuleReadBucketWithOnlyProtoFiles(moduleSet)
	image, err := bufimage.BuildImage(
		ctx,
		logger,
		moduleReadBucket,
	)
	require.NoError(b, err)

	return image
}
