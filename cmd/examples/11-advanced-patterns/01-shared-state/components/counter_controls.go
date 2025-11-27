package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	localComposables "github.com/newbpydev/bubblyui/cmd/examples/11-advanced-patterns/01-shared-state/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// CounterControlsProps defines props for counter controls component
type CounterControlsProps struct {
	OnIncrement func()
	OnDecrement func()
	OnReset     func()
}

// CreateCounterControls creates a component with buttons to control the shared counter
func CreateCounterControls(props CounterControlsProps) (bubbly.Component, error) {
	return bubbly.NewComponent("CounterControls").
		Props(props).
		Setup(func(ctx *bubbly.Context) {
			// Get shared counter instance (same as Display component!)
			counter := localComposables.UseSharedCounter(ctx)
			ctx.Expose("counter", counter)

			// Event handlers - modify shared state
			ctx.On("increment", func(data interface{}) {
				if props.OnIncrement != nil {
					props.OnIncrement()
				}
			})

			ctx.On("decrement", func(data interface{}) {
				if props.OnDecrement != nil {
					props.OnDecrement()
				}
			})

			ctx.On("reset", func(data interface{}) {
				if props.OnReset != nil {
					props.OnReset()
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			counter := ctx.Get("counter").(*localComposables.CounterComposable)

			// Get current count for display
			count := counter.Count.Get().(int)

			// Build controls
			titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
			buttonStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("35")).Bold(true)
			helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)

			content := fmt.Sprintf("%s\n\n", titleStyle.Render("🎮 Counter Controls"))
			content += fmt.Sprintf("Current: %s\n\n", buttonStyle.Render(fmt.Sprintf("%d", count)))
			content += helpStyle.Render("↑/k/+: increment\n")
			content += helpStyle.Render("↓/j/-: decrement\n")
			content += helpStyle.Render("r: reset")

			// Use Card component
			card := components.Card(components.CardProps{
				Title:   "",
				Content: content,
				Padding: 1,
				Width:   40,
			})
			card.Init()
			return card.View()
		}).
		Build()
}
