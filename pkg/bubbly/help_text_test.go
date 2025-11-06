package bubbly

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestComponent_HelpText tests the HelpText() method with various scenarios
func TestComponent_HelpText(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() Component
		expected string
	}{
		{
			name: "empty bindings returns empty string",
			setup: func() Component {
				component, err := NewComponent("TestComponent").
					Template(func(ctx RenderContext) string {
						return "test"
					}).
					Build()
				require.NoError(t, err)
				return component
			},
			expected: "",
		},
		{
			name: "single binding formatted correctly",
			setup: func() Component {
				component, err := NewComponent("TestComponent").
					WithKeyBinding("space", "increment", "Increment counter").
					Template(func(ctx RenderContext) string {
						return "test"
					}).
					Build()
				require.NoError(t, err)
				return component
			},
			expected: "space: Increment counter",
		},
		{
			name: "multiple bindings sorted alphabetically",
			setup: func() Component {
				component, err := NewComponent("TestComponent").
					WithKeyBinding("space", "increment", "Increment counter").
					WithKeyBinding("ctrl+c", "quit", "Quit application").
					WithKeyBinding("down", "next", "Next item").
					Template(func(ctx RenderContext) string {
						return "test"
					}).
					Build()
				require.NoError(t, err)
				return component
			},
			expected: "ctrl+c: Quit application • down: Next item • space: Increment counter",
		},
		{
			name: "duplicate keys show first description",
			setup: func() Component {
				component, err := NewComponent("TestComponent").
					WithConditionalKeyBinding(KeyBinding{
						Key:         "space",
						Event:       "toggle",
						Description: "Toggle in navigation mode",
						Condition:   func() bool { return true },
					}).
					WithConditionalKeyBinding(KeyBinding{
						Key:         "space",
						Event:       "addChar",
						Description: "Add space in input mode",
						Condition:   func() bool { return false },
					}).
					Template(func(ctx RenderContext) string {
						return "test"
					}).
					Build()
				require.NoError(t, err)
				return component
			},
			expected: "space: Toggle in navigation mode",
		},
		{
			name: "empty descriptions skipped",
			setup: func() Component {
				component, err := NewComponent("TestComponent").
					WithKeyBinding("space", "increment", "Increment counter").
					WithKeyBinding("enter", "submit", ""). // Empty description
					WithKeyBinding("ctrl+c", "quit", "Quit application").
					Template(func(ctx RenderContext) string {
						return "test"
					}).
					Build()
				require.NoError(t, err)
				return component
			},
			expected: "ctrl+c: Quit application • space: Increment counter",
		},
		{
			name: "separator formatting correct",
			setup: func() Component {
				component, err := NewComponent("TestComponent").
					WithKeyBinding("a", "action1", "Action 1").
					WithKeyBinding("b", "action2", "Action 2").
					WithKeyBinding("c", "action3", "Action 3").
					Template(func(ctx RenderContext) string {
						return "test"
					}).
					Build()
				require.NoError(t, err)
				return component
			},
			expected: "a: Action 1 • b: Action 2 • c: Action 3",
		},
		{
			name: "all empty descriptions returns empty string",
			setup: func() Component {
				component, err := NewComponent("TestComponent").
					WithKeyBinding("space", "increment", "").
					WithKeyBinding("enter", "submit", "").
					Template(func(ctx RenderContext) string {
						return "test"
					}).
					Build()
				require.NoError(t, err)
				return component
			},
			expected: "",
		},
		{
			name: "complex keys formatted correctly",
			setup: func() Component {
				component, err := NewComponent("TestComponent").
					WithKeyBinding("ctrl+shift+a", "action1", "Complex action").
					WithKeyBinding("alt+enter", "action2", "Alternative action").
					WithKeyBinding("up", "action3", "Arrow action").
					Template(func(ctx RenderContext) string {
						return "test"
					}).
					Build()
				require.NoError(t, err)
				return component
			},
			expected: "alt+enter: Alternative action • ctrl+shift+a: Complex action • up: Arrow action",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := tt.setup()
			result := component.HelpText()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestComponent_HelpText_ThreadSafety tests concurrent access to HelpText()
func TestComponent_HelpText_ThreadSafety(t *testing.T) {
	component, err := NewComponent("TestComponent").
		WithKeyBinding("space", "increment", "Increment counter").
		WithKeyBinding("ctrl+c", "quit", "Quit application").
		Template(func(ctx RenderContext) string {
			return "test"
		}).
		Build()
	require.NoError(t, err)

	// Run HelpText() concurrently from multiple goroutines
	const numGoroutines = 100
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			result := component.HelpText()
			assert.NotEmpty(t, result)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

// TestComponent_HelpText_Integration tests HelpText() with template integration
func TestComponent_HelpText_Integration(t *testing.T) {
	component, err := NewComponent("Counter").
		WithKeyBinding("space", "increment", "Increment counter").
		WithKeyBinding("r", "reset", "Reset counter").
		WithKeyBinding("ctrl+c", "quit", "Quit").
		Setup(func(ctx *Context) {
			count := ctx.Ref(0)
			ctx.Expose("count", count)
		}).
		Template(func(ctx RenderContext) string {
			comp := ctx.component
			helpText := comp.HelpText()
			return "Counter\n\nHelp: " + helpText
		}).
		Build()
	require.NoError(t, err)

	// Initialize component
	component.Init()

	// Render and verify help text is included
	view := component.View()
	assert.Contains(t, view, "ctrl+c: Quit")
	assert.Contains(t, view, "r: Reset counter")
	assert.Contains(t, view, "space: Increment counter")
	assert.Contains(t, view, " • ")
}
