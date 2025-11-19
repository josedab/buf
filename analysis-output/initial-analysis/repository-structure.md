# Buf Repository Structure

**Commit SHA:** `e68c306ab3b6c39eef5abc23725d3e33d4a49cd4`

---

## Root Directory Overview

```
/home/user/buf/
├── cmd/                      # Command-line entry points
├── private/                  # Internal implementation packages
├── proto/                    # Protocol buffer definitions
├── make/                     # Build configuration
├── etc/                      # Additional configuration and templates
├── .github/                  # GitHub Actions workflows
├── Dockerfile.buf            # Docker build configuration
├── Makefile                  # Build orchestration
├── go.mod                    # Go module definition
├── go.sum                    # Dependency checksums
├── buf.yaml                  # Buf configuration
├── buf.lock                  # Dependency lock file
├── README.md                 # Project documentation
├── CHANGELOG.md              # Release notes
└── LICENSE                   # Apache 2.0 license
```

---

## Detailed Directory Breakdown

### `/cmd/` - Command Entry Points (126 files, 29,910 LOC)

```
cmd/
├── buf/                              # Main CLI application
│   ├── buf.go                        # Entry point (625 lines)
│   ├── buf_test.go                   # CLI tests (4,787 lines)
│   ├── testdata/                     # Test fixtures
│   │   ├── workspace/                # Workspace test scenarios
│   │   ├── generate/                 # Generation test cases
│   │   ├── breaking/                 # Breaking change test cases
│   │   └── ...                       # 25+ test scenarios
│   └── internal/
│       └── command/                  # Command implementations
│           ├── build/                # buf build command
│           ├── lint/                 # buf lint command
│           ├── breaking/             # buf breaking command
│           ├── generate/             # buf generate command
│           ├── format/               # buf format command
│           ├── push/                 # buf push command
│           ├── export/               # buf export command
│           ├── convert/              # buf convert command
│           ├── curl/                 # buf curl command
│           ├── lsfiles/              # buf ls-files command
│           ├── stats/                # buf stats command
│           ├── dep/                  # buf dep commands
│           │   ├── deprune/          # buf dep prune
│           │   ├── depupdate/        # buf dep update
│           │   └── depgraph/         # buf dep graph
│           ├── config/               # buf config commands
│           │   ├── configinit/       # buf config init
│           │   ├── configmigrate/    # buf config migrate
│           │   └── configls*/        # buf config ls-* commands
│           ├── registry/             # buf registry commands
│           │   ├── registrylogin/    # buf registry login
│           │   ├── registrylogout/   # buf registry logout
│           │   ├── module/           # buf registry module commands
│           │   ├── plugin/           # buf registry plugin commands
│           │   └── organization/     # buf registry org commands
│           ├── lsp/                  # buf lsp command
│           ├── plugin/               # buf plugin commands
│           ├── beta/                 # Beta features
│           └── alpha/                # Alpha features
├── protoc-gen-buf-breaking/          # Breaking change protoc plugin
│   └── main.go                       # Plugin entry (55 lines)
└── protoc-gen-buf-lint/              # Linting protoc plugin
    └── main.go                       # Plugin entry (55 lines)
```

### `/private/` - Internal Packages (851 files)

#### `/private/buf/` - CLI Support Packages (153 files, 33,477 LOC)

```
private/buf/
├── bufctl/                           # Central controller
│   ├── controller.go                 # Main orchestrator (1,671 lines)
│   ├── option.go                     # Controller options
│   └── bufctl.go                     # Exit codes, utilities
├── bufgen/                           # Code generation
│   ├── generator.go                  # Generation orchestration
│   └── config.go                     # Generation config
├── bufformat/                        # Proto formatting
│   └── formatter.go                  # Format logic (2,493 lines)
├── bufworkspace/                     # Workspace management
│   ├── workspace.go                  # Workspace abstraction
│   ├── workspace_provider.go         # Workspace loading
│   └── workspace_targeting.go        # Target selection
├── buffetch/                         # Input fetching
│   ├── ref_parser.go                 # Reference parsing
│   ├── git_reader.go                 # Git repository fetching
│   └── archive_reader.go             # Archive handling
├── buflsp/                           # Language Server Protocol
│   ├── server.go                     # LSP server
│   ├── completion.go                 # Code completion (1,642 lines)
│   └── diagnostics.go                # Error diagnostics
├── bufcurl/                          # gRPC curl implementation
│   ├── curl.go                       # Main curl logic
│   └── tls.go                        # TLS configuration
├── bufmigrate/                       # Config migration
│   └── migrate.go                    # v1beta1 → v2 migration
├── bufprotopluginexec/               # Plugin execution
│   └── exec.go                       # Plugin runner
├── bufprint/                         # Output formatting
│   └── printer.go                    # Text/JSON/YAML output
├── bufstudioagent/                   # Studio integration
│   └── agent.go                      # Studio agent
├── bufconvert/                       # Format conversion
│   └── convert.go                    # Message conversion
├── buftarget/                        # Target selection
│   └── target.go                     # File targeting
├── bufapp/                           # Application initialization
│   └── app.go                        # App setup
├── bufcli/                           # CLI utilities
│   ├── errors.go                     # Domain errors
│   └── connectclient_config.go       # RPC client config
├── buftesting/                       # Test utilities
│   └── testing.go                    # Test helpers
└── bufwkt/                           # Well-Known Types
    └── wkt.go                        # WKT handling
```

