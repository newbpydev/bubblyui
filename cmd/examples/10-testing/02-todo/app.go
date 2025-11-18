package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	localcomponents "github.com/newbpydev/bubblyui/cmd/examples/10-testing/02-todo/components"
	"github.com/newbpydev/bubblyui/cmd/examples/10-testing/02-todo/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// CreateApp creates the root TodoApp component
// Demonstrates BubblyUI's composable component architecture
func CreateApp() (bubbly.Component, error) {
	builder := bubbly.NewComponent("TodoApp").
		WithAutoCommands(true).
		WithKeyBinding("enter", "submit", "Add/Submit").
		WithKeyBinding("esc", "toggleMode", "Toggle mode").
		WithKeyBinding(" ", "toggle", "Toggle first").
		WithKeyBinding("d", "remove", "Delete first").
		WithKeyBinding("c", "clear", "Clear all").
		WithKeyBinding("a", "toggleAll", "Toggle all").
		WithKeyBinding("backspace", "removeChar", "Backspace").
		WithKeyBinding("ctrl+c", "quit", "Quit").
		// CRITICAL: Capture keyboard for text input
		// Parent captures keys, updates refs, components re-render automatically
		WithMessageHandler(func(comp bubbly.Component, msg tea.Msg) tea.Cmd {
			if keyMsg, ok := msg.(tea.KeyMsg); ok {
				switch keyMsg.Type {
				case tea.KeyRunes:
					// Emit to addChar event - updates inputValue ref
					comp.Emit("addChar", string(keyMsg.Runes))
				}
			}
			return nil
		}).
		Setup(func(ctx *bubbly.Context) {
			// Initialize todos composable (testable business logic)
			todos := composables.UseTodos(ctx, []composables.Todo{})

			// Input value for new todos - shared with TodoInput component
			inputValue := ctx.Ref("")

			// Mode state: false = Navigation, true = Input
			inputMode := ctx.Ref(false)

			// PROVIDE focus state to descendants (BubblyUI Provide/Inject pattern!)
			ctx.Provide("inputModeFocus", inputMode)
			ctx.Provide("focusColor", lipgloss.Color("35"))     // Green for input focus
			ctx.Provide("inactiveColor", lipgloss.Color("240")) // Grey for inactive

			// Create TodoInput component (need to convert interface{} ref to string ref)
			stringInputValue := bubbly.NewRef("")

			input, err := localcomponents.CreateTodoInput(localcomponents.TodoInputProps{
				Value:   stringInputValue,
				Focused: inputMode, // Pass focus state as prop!
				OnSubmit: func(title string) {
					todos.Add(title)
					inputValue.Set("")
					stringInputValue.Set("")
				},
			})
			if err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to create input: %v", err))
				return
			}

			// Create TodoList component
			list, err := localcomponents.CreateTodoList(localcomponents.TodoListProps{
				Todos:    todos.Todos,
				Focused:  inputMode, // Pass same mode state (inverted logic in component)
				OnToggle: todos.Toggle,
				OnRemove: todos.Remove,
			})
			if err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to create list: %v", err))
				return
			}

			// Expose components for template
			if err := ctx.ExposeComponent("input", input); err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to expose input: %v", err))
				return
			}
			if err := ctx.ExposeComponent("list", list); err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to expose list: %v", err))
				return
			}

			// Expose state for tests
			ctx.Expose("inputValue", inputValue)
			ctx.Expose("inputMode", inputMode)
			ctx.Expose("todos", todos.Todos)
			ctx.Expose("total", todos.Total)
			ctx.Expose("completed", todos.Completed)
			ctx.Expose("remaining", todos.Remaining)
			ctx.Expose("allDone", todos.AllDone)

			// Expose actions for tests
			ctx.Expose("add", todos.Add)
			ctx.Expose("toggle", todos.Toggle)
			ctx.Expose("remove", todos.Remove)
			ctx.Expose("clear", todos.Clear)
			ctx.Expose("toggleAll", todos.ToggleAll)

			// Event: Toggle mode (ESC key)
			ctx.On("toggleMode", func(data interface{}) {
				current := inputMode.Get().(bool)
				inputMode.Set(!current)
			})

			// Event: Submit (Enter key)
			ctx.On("submit", func(data interface{}) {
				value := inputValue.Get().(string)
				if value != "" {
					todos.Add(value)
					inputValue.Set("")
				}
			})

			// Event: Toggle first incomplete todo (Space key) - Navigation mode only
			ctx.On("toggle", func(data interface{}) {
				if inputMode.Get().(bool) {
					return // Blocked in input mode
				}
				current := todos.Todos.Get().([]composables.Todo)
				for _, todo := range current {
					if !todo.Completed {
						todos.Toggle(todo.ID)
						break
					}
				}
			})

			// Event: Remove first todo (d key) - Navigation mode only
			ctx.On("remove", func(data interface{}) {
				if inputMode.Get().(bool) {
					return // Blocked in input mode
				}
				current := todos.Todos.Get().([]composables.Todo)
				if len(current) > 0 {
					todos.Remove(current[0].ID)
				}
			})

			// Event: Clear all todos (c key) - Navigation mode only
			ctx.On("clear", func(data interface{}) {
				if inputMode.Get().(bool) {
					return // Blocked in input mode
				}
				todos.Clear()
			})

			// Event: Toggle all todos (a key) - Navigation mode only
			ctx.On("toggleAll", func(data interface{}) {
				if inputMode.Get().(bool) {
					return // Blocked in input mode
				}
				todos.ToggleAll()
			})

			// Event: Add character to input (typing) - Input mode only
			ctx.On("addChar", func(data interface{}) {
				if !inputMode.Get().(bool) {
					return // Blocked in navigation mode
				}
				char := data.(string)
				current := inputValue.Get().(string)
				newValue := current + char
				inputValue.Set(newValue)
				stringInputValue.Set(newValue) // Sync both refs
			})

			// Event: Remove character from input (backspace) - Input mode only
			ctx.On("removeChar", func(data interface{}) {
				if !inputMode.Get().(bool) {
					return // Blocked in navigation mode
				}
				current := inputValue.Get().(string)
				if len(current) > 0 {
					newValue := current[:len(current)-1]
					inputValue.Set(newValue)
					stringInputValue.Set(newValue) // Sync both refs
				}
			})

			// Event: Quit (ctrl+c) - handled by bubbly.Wrap()
			ctx.On("quit", func(data interface{}) {
				// Quit command returns tea.Quit automatically
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Check for errors
			if errMsg := ctx.Get("error"); errMsg != nil {
				return fmt.Sprintf("Error: %v", errMsg)
			}

			// Get components (BubblyUI composable architecture!)
			input := ctx.Get("input").(bubbly.Component)
			list := ctx.Get("list").(bubbly.Component)

			// Get state
			inputMode := ctx.Get("inputMode").(*bubbly.Ref[interface{}]).Get().(bool)
			total := ctx.Get("total").(*bubbly.Computed[interface{}]).Get().(int)
			completed := ctx.Get("completed").(*bubbly.Computed[interface{}]).Get().(int)
			remaining := ctx.Get("remaining").(*bubbly.Computed[interface{}]).Get().(int)

			// Styling
			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("99")).
				MarginBottom(1)

			statsStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("240"))

			helpStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				MarginTop(1)

			// Title
			title := titleStyle.Render("📝 Todo App Example")

			// Mode indicator badge (critical for UX!)
			modeStyle := lipgloss.NewStyle().
				Bold(true).
				Padding(0, 1).
				MarginBottom(1)

			var modeIndicator string
			if inputMode {
				// INPUT MODE - Green background
				modeIndicator = modeStyle.
					Foreground(lipgloss.Color("15")).
					Background(lipgloss.Color("35")).
					Render("✍️  INPUT MODE - Type to add todo, ESC to navigate")
			} else {
				// NAVIGATION MODE - Purple background
				modeIndicator = modeStyle.
					Foreground(lipgloss.Color("15")).
					Background(lipgloss.Color("99")).
					Render("🧭 NAVIGATION MODE - Press ESC to type")
			}

			// Stats
			stats := statsStyle.Render(fmt.Sprintf(
				"Total: %d  |  Completed: %d  |  Remaining: %d",
				total, completed, remaining,
			))

			// Help text (mode-specific)
			comp := ctx.Component()
			var help string
			if inputMode {
				help = helpStyle.Render("esc: navigation • enter: submit • backspace: delete char • ctrl+c: quit")
			} else {
				help = helpStyle.Render(comp.HelpText())
			}

			// Compose layout using BubblyUI components
			content := lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				"",
				modeIndicator, // MODE INDICATOR - Critical visual feedback!
				"",
				input.View(), // TodoInput component
				"",
				list.View(), // TodoList component
				"",
				stats,
				"",
				help,
			)

			return lipgloss.NewStyle().
				Padding(2).
				Render(content)
		})

	return builder.Build()
}
