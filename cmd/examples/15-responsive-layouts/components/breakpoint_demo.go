// Package components provides responsive demo components for the layout showcase.
package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	localComposables "github.com/newbpydev/bubblyui/cmd/examples/15-responsive-layouts/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// CreateBreakpointDemo creates a demo showing current breakpoint information.
// It displays a visual indicator of the current breakpoint and available space.
func CreateBreakpointDemo() (bubbly.Component, error) {
	return bubbly.NewComponent("BreakpointDemo").
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
			height := ws.Height.GetTyped()
			breakpoint := ws.Breakpoint.GetTyped()

			// === TITLE ===
			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Primary)
			title := components.Text(components.TextProps{
				Content: titleStyle.Render("📐 Breakpoint Information"),
			})
			title.Init()

			// === BREAKPOINT INDICATOR ===
			breakpoints := []struct {
				bp       localComposables.Breakpoint
				name     string
				minWidth int
				color    lipgloss.Color
			}{
				{localComposables.BreakpointXS, "XS (<60)", 0, lipgloss.Color("196")},
				{localComposables.BreakpointSM, "SM (60-79)", 60, lipgloss.Color("208")},
				{localComposables.BreakpointMD, "MD (80-119)", 80, lipgloss.Color("226")},
				{localComposables.BreakpointLG, "LG (120-159)", 120, lipgloss.Color("35")},
				{localComposables.BreakpointXL, "XL (160+)", 160, lipgloss.Color("39")},
			}

			var bpItems []interface{}
			for _, bp := range breakpoints {
				var indicator string
				if bp.bp == breakpoint {
					// Active breakpoint
					style := lipgloss.NewStyle().
						Bold(true).
						Foreground(lipgloss.Color("0")).
						Background(bp.color).
						Padding(0, 1)
					indicator = style.Render(fmt.Sprintf("▶ %s ◀", bp.name))
				} else {
					// Inactive breakpoint
					style := lipgloss.NewStyle().
						Foreground(theme.Muted)
					indicator = style.Render(fmt.Sprintf("  %s  ", bp.name))
				}
				text := components.Text(components.TextProps{
					Content: indicator,
				})
				text.Init()
				bpItems = append(bpItems, text)
			}

			bpRow := components.HStack(components.StackProps{
				Items:   bpItems,
				Spacing: 1,
			})
			bpRow.Init()

			// === SIZE INFO ===
			sizeStyle := lipgloss.NewStyle().Foreground(theme.Secondary)
			sizeInfo := components.Text(components.TextProps{
				Content: sizeStyle.Render(fmt.Sprintf("Terminal Size: %d × %d", width, height)),
			})
			sizeInfo.Init()

			// === VISUAL WIDTH BAR ===
			barWidth := width - 10
			if barWidth > 100 {
				barWidth = 100
			}
			if barWidth < 20 {
				barWidth = 20
			}

			// Create a visual bar showing relative width
			filledWidth := (width * barWidth) / 200 // Scale to max 200 cols
			if filledWidth > barWidth {
				filledWidth = barWidth
			}

			barStyle := lipgloss.NewStyle().Foreground(theme.Primary)
			emptyStyle := lipgloss.NewStyle().Foreground(theme.Muted)

			bar := barStyle.Render(repeatChar("█", filledWidth)) +
				emptyStyle.Render(repeatChar("░", barWidth-filledWidth))

			barText := components.Text(components.TextProps{
				Content: fmt.Sprintf("[%s] %d cols", bar, width),
			})
			barText.Init()

			// === LAYOUT ===
			divider := components.Divider(components.DividerProps{
				Length: width - 10,
			})
			divider.Init()

			page := components.VStack(components.StackProps{
				Items:   []interface{}{title, divider, bpRow, sizeInfo, barText},
				Spacing: 1,
			})
			page.Init()

			box := components.Box(components.BoxProps{
				Child:   page,
				Border:  true,
				Padding: 1,
				Width:   width - 4,
			})
			box.Init()

			return box.View()
		}).
		Build()
}

// repeatChar repeats a character n times.
func repeatChar(char string, n int) string {
	if n <= 0 {
		return ""
	}
	result := ""
	for i := 0; i < n; i++ {
		result += char
	}
	return result
}
