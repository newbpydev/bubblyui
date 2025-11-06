package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Todo represents a single todo item
type Todo struct {
	ID          int
	Title       string
	Description string
	Priority    string // "low", "medium", "high"
	Completed   bool
}

// model holds the application state for pure Bubbletea
type model struct {
	// Todo list state
	todos         []Todo
	nextID        int
	selectedIndex int

	// Input mode and form state
	inputMode    bool
	editMode     bool
	focusedField string // "Title", "Description", or "Priority"

	// Form fields
	titleInput    string
	descInput     string
	priorityInput string
}

// Init initializes the model
func (m model) Init() tea.Cmd {
	return nil
}

// Update handles all messages and updates the model
// This is where ALL the boilerplate lives in pure Bubbletea
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle special keys first
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc":
			// Toggle mode
			m.inputMode = !m.inputMode
			if !m.inputMode {
				// Exit edit mode when leaving input mode
				m.editMode = false
				m.titleInput = ""
				m.descInput = ""
				m.priorityInput = "medium"
			}

		case "ctrl+n":
			// New todo - only in navigation mode
			if !m.inputMode {
				m.editMode = false
				m.inputMode = true
				m.titleInput = ""
				m.descInput = ""
				m.priorityInput = "medium"
				m.focusedField = "Title"
			}

		case "ctrl+e":
			// Edit selected todo - only in navigation mode
			if !m.inputMode && !m.editMode && len(m.todos) > 0 {
				if m.selectedIndex >= 0 && m.selectedIndex < len(m.todos) {
					todo := m.todos[m.selectedIndex]
					m.editMode = true
					m.inputMode = true
					m.titleInput = todo.Title
					m.descInput = todo.Description
					m.priorityInput = todo.Priority
					m.focusedField = "Title"
				}
			}

		case "ctrl+d":
			// Delete selected todo - only in navigation mode
			if !m.inputMode && !m.editMode && len(m.todos) > 0 {
				if m.selectedIndex >= 0 && m.selectedIndex < len(m.todos) {
					m.todos = append(m.todos[:m.selectedIndex], m.todos[m.selectedIndex+1:]...)
					// Adjust selection
					if m.selectedIndex >= len(m.todos) && len(m.todos) > 0 {
						m.selectedIndex = len(m.todos) - 1
					}
				}
			}

		case "up":
			// Move selection up - only in navigation mode
			if !m.inputMode && !m.editMode {
				if m.selectedIndex > 0 {
					m.selectedIndex--
				}
			}

		case "down":
			// Move selection down - only in navigation mode
			if !m.inputMode && !m.editMode {
				if m.selectedIndex < len(m.todos)-1 {
					m.selectedIndex++
				}
			}

		case "enter":
			if m.inputMode {
				// In input mode: submit form
				if len(m.titleInput) >= 3 {
					if m.editMode {
						// Update existing todo
						if m.selectedIndex >= 0 && m.selectedIndex < len(m.todos) {
							m.todos[m.selectedIndex].Title = m.titleInput
							m.todos[m.selectedIndex].Description = m.descInput
							m.todos[m.selectedIndex].Priority = m.priorityInput
						}
						m.editMode = false
						m.inputMode = false
					} else {
						// Add new todo
						newTodo := Todo{
							ID:          m.nextID,
							Title:       m.titleInput,
							Description: m.descInput,
							Priority:    m.priorityInput,
							Completed:   false,
						}
						m.todos = append(m.todos, newTodo)
						m.nextID++
					}
					// Clear form
					m.titleInput = ""
					m.descInput = ""
					m.priorityInput = "medium"
				}
			} else {
				// In navigation mode: enter input mode to add new todo
				m.inputMode = true
				m.editMode = false
				m.focusedField = "Title"
			}

		case "tab":
			// Cycle through form fields - only in input mode
			if m.inputMode {
				switch m.focusedField {
				case "Title":
					m.focusedField = "Description"
				case "Description":
					m.focusedField = "Priority"
				case "Priority":
					m.focusedField = "Title"
				}
			}

		case "backspace":
			// Remove character - only in input mode
			if m.inputMode {
				switch m.focusedField {
				case "Title":
					if len(m.titleInput) > 0 {
						m.titleInput = m.titleInput[:len(m.titleInput)-1]
					}
				case "Description":
					if len(m.descInput) > 0 {
						m.descInput = m.descInput[:len(m.descInput)-1]
					}
				case "Priority":
					if len(m.priorityInput) > 0 {
						m.priorityInput = m.priorityInput[:len(m.priorityInput)-1]
					}
				}
			}

		default:
			// Handle space key separately (mode-dependent)
			if msg.Type == tea.KeySpace {
				if !m.inputMode && !m.editMode {
					// Navigation mode: toggle completion
					if len(m.todos) > 0 && m.selectedIndex >= 0 && m.selectedIndex < len(m.todos) {
						m.todos[m.selectedIndex].Completed = !m.todos[m.selectedIndex].Completed
					}
				} else if m.inputMode {
					// Input mode: add space character
					switch m.focusedField {
					case "Title":
						m.titleInput += " "
					case "Description":
						m.descInput += " "
					case "Priority":
						m.priorityInput += " "
					}
				}
			} else if m.inputMode {
				// Handle text input - only in input mode
				if msg.Type == tea.KeyRunes {
					char := string(msg.Runes)
					switch m.focusedField {
					case "Title":
						m.titleInput += char
					case "Description":
						m.descInput += char
					case "Priority":
						m.priorityInput += char
					}
				}
			}
		}
	}

	return m, nil
}

