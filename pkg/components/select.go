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

	// Width sets the width of the select in characters.
	// Optional - if 0, defaults to 30 characters.
	Width int

	// NoBorder removes the border if true.
	// Default is false (border is shown).
	// Useful when embedding in other bordered containers.
	NoBorder bool

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
//
// selectHandleToggle handles the toggle event for opening/closing the dropdown.
func selectHandleToggle[T any](props SelectProps[T], isOpen *bubbly.Ref[bool]) func(interface{}) {
	return func(data interface{}) {
		if !props.Disabled {
			isOpen.Set(!isOpen.GetTyped())
		}
	}
}

// selectHandleNavigation handles up/down navigation in the dropdown.
func selectHandleNavigation[T any](props SelectProps[T], isOpen *bubbly.Ref[bool], selectedIndex *bubbly.Ref[int], delta int) func(interface{}) {
	return func(data interface{}) {
		if !props.Disabled && isOpen.GetTyped() && len(props.Options) > 0 {
			currentIdx := selectedIndex.GetTyped()
			newIdx := currentIdx + delta
			if newIdx < 0 {
				newIdx = len(props.Options) - 1 // Wrap to bottom
			} else if newIdx >= len(props.Options) {
				newIdx = 0 // Wrap to top
			}
			selectedIndex.Set(newIdx)
		}
	}
}

// selectHandleSelect handles the select event for confirming selection.
func selectHandleSelect[T any](props SelectProps[T], isOpen *bubbly.Ref[bool], selectedIndex *bubbly.Ref[int]) func(interface{}) {
	return func(data interface{}) {
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
	}
}

// selectFindInitialIndex finds the initial selected index based on current value.
func selectFindInitialIndex[T any](options []T, currentValue T) int {
	for i, opt := range options {
		if fmt.Sprintf("%v", opt) == fmt.Sprintf("%v", currentValue) {
			return i
		}
	}
	return 0
}

// selectRenderOptionText renders an option as text.
func selectRenderOptionText[T any](opt T, props SelectProps[T]) string {
	if props.RenderOption != nil {
		return props.RenderOption(opt)
	}
	return fmt.Sprintf("%v", opt)
}

// selectGetDisplayValue determines the display value for the select.
func selectGetDisplayValue[T any](props SelectProps[T]) string {
	if len(props.Options) == 0 {
		if props.Placeholder != "" {
			return props.Placeholder
		}
		return "No options"
	}

	currentValue := props.Value.GetTyped()
	valueStr := fmt.Sprintf("%v", currentValue)
	for _, opt := range props.Options {
		if fmt.Sprintf("%v", opt) == valueStr {
			return selectRenderOptionText(currentValue, props)
		}
	}

	if props.Placeholder != "" {
		return props.Placeholder
	}
	return selectRenderOptionText(currentValue, props)
}

// selectApplyBorderColor applies border color based on state.
func selectApplyBorderColor[T any](style lipgloss.Style, props SelectProps[T], isOpen bool, theme Theme) lipgloss.Style {
	if props.NoBorder {
		return style
	}
	if props.Disabled {
		return style.BorderForeground(theme.Muted)
	}
	if isOpen {
		return style.BorderForeground(theme.Primary)
	}
	return style.BorderForeground(theme.Secondary)
}

// selectRenderOptions renders the dropdown options list.
func selectRenderOptions[T any](props SelectProps[T], currentIdx int, theme Theme) string {
	var output strings.Builder
	for i, opt := range props.Options {
		optionText := selectRenderOptionText(opt, props)
		optionStyle := lipgloss.NewStyle().Padding(0, 1)

		if i == currentIdx {
			optionStyle = optionStyle.Foreground(theme.Primary).Bold(true)
			optionText = "> " + optionText
		} else {
			optionStyle = optionStyle.Foreground(theme.Foreground)
			optionText = "  " + optionText
		}

		output.WriteString(optionStyle.Render(optionText))
		output.WriteString("\n")
	}
	return output.String()
}

func Select[T any](props SelectProps[T]) bubbly.Component {
	component, _ := bubbly.NewComponent("Select").
		Props(props).
		Setup(func(ctx *bubbly.Context) {
			theme := injectTheme(ctx)
			isOpen := bubbly.NewRef(false)
			selectedIndex := bubbly.NewRef(selectFindInitialIndex(props.Options, props.Value.GetTyped()))

			ctx.On("toggle", selectHandleToggle(props, isOpen))
			ctx.On("up", selectHandleNavigation(props, isOpen, selectedIndex, -1))
			ctx.On("down", selectHandleNavigation(props, isOpen, selectedIndex, 1))
			ctx.On("select", selectHandleSelect(props, isOpen, selectedIndex))
			ctx.On("close", func(_ interface{}) { isOpen.Set(false) })

			ctx.Expose("theme", theme)
			ctx.Expose("isOpen", isOpen)
			ctx.Expose("selectedIndex", selectedIndex)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			props := ctx.Props().(SelectProps[T])
			theme := ctx.Get("theme").(Theme)
			isOpen := ctx.Get("isOpen").(*bubbly.Ref[bool])
			selectedIndex := ctx.Get("selectedIndex").(*bubbly.Ref[int])

			isOpenState := isOpen.GetTyped()
			currentIdx := selectedIndex.GetTyped()
			displayValue := selectGetDisplayValue(props)

			width := props.Width
			if width <= 0 {
				width = 30
			}

			selectStyle := lipgloss.NewStyle().Padding(0, 1).Width(width)
			if !props.NoBorder {
				selectStyle = selectStyle.Border(theme.GetBorderStyle())
			}
			if props.Disabled {
				selectStyle = selectStyle.Foreground(theme.Muted)
			}
			selectStyle = selectApplyBorderColor(selectStyle, props, isOpenState, theme)

			if props.Style != nil {
				selectStyle = selectStyle.Inherit(*props.Style)
			}

			var output strings.Builder
			if isOpenState && !props.Disabled && len(props.Options) > 0 {
				output.WriteString(selectStyle.Render("▲ " + displayValue))
				output.WriteString("\n")
				output.WriteString(selectRenderOptions(props, currentIdx, theme))
			} else {
				output.WriteString(selectStyle.Render("▼ " + displayValue))
			}

			return output.String()
		}).
		Build()

	return component
}
