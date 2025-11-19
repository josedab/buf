# RFC-0005: Parallel Workspace Processing

**Status:** Draft
**Author:** Codebase Analysis
**Created:** November 18, 2025
**Effort:** 2-3 weeks
**Category:** Strategic

---

## Summary

Enable parallel processing of modules within a workspace to reduce build times for multi-module projects by 50-70%.

---

## Motivation

Large projects often have multiple modules:

```yaml
# buf.work.yaml
modules:
  - path: proto/user
  - path: proto/order
  - path: proto/payment
  - path: proto/notification
  - path: proto/analytics
```

Currently, operations process modules sequentially. With parallel processing:

| Operation | Sequential | Parallel (5 modules) |
|-----------|------------|----------------------|
| Build | 5s | 1.5s |
| Lint | 3s | 0.8s |
| Breaking | 10s | 3s |

---

## Detailed Design

### 1. Parallel Module Compilation

```go
// private/buf/bufworkspace/parallel.go
package bufworkspace

import (
    "context"
    "golang.org/x/sync/errgroup"
)

func (w *workspace) CompileParallel(ctx context.Context) ([]bufimage.Image, error) {
    modules := w.Modules()

    // Build dependency graph
    graph := w.buildDependencyGraph(modules)

    // Find independent modules (no dependencies between them)
    levels := graph.TopologicalLevels()

    var allImages []bufimage.Image

    // Process each level in parallel
    for _, level := range levels {
        images, err := w.compileLevel(ctx, level)
        if err != nil {
            return nil, err
        }
        allImages = append(allImages, images...)
    }

    return allImages, nil
}

func (w *workspace) compileLevel(ctx context.Context, modules []Module) ([]bufimage.Image, error) {
    g, ctx := errgroup.WithContext(ctx)
    images := make([]bufimage.Image, len(modules))

    for i, module := range modules {
        i, module := i, module // Capture for goroutine
        g.Go(func() error {
            image, err := w.compileModule(ctx, module)
            if err != nil {
                return err
            }
            images[i] = image
            return nil
        })
    }

    if err := g.Wait(); err != nil {
        return nil, err
    }

    return images, nil
}
```

### 2. Parallel Lint Execution

```go
// private/bufpkg/bufcheck/parallel.go
package bufcheck

func (c *client) LintParallel(
    ctx context.Context,
    modules []ModuleImages,
) ([]FileAnnotation, error) {
    g, ctx := errgroup.WithContext(ctx)
    results := make(chan []FileAnnotation, len(modules))

    for _, module := range modules {
        module := module
        g.Go(func() error {
            annotations, err := c.Lint(ctx, module.Config, module.Files)
            if err != nil {
                return err
            }
            results <- annotations
            return nil
        })
    }

    // Wait for all to complete
    go func() {
        g.Wait()
        close(results)
    }()

    // Collect results
    var all []FileAnnotation
    for annotations := range results {
        all = append(all, annotations...)
    }

    if err := g.Wait(); err != nil {
        return nil, err
    }

    return all, nil
}
```

### 3. Parallel Plugin Execution

```go
// private/buf/bufgen/parallel.go
package bufgen

func (g *generator) GenerateParallel(
    ctx context.Context,
    image bufimage.Image,
) error {
    eg, ctx := errgroup.WithContext(ctx)

    // Group plugins by output directory to avoid conflicts
    pluginGroups := g.groupPluginsByOutput()

    for _, group := range pluginGroups {
        group := group
        eg.Go(func() error {
            for _, plugin := range group {
                if err := g.executePlugin(ctx, plugin, image); err != nil {
                    return err
                }
            }
            return nil
        })
    }

    return eg.Wait()
}

func (g *generator) groupPluginsByOutput() [][]PluginConfig {
    // Plugins with same output dir must be sequential
    // Different output dirs can be parallel
    groups := make(map[string][]PluginConfig)
    for _, plugin := range g.config.Plugins {
        groups[plugin.Out] = append(groups[plugin.Out], plugin)
    }

    var result [][]PluginConfig
    for _, group := range groups {
        result = append(result, group)
    }
    return result
}
```

