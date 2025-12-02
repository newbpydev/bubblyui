// Package components provides UI components for the quickstart example.
package components

import (
	"github.com/charmbracelet/lipgloss"

	// Clean import paths using alias packages
	"github.com/newbpydev/bubblyui/components"

	// Need pkg/bubbly for Context/RenderContext (builder callback types)
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// HelpPanelProps defines the props for HelpPanel component.
type HelpPanelProps struct {
	// No props needed for static help content
}

// CreateHelpPanel creates a component that displays keyboard shortcuts.
// This demonstrates:
// - Simple static component pattern
// - Using BubblyUI Text component
// - Clean presentation of help information
func CreateHelpPanel(_ HelpPanelProps) (bubbly.Component, error) {
	return bubbly.NewComponent("HelpPanel").
		Setup(func(ctx *bubbly.Context) {
			// Lifecycle hook
			ctx.OnMounted(func() {
				// HelpPanel mounted - visible in DevTools
			})
		}).
		// Template receives RenderContext (no pointer!)
		Template(func(_ bubbly.RenderContext) string {
			// Build help content
			keyStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("99")).
				Bold(true)
			descStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("240"))
			sepStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("238"))

			separator := sepStyle.Render(" | ")

			helpText := keyStyle.Render("Tab") + descStyle.Render(":Focus") + separator +
				keyStyle.Render("j/k") + descStyle.Render(":Navigate") + separator +
				keyStyle.Render("Enter") + descStyle.Render(":Toggle") + separator +
				keyStyle.Render("a") + descStyle.Render(":Add") + separator +
				keyStyle.Render("d") + descStyle.Render(":Delete") + separator +
				keyStyle.Render("f") + descStyle.Render(":Filter") + separator +
				keyStyle.Render("c") + descStyle.Render(":Clear Done") + separator +
				keyStyle.Render("F12") + descStyle.Render(":DevTools") + separator +
				keyStyle.Render("q") + descStyle.Render(":Quit")

			text := components.Text(components.TextProps{
				Content: helpText,
			})
			text.Init() // REQUIRED before View()!

			return text.View()
		}).
		Build()
}
