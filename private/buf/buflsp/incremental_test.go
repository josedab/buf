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
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDependencyGraphBasic(t *testing.T) {
	t.Parallel()
	g := NewDependencyGraph()

	// Set up: a.proto imports b.proto and c.proto
	g.SetDependencies("a.proto", []string{"b.proto", "c.proto"})

	// Check dependencies
	deps := g.GetDependencies("a.proto")
	sort.Strings(deps)
	assert.Equal(t, []string{"b.proto", "c.proto"}, deps)

	// Check dependents
	assert.Equal(t, []string{"a.proto"}, g.GetDependents("b.proto"))
	assert.Equal(t, []string{"a.proto"}, g.GetDependents("c.proto"))
}

func TestDependencyGraphMultipleDependents(t *testing.T) {
	t.Parallel()
	g := NewDependencyGraph()

	// Set up:
	// a.proto imports common.proto
	// b.proto imports common.proto
	g.SetDependencies("a.proto", []string{"common.proto"})
	g.SetDependencies("b.proto", []string{"common.proto"})

	// Both should be dependents of common.proto
	deps := g.GetDependents("common.proto")
	sort.Strings(deps)
	assert.Equal(t, []string{"a.proto", "b.proto"}, deps)
}

func TestDependencyGraphUpdateDependencies(t *testing.T) {
	t.Parallel()
	g := NewDependencyGraph()

	// Initial dependencies
	g.SetDependencies("a.proto", []string{"b.proto", "c.proto"})

	// Update dependencies (remove c.proto, add d.proto)
	g.SetDependencies("a.proto", []string{"b.proto", "d.proto"})

	// Check dependencies
	deps := g.GetDependencies("a.proto")
	sort.Strings(deps)
	assert.Equal(t, []string{"b.proto", "d.proto"}, deps)

	// a.proto should no longer be a dependent of c.proto
	assert.Nil(t, g.GetDependents("c.proto"))

	// a.proto should be a dependent of d.proto
	assert.Equal(t, []string{"a.proto"}, g.GetDependents("d.proto"))
}

func TestDependencyGraphTransitiveDependents(t *testing.T) {
	t.Parallel()
	g := NewDependencyGraph()

	// Set up:
	// a.proto imports b.proto
	// b.proto imports c.proto
	// c.proto imports d.proto
	g.SetDependencies("a.proto", []string{"b.proto"})
	g.SetDependencies("b.proto", []string{"c.proto"})
	g.SetDependencies("c.proto", []string{"d.proto"})

	// Get transitive dependents of d.proto
	deps := g.GetTransitiveDependents("d.proto")
	sort.Strings(deps)
	assert.Equal(t, []string{"a.proto", "b.proto", "c.proto"}, deps)
}

func TestDependencyGraphCircularDependencies(t *testing.T) {
	t.Parallel()
	g := NewDependencyGraph()

	// Set up circular dependency:
	// a.proto imports b.proto
	// b.proto imports a.proto
	g.SetDependencies("a.proto", []string{"b.proto"})
	g.SetDependencies("b.proto", []string{"a.proto"})

	// GetTransitiveDependents should handle circular dependencies without infinite loop
	deps := g.GetTransitiveDependents("a.proto")
	assert.Contains(t, deps, "b.proto")
}

func TestDependencyGraphRemoveFile(t *testing.T) {
	t.Parallel()
	g := NewDependencyGraph()

	// Set up
	g.SetDependencies("a.proto", []string{"b.proto", "c.proto"})
	g.SetDependencies("d.proto", []string{"b.proto"})

	// Remove a.proto
	g.RemoveFile("a.proto")

	// a.proto should no longer exist
	assert.Nil(t, g.GetDependencies("a.proto"))

	// b.proto should only have d.proto as dependent now
	assert.Equal(t, []string{"d.proto"}, g.GetDependents("b.proto"))

	// c.proto should have no dependents
	assert.Nil(t, g.GetDependents("c.proto"))
}

func TestDependencyGraphClear(t *testing.T) {
	t.Parallel()
	g := NewDependencyGraph()

	// Set up
	g.SetDependencies("a.proto", []string{"b.proto"})
	g.SetDependencies("c.proto", []string{"d.proto"})

	assert.Equal(t, 2, g.Size())

	// Clear
	g.Clear()

	assert.Equal(t, 0, g.Size())
	assert.Nil(t, g.GetDependencies("a.proto"))
	assert.Nil(t, g.GetDependents("b.proto"))
}

func TestDependencyGraphSize(t *testing.T) {
	t.Parallel()
	g := NewDependencyGraph()

	assert.Equal(t, 0, g.Size())

	g.SetDependencies("a.proto", []string{"b.proto"})
	assert.Equal(t, 1, g.Size())

	g.SetDependencies("c.proto", []string{"d.proto"})
	assert.Equal(t, 2, g.Size())

	g.RemoveFile("a.proto")
	assert.Equal(t, 1, g.Size())
}

func TestDependencyGraphNoDependencies(t *testing.T) {
	t.Parallel()
	g := NewDependencyGraph()

	// File with no dependencies
	g.SetDependencies("a.proto", nil)

	assert.Nil(t, g.GetDependencies("a.proto"))
	assert.Equal(t, 1, g.Size())
}

func TestDependencyGraphNoDuplicateDependents(t *testing.T) {
	t.Parallel()
	g := NewDependencyGraph()

	// Set same dependencies twice
	g.SetDependencies("a.proto", []string{"b.proto"})
	g.SetDependencies("a.proto", []string{"b.proto"})

	// Should only have one dependent entry
	deps := g.GetDependents("b.proto")
	assert.Len(t, deps, 1)
}

func TestDependencyGraphConcurrency(t *testing.T) {
	t.Parallel()
	g := NewDependencyGraph()

	// Run concurrent operations
	done := make(chan bool)

	go func() {
		for i := 0; i < 100; i++ {
			g.SetDependencies("a.proto", []string{"b.proto", "c.proto"})
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			g.GetDependents("b.proto")
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			g.GetTransitiveDependents("c.proto")
		}
		done <- true
	}()

	// Wait for all goroutines
	<-done
	<-done
	<-done
}
