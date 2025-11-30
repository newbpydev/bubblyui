# Optimization Guide

This guide covers performance optimization workflows, patterns, and techniques for BubblyUI applications.

## Table of Contents

- [Optimization Workflow](#optimization-workflow)
- [Bottleneck Detection](#bottleneck-detection)
- [Render Optimization](#render-optimization)
- [Memory Optimization](#memory-optimization)
- [State Management Optimization](#state-management-optimization)
- [Common Patterns](#common-patterns)
- [Anti-Patterns](#anti-patterns)

## Optimization Workflow

Follow this systematic approach to optimize performance:

### 1. Measure First

Never optimize without measuring:

```go
prof := profiler.New()
prof.Start()

// Run your application
runApplication()

prof.Stop()
report := prof.GenerateReport()

// Analyze the report
for _, bottleneck := range report.Bottlenecks {
    fmt.Printf("[%s] %s: %s\n", 
        bottleneck.Severity, bottleneck.Location, bottleneck.Description)
}
```

### 2. Identify Bottlenecks

Use the bottleneck detector:

```go
detector := profiler.NewBottleneckDetector()

// Set thresholds for your target frame rate
detector.SetThreshold("render", 16*time.Millisecond)  // 60 FPS
detector.SetThreshold("update", 5*time.Millisecond)
detector.SetThreshold("event", 10*time.Millisecond)

// Check operations
if bottleneck := detector.Check("render", renderTime); bottleneck != nil {
    fmt.Printf("Bottleneck: %s\n", bottleneck.Description)
    fmt.Printf("Suggestion: %s\n", bottleneck.Suggestion)
}
```

### 3. Get Recommendations

Use the recommendation engine:

```go
engine := profiler.NewRecommendationEngine()
recommendations := engine.Generate(report)

for _, rec := range recommendations {
    fmt.Printf("[%s] %s\n", rec.Priority, rec.Title)
    fmt.Printf("  Description: %s\n", rec.Description)
    fmt.Printf("  Action: %s\n", rec.Action)
    fmt.Printf("  Impact: %s\n", rec.Impact)
}
```

### 4. Implement Changes

Apply optimizations based on recommendations.

### 5. Verify Improvements

Measure again to verify:

```go
func BenchmarkOptimization(b *testing.B) {
    bp := profiler.NewBenchmarkProfiler(b)
    
    for i := 0; i < b.N; i++ {
        bp.Measure(func() {
            runOptimizedCode()
        })
    }
    
    // Compare with baseline
    baseline, _ := profiler.LoadBaseline("baseline.json")
    info := bp.GetRegressionInfo(baseline)
    
    if info.TimeRegression < 0 {
        fmt.Printf("Improvement: %.2f%%\n", -info.TimeRegression*100)
    }
}
```

## Bottleneck Detection

### Automatic Detection

```go
detector := profiler.NewBottleneckDetector()

// Analyze component metrics
metrics := &profiler.PerformanceMetrics{
    Components: componentMetrics,
}

bottlenecks := detector.Detect(metrics)
for _, b := range bottlenecks {
    fmt.Printf("[%s] %s at %s\n", b.Severity, b.Type, b.Location)
    fmt.Printf("  Impact: %.2f\n", b.Impact)
    fmt.Printf("  Suggestion: %s\n", b.Suggestion)
}
```

### Pattern Analysis

Detect common anti-patterns:

```go
analyzer := profiler.NewPatternAnalyzer()

// Built-in patterns:
// - frequent_rerender: Too many renders with short duration
// - slow_render: Render time exceeds threshold
// - memory_hog: High memory usage
// - render_spike: Occasional very slow renders
// - inefficient_render: Many renders with moderate time

issues := analyzer.Analyze(componentMetrics)
for _, issue := range issues {
    fmt.Printf("Pattern: %s\n", issue.Description)
    fmt.Printf("Suggestion: %s\n", issue.Suggestion)
}
```

### Custom Patterns

Add your own detection patterns:

```go
analyzer.AddPattern(profiler.Pattern{
    Name: "excessive_events",
    Detect: func(m *profiler.ComponentMetrics) bool {
        return m.RenderCount > 100 && m.AvgRenderTime < time.Millisecond
    },
    Severity:    profiler.SeverityMedium,
    Description: "Component renders too frequently",
    Suggestion:  "Implement event debouncing or throttling",
})
```

### Threshold Monitoring

Set up real-time threshold monitoring:

```go
monitor := profiler.NewThresholdMonitor()

// Configure thresholds
monitor.SetThreshold("render", 16*time.Millisecond)
monitor.SetThreshold("update", 5*time.Millisecond)

// Set up alerts
monitor.OnAlert(func(alert *profiler.Alert) {
    log.Printf("[%s] %s exceeded threshold: %v > %v",
        alert.Severity, alert.Operation, alert.Duration, alert.Threshold)
})

// Check operations
if bottleneck := monitor.Check("render", renderTime); bottleneck != nil {
    // Handle bottleneck
}
```

## Render Optimization

### Reduce Render Frequency

**Problem**: Component renders too often

**Solution**: Implement render throttling

```go
// Use a debounced update pattern
type DebouncedComponent struct {
    lastRender time.Time
    minInterval time.Duration
}

func (c *DebouncedComponent) ShouldRender() bool {
    if time.Since(c.lastRender) < c.minInterval {
        return false
    }
    c.lastRender = time.Now()
    return true
}
```

### Optimize View Function

**Problem**: Slow View() function

**Solutions**:

1. **Cache expensive computations**:
```go
type Component struct {
    cachedOutput string
    dirty        bool
}

func (c *Component) View() string {
    if !c.dirty && c.cachedOutput != "" {
        return c.cachedOutput
    }
    c.cachedOutput = c.render()
    c.dirty = false
    return c.cachedOutput
}
```

2. **Use string builders**:
```go
func (c *Component) View() string {
    var b strings.Builder
    b.Grow(1024) // Pre-allocate
    
    b.WriteString(header)
    for _, item := range c.items {
        b.WriteString(item.Render())
    }
    b.WriteString(footer)
    
    return b.String()
}
```

3. **Avoid allocations in hot paths**:
```go
// Bad: Creates new style each render
func (c *Component) View() string {
    style := lipgloss.NewStyle().Bold(true)
    return style.Render(c.text)
}

// Good: Reuse style
var boldStyle = lipgloss.NewStyle().Bold(true)

func (c *Component) View() string {
    return boldStyle.Render(c.text)
}
```

### Lazy Rendering

Only render visible content:

```go
type VirtualList struct {
    items       []Item
    visibleStart int
    visibleEnd   int
    itemHeight   int
}

func (v *VirtualList) View() string {
    var b strings.Builder
    
    // Only render visible items
    for i := v.visibleStart; i < v.visibleEnd && i < len(v.items); i++ {
        b.WriteString(v.items[i].Render())
        b.WriteString("\n")
    }
    
    return b.String()
}
```

## Memory Optimization

### Reduce Allocations

**Problem**: High allocation rate

**Solutions**:

1. **Use sync.Pool for temporary objects**:
```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func render() string {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufferPool.Put(buf)
    }()
    
    // Use buffer
    buf.WriteString("content")
    return buf.String()
}
```

2. **Pre-allocate slices**:
```go
// Bad
items := []Item{}
for _, data := range source {
    items = append(items, processItem(data))
}

// Good
items := make([]Item, 0, len(source))
for _, data := range source {
    items = append(items, processItem(data))
}
```

3. **Reuse buffers**:
```go
type Component struct {
    renderBuf strings.Builder
}

func (c *Component) View() string {
    c.renderBuf.Reset()
    c.renderBuf.Grow(1024)
    
    // Build output
    c.renderBuf.WriteString(c.content)
    
    return c.renderBuf.String()
}
```

### Prevent Memory Leaks

1. **Clean up event handlers**:
```go
func (c *Component) Unmount() {
    // Clear all handlers
    c.handlers = nil
    
    // Cancel any goroutines
    if c.cancel != nil {
        c.cancel()
    }
}
```

2. **Use weak references for caches**:
```go
// Limit cache size
type LRUCache struct {
    maxSize int
    items   map[string]*Item
    order   []string
}

func (c *LRUCache) Set(key string, item *Item) {
    if len(c.items) >= c.maxSize {
        // Evict oldest
        oldest := c.order[0]
        delete(c.items, oldest)
        c.order = c.order[1:]
    }
    c.items[key] = item
    c.order = append(c.order, key)
}
```

3. **Monitor goroutine count**:
```go
func monitorGoroutines() {
    ticker := time.NewTicker(time.Second)
    var lastCount int
    
    for range ticker.C {
        count := runtime.NumGoroutine()
        if count > lastCount+10 {
            log.Printf("Warning: goroutine count increased from %d to %d",
                lastCount, count)
        }
        lastCount = count
    }
}
```

## State Management Optimization

### Batch State Updates

**Problem**: Multiple state updates cause multiple renders

**Solution**: Batch updates together

```go
// Bad: Multiple renders
counter.Set(counter.Get() + 1)
name.Set("new name")
items.Set(append(items.Get(), newItem))

// Good: Single render with batch
ctx.BatchUpdate(func() {
    counter.Set(counter.Get() + 1)
    name.Set("new name")
    items.Set(append(items.Get(), newItem))
})
```

### Use Computed Values

**Problem**: Recalculating derived values on every render

**Solution**: Use computed values with caching

```go
// Bad: Recalculates every render
func (c *Component) View() string {
    total := 0
    for _, item := range c.items.Get() {
        total += item.Price
    }
    return fmt.Sprintf("Total: $%d", total)
}

// Good: Computed value with automatic caching
total := bubbly.NewComputed(func() int {
    sum := 0
    for _, item := range c.items.Get() {
        sum += item.Price
    }
    return sum
})

func (c *Component) View() string {
    return fmt.Sprintf("Total: $%d", total.Get())
}
```

### Selective Reactivity

Only update what changed:

```go
// Watch specific values
bubbly.Watch(selectedIndex, func(newVal, oldVal int) {
    // Only update when selection changes
    c.updateSelection(newVal)
})
```

## Common Patterns

### Memoization

Cache expensive function results:

```go
type MemoizedRenderer struct {
    cache map[string]string
    mu    sync.RWMutex
}

func (m *MemoizedRenderer) Render(key string, fn func() string) string {
    m.mu.RLock()
    if cached, ok := m.cache[key]; ok {
        m.mu.RUnlock()
        return cached
    }
    m.mu.RUnlock()
    
    result := fn()
    
    m.mu.Lock()
    m.cache[key] = result
    m.mu.Unlock()
    
    return result
}
```

### Debouncing

Limit function call frequency:

```go
type Debouncer struct {
    delay time.Duration
    timer *time.Timer
    mu    sync.Mutex
}

func (d *Debouncer) Call(fn func()) {
    d.mu.Lock()
    defer d.mu.Unlock()
    
    if d.timer != nil {
        d.timer.Stop()
    }
    
    d.timer = time.AfterFunc(d.delay, fn)
}
```

### Throttling

Ensure minimum interval between calls:

```go
type Throttler struct {
    interval time.Duration
    lastCall time.Time
    mu       sync.Mutex
}

func (t *Throttler) Call(fn func()) bool {
    t.mu.Lock()
    defer t.mu.Unlock()
    
    if time.Since(t.lastCall) < t.interval {
        return false
    }
    
    t.lastCall = time.Now()
    fn()
    return true
}
```

### Object Pooling

Reuse objects to reduce allocations:

```go
var componentPool = sync.Pool{
    New: func() interface{} {
        return &Component{
            buffer: make([]byte, 0, 1024),
        }
    },
}

func GetComponent() *Component {
    return componentPool.Get().(*Component)
}

func PutComponent(c *Component) {
    c.Reset()
    componentPool.Put(c)
}
```

## Anti-Patterns

### ❌ Premature Optimization

```go
// Don't optimize without measuring
// Bad: Complex optimization without evidence
func (c *Component) View() string {
    // Overly complex caching logic
    // when simple approach is fast enough
}

// Good: Measure first, optimize if needed
```

### ❌ Allocating in Hot Paths

```go
// Bad: Allocates on every call
func (c *Component) View() string {
    style := lipgloss.NewStyle().Bold(true)
    return style.Render(c.text)
}

// Good: Reuse allocations
var boldStyle = lipgloss.NewStyle().Bold(true)
```

### ❌ Unnecessary Renders

```go
// Bad: Renders even when nothing changed
func (c *Component) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Always marks dirty
    c.dirty = true
    return c, nil
}

// Good: Only render when needed
func (c *Component) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    if c.handleMessage(msg) {
        c.dirty = true
    }
    return c, nil
}
```

### ❌ Blocking Operations

```go
// Bad: Blocks the main loop
func (c *Component) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    result := expensiveOperation() // Blocks!
    c.result = result
    return c, nil
}

// Good: Use commands for async operations
func (c *Component) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    return c, func() tea.Msg {
        return resultMsg{expensiveOperation()}
    }
}
```

### ❌ Memory Leaks

```go
// Bad: Never cleans up
func (c *Component) Setup() {
    go func() {
        for {
            // Runs forever, even after unmount
            c.update()
            time.Sleep(time.Second)
        }
    }()
}

// Good: Clean up on unmount
func (c *Component) Setup() {
    ctx, cancel := context.WithCancel(context.Background())
    c.cancel = cancel
    
    go func() {
        ticker := time.NewTicker(time.Second)
        defer ticker.Stop()
        
        for {
            select {
            case <-ctx.Done():
                return
            case <-ticker.C:
                c.update()
            }
        }
    }()
}
```

## Next Steps

- [Benchmarking Guide](benchmarking.md) - Measure optimization impact
- [Profiling Guide](profiling.md) - Find more bottlenecks
