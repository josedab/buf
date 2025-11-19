# Deep Dive: Linting and Breaking Change Detection

**Series:** Buf Deep Dive (2 of 6)
**Commit SHA:** `e68c306ab3b6c39eef5abc23725d3e33d4a49cd4`

---

## What You'll Learn

- How lint rules are structured and executed
- The breaking change detection pipeline
- Rule configuration and customization
- Plugin-based rule architecture

---

## Introduction

Linting and breaking change detection are Buf's killer features. With 40+ lint rules and 53+ breaking change rules, Buf enforces API design best practices and prevents backward-incompatible changes from reaching production.

In this post, we'll explore how these systems work under the hood. We'll trace a lint rule from configuration to execution, understand the breaking change comparison algorithm, and see how you can extend the system with custom rules.

---

## The Check System Architecture

Both linting and breaking change detection share a common architecture. At the center is the [`CheckClient`](https://github.com/bufbuild/buf/blob/e68c306ab3b6c39eef5abc23725d3e33d4a49cd4/private/bufpkg/bufcheck/client.go):

```go
// private/bufpkg/bufcheck/client.go:35-50
type Client interface {
    // Lint checks files against lint rules
    Lint(
        ctx context.Context,
        config LintConfig,
        files []bufimage.ImageFile,
        options ...LintOption,
    ) ([]bufanalysis.FileAnnotation, error)

    // Breaking checks for breaking changes
    Breaking(
        ctx context.Context,
        config BreakingConfig,
        against []bufimage.ImageFile,
        current []bufimage.ImageFile,
        options ...BreakingOption,
    ) ([]bufanalysis.FileAnnotation, error)
}
```

The client delegates to rule implementations through a plugin-like architecture:

```
┌─────────────────┐     ┌──────────────┐     ┌───────────────┐
│   buf lint      │ --> │ CheckClient  │ --> │ Rule Handlers │
│   command       │     │              │     │ (40+ rules)   │
└─────────────────┘     └──────────────┘     └───────────────┘
         │                      │                     │
         │                      │                     │
         v                      v                     v
   buf.yaml config        Image files        FileAnnotations
```

---

## How Lint Rules Work

### Rule Categories

Buf organizes lint rules into categories of increasing strictness:

| Category | Description | Rules |
|----------|-------------|-------|
| MINIMAL | Basic requirements | ~10 rules |
| BASIC | Standard naming | ~20 rules |
| STANDARD | Full conventions (default) | ~30 rules |
| COMMENTS | Documentation requirements | ~5 rules |
| UNARY_RPC | RPC patterns | ~5 rules |

You configure these in `buf.yaml`:

```yaml
lint:
  use:
    - STANDARD
    - COMMENTS
  except:
    - ENUM_ZERO_VALUE_SUFFIX
```

### Rule Implementation

Each rule is implemented as a handler that receives proto types and returns violations. Let's look at a simple rule - [`FIELD_LOWER_SNAKE_CASE`](https://github.com/bufbuild/buf/blob/e68c306ab3b6c39eef5abc23725d3e33d4a49cd4/private/bufpkg/bufcheck/bufcheckserver/internal/bufcheckserverhandle/lint.go):

```go
// private/bufpkg/bufcheck/bufcheckserver/internal/bufcheckserverhandle/lint.go
// (simplified for clarity)

func checkFieldLowerSnakeCase(
    ctx context.Context,
    responseWriter check.ResponseWriter,
    request check.Request,
) error {
    // Iterate over all fields in all messages
    for _, file := range request.Files() {
        for _, message := range file.Messages() {
            for _, field := range message.Fields() {
                name := field.Name()
                expected := toLowerSnakeCase(name)

                if name != expected {
                    responseWriter.AddAnnotation(
                        check.WithMessagef(
                            "Field name %q should be lower_snake_case: %q",
                            name,
                            expected,
                        ),
                        check.WithDescriptor(field),
                    )
                }
            }
        }
    }
    return nil
}
```

The rule:
1. Receives compiled protobuf descriptors via `request.Files()`
2. Iterates through the type hierarchy
3. Checks invariants
4. Reports violations via `responseWriter.AddAnnotation()`

### The Execution Pipeline

When you run `buf lint`, here's what happens:

```go
// Simplified flow in cmd/buf/internal/command/lint/lint.go

func run(ctx context.Context, container appext.Container) error {
    // 1. Load configuration
    config, err := bufconfig.ReadConfig(ctx, "buf.yaml")

    // 2. Get compiled image
    controller := bufctl.NewController(container)
    image, err := controller.GetImage(ctx, input)

    // 3. Create check client
    checkClient := bufcheck.NewClient(logger)

    // 4. Run lint checks
    annotations, err := checkClient.Lint(
        ctx,
        config.Lint(),
        image.Files(),
    )

    // 5. Format and output results
    if len(annotations) > 0 {
        printAnnotations(annotations, format)
        return app.NewExitError(bufctl.ExitCodeFileAnnotation)
    }
    return nil
}
```

### Adding Context to Violations

Violations include precise source locations:

```go
// private/bufpkg/bufcheck/internal/check/annotation.go
type Annotation struct {
    Path        string  // proto/user/v1/user.proto
    StartLine   int     // 42
    StartColumn int     // 3
    EndLine     int     // 42
    EndColumn   int     // 15
    Type        string  // FIELD_LOWER_SNAKE_CASE
    Message     string  // Field name "UserName" should be...
}
```

This enables output like:

```
proto/user/v1/user.proto:42:3: Field name "UserName" should be lower_snake_case: "user_name"
```

---

## Breaking Change Detection

Breaking change detection compares two versions of your API - typically `HEAD` vs a baseline.

### The Comparison Algorithm

```go
// private/bufpkg/bufcheck/client.go (simplified)
func (c *client) Breaking(
    ctx context.Context,
    config BreakingConfig,
    against []ImageFile,  // baseline (old version)
    current []ImageFile,  // new version
    options ...BreakingOption,
) ([]FileAnnotation, error) {
    // Build maps for comparison
    againstMap := buildTypeMap(against)

    for _, file := range current {
        againstFile := againstMap[file.Path()]

        // Check each type against its previous version
        checkMessages(againstFile.Messages(), file.Messages())
        checkEnums(againstFile.Enums(), file.Enums())
        checkServices(againstFile.Services(), file.Services())
    }
}
```

### Breaking Rule Categories

| Category | Protects | Examples |
|----------|----------|----------|
| FILE | Source compatibility | Field deletion, type rename |
| PACKAGE | Package-level | Package rename |
| WIRE | Wire format | Field number change |
| WIRE_JSON | Wire + JSON | JSON name change |

### Example Breaking Rules

**FIELD_NO_DELETE** - Prevents removing fields:

```go
// private/bufpkg/bufcheck/bufcheckserver/internal/bufcheckserverhandle/breaking.go
// (conceptual)

func checkFieldNoDelete(against, current Message) []Annotation {
    var violations []Annotation

    currentFields := mapByNumber(current.Fields())

    for _, field := range against.Fields() {
        if _, exists := currentFields[field.Number()]; !exists {
            violations = append(violations, Annotation{
                Message: fmt.Sprintf(
                    "Previously present field %q with number %d was deleted",
                    field.Name(),
                    field.Number(),
                ),
                Type: "FIELD_NO_DELETE",
            })
        }
    }
    return violations
}
```

**FIELD_SAME_TYPE** - Prevents changing field types:

```go
func checkFieldSameType(against, current Field) *Annotation {
    if against.Type() != current.Type() {
        return &Annotation{
            Message: fmt.Sprintf(
                "Field %q changed type from %q to %q",
                current.Name(),
                against.Type(),
                current.Type(),
            ),
            Type: "FIELD_SAME_TYPE",
        }
    }
    return nil
}
```

### Using Breaking Detection

```bash
# Compare against git branch
buf breaking --against '.git#branch=main'

# Compare against BSR version
buf breaking --against 'buf.build/myorg/mymodule:v1.0.0'

# Compare against local archive
buf breaking --against 'baseline.tar.gz'
```

Configuration in `buf.yaml`:

```yaml
breaking:
  use:
    - WIRE_JSON  # Most strict for APIs
  except:
    - FIELD_SAME_CTYPE
  ignore:
    - internal/  # Ignore internal protos
```

---

## Configuration Deep Dive

### Loading Configuration

Configuration is loaded from [`buf.yaml`](https://github.com/bufbuild/buf/blob/e68c306ab3b6c39eef5abc23725d3e33d4a49cd4/private/bufpkg/bufconfig/buf_yaml_file.go):

```go
// private/bufpkg/bufconfig/lint_config.go (simplified)
type LintConfig interface {
    // UseRuleIDs returns enabled rule IDs
    UseRuleIDs() []string

    // ExceptRuleIDs returns disabled rule IDs
    ExceptRuleIDs() []string

    // IgnorePaths returns paths to skip
    IgnorePaths() []string

    // EnumZeroValueSuffix returns custom suffix
    EnumZeroValueSuffix() string

    // ServiceSuffix returns custom service suffix
    ServiceSuffix() string
}
```

### Rule Resolution

The system resolves rules in this order:

1. Start with category rules (e.g., STANDARD)
2. Add explicitly enabled rules
3. Remove excepted rules
4. Apply path ignores

```go
func resolveRules(config LintConfig, allRules []Rule) []Rule {
    enabled := make(map[string]bool)

    // Add category rules
    for _, id := range config.UseRuleIDs() {
        if category := getCategory(id); category != nil {
            for _, rule := range category.Rules {
                enabled[rule.ID] = true
            }
        } else {
            enabled[id] = true
        }
    }

    // Remove excepted
    for _, id := range config.ExceptRuleIDs() {
        delete(enabled, id)
    }

    // Filter to enabled rules
    var result []Rule
    for _, rule := range allRules {
        if enabled[rule.ID] {
            result = append(result, rule)
        }
    }
    return result
}
```

### Custom Rule Options

Some rules accept custom configuration:

```yaml
lint:
  use:
    - STANDARD
  enum_zero_value_suffix: _UNSPECIFIED  # Default is _UNSPECIFIED
  service_suffix: Service  # Default is Service
  rpc_allow_google_protobuf_empty_requests: true
```

---

## Plugin Architecture for Custom Rules

Buf supports custom rules via the plugin system. Rules can be:

1. **Built-in** - Compiled into buf
2. **Remote** - Downloaded from BSR
3. **WASM** - WebAssembly modules for sandboxed execution

### Writing a Custom Rule

Here's the structure of a custom lint plugin:

```go
// Your custom plugin
package main

import (
    "buf.build/go/bufplugin/check"
)

func main() {
    check.Main(
        check.NewSpec(
            "MY_CUSTOM_RULE",
            "Checks custom invariant",
            check.WithHandler(myCustomHandler),
        ),
    )
}

func myCustomHandler(
    ctx context.Context,
    responseWriter check.ResponseWriter,
    request check.Request,
) error {
    for _, file := range request.Files() {
        // Your custom logic here
        if violation {
            responseWriter.AddAnnotation(
                check.WithMessage("Custom violation detected"),
                check.WithDescriptor(descriptor),
            )
        }
    }
    return nil
}
```

### Registering Custom Rules

In `buf.yaml`:

```yaml
lint:
  use:
    - STANDARD
  plugins:
    - plugin: buf.build/myorg/custom-lint-rules:v1.0.0
```

---

## Performance Considerations

### Parallel Rule Execution

Rules are executed in parallel where possible:

```go
// private/bufpkg/bufcheck/client.go (conceptual)
func (c *client) Lint(ctx context.Context, ...) {
    var wg sync.WaitGroup
    results := make(chan []Annotation, len(rules))

    for _, rule := range rules {
        wg.Add(1)
        go func(r Rule) {
            defer wg.Done()
            annotations := r.Check(ctx, files)
            results <- annotations
        }(rule)
    }

    // Collect results
    wg.Wait()
    close(results)

    var all []Annotation
    for annotations := range results {
        all = append(all, annotations...)
    }
    return all
}
```

### Caching Compilation

The most expensive part is compiling protos. Buf caches images:

```go
// When comparing against git
buf breaking --against '.git#branch=main'

// The baseline image is cached after first compilation
// Subsequent checks reuse the cached image
```

### Incremental Checking

With `--limit-to-input-files`, only modified files are checked:

```bash
buf lint --limit-to-input-files
```

This dramatically speeds up CI checks in large codebases.

---

## Error Output Formats

Buf supports multiple output formats for different use cases:

```bash
# Default text format
buf lint
# proto/user.proto:42:3: Field name "UserName" should be lower_snake_case

# JSON for scripting
buf lint --error-format=json
# {"path":"proto/user.proto","start_line":42,...}

# GitHub Actions
buf lint --error-format=github-actions
# ::error file=proto/user.proto,line=42::Field name...

# JUnit for CI
buf lint --error-format=junit
# <testsuites>...</testsuites>
```

---

## Best Practices

### 1. Start Strict, Relax as Needed

```yaml
lint:
  use:
    - STANDARD
    - COMMENTS
  # Then add exceptions for specific cases
  except:
    - PACKAGE_DIRECTORY_MATCH
```

### 2. Use Ignores for Legacy Code

```yaml
lint:
  ignore:
    - legacy/
  ignore_only:
    FIELD_LOWER_SNAKE_CASE:
      - old_api/
```

### 3. Integrate Breaking Checks in CI

```yaml
# GitHub Actions
- name: Buf Breaking
  run: buf breaking --against '.git#branch=main'
```

### 4. Version Your Baseline

```bash
# Tag releases in BSR
buf push --label v1.0.0

# Check against tagged versions
buf breaking --against 'buf.build/myorg/api:v1.0.0'
```

---

## Key Takeaways

1. **Unified Check System** - Lint and breaking share architecture via `CheckClient`

2. **Rule Handlers** - Each rule is a function that receives descriptors and returns annotations

3. **Configuration Cascade** - Categories → explicit rules → exceptions → ignores

4. **Plugin Support** - Custom rules via remote plugins or WASM

5. **Multiple Formats** - Output adapts to text, JSON, GitHub Actions, JUnit

---

## What's Next

In Post 3, we'll explore the **patterns and practices** that make Buf's codebase maintainable - including the Provider pattern, Option pattern, and testing strategies.

---

## Code References

- Check Client: [`private/bufpkg/bufcheck/client.go`](https://github.com/bufbuild/buf/blob/e68c306ab3b6c39eef5abc23725d3e33d4a49cd4/private/bufpkg/bufcheck/client.go)
- Lint Rules: [`private/bufpkg/bufcheck/bufcheckserver/internal/bufcheckserverhandle/lint.go`](https://github.com/bufbuild/buf/blob/e68c306ab3b6c39eef5abc23725d3e33d4a49cd4/private/bufpkg/bufcheck/bufcheckserver/internal/bufcheckserverhandle/lint.go)
- Breaking Rules: [`private/bufpkg/bufcheck/bufcheckserver/internal/bufcheckserverhandle/breaking.go`](https://github.com/bufbuild/buf/blob/e68c306ab3b6c39eef5abc23725d3e33d4a49cd4/private/bufpkg/bufcheck/bufcheckserver/internal/bufcheckserverhandle/breaking.go)
- Lint Config: [`private/bufpkg/bufconfig/lint_config.go`](https://github.com/bufbuild/buf/blob/e68c306ab3b6c39eef5abc23725d3e33d4a49cd4/private/bufpkg/bufconfig/lint_config.go)

---

*Continue to Post 3: Patterns and Practices in Buf →*
