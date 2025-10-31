package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
)

// UseCounterReturn is the return type for the UseCounter composable.
// It demonstrates a custom composable that wraps UseState with additional logic.
type UseCounterReturn struct {
	Count     *bubbly.Ref[int]
	Increment func()
	Decrement func()
	Reset     func()
	Double    func()
}

// UseCounter is a custom composable that provides counter functionality.
// This demonstrates:
// - Creating reusable composable functions
// - Composable composition (UseCounter uses UseState internally)
// - Returning multiple functions for different operations
//
// This is a COMPOSABLE CHAIN: UseCounter -> UseState -> Ref
func UseCounter(ctx *bubbly.Context, initial int) UseCounterReturn {
	// Use the standard UseState composable internally
	// This is composable composition - one composable calling another
	state := composables.UseState(ctx, initial)

	// Store initial value for reset functionality
	initialValue := initial

	return UseCounterReturn{
		Count: state.Value,
		Increment: func() {
			state.Set(state.Get() + 1)
		},
		Decrement: func() {
			state.Set(state.Get() - 1)
		},
		Reset: func() {
			state.Set(initialValue)
		},
		Double: func() {
			state.Set(state.Get() * 2)
		},
	}
}

// UseDoubleCounterReturn demonstrates a deeper composable chain.
type UseDoubleCounterReturn struct {
	Counter1  UseCounterReturn
	Counter2  UseCounterReturn
	SyncBoth  func()
	ResetBoth func()
}

// UseDoubleCounter demonstrates a DEEPER COMPOSABLE CHAIN.
// This composable uses UseCounter, which uses UseState, which uses Ref.
// Chain: UseDoubleCounter -> UseCounter -> UseState -> Ref
//
// This pattern is useful for:
// - Managing related state together
// - Providing coordinated operations
// - Building complex functionality from simple composables
func UseDoubleCounter(ctx *bubbly.Context, initial1, initial2 int) UseDoubleCounterReturn {
	// Create two independent counters using UseCounter
	// This demonstrates composable reuse and composition
	counter1 := UseCounter(ctx, initial1)
	counter2 := UseCounter(ctx, initial2)

	return UseDoubleCounterReturn{
		Counter1: counter1,
		Counter2: counter2,
		SyncBoth: func() {
			// Sync counter2 to counter1's value
			counter2.Count.Set(counter1.Count.GetTyped())
		},
		ResetBoth: func() {
			counter1.Reset()
			counter2.Reset()
		},
	}
}

// model wraps the counter component for Bubbletea integration
type model struct {
	component bubbly.Component
}

func (m model) Init() tea.Cmd {
	// CRITICAL: Let Bubbletea call Init, don't call it manually
	return m.component.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			m.component.Emit("increment1", nil)
		case "down", "j":
			m.component.Emit("decrement1", nil)
		case "w":
			m.component.Emit("increment2", nil)
		case "s":
			m.component.Emit("decrement2", nil)
		case "d":
			m.component.Emit("double1", nil)
		case "f":
			m.component.Emit("double2", nil)
		case "r":
			m.component.Emit("reset1", nil)
		case "t":
			m.component.Emit("reset2", nil)
		case "y":
			m.component.Emit("sync", nil)
		case "u":
			m.component.Emit("resetBoth", nil)
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

	title := titleStyle.Render("🔢 Composables - Counter Example")

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginBottom(1)

	subtitle := subtitleStyle.Render(
		"Demonstrates: Custom composables, composable chains, and composition patterns",
	)

	componentView := m.component.View()

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(2)

	help := helpStyle.Render(
		"Counter 1: ↑/k: +1 • ↓/j: -1 • d: double • r: reset\n" +
			"Counter 2: w: +1 • s: -1 • f: double • t: reset\n" +
			"Both: y: sync • u: reset both • q: quit",
	)

	return fmt.Sprintf("%s\n%s\n\n%s\n%s\n", title, subtitle, componentView, help)
}

// createCounterDemo creates a component demonstrating composable patterns
func createCounterDemo() (bubbly.Component, error) {
	return bubbly.NewComponent("CounterDemo").
		Setup(func(ctx *bubbly.Context) {
			// Use our custom UseDoubleCounter composable
			// This demonstrates the composable chain in action
			counters := UseDoubleCounter(ctx, 0, 10)

			// Expose state to template
			ctx.Expose("counter1", counters.Counter1.Count)
			ctx.Expose("counter2", counters.Counter2.Count)

			// Event handlers for Counter 1
			ctx.On("increment1", func(_ interface{}) {
				counters.Counter1.Increment()
			})

			ctx.On("decrement1", func(_ interface{}) {
				counters.Counter1.Decrement()
			})

			ctx.On("double1", func(_ interface{}) {
				counters.Counter1.Double()
			})

			ctx.On("reset1", func(_ interface{}) {
				counters.Counter1.Reset()
			})

			// Event handlers for Counter 2
			ctx.On("increment2", func(_ interface{}) {
				counters.Counter2.Increment()
			})

			ctx.On("decrement2", func(_ interface{}) {
				counters.Counter2.Decrement()
			})

			ctx.On("double2", func(_ interface{}) {
				counters.Counter2.Double()
			})

			ctx.On("reset2", func(_ interface{}) {
				counters.Counter2.Reset()
			})

			// Coordinated operations
			ctx.On("sync", func(_ interface{}) {
				counters.SyncBoth()
			})

			ctx.On("resetBoth", func(_ interface{}) {
				counters.ResetBoth()
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Get state
			counter1 := ctx.Get("counter1").(*bubbly.Ref[int])
			counter2 := ctx.Get("counter2").(*bubbly.Ref[int])

			count1 := counter1.GetTyped()
			count2 := counter2.GetTyped()

			// Counter 1 box
			counter1Style := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("63")).
				Padding(2, 4).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("99")).
				Width(40).
				Align(lipgloss.Center)

			counter1Box := counter1Style.Render(fmt.Sprintf("Counter 1: %d", count1))

			// Counter 2 box
			counter2Style := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("170")).
				Padding(2, 4).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("141")).
				Width(40).
				Align(lipgloss.Center)

			counter2Box := counter2Style.Render(fmt.Sprintf("Counter 2: %d", count2))

			// Info box explaining the pattern
			infoStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Padding(1, 2).
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240")).
				Width(40)

			infoBox := infoStyle.Render(
				"Composable Chain:\n" +
					"UseDoubleCounter\n" +
					"  └─> UseCounter (x2)\n" +
					"      └─> UseState\n" +
					"          └─> Ref\n\n" +
					"Pattern: Composables calling\n" +
					"other composables for reuse",
			)

			return lipgloss.JoinVertical(
				lipgloss.Left,
				counter1Box,
				"",
				counter2Box,
				"",
				infoBox,
			)
		}).
		Build()
}

func main() {
	component, err := createCounterDemo()
	if err != nil {
		fmt.Printf("Error creating component: %v\n", err)
		os.Exit(1)
	}

	// CRITICAL: Don't call component.Init() manually
	// Bubbletea will call model.Init() which calls component.Init()

	m := model{component: component}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
