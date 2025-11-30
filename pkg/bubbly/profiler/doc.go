// Package profiler provides comprehensive performance profiling for BubblyUI applications.
//
// The profiler enables runtime performance analysis, CPU and memory profiling,
// rendering benchmarks, component performance tracking, and optimization recommendations.
// It integrates with Go's built-in pprof tools and operates with minimal overhead
// (< 3% when enabled, < 0.1% when disabled).
//
// # Quick Start
//
// Basic profiling setup:
//
//	import "github.com/newbpydev/bubblyui/pkg/bubbly/profiler"
//
//	func main() {
//	    // Create and start profiler
//	    prof := profiler.New()
//	    prof.Start()
//	    defer prof.Stop()
//
//	    // Run your BubblyUI application
//	    tea.NewProgram(app).Run()
//
//	    // Generate performance report
//	    report := prof.GenerateReport()
//	    // Save as HTML for visualization
//	    exporter := profiler.NewExporter()
//	    exporter.ExportHTML(report, "performance-report.html")
//	}
//
// # Configuration
//
// The profiler supports various configuration options:
//
//	prof := profiler.New(
//	    profiler.WithEnabled(true),                           // Start enabled
//	    profiler.WithSamplingRate(0.1),                       // 10% sampling for low overhead
//	    profiler.WithMaxSamples(5000),                        // Limit memory usage
//	    profiler.WithMinimalMetrics(),                        // Production mode
//	    profiler.WithThreshold("render", 16*time.Millisecond), // 60 FPS budget
//	)
//
// Environment variables can also configure the profiler:
//
//	BUBBLY_PROFILER_ENABLED=true
//	BUBBLY_PROFILER_SAMPLING_RATE=0.1
//	BUBBLY_PROFILER_MAX_SAMPLES=5000
//	BUBBLY_PROFILER_MINIMAL_METRICS=true
//
// Load configuration from environment:
//
//	cfg := profiler.ConfigFromEnv()
//
// # CPU Profiling
//
// CPU profiling integrates with Go's pprof:
//
//	// Start CPU profiling
//	cpuProf := profiler.NewCPUProfiler()
//	err := cpuProf.Start("cpu.prof")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Run workload
//	runApplication()
//
//	// Stop and analyze
//	cpuProf.Stop()
//	// Use: go tool pprof cpu.prof
//
// # Memory Profiling
//
// Memory profiling tracks heap allocations and detects leaks:
//
//	memProf := profiler.NewMemoryProfiler()
//
//	// Take snapshots over time
//	memProf.TakeSnapshot()
//	runWorkload()
//	memProf.TakeSnapshot()
//
//	// Check for memory growth
//	growth := memProf.GetMemoryGrowth()
//	if growth > 10*1024*1024 { // 10MB
//	    log.Println("Warning: significant memory growth detected")
//	}
//
//	// Write heap profile for analysis
//	memProf.WriteHeapProfile("heap.prof")
//	// Use: go tool pprof heap.prof
//
// # Memory Leak Detection
//
// The leak detector analyzes memory snapshots:
//
//	ld := profiler.NewLeakDetector()
//
//	// Analyze snapshots for leaks
//	leaks := ld.DetectLeaks(memProf.GetSnapshots())
//	for _, leak := range leaks {
//	    fmt.Printf("Leak: %s (severity: %s)\n", leak.Description, leak.Severity)
//	}
//
//	// Check for goroutine leaks
//	before := runtime.NumGoroutine()
//	runWorkload()
//	after := runtime.NumGoroutine()
//	if leak := ld.DetectGoroutineLeaks(before, after); leak != nil {
//	    fmt.Printf("Goroutine leak: %s\n", leak.Description)
//	}
//
// # Render Performance
//
// Track frames per second and render timing:
//
//	renderProf := profiler.NewRenderProfiler()
//
//	// Record frame timing
//	for {
//	    start := time.Now()
//	    component.View()
//	    renderProf.RecordFrame(time.Since(start))
//	}
//
//	// Get metrics
//	fps := renderProf.GetFPS()
//	dropped := renderProf.GetDroppedFramePercent()
//	fmt.Printf("FPS: %.1f, Dropped: %.1f%%\n", fps, dropped)
//
// # Component Tracking
//
// Track per-component performance:
//
//	tracker := profiler.NewComponentTracker()
//
//	// Record render timing
//	start := time.Now()
//	output := component.View()
//	tracker.RecordRender(component.ID(), component.Name(), time.Since(start))
//
//	// Get metrics
//	metrics := tracker.GetMetrics(component.ID())
//	fmt.Printf("Component %s: %d renders, avg %v\n",
//	    metrics.ComponentName, metrics.RenderCount, metrics.AvgRenderTime)
//
//	// Find slowest components
//	top := tracker.GetTopComponents(5, profiler.SortByAvgRenderTime)
//
// # Bottleneck Detection
//
// Automatically detect performance issues:
//
//	detector := profiler.NewBottleneckDetector()
//	detector.SetThreshold("render", 16*time.Millisecond) // 60 FPS budget
//
//	// Check operations against thresholds
//	if bottleneck := detector.Check("render", renderTime); bottleneck != nil {
//	    fmt.Printf("Bottleneck: %s\n", bottleneck.Description)
//	    fmt.Printf("Suggestion: %s\n", bottleneck.Suggestion)
//	}
//
//	// Analyze component metrics for patterns
//	metrics := &profiler.PerformanceMetrics{Components: componentMetrics}
//	bottlenecks := detector.Detect(metrics)
//
// # Pattern Analysis
//
// Detect common performance anti-patterns:
//
//	analyzer := profiler.NewPatternAnalyzer()
//
//	// Analyze component for patterns
//	issues := analyzer.Analyze(componentMetrics)
//	for _, issue := range issues {
//	    fmt.Printf("Pattern: %s - %s\n", issue.Description, issue.Suggestion)
//	}
//
//	// Add custom patterns
//	analyzer.AddPattern(profiler.Pattern{
//	    Name:        "custom_pattern",
//	    Detect:      func(m *profiler.ComponentMetrics) bool { return m.RenderCount > 10000 },
//	    Severity:    profiler.SeverityHigh,
//	    Description: "Excessive renders detected",
//	    Suggestion:  "Implement shouldComponentUpdate logic",
//	})
//
// # Recommendations
//
// Get actionable optimization suggestions:
//
//	engine := profiler.NewRecommendationEngine()
//
//	// Generate recommendations from report
//	recommendations := engine.Generate(report)
//	for _, rec := range recommendations {
//	    fmt.Printf("[%s] %s\n", rec.Priority, rec.Title)
//	    fmt.Printf("  Action: %s\n", rec.Action)
//	}
//
// # Report Generation
//
// Generate comprehensive performance reports:
//
//	generator := profiler.NewReportGenerator()
//
//	// Aggregate profiling data
//	data := &profiler.ProfileData{
//	    ComponentTracker: tracker,
//	    StartTime:        startTime,
//	    EndTime:          time.Now(),
//	}
//
//	// Generate report
//	report := generator.Generate(data)
//
//	// Export in various formats
//	exporter := profiler.NewExporter()
//	exporter.ExportHTML(report, "report.html")
//	exporter.ExportJSON(report, "report.json")
//	exporter.ExportCSV(report, "report.csv")
//
// # Flame Graphs
//
// Generate flame graph visualizations:
//
//	fgg := profiler.NewFlameGraphGenerator()
//
//	// Generate from CPU profile data
//	svg := fgg.GenerateSVG(cpuProfileData)
//	os.WriteFile("flamegraph.svg", []byte(svg), 0644)
//
// # Timeline Visualization
//
// Generate timeline views of events:
//
//	tg := profiler.NewTimelineGenerator()
//
//	// Add events
//	events := []*profiler.TimedEvent{
//	    profiler.AddEvent("render", profiler.EventTypeRender, time.Now(), 5*time.Millisecond),
//	    profiler.AddEvent("update", profiler.EventTypeUpdate, time.Now(), 2*time.Millisecond),
//	}
//
//	// Generate HTML timeline
//	html := tg.GenerateHTML(events)
//	os.WriteFile("timeline.html", []byte(html), 0644)
//
// # Benchmarking
//
// Integration with Go's testing benchmarks:
//
//	func BenchmarkComponent(b *testing.B) {
//	    bp := profiler.NewBenchmarkProfiler(b)
//
//	    b.ResetTimer()
//	    for i := 0; i < b.N; i++ {
//	        bp.Measure(func() {
//	            component.Render()
//	        })
//	    }
//
//	    // Report custom metrics
//	    bp.ReportMetrics()
//
//	    // Check for regression
//	    baseline, _ := profiler.LoadBaseline("baseline.json")
//	    if err := bp.AssertNoRegression(baseline, 0.10); err != nil {
//	        b.Fatal(err)
//	    }
//	}
//
// # HTTP Endpoints
//
// Expose pprof endpoints for remote profiling:
//
//	h := profiler.NewHTTPHandler(prof)
//	h.Enable() // Disabled by default for production safety
//
//	mux := http.NewServeMux()
//	h.RegisterHandlers(mux, "/debug/pprof")
//
//	// Available endpoints:
//	// /debug/pprof/           - Index page
//	// /debug/pprof/profile    - CPU profile
//	// /debug/pprof/heap       - Heap profile
//	// /debug/pprof/goroutine  - Goroutine stacks
//	// /debug/pprof/block      - Block profile
//	// /debug/pprof/mutex      - Mutex profile
//	// /debug/pprof/trace      - Execution trace
//
// # Dev Tools Integration
//
// Integrate with BubblyUI dev tools:
//
//	dti := profiler.NewDevToolsIntegration(prof)
//	dti.Enable()
//
//	// Register performance panel
//	dti.RegisterPanel("Performance")
//
//	// Send metrics periodically
//	go func() {
//	    ticker := time.NewTicker(100 * time.Millisecond)
//	    for range ticker.C {
//	        dti.SendMetrics()
//	    }
//	}()
//
//	// Get real-time updates
//	dti.OnMetricsUpdate(func(snapshot *profiler.MetricsSnapshot) {
//	    fmt.Printf("FPS: %.1f, Memory: %d bytes\n",
//	        snapshot.FPS, snapshot.MemoryUsage)
//	})
//
// # Component Instrumentation
//
// Automatically instrument components:
//
//	inst := profiler.NewInstrumentor(prof)
//	inst.Enable()
//
//	// Option 1: Manual instrumentation
//	stop := inst.InstrumentRender(component)
//	output := component.View()
//	stop()
//
//	// Option 2: Wrap component for automatic instrumentation
//	wrapped := inst.InstrumentComponent(component)
//	output := wrapped.View() // Automatically timed
//
// # Thread Safety
//
// All profiler types are thread-safe and can be used concurrently from
// multiple goroutines. This is essential for profiling concurrent TUI
// applications where multiple components may render simultaneously.
//
// # Performance Overhead
//
// The profiler is designed for minimal overhead:
//
//   - Disabled: < 0.1% overhead (fast path checks)
//   - Enabled: < 3% overhead (full profiling)
//   - Sampling: Configurable rate reduces overhead further
//   - Minimal mode: Essential metrics only for production
//
// # Best Practices
//
// 1. Use sampling in production (WithSamplingRate(0.01) for 1%)
// 2. Enable minimal metrics mode for production monitoring
// 3. Set appropriate thresholds for your target frame rate
// 4. Use HTTP endpoints for remote debugging
// 5. Save baselines for regression detection in CI/CD
// 6. Generate reports after profiling sessions for analysis
//
// # Integration with pprof
//
// All CPU and memory profiles are compatible with Go's pprof tools:
//
//	# Analyze CPU profile
//	go tool pprof cpu.prof
//
//	# Analyze heap profile
//	go tool pprof heap.prof
//
//	# Web interface
//	go tool pprof -http=:8080 cpu.prof
//
//	# Compare profiles
//	go tool pprof -base=baseline.prof current.prof
//
// # Package Structure
//
// The profiler package is organized into several components:
//
//   - Core: Profiler, Config, MetricCollector
//   - CPU: CPUProfiler, StackAnalyzer
//   - Memory: MemoryProfiler, MemoryTracker, LeakDetector
//   - Render: RenderProfiler, FPSCalculator, ComponentTracker
//   - Analysis: BottleneckDetector, PatternAnalyzer, ThresholdMonitor
//   - Recommendations: RecommendationEngine
//   - Reporting: ReportGenerator, FlameGraphGenerator, TimelineGenerator, Exporter
//   - Integration: BenchmarkProfiler, HTTPHandler, DevToolsIntegration, Instrumentor
//   - Data: TimingTracker, DataAggregator
package profiler
