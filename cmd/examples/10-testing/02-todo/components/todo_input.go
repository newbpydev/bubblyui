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

			// Create Input component using unified pattern
			// Input now uses WithMessageHandler for keyboard input
			// This allows it to work with ExposeComponent like all other components
			inputComp := components.Input(components.InputProps{
				Value:       props.Value,
				Placeholder: "What needs to be done?",
				Width:       50,
				CharLimit:   100,
				NoBorder:    true, // We'll add our own border in the card
			})

			// Use ExposeComponent for automatic Init + parent-child relationship
			// This is now the UNIFIED pattern for ALL components!
			if err := ctx.ExposeComponent("inputComp", inputComp); err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to expose input: %v", err))
				return
			}

			// Forward focus/blur events to the Input component
			// Use internal event names to avoid bubble-back loops
			// (parent emits "setFocus" -> we forward as "focus" to child)
			ctx.On("setFocus", func(data interface{}) {
				comp := ctx.Get("inputComp").(bubbly.Component)
				comp.Emit("focus", nil)
			})

			ctx.On("setBlur", func(data interface{}) {
				comp := ctx.Get("inputComp").(bubbly.Component)
				comp.Emit("blur", nil)
			})

			ctx.On("submit", func(data interface{}) {
				value := props.Value.Get().(string)
				if value != "" && props.OnSubmit != nil {
					props.OnSubmit(value)
					props.Value.Set("")
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			focusedRef := ctx.Get("focused").(*bubbly.Ref[interface{}])
			focusColor := ctx.Get("focusColor").(lipgloss.Color)
			inactiveColor := ctx.Get("inactiveColor").(lipgloss.Color)
			inputComp := ctx.Get("inputComp").(bubbly.Component)

			isFocused := focusedRef.Get().(bool)
			borderColor := inactiveColor
			if isFocused {
				borderColor = focusColor
			}

			// Render the Input component (it has cursor support!)
			inputView := inputComp.View()

			instructColor := inactiveColor
			if isFocused {
				instructColor = lipgloss.Color("250")
			}
			instructions := components.Text(components.TextProps{
				Content: "Press [enter] to add • Press [d] on item to delete",
				Color:   instructColor,
			})
			instructions.Init()

			content := fmt.Sprintf("%s\n\n%s", inputView, instructions.View())

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
