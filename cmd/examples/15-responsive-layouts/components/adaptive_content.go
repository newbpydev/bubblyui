// Package components provides responsive demo components for the layout showcase.
package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	localComposables "github.com/newbpydev/bubblyui/cmd/examples/15-responsive-layouts/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// CreateAdaptiveContent creates content that adapts its layout based on screen size.
// On wide screens: horizontal layout with multiple columns
// On narrow screens: vertical stacked layout
func CreateAdaptiveContent() (bubbly.Component, error) {
	return bubbly.NewComponent("AdaptiveContent").
		Setup(func(ctx *bubbly.Context) {
			ctx.ProvideTheme(bubbly.DefaultTheme)
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)

			windowSize := localComposables.UseSharedWindowSize(ctx)
			ctx.Expose("windowSize", windowSize)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			theme := ctx.Get("theme").(bubbly.Theme)
			ws := ctx.Get("windowSize").(*localComposables.WindowSizeComposable)

			width := ws.Width.GetTyped()
			breakpoint := ws.Breakpoint.GetTyped()

			// Determine layout mode based on breakpoint
			isWide := breakpoint == localComposables.BreakpointLG ||
				breakpoint == localComposables.BreakpointXL
			isMedium := breakpoint == localComposables.BreakpointMD

			// === TITLE ===
			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Primary)

			var layoutMode string
			if isWide {
				layoutMode = "3-Column Horizontal"
			} else if isMedium {
				layoutMode = "2-Column Horizontal"
			} else {
				layoutMode = "Stacked Vertical"
			}

			title := components.Text(components.TextProps{
				Content: titleStyle.Render(fmt.Sprintf("🔄 Adaptive Layout: %s", layoutMode)),
			})
			title.Init()

			// === CONTENT PANELS ===
			panels := []struct {
				title   string
				content string
				icon    string
			}{
				{"Primary", "Main content area that contains the most important information.", "📄"},
				{"Secondary", "Supporting content with additional details and context.", "📋"},
				{"Tertiary", "Extra information, tips, or related content.", "💡"},
			}

			// Calculate panel widths based on layout
			var panelWidth int
			var panelComponents []interface{}

			if isWide {
				// 3-column layout
				panelWidth = (width - 10) / 3
			} else if isMedium {
				// 2-column layout (third panel goes below)
				panelWidth = (width - 8) / 2
			} else {
				// Stacked layout
				panelWidth = width - 6
			}

			if panelWidth < 20 {
				panelWidth = 20
			}

			for _, panel := range panels {
				iconStyle := lipgloss.NewStyle().Foreground(theme.Secondary)
				contentText := iconStyle.Render(panel.icon) + " " + panel.content

				card := components.Card(components.CardProps{
					Title:   panel.title,
					Content: contentText,
					Width:   panelWidth,
				})
				card.Init()
				panelComponents = append(panelComponents, card)
			}

			// === ARRANGE PANELS BASED ON LAYOUT ===
			var contentLayout bubbly.Component

			if isWide {
				// All 3 panels in a row
				contentLayout = components.HStack(components.StackProps{
					Items:   panelComponents,
					Spacing: 1,
				})
			} else if isMedium {
				// First 2 panels in a row, third below
				topRow := components.HStack(components.StackProps{
					Items:   panelComponents[:2],
					Spacing: 1,
				})
				topRow.Init()

				contentLayout = components.VStack(components.StackProps{
					Items:   []interface{}{topRow, panelComponents[2]},
					Spacing: 1,
				})
			} else {
				// All panels stacked vertically
				contentLayout = components.VStack(components.StackProps{
					Items:   panelComponents,
					Spacing: 1,
				})
			}
			contentLayout.Init()

			// === INFO ===
			infoStyle := lipgloss.NewStyle().Foreground(theme.Muted).Italic(true)
			info := components.Text(components.TextProps{
				Content: infoStyle.Render(fmt.Sprintf(
					"Breakpoint: %s | Panel width: %d | Layout: %s",
					breakpoint, panelWidth, layoutMode,
				)),
			})
			info.Init()

			// === LAYOUT ===
			divider := components.Divider(components.DividerProps{
				Length: width - 8,
			})
			divider.Init()

			page := components.VStack(components.StackProps{
				Items:   []interface{}{title, divider, contentLayout, divider, info},
				Spacing: 1,
			})
			page.Init()

			box := components.Box(components.BoxProps{
				Child:   page,
				Border:  true,
				Padding: 1,
				Width:   width - 2,
			})
			box.Init()

			return box.View()
		}).
		Build()
}
