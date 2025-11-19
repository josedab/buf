# RFC-0003: Performance Benchmark Suite

**Status:** Draft
**Author:** Codebase Analysis
**Created:** November 18, 2025
**Effort:** 3-5 dev-days
**Category:** Quick Win

---

## Summary

Create a comprehensive benchmark suite that runs in CI to detect performance regressions and track improvements over time.

---

## Motivation

Buf's value proposition includes performance. Without continuous benchmarking:

1. **Regressions go unnoticed** - Slower builds creep in
2. **Optimization impact unknown** - Can't measure improvements
3. **No performance baseline** - New contributors have no reference
4. **Competition comparison** - Can't track vs protoc

Current state:
- 5 benchmark tests exist
- Not run in CI
- No historical tracking
- No regression detection

---

## Detailed Design

### 1. Benchmark Suite

Create comprehensive benchmarks for key operations:

```go
// private/bufpkg/bufbenchmark/benchmarks_test.go
package bufbenchmark

// Compilation benchmarks
func BenchmarkCompileSmall(b *testing.B)   { benchmarkCompile(b, "small") }    // ~10 files
func BenchmarkCompileMedium(b *testing.B)  { benchmarkCompile(b, "medium") }   // ~100 files
func BenchmarkCompileLarge(b *testing.B)   { benchmarkCompile(b, "large") }    // ~1000 files
func BenchmarkCompileGoogleapis(b *testing.B) { benchmarkCompile(b, "googleapis") }

// Linting benchmarks
func BenchmarkLintSmall(b *testing.B)  { benchmarkLint(b, "small") }
func BenchmarkLintMedium(b *testing.B) { benchmarkLint(b, "medium") }
func BenchmarkLintLarge(b *testing.B)  { benchmarkLint(b, "large") }

// Breaking change benchmarks
func BenchmarkBreakingSmall(b *testing.B)  { benchmarkBreaking(b, "small") }
func BenchmarkBreakingMedium(b *testing.B) { benchmarkBreaking(b, "medium") }
func BenchmarkBreakingLarge(b *testing.B)  { benchmarkBreaking(b, "large") }

// Generation benchmarks
func BenchmarkGenerateGoSingle(b *testing.B)    { benchmarkGenerate(b, "go", 1) }
func BenchmarkGenerateGoMultiple(b *testing.B)  { benchmarkGenerate(b, "go", 4) }

// Plugin benchmarks
func BenchmarkPluginStartupLocal(b *testing.B) { benchmarkPluginStartup(b, "local") }
func BenchmarkPluginStartupWASM(b *testing.B)  { benchmarkPluginStartup(b, "wasm") }

// Image operations
func BenchmarkImageSerialize(b *testing.B)    { benchmarkImageOp(b, "serialize") }
func BenchmarkImageDeserialize(b *testing.B)  { benchmarkImageOp(b, "deserialize") }
func BenchmarkImageFilter(b *testing.B)       { benchmarkImageOp(b, "filter") }
```

### 2. Benchmark Infrastructure

```go
// private/bufpkg/bufbenchmark/setup.go
package bufbenchmark

import (
    "testing"
)

var testDataSets = map[string]*TestData{
    "small":      loadTestData("testdata/small"),      // 10 files
    "medium":     loadTestData("testdata/medium"),     // 100 files
    "large":      loadTestData("testdata/large"),      // 1000 files
    "googleapis": loadTestData("testdata/googleapis"), // 1574 files
}

func benchmarkCompile(b *testing.B, size string) {
    data := testDataSets[size]
    b.ResetTimer()
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        _, err := compile(data.Files)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### 3. CI Integration

Add benchmark workflow:

```yaml
# .github/workflows/benchmark.yaml
name: Benchmarks

on:
  push:
    branches: [main]
  pull_request:

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Run benchmarks
        run: |
          go test -bench=. -benchmem -count=5 \
            ./private/bufpkg/bufbenchmark/... \
            | tee benchmark-results.txt

      - name: Compare with baseline
        uses: benchmark-action/github-action-benchmark@v1
        with:
          tool: 'go'
          output-file-path: benchmark-results.txt
          github-token: ${{ secrets.GITHUB_TOKEN }}
          auto-push: true
          alert-threshold: '150%'  # Alert if 50% slower
          comment-on-alert: true
          fail-on-alert: true
```

### 4. Benchmark Dashboard

GitHub Action will generate a dashboard at:
`https://bufbuild.github.io/buf/dev/bench/`

### 5. Regression Detection

Configure alerts for significant regressions:

```yaml
# Alert thresholds
thresholds:
  BenchmarkCompileSmall: 150%    # Allow 50% variance
  BenchmarkCompileLarge: 120%    # Stricter for large
  BenchmarkLintSmall: 150%
  BenchmarkGenerateGoSingle: 130%
  BenchmarkPluginStartupWASM: 200%  # More variance expected
```

---

## Example Output

### PR Comment

```markdown
## Benchmark Results

| Benchmark | Base | Head | Change |
|-----------|------|------|--------|
| CompileSmall | 45ms | 42ms | -6.7% |
| CompileLarge | 1.2s | 1.3s | +8.3% |
| LintMedium | 120ms | 115ms | -4.2% |
| PluginStartupWASM | 150ms | 145ms | -3.3% |

:warning: **CompileLarge is 8.3% slower than baseline**
```

### Historical Chart

```
Compilation Time (small)
│
│     ╭─╮
│    ╭╯ ╰─╮
│   ╭╯    ╰─────────╮
│  ╭╯               ╰─
│ ╭╯
│╭╯
└─────────────────────────
  Jan  Feb  Mar  Apr  May
```

---

## Implementation Plan

### Phase 1: Benchmark Suite (2 days)
1. Create benchmark test files
2. Set up test data (small/medium/large)
3. Implement core benchmarks

### Phase 2: CI Integration (2 days)
1. Add GitHub workflow
2. Configure benchmark action
3. Set up dashboard

### Phase 3: Alerting (1 day)
1. Configure thresholds
2. Set up PR comments
3. Document process

---

## Backwards Compatibility

**Impact:** None

Pure addition, no existing code changes.

---

## Success Criteria

- [ ] Benchmarks run on every PR
- [ ] Dashboard shows historical trends
- [ ] Regressions block PRs
- [ ] PR comments show comparison
- [ ] Baseline established for all operations

---

## Stakeholder Approval

- [ ] Engineering Lead
- [ ] Performance Team
