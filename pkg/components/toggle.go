package components

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// ToggleProps defines the configuration properties for a Toggle component.
//
// Example usage:
//
//	valueRef := bubbly.NewRef(false)
//	toggle := components.Toggle(components.ToggleProps{
//	    Label: "Enable dark mode",
//	    Value: valueRef,
//	    OnChange: func(enabled bool) {
//	        if enabled {
//	            activateDarkMode()
//	        }
//	    },
//	})
type ToggleProps struct {
	// Label is the text displayed next to the toggle switch.
	// Optional - if empty, only the toggle indicator is shown.
	Label string

	// Value is the reactive reference to the toggle's on/off state.
	// Required - must be a valid Ref[bool].
	// Changes to this ref will update the toggle display.
	Value *bubbly.Ref[bool]

	// OnChange is a callback function executed when the toggle is switched.
	// Receives the new state as a parameter.
	// Optional - if nil, no callback is executed.
	OnChange func(bool)

	// Disabled indicates whether the toggle is disabled.
	// Disabled toggles do not respond to toggle events and are styled differently.
	// Default: false (enabled).
	Disabled bool

	// Common props for all components
	CommonProps
}

// Toggle creates a new Toggle molecule component.
//
// Toggle is an interactive switch element that allows users to turn a feature on or off.
// It supports reactive state binding, callbacks, and disabled states.
//
// The toggle automatically integrates with the theme system via the composition API's
// Provide/Inject mechanism. If no theme is provided, it uses DefaultTheme.
//
// Example:
//
//	valueRef := bubbly.NewRef(false)
//	toggle := components.Toggle(components.ToggleProps{
//	    Label: "Enable notifications",
//	    Value: valueRef,
//	    OnChange: func(enabled bool) {
//	        if enabled {
//	            enableNotifications()
//	        } else {
//	            disableNotifications()
//	        }
//	    },
//	})
//
//	// Initialize and use with Bubbletea
//	toggle.Init()
//	view := toggle.View()
//
// Features:
//   - Reactive on/off state binding with Ref[bool]
//   - Toggle functionality via "toggle" event
//   - OnChange callback support
//   - Disabled state support
//   - Label display
//   - Theme integration
//   - Custom style override
//
// Keyboard interaction:
//   - Space/Enter: Toggle switch (when focused)
//
// Visual indicators:
//   - Off: [OFF] or [─●]
//   - On: [ON ] or [●─]
//
// Accessibility:
//   - Clear visual distinction between on/off states
//   - Disabled state clearly indicated
//   - Keyboard accessible
func Toggle(props ToggleProps) bubbly.Component {
	component, _ := bubbly.NewComponent("Toggle").
		Props(props).
		Setup(func(ctx *bubbly.Context) {
			// Inject theme using helper
			setupTheme(ctx)

			// Register toggle event handler
			ctx.On("toggle", func(data interface{}) {
				// Only handle toggle if not disabled
				if !props.Disabled {
					// Flip the state
					currentValue := props.Value.GetTyped()
					newValue := !currentValue
					props.Value.Set(newValue)

					// Call OnChange callback if provided
					if props.OnChange != nil {
						props.OnChange(newValue)
					}
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			props := ctx.Props().(ToggleProps)
			theme := ctx.Get("theme").(Theme)

			// Get current state
			isOn := props.Value.GetTyped()

			// Determine toggle indicator
			var indicator string
			if isOn {
				indicator = "[ON ]" // On state
			} else {
				indicator = "[OFF]" // Off state
			}

			// Build toggle style
			toggleStyle := lipgloss.NewStyle()

			// Set color based on state
			if props.Disabled {
				// Disabled state: muted color
				toggleStyle = toggleStyle.Foreground(theme.Muted)
			} else if isOn {
				// On state: primary color
				toggleStyle = toggleStyle.Foreground(theme.Primary)
			} else {
				// Off state: secondary color
				toggleStyle = toggleStyle.Foreground(theme.Secondary)
			}

			// Apply custom style if provided
			if props.Style != nil {
				toggleStyle = toggleStyle.Inherit(*props.Style)
			}

			// Build output
			output := indicator
			if props.Label != "" {
				output += " " + props.Label
			}

			// Render with style
			return toggleStyle.Render(output)
		}).
		Build()

	return component
}
