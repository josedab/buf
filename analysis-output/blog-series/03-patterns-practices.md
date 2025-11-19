# Patterns and Practices in Buf

**Series:** Buf Deep Dive (3 of 6)
**Commit SHA:** `e68c306ab3b6c39eef5abc23725d3e33d4a49cd4`

---

## What You'll Learn

- Provider pattern for dependency injection
- Option pattern for configuration
- Error handling with FileAnnotations
- Testing strategies with testdata

---

## Introduction

Buf's codebase contains 134,000+ lines of well-organized Go code. What keeps it maintainable? In this post, we'll explore the design patterns that enable 341 packages to work together cohesively without becoming a tangled mess.

These patterns aren't novel - they're established Go idioms. But Buf's consistent application of them throughout the codebase is exemplary.

---

## Pattern #1: Provider for Dependency Injection

The **Provider pattern** is Buf's approach to dependency injection. Instead of using a DI framework (like Wire or Uber's Dig), Buf defines explicit provider interfaces.

### The Pattern

```go
// Define what you need
type ModuleDataProvider interface {
    GetModuleData(ctx context.Context, key ModuleKey) (ModuleData, error)
}

// Accept it as a parameter
func NewModuleSetBuilder(provider ModuleDataProvider) *ModuleSetBuilder {
    return &ModuleSetBuilder{
        moduleDataProvider: provider,
    }
}
```

### Real Examples in Buf

**CommitProvider** for fetching module commits:

```go
// private/bufpkg/bufmodule/commit_provider.go
type CommitProvider interface {
    // GetCommitsForModuleKeys gets commits for module keys
    GetCommitsForModuleKeys(
        ctx context.Context,
        moduleKeys []ModuleKey,
    ) ([]Commit, error)
}
```

**GraphProvider** for dependency graphs:

```go
// private/bufpkg/bufmodule/graph_provider.go
type GraphProvider interface {
    GetGraph(
        ctx context.Context,
    ) (*dag.Graph[ModuleKey, Module], error)
}
```

**WorkspaceProvider** for loading workspaces:

```go
// private/buf/bufworkspace/workspace_provider.go
type WorkspaceProvider interface {
    GetWorkspaceForBucket(
        ctx context.Context,
        bucket storage.ReadBucket,
        options ...WorkspaceOption,
    ) (Workspace, error)
}
```

### Benefits

1. **Explicit dependencies** - No hidden magic, all deps visible in constructors
2. **Compile-time verification** - Missing implementations fail compilation
3. **Easy mocking** - Test implementations are straightforward
4. **No reflection** - Better performance
5. **Gradual construction** - Build complex objects piece by piece

### Example Usage

```go
// In production code
func main() {
    // Wire up real providers
    registryClient := bufregistry.NewClient(...)
    commitProvider := bufregistrycommit.NewProvider(registryClient)
    moduleDataProvider := bufmoduledata.NewProvider(commitProvider, cache)

    builder := bufmodule.NewModuleSetBuilder(moduleDataProvider)
    // ...
}

// In test code
func TestModuleSetBuilder(t *testing.T) {
    // Use mock provider
    mockProvider := &mockModuleDataProvider{
        data: map[ModuleKey]ModuleData{
            testKey: testData,
        },
    }

    builder := bufmodule.NewModuleSetBuilder(mockProvider)
    // Test builder behavior
}
```

---

## Pattern #2: Functional Options

The **Option pattern** (also called Functional Options) provides flexible, extensible configuration without constructor bloat.

### The Pattern

```go
// Define option type
type ControllerOption func(*controllerOptions)

// Define options struct (unexported)
type controllerOptions struct {
    logger      *slog.Logger
    cacheDir    string
    disableTLS  bool
}

// Create option functions
func WithLogger(logger *slog.Logger) ControllerOption {
    return func(o *controllerOptions) {
        o.logger = logger
    }
}

func WithCacheDir(dir string) ControllerOption {
    return func(o *controllerOptions) {
        o.cacheDir = dir
    }
}

// Apply options in constructor
func NewController(options ...ControllerOption) (*Controller, error) {
    opts := &controllerOptions{
        logger: slog.Default(),  // defaults
    }
    for _, option := range options {
        option(opts)
    }
    return &Controller{
        logger:   opts.logger,
        cacheDir: opts.cacheDir,
    }, nil
}
```

### Real Examples in Buf

**Controller options**:

```go
// private/buf/bufctl/option.go
func WithModuleKeyProvider(provider bufmodule.ModuleKeyProvider) ControllerOption {
    return func(o *controllerOptions) {
        o.moduleKeyProvider = provider
    }
}

func WithGraphProvider(provider bufmodule.GraphProvider) ControllerOption {
    return func(o *controllerOptions) {
        o.graphProvider = provider
    }
}
```

**Workspace options**:

