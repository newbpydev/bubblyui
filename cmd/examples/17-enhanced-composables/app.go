// Package main provides the Enhanced Composables Demo application.
package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	localComponents "github.com/newbpydev/bubblyui/cmd/examples/17-enhanced-composables/components"
	localComposables "github.com/newbpydev/bubblyui/cmd/examples/17-enhanced-composables/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// CreateApp creates the root application component for the enhanced composables demo.
func CreateApp() (bubbly.Component, error) {
	return bubbly.NewComponent("EnhancedComposablesApp").
		WithAutoCommands(true).
		WithKeyBinding("ctrl+c", "quit", "Quit").
		WithMultiKeyBindings("navUp", "Navigate up", "up", "k").
		WithMultiKeyBindings("navDown", "Navigate down", "down", "j").
		WithKeyBinding("tab", "cycleFocus", "Next pane").
		WithKeyBinding("+", "increment", "Increment counter").
		WithKeyBinding("-", "decrement", "Decrement counter").
		WithKeyBinding("u", "undo", "Undo").
		WithKeyBinding("r", "redo", "Redo").
		WithKeyBinding("t", "toggleTimer", "Start/stop timer").
		WithKeyBinding(" ", "toggle", "Toggle dark mode").
		WithKeyBinding("n", "notify", "Show notification").
		WithKeyBinding("enter", "select", "Select item").
		Setup(func(ctx *bubbly.Context) {
			ctx.ProvideTheme(bubbly.DefaultTheme)
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)

			state := localComposables.UseSharedDemoState(ctx)
			ctx.Expose("state", state)

			sidebarList, _ := localComponents.CreateSidebarList()
			counterCard, _ := localComponents.CreateCounterCard()
			timerCard, _ := localComponents.CreateTimerCard()
			collectionsCard, _ := localComponents.CreateCollectionsCard()

			_ = ctx.ExposeComponent("sidebarList", sidebarList)
			_ = ctx.ExposeComponent("counterCard", counterCard)
			_ = ctx.ExposeComponent("timerCard", timerCard)
			_ = ctx.ExposeComponent("collectionsCard", collectionsCard)

			ctx.On("resize", func(data interface{}) {
				if sizeData, ok := data.(map[string]int); ok {
					state.SetSize(sizeData["width"], sizeData["height"])
				}
			})

			ctx.On("navUp", func(_ interface{}) {
				if state.FocusedPane.GetTyped() == localComposables.FocusSidebar {
					state.SidebarUp()
				}
			})

			ctx.On("navDown", func(_ interface{}) {
				if state.FocusedPane.GetTyped() == localComposables.FocusSidebar {
					state.SidebarDown()
				}
			})

			ctx.On("cycleFocus", func(_ interface{}) {
				state.CycleFocus()
			})

			ctx.On("select", func(_ interface{}) {
				if state.FocusedPane.GetTyped() == localComposables.FocusSidebar {
					state.SelectSidebarItem()
				}
			})

			ctx.On("increment", func(_ interface{}) {
				state.Increment()
			})

			ctx.On("decrement", func(_ interface{}) {
				state.Decrement()
			})

			ctx.On("undo", func(_ interface{}) {
				state.Undo()
			})

			ctx.On("redo", func(_ interface{}) {
				state.Redo()
			})

			ctx.On("toggleTimer", func(_ interface{}) {
				state.ToggleTimer()
			})

			ctx.On("toggle", func(_ interface{}) {
				state.ToggleDarkMode()
			})

			ctx.On("notify", func(_ interface{}) {
				val := state.CounterValue.GetTyped()
				state.ShowNotification("Counter", strings.Repeat("*", val/10))
			})

			ctx.OnMounted(func() {
				state.UpdateTimerDisplay()
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			state := ctx.Get("state").(*localComposables.DemoStateComposable)
			theme := ctx.Get("theme").(bubbly.Theme)

			// Sync timer state on each render (asyncWrapperModel ticks every 100ms)
			state.UpdateTimerDisplay()

			width := state.Width.GetTyped()
			height := state.Height.GetTyped()
			focusedPane := state.FocusedPane.GetTyped()
			selectedDetail := state.SelectedDetail.GetTyped()

			sidebarList := ctx.Get("sidebarList").(bubbly.Component)
			counterCard := ctx.Get("counterCard").(bubbly.Component)
			timerCard := ctx.Get("timerCard").(bubbly.Component)
			collectionsCard := ctx.Get("collectionsCard").(bubbly.Component)

			headerHeight := 1
			footerHeight := 1
			mainHeight := height - headerHeight - footerHeight - 2

			if mainHeight < 15 {
				mainHeight = 15
			}

			sidebarWidth := state.SidebarWidth
			contentWidth := width - sidebarWidth - 3

			headerStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205")).
				Background(lipgloss.Color("236")).
				Padding(0, 1).
				Width(width)

			modeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
			headerContent := "Enhanced Composables Demo" + modeStyle.Render(" | "+selectedDetail)
			headerRendered := headerStyle.Render(headerContent)

			sidebarBorderColor := lipgloss.Color("240")
			if focusedPane == localComposables.FocusSidebar {
				sidebarBorderColor = lipgloss.Color("99")
			}

			sidebarBox := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(sidebarBorderColor).
				Width(sidebarWidth - 2).
				Height(mainHeight - 2)

			sidebarRendered := sidebarBox.Render(sidebarList.View())

			mainBorderColor := lipgloss.Color("240")
			if focusedPane == localComposables.FocusMain {
				mainBorderColor = theme.Primary
			}

			cardsContent := lipgloss.JoinVertical(lipgloss.Left,
				counterCard.View(),
				"",
				timerCard.View(),
				"",
				collectionsCard.View(),
			)

			mainBox := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(mainBorderColor).
				Width(contentWidth).
				Height(mainHeight-2).
				Padding(0, 1)

			mainRendered := mainBox.Render(cardsContent)

			mainArea := lipgloss.JoinHorizontal(lipgloss.Top, sidebarRendered, " ", mainRendered)

			var statusText string
			switch focusedPane {
			case localComposables.FocusSidebar:
				statusText = "SIDEBAR | jk: navigate | enter: select | tab: main"
			case localComposables.FocusMain:
				statusText = "MAIN | +/-: counter | u/r: undo/redo | t: timer | space: toggle | n: notify | tab: sidebar"
			}

			statusStyle := lipgloss.NewStyle().
				Foreground(theme.Primary).
				Padding(0, 1)

			statusRendered := statusStyle.Render(statusText)

			notifications := state.Notifications.Notifications.GetTyped()
			var notifOverlay string
			if len(notifications) > 0 {
				var notifLines []string
				for _, n := range notifications {
					notifStyle := lipgloss.NewStyle().
						Background(lipgloss.Color("22")).
						Foreground(lipgloss.Color("15")).
						Padding(0, 1)
					notifLines = append(notifLines, notifStyle.Render("OK "+n.Title+": "+n.Message))
				}
				notifOverlay = "\n" + strings.Join(notifLines, "\n")
			}

			var output strings.Builder
			output.WriteString(headerRendered)
			output.WriteString("\n")
			output.WriteString(mainArea)
			output.WriteString("\n")
			output.WriteString(statusRendered)
			output.WriteString(notifOverlay)

			return output.String()
		}).
		Build()
}
