package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// SpinnerProps defines the configuration properties for a Spinner component.
//
// Example usage:
//
//	spinner := components.Spinner(components.SpinnerProps{
//	    Label:  "Loading...",
//	    Active: true,
//	    Color:  lipgloss.Color("99"),
//	})
type SpinnerProps struct {
	// Label is optional text to display next to the spinner.
	// Typically used to describe what is loading.
	// Optional - if empty, only the spinner animation is shown.
	Label string

	// Active determines whether the spinner is animating.
	// When false, the spinner is hidden or shows a static state.
	// Default: false.
	Active bool

	// Color sets the foreground color of the spinner.
	// Optional - if not specified, uses theme primary color.
	Color lipgloss.Color

	// Common props for all components
	CommonProps
}

// Spinner creates a new Spinner atom component.
//
// Spinner is a loading indicator component that shows an animated symbol
// to indicate background activity or loading states. It can optionally
// display a label describing the operation.
//
// The spinner component automatically integrates with the theme system via the composition API's
// Provide/Inject mechanism. If no theme is provided, it uses DefaultTheme.
//
// Example:
//
//	spinner := components.Spinner(components.SpinnerProps{
//	    Label:  "Loading data...",
//	    Active: true,
//	    Color:  lipgloss.Color("99"),
//	})
//
//	// Initialize and use with Bubbletea
//	spinner.Init()
//	view := spinner.View()
//
// Common use cases:
//   - Loading indicators
//   - Background processing
//   - Data fetching states
//   - Long-running operations
//
// Note: This is a simplified spinner implementation.
// For advanced animations with Bubbletea tick messages,
// consider using the bubbles/spinner package directly.
//
// Accessibility:
//   - Clear visual indication of activity
//   - Optional label for context
//   - Can be hidden when inactive
func Spinner(props SpinnerProps) bubbly.Component {
	component, _ := bubbly.NewComponent("Spinner").
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

			// Create a simple frame counter for animation
			// In a real implementation with Bubbletea, this would use tick messages
			frame := ctx.Ref(0)
			ctx.Expose("frame", frame)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			props := ctx.Props().(SpinnerProps)
			theme := ctx.Get("theme").(Theme)
			frameRef := ctx.Get("frame")

			// Type assert the frame reference
			var frameIndex int
			if fr, ok := frameRef.(*bubbly.Ref[int]); ok {
				if val, ok := fr.Get().(int); ok {
					frameIndex = val
				}
			}

			// If not active, show label only or nothing
			if !props.Active {
				if props.Label != "" {
					return props.Label
				}
				return ""
			}

			// Simple spinner frames (dots animation)
			spinnerFrames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
			currentFrame := spinnerFrames[frameIndex%len(spinnerFrames)]

			// Build spinner style
			spinnerStyle := lipgloss.NewStyle()

			// Apply color
			if props.Color != "" {
				spinnerStyle = spinnerStyle.Foreground(props.Color)
			} else {
				// Use theme primary color
				spinnerStyle = spinnerStyle.Foreground(theme.Primary)
			}

			// Apply custom style if provided
			if props.Style != nil {
				spinnerStyle = spinnerStyle.Inherit(*props.Style)
			}

			// Render spinner with optional label
			result := spinnerStyle.Render(currentFrame)
			if props.Label != "" {
				result = fmt.Sprintf("%s %s", result, props.Label)
			}

			return result
		}).
		Build()

	return component
}
