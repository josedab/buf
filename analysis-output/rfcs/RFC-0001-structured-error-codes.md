# RFC-0001: Structured Error Codes

**Status:** Draft
**Author:** Codebase Analysis
**Created:** November 18, 2025
**Effort:** 3-5 dev-days
**Category:** Quick Win

---

## Summary

Add structured error codes to all buf CLI error outputs to enable programmatic error handling, better tooling integration, and improved debugging.

---

## Motivation

Currently, buf errors are primarily text-based:

```
Error: module "buf.build/myorg/mymodule" not found
```

This makes it difficult to:
1. **Parse errors programmatically** - CI/CD must rely on string matching
2. **Provide context-specific help** - IDEs can't suggest fixes
3. **Track error patterns** - Analytics require error categorization
4. **Localize error messages** - Translations need stable identifiers

Other tools (rustc, eslint, typescript) use error codes effectively:

```
// TypeScript
error TS2322: Type 'string' is not assignable to type 'number'.

// ESLint
error no-unused-vars: 'x' is assigned a value but never used.
```

---

## Detailed Design

### Error Code Format

```
BUF{category}{number}
```

Where:
- **BUF** - Prefix for all buf errors
- **category** - Two-letter category code
- **number** - Three-digit error number

Examples:
- `BUFAU001` - Authentication: Invalid token
- `BUFMD001` - Module: Module not found
- `BUFLR001` - Lint Rule: FIELD_LOWER_SNAKE_CASE violation

### Error Categories

| Code | Category | Description |
|------|----------|-------------|
| AU | Authentication | Token, login, credential errors |
| MD | Module | Module resolution, dependency errors |
| WS | Workspace | Workspace configuration errors |
| CF | Config | buf.yaml, buf.gen.yaml errors |
| LR | Lint Rule | Individual lint rule violations |
| BR | Breaking Rule | Individual breaking rule violations |
| GN | Generate | Code generation errors |
| NW | Network | Connection, TLS, DNS errors |
| FS | Filesystem | File access, permission errors |
| IN | Internal | Bug/system errors |

### Implementation

#### 1. Error Code Registry

```go
// private/buf/bufcli/errorcode/errorcode.go
package errorcode

type Code string

const (
    // Authentication
    AuthInvalidToken      Code = "BUFAU001"
    AuthTokenExpired      Code = "BUFAU002"
    AuthLoginRequired     Code = "BUFAU003"

    // Module
    ModuleNotFound        Code = "BUFMD001"
    ModuleInvalidName     Code = "BUFMD002"
    ModuleDependencyCycle Code = "BUFMD003"

    // Workspace
    WorkspaceNotFound     Code = "BUFWS001"
    WorkspaceInvalid      Code = "BUFWS002"

    // Config
    ConfigNotFound        Code = "BUFCF001"
    ConfigInvalidYAML     Code = "BUFCF002"
    ConfigVersionInvalid  Code = "BUFCF003"

    // Network
    NetworkDNSError       Code = "BUFNW001"
    NetworkTLSError       Code = "BUFNW002"
    NetworkTimeout        Code = "BUFNW003"
)

func (c Code) String() string {
    return string(c)
}

func (c Code) URL() string {
    return fmt.Sprintf("https://buf.build/docs/errors/%s", c)
}
```

#### 2. Structured Error Type

```go
// private/buf/bufcli/errorcode/error.go
type StructuredError struct {
    Code    Code   `json:"code"`
    Message string `json:"message"`
    Detail  string `json:"detail,omitempty"`
    Help    string `json:"help,omitempty"`
}

func (e *StructuredError) Error() string {
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewError(code Code, message string, options ...ErrorOption) *StructuredError {
    err := &StructuredError{
        Code:    code,
        Message: message,
    }
    for _, opt := range options {
        opt(err)
    }
    return err
}
```

#### 3. Integration with Existing Errors

```go
// private/buf/bufcli/errors.go
func NewModuleNotFoundError(moduleName string) error {
    return errorcode.NewError(
        errorcode.ModuleNotFound,
        fmt.Sprintf("module %q not found", moduleName),
        errorcode.WithHelp(fmt.Sprintf(
            "Ensure the module exists on the registry or check your dependencies.\n"+
            "See: %s", errorcode.ModuleNotFound.URL(),
        )),
    )
}
```

