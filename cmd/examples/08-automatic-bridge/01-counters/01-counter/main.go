package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// createCounter creates a simple counter component demonstrating:
// - Automatic reactive bridge (WithAutoCommands)
// - Declarative key bindings (WithKeyBinding)
// - Auto-generated help text (HelpText)
// - Zero boilerplate with bubbly.Wrap()
func createCounter() (bubbly.Component, error) {
	return bubbly.NewComponent("Counter").
		WithAutoCommands(true).                                // Enable automatic reactive bridge
		WithKeyBinding(" ", "increment", "Increment counter"). // Space key is " " not "space"
		WithKeyBinding("r", "reset", "Reset to zero").
		WithKeyBinding("ctrl+c", "quit", "Quit application").
		Setup(func(ctx *bubbly.Context) {
			// Create reactive state
			count := ctx.Ref(0)

			// Expose to template
			ctx.Expose("count", count)

			// Event handlers - no manual Emit() needed!
			// State changes automatically trigger UI updates
			ctx.On("increment", func(_ interface{}) {
				current := count.Get().(int)
				count.Set(current + 1)
				// UI updates automatically - no Emit() needed!
			})

			ctx.On("reset", func(_ interface{}) {
				count.Set(0)
				// UI updates automatically - no Emit() needed!
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Get state
			count := ctx.Get("count").(*bubbly.Ref[interface{}])
			currentCount := count.Get().(int)

			// Get component for help text
			comp := ctx.Component()

			// Styling
			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205")).
				MarginBottom(1)

			counterStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("63")).
				Padding(2, 4).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("99")).
				Width(40).
				Align(lipgloss.Center)

			infoStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Padding(1, 2).
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240")).
				Width(40)

			helpStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				MarginTop(1)

			// Render
			title := titleStyle.Render("🎯 Automatic Bridge - Simple Counter")

			counterBox := counterStyle.Render(fmt.Sprintf("Count: %d", currentCount))

			infoBox := infoStyle.Render(
				"Features:\n" +
					"✓ Automatic reactive bridge\n" +
					"✓ Declarative key bindings\n" +
					"✓ Auto-generated help text\n" +
					"✓ Zero boilerplate code",
			)

			// Auto-generated help text from key bindings
			help := helpStyle.Render(comp.HelpText())

			return lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				"",
				counterBox,
				"",
				infoBox,
				"",
				help,
			)
		}).
		Build()
}

func main() {
	component, err := createCounter()
	if err != nil {
		fmt.Printf("Error creating component: %v\n", err)
		os.Exit(1)
	}

	// Run with bubbly.Run() - ultimate zero boilerplate!
	// No Bubbletea imports, no manual model, no Wrap() call needed
	if err := bubbly.Run(component, bubbly.WithAltScreen()); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
