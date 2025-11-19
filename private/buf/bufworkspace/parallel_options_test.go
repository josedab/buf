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
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultParallelOptions(t *testing.T) {
	t.Parallel()

	opts := DefaultParallelOptions()
	assert.Equal(t, 0, opts.MaxWorkers)
	assert.True(t, opts.Enabled)
	assert.True(t, opts.CancelOnFailure)
}

func TestEffectiveMaxWorkers(t *testing.T) {
	t.Parallel()

	// Default (0) should return NumCPU
	opts := DefaultParallelOptions()
	assert.Equal(t, runtime.NumCPU(), opts.EffectiveMaxWorkers())

	// Explicit value should be returned
	opts.MaxWorkers = 4
	assert.Equal(t, 4, opts.EffectiveMaxWorkers())

	// Negative values should also return NumCPU
	opts.MaxWorkers = -1
	assert.Equal(t, runtime.NumCPU(), opts.EffectiveMaxWorkers())
}

func TestWithMaxWorkers(t *testing.T) {
	t.Parallel()

	opts := ApplyOptions(WithMaxWorkers(8))
	assert.Equal(t, 8, opts.MaxWorkers)
	assert.Equal(t, 8, opts.EffectiveMaxWorkers())
}

func TestWithParallelEnabled(t *testing.T) {
	t.Parallel()

	// Disable parallel
	opts := ApplyOptions(WithParallelEnabled(false))
	assert.False(t, opts.Enabled)

	// Enable parallel
	opts = ApplyOptions(WithParallelEnabled(true))
	assert.True(t, opts.Enabled)
}

func TestWithCancelOnFailure(t *testing.T) {
	t.Parallel()

	// Disable cancel on failure
	opts := ApplyOptions(WithCancelOnFailure(false))
	assert.False(t, opts.CancelOnFailure)

	// Enable cancel on failure
	opts = ApplyOptions(WithCancelOnFailure(true))
	assert.True(t, opts.CancelOnFailure)
}

func TestApplyMultipleOptions(t *testing.T) {
	t.Parallel()

	opts := ApplyOptions(
		WithMaxWorkers(16),
		WithParallelEnabled(false),
		WithCancelOnFailure(false),
	)

	assert.Equal(t, 16, opts.MaxWorkers)
	assert.False(t, opts.Enabled)
	assert.False(t, opts.CancelOnFailure)
}

func TestOptionsOverwrite(t *testing.T) {
	t.Parallel()

	// Later options should overwrite earlier ones
	opts := ApplyOptions(
		WithMaxWorkers(4),
		WithMaxWorkers(8),
	)

	assert.Equal(t, 8, opts.MaxWorkers)
}
