package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// Todo represents a single todo item
type Todo struct {
	Text      string
	Completed bool
}

// model represents the application state using BubblyUI's reactive primitives
type model struct {
	todos          *bubbly.Ref[[]Todo]
	totalCount     *bubbly.Computed[int]
	completedCount *bubbly.Computed[int]
	activeCount    *bubbly.Computed[int]
	cursor         int
	input          string
	inputMode      bool
}

// Init initializes the model
func (m model) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.inputMode {
			return m.handleInputMode(msg)
		}
		return m.handleNormalMode(msg)
	}
	return m, nil
}

// handleInputMode handles input when adding a new todo
func (m model) handleInputMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.input != "" {
			// Add new todo using reactive state
			todos := m.todos.Get()
			todos = append(todos, Todo{Text: m.input, Completed: false})
			m.todos.Set(todos)
			m.input = ""
		}
		m.inputMode = false

	case "esc":
		m.input = ""
		m.inputMode = false

	case "backspace":
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}

	default:
		// Add character to input
		if len(msg.String()) == 1 {
			m.input += msg.String()
		}
	}
	return m, nil
}

// handleNormalMode handles navigation and actions
func (m model) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	todos := m.todos.Get()

	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(todos)-1 {
			m.cursor++
		}

	case "enter", " ":
		// Toggle completion status
		if len(todos) > 0 && m.cursor < len(todos) {
			todos[m.cursor].Completed = !todos[m.cursor].Completed
			m.todos.Set(todos)
		}

	case "d":
		// Delete current todo
		if len(todos) > 0 && m.cursor < len(todos) {
			todos = append(todos[:m.cursor], todos[m.cursor+1:]...)
			m.todos.Set(todos)
			if m.cursor >= len(todos) && m.cursor > 0 {
				m.cursor--
			}
		}

	case "a", "n":
		// Enter input mode to add new todo
		m.inputMode = true
	}

	return m, nil
}

// View renders the UI
func (m model) View() string {
	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	todoStyle := lipgloss.NewStyle().
		PaddingLeft(2)

	cursorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true)

	completedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Strikethrough(true)

	statsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(0, 1).
		MarginTop(1)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	// Build view
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("📝 Reactive Todo List"))
	b.WriteString("\n\n")

	// Input mode
	if m.inputMode {
		b.WriteString("New todo: ")
		b.WriteString(m.input)
		b.WriteString("█\n\n")
		b.WriteString(helpStyle.Render("enter to add • esc to cancel"))
		return b.String()
	}

	// Todo list
	todos := m.todos.Get()
	if len(todos) == 0 {
		b.WriteString(todoStyle.Render("No todos yet. Press 'a' to add one!"))
	} else {
		for i, todo := range todos {
			cursor := "  "
			if i == m.cursor {
				cursor = cursorStyle.Render("→ ")
			}

			checkbox := "[ ]"
			text := todo.Text
			if todo.Completed {
				checkbox = "[✓]"
				text = completedStyle.Render(text)
			}

			line := fmt.Sprintf("%s%s %s", cursor, checkbox, text)
			b.WriteString(todoStyle.Render(line))
			b.WriteString("\n")
		}
	}

	// Stats using computed values - demonstrates reactivity!
	stats := fmt.Sprintf(
		"Total: %d | Active: %d | Completed: %d",
		m.totalCount.Get(),
		m.activeCount.Get(),
		m.completedCount.Get(),
	)
	b.WriteString("\n")
	b.WriteString(statsStyle.Render(stats))

	// Help
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(
		"↑/k up • ↓/j down • space/enter toggle • d delete • a/n add • q quit",
	))

	return b.String()
}

func main() {
	// Create reactive state with initial todos
	todos := bubbly.NewRef([]Todo{
		{Text: "Learn BubblyUI reactivity", Completed: true},
		{Text: "Build awesome TUI apps", Completed: false},
		{Text: "Share with the community", Completed: false},
	})

	// Create computed values that automatically update
	totalCount := bubbly.NewComputed(func() int {
		return len(todos.Get())
	})

	completedCount := bubbly.NewComputed(func() int {
		count := 0
		for _, todo := range todos.Get() {
			if todo.Completed {
				count++
			}
		}
		return count
	})

	activeCount := bubbly.NewComputed(func() int {
		return totalCount.Get() - completedCount.Get()
	})

	// Create model
	m := model{
		todos:          todos,
		totalCount:     totalCount,
		completedCount: completedCount,
		activeCount:    activeCount,
		cursor:         0,
	}

	// Run the program with alternate screen buffer
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
