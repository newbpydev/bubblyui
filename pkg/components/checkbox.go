package components

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// CheckboxProps defines the configuration properties for a Checkbox component.
//
// Example usage:
//
//	checkedRef := bubbly.NewRef(false)
//	checkbox := components.Checkbox(components.CheckboxProps{
//	    Label:   "Accept terms and conditions",
//	    Checked: checkedRef,
//	    OnChange: func(checked bool) {
//	        fmt.Println("Checkbox is now:", checked)
//	    },
//	})
type CheckboxProps struct {
	// Label is the text displayed next to the checkbox.
	// Optional - if empty, only the checkbox indicator is shown.
	Label string

	// Checked is the reactive reference to the checkbox's checked state.
	// Required - must be a valid Ref[bool].
	// Changes to this ref will update the checkbox display.
	Checked *bubbly.Ref[bool]

	// OnChange is a callback function executed when the checkbox is toggled.
	// Receives the new checked state as a parameter.
	// Optional - if nil, no callback is executed.
	OnChange func(bool)

	// Disabled indicates whether the checkbox is disabled.
	// Disabled checkboxes do not respond to toggle events and are styled differently.
	// Default: false (enabled).
	Disabled bool

	// Common props for all components
	CommonProps
}

// Checkbox creates a new Checkbox molecule component.
//
// Checkbox is an interactive toggle element that allows users to select or deselect an option.
// It supports reactive state binding, callbacks, and disabled states.
//
// The checkbox automatically integrates with the theme system via the composition API's
// Provide/Inject mechanism. If no theme is provided, it uses DefaultTheme.
//
// Example:
//
//	checkedRef := bubbly.NewRef(false)
//	checkbox := components.Checkbox(components.CheckboxProps{
//	    Label:   "Enable notifications",
//	    Checked: checkedRef,
//	    OnChange: func(checked bool) {
//	        if checked {
//	            enableNotifications()
//	        } else {
//	            disableNotifications()
//	        }
//	    },
//	})
//
//	// Initialize and use with Bubbletea
//	checkbox.Init()
//	view := checkbox.View()
//
// Features:
//   - Reactive checked state binding with Ref[bool]
//   - Toggle functionality via "toggle" event
//   - OnChange callback support
//   - Disabled state support
//   - Label display
//   - Theme integration
//   - Custom style override
//
// Keyboard interaction:
//   - Space/Enter: Toggle checkbox (when focused)
//
// Visual indicators:
//   - Unchecked: ☐ (or [ ])
//   - Checked: ☑ (or [x])
//
// Accessibility:
//   - Clear visual distinction between checked/unchecked states
//   - Disabled state clearly indicated
//   - Keyboard accessible
func Checkbox(props CheckboxProps) bubbly.Component {
	component, _ := bubbly.NewComponent("Checkbox").
		Props(props).
		Setup(func(ctx *bubbly.Context) {
			// Try to inject theme, fallback to DefaultTheme
			theme := DefaultTheme
			if injected := ctx.Inject("theme", nil); injected != nil {
				if t, ok := injected.(Theme); ok {
					theme = t
				}
			}

			// Register toggle event handler
			ctx.On("toggle", func(data interface{}) {
				// Only handle toggle if not disabled
				if !props.Disabled {
					// Flip the checked state
					currentValue := props.Checked.GetTyped()
					newValue := !currentValue
					props.Checked.Set(newValue)

					// Call OnChange callback if provided
					if props.OnChange != nil {
						props.OnChange(newValue)
					}
				}
			})

			// Expose theme for template
			ctx.Expose("theme", theme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			props := ctx.Props().(CheckboxProps)
			theme := ctx.Get("theme").(Theme)

			// Get current checked state
			isChecked := props.Checked.GetTyped()

			// Determine checkbox indicator
			var indicator string
			if isChecked {
				indicator = "☑" // Checked box (U+2611)
			} else {
				indicator = "☐" // Unchecked box (U+2610)
			}

			// Build checkbox style
			checkboxStyle := lipgloss.NewStyle()

			// Set color based on state
			if props.Disabled {
				// Disabled state: muted color
				checkboxStyle = checkboxStyle.Foreground(theme.Muted)
			} else if isChecked {
				// Checked state: primary color
				checkboxStyle = checkboxStyle.Foreground(theme.Primary)
			} else {
				// Unchecked state: secondary color
				checkboxStyle = checkboxStyle.Foreground(theme.Secondary)
			}

			// Apply custom style if provided
			if props.Style != nil {
				checkboxStyle = checkboxStyle.Inherit(*props.Style)
			}

			// Build output
			output := indicator
			if props.Label != "" {
				output += " " + props.Label
			}

			// Render with style
			return checkboxStyle.Render(output)
		}).
		Build()

	return component
}
