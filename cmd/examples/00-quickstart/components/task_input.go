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
	InputMode *bubblyui.Ref[bool] // Whether input mode is active
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
			isInputMode := props.InputMode != nil && props.InputMode.GetTyped()

			// Build input display
			var content string
			var title string

			if isInputMode {
				// Active input mode - show blinking cursor effect
				title = "✏️  New Task (typing...)"
				if inputText == "" {
					content = "│"
				} else {
					content = inputText + "│"
				}
			} else if props.IsFocused() {
				// Focused but not in input mode
				title = "New Task"
				content = inputText
				if inputText == "" {
					content = "(Press 'a' to add a task)"
				}
			} else {
				// Not focused
				title = "New Task"
				content = inputText
				if inputText == "" {
					content = "(Press 'a' to add a task)"
				}
			}

			// Style the input content
			inputStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))
			if isInputMode {
				// Highlight when typing
				inputStyle = inputStyle.
					Foreground(lipgloss.Color("229")). // Yellow for active input
					Bold(true)
			} else if !props.IsFocused() && inputText == "" {
				inputStyle = inputStyle.Foreground(lipgloss.Color("240"))
			}

			styledContent := inputStyle.Render(content)

			// Add hint when in input mode
			if isInputMode {
				hintStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("240")).
					Italic(true)
				styledContent += "\n" + hintStyle.Render("Enter: submit | Esc: cancel")
			}

			// Wrap in a Card component
			card := components.Card(components.CardProps{
				Title:   title,
				Content: styledContent,
			})
			card.Init() // REQUIRED before View()!

			return card.View()
		}).
		Build()
}
