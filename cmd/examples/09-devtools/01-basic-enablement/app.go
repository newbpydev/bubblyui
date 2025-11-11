package main

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/bubblyui/cmd/examples/09-devtools/01-basic-enablement/components"
	"github.com/newbpydev/bubblyui/cmd/examples/09-devtools/01-basic-enablement/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// CreateApp creates the root application component
// This demonstrates:
// - Composable architecture pattern
// - Using UseCounter composable for shared logic
// - Component composition (display + controls)
// - Exposing state for dev tools inspection
func CreateApp() (bubbly.Component, error) {
	builder := bubbly.NewComponent("CounterApp").
		// Key bindings for counter operations
		WithKeyBinding("i", "increment", "Increment counter").
		WithKeyBinding("d", "decrement", "Decrement counter").
		WithKeyBinding("r", "reset", "Reset counter").
		WithKeyBinding("ctrl+c", "quit", "Quit application")

	builder = builder.Setup(func(ctx *bubbly.Context) {
		// Use counter composable - reusable reactive logic!
		counter := composables.UseCounter(ctx, 0)

		// Create child components with props
		display, err := components.CreateCounterDisplay(components.CounterDisplayProps{
			Count:  counter.Count,
			IsEven: counter.IsEven,
		})
		if err != nil {
			ctx.Expose("error", err)
			return
		}

		controls, err := components.CreateCounterControls(components.CounterControlsProps{
			OnIncrement: counter.Increment,
			OnDecrement: counter.Decrement,
			OnReset:     counter.Reset,
		})
		if err != nil {
			ctx.Expose("error", err)
			return
		}

		// Register event handlers
		ctx.On("increment", func(_ interface{}) {
			counter.Increment()
		})

		ctx.On("decrement", func(_ interface{}) {
			counter.Decrement()
		})

		ctx.On("reset", func(_ interface{}) {
			counter.Reset()
		})

		// Expose composable for dev tools inspection
		ctx.Expose("counter", counter)

		// Expose child components (auto-initializes them)
		ctx.ExposeComponent("display", display)
		ctx.ExposeComponent("controls", controls)

		// Lifecycle hooks
		ctx.OnMounted(func() {
			// App mounted - dev tools will show this in component tree
		})
	})

	builder = builder.Template(func(ctx bubbly.RenderContext) string {
		// Get child components
		display := ctx.Get("display").(bubbly.Component)
		controls := ctx.Get("controls").(bubbly.Component)

		// Create title
		titleStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99")).
			MarginBottom(1)
		title := titleStyle.Render("🎯 Dev Tools Example 01: Basic Enablement")

		// Layout: Title, Display, Controls
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			"",
			display.View(),
			"",
			controls.View(),
		)

		// Add padding
		containerStyle := lipgloss.NewStyle().Padding(2)
		return containerStyle.Render(content)
	})

	return builder.Build()
}
