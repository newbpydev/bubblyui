package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestCreateControlsPanel_Creation tests that CreateControlsPanel creates a valid component.
func TestCreateControlsPanel_Creation(t *testing.T) {
	isRunning := bubbly.NewRef(false)
	focused := bubbly.NewRef(true)

	panel, err := CreateControlsPanel(ControlsPanelProps{
		IsRunning: isRunning,
		Focused:   focused,
		OnToggle:  func() {},
		OnReset:   func() {},
		OnExport:  func() {},
	})

	require.NoError(t, err)
	assert.NotNil(t, panel)
	assert.Equal(t, "ControlsPanel", panel.Name())
}

// TestCreateControlsPanel_Render tests that the panel renders correctly.
func TestCreateControlsPanel_Render(t *testing.T) {
	isRunning := bubbly.NewRef(false)
	focused := bubbly.NewRef(true)

	panel, err := CreateControlsPanel(ControlsPanelProps{
		IsRunning: isRunning,
		Focused:   focused,
		OnToggle:  func() {},
		OnReset:   func() {},
		OnExport:  func() {},
	})

	require.NoError(t, err)
	panel.Init()

	view := panel.View()
	assert.Contains(t, view, "Controls")
	assert.Contains(t, view, "Space")
	assert.Contains(t, view, "Reset")
	assert.Contains(t, view, "Export")
}

// TestCreateControlsPanel_RunningState tests the running state display.
func TestCreateControlsPanel_RunningState(t *testing.T) {
	isRunning := bubbly.NewRef(true)
	focused := bubbly.NewRef(true)

	panel, err := CreateControlsPanel(ControlsPanelProps{
		IsRunning: isRunning,
		Focused:   focused,
		OnToggle:  func() {},
		OnReset:   func() {},
		OnExport:  func() {},
	})

	require.NoError(t, err)
	panel.Init()

	view := panel.View()
	assert.Contains(t, view, "Stop")
}

// TestCreateControlsPanel_StoppedState tests the stopped state display.
func TestCreateControlsPanel_StoppedState(t *testing.T) {
	isRunning := bubbly.NewRef(false)
	focused := bubbly.NewRef(true)

	panel, err := CreateControlsPanel(ControlsPanelProps{
		IsRunning: isRunning,
		Focused:   focused,
		OnToggle:  func() {},
		OnReset:   func() {},
		OnExport:  func() {},
	})

	require.NoError(t, err)
	panel.Init()

	view := panel.View()
	assert.Contains(t, view, "Start")
}

// TestCreateControlsPanel_FocusIndicator tests the focus indicator.
func TestCreateControlsPanel_FocusIndicator(t *testing.T) {
	tests := []struct {
		name     string
		focused  bool
		contains string
	}{
		{"focused", true, "Active"},
		{"unfocused", false, "Inactive"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isRunning := bubbly.NewRef(false)
			focused := bubbly.NewRef(tt.focused)

			panel, err := CreateControlsPanel(ControlsPanelProps{
				IsRunning: isRunning,
				Focused:   focused,
				OnToggle:  func() {},
				OnReset:   func() {},
				OnExport:  func() {},
			})

			require.NoError(t, err)
			panel.Init()

			view := panel.View()
			assert.Contains(t, view, tt.contains)
		})
	}
}

// TestCreateControlsPanel_Callbacks tests that callbacks are called.
func TestCreateControlsPanel_Callbacks(t *testing.T) {
	isRunning := bubbly.NewRef(false)
	focused := bubbly.NewRef(true)

	toggleCalled := false
	resetCalled := false
	exportCalled := false

	panel, err := CreateControlsPanel(ControlsPanelProps{
		IsRunning: isRunning,
		Focused:   focused,
		OnToggle:  func() { toggleCalled = true },
		OnReset:   func() { resetCalled = true },
		OnExport:  func() { exportCalled = true },
	})

	require.NoError(t, err)
	panel.Init()

	// Emit events
	panel.Emit("toggle", nil)
	panel.Emit("reset", nil)
	panel.Emit("export", nil)

	assert.True(t, toggleCalled, "OnToggle should be called")
	assert.True(t, resetCalled, "OnReset should be called")
	assert.True(t, exportCalled, "OnExport should be called")
}

// TestCreateControlsPanel_NilCallbacks tests that nil callbacks don't panic.
func TestCreateControlsPanel_NilCallbacks(t *testing.T) {
	isRunning := bubbly.NewRef(false)
	focused := bubbly.NewRef(true)

	panel, err := CreateControlsPanel(ControlsPanelProps{
		IsRunning: isRunning,
		Focused:   focused,
		OnToggle:  nil,
		OnReset:   nil,
		OnExport:  nil,
	})

	require.NoError(t, err)
	panel.Init()

	// These should not panic
	assert.NotPanics(t, func() {
		panel.Emit("toggle", nil)
		panel.Emit("reset", nil)
		panel.Emit("export", nil)
	})
}
