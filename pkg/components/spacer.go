package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// SpacerProps defines the configuration properties for a Spacer component.
//
// Example usage:
//
//	// Fixed-size spacer
//	spacer := components.Spacer(components.SpacerProps{
//	    Width:  20,
//	    Height: 3,
//	})
//
//	// Flexible spacer (fills available space in parent layout)
//	flexSpacer := components.Spacer(components.SpacerProps{
//	    Flex: true,
//	})
type SpacerProps struct {
	// Flex makes the spacer fill available space in parent layouts.
	// When true, the spacer expands to fill remaining space in HStack/VStack/Flex.
	// Width and Height become minimum dimensions when Flex is true.
	// Default: false (fixed-size spacer)
	Flex bool

	// Width sets the horizontal space in characters.
	// When Flex=false: exact width to render.
	// When Flex=true: minimum width (parent layout may expand beyond this).
	// Optional - if 0, no horizontal space is added (or auto when Flex=true).
	Width int

	// Height sets the vertical space in lines.
	// When Flex=false: exact height to render.
	// When Flex=true: minimum height (parent layout may expand beyond this).
	// Optional - if 0, no vertical space is added (or auto when Flex=true).
	Height int

	// Common props for all components
	CommonProps
}

// IsFlex returns true if this spacer should fill available space.
// Parent layouts (HStack, VStack, Flex) use this to detect flexible spacers
// and distribute remaining space accordingly.
func (p SpacerProps) IsFlex() bool {
	return p.Flex
}

// Spacer creates a new Spacer atom component.
//
// Spacer is a layout utility component that creates empty space in the terminal.
// It can create horizontal space (width), vertical space (height), or both.
// When Flex=true, the spacer acts as a marker for parent layouts (HStack, VStack, Flex)
// to fill remaining available space.
//
// Example:
//
//	// Fixed horizontal spacer
//	hSpacer := components.Spacer(components.SpacerProps{
//	    Width: 10,
//	})
//
//	// Fixed vertical spacer
//	vSpacer := components.Spacer(components.SpacerProps{
//	    Height: 3,
//	})
//
//	// Both dimensions
//	spacer := components.Spacer(components.SpacerProps{
//	    Width:  20,
//	    Height: 5,
//	})
//
//	// Flexible spacer - fills available space in parent layout
//	flexSpacer := components.Spacer(components.SpacerProps{
//	    Flex: true,
//	})
//
//	// Flexible spacer with minimum width
//	flexMinSpacer := components.Spacer(components.SpacerProps{
//	    Flex:  true,
//	    Width: 5, // minimum 5 characters, parent can expand
//	})
//
//	// Initialize and use with Bubbletea
//	spacer.Init()
//	view := spacer.View()
//
// Use cases:
//   - Creating margins between components
//   - Adding padding in layouts
//   - Vertical spacing between sections
//   - Horizontal spacing in rows
//   - Pushing items to opposite ends (Flex=true in HStack)
//   - Creating flexible gaps in toolbars
//
// Accessibility:
//   - Invisible but affects layout
//   - Helps create visual hierarchy
//   - Improves readability
func Spacer(props SpacerProps) bubbly.Component {
	component, _ := bubbly.NewComponent("Spacer").
		Props(props).
		Template(func(ctx bubbly.RenderContext) string {
			props := ctx.Props().(SpacerProps)

			// Build spacer content
			var result strings.Builder

			// Create vertical space (newlines)
			if props.Height > 0 {
				result.WriteString(strings.Repeat("\n", props.Height))
			}

			// Create horizontal space
			if props.Width > 0 {
				// If we have height, we need to add width to each line
				if props.Height > 0 {
					// Replace each newline with spaces + newline
					spaces := strings.Repeat(" ", props.Width)
					lines := make([]string, props.Height)
					for i := range lines {
						lines[i] = spaces
					}
					result.Reset()
					result.WriteString(strings.Join(lines, "\n"))
				} else {
					// Just horizontal space
					result.WriteString(strings.Repeat(" ", props.Width))
				}
			}

			// Apply custom style if provided
			if props.Style != nil {
				style := lipgloss.NewStyle().Inherit(*props.Style)
				return style.Render(result.String())
			}

			return result.String()
		}).
		Build()

	return component
}