#### `/private/bufpkg/` - Core Business Logic (384 files, 74,946 LOC)

```
private/bufpkg/
├── bufcheck/                         # Lint & breaking detection
│   ├── client.go                     # Check client (1,002 lines)
│   ├── bufcheck.go                   # Core logic
│   └── bufcheckserver/               # Check server
│       └── internal/
│           └── bufcheckserverhandle/
│               ├── lint.go           # Lint rules (1,375 lines)
│               └── breaking.go       # Breaking rules (2,397 lines)
├── bufimage/                         # Protobuf images
│   ├── bufimage.go                   # Image abstraction
│   ├── bufimagebuild/                # Image building
│   └── bufimageutil/                 # Image utilities (1,080 lines)
├── bufmodule/                        # Module system
│   ├── module.go                     # Module interface
│   ├── module_set.go                 # Module collections
│   ├── module_key.go                 # Module identifiers
│   └── module_read_bucket.go         # Module reading (959 lines)
├── bufconfig/                        # Configuration
│   ├── buf_yaml_file.go              # YAML parsing (1,451 lines)
│   ├── buf_lock_file.go              # Lock file (856 lines)
│   ├── lint_config.go                # Lint configuration
│   ├── breaking_config.go            # Breaking configuration
│   └── generate_plugin_config.go     # Generation config
├── bufremoteplugin/                  # Remote plugins
│   ├── bufremoteplugindocker/        # Docker execution
│   │   └── docker.go                 # Docker client
│   └── bufremotepluginwasm/          # WASM execution
│       └── wasm.go                   # WASM runtime
├── bufpolicy/                        # Policy management
│   └── policy.go                     # Policy configuration
├── bufprotosource/                   # Proto source handling
│   ├── bufprotosource.go             # Source abstraction (1,328 lines)
│   └── file.go                       # File handling (851 lines)
├── bufplugin/                        # Plugin abstraction
│   └── plugin.go                     # Plugin interface
├── bufregistryapi/                   # BSR API client
│   ├── bufregistryapimodule/         # Module API
│   ├── bufregistryapiplugin/         # Plugin API
│   └── bufregistryapiowner/          # Owner API
├── bufanalysis/                      # Result analysis
│   ├── bufanalysis.go                # Analysis types
│   └── bufanalysistesting/           # Test utilities
├── bufprotoplugin/                   # Plugin protocol
│   └── protoplugin.go                # Protocol handling
├── bufcas/                           # Content-addressed storage
│   └── cas.go                        # CAS implementation
├── bufconnect/                       # Connect RPC client
│   ├── interceptors.go               # RPC interceptors
│   └── errors.go                     # Error handling
├── bufcobra/                         # Cobra extensions
│   └── cobra.go                      # CLI framework
├── bufparse/                         # Proto parsing
│   └── parse.go                      # Parser utilities
├── bufprotocompile/                  # Compilation layer
│   └── compile.go                    # Protobuf compilation
├── bufreflect/                       # Reflection utilities
│   └── reflect.go                    # Proto reflection
└── buftransport/                     # HTTP transport
    └── transport.go                  # Transport config
```

#### `/private/pkg/` - Shared Utilities (222 files, 26,058 LOC)

```
private/pkg/
├── storage/                          # Storage abstraction
│   ├── bucket.go                     # Bucket interface
│   ├── storagemem/                   # In-memory storage
│   ├── storageos/                    # OS filesystem
│   ├── storagegit/                   # Git storage
│   └── storagetesting/               # Test utilities (45K lines)
├── normalpath/                       # Path normalization
│   └── normalpath.go                 # Cross-platform paths
├── git/                              # Git integration
│   └── git.go                        # Git operations
├── netrc/                            # Netrc parsing
│   └── netrc.go                      # Credential file
├── httpauth/                         # HTTP authentication
│   └── httpauth.go                   # Auth handlers
├── oauth2/                           # OAuth2 support
│   └── oauth2.go                     # Device flow
├── cert/                             # Certificate utilities
│   └── cert.go                       # TLS certificates
├── cache/                            # Caching layer
│   └── cache.go                      # Local cache
├── encoding/                         # Encoding utilities
│   └── encoding.go                   # Proto encoding
├── protodescriptor/                  # Descriptor utilities
│   └── descriptor.go                 # Descriptor helpers
├── protoencoding/                    # Proto marshaling
│   └── encoding.go                   # Marshal/unmarshal
├── protogenutil/                     # Generation utilities
│   └── genutil.go                    # Code gen helpers
├── protosourcepath/                  # Source paths
│   └── sourcepath.go                 # Path handling
├── diff/                             # Diff computation
│   └── diff.go                       # File diffing
├── dag/                              # Graph utilities
│   └── dag.go                        # DAG operations
├── bandeps/                          # Dependency analysis
│   └── bandeps.go                    # Banned deps
├── slogapp/                          # Logging
│   ├── slogapp.go                    # Logger factory
│   └── console.go                    # Console handler
├── syserror/                         # System errors
│   └── syserror.go                   # Error wrapper
├── verbose/                          # Verbose mode
│   └── verbose.go                    # Verbose printer
├── netext/                           # Network utilities
│   └── netext.go                     # Hostname validation
└── transport/
    └── http/
        ├── httpclient/               # HTTP client
        └── httpserver/               # HTTP server
```

