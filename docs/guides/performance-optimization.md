# Performance Optimization Guide

This guide covers performance optimization strategies for BubblyUI applications, including when to enable various optimizations, performance targets, and benchmarking methodology.

## 📊 Overview

BubblyUI is designed for performance out of the box, but understanding optimization strategies helps you build faster applications. This guide covers:

- **Timer pooling** - Reduce GC pressure from temporary timers
- **Reflection caching** - Speed up form field access
- **Performance targets** - Know what "good" looks like
- **Benchmarking** - Measure before and after optimizations
- **Trade-offs** - Understand the costs

## 🎯 Performance Targets

### Composable Performance

| Operation | Target | Typical | Notes |
|-----------|--------|---------|-------|
| UseState creation | < 5μs | ~3.5μs | ✅ Under target |
| UseState Set | < 50ns | ~32ns | ✅ Excellent |
| UseState Get | < 20ns | ~15ns | ✅ Excellent |
| UseForm creation | < 20μs | ~15μs | ✅ Under target |
| UseForm SetField | < 500ns | ~343ns | ✅ Good |
| UseAsync creation | < 5μs | ~3.7μs | ✅ Under target |
| Provide/Inject (1-level) | < 500ns | ~200ns | ✅ Excellent |
| Provide/Inject (10-level) | < 2μs | ~1.5μs | ✅ Acceptable |

### Memory Performance

| Metric | Target | Typical | Status |
|--------|--------|---------|--------|
| UseState allocation | < 200B | 128B | ✅ Good |
| UseForm allocation | < 500B | ~300B | ✅ Good |
| UseAsync allocation | < 500B | 352B | ✅ Good |
| Memory growth rate | < 1KB/1000 ops | < 0.1B/op | ✅ Excellent |

### Application Performance

| Metric | Target | Notes |
|--------|--------|-------|
| Component render | < 16ms | For 60 FPS |
| Event handler | < 8ms | Responsive UI |
| State update | < 1ms | Instant feedback |
| Initial load | < 100ms | First render |

## ⚡ Timer Pooling

### What is Timer Pooling?

Timer pooling reuses `time.Timer` instances instead of creating new ones for each debounce/throttle operation. This reduces GC pressure in applications with many timed operations.

**When to Enable:**
- ✅ Applications with **100+ simultaneous debounce/throttle operations**
- ✅ High-frequency timer creation (> 1000/second)
- ✅ GC pause time is a concern
- ✅ You've measured GC overhead from timers

**When NOT to Enable:**
- ❌ < 10 concurrent timers
- ❌ Timers created infrequently
- ❌ No GC pressure observed
- ❌ Before measuring (premature optimization)

### How to Enable

```go
import "github.com/newbpydev/bubblyui/pkg/bubbly/composables/timerpool"

// Enable timer pooling globally
timerpool.EnableGlobalPool(100) // Pool size of 100 timers

// Disable if needed
timerpool.DisableGlobalPool()
```

### Performance Impact

**Without Timer Pooling:**
```
BenchmarkUseDebounce-6    50000    25000 ns/op    450 B/op    5 allocs/op
GC runs: 1000/second
```

**With Timer Pooling:**
```
BenchmarkUseDebounce-6    55000    22000 ns/op    350 B/op    3 allocs/op
GC runs: 200/second
```

**Benefits:**
- ~12% faster operations
- ~22% fewer allocations
- ~80% fewer GC runs
- More predictable latency

**Trade-offs:**
- Memory: ~8KB for pool of 100 timers
- Complexity: Slightly more complex debugging
- Safety: Must ensure timers are stopped properly

### Tuning Pool Size

```go
// Too small: Frequent allocation/deallocation
timerpool.EnableGlobalPool(10)

// Good for most apps: Balanced
timerpool.EnableGlobalPool(100)

// Large apps: More headroom
timerpool.EnableGlobalPool(500)

// Excessive: Wastes memory
timerpool.EnableGlobalPool(10000)
```

**Rule of thumb:** Pool size = 2× peak concurrent timers

## 🔍 Reflection Caching

### What is Reflection Caching?

Reflection caching stores field indices and types for form structs, avoiding repeated reflection lookups. Critical for applications with heavy form usage.

**When to Enable:**
- ✅ **Always enable** for production
- ✅ Applications with forms
- ✅ > 100 form field updates/second
- ✅ Forms with many fields (> 10)

**When NOT to Enable:**
- ❌ Never (always beneficial)
- ❌ Unless debugging reflection issues

### How to Use

Reflection caching is **enabled by default** in BubblyUI. No configuration needed!

```go
// UseForm automatically uses reflection cache
form := UseForm(ctx, UserForm{}, validator)

// Cache hits after first access
form.SetField("Name", "Alice")    // Cache miss (builds cache)
form.SetField("Email", "a@b.com") // Cache hit (fast!)
form.SetField("Name", "Bob")      // Cache hit (fast!)
```

### Performance Impact

