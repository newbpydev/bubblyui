# Design Specification: Performance Profiler

## Component Hierarchy

```
Performance Profiler System
└── Profiler
    ├── Metric Collector
    │   ├── Timing Tracker
    │   ├── Memory Tracker
    │   └── Counter Tracker
    ├── CPU Profiler (pprof integration)
    │   ├── Sample Collector
    │   └── Stack Analyzer
    ├── Memory Profiler (pprof integration)
    │   ├── Heap Analyzer
    │   └── Leak Detector
    ├── Render Profiler
    │   ├── FPS Calculator
    │   ├── Frame Analyzer
    └── Component Tracker
    ├── Bottleneck Detector
    │   ├── Threshold Monitor
    │   ├── Pattern Analyzer
    │   └── Recommendation Engine
    └── Report Generator
        ├── Data Aggregator
        ├── Visualization Builder
        └── Export System
```

---

## Architecture Overview

### High-Level Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                   Application Code                            │
│  (Components, state, events - instrumented for profiling)   │
└───────────────────────────────┬──────────────────────────────┘
                                │
┌───────────────────────────────┴──────────────────────────────┐
│                  Performance Profiler                         │
├──────────────────────────────────────────────────────────────┤
│  ┌──────────────┐    ┌─────────────────┐    ┌────────────┐  │
│  │   Metric     │───→│  Data Storage   │←───│  pprof     │  │
│  │  Collector   │    │  (In-Memory)    │    │Integration │  │
│  └──────────────┘    └─────────────────┘    └────────────┘  │
│                               │                               │
│                               ↓                               │
│                    ┌─────────────────────┐                   │
│                    │  Analysis Engine    │                   │
│                    │  (Bottlenecks, etc) │                   │
│                    └─────────────────────┘                   │
│                               │                               │
│                               ↓                               │
│                    ┌─────────────────────┐                   │
│                    │  Report Generator   │                   │
│                    └─────────────────────┘                   │
└───────────────────────────────┬──────────────────────────────┘
                                │
┌───────────────────────────────┴──────────────────────────────┐
│                     Output                                    │
│  (HTML reports, flame graphs, pprof files, metrics)         │
└──────────────────────────────────────────────────────────────┘
```

---

## Data Flow

### Profiling Data Flow

```
Application Event (render, update, etc.)
    ↓
Instrumentation Hook Captures Start Time
    ↓
Operation Executes
    ↓
Instrumentation Hook Captures End Time
    ↓
Calculate Duration
    ↓
Record Metric in Storage
    ├─ Component metrics
    ├─ Operation metrics
    └─ System metrics
    ↓
Periodic Aggregation
    ├─ Calculate statistics
    ├─ Detect bottlenecks
    └─ Update trends
    ↓
On-Demand or Periodic Reporting
    ├─ Generate flame graphs
    ├─ Create timeline
    ├─ Build recommendations
    └─ Export data
```

---

## Type Definitions

### Core Types

```go
// Profiler is the main profiler instance
type Profiler struct {
    enabled       bool
    collector     *MetricCollector
    cpuProfiler   *CPUProfiler
    memProfiler   *MemoryProfiler
    renderProfiler *RenderProfiler
    detector      *BottleneckDetector
    config        *Config
    mu            sync.RWMutex
}

// MetricCollector collects performance metrics
type MetricCollector struct {
    timings   *TimingTracker
    memory    *MemoryTracker
    counters  *CounterTracker
    mu        sync.RWMutex
}

// TimingTracker tracks operation timings
type TimingTracker struct {
    operations map[string]*TimingStats
    mu         sync.RWMutex
}

// TimingStats holds statistics for an operation
type TimingStats struct {
    Count    int64
    Total    time.Duration
    Min      time.Duration
    Max      time.Duration
    Mean     time.Duration
    P50      time.Duration
    P95      time.Duration
    P99      time.Duration
    samples  []time.Duration
}

// ComponentMetrics tracks component-specific metrics
type ComponentMetrics struct {
    ComponentID   string
    ComponentName string
    RenderCount   int64
    TotalRenderTime time.Duration
    AvgRenderTime   time.Duration
    UpdateCount   int64
    TotalUpdateTime time.Duration
    MemoryUsage   uint64
    Events        int64
}

// BottleneckInfo describes a performance bottleneck
type BottleneckInfo struct {
    Type        BottleneckType
    Location    string
    Severity    Severity
    Impact      float64
    Description string
    Suggestion  string
    Evidence    []string
}

