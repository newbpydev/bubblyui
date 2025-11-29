// Package components provides UI components for the enhanced composables demo.
package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	localComposables "github.com/newbpydev/bubblyui/cmd/examples/17-enhanced-composables/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// CreateTimerCard creates the timer demonstration card component.
func CreateTimerCard() (bubbly.Component, error) {
	return bubbly.NewComponent("TimerCard").
		Setup(func(ctx *bubbly.Context) {
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)
			state := localComposables.UseSharedDemoState(ctx)
			ctx.Expose("state", state)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			theme := ctx.Get("theme").(bubbly.Theme)
			state := ctx.Get("state").(*localComposables.DemoStateComposable)

			isRunning := state.TimerIsRunning.GetTyped()
			isExpired := state.TimerIsExpired.GetTyped()
			remaining := state.TimerRemaining.GetTyped()
			progress := state.TimerProgress.GetTyped()

			// Status with color
			var statusStyle lipgloss.Style
			status := "⏹ Stopped"
			if isRunning {
				status = "▶ Running"
				statusStyle = lipgloss.NewStyle().Foreground(theme.Success)
			} else if isExpired {
				status = "✓ Expired"
				statusStyle = lipgloss.NewStyle().Foreground(theme.Warning)
			} else {
				statusStyle = lipgloss.NewStyle().Foreground(theme.Muted)
			}

			// Colorful progress bar
			barWidth := 28
			filled := int(progress * float64(barWidth))
			var bar strings.Builder

			filledStyle := lipgloss.NewStyle().Foreground(theme.Primary)
			emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

			for i := 0; i < barWidth; i++ {
				if i < filled {
					bar.WriteString(filledStyle.Render("█"))
				} else {
					bar.WriteString(emptyStyle.Render("░"))
				}
			}

			percentStyle := lipgloss.NewStyle().Foreground(theme.Primary).Bold(true)

			content := fmt.Sprintf(
				"%s\nRemaining: %s\n\n%s %s\n\nt: start/stop/reset",
				statusStyle.Render(status),
				remaining.Round(time.Second).String(),
				bar.String(),
				percentStyle.Render(fmt.Sprintf("%.0f%%", progress*100)),
			)

			card := components.Card(components.CardProps{
				Title:   "Timer (UseTimer)",
				Content: content,
				Width:   42,
			})
			card.Init()
			return card.View()
		}).
		Build()
}
