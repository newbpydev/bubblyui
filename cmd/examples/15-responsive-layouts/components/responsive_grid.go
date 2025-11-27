// Package components provides responsive demo components for the layout showcase.
package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	localComposables "github.com/newbpydev/bubblyui/cmd/examples/15-responsive-layouts/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// CreateResponsiveGrid creates a grid that adjusts columns based on terminal width.
// It demonstrates Flex with Wrap for automatic responsive behavior.
func CreateResponsiveGrid() (bubbly.Component, error) {
	return bubbly.NewComponent("ResponsiveGrid").
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
			gridCols := ws.GridColumns.GetTyped()

			// === TITLE ===
			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Primary)
			title := components.Text(components.TextProps{
				Content: titleStyle.Render(fmt.Sprintf("🔲 Responsive Grid (%d columns)", gridCols)),
			})
			title.Init()

			// === DESCRIPTION ===
			descStyle := lipgloss.NewStyle().Foreground(theme.Muted).Italic(true)
			desc := components.Text(components.TextProps{
				Content: descStyle.Render("Cards automatically wrap to fit available width using Flex with Wrap=true"),
			})
			desc.Init()

			// === GRID ITEMS ===
			items := []struct {
				title string
				icon  string
				color lipgloss.Color
			}{
				{"Dashboard", "📊", lipgloss.Color("35")},
				{"Analytics", "📈", lipgloss.Color("99")},
				{"Reports", "📋", lipgloss.Color("205")},
				{"Settings", "⚙️", lipgloss.Color("33")},
				{"Users", "👥", lipgloss.Color("42")},
				{"Messages", "💬", lipgloss.Color("39")},
				{"Calendar", "📅", lipgloss.Color("208")},
				{"Tasks", "✅", lipgloss.Color("35")},
				{"Files", "📁", lipgloss.Color("226")},
				{"Help", "❓", lipgloss.Color("99")},
				{"Profile", "👤", lipgloss.Color("205")},
				{"Logout", "🚪", lipgloss.Color("196")},
			}

			// Calculate card width based on grid columns
			availableWidth := width - 6 // Account for outer box padding/border
			cardWidth := (availableWidth - (gridCols - 1)) / gridCols
			if cardWidth < 10 {
				cardWidth = 10
			}
			if cardWidth > 18 {
				cardWidth = 18
			}

			var cardComponents []bubbly.Component
			for _, item := range items {
				iconStyle := lipgloss.NewStyle().
					Foreground(item.color)

				content := iconStyle.Render(item.icon)
				card := components.Card(components.CardProps{
					Title:   item.title,
					Content: content,
					Width:   cardWidth,
				})
				card.Init()
				cardComponents = append(cardComponents, card)
			}

			// Create flex grid with wrap
			flexWidth := availableWidth
			if flexWidth < 30 {
				flexWidth = 30
			}

			grid := components.Flex(components.FlexProps{
				Items:   cardComponents,
				Justify: components.JustifyStart,
				Gap:     1,
				Width:   flexWidth,
				Wrap:    true,
			})
			grid.Init()

			// === INFO ===
			infoStyle := lipgloss.NewStyle().Foreground(theme.Muted)
			info := components.Text(components.TextProps{
				Content: infoStyle.Render(fmt.Sprintf(
					"Grid: %d cols × %d items | Card width: %d chars | Container: %d chars",
					gridCols, len(items), cardWidth, flexWidth,
				)),
			})
			info.Init()

			// === LAYOUT ===
			divider := components.Divider(components.DividerProps{
				Length: flexWidth,
			})
			divider.Init()

			page := components.VStack(components.StackProps{
				Items:   []interface{}{title, desc, divider, grid, divider, info},
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