// PerformanceReport is the complete performance analysis
type PerformanceReport struct {
    Summary      *Summary
    Components   []*ComponentMetrics
    Bottlenecks  []*BottleneckInfo
    CPUProfile   *CPUProfileData
    MemProfile   *MemProfileData
    Recommendations []*Recommendation
    Timestamp    time.Time
}
```

---

## Metric Collection Architecture

### Timing Instrumentation

```go
type MetricCollector struct {
    timings *TimingTracker
}

// Measure times an operation
func (mc *MetricCollector) Measure(name string, fn func()) {
    if !mc.enabled {
        fn()
        return
    }
    
    start := time.Now()
    fn()
    duration := time.Since(start)
    
    mc.timings.Record(name, duration)
}

// Inline timing for minimal overhead
func (mc *MetricCollector) StartTiming(name string) func() {
    if !mc.enabled {
        return func() {}
    }
    
    start := time.Now()
    return func() {
        duration := time.Since(start)
        mc.timings.Record(name, duration)
    }
}

// Usage in component:
func (c *componentImpl) View() string {
    defer c.profiler.StartTiming("component.render." + c.name)()
    
    // Render logic
    return c.template(c.renderContext)
}
```

### Timing Statistics

```go
type TimingTracker struct {
    operations map[string]*TimingStats
    mu         sync.RWMutex
}

func (tt *TimingTracker) Record(name string, duration time.Duration) {
    tt.mu.Lock()
    defer tt.mu.Unlock()
    
    stats, ok := tt.operations[name]
    if !ok {
        stats = &TimingStats{
            Min: duration,
            Max: duration,
            samples: make([]time.Duration, 0, 1000),
        }
        tt.operations[name] = stats
    }
    
    stats.Count++
    stats.Total += duration
    stats.Mean = time.Duration(int64(stats.Total) / stats.Count)
    
    if duration < stats.Min {
        stats.Min = duration
    }
    if duration > stats.Max {
        stats.Max = duration
    }
    
    // Keep samples for percentile calculation
    if len(stats.samples) < 10000 {
        stats.samples = append(stats.samples, duration)
    } else {
        // Reservoir sampling to keep memory bounded
        i := rand.Intn(int(stats.Count))
        if i < len(stats.samples) {
            stats.samples[i] = duration
        }
    }
}

func (tt *TimingTracker) GetStats(name string) *TimingStats {
    tt.mu.RLock()
    defer tt.mu.RUnlock()
    
    stats := tt.operations[name]
    if stats == nil {
        return nil
    }
    
    // Calculate percentiles
    stats.calculatePercentiles()
    
    return stats
}

func (ts *TimingStats) calculatePercentiles() {
    if len(ts.samples) == 0 {
        return
    }
    
    sorted := make([]time.Duration, len(ts.samples))
    copy(sorted, ts.samples)
    sort.Slice(sorted, func(i, j int) bool {
        return sorted[i] < sorted[j]
    })
    
    ts.P50 = sorted[len(sorted)*50/100]
    ts.P95 = sorted[len(sorted)*95/100]
    ts.P99 = sorted[len(sorted)*99/100]
}
```

---

## CPU Profiling Architecture

### pprof Integration

```go
type CPUProfiler struct {
    active   bool
    file     *os.File
    mu       sync.Mutex
}

func (cp *CPUProfiler) Start(filename string) error {
    cp.mu.Lock()
    defer cp.mu.Unlock()
    
    if cp.active {
        return errors.New("CPU profiling already active")
    }
    
    f, err := os.Create(filename)
    if err != nil {
        return err
    }
    
    cp.file = f
    cp.active = true
    
    return pprof.StartCPUProfile(f)
}

func (cp *CPUProfiler) Stop() error {
    cp.mu.Lock()
    defer cp.mu.Unlock()
    
    if !cp.active {
        return errors.New("CPU profiling not active")
    }
    
    pprof.StopCPUProfile()
    
    err := cp.file.Close()
    cp.file = nil
    cp.active = false
    
    return err
}
```

### Stack Analysis

```go
type StackAnalyzer struct {
    samples map[string]int64
}

