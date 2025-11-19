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
)

// ParallelOptions configures parallel processing behavior.
type ParallelOptions struct {
	// MaxWorkers is the maximum number of concurrent workers.
	// If 0, defaults to runtime.NumCPU().
	MaxWorkers int
	// Enabled controls whether parallel processing is enabled.
	// When false, modules are processed sequentially.
	Enabled bool
	// CancelOnFailure stops all processing when any module fails.
	CancelOnFailure bool
}

// DefaultParallelOptions returns the default parallel options.
// Parallel processing is enabled by default with worker count equal to CPU count.
func DefaultParallelOptions() ParallelOptions {
	return ParallelOptions{
		MaxWorkers:      0, // Will use runtime.NumCPU()
		Enabled:         true,
		CancelOnFailure: true,
	}
}

// EffectiveMaxWorkers returns the actual number of workers to use.
// If MaxWorkers is 0, returns runtime.NumCPU().
func (o ParallelOptions) EffectiveMaxWorkers() int {
	if o.MaxWorkers <= 0 {
		return runtime.NumCPU()
	}
	return o.MaxWorkers
}

// ParallelOption is an option for configuring parallel processing.
type ParallelOption func(*ParallelOptions)

// WithMaxWorkers sets the maximum number of concurrent workers.
func WithMaxWorkers(n int) ParallelOption {
	return func(o *ParallelOptions) {
		o.MaxWorkers = n
	}
}

// WithParallelEnabled sets whether parallel processing is enabled.
func WithParallelEnabled(enabled bool) ParallelOption {
	return func(o *ParallelOptions) {
		o.Enabled = enabled
	}
}

// WithCancelOnFailure sets whether to cancel all processing on first failure.
func WithCancelOnFailure(cancel bool) ParallelOption {
	return func(o *ParallelOptions) {
		o.CancelOnFailure = cancel
	}
}

// ApplyOptions applies the given options to the default parallel options.
func ApplyOptions(opts ...ParallelOption) ParallelOptions {
	options := DefaultParallelOptions()
	for _, opt := range opts {
		opt(&options)
	}
	return options
}
