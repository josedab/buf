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
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/bufbuild/buf/private/bufpkg/bufmodule"
)

// ProgressReporter reports progress for parallel module processing.
type ProgressReporter interface {
	// ModuleStarted is called when a module starts processing.
	ModuleStarted(module bufmodule.Module)
	// ModuleCompleted is called when a module completes successfully.
	ModuleCompleted(module bufmodule.Module, duration time.Duration)
	// ModuleFailed is called when a module fails processing.
	ModuleFailed(module bufmodule.Module, err error)
	// LevelStarted is called when a parallel level starts processing.
	LevelStarted(levelIndex int, modules []bufmodule.Module)
	// LevelCompleted is called when a parallel level completes.
	LevelCompleted(levelIndex int, modules []bufmodule.Module, duration time.Duration)
}

// NoopProgressReporter is a ProgressReporter that does nothing.
type NoopProgressReporter struct{}

// ModuleStarted implements ProgressReporter.
func (NoopProgressReporter) ModuleStarted(bufmodule.Module) {}

// ModuleCompleted implements ProgressReporter.
func (NoopProgressReporter) ModuleCompleted(bufmodule.Module, time.Duration) {}

// ModuleFailed implements ProgressReporter.
func (NoopProgressReporter) ModuleFailed(bufmodule.Module, error) {}

// LevelStarted implements ProgressReporter.
func (NoopProgressReporter) LevelStarted(int, []bufmodule.Module) {}

// LevelCompleted implements ProgressReporter.
func (NoopProgressReporter) LevelCompleted(int, []bufmodule.Module, time.Duration) {}

// ConsoleProgressReporter reports progress to a console writer.
type ConsoleProgressReporter struct {
	writer      io.Writer
	totalCount  int
	operation   string
	mu          sync.Mutex
	completed   int
	startTimes  map[string]time.Time
}

// NewConsoleProgressReporter creates a new ConsoleProgressReporter.
func NewConsoleProgressReporter(writer io.Writer, totalCount int, operation string) *ConsoleProgressReporter {
	return &ConsoleProgressReporter{
		writer:     writer,
		totalCount: totalCount,
		operation:  operation,
		startTimes: make(map[string]time.Time),
	}
}

// ModuleStarted implements ProgressReporter.
func (r *ConsoleProgressReporter) ModuleStarted(module bufmodule.Module) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.startTimes[module.OpaqueID()] = time.Now()
}

// ModuleCompleted implements ProgressReporter.
func (r *ConsoleProgressReporter) ModuleCompleted(module bufmodule.Module, duration time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.completed++
}

// ModuleFailed implements ProgressReporter.
func (r *ConsoleProgressReporter) ModuleFailed(module bufmodule.Module, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.completed++
	fmt.Fprintf(r.writer, "[%d/%d] %s %s... failed: %v\n",
		r.completed, r.totalCount, r.operation, getModuleDisplayName(module), err)
}

// LevelStarted implements ProgressReporter.
func (r *ConsoleProgressReporter) LevelStarted(levelIndex int, modules []bufmodule.Module) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(modules) == 1 {
		fmt.Fprintf(r.writer, "[%d/%d] %s %s...",
			r.completed+1, r.totalCount, r.operation, getModuleDisplayName(modules[0]))
	} else {
		names := make([]string, len(modules))
		for i, m := range modules {
			names[i] = getModuleDisplayName(m)
		}
		startNum := r.completed + 1
		endNum := r.completed + len(modules)
		fmt.Fprintf(r.writer, "[%d-%d/%d] %s %s...",
			startNum, endNum, r.totalCount, r.operation, strings.Join(names, ", "))
	}
}

// LevelCompleted implements ProgressReporter.
func (r *ConsoleProgressReporter) LevelCompleted(levelIndex int, modules []bufmodule.Module, duration time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	fmt.Fprintf(r.writer, " done (%.1fs)\n", duration.Seconds())
}

// getModuleDisplayName returns a display-friendly name for a module.
func getModuleDisplayName(module bufmodule.Module) string {
	if fullName := module.FullName(); fullName != nil {
		return fullName.String()
	}
	return module.OpaqueID()
}

// ProgressTracker tracks progress for multiple concurrent operations.
type ProgressTracker struct {
	reporter   ProgressReporter
	totalCount int
	mu         sync.Mutex
	completed  int
	failed     int
	startTime  time.Time
}

// NewProgressTracker creates a new ProgressTracker.
func NewProgressTracker(reporter ProgressReporter, totalCount int) *ProgressTracker {
	return &ProgressTracker{
		reporter:   reporter,
		totalCount: totalCount,
		startTime:  time.Now(),
	}
}

// RecordStart records the start of processing a module.
func (t *ProgressTracker) RecordStart(module bufmodule.Module) {
	t.reporter.ModuleStarted(module)
}

// RecordSuccess records successful completion of a module.
func (t *ProgressTracker) RecordSuccess(module bufmodule.Module, duration time.Duration) {
	t.mu.Lock()
	t.completed++
	t.mu.Unlock()
	t.reporter.ModuleCompleted(module, duration)
}

// RecordFailure records a failed module.
func (t *ProgressTracker) RecordFailure(module bufmodule.Module, err error) {
	t.mu.Lock()
	t.failed++
	t.mu.Unlock()
	t.reporter.ModuleFailed(module, err)
}

// CompletedCount returns the number of completed modules.
func (t *ProgressTracker) CompletedCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.completed
}

// FailedCount returns the number of failed modules.
func (t *ProgressTracker) FailedCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.failed
}

// TotalDuration returns the total duration since tracking started.
func (t *ProgressTracker) TotalDuration() time.Duration {
	return time.Since(t.startTime)
}
