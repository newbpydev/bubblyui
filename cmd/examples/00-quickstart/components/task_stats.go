// Package components provides UI components for the quickstart example.
package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	// Clean import paths using alias packages
	"github.com/newbpydev/bubblyui/components"

	// Need pkg/bubbly for Context/RenderContext (builder callback types)
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TaskStatsProps defines the props for TaskStats component.
type TaskStatsProps struct {
	ActiveCount func() int
	DoneCount   func() int
	TotalCount  func() int
}

// CreateTaskStats creates a component that displays task statistics.
// This demonstrates:
// - Computed-like pattern with function props
// - Using BubblyUI components for styling
// - Clean separation of data and presentation
func CreateTaskStats(props TaskStatsProps) (bubbly.Component, error) {
	return bubbly.NewComponent("TaskStats").
		Setup(func(ctx *bubbly.Context) {
			// Lifecycle hook
			ctx.OnMounted(func() {
				// TaskStats mounted - visible in DevTools
			})
		}).
		// Template receives RenderContext (no pointer!)
		Template(func(_ bubbly.RenderContext) string {
			active := props.ActiveCount()
			done := props.DoneCount()
			total := props.TotalCount()

			// Calculate completion percentage
			percentage := 0
			if total > 0 {
				percentage = (done * 100) / total
			}

			// Build stats content
			activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
			doneStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("34"))
			totalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
			percentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("141"))

			content := fmt.Sprintf(
				"%s: %d  |  %s: %d  |  %s: %d  |  %s: %d%%",
				activeStyle.Render("Active"),
				active,
				doneStyle.Render("Done"),
				done,
				totalStyle.Render("Total"),
				total,
				percentStyle.Render("Complete"),
				percentage,
			)

			// Create a simple text component
			text := components.Text(components.TextProps{
				Content: content,
			})
			text.Init() // REQUIRED before View()!

			return text.View()
		}).
		Build()
}
