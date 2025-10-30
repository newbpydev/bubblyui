# User Workflow: Performance Profiler

## Developer Personas

### Persona 1: Performance Engineer (Carlos)
- **Background**: 7 years optimizing systems
- **Goal**: Identify and fix performance bottlenecks
- **Pain Point**: No visibility into TUI performance
- **Expects**: Detailed profiling data and actionable insights
- **Success**: Finds bottleneck, optimizes, 5x improvement

### Persona 2: Product Developer (Aisha)
- **Background**: 3 years Go, building production TUI app
- **Goal**: Ensure app meets performance requirements
- **Pain Point**: App feels slow, don't know why
- **Expects**: Quick identification of slow components
- **Success**: Meets 60 FPS target, smooth UX

### Persona 3: CI/CD Engineer (Tom)
- **Background**: 5 years DevOps, managing pipelines
- **Goal**: Detect performance regressions in CI
- **Pain Point**: No automated performance checks
- **Expects**: Benchmark integration, regression detection
- **Success**: CI catches regressions before production

---

## Primary User Journey: First Performance Profile

### Entry Point: App Feels Slow

**Workflow: Profiling Application Performance**

#### Step 1: Enable Profiler
**User Action**: Add profiler to application

```go
import (
    "github.com/newbpydev/bubblyui/pkg/bubbly/profiler"
)

func main() {
    // Enable profiler
    prof := profiler.New()
    prof.Start()
    defer func() {
        prof.Stop()
        
        // Generate report
        report := prof.GenerateReport()
        report.SaveHTML("performance-report.html")
    }()
    
    // Run app
    tea.NewProgram(createApp()).Run()
}
```

**System Response**:
- Profiler initialized
- Metrics collection enabled
- Minimal overhead (< 3%)
- Ready to collect data

**UI Feedback**:
- App runs normally
- No visible change
- Performance data collecting

#### Step 2: Use Application
**User Action**: Exercise app functionality

```
// User interacts with app for 2-3 minutes
// - Navigate between screens
// - Trigger all major features
// - Typical usage patterns
```

**System Response**:
- Metrics collected for every operation
- Render timing tracked
- Memory snapshots taken
- Events recorded

**Data Collected**:
- Component render times
- Update cycle durations
- Memory allocations
- Event processing times

#### Step 3: Stop and Generate Report
**User Action**: Exit application

**System Response**:
- Profiler stops
- Data aggregated
- Report generated
- HTML file created: `performance-report.html`

**Console Output**:
```
Performance Profiler Report
===========================
Duration: 2m 34s
Components analyzed: 15
Total renders: 1,247
Average FPS: 58.3

Report saved to: performance-report.html
```

#### Step 4: View Report
**User Action**: Open HTML report in browser

**Report Shows**:
```
┌─────────────────────────────────────────────────┐
│        Performance Report - MyApp               │
│        Generated: 2024-10-29 17:20:00          │
└─────────────────────────────────────────────────┘

SUMMARY
───────
• Average FPS: 58.3 (Target: 60)
• Dropped Frames: 3.2%
• Total Renders: 1,247
• Memory Usage: 24.3 MB
• Bottlenecks: 3 Critical, 5 High

TOP SLOW COMPONENTS
───────────────────
1. DataTable        ████████████████ 12.8ms avg
2. SearchResults    ████████ 6.3ms avg
3. FilterPanel      ████ 3.1ms avg

RECOMMENDATIONS
───────────────
🔴 CRITICAL: DataTable re-renders 400x, consider memoization
🟡 HIGH: SearchResults performs expensive filtering on every render
🟢 MEDIUM: Enable virtualization for large lists
```

**Journey Milestone**: ✅ Performance issues identified!

---

### Feature Journey: Deep Dive - CPU Profiling

#### Step 5: Enable CPU Profiling
**User Action**: Profile CPU-intensive operations

```go
func main() {
    prof := profiler.New()
    
    // Start CPU profiling
    prof.StartCPUProfile("cpu.prof")
    defer prof.StopCPUProfile()
    
    // Run workload
    runApp()
}
```

**System Response**:
- CPU profiling starts
- Stack samples collected
- pprof file generated
- Overhead: ~5%

#### Step 6: Analyze CPU Profile
**User Action**: Use pprof tools

