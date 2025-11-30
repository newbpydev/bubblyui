package components

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/cmd/examples/11-profiler/cpu/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestFocusPane_String tests the FocusPane String method.
func TestFocusPane_String(t *testing.T) {
	tests := []struct {
		name     string
		pane     FocusPane
		expected string
	}{
		{"profile", FocusProfile, "Profile"},
		{"controls", FocusControls, "Controls"},
		{"results", FocusResults, "Results"},
		{"unknown", FocusPane(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.pane.String())
		})
	}
}

// TestCreateProfilePanel_Creation tests that CreateProfilePanel creates a valid component.
func TestCreateProfilePanel_Creation(t *testing.T) {
	state := bubbly.NewRef(composables.StateIdle)
	filename := bubbly.NewRef("")
	startTime := bubbly.NewRef(time.Time{})
	fileSize := bubbly.NewRef(int64(0))
	focused := bubbly.NewRef(false)

	panel, err := CreateProfilePanel(ProfilePanelProps{
		State:     state,
		Filename:  filename,
		StartTime: startTime,
		FileSize:  fileSize,
		Focused:   focused,
	})

	require.NoError(t, err)
	assert.NotNil(t, panel)
	assert.Equal(t, "ProfilePanel", panel.Name())
}

// TestCreateProfilePanel_IdleState tests rendering in idle state.
func TestCreateProfilePanel_IdleState(t *testing.T) {
	state := bubbly.NewRef(composables.StateIdle)
	filename := bubbly.NewRef("")
	startTime := bubbly.NewRef(time.Time{})
	fileSize := bubbly.NewRef(int64(0))
	focused := bubbly.NewRef(false)

	panel, err := CreateProfilePanel(ProfilePanelProps{
		State:     state,
		Filename:  filename,
		StartTime: startTime,
		FileSize:  fileSize,
		Focused:   focused,
	})

	require.NoError(t, err)
	panel.Init()

	view := panel.View()
	assert.Contains(t, view, "CPU Profile")
	assert.Contains(t, view, "No profile active")
	assert.Contains(t, view, "Space")
}

// TestCreateProfilePanel_ProfilingState tests rendering in profiling state.
func TestCreateProfilePanel_ProfilingState(t *testing.T) {
	state := bubbly.NewRef(composables.StateProfiling)
	filename := bubbly.NewRef("cpu.prof")
	startTime := bubbly.NewRef(time.Now().Add(-5 * time.Second))
	fileSize := bubbly.NewRef(int64(0))
	focused := bubbly.NewRef(true)

	panel, err := CreateProfilePanel(ProfilePanelProps{
		State:     state,
		Filename:  filename,
		StartTime: startTime,
		FileSize:  fileSize,
		Focused:   focused,
	})

	require.NoError(t, err)
	panel.Init()

	view := panel.View()
	assert.Contains(t, view, "CPU Profile")
	assert.Contains(t, view, "cpu.prof")
	assert.Contains(t, view, "Duration")
	assert.Contains(t, view, "Recording")
}

// TestCreateProfilePanel_CompleteState tests rendering in complete state.
func TestCreateProfilePanel_CompleteState(t *testing.T) {
	state := bubbly.NewRef(composables.StateComplete)
	filename := bubbly.NewRef("cpu.prof")
	startTime := bubbly.NewRef(time.Now().Add(-10 * time.Second))
	fileSize := bubbly.NewRef(int64(1024 * 1024)) // 1 MB
	focused := bubbly.NewRef(false)

	panel, err := CreateProfilePanel(ProfilePanelProps{
		State:     state,
		Filename:  filename,
		StartTime: startTime,
		FileSize:  fileSize,
		Focused:   focused,
	})

	require.NoError(t, err)
	panel.Init()

	view := panel.View()
	assert.Contains(t, view, "CPU Profile")
	assert.Contains(t, view, "cpu.prof")
	assert.Contains(t, view, "1.0 MB")
}

// TestCreateProfilePanel_FocusIndicator tests the focus indicator.
func TestCreateProfilePanel_FocusIndicator(t *testing.T) {
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
			filename := bubbly.NewRef("")
			startTime := bubbly.NewRef(time.Time{})
			fileSize := bubbly.NewRef(int64(0))
			focused := bubbly.NewRef(tt.focused)

			panel, err := CreateProfilePanel(ProfilePanelProps{
				State:     state,
				Filename:  filename,
				StartTime: startTime,
				FileSize:  fileSize,
				Focused:   focused,
			})

			require.NoError(t, err)
			panel.Init()

			// Just verify it renders without error
			view := panel.View()
			assert.NotEmpty(t, view)
		})
	}
}
