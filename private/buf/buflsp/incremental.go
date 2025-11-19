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

// This file implements enhanced incremental compilation with dependency tracking.

package buflsp

import (
	"sync"
)

// DependencyGraph tracks file dependencies for incremental compilation.
// This allows us to only recompile affected files when a file changes.
type DependencyGraph struct {
	mu sync.RWMutex
	// dependencies maps a file path to the files it depends on (imports)
	dependencies map[string][]string
	// dependents maps a file path to the files that depend on it
	dependents map[string][]string
}

// NewDependencyGraph creates a new dependency graph.
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		dependencies: make(map[string][]string),
		dependents:   make(map[string][]string),
	}
}

// SetDependencies updates the dependencies for a file.
// This should be called after parsing to record what files a given file imports.
func (g *DependencyGraph) SetDependencies(path string, deps []string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Remove old dependent relationships
	if oldDeps, ok := g.dependencies[path]; ok {
		for _, oldDep := range oldDeps {
			g.removeDependentUnsafe(oldDep, path)
		}
	}

	// Set new dependencies
	g.dependencies[path] = deps

	// Update dependent relationships
	for _, dep := range deps {
		g.addDependentUnsafe(dep, path)
	}
}

// GetDependents returns all files that depend on the given file.
// This is used to determine which files need to be recompiled when a file changes.
func (g *DependencyGraph) GetDependents(path string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if deps, ok := g.dependents[path]; ok {
		// Return a copy to avoid race conditions
		result := make([]string, len(deps))
		copy(result, deps)
		return result
	}
	return nil
}

// GetDependencies returns all files that the given file depends on.
func (g *DependencyGraph) GetDependencies(path string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if deps, ok := g.dependencies[path]; ok {
		// Return a copy to avoid race conditions
		result := make([]string, len(deps))
		copy(result, deps)
		return result
	}
	return nil
}

// GetTransitiveDependents returns all files that directly or transitively depend on the given file.
// This performs a breadth-first search through the dependency graph.
func (g *DependencyGraph) GetTransitiveDependents(path string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	visited := make(map[string]bool)
	var result []string
	queue := []string{path}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		deps, ok := g.dependents[current]
		if !ok {
			continue
		}

		for _, dep := range deps {
			if !visited[dep] {
				visited[dep] = true
				result = append(result, dep)
				queue = append(queue, dep)
			}
		}
	}

	return result
}

// RemoveFile removes a file from the dependency graph.
func (g *DependencyGraph) RemoveFile(path string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Remove from dependencies
	if deps, ok := g.dependencies[path]; ok {
		for _, dep := range deps {
			g.removeDependentUnsafe(dep, path)
		}
		delete(g.dependencies, path)
	}

	// Remove from dependents
	delete(g.dependents, path)
}

// addDependentUnsafe adds a dependent to a file without locking.
// Caller must hold the write lock.
func (g *DependencyGraph) addDependentUnsafe(path, dependent string) {
	deps := g.dependents[path]
	// Check if already exists
	for _, d := range deps {
		if d == dependent {
			return
		}
	}
	g.dependents[path] = append(deps, dependent)
}

// removeDependentUnsafe removes a dependent from a file without locking.
// Caller must hold the write lock.
func (g *DependencyGraph) removeDependentUnsafe(path, dependent string) {
	deps := g.dependents[path]
	for i, d := range deps {
		if d == dependent {
			// Remove by swapping with last and truncating
			deps[i] = deps[len(deps)-1]
			g.dependents[path] = deps[:len(deps)-1]
			return
		}
	}
}

// Clear removes all entries from the dependency graph.
func (g *DependencyGraph) Clear() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.dependencies = make(map[string][]string)
	g.dependents = make(map[string][]string)
}

// Size returns the number of files tracked in the dependency graph.
func (g *DependencyGraph) Size() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return len(g.dependencies)
}
