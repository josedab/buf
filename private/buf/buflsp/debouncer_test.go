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

package buflsp

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.lsp.dev/protocol"
)

func TestDebouncerBasic(t *testing.T) {
	t.Parallel()
	d := newDebouncer()

	var called int32
	uri := protocol.URI("file:///test.proto")

	d.Debounce(uri, 10*time.Millisecond, func() {
		atomic.AddInt32(&called, 1)
	})

	// Wait for debounce to fire
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, int32(1), atomic.LoadInt32(&called))
}

func TestDebouncerCancelsOnRapidCalls(t *testing.T) {
	t.Parallel()
	d := newDebouncer()

	var called int32
	uri := protocol.URI("file:///test.proto")

	// Make several rapid calls
	for i := 0; i < 5; i++ {
		d.Debounce(uri, 50*time.Millisecond, func() {
			atomic.AddInt32(&called, 1)
		})
		time.Sleep(10 * time.Millisecond)
	}

	// Wait for debounce to fire
	time.Sleep(100 * time.Millisecond)

	// Only the last call should have executed
	assert.Equal(t, int32(1), atomic.LoadInt32(&called))
}

func TestDebouncerMultipleURIs(t *testing.T) {
	t.Parallel()
	d := newDebouncer()

	var called1, called2 int32
	uri1 := protocol.URI("file:///test1.proto")
	uri2 := protocol.URI("file:///test2.proto")

	d.Debounce(uri1, 10*time.Millisecond, func() {
		atomic.AddInt32(&called1, 1)
	})

	d.Debounce(uri2, 10*time.Millisecond, func() {
		atomic.AddInt32(&called2, 1)
	})

	// Wait for debounces to fire
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, int32(1), atomic.LoadInt32(&called1))
	assert.Equal(t, int32(1), atomic.LoadInt32(&called2))
}

func TestDebouncerCancel(t *testing.T) {
	t.Parallel()
	d := newDebouncer()

	var called int32
	uri := protocol.URI("file:///test.proto")

	d.Debounce(uri, 50*time.Millisecond, func() {
		atomic.AddInt32(&called, 1)
	})

	// Cancel before it fires
	d.Cancel(uri)

	// Wait to make sure it doesn't fire
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, int32(0), atomic.LoadInt32(&called))
}

func TestDebouncerCancelAll(t *testing.T) {
	t.Parallel()
	d := newDebouncer()

	var called1, called2 int32
	uri1 := protocol.URI("file:///test1.proto")
	uri2 := protocol.URI("file:///test2.proto")

	d.Debounce(uri1, 50*time.Millisecond, func() {
		atomic.AddInt32(&called1, 1)
	})

	d.Debounce(uri2, 50*time.Millisecond, func() {
		atomic.AddInt32(&called2, 1)
	})

	// Cancel all before they fire
	d.CancelAll()

	// Wait to make sure they don't fire
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, int32(0), atomic.LoadInt32(&called1))
	assert.Equal(t, int32(0), atomic.LoadInt32(&called2))
}

func TestDebouncerCancelNonexistent(t *testing.T) {
	t.Parallel()
	d := newDebouncer()

	// Should not panic
	d.Cancel(protocol.URI("file:///nonexistent.proto"))
}
