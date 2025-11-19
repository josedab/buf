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

package bufimage

import (
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestNewStreamingCompiler(t *testing.T) {
	t.Parallel()

	compiler := NewStreamingCompiler(nil)
	assert.NotNil(t, compiler)
}

func TestNewStreamingCompilerWithOptions(t *testing.T) {
	t.Parallel()

	compiler := NewStreamingCompiler(
		nil,
		WithStreamingExcludeSourceCodeInfo(),
		WithStreamingMaxParallelism(4),
	)
	assert.NotNil(t, compiler)
}

func TestStreamingImageBuilder(t *testing.T) {
	t.Parallel()

	builder := NewStreamingImageBuilder()
	assert.NotNil(t, builder)
	assert.Equal(t, 0, builder.FileCount())
}

func TestStreamingImageBuilderAddFile(t *testing.T) {
	t.Parallel()

	builder := NewStreamingImageBuilder()

	// Create a simple file descriptor proto
	fdp := &descriptorpb.FileDescriptorProto{
		Name:   stringPtr("test.proto"),
		Syntax: stringPtr("proto3"),
	}

	result := CompileResult{
		File:                "test.proto",
		FileDescriptorProto: fdp,
		ModuleFullName:      nil,
		CommitID:            uuid.Nil,
		ExternalPath:        "test.proto",
		LocalPath:           "/path/to/test.proto",
		IsImport:            false,
		IsSyntaxUnspecified: false,
	}

	err := builder.AddFile(result)
	require.NoError(t, err)
	assert.Equal(t, 1, builder.FileCount())

	// Verify file can be retrieved
	file := builder.GetFile("test.proto")
	assert.NotNil(t, file)
	assert.Equal(t, "test.proto", file.Path())
}

func TestStreamingImageBuilderOnChange(t *testing.T) {
	t.Parallel()

	builder := NewStreamingImageBuilder()

	var called bool
	var mu sync.Mutex
	builder.SetOnChange(func(file ImageFile) {
		mu.Lock()
		called = true
		mu.Unlock()
	})

	fdp := &descriptorpb.FileDescriptorProto{
		Name:   stringPtr("test.proto"),
		Syntax: stringPtr("proto3"),
	}

	result := CompileResult{
		File:                "test.proto",
		FileDescriptorProto: fdp,
	}

	err := builder.AddFile(result)
	require.NoError(t, err)

	mu.Lock()
	assert.True(t, called)
	mu.Unlock()
}

func TestStreamingImageBuilderBuild(t *testing.T) {
	t.Parallel()

	builder := NewStreamingImageBuilder()

	// Add multiple files
	for i, name := range []string{"a.proto", "b.proto", "c.proto"} {
		fdp := &descriptorpb.FileDescriptorProto{
			Name:   stringPtr(name),
			Syntax: stringPtr("proto3"),
		}

		result := CompileResult{
			File:                name,
			FileDescriptorProto: fdp,
			IsImport:            i > 0, // First is target, rest are imports
		}

		err := builder.AddFile(result)
		require.NoError(t, err)
	}

	assert.Equal(t, 3, builder.FileCount())

	// Build image
	image, err := builder.Build()
	require.NoError(t, err)
	assert.NotNil(t, image)
	assert.Equal(t, 3, len(image.Files()))
}

func TestStreamingImageBuilderError(t *testing.T) {
	t.Parallel()

	builder := NewStreamingImageBuilder()

	// Result with error should return error
	result := CompileResult{
		File:  "test.proto",
		Error: assert.AnError,
	}

	err := builder.AddFile(result)
	assert.Error(t, err)
}

func TestStreamingImageBuilderNilDescriptor(t *testing.T) {
	t.Parallel()

	builder := NewStreamingImageBuilder()

	// Result with nil descriptor should be skipped
	result := CompileResult{
		File:                "test.proto",
		FileDescriptorProto: nil,
	}

	err := builder.AddFile(result)
	require.NoError(t, err)
	assert.Equal(t, 0, builder.FileCount())
}

func TestStreamingImageBuilderConcurrency(t *testing.T) {
	t.Parallel()

	builder := NewStreamingImageBuilder()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			name := stringPtr("test_" + string(rune('a'+idx%26)) + ".proto")
			fdp := &descriptorpb.FileDescriptorProto{
				Name:   name,
				Syntax: stringPtr("proto3"),
			}

			result := CompileResult{
				File:                *name,
				FileDescriptorProto: fdp,
			}

			_ = builder.AddFile(result)
		}(i)
	}

	wg.Wait()

	// Should have added some files (may have duplicates due to same names)
	assert.Greater(t, builder.FileCount(), 0)
}

func TestCompileResultFields(t *testing.T) {
	t.Parallel()

	result := CompileResult{
		File:                    "test.proto",
		FileDescriptorProto:     nil,
		ModuleFullName:          nil,
		CommitID:                uuid.New(),
		ExternalPath:            "external/test.proto",
		LocalPath:               "/local/test.proto",
		IsImport:                true,
		IsSyntaxUnspecified:     true,
		UnusedDependencyIndexes: []int32{1, 2, 3},
		Error:                   nil,
	}

	assert.Equal(t, "test.proto", result.File)
	assert.Nil(t, result.FileDescriptorProto)
	assert.NotEqual(t, uuid.Nil, result.CommitID)
	assert.Equal(t, "external/test.proto", result.ExternalPath)
	assert.Equal(t, "/local/test.proto", result.LocalPath)
	assert.True(t, result.IsImport)
	assert.True(t, result.IsSyntaxUnspecified)
	assert.Equal(t, []int32{1, 2, 3}, result.UnusedDependencyIndexes)
	assert.Nil(t, result.Error)
}

func stringPtr(s string) *string {
	return &s
}
