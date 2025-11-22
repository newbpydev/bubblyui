package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TodoStatsProps defines the props for the TodoStats component
type TodoStatsProps struct {
	Todos *bubbly.Ref[interface{}]
}

// CreateTodoStats creates a reusable statistics component
func CreateTodoStats(props TodoStatsProps) (bubbly.Component, error) {
	return bubbly.NewComponent("TodoStats").
		Setup(func(ctx *bubbly.Context) {
			// Computed: Total count
			totalCount := ctx.Computed(func() interface{} {
				todos := props.Todos.Get().([]Todo)
				return len(todos)
			})
			ctx.Expose("totalCount", totalCount)

			// Computed: Completed count
			completedCount := ctx.Computed(func() interface{} {
				todos := props.Todos.Get().([]Todo)
				count := 0
				for _, todo := range todos {
					if todo.Completed {
						count++
					}
				}
				return count
			})
			ctx.Expose("completedCount", completedCount)

			// Computed: Pending count
			pendingCount := ctx.Computed(func() interface{} {
				total := totalCount.Get().(int)
				completed := completedCount.Get().(int)
				return total - completed
			})
			ctx.Expose("pendingCount", pendingCount)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Get computed values (they're already pointers from ctx.Computed)
			totalCount := ctx.Get("totalCount").(*bubbly.Computed[interface{}])
			completedCount := ctx.Get("completedCount").(*bubbly.Computed[interface{}])
			pendingCount := ctx.Get("pendingCount").(*bubbly.Computed[interface{}])

			statsStyle := lipgloss.NewStyle().
				Bold(true).
				Padding(0, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("99")).
				Width(70)

			return statsStyle.Render(fmt.Sprintf(
				"📊 Total: %d | ✅ Completed: %d | ⏳ Pending: %d",
				totalCount.Get().(int),
				completedCount.Get().(int),
				pendingCount.Get().(int),
			))
		}).
		Build()
}