func (sa *StackAnalyzer) Analyze(profile *pprof.Profile) *CPUProfileData {
    data := &CPUProfileData{
        HotFunctions: make([]*HotFunction, 0),
        CallGraph:    make(map[string][]string),
    }
    
    // Analyze samples
    for sample := range profile.Sample {
        for _, location := range sample.Location {
            for _, line := range location.Line {
                funcName := line.Function.Name
                sa.samples[funcName] += sample.Value[0]
            }
        }
    }
    
    // Find hot functions
    for funcName, samples := range sa.samples {
        if samples > threshold {
            data.HotFunctions = append(data.HotFunctions, &HotFunction{
                Name:    funcName,
                Samples: samples,
                Percent: float64(samples) / float64(totalSamples) * 100,
            })
        }
    }
    
    // Sort by samples
    sort.Slice(data.HotFunctions, func(i, j int) bool {
        return data.HotFunctions[i].Samples > data.HotFunctions[j].Samples
    })
    
    return data
}
```

---

## Memory Profiling Architecture

### Heap Analysis

```go
type MemoryProfiler struct {
    baseline *runtime.MemStats
    snapshots []*runtime.MemStats
}

func (mp *MemoryProfiler) TakeSnapshot() *runtime.MemStats {
    var stats runtime.MemStats
    runtime.ReadMemStats(&stats)
    
    mp.snapshots = append(mp.snapshots, &stats)
    
    return &stats
}

func (mp *MemoryProfiler) DetectLeaks() []*LeakInfo {
    leaks := []*LeakInfo{}
    
    if len(mp.snapshots) < 2 {
        return leaks
    }
    
    first := mp.snapshots[0]
    last := mp.snapshots[len(mp.snapshots)-1]
    
    // Check for memory growth
    heapGrowth := last.HeapAlloc - first.HeapAlloc
    if heapGrowth > thresholdBytes {
        leaks = append(leaks, &LeakInfo{
            Type:        "heap_growth",
            BytesLeaked: heapGrowth,
            Description: fmt.Sprintf("Heap grew by %d bytes", heapGrowth),
            Severity:    calculateSeverity(heapGrowth),
        })
    }
    
    // Check for goroutine leaks
    goroutineGrowth := last.NumGoroutine - first.NumGoroutine
    if goroutineGrowth > goroutineThreshold {
        leaks = append(leaks, &LeakInfo{
            Type:        "goroutine_leak",
            Count:       goroutineGrowth,
            Description: fmt.Sprintf("%d goroutines leaked", goroutineGrowth),
            Severity:    SeverityHigh,
        })
    }
    
    return leaks
}
```

### Allocation Tracking

```go
type AllocationTracker struct {
    allocations map[string]*AllocationStats
}

type AllocationStats struct {
    Count     int64
    TotalSize int64
    AvgSize   int64
}

func (at *AllocationTracker) TrackAllocation(location string, size int64) {
    stats, ok := at.allocations[location]
    if !ok {
        stats = &AllocationStats{}
        at.allocations[location] = stats
    }
    
    stats.Count++
    stats.TotalSize += size
    stats.AvgSize = stats.TotalSize / stats.Count
}

func (at *AllocationTracker) GetHotSpots() []*AllocationHotSpot {
    hotSpots := []*AllocationHotSpot{}
    
    for location, stats := range at.allocations {
        if stats.TotalSize > threshold {
            hotSpots = append(hotSpots, &AllocationHotSpot{
                Location:  location,
                Count:     stats.Count,
                TotalSize: stats.TotalSize,
                AvgSize:   stats.AvgSize,
            })
        }
    }
    
    // Sort by total size
    sort.Slice(hotSpots, func(i, j int) bool {
        return hotSpots[i].TotalSize > hotSpots[j].TotalSize
    })
    
    return hotSpots
}
```

---

## Render Performance Architecture

### FPS Tracking

```go
type RenderProfiler struct {
    frames      []FrameInfo
    lastFrame   time.Time
    fpsSamples  []float64
}

type FrameInfo struct {
    Timestamp time.Time
    Duration  time.Duration
    Dropped   bool
}

func (rp *RenderProfiler) RecordFrame(duration time.Duration) {
    now := time.Now()
    
    frame := FrameInfo{
        Timestamp: now,
        Duration:  duration,
    }
    
    // Check if frame was dropped (> 16.67ms for 60fps)
    if duration > 16*time.Millisecond {
        frame.Dropped = true
    }
    
    rp.frames = append(rp.frames, frame)
    
    // Calculate FPS from last second
    if !rp.lastFrame.IsZero() {
        frameDelta := now.Sub(rp.lastFrame)
        fps := 1.0 / frameDelta.Seconds()
        rp.fpsSamples = append(rp.fpsSamples, fps)
        
        // Keep only last 60 samples
        if len(rp.fpsSamples) > 60 {
            rp.fpsSamples = rp.fpsSamples[1:]
        }
    }
    
    rp.lastFrame = now
}

