// Package composables provides reusable logic for the todo router example.
package composables

import (
	"sync"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// Todo represents a single todo item.
type Todo struct {
	ID          int
	Title       string
	Description string
	Priority    string // "low", "medium", "high"
	Completed   bool
}

// TodosReturn is the return type of UseTodos composable.
type TodosReturn struct {
	Todos          *bubbly.Ref[[]Todo]
	SelectedIndex  *bubbly.Ref[int]
	NextID         *bubbly.Ref[int]
	AddTodo        func(title, description, priority string)
	UpdateTodo     func(id int, title, description, priority string)
	DeleteTodo     func(id int)
	ToggleTodo     func(id int)
	GetTodo        func(id int) *Todo
	GetStats       func() (total, completed, pending int)
	SelectNext     func()
	SelectPrevious func()
}

// sharedTodos holds the singleton instance of todos state.
var (
	sharedTodos     *TodosReturn
	sharedTodosOnce sync.Once
)

// UseTodos returns a shared todo state manager.
// This composable provides CRUD operations for todos and is shared across all components.
func UseTodos() *TodosReturn {
	sharedTodosOnce.Do(func() {
		// Initialize with sample todos
		todos := bubbly.NewRef([]Todo{
			{ID: 1, Title: "Learn BubblyUI Router", Description: "Understand how routing works in BubblyUI", Priority: "high", Completed: true},
			{ID: 2, Title: "Build Todo App", Description: "Create a todo app with multiple pages", Priority: "high", Completed: false},
			{ID: 3, Title: "Add Navigation Guards", Description: "Implement auth guards for protected routes", Priority: "medium", Completed: false},
			{ID: 4, Title: "Style with Lipgloss", Description: "Make the app look beautiful", Priority: "low", Completed: false},
		})
		selectedIndex := bubbly.NewRef(0)
		nextID := bubbly.NewRef(5)

		sharedTodos = &TodosReturn{
			Todos:         todos,
			SelectedIndex: selectedIndex,
			NextID:        nextID,
		}

		// AddTodo adds a new todo
		sharedTodos.AddTodo = func(title, description, priority string) {
			if title == "" {
				return
			}
			if priority == "" {
				priority = "medium"
			}
			id := nextID.GetTyped()
			todoList := todos.GetTyped()
			newTodo := Todo{
				ID:          id,
				Title:       title,
				Description: description,
				Priority:    priority,
				Completed:   false,
			}
			todos.Set(append(todoList, newTodo))
			nextID.Set(id + 1)
		}

		// UpdateTodo updates an existing todo
		sharedTodos.UpdateTodo = func(id int, title, description, priority string) {
			todoList := todos.GetTyped()
			for i, todo := range todoList {
				if todo.ID == id {
					todoList[i].Title = title
					todoList[i].Description = description
					todoList[i].Priority = priority
					todos.Set(todoList)
					return
				}
			}
		}

		// DeleteTodo removes a todo by ID
		sharedTodos.DeleteTodo = func(id int) {
			todoList := todos.GetTyped()
			for i, todo := range todoList {
				if todo.ID == id {
					newList := append(todoList[:i], todoList[i+1:]...)
					todos.Set(newList)
					// Adjust selection if needed
					idx := selectedIndex.GetTyped()
					if idx >= len(newList) && len(newList) > 0 {
						selectedIndex.Set(len(newList) - 1)
					}
					return
				}
			}
		}

		// ToggleTodo toggles the completed status
		sharedTodos.ToggleTodo = func(id int) {
			todoList := todos.GetTyped()
			for i, todo := range todoList {
				if todo.ID == id {
					todoList[i].Completed = !todoList[i].Completed
					todos.Set(todoList)
					return
				}
			}
		}

		// GetTodo returns a todo by ID
		sharedTodos.GetTodo = func(id int) *Todo {
			todoList := todos.GetTyped()
			for _, todo := range todoList {
				if todo.ID == id {
					t := todo // Create copy to avoid returning pointer to loop variable
					return &t
				}
			}
			return nil
		}

		// GetStats returns todo statistics
		sharedTodos.GetStats = func() (total, completed, pending int) {
			todoList := todos.GetTyped()
			total = len(todoList)
			for _, todo := range todoList {
				if todo.Completed {
					completed++
				}
			}
			pending = total - completed
			return
		}

		// SelectNext moves selection down
		sharedTodos.SelectNext = func() {
			todoList := todos.GetTyped()
			idx := selectedIndex.GetTyped()
			if idx < len(todoList)-1 {
				selectedIndex.Set(idx + 1)
			}
		}

		// SelectPrevious moves selection up
		sharedTodos.SelectPrevious = func() {
			idx := selectedIndex.GetTyped()
			if idx > 0 {
				selectedIndex.Set(idx - 1)
			}
		}
	})

	return sharedTodos
}

// ResetTodos resets the shared todos state (useful for testing).
func ResetTodos() {
	sharedTodosOnce = sync.Once{}
	sharedTodos = nil
}