**Without Cache:**
```
BenchmarkUseForm_SetField-6    100000    15000 ns/op    800 B/op    12 allocs/op
```

**With Cache (after warm-up):**
```
BenchmarkUseForm_SetField-6    500000     343 ns/op     80 B/op     2 allocs/op
```

**Benefits:**
- ~97% faster field access
- ~90% fewer allocations
- Constant-time lookups O(1)
- Scales with form size

**Trade-offs:**
- Memory: ~100 bytes per cached form type
- Warm-up: First access slower
- Thread-safe: Lock contention possible (minimal)

### Cache Warm-Up

Pre-populate cache for critical forms:

```go
import "github.com/newbpydev/bubblyui/pkg/bubbly/composables/reflectcache"

// Warm up cache at startup
cache := reflectcache.GetGlobalCache()
cache.WarmUp(reflect.TypeOf(UserForm{}))
cache.WarmUp(reflect.TypeOf(OrderForm{}))
cache.WarmUp(reflect.TypeOf(ProductForm{}))
```

### Cache Statistics

Monitor cache effectiveness:

```go
stats := cache.Stats()
fmt.Printf("Cache hits: %d\n", stats.Hits)
fmt.Printf("Cache misses: %d\n", stats.Misses)
fmt.Printf("Hit rate: %.2f%%\n", stats.HitRate())
```

## 📈 Benchmarking Methodology

### Local Benchmarking

```bash
# Run all benchmarks
go test -bench=. -benchmem ./pkg/bubbly/composables/

# Run specific benchmark
go test -bench=BenchmarkUseState -benchmem ./pkg/bubbly/composables/

# Statistical significance (10 runs)
go test -bench=BenchmarkUseState -benchmem -count=10 ./pkg/bubbly/composables/

# CPU profiling
go test -bench=BenchmarkUseState -cpuprofile=cpu.prof ./pkg/bubbly/composables/

# Memory profiling
go test -bench=BenchmarkUseState -memprofile=mem.prof ./pkg/bubbly/composables/
```

### Baseline Comparison

```bash
# Capture baseline
go test -bench=. -benchmem -count=10 ./pkg/bubbly/composables/ > baseline.txt

# Make changes...

# Capture current
go test -bench=. -benchmem -count=10 ./pkg/bubbly/composables/ > current.txt

# Compare with benchstat
benchstat baseline.txt current.txt
```

### Example Output

```
name              old time/op    new time/op    delta
UseState-6          3488ns ± 2%    3150ns ± 1%   -9.69%  (p=0.000 n=10+10)
UseForm-6          15230ns ± 3%   14100ns ± 2%   -7.42%  (p=0.000 n=10+10)

name              old alloc/op   new alloc/op   delta
UseState-6           128B ± 0%      112B ± 0%  -12.50%  (p=0.000 n=10+10)
UseForm-6            300B ± 0%      280B ± 0%   -6.67%  (p=0.000 n=10+10)
```

**Interpreting Results:**
- **~** = No statistically significant change
- **+X%** = Slower (regression)
- **-X%** = Faster (improvement)
- **p-value** = Statistical significance (< 0.05 = significant)

### Multi-CPU Benchmarks

Test scaling characteristics:

```bash
# Run multi-CPU benchmarks
go test -bench=MultiCPU -benchmem ./pkg/bubbly/composables/
```

```
BenchmarkUseState_MultiCPU/cpu=1-6    33466    3991 ns/op
BenchmarkUseState_MultiCPU/cpu=2-6    33374    3445 ns/op
BenchmarkUseState_MultiCPU/cpu=4-6    33943    3655 ns/op
BenchmarkUseState_MultiCPU/cpu=8-6    34005    3392 ns/op
```

**Analysis:**
- Minimal scaling = Single-threaded operation (expected)
- Linear scaling = Good parallelization
- Degrading = Lock contention or resource limits

### Memory Growth Tests

Detect memory leaks:

```bash
# Run memory growth benchmarks
go test -bench=MemoryGrowth -benchtime=1s ./pkg/bubbly/composables/
```

```
BenchmarkMemoryGrowth_UseState-6    1    500ms    7720 mem-growth-bytes
BenchmarkMemoryGrowth_LongRunning-6 1    2s       20552 mem-growth-bytes
```

**Interpreting:**
- < 0.1 bytes/iteration = Excellent (no leak)
- 0.1-1 bytes/iteration = Acceptable
- > 1 bytes/iteration = Investigate potential leak
- > 10 bytes/iteration = Likely memory leak

## 🔧 Optimization Strategies

### 1. Minimize Composable Chaining

**Slow:**
```go
func UseComplexChain(ctx *bubbly.Context) {
    a := UseState(ctx, 0)
    b := UseState(ctx, a.Get())
    c := UseState(ctx, b.Get())
    d := UseState(ctx, c.Get())
    // 4 composable calls = ~14μs
}
```

