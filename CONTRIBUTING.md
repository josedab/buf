# Contributing to Buf

Thank you for your interest in contributing to Buf! This document provides guidelines and information for contributors.

## Development Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/bufbuild/buf.git
   cd buf
   ```

2. Ensure you have Go installed (check `go.mod` for the required version).

3. Run the tests:
   ```bash
   make test
   ```

## Test Coverage

### Viewing Coverage Locally

Generate and view a coverage report:

```bash
# Generate coverage report
make cover

# Or manually generate and open in browser
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Coverage in CI

Our CI pipeline automatically generates and uploads coverage reports to [Codecov](https://codecov.io/gh/bufbuild/buf). Every pull request will show:

- Overall project coverage and any changes
- Coverage for the specific patch/changes
- Per-file coverage breakdown

### Coverage Requirements

- **Project threshold**: Coverage should not drop more than 2% from the baseline
- **Patch threshold**: New code should have at least 80% coverage

### Package-Specific Targets

Different packages have different coverage requirements:

| Package | Target | Rationale |
|---------|--------|-----------|
| `private/bufpkg/bufcheck/**` | 80% | Core linting/breaking change detection |
| `private/bufpkg/bufimage/**` | 75% | Core image processing |
| `cmd/**` | 60% | CLI commands (often integration-tested) |

### Improving Coverage

When contributing, consider:

1. **Add tests for new functionality**: All new features should have corresponding tests.
2. **Add tests for bug fixes**: Include a test that would have caught the bug.
3. **Focus on critical paths**: Prioritize coverage for error handling and edge cases.

### Files Excluded from Coverage

The following are excluded from coverage metrics:
- Generated code (`private/gen/**`)
- Test data (`**/testdata/**`)
- Test files (`**/*_test.go`)

## Submitting Changes

1. Fork the repository and create your branch from `main`.
2. Add tests for any new functionality.
3. Ensure all tests pass: `make test`
4. Ensure linting passes: `make lint`
5. Submit a pull request.

## Code Style

- Follow the existing code style
- Run `make lint` before submitting
- Use `buf format` for any `.proto` file changes

## Questions?

For questions about contributing, reach out on [Slack](https://buf.build/links/slack) or email [dev@buf.build](mailto:dev@buf.build).
