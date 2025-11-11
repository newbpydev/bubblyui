package components

import (
	"fmt"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// CounterDisplayProps defines the props for CounterDisplay component
type CounterDisplayProps struct {
	Count  *bubbly.Ref[int]
	IsEven *bubbly.Computed[interface{}]
}

// CreateCounterDisplay creates a component that displays the current count
// This demonstrates:
// - Component factory pattern
// - Props-based composition
// - Using BubblyUI Card component
// - Exposing state for dev tools inspection
func CreateCounterDisplay(props CounterDisplayProps) (bubbly.Component, error) {
	builder := bubbly.NewComponent("CounterDisplay")

	builder = builder.Setup(func(ctx *bubbly.Context) {
		// Expose props for template access and dev tools visibility
		ctx.Expose("count", props.Count)
		ctx.Expose("isEven", props.IsEven)

		// Lifecycle hook - demonstrates component mounting
		ctx.OnMounted(func() {
			fmt.Println("[CounterDisplay] Mounted - visible in dev tools!")
		})
	})

	builder = builder.Template(func(ctx bubbly.RenderContext) string {
		// Get current values from reactive state
		count := ctx.Get("count").(*bubbly.Ref[int]).Get().(int)
		isEven := ctx.Get("isEven").(*bubbly.Computed[interface{}]).Get().(bool)

		// Determine parity text
		parity := "odd"
		if isEven {
			parity = "even"
		}

		// Use BubblyUI Card component (not manual Lipgloss!)
		card := components.Card(components.CardProps{
			Title:   "Counter Display",
			Content: fmt.Sprintf("Count: %d (%s)", count, parity),
		})
		card.Init()

		return card.View()
	})

	return builder.Build()
}
