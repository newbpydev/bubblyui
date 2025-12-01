// Package components provides UI components for the quickstart example.
package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	// Clean import paths using alias packages
	"github.com/newbpydev/bubblyui"
	"github.com/newbpydev/bubblyui/components"

	// Need pkg/bubbly for Context/RenderContext (builder callback types)
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TaskStatsProps defines the props for TaskStats component.
type TaskStatsProps struct {
	ActiveCount func() int
	DoneCount   func() int
	TotalCount  func() int
	Filter      *bubblyui.Ref[string] // Current filter: "all", "active", "done"
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

			// Get current filter
			currentFilter := "all"
			if props.Filter != nil {
				currentFilter = props.Filter.GetTyped()
			}

			// Calculate completion percentage
			percentage := 0
			if total > 0 {
				percentage = (done * 100) / total
			}

			// =============================================================================
			// Filter Chips - Shows which filter is active with visual highlighting
			// =============================================================================
			filters := []string{"all", "active", "done"}
			var filterChips []string

			for _, f := range filters {
				var chipStyle lipgloss.Style
				if f == currentFilter {
					// Active filter - highlighted
					chipStyle = lipgloss.NewStyle().
						Background(lipgloss.Color("99")).
						Foreground(lipgloss.Color("0")).
						Bold(true).
						Padding(0, 1)
				} else {
					// Inactive filter - muted
					chipStyle = lipgloss.NewStyle().
						Foreground(lipgloss.Color("240")).
						Padding(0, 1)
				}
				filterChips = append(filterChips, chipStyle.Render(strings.ToUpper(f)))
			}

			filterBar := lipgloss.JoinHorizontal(lipgloss.Center, filterChips...)

			// =============================================================================
			// Stats Line
			// =============================================================================
			activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
			doneStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("34"))
			totalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
			percentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("141"))

			statsLine := fmt.Sprintf(
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

			// Combine filter bar and stats
			filterLabel := lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Render("Filter (f): ")

			content := lipgloss.JoinVertical(
				lipgloss.Left,
				filterLabel+filterBar,
				"",
				statsLine,
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
