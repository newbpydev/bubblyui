// Package composables provides reusable reactive logic for the profiler example.
package composables

import (
	"runtime"
	"sync"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/profiler"
)

// ProfilerMetrics holds the current profiler metrics for display.
type ProfilerMetrics struct {
	FPS             float64
	FrameTime       time.Duration
	MemoryUsage     uint64
	GoroutineCount  int
	RenderCount     int
	BottleneckCount int
	SampleCount     int
}

// ProfilerComposable encapsulates profiler logic with reactive state.
// This follows the Vue-like composable pattern for reusable logic.
type ProfilerComposable struct {
	// Profiler is the underlying profiler instance
	Profiler *profiler.Profiler

	// IsRunning indicates if profiling is active
	IsRunning *bubbly.Ref[bool]

	// Metrics holds the current profiler metrics
	Metrics *bubbly.Ref[*ProfilerMetrics]

	// StartTime is when profiling started
	StartTime *bubbly.Ref[time.Time]

	// Duration is computed from StartTime
	Duration *bubbly.Computed[interface{}]

	// LastExport holds the last export filename (empty if none)
	LastExport *bubbly.Ref[string]

	// Start begins profiling
	Start func()

	// Stop ends profiling
	Stop func()

	// Toggle switches profiling state
	Toggle func()

	// Reset clears all metrics
	Reset func()

	// ExportReport saves the report to a file
	ExportReport func(filename string) error

	// RefreshMetrics updates the metrics from the profiler
	RefreshMetrics func()

	// mu protects concurrent access
	mu sync.Mutex
}

