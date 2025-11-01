# Benchmark Guide

This guide covers running, analyzing, and interpreting benchmarks for BubblyUI applications, including local testing, CI/CD integration, and regression investigation.

## 📊 Overview

Benchmarking helps you understand and track performance over time. This guide covers:

- **Local benchmarking** - Running benchmarks on your machine
- **CI/CD integration** - Automated regression testing
- **Statistical analysis** - Using benchstat for comparisons
- **Baseline management** - Updating performance baselines
- **Regression investigation** - Debugging performance issues

## 🚀 Quick Start

### Run All Benchmarks

```bash
# Run all benchmarks
go test -bench=. -benchmem ./pkg/bubbly/composables/

# Sample output:
# BenchmarkUseState-6          340110    3488 ns/op    128 B/op    3 allocs/op
# BenchmarkUseState_Set-6      37250     32.1 ns/op      0 B/op    0 allocs/op
# BenchmarkUseState_Get-6      76923     15.3 ns/op      0 B/op    0 allocs/op
```

### Run Specific Benchmark

```bash
# Run UseState benchmarks only
go test -bench=UseState -benchmem ./pkg/bubbly/composables/

# Run with pattern matching
go test -bench='UseForm.*' -benchmem ./pkg/bubbly/composables/
```

## 📈 Understanding Benchmark Output

### Reading Results

```
BenchmarkUseState-6    340110    3488 ns/op    128 B/op    3 allocs/op
│                │         │          │           │          │
│                │         │          │           │          └─ Allocations per operation
│                │         │          │           └──────────── Bytes allocated per operation
│                │         │          └──────────────────────── Nanoseconds per operation
│                │         └─────────────────────────────────── Number of iterations
│                └───────────────────────────────────────────── GOMAXPROCS value
└────────────────────────────────────────────────────────────── Benchmark name
```

### Performance Metrics

**Time per operation (ns/op):**
- Lower is better
- Typical: 3-4μs for UseState
- Target: < 5μs

**Bytes per operation (B/op):**
- Lower is better
- Typical: 128B for UseState
- Target: < 200B

**Allocations per operation (allocs/op):**
- Lower is better (fewer = less GC pressure)
- Typical: 3 for UseState
- Target: < 5

### Variance

```
BenchmarkUseState-6    3488 ns/op ± 2%
                                    └─ Variance (stability indicator)
```

- **± 1-2%**: Excellent stability
- **± 3-5%**: Good stability
- **± 6-10%**: Acceptable
- **> 10%**: High variance (investigate)

## 🔬 Local Benchmarking

### Basic Benchmarks

```bash
# Run with memory statistics
go test -bench=. -benchmem ./pkg/bubbly/composables/

# Run for longer (more accurate)
go test -bench=. -benchmem -benchtime=5s ./pkg/bubbly/composables/

# Run multiple times for statistical significance
go test -bench=. -benchmem -count=10 ./pkg/bubbly/composables/
```

### Focused Benchmarks

```bash
# Run specific composable
go test -bench=BenchmarkUseState ./pkg/bubbly/composables/

# Run all form benchmarks
go test -bench='UseForm' ./pkg/bubbly/composables/

# Run multi-CPU benchmarks
go test -bench=MultiCPU ./pkg/bubbly/composables/

# Run memory growth benchmarks
go test -bench=MemoryGrowth ./pkg/bubbly/composables/
```

### Profiling While Benchmarking

```bash
# CPU profiling
go test -bench=BenchmarkUseState -cpuprofile=cpu.prof ./pkg/bubbly/composables/

# Memory profiling
go test -bench=BenchmarkUseState -memprofile=mem.prof ./pkg/bubbly/composables/

# Analyze profiles
go tool pprof cpu.prof
(pprof) top10
(pprof) web
```

## 📊 Statistical Analysis with benchstat

### Installing benchstat

```bash
go install golang.org/x/perf/cmd/benchstat@latest
```

### Comparing Benchmarks

