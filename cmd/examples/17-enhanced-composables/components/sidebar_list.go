// Package components provides UI components for the enhanced composables demo.
package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	localComposables "github.com/newbpydev/bubblyui/cmd/examples/17-enhanced-composables/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

func CreateSidebarList() (bubbly.Component, error) {
	return bubbly.NewComponent("SidebarList").
		Setup(func(ctx *bubbly.Context) {
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)
			state := localComposables.UseSharedDemoState(ctx)
			ctx.Expose("state", state)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			theme := ctx.Get("theme").(bubbly.Theme)
			state := ctx.Get("state").(*localComposables.DemoStateComposable)

			items := state.SidebarItems
			selectedIdx := state.SidebarIndex.GetTyped()
			focusedPane := state.FocusedPane.GetTyped()
			innerWidth := state.SidebarWidth - 4

			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Primary).
				Width(innerWidth).
				MarginBottom(1)

			title := titleStyle.Render("Composables")

			categoryStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Bold(true).
				Width(innerWidth)

			var lines []string
			currentCategory := ""

			for i, item := range items {
				if item.Category != currentCategory {
					currentCategory = item.Category
					headerText := fmt.Sprintf("-- %s --", currentCategory)
					lines = append(lines, categoryStyle.Render(headerText))
				}

				isSelected := focusedPane == localComposables.FocusSidebar && i == selectedIdx

				var style lipgloss.Style
				prefix := "  "

				if isSelected {
					prefix = "> "
					style = lipgloss.NewStyle().
						Bold(true).
						Foreground(lipgloss.Color("0")).
						Background(lipgloss.Color("99")).
						Width(innerWidth)
				} else {
					style = lipgloss.NewStyle().
						Foreground(lipgloss.Color("252")).
						Width(innerWidth)
				}

				lines = append(lines, style.Render(prefix+item.Name))
			}

			scrollInfo := lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Italic(true).
				Render(fmt.Sprintf("\n[%d/%d]", selectedIdx+1, len(items)))

			content := strings.Join(lines, "\n")

			return lipgloss.JoinVertical(lipgloss.Left, title, content, scrollInfo)
		}).
		Build()
}
