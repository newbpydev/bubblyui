package components

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// ButtonVariant defines the visual style variant of a button.
// Variants map to theme colors for consistent styling across the application.
type ButtonVariant string

// Button variant constants.
const (
	// ButtonPrimary represents the primary/default button variant.
	// Used for main actions and primary calls-to-action.
	ButtonPrimary ButtonVariant = "primary"

	// ButtonSecondary represents a secondary/alternative button variant.
	// Used for less prominent actions.
	ButtonSecondary ButtonVariant = "secondary"

	// ButtonDanger represents a destructive/dangerous action variant.
	// Used for delete, remove, or other destructive operations.
	ButtonDanger ButtonVariant = "danger"

	// ButtonSuccess represents a successful/positive action variant.
	// Used for confirmations and positive actions.
	ButtonSuccess ButtonVariant = "success"

	// ButtonWarning represents a warning/caution variant.
	// Used for actions that require user attention.
	ButtonWarning ButtonVariant = "warning"

	// ButtonInfo represents an informational variant.
	// Used for neutral informational actions.
	ButtonInfo ButtonVariant = "info"
)

// ButtonProps defines the configuration properties for a Button component.
//
// Example usage:
//
//	button := components.Button(components.ButtonProps{
//	    Label:   "Submit",
//	    Variant: components.ButtonPrimary,
//	    OnClick: func() {
//	        handleSubmit()
//	    },
//	})
type ButtonProps struct {
	// Label is the text displayed on the button.
	// Required - should not be empty for usability.
	Label string

	// Variant determines the visual style of the button.
	// Defaults to ButtonPrimary if not specified.
	// Valid values: ButtonPrimary, ButtonSecondary, ButtonDanger, ButtonSuccess, ButtonWarning, ButtonInfo.
	Variant ButtonVariant

	// Disabled indicates whether the button is disabled.
	// Disabled buttons do not respond to click events and are styled differently.
	// Default: false (enabled).
	Disabled bool

	// OnClick is the callback function executed when the button is clicked.
	// Only called if the button is not disabled.
	// Optional - if nil, button will not respond to clicks.
	OnClick func()

	// Common props for all components
	CommonProps
}

// Button creates a new Button atom component.
//
// Button is a fundamental interactive element that triggers actions when clicked.
// It supports multiple visual variants, disabled states, and custom click handlers.
//
// The button automatically integrates with the theme system via the composition API's
// Provide/Inject mechanism. If no theme is provided, it uses DefaultTheme.
//
// Example:
//
//	button := components.Button(components.ButtonProps{
//	    Label:   "Save Changes",
//	    Variant: components.ButtonPrimary,
//	    OnClick: func() {
//	        fmt.Println("Saving...")
//	    },
//	})
//
//	// Initialize and use with Bubbletea
//	button.Init()
//	view := button.View()
//
// Keyboard interaction:
//   - Enter/Space: Trigger click event (when focused)
//
// Accessibility:
//   - Clear visual distinction between enabled/disabled states
//   - Keyboard accessible
//   - High contrast variants available via theme
func Button(props ButtonProps) bubbly.Component {
	// Default to primary variant if not specified
	if props.Variant == "" {
		props.Variant = ButtonPrimary
	}

	component, _ := bubbly.NewComponent("Button").
		Props(props).
		Setup(func(ctx *bubbly.Context) {
			// Try to inject theme, fallback to DefaultTheme
			theme := DefaultTheme
			if injected := ctx.Inject("theme", nil); injected != nil {
				if t, ok := injected.(Theme); ok {
					theme = t
				}
			}

			// Expose theme for template
			ctx.Expose("theme", theme)

			// Register click event handler
			ctx.On("click", func(data interface{}) {
				// Only handle click if not disabled and handler exists
				if !props.Disabled && props.OnClick != nil {
					props.OnClick()
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			props := ctx.Props().(ButtonProps)
			theme := ctx.Get("theme").(Theme)

			// Get variant color from theme
			variantColor := theme.GetVariantColor(Variant(props.Variant))

			// Build button style based on state
			buttonStyle := lipgloss.NewStyle().
				Padding(0, 2).
				Bold(true)

			if props.Disabled {
				// Disabled state: muted colors, no background
				buttonStyle = buttonStyle.
					Foreground(theme.Muted).
					Border(theme.GetBorderStyle()).
					BorderForeground(theme.Muted)
			} else {
				// Enabled state: variant colors with background
				buttonStyle = buttonStyle.
					Foreground(lipgloss.Color("230")). // Light text
					Background(variantColor).
					Border(theme.GetBorderStyle()).
					BorderForeground(variantColor)
			}

			// Apply custom style if provided
			if props.Style != nil {
				// Custom style overrides
				buttonStyle = buttonStyle.Inherit(*props.Style)
			}

			// Render button with label
			return buttonStyle.Render(props.Label)
		}).
		Build()

	return component
}
