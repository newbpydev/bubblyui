# Profiling Guide

This guide covers CPU and memory profiling for BubblyUI applications using the built-in profiler and Go's pprof tools.

## Table of Contents

- [CPU Profiling](#cpu-profiling)
- [Memory Profiling](#memory-profiling)
- [Render Profiling](#render-profiling)
- [Component Profiling](#component-profiling)
- [Leak Detection](#leak-detection)
- [Remote Profiling](#remote-profiling)
- [Analyzing Profiles](#analyzing-profiles)

## CPU Profiling

CPU profiling identifies hot functions that consume the most CPU time.

### Basic CPU Profiling

```go
import "github.com/newbpydev/bubblyui/pkg/bubbly/profiler"

// Create CPU profiler
cpuProf := profiler.NewCPUProfiler()

// Start profiling
err := cpuProf.Start("cpu.prof")
if err != nil {
    log.Fatal(err)
}

// Run your workload
runApplication()

// Stop profiling
cpuProf.Stop()

// Analyze with: go tool pprof cpu.prof
```

### Profiling Specific Code Sections

```go
// Profile only the critical section
cpuProf.Start("render.prof")
for i := 0; i < 1000; i++ {
    component.View()
}
cpuProf.Stop()
```

### Using the Main Profiler

```go
prof := profiler.New()
prof.Start()

// CPU profiling is integrated
// Timing data is collected automatically

prof.Stop()
report := prof.GenerateReport()

// Check CPU profile data
if report.CPUProfile != nil {
    for _, fn := range report.CPUProfile.HotFunctions {
        fmt.Printf("%s: %.2f%% (%d samples)\n", 
            fn.Name, fn.Percent, fn.Samples)
    }
}
```

### Stack Analysis

Analyze call stacks to understand CPU usage:

```go
import "github.com/google/pprof/profile"

// Load profile
p, err := profile.Parse(file)
if err != nil {
    log.Fatal(err)
}

// Analyze with StackAnalyzer
analyzer := profiler.NewStackAnalyzer()
data := analyzer.Analyze(p)

// Hot functions
for _, fn := range data.HotFunctions {
    fmt.Printf("%s: %.2f%%\n", fn.Name, fn.Percent)
}

// Call graph
for caller, callees := range data.CallGraph {
    fmt.Printf("%s calls: %v\n", caller, callees)
}
```

## Memory Profiling

Memory profiling tracks heap allocations and helps detect memory leaks.

### Basic Memory Profiling

```go
memProf := profiler.NewMemoryProfiler()

// Take baseline snapshot
memProf.TakeSnapshot()

// Run workload
runApplication()

// Take another snapshot
memProf.TakeSnapshot()

// Check memory growth
growth := memProf.GetMemoryGrowth()
fmt.Printf("Memory growth: %d bytes\n", growth)

// Write heap profile
memProf.WriteHeapProfile("heap.prof")
```

### Tracking Allocations

```go
tracker := profiler.NewMemoryTracker()

// Track allocations by location
tracker.TrackAllocation("component.state", 1024)
tracker.TrackAllocation("buffer.resize", 4096)
tracker.TrackAllocation("component.state", 512)

// Get allocation statistics
stats := tracker.GetAllocation("component.state")
fmt.Printf("Count: %d, Total: %d, Avg: %d\n",
    stats.Count, stats.TotalSize, stats.AvgSize)

// Get all allocation locations
for _, loc := range tracker.GetAllocationLocations() {
    stats := tracker.GetAllocation(loc)
    fmt.Printf("%s: %d bytes in %d allocations\n",
        loc, stats.TotalSize, stats.Count)
}
```

### Memory Snapshots

```go
memProf := profiler.NewMemoryProfiler()

// Take periodic snapshots
for i := 0; i < 10; i++ {
    memProf.TakeSnapshot()
    time.Sleep(time.Second)
}

// Analyze snapshots
snapshots := memProf.GetSnapshots()
for i, snap := range snapshots {
    fmt.Printf("Snapshot %d: HeapAlloc=%d, HeapObjects=%d\n",
        i, snap.HeapAlloc, snap.HeapObjects)
}

// Get growth metrics
heapGrowth := memProf.GetMemoryGrowth()
objectGrowth := memProf.GetHeapObjectGrowth()
```

### GC Analysis

```go
import "runtime/debug"

// Get GC statistics
var gcStats debug.GCStats
debug.ReadGCStats(&gcStats)

fmt.Printf("Last GC: %v\n", gcStats.LastGC)
fmt.Printf("Num GC: %d\n", gcStats.NumGC)
fmt.Printf("Pause Total: %v\n", gcStats.PauseTotal)

// Recent pause times
for i, pause := range gcStats.Pause {
    if i >= 5 {
        break
    }
    fmt.Printf("Pause %d: %v\n", i, pause)
}
```

## Render Profiling

Track render performance including FPS and frame timing.

### Basic Render Profiling

```go
renderProf := profiler.NewRenderProfiler()

// Record frame timing
for {
    start := time.Now()
    output := component.View()
    renderProf.RecordFrame(time.Since(start))
    
    // Render to terminal
    render(output)
}

// Get metrics
fps := renderProf.GetFPS()
dropped := renderProf.GetDroppedFramePercent()
fmt.Printf("FPS: %.1f, Dropped: %.1f%%\n", fps, dropped)
```

### Frame Analysis

```go
renderProf := profiler.NewRenderProfiler()

// Configure for 60 FPS target
config := renderProf.GetConfig()
config.TargetFPS = 60
config.DroppedFrameThreshold = 16670 * time.Microsecond // ~16.67ms
renderProf.SetConfig(config)

// Record frames
for i := 0; i < 100; i++ {
    start := time.Now()
    component.View()
    renderProf.RecordFrame(time.Since(start))
}

// Analyze frames
frames := renderProf.GetFrames()
for _, frame := range frames {
    status := "OK"
    if frame.Dropped {
        status = "DROPPED"
    }
    fmt.Printf("%v: %v [%s]\n", frame.Timestamp, frame.Duration, status)
}

// Get statistics
avgDuration := renderProf.GetAverageFrameDuration()
minDur, maxDur := renderProf.GetMinMaxFrameDuration()
fmt.Printf("Avg: %v, Min: %v, Max: %v\n", avgDuration, minDur, maxDur)
```

### FPS Calculation

```go
fpsCalc := profiler.NewFPSCalculator()

// Add FPS samples
for i := 0; i < 60; i++ {
    fpsCalc.AddSample(60.0 + float64(i%5) - 2.5)
}

// Get statistics
avg := fpsCalc.GetAverage()
min := fpsCalc.GetMin()
max := fpsCalc.GetMax()
stdDev := fpsCalc.GetStandardDeviation()

fmt.Printf("FPS: avg=%.1f, min=%.1f, max=%.1f, stddev=%.2f\n",
    avg, min, max, stdDev)

// Check stability
if fpsCalc.IsStable(5.0) { // Within 5 FPS variance
    fmt.Println("FPS is stable")
}
```

## Component Profiling

Track per-component performance metrics.

### Basic Component Tracking

```go
tracker := profiler.NewComponentTracker()

// Record render timing
start := time.Now()
output := component.View()
tracker.RecordRender(component.ID(), component.Name(), time.Since(start))

// Get metrics
metrics := tracker.GetMetrics(component.ID())
fmt.Printf("Component: %s\n", metrics.ComponentName)
fmt.Printf("Renders: %d\n", metrics.RenderCount)
fmt.Printf("Avg Time: %v\n", metrics.AvgRenderTime)
fmt.Printf("Max Time: %v\n", metrics.MaxRenderTime)
```

### Finding Slow Components

```go
tracker := profiler.NewComponentTracker()

// ... record many renders ...

// Get top 5 slowest by average render time
top := tracker.GetTopComponents(5, profiler.SortByAvgRenderTime)
for i, m := range top {
    fmt.Printf("%d. %s: avg=%v, count=%d\n",
        i+1, m.ComponentName, m.AvgRenderTime, m.RenderCount)
}

// Sort options:
// - SortByTotalRenderTime
// - SortByRenderCount
// - SortByAvgRenderTime
// - SortByMaxRenderTime
```

### Automatic Instrumentation

```go
inst := profiler.NewInstrumentor(prof)
inst.Enable()

// Option 1: Manual instrumentation
stop := inst.InstrumentRender(component)
output := component.View()
stop()

// Option 2: Wrap component
wrapped := inst.InstrumentComponent(component)
output := wrapped.View() // Automatically timed
```

## Leak Detection

Detect memory and goroutine leaks.

### Memory Leak Detection

```go
ld := profiler.NewLeakDetector()

// Configure thresholds
thresholds := ld.GetThresholds()
thresholds.HeapGrowthBytes = 1024 * 1024      // 1MB
thresholds.GoroutineGrowth = 10               // 10 goroutines
thresholds.HeapObjectGrowth = 10000           // 10K objects
ld.SetThresholds(thresholds)

// Analyze memory snapshots
memProf := profiler.NewMemoryProfiler()
memProf.TakeSnapshot()
// ... run workload ...
memProf.TakeSnapshot()

leaks := ld.DetectLeaks(memProf.GetSnapshots())
for _, leak := range leaks {
    fmt.Printf("[%s] %s: %s\n", leak.Severity, leak.Type, leak.Description)
}
```

### Goroutine Leak Detection

```go
ld := profiler.NewLeakDetector()

// Record goroutine count before
before := runtime.NumGoroutine()

// Run workload
runWorkload()

// Record goroutine count after
after := runtime.NumGoroutine()

// Check for leaks
if leak := ld.DetectGoroutineLeaks(before, after); leak != nil {
    fmt.Printf("Goroutine leak: %s (severity: %s)\n",
        leak.Description, leak.Severity)
}
```

## Remote Profiling

Enable HTTP endpoints for remote profiling access.

### Setting Up HTTP Handlers

```go
prof := profiler.New()
h := profiler.NewHTTPHandler(prof)
h.Enable() // Disabled by default for security

mux := http.NewServeMux()
h.RegisterHandlers(mux, "/debug/pprof")

// Start HTTP server
go http.ListenAndServe(":6060", mux)
```

### Available Endpoints

| Endpoint | Description |
|----------|-------------|
| `/debug/pprof/` | Index page with links |
| `/debug/pprof/profile?seconds=30` | CPU profile |
| `/debug/pprof/heap` | Heap profile |
| `/debug/pprof/goroutine?debug=1` | Goroutine stacks |
| `/debug/pprof/block` | Block profile |
| `/debug/pprof/mutex` | Mutex profile |
| `/debug/pprof/trace?seconds=5` | Execution trace |

### Remote Analysis

```bash
# Analyze CPU profile from remote server
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Analyze heap profile
go tool pprof http://localhost:6060/debug/pprof/heap

# View goroutine stacks in browser
curl http://localhost:6060/debug/pprof/goroutine?debug=1

# Collect and analyze trace
curl -o trace.out http://localhost:6060/debug/pprof/trace?seconds=5
go tool trace trace.out
```

## Analyzing Profiles

### Using go tool pprof

```bash
# Interactive mode
go tool pprof cpu.prof

# Common commands in interactive mode:
# top10       - Show top 10 functions
# list func   - Show source for function
# web         - Open graph in browser
# pdf         - Generate PDF report

# Web interface
go tool pprof -http=:8080 cpu.prof

# Compare profiles
go tool pprof -base=before.prof after.prof
```

### Flame Graphs

Generate flame graph visualizations:

```go
fgg := profiler.NewFlameGraphGenerator()

// Generate from CPU profile data
svg := fgg.GenerateSVG(cpuProfileData)

// Save to file
os.WriteFile("flamegraph.svg", []byte(svg), 0644)
```

### Timeline Visualization

Generate timeline views:

```go
tg := profiler.NewTimelineGenerator()

events := []*profiler.TimedEvent{
    profiler.AddEvent("render", profiler.EventTypeRender, t1, 5*time.Millisecond),
    profiler.AddEvent("update", profiler.EventTypeUpdate, t2, 2*time.Millisecond),
}

html := tg.GenerateHTML(events)
os.WriteFile("timeline.html", []byte(html), 0644)
```

### Generating Reports

```go
generator := profiler.NewReportGenerator()

data := &profiler.ProfileData{
    ComponentTracker: tracker,
    StartTime:        startTime,
    EndTime:          time.Now(),
}

report := generator.Generate(data)

// Export in multiple formats
exporter := profiler.NewExporter()
exporter.ExportHTML(report, "report.html")
exporter.ExportJSON(report, "report.json")
exporter.ExportCSV(report, "report.csv")
```

## Best Practices

1. **Profile representative workloads** - Use realistic data and usage patterns
2. **Take multiple samples** - Single profiles can be misleading
3. **Profile in isolation** - Minimize external factors
4. **Compare before and after** - Verify optimization impact
5. **Use appropriate tools** - CPU profiles for CPU, heap profiles for memory
6. **Monitor in production** - Use sampling to reduce overhead

## Common Issues

### Profile is empty or too small

- Increase profiling duration
- Ensure workload is running during profiling
- Check that profiler is enabled

### High overhead

- Use sampling: `WithSamplingRate(0.1)`
- Enable minimal metrics mode
- Profile shorter durations

### Inaccurate results

- Warm up before profiling
- Run multiple iterations
- Account for GC pauses

## Next Steps

- [Optimization Guide](optimization.md) - Apply profiling insights
- [Benchmarking Guide](benchmarking.md) - Measure optimization impact
