package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// tickMsg is sent periodically to trigger UI updates while loading
// This is necessary because async operations update Refs in goroutines,
// but Bubbletea only redraws when Update() is called
type tickMsg time.Time

// tickCmd returns a command that ticks periodically
func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// model wraps the component with async tick support
type model struct {
	component bubbly.Component
	loading   bool // Track if any async operation is in progress
}

func (m model) Init() tea.Cmd {
	// Start ticking to handle initial async load
	return tea.Batch(m.component.Init(), tickCmd())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "r":
			// Refresh data - restart ticking
			m.component.Emit("refresh", nil)
			m.loading = true
			cmds = append(cmds, tickCmd())
		}
	case tickMsg:
		// Continue ticking while loading
		// This ensures UI updates while goroutines are running
		if m.loading {
			cmds = append(cmds, tickCmd())
		}
	}

	// Update component
	updatedComponent, cmd := m.component.Update(msg)
	m.component = updatedComponent.(bubbly.Component)

	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	return m.component.View()
}

func main() {
	// Create mock API with realistic delays
	mockAPI := NewMockGitHubAPI()
	mockAPI.SetDelay(500*time.Millisecond, 500*time.Millisecond)

	// Create app component
	app, err := CreateApp(mockAPI)
	if err != nil {
		fmt.Printf("Error creating app: %v\n", err)
		os.Exit(1)
	}

	// Wrap with async-aware model
	m := model{
		component: app,
		loading:   true,
	}

	// Run with Bubbletea
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running app: %v\n", err)
		os.Exit(1)
	}
}
