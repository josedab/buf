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

// This file implements debouncing for real-time diagnostics.

package buflsp

import (
	"sync"
	"time"

	"go.lsp.dev/protocol"
)

// debouncer manages debounced function calls for URIs.
// This is used to prevent excessive diagnostics computation during rapid typing.
type debouncer struct {
	mu     sync.Mutex
	timers map[protocol.URI]*time.Timer
}

// newDebouncer creates a new debouncer instance.
func newDebouncer() *debouncer {
	return &debouncer{
		timers: make(map[protocol.URI]*time.Timer),
	}
}

// Debounce schedules a function to be called after the specified delay.
// If called again for the same URI before the delay expires, the previous
// call is cancelled and a new one is scheduled.
func (d *debouncer) Debounce(uri protocol.URI, delay time.Duration, fn func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Cancel any existing timer for this URI
	if timer, ok := d.timers[uri]; ok {
		timer.Stop()
	}

	// Schedule a new timer
	d.timers[uri] = time.AfterFunc(delay, func() {
		d.mu.Lock()
		delete(d.timers, uri)
		d.mu.Unlock()
		fn()
	})
}

// Cancel cancels any pending debounced call for the given URI.
func (d *debouncer) Cancel(uri protocol.URI) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if timer, ok := d.timers[uri]; ok {
		timer.Stop()
		delete(d.timers, uri)
	}
}

// CancelAll cancels all pending debounced calls.
func (d *debouncer) CancelAll() {
	d.mu.Lock()
	defer d.mu.Unlock()

	for uri, timer := range d.timers {
		timer.Stop()
		delete(d.timers, uri)
	}
}
