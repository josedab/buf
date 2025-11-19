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

package distributed

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateCacheKey(t *testing.T) {
	t.Parallel()

	files := []File{
		{Path: "foo.proto", Content: []byte("syntax = \"proto3\";")},
		{Path: "bar.proto", Content: []byte("syntax = \"proto3\";")},
	}
	options := CompileOptions{}
	version := "1.0.0"

	key := GenerateCacheKey(files, version, options)

	assert.NotEmpty(t, key.ContentHash)
	assert.Equal(t, version, key.Version)
	assert.Equal(t, options, key.Options)
}

func TestGenerateCacheKeyDeterminism(t *testing.T) {
	t.Parallel()

	files := []File{
		{Path: "foo.proto", Content: []byte("syntax = \"proto3\";")},
		{Path: "bar.proto", Content: []byte("syntax = \"proto3\"; package bar;")},
	}
	options := CompileOptions{}
	version := "1.0.0"

	key1 := GenerateCacheKey(files, version, options)
	key2 := GenerateCacheKey(files, version, options)

	assert.Equal(t, key1.ContentHash, key2.ContentHash)
}

func TestGenerateCacheKeyFileOrderIndependent(t *testing.T) {
	t.Parallel()

	files1 := []File{
		{Path: "foo.proto", Content: []byte("syntax = \"proto3\";")},
		{Path: "bar.proto", Content: []byte("syntax = \"proto3\";")},
	}
	files2 := []File{
		{Path: "bar.proto", Content: []byte("syntax = \"proto3\";")},
		{Path: "foo.proto", Content: []byte("syntax = \"proto3\";")},
	}
	options := CompileOptions{}
	version := "1.0.0"

	key1 := GenerateCacheKey(files1, version, options)
	key2 := GenerateCacheKey(files2, version, options)

	assert.Equal(t, key1.ContentHash, key2.ContentHash, "cache keys should be independent of file order")
}

func TestGenerateCacheKeyDifferentContent(t *testing.T) {
	t.Parallel()

	files1 := []File{
		{Path: "foo.proto", Content: []byte("syntax = \"proto3\";")},
	}
	files2 := []File{
		{Path: "foo.proto", Content: []byte("syntax = \"proto2\";")},
	}
	options := CompileOptions{}
	version := "1.0.0"

	key1 := GenerateCacheKey(files1, version, options)
	key2 := GenerateCacheKey(files2, version, options)

	assert.NotEqual(t, key1.ContentHash, key2.ContentHash, "different content should produce different keys")
}

func TestGenerateCacheKeyDifferentOptions(t *testing.T) {
	t.Parallel()

	files := []File{
		{Path: "foo.proto", Content: []byte("syntax = \"proto3\";")},
	}
	options1 := CompileOptions{ExcludeSourceInfo: false}
	options2 := CompileOptions{ExcludeSourceInfo: true}
	version := "1.0.0"

	key1 := GenerateCacheKey(files, version, options1)
	key2 := GenerateCacheKey(files, version, options2)

	assert.NotEqual(t, key1.ContentHash, key2.ContentHash, "different options should produce different keys")
}

func TestCacheKeyString(t *testing.T) {
	t.Parallel()

	files := []File{
		{Path: "foo.proto", Content: []byte("syntax = \"proto3\";")},
	}
	options := CompileOptions{}
	version := "1.0.0"

	key := GenerateCacheKey(files, version, options)

	str := key.String()
	require.NotEmpty(t, str)
	assert.Equal(t, key.ContentHash, str)
}

func TestCacheKeyFullString(t *testing.T) {
	t.Parallel()

	files := []File{
		{Path: "foo.proto", Content: []byte("syntax = \"proto3\";")},
	}
	options := CompileOptions{}
	version := "1.0.0"

	key := GenerateCacheKey(files, version, options)

	fullStr := key.FullString()
	require.NotEmpty(t, fullStr)
	assert.Contains(t, fullStr, version)
	assert.Contains(t, fullStr, key.ContentHash)
}
