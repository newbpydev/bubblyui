# Benchmarking Guide

This guide covers writing, running, and analyzing benchmarks for BubblyUI applications.

## Table of Contents

- [Writing Benchmarks](#writing-benchmarks)
- [Using BenchmarkProfiler](#using-benchmarkprofiler)
- [Baseline Management](#baseline-management)
- [Regression Detection](#regression-detection)
- [CI/CD Integration](#cicd-integration)
- [Best Practices](#best-practices)
- [Common Patterns](#common-patterns)

## Writing Benchmarks

### Basic Benchmark

```go
func BenchmarkComponentRender(b *testing.B) {
    component := createComponent()
    component.Init()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = component.View()
    }
}
```

### Benchmark with Setup

```go
func BenchmarkComponentUpdate(b *testing.B) {
    component := createComponent()
    component.Init()
    
    msg := tea.KeyMsg{Type: tea.KeyEnter}
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        component.Update(msg)
    }
}
```

### Sub-benchmarks

```go
func BenchmarkComponent(b *testing.B) {
    b.Run("Render", func(b *testing.B) {
        component := createComponent()
        component.Init()
        
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            _ = component.View()
        }
    })
    
    b.Run("Update", func(b *testing.B) {
        component := createComponent()
        component.Init()
        msg := tea.KeyMsg{Type: tea.KeyEnter}
        
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            component.Update(msg)
        }
    })
}
```

### Table-Driven Benchmarks

```go
func BenchmarkListRender(b *testing.B) {
    sizes := []int{10, 100, 1000, 10000}
    
    for _, size := range sizes {
        b.Run(fmt.Sprintf("Size_%d", size), func(b *testing.B) {
            list := createListWithItems(size)
            list.Init()
            
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                _ = list.View()
            }
        })
    }
}
```

## Using BenchmarkProfiler

### Basic Usage

```go
func BenchmarkWithProfiler(b *testing.B) {
    bp := profiler.NewBenchmarkProfiler(b)
    
    component := createComponent()
    component.Init()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        bp.Measure(func() {
            _ = component.View()
        })
    }
    
    // Report custom metrics
    bp.ReportMetrics()
}
```

### Manual Timing

```go
func BenchmarkManualTiming(b *testing.B) {
    bp := profiler.NewBenchmarkProfiler(b)
    
    component := createComponent()
    component.Init()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        stop := bp.StartMeasurement()
        _ = component.View()
        stop()
    }
    
    // Get statistics
    stats := bp.GetStats()
    b.Logf("Mean: %v, P95: %v, P99: %v", stats.Mean, stats.P95, stats.P99)
}
```

### Getting Statistics

```go
func BenchmarkWithStats(b *testing.B) {
    bp := profiler.NewBenchmarkProfiler(b)
    
    for i := 0; i < b.N; i++ {
        bp.Measure(func() {
            doWork()
        })
    }
    
    stats := bp.GetStats()
    
    b.Logf("Iterations: %d", stats.Iterations)
    b.Logf("Mean: %v", stats.Mean)
    b.Logf("Min: %v", stats.Min)
    b.Logf("Max: %v", stats.Max)
    b.Logf("P50: %v", stats.P50)
    b.Logf("P95: %v", stats.P95)
    b.Logf("P99: %v", stats.P99)
}
```

## Baseline Management

### Creating Baselines

```go
func BenchmarkCreateBaseline(b *testing.B) {
    bp := profiler.NewBenchmarkProfiler(b)
    
    for i := 0; i < b.N; i++ {
        bp.Measure(func() {
            doWork()
        })
    }
    
    // Save baseline
    err := bp.SaveBaseline("baseline.json")
    if err != nil {
        b.Fatal(err)
    }
}
```

### Baseline Structure

```json
{
  "name": "BenchmarkComponent",
  "ns_per_op": 1500,
  "alloc_bytes": 256,
  "allocs_per_op": 3,
  "iterations": 1000000,
  "timestamp": "2024-11-30T12:00:00Z",
  "go_version": "go1.22.0",
  "goos": "linux",
  "goarch": "amd64",
  "metadata": {
    "commit": "abc123",
    "branch": "main"
  }
}
```

### Loading Baselines

```go
baseline, err := profiler.LoadBaseline("baseline.json")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Baseline: %s\n", baseline.Name)
fmt.Printf("NsPerOp: %d\n", baseline.NsPerOp)
fmt.Printf("Created: %v\n", baseline.Timestamp)
```

### Baseline with Metadata

```go
func BenchmarkWithMetadata(b *testing.B) {
    bp := profiler.NewBenchmarkProfiler(b)
    
    for i := 0; i < b.N; i++ {
        bp.Measure(func() {
            doWork()
        })
    }
    
    baseline := bp.NewBaseline("BenchmarkComponent")
    baseline.Metadata = map[string]string{
        "commit":  os.Getenv("GIT_COMMIT"),
        "branch":  os.Getenv("GIT_BRANCH"),
        "pr":      os.Getenv("PR_NUMBER"),
    }
    
    baseline.SaveToFile("baseline.json")
}
```

## Regression Detection

### Basic Regression Check

```go
func BenchmarkWithRegression(b *testing.B) {
    bp := profiler.NewBenchmarkProfiler(b)
    
    for i := 0; i < b.N; i++ {
        bp.Measure(func() {
            doWork()
        })
    }
    
    // Load baseline
    baseline, err := profiler.LoadBaseline("baseline.json")
    if err != nil {
        b.Skip("No baseline found")
    }
    
    // Check for regression (10% threshold)
    if err := bp.AssertNoRegression(baseline, 0.10); err != nil {
        b.Fatal(err)
    }
}
```

### Detailed Regression Info

```go
func BenchmarkDetailedRegression(b *testing.B) {
    bp := profiler.NewBenchmarkProfiler(b)
    
    for i := 0; i < b.N; i++ {
        bp.Measure(func() {
            doWork()
        })
    }
    
    baseline, _ := profiler.LoadBaseline("baseline.json")
    info := bp.GetRegressionInfo(baseline)
    
    if info.HasRegression {
        b.Logf("Time regression: %.2f%%", info.TimeRegression*100)
        b.Logf("Memory regression: %.2f%%", info.MemoryRegression*100)
        b.Logf("Alloc regression: %.2f%%", info.AllocRegression*100)
        b.Logf("Details: %s", info.Details)
    }
}
```

### Custom Thresholds

```go
func BenchmarkCustomThresholds(b *testing.B) {
    bp := profiler.NewBenchmarkProfiler(b)
    
    for i := 0; i < b.N; i++ {
        bp.Measure(func() {
            doWork()
        })
    }
    
    baseline, _ := profiler.LoadBaseline("baseline.json")
    info := bp.GetRegressionInfo(baseline)
    
    // Different thresholds for different metrics
    if info.TimeRegression > 0.05 { // 5% time regression
        b.Errorf("Time regression: %.2f%%", info.TimeRegression*100)
    }
    
    if info.MemoryRegression > 0.20 { // 20% memory regression
        b.Errorf("Memory regression: %.2f%%", info.MemoryRegression*100)
    }
}
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Benchmarks

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      
      - name: Download baseline
        uses: actions/download-artifact@v4
        with:
          name: benchmark-baseline
          path: .
        continue-on-error: true
      
      - name: Run benchmarks
        run: |
          go test -bench=. -benchmem -count=5 ./... | tee benchmark.txt
      
      - name: Check for regression
        run: |
          go test -run=TestBenchmarkRegression ./...
        env:
          BASELINE_FILE: baseline.json
      
      - name: Save baseline (main only)
        if: github.ref == 'refs/heads/main'
        uses: actions/upload-artifact@v4
        with:
          name: benchmark-baseline
          path: baseline.json
```

### Regression Test

```go
func TestBenchmarkRegression(t *testing.T) {
    baselineFile := os.Getenv("BASELINE_FILE")
    if baselineFile == "" {
        t.Skip("No baseline file specified")
    }
    
    baseline, err := profiler.LoadBaseline(baselineFile)
    if err != nil {
        t.Skip("Could not load baseline")
    }
    
    // Run benchmark
    result := testing.Benchmark(BenchmarkComponent)
    
    // Create profiler stats
    bp := profiler.NewBenchmarkProfiler(nil)
    bp.SetName("BenchmarkComponent")
    
    // Simulate measurements from benchmark result
    for i := 0; i < result.N; i++ {
        bp.Measure(func() {
            time.Sleep(time.Duration(result.NsPerOp()))
        })
    }
    
    // Check regression
    if err := bp.AssertNoRegression(baseline, 0.10); err != nil {
        t.Fatal(err)
    }
}
```

### Benchmark Comparison Script

```bash
#!/bin/bash

# Run benchmarks and compare with baseline
echo "Running benchmarks..."
go test -bench=. -benchmem -count=5 ./... > current.txt

if [ -f baseline.txt ]; then
    echo "Comparing with baseline..."
    benchstat baseline.txt current.txt
else
    echo "No baseline found, saving current as baseline"
    cp current.txt baseline.txt
fi
```

## Best Practices

### 1. Use Sufficient Iterations

```go
func BenchmarkComponent(b *testing.B) {
    // Let Go determine iteration count
    for i := 0; i < b.N; i++ {
        doWork()
    }
}
```

### 2. Reset Timer After Setup

```go
func BenchmarkWithSetup(b *testing.B) {
    // Setup (not measured)
    component := createExpensiveComponent()
    
    b.ResetTimer() // Start measuring here
    for i := 0; i < b.N; i++ {
        _ = component.View()
    }
}
```

### 3. Stop Timer for Cleanup

```go
func BenchmarkWithCleanup(b *testing.B) {
    for i := 0; i < b.N; i++ {
        component := createComponent()
        _ = component.View()
        
        b.StopTimer()
        component.Cleanup() // Not measured
        b.StartTimer()
    }
}
```

### 4. Report Memory Allocations

```go
func BenchmarkMemory(b *testing.B) {
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        doWork()
    }
}
```

### 5. Use benchstat for Comparison

```bash
# Run benchmarks multiple times
go test -bench=. -count=10 ./... > old.txt

# Make changes, then run again
go test -bench=. -count=10 ./... > new.txt

# Compare
benchstat old.txt new.txt
```

### 6. Avoid Compiler Optimizations

```go
var result string // Package-level to prevent optimization

func BenchmarkComponent(b *testing.B) {
    component := createComponent()
    
    for i := 0; i < b.N; i++ {
        result = component.View() // Assign to prevent optimization
    }
}
```

### 7. Warm Up Before Measuring

```go
func BenchmarkWithWarmup(b *testing.B) {
    component := createComponent()
    
    // Warm up (not measured)
    for i := 0; i < 100; i++ {
        _ = component.View()
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = component.View()
    }
}
```

## Common Patterns

### Benchmark Different Implementations

```go
func BenchmarkImplementations(b *testing.B) {
    data := generateTestData(1000)
    
    b.Run("Implementation_A", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            implementationA(data)
        }
    })
    
    b.Run("Implementation_B", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            implementationB(data)
        }
    })
}
```

### Benchmark Scaling

```go
func BenchmarkScaling(b *testing.B) {
    for _, n := range []int{10, 100, 1000, 10000} {
        b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
            data := generateTestData(n)
            
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                process(data)
            }
        })
    }
}
```

### Benchmark Concurrency

```go
func BenchmarkConcurrency(b *testing.B) {
    for _, goroutines := range []int{1, 2, 4, 8, 16} {
        b.Run(fmt.Sprintf("goroutines=%d", goroutines), func(b *testing.B) {
            b.SetParallelism(goroutines)
            b.RunParallel(func(pb *testing.PB) {
                for pb.Next() {
                    doWork()
                }
            })
        })
    }
}
```

### Benchmark with Profiler Integration

```go
func BenchmarkFullProfile(b *testing.B) {
    bp := profiler.NewBenchmarkProfiler(b)
    
    component := createComponent()
    component.Init()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        bp.Measure(func() {
            _ = component.View()
        })
    }
    
    // Report custom metrics
    bp.ReportMetrics()
    
    // Check regression if baseline exists
    if baseline, err := profiler.LoadBaseline("baseline.json"); err == nil {
        if err := bp.AssertNoRegression(baseline, 0.10); err != nil {
            b.Error(err)
        }
    }
    
    // Save new baseline on success
    if !b.Failed() {
        bp.SaveBaseline("baseline.json")
    }
}
```

## Running Benchmarks

### Basic Run

```bash
go test -bench=. ./...
```

### With Memory Stats

```bash
go test -bench=. -benchmem ./...
```

### Multiple Runs for Statistics

```bash
go test -bench=. -count=10 ./...
```

### Specific Benchmark

```bash
go test -bench=BenchmarkComponent -benchmem ./pkg/components/
```

### With CPU Profile

```bash
go test -bench=. -cpuprofile=cpu.prof ./...
go tool pprof cpu.prof
```

### With Memory Profile

```bash
go test -bench=. -memprofile=mem.prof ./...
go tool pprof mem.prof
```

### With Timeout

```bash
go test -bench=. -benchtime=10s ./...
```

## Next Steps

- [Profiling Guide](profiling.md) - Deep dive into profiling
- [Optimization Guide](optimization.md) - Apply benchmark insights
