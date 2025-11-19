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
	"errors"
	"sort"

	"github.com/bufbuild/buf/private/bufpkg/bufmodule"
)

// ErrCyclicDependency is returned when a cycle is detected in the dependency graph.
var ErrCyclicDependency = errors.New("cyclic dependency detected in module graph")

// DependencyGraph represents a directed acyclic graph of module dependencies.
// It enables topological sorting for parallel processing of modules.
type DependencyGraph struct {
	nodes map[string]*graphNode
}

// graphNode represents a single node in the dependency graph.
type graphNode struct {
	module   bufmodule.Module
	deps     []*graphNode
	depCount int // Number of unprocessed dependencies
}

// NewDependencyGraph creates a new dependency graph from a set of modules.
func NewDependencyGraph(modules []bufmodule.Module) (*DependencyGraph, error) {
	graph := &DependencyGraph{
		nodes: make(map[string]*graphNode),
	}

	// Create nodes for all modules
	for _, module := range modules {
		graph.nodes[module.OpaqueID()] = &graphNode{
			module:   module,
			deps:     make([]*graphNode, 0),
			depCount: 0,
		}
	}

	// Build dependency edges
	for _, module := range modules {
		node := graph.nodes[module.OpaqueID()]
		deps, err := module.ModuleDeps()
		if err != nil {
			return nil, err
		}

		for _, dep := range deps {
			depID := dep.OpaqueID()
			if depNode, ok := graph.nodes[depID]; ok {
				// This module depends on depNode
				node.deps = append(node.deps, depNode)
				node.depCount++
			}
			// If the dependency is not in our node set (external dependency),
			// we ignore it as it doesn't affect our parallel processing
		}
	}

	return graph, nil
}

// TopologicalLevels returns modules grouped by topological levels.
// Modules in the same level can be processed in parallel as they have no
// dependencies on each other.
// Returns nil if a cycle is detected.
func (g *DependencyGraph) TopologicalLevels() ([][]bufmodule.Module, error) {
	if len(g.nodes) == 0 {
		return nil, nil
	}

	// Create a working copy of dependency counts
	depCounts := make(map[string]int)
	for id, node := range g.nodes {
		depCounts[id] = node.depCount
	}

	var levels [][]bufmodule.Module
	processed := make(map[string]bool)

	for len(processed) < len(g.nodes) {
		// Find nodes with no unprocessed dependencies
		var level []bufmodule.Module
		var levelIDs []string

		for id, node := range g.nodes {
			if !processed[id] && depCounts[id] == 0 {
				level = append(level, node.module)
				levelIDs = append(levelIDs, id)
			}
		}

		if len(level) == 0 {
			// No nodes with zero dependencies found, but we still have unprocessed nodes
			// This indicates a cycle
			return nil, ErrCyclicDependency
		}

		// Sort modules in this level by OpaqueID for deterministic ordering
		sort.Slice(level, func(i, j int) bool {
			return level[i].OpaqueID() < level[j].OpaqueID()
		})

		levels = append(levels, level)

		// Mark as processed and decrement dependency counts
		for _, id := range levelIDs {
			processed[id] = true
		}

		// Decrement dependency counts for nodes that depend on the processed nodes
		for _, processedID := range levelIDs {
			for _, node := range g.nodes {
				if processed[node.module.OpaqueID()] {
					continue
				}
				for _, dep := range node.deps {
					if dep.module.OpaqueID() == processedID {
						depCounts[node.module.OpaqueID()]--
					}
				}
			}
		}
	}

	return levels, nil
}

// HasCycle checks if the dependency graph contains any cycles.
func (g *DependencyGraph) HasCycle() bool {
	_, err := g.TopologicalLevels()
	return errors.Is(err, ErrCyclicDependency)
}

// GetDependencies returns the direct dependencies of a module.
func (g *DependencyGraph) GetDependencies(opaqueID string) []bufmodule.Module {
	node, ok := g.nodes[opaqueID]
	if !ok {
		return nil
	}

	deps := make([]bufmodule.Module, len(node.deps))
	for i, depNode := range node.deps {
		deps[i] = depNode.module
	}
	return deps
}

// GetDependents returns modules that depend on the given module.
func (g *DependencyGraph) GetDependents(opaqueID string) []bufmodule.Module {
	var dependents []bufmodule.Module
	for _, node := range g.nodes {
		for _, dep := range node.deps {
			if dep.module.OpaqueID() == opaqueID {
				dependents = append(dependents, node.module)
				break
			}
		}
	}
	return dependents
}

// ModuleCount returns the total number of modules in the graph.
func (g *DependencyGraph) ModuleCount() int {
	return len(g.nodes)
}