```bash
# Capture baseline
go test -bench=. -benchmem -count=10 ./pkg/bubbly/composables/ > baseline.txt

# Make code changes...

# Capture current results
go test -bench=. -benchmem -count=10 ./pkg/bubbly/composables/ > current.txt

# Compare with benchstat
benchstat baseline.txt current.txt
```

### Example Comparison

```
name              old time/op    new time/op    delta
UseState-6          3488ns ± 2%    3150ns ± 1%   -9.69%  (p=0.000 n=10+10)
UseState_Set-6      32.1ns ± 1%    31.9ns ± 1%     ~     (p=0.393 n=10+10)
UseForm-6          15230ns ± 3%   14100ns ± 2%   -7.42%  (p=0.000 n=10+10)

name              old alloc/op   new alloc/op   delta
UseState-6           128B ± 0%      112B ± 0%  -12.50%  (p=0.000 n=10+10)
UseForm-6            300B ± 0%      280B ± 0%   -6.67%  (p=0.000 n=10+10)

name              old allocs/op  new allocs/op  delta
UseState-6           3.00 ± 0%      2.00 ± 0%  -33.33%  (p=0.000 n=10+10)
UseForm-6            5.00 ± 0%      4.00 ± 0%  -20.00%  (p=0.000 n=10+10)
```

### Interpreting benchstat Output

**Delta Symbols:**
- **~** (tilde): No statistically significant change
- **+X%**: Performance degradation (slower)
- **-X%**: Performance improvement (faster)

**P-value:**
- **p < 0.05**: Statistically significant
- **p > 0.05**: Not statistically significant (could be noise)

**Sample count (n=X+Y):**
- First number: baseline samples
- Second number: current samples
- Recommend: n=10+10 minimum

### Advanced benchstat Options

```bash
# Use t-test for significance
benchstat -delta-test=ttest baseline.txt current.txt

# Split by package
benchstat -split pkg baseline.txt current.txt

# Show only significant changes
benchstat -filter='p<0.05' baseline.txt current.txt

# Sort by magnitude
benchstat -sort delta baseline.txt current.txt
```

## 🔄 Baseline Management

### When to Update Baseline

**Update baseline when:**
- ✅ Intentional optimizations made
- ✅ New features with acceptable performance trade-offs
- ✅ Major refactoring completed
- ✅ Team has reviewed and approved changes

**Don't update baseline when:**
- ❌ Unexplained performance regressions
- ❌ Changes haven't been reviewed
- ❌ Temporary performance issues
- ❌ Without documenting why

### How to Update

```bash
# Generate new baseline
go test -bench=. -benchmem -benchtime=1s -count=10 \
  ./pkg/bubbly/composables/ > benchmarks/baseline.txt

# Verify results look reasonable
cat benchmarks/baseline.txt

# Commit with clear message
git add benchmarks/baseline.txt
git commit -m "chore(benchmarks): update baseline after reflection optimization

Performance improvements:
- UseForm SetField: -25% faster due to caching
- UseForm creation: -15% due to optimized reflection

Regression accepted:
- UseState: +2% due to metrics integration (acceptable)

Reviewed by: @team
"
```

### Baseline Documentation Template

```markdown
## Baseline Update - [Date]

**Reason:** [Optimization/Feature/Refactoring]

**Changes:**
- [Composable]: [% change] - [Explanation]
- [Composable]: [% change] - [Explanation]

**Performance Impact:**
- Improvements: [List]
- Acceptable regressions: [List with justification]

**Reviewed by:** [@team members]
**Approved by:** [@lead]
```

## 🐛 Regression Investigation

### Step 1: Identify the Regression

```bash
# Compare with baseline
benchstat benchmarks/baseline.txt new-results.txt

# Look for:
# - Changes > 10% (significant)
# - p-value < 0.05 (statistically significant)
# - Multiple benchmarks affected (systemic issue)
```

### Step 2: Isolate the Problem

