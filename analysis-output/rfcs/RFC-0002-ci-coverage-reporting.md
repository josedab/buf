# RFC-0002: CI Coverage Reporting

**Status:** Draft
**Author:** Codebase Analysis
**Created:** November 18, 2025
**Effort:** 2-3 dev-days
**Category:** Quick Win

---

## Summary

Integrate test coverage reporting into the buf CI pipeline to track coverage over time, identify gaps, and prevent regression.

---

## Motivation

The buf codebase has 660+ test functions and 35,741 lines of test code, but coverage metrics aren't visible in CI. This makes it difficult to:

1. **Track coverage trends** - Is coverage improving or declining?
2. **Identify gaps** - Which packages need more testing?
3. **Enforce minimums** - Block PRs that reduce coverage
4. **Celebrate wins** - Highlight coverage improvements

Current state:
- `make cover` generates reports locally
- No coverage tracking in CI
- No coverage badges or dashboards

---

## Detailed Design

### 1. Coverage Generation in CI

Update `.github/workflows/ci.yaml`:

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Run tests with coverage
        run: |
          go test -race -coverprofile=coverage.out -covermode=atomic ./...

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v4
        with:
          files: coverage.out
          fail_ci_if_error: true
```

### 2. Codecov Configuration

Create `codecov.yml`:

```yaml
coverage:
  precision: 2
  round: down
  range: "60...90"

  status:
    project:
      default:
        target: auto
        threshold: 2%  # Allow 2% decrease

    patch:
      default:
        target: 80%  # New code should be 80%+ covered

parsers:
  go:
    partials_as_hits: true

ignore:
  - "private/gen/**"  # Generated code
  - "**/testdata/**"
  - "**/*_test.go"

comment:
  layout: "reach,diff,files"
  behavior: default
  require_changes: true
```

### 3. Coverage Badges

Add to README.md:

```markdown
[![codecov](https://codecov.io/gh/bufbuild/buf/branch/main/graph/badge.svg)](https://codecov.io/gh/bufbuild/buf)
```

### 4. Package-Level Coverage Targets

```yaml
# codecov.yml
coverage:
  status:
    project:
      # Core packages need higher coverage
      bufcheck:
        paths:
          - "private/bufpkg/bufcheck/**"
        target: 80%

      bufimage:
        paths:
          - "private/bufpkg/bufimage/**"
        target: 75%

      # CLI can have lower coverage
      cmd:
        paths:
          - "cmd/**"
        target: 60%
```

### 5. PR Comments

Codecov will add coverage comments to PRs:

```markdown
## Coverage Report

| Coverage | Δ |
|----------|---|
| Overall | 72.5% (+0.3%) |
| Patch | 85.2% |

### Changed Files

| File | Coverage |
|------|----------|
| private/bufpkg/bufcheck/client.go | 78% (+5%) |
| private/buf/bufctl/controller.go | 65% (-2%) |

### Impacted Files

<details>
<summary>See details</summary>
...
</details>
```

---

## Example Usage

### View Coverage Locally

```bash
# Generate and view coverage
make cover

# Or manually
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Check Coverage in CI

```bash
# CI will fail if coverage drops >2%
# Or if new code has <80% coverage
```

---

## Implementation Plan

### Phase 1: Basic Integration (1 day)
1. Add coverage generation to CI
2. Set up Codecov account
3. Configure basic thresholds

### Phase 2: Configuration (1 day)
1. Define package-specific targets
2. Configure PR comments
3. Add badges to README

### Phase 3: Documentation (0.5 day)
1. Document how to view coverage
2. Add contributing guide section
3. Create coverage improvement guide

---

## Backwards Compatibility

**Impact:** None

This is a pure addition to CI. No code changes required.

---

## Alternatives Considered

### 1. Coveralls
Similar to Codecov but less Go-specific features.

### 2. SonarQube
More comprehensive but higher complexity.

### 3. Custom Solution
Build internal coverage tracking. Higher maintenance burden.

**Decision:** Codecov is the industry standard for open-source Go projects.

---

## Success Criteria

- [ ] Coverage reports in every PR
- [ ] Coverage badge in README
- [ ] Package-level targets defined
- [ ] Coverage trend visible
- [ ] >2% drop blocks PR

---

## Stakeholder Approval

- [ ] Engineering Lead
- [ ] Infrastructure Team
