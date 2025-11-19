# RFC-0009: Custom Lint Rule DSL

**Status:** Draft
**Author:** Codebase Analysis
**Created:** November 18, 2025
**Effort:** 6-8 weeks
**Category:** Long-term

---

## Summary

Create a domain-specific language (DSL) for defining custom lint rules without writing Go/WASM code, making rule authoring accessible to all teams.

---

## Motivation

Currently, custom lint rules require:
1. Go programming knowledge
2. Understanding of bufplugin interfaces
3. WASM compilation setup
4. Plugin distribution

This barrier is too high for most teams. A DSL would:
- Enable rule authoring in minutes instead of hours
- Allow non-Go developers to contribute
- Integrate naturally with buf.yaml
- Support common patterns without code

---

## Detailed Design

### 1. Rule DSL Syntax

```yaml
# buf.yaml
lint:
  use:
    - STANDARD

  custom_rules:
    - id: FIELD_MUST_HAVE_COMMENT
      message: "Field '{{field.name}}' must have a comment"
      severity: error
      match:
        kind: field
        where:
          - comment == ""
          - name != "id"  # Except id fields

    - id: SERVICE_MUST_HAVE_HEALTH_CHECK
      message: "Service '{{service.name}}' must have a health check RPC"
      severity: warning
      match:
        kind: service
        where:
          - not(any(rpcs, name == "HealthCheck"))

    - id: ENUM_VALUES_MUST_BE_PREFIXED
      message: "Enum value '{{value.name}}' must be prefixed with '{{enum.name}}'"
      severity: error
      match:
        kind: enum_value
        where:
          - not(starts_with(name, upper_snake_case(enum.name)))

    - id: MESSAGE_MAX_FIELDS
      message: "Message '{{message.name}}' has {{len(fields)}} fields, max is 20"
      severity: warning
      match:
        kind: message
        where:
          - len(fields) > 20
```

### 2. DSL Expression Language

Based on CEL (Common Expression Language):

```yaml
# Basic comparisons
where:
  - name == "Foo"
  - number > 100
  - type == "string"

# String functions
where:
  - starts_with(name, "Get")
  - ends_with(name, "Request")
  - contains(comment, "deprecated")
  - matches(name, "^[A-Z][a-z]+$")

# List operations
where:
  - len(fields) > 10
  - any(fields, type == "string")
  - all(rpcs, has_comment)

# Case conversion
where:
  - name != lower_snake_case(name)
  - name != pascal_case(name)

# Negation
where:
  - not(deprecated)
  - not(any(options, name == "deprecated"))

# Type checking
where:
  - is_map
  - is_repeated
  - is_oneof
```

### 3. Match Kinds

```yaml
# Available kinds
match:
  kind: message | field | enum | enum_value | service | rpc | file | package | oneof | extension
```

### 4. Context Variables

```yaml
# Available in message context
message.name
message.full_name
message.fields
message.nested_messages
message.nested_enums
message.comment
message.options
message.file

# Available in field context
field.name
field.number
field.type
field.type_name
field.label  # optional, required, repeated
field.comment
field.options
field.message  # parent message
field.is_map
field.is_repeated

# Available in service context
service.name
service.rpcs
service.comment
service.file

# Available in rpc context
rpc.name
rpc.input_type
rpc.output_type
rpc.client_streaming
rpc.server_streaming
rpc.comment
rpc.service

# Available in enum context
enum.name
enum.values
enum.comment

# Available in enum_value context
value.name
value.number
value.comment
value.enum
```

### 5. DSL Implementation

