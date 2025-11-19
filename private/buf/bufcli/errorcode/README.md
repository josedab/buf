# Error Code Package

This package provides structured error codes for the buf CLI, enabling programmatic error handling, better tooling integration, and improved debugging.

## Error Code Format

Error codes follow the format `BUF{category}{number}`:
- **BUF** - Prefix for all buf errors
- **category** - Two-letter category code
- **number** - Three-digit error number

Example: `BUFMD001` for "Module not found"

## Error Categories

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

## Quick Start

### Creating Errors

```go
import "github.com/bufbuild/buf/private/buf/bufcli/errorcode"

// Using helper functions (recommended)
err := errorcode.NewModuleNotFoundError("buf.build/myorg/mymodule")

// Using NewError directly
err := errorcode.NewError(
    errorcode.ModuleNotFound,
    "module not found",
    errorcode.WithHelp("Check your module name"),
)

// Wrapping an existing error
err := errorcode.Wrap(errorcode.NetworkDNSError, originalErr)
```

### Checking Errors

```go
// Check if error has specific code
if errorcode.HasCode(err, errorcode.ModuleNotFound) {
    // Handle module not found
}

// Get code from error
code := errorcode.GetCode(err)

// Check if error is structured
if errorcode.IsStructuredError(err) {
    // Handle structured error
}
```

### Formatting Output

```go
err := errorcode.NewModuleNotFoundError("buf.build/myorg/mymodule")

// Text format
fmt.Print(err.Format())
// Output:
// BUFMD001: module "buf.build/myorg/mymodule" not found
//   → Ensure the module exists on the registry or check your dependencies.
//   → Run 'buf registry login' if this is a private module.
//   → See: https://buf.build/docs/errors/BUFMD001

// JSON format
data, _ := json.Marshal(err)
// {"code":"BUFMD001","message":"module \"buf.build/myorg/mymodule\" not found",...}
```

### Rule Mapping

Convert lint/breaking rule IDs to error codes:

```go
// Lint rules
code := errorcode.LintRuleToCode("FIELD_LOWER_SNAKE_CASE")  // BUFLR015

// Breaking rules
code := errorcode.BreakingRuleToCode("ENUM_NO_DELETE")  // BUFBR001

// Any rule
code := errorcode.RuleToCode("MESSAGE_PASCAL_CASE")  // BUFLR021

// Check rule type
if errorcode.IsLintRule("FIELD_LOWER_SNAKE_CASE") {
    // Handle lint rule
}
```

## Error Reference

Each error code has associated documentation at `https://buf.build/docs/errors/{CODE}`.

### Example: BUFMD001

**Module Not Found**

#### Description
The specified module could not be found in any configured registry.

#### Common Causes
1. Typo in module name
2. Module not published to registry
3. Private module without authentication

#### Solutions
1. Verify the module name is correct
2. Check if the module exists: `buf registry module info <name>`
3. Authenticate: `buf registry login`

## Adding New Error Codes

1. Add the code constant to `errorcode.go` in the appropriate category
2. Add a helper function to `helpers.go` if needed
3. Add rule mapping to `rules.go` if it's a lint/breaking rule
4. Add tests
5. Update documentation

## Testing

```bash
go test ./private/buf/bufcli/errorcode/...
```
