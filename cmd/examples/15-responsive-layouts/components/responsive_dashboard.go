// Package components provides responsive demo components for the layout showcase.
package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	localComposables "github.com/newbpydev/bubblyui/cmd/examples/15-responsive-layouts/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// CreateResponsiveDashboard creates a dashboard that adapts to terminal size.
// It shows/hides sidebar, adjusts card grid columns, and resizes content.
func CreateResponsiveDashboard() (bubbly.Component, error) {
	return bubbly.NewComponent("ResponsiveDashboard").
		Setup(func(ctx *bubbly.Context) {
			ctx.ProvideTheme(bubbly.DefaultTheme)
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)

			// Get shared window size state
			windowSize := localComposables.UseSharedWindowSize(ctx)
			ctx.Expose("windowSize", windowSize)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			theme := ctx.Get("theme").(bubbly.Theme)
			ws := ctx.Get("windowSize").(*localComposables.WindowSizeComposable)

			width := ws.Width.GetTyped()
			height := ws.Height.GetTyped()
			breakpoint := ws.Breakpoint.GetTyped()
			sidebarVisible := ws.SidebarVisible.GetTyped()
			gridCols := ws.GridColumns.GetTyped()
			contentWidth := ws.GetContentWidth()

			// === HEADER ===
			headerStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205"))

			headerText := fmt.Sprintf("📊 Responsive Dashboard [%dx%d] [%s]", width, height, breakpoint)
			header := components.Text(components.TextProps{
				Content: headerStyle.Render(headerText),
			})
			header.Init()

			// === SIDEBAR (only on md+ screens) ===
			var sidebarBox bubbly.Component
			if sidebarVisible {
				menuItems := []string{"📁 Files", "📊 Analytics", "👥 Users", "⚙️ Settings"}
				var sidebarItems []interface{}

				for _, item := range menuItems {
					text := components.Text(components.TextProps{
						Content: item,
					})
					text.Init()
					sidebarItems = append(sidebarItems, text)
				}

				sidebar := components.VStack(components.StackProps{
					Items:       sidebarItems,
					Spacing:     1,
					Divider:     true,
					DividerChar: "─",
				})
				sidebar.Init()

				sidebarBox = components.Box(components.BoxProps{
					Child:   sidebar,
					Title:   "Navigation",
					Border:  true,
					Padding: 1,
					Width:   20,
				})
				sidebarBox.Init()
			}

			// === STAT CARDS (responsive grid) ===
			statCards := []struct {
				title string
				value string
				icon  string
				color lipgloss.Color
			}{
				{"Users", "1,234", "👥", lipgloss.Color("35")},
				{"Revenue", "$45.2K", "💰", lipgloss.Color("99")},
				{"Orders", "89", "📦", lipgloss.Color("205")},
				{"Growth", "+12%", "📈", lipgloss.Color("42")},
				{"Sessions", "5.2K", "🔗", lipgloss.Color("33")},
				{"Bounce", "32%", "↩️", lipgloss.Color("196")},
			}

			// Calculate card width based on available space
			cardWidth := ws.GetCardWidth()
			if cardWidth < 12 {
				cardWidth = 12 // Minimum card width
			}
			if cardWidth > 20 {
				cardWidth = 20 // Maximum card width
			}

			var cardComponents []bubbly.Component
			for _, stat := range statCards {
				valueStyle := lipgloss.NewStyle().
					Bold(true).
					Foreground(stat.color)

				content := fmt.Sprintf("%s\n%s", stat.icon, valueStyle.Render(stat.value))
				card := components.Card(components.CardProps{
					Title:   stat.title,
					Content: content,
					Width:   cardWidth,
				})
				card.Init()
				cardComponents = append(cardComponents, card)
			}

			// Create flex grid with wrap
			flexWidth := contentWidth - 4 // Account for box padding/border
			if flexWidth < 40 {
				flexWidth = 40
			}

			cardGrid := components.Flex(components.FlexProps{
				Items:   cardComponents,
				Justify: components.JustifyStart,
				Gap:     1,
				Width:   flexWidth,
				Wrap:    true, // Enable wrapping for responsive behavior
			})
			cardGrid.Init()

			// === CONTENT AREA ===
			contentTitle := components.Text(components.TextProps{
				Content: lipgloss.NewStyle().Bold(true).Foreground(theme.Primary).Render(
					fmt.Sprintf("📈 Statistics (%d columns)", gridCols),
				),
			})
			contentTitle.Init()

			contentDivider := components.Divider(components.DividerProps{
				Length: flexWidth - 2,
				Label:  "Stats",
			})
			contentDivider.Init()

			// Info about current layout
			infoStyle := lipgloss.NewStyle().Foreground(theme.Muted).Italic(true)
			layoutInfo := fmt.Sprintf(
				"Breakpoint: %s | Sidebar: %v | Grid: %d cols | Card: %d chars",
				breakpoint, sidebarVisible, gridCols, cardWidth,
			)
			info := components.Text(components.TextProps{
				Content: infoStyle.Render(layoutInfo),
			})
			info.Init()

			contentArea := components.VStack(components.StackProps{
				Items:   []interface{}{contentTitle, contentDivider, cardGrid, info},
				Spacing: 1,
			})
			contentArea.Init()

			contentBox := components.Box(components.BoxProps{
				Child:   contentArea,
				Padding: 1,
				Border:  true,
				Width:   contentWidth,
			})
			contentBox.Init()

			// === MAIN LAYOUT ===
			var mainContent bubbly.Component
			if sidebarVisible && sidebarBox != nil {
				mainContent = components.HStack(components.StackProps{
					Items:   []interface{}{sidebarBox, contentBox},
					Spacing: 0,
				})
			} else {
				mainContent = contentBox
			}
			mainContent.Init()

			// === FOOTER ===
			footerStyle := lipgloss.NewStyle().Foreground(theme.Muted)
			footerText := components.Text(components.TextProps{
				Content: footerStyle.Render("Resize terminal to see responsive behavior • Minimum: 60x20"),
			})
			footerText.Init()

			footerCenter := components.Center(components.CenterProps{
				Child: footerText,
				Width: width - 2,
			})
			footerCenter.Init()

			// === FULL PAGE ===
			page := components.VStack(components.StackProps{
				Items:   []interface{}{header, mainContent, footerCenter},
				Spacing: 1,
			})
			page.Init()

			return page.View()
		}).
		Build()
}
