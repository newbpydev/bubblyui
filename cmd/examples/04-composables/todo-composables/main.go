package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
)

// Todo represents a single todo item
type Todo struct {
	ID          int
	Title       string
	Description string
	Priority    string // "low", "medium", "high"
	Completed   bool
}

// TodoInput represents the form input for creating/editing todos
type TodoInput struct {
	Title       string
	Description string
	Priority    string
}

// model wraps the todo component
type model struct {
	component    bubbly.Component
	selectedTodo int  // Index of selected todo in list
	editMode     bool // Whether we're editing an existing todo
	inputMode    bool // Whether we're in input mode (typing in form fields)
}

func (m model) Init() tea.Cmd {
	// CRITICAL: Let Bubbletea call Init, don't call it manually
	return m.component.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle space key first (using msg.Type)
		if msg.Type == tea.KeySpace {
			if !m.inputMode && !m.editMode {
				// Navigation mode: toggle completion
				m.component.Emit("toggleTodo", m.selectedTodo)
			} else if m.inputMode {
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
				if !m.inputMode && m.editMode {
					// Exit edit mode when leaving input mode
					m.editMode = false
					m.component.Emit("clearForm", nil)
				}
			case "ctrl+n":
				// New todo - clear form and enter input mode
				m.editMode = false
				m.inputMode = true
				m.component.Emit("setInputMode", m.inputMode)
				m.component.Emit("clearForm", nil)
			case "ctrl+e":
				// Edit selected todo - only in navigation mode
				if !m.inputMode && !m.editMode {
					m.editMode = true
					m.inputMode = true
					m.component.Emit("setInputMode", m.inputMode)
					m.component.Emit("editTodo", m.selectedTodo)
				}
			case "ctrl+d":
				// Delete selected todo - only in navigation mode
				if !m.inputMode && !m.editMode {
					m.component.Emit("deleteTodo", m.selectedTodo)
				}
			case "enter":
				if m.inputMode {
					// In input mode: submit form
					if m.editMode {
						m.component.Emit("updateTodo", m.selectedTodo)
						m.editMode = false
						m.inputMode = false
						m.component.Emit("setInputMode", m.inputMode)
					} else {
						m.component.Emit("addTodo", nil)
						// Stay in input mode for adding multiple todos
					}
				} else {
					// In navigation mode: enter input mode to add new todo
					m.inputMode = true
					m.component.Emit("setInputMode", m.inputMode)
				}
			case "up":
				// Move selection up - only in navigation mode
				if !m.inputMode && !m.editMode {
					m.component.Emit("selectPrevious", nil)
					if m.selectedTodo > 0 {
						m.selectedTodo--
					}
				}
			case "down":
				// Move selection down - only in navigation mode
				if !m.inputMode && !m.editMode {
					m.component.Emit("selectNext", nil)
					m.selectedTodo++
				}
			case "k":
				// In navigation mode: move up; in input mode: type 'k'
				if !m.inputMode && !m.editMode {
					m.component.Emit("selectPrevious", nil)
					if m.selectedTodo > 0 {
						m.selectedTodo--
					}
				} else if m.inputMode {
					m.component.Emit("addChar", "k")
				}
			case "j":
				// In navigation mode: move down; in input mode: type 'j'
				if !m.inputMode && !m.editMode {
					m.component.Emit("selectNext", nil)
					m.selectedTodo++
				} else if m.inputMode {
					m.component.Emit("addChar", "j")
				}
			case "tab":
				// Cycle through form fields - only in input mode
				if m.inputMode {
					m.component.Emit("nextField", nil)
				}
			case "backspace":
				// Remove character - only in input mode
				if m.inputMode {
					m.component.Emit("removeChar", nil)
				}
			default:
				// Handle text input - only in input mode
				if m.inputMode {
					switch msg.Type {
					case tea.KeyRunes:
						m.component.Emit("addChar", string(msg.Runes))
					}
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

	title := titleStyle.Render("📝 Todo App - UseForm Composable")

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginBottom(1)

	subtitle := subtitleStyle.Render(
		"Demonstrates: UseForm for CRUD operations with validation",
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

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	var help string
	if m.inputMode {
		help = helpStyle.Render(
			"tab: next field • enter: save • esc: cancel • ctrl+c: quit",
		)
	} else {
		help = helpStyle.Render(
			"↑/↓: select • space: toggle • ctrl+e: edit • ctrl+d: delete • ctrl+n: new • enter: add • ctrl+c: quit",
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

	// Description validation (optional but has constraints)
	if len(input.Description) > 200 {
		errors["Description"] = "Description must be less than 200 characters"
	}

	// Priority validation
	validPriorities := map[string]bool{"low": true, "medium": true, "high": true}
	if input.Priority != "" && !validPriorities[input.Priority] {
		errors["Priority"] = "Priority must be: low, medium, or high"
	}

	return errors
}

// createTodoApp creates the todo application component
func createTodoApp() (bubbly.Component, error) {
	return bubbly.NewComponent("TodoApp").
		Setup(func(ctx *bubbly.Context) {
			// Use UseForm for todo input
			form := composables.UseForm(ctx, TodoInput{
				Priority: "medium", // Default priority
			}, validateTodoInput)

			// Todo list state
			todos := ctx.Ref([]Todo{})
			nextID := ctx.Ref(1)

			// UI state
			selectedIndex := ctx.Ref(0)
			focusedField := ctx.Ref("Title") // Title, Description, or Priority
			inputMode := ctx.Ref(false)      // Track input mode for visual feedback

			// Statistics
			totalCount := ctx.Computed(func() interface{} {
				todoList := todos.GetTyped().([]Todo)
				return len(todoList)
			})

			completedCount := ctx.Computed(func() interface{} {
				todoList := todos.GetTyped().([]Todo)
				count := 0
				for _, todo := range todoList {
					if todo.Completed {
						count++
					}
				}
				return count
			})

			pendingCount := ctx.Computed(func() interface{} {
				total := totalCount.GetTyped().(int)
				completed := completedCount.GetTyped().(int)
				return total - completed
			})

			// Expose state to template
			ctx.Expose("form", form)
			ctx.Expose("todos", todos)
			ctx.Expose("selectedIndex", selectedIndex)
			ctx.Expose("focusedField", focusedField)
			ctx.Expose("inputMode", inputMode)
			ctx.Expose("totalCount", totalCount)
			ctx.Expose("completedCount", completedCount)
			ctx.Expose("pendingCount", pendingCount)

			// Event: Set input mode (called from model)
			ctx.On("setInputMode", func(data interface{}) {
				mode := data.(bool)
				inputMode.Set(mode)
			})

			// Event: Add new todo
			ctx.On("addTodo", func(_ interface{}) {
				form.Submit()
				if form.IsValid.GetTyped() {
					input := form.Values.GetTyped()
					todoList := todos.GetTyped().([]Todo)
					id := nextID.GetTyped().(int)

					newTodo := Todo{
						ID:          id,
						Title:       input.Title,
						Description: input.Description,
						Priority:    input.Priority,
						Completed:   false,
					}

					todos.Set(append(todoList, newTodo))
					nextID.Set(id + 1)
					form.Reset()
					form.Values.Set(TodoInput{Priority: "medium"})
				}
			})

			// Event: Edit todo
			ctx.On("editTodo", func(data interface{}) {
				index := data.(int)
				todoList := todos.GetTyped().([]Todo)
				if index >= 0 && index < len(todoList) {
					todo := todoList[index]
					form.Values.Set(TodoInput{
						Title:       todo.Title,
						Description: todo.Description,
						Priority:    todo.Priority,
					})
				}
			})

			// Event: Update todo
			ctx.On("updateTodo", func(data interface{}) {
				index := data.(int)
				form.Submit()
				if form.IsValid.GetTyped() {
					input := form.Values.GetTyped()
					todoList := todos.GetTyped().([]Todo)
					if index >= 0 && index < len(todoList) {
						todoList[index].Title = input.Title
						todoList[index].Description = input.Description
						todoList[index].Priority = input.Priority
						todos.Set(todoList)
						form.Reset()
						form.Values.Set(TodoInput{Priority: "medium"})
					}
				}
			})

			// Event: Delete todo
			ctx.On("deleteTodo", func(data interface{}) {
				index := data.(int)
				todoList := todos.GetTyped().([]Todo)
				if index >= 0 && index < len(todoList) {
					newList := append(todoList[:index], todoList[index+1:]...)
					todos.Set(newList)
					// Adjust selection if needed
					if selectedIndex.GetTyped().(int) >= len(newList) && len(newList) > 0 {
						selectedIndex.Set(len(newList) - 1)
					}
				}
			})

			// Event: Toggle todo completion
			ctx.On("toggleTodo", func(data interface{}) {
				index := data.(int)
				todoList := todos.GetTyped().([]Todo)
				if index >= 0 && index < len(todoList) {
					todoList[index].Completed = !todoList[index].Completed
					todos.Set(todoList)
				}
			})

			// Event: Select previous todo
			ctx.On("selectPrevious", func(_ interface{}) {
				current := selectedIndex.GetTyped().(int)
				if current > 0 {
					selectedIndex.Set(current - 1)
				}
			})

			// Event: Select next todo
			ctx.On("selectNext", func(_ interface{}) {
				current := selectedIndex.GetTyped().(int)
				todoList := todos.GetTyped().([]Todo)
				if current < len(todoList)-1 {
					selectedIndex.Set(current + 1)
				}
			})

			// Event: Clear form
			ctx.On("clearForm", func(_ interface{}) {
				form.Reset()
				form.Values.Set(TodoInput{Priority: "medium"})
			})

			// Event: Next field in form
			ctx.On("nextField", func(_ interface{}) {
				current := focusedField.GetTyped().(string)
				switch current {
				case "Title":
					focusedField.Set("Description")
				case "Description":
					focusedField.Set("Priority")
				case "Priority":
					focusedField.Set("Title")
				}
			})

			// Event: Add character to current field
			ctx.On("addChar", func(data interface{}) {
				char := data.(string)
				field := focusedField.GetTyped().(string)
				input := form.Values.GetTyped()

				var newValue string
				switch field {
				case "Title":
					newValue = input.Title + char
					form.SetField("Title", newValue)
				case "Description":
					newValue = input.Description + char
					form.SetField("Description", newValue)
				case "Priority":
					newValue = input.Priority + char
					form.SetField("Priority", newValue)
				}
			})

			// Event: Remove character from current field
			ctx.On("removeChar", func(_ interface{}) {
				field := focusedField.GetTyped().(string)
				input := form.Values.GetTyped()

				var newValue string
				switch field {
				case "Title":
					if len(input.Title) > 0 {
						newValue = input.Title[:len(input.Title)-1]
						form.SetField("Title", newValue)
					}
				case "Description":
					if len(input.Description) > 0 {
						newValue = input.Description[:len(input.Description)-1]
						form.SetField("Description", newValue)
					}
				case "Priority":
					if len(input.Priority) > 0 {
						newValue = input.Priority[:len(input.Priority)-1]
						form.SetField("Priority", newValue)
					}
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Get state
			form := ctx.Get("form").(composables.UseFormReturn[TodoInput])
			todos := ctx.Get("todos").(*bubbly.Ref[interface{}])
			selectedIndex := ctx.Get("selectedIndex").(*bubbly.Ref[interface{}])
			focusedField := ctx.Get("focusedField").(*bubbly.Ref[interface{}])
			inputModeRef := ctx.Get("inputMode").(*bubbly.Ref[interface{}])
			totalCount := ctx.Get("totalCount").(*bubbly.Computed[interface{}])
			completedCount := ctx.Get("completedCount").(*bubbly.Computed[interface{}])
			pendingCount := ctx.Get("pendingCount").(*bubbly.Computed[interface{}])

			input := form.Values.GetTyped()
			errors := form.Errors.GetTyped()
			isValid := form.IsValid.GetTyped()
			todoList := todos.GetTyped().([]Todo)
			selected := selectedIndex.GetTyped().(int)
			focused := focusedField.GetTyped().(string)
			inInputMode := inputModeRef.GetTyped().(bool)

			// Statistics box
			statsStyle := lipgloss.NewStyle().
				Bold(true).
				Padding(0, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("99")).
				Width(60)

			stats := statsStyle.Render(fmt.Sprintf(
				"📊 Total: %d | ✅ Completed: %d | ⏳ Pending: %d",
				totalCount.GetTyped().(int),
				completedCount.GetTyped().(int),
				pendingCount.GetTyped().(int),
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

			// Build form fields
			var formFields []string

			// Title field
			titleLabel := "Title:"
			if focused == "Title" {
				titleLabel = "▶ " + titleLabel
			} else {
				titleLabel = "  " + titleLabel
			}
			titleValue := input.Title
			if titleValue == "" {
				titleValue = "(empty)"
			}
			titleError := ""
			if err, ok := errors["Title"]; ok {
				titleError = " ❌ " + err
			}
			formFields = append(formFields, titleLabel+" "+titleValue+titleError)

			// Description field
			descLabel := "Description:"
			if focused == "Description" {
				descLabel = "▶ " + descLabel
			} else {
				descLabel = "  " + descLabel
			}
			descValue := input.Description
			if descValue == "" {
				descValue = "(empty)"
			}
			descError := ""
			if err, ok := errors["Description"]; ok {
				descError = " ❌ " + err
			}
			formFields = append(formFields, descLabel+" "+descValue+descError)

			// Priority field
			priorityLabel := "Priority:"
			if focused == "Priority" {
				priorityLabel = "▶ " + priorityLabel
			} else {
				priorityLabel = "  " + priorityLabel
			}
			priorityValue := input.Priority
			if priorityValue == "" {
				priorityValue = "(empty)"
			}
			priorityError := ""
			if err, ok := errors["Priority"]; ok {
				priorityError = " ❌ " + err
			}
			formFields = append(formFields, priorityLabel+" "+priorityValue+priorityError)

			// Form status
			formStatus := ""
			if !isValid && len(errors) > 0 {
				formStatus = fmt.Sprintf("\n❌ Fix %d error%s", len(errors),
					map[bool]string{true: "", false: "s"}[len(errors) == 1])
			} else if isValid {
				formStatus = "\n✓ Ready to submit"
			}

			formBox := formStyle.Render(strings.Join(formFields, "\n") + formStatus)

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

			var todoItems []string
			if len(todoList) == 0 {
				todoItems = append(todoItems, "No todos yet. Press Ctrl+N to create one!")
			} else {
				for i, todo := range todoList {
					cursor := "  "
					if i == selected {
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

					title := todo.Title
					if todo.Completed {
						title = lipgloss.NewStyle().
							Foreground(lipgloss.Color("240")).
							Strikethrough(true).
							Render(title)
					}

					todoItems = append(todoItems, fmt.Sprintf(
						"%s%s %s %s",
						cursor, checkbox, priorityIcon, title,
					))
				}
			}

			todoBox := todoStyle.Render(strings.Join(todoItems, "\n"))

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
		editMode:     false,
		inputMode:    false, // Start in navigation mode
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
