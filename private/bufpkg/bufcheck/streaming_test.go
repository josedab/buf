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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStreamingLintClient(t *testing.T) {
	t.Parallel()

	// Note: This test just verifies construction.
	// Full integration tests would require a mock Client.
	client := NewStreamingLintClient(nil, nil)
	assert.NotNil(t, client)
}

func TestNewStreamingBreakingClient(t *testing.T) {
	t.Parallel()

	client := NewStreamingBreakingClient(nil, nil)
	assert.NotNil(t, client)
}

func TestLintStreamProgress(t *testing.T) {
	t.Parallel()

	progress := LintStreamProgress{
		TotalFiles:      100,
		ProcessedFiles:  50,
		AnnotationCount: 5,
		CurrentFile:     "test.proto",
	}

	assert.Equal(t, 100, progress.TotalFiles)
	assert.Equal(t, 50, progress.ProcessedFiles)
	assert.Equal(t, 5, progress.AnnotationCount)
	assert.Equal(t, "test.proto", progress.CurrentFile)
}
