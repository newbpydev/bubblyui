package main

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/bubblyui/cmd/examples/09-devtools/02-component-inspection/components"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// CreateApp creates the root application component
// This demonstrates:
// - Multi-level component hierarchy (4 levels)
// - Component composition pattern
// - State management across components
// - Exposing complex state for dev tools
func CreateApp() (bubbly.Component, error) {
	builder := bubbly.NewComponent("TodoApp").
		// Key bindings for todo operations
		WithKeyBinding("up", "selectPrevious", "Select previous").
		WithKeyBinding("down", "selectNext", "Select next").
		WithKeyBinding(" ", "toggleTodo", "Toggle completion").
		WithKeyBinding("ctrl+c", "quit", "Quit application")

	builder = builder.Setup(func(ctx *bubbly.Context) {
		// Initialize todos with refs for reactivity
		initialTodos := []components.Todo{
			{
				ID:        1,
				Text:      "Learn BubblyUI",
				Completed: bubbly.NewRef(true),
			},
			{
				ID:        2,
				Text:      "Enable Dev Tools",
				Completed: bubbly.NewRef(true),
			},
			{
				ID:        3,
				Text:      "Inspect Component Tree",
				Completed: bubbly.NewRef(false),
			},
			{
				ID:        4,
				Text:      "View State in DevTools",
				Completed: bubbly.NewRef(false),
			},
			{
				ID:        5,
				Text:      "Navigate with Keyboard",
				Completed: bubbly.NewRef(false),
			},
		}

		// State: Todos list (use NewRef for typed refs)
		todos := bubbly.NewRef(initialTodos)

		// State: Selected index  
		selectedIndex := bubbly.NewRef(0)

		// Create child components
		header, err := components.CreateHeader(components.HeaderProps{
			Title: "Component Inspection Demo",
		})
		if err != nil {
			ctx.Expose("error", err)
			return
		}

		todoList, err := components.CreateTodoList(components.TodoListProps{
			Todos:         todos,
			SelectedIndex: selectedIndex,
		})
		if err != nil {
			ctx.Expose("error", err)
			return
		}

		footer, err := components.CreateFooter(components.FooterProps{
			Todos: todos,
		})
		if err != nil {
			ctx.Expose("error", err)
			return
		}

		// Event handlers
		ctx.On("selectPrevious", func(_ interface{}) {
			current := selectedIndex.Get().(int)
			if current > 0 {
				selectedIndex.Set(current - 1)
			}
			// Recreate TodoList with new selection
			newList, _ := components.CreateTodoList(components.TodoListProps{
				Todos:         todos,
				SelectedIndex: selectedIndex,
			})
			ctx.ExposeComponent("todoList", newList)
		})

		ctx.On("selectNext", func(_ interface{}) {
			current := selectedIndex.Get().(int)
			todoList := todos.Get().([]components.Todo)
			if current < len(todoList)-1 {
				selectedIndex.Set(current + 1)
			}
			// Recreate TodoList with new selection
			newList, _ := components.CreateTodoList(components.TodoListProps{
				Todos:         todos,
				SelectedIndex: selectedIndex,
			})
			ctx.ExposeComponent("todoList", newList)
		})

		ctx.On("toggleTodo", func(_ interface{}) {
			current := selectedIndex.Get().(int)
			todoList := todos.Get().([]components.Todo)
			if current >= 0 && current < len(todoList) {
				todo := todoList[current]
				completed := todo.Completed.Get().(bool)
				todo.Completed.Set(!completed)

				// Trigger update
				todos.Set(todoList)

				// Recreate components
				newList, _ := components.CreateTodoList(components.TodoListProps{
					Todos:         todos,
					SelectedIndex: selectedIndex,
				})
				ctx.ExposeComponent("todoList", newList)

				newFooter, _ := components.CreateFooter(components.FooterProps{
					Todos: todos,
				})
				ctx.ExposeComponent("footer", newFooter)
			}
		})

		// Expose state for dev tools inspection
		ctx.Expose("todos", todos)
		ctx.Expose("selectedIndex", selectedIndex)

		// Expose child components
		ctx.ExposeComponent("header", header)
		ctx.ExposeComponent("todoList", todoList)
		ctx.ExposeComponent("footer", footer)

		// Lifecycle hook
		ctx.OnMounted(func() {
			// App mounted - visible in dev tools component tree
		})
	})

	builder = builder.Template(func(ctx bubbly.RenderContext) string {
		// Get child components
		header := ctx.Get("header").(bubbly.Component)
		todoList := ctx.Get("todoList").(bubbly.Component)
		footer := ctx.Get("footer").(bubbly.Component)

		// Layout: Header → TodoList → Footer
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			header.View(),
			todoList.View(),
			footer.View(),
		)

		// Add container padding
		containerStyle := lipgloss.NewStyle().Padding(1, 2)
		return containerStyle.Render(content)
	})

	return builder.Build()
}
