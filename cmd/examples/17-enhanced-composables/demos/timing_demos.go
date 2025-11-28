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

// CreateUseIntervalDemo creates the UseInterval demo view.
func CreateUseIntervalDemo() (bubbly.Component, error) {
	return bubbly.NewComponent("UseIntervalDemo").
		Setup(func(ctx *bubbly.Context) {
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)
			state := localComposables.UseSharedDemoState(ctx)
			ctx.Expose("state", state)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			theme := ctx.Get("theme").(bubbly.Theme)
			state := ctx.Get("state").(*localComposables.DemoStateComposable)

			// Use shared interval state
			isRunning := state.IntervalRunning.GetTyped()
			count := state.IntervalCount.GetTyped()

			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Primary).
				MarginBottom(1)

			codeStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

			usage := `interval := composables.UseInterval(ctx, func() {
    // Called every 500ms
    refreshData()
}, 500*time.Millisecond)

interval.Start()   // Begin interval
interval.Stop()    // Pause interval
interval.Toggle()  // Toggle running state
interval.Reset()   // Restart timing`

			// Running indicator
			statusStyle := lipgloss.NewStyle().Bold(true)
			status := "⏸ Paused"
			if isRunning {
				status = "▶ Running"
				statusStyle = statusStyle.Foreground(theme.Success)
			} else {
				statusStyle = statusStyle.Foreground(theme.Warning)
			}

			// Tick animation
			dots := strings.Repeat("●", count%5+1) + strings.Repeat("○", 4-count%5)

			stateContent := fmt.Sprintf(
				"Status: %s\nTick Count: %d\nInterval: 500ms\n\nAnimation: %s\n\nPress SPACE: start/stop | r: reset count",
				statusStyle.Render(status), count, dots,
			)

			stateCard := components.Card(components.CardProps{
				Title:   "Interactive Interval Demo",
				Content: stateContent,
				Width:   40,
			})
			stateCard.Init()

			descContent := "UseInterval executes a callback at regular intervals. Automatically cleaned up on unmount. Perfect for periodic updates and polling."

			descCard := components.Card(components.CardProps{
				Title:   "Description",
				Content: descContent,
				Width:   40,
			})
			descCard.Init()

			return lipgloss.JoinVertical(lipgloss.Left,
				titleStyle.Render("UseInterval Demo"),
				"",
				codeStyle.Render(usage),
				"",
				stateCard.View(),
				"",
				descCard.View(),
			)
		}).
		Build()
}

// CreateUseTimeoutDemo creates the UseTimeout demo view.
func CreateUseTimeoutDemo() (bubbly.Component, error) {
	return bubbly.NewComponent("UseTimeoutDemo").
		Setup(func(ctx *bubbly.Context) {
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)
			state := localComposables.UseSharedDemoState(ctx)
			ctx.Expose("state", state)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			theme := ctx.Get("theme").(bubbly.Theme)
			state := ctx.Get("state").(*localComposables.DemoStateComposable)

			// Use shared timeout state
			isPending := state.TimeoutPending.GetTyped()
			hasTriggered := state.TimeoutTriggered.GetTyped()

			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Primary).
				MarginBottom(1)

			codeStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

			usage := `timeout := composables.UseTimeout(ctx, func() {
    // Called after 3 seconds
    showNotification("Done!")
}, 3*time.Second)

timeout.Start()   // Begin countdown
timeout.Stop()    // Cancel timeout
timeout.Reset()   // Restart countdown
isPending := timeout.IsPending.GetTyped()`

			// Status indicator
			statusStyle := lipgloss.NewStyle().Bold(true)
			status := "○ Idle"
			if isPending {
				status = "⏳ Pending..."
				statusStyle = statusStyle.Foreground(theme.Warning)
			} else if hasTriggered {
				status = "✓ Triggered!"
				statusStyle = statusStyle.Foreground(theme.Success)
			} else {
				statusStyle = statusStyle.Foreground(theme.Muted)
			}

			stateContent := fmt.Sprintf(
				"Status: %s\nDelay: 3 seconds\nTriggered: %t\n\nPress SPACE: start | r: reset",
				statusStyle.Render(status), hasTriggered,
			)

			stateCard := components.Card(components.CardProps{
				Title:   "Interactive Timeout Demo",
				Content: stateContent,
				Width:   40,
			})
			stateCard.Init()

			descContent := "UseTimeout executes a callback after a delay. Can be cancelled or reset. Automatically cleaned up on unmount."

			descCard := components.Card(components.CardProps{
				Title:   "Description",
				Content: descContent,
				Width:   40,
			})
			descCard.Init()

			return lipgloss.JoinVertical(lipgloss.Left,
				titleStyle.Render("UseTimeout Demo"),
				"",
				codeStyle.Render(usage),
				"",
				stateCard.View(),
				"",
				descCard.View(),
			)
		}).
		Build()
}

// CreateUseTimerDemo creates the UseTimer demo view.
func CreateUseTimerDemo() (bubbly.Component, error) {
	return bubbly.NewComponent("UseTimerDemo").
		Setup(func(ctx *bubbly.Context) {
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)
			state := localComposables.UseSharedDemoState(ctx)
			ctx.Expose("state", state)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			theme := ctx.Get("theme").(bubbly.Theme)
			state := ctx.Get("state").(*localComposables.DemoStateComposable)

			// Sync timer state
			state.UpdateTimerDisplay()

			isRunning := state.TimerIsRunning.GetTyped()
			isExpired := state.TimerIsExpired.GetTyped()
			remaining := state.TimerRemaining.GetTyped()
			progress := state.TimerProgress.GetTyped()

			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Primary).
				MarginBottom(1)

			codeStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

			usage := `timer := composables.UseTimer(ctx, 30*time.Second,
    composables.WithTickInterval(100*time.Millisecond),
    composables.WithOnExpire(func() {
        playAlarm()
    }),
)
timer.Start()   // Begin countdown
timer.Stop()    // Pause timer
timer.Reset()   // Reset to initial duration
progress := timer.Progress.GetTyped()  // 0.0 to 1.0`

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
			barWidth := 30
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

			stateContent := fmt.Sprintf(
				"%s\nRemaining: %s\nProgress: %.0f%%\n\n%s\n\nPress SPACE: start/stop | r: reset",
				statusStyle.Render(status),
				remaining.Round(time.Second).String(),
				progress*100,
				bar.String(),
			)

			stateCard := components.Card(components.CardProps{
				Title:   "Interactive Timer Demo",
				Content: stateContent,
				Width:   40,
			})
			stateCard.Init()

			descContent := "UseTimer provides a countdown timer with progress tracking. Supports callbacks on expiration and configurable tick intervals."

			descCard := components.Card(components.CardProps{
				Title:   "Description",
				Content: descContent,
				Width:   40,
			})
			descCard.Init()

			return lipgloss.JoinVertical(lipgloss.Left,
				titleStyle.Render("UseTimer Demo"),
				"",
				codeStyle.Render(usage),
				"",
				stateCard.View(),
				"",
				descCard.View(),
			)
		}).
		Build()
}
