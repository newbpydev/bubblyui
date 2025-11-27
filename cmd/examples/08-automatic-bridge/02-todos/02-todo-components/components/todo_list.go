package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// Todo represents a single todo item
type Todo struct {
	ID          int
	Title       string
	Description string
	Priority    string // "low", "medium", "high"
	Completed   bool
}

// TodoListProps defines the props for the TodoList component
type TodoListProps struct {
	Todos         *bubbly.Ref[interface{}]
	SelectedIndex *bubbly.Ref[interface{}]
	InputMode     *bubbly.Ref[interface{}]
	OnToggle      func(index int)
	OnSelect      func(index int)
}

// CreateTodoList creates a reusable todo list component
func CreateTodoList(props TodoListProps) (bubbly.Component, error) {
	return bubbly.NewComponent("TodoList").
		Setup(func(ctx *bubbly.Context) {
			// Event: Toggle todo completion
			ctx.On("toggleTodo", func(data interface{}) {
				index := data.(int)
				if props.OnToggle != nil {
					props.OnToggle(index)
				}
			})

			// Event: Select previous todo
			ctx.On("selectPrevious", func(_ interface{}) {
				currentIndex := props.SelectedIndex.Get().(int)
				if currentIndex > 0 {
					props.SelectedIndex.Set(currentIndex - 1)
					if props.OnSelect != nil {
						props.OnSelect(currentIndex - 1)
					}
				}
			})

			// Event: Select next todo
			ctx.On("selectNext", func(_ interface{}) {
				todoList := props.Todos.Get().([]Todo)
				currentIndex := props.SelectedIndex.Get().(int)
				if currentIndex < len(todoList)-1 {
					props.SelectedIndex.Set(currentIndex + 1)
					if props.OnSelect != nil {
						props.OnSelect(currentIndex + 1)
					}
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			todos := props.Todos.Get().([]Todo)
			selectedIndex := props.SelectedIndex.Get().(int)
			inputMode := props.InputMode.Get().(bool)

			// Todo list - dynamic border color based on mode
			todoBorderColor := "99" // Purple (navigation mode - active)
			if inputMode {
				todoBorderColor = "240" // Dark grey (input mode - inactive)
			}
			todoStyle := lipgloss.NewStyle().
				Padding(1, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(todoBorderColor)).
				Width(70)

			var todoItems []string
			if len(todos) == 0 {
				todoItems = append(todoItems, "No todos yet. Press Ctrl+N to create one!")
			} else {
				for i, todo := range todos {
					cursor := "  "
					if i == selectedIndex {
						cursor = "▶ "
					}

					checkbox := "☐"
					if todo.Completed {
						checkbox = "☑"
					}

					priorityIcon := GetPriorityIcon(todo.Priority)

					titleText := todo.Title
					if todo.Completed {
						titleText = lipgloss.NewStyle().
							Foreground(lipgloss.Color("240")).
							Strikethrough(true).
							Render(titleText)
					}

					todoItems = append(todoItems, fmt.Sprintf(
						"%s%s %s %s",
						cursor, checkbox, priorityIcon, titleText,
					))
				}
			}

			return todoStyle.Render(strings.Join(todoItems, "\n"))
		}).
		Build()
}

// GetPriorityIcon returns the icon for a given priority
func GetPriorityIcon(priority string) string {
	switch priority {
	case "high":
		return "🔴"
	case "medium":
		return "🟡"
	case "low":
		return "🟢"
	default:
		return "⚪"
	}
}