// UseProfiler creates a reusable profiler composable with reactive state.
// This demonstrates the composable pattern - reusable logic that can be shared
// across components, similar to Vue's Composition API.
//
// Example:
//
//	profilerComp := composables.UseProfiler(ctx)
//	profilerComp.Start()
//	// ... later
//	profilerComp.RefreshMetrics()
//	metrics := profilerComp.Metrics.GetTyped()
func UseProfiler(ctx *bubbly.Context) *ProfilerComposable {
	// Create the underlying profiler with sensible defaults
	prof := profiler.New(
		profiler.WithEnabled(true),
		profiler.WithSamplingRate(1.0), // 100% sampling for demo
		profiler.WithMaxSamples(1000),
	)

	// Create reactive state using typed refs
	isRunning := bubbly.NewRef(false)
	metrics := bubbly.NewRef(&ProfilerMetrics{})
	startTime := bubbly.NewRef(time.Time{})
	lastExport := bubbly.NewRef("")

	// Internal tracking state for FPS calculation
	// These are not exposed as refs since they're internal implementation details
	var (
		sampleCount      int
		renderCount      int
		lastSampleTime   time.Time
		frameTimeSamples []time.Duration
	)

	// Create computed duration
	duration := ctx.Computed(func() interface{} {
		start := startTime.GetTyped()
		if start.IsZero() {
			return time.Duration(0)
		}
		return time.Since(start)
	})

	// Create the composable instance
	comp := &ProfilerComposable{
		Profiler:   prof,
		IsRunning:  isRunning,
		Metrics:    metrics,
		StartTime:  startTime,
		Duration:   duration,
		LastExport: lastExport,
	}

	// Define Start method
	comp.Start = func() {
		comp.mu.Lock()
		defer comp.mu.Unlock()

		if isRunning.GetTyped() {
			return // Already running
		}

		prof.Start()
		startTime.Set(time.Now())
		isRunning.Set(true)
		lastExport.Set("")
	}

	// Define Stop method
	comp.Stop = func() {
		comp.mu.Lock()
		defer comp.mu.Unlock()

		if !isRunning.GetTyped() {
			return // Already stopped
		}

		prof.Stop()
		isRunning.Set(false)
	}

	// Define Toggle method
	comp.Toggle = func() {
		if isRunning.GetTyped() {
			comp.Stop()
		} else {
			comp.Start()
		}
	}

	// Define Reset method
	comp.Reset = func() {
		comp.mu.Lock()
		defer comp.mu.Unlock()

		// Reset metrics
		metrics.Set(&ProfilerMetrics{})
		startTime.Set(time.Time{})
		lastExport.Set("")

		// Reset internal tracking state
		sampleCount = 0
		renderCount = 0
		lastSampleTime = time.Time{}
		frameTimeSamples = nil

		// If running, restart the profiler
		if isRunning.GetTyped() {
			prof.Stop()
			prof.Start()
			startTime.Set(time.Now())
		}
	}

	// Define ExportReport method
	comp.ExportReport = func(filename string) error {
		comp.mu.Lock()
		defer comp.mu.Unlock()

		// Get current metrics from our reactive state
		currentMetrics := metrics.GetTyped()

		// Calculate average FPS for report
		var avgFPS float64
		if len(frameTimeSamples) > 0 {
			var totalTime time.Duration
			for _, ft := range frameTimeSamples {
				totalTime += ft
			}
			avgFrameTime := totalTime / time.Duration(len(frameTimeSamples))
			if avgFrameTime > 0 {
				avgFPS = float64(time.Second) / float64(avgFrameTime)
			}
		}

		// Count bottlenecks for report
		bottleneckCount := 0
		for _, ft := range frameTimeSamples {
			if ft > 100*time.Millisecond {
				bottleneckCount++
			}
		}

		// Build bottleneck info if any
		var bottlenecks []*profiler.BottleneckInfo
		if bottleneckCount > 0 {
			bottlenecks = append(bottlenecks, &profiler.BottleneckInfo{
				Type:        profiler.BottleneckTypeSlow,
				Location:    "Frame rendering",
				Severity:    profiler.SeverityMedium,
				Impact:      float64(bottleneckCount) / float64(len(frameTimeSamples)),
				Description: "Some frames took longer than 100ms to render",
				Suggestion:  "Optimize component rendering or reduce update frequency",
			})
		}

		// Build a report with actual data from our metrics
		// The profiler's GenerateReport() returns empty data, so we populate it ourselves
		report := &profiler.Report{
			Summary: &profiler.Summary{
				Duration:        time.Since(startTime.GetTyped()),
				TotalOperations: int64(sampleCount),
				AverageFPS:      avgFPS,
				MemoryUsage:     currentMetrics.MemoryUsage,
				GoroutineCount:  currentMetrics.GoroutineCount,
			},
			Components:      make([]*profiler.ComponentMetrics, 0),
			Bottlenecks:     bottlenecks,
			CPUProfile:      &profiler.CPUProfileData{},
			MemProfile:      &profiler.MemProfileData{HeapAlloc: currentMetrics.MemoryUsage},
			Recommendations: make([]*profiler.Recommendation, 0),
			Timestamp:       time.Now(),
		}

		exporter := profiler.NewExporter()
		err := exporter.ExportHTML(report, filename)
		if err != nil {
			return err
		}

		lastExport.Set(filename)
		return nil
	}

	// Define RefreshMetrics method
	comp.RefreshMetrics = func() {
		comp.mu.Lock()
		defer comp.mu.Unlock()

		// Only track if profiler is running
		if !isRunning.GetTyped() {
			return
		}

		now := time.Now()

		// Calculate frame time since last sample
		var frameTime time.Duration
		if !lastSampleTime.IsZero() {
			frameTime = now.Sub(lastSampleTime)
			// Keep last 60 samples for FPS calculation (rolling window)
			frameTimeSamples = append(frameTimeSamples, frameTime)
			if len(frameTimeSamples) > 60 {
				frameTimeSamples = frameTimeSamples[1:]
			}
		}
		lastSampleTime = now

		// Increment counters
		sampleCount++
		renderCount++ // Each refresh is a "render" in this demo

		// Calculate FPS from frame time samples
		var fps float64
		if len(frameTimeSamples) > 0 {
			var totalTime time.Duration
			for _, ft := range frameTimeSamples {
				totalTime += ft
			}
			avgFrameTime := totalTime / time.Duration(len(frameTimeSamples))
			if avgFrameTime > 0 {
				fps = float64(time.Second) / float64(avgFrameTime)
			}
		}

		// Get memory stats
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		// Detect bottlenecks (simple heuristic: frame time > 100ms is a bottleneck)
		bottleneckCount := 0
		for _, ft := range frameTimeSamples {
			if ft > 100*time.Millisecond {
				bottleneckCount++
			}
		}

		// Update metrics ref
		metrics.Set(&ProfilerMetrics{
			FPS:             fps,
			FrameTime:       frameTime,
			MemoryUsage:     memStats.Alloc,
			GoroutineCount:  runtime.NumGoroutine(),
			RenderCount:     renderCount,
			BottleneckCount: bottleneckCount,
			SampleCount:     sampleCount,
		})
	}

	return comp
}

// FormatBytes formats bytes into a human-readable string.
func FormatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return formatFloat(float64(bytes)/float64(GB)) + " GB"
	case bytes >= MB:
		return formatFloat(float64(bytes)/float64(MB)) + " MB"
	case bytes >= KB:
		return formatFloat(float64(bytes)/float64(KB)) + " KB"
	default:
		return formatInt(int(bytes)) + " B"
	}
}

// FormatDuration formats a duration into a human-readable string.
func FormatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return formatFloat(float64(d.Microseconds())) + " µs"
	}
	if d < time.Second {
		return formatFloat(float64(d.Milliseconds())) + " ms"
	}
	return d.Round(time.Millisecond).String()
}

func formatFloat(f float64) string {
	if f == float64(int(f)) {
		return formatInt(int(f))
	}
	// Format with 1 decimal place
	return string(rune('0'+int(f)/10)) + string(rune('0'+int(f)%10)) + "." + string(rune('0'+int(f*10)%10))
}

func formatInt(i int) string {
	if i == 0 {
		return "0"
	}
	if i < 0 {
		return "-" + formatInt(-i)
	}

	result := ""
	for i > 0 {
		result = string(rune('0'+i%10)) + result
		i /= 10
	}
	return result
}
