package composables

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestCPUProfilerState_String tests the CPUProfilerState String method.
func TestCPUProfilerState_String(t *testing.T) {
	tests := []struct {
		name     string
		state    CPUProfilerState
		expected string
	}{
		{"idle", StateIdle, "Idle"},
		{"profiling", StateProfiling, "Profiling"},
		{"complete", StateComplete, "Complete"},
		{"unknown", CPUProfilerState("unknown"), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.state.String())
		})
	}
}

// TestUseCPUProfiler_Creation tests that UseCPUProfiler creates a valid composable.
func TestUseCPUProfiler_Creation(t *testing.T) {
	var cpuProfiler *CPUProfilerComposable

	comp, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			cpuProfiler = UseCPUProfiler(ctx)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	require.NoError(t, err)
	comp.Init()

	// Verify composable is created
	assert.NotNil(t, cpuProfiler)
	assert.NotNil(t, cpuProfiler.Profiler)
	assert.NotNil(t, cpuProfiler.State)
	assert.NotNil(t, cpuProfiler.Filename)
	assert.NotNil(t, cpuProfiler.StartTime)
	assert.NotNil(t, cpuProfiler.FileSize)
	assert.NotNil(t, cpuProfiler.HotFunctions)
	assert.NotNil(t, cpuProfiler.LastError)

	// Verify initial state
	assert.Equal(t, StateIdle, cpuProfiler.State.GetTyped())
	assert.Equal(t, "", cpuProfiler.Filename.GetTyped())
	assert.True(t, cpuProfiler.StartTime.GetTyped().IsZero())
	assert.Equal(t, int64(0), cpuProfiler.FileSize.GetTyped())
	assert.Empty(t, cpuProfiler.HotFunctions.GetTyped())
	assert.Equal(t, "", cpuProfiler.LastError.GetTyped())
}

// TestUseCPUProfiler_StartStop tests the Start and Stop methods.
func TestUseCPUProfiler_StartStop(t *testing.T) {
	var cpuProfiler *CPUProfilerComposable

	comp, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			cpuProfiler = UseCPUProfiler(ctx)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	require.NoError(t, err)
	comp.Init()

	// Create a temp file for the profile
	tmpFile := t.TempDir() + "/test_cpu.prof"

	// Start profiling
	err = cpuProfiler.Start(tmpFile)
	require.NoError(t, err)

	// Verify state changed
	assert.Equal(t, StateProfiling, cpuProfiler.State.GetTyped())
	assert.Equal(t, tmpFile, cpuProfiler.Filename.GetTyped())
	assert.False(t, cpuProfiler.StartTime.GetTyped().IsZero())

	// Wait a bit to generate some profile data
	time.Sleep(50 * time.Millisecond)

	// Stop profiling
	err = cpuProfiler.Stop()
	require.NoError(t, err)

	// Verify state changed
	assert.Equal(t, StateComplete, cpuProfiler.State.GetTyped())
	assert.Greater(t, cpuProfiler.FileSize.GetTyped(), int64(0))

	// Verify file exists
	_, err = os.Stat(tmpFile)
	assert.NoError(t, err)

	// Cleanup
	os.Remove(tmpFile)
}

// TestUseCPUProfiler_StartTwice tests that starting twice returns an error.
func TestUseCPUProfiler_StartTwice(t *testing.T) {
	var cpuProfiler *CPUProfilerComposable

	comp, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			cpuProfiler = UseCPUProfiler(ctx)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	require.NoError(t, err)
	comp.Init()

	tmpFile := t.TempDir() + "/test_cpu.prof"

	// Start profiling
	err = cpuProfiler.Start(tmpFile)
	require.NoError(t, err)

	// Try to start again - should fail
	err = cpuProfiler.Start(tmpFile + "2")
	assert.Error(t, err)
	assert.NotEmpty(t, cpuProfiler.LastError.GetTyped())

	// Cleanup
	cpuProfiler.Stop()
	os.Remove(tmpFile)
}

// TestUseCPUProfiler_StopWithoutStart tests that stopping without starting returns an error.
func TestUseCPUProfiler_StopWithoutStart(t *testing.T) {
	var cpuProfiler *CPUProfilerComposable

	comp, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			cpuProfiler = UseCPUProfiler(ctx)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	require.NoError(t, err)
	comp.Init()

	// Try to stop without starting - should fail
	err = cpuProfiler.Stop()
	assert.Error(t, err)
	assert.NotEmpty(t, cpuProfiler.LastError.GetTyped())
}

