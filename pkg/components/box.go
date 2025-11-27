// Package components provides layout components for the BubblyUI framework.
package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// BoxProps defines the properties for the Box container component.
// Box is an atom component that provides a generic container with padding,
// border, and title support. It serves as a building block for higher-level layouts.
type BoxProps struct {
	// Child is an optional child component to render inside the box.
	// If provided, it takes precedence over Content.
	Child bubbly.Component

	// Content is the text content to display inside the box.
	// Used when Child is nil.
	Content string

	// Padding sets uniform padding on all sides (in characters/lines).
	// Default is 0 (no padding).
	Padding int

	// PaddingX sets horizontal padding (left and right).
	// If set, overrides Padding for horizontal sides.
	PaddingX int

	// PaddingY sets vertical padding (top and bottom).
	// If set, overrides Padding for vertical sides.
	PaddingY int

	// Border enables a border around the box.
	// Default is false (no border).
	Border bool

	// BorderStyle specifies the border style to use.
	// Default is lipgloss.NormalBorder() when Border is true.
	BorderStyle lipgloss.Border

	// Title is optional text displayed at the top of the box.
	// Rendered as a styled header line inside the box.
	Title string

	// Width sets the fixed width of the box in characters.
	// Default is 0 (auto-width based on content).
	Width int

	// Height sets the fixed height of the box in lines.
	// Default is 0 (auto-height based on content).
	Height int

	// Background sets the background color inside the box.
	// Default is no background color.
	Background lipgloss.Color

	// CommonProps for styling and identification.
	CommonProps
}

// boxApplyDefaults sets default values for BoxProps.
// BorderStyle defaults to NormalBorder when Border is enabled.
func boxApplyDefaults(props *BoxProps) {
	// BorderStyle defaults to NormalBorder when Border is true
	if props.Border && props.BorderStyle.Top == "" {
		props.BorderStyle = lipgloss.NormalBorder()
	}
}

// boxRenderContent builds the box's inner content string.
// It handles Child vs Content precedence and title rendering.
func boxRenderContent(p BoxProps, theme Theme) string {
	var content strings.Builder

	// Render title if present
	if p.Title != "" {
		titleStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Primary)
		content.WriteString(titleStyle.Render(p.Title))
		content.WriteString("\n")
	}

	// Render child or content
	if p.Child != nil {
		content.WriteString(p.Child.View())
	} else if p.Content != "" {
		contentStyle := lipgloss.NewStyle().
			Foreground(theme.Foreground)
		content.WriteString(contentStyle.Render(p.Content))
	}

	return content.String()
}

// boxCreateStyle creates the box's outer style with padding, border, and dimensions.
func boxCreateStyle(p BoxProps, theme Theme) lipgloss.Style {
	style := lipgloss.NewStyle()

	// Apply padding
	// PaddingX/PaddingY override Padding if set
	paddingTop := p.Padding
	paddingRight := p.Padding
	paddingBottom := p.Padding
	paddingLeft := p.Padding

	if p.PaddingY > 0 {
		paddingTop = p.PaddingY
		paddingBottom = p.PaddingY
	}
	if p.PaddingX > 0 {
		paddingLeft = p.PaddingX
		paddingRight = p.PaddingX
	}

	style = style.
		PaddingTop(paddingTop).
		PaddingRight(paddingRight).
		PaddingBottom(paddingBottom).
		PaddingLeft(paddingLeft)

	// Apply border
	if p.Border {
		style = style.
			BorderStyle(p.BorderStyle).
			BorderForeground(theme.Secondary)
	}

	// Apply dimensions
	if p.Width > 0 {
		style = style.Width(p.Width)
	}
	if p.Height > 0 {
		style = style.Height(p.Height)
	}

	// Apply background
	if p.Background != "" {
		style = style.Background(p.Background)
	}

	// Apply custom style if provided
	if p.Style != nil {
		style = style.Inherit(*p.Style)
	}

	return style
}

// Box creates a generic container component.
// The box provides a flexible container with optional padding, border, title,
// and background color. It can contain either a child component or text content.
//
// Features:
//   - Optional child component or text content
//   - Configurable padding (uniform or per-axis)
//   - Optional border with customizable style
//   - Optional title header
//   - Fixed or auto dimensions
//   - Background color support
//   - Theme integration for consistent styling
//   - Custom style override support
//
// Example:
//
//	box := Box(BoxProps{
//	    Content: "Hello, World!",
//	    Padding: 1,
//	    Border:  true,
//	    Title:   "Greeting",
//	})
//
//	// With child component
//	box := Box(BoxProps{
//	    Child:   myComponent,
//	    Border:  true,
//	    Width:   40,
//	})
//
//nolint:dupl // Component creation pattern is intentionally similar across all components
func Box(props BoxProps) bubbly.Component {
	boxApplyDefaults(&props)

	component, _ := bubbly.NewComponent("Box").
		Props(props).
		Setup(func(ctx *bubbly.Context) {
			theme := injectTheme(ctx)
			ctx.Expose("theme", theme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			p := ctx.Props().(BoxProps)
			theme := ctx.Get("theme").(Theme)

			content := boxRenderContent(p, theme)
			boxStyle := boxCreateStyle(p, theme)

			return boxStyle.Render(content)
		}).
		Build()

	return component
}