```bash
$ go tool pprof cpu.prof

(pprof) top10
Showing nodes accounting for 2840ms, 71% of 4000ms total
      flat  flat%   sum%        cum   cum%
     820ms 20.50% 20.50%      820ms 20.50%  DataTable.render
     640ms 16.00% 36.50%      640ms 16.00%  regexp.Match
     380ms  9.50% 46.00%      380ms  9.50%  json.Marshal
     280ms  7.00% 53.00%     1100ms 27.50%  Component.Update
     220ms  5.50% 58.50%      220ms  5.50%  strings.Split

(pprof) list DataTable.render
Total: 4s
ROUTINE ======================== DataTable.render
     820ms      820ms (flat, cum) 20.50% of Total
         .          .   145:func (dt *DataTable) render() string {
      80ms       80ms   146:    rows := []string{}
         .          .   147:    
     740ms      740ms   148:    for _, row := range dt.data {  // HOT SPOT
         .          .   149:        rows = append(rows, dt.formatRow(row))
         .          .   150:    }
         .          .   151:    
         .          .   152:    return strings.Join(rows, "\n")
         .          .   153:}
```

**User Discovers**: Loop in `DataTable.render` is hot spot

**Journey Milestone**: ✅ Hot path identified!

---

### Feature Journey: Memory Profiling

#### Step 7: Detect Memory Leak
**User Action**: Enable memory profiling

```go
prof := profiler.New()
prof.EnableMemoryTracking()

// Take snapshots over time
for i := 0; i < 10; i++ {
    runWorkload()
    prof.TakeMemorySnapshot()
    time.Sleep(30 * time.Second)
}

// Check for leaks
leaks := prof.DetectLeaks()
for _, leak := range leaks {
    fmt.Printf("LEAK: %s - %s\n", leak.Type, leak.Description)
}
```

**System Response**:
- Memory snapshots taken
- Growth analyzed
- Leaks detected

**Output**:
```
Memory Leak Detection
═════════════════════

🔴 LEAK DETECTED: heap_growth
   Initial: 12.3 MB
   Final:   48.7 MB
   Growth:  36.4 MB over 5 minutes
   
   Likely causes:
   • Goroutine leak (15 → 142 goroutines)
   • Unclosed resources
   • Cached data accumulation

🟡 WARNING: allocation_hotspot
   Location: DataTable.formatRow
   Allocations: 12,450 / second
   Total: 24.8 MB allocated
   
   Suggestion: Pre-allocate buffers, reuse objects
```

**User Identifies**: Goroutines not being cleaned up

**Journey Milestone**: ✅ Memory leak found!

---

### Feature Journey: Optimization & Verification

#### Step 8: Optimize Code
**User Action**: Apply optimizations based on recommendations

**Before** (slow):
```go
func (dt *DataTable) render() string {
    rows := []string{}
    for _, row := range dt.data {  // Slow: reallocates every iteration
        rows = append(rows, dt.formatRow(row))
    }
    return strings.Join(rows, "\n")
}
```

**After** (optimized):
```go
func (dt *DataTable) render() string {
    // Pre-allocate with capacity
    rows := make([]string, 0, len(dt.data))
    
    // Reuse buffer for formatting
    buf := &strings.Builder{}
    
    for _, row := range dt.data {
        buf.Reset()
        dt.formatRowInto(buf, row)  // No allocation
        rows = append(rows, buf.String())
    }
    
    return strings.Join(rows, "\n")
}
```

#### Step 9: Benchmark Improvement
**User Action**: Run benchmarks before/after

```bash
$ go test -bench=BenchmarkDataTableRender -benchmem

# Before optimization:
BenchmarkDataTableRender-8    100    12.8 ms/op    2.4 MB/op    450 allocs/op

# After optimization:
BenchmarkDataTableRender-8    500     2.1 ms/op    0.8 MB/op     50 allocs/op
```

**Results**:
- **6x faster** (12.8ms → 2.1ms)
- **3x less memory** (2.4MB → 0.8MB)
- **9x fewer allocations** (450 → 50)

**Journey Milestone**: ✅ Optimization verified with data!

---

## Alternative Workflows

### Workflow A: Continuous Performance Monitoring

#### Entry: Track Performance Over Time

