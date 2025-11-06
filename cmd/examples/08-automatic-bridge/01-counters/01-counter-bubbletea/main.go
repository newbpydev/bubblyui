package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// model holds the application state for pure Bubbletea
type model struct {
	count int
}

// Init initializes the model (no commands needed for this simple app)
func (m model) Init() tea.Cmd {
	return nil
}

// Update handles all messages and updates the model
// This is where ALL the boilerplate lives in pure Bubbletea
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case " ": // Space key
			m.count++
		case "r":
			m.count = 0
		}
	}
	return m, nil
}

// View renders the UI
func (m model) View() string {
	// Styling - identical to BubblyUI version
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	counterStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("63")).
		Padding(2, 4).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Width(40).
		Align(lipgloss.Center)

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Padding(1, 2).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(40)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	// Render
	title := titleStyle.Render("🎯 Pure Bubbletea - Simple Counter")

	counterBox := counterStyle.Render(fmt.Sprintf("Count: %d", m.count))

	infoBox := infoStyle.Render(
		"Pure Bubbletea:\n" +
			"✓ Manual state management\n" +
			"✓ Manual key handling in Update()\n" +
			"✓ Manual help text\n" +
			"✓ Manual model wrapper",
	)

	// Manual help text (no auto-generation)
	help := helpStyle.Render(" : Increment counter • ctrl+c: Quit application • r: Reset to zero")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		counterBox,
		"",
		infoBox,
		"",
		help,
	)
}

func main() {
	// Pure Bubbletea: Manual model creation and initialization
	m := model{
		count: 0,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