```bash
# Run specific benchmark with profiling
go test -bench=BenchmarkUseForm -cpuprofile=cpu.prof ./pkg/bubbly/composables/

# Analyze CPU profile
go tool pprof cpu.prof
(pprof) top10
(pprof) list UseForm
```

### Step 3: Bisect the Changes

```bash
# Find the commit that introduced regression
git bisect start
git bisect bad HEAD
git bisect good <last-good-commit>

# For each commit:
go test -bench=BenchmarkUseForm -count=5 ./pkg/bubbly/composables/
# Mark as good/bad based on results
```

### Step 4: Analyze the Root Cause

**Common causes:**
- Added allocations in hot path
- New reflection calls
- Lock contention
- Inefficient algorithms
- Dependencies updated

**Analysis techniques:**
- CPU profiling
- Memory profiling
- Benchmark comparison
- Code review

### Step 5: Fix or Document

**Option A: Fix the regression**
```bash
# Make fix
# Verify with benchmark
go test -bench=BenchmarkUseForm -count=10 ./pkg/bubbly/composables/
benchstat baseline.txt fixed.txt
```

**Option B: Accept with justification**
```markdown
## Performance Trade-off Accepted

**Regression:** UseForm SetField +15% slower

**Reason:** Added error reporting to observability system

**Justification:**
- Error tracking is critical for production
- 15% of 343ns = 51ns absolute increase
- Still well under 500ns target
- Benefits outweigh performance cost

**Mitigation:**
- Consider async error reporting for future optimization
- Monitor in production for actual impact

**Approved by:** @lead
```

## 🤖 CI/CD Integration

### GitHub Actions Workflow

BubblyUI includes automated benchmark regression testing:

```yaml
# .github/workflows/benchmark.yml
name: Performance Benchmarks

on:
  pull_request:
    branches: [ main, develop ]
  workflow_dispatch:

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.22'
      
      - name: Run benchmarks
        run: |
          go test -bench=. -benchmem -count=10 \
            ./pkg/bubbly/composables/ > new.txt
      
      - name: Compare with baseline
        run: |
          go install golang.org/x/perf/cmd/benchstat@latest
          benchstat benchmarks/baseline.txt new.txt
      
      - name: Check for regressions
        run: |
          benchstat -delta-test=ttest benchmarks/baseline.txt new.txt | \
          grep -E '\+[0-9]{2}\.' && exit 1 || exit 0
```

### Viewing CI Results

**On Pull Requests:**
1. Navigate to the PR
2. Check "Checks" tab
3. Find "Performance Benchmarks" workflow
4. View benchmark comparison in PR comment

**Manual Runs:**
1. Go to "Actions" tab
2. Select "Performance Benchmarks"
3. Click "Run workflow"
4. View results

### Benchmark Artifacts

CI stores benchmark results as artifacts for 30 days:

```bash
# Download from GitHub UI
# Or via GitHub CLI:
gh run download <run-id> -n benchmark-results
```

## 📊 Specialized Benchmarks

### Multi-CPU Benchmarks

Test scaling characteristics:

```bash
# Run multi-CPU benchmarks
go test -bench=MultiCPU ./pkg/bubbly/composables/

# Example output:
# BenchmarkUseState_MultiCPU/cpu=1-6    33466    3991 ns/op
# BenchmarkUseState_MultiCPU/cpu=2-6    33374    3445 ns/op
# BenchmarkUseState_MultiCPU/cpu=4-6    33943    3655 ns/op
# BenchmarkUseState_MultiCPU/cpu=8-6    34005    3392 ns/op
```

**Analysis:**
- Consistent time = Single-threaded (expected for UseState)
- Decreasing time = Good parallelization
- Increasing time = Lock contention issue

### Memory Growth Benchmarks

Detect memory leaks:

```bash
# Run memory growth tests
go test -bench=MemoryGrowth ./pkg/bubbly/composables/

# Example output:
# BenchmarkMemoryGrowth_UseState-6    1    500ms    7720 mem-growth-bytes
# BenchmarkMemoryGrowth_LongRunning-6 1    2s       20552 mem-growth-bytes
```

