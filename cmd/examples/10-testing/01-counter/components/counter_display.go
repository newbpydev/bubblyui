package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// CounterDisplayProps defines props for the counter display component
type CounterDisplayProps struct {
	Count   *bubbly.Ref[interface{}]
	Doubled *bubbly.Computed[interface{}]
	IsEven  *bubbly.Computed[interface{}]
	History *bubbly.Ref[interface{}]
}

// CreateCounterDisplay creates a display component for counter state
func CreateCounterDisplay(props CounterDisplayProps) (bubbly.Component, error) {
	return bubbly.NewComponent("CounterDisplay").
		Props(props).
		Template(func(ctx bubbly.RenderContext) string {
			props := ctx.Props().(CounterDisplayProps)

			countVal := props.Count.Get().(int)
			doubledVal := props.Doubled.Get().(int)
			historyVal := props.History.Get().([]int)
			evenStr := "Odd"
			if props.IsEven.Get().(bool) {
				evenStr = "Even"
			}

			// Main counter box
			counterStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("63")).
				Padding(1, 3).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("99")).
				Width(30).
				Align(lipgloss.Center)

			counterBox := counterStyle.Render(fmt.Sprintf("Count: %d", countVal))

			// Computed values box
			computedStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("170")).
				Padding(1, 2).
				Border(lipgloss.DoubleBorder()).
				BorderForeground(lipgloss.Color("141")).
				Width(30)

			computedBox := computedStyle.Render(fmt.Sprintf(
				"Doubled: %d\nSquared: %d\nParity:  %s",
				doubledVal,
				countVal*countVal,
				evenStr,
			))

			// History box
			historyStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Padding(1, 2).
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240")).
				Width(30)

			historyStr := "History: "
			for i, v := range historyVal {
				if i > 0 {
					historyStr += " → "
				}
				historyStr += fmt.Sprintf("%d", v)
			}

			historyBox := historyStyle.Render(historyStr)

			return lipgloss.JoinVertical(
				lipgloss.Left,
				counterBox,
				"",
				computedBox,
				"",
				historyBox,
			)
		}).
		Build()
}
