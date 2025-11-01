package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// tickMsg is sent on each timer tick
type tickMsg time.Time

// tickCmd returns a command that waits for the next timer tick
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// model wraps the timer component
type model struct {
	component bubbly.Component
}

func (m model) Init() tea.Cmd {
	// Start both component init and timer tick
	return tea.Batch(m.component.Init(), tickCmd())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "space":
			m.component.Emit("toggle-timer", nil)
		case "r":
			m.component.Emit("reset", nil)
		}
	case tickMsg:
		// Forward tick to component
		m.component.Emit("tick", time.Time(msg))
		// Return next tick command to keep timer running
		cmds = append(cmds, tickCmd())
	}

	updatedComponent, cmd := m.component.Update(msg)
	m.component = updatedComponent.(bubbly.Component)

	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	title := titleStyle.Render("⏱️  Lifecycle Hooks - Timer/Interval Example")

	componentView := m.component.View()

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(2)

	help := helpStyle.Render(
		"space: toggle timer • r: reset • q: quit",
	)

	return fmt.Sprintf("%s\n\n%s\n%s\n", title, componentView, help)
}

// createTimerDemo creates a component demonstrating timer/interval management with lifecycle hooks
func createTimerDemo() (bubbly.Component, error) {
	return bubbly.NewComponent("TimerDemo").
		Setup(func(ctx *bubbly.Context) {
			// State
			elapsed := ctx.Ref(0)
			running := ctx.Ref(false)
			tickCount := ctx.Ref(0)
			events := ctx.Ref([]string{})

			// Helper to add event
			addEvent := func(event string) {
				current := events.GetTyped().([]string)
				if len(current) >= 8 {
					current = current[1:]
				}
				current = append(current, event)
				events.Set(current)
			}

			// onMounted: Start timer
			ctx.OnMounted(func() {
				addEvent("✅ onMounted: Component mounted")
				addEvent("⏱️  Starting timer...")
				running.Set(true)
				addEvent("✅ Timer started")
			})

			// onUpdated: Track running state changes
			ctx.OnUpdated(func() {
				isRunning := running.GetTyped().(bool)
				if isRunning {
					addEvent("▶️  Timer resumed")
				} else {
					addEvent("⏸️  Timer paused")
				}
			}, running)

			// onUpdated: Track elapsed time milestones
			ctx.OnUpdated(func() {
				elapsedVal := elapsed.GetTyped().(int)
				if elapsedVal > 0 && elapsedVal%10 == 0 {
					addEvent(fmt.Sprintf("🎯 Milestone: %d seconds!", elapsedVal))
				}
			}, elapsed)

			// onUnmounted: Cleanup will be called automatically
			ctx.OnUnmounted(func() {
				addEvent("🛑 onUnmounted: Component unmounting")
			})

			// Expose state
			ctx.Expose("elapsed", elapsed)
			ctx.Expose("running", running)
			ctx.Expose("tickCount", tickCount)
			ctx.Expose("events", events)

			// Event handlers
			ctx.On("tick", func(data interface{}) {
				if running.GetTyped().(bool) {
					current := elapsed.GetTyped().(int)
					elapsed.Set(current + 1)

					count := tickCount.GetTyped().(int)
					tickCount.Set(count + 1)
				}
			})

			ctx.On("toggle-timer", func(data interface{}) {
				isRunning := running.GetTyped().(bool)
				running.Set(!isRunning)
			})

			ctx.On("reset", func(data interface{}) {
				elapsed.Set(0)
				tickCount.Set(0)
				addEvent("🔄 Timer reset")
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Get state
			elapsed := ctx.Get("elapsed").(*bubbly.Ref[interface{}])
			running := ctx.Get("running").(*bubbly.Ref[interface{}])
			tickCount := ctx.Get("tickCount").(*bubbly.Ref[interface{}])
			events := ctx.Get("events").(*bubbly.Ref[interface{}])

			elapsedVal := elapsed.GetTyped().(int)
			runningVal := running.GetTyped().(bool)
			tickCountVal := tickCount.GetTyped().(int)
			eventsVal := events.GetTyped().([]string)

			// Format elapsed time
			minutes := elapsedVal / 60
			seconds := elapsedVal % 60
			timeStr := fmt.Sprintf("%02d:%02d", minutes, seconds)

			// Timer display box
			timerStyle := lipgloss.NewStyle().
				Bold(true).
				Padding(2, 4).
				Border(lipgloss.RoundedBorder()).
				Width(60).
				Align(lipgloss.Center)

			var timerBox string
			if runningVal {
				timerStyle = timerStyle.
					Foreground(lipgloss.Color("15")).
					Background(lipgloss.Color("35")).
					BorderForeground(lipgloss.Color("99"))
				timerBox = timerStyle.Render(fmt.Sprintf("⏱️  %s ▶️", timeStr))
			} else {
				timerStyle = timerStyle.
					Foreground(lipgloss.Color("15")).
					Background(lipgloss.Color("214")).
					BorderForeground(lipgloss.Color("208"))
				timerBox = timerStyle.Render(fmt.Sprintf("⏱️  %s ⏸️", timeStr))
			}

			// Stats box
			statsStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Padding(1, 2).
				Border(lipgloss.DoubleBorder()).
				BorderForeground(lipgloss.Color("240")).
				Width(60)

			statsBox := statsStyle.Render(fmt.Sprintf(
				"Timer Statistics:\n\n"+
					"Elapsed:    %d seconds\n"+
					"Tick Count: %d\n"+
					"Status:     %s\n"+
					"Minutes:    %d",
				elapsedVal,
				tickCountVal,
				map[bool]string{true: "Running", false: "Paused"}[runningVal],
				minutes,
			))

			// Events log box
			eventsStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("170")).
				Padding(1, 2).
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("141")).
				Width(60).
				Height(10)

			eventsStr := "Lifecycle Events:\n\n"
			for _, event := range eventsVal {
				eventsStr += event + "\n"
			}

			eventsBox := eventsStyle.Render(eventsStr)

			// Info box
			infoStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("99")).
				Padding(1, 2).
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240")).
				Width(60)

			infoBox := infoStyle.Render(
				"Lifecycle Hook Usage:\n" +
					"• onMounted: Start timer and register cleanup\n" +
					"• onUpdated: Track state changes\n" +
					"• OnCleanup: Stop timer and goroutine\n" +
					"• onUnmounted: Final cleanup",
			)

			return lipgloss.JoinVertical(
				lipgloss.Left,
				timerBox,
				"",
				statsBox,
				"",
				eventsBox,
				"",
				infoBox,
			)
		}).
		Build()
}

func main() {
	component, err := createTimerDemo()
	if err != nil {
		fmt.Printf("Error creating component: %v\n", err)
		os.Exit(1)
	}

	// Don't call component.Init() manually - Bubbletea will call model.Init()

	m := model{component: component}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}

	// Unmount component for cleanup demonstration
	if impl, ok := component.(interface{ Unmount() }); ok {
		impl.Unmount()
		fmt.Println("\n✅ Component unmounted - timer stopped and cleaned up")
	}
}
