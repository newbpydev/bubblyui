package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// model represents the application state using BubblyUI's reactive primitives
type model struct {
	count   *bubbly.Ref[int]
	doubled *bubbly.Computed[int]
	cleanup func()
}

// Init initializes the model and returns any initial commands
func (m model) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages and updates the model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			// Cleanup watchers before quitting
			if m.cleanup != nil {
				m.cleanup()
			}
			return m, tea.Quit

		case "up", "k", "+":
			// Increment counter - reactivity handles the rest!
			m.count.Set(m.count.Get() + 1)

		case "down", "j", "-":
			// Decrement counter
			if m.count.Get() > 0 {
				m.count.Set(m.count.Get() - 1)
			}

		case "r":
			// Reset counter
			m.count.Set(0)
		}
	}

	return m, nil
}

// View renders the UI based on the current model state
func (m model) View() string {
	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	counterStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63"))

	computedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Padding(1, 2).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("99"))

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	// Build the view using reactive values
	title := titleStyle.Render("🔢 Reactive Counter Example")

	counter := counterStyle.Render(
		fmt.Sprintf("Count: %d", m.count.Get()),
	)

	computed := computedStyle.Render(
		fmt.Sprintf("Doubled: %d", m.doubled.Get()),
	)

	help := helpStyle.Render(
		"↑/k/+ increment • ↓/j/- decrement • r reset • q quit",
	)

	// Compose the final view
	return fmt.Sprintf(
		"%s\n\n%s\n\n%s\n\n%s\n",
		title,
		counter,
		computed,
		help,
	)
}

func main() {
	// Create reactive state
	count := bubbly.NewRef(0)

	// Create computed value that automatically updates
	doubled := bubbly.NewComputed(func() int {
		return count.Get() * 2
	})

	// Optional: Watch for changes and log them
	cleanup := bubbly.Watch(count, func(newVal, oldVal int) {
		// This demonstrates side effects with watchers
		// In a real app, you might trigger Bubbletea commands here
		_ = newVal // Suppress unused warning
		_ = oldVal
	})

	// Create the model
	m := model{
		count:   count,
		doubled: doubled,
		cleanup: cleanup,
	}

	// Run the Bubbletea program
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
