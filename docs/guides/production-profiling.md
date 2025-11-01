# Production Profiling Guide

This guide covers how to profile BubblyUI applications in production using Go's built-in pprof tooling and custom composable profiling.

## 📊 Overview

Profiling helps identify performance bottlenecks, memory leaks, and resource contention in production applications. BubblyUI provides easy-to-use profiling utilities that integrate seamlessly with Go's standard profiling tools.

**What You Can Profile:**
- ✅ CPU usage (where time is spent)
- ✅ Memory allocations (heap profile)
- ✅ Goroutine activity
- ✅ Mutex contention
- ✅ Blocking operations
- ✅ Composable usage patterns

## 🔐 Security Considerations

**⚠️ CRITICAL: Profiling endpoints expose sensitive runtime information**

### Production Security Best Practices

1. **Bind to localhost only** - Never expose to public internet

```go
// ✅ GOOD - localhost only
monitoring.EnableProfiling("localhost:6060")

// ❌ BAD - exposed to network
monitoring.EnableProfiling("0.0.0.0:6060")
monitoring.EnableProfiling(":6060")
```

2. **Use SSH tunneling** for remote access

```bash
# SSH tunnel from local machine to production server
ssh -L 6060:localhost:6060 user@production-server

# Now access locally
open http://localhost:6060/debug/pprof/
```

3. **Add authentication** if needed

```go
// Custom authenticated handler
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("X-Profiling-Token")
        if token != os.Getenv("PROFILING_TOKEN") {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

4. **Enable only when needed** - Turn off after profiling

```go
// Enable temporarily
monitoring.EnableProfiling("localhost:6060")

// Capture profile
// ...

// Disable when done
monitoring.StopProfiling()
```

5. **Use firewall rules** to restrict access

```bash
# Only allow from specific IP
iptables -A INPUT -p tcp --dport 6060 -s 192.168.1.100 -j ACCEPT
iptables -A INPUT -p tcp --dport 6060 -j DROP
```

## 🚀 Quick Start

### Enable Profiling

```go
package main

import (
    "log"
    
    "github.com/newbpydev/bubblyui/pkg/bubbly"
    "github.com/newbpydev/bubblyui/pkg/bubbly/monitoring"
)

func main() {
    // Enable profiling (localhost only for security)
    if err := monitoring.EnableProfiling("localhost:6060"); err != nil {
        log.Fatalf("Failed to start profiling: %v", err)
    }
    defer monitoring.StopProfiling()
    
    log.Println("Profiling available at http://localhost:6060/debug/pprof/")
    
    // Run your application
    app := bubbly.NewApp(RootComponent)
    app.Run()
}
```

### Access Profiling Endpoints

Once profiling is enabled, these endpoints are available:

| Endpoint | Description |
|----------|-------------|
| `/debug/pprof/` | Index page with all available profiles |
| `/debug/pprof/heap` | Heap memory profile |
| `/debug/pprof/goroutine` | Goroutine stack traces |
| `/debug/pprof/profile` | CPU profile (30s default) |
| `/debug/pprof/trace` | Execution trace |
| `/debug/pprof/block` | Blocking profile |
| `/debug/pprof/mutex` | Mutex contention profile |

## 🔍 CPU Profiling

CPU profiling shows where your application spends execution time.

### Capture CPU Profile

```bash
# Capture 30-second CPU profile
curl -o cpu.prof http://localhost:6060/debug/pprof/profile?seconds=30

# Capture 60-second profile
curl -o cpu.prof http://localhost:6060/debug/pprof/profile?seconds=60
```

### Analyze CPU Profile

```bash
# Interactive analysis
go tool pprof cpu.prof

# Commands in pprof:
(pprof) top10          # Top 10 functions by CPU time
(pprof) list UseState  # Source code with CPU samples
(pprof) web            # Open flame graph in browser
(pprof) pdf            # Generate PDF visualization
```

### Example Output

```
(pprof) top10
Showing nodes accounting for 2.50s, 83.33% of 3.00s total
      flat  flat%   sum%        cum   cum%
     1.20s 40.00% 40.00%      1.50s 50.00%  UseState
     0.60s 20.00% 60.00%      0.80s 26.67%  UseForm.SetField
     0.40s 13.33% 73.33%      0.40s 13.33%  runtime.mallocgc
     0.30s 10.00% 83.33%      0.30s 10.00%  reflect.Value.Field
