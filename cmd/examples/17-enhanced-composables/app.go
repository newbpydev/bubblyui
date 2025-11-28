// Package main provides the Enhanced Composables Demo application.
package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	localComponents "github.com/newbpydev/bubblyui/cmd/examples/17-enhanced-composables/components"
	localComposables "github.com/newbpydev/bubblyui/cmd/examples/17-enhanced-composables/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
)

// getStatusTextForView returns the appropriate status bar text for the given view.
func getStatusTextForView(view localComposables.ViewType, focusedPane localComposables.FocusPane) string {
	if focusedPane == localComposables.FocusSidebar {
		return "SIDEBAR | jk: navigate | enter: select | h: home | tab: main"
	}

	// Main pane - show view-specific controls
	// Note: UseFocus hijacks tab, so we show esc for that view
	switch view {
	case localComposables.ViewHome:
		return "HOME | +/-: counter | u/r: undo/redo | t: timer | space: toggle | n: notify | esc: sidebar"
	case localComposables.ViewUseWindowSize:
		return "UseWindowSize | (resize terminal to see changes) | esc: sidebar"
	case localComposables.ViewUseFocus:
		return "UseFocus | 1/2/3: focus pane | tab: cycle demo panes | esc: sidebar"
	case localComposables.ViewUseScroll:
		return "UseScroll | jk: scroll | g: top | G: bottom | esc: sidebar"
	case localComposables.ViewUseSelection:
		return "UseSelection | jk: move | space: toggle select | esc: sidebar"
	case localComposables.ViewUseMode:
		return "UseMode | 1: normal | 2: insert | 3: visual | 4: command | esc: sidebar"
	case localComposables.ViewUseToggle:
		return "UseToggle | 1/2/3: toggle each | esc: sidebar"
	case localComposables.ViewUseCounter:
		return "UseCounter | +/-: change | r: reset | esc: sidebar"
	case localComposables.ViewUsePrevious:
		return "UsePrevious | +/-: change counter (see previous) | esc: sidebar"
	case localComposables.ViewUseHistory:
		return "UseHistory | +/-: change | u: undo | r: redo | esc: sidebar"
	case localComposables.ViewUseInterval:
		return "UseInterval | space: start/stop | r: reset count | esc: sidebar"
	case localComposables.ViewUseTimeout:
		return "UseTimeout | space: start | r: reset | esc: sidebar"
	case localComposables.ViewUseTimer:
		return "UseTimer | space: start/stop | r: reset | esc: sidebar"
	case localComposables.ViewUseList:
		return "UseList | jk: move | a: add | d: delete | c: clear | esc: sidebar"
	case localComposables.ViewUseMap:
		return "UseMap | a: add | d: delete | c: clear | esc: sidebar"
	case localComposables.ViewUseSet:
		return "UseSet | a: add | d: delete | t: toggle 'bubbly' | c: clear | esc: sidebar"
	case localComposables.ViewUseQueue:
		return "UseQueue | e: enqueue | d: dequeue | c: clear | esc: sidebar"
	case localComposables.ViewUseLogger:
		return "UseLogger | d: debug | i: info | w: warn | e: error | c: clear | esc: sidebar"
	case localComposables.ViewUseNotification:
		return "UseNotification | s: success | e: error | w: warning | i: info | c: clear | esc: sidebar"
	case localComposables.ViewCreateShared:
		return "CreateShared | +/-: change shared counter | esc: sidebar"
	case localComposables.ViewCreateSharedReset:
		return "CreateSharedWithReset | +/-: change | r: reset | esc: sidebar"
	default:
		return "esc: sidebar"
	}
}

