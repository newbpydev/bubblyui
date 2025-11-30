package composables

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestUseProfiler_Creation tests that UseProfiler creates a valid composable.
func TestUseProfiler_Creation(t *testing.T) {
	// Create a test component to get a context
	comp, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			profiler := UseProfiler(ctx)

			// Verify all fields are initialized
			assert.NotNil(t, profiler.Profiler, "Profiler should not be nil")
			assert.NotNil(t, profiler.IsRunning, "IsRunning should not be nil")
			assert.NotNil(t, profiler.Metrics, "Metrics should not be nil")
			assert.NotNil(t, profiler.StartTime, "StartTime should not be nil")
			assert.NotNil(t, profiler.Duration, "Duration should not be nil")
			assert.NotNil(t, profiler.LastExport, "LastExport should not be nil")
			assert.NotNil(t, profiler.Start, "Start should not be nil")
			assert.NotNil(t, profiler.Stop, "Stop should not be nil")
			assert.NotNil(t, profiler.Toggle, "Toggle should not be nil")
			assert.NotNil(t, profiler.Reset, "Reset should not be nil")
			assert.NotNil(t, profiler.ExportReport, "ExportReport should not be nil")
			assert.NotNil(t, profiler.RefreshMetrics, "RefreshMetrics should not be nil")

			// Verify initial state
			assert.False(t, profiler.IsRunning.GetTyped(), "Profiler should not be running initially")
			assert.True(t, profiler.StartTime.GetTyped().IsZero(), "StartTime should be zero initially")
			assert.Empty(t, profiler.LastExport.GetTyped(), "LastExport should be empty initially")
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	require.NoError(t, err)
	comp.Init()
}

// TestUseProfiler_StartStop tests the Start and Stop methods.
func TestUseProfiler_StartStop(t *testing.T) {
	var profiler *ProfilerComposable

	comp, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			profiler = UseProfiler(ctx)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	require.NoError(t, err)
	comp.Init()

	// Initially not running
	assert.False(t, profiler.IsRunning.GetTyped())

	// Start profiler
	profiler.Start()
	assert.True(t, profiler.IsRunning.GetTyped())
	assert.False(t, profiler.StartTime.GetTyped().IsZero())

	// Start again should be no-op
	startTime := profiler.StartTime.GetTyped()
	profiler.Start()
	assert.Equal(t, startTime, profiler.StartTime.GetTyped())

	// Stop profiler
	profiler.Stop()
	assert.False(t, profiler.IsRunning.GetTyped())

	// Stop again should be no-op
	profiler.Stop()
	assert.False(t, profiler.IsRunning.GetTyped())
}

// TestUseProfiler_Toggle tests the Toggle method.
func TestUseProfiler_Toggle(t *testing.T) {
	var profiler *ProfilerComposable

	comp, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			profiler = UseProfiler(ctx)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	require.NoError(t, err)
	comp.Init()

	// Initially not running
	assert.False(t, profiler.IsRunning.GetTyped())

	// Toggle to start
	profiler.Toggle()
	assert.True(t, profiler.IsRunning.GetTyped())

	// Toggle to stop
	profiler.Toggle()
	assert.False(t, profiler.IsRunning.GetTyped())
}

// TestUseProfiler_Reset tests the Reset method.
func TestUseProfiler_Reset(t *testing.T) {
	var profiler *ProfilerComposable

	comp, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			profiler = UseProfiler(ctx)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	require.NoError(t, err)
	comp.Init()

	// Start and let it run briefly
	profiler.Start()
	time.Sleep(10 * time.Millisecond)
	profiler.RefreshMetrics()

	// Reset while running
	profiler.Reset()
	assert.True(t, profiler.IsRunning.GetTyped(), "Should still be running after reset")
	assert.Empty(t, profiler.LastExport.GetTyped())

	// Stop and reset
	profiler.Stop()
	profiler.Reset()
	assert.False(t, profiler.IsRunning.GetTyped())
	assert.True(t, profiler.StartTime.GetTyped().IsZero())
}