**Analysis:**
- < 0.1 bytes/iteration: No leak
- 0.1-1 bytes/iteration: Acceptable
- > 1 bytes/iteration: Investigate
- > 10 bytes/iteration: Likely leak

### Statistical Benchmarks

Enhanced metrics with RunWithStats:

```bash
# Run with detailed stats
go test -bench=WithStats ./pkg/bubbly/composables/

# Example output:
# BenchmarkWithStats_UseState-6    26402    4213 ns/op
#   1.000 gc-runs    18446744073709549568 mem-growth-bytes
```

## 🎯 Benchmark Best Practices

### Writing Benchmarks

**Do:**
- ✅ Use `b.ResetTimer()` after setup
- ✅ Use `b.ReportAllocs()` for memory tracking
- ✅ Prevent compiler optimizations with sink variables
- ✅ Test realistic scenarios
- ✅ Document what you're measuring

**Don't:**
- ❌ Include setup time in measurements
- ❌ Use non-deterministic operations
- ❌ Benchmark trivial operations
- ❌ Ignore variance
- ❌ Run without `-benchmem`

### Running Benchmarks

**Do:**
- ✅ Run multiple times (`-count=10`)
- ✅ Use benchstat for comparisons
- ✅ Close other applications
- ✅ Use consistent hardware
- ✅ Test with realistic data

**Don't:**
- ❌ Run on busy systems
- ❌ Compare different machines
- ❌ Trust single runs
- ❌ Ignore statistical significance
- ❌ Benchmark in VMs (if avoidable)

### Interpreting Results

**Do:**
- ✅ Look for trends, not absolute numbers
- ✅ Consider statistical significance
- ✅ Compare like-with-like
- ✅ Check variance
- ✅ Document findings

**Don't:**
- ❌ Obsess over noise (< 5%)
- ❌ Ignore p-values
- ❌ Compare different benchmark runs
- ❌ Draw conclusions from single data points
- ❌ Micro-optimize without profiling

## 📚 Benchmark Examples

### Example 1: Simple Benchmark

```go
func BenchmarkUseState(b *testing.B) {
    ctx := bubbly.NewTestContext()
    b.ReportAllocs()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        state := UseState(ctx, 42)
        _ = state
    }
}
```

### Example 2: Sub-benchmarks

```go
func BenchmarkUseForm(b *testing.B) {
    sizes := []int{5, 10, 20, 50}
    
    for _, size := range sizes {
        b.Run(fmt.Sprintf("fields=%d", size), func(b *testing.B) {
            form := createFormWithFields(size)
            b.ReportAllocs()
            b.ResetTimer()
            
            for i := 0; i < b.N; i++ {
                form.SetField("Field1", "value")
            }
        })
    }
}
```

### Example 3: Parallel Benchmark

```go
func BenchmarkUseStateParallel(b *testing.B) {
    ctx := bubbly.NewTestContext()
    
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            state := UseState(ctx, 42)
            _ = state
        }
    })
}
```

## 🔗 Related Guides

- [Performance Optimization Guide](./performance-optimization.md) - Optimization strategies
- [Production Profiling Guide](./production-profiling.md) - CPU/memory profiling
- [Production Monitoring Guide](./production-monitoring.md) - Metrics and alerting

## 💡 Benchmarking Tips

1. **Benchmark early** - Establish baselines from the start
2. **Benchmark often** - Track performance continuously
3. **Use CI/CD** - Automate regression detection
4. **Statistical significance** - Always use `-count=10`
5. **Document changes** - Explain baseline updates
6. **Profile hot paths** - Understand why things are slow
7. **Compare fairly** - Same hardware, same conditions
8. **Focus on trends** - Long-term patterns > single runs

---

**Need help?** Check the [BubblyUI documentation](../../README.md) or [open an issue](https://github.com/newbpydev/bubblyui/issues).
