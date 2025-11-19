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
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/bufbuild/buf/private/bufpkg/bufmodule"
	"github.com/stretchr/testify/assert"
)

func TestNoopProgressReporter(t *testing.T) {
	t.Parallel()

	// NoopProgressReporter should not panic when methods are called
	reporter := NoopProgressReporter{}
	module := newTestModule("test", nil)

	reporter.ModuleStarted(module)
	reporter.ModuleCompleted(module, time.Second)
	reporter.ModuleFailed(module, errors.New("test error"))
	reporter.LevelStarted(0, []bufmodule.Module{module})
	reporter.LevelCompleted(0, []bufmodule.Module{module}, time.Second)
}

func TestConsoleProgressReporterSingleModule(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	reporter := NewConsoleProgressReporter(&buf, 1, "Building")
	module := newTestModule("proto/user", nil)

	reporter.LevelStarted(0, []bufmodule.Module{module})
	reporter.LevelCompleted(0, []bufmodule.Module{module}, 500*time.Millisecond)

	output := buf.String()
	assert.Contains(t, output, "[1/1]")
	assert.Contains(t, output, "Building")
	assert.Contains(t, output, "proto/user")
	assert.Contains(t, output, "done")
}

func TestConsoleProgressReporterMultipleModules(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	reporter := NewConsoleProgressReporter(&buf, 3, "Linting")

	modules := []bufmodule.Module{
		newTestModule("proto/user", nil),
		newTestModule("proto/order", nil),
		newTestModule("proto/payment", nil),
	}

	reporter.LevelStarted(0, modules)
	reporter.LevelCompleted(0, modules, time.Second)

	output := buf.String()
	assert.Contains(t, output, "[1-3/3]")
	assert.Contains(t, output, "proto/user")
	assert.Contains(t, output, "proto/order")
	assert.Contains(t, output, "proto/payment")
}

func TestConsoleProgressReporterFailure(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	reporter := NewConsoleProgressReporter(&buf, 1, "Building")
	module := newTestModule("proto/broken", nil)

	reporter.ModuleFailed(module, errors.New("compilation error"))

	output := buf.String()
	assert.Contains(t, output, "failed")
	assert.Contains(t, output, "compilation error")
}

func TestProgressTracker(t *testing.T) {
	t.Parallel()

	reporter := NoopProgressReporter{}
	tracker := NewProgressTracker(reporter, 5)

	module1 := newTestModule("module1", nil)
	module2 := newTestModule("module2", nil)

	// Record success
	tracker.RecordStart(module1)
	tracker.RecordSuccess(module1, time.Second)

	// Record failure
	tracker.RecordStart(module2)
	tracker.RecordFailure(module2, errors.New("test error"))

	assert.Equal(t, 1, tracker.CompletedCount())
	assert.Equal(t, 1, tracker.FailedCount())
	assert.Greater(t, tracker.TotalDuration(), time.Duration(0))
}

func TestProgressTrackerConcurrency(t *testing.T) {
	t.Parallel()

	reporter := NoopProgressReporter{}
	tracker := NewProgressTracker(reporter, 100)

	// Simulate concurrent completions
	done := make(chan struct{})
	for i := 0; i < 100; i++ {
		go func(i int) {
			module := newTestModule(string(rune('a'+i%26)), nil)
			tracker.RecordStart(module)
			if i%2 == 0 {
				tracker.RecordSuccess(module, time.Millisecond)
			} else {
				tracker.RecordFailure(module, errors.New("error"))
			}
			done <- struct{}{}
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	assert.Equal(t, 50, tracker.CompletedCount())
	assert.Equal(t, 50, tracker.FailedCount())
}

func TestGetModuleDisplayName(t *testing.T) {
	t.Parallel()

	// Module without FullName should return OpaqueID
	module := newTestModule("test-opaque-id", nil)
	name := getModuleDisplayName(module)
	assert.Equal(t, "test-opaque-id", name)
}

func TestConsoleProgressReporterMultipleLevels(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	reporter := NewConsoleProgressReporter(&buf, 4, "Compiling")

	// Level 0: single module
	level0 := []bufmodule.Module{newTestModule("base", nil)}
	reporter.LevelStarted(0, level0)
	reporter.ModuleStarted(level0[0])
	reporter.ModuleCompleted(level0[0], 100*time.Millisecond)
	reporter.LevelCompleted(0, level0, 100*time.Millisecond)

	// Level 1: multiple modules
	level1 := []bufmodule.Module{
		newTestModule("user", nil),
		newTestModule("order", nil),
		newTestModule("payment", nil),
	}
	reporter.LevelStarted(1, level1)
	reporter.LevelCompleted(1, level1, 200*time.Millisecond)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.GreaterOrEqual(t, len(lines), 2)
}