// View renders the UI
func (m model) View() string {
	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)
	titleText := titleStyle.Render("📝 Todo App - Pure Bubbletea")

	// Statistics
	totalCount := len(m.todos)
	completedCount := 0
	for _, todo := range m.todos {
		if todo.Completed {
			completedCount++
		}
	}
	pendingCount := totalCount - completedCount

	statsStyle := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Width(70)
	stats := statsStyle.Render(fmt.Sprintf(
		"📊 Total: %d | ✅ Completed: %d | ⏳ Pending: %d",
		totalCount, completedCount, pendingCount,
	))

	// Form box - dynamic border color based on mode
	formBorderColor := "240" // Dark grey (navigation mode - inactive)
	if m.inputMode {
		formBorderColor = "35" // Green (input mode - active)
	}
	formStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(formBorderColor)).
		Width(70)

	// Build form fields
	var formFields []string

	// Title field
	titleLabel := "Title:"
	if m.focusedField == "Title" {
		titleLabel = "▶ " + titleLabel
	} else {
		titleLabel = "  " + titleLabel
	}
	titleValue := m.titleInput
	if titleValue == "" {
		titleValue = "(empty)"
	}
	titleError := ""
	if len(m.titleInput) > 0 && len(m.titleInput) < 3 {
		titleError = " ❌ Must be at least 3 characters"
	}
	formFields = append(formFields, titleLabel+" "+titleValue+titleError)

	// Description field
	descLabel := "Description:"
	if m.focusedField == "Description" {
		descLabel = "▶ " + descLabel
	} else {
		descLabel = "  " + descLabel
	}
	descValue := m.descInput
	if descValue == "" {
		descValue = "(empty)"
	}
	formFields = append(formFields, descLabel+" "+descValue)

	// Priority field
	priorityLabel := "Priority:"
	if m.focusedField == "Priority" {
		priorityLabel = "▶ " + priorityLabel
	} else {
		priorityLabel = "  " + priorityLabel
	}
	priorityValue := m.priorityInput
	if priorityValue == "" {
		priorityValue = "(empty)"
	}
	formFields = append(formFields, priorityLabel+" "+priorityValue)

	// Form status
	formStatus := ""
	if len(m.titleInput) >= 3 {
		formStatus = "\n✓ Ready to submit"
	} else if len(m.titleInput) > 0 {
		formStatus = "\n❌ Title too short"
	}

	formBox := formStyle.Render(strings.Join(formFields, "\n") + formStatus)

	// Todo list - dynamic border color based on mode
	todoBorderColor := "99" // Purple (navigation mode - active)
	if m.inputMode {
		todoBorderColor = "240" // Dark grey (input mode - inactive)
	}
	todoStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(todoBorderColor)).
		Width(70)

	var todoItems []string
	if len(m.todos) == 0 {
		todoItems = append(todoItems, "No todos yet. Press Ctrl+N to create one!")
	} else {
		for i, todo := range m.todos {
			cursor := "  "
			if i == m.selectedIndex {
				cursor = "▶ "
			}

			checkbox := "☐"
			if todo.Completed {
				checkbox = "☑"
			}

			priorityIcon := ""
			switch todo.Priority {
			case "high":
				priorityIcon = "🔴"
			case "medium":
				priorityIcon = "🟡"
			case "low":
				priorityIcon = "🟢"
			}

			titleText := todo.Title
			if todo.Completed {
				titleText = lipgloss.NewStyle().
					Foreground(lipgloss.Color("240")).
					Strikethrough(true).
					Render(titleText)
			}

			todoItems = append(todoItems, fmt.Sprintf(
				"%s%s %s %s",
				cursor, checkbox, priorityIcon, titleText,
			))
		}
	}

	todoBox := todoStyle.Render(strings.Join(todoItems, "\n"))

	// Mode indicator
	modeStyle := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1).
		MarginTop(1)

	var modeIndicator string
	if m.inputMode {
		modeStyle = modeStyle.
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("35"))
		if m.editMode {
			modeIndicator = modeStyle.Render("✏️  EDIT MODE - Type to edit, ESC to cancel")
		} else {
			modeIndicator = modeStyle.Render("✍️  INPUT MODE - Type to add todo, ESC to navigate")
		}
	} else {
		modeStyle = modeStyle.
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("99"))
		modeIndicator = modeStyle.Render("🧭 NAVIGATION MODE - Use shortcuts, ENTER to add todo")
	}

	// Help text (manual - must keep in sync with Update())
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	var help string
	if m.inputMode {
		help = helpStyle.Render("tab: next field • enter: save • esc: cancel • ctrl+c: quit")
	} else {
		help = helpStyle.Render("↑/↓: select • space: toggle • ctrl+e: edit • ctrl+d: delete • ctrl+n: new • enter: add • ctrl+c: quit")
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		titleText,
		"",
		stats,
		"",
		formBox,
		"",
		todoBox,
		"",
		modeIndicator,
		help,
	)
}

func main() {
	// Pure Bubbletea: Manual model creation and initialization
	m := model{
		todos:         []Todo{},
		nextID:        1,
		selectedIndex: 0,
		inputMode:     false,
		editMode:      false,
		focusedField:  "Title",
		titleInput:    "",
		descInput:     "",
		priorityInput: "medium",
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