#### `/private/gen/` - Generated Code (79,289 LOC)

```
private/gen/
└── proto/                            # Generated protobuf code
    ├── go/                           # Go generated code
    │   └── buf/
    │       └── alpha/
    │           ├── image/            # Image definitions
    │           ├── module/           # Module API
    │           ├── registry/         # Registry services
    │           ├── lint/             # Lint config
    │           ├── breaking/         # Breaking config
    │           └── ...               # Other services
    └── bufbuild/                     # Buf-specific protos
```

### `/proto/` - Protocol Buffer Definitions

```
proto/
└── buf/
    └── alpha/
        ├── image/
        │   └── v1/
        │       └── image.proto       # Image format definition
        ├── module/
        │   └── v1alpha1/
        │       └── module.proto      # Module API
        ├── registry/
        │   └── v1alpha1/             # Registry services
        │       ├── module.proto      # Module service
        │       ├── plugin.proto      # Plugin service
        │       ├── organization.proto # Organization service
        │       ├── repository.proto  # Repository service
        │       └── ...               # 15+ service definitions
        ├── lint/
        │   └── v1/
        │       └── config.proto      # Lint configuration
        ├── breaking/
        │   └── v1/
        │       └── config.proto      # Breaking configuration
        ├── studio/
        │   └── v1alpha1/
        │       └── invoke.proto      # Studio integration
        ├── webhook/
        │   └── v1alpha1/
        │       └── event.proto       # Webhook events
        └── audit/
            └── v1alpha1/
                └── event.proto       # Audit logging
```

### `/make/` - Build Configuration

```
make/
├── buf/
│   └── all.mk                        # Buf-specific build rules
└── go/                               # Go build framework (makego)
```

### `/.github/` - CI/CD Configuration

```
.github/
├── workflows/
│   ├── ci.yaml                       # Main CI pipeline
│   ├── buf-ci.yaml                   # Buf checks
│   ├── build-and-draft-release.yaml  # Release builds
│   ├── docker-publish.yaml           # Docker publishing
│   ├── create-release-pr.yaml        # Release automation
│   ├── windows.yaml                  # Windows testing
│   ├── verify-changelog.yaml         # Changelog checks
│   ├── codeql.yaml                   # Security scanning
│   └── buf-binary-size.yaml          # Binary size tracking
└── ISSUE_TEMPLATE/                   # Issue templates
```

---

## Package Relationships

### Core Dependencies

```
cmd/buf/buf.go
    └── bufctl.Controller
        ├── bufworkspace (module loading)
        │   ├── bufmodule (module abstraction)
        │   ├── buffetch (input fetching)
        │   └── bufparse (proto parsing)
        ├── bufimage (compilation)
        │   └── bufprotocompile (protobuf compiler)
        ├── bufcheck (lint/breaking)
        │   └── bufconfig (configuration)
        ├── bufgen (code generation)
        │   └── bufremoteplugin (plugin execution)
        └── bufformat (formatting)
            └── bufparse (proto parsing)
```

### Layer Dependencies

```
CLI Layer (cmd/buf/)
    ↓ uses
Application Layer (bufctl/)
    ↓ orchestrates
Domain Layer (bufpkg/)
    ↓ utilizes
Infrastructure Layer (pkg/)
```

---

## File Count Summary

| Directory | Go Files | Purpose |
|-----------|----------|---------|
| cmd/ | 126 | CLI entry points |
| private/buf/ | 153 | CLI support packages |
| private/bufpkg/ | 384 | Core business logic |
| private/pkg/ | 222 | Shared utilities |
| private/gen/ | 92 | Generated code |
| **Total** | **977** | |

---

## Key Files to Study

### Entry Points
- `cmd/buf/buf.go` - Main CLI entry
- `private/buf/bufctl/controller.go` - Central orchestrator

### Core Logic
- `private/bufpkg/bufcheck/client.go` - Lint/breaking client
- `private/bufpkg/bufimage/bufimage.go` - Image abstraction
- `private/bufpkg/bufmodule/module.go` - Module interface

### Configuration
- `private/bufpkg/bufconfig/buf_yaml_file.go` - YAML parsing
- `private/bufpkg/bufconfig/lint_config.go` - Lint rules

### Testing
- `cmd/buf/buf_test.go` - Comprehensive CLI tests
- `private/bufpkg/bufcheck/lint_test.go` - Lint tests

---

*Generated from commit `e68c306ab3b6c39eef5abc23725d3e33d4a49cd4`*
