# Buf Codebase Metrics Summary

**Commit SHA:** `e68c306ab3b6c39eef5abc23725d3e33d4a49cd4`
**Analysis Date:** November 18, 2025

---

## Code Size Metrics

### Lines of Code

| Category | LOC | Percentage |
|----------|-----|------------|
| **Total** | 244,395 | 100% |
| Source (non-generated) | 134,481 | 55.0% |
| Generated (protobuf) | 79,289 | 32.4% |
| Test code | 35,741 | 14.6% |

### File Counts

| Category | Count |
|----------|-------|
| Total Go files | 977 |
| Test files | 106 |
| Packages | 341 |
| Commands | 21 major groups |

### Lines per Directory

| Directory | Files | LOC | Non-Gen LOC |
|-----------|-------|-----|-------------|
| private/bufpkg | 384 | 74,946 | 60,124 |
| private/buf | 153 | 33,477 | 29,258 |
| private/pkg | 222 | 26,058 | 26,058 |
| cmd/buf | 123 | 27,896 | 19,412 |
| private/gen | 92 | 79,289 | — |

---

## Code Quality Metrics

### File Size Distribution

| Metric | Value |
|--------|-------|
| Average file size | 354 LOC |
| Median file size | ~200 LOC |
| Maximum file size | 2,493 LOC |
| Standard deviation | 4,235 LOC |

### Largest Files

| File | LOC | Package |
|------|-----|---------|
| bufformat/formatter.go | 2,493 | bufformat |
| bufcheckserverhandle/breaking.go | 2,397 | bufcheck |
| bufctl/controller.go | 1,671 | bufctl |
| buflsp/completion.go | 1,642 | buflsp |
| bufconfig/buf_yaml_file.go | 1,451 | bufconfig |
| bufcheckserverhandle/lint.go | 1,375 | bufcheck |
| bufprotosource/bufprotosource.go | 1,328 | bufprotosource |
| curl/curl.go | 1,195 | curl |
| buflintvalidate/field.go | 1,178 | buflintvalidate |
| bufimageutil/bufimageutil.go | 1,080 | bufimageutil |

### Code Duplication Indicators

| Metric | Value |
|--------|-------|
| TODO/FIXME comments | 285 |
| Interface definitions | 340+ |
| Unique packages | 341 |

---

## Testing Metrics

### Test Coverage

| Metric | Value |
|--------|-------|
| Test files | 106 |
| Test functions | 660+ |
| Test LOC | 35,741 |
| Test:Source ratio | 26.6% |
| Parallel tests | 943 |
| Benchmark tests | 5 |

### Test Distribution

| Package | Test Files | Test Functions |
|---------|------------|----------------|
| cmd/buf | 1 | 200+ |
| bufcheck | 5 | 100+ |
| bufimage | 8 | 80+ |
| bufmodule | 12 | 150+ |
| bufformat | 3 | 50+ |

### Test Infrastructure

| Component | Count |
|-----------|-------|
| Testdata directories | 24 |
| Testing utilities | 9 packages |
| Platform-specific tests | 11 |
| Test fixtures (googleapis) | 1,574 files |

---

## Dependency Metrics

### External Dependencies

| Category | Count |
|----------|-------|
| Total modules | 162 |
| Direct dependencies | 48 |
| Indirect dependencies | 114 |

### Dependency Categories

