package bubbly

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

// defaultRenderer is a shared Lipgloss renderer instance used by all components.
// It's initialized once and reused for performance.
var defaultRenderer = lipgloss.NewRenderer(os.Stdout)

// NewRenderer returns a Lipgloss renderer for the component.
// This allows templates to create custom renderers if needed for specific
// output destinations (e.g., SSH sessions, files).
//
// For most use cases, using NewStyle() directly is simpler and more efficient
// as it uses a shared default renderer.
//
// Example:
//
//	Template(func(ctx RenderContext) string {
//	    renderer := ctx.NewRenderer()
//	    style := renderer.NewStyle().Bold(true)
//	    return style.Render("Custom renderer")
//	})
func (ctx RenderContext) NewRenderer() *lipgloss.Renderer {
	return defaultRenderer
}

// NewStyle creates a new Lipgloss style for use in templates.
// This is the primary way to apply styling to rendered text.
//
// The style can be configured with various properties like colors, padding,
// margins, borders, alignment, and more. Styles are composable and can be
// inherited from other styles.
//
// Example:
//
//	Template(func(ctx RenderContext) string {
//	    style := ctx.NewStyle().
//	        Bold(true).
//	        Foreground(lipgloss.Color("63")).
//	        Padding(1, 2)
//	    return style.Render("Styled text")
//	})
//
// For more complex styling patterns:
//
//	Template(func(ctx RenderContext) string {
//	    // Base style
//	    baseStyle := ctx.NewStyle().
//	        Foreground(lipgloss.Color("229")).
//	        Padding(1)
//
//	    // Inherit and extend
//	    headerStyle := ctx.NewStyle().
//	        Bold(true).
//	        Inherit(baseStyle)
//
//	    return headerStyle.Render("Header")
//	})
func (ctx RenderContext) NewStyle() lipgloss.Style {
	return defaultRenderer.NewStyle()
}