```

**Interpretation:**
- **flat**: Time spent in this function only
- **cum**: Cumulative time (including called functions)
- UseState is the hotspot (40% of CPU time)

### Common CPU Issues

**Issue: High CPU in UseState**
```go
// Problem: Expensive computation in composable
state := UseState(ctx, computeExpensiveValue())

// Solution: Use lazy initialization
state := UseState(ctx, defaultValue)
ctx.OnMounted(func() {
    result := computeExpensiveValue()
    state.Set(result)
})
```

**Issue: Excessive reactivity updates**
```go
// Problem: Reactive updates in hot loop
for i := 0; i < 10000; i++ {
    counter.Set(i) // Triggers reactivity each time
}

// Solution: Batch updates
counter.Set(10000) // Single update
```

## 💾 Memory Profiling

Memory profiling identifies allocation hotspots and potential leaks.

### Capture Heap Profile

```bash
# Capture heap profile
curl -o heap.prof http://localhost:6060/debug/pprof/heap

# Capture with allocations (instead of in-use memory)
curl -o alloc.prof 'http://localhost:6060/debug/pprof/heap?alloc_space=1'
```

### Analyze Heap Profile

```bash
# Interactive analysis
go tool pprof heap.prof

# Commands:
(pprof) top10           # Top 10 allocators
(pprof) list UseForm    # Source with allocations
(pprof) web             # Flame graph
(pprof) tree            # Call tree
```

### Example Output

```
(pprof) top10
Showing nodes accounting for 150MB, 75% of 200MB total
      flat  flat%   sum%        cum   cum%
      60MB 30.00% 30.00%       80MB 40.00%  UseForm
      40MB 20.00% 50.00%       50MB 25.00%  NewRef
      30MB 15.00% 65.00%       30MB 15.00%  reflect.New
      20MB 10.00% 75.00%       20MB 10.00%  bubbly.NewComputed
```

**Interpretation:**
- UseForm allocates 60MB directly
- 80MB cumulative (including called functions)
- NewRef is second largest allocator

### Common Memory Issues

**Issue: Memory leak in composable**
```go
// Problem: Slice grows unbounded
state := UseState(ctx, []string{})
ctx.On("addItem", func(item string) {
    current := state.Get()
    state.Set(append(current, item)) // Slice grows forever
})

// Solution: Limit size or clean up old entries
ctx.On("addItem", func(item string) {
    current := state.Get()
    if len(current) > 1000 {
        current = current[len(current)-500:] // Keep last 500
    }
    state.Set(append(current, item))
})
```

**Issue: Large allocations in forms**
```go
// Problem: Reflection cache grows unbounded
// Solution: Use WarmUp with expected types
cache := reflectcache.NewFieldCache()
cache.WarmUp(reflect.TypeOf(UserForm{}))
cache.WarmUp(reflect.TypeOf(OrderForm{}))
```

## 🔄 Goroutine Profiling

Goroutine profiling shows all running goroutines and their state.

### Capture Goroutine Profile

```bash
# Capture goroutine profile
curl -o goroutine.prof http://localhost:6060/debug/pprof/goroutine
```

### Analyze Goroutines

```bash
go tool pprof goroutine.prof

(pprof) top10          # Goroutines by count
(pprof) traces         # Show goroutine traces
```

### Example: Finding Goroutine Leaks

```bash
# Check number of goroutines over time
watch -n 1 'curl -s http://localhost:6060/debug/pprof/goroutine | grep "goroutine profile:"'

# Output:
goroutine profile: total 50    # t=0
goroutine profile: total 75    # t=10s (growing)
goroutine profile: total 120   # t=20s (leak!)
```

**Issue: Goroutine leak in UseAsync**
```go
// Problem: Execute() doesn't clean up goroutines
async := UseAsync(ctx, fetchData)
for i := 0; i < 1000; i++ {
    async.Execute() // Spawns 1000 goroutines
}

