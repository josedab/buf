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

package bufgen

import (
	"context"
	"errors"
	"testing"

	"github.com/bufbuild/buf/private/bufpkg/bufconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupPluginsByOutputEmpty(t *testing.T) {
	t.Parallel()

	groups := GroupPluginsByOutput(nil)
	assert.Nil(t, groups)
}

func TestGroupPluginsByOutputSinglePlugin(t *testing.T) {
	t.Parallel()

	plugins := []bufconfig.GeneratePluginConfig{
		newMockPluginConfig("plugin1", "out1"),
	}

	groups := GroupPluginsByOutput(plugins)
	require.Len(t, groups, 1)
	assert.Len(t, groups[0].Plugins, 1)
}

func TestGroupPluginsByOutputSameOutput(t *testing.T) {
	t.Parallel()

	// All plugins have same output directory
	plugins := []bufconfig.GeneratePluginConfig{
		newMockPluginConfig("plugin1", "gen/go"),
		newMockPluginConfig("plugin2", "gen/go"),
		newMockPluginConfig("plugin3", "gen/go"),
	}

	groups := GroupPluginsByOutput(plugins)
	require.Len(t, groups, 1)
	assert.Len(t, groups[0].Plugins, 3)
}

func TestGroupPluginsByOutputDifferentOutputs(t *testing.T) {
	t.Parallel()

	// Plugins have different output directories
	plugins := []bufconfig.GeneratePluginConfig{
		newMockPluginConfig("go-plugin", "gen/go"),
		newMockPluginConfig("java-plugin", "gen/java"),
		newMockPluginConfig("python-plugin", "gen/python"),
	}

	groups := GroupPluginsByOutput(plugins)
	require.Len(t, groups, 3)

	// Each group should have one plugin
	for _, group := range groups {
		assert.Len(t, group.Plugins, 1)
	}
}

func TestGroupPluginsByOutputMixed(t *testing.T) {
	t.Parallel()

	// Mix of same and different output directories
	plugins := []bufconfig.GeneratePluginConfig{
		newMockPluginConfig("go-proto", "gen/go"),
		newMockPluginConfig("go-grpc", "gen/go"),
		newMockPluginConfig("java-proto", "gen/java"),
	}

	groups := GroupPluginsByOutput(plugins)
	require.Len(t, groups, 2)

	// Find the go group
	var goGroup, javaGroup *PluginGroup
	for i := range groups {
		if len(groups[i].Plugins) == 2 {
			goGroup = &groups[i]
		} else {
			javaGroup = &groups[i]
		}
	}

	require.NotNil(t, goGroup)
	require.NotNil(t, javaGroup)
	assert.Len(t, goGroup.Plugins, 2)
	assert.Len(t, javaGroup.Plugins, 1)
}

func TestCreatePluginExecutionPlanEmpty(t *testing.T) {
	t.Parallel()

	plan := CreatePluginExecutionPlan(nil)
	assert.Nil(t, plan.ParallelGroups)
}

func TestCreatePluginExecutionPlanSinglePlugin(t *testing.T) {
	t.Parallel()

	plugins := []bufconfig.GeneratePluginConfig{
		newMockPluginConfig("plugin1", "out1"),
	}

	plan := CreatePluginExecutionPlan(plugins)
	require.Len(t, plan.ParallelGroups, 1)
	assert.Len(t, plan.ParallelGroups[0], 1)
}

func TestCreatePluginExecutionPlanMaximizesParallelism(t *testing.T) {
	t.Parallel()

	// Two plugins for go, one for java
	// Execution should be:
	// Batch 1: go-proto, java-proto (parallel)
	// Batch 2: go-grpc (sequential after go-proto)
	plugins := []bufconfig.GeneratePluginConfig{
		newMockPluginConfig("go-proto", "gen/go"),
		newMockPluginConfig("go-grpc", "gen/go"),
		newMockPluginConfig("java-proto", "gen/java"),
	}

	plan := CreatePluginExecutionPlan(plugins)
	require.Len(t, plan.ParallelGroups, 2)

	// First batch should have 2 plugins (one from each output dir)
	assert.Len(t, plan.ParallelGroups[0], 2)
	// Second batch should have 1 plugin (second go plugin)
	assert.Len(t, plan.ParallelGroups[1], 1)
}