| Category | Count | Notable |
|----------|-------|---------|
| Protobuf ecosystem | 8 | protocompile, protoreflect |
| Buf internal | 7 | buf.build/go/* |
| RPC/Networking | 4 | Connect, QUIC |
| Container | 2 | Docker, go-containerregistry |
| CLI | 3 | Cobra, pflag |
| LSP | 3 | jsonrpc2, protocol |
| Compression | 2 | klauspost/compress |
| Utilities | 10+ | uuid, yaml, etc. |
| Testing | 2 | testify, go-cmp |

### Dependency Health

| Status | Count |
|--------|-------|
| Stable (v1+) | 28 |
| Pre-stable (v0.x) | 17 |
| +incompatible | 3 |
| Pseudo-versions | 7 |
| Beta | 1 |

---

## Documentation Metrics

### Code Documentation

| Metric | Status |
|--------|--------|
| Package docs | Required (godoclint) |
| Public symbol docs | Required |
| Comment length | Max 120 chars |
| Doc format | Start with symbol name |

### Project Documentation

| Document | Lines |
|----------|-------|
| README.md | 168 |
| CHANGELOG.md | 5000+ |
| Contributing guide | Yes |
| License | Apache 2.0 |

---

## Linting Metrics

### Active Linters

| Count | Category |
|-------|----------|
| 23 | Total active linters |
| 50+ | Specific exclusions |
| 9 | Pre-commit hooks |

### Key Linters

| Linter | Purpose |
|--------|---------|
| staticcheck | Static analysis |
| gosec | Security checks |
| errcheck | Error handling |
| govet | Correctness |
| forcetypeassert | Type safety |
| bodyclose | Resource leaks |
| misspell | Spelling |
| paralleltest | Test parallelization |

### Forbidden Patterns (forbidigo)

| Pattern | Alternative |
|---------|-------------|
| errgroup.* | pkg/thread.Parallelize |
| exec.Cmd* | pkg/standard/xos/xexec |
| os.Rename | (see #639) |
| os.Getwd | pkg/osext.Getwd |
| fmt.Print* | Debug banned |
| log.* | Debug banned |

---

## Architecture Metrics

### Package Distribution

| Layer | Packages | LOC |
|-------|----------|-----|
| CLI (cmd/) | 126 | 29,910 |
| Application (buf/) | 153 | 33,477 |
| Domain (bufpkg/) | 384 | 74,946 |
| Infrastructure (pkg/) | 222 | 26,058 |

### Core Abstractions

| Abstraction | Interfaces | Implementations |
|-------------|------------|-----------------|
| Module | 5+ | 10+ |
| Image | 3+ | 5+ |
| Workspace | 3+ | 5+ |
| Controller | 1 | 1 |
| CheckClient | 1 | 3+ |

### Design Patterns

| Pattern | Usage |
|---------|-------|
| Provider/DI | 50+ providers |
| Option | 30+ option types |
| Builder | 10+ builders |
| Decorator | 20+ decorators |

---

## Performance Metrics

### Feature Counts

| Feature | Count |
|---------|-------|
| Lint rules | 40+ |
| Breaking rules | 53+ |
| CLI commands | 21 groups |
| Output formats | 6 (text, json, yaml, junit, msvs, github-actions) |

### Supported Platforms

| Platform | Architectures |
|----------|---------------|
| Linux | x86_64, ARM64, ARMv7, RISC-V, ppc64le, s390x |
| macOS | x86_64, ARM64 |
| Windows | x86_64 |
| Docker | Multi-arch |

---

## Security Metrics

### Authentication Methods

| Method | Implementation |
|--------|----------------|
| OAuth2 Device Flow | pkg/oauth2 |
| Token-based | BUF_TOKEN env |
| Netrc | pkg/netrc |
| HTTP Basic | pkg/httpauth |

### Security Controls

| Control | Status |
|---------|--------|
| Credential file perms | 0600 |
| TLS enforcement | Yes |
| Input validation | Comprehensive |
| Sensitive data logging | Blocked |

---

## CI/CD Metrics

### GitHub Workflows

| Workflow | Purpose |
|----------|---------|
| ci.yaml | Main CI pipeline |
| buf-ci.yaml | Buf checks |
| build-and-draft-release.yaml | Release builds |
| docker-publish.yaml | Docker publishing |
| windows.yaml | Windows testing |
| codeql.yaml | Security scanning |

### Build Targets

| Target | Purpose |
|--------|---------|
| make test | Full test suite |
| make lint | Run linters |
| make generate | Code generation |
| make install | Install binary |
| make cover | Coverage report |

---

## Quality Score Summary

| Category | Score | Notes |
|----------|-------|-------|
| Architecture | 9/10 | Clean layered design |
| Testing | 8/10 | Good coverage, could be higher |
| Documentation | 8/10 | Well-documented, enforced |
| Security | 9/10 | Strong security posture |
| Maintainability | 8/10 | Some large files |
| **Overall** | **8.4/10** | Production-grade |

---

## Recommendations

### Improve

1. **File Size**: Refactor formatter.go (2,493 LOC) and breaking.go (2,397 LOC)
2. **Test Coverage**: Increase from 27% to 40%+
3. **TODOs**: Review and address 285 TODO/FIXME comments
4. **Benchmarks**: Add more performance benchmarks

### Maintain

1. **Architecture**: Keep clean layer separation
2. **Linting**: Continue strict linting enforcement
3. **Security**: Maintain strong credential handling
4. **Documentation**: Keep enforcing doc requirements

---

*Generated from commit `e68c306ab3b6c39eef5abc23725d3e33d4a49cd4`*