// Solution: Use context with timeout
fetcher := func() (*Data, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    return fetchDataWithContext(ctx)
}
async := UseAsync(ctx, fetcher)
```

## 🔒 Mutex & Block Profiling

Profile lock contention and blocking operations.

### Enable Mutex Profiling

```go
import "runtime"

func main() {
    // Enable mutex profiling
    runtime.SetMutexProfileFraction(1)
    
    // Enable block profiling
    runtime.SetBlockProfileRate(1)
    
    monitoring.EnableProfiling("localhost:6060")
    // ...
}
```

### Capture Profiles

```bash
# Mutex contention
curl -o mutex.prof http://localhost:6060/debug/pprof/mutex

# Blocking operations
curl -o block.prof http://localhost:6060/debug/pprof/block
```

### Analyze

```bash
go tool pprof mutex.prof

(pprof) top10          # Top mutex contention points
(pprof) list component # Source with contention
```

## 🎯 Composable Profiling

Profile composable usage patterns in your application.

### Profile Composables

```go
// Profile composables for 60 seconds
profile := monitoring.ProfileComposables(60 * time.Second)

// Print summary
fmt.Println(profile.Summary())

// Analyze specific composable
if stats, ok := profile.Calls["UseState"]; ok {
    fmt.Printf("UseState:\n")
    fmt.Printf("  Calls: %d\n", stats.Count)
    fmt.Printf("  Avg time: %v\n", stats.AverageTime)
    fmt.Printf("  Total alloc: %d bytes\n", stats.Allocations)
}
```

### Example Output

```
Composable Profile (1m0s):

UseState: 15420 calls, avg 3.5µs, 1974720 bytes allocated
UseForm: 3200 calls, avg 15.2µs, 819200 bytes allocated
UseAsync: 890 calls, avg 3.7µs, 313280 bytes allocated
```

### Track Composables in Production

```go
// Log composable stats periodically
go func() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        profile := monitoring.ProfileComposables(10 * time.Second)
        log.Printf("Composable stats:\n%s", profile.Summary())
    }
}()
```

## 📈 Flame Graphs

Flame graphs visualize where time/memory is spent.

### Generate Flame Graph

```bash
# Install flamegraph tool
go install github.com/google/pprof@latest

# Generate flame graph
pprof -http=:8080 cpu.prof

# Opens browser with interactive flame graph
```

### Interpret Flame Graphs

- **Width**: Time spent in function (wider = more time)
- **Height**: Call stack depth
- **Color**: Different files/packages
- **Click**: Zoom into specific call path

**Example:**
```
┌────────────────────────────────────────────┐
│          main.main                          │  ← Entry point
├────────────┬──────────────┬────────────────┤
│  UseState  │   UseForm    │   UseAsync     │  ← Composables
├────┬───────┼──────┬───────┼────────────────┤
│NewRef│Watch│SetField│Valid│   Execute      │  ← Internal calls
└────┴───────┴──────┴───────┴────────────────┘
```

## 🐛 Common Issues & Solutions

### Issue: High Memory Usage

**Symptoms:**
- Memory grows over time
- OOM kills in production

**Investigation:**
```bash
# Capture heap profile
curl -o heap.prof http://localhost:6060/debug/pprof/heap

# Find top allocators
go tool pprof -top heap.prof
```

**Solutions:**
1. Limit cache sizes
2. Clean up old reactive references
3. Use weak references for caches
4. Profile with `alloc_space` to find total allocations

### Issue: Slow Response Times

**Symptoms:**
- High latency
- CPU at 100%

**Investigation:**
```bash
# Capture CPU profile during slow period
curl -o cpu.prof http://localhost:6060/debug/pprof/profile?seconds=30

# Find hotspots
go tool pprof -top cpu.prof
```

**Solutions:**
1. Optimize hot composables
2. Reduce reactive updates
3. Cache expensive computations
4. Use lazy initialization

### Issue: Goroutine Explosion

**Symptoms:**
- Thousands of goroutines
- System slowdown

**Investigation:**
```bash
# Check goroutine count
curl -s http://localhost:6060/debug/pprof/goroutine | grep "goroutine profile:"

