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

// panelLayoutApplyDefaults sets default values for PanelLayoutProps.
func panelLayoutApplyDefaults(props *PanelLayoutProps) {
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
}

// panelLayoutRenderPanel renders a single panel with the given dimensions and options.
func panelLayoutRenderPanel(comp bubbly.Component, width, height int, showBorder bool, theme Theme, borderSides ...bool) string {
	if comp == nil {
		return lipgloss.NewStyle().Width(width).Height(height).Render("")
	}

	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(1)

	if showBorder {
		if len(borderSides) == 4 {
			style = style.Border(lipgloss.NormalBorder(), borderSides[0], borderSides[1], borderSides[2], borderSides[3])
		} else {
			style = style.Border(lipgloss.NormalBorder())
		}
		style = style.BorderForeground(theme.Secondary)
	}

	return style.Render(comp.View())
}

// panelLayoutRenderVertical renders vertical split layout.
func panelLayoutRenderVertical(p PanelLayoutProps, theme Theme) string {
	topHeight := int(float64(p.Height) * p.SplitRatio)
	bottomHeight := p.Height - topHeight

	topPanel := panelLayoutRenderPanel(p.Left, p.Width, topHeight, p.ShowBorder, theme, false, false, true, false)
	bottomPanel := panelLayoutRenderPanel(p.Right, p.Width, bottomHeight, p.ShowBorder, theme)

	return lipgloss.JoinVertical(lipgloss.Left, topPanel, bottomPanel)
}

// panelLayoutRenderHorizontal renders horizontal split layout.
func panelLayoutRenderHorizontal(p PanelLayoutProps, theme Theme) string {
	leftWidth := int(float64(p.Width) * p.SplitRatio)
	rightWidth := p.Width - leftWidth

	leftPanel := panelLayoutRenderPanel(p.Left, leftWidth, p.Height, p.ShowBorder, theme, false, true, false, false)
	rightPanel := panelLayoutRenderPanel(p.Right, rightWidth, p.Height, p.ShowBorder, theme)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
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
	panelLayoutApplyDefaults(&props)

	component, _ := bubbly.NewComponent("PanelLayout").
		Props(props).
		Setup(func(ctx *bubbly.Context) {
			theme := injectTheme(ctx)
			ctx.Expose("theme", theme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			p := ctx.Props().(PanelLayoutProps)
			theme := ctx.Get("theme").(Theme)

			var result string
			if p.Direction == "vertical" {
				result = panelLayoutRenderVertical(p, theme)
			} else {
				result = panelLayoutRenderHorizontal(p, theme)
			}

			if p.Style != nil {
				result = lipgloss.NewStyle().Inherit(*p.Style).Render(result)
			}

			return result
		}).
		Build()

	return component
}