func TestExecutePlanSuccess(t *testing.T) {
	t.Parallel()

	plugins := []bufconfig.GeneratePluginConfig{
		newMockPluginConfig("plugin1", "out1"),
		newMockPluginConfig("plugin2", "out2"),
	}

	plan := CreatePluginExecutionPlan(plugins)

	executed := make([]string, 0)
	err := ExecutePlan(
		context.Background(),
		plan,
		func(_ context.Context, plugin bufconfig.GeneratePluginConfig) error {
			executed = append(executed, plugin.Name())
			return nil
		},
	)

	require.NoError(t, err)
	assert.Len(t, executed, 2)
}

func TestExecutePlanFailure(t *testing.T) {
	t.Parallel()

	plugins := []bufconfig.GeneratePluginConfig{
		newMockPluginConfig("plugin1", "out1"),
	}

	plan := CreatePluginExecutionPlan(plugins)

	expectedErr := errors.New("execution failed")
	err := ExecutePlan(
		context.Background(),
		plan,
		func(_ context.Context, _ bufconfig.GeneratePluginConfig) error {
			return expectedErr
		},
	)

	assert.ErrorIs(t, err, expectedErr)
}

func TestGetParallelStats(t *testing.T) {
	t.Parallel()

	plugins := []bufconfig.GeneratePluginConfig{
		newMockPluginConfig("go-proto", "gen/go"),
		newMockPluginConfig("go-grpc", "gen/go"),
		newMockPluginConfig("java-proto", "gen/java"),
		newMockPluginConfig("python-proto", "gen/python"),
	}

	stats := GetParallelStats(plugins)

	assert.Equal(t, 4, stats.TotalPlugins)
	assert.Equal(t, 3, stats.UniqueOutputDirs)
	assert.Equal(t, 3, stats.MaxParallelism) // go, java, python can run in parallel
	assert.Equal(t, 2, stats.ParallelGroups) // 2 batches total
}

func TestGetParallelStatsEmpty(t *testing.T) {
	t.Parallel()

	stats := GetParallelStats(nil)

	assert.Equal(t, 0, stats.TotalPlugins)
	assert.Equal(t, 0, stats.UniqueOutputDirs)
	assert.Equal(t, 0, stats.MaxParallelism)
	assert.Equal(t, 0, stats.ParallelGroups)
}

// mockPluginConfig is a mock implementation of bufconfig.GeneratePluginConfig for testing
type mockPluginConfig struct {
	name string
	out  string
}

func newMockPluginConfig(name, out string) *mockPluginConfig {
	return &mockPluginConfig{name: name, out: out}
}

func (m *mockPluginConfig) Type() bufconfig.GeneratePluginConfigType {
	return bufconfig.GeneratePluginConfigTypeLocal
}
func (m *mockPluginConfig) Name() string                      { return m.name }
func (m *mockPluginConfig) Out() string                       { return m.out }
func (m *mockPluginConfig) Opt() string                       { return "" }
func (m *mockPluginConfig) IncludeImports() bool              { return false }
func (m *mockPluginConfig) IncludeWKT() bool                  { return false }
func (m *mockPluginConfig) Strategy() bufconfig.GenerateStrategy { return bufconfig.GenerateStrategyAll }
func (m *mockPluginConfig) Path() []string                    { return nil }
func (m *mockPluginConfig) ProtocPath() []string              { return nil }
func (m *mockPluginConfig) RemoteHost() string                { return "" }
func (m *mockPluginConfig) IncludeTypes() []string            { return nil }
func (m *mockPluginConfig) ExcludeTypes() []string            { return nil }
func (m *mockPluginConfig) Revision() int                     { return 0 }
