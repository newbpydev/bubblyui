// Package components provides UI components for the quickstart example.
package components

import (
	"github.com/charmbracelet/lipgloss"

	// Clean import paths using alias packages
	"github.com/newbpydev/bubblyui"
	"github.com/newbpydev/bubblyui/components"

	// Need pkg/bubbly for Context/RenderContext (builder callback types)
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TaskInputProps defines the props for TaskInput component.
type TaskInputProps struct {
	InputText *bubblyui.Ref[string]
	IsFocused func() bool
	OnSubmit  func(text string)
}

// CreateTaskInput creates a component for adding new tasks.
// This demonstrates:
// - Text input handling with reactive state
// - Using BubblyUI Card component
// - Event-based communication with parent
func CreateTaskInput(props TaskInputProps) (bubbly.Component, error) {
	return bubbly.NewComponent("TaskInput").
		Setup(func(ctx *bubbly.Context) {
			// Expose props for template access and DevTools visibility
			ctx.Expose("inputText", props.InputText)

			// Handle submit event
			ctx.On("submit", func(_ interface{}) {
				// Use GetTyped() for type-safe access
				text := props.InputText.GetTyped()
				if text != "" && props.OnSubmit != nil {
					props.OnSubmit(text)
					props.InputText.Set("") // Clear input after submit
				}
			})

			// Lifecycle hook
			ctx.OnMounted(func() {
				// TaskInput mounted - visible in DevTools
			})
		}).
		// Template receives RenderContext (no pointer!)
		Template(func(_ bubbly.RenderContext) string {
			// Use GetTyped() for type-safe access
			inputText := props.InputText.GetTyped()

			// Build input display
			var content string
			if props.IsFocused() {
				// Show cursor when focused
				content = inputText + "_"
				if inputText == "" {
					content = "Type task and press Enter..._"
				}
			} else {
				content = inputText
				if inputText == "" {
					content = "(Press Tab to focus input)"
				}
			}

			// Style the input content
			inputStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))
			if !props.IsFocused() && inputText == "" {
				inputStyle = inputStyle.Foreground(lipgloss.Color("240"))
			}

			styledContent := inputStyle.Render(content)

			// Wrap in a Card component (simple props - no BorderStyle/BorderColor)
			card := components.Card(components.CardProps{
				Title:   "New Task",
				Content: styledContent,
			})
			card.Init() // REQUIRED before View()!

			return card.View()
		}).
		Build()
}
