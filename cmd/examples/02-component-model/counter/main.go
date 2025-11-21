package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// createCounter creates a counter component with advanced state management
func createCounter() (bubbly.Component, error) {
	return bubbly.NewComponent("Counter").
		WithKeyBinding("up", "increment", "Increment counter").
		WithKeyBinding("k", "increment", "Increment counter").
		WithKeyBinding("+", "increment", "Increment counter").
		WithKeyBinding("down", "decrement", "Decrement counter").
		WithKeyBinding("j", "decrement", "Decrement counter").
		WithKeyBinding("-", "decrement", "Decrement counter").
		WithKeyBinding("r", "reset", "Reset to zero").
		WithKeyBinding("d", "double", "Double the count").
		WithKeyBinding("h", "halve", "Halve the count").
		WithKeyBinding("q", "quit", "Quit application").
		Setup(func(ctx *bubbly.Context) {
			// Reactive state
			count := ctx.Ref(0)
			history := ctx.Ref([]int{0})

			// Computed values
			doubled := ctx.Computed(func() interface{} {
				return count.GetTyped().(int) * 2
			})

			squared := ctx.Computed(func() interface{} {
				c := count.GetTyped().(int)
				return c * c
			})

			isEven := ctx.Computed(func() interface{} {
				return count.GetTyped().(int)%2 == 0
			})

			// Expose state and computed values
			ctx.Expose("count", count)
			ctx.Expose("history", history)
			ctx.Expose("doubled", doubled)
			ctx.Expose("squared", squared)
			ctx.Expose("isEven", isEven)

			// Helper to add to history
			addToHistory := func(newVal int) {
				hist := history.GetTyped().([]int)
				// Keep last 5 values
				if len(hist) >= 5 {
					hist = hist[1:]
				}
				hist = append(hist, newVal)
				history.Set(hist)
			}

			// Event handlers
			ctx.On("increment", func(data interface{}) {
				newVal := count.GetTyped().(int) + 1
				count.Set(newVal)
				addToHistory(newVal)
			})

			ctx.On("decrement", func(data interface{}) {
				newVal := count.GetTyped().(int) - 1
				count.Set(newVal)
				addToHistory(newVal)
			})

			ctx.On("reset", func(data interface{}) {
				count.Set(0)
				history.Set([]int{0})
			})

			ctx.On("double", func(data interface{}) {
				newVal := count.GetTyped().(int) * 2
				count.Set(newVal)
				addToHistory(newVal)
			})

			ctx.On("halve", func(data interface{}) {
				newVal := count.GetTyped().(int) / 2
				count.Set(newVal)
				addToHistory(newVal)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Get component for help text
			comp := ctx.Component()

			// Title
			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205")).
				MarginBottom(1)

			title := titleStyle.Render("🔢 Counter Component - State Management")

			// Get state
			count := ctx.Get("count").(*bubbly.Ref[interface{}])
			history := ctx.Get("history").(*bubbly.Ref[interface{}])
			doubled := ctx.Get("doubled").(*bubbly.Computed[interface{}])
			squared := ctx.Get("squared").(*bubbly.Computed[interface{}])
			isEven := ctx.Get("isEven").(*bubbly.Computed[interface{}])

			countVal := count.GetTyped().(int)
			historyVal := history.GetTyped().([]int)

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

			evenStr := "Odd"
			if isEven.GetTyped().(bool) {
				evenStr = "Even"
			}

			computedBox := computedStyle.Render(fmt.Sprintf(
				"Doubled: %d\nSquared: %d\nParity:  %s",
				doubled.GetTyped().(int),
				squared.GetTyped().(int),
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

			// Help text
			helpStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				MarginTop(2)

			help := helpStyle.Render(comp.HelpText())

			return lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				"",
				counterBox,
				"",
				computedBox,
				"",
				historyBox,
				"",
				help,
			)
		}).
		Build()
}

func main() {
	counter, err := createCounter()
	if err != nil {
		fmt.Printf("Error creating counter: %v\n", err)
		os.Exit(1)
	}

	// Run with bubbly.Run() - zero boilerplate!
	// No manual model, no manual key routing, auto-generated help text
	if err := bubbly.Run(counter, bubbly.WithAltScreen()); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