func (rp *RenderProfiler) GetFPS() float64 {
    if len(rp.fpsSamples) == 0 {
        return 0
    }
    
    sum := 0.0
    for _, fps := range rp.fpsSamples {
        sum += fps
    }
    
    return sum / float64(len(rp.fpsSamples))
}

func (rp *RenderProfiler) GetDroppedFramePercent() float64 {
    if len(rp.frames) == 0 {
        return 0
    }
    
    dropped := 0
    for _, frame := range rp.frames {
        if frame.Dropped {
            dropped++
        }
    }
    
    return float64(dropped) / float64(len(rp.frames)) * 100
}
```

---

## Bottleneck Detection Architecture

### Threshold Monitoring

```go
type BottleneckDetector struct {
    thresholds map[string]time.Duration
    violations map[string]int
}

func (bd *BottleneckDetector) Check(operation string, duration time.Duration) *BottleneckInfo {
    threshold, ok := bd.thresholds[operation]
    if !ok {
        threshold = defaultThreshold
    }
    
    if duration > threshold {
        bd.violations[operation]++
        
        return &BottleneckInfo{
            Type:        BottleneckTypeSlow,
            Location:    operation,
            Severity:    calculateSeverity(duration, threshold),
            Impact:      float64(duration) / float64(threshold),
            Description: fmt.Sprintf("%s took %s (threshold: %s)", 
                operation, duration, threshold),
            Suggestion:  generateSuggestion(operation, duration),
        }
    }
    
    return nil
}

func calculateSeverity(duration, threshold time.Duration) Severity {
    ratio := float64(duration) / float64(threshold)
    
    if ratio > 5.0 {
        return SeverityCritical
    } else if ratio > 3.0 {
        return SeverityHigh
    } else if ratio > 2.0 {
        return SeverityMedium
    } else {
        return SeverityLow
    }
}
```

### Pattern Analysis

```go
type PatternAnalyzer struct {
    patterns []Pattern
}

type Pattern struct {
    Name        string
    Detect      func(*ComponentMetrics) bool
    Severity    Severity
    Description string
    Suggestion  string
}

func (pa *PatternAnalyzer) Analyze(metrics *ComponentMetrics) []*BottleneckInfo {
    bottlenecks := []*BottleneckInfo{}
    
    for _, pattern := range pa.patterns {
        if pattern.Detect(metrics) {
            bottlenecks = append(bottlenecks, &BottleneckInfo{
                Type:        BottleneckTypePattern,
                Location:    metrics.ComponentName,
                Severity:    pattern.Severity,
                Description: pattern.Description,
                Suggestion:  pattern.Suggestion,
            })
        }
    }
    
    return bottlenecks
}

// Common patterns
var commonPatterns = []Pattern{
    {
        Name: "frequent_rerender",
        Detect: func(m *ComponentMetrics) bool {
            return m.RenderCount > 1000 && m.AvgRenderTime < time.Millisecond
        },
        Severity:    SeverityMedium,
        Description: "Component re-renders very frequently",
        Suggestion:  "Consider memoization or shouldComponentUpdate logic",
    },
    {
        Name: "slow_render",
        Detect: func(m *ComponentMetrics) bool {
            return m.AvgRenderTime > 10*time.Millisecond
        },
        Severity:    SeverityHigh,
        Description: "Component render is slow",
        Suggestion:  "Profile render function, optimize template",
    },
}
```

---

## Recommendation Engine Architecture

```go
type RecommendationEngine struct {
    rules []RecommendationRule
}

type RecommendationRule struct {
    Name       string
    Condition  func(*PerformanceReport) bool
    Priority   Priority
    Category   Category
    Title      string
    Description string
    Action     string
    Impact     ImpactLevel
}

func (re *RecommendationEngine) Generate(report *PerformanceReport) []*Recommendation {
    recommendations := []*Recommendation{}
    
    for _, rule := range re.rules {
        if rule.Condition(report) {
            recommendations = append(recommendations, &Recommendation{
                Title:       rule.Title,
                Description: rule.Description,
                Action:      rule.Action,
                Priority:    rule.Priority,
                Category:    rule.Category,
                Impact:      rule.Impact,
            })
        }
    }
    
    // Sort by priority
    sort.Slice(recommendations, func(i, j int) bool {
        return recommendations[i].Priority > recommendations[j].Priority
    })
    
    return recommendations
}

