package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// FooterProps defines the props for Footer component
type FooterProps struct {
	Todos *bubbly.Ref[[]Todo]
}

// CreateFooter creates a footer with statistics
// Demonstrates:
// - Computed values from props
// - Using Badge component
// - State derivation
func CreateFooter(props FooterProps) (bubbly.Component, error) {
	builder := bubbly.NewComponent("Footer")

	builder = builder.Setup(func(ctx *bubbly.Context) {
		// Expose todos for dev tools
		ctx.Expose("todos", props.Todos)

		// Computed: Total count
		totalCount := ctx.Computed(func() interface{} {
			todos := props.Todos.Get().([]Todo)
			return len(todos)
		})

		// Computed: Completed count
		completedCount := ctx.Computed(func() interface{} {
			todos := props.Todos.Get().([]Todo)
			count := 0
			for _, todo := range todos {
				if todo.Completed.Get().(bool) {
					count++
				}
			}
			return count
		})

		// Computed: Pending count
		pendingCount := ctx.Computed(func() interface{} {
			todos := props.Todos.Get().([]Todo)
			count := 0
			for _, todo := range todos {
				if !todo.Completed.Get().(bool) {
					count++
				}
			}
			return count
		})

		// Expose computed values for dev tools inspection
		ctx.Expose("totalCount", totalCount)
		ctx.Expose("completedCount", completedCount)
		ctx.Expose("pendingCount", pendingCount)
	})

	builder = builder.Template(func(ctx bubbly.RenderContext) string {
		total := ctx.Get("totalCount").(*bubbly.Computed[interface{}]).Get().(int)
		completed := ctx.Get("completedCount").(*bubbly.Computed[interface{}]).Get().(int)
		pending := ctx.Get("pendingCount").(*bubbly.Computed[interface{}]).Get().(int)

		// Create badges for statistics
		totalBadge := components.Badge(components.BadgeProps{
			Label:   fmt.Sprintf("Total: %d", total),
			Variant: components.VariantInfo,
		})
		totalBadge.Init()

		completedBadge := components.Badge(components.BadgeProps{
			Label:   fmt.Sprintf("Done: %d", completed),
			Variant: components.VariantSuccess,
		})
		completedBadge.Init()

		pendingBadge := components.Badge(components.BadgeProps{
			Label:   fmt.Sprintf("Pending: %d", pending),
			Variant: components.VariantWarning,
		})
		pendingBadge.Init()

		// Layout badges horizontally
		badges := lipgloss.JoinHorizontal(
			lipgloss.Left,
			totalBadge.View(),
			"  ",
			completedBadge.View(),
			"  ",
			pendingBadge.View(),
		)

		// Add help text
		helpStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			MarginTop(1)
		help := helpStyle.Render("Press [F12] to toggle dev tools • Navigate with ↑↓ • Press [space] to toggle • [ctrl+c] to quit")

		// Combine badges and help
		footerStyle := lipgloss.NewStyle().
			MarginTop(1).
			Padding(1).
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(lipgloss.Color("240"))

		content := lipgloss.JoinVertical(
			lipgloss.Left,
			badges,
			help,
		)

		return footerStyle.Render(content)
	})

	return builder.Build()
}
