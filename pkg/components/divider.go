// Package components provides layout components for the BubblyUI framework.
package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// Default divider characters for horizontal and vertical orientations.
const (
	// DefaultHorizontalChar is the default character for horizontal dividers.
	DefaultHorizontalChar = "─"

	// DefaultVerticalChar is the default character for vertical dividers.
	DefaultVerticalChar = "│"

	// DefaultDividerLength is the default length when not specified.
	DefaultDividerLength = 20
)

// DividerProps defines the properties for the Divider component.
// Divider is an atom component that renders a horizontal or vertical
// separator line with optional label text.
type DividerProps struct {
	// Vertical renders a vertical divider if true.
	// Default: false (horizontal)
	Vertical bool

	// Length is the divider length in characters (horizontal) or lines (vertical).
	// Default: 20
	Length int

	// Label is optional text centered on the divider.
	// For horizontal dividers, the label appears in the middle of the line.
	// For vertical dividers, the label appears in the middle row.
	Label string

	// Char is the divider character.
	// Default: "─" (horizontal) or "│" (vertical)
	Char string

	// CommonProps for styling and identification.
	CommonProps
}

// dividerApplyDefaults sets default values for DividerProps.
func dividerApplyDefaults(props *DividerProps) {
	// Set default length
	if props.Length <= 0 {
		props.Length = DefaultDividerLength
	}

	// Set default character based on orientation
	if props.Char == "" {
		if props.Vertical {
			props.Char = DefaultVerticalChar
		} else {
			props.Char = DefaultHorizontalChar
		}
	}
}

// dividerRenderHorizontal renders a horizontal divider line.
func dividerRenderHorizontal(p DividerProps) string {
	if p.Label == "" {
		// No label - just render the line
		return strings.Repeat(p.Char, p.Length)
	}

	// With label - center it on the line
	labelLen := lipgloss.Width(p.Label)

	// If label is longer than or equal to length, just return the label
	if labelLen >= p.Length {
		return p.Label
	}

	// Calculate space for divider characters on each side
	// We add spaces around the label for visual separation
	labelWithSpaces := " " + p.Label + " "
	labelWithSpacesLen := lipgloss.Width(labelWithSpaces)

	// If label with spaces is too long, reduce padding
	if labelWithSpacesLen >= p.Length {
		labelWithSpaces = p.Label
		labelWithSpacesLen = labelLen
	}

	remainingSpace := p.Length - labelWithSpacesLen
	leftLen := remainingSpace / 2
	rightLen := remainingSpace - leftLen

	leftLine := strings.Repeat(p.Char, leftLen)
	rightLine := strings.Repeat(p.Char, rightLen)

	return leftLine + labelWithSpaces + rightLine
}

// dividerRenderVertical renders a vertical divider line.
func dividerRenderVertical(p DividerProps) string {
	if p.Label == "" {
		// No label - just render vertical lines
		lines := make([]string, p.Length)
		for i := range lines {
			lines[i] = p.Char
		}
		return strings.Join(lines, "\n")
	}

	// With label - center it vertically
	lines := make([]string, p.Length)

	// Calculate the middle position for the label
	middleIndex := p.Length / 2

	for i := range lines {
		if i == middleIndex {
			lines[i] = p.Label
		} else {
			lines[i] = p.Char
		}
	}

	return strings.Join(lines, "\n")
}

// dividerCreateStyle creates the divider's style with theme colors.
func dividerCreateStyle(p DividerProps, theme Theme) lipgloss.Style {
	style := lipgloss.NewStyle().
		Foreground(theme.Muted)

	// Apply custom style if provided
	if p.Style != nil {
		style = style.Inherit(*p.Style)
	}

	return style
}

// Divider creates a separator line component.
// The divider can be horizontal or vertical, with an optional centered label.
//
// Features:
//   - Horizontal or vertical orientation
//   - Configurable length
//   - Optional centered label text
//   - Customizable divider character
//   - Theme integration (uses theme.Muted for color)
//   - Custom style override support
//
// Example:
//
//	// Simple horizontal divider
//	divider := Divider(DividerProps{
//	    Length: 40,
//	})
//
//	// Divider with label
//	divider := Divider(DividerProps{
//	    Label:  "OR",
//	    Length: 30,
//	})
//
//	// Vertical divider
//	divider := Divider(DividerProps{
//	    Vertical: true,
//	    Length:   10,
//	})
//
//	// Custom character
//	divider := Divider(DividerProps{
//	    Char:   "═",
//	    Length: 40,
//	})
//
//nolint:dupl // Component creation pattern is intentionally similar across all components
func Divider(props DividerProps) bubbly.Component {
	dividerApplyDefaults(&props)

	component, _ := bubbly.NewComponent("Divider").
		Props(props).
		Setup(func(ctx *bubbly.Context) {
			theme := injectTheme(ctx)
			ctx.Expose("theme", theme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			p := ctx.Props().(DividerProps)
			theme := ctx.Get("theme").(Theme)

			var content string
			if p.Vertical {
				content = dividerRenderVertical(p)
			} else {
				content = dividerRenderHorizontal(p)
			}

			style := dividerCreateStyle(p, theme)
			return style.Render(content)
		}).
		Build()

	return component
}
