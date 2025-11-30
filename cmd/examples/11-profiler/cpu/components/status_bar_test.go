package components

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/cmd/examples/11-profiler/cpu/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestCreateStatusBar_Creation tests that CreateStatusBar creates a valid component.
func TestCreateStatusBar_Creation(t *testing.T) {
	state := bubbly.NewRef(composables.StateIdle)
	startTime := bubbly.NewRef(time.Time{})
	filename := bubbly.NewRef("")
	focusedPane := bubbly.NewRef(FocusControls)
	hasResults := bubbly.NewRef(false)
	lastError := bubbly.NewRef("")

	bar, err := CreateStatusBar(StatusBarProps{
		State:       state,
		StartTime:   startTime,
		Filename:    filename,
		FocusedPane: focusedPane,
		HasResults:  hasResults,
		LastError:   lastError,
	})

	require.NoError(t, err)
	assert.NotNil(t, bar)
	assert.Equal(t, "StatusBar", bar.Name())
}

// TestCreateStatusBar_IdleState tests rendering in idle state.
func TestCreateStatusBar_IdleState(t *testing.T) {
	state := bubbly.NewRef(composables.StateIdle)
	startTime := bubbly.NewRef(time.Time{})
	filename := bubbly.NewRef("")
	focusedPane := bubbly.NewRef(FocusControls)
	hasResults := bubbly.NewRef(false)
	lastError := bubbly.NewRef("")

	bar, err := CreateStatusBar(StatusBarProps{
		State:       state,
		StartTime:   startTime,
		Filename:    filename,
		FocusedPane: focusedPane,
		HasResults:  hasResults,
		LastError:   lastError,
	})

	require.NoError(t, err)
	bar.Init()

	view := bar.View()
	assert.Contains(t, view, "IDLE")
	assert.Contains(t, view, "Focus")
	assert.Contains(t, view, "Controls")
	assert.Contains(t, view, "Start profiling")
}

// TestCreateStatusBar_ProfilingState tests rendering in profiling state.
func TestCreateStatusBar_ProfilingState(t *testing.T) {
	state := bubbly.NewRef(composables.StateProfiling)
	startTime := bubbly.NewRef(time.Now().Add(-5 * time.Second))
	filename := bubbly.NewRef("cpu.prof")
	focusedPane := bubbly.NewRef(FocusControls)
	hasResults := bubbly.NewRef(false)
	lastError := bubbly.NewRef("")

	bar, err := CreateStatusBar(StatusBarProps{
		State:       state,
		StartTime:   startTime,
		Filename:    filename,
		FocusedPane: focusedPane,
		HasResults:  hasResults,
		LastError:   lastError,
	})

	require.NoError(t, err)
	bar.Init()

	view := bar.View()
	assert.Contains(t, view, "PROFILING")
	assert.Contains(t, view, "Duration")
	assert.Contains(t, view, "Stop profiling")
}

// TestCreateStatusBar_CompleteState tests rendering in complete state.
func TestCreateStatusBar_CompleteState(t *testing.T) {
	state := bubbly.NewRef(composables.StateComplete)
	startTime := bubbly.NewRef(time.Now().Add(-10 * time.Second))
	filename := bubbly.NewRef("cpu.prof")
	focusedPane := bubbly.NewRef(FocusControls)
	hasResults := bubbly.NewRef(false)
	lastError := bubbly.NewRef("")

	bar, err := CreateStatusBar(StatusBarProps{
		State:       state,
		StartTime:   startTime,
		Filename:    filename,
		FocusedPane: focusedPane,
		HasResults:  hasResults,
		LastError:   lastError,
	})

	require.NoError(t, err)
	bar.Init()

	view := bar.View()
	assert.Contains(t, view, "COMPLETE")
	assert.Contains(t, view, "Analyze")
}

// TestCreateStatusBar_AnalyzedState tests rendering when results are available.
func TestCreateStatusBar_AnalyzedState(t *testing.T) {
	state := bubbly.NewRef(composables.StateComplete)
	startTime := bubbly.NewRef(time.Now().Add(-10 * time.Second))
	filename := bubbly.NewRef("cpu.prof")
	focusedPane := bubbly.NewRef(FocusResults)
	hasResults := bubbly.NewRef(true)
	lastError := bubbly.NewRef("")

	bar, err := CreateStatusBar(StatusBarProps{
		State:       state,
		StartTime:   startTime,
		Filename:    filename,
		FocusedPane: focusedPane,
		HasResults:  hasResults,
		LastError:   lastError,
	})

	require.NoError(t, err)
	bar.Init()

	view := bar.View()
	assert.Contains(t, view, "ANALYZED")
	assert.Contains(t, view, "go tool pprof")
	assert.Contains(t, view, "cpu.prof")
}

// TestCreateStatusBar_WithError tests rendering with an error.
func TestCreateStatusBar_WithError(t *testing.T) {
	state := bubbly.NewRef(composables.StateIdle)
	startTime := bubbly.NewRef(time.Time{})
	filename := bubbly.NewRef("")
	focusedPane := bubbly.NewRef(FocusControls)
	hasResults := bubbly.NewRef(false)
	lastError := bubbly.NewRef("CPU profiling already active")

	bar, err := CreateStatusBar(StatusBarProps{
		State:       state,
		StartTime:   startTime,
		Filename:    filename,
		FocusedPane: focusedPane,
		HasResults:  hasResults,
		LastError:   lastError,
	})

	require.NoError(t, err)
	bar.Init()

	view := bar.View()
	assert.Contains(t, view, "Error")
	assert.Contains(t, view, "CPU profiling already active")
}

// TestCreateStatusBar_FocusPanes tests different focus pane displays.
func TestCreateStatusBar_FocusPanes(t *testing.T) {
	tests := []struct {
		name     string
		pane     FocusPane
		expected string
	}{
		{"profile", FocusProfile, "Profile"},
		{"controls", FocusControls, "Controls"},
		{"results", FocusResults, "Results"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := bubbly.NewRef(composables.StateIdle)
			startTime := bubbly.NewRef(time.Time{})
			filename := bubbly.NewRef("")
			focusedPane := bubbly.NewRef(tt.pane)
			hasResults := bubbly.NewRef(false)
			lastError := bubbly.NewRef("")

			bar, err := CreateStatusBar(StatusBarProps{
				State:       state,
				StartTime:   startTime,
				Filename:    filename,
				FocusedPane: focusedPane,
				HasResults:  hasResults,
				LastError:   lastError,
			})

			require.NoError(t, err)
			bar.Init()

			view := bar.View()
			assert.Contains(t, view, tt.expected)
		})
	}
}

// TestGetHelpText tests the context-aware help text.
func TestGetHelpText(t *testing.T) {
	tests := []struct {
		name        string
		state       composables.CPUProfilerState
		focusedPane FocusPane
		hasResults  bool
		contains    string
	}{
		{"idle controls", composables.StateIdle, FocusControls, false, "Start profiling"},
		{"idle profile", composables.StateIdle, FocusProfile, false, "Switch to Controls"},
		{"profiling controls", composables.StateProfiling, FocusControls, false, "Stop profiling"},
		{"complete no results", composables.StateComplete, FocusControls, false, "Analyze"},
		{"complete with results", composables.StateComplete, FocusControls, true, "Reset"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpText := getHelpText(tt.state, tt.focusedPane, tt.hasResults)
			assert.Contains(t, helpText, tt.contains)
		})
	}
}
