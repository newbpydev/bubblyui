package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	localComposables "github.com/newbpydev/bubblyui/cmd/examples/11-advanced-patterns/01-shared-state/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// CounterDisplayProps defines props for counter display component
type CounterDisplayProps struct {
	Count   *bubbly.Ref[interface{}]
	Doubled *bubbly.Computed[interface{}]
	IsEven  *bubbly.Computed[interface{}]
	History *bubbly.Ref[interface{}]
}

// CreateCounterDisplay creates a component that displays the shared counter state
func CreateCounterDisplay(props CounterDisplayProps) (bubbly.Component, error) {
	return bubbly.NewComponent("CounterDisplay").
		Props(props).
		Setup(func(ctx *bubbly.Context) {
			// Get shared counter instance
			counter := localComposables.UseSharedCounter(ctx)

			// Expose for template
			ctx.Expose("counter", counter)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			counter := ctx.Get("counter").(*localComposables.CounterComposable)

			// Get current values
			count := counter.Count.Get().(int)
			doubled := counter.Doubled.Get().(int)
			isEven := counter.IsEven.Get().(bool)
			history := counter.History.Get().([]int)

			// Build content
			var content strings.Builder
			content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("35")).Render("📊 Counter Display"))
			content.WriteString("\n\n")
			content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Current Value:"))
			content.WriteString("\n")
			content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("35")).Width(20).Align(lipgloss.Center).Render(fmt.Sprintf("%d", count)))
			content.WriteString("\n\n")
			content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Doubled: "))
			content.WriteString(fmt.Sprintf("%d\n", doubled))
			content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Is Even: "))
			if isEven {
				content.WriteString("✓ Yes\n")
			} else {
				content.WriteString("✗ No\n")
			}

			// History
			content.WriteString("\n")
			content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("History: "))
			histStr := make([]string, len(history))
			for i, v := range history {
				histStr[i] = fmt.Sprintf("%d", v)
			}
			content.WriteString(strings.Join(histStr, " → "))

			// Use Card component
			card := components.Card(components.CardProps{
				Title:   "",
				Content: content.String(),
				Padding: 1,
				Width:   40,
			})
			card.Init()
			return card.View()
		}).
		Build()
}
