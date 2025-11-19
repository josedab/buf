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
	"testing"

	"github.com/bufbuild/buf/private/bufpkg/bufmodule"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDependencyGraphEmpty(t *testing.T) {
	t.Parallel()

	graph, err := NewDependencyGraph(nil)
	require.NoError(t, err)

	levels, err := graph.TopologicalLevels()
	require.NoError(t, err)
	assert.Nil(t, levels)
	assert.Equal(t, 0, graph.ModuleCount())
}

func TestDependencyGraphSingleModule(t *testing.T) {
	t.Parallel()

	module := newTestModule("module1", nil)
	graph, err := NewDependencyGraph([]bufmodule.Module{module})
	require.NoError(t, err)

	levels, err := graph.TopologicalLevels()
	require.NoError(t, err)
	require.Len(t, levels, 1)
	require.Len(t, levels[0], 1)
	assert.Equal(t, "module1", levels[0][0].OpaqueID())
	assert.Equal(t, 1, graph.ModuleCount())
}

func TestDependencyGraphIndependentModules(t *testing.T) {
	t.Parallel()

	// Three independent modules should all be in the same level
	modules := []bufmodule.Module{
		newTestModule("module1", nil),
		newTestModule("module2", nil),
		newTestModule("module3", nil),
	}

	graph, err := NewDependencyGraph(modules)
	require.NoError(t, err)

	levels, err := graph.TopologicalLevels()
	require.NoError(t, err)
	require.Len(t, levels, 1)
	assert.Len(t, levels[0], 3)
}

func TestDependencyGraphLinearDependencies(t *testing.T) {
	t.Parallel()

	// module3 -> module2 -> module1 (linear chain)
	module1 := newTestModule("module1", nil)
	module2 := newTestModule("module2", []string{"module1"})
	module3 := newTestModule("module3", []string{"module2"})

	modules := []bufmodule.Module{module3, module2, module1}

	graph, err := NewDependencyGraph(modules)
	require.NoError(t, err)

	levels, err := graph.TopologicalLevels()
	require.NoError(t, err)
	require.Len(t, levels, 3)

	// Level 0: module1 (no dependencies)
	assert.Len(t, levels[0], 1)
	assert.Equal(t, "module1", levels[0][0].OpaqueID())

	// Level 1: module2 (depends on module1)
	assert.Len(t, levels[1], 1)
	assert.Equal(t, "module2", levels[1][0].OpaqueID())

	// Level 2: module3 (depends on module2)
	assert.Len(t, levels[2], 1)
	assert.Equal(t, "module3", levels[2][0].OpaqueID())
}

func TestDependencyGraphDiamondDependencies(t *testing.T) {
	t.Parallel()

	// Diamond pattern:
	//     module4
	//    /      \
	// module2  module3
	//    \      /
	//     module1
	module1 := newTestModule("module1", nil)
	module2 := newTestModule("module2", []string{"module1"})
	module3 := newTestModule("module3", []string{"module1"})
	module4 := newTestModule("module4", []string{"module2", "module3"})

	modules := []bufmodule.Module{module4, module3, module2, module1}

	graph, err := NewDependencyGraph(modules)
	require.NoError(t, err)

	levels, err := graph.TopologicalLevels()
	require.NoError(t, err)
	require.Len(t, levels, 3)

	// Level 0: module1
	assert.Len(t, levels[0], 1)
	assert.Equal(t, "module1", levels[0][0].OpaqueID())

	// Level 1: module2 and module3 (can be parallel)
	assert.Len(t, levels[1], 2)

	// Level 2: module4
	assert.Len(t, levels[2], 1)
	assert.Equal(t, "module4", levels[2][0].OpaqueID())
}

func TestDependencyGraphGetDependencies(t *testing.T) {
	t.Parallel()

	module1 := newTestModule("module1", nil)
	module2 := newTestModule("module2", []string{"module1"})

	modules := []bufmodule.Module{module1, module2}

	graph, err := NewDependencyGraph(modules)
	require.NoError(t, err)

	// module1 has no dependencies
	deps := graph.GetDependencies("module1")
	assert.Empty(t, deps)

	// module2 depends on module1
	deps = graph.GetDependencies("module2")
	require.Len(t, deps, 1)
	assert.Equal(t, "module1", deps[0].OpaqueID())
}

func TestDependencyGraphGetDependents(t *testing.T) {
	t.Parallel()

	module1 := newTestModule("module1", nil)
	module2 := newTestModule("module2", []string{"module1"})
	module3 := newTestModule("module3", []string{"module1"})

	modules := []bufmodule.Module{module1, module2, module3}

	graph, err := NewDependencyGraph(modules)
	require.NoError(t, err)

	// module1 has two dependents
	dependents := graph.GetDependents("module1")
	assert.Len(t, dependents, 2)

	// module2 has no dependents
	dependents = graph.GetDependents("module2")
	assert.Empty(t, dependents)
}

func TestDependencyGraphExternalDependencies(t *testing.T) {
	t.Parallel()

	// Module depends on something not in our set (external dependency)
	module := newTestModule("module1", []string{"external-module"})

	graph, err := NewDependencyGraph([]bufmodule.Module{module})
	require.NoError(t, err)

	levels, err := graph.TopologicalLevels()
	require.NoError(t, err)
	require.Len(t, levels, 1)
	assert.Len(t, levels[0], 1)
}

// testModule is a test implementation of bufmodule.Module
type testModule struct {
	bufmodule.Module
	opaqueID string
	deps     []string
}

func newTestModule(opaqueID string, deps []string) *testModule {
	return &testModule{
		opaqueID: opaqueID,
		deps:     deps,
	}
}

func (m *testModule) OpaqueID() string {
	return m.opaqueID
}

func (m *testModule) ModuleDeps() ([]bufmodule.ModuleDep, error) {
	deps := make([]bufmodule.ModuleDep, len(m.deps))
	for i, depID := range m.deps {
		deps[i] = &testModuleDep{opaqueID: depID}
	}
	return deps, nil
}

// testModuleDep is a test implementation of bufmodule.ModuleDep
type testModuleDep struct {
	bufmodule.ModuleDep
	opaqueID string
}

func (d *testModuleDep) OpaqueID() string {
	return d.opaqueID
}
