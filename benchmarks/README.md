# Performance Benchmarks

This directory contains baseline performance benchmarks for BubblyUI composables and automated regression testing configuration.

## 📊 Overview

Performance regression testing runs automatically on all pull requests, comparing new benchmark results against the baseline to detect performance degradations.

**Key Features:**
- ✅ Automated benchmark runs on PRs
- ✅ Statistical significance testing (count=10)
- ✅ Automatic regression detection (>10% threshold)
- ✅ PR comments with detailed comparison
- ✅ Benchmark result artifacts

## 📁 Directory Structure

```
benchmarks/
├── README.md         # This file - documentation and usage guide
└── baseline.txt      # Baseline benchmark results for comparison
```

## 🔍 Understanding Benchmark Results

### Benchmark Output Format

```
BenchmarkUseState-6    340110    3488 ns/op    128 B/op    3 allocs/op
```

- **BenchmarkUseState**: Name of the benchmark
- **-6**: GOMAXPROCS (number of CPUs used)
- **340110**: Number of iterations
- **3488 ns/op**: Nanoseconds per operation (lower is better)
- **128 B/op**: Bytes allocated per operation (lower is better)
- **3 allocs/op**: Number of allocations per operation (lower is better)

### Comparison Symbols

When comparing benchmarks with `benchstat`, you'll see:

- **~**: No statistically significant change (p > 0.05)
- **+X%**: Performance degradation - operations are X% slower
- **-X%**: Performance improvement - operations are X% faster

**Example:**

```
name              old time/op    new time/op    delta
UseState-6          3488ns ± 2%    3850ns ± 3%  +10.38%
UseState_Set-6      32.1ns ± 1%    31.9ns ± 1%     ~
```

This shows:
- `UseState` got 10.38% slower (regression)
- `UseState_Set` has no significant change

## 🚦 CI/CD Integration

### Automated Workflow

The `.github/workflows/benchmark.yml` workflow runs on:
- All pull requests to `main` or `develop`
- Manual workflow dispatch

**Steps:**
1. **Run benchmarks** - Execute all benchmarks 10 times for statistical significance
2. **Compare with baseline** - Use `benchstat` to compare results
3. **Check for regressions** - Fail if performance degrades >10%
4. **Comment on PR** - Post comparison results as PR comment
5. **Upload artifacts** - Store benchmark results for 30 days

### Regression Threshold

**The CI fails if:**
- Any benchmark shows >10% performance degradation
- The degradation is statistically significant (p ≤ 0.05)

**The CI passes if:**
- All benchmarks are within 10% of baseline
- Performance improvements (faster)
- No statistically significant changes

## 🔧 Manual Benchmark Testing

### Run Benchmarks Locally

```bash
# Run all composable benchmarks
go test -bench=. -benchmem ./pkg/bubbly/composables/

# Run specific benchmark
go test -bench=BenchmarkUseState -benchmem ./pkg/bubbly/composables/

# Run with statistical significance (count=10)
go test -bench=. -benchmem -count=10 ./pkg/bubbly/composables/
```

### Compare with Baseline

```bash
# Install benchstat (if not already installed)
go install golang.org/x/perf/cmd/benchstat@latest

# Run new benchmarks
go test -bench=. -benchmem -count=10 ./pkg/bubbly/composables/ > new.txt

# Compare with baseline
benchstat benchmarks/baseline.txt new.txt

# Check for statistical significance
benchstat -delta-test=ttest benchmarks/baseline.txt new.txt
```

### Example Comparison Output

```
name                                                old time/op    new time/op    delta
UseState-6                                            3488ns ± 2%    3421ns ± 1%   -1.92%  (p=0.000 n=10+10)
UseState_Set-6                                        32.1ns ± 1%    32.3ns ± 2%     ~     (p=0.393 n=10+10)
UseState_Get-6                                        15.3ns ± 2%    15.2ns ± 1%     ~     (p=0.165 n=10+10)
UseAsync-6                                            3753ns ± 2%    3812ns ± 3%   +1.57%  (p=0.023 n=10+10)
UseForm_SetField-6                                     343ns ± 3%     355ns ± 2%   +3.50%  (p=0.001 n=10+10)
UseForm_SetField_WithCache-6                           447ns ± 2%     449ns ± 1%     ~     (p=0.529 n=10+10)

name                                                old alloc/op   new alloc/op   delta
UseState-6                                             128B ± 0%      128B ± 0%     ~     (all equal)
UseState_Set-6                                        0.00B          0.00B          ~     (all equal)
UseState_Get-6                                        0.00B          0.00B          ~     (all equal)
UseAsync-6                                             352B ± 0%      352B ± 0%     ~     (all equal)

name                                                old allocs/op  new allocs/op  delta
UseState-6                                             3.00 ± 0%      3.00 ± 0%     ~     (all equal)
UseState_Set-6                                         0.00           0.00          ~     (all equal)
```

## 🔄 Updating the Baseline

### When to Update

Update the baseline when:
- ✅ **Intentional optimizations** - You've made changes that improve performance
- ✅ **Acceptable regressions** - Performance trade-off for new features (document why)
- ✅ **Structural changes** - Major refactoring that changes performance characteristics
- ✅ **After review** - Team has reviewed and approved the performance changes

### How to Update

1. **Run benchmarks with statistical significance:**

