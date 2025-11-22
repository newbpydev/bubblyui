package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/cmd/examples/11-advanced-patterns/01-shared-state/components"
	localComposables "github.com/newbpydev/bubblyui/cmd/examples/11-advanced-patterns/01-shared-state/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// CreateApp creates the root application component demonstrating shared state
func CreateApp() (bubbly.Component, error) {
	return bubbly.NewComponent("SharedStateApp").
		WithMultiKeyBindings("increment", "Increment counter", "up", "k", "+").
		WithMultiKeyBindings("decrement", "Decrement counter", "down", "j", "-").
		WithKeyBinding("r", "reset", "Reset counter").
		WithKeyBinding("q", "quit", "Quit application").
		WithKeyBinding("ctrl+c", "quit", "Quit application").
		Setup(func(ctx *bubbly.Context) {
			// Get shared counter instance
			counter := localComposables.UseSharedCounter(ctx)

			// Create display component (reads shared counter)
			display, err := components.CreateCounterDisplay(components.CounterDisplayProps{
				Count:   counter.Count,
				Doubled: counter.Doubled,
				IsEven:  counter.IsEven,
				History: counter.History,
			})
			if err != nil {
				panic(err)
			}

			// Create controls component (modifies shared counter)
			controls, err := components.CreateCounterControls(components.CounterControlsProps{
				OnIncrement: counter.Increment,
				OnDecrement: counter.Decrement,
				OnReset:     counter.Reset,
			})
			if err != nil {
				panic(err)
			}

			// Expose components
			if err := ctx.ExposeComponent("display", display); err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to expose display: %v", err))
				return
			}
			if err := ctx.ExposeComponent("controls", controls); err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to expose controls: %v", err))
				return
			}

			// Event handlers
			ctx.On("increment", func(_ interface{}) { counter.Increment() })
			ctx.On("decrement", func(_ interface{}) { counter.Decrement() })
			ctx.On("reset", func(_ interface{}) { counter.Reset() })
		}).
		Template(func(ctx bubbly.RenderContext) string {
			display := ctx.Get("display").(bubbly.Component)
			controls := ctx.Get("controls").(bubbly.Component)

			// Styles
			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205")).
				MarginBottom(1)

			subtitleStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("99")).
				Italic(true).
				MarginBottom(1)

			helpStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				MarginTop(2)

			// Layout - side by side
			displayView := display.View()
			controlsView := controls.View()

			combined := lipgloss.JoinHorizontal(
				lipgloss.Top,
				displayView,
				"  ", // Spacing
				controlsView,
			)

			// Build full view
			title := titleStyle.Render("🔄 Shared State Example")
			subtitle := subtitleStyle.Render("Two components, one counter - powered by CreateShared()")
			help := helpStyle.Render("q: quit • Try incrementing/decrementing - both components update!")

			return fmt.Sprintf("%s\n%s\n\n%s\n\n%s\n", title, subtitle, combined, help)
		}).
		Build()
}