// TestUseCPUProfiler_Analyze tests the Analyze method.
func TestUseCPUProfiler_Analyze(t *testing.T) {
	var cpuProfiler *CPUProfilerComposable

	comp, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			cpuProfiler = UseCPUProfiler(ctx)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	require.NoError(t, err)
	comp.Init()

	tmpFile := t.TempDir() + "/test_cpu.prof"

	// Start and stop profiling
	err = cpuProfiler.Start(tmpFile)
	require.NoError(t, err)
	time.Sleep(50 * time.Millisecond)
	err = cpuProfiler.Stop()
	require.NoError(t, err)

	// Analyze
	err = cpuProfiler.Analyze()
	require.NoError(t, err)

	// Verify hot functions are populated
	hotFuncs := cpuProfiler.HotFunctions.GetTyped()
	assert.NotEmpty(t, hotFuncs)
	assert.Greater(t, len(hotFuncs), 0)

	// Verify first hot function has expected fields
	assert.NotEmpty(t, hotFuncs[0].Name)
	assert.Greater(t, hotFuncs[0].Samples, int64(0))
	assert.Greater(t, hotFuncs[0].Percent, float64(0))

	// Cleanup
	os.Remove(tmpFile)
}

// TestUseCPUProfiler_AnalyzeWithoutComplete tests that analyzing without completing returns an error.
func TestUseCPUProfiler_AnalyzeWithoutComplete(t *testing.T) {
	var cpuProfiler *CPUProfilerComposable

	comp, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			cpuProfiler = UseCPUProfiler(ctx)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	require.NoError(t, err)
	comp.Init()

	// Try to analyze in idle state - should fail
	err = cpuProfiler.Analyze()
	assert.Error(t, err)
	assert.NotEmpty(t, cpuProfiler.LastError.GetTyped())
}

// TestUseCPUProfiler_Reset tests the Reset method.
func TestUseCPUProfiler_Reset(t *testing.T) {
	var cpuProfiler *CPUProfilerComposable

	comp, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			cpuProfiler = UseCPUProfiler(ctx)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	require.NoError(t, err)
	comp.Init()

	tmpFile := t.TempDir() + "/test_cpu.prof"

	// Start, stop, and analyze
	err = cpuProfiler.Start(tmpFile)
	require.NoError(t, err)
	time.Sleep(50 * time.Millisecond)
	err = cpuProfiler.Stop()
	require.NoError(t, err)
	err = cpuProfiler.Analyze()
	require.NoError(t, err)

	// Verify we have data
	assert.Equal(t, StateComplete, cpuProfiler.State.GetTyped())
	assert.NotEmpty(t, cpuProfiler.HotFunctions.GetTyped())

	// Reset
	cpuProfiler.Reset()

	// Verify all state is cleared
	assert.Equal(t, StateIdle, cpuProfiler.State.GetTyped())
	assert.Equal(t, "", cpuProfiler.Filename.GetTyped())
	assert.True(t, cpuProfiler.StartTime.GetTyped().IsZero())
	assert.Equal(t, int64(0), cpuProfiler.FileSize.GetTyped())
	assert.Empty(t, cpuProfiler.HotFunctions.GetTyped())
	assert.Equal(t, "", cpuProfiler.LastError.GetTyped())

	// Cleanup
	os.Remove(tmpFile)
}

// TestUseCPUProfiler_ResetWhileProfiling tests resetting while profiling is active.
func TestUseCPUProfiler_ResetWhileProfiling(t *testing.T) {
	var cpuProfiler *CPUProfilerComposable

	comp, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			cpuProfiler = UseCPUProfiler(ctx)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	require.NoError(t, err)
	comp.Init()

	tmpFile := t.TempDir() + "/test_cpu.prof"

	// Start profiling
	err = cpuProfiler.Start(tmpFile)
	require.NoError(t, err)
	assert.Equal(t, StateProfiling, cpuProfiler.State.GetTyped())

	// Reset while profiling
	cpuProfiler.Reset()

	// Verify state is cleared
	assert.Equal(t, StateIdle, cpuProfiler.State.GetTyped())

	// Cleanup
	os.Remove(tmpFile)
}

// TestFormatBytes tests the FormatBytes function.
func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"zero bytes", 0, "0 B"},
		{"bytes", 500, "500 B"},
		{"kilobytes", 1024, "1.0 KB"},
		{"kilobytes with decimal", 1536, "1.5 KB"},
		{"megabytes", 1024 * 1024, "1.0 MB"},
		{"megabytes with decimal", int64(1.5 * 1024 * 1024), "1.5 MB"},
		{"gigabytes", 1024 * 1024 * 1024, "1.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBytes(tt.bytes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFormatDuration tests the FormatDuration function.
func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"milliseconds", 500 * time.Millisecond, "500ms"},
		{"seconds", 30 * time.Second, "30s"},
		{"minutes", 2*time.Minute + 30*time.Second, "2m 30s"},
		{"hours", 1*time.Hour + 30*time.Minute + 45*time.Second, "1h 30m 45s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}
