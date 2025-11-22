package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// FormFieldProps defines props for a form field component
type FormFieldProps struct {
	Label       string
	Value       *bubbly.Ref[string]
	Placeholder string
	Focused     *bubbly.Ref[interface{}] // Is this field focused?
	Error       *bubbly.Ref[interface{}] // Validation error ref
	Width       int
}

// CreateFormField creates a reusable form field component
// Wraps Input component like TodoInput wraps Input
func CreateFormField(props FormFieldProps) (bubbly.Component, error) {
	return bubbly.NewComponent("FormField").
		Setup(func(ctx *bubbly.Context) {
			ctx.Expose("label", props.Label)
			ctx.Expose("focused", props.Focused)
			ctx.Expose("error", props.Error)

			// INJECT theme colors from parent
			focusColor := lipgloss.Color("35")     // Green
			inactiveColor := lipgloss.Color("240") // Dark grey
			errorColor := lipgloss.Color("196")    // Red

			if injected := ctx.Inject("focusColor", nil); injected != nil {
				focusColor = injected.(lipgloss.Color)
			}
			if injected := ctx.Inject("inactiveColor", nil); injected != nil {
				inactiveColor = injected.(lipgloss.Color)
			}
			if injected := ctx.Inject("errorColor", nil); injected != nil {
				errorColor = injected.(lipgloss.Color)
			}

			ctx.Expose("focusColor", focusColor)
			ctx.Expose("inactiveColor", inactiveColor)
			ctx.Expose("errorColor", errorColor)

			// Create Input component with proper width
			width := props.Width
			if width == 0 {
				width = 40
			}

			inputComp := components.Input(components.InputProps{
				Value:       props.Value,
				Placeholder: props.Placeholder,
				Width:       width,
				CharLimit:   100,
				NoBorder:    true,
			})

			// Use ExposeComponent for unified pattern (like TodoInput does)
			if err := ctx.ExposeComponent("inputComp", inputComp); err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to expose input: %v", err))
				return
			}

			// Forward focus/blur events to Input component
			// (parent emits "setFocus" -> we forward as "focus")
			ctx.On("setFocus", func(data interface{}) {
				comp := ctx.Get("inputComp").(bubbly.Component)
				comp.Emit("focus", nil)
			})

			ctx.On("setBlur", func(data interface{}) {
				comp := ctx.Get("inputComp").(bubbly.Component)
				comp.Emit("blur", nil)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			label := ctx.Get("label").(string)
			focusedRef := ctx.Get("focused").(*bubbly.Ref[interface{}])
			errorRef := ctx.Get("error").(*bubbly.Ref[interface{}])
			focusColor := ctx.Get("focusColor").(lipgloss.Color)
			inactiveColor := ctx.Get("inactiveColor").(lipgloss.Color)
			errorColor := ctx.Get("errorColor").(lipgloss.Color)
			inputComp := ctx.Get("inputComp").(bubbly.Component)

			isFocused := focusedRef.Get().(bool)
			errorMsg := errorRef.Get().(string)

			// Label style
			labelColor := inactiveColor
			if isFocused {
				labelColor = focusColor
			}
			if errorMsg != "" {
				labelColor = errorColor
			}

			labelStyle := lipgloss.NewStyle().
				Foreground(labelColor).
				Bold(true).
				Width(18)

			labelText := labelStyle.Render(label + ":")

			// Input border
			borderColor := inactiveColor
			if isFocused {
				borderColor = focusColor
			}
			if errorMsg != "" {
				borderColor = errorColor
			}

			inputBorder := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(borderColor).
				Padding(0, 1)

			inputView := inputBorder.Render(inputComp.View())

			// Error message
			segments := []string{labelText, inputView}
			errorView := ""
			if errorMsg != "" {
				errorStyle := lipgloss.NewStyle().
					Foreground(errorColor).
					PaddingLeft(2)
				errorView = errorStyle.Render("⚠ " + errorMsg)
				segments = append(segments, errorView)
			}

			return lipgloss.JoinVertical(lipgloss.Left, segments...)
		}).
		Build()
}
