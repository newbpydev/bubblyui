package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// AppLayoutProps defines the properties for the AppLayout component.
// AppLayout is a template component that provides a full application layout structure.
type AppLayoutProps struct {
	// Header is the top section component (optional).
	// Typically contains application title, navigation, or branding.
	Header bubbly.Component

	// Sidebar is the left section component (optional).
	// Typically contains navigation menu or secondary content.
	Sidebar bubbly.Component

	// Content is the main center section component (required).
	// Contains the primary application content.
	Content bubbly.Component

	// Footer is the bottom section component (optional).
	// Typically contains copyright, links, or status information.
	Footer bubbly.Component

	// Width sets the total layout width in characters.
	// Default is 80 if not specified.
	Width int

	// Height sets the total layout height in lines.
	// Default is 24 if not specified.
	Height int

	// SidebarWidth sets the sidebar width in characters.
	// Default is 20 if not specified.
	SidebarWidth int

	// HeaderHeight sets the header height in lines.
	// Default is 3 if not specified.
	HeaderHeight int

	// FooterHeight sets the footer height in lines.
	// Default is 2 if not specified.
	FooterHeight int

	// CommonProps for styling and identification.
	CommonProps
}

// AppLayout creates a full application layout template component.
// The layout positions Header, Sidebar, Content, and Footer sections using Lipgloss layout functions.
//
// Layout Structure:
//
//	┌─────────────────────────────────┐
//	│          Header (full width)    │
//	├──────────┬──────────────────────┤
//	│ Sidebar  │      Content         │
//	│          │                      │
//	├──────────┴──────────────────────┤
//	│          Footer (full width)    │
//	└─────────────────────────────────┘
//
// Features:
//   - Full application layout with four sections
//   - Responsive to terminal size
//   - Configurable dimensions for each section
//   - Theme integration for consistent styling
//   - Custom style override support
//   - Lipgloss-based layout positioning
//
// Example:
//
//	layout := AppLayout(AppLayoutProps{
//	    Header:  Text(TextProps{Content: "My App"}),
//	    Sidebar: Menu(MenuProps{Items: menuItems}),
//	    Content: Card(CardProps{Title: "Dashboard"}),
//	    Footer:  Text(TextProps{Content: "© 2024"}),
//	})
func AppLayout(props AppLayoutProps) bubbly.Component {
	// Set defaults
	if props.Width == 0 {
		props.Width = 80
	}
	if props.Height == 0 {
		props.Height = 24
	}
	if props.SidebarWidth == 0 {
		props.SidebarWidth = 20
	}
	if props.HeaderHeight == 0 {
		props.HeaderHeight = 3
	}
	if props.FooterHeight == 0 {
		props.FooterHeight = 2
	}

	component, _ := bubbly.NewComponent("AppLayout").
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
			p := ctx.Props().(AppLayoutProps)
			theme := ctx.Get("theme").(Theme)

			var output strings.Builder

			// Calculate dimensions
			contentHeight := p.Height - p.HeaderHeight - p.FooterHeight
			contentWidth := p.Width - p.SidebarWidth

			// Render Header (full width)
			if p.Header != nil {
				headerStyle := lipgloss.NewStyle().
					Width(p.Width).
					Height(p.HeaderHeight).
					Border(lipgloss.NormalBorder(), false, false, true, false).
					BorderForeground(theme.Secondary)

				headerContent := p.Header.View()
				output.WriteString(headerStyle.Render(headerContent))
				output.WriteString("\n")
			}

			// Render Main Area (Sidebar + Content)
			var mainArea string

			// Sidebar section
			if p.Sidebar != nil {
				sidebarStyle := lipgloss.NewStyle().
					Width(p.SidebarWidth).
					Height(contentHeight).
					Border(lipgloss.NormalBorder(), false, true, false, false).
					BorderForeground(theme.Secondary).
					Padding(1)

				sidebarContent := p.Sidebar.View()
				sidebarRendered := sidebarStyle.Render(sidebarContent)

				// Content section
				var contentRendered string
				if p.Content != nil {
					contentStyle := lipgloss.NewStyle().
						Width(contentWidth).
						Height(contentHeight).
						Padding(1, 2)

					contentContent := p.Content.View()
					contentRendered = contentStyle.Render(contentContent)
				} else {
					// Empty content placeholder
					contentStyle := lipgloss.NewStyle().
						Width(contentWidth).
						Height(contentHeight)
					contentRendered = contentStyle.Render("")
				}

				// Join sidebar and content horizontally
				mainArea = lipgloss.JoinHorizontal(
					lipgloss.Top,
					sidebarRendered,
					contentRendered,
				)
			} else {
				// No sidebar, content takes full width
				if p.Content != nil {
					contentStyle := lipgloss.NewStyle().
						Width(p.Width).
						Height(contentHeight).
						Padding(1, 2)

					contentContent := p.Content.View()
					mainArea = contentStyle.Render(contentContent)
				} else {
					// Empty content placeholder
					contentStyle := lipgloss.NewStyle().
						Width(p.Width).
						Height(contentHeight)
					mainArea = contentStyle.Render("")
				}
			}

			output.WriteString(mainArea)

			// Render Footer (full width)
			if p.Footer != nil {
				output.WriteString("\n")
				footerStyle := lipgloss.NewStyle().
					Width(p.Width).
					Height(p.FooterHeight).
					Border(lipgloss.NormalBorder(), true, false, false, false).
					BorderForeground(theme.Secondary)

				footerContent := p.Footer.View()
				output.WriteString(footerStyle.Render(footerContent))
			}

			// Apply custom style if provided
			result := output.String()
			if p.Style != nil {
				containerStyle := lipgloss.NewStyle().Inherit(*p.Style)
				result = containerStyle.Render(result)
			}

			return result
		}).
		Build()

	return component
}
