# Buf Terminology Glossary

**Commit SHA:** `e68c306ab3b6c39eef5abc23725d3e33d4a49cd4`

---

## Core Concepts

### Module

A collection of Protobuf files that are versioned and distributed together. The fundamental unit of organization in Buf.

**Key File:** [`private/bufpkg/bufmodule/module.go`](https://github.com/bufbuild/buf/blob/e68c306ab3b6c39eef5abc23725d3e33d4a49cd4/private/bufpkg/bufmodule/module.go)

```go
type Module interface {
    ModuleKey() ModuleKey
    Files() []File
    // ...
}
```

### ModuleKey

A unique identifier for a module, typically in the format `buf.build/owner/repository`.

**Example:** `buf.build/googleapis/googleapis`

### ModuleSet

A collection of modules that can be processed together, including transitive dependencies.

### Image

The compiled representation of Protobuf files. Contains FileDescriptors with all type information, source code info, and metadata.

**Key File:** [`private/bufpkg/bufimage/bufimage.go`](https://github.com/bufbuild/buf/blob/e68c306ab3b6c39eef5abc23725d3e33d4a49cd4/private/bufpkg/bufimage/bufimage.go)

### Workspace

A multi-module project container that manages multiple related modules and their dependencies.

**Configuration:** `buf.work.yaml`

**Key File:** [`private/buf/bufworkspace/workspace.go`](https://github.com/bufbuild/buf/blob/e68c306ab3b6c39eef5abc23725d3e33d4a49cd4/private/buf/bufworkspace/workspace.go)

### Controller

The central orchestrator that coordinates all buf operations. Acts as a facade over the workspace, module resolution, compilation, and processing systems.

**Key File:** [`private/buf/bufctl/controller.go`](https://github.com/bufbuild/buf/blob/e68c306ab3b6c39eef5abc23725d3e33d4a49cd4/private/buf/bufctl/controller.go)

---

## BSR (Buf Schema Registry)

### BSR

The Buf Schema Registry - a centralized repository for storing, versioning, and distributing Protobuf modules. Similar to npm for JavaScript or Maven for Java.

**URL:** `buf.build`

### Commit

A specific version of a module in the BSR. Immutable and content-addressed.

### Label

A human-readable reference to a commit, similar to git tags. Used for semantic versioning.

**Example:** `v1.0.0`, `latest`

### Remote

A BSR instance. Default is `buf.build` but can be self-hosted.

### Owner

An organization or user that owns modules in the BSR.

---

## Configuration

### buf.yaml

The primary configuration file for a Buf module. Defines the module name, dependencies, lint rules, and breaking change rules.

**Example:**
```yaml
version: v2
name: buf.build/myorg/mymodule
deps:
  - buf.build/googleapis/googleapis
lint:
  use:
    - STANDARD
breaking:
  use:
    - FILE
```

### buf.work.yaml

Workspace configuration file that defines multiple modules in a project.

**Example:**
```yaml
version: v2
modules:
  - path: proto/user
  - path: proto/order
```

### buf.gen.yaml

Code generation configuration file that defines plugins, outputs, and options.

**Example:**
```yaml
version: v2
plugins:
  - remote: buf.build/protocolbuffers/go
    out: gen
```

### buf.lock

Dependency lock file that records exact versions of dependencies for reproducible builds.

---

## Linting & Breaking Changes

### FileAnnotation

A diagnostic message associated with a specific location in a Protobuf file. Used for lint errors, breaking change violations, and compilation errors.

**Structure:**
```go
type FileAnnotation struct {
    Path    string  // File path
    Line    int     // Line number
    Column  int     // Column number
    Message string  // Error message
    Type    string  // Rule ID
}
```

### Lint Rule

A check that enforces a specific aspect of API design. Buf has 40+ built-in rules.

**Categories:**
- MINIMAL - Minimum requirements
- BASIC - Basic API design
- STANDARD - Standard conventions (default)
- COMMENTS - Documentation requirements

**Example Rules:** `FIELD_LOWER_SNAKE_CASE`, `SERVICE_SUFFIX`, `ENUM_ZERO_VALUE_SUFFIX`

### Breaking Rule

A check that detects backward-incompatible changes. Buf has 53+ built-in rules.

**Categories:**
- FILE - Source-level compatibility
- PACKAGE - Package-level compatibility
- WIRE - Wire format compatibility
- WIRE_JSON - Wire and JSON compatibility

**Example Rules:** `FIELD_NO_DELETE`, `MESSAGE_NO_REMOVE_STANDARD_DESCRIPTOR_ACCESSOR`

### Check Client

The interface for executing lint and breaking change checks.

**Key File:** [`private/bufpkg/bufcheck/client.go`](https://github.com/bufbuild/buf/blob/e68c306ab3b6c39eef5abc23725d3e33d4a49cd4/private/bufpkg/bufcheck/client.go)

---

## Code Generation

### Plugin

A program that generates code from compiled Protobuf. Can be local binary, remote BSR-hosted, or WASM module.

**Types:**
- **Local**: Binary on disk (e.g., `protoc-gen-go`)
- **Remote**: Hosted on BSR (e.g., `buf.build/protocolbuffers/go`)
- **WASM**: WebAssembly module for sandboxed execution

### Generator

The orchestrator that manages plugin execution for code generation.

**Key File:** [`private/buf/bufgen/generator.go`](https://github.com/bufbuild/buf/blob/e68c306ab3b6c39eef5abc23725d3e33d4a49cd4/private/buf/bufgen/generator.go)

### Strategy

How files are passed to plugins during generation.

- **all**: All files passed at once
- **directory**: Files passed by directory

---

## Storage & Fetching

### Bucket

An abstraction over file storage, similar to a filesystem or S3 bucket.

**Types:**
- `ReadBucket` - Read-only access
- `WriteBucket` - Write-only access
- `ReadWriteBucket` - Full access

**Key File:** [`private/pkg/storage/bucket.go`](https://github.com/bufbuild/buf/blob/e68c306ab3b6c39eef5abc23725d3e33d4a49cd4/private/pkg/storage/bucket.go)

### Ref (Reference)

A reference to input that can be resolved to content. Can be a file, directory, git repository, archive, or image.

**Examples:**
- `.` - Current directory
- `path/to/dir` - Local directory
- `https://github.com/org/repo.git` - Git repository
- `buf.build/owner/module` - BSR module

### Fetcher

Component that retrieves content from various sources (files, git, HTTP, archives).

**Key File:** [`private/buf/buffetch/`](https://github.com/bufbuild/buf/blob/e68c306ab3b6c39eef5abc23725d3e33d4a49cd4/private/buf/buffetch/)

---

## Compilation

### protocompile

The pure-Go Protobuf compiler used by Buf. Faster and more deterministic than protoc.

**Dependency:** `github.com/bufbuild/protocompile`

### FileDescriptor

The compiled representation of a single .proto file containing all type information.

### SourceInfo

Source code information (comments, locations) preserved from the original .proto file.

---

## Authentication

### Token

Authentication credential for BSR access. Can be provided via:
- `BUF_TOKEN` environment variable
- `.netrc` file
- Interactive login

### Netrc

Standard Unix credential file format used to store BSR tokens.

**Location:** `~/.netrc` (Unix), `~/_netrc` (Windows)

**Format:**
```
machine buf.build
  login <username>
  password <token>
```

---

## Architecture Terms

### Provider Pattern

Design pattern where interfaces define data providers that can be injected. Used throughout Buf for dependency injection without frameworks.

**Example:**
```go
type ModuleDataProvider interface {
    GetModuleData(ctx context.Context, key ModuleKey) (ModuleData, error)
}
```

### Option Pattern

Design pattern where configuration is passed as variadic function options. Provides flexible, extensible configuration.

**Example:**
```go
func NewController(options ...ControllerOption) (*Controller, error)

type ControllerOption func(*controllerOptions)

func WithLogger(logger *slog.Logger) ControllerOption {
    return func(o *controllerOptions) {
        o.logger = logger
    }
}
```

### Interceptor

Middleware that wraps RPC calls to add cross-cutting concerns like logging, auth, tracing.

**Key File:** [`private/bufpkg/bufconnect/interceptors.go`](https://github.com/bufbuild/buf/blob/e68c306ab3b6c39eef5abc23725d3e33d4a49cd4/private/bufpkg/bufconnect/interceptors.go)

---

## CLI Terms

### Exit Code

Numeric value returned by buf to indicate success or failure.

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General failure |
| 100 | File annotation failure (lint/breaking errors) |

### Output Format

Format for command output.

| Format | Use Case |
|--------|----------|
| text | Human-readable (default) |
| json | Machine-parseable |
| msvs | Visual Studio |
| junit | CI/CD test reporting |
| github-actions | GitHub Actions annotations |

---

## Observability Terms

### slog

Go's standard structured logging library used by Buf (Go 1.24+).

### OpenTelemetry

Observability framework for tracing and metrics. Buf uses `otelconnect` for RPC tracing.

### xslog.DebugProfile

Profiling utility that logs function execution duration at debug level.

---

## File Types

### .proto

Protocol Buffer definition file.

### .bin

Binary serialized Protobuf image.

### .json

JSON-encoded Protobuf data.

### .txtpb

Text-format Protobuf (human-readable).

---

## Version Formats

### v1, v1beta1, v2

Configuration file versions. Current is v2.

### Commit ID

SHA-256 based identifier for BSR commits.

### Module Reference

Format: `buf.build/owner/repository:reference`

**Examples:**
- `buf.build/googleapis/googleapis:v1.0.0`
- `buf.build/envoyproxy/envoy@<commit>`

---

## Common Abbreviations

| Abbrev | Full Form |
|--------|-----------|
| BSR | Buf Schema Registry |
| WKT | Well-Known Types |
| LSP | Language Server Protocol |
| RPC | Remote Procedure Call |
| CAS | Content-Addressed Storage |
| TLS | Transport Layer Security |
| mTLS | Mutual TLS |
| DI | Dependency Injection |

---

*Generated from commit `e68c306ab3b6c39eef5abc23725d3e33d4a49cd4`*
