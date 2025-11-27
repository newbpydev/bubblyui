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
//
// radioFindInitialIndex finds the initial highlighted index based on current value.
func radioFindInitialIndex[T any](options []T, currentValue T) int {
	for i, opt := range options {
		if fmt.Sprintf("%v", opt) == fmt.Sprintf("%v", currentValue) {
			return i
		}
	}
	return 0
}

// radioHandleNavigation creates a navigation handler for up/down movement.
func radioHandleNavigation[T any](props RadioProps[T], highlightedIndex *bubbly.Ref[int], delta int) func(interface{}) {
	return func(_ interface{}) {
		if props.Disabled || len(props.Options) == 0 {
			return
		}
		currentIdx := highlightedIndex.GetTyped()
		newIdx := currentIdx + delta
		if newIdx < 0 {
			newIdx = len(props.Options) - 1
		} else if newIdx >= len(props.Options) {
			newIdx = 0
		}
		highlightedIndex.Set(newIdx)
	}
}

// radioHandleSelect creates a select handler for confirming selection.
func radioHandleSelect[T any](props RadioProps[T], highlightedIndex *bubbly.Ref[int]) func(interface{}) {
	return func(_ interface{}) {
		if props.Disabled || len(props.Options) == 0 {
			return
		}
		idx := highlightedIndex.GetTyped()
		if idx >= 0 && idx < len(props.Options) {
			selectedValue := props.Options[idx]
			props.Value.Set(selectedValue)
			if props.OnChange != nil {
				props.OnChange(selectedValue)
			}
		}
	}
}

// radioGetOptionStyle returns the style for a radio option based on state.
func radioGetOptionStyle[T any](opt T, index int, currentValue T, currentIdx int, props RadioProps[T], theme Theme) lipgloss.Style {
	optionStyle := lipgloss.NewStyle()
	isSelected := fmt.Sprintf("%v", opt) == fmt.Sprintf("%v", currentValue)

	if props.Disabled {
		optionStyle = optionStyle.Foreground(theme.Muted)
	} else if index == currentIdx {
		optionStyle = optionStyle.Foreground(theme.Primary).Bold(true)
	} else if isSelected {
		optionStyle = optionStyle.Foreground(theme.Primary)
	} else {
		optionStyle = optionStyle.Foreground(theme.Foreground)
	}

	if props.Style != nil {
		optionStyle = optionStyle.Inherit(*props.Style)
	}
	return optionStyle
}

// radioRenderOption renders a single radio option with indicator.
func radioRenderOption[T any](opt T, _ int, currentValue T, props RadioProps[T]) string {
	var optionText string
	if props.RenderOption != nil {
		optionText = props.RenderOption(opt)
	} else {
		optionText = fmt.Sprintf("%v", opt)
	}

	isSelected := fmt.Sprintf("%v", opt) == fmt.Sprintf("%v", currentValue)
	indicator := "( )"
	if isSelected {
		indicator = "(●)"
	}
	return indicator + " " + optionText
}

func Radio[T any](props RadioProps[T]) bubbly.Component {
	component, _ := bubbly.NewComponent("Radio").
		Props(props).
		Setup(func(ctx *bubbly.Context) {
			theme := injectTheme(ctx)
			highlightedIndex := bubbly.NewRef(radioFindInitialIndex(props.Options, props.Value.GetTyped()))

			ctx.On("up", radioHandleNavigation(props, highlightedIndex, -1))
			ctx.On("down", radioHandleNavigation(props, highlightedIndex, 1))
			ctx.On("select", radioHandleSelect(props, highlightedIndex))

			ctx.Expose("theme", theme)
			ctx.Expose("highlightedIndex", highlightedIndex)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			props := ctx.Props().(RadioProps[T])
			theme := ctx.Get("theme").(Theme)
			highlightedIndex := ctx.Get("highlightedIndex").(*bubbly.Ref[int])

			if len(props.Options) == 0 {
				return lipgloss.NewStyle().Foreground(theme.Muted).Render("No options available")
			}

			currentValue := props.Value.GetTyped()
			currentIdx := highlightedIndex.GetTyped()

			var output strings.Builder
			for i, opt := range props.Options {
				line := radioRenderOption(opt, i, currentValue, props)
				optionStyle := radioGetOptionStyle(opt, i, currentValue, currentIdx, props, theme)
				output.WriteString(optionStyle.Render(line))

				if i < len(props.Options)-1 {
					output.WriteString("\n")
				}
			}

			return output.String()
		}).
		Build()

	return component
}