```go
// private/buf/bufworkspace/option.go
func WithTargetPaths(paths []string) WorkspaceOption {
    return func(o *workspaceOptions) {
        o.targetPaths = paths
    }
}

func WithTargetExcludePaths(paths []string) WorkspaceOption {
    return func(o *workspaceOptions) {
        o.targetExcludePaths = paths
    }
}
```

**Image options**:

```go
// private/bufpkg/bufimage/option.go
func WithExcludeSourceCodeInfo() ImageOption {
    return func(o *imageOptions) {
        o.excludeSourceCodeInfo = true
    }
}
```

### Benefits

1. **Extensible** - Add new options without breaking existing code
2. **Self-documenting** - Option names describe what they do
3. **Optional defaults** - Sensible defaults with easy overrides
4. **Chainable** - Compose multiple options naturally

### Example Usage

```go
// Simple usage with defaults
controller, _ := bufctl.NewController()

// Customized
controller, _ := bufctl.NewController(
    bufctl.WithLogger(logger),
    bufctl.WithModuleKeyProvider(provider),
    bufctl.WithCacheDir("/tmp/buf-cache"),
)
```

---

## Pattern #3: Error Handling with Context

Buf has sophisticated error handling that preserves context while remaining user-friendly.

### FileAnnotation for Structured Errors

Instead of plain errors, lint/breaking violations use FileAnnotations:

```go
// private/bufpkg/bufanalysis/bufanalysis.go
type FileAnnotation interface {
    FileInfo() FileInfo
    StartLine() int
    StartColumn() int
    EndLine() int
    EndColumn() int
    Type() string
    Message() string
}

type FileAnnotationSet interface {
    FileAnnotations() []FileAnnotation
    Error() string
}
```

This enables:
- Multiple errors per invocation
- Precise source locations
- Structured output (JSON, GitHub Actions)
- Aggregation across files

### Error Wrapping with Context

Domain errors preserve context while remaining actionable:

```go
// private/buf/bufcli/errors.go
func NewModuleNotFoundError(moduleName string) error {
    return fmt.Errorf("module %q not found", moduleName)
}

func NewAuthenticationError(remote string, cause error) error {
    return fmt.Errorf(
        "authentication failed for %q: %w\n"+
        "Run 'buf registry login %s' to authenticate",
        remote,
        cause,
        remote,
    )
}
```

### Global Error Interceptor

The CLI has a global error interceptor that enriches errors:

```go
// cmd/buf/buf.go:480-550 (simplified)
func newErrorInterceptor() appext.Interceptor {
    return func(ctx context.Context, container appext.Container, next func() error) error {
        err := next()
        if err == nil {
            return nil
        }

        // Enhance specific error types
        var connectErr *connect.Error
        if errors.As(err, &connectErr) {
            return enhanceConnectError(connectErr)
        }

        var authErr *AuthenticationError
        if errors.As(err, &authErr) {
            return fmt.Errorf(
                "%w\nHint: Check your BUF_TOKEN or run 'buf registry login'",
                err,
            )
        }

        return err
    }
}
```

---

## Pattern #4: Table-Driven Tests

Buf uses table-driven tests extensively, especially for rule testing.

### The Pattern

```go
func TestFieldLowerSnakeCase(t *testing.T) {
    tests := []struct {
        name        string
        input       string
        wantErrors  []string
    }{
        {
            name:  "valid_field",
            input: "message Foo { string user_name = 1; }",
            wantErrors: nil,
        },
        {
            name:  "invalid_camelCase",
            input: "message Foo { string userName = 1; }",
            wantErrors: []string{
                `Field name "userName" should be lower_snake_case`,
            },
        },
        {
            name:  "invalid_PascalCase",
            input: "message Foo { string UserName = 1; }",
            wantErrors: []string{
                `Field name "UserName" should be lower_snake_case`,
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            errors := runLintRule("FIELD_LOWER_SNAKE_CASE", tt.input)
            assert.Equal(t, tt.wantErrors, errors)
        })
    }
}
```

### Testdata Directory Pattern

For complex scenarios, Buf uses testdata directories:

```
cmd/buf/testdata/
├── workspace/
│   ├── fail/
│   │   └── buf.yaml
│   │   └── proto/
│   │       └── invalid.proto
│   └── success/
│       └── buf.yaml
│       └── proto/
│           └── valid.proto
├── breaking/
│   ├── against/
│   │   └── proto/
│   └── current/
│       └── proto/
└── generate/
    └── ...
```

Tests read from these directories:

```go
// cmd/buf/buf_test.go
func TestWorkspaceSuccess(t *testing.T) {
    testCases := []struct {
        name string
        dir  string
    }{
        {"basic", "testdata/workspace/success/basic"},
        {"multi-module", "testdata/workspace/success/multi"},
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            _, err := runBuf(t, tc.dir, "build")
            require.NoError(t, err)
        })
    }
}
```

### Testing Utilities

Buf has dedicated testing utility packages:

