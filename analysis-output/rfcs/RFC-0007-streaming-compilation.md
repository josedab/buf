# RFC-0007: Streaming Compilation Pipeline

**Status:** Draft
**Author:** Codebase Analysis
**Created:** November 18, 2025
**Effort:** 6-8 weeks
**Category:** Long-term

---

## Summary

Implement a streaming compilation pipeline that processes files incrementally, enabling faster builds for large codebases and better perceived performance.

---

## Motivation

Large codebases (10,000+ proto files) have significant compilation times:

| Files | Current Time | User Experience |
|-------|--------------|-----------------|
| 1,000 | 3-5s | Acceptable |
| 5,000 | 15-30s | Frustrating |
| 10,000 | 45-90s | Unacceptable |
| 20,000 | 2-5min | Deal-breaker |

Problems with current batch approach:
1. **All-or-nothing** - Must wait for complete compilation
2. **Memory intensive** - Entire image in memory
3. **Poor error feedback** - Errors shown only at end
4. **No incremental builds** - Small change = full rebuild

Streaming compilation addresses all these issues.

---

## Detailed Design

### 1. Streaming API

```go
// private/bufpkg/bufprotocompile/streaming.go
package bufprotocompile

type StreamingCompiler interface {
    // CompileStream returns results as they become available
    CompileStream(
        ctx context.Context,
        files []string,
    ) (<-chan CompileResult, error)
}

type CompileResult struct {
    File   string
    Result *descriptorpb.FileDescriptorProto
    Error  error
}

func (c *streamingCompiler) CompileStream(
    ctx context.Context,
    files []string,
) (<-chan CompileResult, error) {
    results := make(chan CompileResult, 100)

    go func() {
        defer close(results)

        // Build dependency graph
        graph := c.buildDependencyGraph(files)

        // Process in topological order
        for _, level := range graph.TopologicalLevels() {
            // Process level in parallel
            var wg sync.WaitGroup
            for _, file := range level {
                wg.Add(1)
                go func(f string) {
                    defer wg.Done()

                    result, err := c.compileFile(ctx, f)
                    results <- CompileResult{
                        File:   f,
                        Result: result,
                        Error:  err,
                    }
                }(file)
            }
            wg.Wait()
        }
    }()

    return results, nil
}
```

### 2. Incremental Image Building

```go
// private/bufpkg/bufimage/streaming.go
package bufimage

type StreamingImageBuilder struct {
    files    map[string]ImageFile
    mu       sync.RWMutex
    onChange func(ImageFile)
}

func (b *StreamingImageBuilder) AddFile(result CompileResult) error {
    if result.Error != nil {
        return result.Error
    }

    file := newImageFile(result.Result)

    b.mu.Lock()
    b.files[result.File] = file
    b.mu.Unlock()

    if b.onChange != nil {
        b.onChange(file)
    }

    return nil
}

func (b *StreamingImageBuilder) Build() Image {
    b.mu.RLock()
    defer b.mu.RUnlock()

    files := make([]ImageFile, 0, len(b.files))
    for _, f := range b.files {
        files = append(files, f)
    }
    return newImage(files)
}
```

### 3. Progressive Lint Results

```go
// private/bufpkg/bufcheck/streaming.go
package bufcheck

func (c *client) LintStream(
    ctx context.Context,
    config LintConfig,
    files <-chan ImageFile,
) (<-chan FileAnnotation, error) {
    results := make(chan FileAnnotation, 100)

    go func() {
        defer close(results)

        for file := range files {
            annotations := c.lintFile(ctx, config, file)
            for _, ann := range annotations {
                results <- ann
            }
        }
    }()

    return results, nil
}
```

### 4. Cache-Aware Compilation

```go
// private/bufpkg/bufprotocompile/cache.go
package bufprotocompile

type CachedCompiler struct {
    cache    *CompilationCache
    compiler StreamingCompiler
}

func (c *CachedCompiler) CompileStream(
    ctx context.Context,
    files []string,
) (<-chan CompileResult, error) {
    results := make(chan CompileResult, 100)

    go func() {
        defer close(results)

        var toCompile []string

        for _, file := range files {
            // Check cache
            if cached, ok := c.cache.Get(file); ok {
                results <- CompileResult{
                    File:   file,
                    Result: cached,
                }
                continue
            }
            toCompile = append(toCompile, file)
        }

        // Compile uncached files
        if len(toCompile) > 0 {
            stream, _ := c.compiler.CompileStream(ctx, toCompile)
            for result := range stream {
                if result.Error == nil {
                    c.cache.Put(result.File, result.Result)
                }
                results <- result
            }
        }
    }()

    return results, nil
}
```

### 5. Progress Reporting

```go
// private/buf/bufctl/progress.go
package bufctl

type StreamingProgress struct {
    Total     int
    Completed int
    Failed    int
    Current   string
}

func (c *controller) BuildWithProgress(
    ctx context.Context,
    input string,
    progress func(StreamingProgress),
) (Image, error) {
    files := c.listFiles(ctx, input)
    total := len(files)

    stream, err := c.compiler.CompileStream(ctx, files)
    if err != nil {
        return nil, err
    }

    builder := bufimage.NewStreamingImageBuilder()
    completed := 0
    failed := 0

    for result := range stream {
        if result.Error != nil {
            failed++
        } else {
            builder.AddFile(result)
            completed++
        }

        progress(StreamingProgress{
            Total:     total,
            Completed: completed,
            Failed:    failed,
            Current:   result.File,
        })
    }

    return builder.Build(), nil
}
```

### 6. CLI Integration

```bash
# Shows progress during build
$ buf build --progress

Compiling: [################----] 80% (8000/10000)
Current: proto/analytics/v1/events.proto
Errors: 3

# Stream errors as they occur
$ buf lint --stream

proto/user/v1/user.proto:42:3: FIELD_LOWER_SNAKE_CASE
proto/order/v1/order.proto:18:5: ENUM_ZERO_VALUE_SUFFIX
[streaming... 234 files processed]
```

---

## Implementation Plan

### Phase 1: Core Streaming (3 weeks)
1. Implement StreamingCompiler
2. Add dependency graph
3. Parallel level processing

### Phase 2: Image Builder (1 week)
1. Streaming image builder
2. Incremental updates
3. Memory optimization

### Phase 3: Cache Integration (2 weeks)
1. Compilation cache
2. Cache invalidation
3. Persistent storage

### Phase 4: CLI & Polish (2 weeks)
1. Progress reporting
2. Streaming output
3. Documentation

---

## Backwards Compatibility

**Impact:** Medium

- New API alongside existing
- Existing behavior unchanged
- Migration path for callers

**Migration:**

```go
// Old API (still works)
image, err := controller.GetImage(ctx, input)

// New API (streaming)
stream, err := controller.GetImageStream(ctx, input)
for result := range stream {
    // Process as available
}
```

---

## Success Criteria

- [ ] 70% faster perceived time (first result)
- [ ] Linear memory scaling
- [ ] Incremental cache hits >80%
- [ ] Progress reporting in CLI
- [ ] Same correctness as batch

---

## Stakeholder Approval

- [ ] Engineering Lead
- [ ] Product
- [ ] Architecture Review
