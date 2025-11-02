package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// Todo represents a single todo item
type Todo struct {
	ID        int
	Title     string
	Completed bool
}

// TodoInput represents the form input for creating todos
type TodoInput struct {
	Title string
}

// model wraps the todo component
type model struct {
	component    bubbly.Component
	selectedTodo int  // Index of selected todo in list
	inputMode    bool // Whether we're in input mode (typing in form)
}

func (m model) Init() tea.Cmd {
	// CRITICAL: Let Bubbletea call Init, don't call it manually
	// Return both component init and cursor blink command
	return tea.Batch(
		m.component.Init(),
		composables.BlinkCmd(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle space key first (using msg.Type)
		if msg.Type == tea.KeySpace {
			if !m.inputMode {
				// Navigation mode: toggle completion
				m.component.Emit("toggleTodo", m.selectedTodo)
			} else {
				// Input mode: add space character
				m.component.Emit("addChar", " ")
			}
		} else {
			// Handle other keys using msg.String()
			switch msg.String() {
			case "ctrl+c":
				// Always allow quit with Ctrl+C
				return m, tea.Quit
			case "esc":
				// ESC toggles between input mode and navigation mode
				m.inputMode = !m.inputMode
				m.component.Emit("setInputMode", m.inputMode)
			case "ctrl+n":
				// New todo - enter input mode
				m.inputMode = true
				m.component.Emit("setInputMode", m.inputMode)
				m.component.Emit("clearForm", nil)
			case "ctrl+d":
				// Delete selected todo - only in navigation mode
				if !m.inputMode {
					m.component.Emit("deleteTodo", m.selectedTodo)
				}
			case "enter":
				if m.inputMode {
					// In input mode: submit form
					m.component.Emit("addTodo", nil)
					// Stay in input mode for adding multiple todos
				} else {
					// In navigation mode: enter input mode to add new todo
					m.inputMode = true
					m.component.Emit("setInputMode", m.inputMode)
				}
			case "up":
				// Move selection up - only in navigation mode
				if !m.inputMode {
					m.component.Emit("selectPrevious", nil)
					if m.selectedTodo > 0 {
						m.selectedTodo--
					}
				}
			case "down":
				// Move selection down - only in navigation mode
				if !m.inputMode {
					m.component.Emit("selectNext", nil)
					m.selectedTodo++
				}
			case "k":
				// In navigation mode: move up; in input mode: forward to text input
				if !m.inputMode {
					m.component.Emit("selectPrevious", nil)
					if m.selectedTodo > 0 {
						m.selectedTodo--
					}
				} else {
					m.component.Emit("textInputUpdate", msg)
				}
			case "j":
				// In navigation mode: move down; in input mode: forward to text input
				if !m.inputMode {
					m.component.Emit("selectNext", nil)
					m.selectedTodo++
				} else {
					m.component.Emit("textInputUpdate", msg)
				}
			default:
				// Forward all other keys to text input in input mode
				// This includes typing, backspace, delete, arrow keys for cursor movement
				if m.inputMode {
					m.component.Emit("textInputUpdate", msg)
				}
			}
		}
	}

	updatedComponent, cmd := m.component.Update(msg)
	m.component = updatedComponent.(bubbly.Component)

	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

func (m model) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	title := titleStyle.Render("📝 Todo App - Built-in Components Demo")

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginBottom(1)

	subtitle := subtitleStyle.Render(
		"Demonstrates: Form + List components with reactive state",
	)

	componentView := m.component.View()

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
		modeIndicator = modeStyle.Render("✍️  INPUT MODE - Type to add todo, ESC to navigate")
	} else {
		modeStyle = modeStyle.
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("99"))
		modeIndicator = modeStyle.Render("🧭 NAVIGATION MODE - Use shortcuts, ENTER to add todo")
	}

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	var help string
	if m.inputMode {
		help = helpStyle.Render(
			"enter: save • esc: cancel • ctrl+c: quit",
		)
	} else {
		help = helpStyle.Render(
			"↑/↓/j/k: select • space: toggle • ctrl+d: delete • ctrl+n: new • enter: add • ctrl+c: quit",
		)
	}

	return fmt.Sprintf("%s\n%s\n\n%s\n%s\n%s\n", title, subtitle, componentView, modeIndicator, help)
}

