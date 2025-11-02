package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// SelectProps defines the configuration properties for a Select component.
//
// Select is a generic component that works with any type T.
//
// Example usage:
//
//	valueRef := bubbly.NewRef("option1")
//	selectComp := components.Select(components.SelectProps[string]{
//	    Value:   valueRef,
//	    Options: []string{"option1", "option2", "option3"},
//	    OnChange: func(value string) {
//	        fmt.Println("Selected:", value)
//	    },
//	})
type SelectProps[T any] struct {
	// Value is the reactive reference to the selected value.
	// Required - must be a valid Ref[T].
	// Changes to this ref will update the select display.
	Value *bubbly.Ref[T]

	// Options is the list of available options to choose from.
	// Required - should not be empty for usability.
	Options []T

	// OnChange is a callback function executed when the selection changes.
	// Receives the newly selected value as a parameter.
	// Optional - if nil, no callback is executed.
	OnChange func(T)

	// Placeholder is the text displayed when no value is selected.
	// Optional - if empty, shows the first option or empty string.
	Placeholder string

	// Disabled indicates whether the select is disabled.
	// Disabled selects do not respond to events and are styled differently.
	// Default: false (enabled).
	Disabled bool

	// RenderOption is a custom function to render each option as a string.
	// Optional - if nil, uses fmt.Sprintf("%v", option) for default rendering.
	// Useful for complex types that need custom display logic.
	RenderOption func(T) string

	// Common props for all components
	CommonProps
}

