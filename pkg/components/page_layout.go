package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// PageLayoutProps defines the properties for the PageLayout component.
// PageLayout is a template component that provides a simple page structure.
type PageLayoutProps struct {
	// Title is the page title section component (optional).
	// Typically contains page heading or breadcrumbs.
	Title bubbly.Component

	// Content is the main page content section component (required).
	// Contains the primary page content.
	Content bubbly.Component

	// Actions is the page actions section component (optional).
	// Typically contains buttons or action controls at the bottom.
	Actions bubbly.Component

	// Width sets the total layout width in characters.
	// Default is 80 if not specified.
	Width int

	// Spacing sets the vertical spacing between sections in lines.
	// Default is 2 if not specified.
	Spacing int

	// CommonProps for styling and identification.
	CommonProps
}

// PageLayout creates a simple page structure template component.
// The layout positions Title, Content, and Actions sections vertically using Lipgloss layout functions.
//
// Layout Structure:
//
//	┌─────────────────────────────────┐
//	│          Title                  │
//	│                                 │
//	│          Content                │
//	│          (main area)            │
//	│                                 │
//	│          Actions                │
//	└─────────────────────────────────┘
//
// Features:
//   - Simple vertical page structure with three sections
//   - Configurable width and spacing
//   - Theme integration for consistent styling
//   - Custom style override support
//   - Lipgloss-based layout positioning
//
// Example:
//
//	layout := PageLayout(PageLayoutProps{
//	    Title:   Text(TextProps{Content: "Dashboard"}),
//	    Content: Card(CardProps{Title: "Stats"}),
//	    Actions: Button(ButtonProps{Label: "Refresh"}),
//	})
func PageLayout(props PageLayoutProps) bubbly.Component {
	// Set defaults
	if props.Width == 0 {
		props.Width = 80
	}
	if props.Spacing == 0 {
		props.Spacing = 2
	}

	component, _ := bubbly.NewComponent("PageLayout").
		Props(props).
		Setup(func(ctx *bubbly.Context) {
			// Inject theme
			theme := DefaultTheme
			if injected := ctx.Inject("theme", nil); injected != nil {
				if t, ok := injected.(Theme); ok {
					theme = t
				}
			}
			ctx.Expose("theme", theme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			p := ctx.Props().(PageLayoutProps)
			theme := ctx.Get("theme").(Theme)

			var sections []string

			// Render Title section
			if p.Title != nil {
				titleStyle := lipgloss.NewStyle().
					Width(p.Width).
					Bold(true).
					Foreground(theme.Primary).
					Padding(1, 2)

				titleContent := p.Title.View()
				sections = append(sections, titleStyle.Render(titleContent))
			}

			// Render Content section
			if p.Content != nil {
				contentStyle := lipgloss.NewStyle().
					Width(p.Width).
					Padding(1, 2)

				contentContent := p.Content.View()
				sections = append(sections, contentStyle.Render(contentContent))
			}

			// Render Actions section
			if p.Actions != nil {
				actionsStyle := lipgloss.NewStyle().
					Width(p.Width).
					Padding(1, 2).
					Align(lipgloss.Right)

				actionsContent := p.Actions.View()
				sections = append(sections, actionsStyle.Render(actionsContent))
			}

			// Join sections vertically with spacing
			spacer := strings.Repeat("\n", p.Spacing)
			result := strings.Join(sections, spacer)

			// Apply custom style if provided
			if p.Style != nil {
				containerStyle := lipgloss.NewStyle().Inherit(*p.Style)
				result = containerStyle.Render(result)
			}

			return result
		}).
		Build()

	return component
}
