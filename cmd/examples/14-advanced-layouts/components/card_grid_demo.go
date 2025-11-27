// Package components provides demo components for the layout showcase.
package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// CardGridDemoProps defines props for the card grid demo component.
type CardGridDemoProps struct{}

// CreateCardGridDemo creates a wrapping card grid demonstration.
// This showcases Flex with Wrap=true for responsive card layouts.
func CreateCardGridDemo() (bubbly.Component, error) {
	return bubbly.NewComponent("CardGridDemo").
		Setup(func(ctx *bubbly.Context) {
			ctx.ProvideTheme(bubbly.DefaultTheme)
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			theme := ctx.Get("theme").(bubbly.Theme)

			// === TITLE ===
			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Primary)
			title := components.Text(components.TextProps{
				Content: titleStyle.Render("📦 Card Grid Demo (Flex Wrap)"),
			})
			title.Init()

			// === DESCRIPTION ===
			descStyle := lipgloss.NewStyle().Foreground(theme.Muted).Italic(true)
			desc := components.Text(components.TextProps{
				Content: descStyle.Render("Cards automatically wrap to next row when they exceed container width"),
			})
			desc.Init()

			// === PRODUCT CARDS ===
			products := []struct {
				name  string
				price string
				icon  string
				color lipgloss.Color
			}{
				{"Laptop", "$999", "💻", lipgloss.Color("35")},
				{"Phone", "$699", "📱", lipgloss.Color("99")},
				{"Tablet", "$449", "📟", lipgloss.Color("205")},
				{"Watch", "$299", "⌚", lipgloss.Color("42")},
				{"Headphones", "$199", "🎧", lipgloss.Color("214")},
				{"Camera", "$549", "📷", lipgloss.Color("141")},
			}

			var cardComponents []bubbly.Component
			for _, product := range products {
				priceStyle := lipgloss.NewStyle().
					Bold(true).
					Foreground(product.color)

				content := fmt.Sprintf("%s\n%s", product.icon, priceStyle.Render(product.price))
				card := components.Card(components.CardProps{
					Title:   product.name,
					Content: content,
					Width:   14,
				})
				card.Init()
				cardComponents = append(cardComponents, card)
			}

			// === FLEX GRID WITH WRAP ===
			grid := components.Flex(components.FlexProps{
				Items:   cardComponents,
				Justify: components.JustifyStart,
				Gap:     2,
				Wrap:    true,
				Width:   50, // Smaller width forces wrapping
			})
			grid.Init()

			gridBox := components.Box(components.BoxProps{
				Child:       grid,
				Border:      true,
				BorderStyle: lipgloss.RoundedBorder(),
				Padding:     1,
				Title:       "Product Grid (Width: 50)",
			})
			gridBox.Init()

			// === SECOND GRID - SPACE EVENLY ===
			var cardComponents2 []bubbly.Component
			for _, product := range products[:4] {
				priceStyle := lipgloss.NewStyle().
					Bold(true).
					Foreground(product.color)

				content := fmt.Sprintf("%s\n%s", product.icon, priceStyle.Render(product.price))
				card := components.Card(components.CardProps{
					Title:   product.name,
					Content: content,
					Width:   14,
				})
				card.Init()
				cardComponents2 = append(cardComponents2, card)
			}

			grid2 := components.Flex(components.FlexProps{
				Items:   cardComponents2,
				Justify: components.JustifySpaceEvenly,
				Gap:     1,
				Wrap:    true,
				Width:   70,
			})
			grid2.Init()

			grid2Box := components.Box(components.BoxProps{
				Child:       grid2,
				Border:      true,
				BorderStyle: lipgloss.RoundedBorder(),
				Padding:     1,
				Title:       "Space Evenly Grid (Width: 70)",
			})
			grid2Box.Init()

			// === LAYOUT ===
			divider := components.Divider(components.DividerProps{
				Length: 70,
			})
			divider.Init()

			page := components.VStack(components.StackProps{
				Items: []interface{}{
					title,
					desc,
					divider,
					gridBox,
					divider,
					grid2Box,
				},
				Spacing: 1,
			})
			page.Init()

			return page.View()
		}).
		Build()
}
