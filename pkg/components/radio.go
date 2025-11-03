package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// RadioProps defines the configuration properties for a Radio component.
//
// Radio is a generic component that works with any type T.
//
// Example usage:
//
//	valueRef := bubbly.NewRef("option1")
//	radio := components.Radio(components.RadioProps[string]{
//	    Value:   valueRef,
//	    Options: []string{"option1", "option2", "option3"},
//	    OnChange: func(value string) {
//	        fmt.Println("Selected:", value)
//	    },
//	})
type RadioProps[T any] struct {
	// Value is the reactive reference to the selected value.
	// Required - must be a valid Ref[T].
	// Changes to this ref will update the radio display.
	Value *bubbly.Ref[T]

	// Options is the list of available options to choose from.
	// Required - should not be empty for usability.
	Options []T

	// OnChange is a callback function executed when the selection changes.
	// Receives the newly selected value as a parameter.
	// Optional - if nil, no callback is executed.
	OnChange func(T)

	// Disabled indicates whether the radio is disabled.
	// Disabled radios do not respond to events and are styled differently.
	// Default: false (enabled).
	Disabled bool

	// RenderOption is a custom function to render each option as a string.
	// Optional - if nil, uses fmt.Sprintf("%v", option) for default rendering.
	// Useful for complex types that need custom display logic.
	RenderOption func(T) string

	// Common props for all components
	CommonProps
}

// Radio creates a new Radio molecule component with generic type support.
//
// Radio is an interactive selection element that allows users to choose one option
// from a list. Unlike Select, all options are always visible.
//
// The radio automatically integrates with the theme system via the composition API's
// Provide/Inject mechanism. If no theme is provided, it uses DefaultTheme.
//
// Example:
//
//	valueRef := bubbly.NewRef("dark")
//	radio := components.Radio(components.RadioProps[string]{
//	    Value:   valueRef,
//	    Options: []string{"light", "dark", "auto"},
//	    OnChange: func(theme string) {
//	        applyTheme(theme)
//	    },
//	})
//
//	// Initialize and use with Bubbletea
//	radio.Init()
//	view := radio.View()
//
// Features:
//   - Generic type support for any option type
//   - Reactive value binding with Ref[T]
//   - Keyboard navigation (up/down arrows)
//   - Selection with Enter/Space
//   - OnChange callback support
//   - Disabled state support
//   - Custom option rendering
//   - Theme integration
//   - Custom style override
//
// Keyboard interaction:
//   - Up Arrow: Navigate to previous option
//   - Down Arrow: Navigate to next option
//   - Enter/Space: Select highlighted option
//
// Visual indicators:
//   - Selected: (●) Option
//   - Unselected: ( ) Option
//
// Accessibility:
//   - Clear visual distinction between states
//   - Highlighted selection
//   - Disabled state clearly indicated
//   - Keyboard accessible
func Radio[T any](props RadioProps[T]) bubbly.Component {
	component, _ := bubbly.NewComponent("Radio").
		Props(props).
		Setup(func(ctx *bubbly.Context) {
			// Try to inject theme, fallback to DefaultTheme
			theme := DefaultTheme
			if injected := ctx.Inject("theme", nil); injected != nil {
				if t, ok := injected.(Theme); ok {
					theme = t
				}
			}

			// Create internal state for highlighted index
			highlightedIndex := bubbly.NewRef(0)

			// Find initial highlighted index based on current value
			currentValue := props.Value.GetTyped()
			for i, opt := range props.Options {
				if fmt.Sprintf("%v", opt) == fmt.Sprintf("%v", currentValue) {
					highlightedIndex.Set(i)
					break
				}
			}

			// Register up navigation event
			ctx.On("up", func(data interface{}) {
				if !props.Disabled && len(props.Options) > 0 {
					currentIdx := highlightedIndex.GetTyped()
					newIdx := currentIdx - 1
					if newIdx < 0 {
						newIdx = len(props.Options) - 1 // Wrap to bottom
					}
					highlightedIndex.Set(newIdx)
				}
			})

			// Register down navigation event
			ctx.On("down", func(data interface{}) {
				if !props.Disabled && len(props.Options) > 0 {
					currentIdx := highlightedIndex.GetTyped()
					newIdx := (currentIdx + 1) % len(props.Options) // Wrap to top
					highlightedIndex.Set(newIdx)
				}
			})

			// Register select event (confirm selection)
			ctx.On("select", func(data interface{}) {
				if !props.Disabled && len(props.Options) > 0 {
					idx := highlightedIndex.GetTyped()
					if idx >= 0 && idx < len(props.Options) {
						selectedValue := props.Options[idx]
						props.Value.Set(selectedValue)

						// Call OnChange callback if provided
						if props.OnChange != nil {
							props.OnChange(selectedValue)
						}
					}
				}
			})

			// Expose internal state
			ctx.Expose("theme", theme)
			ctx.Expose("highlightedIndex", highlightedIndex)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			props := ctx.Props().(RadioProps[T])
			theme := ctx.Get("theme").(Theme)
			highlightedIndex := ctx.Get("highlightedIndex").(*bubbly.Ref[int])

			// Get current state
			currentValue := props.Value.GetTyped()
			currentIdx := highlightedIndex.GetTyped()

			// Helper function to render an option
			renderOption := func(opt T) string {
				if props.RenderOption != nil {
					return props.RenderOption(opt)
				}
				return fmt.Sprintf("%v", opt)
			}

			// Handle empty options
			if len(props.Options) == 0 {
				emptyStyle := lipgloss.NewStyle().Foreground(theme.Muted)
				return emptyStyle.Render("No options available")
			}

			var output strings.Builder

			// Render each option
			for i, opt := range props.Options {
				optionText := renderOption(opt)

				// Determine if this option is selected
				isSelected := fmt.Sprintf("%v", opt) == fmt.Sprintf("%v", currentValue)

				// Determine indicator
				var indicator string
				if isSelected {
					indicator = "(●)" // Selected (filled circle)
				} else {
					indicator = "( )" // Unselected (empty circle)
				}

				// Build option style
				optionStyle := lipgloss.NewStyle()

				if props.Disabled {
					// Disabled state
					optionStyle = optionStyle.Foreground(theme.Muted)
				} else if i == currentIdx {
					// Highlighted option
					optionStyle = optionStyle.Foreground(theme.Primary).Bold(true)
				} else if isSelected {
					// Selected but not highlighted
					optionStyle = optionStyle.Foreground(theme.Primary)
				} else {
					// Normal option
					optionStyle = optionStyle.Foreground(theme.Foreground)
				}

				// Apply custom style if provided
				if props.Style != nil {
					optionStyle = optionStyle.Inherit(*props.Style)
				}

				// Build line
				line := indicator + " " + optionText
				output.WriteString(optionStyle.Render(line))

				// Add newline except for last option
				if i < len(props.Options)-1 {
					output.WriteString("\n")
				}
			}

			return output.String()
		}).
		Build()

	return component
}
