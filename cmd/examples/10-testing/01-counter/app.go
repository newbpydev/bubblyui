package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/bubblyui/cmd/examples/10-testing/01-counter/components"
	"github.com/newbpydev/bubblyui/cmd/examples/10-testing/01-counter/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// CreateApp creates the root counter application component
func CreateApp() (bubbly.Component, error) {
	return bubbly.NewComponent("CounterApp").
		WithMultiKeyBindings("increment", "Increment counter", "up", "k", "+").
		WithMultiKeyBindings("decrement", "Decrement counter", "down", "j", "-").
		WithKeyBinding("r", "reset", "Reset counter").
		WithKeyBinding("q", "quit", "Quit application").
		WithKeyBinding("ctrl+c", "quit", "Quit application").
		Setup(func(ctx *bubbly.Context) {
			// Use the counter composable for state management
			counter := composables.UseCounter(ctx, 0)

			// Create display component with counter state
			display, err := components.CreateCounterDisplay(components.CounterDisplayProps{
				Count:   counter.Count,
				Doubled: counter.Doubled,
				IsEven:  counter.IsEven,
				History: counter.History,
			})
			if err != nil {
				panic(err)
			}

			// Expose display component
			ctx.ExposeComponent("display", display)

			// Expose counter for testing
			ctx.Expose("counter", counter)

			// Event handlers - delegate to composable
			ctx.On("increment", func(data interface{}) {
				counter.Increment()
			})

			ctx.On("decrement", func(data interface{}) {
				counter.Decrement()
			})

			ctx.On("reset", func(data interface{}) {
				counter.Reset()
			})

			ctx.On("set", func(data interface{}) {
				if val, ok := data.(int); ok {
					counter.SetValue(val)
				}
			})

			// Handle quit event
			ctx.On("quit", func(data interface{}) {
				// Quit is handled by bubbly.Wrapper automatically
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			display := ctx.Get("display").(bubbly.Component)

			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205")).
				MarginBottom(1)

			title := titleStyle.Render("🔢 Counter App - Testing Example")

			helpStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				MarginTop(2)

			help := helpStyle.Render(
				"↑/k/+: increment • ↓/j/-: decrement • r: reset • q: quit",
			)

			return fmt.Sprintf("%s\n\n%s\n%s\n", title, display.View(), help)
		}).
		Build()
}
