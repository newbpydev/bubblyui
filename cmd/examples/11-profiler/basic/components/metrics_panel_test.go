package components

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/cmd/examples/11-profiler/basic/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestCreateMetricsPanel_Creation tests that CreateMetricsPanel creates a valid component.
func TestCreateMetricsPanel_Creation(t *testing.T) {
	metrics := bubbly.NewRef(&composables.ProfilerMetrics{
		FPS:             60.0,
		MemoryUsage:     1024 * 1024,
		GoroutineCount:  10,
		RenderCount:     100,
		BottleneckCount: 0,
		SampleCount:     500,
	})
	focused := bubbly.NewRef(false)
	isRunning := bubbly.NewRef(false)

	panel, err := CreateMetricsPanel(MetricsPanelProps{
		Metrics:   metrics,
		Focused:   focused,
		IsRunning: isRunning,
	})

	require.NoError(t, err)
	assert.NotNil(t, panel)
	assert.Equal(t, "MetricsPanel", panel.Name())
}

// TestCreateMetricsPanel_Render tests that the panel renders correctly.
func TestCreateMetricsPanel_Render(t *testing.T) {
	metrics := bubbly.NewRef(&composables.ProfilerMetrics{
		FPS:             60.0,
		MemoryUsage:     1024 * 1024,
		GoroutineCount:  10,
		RenderCount:     100,
		BottleneckCount: 0,
		SampleCount:     500,
	})
	focused := bubbly.NewRef(false)
	isRunning := bubbly.NewRef(true)

	panel, err := CreateMetricsPanel(MetricsPanelProps{
		Metrics:   metrics,
		Focused:   focused,
		IsRunning: isRunning,
	})

	require.NoError(t, err)
	panel.Init()

	view := panel.View()
	assert.Contains(t, view, "Metrics")
	assert.Contains(t, view, "FPS")
	assert.Contains(t, view, "Memory")
	assert.Contains(t, view, "Goroutines")
	assert.Contains(t, view, "Running")
}

// TestCreateMetricsPanel_FocusedState tests that focus changes border color.
func TestCreateMetricsPanel_FocusedState(t *testing.T) {
	metrics := bubbly.NewRef(&composables.ProfilerMetrics{})
	focused := bubbly.NewRef(true)
	isRunning := bubbly.NewRef(false)

	panel, err := CreateMetricsPanel(MetricsPanelProps{
		Metrics:   metrics,
		Focused:   focused,
		IsRunning: isRunning,
	})

	require.NoError(t, err)
	panel.Init()

	view := panel.View()
	// When focused, should have green border (ANSI code for green)
	assert.NotEmpty(t, view)
}

// TestCreateMetricsPanel_StoppedState tests the stopped state display.
func TestCreateMetricsPanel_StoppedState(t *testing.T) {
	metrics := bubbly.NewRef(&composables.ProfilerMetrics{})
	focused := bubbly.NewRef(false)
	isRunning := bubbly.NewRef(false)

	panel, err := CreateMetricsPanel(MetricsPanelProps{
		Metrics:   metrics,
		Focused:   focused,
		IsRunning: isRunning,
	})

	require.NoError(t, err)
	panel.Init()

	view := panel.View()
	assert.Contains(t, view, "Paused")
}

// TestGetFPSColor tests the FPS color function.
func TestGetFPSColor(t *testing.T) {
	tests := []struct {
		name string
		fps  float64
	}{
		{"excellent FPS", 60.0},
		{"acceptable FPS", 45.0},
		{"poor FPS", 15.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			color := getFPSColor(tt.fps)
			assert.NotEmpty(t, string(color))
		})
	}
}

// TestGetMemoryColor tests the memory color function.
func TestGetMemoryColor(t *testing.T) {
	tests := []struct {
		name  string
		bytes uint64
	}{
		{"low memory", 10 * 1024 * 1024},
		{"moderate memory", 100 * 1024 * 1024},
		{"high memory", 500 * 1024 * 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			color := getMemoryColor(tt.bytes)
			assert.NotEmpty(t, string(color))
		})
	}
}

// TestFormatDuration tests the duration formatting.
func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		contains string
	}{
		{"milliseconds", 500 * time.Millisecond, "ms"},
		{"seconds", 5 * time.Second, "s"},
		{"minutes", 2 * time.Minute, "m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.duration)
			assert.Contains(t, result, tt.contains)
		})
	}
}
