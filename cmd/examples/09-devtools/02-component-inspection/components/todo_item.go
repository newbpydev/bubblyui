package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// TodoItemProps defines the props for TodoItem component
type TodoItemProps struct {
	ID         int
	Text       string
	Completed  *bubbly.Ref[bool]
	IsSelected bool
}

// CreateTodoItem creates a todo item component
// Demonstrates:
// - Leaf component in hierarchy
// - State management with Ref
// - Conditional styling based on props
// - Using Checkbox component
func CreateTodoItem(props TodoItemProps) (bubbly.Component, error) {
	builder := bubbly.NewComponent(fmt.Sprintf("TodoItem#%d", props.ID))

	builder = builder.Setup(func(ctx *bubbly.Context) {
		// Expose props for dev tools inspection
		ctx.Expose("id", props.ID)
		ctx.Expose("text", props.Text)
		ctx.Expose("completed", props.Completed)
		ctx.Expose("isSelected", props.IsSelected)
	})

	builder = builder.Template(func(ctx bubbly.RenderContext) string {
		id := ctx.Get("id").(int)
		text := ctx.Get("text").(string)
		completed := ctx.Get("completed").(*bubbly.Ref[bool])
		isSelected := ctx.Get("isSelected").(bool)

		// Get completed state
		isDone := completed.Get().(bool)

		// Use Checkbox component
		checkbox := components.Checkbox(components.CheckboxProps{
			Label:   text,
			Checked: completed,
		})
		checkbox.Init()

		// Style based on state
		itemStyle := lipgloss.NewStyle().
			Padding(0, 1)

		if isSelected {
			// Highlight selected item
			itemStyle = itemStyle.
				Background(lipgloss.Color("99")).
				Foreground(lipgloss.Color("230"))
		} else if isDone {
			// Dim completed items
			itemStyle = itemStyle.
				Foreground(lipgloss.Color("240"))
		}

		content := fmt.Sprintf("[%d] %s", id, checkbox.View())
		return itemStyle.Render(content)
	})

	return builder.Build()
}
