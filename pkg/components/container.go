// Package components provides layout components for the BubblyUI framework.
package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// ContainerProps defines the properties for the Container layout component.
// Container is a molecule component that constrains content width and optionally
// centers it horizontally within the available space.
//
// Container is useful for creating readable content layouts by limiting line
// length to comfortable reading widths (typically 60-80 characters).
type ContainerProps struct {
	// Child is the component to render inside the container.
	// If nil, the Container renders an empty space with the specified width.
	Child bubbly.Component

	// Size is a preset container size that determines the max-width.
	// Available sizes: ContainerSm (40), ContainerMd (60), ContainerLg (80),
	// ContainerXl (100), ContainerFull (no constraint).
	// Default is ContainerMd (60 characters).
	Size ContainerSize

	// MaxWidth overrides Size with a custom maximum width in characters.
	// When MaxWidth > 0, it takes precedence over the Size preset.
	// Default is 0 (use Size preset).
	MaxWidth int

	// Centered horizontally centers the content within the container width.
	// When true, content is centered using Lipgloss alignment.
	// Default is true.
	// Use CenteredSet to explicitly disable centering (set Centered=false, CenteredSet=true).
	Centered bool

	// CenteredSet indicates whether Centered was explicitly set.
	// This allows distinguishing between "not set" (default true) and "explicitly false".
	CenteredSet bool

	// CommonProps for styling and identification.
	CommonProps
}

// containerApplyDefaults applies default values to ContainerProps.
// - Size defaults to ContainerMd (60 characters)
// - Centered defaults to true (when CenteredSet is false)
func containerApplyDefaults(p *ContainerProps) {
	// Default size is ContainerMd
	if p.Size == "" {
		p.Size = ContainerMd
	}

	// Default Centered to true if not explicitly set
	if !p.CenteredSet {
		p.Centered = true
	}
}

// containerGetWidth returns the effective width for the container.
// MaxWidth takes precedence over Size preset.
// Returns 0 for ContainerFull (no width constraint).
func containerGetWidth(p ContainerProps) int {
	// MaxWidth overrides Size when > 0
	if p.MaxWidth > 0 {
		return p.MaxWidth
	}

	// Use Size preset width
	return p.Size.Width()
}

// containerRenderContent renders the child component content.
// Returns empty string if child is nil.
func containerRenderContent(p ContainerProps) string {
	if p.Child == nil {
		return ""
	}
	return p.Child.View()
}

// containerApplyWidth applies width constraint to the content.
// If width is 0 (ContainerFull), returns content unchanged.
func containerApplyWidth(content string, width int, centered bool) string {
	// No width constraint for ContainerFull (width=0)
	if width == 0 {
		return content
	}

	// Create style with width constraint
	style := lipgloss.NewStyle().Width(width)

	// Apply horizontal alignment based on Centered flag
	if centered {
		style = style.Align(lipgloss.Center)
	} else {
		style = style.Align(lipgloss.Left)
	}

	return style.Render(content)
}

// containerRenderEmpty renders an empty container with the specified width.
func containerRenderEmpty(width int) string {
	if width == 0 {
		return ""
	}
	// Create empty space with the specified width
	return strings.Repeat(" ", width)
}

// Container creates a width-constrained container component.
// The component limits content width to improve readability and optionally
// centers the content horizontally.
//
// Features:
//   - Preset sizes for common widths (sm=40, md=60, lg=80, xl=100)
//   - Custom max-width override
//   - Horizontal centering (enabled by default)
//   - Full-width mode (no constraint)
//   - Theme integration for consistent styling
//   - Custom style override support
//
// Width Behavior:
//   - Size preset determines default width
//   - MaxWidth > 0 overrides Size preset
//   - ContainerFull disables width constraint
//
// Centering Behavior:
//   - Centered=true (default): content is horizontally centered
//   - Centered=false: content is left-aligned
//
// Example:
//
//	// Default container (60 chars, centered)
//	container := Container(ContainerProps{
//	    Child: myContent,
//	})
//
//	// Large container, not centered
//	container := Container(ContainerProps{
//	    Child:    myContent,
//	    Size:     ContainerLg,
//	    Centered: false,
//	})
//
//	// Custom width
//	container := Container(ContainerProps{
//	    Child:    myContent,
//	    MaxWidth: 50,
//	})
//
//	// Full width (no constraint)
//	container := Container(ContainerProps{
//	    Child: myContent,
//	    Size:  ContainerFull,
//	})
//
//	// Readable content layout pattern
//	page := VStack(StackProps{
//	    Items: []bubbly.Component{
//	        header,
//	        Container(ContainerProps{
//	            Child: article,
//	            Size:  ContainerLg,
//	        }),
//	        footer,
//	    },
//	})
//
//nolint:dupl // Component creation pattern is intentionally similar across all components
func Container(props ContainerProps) bubbly.Component {
	// Apply defaults before building component
	containerApplyDefaults(&props)

	component, _ := bubbly.NewComponent("Container").
		Props(props).
		Setup(func(ctx *bubbly.Context) {
			theme := injectTheme(ctx)
			ctx.Expose("theme", theme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			p := ctx.Props().(ContainerProps)

			// Get effective width
			width := containerGetWidth(p)

			// Render child content
			content := containerRenderContent(p)

			// Handle empty content
			if content == "" {
				return containerRenderEmpty(width)
			}

			// Apply width constraint and centering
			result := containerApplyWidth(content, width, p.Centered)

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
