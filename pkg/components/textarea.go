package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TextAreaProps defines the configuration properties for a TextArea component.
//
// Example usage:
//
//	valueRef := bubbly.NewRef("Enter your\nmulti-line\ntext here")
//	textarea := components.TextArea(components.TextAreaProps{
//	    Value:       valueRef,
//	    Placeholder: "Type something...",
//	    Rows:        5,
//	    OnChange: func(value string) {
//	        fmt.Println("Text changed:", value)
//	    },
//	})
type TextAreaProps struct {
	// Value is the reactive reference to the textarea's text content.
	// Required - must be a valid Ref[string].
	// Changes to this ref will update the textarea display.
	// Supports multi-line text with \n for line breaks.
	Value *bubbly.Ref[string]

	// Placeholder is the text displayed when the textarea is empty.
	// Optional - if empty, no placeholder is shown.
	Placeholder string

	// Rows is the height of the textarea in lines.
	// Default: 3 if not specified or <= 0.
	Rows int

	// MaxLength is the maximum number of characters allowed.
	// Optional - if 0, no limit is enforced.
	MaxLength int

	// OnChange is a callback function executed when the text changes.
	// Receives the new text value as a parameter.
	// Optional - if nil, no callback is executed.
	OnChange func(string)

	// Validate is a function to validate the current text.
	// Returns an error if validation fails, nil if valid.
	// Optional - if nil, no validation is performed.
	Validate func(string) error

	// Disabled indicates whether the textarea is disabled.
	// Disabled textareas do not respond to events and are styled differently.
	// Default: false (enabled).
	Disabled bool

	// Common props for all components
	CommonProps
}

// TextArea creates a new TextArea molecule component.
//
// TextArea is a multi-line text input element that allows users to enter longer text content.
// It supports reactive state binding, validation, and callbacks.
//
// The textarea automatically integrates with the theme system via the composition API's
// Provide/Inject mechanism. If no theme is provided, it uses DefaultTheme.
//
// Example:
//
//	valueRef := bubbly.NewRef("")
//	textarea := components.TextArea(components.TextAreaProps{
//	    Value:       valueRef,
//	    Placeholder: "Enter your comment...",
//	    Rows:        5,
//	    MaxLength:   500,
//	    Validate: func(text string) error {
//	        if len(text) < 10 {
//	            return errors.New("Comment must be at least 10 characters")
//	        }
//	        return nil
//	    },
//	    OnChange: func(text string) {
//	        fmt.Println("Text length:", len(text))
//	    },
//	})
//
//	// Initialize and use with Bubbletea
//	textarea.Init()
//	view := textarea.View()
//
// Features:
//   - Reactive multi-line text binding with Ref[string]
//   - Placeholder support
//   - Configurable height (rows)
//   - Maximum length enforcement
//   - Validation support
//   - OnChange callback support
//   - Disabled state support
//   - Theme integration
//   - Custom style override
//
// Visual layout:
//   - Bordered box containing text lines
//   - Each line displayed separately
//   - Placeholder shown when empty
//   - Error message displayed below if validation fails
//
// Accessibility:
//   - Clear visual distinction for disabled state
//   - Error messages clearly displayed
//   - Keyboard accessible
func TextArea(props TextAreaProps) bubbly.Component {
	component, _ := bubbly.NewComponent("TextArea").
		Props(props).
		Setup(func(ctx *bubbly.Context) {
			// Try to inject theme, fallback to DefaultTheme
			theme := DefaultTheme
			if injected := ctx.Inject("theme", nil); injected != nil {
				if t, ok := injected.(Theme); ok {
					theme = t
				}
			}

			// Create internal state for validation error
			validationError := bubbly.NewRef[error](nil)

			// Watch value changes for validation
			if props.Validate != nil {
				bubbly.Watch(props.Value, func(oldValue, newValue string) {
					err := props.Validate(newValue)
					validationError.Set(err)
				})
			}

			// Register change event handler
			ctx.On("change", func(data interface{}) {
				if !props.Disabled {
					if newValue, ok := data.(string); ok {
						// Enforce max length if specified
						if props.MaxLength > 0 && len(newValue) > props.MaxLength {
							newValue = newValue[:props.MaxLength]
						}

						props.Value.Set(newValue)

						// Call OnChange callback if provided
						if props.OnChange != nil {
							props.OnChange(newValue)
						}
					}
				}
			})

			// Expose internal state
			ctx.Expose("theme", theme)
			ctx.Expose("validationError", validationError)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			props := ctx.Props().(TextAreaProps)
			theme := ctx.Get("theme").(Theme)
			validationError := ctx.Get("validationError").(*bubbly.Ref[error])

			// Get current text
			text := props.Value.GetTyped()

			// Determine rows (default to 3 if not specified)
			rows := props.Rows
			if rows <= 0 {
				rows = 3
			}

			// Build textarea style
			textareaStyle := lipgloss.NewStyle().
				Border(theme.GetBorderStyle()).
				Padding(0, 1).
				Width(40) // Default width

			// Set border color based on state
			if props.Disabled {
				textareaStyle = textareaStyle.BorderForeground(theme.Muted)
			} else if validationError.GetTyped() != nil {
				textareaStyle = textareaStyle.BorderForeground(theme.Danger)
			} else {
				textareaStyle = textareaStyle.BorderForeground(theme.Secondary)
			}

			// Apply custom style if provided
			if props.Style != nil {
				textareaStyle = textareaStyle.Inherit(*props.Style)
			}

			// Build content
			var content strings.Builder

			if text == "" && props.Placeholder != "" {
				// Show placeholder
				placeholderStyle := lipgloss.NewStyle().Foreground(theme.Muted)
				content.WriteString(placeholderStyle.Render(props.Placeholder))
			} else {
				// Split text into lines
				lines := strings.Split(text, "\n")

				// Render lines (limit to rows)
				displayLines := lines
				if len(lines) > rows {
					// Show last N lines if content exceeds rows
					displayLines = lines[len(lines)-rows:]
				}

				for i, line := range displayLines {
					if props.Disabled {
						lineStyle := lipgloss.NewStyle().Foreground(theme.Muted)
						content.WriteString(lineStyle.Render(line))
					} else {
						content.WriteString(line)
					}

					// Add newline except for last line
					if i < len(displayLines)-1 {
						content.WriteString("\n")
					}
				}

				// Pad with empty lines if needed to reach rows
				for i := len(displayLines); i < rows; i++ {
					content.WriteString("\n")
				}
			}

			// Render textarea
			output := textareaStyle.Render(content.String())

			// Add validation error if present
			if err := validationError.GetTyped(); err != nil {
				errorStyle := lipgloss.NewStyle().
					Foreground(theme.Danger).
					Italic(true)
				output += "\n" + errorStyle.Render("✗ "+err.Error())
			}

			return output
		}).
		Build()

	return component
}
