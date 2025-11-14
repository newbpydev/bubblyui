package main

import "github.com/newbpydev/bubblyui/pkg/bubbly"

// TodosComposable encapsulates todo list logic and reactive state
type TodosComposable struct {
	Items          *bubbly.Ref[[]Todo]
	Add            func(title string)
	Toggle         func(index int)
	Delete         func(index int)
	CompletedCount *bubbly.Computed[interface{}]
	TotalCount     *bubbly.Computed[interface{}]
}

// UseTodos creates a todos composable with reactive state
// This follows Vue's Composition API pattern for reusable logic
func UseTodos(ctx *bubbly.Context) *TodosComposable {
	// Create reactive state
	items := bubbly.NewRef([]Todo{
		{ID: 1, Title: "Connect AI assistant to MCP server", Completed: false},
		{ID: 2, Title: "Query todo list via AI", Completed: false},
		{ID: 3, Title: "Inspect component state", Completed: false},
	})

	// Create computed values (automatically update when items change)
	completedCount := ctx.Computed(func() interface{} {
		todos := items.Get().([]Todo)
		count := 0
		for _, todo := range todos {
			if todo.Completed {
				count++
			}
		}
		return count
	})

	totalCount := ctx.Computed(func() interface{} {
		todos := items.Get().([]Todo)
		return len(todos)
	})

	// Define methods
	add := func(title string) {
		todos := items.Get().([]Todo)
		newID := 1
		if len(todos) > 0 {
			newID = todos[len(todos)-1].ID + 1
		}
		newTodo := Todo{
			ID:        newID,
			Title:     title,
			Completed: false,
		}
		items.Set(append(todos, newTodo))
	}

	toggle := func(index int) {
		todos := items.Get().([]Todo)
		if index >= 0 && index < len(todos) {
			todos[index].Completed = !todos[index].Completed
			items.Set(todos)
		}
	}

	deleteTodo := func(index int) {
		todos := items.Get().([]Todo)
		if index >= 0 && index < len(todos) {
			newTodos := append(todos[:index], todos[index+1:]...)
			items.Set(newTodos)
		}
	}

	return &TodosComposable{
		Items:          items,
		Add:            add,
		Toggle:         toggle,
		Delete:         deleteTodo,
		CompletedCount: completedCount,
		TotalCount:     totalCount,
	}
}