// validateTodoInput validates the todo input form
func validateTodoInput(input TodoInput) map[string]string {
	errors := make(map[string]string)

	// Title validation
	if len(input.Title) == 0 {
		errors["Title"] = "Title is required"
	} else if len(input.Title) < 3 {
		errors["Title"] = "Title must be at least 3 characters"
	} else if len(input.Title) > 50 {
		errors["Title"] = "Title must be less than 50 characters"
	}

	return errors
}

// createTodoApp creates the todo application component
func createTodoApp() (bubbly.Component, error) {
	return bubbly.NewComponent("TodoApp").
		Setup(func(ctx *bubbly.Context) {
			// Provide theme for child components
			ctx.Provide("theme", components.DefaultTheme)

			// Todo list state
			todos := bubbly.NewRef([]Todo{})
			nextID := bubbly.NewRef(1)

			// Form input state with cursor support
			textInput := composables.UseTextInput(composables.UseTextInputConfig{
				Placeholder:        "Enter todo title...",
				CharLimit:          100,
				Width:              56, // 60 - padding
				ShowCursorPosition: true,
			})

			// UI state
			selectedIndex := bubbly.NewRef(0)
			inputMode := bubbly.NewRef(false) // Track input mode for visual feedback

			// Statistics
			totalCount := ctx.Computed(func() interface{} {
				todoList := todos.Get().([]Todo)
				return len(todoList)
			})

			completedCount := ctx.Computed(func() interface{} {
				todoList := todos.Get().([]Todo)
				count := 0
				for _, todo := range todoList {
					if todo.Completed {
						count++
					}
				}
				return count
			})

			pendingCount := ctx.Computed(func() interface{} {
				total := totalCount.Get().(int)
				completed := completedCount.Get().(int)
				return total - completed
			})

			// Note: We're not using the Form component directly here,
			// instead we're building a simpler input interface for this example.
			// The Form component would be used in more complex scenarios.

			// Create list component
			list := components.List(components.ListProps[Todo]{
				Items: todos,
				RenderItem: func(todo Todo, i int) string {
					cursor := "  "
					if i == selectedIndex.Get().(int) {
						cursor = "▶ "
					}

					checkbox := "☐"
					if todo.Completed {
						checkbox = "☑"
					}

					title := todo.Title
					if todo.Completed {
						title = lipgloss.NewStyle().
							Foreground(lipgloss.Color("240")).
							Strikethrough(true).
							Render(title)
					}

					return fmt.Sprintf("%s%s %s", cursor, checkbox, title)
				},
				Height: 10,
			})

			// Initialize child components
			list.Init()

			// Expose state to template
			ctx.Expose("list", list)
			ctx.Expose("textInput", textInput)
			ctx.Expose("todos", todos)
			ctx.Expose("selectedIndex", selectedIndex)
			ctx.Expose("inputMode", inputMode)
			ctx.Expose("totalCount", totalCount)
			ctx.Expose("completedCount", completedCount)
			ctx.Expose("pendingCount", pendingCount)

			// Event: Set input mode (called from model)
			ctx.On("setInputMode", func(data interface{}) {
				mode := data.(bool)
				inputMode.Set(mode)

				// Focus/blur text input based on mode
				if mode {
					textInput.Focus()
				} else {
					textInput.Blur()
				}
			})

			// Event: Add new todo
			ctx.On("addTodo", func(_ interface{}) {
				// Validate manually
				input := TodoInput{Title: textInput.Value()}
				errors := validateTodoInput(input)

				if len(errors) == 0 {
					todoList := todos.Get().([]Todo)
					id := nextID.Get().(int)

					newTodo := Todo{
						ID:        id,
						Title:     input.Title,
						Completed: false,
					}

					todos.Set(append(todoList, newTodo))
					nextID.Set(id + 1)
					textInput.Reset()
				}
			})

			// Event: Delete todo
			ctx.On("deleteTodo", func(data interface{}) {
				index := data.(int)
				todoList := todos.Get().([]Todo)
				if index >= 0 && index < len(todoList) {
					newList := append(todoList[:index], todoList[index+1:]...)
					todos.Set(newList)
					// Adjust selection if needed
					if selectedIndex.Get().(int) >= len(newList) && len(newList) > 0 {
						selectedIndex.Set(len(newList) - 1)
					}
				}
			})

			// Event: Toggle todo completion
			ctx.On("toggleTodo", func(data interface{}) {
				index := data.(int)
				todoList := todos.Get().([]Todo)
				if index >= 0 && index < len(todoList) {
					todoList[index].Completed = !todoList[index].Completed
					todos.Set(todoList)
				}
			})

			// Event: Select previous todo
			ctx.On("selectPrevious", func(_ interface{}) {
				current := selectedIndex.Get().(int)
				if current > 0 {
					selectedIndex.Set(current - 1)
				}
			})

			// Event: Select next todo
			ctx.On("selectNext", func(_ interface{}) {
				current := selectedIndex.Get().(int)
				todoList := todos.Get().([]Todo)
				if current < len(todoList)-1 {
					selectedIndex.Set(current + 1)
				}
			})

			// Event: Clear form
			ctx.On("clearForm", func(_ interface{}) {
				textInput.Reset()
			})

			// Event: Update text input (forward Bubbletea messages for cursor support)
			ctx.On("textInputUpdate", func(data interface{}) {
				msg := data.(tea.Msg)
				textInput.Update(msg)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Get state
			list := ctx.Get("list").(bubbly.Component)
			textInput := ctx.Get("textInput").(*composables.TextInputResult)
			inputModeRefRaw := ctx.Get("inputMode")
			totalCountRaw := ctx.Get("totalCount")
			completedCountRaw := ctx.Get("completedCount")
			pendingCountRaw := ctx.Get("pendingCount")

			// Type assert to correct types
			var inInputMode bool
			var totalCountVal, completedCountVal, pendingCountVal int

			if ref, ok := inputModeRefRaw.(*bubbly.Ref[bool]); ok {
				inInputMode = ref.Get().(bool)
			}
			if comp, ok := totalCountRaw.(*bubbly.Computed[interface{}]); ok {
				totalCountVal = comp.Get().(int)
			}
			if comp, ok := completedCountRaw.(*bubbly.Computed[interface{}]); ok {
				completedCountVal = comp.Get().(int)
			}
			if comp, ok := pendingCountRaw.(*bubbly.Computed[interface{}]); ok {
				pendingCountVal = comp.Get().(int)
			}

			// Statistics box
			statsStyle := lipgloss.NewStyle().
				Bold(true).
				Padding(0, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("99")).
				Width(60)

			stats := statsStyle.Render(fmt.Sprintf(
				"📊 Total: %d | ✅ Completed: %d | ⏳ Pending: %d",
				totalCountVal,
				completedCountVal,
				pendingCountVal,
			))

			// Form box - dynamic border color based on mode
			formBorderColor := "240" // Dark grey (navigation mode)
			if inInputMode {
				formBorderColor = "35" // Green (input mode)
			}
			formStyle := lipgloss.NewStyle().
				Padding(1, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(formBorderColor)).
				Width(60)

			// Build form display with text input (shows cursor)
			formDisplay := lipgloss.JoinVertical(
				lipgloss.Left,
				lipgloss.NewStyle().Bold(true).Render("New Todo:"),
				"",
				textInput.View(),
			)
			formBox := formStyle.Render(formDisplay)

			// Todo list - dynamic border color based on mode
			todoBorderColor := "99" // Purple (navigation mode - active)
			if inInputMode {
				todoBorderColor = "240" // Dark grey (input mode - inactive)
			}
			todoStyle := lipgloss.NewStyle().
				Padding(1, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(todoBorderColor)).
				Width(60)

			// Render list
			listView := list.View()
			if strings.TrimSpace(listView) == "" {
				listView = "No todos yet. Press Ctrl+N to create one!"
			}
			todoBox := todoStyle.Render(listView)

			return lipgloss.JoinVertical(
				lipgloss.Left,
				stats,
				"",
				formBox,
				"",
				todoBox,
			)
		}).
		Build()
}

func main() {
	component, err := createTodoApp()
	if err != nil {
		fmt.Printf("Error creating component: %v\n", err)
		os.Exit(1)
	}

	// CRITICAL: Don't call component.Init() manually
	// Bubbletea will call model.Init() which calls component.Init()

	m := model{
		component:    component,
		selectedTodo: 0,
		inputMode:    false, // Start in navigation mode
	}

	// Use tea.WithAltScreen() for full terminal screen mode
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