```bash
go test -bench=. -benchmem -benchtime=1s -count=10 ./pkg/bubbly/composables/ > benchmarks/baseline.txt
```

2. **Verify the results look reasonable:**

```bash
cat benchmarks/baseline.txt
```

3. **Commit the new baseline:**

```bash
git add benchmarks/baseline.txt
git commit -m "chore(benchmarks): update baseline after [reason]

Reason: [Explain why baseline is being updated]

Performance changes:
- UseState: [describe change]
- UseForm: [describe change]
- etc.
"
```

4. **Document in PR description:**

When updating the baseline in a PR, include:
- **Why** the baseline is being updated
- **What** changed in the implementation
- **Performance impact** (improvements or acceptable regressions)
- **Trade-offs** if any (e.g., "5% slower but 20% less memory")

### Example Commit Message

```
chore(benchmarks): update baseline after metrics integration

Reason: Added optional metrics collection to composables (Task 8.5)

Performance changes:
- UseState: No significant change (~0.1ns overhead with NoOp)
- UseForm: No significant change
- UseAsync: No significant change
- Reflection cache: Cache metrics added with zero overhead

The metrics integration uses the NoOp pattern by default, resulting
in negligible performance impact (within measurement noise).
```

## ⚠️ When CI Fails

If the benchmark CI fails on your PR:

### 1. Review the Comparison

Check the PR comment or workflow logs for the benchmark comparison:
- Identify which benchmarks regressed
- Check the magnitude of regression
- Look for statistical significance

### 2. Investigate the Cause

Common causes of regressions:
- **Added functionality** - New features may have overhead
- **Allocations** - New allocations in hot paths
- **Lock contention** - Added synchronization
- **Indirect changes** - Dependencies or imports

### 3. Options

**Option A: Fix the Regression**
- Profile the code to find the bottleneck
- Optimize the hot path
- Reduce allocations
- Re-run benchmarks

**Option B: Accept the Regression (with justification)**
- Document why the regression is acceptable
- Explain the trade-offs (e.g., maintainability vs. performance)
- Get team approval
- Update the baseline with a clear commit message

**Option C: Refactor the Approach**
- Consider alternative implementations
- Use lazy initialization
- Cache more aggressively
- Defer expensive operations

### 4. Example Investigation

```bash
# Profile the slow benchmark
go test -bench=BenchmarkUseForm -benchmem -cpuprofile=cpu.prof ./pkg/bubbly/composables/

# View profile
go tool pprof cpu.prof
(pprof) top10
(pprof) list UseForm

# Check for allocations
go test -bench=BenchmarkUseForm -benchmem -memprofile=mem.prof ./pkg/bubbly/composables/
go tool pprof mem.prof
```

## 📈 Benchmark Best Practices

### Writing Good Benchmarks

1. **Isolate what you're measuring:**
```go
func BenchmarkUseState(b *testing.B) {
    ctx := createTestContext()
    
    b.ResetTimer() // Reset timer after setup
    for i := 0; i < b.N; i++ {
        _ = UseState(ctx, 42)
    }
}
```

2. **Prevent compiler optimizations:**
```go
var result UseStateReturn[int]

func BenchmarkUseState(b *testing.B) {
    var r UseStateReturn[int]
    for i := 0; i < b.N; i++ {
        r = UseState(ctx, 42)
    }
    result = r // Prevent dead code elimination
}
```

3. **Use sub-benchmarks for comparison:**
```go
func BenchmarkUseForm(b *testing.B) {
    b.Run("WithCache", func(b *testing.B) {
        // benchmark with cache
    })
    
    b.Run("WithoutCache", func(b *testing.B) {
        // benchmark without cache
    })
}
```

### Running Reliable Benchmarks

```bash
# Increase sample size for stability
go test -bench=. -benchtime=1s -count=10

# Reduce noise from other processes
go test -bench=. -benchtime=1s -count=10 -cpu=1

# Save to file for comparison
go test -bench=. -benchtime=1s -count=10 > results.txt
```

### Interpreting Results

**Good variance:** ±2-3%
```
BenchmarkUseState-6    3488ns ± 2%
```

**High variance:** ±10%+ (may indicate unstable benchmark)
```
BenchmarkUseState-6    3488ns ± 12%  # ⚠️ Too much variance
```

## 🔗 Related Documentation

- [GitHub Actions Workflow](../.github/workflows/benchmark.yml)
- [Go Benchmark Documentation](https://golang.org/pkg/testing/#hdr-Benchmarks)
- [benchstat Tool](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat)
- [Performance Testing Best Practices](https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go)

## 📝 Notes

- Benchmarks are CPU and system-dependent
- Results will vary between machines and environments
- GitHub Actions uses Ubuntu runners (may differ from local)
- Baseline is generated on AMD Ryzen 5 4500U (update as needed)
- Statistical significance requires multiple runs (count=10 minimum)

## 🎯 Performance Targets

Current performance targets for composables:

| Operation | Target | Current |
|-----------|--------|---------|
| UseState creation | < 5μs | ~3.5μs ✅ |
| UseState Set | < 50ns | ~32ns ✅ |
| UseState Get | < 20ns | ~15ns ✅ |
| UseForm creation | < 20μs | ~15μs ✅ |
| UseForm SetField | < 500ns | ~343ns ✅ |
| UseAsync creation | < 5μs | ~3.7μs ✅ |

**All targets are met! 🎉**
