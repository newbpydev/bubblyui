package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/cmd/examples/08-automatic-bridge/02-todos/02-todo-components/components"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// createTodoApp creates the main todo application with component composition
func createTodoApp() (bubbly.Component, error) {
	// Shared state refs (will be passed to child components)
	var inputModeRef *bubbly.Ref[interface{}]

	builder := bubbly.NewComponent("TodoApp").
		WithAutoCommands(true).
		// Key bindings
		WithKeyBinding("ctrl+c", "quit", "Quit application").
		WithKeyBinding("ctrl+n", "newTodo", "New todo").
		WithKeyBinding("ctrl+e", "editTodo", "Edit selected").
		WithKeyBinding("ctrl+d", "deleteTodo", "Delete selected").
		WithKeyBinding("up", "selectPrevious", "Previous todo").
		WithKeyBinding("down", "selectNext", "Next todo").
		WithKeyBinding("enter", "handleEnter", "Add/Submit").
		WithKeyBinding("esc", "toggleMode", "Toggle mode").
		WithKeyBinding("tab", "nextField", "Next field").
		WithKeyBinding("backspace", "removeChar", "Delete character").
		// Conditional key bindings for space
		WithConditionalKeyBinding(bubbly.KeyBinding{
			Key:         " ",
			Event:       "toggleTodo",
			Description: "Toggle completion",
			Condition: func() bool {
				if inputModeRef == nil {
					return false
				}
				return !inputModeRef.Get().(bool)
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
				return inputModeRef.Get().(bool)
			},
		}).
		// Message handler for character input
		WithMessageHandler(func(comp bubbly.Component, msg tea.Msg) tea.Cmd {
			if keyMsg, ok := msg.(tea.KeyMsg); ok {
				switch keyMsg.Type {
				case tea.KeyRunes:
					comp.Emit("addChar", string(keyMsg.Runes))
				}
			}
			return nil
		})

	builder = builder.Setup(func(ctx *bubbly.Context) {
		// State: Todo list (using ctx.Ref for consistency)
		todos := ctx.Ref([]components.Todo{})
		nextID := ctx.Ref(1)
		selectedIndex := ctx.Ref(0)

		// State: Input mode and form (using ctx.Ref for consistency)
		inputMode := ctx.Ref(false)
		editMode := ctx.Ref(false)
		focusedField := ctx.Ref("Title")

		// Assign to outer variable for conditional key bindings
		inputModeRef = inputMode

		// Create child components with props
		todoForm, err := components.CreateTodoForm(components.TodoFormProps{
			InputMode:    inputMode,
			FocusedField: focusedField,
			OnSubmit: func(data components.TodoFormData) {
				if editMode.Get().(bool) {
					// Update existing todo
					todoList := todos.Get().([]components.Todo)
					index := selectedIndex.Get().(int)
					if index >= 0 && index < len(todoList) {
						todoList[index].Title = data.Title
						todoList[index].Description = data.Description
						todoList[index].Priority = data.Priority
						todos.Set(todoList)
					}
					editMode.Set(false)
				} else {
					// Add new todo
					newTodo := components.Todo{
						ID:          nextID.Get().(int),
						Title:       data.Title,
						Description: data.Description,
						Priority:    data.Priority,
						Completed:   false,
					}
					todoList := todos.Get().([]components.Todo)
					todos.Set(append(todoList, newTodo))
					nextID.Set(nextID.Get().(int) + 1)
				}
				inputMode.Set(false)
			},
			OnCancel: func() {
				inputMode.Set(false)
				editMode.Set(false)
			},
		})
		if err != nil {
			ctx.Expose("error", err)
			return
		}

		todoList, err := components.CreateTodoList(components.TodoListProps{
			Todos:         todos,
			SelectedIndex: selectedIndex,
			InputMode:     inputMode,
			OnToggle: func(index int) {
				todoList := todos.Get().([]components.Todo)
				if index >= 0 && index < len(todoList) {
					todoList[index].Completed = !todoList[index].Completed
					todos.Set(todoList)
				}
			},
			OnSelect: func(index int) {
				selectedIndex.Set(index)
			},
		})
		if err != nil {
			ctx.Expose("error", err)
			return
		}

		todoStats, err := components.CreateTodoStats(components.TodoStatsProps{
			Todos: todos,
		})
		if err != nil {
			ctx.Expose("error", err)
			return
		}

		// Auto-initialize and expose child components
		// ExposeComponent automatically calls Init() if not already initialized
		ctx.ExposeComponent("todoForm", todoForm)
		ctx.ExposeComponent("todoList", todoList)
		ctx.ExposeComponent("todoStats", todoStats)
		ctx.Expose("inputMode", inputMode)
		ctx.Expose("editMode", editMode)

		// Event: Quit
		ctx.On("quit", func(_ interface{}) {
			// Handled by wrapper
		})

		// Event: New todo
		ctx.On("newTodo", func(_ interface{}) {
			if !inputMode.Get().(bool) {
				editMode.Set(false)
				inputMode.Set(true)
				inputModeRef.Set(true)
				focusedField.Set("Title")
				todoForm.Emit("clearForm", nil)
			}
		})

		// Event: Edit todo
		ctx.On("editTodo", func(_ interface{}) {
			if !inputMode.Get().(bool) && !editMode.Get().(bool) {
				todoList := todos.Get().([]components.Todo)
				index := selectedIndex.Get().(int)
				if len(todoList) > 0 && index >= 0 && index < len(todoList) {
					todo := todoList[index]
					editMode.Set(true)
					inputMode.Set(true)
					inputModeRef.Set(true)
					focusedField.Set("Title")
					todoForm.Emit("setFormData", components.TodoFormData{
						Title:       todo.Title,
						Description: todo.Description,
						Priority:    todo.Priority,
					})
				}
			}
		})

		// Event: Delete todo
		ctx.On("deleteTodo", func(_ interface{}) {
			if !inputMode.Get().(bool) && !editMode.Get().(bool) {
				todoList := todos.Get().([]components.Todo)
				index := selectedIndex.Get().(int)
				if len(todoList) > 0 && index >= 0 && index < len(todoList) {
					todos.Set(append(todoList[:index], todoList[index+1:]...))
					// Adjust selection
					if index >= len(todoList)-1 && len(todoList) > 1 {
						selectedIndex.Set(len(todoList) - 2)
					}
				}
			}
		})

		// Event: Select previous
		ctx.On("selectPrevious", func(_ interface{}) {
			if !inputMode.Get().(bool) && !editMode.Get().(bool) {
				todoList.Emit("selectPrevious", nil)
			}
		})

		// Event: Select next
		ctx.On("selectNext", func(_ interface{}) {
			if !inputMode.Get().(bool) && !editMode.Get().(bool) {
				todoList.Emit("selectNext", nil)
			}
		})

		// Event: Handle enter key
		ctx.On("handleEnter", func(_ interface{}) {
			if inputMode.Get().(bool) {
				// Submit form
				todoForm.Emit("submitForm", nil)
			} else {
				// Enter input mode to add new todo
				editMode.Set(false)
				inputMode.Set(true)
				inputModeRef.Set(true)
				focusedField.Set("Title")
				todoForm.Emit("clearForm", nil)
			}
		})

		// Event: Toggle mode
		ctx.On("toggleMode", func(_ interface{}) {
			newMode := !inputMode.Get().(bool)
			inputMode.Set(newMode)
			inputModeRef.Set(newMode)
			if !newMode {
				editMode.Set(false)
				todoForm.Emit("clearForm", nil)
			}
		})

		// Event: Next field
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

		// Event: Remove character
		ctx.On("removeChar", func(_ interface{}) {
			if inputMode.Get().(bool) {
				todoForm.Emit("removeChar", nil)
			}
		})

		// Event: Add character
		ctx.On("addChar", func(data interface{}) {
			if inputMode.Get().(bool) {
				todoForm.Emit("addChar", data)
			}
		})

		// Event: Toggle todo
		ctx.On("toggleTodo", func(_ interface{}) {
			if !inputMode.Get().(bool) && !editMode.Get().(bool) {
				index := selectedIndex.Get().(int)
				todoList.Emit("toggleTodo", index)
			}
		})
	})

	builder = builder.Template(func(ctx bubbly.RenderContext) string {
		// Get child components
		todoForm := ctx.Get("todoForm").(bubbly.Component)
		todoList := ctx.Get("todoList").(bubbly.Component)
		todoStats := ctx.Get("todoStats").(bubbly.Component)

		// Get state for mode indicator
		inputMode := ctx.Get("inputMode").(*bubbly.Ref[interface{}]).Get().(bool)
		editMode := ctx.Get("editMode").(*bubbly.Ref[interface{}]).Get().(bool)

		// Title
		titleStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)
		title := titleStyle.Render("📝 Todo App - Component-Based (Vue-like)")

		// Render child components
		stats := todoStats.View()
		form := todoForm.View()
		list := todoList.View()

		// Mode indicator and help
		modeIndicator := components.RenderModeIndicator(inputMode, editMode)
		help := components.RenderFormHelp(inputMode)

		return lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			"",
			stats,
			"",
			form,
			"",
			list,
			"",
			modeIndicator,
			help,
		)
	})

	return builder.Build()
}

func main() {
	component, err := createTodoApp()
	if err != nil {
		fmt.Printf("Error creating component: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(bubbly.Wrap(component), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
