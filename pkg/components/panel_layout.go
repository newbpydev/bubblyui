package components

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// PanelLayoutProps defines the properties for the PanelLayout component.
// PanelLayout is a template component that provides a split panel layout.
type PanelLayoutProps struct {
	// Left is the left panel component (or top panel in vertical mode).
	// Optional - if empty, only Right panel is displayed.
	Left bubbly.Component

	// Right is the right panel component (or bottom panel in vertical mode).
	// Optional - if empty, only Left panel is displayed.
	Right bubbly.Component

	// Direction sets the split direction: "horizontal" (left/right) or "vertical" (top/bottom).
	// Default is "horizontal" if not specified.
	Direction string

	// SplitRatio sets the ratio between left/right (or top/bottom) panels.
	// Value between 0.0 and 1.0. Default is 0.5 (50/50 split).
	// For example, 0.3 means 30% left, 70% right.
	SplitRatio float64

	// Width sets the total layout width in characters.
	// Default is 80 if not specified.
	Width int

	// Height sets the total layout height in lines.
	// Default is 24 if not specified.
	Height int

	// ShowBorder enables borders around panels.
	// Default is false.
	ShowBorder bool

	// CommonProps for styling and identification.
	CommonProps
}

// PanelLayout creates a split panel layout template component.
// The layout splits the space either horizontally (left/right) or vertically (top/bottom).
//
// Layout Structure (Horizontal):
//
//	┌──────────┬──────────────────────┐
//	│          │                      │
//	│   Left   │       Right          │
//	│          │                      │
//	└──────────┴──────────────────────┘
//
// Layout Structure (Vertical):
//
//	┌─────────────────────────────────┐
//	│             Top                 │
//	├─────────────────────────────────┤
//	│            Bottom               │
//	└─────────────────────────────────┘
//
// Features:
//   - Horizontal or vertical split panels
//   - Configurable split ratio
//   - Optional borders
//   - Theme integration for consistent styling
//   - Custom style override support
//   - Perfect for master-detail views
//
// Example:
//
//	layout := PanelLayout(PanelLayoutProps{
//	    Left:       List(ListProps{Items: items}),
//	    Right:      Card(CardProps{Title: "Details"}),
//	    SplitRatio: 0.3, // 30% left, 70% right
//	})
func PanelLayout(props PanelLayoutProps) bubbly.Component {
	// Set defaults
	if props.Width == 0 {
		props.Width = 80
	}
	if props.Height == 0 {
		props.Height = 24
	}
	if props.SplitRatio == 0 {
		props.SplitRatio = 0.5
	}
	if props.Direction == "" {
		props.Direction = "horizontal"
	}

	component, _ := bubbly.NewComponent("PanelLayout").
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
			p := ctx.Props().(PanelLayoutProps)
			theme := ctx.Get("theme").(Theme)

			var result string

			if p.Direction == "vertical" {
				// Vertical split (top/bottom)
				topHeight := int(float64(p.Height) * p.SplitRatio)
				bottomHeight := p.Height - topHeight

				// Render top panel (using Left prop)
				var topPanel string
				if p.Left != nil {
					topStyle := lipgloss.NewStyle().
						Width(p.Width).
						Height(topHeight).
						Padding(1)

					if p.ShowBorder {
						topStyle = topStyle.
							Border(lipgloss.NormalBorder(), false, false, true, false).
							BorderForeground(theme.Secondary)
					}

					topContent := p.Left.View()
					topPanel = topStyle.Render(topContent)
				} else {
					topStyle := lipgloss.NewStyle().
						Width(p.Width).
						Height(topHeight)
					topPanel = topStyle.Render("")
				}

				// Render bottom panel (using Right prop)
				var bottomPanel string
				if p.Right != nil {
					bottomStyle := lipgloss.NewStyle().
						Width(p.Width).
						Height(bottomHeight).
						Padding(1)

					if p.ShowBorder {
						bottomStyle = bottomStyle.
							Border(lipgloss.NormalBorder()).
							BorderForeground(theme.Secondary)
					}

					bottomContent := p.Right.View()
					bottomPanel = bottomStyle.Render(bottomContent)
				} else {
					bottomStyle := lipgloss.NewStyle().
						Width(p.Width).
						Height(bottomHeight)
					bottomPanel = bottomStyle.Render("")
				}

				// Join vertically
				result = lipgloss.JoinVertical(
					lipgloss.Left,
					topPanel,
					bottomPanel,
				)
			} else {
				// Horizontal split (left/right)
				leftWidth := int(float64(p.Width) * p.SplitRatio)
				rightWidth := p.Width - leftWidth

				// Render left panel
				var leftPanel string
				if p.Left != nil {
					leftStyle := lipgloss.NewStyle().
						Width(leftWidth).
						Height(p.Height).
						Padding(1)

					if p.ShowBorder {
						leftStyle = leftStyle.
							Border(lipgloss.NormalBorder(), false, true, false, false).
							BorderForeground(theme.Secondary)
					}

					leftContent := p.Left.View()
					leftPanel = leftStyle.Render(leftContent)
				} else {
					leftStyle := lipgloss.NewStyle().
						Width(leftWidth).
						Height(p.Height)
					leftPanel = leftStyle.Render("")
				}

				// Render right panel
				var rightPanel string
				if p.Right != nil {
					rightStyle := lipgloss.NewStyle().
						Width(rightWidth).
						Height(p.Height).
						Padding(1)

					if p.ShowBorder {
						rightStyle = rightStyle.
							Border(lipgloss.NormalBorder()).
							BorderForeground(theme.Secondary)
					}

					rightContent := p.Right.View()
					rightPanel = rightStyle.Render(rightContent)
				} else {
					rightStyle := lipgloss.NewStyle().
						Width(rightWidth).
						Height(p.Height)
					rightPanel = rightStyle.Render("")
				}

				// Join horizontally
				result = lipgloss.JoinHorizontal(
					lipgloss.Top,
					leftPanel,
					rightPanel,
				)
			}

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
