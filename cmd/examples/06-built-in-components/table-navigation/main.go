package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// User represents a user in our table
type User struct {
	ID     int
	Name   string
	Email  string
	Status string
}

type model struct {
	table          bubbly.Component
	selectedUser   string
	selectedStatus string
}

func initialModel() model {
	// Sample data
	users := []User{
		{ID: 1, Name: "Alice Johnson", Email: "alice@example.com", Status: "Active"},
		{ID: 2, Name: "Bob Smith", Email: "bob@example.com", Status: "Active"},
		{ID: 3, Name: "Charlie Brown", Email: "charlie@example.com", Status: "Inactive"},
		{ID: 4, Name: "Diana Prince", Email: "diana@example.com", Status: "Active"},
		{ID: 5, Name: "Eve Wilson", Email: "eve@example.com", Status: "Active"},
	}

	usersRef := bubbly.NewRef(users)

	// Create table with keyboard navigation
	table := components.Table(components.TableProps[User]{
		Data: usersRef,
		Columns: []components.TableColumn[User]{
			{Header: "ID", Field: "ID", Width: 5},
			{Header: "Name", Field: "Name", Width: 20},
			{Header: "Email", Field: "Email", Width: 25},
			{Header: "Status", Field: "Status", Width: 10},
		},
		OnRowClick: func(user User, index int) {
			// This will be called when user presses Enter or clicks a row
			// In a real app, you might open a detail view or edit form
		},
	})

	table.Init()

	return model{
		table:          table,
		selectedUser:   "None",
		selectedStatus: "Use ↑/↓ or k/j to navigate, Enter to select",
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			// Navigate up in table
			m.table.Emit("keyUp", nil)
			m.selectedStatus = "Navigating..."

		case "down", "j":
			// Navigate down in table
			m.table.Emit("keyDown", nil)
			m.selectedStatus = "Navigating..."

		case "enter", " ":
			// Confirm selection
			m.table.Emit("keyEnter", nil)
			m.selectedStatus = "Row selected! (callback triggered)"
		}
	}

	// Update table component
	_, cmd := m.table.Update(msg)

	return m, cmd
}

func (m model) View() string {
	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99")).
		MarginBottom(1)

	title := titleStyle.Render("🎯 Table Keyboard Navigation Demo")

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		MarginTop(1).
		MarginBottom(1)

	help := helpStyle.Render(
		"Controls:\n" +
			"  ↑/↓ or k/j : Navigate rows\n" +
			"  Enter/Space : Select row\n" +
			"  q or Ctrl+C : Quit",
	)

	// Status
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("35")).
		Bold(true).
		MarginTop(1)

	status := statusStyle.Render(fmt.Sprintf("Status: %s", m.selectedStatus))

	// Combine all parts
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		help,
		m.table.View(),
		status,
	)
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
