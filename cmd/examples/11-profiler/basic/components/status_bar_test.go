package components

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestFocusPane_String tests the FocusPane String method.
func TestFocusPane_String(t *testing.T) {
	tests := []struct {
		name     string
		pane     FocusPane
		expected string
	}{
		{"metrics", FocusMetrics, "Metrics"},
		{"controls", FocusControls, "Controls"},
		{"unknown", FocusPane(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.pane.String())
		})
	}
}

// TestCreateStatusBar_Creation tests that CreateStatusBar creates a valid component.
func TestCreateStatusBar_Creation(t *testing.T) {
	isRunning := bubbly.NewRef(false)
	focusedPane := bubbly.NewRef(FocusControls)
	lastExport := bubbly.NewRef("")

	// Create a mock context for computed
	var duration *bubbly.Computed[interface{}]
	comp, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			duration = ctx.Computed(func() interface{} {
				return time.Duration(0)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	require.NoError(t, err)
	comp.Init()

	statusBar, err := CreateStatusBar(StatusBarProps{
		IsRunning:   isRunning,
		Duration:    duration,
		FocusedPane: focusedPane,
		LastExport:  lastExport,
	})

	require.NoError(t, err)
	assert.NotNil(t, statusBar)
	assert.Equal(t, "StatusBar", statusBar.Name())
}

// TestCreateStatusBar_Render tests that the status bar renders correctly.
func TestCreateStatusBar_Render(t *testing.T) {
	isRunning := bubbly.NewRef(false)
	focusedPane := bubbly.NewRef(FocusControls)
	lastExport := bubbly.NewRef("")

	var duration *bubbly.Computed[interface{}]
	comp, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			duration = ctx.Computed(func() interface{} {
				return time.Duration(0)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	require.NoError(t, err)
	comp.Init()

	statusBar, err := CreateStatusBar(StatusBarProps{
		IsRunning:   isRunning,
		Duration:    duration,
		FocusedPane: focusedPane,
		LastExport:  lastExport,
	})

	require.NoError(t, err)
	statusBar.Init()

	view := statusBar.View()
	assert.Contains(t, view, "STOPPED")
	assert.Contains(t, view, "Duration")
	assert.Contains(t, view, "Focus")
	assert.Contains(t, view, "Controls")
}

// TestCreateStatusBar_RunningState tests the running state display.
func TestCreateStatusBar_RunningState(t *testing.T) {
	isRunning := bubbly.NewRef(true)
	focusedPane := bubbly.NewRef(FocusMetrics)
	lastExport := bubbly.NewRef("")

	var duration *bubbly.Computed[interface{}]
	comp, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			duration = ctx.Computed(func() interface{} {
				return 5 * time.Second
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	require.NoError(t, err)
	comp.Init()

	statusBar, err := CreateStatusBar(StatusBarProps{
		IsRunning:   isRunning,
		Duration:    duration,
		FocusedPane: focusedPane,
		LastExport:  lastExport,
	})

	require.NoError(t, err)
	statusBar.Init()

	view := statusBar.View()
	assert.Contains(t, view, "RUNNING")
	assert.Contains(t, view, "Metrics")
}

// TestCreateStatusBar_ExportStatus tests the export status display.
func TestCreateStatusBar_ExportStatus(t *testing.T) {
	tests := []struct {
		name       string
		lastExport string
		contains   string
	}{
		{"no export", "", "No export yet"},
		{"with export", "report.html", "Exported"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isRunning := bubbly.NewRef(false)
			focusedPane := bubbly.NewRef(FocusControls)
			lastExport := bubbly.NewRef(tt.lastExport)

			var duration *bubbly.Computed[interface{}]
			comp, err := bubbly.NewComponent("TestComponent").
				Setup(func(ctx *bubbly.Context) {
					duration = ctx.Computed(func() interface{} {
						return time.Duration(0)
					})
				}).
				Template(func(ctx bubbly.RenderContext) string {
					return "test"
				}).
				Build()
			require.NoError(t, err)
			comp.Init()

			statusBar, err := CreateStatusBar(StatusBarProps{
				IsRunning:   isRunning,
				Duration:    duration,
				FocusedPane: focusedPane,
				LastExport:  lastExport,
			})

			require.NoError(t, err)
			statusBar.Init()

			view := statusBar.View()
			assert.Contains(t, view, tt.contains)
		})
	}
}

// TestFormatDuration_StatusBar tests the duration formatting in status bar.
func TestFormatDuration_StatusBar(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"zero", 0, "0s"},
		{"seconds", 30 * time.Second, "30s"},
		{"minutes", 2*time.Minute + 30*time.Second, "2m 30s"},
		{"hours", 1*time.Hour + 30*time.Minute + 45*time.Second, "1h 30m 45s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}
