package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// InputType defines the type of input field.
// Different types may have different validation and display behavior.
type InputType string

// Input type constants.
const (
	// InputText represents a standard text input field.
	// Displays characters as typed.
	InputText InputType = "text"

	// InputPassword represents a password input field.
	// Masks characters with asterisks for security.
	InputPassword InputType = "password"

	// InputEmail represents an email input field.
	// Displays characters as typed, intended for email addresses.
	InputEmail InputType = "email"
)

// InputProps defines the configuration properties for an Input component.
//
// Example usage:
//
//	valueRef := bubbly.NewRef("")
//	input := components.Input(components.InputProps{
//	    Value:       valueRef,
//	    Placeholder: "Enter your name",
//	    Type:        components.InputText,
//	    Validate: func(s string) error {
//	        if len(s) < 3 {
//	            return errors.New("name must be at least 3 characters")
//	        }
//	        return nil
//	    },
//	})
type InputProps struct {
	// Value is the reactive reference to the input's value.
	// Required - must be a valid Ref[string].
	// Changes to this ref will update the input display.
	Value *bubbly.Ref[string]

	// Placeholder is the text displayed when the input is empty.
	// Optional - if empty, no placeholder is shown.
	Placeholder string

	// Type determines the input field type.
	// Valid values: InputText, InputPassword, InputEmail.
	// Default: InputText.
	Type InputType

	// Validate is a function called when the value changes.
	// If it returns an error, the error message is displayed below the input.
	// Optional - if nil, no validation is performed.
	Validate func(string) error

	// OnChange is a callback function executed when the value changes.
	// Called after validation.
	// Optional - if nil, no callback is executed.
	OnChange func(string)

	// OnBlur is a callback function executed when the input loses focus.
	// Optional - if nil, no callback is executed.
	OnBlur func()

	// Width sets the width of the input field in characters.
	// Optional - if 0, defaults to 30 characters.
	Width int

	// Common props for all components
	CommonProps
}

// Input creates a new Input molecule component.
//
// Input is an interactive text field that supports reactive value binding,
// validation, focus states, and different input types (text, password, email).
//
// The input automatically integrates with the theme system via the composition API's
// Provide/Inject mechanism. If no theme is provided, it uses DefaultTheme.
//
// Example:
//
//	valueRef := bubbly.NewRef("")
//	input := components.Input(components.InputProps{
//	    Value:       valueRef,
//	    Placeholder: "Enter email",
//	    Type:        components.InputEmail,
//	    Validate: func(s string) error {
//	        if !strings.Contains(s, "@") {
//	            return errors.New("invalid email address")
//	        }
//	        return nil
//	    },
//	    OnChange: func(value string) {
//	        fmt.Println("Email changed:", value)
//	    },
//	})
//
//	// Initialize and use with Bubbletea
//	input.Init()
//	view := input.View()
//
// Features:
//   - Reactive value binding with Ref[string]
//   - Real-time validation with error display
//   - Focus state management
//   - Password masking
//   - Placeholder support
//   - Custom width
//   - Theme integration
//
// Keyboard interaction:
//   - Type to input text
//   - Tab/Shift+Tab for focus navigation
//   - Esc to blur
//
// Accessibility:
//   - Clear visual distinction between focused/unfocused states
//   - Error messages displayed inline
//   - High contrast colors for error states
func Input(props InputProps) bubbly.Component {
	// Default to text type if not specified
	if props.Type == "" {
		props.Type = InputText
	}

	// Default width if not specified
	if props.Width == 0 {
		props.Width = 30
	}

	component, _ := bubbly.NewComponent("Input").
		Props(props).
		Setup(func(ctx *bubbly.Context) {
			// Try to inject theme, fallback to DefaultTheme
			theme := DefaultTheme
			if injected := ctx.Inject("theme", nil); injected != nil {
				if t, ok := injected.(Theme); ok {
					theme = t
				}
			}

			// Create internal state with typed refs
			errorRef := bubbly.NewRef[error](nil)
			focusedRef := bubbly.NewRef(false)

			// Watch value changes for validation
			if props.Validate != nil {
				bubbly.Watch(props.Value, func(newVal, oldVal string) {
					err := props.Validate(newVal)
					errorRef.Set(err)

					// Call OnChange callback if provided
					if props.OnChange != nil {
						props.OnChange(newVal)
					}
				})
			} else if props.OnChange != nil {
				// If no validation but OnChange provided, still watch for changes
				bubbly.Watch(props.Value, func(newVal, oldVal string) {
					props.OnChange(newVal)
				})
			}

			// Register event handlers
			ctx.On("input", func(data interface{}) {
				if newValue, ok := data.(string); ok {
					props.Value.Set(newValue)
				}
			})

			ctx.On("focus", func(_ interface{}) {
				focusedRef.Set(true)
			})

			ctx.On("blur", func(_ interface{}) {
				focusedRef.Set(false)

				// Call OnBlur callback if provided
				if props.OnBlur != nil {
					props.OnBlur()
				}
			})

			// Expose internal state
			ctx.Expose("theme", theme)
			ctx.Expose("error", errorRef)
			ctx.Expose("focused", focusedRef)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			props := ctx.Props().(InputProps)
			theme := ctx.Get("theme").(Theme)
			errorRef := ctx.Get("error").(*bubbly.Ref[error])
			focusedRef := ctx.Get("focused").(*bubbly.Ref[bool])

			// Get current state - use GetTyped() for type safety
			value := props.Value.GetTyped()
			currentError := errorRef.GetTyped()
			hasError := currentError != nil
			isFocused := focusedRef.GetTyped()

			// Build input style based on state
			inputStyle := lipgloss.NewStyle().
				Width(props.Width).
				Padding(0, 1).
				Border(theme.GetBorderStyle())

			// Set border color based on state
			if hasError {
				// Error state: red border
				inputStyle = inputStyle.BorderForeground(theme.Danger)
			} else if isFocused {
				// Focused state: primary color border
				inputStyle = inputStyle.BorderForeground(theme.Primary)
			} else {
				// Normal state: secondary/muted border
				inputStyle = inputStyle.BorderForeground(theme.Secondary)
			}

			// Determine display value
			displayValue := value
			if displayValue == "" && !isFocused {
				// Show placeholder when empty and not focused
				displayValue = props.Placeholder
				inputStyle = inputStyle.Foreground(theme.Muted)
			} else if props.Type == InputPassword && displayValue != "" {
				// Mask password with asterisks
				displayValue = strings.Repeat("*", len(displayValue))
			}

			// Apply custom style if provided
			if props.Style != nil {
				inputStyle = inputStyle.Inherit(*props.Style)
			}

			// Render input field
			result := inputStyle.Render(displayValue)

			// Add error message if present
			if currentError != nil {
				errorStyle := lipgloss.NewStyle().
					Foreground(theme.Danger).
					Italic(true)
				result += "\n" + errorStyle.Render(currentError.Error())
			}

			return result
		}).
		Build()

	return component
}