### 4. Dependency Graph

Handle inter-module dependencies:

```go
// private/buf/bufworkspace/graph.go
package bufworkspace

type DependencyGraph struct {
    nodes map[ModuleKey]*graphNode
}

type graphNode struct {
    module   Module
    deps     []*graphNode
    depCount int  // Number of unprocessed dependencies
}

func (g *DependencyGraph) TopologicalLevels() [][]Module {
    var levels [][]Module

    remaining := make(map[ModuleKey]*graphNode)
    for k, v := range g.nodes {
        remaining[k] = v
    }

    for len(remaining) > 0 {
        // Find nodes with no unprocessed dependencies
        var level []Module
        for key, node := range remaining {
            if node.depCount == 0 {
                level = append(level, node.module)
                delete(remaining, key)
            }
        }

        if len(level) == 0 {
            // Cycle detected
            return nil
        }

        // Decrement dep counts
        for _, mod := range level {
            for _, node := range remaining {
                for _, dep := range node.deps {
                    if dep.module.ModuleKey() == mod.ModuleKey() {
                        node.depCount--
                    }
                }
            }
        }

        levels = append(levels, level)
    }

    return levels
}
```

### 5. Concurrency Control

```go
// private/buf/bufworkspace/options.go
type ParallelOptions struct {
    MaxWorkers int  // 0 = runtime.NumCPU()
    Enabled    bool
}

// Default: parallel enabled with CPU count workers
var defaultParallelOptions = ParallelOptions{
    MaxWorkers: 0,
    Enabled:    true,
}
```

### 6. Progress Reporting

```go
// private/buf/bufworkspace/progress.go
type ProgressReporter interface {
    ModuleStarted(module Module)
    ModuleCompleted(module Module, duration time.Duration)
    ModuleFailed(module Module, err error)
}

// Console output
// [1/5] Compiling proto/user... done (0.3s)
// [2/5] Compiling proto/order... done (0.4s)
// [3-5/5] Compiling proto/payment, proto/notification, proto/analytics... done (0.5s)
```

---

## Example Performance

### Before (Sequential)

```
$ time buf build

Building proto/user... done (1.0s)
Building proto/order... done (1.2s)
Building proto/payment... done (0.8s)
Building proto/notification... done (0.9s)
Building proto/analytics... done (1.1s)

real    0m5.0s
```

### After (Parallel)

```
$ time buf build

[1/5] Building proto/user... done (1.0s)
[2-5/5] Building proto/order, proto/payment, proto/notification, proto/analytics... done (1.2s)

real    0m2.2s
```

---

## Implementation Plan

### Phase 1: Dependency Graph (0.5 week)
1. Implement DependencyGraph
2. Add topological sorting
3. Handle cycles

### Phase 2: Parallel Compilation (1 week)
1. Implement parallel module compilation
2. Add concurrency control
3. Handle errors properly

### Phase 3: Parallel Operations (1 week)
1. Parallel lint execution
2. Parallel generation
3. Parallel breaking detection

### Phase 4: Polish (0.5 week)
1. Progress reporting
2. Configuration options
3. Documentation

---

## Backwards Compatibility

**Impact:** Low

- Parallel mode is default but can be disabled
- Same results, different execution order
- Error messages may differ (first error vs all errors)

**Disable parallel:**

```bash
buf build --parallel=false
```

---

## Success Criteria

- [ ] 50%+ reduction in multi-module build time
- [ ] Correct handling of dependencies
- [ ] Clear progress reporting
- [ ] Graceful error handling
- [ ] Memory usage bounded

---

## Stakeholder Approval

- [ ] Engineering Lead
- [ ] Performance Team