**Scenario**: Monitor performance in CI/CD

```go
// In test suite
func BenchmarkAppPerformance(b *testing.B) {
    prof := profiler.NewBenchmarkProfiler(b)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        prof.Measure(func() {
            runAppCycle()
        })
    }
    
    // Assert no regression
    baseline := loadBaseline()
    if prof.HasRegression(baseline, 0.10) {  // 10% threshold
        b.Errorf("Performance regression detected")
    }
}
```

**CI Output**:
```
=== BENCH BenchmarkAppPerformance
BenchmarkAppPerformance-8    1000    1.23 ms/op
Previous baseline:           1000    1.15 ms/op
Regression: +7% (within 10% threshold)
PASS
```

**Result**: Performance tracked automatically in CI

---

### Workflow B: Production Monitoring

#### Entry: Debug Production Performance

**Scenario**: Enable low-overhead profiling in production

```go
func main() {
    // Minimal overhead production profiling
    prof := profiler.New(
        profiler.WithSamplingRate(0.01),  // 1% sampling
        profiler.WithMinimalMetrics(),     // Only critical metrics
    )
    
    // Periodic snapshots
    go func() {
        ticker := time.NewTicker(5 * time.Minute)
        for range ticker.C {
            snapshot := prof.Snapshot()
            
            // Send to monitoring system
            sendToDatadog(snapshot)
            
            // Check for issues
            if snapshot.AvgFPS < 30 {
                alert("Low FPS detected in production")
            }
        }
    }()
    
    runApp()
}
```

**Monitoring Dashboard Shows**:
- Real-time FPS
- Memory usage trends
- Component render times
- Alerts on degradation

**Result**: Production performance visibility

---

## Error Recovery Workflows

### Error Flow 1: Profiler Overhead Too High

**Trigger**: App noticeably slower with profiler enabled

**User Sees**:
```
⚠️ Warning: Profiler overhead detected
Current overhead: 8.2%
Recommended: < 3%

Suggestions:
• Increase sampling rate (current: 100%)
• Disable detailed metrics
• Profile for shorter duration
```

**Recovery**:
```go
// Reduce overhead
prof := profiler.New(
    profiler.WithSamplingRate(0.1),  // 10% sampling
    profiler.WithBasicMetricsOnly(),
)
```

**Result**: Overhead reduced to < 3%

---

### Error Flow 2: Report Generation Failed

**Trigger**: Report generation error

**User Sees**:
```
Error generating performance report: out of memory

The profiler collected too much data (2.3 GB).

Solutions:
• Profile for shorter duration
• Reduce sampling rate
• Disable detailed stack traces
• Increase available memory
```

**Recovery**:
1. Profile for shorter period
2. Reduce detail level
3. Stream report generation

---

### Error Flow 3: No Bottlenecks Found

**Trigger**: App is slow but profiler finds nothing

**User Sees**:
```
Performance Report: No Critical Bottlenecks Detected

However, overall performance is below target:
• Target FPS: 60
• Actual FPS: 45
• All components under threshold individually

This suggests:
✓ Issue may be aggregate/cumulative
✓ Too many small operations
✓ External factors (system load)

Recommendations:
1. Reduce total number of components
2. Batch operations
3. Profile with higher sampling rate
```

**Recovery**: Look at aggregate effects, not individual bottlenecks

---

## State Transition Diagrams

### Profiling Session Lifecycle
```
Application Start
    ↓
Profiler Initialized
    ↓
Start() Called
    ├─ Begin metric collection
    ├─ Install hooks
    └─ Initialize storage
    ↓
Running (Collecting Data)
    ├─ Record timings
    ├─ Track memory
    ├─ Detect bottlenecks
    └─ Continuous monitoring
    ↓
Stop() Called
    ├─ Finalize metrics
    ├─ Aggregate data
    └─ Prepare for report
    ↓
Report Generation
    ├─ Analyze bottlenecks
    ├─ Generate recommendations
    ├─ Create visualizations
    └─ Export formats
    ↓
Report Delivered
    ↓
Cleanup & Exit
```

---

## Integration Points Map

