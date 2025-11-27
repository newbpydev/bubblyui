package components

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// Todo represents a todo item data structure
type Todo struct {
	ID        int
	Text      string
	Completed *bubbly.Ref[bool]
}

// TodoListProps defines the props for TodoList component
type TodoListProps struct {
	Todos         *bubbly.Ref[[]Todo]
	SelectedIndex *bubbly.Ref[int]
}

// CreateTodoList creates a list of todo items
// Demonstrates:
// - Parent component composing children
// - Dynamic child creation
// - State passed to children via props
// - ExposeComponent for child registration
func CreateTodoList(props TodoListProps) (bubbly.Component, error) {
	builder := bubbly.NewComponent("TodoList")

	builder = builder.Setup(func(ctx *bubbly.Context) {
		// Expose props for dev tools
		ctx.Expose("todos", props.Todos)
		ctx.Expose("selectedIndex", props.SelectedIndex)

		// Get current todos
		todos := props.Todos.Get().([]Todo)
		selectedIdx := props.SelectedIndex.Get().(int)

		// Create TodoItem components dynamically
		for i, todo := range todos {
			item, err := CreateTodoItem(TodoItemProps{
				ID:         todo.ID,
				Text:       todo.Text,
				Completed:  todo.Completed,
				IsSelected: i == selectedIdx,
			})
			if err != nil {
				continue
			}

			// ExposeComponent registers child in component tree
			ctx.ExposeComponent(todo.Text, item)
		}
	})

	builder = builder.Template(func(ctx bubbly.RenderContext) string {
		todos := ctx.Get("todos").(*bubbly.Ref[[]Todo]).Get().([]Todo)

		if len(todos) == 0 {
			emptyStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Italic(true).
				Padding(1)
			return emptyStyle.Render("No todos yet. This is a demo for component inspection!")
		}

		// Render all child components
		var items []string
		for _, todo := range todos {
			if comp := ctx.Get(todo.Text); comp != nil {
				if component, ok := comp.(bubbly.Component); ok {
					items = append(items, component.View())
				}
			}
		}

		// Join items vertically
		listStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1)

		return listStyle.Render(lipgloss.JoinVertical(lipgloss.Left, items...))
	})

	return builder.Build()
}
