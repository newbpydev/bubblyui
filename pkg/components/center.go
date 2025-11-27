// Package components provides layout components for the BubblyUI framework.
package components

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// CenterProps defines the properties for the Center layout component.
// Center is a molecule component that centers its child component horizontally
// and/or vertically within a container of specified dimensions.
//
// By default, when neither Horizontal nor Vertical flags are set, the component
// centers content in both directions. Setting either flag to true enables
// centering only in that specific direction.
type CenterProps struct {
	// Child is the component to center within the container.
	// If nil, the Center component renders an empty container.
	Child bubbly.Component

	// Width sets the container width in characters.
	// Default is 0 (auto-width based on content).
	// Horizontal centering requires a Width > 0.
	Width int

	// Height sets the container height in lines.
	// Default is 0 (auto-height based on content).
	// Vertical centering requires a Height > 0.
	Height int

	// Horizontal enables horizontal-only centering when true.
	// When false and Vertical is also false, centers both directions (default).
	// Default is false.
	Horizontal bool

	// Vertical enables vertical-only centering when true.
	// When false and Horizontal is also false, centers both directions (default).
	// Default is false.
	Vertical bool

	// CommonProps for styling and identification.
	CommonProps
}

// centerShouldCenterHorizontal determines if horizontal centering should be applied.
// Returns true if:
// - Horizontal flag is explicitly set, OR
// - Both Horizontal and Vertical flags are false (default: center both)
func centerShouldCenterHorizontal(p CenterProps) bool {
	// Default behavior: center both when neither flag is set
	if !p.Horizontal && !p.Vertical {
		return true
	}
	return p.Horizontal
}

// centerShouldCenterVertical determines if vertical centering should be applied.
// Returns true if:
// - Vertical flag is explicitly set, OR
// - Both Horizontal and Vertical flags are false (default: center both)
func centerShouldCenterVertical(p CenterProps) bool {
	// Default behavior: center both when neither flag is set
	if !p.Horizontal && !p.Vertical {
		return true
	}
	return p.Vertical
}

// centerRenderContent renders the child component and applies centering.
func centerRenderContent(p CenterProps) string {
	// Handle nil child
	if p.Child == nil {
		return centerRenderEmpty(p)
	}

	// Render child content
	content := p.Child.View()

	// Determine centering behavior
	centerH := centerShouldCenterHorizontal(p)
	centerV := centerShouldCenterVertical(p)

	// Apply centering based on flags and dimensions
	return centerApplyPlacement(content, p, centerH, centerV)
}

// centerRenderEmpty handles rendering when child is nil.
func centerRenderEmpty(p CenterProps) string {
	if p.Width > 0 && p.Height > 0 {
		return lipgloss.Place(p.Width, p.Height, lipgloss.Center, lipgloss.Center, "")
	}
	return ""
}

// centerApplyPlacement applies the appropriate centering based on flags and dimensions.
func centerApplyPlacement(content string, p CenterProps, centerH, centerV bool) string {
	// Center both directions when both flags and dimensions are set
	if centerH && centerV && p.Width > 0 && p.Height > 0 {
		return lipgloss.Place(p.Width, p.Height, lipgloss.Center, lipgloss.Center, content)
	}

	// Center horizontally only
	if centerH && p.Width > 0 {
		return centerApplyHorizontal(content, p)
	}

	// Center vertically only
	if centerV && p.Height > 0 {
		return centerApplyVertical(content, p)
	}

	// No centering possible (no dimensions specified)
	return content
}

// centerApplyHorizontal applies horizontal centering with optional vertical placement.
func centerApplyHorizontal(content string, p CenterProps) string {
	result := lipgloss.PlaceHorizontal(p.Width, lipgloss.Center, content)
	if p.Height > 0 {
		result = lipgloss.PlaceVertical(p.Height, lipgloss.Top, result)
	}
	return result
}

// centerApplyVertical applies vertical centering with optional horizontal placement.
func centerApplyVertical(content string, p CenterProps) string {
	result := lipgloss.PlaceVertical(p.Height, lipgloss.Center, content)
	if p.Width > 0 {
		result = lipgloss.PlaceHorizontal(p.Width, lipgloss.Left, result)
	}
	return result
}

// Center creates a centering layout component.
// The component centers its child horizontally and/or vertically within
// a container of specified dimensions.
//
// Features:
//   - Centers content horizontally within container
//   - Centers content vertically within container
//   - Centers both directions by default
//   - Supports fixed or auto dimensions
//   - Theme integration for consistent styling
//   - Custom style override support
//
// Centering Behavior:
//   - When neither Horizontal nor Vertical is set: centers both directions (default)
//   - When Horizontal=true: centers only horizontally
//   - When Vertical=true: centers only vertically
//   - When both are true: centers both directions
//
// Dimension Requirements:
//   - Horizontal centering requires Width > 0
//   - Vertical centering requires Height > 0
//   - Auto-sizing (0) uses content dimensions
//
// Example:
//
//	// Center both directions (default)
//	center := Center(CenterProps{
//	    Child:  myComponent,
//	    Width:  80,
//	    Height: 24,
//	})
//
//	// Center horizontally only
//	center := Center(CenterProps{
//	    Child:      myComponent,
//	    Width:      80,
//	    Horizontal: true,
//	})
//
//	// Center vertically only
//	center := Center(CenterProps{
//	    Child:    myComponent,
//	    Height:   24,
//	    Vertical: true,
//	})
//
//	// Modal centering pattern
//	modal := Center(CenterProps{
//	    Child: Card(CardProps{
//	        Title:   "Confirm",
//	        Content: "Are you sure?",
//	    }),
//	    Width:  80,
//	    Height: 24,
//	})
//
//nolint:dupl // Component creation pattern is intentionally similar across all components
func Center(props CenterProps) bubbly.Component {
	component, _ := bubbly.NewComponent("Center").
		Props(props).
		Setup(func(ctx *bubbly.Context) {
			theme := injectTheme(ctx)
			ctx.Expose("theme", theme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			p := ctx.Props().(CenterProps)

			// Render centered content
			result := centerRenderContent(p)

			// Apply custom style if provided
			if p.Style != nil {
				style := lipgloss.NewStyle().Inherit(*p.Style)
				return style.Render(result)
			}

			return result
		}).
		Build()

	return component
}
