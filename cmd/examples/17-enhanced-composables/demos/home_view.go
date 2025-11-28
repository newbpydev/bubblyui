// Package demos provides demo views for each composable.
package demos

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	localComposables "github.com/newbpydev/bubblyui/cmd/examples/17-enhanced-composables/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// CreateHomeView creates the home view with overview cards.
func CreateHomeView() (bubbly.Component, error) {
	return bubbly.NewComponent("HomeView").
		Setup(func(ctx *bubbly.Context) {
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)
			state := localComposables.UseSharedDemoState(ctx)
			ctx.Expose("state", state)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			theme := ctx.Get("theme").(bubbly.Theme)
			state := ctx.Get("state").(*localComposables.DemoStateComposable)

			// Counter Card
			counterVal := state.CounterValue.GetTyped()
			prevVal := state.PreviousVal.GetTyped()
			canUndo := state.CanUndo.GetTyped()
			canRedo := state.CanRedo.GetTyped()

			prevStr := "none"
			if prevVal != nil {
				prevStr = fmt.Sprintf("%d", *prevVal)
			}

			counterContent := fmt.Sprintf(
				"Value: %d (prev: %s)\nCan Undo: %t | Can Redo: %t\n\n+/-: change | u/r: undo/redo",
				counterVal, prevStr, canUndo, canRedo,
			)

			counterCard := components.Card(components.CardProps{
				Title:   "Counter (UseCounter + UseHistory)",
				Content: counterContent,
				Width:   42,
			})
			counterCard.Init()

			// Timer Card
			isRunning := state.TimerIsRunning.GetTyped()
			isExpired := state.TimerIsExpired.GetTyped()
			remaining := state.TimerRemaining.GetTyped()
			progress := state.TimerProgress.GetTyped()

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

			timerContent := fmt.Sprintf(
				"%s\nRemaining: %s\n\n%s %s\n\nt: start/stop/reset",
				statusStyle.Render(status),
				remaining.Round(time.Second).String(),
				bar.String(),
				percentStyle.Render(fmt.Sprintf("%.0f%%", progress*100)),
			)

			timerCard := components.Card(components.CardProps{
				Title:   "Timer (UseTimer)",
				Content: timerContent,
				Width:   42,
			})
			timerCard.Init()

			// Collections Card
			tasks := state.TaskList.GetTyped()
			tags := state.TagsSlice.GetTyped()
			darkMode := state.DarkMode.GetTyped()

			collectionsContent := fmt.Sprintf(
				"Tasks (UseList): %d items\nTags (UseSet): [%s]\nDark Mode (UseToggle): %t\n\nspace: toggle dark mode",
				len(tasks),
				strings.Join(tags, ", "),
				darkMode,
			)

			collectionsCard := components.Card(components.CardProps{
				Title:   "Collections & Toggle",
				Content: collectionsContent,
				Width:   42,
			})
			collectionsCard.Init()

			// Welcome message
			welcomeStyle := lipgloss.NewStyle().
				Foreground(theme.Primary).
				Bold(true).
				MarginBottom(1)

			welcome := welcomeStyle.Render("Welcome to Enhanced Composables Demo!")
			hint := lipgloss.NewStyle().
				Foreground(theme.Muted).
				Italic(true).
				Render("Select a composable from the sidebar and press Enter to see its dedicated demo.")

			return lipgloss.JoinVertical(lipgloss.Left,
				welcome,
				hint,
				"",
				counterCard.View(),
				"",
				timerCard.View(),
				"",
				collectionsCard.View(),
			)
		}).
		Build()
}