// CreateApp creates the root application component for the enhanced composables demo.
func CreateApp() (bubbly.Component, error) {
	return bubbly.NewComponent("EnhancedComposablesApp").
		WithAutoCommands(true).
		WithKeyBinding("ctrl+c", "quit", "Quit").
		WithMultiKeyBindings("navUp", "Navigate up", "up", "k").
		WithMultiKeyBindings("navDown", "Navigate down", "down", "j").
		WithKeyBinding("tab", "cycleFocus", "Next pane").
		WithKeyBinding("+", "increment", "Increment").
		WithKeyBinding("-", "decrement", "Decrement").
		WithKeyBinding("u", "undo", "Undo").
		WithKeyBinding("r", "actionR", "Reset/Redo").
		WithKeyBinding("t", "actionT", "Toggle timer").
		WithKeyBinding(" ", "actionSpace", "Toggle/Select").
		WithKeyBinding("n", "notify", "Show notification").
		WithKeyBinding("enter", "select", "Select item").
		WithKeyBinding("h", "goHome", "Go to home").
		WithKeyBinding("esc", "focusSidebar", "Focus sidebar").
		WithKeyBinding("g", "actionG", "Go to top").
		WithKeyBinding("G", "actionShiftG", "Go to bottom").
		WithKeyBinding("1", "action1", "Action 1").
		WithKeyBinding("2", "action2", "Action 2").
		WithKeyBinding("3", "action3", "Action 3").
		WithKeyBinding("4", "action4", "Action 4").
		WithKeyBinding("a", "actionA", "Add").
		WithKeyBinding("d", "actionD", "Delete/Debug").
		WithKeyBinding("c", "actionC", "Clear").
		WithKeyBinding("e", "actionE", "Enqueue/Error").
		WithKeyBinding("i", "actionI", "Info").
		WithKeyBinding("w", "actionW", "Warning").
		WithKeyBinding("s", "actionS", "Success").
		// Handle window resize messages from Bubbletea
		WithMessageHandler(func(comp bubbly.Component, msg tea.Msg) tea.Cmd {
			switch msg := msg.(type) {
			case tea.WindowSizeMsg:
				// Emit resize event with actual terminal dimensions
				comp.Emit("resize", map[string]int{
					"width":  msg.Width,
					"height": msg.Height,
				})
				return nil
			}
			return nil
		}).
		Setup(func(ctx *bubbly.Context) {
			ctx.ProvideTheme(bubbly.DefaultTheme)
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)

			state := localComposables.UseSharedDemoState(ctx)
			ctx.Expose("state", state)

			sidebarList, _ := localComponents.CreateSidebarList()
			contentArea, _ := localComponents.CreateContentArea()

			_ = ctx.ExposeComponent("sidebarList", sidebarList)
			_ = ctx.ExposeComponent("contentArea", contentArea)

			ctx.On("resize", func(data interface{}) {
				if sizeData, ok := data.(map[string]int); ok {
					state.SetSize(sizeData["width"], sizeData["height"])
				}
			})

			ctx.On("navUp", func(_ interface{}) {
				view := state.ActiveView.GetTyped()
				if state.FocusedPane.GetTyped() == localComposables.FocusSidebar {
					state.SidebarUp()
				} else {
					// View-specific up navigation
					switch view {
					case localComposables.ViewUseScroll:
						state.ScrollDemoUp()
					case localComposables.ViewUseSelection:
						state.SelectionDemoUp()
					case localComposables.ViewUseList:
						state.ListDemoUp()
					}
				}
			})

			ctx.On("navDown", func(_ interface{}) {
				view := state.ActiveView.GetTyped()
				if state.FocusedPane.GetTyped() == localComposables.FocusSidebar {
					state.SidebarDown()
				} else {
					// View-specific down navigation
					switch view {
					case localComposables.ViewUseScroll:
						state.ScrollDemoDown()
					case localComposables.ViewUseSelection:
						state.SelectionDemoDown()
					case localComposables.ViewUseList:
						state.ListDemoDown()
					}
				}
			})

			ctx.On("cycleFocus", func(_ interface{}) {
				view := state.ActiveView.GetTyped()
				if view == localComposables.ViewUseFocus && state.FocusedPane.GetTyped() == localComposables.FocusMain {
					state.FocusDemoNext()
				} else {
					state.CycleFocus()
				}
			})

			ctx.On("select", func(_ interface{}) {
				if state.FocusedPane.GetTyped() == localComposables.FocusSidebar {
					state.SelectSidebarItem()
				}
			})

			ctx.On("increment", func(_ interface{}) {
				view := state.ActiveView.GetTyped()
				if state.FocusedPane.GetTyped() == localComposables.FocusMain {
					switch view {
					case localComposables.ViewUseCounter:
						state.LocalCounterIncrement()
					case localComposables.ViewUsePrevious, localComposables.ViewUseHistory,
						localComposables.ViewHome, localComposables.ViewCreateShared,
						localComposables.ViewCreateSharedReset:
						state.Increment()
					}
				}
			})

			ctx.On("decrement", func(_ interface{}) {
				view := state.ActiveView.GetTyped()
				if state.FocusedPane.GetTyped() == localComposables.FocusMain {
					switch view {
					case localComposables.ViewUseCounter:
						state.LocalCounterDecrement()
					case localComposables.ViewUsePrevious, localComposables.ViewUseHistory,
						localComposables.ViewHome, localComposables.ViewCreateShared,
						localComposables.ViewCreateSharedReset:
						state.Decrement()
					}
				}
			})

			ctx.On("undo", func(_ interface{}) {
				view := state.ActiveView.GetTyped()
				if state.FocusedPane.GetTyped() == localComposables.FocusMain {
					if view == localComposables.ViewUseHistory || view == localComposables.ViewHome {
						state.Undo()
					}
				}
			})

			ctx.On("actionR", func(_ interface{}) {
				view := state.ActiveView.GetTyped()
				if state.FocusedPane.GetTyped() == localComposables.FocusMain {
					switch view {
					case localComposables.ViewUseHistory, localComposables.ViewHome:
						state.Redo()
					case localComposables.ViewUseCounter:
						state.LocalCounterReset()
					case localComposables.ViewUseInterval:
						state.IntervalReset()
					case localComposables.ViewUseTimeout:
						state.TimeoutReset()
					case localComposables.ViewUseTimer:
						state.TimerReset()
					case localComposables.ViewCreateSharedReset:
						state.SharedReset()
					}
				}
			})

			ctx.On("actionT", func(_ interface{}) {
				view := state.ActiveView.GetTyped()
				if state.FocusedPane.GetTyped() == localComposables.FocusMain {
					switch view {
					case localComposables.ViewHome, localComposables.ViewUseTimer:
						state.ToggleTimer()
					case localComposables.ViewUseSet:
						state.SetDemoToggle()
					}
				}
			})

			ctx.On("actionSpace", func(_ interface{}) {
				view := state.ActiveView.GetTyped()
				if state.FocusedPane.GetTyped() == localComposables.FocusMain {
					switch view {
					case localComposables.ViewHome:
						state.ToggleDarkMode()
					case localComposables.ViewUseSelection:
						state.SelectionDemoToggle()
					case localComposables.ViewUseInterval:
						state.IntervalToggle()
					case localComposables.ViewUseTimeout:
						state.TimeoutStart()
					case localComposables.ViewUseTimer:
						state.ToggleTimer()
					}
				}
			})

			ctx.On("notify", func(_ interface{}) {
				view := state.ActiveView.GetTyped()
				if state.FocusedPane.GetTyped() == localComposables.FocusMain {
					if view == localComposables.ViewHome {
						val := state.CounterValue.GetTyped()
						state.ShowNotification("Counter", strings.Repeat("*", val/10))
					}
				}
			})

			ctx.On("goHome", func(_ interface{}) {
				state.GoHome()
			})

			ctx.On("focusSidebar", func(_ interface{}) {
				state.FocusedPane.Set(localComposables.FocusSidebar)
			})

			ctx.On("actionG", func(_ interface{}) {
				view := state.ActiveView.GetTyped()
				if state.FocusedPane.GetTyped() == localComposables.FocusMain {
					if view == localComposables.ViewUseScroll {
						state.ScrollDemoTop()
					}
				}
			})

			ctx.On("actionShiftG", func(_ interface{}) {
				view := state.ActiveView.GetTyped()
				if state.FocusedPane.GetTyped() == localComposables.FocusMain {
					if view == localComposables.ViewUseScroll {
						state.ScrollDemoBottom()
					}
				}
			})

			ctx.On("action1", func(_ interface{}) {
				view := state.ActiveView.GetTyped()
				if state.FocusedPane.GetTyped() == localComposables.FocusMain {
					switch view {
					case localComposables.ViewUseFocus:
						state.FocusDemoSet(1)
					case localComposables.ViewUseMode:
						state.ModeDemoSet(1)
					case localComposables.ViewUseToggle:
						state.ToggleDemoToggle(1)
					}
				}
			})

			ctx.On("action2", func(_ interface{}) {
				view := state.ActiveView.GetTyped()
				if state.FocusedPane.GetTyped() == localComposables.FocusMain {
					switch view {
					case localComposables.ViewUseFocus:
						state.FocusDemoSet(2)
					case localComposables.ViewUseMode:
						state.ModeDemoSet(2)
					case localComposables.ViewUseToggle:
						state.ToggleDemoToggle(2)
					}
				}
			})

			ctx.On("action3", func(_ interface{}) {
				view := state.ActiveView.GetTyped()
				if state.FocusedPane.GetTyped() == localComposables.FocusMain {
					switch view {
					case localComposables.ViewUseFocus:
						state.FocusDemoSet(3)
					case localComposables.ViewUseMode:
						state.ModeDemoSet(3)
					case localComposables.ViewUseToggle:
						state.ToggleDemoToggle(3)
					}
				}
			})

			ctx.On("action4", func(_ interface{}) {
				view := state.ActiveView.GetTyped()
				if state.FocusedPane.GetTyped() == localComposables.FocusMain {
					if view == localComposables.ViewUseMode {
						state.ModeDemoSet(4)
					}
				}
			})

			ctx.On("actionA", func(_ interface{}) {
				view := state.ActiveView.GetTyped()
				if state.FocusedPane.GetTyped() == localComposables.FocusMain {
					switch view {
					case localComposables.ViewUseList:
						state.ListDemoAdd()
					case localComposables.ViewUseMap:
						state.MapDemoAdd()
					case localComposables.ViewUseSet:
						state.SetDemoAdd()
					}
				}
			})

			ctx.On("actionD", func(_ interface{}) {
				view := state.ActiveView.GetTyped()
				if state.FocusedPane.GetTyped() == localComposables.FocusMain {
					switch view {
					case localComposables.ViewUseList:
						state.ListDemoDelete()
					case localComposables.ViewUseMap:
						state.MapDemoDelete()
					case localComposables.ViewUseSet:
						state.SetDemoDelete()
					case localComposables.ViewUseQueue:
						state.QueueDemoDequeue()
					case localComposables.ViewUseLogger:
						state.LoggerDemoLog("DEBUG", "Debug message logged")
					}
				}
			})

			ctx.On("actionC", func(_ interface{}) {
				view := state.ActiveView.GetTyped()
				if state.FocusedPane.GetTyped() == localComposables.FocusMain {
					switch view {
					case localComposables.ViewUseList:
						state.ListDemoClear()
					case localComposables.ViewUseMap:
						state.MapDemoClear()
					case localComposables.ViewUseSet:
						state.SetDemoClear()
					case localComposables.ViewUseQueue:
						state.QueueDemoClear()
					case localComposables.ViewUseLogger:
						state.LoggerDemoClear()
					case localComposables.ViewUseNotification:
						state.NotificationDemoClear()
					}
				}
			})

			ctx.On("actionE", func(_ interface{}) {
				view := state.ActiveView.GetTyped()
				if state.FocusedPane.GetTyped() == localComposables.FocusMain {
					switch view {
					case localComposables.ViewUseQueue:
						state.QueueDemoEnqueue()
					case localComposables.ViewUseLogger:
						state.LoggerDemoLog("ERROR", "Error message logged")
					case localComposables.ViewUseNotification:
						state.NotificationDemoShow("error")
					}
				}
			})

			ctx.On("actionI", func(_ interface{}) {
				view := state.ActiveView.GetTyped()
				if state.FocusedPane.GetTyped() == localComposables.FocusMain {
					switch view {
					case localComposables.ViewUseLogger:
						state.LoggerDemoLog("INFO", "Info message logged")
					case localComposables.ViewUseNotification:
						state.NotificationDemoShow("info")
					}
				}
			})

			ctx.On("actionW", func(_ interface{}) {
				view := state.ActiveView.GetTyped()
				if state.FocusedPane.GetTyped() == localComposables.FocusMain {
					switch view {
					case localComposables.ViewUseLogger:
						state.LoggerDemoLog("WARN", "Warning message logged")
					case localComposables.ViewUseNotification:
						state.NotificationDemoShow("warning")
					}
				}
			})

			ctx.On("actionS", func(_ interface{}) {
				view := state.ActiveView.GetTyped()
				if state.FocusedPane.GetTyped() == localComposables.FocusMain {
					if view == localComposables.ViewUseNotification {
						state.NotificationDemoShow("success")
					}
				}
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
			activeView := state.ActiveView.GetTyped()

			sidebarList := ctx.Get("sidebarList").(bubbly.Component)
			contentArea := ctx.Get("contentArea").(bubbly.Component)

			// Update header based on active view
			headerTitle := "Home"
			if activeView != localComposables.ViewHome {
				headerTitle = string(activeView)
			}

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
			headerContent := "Enhanced Composables Demo" + modeStyle.Render(" | "+headerTitle)
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

			mainBox := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(mainBorderColor).
				Width(contentWidth).
				Height(mainHeight-2).
				Padding(0, 1)

			mainRendered := mainBox.Render(contentArea.View())

			mainArea := lipgloss.JoinHorizontal(lipgloss.Top, sidebarRendered, " ", mainRendered)

			// Dynamic status bar based on active view
			statusText := getStatusTextForView(activeView, focusedPane)

			statusStyle := lipgloss.NewStyle().
				Foreground(theme.Primary).
				Padding(0, 1)

			statusRendered := statusStyle.Render(statusText)

			notifications := state.Notifications.Notifications.GetTyped()
			var notifOverlay string
			if len(notifications) > 0 {
				var notifLines []string
				for _, n := range notifications {
					// Use notification type to determine color
					var bgColor lipgloss.Color
					var icon string
					switch n.Type {
					case composables.NotificationSuccess:
						bgColor = lipgloss.Color("22") // Green
						icon = "✓"
					case composables.NotificationError:
						bgColor = lipgloss.Color("124") // Red
						icon = "✗"
					case composables.NotificationWarning:
						bgColor = lipgloss.Color("208") // Orange/Yellow
						icon = "⚠"
					case composables.NotificationInfo:
						bgColor = lipgloss.Color("33") // Blue
						icon = "ℹ"
					default:
						bgColor = lipgloss.Color("240") // Gray
						icon = "●"
					}
					notifStyle := lipgloss.NewStyle().
						Background(bgColor).
						Foreground(lipgloss.Color("15")).
						Padding(0, 1)
					notifLines = append(notifLines, notifStyle.Render(icon+" "+n.Title+": "+n.Message))
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
