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
	ID          int
	Title       string
	Description string
	Priority    string // "low", "medium", "high"
	Completed   bool
}

// createTodoApp creates the todo application component with declarative key bindings
func createTodoApp() (bubbly.Component, error) {
	// Shared state for conditional key bindings
	// We need to create these refs outside the Setup so they can be used in Condition functions
	var inputModeRef *bubbly.Ref[interface{}]

	// Create component with automatic bridge and declarative key bindings
	builder := bubbly.NewComponent("TodoApp").
		WithAutoCommands(true). // Enable automatic reactive bridge
		// Standard key bindings (always active)
		WithKeyBinding("ctrl+c", "quit", "Quit application").
		WithKeyBinding("ctrl+n", "newTodo", "New todo").
		WithKeyBinding("ctrl+e", "editTodo", "Edit selected").
		WithKeyBinding("ctrl+d", "deleteTodo", "Delete selected").
		WithKeyBinding("up", "selectPrevious", "Previous todo").
		WithKeyBinding("down", "selectNext", "Next todo").
		WithKeyBinding("enter", "handleEnter", "Add/Submit").
		WithKeyBinding("esc", "toggleMode", "Toggle mode").
		WithKeyBinding("tab", "nextField", "Next field").
		WithKeyBinding("backspace", "removeChar", "Delete character")

	// Add conditional key bindings for space key (mode-based)
	// CRITICAL: Use " " (space character), not "space" string
	builder = builder.
		WithConditionalKeyBinding(bubbly.KeyBinding{
			Key:         " ",
			Event:       "toggleTodo",
			Description: "Toggle completion",
			Condition: func() bool {
				if inputModeRef == nil {
					return false
				}
				return !inputModeRef.Get().(bool) // Only in navigation mode
			},
		}).
		WithConditionalKeyBinding(bubbly.KeyBinding{
			Key:         " ",
			Event:       "addChar",
			Data:        " ",
			Description: "Add space",
			Condition: func() bool {
				if inputModeRef == nil {
					return false
				}
				return inputModeRef.Get().(bool) // Only in input mode
			},
		}).
		// Message handler for character input (escape hatch for text typing)
		// Declarative key bindings handle specific keys, but we need this for any character
		WithMessageHandler(func(comp bubbly.Component, msg tea.Msg) tea.Cmd {
			// Capture all character input and emit to addChar event
			// The event handler will check if we're in input mode
			if keyMsg, ok := msg.(tea.KeyMsg); ok {
				switch keyMsg.Type {
				case tea.KeyRunes:
					// Regular character input (a-z, A-Z, 0-9, punctuation, etc.)
					comp.Emit("addChar", string(keyMsg.Runes))
				}
			}
			return nil
		})

	// Now add Setup
	builder = builder.Setup(func(ctx *bubbly.Context) {
		// State: Todo list
		todos := ctx.Ref([]Todo{})
		nextID := ctx.Ref(1)
		selectedIndex := ctx.Ref(0)

		// State: Input mode and form
		inputMode := ctx.Ref(false)
		inputModeRef = inputMode // Assign to outer variable for conditional key bindings
		editMode := ctx.Ref(false)
		focusedField := ctx.Ref("Title") // "Title", "Description", or "Priority"

		// Form fields
		titleInput := ctx.Ref("")
		descInput := ctx.Ref("")
		priorityInput := ctx.Ref("medium")

		// Expose state for conditional key bindings and template
		ctx.Expose("todos", todos)
		ctx.Expose("selectedIndex", selectedIndex)
		ctx.Expose("inputMode", inputMode)
		ctx.Expose("editMode", editMode)
		ctx.Expose("focusedField", focusedField)
		ctx.Expose("titleInput", titleInput)
		ctx.Expose("descInput", descInput)
		ctx.Expose("priorityInput", priorityInput)

		// Statistics (computed values)
		totalCount := ctx.Computed(func() interface{} {
			return len(todos.Get().([]Todo))
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
			return totalCount.Get().(int) - completedCount.Get().(int)
		})
		ctx.Expose("totalCount", totalCount)
		ctx.Expose("completedCount", completedCount)
		ctx.Expose("pendingCount", pendingCount)

		// Event: Quit application
		ctx.On("quit", func(_ interface{}) {
			// Handled by wrapper - returns tea.Quit
		})

		// Event: Toggle mode (navigation <-> input)
		ctx.On("toggleMode", func(_ interface{}) {
			currentMode := inputMode.Get().(bool)
			inputMode.Set(!currentMode)

			// Exit edit mode when leaving input mode
			if currentMode { // Was in input mode, now going to navigation
				editMode.Set(false)
				// Clear form
				titleInput.Set("")
				descInput.Set("")
				priorityInput.Set("medium")
			}
		})

		// Event: New todo
		ctx.On("newTodo", func(_ interface{}) {
			// Only in navigation mode
			if !inputMode.Get().(bool) {
				editMode.Set(false)
				inputMode.Set(true)
				// Clear form
				titleInput.Set("")
				descInput.Set("")
				priorityInput.Set("medium")
				focusedField.Set("Title")
			}
		})

		// Event: Edit selected todo
		ctx.On("editTodo", func(_ interface{}) {
			// Only in navigation mode
			if !inputMode.Get().(bool) {
				todoList := todos.Get().([]Todo)
				selected := selectedIndex.Get().(int)

				if len(todoList) > 0 && selected >= 0 && selected < len(todoList) {
					todo := todoList[selected]
					editMode.Set(true)
					inputMode.Set(true)
					titleInput.Set(todo.Title)
					descInput.Set(todo.Description)
					priorityInput.Set(todo.Priority)
					focusedField.Set("Title")
				}
			}
		})

		// Event: Delete selected todo
		ctx.On("deleteTodo", func(_ interface{}) {
			// Only in navigation mode
			if !inputMode.Get().(bool) {
				todoList := todos.Get().([]Todo)
				selected := selectedIndex.Get().(int)

				if len(todoList) > 0 && selected >= 0 && selected < len(todoList) {
					newList := append(todoList[:selected], todoList[selected+1:]...)
					todos.Set(newList)

					// Adjust selection
					if selected >= len(newList) && len(newList) > 0 {
						selectedIndex.Set(len(newList) - 1)
					}
				}
			}
		})

		// Event: Select previous todo
		ctx.On("selectPrevious", func(_ interface{}) {
			// Only in navigation mode
			if !inputMode.Get().(bool) {
				current := selectedIndex.Get().(int)
				if current > 0 {
					selectedIndex.Set(current - 1)
				}
			}
		})

		// Event: Select next todo
		ctx.On("selectNext", func(_ interface{}) {
			// Only in navigation mode
			if !inputMode.Get().(bool) {
				todoList := todos.Get().([]Todo)
				current := selectedIndex.Get().(int)
				if current < len(todoList)-1 {
					selectedIndex.Set(current + 1)
				}
			}
		})

		// Event: Toggle todo completion
		ctx.On("toggleTodo", func(_ interface{}) {
			// Only in navigation mode
			if !inputMode.Get().(bool) {
				todoList := todos.Get().([]Todo)
				selected := selectedIndex.Get().(int)

				if len(todoList) > 0 && selected >= 0 && selected < len(todoList) {
					todoList[selected].Completed = !todoList[selected].Completed
					todos.Set(todoList)
				}
			}
		})

		// Event: Handle Enter key (context-dependent)
		ctx.On("handleEnter", func(_ interface{}) {
			if inputMode.Get().(bool) {
				// In input mode: submit form
				title := titleInput.Get().(string)

				// Validation
				if len(title) < 3 {
					return // Don't submit if title too short
				}

				if editMode.Get().(bool) {
					// Update existing todo
					todoList := todos.Get().([]Todo)
					selected := selectedIndex.Get().(int)

					if selected >= 0 && selected < len(todoList) {
						todoList[selected].Title = title
						todoList[selected].Description = descInput.Get().(string)
						todoList[selected].Priority = priorityInput.Get().(string)
						todos.Set(todoList)
					}

					// Exit edit mode
					editMode.Set(false)
					inputMode.Set(false)
				} else {
					// Add new todo
					todoList := todos.Get().([]Todo)
					id := nextID.Get().(int)

					newTodo := Todo{
						ID:          id,
						Title:       title,
						Description: descInput.Get().(string),
						Priority:    priorityInput.Get().(string),
						Completed:   false,
					}

					todos.Set(append(todoList, newTodo))
					nextID.Set(id + 1)
				}

				// Clear form
				titleInput.Set("")
				descInput.Set("")
				priorityInput.Set("medium")
			} else {
				// In navigation mode: enter input mode to add new todo
				inputMode.Set(true)
				editMode.Set(false)
				focusedField.Set("Title")
			}
		})

		// Event: Next field (Tab in input mode)
		ctx.On("nextField", func(_ interface{}) {
			if inputMode.Get().(bool) {
				current := focusedField.Get().(string)
				switch current {
				case "Title":
					focusedField.Set("Description")
				case "Description":
					focusedField.Set("Priority")
				case "Priority":
					focusedField.Set("Title")
				}
			}
		})

		// Event: Add character to current field
		ctx.On("addChar", func(data interface{}) {
			if inputMode.Get().(bool) {
				char := data.(string)
				field := focusedField.Get().(string)

				switch field {
				case "Title":
					current := titleInput.Get().(string)
					titleInput.Set(current + char)
				case "Description":
					current := descInput.Get().(string)
					descInput.Set(current + char)
				case "Priority":
					current := priorityInput.Get().(string)
					priorityInput.Set(current + char)
				}
			}
		})

		// Event: Remove character from current field
		ctx.On("removeChar", func(_ interface{}) {
			if inputMode.Get().(bool) {
				field := focusedField.Get().(string)

				switch field {
				case "Title":
					current := titleInput.Get().(string)
					if len(current) > 0 {
						titleInput.Set(current[:len(current)-1])
					}
				case "Description":
					current := descInput.Get().(string)
					if len(current) > 0 {
						descInput.Set(current[:len(current)-1])
					}
				case "Priority":
					current := priorityInput.Get().(string)
					if len(current) > 0 {
						priorityInput.Set(current[:len(current)-1])
					}
				}
			}
		})
	}).
		Template(func(ctx bubbly.RenderContext) string {
			// Get state
			todos := ctx.Get("todos").(*bubbly.Ref[interface{}])
			selectedIndex := ctx.Get("selectedIndex").(*bubbly.Ref[interface{}])
			inputMode := ctx.Get("inputMode").(*bubbly.Ref[interface{}])
			editMode := ctx.Get("editMode").(*bubbly.Ref[interface{}])
			focusedField := ctx.Get("focusedField").(*bubbly.Ref[interface{}])
			titleInput := ctx.Get("titleInput").(*bubbly.Ref[interface{}])
			descInput := ctx.Get("descInput").(*bubbly.Ref[interface{}])
			priorityInput := ctx.Get("priorityInput").(*bubbly.Ref[interface{}])
			totalCount := ctx.Get("totalCount").(*bubbly.Computed[interface{}])
			completedCount := ctx.Get("completedCount").(*bubbly.Computed[interface{}])
			pendingCount := ctx.Get("pendingCount").(*bubbly.Computed[interface{}])

			// Extract values
			todoList := todos.Get().([]Todo)
			selected := selectedIndex.Get().(int)
			inInputMode := inputMode.Get().(bool)
			inEditMode := editMode.Get().(bool)
			focused := focusedField.Get().(string)
			title := titleInput.Get().(string)
			desc := descInput.Get().(string)
			priority := priorityInput.Get().(string)

			// Title
			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205")).
				MarginBottom(1)
			titleText := titleStyle.Render("📝 Todo App - Declarative Key Bindings")

			// Statistics
			statsStyle := lipgloss.NewStyle().
				Bold(true).
				Padding(0, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("99")).
				Width(70)
			stats := statsStyle.Render(fmt.Sprintf(
				"📊 Total: %d | ✅ Completed: %d | ⏳ Pending: %d",
				totalCount.Get().(int),
				completedCount.Get().(int),
				pendingCount.Get().(int),
			))

			// Form box - dynamic border color based on mode
			formBorderColor := "240" // Dark grey (navigation mode - inactive)
			if inInputMode {
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
			if focused == "Title" {
				titleLabel = "▶ " + titleLabel
			} else {
				titleLabel = "  " + titleLabel
			}
			titleValue := title
			if titleValue == "" {
				titleValue = "(empty)"
			}
			titleError := ""
			if len(title) > 0 && len(title) < 3 {
				titleError = " ❌ Must be at least 3 characters"
			}
			formFields = append(formFields, titleLabel+" "+titleValue+titleError)

			// Description field
			descLabel := "Description:"
			if focused == "Description" {
				descLabel = "▶ " + descLabel
			} else {
				descLabel = "  " + descLabel
			}
			descValue := desc
			if descValue == "" {
				descValue = "(empty)"
			}
			formFields = append(formFields, descLabel+" "+descValue)

			// Priority field
			priorityLabel := "Priority:"
			if focused == "Priority" {
				priorityLabel = "▶ " + priorityLabel
			} else {
				priorityLabel = "  " + priorityLabel
			}
			priorityValue := priority
			if priorityValue == "" {
				priorityValue = "(empty)"
			}
			formFields = append(formFields, priorityLabel+" "+priorityValue)

			// Form status
			formStatus := ""
			if len(title) >= 3 {
				formStatus = "\n✓ Ready to submit"
			} else if len(title) > 0 {
				formStatus = "\n❌ Title too short"
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
				Width(70)

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
			if inInputMode {
				modeStyle = modeStyle.
					Foreground(lipgloss.Color("15")).
					Background(lipgloss.Color("35"))
				if inEditMode {
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

			// Help text (auto-generated from key bindings)
			comp := ctx.Component()
			helpStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				MarginTop(1)

			var help string
			if inInputMode {
				help = helpStyle.Render("tab: next field • enter: save • esc: cancel • ctrl+c: quit")
			} else {
				help = helpStyle.Render(comp.HelpText())
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
		})

	// Build the component
	component, err := builder.Build()
	if err != nil {
		return nil, err
	}

	return component, nil
}

func main() {
	component, err := createTodoApp()
	if err != nil {
		fmt.Printf("Error creating component: %v\n", err)
		os.Exit(1)
	}

	// Zero-boilerplate integration with bubbly.Wrap()!
	// No manual wrapper model needed - automatic bridge handles everything
	p := tea.NewProgram(bubbly.Wrap(component), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
