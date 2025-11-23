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

	// Width sets the width of the textarea in characters.
	// Optional - if 0, defaults to 40 characters.
	Width int

	// NoBorder removes the border if true.
	// Default is false (border is shown).
	// Useful when embedding in other bordered containers.
	NoBorder bool

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
//
// textareaApplyDefaults returns rows and width with defaults applied.
func textareaApplyDefaults(props TextAreaProps) (int, int) {
	rows := props.Rows
	if rows <= 0 {
		rows = 3
	}
	width := props.Width
	if width <= 0 {
		width = 40
	}
	return rows, width
}

// textareaGetBorderColor returns the border color based on state.
func textareaGetBorderColor(props TextAreaProps, validationError error, theme Theme) lipgloss.TerminalColor {
	if props.Disabled {
		return theme.Muted
	}
	if validationError != nil {
		return theme.Danger
	}
	return theme.Secondary
}

// textareaRenderContent renders the text content area.
func textareaRenderContent(text string, props TextAreaProps, rows int, theme Theme) string {
	var content strings.Builder

	if text == "" && props.Placeholder != "" {
		return lipgloss.NewStyle().Foreground(theme.Muted).Render(props.Placeholder)
	}

	lines := strings.Split(text, "\n")
	displayLines := lines
	if len(lines) > rows {
		displayLines = lines[len(lines)-rows:]
	}

	for i, line := range displayLines {
		if props.Disabled {
			content.WriteString(lipgloss.NewStyle().Foreground(theme.Muted).Render(line))
		} else {
			content.WriteString(line)
		}
		if i < len(displayLines)-1 {
			content.WriteString("\n")
		}
	}

	for i := len(displayLines); i < rows; i++ {
		content.WriteString("\n")
	}

	return content.String()
}

func TextArea(props TextAreaProps) bubbly.Component {
	component, _ := bubbly.NewComponent("TextArea").
		Props(props).
		Setup(func(ctx *bubbly.Context) {
			theme := injectTheme(ctx)
			validationError := bubbly.NewRef[error](nil)

			if props.Validate != nil {
				bubbly.Watch(props.Value, func(_, newValue string) {
					validationError.Set(props.Validate(newValue))
				})
			}

			ctx.On("change", func(data interface{}) {
				if props.Disabled {
					return
				}
				if newValue, ok := data.(string); ok {
					if props.MaxLength > 0 && len(newValue) > props.MaxLength {
						newValue = newValue[:props.MaxLength]
					}
					props.Value.Set(newValue)
					if props.OnChange != nil {
						props.OnChange(newValue)
					}
				}
			})

			ctx.Expose("theme", theme)
			ctx.Expose("validationError", validationError)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			props := ctx.Props().(TextAreaProps)
			theme := ctx.Get("theme").(Theme)
			validationError := ctx.Get("validationError").(*bubbly.Ref[error])

			text := props.Value.GetTyped()
			rows, width := textareaApplyDefaults(props)

			textareaStyle := lipgloss.NewStyle().Padding(0, 1).Width(width)
			if !props.NoBorder {
				textareaStyle = textareaStyle.
					Border(theme.GetBorderStyle()).
					BorderForeground(textareaGetBorderColor(props, validationError.GetTyped(), theme))
			}

			if props.Style != nil {
				textareaStyle = textareaStyle.Inherit(*props.Style)
			}

			content := textareaRenderContent(text, props, rows, theme)
			output := textareaStyle.Render(content)

			if err := validationError.GetTyped(); err != nil {
				errorStyle := lipgloss.NewStyle().Foreground(theme.Danger).Italic(true)
				output += "\n" + errorStyle.Render("✗ "+err.Error())
			}

			return output
		}).
		Build()

	return component
}