### Feature Cross-Reference
```
11-performance-profiler
    ← Profiles: 01-reactivity-system (state updates)
    ← Profiles: 02-component-model (renders)
    ← Profiles: 03-lifecycle-hooks (hook timing)
    ← Profiles: 08-automatic-reactive-bridge (commands)
    → Integrates: 09-dev-tools (visualization)
    → Integrates: 10-testing-utilities (benchmarks)
    → Provides: pprof files, reports, metrics
```

---

## User Success Paths

### Path 1: Quick Profile (< 10 minutes)
```
Enable profiler → Run app → View report → Identify issue → Success! 🎉
Time to insight: 5-10 minutes
```

### Path 2: Deep Analysis (< 1 hour)
```
CPU profile → Memory profile → Optimize → Benchmark → Verify → Success! 🎉
Performance gain: 2-10x typical
```

### Path 3: Continuous Monitoring (ongoing)
```
CI benchmarks → Production monitoring → Regression alerts → Fix → Success! 🎉
Prevent performance degradation
```

---

## Common Patterns

### Pattern 1: Development Profiling
```go
func main() {
    if os.Getenv("PROFILE") == "1" {
        prof := profiler.New()
        prof.Start()
        defer func() {
            prof.Stop()
            prof.GenerateReport().SaveHTML("profile.html")
            fmt.Println("Profile saved to profile.html")
        }()
    }
    
    runApp()
}
```

### Pattern 2: Benchmark-Driven Optimization
```go
func BenchmarkComponent(b *testing.B) {
    component := createComponent()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = component.View()
    }
}

// Run: go test -bench=. -benchmem
// Optimize
// Re-run and compare
```

### Pattern 3: Production Profiling
```go
func enableProductionProfiling() {
    prof := profiler.New(
        profiler.WithSamplingRate(0.01),
        profiler.WithMinimalMetrics(),
    )
    
    http.HandleFunc("/debug/pprof/profile", func(w http.ResponseWriter, r *http.Request) {
        prof.ServeCPUProfile(w, r)
    })
}
```

---

## Tips & Tricks

### Tip 1: Profile Representative Workloads
Profile during typical usage, not artificial scenarios. Real user patterns matter.

### Tip 2: Use pprof Tools
```bash
# CPU profile analysis
go tool pprof -http=:8080 cpu.prof

# Memory profile
go tool pprof -http=:8080 mem.prof

# Compare before/after
go tool pprof -base=before.prof after.prof
```

### Tip 3: Measure Before and After
Always benchmark before optimizing. Measure after to verify improvement. Don't guess.

### Tip 4: Focus on the Biggest Wins
Optimize the slowest 20% first. Pareto principle applies to performance.

### Tip 5: Automate in CI
```yaml
# .github/workflows/benchmark.yml
- name: Run benchmarks
  run: |
    go test -bench=. -benchmem > new.txt
    benchcmp old.txt new.txt
```

---

## Performance Optimization Workflow

### Step-by-Step Process

1. **Measure Baseline**
   - Profile current performance
   - Establish baseline metrics
   - Identify top bottlenecks

2. **Hypothesize**
   - Review recommendations
   - Identify likely cause
   - Plan optimization approach

3. **Optimize**
   - Make targeted changes
   - Keep changes focused
   - Document what changed

4. **Measure Again**
   - Profile after optimization
   - Compare before/after
   - Verify improvement

5. **Iterate**
   - Continue with next bottleneck
   - Track cumulative improvements
   - Stop when targets met

---

## Summary

The Performance Profiler enables developers to identify and fix performance bottlenecks through comprehensive profiling (CPU, memory, render timing), automatic bottleneck detection with severity ranking, actionable optimization recommendations, detailed HTML reports with flame graphs, integration with Go's pprof tools, and minimal overhead (< 3%) for development profiling or (< 0.1%) production monitoring. Common workflows include development profiling (5-10 minutes to insight), deep optimization analysis (1 hour for 2-10x improvement), continuous benchmarking in CI/CD, and production monitoring for performance tracking.

**Key Success Factors**:
- ✅ Easy enablement (one-line setup)
- ✅ Comprehensive metrics (CPU, memory, render)
- ✅ Actionable insights (specific recommendations)
- ✅ Low overhead (usable in production)
- ✅ Standard tools (pprof integration)
- ✅ Clear visualizations (flame graphs, timelines)
- ✅ Proven impact (2-10x typical improvements)
