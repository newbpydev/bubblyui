package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// model wraps the lifecycle demo component
type model struct {
	component bubbly.Component
}

func (m model) Init() tea.Cmd {
	return m.component.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "space", "enter":
			m.component.Emit("trigger-update", nil)
		case "r":
			m.component.Emit("reset", nil)
		}
	}

	updatedComponent, cmd := m.component.Update(msg)
	m.component = updatedComponent.(bubbly.Component)
	return m, cmd
}

func (m model) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	title := titleStyle.Render("🔄 Lifecycle Hooks - Basic Example")

	componentView := m.component.View()

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(2)

	help := helpStyle.Render(
		"space/enter: trigger update • r: reset • q: quit",
	)

	return fmt.Sprintf("%s\n\n%s\n%s\n", title, componentView, help)
}

// createLifecycleDemo creates a component demonstrating all lifecycle hooks
func createLifecycleDemo() (bubbly.Component, error) {
	return bubbly.NewComponent("LifecycleDemo").
		Setup(func(ctx *bubbly.Context) {
			// State
			updateCount := ctx.Ref(0)
			events := ctx.Ref([]string{"Component created"})

			// Helper to add event
			addEvent := func(event string) {
				current := events.Get().([]string)
				// Keep last 10 events
				if len(current) >= 10 {
					current = current[1:]
				}
				current = append(current, event)
				events.Set(current)
			}

			// onMounted: Runs after component is mounted
			ctx.OnMounted(func() {
				addEvent("✅ onMounted: Component mounted and ready")
			})

			// onMounted: Multiple hooks execute in order
			ctx.OnMounted(func() {
				addEvent("✅ onMounted: Second mounted hook")
			})

			// onUpdated: Runs after every update (no dependencies)
			ctx.OnUpdated(func() {
				count := updateCount.Get().(int)
				addEvent(fmt.Sprintf("🔄 onUpdated: Update #%d", count))
			})

			// onUpdated: With dependencies - only runs when updateCount changes
			ctx.OnUpdated(func() {
				count := updateCount.Get().(int)
				if count > 0 && count%5 == 0 {
					addEvent(fmt.Sprintf("🎯 onUpdated (deps): Milestone at %d updates!", count))
				}
			}, updateCount)

			// onUnmounted: Cleanup when component is removed
			ctx.OnUnmounted(func() {
				addEvent("🛑 onUnmounted: Component unmounting")
			})

			// Manual cleanup registration
			ctx.OnCleanup(func() {
				addEvent("🧹 Cleanup: Manual cleanup executed")
			})

			// Expose state
			ctx.Expose("updateCount", updateCount)
			ctx.Expose("events", events)

			// Event handlers
			ctx.On("trigger-update", func(data interface{}) {
				count := updateCount.Get().(int)
				updateCount.Set(count + 1)
			})

			ctx.On("reset", func(data interface{}) {
				updateCount.Set(0)
				events.Set([]string{"Component reset"})
				addEvent("🔄 State reset")
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Get state
			updateCount := ctx.Get("updateCount").(*bubbly.Ref[interface{}])
			events := ctx.Get("events").(*bubbly.Ref[interface{}])

			countVal := updateCount.Get().(int)
			eventsVal := events.Get().([]string)

			// Counter box
			counterStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("63")).
				Padding(1, 3).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("99")).
				Width(50).
				Align(lipgloss.Center)

			counterBox := counterStyle.Render(fmt.Sprintf("Update Count: %d", countVal))

			// Events log box
			eventsStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Padding(1, 2).
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240")).
				Width(50).
				Height(12)

			eventsStr := "Lifecycle Events:\n\n"
			for _, event := range eventsVal {
				eventsStr += event + "\n"
			}

			eventsBox := eventsStyle.Render(eventsStr)

			// Info box
			infoStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("170")).
				Padding(1, 2).
				Border(lipgloss.DoubleBorder()).
				BorderForeground(lipgloss.Color("141")).
				Width(50)

			infoBox := infoStyle.Render(
				"Lifecycle Hooks Demonstrated:\n" +
					"• onMounted: Initialization (runs once)\n" +
					"• onUpdated: React to changes\n" +
					"• onUpdated with deps: Conditional updates\n" +
					"• onUnmounted: Cleanup (on quit)\n" +
					"• OnCleanup: Manual cleanup registration",
			)

			return lipgloss.JoinVertical(
				lipgloss.Left,
				counterBox,
				"",
				eventsBox,
				"",
				infoBox,
			)
		}).
		Build()
}

func main() {
	component, err := createLifecycleDemo()
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
		fmt.Println("\n✅ Component unmounted - cleanup hooks executed")
	}
}