```go
// private/bufpkg/bufcheck/dsl/dsl.go
package dsl

type RuleSpec struct {
    ID       string   `yaml:"id"`
    Message  string   `yaml:"message"`
    Severity string   `yaml:"severity"`
    Match    Match    `yaml:"match"`
}

type Match struct {
    Kind  string   `yaml:"kind"`
    Where []string `yaml:"where"`
}

func CompileRule(spec RuleSpec) (check.Rule, error) {
    // Parse CEL expressions
    env, err := cel.NewEnv(
        cel.Declarations(declsForKind(spec.Match.Kind)...),
    )
    if err != nil {
        return nil, err
    }

    var programs []cel.Program
    for _, expr := range spec.Match.Where {
        ast, issues := env.Compile(expr)
        if issues != nil && issues.Err() != nil {
            return nil, issues.Err()
        }
        prg, err := env.Program(ast)
        if err != nil {
            return nil, err
        }
        programs = append(programs, prg)
    }

    return &dslRule{
        spec:     spec,
        programs: programs,
    }, nil
}

type dslRule struct {
    spec     RuleSpec
    programs []cel.Program
}

func (r *dslRule) Check(ctx context.Context, file ImageFile) []FileAnnotation {
    var annotations []FileAnnotation

    for _, item := range itemsOfKind(file, r.spec.Match.Kind) {
        if r.matches(item) {
            annotations = append(annotations, FileAnnotation{
                File:    file.Path(),
                Line:    item.Line(),
                Column:  item.Column(),
                Type:    r.spec.ID,
                Message: r.formatMessage(item),
            })
        }
    }

    return annotations
}

func (r *dslRule) matches(item any) bool {
    activation := activationForItem(item)

    for _, prg := range r.programs {
        result, _, err := prg.Eval(activation)
        if err != nil || result.Value() != true {
            return false
        }
    }
    return true
}
```

### 6. Rule Testing

```yaml
# buf.yaml
lint:
  custom_rules:
    - id: FIELD_MUST_HAVE_COMMENT
      ...
      tests:
        - input: |
            message User {
              string name = 1;
            }
          expect_violation: true

        - input: |
            message User {
              // User's name
              string name = 1;
            }
          expect_violation: false
```

### 7. Rule Libraries

Share rules across projects:

```yaml
# buf.yaml
lint:
  custom_rules:
    # Import from BSR
    - import: buf.build/myorg/lint-rules:v1.0.0

    # Local rules
    - id: MY_LOCAL_RULE
      ...
```

---

## Example Rules

### 1. Require Comments on Public Fields

```yaml
- id: PUBLIC_FIELD_COMMENT
  message: "Public field '{{field.name}}' must have a comment"
  match:
    kind: field
    where:
      - comment == ""
      - not(starts_with(name, "_"))
```

### 2. Enforce RPC Naming

```yaml
- id: RPC_NAMING
  message: "RPC '{{rpc.name}}' must start with a verb"
  match:
    kind: rpc
    where:
      - not(matches(name, "^(Get|List|Create|Update|Delete|Watch|Search)"))
```

### 3. Limit Message Complexity

```yaml
- id: MESSAGE_COMPLEXITY
  message: "Message '{{message.name}}' has {{len(fields)}} fields (max 30)"
  match:
    kind: message
    where:
      - len(fields) > 30
```

### 4. Require Enum Prefix

```yaml
- id: ENUM_VALUE_PREFIX
  message: "Enum value must be prefixed with enum name"
  match:
    kind: enum_value
    where:
      - not(starts_with(name, upper_snake_case(enum.name) + "_"))
```

---

## Implementation Plan

### Phase 1: Core DSL (3 weeks)
1. DSL syntax parser
2. CEL integration
3. Basic match kinds

### Phase 2: All Match Kinds (2 weeks)
1. Complete all kinds
2. Context variables
3. Built-in functions

### Phase 3: Testing & Distribution (2 weeks)
1. Rule testing framework
2. Rule libraries
3. BSR integration

### Phase 4: Documentation (1 week)
1. User guide
2. Examples
3. Migration guide

---

## Backwards Compatibility

**Impact:** None

DSL rules are additive. Existing Go/WASM plugins continue to work.

---

## Success Criteria

- [ ] 90% of common rules expressible in DSL
- [ ] Rule authoring time <10 minutes
- [ ] Clear error messages for invalid DSL
- [ ] Rule testing support
- [ ] Library sharing via BSR

---

## Stakeholder Approval

- [ ] Engineering Lead
- [ ] Product
- [ ] Developer Experience Team
