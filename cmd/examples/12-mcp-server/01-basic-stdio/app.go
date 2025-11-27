package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// CreateApp creates the root application component using composable architecture
func CreateApp() (bubbly.Component, error) {
	builder := bubbly.NewComponent("MCPCounterApp").
		WithKeyBinding(" ", "increment", "Increment counter").
		WithKeyBinding("r", "reset", "Reset counter").
		WithKeyBinding("ctrl+c", "quit", "Quit application")

	builder = builder.Setup(func(ctx *bubbly.Context) {
		// Use counter composable for reactive state management
		counter := UseCounter(ctx, 0)

		// Expose state for MCP inspection
		ctx.Expose("count", counter.Count)
		ctx.Expose("isEven", counter.IsEven)

		// Register event handlers
		ctx.On("increment", func(_ interface{}) {
			counter.Increment()
		})

		ctx.On("reset", func(_ interface{}) {
			counter.Reset()
		})

		// Log when component is mounted (stderr for MCP compatibility)
		ctx.OnMounted(func() {
			fmt.Fprintln(os.Stderr, " Counter app mounted - ready for MCP inspection!")
		})
	})

	builder = builder.Template(func(ctx bubbly.RenderContext) string {
		count := ctx.Get("count").(*bubbly.Ref[int]).Get().(int)
		isEven := ctx.Get("isEven").(*bubbly.Computed[interface{}]).Get().(bool)

		// Use BubblyUI Card component for display
		card := components.Card(components.CardProps{
			Title:   "MCP Counter Example",
			Content: renderCounterContent(count, isEven),
			Width:   60,
		})
		card.Init()

		// Use BubblyUI Text component for help
		helpText := components.Text(components.TextProps{
			Content: "💡 Ask your AI: 'What's the current counter value?' or 'Show me the component tree'",
			Color:   lipgloss.Color("240"), // Muted grey
		})
		helpText.Init()

		return lipgloss.JoinVertical(
			lipgloss.Center,
			"",
			card.View(),
			"",
			helpText.View(),
			"",
		)
	})

	return builder.Build()
}

// renderCounterContent creates the content for the counter card using BubblyUI components
func renderCounterContent(count int, isEven bool) string {
	// Use BubblyUI Text component for count display
	countText := components.Text(components.TextProps{
		Content: fmt.Sprintf("Count: %d", count),
		Bold:    true,
	})
	countText.Init()

	// Use BubblyUI Badge component for even/odd indicator
	label := "ODD"
	variant := components.VariantWarning
	if isEven {
		label = "EVEN"
		variant = components.VariantSuccess
	}

	parityBadge := components.Badge(components.BadgeProps{
		Label:   label,
		Variant: variant,
	})
	parityBadge.Init()

	// Use BubblyUI Text component for description
	descText := components.Text(components.TextProps{
		Content: "This counter is exposed to MCP for AI inspection.",
		Color:   lipgloss.Color("240"), // Muted grey
	})
	descText.Init()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		countText.View(),
		"",
		parityBadge.View(),
		"",
		descText.View(),
		"",
	)
}
