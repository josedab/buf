# RFC-0004: Plugin Caching Strategy

**Status:** Draft
**Author:** Codebase Analysis
**Created:** November 18, 2025
**Effort:** 2-3 weeks
**Category:** Strategic

---

## Summary

Implement comprehensive plugin caching to eliminate startup overhead for repeated plugin invocations, reducing code generation time by up to 80%.

---

## Motivation

Plugin startup is the primary bottleneck in `buf generate`:

| Plugin Type | Current Startup | Target |
|-------------|-----------------|--------|
| Remote (first run) | 2-10s (download) | <1s (cached) |
| WASM instantiation | 50-200ms | <10ms |
| Docker container | 500ms-2s | <100ms |

For a typical 4-plugin generation:
- **Current:** 8-15 seconds
- **Target:** 1-3 seconds

This affects:
- Developer experience (slow iteration)
- CI/CD costs (longer pipeline times)
- Resource usage (redundant downloads)

---

## Detailed Design

### 1. Plugin Cache Architecture

```
~/.cache/buf/v2/
├── plugins/
│   ├── registry/
│   │   └── buf.build/
│   │       └── protocolbuffers/
│   │           └── go/
│   │               └── v1.28.1/
│   │                   ├── metadata.json
│   │                   └── plugin.wasm
│   ├── instances/
│   │   └── [sha256]/
│   │       └── compiled.wasm  # Pre-compiled WASM
│   └── docker/
│       └── [image-id]/
│           └── container.json  # Warm container config
└── cache.db  # SQLite for metadata
```

### 2. WASM Module Caching

Pre-compile WASM modules to native code:

```go
// private/bufpkg/bufremoteplugin/bufremotepluginwasm/cache.go
package bufremotepluginwasm

type CompiledModuleCache struct {
    cacheDir string
    runtime  wazero.Runtime
}

func (c *CompiledModuleCache) GetOrCompile(
    ctx context.Context,
    wasmBytes []byte,
) (wazero.CompiledModule, error) {
    hash := sha256.Sum256(wasmBytes)
    cachePath := filepath.Join(c.cacheDir, hex.EncodeToString(hash[:]))

    // Try loading pre-compiled
    if cached, err := c.runtime.CompilationCache().Get(cachePath); err == nil {
        return cached, nil
    }

    // Compile and cache
    compiled, err := c.runtime.CompileModule(ctx, wasmBytes)
    if err != nil {
        return nil, err
    }

    // Store compiled module
    err = c.runtime.CompilationCache().Put(cachePath, compiled)
    if err != nil {
        // Log but don't fail
        log.Warn("failed to cache compiled module", "error", err)
    }

    return compiled, nil
}
```

### 3. Docker Container Pooling

Keep warm containers for frequently-used plugins:

```go
// private/bufpkg/bufremoteplugin/bufremoteplugindocker/pool.go
package bufremoteplugindocker

type ContainerPool struct {
    client     *docker.Client
    containers map[string]*Container  // image -> warm container
    mu         sync.RWMutex
    maxSize    int
}

func (p *ContainerPool) Get(ctx context.Context, image string) (*Container, error) {
    p.mu.RLock()
    if container, ok := p.containers[image]; ok {
        p.mu.RUnlock()
        return container, nil
    }
    p.mu.RUnlock()

    // Create new container
    container, err := p.createWarmContainer(ctx, image)
    if err != nil {
        return nil, err
    }

    p.mu.Lock()
    // Evict if at capacity
    if len(p.containers) >= p.maxSize {
        p.evictOldest()
    }
    p.containers[image] = container
    p.mu.Unlock()

    return container, nil
}

func (p *ContainerPool) createWarmContainer(ctx context.Context, image string) (*Container, error) {
    // Create container but don't start
    container, err := p.client.CreateContainer(ctx, &docker.Config{
        Image: image,
        // Keep container alive
        Cmd: []string{"sleep", "infinity"},
    })
    if err != nil {
        return nil, err
    }

    // Start container
    err = p.client.StartContainer(ctx, container.ID)
    if err != nil {
        return nil, err
    }

    return &Container{
        ID:     container.ID,
        Image:  image,
        Ready:  true,
        client: p.client,
    }, nil
}
```

### 4. Plugin Metadata Caching

Cache plugin resolution results:

