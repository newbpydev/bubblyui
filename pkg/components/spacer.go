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
//	spacer := components.Spacer(components.SpacerProps{
//	    Width:  20,
//	    Height: 3,
//	})
type SpacerProps struct {
	// Width sets the horizontal space in characters.
	// Optional - if 0, no horizontal space is added.
	Width int

	// Height sets the vertical space in lines.
	// Optional - if 0, no vertical space is added.
	Height int

	// Common props for all components
	CommonProps
}

// Spacer creates a new Spacer atom component.
//
// Spacer is a layout utility component that creates empty space in the terminal.
// It can create horizontal space (width), vertical space (height), or both.
//
// Example:
//
//	// Horizontal spacer
//	hSpacer := components.Spacer(components.SpacerProps{
//	    Width: 10,
//	})
//
//	// Vertical spacer
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
//	// Initialize and use with Bubbletea
//	spacer.Init()
//	view := spacer.View()
//
// Use cases:
//   - Creating margins between components
//   - Adding padding in layouts
//   - Vertical spacing between sections
//   - Horizontal spacing in rows
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