// Example rules
var memoizationRule = RecommendationRule{
    Name: "suggest_memoization",
    Condition: func(r *PerformanceReport) bool {
        for _, comp := range r.Components {
            if comp.RenderCount > 100 && comp.AvgRenderTime > 5*time.Millisecond {
                return true
            }
        }
        return false
    },
    Priority:    PriorityHigh,
    Category:    CategoryOptimization,
    Title:       "Implement Component Memoization",
    Description: "Some components render frequently with expensive operations",
    Action:      "Add memoization to prevent unnecessary re-renders",
    Impact:      ImpactHigh,
}
```

---

## Report Generation Architecture

### HTML Report

```go
type ReportGenerator struct {
    templates *template.Template
}

func (rg *ReportGenerator) GenerateHTML(report *PerformanceReport) (string, error) {
    var buf bytes.Buffer
    
    data := map[string]interface{}{
        "Summary":        report.Summary,
        "Components":     report.Components,
        "Bottlenecks":    report.Bottlenecks,
        "Recommendations": report.Recommendations,
        "Timestamp":      report.Timestamp,
        "FlameGraph":     rg.generateFlameGraphSVG(report.CPUProfile),
    }
    
    err := rg.templates.ExecuteTemplate(&buf, "report.html", data)
    if err != nil {
        return "", err
    }
    
    return buf.String(), nil
}
```

### Flame Graph Generation

```go
type FlameGraphGenerator struct {
    width  int
    height int
}

func (fgg *FlameGraphGenerator) Generate(profile *CPUProfileData) string {
    svg := &strings.Builder{}
    
    // SVG header
    svg.WriteString(fmt.Sprintf(`<svg width="%d" height="%d">`, fgg.width, fgg.height))
    
    // Build flame graph layers
    root := buildCallTree(profile)
    fgg.renderNode(svg, root, 0, fgg.width, 0)
    
    // SVG footer
    svg.WriteString("</svg>")
    
    return svg.String()
}

func (fgg *FlameGraphGenerator) renderNode(svg *strings.Builder, node *CallNode, x, width, depth int) {
    if width < 1 {
        return
    }
    
    y := depth * 20
    
    // Render rectangle
    svg.WriteString(fmt.Sprintf(
        `<rect x="%d" y="%d" width="%d" height="18" fill="%s"/>`,
        x, y, width, getColor(depth)))
    
    // Render text
    svg.WriteString(fmt.Sprintf(
        `<text x="%d" y="%d">%s</text>`,
        x+2, y+14, truncate(node.Name, width)))
    
    // Render children
    childX := x
    for _, child := range node.Children {
        childWidth := int(float64(width) * child.Percent / node.Percent)
        fgg.renderNode(svg, child, childX, childWidth, depth+1)
        childX += childWidth
    }
}
```

---

## Known Limitations & Solutions

### Limitation 1: Profiling Overhead
**Problem**: Profiling impacts performance  
**Current Design**: ~3% overhead target  
**Solution**: Sampling, conditional profiling, minimal instrumentation  
**Benefits**: Usable in production  
**Priority**: HIGH

### Limitation 2: Memory for Metrics
**Problem**: Storing all metrics uses memory  
**Current Design**: Bounded storage  
**Solution**: Reservoir sampling, aggregation, configurable detail level  
**Benefits**: Controlled memory usage  
**Priority**: HIGH

### Limitation 3: pprof File Size
**Problem**: CPU profile files can be large  
**Current Design**: Standard pprof format  
**Solution**: Compression, duration limits, sample rate tuning  
**Benefits**: Manageable file sizes  
**Priority**: MEDIUM

### Limitation 4: TUI-Specific Metrics
**Problem**: No standard for TUI performance  
**Current Design**: Custom metrics  
**Solution**: Define TUI-specific benchmarks, document methodology  
**Benefits**: Clear performance expectations  
**Priority**: MEDIUM

---

## Future Enhancements

### Phase 4+
1. **Real-time Profiling Dashboard**: Live performance visualization
2. **Distributed Tracing**: Track performance across services
3. **ML-Based Optimization**: AI-powered optimization suggestions
4. **Performance Budgets**: Enforce performance constraints
5. **Automated Optimization**: Auto-apply safe optimizations

---

## Summary

The Performance Profiler system provides comprehensive performance analysis through metric collection (< 3% overhead), CPU and memory profiling with pprof integration, render performance tracking with FPS calculation, bottleneck detection with threshold monitoring, and optimization recommendations from pattern analysis. The system generates detailed reports with flame graphs, timeline visualizations, and actionable suggestions, integrates with Go's built-in benchmarking tools, and supports both development profiling and production monitoring with minimal overhead.