#### 4. Output Format Updates

**Text format (default):**
```
BUFMD001: module "buf.build/myorg/mymodule" not found
  → Ensure the module exists on the registry or check your dependencies.
  → See: https://buf.build/docs/errors/BUFMD001
```

**JSON format:**
```json
{
  "code": "BUFMD001",
  "message": "module \"buf.build/myorg/mymodule\" not found",
  "help": "Ensure the module exists on the registry...",
  "url": "https://buf.build/docs/errors/BUFMD001"
}
```

#### 5. Lint/Breaking Rule Codes

Map existing rule IDs to error codes:

```go
// Each lint rule gets a unique code
const (
    LintFieldLowerSnakeCase    Code = "BUFLR001"
    LintEnumPascalCase         Code = "BUFLR002"
    LintServiceSuffix          Code = "BUFLR003"
    // ... 40+ rules
)

var lintRuleToCode = map[string]Code{
    "FIELD_LOWER_SNAKE_CASE": LintFieldLowerSnakeCase,
    "ENUM_PASCAL_CASE":       LintEnumPascalCase,
    // ...
}
```

### Documentation

Create error documentation at `https://buf.build/docs/errors/`:

```markdown
# BUFMD001: Module Not Found

## Description
The specified module could not be found in any configured registry.

## Common Causes
1. Typo in module name
2. Module not published to registry
3. Private module without authentication

## Solutions
1. Verify the module name is correct
2. Check if the module exists: `buf registry module info <name>`
3. Authenticate: `buf registry login`

## Related Errors
- BUFAU003: Authentication required
- BUFNW001: DNS resolution failed
```

---

## Example Usage

### Before

```
$ buf build
Error: module "buf.build/myorg/api" not found
```

### After

```
$ buf build
BUFMD001: module "buf.build/myorg/api" not found
  → Run 'buf registry login' if this is a private module
  → See: https://buf.build/docs/errors/BUFMD001
```

### Programmatic Usage

```bash
# Parse error code
ERROR=$(buf build 2>&1)
CODE=$(echo "$ERROR" | grep -oP 'BUF[A-Z]{2}\d{3}')

case "$CODE" in
    BUFAU*)
        echo "Authentication error - running login"
        buf registry login
        ;;
    BUFMD*)
        echo "Module error - checking registry"
        ;;
esac
```

---

## Implementation Plan

### Phase 1: Core Infrastructure (2 days)
1. Create error code registry
2. Implement StructuredError type
3. Update error output formatting

### Phase 2: Migration (2 days)
1. Convert existing errors to structured format
2. Map lint/breaking rules to codes
3. Update tests

### Phase 3: Documentation (1 day)
1. Create error documentation
2. Update user-facing docs
3. Add error code index

---

## Backwards Compatibility

**Impact:** Low

- Error codes are additive to existing messages
- Default text format remains readable
- JSON format gains new fields
- Existing error parsing may need updates

**Migration:**
- Deprecation warning for string-based error parsing
- Recommend using error codes after 6-month transition

---

## Alternatives Considered

### 1. Numeric-Only Codes
```
Error 1001: Module not found
```

**Rejected:** Less memorable, harder to search

### 2. Hierarchical Codes
```
buf.module.not_found
```

**Rejected:** Longer to type, harder to parse

### 3. Keep Current Format
```
Error: module not found
```

**Rejected:** Doesn't solve programmatic parsing

---

## Open Questions

1. **Code stability:** Should codes be immutable forever, or can they be deprecated?
   - Recommendation: Immutable with deprecation notices

2. **Localization:** How should error messages be localized?
   - Recommendation: Error codes are stable, messages can be translated

3. **Code allocation:** How should new codes be assigned?
   - Recommendation: Sequential within category, maintain registry

---

## Success Criteria

- [ ] All errors have structured codes
- [ ] Error documentation published
- [ ] JSON output includes codes
- [ ] CI/CD examples updated
- [ ] IDE integration guide available
- [ ] >95% programmatic parse success rate

---

## Stakeholder Approval

- [ ] Engineering Lead
- [ ] Developer Experience Team
- [ ] Documentation Team
