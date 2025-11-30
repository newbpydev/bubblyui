package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/cmd/examples/11-profiler/cpu/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestCreateResultsPanel_Creation tests that CreateResultsPanel creates a valid component.
func TestCreateResultsPanel_Creation(t *testing.T) {
	state := bubbly.NewRef(composables.StateIdle)
	hotFunctions := bubbly.NewRef([]composables.HotFunctionInfo{})
	filename := bubbly.NewRef("")
	focused := bubbly.NewRef(false)

	panel, err := CreateResultsPanel(ResultsPanelProps{
		State:        state,
		HotFunctions: hotFunctions,
		Filename:     filename,
		Focused:      focused,
	})

	require.NoError(t, err)
	assert.NotNil(t, panel)
	assert.Equal(t, "ResultsPanel", panel.Name())
}

// TestCreateResultsPanel_IdleState tests rendering in idle state.
func TestCreateResultsPanel_IdleState(t *testing.T) {
	state := bubbly.NewRef(composables.StateIdle)
	hotFunctions := bubbly.NewRef([]composables.HotFunctionInfo{})
	filename := bubbly.NewRef("")
	focused := bubbly.NewRef(false)

	panel, err := CreateResultsPanel(ResultsPanelProps{
		State:        state,
		HotFunctions: hotFunctions,
		Filename:     filename,
		Focused:      focused,
	})

	require.NoError(t, err)
	panel.Init()

	view := panel.View()
	assert.Contains(t, view, "Hot Functions")
	assert.Contains(t, view, "No results yet")
}

// TestCreateResultsPanel_ProfilingState tests rendering in profiling state.
func TestCreateResultsPanel_ProfilingState(t *testing.T) {
	state := bubbly.NewRef(composables.StateProfiling)
	hotFunctions := bubbly.NewRef([]composables.HotFunctionInfo{})
	filename := bubbly.NewRef("cpu.prof")
	focused := bubbly.NewRef(false)

	panel, err := CreateResultsPanel(ResultsPanelProps{
		State:        state,
		HotFunctions: hotFunctions,
		Filename:     filename,
		Focused:      focused,
	})

	require.NoError(t, err)
	panel.Init()

	view := panel.View()
	assert.Contains(t, view, "Hot Functions")
	assert.Contains(t, view, "Profiling in progress")
}

// TestCreateResultsPanel_CompleteNoResults tests rendering in complete state without results.
func TestCreateResultsPanel_CompleteNoResults(t *testing.T) {
	state := bubbly.NewRef(composables.StateComplete)
	hotFunctions := bubbly.NewRef([]composables.HotFunctionInfo{})
	filename := bubbly.NewRef("cpu.prof")
	focused := bubbly.NewRef(false)

	panel, err := CreateResultsPanel(ResultsPanelProps{
		State:        state,
		HotFunctions: hotFunctions,
		Filename:     filename,
		Focused:      focused,
	})

	require.NoError(t, err)
	panel.Init()

	view := panel.View()
	assert.Contains(t, view, "Hot Functions")
	assert.Contains(t, view, "Profile complete")
	assert.Contains(t, view, "analyze")
}

// TestCreateResultsPanel_WithResults tests rendering with hot functions.
func TestCreateResultsPanel_WithResults(t *testing.T) {
	state := bubbly.NewRef(composables.StateComplete)
	hotFunctions := bubbly.NewRef([]composables.HotFunctionInfo{
		{Name: "runtime.mallocgc", Samples: 1250, Percent: 25.0},
		{Name: "main.processData", Samples: 800, Percent: 16.0},
		{Name: "encoding/json.decode", Samples: 650, Percent: 13.0},
	})
	filename := bubbly.NewRef("cpu.prof")
	focused := bubbly.NewRef(true)

	panel, err := CreateResultsPanel(ResultsPanelProps{
		State:        state,
		HotFunctions: hotFunctions,
		Filename:     filename,
		Focused:      focused,
	})

	require.NoError(t, err)
	panel.Init()

	view := panel.View()
	assert.Contains(t, view, "Hot Functions")
	assert.Contains(t, view, "Top CPU Consumers")
	assert.Contains(t, view, "runtime.mallocgc")
	assert.Contains(t, view, "25.0%")
	assert.Contains(t, view, "1250 samples")
	assert.Contains(t, view, "go tool pprof")
}

// TestCreateResultsPanel_LongFunctionNames tests truncation of long function names.
func TestCreateResultsPanel_LongFunctionNames(t *testing.T) {
	state := bubbly.NewRef(composables.StateComplete)
	hotFunctions := bubbly.NewRef([]composables.HotFunctionInfo{
		{Name: "github.com/very/long/package/path/to/some/deeply/nested/function.VeryLongFunctionName", Samples: 100, Percent: 10.0},
	})
	filename := bubbly.NewRef("cpu.prof")
	focused := bubbly.NewRef(false)

	panel, err := CreateResultsPanel(ResultsPanelProps{
		State:        state,
		HotFunctions: hotFunctions,
		Filename:     filename,
		Focused:      focused,
	})

	require.NoError(t, err)
	panel.Init()

	view := panel.View()
	assert.Contains(t, view, "...")
}

// TestCreateResultsPanel_MoreThanFive tests the "and X more" message.
func TestCreateResultsPanel_MoreThanFive(t *testing.T) {
	state := bubbly.NewRef(composables.StateComplete)
	hotFunctions := bubbly.NewRef([]composables.HotFunctionInfo{
		{Name: "func1", Samples: 100, Percent: 20.0},
		{Name: "func2", Samples: 90, Percent: 18.0},
		{Name: "func3", Samples: 80, Percent: 16.0},
		{Name: "func4", Samples: 70, Percent: 14.0},
		{Name: "func5", Samples: 60, Percent: 12.0},
		{Name: "func6", Samples: 50, Percent: 10.0},
		{Name: "func7", Samples: 40, Percent: 8.0},
	})
	filename := bubbly.NewRef("cpu.prof")
	focused := bubbly.NewRef(false)

	panel, err := CreateResultsPanel(ResultsPanelProps{
		State:        state,
		HotFunctions: hotFunctions,
		Filename:     filename,
		Focused:      focused,
	})

	require.NoError(t, err)
	panel.Init()

	view := panel.View()
	assert.Contains(t, view, "and 2 more")
}

// TestCreateResultsPanel_FocusIndicator tests the focus indicator.
func TestCreateResultsPanel_FocusIndicator(t *testing.T) {
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
			hotFunctions := bubbly.NewRef([]composables.HotFunctionInfo{})
			filename := bubbly.NewRef("")
			focused := bubbly.NewRef(tt.focused)

			panel, err := CreateResultsPanel(ResultsPanelProps{
				State:        state,
				HotFunctions: hotFunctions,
				Filename:     filename,
				Focused:      focused,
			})

			require.NoError(t, err)
			panel.Init()

			// Just verify it renders without error
			view := panel.View()
			assert.NotEmpty(t, view)
		})
	}
}

// TestRenderHotFunctions_Empty tests rendering empty hot functions.
func TestRenderHotFunctions_Empty(t *testing.T) {
	result := renderHotFunctions([]composables.HotFunctionInfo{})
	assert.Contains(t, result, "No hot functions found")
}
