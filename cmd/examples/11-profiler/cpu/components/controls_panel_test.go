package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/cmd/examples/11-profiler/cpu/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestCreateControlsPanel_Creation tests that CreateControlsPanel creates a valid component.
func TestCreateControlsPanel_Creation(t *testing.T) {
	state := bubbly.NewRef(composables.StateIdle)
	focused := bubbly.NewRef(false)
	hasResults := bubbly.NewRef(false)

	panel, err := CreateControlsPanel(ControlsPanelProps{
		State:      state,
		Focused:    focused,
		HasResults: hasResults,
	})

	require.NoError(t, err)
	assert.NotNil(t, panel)
	assert.Equal(t, "ControlsPanel", panel.Name())
}

// TestCreateControlsPanel_IdleState tests rendering in idle state.
func TestCreateControlsPanel_IdleState(t *testing.T) {
	state := bubbly.NewRef(composables.StateIdle)
	focused := bubbly.NewRef(true)
	hasResults := bubbly.NewRef(false)

	panel, err := CreateControlsPanel(ControlsPanelProps{
		State:      state,
		Focused:    focused,
		HasResults: hasResults,
	})

	require.NoError(t, err)
	panel.Init()

	view := panel.View()
	assert.Contains(t, view, "Controls")
	assert.Contains(t, view, "Ready")
	assert.Contains(t, view, "Start Profiling")
}

// TestCreateControlsPanel_ProfilingState tests rendering in profiling state.
func TestCreateControlsPanel_ProfilingState(t *testing.T) {
	state := bubbly.NewRef(composables.StateProfiling)
	focused := bubbly.NewRef(true)
	hasResults := bubbly.NewRef(false)

	panel, err := CreateControlsPanel(ControlsPanelProps{
		State:      state,
		Focused:    focused,
		HasResults: hasResults,
	})

	require.NoError(t, err)
	panel.Init()

	view := panel.View()
	assert.Contains(t, view, "Controls")
	assert.Contains(t, view, "Recording")
	assert.Contains(t, view, "Stop Profiling")
}

// TestCreateControlsPanel_CompleteState tests rendering in complete state.
func TestCreateControlsPanel_CompleteState(t *testing.T) {
	state := bubbly.NewRef(composables.StateComplete)
	focused := bubbly.NewRef(true)
	hasResults := bubbly.NewRef(false)

	panel, err := CreateControlsPanel(ControlsPanelProps{
		State:      state,
		Focused:    focused,
		HasResults: hasResults,
	})

	require.NoError(t, err)
	panel.Init()

	view := panel.View()
	assert.Contains(t, view, "Controls")
	assert.Contains(t, view, "Complete")
	assert.Contains(t, view, "Analyze Results")
	assert.Contains(t, view, "Reset")
}

// TestCreateControlsPanel_AnalyzedState tests rendering when results are available.
func TestCreateControlsPanel_AnalyzedState(t *testing.T) {
	state := bubbly.NewRef(composables.StateComplete)
	focused := bubbly.NewRef(true)
	hasResults := bubbly.NewRef(true)

	panel, err := CreateControlsPanel(ControlsPanelProps{
		State:      state,
		Focused:    focused,
		HasResults: hasResults,
	})

	require.NoError(t, err)
	panel.Init()

	view := panel.View()
	assert.Contains(t, view, "Controls")
	assert.Contains(t, view, "Analyzed")
}

// TestCreateControlsPanel_Callbacks tests that callbacks are invoked.
func TestCreateControlsPanel_Callbacks(t *testing.T) {
	state := bubbly.NewRef(composables.StateIdle)
	focused := bubbly.NewRef(true)
	hasResults := bubbly.NewRef(false)

	startCalled := false
	stopCalled := false
	analyzeCalled := false
	resetCalled := false

	panel, err := CreateControlsPanel(ControlsPanelProps{
		State:      state,
		Focused:    focused,
		HasResults: hasResults,
		OnStart:    func() { startCalled = true },
		OnStop:     func() { stopCalled = true },
		OnAnalyze:  func() { analyzeCalled = true },
		OnReset:    func() { resetCalled = true },
	})

	require.NoError(t, err)
	panel.Init()

	// Emit events and verify callbacks
	panel.Emit("start", nil)
	assert.True(t, startCalled)

	panel.Emit("stop", nil)
	assert.True(t, stopCalled)

	panel.Emit("analyze", nil)
	assert.True(t, analyzeCalled)

	panel.Emit("reset", nil)
	assert.True(t, resetCalled)
}

// TestCreateControlsPanel_NilCallbacks tests that nil callbacks don't panic.
func TestCreateControlsPanel_NilCallbacks(t *testing.T) {
	state := bubbly.NewRef(composables.StateIdle)
	focused := bubbly.NewRef(true)
	hasResults := bubbly.NewRef(false)

	panel, err := CreateControlsPanel(ControlsPanelProps{
		State:      state,
		Focused:    focused,
		HasResults: hasResults,
		// No callbacks provided
	})

	require.NoError(t, err)
	panel.Init()

	// Emit events - should not panic
	assert.NotPanics(t, func() {
		panel.Emit("start", nil)
		panel.Emit("stop", nil)
		panel.Emit("analyze", nil)
		panel.Emit("reset", nil)
	})
}

// TestCreateControlsPanel_FocusIndicator tests the focus indicator.
func TestCreateControlsPanel_FocusIndicator(t *testing.T) {
	tests := []struct {
		name    string
		focused bool
	}{
		{"focused", true},
		{"not focused", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := bubbly.NewRef(composables.StateIdle)
			focused := bubbly.NewRef(tt.focused)
			hasResults := bubbly.NewRef(false)

			panel, err := CreateControlsPanel(ControlsPanelProps{
				State:      state,
				Focused:    focused,
				HasResults: hasResults,
			})

			require.NoError(t, err)
			panel.Init()

			// Just verify it renders without error
			view := panel.View()
			assert.NotEmpty(t, view)
		})
	}
}
