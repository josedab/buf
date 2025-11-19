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

/*
Parallel Workspace Processing

This package provides parallel processing capabilities for buf workspaces
with multiple modules. It enables significant build time reductions for
large projects by processing independent modules concurrently.

# Overview

The parallel processing system consists of several components:

  - DependencyGraph: Analyzes module dependencies and groups independent
    modules for parallel execution
  - ParallelCompiler: Compiles multiple modules in parallel while respecting
    dependency order
  - ProgressReporter: Provides real-time feedback on parallel processing
    progress
  - ParallelOptions: Configures parallelism behavior (worker count, cancellation)

# Dependency Analysis

Modules are organized into topological levels based on their dependencies.
Modules at the same level have no dependencies on each other and can be
processed in parallel:

	Level 0: [base-module]           <- No dependencies, processed first
	Level 1: [user, order, payment]  <- Depend only on base-module
	Level 2: [analytics]             <- Depends on user and order

# Usage

Basic parallel compilation:

	compiler := NewParallelCompiler(logger, reporter,
	    WithMaxWorkers(8),
	    WithCancelOnFailure(true),
	)
	images, err := compiler.CompileModules(ctx, modules)

Using workspace convenience function:

	images, err := CompileWorkspace(ctx, logger, workspace, reporter)

# Performance

Parallel processing can reduce build times by 50-70% for multi-module
workspaces, depending on the number of modules and their dependency structure.

Example performance comparison:

	Sequential (5 modules): 5.0s
	Parallel (5 modules):   2.2s  (56% reduction)

# Configuration

Parallel processing is enabled by default. To disable:

	compiler := NewParallelCompiler(logger, reporter,
	    WithParallelEnabled(false),
	)

Worker count defaults to runtime.NumCPU() but can be customized:

	compiler := NewParallelCompiler(logger, reporter,
	    WithMaxWorkers(16),
	)

# Progress Reporting

Use ConsoleProgressReporter for terminal output:

	reporter := NewConsoleProgressReporter(os.Stdout, len(modules), "Building")

Output format:

	[1/5] Building proto/user... done (0.3s)
	[2-4/5] Building proto/order, proto/payment, proto/notification... done (0.5s)
	[5/5] Building proto/analytics... done (0.4s)

# Error Handling

By default, processing stops on first error (CancelOnFailure). To collect
all errors:

	compiler := NewParallelCompiler(logger, reporter,
	    WithCancelOnFailure(false),
	)
*/
package bufworkspace