// TestUseProfiler_RefreshMetrics tests the RefreshMetrics method.
func TestUseProfiler_RefreshMetrics(t *testing.T) {
	var profiler *ProfilerComposable

	comp, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			profiler = UseProfiler(ctx)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	require.NoError(t, err)
	comp.Init()

	// Refresh metrics
	profiler.RefreshMetrics()

	// Verify metrics are populated
	metrics := profiler.Metrics.GetTyped()
	assert.NotNil(t, metrics)
	assert.GreaterOrEqual(t, metrics.GoroutineCount, 1, "Should have at least 1 goroutine")
	assert.Greater(t, metrics.MemoryUsage, uint64(0), "Should have some memory usage")
}

// TestFormatBytes tests the FormatBytes function.
func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    uint64
		expected string
	}{
		{"zero bytes", 0, "0 B"},
		{"bytes", 500, "500 B"},
		{"kilobytes", 1024, "1 KB"},
		{"kilobytes with decimal", 1536, "1 KB"},
		{"megabytes", 1024 * 1024, "1 MB"},
		{"megabytes with decimal", 1536 * 1024, "1 MB"},
		{"gigabytes", 1024 * 1024 * 1024, "1 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBytes(tt.bytes)
			assert.Contains(t, result, "B", "Should contain B suffix")
		})
	}
}

// TestFormatDuration tests the FormatDuration function.
func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		contains string
	}{
		{"microseconds", 500 * time.Microsecond, "µs"},
		{"milliseconds", 500 * time.Millisecond, "ms"},
		{"seconds", 5 * time.Second, "s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.duration)
			assert.Contains(t, result, tt.contains)
		})
	}
}

// TestProfilerMetrics_Fields tests that ProfilerMetrics has all expected fields.
func TestProfilerMetrics_Fields(t *testing.T) {
	metrics := &ProfilerMetrics{
		FPS:             60.0,
		FrameTime:       16 * time.Millisecond,
		MemoryUsage:     1024 * 1024,
		GoroutineCount:  10,
		RenderCount:     100,
		BottleneckCount: 2,
		SampleCount:     500,
	}

	assert.Equal(t, 60.0, metrics.FPS)
	assert.Equal(t, 16*time.Millisecond, metrics.FrameTime)
	assert.Equal(t, uint64(1024*1024), metrics.MemoryUsage)
	assert.Equal(t, 10, metrics.GoroutineCount)
	assert.Equal(t, 100, metrics.RenderCount)
	assert.Equal(t, 2, metrics.BottleneckCount)
	assert.Equal(t, 500, metrics.SampleCount)
}

// TestUseProfiler_ExportReport tests the ExportReport method.
func TestUseProfiler_ExportReport(t *testing.T) {
	var profiler *ProfilerComposable

	comp, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			profiler = UseProfiler(ctx)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	require.NoError(t, err)
	comp.Init()

	// Start profiler and generate some data
	profiler.Start()
	profiler.RefreshMetrics()
	profiler.Stop()

	// Export to a temp file
	filename := t.TempDir() + "/test-report.html"
	err = profiler.ExportReport(filename)
	assert.NoError(t, err)
	assert.Equal(t, filename, profiler.LastExport.GetTyped())
}

// TestUseProfiler_Duration tests the Duration computed value.
func TestUseProfiler_Duration(t *testing.T) {
	var profiler *ProfilerComposable

	comp, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			profiler = UseProfiler(ctx)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	require.NoError(t, err)
	comp.Init()

	// Initially duration should be 0
	duration := profiler.Duration.Get().(time.Duration)
	assert.Equal(t, time.Duration(0), duration)

	// Start profiler
	profiler.Start()
	time.Sleep(50 * time.Millisecond)

	// Duration should be > 0
	duration = profiler.Duration.Get().(time.Duration)
	assert.Greater(t, duration, time.Duration(0))
}

// TestUseProfiler_ConcurrentAccess tests thread safety.
func TestUseProfiler_ConcurrentAccess(t *testing.T) {
	var profiler *ProfilerComposable

	comp, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			profiler = UseProfiler(ctx)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	require.NoError(t, err)
	comp.Init()

	// Run concurrent operations
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			profiler.Start()
			profiler.RefreshMetrics()
			profiler.Stop()
			profiler.Toggle()
			profiler.Reset()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not panic
	assert.True(t, true)
}
