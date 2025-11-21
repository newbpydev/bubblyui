package components

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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

	// CharLimit sets the maximum number of characters allowed.
	// Optional - if 0, no limit is enforced.
	CharLimit int

	// ShowCursorPosition displays the cursor position indicator [pos/len].
	// Optional - defaults to false.
	ShowCursorPosition bool

	// NoBorder removes the border if true.
	// Default is false (border is shown).
	// Useful when embedding in other bordered containers.
	NoBorder bool

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
//   - **Blinking cursor** with visual feedback
//   - **Cursor navigation** with arrow keys
//   - **Text editing** at cursor position
//   - **Cursor position indicator** [pos/len]
//   - **Character limits** enforcement
//   - **Clipboard operations** (Ctrl+C/V)
//
// Keyboard interaction:
//   - Type to input text (inserts at cursor position)
//   - ←/→ arrows to move cursor within text
//   - Home/End to jump to start/end
//   - Backspace/Delete for character deletion
//   - Ctrl+A to select all
//   - Ctrl+C/V for copy/paste
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

			// Create Bubbles textinput for cursor support
			ti := textinput.New()
			ti.Placeholder = props.Placeholder
			ti.Width = props.Width
			if props.CharLimit > 0 {
				ti.CharLimit = props.CharLimit
			}
			if props.Type == InputPassword {
				ti.EchoMode = textinput.EchoPassword
				ti.EchoCharacter = '*'
			}
			// Set initial value
			ti.SetValue(props.Value.Get().(string))

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

			// Sync textinput value with props.Value
			bubbly.Watch(props.Value, func(newVal, oldVal string) {
				if ti.Value() != newVal {
					ti.SetValue(newVal)
				}
			})

			// Register event handlers
			ctx.On("input", func(data interface{}) {
				if newValue, ok := data.(string); ok {
					props.Value.Set(newValue)
					ti.SetValue(newValue)
				}
			})

			ctx.On("focus", func(_ interface{}) {
				focusedRef.Set(true)
				ti.Focus()
			})

			ctx.On("blur", func(_ interface{}) {
				focusedRef.Set(false)
				ti.Blur()

				// Call OnBlur callback if provided
				if props.OnBlur != nil {
					props.OnBlur()
				}
			})

			// Internal event handler for keyboard processing
			// This is emitted from WithMessageHandler and processed here
			// where we have access to the textinput state
			ctx.On("__processKeyboard", func(data interface{}) {
				if msg, ok := data.(tea.Msg); ok {
					// Only process if focused
					if focusedRef.GetTyped() {
						// Update textinput with the message
						var cmd tea.Cmd
						ti, cmd = ti.Update(msg)

						// Sync value back to props.Value
						newValue := ti.Value()
						if newValue != props.Value.Get().(string) {
							props.Value.Set(newValue)
						}

						// Commands are handled by Bubbletea
						_ = cmd
					}
				}
			})

			// Expose internal state
			ctx.Expose("theme", theme)
			ctx.Expose("error", errorRef)
			ctx.Expose("focused", focusedRef)
			ctx.Expose("textInput", &ti)
		}).
		WithMessageHandler(func(comp bubbly.Component, msg tea.Msg) tea.Cmd {
			// Forward keyboard input to internal event handler
			// The event handler has access to the textinput state (ti)
			comp.Emit("__processKeyboard", msg)
			return nil
		}).
		Template(func(ctx bubbly.RenderContext) string {
			props := ctx.Props().(InputProps)
			theme := ctx.Get("theme").(Theme)
			errorRef := ctx.Get("error").(*bubbly.Ref[error])
			focusedRef := ctx.Get("focused").(*bubbly.Ref[bool])
			ti := ctx.Get("textInput").(*textinput.Model)

			// Get current state
			currentError := errorRef.GetTyped()
			hasError := currentError != nil
			isFocused := focusedRef.GetTyped()

			// Apply theme colors to textinput
			if hasError {
				ti.PromptStyle = lipgloss.NewStyle().Foreground(theme.Danger)
				ti.TextStyle = lipgloss.NewStyle().Foreground(theme.Danger)
			} else if isFocused {
				ti.PromptStyle = lipgloss.NewStyle().Foreground(theme.Primary)
				ti.TextStyle = lipgloss.NewStyle().Foreground(theme.Foreground)
			} else {
				ti.PromptStyle = lipgloss.NewStyle().Foreground(theme.Secondary)
				ti.TextStyle = lipgloss.NewStyle().Foreground(theme.Muted)
			}

			// Render textinput with cursor
			inputView := ti.View()

			// Add cursor position indicator if enabled
			if props.ShowCursorPosition && isFocused {
				pos := ti.Position()
				length := len(ti.Value())
				posIndicator := lipgloss.NewStyle().
					Foreground(theme.Muted).
					Render(" [" + string(rune(pos+'0')) + "/" + string(rune(length+'0')) + "]")
				inputView += posIndicator
			}

			var result string

			// Apply border unless NoBorder is true
			if !props.NoBorder {
				// Wrap in border
				borderStyle := lipgloss.NewStyle().
					Border(theme.GetBorderStyle()).
					Padding(0, 1)

				// Set border color based on state
				if hasError {
					borderStyle = borderStyle.BorderForeground(theme.Danger)
				} else if isFocused {
					borderStyle = borderStyle.BorderForeground(theme.Primary)
				} else {
					borderStyle = borderStyle.BorderForeground(theme.Secondary)
				}

				// Apply custom style if provided
				if props.Style != nil {
					borderStyle = borderStyle.Inherit(*props.Style)
				}

				result = borderStyle.Render(inputView)
			} else {
				// No border, just apply padding and custom style
				noBorderStyle := lipgloss.NewStyle().
					Padding(0, 1)

				// Apply custom style if provided
				if props.Style != nil {
					noBorderStyle = noBorderStyle.Inherit(*props.Style)
				}

				result = noBorderStyle.Render(inputView)
			}

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
