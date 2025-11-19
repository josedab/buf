# RFC-0008: Distributed Build Cache

**Status:** Draft
**Author:** Codebase Analysis
**Created:** November 18, 2025
**Effort:** 4-6 weeks
**Category:** Long-term

---

## Summary

Implement a distributed build cache that allows teams to share compilation results, dramatically reducing build times across team members and CI.

---

## Motivation

Each developer and CI job currently compiles from scratch:

| Scenario | Local Cache | Distributed Cache |
|----------|-------------|-------------------|
| Dev A builds | 30s | 30s (populate) |
| Dev B builds same | 30s | <1s (hit) |
| CI builds same | 30s | <1s (hit) |
| Team of 20 | 20 × 30s = 10min | 30s + 19 × 1s = 49s |

Annual savings for large team: **hundreds of developer hours**.

---

## Detailed Design

### 1. Cache Key Generation

```go
// private/buf/bufcache/distributed/key.go
package distributed

type CacheKey struct {
    // Content hash of input files
    ContentHash string `json:"content_hash"`
    // Buf version
    Version string `json:"version"`
    // Compilation options
    Options CompileOptions `json:"options"`
}

func GenerateCacheKey(files []File, options CompileOptions) CacheKey {
    h := sha256.New()

    // Sort files for determinism
    sort.Slice(files, func(i, j int) bool {
        return files[i].Path < files[j].Path
    })

    for _, file := range files {
        h.Write([]byte(file.Path))
        h.Write(file.Content)
    }

    return CacheKey{
        ContentHash: hex.EncodeToString(h.Sum(nil)),
        Version:     version.Version,
        Options:     options,
    }
}
```

### 2. Cache Backend Interface

```go
// private/buf/bufcache/distributed/backend.go
package distributed

type CacheBackend interface {
    Get(ctx context.Context, key CacheKey) ([]byte, error)
    Put(ctx context.Context, key CacheKey, data []byte) error
    Delete(ctx context.Context, key CacheKey) error
    Stats(ctx context.Context) (*CacheStats, error)
}

type CacheStats struct {
    Hits       int64
    Misses     int64
    Size       int64
    EntryCount int64
}
```

### 3. S3 Backend

```go
// private/buf/bufcache/distributed/s3.go
package distributed

type S3Backend struct {
    client *s3.Client
    bucket string
    prefix string
}

func (b *S3Backend) Get(ctx context.Context, key CacheKey) ([]byte, error) {
    result, err := b.client.GetObject(ctx, &s3.GetObjectInput{
        Bucket: &b.bucket,
        Key:    aws.String(b.keyPath(key)),
    })
    if err != nil {
        var nsk *types.NoSuchKey
        if errors.As(err, &nsk) {
            return nil, ErrCacheMiss
        }
        return nil, err
    }
    defer result.Body.Close()

    return io.ReadAll(result.Body)
}

func (b *S3Backend) Put(ctx context.Context, key CacheKey, data []byte) error {
    _, err := b.client.PutObject(ctx, &s3.PutObjectInput{
        Bucket: &b.bucket,
        Key:    aws.String(b.keyPath(key)),
        Body:   bytes.NewReader(data),
    })
    return err
}

func (b *S3Backend) keyPath(key CacheKey) string {
    return fmt.Sprintf("%s/%s/%s.bin", b.prefix, key.Version, key.ContentHash)
}
```

### 4. GCS Backend

```go
// private/buf/bufcache/distributed/gcs.go
package distributed

type GCSBackend struct {
    client *storage.Client
    bucket string
    prefix string
}

// Similar implementation to S3...
```

### 5. HTTP Backend

```go
// private/buf/bufcache/distributed/http.go
package distributed

type HTTPBackend struct {
    baseURL string
    client  *http.Client
    token   string
}

func (b *HTTPBackend) Get(ctx context.Context, key CacheKey) ([]byte, error) {
    url := fmt.Sprintf("%s/cache/%s", b.baseURL, key.ContentHash)

    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    req.Header.Set("Authorization", "Bearer "+b.token)

    resp, err := b.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusNotFound {
        return nil, ErrCacheMiss
    }

    return io.ReadAll(resp.Body)
}
```

### 6. Cache Integration

```go
// private/buf/bufctl/cache.go
package bufctl

type CachedController struct {
    controller Controller
    cache      distributed.CacheBackend
}

func (c *CachedController) GetImage(
    ctx context.Context,
    input string,
) (bufimage.Image, error) {
    // Generate cache key
    files := c.controller.ListFiles(ctx, input)
    key := distributed.GenerateCacheKey(files, c.options)

    // Check cache
    if data, err := c.cache.Get(ctx, key); err == nil {
        image, err := deserializeImage(data)
        if err == nil {
            return image, nil
        }
        // Cache corruption, fall through
    }

    // Build image
    image, err := c.controller.GetImage(ctx, input)
    if err != nil {
        return nil, err
    }

    // Populate cache (async)
    go func() {
        data := serializeImage(image)
        c.cache.Put(context.Background(), key, data)
    }()

    return image, nil
}
```

### 7. Configuration

```yaml
# buf.yaml or global config
cache:
  distributed:
    enabled: true
    backend: s3
    bucket: my-team-buf-cache
    prefix: buf/v1
    region: us-east-1

    # Or HTTP backend
    # backend: http
    # url: https://cache.mycompany.com
    # token_env: BUF_CACHE_TOKEN

    # Or GCS
    # backend: gcs
    # bucket: my-team-buf-cache
```

### 8. CLI Commands

```bash
# Check cache status
buf cache distributed status
# Backend: s3://my-team-buf-cache/buf/v1
# Entries: 1,234
# Size: 2.3 GB
# Hit rate: 87%

# Warm cache with current project
buf cache distributed warm

# Clear remote cache
buf cache distributed clear --older-than 30d
```

---

## Security Considerations

### Access Control
- Use IAM roles for S3/GCS
- Token-based auth for HTTP
- Encryption at rest

### Data Integrity
- SHA-256 content verification
- Version isolation
- Corruption detection

### Privacy
- No source code in cache (only compiled descriptors)
- Team-isolated buckets
- Audit logging

---

## Implementation Plan

### Phase 1: Core Infrastructure (2 weeks)
1. Cache key generation
2. Backend interface
3. S3 implementation

### Phase 2: Integration (2 weeks)
1. Controller integration
2. Configuration
3. CLI commands

### Phase 3: Additional Backends (1 week)
1. GCS backend
2. HTTP backend
3. Local backend (for testing)

### Phase 4: Polish (1 week)
1. Monitoring/metrics
2. Documentation
3. Migration guide

---

## Backwards Compatibility

**Impact:** None

Distributed cache is opt-in. Local cache behavior unchanged.

---

## Success Criteria

- [ ] >80% cache hit rate for teams
- [ ] <1s cache retrieval
- [ ] Works with S3, GCS, HTTP
- [ ] Proper access control
- [ ] Cache size <10GB per project

---

## Stakeholder Approval

- [ ] Engineering Lead
- [ ] Product
- [ ] Security Team
- [ ] Infrastructure Team
