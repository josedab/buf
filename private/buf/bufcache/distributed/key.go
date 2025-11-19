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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
)

// CacheKey represents a unique identifier for a cached build result.
type CacheKey struct {
	// ContentHash is the SHA-256 hash of all input files.
	ContentHash string `json:"content_hash"`
	// Version is the buf CLI version.
	Version string `json:"version"`
	// Options contains the compilation options used.
	Options CompileOptions `json:"options"`
}

// CompileOptions represents the options used during compilation
// that affect the output.
type CompileOptions struct {
	// ExcludeSourceInfo excludes source code info from the image.
	ExcludeSourceInfo bool `json:"exclude_source_info,omitempty"`
	// ExcludeSourceRetentionOptions excludes source-retention options.
	ExcludeSourceRetentionOptions bool `json:"exclude_source_retention_options,omitempty"`
	// TargetPaths are the specific paths to target.
	TargetPaths []string `json:"target_paths,omitempty"`
	// ExcludePaths are paths to exclude.
	ExcludePaths []string `json:"exclude_paths,omitempty"`
}

// File represents a file to be included in cache key generation.
type File struct {
	// Path is the normalized path of the file.
	Path string
	// Content is the file content.
	Content []byte
}

// GenerateCacheKey creates a deterministic cache key from the given files and options.
func GenerateCacheKey(files []File, version string, options CompileOptions) CacheKey {
	h := sha256.New()

	// Sort files for determinism
	sortedFiles := make([]File, len(files))
	copy(sortedFiles, files)
	sort.Slice(sortedFiles, func(i, j int) bool {
		return sortedFiles[i].Path < sortedFiles[j].Path
	})

	for _, file := range sortedFiles {
		h.Write([]byte(file.Path))
		h.Write([]byte{0}) // null separator
		h.Write(file.Content)
		h.Write([]byte{0}) // null separator
	}

	// Include options in the hash for full determinism
	optionsJSON, _ := json.Marshal(options)
	h.Write(optionsJSON)

	return CacheKey{
		ContentHash: hex.EncodeToString(h.Sum(nil)),
		Version:     version,
		Options:     options,
	}
}

// String returns a string representation of the cache key suitable for use as a filename.
func (k CacheKey) String() string {
	return k.ContentHash
}

// FullString returns a full string representation including version and options hash.
func (k CacheKey) FullString() string {
	optionsJSON, _ := json.Marshal(k.Options)
	optionsHash := sha256.Sum256(optionsJSON)
	return k.Version + "/" + k.ContentHash + "/" + hex.EncodeToString(optionsHash[:8])
}
