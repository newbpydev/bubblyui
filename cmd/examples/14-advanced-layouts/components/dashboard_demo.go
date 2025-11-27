// Package components provides demo components for the layout showcase.
package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// DashboardDemoProps defines props for the dashboard demo component.
type DashboardDemoProps struct{}

// CreateDashboardDemo creates a dashboard layout demonstration.
// This showcases HStack, VStack, Flex, Spacer, Divider, and Box components
// in a real-world dashboard pattern.
func CreateDashboardDemo() (bubbly.Component, error) {
	return bubbly.NewComponent("DashboardDemo").
		Setup(func(ctx *bubbly.Context) {
			// Provide theme for child components
			ctx.ProvideTheme(bubbly.DefaultTheme)
			// Use theme and expose for template
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			theme := ctx.Get("theme").(bubbly.Theme)

			// === HEADER ===
			// Logo
			logoStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205"))
			logo := components.Text(components.TextProps{
				Content: logoStyle.Render("📊 Dashboard"),
			})
			logo.Init()

			// Flexible spacer pushes actions to the right
			spacer := components.Spacer(components.SpacerProps{Flex: true})
			spacer.Init()

			// Action buttons
			settingsBtn := components.Button(components.ButtonProps{
				Label:   "⚙️ Settings",
				Variant: "secondary",
			})
			settingsBtn.Init()

			profileBtn := components.Button(components.ButtonProps{
				Label:   "👤 Profile",
				Variant: "primary",
			})
			profileBtn.Init()

			// Header HStack
			header := components.HStack(components.StackProps{
				Items:   []interface{}{logo, spacer, settingsBtn, profileBtn},
				Spacing: 2,
				Align:   components.AlignItemsCenter,
			})
			header.Init()

			// Header box with border
			// Width = sidebar (20) + contentBox (71 + 2 border) = 93
			headerBox := components.Box(components.BoxProps{
				Child:   header,
				Padding: 1,
				Border:  true,
				Width:   93,
			})
			headerBox.Init()

			// === SIDEBAR ===
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

			sidebarBox := components.Box(components.BoxProps{
				Child:   sidebar,
				Title:   "Navigation",
				Border:  true,
				Padding: 1,
				Width:   20,
			})
			sidebarBox.Init()

			// === MAIN CONTENT - Stat Cards ===
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
			}

			var cardComponents []bubbly.Component
			for _, stat := range statCards {
				valueStyle := lipgloss.NewStyle().
					Bold(true).
					Foreground(stat.color)

				content := fmt.Sprintf("%s\n%s", stat.icon, valueStyle.Render(stat.value))
				// Width must be at least 14 to fit: 2 (border) + 2 (padding) + content
				// "Revenue" is 7 chars, "$45.2K" is 6 chars - need at least 14 for proper borders
				card := components.Card(components.CardProps{
					Title:   stat.title,
					Content: content,
					Width:   14,
				})
				card.Init()
				cardComponents = append(cardComponents, card)
			}

			// Flex layout for cards with space-between
			// Card Width=14 renders to 16 chars (14 + 2 for border)
			// 4 cards × 16 = 64, plus 3 gaps × 1 = 3, total = 67 chars
			cardGrid := components.Flex(components.FlexProps{
				Items:   cardComponents,
				Justify: components.JustifySpaceBetween,
				Gap:     1,
				Width:   67,
			})
			cardGrid.Init()

			// === CONTENT AREA ===
			contentTitle := components.Text(components.TextProps{
				Content: lipgloss.NewStyle().Bold(true).Foreground(theme.Primary).Render("📈 Statistics Overview"),
			})
			contentTitle.Init()

			contentDivider := components.Divider(components.DividerProps{
				Length: 65,
				Label:  "Stats",
			})
			contentDivider.Init()

			contentArea := components.VStack(components.StackProps{
				Items:   []interface{}{contentTitle, contentDivider, cardGrid},
				Spacing: 1,
			})
			contentArea.Init()

			// Box Width must accommodate content (67 chars) + padding (2) = 69
			// Lipgloss Width sets content area, padding is inside that
			// So we need Width >= 67 + 2 (padding left/right) = 69
			// Adding 2 more for safety margin = 71
			contentBox := components.Box(components.BoxProps{
				Child:   contentArea,
				Padding: 1,
				Border:  true,
				Width:   71,
			})
			contentBox.Init()

			// === MAIN LAYOUT ===
			// Sidebar + Content horizontally
			mainContent := components.HStack(components.StackProps{
				Items:   []interface{}{sidebarBox, contentBox},
				Spacing: 0,
			})
			mainContent.Init()

			// === FOOTER ===
			footerText := components.Text(components.TextProps{
				Content: lipgloss.NewStyle().Foreground(theme.Muted).Render("© 2025 BubblyUI • Advanced Layout System Demo"),
			})
			footerText.Init()

			footerCenter := components.Center(components.CenterProps{
				Child: footerText,
				Width: 93,
			})
			footerCenter.Init()

			// === FULL PAGE ===
			page := components.VStack(components.StackProps{
				Items:   []interface{}{headerBox, mainContent, footerCenter},
				Spacing: 0,
			})
			page.Init()

			return page.View()
		}).
		Build()
}
