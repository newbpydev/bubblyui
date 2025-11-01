package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// model wraps the counter component
type model struct {
	counter bubbly.Component
}

func (m model) Init() tea.Cmd {
	return m.counter.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k", "+":
			m.counter.Emit("increment", nil)
		case "down", "j", "-":
			m.counter.Emit("decrement", nil)
		case "r":
			m.counter.Emit("reset", nil)
		case "d":
			m.counter.Emit("double", nil)
		case "h":
			m.counter.Emit("halve", nil)
		}
	}

	updatedComponent, cmd := m.counter.Update(msg)
	m.counter = updatedComponent.(bubbly.Component)
	return m, cmd
}

func (m model) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	title := titleStyle.Render("🔢 Counter Component - State Management")

	componentView := m.counter.View()

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(2)

	help := helpStyle.Render(
		"↑/k/+: increment • ↓/j/-: decrement • d: double • h: halve • r: reset • q: quit",
	)

	return fmt.Sprintf("%s\n\n%s\n%s\n", title, componentView, help)
}

// createCounter creates a counter component with advanced state management
func createCounter() (bubbly.Component, error) {
	return bubbly.NewComponent("Counter").
		Setup(func(ctx *bubbly.Context) {
			// Reactive state
			count := ctx.Ref(0)
			history := ctx.Ref([]int{0})

			// Computed values
			doubled := ctx.Computed(func() interface{} {
				return count.GetTyped().(int) * 2
			})

			squared := ctx.Computed(func() interface{} {
				c := count.GetTyped().(int)
				return c * c
			})

			isEven := ctx.Computed(func() interface{} {
				return count.GetTyped().(int)%2 == 0
			})

			// Expose state and computed values
			ctx.Expose("count", count)
			ctx.Expose("history", history)
			ctx.Expose("doubled", doubled)
			ctx.Expose("squared", squared)
			ctx.Expose("isEven", isEven)

			// Helper to add to history
			addToHistory := func(newVal int) {
				hist := history.GetTyped().([]int)
				// Keep last 5 values
				if len(hist) >= 5 {
					hist = hist[1:]
				}
				hist = append(hist, newVal)
				history.Set(hist)
			}

			// Event handlers
			ctx.On("increment", func(data interface{}) {
				newVal := count.GetTyped().(int) + 1
				count.Set(newVal)
				addToHistory(newVal)
			})

			ctx.On("decrement", func(data interface{}) {
				newVal := count.GetTyped().(int) - 1
				count.Set(newVal)
				addToHistory(newVal)
			})

			ctx.On("reset", func(data interface{}) {
				count.Set(0)
				history.Set([]int{0})
			})

			ctx.On("double", func(data interface{}) {
				newVal := count.GetTyped().(int) * 2
				count.Set(newVal)
				addToHistory(newVal)
			})

			ctx.On("halve", func(data interface{}) {
				newVal := count.GetTyped().(int) / 2
				count.Set(newVal)
				addToHistory(newVal)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Get state
			count := ctx.Get("count").(*bubbly.Ref[interface{}])
			history := ctx.Get("history").(*bubbly.Ref[interface{}])
			doubled := ctx.Get("doubled").(*bubbly.Computed[interface{}])
			squared := ctx.Get("squared").(*bubbly.Computed[interface{}])
			isEven := ctx.Get("isEven").(*bubbly.Computed[interface{}])

			countVal := count.GetTyped().(int)
			historyVal := history.GetTyped().([]int)

			// Main counter box
			counterStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("63")).
				Padding(1, 3).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("99")).
				Width(30).
				Align(lipgloss.Center)

			counterBox := counterStyle.Render(fmt.Sprintf("Count: %d", countVal))

			// Computed values box
			computedStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("170")).
				Padding(1, 2).
				Border(lipgloss.DoubleBorder()).
				BorderForeground(lipgloss.Color("141")).
				Width(30)

			evenStr := "Odd"
			if isEven.GetTyped().(bool) {
				evenStr = "Even"
			}

			computedBox := computedStyle.Render(fmt.Sprintf(
				"Doubled: %d\nSquared: %d\nParity:  %s",
				doubled.GetTyped().(int),
				squared.GetTyped().(int),
				evenStr,
			))

			// History box
			historyStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Padding(1, 2).
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240")).
				Width(30)

			historyStr := "History: "
			for i, v := range historyVal {
				if i > 0 {
					historyStr += " → "
				}
				historyStr += fmt.Sprintf("%d", v)
			}

			historyBox := historyStyle.Render(historyStr)

			return lipgloss.JoinVertical(
				lipgloss.Left,
				counterBox,
				"",
				computedBox,
				"",
				historyBox,
			)
		}).
		Build()
}

func main() {
	counter, err := createCounter()
	if err != nil {
		fmt.Printf("Error creating counter: %v\n", err)
		os.Exit(1)
	}

	counter.Init()

	m := model{counter: counter}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