# Analyze goroutine stacks
go tool pprof -traces goroutine.prof
```

**Solutions:**
1. Use context with timeout
2. Limit concurrent operations
3. Use worker pools
4. Clean up goroutines in unmount hooks

## 📚 Advanced Techniques

### Continuous Profiling

Set up continuous profiling to production:

```go
// Profile periodically and upload to storage
go func() {
    ticker := time.NewTicker(10 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        // Capture CPU profile
        resp, _ := http.Get("http://localhost:6060/debug/pprof/profile?seconds=10")
        defer resp.Body.Close()
        
        // Upload to S3/GCS
        uploadProfile("cpu", time.Now(), resp.Body)
        
        // Capture heap profile
        resp, _ = http.Get("http://localhost:6060/debug/pprof/heap")
        defer resp.Body.Close()
        
        uploadProfile("heap", time.Now(), resp.Body)
    }
}()
```

### Profile Comparison

Compare profiles to track regressions:

```bash
# Capture baseline
curl -o baseline.prof http://localhost:6060/debug/pprof/profile?seconds=30

# After changes, capture new profile
curl -o current.prof http://localhost:6060/debug/pprof/profile?seconds=30

# Compare
go tool pprof -base baseline.prof current.prof

# Shows differences
(pprof) top10  # Functions with biggest changes
```

### Custom Profiling Markers

Add custom markers to identify specific operations:

```go
import "runtime/pprof"

func processLargeDataset(data []Item) {
    // Add custom label for profiling
    labels := pprof.Labels("operation", "process_dataset", "size", strconv.Itoa(len(data)))
    pprof.Do(context.Background(), labels, func(ctx context.Context) {
        // Process data
        for _, item := range data {
            UseForm(ctx, item)
        }
    })
}
```

## 🔗 External Tools

### Pyroscope (Continuous Profiling)

```go
import "github.com/pyroscope-io/client/pyroscope"

func main() {
    pyroscope.Start(pyroscope.Config{
        ApplicationName: "bubblyui-app",
        ServerAddress:   "http://pyroscope:4040",
    })
    
    // Your application
}
```

### Datadog APM

```go
import "gopkg.in/DataDog/dd-trace-go.v1/profiler"

func main() {
    profiler.Start(
        profiler.WithService("bubblyui-app"),
        profiler.WithEnv("production"),
    )
    defer profiler.Stop()
    
    // Your application
}
```

## 📋 Profiling Checklist

Before profiling:
- [ ] Enable profiling on localhost only
- [ ] Set up SSH tunnel for remote access
- [ ] Enable mutex/block profiling if needed
- [ ] Verify application is under typical load

During profiling:
- [ ] Capture multiple profiles (CPU, heap, goroutine)
- [ ] Profile during normal and peak load
- [ ] Document baseline metrics
- [ ] Save profiles for comparison

After profiling:
- [ ] Analyze top hotspots
- [ ] Identify patterns (leaks, contention, allocations)
- [ ] Create flamegraphs for visualization
- [ ] Document findings and solutions
- [ ] Disable profiling when done

## 🎓 Resources

- [Go pprof Documentation](https://pkg.go.dev/net/http/pprof)
- [Profiling Go Programs](https://go.dev/blog/pprof)
- [Flame Graphs](http://www.brendangregg.com/flamegraphs.html)
- [Go Performance Workshop](https://dave.cheney.net/high-performance-go-workshop/gopherchina-2019.html)
- [Memory Profiling in Go](https://go.dev/blog/profiling-go-programs)

## 💡 Tips & Best Practices

1. **Profile in production** - Development doesn't show real patterns
2. **Compare baselines** - Track changes over time
3. **Focus on biggest wins** - Optimize top 3 hotspots first
4. **Measure impact** - Verify optimizations actually help
5. **Document findings** - Share learnings with team
6. **Automate profiling** - Continuous profiling catches regressions
7. **Use multiple profile types** - CPU + memory + goroutines = full picture
8. **Profile under load** - Idle applications don't show issues

---

**Need help?** Check the [BubblyUI documentation](../../README.md) or [open an issue](https://github.com/newbpydev/bubblyui/issues).
