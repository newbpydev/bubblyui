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

		report := prof.GenerateReport()
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

		// Get memory stats
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		// Get current metrics from profiler
		report := prof.GenerateReport()

		// Calculate FPS and frame time from report
		var fps float64
		var renderCount int
		var bottleneckCount int
		var sampleCount int

		if report != nil && report.Summary != nil {
			fps = report.Summary.AverageFPS
			sampleCount = int(report.Summary.TotalOperations)
		}

		if report != nil {
			bottleneckCount = len(report.Bottlenecks)
			// Count total renders from component metrics
			for _, cm := range report.Components {
				renderCount += int(cm.RenderCount)
			}
		}

		// Update metrics ref
		metrics.Set(&ProfilerMetrics{
			FPS:             fps,
			FrameTime:       time.Duration(0), // Will be calculated from FPS if needed
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