**Fast:**
```go
func UseSimple(ctx *bubbly.Context) {
    state := UseState(ctx, 0)
    // 1 composable call = ~3.5μs
}
```

### 2. Batch State Updates

**Slow:**
```go
// Triggers 3 reactive updates
state1.Set(1)
state2.Set(2)
state3.Set(3)
```

**Fast:**
```go
// Batch updates if possible
type AppState struct {
    Value1 int
    Value2 int
    Value3 int
}

state := UseState(ctx, AppState{})
state.Set(AppState{Value1: 1, Value2: 2, Value3: 3})
// 1 reactive update
```

### 3. Use Computed for Derived Values

**Slow:**
```go
func GetTotal(items []*Item) int {
    total := 0
    for _, item := range items {
        total += item.Price // Recalculated every time
    }
    return total
}
```

**Fast:**
```go
items := UseState(ctx, []*Item{})
total := bubbly.NewComputed(func() int {
    t := 0
    for _, item := range items.Get() {
        t += item.Price
    }
    return t // Cached until items change
})
```

### 4. Debounce High-Frequency Updates

**Slow:**
```go
// Updates on every keystroke
input.OnChange(func(value string) {
    search(value) // API call every keystroke!
})
```

**Fast:**
```go
// Debounce to reduce API calls
debouncedSearch := UseDebounce(ctx, searchRef, 300)
input.OnChange(func(value string) {
    searchRef.Set(value)
    // search() called only after 300ms of no changes
})
```

### 5. Lazy Initialize Expensive Operations

**Slow:**
```go
// Computed on component creation
heavyData := computeExpensiveData()
state := UseState(ctx, heavyData)
```

**Fast:**
```go
// Lazy initialization
state := UseState(ctx, nil)
ctx.OnMounted(func() {
    heavyData := computeExpensiveData()
    state.Set(heavyData) // Computed only when mounted
})
```

### 6. Use Provide/Inject for Shared State

**Slow:**
```go
// Props drilling through 5 levels
func GrandParent() {
    data := getData()
    Parent(data) // Pass down
}

func Parent(data Data) {
    Child(data) // Pass down
}
// ... 3 more levels
```

**Fast:**
```go
// Provide at top level
func GrandParent(ctx *bubbly.Context) {
    data := getData()
    ctx.Provide("appData", data)
}

// Inject anywhere in tree
func DeepChild(ctx *bubbly.Context) {
    data := ctx.Inject("appData", Data{})
}
```

## 📊 Performance Monitoring

### Development Monitoring

```go
import "github.com/newbpydev/bubblyui/pkg/bubbly/monitoring"

// Set up console metrics
metrics := monitoring.NewConsoleMetrics()
monitoring.SetGlobalMetrics(metrics)

// Metrics are now logged to console
```

### Production Monitoring

```go
import "github.com/newbpydev/bubblyui/pkg/bubbly/monitoring"

// Set up Prometheus metrics
metrics := monitoring.NewPrometheusMetrics()
monitoring.SetGlobalMetrics(metrics)

// Expose /metrics endpoint
http.Handle("/metrics", promhttp.Handler())
```

### Profiling in Production

```go
import "github.com/newbpydev/bubblyui/pkg/bubbly/monitoring"

// Enable profiling (localhost only!)
if err := monitoring.EnableProfiling("localhost:6060"); err != nil {
    log.Fatal(err)
}
defer monitoring.StopProfiling()

// Access via SSH tunnel:
// ssh -L 6060:localhost:6060 user@production
// curl http://localhost:6060/debug/pprof/
```

## 🎯 Optimization Checklist

Before optimizing:
- [ ] Profile to find actual bottlenecks
- [ ] Measure baseline performance
- [ ] Set clear performance targets
- [ ] Identify root cause of slowness

When optimizing:
- [ ] Focus on biggest bottlenecks first
- [ ] Make one change at a time
- [ ] Benchmark after each change
- [ ] Verify improvement is statistically significant

After optimizing:
- [ ] Document what was changed and why
- [ ] Update performance baselines
- [ ] Monitor in production
- [ ] Share learnings with team

## 📚 Related Guides

- [Production Profiling Guide](./production-profiling.md) - CPU/memory profiling
- [Production Monitoring Guide](./production-monitoring.md) - Metrics and alerting
- [Benchmark Guide](./benchmark-guide.md) - Running and analyzing benchmarks

## 💡 Performance Tips

1. **Measure first** - Profile before optimizing
2. **Start simple** - Don't enable all optimizations upfront
3. **Know your bottlenecks** - 80/20 rule applies
4. **Test under load** - Benchmark with realistic data
5. **Monitor production** - Optimization != production performance
6. **Document decisions** - Explain why optimizations were needed
7. **Avoid premature optimization** - Clarity > speed until proven slow
8. **Use the right tool** - Timer pooling vs reflection caching vs algorithmic improvements

---

**Need help?** Check the [BubblyUI documentation](../../README.md) or [open an issue](https://github.com/newbpydev/bubblyui/issues).
