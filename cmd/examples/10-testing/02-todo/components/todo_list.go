package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/cmd/examples/10-testing/02-todo/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// TodoListProps defines the props for the TodoList component
type TodoListProps struct {
	Todos    *bubbly.Ref[interface{}]
	Focused  *bubbly.Ref[interface{}] // Focused state
	OnToggle func(id int64)
	OnRemove func(id int64)
}

// CreateTodoList creates a new TodoList component
// Displays a list of todos with toggle and remove actions
func CreateTodoList(props TodoListProps) (bubbly.Component, error) {
	return bubbly.NewComponent("TodoList").
		Setup(func(ctx *bubbly.Context) {
			ctx.Expose("todos", props.Todos)
			ctx.Expose("focused", props.Focused)
			ctx.Expose("onToggle", props.OnToggle)
			ctx.Expose("onRemove", props.OnRemove)

			// INJECT colors from parent
			navColor := lipgloss.Color("99")
			inactiveColor := lipgloss.Color("240")
			if injected := ctx.Inject("inactiveColor", nil); injected != nil {
				inactiveColor = injected.(lipgloss.Color)
			}
			ctx.Expose("navColor", navColor)
			ctx.Expose("inactiveColor", inactiveColor)

			ctx.On("toggle", func(data interface{}) {
				if id, ok := data.(int64); ok && props.OnToggle != nil {
					props.OnToggle(id)
				}
			})

			ctx.On("remove", func(data interface{}) {
				if id, ok := data.(int64); ok && props.OnRemove != nil {
					props.OnRemove(id)
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			todosRef := ctx.Get("todos").(*bubbly.Ref[interface{}])
			focusedRef := ctx.Get("focused").(*bubbly.Ref[interface{}])
			navColor := ctx.Get("navColor").(lipgloss.Color)
			inactiveColor := ctx.Get("inactiveColor").(lipgloss.Color)

			todos := todosRef.Get().([]composables.Todo)
			isInInputMode := focusedRef.Get().(bool)
			isListFocused := !isInInputMode

			borderColor := inactiveColor
			if isListFocused {
				borderColor = navColor
			}

			if len(todos) == 0 {
				emptyText := components.Text(components.TextProps{
					Content: "No todos yet. Add one above!",
					Color:   inactiveColor,
				})
				emptyText.Init()

				emptyCard := lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(borderColor).
					Padding(1).
					Render(fmt.Sprintf("%s\n\n%s",
						lipgloss.NewStyle().Bold(true).Foreground(borderColor).Render("Todo List"),
						emptyText.View()))

				return emptyCard
			}

			var items []string
			for i, todo := range todos {
				var status string
				if todo.Completed {
					status = "✓"
				} else {
					status = "○"
				}

				titleStyle := lipgloss.NewStyle()
				if todo.Completed {
					titleStyle = titleStyle.Foreground(inactiveColor).Strikethrough(true)
				} else {
					if isListFocused {
						titleStyle = titleStyle.Foreground(lipgloss.Color("252"))
					} else {
						titleStyle = titleStyle.Foreground(inactiveColor)
					}
				}

				itemText := fmt.Sprintf("[%s] %d. %s", status, i+1, todo.Title)

				deleteColor := inactiveColor
				if isListFocused {
					deleteColor = lipgloss.Color("160")
				}
				item := titleStyle.Render(itemText) + lipgloss.NewStyle().
					Foreground(deleteColor).
					Render(" [x]")

				items = append(items, item)
			}

			content := strings.Join(items, "\n")

			cardStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(borderColor).
				Padding(1)

			title := lipgloss.NewStyle().
				Bold(true).
				Foreground(borderColor).
				Render("Todo List")

			return cardStyle.Render(fmt.Sprintf("%s\n\n%s", title, content))
		}).
		Build()
}