```go
// private/bufpkg/bufanalysis/bufanalysistesting/bufanalysistesting.go
func AssertFileAnnotationsEqual(
    t *testing.T,
    expected []bufanalysis.FileAnnotation,
    actual []bufanalysis.FileAnnotation,
)

// private/pkg/storage/storagetesting/storagetesting.go
func NewReadBucket(t *testing.T, files map[string]string) storage.ReadBucket
```

---

## Pattern #5: Decorator/Wrapper

Buf uses the decorator pattern to add behavior without modifying core types.

### Example: Debug Logging Interceptor

```go
// private/bufpkg/bufconnect/interceptors.go
func NewDebugLoggingInterceptor(logger *slog.Logger) connect.UnaryInterceptorFunc {
    return func(next connect.UnaryFunc) connect.UnaryFunc {
        return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
            start := time.Now()

            // Call the actual handler
            resp, err := next(ctx, req)

            // Log after completion
            logger.Debug(
                "RPC completed",
                "procedure", req.Spec().Procedure,
                "duration", time.Since(start),
                "error", err,
            )

            return resp, err
        }
    }
}
```

### Example: Caching Wrapper

```go
// Conceptual caching decorator
type cachingModuleDataProvider struct {
    delegate ModuleDataProvider
    cache    map[ModuleKey]ModuleData
    mu       sync.RWMutex
}

func (p *cachingModuleDataProvider) GetModuleData(
    ctx context.Context,
    key ModuleKey,
) (ModuleData, error) {
    // Check cache
    p.mu.RLock()
    if data, ok := p.cache[key]; ok {
        p.mu.RUnlock()
        return data, nil
    }
    p.mu.RUnlock()

    // Delegate to underlying provider
    data, err := p.delegate.GetModuleData(ctx, key)
    if err != nil {
        return nil, err
    }

    // Cache result
    p.mu.Lock()
    p.cache[key] = data
    p.mu.Unlock()

    return data, nil
}
```

---

## Pattern #6: Strategy

Different strategies for the same operation, selected at runtime.

### Generation Strategies

```go
// private/buf/bufgen/strategy.go
type Strategy interface {
    Generate(ctx context.Context, image Image, plugin Plugin) error
}

type allStrategy struct{}  // Pass all files at once

type directoryStrategy struct{}  // Pass files by directory
```

Configuration selects strategy:

```yaml
# buf.gen.yaml
plugins:
  - remote: buf.build/protocolbuffers/go
    out: gen
    strategy: directory  # or "all"
```

---

## Code Organization Principles

### 1. Package by Feature, Not Layer

```
# Good - organized by feature
bufcheck/
  client.go
  config.go
  lint.go
  breaking.go

# Avoid - organized by type
interfaces/
  check_client.go
implementations/
  check_client_impl.go
```

### 2. Internal Packages for Implementation Details

```
bufcheck/
  client.go           # Public API
  bufcheckserver/     # Internal
    internal/
      bufcheckserverhandle/
        lint.go       # Implementation details
```

### 3. Explicit Public APIs

Each package has clear entry points:

```go
// private/bufpkg/bufcheck/bufcheck.go

// Package bufcheck provides linting and breaking change detection.
package bufcheck

// Public types and functions
type Client interface { ... }
func NewClient(options ...ClientOption) Client
```

---

## Key Takeaways

1. **Provider Pattern** - Explicit dependency injection through interfaces

2. **Option Pattern** - Flexible configuration with sensible defaults

3. **FileAnnotations** - Rich error context for diagnostics

4. **Table-Driven Tests** - Comprehensive coverage with clear structure

5. **Decorator Pattern** - Add behavior without modifying core types

6. **Strategy Pattern** - Runtime selection of algorithms

---

## What's Next

In Post 4, we'll explore **extending and integrating Buf** - the plugin architecture, code generation pipeline, and BSR API integration.

---

## Code References

- Controller Options: [`private/buf/bufctl/option.go`](https://github.com/bufbuild/buf/blob/e68c306ab3b6c39eef5abc23725d3e33d4a49cd4/private/buf/bufctl/option.go)
- FileAnnotation: [`private/bufpkg/bufanalysis/bufanalysis.go`](https://github.com/bufbuild/buf/blob/e68c306ab3b6c39eef5abc23725d3e33d4a49cd4/private/bufpkg/bufanalysis/bufanalysis.go)
- Interceptors: [`private/bufpkg/bufconnect/interceptors.go`](https://github.com/bufbuild/buf/blob/e68c306ab3b6c39eef5abc23725d3e33d4a49cd4/private/bufpkg/bufconnect/interceptors.go)
- Test Utilities: [`private/bufpkg/bufanalysis/bufanalysistesting/`](https://github.com/bufbuild/buf/blob/e68c306ab3b6c39eef5abc23725d3e33d4a49cd4/private/bufpkg/bufanalysis/bufanalysistesting/)

---

*Continue to Post 4: Extending and Integrating Buf →*
