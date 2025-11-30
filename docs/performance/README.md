# BubblyUI Performance Guide

This guide covers performance profiling, optimization, and benchmarking for BubblyUI applications.

## Overview

BubblyUI includes a comprehensive performance profiler that enables:

- **Runtime Performance Analysis** - Track component render times, update cycles, and event handling
- **CPU Profiling** - Integrate with Go's pprof for CPU analysis
- **Memory Profiling** - Detect memory leaks and allocation hot spots
- **Render Performance** - Monitor FPS and frame timing
- **Bottleneck Detection** - Automatically identify performance issues
- **Optimization Recommendations** - Get actionable suggestions

## Quick Start

```go
import "github.com/newbpydev/bubblyui/pkg/bubbly/profiler"

func main() {
    // Create and start profiler
    prof := profiler.New()
    prof.Start()
    defer prof.Stop()

    // Run your BubblyUI application
    tea.NewProgram(app).Run()

    // Generate performance report
    report := prof.GenerateReport()
    exporter := profiler.NewExporter()
    exporter.ExportHTML(report, "performance-report.html")
}
```

## Documentation

| Guide | Description |
|-------|-------------|
| [Profiling Guide](profiling.md) | CPU and memory profiling with pprof integration |
| [Optimization Guide](optimization.md) | Performance optimization workflows and patterns |
| [Benchmarking Guide](benchmarking.md) | Writing and running benchmarks |

## Performance Targets

BubblyUI profiler is designed with these performance characteristics:

| Metric | Target |
|--------|--------|
| Profiler overhead (enabled) | < 3% |
| Profiler overhead (disabled) | < 0.1% |
| Timing accuracy | ±1ms |
| Data collection | < 100μs per metric |
| Report generation | < 1s |
| Memory overhead | < 10MB |

## Key Concepts

### Profiler Modes

1. **Development Mode** (default)
   - Full metric collection
   - Detailed timing statistics
   - All bottleneck detection enabled

2. **Production Mode** (`WithMinimalMetrics()`)
   - Essential metrics only
   - Reduced overhead
   - Sampling enabled

3. **Disabled Mode**
   - Near-zero overhead
   - Fast path checks only

### Metric Types

- **Timing Metrics** - Duration of operations (render, update, events)
- **Counter Metrics** - Counts of operations (render count, event count)
- **Memory Metrics** - Heap allocations, object counts
- **Render Metrics** - FPS, frame timing, dropped frames

## Configuration

### Programmatic Configuration

```go
prof := profiler.New(
    profiler.WithEnabled(true),                           // Start enabled
    profiler.WithSamplingRate(0.1),                       // 10% sampling
    profiler.WithMaxSamples(5000),                        // Limit memory
    profiler.WithMinimalMetrics(),                        // Production mode
    profiler.WithThreshold("render", 16*time.Millisecond), // 60 FPS budget
)
```

### Environment Variables

```bash
export BUBBLY_PROFILER_ENABLED=true
export BUBBLY_PROFILER_SAMPLING_RATE=0.1
export BUBBLY_PROFILER_MAX_SAMPLES=5000
export BUBBLY_PROFILER_MINIMAL_METRICS=true
```

## Common Workflows

### 1. Development Profiling

Profile during development to identify issues early:

```go
func main() {
    prof := profiler.New()
    prof.Start()
    defer func() {
        prof.Stop()
        report := prof.GenerateReport()
        exporter := profiler.NewExporter()
        exporter.ExportHTML(report, "dev-profile.html")
    }()

    tea.NewProgram(app).Run()
}
```

### 2. CI/CD Benchmarking

Detect performance regressions in CI:

```go
func BenchmarkApp(b *testing.B) {
    bp := profiler.NewBenchmarkProfiler(b)

    for i := 0; i < b.N; i++ {
        bp.Measure(func() {
            runAppCycle()
        })
    }

    baseline, _ := profiler.LoadBaseline("baseline.json")
    if err := bp.AssertNoRegression(baseline, 0.10); err != nil {
        b.Fatal(err)
    }
}
```

### 3. Production Monitoring

Minimal overhead monitoring in production:

```go
prof := profiler.New(
    profiler.WithSamplingRate(0.01),  // 1% sampling
    profiler.WithMinimalMetrics(),
)

go func() {
    ticker := time.NewTicker(5 * time.Minute)
    for range ticker.C {
        snapshot := prof.GenerateReport()
        sendToMonitoring(snapshot)
    }
}()
```

## Integration with pprof

All profiles are compatible with Go's pprof tools:

```bash
# Analyze CPU profile
go tool pprof cpu.prof

# Analyze heap profile
go tool pprof heap.prof

# Web interface
go tool pprof -http=:8080 cpu.prof

# Compare profiles
go tool pprof -base=baseline.prof current.prof
```

## HTTP Endpoints

Expose profiling endpoints for remote access:

```go
h := profiler.NewHTTPHandler(prof)
h.Enable()

mux := http.NewServeMux()
h.RegisterHandlers(mux, "/debug/pprof")
```

Available endpoints:
- `/debug/pprof/` - Index page
- `/debug/pprof/profile` - CPU profile
- `/debug/pprof/heap` - Heap profile
- `/debug/pprof/goroutine` - Goroutine stacks
- `/debug/pprof/block` - Block profile
- `/debug/pprof/mutex` - Mutex profile
- `/debug/pprof/trace` - Execution trace

## Best Practices

1. **Profile in realistic conditions** - Use production-like data and workloads
2. **Establish baselines** - Save baselines for regression detection
3. **Use sampling in production** - Reduce overhead with sampling
4. **Focus on hot paths** - Optimize the most impactful code first
5. **Measure before and after** - Verify optimizations with benchmarks
6. **Monitor continuously** - Set up production monitoring

## Troubleshooting

### High Profiler Overhead

- Enable sampling: `WithSamplingRate(0.1)`
- Use minimal metrics: `WithMinimalMetrics()`
- Reduce max samples: `WithMaxSamples(1000)`

### Inaccurate Timing

- Ensure sufficient iterations in benchmarks
- Use warm-up periods before measuring
- Account for GC pauses

### Memory Issues

- Check for goroutine leaks with `runtime.NumGoroutine()`
- Use heap profiles to find allocation hot spots
- Monitor GC pause times

## Next Steps

- [Profiling Guide](profiling.md) - Deep dive into CPU and memory profiling
- [Optimization Guide](optimization.md) - Learn optimization techniques
- [Benchmarking Guide](benchmarking.md) - Write effective benchmarks
