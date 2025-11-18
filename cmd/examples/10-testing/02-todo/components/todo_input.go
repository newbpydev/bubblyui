package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// TodoInputProps defines the props for the TodoInput component
type TodoInputProps struct {
	Value    *bubbly.Ref[string]
	Focused  *bubbly.Ref[interface{}] // Focused state
	OnSubmit func(title string)
}

// CreateTodoInput creates a new TodoInput component
// Allows user to enter new todo items
func CreateTodoInput(props TodoInputProps) (bubbly.Component, error) {
	return bubbly.NewComponent("TodoInput").
		Setup(func(ctx *bubbly.Context) {
			ctx.Expose("value", props.Value)
			ctx.Expose("focused", props.Focused)
			ctx.Expose("onSubmit", props.OnSubmit)

			// INJECT colors from parent
			focusColor := lipgloss.Color("35")
			inactiveColor := lipgloss.Color("240")
			if injected := ctx.Inject("focusColor", nil); injected != nil {
				focusColor = injected.(lipgloss.Color)
			}
			if injected := ctx.Inject("inactiveColor", nil); injected != nil {
				inactiveColor = injected.(lipgloss.Color)
			}
			ctx.Expose("focusColor", focusColor)
			ctx.Expose("inactiveColor", inactiveColor)

			ctx.On("submit", func(data interface{}) {
				value := props.Value.Get().(string)
				if value != "" && props.OnSubmit != nil {
					props.OnSubmit(value)
					props.Value.Set("")
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			valueRef := ctx.Get("value").(*bubbly.Ref[string])
			focusedRef := ctx.Get("focused").(*bubbly.Ref[interface{}])
			focusColor := ctx.Get("focusColor").(lipgloss.Color)
			inactiveColor := ctx.Get("inactiveColor").(lipgloss.Color)

			isFocused := focusedRef.Get().(bool)
			borderColor := inactiveColor
			if isFocused {
				borderColor = focusColor
			}

			input := components.Input(components.InputProps{
				Value:       valueRef,
				Placeholder: "What needs to be done?",
				Width:       50,
				CharLimit:   100,
			})
			input.Init()

			instructColor := inactiveColor
			if isFocused {
				instructColor = lipgloss.Color("250")
			}
			instructions := components.Text(components.TextProps{
				Content: "Press [enter] to add • Press [d] on item to delete",
				Color:   instructColor,
			})
			instructions.Init()

			content := fmt.Sprintf("%s\n\n%s", input.View(), instructions.View())

			cardStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(borderColor).
				Padding(1)

			title := lipgloss.NewStyle().
				Bold(true).
				Foreground(borderColor).
				Render("Add New Todo")

			return cardStyle.Render(fmt.Sprintf("%s\n\n%s", title, content))
		}).
		Build()
}
