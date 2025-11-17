package testutil

import (
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewKeyBindingsTester tests creating a new KeyBindingsTester
func TestNewKeyBindingsTester(t *testing.T) {
	// Create a component with key bindings
	comp, err := bubbly.NewComponent("TestComponent").
		WithKeyBinding(" ", "increment", "Increment counter").
		WithKeyBinding("r", "reset", "Reset counter").
		Setup(func(ctx *bubbly.Context) {
			// Empty setup
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Test"
		}).
		Build()
	require.NoError(t, err)

	// Create tester
	tester := NewKeyBindingsTester(comp)

	// Assert tester created
	assert.NotNil(t, tester)
	assert.NotNil(t, tester.component)
	assert.NotNil(t, tester.bindings)
	assert.Equal(t, comp, tester.component)
}

// TestKeyBindingsTester_SimulateKeyPress tests key press simulation
func TestKeyBindingsTester_SimulateKeyPress(t *testing.T) {
	// Create a component with key bindings
	eventFired := false
	comp, err := bubbly.NewComponent("TestComponent").
		WithKeyBinding(" ", "increment", "Increment counter").
		Setup(func(ctx *bubbly.Context) {
			ctx.On("increment", func(_ interface{}) {
				eventFired = true
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Test"
		}).
		Build()
	require.NoError(t, err)

	// Initialize component
	comp.Init()

	// Create tester
	tester := NewKeyBindingsTester(comp)

	// Simulate space key press
	cmd := tester.SimulateKeyPress(" ")

	// Assert event was fired
	assert.True(t, eventFired, "Expected event to be fired")

	// Command may or may not be nil depending on auto-commands
	_ = cmd
}

// TestKeyBindingsTester_SimulateKeyPress_MultipleKeys tests multiple key presses
func TestKeyBindingsTester_SimulateKeyPress_MultipleKeys(t *testing.T) {
	incrementCount := 0
	resetCount := 0

	comp, err := bubbly.NewComponent("TestComponent").
		WithKeyBinding(" ", "increment", "Increment counter").
		WithKeyBinding("r", "reset", "Reset counter").
		Setup(func(ctx *bubbly.Context) {
			ctx.On("increment", func(_ interface{}) {
				incrementCount++
			})
			ctx.On("reset", func(_ interface{}) {
				resetCount++
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Test"
		}).
		Build()
	require.NoError(t, err)

	comp.Init()

	tester := NewKeyBindingsTester(comp)

	// Simulate multiple key presses
	tester.SimulateKeyPress(" ")
	tester.SimulateKeyPress(" ")
	tester.SimulateKeyPress("r")

	// Assert events fired correctly
	assert.Equal(t, 2, incrementCount)
	assert.Equal(t, 1, resetCount)
}

// TestKeyBindingsTester_AssertHelpText tests help text assertion
func TestKeyBindingsTester_AssertHelpText(t *testing.T) {
	comp, err := bubbly.NewComponent("TestComponent").
		WithKeyBinding(" ", "increment", "Increment counter").
		WithKeyBinding("r", "reset", "Reset counter").
		Setup(func(ctx *bubbly.Context) {
			// Empty setup
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Test"
		}).
		Build()
	require.NoError(t, err)

	tester := NewKeyBindingsTester(comp)

	// Assert help text (should contain both bindings)
	// Format: "key: description • key: description"
	tester.AssertHelpText(t, " : Increment counter • r: Reset counter")
}

// TestKeyBindingsTester_AssertHelpText_Failure tests help text assertion failure
func TestKeyBindingsTester_AssertHelpText_Failure(t *testing.T) {
	comp, err := bubbly.NewComponent("TestComponent").
		WithKeyBinding(" ", "increment", "Increment counter").
		Setup(func(ctx *bubbly.Context) {
			// Empty setup
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Test"
		}).
		Build()
	require.NoError(t, err)

	tester := NewKeyBindingsTester(comp)

	// Create mock testing.T to capture error
	mockT := &mockTestingT{}

	// Assert wrong help text (should fail)
	tester.AssertHelpText(mockT, "wrong help text")

	// Verify error was called
	assert.True(t, len(mockT.errors) > 0, "Expected error to be recorded")
	assert.Contains(t, mockT.errors[0], "expected help text")
}

// TestKeyBindingsTester_DetectConflicts tests conflict detection
func TestKeyBindingsTester_DetectConflicts(t *testing.T) {
	tests := []struct {
		name              string
		bindings          []struct{ key, event, desc string }
		expectedConflicts int
	}{
		{
			name: "no conflicts",
			bindings: []struct{ key, event, desc string }{
				{" ", "increment", "Increment"},
				{"r", "reset", "Reset"},
			},
			expectedConflicts: 0,
		},
		{
			name: "one conflict",
			bindings: []struct{ key, event, desc string }{
				{" ", "increment", "Increment"},
				{" ", "toggle", "Toggle"},
			},
			expectedConflicts: 1,
		},
		{
			name: "multiple conflicts",
			bindings: []struct{ key, event, desc string }{
				{" ", "increment", "Increment"},
				{" ", "toggle", "Toggle"},
				{"r", "reset", "Reset"},
				{"r", "refresh", "Refresh"},
			},
			expectedConflicts: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build component with bindings
			builder := bubbly.NewComponent("TestComponent")
			for _, b := range tt.bindings {
				builder = builder.WithKeyBinding(b.key, b.event, b.desc)
			}
			comp, err := builder.
				Setup(func(ctx *bubbly.Context) {
					// Empty setup
				}).
				Template(func(ctx bubbly.RenderContext) string {
					return "Test"
				}).
				Build()
			require.NoError(t, err)

			tester := NewKeyBindingsTester(comp)

			// Detect conflicts
			conflicts := tester.DetectConflicts()

			// Assert conflict count
			assert.Len(t, conflicts, tt.expectedConflicts)
		})
	}
}

// TestKeyBindingsTester_GetBindings tests retrieving bindings
func TestKeyBindingsTester_GetBindings(t *testing.T) {
	comp, err := bubbly.NewComponent("TestComponent").
		WithKeyBinding(" ", "increment", "Increment counter").
		WithKeyBinding("r", "reset", "Reset counter").
		Setup(func(ctx *bubbly.Context) {
			// Empty setup
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Test"
		}).
		Build()
	require.NoError(t, err)

	tester := NewKeyBindingsTester(comp)

	// Get bindings
	bindings := tester.bindings

	// Assert bindings retrieved
	assert.NotNil(t, bindings)
	assert.Len(t, bindings, 2)
	assert.Contains(t, bindings, " ")
	assert.Contains(t, bindings, "r")
}

// TestKeyBindingsTester_SimulateKeyPress_UnknownKey tests unknown key handling
func TestKeyBindingsTester_SimulateKeyPress_UnknownKey(t *testing.T) {
	comp, err := bubbly.NewComponent("TestComponent").
		WithKeyBinding(" ", "increment", "Increment counter").
		Setup(func(ctx *bubbly.Context) {
			// Empty setup
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Test"
		}).
		Build()
	require.NoError(t, err)

	comp.Init()

	tester := NewKeyBindingsTester(comp)

	// Simulate unknown key (should not panic)
	cmd := tester.SimulateKeyPress("unknown")

	// Should return nil or a valid command
	_ = cmd
}

// TestKeyBindingsTester_SimulateKeyPress_SpecialKeys tests special key handling
func TestKeyBindingsTester_SimulateKeyPress_SpecialKeys(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{"space", " "},
		{"enter", "enter"},
		{"esc", "esc"},
		{"ctrl+c", "ctrl+c"},
		{"up", "up"},
		{"down", "down"},
		{"left", "left"},
		{"right", "right"},
		{"tab", "tab"},
		{"backspace", "backspace"},
		{"delete", "delete"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventFired := false
			comp, err := bubbly.NewComponent("TestComponent").
				WithKeyBinding(tt.key, "action", "Do action").
				Setup(func(ctx *bubbly.Context) {
					ctx.On("action", func(_ interface{}) {
						eventFired = true
					})
				}).
				Template(func(ctx bubbly.RenderContext) string {
					return "Test"
				}).
				Build()
			require.NoError(t, err)

			comp.Init()

			tester := NewKeyBindingsTester(comp)

			// Simulate key press
			tester.SimulateKeyPress(tt.key)

			// Assert event was fired
			assert.True(t, eventFired, "Expected event to be fired for key: %s", tt.key)
		})
	}
}

// TestKeyBindingsTester_SimulateKeyPress_SingleCharKeys tests single character key handling
func TestKeyBindingsTester_SimulateKeyPress_SingleCharKeys(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{"letter_a", "a"},
		{"letter_z", "z"},
		{"number_1", "1"},
		{"number_9", "9"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventFired := false
			comp, err := bubbly.NewComponent("TestComponent").
				WithKeyBinding(tt.key, "action", "Do action").
				Setup(func(ctx *bubbly.Context) {
					ctx.On("action", func(_ interface{}) {
						eventFired = true
					})
				}).
				Template(func(ctx bubbly.RenderContext) string {
					return "Test"
				}).
				Build()
			require.NoError(t, err)

			comp.Init()

			tester := NewKeyBindingsTester(comp)

			// Simulate key press
			tester.SimulateKeyPress(tt.key)

			// Assert event was fired
			assert.True(t, eventFired, "Expected event to be fired for key: %s", tt.key)
		})
	}
}

// TestKeyBindingsTester_SimulateKeyPress_ComplexKeys tests complex key combinations
func TestKeyBindingsTester_SimulateKeyPress_ComplexKeys(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{"ctrl+a", "ctrl+a"},
		{"ctrl+d", "ctrl+d"},
		{"alt+enter", "alt+enter"},
		{"shift+tab", "shift+tab"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventFired := false
			comp, err := bubbly.NewComponent("TestComponent").
				WithKeyBinding(tt.key, "action", "Do action").
				Setup(func(ctx *bubbly.Context) {
					ctx.On("action", func(_ interface{}) {
						eventFired = true
					})
				}).
				Template(func(ctx bubbly.RenderContext) string {
					return "Test"
				}).
				Build()
			require.NoError(t, err)

			comp.Init()

			tester := NewKeyBindingsTester(comp)

			// Simulate key press
			tester.SimulateKeyPress(tt.key)

			// Assert event was fired
			assert.True(t, eventFired, "Expected event to be fired for key: %s", tt.key)
		})
	}
}