// Select creates a new Select molecule component with generic type support.
//
// Select is an interactive dropdown element that allows users to choose one option
// from a list. It supports reactive state binding, keyboard navigation, and callbacks.
//
// The select automatically integrates with the theme system via the composition API's
// Provide/Inject mechanism. If no theme is provided, it uses DefaultTheme.
//
// Example:
//
//	valueRef := bubbly.NewRef("dark")
//	selectComp := components.Select(components.SelectProps[string]{
//	    Value:       valueRef,
//	    Options:     []string{"light", "dark", "auto"},
//	    Placeholder: "Choose theme",
//	    OnChange: func(theme string) {
//	        applyTheme(theme)
//	    },
//	})
//
//	// Initialize and use with Bubbletea
//	selectComp.Init()
//	view := selectComp.View()
//
// Features:
//   - Generic type support for any option type
//   - Reactive value binding with Ref[T]
//   - Dropdown open/close functionality
//   - Keyboard navigation (up/down arrows)
//   - Selection with Enter/Space
//   - OnChange callback support
//   - Disabled state support
//   - Custom option rendering
//   - Theme integration
//   - Custom style override
//
// Keyboard interaction:
//   - Space/Enter: Toggle dropdown open/closed
//   - Up Arrow: Navigate to previous option
//   - Down Arrow: Navigate to next option
//   - Enter (when open): Select highlighted option
//   - Escape: Close dropdown without selecting
//
// Visual indicators:
//   - Closed: ▼ indicator
//   - Open: ▲ indicator with options list
//
// Accessibility:
//   - Clear visual distinction between states
//   - Highlighted selection in dropdown
//   - Disabled state clearly indicated
//   - Keyboard accessible
func Select[T any](props SelectProps[T]) bubbly.Component {
	component, _ := bubbly.NewComponent("Select").
		Props(props).
		Setup(func(ctx *bubbly.Context) {
			// Try to inject theme, fallback to DefaultTheme
			theme := DefaultTheme
			if injected := ctx.Inject("theme", nil); injected != nil {
				if t, ok := injected.(Theme); ok {
					theme = t
				}
			}

			// Create internal state
			isOpen := bubbly.NewRef(false)
			selectedIndex := bubbly.NewRef(0)

			// Find initial selected index based on current value
			currentValue := props.Value.GetTyped()
			for i, opt := range props.Options {
				if fmt.Sprintf("%v", opt) == fmt.Sprintf("%v", currentValue) {
					selectedIndex.Set(i)
					break
				}
			}

			// Register toggle event handler (open/close dropdown)
			ctx.On("toggle", func(data interface{}) {
				if !props.Disabled {
					isOpen.Set(!isOpen.GetTyped())
				}
			})

			// Register up navigation event
			ctx.On("up", func(data interface{}) {
				if !props.Disabled && isOpen.GetTyped() && len(props.Options) > 0 {
					currentIdx := selectedIndex.GetTyped()
					newIdx := currentIdx - 1
					if newIdx < 0 {
						newIdx = len(props.Options) - 1 // Wrap to bottom
					}
					selectedIndex.Set(newIdx)
				}
			})

			// Register down navigation event
			ctx.On("down", func(data interface{}) {
				if !props.Disabled && isOpen.GetTyped() && len(props.Options) > 0 {
					currentIdx := selectedIndex.GetTyped()
					newIdx := (currentIdx + 1) % len(props.Options) // Wrap to top
					selectedIndex.Set(newIdx)
				}
			})

			// Register select event (confirm selection)
			ctx.On("select", func(data interface{}) {
				if !props.Disabled && isOpen.GetTyped() && len(props.Options) > 0 {
					idx := selectedIndex.GetTyped()
					if idx >= 0 && idx < len(props.Options) {
						selectedValue := props.Options[idx]
						props.Value.Set(selectedValue)

						// Call OnChange callback if provided
						if props.OnChange != nil {
							props.OnChange(selectedValue)
						}

						// Close dropdown after selection
						isOpen.Set(false)
					}
				}
			})

			// Register close event (close without selecting)
			ctx.On("close", func(data interface{}) {
				isOpen.Set(false)
			})

			// Expose internal state
			ctx.Expose("theme", theme)
			ctx.Expose("isOpen", isOpen)
			ctx.Expose("selectedIndex", selectedIndex)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			props := ctx.Props().(SelectProps[T])
			theme := ctx.Get("theme").(Theme)
			isOpen := ctx.Get("isOpen").(*bubbly.Ref[bool])
			selectedIndex := ctx.Get("selectedIndex").(*bubbly.Ref[int])

			// Get current state
			isOpenState := isOpen.GetTyped()
			currentValue := props.Value.GetTyped()
			currentIdx := selectedIndex.GetTyped()

			// Helper function to render an option
			renderOption := func(opt T) string {
				if props.RenderOption != nil {
					return props.RenderOption(opt)
				}
				return fmt.Sprintf("%v", opt)
			}

			// Determine display value
			var displayValue string
			if len(props.Options) == 0 {
				displayValue = props.Placeholder
				if displayValue == "" {
					displayValue = "No options"
				}
			} else {
				// Check if current value is in options
				valueStr := fmt.Sprintf("%v", currentValue)
				found := false
				for _, opt := range props.Options {
					if fmt.Sprintf("%v", opt) == valueStr {
						displayValue = renderOption(currentValue)
						found = true
						break
					}
				}
				if !found && props.Placeholder != "" {
					displayValue = props.Placeholder
				} else if !found {
					displayValue = renderOption(currentValue)
				}
			}

			// Build select style
			selectStyle := lipgloss.NewStyle().
				Padding(0, 1).
				Border(theme.GetBorderStyle())

			// Set color based on state
			if props.Disabled {
				selectStyle = selectStyle.Foreground(theme.Muted)
			} else if isOpenState {
				selectStyle = selectStyle.BorderForeground(theme.Primary)
			} else {
				selectStyle = selectStyle.BorderForeground(theme.Secondary)
			}

			// Apply custom style if provided
			if props.Style != nil {
				selectStyle = selectStyle.Inherit(*props.Style)
			}

			var output strings.Builder

			if isOpenState && !props.Disabled && len(props.Options) > 0 {
				// Open state: show options list
				indicator := "▲"
				output.WriteString(selectStyle.Render(indicator + " " + displayValue))
				output.WriteString("\n")

				// Render options
				for i, opt := range props.Options {
					optionText := renderOption(opt)
					optionStyle := lipgloss.NewStyle().Padding(0, 1)

					if i == currentIdx {
						// Highlighted option
						optionStyle = optionStyle.
							Foreground(theme.Primary).
							Bold(true)
						optionText = "> " + optionText
					} else {
						optionStyle = optionStyle.Foreground(theme.Foreground)
						optionText = "  " + optionText
					}

					output.WriteString(optionStyle.Render(optionText))
					output.WriteString("\n")
				}
			} else {
				// Closed state: show selected value
				indicator := "▼"
				output.WriteString(selectStyle.Render(indicator + " " + displayValue))
			}

			return output.String()
		}).
		Build()

	return component
}