```go
// private/bufpkg/bufremoteplugin/cache.go
type PluginMetadataCache struct {
    db *sql.DB
}

type CachedPlugin struct {
    Ref         string    `db:"ref"`
    Digest      string    `db:"digest"`
    Type        string    `db:"type"`  // wasm, docker, binary
    LocalPath   string    `db:"local_path"`
    ResolvedAt  time.Time `db:"resolved_at"`
    ExpiresAt   time.Time `db:"expires_at"`
}

func (c *PluginMetadataCache) Get(ref string) (*CachedPlugin, error) {
    var plugin CachedPlugin
    err := c.db.QueryRow(`
        SELECT ref, digest, type, local_path, resolved_at, expires_at
        FROM plugins
        WHERE ref = ? AND expires_at > ?
    `, ref, time.Now()).Scan(
        &plugin.Ref, &plugin.Digest, &plugin.Type,
        &plugin.LocalPath, &plugin.ResolvedAt, &plugin.ExpiresAt,
    )
    if err != nil {
        return nil, err
    }
    return &plugin, nil
}
```

### 5. Cache Invalidation

```go
// Invalidation strategies
type InvalidationStrategy int

const (
    // Time-based: check for updates after TTL
    TTLBased InvalidationStrategy = iota
    // Version-based: invalidate when version changes
    VersionBased
    // Manual: user runs buf cache clean
    Manual
)

// Default TTLs
var defaultTTLs = map[string]time.Duration{
    "remote-plugin":    24 * time.Hour,
    "compiled-wasm":    7 * 24 * time.Hour,
    "container-pool":   1 * time.Hour,
}
```

### 6. CLI Commands

```bash
# View cache status
buf cache status
# Plugin cache: 156MB (12 plugins)
# WASM compiled: 89MB (8 modules)
# Container pool: 3 warm containers

# Clear specific cache
buf cache clean --plugins
buf cache clean --wasm
buf cache clean --containers

# Clear all
buf cache clean --all

# Pre-warm cache
buf cache warm --config buf.gen.yaml
# Downloading and caching 4 plugins...
```

---

## Example Performance Improvements

### Before (no caching)

```
$ time buf generate
Plugin buf.build/protocolbuffers/go: downloading... 3.2s
Plugin buf.build/protocolbuffers/go: instantiating... 0.15s
Plugin buf.build/grpc/go: downloading... 2.8s
Plugin buf.build/grpc/go: instantiating... 0.12s
Plugin buf.build/connect-go: downloading... 2.5s
Plugin buf.build/connect-go: instantiating... 0.14s
Generating...

real    0m9.234s
```

### After (cached)

```
$ time buf generate
Plugin buf.build/protocolbuffers/go: cached
Plugin buf.build/grpc/go: cached
Plugin buf.build/connect-go: cached
Generating...

real    0m1.456s
```

---

## Implementation Plan

### Phase 1: WASM Compilation Caching (1 week)
1. Implement CompiledModuleCache
2. Add wazero compilation cache integration
3. Test with existing WASM plugins

### Phase 2: Plugin Metadata Caching (0.5 week)
1. Implement SQLite-based metadata cache
2. Add cache invalidation logic
3. Integrate with plugin resolution

### Phase 3: Docker Container Pooling (1 week)
1. Implement ContainerPool
2. Add warm container management
3. Handle container lifecycle

### Phase 4: CLI and Polish (0.5 week)
1. Add cache management commands
2. Add cache status reporting
3. Documentation and testing

---

## Backwards Compatibility

**Impact:** Low

- Cache is opt-in (enabled by default)
- Existing behavior preserved on cache miss
- Clear migration path (cache builds up naturally)

**Configuration:**

```yaml
# buf.yaml (or global config)
cache:
  enabled: true
  directory: ~/.cache/buf
  plugins:
    ttl: 24h
    max_size: 500MB
  wasm:
    compile: true
  docker:
    pool_size: 5
```

---

## Alternatives Considered

### 1. Process Pooling
Keep plugin processes alive between runs.

**Rejected:** Complex lifecycle management, memory overhead.

### 2. Daemon Mode
Run buf as a daemon.

**Rejected:** Changes usage model significantly.

### 3. Remote Execution
Execute plugins on BSR servers.

**Rejected:** Latency, privacy concerns.

---

## Success Criteria

- [ ] WASM plugins <10ms startup (from cache)
- [ ] Docker plugins <100ms startup (warm pool)
- [ ] 80% reduction in total generation time
- [ ] Cache size <500MB for typical usage
- [ ] Cache management commands available

---

## Stakeholder Approval

- [ ] Engineering Lead
- [ ] Product
- [ ] Performance Team
